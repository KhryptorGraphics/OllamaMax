package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// PrometheusExporter exports metrics to Prometheus
type PrometheusExporter struct {
	config   *PrometheusConfig
	registry *prometheus.Registry

	// Prometheus metric types
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec

	// HTTP server for metrics endpoint
	server *http.Server

	// Lifecycle
	mu      sync.RWMutex
	started bool
}

// PrometheusConfig configures the Prometheus exporter
type PrometheusConfig struct {
	// HTTP server settings
	ListenAddress string `json:"listen_address"`
	MetricsPath   string `json:"metrics_path"`

	// Metric settings
	Namespace string `json:"namespace"`
	Subsystem string `json:"subsystem"`

	// Collection settings
	EnableGoMetrics      bool          `json:"enable_go_metrics"`
	EnableProcessMetrics bool          `json:"enable_process_metrics"`
	GatherInterval       time.Duration `json:"gather_interval"`

	// HTTP settings
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// MetricNamingConvention defines standard metric naming
type MetricNamingConvention struct {
	// Component prefixes
	SchedulerPrefix      string
	ConsensusPrefix      string
	P2PPrefix            string
	APIPrefix            string
	FaultTolerancePrefix string
	ModelPrefix          string

	// Common suffixes
	TotalSuffix    string
	DurationSuffix string
	SizeSuffix     string
	CountSuffix    string
	RateSuffix     string
	ErrorsSuffix   string
	SuccessSuffix  string
}

// StandardMetricNames provides standardized metric names
var StandardMetricNames = &MetricNamingConvention{
	// Component prefixes
	SchedulerPrefix:      "scheduler",
	ConsensusPrefix:      "consensus",
	P2PPrefix:            "p2p",
	APIPrefix:            "api",
	FaultTolerancePrefix: "fault_tolerance",
	ModelPrefix:          "model",

	// Common suffixes
	TotalSuffix:    "total",
	DurationSuffix: "duration_seconds",
	SizeSuffix:     "size_bytes",
	CountSuffix:    "count",
	RateSuffix:     "rate",
	ErrorsSuffix:   "errors_total",
	SuccessSuffix:  "success_total",
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(config *PrometheusConfig) *PrometheusExporter {
	if config == nil {
		config = DefaultPrometheusConfig()
	}

	// Create custom registry
	registry := prometheus.NewRegistry()

	// Add Go and process metrics if enabled
	if config.EnableGoMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
	}
	if config.EnableProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	return &PrometheusExporter{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}
}

// DefaultPrometheusConfig returns default configuration
func DefaultPrometheusConfig() *PrometheusConfig {
	return &PrometheusConfig{
		ListenAddress:        ":9090",
		MetricsPath:          "/metrics",
		Namespace:            "ollama",
		Subsystem:            "distributed",
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
		GatherInterval:       15 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		IdleTimeout:          60 * time.Second,
	}
}

// Start starts the Prometheus exporter
func (pe *PrometheusExporter) Start(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.started {
		return nil
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle(pe.config.MetricsPath, promhttp.HandlerFor(pe.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Registry:          pe.registry,
	}))

	// Add health check endpoint
	mux.HandleFunc("/health", pe.healthHandler)
	mux.HandleFunc("/ready", pe.readyHandler)

	pe.server = &http.Server{
		Addr:         pe.config.ListenAddress,
		Handler:      mux,
		ReadTimeout:  pe.config.ReadTimeout,
		WriteTimeout: pe.config.WriteTimeout,
		IdleTimeout:  pe.config.IdleTimeout,
	}

	// Start server in background
	go func() {
		log.Info().
			Str("address", pe.config.ListenAddress).
			Str("metrics_path", pe.config.MetricsPath).
			Msg("Starting Prometheus metrics server")

		if err := pe.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Prometheus metrics server failed")
		}
	}()

	pe.started = true
	log.Info().Msg("Prometheus exporter started")
	return nil
}

// Stop stops the Prometheus exporter
func (pe *PrometheusExporter) Stop() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.started || pe.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := pe.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown Prometheus server gracefully")
		return err
	}

	pe.started = false
	log.Info().Msg("Prometheus exporter stopped")
	return nil
}

