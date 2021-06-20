package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ArgoCDApplication struct {
	Iface ArgoCDApplicationIFace
}

type ArgoCDApplicationIFace interface {
	GetArgoCDApplication(ctx context.Context, namespace, name string) (*models.ArgoCDApplication, error)
	SyncArgoCDApplicationStatus(ctx context.Context, app *models.ArgoCDApplication) error
}

func NewArgoCDApplication(iface ArgoCDApplicationIFace) *ArgoCDApplication {
	return &ArgoCDApplication{iface}
}
