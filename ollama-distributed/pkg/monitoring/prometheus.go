package monitoring

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics holds all Prometheus metrics for the fault tolerance system
type PrometheusMetrics struct {
	// Node and cluster metrics
	NodeStatus           *prometheus.GaugeVec
	ClusterSize          prometheus.Gauge
	LeaderElectionCount  prometheus.Counter
	ConsensusLatency     prometheus.Histogram

	// Fault tolerance metrics
	HealingAttempts      *prometheus.CounterVec
	HealingSuccessRate   prometheus.Gauge
	PredictionAccuracy   prometheus.Gauge
	RecoveryTime         prometheus.Histogram
	NodeFailures         *prometheus.CounterVec

	// Performance metrics
	RequestDuration      prometheus.Histogram
	RequestTotal         *prometheus.CounterVec
	ErrorRate            prometheus.Gauge
	Throughput           prometheus.Gauge

	// Resource metrics
	CPUUsage             *prometheus.GaugeVec
	MemoryUsage          *prometheus.GaugeVec
	DiskUsage            *prometheus.GaugeVec
	NetworkIO            *prometheus.CounterVec

	// Configuration metrics
	ConfigReloads        *prometheus.CounterVec
	ConfigValidation     *prometheus.CounterVec
	HotReloadLatency     prometheus.Histogram

	// Predictive detection metrics
	PredictionLatency    prometheus.Histogram
	PredictionConfidence prometheus.Histogram
	FalsePositives       prometheus.Counter
	FalseNegatives       prometheus.Counter

	// Self-healing metrics
	HealingLatency       prometheus.Histogram
	HealingStrategies    *prometheus.CounterVec
	LearningAccuracy     prometheus.Gauge
	ProactiveActions     prometheus.Counter

	// Redundancy metrics
	ReplicationFactor    *prometheus.GaugeVec
	ReplicaHealth        *prometheus.GaugeVec
	LoadDistribution     *prometheus.GaugeVec
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	metrics := &PrometheusMetrics{
		// Node and cluster metrics
		NodeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_node_status",
				Help: "Status of cluster nodes (0=down, 1=up)",
			},
			[]string{"node_id", "node_name", "region", "zone"},
		),
		ClusterSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_cluster_size",
				Help: "Number of active nodes in the cluster",
			},
		),
		LeaderElectionCount: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ollama_leader_election_total",
				Help: "Total number of leader elections",
			},
		),
		ConsensusLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_consensus_latency_seconds",
				Help:    "Latency of consensus operations",
				Buckets: prometheus.DefBuckets,
			},
		),

		// Fault tolerance metrics
		HealingAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_fault_tolerance_healing_attempts_total",
				Help: "Total number of healing attempts",
			},
			[]string{"node_id", "strategy", "result"},
		),
		HealingSuccessRate: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_fault_tolerance_healing_success_rate",
				Help: "Success rate of healing attempts (0.0-1.0)",
			},
		),
		PredictionAccuracy: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_fault_tolerance_prediction_accuracy",
				Help: "Accuracy of fault predictions (0.0-1.0)",
			},
		),
		RecoveryTime: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_fault_tolerance_recovery_time_seconds",
				Help:    "Time taken to recover from failures",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
		),
		NodeFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_fault_tolerance_node_failures_total",
				Help: "Total number of node failures detected",
			},
			[]string{"node_id", "failure_type", "severity"},
		),

		// Performance metrics
		RequestDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_request_duration_seconds",
				Help:    "Duration of inference requests",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5, 10},
			},
		),
		RequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_request_total",
				Help: "Total number of requests processed",
			},
			[]string{"method", "status", "node_id"},
		),
		ErrorRate: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_error_rate",
				Help: "Current error rate (0.0-1.0)",
			},
		),
		Throughput: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_throughput_requests_per_second",
				Help: "Current throughput in requests per second",
			},
		),

		// Resource metrics
		CPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_cpu_usage_percent",
				Help: "CPU usage percentage by component",
			},
			[]string{"node_id", "component"},
		),
		MemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_memory_usage_bytes",
				Help: "Memory usage in bytes by component",
			},
			[]string{"node_id", "component"},
		),
		DiskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_disk_usage_bytes",
				Help: "Disk usage in bytes by component",
			},
			[]string{"node_id", "component", "mount_point"},
		),
		NetworkIO: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_network_io_bytes_total",
				Help: "Total network I/O in bytes",
			},
			[]string{"node_id", "direction", "interface"},
		),

		// Configuration metrics
		ConfigReloads: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_config_reloads_total",
				Help: "Total number of configuration reloads",
			},
			[]string{"result", "component"},
		),
		ConfigValidation: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_config_validation_total",
				Help: "Total number of configuration validations",
			},
			[]string{"result", "validation_type"},
		),
		HotReloadLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_config_hot_reload_latency_seconds",
				Help:    "Latency of hot configuration reloads",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2, 5},
			},
		),

		// Predictive detection metrics
		PredictionLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_prediction_latency_seconds",
				Help:    "Latency of fault predictions",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 2},
			},
		),
		PredictionConfidence: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_prediction_confidence",
				Help:    "Confidence scores of fault predictions",
				Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
		),
		FalsePositives: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ollama_prediction_false_positives_total",
				Help: "Total number of false positive predictions",
			},
		),
		FalseNegatives: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ollama_prediction_false_negatives_total",
				Help: "Total number of false negative predictions",
			},
		),

		// Self-healing metrics
		HealingLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_healing_latency_seconds",
				Help:    "Latency of healing operations",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
		),
		HealingStrategies: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_healing_strategies_total",
				Help: "Total number of healing strategies executed",
			},
			[]string{"strategy", "result"},
		),
		LearningAccuracy: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_learning_accuracy",
				Help: "Accuracy of learning-based adaptations (0.0-1.0)",
			},
		),
		ProactiveActions: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "ollama_proactive_actions_total",
				Help: "Total number of proactive healing actions",
			},
		),

		// Redundancy metrics
		ReplicationFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_replication_factor",
				Help: "Current replication factor by component",
			},
			[]string{"component", "service"},
		),
		ReplicaHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_replica_health",
				Help: "Health status of replicas (0=unhealthy, 1=healthy)",
			},
			[]string{"component", "replica_id"},
		),
		LoadDistribution: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_load_distribution",
				Help: "Load distribution across nodes (0.0-1.0)",
			},
			[]string{"node_id", "metric_type"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		metrics.NodeStatus,
		metrics.ClusterSize,
		metrics.LeaderElectionCount,
		metrics.ConsensusLatency,
		metrics.HealingAttempts,
		metrics.HealingSuccessRate,
		metrics.PredictionAccuracy,
		metrics.RecoveryTime,
		metrics.NodeFailures,
		metrics.RequestDuration,
		metrics.RequestTotal,
		metrics.ErrorRate,
		metrics.Throughput,
		metrics.CPUUsage,
		metrics.MemoryUsage,
		metrics.DiskUsage,
		metrics.NetworkIO,
		metrics.ConfigReloads,
		metrics.ConfigValidation,
		metrics.HotReloadLatency,
		metrics.PredictionLatency,
		metrics.PredictionConfidence,
		metrics.FalsePositives,
		metrics.FalseNegatives,
		metrics.HealingLatency,
		metrics.HealingStrategies,
		metrics.LearningAccuracy,
		metrics.ProactiveActions,
		metrics.ReplicationFactor,
		metrics.ReplicaHealth,
		metrics.LoadDistribution,
	)

	return metrics
}

