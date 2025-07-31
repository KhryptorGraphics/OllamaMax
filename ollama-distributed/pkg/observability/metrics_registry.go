package observability

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// MetricsRegistry provides a centralized registry for all system metrics
type MetricsRegistry struct {
	collector          *MetricsCollector
	prometheusExporter *PrometheusExporter

	// Standard metrics for all components
	schedulerMetrics      *SchedulerMetrics
	consensusMetrics      *ConsensusMetrics
	p2pMetrics            *P2PMetrics
	apiMetrics            *APIMetrics
	faultToleranceMetrics *FaultToleranceMetrics
	modelMetrics          *ModelMetrics

	mu sync.RWMutex
}

// SchedulerMetrics contains all scheduler-related metrics
type SchedulerMetrics struct {
	TasksTotal           *prometheus.CounterVec
	TasksActive          *prometheus.GaugeVec
	TaskDuration         *prometheus.HistogramVec
	TaskErrors           *prometheus.CounterVec
	LoadBalancerRequests *prometheus.CounterVec
	NodeUtilization      *prometheus.GaugeVec
}

// ConsensusMetrics contains all consensus-related metrics
type ConsensusMetrics struct {
	LeaderElections *prometheus.CounterVec
	LogEntries      *prometheus.CounterVec
	CommitLatency   *prometheus.HistogramVec
	QuorumStatus    *prometheus.GaugeVec
	NodeStates      *prometheus.GaugeVec
	ConsensusErrors *prometheus.CounterVec
}

// P2PMetrics contains all P2P network metrics
type P2PMetrics struct {
	ConnectionsTotal  *prometheus.CounterVec
	ConnectionsActive *prometheus.GaugeVec
	MessagesSent      *prometheus.CounterVec
	MessagesReceived  *prometheus.CounterVec
	NetworkLatency    *prometheus.HistogramVec
	BandwidthUsage    *prometheus.GaugeVec
	PeerDiscovery     *prometheus.CounterVec
}

// APIMetrics contains all API gateway metrics
type APIMetrics struct {
	RequestsTotal        *prometheus.CounterVec
	RequestDuration      *prometheus.HistogramVec
	ResponseSize         *prometheus.HistogramVec
	ActiveConnections    *prometheus.GaugeVec
	WebSocketConnections *prometheus.GaugeVec
	RateLimitHits        *prometheus.CounterVec
}

// FaultToleranceMetrics contains all fault tolerance metrics
type FaultToleranceMetrics struct {
	FaultsDetected     *prometheus.CounterVec
	RecoveryAttempts   *prometheus.CounterVec
	RecoverySuccess    *prometheus.CounterVec
	PredictionAccuracy *prometheus.GaugeVec
	HealingOperations  *prometheus.CounterVec
	SystemHealth       *prometheus.GaugeVec
}

// ModelMetrics contains all model management metrics
type ModelMetrics struct {
	ModelsLoaded          *prometheus.GaugeVec
	ModelRequests         *prometheus.CounterVec
	ModelLatency          *prometheus.HistogramVec
	ModelErrors           *prometheus.CounterVec
	ReplicationOperations *prometheus.CounterVec
	StorageUsage          *prometheus.GaugeVec
}

// NewMetricsRegistry creates a new centralized metrics registry
func NewMetricsRegistry(config *MetricsConfig) *MetricsRegistry {
	if config == nil {
		config = &MetricsConfig{
			Namespace:          "ollama",
			Subsystem:          "distributed",
			CollectionInterval: 15 * time.Second,
			EnablePrometheus:   true,
			PrometheusConfig:   DefaultPrometheusConfig(),
		}
	}

	collector := NewMetricsCollector(config)

	registry := &MetricsRegistry{
		collector:          collector,
		prometheusExporter: collector.prometheusExporter,
	}

	// Initialize all component metrics
	registry.initializeSchedulerMetrics()
	registry.initializeConsensusMetrics()
	registry.initializeP2PMetrics()
	registry.initializeAPIMetrics()
	registry.initializeFaultToleranceMetrics()
	registry.initializeModelMetrics()

	return registry
}

// Start starts the metrics registry and all exporters
func (mr *MetricsRegistry) Start() error {
	if err := mr.collector.Start(); err != nil {
		return err
	}

	log.Info().Msg("Metrics registry started with Prometheus integration")
	return nil
}

// Stop stops the metrics registry
func (mr *MetricsRegistry) Stop() error {
	return mr.collector.Close()
}

// GetSchedulerMetrics returns scheduler metrics
func (mr *MetricsRegistry) GetSchedulerMetrics() *SchedulerMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.schedulerMetrics
}

// GetConsensusMetrics returns consensus metrics
func (mr *MetricsRegistry) GetConsensusMetrics() *ConsensusMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.consensusMetrics
}

