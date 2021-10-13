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
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

const (
	finalizer                                   = "reviewapp.finalizers.cloudnativedays.jp"
	annotationAppOrgNameForArgoCDApplication    = "dreamkast.cloudnativedays.jp/app-organization"
	annotationAppRepoNameForArgoCDApplication   = "dreamkast.cloudnativedays.jp/app-repository"
	annotationAppCommitHashForArgoCDApplication = "dreamkast.cloudnativedays.jp/app-commit-hash"
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

// comment: この関数内で呼んでいるreconcile~の関数名ですが、reconcileとCheck、reconcileとUpdateのように動詞が続いているのが気になりました。
// reconcileCheckAppRepository なら reconcileAppRepository , reconcileUpdateInfraReposiotry なら reconcileInfraReposiotry という名前がよいように思います。
// 追記: でもreconcileDeleteは割とよく使われてるしあんまり気にしなくてもいいかも
func (r *ReviewAppReconciler) reconcile(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (result ctrl.Result, err error) {

	errs := []error{}
	if reflect.DeepEqual(ra.Status, dreamkastv1beta1.ReviewAppStatus{}) ||
		ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeWatchingAppRepo {
		result, err = r.reconcileCheckAppRepository(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeWatchingTemplates {
		result, err = r.reconcileCheckAtAndMt(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeCheckedAppRepo {
		result, err = r.reconcileUpdateInfraReposiotry(ctx, ra)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo {
		result, err = r.reconcileSendMessageToAppRepoPR(ctx, ra)
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

func (r *ReviewAppReconciler) reconcileCheckAppRepository(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	var updated bool

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx, r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
			return ctrl.Result{}, nil
		}
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

	if pr.HeadCommitSha != ra.Status.Sync.AppRepoLatestCommitSha {
		updated = true
	}

	// get ArgoCD Application name
	argocdAppNamespacedName, err := kubernetes.PickNamespacedNameFromObjectStr(ctx, ra.Spec.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.ApplicationName = argocdAppNamespacedName.Name
	ra.Status.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	ra.Status.Sync.AppRepoLatestCommitSha = pr.HeadCommitSha
	if updated {
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeCheckedAppRepo
	} else {
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeWatchingTemplates
	}
	return ctrl.Result{}, nil
}

// comment: 何を行う関数なのか、関数名から読み取れませんでした
func (r *ReviewAppReconciler) reconcileCheckAtAndMt(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	var updated bool
	if !reflect.DeepEqual(ra.Spec.Application, ra.Status.Sync.Application) {
		updated = true
	}
	if !reflect.DeepEqual(ra.Spec.Manifests, ra.Status.Sync.Manifests) {
		updated = true
	}

	// get ArgoCD Application name
	argocdAppNamespacedName, err := kubernetes.PickNamespacedNameFromObjectStr(ctx, ra.Spec.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.ApplicationName = argocdAppNamespacedName.Name
	ra.Status.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	if updated {
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeCheckedAppRepo
	} else {
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeWatchingAppRepo
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileUpdateInfraReposiotry(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {

	// set annotations to Argo CD Application
	argocdAppStr := ra.Spec.Application
	argocdAppStr, err := kubernetes.SetAnnotationToObjectStr(ctx,
		argocdAppStr, annotationAppOrgNameForArgoCDApplication, ra.Spec.AppTarget.Organization,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	argocdAppStr, err = kubernetes.SetAnnotationToObjectStr(ctx,
		argocdAppStr, annotationAppRepoNameForArgoCDApplication, ra.Spec.AppTarget.Repository,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	argocdAppStr, err = kubernetes.SetAnnotationToObjectStr(ctx,
		argocdAppStr, annotationAppCommitHashForArgoCDApplication, ra.Status.Sync.AppRepoLatestCommitSha,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	argocdAppStrWithoutAnnotations := ra.Spec.Application
	ra.Spec.Application = argocdAppStr

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

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	updateManifestParam := services.UpdateManifestsParam{
		Org:    ra.Spec.InfraTarget.Organization,
		Repo:   ra.Spec.InfraTarget.Repository,
		Branch: ra.Spec.InfraTarget.Branch,
		CommitMsg: fmt.Sprintf(
			"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.AppTarget.Organization,
			ra.Spec.AppTarget.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		Username: ra.Spec.InfraTarget.Username,
		Token:    gitRemoteRepoCred,
	}
	gp, err := r.GitRemoteRepoInfraService.UpdateManifests(ctx, updateManifestParam, ra)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo
	ra.Status.Sync.InfraRepoLatestCommitSha = gp.LatestCommitSha
	ra.Status.Sync.Application = argocdAppStrWithoutAnnotations
	ra.Status.Sync.Manifests = ra.Spec.Manifests

	return ctrl.Result{}, nil
}

// comment: 一度PRにコメントした後にReviewAppが更新されると、投稿済みコメントを消して再度コメントしてもいいかもしれないなと思いました。
// そうすると積んだコミットがArgoCDに反映されたのがわかって便利かなと。
func (r *ReviewAppReconciler) reconcileSendMessageToAppRepoPR(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
	// check appRepoSha from annotations in ArgoCD Application
	hashInArgoCDApplication, err := kubernetes.GetArgoCDAppAnnotation(
		ctx, r.Client, ra.Status.Sync.ApplicationNamespace, ra.Status.Sync.ApplicationName, annotationAppCommitHashForArgoCDApplication,
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
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	//
	param := services.IsApplicationUpdatedParam{
		Org:                     ra.Spec.AppTarget.Organization,
		Repo:                    ra.Spec.AppTarget.Repository,
		PrNum:                   ra.Spec.AppPrNum,
		Username:                ra.Spec.AppTarget.Username,
		Token:                   gitRemoteRepoCred,
		HashInRA:                ra.Status.Sync.AppRepoLatestCommitSha,
		HashInArgoCDApplication: hashInArgoCDApplication,
	}
	updated, err := r.GitRemoteRepoAppService.IsApplicationUpdated(ctx, param)
	if err != nil {
		if myerrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// if ArgoCD Application updated, send message to PR of AppRepo
	if updated {
		if ra.Spec.AppConfig.Message != "" &&
			(ra.Spec.AppConfig.SendMessageEveryTime || !ra.Status.AlreadySentMessage) {

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
		ra.Status.AlreadySentMessage = true
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) (ctrl.Result, error) {
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
		For(&dreamkastv1beta1.ReviewApp{}).
		Complete(r)
}
