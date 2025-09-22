package models

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/protocols"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ShardLocation represents where a shard is stored
type ShardLocation struct {
	ShardID      string    `json:"shard_id"`
	NodeID       string    `json:"node_id"`
	StoragePath  string    `json:"storage_path"`
	IsLocal      bool      `json:"is_local"`
	IsAvailable  bool      `json:"is_available"`
	LastVerified time.Time `json:"last_verified"`
	AccessCount  int64     `json:"access_count"`
	Latency      time.Duration `json:"latency"`
}

// ShardIndex maintains an index of shard locations
type ShardIndex struct {
	ShardID       string           `json:"shard_id"`
	ModelID       string           `json:"model_id"`
	PrimaryNode   string           `json:"primary_node"`
	ReplicaNodes  []string         `json:"replica_nodes"`
	Locations     []*ShardLocation `json:"locations"`
	Checksum      string           `json:"checksum"`
	Size          int64            `json:"size"`
	CreatedAt     time.Time        `json:"created_at"`
	LastAccessed  time.Time        `json:"last_accessed"`
	AccessPattern string           `json:"access_pattern"` // sequential, random, hot
	Health        ShardHealth      `json:"health"`
}

// ShardHealth represents the health status of a shard
type ShardHealth struct {
	Status           ShardHealthStatus `json:"status"`
	LastCheck        time.Time         `json:"last_check"`
	FailureCount     int               `json:"failure_count"`
	CorruptionDetected bool            `json:"corruption_detected"`
	RepairAttempts   int               `json:"repair_attempts"`
}

// ShardHealthStatus represents shard health states
type ShardHealthStatus string

const (
	ShardHealthHealthy    ShardHealthStatus = "healthy"
	ShardHealthDegraded   ShardHealthStatus = "degraded"
	ShardHealthUnavailable ShardHealthStatus = "unavailable"
	ShardHealthCorrupted  ShardHealthStatus = "corrupted"
)

// ShardReplicationManager manages shard replication
type ShardReplicationManager struct {
	mu                sync.RWMutex
	registry          *ShardRegistry
	replicationFactor int
	replicationQueue  chan *ShardReplicationTask
	workers           []*ShardReplicationWorker
	config            *ReplicationConfig
}

// ShardReplicationTask represents a shard replication task
type ShardReplicationTask struct {
	ShardID     string
	SourceNode  string
	TargetNodes []string
	Priority    int
	Retries     int
	CreatedAt   time.Time
}

// ShardReplicationWorker processes shard replication tasks
type ShardReplicationWorker struct {
	ID       int
	manager  *ShardReplicationManager
	stopChan chan struct{}
}

// ReplicationConfig contains replication configuration
type ReplicationConfig struct {
	MinReplicas       int
	MaxReplicas       int
	ReplicationFactor int
	CheckInterval     time.Duration
	RepairTimeout     time.Duration
	MaxRetries        int
}

// ShardDiscovery handles shard discovery across the network
type ShardDiscovery struct {
	registry         *ShardRegistry
	p2pNode          *p2p.Node
	discoveryCache   map[string]*DiscoveryResult
	cacheMutex       sync.RWMutex
	discoveryTimeout time.Duration
	announceInterval time.Duration
}

// DiscoveryResult represents a shard discovery result
type DiscoveryResult struct {
	ShardID   string
	Locations []*ShardLocation
	Timestamp time.Time
	TTL       time.Duration
}

// ShardRegistry maintains a distributed registry of model shards
type ShardRegistry struct {
	mu              sync.RWMutex
	shards          map[string]*ShardIndex        // shardID -> index
	modelShards     map[string][]string           // modelID -> shardIDs
	nodeShards      map[string][]string           // nodeID -> shardIDs
	replicationMgr  *ShardReplicationManager
	discovery       *ShardDiscovery
	p2pNode         *p2p.Node
	logger          *slog.Logger

	// Conflict resolution
	conflictResolver *ConflictResolver

	// Cleanup management
	cleanupInterval  time.Duration
	lastCleanup      time.Time
}

