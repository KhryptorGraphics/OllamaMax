package models

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// NetworkTopology represents the network topology for model distribution
type NetworkTopology struct {
	nodes      map[string]*TopologyNode
	nodesMutex sync.RWMutex

	// Topology metadata
	Type         TopologyType `json:"type"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	NodeCount    int          `json:"node_count"`
	avgLatency   time.Duration
	avgBandwidth int64

	logger *slog.Logger
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

	// Status
	Status   NodeStatus `json:"status"`
	LastSeen time.Time  `json:"last_seen"`
	Health   float64    `json:"health"`
}

// TopologyConnection represents a connection between nodes
type TopologyConnection struct {
	TargetID  string        `json:"target_id"`
	Weight    float64       `json:"weight"`
	Latency   time.Duration `json:"latency"`
	Bandwidth int64         `json:"bandwidth"`
	Quality   float64       `json:"quality"`
	Status    string        `json:"status"`
}

// TopologyType represents the type of network topology
type TopologyType string

const (
	TopologyMesh   TopologyType = "mesh"
	TopologyRing   TopologyType = "ring"
	TopologyTree   TopologyType = "tree"
	TopologyStar   TopologyType = "star"
	TopologyHybrid TopologyType = "hybrid"
)

// NodeStatus represents the status of a topology node
type NodeStatus string

const (
	NodeStatusActive       NodeStatus = "active"
	NodeStatusInactive     NodeStatus = "inactive"
	NodeStatusConnecting   NodeStatus = "connecting"
	NodeStatusDisconnected NodeStatus = "disconnected"
	NodeStatusFailed       NodeStatus = "failed"
)

// DistributionStrategy defines how models are distributed across the network
type DistributionStrategy interface {
	GetName() string
	SelectTargetNodes(sourceNode string, availableNodes []*TopologyNode, criteria map[string]interface{}) ([]*TopologyNode, error)
	CalculateDistributionPlan(modelName string, topology *NetworkTopology) (*DistributionPlan, error)
	OptimizeDistribution(currentPlan *DistributionPlan, topology *NetworkTopology) (*DistributionPlan, error)
}

// DistributionPlan represents a plan for distributing a model
type DistributionPlan struct {
	ModelName     string                 `json:"model_name"`
	SourceNode    string                 `json:"source_node"`
	TargetNodes   []string               `json:"target_nodes"`
	Strategy      string                 `json:"strategy"`
	Priority      int                    `json:"priority"`
	EstimatedTime time.Duration          `json:"estimated_time"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
}

// GeographicDistributionStrategy implements geographic-aware distribution
type GeographicDistributionStrategy struct {
	regionPriority map[string]int
	logger         *slog.Logger
}

// LoadBalancedDistributionStrategy implements load-balanced distribution
type LoadBalancedDistributionStrategy struct {
	maxLoadThreshold float64
	logger           *slog.Logger
}

// ReplicationSummary provides a summary of replication status
type ReplicationSummary struct {
	QueueLength int            `json:"queue_length"`
	WorkerCount int            `json:"worker_count"`
	Models      map[string]int `json:"models"` // model name -> replica count
}

// NewNetworkTopology creates a new network topology
func NewNetworkTopology(topologyType TopologyType, logger *slog.Logger) *NetworkTopology {
	return &NetworkTopology{
		nodes:     make(map[string]*TopologyNode),
		Type:      topologyType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		NodeCount: 0,
		logger:    logger,
	}
}

// NewGeographicDistributionStrategy creates a new geographic distribution strategy
func NewGeographicDistributionStrategy(logger *slog.Logger) *GeographicDistributionStrategy {
	return &GeographicDistributionStrategy{
		regionPriority: make(map[string]int),
		logger:         logger,
	}
}

// NewLoadBalancedDistributionStrategy creates a new load-balanced distribution strategy
func NewLoadBalancedDistributionStrategy(maxLoadThreshold float64, logger *slog.Logger) *LoadBalancedDistributionStrategy {
	return &LoadBalancedDistributionStrategy{
		maxLoadThreshold: maxLoadThreshold,
		logger:           logger,
	}
}

// NetworkTopology methods

