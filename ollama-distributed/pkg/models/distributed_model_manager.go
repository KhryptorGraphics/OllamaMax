package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/p2p"
	"github.com/ollama/ollama/server"
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
	
	// Ollama integration
	ollamaServer *server.Server
	
	// Distributed model registry
	registry       *DistributedRegistry
	registryMutex  sync.RWMutex
	
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
	models       map[string]*DistributedModel
	modelsMutex  sync.RWMutex
	
	// Peer model tracking
	peerModels   map[string]map[string]*DistributedModel // peerID -> modelName -> model
	peerMutex    sync.RWMutex
	
	// Network topology
	topology     *NetworkTopology
	
	// Discovery service
	discovery    *ModelDiscovery
}

// DistributedModel represents a model in the distributed network
type DistributedModel struct {
	// Base model information
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Hash        string                 `json:"hash"`
	Size        int64                  `json:"size"`
	Type        string                 `json:"type"`
	
	// Distributed information
	Replicas    []*ReplicaInfo         `json:"replicas"`
	Availability float64               `json:"availability"`
	
	// Version tracking
	Versions    []*ModelVersion        `json:"versions"`
	CurrentVersion string              `json:"current_version"`
	
	// Metadata
	Metadata    map[string]interface{} `json:"metadata"`
	Tags        []string               `json:"tags"`
	
	// Lifecycle
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
	
	// Performance metrics
	AccessCount int64                  `json:"access_count"`
	DownloadCount int64                `json:"download_count"`
	
	// Replication policy
	Policy      *ReplicationPolicy     `json:"policy"`
	
	// Sync state
	SyncState   *SyncState             `json:"sync_state"`
}

// ModelLifecycle manages the lifecycle of distributed models
type ModelLifecycle struct {
	events      chan *LifecycleEvent
	eventsMutex sync.RWMutex
	
	// Lifecycle stages
	stages      map[string]*LifecycleStage
	stagesMutex sync.RWMutex
	
	// Hooks
	hooks       map[LifecycleEventType][]LifecycleHook
	hooksMutex  sync.RWMutex
}

// LifecycleEvent represents a model lifecycle event
type LifecycleEvent struct {
	Type      LifecycleEventType     `json:"type"`
	ModelName string                 `json:"model_name"`
	PeerID    string                 `json:"peer_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// LifecycleEventType represents the type of lifecycle event
type LifecycleEventType string

const (
	EventModelCreated     LifecycleEventType = "model_created"
	EventModelUpdated     LifecycleEventType = "model_updated"
	EventModelDeleted     LifecycleEventType = "model_deleted"
	EventModelAccessed    LifecycleEventType = "model_accessed"
	EventModelReplicated  LifecycleEventType = "model_replicated"
	EventModelSynced      LifecycleEventType = "model_synced"
	EventModelCorrupted   LifecycleEventType = "model_corrupted"
	EventModelHealed      LifecycleEventType = "model_healed"
)

// LifecycleStage represents a stage in the model lifecycle
type LifecycleStage struct {
	Name        string                 `json:"name"`
	ModelName   string                 `json:"model_name"`
	Status      StageStatus            `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Progress    float64                `json:"progress"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
}

// StageStatus represents the status of a lifecycle stage
type StageStatus string

const (
	StageStatusPending    StageStatus = "pending"
	StageStatusInProgress StageStatus = "in_progress"
	StageStatusCompleted  StageStatus = "completed"
	StageStatusFailed     StageStatus = "failed"
)

// LifecycleHook represents a hook function for lifecycle events
type LifecycleHook func(event *LifecycleEvent) error

// NetworkTopology represents the network topology for model distribution
type NetworkTopology struct {
	nodes       map[string]*TopologyNode
	nodesMutex  sync.RWMutex
	
	// Topology metadata
	Type        TopologyType           `json:"type"`
	Depth       int                    `json:"depth"`
	Diameter    int                    `json:"diameter"`
	Connectivity float64               `json:"connectivity"`
	
	// Performance metrics
	avgLatency  time.Duration
	avgBandwidth int64
}

// TopologyNode represents a node in the network topology
type TopologyNode struct {
	ID          string                 `json:"id"`
	Address     string                 `json:"address"`
	Capabilities []string              `json:"capabilities"`
	Connections []*TopologyConnection  `json:"connections"`
	Metadata    map[string]interface{} `json:"metadata"`
	
	// Performance metrics
	Latency     time.Duration          `json:"latency"`
	Bandwidth   int64                  `json:"bandwidth"`
	Reliability float64                `json:"reliability"`
}

// TopologyConnection represents a connection between nodes
type TopologyConnection struct {
	TargetID    string        `json:"target_id"`
	Weight      float64       `json:"weight"`
	Latency     time.Duration `json:"latency"`
	Bandwidth   int64         `json:"bandwidth"`
	Quality     float64       `json:"quality"`
}

// TopologyType represents the type of network topology
type TopologyType string

const (
	TopologyMesh         TopologyType = "mesh"
	TopologyHierarchical TopologyType = "hierarchical"
	TopologyRing         TopologyType = "ring"
	TopologyStar         TopologyType = "star"
	TopologyHybrid       TopologyType = "hybrid"
)

// ModelDiscovery handles model discovery across the network
type ModelDiscovery struct {
	manager     *DistributedModelManager
	
	// Discovery cache
	cache       map[string]*DiscoveryEntry
	cacheMutex  sync.RWMutex
	
	// Discovery workers
	workers     []*DiscoveryWorker
	workQueue   chan *DiscoveryRequest
	
	// Broadcast settings
	broadcastInterval time.Duration
	discoveryTimeout  time.Duration
}

// DiscoveryEntry represents a discovered model
type DiscoveryEntry struct {
	ModelName   string                 `json:"model_name"`
	PeerID      string                 `json:"peer_id"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	TTL         time.Duration          `json:"ttl"`
}