// ConflictResolver handles shard state conflicts
type ConflictResolver struct {
	mu               sync.RWMutex
	resolutionLog    []*ConflictResolution
	strategies       map[string]ResolutionStrategy
}

// ConflictResolution records a conflict resolution
type ConflictResolution struct {
	ShardID      string
	ConflictType string
	Resolution   string
	Timestamp    time.Time
	Details      map[string]interface{}
}

// ResolutionStrategy defines how to resolve conflicts
type ResolutionStrategy func(conflicts []ShardIndex) (*ShardIndex, error)

// NewShardRegistry creates a new shard registry
func NewShardRegistry(p2pNode *p2p.Node, logger *slog.Logger) *ShardRegistry {
	registry := &ShardRegistry{
		shards:          make(map[string]*ShardIndex),
		modelShards:     make(map[string][]string),
		nodeShards:      make(map[string][]string),
		p2pNode:         p2pNode,
		logger:          logger,
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}

	// Initialize replication manager
	replicationConfig := &ReplicationConfig{
		MinReplicas:       2,
		MaxReplicas:       5,
		ReplicationFactor: 3,
		CheckInterval:     30 * time.Second,
		RepairTimeout:     5 * time.Minute,
		MaxRetries:        3,
	}
	registry.replicationMgr = NewShardReplicationManager(registry, replicationConfig)

	// Initialize discovery service
	registry.discovery = &ShardDiscovery{
		registry:         registry,
		p2pNode:          p2pNode,
		discoveryCache:   make(map[string]*DiscoveryResult),
		discoveryTimeout: 10 * time.Second,
		announceInterval: 30 * time.Second,
	}

	// Initialize conflict resolver
	registry.conflictResolver = NewConflictResolver()

	// Start background services
	go registry.maintenanceLoop()
	go registry.discovery.start()
	go registry.replicationMgr.start()

	return registry
}

// RegisterModelShard registers a new shard in the registry from a ModelShard
func (r *ShardRegistry) RegisterModelShard(shard *ModelShard) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create shard index
	index := &ShardIndex{
		ShardID:      shard.ID,
		ModelID:      shard.ModelID,
		PrimaryNode:  r.p2pNode.ID().String(),
		ReplicaNodes: []string{},
		Locations:    make([]*ShardLocation, 0),
		Checksum:     shard.Checksum,
		Size:         shard.Size,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Health: ShardHealth{
			Status:    ShardHealthHealthy,
			LastCheck: time.Now(),
		},
	}

	// Add primary location
	primaryLocation := &ShardLocation{
		ShardID:      shard.ID,
		NodeID:       r.p2pNode.ID().String(),
		StoragePath:  fmt.Sprintf("/shards/%s/%s", shard.ModelID, shard.ID),
		IsLocal:      true,
		IsAvailable:  true,
		LastVerified: time.Now(),
	}
	index.Locations = append(index.Locations, primaryLocation)

	// Add to registry
	r.shards[shard.ID] = index

	// Update model shards mapping
	r.modelShards[shard.ModelID] = append(r.modelShards[shard.ModelID], shard.ID)

	// Update node shards mapping
	r.nodeShards[r.p2pNode.ID().String()] = append(r.nodeShards[r.p2pNode.ID().String()], shard.ID)

	// Trigger replication if needed
	if len(shard.NodeAssignments) > 1 {
		task := &ShardReplicationTask{
			ShardID:     shard.ID,
			SourceNode:  r.p2pNode.ID().String(),
			TargetNodes: shard.NodeAssignments[1:], // Replicate to other assigned nodes
			Priority:    shard.Priority,
			CreatedAt:   time.Now(),
		}
		r.replicationMgr.queueReplication(task)
	}

	r.logger.Info("shard registered", "shard", shard.ID, "model", shard.ModelID)
	return nil
}

