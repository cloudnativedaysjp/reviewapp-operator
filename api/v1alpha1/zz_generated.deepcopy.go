//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationTemplate) DeepCopyInto(out *ApplicationTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationTemplate.
func (in *ApplicationTemplate) DeepCopy() *ApplicationTemplate {
	if in == nil {
		return nil
	}
	out := new(ApplicationTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationTemplateList) DeepCopyInto(out *ApplicationTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ApplicationTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationTemplateList.
func (in *ApplicationTemplateList) DeepCopy() *ApplicationTemplateList {
	if in == nil {
		return nil
	}
	out := new(ApplicationTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationTemplateSpec) DeepCopyInto(out *ApplicationTemplateSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationTemplateSpec.
func (in *ApplicationTemplateSpec) DeepCopy() *ApplicationTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestsCache) DeepCopyInto(out *ManifestsCache) {
	*out = *in
	if in.Manifests != nil {
		in, out := &in.Manifests, &out.Manifests
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestsCache.
func (in *ManifestsCache) DeepCopy() *ManifestsCache {
	if in == nil {
		return nil
	}
	out := new(ManifestsCache)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestsTemplate) DeepCopyInto(out *ManifestsTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestsTemplate.
func (in *ManifestsTemplate) DeepCopy() *ManifestsTemplate {
	if in == nil {
		return nil
	}
	out := new(ManifestsTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ManifestsTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestsTemplateList) DeepCopyInto(out *ManifestsTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ManifestsTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestsTemplateList.
func (in *ManifestsTemplateList) DeepCopy() *ManifestsTemplateList {
	if in == nil {
		return nil
	}
	out := new(ManifestsTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ManifestsTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ManifestsTemplateSpec) DeepCopyInto(out *ManifestsTemplateSpec) {
	*out = *in
	if in.CandidateData != nil {
		in, out := &in.CandidateData, &out.CandidateData
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.StableData != nil {
		in, out := &in.StableData, &out.StableData
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ManifestsTemplateSpec.
func (in *ManifestsTemplateSpec) DeepCopy() *ManifestsTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(ManifestsTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewApp) DeepCopyInto(out *ReviewApp) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	in.Tmp.DeepCopyInto(&out.Tmp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewApp.
func (in *ReviewApp) DeepCopy() *ReviewApp {
	if in == nil {
		return nil
	}
	out := new(ReviewApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReviewApp) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppList) DeepCopyInto(out *ReviewAppList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ReviewApp, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppList.
func (in *ReviewAppList) DeepCopy() *ReviewAppList {
	if in == nil {
		return nil
	}
	out := new(ReviewAppList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReviewAppList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManager) DeepCopyInto(out *ReviewAppManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManager.
func (in *ReviewAppManager) DeepCopy() *ReviewAppManager {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReviewAppManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerList) DeepCopyInto(out *ReviewAppManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ReviewAppManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerList.
func (in *ReviewAppManagerList) DeepCopy() *ReviewAppManagerList {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ReviewAppManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpec) DeepCopyInto(out *ReviewAppManagerSpec) {
	*out = *in
	in.AppTarget.DeepCopyInto(&out.AppTarget)
	out.AppConfig = in.AppConfig
	in.InfraTarget.DeepCopyInto(&out.InfraTarget)
	in.InfraConfig.DeepCopyInto(&out.InfraConfig)
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpec.
func (in *ReviewAppManagerSpec) DeepCopy() *ReviewAppManagerSpec {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecAppConfig) DeepCopyInto(out *ReviewAppManagerSpecAppConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecAppConfig.
func (in *ReviewAppManagerSpecAppConfig) DeepCopy() *ReviewAppManagerSpecAppConfig {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecAppConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecAppTarget) DeepCopyInto(out *ReviewAppManagerSpecAppTarget) {
	*out = *in
	if in.GitSecretRef != nil {
		in, out := &in.GitSecretRef, &out.GitSecretRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.IgnoreLabels != nil {
		in, out := &in.IgnoreLabels, &out.IgnoreLabels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecAppTarget.
func (in *ReviewAppManagerSpecAppTarget) DeepCopy() *ReviewAppManagerSpecAppTarget {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecAppTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecInfraArgoCDApp) DeepCopyInto(out *ReviewAppManagerSpecInfraArgoCDApp) {
	*out = *in
	out.Template = in.Template
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecInfraArgoCDApp.
func (in *ReviewAppManagerSpecInfraArgoCDApp) DeepCopy() *ReviewAppManagerSpecInfraArgoCDApp {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecInfraArgoCDApp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecInfraConfig) DeepCopyInto(out *ReviewAppManagerSpecInfraConfig) {
	*out = *in
	in.Manifests.DeepCopyInto(&out.Manifests)
	out.ArgoCDApp = in.ArgoCDApp
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecInfraConfig.
func (in *ReviewAppManagerSpecInfraConfig) DeepCopy() *ReviewAppManagerSpecInfraConfig {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecInfraConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecInfraManifests) DeepCopyInto(out *ReviewAppManagerSpecInfraManifests) {
	*out = *in
	if in.Templates != nil {
		in, out := &in.Templates, &out.Templates
		*out = make([]NamespacedName, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecInfraManifests.
func (in *ReviewAppManagerSpecInfraManifests) DeepCopy() *ReviewAppManagerSpecInfraManifests {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecInfraManifests)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerSpecInfraTarget) DeepCopyInto(out *ReviewAppManagerSpecInfraTarget) {
	*out = *in
	if in.GitSecretRef != nil {
		in, out := &in.GitSecretRef, &out.GitSecretRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerSpecInfraTarget.
func (in *ReviewAppManagerSpecInfraTarget) DeepCopy() *ReviewAppManagerSpecInfraTarget {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerSpecInfraTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerStatus) DeepCopyInto(out *ReviewAppManagerStatus) {
	*out = *in
	if in.SyncedPullRequests != nil {
		in, out := &in.SyncedPullRequests, &out.SyncedPullRequests
		*out = make([]ReviewAppManagerStatusSyncedPullRequests, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerStatus.
func (in *ReviewAppManagerStatus) DeepCopy() *ReviewAppManagerStatus {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppManagerStatusSyncedPullRequests) DeepCopyInto(out *ReviewAppManagerStatusSyncedPullRequests) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppManagerStatusSyncedPullRequests.
func (in *ReviewAppManagerStatusSyncedPullRequests) DeepCopy() *ReviewAppManagerStatusSyncedPullRequests {
	if in == nil {
		return nil
	}
	out := new(ReviewAppManagerStatusSyncedPullRequests)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppSpec) DeepCopyInto(out *ReviewAppSpec) {
	*out = *in
	in.AppTarget.DeepCopyInto(&out.AppTarget)
	out.AppConfig = in.AppConfig
	in.InfraTarget.DeepCopyInto(&out.InfraTarget)
	in.InfraConfig.DeepCopyInto(&out.InfraConfig)
	if in.Variables != nil {
		in, out := &in.Variables, &out.Variables
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppSpec.
func (in *ReviewAppSpec) DeepCopy() *ReviewAppSpec {
	if in == nil {
		return nil
	}
	out := new(ReviewAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppStatus) DeepCopyInto(out *ReviewAppStatus) {
	*out = *in
	out.Sync = in.Sync
	in.ManifestsCache.DeepCopyInto(&out.ManifestsCache)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppStatus.
func (in *ReviewAppStatus) DeepCopy() *ReviewAppStatus {
	if in == nil {
		return nil
	}
	out := new(ReviewAppStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppTmp) DeepCopyInto(out *ReviewAppTmp) {
	*out = *in
	in.PullRequest.DeepCopyInto(&out.PullRequest)
	if in.Manifests != nil {
		in, out := &in.Manifests, &out.Manifests
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppTmp.
func (in *ReviewAppTmp) DeepCopy() *ReviewAppTmp {
	if in == nil {
		return nil
	}
	out := new(ReviewAppTmp)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReviewAppTmpPr) DeepCopyInto(out *ReviewAppTmpPr) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReviewAppTmpPr.
func (in *ReviewAppTmpPr) DeepCopy() *ReviewAppTmpPr {
	if in == nil {
		return nil
	}
	out := new(ReviewAppTmpPr)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SyncStatus) DeepCopyInto(out *SyncStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SyncStatus.
func (in *SyncStatus) DeepCopy() *SyncStatus {
	if in == nil {
		return nil
	}
	out := new(SyncStatus)
	in.DeepCopyInto(out)
	return out
}
