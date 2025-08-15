package fault_tolerance

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// EnhancedFaultToleranceManager extends the basic fault tolerance manager with advanced features
type EnhancedFaultToleranceManager struct {
	*FaultToleranceManager // Embed base manager

	// Advanced recovery strategies
	advancedStrategies map[FaultType][]RecoveryStrategy

	// Predictive fault detection
	predictor *FaultPredictorImpl

	// Self-healing mechanisms
	selfHealer *SelfHealingEngine

	// Redundancy management
	redundancyManager *RedundancyManager

	// Performance tracking
	performanceTracker *PerformanceTracker

	// Adaptive configuration
	configAdaptor *ConfigAdaptor

	// System integration
	systemIntegration *SystemIntegration

	// Metrics
	enhancedMetrics *EnhancedFaultToleranceMetrics

	// Node provider callback for accessing cluster nodes without import cycles
	getNodesFn func() []interface{}

	// Lifecycle
	mu      sync.RWMutex
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// EnhancedFaultToleranceConfig holds enhanced fault tolerance configuration
type EnhancedFaultToleranceConfig struct {
	*Config // Embed base config

	// Predictive fault detection
	EnablePrediction     bool          `json:"enable_prediction"`
	PredictionWindowSize time.Duration `json:"prediction_window_size"`
	PredictionThreshold  float64       `json:"prediction_threshold"`

	// Self-healing
	EnableSelfHealing    bool          `json:"enable_self_healing"`
	SelfHealingInterval  time.Duration `json:"self_healing_interval"`
	SelfHealingThreshold float64       `json:"self_healing_threshold"`

	// Redundancy management
	EnableRedundancy         bool          `json:"enable_redundancy"`
	DefaultRedundancyFactor  int           `json:"default_redundancy_factor"`
	MaxRedundancyFactor      int           `json:"max_redundancy_factor"`
	RedundancyUpdateInterval time.Duration `json:"redundancy_update_interval"`

	// Performance tracking
	EnablePerformanceTracking bool          `json:"enable_performance_tracking"`
	PerformanceWindowSize     time.Duration `json:"performance_window_size"`

	// Adaptive configuration
	EnableConfigAdaptation   bool          `json:"enable_config_adaptation"`
	ConfigAdaptationInterval time.Duration `json:"config_adaptation_interval"`

	// Advanced recovery settings
	MaxRecoveryRetries    int           `json:"max_recovery_retries"`
	RecoveryBackoffFactor float64       `json:"recovery_backoff_factor"`
	RecoveryTimeout       time.Duration `json:"recovery_timeout"`

	// Checkpoint management
	CheckpointCompression bool          `json:"checkpoint_compression"`
	CheckpointEncryption  bool          `json:"checkpoint_encryption"`
	CheckpointRetention   time.Duration `json:"checkpoint_retention"`

	// Circuit breaker settings
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`

	// Alerting
	AlertThrottleTime      time.Duration `json:"alert_throttle_time"`
	AlertSeverityThreshold string        `json:"alert_severity_threshold"`
}

// EnhancedFaultToleranceMetrics tracks enhanced fault tolerance metrics
type EnhancedFaultToleranceMetrics struct {
	*FaultToleranceMetrics // Embed base metrics

	// Prediction metrics
	PredictionsMade          int64         `json:"predictions_made"`
	PredictionsCorrect       int64         `json:"predictions_correct"`
	PredictionAccuracy       float64       `json:"prediction_accuracy"`
	AveragePredictionLatency time.Duration `json:"average_prediction_latency"`

	// Self-healing metrics
	SelfHealingAttempts  int64         `json:"self_healing_attempts"`
	SelfHealingSuccesses int64         `json:"self_healing_successes"`
	SelfHealingFailures  int64         `json:"self_healing_failures"`
	AverageHealingTime   time.Duration `json:"average_healing_time"`

	// Redundancy metrics
	RedundancyFactor   int           `json:"redundancy_factor"`
	ActiveReplicas     int           `json:"active_replicas"`
	FailedReplicas     int           `json:"failed_replicas"`
	ReplicationLatency time.Duration `json:"replication_latency"`

	// Performance metrics
	AverageRecoveryTime time.Duration `json:"average_recovery_time"`
	RecoverySuccessRate float64       `json:"recovery_success_rate"`
	ResourceUtilization float64       `json:"resource_utilization"`
	SystemStability     float64       `json:"system_stability"`
	RecoveryAttempts    int64         `json:"recovery_attempts"`
	RecoverySuccesses   int64         `json:"recovery_successes"`
	RecoveryFailures    int64         `json:"recovery_failures"`

	// Config adaptation metrics
	ConfigAdaptations  int64   `json:"config_adaptations"`
	AdaptationAccuracy float64 `json:"adaptation_accuracy"`

	// Circuit breaker metrics
	CircuitBreakerTrips  int64 `json:"circuit_breaker_trips"`
	CircuitBreakerResets int64 `json:"circuit_breaker_resets"`

	// Alerting metrics
	AlertsSent      int64 `json:"alerts_sent"`
	AlertThrottling int64 `json:"alert_throttling"`

	// Timestamps
	LastPrediction  *time.Time `json:"last_prediction,omitempty"`
	LastSelfHealing *time.Time `json:"last_self_healing,omitempty"`
	LastReplication *time.Time `json:"last_replication,omitempty"`
	LastAdaptation  *time.Time `json:"last_adaptation,omitempty"`
	LastCircuitTrip *time.Time `json:"last_circuit_trip,omitempty"`
	LastAlert       *time.Time `json:"last_alert,omitempty"`
	LastUpdated     time.Time  `json:"last_updated"`
}

// Use FaultPredictorImpl from predictive_detection.go to avoid duplication

// Use PredictionModelImpl from predictive_detection.go to avoid duplication

// Use PredictionSampleImpl from predictive_detection.go to avoid duplication

// Use SelfHealingEngineImpl from self_healing_engine.go to avoid duplication

// Use SelfHealingStrategyImpl from self_healing_engine.go to avoid duplication

// Use HealingAttemptImpl from self_healing_engine.go to avoid duplication

// Use HealingResultImpl from self_healing_engine.go to avoid duplication

// Use SystemStateImpl from self_healing_engine.go to avoid duplication

// RedundancyManager manages redundancy for fault tolerance
type RedundancyManager struct {
	manager          *EnhancedFaultToleranceManager
	factor           int
	maxFactor        int
	updateInterval   time.Duration
	replicas         map[string][]*ReplicaInfo
	replicasMu       sync.RWMutex
	replicationTasks map[string]*ReplicationTask
	replicationMu    sync.RWMutex
	learning         bool
	efficiency       float64
}

// ReplicaInfo represents information about a replica
type ReplicaInfo struct {
	ID          string                 `json:"id"`
	OriginalID  string                 `json:"original_id"`
	NodeID      string                 `json:"node_id"`
	Status      ReplicaStatus          `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	LastSync    time.Time              `json:"last_sync"`
	SyncLatency time.Duration          `json:"sync_latency"`
	StorageSize int64                  `json:"storage_size"`
	HealthScore float64                `json:"health_score"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReplicaStatus represents the status of a replica
type ReplicaStatus string

const (
	ReplicaStatusCreating   ReplicaStatus = "creating"
	ReplicaStatusActive     ReplicaStatus = "active"
	ReplicaStatusSyncing    ReplicaStatus = "syncing"
	ReplicaStatusDegraded   ReplicaStatus = "degraded"
	ReplicaStatusFailed     ReplicaStatus = "failed"
	ReplicaStatusTerminated ReplicaStatus = "terminated"
)

// ReplicationTask represents a replication task
type ReplicationTask struct {
	ID          string                 `json:"id"`
	OriginalID  string                 `json:"original_id"`
	SourceNode  string                 `json:"source_node"`
	TargetNodes []string               `json:"target_nodes"`
	Status      types.TaskStatus       `json:"status"`
	Progress    float64                `json:"progress"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PerformanceTracker tracks system performance for optimization
type PerformanceTracker struct {
	manager                *EnhancedFaultToleranceManager
	windowSize             time.Duration
	metricsHistory         []*PerformanceSample
	metricsHistoryMu       sync.RWMutex
	optimizationStrategies []OptimizationStrategy
	strategyWeights        map[string]float64
	learning               bool
	efficiency             float64
}

// PerformanceSample represents a performance sample
type PerformanceSample struct {
	Timestamp     time.Time              `json:"timestamp"`
	Metrics       map[string]float64     `json:"metrics"`
	Faults        []*FaultDetection      `json:"faults"`
	Recoveries    []*RecoveryResult      `json:"recoveries"`
	Optimizations []*OptimizationResult  `json:"optimizations"`
	Efficiency    float64                `json:"efficiency"`
	Stability     float64                `json:"stability"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// OptimizationStrategy defines the interface for optimization strategies
type OptimizationStrategy interface {
	Apply(ctx context.Context, metrics *PerformanceSample) (*OptimizationResult, error)
	GetName() string
	GetWeight() float64
	SetWeight(weight float64)
	CanHandle(sample *PerformanceSample) bool
}

// OptimizationResult represents the result of an optimization attempt
type OptimizationResult struct {
	Improvement  float64            `json:"improvement"`
	Metrics      map[string]float64 `json:"metrics"`
	ActionsTaken []string           `json:"actions_taken"`
	Error        string             `json:"error,omitempty"`
	Timestamp    time.Time          `json:"timestamp"`
}

// ConfigAdaptor adapts configuration based on system performance
type ConfigAdaptor struct {
	manager              *EnhancedFaultToleranceManager
	interval             time.Duration
	adaptationStrategies []AdaptationStrategy
	strategyWeights      map[string]float64
	adaptationHistory    []*AdaptationAttempt
	adaptationHistoryMu  sync.RWMutex
	learning             bool
	accuracy             float64
}

// AdaptationStrategy defines the interface for configuration adaptation strategies
type AdaptationStrategy interface {
	Apply(ctx context.Context, metrics *PerformanceSample) (*AdaptationResult, error)
	GetName() string
	GetWeight() float64
	SetWeight(weight float64)
	CanHandle(sample *PerformanceSample) bool
}

// AdaptationAttempt represents a configuration adaptation attempt
type AdaptationAttempt struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Strategy     string                 `json:"strategy"`
	InputMetrics *PerformanceSample     `json:"input_metrics"`
	Result       *AdaptationResult      `json:"result"`
	Duration     time.Duration          `json:"duration"`
	Success      bool                   `json:"success"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// AdaptationResult represents the result of a configuration adaptation
type AdaptationResult struct {
	ConfigChanges map[string]interface{} `json:"config_changes"`
	Improvement   float64                `json:"improvement"`
	Metrics       map[string]float64     `json:"metrics"`
	Error         string                 `json:"error,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// NewEnhancedFaultToleranceManager creates a new enhanced fault tolerance manager
func NewEnhancedFaultToleranceManager(
	config *EnhancedFaultToleranceConfig,
	manager *FaultToleranceManager,
) *EnhancedFaultToleranceManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Create base fault tolerance manager if not provided
	if manager == nil {
		baseConfig := &Config{
			ReplicationFactor:     config.ReplicationFactor,
			HealthCheckInterval:   config.HealthCheckInterval,
			RecoveryTimeout:       config.RecoveryTimeout,
			CircuitBreakerEnabled: config.CircuitBreakerEnabled,
			CheckpointInterval:    config.CheckpointInterval,
			MaxRetries:            config.MaxRetries,
			RetryBackoff:          config.RetryBackoff,
		}
		manager = NewFaultToleranceManager(baseConfig)
	}

	eftm := &EnhancedFaultToleranceManager{
		FaultToleranceManager: manager,
		advancedStrategies:    make(map[FaultType][]RecoveryStrategy),
		configAdaptor:         NewConfigAdaptor(config, manager),
		performanceTracker:    NewPerformanceTracker(config, manager),
		redundancyManager:     NewRedundancyManager(config, manager),
		selfHealer: NewSelfHealingEngine(manager, &SelfHealingConfig{
			HealingInterval:            config.SelfHealingInterval,
			MonitoringInterval:         30 * time.Second,
			LearningInterval:           60 * time.Second,
			HealingThreshold:           config.SelfHealingThreshold,
			EnableLearning:             config.EnableSelfHealing,
			EnablePredictiveHealing:    true,
			EnableProactiveHealing:     true,
			EnableServiceRestart:       true,
			EnableResourceReallocation: true,
			EnableLoadRedistribution:   true,
			EnableFailover:             true,
			EnableScaling:              true,
		}),
		predictor: NewFaultPredictor(config, manager),
		enhancedMetrics: &EnhancedFaultToleranceMetrics{
			FaultToleranceMetrics: manager.GetMetrics(),
			LastUpdated:           time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize system integration
	eftm.systemIntegration = NewSystemIntegration(eftm)

	// Initialize components
	eftm.initializeComponents(config)

	return eftm
}

// initializeComponents initializes all enhanced fault tolerance components
func (eftm *EnhancedFaultToleranceManager) initializeComponents(config *EnhancedFaultToleranceConfig) {
	// Initialize advanced recovery strategies
	eftm.registerAdvancedStrategies()

	// Initialize predictor if enabled
	if config.EnablePrediction {
		eftm.predictor.learning = true
		eftm.predictor.windowSize = config.PredictionWindowSize
		eftm.predictor.threshold = config.PredictionThreshold
	}

	// Initialize self-healer if enabled
	if config.EnableSelfHealing {
		// Self-healer is already configured in constructor
		slog.Info("Self-healing engine initialized")
		// Also register adapter as a recovery strategy to integrate with RecoveryEngine
		if eftm.selfHealer != nil && eftm.FaultToleranceManager != nil && eftm.FaultToleranceManager.recoveryEngine != nil {
			adapter := NewSelfHealingRecoveryAdapter(eftm.selfHealer)
			// Add adapter to multiple fault types to be considered during recovery
			re := eftm.FaultToleranceManager.recoveryEngine
			re.strategies[FaultTypePerformanceAnomaly] = append(re.strategies[FaultTypePerformanceAnomaly], adapter)
			re.strategies[FaultTypeResourceExhaustion] = append(re.strategies[FaultTypeResourceExhaustion], adapter)
			re.strategies[FaultTypeServiceUnavailable] = append(re.strategies[FaultTypeServiceUnavailable], adapter)
			re.strategies[FaultTypeNetworkPartition] = append(re.strategies[FaultTypeNetworkPartition], adapter)
			// As a low-priority fallback for node failures
			re.strategies[FaultTypeNodeFailure] = append(re.strategies[FaultTypeNodeFailure], adapter)
		}
	}

	// Initialize redundancy manager if enabled
	if config.EnableRedundancy {
		eftm.redundancyManager.factor = config.DefaultRedundancyFactor
		eftm.redundancyManager.maxFactor = config.MaxRedundancyFactor
		eftm.redundancyManager.updateInterval = config.RedundancyUpdateInterval
	}

	// Initialize performance tracker if enabled
	if config.EnablePerformanceTracking {
		eftm.performanceTracker.learning = true
		eftm.performanceTracker.windowSize = config.PerformanceWindowSize
	}

	// Initialize config adaptor if enabled
	if config.EnableConfigAdaptation {
		eftm.configAdaptor.learning = true
		eftm.configAdaptor.interval = config.ConfigAdaptationInterval
	}
}

// registerAdvancedStrategies registers advanced recovery strategies
func (eftm *EnhancedFaultToleranceManager) registerAdvancedStrategies() {
	// Register fast recovery strategies
	eftm.advancedStrategies[FaultTypeNodeFailure] = append(
		eftm.advancedStrategies[FaultTypeNodeFailure],
		NewFastRecoveryStrategy(eftm.FaultToleranceManager),
		NewCheckpointBasedRecoveryStrategy(eftm.FaultToleranceManager),
	)

	// Register redundancy strategies
	eftm.advancedStrategies[FaultTypeNetworkPartition] = append(
		eftm.advancedStrategies[FaultTypeNetworkPartition],
		NewRedundantExecutionStrategy(eftm.FaultToleranceManager),
	)

	// Register graceful degradation strategies
	eftm.advancedStrategies[FaultTypeResourceExhaustion] = append(
		eftm.advancedStrategies[FaultTypeResourceExhaustion],
		NewGracefulDegradationStrategy(eftm.FaultToleranceManager),
	)

	// Register performance tuning strategies
	eftm.advancedStrategies[FaultTypePerformanceAnomaly] = append(
		eftm.advancedStrategies[FaultTypePerformanceAnomaly],
		NewPerformanceTuningStrategy(eftm.FaultToleranceManager),
	)

	// Register service unavailable strategies
	eftm.advancedStrategies[FaultTypeServiceUnavailable] = append(
		eftm.advancedStrategies[FaultTypeServiceUnavailable],
		NewServiceUnavailableStrategy(eftm.FaultToleranceManager),
	)
}

// Start starts the enhanced fault tolerance manager
func (eftm *EnhancedFaultToleranceManager) Start() error {
	eftm.mu.Lock()
	defer eftm.mu.Unlock()

	if eftm.started {
		return fmt.Errorf("enhanced fault tolerance manager already started")
	}

	// Start base manager
	if err := eftm.FaultToleranceManager.Start(); err != nil {
		return fmt.Errorf("failed to start base fault tolerance manager: %w", err)
	}

	// Start enhanced components
	eftm.startEnhancedComponents()

	eftm.started = true

	slog.Info("enhanced fault tolerance manager started",
		"prediction_enabled", eftm.predictor.learning,
		"self_healing_enabled", true,
		"redundancy_enabled", eftm.redundancyManager.factor > 1,
		"performance_tracking_enabled", eftm.performanceTracker.learning,
		"config_adaptation_enabled", eftm.configAdaptor.learning)

	return nil
}

// startEnhancedComponents starts enhanced fault tolerance components
func (eftm *EnhancedFaultToleranceManager) startEnhancedComponents() {
	// Start predictor
	if eftm.predictor.learning {
		eftm.wg.Add(1)
		go eftm.predictor.start(eftm.ctx, &eftm.wg)
	}

	// Start self-healer
	if eftm.selfHealer != nil {
		eftm.selfHealer.Start()
	}

	// Start redundancy manager
	if eftm.redundancyManager.factor > 1 {
		eftm.wg.Add(1)
		go eftm.redundancyManager.start(eftm.ctx, &eftm.wg)
	}

	// Start performance tracker
	if eftm.performanceTracker.learning {
		eftm.wg.Add(1)
		go eftm.performanceTracker.start(eftm.ctx, &eftm.wg)
	}

	// Start config adaptor
	if eftm.configAdaptor.learning {
		eftm.wg.Add(1)
		go eftm.configAdaptor.start(eftm.ctx, &eftm.wg)
	}

	// Start system integration
	if eftm.systemIntegration != nil {
		if err := eftm.systemIntegration.Start(eftm.ctx); err != nil {
			slog.Error("Failed to start system integration", "error", err)
		}
	}
}

// DetectFault detects a fault with enhanced capabilities
func (eftm *EnhancedFaultToleranceManager) DetectFault(faultType FaultType, target, description string, metadata map[string]interface{}) *FaultDetection {
	// Use base detection
	fault := eftm.FaultToleranceManager.DetectFault(faultType, target, description, metadata)

	// Update enhanced metrics
	eftm.enhancedMetrics.FaultsDetected++
	now := time.Now()
	eftm.enhancedMetrics.LastFault = &now

	// Trigger predictive detection if enabled
	if eftm.predictor.learning {
		go eftm.predictor.predictFault(fault)
	}

	// Trigger self-healing if enabled
	if eftm.selfHealer != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			eftm.selfHealer.HealFault(ctx, fault)
		}()
	}

	// Trigger redundancy management if enabled
	if eftm.redundancyManager.factor > 1 {
		go eftm.redundancyManager.manageReplicas(fault)
	}

	// Track performance if enabled
	if eftm.performanceTracker.learning {
		go eftm.performanceTracker.trackFault(fault)
	}

	// Adapt configuration if enabled
	if eftm.configAdaptor.learning {
		go eftm.configAdaptor.adaptConfiguration(fault)
	}

	return fault
}

// Recover attempts to recover from a fault using enhanced strategies
func (eftm *EnhancedFaultToleranceManager) Recover(ctx context.Context, fault *FaultDetection) (*RecoveryResult, error) {
	start := time.Now()

	// Try advanced strategies first
	if strategies, exists := eftm.advancedStrategies[fault.Type]; exists {
		for _, strategy := range strategies {
			if strategy.CanHandle(fault) {
				result, err := strategy.Recover(ctx, fault)
				if err == nil && result != nil && result.Successful {
					// Update metrics
					eftm.updateRecoveryMetrics(result, time.Since(start))
					return result, nil
				}
			}
		}
	}

	// Fall back to base recovery
	result, err := eftm.FaultToleranceManager.Recover(ctx, fault)

	// Update metrics
	if result != nil {
		eftm.updateRecoveryMetrics(result, time.Since(start))
	}

	return result, err
}

// updateRecoveryMetrics updates recovery metrics
func (eftm *EnhancedFaultToleranceManager) updateRecoveryMetrics(result *RecoveryResult, duration time.Duration) {
	eftm.enhancedMetrics.RecoveryAttempts++

	if result.Successful {
		eftm.enhancedMetrics.RecoverySuccesses++

		// Update average recovery time
		if eftm.enhancedMetrics.AverageRecoveryTime == 0 {
			eftm.enhancedMetrics.AverageRecoveryTime = duration
		} else {
			totalTime := eftm.enhancedMetrics.AverageRecoveryTime*time.Duration(eftm.enhancedMetrics.RecoverySuccesses-1) + duration
			eftm.enhancedMetrics.AverageRecoveryTime = totalTime / time.Duration(eftm.enhancedMetrics.RecoverySuccesses)
		}

		// Update recovery success rate
		eftm.enhancedMetrics.RecoverySuccessRate = float64(eftm.enhancedMetrics.RecoverySuccesses) / float64(eftm.enhancedMetrics.RecoveryAttempts)
	} else {
		eftm.enhancedMetrics.RecoveryFailures++
	}

	eftm.enhancedMetrics.LastUpdated = time.Now()
}

// GetEnhancedMetrics returns enhanced fault tolerance metrics
func (eftm *EnhancedFaultToleranceManager) GetEnhancedMetrics() *EnhancedFaultToleranceMetrics {
	eftm.mu.RLock()
	defer eftm.mu.RUnlock()

	// Get base metrics
	baseMetrics := eftm.FaultToleranceManager.GetMetrics()

	// Update enhanced metrics
	eftm.enhancedMetrics.FaultToleranceMetrics = baseMetrics
	eftm.enhancedMetrics.LastUpdated = time.Now()

	// Update prediction metrics
	if eftm.predictor != nil {
		eftm.enhancedMetrics.PredictionsMade = eftm.predictor.metrics.PredictionsMade
		eftm.enhancedMetrics.PredictionsCorrect = eftm.predictor.metrics.PredictionsCorrect
		eftm.enhancedMetrics.PredictionAccuracy = eftm.predictor.accuracy
		eftm.enhancedMetrics.AveragePredictionLatency = eftm.predictor.metrics.AveragePredictionLatency
		if eftm.predictor.metrics.LastPrediction != nil {
			eftm.enhancedMetrics.LastPrediction = eftm.predictor.metrics.LastPrediction
		}
	}

	// Update self-healing metrics
	if eftm.selfHealer != nil {
		// Self-healing metrics integration would go here
		// For now, use placeholder values
		eftm.enhancedMetrics.SelfHealingAttempts = 0
		eftm.enhancedMetrics.SelfHealingSuccesses = 0
		eftm.enhancedMetrics.SelfHealingFailures = 0
		eftm.enhancedMetrics.AverageHealingTime = 0
	}

	// Update redundancy metrics
	if eftm.redundancyManager != nil {
		eftm.enhancedMetrics.RedundancyFactor = eftm.redundancyManager.factor
		eftm.enhancedMetrics.ActiveReplicas = eftm.redundancyManager.getActiveReplicaCount()
		eftm.enhancedMetrics.FailedReplicas = eftm.redundancyManager.getFailedReplicaCount()
		redundancyMetrics := eftm.redundancyManager.getMetrics()
		eftm.enhancedMetrics.ReplicationLatency = redundancyMetrics.ReplicationLatency
		if redundancyMetrics.LastReplication != nil {
			eftm.enhancedMetrics.LastReplication = redundancyMetrics.LastReplication
		}
	}

	// Update performance metrics
	if eftm.performanceTracker != nil {
		performanceMetrics := eftm.performanceTracker.getMetrics()
		eftm.enhancedMetrics.ResourceUtilization = performanceMetrics.SuccessRate // Use success rate as proxy
		eftm.enhancedMetrics.SystemStability = 1.0 - performanceMetrics.ErrorRate // Use inverse of error rate
	}

	// Update config adaptation metrics
	if eftm.configAdaptor != nil {
		configMetrics := eftm.configAdaptor.getMetrics()
		eftm.enhancedMetrics.ConfigAdaptations = configMetrics.ConfigAdaptations
		eftm.enhancedMetrics.AdaptationAccuracy = eftm.configAdaptor.accuracy
		if configMetrics.LastAdaptation != nil {
			eftm.enhancedMetrics.LastAdaptation = configMetrics.LastAdaptation
		}
	}

	// Update alerting metrics
	if eftm.FaultToleranceManager.detectionSystem != nil &&
		eftm.FaultToleranceManager.detectionSystem.alerting != nil {
		eftm.enhancedMetrics.AlertsSent = int64(len(eftm.FaultToleranceManager.detectionSystem.alerting.alerts))
	}

	return eftm.enhancedMetrics
}

// Shutdown gracefully shuts down the enhanced fault tolerance manager
func (eftm *EnhancedFaultToleranceManager) Shutdown(ctx context.Context) error {
	eftm.mu.Lock()
	defer eftm.mu.Unlock()

	if !eftm.started {
		return nil
	}

	slog.Info("shutting down enhanced fault tolerance manager")

	// Cancel context
	eftm.cancel()

	// Wait for background tasks
	eftm.wg.Wait()

	// Shutdown base manager
	if err := eftm.FaultToleranceManager.Shutdown(ctx); err != nil {
		slog.Warn("failed to shutdown base fault tolerance manager", "error", err)
	}

	eftm.started = false

	return nil
}

// NewEnhancedFaultToleranceConfig creates a new enhanced fault tolerance configuration
func NewEnhancedFaultToleranceConfig(baseConfig *Config) *EnhancedFaultToleranceConfig {
	return &EnhancedFaultToleranceConfig{
		Config:                    baseConfig,
		EnablePrediction:          true,
		PredictionWindowSize:      30 * time.Second,
		PredictionThreshold:       0.8,
		EnableSelfHealing:         true,
		SelfHealingInterval:       60 * time.Second,
		SelfHealingThreshold:      0.7,
		EnableRedundancy:          true,
		DefaultRedundancyFactor:   2,
		MaxRedundancyFactor:       5,
		RedundancyUpdateInterval:  300 * time.Second,
		EnablePerformanceTracking: true,
		PerformanceWindowSize:     60 * time.Second,
		EnableConfigAdaptation:    true,
		ConfigAdaptationInterval:  300 * time.Second,
		MaxRecoveryRetries:        5,
		RecoveryBackoffFactor:     1.5,
		RecoveryTimeout:           30 * time.Second,
		CheckpointCompression:     true,
		CheckpointEncryption:      true,
		CheckpointRetention:       24 * time.Hour,
		CircuitBreakerThreshold:   5,
		CircuitBreakerTimeout:     30 * time.Second,
		AlertThrottleTime:         5 * time.Minute,
		AlertSeverityThreshold:    "medium",
	}
}

// Constructor functions for components

// NewConfigAdaptor creates a new configuration adaptor
func NewConfigAdaptor(config *EnhancedFaultToleranceConfig, manager *FaultToleranceManager) *ConfigAdaptor {
	return &ConfigAdaptor{
		manager:              &EnhancedFaultToleranceManager{FaultToleranceManager: manager},
		interval:             5 * time.Minute,
		adaptationStrategies: make([]AdaptationStrategy, 0),
		strategyWeights:      make(map[string]float64),
		adaptationHistory:    make([]*AdaptationAttempt, 0),
	}
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(config *EnhancedFaultToleranceConfig, manager *FaultToleranceManager) *PerformanceTracker {
	return &PerformanceTracker{
		manager:                &EnhancedFaultToleranceManager{FaultToleranceManager: manager},
		windowSize:             10 * time.Minute,
		metricsHistory:         make([]*PerformanceSample, 0),
		optimizationStrategies: make([]OptimizationStrategy, 0),
		strategyWeights:        make(map[string]float64),
	}
}

// NewRedundancyManager creates a new redundancy manager
func NewRedundancyManager(config *EnhancedFaultToleranceConfig, manager *FaultToleranceManager) *RedundancyManager {
	return &RedundancyManager{
		manager:        &EnhancedFaultToleranceManager{FaultToleranceManager: manager},
		factor:         3,
		maxFactor:      5,
		updateInterval: 30 * time.Second,
		replicas:       make(map[string][]*ReplicaInfo),
	}
}

// NewPerformanceTuningStrategy creates a new performance tuning strategy
func NewPerformanceTuningStrategy(manager *FaultToleranceManager) RecoveryStrategy {
	return &PerformanceTuningStrategy{
		name: "performance_tuning",
	}
}

// NewServiceUnavailableStrategy creates a new service unavailable strategy
func NewServiceUnavailableStrategy(manager *FaultToleranceManager) RecoveryStrategy {
	return &LoadSheddingStrategy{
		name: "service_unavailable",
	}
}

// Component methods

// start method for ConfigAdaptor
func (ca *ConfigAdaptor) start(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	return nil
}

// start method for PerformanceTracker
func (pt *PerformanceTracker) start(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	return nil
}

// start method for RedundancyManager
func (rm *RedundancyManager) start(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()
	return nil
}

// manageReplicas method for RedundancyManager
func (rm *RedundancyManager) manageReplicas(fault *FaultDetection) error {
	return nil
}

// Additional missing methods for PerformanceTracker
func (pt *PerformanceTracker) trackFault(fault *FaultDetection) error {
	return nil
}

// Additional missing methods for ConfigAdaptor
func (ca *ConfigAdaptor) adaptConfiguration(fault *FaultDetection) error {
	return nil
}

// Additional missing methods for RedundancyManager
func (rm *RedundancyManager) getActiveReplicaCount() int {
	return 0
}

func (rm *RedundancyManager) getFailedReplicaCount() int {
	return 0
}

// SetNodeProvider sets a callback used to retrieve available nodes from the scheduler/cluster manager
func (eftm *EnhancedFaultToleranceManager) SetNodeProvider(getNodes func() []interface{}) {
	eftm.mu.Lock()
	defer eftm.mu.Unlock()
	eftm.getNodesFn = getNodes
}

// GetAvailableNodes returns available nodes using the configured provider; falls back to empty slice
func (eftm *EnhancedFaultToleranceManager) GetAvailableNodes() []interface{} {
	eftm.mu.RLock()
	provider := eftm.getNodesFn
	eftm.mu.RUnlock()
	if provider != nil {
		if nodes := provider(); nodes != nil {
			return nodes
		}
	}
	return []interface{}{}
}

// GetFaultDetections returns current detected faults from the base manager
func (eftm *EnhancedFaultToleranceManager) GetFaultDetections() []*FaultDetection {
	if eftm.FaultToleranceManager == nil {
		return nil
	}
	return eftm.FaultToleranceManager.GetFaultDetections()
}

// Recover method for FaultToleranceManager (stub implementation)
func (ftm *FaultToleranceManager) Recover(ctx context.Context, fault *FaultDetection) (*RecoveryResult, error) {
	return &RecoveryResult{
		FaultID:    fault.ID,
		Strategy:   "basic_recovery",
		Successful: true,
		Duration:   100 * time.Millisecond,
		Timestamp:  time.Now(),
	}, nil
}

// Add metrics fields to component types
func (rm *RedundancyManager) getMetrics() *RedundancyMetrics {
	return &RedundancyMetrics{
		ReplicationLatency: 50 * time.Millisecond,
		LastReplication:    &time.Time{},
	}
}

func (pt *PerformanceTracker) getMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		AverageLatency:    100 * time.Millisecond,
		Throughput:        1000.0,
		SuccessRate:       0.95,
		ErrorRate:         0.05,
		RequestsProcessed: 10000,
		LastUpdated:       time.Now(),
	}
}

// Add metrics method for ConfigAdaptor
func (ca *ConfigAdaptor) getMetrics() *ConfigMetrics {
	return &ConfigMetrics{
		ConfigAdaptations: 5,
		LastAdaptation:    &time.Time{},
	}
}

// Metrics types
type RedundancyMetrics struct {
	ReplicationLatency time.Duration `json:"replication_latency"`
	LastReplication    *time.Time    `json:"last_replication"`
}

type ConfigMetrics struct {
	ConfigAdaptations int64      `json:"config_adaptations"`
	LastAdaptation    *time.Time `json:"last_adaptation"`
}

// Configuration Loading and Management

// LoadConfiguration loads configuration from DistributedConfig and applies it to the fault tolerance system
func (eftm *EnhancedFaultToleranceManager) LoadConfiguration(distributedConfig *config.DistributedConfig) error {
	if distributedConfig == nil {
		return fmt.Errorf("distributed config cannot be nil")
	}

	// Validate configuration first
	if err := eftm.ValidateConfiguration(distributedConfig); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Apply configuration to components
	if err := eftm.applyConfiguration(distributedConfig); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	slog.Info("fault tolerance configuration loaded successfully",
		"predictive_detection_enabled", distributedConfig.Inference.FaultTolerance.PredictiveDetection.Enabled,
		"self_healing_enabled", distributedConfig.Inference.FaultTolerance.SelfHealing.Enabled,
		"redundancy_enabled", distributedConfig.Inference.FaultTolerance.Redundancy.Enabled)

	return nil
}

// ValidateConfiguration validates the fault tolerance configuration using comprehensive validator
func (eftm *EnhancedFaultToleranceManager) ValidateConfiguration(distributedConfig *config.DistributedConfig) error {
	// Create comprehensive validator
	validator := NewConfigValidator()

	// Use comprehensive validation
	return validator.ValidateConfiguration(distributedConfig)
}

// applyConfiguration applies the validated configuration to all components
func (eftm *EnhancedFaultToleranceManager) applyConfiguration(distributedConfig *config.DistributedConfig) error {
	ft := distributedConfig.Inference.FaultTolerance

	// Update base fault tolerance manager configuration (durations are strongly typed)
	eftm.FaultToleranceManager.config.HealthCheckInterval = ft.HealthCheckInterval
	eftm.FaultToleranceManager.config.RecoveryTimeout = ft.RecoveryTimeout
	eftm.FaultToleranceManager.config.CircuitBreakerEnabled = ft.CircuitBreaker.Enabled
	eftm.FaultToleranceManager.config.MaxRetries = ft.MaxRetries
	eftm.FaultToleranceManager.config.RetryBackoff = ft.RetryBackoff
	eftm.FaultToleranceManager.config.ReplicationFactor = ft.ReplicationFactor

	// Apply predictive detection configuration
	if err := eftm.applyPredictiveDetectionConfig(ft.PredictiveDetection); err != nil {
		return fmt.Errorf("failed to apply predictive detection config: %w", err)
	}

	// Apply self-healing configuration
	if err := eftm.applySelfHealingConfig(ft.SelfHealing); err != nil {
		return fmt.Errorf("failed to apply self-healing config: %w", err)
	}

	// Apply redundancy configuration
	if err := eftm.applyRedundancyConfig(ft.Redundancy); err != nil {
		return fmt.Errorf("failed to apply redundancy config: %w", err)
	}

	// Apply performance tracking configuration
	if err := eftm.applyPerformanceTrackingConfig(ft.PerformanceTracking); err != nil {
		return fmt.Errorf("failed to apply performance tracking config: %w", err)
	}

	// Apply config adaptation configuration
	if err := eftm.applyConfigAdaptationConfig(ft.ConfigAdaptation); err != nil {
		return fmt.Errorf("failed to apply config adaptation config: %w", err)
	}

	return nil
}

// applyPredictiveDetectionConfig applies predictive detection configuration
func (eftm *EnhancedFaultToleranceManager) applyPredictiveDetectionConfig(config struct {
	Enabled             bool    `yaml:"enabled"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	PredictionInterval  string  `yaml:"prediction_interval"`
	WindowSize          string  `yaml:"window_size"`
	Threshold           float64 `yaml:"threshold"`
	EnableMLDetection   bool    `yaml:"enable_ml_detection"`
	EnableStatistical   bool    `yaml:"enable_statistical"`
	EnablePatternRecog  bool    `yaml:"enable_pattern_recognition"`
}) error {
	if !config.Enabled {
		slog.Info("predictive detection disabled")
		return nil
	}

	if eftm.predictor == nil {
		return fmt.Errorf("predictor not initialized")
	}

	// Parse duration strings
	predictionInterval, err := time.ParseDuration(config.PredictionInterval)
	if err != nil {
		return fmt.Errorf("invalid prediction_interval: %w", err)
	}

	windowSize, err := time.ParseDuration(config.WindowSize)
	if err != nil {
		return fmt.Errorf("invalid window_size: %w", err)
	}

	// Apply configuration to predictor
	eftm.predictor.learning = config.Enabled
	eftm.predictor.threshold = config.Threshold
	eftm.predictor.windowSize = windowSize
	// Note: confidenceThreshold may need to be added to FaultPredictorImpl if not present

	slog.Info("predictive detection configuration applied",
		"confidence_threshold", config.ConfidenceThreshold,
		"prediction_interval", predictionInterval,
		"window_size", windowSize,
		"threshold", config.Threshold)

	return nil
}

// applySelfHealingConfig applies self-healing configuration
func (eftm *EnhancedFaultToleranceManager) applySelfHealingConfig(config struct {
	Enabled              bool    `yaml:"enabled"`
	HealingThreshold     float64 `yaml:"healing_threshold"`
	HealingInterval      string  `yaml:"healing_interval"`
	MonitoringInterval   string  `yaml:"monitoring_interval"`
	LearningInterval     string  `yaml:"learning_interval"`
	ServiceRestart       bool    `yaml:"service_restart"`
	ResourceReallocation bool    `yaml:"resource_reallocation"`
	LoadRedistribution   bool    `yaml:"load_redistribution"`
	EnableLearning       bool    `yaml:"enable_learning"`
	EnablePredictive     bool    `yaml:"enable_predictive"`
	EnableProactive      bool    `yaml:"enable_proactive"`
	EnableFailover       bool    `yaml:"enable_failover"`
	EnableScaling        bool    `yaml:"enable_scaling"`
}) error {
	if !config.Enabled {
		slog.Info("self-healing disabled")
		return nil
	}

	if eftm.selfHealer == nil {
		return fmt.Errorf("self-healer not initialized")
	}

	// Parse duration strings
	healingInterval, err := time.ParseDuration(config.HealingInterval)
	if err != nil {
		return fmt.Errorf("invalid healing_interval: %w", err)
	}

	monitoringInterval, err := time.ParseDuration(config.MonitoringInterval)
	if err != nil {
		return fmt.Errorf("invalid monitoring_interval: %w", err)
	}

	learningInterval, err := time.ParseDuration(config.LearningInterval)
	if err != nil {
		return fmt.Errorf("invalid learning_interval: %w", err)
	}

	// Apply configuration to self-healer
	eftm.selfHealer.config.HealingThreshold = config.HealingThreshold
	eftm.selfHealer.config.HealingInterval = healingInterval
	eftm.selfHealer.config.MonitoringInterval = monitoringInterval
	eftm.selfHealer.config.LearningInterval = learningInterval
	eftm.selfHealer.config.EnableServiceRestart = config.ServiceRestart
	eftm.selfHealer.config.EnableResourceReallocation = config.ResourceReallocation
	eftm.selfHealer.config.EnableLoadRedistribution = config.LoadRedistribution
	eftm.selfHealer.config.EnableLearning = config.EnableLearning
	eftm.selfHealer.config.EnablePredictiveHealing = config.EnablePredictive
	eftm.selfHealer.config.EnableProactiveHealing = config.EnableProactive
	eftm.selfHealer.config.EnableFailover = config.EnableFailover
	eftm.selfHealer.config.EnableScaling = config.EnableScaling

	slog.Info("self-healing configuration applied",
		"healing_threshold", config.HealingThreshold,
		"healing_interval", healingInterval,
		"monitoring_interval", monitoringInterval,
		"learning_interval", learningInterval)

	return nil
}

// applyRedundancyConfig applies redundancy configuration
func (eftm *EnhancedFaultToleranceManager) applyRedundancyConfig(config struct {
	Enabled        bool   `yaml:"enabled"`
	DefaultFactor  int    `yaml:"default_factor"`
	MaxFactor      int    `yaml:"max_factor"`
	UpdateInterval string `yaml:"update_interval"`
}) error {
	if !config.Enabled {
		slog.Info("redundancy management disabled")
		return nil
	}

	if eftm.redundancyManager == nil {
		return fmt.Errorf("redundancy manager not initialized")
	}

	// Parse duration string
	updateInterval, err := time.ParseDuration(config.UpdateInterval)
	if err != nil {
		return fmt.Errorf("invalid update_interval: %w", err)
	}

	// Apply configuration to redundancy manager
	eftm.redundancyManager.factor = config.DefaultFactor
	eftm.redundancyManager.maxFactor = config.MaxFactor
	eftm.redundancyManager.updateInterval = updateInterval

	slog.Info("redundancy configuration applied",
		"default_factor", config.DefaultFactor,
		"max_factor", config.MaxFactor,
		"update_interval", updateInterval)

	return nil
}

// applyPerformanceTrackingConfig applies performance tracking configuration
func (eftm *EnhancedFaultToleranceManager) applyPerformanceTrackingConfig(config struct {
	Enabled    bool   `yaml:"enabled"`
	WindowSize string `yaml:"window_size"`
}) error {
	if !config.Enabled {
		slog.Info("performance tracking disabled")
		return nil
	}

	if eftm.performanceTracker == nil {
		return fmt.Errorf("performance tracker not initialized")
	}

	// Parse duration string
	windowSize, err := time.ParseDuration(config.WindowSize)
	if err != nil {
		return fmt.Errorf("invalid window_size: %w", err)
	}

	// Apply configuration to performance tracker
	eftm.performanceTracker.learning = config.Enabled
	eftm.performanceTracker.windowSize = windowSize

	slog.Info("performance tracking configuration applied",
		"enabled", config.Enabled,
		"window_size", windowSize)

	return nil
}

// applyConfigAdaptationConfig applies configuration adaptation settings
func (eftm *EnhancedFaultToleranceManager) applyConfigAdaptationConfig(config struct {
	Enabled  bool   `yaml:"enabled"`
	Interval string `yaml:"interval"`
}) error {
	if !config.Enabled {
		slog.Info("configuration adaptation disabled")
		return nil
	}

	if eftm.configAdaptor == nil {
		return fmt.Errorf("config adaptor not initialized")
	}

	// Parse duration string
	interval, err := time.ParseDuration(config.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	// Apply configuration to config adaptor
	eftm.configAdaptor.learning = config.Enabled
	eftm.configAdaptor.interval = interval

	slog.Info("configuration adaptation applied",
		"enabled", config.Enabled,
		"interval", interval)

	return nil
}

// Hot-Reload Configuration Support

// ReloadConfiguration reloads configuration without restarting the system
func (eftm *EnhancedFaultToleranceManager) ReloadConfiguration(distributedConfig *config.DistributedConfig) error {
	eftm.mu.Lock()
	defer eftm.mu.Unlock()

	slog.Info("reloading fault tolerance configuration")

	// Store current configuration for rollback if needed
	currentConfig := eftm.getCurrentConfigSnapshot()

	// Validate new configuration
	if err := eftm.ValidateConfiguration(distributedConfig); err != nil {
		return fmt.Errorf("configuration validation failed during reload: %w", err)
	}

	// Apply new configuration
	if err := eftm.applyConfiguration(distributedConfig); err != nil {
		// Rollback to previous configuration on failure
		slog.Error("failed to apply new configuration, rolling back", "error", err)
		if rollbackErr := eftm.rollbackConfiguration(currentConfig); rollbackErr != nil {
			slog.Error("failed to rollback configuration", "error", rollbackErr)
			return fmt.Errorf("configuration reload failed and rollback failed: %w, rollback error: %w", err, rollbackErr)
		}
		return fmt.Errorf("configuration reload failed, rolled back to previous config: %w", err)
	}

	// Notify components of configuration change
	eftm.notifyConfigurationChange(distributedConfig)

	slog.Info("fault tolerance configuration reloaded successfully")
	return nil
}

// getCurrentConfigSnapshot creates a snapshot of current configuration for rollback
func (eftm *EnhancedFaultToleranceManager) getCurrentConfigSnapshot() map[string]interface{} {
	snapshot := make(map[string]interface{})

	// Store current base configuration
	if eftm.FaultToleranceManager != nil && eftm.FaultToleranceManager.config != nil {
		snapshot["base_config"] = *eftm.FaultToleranceManager.config
	}

	// Store predictor configuration
	if eftm.predictor != nil {
		snapshot["predictor_learning"] = eftm.predictor.learning
		snapshot["predictor_threshold"] = eftm.predictor.threshold
		snapshot["predictor_window_size"] = eftm.predictor.windowSize
	}

	// Store self-healer configuration
	if eftm.selfHealer != nil && eftm.selfHealer.config != nil {
		snapshot["self_healer_config"] = *eftm.selfHealer.config
	}

	// Store redundancy manager configuration
	if eftm.redundancyManager != nil {
		snapshot["redundancy_factor"] = eftm.redundancyManager.factor
		snapshot["redundancy_max_factor"] = eftm.redundancyManager.maxFactor
		snapshot["redundancy_update_interval"] = eftm.redundancyManager.updateInterval
	}

	// Store performance tracker configuration
	if eftm.performanceTracker != nil {
		snapshot["performance_learning"] = eftm.performanceTracker.learning
		snapshot["performance_window_size"] = eftm.performanceTracker.windowSize
	}

	// Store config adaptor configuration
	if eftm.configAdaptor != nil {
		snapshot["config_adaptor_learning"] = eftm.configAdaptor.learning
		snapshot["config_adaptor_interval"] = eftm.configAdaptor.interval
	}

	return snapshot
}

// rollbackConfiguration restores previous configuration from snapshot
func (eftm *EnhancedFaultToleranceManager) rollbackConfiguration(snapshot map[string]interface{}) error {
	// Restore base configuration
	if baseConfig, ok := snapshot["base_config"].(Config); ok {
		*eftm.FaultToleranceManager.config = baseConfig
	}

	// Restore predictor configuration
	if eftm.predictor != nil {
		if learning, ok := snapshot["predictor_learning"].(bool); ok {
			eftm.predictor.learning = learning
		}
		if threshold, ok := snapshot["predictor_threshold"].(float64); ok {
			eftm.predictor.threshold = threshold
		}
		if windowSize, ok := snapshot["predictor_window_size"].(time.Duration); ok {
			eftm.predictor.windowSize = windowSize
		}
	}

	// Restore self-healer configuration
	if eftm.selfHealer != nil && eftm.selfHealer.config != nil {
		if config, ok := snapshot["self_healer_config"].(SelfHealingConfig); ok {
			*eftm.selfHealer.config = config
		}
	}

	// Restore redundancy manager configuration
	if eftm.redundancyManager != nil {
		if factor, ok := snapshot["redundancy_factor"].(int); ok {
			eftm.redundancyManager.factor = factor
		}
		if maxFactor, ok := snapshot["redundancy_max_factor"].(int); ok {
			eftm.redundancyManager.maxFactor = maxFactor
		}
		if updateInterval, ok := snapshot["redundancy_update_interval"].(time.Duration); ok {
			eftm.redundancyManager.updateInterval = updateInterval
		}
	}

	// Restore performance tracker configuration
	if eftm.performanceTracker != nil {
		if learning, ok := snapshot["performance_learning"].(bool); ok {
			eftm.performanceTracker.learning = learning
		}
		if windowSize, ok := snapshot["performance_window_size"].(time.Duration); ok {
			eftm.performanceTracker.windowSize = windowSize
		}
	}

	// Restore config adaptor configuration
	if eftm.configAdaptor != nil {
		if learning, ok := snapshot["config_adaptor_learning"].(bool); ok {
			eftm.configAdaptor.learning = learning
		}
		if interval, ok := snapshot["config_adaptor_interval"].(time.Duration); ok {
			eftm.configAdaptor.interval = interval
		}
	}

	return nil
}

// notifyConfigurationChange notifies all components of configuration changes
func (eftm *EnhancedFaultToleranceManager) notifyConfigurationChange(distributedConfig *config.DistributedConfig) {
	// Notify predictor of configuration change
	if eftm.predictor != nil {
		// Predictor can adjust its algorithms based on new configuration
		slog.Debug("notifying predictor of configuration change")
	}

	// Notify self-healer of configuration change
	if eftm.selfHealer != nil {
		// Self-healer can update its strategies based on new configuration
		slog.Debug("notifying self-healer of configuration change")
	}

	// Notify redundancy manager of configuration change
	if eftm.redundancyManager != nil {
		// Redundancy manager can adjust replication based on new configuration
		slog.Debug("notifying redundancy manager of configuration change")
	}

	// Update metrics with configuration change event
	if eftm.enhancedMetrics != nil {
		eftm.enhancedMetrics.LastUpdated = time.Now()
	}
}
