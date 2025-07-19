package distributed

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ollama/ollama/api"
)

// FaultToleranceManager manages all fault tolerance mechanisms
type FaultToleranceManager struct {
	detectionSystem *FaultDetector
	recoveryEngine  *RecoveryEngine
	replicationMgr  *ReplicationManager
	circuitBreaker  *CircuitBreaker
	checkpointing   *CheckpointManager
	
	config *FaultToleranceConfig
	logger Logger
	
	// State management
	systemState  *SystemState
	stateMutex   sync.RWMutex
	shutdownCh   chan struct{}
	wg           sync.WaitGroup
}

// FaultToleranceConfig defines configuration for fault tolerance
type FaultToleranceConfig struct {
	// Detection settings
	HealthCheckInterval    time.Duration
	FailureThreshold      int
	RecoveryThreshold     int
	NetworkTimeout        time.Duration
	
	// Recovery settings
	MaxRetries            int
	BackoffMultiplier     float64
	MaxBackoffDelay       time.Duration
	GracefulDegradation   bool
	
	// Replication settings
	ReplicationFactor     int
	ConsistencyLevel      string
	AutoFailover          bool
	
	// Circuit breaker settings
	CircuitBreakerEnabled bool
	FailureRateThreshold  float64
	RequestVolumeThreshold int
	SleepWindow           time.Duration
	
	// Checkpointing settings
	CheckpointInterval    time.Duration
	CheckpointRetention   time.Duration
	CompressionEnabled    bool
	EncryptionEnabled     bool
}

// SystemState tracks the overall system state
type SystemState struct {
	Health           SystemHealth
	ActiveNodes      map[string]*NodeState
	FailedNodes      map[string]*FailedNode
	PartitionedNodes map[string]*PartitionedNode
	LastUpdated      time.Time
}

type SystemHealth string

const (
	SystemHealthy     SystemHealth = "healthy"
	SystemDegraded    SystemHealth = "degraded"
	SystemUnhealthy   SystemHealth = "unhealthy"
	SystemPartitioned SystemHealth = "partitioned"
)

type NodeState struct {
	ID             string
	Status         NodeStatus
	Health         HealthStatus
	LastSeen       time.Time
	FailureCount   int
	RecoveryCount  int
	Metrics        *NodeMetrics
	Replicas       []string
}

type HealthStatus string

const (
	HealthStatusHealthy     HealthStatus = "healthy"
	HealthStatusDegraded    HealthStatus = "degraded"
	HealthStatusUnhealthy   HealthStatus = "unhealthy"
	HealthStatusUnreachable HealthStatus = "unreachable"
)

type FailedNode struct {
	ID            string
	FailureTime   time.Time
	FailureReason string
	RecoveryPlan  *RecoveryPlan
	Replicas      []string
}

type PartitionedNode struct {
	ID             string
	PartitionTime  time.Time
	LastContact    time.Time
	IsolatedNodes  []string
	ReconnectPlan  *ReconnectPlan
}

// FaultDetector detects various types of failures
type FaultDetector struct {
	healthCheckers map[string]HealthChecker
	monitors       []SystemMonitor
	alerting       *AlertingSystem
	thresholds     map[string]float64
	
	// Detection state
	detectionResults chan DetectionResult
	stopCh          chan struct{}
	wg              sync.WaitGroup
}

type DetectionResult struct {
	NodeID        string
	FaultType     FaultType
	Severity      Severity
	Timestamp     time.Time
	Details       map[string]interface{}
	Remediation   []RemediationAction
}

type FaultType string

