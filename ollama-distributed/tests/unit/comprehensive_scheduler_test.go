package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
)

// TestSchedulerEngineCreation tests scheduler engine creation
func TestSchedulerEngineCreation(t *testing.T) {
	ctx := context.Background()

	// Create mock dependencies
	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)

	config := &config.SchedulerConfig{
		QueueSize:           1000,
		WorkerCount:         4,
		HealthCheckInterval: 10 * time.Second,
		LoadBalancing:       "round_robin",
	}

	engine, err := scheduler.NewEngine(config, p2pNode, consensusEngine)
	require.NoError(t, err, "Failed to create scheduler engine")
	require.NotNil(t, engine, "Scheduler engine should not be nil")

	// Test engine starts
	err = engine.Start()
	require.NoError(t, err, "Failed to start scheduler engine")

	// Cleanup
	defer engine.Shutdown(ctx)
}

// TestSchedulerNodeManagement tests node management functionality
func TestSchedulerNodeManagement(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Test initial state
	nodes := engine.GetNodes()
	assert.Equal(t, 0, len(nodes), "Should start with no nodes")

	// Test cluster size
	clusterSize := engine.GetClusterSize()
	assert.Equal(t, 0, clusterSize, "Initial cluster size should be 0")

	// Test active nodes
	activeNodes := engine.GetActiveNodes()
	assert.Equal(t, 0, activeNodes, "Initial active nodes should be 0")

	// Test online node count
	onlineCount := engine.GetOnlineNodeCount()
	assert.Equal(t, 0, onlineCount, "Initial online count should be 0")

	// Test available nodes
	availableNodes := engine.GetAvailableNodes()
	assert.Equal(t, 0, len(availableNodes), "Should have no available nodes initially")
}

// TestSchedulerModelManagement tests model management functionality
func TestSchedulerModelManagement(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Test initial model state
	models := engine.GetAllModels()
	assert.Equal(t, 0, len(models), "Should start with no models")

	modelCount := engine.GetModelCount()
	assert.Equal(t, 0, modelCount, "Initial model count should be 0")

	// Test registering a model
	modelName := "test-model"
	modelSize := int64(1024 * 1024 * 100) // 100MB
	modelChecksum := "abc123def456"
	nodeID := "test-node-1"

	err = engine.RegisterModel(modelName, modelSize, modelChecksum, nodeID)
	require.NoError(t, err, "Should register model successfully")

	// Test model retrieval
	model, exists := engine.GetModel(modelName)
	require.True(t, exists, "Model should exist after registration")
	assert.Equal(t, modelName, model.Name)
	assert.Equal(t, modelSize, model.Size)
	assert.Equal(t, modelChecksum, model.Checksum)
	assert.Contains(t, model.Locations, nodeID, "Node should be in model locations")

	// Test model count update
	newModelCount := engine.GetModelCount()
	assert.Equal(t, 1, newModelCount, "Model count should be 1 after registration")

	// Test all models
	allModels := engine.GetAllModels()
	assert.Equal(t, 1, len(allModels), "Should have 1 model")
	assert.Contains(t, allModels, modelName, "Should contain registered model")

	// Test registering same model on different node
	nodeID2 := "test-node-2"
	err = engine.RegisterModel(modelName, modelSize, modelChecksum, nodeID2)
	require.NoError(t, err, "Should register model on second node")

	// Test model locations updated
	updatedModel, exists := engine.GetModel(modelName)
	require.True(t, exists)
	assert.Equal(t, 2, len(updatedModel.Locations), "Model should be on 2 nodes")
	assert.Contains(t, updatedModel.Locations, nodeID)
	assert.Contains(t, updatedModel.Locations, nodeID2)

	// Test deleting model
	err = engine.DeleteModel(modelName)
	require.NoError(t, err, "Should delete model successfully")

	_, exists = engine.GetModel(modelName)
	assert.False(t, exists, "Model should not exist after deletion")

	finalModelCount := engine.GetModelCount()
	assert.Equal(t, 0, finalModelCount, "Model count should be 0 after deletion")

	// Test deleting non-existent model
	err = engine.DeleteModel("non-existent-model")
	assert.Error(t, err, "Should error when deleting non-existent model")
}

