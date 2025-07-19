package unit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// TestConsensusEngineCreation tests consensus engine creation
func TestConsensusEngineCreation(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "consensus-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create test P2P node
	p2pConfig := &config.NodeConfig{
		Listen:         "/ip4/127.0.0.1/tcp/0",
		BootstrapPeers: []string{},
		EnableDHT:      false,
		EnableMDNS:     false,
	}
	
	ctx := context.Background()
	p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Stop()
	
	// Start P2P node
	err = p2pNode.Start()
	require.NoError(t, err)
	
	// Create consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:            tempDir,
		BindAddr:           "127.0.0.1:0",
		Bootstrap:          true,
		LogLevel:           "INFO",
		HeartbeatTimeout:   1000 * time.Millisecond,
		ElectionTimeout:    1000 * time.Millisecond,
		CommitTimeout:      50 * time.Millisecond,
		MaxAppendEntries:   64,
		SnapshotInterval:   120 * time.Second,
		SnapshotThreshold:  8192,
	}
	
	// Create consensus engine
	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err, "Failed to create consensus engine")
	require.NotNil(t, engine, "Consensus engine should not be nil")
	
	// Test engine starts
	err = engine.Start()
	require.NoError(t, err, "Failed to start consensus engine")
	
	// Cleanup
	defer engine.Shutdown(ctx)
}

// TestConsensusEngineLeadership tests leadership functionality
func TestConsensusEngineLeadership(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-leadership-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create single-node cluster
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership election
	time.Sleep(2 * time.Second)
	
	// Test leadership status
	assert.True(t, engine.IsLeader(), "Node should become leader in single-node cluster")
	
	// Test leader address
	leader := engine.Leader()
	assert.NotEmpty(t, leader, "Should have leader address")
	
	// Test leadership changes channel
	select {
	case isLeader := <-engine.LeadershipChanges():
		assert.True(t, isLeader, "Should receive leadership true")
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive leadership change notification")
	}
}

// TestConsensusBasicOperations tests basic consensus operations
func TestConsensusBasicOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-ops-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(t, engine.IsLeader(), "Node should be leader")
	
	// Test Apply operation
	testKey := "test-key"
	testValue := "test-value"
	metadata := map[string]interface{}{"timestamp": time.Now().Unix()}
	
	err = engine.Apply(testKey, testValue, metadata)
	require.NoError(t, err, "Apply operation should succeed")
	
	// Test Get operation
	value, exists := engine.Get(testKey)
	require.True(t, exists, "Key should exist after apply")
	assert.Equal(t, testValue, value, "Retrieved value should match applied value")
	
	// Test GetAll operation
	allValues := engine.GetAll()
	assert.Contains(t, allValues, testKey, "GetAll should contain applied key")
	assert.Equal(t, testValue, allValues[testKey], "GetAll should return correct value")
}

// TestConsensusDeleteOperations tests delete operations
func TestConsensusDeleteOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-delete-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(t, engine.IsLeader(), "Node should be leader")
	
	// Apply a value first
	testKey := "delete-test-key"
	testValue := "delete-test-value"
	
	err = engine.Apply(testKey, testValue, nil)
	require.NoError(t, err)
	
	// Verify value exists
	_, exists := engine.Get(testKey)
	require.True(t, exists, "Key should exist before delete")
	
	// Delete the value
	err = engine.Delete(testKey)
	require.NoError(t, err, "Delete operation should succeed")
	
	// Verify value is deleted
	_, exists = engine.Get(testKey)
	assert.False(t, exists, "Key should not exist after delete")
	
	// Verify delete of non-existent key doesn't error
	err = engine.Delete("non-existent-key")
	require.NoError(t, err, "Delete of non-existent key should not error")
}

// TestConsensusMultipleOperations tests multiple concurrent operations
func TestConsensusMultipleOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-multi-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(t, engine.IsLeader(), "Node should be leader")
	
	// Apply multiple values
	numOperations := 10
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		metadata := map[string]interface{}{"index": i}
		
		err = engine.Apply(key, value, metadata)
		require.NoError(t, err, "Apply operation %d should succeed", i)
	}
	
	// Verify all values exist
	allValues := engine.GetAll()
	assert.Equal(t, numOperations, len(allValues), "Should have all applied values")
	
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("key-%d", i)
		expectedValue := fmt.Sprintf("value-%d", i)
		
		value, exists := engine.Get(key)
		require.True(t, exists, "Key %s should exist", key)
		assert.Equal(t, expectedValue, value, "Value for key %s should match", key)
	}
}

