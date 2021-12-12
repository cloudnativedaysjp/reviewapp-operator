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
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/metrics"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	GitRemoteRepoAppService   *services.GitRemoteRepoAppService
	GitRemoteRepoInfraService *services.GitRemoteRepoInfraService
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argoproj.io,resources=applications,verbs=get;list;watch

const (
	finalizer = "reviewapp.finalizers.cloudnativedays.jp"
)

func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(fmt.Sprintf("fetching ReviewApp resource: %s/%s", req.Namespace, req.Name))
	ra, err := kubernetes.GetReviewApp(ctx, r.Client, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ra), req.Namespace, req.Name))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if result, err := r.prepare(ctx, ra); err != nil {
		if myerrors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	}

	// Add Finalizers
	if err := kubernetes.AddFinalizersToReviewApp(ctx, r.Client, ra, finalizer); err != nil {
		return ctrl.Result{}, err
	}

	// Handle deletion reconciliation loop.
	if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, ra)
	}
	return r.reconcile(ctx, ra)
}

func (r *ReviewAppReconciler) reconcile(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (result ctrl.Result, err error) {
	metrics.SetMetrics(*ra)

	// run/skip processes by ReviewApp state
	errs := []error{}
	if reflect.DeepEqual(ra.Status, dreamkastv1alpha1.ReviewAppStatus{}) ||
		ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo {
		result, err = r.confirmAppRepoIsUpdated(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeWatchingTemplates {
		result, err = r.confirmTemplatesAreUpdated(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo {
		result, err = r.deployReviewAppManifestsToInfraRepo(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo {
		result, err = r.commentToAppRepoPullRequest(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// update status
	if err := kubernetes.UpdateReviewAppStatus(ctx, r.Client, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, kerrors.NewAggregate(errs)
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (ctrl.Result, error) {
	metrics.RemoveMetrics(*ra)

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
		r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key,
	)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// delete some manifests
	deleteManifestsParam := services.DeleteManifestsParam{
		Org:    ra.Spec.InfraTarget.Organization,
		Repo:   ra.Spec.InfraTarget.Repository,
		Branch: ra.Spec.InfraTarget.Branch,
		CommitMsg: fmt.Sprintf(
			"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.AppTarget.Organization,
			ra.Spec.AppTarget.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		Username: ra.Spec.InfraTarget.Username,
		Token:    gitRemoteRepoCred,
	}
	if _, err := r.GitRemoteRepoInfraService.DeleteManifests(ctx, deleteManifestsParam, ra); err != nil {
		return ctrl.Result{}, err
	}

	// Remove Finalizers
	if err := kubernetes.RemoveFinalizersToReviewApp(ctx, r.Client, ra, finalizer); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReviewAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dreamkastv1alpha1.ReviewApp{}).
		Complete(r)
}
