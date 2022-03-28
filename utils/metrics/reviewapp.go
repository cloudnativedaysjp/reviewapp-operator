package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
)

const (
	metricsRaNamespace = "reviewapp_status"
)

var (
	UpVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsRaNamespace,
		Name:      "up",
		Help:      "Operator's status is healthy when this flag equals 1",
	}, []string{"name", "namespace", "appOrganization", "appRepository", "infraOrganization", "infraRepository"})
	UnknownVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsRaNamespace,
		Name:      "unknown",
		Help:      "Operator's status is unknown when this flag equals 1",
	}, []string{"name", "namespace", "appOrganization", "appRepository", "infraOrganization", "infraRepository"})
)

func init() {
	metrics.Registry.MustRegister(UpVec)
}

func SetMetricsUp(ra dreamkastv1alpha1.ReviewApp) {
	UpVec.WithLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	).Set(1)
}

func SetMetricsUnknown(ra dreamkastv1alpha1.ReviewApp) {
	UnknownVec.WithLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	).Set(1)
}

func RemoveMetrics(ra dreamkastv1alpha1.ReviewApp) {
	UpVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	)
	UnknownVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	)
}
