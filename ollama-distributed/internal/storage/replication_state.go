package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// NodeManager manages storage nodes in the cluster
type NodeManager struct {
	logger *slog.Logger

	// Node registry
	nodes      map[string]*StorageNode
	nodesMutex sync.RWMutex

	// Node selection
	selector *NodeSelector

	// Health monitoring
	healthChecker *NodeHealthChecker

	// Configuration
	config *NodeManagerConfig
}

// NodeManagerConfig contains configuration for node management
type NodeManagerConfig struct {
	HeartbeatInterval   time.Duration `json:"heartbeat_interval"`
	FailureTimeout      time.Duration `json:"failure_timeout"`
	MaxFailures         int           `json:"max_failures"`
	EnableLoadBalancing bool          `json:"enable_load_balancing"`
	PreferLocalReplicas bool          `json:"prefer_local_replicas"`
}

// StorageNode represents a storage node in the cluster
type StorageNode struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	Region       string                 `json:"region"`
	Zone         string                 `json:"zone"`
	Status       string                 `json:"status"` // active, inactive, draining, failed
	Connected    bool                   `json:"connected"`
	LastSeen     time.Time              `json:"last_seen"`
	JoinedAt     time.Time              `json:"joined_at"`
	Version      string                 `json:"version"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Capacity and performance
	Capacity *NodeCapacity     `json:"capacity"`
	Health   *NodeHealthStatus `json:"health"`

	// Load balancing
	LoadFactor   float64       `json:"load_factor"`
	RequestCount int64         `json:"request_count"`
	ErrorCount   int64         `json:"error_count"`
	ResponseTime time.Duration `json:"response_time"`

	mutex sync.RWMutex
}

// NodeCapacity represents node storage capacity
type NodeCapacity struct {
	TotalBytes     int64   `json:"total_bytes"`
	UsedBytes      int64   `json:"used_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	ObjectCount    int64   `json:"object_count"`
}

// NodeHealthStatus represents node health information
type NodeHealthStatus struct {
	Status         string          `json:"status"` // healthy, degraded, unhealthy, down
	LastCheck      time.Time       `json:"last_check"`
	Checks         map[string]bool `json:"checks"`
	Errors         []string        `json:"errors"`
	Warnings       []string        `json:"warnings"`
	ResponseTime   time.Duration   `json:"response_time"`
	SuccessRate    float64         `json:"success_rate"`
	TotalRequests  int64           `json:"total_requests"`
	FailedRequests int64           `json:"failed_requests"`
}

// NodeSelector selects optimal nodes for replication
type NodeSelector struct {
	manager *NodeManager
	logger  *slog.Logger

	// Selection strategies
	strategies map[string]SelectionStrategy
}

// SelectionStrategy defines node selection interface
type SelectionStrategy interface {
	SelectNodes(availableNodes []*StorageNode, count int, constraints map[string]interface{}) ([]*StorageNode, error)
}

// LoadBalancedSelection implements load-balanced node selection
type LoadBalancedSelection struct{}

// GeographicSelection implements geographic-aware node selection
type GeographicSelection struct {
	preferredRegions []string
}

// CapacityBasedSelection implements capacity-based node selection
type CapacityBasedSelection struct{}

// NodeHealthChecker monitors node health
type NodeHealthChecker struct {
	manager *NodeManager
	logger  *slog.Logger

	// Health check configuration
	interval time.Duration

	// Background task
	ctx    context.Context
	cancel context.CancelFunc
}

// ReplicationHealth tracks overall replication health
type ReplicationHealth struct {
	OverallStatus    string                  `json:"overall_status"`
	HealthyNodes     int                     `json:"healthy_nodes"`
	TotalNodes       int                     `json:"total_nodes"`
	ActiveOperations int                     `json:"active_operations"`
	FailedOperations int                     `json:"failed_operations"`
	AverageLatency   time.Duration           `json:"average_latency"`
	ThroughputMBps   float64                 `json:"throughput_mbps"`
	LastHealthCheck  time.Time               `json:"last_health_check"`
	RegionHealth     map[string]RegionHealth `json:"region_health"`
}

