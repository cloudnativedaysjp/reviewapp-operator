package models

import (
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

/* ReviewApp or ReviewAppManager */

type ReviewAppOrReviewAppManager interface {
	NamespaceName() types.NamespacedName
	AppRepoTarget() AppRepoTarget
	AppRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecAppConfig
	InfraRepoTarget() InfraRepoTarget
	InfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig
	Variables() []string
}

/* ReviewApp */

type ReviewApp dreamkastv1alpha1.ReviewApp

func NewReviewApp(ra *dreamkastv1alpha1.ReviewApp) ReviewApp {
	return ReviewApp(*ra)
}

func (m ReviewApp) NamespaceName() types.NamespacedName {
	return types.NamespacedName{Namespace: m.Namespace, Name: m.Name}
}
func (m ReviewApp) AppRepoTarget() AppRepoTarget {
	return AppRepoTarget(m.Spec.AppTarget)
}
func (m ReviewApp) AppRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecAppConfig {
	return m.Spec.AppConfig
}
func (m ReviewApp) InfraRepoTarget() InfraRepoTarget {
	return InfraRepoTarget(m.Spec.InfraTarget)
}
func (m ReviewApp) InfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig {
	return m.Spec.InfraConfig
}
func (m ReviewApp) Variables() []string {
	return m.Spec.Variables
}

func (m ReviewApp) ToReviewAppCR() *dreamkastv1alpha1.ReviewApp {
	ra := dreamkastv1alpha1.ReviewApp(m)
	return &ra
}

func (m ReviewApp) AtFilepath() string {
	return m.Spec.InfraConfig.ArgoCDApp.Filepath
}

func (m ReviewApp) MtDirpath() string {
	return m.Spec.InfraConfig.Manifests.Dirpath
}

func (m ReviewApp) PrNum() int {
	return m.Spec.AppPrNum
}

func (m ReviewApp) HavingPreStopJob() bool {
	return m.Spec.PreStopJob.Namespace != "" && m.Spec.PreStopJob.Name != ""
}

func (m ReviewApp) HasMessageAlreadyBeenSent() bool {
	status := m.GetStatus()
	return m.Spec.AppConfig.Message == "" || (!m.Spec.AppConfig.SendMessageEveryTime && status.AlreadySentMessage)
}

func (m ReviewApp) GetStatus() ReviewAppStatus {
	return ReviewAppStatus(m.Status)
}

/* ReviewAppStatus */

type ReviewAppStatus dreamkastv1alpha1.ReviewAppStatus

func (m ReviewAppStatus) UpdateStatusOfAppRepo(pr PullRequest) (ReviewAppStatus, bool) {
	updated := false
	m.Sync.AppRepoBranch = pr.Branch
	if m.Sync.AppRepoLatestCommitSha != pr.HeadCommitSha {
		updated = true
	}
	m.Sync.AppRepoLatestCommitSha = pr.HeadCommitSha
	return m, updated
}

func (m ReviewAppStatus) UpdateStatusOfApplication(application Application) (ReviewAppStatus, bool, error) {
	updated := false
	argocdAppNamespacedName, err := application.NamespacedName()
	if err != nil {
		return ReviewAppStatus{}, false, err
	}
	if m.Sync.ApplicationName != argocdAppNamespacedName.Name || m.Sync.ApplicationNamespace != argocdAppNamespacedName.Namespace {
		updated = true
	}
	if !reflect.DeepEqual(string(application), m.ManifestsCache.Application) {
		updated = true
	}
	m.Sync.ApplicationName = argocdAppNamespacedName.Name
	m.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	return m, updated, nil
}

func (m ReviewAppStatus) WasManifestsUpdated(manifests Manifests) bool {
	updated := false
	if !reflect.DeepEqual(map[string]string(manifests), m.ManifestsCache.Manifests) {
		updated = true
	}
	return updated
}

func (m ReviewAppStatus) HasApplicationBeenUpdated(hash string) bool {
	return m.Sync.AppRepoLatestCommitSha == hash
}

/* ReviewAppManager */

type ReviewAppManager dreamkastv1alpha1.ReviewAppManager

func NewReviewAppManager(ram dreamkastv1alpha1.ReviewAppManager) ReviewAppManager {
	return ReviewAppManager(ram)
}

func (m ReviewAppManager) NamespaceName() types.NamespacedName {
	return types.NamespacedName{Namespace: m.Namespace, Name: m.Name}
}
func (m ReviewAppManager) AppRepoTarget() AppRepoTarget {
	return AppRepoTarget(m.Spec.AppTarget)
}
func (m ReviewAppManager) AppRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecAppConfig {
	return m.Spec.AppConfig
}
func (m ReviewAppManager) InfraRepoTarget() InfraRepoTarget {
	return InfraRepoTarget(m.Spec.InfraTarget)
}
func (m ReviewAppManager) InfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig {
	return m.Spec.InfraConfig
}
func (m ReviewAppManager) Variables() []string {
	return m.Spec.Variables
}

func (m ReviewAppManager) ToReviewAppCR() *dreamkastv1alpha1.ReviewAppManager {
	ram := dreamkastv1alpha1.ReviewAppManager(m)
	return &ram
}

func (m ReviewAppManager) GenerateReviewApp(pr PullRequest, v Templator) (ReviewApp, error) {
	ra := ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ReviewAppName(pr),
			Namespace: m.Namespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppSpec{
			AppTarget:   m.Spec.AppTarget,
			InfraTarget: m.Spec.InfraTarget,
			Variables:   m.Spec.Variables,
			PreStopJob:  m.Spec.PreStopJob,
			AppPrNum:    pr.Number,
		},
	}
	{ // template from ram.Spec.AppConfig to ra.Spec.AppConfig
		out, err := yaml.Marshal(&m.Spec.AppConfig)
		if err != nil {
			return ReviewApp{}, err
		}
		appConfigStr, err := v.Templating(string(out))
		if err != nil {
			return ReviewApp{}, err
		}
		if err := yaml.Unmarshal([]byte(appConfigStr), &ra.Spec.AppConfig); err != nil {
			return ReviewApp{}, err
		}
	}
	{ // template from ram.Spec.InfraConfig to ra.Spec.InfraConfig
		out, err := yaml.Marshal(&m.Spec.InfraConfig)
		if err != nil {
			return ReviewApp{}, err
		}
		infraConfigStr, err := v.Templating(string(out))
		if err != nil {
			return ReviewApp{}, err
		}
		if err := yaml.Unmarshal([]byte(infraConfigStr), &ra.Spec.InfraConfig); err != nil {
			return ReviewApp{}, err
		}
	}
	return ra, nil
}

