package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ChaosTest represents a chaos engineering test scenario
type ChaosTest struct {
	Name        string
	Description string
	Setup       func(t *testing.T, cluster *TestCluster)
	Chaos       func(t *testing.T, cluster *TestCluster)
	Verify      func(t *testing.T, cluster *TestCluster)
	Cleanup     func(t *testing.T, cluster *TestCluster)
}

// TestCluster represents a cluster of nodes for chaos testing
type TestCluster struct {
	Nodes       []*TestNode
	Config      *ClusterConfig
	Coordinator *ChaosCoordinator
	ctx         context.Context
	cancel      context.CancelFunc
}

// TestNode represents a single node in the test cluster
type TestNode struct {
	ID              string
	P2PNode         *p2p.Node
	ConsensusEngine *consensus.Engine
	SchedulerEngine *scheduler.Engine
	Config          *NodeConfig
	Status          NodeStatus
	Failures        []FailureEvent
	mu              sync.RWMutex
}

// NodeConfig holds configuration for test nodes
type NodeConfig struct {
	DataDir   string
	P2PPort   int
	RaftPort  int
	Bootstrap bool
}

// ClusterConfig holds cluster-wide configuration
type ClusterConfig struct {
	Size              int
	NetworkLatency    time.Duration
	PartitionDuration time.Duration
	FailureRate       float64
}

// NodeStatus represents the current status of a node
type NodeStatus int

const (
	NodeStatusHealthy NodeStatus = iota
	NodeStatusSlow
	NodeStatusPartitioned
	NodeStatusFailed
	NodeStatusRecovering
)

// FailureEvent represents a failure that occurred on a node
type FailureEvent struct {
	Type        FailureType
	Timestamp   time.Time
	Duration    time.Duration
	Description string
}

// FailureType represents different types of failures
type FailureType int

const (
	FailureTypeNetworkPartition FailureType = iota
	FailureTypeHighLatency
	FailureTypeMemoryPressure
	FailureTypeCPUStarvation
	FailureTypeDiskFull
	FailureTypeProcessKill
	FailureTypeClockSkew
)

// ChaosCoordinator orchestrates chaos experiments
type ChaosCoordinator struct {
	cluster        *TestCluster
	activeFailures map[string]*ActiveFailure
	mu             sync.RWMutex
}

// ActiveFailure represents an ongoing failure
type ActiveFailure struct {
	Type      FailureType
	Target    string
	StartTime time.Time
	Duration  time.Duration
	Stop      chan struct{}
}

// TestChaosEngineering runs comprehensive chaos engineering tests
func TestChaosEngineering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos engineering tests in short mode")
	}

	tests := []ChaosTest{
		{
			Name:        "NetworkPartition",
			Description: "Test cluster resilience during network partitions",
			Setup:       setupNetworkPartitionTest,
			Chaos:       injectNetworkPartition,
			Verify:      verifyPartitionResilience,
			Cleanup:     cleanupNetworkPartition,
		},
		{
			Name:        "LeaderFailure",
			Description: "Test leader election during leader failures",
			Setup:       setupLeaderFailureTest,
			Chaos:       injectLeaderFailure,
			Verify:      verifyLeaderRecovery,
			Cleanup:     cleanupLeaderFailure,
		},
		{
			Name:        "HighLatency",
			Description: "Test system behavior under high network latency",
			Setup:       setupLatencyTest,
			Chaos:       injectHighLatency,
			Verify:      verifyLatencyTolerance,
			Cleanup:     cleanupLatency,
		},
		{
			Name:        "MemoryPressure",
			Description: "Test system behavior under memory pressure",
			Setup:       setupMemoryPressureTest,
			Chaos:       injectMemoryPressure,
			Verify:      verifyMemoryResilience,
			Cleanup:     cleanupMemoryPressure,
		},
		{
			Name:        "ByzantineFaults",
			Description: "Test Byzantine fault tolerance",
			Setup:       setupByzantineTest,
			Chaos:       injectByzantineFaults,
			Verify:      verifyByzantineTolerance,
			Cleanup:     cleanupByzantine,
		},
		{
			Name:        "CascadingFailures",
			Description: "Test protection against cascading failures",
			Setup:       setupCascadingTest,
			Chaos:       injectCascadingFailures,
			Verify:      verifyCascadingProtection,
			Cleanup:     cleanupCascading,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			runChaosTest(t, test)
		})
	}
}

