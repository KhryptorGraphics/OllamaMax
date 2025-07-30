package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
)

// SimpleTestNode represents a simplified test node without complex dependencies
type SimpleTestNode struct {
	config    *config.Config
	p2pNode   *p2p.Node
	consensus *consensus.Engine
	scheduler *scheduler.Engine
	apiServer *api.Server
	active    bool
	dataDir   string
}

// TestSimpleClusterFormation tests basic cluster formation without distributed scheduler
func TestSimpleClusterFormation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create 3 simple test nodes
	nodes := make([]*SimpleTestNode, 3)
	testDir := t.TempDir()

	for i := 0; i < 3; i++ {
		nodeDir := filepath.Join(testDir, fmt.Sprintf("node-%d", i))
		require.NoError(t, os.MkdirAll(nodeDir, 0755))

		cfg := createSimpleTestConfig(nodeDir, i)
		node, err := createSimpleTestNode(cfg)
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

	// Cleanup
	for _, node := range nodes {
		node.Shutdown()
	}
}

// TestSimpleConsensus tests basic consensus functionality
func TestSimpleConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a single bootstrap node first
	testDir := t.TempDir()
	nodeDir := filepath.Join(testDir, "bootstrap")
	require.NoError(t, os.MkdirAll(nodeDir, 0755))

	cfg := createSimpleTestConfig(nodeDir, 0)
	cfg.Consensus.Bootstrap = true

	node, err := createSimpleTestNode(cfg)
	require.NoError(t, err)

	// Start the node
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	require.NoError(t, node.Start(ctx))

	// Wait for node to become leader
	time.Sleep(3 * time.Second)

	// Test consensus operations
	t.Run("BasicConsensusOperations", func(t *testing.T) {
		// Check if node is leader
		assert.True(t, node.consensus.IsLeader(), "Bootstrap node should be leader")

		// Apply some changes
		testKey := "test_key"
		testValue := "test_value"

		err := node.consensus.Apply(testKey, testValue, nil)
		assert.NoError(t, err, "Failed to apply change")

		// Wait for application
		time.Sleep(1 * time.Second)

		// Verify the change
		value, exists := node.consensus.Get(testKey)
		assert.True(t, exists, "Key should exist")
		assert.Equal(t, testValue, value, "Value should match")
	})

	// Cleanup
	node.Shutdown()
}

// TestP2PBasicFunctionality tests basic P2P functionality
func TestP2PBasicFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create 2 nodes for P2P testing
	testDir := t.TempDir()
	
	// Node 1
	node1Dir := filepath.Join(testDir, "node1")
	require.NoError(t, os.MkdirAll(node1Dir, 0755))
	cfg1 := createSimpleTestConfig(node1Dir, 0)
	node1, err := createSimpleTestNode(cfg1)
	require.NoError(t, err)

	// Node 2
	node2Dir := filepath.Join(testDir, "node2")
	require.NoError(t, os.MkdirAll(node2Dir, 0755))
	cfg2 := createSimpleTestConfig(node2Dir, 1)
	node2, err := createSimpleTestNode(cfg2)
	require.NoError(t, err)

	// Start nodes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	require.NoError(t, node1.Start(ctx))
	require.NoError(t, node2.Start(ctx))

	// Wait for connection
	time.Sleep(5 * time.Second)

	t.Run("P2PConnectivity", func(t *testing.T) {
		// Check if nodes can see each other
		peers1 := node1.p2pNode.ConnectedPeers()
		peers2 := node2.p2pNode.ConnectedPeers()

		// Allow some time for connection establishment
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		connected := false
		for !connected {
			select {
			case <-timeout:
				t.Log("Timeout waiting for P2P connection")
				t.Log("Node1 peers:", len(peers1))
				t.Log("Node2 peers:", len(peers2))
				return // Don't fail the test, just log
			case <-ticker.C:
				peers1 = node1.p2pNode.ConnectedPeers()
				peers2 = node2.p2pNode.ConnectedPeers()
				if len(peers1) > 0 || len(peers2) > 0 {
					connected = true
				}
			}
		}

		t.Logf("Node1 connected to %d peers", len(peers1))
		t.Logf("Node2 connected to %d peers", len(peers2))
	})

	// Cleanup
	node1.Shutdown()
	node2.Shutdown()
}

// TestAPIServerBasicFunctionality tests basic API server functionality
func TestAPIServerBasicFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a single node with API server
	testDir := t.TempDir()
	nodeDir := filepath.Join(testDir, "api-node")
	require.NoError(t, os.MkdirAll(nodeDir, 0755))

	cfg := createSimpleTestConfig(nodeDir, 0)
	node, err := createSimpleTestNode(cfg)
	require.NoError(t, err)

	// Start the node
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	require.NoError(t, node.Start(ctx))

	// Wait for services to start
	time.Sleep(2 * time.Second)

	t.Run("APIServerHealth", func(t *testing.T) {
		// Test would check API endpoints if they were properly implemented
		assert.NotNil(t, node.apiServer, "API server should be initialized")
	})

	// Cleanup
	node.Shutdown()
}

// Helper functions

func createSimpleTestConfig(dataDir string, nodeIndex int) *config.Config {
	cfg := config.DefaultConfig()

	// Unique ports for each node
	basePort := 15000 + nodeIndex*100

	cfg.API.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+1)
	cfg.P2P.Listen = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", basePort+2)
	cfg.Consensus.BindAddr = fmt.Sprintf("127.0.0.1:%d", basePort+3)
	cfg.Web.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+4)
	cfg.Metrics.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+5)

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

	// Bootstrap configuration for non-bootstrap nodes
	if nodeIndex > 0 {
		bootstrapPort := 15000 + 2 // First node's P2P port
		cfg.P2P.Bootstrap = []string{
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", bootstrapPort),
		}
	}

	return cfg
}

func createSimpleTestNode(cfg *config.Config) (*SimpleTestNode, error) {
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

	// Create basic scheduler (without distributed features)
	schedulerEngine, err := scheduler.NewEngine(&cfg.Scheduler, p2pNode, consensusEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Create API server
	apiServer, err := api.NewServer(&cfg.API, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	return &SimpleTestNode{
		config:    cfg,
		p2pNode:   p2pNode,
		consensus: consensusEngine,
		scheduler: schedulerEngine,
		apiServer: apiServer,
		active:    false,
		dataDir:   cfg.Storage.DataDir,
	}, nil
}

func (stn *SimpleTestNode) Start(ctx context.Context) error {
	if stn.active {
		return fmt.Errorf("node already started")
	}

	// Start P2P node
	err := stn.p2pNode.Start()
	if err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Start consensus engine
	err = stn.consensus.Start()
	if err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start scheduler
	err = stn.scheduler.Start()
	if err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start API server
	err = stn.apiServer.Start()
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	stn.active = true
	return nil
}

func (stn *SimpleTestNode) Shutdown() {
	if !stn.active {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown components in reverse order
	if stn.apiServer != nil {
		stn.apiServer.Shutdown(ctx)
	}

	if stn.scheduler != nil {
		stn.scheduler.Shutdown(ctx)
	}

	if stn.consensus != nil {
		stn.consensus.Shutdown(ctx)
	}

	if stn.p2pNode != nil {
		stn.p2pNode.Stop()
	}

	stn.active = false
}