// convertToProtocolLocations converts ShardLocation to protocols.ShardNodeLocation
func convertToProtocolLocations(locations []*ShardLocation) []protocols.ShardNodeLocation {
	result := make([]protocols.ShardNodeLocation, len(locations))
	for i, loc := range locations {
		// Convert peer ID from string to peer.ID
		peerID, _ := peer.Decode(loc.NodeID)
		result[i] = protocols.ShardNodeLocation{
			NodeID:      loc.NodeID,
			PeerID:      peerID,
			IsAvailable: loc.IsAvailable,
			IsLoaded:    true, // Assume loaded if available
			IsLocal:     loc.IsLocal,
			StoragePath: loc.StoragePath,
			LastSeen:    loc.LastVerified,
		}
	}
	return result
}

// RegisterShard implements the protocols.ShardRegistry interface
func (r *ShardRegistry) RegisterShard(shardID, modelName string, location protocols.ShardNodeLocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if shard already exists
	index, exists := r.shards[shardID]
	if !exists {
		// Create new shard index
		index = &ShardIndex{
			ShardID:      shardID,
			ModelID:      modelName,
			PrimaryNode:  location.NodeID,
			ReplicaNodes: []string{},
			Locations:    make([]*ShardLocation, 0),
			CreatedAt:    time.Now(),
			LastAccessed: time.Now(),
			Health: ShardHealth{
				Status:    ShardHealthHealthy,
				LastCheck: time.Now(),
			},
		}
		r.shards[shardID] = index

		// Update model shards mapping
		r.modelShards[modelName] = append(r.modelShards[modelName], shardID)
	}

	// Convert protocol location to internal location
	shardLocation := &ShardLocation{
		ShardID:      shardID,
		NodeID:       location.NodeID,
		StoragePath:  location.StoragePath,
		IsLocal:      location.IsLocal,
		IsAvailable:  location.IsAvailable,
		LastVerified: location.LastSeen,
	}

	// Add location if not already present
	locationExists := false
	for _, loc := range index.Locations {
		if loc.NodeID == location.NodeID {
			// Update existing location
			loc.IsAvailable = location.IsAvailable
			loc.IsLocal = location.IsLocal
			loc.StoragePath = location.StoragePath
			loc.LastVerified = location.LastSeen
			locationExists = true
			break
		}
	}

	if !locationExists {
		index.Locations = append(index.Locations, shardLocation)

		// Update node shards mapping
		r.nodeShards[location.NodeID] = append(r.nodeShards[location.NodeID], shardID)
	}

	r.logger.Info("shard registered via protocol", "shard", shardID, "model", modelName, "node", location.NodeID)
	return nil
}

// LocateShard finds the locations of a shard (implements protocols.ShardRegistry interface)
func (r *ShardRegistry) LocateShard(shardID string) ([]protocols.ShardNodeLocation, error) {
	r.mu.RLock()
	index, exists := r.shards[shardID]
	r.mu.RUnlock()

	if exists && len(index.Locations) > 0 {
		// Update access metadata
		r.mu.Lock()
		index.LastAccessed = time.Now()
		r.mu.Unlock()

		return convertToProtocolLocations(index.Locations), nil
	}

	// Try discovery if not found locally
	locations, err := r.discovery.discoverShard(shardID)
	if err != nil {
		return nil, fmt.Errorf("shard %s not found: %w", shardID, err)
	}

	// Cache discovered locations
	r.mu.Lock()
	if index == nil {
		index = &ShardIndex{
			ShardID:      shardID,
			Locations:    locations,
			LastAccessed: time.Now(),
			Health: ShardHealth{
				Status:    ShardHealthHealthy,
				LastCheck: time.Now(),
			},
		}
		r.shards[shardID] = index
	} else {
		index.Locations = locations
	}
	r.mu.Unlock()

	return convertToProtocolLocations(locations), nil
}

