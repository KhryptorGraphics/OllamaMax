package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ReplicationEngine manages data replication across storage nodes
type ReplicationEngine struct {
	logger *slog.Logger
	config *ReplicationConfig

	// Node management
	nodeManager *NodeManager

	// Replication coordination
	coordinator *ReplicationCoordinator

	// Replication strategies
	strategies map[string]ReplicationStrategy

	// Policies
	policies      map[string]*ReplicationPolicy
	policiesMutex sync.RWMutex

	// Metrics and monitoring
	metrics      *ReplicationMetrics
	metricsMutex sync.RWMutex

	// Background task control
	ctx    context.Context
	cancel context.CancelFunc
}

// ReplicationConfig contains replication engine configuration
type ReplicationConfig struct {
	// Basic settings
	DefaultReplicationFactor int           `json:"default_replication_factor"`
	MaxConcurrentSyncs       int           `json:"max_concurrent_syncs"`
	SyncTimeout              time.Duration `json:"sync_timeout"`
	RetryInterval            time.Duration `json:"retry_interval"`
	MaxRetries               int           `json:"max_retries"`

	// Strategy settings
	DefaultStrategy     string `json:"default_strategy"`
	EnableEagerStrategy bool   `json:"enable_eager_strategy"`
	EnableLazyStrategy  bool   `json:"enable_lazy_strategy"`
	EnableGeoStrategy   bool   `json:"enable_geo_strategy"`

	// Health and monitoring
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
	EnableMetrics       bool          `json:"enable_metrics"`

	// Performance tuning
	BatchSize          int           `json:"batch_size"`
	CompressionEnabled bool          `json:"compression_enabled"`
	EncryptionEnabled  bool          `json:"encryption_enabled"`
	ChecksumValidation bool          `json:"checksum_validation"`
	TransferTimeout    time.Duration `json:"transfer_timeout"`
}

// Note: ReplicationPolicy is defined in interface.go

// ReplicationMetrics tracks replication performance and health
type ReplicationMetrics struct {
	// Operation counts
	TotalOperations int64 `json:"total_operations"`
	SuccessfulOps   int64 `json:"successful_ops"`
	FailedOps       int64 `json:"failed_ops"`
	PendingOps      int64 `json:"pending_ops"`

	// Performance metrics
	AverageLatency    time.Duration `json:"average_latency"`
	ThroughputMBps    float64       `json:"throughput_mbps"`
	BytesReplicated   int64         `json:"bytes_replicated"`
	ObjectsReplicated int64         `json:"objects_replicated"`

	// Health metrics
	HealthyNodes      int     `json:"healthy_nodes"`
	TotalNodes        int     `json:"total_nodes"`
	ReplicationHealth float64 `json:"replication_health"`
	ConsistencyRatio  float64 `json:"consistency_ratio"`

	// Timing
	LastUpdate         time.Time `json:"last_update"`
	LastSuccessfulSync time.Time `json:"last_successful_sync"`
	LastFailedSync     time.Time `json:"last_failed_sync"`
}

// NewReplicationEngine creates a new replication engine
func NewReplicationEngine(config *ReplicationConfig, logger *slog.Logger) (*ReplicationEngine, error) {
	if config == nil {
		config = DefaultReplicationConfig()
	}

	re := &ReplicationEngine{
		logger:     logger,
		config:     config,
		strategies: make(map[string]ReplicationStrategy),
		policies:   make(map[string]*ReplicationPolicy),
		metrics:    NewReplicationMetrics(),
	}

	// Initialize node manager
	nodeConfig := &NodeManagerConfig{
		HeartbeatInterval:   config.HealthCheckInterval,
		FailureTimeout:      config.SyncTimeout,
		MaxFailures:         config.MaxRetries,
		EnableLoadBalancing: true,
		PreferLocalReplicas: true,
	}

	var err error
	re.nodeManager, err = NewNodeManager(nodeConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create node manager: %w", err)
	}

	// Initialize coordinator
	re.coordinator = NewReplicationCoordinator(re, logger)

	// Initialize strategies
	re.initializeStrategies()

	// Set default policy
	re.policies["default"] = GetDefaultReplicationPolicy()

	return re, nil
}

// DefaultReplicationConfig returns default replication configuration
func DefaultReplicationConfig() *ReplicationConfig {
	return &ReplicationConfig{
		DefaultReplicationFactor: 3,
		MaxConcurrentSyncs:       10,
		SyncTimeout:              30 * time.Second,
		RetryInterval:            5 * time.Second,
		MaxRetries:               3,
		DefaultStrategy:          "eager",
		EnableEagerStrategy:      true,
		EnableLazyStrategy:       true,
		EnableGeoStrategy:        true,
		HealthCheckInterval:      30 * time.Second,
		MetricsInterval:          60 * time.Second,
		EnableMetrics:            true,
		BatchSize:                100,
		CompressionEnabled:       true,
		EncryptionEnabled:        true,
		ChecksumValidation:       true,
		TransferTimeout:          5 * time.Minute,
	}
}

// Start starts the replication engine
func (re *ReplicationEngine) Start(ctx context.Context) error {
	re.ctx, re.cancel = context.WithCancel(ctx)

	// Start node manager
	if err := re.nodeManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start node manager: %w", err)
	}

	// Start coordinator
	if err := re.coordinator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start coordinator: %w", err)
	}

	// Start background tasks
	go re.metricsCollectionRoutine()
	go re.healthMonitoringRoutine()

	re.logger.Info("replication engine started")
	return nil
}