func (m ReviewAppManager) ReviewAppName(pr PullRequest) string {
	return fmt.Sprintf("%s-%s-%s-%d",
		m.Name,
		strings.ToLower(pr.Organization),
		strings.ToLower(pr.Repository),
		pr.Number,
	)
}

func (m ReviewAppManager) ListOutOfSyncReviewAppName(prs []PullRequest) []string {
	var result []string
loop:
	for _, a := range m.Status.SyncedPullRequests {
		for _, b := range prs {
			if a.Organization == b.Organization && a.Repository == b.Repository && a.Number == b.Number {
				continue loop
			}
		}
		result = append(result, m.ReviewAppName(PullRequest{
			Organization: a.Organization,
			Repository:   a.Repository,
			Number:       a.Number,
		}))
	}
	return result
}

/* AppRepoTarget Or InfraRepoTarget */

type AppOrInfraRepoTarget interface {
	GitSecretSelector() (*corev1.SecretKeySelector, error)
}

/* AppRepoTarget  */

type AppRepoTarget dreamkastv1alpha1.ReviewAppManagerSpecAppTarget

func (m AppRepoTarget) GitSecretSelector() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("gitSecretRef is not set")
	}
	return m.GitSecretRef, nil
}

/* InfraRepoTarget  */

type InfraRepoTarget dreamkastv1alpha1.ReviewAppManagerSpecInfraTarget

func (m InfraRepoTarget) GitSecretSelector() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("gitSecretRef is not set")
	}
	return m.GitSecretRef, nil
}