const (
	FaultTypeNodeFailure      FaultType = "node_failure"
	FaultTypeNetworkPartition FaultType = "network_partition"
	FaultTypeResourceExhaustion FaultType = "resource_exhaustion"
	FaultTypePerformanceDegradation FaultType = "performance_degradation"
	FaultTypeMemoryLeak       FaultType = "memory_leak"
	FaultTypeHighLatency      FaultType = "high_latency"
	FaultTypeCorruption       FaultType = "data_corruption"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type RemediationAction struct {
	Type        string
	Priority    int
	Automatic   bool
	Description string
	Parameters  map[string]interface{}
}

// HealthChecker interface for different health check implementations
type HealthChecker interface {
	CheckHealth(ctx context.Context, nodeID string) (*HealthCheckResult, error)
	GetHealthStatus(nodeID string) HealthStatus
	SetThresholds(thresholds map[string]float64)
}

type HealthCheckResult struct {
	NodeID      string
	Status      HealthStatus
	Metrics     map[string]float64
	Timestamp   time.Time
	Errors      []string
	Warnings    []string
}

// SystemMonitor monitors system-wide metrics
type SystemMonitor interface {
	Monitor(ctx context.Context) (*SystemMetrics, error)
	GetAlerts() []Alert
	SetThresholds(thresholds map[string]float64)
}

type Alert struct {
	ID          string
	Type        string
	Severity    Severity
	Message     string
	Timestamp   time.Time
	NodeID      string
	Metadata    map[string]interface{}
	Resolved    bool
	ResolvedAt  time.Time
}

// RecoveryEngine handles automatic recovery
type RecoveryEngine struct {
	strategies     map[FaultType]RecoveryStrategy
	executionQueue chan RecoveryTask
	activeRecovery map[string]*RecoveryExecution
	history        *RecoveryHistory
	
	config *RecoveryConfig
	mutex  sync.RWMutex
}

type RecoveryConfig struct {
	MaxConcurrentRecoveries int
	RecoveryTimeout        time.Duration
	RetryBackoff          time.Duration
	MaxRetries            int
	NotificationEnabled   bool
}

type RecoveryStrategy interface {
	CanHandle(fault FaultType) bool
	CreateRecoveryPlan(fault DetectionResult) (*RecoveryPlan, error)
	Execute(ctx context.Context, plan *RecoveryPlan) error
	Rollback(ctx context.Context, plan *RecoveryPlan) error
}

type RecoveryPlan struct {
	ID          string
	FaultType   FaultType
	NodeID      string
	Steps       []RecoveryStep
	Rollback    []RecoveryStep
	Timeout     time.Duration
	Priority    int
	CreatedAt   time.Time
	Dependencies []string
}

type RecoveryStep struct {
	ID          string
	Type        string
	Action      string
	Parameters  map[string]interface{}
	Timeout     time.Duration
	Retries     int
	Critical    bool
	Completed   bool
	Error       error
}

type RecoveryTask struct {
	Plan      *RecoveryPlan
	Context   context.Context
	Callback  func(error)
	StartTime time.Time
}

type RecoveryExecution struct {
	Task        *RecoveryTask
	CurrentStep int
	Status      ExecutionStatus
	Error       error
	StartTime   time.Time
	EndTime     time.Time
}

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusRolledBack ExecutionStatus = "rolled_back"
)

// ReplicationManager handles model and data replication
type ReplicationManager struct {
	replicas      map[string][]ReplicaInfo
	replicationCh chan ReplicationTask
	
	config *ReplicationConfig
	mutex  sync.RWMutex
}

type ReplicationConfig struct {
	ReplicationFactor    int
	ConsistencyLevel     string
	AutoRepair          bool
	ReplicationTimeout  time.Duration
	ChecksumValidation  bool
}

type ReplicaInfo struct {
	NodeID      string
	Path        string
	Checksum    string
	Size        int64
	LastUpdated time.Time
	Status      ReplicaStatus
}

type ReplicaStatus string

const (
	ReplicaStatusActive     ReplicaStatus = "active"
	ReplicaStatusStale      ReplicaStatus = "stale"
	ReplicaStatusCorrupted  ReplicaStatus = "corrupted"
	ReplicaStatusUnreachable ReplicaStatus = "unreachable"
)

type ReplicationTask struct {
	Type        string
	SourceNode  string
	TargetNodes []string
	ModelName   string
	Priority    int
	Callback    func(error)
}

// CircuitBreaker prevents cascading failures
type CircuitBreaker struct {
	state         CircuitState
	failureCount  int64
	successCount  int64
	lastFailTime  time.Time
	nextRetryTime time.Time
	
	config *CircuitBreakerConfig
	mutex  sync.RWMutex
}

type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half_open"
)

type CircuitBreakerConfig struct {
	FailureRateThreshold   float64
	RequestVolumeThreshold int
	SleepWindow           time.Duration
	MaxRequestsHalfOpen   int
}

