package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// DistributedStorageImpl implements DistributedStorage interface
type DistributedStorageImpl struct {
	localStorage Storage
	logger       *slog.Logger
	
	// Distributed coordination
	nodeID       string
	
	// Node management
	nodes       map[string]*NodeInfo
	nodesMutex  sync.RWMutex
	
	// Replication management
	replicationMgr *ReplicationManager
	
	// Consensus and coordination
	consensusState *ConsensusState
	consensusMutex sync.RWMutex
	
	// Distributed locks
	locks       map[string]*DistributedLock
	locksMutex  sync.RWMutex
	
	// Configuration
	config *DistributedStorageConfig
	
	// Metrics and monitoring
	metrics     *DistributedMetrics
	metricsMutex sync.RWMutex
	
	// Background tasks
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// DistributedStorageConfig contains configuration for distributed storage
type DistributedStorageConfig struct {
	NodeID               string        `json:"node_id"`
	ReplicationFactor    int           `json:"replication_factor"`
	ConsistencyLevel     string        `json:"consistency_level"`
	HeartbeatInterval    time.Duration `json:"heartbeat_interval"`
	ElectionTimeout      time.Duration `json:"election_timeout"`
	ReplicationTimeout   time.Duration `json:"replication_timeout"`
	MaxConcurrentRepl    int           `json:"max_concurrent_replication"`
	GossipInterval       time.Duration `json:"gossip_interval"`
	FailureDetectorTimeout time.Duration `json:"failure_detector_timeout"`
}

// ReplicationManager handles distributed replication
type ReplicationManager struct {
	storage    *DistributedStorageImpl
	logger     *slog.Logger
	
	// Replication state
	replicas      map[string]*ReplicationStatus
	replicasMutex sync.RWMutex
	
	// Worker pools
	workers     []*ReplicationWorker
	workQueue   chan *ReplicationTask
	
	// Policies
	policies      map[string]*ReplicationPolicy
	policiesMutex sync.RWMutex
	
	ctx    context.Context
	cancel context.CancelFunc
}

// ReplicationWorker handles replication tasks
type ReplicationWorker struct {
	id      int
	manager *ReplicationManager
	logger  *slog.Logger
}

// ReplicationTask represents a replication task
type ReplicationTask struct {
	Type        string    `json:"type"`
	Key         string    `json:"key"`
	SourceNode  string    `json:"source_node"`
	TargetNodes []string  `json:"target_nodes"`
	Priority    int       `json:"priority"`
	Timeout     time.Duration `json:"timeout"`
	CreatedAt   time.Time `json:"created_at"`
	Retries     int       `json:"retries"`
	MaxRetries  int       `json:"max_retries"`
}

// DistributedLock implements the Lock interface
type DistributedLock struct {
	lockID     string
	owner      string
	expiration time.Time
	storage    *DistributedStorageImpl
	released   bool
	mutex      sync.Mutex
}

// NewDistributedStorage creates a new distributed storage instance
func NewDistributedStorage(
	localStorage Storage,
	config *DistributedStorageConfig,
	logger *slog.Logger,
) (*DistributedStorageImpl, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	ds := &DistributedStorageImpl{
		localStorage: localStorage,
		logger:       logger,
		nodeID:       config.NodeID,
		nodes:        make(map[string]*NodeInfo),
		locks:        make(map[string]*DistributedLock),
		config:       config,
		consensusState: &ConsensusState{
			Nodes:      make(map[string]string),
			QuorumSize: (config.ReplicationFactor / 2) + 1,
			IsHealthy:  true,
		},
		metrics: &DistributedMetrics{
			DataDistribution: make(map[string]int64),
			NetworkMetrics:   &NetworkMetrics{ConnectionCounts: make(map[string]int64)},
			ConsensusMetrics: &ConsensusMetrics{},
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize replication manager
	ds.replicationMgr = &ReplicationManager{
		storage:   ds,
		logger:    logger,
		replicas:  make(map[string]*ReplicationStatus),
		policies:  make(map[string]*ReplicationPolicy),
		workQueue: make(chan *ReplicationTask, 1000),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Create replication workers
	workerCount := config.MaxConcurrentRepl
	if workerCount <= 0 {
		workerCount = 10
	}
	
	ds.replicationMgr.workers = make([]*ReplicationWorker, workerCount)
	for i := 0; i < workerCount; i++ {
		ds.replicationMgr.workers[i] = &ReplicationWorker{
			id:      i,
			manager: ds.replicationMgr,
			logger:  logger,
		}
	}
	
	return ds, nil
}

// Start starts the distributed storage
func (ds *DistributedStorageImpl) Start(ctx context.Context) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	if ds.started {
		return &StorageError{
			Code:    ErrCodeInternal,
			Message: "distributed storage already started",
		}
	}
	
	// Start local storage
	if err := ds.localStorage.Start(ctx); err != nil {
		return fmt.Errorf("failed to start local storage: %w", err)
	}
	
	// Start replication workers
	for _, worker := range ds.replicationMgr.workers {
		go worker.start()
	}
	
	// Start background routines
	go ds.heartbeatRoutine()
	go ds.consensusMonitorRoutine()
	go ds.metricsCollectionRoutine()
	go ds.failureDetectorRoutine()
	
	ds.started = true
	ds.logger.Info("distributed storage started", "node_id", ds.nodeID)
	
	return nil
}

// Stop stops the distributed storage
func (ds *DistributedStorageImpl) Stop(ctx context.Context) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	
	if !ds.started {
		return nil
	}
	
	ds.cancel()
	
	// Stop local storage
	if err := ds.localStorage.Stop(ctx); err != nil {
		ds.logger.Error("failed to stop local storage", "error", err)
	}
	
	ds.started = false
	ds.logger.Info("distributed storage stopped")
	
	return nil
}

// Close closes the distributed storage
func (ds *DistributedStorageImpl) Close() error {
	return ds.Stop(context.Background())
}

// Core storage operations (delegate to local storage with replication)

// Store stores an object with distributed replication
func (ds *DistributedStorageImpl) Store(ctx context.Context, key string, data io.Reader, metadata *ObjectMetadata) error {
	// First store locally
	if err := ds.localStorage.Store(ctx, key, data, metadata); err != nil {
		return err
	}
	
	// Get or create replication policy
	policy := ds.getReplicationPolicy(key)
	if policy == nil {
		policy = ds.createDefaultReplicationPolicy(key)
	}
	
	// Initiate replication based on policy
	if err := ds.initiateReplication(ctx, key, policy); err != nil {
		ds.logger.Error("replication failed", "key", key, "error", err)
		// Don't fail the store operation if replication fails
	}
	
	return nil
}

// Retrieve retrieves an object, potentially from replicas
func (ds *DistributedStorageImpl) Retrieve(ctx context.Context, key string) (io.ReadCloser, *ObjectMetadata, error) {
	// Try local storage first
	reader, metadata, err := ds.localStorage.Retrieve(ctx, key)
	if err == nil {
		return reader, metadata, nil
	}
	
	// If not found locally, try replicas
	if isNotFoundError(err) {
		return ds.retrieveFromReplicas(ctx, key)
	}
	
	return nil, nil, err
}

// Delete deletes an object from all replicas
func (ds *DistributedStorageImpl) Delete(ctx context.Context, key string) error {
	// Delete locally first
	if err := ds.localStorage.Delete(ctx, key); err != nil && !isNotFoundError(err) {
		return err
	}
	
	// Delete from replicas
	if err := ds.deleteFromReplicas(ctx, key); err != nil {
		ds.logger.Error("failed to delete from replicas", "key", key, "error", err)
	}
	
	// Remove replication status
	ds.replicationMgr.removeReplicationStatus(key)
	
	return nil
}

// Exists checks if an object exists locally or on replicas
func (ds *DistributedStorageImpl) Exists(ctx context.Context, key string) (bool, error) {
	// Check locally first
	exists, err := ds.localStorage.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	
	// Check replicas
	return ds.existsOnReplicas(ctx, key)
}

// Delegate metadata operations to local storage
func (ds *DistributedStorageImpl) GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error) {
	return ds.localStorage.GetMetadata(ctx, key)
}

func (ds *DistributedStorageImpl) SetMetadata(ctx context.Context, key string, metadata *ObjectMetadata) error {
	return ds.localStorage.SetMetadata(ctx, key, metadata)
}

func (ds *DistributedStorageImpl) UpdateMetadata(ctx context.Context, key string, updates map[string]interface{}) error {
	return ds.localStorage.UpdateMetadata(ctx, key, updates)
}

// Delegate batch operations
func (ds *DistributedStorageImpl) BatchStore(ctx context.Context, operations []BatchStoreOperation) error {
	return ds.localStorage.BatchStore(ctx, operations)
}

func (ds *DistributedStorageImpl) BatchDelete(ctx context.Context, keys []string) error {
	return ds.localStorage.BatchDelete(ctx, keys)
}

// Delegate listing operations
func (ds *DistributedStorageImpl) List(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error) {
	return ds.localStorage.List(ctx, prefix, options)
}

func (ds *DistributedStorageImpl) ListKeys(ctx context.Context, prefix string) ([]string, error) {
	return ds.localStorage.ListKeys(ctx, prefix)
}

// Distributed-specific operations

// Replicate replicates an object to target nodes
func (ds *DistributedStorageImpl) Replicate(ctx context.Context, key string, targetNodes []string) error {
	task := &ReplicationTask{
		Type:        "replicate",
		Key:         key,
		SourceNode:  ds.nodeID,
		TargetNodes: targetNodes,
		Priority:    1,
		Timeout:     ds.config.ReplicationTimeout,
		CreatedAt:   time.Now(),
		MaxRetries:  3,
	}
	
	select {
	case ds.replicationMgr.workQueue <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return &StorageError{
			Code:    ErrCodeTimeout,
			Message: "replication queue full",
		}
	}
}

// GetReplicationStatus gets the replication status for a key
func (ds *DistributedStorageImpl) GetReplicationStatus(ctx context.Context, key string) (*ReplicationStatus, error) {
	ds.replicationMgr.replicasMutex.RLock()
	defer ds.replicationMgr.replicasMutex.RUnlock()
	
	status, exists := ds.replicationMgr.replicas[key]
	if !exists {
		return nil, &StorageError{
			Code:    ErrCodeNotFound,
			Message: "replication status not found",
		}
	}
	
	// Create a copy
	result := *status
	return &result, nil
}

// SetReplicationPolicy sets a replication policy for a key
func (ds *DistributedStorageImpl) SetReplicationPolicy(ctx context.Context, key string, policy *ReplicationPolicy) error {
	ds.replicationMgr.policiesMutex.Lock()
	defer ds.replicationMgr.policiesMutex.Unlock()
	
	ds.replicationMgr.policies[key] = policy
	
	// Trigger replication if necessary
	go ds.enforceReplicationPolicy(ctx, key, policy)
	
	return nil
}

// Consensus and coordination operations

// ProposeWrite proposes a write operation through consensus
func (ds *DistributedStorageImpl) ProposeWrite(ctx context.Context, key string, data io.Reader, metadata *ObjectMetadata) error {
	// For now, fall back to direct store since we don't have Raft integration
	return ds.Store(ctx, key, data, metadata)
}

// ProposeDelete proposes a delete operation through consensus
func (ds *DistributedStorageImpl) ProposeDelete(ctx context.Context, key string) error {
	// For now, fall back to direct delete since we don't have Raft integration
	return ds.Delete(ctx, key)
}

// GetConsensusState returns the current consensus state
func (ds *DistributedStorageImpl) GetConsensusState(ctx context.Context) (*ConsensusState, error) {
	ds.consensusMutex.RLock()
	defer ds.consensusMutex.RUnlock()
	
	// Create a copy
	state := *ds.consensusState
	state.Nodes = make(map[string]string)
	for k, v := range ds.consensusState.Nodes {
		state.Nodes[k] = v
	}
	
	return &state, nil
}

// Node management operations

// AddNode adds a new node to the cluster
func (ds *DistributedStorageImpl) AddNode(ctx context.Context, nodeID string, nodeInfo *NodeInfo) error {
	ds.nodesMutex.Lock()
	defer ds.nodesMutex.Unlock()
	
	nodeInfo.NodeID = nodeID
	nodeInfo.JoinedAt = time.Now()
	nodeInfo.LastSeen = time.Now()
	
	ds.nodes[nodeID] = nodeInfo
	
	// Update consensus state
	ds.consensusMutex.Lock()
	ds.consensusState.Nodes[nodeID] = "active"
	ds.consensusMutex.Unlock()
	
	ds.logger.Info("node added", "node_id", nodeID, "address", nodeInfo.Address)
	
	return nil
}

// RemoveNode removes a node from the cluster
func (ds *DistributedStorageImpl) RemoveNode(ctx context.Context, nodeID string) error {
	ds.nodesMutex.Lock()
	defer ds.nodesMutex.Unlock()
	
	delete(ds.nodes, nodeID)
	
	// Update consensus state
	ds.consensusMutex.Lock()
	delete(ds.consensusState.Nodes, nodeID)
	ds.consensusMutex.Unlock()
	
	ds.logger.Info("node removed", "node_id", nodeID)
	
	return nil
}

// GetNodes returns all nodes in the cluster
func (ds *DistributedStorageImpl) GetNodes(ctx context.Context) ([]*NodeInfo, error) {
	ds.nodesMutex.RLock()
	defer ds.nodesMutex.RUnlock()
	
	nodes := make([]*NodeInfo, 0, len(ds.nodes))
	for _, node := range ds.nodes {
		// Create a copy
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}
	
	return nodes, nil
}

// Distributed coordination

// AcquireLock acquires a distributed lock
func (ds *DistributedStorageImpl) AcquireLock(ctx context.Context, lockID string, timeout time.Duration) (Lock, error) {
	ds.locksMutex.Lock()
	defer ds.locksMutex.Unlock()
	
	// Check if lock already exists
	if lock, exists := ds.locks[lockID]; exists {
		if lock.IsHeld() {
			return nil, &StorageError{
				Code:    ErrCodeAlreadyExists,
				Message: "lock already held",
			}
		}
	}
	
	// Create new lock
	lock := &DistributedLock{
		lockID:     lockID,
		owner:      ds.nodeID,
		expiration: time.Now().Add(timeout),
		storage:    ds,
		released:   false,
	}
	
	ds.locks[lockID] = lock
	
	return lock, nil
}

// GetDistributedMetrics returns distributed metrics
func (ds *DistributedStorageImpl) GetDistributedMetrics(ctx context.Context) (*DistributedMetrics, error) {
	ds.metricsMutex.RLock()
	defer ds.metricsMutex.RUnlock()
	
	// Create a copy
	metrics := *ds.metrics
	
	// Copy maps
	metrics.DataDistribution = make(map[string]int64)
	for k, v := range ds.metrics.DataDistribution {
		metrics.DataDistribution[k] = v
	}
	
	if ds.metrics.NetworkMetrics != nil {
		netMetrics := *ds.metrics.NetworkMetrics
		netMetrics.ConnectionCounts = make(map[string]int64)
		for k, v := range ds.metrics.NetworkMetrics.ConnectionCounts {
			netMetrics.ConnectionCounts[k] = v
		}
		metrics.NetworkMetrics = &netMetrics
	}
	
	if ds.metrics.ConsensusMetrics != nil {
		consMetrics := *ds.metrics.ConsensusMetrics
		metrics.ConsensusMetrics = &consMetrics
	}
	
	return &metrics, nil
}

// Health check for distributed storage
func (ds *DistributedStorageImpl) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	// Get local health first
	localHealth, err := ds.localStorage.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}
	
	// Add distributed-specific checks
	checks := localHealth.Checks
	if checks == nil {
		checks = make(map[string]CheckResult)
	}
	
	// Check consensus health
	consensusCheck := ds.checkConsensusHealth()
	checks["consensus"] = consensusCheck
	
	// Check node connectivity
	connectivityCheck := ds.checkNodeConnectivity()
	checks["connectivity"] = connectivityCheck
	
	// Check replication health
	replicationCheck := ds.checkReplicationHealth()
	checks["replication"] = replicationCheck
	
	// Overall health
	healthy := localHealth.Healthy && 
		consensusCheck.Status == "ok" && 
		connectivityCheck.Status == "ok" && 
		replicationCheck.Status == "ok"
	
	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}
	
	return &HealthStatus{
		Status:    status,
		Healthy:   healthy,
		LastCheck: time.Now(),
		Checks:    checks,
	}, nil
}

