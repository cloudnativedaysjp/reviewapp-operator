package githubapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type GitHubApiGateway struct {
	logger logr.Logger

	username string
	ts       oauth2.TokenSource
}

func NewGitHubApiGateway(l logr.Logger, username string) *GitHubApiGateway {
	return &GitHubApiGateway{logger: l, username: username}
}

func (g *GitHubApiGateway) WithCredential(token string) error {
	// 既にtokenを持っているなら早期リターン
	if g.ts != nil {
		if _, err := g.ts.Token(); err == nil {
			return nil
		}
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return err
	}
	g.ts = ts
	return nil
}

// TODO: 検索条件を指定可能にする (例. label xxx が付与されている PR は対象外)
func (g *GitHubApiGateway) ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequest, error) {
	client := github.NewClient(oauth2.NewClient(ctx, g.ts))
	prs, _, err := client.PullRequests.List(ctx, org, repo, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return nil, err
	}

	var result []*models.PullRequest
	for _, pr := range prs {
		result = append(result, models.NewPullRequest(org, repo, *pr.Number, *pr.Head.SHA))
	}
	return result, nil
}

func (g *GitHubApiGateway) GetOpenPullRequest(ctx context.Context, org, repo string, prNum int) (*models.PullRequest, error) {
	client := github.NewClient(oauth2.NewClient(ctx, g.ts))
	pr, _, err := client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return nil, err
	}
	return &models.PullRequest{
		Organization:  org,
		Repository:    repo,
		Number:        prNum,
		HeadCommitSha: *pr.Head.SHA,
	}, nil
}

func (g *GitHubApiGateway) CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error {
	client := github.NewClient(oauth2.NewClient(ctx, g.ts))
	// get User
	u, _, err := client.Users.Get(ctx, g.username)
	if err != nil {
		return err
	}
	// post comment to PR
	if _, _, err := client.Issues.CreateComment(ctx, pr.Organization, pr.Repository, pr.Number, &github.IssueComment{
		Body: &comment,
		User: u,
	}); err != nil {
		return err
	}
	return nil
}

func (g *GitHubApiGateway) GetCommitHashes(ctx context.Context, prModel models.PullRequest) ([]string, error) {
	client := github.NewClient(oauth2.NewClient(ctx, g.ts))
	prs, _, err := client.PullRequests.ListCommits(ctx, prModel.Organization, prModel.Repository, prModel.Number, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, pr := range prs {
		result = append(result, *pr.Commit.SHA)
	}
	return result, nil
}
