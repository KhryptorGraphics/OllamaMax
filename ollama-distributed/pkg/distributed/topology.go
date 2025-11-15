package distributed

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// ClusterTier represents the network tier of a cluster
type ClusterTier int

const (
	TierLocal    ClusterTier = iota // LAN/VPN cluster (<10ms latency)
	TierRegional                    // Regional cluster (10-50ms latency)
	TierGlobal                      // Global cluster (>50ms latency)
)

func (t ClusterTier) String() string {
	switch t {
	case TierLocal:
		return "local"
	case TierRegional:
		return "regional"
	case TierGlobal:
		return "global"
	default:
		return "unknown"
	}
}

// NetworkMetrics represents network quality metrics for a peer
type NetworkMetrics struct {
	RTT          time.Duration // Round-trip time
	Bandwidth    float64       // Estimated bandwidth in MB/s
	PacketLoss   float64       // Packet loss rate (0.0-1.0)
	Jitter       time.Duration // Network jitter
	LastMeasured time.Time     // Last measurement timestamp
	MeasureCount int           // Number of measurements
	AvgRTT       time.Duration // Average RTT over time
	AvgBandwidth float64       // Average bandwidth over time
}

// ClusterNode represents a node in the cluster topology
type ClusterNode struct {
	PeerID        peer.ID
	ClusterID     string
	Tier          ClusterTier
	Metrics       *NetworkMetrics
	IsCoordinator bool
	LastSeen      time.Time
	Capabilities  map[string]interface{} // GPU memory, CPU cores, etc.
}

// Cluster represents a group of nodes with similar network characteristics
type Cluster struct {
	ID            string
	Tier          ClusterTier
	Coordinator   peer.ID
	Nodes         map[peer.ID]*ClusterNode
	NodesMux      sync.RWMutex
	CreatedAt     time.Time
	LastUpdate    time.Time
	AvgLatency    time.Duration      // Average intra-cluster latency
	TotalCapacity map[string]float64 // Aggregated capabilities
}

// ClusterTopology manages the hierarchical cluster structure
type ClusterTopology struct {
	LocalNodeID peer.ID
	Clusters    map[string]*Cluster
	ClustersMux sync.RWMutex

	// Node to cluster mapping
	NodeCluster map[peer.ID]string
	NodeMux     sync.RWMutex

	// Local cluster (this node's cluster)
	LocalClusterID string

	// Configuration
	LocalLatencyThreshold    time.Duration // <10ms = local
	RegionalLatencyThreshold time.Duration // <50ms = regional
	MeasurementInterval      time.Duration // How often to measure
	ClusterFormationDelay    time.Duration // Wait before forming cluster

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	metrics *TopologyMetrics
}

// TopologyMetrics tracks topology performance
type TopologyMetrics struct {
	TotalClusters     int
	LocalClusters     int
	RegionalClusters  int
	GlobalClusters    int
	TotalNodes        int
	ClusterFormations int
	ClusterMerges     int
	ClusterSplits     int
	LastUpdate        time.Time
}

// NewClusterTopology creates a new hierarchical cluster topology manager
func NewClusterTopology(ctx context.Context, localNodeID peer.ID) *ClusterTopology {
	ctx, cancel := context.WithCancel(ctx)

	return &ClusterTopology{
		LocalNodeID:              localNodeID,
		Clusters:                 make(map[string]*Cluster),
		NodeCluster:              make(map[peer.ID]string),
		LocalLatencyThreshold:    10 * time.Millisecond,
		RegionalLatencyThreshold: 50 * time.Millisecond,
		MeasurementInterval:      30 * time.Second,
		ClusterFormationDelay:    5 * time.Second,
		ctx:                      ctx,
		cancel:                   cancel,
		metrics: &TopologyMetrics{
			LastUpdate: time.Now(),
		},
	}
}

// Start starts the topology manager
func (ct *ClusterTopology) Start() error {
	log.Printf("Starting cluster topology manager for node %s", ct.LocalNodeID)

	// Start periodic tasks
	ct.wg.Add(2)
	go ct.measurementTask()
	go ct.clusterMaintenanceTask()

	return nil
}

