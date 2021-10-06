package services

import (
	"context"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/go-logr/logr"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitRemoteRepoAppService struct {
	gitapi gateways.GitApiIFace

	Log logr.Logger
}

func NewGitRemoteRepoAppService(gitapi gateways.GitApiIFace, logger logr.Logger) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{gitapi, logger}
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, org, repo string, username, token string) ([]*models.PullRequest, error) {
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return nil, err
	}
	prs, err := s.gitapi.ListOpenPullRequests(ctx, org, repo)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *GitRemoteRepoAppService) GetOpenPullRequest(ctx context.Context, org, repo string, prNum int, username, token string) (*models.PullRequest, error) {
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return nil, err
	}
	pr, err := s.gitapi.GetOpenPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *GitRemoteRepoAppService) SendMessage(ctx context.Context, pr *models.PullRequest, message string, username, token string) error {
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return err
	}
	if err := s.gitapi.CommentToPullRequest(ctx, *pr, message); err != nil {
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
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return false, err
	}
	pr, err := s.gitapi.GetOpenPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return false, err
	}
	hashes, err := s.gitapi.GetCommitHashes(ctx, *pr)
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
