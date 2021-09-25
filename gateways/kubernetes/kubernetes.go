package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	errors_ "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
)

type KubernetesGateway struct {
	client.Client
	logger logr.Logger
}

func NewKubernetesGateway(c client.Client, l logr.Logger) (*KubernetesGateway, error) {
	// this app depend on reviewapp-operator CRD.
	// return error if ReviewAppManager and ReviewApp CRD is not installed
	var ramList dreamkastv1beta1.ReviewAppManagerList
	if err := c.List(context.Background(), &ramList); err != nil {
		return nil, err
	}
	var raList dreamkastv1beta1.ReviewAppList
	if err := c.List(context.Background(), &raList); err != nil {
		return nil, err
	}

	// this app depend on ArgoCD.
	// return error if ArgoCD Application CRD is not installed
	var a argocd_application_v1alpha1.ApplicationList
	if err := c.List(context.Background(), &a); err != nil {
		return nil, err
	}

	return &KubernetesGateway{c, l}, nil
}

// ReviewApp

func (g *KubernetesGateway) GetReviewAppConfig(ctx context.Context, namespace, name string, isCandidate bool) (*models.ReviewAppConfig, error) {
	var ram dreamkastv1beta1.ReviewAppManager
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ram); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(ram), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(ram)))
			return nil, models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	rac := models.NewReviewAppConfig()
	rac.ReviewAppManager = ram

	// sync ApplicationTemplate
	if err := g.syncApplicationTemplate(ctx, rac, ram.Spec.InfraConfig.ArgoCDApp.Template.Namespace, ram.Spec.InfraConfig.ArgoCDApp.Template.Name, isCandidate); err != nil {
		return nil, err
	}

	// sync ManifestsTemplate
	for _, nn := range ram.Spec.InfraConfig.Manifests.Templates {
		if err := g.syncManifestsTemplate(ctx, rac, nn.Namespace, nn.Name, isCandidate); err != nil {
			return nil, err
		}
	}

	return rac, nil
}

func (g *KubernetesGateway) syncApplicationTemplate(ctx context.Context, rac *models.ReviewAppConfig, namespace, name string, isCandidate bool) error {
	var at dreamkastv1beta1.ApplicationTemplate
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &at); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(at), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(at)))
			return models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}

	if isCandidate {
		rac.ApplicationTemplate = at.Spec.CandidateTemplate
	} else {
		rac.ApplicationTemplate = at.Spec.StableTemplate
	}
	return nil
}

func (g *KubernetesGateway) syncManifestsTemplate(ctx context.Context, rac *models.ReviewAppConfig, namespace, name string, isCandidate bool) error {
	var mt dreamkastv1beta1.ManifestsTemplate
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &mt); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(mt), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(mt)))
			return models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}

	var data map[string]string
	if isCandidate {
		data = mt.Spec.CandidateData
	} else {
		data = mt.Spec.StableData
	}
	for key, val := range data {
		rac.ManifestsTemplate[key] = val
	}
	return nil
}

func (g *KubernetesGateway) UpdateReviewAppManagerStatus(ctx context.Context, rac *models.ReviewAppConfig) error {
	var ramCurrent dreamkastv1beta1.ReviewAppManager
	if err := g.Client.Get(ctx, types.NamespacedName{Name: rac.ReviewAppManager.Name, Namespace: rac.ReviewAppManager.Namespace}, &ramCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to get %s: %w", reflect.TypeOf(ramCurrent), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(ramCurrent)))
			return models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}
	patch := client.MergeFrom(&ramCurrent)

	ram := *rac.ReviewAppManager.DeepCopy()
	if err := g.Status().Patch(ctx, &ram, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(ram), err)
	}
	return nil
}

func (g *KubernetesGateway) AddFinalizersToReviewApp(ctx context.Context, raModel *models.ReviewApp, finalizers ...string) error {
	ra := *raModel.DeepCopy()
	raPatched := *raModel.DeepCopy()
	for _, f := range finalizers {
		controllerutil.AddFinalizer(&raPatched, f)
	}
	patch := client.MergeFrom(&ra)
	if err := g.Patch(ctx, &raPatched, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raPatched), err)
	}
	return nil
}

func (g *KubernetesGateway) RemoveFinalizersToReviewApp(ctx context.Context, raModel *models.ReviewApp, finalizers ...string) error {
	ra := *raModel.DeepCopy()
	raPatched := *raModel.DeepCopy()
	for _, f := range finalizers {
		if controllerutil.ContainsFinalizer(&raPatched, f) {
			controllerutil.RemoveFinalizer(&raPatched, f)
		}
	}
	patch := client.MergeFrom(&ra)
	if err := g.Patch(ctx, &raPatched, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raPatched), err)
	}
	return nil
}

// ReviewApp

