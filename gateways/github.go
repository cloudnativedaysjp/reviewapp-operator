package gateways

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
)

const CandidateLabelName = "candidate-template"

type PullRequest struct {
	Organization  string
	Repository    string
	Branch        string
	Number        int
	HeadCommitSha string
	Labels        []string
}

func NewPullRequest(organization, repository, branch string, number int, headCommitSha string, labels []string) *PullRequest {
	return &PullRequest{
		Organization:  organization,
		Repository:    repository,
		Branch:        branch,
		Number:        number,
		HeadCommitSha: headCommitSha,
		Labels:        labels,
	}
}

type GitHubIFace interface {
	WithCredential(username, token string) error
	ListOpenPullRequests(ctx context.Context, org, repo string) ([]*PullRequest, error)
	GetPullRequest(ctx context.Context, org, repo string, prNum int) (*PullRequest, error)
	CommentToPullRequest(ctx context.Context, pr PullRequest, comment string) error
	GetCommitHashes(ctx context.Context, pr PullRequest) ([]string, error)
}

type GitHub struct {
	logger logr.Logger

	username string
	client   *github.Client
}

func NewGitHub(l logr.Logger) *GitHub {
	return &GitHub{logger: l}
}

func (g *GitHub) WithCredential(username, token string) error {
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
func (g *GitHub) ListOpenPullRequests(ctx context.Context, org, repo string) ([]*PullRequest, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitHub have no client")
	}
	prs, _, err := g.client.PullRequests.List(ctx, org, repo, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	var result []*PullRequest
	for _, pr := range prs {
		var labels []string
		for _, l := range pr.Labels {
			labels = append(labels, *l.Name)
		}
		result = append(result, NewPullRequest(org, repo, pr.Head.GetRef(), pr.GetNumber(), pr.Head.GetSHA(), labels))
	}
	return result, nil
}

func (g *GitHub) GetPullRequest(ctx context.Context, org, repo string, prNum int) (*PullRequest, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitHub have no client")
	}
	pr, _, err := g.client.PullRequests.Get(ctx, org, repo, prNum)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, *l.Name)
	}
	return NewPullRequest(org, repo, pr.Head.GetRef(), prNum, pr.Head.GetSHA(), labels), nil
}

func (g *GitHub) CommentToPullRequest(ctx context.Context, pr PullRequest, comment string) error {
	if !g.haveClient(ctx) {
		return xerrors.Errorf("GitHub have no client")
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

func (g *GitHub) GetCommitHashes(ctx context.Context, prModel PullRequest) ([]string, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitHub have no client")
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

func (g *GitHub) haveClient(ctx context.Context) bool {
	if g.client != nil {
		if _, _, err := g.client.Users.Get(ctx, g.username); err == nil {
			return true
		}
	}
	return false
}
