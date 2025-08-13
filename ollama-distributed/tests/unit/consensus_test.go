package unit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsensusEngine(t *testing.T) {
	// Use the helper function to create a mock consensus engine
	engine := createMockConsensusEngine(t)
	require.NotNil(t, engine)

	// Start the engine
	err := engine.Start()
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

// Helper functions and mocks

// createMockP2PNode is defined in test_helpers.go to avoid duplication

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
	engine := createMockConsensusEngine(&testing.T{})

	err := engine.Start()
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
	engine := createMockConsensusEngine(&testing.T{})

	err := engine.Start()
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
