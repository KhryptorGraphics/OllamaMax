package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/protocols"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
	"github.com/libp2p/go-libp2p/core/peer"
)

// DistributedModelManager extends Ollama's model management with distributed capabilities
type DistributedModelManager struct {
	// Core components
	localManager       *Manager
	syncManager        *SyncManager
	replicationManager *ReplicationManager
	casStore           *ContentAddressedStore
	deltaTracker       *DeltaTracker

	// Advanced components for large-scale model distribution
	shardManager         *ModelShardManager
	memoryDistributor    *MemoryAwareDistributor
	transferOrchestrator *ChunkTransferOrchestrator
	shardRegistry        *ShardRegistry
	modelLoader          *ShardedModelLoader
	p2pTransferEngine    *P2PTransferEngine
	integrityVerifier    *IntegrityVerifier
	versionManager       *BasicVersionManager
	shardProtocolHandler *protocols.ShardProtocolHandler

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
	monitor     *PerformanceMonitor
	distMonitor *DistributedModelPerformanceMonitor

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

	// Network topology
	topology *NetworkTopology

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
	Type    string `json:"type"`

	// Distributed information
	Replicas     []*ReplicaInfo `json:"replicas"`
	Availability float64        `json:"availability"`

	// Version tracking
	Versions       []*ModelVersion `json:"versions"`
	CurrentVersion string          `json:"current_version"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
	Tags     []string               `json:"tags"`

	// Lifecycle
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AccessedAt time.Time `json:"accessed_at"`

	// Performance metrics
	AccessCount   int64 `json:"access_count"`
	DownloadCount int64 `json:"download_count"`

	// Replication policy
	Policy *ReplicationPolicy `json:"policy"`

	// Sync state
	SyncState *SyncState `json:"sync_state"`
}

// ModelLifecycle manages the lifecycle of distributed models
type ModelLifecycle struct {
	events      chan *LifecycleEvent
	eventsMutex sync.RWMutex

	// Lifecycle stages
	stages      map[string]*LifecycleStage
	stagesMutex sync.RWMutex

	// Hooks
	hooks      map[LifecycleEventType][]LifecycleHook
	hooksMutex sync.RWMutex
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
	EventModelCreated    LifecycleEventType = "model_created"
	EventModelUpdated    LifecycleEventType = "model_updated"
	EventModelDeleted    LifecycleEventType = "model_deleted"
	EventModelAccessed   LifecycleEventType = "model_accessed"
	EventModelReplicated LifecycleEventType = "model_replicated"
	EventModelSynced     LifecycleEventType = "model_synced"
	EventModelCorrupted  LifecycleEventType = "model_corrupted"
	EventModelHealed     LifecycleEventType = "model_healed"
)

// LifecycleStage represents a stage in the model lifecycle
type LifecycleStage struct {
	Name      string                 `json:"name"`
	ModelName string                 `json:"model_name"`
	Status    StageStatus            `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Progress  float64                `json:"progress"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
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
	nodes      map[string]*TopologyNode
	nodesMutex sync.RWMutex

	// Topology metadata
	Type         TopologyType `json:"type"`
	Depth        int          `json:"depth"`
	Diameter     int          `json:"diameter"`
	Connectivity float64      `json:"connectivity"`

	// Performance metrics
	avgLatency   time.Duration
	avgBandwidth int64
}

// TopologyNode represents a node in the network topology
type TopologyNode struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Capabilities []string               `json:"capabilities"`
	Connections  []*TopologyConnection  `json:"connections"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Performance metrics
	Latency     time.Duration `json:"latency"`
	Bandwidth   int64         `json:"bandwidth"`
	Reliability float64       `json:"reliability"`
}

// TopologyConnection represents a connection between nodes
type TopologyConnection struct {
	TargetID  string        `json:"target_id"`
	Weight    float64       `json:"weight"`
	Latency   time.Duration `json:"latency"`
	Bandwidth int64         `json:"bandwidth"`
	Quality   float64       `json:"quality"`
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
	manager *DistributedModelManager

	// Discovery cache
	cache      map[string]*DiscoveryEntry
	cacheMutex sync.RWMutex

	// Discovery workers
	workers   []*DiscoveryWorker
	workQueue chan *DiscoveryRequest

	// Broadcast settings
	broadcastInterval time.Duration
	discoveryTimeout  time.Duration
}

