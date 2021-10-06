package services

import (
	"context"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"path/filepath"

	"github.com/cenkalti/backoff/v4"
	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitRemoteRepoInfraService struct {
	gitCommand gateways.GitCommandIFace
}

func NewGitRemoteRepoInfraService(gitCodeIF gateways.GitCommandIFace) *GitRemoteRepoInfraService {
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

func (s GitRemoteRepoInfraService) UpdateManifests(
	ctx context.Context, org, repo, branch, commitMsg string,
	username, token string,
	ra *dreamkastv1beta1.ReviewApp,
) (*models.GitProject, error) {
	inputManifests := append([]UpdateManifestsInput{}, UpdateManifestsInput{
		Content: ra.Spec.Application,
		Path:    ra.Spec.InfraConfig.ArgoCDApp.Filepath,
	})
	for filename, manifest := range ra.Spec.Manifests {
		inputManifests = append(inputManifests, UpdateManifestsInput{
			Content: manifest,
			Path:    filepath.Join(ra.Spec.InfraConfig.Manifests.Dirpath, filename),
		})
	}

	var gp *models.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			if err := s.gitCommand.WithCredential(username, token); err != nil {
				return err
			}
			m, err := s.gitCommand.Pull(ctx, org, repo, branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCommand.CreateFile(ctx, *m, manifest.Path, []byte(manifest.Content)); err != nil {
					return err
				}
			}
			_, err = s.gitCommand.CommitAndPush(ctx, *m, commitMsg)
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

func (s GitRemoteRepoInfraService) DeleteManifests(
	ctx context.Context, org, repo, branch, commitMsg string,
	username, token string,
	ra *dreamkastv1beta1.ReviewApp,
) (*models.GitProject, error) {
	inputManifests := append([]DeleteManifestsInput{}, DeleteManifestsInput{
		Path: ra.Spec.InfraConfig.ArgoCDApp.Filepath,
	})
	for filename := range ra.Spec.Manifests {
		inputManifests = append(inputManifests, DeleteManifestsInput{
			Path: filepath.Join(ra.Spec.InfraConfig.Manifests.Dirpath, filename),
		})
	}

	var gp *models.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCommand.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			if err := s.gitCommand.WithCredential(username, token); err != nil {
				return err
			}
			m, err := s.gitCommand.Pull(ctx, org, repo, branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCommand.DeleteFile(ctx, *m, manifest.Path); err != nil {
					return err
				}
			}
			_, err = s.gitCommand.CommitAndPush(ctx, *m, commitMsg)
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
