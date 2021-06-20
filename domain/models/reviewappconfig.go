package models

import (
	mapset "github.com/deckarep/golang-set"

	dreamkastv1beta1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1beta1"
)

type ReviewAppConfig struct {
	ReviewApp           dreamkastv1beta1.ReviewApp
	ApplicationTemplate dreamkastv1beta1.ApplicationTemplate
	ManifestsTemplate   mapset.Set // set of dreamkastv1beta1.ManifestsTemplate
}
