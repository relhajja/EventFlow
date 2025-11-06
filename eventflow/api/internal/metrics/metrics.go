package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FunctionInvocations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eventflow_function_invocations_total",
			Help: "Total number of function invocations",
		},
		[]string{"function", "namespace"},
	)

	FunctionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "eventflow_function_duration_seconds",
			Help:    "Duration of function invocations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"function", "namespace"},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eventflow_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ActiveFunctions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eventflow_active_functions",
			Help: "Number of active functions",
		},
		[]string{"namespace"},
	)
)

func Init() {
	// Metrics are automatically registered with prometheus.DefaultRegisterer
	// via promauto
}