// UpdateShardStatusByNode updates the status of a shard by node ID
func (r *ShardRegistry) UpdateShardStatusByNode(nodeID, shardID string, available bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	index, exists := r.shards[shardID]
	if !exists {
		return fmt.Errorf("shard %s not found", shardID)
	}

	// Update location status
	for _, loc := range index.Locations {
		if loc.NodeID == nodeID {
			loc.IsAvailable = available
			loc.LastVerified = time.Now()
			break
		}
	}

	// Update health status
	if !available {
		index.Health.FailureCount++
		if index.Health.FailureCount > 3 {
			index.Health.Status = ShardHealthDegraded
		}
		if index.Health.FailureCount > 10 {
			index.Health.Status = ShardHealthUnavailable
		}
	} else {
		index.Health.FailureCount = 0
		index.Health.Status = ShardHealthHealthy
	}

	index.Health.LastCheck = time.Now()

	return nil
}

// UpdateShardStatus implements the protocols.ShardRegistry interface
func (r *ShardRegistry) UpdateShardStatus(shardID string, status protocols.ShardStatusMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	index, exists := r.shards[shardID]
	if !exists {
		return fmt.Errorf("shard %s not found", shardID)
	}

	// Update location status
	for _, loc := range index.Locations {
		if loc.NodeID == status.NodeID {
			loc.IsAvailable = status.IsAvailable
			loc.StoragePath = status.StoragePath
			loc.LastVerified = status.LastAccessed
			break
		}
	}

	// Update health status
	if !status.IsAvailable {
		index.Health.FailureCount++
		if index.Health.FailureCount > 3 {
			index.Health.Status = ShardHealthDegraded
		}
		if index.Health.FailureCount > 10 {
			index.Health.Status = ShardHealthUnavailable
		}
	} else {
		index.Health.FailureCount = 0
		index.Health.Status = ShardHealthHealthy
	}

	index.Health.LastCheck = time.Now()

	return nil
}

// GetShardsByModel returns all shards for a model
func (r *ShardRegistry) GetShardsByModel(modelID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shardIDs, exists := r.modelShards[modelID]
	if !exists {
		return nil, fmt.Errorf("no shards found for model %s", modelID)
	}

	return shardIDs, nil
}

// GetShardsByNode returns all shards on a node
func (r *ShardRegistry) GetShardsByNode(nodeID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.nodeShards[nodeID]
}

// GetLocalShards returns all shards stored on the local node
func (r *ShardRegistry) GetLocalShards() ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodeID := r.p2pNode.ID().String()
	shards := r.nodeShards[nodeID]
	if shards == nil {
		return []string{}, nil
	}

	// Return a copy to avoid race conditions
	result := make([]string, len(shards))
	copy(result, shards)
	return result, nil
}

// maintenanceLoop performs periodic maintenance tasks
func (r *ShardRegistry) maintenanceLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.performMaintenance()
	}
}

// performMaintenance executes maintenance tasks
func (r *ShardRegistry) performMaintenance() {
	now := time.Now()

	// Cleanup orphaned shards
	if now.Sub(r.lastCleanup) > r.cleanupInterval {
		r.cleanupOrphanedShards()
		r.lastCleanup = now
	}

	// Verify shard health
	r.verifyShardHealth()

	// Check replication status
	r.checkReplicationStatus()
}

// cleanupOrphanedShards removes shards with no valid locations
func (r *ShardRegistry) cleanupOrphanedShards() {
	r.mu.Lock()
	defer r.mu.Unlock()

	orphaned := []string{}
	for shardID, index := range r.shards {
		hasValidLocation := false
		for _, loc := range index.Locations {
			if loc.IsAvailable {
				hasValidLocation = true
				break
			}
		}

		if !hasValidLocation {
			orphaned = append(orphaned, shardID)
		}
	}

	// Remove orphaned shards
	for _, shardID := range orphaned {
		delete(r.shards, shardID)
		r.logger.Info("removed orphaned shard", "shard", shardID)
	}
}

// verifyShardHealth checks the health of all shards
func (r *ShardRegistry) verifyShardHealth() {
	r.mu.RLock()
	shards := make([]*ShardIndex, 0, len(r.shards))
	for _, shard := range r.shards {
		shards = append(shards, shard)
	}
	r.mu.RUnlock()

	for _, shard := range shards {
		// Check if shard needs verification
		if time.Since(shard.Health.LastCheck) > 5*time.Minute {
			go r.verifyShard(shard)
		}
	}
}

