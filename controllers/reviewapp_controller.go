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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/apprepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/infrarepo"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	GitRemoteRepoAppService   *apprepo.GitRemoteRepoAppService
	GitRemoteRepoInfraService *infrarepo.GitRemoteRepoInfraService
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update

const finalizer = "reviewapp.finalizers.cloudnativedays.jp"

func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ra *dreamkastv1beta1.ReviewApp
	r.Log.Info(fmt.Sprintf("fetching %s resource: %s/%s", reflect.TypeOf(ra), req.Namespace, req.Name))
	ra, err := kubernetes.GetReviewApp(ctx, r.Client, req.Namespace, req.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ra), req.Namespace, req.Name))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
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

func (r *ReviewAppReconciler) reconcile(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (result ctrl.Result, err error) {

	emptyStatus := dreamkastv1beta1.ReviewAppStatus{}
	if ra.Status == emptyStatus || ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeWatchingAppRepo {
		result, err = r.reconcileCheckAppRepository(ctx, ra)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeCheckedAppRepo {
		result, err = r.reconcileUpdateInfraReposiotry(ctx, ra)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo {
		result, err = r.reconcileSendMessageToAppRepoPR(ctx, ra)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return
}

func (r *ReviewAppReconciler) reconcileCheckAppRepository(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx, r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key)
	if err != nil {
		return ctrl.Result{}, err
	}

	// check PRs specified by spec.appRepo.repository
	pr, err := r.GitRemoteRepoAppService.GetOpenPullRequest(ctx,
		ra.Spec.AppTarget.Organization, ra.Spec.AppTarget.Repository, ra.Spec.AppPrNum,
		ra.Spec.AppTarget.Username, gitRemoteRepoCred,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// if RA.status already applied with above PR, ealry return
	if pr.HeadCommitSha == ra.Status.Sync.AppRepoLatestCommitSha {
		return ctrl.Result{}, nil
	}

	// get ArgoCD Application name
	argocdAppNamespacedName, err := kubernetes.PickNamespacedNameFromArgoCDAppStr(ctx, ra.Spec.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync = dreamkastv1beta1.SyncStatus{
		Status:                 dreamkastv1beta1.SyncStatusCodeCheckedAppRepo,
		ApplicationName:        argocdAppNamespacedName.Name,
		ApplicationNamespace:   argocdAppNamespacedName.Namespace,
		AppRepoLatestCommitSha: pr.HeadCommitSha,
	}
	if err := kubernetes.UpdateReviewAppStatus(ctx, r.Client, ra); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileUpdateInfraReposiotry(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {

	// set annotations to Argo CD Application
	argocdAppStr := ra.Spec.Application
	argocdAppStr, err := kubernetes.SetAnnotationToArgoCDAppStr(ctx,
		argocdAppStr, kubernetes.AnnotationAppOrgNameForArgoCDApplication, ra.Spec.AppTarget.Organization,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	argocdAppStr, err = kubernetes.SetAnnotationToArgoCDAppStr(ctx,
		argocdAppStr, kubernetes.AnnotationAppRepoNameForArgoCDApplication, ra.Spec.AppTarget.Repository,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	argocdAppStr, err = kubernetes.SetAnnotationToArgoCDAppStr(ctx,
		argocdAppStr, kubernetes.AnnotationAppCommitHashForArgoCDApplication, ra.Status.Sync.AppRepoLatestCommitSha,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	ra.Spec.Application = argocdAppStr

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
		r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	gp, err := r.GitRemoteRepoInfraService.UpdateManifests(ctx,
		ra.Spec.InfraTarget.Organization, ra.Spec.InfraTarget.Repository, ra.Spec.InfraTarget.Branch,
		fmt.Sprintf(
			"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.AppTarget.Organization,
			ra.Spec.AppTarget.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		ra.Spec.InfraTarget.Username, gitRemoteRepoCred, ra,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo
	ra.Status.Sync.InfraRepoLatestCommitSha = gp.LatestCommitSha
	if err := kubernetes.UpdateReviewAppStatus(ctx, r.Client, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileSendMessageToAppRepoPR(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	// check appRepoSha from annotations in ArgoCD Application
	hashInArgoCDApplication, err := kubernetes.GetArgoCDAppAnnotation(
		ctx, r.Client, ra.Status.Sync.ApplicationNamespace, ra.Status.Sync.ApplicationName, kubernetes.AnnotationAppCommitHashForArgoCDApplication,
	)
	if err != nil {
		if myerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
		r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	//
	updated, err := r.GitRemoteRepoAppService.CheckApplicationUpdated(ctx,
		ra.Spec.AppTarget.Organization, ra.Spec.AppTarget.Repository, ra.Spec.AppPrNum,
		ra.Spec.AppTarget.Username, gitRemoteRepoCred,
		ra.Status.Sync.AppRepoLatestCommitSha, hashInArgoCDApplication,
	)
	if err != nil {
		if myerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// if ArgoCD Application updated, send message to PR of AppRepo
	if updated {
		if ra.Spec.AppConfig.Message != "" {

			// get gitRemoteRepo credential from Secret
			gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
				r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key,
			)
			if err != nil {
				return ctrl.Result{}, err
			}

			//
			pr, err := r.GitRemoteRepoAppService.GetOpenPullRequest(ctx,
				ra.Spec.AppTarget.Organization, ra.Spec.AppTarget.Repository, ra.Spec.AppPrNum,
				ra.Spec.AppTarget.Username, gitRemoteRepoCred,
			)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Send Message to AppRepo's PR
			if err := r.GitRemoteRepoAppService.SendMessage(ctx,
				pr, ra.Spec.AppConfig.Message, ra.Spec.AppTarget.Username, gitRemoteRepoCred,
			); err != nil {
				return ctrl.Result{}, err
			}
		}

		// update ReviewApp.Status
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeWatchingAppRepo
		if err := kubernetes.UpdateReviewAppStatus(ctx, r.Client, ra); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx,
		r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// delete some manifests
	if _, err := r.GitRemoteRepoInfraService.DeleteManifests(ctx,
		ra.Spec.InfraTarget.Organization, ra.Spec.InfraTarget.Repository, ra.Spec.InfraTarget.Branch,
		fmt.Sprintf(
			"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.AppTarget.Organization,
			ra.Spec.AppTarget.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		ra.Spec.InfraTarget.Username, gitRemoteRepoCred, ra,
	); err != nil {
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
		For(&dreamkastv1beta1.ReviewApp{}).
		Complete(r)
}
