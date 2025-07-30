package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClusterSetup tests the basic cluster setup and functionality
func TestClusterSetup(t *testing.T) {
	// Create temporary directories for test nodes
	testDir := t.TempDir()
	
	// Create test configurations for 3 nodes
	nodes := make([]*TestNode, 3)
	
	for i := 0; i < 3; i++ {
		nodeDir := filepath.Join(testDir, fmt.Sprintf("node%d", i))
		require.NoError(t, os.MkdirAll(nodeDir, 0755))
		
		cfg := createTestConfig(nodeDir, i)
		node, err := NewTestNode(cfg)
		require.NoError(t, err)
		
		nodes[i] = node
	}
	
	// Start all nodes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	for i, node := range nodes {
		require.NoError(t, node.Start(ctx), "Failed to start node %d", i)
	}
	
	// Wait for nodes to connect
	time.Sleep(5 * time.Second)
	
	// Bootstrap the first node
	require.NoError(t, nodes[0].Bootstrap())
	
	// Join other nodes to the cluster
	for i := 1; i < len(nodes); i++ {
		require.NoError(t, nodes[i].JoinCluster(nodes[0].GetAddress()))
	}
	
	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)
	
	// Verify cluster state
	t.Run("VerifyClusterState", func(t *testing.T) {
		verifyClusterState(t, nodes)
	})
	
	// Test consensus
	t.Run("TestConsensus", func(t *testing.T) {
		testConsensus(t, nodes)
	})
	
	// Test model distribution
	t.Run("TestModelDistribution", func(t *testing.T) {
		testModelDistribution(t, nodes)
	})
	
	// Test fault tolerance
	t.Run("TestFaultTolerance", func(t *testing.T) {
		testFaultTolerance(t, nodes)
	})
	
	// Cleanup
	for _, node := range nodes {
		node.Shutdown()
	}
}

// TestNode represents a test node
type TestNode struct {
	config    *config.Config
	p2p       *p2p.Node
	consensus *consensus.Engine
	scheduler *scheduler.Engine
	api       *api.Server
}

// NewTestNode creates a new test node
func NewTestNode(cfg *config.Config) (*TestNode, error) {
	ctx := context.Background()
	
	// Create P2P node
	p2pNode, err := p2p.NewNode(ctx, &cfg.P2P)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P node: %w", err)
	}
	
	// Create consensus engine
	consensusEngine, err := consensus.NewEngine(&cfg.Consensus, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create consensus engine: %w", err)
	}
	
	// Create scheduler
	schedulerEngine, err := scheduler.NewEngine(&cfg.Scheduler, p2pNode, consensusEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}
	
	// Create API server
	apiServer, err := api.NewServer(&cfg.API, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}
	
	return &TestNode{
		config:    cfg,
		p2p:       p2pNode,
		consensus: consensusEngine,
		scheduler: schedulerEngine,
		api:       apiServer,
	}, nil
}

// Start starts the test node
func (tn *TestNode) Start(ctx context.Context) error {
	if err := tn.p2p.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}
	
	if err := tn.consensus.Start(); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}
	
	if err := tn.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	if err := tn.api.Start(); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	
	return nil
}

// Bootstrap bootstraps the node as cluster leader
func (tn *TestNode) Bootstrap() error {
	// TODO: Implement bootstrap functionality
	return nil
}

// JoinCluster joins the node to an existing cluster
func (tn *TestNode) JoinCluster(leaderAddr string) error {
	// TODO: Implement cluster join functionality
	return nil
}

// GetAddress returns the node's address
func (tn *TestNode) GetAddress() string {
	return tn.config.Consensus.BindAddr
}

// Shutdown shuts down the test node
func (tn *TestNode) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if tn.api != nil {
		tn.api.Shutdown(ctx)
	}
	
	if tn.scheduler != nil {
		tn.scheduler.Shutdown(ctx)
	}
	
	if tn.consensus != nil {
		tn.consensus.Shutdown(ctx)
	}
	
	if tn.p2p != nil {
		tn.p2p.Shutdown(ctx)
	}
}

// createTestConfig creates a test configuration for a node
func createTestConfig(dataDir string, nodeIndex int) *config.Config {
	cfg := config.DefaultConfig()
	
	// Unique ports for each node
	basePort := 10000 + nodeIndex*100
	
	cfg.API.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+34)
	cfg.P2P.Listen = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", basePort+1)
	cfg.Consensus.BindAddr = fmt.Sprintf("127.0.0.1:%d", basePort+70)
	cfg.Web.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+80)
	cfg.Metrics.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+90)
	
	// Data directories
	cfg.Storage.DataDir = filepath.Join(dataDir, "data")
	cfg.Storage.ModelDir = filepath.Join(dataDir, "models")
	cfg.Storage.CacheDir = filepath.Join(dataDir, "cache")
	cfg.Consensus.DataDir = filepath.Join(dataDir, "consensus")
	
	// Disable TLS for testing
	cfg.Security.TLS.Enabled = false
	cfg.Security.Auth.Enabled = false
	cfg.Web.TLS.Enabled = false
	
	// Test-specific settings
	cfg.Consensus.Bootstrap = (nodeIndex == 0)
	cfg.Scheduler.WorkerCount = 2
	cfg.P2P.ConnMgrLow = 1
	cfg.P2P.ConnMgrHigh = 10
	
	return cfg
}

