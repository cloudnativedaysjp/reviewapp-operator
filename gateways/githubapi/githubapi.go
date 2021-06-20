package githubapi

import (
	"context"
	"fmt"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	"github.com/go-logr/logr"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
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
	if _, err := g.ts.Token(); err == nil {
		return nil
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
func (g *GitHubApiGateway) ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequestApp, error) {
	client := github.NewClient(oauth2.NewClient(ctx, g.ts))
	prs, _, err := client.PullRequests.List(ctx, org, repo, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return nil, err
	}

	var result []*models.PullRequestApp
	for _, pr := range prs {
		result = append(result, models.NewPullRequestApp(org, repo, *pr.Number, *pr.Head.SHA))
	}
	return result, nil
}

func (g *GitHubApiGateway) CommentToPullRequest(pr models.PullRequestApp, comment string) error {
	// TODO
	return fmt.Errorf("not implemented")
}