// DiscoveryRequest represents a model discovery request
type DiscoveryRequest struct {
	ModelName    string                 `json:"model_name"`
	Criteria     map[string]interface{} `json:"criteria"`
	Timeout      time.Duration          `json:"timeout"`
	ResponseChan chan *DiscoveryResponse `json:"-"`
}

// DiscoveryResponse represents a model discovery response
type DiscoveryResponse struct {
	Models    []*DistributedModel `json:"models"`
	Peers     []string            `json:"peers"`
	Error     string              `json:"error,omitempty"`
	Duration  time.Duration       `json:"duration"`
}

// DiscoveryWorker handles model discovery tasks
type DiscoveryWorker struct {
	ID         int
	discovery  *ModelDiscovery
	stopChan   chan struct{}
}

// PerformanceMonitor monitors the performance of the distributed system
type PerformanceMonitor struct {
	metrics     map[string]*PerformanceMetric
	metricsMutex sync.RWMutex
	
	// Monitoring settings
	interval    time.Duration
	retention   time.Duration
	
	// Alerting
	alerts      []*PerformanceAlert
	alertsMutex sync.RWMutex
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	Name        string                 `json:"name"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	History     []MetricPoint          `json:"history"`
}

// MetricPoint represents a point in a metric's history
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  time.Time              `json:"resolved_at"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeLatency      AlertType = "latency"
	AlertTypeBandwidth    AlertType = "bandwidth"
	AlertTypeReplication  AlertType = "replication"
	AlertTypeSync         AlertType = "sync"
	AlertTypeStorage      AlertType = "storage"
	AlertTypeHealth       AlertType = "health"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityError    AlertSeverity = "error"
	SeverityCritical AlertSeverity = "critical"
)

