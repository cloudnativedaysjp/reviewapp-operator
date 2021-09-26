package services

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/repositories"
	"github.com/go-logr/logr"
)

type GitRemoteRepoAppService struct {
	gitPrRepo repositories.PullRequestIFace

	Log logr.Logger
}

func NewGitRemoteRepoAppService(prIF repositories.PullRequestIFace, logger logr.Logger) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{prIF, logger}
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, org, repo string, username, token string) ([]*models.PullRequest, error) {
	if err := s.gitPrRepo.WithCredential(username, token); err != nil {
		return nil, err
	}
	prs, err := s.gitPrRepo.ListOpenPullRequests(ctx, org, repo)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *GitRemoteRepoAppService) GetOpenPullRequest(ctx context.Context, org, repo string, prNum int, username, token string) (*models.PullRequest, error) {
	if err := s.gitPrRepo.WithCredential(username, token); err != nil {
		return nil, err
	}
	pr, err := s.gitPrRepo.GetOpenPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *GitRemoteRepoAppService) SendMessage(ctx context.Context, pr *models.PullRequest, message string, username, token string) error {
	if err := s.gitPrRepo.WithCredential(username, token); err != nil {
		return err
	}
	if err := s.gitPrRepo.CommentToPullRequest(ctx, *pr, message); err != nil {
		return err
	}
	return nil
}

func (s GitRemoteRepoAppService) CheckApplicationUpdated(
	ctx context.Context, org, repo string, prNum int,
	username, token string,
	hashInRA, hashInArgoCDApplication string,
) (bool, error) {
	if hashInRA == hashInArgoCDApplication {
		return true, nil
	}
	if err := s.gitPrRepo.WithCredential(username, token); err != nil {
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
