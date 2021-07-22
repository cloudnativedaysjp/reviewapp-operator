package repositories

import (
	"context"
	"io"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type GitCodeIFace interface {
	WithCredential(models.GitRepoCredential) error
	Pull(ctx context.Context, org, repo string) (models.GitProject, error)
	CheckDirectoryExistence(ctx context.Context, project models.GitProject, dirname string) error
	WithCreateFile(ctx context.Context, project models.GitProject, filename string, contents io.Reader) error
	WithDeleteFile(ctx context.Context, project models.GitProject, filename string) error
	CommitAndPush(ctx context.Context, project models.GitProject, message string) error
}
