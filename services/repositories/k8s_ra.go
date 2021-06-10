package repositories

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sReviewAppClientFactory interface {
	NewRepository(client.Client, logr.Logger) (K8sReviewAppClientRepository, error)
}

type K8sReviewAppClientRepository interface {
	ApplyReviewAppInstance(ctx context.Context, namespacedName client.ObjectKey) error
	GetSecretValue(ctx context.Context, namespacedName client.ObjectKey, key string) (string, error)
}