// CheckpointManager handles system checkpointing
type CheckpointManager struct {
	storage       CheckpointStorage
	scheduler     *CheckpointScheduler
	compressor    CompressionAlgorithm
	encryptor     EncryptionMethod
	cleaner       *CheckpointCleaner
	
	config *CheckpointConfig
}

type CheckpointConfig struct {
	Interval       time.Duration
	Retention      time.Duration
	Compression    bool
	Encryption     bool
	BackupEnabled  bool
	BackupLocation string
}

type CheckpointStorage interface {
	Store(checkpoint *Checkpoint) error
	Retrieve(id string) (*Checkpoint, error)
	List() ([]CheckpointInfo, error)
	Delete(id string) error
}

type Checkpoint struct {
	ID            string
	Timestamp     time.Time
	SystemState   *SystemState
	NodeStates    map[string]*NodeState
	RequestQueue  []InferenceRequest
	ModelRegistry map[string]*ModelInfo
	Configuration map[string]interface{}
	Metadata      map[string]interface{}
	Checksum      string
	Size          int64
}

type CheckpointInfo struct {
	ID        string
	Timestamp time.Time
	Size      int64
	Checksum  string
}

// NewFaultToleranceManager creates a new fault tolerance manager
func NewFaultToleranceManager(config *FaultToleranceConfig) *FaultToleranceManager {
	ftm := &FaultToleranceManager{
		config:     config,
		systemState: &SystemState{
			Health:           SystemHealthy,
			ActiveNodes:      make(map[string]*NodeState),
			FailedNodes:      make(map[string]*FailedNode),
			PartitionedNodes: make(map[string]*PartitionedNode),
		},
		shutdownCh: make(chan struct{}),
	}

	// Initialize components
	ftm.detectionSystem = NewFaultDetector(config)
	ftm.recoveryEngine = NewRecoveryEngine(config)
	ftm.replicationMgr = NewReplicationManager(config)
	ftm.circuitBreaker = NewCircuitBreaker(config)
	ftm.checkpointing = NewCheckpointManager(config)

	return ftm
}

// Start starts the fault tolerance manager
func (ftm *FaultToleranceManager) Start(ctx context.Context) error {
	ftm.wg.Add(1)
	go ftm.run(ctx)
	
	return nil
}

// Stop stops the fault tolerance manager
func (ftm *FaultToleranceManager) Stop(ctx context.Context) error {
	close(ftm.shutdownCh)
	ftm.wg.Wait()
	return nil
}

// run is the main loop for the fault tolerance manager
func (ftm *FaultToleranceManager) run(ctx context.Context) {
	defer ftm.wg.Done()

	ticker := time.NewTicker(ftm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ftm.shutdownCh:
			return
		case <-ticker.C:
			ftm.performHealthChecks(ctx)
		case detection := <-ftm.detectionSystem.detectionResults:
			ftm.handleDetection(ctx, detection)
		}
	}
}

// performHealthChecks performs periodic health checks
func (ftm *FaultToleranceManager) performHealthChecks(ctx context.Context) {
	ftm.stateMutex.RLock()
	nodes := make([]string, 0, len(ftm.systemState.ActiveNodes))
	for nodeID := range ftm.systemState.ActiveNodes {
		nodes = append(nodes, nodeID)
	}
	ftm.stateMutex.RUnlock()

	for _, nodeID := range nodes {
		go ftm.checkNodeHealth(ctx, nodeID)
	}
}

// checkNodeHealth checks the health of a specific node
func (ftm *FaultToleranceManager) checkNodeHealth(ctx context.Context, nodeID string) {
	for checkType, checker := range ftm.detectionSystem.healthCheckers {
		result, err := checker.CheckHealth(ctx, nodeID)
		if err != nil {
			ftm.logger.Error("Health check failed", "node", nodeID, "type", checkType, "error", err)
			continue
		}

		if result.Status != HealthStatusHealthy {
			detection := DetectionResult{
				NodeID:    nodeID,
				FaultType: ftm.mapHealthToFaultType(result.Status),
				Severity:  ftm.mapHealthToSeverity(result.Status),
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"health_result": result,
					"check_type":    checkType,
				},
			}

			select {
			case ftm.detectionSystem.detectionResults <- detection:
			case <-ctx.Done():
				return
			}
		}
	}
}

