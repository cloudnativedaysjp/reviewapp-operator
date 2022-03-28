package kubernetes

import (
	"context"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetArgoCDAppFromReviewAppStatus(ctx context.Context, ra models.ReviewApp) (models.Application, error) {
	var a argocd_application_v1alpha1.Application
	nn := types.NamespacedName{Namespace: ra.Status.Sync.ApplicationNamespace, Name: ra.Status.Sync.ApplicationName}
	if err := c.Get(ctx, nn, &a); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(a), err)
		if apierrors.IsNotFound(err) {
			return "", myerrors.NewK8sObjectNotFound(wrapedErr, &a, nn)
		}
		return "", wrapedErr
	}
	b, err := yaml.Marshal(&a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return models.Application(b), nil
}
