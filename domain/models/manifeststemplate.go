package models

import (
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

type ManifestsTemplate dreamkastv1alpha1.ManifestsTemplate

func (m ManifestsTemplate) GetStableMap() map[string]string {
	return m.Spec.StableData
}

func (m ManifestsTemplate) GetCandidateMap() map[string]string {
	return m.Spec.CandidateData
}

func (m ManifestsTemplate) AppendOrUpdate(mt ManifestsTemplate) ManifestsTemplate {
	for k, v := range mt.GetStableMap() {
		m.Spec.StableData[k] = v
	}
	for k, v := range mt.GetCandidateMap() {
		m.Spec.CandidateData[k] = v
	}
	return m
}

func (m ManifestsTemplate) GenerateManifests(pr PullRequest, v Templator) (Manifests, error) {
	var template map[string]string
	var err error
	if pr.IsCandidate() {
		template = m.GetCandidateMap()
	} else {
		template = m.GetStableMap()
	}
	manifests := make(map[string]string)
	for key, val := range template {
		manifests[key], err = v.Templating(val)
		if err != nil {
			return nil, err
		}
	}
	return Manifests(manifests), nil
}

/* Manifests  */

type Manifests map[string]string
