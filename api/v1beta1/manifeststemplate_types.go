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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManifestsTemplateStatus defines the observed state of ManifestsTemplate
type ManifestsTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=mt
//+kubebuilder:subresource:status

// ManifestsTemplate is the Schema for the manifeststemplates API
type ManifestsTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Data is field that be given various resources' manifest.
	Data map[string]string `json:"data,omitempty"`

	Status ManifestsTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManifestsTemplateList contains a list of ManifestsTemplate
type ManifestsTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManifestsTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManifestsTemplate{}, &ManifestsTemplateList{})
}
