package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

const (
	metricsRaNamespace = "reviewapp"
)

var (
	HealthyVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsRaNamespace,
		Name:      "healthy",
		Help:      "The cluster status about healthy condition",
	}, []string{"name", "namespace", "appOrganization", "appRepository", "infraOrganization", "infraRepository"})
)

func init() {
	metrics.Registry.MustRegister(HealthyVec)
}

func SetMetrics(ra dreamkastv1alpha1.ReviewApp) {
	HealthyVec.WithLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	).Set(1)
}

func RemoveMetrics(ra dreamkastv1alpha1.ReviewApp) {
	HealthyVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	)
}
