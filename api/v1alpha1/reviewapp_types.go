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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReviewAppSpec defines the desired state of ReviewApp
type ReviewAppSpec struct {

	// TODO
	AppTarget ReviewAppManagerSpecAppTarget `json:"appRepoTarget"`

	// TODO
	AppConfig ReviewAppManagerSpecAppConfig `json:"appRepoConfig"`

	// TODO
	InfraTarget ReviewAppManagerSpecInfraTarget `json:"infraRepoTarget"`

	// TODO
	InfraConfig ReviewAppManagerSpecInfraConfig `json:"infraRepoConfig"`

	// PreStopJob is specified JobTemplate that executed at previous of stopped ReviewApp
	PreStopJob NamespacedName `json:"preStopJob,omitempty"`

	// Variables is available to use input of Application & Manifest Template
	Variables []string `json:"variables,omitempty"`

	// AppPrNum is watched PR's number by this RA
	AppPrNum int `json:"appRepoPrNum"`
}

// ReviewAppStatus defines the observed state of ReviewApp
type ReviewAppStatus struct {
	// TODO
	Sync SyncStatus `json:"sync,omitempty"`

	// ManifestsCache is used in "confirm Templates Are Updated" for confirm templates updated
	ManifestsCache ManifestsCache `json:"manifestsCache,omitempty"`

	// AlreadySentMessage is used to decide sending message to AppRepo's PR when Spec.AppConfig.SendMessageOnlyFirstTime is true.
	AlreadySentMessage bool `json:"alreadySentMessage,omitempty"`
}

type SyncStatus struct {

	// Status is the sync state of the comparison
	Status SyncStatusCode `json:"status"`

	// TODO
	ApplicationName string `json:"applicationName,omitempty"`

	// TODO
	ApplicationNamespace string `json:"applicationNamespace,omitempty"`

	// TODO
	AppRepoBranch string `json:"appRepoBranch,omitempty"`

	// TODO
	AppRepoLatestCommitSha string `json:"appRepoLatestCommitSha,omitempty"`

	// TODO
	InfraRepoLatestCommitSha string `json:"infraRepoLatestCommitSha,omitempty"`
}

type ManifestsCache struct {

	// Application is manifest of ArgoCD Application resource
	Application string `json:"application,omitempty"`

	// Manifests is other manifests
	Manifests map[string]string `json:"manifests,omitempty"`
}

// SyncStatusCode is a type which represents possible comparison results
type SyncStatusCode string

// Possible comparison results
const (
	// SyncStatusCodeUnknown indicates that the status of a sync could not be reliably determined
	SyncStatusCodeUnknown SyncStatusCode = "Unknown"
	// SyncStatusCodeWatchingAppRepo indicates that TODO
	SyncStatusCodeWatchingAppRepo SyncStatusCode = "WatchingAppRepo"
	// SyncStatusCodeWatchingTemplates indicates that TODO
	SyncStatusCodeWatchingTemplates SyncStatusCode = "WatchingTemplates"
	// SyncStatusCodeNeedToUpdateInfraRepo indicates that watched updated app repo & will update manifests to infra repo
	SyncStatusCodeNeedToUpdateInfraRepo SyncStatusCode = "NeedToUpdateInfraRepo"
	// SyncStatusCodeUpdatedInfraRepo indicates that watched updated manifest repo & wait ArgoCD Application updated
	SyncStatusCodeUpdatedInfraRepo SyncStatusCode = "UpdatedInfraRepo"
)

type ReviewAppTmp struct {
	PullRequest                ReviewAppTmpPr
	Application                string
	ApplicationWithAnnotations string
	Manifests                  map[string]string
}

type ReviewAppTmpPr struct {
	Organization  string
	Repository    string
	Branch        string
	Number        int
	HeadCommitSha string
	Title         string
	Labels        []string
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ra
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="app_organization",type="string",JSONPath=".spec.appRepoTarget.organization",description="Name of Application Repository's Organization"
//+kubebuilder:printcolumn:name="app_repository",type="string",JSONPath=".spec.appRepoTarget.repository",description="Name of Application Repository"
//+kubebuilder:printcolumn:name="app_pr_num",type="integer",JSONPath=".spec.appRepoPrNum",description="Number of Application Repository's PullRequest"
//+kubebuilder:printcolumn:name="infra_organization",type="string",JSONPath=".spec.infraRepoTarget.organization",description="Name of Infra Repository's Organization"
//+kubebuilder:printcolumn:name="infra_repository",type="string",JSONPath=".spec.infraRepoTarget.repository",description="Name of Infra Repository"

// ReviewApp is the Schema for the reviewapp API
type ReviewApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReviewAppSpec   `json:"spec,omitempty"`
	Status ReviewAppStatus `json:"status,omitempty"`

	Tmp ReviewAppTmp `json:"-"`
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
