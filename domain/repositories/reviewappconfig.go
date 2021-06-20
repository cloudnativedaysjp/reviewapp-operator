package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ReviewAppConfig struct {
	Iface ReviewAppConfigIFace
}

type ReviewAppConfigIFace interface {
	GetReviewAppConfig(ctx context.Context, namespace, name string) (*models.ReviewAppConfig, error)
	UpdateReviewAppStatus(ctx context.Context, rac *models.ReviewAppConfig) error
}

func NewReviewAppConfig(iface ReviewAppConfigIFace) *ReviewAppConfig {
	return &ReviewAppConfig{iface}
}