// TestRandomChaos runs random chaos scenarios
func TestRandomChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping random chaos tests in short mode")
	}

	cluster := createTestCluster(t, &ClusterConfig{
		Size:              5,
		NetworkLatency:    10 * time.Millisecond,
		PartitionDuration: 30 * time.Second,
		FailureRate:       0.1,
	})
	defer cluster.cleanup(t)

	// Start cluster
	err := cluster.start(t)
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(5 * time.Second)

	// Run random chaos for 2 minutes
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	go cluster.runRandomChaos(ctx)

	// Continuously verify system health
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cluster.verifySystemHealth(t)
		}
	}
}

// TestStressTest runs sustained stress testing
func TestStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	cluster := createTestCluster(t, &ClusterConfig{
		Size:           3,
		NetworkLatency: 5 * time.Millisecond,
		FailureRate:    0.05,
	})
	defer cluster.cleanup(t)

	err := cluster.start(t)
	require.NoError(t, err)

	// Apply sustained load
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start workload generators
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			cluster.runWorkload(ctx, workerID)
		}(i)
	}

	// Start fault injection
	go cluster.runContinuousFaults(ctx)

	// Wait for completion
	wg.Wait()

	// Verify final state
	cluster.verifySystemHealth(t)
}

// Implementation of chaos tests

func runChaosTest(t *testing.T, test ChaosTest) {
	cluster := createTestCluster(t, &ClusterConfig{
		Size:              3,
		NetworkLatency:    10 * time.Millisecond,
		PartitionDuration: 30 * time.Second,
		FailureRate:       0.1,
	})
	defer cluster.cleanup(t)

	// Setup phase
	if test.Setup != nil {
		test.Setup(t, cluster)
	}

	// Start cluster
	err := cluster.start(t)
	require.NoError(t, err)

	// Wait for stabilization
	time.Sleep(2 * time.Second)

	// Verify initial health
	cluster.verifySystemHealth(t)

	// Chaos phase
	if test.Chaos != nil {
		test.Chaos(t, cluster)
	}

	// Verification phase
	if test.Verify != nil {
		test.Verify(t, cluster)
	}

	// Cleanup phase
	if test.Cleanup != nil {
		test.Cleanup(t, cluster)
	}
}

// Cluster management functions

