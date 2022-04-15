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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// JobTemplateSpec defines the desired state of JobTemplate
type JobTemplateSpec struct {

	// +kubebuilder:validation:Required
	// CandidateTemplate is included Job manifest. (apiVersion, kind, metadata, spec, ...)
	CandidateTemplate batchv1.Job `json:"candidateTemplate,omitempty"`

	// +kubebuilder:validation:Required
	// StableTemplate is included Job manifest. (apiVersion, kind, metadata, spec, ...)
	StableTemplate batchv1.Job `json:"stableTemplate,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=jt

// JobTemplate is the Schema for the jobtemplates API
type JobTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec JobTemplateSpec `json:"spec,omitempty"`
}

func (JobTemplate) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "JobTemplate",
	}
}

//+kubebuilder:object:root=true

// JobTemplateList contains a list of JobTemplate
type JobTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JobTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JobTemplate{}, &JobTemplateList{})
}