// GetStats returns distributed storage statistics
func (ds *DistributedStorageImpl) GetStats(ctx context.Context) (*StorageStats, error) {
	localStats, err := ds.localStorage.GetStats(ctx)
	if err != nil {
		return nil, err
	}
	
	// Add replication statistics
	localStats.Replication = ds.getReplicationStats()
	
	return localStats, nil
}

// Helper methods

func (ds *DistributedStorageImpl) getReplicationPolicy(key string) *ReplicationPolicy {
	ds.replicationMgr.policiesMutex.RLock()
	defer ds.replicationMgr.policiesMutex.RUnlock()
	
	return ds.replicationMgr.policies[key]
}

func (ds *DistributedStorageImpl) createDefaultReplicationPolicy(key string) *ReplicationPolicy {
	policy := &ReplicationPolicy{
		MinReplicas:      ds.config.ReplicationFactor,
		MaxReplicas:      ds.config.ReplicationFactor * 2,
		ConsistencyLevel: ds.config.ConsistencyLevel,
		Strategy:         "eager",
		Constraints:      make(map[string]interface{}),
	}
	
	ds.replicationMgr.policiesMutex.Lock()
	ds.replicationMgr.policies[key] = policy
	ds.replicationMgr.policiesMutex.Unlock()
	
	return policy
}