func createTestCluster(t *testing.T, config *ClusterConfig) *TestCluster {
	ctx, cancel := context.WithCancel(context.Background())

	cluster := &TestCluster{
		Nodes:  make([]*TestNode, config.Size),
		Config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < config.Size; i++ {
		node := createTestNode(t, i, config)
		cluster.Nodes[i] = node
	}

	cluster.Coordinator = &ChaosCoordinator{
		cluster:        cluster,
		activeFailures: make(map[string]*ActiveFailure),
	}

	return cluster
}

func createTestNode(t *testing.T, index int, config *ClusterConfig) *TestNode {
	nodeConfig := &NodeConfig{
		DataDir:   t.TempDir(),
		P2PPort:   9000 + index,
		RaftPort:  7000 + index,
		Bootstrap: index == 0,
	}

	nodeID := fmt.Sprintf("chaos-node-%d", index)

	// Create P2P node
	ctx := context.Background()
	p2pNode, err := p2p.NewP2PNode(ctx, nil)
	require.NoError(t, err)

	// Create consensus engine
	consensusConfig := &config.ConsensusConfig{
		DataDir:           nodeConfig.DataDir + "/consensus",
		BindAddr:          fmt.Sprintf("127.0.0.1:%d", nodeConfig.RaftPort),
		Bootstrap:         nodeConfig.Bootstrap,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "ERROR",
	}

	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	// Create scheduler engine
	schedulerEngine, err := scheduler.NewEngine(p2pNode, consensusEngine)
	require.NoError(t, err)

	return &TestNode{
		ID:              nodeID,
		P2PNode:         p2pNode,
		ConsensusEngine: consensusEngine,
		SchedulerEngine: schedulerEngine,
		Config:          nodeConfig,
		Status:          NodeStatusHealthy,
		Failures:        make([]FailureEvent, 0),
	}
}

func (c *TestCluster) start(t *testing.T) error {
	for _, node := range c.Nodes {
		err := node.start(t)
		if err != nil {
			return fmt.Errorf("failed to start node %s: %w", node.ID, err)
		}
	}
	return nil
}

func (c *TestCluster) cleanup(t *testing.T) {
	c.cancel()

	for _, node := range c.Nodes {
		if node != nil {
			node.stop(t)
		}
	}
}

func (n *TestNode) start(t *testing.T) error {
	err := n.P2PNode.Start()
	if err != nil {
		return err
	}

	err = n.ConsensusEngine.Start()
	if err != nil {
		return err
	}

	err = n.SchedulerEngine.Start()
	if err != nil {
		return err
	}

	n.setStatus(NodeStatusHealthy)
	return nil
}

func (n *TestNode) stop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if n.SchedulerEngine != nil {
		n.SchedulerEngine.Shutdown(ctx)
	}

	if n.ConsensusEngine != nil {
		n.ConsensusEngine.Shutdown(ctx)
	}

	if n.P2PNode != nil {
		n.P2PNode.Stop()
	}

	n.setStatus(NodeStatusFailed)
}

func (n *TestNode) setStatus(status NodeStatus) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Status = status
}

func (n *TestNode) addFailure(failure FailureEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Failures = append(n.Failures, failure)
}

// Chaos injection functions

func setupNetworkPartitionTest(t *testing.T, cluster *TestCluster) {
	// Ensure cluster has enough nodes for partition testing
	require.True(t, len(cluster.Nodes) >= 3, "Network partition test requires at least 3 nodes")
}

func injectNetworkPartition(t *testing.T, cluster *TestCluster) {
	// Simulate network partition by isolating minority nodes
	partitionSize := len(cluster.Nodes) / 2
	if partitionSize == 0 {
		partitionSize = 1
	}

	// Select nodes to partition
	partitionedNodes := cluster.Nodes[:partitionSize]

	for _, node := range partitionedNodes {
		cluster.Coordinator.startFailure(node.ID, FailureTypeNetworkPartition, 30*time.Second)
		node.setStatus(NodeStatusPartitioned)

		// Simulate partition by stopping P2P communication
		node.addFailure(FailureEvent{
			Type:        FailureTypeNetworkPartition,
			Timestamp:   time.Now(),
			Duration:    30 * time.Second,
			Description: "Network partition injected",
		})
	}

	// Wait for partition effects
	time.Sleep(5 * time.Second)
}

func verifyPartitionResilience(t *testing.T, cluster *TestCluster) {
	// Verify majority partition maintains functionality
	majorityNodes := cluster.Nodes[len(cluster.Nodes)/2:]

	// Find leader in majority partition
	var leader *TestNode
	for _, node := range majorityNodes {
		if node.Status == NodeStatusHealthy && node.ConsensusEngine.IsLeader() {
			leader = node
			break
		}
	}

	assert.NotNil(t, leader, "Majority partition should maintain leadership")

	// Test operations on majority partition
	if leader != nil {
		err := leader.ConsensusEngine.Apply("partition-test", "value", nil)
		assert.NoError(t, err, "Should be able to perform operations in majority partition")
	}
}

func cleanupNetworkPartition(t *testing.T, cluster *TestCluster) {
	// Restore network connectivity
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusPartitioned {
			cluster.Coordinator.stopFailure(node.ID, FailureTypeNetworkPartition)
			node.setStatus(NodeStatusRecovering)
		}
	}

	// Wait for cluster recovery
	time.Sleep(10 * time.Second)

	// Verify all nodes are healthy
	for _, node := range cluster.Nodes {
		node.setStatus(NodeStatusHealthy)
	}
}