// AddNode adds a node to the topology
func (nt *NetworkTopology) AddNode(node *TopologyNode) {
	nt.nodesMutex.Lock()
	defer nt.nodesMutex.Unlock()

	node.LastSeen = time.Now()
	node.Status = NodeStatusActive
	nt.nodes[node.ID] = node
	nt.NodeCount = len(nt.nodes)
	nt.UpdatedAt = time.Now()

	nt.logger.Info("node added to topology", "node_id", node.ID, "address", node.Address)
}

// RemoveNode removes a node from the topology
func (nt *NetworkTopology) RemoveNode(nodeID string) {
	nt.nodesMutex.Lock()
	defer nt.nodesMutex.Unlock()

	if _, exists := nt.nodes[nodeID]; exists {
		delete(nt.nodes, nodeID)
		nt.NodeCount = len(nt.nodes)
		nt.UpdatedAt = time.Now()

		nt.logger.Info("node removed from topology", "node_id", nodeID)
	}
}

// GetNode retrieves a node from the topology
func (nt *NetworkTopology) GetNode(nodeID string) (*TopologyNode, bool) {
	nt.nodesMutex.RLock()
	defer nt.nodesMutex.RUnlock()

	node, exists := nt.nodes[nodeID]
	if exists {
		// Return a copy to avoid race conditions
		nodeCopy := *node
		return &nodeCopy, true
	}
	return nil, false
}

