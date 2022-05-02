package models

import (
	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type ManifestsTemplate dreamkastv1alpha1.ManifestsTemplate

func (m ManifestsTemplate) StableMap() (map[string]string, error) {
	return m.toMapStrStr(m.Spec.StableData)
}

func (m ManifestsTemplate) CandidateMap() (map[string]string, error) {
	return m.toMapStrStr(m.Spec.CandidateData)
}

func (m ManifestsTemplate) toMapStrStr(d map[string]unstructured.Unstructured) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range d {
		b, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
		result[k] = string(b)
	}
	return result, nil
}

func (m ManifestsTemplate) AppendOrUpdate(mt ManifestsTemplate) ManifestsTemplate {
	if m.Spec.StableData == nil {
		m.Spec.StableData = make(map[string]unstructured.Unstructured)
	}
	for k, v := range mt.Spec.StableData {
		m.Spec.StableData[k] = v
	}
	if m.Spec.CandidateData == nil {
		m.Spec.CandidateData = make(map[string]unstructured.Unstructured)
	}
	for k, v := range mt.Spec.CandidateData {
		m.Spec.CandidateData[k] = v
	}
	return m
}

func (m ManifestsTemplate) GenerateManifests(pr PullRequest, v Templator) (Manifests, error) {
	var template map[string]string
	var err error
	if pr.IsCandidate() {
		template, err = m.CandidateMap()
	} else {
		template, err = m.StableMap()
	}
	if err != nil {
		return nil, err
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
