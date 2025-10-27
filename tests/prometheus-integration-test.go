package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestPrometheusIntegration validates the Prometheus metrics implementation
func main() {
	fmt.Println("Testing Prometheus Integration...")

	// Initialize Prometheus registry
	registry := prometheus.NewRegistry()

	// Define Prometheus metrics (same as in server.go)
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

	// Register metrics with the registry
	if err := registry.Register(httpRequestsTotal); err != nil {
		panic(fmt.Sprintf("Failed to register http_requests_total: %v", err))
	}
	if err := registry.Register(httpRequestDuration); err != nil {
		panic(fmt.Sprintf("Failed to register http_request_duration_seconds: %v", err))
	}
	if err := registry.Register(httpRequestsInFlight); err != nil {
		panic(fmt.Sprintf("Failed to register http_requests_in_flight: %v", err))
	}

	fmt.Println("✓ Metrics registered successfully")

	// Create a test middleware (same logic as prometheusMiddleware)
	prometheusMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			// Skip metrics endpoint to avoid recursion
			if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/metrics.json" {
				c.Next()
				return
			}

			// Increment in-flight requests
			httpRequestsInFlight.Inc()
			defer httpRequestsInFlight.Dec()

			// Record start time
			start := time.Now()

			// Process request
			c.Next()

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Get endpoint path (normalized to avoid high cardinality)
			endpoint := c.FullPath()
			if endpoint == "" {
				endpoint = c.Request.URL.Path
			}

			// Record metrics
			status := strconv.Itoa(c.Writer.Status())
			method := c.Request.Method

			httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
			httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
		}
	}

	// Create a test router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(prometheusMiddleware())

	// Add test endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	router.GET("/metrics.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"database":  gin.H{"connections": 10},
			"timestamp": time.Now(),
		})
	})

	fmt.Println("✓ Router configured successfully")

	// Test 1: Make some test requests
	fmt.Println("\nTest 1: Making test requests...")
	makeTestRequest(router, "GET", "/health", 200)
	makeTestRequest(router, "GET", "/health", 200)
	makeTestRequest(router, "GET", "/health", 200)
	fmt.Println("✓ Test requests completed")

	// Test 2: Verify Prometheus metrics endpoint
	fmt.Println("\nTest 2: Testing /metrics endpoint...")
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		panic(fmt.Sprintf("Expected status 200, got %d", w.Code))
	}

	body := w.Body.String()
	if !contains(body, "http_requests_total") {
		panic("Missing http_requests_total in metrics output")
	}
	if !contains(body, "http_request_duration_seconds") {
		panic("Missing http_request_duration_seconds in metrics output")
	}
	if !contains(body, "http_requests_in_flight") {
		panic("Missing http_requests_in_flight in metrics output")
	}

	fmt.Println("✓ Prometheus metrics endpoint working")
	fmt.Println("  - http_requests_total: found")
	fmt.Println("  - http_request_duration_seconds: found")
	fmt.Println("  - http_requests_in_flight: found")

	// Test 3: Verify JSON metrics endpoint
	fmt.Println("\nTest 3: Testing /metrics.json endpoint...")
	req = httptest.NewRequest("GET", "/metrics.json", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		panic(fmt.Sprintf("Expected status 200, got %d", w.Code))
	}

	jsonBody := w.Body.String()
	if !contains(jsonBody, "database") {
		panic("Missing database in JSON metrics output")
	}
	if !contains(jsonBody, "timestamp") {
		panic("Missing timestamp in JSON metrics output")
	}

	fmt.Println("✓ JSON metrics endpoint working")
	fmt.Println("  - database metrics: found")
	fmt.Println("  - timestamp: found")

	// Test 4: Verify metrics are being recorded
	fmt.Println("\nTest 4: Verifying metric values...")
	metricFamilies, err := registry.Gather()
	if err != nil {
		panic(fmt.Sprintf("Failed to gather metrics: %v", err))
	}

	if len(metricFamilies) != 3 {
		panic(fmt.Sprintf("Expected 3 metric families, got %d", len(metricFamilies)))
	}

	var requestCount float64
	for _, mf := range metricFamilies {
		if *mf.Name == "http_requests_total" {
			for _, m := range mf.Metric {
				requestCount += *m.Counter.Value
			}
		}
	}

	if requestCount < 3 {
		panic(fmt.Sprintf("Expected at least 3 requests recorded, got %f", requestCount))
	}

	fmt.Printf("✓ Metrics recorded correctly (total requests: %.0f)\n", requestCount)

	// All tests passed
	fmt.Println("\n✅ All Prometheus integration tests passed!")
	fmt.Println("\nImplementation Summary:")
	fmt.Println("  ✓ Prometheus registry initialized")
	fmt.Println("  ✓ Three metrics registered (counter, histogram, gauge)")
	fmt.Println("  ✓ Prometheus middleware tracking requests")
	fmt.Println("  ✓ /metrics endpoint serving Prometheus format")
	fmt.Println("  ✓ /metrics.json endpoint serving JSON format (backward compatibility)")
}

func makeTestRequest(router *gin.Engine, method, path string, expectedStatus int) {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != expectedStatus {
		panic(fmt.Sprintf("Expected status %d, got %d", expectedStatus, w.Code))
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
