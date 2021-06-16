package repositories

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

type KubernetesFactory interface {
	NewRepository(client.Client, logr.Logger) (KubernetesRepository, error)
}

type KubernetesRepository interface {
	ApplyReviewAppInstanceFromReviewApp(ctx context.Context, rai *dreamkastv1beta1.ReviewAppInstance, ra *dreamkastv1beta1.ReviewApp) error
	GetSecretValue(ctx context.Context, namespacedName client.ObjectKey, key string) (string, error)
	GetArgoCDApplicationStatus(ctx context.Context, namespacedName client.ObjectKey) (*ArgoCDStatusOutput, error)
	GetApplicationTemplate(ctx context.Context, namespacedName client.ObjectKey) (*dreamkastv1beta1.ApplicationTemplate, error)
	GetManifestTemplate(ctx context.Context, namespacedName client.ObjectKey) (*dreamkastv1beta1.ManifestsTemplate, error)
	UpdateReviewAppStatus(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) error
	DeleteReviewAppInstance(ctx context.Context, namespacedName client.ObjectKey) error
}
