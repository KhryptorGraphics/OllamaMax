package distributed

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// Start starts the cluster manager
func (cm *ClusterManager) Start(ctx context.Context) error {
	// Start node discovery
	go cm.discovery.start(ctx)
	
	// Start health checker
	go cm.healthChecker.start(ctx)
	
	// Start heartbeat processor
	go cm.processHeartbeats(ctx)
	
	// Register local node
	if err := cm.registerLocalNode(); err != nil {
		return fmt.Errorf("failed to register local node: %v", err)
	}
	
	slog.Info("cluster manager started")
	return nil
}

// registerLocalNode registers the local node with the cluster
func (cm *ClusterManager) registerLocalNode() error {
	localNode := &NodeInfo{
		ID:           cm.scheduler.config.NodeID,
		Address:      cm.scheduler.p2pNode.GetAddress(),
		Status:       NodeStatusOnline,
		Capacity:     cm.getLocalCapacity(),
		Usage:        cm.getLocalUsage(),
		Models:       cm.getLocalModels(),
		GPUs:         cm.getLocalGPUs(),
		LastSeen:     time.Now(),
		Latency:      0,
		Bandwidth:    cm.getLocalBandwidth(),
		Capabilities: cm.getLocalCapabilities(),
		Metadata:     make(map[string]interface{}),
	}
	
	cm.nodesMu.Lock()
	cm.nodes[localNode.ID] = localNode
	cm.nodesMu.Unlock()
	
	// Announce to cluster
	announcement := &NodeAnnouncement{
		Node:      localNode,
		Action:    "join",
		Timestamp: time.Now(),
	}
	
	cm.discovery.broadcast <- announcement
	
	return nil
}

// getLocalCapacity returns the capacity of the local node
func (cm *ClusterManager) getLocalCapacity() *ResourceCapacity {
	// Get local system information
	// This would typically use system APIs to get actual capacity
	return &ResourceCapacity{
		CPUCores:         int64(4), // Example values
		MemoryBytes:      int64(16 * 1024 * 1024 * 1024), // 16GB
		DiskBytes:        int64(1024 * 1024 * 1024 * 1024), // 1TB
		GPUCount:         1,
		GPUMemoryBytes:   int64(8 * 1024 * 1024 * 1024), // 8GB
		NetworkBandwidth: int64(1000 * 1000 * 1000), // 1Gbps
		ComputeScore:     1.0,
	}
}

// getLocalUsage returns the current usage of the local node
func (cm *ClusterManager) getLocalUsage() *ResourceUsage {
	// Get local system usage
	// This would typically use system APIs to get actual usage
	return &ResourceUsage{
		CPUUtilization:    0.3,
		MemoryUtilization: 0.5,
		DiskUtilization:   0.2,
		GPUUtilization:    0.0,
		NetworkUtilization: 0.1,
		ActiveRequests:    0,
		QueuedRequests:    0,
		LoadAverage:       0.5,
	}
}

// getLocalModels returns the models available on the local node
func (cm *ClusterManager) getLocalModels() []string {
	// Get models from local scheduler
	// This would interface with the local Ollama server
	return []string{} // Placeholder
}

// getLocalGPUs returns the GPUs available on the local node
func (cm *ClusterManager) getLocalGPUs() []interface{} {
	// Get GPU information from local system
	// This would use the discover package from Ollama
	return []interface{}{} // Placeholder
}

// getLocalBandwidth returns the network bandwidth of the local node
func (cm *ClusterManager) getLocalBandwidth() int64 {
	// Measure or estimate local network bandwidth
	return 1000 * 1000 * 1000 // 1Gbps placeholder
}

// getLocalCapabilities returns the capabilities of the local node
func (cm *ClusterManager) getLocalCapabilities() []string {
	return []string{"inference", "embedding", "classification"}
}

// processHeartbeats processes heartbeat messages
func (cm *ClusterManager) processHeartbeats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case heartbeat := <-cm.heartbeat:
			cm.handleHeartbeat(heartbeat)
		}
	}
}

