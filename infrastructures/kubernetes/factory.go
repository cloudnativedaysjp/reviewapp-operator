package kubernetes

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type KubernetesFactoryImpl struct{}

func (kfi KubernetesFactoryImpl) NewRepository(c client.Client, l logr.Logger) (repositories.KubernetesRepository, error) {
	return NewKubernetesInfra(c, l)
}
