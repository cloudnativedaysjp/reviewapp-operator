package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	errors_ "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/services/repositories"
)

type KubernetesInfra struct {
	client.Client
	Log logr.Logger
}

func NewKubernetesInfra(c client.Client, l logr.Logger) (*KubernetesInfra, error) {
	// this app depend on reviewapp-operator CRD.
	// return error if ReviewApp and ReviewAppInstance CRD is not installed
	var raList dreamkastv1beta1.ReviewAppList
	if err := c.List(context.Background(), &raList); err != nil {
		return nil, err
	}
	var raiList dreamkastv1beta1.ReviewAppInstanceList
	if err := c.List(context.Background(), &raiList); err != nil {
		return nil, err
	}

	// this app depend on ArgoCD.
	// return error if ArgoCD Application CRD is not installed
	// TODO: list に失敗する
	var a argocd_application_v1alpha1.ApplicationList
	if err := c.List(context.Background(), &a); err != nil {
		return nil, err
	}

	return &KubernetesInfra{c, l}, nil
}

func (ki *KubernetesInfra) ApplyReviewAppInstanceFromReviewApp(ctx context.Context, rai *dreamkastv1beta1.ReviewAppInstance, ra *dreamkastv1beta1.ReviewApp) error {
	if _, err := ctrl.CreateOrUpdate(ctx, ki.Client, rai, func() (err error) {
		if err := ctrl.SetControllerReference(ra, rai, ki.Scheme()); err != nil {
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

func (ki *KubernetesInfra) GetArgoCDApplicationStatus(ctx context.Context, namespacedName client.ObjectKey) (*repositories.ArgoCDStatusOutput, error) {
	var a argocd_application_v1alpha1.Application
	if err := ki.Client.Get(ctx, namespacedName, &a); err != nil {
		if errors_.IsNotFound(err) {
			ki.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(a)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &repositories.ArgoCDStatusOutput{
		Status: string(a.Status.Health.Status),
		// TODO
	}, nil
}

func (ki *KubernetesInfra) UpdateReviewAppStatus(ctx context.Context, ra *dreamkastv1beta1.ReviewApp) error {
	if err := ki.Status().Update(ctx, ra); err != nil {
		ki.Log.Error(err, err.Error())
		return err
	}
	return nil
}

func (ki *KubernetesInfra) DeleteReviewAppInstance(ctx context.Context, namespacedName client.ObjectKey) error {
	rai := dreamkastv1beta1.ReviewAppInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
	}
	if err := ki.Delete(ctx, &rai); err != nil {
		ki.Log.Error(err, err.Error())
		return err
	}
	return nil
}

func (ki *KubernetesInfra) GetApplicationTemplate(ctx context.Context, namespacedName client.ObjectKey) (*dreamkastv1beta1.ApplicationTemplate, error) {
	var at dreamkastv1beta1.ApplicationTemplate
	if err := ki.Client.Get(ctx, namespacedName, &at); err != nil {
		if errors_.IsNotFound(err) {
			ki.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(at)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &at, nil
}

func (ki *KubernetesInfra) GetManifestTemplate(ctx context.Context, namespacedName client.ObjectKey) (*dreamkastv1beta1.ManifestsTemplate, error) {
	var mt dreamkastv1beta1.ManifestsTemplate
	if err := ki.Client.Get(ctx, namespacedName, &mt); err != nil {
		if errors_.IsNotFound(err) {
			ki.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(mt)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &mt, nil
}
