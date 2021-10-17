package services

import (
	"context"
	"path/filepath"

	"github.com/cenkalti/backoff/v4"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
)

type GitRemoteRepoInfraService struct {
	gitCommand gateways.GitIFace
}

func NewGitRemoteRepoInfraService(gitCodeIF gateways.GitIFace) *GitRemoteRepoInfraService {
	return &GitRemoteRepoInfraService{gitCodeIF}
}

/* Inputs of some functions */
type UpdateManifestsInput struct {
	Content string
	Path    string
}
type DeleteManifestsInput struct {
	Path string
}

var (
	b = backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
)

type UpdateManifestsParam struct {
	Org       string
	Repo      string
	Branch    string
	CommitMsg string
	Username  string
	Token     string
}

func (s GitRemoteRepoInfraService) UpdateManifests(ctx context.Context,
	param UpdateManifestsParam, ra *dreamkastv1alpha1.ReviewApp,
) (*gateways.GitProject, error) {
	inputManifests := append([]UpdateManifestsInput{}, UpdateManifestsInput{
		Content: ra.Tmp.ApplicationWithAnnotations,
		Path:    ra.Spec.InfraConfig.ArgoCDApp.Filepath,
	})
	for filename, manifest := range ra.Tmp.Manifests {
		inputManifests = append(inputManifests, UpdateManifestsInput{
			Content: manifest,
			Path:    filepath.Join(ra.Spec.InfraConfig.Manifests.Dirpath, filename),
		})
	}

	var gp *gateways.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			if err := s.gitCommand.WithCredential(param.Username, param.Token); err != nil {
				return err
			}
			m, err := s.gitCommand.ForceClone(ctx, param.Org, param.Repo, param.Branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCommand.CreateFile(ctx, *m, manifest.Path, []byte(manifest.Content)); err != nil {
					return err
				}
			}
			_, err = s.gitCommand.CommitAndPush(ctx, *m, param.CommitMsg)
			if err != nil {
				return err
			}
			gp = m
			return nil
		}, b); err != nil {
		return nil, err
	}
	return gp, nil
}

type DeleteManifestsParam struct {
	Org       string
	Repo      string
	Branch    string
	CommitMsg string
	Username  string
	Token     string
}

func (s GitRemoteRepoInfraService) DeleteManifests(ctx context.Context,
	param DeleteManifestsParam, ra *dreamkastv1alpha1.ReviewApp,
) (*gateways.GitProject, error) {
	inputManifests := append([]DeleteManifestsInput{}, DeleteManifestsInput{
		Path: ra.Spec.InfraConfig.ArgoCDApp.Filepath,
	})
	for filename := range ra.Tmp.Manifests {
		inputManifests = append(inputManifests, DeleteManifestsInput{
			Path: filepath.Join(ra.Spec.InfraConfig.Manifests.Dirpath, filename),
		})
	}

	var gp *gateways.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			if err := s.gitCommand.WithCredential(param.Username, param.Token); err != nil {
				return err
			}
			m, err := s.gitCommand.ForceClone(ctx, param.Org, param.Repo, param.Branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCommand.DeleteFile(ctx, *m, manifest.Path); err != nil {
					return err
				}
			}
			_, err = s.gitCommand.CommitAndPush(ctx, *m, param.CommitMsg)
			if err != nil {
				return err
			}
			gp = m
			return nil
		}, b); err != nil {
		return nil, err
	}
	return gp, nil
}