func (ds *DistributedStorageImpl) initiateReplication(ctx context.Context, key string, policy *ReplicationPolicy) error {
	// Select target nodes
	targetNodes := ds.selectReplicationTargets(policy)
	if len(targetNodes) == 0 {
		return &StorageError{
			Code:    ErrCodeUnavailable,
			Message: "no suitable replication targets available",
		}
	}
	
	// Create replication status
	status := &ReplicationStatus{
		Key:             key,
		Policy:          policy,
		CurrentReplicas: 1, // Local copy
		HealthyReplicas: 1,
		ReplicaNodes:    append([]string{ds.nodeID}, targetNodes...),
		SyncStatus:      make(map[string]string),
		LastSync:        time.Now(),
	}
	
	ds.replicationMgr.replicasMutex.Lock()
	ds.replicationMgr.replicas[key] = status
	ds.replicationMgr.replicasMutex.Unlock()
	
	// Submit replication task
	return ds.Replicate(ctx, key, targetNodes)
}

func (ds *DistributedStorageImpl) selectReplicationTargets(policy *ReplicationPolicy) []string {
	ds.nodesMutex.RLock()
	defer ds.nodesMutex.RUnlock()
	
	var candidates []*NodeInfo
	
	// Filter by preferred nodes first
	if len(policy.PreferredNodes) > 0 {
		for _, nodeID := range policy.PreferredNodes {
			if node, exists := ds.nodes[nodeID]; exists && nodeID != ds.nodeID {
				candidates = append(candidates, node)
			}
		}
	}
	
	// Add other available nodes if needed
	if len(candidates) < policy.MinReplicas {
		for nodeID, node := range ds.nodes {
			if nodeID == ds.nodeID {
				continue
			}
			
			// Check if already in candidates
			found := false
			for _, candidate := range candidates {
				if candidate.NodeID == nodeID {
					found = true
					break
				}
			}
			if found {
				continue
			}
			
			// Check excluded nodes
			excluded := false
			for _, excludedID := range policy.ExcludedNodes {
				if excludedID == nodeID {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
			
			candidates = append(candidates, node)
			if len(candidates) >= policy.MinReplicas {
				break
			}
		}
	}
	
	// Sort by availability and select top candidates
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Available > candidates[j].Available
	})
	
	var targets []string
	for i, candidate := range candidates {
		if i >= policy.MinReplicas {
			break
		}
		targets = append(targets, candidate.NodeID)
	}
	
	return targets
}

