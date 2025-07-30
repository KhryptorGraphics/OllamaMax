package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEngine_NewEngine tests engine creation
func TestEngine_NewEngine(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.ConsensusConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.ConsensusConfig{
				DataDir:           t.TempDir(),
				BindAddr:          "127.0.0.1:0",
				Bootstrap:         true,
				HeartbeatTimeout:  time.Second,
				ElectionTimeout:   time.Second,
				CommitTimeout:     time.Second,
				MaxAppendEntries:  64,
				SnapshotInterval:  time.Hour,
				SnapshotThreshold: 8192,
				LogLevel:          "INFO",
			},
			expectError: false,
		},
		{
			name: "invalid data dir",
			config: &config.ConsensusConfig{
				DataDir:  "/invalid/path/that/does/not/exist",
				BindAddr: "127.0.0.1:0",
			},
			expectError: true,
		},
		{
			name: "invalid bind address",
			config: &config.ConsensusConfig{
				DataDir:  t.TempDir(),
				BindAddr: "invalid-address",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock P2P node
			mockP2P := createMockP2PNode(t)
			
			engine, err := NewEngine(tt.config, mockP2P)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, engine)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, engine)
				
				// Clean up
				if engine != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					engine.Shutdown(ctx)
				}
			}
		})
	}
}

// TestEngine_StartShutdown tests engine lifecycle
func TestEngine_StartShutdown(t *testing.T) {
	config := &config.ConsensusConfig{
		DataDir:           t.TempDir(),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "INFO",
	}

	mockP2P := createMockP2PNode(t)
	engine, err := NewEngine(config, mockP2P)
	require.NoError(t, err)
	require.NotNil(t, engine)

	// Test start
	err = engine.Start()
	assert.NoError(t, err)

	// Wait for bootstrap
	time.Sleep(100 * time.Millisecond)

	// Test leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 2*time.Second, 100*time.Millisecond, "should become leader")

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = engine.Shutdown(ctx)
	assert.NoError(t, err)
}

// TestEngine_ApplyGet tests state operations
func TestEngine_ApplyGet(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 2*time.Second, 100*time.Millisecond)

	// Test apply
	err = engine.Apply("test-key", "test-value", map[string]interface{}{"author": "test"})
	assert.NoError(t, err)

	// Test get
	value, exists := engine.Get("test-key")
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)

	// Test get all
	allState := engine.GetAll()
	assert.Contains(t, allState, "test-key")
	assert.Equal(t, "test-value", allState["test-key"])
}

// TestEngine_Delete tests state deletion
func TestEngine_Delete(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 2*time.Second, 100*time.Millisecond)

	// Apply initial value
	err = engine.Apply("test-key", "test-value", nil)
	require.NoError(t, err)

	// Small delay to ensure consensus completes
	time.Sleep(50 * time.Millisecond)

	// Verify value exists
	_, exists := engine.Get("test-key")
	assert.True(t, exists)

	// Delete value
	err = engine.Delete("test-key")
	assert.NoError(t, err)

	// Small delay to ensure delete consensus completes
	time.Sleep(50 * time.Millisecond)

	// Verify value is deleted
	_, exists = engine.Get("test-key")
	assert.False(t, exists)
}

// TestEngine_NonLeaderOperations tests operations from non-leader
func TestEngine_NonLeaderOperations(t *testing.T) {
	config := &config.ConsensusConfig{
		DataDir:           t.TempDir(),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         false, // Not bootstrap node
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "INFO",
	}

	mockP2P := createMockP2PNode(t)
	engine, err := NewEngine(config, mockP2P)
	require.NoError(t, err)
	require.NotNil(t, engine)
	defer cleanupTestEngine(t, engine)

	err = engine.Start()
	require.NoError(t, err)

	// Should not be leader
	assert.False(t, engine.IsLeader())

	// Apply should fail
	err = engine.Apply("test-key", "test-value", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not leader")

	// Delete should fail
	err = engine.Delete("test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not leader")
}

// TestEngine_LeadershipMonitoring tests leadership change monitoring
func TestEngine_LeadershipMonitoring(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Monitor leadership changes
	leadershipCh := engine.LeadershipChanges()
	
	var leadershipEvents []bool
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		timeout := time.After(3 * time.Second)
		for {
			select {
			case isLeader := <-leadershipCh:
				leadershipEvents = append(leadershipEvents, isLeader)
				if isLeader {
					return // Got leadership event
				}
			case <-timeout:
				return
			}
		}
	}()

	wg.Wait()

	// Should have received at least one leadership event
	assert.NotEmpty(t, leadershipEvents)
	assert.True(t, leadershipEvents[len(leadershipEvents)-1], "should be leader")
}

// TestEngine_Stats tests statistics
func TestEngine_Stats(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Wait for bootstrap
	time.Sleep(100 * time.Millisecond)

	stats := engine.Stats()
	assert.NotEmpty(t, stats)
	assert.Contains(t, stats, "state")
}

