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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	"github.com/cloudnativedaysjp/reviewapp-operator/mock"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/gitcommand"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/githubapi"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/gateways/kubernetes"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/models"
)

var (
	testLogger = glogr.NewWithOptions(glogr.Options{LogCaller: glogr.None})
	testScheme = runtime.NewScheme()
	testCtx    = context.Background()

	testRaNormal,
	testPrNormal,
	testAtNormal,
	testMtNormal,
	testAppNormal,
	testManifestsNormal,
	testPreStopJtNormal,
	testPreStopJobNormal = testutils.GenerateObjects("testset_normal")

	_,
	_,
	_,
	_,
	testAppNormal_updated,
	testManifestsNormal_updated,
	_,
	_ = testutils.GenerateObjects("testset_normal_updated")
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestReviewAppReconciler_prepare(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder int
		K8sRepository       func() kubernetes.KubernetesIface
	}
	type args struct {
		ra dreamkastv1alpha1.ReviewApp
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
				K8sRepository: func() kubernetes.KubernetesIface {
					m := mock.NewMockKubernetesIface(mockCtrl)
					m.EXPECT().GetPullRequest(testCtx, testPrNormal.Namespace, testPrNormal.Name).
						Return(testPrNormal, nil)
					m.EXPECT().GetApplicationTemplate(testCtx, testRaNormal.Spec.ReviewAppCommonSpec).
						Return(testAtNormal, nil)
					m.EXPECT().GetManifestsTemplate(testCtx, testRaNormal.Spec.ReviewAppCommonSpec).
						Return([]dreamkastv1alpha1.ManifestsTemplate{testMtNormal}, nil)
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
				Log:      testLogger,
				Scheme:   testScheme,
				Recorder: record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8s:      tt.fields.K8sRepository(),
			}
			dto, _, result, err := r.prepare(testCtx, tt.args.ra)
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
		wantRaStatus dreamkastv1alpha1.ReviewAppStatus
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
					ReviewApp: func() dreamkastv1alpha1.ReviewApp {
						m := testRaNormal
						m.Status = dreamkastv1alpha1.ReviewAppStatus{}
						return m
					}(),
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal,
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.ReviewAppStatusSync{
					Status: dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{},
				PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
					Number:           testPrNormal.Spec.Number,
					BaseBranch:       testPrNormal.Status.BaseBranch,
					HeadBranch:       testPrNormal.Status.HeadBranch,
					LatestCommitHash: testPrNormal.Status.LatestCommitHash,
					Title:            testPrNormal.Status.Title,
					Labels:           testPrNormal.Status.Labels,
				},
			},
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
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.ReviewAppStatusSync{
					Status:             dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					AlreadySentMessage: true,
				},
				PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
					Number:           testPrNormal.Spec.Number,
					BaseBranch:       testPrNormal.Status.BaseBranch,
					HeadBranch:       testPrNormal.Status.HeadBranch,
					LatestCommitHash: testPrNormal.Status.LatestCommitHash,
					Title:            testPrNormal.Status.Title,
					Labels:           testPrNormal.Status.Labels,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					ApplicationBase64:    testAppNormal.ToBase64(),
					ManifestsBase64:      testManifestsNormal.ToBase64(),
				},
			},
			wantResult: ctrl.Result{},
		},
		{
			name: "[normal] updated ApplicationTemplate",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp: testutil_withReviewAppStatus(testRaNormal,
						"argocd", "sample-1", testPrNormal.Status.LatestCommitHash),
					PullRequest: testPrNormal,
					Application: testAppNormal_updated, // updated
					Manifests:   testManifestsNormal,
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.ReviewAppStatusSync{
					Status:             dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					AlreadySentMessage: true,
				},
				PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
					Number:           testPrNormal.Spec.Number,
					BaseBranch:       testPrNormal.Status.BaseBranch,
					HeadBranch:       testPrNormal.Status.HeadBranch,
					LatestCommitHash: testPrNormal.Status.LatestCommitHash,
					Title:            testPrNormal.Status.Title,
					Labels:           testPrNormal.Status.Labels,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					ApplicationBase64:    testAppNormal.ToBase64(),
					ManifestsBase64:      testManifestsNormal.ToBase64(),
				},
			},
			wantResult: ctrl.Result{},
		},
		{
			name: "[normal] updated ManifestsTemplate",
			fields: fields{
				NumOfCalledRecorder: 0,
			},
			args: args{
				dto: ReviewAppPhaseDTO{
					ReviewApp: testutil_withReviewAppStatus(testRaNormal,
						"argocd", "sample-1", testPrNormal.Status.LatestCommitHash),
					PullRequest: testPrNormal,
					Application: testAppNormal,
					Manifests:   testManifestsNormal_updated, // updated
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.ReviewAppStatusSync{
					Status:             dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					AlreadySentMessage: true,
				},
				PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
					Number:           testPrNormal.Spec.Number,
					BaseBranch:       testPrNormal.Status.BaseBranch,
					HeadBranch:       testPrNormal.Status.HeadBranch,
					LatestCommitHash: testPrNormal.Status.LatestCommitHash,
					Title:            testPrNormal.Status.Title,
					Labels:           testPrNormal.Status.Labels,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					ApplicationName:      "sample-1",
					ApplicationNamespace: "argocd",
					ApplicationBase64:    testAppNormal.ToBase64(),
					ManifestsBase64:      testManifestsNormal.ToBase64(),
				},
			},
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
			raStatus, err := r.confirmUpdated(testCtx, tt.args.dto, &Result{})
			raStatus.PullRequestCache.SyncedTimestamp = metav1.Time{} // ignore timestamp
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.confirmUpdated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.confirmUpdated() is unexpected:\n%v", diff)
			}
			// TODO
			// if diff := cmp.Diff(result, tt.wantResult); diff != "" {
			// 	t.Errorf("result in ReviewAppReconciler.confirmUpdated() is unexpected:\n%v", diff)
			// }
		})
	}
}

