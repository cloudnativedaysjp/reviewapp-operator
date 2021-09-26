package git_iface

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/models"
)

type GitCommandIFace interface {
	WithCredential(username, token string) error
	Pull(ctx context.Context, org, repo, branch string) (*models.GitProject, error)
	CreateFile(ctx context.Context, gp models.GitProject, filename string, contents []byte) error
	DeleteFile(ctx context.Context, gp models.GitProject, filename string) error
	CommitAndPush(ctx context.Context, gp models.GitProject, message string) (*models.GitProject, error)
}