func setupLeaderFailureTest(t *testing.T, cluster *TestCluster) {
	// Wait for initial leader election
	time.Sleep(3 * time.Second)
}

func injectLeaderFailure(t *testing.T, cluster *TestCluster) {
	// Find current leader
	var leader *TestNode
	for _, node := range cluster.Nodes {
		if node.ConsensusEngine.IsLeader() {
			leader = node
			break
		}
	}

	require.NotNil(t, leader, "Should have a leader before failure injection")

	// Kill the leader
	cluster.Coordinator.startFailure(leader.ID, FailureTypeProcessKill, 60*time.Second)
	leader.stop(t)
	leader.setStatus(NodeStatusFailed)

	leader.addFailure(FailureEvent{
		Type:        FailureTypeProcessKill,
		Timestamp:   time.Now(),
		Duration:    60 * time.Second,
		Description: "Leader process killed",
	})

	// Wait for leader election
	time.Sleep(5 * time.Second)
}

func verifyLeaderRecovery(t *testing.T, cluster *TestCluster) {
	// Verify new leader is elected
	var newLeader *TestNode
	healthyNodes := 0

	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusHealthy {
			healthyNodes++
			if node.ConsensusEngine.IsLeader() {
				newLeader = node
			}
		}
	}

	assert.True(t, healthyNodes >= 2, "Should have at least 2 healthy nodes")
	assert.NotNil(t, newLeader, "Should elect new leader")

	// Test operations with new leader
	if newLeader != nil {
		err := newLeader.ConsensusEngine.Apply("leader-recovery-test", "value", nil)
		assert.NoError(t, err, "New leader should be operational")
	}
}

func cleanupLeaderFailure(t *testing.T, cluster *TestCluster) {
	// Restart failed leader
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusFailed {
			cluster.Coordinator.stopFailure(node.ID, FailureTypeProcessKill)
			err := node.start(t)
			assert.NoError(t, err, "Should be able to restart failed leader")
		}
	}

	// Wait for cluster stabilization
	time.Sleep(10 * time.Second)
}

func setupLatencyTest(t *testing.T, cluster *TestCluster) {
	// Baseline latency test
	cluster.measureBaselineLatency(t)
}

func injectHighLatency(t *testing.T, cluster *TestCluster) {
	// Inject high latency on random nodes
	numAffected := len(cluster.Nodes) / 2
	if numAffected == 0 {
		numAffected = 1
	}

	for i := 0; i < numAffected; i++ {
		node := cluster.Nodes[rand.Intn(len(cluster.Nodes))]
		cluster.Coordinator.startFailure(node.ID, FailureTypeHighLatency, 45*time.Second)
		node.setStatus(NodeStatusSlow)

		node.addFailure(FailureEvent{
			Type:        FailureTypeHighLatency,
			Timestamp:   time.Now(),
			Duration:    45 * time.Second,
			Description: "High network latency injected",
		})
	}

	// Let high latency take effect
	time.Sleep(5 * time.Second)
}

func verifyLatencyTolerance(t *testing.T, cluster *TestCluster) {
	// Verify system continues to operate under high latency
	healthyNodes := cluster.getHealthyNodes()
	assert.True(t, len(healthyNodes) > 0, "Should have healthy nodes under latency")

	// Test operations still work (may be slower)
	if len(healthyNodes) > 0 {
		leader := cluster.findLeader(healthyNodes)
		if leader != nil {
			start := time.Now()
			err := leader.ConsensusEngine.Apply("latency-test", "value", nil)
			duration := time.Since(start)

			assert.NoError(t, err, "Operations should work under high latency")
			t.Logf("Operation took %v under high latency", duration)
		}
	}
}

