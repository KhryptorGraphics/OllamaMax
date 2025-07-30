package consensus

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/libp2p/go-libp2p/core/peer"
)

// LeaderElectionManager manages advanced leader election with priority-based selection
type LeaderElectionManager struct {
	engine *Engine
	mu     sync.RWMutex

	// Node capabilities and priorities
	nodeCapabilities map[raft.ServerID]*NodeCapability
	electionHistory  []*ElectionEvent

	// Configuration
	config *ElectionConfig

	// Metrics
	metrics *ElectionMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NodeCapability represents the capabilities of a node for leader election
type NodeCapability struct {
	NodeID raft.ServerID `json:"node_id"`
	PeerID peer.ID       `json:"peer_id"`

	// Hardware capabilities
	CPUCores         int   `json:"cpu_cores"`
	MemoryGB         int   `json:"memory_gb"`
	StorageGB        int   `json:"storage_gb"`
	NetworkBandwidth int64 `json:"network_bandwidth"`

	// Performance metrics
	Latency     time.Duration `json:"latency"`
	Throughput  float64       `json:"throughput"`
	Reliability float64       `json:"reliability"` // 0.0 to 1.0

	// Operational status
	Uptime            time.Duration `json:"uptime"`
	LoadAverage       float64       `json:"load_average"`
	ActiveConnections int           `json:"active_connections"`

	// Geographic information
	Region string `json:"region"`
	Zone   string `json:"zone"`

	// Priority calculation
	Priority    float64   `json:"priority"`
	LastUpdated time.Time `json:"last_updated"`
}

// ElectionEvent represents a leader election event
type ElectionEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	EventType  ElectionEventType `json:"event_type"`
	OldLeader  raft.ServerID     `json:"old_leader,omitempty"`
	NewLeader  raft.ServerID     `json:"new_leader,omitempty"`
	Candidates []raft.ServerID   `json:"candidates"`
	Reason     string            `json:"reason"`
	Duration   time.Duration     `json:"duration"`
}

// ElectionEventType represents the type of election event
type ElectionEventType string

const (
	ElectionStarted   ElectionEventType = "election_started"
	ElectionCompleted ElectionEventType = "election_completed"
	LeaderChanged     ElectionEventType = "leader_changed"
	LeaderFailed      ElectionEventType = "leader_failed"
)

// ElectionConfig configures the leader election system
type ElectionConfig struct {
	// Priority weights for different factors
	HardwareWeight    float64
	PerformanceWeight float64
	ReliabilityWeight float64
	GeographicWeight  float64
	UptimeWeight      float64

	// Election behavior
	MinElectionTimeout time.Duration
	MaxElectionTimeout time.Duration
	PreferredRegions   []string

	// Capability refresh
	CapabilityRefreshInterval time.Duration
	CapabilityTimeout         time.Duration
}

// ElectionMetrics tracks election performance
type ElectionMetrics struct {
	TotalElections      int64         `json:"total_elections"`
	AverageElectionTime time.Duration `json:"average_election_time"`
	LeaderChanges       int64         `json:"leader_changes"`
	FailedElections     int64         `json:"failed_elections"`
	LastElection        time.Time     `json:"last_election"`
	CurrentLeader       raft.ServerID `json:"current_leader"`
}

// NewLeaderElectionManager creates a new leader election manager
func NewLeaderElectionManager(engine *Engine, config *ElectionConfig) *LeaderElectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &ElectionConfig{
			HardwareWeight:            0.3,
			PerformanceWeight:         0.3,
			ReliabilityWeight:         0.2,
			GeographicWeight:          0.1,
			UptimeWeight:              0.1,
			MinElectionTimeout:        150 * time.Millisecond,
			MaxElectionTimeout:        300 * time.Millisecond,
			PreferredRegions:          []string{},
			CapabilityRefreshInterval: 30 * time.Second,
			CapabilityTimeout:         10 * time.Second,
		}
	}

	lem := &LeaderElectionManager{
		engine:           engine,
		nodeCapabilities: make(map[raft.ServerID]*NodeCapability),
		electionHistory:  make([]*ElectionEvent, 0),
		config:           config,
		metrics:          &ElectionMetrics{},
		ctx:              ctx,
		cancel:           cancel,
	}

	// Start background tasks
	lem.wg.Add(2)
	go lem.capabilityRefreshLoop()
	go lem.electionMonitorLoop()

	return lem
}

