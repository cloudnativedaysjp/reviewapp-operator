package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type PullRequestIFace interface {
	WithCredential(string) error
	ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequest, error)
	CommentToPullRequest(pr models.PullRequest, comment string) error
}
