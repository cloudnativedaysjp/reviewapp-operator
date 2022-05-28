package template

import (
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

func (v Templator) Application(m dreamkastv1alpha1.ApplicationTemplate, pr dreamkastv1alpha1.PullRequest) (dreamkastv1alpha1.Application, error) {
	var template string
	var err error
	if pr.IsCandidate() {
		template = string(m.Spec.CandidateTemplate)
	} else {
		template = string(m.Spec.StableTemplate)
	}
	application, err := v.Templating(template)
	if err != nil {
		return "", err
	}
	return dreamkastv1alpha1.Application(application), nil
}
