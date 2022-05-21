/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/base64"

	"golang.org/x/xerrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	myerrors "github.com/cloudnativedaysjp/reviewapp-operator/pkg/errors"
)

// ApplicationTemplateSpec defines the desired state of ApplicationTemplate
type ApplicationTemplateSpec struct {

	// CandidateTemplate is included ArgoCD Application manifest. (apiVersion, kind, metadata, spec, ...)
	CandidateTemplate Application `json:"candidate,omitempty"`

	// StableTemplate is included ArgoCD Application manifest. (apiVersion, kind, metadata, spec, ...)
	StableTemplate Application `json:"stable,omitempty"`
}

const (
	AnnotationAppOrgNameForArgoCDApplication    = "dreamkast.cloudnativedays.jp/app-organization"
	AnnotationAppRepoNameForArgoCDApplication   = "dreamkast.cloudnativedays.jp/app-repository"
	AnnotationAppCommitHashForArgoCDApplication = "dreamkast.cloudnativedays.jp/app-commit-hash"
)

type Application string
type ApplicationBase64 string

func (m Application) NamespacedName() (types.NamespacedName, error) {
	var obj unstructured.Unstructured
	err := yaml.Unmarshal([]byte(m), &obj)
	if err != nil {
		return types.NamespacedName{}, xerrors.Errorf("%w", err)
	}
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}, nil
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

func (m Application) SetSomeAnnotations(ra ReviewApp) (Application, error) {
	appWithAnnotations := m
	appWithAnnotations, err := appWithAnnotations.setAnnotation(
		AnnotationAppOrgNameForArgoCDApplication, ra.Spec.AppTarget.Organization)
	if err != nil {
		return "", err
	}
	appWithAnnotations, err = appWithAnnotations.setAnnotation(
		AnnotationAppRepoNameForArgoCDApplication, ra.Spec.AppTarget.Repository)
	if err != nil {
		return "", err
	}
	appWithAnnotations, err = appWithAnnotations.setAnnotation(
		AnnotationAppCommitHashForArgoCDApplication, ra.Status.PullRequestCache.LatestCommitHash)
	if err != nil {
		return "", err
	}
	return appWithAnnotations, nil
}

func (m Application) setAnnotation(annotationKey, annotationValue string) (Application, error) {
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

func (m Application) ToBase64() ApplicationBase64 {
	return ApplicationBase64(base64.StdEncoding.EncodeToString([]byte(m)))
}

func (m ApplicationBase64) Decode() (Application, error) {
	app, err := base64.URLEncoding.DecodeString(string(m))
	return Application(app), err
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=at

// ApplicationTemplate is the Schema for the applicationtemplates API
type ApplicationTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ApplicationTemplateSpec `json:"spec"`
}

func (ApplicationTemplate) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "ApplicationTemplate",
	}
}

//+kubebuilder:object:root=true

// ApplicationTemplateList contains a list of ApplicationTemplate
type ApplicationTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationTemplate{}, &ApplicationTemplateList{})
}
