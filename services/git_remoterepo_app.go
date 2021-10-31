package services

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
)

type GitRemoteRepoAppService struct {
	GitApi gateways.GitHubIFace
}

func NewGitRemoteRepoAppService(gitapi gateways.GitHubIFace) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{gitapi}
}

func (s *GitRemoteRepoAppService) ListOpenPullRequest(ctx context.Context, org, repo string, username, token string) ([]*gateways.PullRequest, error) {
	if err := s.GitApi.WithCredential(username, token); err != nil {
		return nil, err
	}
	prs, err := s.GitApi.ListOpenPullRequests(ctx, org, repo)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *GitRemoteRepoAppService) GetPullRequest(ctx context.Context, org, repo string, prNum int, username, token string) (*gateways.PullRequest, error) {
	if err := s.GitApi.WithCredential(username, token); err != nil {
		return nil, err
	}
	pr, err := s.GitApi.GetPullRequest(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *GitRemoteRepoAppService) IsCandidatePr(pr *gateways.PullRequest) bool {
	isCandidate := false
	for _, l := range pr.Labels {
		if l == gateways.CandidateLabelName {
			isCandidate = true
		}
	}
	return isCandidate
}

func (s *GitRemoteRepoAppService) SendMessage(ctx context.Context, pr *gateways.PullRequest, message string, username, token string) error {
	if err := s.GitApi.WithCredential(username, token); err != nil {
		return err
	}
	if err := s.GitApi.CommentToPullRequest(ctx, *pr, message); err != nil {
		return err
	}
	return nil
}

type IsApplicationUpdatedParam struct {
	HashInRA                string
	HashInArgoCDApplication string
}

func (s GitRemoteRepoAppService) IsApplicationUpdated(ctx context.Context, param IsApplicationUpdatedParam) bool {
	if param.HashInRA != "" && param.HashInArgoCDApplication != "" && param.HashInRA == param.HashInArgoCDApplication {
		return true
	}
	return false
}
