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
	argocd_application_v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationTemplateSpec defines the desired state of ApplicationTemplate
type ApplicationTemplateSpec struct {

	// Template is included ArgoCD Application manifest. (apiVersion, kind, metadata, spec, ...)
	Template argocd_application_v1alpha1.Application `json:"template,omitempty"`
}

// ApplicationTemplateStatus defines the observed state of ApplicationTemplate
type ApplicationTemplateStatus struct {
	// TODO
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ApplicationTemplate is the Schema for the applicationtemplates API
type ApplicationTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationTemplateSpec   `json:"spec,omitempty"`
	Status ApplicationTemplateStatus `json:"status,omitempty"`
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