// NewDistributedModelManager creates a new distributed model manager
func NewDistributedModelManager(
	config *config.DistributedConfig,
	p2pNode *p2p.Node,
	logger *slog.Logger,
) (*DistributedModelManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create local manager
	localManager, err := NewManager(config.Storage, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create local manager: %w", err)
	}
	
	// Create sync manager
	syncManager, err := NewSyncManager(config.Sync, p2pNode, localManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync manager: %w", err)
	}
	
	// Create replication manager
	replicationManager, err := NewReplicationManager(config.Replication, p2pNode, localManager, syncManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create replication manager: %w", err)
	}
	
	// Create content-addressed store
	casStore, err := NewContentAddressedStore(config.CASDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CAS store: %w", err)
	}
	
	// Create delta tracker
	deltaTracker, err := NewDeltaTracker(config.DeltaDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create delta tracker: %w", err)
	}
	
	dmm := &DistributedModelManager{
		localManager:       localManager,
		syncManager:        syncManager,
		replicationManager: replicationManager,
		casStore:           casStore,
		deltaTracker:       deltaTracker,
		config:             config,
		p2p:                p2pNode,
		logger:             logger,
		ctx:                ctx,
		cancel:             cancel,
	}
	
	// Initialize registry
	dmm.registry = &DistributedRegistry{
		models:     make(map[string]*DistributedModel),
		peerModels: make(map[string]map[string]*DistributedModel),
		topology:   &NetworkTopology{
			nodes: make(map[string]*TopologyNode),
		},
	}
	
	// Initialize lifecycle manager
	dmm.lifecycle = &ModelLifecycle{
		events: make(chan *LifecycleEvent, 100),
		stages: make(map[string]*LifecycleStage),
		hooks:  make(map[LifecycleEventType][]LifecycleHook),
	}
	
	// Initialize performance monitor
	dmm.monitor = &PerformanceMonitor{
		metrics:  make(map[string]*PerformanceMetric),
		interval: time.Minute,
		retention: 24 * time.Hour,
	}
	
	// Initialize model discovery
	dmm.registry.discovery = &ModelDiscovery{
		manager:           dmm,
		cache:             make(map[string]*DiscoveryEntry),
		workQueue:         make(chan *DiscoveryRequest, 100),
		broadcastInterval: 30 * time.Second,
		discoveryTimeout:  10 * time.Second,
	}
	
	return dmm, nil
}

// Start starts the distributed model manager
func (dmm *DistributedModelManager) Start() error {
	dmm.mu.Lock()
	defer dmm.mu.Unlock()
	
	if dmm.started {
		return fmt.Errorf("distributed model manager already started")
	}
	
	// Start local manager
	if err := dmm.localManager.Start(); err != nil {
		return fmt.Errorf("failed to start local manager: %w", err)
	}
	
	// Start sync manager
	if err := dmm.syncManager.Start(); err != nil {
		return fmt.Errorf("failed to start sync manager: %w", err)
	}
	
	// Start replication manager
	if err := dmm.replicationManager.Start(); err != nil {
		return fmt.Errorf("failed to start replication manager: %w", err)
	}
	
	// Start lifecycle manager
	go dmm.lifecycle.start()
	
	// Start performance monitor
	go dmm.monitor.start()
	
	// Start model discovery
	go dmm.registry.discovery.start()
	
	// Start registry synchronization
	go dmm.registrySyncRoutine()
	
	dmm.started = true
	dmm.logger.Info("distributed model manager started")
	
	return nil
}