// verifyShard verifies a single shard's health
func (r *ShardRegistry) verifyShard(shard *ShardIndex) {
	// Verify each location
	for _, loc := range shard.Locations {
		if loc.IsLocal {
			// Verify local shard
			// In real implementation, would check file integrity
			loc.IsAvailable = true
			loc.LastVerified = time.Now()
		} else {
			// Verify remote shard via P2P
			// In real implementation, would ping the remote node
			loc.LastVerified = time.Now()
		}
	}

	// Update health status
	r.mu.Lock()
	shard.Health.LastCheck = time.Now()
	r.mu.Unlock()
}

// checkReplicationStatus ensures adequate replication
func (r *ShardRegistry) checkReplicationStatus() {
	r.mu.RLock()
	underReplicated := []string{}

	for shardID, index := range r.shards {
		activeReplicas := 0
		for _, loc := range index.Locations {
			if loc.IsAvailable {
				activeReplicas++
			}
		}

		if activeReplicas < r.replicationMgr.config.MinReplicas {
			underReplicated = append(underReplicated, shardID)
		}
	}
	r.mu.RUnlock()

	// Queue replication tasks for under-replicated shards
	for _, shardID := range underReplicated {
		r.replicationMgr.repairShard(shardID)
	}
}

// NewShardReplicationManager creates a new replication manager
func NewShardReplicationManager(registry *ShardRegistry, config *ReplicationConfig) *ShardReplicationManager {
	mgr := &ShardReplicationManager{
		registry:         registry,
		replicationFactor: config.ReplicationFactor,
		replicationQueue: make(chan *ReplicationTask, 100),
		config:           config,
	}

	// Start replication workers
	mgr.workers = make([]*ShardReplicationWorker, 3)
	for i := 0; i < 3; i++ {
		mgr.workers[i] = &ShardReplicationWorker{
			ID:       i,
			manager:  mgr,
			stopChan: make(chan struct{}),
		}
	}

	return mgr
}

// start starts the replication manager
func (m *ShardReplicationManager) start() {
	// Start workers
	for _, worker := range m.workers {
		go worker.start()
	}

	// Start replication monitor
	go m.monitorReplication()
}

// queueReplication adds a replication task to the queue
func (m *ShardReplicationManager) queueReplication(task *ShardReplicationTask) {
	select {
	case m.replicationQueue <- task:
	default:
		// Queue full, log warning
		m.registry.logger.Warn("replication queue full", "shard", task.ShardID)
	}
}

// repairShard creates replicas for an under-replicated shard
func (m *ShardReplicationManager) repairShard(shardID string) {
	m.mu.RLock()
	index := m.registry.shards[shardID]
	m.mu.RUnlock()

	if index == nil {
		return
	}

	// Find available source node
	var sourceNode string
	for _, loc := range index.Locations {
		if loc.IsAvailable {
			sourceNode = loc.NodeID
			break
		}
	}

	if sourceNode == "" {
		m.registry.logger.Error("no available source for shard repair", "shard", shardID)
		return
	}

	// Find target nodes for replication
	targetNodes := m.selectReplicationTargets(shardID, m.config.MinReplicas)

	if len(targetNodes) > 0 {
		task := &ShardReplicationTask{
			ShardID:     shardID,
			SourceNode:  sourceNode,
			TargetNodes: targetNodes,
			Priority:    10, // High priority for repair
			CreatedAt:   time.Now(),
		}
		m.queueReplication(task)
	}
}

// selectReplicationTargets selects nodes for shard replication
func (m *ShardReplicationManager) selectReplicationTargets(shardID string, count int) []string {
	// In real implementation, would select based on:
	// - Node capacity
	// - Network topology
	// - Current load
	// - Geographic distribution

	targets := []string{}
	// Simplified: just return some peer IDs
	peers := m.registry.p2pNode.GetConnectedPeers()
	for i, peer := range peers {
		if i >= count {
			break
		}
		targets = append(targets, peer.String())
	}

	return targets
}

