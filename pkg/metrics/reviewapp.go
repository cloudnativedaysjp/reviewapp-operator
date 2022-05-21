package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace = "reviewapp_operator"
)

var (
	UpVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "up_environment",
		Help:      "review-app is up when this flag equals 1",
	}, []string{"name", "namespace", "appOrganization", "appRepository", "infraOrganization", "infraRepository"})
	RequestToGitHubApiCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "github_api_requests_count",
		Help:      "The number of Requesting to GitHub API",
	}, []string{"name", "namespace", "kind"})
)

func Register(registry prometheus.Registerer) {
	registry.MustRegister(UpVec)
	registry.MustRegister(RequestToGitHubApiCounterVec)
}
