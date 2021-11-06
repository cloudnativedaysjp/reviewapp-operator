//+build unit_test

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/mock"
	"github.com/cloudnativedaysjp/reviewapp-operator/services"
	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	logger = glogr.NewWithOptions(glogr.Options{LogCaller: glogr.None})
	scheme = runtime.NewScheme()
)

func TestReviewAppReconciler_prepare(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testRaNamespace := "test-ra-namespace"
	testRaName := "test-ra-name"
	testAppTargetOrg := "testorg"
	testAppTargetRepo := "test-apprepo"
	testAppTargetPrNum := 1024
	testAppTargetUsername := "apptarget-username"
	testAppTargetToken := "apptarget-token"
	testAppTargetBranch := "apptarget-branch"
	testAppTargetHeadCommitSha := "1234567"
	testInfraTargetOrg := "testorg"
	testInfraTargetRepo := "test-infrarepo"
	testSecretName := "test-secret"
	testSecretKey := "secret-key"
	testAtNamespace := "namespace-at"
	testAtName := "test-at"
	testAtContent := newApplication_forPhaseTest("argocd", fmt.Sprintf("%s-%d", testAppTargetRepo, testAppTargetPrNum))
	testMtNamespace := "namespace-mt"
	testMtName := "test-mt"
	testMtContents := map[string]string{
		"manifest01.yaml": fmt.Sprintf(`kind: Namespace
metadata:
  name: %s-%d`, testAppTargetRepo, testAppTargetPrNum),
		"manifest02.yaml": fmt.Sprintf(`kind: ConfigMap
metadata:
  namespace: %s-%d 
  name: %s`, testAppTargetRepo, testAppTargetPrNum, testAppTargetOrg),
	}
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dreamkastv1alpha1.AddToScheme(scheme))

	type fields struct {
		Client                    client.Client
		Log                       logr.Logger
		Scheme                    *runtime.Scheme
		GitRemoteRepoAppService   func() *services.GitRemoteRepoAppService
		GitRemoteRepoInfraService func() *services.GitRemoteRepoInfraService
	}
	type args struct {
		ctx context.Context
		ra  *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantTmp dreamkastv1alpha1.ReviewAppTmp
		wantErr bool
	}{
		{
			name: "normal: test",
			fields: fields{
				Client: fake.NewClientBuilder().WithScheme(scheme).
					WithObjects(
						&corev1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: testRaNamespace,
								Name:      testSecretName,
							},
							Data: map[string][]byte{
								testSecretKey: []byte(testAppTargetToken),
							},
						},
						&dreamkastv1alpha1.ApplicationTemplate{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: testAtNamespace,
								Name:      testAtName,
							},
							Spec: dreamkastv1alpha1.ApplicationTemplateSpec{
								StableTemplate:    testAtContent,
								CandidateTemplate: testAtContent,
							},
						},
						&dreamkastv1alpha1.ManifestsTemplate{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: testMtNamespace,
								Name:      testMtName,
							},
							Spec: dreamkastv1alpha1.ManifestsTemplateSpec{
								StableData:    testMtContents,
								CandidateData: testMtContents,
							},
						},
					).Build(),
				Log:    logger,
				Scheme: scheme,
				GitRemoteRepoAppService: func() *services.GitRemoteRepoAppService {
					c := mock.NewMockGitHubIFace(mockCtrl)
					c.EXPECT().WithCredential(testAppTargetUsername, testAppTargetToken).Return(nil)
					c.EXPECT().GetPullRequest(context.Background(), testAppTargetOrg, testAppTargetRepo, testAppTargetPrNum).Return(&gateways.PullRequest{
						Organization:  testAppTargetOrg,
						Repository:    testAppTargetRepo,
						Branch:        testAppTargetBranch,
						Number:        testAppTargetPrNum,
						HeadCommitSha: testAppTargetHeadCommitSha,
						Labels:        []string{},
					}, nil)
					return services.NewGitRemoteRepoAppService(c)
				},
				GitRemoteRepoInfraService: func() *services.GitRemoteRepoInfraService { // unused
					c := mock.NewMockGitIFace(mockCtrl)
					return services.NewGitRemoteRepoInfraService(c)
				},
			},
			args: args{ctx: context.Background(),
				ra: &dreamkastv1alpha1.ReviewApp{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testRaNamespace,
						Name:      testRaName,
					},
					Spec: dreamkastv1alpha1.ReviewAppSpec{
						AppPrNum: testAppTargetPrNum,
						AppTarget: dreamkastv1alpha1.ReviewAppManagerSpecAppTarget{
							Organization: testAppTargetOrg,
							Repository:   testAppTargetRepo,
							Username:     testAppTargetUsername,
							GitSecretRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: testSecretName},
								Key:                  testSecretKey,
							},
						},
						InfraTarget: dreamkastv1alpha1.ReviewAppManagerSpecInfraTarget{
							Organization: testInfraTargetOrg,
							Repository:   testInfraTargetRepo,
							GitSecretRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: testSecretName},
								Key:                  testSecretKey,
							},
						},
						InfraConfig: dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig{
							ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
								Template: dreamkastv1alpha1.NamespacedName{
									Namespace: testAtNamespace, Name: testAtName,
								},
							},
							Manifests: dreamkastv1alpha1.ReviewAppManagerSpecInfraManifests{
								Templates: []dreamkastv1alpha1.NamespacedName{
									{Namespace: testMtNamespace, Name: testMtName},
								},
							},
						},
					},
					Status: dreamkastv1alpha1.ReviewAppStatus{
						Sync: dreamkastv1alpha1.SyncStatus{
							Status:                 "",
							AppRepoLatestCommitSha: testAppTargetHeadCommitSha,
						},
					},
				},
			},
			wantTmp: dreamkastv1alpha1.ReviewAppTmp{
				PullRequest: dreamkastv1alpha1.ReviewAppTmpPr{
					Organization:  testAppTargetOrg,
					Repository:    testAppTargetRepo,
					Branch:        testAppTargetBranch,
					Number:        testAppTargetPrNum,
					HeadCommitSha: testAppTargetHeadCommitSha,
					Labels:        []string{},
				},
				Application: testAtContent,
				Manifests:   testMtContents,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReviewAppReconciler{
				Client:                    tt.fields.Client,
				Log:                       tt.fields.Log,
				Scheme:                    tt.fields.Scheme,
				GitRemoteRepoAppService:   tt.fields.GitRemoteRepoAppService(),
				GitRemoteRepoInfraService: tt.fields.GitRemoteRepoInfraService(),
			}
			_, err := r.prepare(tt.args.ctx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantTmp, tt.args.ra.Tmp); diff != "" {
				t.Errorf("ReviewAppReconciler.prepare() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_confirmAppRepoIsUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testApplicationName := "test-application-name"
	testApplicationNamespace := "test-application-namespace"
	testApplication := newApplication_forPhaseTest(testApplicationNamespace, testApplicationName)
	testAppRepoHeadCommitSha := "1234567"

	type fields struct {
		Client                    client.Client
		Log                       logr.Logger
		Scheme                    *runtime.Scheme
		GitRemoteRepoAppService   func() *services.GitRemoteRepoAppService
		GitRemoteRepoInfraService func() *services.GitRemoteRepoInfraService
	}
	type args struct {
		ctx context.Context
		ra  *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus dreamkastv1alpha1.ReviewAppStatus
		wantErr      bool
	}{
		{
			name: "normal: first time (ra.status.sync.status is empty)",
			fields: fields{
				Client: fake.NewClientBuilder().Build(), // unused
				Log:    logger,
				Scheme: scheme,
				GitRemoteRepoAppService: func() *services.GitRemoteRepoAppService { // unused
					c := mock.NewMockGitHubIFace(mockCtrl)
					return services.NewGitRemoteRepoAppService(c)
				},
				GitRemoteRepoInfraService: func() *services.GitRemoteRepoInfraService { // unused
					c := mock.NewMockGitIFace(mockCtrl)
					return services.NewGitRemoteRepoInfraService(c)
				},
			},
			args: args{ctx: context.Background(),
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{},
					Status: dreamkastv1alpha1.ReviewAppStatus{
						Sync: dreamkastv1alpha1.SyncStatus{
							Status: "",
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						Application: testApplication,
						PullRequest: dreamkastv1alpha1.ReviewAppTmpPr{
							HeadCommitSha: testAppRepoHeadCommitSha,
						},
					},
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status:                 dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					ApplicationName:        testApplicationName,
					ApplicationNamespace:   testApplicationNamespace,
					AppRepoLatestCommitSha: testAppRepoHeadCommitSha,
				},
			},
			wantErr: false,
		},
		{
			name: "normal: after the second time: ra.Tmp.PullRequest.HeadCommitSha == ra.Status.Sync.AppRepoLatestCommitSha",
			fields: fields{
				Client: fake.NewClientBuilder().Build(), // unused
				Log:    logger,
				Scheme: scheme,
				GitRemoteRepoAppService: func() *services.GitRemoteRepoAppService { // unused
					c := mock.NewMockGitHubIFace(mockCtrl)
					return services.NewGitRemoteRepoAppService(c)
				},
				GitRemoteRepoInfraService: func() *services.GitRemoteRepoInfraService { // unused
					c := mock.NewMockGitIFace(mockCtrl)
					return services.NewGitRemoteRepoInfraService(c)
				},
			},
			args: args{ctx: context.Background(),
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{},
					Status: dreamkastv1alpha1.ReviewAppStatus{
						Sync: dreamkastv1alpha1.SyncStatus{
							Status:                 dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo,
							ApplicationName:        testApplicationName,
							ApplicationNamespace:   testApplicationNamespace,
							AppRepoLatestCommitSha: testAppRepoHeadCommitSha,
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						Application: testApplication,
						PullRequest: dreamkastv1alpha1.ReviewAppTmpPr{
							HeadCommitSha: testAppRepoHeadCommitSha,
						},
					},
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status:                 dreamkastv1alpha1.SyncStatusCodeWatchingTemplates,
					ApplicationName:        testApplicationName,
					ApplicationNamespace:   testApplicationNamespace,
					AppRepoLatestCommitSha: testAppRepoHeadCommitSha,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReviewAppReconciler{
				Client:                    tt.fields.Client,
				Log:                       tt.fields.Log,
				Scheme:                    tt.fields.Scheme,
				GitRemoteRepoAppService:   tt.fields.GitRemoteRepoAppService(),
				GitRemoteRepoInfraService: tt.fields.GitRemoteRepoInfraService(),
			}
			_, err := r.confirmAppRepoIsUpdated(tt.args.ctx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.confirmAppRepoIsUpdated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantRaStatus, tt.args.ra.Status); diff != "" {
				t.Errorf("ReviewAppReconciler.confirmAppRepoIsUpdated() is unexpected:\n%v", diff)
			}
		})
	}
}

func TestReviewAppReconciler_confirmTemplatesAreUpdated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	testApplicationName := "test-application-name"
	testApplicationNamespace := "test-application-namespace"
	testManifestNamespace := "test-manifest-namespace"
	testApplication := newApplication_forPhaseTest(testApplicationNamespace, testApplicationName)
	testManifests := map[string]string{
		"manifest.yaml": fmt.Sprintf(`kind: Namespace
metadata:
  name: %s
	`, testManifestNamespace),
	}

	type fields struct {
		Client                    client.Client
		Log                       logr.Logger
		Scheme                    *runtime.Scheme
		GitRemoteRepoAppService   func() *services.GitRemoteRepoAppService
		GitRemoteRepoInfraService func() *services.GitRemoteRepoInfraService
	}
	type args struct {
		ctx context.Context
		ra  *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantRaStatus dreamkastv1alpha1.ReviewAppStatus
		wantErr      bool
	}{
		{
			name: "normal: first time (ra.status.manifestsCache is empty)",
			fields: fields{
				Client: fake.NewClientBuilder().Build(), // unused
				Log:    logger,
				Scheme: scheme,
				GitRemoteRepoAppService: func() *services.GitRemoteRepoAppService { // unused
					c := mock.NewMockGitHubIFace(mockCtrl)
					return services.NewGitRemoteRepoAppService(c)
				},
				GitRemoteRepoInfraService: func() *services.GitRemoteRepoInfraService { // unused
					c := mock.NewMockGitIFace(mockCtrl)
					return services.NewGitRemoteRepoInfraService(c)
				},
			},
			args: args{ctx: context.Background(),
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{},
					Status: dreamkastv1alpha1.ReviewAppStatus{
						Sync: dreamkastv1alpha1.SyncStatus{
							Status:               dreamkastv1alpha1.SyncStatusCodeWatchingTemplates,
							ApplicationName:      testApplicationName,
							ApplicationNamespace: testApplicationNamespace,
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						Application: testApplication,
						Manifests:   testManifests,
					},
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status:               dreamkastv1alpha1.SyncStatusCodeNeedToUpdateInfraRepo,
					ApplicationName:      testApplicationName,
					ApplicationNamespace: testApplicationNamespace,
				},
			},
			wantErr: false,
		},
		{
			name: "normal: after the second time: ra.Tmp.Application == ra.Status.ManifestsCache.Application && ra.Tmp.Manifests == ra.Status.ManifestsCache.Manifests",
			fields: fields{
				Client: fake.NewClientBuilder().Build(), // unused
				Log:    logger,
				Scheme: scheme,
				GitRemoteRepoAppService: func() *services.GitRemoteRepoAppService { // unused
					c := mock.NewMockGitHubIFace(mockCtrl)
					return services.NewGitRemoteRepoAppService(c)
				},
				GitRemoteRepoInfraService: func() *services.GitRemoteRepoInfraService { // unused
					c := mock.NewMockGitIFace(mockCtrl)
					return services.NewGitRemoteRepoInfraService(c)
				},
			},
			args: args{ctx: context.Background(),
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{},
					Status: dreamkastv1alpha1.ReviewAppStatus{
						Sync: dreamkastv1alpha1.SyncStatus{
							Status:               dreamkastv1alpha1.SyncStatusCodeWatchingTemplates,
							ApplicationName:      testApplicationName,
							ApplicationNamespace: testApplicationNamespace,
						},
						ManifestsCache: dreamkastv1alpha1.ManifestsCache{
							Application: testApplication,
							Manifests:   testManifests,
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						Application: testApplication,
						Manifests:   testManifests,
					},
				},
			},
			wantRaStatus: dreamkastv1alpha1.ReviewAppStatus{
				Sync: dreamkastv1alpha1.SyncStatus{
					Status:               dreamkastv1alpha1.SyncStatusCodeWatchingAppRepo,
					ApplicationName:      testApplicationName,
					ApplicationNamespace: testApplicationNamespace,
				},
				ManifestsCache: dreamkastv1alpha1.ManifestsCache{
					Application: testApplication,
					Manifests:   testManifests,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReviewAppReconciler{
				Client:                    tt.fields.Client,
				Log:                       tt.fields.Log,
				Scheme:                    tt.fields.Scheme,
				GitRemoteRepoAppService:   tt.fields.GitRemoteRepoAppService(),
				GitRemoteRepoInfraService: tt.fields.GitRemoteRepoInfraService(),
			}
			_, err := r.confirmTemplatesAreUpdated(tt.args.ctx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.confirmTemplatesAreUpdated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantRaStatus, tt.args.ra.Status); diff != "" {
				t.Errorf("ReviewAppReconciler.confirmTemplatesAreUpdated() is unexpected:\n%v", diff)
			}
		})
	}
}

func newApplication_forPhaseTest(namespace, name string) string {
	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  namespace: %s
  name: %s
spec:
  project: dummy
  destination:
    namespace: dummy
    server: dummy
  source:
    repoURL: dummy
    path: dummy
    targetRevision: dummy`, namespace, name)
}

// TODO
func TestReviewAppReconciler_deployReviewAppManifestsToInfraRepo(t *testing.T) {
	type fields struct {
		Client                    client.Client
		Log                       logr.Logger
		Scheme                    *runtime.Scheme
		GitRemoteRepoAppService   func() *services.GitRemoteRepoAppService
		GitRemoteRepoInfraService func() *services.GitRemoteRepoInfraService
	}
	type args struct {
		ctx context.Context
		ra  *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReviewAppReconciler{
				Client:                    tt.fields.Client,
				Log:                       tt.fields.Log,
				Scheme:                    tt.fields.Scheme,
				GitRemoteRepoAppService:   tt.fields.GitRemoteRepoAppService(),
				GitRemoteRepoInfraService: tt.fields.GitRemoteRepoInfraService(),
			}
			got, err := r.deployReviewAppManifestsToInfraRepo(tt.args.ctx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReviewAppReconciler.deployReviewAppManifestsToInfraRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO
func TestReviewAppReconciler_commentToAppRepoPullRequest(t *testing.T) {
	type fields struct {
		Client                    client.Client
		Log                       logr.Logger
		Scheme                    *runtime.Scheme
		GitRemoteRepoAppService   func() *services.GitRemoteRepoAppService
		GitRemoteRepoInfraService func() *services.GitRemoteRepoInfraService
	}
	type args struct {
		ctx context.Context
		ra  *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReviewAppReconciler{
				Client:                    tt.fields.Client,
				Log:                       tt.fields.Log,
				Scheme:                    tt.fields.Scheme,
				GitRemoteRepoAppService:   tt.fields.GitRemoteRepoAppService(),
				GitRemoteRepoInfraService: tt.fields.GitRemoteRepoInfraService(),
			}
			got, err := r.commentToAppRepoPullRequest(tt.args.ctx, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReviewAppReconciler.commentToAppRepoPullRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
