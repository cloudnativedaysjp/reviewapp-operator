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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReviewAppSpec defines the desired state of ReviewApp
type ReviewAppSpec struct {

	// App is config of application repository
	App ReviewAppSpecApp `json:"app,omitempty"`

	// Infra is config of manifest repository
	Infra ReviewAppSpecInfra `json:"infra,omitempty"`

	// Variables is available to use input of Application & Manifest Template
	Variables []string `json:"variables,omitempty"`
}

type ReviewAppSpecApp struct {

	// TODO
	Repository string `json:"repository,omitempty"`

	// GitSecretRef is specifying secret for accessing Git remote-repo
	GitSecretRef *corev1.SecretKeySelector `json:"gitSecretRef,omitempty"`

	// IgnoreLabels is TODO
	IgnoreLabels []string `json:"ignoreLabels,omitempty"`

	// IgnoreTitleExp is TODO
	IgnoreTitleExp string `json:"ignoreTitleExp,omitempty"`
}

type ReviewAppSpecInfra struct {

	// TODO
	Repository string `json:"repository,omitempty"`

	// GitSecretRef is specifying secret for accessing Git remote-repo
	GitSecretRef *corev1.SecretKeySelector `json:"gitSecretRef,omitempty"`

	Manifests ReviewAppSpecInfraManifests `json:"manifests,omitempty"`

	ArgoCDApp ReviewAppSpecInfraArgoCDApp `json:"argocdApp,omitempty"`
}

type ReviewAppSpecInfraManifests struct {
	// Templates is specifying list of ManifestTemplate resources
	Templates []string `json:"templatesName,omitempty"`

	// Dirpath is directory path of deploying TemplateManifests
	// Allow Go-Template notation
	Dirpath string `json:"dirpath,omitempty"`
}

type ReviewAppSpecInfraArgoCDApp struct {

	// Template is specifying ApplicationTemplate resources
	Template string `json:"templateName,omitempty"`

	// Filepath is file path of deploying ApplicationTemplate
	// Allow Go-Template notation
	Filepath string `json:"filepath,omitempty"`
}

// ReviewAppStatus defines the observed state of ReviewApp
type ReviewAppStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ReviewApp is the Schema for the reviewapps API
type ReviewApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReviewAppSpec   `json:"spec,omitempty"`
	Status ReviewAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReviewAppList contains a list of ReviewApp
type ReviewAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReviewApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReviewApp{}, &ReviewAppList{})
}