// RegisterCounter registers a new counter metric
func (pe *PrometheusExporter) RegisterCounter(name, help string, labels []string) *prometheus.CounterVec {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	fullName := pe.buildMetricName(name)

	if existing, exists := pe.counters[fullName]; exists {
		return existing
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: pe.config.Namespace,
			Subsystem: pe.config.Subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	pe.registry.MustRegister(counter)
	pe.counters[fullName] = counter

	return counter
}

// RegisterGauge registers a new gauge metric
func (pe *PrometheusExporter) RegisterGauge(name, help string, labels []string) *prometheus.GaugeVec {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	fullName := pe.buildMetricName(name)

	if existing, exists := pe.gauges[fullName]; exists {
		return existing
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: pe.config.Namespace,
			Subsystem: pe.config.Subsystem,
			Name:      name,
			Help:      help,
		},
		labels,
	)

	pe.registry.MustRegister(gauge)
	pe.gauges[fullName] = gauge

	return gauge
}

// RegisterHistogram registers a new histogram metric
func (pe *PrometheusExporter) RegisterHistogram(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	fullName := pe.buildMetricName(name)

	if existing, exists := pe.histograms[fullName]; exists {
		return existing
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: pe.config.Namespace,
			Subsystem: pe.config.Subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	pe.registry.MustRegister(histogram)
	pe.histograms[fullName] = histogram

	return histogram
}

// RegisterSummary registers a new summary metric
func (pe *PrometheusExporter) RegisterSummary(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	fullName := pe.buildMetricName(name)

	if existing, exists := pe.summaries[fullName]; exists {
		return existing
	}

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  pe.config.Namespace,
			Subsystem:  pe.config.Subsystem,
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)

	pe.registry.MustRegister(summary)
	pe.summaries[fullName] = summary

	return summary
}

// buildMetricName builds a standardized metric name
func (pe *PrometheusExporter) buildMetricName(name string) string {
	// Ensure metric name follows Prometheus conventions
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")

	return fmt.Sprintf("%s_%s_%s", pe.config.Namespace, pe.config.Subsystem, name)
}

// healthHandler handles health check requests
func (pe *PrometheusExporter) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyHandler handles readiness check requests
func (pe *PrometheusExporter) readyHandler(w http.ResponseWriter, r *http.Request) {
	pe.mu.RLock()
	started := pe.started
	pe.mu.RUnlock()

	if started {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
	}
}

// GetRegistry returns the Prometheus registry
func (pe *PrometheusExporter) GetRegistry() *prometheus.Registry {
	return pe.registry
}

// GetMetricsURL returns the metrics endpoint URL
func (pe *PrometheusExporter) GetMetricsURL() string {
	return fmt.Sprintf("http://%s%s", pe.config.ListenAddress, pe.config.MetricsPath)
}

// Global Prometheus exporter instance for middleware
var globalPrometheusExporter *PrometheusExporter
var globalPrometheusOnce sync.Once

// SetGlobalPrometheusExporter sets the global Prometheus exporter for middleware
func SetGlobalPrometheusExporter(exporter *PrometheusExporter) {
	globalPrometheusExporter = exporter
}

// GinMetricsMiddleware provides Gin middleware for Prometheus metrics collection
func GinMetricsMiddleware() gin.HandlerFunc {
	// Initialize metrics once
	globalPrometheusOnce.Do(func() {
		if globalPrometheusExporter == nil {
			// Create default exporter if none set
			globalPrometheusExporter = NewPrometheusExporter(DefaultPrometheusConfig())
		}

		// Register HTTP metrics
		globalPrometheusExporter.RegisterCounter(
			"http_requests_total",
			"Total number of HTTP requests",
			[]string{"method", "path", "status"},
		)

		globalPrometheusExporter.RegisterHistogram(
			"http_request_duration_seconds",
			"HTTP request duration in seconds",
			[]string{"method", "path", "status"},
			[]float64{0.001, 0.01, 0.1, 0.5, 1, 2.5, 5, 10},
		)

		globalPrometheusExporter.RegisterGauge(
			"http_requests_in_flight",
			"Number of HTTP requests currently being processed",
			[]string{"method", "path"},
		)
	})

	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// If path is empty (404), use the request URI
		if path == "" {
			path = c.Request.URL.Path
		}

		// Increment in-flight requests
		inFlightGauge := globalPrometheusExporter.gauges["ollama_distributed_http_requests_in_flight"]
		if inFlightGauge != nil {
			inFlightGauge.WithLabelValues(method, path).Inc()
			defer inFlightGauge.WithLabelValues(method, path).Dec()
		}

		// Process request
		c.Next()

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		// Increment request counter
		requestCounter := globalPrometheusExporter.counters["ollama_distributed_http_requests_total"]
		if requestCounter != nil {
			requestCounter.WithLabelValues(method, path, status).Inc()
		}

		// Record request duration
		durationHistogram := globalPrometheusExporter.histograms["ollama_distributed_http_request_duration_seconds"]
		if durationHistogram != nil {
			durationHistogram.WithLabelValues(method, path, status).Observe(duration)
		}
	}
}
