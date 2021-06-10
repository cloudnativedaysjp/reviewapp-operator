package k8s_rai_client

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type KubernetesFactoryImpl struct{}

func (kfi KubernetesFactoryImpl) NewRepository(c client.Client, l logr.Logger) (repositories.K8sReviewAppInstanceClientRepository, error) {
	return NewKubernetesInfra(c, l)
}
