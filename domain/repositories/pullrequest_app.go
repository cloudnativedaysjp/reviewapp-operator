package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type PullRequestApp struct {
	Iface PullRequestAppIFace
}

type PullRequestAppIFace interface {
	WithCredential(string) error
	ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequestApp, error)
	CommentToPullRequest(pr models.PullRequestApp, comment string) error
}

func NewPullRequestApp(iface PullRequestAppIFace) *PullRequestApp {
	return &PullRequestApp{iface}
}