// monitorReplication monitors replication health
func (m *ShardReplicationManager) monitorReplication() {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.checkReplicationHealth()
	}
}

// checkReplicationHealth verifies replication status
func (m *ShardReplicationManager) checkReplicationHealth() {
	// Check all shards for proper replication
	m.registry.checkReplicationStatus()
}

// ShardReplicationWorker methods

// start starts the shard replication worker
func (w *ShardReplicationWorker) start() {
	for {
		select {
		case <-w.stopChan:
			return
		case task := <-w.manager.replicationQueue:
			w.processShardTask(task)
		}
	}
}

// processShardTask processes a shard replication task
func (w *ShardReplicationWorker) processShardTask(task *ShardReplicationTask) {
	// In real implementation, would:
	// 1. Connect to source node
	// 2. Transfer shard data to target nodes
	// 3. Verify integrity
	// 4. Update registry

	w.manager.registry.logger.Info("processing shard replication task",
		"worker", w.ID,
		"shard", task.ShardID,
		"targets", len(task.TargetNodes))

	// Simulate replication
	time.Sleep(2 * time.Second)

	// Update registry with new locations
	w.manager.mu.Lock()
	if index, exists := w.manager.registry.shards[task.ShardID]; exists {
		for _, nodeID := range task.TargetNodes {
			newLocation := &ShardLocation{
				ShardID:      task.ShardID,
				NodeID:       nodeID,
				StoragePath:  fmt.Sprintf("/shards/%s/%s", index.ModelID, task.ShardID),
				IsLocal:      false,
				IsAvailable:  true,
				LastVerified: time.Now(),
			}
			index.Locations = append(index.Locations, newLocation)
			index.ReplicaNodes = append(index.ReplicaNodes, nodeID)
		}
	}
	w.manager.mu.Unlock()
}

// ShardDiscovery methods

// start starts the discovery service
func (d *ShardDiscovery) start() {
	// Start announcement routine
	go d.announceShards()

	// Start discovery listener
	go d.listenForAnnouncements()
}

// announceShards periodically announces local shards
func (d *ShardDiscovery) announceShards() {
	ticker := time.NewTicker(d.announceInterval)
	defer ticker.Stop()

	for range ticker.C {
		d.broadcastLocalShards()
	}
}

// broadcastLocalShards broadcasts information about local shards
func (d *ShardDiscovery) broadcastLocalShards() {
	d.registry.mu.RLock()
	localShards := d.registry.nodeShards[d.p2pNode.ID().String()]
	d.registry.mu.RUnlock()

	if len(localShards) == 0 {
		return
	}

	// Prepare announcement
	announcement := map[string]interface{}{
		"type":      "shard_announcement",
		"node_id":   d.p2pNode.ID().String(),
		"shards":    localShards,
		"timestamp": time.Now().Unix(),
	}

	// Broadcast to peers
	peers := d.p2pNode.GetConnectedPeers()
	for _, peer := range peers {
		go d.sendAnnouncementToPeer(peer.String(), announcement)
	}
}

// sendAnnouncementToPeer sends shard announcement to a peer
func (d *ShardDiscovery) sendAnnouncementToPeer(peerID string, announcement map[string]interface{}) {
	// In real implementation, would use P2P messaging
	d.registry.logger.Debug("announcing shards to peer", "peer", peerID, "shards", announcement["shards"])
}

// listenForAnnouncements listens for shard announcements from peers
func (d *ShardDiscovery) listenForAnnouncements() {
	// In real implementation, would listen on P2P network
	// For now, this is a placeholder
}

