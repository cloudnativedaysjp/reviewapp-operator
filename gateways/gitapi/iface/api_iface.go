package gitapi_iface

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitApiIFace interface {
	WithCredential(username, token string) error
	ListOpenPullRequests(ctx context.Context, org, repo string) ([]*models.PullRequest, error)
	GetOpenPullRequest(ctx context.Context, org, repo string, prNum int) (*models.PullRequest, error)
	CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error
	GetCommitHashes(ctx context.Context, pr models.PullRequest) ([]string, error)
}
