package template

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

func (v Templator) ReviewApp(
	m dreamkastv1alpha1.ReviewAppManager, pr dreamkastv1alpha1.PullRequest,
) (dreamkastv1alpha1.ReviewApp, error) {
	ra := dreamkastv1alpha1.ReviewApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.ReviewAppName(pr.Spec),
			Namespace: m.Namespace,
		},
		Spec: dreamkastv1alpha1.ReviewAppSpec{
			ReviewAppCommonSpec: dreamkastv1alpha1.ReviewAppCommonSpec{
				AppTarget:   m.Spec.AppTarget,
				InfraTarget: m.Spec.InfraTarget,
				Variables:   m.Spec.Variables,
				PreStopJob:  m.Spec.PreStopJob,
			},
			PullRequest: dreamkastv1alpha1.NamespacedName{
				Namespace: m.Namespace,
				Name:      m.Spec.AppTarget.PullRequestName(pr.Spec.Number),
			},
		},
		Status: dreamkastv1alpha1.ReviewAppStatus{
			Sync: dreamkastv1alpha1.ReviewAppStatusSync{},
			PullRequestCache: dreamkastv1alpha1.ReviewAppStatusPullRequestCache{
				HeadBranch:       pr.Status.HeadBranch,
				LatestCommitHash: pr.Status.LatestCommitHash,
				Title:            pr.Status.Title,
				Labels:           pr.Status.Labels,
				SyncedTimestamp:  metav1.Now(),
			},
		},
	}
	{ // template from ram.Spec.AppConfig to ra.Spec.AppConfig
		out, err := yaml.Marshal(&m.Spec.AppConfig)
		if err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
		appConfigStr, err := v.Templating(string(out))
		if err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
		if err := yaml.Unmarshal([]byte(appConfigStr), &ra.Spec.AppConfig); err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
	}
	{ // template from ram.Spec.InfraConfig to ra.Spec.InfraConfig
		out, err := yaml.Marshal(&m.Spec.InfraConfig)
		if err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
		infraConfigStr, err := v.Templating(string(out))
		if err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
		if err := yaml.Unmarshal([]byte(infraConfigStr), &ra.Spec.InfraConfig); err != nil {
			return dreamkastv1alpha1.ReviewApp{}, err
		}
	}
	return ra, nil
}
