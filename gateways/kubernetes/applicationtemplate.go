package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetApplicationTemplate(ctx context.Context, m models.ReviewAppOrReviewAppManager) (models.ApplicationTemplate, error) {
	var at dreamkastv1alpha1.ApplicationTemplate
	conf := m.InfraRepoConfig()
	nn := types.NamespacedName{Name: conf.ArgoCDApp.Template.Name, Namespace: conf.ArgoCDApp.Template.Namespace}
	if err := c.Get(ctx, nn, &at); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(at), err)
		if apierrors.IsNotFound(err) {
			return models.ApplicationTemplate{}, myerrors.NewK8sObjectNotFound(wrapedErr, at.GVK(), nn)
		}
		return models.ApplicationTemplate{}, wrapedErr
	}
	at.SetGroupVersionKind(at.GVK())
	return models.ApplicationTemplate(at), nil
}
