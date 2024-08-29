package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// prometheus metrics
type metrics struct {
	SPQueryCount       *prometheus.CounterVec
	httpDuration       *prometheus.HistogramVec
	durationSummary    prometheus.Summary
	responseStatusCode *prometheus.CounterVec
	totalRequests      *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		SPQueryCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "navigatorx",
			Name:      "shortestpath_query_count",
			Help:      "The total number of shortest path query",
		}, []string{"shortestpath_bidirectional_dijsktra"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "navigatorx",
			Name:      "request_duration_seconds",
			Help:      "The duration of request",
			Buckets:   []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3}, // 0.001 = 1ms
		}, []string{"method", "path"}),
		durationSummary: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  "navigatorx",
			Name:       "shortestpath_request_duration_summary_seconds",
			Help:       "The duration of request",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}),
		responseStatusCode: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "navigatorx",
				Name:      "response_status_code",
				Help:      "The status code of http response",
			}, []string{"status", "method", "path"},
		),
		totalRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "navigatorx",
				Name:      "total_requests",
				Help:      "The total number of requests",
			}, []string{"path", "method", "status"},
		),
	}
	reg.MustRegister(m.SPQueryCount, m.httpDuration, m.durationSummary, m.responseStatusCode, m.totalRequests)
	return m
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func PromeHttpMiddleware(m *metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			rw := NewResponseWriter(w)
			timer := prometheus.NewTimer(m.httpDuration.With(prometheus.Labels{"method": r.Method, "path": path}))
			now := time.Now()

			next.ServeHTTP(rw, r)

			statusCode := rw.statusCode

			m.responseStatusCode.With(prometheus.Labels{"status": strconv.Itoa(statusCode), "method": r.Method, "path": path}).Inc()
			m.totalRequests.With(prometheus.Labels{"path": path, "method": r.Method, "status": strconv.Itoa(statusCode)}).Inc()
			timer.ObserveDuration()
			m.durationSummary.Observe(time.Since(now).Seconds())

		})
	}
}