// RegionHealth tracks health per region
type RegionHealth struct {
	Region         string        `json:"region"`
	HealthyNodes   int           `json:"healthy_nodes"`
	TotalNodes     int           `json:"total_nodes"`
	AverageLatency time.Duration `json:"average_latency"`
	Status         string        `json:"status"`
}

// NewNodeManager creates a new node manager
func NewNodeManager(config *NodeManagerConfig, logger *slog.Logger) (*NodeManager, error) {
	nm := &NodeManager{
		logger: logger,
		nodes:  make(map[string]*StorageNode),
		config: config,
	}

	// Initialize node selector
	nm.selector = &NodeSelector{
		manager: nm,
		logger:  logger,
		strategies: map[string]SelectionStrategy{
			"load_balanced": &LoadBalancedSelection{},
			"geographic":    &GeographicSelection{},
			"capacity":      &CapacityBasedSelection{},
		},
	}

	// Initialize health checker
	nm.healthChecker = &NodeHealthChecker{
		manager:  nm,
		logger:   logger,
		interval: config.HeartbeatInterval,
	}

	return nm, nil
}

// Start starts the node manager
func (nm *NodeManager) Start(ctx context.Context) error {
	// Start health checker
	return nm.healthChecker.Start(ctx)
}

// Stop stops the node manager
func (nm *NodeManager) Stop(ctx context.Context) error {
	// Stop health checker
	return nm.healthChecker.Stop(ctx)
}

// AddNode adds a node to the cluster
func (nm *NodeManager) AddNode(ctx context.Context, node *StorageNode) error {
	nm.nodesMutex.Lock()
	defer nm.nodesMutex.Unlock()

	node.JoinedAt = time.Now()
	node.LastSeen = time.Now()
	node.Status = "active"
	node.Connected = true

	nm.nodes[node.ID] = node
	nm.logger.Info("node added", "node_id", node.ID, "address", node.Address)

	return nil
}

// RemoveNode removes a node from the cluster
func (nm *NodeManager) RemoveNode(ctx context.Context, nodeID string) error {
	nm.nodesMutex.Lock()
	defer nm.nodesMutex.Unlock()

	if node, exists := nm.nodes[nodeID]; exists {
		node.Status = "leaving"
		delete(nm.nodes, nodeID)
		nm.logger.Info("node removed", "node_id", nodeID)
	} else {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	return nil
}

// GetAllNodes returns all nodes in the cluster
func (nm *NodeManager) GetAllNodes() []*StorageNode {
	nm.nodesMutex.RLock()
	defer nm.nodesMutex.RUnlock()

	nodes := make([]*StorageNode, 0, len(nm.nodes))
	for _, node := range nm.nodes {
		// Create a copy to avoid race conditions
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// GetHealthyNodes returns only healthy nodes
func (nm *NodeManager) GetHealthyNodes() []*StorageNode {
	allNodes := nm.GetAllNodes()
	var healthyNodes []*StorageNode

	for _, node := range allNodes {
		if node.Health.Status == "healthy" && node.Connected {
			healthyNodes = append(healthyNodes, node)
		}
	}

	return healthyNodes
}

// SelectNodes selects nodes using the specified strategy
func (nm *NodeManager) SelectNodes(strategy string, count int, constraints map[string]interface{}) ([]*StorageNode, error) {
	return nm.selector.SelectNodes(strategy, count, constraints)
}

// NodeSelector implementation

// SelectNodes selects nodes using the specified strategy
func (ns *NodeSelector) SelectNodes(strategy string, count int, constraints map[string]interface{}) ([]*StorageNode, error) {
	availableNodes := ns.manager.GetHealthyNodes()

	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	selectionStrategy, exists := ns.strategies[strategy]
	if !exists {
		selectionStrategy = ns.strategies["load_balanced"] // Default strategy
	}

	return selectionStrategy.SelectNodes(availableNodes, count, constraints)
}

// Selection strategy implementations

// LoadBalancedSelection selects nodes based on load
func (lbs *LoadBalancedSelection) SelectNodes(availableNodes []*StorageNode, count int, constraints map[string]interface{}) ([]*StorageNode, error) {
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Sort by load factor (ascending)
	sortedNodes := make([]*StorageNode, len(availableNodes))
	copy(sortedNodes, availableNodes)

	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].LoadFactor < sortedNodes[j].LoadFactor
	})

	if count > len(sortedNodes) {
		count = len(sortedNodes)
	}

	return sortedNodes[:count], nil
}

