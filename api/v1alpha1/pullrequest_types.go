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
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PullRequestSpec defines the desired state of PullRequest
type PullRequestSpec struct {

	// TODO
	AppTarget ReviewAppCommonSpecAppTarget `json:"appRepoTarget"`

	// TODO
	Number int `json:"number,omitempty"`
}

// PullRequestStatus defines the observed state of PullRequest
type PullRequestStatus struct {

	// TODO
	HeadBranch string `json:"headBranch,omitempty"`

	// TODO
	BaseBranch string `json:"baseBranch,omitempty"`

	// TODO
	LatestCommitHash string `json:"latestCommitHash,omitempty"`

	// TODO
	Title string `json:"title,omitempty"`

	// TODO
	Labels []string `json:"labels,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=pr
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="organization",type="string",JSONPath=".spec.appRepoTarget.organization",description="Name of Application Repository's Organization"
//+kubebuilder:printcolumn:name="repository",type="string",JSONPath=".spec.appRepoTarget.repository",description="Name of Application Repository"
//+kubebuilder:printcolumn:name="number",type="integer",JSONPath=".spec.number",description="Number of Application Repository's PullRequest"

// PullRequest is the Schema for the pullrequests API
type PullRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PullRequestSpec   `json:"spec,omitempty"`
	Status PullRequestStatus `json:"status,omitempty"`
}

func (PullRequest) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "PullRequest",
	}
}

const (
	candidateLabelName = "candidate-template"
)

func (m PullRequest) IsCandidate() bool {
	isCandidate := false
	for _, l := range m.Labels {
		if l == candidateLabelName {
			isCandidate = true
		}
	}
	return isCandidate
}

func (m PullRequest) MustBeIgnored(appTarget ReviewAppCommonSpecAppTarget) bool {
	for _, actualLabel := range m.Status.Labels {
		for _, ignoreLabel := range appTarget.IgnoreLabels {
			if actualLabel == ignoreLabel {
				return true
			}
		}
	}
	if appTarget.IgnoreTitleExp != "" {
		r := regexp.MustCompile(appTarget.IgnoreTitleExp)
		if r.Match([]byte(m.Status.Title)) {
			return true
		}
	}
	return false
}

//+kubebuilder:object:root=true

// PullRequestList contains a list of PullRequest
type PullRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PullRequest `json:"items"`
}

func (m PullRequestList) ExcludeSpecificPR(appTarget ReviewAppCommonSpecAppTarget) PullRequestList {
	remove := func(m PullRequestList, idx int) PullRequestList {
		items := m.Items
		if idx == len(items)-1 {
			m.Items = items[:idx]
			return m
		} else {
			m.Items = append(items[:idx], items[idx+1:]...)
			return m
		}
	}
	ignoreLabels := appTarget.IgnoreLabels
	removedCount := 0
	for idx, pr := range m.Items {
		for _, actualLabel := range pr.Status.Labels {
			for _, ignoreLabel := range ignoreLabels {
				if actualLabel == ignoreLabel {
					m = remove(m, idx-removedCount)
					removedCount += 1
				}
			}
		}
	}
	ignoreTitleExp := appTarget.IgnoreTitleExp
	removedCount = 0
	if ignoreTitleExp != "" {
		r := regexp.MustCompile(ignoreTitleExp)
		for idx, pr := range m.Items {
			if r.Match([]byte(pr.Status.Title)) {
				m = remove(m, idx-removedCount)
				removedCount += 1
			}
		}
	}
	return m
}

func init() {
	SchemeBuilder.Register(&PullRequest{}, &PullRequestList{})
}
