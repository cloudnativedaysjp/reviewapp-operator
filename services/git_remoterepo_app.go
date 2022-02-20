package services

import (
	"context"
	"regexp"

	"github.com/cloudnativedaysjp/reviewapp-operator/gateways"
)

type GitRemoteRepoAppService struct {
	gitapi gateways.GitHubIFace
}

func NewGitRemoteRepoAppService(gitapi gateways.GitHubIFace) *GitRemoteRepoAppService {
	return &GitRemoteRepoAppService{gitapi}
}

func (s *GitRemoteRepoAppService) ListOpenPullRequestWithSpecificConditions(
	ctx context.Context, org, repo string, username, token string,
	ignoreLabels []string, ignoreTitleExp string,
) ([]*gateways.PullRequest, error) {
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return nil, err
	}
	prs, err := s.gitapi.ListOpenPullRequests(ctx, org, repo)
	if err != nil {
		return nil, err
	}

	// prepare anonymous function using below
	remove := func(slice []*gateways.PullRequest, idx int) []*gateways.PullRequest {
		if idx == len(prs)-1 {
			return slice[:idx]
		} else {
			return append(slice[:idx], slice[idx+1:]...)
		}
	}

	// exclude PRs with specific labels
	removedCount := 0
	for idx, pr := range prs {
		for _, actualLabel := range pr.Labels {
			for _, ignoreLabel := range ignoreLabels {
				if actualLabel == ignoreLabel {
					prs = remove(prs, idx-removedCount)
					removedCount += 1
				}
			}
		}
	}
	// exclude PRs with specific titles
	removedCount = 0
	if ignoreTitleExp != "" {
		r := regexp.MustCompile(ignoreTitleExp)
		for idx, pr := range prs {
			if r.Match([]byte(pr.Title)) {
				prs = remove(prs, idx-removedCount)
				removedCount += 1
			}
		}
	}

	return prs, nil
}

func (s *GitRemoteRepoAppService) GetPullRequest(ctx context.Context, org, repo string, prNum int, username, token string) (*gateways.PullRequest, error) {
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return nil, err
	}
	pr, err := s.gitapi.GetPullRequest(ctx, org, repo, prNum)
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
	if err := s.gitapi.WithCredential(username, token); err != nil {
		return err
	}
	if err := s.gitapi.CommentToPullRequest(ctx, *pr, message); err != nil {
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