// MetricsServer provides HTTP endpoint for Prometheus metrics
type MetricsServer struct {
	server  *http.Server
	metrics *PrometheusMetrics
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port string, metrics *PrometheusMetrics) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return &MetricsServer{
		server:  server,
		metrics: metrics,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start() error {
	return ms.server.ListenAndServe()
}

// Stop stops the metrics server
func (ms *MetricsServer) Stop(ctx context.Context) error {
	return ms.server.Shutdown(ctx)
}

// UpdateNodeStatus updates the status of a node
func (pm *PrometheusMetrics) UpdateNodeStatus(nodeID, nodeName, region, zone string, status float64) {
	pm.NodeStatus.WithLabelValues(nodeID, nodeName, region, zone).Set(status)
}

// UpdateClusterSize updates the cluster size
func (pm *PrometheusMetrics) UpdateClusterSize(size float64) {
	pm.ClusterSize.Set(size)
}

// RecordLeaderElection records a leader election event
func (pm *PrometheusMetrics) RecordLeaderElection() {
	pm.LeaderElectionCount.Inc()
}

// RecordConsensusLatency records consensus operation latency
func (pm *PrometheusMetrics) RecordConsensusLatency(duration time.Duration) {
	pm.ConsensusLatency.Observe(duration.Seconds())
}

// RecordHealingAttempt records a healing attempt
func (pm *PrometheusMetrics) RecordHealingAttempt(nodeID, strategy, result string) {
	pm.HealingAttempts.WithLabelValues(nodeID, strategy, result).Inc()
}

// UpdateHealingSuccessRate updates the healing success rate
func (pm *PrometheusMetrics) UpdateHealingSuccessRate(rate float64) {
	pm.HealingSuccessRate.Set(rate)
}

// UpdatePredictionAccuracy updates the prediction accuracy
func (pm *PrometheusMetrics) UpdatePredictionAccuracy(accuracy float64) {
	pm.PredictionAccuracy.Set(accuracy)
}

// RecordRecoveryTime records recovery time
func (pm *PrometheusMetrics) RecordRecoveryTime(duration time.Duration) {
	pm.RecoveryTime.Observe(duration.Seconds())
}

// RecordNodeFailure records a node failure
func (pm *PrometheusMetrics) RecordNodeFailure(nodeID, failureType, severity string) {
	pm.NodeFailures.WithLabelValues(nodeID, failureType, severity).Inc()
}

// RecordRequest records a request
func (pm *PrometheusMetrics) RecordRequest(method, status, nodeID string, duration time.Duration) {
	pm.RequestDuration.Observe(duration.Seconds())
	pm.RequestTotal.WithLabelValues(method, status, nodeID).Inc()
}

// UpdateErrorRate updates the error rate
func (pm *PrometheusMetrics) UpdateErrorRate(rate float64) {
	pm.ErrorRate.Set(rate)
}

// UpdateThroughput updates the throughput
func (pm *PrometheusMetrics) UpdateThroughput(rps float64) {
	pm.Throughput.Set(rps)
}