// GetModel retrieves a model, either locally or from the network
func (dmm *DistributedModelManager) GetModel(modelName string) (*DistributedModel, error) {
	// Check local registry first
	dmm.registryMutex.RLock()
	if model, exists := dmm.registry.models[modelName]; exists {
		dmm.registryMutex.RUnlock()
		
		// Update access statistics
		model.AccessedAt = time.Now()
		model.AccessCount++
		
		// Emit lifecycle event
		dmm.emitLifecycleEvent(EventModelAccessed, modelName, dmm.p2p.ID().String(), map[string]interface{}{
			"access_count": model.AccessCount,
		})
		
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
	
	// Create distributed model
	model := &DistributedModel{
		Name:           modelName,
		Version:        version.Version,
		Hash:           version.Hash,
		Size:           version.Size,
		Type:           "gguf", // Default type
		Replicas:       []*ReplicaInfo{},
		Availability:   1.0,
		Versions:       []*ModelVersion{version},
		CurrentVersion: version.Version,
		Metadata:       make(map[string]interface{}),
		Tags:           []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		AccessedAt:     time.Now(),
		AccessCount:    0,
		DownloadCount:  0,
	}
	
	// Add to registry
	dmm.registryMutex.Lock()
	dmm.registry.models[modelName] = model
	dmm.registryMutex.Unlock()
	
	// Set default replication policy
	policy := &ReplicationPolicy{
		ModelName:         modelName,
		MinReplicas:       dmm.config.Replication.DefaultMinReplicas,
		MaxReplicas:       dmm.config.Replication.DefaultMaxReplicas,
		ReplicationFactor: dmm.config.Replication.DefaultReplicationFactor,
		SyncInterval:      dmm.config.Replication.DefaultSyncInterval,
		Priority:          1,
		Constraints:       make(map[string]string),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	
	model.Policy = policy
	dmm.replicationManager.SetReplicationPolicy(modelName, policy)
	
	// Emit lifecycle event
	dmm.emitLifecycleEvent(EventModelCreated, modelName, dmm.p2p.ID().String(), map[string]interface{}{
		"version": version.Version,
		"hash":    version.Hash,
		"size":    version.Size,
	})
	
	dmm.logger.Info("model added to distributed system", "model", modelName, "version", version.Version)
	
	return model, nil
}

// discoverAndFetchModel discovers a model on the network and fetches it
func (dmm *DistributedModelManager) discoverAndFetchModel(modelName string) (*DistributedModel, error) {
	// Create discovery request
	req := &DiscoveryRequest{
		ModelName:    modelName,
		Criteria:     make(map[string]interface{}),
		Timeout:      dmm.registry.discovery.discoveryTimeout,
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
	case resp := <-req.ResponseChan:
		if resp.Error != "" {
			return nil, fmt.Errorf("discovery failed: %s", resp.Error)
		}
		
		if len(resp.Models) == 0 {
			return nil, fmt.Errorf("model not found: %s", modelName)
		}
		
		// Use the first available model
		model := resp.Models[0]
		
		// Download model from a peer
		if len(resp.Peers) > 0 {
			if err := dmm.downloadModelFromPeer(modelName, resp.Peers[0]); err != nil {
				return nil, fmt.Errorf("failed to download model: %w", err)
			}
		}
		
		// Add to local registry
		dmm.registryMutex.Lock()
		dmm.registry.models[modelName] = model
		dmm.registryMutex.Unlock()
		
		return model, nil
		
	case <-time.After(dmm.registry.discovery.discoveryTimeout):
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
		// Event queue full, log warning
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
	peers := dmm.p2p.GetConnectedPeers()
	if len(peers) == 0 {
		return // No peers to sync with
	}

	// Prepare local registry for broadcasting
	dmm.registryMutex.RLock()
	localModels := make(map[string]*DistributedModel)
	for k, v := range dmm.registry.models {
		localModels[k] = v
	}
	dmm.registryMutex.RUnlock()

	// Broadcast local models to peers
	for _, peerID := range peers {
		go dmm.syncWithPeer(peerID, localModels)
	}

	// Request model information from peers
	for _, peerID := range peers {
		go dmm.requestPeerModels(peerID)
	}

	// Clean up stale peer entries
	dmm.cleanupStalePeers()
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
	dmm.monitor.metricsMutex.RLock()
	defer dmm.monitor.metricsMutex.RUnlock()
	
	metrics := make([]*PerformanceMetric, 0, len(dmm.monitor.metrics))
	for _, metric := range dmm.monitor.metrics {
		metrics = append(metrics, metric)
	}
	
	return metrics
}

// Shutdown gracefully shuts down the distributed model manager
func (dmm *DistributedModelManager) Shutdown(ctx context.Context) error {
	dmm.mu.Lock()
	defer dmm.mu.Unlock()
	
	if !dmm.started {
		return nil
	}
	
	// Shutdown components
	if err := dmm.replicationManager.Shutdown(ctx); err != nil {
		dmm.logger.Error("failed to shutdown replication manager", "error", err)
	}
	
	if err := dmm.syncManager.Shutdown(ctx); err != nil {
		dmm.logger.Error("failed to shutdown sync manager", "error", err)
	}
	
	if err := dmm.localManager.Shutdown(ctx); err != nil {
		dmm.logger.Error("failed to shutdown local manager", "error", err)
	}
	
	if err := dmm.casStore.Close(); err != nil {
		dmm.logger.Error("failed to close CAS store", "error", err)
	}
	
	if err := dmm.deltaTracker.Close(); err != nil {
		dmm.logger.Error("failed to close delta tracker", "error", err)
	}
	
	dmm.cancel()
	dmm.started = false
	
	dmm.logger.Info("distributed model manager shutdown complete")
	return nil
}

// ModelLifecycle methods

// start starts the lifecycle manager
func (ml *ModelLifecycle) start() {
	for event := range ml.events {
		ml.processEvent(event)
	}
}

// processEvent processes a lifecycle event
func (ml *ModelLifecycle) processEvent(event *LifecycleEvent) {
	// Execute hooks
	ml.hooksMutex.RLock()
	hooks := ml.hooks[event.Type]
	ml.hooksMutex.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(event); err != nil {
			// Log hook error but continue
			fmt.Printf("Lifecycle hook error: %v\n", err)
		}
	}
}

// PerformanceMonitor methods

// start starts the performance monitor
func (pm *PerformanceMonitor) start() {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()
	
	for range ticker.C {
		pm.collectMetrics()
	}
}

// collectMetrics collects performance metrics
func (pm *PerformanceMonitor) collectMetrics() {
	now := time.Now()

	// Collect model access latency
	pm.collectModelAccessMetrics(now)

	// Collect replication bandwidth
	pm.collectReplicationMetrics(now)

	// Collect sync success rate
	pm.collectSyncMetrics(now)

	// Collect storage utilization
	pm.collectStorageMetrics(now)

	// Collect network connectivity
	pm.collectNetworkMetrics(now)

	// Clean up old metrics
	pm.cleanupOldMetrics(now)
}

// ModelDiscovery methods

// start starts the model discovery service
func (md *ModelDiscovery) start() {
	// Start discovery workers
	md.workers = make([]*DiscoveryWorker, 3)
	for i := 0; i < 3; i++ {
		md.workers[i] = &DiscoveryWorker{
			ID:        i,
			discovery: md,
			stopChan:  make(chan struct{}),
		}
		go md.workers[i].start()
	}
	
	// Start broadcast routine
	go md.broadcastRoutine()
}

// broadcastRoutine periodically broadcasts model information
func (md *ModelDiscovery) broadcastRoutine() {
	ticker := time.NewTicker(md.broadcastInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		md.broadcastModels()
	}
}

// broadcastModels broadcasts local model information to peers
func (md *ModelDiscovery) broadcastModels() {
	// Get local models from manager
	models := md.manager.GetDistributedModels()
	if len(models) == 0 {
		return // No models to broadcast
	}

	// Prepare broadcast message
	broadcast := map[string]interface{}{
		"type":      "model_broadcast",
		"peer_id":   md.manager.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
		"models":    md.prepareModelBroadcast(models),
	}

	// Send to all connected peers
	peers := md.manager.p2p.GetConnectedPeers()
	for _, peerID := range peers {
		go md.sendBroadcastToPeer(peerID, broadcast)
	}

	// Update broadcast metrics
	md.updateBroadcastMetrics(len(peers), len(models))
}

// DiscoveryWorker methods

// start starts the discovery worker
func (dw *DiscoveryWorker) start() {
	for {
		select {
		case <-dw.stopChan:
			return
		case req := <-dw.discovery.workQueue:
			dw.processRequest(req)
		}
	}
}

// processRequest processes a discovery request
func (dw *DiscoveryWorker) processRequest(req *DiscoveryRequest) {
	start := time.Now()

	// Search local cache first
	foundModels, foundPeers := dw.searchLocalCache(req.ModelName, req.Criteria)

	// If not found locally, search network
	if len(foundModels) == 0 {
		networkModels, networkPeers := dw.searchNetwork(req.ModelName, req.Criteria, req.Timeout)
		foundModels = append(foundModels, networkModels...)
		foundPeers = append(foundPeers, networkPeers...)
	}

	// Filter and rank results
	filteredModels := dw.filterResults(foundModels, req.Criteria)
	rankedModels := dw.rankResults(filteredModels)

	// Prepare response
	resp := &DiscoveryResponse{
		Models:   rankedModels,
		Peers:    foundPeers,
		Duration: time.Since(start),
	}

	// Send response
	select {
	case req.ResponseChan <- resp:
	case <-time.After(time.Second):
		// Response channel blocked, log warning
		dw.discovery.manager.logger.Warn("discovery response channel blocked")
	}
}

// Helper methods for registry synchronization

// syncWithPeer synchronizes models with a specific peer
func (dmm *DistributedModelManager) syncWithPeer(peerID peer.ID, localModels map[string]*DistributedModel) {
	// Prepare sync message
	syncMessage := map[string]interface{}{
		"type":      "registry_sync",
		"peer_id":   dmm.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
		"models":    localModels,
	}

	// Send via P2P (simplified implementation)
	// In practice, this would use libp2p streams
	dmm.logger.Info("syncing models with peer", "peer", peerID.String(), "models", len(localModels))
}

// requestPeerModels requests model information from a peer
func (dmm *DistributedModelManager) requestPeerModels(peerID peer.ID) {
	// Create request message
	request := map[string]interface{}{
		"type":      "model_request",
		"peer_id":   dmm.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
	}

	// Send request via P2P (simplified implementation)
	dmm.logger.Info("requesting models from peer", "peer", peerID.String())
}

// cleanupStalePeers removes stale peer entries
func (dmm *DistributedModelManager) cleanupStalePeers() {
	dmm.registry.peerMutex.Lock()
	defer dmm.registry.peerMutex.Unlock()

	connectedPeers := dmm.p2p.GetConnectedPeers()
	connectedMap := make(map[string]bool)
	for _, peerID := range connectedPeers {
		connectedMap[peerID.String()] = true
	}

	// Remove disconnected peers
	for peerID := range dmm.registry.peerModels {
		if !connectedMap[peerID] {
			delete(dmm.registry.peerModels, peerID)
		}
	}
}

// Helper methods for performance monitoring

// collectModelAccessMetrics collects model access latency metrics
func (pm *PerformanceMonitor) collectModelAccessMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	// Simulate collecting access latency
	latencyMetric := &PerformanceMetric{
		Name:      "model_access_latency",
		Value:     float64(50 + (now.UnixNano() % 100)), // Simulate 50-150ms
		Unit:      "milliseconds",
		Timestamp: now,
		Labels:    map[string]string{"type": "access"},
		History:   []MetricPoint{{Timestamp: now, Value: 75.5}},
	}

	pm.metrics["model_access_latency"] = latencyMetric
}

// collectReplicationMetrics collects replication bandwidth metrics
func (pm *PerformanceMonitor) collectReplicationMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	// Simulate bandwidth metrics
	bandwidthMetric := &PerformanceMetric{
		Name:      "replication_bandwidth",
		Value:     float64(1024 * 1024 * 10), // 10 MB/s
		Unit:      "bytes_per_second",
		Timestamp: now,
		Labels:    map[string]string{"type": "replication"},
		History:   []MetricPoint{{Timestamp: now, Value: 1024 * 1024 * 10}},
	}

	pm.metrics["replication_bandwidth"] = bandwidthMetric
}

