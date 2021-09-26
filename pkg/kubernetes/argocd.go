package kubernetes

import (
	"context"
	"reflect"

	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"golang.org/x/xerrors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
)

const (
	AnnotationAppOrgNameForArgoCDApplication    = "dreamkast.cloudnativedays.jp/app-organization"
	AnnotationAppRepoNameForArgoCDApplication   = "dreamkast.cloudnativedays.jp/app-repository"
	AnnotationAppCommitHashForArgoCDApplication = "dreamkast.cloudnativedays.jp/app-commit-hash"
)

func PickNamespacedNameFromArgoCDAppStr(ctx context.Context, applicationStr string) (types.NamespacedName, error) {
	var a argocd_application_v1alpha1.Application
	err := yaml.Unmarshal([]byte(applicationStr), &a)
	if err != nil {
		return types.NamespacedName{}, xerrors.Errorf("%w", err)
	}
	return types.NamespacedName{Namespace: a.Namespace, Name: a.Name}, nil
}

func SetAnnotationToArgoCDAppStr(ctx context.Context, applicationStr string, annotationKey, annotationValue string) (string, error) {
	var a argocd_application_v1alpha1.Application
	err := yaml.Unmarshal([]byte(applicationStr), &a)
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
	return string(b), nil
}

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
