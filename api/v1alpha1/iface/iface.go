package v1alpha1_iface

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

// ReviewAppOrReviewAppManager
type ReviewAppOrReviewAppManager interface {
	GetAppTarget() dreamkastv1alpha1.ReviewAppCommonSpecAppTarget
	GetAppConfig() dreamkastv1alpha1.ReviewAppCommonSpecAppConfig
	GetInfraTarget() dreamkastv1alpha1.ReviewAppCommonSpecInfraTarget
	GetInfraConfig() dreamkastv1alpha1.ReviewAppCommonSpecInfraConfig
	GetPreStopJob() types.NamespacedName
	GetVariables() []string
}

type AppOrInfraRepoTarget interface {
	GitSecretSelector() (*corev1.SecretKeySelector, error)
}
