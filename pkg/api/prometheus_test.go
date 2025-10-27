package api

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusMetricsRegistration(t *testing.T) {
	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()

	// Define metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Register metrics
	err := registry.Register(httpRequestsTotal)
	require.NoError(t, err, "Failed to register http_requests_total")

	err = registry.Register(httpRequestDuration)
	require.NoError(t, err, "Failed to register http_request_duration_seconds")

	err = registry.Register(httpRequestsInFlight)
	require.NoError(t, err, "Failed to register http_requests_in_flight")

	// Verify metrics can be incremented
	httpRequestsTotal.WithLabelValues("GET", "/api/v1/health", "200").Inc()
	httpRequestDuration.WithLabelValues("GET", "/api/v1/health", "200").Observe(0.123)
	httpRequestsInFlight.Inc()
	httpRequestsInFlight.Dec()

	// Gather metrics to verify they exist
	metricFamilies, err := registry.Gather()
	require.NoError(t, err, "Failed to gather metrics")

	// Should have 3 metric families
	assert.Equal(t, 3, len(metricFamilies), "Expected 3 metric families")

	// Verify metric names
	metricNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		metricNames[*mf.Name] = true
	}

	assert.True(t, metricNames["http_requests_total"], "Missing http_requests_total metric")
	assert.True(t, metricNames["http_request_duration_seconds"], "Missing http_request_duration_seconds metric")
	assert.True(t, metricNames["http_requests_in_flight"], "Missing http_requests_in_flight metric")
}

func TestPrometheusMetricsLabels(t *testing.T) {
	registry := prometheus.NewRegistry()

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	err := registry.Register(httpRequestsTotal)
	require.NoError(t, err)

	// Test different label combinations
	httpRequestsTotal.WithLabelValues("GET", "/api/v1/models", "200").Inc()
	httpRequestsTotal.WithLabelValues("POST", "/api/v1/models", "201").Inc()
	httpRequestsTotal.WithLabelValues("GET", "/api/v1/models", "404").Inc()
	httpRequestsTotal.WithLabelValues("DELETE", "/api/v1/models/123", "204").Inc()

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find the http_requests_total metric
	var found bool
	for _, mf := range metricFamilies {
		if *mf.Name == "http_requests_total" {
			found = true
			// Should have 4 different label combinations
			assert.Equal(t, 4, len(mf.Metric), "Expected 4 different label combinations")
		}
	}

	assert.True(t, found, "http_requests_total metric not found")
}

func TestPrometheusHistogramBuckets(t *testing.T) {
	registry := prometheus.NewRegistry()

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	err := registry.Register(httpRequestDuration)
	require.NoError(t, err)

	// Record some observations
	httpRequestDuration.WithLabelValues("GET", "/api/v1/health", "200").Observe(0.001)
	httpRequestDuration.WithLabelValues("GET", "/api/v1/health", "200").Observe(0.5)
	httpRequestDuration.WithLabelValues("GET", "/api/v1/health", "200").Observe(2.5)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find the histogram metric
	var found bool
	for _, mf := range metricFamilies {
		if *mf.Name == "http_request_duration_seconds" {
			found = true
			assert.Equal(t, 1, len(mf.Metric), "Expected 1 histogram")

			// Verify histogram has buckets
			histogram := mf.Metric[0].Histogram
			assert.NotNil(t, histogram, "Histogram should not be nil")
			assert.Greater(t, len(histogram.Bucket), 0, "Histogram should have buckets")
			assert.Equal(t, uint64(3), *histogram.SampleCount, "Expected 3 samples")
		}
	}

	assert.True(t, found, "http_request_duration_seconds metric not found")
}

func TestPrometheusGaugeOperations(t *testing.T) {
	registry := prometheus.NewRegistry()

	httpRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	err := registry.Register(httpRequestsInFlight)
	require.NoError(t, err)

	// Test gauge operations
	httpRequestsInFlight.Set(0)
	httpRequestsInFlight.Inc()
	httpRequestsInFlight.Inc()
	httpRequestsInFlight.Inc()
	httpRequestsInFlight.Dec()

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find the gauge metric
	var found bool
	var gaugeValue float64
	for _, mf := range metricFamilies {
		if *mf.Name == "http_requests_in_flight" {
			found = true
			assert.Equal(t, 1, len(mf.Metric), "Expected 1 gauge")
			gaugeValue = *mf.Metric[0].Gauge.Value
		}
	}

	assert.True(t, found, "http_requests_in_flight metric not found")
	assert.Equal(t, float64(2), gaugeValue, "Expected gauge value to be 2")
}
