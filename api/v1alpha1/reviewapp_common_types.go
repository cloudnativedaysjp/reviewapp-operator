package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type ReviewAppCommonSpec struct {
	// TODO
	AppTarget ReviewAppCommonSpecAppTarget `json:"appRepoTarget"`

	// TODO
	AppConfig ReviewAppCommonSpecAppConfig `json:"appRepoConfig"`

	// TODO
	InfraTarget ReviewAppCommonSpecInfraTarget `json:"infraRepoTarget"`

	// TODO
	InfraConfig ReviewAppCommonSpecInfraConfig `json:"infraRepoConfig"`

	// PreStopJob is specified JobTemplate that executed at previous of stopped ReviewApp
	PreStopJob NamespacedName `json:"preStopJob,omitempty"`

	// Variables is available to use input of Application & Manifest Template
	Variables []string `json:"variables,omitempty"`
}

func (m ReviewAppCommonSpec) HavePreStopJob() bool {
	return m.PreStopJob.Namespace != "" && m.PreStopJob.Name != ""
}

type ReviewAppCommonSpecAppTarget struct {

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
}

func (m ReviewAppCommonSpecAppTarget) GitSecretSelector() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("gitSecretRef is not set")
	}
	return m.GitSecretRef, nil
}

// TODO: upper limit of Kubernetes object name is 63, but this logic can more 64 charactors.
func (m ReviewAppCommonSpecAppTarget) PullRequestName(number int) string {
	toObjName := func(base string) string {
		return strings.ToLower(
			strings.ReplaceAll(
				strings.ReplaceAll(base,
					"_", "-"),
				".", "-"),
		)
	}
	return fmt.Sprintf("%s-%s-%d",
		toObjName(m.Organization),
		toObjName(m.Repository),
		number,
	)
}

type ReviewAppCommonSpecAppConfig struct {

	// Message is output to specified App Repository's PR when reviewapp is synced
	// +optional
	Message string `json:"message,omitempty"`

	// SendMessageEveryTime is flag. Controller send comment to App Repository's PR only first time if flag is false.
	// +kubebuilder:default=false
	// +optional
	SendMessageEveryTime bool `json:"sendMessageEveryTime,omitempty"`
}

type ReviewAppCommonSpecInfraTarget struct {

	// TODO
	Organization string `json:"organization"`

	// TODO
	Repository string `json:"repository"`

	// TODO
	Username string `json:"username"`

	// TODO
	Branch string `json:"branch"`

	// GitSecretRef is specifying secret for accessing Git remote-repo
	GitSecretRef *corev1.SecretKeySelector `json:"gitSecretRef,omitempty"`
}

func (m ReviewAppCommonSpecInfraTarget) GitSecretSelector() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("gitSecretRef is not set")
	}
	return m.GitSecretRef, nil
}

type ReviewAppCommonSpecInfraConfig struct {

	// TODO
	Manifests ReviewAppCommonSpecInfraManifests `json:"manifests,omitempty"`

	// TODO
	ArgoCDApp ReviewAppCommonSpecInfraArgoCDApp `json:"argocdApp,omitempty"`
}

type ReviewAppCommonSpecInfraManifests struct {
	// Templates is specifying list of ManifestTemplate resources
	Templates []NamespacedName `json:"templates,omitempty"`

	// Dirpath is directory path of deploying TemplateManifests
	// Allow Go-Template notation
	Dirpath string `json:"dirpath,omitempty"`
}

type ReviewAppCommonSpecInfraArgoCDApp struct {

	// Template is specifying ApplicationTemplate resources
	Template NamespacedName `json:"template,omitempty"`

	// Filepath is file path of deploying ApplicationTemplate
	// Allow Go-Template notation
	Filepath string `json:"filepath,omitempty"`
}
