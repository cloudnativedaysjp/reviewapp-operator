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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ManifestsTemplateSpec struct {
	// CandidateData is field that be given various resources' manifest.
	CandidateData Manifests `json:"candidate,omitempty"`

	// StableData is field that be given various resources' manifest.
	StableData Manifests `json:"stable,omitempty"`
}

type Manifests map[string]string
type ManifestsBase64 map[string]string

func (m Manifests) ToBase64() ManifestsBase64 {
	manifestsBase64 := make(ManifestsBase64)
	for k, v := range m {
		manifestsBase64[k] = base64.StdEncoding.EncodeToString([]byte(v))
	}
	return manifestsBase64
}

func (m ManifestsBase64) Decode() (Manifests, error) {
	manifests := make(Manifests)
	for k, v := range m {
		m, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		manifests[k] = string(m)
	}
	return manifests, nil
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=mt

// ManifestsTemplate is the Schema for the manifeststemplates API
type ManifestsTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ManifestsTemplateSpec `json:"spec"`
}

func (ManifestsTemplate) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "ManifestsTemplate",
	}
}

func (m ManifestsTemplate) AppendOrUpdate(mt ManifestsTemplate) ManifestsTemplate {
	if m.Spec.StableData == nil {
		m.Spec.StableData = make(map[string]string)
	}
	for k, v := range mt.Spec.StableData {
		m.Spec.StableData[k] = v
	}
	if m.Spec.CandidateData == nil {
		m.Spec.CandidateData = make(map[string]string)
	}
	for k, v := range mt.Spec.CandidateData {
		m.Spec.CandidateData[k] = v
	}
	return m
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
