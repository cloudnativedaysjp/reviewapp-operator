package controllers

import (
	"context"
	"fmt"
	"reflect"

	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/template"
)

const (
	annotationAppOrgNameForArgoCDApplication    = "dreamkast.cloudnativedays.jp/app-organization"
	annotationAppRepoNameForArgoCDApplication   = "dreamkast.cloudnativedays.jp/app-repository"
	annotationAppCommitHashForArgoCDApplication = "dreamkast.cloudnativedays.jp/app-commit-hash"
)

func (r *ReviewAppReconciler) prepare(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (result ctrl.Result, err error) {
	ra.Tmp.Manifests = make(map[string]string)

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoCred, err := kubernetes.GetSecretValue(ctx, r.Client, ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
		}
		return ctrl.Result{}, err
	}
	// check PRs specified by spec.appRepo.repository
	pr, err := r.GitRemoteRepoAppService.GetPullRequest(ctx,
		ra.Spec.AppTarget.Organization, ra.Spec.AppTarget.Repository, ra.Spec.AppPrNum,
		ra.Spec.AppTarget.Username, gitRemoteRepoCred,
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	ra.Tmp.PullRequest = (dreamkastv1alpha1.ReviewAppTmpPr)(*pr)

	// template ApplicationTemplate & ManifestsTemplate
	v := template.NewTemplateValue(
		pr.Organization, pr.Repository, pr.Branch, pr.Number,
		ra.Spec.InfraTarget.Organization, ra.Spec.InfraTarget.Repository,
		kubernetes.PickVariablesFromReviewApp(ctx, ra),
	)
	v = v.WithAppRepoLatestCommitSha(pr.HeadCommitSha)

	// get ApplicationTemplate & template to applicationStr
	at, err := kubernetes.GetApplicationTemplate(ctx, r.Client, ra.Spec.InfraConfig.ArgoCDApp.Template.Namespace, ra.Spec.InfraConfig.ArgoCDApp.Template.Name)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(at), ra.Spec.InfraConfig.ArgoCDApp.Template.Namespace, ra.Spec.InfraConfig.ArgoCDApp.Template.Name))
		}
		return ctrl.Result{}, err
	}
	if r.GitRemoteRepoAppService.IsCandidatePr(pr) {
		ra.Tmp.Application, err = v.Templating(at.Spec.CandidateTemplate)
	} else {
		ra.Tmp.Application, err = v.Templating(at.Spec.StableTemplate)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// get ManifestsTemplate & template to manifestsStr
	for _, mtNN := range ra.Spec.InfraConfig.Manifests.Templates {
		mt, err := kubernetes.GetManifestsTemplate(ctx, r.Client, mtNN.Namespace, mtNN.Name)
		if err != nil {
			if myerrors.IsNotFound(err) {
				r.Log.Info(fmt.Sprintf("%s %s/%s not found", reflect.TypeOf(mt), mtNN.Namespace, mtNN.Name))
			}
			return ctrl.Result{}, err
		}
		if r.GitRemoteRepoAppService.IsCandidatePr(pr) {
			ra.Tmp.Manifests, err = v.MapTemplatingAndAppend(ra.Tmp.Manifests, mt.Spec.CandidateData)
		} else {
			ra.Tmp.Manifests, err = v.MapTemplatingAndAppend(ra.Tmp.Manifests, mt.Spec.StableData)
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) confirmAppRepoIsUpdated(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (ctrl.Result, error) {
	var updated bool
	if ra.Tmp.PullRequest.HeadCommitSha != ra.Status.Sync.AppRepoLatestCommitSha {
		updated = true
		ra.Status.Sync.AppRepoLatestCommitSha = ra.Tmp.PullRequest.HeadCommitSha
	}

	// get ArgoCD Application name
	argocdAppNamespacedName, err := kubernetes.PickNamespacedNameFromObjectStr(ctx, ra.Tmp.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.ApplicationName = argocdAppNamespacedName.Name
	ra.Status.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	if updated {
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo
	} else {
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeWatchingTemplates
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) confirmTemplatesAreUpdated(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (ctrl.Result, error) {
	// confirm
	var updated bool
	if !reflect.DeepEqual(ra.Tmp.Application, ra.Status.ManifestsCache.Application) {
		updated = true
	}
	if !reflect.DeepEqual(ra.Tmp.Manifests, ra.Status.ManifestsCache.Manifests) {
		updated = true
	}

	// get ArgoCD Application name
	argocdAppNamespacedName, err := kubernetes.PickNamespacedNameFromObjectStr(ctx, ra.Tmp.Application)
	if err != nil {
		return ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.ApplicationName = argocdAppNamespacedName.Name
	ra.Status.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	if updated {
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo
	} else {
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo
	}
	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) deployReviewAppManifestsToInfraRepo(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (ctrl.Result, error) {

	// set annotations to Argo CD Application
	argocdAppStr := ra.Tmp.Application
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
	ra.Tmp.ApplicationWithAnnotations = argocdAppStr

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
	ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo
	ra.Status.Sync.InfraRepoLatestCommitSha = gp.LatestCommitSha
	ra.Status.ManifestsCache.Application = ra.Tmp.Application
	ra.Status.ManifestsCache.Manifests = ra.Tmp.Manifests

	return ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) commentToAppRepoPullRequest(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) (ctrl.Result, error) {
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

	// if ArgoCD Application updated, send message to PR of AppRepo
	updated := r.GitRemoteRepoAppService.IsApplicationUpdated(ctx, services.IsApplicationUpdatedParam{
		HashInRA:                ra.Status.Sync.AppRepoLatestCommitSha,
		HashInArgoCDApplication: hashInArgoCDApplication,
	})
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

			// Send Message to AppRepo's PR
			pr := (*gateways.PullRequest)(&ra.Tmp.PullRequest)
			if err := r.GitRemoteRepoAppService.SendMessage(ctx,
				pr, ra.Spec.AppConfig.Message, ra.Spec.AppTarget.Username, gitRemoteRepoCred,
			); err != nil {
				return ctrl.Result{}, err
			}
		}

		// update ReviewApp.Status
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo
		ra.Status.AlreadySentMessage = true
	}

	return ctrl.Result{}, nil
}