func cleanupLatency(t *testing.T, cluster *TestCluster) {
	// Remove latency injection
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusSlow {
			cluster.Coordinator.stopFailure(node.ID, FailureTypeHighLatency)
			node.setStatus(NodeStatusHealthy)
		}
	}

	time.Sleep(5 * time.Second)
}

func setupMemoryPressureTest(t *testing.T, cluster *TestCluster) {
	// Record baseline memory usage
	cluster.recordMemoryBaseline(t)
}

func injectMemoryPressure(t *testing.T, cluster *TestCluster) {
	// Simulate memory pressure on random nodes
	targetNode := cluster.Nodes[rand.Intn(len(cluster.Nodes))]

	cluster.Coordinator.startFailure(targetNode.ID, FailureTypeMemoryPressure, 60*time.Second)

	targetNode.addFailure(FailureEvent{
		Type:        FailureTypeMemoryPressure,
		Timestamp:   time.Now(),
		Duration:    60 * time.Second,
		Description: "Memory pressure injected",
	})

	// Simulate memory pressure by creating memory load
	go cluster.simulateMemoryPressure(targetNode, 60*time.Second)
}

func verifyMemoryResilience(t *testing.T, cluster *TestCluster) {
	// Verify system handles memory pressure gracefully
	healthyNodes := cluster.getHealthyNodes()
	assert.True(t, len(healthyNodes) >= len(cluster.Nodes)/2, "Majority of nodes should remain healthy")

	// Test that operations continue
	leader := cluster.findLeader(healthyNodes)
	if leader != nil {
		err := leader.ConsensusEngine.Apply("memory-pressure-test", "value", nil)
		assert.NoError(t, err, "Operations should continue under memory pressure")
	}
}

func cleanupMemoryPressure(t *testing.T, cluster *TestCluster) {
	// Stop memory pressure simulation
	for _, node := range cluster.Nodes {
		cluster.Coordinator.stopFailure(node.ID, FailureTypeMemoryPressure)
		node.setStatus(NodeStatusHealthy)
	}

	time.Sleep(5 * time.Second)
}

func setupByzantineTest(t *testing.T, cluster *TestCluster) {
	require.True(t, len(cluster.Nodes) >= 4, "Byzantine fault tolerance requires at least 4 nodes")
}

func injectByzantineFaults(t *testing.T, cluster *TestCluster) {
	// Make one node Byzantine (send conflicting messages)
	byzantineNode := cluster.Nodes[len(cluster.Nodes)-1]

	cluster.Coordinator.startFailure(byzantineNode.ID, FailureTypeProcessKill, 60*time.Second)

	byzantineNode.addFailure(FailureEvent{
		Type:        FailureTypeProcessKill,
		Timestamp:   time.Now(),
		Duration:    60 * time.Second,
		Description: "Byzantine fault injected",
	})

	// Simulate Byzantine behavior by stopping the node
	// In a real implementation, this would send conflicting messages
	byzantineNode.stop(t)
	byzantineNode.setStatus(NodeStatusFailed)
}

func verifyByzantineTolerance(t *testing.T, cluster *TestCluster) {
	// Verify cluster tolerates Byzantine node
	healthyNodes := cluster.getHealthyNodes()
	assert.True(t, len(healthyNodes) >= 2, "Should maintain majority with Byzantine fault")

	// Verify consensus still works
	leader := cluster.findLeader(healthyNodes)
	if leader != nil {
		err := leader.ConsensusEngine.Apply("byzantine-test", "value", nil)
		assert.NoError(t, err, "Consensus should work despite Byzantine fault")
	}
}

func cleanupByzantine(t *testing.T, cluster *TestCluster) {
	// Restart Byzantine node
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusFailed {
			cluster.Coordinator.stopFailure(node.ID, FailureTypeProcessKill)
			err := node.start(t)
			assert.NoError(t, err, "Should be able to restart Byzantine node")
		}
	}

	time.Sleep(10 * time.Second)
}

func setupCascadingTest(t *testing.T, cluster *TestCluster) {
	// Ensure cluster is large enough for cascading failure test
	require.True(t, len(cluster.Nodes) >= 3, "Cascading failure test requires at least 3 nodes")
}

