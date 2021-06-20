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

// ReviewAppInstanceSpec defines the desired state of ReviewAppInstance
type ReviewAppInstanceSpec struct {

	// App is config of application repository
	App ReviewAppSpecApp `json:"app,omitempty"`

	// Infra is config of manifest repository
	Infra ReviewAppSpecInfra `json:"infra,omitempty"`

	// Application is manifest of ArgoCD Application resource
	Application string `json:"application,omitempty"`

	// Manifests
	Manifests map[string]string `json:"manifests,omitempty"`
}

// ReviewAppInstanceStatus defines the observed state of ReviewAppInstance
type ReviewAppInstanceStatus struct {
	SyncedArtifacts []SyncedArtifact `json:"syncedArtifacts,omitempty"`
}

type SyncedArtifact struct {

	// TODO
	ApplicationName string `json:"applicationName,omitempty"`

	// TODO
	AppRepoPrNum uint `json:"appRepoPrNum,omitempty"`

	// TODO
	AppRepoLatestCommitSha string `json:"appRepoLatestCommitSha,omitempty"`

	// TODO
	InfraRepoLatestCommitSha string `json:"infraRepoLatestCommitSha,omitempty"`

	// TODO
	Notified bool `json:"notifid,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=rai
//+kubebuilder:subresource:status

// ReviewAppInstance is the Schema for the reviewappinstances API
type ReviewAppInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReviewAppInstanceSpec   `json:"spec,omitempty"`
	Status ReviewAppInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReviewAppInstanceList contains a list of ReviewAppInstance
type ReviewAppInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReviewAppInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReviewAppInstance{}, &ReviewAppInstanceList{})
}
