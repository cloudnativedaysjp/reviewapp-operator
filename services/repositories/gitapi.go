package repositories

import (
	"context"

	"github.com/go-logr/logr"
)

type GitApiFactory interface {
	NewRepository(l logr.Logger, username, token string) (GitApiRepository, error)
}

type GitApiRepository interface {
	ListPullRequestsWithOpen(ctx context.Context, org, repo string) ([]ListPullRequestsOutput, error)
}
