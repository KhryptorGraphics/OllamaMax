package fault_tolerance

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// FaultToleranceManager manages fault tolerance and recovery mechanisms
type FaultToleranceManager struct {
	config           *Config
	detectionSystem  *FaultDetector
	recoveryEngine   *RecoveryEngine
	replicationMgr   *ReplicationManager
	circuitBreaker   *CircuitBreaker
	checkpointing    *CheckpointManager
	metrics          *FaultToleranceMetrics
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	started          bool
}

// Config holds fault tolerance configuration
type Config struct {
	ReplicationFactor     int           `json:"replication_factor"`
	HealthCheckInterval   time.Duration `json:"health_check_interval"`
	RecoveryTimeout       time.Duration `json:"recovery_timeout"`
	CircuitBreakerEnabled bool          `json:"circuit_breaker_enabled"`
	CheckpointInterval    time.Duration `json:"checkpoint_interval"`
	MaxRetries            int           `json:"max_retries"`
	RetryBackoff          time.Duration `json:"retry_backoff"`
}

// FaultDetector monitors system health and detects faults
type FaultDetector struct {
	manager        *FaultToleranceManager
	healthCheckers map[string]HealthChecker
	monitors       []SystemMonitor
	alerting       *AlertingSystem
	thresholds     map[string]float64
	detections     map[string]*FaultDetection
	detectionsMu   sync.RWMutex
}

// HealthChecker interface for different health checking mechanisms
type HealthChecker interface {
	Check(ctx context.Context, target string) (*HealthResult, error)
	GetName() string
}

// SystemMonitor interface for system monitoring
type SystemMonitor interface {
	Monitor(ctx context.Context) (*MonitorResult, error)
	GetName() string
}

// AlertingSystem manages fault alerts
type AlertingSystem struct {
	alerts    []*FaultAlert
	alertsMu  sync.RWMutex
	handlers  map[string]AlertHandler
	config    *AlertConfig
}

// AlertHandler interface for handling alerts
type AlertHandler interface {
	Handle(alert *FaultAlert) error
	GetName() string
}

// AlertConfig holds alerting configuration
type AlertConfig struct {
	Enabled       bool          `json:"enabled"`
	Channels      []string      `json:"channels"`
	ThrottleTime  time.Duration `json:"throttle_time"`
	SeverityLevel string        `json:"severity_level"`
}