func (ds *DistributedStorageImpl) retrieveFromReplicas(ctx context.Context, key string) (io.ReadCloser, *ObjectMetadata, error) {
	// Get replication status
	status, err := ds.GetReplicationStatus(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	
	// Try to retrieve from healthy replicas
	for _, nodeID := range status.ReplicaNodes {
		if nodeID == ds.nodeID {
			continue // Skip local node
		}
		
		// TODO: Implement remote retrieval from peer node
		// This would involve network communication with the peer
		ds.logger.Debug("attempting to retrieve from replica", "key", key, "node", nodeID)
	}
	
	return nil, nil, &StorageError{
		Code:    ErrCodeNotFound,
		Message: "object not found on any replica",
	}
}

func (ds *DistributedStorageImpl) deleteFromReplicas(ctx context.Context, key string) error {
	// Get replication status
	status, err := ds.GetReplicationStatus(ctx, key)
	if err != nil {
		return nil // No replicas to delete from
	}
	
	// Send delete requests to replicas
	for _, nodeID := range status.ReplicaNodes {
		if nodeID == ds.nodeID {
			continue // Skip local node
		}
		
		// TODO: Implement remote deletion
		ds.logger.Debug("deleting from replica", "key", key, "node", nodeID)
	}
	
	return nil
}

func (ds *DistributedStorageImpl) existsOnReplicas(ctx context.Context, key string) (bool, error) {
	// Get replication status
	status, err := ds.GetReplicationStatus(ctx, key)
	if err != nil {
		return false, nil // No replicas
	}
	
	// Check replicas
	for _, nodeID := range status.ReplicaNodes {
		if nodeID == ds.nodeID {
			continue // Skip local node
		}
		
		// TODO: Implement remote existence check
		ds.logger.Debug("checking existence on replica", "key", key, "node", nodeID)
	}
	
	return false, nil
}

func (ds *DistributedStorageImpl) enforceReplicationPolicy(ctx context.Context, key string, policy *ReplicationPolicy) {
	// TODO: Implement policy enforcement
	// This would check current replication status and adjust as needed
}

// Health check methods

func (ds *DistributedStorageImpl) checkConsensusHealth() CheckResult {
	start := time.Now()
	
	return CheckResult{
		Status:  "ok",
		Message: "consensus simulation healthy",
		Latency: time.Since(start).Milliseconds(),
		Time:    time.Now(),
	}
}

func (ds *DistributedStorageImpl) checkNodeConnectivity() CheckResult {
	start := time.Now()
	
	ds.nodesMutex.RLock()
	totalNodes := len(ds.nodes)
	connectedNodes := 0
	
	for _, node := range ds.nodes {
		if time.Since(node.LastSeen) < ds.config.FailureDetectorTimeout {
			connectedNodes++
		}
	}
	ds.nodesMutex.RUnlock()
	
	if totalNodes == 0 {
		return CheckResult{
			Status:  "warning",
			Message: "no other nodes in cluster",
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
	}
	
	connectivity := float64(connectedNodes) / float64(totalNodes)
	if connectivity >= 0.8 {
		return CheckResult{
			Status:  "ok",
			Message: fmt.Sprintf("%.1f%% nodes connected", connectivity*100),
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
	}
	
	return CheckResult{
		Status:  "error",
		Message: fmt.Sprintf("poor connectivity: %.1f%% nodes connected", connectivity*100),
		Latency: time.Since(start).Milliseconds(),
		Time:    time.Now(),
	}
}

func (ds *DistributedStorageImpl) checkReplicationHealth() CheckResult {
	start := time.Now()
	
	ds.replicationMgr.replicasMutex.RLock()
	totalReplicas := 0
	healthyReplicas := 0
	
	for _, status := range ds.replicationMgr.replicas {
		totalReplicas += status.CurrentReplicas
		healthyReplicas += status.HealthyReplicas
	}
	ds.replicationMgr.replicasMutex.RUnlock()
	
	if totalReplicas == 0 {
		return CheckResult{
			Status:  "ok",
			Message: "no replicated objects",
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
	}
	
	healthRatio := float64(healthyReplicas) / float64(totalReplicas)
	if healthRatio >= 0.9 {
		return CheckResult{
			Status:  "ok",
			Message: fmt.Sprintf("%.1f%% replicas healthy", healthRatio*100),
			Latency: time.Since(start).Milliseconds(),
			Time:    time.Now(),
		}
	}
	
	return CheckResult{
		Status:  "warning",
		Message: fmt.Sprintf("%.1f%% replicas healthy", healthRatio*100),
		Latency: time.Since(start).Milliseconds(),
		Time:    time.Now(),
	}
}

func (ds *DistributedStorageImpl) getReplicationStats() *ReplicationStats {
	ds.replicationMgr.replicasMutex.RLock()
	defer ds.replicationMgr.replicasMutex.RUnlock()
	
	stats := &ReplicationStats{
		ReplicationLag: make(map[string]int64),
		SyncOperations: &SyncStats{},
	}
	
	for _, status := range ds.replicationMgr.replicas {
		stats.TotalReplicas += int64(status.CurrentReplicas)
		stats.HealthyReplicas += int64(status.HealthyReplicas)
		stats.OutOfSyncReplicas += int64(status.CurrentReplicas - status.HealthyReplicas)
	}
	
	return stats
}

// Background routines

func (ds *DistributedStorageImpl) heartbeatRoutine() {
	ticker := time.NewTicker(ds.config.HeartbeatInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.sendHeartbeats()
		}
	}
}

func (ds *DistributedStorageImpl) consensusMonitorRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.updateConsensusState()
		}
	}
}