func (g *KubernetesGateway) GetReviewApp(ctx context.Context, namespace, name string) (*models.ReviewApp, error) {
	var ra dreamkastv1beta1.ReviewApp
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &ra); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ra), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(ra)))
			return nil, models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &models.ReviewApp{ReviewApp: ra}, nil
}

func (g *KubernetesGateway) GetReviewAppManagerFromReviewApp(ctx context.Context, ra *models.ReviewApp) (*models.ReviewAppConfig, error) {
	var ram dreamkastv1beta1.ReviewAppManager

	// get RAM name/namespace in OwnerReferences
	gvk := ram.GroupVersionKind()
	var ramName string
	for _, or := range ra.OwnerReferences {
		if or.APIVersion == gvk.Version && or.Kind == gvk.Kind {
			ramName = or.Name
		}
	}

	if err := g.Client.Get(ctx, types.NamespacedName{Name: ramName, Namespace: ra.Namespace}, &ram); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(ram), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(ra)))
			return nil, models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return nil, wrapedErr
	}
	return &models.ReviewAppConfig{ReviewAppManager: ram}, nil
}

func (g *KubernetesGateway) ApplyReviewAppWithOwnerRef(ctx context.Context, ra *models.ReviewApp, owner metav1.Object) error {
	tmp := &dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ra.ReviewApp.Name,
			Namespace: ra.ReviewApp.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, g.Client, tmp, func() (err error) {
		tmp.Spec = ra.Spec
		if err := ctrl.SetControllerReference(owner, tmp, g.Scheme()); err != nil {
			g.logger.Error(err, "unable to set ownerReference from ReviewAppManager to ReviewApp")
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (g *KubernetesGateway) UpdateReviewAppStatus(ctx context.Context, raModel *models.ReviewApp) error {
	var raCurrent dreamkastv1beta1.ReviewApp
	if err := g.Client.Get(ctx, types.NamespacedName{Name: raModel.Name, Namespace: raModel.Namespace}, &raCurrent); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(raCurrent), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(raCurrent)))
			return models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return wrapedErr
	}
	patch := client.MergeFrom(&raCurrent)

	ra := *raModel.DeepCopy()
	if err := g.Status().Patch(ctx, &ra, patch); err != nil {
		return xerrors.Errorf("Error to Patch %s: %w", reflect.TypeOf(raCurrent), err)
	}
	return nil
}

func (g *KubernetesGateway) DeleteReviewApp(ctx context.Context, namespace, name string) error {
	ra := dreamkastv1beta1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if err := g.Delete(ctx, &ra); err != nil {
		return xerrors.Errorf("Error to Delete %s: %w", reflect.TypeOf(ra), err)
	}
	return nil
}

// Secret

func (g *KubernetesGateway) GetSecretValue(ctx context.Context, namespace, name, key string) (string, error) {
	// get value from secret
	var secret corev1.Secret
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		wrapedErr := xerrors.Errorf("Error to Delete %s: %w", reflect.TypeOf(secret), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(secret)))
			return "", models.K8sRsourceNotFound{Err: wrapedErr}
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

// ArgoCD Application

func (g *KubernetesGateway) GetAnnotationOfArgoCDApplication(ctx context.Context, namespace, name, annotationKey string) (string, error) {
	var a argocd_application_v1alpha1.Application
	if err := g.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &a); err != nil {
		wrapedErr := xerrors.Errorf("Error to Get %s: %w", reflect.TypeOf(a), err)
		if errors_.IsNotFound(err) {
			g.logger.Info(fmt.Sprintf("%s not found", reflect.TypeOf(a)))
			return "", models.K8sRsourceNotFound{Err: wrapedErr}
		}
		return "", wrapedErr
	}
	return a.Annotations[annotationKey], nil
}

func (g *KubernetesGateway) PrintArgoCDApplicationWithAnnotation(ctx context.Context, application models.ArgoCDApplicationString, annotationKey, annotationValue string) (models.ArgoCDApplicationString, error) {
	var a argocd_application_v1alpha1.Application
	err := yaml.Unmarshal([]byte(application), &a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	if a.Annotations == nil {
		a.Annotations = map[string]string{}
	}
	a.Annotations[annotationKey] = annotationValue

	b, err := yaml.Marshal(&a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return models.ArgoCDApplicationString(b), nil
}

func (g *KubernetesGateway) PrintArgoCDApplicationName(ctx context.Context, application models.ArgoCDApplicationString) (string, error) {
	var a argocd_application_v1alpha1.Application
	err := yaml.Unmarshal([]byte(application), &a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return a.Name, nil
}

func (g *KubernetesGateway) PrintArgoCDApplicationNamespace(ctx context.Context, application models.ArgoCDApplicationString) (string, error) {
	var a argocd_application_v1alpha1.Application
	err := yaml.Unmarshal([]byte(application), &a)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return a.Namespace, nil
}
