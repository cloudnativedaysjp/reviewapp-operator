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
	"reflect"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ReviewAppSpec defines the desired state of ReviewApp
type ReviewAppSpec struct {
	ReviewAppCommonSpec `json:",inline"`

	// PullRequestName is object name of PullRequest resource
	PullRequest NamespacedName `json:"pullRequest"`
}

// ReviewAppStatus defines the observed state of ReviewApp
type ReviewAppStatus struct {

	// TODO
	Sync ReviewAppStatusSync `json:"sync,omitempty"`

	// TODO
	PullRequestCache ReviewAppStatusPullRequestCache `json:"pullRequestCache,omitempty"`

	// ManifestsCache is used in "confirm Templates Are Updated" for confirm templates updated
	ManifestsCache ManifestsCache `json:"manifestsCache,omitempty"`
}

func (m ReviewAppStatus) HaveManifestsTemplateBeenUpdated(manifests Manifests) bool {
	return !reflect.DeepEqual(manifests.ToBase64(), m.ManifestsCache.ManifestsBase64)
}

func (m ReviewAppStatus) HasApplicationTemplateBeenUpdated(application Application) bool {
	return application.ToBase64() != m.ManifestsCache.ApplicationBase64
}

func (m ReviewAppStatus) HasPullRequestBeenUpdated(hash string) bool {
	return m.PullRequestCache.LatestCommitHash != hash
}

func (m ReviewAppStatus) HasArgoCDApplicationBeenUpdated(hash string) bool {
	return m.PullRequestCache.LatestCommitHash == hash
}

type ReviewAppStatusSync struct {

	// Status is the sync state of the comparison
	Status SyncStatusCode `json:"status,omitempty"`

	// AlreadySentMessage is used to decide sending message to AppRepo's PR when Spec.AppConfig.SendMessageOnlyFirstTime is true.
	AlreadySentMessage bool `json:"alreadySentMessage,omitempty"`
}

type ReviewAppStatusPullRequestCache struct {

	// TODO
	Number int `json:"number,omitempty"`

	// TODO
	BaseBranch string `json:"baseBranch,omitempty"`

	// TODO
	HeadBranch string `json:"headBranch,omitempty"`

	// TODO
	LatestCommitHash string `json:"latestCommitHash,omitempty"`

	// TODO
	Title string `json:"title,omitempty"`

	// TODO
	Labels []string `json:"labels,omitempty"`

	// TODO
	SyncedTimestamp metav1.Time `json:"syncedTimestamp,omitempty"`
}

func (m *ReviewAppStatusPullRequestCache) IsEmpty() bool {
	return m.Number == 0 || m.BaseBranch == "" ||
		m.HeadBranch == "" || m.LatestCommitHash == ""
}

func (m *ReviewAppStatusPullRequestCache) UpdateCache(pr PullRequest) {
	updated := false
	copyString := func(dst, src *string) {
		if *dst != *src {
			updated = true
			*dst = *src
		}
	}
	copySlice := func(dst, src *[]string) {
		if len(*dst) == len(*src) {
			var dstSorted, srcSorted []string
			copy(dstSorted, *dst)
			copy(srcSorted, *src)
			sort.SliceStable(dstSorted, func(i, j int) bool { return dstSorted[i] < dstSorted[j] })
			sort.SliceStable(srcSorted, func(i, j int) bool { return srcSorted[i] < srcSorted[j] })
			if !reflect.DeepEqual(dstSorted, srcSorted) {
				updated = true
				copy(*dst, *src)
			}
		}
	}
	copyString(&m.HeadBranch, &pr.Status.HeadBranch)
	copyString(&m.LatestCommitHash, &pr.Status.LatestCommitHash)
	copyString(&m.Title, &pr.Status.Title)
	copySlice(&m.Labels, &pr.Status.Labels)
	if updated {
		m.SyncedTimestamp = metav1.Now()
	}
}

type ManifestsCache struct {

	// TODO
	ApplicationName string `json:"applicationName,omitempty"`

	// TODO
	ApplicationNamespace string `json:"applicationNamespace,omitempty"`

	// Application is manifest of ArgoCD Application resource
	ApplicationBase64 ApplicationBase64 `json:"application,omitempty"`

	// Manifests is other manifests
	ManifestsBase64 ManifestsBase64 `json:"manifests,omitempty"`
}

// SyncStatusCode is a type which represents possible comparison results
type SyncStatusCode string

// Possible comparison results
const (
	// SyncStatusCodeUnknown indicates that the status of a sync could not be reliably determined
	SyncStatusCodeUnknown SyncStatusCode = "Unknown"
	// SyncStatusCodeInitialize indicates that ReviewApp Object is created now.
	SyncStatusCodeInitialize SyncStatusCode = "Initialize"
	// SyncStatusCodeWatchingAppRepo indicates that ReviewApp Object is no changing.
	SyncStatusCodeWatchingAppRepoAndTemplates SyncStatusCode = "WatchingAppRepoAndTemplates"
	// SyncStatusCodeNeedToUpdateInfraRepo indicates that ReviewApp Object was updated. Operator will update manifests to infra repo.
	SyncStatusCodeNeedToUpdateInfraRepo SyncStatusCode = "NeedToUpdateInfraRepo"
	// SyncStatusCodeUpdatedInfraRepo indicates that ReviewApp manifests was deployed to infra repo. Operator is waiting ArgoCD Application updated
	SyncStatusCodeUpdatedInfraRepo SyncStatusCode = "UpdatedInfraRepo"
)

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ra
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="app_organization",type="string",JSONPath=".spec.appRepoTarget.organization",description="Name of Application Repository's Organization"
//+kubebuilder:printcolumn:name="app_repository",type="string",JSONPath=".spec.appRepoTarget.repository",description="Name of Application Repository"
//+kubebuilder:printcolumn:name="app_pr_num",type="integer",JSONPath=".status.pullRequestCache.number",description="Number of Application Repository's PullRequest"
//+kubebuilder:printcolumn:name="infra_organization",type="string",JSONPath=".spec.infraRepoTarget.organization",description="Name of Infra Repository's Organization"
//+kubebuilder:printcolumn:name="infra_repository",type="string",JSONPath=".spec.infraRepoTarget.repository",description="Name of Infra Repository"

// ReviewApp is the Schema for the reviewapp API
type ReviewApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReviewAppSpec   `json:"spec,omitempty"`
	Status ReviewAppStatus `json:"status,omitempty"`
}

func (ReviewApp) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "ReviewApp",
	}
}

func (m ReviewApp) HasMessageAlreadyBeenSent() bool {
	return m.Spec.AppConfig.Message == "" ||
		(!m.Spec.AppConfig.SendMessageEveryTime && m.Status.Sync.AlreadySentMessage)
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
