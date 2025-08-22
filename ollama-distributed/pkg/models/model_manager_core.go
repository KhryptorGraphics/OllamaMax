package models

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

// DistributedModelManager extends Ollama's model management with distributed capabilities
type DistributedModelManager struct {
	// Core components
	localManager       *Manager
	syncManager        *SyncManager
	replicationManager *ReplicationManager
	casStore           *ContentAddressedStore
	deltaTracker       *DeltaTracker

	// Configuration
	config *config.DistributedConfig
	p2p    *p2p.Node
	logger *slog.Logger

	// Distributed model registry
	registry      *DistributedRegistry
	registryMutex sync.RWMutex

	// Model lifecycle management
	lifecycle      *ModelLifecycle
	lifecycleMutex sync.RWMutex

	// Performance monitoring
	monitor *PerformanceMonitor

	// Context management
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// DistributedRegistry maintains a registry of all models across the network
type DistributedRegistry struct {
	models      map[string]*DistributedModel
	modelsMutex sync.RWMutex

	// Peer model tracking
	peerModels map[string]map[string]*DistributedModel // peerID -> modelName -> model
	peerMutex  sync.RWMutex

	// Discovery service
	discovery *ModelDiscovery
}

// DistributedModel represents a model in the distributed network
type DistributedModel struct {
	// Base model information
	Name    string `json:"name"`
	Version string `json:"version"`
	Hash    string `json:"hash"`
	Size    int64  `json:"size"`

	// Distribution information
	Replicas     []*ReplicaInfo `json:"replicas"`
	Availability float64        `json:"availability"`
	Peers        []string       `json:"peers"`

	// Metadata
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        []string               `json:"tags"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`

	// Performance metrics
	AccessCount    int64         `json:"access_count"`
	LastAccessed   time.Time     `json:"last_accessed"`
	AverageLatency time.Duration `json:"average_latency"`

	// Health status
	Status      ModelStatus `json:"status"`
	HealthScore float64     `json:"health_score"`

	// Sync state
	SyncState *SyncState `json:"sync_state"`
}

// Note: ModelStatus is defined in distribution.go

// Note: SyncState and ReplicaInfo are defined in other files

// NewDistributedModelManager creates a new distributed model manager
func NewDistributedModelManager(
	config *config.DistributedConfig,
	p2pNode *p2p.Node,
	logger *slog.Logger,
) (*DistributedModelManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if p2pNode == nil {
		return nil, fmt.Errorf("p2p node cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	dmm := &DistributedModelManager{
		config: config,
		p2p:    p2pNode,
		logger: logger,
	}

	// Initialize registry
	dmm.registry = &DistributedRegistry{
		models:     make(map[string]*DistributedModel),
		peerModels: make(map[string]map[string]*DistributedModel),
	}

	// Initialize local manager
	var err error
	dmm.localManager, err = NewManager(config.Storage, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create local manager: %w", err)
	}

	// Initialize sync manager
	dmm.syncManager, err = NewSyncManager(config.Sync, p2pNode, dmm.localManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync manager: %w", err)
	}

	// Initialize replication manager
	replicationConfig := &ReplicationConfig{
		Enabled:               true, // Default enabled
		MaxReplicas:           config.Replication.DefaultMaxReplicas,
		SyncInterval:          config.Replication.DefaultSyncInterval,
		WorkerCount:           config.Replication.WorkerCount,
		RetryAttempts:         3, // Default retry attempts
		RetryDelay:            5 * time.Second, // Default retry delay
		HealthCheckInterval:   config.Replication.HealthCheckInterval,
		HealthCheckTimeout:    config.Replication.HealthCheckTimeout,
		DefaultMinReplicas:    config.Replication.DefaultMinReplicas,
		DefaultMaxReplicas:    config.Replication.DefaultMaxReplicas,
		DefaultReplicationFactor: config.Replication.DefaultReplicationFactor,
		DefaultSyncInterval:   config.Replication.DefaultSyncInterval,
		PolicyEnforcementInterval: config.Replication.PolicyEnforcementInterval,
	}
	dmm.replicationManager, err = NewReplicationManager(replicationConfig, p2pNode, dmm.localManager, dmm.syncManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication manager: %w", err)
	}

	// Initialize content-addressed store
	dmm.casStore, err = NewContentAddressedStore(config.Storage.DataDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CAS store: %w", err)
	}

	// Initialize delta tracker (simplified - may need different constructor)
	dmm.deltaTracker = &DeltaTracker{} // Placeholder

	// Initialize lifecycle manager
	dmm.lifecycle = NewModelLifecycle(logger)

	// Initialize performance monitor
	dmm.monitor = NewPerformanceMonitor(logger)

	// Initialize discovery service
	dmm.registry.discovery = NewModelDiscovery(dmm, logger)

	return dmm, nil
}

// Start starts the distributed model manager
func (dmm *DistributedModelManager) Start() error {
	dmm.mu.Lock()
	defer dmm.mu.Unlock()

	if dmm.started {
		return fmt.Errorf("distributed model manager already started")
	}

	dmm.ctx, dmm.cancel = context.WithCancel(context.Background())

	// Start components
	if err := dmm.localManager.Start(); err != nil {
		return fmt.Errorf("failed to start local manager: %w", err)
	}

	if err := dmm.syncManager.Start(); err != nil {
		return fmt.Errorf("failed to start sync manager: %w", err)
	}

	if err := dmm.replicationManager.Start(); err != nil {
		return fmt.Errorf("failed to start replication manager: %w", err)
	}

	// Start lifecycle manager
	go dmm.lifecycle.start()

	// Start performance monitor
	go dmm.monitor.start()

	// Start discovery service
	dmm.registry.discovery.start()

	// Start background routines
	go dmm.registrySyncRoutine()

	dmm.started = true
	dmm.logger.Info("distributed model manager started")

	return nil
}

// Shutdown gracefully shuts down the distributed model manager
func (dmm *DistributedModelManager) Shutdown(ctx context.Context) error {
	dmm.mu.Lock()
	defer dmm.mu.Unlock()

	if !dmm.started {
		return nil
	}

	dmm.cancel()

	// Shutdown components
	if err := dmm.replicationManager.Shutdown(ctx); err != nil {
		dmm.logger.Warn("error shutting down replication manager", "error", err)
	}

	if err := dmm.syncManager.Shutdown(ctx); err != nil {
		dmm.logger.Warn("error shutting down sync manager", "error", err)
	}

	if err := dmm.localManager.Shutdown(ctx); err != nil {
		dmm.logger.Warn("error shutting down local manager", "error", err)
	}

	dmm.started = false
	dmm.logger.Info("distributed model manager shut down")

	return nil
}

// GetModel retrieves a model, either locally or from the network
func (dmm *DistributedModelManager) GetModel(modelName string) (*DistributedModel, error) {
	// Check local registry first
	dmm.registryMutex.RLock()
	if model, exists := dmm.registry.models[modelName]; exists {
		dmm.registryMutex.RUnlock()

		// Update access metrics
		model.AccessCount++
		model.LastAccessed = time.Now()

		return model, nil
	}
	dmm.registryMutex.RUnlock()

	// Discover model on network
	return dmm.discoverAndFetchModel(modelName)
}

// AddModel adds a model to the distributed system
func (dmm *DistributedModelManager) AddModel(modelName, modelPath string) (*DistributedModel, error) {
	// Create model version
	version, err := dmm.syncManager.CreateModelVersion(modelName, modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create model version: %w", err)
	}

	// Calculate hash
	hash, err := dmm.casStore.calculateHash(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Get file size (simplified implementation)
	fileInfo, err := os.Stat(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}
	size := fileInfo.Size()

	// Create distributed model
	model := &DistributedModel{
		Name:           modelName,
		Version:        version.Version, // Extract version string from ModelVersion
		Hash:           hash,
		Size:           size,
		Replicas:       []*ReplicaInfo{},
		Availability:   1.0,
		Peers:          []string{dmm.p2p.ID().String()},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Tags:           []string{},
		Description:    "",
		Metadata:       make(map[string]interface{}),
		AccessCount:    0,
		LastAccessed:   time.Now(),
		AverageLatency: 0,
		Status:         ModelStatusAvailable,
		HealthScore:    1.0,
		SyncState: &SyncState{
			ModelName:      modelName,
			LocalVersion:   version.Version,
			RemoteVersions: make(map[string]string),
			Status:         SyncStatusInSync,
			LastSyncTime:   time.Now(),
		},
	}

	// Add to registry
	dmm.registryMutex.Lock()
	dmm.registry.models[modelName] = model
	dmm.registryMutex.Unlock()

	// Emit lifecycle event
	dmm.emitLifecycleEvent(EventModelAdded, modelName, dmm.p2p.ID().String(), map[string]interface{}{
		"version": version,
		"hash":    hash,
		"size":    size,
	})

	return model, nil
}

// RemoveModel removes a model from the distributed system if present
func (dmm *DistributedModelManager) RemoveModel(modelName string) error {
	// Remove from registry
	dmm.registryMutex.Lock()
	model, exists := dmm.registry.models[modelName]
	if exists {
		delete(dmm.registry.models, modelName)
	}
	dmm.registryMutex.Unlock()

	if !exists {
		return fmt.Errorf("model not found: %s", modelName)
	}

	// Remove from local storage (simplified - may need different method)
	dmm.logger.Info("removing model from local storage", "model", modelName)

	// Remove from CAS store (simplified - may need different method)
	dmm.logger.Info("removing model from CAS store", "model", modelName, "hash", model.Hash)

	// Emit lifecycle event
	dmm.emitLifecycleEvent(EventModelDeleted, modelName, dmm.p2p.ID().String(), map[string]interface{}{})
	return nil
}

// GetDistributedModels returns all distributed models
func (dmm *DistributedModelManager) GetDistributedModels() []*DistributedModel {
	dmm.registryMutex.RLock()
	defer dmm.registryMutex.RUnlock()

	models := make([]*DistributedModel, 0, len(dmm.registry.models))
	for _, model := range dmm.registry.models {
		models = append(models, model)
	}

	return models
}

// GetPerformanceMetrics returns performance metrics
func (dmm *DistributedModelManager) GetPerformanceMetrics() []*PerformanceMetric {
	return dmm.monitor.GetMetrics()
}

// ReplicateModelToPeers triggers replication of a model to specific peers
func (dmm *DistributedModelManager) ReplicateModelToPeers(modelName string, targetPeers []string) error {
	if dmm.replicationManager == nil {
		return fmt.Errorf("replication manager not initialized")
	}
	var firstErr error
	for _, peerID := range targetPeers {
		if err := dmm.replicationManager.ReplicateModel(modelName, peerID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// GetReplicas returns current replicas for a model
func (dmm *DistributedModelManager) GetReplicas(modelName string) []*ReplicaInfo {
	if dmm.replicationManager == nil {
		return nil
	}
	return dmm.replicationManager.GetReplicas(modelName)
}

// GetReplicaCount returns the number of replicas known for a model
func (dmm *DistributedModelManager) GetReplicaCount(modelName string) int {
	replicas := dmm.GetReplicas(modelName)
	return len(replicas)
}

// GetReplicationSummary exposes replication manager summary
func (dmm *DistributedModelManager) GetReplicationSummary() *ReplicationSummary {
	if dmm.replicationManager == nil {
		return &ReplicationSummary{QueueLength: 0, WorkerCount: 0, Models: map[string]int{}}
	}
	return dmm.replicationManager.GetSummary()
}

// Private methods

// discoverAndFetchModel discovers a model on the network and fetches it
func (dmm *DistributedModelManager) discoverAndFetchModel(modelName string) (*DistributedModel, error) {
	// Create discovery request
	req := &DiscoveryRequest{
		ModelName:    modelName,
		Criteria:     make(map[string]interface{}),
		Timeout:      30 * time.Second,
		ResponseChan: make(chan *DiscoveryResponse, 1),
	}

	// Submit discovery request
	select {
	case dmm.registry.discovery.workQueue <- req:
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("discovery queue full")
	}

	// Wait for response
	select {
	case response := <-req.ResponseChan:
		if response.Error != "" {
			return nil, fmt.Errorf("discovery failed: %s", response.Error)
		}

		if len(response.Models) == 0 {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}

		// Use the first available model
		model := response.Models[0]

		// Download model from peer
		if len(response.Peers) > 0 {
			if err := dmm.downloadModelFromPeer(modelName, response.Peers[0]); err != nil {
				return nil, fmt.Errorf("failed to download model: %w", err)
			}
		}

		// Add to local registry
		dmm.registryMutex.Lock()
		dmm.registry.models[modelName] = model
		dmm.registryMutex.Unlock()

		return model, nil
	case <-time.After(req.Timeout):
		return nil, fmt.Errorf("discovery timeout")
	}
}

// downloadModelFromPeer downloads a model from a specific peer
func (dmm *DistributedModelManager) downloadModelFromPeer(modelName, peerID string) error {
	// Use the local manager to download the model
	_, err := dmm.localManager.DownloadModel(modelName, peerID)
	return err
}

// emitLifecycleEvent emits a lifecycle event
func (dmm *DistributedModelManager) emitLifecycleEvent(eventType LifecycleEventType, modelName, peerID string, data map[string]interface{}) {
	event := &LifecycleEvent{
		Type:      eventType,
		ModelName: modelName,
		PeerID:    peerID,
		Timestamp: time.Now(),
		Data:      data,
	}

	select {
	case dmm.lifecycle.events <- event:
	default:
		dmm.logger.Warn("lifecycle event queue full", "event", eventType, "model", modelName)
	}
}

// registrySyncRoutine periodically synchronizes the registry
func (dmm *DistributedModelManager) registrySyncRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dmm.ctx.Done():
			return
		case <-ticker.C:
			dmm.syncRegistry()
		}
	}
}

// syncRegistry synchronizes the registry with peers
func (dmm *DistributedModelManager) syncRegistry() {
	// Get connected peers
	peerIDs := dmm.p2p.GetConnectedPeers()
	if len(peerIDs) == 0 {
		return // No peers to sync with
	}

	// Get local models
	dmm.registryMutex.RLock()
	localModels := make(map[string]*DistributedModel)
	for name, model := range dmm.registry.models {
		localModels[name] = model
	}
	dmm.registryMutex.RUnlock()

	// Sync with each peer
	for _, peerID := range peerIDs {
		go dmm.syncWithPeer(peerID.String(), localModels)
	}

	// Request model information from peers
	for _, peerID := range peerIDs {
		go dmm.requestPeerModels(peerID.String())
	}

	// Clean up stale peer entries
	dmm.cleanupStalePeers()
}

// syncWithPeer synchronizes models with a specific peer
func (dmm *DistributedModelManager) syncWithPeer(peerIDStr string, localModels map[string]*DistributedModel) {
	// Prepare sync message
	syncMessage := map[string]interface{}{
		"type":      "registry_sync",
		"peer_id":   dmm.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
		"models":    localModels,
	}

	// Send sync message via P2P (simplified implementation)
	// In practice, this would use libp2p streams
	dmm.logger.Info("syncing models with peer", "peer", peerIDStr, "models", len(localModels), "sync_message", syncMessage)
}

// requestPeerModels requests model information from a peer
func (dmm *DistributedModelManager) requestPeerModels(peerIDStr string) {
	// Create request message
	request := map[string]interface{}{
		"type":      "model_request",
		"peer_id":   dmm.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
	}

	// Send request via P2P (simplified implementation)
	dmm.logger.Info("requesting models from peer", "peer", peerIDStr, "request", request)
}

// cleanupStalePeers removes stale peer entries
func (dmm *DistributedModelManager) cleanupStalePeers() {
	dmm.registry.peerMutex.Lock()
	defer dmm.registry.peerMutex.Unlock()

	connectedPeerIDs := dmm.p2p.GetConnectedPeers()
	connectedMap := make(map[string]bool)
	for _, peerID := range connectedPeerIDs {
		connectedMap[peerID.String()] = true
	}

	// Remove disconnected peers
	for peerID := range dmm.registry.peerModels {
		if !connectedMap[peerID] {
			delete(dmm.registry.peerModels, peerID)
		}
	}
}