func TestReviewAppReconciler_deployReviewAppManifestsToInfraRepo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder int
		K8sRepository       func() kubernetes.KubernetesIface
		GitApiRepository    func() githubapi.GitApiIface
		GitLocalRepoIface   func() gitcommand.GitLocalRepoIface
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus dreamkastv1alpha1.ReviewAppStatus
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
				Log:          testLogger,
				Scheme:       testScheme,
				Recorder:     record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8s:          tt.fields.K8sRepository(),
				GitApi:       tt.fields.GitApiRepository(),
				GitLocalRepo: tt.fields.GitLocalRepoIface(),
			}
			raStatus, err := r.deployReviewAppManifestsToInfraRepo(testCtx, tt.args.dto, &Result{})
			raStatus.PullRequestCache.SyncedTimestamp = metav1.Time{} // ignore timestamp
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() is unexpected:\n%v", diff)
			}
			// TODO
			//if diff := cmp.Diff(result, tt.wantResult); diff != "" {
			//	t.Errorf("result in ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() is unexpected:\n%v", diff)
			//}
		})
	}
}

func TestReviewAppReconciler_commentToAppRepoPullRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		NumOfCalledRecorder int
		K8sRepository       func() kubernetes.KubernetesIface
		GitApiRepository    func() githubapi.GitApiIface
		GitLocalRepoIface   func() gitcommand.GitLocalRepoIface
	}
	type args struct {
		dto ReviewAppPhaseDTO
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus dreamkastv1alpha1.ReviewAppStatus
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
				Log:          testLogger,
				Scheme:       testScheme,
				Recorder:     record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8s:          tt.fields.K8sRepository(),
				GitApi:       tt.fields.GitApiRepository(),
				GitLocalRepo: tt.fields.GitLocalRepoIface(),
			}
			raStatus, err := r.commentToAppRepoPullRequest(testCtx, tt.args.dto, &Result{})
			raStatus.PullRequestCache.SyncedTimestamp = metav1.Time{} // ignore timestamp
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(raStatus, tt.wantRaStatus); diff != "" {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() is unexpected:\n%v", diff)
			}
			// TODO
			//if diff := cmp.Diff(result, tt.wantResult); diff != "" {
			//	t.Errorf("result in ReviewAppReconciler.commentToAppRepoPullRequest() is unexpected:\n%v", diff)
			//}
		})
	}
}

func TestReviewAppReconciler_reconcileDelete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testSecretToken := "test-token"
	testInfraRepoLatestCommitHash := "12345678"

	type fields struct {
		NumOfCalledRecorder int
		K8sRepository       func() kubernetes.KubernetesIface
		GitLocalRepoIface   func() gitcommand.GitLocalRepoIface
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
				K8sRepository: func() kubernetes.KubernetesIface {
					m := mock.NewMockKubernetesIface(mockCtrl)
					m.EXPECT().GetPreStopJobTemplate(testCtx, testRaNormal).
						Return(testPreStopJtNormal, nil)
					m.EXPECT().CreateJob(testCtx, &testPreStopJobNormal).
						Return(nil)
					m.EXPECT().GetLatestJobFromLabel(testCtx, testPreStopJobNormal.Namespace, dreamkastv1alpha1.LabelReviewAppNameForJob, testRaNormal.Name).
						Return(testutil_withJobStatus(testPreStopJobNormal, true), nil)
					m.EXPECT().GetSecretValue(testCtx, testRaNormal.Namespace, testRaNormal.Spec.InfraTarget).
						Return(testSecretToken, nil)
					m.EXPECT().RemoveFinalizersFromReviewApp(testCtx, testRaNormal, raFinalizer).
						Return(nil)
					return m
				},
				GitLocalRepoIface: func() gitcommand.GitLocalRepoIface {
					// vars
					infraRepoTarget := testRaNormal.Spec.InfraTarget
					localDir := models.NewInfraRepoLocalDir(fmt.Sprintf("/tmp/%s/%s", infraRepoTarget.Organization, infraRepoTarget.Repository)).SetLatestCommitHash(testInfraRepoLatestCommitHash)
					// mock
					m := mock.NewMockGitLocalRepoIface(mockCtrl)
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
				Log:          testLogger,
				Scheme:       testScheme,
				Recorder:     record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8s:          tt.fields.K8sRepository(),
				GitLocalRepo: tt.fields.GitLocalRepoIface(),
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

func testutil_withReviewAppStatus(m dreamkastv1alpha1.ReviewApp, appNamespace, appName, commitHash string) dreamkastv1alpha1.ReviewApp {
	m.Status = dreamkastv1alpha1.ReviewAppStatus{
		Sync: dreamkastv1alpha1.ReviewAppStatusSync{
			Status:             dreamkastv1alpha1.SyncStatusCodeWatchingAppRepoAndTemplates,
			AlreadySentMessage: true,
		},
		PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
			Number:           testPrNormal.Spec.Number,
			BaseBranch:       testPrNormal.Status.BaseBranch,
			HeadBranch:       testPrNormal.Status.HeadBranch,
			LatestCommitHash: commitHash,
			Title:            testPrNormal.Status.Title,
			Labels:           testPrNormal.Status.Labels,
		},
		ManifestsCache: dreamkastv1alpha1.ManifestsCache{
			ApplicationName:      appName,
			ApplicationNamespace: appNamespace,
			ApplicationBase64:    testAppNormal.ToBase64(),
			ManifestsBase64:      testManifestsNormal.ToBase64(),
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
