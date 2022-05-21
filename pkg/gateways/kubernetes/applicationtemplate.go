package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
)

func (c Client) GetApplicationTemplate(ctx context.Context, spec dreamkastv1alpha1.ReviewAppCommonSpec) (dreamkastv1alpha1.ApplicationTemplate, error) {
	var at dreamkastv1alpha1.ApplicationTemplate
	conf := spec.InfraConfig
	nn := types.NamespacedName{Name: conf.ArgoCDApp.Template.Name, Namespace: conf.ArgoCDApp.Template.Namespace}
	if err := c.Get(ctx, nn, &at); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(at), err)
		if apierrors.IsNotFound(err) {
			return dreamkastv1alpha1.ApplicationTemplate{}, myerrors.NewK8sObjectNotFound(wrapedErr, at.GVK(), nn)
		}
		return dreamkastv1alpha1.ApplicationTemplate{}, wrapedErr
	}
	at.SetGroupVersionKind(at.GVK())
	return at, nil
}
