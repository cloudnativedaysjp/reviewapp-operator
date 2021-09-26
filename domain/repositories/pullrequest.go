package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type PullRequestIFace interface {
	WithCredential(username, token string) error
	ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequest, error)
	GetOpenPullRequest(ctx context.Context, org, repo string, prNum int) (*models.PullRequest, error)
	CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error
	GetCommitHashes(ctx context.Context, pr models.PullRequest) ([]string, error)
}
