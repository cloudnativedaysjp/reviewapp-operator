package kubernetes

import (
	"context"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetArgoCDAppFromReviewAppStatus(ctx context.Context, raStatus models.ReviewAppStatus) (models.Application, error) {
	var a argocd_application_v1alpha1.Application
	gvk := schema.GroupVersionKind{
		Group:   argocd_application_v1alpha1.SchemeGroupVersion.Group,
		Version: argocd_application_v1alpha1.SchemeGroupVersion.Version,
		Kind:    "Application",
	}
	nn := types.NamespacedName{Namespace: raStatus.Sync.ApplicationNamespace, Name: raStatus.Sync.ApplicationName}
	if err := c.Get(ctx, nn, &a); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(a), err)
		if apierrors.IsNotFound(err) {
			return "", myerrors.NewK8sObjectNotFound(wrapedErr, gvk, nn)
		}
		return "", wrapedErr
	}
	a.SetGroupVersionKind(gvk)
	b, err := yaml.Marshal(&a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return models.Application(b), nil
}
