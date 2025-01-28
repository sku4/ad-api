package middleware

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//nolint:mnd
var (
	InFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "api_rest",
		Subsystem: "http",
		Name:      "in_flight_requests_total",
	})
	SummaryResponseTime = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "api_rest",
		Subsystem: "http",
		Name:      "summary_response_time_seconds",
		Objectives: map[float64]float64{
			0.5:  0.1,
			0.9:  0.01,
			0.99: 0.001,
		},
	})
	HistogramResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "api_rest",
			Subsystem: "http",
			Name:      "histogram_response_time_seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 4},
		},
		[]string{"code", "url"},
	)
)

func Metrics(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := NewResponseWrapper(w)

		startTime := time.Now()
		next.ServeHTTP(wrapper, r)
		duration := time.Since(startTime)

		SummaryResponseTime.Observe(duration.Seconds())
		HistogramResponseTime.
			WithLabelValues(http.StatusText(wrapper.statusCode), r.URL.Path).
			Observe(duration.Seconds())
	})
	wrappedHandler := promhttp.InstrumentHandlerInFlight(InFlightRequests, handler)

	return wrappedHandler
}