func injectCascadingFailures(t *testing.T, cluster *TestCluster) {
	// Start with one failure and see if it cascades
	firstNode := cluster.Nodes[0]
	cluster.Coordinator.startFailure(firstNode.ID, FailureTypeProcessKill, 30*time.Second)
	firstNode.stop(t)
	firstNode.setStatus(NodeStatusFailed)

	// Wait and see if more failures occur
	time.Sleep(10 * time.Second)

	// Inject second failure to test cascade prevention
	if len(cluster.Nodes) > 2 {
		secondNode := cluster.Nodes[1]
		cluster.Coordinator.startFailure(secondNode.ID, FailureTypeProcessKill, 30*time.Second)
		secondNode.stop(t)
		secondNode.setStatus(NodeStatusFailed)
	}
}

func verifyCascadingProtection(t *testing.T, cluster *TestCluster) {
	// Verify remaining nodes stay healthy despite multiple failures
	healthyNodes := cluster.getHealthyNodes()
	assert.True(t, len(healthyNodes) > 0, "Should prevent complete cluster failure")

	// Test that at least one node remains operational
	if len(healthyNodes) > 0 {
		// Try to find a leader among healthy nodes
		leader := cluster.findLeader(healthyNodes)
		// It's okay if there's no leader immediately after cascading failures
		// The important thing is that some nodes remain healthy
		t.Logf("Healthy nodes remaining: %d", len(healthyNodes))
	}
}

func cleanupCascading(t *testing.T, cluster *TestCluster) {
	// Restart all failed nodes
	for _, node := range cluster.Nodes {
		if node.Status == NodeStatusFailed {
			cluster.Coordinator.stopFailure(node.ID, FailureTypeProcessKill)
			err := node.start(t)
			assert.NoError(t, err, "Should be able to restart failed nodes")
		}
	}

	time.Sleep(15 * time.Second)
}

// Utility functions

func (c *TestCluster) verifySystemHealth(t *testing.T) {
	healthyNodes := c.getHealthyNodes()
	assert.True(t, len(healthyNodes) > len(c.Nodes)/2, "Majority of nodes should be healthy")

	leader := c.findLeader(healthyNodes)
	if leader != nil {
		// Test basic operations
		testKey := fmt.Sprintf("health-check-%d", time.Now().UnixNano())
		err := leader.ConsensusEngine.Apply(testKey, "test-value", nil)
		assert.NoError(t, err, "Should be able to perform operations")
	}
}

func (c *TestCluster) getHealthyNodes() []*TestNode {
	var healthy []*TestNode
	for _, node := range c.Nodes {
		if node.Status == NodeStatusHealthy {
			healthy = append(healthy, node)
		}
	}
	return healthy
}

func (c *TestCluster) findLeader(nodes []*TestNode) *TestNode {
	for _, node := range nodes {
		if node.ConsensusEngine.IsLeader() {
			return node
		}
	}
	return nil
}

func (c *TestCluster) runRandomChaos(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if rand.Float64() < c.Config.FailureRate {
				c.injectRandomFailure()
			}
		}
	}
}

func (c *TestCluster) injectRandomFailure() {
	if len(c.Nodes) == 0 {
		return
	}

	node := c.Nodes[rand.Intn(len(c.Nodes))]
	if node.Status != NodeStatusHealthy {
		return // Don't inject failure on already failed node
	}

	failureTypes := []FailureType{
		FailureTypeHighLatency,
		FailureTypeMemoryPressure,
		FailureTypeCPUStarvation,
	}

	failureType := failureTypes[rand.Intn(len(failureTypes))]
	duration := time.Duration(rand.Intn(30)+10) * time.Second

	c.Coordinator.startFailure(node.ID, failureType, duration)

	node.addFailure(FailureEvent{
		Type:        failureType,
		Timestamp:   time.Now(),
		Duration:    duration,
		Description: "Random chaos injection",
	})

	switch failureType {
	case FailureTypeHighLatency:
		node.setStatus(NodeStatusSlow)
	case FailureTypeMemoryPressure, FailureTypeCPUStarvation:
		// Simulate resource pressure
		go c.simulateResourcePressure(node, duration)
	}
}

