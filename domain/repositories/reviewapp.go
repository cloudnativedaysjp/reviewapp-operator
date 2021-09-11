package repositories

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type ReviewAppIFace interface {
	GetReviewApp(ctx context.Context, namespace, name string) (*models.ReviewApp, error)
	GetReviewAppManagerFromReviewApp(ctx context.Context, ra *models.ReviewApp) (*models.ReviewAppConfig, error)
	ApplyReviewAppWithOwnerRef(ctx context.Context, ra *models.ReviewApp, owner metav1.Object) error
	UpdateReviewAppStatus(ctx context.Context, ra *models.ReviewApp) error
	DeleteReviewApp(ctx context.Context, namespace, name string) error
}
