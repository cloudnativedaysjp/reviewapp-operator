package githubapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

const CandidateLabelName = "candidate-template"

type GitHub struct {
	logger logr.Logger

	username string
	client   *github.Client
}

func NewGitHub(l logr.Logger) *GitHub {
	return &GitHub{logger: l}
}

func (g *GitHub) WithCredential(credential models.GitCredential) error {
	ctx := context.Background()
	// 既に client を持っているなら早期リターン
	if g.haveClient(ctx) {
		return nil
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: credential.Token()},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, g.username); err != nil {
		return xerrors.Errorf("%w", err)
	}
	g.username = credential.Username()
	g.client = client
	return nil
}

func (g *GitHub) ListOpenPullRequests(ctx context.Context, appRepoTarget models.AppRepoTarget) ([]models.PullRequest, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitHub have no client")
	}
	prs, _, err := g.client.PullRequests.List(ctx, appRepoTarget.Organization, appRepoTarget.Repository, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	var result []models.PullRequest
	for _, pr := range prs {
		var labels []string
		for _, l := range pr.Labels {
			labels = append(labels, *l.Name)
		}
		result = append(result, models.NewPullRequest(appRepoTarget.Organization, appRepoTarget.Repository, pr.Head.GetRef(), pr.GetNumber(), pr.Head.GetSHA(), pr.GetTitle(), labels))
	}
	return result, nil
}

func (g *GitHub) GetPullRequest(ctx context.Context, appRepoTarget models.AppRepoTarget, prNum int) (models.PullRequest, error) {
	if !g.haveClient(ctx) {
		return models.PullRequest{}, xerrors.Errorf("GitHub have no client")
	}
	pr, _, err := g.client.PullRequests.Get(ctx, appRepoTarget.Organization, appRepoTarget.Repository, prNum)
	if err != nil {
		return models.PullRequest{}, xerrors.Errorf("%w", err)
	}
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, *l.Name)
	}
	return models.NewPullRequest(appRepoTarget.Organization, appRepoTarget.Repository, pr.Head.GetRef(), prNum, pr.Head.GetSHA(), pr.GetTitle(), labels), nil
}

func (g *GitHub) CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error {
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

func (g *GitHub) GetCommitHashes(ctx context.Context, prModel models.PullRequest) ([]string, error) {
	if !g.haveClient(ctx) {
		return nil, xerrors.Errorf("GitHub have no client")
	}
	prCommits, _, err := g.client.PullRequests.ListCommits(ctx, prModel.Organization, prModel.Repository, prModel.Number, &github.ListOptions{})
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	result := []string{}
	for _, prCommit := range prCommits {
		result = append(result, prCommit.Commit.GetSHA())
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
