package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetReviewApp(ctx context.Context, namespace, name string) (*dreamkastv1alpha1.ReviewApp, error) {
	var ra dreamkastv1alpha1.ReviewApp
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	if err := c.Get(ctx, nn, &ra); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ra), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.NewK8sObjectNotFound(wrapedErr, &ra, nn)
		}
		return nil, wrapedErr
	}
	return &ra, nil
}

func (c Client) ApplyReviewAppWithOwnerRef(ctx context.Context, ra models.ReviewApp, owner metav1.Object) error {
	raApplied := &dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ra.Name,
			Namespace: ra.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, c, raApplied, func() (err error) {
		raApplied.Spec = ra.Spec
		if err := ctrl.SetControllerReference(owner, raApplied, c.Scheme()); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (c Client) UpdateReviewAppStatus(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp) error {
	var raCurrent dreamkastv1alpha1.ReviewApp
	nn := types.NamespacedName{Name: ra.Name, Namespace: ra.Namespace}
	if err := c.Get(ctx, nn, &raCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(raCurrent), err)
		if apierrors.IsNotFound(err) {
			return myerrors.NewK8sObjectNotFound(wrapedErr, &raCurrent, nn)
		}
		return wrapedErr
	}

	raCurrent.Status = ra.Status
	if err := c.Status().Update(ctx, &raCurrent); err != nil {
		return xerrors.Errorf("Error to Update %s: %w", reflect.TypeOf(raCurrent), err)
	}
	return nil
}

func (c Client) DeleteReviewApp(ctx context.Context, namespace, name string) error {
	ra := dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, &ra); err != nil {
		return xerrors.Errorf("Error to Delete %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

func (c Client) AddFinalizersToReviewApp(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dreamkast.cloudnativedays.jp",
		Version: "v1alpha1",
		Kind:    "ReviewApp",
	})
	patch.SetNamespace(ra.Namespace)
	patch.SetName(ra.Name)
	patch.SetFinalizers(ra.Finalizers)
	for _, f := range finalizers {
		controllerutil.AddFinalizer(patch, f)
	}

	if err := c.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "reviewapp-operator",
		Force:        pointer.Bool(true),
	}); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

func (c Client) RemoveFinalizersFromReviewApp(ctx context.Context, ra *dreamkastv1alpha1.ReviewApp, finalizers ...string) error {
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "dreamkast.cloudnativedays.jp",
		Version: "v1alpha1",
		Kind:    "ReviewApp",
	})
	patch.SetNamespace(ra.Namespace)
	patch.SetName(ra.Name)
	patch.SetFinalizers(ra.Finalizers)
	for _, f := range finalizers {
		if controllerutil.ContainsFinalizer(patch, f) {
			controllerutil.RemoveFinalizer(patch, f)
		}
	}
	if err := c.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: "reviewapp-operator",
		Force:        pointer.Bool(true),
	}); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}