// discoverShard discovers a shard on the network
func (d *ShardDiscovery) discoverShard(shardID string) ([]*ShardLocation, error) {
	// Check cache first
	d.cacheMutex.RLock()
	if result, exists := d.discoveryCache[shardID]; exists {
		if time.Since(result.Timestamp) < result.TTL {
			d.cacheMutex.RUnlock()
			return result.Locations, nil
		}
	}
	d.cacheMutex.RUnlock()

	// Query peers for shard
	locations := []*ShardLocation{}
	peers := d.p2pNode.GetConnectedPeers()

	for _, peer := range peers {
		// In real implementation, would query peer via P2P
		// For now, simulate finding shard on some peers
		if len(locations) < 2 { // Simulate finding on first 2 peers
			location := &ShardLocation{
				ShardID:      shardID,
				NodeID:       peer.String(),
				IsLocal:      false,
				IsAvailable:  true,
				LastVerified: time.Now(),
			}
			locations = append(locations, location)
		}
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("shard not found on network")
	}

	// Cache discovery result
	d.cacheMutex.Lock()
	d.discoveryCache[shardID] = &DiscoveryResult{
		ShardID:   shardID,
		Locations: locations,
		Timestamp: time.Now(),
		TTL:       5 * time.Minute,
	}
	d.cacheMutex.Unlock()

	return locations, nil
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver() *ConflictResolver {
	resolver := &ConflictResolver{
		resolutionLog: make([]*ConflictResolution, 0),
		strategies:    make(map[string]ResolutionStrategy),
	}

	// Register default strategies
	resolver.strategies["timestamp"] = timestampResolutionStrategy
	resolver.strategies["majority"] = majorityResolutionStrategy
	resolver.strategies["checksum"] = checksumResolutionStrategy

	return resolver
}

// Default resolution strategies

func timestampResolutionStrategy(conflicts []ShardIndex) (*ShardIndex, error) {
	// Select the most recently updated shard
	var newest *ShardIndex
	for i := range conflicts {
		if newest == nil || conflicts[i].LastAccessed.After(newest.LastAccessed) {
			newest = &conflicts[i]
		}
	}
	return newest, nil
}

func majorityResolutionStrategy(conflicts []ShardIndex) (*ShardIndex, error) {
	// Select the shard with most replicas
	var best *ShardIndex
	maxReplicas := 0
	for i := range conflicts {
		if len(conflicts[i].ReplicaNodes) > maxReplicas {
			best = &conflicts[i]
			maxReplicas = len(conflicts[i].ReplicaNodes)
		}
	}
	return best, nil
}

func checksumResolutionStrategy(conflicts []ShardIndex) (*ShardIndex, error) {
	// Group by checksum and select majority
	checksumCount := make(map[string]int)
	checksumToShard := make(map[string]*ShardIndex)

	for i := range conflicts {
		checksum := conflicts[i].Checksum
		checksumCount[checksum]++
		checksumToShard[checksum] = &conflicts[i]
	}

	// Find most common checksum
	maxCount := 0
	var bestChecksum string
	for checksum, count := range checksumCount {
		if count > maxCount {
			maxCount = count
			bestChecksum = checksum
		}
	}

	return checksumToShard[bestChecksum], nil
}

// GetShardHealth returns the health status of a shard
func (r *ShardRegistry) GetShardHealth(shardID string) (*ShardHealth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	index, exists := r.shards[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found", shardID)
	}

	return &index.Health, nil
}

// GetRegistryStats returns statistics about the registry
func (r *ShardRegistry) GetRegistryStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalShards := len(r.shards)
	totalModels := len(r.modelShards)
	totalNodes := len(r.nodeShards)

	healthyShards := 0
	degradedShards := 0
	unavailableShards := 0

	for _, index := range r.shards {
		switch index.Health.Status {
		case ShardHealthHealthy:
			healthyShards++
		case ShardHealthDegraded:
			degradedShards++
		case ShardHealthUnavailable:
			unavailableShards++
		}
	}

	return map[string]interface{}{
		"total_shards":       totalShards,
		"total_models":       totalModels,
		"total_nodes":        totalNodes,
		"healthy_shards":     healthyShards,
		"degraded_shards":    degradedShards,
		"unavailable_shards": unavailableShards,
		"replication_factor": r.replicationMgr.replicationFactor,
	}
}

