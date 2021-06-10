package k8s_ra_client

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

type KubernetesInfra struct {
	client.Client
	Log logr.Logger
}

func NewKubernetesInfra(c client.Client, l logr.Logger) (*KubernetesInfra, error) {
	// this app depend on reviewapp-operator CRD.
	// return error if ReviewApp and ReviewAppInstance CRD is not defined
	var raList dreamkastv1beta1.ReviewAppList
	if err := c.List(context.Background(), &raList); err != nil {
		return nil, err
	}
	var raiList dreamkastv1beta1.ReviewAppInstanceList
	if err := c.List(context.Background(), &raiList); err != nil {
		return nil, err
	}

	return &KubernetesInfra{c, l}, nil
}

func (ki *KubernetesInfra) ApplyReviewAppInstance(ctx context.Context, namespacedName client.ObjectKey) error {
	reviewApp := dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}
	reviewAppInstance := dreamkastv1beta1.ReviewAppInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, ki.Client, &reviewAppInstance, func() (err error) {
		// TODO
		//reviewAppInstance.Spec =

		if err != nil {
			return err
		}
		if err := ctrl.SetControllerReference(&reviewApp, &reviewAppInstance, ki.Scheme()); err != nil {
			ki.Log.Error(err, "unable to set ownerReference from ReviewApp to ReviewAppInstance")
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (ki *KubernetesInfra) GetSecretValue(ctx context.Context, namespacedName client.ObjectKey, key string) (string, error) {
	// get value from secret
	var secret corev1.Secret
	if err := ki.Client.Get(ctx, namespacedName, &secret); err != nil {
		return "", err
	}
	d, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("Secret %s does not have key %s", namespacedName.Name, key)
	}
	// base64 decode
	return string(d), nil
}