// handleDetection handles a fault detection
func (ftm *FaultToleranceManager) handleDetection(ctx context.Context, detection DetectionResult) {
	ftm.logger.Info("Fault detected", "node", detection.NodeID, "type", detection.FaultType, "severity", detection.Severity)

	// Update system state
	ftm.updateSystemState(detection)

	// Create recovery plan
	plan, err := ftm.recoveryEngine.CreateRecoveryPlan(detection)
	if err != nil {
		ftm.logger.Error("Failed to create recovery plan", "error", err)
		return
	}

	// Execute recovery
	if plan != nil {
		ftm.executeRecovery(ctx, plan)
	}
}

// updateSystemState updates the system state based on detection
func (ftm *FaultToleranceManager) updateSystemState(detection DetectionResult) {
	ftm.stateMutex.Lock()
	defer ftm.stateMutex.Unlock()

	nodeState, exists := ftm.systemState.ActiveNodes[detection.NodeID]
	if !exists {
		return
	}

	nodeState.Health = ftm.mapFaultTypeToHealth(detection.FaultType)
	nodeState.FailureCount++
	nodeState.LastSeen = time.Now()

	// Move to failed nodes if critical
	if detection.Severity == SeverityCritical {
		delete(ftm.systemState.ActiveNodes, detection.NodeID)
		ftm.systemState.FailedNodes[detection.NodeID] = &FailedNode{
			ID:            detection.NodeID,
			FailureTime:   time.Now(),
			FailureReason: string(detection.FaultType),
		}
	}

	// Update overall system health
	ftm.updateSystemHealth()
}

// updateSystemHealth updates the overall system health
func (ftm *FaultToleranceManager) updateSystemHealth() {
	totalNodes := len(ftm.systemState.ActiveNodes) + len(ftm.systemState.FailedNodes)
	if totalNodes == 0 {
		ftm.systemState.Health = SystemUnhealthy
		return
	}

	failedRatio := float64(len(ftm.systemState.FailedNodes)) / float64(totalNodes)
	
	switch {
	case failedRatio > 0.5:
		ftm.systemState.Health = SystemUnhealthy
	case failedRatio > 0.3:
		ftm.systemState.Health = SystemDegraded
	case len(ftm.systemState.PartitionedNodes) > 0:
		ftm.systemState.Health = SystemPartitioned
	default:
		ftm.systemState.Health = SystemHealthy
	}
}

// executeRecovery executes a recovery plan
func (ftm *FaultToleranceManager) executeRecovery(ctx context.Context, plan *RecoveryPlan) {
	task := &RecoveryTask{
		Plan:      plan,
		Context:   ctx,
		StartTime: time.Now(),
		Callback: func(err error) {
			if err != nil {
				ftm.logger.Error("Recovery failed", "plan", plan.ID, "error", err)
			} else {
				ftm.logger.Info("Recovery completed", "plan", plan.ID)
			}
		},
	}

	select {
	case ftm.recoveryEngine.executionQueue <- task:
	case <-ctx.Done():
		return
	}
}

// Specific recovery strategies

// GracefulDegradationStrategy implements graceful degradation
type GracefulDegradationStrategy struct {
	config *GracefulDegradationConfig
}

type GracefulDegradationConfig struct {
	QualityReduction    float64
	FeatureDisabling    []string
	ModelDowngrading    bool
	RequestThrottling   bool
}

func (gds *GracefulDegradationStrategy) CanHandle(fault FaultType) bool {
	return fault == FaultTypeResourceExhaustion || fault == FaultTypePerformanceDegradation
}

func (gds *GracefulDegradationStrategy) CreateRecoveryPlan(fault DetectionResult) (*RecoveryPlan, error) {
	plan := &RecoveryPlan{
		ID:        generateRecoveryID(),
		FaultType: fault.FaultType,
		NodeID:    fault.NodeID,
		Steps:     make([]RecoveryStep, 0),
		Timeout:   5 * time.Minute,
		Priority:  2,
		CreatedAt: time.Now(),
	}

	// Add degradation steps
	if gds.config.QualityReduction > 0 {
		plan.Steps = append(plan.Steps, RecoveryStep{
			ID:         "reduce_quality",
			Type:       "quality_reduction",
			Action:     "reduce_inference_quality",
			Parameters: map[string]interface{}{"factor": gds.config.QualityReduction},
			Timeout:    30 * time.Second,
			Retries:    0,
			Critical:   false,
		})
	}

	if gds.config.ModelDowngrading {
		plan.Steps = append(plan.Steps, RecoveryStep{
			ID:         "downgrade_model",
			Type:       "model_downgrade",
			Action:     "switch_to_smaller_model",
			Parameters: map[string]interface{}{},
			Timeout:    2 * time.Minute,
			Retries:    1,
			Critical:   false,
		})
	}

	return plan, nil
}

