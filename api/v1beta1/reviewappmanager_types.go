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

// ReviewAppManagerSpec defines the desired state of ReviewAppManager
type ReviewAppManagerSpec struct {

	// App is config of application repository
	App ReviewAppManagerSpecApp `json:"appRepo"`

	// Infra is config of manifest repository
	Infra ReviewAppManagerSpecInfra `json:"infraRepo"`

	// Variables is available to use input of Application & Manifest Template
	Variables []string `json:"variables,omitempty"`
}

type ReviewAppManagerSpecApp struct {

	// TODO
	Organization string `json:"organization"`

	// TODO
	Repository string `json:"repository"`

	// TODO
	Username string `json:"username"`

	// GitSecretRef is specifying secret for accessing Git remote-repo
	GitSecretRef *corev1.SecretKeySelector `json:"gitSecretRef,omitempty"`

	// IgnoreLabels is TODO
	IgnoreLabels []string `json:"ignoreLabels,omitempty"`

	// IgnoreTitleExp is TODO
	IgnoreTitleExp string `json:"ignoreTitleExp,omitempty"`

	// Message is output to specified App Repositories' PR when reviewapp is synced
	Message string `json:"message,omitempty"`
}

type ReviewAppManagerSpecInfra struct {

	// TODO
	Organization string `json:"organization"`

	// TODO
	Repository string `json:"repository"`

	// TODO
	Username string `json:"username"`

	// TODO
	TargetBranch string `json:"targetBranch"`

	// GitSecretRef is specifying secret for accessing Git remote-repo
	GitSecretRef *corev1.SecretKeySelector `json:"gitSecretRef,omitempty"`

	Manifests ReviewAppManagerSpecInfraManifests `json:"manifests,omitempty"`

	ArgoCDApp ReviewAppManagerSpecInfraArgoCDApp `json:"argocdApp,omitempty"`
}

type ReviewAppManagerSpecInfraManifests struct {
	// Templates is specifying list of ManifestTemplate resources
	Templates []NamespacedName `json:"templates,omitempty"`

	// Dirpath is directory path of deploying TemplateManifests
	// Allow Go-Template notation
	Dirpath string `json:"dirpath,omitempty"`
}

type ReviewAppManagerSpecInfraArgoCDApp struct {

	// Template is specifying ApplicationTemplate resources
	Template NamespacedName `json:"template,omitempty"`

	// Filepath is file path of deploying ApplicationTemplate
	// Allow Go-Template notation
	Filepath string `json:"filepath,omitempty"`
}

// ReviewAppManagerStatus defines the observed state of ReviewAppManager
type ReviewAppManagerStatus struct {

	// TODO
	SyncedPullRequests []ReviewAppManagerStatusSyncedPullRequests `json:"syncedPullRequests,omitempty"`
}

type ReviewAppManagerStatusSyncedPullRequests struct {

	// TODO
	Organization string `json:"organization,omitempty"`

	// TODO
	Repository string `json:"repository,omitempty"`

	// TODO
	Number int `json:"number,omitempty"`

	// TODO
	ReviewAppName string `json:"reviewAppName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ram
//+kubebuilder:subresource:status

// ReviewAppManager is the Schema for the reviewappmanagers API
type ReviewAppManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReviewAppManagerSpec   `json:"spec,omitempty"`
	Status ReviewAppManagerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ReviewAppManagerList contains a list of ReviewAppManager
type ReviewAppManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReviewAppManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReviewAppManager{}, &ReviewAppManagerList{})
}