// TestConsensusConfiguration tests cluster configuration operations
func TestConsensusConfiguration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(t, engine.IsLeader(), "Node should be leader")
	
	// Test getting configuration
	config, err := engine.GetConfiguration()
	require.NoError(t, err, "Should get configuration successfully")
	require.NotNil(t, config, "Configuration should not be nil")
	
	// Should have at least one server (this node)
	assert.GreaterOrEqual(t, len(config.Servers), 1, "Should have at least one server")
	
	// Test adding a voter (simulate adding another node)
	testNodeID := "test-node-2"
	testAddress := "127.0.0.1:8001"
	
	err = engine.AddVoter(testNodeID, testAddress)
	require.NoError(t, err, "Should add voter successfully")
	
	// Get updated configuration
	newConfig, err := engine.GetConfiguration()
	require.NoError(t, err, "Should get updated configuration")
	
	// Should now have two servers
	assert.Equal(t, 2, len(newConfig.Servers), "Should have two servers after adding voter")
	
	// Test removing the server
	err = engine.RemoveServer(testNodeID)
	require.NoError(t, err, "Should remove server successfully")
	
	// Get final configuration
	finalConfig, err := engine.GetConfiguration()
	require.NoError(t, err, "Should get final configuration")
	
	// Should be back to one server
	assert.Equal(t, 1, len(finalConfig.Servers), "Should have one server after removal")
}

// TestConsensusStats tests statistics collection
func TestConsensusStats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-stats-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for initialization
	time.Sleep(1 * time.Second)
	
	// Test stats collection
	stats := engine.Stats()
	require.NotNil(t, stats, "Stats should not be nil")
	
	// Check for expected stats keys
	expectedKeys := []string{"state", "term", "last_log_index", "last_log_term", "commit_index"}
	for _, key := range expectedKeys {
		_, exists := stats[key]
		assert.True(t, exists, "Stats should contain key: %s", key)
	}
	
	// Apply some operations to generate more stats
	if engine.IsLeader() {
		for i := 0; i < 5; i++ {
			err = engine.Apply(fmt.Sprintf("stats-key-%d", i), fmt.Sprintf("stats-value-%d", i), nil)
			require.NoError(t, err)
		}
		
		// Get updated stats
		newStats := engine.Stats()
		assert.NotEqual(t, stats["commit_index"], newStats["commit_index"], 
			"Commit index should increase after operations")
	}
}

// TestConsensusFSMOperations tests FSM-specific operations
func TestConsensusFSMOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-fsm-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(t, engine.IsLeader(), "Node should be leader")
	
	// Test different operation types
	testCases := []struct {
		operation string
		key       string
		value     interface{}
	}{
		{"set", "string-key", "string-value"},
		{"set", "int-key", 42},
		{"set", "float-key", 3.14},
		{"set", "bool-key", true},
		{"set", "map-key", map[string]interface{}{"nested": "value"}},
		{"set", "slice-key", []string{"item1", "item2", "item3"}},
	}
	
	// Apply all test cases
	for _, tc := range testCases {
		err = engine.Apply(tc.key, tc.value, map[string]interface{}{"type": tc.operation})
		require.NoError(t, err, "Apply operation should succeed for %s", tc.key)
	}
	
	// Verify all values
	for _, tc := range testCases {
		value, exists := engine.Get(tc.key)
		require.True(t, exists, "Key %s should exist", tc.key)
		assert.Equal(t, tc.value, value, "Value for key %s should match", tc.key)
	}
	
	// Test update operations
	updateKey := "update-test"
	originalValue := "original"
	updatedValue := "updated"
	
	// Apply original value
	err = engine.Apply(updateKey, originalValue, nil)
	require.NoError(t, err)
	
	// Update value
	err = engine.Apply(updateKey, updatedValue, map[string]interface{}{"type": "update"})
	require.NoError(t, err)
	
	// Verify update
	value, exists := engine.Get(updateKey)
	require.True(t, exists)
	assert.Equal(t, updatedValue, value, "Value should be updated")
}

