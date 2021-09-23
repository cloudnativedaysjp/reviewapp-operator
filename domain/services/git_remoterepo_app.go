package services

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/go-logr/logr"
)

type GitRemoteRepoAppService struct {
	gitPrRepo     repositories.PullRequestIFace
	k8sSecretRepo repositories.SecretIFace

	Log logr.Logger
}

func NewGitRemoteRepoAppService(prIF repositories.PullRequestIFace, secretIF repositories.SecretIFace, logger logr.Logger) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{prIF, secretIF, logger}
}

type AccessToAppRepoInput struct {
	SecretNamespace string
	SecretName      string
	SecretKey       string
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, org, repo string, input AccessToAppRepoInput) ([]*models.PullRequest, error) {
	credential, err := s.k8sSecretRepo.GetSecretValue(ctx, input.SecretNamespace, input.SecretName, input.SecretKey)
	if err != nil {
		return nil, err
	}
	if err := s.gitPrRepo.WithCredential(credential); err != nil {
		return nil, err
	}
	prs, err := s.gitPrRepo.ListOpenPullRequests(ctx, org, repo)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *GitRemoteRepoAppService) GetOpenPullRequest(ctx context.Context, org, repo string, prNum int, input AccessToAppRepoInput) (*models.PullRequest, error) {
	credential, err := s.k8sSecretRepo.GetSecretValue(ctx, input.SecretNamespace, input.SecretName, input.SecretKey)
	if err != nil {
		return nil, err
	}
	if err := s.gitPrRepo.WithCredential(credential); err != nil {
		return nil, err
	}
	pr, err := s.gitPrRepo.GetOpenPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *GitRemoteRepoAppService) SendMessage(ctx context.Context, pr *models.PullRequest, message string, input AccessToAppRepoInput) error {
	credential, err := s.k8sSecretRepo.GetSecretValue(ctx, input.SecretNamespace, input.SecretName, input.SecretKey)
	if err != nil {
		return err
	}
	if err := s.gitPrRepo.WithCredential(credential); err != nil {
		return err
	}
	if err := s.gitPrRepo.CommentToPullRequest(ctx, *pr, message); err != nil {
		return err
	}
	return nil
}

func (s GitRemoteRepoAppService) CheckApplicationUpdated(ctx context.Context,
	org, repo string, prNum int,
	inputSecret AccessToInfraRepoInput,
	hashInRA, hashInArgoCDApplication string,
) (bool, error) {
	if hashInRA == hashInArgoCDApplication {
		return true, nil
	}
	credential, err := s.k8sSecretRepo.GetSecretValue(ctx, inputSecret.Namespace, inputSecret.Name, inputSecret.Key)
	if err != nil {
		return false, err
	}
	if err := s.gitPrRepo.WithCredential(credential); err != nil {
		return false, err
	}
	pr, err := s.gitPrRepo.GetOpenPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return false, err
	}
	hashes, err := s.gitPrRepo.GetCommitHashes(ctx, *pr)
	if err != nil {
		return false, err
	}

	for _, hash := range hashes {
		if hash == hashInArgoCDApplication {
			return true, nil
		}
	}
	return false, nil
}