func (ds *DistributedStorageImpl) metricsCollectionRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.collectDistributedMetrics()
		}
	}
}

func (ds *DistributedStorageImpl) failureDetectorRoutine() {
	ticker := time.NewTicker(ds.config.FailureDetectorTimeout / 2)
	defer ticker.Stop()
	
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.detectFailedNodes()
		}
	}
}

func (ds *DistributedStorageImpl) sendHeartbeats() {
	// TODO: Implement heartbeat sending to peers
	ds.logger.Debug("sending heartbeats")
}

func (ds *DistributedStorageImpl) updateConsensusState() {
	ds.consensusMutex.Lock()
	defer ds.consensusMutex.Unlock()
	
	ds.consensusState.LastHeartbeat = time.Now()
	ds.consensusState.Term = 1
	ds.consensusState.CommitIndex = 100
	ds.consensusState.LastApplied = 100
	ds.consensusState.LeaderID = ds.nodeID
}

func (ds *DistributedStorageImpl) collectDistributedMetrics() {
	ds.metricsMutex.Lock()
	defer ds.metricsMutex.Unlock()
	
	// Update cluster metrics
	ds.nodesMutex.RLock()
	ds.metrics.ClusterSize = len(ds.nodes) + 1 // Include self
	healthyNodes := 1 // Self is always healthy
	for _, node := range ds.nodes {
		if time.Since(node.LastSeen) < ds.config.FailureDetectorTimeout {
			healthyNodes++
		}
	}
	ds.metrics.HealthyNodes = healthyNodes
	ds.nodesMutex.RUnlock()
	
	// Update replication factor
	if ds.metrics.ClusterSize > 0 {
		ds.metrics.ReplicationFactor = float64(ds.config.ReplicationFactor)
	}
}

