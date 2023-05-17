package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func HTTPMetrics(reg prometheus.Registerer) func(http.Handler) http.Handler {
	httpRequestsDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_requests_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.ExponentialBuckets(0.05, 2, 8),
	}, []string{"method", "endpoint", "status_code"})
	reg.MustRegister(httpRequestsDuration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()

			rw := newStatusWriter(w)
			next.ServeHTTP(rw, r)

			statusCode := strconv.Itoa(rw.statusCode)
			d := time.Since(t).Seconds()
			// FIXME(fkrestan): this needs some URL path filtering.
			httpRequestsDuration.WithLabelValues(
				r.Method, r.URL.Path, statusCode).Observe(d)
		})
	}
}
