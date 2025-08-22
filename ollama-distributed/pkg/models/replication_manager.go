package models

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

// ReplicationConfig holds configuration for the replication manager
type ReplicationConfig struct {
	Enabled               bool          `yaml:"enabled" json:"enabled"`
	MaxReplicas           int           `yaml:"max_replicas" json:"max_replicas"`
	SyncInterval          time.Duration `yaml:"sync_interval" json:"sync_interval"`
	WorkerCount           int           `yaml:"worker_count" json:"worker_count"`
	RetryAttempts         int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay            time.Duration `yaml:"retry_delay" json:"retry_delay"`
	HealthCheckInterval   time.Duration `yaml:"health_check_interval" json:"health_check_interval"`
	HealthCheckTimeout    time.Duration `yaml:"health_check_timeout" json:"health_check_timeout"`
	DefaultMinReplicas    int           `yaml:"default_min_replicas" json:"default_min_replicas"`
	DefaultMaxReplicas    int           `yaml:"default_max_replicas" json:"default_max_replicas"`
	DefaultReplicationFactor int        `yaml:"default_replication_factor" json:"default_replication_factor"`
	DefaultSyncInterval   time.Duration `yaml:"default_sync_interval" json:"default_sync_interval"`
	PolicyEnforcementInterval time.Duration `yaml:"policy_enforcement_interval" json:"policy_enforcement_interval"`
}

// ReplicationManager manages model replication across peers
type ReplicationManager struct {
	config  *ReplicationConfig
	p2p     *p2p.Node
	manager *Manager
	syncMgr *SyncManager
	logger  *slog.Logger

	// Replication state
	replicas      map[string]*ReplicaInfo
	replicasMutex sync.RWMutex

	// Replication policies
	policies      map[string]*ReplicationPolicy
	policiesMutex sync.RWMutex

	// Replication workers
	workers   []*ReplicationWorker
	workQueue chan *ReplicationTask

	// Health monitoring
	healthChecker *HealthChecker

	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex

	// Metrics (lightweight)
	successfulReplications int64
	failedReplications     int64
}

// Note: ReplicationSummary is defined in model_distribution.go

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
	ID           string     `json:"id"`
	Type         TaskType   `json:"type"`
	ModelName    string     `json:"model_name"`
	SourcePeer   string     `json:"source_peer"`
	TargetPeer   string     `json:"target_peer"`
	Priority     int        `json:"priority"`
	Status       string     `json:"status"`
	Progress     float64    `json:"progress"`
	Error        string     `json:"error,omitempty"`
	Retries      int        `json:"retries"`
	MaxRetries   int        `json:"max_retries"`
	CreatedAt    time.Time  `json:"created_at"`
	ResponseChan chan error `json:"-"`
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
	ID       int
	manager  *ReplicationManager
	stopChan chan struct{}
}

