package controllers

import (
	"context"
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
		return nil, ctrl.Result{}, err
	}

	// check PRs specified by spec.appRepo.repository
	pr, raStatus, err := r.PullRequestService.Get(ctx, ra, models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken), datetimeFactoryForRA)
	if err != nil {
		return nil, ctrl.Result{}, err
	}
	ra.Status = dreamkastv1alpha1.ReviewAppStatus(raStatus)

	// template ApplicationTemplate & ManifestsTemplate
	v := models.NewTemplator(ra, pr)

	// get ApplicationTemplate & template to applicationStr
	at, err := r.K8sRepository.GetApplicationTemplate(ctx, ra)
	if err != nil {
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

func (r *ReviewAppReconciler) confirmUpdated(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewAppStatus, ctrl.Result, error) {
	ra := dto.ReviewApp
	raStatus := ra.GetStatus()
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests

	// Is App Repo updated?
	raStatus, updatedAppRepo := raStatus.UpdateStatusOfAppRepo(pr)
	// Is ApplicationTemplate updated?
	raStatus, updatedAt, err := raStatus.UpdateStatusOfApplication(application)
	if err != nil {
		return raStatus, ctrl.Result{}, err
	}
	// Is ManifestsTemplate updated?
	updatedMt := raStatus.WasManifestsUpdated(manifests)

	// update ReviewApp.Status
	if updatedAppRepo || updatedAt || updatedMt {
		raStatus.Sync.Status = dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo
	}
	return raStatus, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) deployReviewAppManifestsToInfraRepo(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewAppStatus, ctrl.Result, error) {
	ra := dto.ReviewApp
	raStatus := ra.GetStatus()
	infraRepoTarget := ra.InfraRepoTarget()
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests

	// set annotations to Argo CD Application
	appWithAnnotations, err := application.SetSomeAnnotations(ra)
	if err != nil {
		return raStatus, ctrl.Result{}, err
	}

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, infraRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
			r.Log.Info(err.Error())
			return raStatus, ctrl.Result{}, nil
		}
		return raStatus, ctrl.Result{}, err
	}
	if err := r.GitCommandRepository.WithCredential(models.NewGitCredential(ra.Spec.InfraTarget.Username, gitRemoteRepoToken)); err != nil {
		return raStatus, ctrl.Result{}, err
	}

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	var localDir models.InfraRepoLocalDir
	if err := backoff.Retry(func() error {
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
		return raStatus, ctrl.Result{}, err
	}

	// update ReviewApp.Status
	raStatus.Sync.Status = dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo
	raStatus.ManifestsCache.Application = string(application)
	raStatus.ManifestsCache.Manifests = manifests

	return raStatus, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) commentToAppRepoPullRequest(ctx context.Context, dto ReviewAppPhaseDTO) (models.ReviewAppStatus, ctrl.Result, error) {
	ra := dto.ReviewApp
	raStatus := ra.GetStatus()
	appTarget := ra.AppRepoTarget()
	pr := dto.PullRequest

	// check appRepoSha from annotations in ArgoCD Application
	application, err := r.K8sRepository.GetArgoCDAppFromReviewAppStatus(ctx, ra.GetStatus())
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			return raStatus, ctrl.Result{}, nil
		}
		return raStatus, ctrl.Result{}, err
	}
	hashInArgoCDApplication, err := application.Annotation(models.AnnotationAppCommitHashForArgoCDApplication)
	if err != nil {
		return raStatus, ctrl.Result{}, err
	}

	// if ArgoCD Application has not been updated, early return.
	if !raStatus.HasApplicationBeenUpdated(hashInArgoCDApplication) {
		return raStatus, ctrl.Result{}, nil
	}

	// send message to PR of AppRepo
	if !ra.HasMessageAlreadyBeenSent() {
		// get gitRemoteRepo credential from Secret
		gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, appTarget)
		if err != nil {
			if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
				r.Log.Info(err.Error())
				return raStatus, ctrl.Result{}, nil
			}
			return raStatus, ctrl.Result{}, err
		}
		if err := r.GitApiRepository.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
			return raStatus, ctrl.Result{}, err
		}
		// Send Message to AppRepo's PR
		if err := r.GitApiRepository.CommentToPullRequest(ctx, pr, ra.Spec.AppConfig.Message); err != nil {
			return raStatus, ctrl.Result{}, err
		}
		// add metrics
		metrics.RequestToGitHubApiCounterVec.WithLabelValues(
			ra.Name,
			ra.Namespace,
			"ReviewApp",
		).Add(1)
	}

	// update ReviewApp.Status
	raStatus.Sync.Status = dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates
	raStatus.Sync.AlreadySentMessage = true

	return raStatus, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, dto ReviewAppPhaseDTO) (ctrl.Result, error) {
	ra := dto.ReviewApp
	raSource := ra.ToReviewAppCR()
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
				r.Recorder.Eventf(raSource, corev1.EventTypeWarning, "preStopJob", "not found JobTemplate %s: %s", jt.Name, err)
			}
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
				r.Recorder.Eventf(raSource, corev1.EventTypeNormal, "finish preStopJob", "preStopJob (%s: %s) is succeeded", models.LabelReviewAppNameForJob, ra.Name)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

finalize:
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8sRepository.GetSecretValue(ctx, ra.Namespace, infraRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := r.GitCommandRepository.WithCredential(models.NewGitCredential(ra.Spec.InfraTarget.Username, gitRemoteRepoToken)); err != nil {
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
	r.removeMetrics(ra)

	return ctrl.Result{}, nil
}