// handleHeartbeat handles a heartbeat message
func (cm *ClusterManager) handleHeartbeat(heartbeat *HeartbeatMessage) {
	cm.nodesMu.Lock()
	defer cm.nodesMu.Unlock()
	
	if node, exists := cm.nodes[heartbeat.NodeID]; exists {
		// Update existing node
		node.Status = heartbeat.Status
		node.Capacity = heartbeat.Capacity
		node.Usage = heartbeat.Usage
		node.Models = heartbeat.Models
		node.LastSeen = heartbeat.Timestamp
		node.Metadata = heartbeat.Metadata
	} else {
		// Create new node from heartbeat
		cm.nodes[heartbeat.NodeID] = &NodeInfo{
			ID:           heartbeat.NodeID,
			Status:       heartbeat.Status,
			Capacity:     heartbeat.Capacity,
			Usage:        heartbeat.Usage,
			Models:       heartbeat.Models,
			LastSeen:     heartbeat.Timestamp,
			Metadata:     heartbeat.Metadata,
		}
	}
}

// GetAvailableNodes returns all available nodes in the cluster
func (cm *ClusterManager) GetAvailableNodes() []*NodeInfo {
	cm.nodesMu.RLock()
	defer cm.nodesMu.RUnlock()
	
	var available []*NodeInfo
	for _, node := range cm.nodes {
		if node.Status == NodeStatusOnline {
			available = append(available, node)
		}
	}
	
	return available
}

// GetAllNodes returns all nodes in the cluster
func (cm *ClusterManager) GetAllNodes() []*NodeInfo {
	cm.nodesMu.RLock()
	defer cm.nodesMu.RUnlock()
	
	nodes := make([]*NodeInfo, 0, len(cm.nodes))
	for _, node := range cm.nodes {
		nodes = append(nodes, node)
	}
	
	return nodes
}

// GetNode returns a specific node by ID
func (cm *ClusterManager) GetNode(nodeID string) (*NodeInfo, bool) {
	cm.nodesMu.RLock()
	defer cm.nodesMu.RUnlock()
	
	node, exists := cm.nodes[nodeID]
	return node, exists
}

// RegisterModel registers a model in the cluster
func (cm *ClusterManager) RegisterModel(name, path string, size int64, checksum string, nodeID string) error {
	cm.modelsMu.Lock()
	defer cm.modelsMu.Unlock()
	
	if model, exists := cm.models[name]; exists {
		// Update existing model
		if !containsString(model.Locations, nodeID) {
			model.Locations = append(model.Locations, nodeID)
		}
		model.AccessCount++
		model.LastAccessed = time.Now()
	} else {
		// Create new model
		cm.models[name] = &ModelInfo{
			Name:             name,
			Path:             path,
			Size:             size,
			Checksum:         checksum,
			Locations:        []string{nodeID},
			ReplicationFactor: 1,
			AccessCount:      1,
			LastAccessed:     time.Now(),
			Popularity:       0.0,
			Metadata:         make(map[string]string),
		}
	}
	
	return nil
}

// GetModel returns a specific model by name
func (cm *ClusterManager) GetModel(name string) (*ModelInfo, bool) {
	cm.modelsMu.RLock()
	defer cm.modelsMu.RUnlock()
	
	model, exists := cm.models[name]
	return model, exists
}

// GetAllModels returns all models in the cluster
func (cm *ClusterManager) GetAllModels() map[string]*ModelInfo {
	cm.modelsMu.RLock()
	defer cm.modelsMu.RUnlock()
	
	models := make(map[string]*ModelInfo)
	for k, v := range cm.models {
		models[k] = v
	}
	
	return models
}

// UpdateNodeStatus updates the status of a node
func (cm *ClusterManager) UpdateNodeStatus(nodeID string, status NodeStatus) error {
	cm.nodesMu.Lock()
	defer cm.nodesMu.Unlock()
	
	if node, exists := cm.nodes[nodeID]; exists {
		node.Status = status
		node.LastSeen = time.Now()
		return nil
	}
	
	return fmt.Errorf("node not found: %s", nodeID)
}

