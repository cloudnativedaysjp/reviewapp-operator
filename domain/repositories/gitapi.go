package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type GitAPI interface {
	WithCredential(credential models.GitCredential) error
	ListOpenPullRequests(ctx context.Context, appRepoTarget models.AppRepoTarget) (models.PullRequests, error)
	GetPullRequest(ctx context.Context, appRepoTarget models.AppRepoTarget, prNum int) (models.PullRequest, error)
	CommentToPullRequest(ctx context.Context, pr models.PullRequest, comment string) error
	GetCommitHashes(ctx context.Context, pr models.PullRequest) ([]string, error)
}
