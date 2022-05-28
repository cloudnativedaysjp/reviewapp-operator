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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ReviewAppManagerSpec defines the desired state of ReviewAppManager
type ReviewAppManagerSpec struct {
	ReviewAppCommonSpec `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=ram
//+kubebuilder:printcolumn:name="app_organization",type="string",JSONPath=".spec.appRepoTarget.organization",description="Name of Application Repository's Organization"
//+kubebuilder:printcolumn:name="app_repository",type="string",JSONPath=".spec.appRepoTarget.repository",description="Name of Application Repository"
//+kubebuilder:printcolumn:name="infra_organization",type="string",JSONPath=".spec.infraRepoTarget.organization",description="Name of Infra Repository's Organization"
//+kubebuilder:printcolumn:name="infra_repository",type="string",JSONPath=".spec.infraRepoTarget.repository",description="Name of Infra Repository"

// ReviewAppManager is the Schema for the reviewappmanagers API
type ReviewAppManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ReviewAppManagerSpec `json:"spec,omitempty"`
}

func (ReviewAppManager) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   GroupVersion.Group,
		Version: GroupVersion.Version,
		Kind:    "ReviewAppManager",
	}
}

// TODO: upper limit of Kubernetes object name is 63, but this logic can more 64 charactors.
func (m ReviewAppManager) ReviewAppName(pr PullRequestSpec) string {
	toObjName := func(base string) string {
		return strings.ToLower(
			strings.ReplaceAll(
				strings.ReplaceAll(base,
					"_", "-"),
				".", "-"),
		)
	}
	return fmt.Sprintf("%s-%s-%s-%d",
		m.Name,
		toObjName(pr.AppTarget.Organization),
		toObjName(pr.AppTarget.Repository),
		pr.Number,
	)
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