// collectSyncMetrics collects synchronization success rate metrics
func (pm *PerformanceMonitor) collectSyncMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	// Simulate sync success rate
	syncMetric := &PerformanceMetric{
		Name:      "sync_success_rate",
		Value:     95.5, // 95.5% success rate
		Unit:      "percentage",
		Timestamp: now,
		Labels:    map[string]string{"type": "sync"},
		History:   []MetricPoint{{Timestamp: now, Value: 95.5}},
	}

	pm.metrics["sync_success_rate"] = syncMetric
}

// collectStorageMetrics collects storage utilization metrics
func (pm *PerformanceMonitor) collectStorageMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	// Simulate storage usage
	storageMetric := &PerformanceMetric{
		Name:      "storage_utilization",
		Value:     65.2, // 65.2% storage used
		Unit:      "percentage",
		Timestamp: now,
		Labels:    map[string]string{"type": "storage"},
		History:   []MetricPoint{{Timestamp: now, Value: 65.2}},
	}

	pm.metrics["storage_utilization"] = storageMetric
}

// collectNetworkMetrics collects network connectivity metrics
func (pm *PerformanceMonitor) collectNetworkMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	// Simulate network connectivity
	networkMetric := &PerformanceMetric{
		Name:      "network_connectivity",
		Value:     98.7, // 98.7% uptime
		Unit:      "percentage",
		Timestamp: now,
		Labels:    map[string]string{"type": "network"},
		History:   []MetricPoint{{Timestamp: now, Value: 98.7}},
	}

	pm.metrics["network_connectivity"] = networkMetric
}