// Stop stops the topology manager
func (ct *ClusterTopology) Stop() error {
	log.Printf("Stopping cluster topology manager")
	ct.cancel()
	ct.wg.Wait()
	return nil
}

// AddNode adds a node to the topology with initial metrics
func (ct *ClusterTopology) AddNode(peerID peer.ID, metrics *NetworkMetrics) error {
	// Determine tier based on latency
	tier := ct.determineTier(metrics.RTT)

	// Create cluster node
	node := &ClusterNode{
		PeerID:       peerID,
		Tier:         tier,
		Metrics:      metrics,
		LastSeen:     time.Now(),
		Capabilities: make(map[string]interface{}),
	}

	// Find or create appropriate cluster
	clusterID, err := ct.findOrCreateCluster(node)
	if err != nil {
		return fmt.Errorf("failed to assign node to cluster: %w", err)
	}

	node.ClusterID = clusterID

	// Add to cluster
	ct.ClustersMux.Lock()
	cluster := ct.Clusters[clusterID]
	ct.ClustersMux.Unlock()

	cluster.NodesMux.Lock()
	cluster.Nodes[peerID] = node
	cluster.NodesMux.Unlock()

	// Update node mapping
	ct.NodeMux.Lock()
	ct.NodeCluster[peerID] = clusterID
	ct.NodeMux.Unlock()

	log.Printf("Added node %s to cluster %s (tier: %s, RTT: %v)", peerID, clusterID, tier, metrics.RTT)

	return nil
}

// determineTier determines the cluster tier based on latency
func (ct *ClusterTopology) determineTier(rtt time.Duration) ClusterTier {
	if rtt < ct.LocalLatencyThreshold {
		return TierLocal
	} else if rtt < ct.RegionalLatencyThreshold {
		return TierRegional
	}
	return TierGlobal
}

// findOrCreateCluster finds an appropriate cluster for a node or creates a new one
func (ct *ClusterTopology) findOrCreateCluster(node *ClusterNode) (string, error) {
	ct.ClustersMux.RLock()
	// Try to find existing cluster of same tier with similar latency
	for clusterID, cluster := range ct.Clusters {
		if cluster.Tier == node.Tier {
			// Check if latency is compatible (within 20% of cluster average)
			latencyDiff := float64(node.Metrics.RTT-cluster.AvgLatency) / float64(cluster.AvgLatency)
			if latencyDiff < 0 {
				latencyDiff = -latencyDiff
			}

			if latencyDiff < 0.2 {
				ct.ClustersMux.RUnlock()
				return clusterID, nil
			}
		}
	}
	ct.ClustersMux.RUnlock()

	// No suitable cluster found, create new one
	return ct.createCluster(node)
}

// createCluster creates a new cluster for a node
func (ct *ClusterTopology) createCluster(node *ClusterNode) (string, error) {
	clusterID := fmt.Sprintf("cluster-%s-%d", node.Tier, time.Now().UnixNano())

	cluster := &Cluster{
		ID:            clusterID,
		Tier:          node.Tier,
		Coordinator:   node.PeerID, // First node becomes coordinator
		Nodes:         make(map[peer.ID]*ClusterNode),
		CreatedAt:     time.Now(),
		LastUpdate:    time.Now(),
		AvgLatency:    node.Metrics.RTT,
		TotalCapacity: make(map[string]float64),
	}

	ct.ClustersMux.Lock()
	ct.Clusters[clusterID] = cluster
	ct.ClustersMux.Unlock()

	ct.metrics.TotalClusters++
	switch node.Tier {
	case TierLocal:
		ct.metrics.LocalClusters++
	case TierRegional:
		ct.metrics.RegionalClusters++
	case TierGlobal:
		ct.metrics.GlobalClusters++
	}

	log.Printf("Created new cluster %s (tier: %s)", clusterID, node.Tier)

	return clusterID, nil
}