func (gds *GracefulDegradationStrategy) Execute(ctx context.Context, plan *RecoveryPlan) error {
	for i, step := range plan.Steps {
		err := gds.executeStep(ctx, &step)
		if err != nil {
			if step.Critical {
				return fmt.Errorf("critical step failed: %w", err)
			}
			// Continue with non-critical steps
		}
		plan.Steps[i].Completed = true
	}
	return nil
}

func (gds *GracefulDegradationStrategy) executeStep(ctx context.Context, step *RecoveryStep) error {
	switch step.Action {
	case "reduce_inference_quality":
		return gds.reduceInferenceQuality(ctx, step.Parameters)
	case "switch_to_smaller_model":
		return gds.switchToSmallerModel(ctx, step.Parameters)
	default:
		return fmt.Errorf("unknown recovery action: %s", step.Action)
	}
}

func (gds *GracefulDegradationStrategy) reduceInferenceQuality(ctx context.Context, params map[string]interface{}) error {
	// Implementation for reducing inference quality
	return nil
}

func (gds *GracefulDegradationStrategy) switchToSmallerModel(ctx context.Context, params map[string]interface{}) error {
	// Implementation for switching to smaller model
	return nil
}

func (gds *GracefulDegradationStrategy) Rollback(ctx context.Context, plan *RecoveryPlan) error {
	// Implementation for rolling back graceful degradation
	return nil
}

// Request Migration Strategy
type RequestMigrationStrategy struct {
	loadBalancer LoadBalancer
	coordinator  *RequestCoordinator
}

func (rms *RequestMigrationStrategy) CanHandle(fault FaultType) bool {
	return fault == FaultTypeNodeFailure || fault == FaultTypeNetworkPartition
}

func (rms *RequestMigrationStrategy) CreateRecoveryPlan(fault DetectionResult) (*RecoveryPlan, error) {
	plan := &RecoveryPlan{
		ID:        generateRecoveryID(),
		FaultType: fault.FaultType,
		NodeID:    fault.NodeID,
		Steps:     make([]RecoveryStep, 0),
		Timeout:   2 * time.Minute,
		Priority:  1,
		CreatedAt: time.Now(),
	}

	// Add migration steps
	plan.Steps = append(plan.Steps, RecoveryStep{
		ID:         "migrate_active_requests",
		Type:       "request_migration",
		Action:     "migrate_requests",
		Parameters: map[string]interface{}{"source_node": fault.NodeID},
		Timeout:    1 * time.Minute,
		Retries:    2,
		Critical:   true,
	})

	plan.Steps = append(plan.Steps, RecoveryStep{
		ID:         "update_routing",
		Type:       "routing_update",
		Action:     "exclude_node_from_routing",
		Parameters: map[string]interface{}{"node_id": fault.NodeID},
		Timeout:    10 * time.Second,
		Retries:    3,
		Critical:   true,
	})

	return plan, nil
}

func (rms *RequestMigrationStrategy) Execute(ctx context.Context, plan *RecoveryPlan) error {
	for i, step := range plan.Steps {
		err := rms.executeStep(ctx, &step)
		if err != nil {
			if step.Critical {
				return fmt.Errorf("critical step failed: %w", err)
			}
		}
		plan.Steps[i].Completed = true
	}
	return nil
}

func (rms *RequestMigrationStrategy) executeStep(ctx context.Context, step *RecoveryStep) error {
	switch step.Action {
	case "migrate_requests":
		return rms.migrateRequests(ctx, step.Parameters)
	case "exclude_node_from_routing":
		return rms.excludeNodeFromRouting(ctx, step.Parameters)
	default:
		return fmt.Errorf("unknown recovery action: %s", step.Action)
	}
}

