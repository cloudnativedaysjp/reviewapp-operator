package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ReviewAppConfigIFace interface {
	GetReviewAppConfig(ctx context.Context, namespace, name string) (*models.ReviewAppConfig, error)
	UpdateReviewAppManagerStatus(ctx context.Context, rac *models.ReviewAppConfig) error
	AddFinalizersToReviewApp(ctx context.Context, ra *models.ReviewApp, finalizers ...string) error
	RemoveFinalizersToReviewApp(ctx context.Context, ra *models.ReviewApp, finalizers ...string) error
}