// UpdateNodeCapability updates the capability information for a node
func (lem *LeaderElectionManager) UpdateNodeCapability(capability *NodeCapability) {
	lem.mu.Lock()
	defer lem.mu.Unlock()

	capability.LastUpdated = time.Now()
	capability.Priority = lem.calculatePriority(capability)
	lem.nodeCapabilities[capability.NodeID] = capability
}

// calculatePriority calculates the leadership priority for a node
func (lem *LeaderElectionManager) calculatePriority(capability *NodeCapability) float64 {
	// Hardware score (0-1)
	hardwareScore := lem.calculateHardwareScore(capability)

	// Performance score (0-1)
	performanceScore := lem.calculatePerformanceScore(capability)

	// Reliability score (already 0-1)
	reliabilityScore := capability.Reliability

	// Geographic score (0-1)
	geographicScore := lem.calculateGeographicScore(capability)

	// Uptime score (0-1)
	uptimeScore := lem.calculateUptimeScore(capability)

	// Weighted average
	priority := (hardwareScore * lem.config.HardwareWeight) +
		(performanceScore * lem.config.PerformanceWeight) +
		(reliabilityScore * lem.config.ReliabilityWeight) +
		(geographicScore * lem.config.GeographicWeight) +
		(uptimeScore * lem.config.UptimeWeight)

	return priority
}

// calculateHardwareScore calculates hardware capability score
func (lem *LeaderElectionManager) calculateHardwareScore(capability *NodeCapability) float64 {
	// Normalize hardware specs (simplified scoring)
	cpuScore := float64(capability.CPUCores) / 32.0                                    // Assume max 32 cores
	memoryScore := float64(capability.MemoryGB) / 128.0                                // Assume max 128GB
	storageScore := float64(capability.StorageGB) / 1000.0                             // Assume max 1TB
	bandwidthScore := float64(capability.NetworkBandwidth) / (10 * 1024 * 1024 * 1024) // 10Gbps

	// Cap at 1.0
	if cpuScore > 1.0 {
		cpuScore = 1.0
	}
	if memoryScore > 1.0 {
		memoryScore = 1.0
	}
	if storageScore > 1.0 {
		storageScore = 1.0
	}
	if bandwidthScore > 1.0 {
		bandwidthScore = 1.0
	}

	return (cpuScore + memoryScore + storageScore + bandwidthScore) / 4.0
}

// calculatePerformanceScore calculates performance score
func (lem *LeaderElectionManager) calculatePerformanceScore(capability *NodeCapability) float64 {
	// Lower latency is better (invert and normalize)
	latencyScore := 1.0 - (float64(capability.Latency.Milliseconds()) / 1000.0) // Assume max 1s latency
	if latencyScore < 0 {
		latencyScore = 0
	}

	// Higher throughput is better (normalize)
	throughputScore := capability.Throughput / 1000.0 // Assume max 1000 ops/sec
	if throughputScore > 1.0 {
		throughputScore = 1.0
	}

	// Lower load average is better (invert and normalize)
	loadScore := 1.0 - (capability.LoadAverage / 10.0) // Assume max load of 10
	if loadScore < 0 {
		loadScore = 0
	}

	return (latencyScore + throughputScore + loadScore) / 3.0
}

// calculateGeographicScore calculates geographic preference score
func (lem *LeaderElectionManager) calculateGeographicScore(capability *NodeCapability) float64 {
	// If no preferred regions, all regions are equal
	if len(lem.config.PreferredRegions) == 0 {
		return 1.0
	}

	// Check if node is in preferred region
	for _, region := range lem.config.PreferredRegions {
		if capability.Region == region {
			return 1.0
		}
	}

	return 0.5 // Partial score for non-preferred regions
}

// calculateUptimeScore calculates uptime score
func (lem *LeaderElectionManager) calculateUptimeScore(capability *NodeCapability) float64 {
	// Normalize uptime (assume max relevant uptime is 30 days)
	maxUptime := 30 * 24 * time.Hour
	uptimeScore := float64(capability.Uptime) / float64(maxUptime)
	if uptimeScore > 1.0 {
		uptimeScore = 1.0
	}

	return uptimeScore
}

