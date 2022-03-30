package models

import (
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

type ManifestsTemplate dreamkastv1alpha1.ManifestsTemplate

func (m ManifestsTemplate) StableMap() map[string]string {
	return m.Spec.StableData
}

func (m ManifestsTemplate) CandidateMap() map[string]string {
	return m.Spec.CandidateData
}

func (m ManifestsTemplate) AppendOrUpdate(mt ManifestsTemplate) ManifestsTemplate {
	if m.Spec.StableData == nil {
		m.Spec.StableData = make(map[string]string)
	}
	for k, v := range mt.StableMap() {
		m.Spec.StableData[k] = v
	}
	if m.Spec.CandidateData == nil {
		m.Spec.CandidateData = make(map[string]string)
	}
	for k, v := range mt.CandidateMap() {
		m.Spec.CandidateData[k] = v
	}
	return m
}

func (m ManifestsTemplate) GenerateManifests(pr PullRequest, v Templator) (Manifests, error) {
	var template map[string]string
	var err error
	if pr.IsCandidate() {
		template = m.CandidateMap()
	} else {
		template = m.StableMap()
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