// verifyClusterState verifies the cluster state
func verifyClusterState(t *testing.T, nodes []*TestNode) {
	// Check that all nodes are connected
	for i, node := range nodes {
		peers := node.p2p.ConnectedPeers()
		assert.GreaterOrEqual(t, len(peers), 1, "Node %d should have at least 1 peer", i)
	}
	
	// Check that there's exactly one leader
	leaderCount := 0
	for i, node := range nodes {
		if node.consensus.IsLeader() {
			leaderCount++
			t.Logf("Node %d is the leader", i)
		}
	}
	assert.Equal(t, 1, leaderCount, "There should be exactly one leader")
}

// testConsensus tests consensus functionality
func testConsensus(t *testing.T, nodes []*TestNode) {
	// Find the leader
	var leader *TestNode
	for _, node := range nodes {
		if node.consensus.IsLeader() {
			leader = node
			break
		}
	}
	
	require.NotNil(t, leader, "No leader found")
	
	// Apply some changes through the leader
	testKey := "test_key"
	testValue := "test_value"
	
	err := leader.consensus.Apply(testKey, testValue, nil)
	require.NoError(t, err, "Failed to apply change through leader")
	
	// Wait for replication
	time.Sleep(2 * time.Second)
	
	// Verify all nodes have the same state
	for i, node := range nodes {
		value, exists := node.consensus.Get(testKey)
		assert.True(t, exists, "Node %d should have the key", i)
		assert.Equal(t, testValue, value, "Node %d should have the correct value", i)
	}
}

// testModelDistribution tests model distribution functionality
func testModelDistribution(t *testing.T, nodes []*TestNode) {
	// TODO: Implement model distribution tests
	t.Skip("Model distribution tests not implemented yet")
}

// testFaultTolerance tests fault tolerance
func testFaultTolerance(t *testing.T, nodes []*TestNode) {
	if len(nodes) < 3 {
		t.Skip("Need at least 3 nodes for fault tolerance test")
	}
	
	// Find a non-leader node to shutdown
	var nodeToShutdown *TestNode
	var nodeIndex int
	for i, node := range nodes {
		if !node.consensus.IsLeader() {
			nodeToShutdown = node
			nodeIndex = i
			break
		}
	}
	
	require.NotNil(t, nodeToShutdown, "No non-leader node found")
	
	// Shutdown the node
	t.Logf("Shutting down node %d", nodeIndex)
	nodeToShutdown.Shutdown()
	
	// Wait for cluster to detect the failure
	time.Sleep(5 * time.Second)
	
	// Verify cluster is still operational
	var leader *TestNode
	for i, node := range nodes {
		if i == nodeIndex {
			continue // Skip the shutdown node
		}
		if node.consensus.IsLeader() {
			leader = node
			break
		}
	}
	
	require.NotNil(t, leader, "No leader found after node shutdown")
	
	// Test that we can still apply changes
	testKey := "fault_tolerance_test"
	testValue := "still_working"
	
	err := leader.consensus.Apply(testKey, testValue, nil)
	require.NoError(t, err, "Failed to apply change after node shutdown")
	
	// Verify remaining nodes have the change
	for i, node := range nodes {
		if i == nodeIndex {
			continue // Skip the shutdown node
		}
		
		value, exists := node.consensus.Get(testKey)
		assert.True(t, exists, "Node %d should have the key after fault", i)
		assert.Equal(t, testValue, value, "Node %d should have the correct value after fault", i)
	}
}

// Benchmark tests

func BenchmarkClusterThroughput(b *testing.B) {
	// Create a test cluster
	testDir := b.TempDir()
	nodes := make([]*TestNode, 3)
	
	for i := 0; i < 3; i++ {
		nodeDir := filepath.Join(testDir, fmt.Sprintf("node%d", i))
		require.NoError(b, os.MkdirAll(nodeDir, 0755))
		
		cfg := createTestConfig(nodeDir, i)
		node, err := NewTestNode(cfg)
		require.NoError(b, err)
		
		nodes[i] = node
	}
	
	// Start all nodes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	for i, node := range nodes {
		require.NoError(b, node.Start(ctx), "Failed to start node %d", i)
	}
	
	// Wait for cluster to stabilize
	time.Sleep(5 * time.Second)
	
	// Find the leader
	var leader *TestNode
	for _, node := range nodes {
		if node.consensus.IsLeader() {
			leader = node
			break
		}
	}
	
	require.NotNil(b, leader, "No leader found")
	
	b.ResetTimer()
	
	// Benchmark consensus operations
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			value := fmt.Sprintf("bench_value_%d", i)
			
			err := leader.consensus.Apply(key, value, nil)
			if err != nil {
				b.Errorf("Failed to apply change: %v", err)
			}
			
			i++
		}
	})
	
	// Cleanup
	for _, node := range nodes {
		node.Shutdown()
	}
}

func BenchmarkModelDistribution(b *testing.B) {
	// TODO: Implement model distribution benchmarks
	b.Skip("Model distribution benchmarks not implemented yet")
}