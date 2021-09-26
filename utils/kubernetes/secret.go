package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func GetSecretValue(ctx context.Context, c client.Client, namespace, name, key string) (string, error) {
	// get value from secret
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		wrapedErr := xerrors.Errorf("Error to Delete %s: %w", reflect.TypeOf(secret), err)
		if apierrors.IsNotFound(err) {
			return "", myerrors.K8sResourceNotFound{Err: wrapedErr}
		}
		return "", wrapedErr
	}
	d, ok := secret.Data[key]
	if !ok {
		return "", xerrors.Errorf("Secret %s does not have key %s", name, key)
	}
	// base64 decode
	return string(d), nil
}