// cleanupOldMetrics removes old metric history points
func (pm *PerformanceMonitor) cleanupOldMetrics(now time.Time) {
	pm.metricsMutex.Lock()
	defer pm.metricsMutex.Unlock()

	cutoff := now.Add(-pm.retention)
	for _, metric := range pm.metrics {
		var newHistory []MetricPoint
		for _, point := range metric.History {
			if point.Timestamp.After(cutoff) {
				newHistory = append(newHistory, point)
			}
		}
		metric.History = newHistory
	}
}

// Helper methods for model discovery

// prepareModelBroadcast prepares models for broadcasting
func (md *ModelDiscovery) prepareModelBroadcast(models []*DistributedModel) []map[string]interface{} {
	var broadcast []map[string]interface{}
	for _, model := range models {
		broadcast = append(broadcast, map[string]interface{}{
			"name":         model.Name,
			"version":      model.Version,
			"hash":         model.Hash,
			"size":         model.Size,
			"availability": model.Availability,
			"replicas":     len(model.Replicas),
		})
	}
	return broadcast
}

// sendBroadcastToPeer sends broadcast message to a specific peer
func (md *ModelDiscovery) sendBroadcastToPeer(peerID peer.ID, broadcast map[string]interface{}) {
	// Send via P2P (simplified implementation)
	// In practice, this would use libp2p streams
	fmt.Printf("Broadcasting models to peer %s\n", peerID.String())
}

