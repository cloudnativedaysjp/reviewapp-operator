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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type KubernetesGateway struct {
	client.Client
	Log logr.Logger
}

func NewKubernetesGateway(c client.Client, l logr.Logger) (*KubernetesGateway, error) {
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

	return &KubernetesGateway{c, l}, nil
}

func (g *KubernetesGateway) GetReviewAppConfig(ctx context.Context, namespace, name string) (*models.ReviewAppConfig, error) {
	var ra dreamkastv1beta1.ReviewApp
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ra); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(ra)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	rac := new(models.ReviewAppConfig)
	rac.ReviewApp = ra

	// sync ApplicationTemplate
	if err := g.syncApplicationTemplate(ctx, rac, ra.Spec.Infra.ArgoCDApp.Template.Namespace, ra.Spec.Infra.ArgoCDApp.Template.Name); err != nil {
		return nil, err
	}

	// sync ManifestsTemplate
	for _, nn := range ra.Spec.Infra.Manifests.Templates {
		if err := g.syncManifestsTemplate(ctx, rac, nn.Namespace, nn.Name); err != nil {
			return nil, err
		}
	}

	// return
	return rac, nil
}

func (g *KubernetesGateway) syncApplicationTemplate(ctx context.Context, rac *models.ReviewAppConfig, name string, namespace string) error {
	var at dreamkastv1beta1.ApplicationTemplate
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &at); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(at)))
			return err
		}
		return client.IgnoreNotFound(err)
	}

	rac.ApplicationTemplate = at
	return nil
}

func (g *KubernetesGateway) syncManifestsTemplate(ctx context.Context, rac *models.ReviewAppConfig, namespace, name string) error {
	var mt dreamkastv1beta1.ManifestsTemplate
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &mt); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(mt)))
			return err
		}
		return client.IgnoreNotFound(err)
	}

	rac.ManifestsTemplate.Add(mt)
	return nil
}

func (g *KubernetesGateway) UpdateReviewAppStatus(ctx context.Context, rac *models.ReviewAppConfig) error {
	var ra dreamkastv1beta1.ReviewApp
	ra = *rac.ReviewApp.DeepCopy()
	if err := g.Status().Update(ctx, &ra); err != nil {
		g.Log.Error(err, err.Error())
		return err
	}
	return nil
}

func (g *KubernetesGateway) GetReviewAppInstance(ctx context.Context, namespace, name string) (*models.ReviewAppInstance, error) {
	var rai dreamkastv1beta1.ReviewAppInstance
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &rai); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(rai)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &models.ReviewAppInstance{rai}, nil
}

func (g *KubernetesGateway) ApplyReviewAppInstanceWithOwnerRef(ctx context.Context, rai models.ReviewAppInstance, owner metav1.Object) error {
	tmp := &dreamkastv1beta1.ReviewAppInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rai.ReviewAppInstance.Name,
			Namespace: rai.ReviewAppInstance.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, g.Client, tmp, func() (err error) {
		tmp.Spec = rai.ReviewAppInstance.Spec
		if err := ctrl.SetControllerReference(owner, tmp, g.Scheme()); err != nil {
			g.Log.Error(err, "unable to set ownerReference from ReviewApp to ReviewAppInstance")
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (g *KubernetesGateway) DeleteReviewAppInstance(ctx context.Context, namespace, name string) error {
	rai := dreamkastv1beta1.ReviewAppInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := g.Delete(ctx, &rai); err != nil {
		g.Log.Error(err, err.Error())
		return err
	}
	return nil
}

func (g *KubernetesGateway) GetArgoCDApplication(ctx context.Context, namespace, name string) (*models.ArgoCDApplication, error) {
	var a argocd_application_v1alpha1.Application
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &a); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(a)))
			return nil, err
		}
		return nil, client.IgnoreNotFound(err)
	}
	return &models.ArgoCDApplication{
		ObjectMeta: a.ObjectMeta,
		Status:     string(a.Status.Health.Status),
	}, nil

}

func (g *KubernetesGateway) SyncArgoCDApplicationStatus(ctx context.Context, app *models.ArgoCDApplication) error {
	var a argocd_application_v1alpha1.Application
	if err := g.Client.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, &a); err != nil {
		if errors_.IsNotFound(err) {
			g.Log.Info(fmt.Sprintf("%s not found", reflect.TypeOf(a)))
			return err
		}
		return client.IgnoreNotFound(err)
	}
	app.Status = string(a.Status.Health.Status)
	return nil
}

func (g *KubernetesGateway) GetSecretValue(ctx context.Context, namespace, name, key string) (string, error) {
	// get value from secret
	var secret corev1.Secret
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		return "", err
	}
	d, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("Secret %s does not have key %s", name, key)
	}
	// base64 decode
	return string(d), nil
}