// GetAllNodes returns all nodes in the topology
func (nt *NetworkTopology) GetAllNodes() []*TopologyNode {
	nt.nodesMutex.RLock()
	defer nt.nodesMutex.RUnlock()

	nodes := make([]*TopologyNode, 0, len(nt.nodes))
	for _, node := range nt.nodes {
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// GetActiveNodes returns only active nodes
func (nt *NetworkTopology) GetActiveNodes() []*TopologyNode {
	allNodes := nt.GetAllNodes()
	var activeNodes []*TopologyNode

	for _, node := range allNodes {
		if node.Status == NodeStatusActive {
			activeNodes = append(activeNodes, node)
		}
	}

	return activeNodes
}

// UpdateNodeStatus updates the status of a node
func (nt *NetworkTopology) UpdateNodeStatus(nodeID string, status NodeStatus) {
	nt.nodesMutex.Lock()
	defer nt.nodesMutex.Unlock()

	if node, exists := nt.nodes[nodeID]; exists {
		node.Status = status
		node.LastSeen = time.Now()
		nt.UpdatedAt = time.Now()

		nt.logger.Debug("node status updated", "node_id", nodeID, "status", status)
	}
}

// AddConnection adds a connection between two nodes
func (nt *NetworkTopology) AddConnection(sourceID, targetID string, connection *TopologyConnection) {
	nt.nodesMutex.Lock()
	defer nt.nodesMutex.Unlock()

	if sourceNode, exists := nt.nodes[sourceID]; exists {
		connection.TargetID = targetID
		sourceNode.Connections = append(sourceNode.Connections, connection)
		nt.UpdatedAt = time.Now()

		nt.logger.Debug("connection added", "source", sourceID, "target", targetID)
	}
}

// CalculateAverageLatency calculates the average latency across all connections
func (nt *NetworkTopology) CalculateAverageLatency() time.Duration {
	nt.nodesMutex.RLock()
	defer nt.nodesMutex.RUnlock()

	var totalLatency time.Duration
	var connectionCount int

	for _, node := range nt.nodes {
		for _, connection := range node.Connections {
			totalLatency += connection.Latency
			connectionCount++
		}
	}

	if connectionCount == 0 {
		return 0
	}

	nt.avgLatency = totalLatency / time.Duration(connectionCount)
	return nt.avgLatency
}

// CalculateAverageBandwidth calculates the average bandwidth across all connections
func (nt *NetworkTopology) CalculateAverageBandwidth() int64 {
	nt.nodesMutex.RLock()
	defer nt.nodesMutex.RUnlock()

	var totalBandwidth int64
	var connectionCount int

	for _, node := range nt.nodes {
		for _, connection := range node.Connections {
			totalBandwidth += connection.Bandwidth
			connectionCount++
		}
	}

	if connectionCount == 0 {
		return 0
	}

	nt.avgBandwidth = totalBandwidth / int64(connectionCount)
	return nt.avgBandwidth
}

// GetTopologyStatistics returns statistics about the topology
func (nt *NetworkTopology) GetTopologyStatistics() map[string]interface{} {
	nt.nodesMutex.RLock()
	defer nt.nodesMutex.RUnlock()

	stats := map[string]interface{}{
		"type":              nt.Type,
		"node_count":        len(nt.nodes),
		"active_nodes":      0,
		"inactive_nodes":    0,
		"total_connections": 0,
		"average_latency":   nt.CalculateAverageLatency(),
		"average_bandwidth": nt.CalculateAverageBandwidth(),
	}

	for _, node := range nt.nodes {
		if node.Status == NodeStatusActive {
			stats["active_nodes"] = stats["active_nodes"].(int) + 1
		} else {
			stats["inactive_nodes"] = stats["inactive_nodes"].(int) + 1
		}
		stats["total_connections"] = stats["total_connections"].(int) + len(node.Connections)
	}

	return stats
}

// GeographicDistributionStrategy methods

// GetName returns the strategy name
func (gds *GeographicDistributionStrategy) GetName() string {
	return "geographic"
}

// SelectTargetNodes selects target nodes based on geographic distribution
func (gds *GeographicDistributionStrategy) SelectTargetNodes(sourceNode string, availableNodes []*TopologyNode, criteria map[string]interface{}) ([]*TopologyNode, error) {
	// Group nodes by region
	regionNodes := make(map[string][]*TopologyNode)
	for _, node := range availableNodes {
		region := "default"
		if regionValue, exists := node.Metadata["region"]; exists {
			if regionStr, ok := regionValue.(string); ok {
				region = regionStr
			}
		}
		regionNodes[region] = append(regionNodes[region], node)
	}

	// Select nodes from different regions
	var selectedNodes []*TopologyNode
	maxNodes := 3 // Default max nodes
	if maxNodesValue, exists := criteria["max_nodes"]; exists {
		if maxNodesInt, ok := maxNodesValue.(int); ok {
			maxNodes = maxNodesInt
		}
	}

	// Select best node from each region
	for region, nodes := range regionNodes {
		if len(selectedNodes) >= maxNodes {
			break
		}

		// Find best node in this region
		var bestNode *TopologyNode
		bestScore := -1.0

		for _, node := range nodes {
			score := gds.calculateNodeScore(node, region)
			if score > bestScore {
				bestScore = score
				bestNode = node
			}
		}

		if bestNode != nil {
			selectedNodes = append(selectedNodes, bestNode)
		}
	}

	return selectedNodes, nil
}

// CalculateDistributionPlan calculates a distribution plan for a model
func (gds *GeographicDistributionStrategy) CalculateDistributionPlan(modelName string, topology *NetworkTopology) (*DistributionPlan, error) {
	activeNodes := topology.GetActiveNodes()
	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	// Select target nodes using geographic strategy
	criteria := map[string]interface{}{
		"max_nodes": 3,
	}
	targetNodes, err := gds.SelectTargetNodes("", activeNodes, criteria)
	if err != nil {
		return nil, err
	}

	// Extract node IDs
	var targetNodeIDs []string
	for _, node := range targetNodes {
		targetNodeIDs = append(targetNodeIDs, node.ID)
	}

	plan := &DistributionPlan{
		ModelName:     modelName,
		TargetNodes:   targetNodeIDs,
		Strategy:      gds.GetName(),
		Priority:      1,
		EstimatedTime: 5 * time.Minute, // Estimate based on network conditions
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
	}

	return plan, nil
}

// OptimizeDistribution optimizes an existing distribution plan
func (gds *GeographicDistributionStrategy) OptimizeDistribution(currentPlan *DistributionPlan, topology *NetworkTopology) (*DistributionPlan, error) {
	// For geographic strategy, optimization focuses on regional distribution
	optimizedPlan := *currentPlan
	optimizedPlan.Metadata["optimized"] = true
	optimizedPlan.Metadata["optimization_time"] = time.Now()

	return &optimizedPlan, nil
}

// calculateNodeScore calculates a score for node selection
func (gds *GeographicDistributionStrategy) calculateNodeScore(node *TopologyNode, region string) float64 {
	score := 0.0

	// Health score
	score += node.Health * 100.0

	// Reliability score
	score += node.Reliability * 50.0

	// Regional priority score
	if priority, exists := gds.regionPriority[region]; exists {
		score += float64(priority) * 10.0
	}

	// Bandwidth score (higher is better)
	if node.Bandwidth > 0 {
		score += float64(node.Bandwidth) / 1000000.0 // Convert to MB/s
	}

	return score
}

// SetRegionPriority sets the priority for a region
func (gds *GeographicDistributionStrategy) SetRegionPriority(region string, priority int) {
	gds.regionPriority[region] = priority
}

// LoadBalancedDistributionStrategy methods

// GetName returns the strategy name
func (lbds *LoadBalancedDistributionStrategy) GetName() string {
	return "load_balanced"
}

// SelectTargetNodes selects target nodes based on load balancing
func (lbds *LoadBalancedDistributionStrategy) SelectTargetNodes(sourceNode string, availableNodes []*TopologyNode, criteria map[string]interface{}) ([]*TopologyNode, error) {
	// Filter nodes by load threshold
	var eligibleNodes []*TopologyNode
	for _, node := range availableNodes {
		if lbds.calculateNodeLoad(node) < lbds.maxLoadThreshold {
			eligibleNodes = append(eligibleNodes, node)
		}
	}

	if len(eligibleNodes) == 0 {
		return nil, fmt.Errorf("no nodes available under load threshold")
	}

	// Sort by load (ascending)
	for i := 0; i < len(eligibleNodes)-1; i++ {
		for j := i + 1; j < len(eligibleNodes); j++ {
			if lbds.calculateNodeLoad(eligibleNodes[i]) > lbds.calculateNodeLoad(eligibleNodes[j]) {
				eligibleNodes[i], eligibleNodes[j] = eligibleNodes[j], eligibleNodes[i]
			}
		}
	}

	// Select top nodes
	maxNodes := 3
	if maxNodesValue, exists := criteria["max_nodes"]; exists {
		if maxNodesInt, ok := maxNodesValue.(int); ok {
			maxNodes = maxNodesInt
		}
	}

	if maxNodes > len(eligibleNodes) {
		maxNodes = len(eligibleNodes)
	}

	return eligibleNodes[:maxNodes], nil
}

// CalculateDistributionPlan calculates a load-balanced distribution plan
func (lbds *LoadBalancedDistributionStrategy) CalculateDistributionPlan(modelName string, topology *NetworkTopology) (*DistributionPlan, error) {
	activeNodes := topology.GetActiveNodes()
	if len(activeNodes) == 0 {
		return nil, fmt.Errorf("no active nodes available")
	}

	criteria := map[string]interface{}{
		"max_nodes": 3,
	}
	targetNodes, err := lbds.SelectTargetNodes("", activeNodes, criteria)
	if err != nil {
		return nil, err
	}

	var targetNodeIDs []string
	for _, node := range targetNodes {
		targetNodeIDs = append(targetNodeIDs, node.ID)
	}

	plan := &DistributionPlan{
		ModelName:     modelName,
		TargetNodes:   targetNodeIDs,
		Strategy:      lbds.GetName(),
		Priority:      1,
		EstimatedTime: 3 * time.Minute,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
	}

	return plan, nil
}

// OptimizeDistribution optimizes a load-balanced distribution plan
func (lbds *LoadBalancedDistributionStrategy) OptimizeDistribution(currentPlan *DistributionPlan, topology *NetworkTopology) (*DistributionPlan, error) {
	optimizedPlan := *currentPlan
	optimizedPlan.Metadata["optimized"] = true
	optimizedPlan.Metadata["optimization_time"] = time.Now()

	return &optimizedPlan, nil
}

// calculateNodeLoad calculates the current load of a node
func (lbds *LoadBalancedDistributionStrategy) calculateNodeLoad(node *TopologyNode) float64 {
	// Simplified load calculation
	// In practice, this would consider CPU, memory, network usage, etc.
	baseLoad := 1.0 - node.Health // Lower health = higher load

	// Add connection-based load
	connectionLoad := float64(len(node.Connections)) * 0.1

	return baseLoad + connectionLoad
}