// GetBestLeaderCandidate returns the best candidate for leadership
func (lem *LeaderElectionManager) GetBestLeaderCandidate() *NodeCapability {
	lem.mu.RLock()
	defer lem.mu.RUnlock()

	if len(lem.nodeCapabilities) == 0 {
		return nil
	}

	// Convert to slice and sort by priority
	candidates := make([]*NodeCapability, 0, len(lem.nodeCapabilities))
	for _, capability := range lem.nodeCapabilities {
		candidates = append(candidates, capability)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Priority > candidates[j].Priority
	})

	return candidates[0]
}

// GetLeadershipRanking returns nodes ranked by leadership priority
func (lem *LeaderElectionManager) GetLeadershipRanking() []*NodeCapability {
	lem.mu.RLock()
	defer lem.mu.RUnlock()

	candidates := make([]*NodeCapability, 0, len(lem.nodeCapabilities))
	for _, capability := range lem.nodeCapabilities {
		candidates = append(candidates, capability)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Priority > candidates[j].Priority
	})

	return candidates
}

// RecordElectionEvent records an election event
func (lem *LeaderElectionManager) RecordElectionEvent(event *ElectionEvent) {
	lem.mu.Lock()
	defer lem.mu.Unlock()

	lem.electionHistory = append(lem.electionHistory, event)

	// Limit history size
	if len(lem.electionHistory) > 1000 {
		lem.electionHistory = lem.electionHistory[1:]
	}

	// Update metrics
	lem.metrics.TotalElections++
	lem.metrics.LastElection = event.Timestamp

	if event.EventType == ElectionCompleted {
		lem.metrics.CurrentLeader = event.NewLeader
		if event.Duration > 0 {
			// Update average election time
			totalTime := time.Duration(lem.metrics.TotalElections-1)*lem.metrics.AverageElectionTime + event.Duration
			lem.metrics.AverageElectionTime = totalTime / time.Duration(lem.metrics.TotalElections)
		}
	}

	if event.EventType == LeaderChanged {
		lem.metrics.LeaderChanges++
	}
}

// GetElectionMetrics returns election metrics
func (lem *LeaderElectionManager) GetElectionMetrics() *ElectionMetrics {
	lem.mu.RLock()
	defer lem.mu.RUnlock()

	metrics := *lem.metrics
	return &metrics
}

// capabilityRefreshLoop periodically refreshes node capabilities
func (lem *LeaderElectionManager) capabilityRefreshLoop() {
	defer lem.wg.Done()

	ticker := time.NewTicker(lem.config.CapabilityRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lem.ctx.Done():
			return
		case <-ticker.C:
			lem.refreshCapabilities()
		}
	}
}

// refreshCapabilities refreshes capabilities for all known nodes
func (lem *LeaderElectionManager) refreshCapabilities() {
	lem.mu.RLock()
	nodeIDs := make([]raft.ServerID, 0, len(lem.nodeCapabilities))
	for nodeID := range lem.nodeCapabilities {
		nodeIDs = append(nodeIDs, nodeID)
	}
	lem.mu.RUnlock()

	// In a real implementation, you would query each node for its current capabilities
	// For now, we'll just update the priority scores based on existing data
	for _, nodeID := range nodeIDs {
		lem.mu.Lock()
		if capability, exists := lem.nodeCapabilities[nodeID]; exists {
			capability.Priority = lem.calculatePriority(capability)
			capability.LastUpdated = time.Now()
		}
		lem.mu.Unlock()
	}
}

// electionMonitorLoop monitors election events
func (lem *LeaderElectionManager) electionMonitorLoop() {
	defer lem.wg.Done()

	// Monitor leadership changes
	for {
		select {
		case <-lem.ctx.Done():
			return
		case isLeader := <-lem.engine.leaderCh:
			if isLeader {
				lem.RecordElectionEvent(&ElectionEvent{
					Timestamp: time.Now(),
					EventType: ElectionCompleted,
					NewLeader: raft.ServerID(lem.engine.p2p.ID().String()),
					Reason:    "Leadership acquired",
				})
			} else {
				lem.RecordElectionEvent(&ElectionEvent{
					Timestamp: time.Now(),
					EventType: LeaderChanged,
					OldLeader: raft.ServerID(lem.engine.p2p.ID().String()),
					Reason:    "Leadership lost",
				})
			}
		}
	}
}

// Close closes the leader election manager
func (lem *LeaderElectionManager) Close() error {
	lem.cancel()
	lem.wg.Wait()
	return nil
}
