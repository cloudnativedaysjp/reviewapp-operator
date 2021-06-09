package repositories

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesFactory interface {
	NewRepository(client.Client, logr.Logger) KubernetesRepository
}

type KubernetesRepository interface {
	// TODO
	GetArgoCDApplicationStatus()
}
