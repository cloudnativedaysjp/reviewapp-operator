package models

import (
	"golang.org/x/xerrors"
	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

const (
	LabelReviewAppNameForJob = "dreamkast.cloudnativedays.jp/parent-reviewapp"
)

type JobTemplate dreamkastv1alpha1.JobTemplate

func (m JobTemplate) GenerateJob(ra ReviewApp, pr PullRequest, v Templator) (*batchv1.Job, error) {
	var template string
	var err error
	/* TODO
	if pr.IsCandidate() {
		template = m.GetCandidateStr()
	} else {
		template = m.GetStableStr()
	}
	*/
	template = m.Spec.Template
	jobStr, err := v.Templating(template)
	if err != nil {
		return nil, err
	}
	var job batchv1.Job
	if err := yaml.Unmarshal([]byte(jobStr), &job); err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	// Set Label

	job.SetLabels(map[string]string{LabelReviewAppNameForJob: ra.Name})
	return &job, nil
}