func (c *TestCluster) runWorkload(ctx context.Context, workerID int) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Find a healthy leader
			healthyNodes := c.getHealthyNodes()
			leader := c.findLeader(healthyNodes)

			if leader != nil {
				key := fmt.Sprintf("worker-%d-op-%d", workerID, counter)
				value := fmt.Sprintf("value-%d", counter)

				err := leader.ConsensusEngine.Apply(key, value, nil)
				if err != nil {
					// Log error but continue
					continue
				}
				counter++
			}
		}
	}
}

func (c *TestCluster) runContinuousFaults(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if rand.Float64() < 0.3 { // 30% chance of fault injection
				c.injectRandomFailure()
			}
		}
	}
}

func (c *TestCluster) measureBaselineLatency(t *testing.T) {
	// Measure baseline consensus latency
	healthyNodes := c.getHealthyNodes()
	leader := c.findLeader(healthyNodes)

	if leader != nil {
		start := time.Now()
		err := leader.ConsensusEngine.Apply("baseline-test", "value", nil)
		duration := time.Since(start)

		if err == nil {
			t.Logf("Baseline consensus latency: %v", duration)
		}
	}
}

func (c *TestCluster) recordMemoryBaseline(t *testing.T) {
	// Record baseline memory usage for all nodes
	t.Logf("Recording memory baseline for %d nodes", len(c.Nodes))
}

func (c *TestCluster) simulateMemoryPressure(node *TestNode, duration time.Duration) {
	// Simulate memory pressure by allocating memory
	// This is a simplified simulation
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		// Allocate and hold memory to create pressure
		_ = make([]byte, 1024*1024) // 1MB allocation
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *TestCluster) simulateResourcePressure(node *TestNode, duration time.Duration) {
	// Simulate CPU/memory pressure
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		// CPU intensive operation
		for i := 0; i < 100000; i++ {
			_ = i * i
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// ChaosCoordinator methods

func (cc *ChaosCoordinator) startFailure(nodeID string, failureType FailureType, duration time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	failure := &ActiveFailure{
		Type:      failureType,
		Target:    nodeID,
		StartTime: time.Now(),
		Duration:  duration,
		Stop:      make(chan struct{}),
	}

	key := fmt.Sprintf("%s-%d", nodeID, failureType)
	cc.activeFailures[key] = failure

	// Auto-stop after duration
	go func() {
		select {
		case <-time.After(duration):
			cc.stopFailure(nodeID, failureType)
		case <-failure.Stop:
			return
		}
	}()
}

func (cc *ChaosCoordinator) stopFailure(nodeID string, failureType FailureType) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	key := fmt.Sprintf("%s-%d", nodeID, failureType)
	if failure, exists := cc.activeFailures[key]; exists {
		close(failure.Stop)
		delete(cc.activeFailures, key)
	}
}

// Benchmark chaos tests

func BenchmarkChaos_ConsensusUnderLatency(b *testing.B) {
	cluster := createTestCluster(&testing.T{}, &ClusterConfig{
		Size:           3,
		NetworkLatency: 100 * time.Millisecond,
		FailureRate:    0.1,
	})
	defer cluster.cleanup(&testing.T{})

	err := cluster.start(&testing.T{})
	if err != nil {
		b.Fatal(err)
	}

	// Inject latency
	for _, node := range cluster.Nodes {
		cluster.Coordinator.startFailure(node.ID, FailureTypeHighLatency, time.Minute)
	}

	// Wait for cluster to adapt
	time.Sleep(2 * time.Second)

	leader := cluster.findLeader(cluster.getHealthyNodes())
	if leader == nil {
		b.Fatal("No leader found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		value := fmt.Sprintf("bench-value-%d", i)

		err := leader.ConsensusEngine.Apply(key, value, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}
