package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector collects and exposes system metrics
type MetricsCollector struct {
	config *MetricsConfig

	// Prometheus metrics
	registry *prometheus.Registry

	// System metrics
	systemMetrics *SystemMetrics

	// Application metrics
	appMetrics *ApplicationMetrics

	// P2P metrics
	p2pMetrics *P2PMetrics

	// HTTP server for metrics endpoint
	server *http.Server

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	// Server settings
	ListenAddress string `yaml:"listen_address"`
	MetricsPath   string `yaml:"metrics_path"`

	// Collection settings
	CollectionInterval  time.Duration `yaml:"collection_interval"`
	EnableSystemMetrics bool          `yaml:"enable_system_metrics"`
	EnableAppMetrics    bool          `yaml:"enable_app_metrics"`
	EnableP2PMetrics    bool          `yaml:"enable_p2p_metrics"`

	// Retention settings
	MetricsRetention time.Duration `yaml:"metrics_retention"`

	// Labels
	DefaultLabels map[string]string `yaml:"default_labels"`
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		ListenAddress:       ":9090",
		MetricsPath:         "/metrics",
		CollectionInterval:  15 * time.Second,
		EnableSystemMetrics: true,
		EnableAppMetrics:    true,
		EnableP2PMetrics:    true,
		MetricsRetention:    24 * time.Hour,
		DefaultLabels: map[string]string{
			"service": "ollama-distributed",
			"version": "1.0.0",
		},
	}
}

// SystemMetrics holds system-level Prometheus metrics
type SystemMetrics struct {
	// CPU metrics
	CPUUsage    prometheus.Gauge
	CPUCores    prometheus.Gauge
	LoadAverage *prometheus.GaugeVec

	// Memory metrics
	MemoryUsage     prometheus.Gauge
	MemoryTotal     prometheus.Gauge
	MemoryAvailable prometheus.Gauge
	GCDuration      prometheus.Histogram

	// Disk metrics
	DiskUsage *prometheus.GaugeVec
	DiskIO    *prometheus.CounterVec

	// Network metrics
	NetworkIO *prometheus.CounterVec

	// Process metrics
	ProcessCount    prometheus.Gauge
	FileDescriptors prometheus.Gauge
}

// ApplicationMetrics holds application-level Prometheus metrics
type ApplicationMetrics struct {
	// Request metrics
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
	HTTPErrors   *prometheus.CounterVec

	// Model metrics
	ModelsLoaded  prometheus.Gauge
	ModelRequests *prometheus.CounterVec
	ModelLatency  *prometheus.HistogramVec
	ModelErrors   *prometheus.CounterVec

	// Node metrics
	NodesActive prometheus.Gauge
	NodesTotal  prometheus.Gauge
	NodeHealth  *prometheus.GaugeVec

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
	CacheSize   *prometheus.GaugeVec

	// Queue metrics
	QueueSize    *prometheus.GaugeVec
	QueueLatency *prometheus.HistogramVec
}