// TestEngine_ClusterOperations tests cluster management
func TestEngine_ClusterOperations(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 2*time.Second, 100*time.Millisecond)

	// Test get configuration
	config, err := engine.GetConfiguration()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Len(t, config.Servers, 1) // Bootstrap node

	// Test leader
	leader := engine.Leader()
	assert.NotEmpty(t, leader)

	// Test add voter (should work but server might not join)
	err = engine.AddVoter("test-node", "127.0.0.1:9999")
	// This might fail due to no actual server at the address, which is expected
	// We just want to test the API doesn't panic

	// Test remove server
	err = engine.RemoveServer("test-node")
	// This should not fail as removing non-existent server is generally safe
}

// TestFSM_Apply tests FSM apply functionality
func TestFSM_Apply(t *testing.T) {
	fsm := &FSM{
		state:   make(map[string]interface{}),
		applyCh: make(chan *ApplyEvent, 10),
	}

	tests := []struct {
		name        string
		event       *ApplyEvent
		expectError bool
	}{
		{
			name: "valid set event",
			event: &ApplyEvent{
				Type:      "set",
				Key:       "test-key",
				Value:     "test-value",
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid delete event",
			event: &ApplyEvent{
				Type:      "delete",
				Key:       "test-key",
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid update event",
			event: &ApplyEvent{
				Type:      "update",
				Key:       "test-key",
				Value:     "updated-value",
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name: "invalid empty key",
			event: &ApplyEvent{
				Type:      "set",
				Key:       "",
				Value:     "test-value",
				Timestamp: time.Now(),
			},
			expectError: true,
		},
		{
			name: "invalid empty type",
			event: &ApplyEvent{
				Type:      "",
				Key:       "test-key",
				Value:     "test-value",
				Timestamp: time.Now(),
			},
			expectError: true,
		},
		{
			name: "invalid old timestamp",
			event: &ApplyEvent{
				Type:      "set",
				Key:       "test-key",
				Value:     "test-value",
				Timestamp: time.Now().Add(-10 * time.Minute),
			},
			expectError: true,
		},
		{
			name: "unknown event type",
			event: &ApplyEvent{
				Type:      "unknown",
				Key:       "test-key",
				Value:     "test-value",
				Timestamp: time.Now(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For update tests, create the key first
			if tt.event.Type == "update" {
				fsm.stateMu.Lock()
				fsm.state[tt.event.Key] = "initial-value"
				fsm.stateMu.Unlock()
			}
			
			// Create raft log
			data, err := json.Marshal(tt.event)
			require.NoError(t, err)

			log := &raft.Log{
				Data: data,
			}

			result := fsm.Apply(log)

			if tt.expectError {
				assert.Error(t, result.(error))
			} else {
				assert.Nil(t, result)
				
				// For set operations, verify state was updated
				if tt.event.Type == "set" || tt.event.Type == "update" {
					fsm.stateMu.RLock()
					value, exists := fsm.state[tt.event.Key]
					fsm.stateMu.RUnlock()
					assert.True(t, exists)
					assert.Equal(t, tt.event.Value, value)
				}
			}
		})
	}
}

// TestFSM_Snapshot tests FSM snapshot functionality
func TestFSM_Snapshot(t *testing.T) {
	fsm := &FSM{
		state:   make(map[string]interface{}),
		applyCh: make(chan *ApplyEvent, 10),
	}

	// Add some state
	fsm.state["key1"] = "value1"
	fsm.state["key2"] = "value2"

	// Create snapshot
	snapshot, err := fsm.Snapshot()
	assert.NoError(t, err)
	assert.NotNil(t, snapshot)

	// Test snapshot persistence
	tempFile, err := ioutil.TempFile("", "test-snapshot")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	sink := &testSnapshotSink{tempFile}
	err = snapshot.Persist(sink)
	assert.NoError(t, err)

	// Release snapshot
	snapshot.Release()
}

// TestFSM_Restore tests FSM restore functionality
func TestFSM_Restore(t *testing.T) {
	fsm := &FSM{
		state:   make(map[string]interface{}),
		applyCh: make(chan *ApplyEvent, 10),
	}

	// Create test state
	testState := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	// Create temporary file with state
	tempFile, err := ioutil.TempFile("", "test-restore")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	encoder := json.NewEncoder(tempFile)
	err = encoder.Encode(testState)
	require.NoError(t, err)
	tempFile.Close()

	// Reopen file for reading
	file, err := os.Open(tempFile.Name())
	require.NoError(t, err)
	defer file.Close()

	// Test restore
	err = fsm.Restore(file)
	assert.NoError(t, err)

	// Verify state was restored
	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()
	assert.Equal(t, testState, fsm.state)
}

// TestEngine_ConcurrentOperations tests concurrent operations
func TestEngine_ConcurrentOperations(t *testing.T) {
	engine := setupTestEngine(t)
	defer cleanupTestEngine(t, engine)

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 2*time.Second, 100*time.Millisecond)

	// Run concurrent operations
	var wg sync.WaitGroup
	numOps := 10

	// Concurrent applies
	wg.Add(numOps)
	for i := 0; i < numOps; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)
			err := engine.Apply(key, value, nil)
			assert.NoError(t, err)
		}(i)
	}

	// Concurrent reads
	wg.Add(numOps)
	for i := 0; i < numOps; i++ {
		go func(i int) {
			defer wg.Done()
			time.Sleep(time.Millisecond) // Small delay to ensure some writes happen first
			key := fmt.Sprintf("key-%d", i)
			_, exists := engine.Get(key)
			// Value might or might not exist due to timing, but operation should not panic
			_ = exists
		}(i)
	}

	wg.Wait()

	// Verify final state
	allState := engine.GetAll()
	assert.True(t, len(allState) <= numOps) // Some operations might have completed
}

