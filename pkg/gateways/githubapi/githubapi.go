package githubapi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/pkg/models"
)

type GitApiIface interface {
	WithCredential(credential models.GitCredential) error
	ListOpenPullRequests(ctx context.Context, appRepoTarget dreamkastv1alpha1.ReviewAppCommonSpecAppTarget) (dreamkastv1alpha1.PullRequestList, error)
	GetPullRequest(ctx context.Context, appRepoTarget dreamkastv1alpha1.ReviewAppCommonSpecAppTarget, prNum int) (dreamkastv1alpha1.PullRequest, error)
	CommentToPullRequest(ctx context.Context, pr dreamkastv1alpha1.PullRequest, comment string) error
}

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
	// early return if instance has already had client
	if g.haveClient(ctx) {
		return nil
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: credential.Token()},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	if _, _, err := client.Users.Get(ctx, credential.Username()); err != nil {
		return xerrors.Errorf("%w", err)
	}
	g.username = credential.Username()
	g.client = client
	return nil
}

func (g *GitHub) ListOpenPullRequests(ctx context.Context,
	appRepoTarget dreamkastv1alpha1.ReviewAppCommonSpecAppTarget,
) (dreamkastv1alpha1.PullRequestList, error) {
	if !g.haveClient(ctx) {
		return dreamkastv1alpha1.PullRequestList{}, xerrors.Errorf("GitHub have no client")
	}
	prs, _, err := g.client.PullRequests.List(ctx, appRepoTarget.Organization, appRepoTarget.Repository, &github.PullRequestListOptions{State: "open"})
	if err != nil {
		return dreamkastv1alpha1.PullRequestList{}, xerrors.Errorf("%w", err)
	}

	var result []dreamkastv1alpha1.PullRequest
	for _, pr := range prs {
		var labels []string
		for _, l := range pr.Labels {
			labels = append(labels, *l.Name)
		}
		prObj := dreamkastv1alpha1.PullRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name: appRepoTarget.PullRequestName(pr.GetNumber()),
			},
			Spec: dreamkastv1alpha1.PullRequestSpec{
				AppTarget: appRepoTarget,
				Number:    pr.GetNumber(),
			},
			Status: dreamkastv1alpha1.PullRequestStatus{
				BaseBranch:       pr.Base.GetRef(),
				HeadBranch:       pr.Head.GetRef(),
				LatestCommitHash: pr.Head.GetSHA(),
				Title:            pr.GetTitle(),
				Labels:           labels,
			},
		}
		prObj.SetGroupVersionKind(prObj.GVK())
		result = append(result, prObj)
	}
	return dreamkastv1alpha1.PullRequestList{Items: result}, nil
}

func (g *GitHub) GetPullRequest(ctx context.Context,
	appRepoTarget dreamkastv1alpha1.ReviewAppCommonSpecAppTarget, prNum int,
) (dreamkastv1alpha1.PullRequest, error) {
	if !g.haveClient(ctx) {
		return dreamkastv1alpha1.PullRequest{}, xerrors.Errorf("GitHub have no client")
	}
	pr, _, err := g.client.PullRequests.Get(ctx, appRepoTarget.Organization, appRepoTarget.Repository, prNum)
	if err != nil {
		return dreamkastv1alpha1.PullRequest{}, xerrors.Errorf("%w", err)
	}
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, *l.Name)
	}
	prObj := dreamkastv1alpha1.PullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: appRepoTarget.PullRequestName(pr.GetNumber()),
		},
		Spec: dreamkastv1alpha1.PullRequestSpec{
			AppTarget: appRepoTarget,
			Number:    pr.GetNumber(),
		},
		Status: dreamkastv1alpha1.PullRequestStatus{
			BaseBranch:       pr.Base.GetRef(),
			HeadBranch:       pr.Head.GetRef(),
			LatestCommitHash: pr.Head.GetSHA(),
			Title:            pr.GetTitle(),
			Labels:           labels,
		},
	}
	prObj.SetGroupVersionKind(prObj.GVK())
	return prObj, nil
}

func (g *GitHub) CommentToPullRequest(
	ctx context.Context, pr dreamkastv1alpha1.PullRequest, comment string,
) error {
	if !g.haveClient(ctx) {
		return xerrors.Errorf("GitHub have no client")
	}
	// get User
	u, _, err := g.client.Users.Get(ctx, g.username)
	if err != nil {
		return xerrors.Errorf("%w", err)
	}
	// post comment to PR
	if _, _, err := g.client.Issues.CreateComment(ctx,
		pr.Spec.AppTarget.Organization, pr.Spec.AppTarget.Repository,
		pr.Spec.Number, &github.IssueComment{
			Body: &comment,
			User: u,
		}); err != nil {
		return xerrors.Errorf("%w", err)
	}
	return nil
}

func (g *GitHub) haveClient(ctx context.Context) bool {
	if g.client != nil {
		if _, _, err := g.client.Users.Get(ctx, g.username); err == nil {
			return true
		}
	}
	return false
}
