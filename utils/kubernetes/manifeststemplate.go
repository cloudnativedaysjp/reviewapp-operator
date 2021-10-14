package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func GetManifestsTemplate(ctx context.Context, c client.Client, namespace, name string) (*dreamkastv1alpha1.ManifestsTemplate, error) {
	var mt dreamkastv1alpha1.ManifestsTemplate
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &mt); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(mt), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &mt, nil
}
