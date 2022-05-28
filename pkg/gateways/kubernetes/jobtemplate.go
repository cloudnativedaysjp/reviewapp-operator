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

func (c Client) GetPreStopJobTemplate(ctx context.Context, ra dreamkastv1alpha1.ReviewApp) (dreamkastv1alpha1.JobTemplate, error) {
	return c.getJobTemplate(ctx, ra.Spec.PreStopJob.Namespace, ra.Spec.PreStopJob.Name)
}

func (c Client) getJobTemplate(ctx context.Context, namespace, name string) (dreamkastv1alpha1.JobTemplate, error) {
	var jt dreamkastv1alpha1.JobTemplate
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	if err := c.Get(ctx, nn, &jt); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(jt), err)
		if apierrors.IsNotFound(err) {
			return dreamkastv1alpha1.JobTemplate{}, myerrors.NewK8sObjectNotFound(wrapedErr, jt.GVK(), nn)
		}
		return dreamkastv1alpha1.JobTemplate{}, wrapedErr
	}
	jt.SetGroupVersionKind(jt.GVK())
	return dreamkastv1alpha1.JobTemplate(jt), nil
}
