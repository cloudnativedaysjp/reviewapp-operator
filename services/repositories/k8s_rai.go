package repositories

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sReviewAppInstanceClientFactory interface {
	NewRepository(client.Client, logr.Logger) (K8sReviewAppInstanceClientRepository, error)
}

type K8sReviewAppInstanceClientRepository interface {
	GetArgoCDApplicationStatus(ctx context.Context, namespacedName client.ObjectKey) (*ArgoCDStatusOutput, error)
}