// Stop stops the replication engine
func (re *ReplicationEngine) Stop(ctx context.Context) error {
	re.cancel()

	// Stop coordinator
	if err := re.coordinator.Stop(ctx); err != nil {
		re.logger.Warn("error stopping coordinator", "error", err)
	}

	// Stop node manager
	if err := re.nodeManager.Stop(ctx); err != nil {
		re.logger.Warn("error stopping node manager", "error", err)
	}

	re.logger.Info("replication engine stopped")
	return nil
}

// Replicate replicates an object to target nodes
func (re *ReplicationEngine) Replicate(ctx context.Context, key string, metadata *ObjectMetadata, policyName string) error {
	// Get replication policy
	policy, err := re.GetPolicy(policyName)
	if err != nil {
		policy = re.policies["default"]
	}

	// Get replication strategy
	strategy, exists := re.strategies[policy.Strategy]
	if !exists {
		strategy = re.strategies[re.config.DefaultStrategy]
	}

	// Check if replication is needed
	if !strategy.ShouldReplicate(key, metadata, policy) {
		return nil
	}

	// Get available nodes
	availableNodes := re.nodeManager.GetHealthyNodes()
	if len(availableNodes) == 0 {
		return fmt.Errorf("no healthy nodes available for replication")
	}

	// Select target nodes
	targetNodes, err := strategy.SelectTargetNodes("", availableNodes, policy)
	if err != nil {
		return fmt.Errorf("failed to select target nodes: %w", err)
	}

	// Create replication operation
	operation := NewReplicationOperation("replicate", key, "", extractNodeIDs(targetNodes), policy)

	// Submit operation
	if err := re.coordinator.SubmitReplication(operation); err != nil {
		return fmt.Errorf("failed to submit replication: %w", err)
	}

	// Wait for completion if synchronous
	if strategy.GetConsistencyLevel() == "strong" {
		select {
		case err := <-operation.ResultChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(re.config.SyncTimeout):
			return fmt.Errorf("replication timeout")
		}
	}

	return nil
}

// AddNode adds a storage node to the cluster
func (re *ReplicationEngine) AddNode(ctx context.Context, node *StorageNode) error {
	return re.nodeManager.AddNode(ctx, node)
}

// RemoveNode removes a storage node from the cluster
func (re *ReplicationEngine) RemoveNode(ctx context.Context, nodeID string) error {
	return re.nodeManager.RemoveNode(ctx, nodeID)
}

// GetNodes returns all storage nodes
func (re *ReplicationEngine) GetNodes() []*StorageNode {
	return re.nodeManager.GetAllNodes()
}

// GetHealthyNodes returns only healthy storage nodes
func (re *ReplicationEngine) GetHealthyNodes() []*StorageNode {
	return re.nodeManager.GetHealthyNodes()
}

// SetPolicy sets a replication policy
func (re *ReplicationEngine) SetPolicy(name string, policy *ReplicationPolicy) error {
	if err := ValidateReplicationPolicy(policy); err != nil {
		return err
	}

	re.policiesMutex.Lock()
	defer re.policiesMutex.Unlock()

	re.policies[name] = policy
	return nil
}

// GetPolicy gets a replication policy
func (re *ReplicationEngine) GetPolicy(name string) (*ReplicationPolicy, error) {
	re.policiesMutex.RLock()
	defer re.policiesMutex.RUnlock()

	policy, exists := re.policies[name]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", name)
	}

	return policy, nil
}

// GetMetrics returns current replication metrics
func (re *ReplicationEngine) GetMetrics() *ReplicationMetrics {
	re.metricsMutex.RLock()
	defer re.metricsMutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *re.metrics
	return &metrics
}

// initializeStrategies initializes replication strategies
func (re *ReplicationEngine) initializeStrategies() {
	if re.config.EnableEagerStrategy {
		re.strategies["eager"] = NewEagerReplicationStrategy(re.config)
	}

	if re.config.EnableLazyStrategy {
		re.strategies["lazy"] = NewLazyReplicationStrategy(re.config)
	}

	if re.config.EnableGeoStrategy {
		re.strategies["geographic"] = NewGeographicReplicationStrategy(re.config)
	}
}

// Background routines

func (re *ReplicationEngine) metricsCollectionRoutine() {
	if !re.config.EnableMetrics {
		return
	}

	ticker := time.NewTicker(re.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-re.ctx.Done():
			return
		case <-ticker.C:
			re.collectMetrics()
		}
	}
}

func (re *ReplicationEngine) healthMonitoringRoutine() {
	ticker := time.NewTicker(re.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-re.ctx.Done():
			return
		case <-ticker.C:
			re.monitorHealth()
		}
	}
}

func (re *ReplicationEngine) collectMetrics() {
	re.metricsMutex.Lock()
	defer re.metricsMutex.Unlock()

	// Update node counts
	allNodes := re.nodeManager.GetAllNodes()
	healthyNodes := re.nodeManager.GetHealthyNodes()

	re.metrics.TotalNodes = len(allNodes)
	re.metrics.HealthyNodes = len(healthyNodes)

	if re.metrics.TotalNodes > 0 {
		re.metrics.ReplicationHealth = float64(re.metrics.HealthyNodes) / float64(re.metrics.TotalNodes)
	}

	re.metrics.LastUpdate = time.Now()
}

func (re *ReplicationEngine) monitorHealth() {
	// Health monitoring is handled by the node manager
	// This routine can be used for additional health checks
}

// NewReplicationMetrics creates a new replication metrics instance
func NewReplicationMetrics() *ReplicationMetrics {
	return &ReplicationMetrics{
		LastUpdate: time.Now(),
	}
}
