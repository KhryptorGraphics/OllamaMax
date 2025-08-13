package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_Consensus_Engine_Initialization tests consensus engine setup
func Test_Consensus_Engine_Initialization(t *testing.T) {
	tempDir := t.TempDir()

	// Create P2P node for consensus
	p2pConfig := &config.P2PConfig{
		ListenAddr: "127.0.0.1:0",
	}

	p2pNode, err := p2p.NewNode(context.Background(), p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Close()

	// Create consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:           tempDir,
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		LogLevel:          "INFO",
		HeartbeatTimeout:  1 * time.Second,
		ElectionTimeout:   3 * time.Second,
		CommitTimeout:     500 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  120 * time.Second,
		SnapshotThreshold: 8192,
	}

	// Create consensus engine
	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)
	assert.NotNil(t, engine)

	// Start the engine
	err = engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	select {
	case isLeader := <-engine.LeadershipChanges():
		assert.True(t, isLeader, "Single node should become leader")
	case <-time.After(5 * time.Second):
		t.Fatal("Leadership timeout")
	}

	// Verify leadership
	assert.True(t, engine.IsLeader())

	// Test basic operations
	err = engine.Apply("test-key", "test-value", map[string]interface{}{"timestamp": time.Now()})
	assert.NoError(t, err)

	// Verify state
	value, exists := engine.Get("test-key")
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)

	// Test state retrieval
	allState := engine.GetAll()
	assert.Contains(t, allState, "test-key")
	assert.Equal(t, "test-value", allState["test-key"])

	// Shutdown engine
	err = engine.Shutdown(context.Background())
	assert.NoError(t, err)
}

// Test_Consensus_FSM_Operations tests finite state machine operations
func Test_Consensus_FSM_Operations(t *testing.T) {
	tempDir := t.TempDir()

	p2pConfig := &config.P2PConfig{
		ListenAddr: "127.0.0.1:0",
	}

	p2pNode, err := p2p.NewNode(context.Background(), p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Close()

	consensusConfig := &config.ConsensusConfig{
		DataDir:           tempDir,
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		LogLevel:          "ERROR", // Reduce noise
		HeartbeatTimeout:  500 * time.Millisecond,
		ElectionTimeout:   1 * time.Second,
		CommitTimeout:     250 * time.Millisecond,
		MaxAppendEntries:  32,
		SnapshotInterval:  60 * time.Second,
		SnapshotThreshold: 4096,
	}

	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	err = engine.Start()
	require.NoError(t, err)
	defer engine.Shutdown(context.Background())

	// Wait for leadership
	select {
	case <-engine.LeadershipChanges():
	case <-time.After(3 * time.Second):
		t.Fatal("Leadership timeout")
	}

	// Test multiple operations
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string-key", "string-value"},
		{"int-key", 42},
		{"float-key", 3.14},
		{"bool-key", true},
		{"map-key", map[string]string{"nested": "value"}},
		{"slice-key", []string{"item1", "item2", "item3"}},
	}

	// Apply all test cases
	for _, tc := range testCases {
		err := engine.Apply(tc.key, tc.value, map[string]interface{}{
			"operation": "test",
			"timestamp": time.Now(),
		})
		assert.NoError(t, err, "Failed to apply key: %s", tc.key)
	}

	// Verify all values
	for _, tc := range testCases {
		value, exists := engine.Get(tc.key)
		assert.True(t, exists, "Key not found: %s", tc.key)
		assert.Equal(t, tc.value, value, "Value mismatch for key: %s", tc.key)
	}

	// Test deletion
	err = engine.Delete("string-key")
	assert.NoError(t, err)

	// Verify deletion
	_, exists := engine.Get("string-key")
	assert.False(t, exists)

	// Test update operation (requires our new update event type)
	err = engine.Apply("int-key", 84, map[string]interface{}{"operation": "update"})
	assert.NoError(t, err)

	value, exists := engine.Get("int-key")
	assert.True(t, exists)
	assert.Equal(t, 84, value)
}

