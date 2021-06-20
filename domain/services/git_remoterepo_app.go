package services

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
)

type GitRemoteRepoAppService struct {
	prRepo     repositories.PullRequestAppIFace
	secretRepo repositories.GitRepoSecretIFace
}

func NewGitRemoteRepoAppService(pr repositories.PullRequestAppIFace, secret repositories.GitRepoSecretIFace) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{pr, secret}
}

type ListOpenPullRequestInput struct {
	SecretNamespace     string
	SecretName          string
	SecretKey           string
	GitOrganization     string
	GitRemoteRepository string
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, input ListOpenPullRequestInput) ([]*models.PullRequestApp, error) {
	credential, err := s.secretRepo.GetSecretValue(ctx, input.SecretNamespace, input.SecretName, input.SecretKey)
	if err != nil {
		return nil, err
	}
	if err := s.prRepo.WithCredential(credential); err != nil {
		return nil, err
	}
	prs, err := s.prRepo.ListOpenPullRequests(ctx, input.GitOrganization, input.GitRemoteRepository)
	if err != nil {
		return nil, err
	}
	return prs, nil
}