// P2PMetrics holds P2P-specific Prometheus metrics
type P2PMetrics struct {
	// Connection metrics
	PeerConnections prometheus.Gauge
	PeerLatency     *prometheus.HistogramVec
	PeerErrors      *prometheus.CounterVec

	// Message metrics
	MessagesReceived *prometheus.CounterVec
	MessagesSent     *prometheus.CounterVec
	MessageLatency   *prometheus.HistogramVec

	// Consensus metrics
	ConsensusRounds  prometheus.Counter
	ConsensusLatency prometheus.Histogram
	LeaderChanges    prometheus.Counter

	// Bandwidth metrics
	BandwidthUsage  *prometheus.GaugeVec
	DataTransferred *prometheus.CounterVec
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MetricsConfig) *MetricsCollector {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create custom registry
	registry := prometheus.NewRegistry()

	collector := &MetricsCollector{
		config:   config,
		registry: registry,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize metrics
	if config.EnableSystemMetrics {
		collector.systemMetrics = collector.initSystemMetrics()
	}

	if config.EnableAppMetrics {
		collector.appMetrics = collector.initApplicationMetrics()
	}

	if config.EnableP2PMetrics {
		collector.p2pMetrics = collector.initP2PMetrics()
	}

	return collector
}

// Start starts the metrics collector
func (mc *MetricsCollector) Start() error {
	// Start metrics collection
	mc.wg.Add(1)
	go mc.runCollection()

	// Start HTTP server
	mux := http.NewServeMux()
	mux.Handle(mc.config.MetricsPath, promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", mc.healthHandler)

	mc.server = &http.Server{
		Addr:    mc.config.ListenAddress,
		Handler: mux,
	}

	go func() {
		if err := mc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	fmt.Printf("Metrics server started on %s%s\n", mc.config.ListenAddress, mc.config.MetricsPath)
	return nil
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop() error {
	mc.cancel()
	mc.wg.Wait()

	if mc.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return mc.server.Shutdown(ctx)
	}

	return nil
}

// initSystemMetrics initializes system metrics
func (mc *MetricsCollector) initSystemMetrics() *SystemMetrics {
	metrics := &SystemMetrics{
		CPUUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_cpu_usage_percent",
			Help: "Current CPU usage percentage",
		}),
		CPUCores: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_cpu_cores",
			Help: "Number of CPU cores",
		}),
		LoadAverage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_load_average",
			Help: "System load average",
		}, []string{"period"}),
		MemoryUsage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		}),
		MemoryTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_memory_total_bytes",
			Help: "Total system memory in bytes",
		}),
		MemoryAvailable: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_memory_available_bytes",
			Help: "Available system memory in bytes",
		}),
		GCDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "system_gc_duration_seconds",
			Help: "Garbage collection duration",
		}),
		DiskUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "system_disk_usage_bytes",
			Help: "Disk usage in bytes",
		}, []string{"device", "mountpoint"}),
		DiskIO: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "system_disk_io_bytes_total",
			Help: "Total disk I/O in bytes",
		}, []string{"device", "direction"}),
		NetworkIO: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "system_network_io_bytes_total",
			Help: "Total network I/O in bytes",
		}, []string{"interface", "direction"}),
		ProcessCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_processes_total",
			Help: "Total number of processes",
		}),
		FileDescriptors: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "system_file_descriptors_open",
			Help: "Number of open file descriptors",
		}),
	}

	// Register metrics
	mc.registry.MustRegister(
		metrics.CPUUsage,
		metrics.CPUCores,
		metrics.LoadAverage,
		metrics.MemoryUsage,
		metrics.MemoryTotal,
		metrics.MemoryAvailable,
		metrics.GCDuration,
		metrics.DiskUsage,
		metrics.DiskIO,
		metrics.NetworkIO,
		metrics.ProcessCount,
		metrics.FileDescriptors,
	)

	return metrics
}

// initP2PMetrics initializes P2P metrics
func (mc *MetricsCollector) initP2PMetrics() *P2PMetrics {
	metrics := &P2PMetrics{
		PeerConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "p2p_peer_connections_total",
			Help: "Number of active peer connections",
		}),
		PeerLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "p2p_peer_latency_seconds",
			Help: "Peer connection latency",
		}, []string{"peer_id"}),
		PeerErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "p2p_peer_errors_total",
			Help: "Total number of peer errors",
		}, []string{"peer_id", "error_type"}),
		MessagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "p2p_messages_received_total",
			Help: "Total number of messages received",
		}, []string{"message_type", "peer_id"}),
		MessagesSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "p2p_messages_sent_total",
			Help: "Total number of messages sent",
		}, []string{"message_type", "peer_id"}),
		MessageLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "p2p_message_latency_seconds",
			Help: "Message processing latency",
		}, []string{"message_type"}),
		ConsensusRounds: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "p2p_consensus_rounds_total",
			Help: "Total number of consensus rounds",
		}),
		ConsensusLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "p2p_consensus_latency_seconds",
			Help: "Consensus round latency",
		}),
		LeaderChanges: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "p2p_leader_changes_total",
			Help: "Total number of leader changes",
		}),
		BandwidthUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "p2p_bandwidth_usage_bytes_per_second",
			Help: "P2P bandwidth usage",
		}, []string{"direction"}),
		DataTransferred: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "p2p_data_transferred_bytes_total",
			Help: "Total data transferred over P2P",
		}, []string{"direction", "peer_id"}),
	}

	// Register metrics
	mc.registry.MustRegister(
		metrics.PeerConnections,
		metrics.PeerLatency,
		metrics.PeerErrors,
		metrics.MessagesReceived,
		metrics.MessagesSent,
		metrics.MessageLatency,
		metrics.ConsensusRounds,
		metrics.ConsensusLatency,
		metrics.LeaderChanges,
		metrics.BandwidthUsage,
		metrics.DataTransferred,
	)

	return metrics
}

