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
	"path/filepath"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/wire"
	"github.com/go-logr/logr"
)

// ReviewAppReconciler reconciles a ReviewApp object
type ReviewAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dreamkast.cloudnativedays.jp,resources=reviewapps/finalizers,verbs=update

const finalizer = "reviewapp.finalizers.cloudnativedays.jp"

func (r *ReviewAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ra dreamkastv1beta1.ReviewApp
	r.Log.Info(fmt.Sprintf("fetching %s resource: %s/%s", reflect.TypeOf(ra), req.Namespace, req.Name))
	if err := r.Get(ctx, req.NamespacedName, &ra); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(ra), req.Namespace, req.Name))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// initalize some service
	gitRemoteRepoAppService, err := wire.NewGitRemoteRepoAppService(r.Log, r.Client, ra.Spec.App.Username)
	if err != nil {
		return ctrl.Result{}, err
	}
	gitRemoteRepoInfraService, err := wire.NewGitRemoteRepoInfraService(r.Log, r.Client, ra.Spec.App.Username)
	if err != nil {
		return ctrl.Result{}, err
	}
	k8sService, err := wire.NewKubernetesService(r.Log, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Add Finalizers
	if err := k8sService.AddFinalizersToReviewApp(ctx, &ra, finalizer); err != nil {
		return ctrl.Result{}, err
	}

	// Handle deletion reconciliation loop.
	if !ra.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, k8sService, gitRemoteRepoInfraService, &ra)
	}

	return r.reconcile(ctx, gitRemoteRepoAppService, gitRemoteRepoInfraService, k8sService, &ra)
}

