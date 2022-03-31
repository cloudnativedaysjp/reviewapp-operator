//go:build !integration_test
// +build !integration_test

package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/glogr"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cloudnativedaysjp/reviewapp-operator/controllers/testutils"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/mock"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
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
	testPreStopJobNormal = testutils.GenerateObjects("testset_normal")
	testPrNormal = models.PullRequest{
		Organization:  testRaNormal.AppRepoTarget().Organization,
		Repository:    testRaNormal.AppRepoTarget().Repository,
		Branch:        "test",
		Number:        testRaNormal.PrNum(),
		HeadCommitSha: "1234567",
		Title:         "TEST",
		Labels:        []string{},
	}
)

func TestReviewAppReconciler_prepare(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testSecretToken := "test-token"

	type fields struct {
		NumOfCalledRecorder  int
		K8sRepository        func() repositories.KubernetesRepository
		GitApiRepository     func() repositories.GitAPI
		GitCommandRepository func() repositories.GitCommand
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
			name: "[testset_normal] normal",
			fields: fields{
				NumOfCalledRecorder: 0,
				K8sRepository: func() repositories.KubernetesRepository {
					m := mock.NewMockKubernetesRepository(mockCtrl)
					m.EXPECT().GetSecretValue(testCtx, testRaNormal.Namespace, testRaNormal.AppRepoTarget()).Return(testSecretToken, nil)
					m.EXPECT().GetApplicationTemplate(testCtx, testRaNormal).Return(testAtNormal, nil)
					m.EXPECT().GetManifestsTemplate(testCtx, testRaNormal).Return([]models.ManifestsTemplate{testMtNormal}, nil)
					return m
				},
				GitApiRepository: func() repositories.GitAPI {
					m := mock.NewMockGitAPI(mockCtrl)
					m.EXPECT().WithCredential(models.NewGitCredential(testRaNormal.AppRepoTarget().Username, testSecretToken)).Return(nil)
					m.EXPECT().GetPullRequest(testCtx, testRaNormal.AppRepoTarget(), testRaNormal.PrNum()).Return(testPrNormal, nil)
					return m
				},
				GitCommandRepository: func() repositories.GitCommand { return mock.NewMockGitCommand(mockCtrl) }, // unused
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
			r := &ReviewAppReconciler{
				Log:                  testLogger,
				Scheme:               testScheme,
				Recorder:             record.NewFakeRecorder(tt.fields.NumOfCalledRecorder),
				K8sRepository:        tt.fields.K8sRepository(),
				GitApiRepository:     tt.fields.GitApiRepository(),
				GitCommandRepository: tt.fields.GitCommandRepository(),
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