// runCollection runs the metrics collection loop
func (mc *MetricsCollector) runCollection() {
	defer mc.wg.Done()

	ticker := time.NewTicker(mc.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-ticker.C:
			mc.collectMetrics()
		}
	}
}

// collectMetrics collects all enabled metrics
func (mc *MetricsCollector) collectMetrics() {
	if mc.config.EnableSystemMetrics && mc.systemMetrics != nil {
		mc.collectSystemMetrics()
	}

	if mc.config.EnableAppMetrics && mc.appMetrics != nil {
		mc.collectApplicationMetrics()
	}

	if mc.config.EnableP2PMetrics && mc.p2pMetrics != nil {
		mc.collectP2PMetrics()
	}
}

// collectSystemMetrics collects system-level metrics
func (mc *MetricsCollector) collectSystemMetrics() {
	// TODO: Implement actual system metrics collection
	// This would typically use system calls or libraries like gopsutil

	// Placeholder values for demonstration
	mc.systemMetrics.CPUUsage.Set(25.5)
	mc.systemMetrics.CPUCores.Set(8)
	mc.systemMetrics.LoadAverage.WithLabelValues("1m").Set(1.2)
	mc.systemMetrics.LoadAverage.WithLabelValues("5m").Set(1.1)
	mc.systemMetrics.LoadAverage.WithLabelValues("15m").Set(1.0)

	mc.systemMetrics.MemoryUsage.Set(4 * 1024 * 1024 * 1024)      // 4GB
	mc.systemMetrics.MemoryTotal.Set(16 * 1024 * 1024 * 1024)     // 16GB
	mc.systemMetrics.MemoryAvailable.Set(12 * 1024 * 1024 * 1024) // 12GB

	mc.systemMetrics.ProcessCount.Set(150)
	mc.systemMetrics.FileDescriptors.Set(1024)
}

// collectApplicationMetrics collects application-level metrics
func (mc *MetricsCollector) collectApplicationMetrics() {
	// TODO: Implement actual application metrics collection
	// This would collect metrics from the application components

	// Placeholder values for demonstration
	mc.appMetrics.ModelsLoaded.Set(3)
	mc.appMetrics.NodesActive.Set(5)
	mc.appMetrics.NodesTotal.Set(7)
}

// collectP2PMetrics collects P2P-specific metrics
func (mc *MetricsCollector) collectP2PMetrics() {
	// TODO: Implement actual P2P metrics collection
	// This would collect metrics from the P2P subsystem

	// Placeholder values for demonstration
	mc.p2pMetrics.PeerConnections.Set(4)
	mc.p2pMetrics.BandwidthUsage.WithLabelValues("inbound").Set(1024 * 1024) // 1MB/s
	mc.p2pMetrics.BandwidthUsage.WithLabelValues("outbound").Set(512 * 1024) // 512KB/s
}

// healthHandler handles health check requests
func (mc *MetricsCollector) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// RecordHTTPRequest records an HTTP request metric
func (mc *MetricsCollector) RecordHTTPRequest(method, endpoint, status string, duration time.Duration) {
	if mc.appMetrics != nil {
		mc.appMetrics.HTTPRequests.WithLabelValues(method, endpoint, status).Inc()
		mc.appMetrics.HTTPDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	}
}

// RecordHTTPError records an HTTP error metric
func (mc *MetricsCollector) RecordHTTPError(method, endpoint, errorType string) {
	if mc.appMetrics != nil {
		mc.appMetrics.HTTPErrors.WithLabelValues(method, endpoint, errorType).Inc()
	}
}

// RecordModelRequest records a model request metric
func (mc *MetricsCollector) RecordModelRequest(model, operation string, duration time.Duration) {
	if mc.appMetrics != nil {
		mc.appMetrics.ModelRequests.WithLabelValues(model, operation).Inc()
		mc.appMetrics.ModelLatency.WithLabelValues(model, operation).Observe(duration.Seconds())
	}
}

