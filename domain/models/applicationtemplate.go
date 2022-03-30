package models

import (
	"golang.org/x/xerrors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/errors"
)

/* ApplicationTemplate */

type ApplicationTemplate dreamkastv1alpha1.ApplicationTemplate

func (m ApplicationTemplate) StableStr() string {
	return m.Spec.StableTemplate
}

func (m ApplicationTemplate) CandidateStr() string {
	return m.Spec.CandidateTemplate
}

func (m ApplicationTemplate) GenerateApplication(pr PullRequest, v Templator) (Application, error) {
	var template string
	var err error
	if pr.IsCandidate() {
		template = m.CandidateStr()
	} else {
		template = m.StableStr()
	}
	application, err := v.Templating(template)
	if err != nil {
		return "", err
	}
	return Application(application), nil
}

/* Application */

type Application string

func (m Application) NamespacedName() (types.NamespacedName, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(m), &obj)
	if err != nil {
		return types.NamespacedName{}, xerrors.Errorf("%w", err)
	}
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, nil
}

func (m Application) SetAnnotation(annotationKey, annotationValue string) (Application, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(m), &obj)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[annotationKey] = annotationValue
	obj.SetAnnotations(annotations)

	b, err := yaml.Marshal(&obj)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	return Application(b), nil
}

func (m Application) Annotation(annotationKey string) (string, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(m), &obj)
	if err != nil {
		return "", xerrors.Errorf("%w", err)
	}
	val, ok := obj.GetAnnotations()[annotationKey]
	if !ok {
		return "", myerrors.NewKeyIsMissing("Application.metadata.annotations", annotationKey)
	}
	return val, nil
}