func (rms *RequestMigrationStrategy) migrateRequests(ctx context.Context, params map[string]interface{}) error {
	// Implementation for migrating active requests
	return nil
}

func (rms *RequestMigrationStrategy) excludeNodeFromRouting(ctx context.Context, params map[string]interface{}) error {
	// Implementation for excluding node from routing
	return nil
}

func (rms *RequestMigrationStrategy) Rollback(ctx context.Context, plan *RecoveryPlan) error {
	// Implementation for rolling back request migration
	return nil
}

// Helper functions and types

type RecoveryHistory struct {
	recoveries []RecoveryRecord
	mutex      sync.RWMutex
}

type RecoveryRecord struct {
	ID        string
	NodeID    string
	FaultType FaultType
	Plan      *RecoveryPlan
	Status    ExecutionStatus
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

type ReconnectPlan struct {
	ID          string
	NodeID      string
	Steps       []ReconnectStep
	Timeout     time.Duration
	CreatedAt   time.Time
}

type ReconnectStep struct {
	ID         string
	Action     string
	Parameters map[string]interface{}
	Timeout    time.Duration
}

type CompressionAlgorithm interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
}

type EncryptionMethod interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type CheckpointScheduler struct {
	interval time.Duration
	stopCh   chan struct{}
}

type CheckpointCleaner struct {
	retention time.Duration
	stopCh    chan struct{}
}

type AlertingSystem struct {
	alerts []Alert
	mutex  sync.RWMutex
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// Helper functions

func generateRecoveryID() string {
	return fmt.Sprintf("recovery_%d", time.Now().UnixNano())
}

func (ftm *FaultToleranceManager) mapHealthToFaultType(status HealthStatus) FaultType {
	switch status {
	case HealthStatusUnreachable:
		return FaultTypeNodeFailure
	case HealthStatusUnhealthy:
		return FaultTypeResourceExhaustion
	case HealthStatusDegraded:
		return FaultTypePerformanceDegradation
	default:
		return FaultTypeNodeFailure
	}
}

func (ftm *FaultToleranceManager) mapHealthToSeverity(status HealthStatus) Severity {
	switch status {
	case HealthStatusUnreachable:
		return SeverityCritical
	case HealthStatusUnhealthy:
		return SeverityHigh
	case HealthStatusDegraded:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func (ftm *FaultToleranceManager) mapFaultTypeToHealth(faultType FaultType) HealthStatus {
	switch faultType {
	case FaultTypeNodeFailure:
		return HealthStatusUnreachable
	case FaultTypeResourceExhaustion:
		return HealthStatusUnhealthy
	case FaultTypePerformanceDegradation:
		return HealthStatusDegraded
	default:
		return HealthStatusUnhealthy
	}
}

// Constructor functions

func NewFaultDetector(config *FaultToleranceConfig) *FaultDetector {
	return &FaultDetector{
		healthCheckers:   make(map[string]HealthChecker),
		monitors:         make([]SystemMonitor, 0),
		thresholds:       make(map[string]float64),
		detectionResults: make(chan DetectionResult, 100),
		stopCh:          make(chan struct{}),
	}
}

func NewRecoveryEngine(config *FaultToleranceConfig) *RecoveryEngine {
	return &RecoveryEngine{
		strategies:     make(map[FaultType]RecoveryStrategy),
		executionQueue: make(chan RecoveryTask, 100),
		activeRecovery: make(map[string]*RecoveryExecution),
		history:        &RecoveryHistory{},
	}
}

func NewReplicationManager(config *FaultToleranceConfig) *ReplicationManager {
	return &ReplicationManager{
		replicas:      make(map[string][]ReplicaInfo),
		replicationCh: make(chan ReplicationTask, 100),
	}
}

func NewCircuitBreaker(config *FaultToleranceConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state: CircuitStateClosed,
	}
}

func NewCheckpointManager(config *FaultToleranceConfig) *CheckpointManager {
	return &CheckpointManager{
		scheduler: &CheckpointScheduler{
			interval: config.CheckpointInterval,
			stopCh:   make(chan struct{}),
		},
		cleaner: &CheckpointCleaner{
			retention: config.CheckpointRetention,
			stopCh:    make(chan struct{}),
		},
	}
}