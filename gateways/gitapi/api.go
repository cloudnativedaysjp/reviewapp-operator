package gitapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitApiDriver struct {
	logger logr.Logger

	username string
	client   *github.Client
}

func NewGitApiDriver(l logr.Logger) *GitApiDriver {
	return &GitApiDriver{logger: l}
}

func (g *GitApiDriver) WithCredential(username, token string) error {
	ctx := context.Background()
	// 既に client を持っているなら早期リターン
	if g.haveClient(ctx) {
		return nil
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return xerrors.Errorf("%w", err)
	}
	g.username = username
	g.client = client
	return nil
}

// TODO: 検索条件を指定可能にする (例. label xxx が付与されている PR は対象外)
func (g *GitApiDriver) ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequest, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitApiDriver have no client")
	}
	prs, _, err := g.client.PullRequests.List(ctx, org, repo, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	var result []*models.PullRequest
	for _, pr := range prs {
		var labels []string
		for _, l := range pr.Labels {
			labels = append(labels, *l.Name)
		}
		result = append(result, models.NewPullRequest(org, repo, *pr.Number, *pr.Head.SHA, labels))
	}
	return result, nil
}

func (g *GitApiDriver) GetOpenPullRequest(ctx context.Context, org, repo string, prNum int) (*models.PullRequest, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitApiDriver have no client")
	}
	pr, _, err := g.client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, *l.Name)
	}
	return &models.PullRequest{
		Organization:  org,
		Repository:    repo,
		Number:        prNum,
		HeadCommitSha: *pr.Head.SHA,
		Labels:        labels,
	}, nil
}

func (g *GitApiDriver) CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error {
	if !g.haveClient(ctx) {
		return xerrors.Errorf("GitApiDriver have no client")
	}
	// get User
	u, _, err := g.client.Users.Get(ctx, g.username)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	// post comment to PR
	if _, _, err := g.client.Issues.CreateComment(ctx, pr.Organization, pr.Repository, pr.Number, &github.IssueComment{
		Body: &comment,
		User: u,
	}); err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (g *GitApiDriver) GetCommitHashes(ctx context.Context, prModel models.PullRequest) ([]string, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitApiDriver have no client")
	}
	prs, _, err := g.client.PullRequests.ListCommits(ctx, prModel.Organization, prModel.Repository, prModel.Number, &github.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	result := []string{}
	for _, pr := range prs {
		result = append(result, *pr.Commit.SHA)
	}
	return result, nil
}

func (g *GitApiDriver) haveClient(ctx context.Context) bool {
	if g.client != nil {
		if _, _, err := g.client.Users.Get(ctx, g.username); err == nil {
			return true
		}
	}
	return false
}