// UpdateNodeMetrics updates network metrics for a node
func (ct *ClusterTopology) UpdateNodeMetrics(peerID peer.ID, metrics *NetworkMetrics) error {
	ct.NodeMux.RLock()
	clusterID, exists := ct.NodeCluster[peerID]
	ct.NodeMux.RUnlock()

	if !exists {
		return fmt.Errorf("node %s not found in topology", peerID)
	}

	ct.ClustersMux.RLock()
	cluster := ct.Clusters[clusterID]
	ct.ClustersMux.RUnlock()

	cluster.NodesMux.Lock()
	node, exists := cluster.Nodes[peerID]
	if !exists {
		cluster.NodesMux.Unlock()
		return fmt.Errorf("node %s not found in cluster %s", peerID, clusterID)
	}

	// Update metrics with exponential moving average
	alpha := 0.3 // Smoothing factor
	node.Metrics.AvgRTT = time.Duration(float64(node.Metrics.AvgRTT)*(1-alpha) + float64(metrics.RTT)*alpha)
	node.Metrics.AvgBandwidth = node.Metrics.AvgBandwidth*(1-alpha) + metrics.Bandwidth*alpha
	node.Metrics.RTT = metrics.RTT
	node.Metrics.Bandwidth = metrics.Bandwidth
	node.Metrics.PacketLoss = metrics.PacketLoss
	node.Metrics.Jitter = metrics.Jitter
	node.Metrics.LastMeasured = time.Now()
	node.Metrics.MeasureCount++
	node.LastSeen = time.Now()

	cluster.NodesMux.Unlock()

	// Check if node should be moved to different tier
	newTier := ct.determineTier(node.Metrics.AvgRTT)
	if newTier != node.Tier {
		log.Printf("Node %s tier changed from %s to %s (RTT: %v)", peerID, node.Tier, newTier, node.Metrics.AvgRTT)
		return ct.moveNodeToTier(peerID, newTier)
	}

	return nil
}

// moveNodeToTier moves a node to a different tier
func (ct *ClusterTopology) moveNodeToTier(peerID peer.ID, newTier ClusterTier) error {
	// Remove from current cluster
	ct.NodeMux.RLock()
	oldClusterID := ct.NodeCluster[peerID]
	ct.NodeMux.RUnlock()

	ct.ClustersMux.RLock()
	oldCluster := ct.Clusters[oldClusterID]
	ct.ClustersMux.RUnlock()

	oldCluster.NodesMux.Lock()
	node := oldCluster.Nodes[peerID]
	delete(oldCluster.Nodes, peerID)
	oldCluster.NodesMux.Unlock()

	// Update tier
	node.Tier = newTier

	// Find or create new cluster
	newClusterID, err := ct.findOrCreateCluster(node)
	if err != nil {
		return fmt.Errorf("failed to move node to new tier: %w", err)
	}

	node.ClusterID = newClusterID

	// Add to new cluster
	ct.ClustersMux.RLock()
	newCluster := ct.Clusters[newClusterID]
	ct.ClustersMux.RUnlock()

	newCluster.NodesMux.Lock()
	newCluster.Nodes[peerID] = node
	newCluster.NodesMux.Unlock()

	// Update mapping
	ct.NodeMux.Lock()
	ct.NodeCluster[peerID] = newClusterID
	ct.NodeMux.Unlock()

	log.Printf("Moved node %s from cluster %s to %s (tier: %s)", peerID, oldClusterID, newClusterID, newTier)

	return nil
}

// GetCluster returns a cluster by ID
func (ct *ClusterTopology) GetCluster(clusterID string) (*Cluster, error) {
	ct.ClustersMux.RLock()
	defer ct.ClustersMux.RUnlock()

	cluster, exists := ct.Clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterID)
	}

	return cluster, nil
}

// GetNodeCluster returns the cluster ID for a node
func (ct *ClusterTopology) GetNodeCluster(peerID peer.ID) (string, error) {
	ct.NodeMux.RLock()
	defer ct.NodeMux.RUnlock()

	clusterID, exists := ct.NodeCluster[peerID]
	if !exists {
		return "", fmt.Errorf("node %s not found in topology", peerID)
	}

	return clusterID, nil
}

