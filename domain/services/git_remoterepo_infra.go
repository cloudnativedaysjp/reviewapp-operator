package services

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/go-logr/logr"
)

type GitRemoteRepoInfraService struct {
	gitCodeRepo   repositories.GitCodeIFace
	k8sSecretRepo repositories.SecretIFace

	Log logr.Logger
}

func NewGitRemoteRepoInfraService(gitCodeIF repositories.GitCodeIFace, secretIF repositories.SecretIFace, logger logr.Logger) *GitRemoteRepoInfraService {
	return &GitRemoteRepoInfraService{gitCodeIF, secretIF, logger}
}

/* Inputs of some functions */
type AccessToInfraRepoInput struct {
	Namespace string
	Name      string
	Key       string
}
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

func (s GitRemoteRepoInfraService) UpdateManifests(ctx context.Context, org, repo, branch, commitMsg string, inputSecret AccessToInfraRepoInput, inputManifests ...UpdateManifestsInput) (*models.GitProject, error) {
	var gp *models.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCodeRepo.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			credential, err := s.k8sSecretRepo.GetSecretValue(ctx, inputSecret.Namespace, inputSecret.Name, inputSecret.Key)
			if err != nil {
				return err
			}
			if err := s.gitCodeRepo.WithCredential(credential); err != nil {
				return err
			}
			m, err := s.gitCodeRepo.Pull(ctx, org, repo, branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCodeRepo.CreateFile(ctx, *m, manifest.Path, []byte(manifest.Content)); err != nil {
					return err
				}
			}
			_, err = s.gitCodeRepo.CommitAndPush(ctx, *m, commitMsg)
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

func (s GitRemoteRepoInfraService) DeleteManifests(ctx context.Context, org, repo, branch, commitMsg string, inputSecret AccessToInfraRepoInput, inputManifests ...DeleteManifestsInput) (*models.GitProject, error) {
	var gp *models.GitProject
	// 処理中に誰かが同一ブランチにpushすると s.gitCodeRepo.CommitAndPush() に失敗するため、リトライする
	if err := backoff.Retry(
		func() error {
			credential, err := s.k8sSecretRepo.GetSecretValue(ctx, inputSecret.Namespace, inputSecret.Name, inputSecret.Key)
			if err != nil {
				return err
			}
			if err := s.gitCodeRepo.WithCredential(credential); err != nil {
				return err
			}
			m, err := s.gitCodeRepo.Pull(ctx, org, repo, branch)
			if err != nil {
				return err
			}
			for _, manifest := range inputManifests {
				if err := s.gitCodeRepo.DeleteFile(ctx, *m, manifest.Path); err != nil {
					return err
				}
			}
			_, err = s.gitCodeRepo.CommitAndPush(ctx, *m, commitMsg)
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