// TestConsensusErrorHandling tests error handling scenarios
func TestConsensusErrorHandling(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-error-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Test with invalid data directory permissions
	invalidDir := filepath.Join(tempDir, "invalid")
	err = os.MkdirAll(invalidDir, 0000) // No permissions
	require.NoError(t, err)
	defer os.Chmod(invalidDir, 0755) // Restore permissions for cleanup
	
	invalidEngine := createTestConsensusEngineWithDir(t, invalidDir, true)
	err = invalidEngine.Start()
	assert.Error(t, err, "Should fail to start with invalid permissions")
	
	// Test non-leader operations
	engine := createTestConsensusEngine(t, tempDir, false) // Not bootstrap
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Try operations as non-leader
	err = engine.Apply("test", "value", nil)
	assert.Error(t, err, "Non-leader should not be able to apply")
	
	err = engine.Delete("test")
	assert.Error(t, err, "Non-leader should not be able to delete")
	
	err = engine.AddVoter("test", "address")
	assert.Error(t, err, "Non-leader should not be able to add voter")
	
	err = engine.RemoveServer("test")
	assert.Error(t, err, "Non-leader should not be able to remove server")
}

// TestConsensusShutdown tests graceful shutdown
func TestConsensusShutdown(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "consensus-shutdown-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(t, tempDir, true)
	
	err = engine.Start()
	require.NoError(t, err)
	
	// Apply some operations
	if engine.IsLeader() {
		err = engine.Apply("shutdown-test", "value", nil)
		require.NoError(t, err)
	}
	
	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = engine.Shutdown(ctx)
	require.NoError(t, err, "Shutdown should succeed")
	
	// Test operations after shutdown should fail
	err = engine.Apply("after-shutdown", "value", nil)
	assert.Error(t, err, "Operations should fail after shutdown")
}

// TestConsensusMultiNodeCluster tests multi-node cluster operations
func TestConsensusMultiNodeCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-node test in short mode")
	}
	
	// Create three-node cluster
	nodes := createTestCluster(t, 3)
	defer func() {
		for _, node := range nodes {
			node.engine.Shutdown(context.Background())
			os.RemoveAll(node.dataDir)
		}
	}()
	
	// Start all nodes
	for i, node := range nodes {
		err := node.engine.Start()
		require.NoError(t, err, "Node %d should start successfully", i)
	}
	
	// Wait for cluster formation and leader election
	time.Sleep(5 * time.Second)
	
	// Find the leader
	var leader *testClusterNode
	leaderCount := 0
	for _, node := range nodes {
		if node.engine.IsLeader() {
			leader = node
			leaderCount++
		}
	}
	
	require.Equal(t, 1, leaderCount, "Should have exactly one leader")
	require.NotNil(t, leader, "Should have a leader")
	
	// Test operations on leader
	testKey := "cluster-test"
	testValue := "cluster-value"
	
	err := leader.engine.Apply(testKey, testValue, nil)
	require.NoError(t, err, "Leader should be able to apply operations")
	
	// Wait for replication
	time.Sleep(2 * time.Second)
	
	// Verify replication to all nodes
	for i, node := range nodes {
		value, exists := node.engine.Get(testKey)
		assert.True(t, exists, "Node %d should have replicated value", i)
		assert.Equal(t, testValue, value, "Node %d should have correct value", i)
	}
	
	// Test configuration management
	config, err := leader.engine.GetConfiguration()
	require.NoError(t, err)
	assert.Equal(t, 3, len(config.Servers), "Should have 3 servers in configuration")
}

// Helper functions

type testClusterNode struct {
	engine   *consensus.Engine
	p2pNode  *p2p.P2PNode
	dataDir  string
}

func createTestConsensusEngine(t *testing.T, dataDir string, bootstrap bool) *consensus.Engine {
	return createTestConsensusEngineWithDir(t, dataDir, bootstrap)
}

