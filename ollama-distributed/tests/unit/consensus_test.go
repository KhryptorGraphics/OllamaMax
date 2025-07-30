package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsensusEngine(t *testing.T) {
	// Create temporary directory for test
	testDir := t.TempDir()
	
	// Create test configuration
	cfg := &config.ConsensusConfig{
		DataDir:           filepath.Join(testDir, "consensus"),
		BindAddr:          "127.0.0.1:0", // Use random port
		Bootstrap:         true,
		LogLevel:          "ERROR",
		HeartbeatTimeout:  100 * time.Millisecond,
		ElectionTimeout:   100 * time.Millisecond,
		CommitTimeout:     10 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  10 * time.Second,
		SnapshotThreshold: 1024,
	}
	
	// Create P2P node (mock)
	p2pNode := createMockP2PNode(t)
	
	// Create consensus engine
	engine, err := consensus.NewEngine(cfg, p2pNode)
	require.NoError(t, err)
	
	// Start the engine
	err = engine.Start()
	require.NoError(t, err)
	
	// Wait for leadership
	time.Sleep(200 * time.Millisecond)
	
	// Test basic operations
	t.Run("TestBasicOperations", func(t *testing.T) {
		testBasicOperations(t, engine)
	})
	
	t.Run("TestStateManagement", func(t *testing.T) {
		testStateManagement(t, engine)
	})
	
	t.Run("TestLeadership", func(t *testing.T) {
		testLeadership(t, engine)
	})
	
	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = engine.Shutdown(ctx)
	require.NoError(t, err)
}

func testBasicOperations(t *testing.T, engine *consensus.Engine) {
	// Test Apply operation
	testKey := "test_key"
	testValue := "test_value"
	
	err := engine.Apply(testKey, testValue, nil)
	require.NoError(t, err)
	
	// Test Get operation
	value, exists := engine.Get(testKey)
	assert.True(t, exists)
	assert.Equal(t, testValue, value)
	
	// Test Delete operation
	err = engine.Delete(testKey)
	require.NoError(t, err)
	
	// Verify deletion
	_, exists = engine.Get(testKey)
	assert.False(t, exists)
}

func testStateManagement(t *testing.T, engine *consensus.Engine) {
	// Apply multiple key-value pairs
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": map[string]string{"nested": "value"},
	}
	
	for key, value := range testData {
		err := engine.Apply(key, value, nil)
		require.NoError(t, err)
	}
	
	// Get all state
	allState := engine.GetAll()
	
	// Verify all keys are present
	for key, expectedValue := range testData {
		actualValue, exists := allState[key]
		assert.True(t, exists, "Key %s should exist", key)
		assert.Equal(t, expectedValue, actualValue, "Value for key %s should match", key)
	}
}

func testLeadership(t *testing.T, engine *consensus.Engine) {
	// Should be leader since we bootstrapped
	assert.True(t, engine.IsLeader())
	
	// Test leadership changes channel
	leadershipCh := engine.LeadershipChanges()
	
	// Should receive initial leadership notification
	select {
	case isLeader := <-leadershipCh:
		assert.True(t, isLeader)
	case <-time.After(time.Second):
		t.Fatal("Expected leadership notification")
	}
	
	// Test stats
	stats := engine.Stats()
	assert.NotEmpty(t, stats)
	assert.Contains(t, stats, "state")
}

