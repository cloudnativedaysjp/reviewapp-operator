package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type GitCommand interface {
	WithCredential(credential models.GitCredential) error
	ForceClone(context.Context, models.InfraRepoTarget) (models.InfraRepoLocalDir, error)
	CreateFiles(context.Context, models.InfraRepoLocalDir, ...models.File) error
	DeleteFiles(context.Context, models.InfraRepoLocalDir, ...models.File) error
	CommitAndPush(ctx context.Context, gp models.InfraRepoLocalDir, message string) (*models.InfraRepoLocalDir, error)
}