// GeographicSelection selects nodes based on geographic distribution
func (gs *GeographicSelection) SelectNodes(availableNodes []*StorageNode, count int, constraints map[string]interface{}) ([]*StorageNode, error) {
	// TODO: Implement geographic selection based on regions/zones
	return availableNodes[:min(count, len(availableNodes))], nil
}

// CapacityBasedSelection selects nodes based on available capacity
func (cbs *CapacityBasedSelection) SelectNodes(availableNodes []*StorageNode, count int, constraints map[string]interface{}) ([]*StorageNode, error) {
	// Sort by available capacity (descending)
	sortedNodes := make([]*StorageNode, len(availableNodes))
	copy(sortedNodes, availableNodes)

	sort.Slice(sortedNodes, func(i, j int) bool {
		if sortedNodes[i].Capacity == nil || sortedNodes[j].Capacity == nil {
			return false
		}
		return sortedNodes[i].Capacity.AvailableBytes > sortedNodes[j].Capacity.AvailableBytes
	})

	if count > len(sortedNodes) {
		count = len(sortedNodes)
	}

	return sortedNodes[:count], nil
}

// Utility functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NodeHealthChecker implementation

// Start starts the health checker
func (nhc *NodeHealthChecker) Start(ctx context.Context) error {
	nhc.ctx, nhc.cancel = context.WithCancel(ctx)

	go nhc.healthCheckRoutine()

	nhc.logger.Info("node health checker started")
	return nil
}

// Stop stops the health checker
func (nhc *NodeHealthChecker) Stop(ctx context.Context) error {
	if nhc.cancel != nil {
		nhc.cancel()
	}

	nhc.logger.Info("node health checker stopped")
	return nil
}

// healthCheckRoutine runs periodic health checks
func (nhc *NodeHealthChecker) healthCheckRoutine() {
	ticker := time.NewTicker(nhc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-nhc.ctx.Done():
			return
		case <-ticker.C:
			nhc.checkAllNodes()
		}
	}
}

// checkAllNodes checks health of all nodes
func (nhc *NodeHealthChecker) checkAllNodes() {
	nodes := nhc.manager.GetAllNodes()

	for _, node := range nodes {
		go nhc.checkNode(node)
	}
}

// checkNode performs health check on a single node
func (nhc *NodeHealthChecker) checkNode(node *StorageNode) {
	start := time.Now()

	// TODO: Implement actual health checks
	// This would involve network calls to the node

	// Simulate health check
	node.mutex.Lock()
	defer node.mutex.Unlock()

	node.LastSeen = time.Now()
	node.Health.LastCheck = time.Now()
	node.Health.ResponseTime = time.Since(start)

	// Update health status based on response time and other factors
	if node.Health.ResponseTime < 100*time.Millisecond {
		node.Health.Status = "healthy"
	} else if node.Health.ResponseTime < 500*time.Millisecond {
		node.Health.Status = "degraded"
	} else {
		node.Health.Status = "unhealthy"
	}

	// Update success rate
	node.Health.TotalRequests++
	if node.Health.Status == "healthy" {
		node.Health.SuccessRate = float64(node.Health.TotalRequests-node.Health.FailedRequests) / float64(node.Health.TotalRequests)
	} else {
		node.Health.FailedRequests++
		node.Health.SuccessRate = float64(node.Health.TotalRequests-node.Health.FailedRequests) / float64(node.Health.TotalRequests)
	}
}