func createTestConsensusEngineWithDir(t *testing.T, dataDir string, bootstrap bool) *consensus.Engine {
	// Create test P2P node
	p2pConfig := &config.NodeConfig{
		Listen:         "/ip4/127.0.0.1/tcp/0",
		BootstrapPeers: []string{},
		EnableDHT:      false,
		EnableMDNS:     false,
	}
	
	ctx := context.Background()
	p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
	require.NoError(t, err)
	
	err = p2pNode.Start()
	require.NoError(t, err)
	
	// Create consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:            dataDir,
		BindAddr:           "127.0.0.1:0",
		Bootstrap:          bootstrap,
		LogLevel:           "ERROR", // Reduce log noise in tests
		HeartbeatTimeout:   500 * time.Millisecond,
		ElectionTimeout:    500 * time.Millisecond,
		CommitTimeout:      50 * time.Millisecond,
		MaxAppendEntries:   64,
		SnapshotInterval:   120 * time.Second,
		SnapshotThreshold:  8192,
	}
	
	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)
	
	return engine
}

func createTestCluster(t *testing.T, nodeCount int) []*testClusterNode {
	nodes := make([]*testClusterNode, nodeCount)
	
	for i := 0; i < nodeCount; i++ {
		tempDir, err := os.MkdirTemp("", fmt.Sprintf("consensus-cluster-%d-*", i))
		require.NoError(t, err)
		
		// Create P2P node
		p2pConfig := &config.NodeConfig{
			Listen:         fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 12000+i),
			BootstrapPeers: []string{},
			EnableDHT:      false,
			EnableMDNS:     false,
		}
		
		ctx := context.Background()
		p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
		require.NoError(t, err)
		
		err = p2pNode.Start()
		require.NoError(t, err)
		
		// Create consensus engine
		consensusConfig := &config.ConsensusConfig{
			DataDir:            tempDir,
			BindAddr:           fmt.Sprintf("127.0.0.1:%d", 13000+i),
			Bootstrap:          i == 0, // First node bootstraps
			LogLevel:           "ERROR",
			HeartbeatTimeout:   1000 * time.Millisecond,
			ElectionTimeout:    1000 * time.Millisecond,
			CommitTimeout:      50 * time.Millisecond,
			MaxAppendEntries:   64,
			SnapshotInterval:   120 * time.Second,
			SnapshotThreshold:  8192,
		}
		
		engine, err := consensus.NewEngine(consensusConfig, p2pNode)
		require.NoError(t, err)
		
		nodes[i] = &testClusterNode{
			engine:  engine,
			p2pNode: p2pNode,
			dataDir: tempDir,
		}
	}
	
	return nodes
}

// BenchmarkConsensusOperations benchmarks consensus operations
func BenchmarkConsensusOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "consensus-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)
	
	engine := createTestConsensusEngine(b, tempDir, true)
	defer engine.Shutdown(context.Background())
	
	err = engine.Start()
	require.NoError(b, err)
	
	// Wait for leadership
	time.Sleep(2 * time.Second)
	require.True(b, engine.IsLeader(), "Node should be leader")
	
	b.ResetTimer()
	
	b.Run("Apply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench-key-%d", i)
			value := fmt.Sprintf("bench-value-%d", i)
			err := engine.Apply(key, value, nil)
			if err != nil {
				b.Fatalf("Apply failed: %v", err)
			}
		}
	})
	
	b.Run("Get", func(b *testing.B) {
		// Pre-populate some data
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("get-bench-key-%d", i)
			value := fmt.Sprintf("get-bench-value-%d", i)
			engine.Apply(key, value, nil)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("get-bench-key-%d", i%100)
			_, _ = engine.Get(key)
		}
	})
	
	b.Run("Delete", func(b *testing.B) {
		// Pre-populate data for deletion
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-bench-key-%d", i)
			value := fmt.Sprintf("delete-bench-value-%d", i)
			engine.Apply(key, value, nil)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("delete-bench-key-%d", i)
			err := engine.Delete(key)
			if err != nil {
				b.Fatalf("Delete failed: %v", err)
			}
		}
	})
}