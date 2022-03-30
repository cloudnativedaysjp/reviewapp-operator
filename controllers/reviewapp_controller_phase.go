package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils/metrics"
)

const (
	annotationAppOrgNameForArgoCDApplication    = "dreamkast.cloudnativedays.jp/app-organization"
	annotationAppRepoNameForArgoCDApplication   = "dreamkast.cloudnativedays.jp/app-repository"
	annotationAppCommitHashForArgoCDApplication = "dreamkast.cloudnativedays.jp/app-commit-hash"

	preStopJobTimeoutSecond = 300
)

var (
	backoffRetryCount = backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
)

type ReviewAppPhaseDTO struct {
	ReviewApp   models.ReviewApp
	PullRequest models.PullRequest
	Application models.Application
	Manifests   models.Manifests
}

func (r *ReviewAppReconciler) prepare(ctx context.Context, ra models.ReviewApp) (*ReviewAppPhaseDTO, ctrl.Result, error) {
	appRepoTarget := ra.AppRepoTarget()

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, appRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
		}
		return nil, ctrl.Result{}, err
	}
	if err := r.GitApiRepository.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
		return nil, ctrl.Result{}, err
	}

	// check PRs specified by spec.appRepo.repository
	pr, err := r.GitApiRepository.GetPullRequest(ctx, appRepoTarget, ra.PrNum())
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	// template ApplicationTemplate & ManifestsTemplate
	v := models.NewTemplator(ra, pr)

	// get ApplicationTemplate & template to applicationStr
	at, err := r.K8sRepository.GetApplicationTemplate(ctx, ra)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
		}
		return nil, ctrl.Result{}, err
	}
	application, err := at.GenerateApplication(pr, v)
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	// get ManifestsTemplate & template to manifestsStr
	ra.GroupVersionKind()
	mts, err := r.K8sRepository.GetManifestsTemplate(ctx, ra)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
		}
		return nil, ctrl.Result{}, err
	}
	var mt models.ManifestsTemplate
	for _, mtOne := range mts {
		mt = mt.AppendOrUpdate(mtOne)
	}
	manifests, err := mt.GenerateManifests(pr, v)
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	return &ReviewAppPhaseDTO{ra, pr, application, manifests}, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) confirmUpdated(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewApp, ctrl.Result, error) {
	ra := dto.ReviewApp
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests

	// Is App Repo updated?
	ra, updatedAppRepo := ra.UpdateStatusOfAppRepo(pr)
	// Is ApplicationTemplate updated?
	ra, updatedAt, err := ra.UpdateStatusOfApplication(application)
	if err != nil {
		return ra, ctrl.Result{}, err
	}
	// Is ManifestsTemplate updated?
	updatedMt := ra.StatusOfManifestsWasUpdated(manifests)

	// update ReviewApp.Status
	if updatedAppRepo || updatedAt || updatedMt {
		ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo
	}
	return ra, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) deployReviewAppManifestsToInfraRepo(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewApp, ctrl.Result, error) {
	ra := dto.ReviewApp
	infraRepoTarget := ra.InfraRepoTarget()
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests

	// set annotations to Argo CD Application
	// TODO: model 化
	appWithAnnotations := application
	appWithAnnotations, err := appWithAnnotations.SetAnnotation(annotationAppOrgNameForArgoCDApplication, ra.Spec.AppTarget.Organization)
	if err != nil {
		return ra, ctrl.Result{}, err
	}
	appWithAnnotations, err = appWithAnnotations.SetAnnotation(annotationAppRepoNameForArgoCDApplication, ra.Spec.AppTarget.Repository)
	if err != nil {
		return ra, ctrl.Result{}, err
	}
	appWithAnnotations, err = appWithAnnotations.SetAnnotation(annotationAppCommitHashForArgoCDApplication, ra.Status.Sync.AppRepoLatestCommitSha)
	if err != nil {
		return ra, ctrl.Result{}, err
	}

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, infraRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) {
			// TODO
			r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
		}
		return ra, ctrl.Result{}, err
	}
	if err := r.GitCommandRepository.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
		return ra, ctrl.Result{}, err
	}

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	var localDir models.InfraRepoLocalDir
	if err := backoff.Retry(
		func() error {
			// clone
			localDir, err = r.GitCommandRepository.ForceClone(ctx, infraRepoTarget)
			if err != nil {
				return err
			}
			// create files
			files := append([]models.File{}, models.NewFileFromApplication(ra, appWithAnnotations, pr, localDir))
			files = append(files, models.NewFilesFromManifests(ra, manifests, pr, localDir)...)
			if err := r.GitCommandRepository.CreateFiles(ctx, localDir, files...); err != nil {
				return err
			}
			// commmit & push
			if _, err := r.GitCommandRepository.CommitAndPush(ctx, localDir, localDir.CommitMsgUpdate(ra)); err != nil {
				return err
			}
			return nil
		}, backoffRetryCount); err != nil {
		return ra, ctrl.Result{}, err
	}

	// update ReviewApp.Status
	ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo
	ra.Status.Sync.InfraRepoLatestCommitSha = localDir.LatestCommitSha()
	ra.Status.ManifestsCache.Application = string(application)
	ra.Status.ManifestsCache.Manifests = manifests

	return ra, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) commentToAppRepoPullRequest(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewApp, ctrl.Result, error) {
	ra := dto.ReviewApp
	appTarget := ra.AppRepoTarget()
	pr := dto.PullRequest

	// check appRepoSha from annotations in ArgoCD Application
	application, err := r.K8sRepository.GetArgoCDAppFromReviewAppStatus(ctx, ra)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			return ra, ctrl.Result{}, nil
		}
		return ra, ctrl.Result{}, err
	}
	hashInArgoCDApplication, err := application.Annotation(annotationAppCommitHashForArgoCDApplication)
	if err != nil {
		return ra, ctrl.Result{}, err
	}

	// if ArgoCD Application has not been updated, early return.
	if !ra.HasApplicationBeenUpdated(hashInArgoCDApplication) {
		return ra, ctrl.Result{}, nil
	}

	// send message to PR of AppRepo
	if !ra.HasMessageAlreadyBeenSent() {
		// get gitRemoteRepo credential from Secret
		gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, appTarget)
		if err != nil {
			if myerrors.IsNotFound(err) {
				// TODO
				r.Log.Info(fmt.Sprintf("Secret %s/%s data[%s] not found", ra.Namespace, ra.Spec.AppTarget.GitSecretRef.Name, ra.Spec.AppTarget.GitSecretRef.Key))
			}
			return ra, ctrl.Result{}, err
		}
		if err := r.GitCommandRepository.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
			return ra, ctrl.Result{}, err
		}
		// Send Message to AppRepo's PR
		if err := r.GitApiRepository.CommentToPullRequest(ctx, pr, ra.Spec.AppConfig.Message); err != nil {
			return ra, ctrl.Result{}, err
		}
	}

	// update ReviewApp.Status
	ra.Status.Sync.Status = dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo
	ra.Status.AlreadySentMessage = true

	return ra, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, dto ReviewAppPhaseDTO) (ctrl.Result, error) {
	ra := dto.ReviewApp
	raSource := ra.ToReviewAppCR()
	appRepoTarget := ra.AppRepoTarget()
	infraRepoTarget := ra.InfraRepoTarget()
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests
	// run preStop Job
	if ra.HavingPreStopJob() {
		// init templator
		v := models.NewTemplator(ra, pr)

		jt, err := r.K8sRepository.GetPreStopJobTemplate(ctx, ra)
		if err != nil {
			if myerrors.IsNotFound(err) {
				r.Log.Info(err.Error())
			}
			r.Recorder.Eventf(raSource, corev1.EventTypeWarning, "preStopJob", "not found JobTemplate %s: %s", jt.Name, err)
			goto finalize
		}

		// get Job Object
		preStopJob, err := jt.GenerateJob(ra, pr, v)
		if err != nil {
			r.Recorder.Eventf(raSource, corev1.EventTypeWarning, "failed to run preStopJob", "cannot unmarshal .spec.template of JobTemplate %s: %s", jt.Name, err)
			goto finalize
		}

		// create job & wait until Job completed on singleflight
		r.Recorder.Eventf(raSource, corev1.EventTypeNormal, "running preStopJob", "running preStopJob (%s: %s)", models.LabelReviewAppNameForJob, ra.Name)
		if err := r.K8sRepository.CreateJob(ctx, preStopJob); err != nil {
			r.Recorder.Eventf(raSource, corev1.EventTypeWarning, "failed to run preStopJob", "cannot create Job (%s: %s): %s", models.LabelReviewAppNameForJob, ra.Name, err)
			goto finalize
		}
		timeout := time.Now().Add(preStopJobTimeoutSecond * time.Second)
		for {
			if time.Since(timeout) >= 0 {
				r.Recorder.Eventf(raSource, corev1.EventTypeWarning, "preStopJob is timeout", "preStopJob (%s: %s) is timeout (%ds)", models.LabelReviewAppNameForJob, ra.Name, preStopJobTimeoutSecond)
				goto finalize
			}
			appliedPreStopJob, err := r.K8sRepository.GetLatestJobFromLabel(ctx, preStopJob.Namespace, models.LabelReviewAppNameForJob, ra.Name)
			if err != nil {
				return ctrl.Result{}, err
			}
			if appliedPreStopJob.Status.Succeeded != 0 {
				r.Recorder.Eventf(raSource, corev1.EventTypeNormal, "finish preStopJob", "preStopJob (%s: %s) is succeeded", models.LabelReviewAppNameForJob, ra.Name, preStopJobTimeoutSecond)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

finalize:
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, appRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
		}
		return ctrl.Result{}, err
	}
	if err := r.GitApiRepository.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
		return ctrl.Result{}, err
	}

	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	var localDir models.InfraRepoLocalDir
	if err := backoff.Retry(
		func() error {
			// clone
			localDir, err = r.GitCommandRepository.ForceClone(ctx, infraRepoTarget)
			if err != nil {
				return err
			}
			// delete files
			files := append([]models.File{}, models.NewFileFromApplication(ra, application, pr, localDir))
			files = append(files, models.NewFilesFromManifests(ra, manifests, pr, localDir)...)
			if err := r.GitCommandRepository.DeleteFiles(ctx, localDir, files...); err != nil {
				return err
			}
			// commmit & push
			if _, err := r.GitCommandRepository.CommitAndPush(ctx, localDir, localDir.CommitMsgDeletion(ra)); err != nil {
				return err
			}
			return nil
		}, backoffRetryCount); err != nil {
		return ctrl.Result{}, err
	}

	// Remove Finalizers
	if err := r.K8sRepository.RemoveFinalizersFromReviewApp(ctx, ra, finalizer); err != nil {
		return ctrl.Result{}, err
	}

	// remove metrics
	metrics.RemoveMetrics(ra)

	return ctrl.Result{}, nil
}
