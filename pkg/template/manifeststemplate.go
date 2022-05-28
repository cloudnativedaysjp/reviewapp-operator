package template

import (
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

func (v Templator) Manifests(m dreamkastv1alpha1.ManifestsTemplate, pr dreamkastv1alpha1.PullRequest) (dreamkastv1alpha1.Manifests, error) {
	var template map[string]string
	var err error
	if pr.IsCandidate() {
		template = m.Spec.CandidateData
	} else {
		template = m.Spec.StableData
	}
	manifests := make(map[string]string)
	for key, val := range template {
		manifests[key], err = v.Templating(val)
		if err != nil {
			return nil, err
		}
	}
	return dreamkastv1alpha1.Manifests(manifests), nil
}