// GetP2PMetrics returns P2P metrics
func (mr *MetricsRegistry) GetP2PMetrics() *P2PMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.p2pMetrics
}

// GetAPIMetrics returns API metrics
func (mr *MetricsRegistry) GetAPIMetrics() *APIMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.apiMetrics
}

// GetFaultToleranceMetrics returns fault tolerance metrics
func (mr *MetricsRegistry) GetFaultToleranceMetrics() *FaultToleranceMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.faultToleranceMetrics
}

// GetModelMetrics returns model metrics
func (mr *MetricsRegistry) GetModelMetrics() *ModelMetrics {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.modelMetrics
}

// GetPrometheusExporter returns the Prometheus exporter
func (mr *MetricsRegistry) GetPrometheusExporter() *PrometheusExporter {
	return mr.prometheusExporter
}

// initializeSchedulerMetrics initializes scheduler metrics
func (mr *MetricsRegistry) initializeSchedulerMetrics() {
	mr.schedulerMetrics = &SchedulerMetrics{
		TasksTotal: mr.prometheusExporter.RegisterCounter(
			"scheduler_tasks_total",
			"Total number of tasks processed by the scheduler",
			[]string{"status", "node_id", "task_type"},
		),
		TasksActive: mr.prometheusExporter.RegisterGauge(
			"scheduler_tasks_active",
			"Number of currently active tasks",
			[]string{"node_id", "task_type"},
		),
		TaskDuration: mr.prometheusExporter.RegisterHistogram(
			"scheduler_task_duration_seconds",
			"Duration of task execution in seconds",
			[]string{"task_type", "node_id"},
			[]float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		),
		TaskErrors: mr.prometheusExporter.RegisterCounter(
			"scheduler_task_errors_total",
			"Total number of task errors",
			[]string{"error_type", "node_id", "task_type"},
		),
		LoadBalancerRequests: mr.prometheusExporter.RegisterCounter(
			"scheduler_load_balancer_requests_total",
			"Total number of load balancer requests",
			[]string{"strategy", "node_id"},
		),
		NodeUtilization: mr.prometheusExporter.RegisterGauge(
			"scheduler_node_utilization",
			"Current utilization of cluster nodes",
			[]string{"node_id", "resource_type"},
		),
	}
}

