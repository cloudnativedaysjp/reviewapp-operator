package services

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/mock"
	"github.com/golang/mock/gomock"
)

func TestGitRemoteRepoInfraService_UpdateManifests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testUsername := "testuser"
	testToken := "testtoken"
	testOrg := "testorg"
	testRepo := "testrepo"
	testBranchName := "testbranch"
	testLatestCommitSha := "1234567"
	testArgoCDFilePath := "/path/to/argocd.yaml"
	testManifestsDirPath := "/path/to/manifests"
	testManifestContent := "testContent"
	testMessage := "testmessage"
	testGitProject := &gateways.GitProject{DownloadDir: "/dummy", LatestCommitSha: testLatestCommitSha}

	type fields struct {
		gitCommand func() gateways.GitIFace
	}
	type args struct {
		ctx   context.Context
		param UpdateManifestsParam
		ra    *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *gateways.GitProject
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{func() gateways.GitIFace {
				c := mock.NewMockGitIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ForceClone(context.Background(), testOrg, testRepo, testBranchName).Return(testGitProject, nil)
				for _, path := range []string{
					testArgoCDFilePath,
					filepath.Join(testManifestsDirPath, "01.yaml"),
					filepath.Join(testManifestsDirPath, "02.yaml"),
					filepath.Join(testManifestsDirPath, "03.yaml"),
				} {
					c.EXPECT().CreateFile(context.Background(), *testGitProject, path, []byte(testManifestContent)).Return(nil)
				}
				c.EXPECT().CommitAndPush(context.Background(), *testGitProject, testMessage).Return(testGitProject, nil)
				return c
			}},
			args: args{
				ctx:   context.Background(),
				param: UpdateManifestsParam{testOrg, testRepo, testBranchName, testMessage, testUsername, testToken},
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{
						InfraConfig: dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig{
							ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
								Filepath: testArgoCDFilePath,
							},
							Manifests: dreamkastv1alpha1.ReviewAppManagerSpecInfraManifests{
								Dirpath: testManifestsDirPath,
							},
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						ApplicationWithAnnotations: testManifestContent,
						Manifests: map[string]string{
							"01.yaml": testManifestContent,
							"02.yaml": testManifestContent,
							"03.yaml": testManifestContent,
						},
					},
				},
			},
			want:    testGitProject,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := GitRemoteRepoInfraService{
				gitCommand: tt.fields.gitCommand(),
			}
			got, err := s.UpdateManifests(tt.args.ctx, tt.args.param, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRemoteRepoInfraService.UpdateManifests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitRemoteRepoInfraService.UpdateManifests() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitRemoteRepoInfraService_DeleteManifests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testUsername := "testuser"
	testToken := "testtoken"
	testOrg := "testorg"
	testRepo := "testrepo"
	testBranchName := "testbranch"
	testLatestCommitSha := "1234567"
	testArgoCDFilePath := "/path/to/argocd.yaml"
	testManifestsDirPath := "/path/to/manifests"
	testMessage := "testmessage"
	testGitProject := &gateways.GitProject{DownloadDir: "/dummy", LatestCommitSha: testLatestCommitSha}

	type fields struct {
		gitCommand func() gateways.GitIFace
	}
	type args struct {
		ctx   context.Context
		param DeleteManifestsParam
		ra    *dreamkastv1alpha1.ReviewApp
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *gateways.GitProject
		wantErr bool
	}{
		{
			name: "normal",
			fields: fields{func() gateways.GitIFace {
				c := mock.NewMockGitIFace(ctrl)
				c.EXPECT().WithCredential(testUsername, testToken).Return(nil)
				c.EXPECT().ForceClone(context.Background(), testOrg, testRepo, testBranchName).Return(testGitProject, nil)
				for _, path := range []string{
					testArgoCDFilePath,
					filepath.Join(testManifestsDirPath, "01.yaml"),
					filepath.Join(testManifestsDirPath, "02.yaml"),
					filepath.Join(testManifestsDirPath, "03.yaml"),
				} {
					c.EXPECT().DeleteFile(context.Background(), *testGitProject, path).Return(nil)
				}
				c.EXPECT().CommitAndPush(context.Background(), *testGitProject, testMessage).Return(testGitProject, nil)
				return c
			}},
			args: args{
				ctx:   context.Background(),
				param: DeleteManifestsParam{testOrg, testRepo, testBranchName, testMessage, testUsername, testToken},
				ra: &dreamkastv1alpha1.ReviewApp{
					Spec: dreamkastv1alpha1.ReviewAppSpec{
						InfraConfig: dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig{
							ArgoCDApp: dreamkastv1alpha1.ReviewAppManagerSpecInfraArgoCDApp{
								Filepath: testArgoCDFilePath,
							},
							Manifests: dreamkastv1alpha1.ReviewAppManagerSpecInfraManifests{
								Dirpath: testManifestsDirPath,
							},
						},
					},
					Tmp: dreamkastv1alpha1.ReviewAppTmp{
						ApplicationWithAnnotations: "",
						Manifests: map[string]string{
							"01.yaml": "",
							"02.yaml": "",
							"03.yaml": "",
						},
					},
				},
			},
			want:    testGitProject,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := GitRemoteRepoInfraService{
				gitCommand: tt.fields.gitCommand(),
			}
			got, err := s.DeleteManifests(tt.args.ctx, tt.args.param, tt.args.ra)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitRemoteRepoInfraService.DeleteManifests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitRemoteRepoInfraService.DeleteManifests() = %v, want %v", got, tt.want)
			}
		})
	}
}
