package githubapi

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
	"github.com/go-logr/logr"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GitHubApiInfra struct {
	Log logr.Logger

	username string
	ts       oauth2.TokenSource
}

func NewGitHubApiInfra(l logr.Logger, username, token string) (*GitHubApiInfra, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, username); err != nil {
		return nil, err
	}

	return &GitHubApiInfra{l, username, ts}, nil
}

// TODO: 検索条件を指定可能にする (例. label xxx が付与されている PR は対象外)
func (ki *GitHubApiInfra) ListPullRequests(ctx context.Context, org, repo string) ([]repositories.ListPullRequestsOutput, error) {
	client := github.NewClient(oauth2.NewClient(ctx, ki.ts))
	prs, _, err := client.PullRequests.List(ctx, org, repo, nil)
	if err != nil {
		return nil, err
	}

	var result []repositories.ListPullRequestsOutput
	for _, pr := range prs {
		result = append(result, repositories.ListPullRequestsOutput{
			Number:        *pr.Number,
			HeadCommitSha: *pr.Head.SHA,
		})
	}
	return result, nil
}

func (ki *GitHubApiInfra) CommentToPullRequest(prNum uint) {
	// TODO
}
