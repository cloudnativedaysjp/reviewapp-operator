//go:build !integration_test
// +build !integration_test

package controllers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-logr/glogr"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/mock"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/services"
	"github.com/cloudnativedaysjp/reviewapp-operator/utils"
)

var (
	testLogger = glogr.NewWithOptions(glogr.Options{LogCaller: glogr.None})
	testScheme = runtime.NewScheme()
	testCtx    = context.Background()

	testRaNormal,
	testAtNormal,
	testMtNormal,
	testAppNormal,
	testManifestsNormal,
	testPreStopJtNormal,
	testPreStopJobNormal = testutils.GenerateObjects("testset_normal")
	testPrNormal = models.PullRequest{
		Organization:     testRaNormal.AppRepoTarget().Organization,
		Repository:       testRaNormal.AppRepoTarget().Repository,
		Branch:           "test",
		Number:           testRaNormal.PrNum(),
		LatestCommitHash: "testset_normal",
		Title:            "TEST",
		Labels:           []string{},
	}

	_,
	_,
	_,
	testAppNormal_updated,
	testManifestsNormal_updated,
	_,
	_ = testutils.GenerateObjects("testset_normal_updated")
)

func TestMain(m *testing.M) {
	datetimeFactoryForRA = utils.NewDatetimeFactory()
	code := m.Run()
	os.Exit(code)
}

