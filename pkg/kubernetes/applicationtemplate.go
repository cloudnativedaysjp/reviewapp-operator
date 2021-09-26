package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
)

func GetApplicationTemplate(ctx context.Context, c client.Client, namespace, name string) (*dreamkastv1beta1.ApplicationTemplate, error) {
	var at dreamkastv1beta1.ApplicationTemplate
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &at); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(at), err)
		if apierrors.IsNotFound(err) {
			return nil, myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &at, nil
}