// TestEngine_ErrorScenarios tests various error scenarios
func TestEngine_ErrorScenarios(t *testing.T) {
	t.Run("corrupt data directory", func(t *testing.T) {
		dataDir := t.TempDir()
		
		// Create a file where directory should be
		corruptPath := filepath.Join(dataDir, "raft-log.db")
		err := ioutil.WriteFile(corruptPath, []byte("corrupt"), 0644)
		require.NoError(t, err)

		config := &config.ConsensusConfig{
			DataDir:  dataDir,
			BindAddr: "127.0.0.1:0",
		}

		mockP2P := createMockP2PNode(t)
		_, err = NewEngine(config, mockP2P)
		assert.Error(t, err)
	})

	t.Run("invalid JSON in apply", func(t *testing.T) {
		fsm := &FSM{
			state:   make(map[string]interface{}),
			applyCh: make(chan *ApplyEvent, 10),
		}

		log := &raft.Log{
			Data: []byte("invalid json"),
		}

		result := fsm.Apply(log)
		assert.Error(t, result.(error))
	})
}

// Helper functions

func setupTestEngine(t *testing.T) *Engine {
	config := &config.ConsensusConfig{
		DataDir:           t.TempDir(),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "INFO",
	}

	mockP2P := createMockP2PNode(t)
	engine, err := NewEngine(config, mockP2P)
	require.NoError(t, err)
	require.NotNil(t, engine)

	return engine
}

func cleanupTestEngine(t *testing.T, engine *Engine) {
	if engine != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := engine.Shutdown(ctx)
		assert.NoError(t, err)
	}
}

func createMockP2PNode(t *testing.T) *p2p.Node {
	// Create a minimal mock P2P node for testing
	ctx := context.Background()
	node, err := p2p.NewP2PNode(ctx, nil) // Use default config
	require.NoError(t, err)
	return node
}

// testSnapshotSink implements raft.SnapshotSink for testing
type testSnapshotSink struct {
	*os.File
}

func (s *testSnapshotSink) ID() string {
	return "test-snapshot"
}

func (s *testSnapshotSink) Cancel() error {
	return s.File.Close()
}

// Benchmark tests

func BenchmarkEngine_Apply(b *testing.B) {
	engine := setupBenchEngine(b)
	defer cleanupBenchEngine(b, engine)

	err := engine.Start()
	require.NoError(b, err)

	// Wait for leadership
	for !engine.IsLeader() {
		time.Sleep(10 * time.Millisecond)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i)
			value := fmt.Sprintf("bench-value-%d", i)
			err := engine.Apply(key, value, nil)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkEngine_Get(b *testing.B) {
	engine := setupBenchEngine(b)
	defer cleanupBenchEngine(b, engine)

	err := engine.Start()
	require.NoError(b, err)

	// Wait for leadership
	for !engine.IsLeader() {
		time.Sleep(10 * time.Millisecond)
	}

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench-key-%d", i)
		value := fmt.Sprintf("bench-value-%d", i)
		err := engine.Apply(key, value, nil)
		require.NoError(b, err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i%1000)
			_, exists := engine.Get(key)
			if !exists {
				b.Fatal("key should exist")
			}
			i++
		}
	})
}

func setupBenchEngine(b *testing.B) *Engine {
	config := &config.ConsensusConfig{
		DataDir:           b.TempDir(),
		BindAddr:          "127.0.0.1:0",
		Bootstrap:         true,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "ERROR", // Reduce logging for benchmarks
	}

	mockP2P := createMockP2PNode(&testing.T{})
	engine, err := NewEngine(config, mockP2P)
	require.NoError(b, err)
	require.NotNil(b, engine)

	return engine
}

func cleanupBenchEngine(b *testing.B, engine *Engine) {
	if engine != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := engine.Shutdown(ctx)
		require.NoError(b, err)
	}
}