// GetLocalCluster returns the local cluster (this node's cluster)
func (ct *ClusterTopology) GetLocalCluster() (*Cluster, error) {
	if ct.LocalClusterID == "" {
		return nil, fmt.Errorf("local cluster not set")
	}

	return ct.GetCluster(ct.LocalClusterID)
}

// GetClustersByTier returns all clusters of a specific tier
func (ct *ClusterTopology) GetClustersByTier(tier ClusterTier) []*Cluster {
	ct.ClustersMux.RLock()
	defer ct.ClustersMux.RUnlock()

	clusters := make([]*Cluster, 0)
	for _, cluster := range ct.Clusters {
		if cluster.Tier == tier {
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// measurementTask periodically measures network metrics
func (ct *ClusterTopology) measurementTask() {
	defer ct.wg.Done()

	ticker := time.NewTicker(ct.MeasurementInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ct.ctx.Done():
			return
		case <-ticker.C:
			ct.performMeasurements()
		}
	}
}

// performMeasurements measures network metrics for all nodes
func (ct *ClusterTopology) performMeasurements() {
	// TODO: Implement actual network measurements (ping, bandwidth test)
	// For now, this is a placeholder
	log.Printf("Performing network measurements for %d clusters", len(ct.Clusters))
}

// clusterMaintenanceTask performs periodic cluster maintenance
func (ct *ClusterTopology) clusterMaintenanceTask() {
	defer ct.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ct.ctx.Done():
			return
		case <-ticker.C:
			ct.performMaintenance()
		}
	}
}

// performMaintenance performs cluster maintenance tasks
func (ct *ClusterTopology) performMaintenance() {
	// Remove stale nodes
	ct.removeStaleNodes()

	// Rebalance clusters if needed
	ct.rebalanceClusters()

	// Update metrics
	ct.updateMetrics()
}

// removeStaleNodes removes nodes that haven't been seen recently
func (ct *ClusterTopology) removeStaleNodes() {
	staleThreshold := 5 * time.Minute

	ct.ClustersMux.RLock()
	clusters := make([]*Cluster, 0, len(ct.Clusters))
	for _, cluster := range ct.Clusters {
		clusters = append(clusters, cluster)
	}
	ct.ClustersMux.RUnlock()

	for _, cluster := range clusters {
		cluster.NodesMux.Lock()
		for peerID, node := range cluster.Nodes {
			if time.Since(node.LastSeen) > staleThreshold {
				log.Printf("Removing stale node %s from cluster %s", peerID, cluster.ID)
				delete(cluster.Nodes, peerID)

				ct.NodeMux.Lock()
				delete(ct.NodeCluster, peerID)
				ct.NodeMux.Unlock()
			}
		}
		cluster.NodesMux.Unlock()
	}
}

// rebalanceClusters rebalances clusters if needed
func (ct *ClusterTopology) rebalanceClusters() {
	// TODO: Implement cluster rebalancing logic
	// - Merge small clusters
	// - Split large clusters
	// - Elect new coordinators if needed
}

// updateMetrics updates topology metrics
func (ct *ClusterTopology) updateMetrics() {
	ct.ClustersMux.RLock()
	defer ct.ClustersMux.RUnlock()

	ct.metrics.TotalClusters = len(ct.Clusters)
	ct.metrics.TotalNodes = 0
	ct.metrics.LocalClusters = 0
	ct.metrics.RegionalClusters = 0
	ct.metrics.GlobalClusters = 0

	for _, cluster := range ct.Clusters {
		cluster.NodesMux.RLock()
		ct.metrics.TotalNodes += len(cluster.Nodes)
		cluster.NodesMux.RUnlock()

		switch cluster.Tier {
		case TierLocal:
			ct.metrics.LocalClusters++
		case TierRegional:
			ct.metrics.RegionalClusters++
		case TierGlobal:
			ct.metrics.GlobalClusters++
		}
	}

	ct.metrics.LastUpdate = time.Now()
}

// GetMetrics returns current topology metrics
func (ct *ClusterTopology) GetMetrics() *TopologyMetrics {
	return ct.metrics
}