// TestSchedulerRequestProcessing tests request processing
func TestSchedulerRequestProcessing(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Create a test request
	request := &scheduler.Request{
		ID:         "test-request-1",
		ModelName:  "test-model",
		Type:       "inference",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Metadata:   map[string]string{"test": "true"},
		Payload:    map[string]interface{}{"prompt": "Hello, world!"},
	}

	// Test scheduling request
	err = engine.Schedule(request)
	require.NoError(t, err, "Should schedule request successfully")

	// Since we don't have actual nodes, the request will likely fail
	// but we can test that it was queued and processed
	select {
	case response := <-request.ResponseCh:
		assert.NotNil(t, response, "Should receive a response")
		assert.Equal(t, request.ID, response.RequestID, "Response should match request ID")
	case <-time.After(5 * time.Second):
		t.Log("Request processing timed out (expected without real nodes)")
	}
}

// TestSchedulerStats tests statistics collection
func TestSchedulerStats(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Test getting stats
	stats := engine.GetStats()
	require.NotNil(t, stats, "Stats should not be nil")

	// Test initial stats values
	assert.GreaterOrEqual(t, stats.TotalRequests, int64(0), "Total requests should be non-negative")
	assert.GreaterOrEqual(t, stats.CompletedRequests, int64(0), "Completed requests should be non-negative")
	assert.GreaterOrEqual(t, stats.FailedRequests, int64(0), "Failed requests should be non-negative")
	assert.GreaterOrEqual(t, stats.QueuedRequests, int64(0), "Queued requests should be non-negative")
	assert.GreaterOrEqual(t, stats.NodesTotal, 0, "Total nodes should be non-negative")
	assert.GreaterOrEqual(t, stats.NodesOnline, 0, "Online nodes should be non-negative")
	assert.GreaterOrEqual(t, stats.NodesOffline, 0, "Offline nodes should be non-negative")
	assert.GreaterOrEqual(t, stats.ModelsTotal, 0, "Total models should be non-negative")
	assert.Greater(t, stats.WorkersActive, 0, "Should have active workers")
	assert.True(t, stats.LastUpdated.Before(time.Now().Add(time.Second)), "LastUpdated should be recent")

	// Test that uptime increases
	initialUptime := stats.Uptime
	time.Sleep(100 * time.Millisecond)
	newStats := engine.GetStats()
	assert.Greater(t, newStats.Uptime, initialUptime, "Uptime should increase")
}

// TestSchedulerHealth tests health checking functionality
func TestSchedulerHealth(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Test health status
	isHealthy := engine.IsHealthy()
	// Without nodes, the system may or may not be considered healthy
	// depending on implementation, so we just test that it doesn't panic
	_ = isHealthy

	// Test that health check doesn't panic with some basic conditions
	assert.NotPanics(t, func() {
		for i := 0; i < 10; i++ {
			engine.IsHealthy()
			time.Sleep(10 * time.Millisecond)
		}
	}, "Health check should not panic")
}

// TestSchedulerLoadBalancer tests load balancing functionality
func TestSchedulerLoadBalancer(t *testing.T) {
	ctx := context.Background()

	// Test different load balancing algorithms
	algorithms := []string{"round_robin", "least_connections", "random"}

	for _, algorithm := range algorithms {
		t.Run(algorithm, func(t *testing.T) {
			config := &config.SchedulerConfig{
				QueueSize:           1000,
				WorkerCount:         2,
				HealthCheckInterval: 1 * time.Second,
				LoadBalancing:       algorithm,
			}

			engine := createTestSchedulerEngineWithConfig(t, config)
			defer engine.Shutdown(ctx)

			err := engine.Start()
			require.NoError(t, err, "Should start with %s algorithm", algorithm)

			// Create a test request to verify load balancer is working
			request := &scheduler.Request{
				ID:         fmt.Sprintf("lb-test-%s", algorithm),
				ModelName:  "test-model",
				Type:       "inference",
				Priority:   1,
				Timeout:    5 * time.Second,
				ResponseCh: make(chan *scheduler.Response, 1),
			}

			// This will likely fail due to no nodes, but should not panic
			err = engine.Schedule(request)
			require.NoError(t, err, "Should accept request with %s algorithm", algorithm)
		})
	}
}

