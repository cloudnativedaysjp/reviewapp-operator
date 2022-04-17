package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace   = "reviewapp_operator"
	reviewappSubsystem = "reviewapp"
)

var (
	UpVec                        *prometheus.GaugeVec
	RequestToGitHubApiCounterVec *prometheus.CounterVec
)

func Register(registry prometheus.Registerer) {
	UpVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: reviewappSubsystem,
		Name:      "up",
		Help:      "Operator's status is healthy when this flag equals 1",
	}, []string{"name", "namespace", "appOrganization", "appRepository", "infraOrganization", "infraRepository"})
	registry.MustRegister(UpVec)

	RequestToGitHubApiCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Subsystem: reviewappSubsystem,
		Name:      "",
		Help:      "Operator's status is healthy when this flag equals 1",
	}, []string{"name", "namespace", "kind"})
	registry.MustRegister(RequestToGitHubApiCounterVec)
}