func (r *ReviewAppReconciler) reconcile(ctx context.Context,
	gitRemoteRepoAppService *services.GitRemoteRepoAppService,
	gitRemoteRepoInfraService *services.GitRemoteRepoInfraService,
	k8sService *services.KubernetesService,
	ra *dreamkastv1beta1.ReviewApp,
) (result ctrl.Result, err error) {

	emptyStatus := dreamkastv1beta1.ReviewAppStatus{}
	if ra.Status == emptyStatus || ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeWatchingAppRepo {
		result, err = r.reconcileCheckAppRepository(ctx, ra, gitRemoteRepoAppService, k8sService)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeCheckedAppRepo {
		result, err = r.reconcileUpdateInfraReposiotry(ctx, ra, gitRemoteRepoInfraService, k8sService)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if ra.Status.Sync.Status == dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo {
		result, err = r.reconcileSendMessageToAppRepoPR(ctx, ra, gitRemoteRepoAppService, gitRemoteRepoInfraService, k8sService)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return
}

func (r *ReviewAppReconciler) reconcileCheckAppRepository(ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
	gitRemoteRepoAppService *services.GitRemoteRepoAppService,
	k8sService *services.KubernetesService,
) (ctrl.Result, error) {
	// check PRs specified by spec.appRepo.repository
	pr, err := gitRemoteRepoAppService.GetOpenPullRequest(ctx,
		ra.Spec.App.Organization, ra.Spec.App.Repository, ra.Spec.AppPrNum,
		services.AccessToAppRepoInput{
			SecretNamespace: ra.Namespace,
			SecretName:      ra.Spec.App.GitSecretRef.Name,
			SecretKey:       ra.Spec.App.GitSecretRef.Key,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// if RA.status already applied with above PR, ealry return
	if pr.HeadCommitSha == ra.Status.Sync.AppRepoLatestCommitSha {
		return ctrl.Result{}, nil
	}

	// get ArgoCD Application name
	applicationName, err := k8sService.GetArgoCDApplicationName(ctx, ra.Spec.Application)
	if err != nil {
		return ctrl.Result{}, err
	}
	applicationNamespace, err := k8sService.GetArgoCDApplicationNamespace(ctx, ra.Spec.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync = dreamkastv1beta1.SyncStatus{
		Status:                 dreamkastv1beta1.SyncStatusCodeCheckedAppRepo,
		ApplicationName:        applicationName,
		ApplicationNamespace:   applicationNamespace,
		AppRepoLatestCommitSha: pr.HeadCommitSha,
	}
	if err := k8sService.UpdateReviewAppStatus(ctx, ra); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileUpdateInfraReposiotry(ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
	gitRemoteRepoInfraService *services.GitRemoteRepoInfraService,
	k8sService *services.KubernetesService,
) (ctrl.Result, error) {

	// add annotations to Application
	applicationWithAnnotations, err := k8sService.GetArgoCDApplicationWithAnnotations(
		ctx, ra.Spec.Application,
		ra.Spec.App.Organization, ra.Spec.App.Repository, ra.Status.Sync.AppRepoLatestCommitSha,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	inputSecret := services.AccessToInfraRepoInput{
		Namespace: ra.Namespace,
		Name:      ra.Spec.Infra.GitSecretRef.Name,
		Key:       ra.Spec.Infra.GitSecretRef.Key,
	}
	inputManifests := append([]services.UpdateManifestsInput{}, services.UpdateManifestsInput{
		Content: applicationWithAnnotations,
		Path:    ra.Spec.Infra.ArgoCDApp.Filepath,
	})
	for filename, manifest := range ra.Spec.Manifests {
		inputManifests = append(inputManifests, services.UpdateManifestsInput{
			Content: manifest,
			Path:    filepath.Join(ra.Spec.Infra.Manifests.Dirpath, filename),
		})
	}
	gp, err := gitRemoteRepoInfraService.UpdateManifests(ctx,
		ra.Spec.Infra.Organization,
		ra.Spec.Infra.Repository,
		ra.Spec.Infra.TargetBranch,
		fmt.Sprintf(
			"Automatic update by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.App.Organization,
			ra.Spec.App.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		inputSecret, inputManifests...,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeUpdatedInfraRepo
	ra.Status.Sync.InfraRepoLatestCommitSha = gp.LatestCommitSha
	if err := k8sService.UpdateReviewAppStatus(ctx, ra); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileSendMessageToAppRepoPR(ctx context.Context,
	ra *dreamkastv1beta1.ReviewApp,
	gitRemoteRepoAppService *services.GitRemoteRepoAppService,
	gitRemoteRepoInfraService *services.GitRemoteRepoInfraService,
	k8sService *services.KubernetesService,
) (ctrl.Result, error) {
	// check appRepoSha in annotations of ArgoCD Application
	hashInArgoCDApplication, err := k8sService.ArgoCDApplictionIFace.GetAnnotationOfArgoCDApplication(
		ctx, ra.Status.Sync.ApplicationNamespace, ra.Status.Sync.ApplicationName, models.AnnotationAppCommitHashForArgoCDApplication,
	)
	if err != nil {
		if models.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// compare ra.Status.Sync.InfraRepoLatestCommitSha with Application.annotation
	inputSecret := services.AccessToInfraRepoInput{
		Namespace: ra.Namespace,
		Name:      ra.Spec.Infra.GitSecretRef.Name,
		Key:       ra.Spec.Infra.GitSecretRef.Key,
	}
	updated, err := gitRemoteRepoAppService.CheckApplicationUpdated(ctx,
		ra.Spec.App.Organization,
		ra.Spec.App.Repository,
		ra.Spec.AppPrNum,
		inputSecret,
		ra.Status.Sync.AppRepoLatestCommitSha, hashInArgoCDApplication,
	)
	if err != nil {
		if models.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// if ArgoCD Application updated, send message to PR of AppRepo
	if updated {
		pr, err := gitRemoteRepoAppService.GetOpenPullRequest(ctx,
			ra.Spec.App.Organization,
			ra.Spec.App.Repository,
			ra.Spec.AppPrNum,
			services.AccessToAppRepoInput{
				SecretNamespace: ra.Namespace,
				SecretName:      ra.Spec.App.GitSecretRef.Name,
				SecretKey:       ra.Spec.App.GitSecretRef.Key,
			},
		)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err := gitRemoteRepoAppService.SendMessage(ctx, pr, ra.Spec.App.Message,
			services.AccessToAppRepoInput{
				SecretNamespace: ra.Namespace,
				SecretName:      ra.Spec.App.GitSecretRef.Name,
				SecretKey:       ra.Spec.App.GitSecretRef.Key,
			},
		); err != nil {
			return ctrl.Result{}, err
		}

		// update ReviewApp.Status
		ra.Status.Sync.Status = dreamkastv1beta1.SyncStatusCodeWatchingAppRepo
		if err := k8sService.UpdateReviewAppStatus(ctx, ra); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context,
	k8sService *services.KubernetesService,
	gitRemoteRepoInfraService *services.GitRemoteRepoInfraService,
	ra *dreamkastv1beta1.ReviewApp,
) (ctrl.Result, error) {
	// delete some manifests
	inputSecret := services.AccessToInfraRepoInput{
		Namespace: ra.Namespace,
		Name:      ra.Spec.Infra.GitSecretRef.Name,
		Key:       ra.Spec.Infra.GitSecretRef.Key,
	}
	inputManifests := append([]services.DeleteManifestsInput{}, services.DeleteManifestsInput{
		Path: ra.Spec.Infra.ArgoCDApp.Filepath,
	})
	for filename := range ra.Spec.Manifests {
		inputManifests = append(inputManifests, services.DeleteManifestsInput{
			Path: filepath.Join(ra.Spec.Infra.Manifests.Dirpath, filename),
		})
	}
	_, err := gitRemoteRepoInfraService.DeleteManifests(ctx,
		ra.Spec.Infra.Organization,
		ra.Spec.Infra.Repository,
		ra.Spec.Infra.TargetBranch,
		fmt.Sprintf(
			"Automatic GC by cloudnativedays/reviewapp-operator (%s/%s@%s)",
			ra.Spec.App.Organization,
			ra.Spec.App.Repository,
			ra.Status.Sync.AppRepoLatestCommitSha,
		),
		inputSecret, inputManifests...,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Remove Finalizers
	if err := k8sService.RemoveFinalizersToReviewApp(ctx, ra, finalizer); err != nil {
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