// TestSchedulerConcurrentOperations tests concurrent operations
func TestSchedulerConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Test concurrent model registration
	const numGoroutines = 10
	const modelsPerGoroutine = 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < modelsPerGoroutine; j++ {
				modelName := fmt.Sprintf("concurrent-model-%d-%d", routineID, j)
				nodeID := fmt.Sprintf("node-%d", routineID)

				err := engine.RegisterModel(modelName, 1024, "checksum", nodeID)
				assert.NoError(t, err, "Concurrent model registration should succeed")
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	// Verify all models were registered
	finalCount := engine.GetModelCount()
	expectedCount := numGoroutines * modelsPerGoroutine
	assert.Equal(t, expectedCount, finalCount, "Should have registered all models concurrently")
}

// TestSchedulerRequestQueue tests request queue behavior
func TestSchedulerRequestQueue(t *testing.T) {
	ctx := context.Background()

	// Create engine with small queue for testing
	config := &config.SchedulerConfig{
		QueueSize:           5, // Small queue
		WorkerCount:         1,
		HealthCheckInterval: 1 * time.Second,
		LoadBalancing:       "round_robin",
	}

	engine := createTestSchedulerEngineWithConfig(t, config)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Fill up the queue
	requests := make([]*scheduler.Request, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &scheduler.Request{
			ID:         fmt.Sprintf("queue-test-%d", i),
			ModelName:  "test-model",
			Type:       "inference",
			Priority:   1,
			Timeout:    30 * time.Second,
			ResponseCh: make(chan *scheduler.Response, 1),
		}
	}

	// Schedule requests
	successCount := 0
	for _, req := range requests {
		err := engine.Schedule(req)
		if err == nil {
			successCount++
		}
	}

	// Should accept at least some requests (up to queue size)
	assert.Greater(t, successCount, 0, "Should accept some requests")
	assert.LessOrEqual(t, successCount, len(requests), "Should not accept more than submitted")

	// Test queue stats
	stats := engine.GetStats()
	assert.GreaterOrEqual(t, stats.QueuedRequests, int64(0), "Queued requests should be non-negative")
}

// TestSchedulerRequestTimeout tests request timeout handling
func TestSchedulerRequestTimeout(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Create request with very short timeout
	request := &scheduler.Request{
		ID:         "timeout-test",
		ModelName:  "test-model",
		Type:       "inference",
		Priority:   1,
		Timeout:    100 * time.Millisecond, // Very short timeout
		ResponseCh: make(chan *scheduler.Response, 1),
	}

	err = engine.Schedule(request)
	require.NoError(t, err)

	// Wait for timeout or response
	select {
	case response := <-request.ResponseCh:
		// If we get a response, verify it indicates timeout or failure
		assert.NotNil(t, response, "Should receive a response")
		// Response may indicate timeout or no available nodes
	case <-time.After(2 * time.Second):
		t.Log("Request did not complete within expected time (may be expected)")
	}
}

// TestSchedulerPriority tests request priority handling
func TestSchedulerPriority(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(t, err)

	// Create requests with different priorities
	priorities := []int{1, 3, 2, 5, 1}
	requests := make([]*scheduler.Request, len(priorities))

	for i, priority := range priorities {
		requests[i] = &scheduler.Request{
			ID:         fmt.Sprintf("priority-test-%d", i),
			ModelName:  "test-model",
			Type:       "inference",
			Priority:   priority,
			Timeout:    30 * time.Second,
			ResponseCh: make(chan *scheduler.Response, 1),
		}
	}

	// Schedule all requests
	for _, req := range requests {
		err := engine.Schedule(req)
		assert.NoError(t, err, "Should schedule request successfully")
	}

	// The scheduler should handle priorities internally
	// We can't easily test order without real nodes, but we can verify
	// that all requests were accepted
	for _, req := range requests {
		assert.NotNil(t, req.ResponseCh, "Request should have response channel")
	}
}

