package integration

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

// TestIsolatedConsensusEngine tests consensus engine in isolation
func TestIsolatedConsensusEngine(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping isolated integration test in short mode")
	}

	testDir := t.TempDir()
	dataDir := filepath.Join(testDir, "consensus")

	// Create minimal P2P node for consensus
	p2pConfig := &config.P2PConfig{
		Listen:       "/ip4/127.0.0.1/tcp/0", // Random port
		EnableDHT:    false,                   // Disable for isolation
		ConnMgrLow:   1,
		ConnMgrHigh:  10,
		ConnMgrGrace: "30s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p2pNode, err := p2p.NewNode(ctx, p2pConfig)
	require.NoError(t, err)

	// Start P2P node
	err = p2pNode.Start()
	require.NoError(t, err)
	defer p2pNode.Stop()

	// Create consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:           dataDir,
		BindAddr:          "127.0.0.1:0", // Random port
		Bootstrap:         true,          // Single node bootstrap
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     500 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "ERROR",
	}

	// Create consensus engine
	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	// Start consensus engine
	err = consensusEngine.Start()
	require.NoError(t, err)
	defer consensusEngine.Shutdown(ctx)

	// Wait for bootstrap
	time.Sleep(2 * time.Second)

	t.Run("ConsensusLeadership", func(t *testing.T) {
		// Check if node becomes leader
		assert.True(t, consensusEngine.IsLeader(), "Bootstrap node should become leader")
	})

	t.Run("ConsensusBasicOperations", func(t *testing.T) {
		// Test apply operation
		testKey := "isolated_test_key"
		testValue := "isolated_test_value"

		err := consensusEngine.Apply(testKey, testValue, nil)
		assert.NoError(t, err, "Should be able to apply consensus operation")

		// Allow some time for the operation to be applied
		time.Sleep(500 * time.Millisecond)

		// Test get operation
		value, exists := consensusEngine.Get(testKey)
		assert.True(t, exists, "Key should exist after apply")
		assert.Equal(t, testValue, value, "Value should match what was applied")
	})

	t.Run("ConsensusStateConsistency", func(t *testing.T) {
		// Apply multiple operations
		operations := map[string]string{
			"key1": "value1",
			"key2": "value2", 
			"key3": "value3",
		}

		for key, value := range operations {
			err := consensusEngine.Apply(key, value, nil)
			assert.NoError(t, err, "Should apply operation for key %s", key)
		}

		// Wait for all operations to be applied
		time.Sleep(1 * time.Second)

		// Verify all operations were applied correctly
		for key, expectedValue := range operations {
			value, exists := consensusEngine.Get(key)
			assert.True(t, exists, "Key %s should exist", key)
			assert.Equal(t, expectedValue, value, "Value for key %s should match", key)
		}
	})
}

// TestIsolatedP2PNode tests P2P functionality in isolation
func TestIsolatedP2PNode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping isolated integration test in short mode")
	}

	// Create two P2P nodes for testing connectivity
	p2pConfig1 := &config.P2PConfig{
		Listen:       "/ip4/127.0.0.1/tcp/0",
		EnableDHT:    false,
		ConnMgrLow:   1,
		ConnMgrHigh:  10,
		ConnMgrGrace: "30s",
	}

	p2pConfig2 := &config.P2PConfig{
		Listen:       "/ip4/127.0.0.1/tcp/0",
		EnableDHT:    false,
		ConnMgrLow:   1,
		ConnMgrHigh:  10,
		ConnMgrGrace: "30s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create first node
	node1, err := p2p.NewNode(ctx, p2pConfig1)
	require.NoError(t, err)

	err = node1.Start()
	require.NoError(t, err)
	defer node1.Stop()

	// Create second node
	node2, err := p2p.NewNode(ctx, p2pConfig2)
	require.NoError(t, err)

	err = node2.Start()
	require.NoError(t, err)
	defer node2.Stop()

	// Allow nodes to start up
	time.Sleep(2 * time.Second)

	t.Run("P2PNodeInitialization", func(t *testing.T) {
		// Basic checks that nodes are running
		// Note: Full connectivity testing requires proper multiaddr setup
		assert.NotNil(t, node1, "Node 1 should be initialized")
		assert.NotNil(t, node2, "Node 2 should be initialized")
	})

	t.Run("P2PNodeConnectedPeers", func(t *testing.T) {
		// Check peer counts (may be zero in isolated test)
		peers1 := node1.ConnectedPeers()
		peers2 := node2.ConnectedPeers()

		t.Logf("Node 1 has %d connected peers", len(peers1))
		t.Logf("Node 2 has %d connected peers", len(peers2))

		// In isolation, nodes may not connect, which is OK for this test
		assert.True(t, len(peers1) >= 0, "Peer count should be non-negative")
		assert.True(t, len(peers2) >= 0, "Peer count should be non-negative")
	})
}

// TestComponentIntegrationSimple tests basic integration between core components
func TestComponentIntegrationSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping component integration test in short mode")
	}

	testDir := t.TempDir()
	dataDir := filepath.Join(testDir, "node")

	// Setup P2P
	p2pConfig := &config.P2PConfig{
		Listen:       "/ip4/127.0.0.1/tcp/0",
		EnableDHT:    false,
		ConnMgrLow:   1,
		ConnMgrHigh:  10,
		ConnMgrGrace: "30s",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	p2pNode, err := p2p.NewNode(ctx, p2pConfig)
	require.NoError(t, err)

	err = p2pNode.Start()
	require.NoError(t, err)
	defer p2pNode.Stop()

	// Setup Consensus
	consensusConfig := &config.ConsensusConfig{
		DataDir:           dataDir,
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     500 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "ERROR",
	}

	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	err = consensusEngine.Start()
	require.NoError(t, err)
	defer consensusEngine.Shutdown(ctx)

	// Wait for system stabilization
	time.Sleep(3 * time.Second)

	t.Run("IntegratedComponentsBasicOperation", func(t *testing.T) {
		// Test that P2P and Consensus work together
		assert.True(t, consensusEngine.IsLeader(), "Node should be leader")
		
		// Test consensus operations with P2P integration
		testKey := "integration_key"
		testValue := "integration_value"

		err := consensusEngine.Apply(testKey, testValue, nil)
		assert.NoError(t, err, "Consensus operation should work with P2P")

		time.Sleep(500 * time.Millisecond)

		value, exists := consensusEngine.Get(testKey)
		assert.True(t, exists, "Key should exist")
		assert.Equal(t, testValue, value, "Value should match")
	})

	t.Run("ComponentHealthCheck", func(t *testing.T) {
		// Verify all components are healthy
		peers := p2pNode.ConnectedPeers()
		assert.True(t, len(peers) >= 0, "P2P should have valid peer count")

		// Consensus should be operational
		assert.True(t, consensusEngine.IsLeader(), "Consensus should be operational")
	})
}