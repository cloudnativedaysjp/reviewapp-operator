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

func (c Client) GetReviewAppManager(ctx context.Context, namespace, name string) (dreamkastv1alpha1.ReviewAppManager, error) {
	var ram dreamkastv1alpha1.ReviewAppManager
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	if err := c.Get(ctx, nn, &ram); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ram), err)
		if apierrors.IsNotFound(err) {
			return dreamkastv1alpha1.ReviewAppManager{}, myerrors.NewK8sObjectNotFound(wrapedErr, ram.GVK(), nn)
		}
		return dreamkastv1alpha1.ReviewAppManager{}, wrapedErr
	}
	ram.SetGroupVersionKind(ram.GVK())
	return ram, nil
}