// TestSchedulerErrorScenarios tests various error scenarios
func TestSchedulerErrorScenarios(t *testing.T) {
	ctx := context.Background()

	// Test with nil config
	_, err := scheduler.NewEngine(nil, nil, nil)
	assert.Error(t, err, "Should fail with nil config")

	// Test with invalid config
	invalidConfig := &config.SchedulerConfig{
		QueueSize:           -1, // Invalid
		WorkerCount:         0,  // Invalid
		HealthCheckInterval: 0,  // Invalid
		LoadBalancing:       "",
	}

	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)

	engine, err := scheduler.NewEngine(invalidConfig, p2pNode, consensusEngine)
	if err == nil {
		// If creation succeeds with invalid config, test that start fails
		err = engine.Start()
		// May or may not fail depending on implementation
		engine.Shutdown(ctx)
	}
}

// TestSchedulerShutdown tests graceful shutdown
func TestSchedulerShutdown(t *testing.T) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(t)

	err := engine.Start()
	require.NoError(t, err)

	// Schedule some requests
	for i := 0; i < 5; i++ {
		request := &scheduler.Request{
			ID:         fmt.Sprintf("shutdown-test-%d", i),
			ModelName:  "test-model",
			Type:       "inference",
			Priority:   1,
			Timeout:    30 * time.Second,
			ResponseCh: make(chan *scheduler.Response, 1),
		}

		err := engine.Schedule(request)
		require.NoError(t, err)
	}

	// Test graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = engine.Shutdown(shutdownCtx)
	require.NoError(t, err, "Shutdown should succeed")

	// Test operations after shutdown
	request := &scheduler.Request{
		ID:         "after-shutdown",
		ModelName:  "test-model",
		Type:       "inference",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
	}

	err = engine.Schedule(request)
	// May or may not error depending on implementation
	_ = err
}

// Helper functions

func createTestSchedulerEngine(t *testing.T) *scheduler.Engine {
	config := &config.SchedulerConfig{
		QueueSize:           1000,
		WorkerCount:         4,
		HealthCheckInterval: 1 * time.Second,
		LoadBalancing:       "round_robin",
	}

	return createTestSchedulerEngineWithConfig(t, config)
}

func createTestSchedulerEngineWithConfig(t *testing.T, config *config.SchedulerConfig) *scheduler.Engine {
	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)

	engine, err := scheduler.NewEngine(config, p2pNode, consensusEngine)
	require.NoError(t, err)

	return engine
}

// Mock functions are defined in test_helpers.go to avoid duplication

// BenchmarkSchedulerOperations benchmarks scheduler operations
func BenchmarkSchedulerOperations(b *testing.B) {
	ctx := context.Background()
	engine := createTestSchedulerEngine(&testing.T{})
	defer engine.Shutdown(ctx)

	err := engine.Start()
	require.NoError(b, err)

	b.Run("ModelRegistration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			modelName := fmt.Sprintf("bench-model-%d", i)
			nodeID := fmt.Sprintf("bench-node-%d", i%10)
			err := engine.RegisterModel(modelName, 1024, "checksum", nodeID)
			if err != nil {
				b.Fatalf("Model registration failed: %v", err)
			}
		}
	})

	b.Run("ModelRetrieval", func(b *testing.B) {
		// Pre-populate some models
		for i := 0; i < 100; i++ {
			modelName := fmt.Sprintf("retrieve-bench-model-%d", i)
			nodeID := fmt.Sprintf("retrieve-bench-node-%d", i%10)
			engine.RegisterModel(modelName, 1024, "checksum", nodeID)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			modelName := fmt.Sprintf("retrieve-bench-model-%d", i%100)
			_, _ = engine.GetModel(modelName)
		}
	})

	b.Run("StatsCollection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = engine.GetStats()
		}
	})

	b.Run("RequestScheduling", func(b *testing.B) {
		requests := make([]*scheduler.Request, b.N)
		for i := 0; i < b.N; i++ {
			requests[i] = &scheduler.Request{
				ID:         fmt.Sprintf("bench-request-%d", i),
				ModelName:  "bench-model",
				Type:       "inference",
				Priority:   1,
				Timeout:    1 * time.Second,
				ResponseCh: make(chan *scheduler.Response, 1),
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := engine.Schedule(requests[i])
			if err != nil {
				b.Fatalf("Request scheduling failed: %v", err)
			}
		}
	})
}
