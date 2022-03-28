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
	GatNamespaceName() types.NamespacedName
	GetAppRepoTarget() AppRepoTarget
	GetInfraRepoTarget() InfraRepoTarget
	GetInfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig
	GetVariables() []string
}

/* ReviewApp */

type ReviewApp dreamkastv1alpha1.ReviewApp

func NewReviewApp(ra *dreamkastv1alpha1.ReviewApp) ReviewApp {
	return ReviewApp(*ra)
}

func (m ReviewApp) GatNamespaceName() types.NamespacedName {
	return types.NamespacedName{Namespace: m.Namespace, Name: m.Name}
}
func (m ReviewApp) GetAppRepoTarget() AppRepoTarget {
	return AppRepoTarget(m.Spec.AppTarget)
}
func (m ReviewApp) GetInfraRepoTarget() InfraRepoTarget {
	return InfraRepoTarget(m.Spec.InfraTarget)
}
func (m ReviewApp) GetInfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig {
	return m.Spec.InfraConfig
}
func (m ReviewApp) GetVariables() []string {
	return m.Spec.Variables
}

func (m ReviewApp) ToReviewAppCR() *dreamkastv1alpha1.ReviewApp {
	ra := dreamkastv1alpha1.ReviewApp(m)
	return &ra
}

func (m ReviewApp) GetAtFilepath() string {
	return m.Spec.InfraConfig.ArgoCDApp.Filepath
}

func (m ReviewApp) GetMtDirpath() string {
	return m.Spec.InfraConfig.Manifests.Dirpath
}

func (m ReviewApp) GetPrNum() int {
	return m.Spec.AppPrNum
}

func (m ReviewApp) UpdateStatusOfAppRepo(pr PullRequest) (ReviewApp, bool) {
	updated := false
	m.Status.Sync.AppRepoBranch = pr.Branch
	if m.Status.Sync.AppRepoLatestCommitSha != pr.HeadCommitSha {
		updated = true
	}
	m.Status.Sync.AppRepoLatestCommitSha = pr.HeadCommitSha
	return m, updated
}

func (m ReviewApp) UpdateStatusOfApplication(application Application) (ReviewApp, bool, error) {
	updated := false
	argocdAppNamespacedName, err := application.GetNamespacedName()
	if err != nil {
		return ReviewApp{}, false, err
	}
	if m.Status.Sync.ApplicationName != argocdAppNamespacedName.Name || m.Status.Sync.ApplicationNamespace != argocdAppNamespacedName.Namespace {
		updated = true
	}
	if !reflect.DeepEqual(application, m.Status.ManifestsCache.Application) {
		updated = true
	}
	m.Status.Sync.ApplicationName = argocdAppNamespacedName.Name
	m.Status.Sync.ApplicationNamespace = argocdAppNamespacedName.Namespace
	return m, updated, nil
}

func (m ReviewApp) StatusOfManifestsWasUpdated(manifests Manifests) bool {
	updated := false
	if !reflect.DeepEqual(manifests, m.Status.ManifestsCache.Manifests) {
		updated = true
	}
	return updated
}

func (m ReviewApp) HasApplicationBeenUpdated(hash string) bool {
	return m.Status.Sync.AppRepoLatestCommitSha != hash
}

func (m ReviewApp) HasMessageAlreadyBeenSent() bool {
	return m.Spec.AppConfig.Message == "" || (!m.Spec.AppConfig.SendMessageEveryTime && m.Status.AlreadySentMessage)
}

func (m ReviewApp) HavingPreStopJob() bool {
	return m.Spec.PreStopJob.Namespace != "" && m.Spec.PreStopJob.Name != ""
}

/* ReviewAppManager */

type ReviewAppManager dreamkastv1alpha1.ReviewAppManager

func NewReviewAppManager(ram dreamkastv1alpha1.ReviewAppManager) ReviewAppManager {
	return ReviewAppManager(ram)
}

func (m ReviewAppManager) GatNamespaceName() types.NamespacedName {
	return types.NamespacedName{Namespace: m.Namespace, Name: m.Name}
}
func (m ReviewAppManager) GetAppRepoTarget() AppRepoTarget {
	return AppRepoTarget(m.Spec.AppTarget)
}
func (m ReviewAppManager) GetInfraRepoTarget() InfraRepoTarget {
	return InfraRepoTarget(m.Spec.InfraTarget)
}
func (m ReviewAppManager) GetInfraRepoConfig() dreamkastv1alpha1.ReviewAppManagerSpecInfraConfig {
	return m.Spec.InfraConfig
}
func (m ReviewAppManager) GetVariables() []string {
	return m.Spec.Variables
}

func (m ReviewAppManager) GenerateReviewApp(pr PullRequest, v Templator) (ReviewApp, error) {
	ra := ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.GetReviewAppName(pr),
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

func (m ReviewAppManager) GetReviewAppName(pr PullRequest) string {
	return fmt.Sprintf("%s-%s-%s-%d",
		m.Name,
		strings.ToLower(pr.Organization),
		strings.ToLower(pr.Repository),
		pr.Number,
	)
}

func (m ReviewAppManager) ListOutOfSyncPullRequests(prs []PullRequest) []PullRequest {
	var result []PullRequest
	for _, a := range m.Status.SyncedPullRequests {
		for _, b := range prs {
			if !(a.Organization == b.Organization && a.Repository == b.Repository && a.Number == b.Number) {
				result = append(result, b)
			}
		}
	}
	return result
}

/* AppRepoTarget Or InfraRepoTarget */

type AppOrInfraRepoTarget interface {
	GetGitSecretRef() (*corev1.SecretKeySelector, error)
}

/* AppRepoTarget  */

type AppRepoTarget dreamkastv1alpha1.ReviewAppManagerSpecAppTarget

func (m AppRepoTarget) GetGitSecretRef() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("TODO")
	}
	return m.GitSecretRef, nil
}

/* InfraRepoTarget  */

type InfraRepoTarget dreamkastv1alpha1.ReviewAppManagerSpecInfraTarget

func (m InfraRepoTarget) GetGitSecretRef() (*corev1.SecretKeySelector, error) {
	if m.GitSecretRef == nil || reflect.ValueOf(m.GitSecretRef).IsNil() {
		return nil, fmt.Errorf("TODO")
	}
	return m.GitSecretRef, nil
}
