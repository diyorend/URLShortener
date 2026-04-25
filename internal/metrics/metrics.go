package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metric objects.
// promauto registers them automatically — no manual Register() calls needed.
type Metrics struct {
	// Counter: only goes up. Total shorten requests, labeled by outcome.
	ShortenTotal *prometheus.CounterVec

	// Counter: total redirects, labeled by "hit" or "not_found".
	RedirectTotal *prometheus.CounterVec

	// Histogram: records latency distributions.
	// Automatically gives you p50, p95, p99 in Grafana with no extra work.
	RequestDuration *prometheus.HistogramVec
}

// New creates and registers all metrics with the default Prometheus registry.
func New() *Metrics {
	return &Metrics{
		ShortenTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "urlshortener_shorten_total",
			Help: "Total number of shorten requests.",
		}, []string{"status"}),

		RedirectTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "urlshortener_redirect_total",
			Help: "Total number of redirect lookups.",
		}, []string{"status"}),

		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "urlshortener_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method"}),
	}
}
