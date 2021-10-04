package kubernetes

import (
	"context"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func GetArgoCDAppAnnotation(ctx context.Context, c client.Client, namespace, name, annotationKey string) (string, error) {
	var a argocd_application_v1alpha1.Application
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &a); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(a), err)
		if apierrors.IsNotFound(err) {
			return "", myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return "", wrapedErr
	}
	return a.Annotations[annotationKey], nil
}