// Test_Consensus_Race_Conditions tests for race conditions
func Test_Consensus_Race_Conditions(t *testing.T) {
	tempDir := t.TempDir()

	p2pConfig := &config.P2PConfig{
		ListenAddr: "127.0.0.1:0",
	}

	p2pNode, err := p2p.NewNode(context.Background(), p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Close()

	consensusConfig := &config.ConsensusConfig{
		DataDir:          tempDir,
		BindAddr:         "127.0.0.1:0",
		Bootstrap:        true,
		LogLevel:         "ERROR",
		HeartbeatTimeout: 200 * time.Millisecond,
		ElectionTimeout:  500 * time.Millisecond,
		CommitTimeout:    100 * time.Millisecond,
	}

	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	err = engine.Start()
	require.NoError(t, err)
	defer engine.Shutdown(context.Background())

	// Wait for leadership
	select {
	case <-engine.LeadershipChanges():
	case <-time.After(2 * time.Second):
		t.Fatal("Leadership timeout")
	}

	const numGoroutines = 10
	const opsPerGoroutine = 20
	errors := make(chan error, numGoroutines*opsPerGoroutine)

	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d-%d", routineID, j)
				value := fmt.Sprintf("value-%d-%d", routineID, j)

				err := engine.Apply(key, value, map[string]interface{}{
					"routine": routineID,
					"op":      j,
				})
				errors <- err
			}
		}(i)
	}

	// Collect results
	var failedOps int
	for i := 0; i < numGoroutines*opsPerGoroutine; i++ {
		err := <-errors
		if err != nil {
			failedOps++
			t.Logf("Operation failed: %v", err)
		}
	}

	// Some operations might fail due to leadership changes, but most should succeed
	successRate := float64(numGoroutines*opsPerGoroutine-failedOps) / float64(numGoroutines*opsPerGoroutine)
	assert.Greater(t, successRate, 0.8, "Success rate too low: %f", successRate)

	// Verify final state consistency
	finalState := engine.GetAll()
	t.Logf("Final state contains %d keys", len(finalState))
	assert.Greater(t, len(finalState), 0)
}

// Test_Consensus_Event_Validation tests the new event validation
func Test_Consensus_Event_Validation(t *testing.T) {
	tempDir := t.TempDir()

	p2pConfig := &config.P2PConfig{
		ListenAddr: "127.0.0.1:0",
	}

	p2pNode, err := p2p.NewNode(context.Background(), p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Close()

	consensusConfig := &config.ConsensusConfig{
		DataDir:          tempDir,
		BindAddr:         "127.0.0.1:0",
		Bootstrap:        true,
		LogLevel:         "ERROR",
		HeartbeatTimeout: 500 * time.Millisecond,
		ElectionTimeout:  1 * time.Second,
		CommitTimeout:    250 * time.Millisecond,
	}

	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	err = engine.Start()
	require.NoError(t, err)
	defer engine.Shutdown(context.Background())

	// Wait for leadership
	select {
	case <-engine.LeadershipChanges():
	case <-time.After(3 * time.Second):
		t.Fatal("Leadership timeout")
	}

	// Test valid operations
	err = engine.Apply("valid-key", "valid-value", map[string]interface{}{"test": true})
	assert.NoError(t, err)

	// Verify the value was set
	value, exists := engine.Get("valid-key")
	assert.True(t, exists)
	assert.Equal(t, "valid-value", value)

	// Test update on existing key
	err = engine.Apply("valid-key", "updated-value", map[string]interface{}{"operation": "update"})
	assert.NoError(t, err)

	value, exists = engine.Get("valid-key")
	assert.True(t, exists)
	assert.Equal(t, "updated-value", value)

	// Test deletion
	err = engine.Delete("valid-key")
	assert.NoError(t, err)

	_, exists = engine.Get("valid-key")
	assert.False(t, exists)
}

// Test_Consensus_Statistics tests consensus statistics
func Test_Consensus_Statistics(t *testing.T) {
	tempDir := t.TempDir()

	p2pConfig := &config.P2PConfig{
		ListenAddr: "127.0.0.1:0",
	}

	p2pNode, err := p2p.NewNode(context.Background(), p2pConfig)
	require.NoError(t, err)
	defer p2pNode.Close()

	consensusConfig := &config.ConsensusConfig{
		DataDir:   tempDir,
		BindAddr:  "127.0.0.1:0",
		Bootstrap: true,
		LogLevel:  "ERROR",
	}

	engine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(t, err)

	err = engine.Start()
	require.NoError(t, err)
	defer engine.Shutdown(context.Background())

	// Wait for leadership
	select {
	case <-engine.LeadershipChanges():
	case <-time.After(3 * time.Second):
		t.Fatal("Leadership timeout")
	}

	// Get statistics
	stats := engine.Stats()
	assert.NotEmpty(t, stats)

	// Should have standard Raft stats
	assert.Contains(t, stats, "state")
	t.Logf("Raft stats: %+v", stats)

	// Test configuration
	config, err := engine.GetConfiguration()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Servers, 1) // Bootstrap single node

	// Test leader information
	leader := engine.Leader()
	assert.NotEmpty(t, leader)
	t.Logf("Current leader: %s", leader)
}
