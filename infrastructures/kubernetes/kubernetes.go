package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
	"github.com/go-logr/logr"
	errors_ "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesInfra struct {
	client.Client
	Log logr.Logger
}

func NewKubernetesInfra(c client.Client, l logr.Logger) *KubernetesInfra {
	return &KubernetesInfra{c, l}
}

func (ki *KubernetesInfra) GetArgoCDApplicationStatus(ctx context.Context, namespacedName client.ObjectKey) (*repositories.ArgoCDStatusOutput, error) {
	var a argocd_application_v1alpha1.Application
	if err := ki.Client.Get(ctx, namespacedName, &a); err != nil {
		if errors_.IsNotFound(err) {
			ki.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(a)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &repositories.ArgoCDStatusOutput{
		Status: string(a.Status.Health.Status),
		// TODO
	}, nil
}
