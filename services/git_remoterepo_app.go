package services

import (
	"context"
	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitRemoteRepoAppService struct {
	gitapi gateways.GitHubApiIFace
}

func NewGitRemoteRepoAppService(gitapi gateways.GitHubApiIFace) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{gitapi}
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

type IsApplicationUpdatedParam struct {
	Org                     string
	Repo                    string
	PrNum                   int
	Username                string
	Token                   string
	HashInRA                string
	HashInArgoCDApplication string
}

func (s GitRemoteRepoAppService) IsApplicationUpdated(ctx context.Context, param IsApplicationUpdatedParam) (bool, error) {
	if param.HashInRA == param.HashInArgoCDApplication {
		return true, nil
	}
	if err := s.gitapi.WithCredential(param.Username, param.Token); err != nil {
		return false, err
	}
	pr, err := s.gitapi.GetOpenPullRequest(ctx, param.Org, param.Repo, param.PrNum)
	if err != nil {
		return false, err
	}
	hashes, err := s.gitapi.GetCommitHashes(ctx, *pr)
	if err != nil {
		return false, err
	}

	for _, hash := range hashes {
		if hash == param.HashInArgoCDApplication {
			return true, nil
		}
	}
	return false, nil
}