// SendHeartbeat sends a heartbeat to the cluster
func (cm *ClusterManager) SendHeartbeat() {
	localNode := cm.getLocalNode()
	if localNode == nil {
		return
	}
	
	heartbeat := &HeartbeatMessage{
		NodeID:    localNode.ID,
		Timestamp: time.Now(),
		Status:    localNode.Status,
		Capacity:  localNode.Capacity,
		Usage:     cm.getLocalUsage(), // Get current usage
		Models:    localNode.Models,
		Metadata:  localNode.Metadata,
	}
	
	// Send to all peers via P2P
	if err := cm.scheduler.p2pNode.Broadcast("heartbeat", heartbeat); err != nil {
		slog.Warn("failed to send heartbeat", "error", err)
	}
}

// getLocalNode returns the local node information
func (cm *ClusterManager) getLocalNode() *NodeInfo {
	cm.nodesMu.RLock()
	defer cm.nodesMu.RUnlock()
	
	return cm.nodes[cm.scheduler.config.NodeID]
}

// Shutdown gracefully shuts down the cluster manager
func (cm *ClusterManager) Shutdown(ctx context.Context) error {
	// Stop health checker
	close(cm.healthChecker.stopCh)
	
	// Send leave announcement
	localNode := cm.getLocalNode()
	if localNode != nil {
		announcement := &NodeAnnouncement{
			Node:      localNode,
			Action:    "leave",
			Timestamp: time.Now(),
		}
		
		select {
		case cm.discovery.broadcast <- announcement:
		case <-time.After(5 * time.Second):
			// Timeout sending leave announcement
		}
	}
	
	slog.Info("cluster manager shutdown")
	return nil
}

// containsString checks if a string slice contains a specific string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// NodeDiscovery methods

// start starts the node discovery process
func (nd *NodeDiscovery) start(ctx context.Context) {
	// Start broadcast handler
	go nd.handleBroadcasts(ctx)
	
	// Start periodic discovery
	go nd.periodicDiscovery(ctx)
}

// handleBroadcasts handles node announcements
func (nd *NodeDiscovery) handleBroadcasts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case announcement := <-nd.broadcast:
			nd.handleAnnouncement(announcement)
		}
	}
}

// handleAnnouncement handles a node announcement
func (nd *NodeDiscovery) handleAnnouncement(announcement *NodeAnnouncement) {
	nd.registeredMu.Lock()
	defer nd.registeredMu.Unlock()
	
	switch announcement.Action {
	case "join":
		nd.registered[announcement.Node.ID] = announcement.Node
		slog.Info("node joined cluster", "node_id", announcement.Node.ID)
		
	case "leave":
		delete(nd.registered, announcement.Node.ID)
		slog.Info("node left cluster", "node_id", announcement.Node.ID)
		
	case "update":
		nd.registered[announcement.Node.ID] = announcement.Node
		slog.Debug("node updated", "node_id", announcement.Node.ID)
	}
	
	// Update cluster manager
	nd.manager.nodesMu.Lock()
	if announcement.Action == "leave" {
		delete(nd.manager.nodes, announcement.Node.ID)
	} else {
		nd.manager.nodes[announcement.Node.ID] = announcement.Node
	}
	nd.manager.nodesMu.Unlock()
}

// periodicDiscovery performs periodic node discovery
func (nd *NodeDiscovery) periodicDiscovery(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nd.discoverNodes()
		}
	}
}

// discoverNodes discovers new nodes in the network
func (nd *NodeDiscovery) discoverNodes() {
	// Query P2P network for new peers
	peers := nd.manager.scheduler.p2pNode.GetAllPeers()
	
	for peerID, peerInfo := range peers {
		nodeID := peerID.String()
		
		nd.registeredMu.RLock()
		_, exists := nd.registered[nodeID]
		nd.registeredMu.RUnlock()
		
		if !exists {
			// New peer discovered
			newNode := &NodeInfo{
				ID:           nodeID,
				Address:      peerInfo.Addresses[0].String(),
				Status:       NodeStatusOnline,
				Capacity:     &ResourceCapacity{},
				Usage:        &ResourceUsage{},
				Models:       []string{},
				LastSeen:     time.Now(),
				Metadata:     make(map[string]interface{}),
			}
			
			announcement := &NodeAnnouncement{
				Node:      newNode,
				Action:    "join",
				Timestamp: time.Now(),
			}
			
			nd.handleAnnouncement(announcement)
		}
	}
}