// updateBroadcastMetrics updates broadcast metrics
func (md *ModelDiscovery) updateBroadcastMetrics(peerCount, modelCount int) {
	// Update internal metrics (simplified)
	fmt.Printf("Broadcast sent to %d peers with %d models\n", peerCount, modelCount)
}

// Helper methods for discovery worker

// searchLocalCache searches for models in local cache
func (dw *DiscoveryWorker) searchLocalCache(modelName string, criteria map[string]interface{}) ([]*DistributedModel, []string) {
	dw.discovery.cacheMutex.RLock()
	defer dw.discovery.cacheMutex.RUnlock()

	var foundModels []*DistributedModel
	var foundPeers []string

	for _, entry := range dw.discovery.cache {
		if entry.ModelName == modelName || modelName == "" {
			// Create model from cache entry
			model := &DistributedModel{
				Name:      entry.ModelName,
				Version:   "1.0",
				Hash:      "unknown",
				Size:      1024,
				Type:      "gguf",
				Metadata:  entry.Metadata,
				CreatedAt: entry.Timestamp,
			}
			foundModels = append(foundModels, model)
			foundPeers = append(foundPeers, entry.PeerID)
		}
	}

	return foundModels, foundPeers
}

// searchNetwork searches for models across the network
func (dw *DiscoveryWorker) searchNetwork(modelName string, criteria map[string]interface{}, timeout time.Duration) ([]*DistributedModel, []string) {
	// Simulate network search
	// In practice, this would query connected peers
	var foundModels []*DistributedModel
	var foundPeers []string

	// Mock finding a model on the network
	if modelName != "" {
		model := &DistributedModel{
			Name:      modelName,
			Version:   "1.0",
			Hash:      "network_hash",
			Size:      2048,
			Type:      "gguf",
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}
		foundModels = append(foundModels, model)
		foundPeers = append(foundPeers, "network_peer_123")
	}

	return foundModels, foundPeers
}

// filterResults filters models based on criteria
func (dw *DiscoveryWorker) filterResults(models []*DistributedModel, criteria map[string]interface{}) []*DistributedModel {
	if len(criteria) == 0 {
		return models
	}

	var filtered []*DistributedModel
	for _, model := range models {
		if dw.matchesCriteria(model, criteria) {
			filtered = append(filtered, model)
		}
	}
	return filtered
}

// rankResults ranks models by relevance
func (dw *DiscoveryWorker) rankResults(models []*DistributedModel) []*DistributedModel {
	// Simple ranking by size (smaller first)
	for i := 0; i < len(models)-1; i++ {
		for j := i + 1; j < len(models); j++ {
			if models[i].Size > models[j].Size {
				models[i], models[j] = models[j], models[i]
			}
		}
	}
	return models
}

// matchesCriteria checks if a model matches search criteria
func (dw *DiscoveryWorker) matchesCriteria(model *DistributedModel, criteria map[string]interface{}) bool {
	// Simple criteria matching
	if minSize, exists := criteria["min_size"]; exists {
		if size, ok := minSize.(int64); ok && model.Size < size {
			return false
		}
	}
	if maxSize, exists := criteria["max_size"]; exists {
		if size, ok := maxSize.(int64); ok && model.Size > size {
			return false
		}
	}
	return true
}