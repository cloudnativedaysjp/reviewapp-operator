package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/cloudnativedaysjp/reviewapp-operator/domain/models"
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
)

func init() {
	metrics.Registry.MustRegister(UpVec)
}

func SetMetricsUp(ra models.ReviewApp) {
	UpVec.WithLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	).Set(1)
}

func RemoveMetrics(ra models.ReviewApp) {
	UpVec.DeleteLabelValues(
		ra.Name,
		ra.Namespace,
		ra.Spec.AppTarget.Organization,
		ra.Spec.AppTarget.Repository,
		ra.Spec.InfraTarget.Organization,
		ra.Spec.InfraTarget.Organization,
	)
}
