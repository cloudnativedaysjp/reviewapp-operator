package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type GitCodeIFace interface {
	WithCredential(username, token string) error
	Pull(ctx context.Context, org, repo, branch string) (*models.GitProject, error)
	CreateFile(ctx context.Context, gp models.GitProject, filename string, contents []byte) error
	DeleteFile(ctx context.Context, gp models.GitProject, filename string) error
	CommitAndPush(ctx context.Context, gp models.GitProject, message string) (*models.GitProject, error)
}