func TestFSM(t *testing.T) {
	// Create test FSM
	fsm := &consensus.FSM{
		State:   make(map[string]interface{}),
		ApplyCh: make(chan *consensus.ApplyEvent, 10),
	}
	
	// Test Apply
	testEvent := &consensus.ApplyEvent{
		Type:      "set",
		Key:       "test_key",
		Value:     "test_value",
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// Mock log entry
	logEntry := &raft.Log{
		Data: marshalEvent(t, testEvent),
	}
	
	result := fsm.Apply(logEntry)
	assert.Nil(t, result)
	
	// Verify state was updated
	assert.Equal(t, "test_value", fsm.State["test_key"])
	
	// Test delete
	deleteEvent := &consensus.ApplyEvent{
		Type:      "delete",
		Key:       "test_key",
		Timestamp: time.Now(),
	}
	
	logEntry = &raft.Log{
		Data: marshalEvent(t, deleteEvent),
	}
	
	result = fsm.Apply(logEntry)
	assert.Nil(t, result)
	
	// Verify key was deleted
	_, exists := fsm.State["test_key"]
	assert.False(t, exists)
}

func TestSnapshot(t *testing.T) {
	// Create test FSM with some state
	fsm := &consensus.FSM{
		State: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": map[string]string{"nested": "value"},
		},
		ApplyCh: make(chan *consensus.ApplyEvent, 10),
	}
	
	// Create snapshot
	snapshot, err := fsm.Snapshot()
	require.NoError(t, err)
	
	// Create a temporary file to persist the snapshot
	tmpFile, err := os.CreateTemp("", "snapshot_test_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	// Create mock sink
	sink := &mockSnapshotSink{file: tmpFile}
	
	// Persist snapshot
	err = snapshot.Persist(sink)
	require.NoError(t, err)
	
	// Create new FSM to restore into
	newFSM := &consensus.FSM{
		State:   make(map[string]interface{}),
		ApplyCh: make(chan *consensus.ApplyEvent, 10),
	}
	
	// Restore from snapshot
	tmpFile.Seek(0, 0)
	err = newFSM.Restore(tmpFile)
	require.NoError(t, err)
	
	// Verify restored state
	assert.Equal(t, fsm.State, newFSM.State)
}

// Helper functions and mocks

func createMockP2PNode(t *testing.T) *p2p.Node {
	// Create a minimal P2P node for testing
	ctx := context.Background()
	
	cfg := &config.P2PConfig{
		Listen:      "/ip4/127.0.0.1/tcp/0",
		Bootstrap:   []string{},
		EnableDHT:   false,
		EnablePubSub: false,
	}
	
	node, err := p2p.NewNode(ctx, cfg)
	require.NoError(t, err)
	
	return node
}

func marshalEvent(t *testing.T, event *consensus.ApplyEvent) []byte {
	data, err := json.Marshal(event)
	require.NoError(t, err)
	return data
}

type mockSnapshotSink struct {
	file *os.File
}

func (s *mockSnapshotSink) Write(p []byte) (int, error) {
	return s.file.Write(p)
}

func (s *mockSnapshotSink) Close() error {
	return s.file.Close()
}

func (s *mockSnapshotSink) ID() string {
	return "mock_snapshot"
}

func (s *mockSnapshotSink) Cancel() error {
	return s.file.Close()
}

// Benchmark tests

func BenchmarkConsensusApply(b *testing.B) {
	// Create test engine
	testDir := b.TempDir()
	
	cfg := &config.ConsensusConfig{
		DataDir:           filepath.Join(testDir, "consensus"),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		LogLevel:          "ERROR",
		HeartbeatTimeout:  100 * time.Millisecond,
		ElectionTimeout:   100 * time.Millisecond,
		CommitTimeout:     10 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  10 * time.Second,
		SnapshotThreshold: 1024,
	}
	
	p2pNode := createMockP2PNode(b)
	engine, err := consensus.NewEngine(cfg, p2pNode)
	require.NoError(b, err)
	
	err = engine.Start()
	require.NoError(b, err)
	
	// Wait for leadership
	time.Sleep(200 * time.Millisecond)
	
	b.ResetTimer()
	
	// Benchmark Apply operations
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			value := fmt.Sprintf("bench_value_%d", i)
			
			err := engine.Apply(key, value, nil)
			if err != nil {
				b.Errorf("Failed to apply: %v", err)
			}
			
			i++
		}
	})
	
	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	engine.Shutdown(ctx)
}

func BenchmarkConsensusGet(b *testing.B) {
	// Create test engine with some data
	testDir := b.TempDir()
	
	cfg := &config.ConsensusConfig{
		DataDir:           filepath.Join(testDir, "consensus"),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		LogLevel:          "ERROR",
		HeartbeatTimeout:  100 * time.Millisecond,
		ElectionTimeout:   100 * time.Millisecond,
		CommitTimeout:     10 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  10 * time.Second,
		SnapshotThreshold: 1024,
	}
	
	p2pNode := createMockP2PNode(b)
	engine, err := consensus.NewEngine(cfg, p2pNode)
	require.NoError(b, err)
	
	err = engine.Start()
	require.NoError(b, err)
	
	// Wait for leadership
	time.Sleep(200 * time.Millisecond)
	
	// Populate with test data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err := engine.Apply(key, value, nil)
		require.NoError(b, err)
	}
	
	b.ResetTimer()
	
	// Benchmark Get operations
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			_, exists := engine.Get(key)
			if !exists {
				b.Errorf("Key %s should exist", key)
			}
			i++
		}
	})
	
	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	engine.Shutdown(ctx)
}