// RecordModelError records a model error metric
func (mc *MetricsCollector) RecordModelError(model, errorType string) {
	if mc.appMetrics != nil {
		mc.appMetrics.ModelErrors.WithLabelValues(model, errorType).Inc()
	}
}

// RecordCacheHit records a cache hit metric
func (mc *MetricsCollector) RecordCacheHit(cacheName string) {
	if mc.appMetrics != nil {
		mc.appMetrics.CacheHits.WithLabelValues(cacheName).Inc()
	}
}

// RecordCacheMiss records a cache miss metric
func (mc *MetricsCollector) RecordCacheMiss(cacheName string) {
	if mc.appMetrics != nil {
		mc.appMetrics.CacheMisses.WithLabelValues(cacheName).Inc()
	}
}

// UpdateCacheSize updates cache size metric
func (mc *MetricsCollector) UpdateCacheSize(cacheName string, size float64) {
	if mc.appMetrics != nil {
		mc.appMetrics.CacheSize.WithLabelValues(cacheName).Set(size)
	}
}

// RecordP2PMessage records a P2P message metric
func (mc *MetricsCollector) RecordP2PMessage(messageType, peerID, direction string, duration time.Duration) {
	if mc.p2pMetrics != nil {
		if direction == "received" {
			mc.p2pMetrics.MessagesReceived.WithLabelValues(messageType, peerID).Inc()
		} else {
			mc.p2pMetrics.MessagesSent.WithLabelValues(messageType, peerID).Inc()
		}
		mc.p2pMetrics.MessageLatency.WithLabelValues(messageType).Observe(duration.Seconds())
	}
}

// RecordConsensusRound records a consensus round metric
func (mc *MetricsCollector) RecordConsensusRound(duration time.Duration) {
	if mc.p2pMetrics != nil {
		mc.p2pMetrics.ConsensusRounds.Inc()
		mc.p2pMetrics.ConsensusLatency.Observe(duration.Seconds())
	}
}

// RecordLeaderChange records a leader change metric
func (mc *MetricsCollector) RecordLeaderChange() {
	if mc.p2pMetrics != nil {
		mc.p2pMetrics.LeaderChanges.Inc()
	}
}

// initApplicationMetrics initializes application metrics
func (mc *MetricsCollector) initApplicationMetrics() *ApplicationMetrics {
	metrics := &ApplicationMetrics{
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "endpoint", "status"}),
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "HTTP request duration",
		}, []string{"method", "endpoint"}),
		HTTPErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		}, []string{"method", "endpoint", "error_type"}),
		ModelsLoaded: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "models_loaded_total",
			Help: "Number of models currently loaded",
		}),
		ModelRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "model_requests_total",
			Help: "Total number of model requests",
		}, []string{"model", "operation"}),
		ModelLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "model_request_duration_seconds",
			Help: "Model request duration",
		}, []string{"model", "operation"}),
		ModelErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "model_errors_total",
			Help: "Total number of model errors",
		}, []string{"model", "error_type"}),
		NodesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nodes_active_total",
			Help: "Number of active nodes",
		}),
		NodesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nodes_total",
			Help: "Total number of nodes",
		}),
		NodeHealth: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "node_health_status",
			Help: "Node health status (1=healthy, 0=unhealthy)",
		}, []string{"node_id"}),
		CacheHits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		}, []string{"cache_name"}),
		CacheMisses: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		}, []string{"cache_name"}),
		CacheSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cache_size_bytes",
			Help: "Cache size in bytes",
		}, []string{"cache_name"}),
		QueueSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "queue_size_total",
			Help: "Queue size",
		}, []string{"queue_name"}),
		QueueLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "queue_latency_seconds",
			Help: "Queue processing latency",
		}, []string{"queue_name"}),
	}

	// Register metrics
	mc.registry.MustRegister(
		metrics.HTTPRequests,
		metrics.HTTPDuration,
		metrics.HTTPErrors,
		metrics.ModelsLoaded,
		metrics.ModelRequests,
		metrics.ModelLatency,
		metrics.ModelErrors,
		metrics.NodesActive,
		metrics.NodesTotal,
		metrics.NodeHealth,
		metrics.CacheHits,
		metrics.CacheMisses,
		metrics.CacheSize,
		metrics.QueueSize,
		metrics.QueueLatency,
	)

	return metrics
}
