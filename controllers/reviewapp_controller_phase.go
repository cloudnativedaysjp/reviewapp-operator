package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/metrics"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/template"
)

const (
	preStopJobTimeoutSecond = 300
)

var (
	backoffRetryCount = backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
)

type ReviewAppPhaseDTO struct {
	// immutable
	ReviewApp   dreamkastv1alpha1.ReviewApp
	PullRequest dreamkastv1alpha1.PullRequest
	Application dreamkastv1alpha1.Application
	Manifests   dreamkastv1alpha1.Manifests
}

type Result struct {
	Result           ctrl.Result
	StatusWasUpdated bool
}

func (r *ReviewAppReconciler) prepare(ctx context.Context, ra dreamkastv1alpha1.ReviewApp,
) (*ReviewAppPhaseDTO, bool, ctrl.Result, error) {
	// check PRs specified by spec.appRepo.repository
	// If object does not exist, using cache from RA.status
	usingPrCache := false
	pr, err := r.K8s.GetPullRequest(ctx, ra.Spec.PullRequest.Namespace, ra.Spec.PullRequest.Name)
	if err != nil {
		if ra.Status.PullRequestCache.IsEmpty() {
			r.Log.Info(err.Error())
			return nil, usingPrCache, ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		usingPrCache = true
		pr = dreamkastv1alpha1.PullRequest{
			Spec: dreamkastv1alpha1.PullRequestSpec{
				AppTarget: ra.Spec.AppTarget,
				Number:    ra.Status.PullRequestCache.Number,
			},
			Status: dreamkastv1alpha1.PullRequestStatus{
				HeadBranch:       ra.Status.PullRequestCache.HeadBranch,
				BaseBranch:       ra.Status.PullRequestCache.BaseBranch,
				LatestCommitHash: ra.Status.PullRequestCache.LatestCommitHash,
				Title:            ra.Status.PullRequestCache.Title,
				Labels:           ra.Status.PullRequestCache.Labels,
			},
		}
	}

	// template ApplicationTemplate & ManifestsTemplate
	v := template.NewTemplator(ra.Spec.ReviewAppCommonSpec, pr)

	// get ApplicationTemplate & template to applicationStr
	// If object does not exist, using cache from RA.status
	var application dreamkastv1alpha1.Application
	if err := func() error {
		at, err := r.K8s.GetApplicationTemplate(ctx, ra.Spec.ReviewAppCommonSpec)
		if err != nil {
			r.Log.Info(fmt.Sprintf("%v: using cache", err))
			return err
		}
		application, err = v.Application(at, pr)
		if err != nil {
			// TODO: logging
			return err
		}
		return nil
	}(); err != nil {
		if ra.Status.ManifestsCache.ApplicationBase64 == "" {
			return nil, usingPrCache, ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		application, err = ra.Status.ManifestsCache.ApplicationBase64.Decode()
		if err != nil {
			return nil, usingPrCache, ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	// get ManifestsTemplate & template to manifestsStr
	// If object does not exist, using cache from RA.status
	var manifests dreamkastv1alpha1.Manifests
	if err := func() error {
		mts, err := r.K8s.GetManifestsTemplate(ctx, ra.Spec.ReviewAppCommonSpec)
		if err != nil {
			r.Log.Info(fmt.Sprintf("%v: using cache", err))
			return err
		}
		var mt dreamkastv1alpha1.ManifestsTemplate
		for _, mtOne := range mts {
			mt = mt.AppendOrUpdate(mtOne)
		}
		manifests, err = v.Manifests(mt, pr)
		if err != nil {
			// TODO: logging
			return err
		}
		return nil
	}(); err != nil {
		if len(ra.Status.ManifestsCache.ManifestsBase64) == 0 {
			return nil, usingPrCache, ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		manifests, err = ra.Status.ManifestsCache.ManifestsBase64.Decode()
		if err != nil {
			return nil, usingPrCache, ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
	}

	return &ReviewAppPhaseDTO{ra, pr, application, manifests}, usingPrCache, ctrl.Result{}, nil
}

func (r *ReviewAppReconciler) confirmUpdated(ctx context.Context,
	dto ReviewAppPhaseDTO, res *Result,
) (dreamkastv1alpha1.ReviewAppStatus, error) {
	raStatus := dto.ReviewApp.Status

	updatedAppRepo := raStatus.HasPullRequestBeenUpdated(dto.PullRequest.Status.LatestCommitHash)
	updatedAt := raStatus.HasApplicationTemplateBeenUpdated(dto.Application)
	updatedMt := raStatus.HaveManifestsTemplateBeenUpdated(dto.Manifests)

	// update ReviewApp.Status
	if updatedAppRepo || updatedAt || updatedMt {
		raStatus.Sync.Status = dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo
		raStatus.PullRequestCache = dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
			Number:           dto.PullRequest.Spec.Number,
			BaseBranch:       dto.PullRequest.Status.BaseBranch,
			HeadBranch:       dto.PullRequest.Status.HeadBranch,
			LatestCommitHash: dto.PullRequest.Status.LatestCommitHash,
			Title:            dto.PullRequest.Status.Title,
			Labels:           dto.PullRequest.Status.Labels,
			SyncedTimestamp:  metav1.Now(),
		}
		res.StatusWasUpdated = true
	}
	return raStatus, nil
}

func (r *ReviewAppReconciler) deployReviewAppManifestsToInfraRepo(ctx context.Context,
	dto ReviewAppPhaseDTO, res *Result,
) (dreamkastv1alpha1.ReviewAppStatus, error) {
	ra := dto.ReviewApp
	raStatus := ra.Status
	infraRepoTarget := ra.Spec.InfraTarget
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests

	applicationNN, err := application.NamespacedName()
	if err != nil {
		return raStatus, err
	}

	// set annotations to Argo CD Application
	appWithAnnotations, err := application.SetSomeAnnotations(ra)
	if err != nil {
		return raStatus, err
	}

	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8s.GetSecretValue(ctx, ra.Namespace, infraRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
			r.Log.Info(err.Error())
			res.Result.RequeueAfter = 10 * time.Second
			return raStatus, nil
		}
		return raStatus, err
	}
	if err := r.GitLocalRepo.WithCredential(models.NewGitCredential(ra.Spec.InfraTarget.Username, gitRemoteRepoToken)); err != nil {
		return raStatus, err
	}

	// update Application & other manifests from ApplicationTemplate & ManifestsTemplate to InfraRepo
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	var localDir models.InfraRepoLocalDir
	if err := backoff.Retry(func() error {
		// clone
		localDir, err = r.GitLocalRepo.ForceClone(ctx, infraRepoTarget)
		if err != nil {
			return err
		}
		// create files
		files := append([]models.File{}, models.NewFileFromApplication(ra, appWithAnnotations, pr, localDir))
		files = append(files, models.NewFilesFromManifests(ra, manifests, pr, localDir)...)
		if err := r.GitLocalRepo.CreateFiles(ctx, localDir, files...); err != nil {
			return err
		}
		// commmit & push
		if _, err := r.GitLocalRepo.CommitAndPush(ctx, localDir, localDir.CommitMsgUpdate(ra)); err != nil {
			return err
		}
		return nil
	}, backoffRetryCount); err != nil {
		return raStatus, err
	}

	// update ReviewApp.Status
	raStatus.Sync.Status = dreamkastv1alpha1.SyncStatusCodeUpdatedInfraRepo
	raStatus.ManifestsCache.ApplicationName = applicationNN.Name
	raStatus.ManifestsCache.ApplicationNamespace = applicationNN.Namespace
	raStatus.ManifestsCache.ApplicationBase64 = application.ToBase64()
	raStatus.ManifestsCache.ManifestsBase64 = manifests.ToBase64()
	res.StatusWasUpdated = true

	return raStatus, nil
}

func (r *ReviewAppReconciler) commentToAppRepoPullRequest(ctx context.Context,
	dto ReviewAppPhaseDTO, res *Result,
) (dreamkastv1alpha1.ReviewAppStatus, error) {
	ra := dto.ReviewApp
	raStatus := ra.Status
	appTarget := ra.Spec.AppTarget
	pr := dto.PullRequest

	// check appRepoSha from annotations in ArgoCD Application
	application, err := r.K8s.GetArgoCDAppFromReviewAppStatus(ctx, raStatus)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			res.Result.RequeueAfter = 10 * time.Second
			return raStatus, nil
		}
		return raStatus, err
	}
	hashInArgoCDApplication, err := application.Annotation(dreamkastv1alpha1.AnnotationAppCommitHashForArgoCDApplication)
	if err != nil {
		return raStatus, err
	}

	// if ArgoCD Application has not been updated, early return.
	if !raStatus.HasArgoCDApplicationBeenUpdated(hashInArgoCDApplication) {
		res.Result.RequeueAfter = 10 * time.Second
		return raStatus, nil
	}

	// send message to PR of AppRepo
	if !ra.HasMessageAlreadyBeenSent() {
		// get gitRemoteRepo credential from Secret
		gitRemoteRepoToken, err := r.K8s.GetSecretValue(ctx, ra.Namespace, appTarget)
		if err != nil {
			if myerrors.IsNotFound(err) || myerrors.IsKeyMissing(err) {
				r.Log.Info(err.Error())
				res.Result.RequeueAfter = 10 * time.Second
				return raStatus, nil
			}
			return raStatus, err
		}
		if err := r.GitApi.WithCredential(models.NewGitCredential(ra.Spec.AppTarget.Username, gitRemoteRepoToken)); err != nil {
			return raStatus, err
		}
		// Send Message to AppRepo's PR
		if err := r.GitApi.CommentToPullRequest(ctx, pr, ra.Spec.AppConfig.Message); err != nil {
			return raStatus, err
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
	res.StatusWasUpdated = true

	return raStatus, nil
}

func (r *ReviewAppReconciler) reconcileDelete(ctx context.Context, dto ReviewAppPhaseDTO) (ctrl.Result, error) {
	ra := dto.ReviewApp
	infraRepoTarget := ra.Spec.InfraTarget
	pr := dto.PullRequest
	application := dto.Application
	manifests := dto.Manifests
	// run preStop Job
	if ra.Spec.ReviewAppCommonSpec.HavePreStopJob() {
		// init templator
		v := template.NewTemplator(ra.Spec.ReviewAppCommonSpec, pr)

		jt, err := r.K8s.GetPreStopJobTemplate(ctx, ra)
		if err != nil {
			if myerrors.IsNotFound(err) {
				r.Log.Info(err.Error())
				r.Recorder.Eventf(&ra, corev1.EventTypeWarning, "preStopJob", "not found JobTemplate %s: %s", jt.Name, err)
			}
			goto finalize
		}

		// get Job Object
		preStopJob, err := v.Job(jt, ra, pr)
		if err != nil {
			r.Recorder.Eventf(&ra, corev1.EventTypeWarning, "failed to run preStopJob", "cannot unmarshal .spec.template of JobTemplate %s: %s", jt.Name, err)
			goto finalize
		}

		// create job & wait until Job completed on singleflight
		r.Recorder.Eventf(&ra, corev1.EventTypeNormal, "running preStopJob", "running preStopJob (%s: %s)", dreamkastv1alpha1.LabelReviewAppNameForJob, ra.Name)
		if err := r.K8s.CreateJob(ctx, preStopJob); err != nil {
			r.Recorder.Eventf(&ra, corev1.EventTypeWarning, "failed to run preStopJob", "cannot create Job (%s: %s): %s", dreamkastv1alpha1.LabelReviewAppNameForJob, ra.Name, err)
			goto finalize
		}
		timeout := time.Now().Add(preStopJobTimeoutSecond * time.Second)
		for {
			if time.Since(timeout) >= 0 {
				r.Recorder.Eventf(&ra, corev1.EventTypeWarning, "preStopJob is timeout", "preStopJob (%s: %s) is timeout (%ds)", dreamkastv1alpha1.LabelReviewAppNameForJob, ra.Name, preStopJobTimeoutSecond)
				goto finalize
			}
			appliedPreStopJob, err := r.K8s.GetLatestJobFromLabel(ctx, preStopJob.Namespace, dreamkastv1alpha1.LabelReviewAppNameForJob, ra.Name)
			if err != nil {
				return ctrl.Result{}, err
			}
			if appliedPreStopJob.Status.Succeeded != 0 {
				r.Recorder.Eventf(&ra, corev1.EventTypeNormal, "finish preStopJob", "preStopJob (%s: %s) is succeeded", dreamkastv1alpha1.LabelReviewAppNameForJob, ra.Name)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

finalize:
	// get gitRemoteRepo credential from Secret
	gitRemoteRepoToken, err := r.K8s.GetSecretValue(ctx, ra.Namespace, infraRepoTarget)
	if err != nil {
		if myerrors.IsNotFound(err) {
			r.Log.Info(err.Error())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if err := r.GitLocalRepo.WithCredential(models.NewGitCredential(ra.Spec.InfraTarget.Username, gitRemoteRepoToken)); err != nil {
		return ctrl.Result{}, err
	}

	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	var localDir models.InfraRepoLocalDir
	if err := backoff.Retry(
		func() error {
			// clone
			localDir, err = r.GitLocalRepo.ForceClone(ctx, infraRepoTarget)
			if err != nil {
				return err
			}
			// delete files
			files := append([]models.File{}, models.NewFileFromApplication(ra, application, pr, localDir))
			files = append(files, models.NewFilesFromManifests(ra, manifests, pr, localDir)...)
			if err := r.GitLocalRepo.DeleteFiles(ctx, localDir, files...); err != nil {
				return err
			}
			// commmit & push
			if _, err := r.GitLocalRepo.CommitAndPush(ctx, localDir, localDir.CommitMsgDeletion(ra)); err != nil {
				return err
			}
			return nil
		}, backoffRetryCount); err != nil {
		return ctrl.Result{}, err
	}

	// Remove Finalizers
	if err := r.K8s.RemoveFinalizersFromReviewApp(ctx, ra, raFinalizer); err != nil {
		return ctrl.Result{}, err
	}

	// remove metrics
	r.removeMetrics(ra)

	return ctrl.Result{}, nil
}
