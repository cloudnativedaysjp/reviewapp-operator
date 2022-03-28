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

func (c Client) GetReviewAppManager(ctx context.Context, namespace, name string) (*dreamkastv1alpha1.ReviewAppManager, error) {
	var ram dreamkastv1alpha1.ReviewAppManager
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	if err := c.Get(ctx, nn, &ram); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ram), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.NewK8sObjectNotFound(wrapedErr, ram.GVK(), nn)
		}
		return nil, wrapedErr
	}
	ram.SetGroupVersionKind(ram.GVK())
	return &ram, nil
}

func (c Client) UpdateReviewAppManagerStatus(ctx context.Context, ram models.ReviewAppManager) error {
	var ramCurrent dreamkastv1alpha1.ReviewAppManager
	nn := types.NamespacedName{Name: ram.Name, Namespace: ram.Namespace}
	if err := c.Get(ctx, nn, &ramCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ramCurrent), err)
		if apierrors.IsNotFound(err) {
			return myerrors.NewK8sObjectNotFound(wrapedErr, ramCurrent.GVK(), nn)
		}
		return wrapedErr
	}
	ramCurrent.Status = ram.Status
	if err := c.Status().Update(ctx, &ramCurrent); err != nil {
		return xerrors.Errorf("Error to Update %s: %w", reflect.TypeOf(ramCurrent), err)
	}
	return nil
}