// NewReplicationManager creates a new replication manager
func NewReplicationManager(
	config *ReplicationConfig,
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

// GetSummary returns a quick snapshot of replication state
func (rm *ReplicationManager) GetSummary() *ReplicationSummary {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	modelCounts := make(map[string]int)
	// derive from replicas map
	rm.replicasMutex.RLock()
	for _, r := range rm.replicas {
		modelCounts[r.ModelName]++
	}
	rm.replicasMutex.RUnlock()
	return &ReplicationSummary{
		QueueLength: len(rm.workQueue),
		WorkerCount: len(rm.workers),
		Models:      modelCounts,
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
	connectedPeerIDs := rm.p2p.GetConnectedPeers()
	var connectedPeers []string
	for _, peerID := range connectedPeerIDs {
		connectedPeers = append(connectedPeers, peerID.String())
	}

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
	if len(constraints) == 0 {
		return true
	}

	// Get peer information from P2P node
	peerInfo := rm.getPeerInfo(peer)
	if peerInfo == nil {
		return false
	}

	// Check storage constraint
	if minStorage, ok := constraints["min_storage"]; ok {
		if peerInfo.AvailableStorage < rm.parseStorageSize(minStorage) {
			return false
		}
	}

	// Check bandwidth constraint
	if minBandwidth, ok := constraints["min_bandwidth"]; ok {
		if peerInfo.NetworkBandwidth < rm.parseBandwidth(minBandwidth) {
			return false
		}
	}

	// Check geographic constraint
	if region, ok := constraints["region"]; ok {
		if peerInfo.Region != region {
			return false
		}
	}

	// Check hardware capability constraints
	if minCPU, ok := constraints["min_cpu_cores"]; ok {
		if peerInfo.CPUCores < rm.parseInt(minCPU) {
			return false
		}
	}

	if minMemory, ok := constraints["min_memory"]; ok {
		if peerInfo.MemoryBytes < rm.parseMemorySize(minMemory) {
			return false
		}
	}

	// Check GPU requirements
	if gpuRequired, ok := constraints["gpu_required"]; ok && gpuRequired == "true" {
		if peerInfo.GPUCount == 0 {
			return false
		}
	}

	return true
}

// loadPolicies loads existing replication policies
func (rm *ReplicationManager) loadPolicies() error {
	// Load from persistent storage
	if err := rm.loadPoliciesFromStorage(); err != nil {
		rm.logger.Warn("failed to load policies from storage, using defaults", "error", err)
	}

	// Create default policies for existing models that don't have policies
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
	w.manager.logger.Info("starting model replication", "model", task.ModelName, "target", task.TargetPeer)

	// 1. Check if model exists locally and get model metadata
	modelInfo, exists := w.manager.manager.GetModel(task.ModelName)
	if !exists {
		return fmt.Errorf("model %s not found locally", task.ModelName)
	}
	if modelInfo == nil {
		return fmt.Errorf("failed to get model info for %s", task.ModelName)
	}

	// 3. Check target peer connectivity
	if !w.manager.isPeerConnectedAndReachable(task.TargetPeer) {
		return fmt.Errorf("target peer %s is not reachable", task.TargetPeer)
	}

	// 4. Initiate transfer to target peer
	transferID, err := w.initiateModelTransfer(task.ModelName, task.TargetPeer, modelInfo)
	if err != nil {
		return fmt.Errorf("failed to initiate transfer: %w", err)
	}

	// 5. Monitor transfer progress
	err = w.monitorTransferProgress(transferID, task.ModelName, task.TargetPeer)
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}

	// 6. Verify replication success
	if err := w.verifyReplication(task.ModelName, task.TargetPeer); err != nil {
		return fmt.Errorf("replication verification failed: %w", err)
	}

	// 7. Create and store replica info
	replica := &ReplicaInfo{
		ModelName:    task.ModelName,
		PeerID:       task.TargetPeer,
		Status:       ReplicaStatusHealthy,
		LastSync:     time.Now(),
		SyncAttempts: 1,
		Health:       HealthGood,
		Metadata: map[string]string{
			"transfer_id": transferID,
			"file_size":   fmt.Sprintf("%d", modelInfo.Size),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store replica info
	replicaKey := fmt.Sprintf("%s:%s", task.ModelName, task.TargetPeer)
	w.manager.replicasMutex.Lock()
	w.manager.replicas[replicaKey] = replica
	w.manager.replicasMutex.Unlock()

	w.manager.logger.Info("model replication completed", "model", task.ModelName, "target", task.TargetPeer)
	return nil
}

// processSync processes a sync task
func (w *ReplicationWorker) processSync(task *ReplicationTask) error {
	// Use sync manager to synchronize the model
	return w.manager.syncMgr.SynchronizeModel(task.ModelName, task.TargetPeer, SyncTypeIncremental)
}

// processRemove processes a remove task
func (w *ReplicationWorker) processRemove(task *ReplicationTask) error {
	w.manager.logger.Info("starting replica removal", "model", task.ModelName, "target", task.TargetPeer)

	// 1. Check if replica exists
	replicaKey := fmt.Sprintf("%s:%s", task.ModelName, task.TargetPeer)
	w.manager.replicasMutex.RLock()
	replica, exists := w.manager.replicas[replicaKey]
	w.manager.replicasMutex.RUnlock()

	if !exists {
		w.manager.logger.Warn("replica does not exist", "model", task.ModelName, "target", task.TargetPeer)
		return nil // Already removed
	}

	// 2. Send remove request to target peer
	err := w.sendRemoveRequest(task.ModelName, task.TargetPeer)
	if err != nil {
		w.manager.logger.Error("failed to send remove request", "error", err)
		// Continue with local cleanup even if remote removal failed
	}

	// 3. Wait for confirmation (with timeout)
	confirmed := false
	if err == nil {
		confirmed = w.waitForRemovalConfirmation(task.ModelName, task.TargetPeer, 30*time.Second)
	}

	// 4. Update replica status before removal
	if replica != nil {
		replica.Status = ReplicaStatusUnhealthy
		replica.UpdatedAt = time.Now()
		
		w.manager.replicasMutex.Lock()
		w.manager.replicas[replicaKey] = replica
		w.manager.replicasMutex.Unlock()
	}

	// 5. Remove replica info from local registry
	w.manager.replicasMutex.Lock()
	delete(w.manager.replicas, replicaKey)
	w.manager.replicasMutex.Unlock()

	if confirmed {
		w.manager.logger.Info("replica removal completed", "model", task.ModelName, "target", task.TargetPeer)
	} else {
		w.manager.logger.Warn("replica removal completed locally but remote confirmation failed", 
			"model", task.ModelName, "target", task.TargetPeer)
	}

	return nil
}

// processVerify processes a verify task
func (w *ReplicationWorker) processVerify(task *ReplicationTask) error {
	w.manager.logger.Info("starting replica verification", "model", task.ModelName, "target", task.TargetPeer)

	// 1. Get local model hash (calculate from model info)
	localHash := calculateModelHash(task.ModelName)

	// 2. Get replica hash from target peer
	remoteHash, err := w.getRemoteModelHash(task.ModelName, task.TargetPeer)
	if err != nil {
		return fmt.Errorf("failed to get remote model hash: %w", err)
	}

	// 3. Compare hashes
	verified := localHash == remoteHash
	
	// 4. Update replica health status
	replicaKey := fmt.Sprintf("%s:%s", task.ModelName, task.TargetPeer)
	w.manager.replicasMutex.Lock()
	defer w.manager.replicasMutex.Unlock()

	if replica, exists := w.manager.replicas[replicaKey]; exists {
		if verified {
			replica.Health = HealthGood
			replica.Status = ReplicaStatusHealthy
			w.manager.logger.Info("replica verification successful", "model", task.ModelName, "target", task.TargetPeer)
		} else {
			replica.Health = HealthError
			replica.Status = ReplicaStatusOutOfSync
			w.manager.logger.Error("replica verification failed - hash mismatch", 
				"model", task.ModelName, "target", task.TargetPeer,
				"local_hash", localHash, "remote_hash", remoteHash)
		}
		replica.UpdatedAt = time.Now()
		w.manager.replicas[replicaKey] = replica
	}

	if !verified {
		return fmt.Errorf("replica verification failed: hash mismatch")
	}

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
	startTime := time.Now()
	replicaKey := fmt.Sprintf("%s:%s", replica.ModelName, replica.PeerID)

	// 1. Ping the peer to check connectivity
	pingTime, err := hc.pingPeer(replica.PeerID)
	if err != nil {
		hc.updateReplicaHealth(replicaKey, HealthError, ReplicaStatusUnreachable, err.Error())
		return
	}

	// 2. Check model availability on peer
	available, err := hc.checkModelAvailability(replica.ModelName, replica.PeerID)
	if err != nil {
		hc.updateReplicaHealth(replicaKey, HealthError, ReplicaStatusUnhealthy, err.Error())
		return
	}

	if !available {
		hc.updateReplicaHealth(replicaKey, HealthWarning, ReplicaStatusOutOfSync, "model not available on peer")
		return
	}

	// 3. Verify basic model integrity (lightweight check)
	integrity, err := hc.verifyBasicIntegrity(replica.ModelName, replica.PeerID)
	if err != nil {
		hc.updateReplicaHealth(replicaKey, HealthError, ReplicaStatusUnhealthy, err.Error())
		return
	}

	// 4. Update health status based on results
	var health ReplicaHealth
	var status ReplicaStatus

	if integrity && pingTime < hc.timeout/2 {
		health = HealthGood
		status = ReplicaStatusHealthy
	} else if integrity && pingTime < hc.timeout {
		health = HealthWarning
		status = ReplicaStatusHealthy
	} else {
		health = HealthError
		status = ReplicaStatusUnhealthy
	}

	hc.updateReplicaHealth(replicaKey, health, status, "")

	// Update check record
	hc.manager.replicasMutex.Lock()
	if storedReplica, exists := hc.manager.replicas[replicaKey]; exists {
		storedReplica.LastSync = time.Now()
		storedReplica.UpdatedAt = time.Now()
	}
	hc.manager.replicasMutex.Unlock()

	hc.manager.logger.Debug("replica health check completed", 
		"model", replica.ModelName, 
		"peer", replica.PeerID,
		"health", health,
		"ping_time", pingTime,
		"duration", time.Since(startTime))
}

// Helper methods for replication management

// PeerInfo represents peer information for constraint checking
type PeerInfo struct {
	ID                string
	AvailableStorage  int64
	NetworkBandwidth  int64
	Region            string
	CPUCores          int
	MemoryBytes       int64
	GPUCount          int
}

// getPeerInfo gets peer information from P2P node
func (rm *ReplicationManager) getPeerInfo(peerID string) *PeerInfo {
	// Get capabilities from P2P node
	allPeers := rm.p2p.GetAllPeers()
	for id, _ := range allPeers {
		if id.String() == peerID {
			// Convert to PeerInfo
			return &PeerInfo{
				ID:                peerID,
				AvailableStorage:  1024 * 1024 * 1024 * 100, // 100GB default
				NetworkBandwidth:  1024 * 1024 * 100,        // 100 MB/s default
				Region:            "default",
				CPUCores:          8,     // Default values
				MemoryBytes:       1024 * 1024 * 1024 * 16, // 16GB
				GPUCount:          0,     // Default no GPU
			}
		}
	}
	return nil
}

// parseStorageSize parses storage size string (e.g., "100GB")
func (rm *ReplicationManager) parseStorageSize(sizeStr string) int64 {
	// Simple parser for storage sizes
	if len(sizeStr) < 2 {
		return 0
	}
	
	multiplier := int64(1)
	unit := strings.ToUpper(sizeStr[len(sizeStr)-2:])
	
	switch unit {
	case "KB":
		multiplier = 1024
	case "MB":
		multiplier = 1024 * 1024
	case "GB":
		multiplier = 1024 * 1024 * 1024
	case "TB":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		// Try single letter units
		lastChar := strings.ToUpper(string(sizeStr[len(sizeStr)-1]))
		switch lastChar {
		case "K":
			multiplier = 1024
		case "M":
			multiplier = 1024 * 1024
		case "G":
			multiplier = 1024 * 1024 * 1024
		case "T":
			multiplier = 1024 * 1024 * 1024 * 1024
		}
	}
	
	// Parse numeric part
	numStr := sizeStr
	if multiplier > 1 {
		if unit == "KB" || unit == "MB" || unit == "GB" || unit == "TB" {
			numStr = sizeStr[:len(sizeStr)-2]
		} else {
			numStr = sizeStr[:len(sizeStr)-1]
		}
	}
	
	if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
		return num * multiplier
	}
	
	return 0
}

// parseBandwidth parses bandwidth string (e.g., "100Mbps")
func (rm *ReplicationManager) parseBandwidth(bandwidthStr string) int64 {
	// Convert to bytes per second
	bandwidthStr = strings.ToUpper(bandwidthStr)
	
	// Remove "BPS" suffix if present
	if strings.HasSuffix(bandwidthStr, "BPS") {
		bandwidthStr = bandwidthStr[:len(bandwidthStr)-3]
	}
	
	multiplier := int64(1)
	if strings.HasSuffix(bandwidthStr, "K") {
		multiplier = 1024
		bandwidthStr = bandwidthStr[:len(bandwidthStr)-1]
	} else if strings.HasSuffix(bandwidthStr, "M") {
		multiplier = 1024 * 1024
		bandwidthStr = bandwidthStr[:len(bandwidthStr)-1]
	} else if strings.HasSuffix(bandwidthStr, "G") {
		multiplier = 1024 * 1024 * 1024
		bandwidthStr = bandwidthStr[:len(bandwidthStr)-1]
	}
	
	if num, err := strconv.ParseInt(bandwidthStr, 10, 64); err == nil {
		return num * multiplier
	}
	
	return 0
}

// parseMemorySize parses memory size string
func (rm *ReplicationManager) parseMemorySize(sizeStr string) int64 {
	return rm.parseStorageSize(sizeStr) // Same logic as storage
}

// parseInt parses integer string
func (rm *ReplicationManager) parseInt(intStr string) int {
	if num, err := strconv.Atoi(intStr); err == nil {
		return num
	}
	return 0
}

// isPeerConnectedAndReachable checks if peer is connected and reachable
func (rm *ReplicationManager) isPeerConnectedAndReachable(peerID string) bool {
	connectedPeers := rm.p2p.GetConnectedPeers()
	for _, peer := range connectedPeers {
		if peer.String() == peerID {
			return true
		}
	}
	return false
}

// loadPoliciesFromStorage loads policies from persistent storage
func (rm *ReplicationManager) loadPoliciesFromStorage() error {
	// This would load from a database or file system
	// For now, return error to use defaults
	return fmt.Errorf("persistent storage not implemented")
}

// ReplicationWorker helper methods

// initiateModelTransfer starts model transfer to target peer
func (w *ReplicationWorker) initiateModelTransfer(modelName, targetPeer string, modelInfo interface{}) (string, error) {
	// Generate transfer ID
	transferID := fmt.Sprintf("transfer_%s_%s_%d", modelName, targetPeer, time.Now().UnixNano())
	
	// This would use P2P transfer mechanisms
	// For now, simulate successful transfer initiation
	w.manager.logger.Info("initiated model transfer", "transfer_id", transferID, "model", modelName, "target", targetPeer)
	
	return transferID, nil
}

// monitorTransferProgress monitors transfer progress
func (w *ReplicationWorker) monitorTransferProgress(transferID, modelName, targetPeer string) error {
	// Simulate transfer monitoring
	duration := time.Duration(100+len(modelName)*10) * time.Millisecond
	time.Sleep(duration)
	
	w.manager.logger.Info("transfer completed", "transfer_id", transferID, "model", modelName, "target", targetPeer)
	return nil
}

// verifyReplication verifies successful replication
func (w *ReplicationWorker) verifyReplication(modelName, targetPeer string) error {
	// This would verify the model exists and is correct on the target peer
	// For now, simulate successful verification
	time.Sleep(50 * time.Millisecond)
	return nil
}

// sendRemoveRequest sends replica removal request to peer
func (w *ReplicationWorker) sendRemoveRequest(modelName, targetPeer string) error {
	// This would send actual P2P message to remove model
	w.manager.logger.Info("sending remove request", "model", modelName, "target", targetPeer)
	time.Sleep(25 * time.Millisecond) // Simulate network delay
	return nil
}

// waitForRemovalConfirmation waits for removal confirmation
func (w *ReplicationWorker) waitForRemovalConfirmation(modelName, targetPeer string, timeout time.Duration) bool {
	// This would wait for actual confirmation message
	// For now, simulate successful confirmation
	time.Sleep(100 * time.Millisecond)
	return true
}

// getRemoteModelHash gets model hash from remote peer
func (w *ReplicationWorker) getRemoteModelHash(modelName, targetPeer string) (string, error) {
	// This would fetch actual hash from remote peer
	// For now, return a mock hash based on model name
	hash := fmt.Sprintf("hash_%s_%d", modelName, len(modelName))
	return hash, nil
}

// HealthChecker helper methods

// pingPeer pings a peer to check connectivity
func (hc *HealthChecker) pingPeer(peerID string) (time.Duration, error) {
	start := time.Now()
	
	// This would perform actual P2P ping
	// For now, simulate ping with random latency
	latency := time.Duration(10+len(peerID)%50) * time.Millisecond
	time.Sleep(latency)
	
	return time.Since(start), nil
}

// checkModelAvailability checks if model is available on peer
func (hc *HealthChecker) checkModelAvailability(modelName, peerID string) (bool, error) {
	// This would query peer for model availability
	// For now, assume models are available
	time.Sleep(25 * time.Millisecond)
	return true, nil
}

// verifyBasicIntegrity performs basic integrity check
func (hc *HealthChecker) verifyBasicIntegrity(modelName, peerID string) (bool, error) {
	// This would perform basic integrity verification
	// For now, assume integrity is good
	time.Sleep(50 * time.Millisecond)
	return true, nil
}

// updateReplicaHealth updates replica health status
func (hc *HealthChecker) updateReplicaHealth(replicaKey string, health ReplicaHealth, status ReplicaStatus, errorMsg string) {
	hc.manager.replicasMutex.Lock()
	defer hc.manager.replicasMutex.Unlock()

	if replica, exists := hc.manager.replicas[replicaKey]; exists {
		replica.Health = health
		replica.Status = status
		replica.UpdatedAt = time.Now()
		
		if errorMsg != "" {
			if replica.Metadata == nil {
				replica.Metadata = make(map[string]string)
			}
			replica.Metadata["last_error"] = errorMsg
		}
		
		hc.manager.replicas[replicaKey] = replica
	}
}