// FaultDetection represents a detected fault
type FaultDetection struct {
	ID          string                 `json:"id"`
	Type        FaultType              `json:"type"`
	Severity    FaultSeverity          `json:"severity"`
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	DetectedAt  time.Time              `json:"detected_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Status      FaultStatus            `json:"status"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// FaultType represents the type of fault
type FaultType string

const (
	FaultTypeNodeFailure     FaultType = "node_failure"
	FaultTypeNetworkPartition FaultType = "network_partition"
	FaultTypeResourceExhaustion FaultType = "resource_exhaustion"
	FaultTypePerformanceAnomaly FaultType = "performance_anomaly"
	FaultTypeServiceUnavailable FaultType = "service_unavailable"
)

// FaultSeverity represents the severity of a fault
type FaultSeverity string

const (
	FaultSeverityLow      FaultSeverity = "low"
	FaultSeverityMedium   FaultSeverity = "medium"
	FaultSeverityHigh     FaultSeverity = "high"
	FaultSeverityCritical FaultSeverity = "critical"
)

// FaultStatus represents the status of a fault
type FaultStatus string

const (
	FaultStatusDetected   FaultStatus = "detected"
	FaultStatusRecovering FaultStatus = "recovering"
	FaultStatusResolved   FaultStatus = "resolved"
	FaultStatusPersistent FaultStatus = "persistent"
)

// HealthResult represents a health check result
type HealthResult struct {
	Target    string                 `json:"target"`
	Healthy   bool                   `json:"healthy"`
	Latency   time.Duration          `json:"latency"`
	Error     string                 `json:"error,omitempty"`
	Metrics   map[string]interface{} `json:"metrics"`
	Timestamp time.Time              `json:"timestamp"`
}

// MonitorResult represents a system monitoring result
type MonitorResult struct {
	System    string                 `json:"system"`
	Healthy   bool                   `json:"healthy"`
	Metrics   map[string]interface{} `json:"metrics"`
	Anomalies []string               `json:"anomalies"`
	Timestamp time.Time              `json:"timestamp"`
}

// FaultAlert represents a fault alert
type FaultAlert struct {
	ID          string                 `json:"id"`
	FaultID     string                 `json:"fault_id"`
	Severity    FaultSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Handled     bool                   `json:"handled"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecoveryEngine handles fault recovery
type RecoveryEngine struct {
	manager         *FaultToleranceManager
	strategies      map[FaultType][]RecoveryStrategy
	recoveryQueue   chan *RecoveryRequest
	recoveryHistory []*RecoveryAttempt
	historyMu       sync.RWMutex
}

// RecoveryStrategy interface for different recovery strategies
type RecoveryStrategy interface {
	Recover(ctx context.Context, fault *FaultDetection) (*RecoveryResult, error)
	GetName() string
	CanHandle(fault *FaultDetection) bool
}

// RecoveryRequest represents a recovery request
type RecoveryRequest struct {
	Fault     *FaultDetection
	Strategy  string
	Priority  int
	Timestamp time.Time
}

// RecoveryResult represents the result of a recovery attempt
type RecoveryResult struct {
	FaultID     string                 `json:"fault_id"`
	Strategy    string                 `json:"strategy"`
	Successful  bool                   `json:"successful"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RecoveryAttempt represents a recovery attempt
type RecoveryAttempt struct {
	ID        string          `json:"id"`
	FaultID   string          `json:"fault_id"`
	Strategy  string          `json:"strategy"`
	Result    *RecoveryResult `json:"result"`
	Timestamp time.Time       `json:"timestamp"`
}

// ReplicationManager handles model and data replication
type ReplicationManager struct {
	manager         *FaultToleranceManager
	replicationJobs map[string]*ReplicationJob
	jobsMu          sync.RWMutex
	factor          int
	strategy        ReplicationStrategy
}

// ReplicationJob represents a replication job
type ReplicationJob struct {
	ID            string                 `json:"id"`
	Type          ReplicationType        `json:"type"`
	Source        string                 `json:"source"`
	Targets       []string               `json:"targets"`
	Status        ReplicationStatus      `json:"status"`
	Progress      float64                `json:"progress"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ReplicationType represents the type of replication
type ReplicationType string

const (
	ReplicationTypeModel ReplicationType = "model"
	ReplicationTypeData  ReplicationType = "data"
	ReplicationTypeState ReplicationType = "state"
)

// ReplicationStatus represents the status of replication
type ReplicationStatus string

const (
	ReplicationStatusPending    ReplicationStatus = "pending"
	ReplicationStatusInProgress ReplicationStatus = "in_progress"
	ReplicationStatusCompleted  ReplicationStatus = "completed"
	ReplicationStatusFailed     ReplicationStatus = "failed"
)

// ReplicationStrategy represents the replication strategy
type ReplicationStrategy string

const (
	ReplicationStrategyImmediate ReplicationStrategy = "immediate"
	ReplicationStrategyLazy      ReplicationStrategy = "lazy"
	ReplicationStrategyAdaptive  ReplicationStrategy = "adaptive"
)

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	manager         *FaultToleranceManager
	circuits        map[string]*Circuit
	circuitsMu      sync.RWMutex
	defaultConfig   *CircuitConfig
}

// Circuit represents a circuit breaker
type Circuit struct {
	Name          string              `json:"name"`
	State         CircuitState        `json:"state"`
	Config        *CircuitConfig      `json:"config"`
	FailureCount  int                 `json:"failure_count"`
	SuccessCount  int                 `json:"success_count"`
	LastFailure   time.Time           `json:"last_failure"`
	LastSuccess   time.Time           `json:"last_success"`
	StateChanged  time.Time           `json:"state_changed"`
	mu            sync.RWMutex
}

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

// CircuitConfig holds circuit breaker configuration
type CircuitConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

// CheckpointManager handles checkpointing and recovery
type CheckpointManager struct {
	manager     *FaultToleranceManager
	storage     CheckpointStorage
	frequency   time.Duration
	compression CompressionAlgorithm
	encryption  EncryptionMethod
	cleanup     CleanupPolicy
	checkpoints map[string]*Checkpoint
	checkpointsMu sync.RWMutex
}

// CheckpointStorage interface for checkpoint storage
type CheckpointStorage interface {
	Store(checkpoint *Checkpoint) error
	Load(id string) (*Checkpoint, error)
	List() ([]*Checkpoint, error)
	Delete(id string) error
}

// CompressionAlgorithm interface for compression
type CompressionAlgorithm interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	GetName() string
}

// EncryptionMethod interface for encryption
type EncryptionMethod interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	GetName() string
}

// CleanupPolicy interface for cleanup policies
type CleanupPolicy interface {
	ShouldCleanup(checkpoint *Checkpoint) bool
	GetName() string
}

// Checkpoint represents a system checkpoint
type Checkpoint struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	ModelState    ModelState             `json:"model_state"`
	RequestQueue  []Request              `json:"request_queue"`
	NodeStates    map[string]NodeState   `json:"node_states"`
	Metadata      map[string]interface{} `json:"metadata"`
	Size          int64                  `json:"size"`
	Compressed    bool                   `json:"compressed"`
	Encrypted     bool                   `json:"encrypted"`
}

// ModelState represents the state of a model
type ModelState struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	State       map[string]interface{} `json:"state"`
	Weights     []byte                 `json:"weights"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Request represents a request in the system
type Request struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// NodeState represents the state of a node
type NodeState struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Resources map[string]interface{} `json:"resources"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// FaultToleranceMetrics tracks fault tolerance metrics
type FaultToleranceMetrics struct {
	FaultsDetected      int64     `json:"faults_detected"`
	FaultsResolved      int64     `json:"faults_resolved"`
	RecoveryAttempts    int64     `json:"recovery_attempts"`
	SuccessfulRecoveries int64    `json:"successful_recoveries"`
	AverageRecoveryTime time.Duration `json:"average_recovery_time"`
	Uptime              time.Duration `json:"uptime"`
	LastFault           *time.Time    `json:"last_fault"`
	LastRecovery        *time.Time    `json:"last_recovery"`
}

// NewFaultToleranceManager creates a new fault tolerance manager
func NewFaultToleranceManager(config *Config) *FaultToleranceManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	ftm := &FaultToleranceManager{
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		metrics: &FaultToleranceMetrics{},
	}
	
	// Initialize components
	ftm.initializeComponents()
	
	return ftm
}

// initializeComponents initializes all fault tolerance components
func (ftm *FaultToleranceManager) initializeComponents() {
	// Initialize fault detector
	ftm.detectionSystem = &FaultDetector{
		manager:        ftm,
		healthCheckers: make(map[string]HealthChecker),
		monitors:       make([]SystemMonitor, 0),
		thresholds:     make(map[string]float64),
		detections:     make(map[string]*FaultDetection),
	}
	
	// Initialize alerting system
	ftm.detectionSystem.alerting = &AlertingSystem{
		alerts:   make([]*FaultAlert, 0),
		handlers: make(map[string]AlertHandler),
		config: &AlertConfig{
			Enabled:       true,
			Channels:      []string{"log", "email"},
			ThrottleTime:  5 * time.Minute,
			SeverityLevel: "medium",
		},
	}
	
	// Initialize recovery engine
	ftm.recoveryEngine = &RecoveryEngine{
		manager:         ftm,
		strategies:      make(map[FaultType][]RecoveryStrategy),
		recoveryQueue:   make(chan *RecoveryRequest, 100),
		recoveryHistory: make([]*RecoveryAttempt, 0),
	}
	
	// Initialize replication manager
	ftm.replicationMgr = &ReplicationManager{
		manager:         ftm,
		replicationJobs: make(map[string]*ReplicationJob),
		factor:          ftm.config.ReplicationFactor,
		strategy:        ReplicationStrategyAdaptive,
	}
	
	// Initialize circuit breaker
	ftm.circuitBreaker = &CircuitBreaker{
		manager:  ftm,
		circuits: make(map[string]*Circuit),
		defaultConfig: &CircuitConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
			ResetTimeout:     60 * time.Second,
		},
	}
	
	// Initialize checkpoint manager
	ftm.checkpointing = &CheckpointManager{
		manager:     ftm,
		frequency:   ftm.config.CheckpointInterval,
		checkpoints: make(map[string]*Checkpoint),
	}
	
	// Register default recovery strategies
	ftm.registerDefaultStrategies()
	
	// Register default health checkers
	ftm.registerDefaultHealthCheckers()
}

// registerDefaultStrategies registers default recovery strategies
func (ftm *FaultToleranceManager) registerDefaultStrategies() {
	// Node failure recovery strategies
	ftm.recoveryEngine.strategies[FaultTypeNodeFailure] = []RecoveryStrategy{
		&GracefulDegradationStrategy{},
		&RequestMigrationStrategy{},
		&ModelReplicationStrategy{},
	}
	
	// Network partition recovery strategies
	ftm.recoveryEngine.strategies[FaultTypeNetworkPartition] = []RecoveryStrategy{
		&PartitionToleranceStrategy{},
		&RequestMigrationStrategy{},
	}
	
	// Resource exhaustion recovery strategies
	ftm.recoveryEngine.strategies[FaultTypeResourceExhaustion] = []RecoveryStrategy{
		&ResourceScalingStrategy{},
		&LoadSheddingStrategy{},
	}
	
	// Performance anomaly recovery strategies
	ftm.recoveryEngine.strategies[FaultTypePerformanceAnomaly] = []RecoveryStrategy{
		&PerformanceTuningStrategy{},
		&LoadBalancingStrategy{},
	}
}

// registerDefaultHealthCheckers registers default health checkers
func (ftm *FaultToleranceManager) registerDefaultHealthCheckers() {
	ftm.detectionSystem.AddHealthChecker("node", NewNodeHealthChecker())
	ftm.detectionSystem.AddHealthChecker("network", NewNetworkHealthChecker())
	ftm.detectionSystem.AddHealthChecker("resource", NewResourceHealthChecker())
	ftm.detectionSystem.AddHealthChecker("performance", NewPerformanceHealthChecker())
}

// Start starts the fault tolerance manager
func (ftm *FaultToleranceManager) Start() error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()
	
	if ftm.started {
		return fmt.Errorf("fault tolerance manager already started")
	}
	
	// Start fault detection
	go ftm.detectionSystem.Start(ftm.ctx)
	
	// Start recovery engine
	go ftm.recoveryEngine.Start(ftm.ctx)
	
	// Start checkpointing
	go ftm.checkpointing.Start(ftm.ctx)
	
	ftm.started = true
	
	slog.Info("fault tolerance manager started",
		"replication_factor", ftm.config.ReplicationFactor,
		"health_check_interval", ftm.config.HealthCheckInterval,
		"circuit_breaker_enabled", ftm.config.CircuitBreakerEnabled)
	
	return nil
}

// AddHealthChecker adds a health checker to the fault detector
func (fd *FaultDetector) AddHealthChecker(name string, checker HealthChecker) {
	fd.healthCheckers[name] = checker
}

// Start method for FaultDetector
func (fd *FaultDetector) Start(ctx context.Context) error {
	// Implementation for starting fault detector
	slog.Info("fault detector started")
	return nil
}

// Start method for RecoveryEngine
func (re *RecoveryEngine) Start(ctx context.Context) error {
	// Implementation for starting recovery engine
	slog.Info("recovery engine started")
	return nil
}

// CreateCheckpoint creates a new checkpoint
func (cm *CheckpointManager) CreateCheckpoint() *Checkpoint {
	checkpoint := &Checkpoint{
		ID:           fmt.Sprintf("checkpoint_%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
		ModelState:   ModelState{},
		RequestQueue: []Request{},
		NodeStates:   make(map[string]NodeState),
		Metadata:     make(map[string]interface{}),
		Size:         0,
		Compressed:   false,
		Encrypted:    false,
	}
	
	// Store system metadata (placeholder)
	checkpoint.Metadata["system_health"] = "ok"
	checkpoint.Metadata["active_connections"] = 100
	checkpoint.Metadata["memory_usage"] = "500MB"
	
	return checkpoint
}

// Start method for CheckpointManager
func (cm *CheckpointManager) Start(ctx context.Context) error {
	// Implementation for starting checkpoint manager
	slog.Info("checkpoint manager started")
	return nil
}

// DetectFault detects a fault in the system
func (ftm *FaultToleranceManager) DetectFault(faultType FaultType, target, description string, metadata map[string]interface{}) *FaultDetection {
	fault := &FaultDetection{
		ID:          fmt.Sprintf("fault_%d", time.Now().UnixNano()),
		Type:        faultType,
		Severity:    ftm.determineSeverity(faultType, metadata),
		Target:      target,
		Description: description,
		DetectedAt:  time.Now(),
		Status:      FaultStatusDetected,
		Metadata:    metadata,
	}
	
	// Store fault detection
	ftm.detectionSystem.detectionsMu.Lock()
	ftm.detectionSystem.detections[fault.ID] = fault
	ftm.detectionSystem.detectionsMu.Unlock()
	
	// Update metrics
	ftm.metrics.FaultsDetected++
	now := time.Now()
	ftm.metrics.LastFault = &now
	
	// Create alert
	alert := &FaultAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		FaultID:   fault.ID,
		Severity:  fault.Severity,
		Message:   fmt.Sprintf("Fault detected: %s - %s", fault.Type, fault.Description),
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	
	// Send alert
	ftm.detectionSystem.alerting.sendAlert(alert)
	
	// Trigger recovery
	go ftm.triggerRecovery(fault)
	
	slog.Warn("fault detected",
		"fault_id", fault.ID,
		"type", fault.Type,
		"severity", fault.Severity,
		"target", fault.Target,
		"description", fault.Description)
	
	return fault
}

// determineSeverity determines the severity of a fault
func (ftm *FaultToleranceManager) determineSeverity(faultType FaultType, metadata map[string]interface{}) FaultSeverity {
	switch faultType {
	case FaultTypeNodeFailure:
		return FaultSeverityHigh
	case FaultTypeNetworkPartition:
		return FaultSeverityCritical
	case FaultTypeResourceExhaustion:
		return FaultSeverityHigh
	case FaultTypePerformanceAnomaly:
		return FaultSeverityMedium
	case FaultTypeServiceUnavailable:
		return FaultSeverityHigh
	default:
		return FaultSeverityMedium
	}
}

// triggerRecovery triggers recovery for a fault
func (ftm *FaultToleranceManager) triggerRecovery(fault *FaultDetection) {
	recoveryRequest := &RecoveryRequest{
		Fault:     fault,
		Priority:  ftm.getPriority(fault.Severity),
		Timestamp: time.Now(),
	}
	
	select {
	case ftm.recoveryEngine.recoveryQueue <- recoveryRequest:
		slog.Debug("recovery request queued", "fault_id", fault.ID)
	case <-time.After(5 * time.Second):
		slog.Warn("recovery queue full, dropping request", "fault_id", fault.ID)
	}
}

// getPriority gets priority based on severity
func (ftm *FaultToleranceManager) getPriority(severity FaultSeverity) int {
	switch severity {
	case FaultSeverityCritical:
		return 1
	case FaultSeverityHigh:
		return 2
	case FaultSeverityMedium:
		return 3
	case FaultSeverityLow:
		return 4
	default:
		return 5
	}
}

// GetMetrics returns fault tolerance metrics
func (ftm *FaultToleranceManager) GetMetrics() *FaultToleranceMetrics {
	ftm.mu.RLock()
	defer ftm.mu.RUnlock()
	
	// Calculate uptime
	if ftm.started {
		ftm.metrics.Uptime = time.Since(time.Now().Add(-ftm.metrics.Uptime))
	}
	
	// Calculate average recovery time
	ftm.recoveryEngine.historyMu.RLock()
	if len(ftm.recoveryEngine.recoveryHistory) > 0 {
		totalTime := time.Duration(0)
		for _, attempt := range ftm.recoveryEngine.recoveryHistory {
			if attempt.Result != nil {
				totalTime += attempt.Result.Duration
			}
		}
		ftm.metrics.AverageRecoveryTime = totalTime / time.Duration(len(ftm.recoveryEngine.recoveryHistory))
	}
	ftm.recoveryEngine.historyMu.RUnlock()
	
	return ftm.metrics
}

// GetFaultDetections returns all fault detections
func (ftm *FaultToleranceManager) GetFaultDetections() []*FaultDetection {
	ftm.detectionSystem.detectionsMu.RLock()
	defer ftm.detectionSystem.detectionsMu.RUnlock()
	
	detections := make([]*FaultDetection, 0, len(ftm.detectionSystem.detections))
	for _, detection := range ftm.detectionSystem.detections {
		detections = append(detections, detection)
	}
	
	return detections
}

// GetRecoveryHistory returns recovery history
func (ftm *FaultToleranceManager) GetRecoveryHistory() []*RecoveryAttempt {
	ftm.recoveryEngine.historyMu.RLock()
	defer ftm.recoveryEngine.historyMu.RUnlock()
	
	history := make([]*RecoveryAttempt, len(ftm.recoveryEngine.recoveryHistory))
	copy(history, ftm.recoveryEngine.recoveryHistory)
	
	return history
}

// Shutdown gracefully shuts down the fault tolerance manager
func (ftm *FaultToleranceManager) Shutdown(ctx context.Context) error {
	ftm.mu.Lock()
	defer ftm.mu.Unlock()
	
	if !ftm.started {
		return nil
	}
	
	slog.Info("shutting down fault tolerance manager")
	
	// Cancel context
	ftm.cancel()
	
	// Wait for components to shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// Shutdown components
	if err := ftm.shutdownComponents(shutdownCtx); err != nil {
		slog.Warn("error during shutdown", "error", err)
	}
	
	ftm.started = false
	
	return nil
}

// shutdownComponents shuts down all components
func (ftm *FaultToleranceManager) shutdownComponents(ctx context.Context) error {
	// Create final checkpoint
	if ftm.checkpointing != nil {
		checkpoint := ftm.checkpointing.CreateCheckpoint()
		if checkpoint == nil {
			slog.Warn("failed to create final checkpoint")
		}
	}
	
	// Close recovery queue
	close(ftm.recoveryEngine.recoveryQueue)
	
	return nil
}
