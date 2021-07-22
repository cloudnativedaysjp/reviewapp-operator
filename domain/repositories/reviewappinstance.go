package repositories

import (
	"context"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReviewAppInstanceIFace interface {
	GetReviewAppInstance(ctx context.Context, namespace, name string) (*models.ReviewAppInstance, error)
	ApplyReviewAppInstanceWithOwnerRef(ctx context.Context, rai models.ReviewAppInstance, owner metav1.Object) error
	DeleteReviewAppInstance(ctx context.Context, namespace, name string) error
}