// HealthChecker methods

// start starts the health checker
func (hc *HealthChecker) start(ctx context.Context) {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkAllNodes()
		}
	}
}

// checkAllNodes checks the health of all nodes
func (hc *HealthChecker) checkAllNodes() {
	nodes := hc.manager.GetAllNodes()
	
	for _, node := range nodes {
		go hc.checkNode(node)
	}
}

// checkNode checks the health of a specific node
func (hc *HealthChecker) checkNode(node *NodeInfo) {
	start := time.Now()
	
	// Perform health check (ping)
	err := hc.manager.scheduler.p2pNode.Ping(node.Address)
	latency := time.Since(start)
	
	hc.checksMu.Lock()
	defer hc.checksMu.Unlock()
	
	check, exists := hc.checks[node.ID]
	if !exists {
		check = &HealthCheck{
			NodeID:              node.ID,
			ConsecutiveFailures: 0,
		}
		hc.checks[node.ID] = check
	}
	
	check.LastCheck = time.Now()
	check.Latency = latency
	
	if err != nil {
		check.Status = "unhealthy"
		check.Error = err.Error()
		check.ConsecutiveFailures++
		
		// Mark node as offline if too many failures
		if check.ConsecutiveFailures >= 3 {
			hc.manager.UpdateNodeStatus(node.ID, NodeStatusOffline)
		}
	} else {
		check.Status = "healthy"
		check.Error = ""
		check.ConsecutiveFailures = 0
		
		// Mark node as online
		hc.manager.UpdateNodeStatus(node.ID, NodeStatusOnline)
	}
}

// GetHealthStatus returns the health status of all nodes
func (hc *HealthChecker) GetHealthStatus() map[string]*HealthCheck {
	hc.checksMu.RLock()
	defer hc.checksMu.RUnlock()
	
	status := make(map[string]*HealthCheck)
	for k, v := range hc.checks {
		status[k] = v
	}
	
	return status
}

// MetricsCollector methods

// GetMetrics returns the current performance metrics
func (mc *MetricsCollector) GetMetrics() *PerformanceMetrics {
	mc.metricsMu.RLock()
	defer mc.metricsMu.RUnlock()
	
	return mc.metrics
}

// UpdateMetrics updates the performance metrics
func (mc *MetricsCollector) UpdateMetrics(sample MetricSample) {
	mc.metricsMu.Lock()
	defer mc.metricsMu.Unlock()
	
	// Add sample
	mc.samplesMu.Lock()
	mc.samples = append(mc.samples, sample)
	
	// Keep only last 1000 samples
	if len(mc.samples) > 1000 {
		mc.samples = mc.samples[len(mc.samples)-1000:]
	}
	mc.samplesMu.Unlock()
	
	// Update metrics
	mc.metrics.TotalRequests++
	mc.metrics.AverageLatency = mc.calculateAverageLatency()
	mc.metrics.Throughput = mc.calculateThroughput()
	mc.metrics.ResourceUtilization = sample.CPUUsage
	mc.metrics.LastUpdated = time.Now()
}

// calculateAverageLatency calculates the average latency from samples
func (mc *MetricsCollector) calculateAverageLatency() time.Duration {
	mc.samplesMu.RLock()
	defer mc.samplesMu.RUnlock()
	
	if len(mc.samples) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, sample := range mc.samples {
		total += sample.Latency
	}
	
	return total / time.Duration(len(mc.samples))
}

// calculateThroughput calculates the throughput from samples
func (mc *MetricsCollector) calculateThroughput() float64 {
	mc.samplesMu.RLock()
	defer mc.samplesMu.RUnlock()
	
	if len(mc.samples) == 0 {
		return 0
	}
	
	var total float64
	for _, sample := range mc.samples {
		total += sample.Throughput
	}
	
	return total / float64(len(mc.samples))
}