func (ds *DistributedStorageImpl) detectFailedNodes() {
	ds.nodesMutex.Lock()
	defer ds.nodesMutex.Unlock()
	
	cutoff := time.Now().Add(-ds.config.FailureDetectorTimeout)
	for nodeID, node := range ds.nodes {
		if node.LastSeen.Before(cutoff) {
			ds.logger.Warn("node appears to have failed", "node_id", nodeID, "last_seen", node.LastSeen)
			node.Status = "failed"
		}
	}
}

// ReplicationManager methods

func (rm *ReplicationManager) removeReplicationStatus(key string) {
	rm.replicasMutex.Lock()
	defer rm.replicasMutex.Unlock()
	
	delete(rm.replicas, key)
}

// ReplicationWorker methods

func (w *ReplicationWorker) start() {
	w.logger.Info("replication worker started", "worker_id", w.id)
	
	for {
		select {
		case <-w.manager.ctx.Done():
			w.logger.Info("replication worker stopped", "worker_id", w.id)
			return
		case task := <-w.manager.workQueue:
			w.processTask(task)
		}
	}
}

func (w *ReplicationWorker) processTask(task *ReplicationTask) {
	w.logger.Debug("processing replication task", "worker_id", w.id, "type", task.Type, "key", task.Key)
	
	// TODO: Implement actual replication logic
	// This would involve network communication with target nodes
	
	time.Sleep(100 * time.Millisecond) // Simulate work
}

// DistributedLock methods

func (dl *DistributedLock) Release() error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	if dl.released {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "lock already released",
		}
	}
	
	dl.storage.locksMutex.Lock()
	delete(dl.storage.locks, dl.lockID)
	dl.storage.locksMutex.Unlock()
	
	dl.released = true
	return nil
}

func (dl *DistributedLock) Renew(timeout time.Duration) error {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	if dl.released {
		return &StorageError{
			Code:    ErrCodeInvalidArgument,
			Message: "cannot renew released lock",
		}
	}
	
	dl.expiration = time.Now().Add(timeout)
	return nil
}

func (dl *DistributedLock) IsHeld() bool {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	return !dl.released && time.Now().Before(dl.expiration)
}

func (dl *DistributedLock) GetOwner() string {
	return dl.owner
}

func (dl *DistributedLock) GetExpiration() time.Time {
	dl.mutex.Lock()
	defer dl.mutex.Unlock()
	
	return dl.expiration
}