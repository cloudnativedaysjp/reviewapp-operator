package services

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
)

type GitRemoteRepoAppService struct {
	gitPrRepo repositories.PullRequestIFace
	k8sRepo   repositories.SecretIFace
}

func NewGitRemoteRepoAppService(pr repositories.PullRequestIFace, secret repositories.SecretIFace) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{pr, secret}
}

type ListOpenPullRequestInput struct {
	SecretNamespace     string
	SecretName          string
	SecretKey           string
	GitOrganization     string
	GitRemoteRepository string
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, input ListOpenPullRequestInput) ([]*models.PullRequest, error) {
	credential, err := s.k8sRepo.GetSecretValue(ctx, input.SecretNamespace, input.SecretName, input.SecretKey)
	if err != nil {
		return nil, err
	}
	if err := s.gitPrRepo.WithCredential(credential); err != nil {
		return nil, err
	}
	prs, err := s.gitPrRepo.ListOpenPullRequests(ctx, input.GitOrganization, input.GitRemoteRepository)
	if err != nil {
		return nil, err
	}
	return prs, nil
}
