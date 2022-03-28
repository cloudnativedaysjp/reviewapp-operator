package kubernetes

import (
	"context"
	"reflect"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

func (c Client) GetSecretValue(ctx context.Context, namespace string, m models.AppOrInfraRepoTarget) (string, error) {
	// get value from secret
	var secret corev1.Secret
	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	}
	secretRef, err := m.GetGitSecretRef()
	if err != nil {
		// Secret is not set
		return "", nil
	}
	nn := types.NamespacedName{Namespace: namespace, Name: secretRef.Name}
	if err := c.Get(ctx, nn, &secret); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(secret), err)
		if apierrors.IsNotFound(err) {
			return "", myerrors.NewK8sObjectNotFound(wrapedErr, gvk, nn)
		}
		return "", wrapedErr
	}
	d, ok := secret.Data[secretRef.Key]
	if !ok {
		return "", xerrors.Errorf("Secret %s does not have key %s", secretRef.Name, secretRef.Key)
	}
	return string(d), nil
}
