package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// ReplicationManager manages model replication across peers
type ReplicationManager struct {
	config     *config.ReplicationConfig
	p2p        *p2p.Node
	manager    *Manager
	syncMgr    *SyncManager
	logger     *slog.Logger
	
	// Replication state
	replicas      map[string]*ReplicaInfo
	replicasMutex sync.RWMutex
	
	// Replication policies
	policies      map[string]*ReplicationPolicy
	policiesMutex sync.RWMutex
	
	// Replication workers
	workers     []*ReplicationWorker
	workQueue   chan *ReplicationTask
	
	// Health monitoring
	healthChecker *HealthChecker
	
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// ReplicaInfo contains information about a model replica
type ReplicaInfo struct {
	ModelName    string            `json:"model_name"`
	PeerID       string            `json:"peer_id"`
	Status       ReplicaStatus     `json:"status"`
	LastSync     time.Time         `json:"last_sync"`
	SyncAttempts int               `json:"sync_attempts"`
	Health       ReplicaHealth     `json:"health"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ReplicaStatus represents the status of a replica
type ReplicaStatus string

const (
	ReplicaStatusHealthy     ReplicaStatus = "healthy"
	ReplicaStatusSyncing     ReplicaStatus = "syncing"
	ReplicaStatusOutOfSync   ReplicaStatus = "out_of_sync"
	ReplicaStatusUnhealthy   ReplicaStatus = "unhealthy"
	ReplicaStatusUnreachable ReplicaStatus = "unreachable"
)

// ReplicaHealth represents the health status of a replica
type ReplicaHealth string

const (
	HealthGood    ReplicaHealth = "good"
	HealthWarning ReplicaHealth = "warning"
	HealthError   ReplicaHealth = "error"
)

// ReplicationPolicy defines how a model should be replicated
type ReplicationPolicy struct {
	ModelName         string            `json:"model_name"`
	MinReplicas       int               `json:"min_replicas"`
	MaxReplicas       int               `json:"max_replicas"`
	PreferredPeers    []string          `json:"preferred_peers"`
	ExcludedPeers     []string          `json:"excluded_peers"`
	ReplicationFactor int               `json:"replication_factor"`
	SyncInterval      time.Duration     `json:"sync_interval"`
	Priority          int               `json:"priority"`
	Constraints       map[string]string `json:"constraints"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ReplicationTask represents a replication task
type ReplicationTask struct {
	Type         TaskType    `json:"type"`
	ModelName    string      `json:"model_name"`
	SourcePeer   string      `json:"source_peer"`
	TargetPeer   string      `json:"target_peer"`
	Priority     int         `json:"priority"`
	Retries      int         `json:"retries"`
	MaxRetries   int         `json:"max_retries"`
	CreatedAt    time.Time   `json:"created_at"`
	ResponseChan chan error  `json:"-"`
}

// TaskType represents the type of replication task
type TaskType string

const (
	TaskTypeReplicate TaskType = "replicate"
	TaskTypeSync      TaskType = "sync"
	TaskTypeRemove    TaskType = "remove"
	TaskTypeVerify    TaskType = "verify"
)

// ReplicationWorker handles replication tasks
type ReplicationWorker struct {
	ID         int
	manager    *ReplicationManager
	stopChan   chan struct{}
}

// HealthChecker monitors replica health
type HealthChecker struct {
	manager       *ReplicationManager
	checkInterval time.Duration
	timeout       time.Duration
	stopChan      chan struct{}
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(
	config *config.ReplicationConfig,
	p2pNode *p2p.Node,
	manager *Manager,
	syncMgr *SyncManager,
	logger *slog.Logger,
) (*ReplicationManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	rm := &ReplicationManager{
		config:    config,
		p2p:       p2pNode,
		manager:   manager,
		syncMgr:   syncMgr,
		logger:    logger,
		replicas:  make(map[string]*ReplicaInfo),
		policies:  make(map[string]*ReplicationPolicy),
		workQueue: make(chan *ReplicationTask, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
	
	// Create workers
	rm.workers = make([]*ReplicationWorker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		rm.workers[i] = &ReplicationWorker{
			ID:       i,
			manager:  rm,
			stopChan: make(chan struct{}),
		}
	}
	
	// Create health checker
	rm.healthChecker = &HealthChecker{
		manager:       rm,
		checkInterval: config.HealthCheckInterval,
		timeout:       config.HealthCheckTimeout,
		stopChan:      make(chan struct{}),
	}
	
	return rm, nil
}

// Start starts the replication manager
func (rm *ReplicationManager) Start() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if rm.started {
		return fmt.Errorf("replication manager already started")
	}
	
	// Load existing policies
	if err := rm.loadPolicies(); err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}
	
	// Start workers
	for _, worker := range rm.workers {
		go worker.start()
	}
	
	// Start health checker
	go rm.healthChecker.start()
	
	// Start policy enforcement routine
	go rm.policyEnforcementRoutine()
	
	rm.started = true
	rm.logger.Info("replication manager started", "workers", len(rm.workers))
	
	return nil
}

// SetReplicationPolicy sets a replication policy for a model
func (rm *ReplicationManager) SetReplicationPolicy(modelName string, policy *ReplicationPolicy) error {
	policy.ModelName = modelName
	policy.UpdatedAt = time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	
	rm.policiesMutex.Lock()
	rm.policies[modelName] = policy
	rm.policiesMutex.Unlock()
	
	// Trigger policy enforcement
	go rm.enforcePolicy(modelName)
	
	rm.logger.Info("replication policy set", "model", modelName, "min_replicas", policy.MinReplicas, "max_replicas", policy.MaxReplicas)
	
	return nil
}

// GetReplicationPolicy gets the replication policy for a model
func (rm *ReplicationManager) GetReplicationPolicy(modelName string) (*ReplicationPolicy, bool) {
	rm.policiesMutex.RLock()
	defer rm.policiesMutex.RUnlock()
	
	policy, exists := rm.policies[modelName]
	return policy, exists
}

// ReplicateModel replicates a model to a specific peer
func (rm *ReplicationManager) ReplicateModel(modelName, targetPeer string) error {
	task := &ReplicationTask{
		Type:         TaskTypeReplicate,
		ModelName:    modelName,
		TargetPeer:   targetPeer,
		Priority:     1,
		MaxRetries:   3,
		CreatedAt:    time.Now(),
		ResponseChan: make(chan error, 1),
	}
	
	select {
	case rm.workQueue <- task:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("replication queue full")
	}
	
	select {
	case err := <-task.ResponseChan:
		return err
	case <-time.After(10 * time.Minute):
		return fmt.Errorf("replication timeout")
	}
}

// GetReplicas returns all replicas for a model
func (rm *ReplicationManager) GetReplicas(modelName string) []*ReplicaInfo {
	rm.replicasMutex.RLock()
	defer rm.replicasMutex.RUnlock()
	
	var replicas []*ReplicaInfo
	for _, replica := range rm.replicas {
		if replica.ModelName == modelName {
			replicas = append(replicas, replica)
		}
	}
	
	return replicas
}

// GetAllReplicas returns all replicas
func (rm *ReplicationManager) GetAllReplicas() []*ReplicaInfo {
	rm.replicasMutex.RLock()
	defer rm.replicasMutex.RUnlock()
	
	replicas := make([]*ReplicaInfo, 0, len(rm.replicas))
	for _, replica := range rm.replicas {
		replicas = append(replicas, replica)
	}
	
	return replicas
}

// enforcePolicy enforces the replication policy for a model
func (rm *ReplicationManager) enforcePolicy(modelName string) {
	policy, exists := rm.GetReplicationPolicy(modelName)
	if !exists {
		return
	}
	
	replicas := rm.GetReplicas(modelName)
	currentReplicas := len(replicas)
	
	rm.logger.Info("enforcing policy", "model", modelName, "current_replicas", currentReplicas, "min_replicas", policy.MinReplicas)
	
	if currentReplicas < policy.MinReplicas {
		// Need to create more replicas
		needed := policy.MinReplicas - currentReplicas
		rm.logger.Info("need more replicas", "model", modelName, "needed", needed)
		
		// Find suitable peers
		peers := rm.findSuitablePeers(modelName, policy, needed)
		
		for _, peer := range peers {
			task := &ReplicationTask{
				Type:         TaskTypeReplicate,
				ModelName:    modelName,
				TargetPeer:   peer,
				Priority:     policy.Priority,
				MaxRetries:   3,
				CreatedAt:    time.Now(),
				ResponseChan: make(chan error, 1),
			}
			
			select {
			case rm.workQueue <- task:
			default:
				rm.logger.Error("replication queue full", "model", modelName, "peer", peer)
			}
		}
	} else if currentReplicas > policy.MaxReplicas {
		// Need to remove some replicas
		excess := currentReplicas - policy.MaxReplicas
		rm.logger.Info("need fewer replicas", "model", modelName, "excess", excess)
		
		// Find replicas to remove (prefer unhealthy ones)
		toRemove := rm.selectReplicasToRemove(modelName, excess)
		
		for _, replica := range toRemove {
			task := &ReplicationTask{
				Type:         TaskTypeRemove,
				ModelName:    modelName,
				TargetPeer:   replica.PeerID,
				Priority:     policy.Priority,
				MaxRetries:   3,
				CreatedAt:    time.Now(),
				ResponseChan: make(chan error, 1),
			}
			
			select {
			case rm.workQueue <- task:
			default:
				rm.logger.Error("replication queue full", "model", modelName, "peer", replica.PeerID)
			}
		}
	}
}

// findSuitablePeers finds suitable peers for replication
func (rm *ReplicationManager) findSuitablePeers(modelName string, policy *ReplicationPolicy, count int) []string {
	// Get all connected peers
	connectedPeers := rm.p2p.GetConnectedPeers()
	
	// Filter based on policy
	var suitable []string
	existing := make(map[string]bool)
	
	// Get existing replicas
	replicas := rm.GetReplicas(modelName)
	for _, replica := range replicas {
		existing[replica.PeerID] = true
	}
	
	// Check preferred peers first
	for _, peer := range policy.PreferredPeers {
		if len(suitable) >= count {
			break
		}
		
		if existing[peer] {
			continue // Already has replica
		}
		
		if rm.isPeerConnected(peer, connectedPeers) {
			suitable = append(suitable, peer)
		}
	}
	
	// Add other suitable peers if needed
	for _, peer := range connectedPeers {
		if len(suitable) >= count {
			break
		}
		
		if existing[peer] {
			continue // Already has replica
		}
		
		if rm.isPeerExcluded(peer, policy.ExcludedPeers) {
			continue // Excluded
		}
		
		if rm.isPeerSuitable(peer, policy.Constraints) {
			suitable = append(suitable, peer)
		}
	}
	
	return suitable
}

// selectReplicasToRemove selects replicas to remove
func (rm *ReplicationManager) selectReplicasToRemove(modelName string, count int) []*ReplicaInfo {
	replicas := rm.GetReplicas(modelName)
	
	// Sort by health and last sync time (prefer to remove unhealthy ones)
	// This is a simplified selection logic
	var toRemove []*ReplicaInfo
	
	for _, replica := range replicas {
		if len(toRemove) >= count {
			break
		}
		
		if replica.Health == HealthError || replica.Status == ReplicaStatusUnhealthy {
			toRemove = append(toRemove, replica)
		}
	}
	
	// If still need more, remove based on last sync time
	if len(toRemove) < count {
		for _, replica := range replicas {
			if len(toRemove) >= count {
				break
			}
			
			alreadySelected := false
			for _, selected := range toRemove {
				if selected.PeerID == replica.PeerID {
					alreadySelected = true
					break
				}
			}
			
			if !alreadySelected {
				toRemove = append(toRemove, replica)
			}
		}
	}
	
	return toRemove
}

// isPeerConnected checks if a peer is connected
func (rm *ReplicationManager) isPeerConnected(peer string, connectedPeers []string) bool {
	for _, connected := range connectedPeers {
		if connected == peer {
			return true
		}
	}
	return false
}

// isPeerExcluded checks if a peer is excluded
func (rm *ReplicationManager) isPeerExcluded(peer string, excludedPeers []string) bool {
	for _, excluded := range excludedPeers {
		if excluded == peer {
			return true
		}
	}
	return false
}

// isPeerSuitable checks if a peer meets the constraints
func (rm *ReplicationManager) isPeerSuitable(peer string, constraints map[string]string) bool {
	// TODO: Implement constraint checking
	// This would check things like:
	// - Available storage space
	// - Network bandwidth
	// - Geographic location
	// - Hardware capabilities
	return true
}

// loadPolicies loads existing replication policies
func (rm *ReplicationManager) loadPolicies() error {
	// TODO: Load policies from persistent storage
	// For now, create default policies for existing models
	
	models := rm.manager.GetAllModels()
	for modelName := range models {
		if _, exists := rm.GetReplicationPolicy(modelName); !exists {
			// Create default policy
			policy := &ReplicationPolicy{
				ModelName:         modelName,
				MinReplicas:       rm.config.DefaultMinReplicas,
				MaxReplicas:       rm.config.DefaultMaxReplicas,
				ReplicationFactor: rm.config.DefaultReplicationFactor,
				SyncInterval:      rm.config.DefaultSyncInterval,
				Priority:          1,
				Constraints:       make(map[string]string),
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}
			
			rm.policiesMutex.Lock()
			rm.policies[modelName] = policy
			rm.policiesMutex.Unlock()
		}
	}
	
	return nil
}

// policyEnforcementRoutine runs periodic policy enforcement
func (rm *ReplicationManager) policyEnforcementRoutine() {
	ticker := time.NewTicker(rm.config.PolicyEnforcementInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.enforceAllPolicies()
		}
	}
}

// enforceAllPolicies enforces all replication policies
func (rm *ReplicationManager) enforceAllPolicies() {
	rm.policiesMutex.RLock()
	policies := make(map[string]*ReplicationPolicy)
	for k, v := range rm.policies {
		policies[k] = v
	}
	rm.policiesMutex.RUnlock()
	
	for modelName := range policies {
		rm.enforcePolicy(modelName)
	}
}

// Shutdown gracefully shuts down the replication manager
func (rm *ReplicationManager) Shutdown(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	if !rm.started {
		return nil
	}
	
	// Stop workers
	for _, worker := range rm.workers {
		close(worker.stopChan)
	}
	
	// Stop health checker
	close(rm.healthChecker.stopChan)
	
	rm.cancel()
	rm.started = false
	
	rm.logger.Info("replication manager shutdown complete")
	return nil
}

// ReplicationWorker methods

// start starts the replication worker
func (w *ReplicationWorker) start() {
	w.manager.logger.Info("replication worker started", "worker_id", w.ID)
	
	for {
		select {
		case <-w.stopChan:
			w.manager.logger.Info("replication worker stopped", "worker_id", w.ID)
			return
		case task := <-w.manager.workQueue:
			w.processTask(task)
		}
	}
}

// processTask processes a replication task
func (w *ReplicationWorker) processTask(task *ReplicationTask) {
	w.manager.logger.Info("processing replication task", "worker_id", w.ID, "type", task.Type, "model", task.ModelName, "peer", task.TargetPeer)
	
	var err error
	
	switch task.Type {
	case TaskTypeReplicate:
		err = w.processReplicate(task)
	case TaskTypeSync:
		err = w.processSync(task)
	case TaskTypeRemove:
		err = w.processRemove(task)
	case TaskTypeVerify:
		err = w.processVerify(task)
	default:
		err = fmt.Errorf("unknown task type: %s", task.Type)
	}
	
	if err != nil {
		w.manager.logger.Error("replication task failed", "worker_id", w.ID, "type", task.Type, "model", task.ModelName, "error", err)
		
		// Retry if possible
		if task.Retries < task.MaxRetries {
			task.Retries++
			go func() {
				time.Sleep(time.Duration(task.Retries) * time.Second)
				select {
				case w.manager.workQueue <- task:
				default:
					// Queue full, send error
					select {
					case task.ResponseChan <- fmt.Errorf("retry failed: queue full"):
					default:
					}
				}
			}()
			return
		}
	} else {
		w.manager.logger.Info("replication task completed", "worker_id", w.ID, "type", task.Type, "model", task.ModelName)
	}
	
	// Send response
	select {
	case task.ResponseChan <- err:
	case <-time.After(time.Second):
		// Response channel blocked
	}
}

// processReplicate processes a replicate task
func (w *ReplicationWorker) processReplicate(task *ReplicationTask) error {
	// TODO: Implement actual replication logic
	// This would involve:
	// 1. Checking if model exists locally
	// 2. Initiating transfer to target peer
	// 3. Monitoring transfer progress
	// 4. Updating replica information
	
	time.Sleep(100 * time.Millisecond) // Simulate work
	
	// Create replica info
	replica := &ReplicaInfo{
		ModelName:    task.ModelName,
		PeerID:       task.TargetPeer,
		Status:       ReplicaStatusHealthy,
		LastSync:     time.Now(),
		SyncAttempts: 1,
		Health:       HealthGood,
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Store replica info
	replicaKey := fmt.Sprintf("%s:%s", task.ModelName, task.TargetPeer)
	w.manager.replicasMutex.Lock()
	w.manager.replicas[replicaKey] = replica
	w.manager.replicasMutex.Unlock()
	
	return nil
}

// processSync processes a sync task
func (w *ReplicationWorker) processSync(task *ReplicationTask) error {
	// Use sync manager to synchronize the model
	return w.manager.syncMgr.SynchronizeModel(task.ModelName, task.TargetPeer, SyncTypeIncremental)
}

// processRemove processes a remove task
func (w *ReplicationWorker) processRemove(task *ReplicationTask) error {
	// TODO: Implement replica removal logic
	// This would involve:
	// 1. Sending remove request to target peer
	// 2. Waiting for confirmation
	// 3. Updating replica information
	
	time.Sleep(50 * time.Millisecond) // Simulate work
	
	// Remove replica info
	replicaKey := fmt.Sprintf("%s:%s", task.ModelName, task.TargetPeer)
	w.manager.replicasMutex.Lock()
	delete(w.manager.replicas, replicaKey)
	w.manager.replicasMutex.Unlock()
	
	return nil
}

// processVerify processes a verify task
func (w *ReplicationWorker) processVerify(task *ReplicationTask) error {
	// TODO: Implement replica verification logic
	// This would involve:
	// 1. Getting replica hash from target peer
	// 2. Comparing with local hash
	// 3. Updating replica health status
	
	time.Sleep(25 * time.Millisecond) // Simulate work
	return nil
}

// HealthChecker methods

// start starts the health checker
func (hc *HealthChecker) start() {
	hc.manager.logger.Info("health checker started")
	
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-hc.stopChan:
			hc.manager.logger.Info("health checker stopped")
			return
		case <-ticker.C:
			hc.checkAllReplicas()
		}
	}
}

// checkAllReplicas checks the health of all replicas
func (hc *HealthChecker) checkAllReplicas() {
	replicas := hc.manager.GetAllReplicas()
	
	for _, replica := range replicas {
		go hc.checkReplica(replica)
	}
}

// checkReplica checks the health of a single replica
func (hc *HealthChecker) checkReplica(replica *ReplicaInfo) {
	// TODO: Implement actual health checking
	// This would involve:
	// 1. Pinging the peer
	// 2. Checking model availability
	// 3. Verifying model integrity
	// 4. Updating health status
	
	// For now, simulate health check
	healthy := true // Assume healthy for simulation
	
	hc.manager.replicasMutex.Lock()
	defer hc.manager.replicasMutex.Unlock()
	
	replicaKey := fmt.Sprintf("%s:%s", replica.ModelName, replica.PeerID)
	if storedReplica, exists := hc.manager.replicas[replicaKey]; exists {
		if healthy {
			storedReplica.Health = HealthGood
			storedReplica.Status = ReplicaStatusHealthy
		} else {
			storedReplica.Health = HealthError
			storedReplica.Status = ReplicaStatusUnhealthy
		}
		storedReplica.UpdatedAt = time.Now()
	}
}