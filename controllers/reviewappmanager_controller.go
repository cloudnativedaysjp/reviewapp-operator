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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/metrics"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
)

var (
	datetimeFactoryForRAM = utils.NewDatetimeFactory()
)

// ReviewAppManagerReconciler reconciles a ReviewAppManager object
type ReviewAppManagerReconciler struct {
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	K8sRepository    repositories.KubernetesRepository
	GitApiRepository repositories.GitAPI
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewappmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=applicationtemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=manifeststemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *ReviewAppManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("fetching ReviewAppManager resource: %s/%s", req.Namespace, req.Name))
	ram, err := r.K8sRepository.GetReviewAppManager(ctx, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.removeMetrics(req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, ram)
}

func (r *ReviewAppManagerReconciler) reconcile(ctx context.Context, ram models.ReviewAppManager) (ctrl.Result, error) {
	// init model
	appRepoTarget := ram.AppRepoTarget()

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ram.Namespace, &appRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
			r.Log.Info(err.Error())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// set credential
	gitRemoteRepoCred := models.NewGitCredential(ram.Spec.AppTarget.Username, gitRemoteRepoToken)
	if err := r.GitApiRepository.WithCredential(gitRemoteRepoCred); err != nil {
		return ctrl.Result{}, err
	}
	// list PRs
	prs, err := r.GitApiRepository.ListOpenPullRequests(ctx, appRepoTarget)
	if err != nil {
		return ctrl.Result{}, err
	}
	// add metrics
	metrics.RequestToGitHubApiCounterVec.WithLabelValues(
		ram.Name,
		ram.Namespace,
		"ReviewAppManager",
	).Add(1)

	// exclude PRs with specific labels
	prs = prs.ExcludeSpecificPR(ram)
	// apply ReviewApp
	var syncedPullRequests []dreamkastv1alpha1.ReviewAppManagerStatusSyncedPullRequests
	for _, pr := range prs {
		// init templator
		v := models.NewTemplator(ram, pr)
		// generate RA
		ra, err := ram.GenerateReviewApp(pr, v, datetimeFactoryForRAM)
		if err != nil {
			return ctrl.Result{}, err
		}
		// get RA
		if _, err := r.K8sRepository.GetReviewApp(ctx, ra.Namespace, ra.Name); err != nil {
			if !myerrors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			// if ReviewApp Object has not existed, set to status.sync.status
			ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeInitialize
		}
		// apply RA
		if err := r.K8sRepository.ApplyReviewAppWithOwnerRef(ctx, ra, ram); err != nil {
			return ctrl.Result{}, err
		}
		// update Status
		if err := r.K8sRepository.PatchReviewAppStatus(ctx, ra); err != nil {
			return ctrl.Result{}, err
		}
		// update values for updating RAM.status
		syncedPullRequests = append(syncedPullRequests, dreamkastv1alpha1.ReviewAppManagerStatusSyncedPullRequests{
			Organization:  pr.Organization,
			Repository:    pr.Repository,
			Number:        pr.Number,
			ReviewAppName: ra.Name,
		})
	}
	// delete RA that only exists ResourceStatus
	for _, name := range ram.ListOutOfSyncReviewAppName(prs) {
		if err := r.K8sRepository.DeleteReviewApp(ctx, ram.Namespace, name); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}
	// update ReviewAppManager Status
	ram.Status.SyncedPullRequests = syncedPullRequests
	if err := r.K8sRepository.UpdateReviewAppManagerStatus(ctx, ram); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppManagerReconciler) removeMetrics(name, namespace string) {
	metrics.RequestToGitHubApiCounterVec.DeleteLabelValues(
		name,
		namespace,
		"ReviewAppManager",
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	setupLog := ctrl.Log.WithName("setup")
	var err error
	r.K8sRepository, err = wire.NewKubernetesRepository(r.Log, mgr.GetClient())
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewKubernetesRepository")
		os.Exit(1)
	}
	r.GitApiRepository, err = wire.NewGitHubAPIRepository(r.Log)
	if err != nil {
		setupLog.Error(err, "unable to initialize", "wire.NewGitHubAPIRepository")
		os.Exit(1)
	}
	mapFunc := handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		rams := dreamkastv1alpha1.ReviewAppManagerList{}
		_ = mgr.GetCache().List(context.Background(), &rams)
		for _, ram := range rams.Items {
			nn := dreamkastv1alpha1.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}
			switch object.(type) {
			case *dreamkastv1alpha1.ApplicationTemplate:
				if nn == ram.Spec.InfraConfig.ArgoCDApp.Template {
					return []reconcile.Request{{
						NamespacedName: types.NamespacedName{
							Name:      ram.Name,
							Namespace: ram.Namespace,
						},
					}}
				}
			case *dreamkastv1alpha1.ManifestsTemplate:
				for _, template := range ram.Spec.InfraConfig.Manifests.Templates {
					if nn == template {
						return []reconcile.Request{{
							NamespacedName: types.NamespacedName{
								Name:      ram.Name,
								Namespace: ram.Namespace,
							},
						}}
					}
				}
			}
		}
		return []reconcile.Request{}
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1alpha1.ReviewAppManager{}).
		Owns(&dreamkastv1alpha1.ReviewApp{}).
		Watches(
			&source.Kind{Type: &dreamkastv1alpha1.ApplicationTemplate{}},
			mapFunc,
		).
		Watches(
			&source.Kind{Type: &dreamkastv1alpha1.ManifestsTemplate{}},
			mapFunc,
		).
		Complete(r)
}
