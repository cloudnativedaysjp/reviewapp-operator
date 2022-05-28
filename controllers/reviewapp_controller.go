/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/metrics"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

const (
	raFinalizer = "reviewapp.finalizers.cloudnativedays.jp"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	K8s          kubernetes.KubernetesIface
	GitApi       githubapi.GitApiIface
	GitLocalRepo gitcommand.GitLocalRepoIface
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch

func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("fetching ReviewApp resource: %s/%s", req.Namespace, req.Name))
	ra, err := r.K8s.GetReviewApp(ctx, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	dto, usingPrCache, result, err := r.prepare(ctx, ra)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			return result, nil
		}
		return result, err
	}

	// Add Finalizers
	if err := r.K8s.AddFinalizersToReviewApp(ctx, ra, raFinalizer); err != nil {
		return ctrl.Result{}, err
	}

	// Handle deletion reconciliation loop.
	if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, *dto)
	}

	// if PR object is not found, delete RA object myself
	if usingPrCache {
		r.Log.Info(fmt.Sprintf("%v: delete my object myself", err))
		if err := r.K8s.DeleteReviewApp(ctx, ra.Namespace, ra.Name); err != nil {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, *dto)
}

func (r *ReviewAppReconciler) reconcile(ctx context.Context, dto ReviewAppPhaseDTO) (ctrl.Result, error) {
	ra := &dto.ReviewApp
	res := &Result{}

	// set metrics
	metrics.UpVec.WithLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	).Set(1)

	// if s.sync.status is empty, set SyncStatusCodeInitialize
	if reflect.DeepEqual(ra.Status, dreamkastv1alpha1.ReviewAppStatus{}) {
		dto.ReviewApp.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeInitialize
	}
	// run/skip each phase by ReviewApp state
	type phaseWithCond struct {
		cond  bool
		phase func(context.Context, ReviewAppPhaseDTO, *Result) (dreamkastv1alpha1.ReviewAppStatus, error)
	}
	var err error
	for _, p := range []phaseWithCond{
		{
			cond: ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeInitialize ||
				ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates,
			phase: r.confirmUpdated,
		},
		{
			cond:  ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
			phase: r.deployReviewAppManifestsToInfraRepo,
		},
		{
			cond:  ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo,
			phase: r.commentToAppRepoPullRequest,
		},
	} {
		if !p.cond {
			continue
		}
		ra.Status, err = p.phase(ctx, dto, res)
		if err != nil {
			break
		}
	}

	// if status does not update, requeue Reconcile
	if !res.StatusWasUpdated {
		return res.Result, err
	}
	if errStatus := r.K8s.PatchReviewAppStatus(ctx, *ra); errStatus != nil {
		return res.Result, kerrors.NewAggregate(append([]error{}, err, errStatus))
	}
	return res.Result, err
}

func (r *ReviewAppReconciler) removeMetrics(ra dreamkastv1alpha1.ReviewApp) {
	metrics.UpVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	)
	metrics.RequestToGitHubApiCounterVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		"ReviewApp",
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")
	var err error
	r.K8s, err = wire.NewKubernetes(r.Log, mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewKubernetesRepository")
		os.Exit(1)
	}
	r.GitApi, err = wire.NewGitHubApi(r.Log)
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewGitHubAPIRepository")
		os.Exit(1)
	}
	r.GitLocalRepo, err = wire.NewGitLocalRepo(r.Log, exec.New())
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewGitCommandRepository")
		os.Exit(1)
	}
	mapFunc := handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		ras := dreamkastv1alpha1.ReviewAppList{}
		_ = mgr.GetCache().List(context.Background(), &ras)
		for _, ra := range ras.Items {
			nn := dreamkastv1alpha1.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}
			switch object.(type) {
			case *dreamkastv1alpha1.ApplicationTemplate:
				if nn == ra.Spec.InfraConfig.ArgoCDApp.Template {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name:      ra.Name,
							Namespace: ra.Namespace,
						},
					}}
				}
			case *dreamkastv1alpha1.ManifestsTemplate:
				for _, template := range ra.Spec.InfraConfig.Manifests.Templates {
					if nn == template {
						return []reconcile.Request{{
							NamespacedName: types.NamespacedName{
								Name:      ra.Name,
								Namespace: ra.Namespace,
							},
						}}
					}
				}
			case *dreamkastv1alpha1.PullRequest:
				if nn == ra.Spec.PullRequest {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name:      ra.Name,
							Namespace: ra.Namespace,
						},
					}}
				}
			}
		}
		return []reconcile.Request{}
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1alpha1.ReviewApp{}).
		Watches(
			&source.Kind{Type: &dreamkastv1alpha1.ApplicationTemplate{}},
			mapFunc,
		).
		Watches(
			&source.Kind{Type: &dreamkastv1alpha1.ManifestsTemplate{}},
			mapFunc,
		).
		Watches(
			&source.Kind{Type: &dreamkastv1alpha1.PullRequest{}},
			mapFunc,
		).
		Complete(r)
}