// DiscoveryEntry represents a discovered model
type DiscoveryEntry struct {
	ModelName string                 `json:"model_name"`
	PeerID    string                 `json:"peer_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       time.Duration          `json:"ttl"`
}

// DiscoveryRequest represents a model discovery request
type DiscoveryRequest struct {
	ModelName    string                  `json:"model_name"`
	Criteria     map[string]interface{}  `json:"criteria"`
	Timeout      time.Duration           `json:"timeout"`
	ResponseChan chan *DiscoveryResponse `json:"-"`
}

// DiscoveryResponse represents a model discovery response
type DiscoveryResponse struct {
	Models   []*DistributedModel `json:"models"`
	Peers    []string            `json:"peers"`
	Error    string              `json:"error,omitempty"`
	Duration time.Duration       `json:"duration"`
}

// DiscoveryWorker handles model discovery tasks
type DiscoveryWorker struct {
	ID        int
	discovery *ModelDiscovery
	stopChan  chan struct{}
}

// PerformanceMonitor monitors the performance of the distributed system
type PerformanceMonitor struct {
	metrics      map[string]*PerformanceMetric
	metricsMutex sync.RWMutex

	// Monitoring settings
	interval  time.Duration
	retention time.Duration

	// Alerting
	alerts      []*PerformanceAlert
	alertsMutex sync.RWMutex
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
	History   []MetricPoint     `json:"history"`
}

// MetricPoint represents a point in a metric's history
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID         string                 `json:"id"`
	Type       AlertType              `json:"type"`
	Severity   AlertSeverity          `json:"severity"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt time.Time              `json:"resolved_at"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeLatency     AlertType = "latency"
	AlertTypeBandwidth   AlertType = "bandwidth"
	AlertTypeReplication AlertType = "replication"
	AlertTypeSync        AlertType = "sync"
	AlertTypeStorage     AlertType = "storage"
	AlertTypeHealth      AlertType = "health"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
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

	// Initialize advanced components
	shardingConfig := DefaultShardingConfig()
	shardManager := NewModelShardManager(shardingConfig)

	distributorConfig := DefaultDistributorConfig()
	memoryDistributor := NewMemoryAwareDistributor(distributorConfig)

	// Create LocalFileStore for shard storage
	localFileStore := protocols.NewLocalFileStore(config.Storage.BasePath + "/shards")

	// Initialize P2P file transfer client for real remote transfers with LocalFileStore
	p2pClient, err := protocols.NewFileTransferClient(p2pNode.Host(), localFileStore, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P client: %w", err)
	}

	// Register a real FileTransferHandler on the host with LocalFileStore
	fileHandler := protocols.NewFileTransferHandler(localFileStore, protocols.DefaultFileTransferConfig())
	p2pNode.Host().SetStreamHandler(protocols.FileTransferProtocol, fileHandler.HandleStream)

	// Initialize P2P transfer engine with real FileTransferClient
	transferConfig := &TransferConfig{
		ChunkSize:           1024 * 1024, // 1MB chunks
		MaxConcurrentChunks: 10,
		RetryAttempts:       3,
		TransferTimeout:     5 * time.Minute,
		VerificationTimeout: 30 * time.Second,
		EnableCompression:   true,
		EnableEncryption:    false,
		CacheChunks:         true,
		MaxCacheSize:        100 * 1024 * 1024, // 100MB cache
		StorageDir:          config.Storage.BasePath + "/chunks",
	}
	p2pTransferEngine := NewP2PTransferEngine(transferConfig, p2pClient)

	// Initialize integrity verifier
	verificationConfig := &VerificationConfig{
		HashAlgorithms: []HashAlgorithm{HashAlgorithmSHA256, HashAlgorithmSHA512},
		EnableCaching:  true,
		CacheTimeout:   10 * time.Minute,
		MaxCacheSize:   1000,
		BufferSize:     8192,
	}
	integrityVerifier := NewIntegrityVerifier(verificationConfig)

	// Initialize version manager
	versionConfig := &VersionConfig{
		StoragePath:       config.Storage.BasePath,
		MaxVersions:       10,
		RetentionPeriod:   30 * 24 * time.Hour,
		EnableAutoCleanup: true,
		CleanupInterval:   24 * time.Hour,
	}
	versionManager := NewBasicVersionManager(versionConfig)

	// Initialize shard registry
	shardRegistry := NewShardRegistry(p2pNode, logger)

	// Initialize chunk transfer orchestrator
	orchestratorConfig := DefaultOrchestratorConfig()
	transferOrchestrator := NewChunkTransferOrchestrator(
		p2pTransferEngine,
		integrityVerifier,
		shardRegistry,
		orchestratorConfig,
	)

	// Wire FileTransferClient to orchestrator for real P2P chunk streaming
	transferOrchestrator.SetFileTransferClient(p2pClient)

	// Initialize sharded model loader with P2P client for real remote transfers
	modelLoader := NewShardedModelLoader(
		shardRegistry,
		shardManager,
		p2pClient,
		p2pNode.ID().String(), // Use actual peer ID from p2pNode
		logger,
	)

	// Initialize distributed performance monitor
	distMonitor := NewDistributedModelPerformanceMonitor(logger)

	// Initialize shard protocol handler for P2P shard announcements
	shardProtocolHandler := protocols.NewShardProtocolHandler(p2pNode.Host(), logger, shardRegistry)

	dmm := &DistributedModelManager{
		localManager:         localManager,
		syncManager:          syncManager,
		replicationManager:   replicationManager,
		casStore:             casStore,
		deltaTracker:         deltaTracker,
		shardManager:         shardManager,
		memoryDistributor:    memoryDistributor,
		transferOrchestrator: transferOrchestrator,
		shardRegistry:        shardRegistry,
		modelLoader:          modelLoader,
		p2pTransferEngine:    p2pTransferEngine,
		integrityVerifier:    integrityVerifier,
		versionManager:       versionManager,
		shardProtocolHandler: shardProtocolHandler,
		config:               config,
		p2p:                  p2pNode,
		logger:               logger,
		ctx:                  ctx,
		cancel:               cancel,
		distMonitor:          distMonitor,
	}

	// Initialize registry
	dmm.registry = &DistributedRegistry{
		models:     make(map[string]*DistributedModel),
		peerModels: make(map[string]map[string]*DistributedModel),
		topology: &NetworkTopology{
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
		metrics:   make(map[string]*PerformanceMetric),
		interval:  time.Minute,
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

	// Wire orchestrator to model loader for remote shard loading
	modelLoader.SetOrchestrator(transferOrchestrator)

	return dmm, nil
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

	// Decrement CAS reference for the model hash
	if model != nil && dmm.casStore != nil && model.Hash != "" {
		_ = dmm.casStore.DecrementReference(model.Hash)
	}

	// Clear replication policy if set
	if dmm.replicationManager != nil {
		if _, exists := dmm.replicationManager.GetReplicationPolicy(modelName); exists {
			_ = dmm.replicationManager.SetReplicationPolicy(modelName, &ReplicationPolicy{
				ModelName:         modelName,
				MinReplicas:       0,
				MaxReplicas:       0,
				ReplicationFactor: 0,
				SyncInterval:      time.Hour,
				Priority:          0,
			})
		}
	}

	// Emit lifecycle event
	dmm.emitLifecycleEvent(EventModelDeleted, modelName, dmm.p2p.ID().String(), map[string]interface{}{})
	return nil
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

// AddModel adds a model to the distributed system with optional sharding for large models
func (dmm *DistributedModelManager) AddModel(modelName, modelPath string) (*DistributedModel, error) {
	// Create model version
	version, err := dmm.syncManager.CreateModelVersion(modelName, modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create model version: %w", err)
	}

	// Check if model needs sharding (>10GB)
	var shardPlan *ShardPlan
	if version.Size > 10*1024*1024*1024 { // 10GB threshold
		// Get node capabilities
		nodeCapabilities := dmm.getNodeCapabilities()

		// Create shard plan
		shardPlan, err = dmm.shardManager.CreateShardPlan(
			modelName,
			modelPath,
			version.Size,
			nodeCapabilities,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create shard plan: %w", err)
		}

		// Validate shard plan
		if err := dmm.shardManager.ValidateShardPlan(shardPlan); err != nil {
			return nil, fmt.Errorf("shard plan validation failed: %w", err)
		}

		// Create distribution plan
		modelAnalysis := &partitioning.ModelAnalysis{
			ModelName:     modelName,
			ParameterSize: version.Size / 1000000000, // Convert to billions of parameters (approximate)
			MemoryReqs:    &partitioning.MemoryRequirements{TotalRequired: version.Size},
		}

		distPlan, err := dmm.memoryDistributor.CalculateDistributionPlan(
			modelAnalysis,
			shardPlan,
			nodeCapabilities,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create distribution plan: %w", err)
		}

		// Validate memory requirements
		if err := dmm.memoryDistributor.ValidateMemoryRequirements(distPlan); err != nil {
			return nil, fmt.Errorf("memory validation failed: %w", err)
		}

		// Execute shard distribution
		if err := dmm.executeShardDistribution(shardPlan, distPlan); err != nil {
			return nil, fmt.Errorf("shard distribution failed: %w", err)
		}

		// Register shards in the registry
		for _, shard := range shardPlan.Shards {
			if err := dmm.shardRegistry.RegisterModelShard(shard); err != nil {
				dmm.logger.Warn("failed to register shard", "shard", shard.ID, "error", err)
			}
		}
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

	// Store shard plan reference if created
	if shardPlan != nil {
		model.Metadata["shard_plan_id"] = shardPlan.ID
		model.Metadata["total_shards"] = shardPlan.TotalShards
		model.Metadata["sharding_strategy"] = shardPlan.Strategy
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

// downloadModelFromPeer downloads a model from a specific peer, handling sharded models
func (dmm *DistributedModelManager) downloadModelFromPeer(modelName, peerID string) error {
	// Check if model is sharded
	shardPlan, err := dmm.shardManager.GetShardPlan(modelName)
	if err == nil && shardPlan != nil {
		// Model is sharded, use chunk transfer orchestrator
		return dmm.downloadShardedModel(modelName, shardPlan, peerID)
	}

	// Use the local manager for non-sharded models
	_, err = dmm.localManager.DownloadModel(modelName, peerID)
	return err
}

// downloadShardedModel downloads a sharded model using the chunk transfer orchestrator
func (dmm *DistributedModelManager) downloadShardedModel(modelName string, shardPlan *ShardPlan, peerID string) error {
	ctx := context.Background()

	// Transfer each shard
	for _, shard := range shardPlan.Shards {
		// Find source node for this shard
		sourceNode := peerID
		if len(shard.NodeAssignments) > 0 {
			// Prefer assigned nodes
			sourceNode = shard.NodeAssignments[0]
		}

		// Orchestrate shard transfer
		transfer, err := dmm.transferOrchestrator.OrchestateShardTransfer(
			ctx,
			shard,
			sourceNode,
			dmm.p2p.ID().String(),
			TransferPriorityNormal,
		)
		if err != nil {
			return fmt.Errorf("failed to transfer shard %s: %w", shard.ID, err)
		}

		// Wait for transfer completion
		if err := dmm.waitForTransfer(transfer.ID); err != nil {
			return fmt.Errorf("shard transfer %s failed: %w", transfer.ID, err)
		}
	}

	// Load the assembled model
	if err := dmm.modelLoader.LoadModel(modelName, shardPlan.Shards); err != nil {
		return fmt.Errorf("failed to load assembled model: %w", err)
	}

	return nil
}

// ShardModel creates a sharded version of a large model
func (dmm *DistributedModelManager) ShardModel(modelName, modelPath string, modelSize int64) (*ShardPlan, error) {
	nodeCapabilities := dmm.getNodeCapabilities()
	return dmm.shardManager.CreateShardPlan(modelName, modelPath, modelSize, nodeCapabilities)
}

// DistributeShards coordinates shard placement across nodes
func (dmm *DistributedModelManager) DistributeShards(shardPlan *ShardPlan) error {
	// Create model analysis
	modelAnalysis := &partitioning.ModelAnalysis{
		ModelName:     shardPlan.ModelID,
		ParameterSize: shardPlan.TotalModelSize / 1000000000, // Convert to billions of parameters (approximate)
		MemoryReqs:    &partitioning.MemoryRequirements{TotalRequired: shardPlan.TotalModelSize},
	}

	// Get node capabilities
	nodeCapabilities := dmm.getNodeCapabilities()

	// Create distribution plan
	distPlan, err := dmm.memoryDistributor.CalculateDistributionPlan(
		modelAnalysis,
		shardPlan,
		nodeCapabilities,
	)
	if err != nil {
		return err
	}

	// Execute distribution
	return dmm.executeShardDistribution(shardPlan, distPlan)
}

// AssembleModel reconstructs a model from its shards
func (dmm *DistributedModelManager) AssembleModel(modelName string) error {
	shardPlan, err := dmm.shardManager.GetShardPlan(modelName)
	if err != nil {
		return fmt.Errorf("shard plan not found: %w", err)
	}

	return dmm.modelLoader.LoadModel(modelName, shardPlan.Shards)
}

// ShardRegistry returns the shard registry
func (dmm *DistributedModelManager) ShardRegistry() *ShardRegistry {
	return dmm.shardRegistry
}

// ExecuteShardDistribution executes the distribution of shards (exported for inference engine)
func (dmm *DistributedModelManager) ExecuteShardDistribution(shardPlan *ShardPlan, distPlan *DistributionPlan) error {
	return dmm.executeShardDistribution(shardPlan, distPlan)
}

// executeShardDistribution executes the distribution of shards across nodes
func (dmm *DistributedModelManager) executeShardDistribution(shardPlan *ShardPlan, distPlan *DistributionPlan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	var transferErrors []error
	var errorMutex sync.Mutex

	// Execute transfers for each shard assignment
	for nodeID, shardIDs := range distPlan.ShardAssignments {
		for _, shardID := range shardIDs {
			// Find the shard
			var targetShard *ModelShard
			for _, shard := range shardPlan.Shards {
				if shard.ID == shardID {
					targetShard = shard
					break
				}
			}

			if targetShard == nil {
				return fmt.Errorf("shard %s not found", shardID)
			}

			wg.Add(1)
			go func(shard *ModelShard, targetNodeID, shardIdentifier string) {
				defer wg.Done()

				// Orchestrate transfer to target node
				transfer, err := dmm.transferOrchestrator.OrchestateShardTransfer(
					ctx,
					shard,
					dmm.p2p.ID().String(), // Source is local node
					targetNodeID,
					TransferPriorityNormal,
				)
				if err != nil {
					dmm.logger.Error("failed to start shard transfer", "shard", shardIdentifier, "node", targetNodeID, "error", err)
					errorMutex.Lock()
					transferErrors = append(transferErrors, fmt.Errorf("failed to transfer shard %s to node %s: %w", shardIdentifier, targetNodeID, err))
					errorMutex.Unlock()
					return
				}

				// Wait for transfer completion
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						dmm.logger.Error("shard transfer timeout", "shard", shardIdentifier, "node", targetNodeID)
						errorMutex.Lock()
						transferErrors = append(transferErrors, fmt.Errorf("timeout transferring shard %s to node %s", shardIdentifier, targetNodeID))
						errorMutex.Unlock()
						return
					case <-ticker.C:
						status, err := dmm.transferOrchestrator.GetTransferStatus(transfer.ID)
						if err != nil {
							continue // Retry on status check error
						}

						switch status.Status {
						case TransferStatusCompleted:
							// Transfer successful - update registry
							assembledPath := dmm.transferOrchestrator.GetAssembledShardPath(shardIdentifier)

							// Build shard location for target node
							targetPeerID, err := peer.Decode(targetNodeID)
							if err != nil {
								dmm.logger.Warn("failed to decode target peer ID", "node", targetNodeID, "error", err)
								targetPeerID = peer.ID("")
							}

							location := protocols.ShardNodeLocation{
								NodeID:      targetNodeID,
								PeerID:      targetPeerID,
								IsAvailable: true,
								IsLoaded:    true,
								IsLocal:     targetNodeID == dmm.p2p.ID().String(),
								StoragePath: assembledPath,
								LastSeen:    time.Now(),
							}

							// Register shard in registry for target node
							err = dmm.shardRegistry.RegisterShard(shardIdentifier, shard.ModelID, location)
							if err != nil {
								dmm.logger.Warn("failed to register shard in registry", "shard", shardIdentifier, "node", targetNodeID, "error", err)
							}

							// Update shard status
							statusMsg := protocols.ShardStatusMessage{
								ShardID:      shardIdentifier,
								ModelName:    shard.ModelID,
								NodeID:       targetNodeID,
								IsAvailable:  true,
								IsLoaded:     true,
								StoragePath:  assembledPath,
								Size:         shard.Size,
								Checksum:     shard.Checksum,
								LastAccessed: time.Now(),
							}

							err = dmm.shardRegistry.UpdateShardStatus(shardIdentifier, statusMsg)
							if err != nil {
								dmm.logger.Warn("failed to update shard status", "shard", shardIdentifier, "node", targetNodeID, "error", err)
							}

							// Broadcast shard availability
							dmm.broadcastShardAvailability(shardIdentifier, targetNodeID)

							dmm.logger.Info("shard transfer completed successfully",
								"shard", shardIdentifier,
								"node", targetNodeID,
								"transfer_id", transfer.ID)
							return

						case TransferStatusFailed:
							dmm.logger.Error("shard transfer failed", "shard", shardIdentifier, "node", targetNodeID, "error", status.LastError)
							errorMutex.Lock()
							transferErrors = append(transferErrors, fmt.Errorf("transfer failed for shard %s to node %s: %s", shardIdentifier, targetNodeID, status.LastError))
							errorMutex.Unlock()
							return

						case TransferStatusCancelled:
							dmm.logger.Warn("shard transfer cancelled", "shard", shardIdentifier, "node", targetNodeID)
							return
						}
					}
				}
			}(targetShard, nodeID, shardID)
		}
	}

	// Wait for all transfers to complete
	wg.Wait()

	// Check if all transfers succeeded
	if len(transferErrors) > 0 {
		// Log all errors
		for _, err := range transferErrors {
			dmm.logger.Error("shard distribution error", "error", err)
		}
		return fmt.Errorf("shard distribution completed with %d errors", len(transferErrors))
	}

	// Verify all shards are properly registered
	for nodeID, shardIDs := range distPlan.ShardAssignments {
		for _, shardID := range shardIDs {
			locations, err := dmm.shardRegistry.LocateShard(shardID)
			if err != nil || len(locations) == 0 {
				dmm.logger.Warn("shard not found in registry after distribution", "shard", shardID, "node", nodeID)
			}
		}
	}

	return nil
}

// getNodeCapabilities retrieves capabilities of all nodes in the cluster
func (dmm *DistributedModelManager) getNodeCapabilities() []*NodeCapabilities {
	peerIDs := dmm.p2p.GetConnectedPeers()
	capabilities := make([]*NodeCapabilities, 0, len(peerIDs)+1)

	// Add local node capabilities
	localCap := &NodeCapabilities{
		NodeID:            dmm.p2p.ID().String(),
		AvailableMemory:   32 * 1024 * 1024 * 1024, // 32GB default
		TotalMemory:       64 * 1024 * 1024 * 1024, // 64GB default
		GPUMemory:         16 * 1024 * 1024 * 1024, // 16GB default
		CPUCores:          16,
		GPUCount:          1,
		NetworkBandwidth:  1000 * 1024 * 1024 / 8,    // 1Gbps
		StorageCapacity:   1024 * 1024 * 1024 * 1024, // 1TB
		ComputeCapability: 8.6,                       // Default CUDA capability
		IsHealthy:         true,
	}
	capabilities = append(capabilities, localCap)

	// Add peer capabilities (would query peers in real implementation)
	for _, peerID := range peerIDs {
		peerCap := &NodeCapabilities{
			NodeID:            peerID.String(),
			AvailableMemory:   16 * 1024 * 1024 * 1024, // 16GB default
			TotalMemory:       32 * 1024 * 1024 * 1024, // 32GB default
			GPUMemory:         8 * 1024 * 1024 * 1024,  // 8GB default
			CPUCores:          8,
			GPUCount:          1,
			NetworkBandwidth:  100 * 1024 * 1024 / 8,    // 100Mbps
			StorageCapacity:   500 * 1024 * 1024 * 1024, // 500GB
			ComputeCapability: 7.5,
			IsHealthy:         true,
		}
		capabilities = append(capabilities, peerCap)
	}

	return capabilities
}

// waitForTransfer waits for a transfer to complete
func (dmm *DistributedModelManager) waitForTransfer(transferID string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ticker.C:
			transfer, err := dmm.transferOrchestrator.GetTransferStatus(transferID)
			if err != nil {
				return err
			}

			switch transfer.Status {
			case TransferStatusCompleted:
				return nil
			case TransferStatusFailed:
				return fmt.Errorf("transfer failed: %s", transfer.LastError)
			}

		case <-timeout:
			return fmt.Errorf("transfer timeout")
		}
	}
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
	// First, synchronize shard registry with P2P protocol
	dmm.syncShardRegistry()

	// Then, synchronize model registry with peers
	// Get connected peers
	peerIDs := dmm.p2p.GetConnectedPeers()
	if len(peerIDs) == 0 {
		return // No peers to sync with
	}

	var peers []string
	for _, peerID := range peerIDs {
		peers = append(peers, peerID.String())
	}

	// Prepare local registry for broadcasting
	dmm.registryMutex.RLock()
	localModels := make(map[string]*DistributedModel)
	for k, v := range dmm.registry.models {
		localModels[k] = v
	}
	dmm.registryMutex.RUnlock()

	// Broadcast local models to peers
	for _, peerStr := range peers {
		go dmm.syncWithPeer(peerStr, localModels)
	}

	// Request model information from peers
	for _, peerStr := range peers {
		go dmm.requestPeerModels(peerStr)
	}

	// Clean up stale peer entries
	dmm.cleanupStalePeers()
}

// GetReplicationSummary exposes replication manager summary
func (dmm *DistributedModelManager) GetReplicationSummary() *ReplicationSummary {
	if dmm.replicationManager == nil {
		return &ReplicationSummary{QueueLength: 0, WorkerCount: 0, Models: map[string]int{}}
	}
	return dmm.replicationManager.GetSummary()
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
	peerIDs := md.manager.p2p.GetConnectedPeers()
	for _, peerID := range peerIDs {
		go md.sendBroadcastToPeer(peerID, broadcast)
	}

	// Update broadcast metrics
	md.updateBroadcastMetrics(len(peerIDs), len(models))
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
func (dmm *DistributedModelManager) syncWithPeer(peerIDStr string, localModels map[string]*DistributedModel) {
	// Prepare sync message
	syncMessage := map[string]interface{}{
		"type":      "registry_sync",
		"peer_id":   dmm.p2p.ID().String(),
		"timestamp": time.Now().Unix(),
		"models":    localModels,
	}

	// Send via P2P (simplified implementation)
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

// GetShardPlan retrieves the shard plan for a model
func (dmm *DistributedModelManager) GetShardPlan(modelName string) (*ShardPlan, error) {
	if dmm.shardManager == nil {
		return nil, fmt.Errorf("shard manager not initialized")
	}
	return dmm.shardManager.GetShardPlan(modelName)
}

// LoadShardOnDemand loads a specific shard on demand
func (dmm *DistributedModelManager) LoadShardOnDemand(modelName string, shardIndex int) error {
	if dmm.modelLoader == nil {
		return fmt.Errorf("model loader not initialized")
	}
	return dmm.modelLoader.LoadShardOnDemand(modelName, shardIndex)
}

// broadcastShardAvailability broadcasts shard availability to the network
func (dmm *DistributedModelManager) broadcastShardAvailability(shardID, nodeID string) {
	if dmm.p2p == nil || dmm.shardProtocolHandler == nil {
		return
	}

	// Get the model name for this shard from registry
	modelName := ""
	locations, err := dmm.shardRegistry.LocateShard(shardID)
	if err == nil && len(locations) > 0 {
		// Try to extract model name from shard metadata
		// For now, use a placeholder - in a full implementation,
		// the shard registry would store model associations
		modelName = "unknown"
	}

	// Create shard status message for the protocol
	status := protocols.ShardStatusMessage{
		ShardID:      shardID,
		ModelName:    modelName,
		NodeID:       nodeID,
		IsAvailable:  true,
		IsLoaded:     true,
		StoragePath:  "", // Will be filled by the receiving node
		Size:         0,  // Will be filled by the receiving node
		Checksum:     "", // Will be filled by the receiving node
		LastAccessed: time.Now(),
	}

	// Update shard status in registry first
	err = dmm.shardRegistry.UpdateShardStatus(shardID, status)
	if err != nil {
		dmm.logger.Warn("failed to update shard status in registry",
			"shard", shardID,
			"error", err)
	}

	// Announce the shard via the protocol handler
	err = dmm.shardProtocolHandler.AnnounceShards(modelName, []string{shardID})
	if err != nil {
		dmm.logger.Error("failed to announce shard availability",
			"shard", shardID,
			"model", modelName,
			"error", err)
	} else {
		dmm.logger.Info("successfully announced shard availability",
			"shard", shardID,
			"node", nodeID,
			"model", modelName)
	}
}

// announceAllLocalShards announces all locally available shards to the network
func (dmm *DistributedModelManager) announceAllLocalShards() {
	if dmm.shardProtocolHandler == nil || dmm.shardRegistry == nil {
		return
	}

	// Get all local shards from the registry
	localShards, err := dmm.shardRegistry.GetLocalShards()
	if err != nil {
		dmm.logger.Warn("failed to get local shards for announcement", "error", err)
		return
	}

	if len(localShards) == 0 {
		dmm.logger.Debug("no local shards to announce")
		return
	}

	// Group shards by model for efficient announcement
	shardsByModel := make(map[string][]string)
	for _, shardID := range localShards {
		// Get model name for this shard
		locations, err := dmm.shardRegistry.LocateShard(shardID)
		if err != nil {
			continue
		}

		// Find the local location to get model info
		modelName := "unknown"
		for _, loc := range locations {
			if loc.IsLocal {
				// In a full implementation, model name would be stored with location
				// For now, we'll use a placeholder
				modelName = "unknown"
				break
			}
		}

		if _, exists := shardsByModel[modelName]; !exists {
			shardsByModel[modelName] = make([]string, 0)
		}
		shardsByModel[modelName] = append(shardsByModel[modelName], shardID)
	}

	// Announce each model's shards
	for modelName, shardIDs := range shardsByModel {
		err := dmm.shardProtocolHandler.AnnounceShards(modelName, shardIDs)
		if err != nil {
			dmm.logger.Error("failed to announce model shards",
				"model", modelName,
				"shard_count", len(shardIDs),
				"error", err)
		} else {
			dmm.logger.Info("announced model shards",
				"model", modelName,
				"shard_count", len(shardIDs))
		}
	}
}

// syncShardRegistry synchronizes the shard registry with the network
func (dmm *DistributedModelManager) syncShardRegistry() {
	// Announce all local shards to keep the network updated
	dmm.announceAllLocalShards()

	// The registry synchronization is handled by the protocol handler
	// which automatically processes incoming announcements and updates
	// the registry with remote shard information

	dmm.logger.Debug("shard registry synchronization completed")
}
