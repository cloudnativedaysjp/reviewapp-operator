package k8s_ra_client

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type KubernetesFactoryImpl struct{}

func (kfi KubernetesFactoryImpl) NewRepository(c client.Client, l logr.Logger) (repositories.K8sReviewAppClientRepository, error) {
	return NewKubernetesInfra(c, l)
}