// initializeConsensusMetrics initializes consensus metrics
func (mr *MetricsRegistry) initializeConsensusMetrics() {
	mr.consensusMetrics = &ConsensusMetrics{
		LeaderElections: mr.prometheusExporter.RegisterCounter(
			"consensus_leader_elections_total",
			"Total number of leader elections",
			[]string{"node_id", "result"},
		),
		LogEntries: mr.prometheusExporter.RegisterCounter(
			"consensus_log_entries_total",
			"Total number of log entries",
			[]string{"node_id", "entry_type"},
		),
		CommitLatency: mr.prometheusExporter.RegisterHistogram(
			"consensus_commit_latency_seconds",
			"Latency of consensus commits in seconds",
			[]string{"node_id"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		),
		QuorumStatus: mr.prometheusExporter.RegisterGauge(
			"consensus_quorum_status",
			"Current quorum status (1 = has quorum, 0 = no quorum)",
			[]string{"cluster_id"},
		),
		NodeStates: mr.prometheusExporter.RegisterGauge(
			"consensus_node_states",
			"Current state of consensus nodes (0=follower, 1=candidate, 2=leader)",
			[]string{"node_id"},
		),
		ConsensusErrors: mr.prometheusExporter.RegisterCounter(
			"consensus_errors_total",
			"Total number of consensus errors",
			[]string{"error_type", "node_id"},
		),
	}
}

// initializeP2PMetrics initializes P2P network metrics
func (mr *MetricsRegistry) initializeP2PMetrics() {
	mr.p2pMetrics = &P2PMetrics{
		ConnectionsTotal: mr.prometheusExporter.RegisterCounter(
			"p2p_connections_total",
			"Total number of P2P connections",
			[]string{"direction", "peer_id"},
		),
		ConnectionsActive: mr.prometheusExporter.RegisterGauge(
			"p2p_connections_active",
			"Number of active P2P connections",
			[]string{"protocol"},
		),
		MessagesSent: mr.prometheusExporter.RegisterCounter(
			"p2p_messages_sent_total",
			"Total number of P2P messages sent",
			[]string{"message_type", "peer_id"},
		),
		MessagesReceived: mr.prometheusExporter.RegisterCounter(
			"p2p_messages_received_total",
			"Total number of P2P messages received",
			[]string{"message_type", "peer_id"},
		),
		NetworkLatency: mr.prometheusExporter.RegisterHistogram(
			"p2p_network_latency_seconds",
			"P2P network latency in seconds",
			[]string{"peer_id"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		),
		BandwidthUsage: mr.prometheusExporter.RegisterGauge(
			"p2p_bandwidth_usage_bytes_per_second",
			"P2P bandwidth usage in bytes per second",
			[]string{"direction", "peer_id"},
		),
		PeerDiscovery: mr.prometheusExporter.RegisterCounter(
			"p2p_peer_discovery_total",
			"Total number of peer discovery events",
			[]string{"discovery_type", "result"},
		),
	}
}

// initializeAPIMetrics initializes API gateway metrics
func (mr *MetricsRegistry) initializeAPIMetrics() {
	mr.apiMetrics = &APIMetrics{
		RequestsTotal: mr.prometheusExporter.RegisterCounter(
			"api_requests_total",
			"Total number of API requests",
			[]string{"method", "endpoint", "status_code"},
		),
		RequestDuration: mr.prometheusExporter.RegisterHistogram(
			"api_request_duration_seconds",
			"Duration of API requests in seconds",
			[]string{"method", "endpoint"},
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		),
		ResponseSize: mr.prometheusExporter.RegisterHistogram(
			"api_response_size_bytes",
			"Size of API responses in bytes",
			[]string{"method", "endpoint"},
			[]float64{100, 1000, 10000, 100000, 1000000, 10000000},
		),
		ActiveConnections: mr.prometheusExporter.RegisterGauge(
			"api_active_connections",
			"Number of active API connections",
			[]string{"connection_type"},
		),
		WebSocketConnections: mr.prometheusExporter.RegisterGauge(
			"api_websocket_connections",
			"Number of active WebSocket connections",
			[]string{"endpoint"},
		),
		RateLimitHits: mr.prometheusExporter.RegisterCounter(
			"api_rate_limit_hits_total",
			"Total number of rate limit hits",
			[]string{"endpoint", "client_id"},
		),
	}
}

// initializeFaultToleranceMetrics initializes fault tolerance metrics
func (mr *MetricsRegistry) initializeFaultToleranceMetrics() {
	mr.faultToleranceMetrics = &FaultToleranceMetrics{
		FaultsDetected: mr.prometheusExporter.RegisterCounter(
			"fault_tolerance_faults_detected_total",
			"Total number of faults detected",
			[]string{"fault_type", "component", "severity"},
		),
		RecoveryAttempts: mr.prometheusExporter.RegisterCounter(
			"fault_tolerance_recovery_attempts_total",
			"Total number of recovery attempts",
			[]string{"recovery_type", "component"},
		),
		RecoverySuccess: mr.prometheusExporter.RegisterCounter(
			"fault_tolerance_recovery_success_total",
			"Total number of successful recoveries",
			[]string{"recovery_type", "component"},
		),
		PredictionAccuracy: mr.prometheusExporter.RegisterGauge(
			"fault_tolerance_prediction_accuracy",
			"Accuracy of fault prediction models",
			[]string{"model_type", "component"},
		),
		HealingOperations: mr.prometheusExporter.RegisterCounter(
			"fault_tolerance_healing_operations_total",
			"Total number of self-healing operations",
			[]string{"healing_type", "component"},
		),
		SystemHealth: mr.prometheusExporter.RegisterGauge(
			"fault_tolerance_system_health",
			"Overall system health score (0-1)",
			[]string{"component", "subsystem"},
		),
	}
}

// initializeModelMetrics initializes model management metrics
func (mr *MetricsRegistry) initializeModelMetrics() {
	mr.modelMetrics = &ModelMetrics{
		ModelsLoaded: mr.prometheusExporter.RegisterGauge(
			"model_loaded_count",
			"Number of currently loaded models",
			[]string{"model_name", "node_id"},
		),
		ModelRequests: mr.prometheusExporter.RegisterCounter(
			"model_requests_total",
			"Total number of model inference requests",
			[]string{"model_name", "node_id", "status"},
		),
		ModelLatency: mr.prometheusExporter.RegisterHistogram(
			"model_inference_latency_seconds",
			"Model inference latency in seconds",
			[]string{"model_name", "node_id"},
			[]float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
		),
		ModelErrors: mr.prometheusExporter.RegisterCounter(
			"model_errors_total",
			"Total number of model errors",
			[]string{"model_name", "error_type", "node_id"},
		),
		ReplicationOperations: mr.prometheusExporter.RegisterCounter(
			"model_replication_operations_total",
			"Total number of model replication operations",
			[]string{"operation_type", "model_name", "status"},
		),
		StorageUsage: mr.prometheusExporter.RegisterGauge(
			"model_storage_usage_bytes",
			"Model storage usage in bytes",
			[]string{"model_name", "node_id", "storage_type"},
		),
	}
}
