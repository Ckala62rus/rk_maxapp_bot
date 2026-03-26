package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics definitions.
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	httpInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "in_flight_requests",
			Help:      "Current number of in-flight HTTP requests.",
		},
	)
)

// init registers metrics once on package load.
func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpInFlight)
}

// Metrics collects Prometheus metrics for each request.
func Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip metrics for metrics endpoint to avoid recursion.
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// Track in-flight requests.
			httpInFlight.Inc()
			defer httpInFlight.Dec()

			// Measure duration.
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			// Serve request.
			next.ServeHTTP(rec, r)

			// Resolve route pattern for labels.
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}

			// Update counters/histograms.
			httpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
			httpRequestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		})
	}
}
