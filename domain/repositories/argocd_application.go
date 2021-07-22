package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ArgoCDApplictionIFace interface {
	GetArgoCDApplication(ctx context.Context, namespace, name string) (*models.ArgoCDApplication, error)
	SyncArgoCDApplicationStatus(ctx context.Context, app *models.ArgoCDApplication) error
}