func TestReviewAppReconciler_prepare(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testSecretToken := "test-token"

	type fields struct {
		NumOfCalledRecorder int
		K8sRepository       func() repositories.KubernetesRepository
		PullRequestService  func() services.PullRequestServiceIface
	}
	type args struct {
		ra models.ReviewApp
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantDTO    *ReviewAppPhaseDTO
		wantResult ctrl.Result
		wantErr    bool
	}{
		{
			name: "testset_normal",
			fields: fields{
				NumOfCalledRecorder: 0,
				K8sRepository: func() repositories.KubernetesRepository {
					m := mock.NewMockKubernetesRepository(mockCtrl)
					m.EXPECT().GetSecretValue(testCtx, testRaNormal.Namespace, testRaNormal.AppRepoTarget()).
						Return(testSecretToken, nil)
					m.EXPECT().GetApplicationTemplate(testCtx, testRaNormal).
						Return(testAtNormal, nil)
					m.EXPECT().GetManifestsTemplate(testCtx, testRaNormal).
						Return([]models.ManifestsTemplate{testMtNormal}, nil)
					return m
				},
				PullRequestService: func() services.PullRequestServiceIface {
					m := mock.NewMockPullRequestServiceIface(mockCtrl)
					m.EXPECT().Get(testCtx, testRaNormal, models.NewGitCredential(testRaNormal.AppRepoTarget().Username, testSecretToken), datetimeFactoryForRA).
						Return(testPrNormal, models.ReviewAppStatus(testRaNormal.Status), nil)
					return m
				},
			},
			args: args{ra: testRaNormal},
			wantDTO: &ReviewAppPhaseDTO{
				ReviewApp:   testRaNormal,
				PullRequest: testPrNormal,
				Application: testAppNormal,
				Manifests:   testManifestsNormal,
			},
			wantResult: ctrl.Result{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &ReviewAppReconciler{
				Log:                testLogger,
				Scheme:             testScheme,
				Recorder:           record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8sRepository:      tt.fields.K8sRepository(),
				PullRequestService: tt.fields.PullRequestService(),
			}
			dto, result, err := r.prepare(testCtx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(dto, tt.wantDTO); diff != "" {
				t.Errorf("dto in ReviewAppReconciler.prepare() is unexpected:\n%v", diff)
			}
			if diff := cmp.Diff(result, tt.wantResult); diff != "" {
				t.Errorf("result in ReviewAppReconciler.prepare() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_confirmUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder int
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus models.ReviewAppStatus
		wantResult   ctrl.Result
		wantErr      bool
	}{
		{
			name: "[normal] first time",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp: func() models.ReviewApp {
						m := testRaNormal
						m.Status = dreamkastv1alpha1.ReviewAppStatus{}
						return m
					}(),
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal,
				},
			},
			wantRaStatus: models.ReviewAppStatus(dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status:               dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					SyncedPullRequest: dreamkastv1alpha1.ReviewAppStatusSyncedPullRequest{
						Branch:           testPrNormal.Branch,
						LatestCommitHash: testPrNormal.LatestCommitHash,
						Title:            testPrNormal.Title,
						Labels:           testPrNormal.Labels,
					},
				},
			}),
			wantResult: ctrl.Result{},
		},
		{
			name: "[normal] updated commitHash",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp:   testutil_withReviewAppStatus(testRaNormal, "argocd", "sample-1", "updated-commit-hash"), // updated
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal,
				},
			},
			wantRaStatus: models.ReviewAppStatus(dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status: dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					SyncedPullRequest: dreamkastv1alpha1.ReviewAppStatusSyncedPullRequest{
						Branch:           testPrNormal.Branch,
						LatestCommitHash: testPrNormal.LatestCommitHash,
						Title:            testPrNormal.Title,
						Labels:           testPrNormal.Labels,
					},
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					AlreadySentMessage:   true,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					Application: string(testAppNormal),
					Manifests:   testManifestsNormal,
				},
			}),
			wantResult: ctrl.Result{},
		},
		{
			name: "[normal] updated ApplicationTemplate",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp:   testutil_withReviewAppStatus(testRaNormal, "argocd", "sample-1", testPrNormal.LatestCommitHash),
					PullRequest: testPrNormal,
					Application: testAppNormal_updated, // updated
					Manifests:   testManifestsNormal,
				},
			},
			wantRaStatus: models.ReviewAppStatus(dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status: dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					SyncedPullRequest: dreamkastv1alpha1.ReviewAppStatusSyncedPullRequest{
						Branch:           testPrNormal.Branch,
						LatestCommitHash: testPrNormal.LatestCommitHash,
						Title:            testPrNormal.Title,
						Labels:           testPrNormal.Labels,
					},
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					AlreadySentMessage:   true,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					Application: string(testAppNormal),
					Manifests:   testManifestsNormal,
				},
			}),
			wantResult: ctrl.Result{},
		},
		{
			name: "[normal] updated ManifestsTemplate",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp:   testutil_withReviewAppStatus(testRaNormal, "argocd", "sample-1", testPrNormal.LatestCommitHash),
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal_updated, // updated
				},
			},
			wantRaStatus: models.ReviewAppStatus(dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status: dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					SyncedPullRequest: dreamkastv1alpha1.ReviewAppStatusSyncedPullRequest{
						Branch:           testPrNormal.Branch,
						LatestCommitHash: testPrNormal.LatestCommitHash,
						Title:            testPrNormal.Title,
						Labels:           testPrNormal.Labels,
					},
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					AlreadySentMessage:   true,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					Application: string(testAppNormal),
					Manifests:   testManifestsNormal,
				},
			}),
			wantResult: ctrl.Result{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &ReviewAppReconciler{
				Log:      testLogger,
				Scheme:   testScheme,
				Recorder: record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
			}
			raStatus, result, err := r.confirmUpdated(testCtx, tt.args.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.confirmUpdated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.confirmUpdated() is unexpected:\n%v", diff)
			}
			if diff := cmp.Diff(result, tt.wantResult); diff != "" {
				t.Errorf("result in ReviewAppReconciler.confirmUpdated() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_deployReviewAppManifestsToInfraRepo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder  int
		K8sRepository        func() repositories.KubernetesRepository
		GitApiRepository     func() repositories.GitAPI
		GitCommandRepository func() repositories.GitCommand
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus models.ReviewAppStatus
		wantResult   ctrl.Result
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &ReviewAppReconciler{
				Log:                  testLogger,
				Scheme:               testScheme,
				Recorder:             record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8sRepository:        tt.fields.K8sRepository(),
				GitApiRepository:     tt.fields.GitApiRepository(),
				GitCommandRepository: tt.fields.GitCommandRepository(),
			}
			raStatus, result, err := r.deployReviewAppManifestsToInfraRepo(testCtx, tt.args.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() is unexpected:\n%v", diff)
			}
			if diff := cmp.Diff(result, tt.wantResult); diff != "" {
				t.Errorf("result in ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_commentToAppRepoPullRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder  int
		K8sRepository        func() repositories.KubernetesRepository
		GitApiRepository     func() repositories.GitAPI
		GitCommandRepository func() repositories.GitCommand
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus models.ReviewAppStatus
		wantResult   ctrl.Result
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &ReviewAppReconciler{
				Log:                  testLogger,
				Scheme:               testScheme,
				Recorder:             record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8sRepository:        tt.fields.K8sRepository(),
				GitApiRepository:     tt.fields.GitApiRepository(),
				GitCommandRepository: tt.fields.GitCommandRepository(),
			}
			raStatus, result, err := r.commentToAppRepoPullRequest(testCtx, tt.args.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() is unexpected:\n%v", diff)
			}
			if diff := cmp.Diff(result, tt.wantResult); diff != "" {
				t.Errorf("result in ReviewAppReconciler.commentToAppRepoPullRequest() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_reconcileDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testSecretToken := "test-token"
	testInfraRepoLatestCommitHash := "12345678"

	type fields struct {
		NumOfCalledRecorder  int
		K8sRepository        func() repositories.KubernetesRepository
		GitCommandRepository func() repositories.GitCommand
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		{
			name: "[normal] having preStopJob",
			fields: fields{
				NumOfCalledRecorder: 2,
				K8sRepository: func() repositories.KubernetesRepository {
					m := mock.NewMockKubernetesRepository(mockCtrl)
					m.EXPECT().GetPreStopJobTemplate(testCtx, testRaNormal).
						Return(testPreStopJtNormal, nil)
					m.EXPECT().CreateJob(testCtx, &testPreStopJobNormal).
						Return(nil)
					m.EXPECT().GetLatestJobFromLabel(testCtx, testPreStopJobNormal.Namespace, models.LabelReviewAppNameForJob, testRaNormal.Name).
						Return(testutil_withJobStatus(testPreStopJobNormal, true), nil)
					m.EXPECT().GetSecretValue(testCtx, testRaNormal.Namespace, testRaNormal.InfraRepoTarget()).
						Return(testSecretToken, nil)
					m.EXPECT().RemoveFinalizersFromReviewApp(testCtx, testRaNormal, finalizer).
						Return(nil)
					return m
				},
				GitCommandRepository: func() repositories.GitCommand {
					// vars
					infraRepoTarget := testRaNormal.InfraRepoTarget()
					localDir := models.NewInfraRepoLocal(fmt.Sprintf("/tmp/%s/%s", infraRepoTarget.Organization, infraRepoTarget.Repository)).SetLatestCommitHash(testInfraRepoLatestCommitHash)
					// mock
					m := mock.NewMockGitCommand(mockCtrl)
					m.EXPECT().WithCredential(models.NewGitCredential(infraRepoTarget.Username, testSecretToken)).
						Return(nil)
					m.EXPECT().ForceClone(testCtx, infraRepoTarget).
						Return(localDir, nil)
					// DeleteFiles の引数は順不同なので gomock.Any を利用
					m.EXPECT().DeleteFiles(testCtx, localDir, gomock.Any(), gomock.Any(), gomock.Any()).
						Return(nil)
					m.EXPECT().CommitAndPush(testCtx, localDir, localDir.CommitMsgDeletion(testRaNormal))
					return m
				},
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp:   testRaNormal,
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := &ReviewAppReconciler{
				Log:                  testLogger,
				Scheme:               testScheme,
				Recorder:             record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8sRepository:        tt.fields.K8sRepository(),
				GitCommandRepository: tt.fields.GitCommandRepository(),
			}
			result, err := r.reconcileDelete(testCtx, tt.args.dto)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.reconcileDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(result, tt.want); diff != "" {
				t.Errorf("result in ReviewAppReconciler.reconcileDelete() is unexpected:\n%v", diff)
			}
		})
	}
}

func testutil_withReviewAppStatus(m models.ReviewApp, appNamespace, appName, commitHash string) models.ReviewApp {
	m.Status = dreamkastv1alpha1.ReviewAppStatus{
		Sync: dreamkastv1alpha1.SyncStatus{
			Status: dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates,
			SyncedPullRequest: dreamkastv1alpha1.ReviewAppStatusSyncedPullRequest{
				Branch:           testPrNormal.Branch,
				LatestCommitHash: commitHash,
				Title:            testPrNormal.Title,
				Labels:           testPrNormal.Labels,
			},
			ApplicationName:      appName,
			ApplicationNamespace: appNamespace,
			AlreadySentMessage:   true,
		},
		ManifestsCache: dreamkastv1alpha1.ManifestsCache{
			Application: string(testAppNormal),
			Manifests:   testManifestsNormal,
		},
	}
	return m
}

func testutil_withJobStatus(job batchv1.Job, succeeded bool) *batchv1.Job {
	if succeeded {
		job.Status.Succeeded = 1
	}
	return &job
}
