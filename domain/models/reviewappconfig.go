package models

import (
	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

type ReviewAppConfig struct {
	ReviewAppManager    dreamkastv1beta1.ReviewAppManager
	ApplicationTemplate dreamkastv1beta1.ApplicationTemplate
	ManifestsTemplate   map[string]string
}

func NewReviewAppConfig() *ReviewAppConfig {
	rac := &ReviewAppConfig{}
	rac.ManifestsTemplate = make(map[string]string)
	return rac
}
