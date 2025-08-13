//go:build ignore

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// TestDistributedInference tests distributed inference across multiple nodes
func TestDistributedInference(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test cluster
	cluster, err := NewTestCluster(3)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	t.Run("TestBasicInference", func(t *testing.T) {
		// Create inference request
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "What is the capital of France?",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  50,
			},
		}

		// Send request to cluster
		response, err := cluster.GetLeader().ProcessInference(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Response)
		assert.True(t, response.Done)
	})

	t.Run("TestLayerwiseDistribution", func(t *testing.T) {
		// Create request that should trigger layerwise distribution
		req := &api.InferenceRequest{
			Model:  "llama3.2:8b", // Larger model
			Prompt: "Write a detailed explanation of quantum computing",
			Options: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  500,
			},
		}

		// Send request
		response, err := cluster.GetLeader().ProcessInference(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Response)

		// Verify distribution occurred
		tasks := cluster.GetLeader().GetActiveTasks()
		assert.Greater(t, len(tasks), 0)

		// Check that multiple nodes were used
		usedNodes := make(map[string]bool)
		for _, task := range tasks {
			for _, node := range task.Nodes {
				usedNodes[node.ID] = true
			}
		}
		assert.Greater(t, len(usedNodes), 1)
	})

	t.Run("TestLoadBalancing", func(t *testing.T) {
		// Send multiple requests to test load balancing
		requests := make([]*api.InferenceRequest, 10)
		for i := 0; i < 10; i++ {
			requests[i] = &api.InferenceRequest{
				Model:  "llama3.2:1b",
				Prompt: fmt.Sprintf("Question %d: What is 2+2?", i),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  20,
				},
			}
		}

		// Send requests concurrently
		responses := make(chan *api.InferenceResponse, len(requests))
		errors := make(chan error, len(requests))

		for _, req := range requests {
			go func(r *api.InferenceRequest) {
				resp, err := cluster.GetLeader().ProcessInference(context.Background(), r)
				if err != nil {
					errors <- err
				} else {
					responses <- resp
				}
			}(req)
		}

		// Collect responses
		successCount := 0
		errorCount := 0
		timeout := time.After(60 * time.Second)

		for i := 0; i < len(requests); i++ {
			select {
			case resp := <-responses:
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Response)
				successCount++
			case err := <-errors:
				t.Logf("Request error: %v", err)
				errorCount++
			case <-timeout:
				t.Fatal("Timeout waiting for responses")
			}
		}

		// Check results
		assert.Equal(t, len(requests), successCount+errorCount)
		assert.Greater(t, successCount, len(requests)/2) // At least 50% success rate

		// Check load distribution
		nodeLoads := cluster.GetNodeLoads()
		assert.Greater(t, len(nodeLoads), 1)

		// Verify load was distributed
		totalLoad := 0
		for _, load := range nodeLoads {
			totalLoad += load
		}
		assert.Greater(t, totalLoad, 0)
	})

	t.Run("TestModelSync", func(t *testing.T) {
		// Get cluster leader
		leader := cluster.GetLeader()

		// Register a new model on leader
		modelInfo := &distributed.ModelInfo{
			Name:              "test-sync-model",
			Path:              "/tmp/test-model",
			Size:              1024 * 1024 * 1024, // 1GB
			Checksum:          "test-checksum-123",
			ReplicationFactor: 2,
			Locations:         []string{leader.GetID()},
			LastAccessed:      time.Now(),
			Popularity:        0.5,
		}

		err := leader.RegisterModel(modelInfo)
		require.NoError(t, err)

		// Wait for model to sync to other nodes
		time.Sleep(30 * time.Second)

		// Verify model is available on multiple nodes
		replicaCount := 0
		for _, node := range cluster.GetNodes() {
			if node.HasModel(modelInfo.Name) {
				replicaCount++
			}
		}

		assert.GreaterOrEqual(t, replicaCount, modelInfo.ReplicationFactor)
	})
}

// TestDistributedFailover tests failover scenarios
func TestDistributedFailover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test cluster
	cluster, err := NewTestCluster(5)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(15 * time.Second)

	t.Run("TestNodeFailure", func(t *testing.T) {
		// Get initial cluster state
		initialNodes := cluster.GetActiveNodes()
		assert.Equal(t, 5, len(initialNodes))

		// Simulate node failure
		failedNode := cluster.GetNodes()[2] // Fail a non-leader node
		err := failedNode.Shutdown()
		require.NoError(t, err)

		// Wait for failure detection
		time.Sleep(30 * time.Second)

		// Verify cluster is still operational
		remainingNodes := cluster.GetActiveNodes()
		assert.Equal(t, 4, len(remainingNodes))

		// Test inference still works
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "Test after node failure",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		response, err := cluster.GetLeader().ProcessInference(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Response)
	})

	t.Run("TestLeaderFailover", func(t *testing.T) {
		// Get current leader
		oldLeader := cluster.GetLeader()
		oldLeaderID := oldLeader.GetID()

		// Simulate leader failure
		err := oldLeader.Shutdown()
		require.NoError(t, err)

		// Wait for leader election
		time.Sleep(45 * time.Second)

		// Verify new leader is elected
		newLeader := cluster.GetLeader()
		assert.NotNil(t, newLeader)
		assert.NotEqual(t, oldLeaderID, newLeader.GetID())

		// Test inference with new leader
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "Test after leader failover",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		response, err := newLeader.ProcessInference(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Response)
	})

	t.Run("TestPartialFailure", func(t *testing.T) {
		// Create inference request that will be distributed
		req := &api.InferenceRequest{
			Model:  "llama3.2:8b",
			Prompt: "Complex task requiring multiple nodes",
			Options: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  200,
			},
		}

		// Start inference
		responseChan := make(chan *api.InferenceResponse, 1)
		errorChan := make(chan error, 1)

		go func() {
			resp, err := cluster.GetLeader().ProcessInference(context.Background(), req)
			if err != nil {
				errorChan <- err
			} else {
				responseChan <- resp
			}
		}()

		// Wait a bit then simulate partial failure
		time.Sleep(5 * time.Second)

		// Fail one more node during processing
		if len(cluster.GetActiveNodes()) > 2 {
			failedNode := cluster.GetActiveNodes()[1]
			err := failedNode.Shutdown()
			require.NoError(t, err)
		}

		// Wait for response
		select {
		case response := <-responseChan:
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.Response)
		case err := <-errorChan:
			t.Logf("Inference failed after partial failure: %v", err)
			// This is acceptable - the system detected the failure
		case <-time.After(120 * time.Second):
			t.Fatal("Timeout waiting for inference response")
		}
	})
}

// TestDistributedConsensus tests consensus mechanisms
func TestDistributedConsensus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test cluster
	cluster, err := NewTestCluster(3)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	t.Run("TestConsensusOperations", func(t *testing.T) {
		leader := cluster.GetLeader()
		assert.NotNil(t, leader)

		// Apply a series of consensus operations
		operations := []string{
			"operation1",
			"operation2",
			"operation3",
		}

		for i, op := range operations {
			key := fmt.Sprintf("test_key_%d", i)
			err := leader.ApplyConsensusOperation(key, op)
			require.NoError(t, err)
		}

		// Wait for replication
		time.Sleep(5 * time.Second)

		// Verify all nodes have the same state
		for i, op := range operations {
			key := fmt.Sprintf("test_key_%d", i)

			for _, node := range cluster.GetNodes() {
				if !node.IsActive() {
					continue
				}

				value, exists := node.GetConsensusValue(key)
				assert.True(t, exists, "Node %s should have key %s", node.GetID(), key)
				assert.Equal(t, op, value, "Node %s should have correct value for key %s", node.GetID(), key)
			}
		}
	})

	t.Run("TestConsensusConsistency", func(t *testing.T) {
		// Concurrent consensus operations
		numOperations := 50
		done := make(chan bool, numOperations)
		errors := make(chan error, numOperations)

		for i := 0; i < numOperations; i++ {
			go func(idx int) {
				key := fmt.Sprintf("concurrent_key_%d", idx)
				value := fmt.Sprintf("concurrent_value_%d", idx)

				err := cluster.GetLeader().ApplyConsensusOperation(key, value)
				if err != nil {
					errors <- err
				} else {
					done <- true
				}
			}(i)
		}

		// Wait for all operations
		successCount := 0
		errorCount := 0
		timeout := time.After(60 * time.Second)

		for i := 0; i < numOperations; i++ {
			select {
			case <-done:
				successCount++
			case err := <-errors:
				t.Logf("Consensus operation error: %v", err)
				errorCount++
			case <-timeout:
				t.Fatal("Timeout waiting for consensus operations")
			}
		}

		assert.Equal(t, numOperations, successCount+errorCount)
		assert.Greater(t, successCount, numOperations*4/5) // At least 80% success rate

		// Wait for replication
		time.Sleep(10 * time.Second)

		// Verify consistency across all nodes
		for i := 0; i < successCount; i++ {
			key := fmt.Sprintf("concurrent_key_%d", i)
			expectedValue := fmt.Sprintf("concurrent_value_%d", i)

			values := make(map[string]int)
			for _, node := range cluster.GetActiveNodes() {
				if value, exists := node.GetConsensusValue(key); exists {
					values[value]++
				}
			}

			// All nodes should have the same value
			assert.LessOrEqual(t, len(values), 1, "Inconsistent values for key %s: %v", key, values)

			if len(values) == 1 {
				for value := range values {
					assert.Equal(t, expectedValue, value, "Incorrect value for key %s", key)
				}
			}
		}
	})
}

// TestDistributedModelManagement tests model management across the cluster
func TestDistributedModelManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test cluster
	cluster, err := NewTestCluster(4)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	t.Run("TestModelReplication", func(t *testing.T) {
		// Create test model
		modelInfo := &distributed.ModelInfo{
			Name:              "replication-test-model",
			Path:              "/tmp/replication-test",
			Size:              512 * 1024 * 1024, // 512MB
			Checksum:          "replication-checksum-abc",
			ReplicationFactor: 3,
			Locations:         []string{cluster.GetLeader().GetID()},
			LastAccessed:      time.Now(),
			Popularity:        0.8,
		}

		// Register model
		err := cluster.GetLeader().RegisterModel(modelInfo)
		require.NoError(t, err)

		// Wait for replication
		time.Sleep(60 * time.Second)

		// Verify replication
		replicaCount := 0
		for _, node := range cluster.GetActiveNodes() {
			if node.HasModel(modelInfo.Name) {
				replicaCount++
			}
		}

		assert.GreaterOrEqual(t, replicaCount, modelInfo.ReplicationFactor)
	})

	t.Run("TestModelEviction", func(t *testing.T) {
		// Fill up nodes with models
		for i := 0; i < 10; i++ {
			modelInfo := &distributed.ModelInfo{
				Name:              fmt.Sprintf("eviction-test-model-%d", i),
				Size:              100 * 1024 * 1024, // 100MB
				Checksum:          fmt.Sprintf("eviction-checksum-%d", i),
				ReplicationFactor: 2,
				Locations:         []string{cluster.GetLeader().GetID()},
				LastAccessed:      time.Now().Add(-time.Duration(i) * time.Hour),
				Popularity:        1.0 / float64(i+1),
			}

			err := cluster.GetLeader().RegisterModel(modelInfo)
			require.NoError(t, err)
		}

		// Wait for replication and potential eviction
		time.Sleep(30 * time.Second)

		// Verify eviction policy worked
		// Less popular models should be evicted first
		nodes := cluster.GetActiveNodes()
		for _, node := range nodes {
			models := node.GetModels()

			// Check that popular models are still there
			hasPopularModel := false
			for _, model := range models {
				if model.Name == "eviction-test-model-0" { // Most popular
					hasPopularModel = true
					break
				}
			}

			if len(models) > 0 {
				assert.True(t, hasPopularModel, "Node %s should retain popular models", node.GetID())
			}
		}
	})

	t.Run("TestModelVersioning", func(t *testing.T) {
		// Create initial model version
		modelInfo := &distributed.ModelInfo{
			Name:              "versioning-test-model",
			Version:           "1.0.0",
			Size:              256 * 1024 * 1024, // 256MB
			Checksum:          "version-1-checksum",
			ReplicationFactor: 2,
			Locations:         []string{cluster.GetLeader().GetID()},
			LastAccessed:      time.Now(),
			Popularity:        0.6,
		}

		err := cluster.GetLeader().RegisterModel(modelInfo)
		require.NoError(t, err)

		// Wait for replication
		time.Sleep(20 * time.Second)

		// Update model to new version
		modelInfo.Version = "1.1.0"
		modelInfo.Checksum = "version-1-1-checksum"
		modelInfo.Size = 280 * 1024 * 1024 // 280MB

		err = cluster.GetLeader().UpdateModel(modelInfo)
		require.NoError(t, err)

		// Wait for update propagation
		time.Sleep(30 * time.Second)

		// Verify all nodes have the new version
		for _, node := range cluster.GetActiveNodes() {
			if node.HasModel(modelInfo.Name) {
				model := node.GetModel(modelInfo.Name)
				assert.Equal(t, "1.1.0", model.Version)
				assert.Equal(t, "version-1-1-checksum", model.Checksum)
			}
		}
	})
}

// TestDistributedPerformance tests performance characteristics
func TestDistributedPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test cluster
	cluster, err := NewTestCluster(3)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	t.Run("TestThroughput", func(t *testing.T) {
		// Test concurrent inference requests
		numRequests := 100
		concurrency := 10

		requests := make(chan *api.InferenceRequest, numRequests)
		responses := make(chan *api.InferenceResponse, numRequests)
		errors := make(chan error, numRequests)

		// Generate requests
		for i := 0; i < numRequests; i++ {
			requests <- &api.InferenceRequest{
				Model:  "llama3.2:1b",
				Prompt: fmt.Sprintf("Performance test request %d", i),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  50,
				},
			}
		}
		close(requests)

		// Process requests concurrently
		startTime := time.Now()

		for i := 0; i < concurrency; i++ {
			go func() {
				for req := range requests {
					resp, err := cluster.GetLeader().ProcessInference(context.Background(), req)
					if err != nil {
						errors <- err
					} else {
						responses <- resp
					}
				}
			}()
		}

		// Collect results
		successCount := 0
		errorCount := 0
		timeout := time.After(300 * time.Second) // 5 minutes

		for i := 0; i < numRequests; i++ {
			select {
			case resp := <-responses:
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.Response)
				successCount++
			case err := <-errors:
				t.Logf("Request error: %v", err)
				errorCount++
			case <-timeout:
				t.Fatal("Timeout waiting for responses")
			}
		}

		duration := time.Since(startTime)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Processed %d requests in %v", successCount, duration)
		t.Logf("Throughput: %.2f requests/second", throughput)
		t.Logf("Success rate: %.2f%%", float64(successCount)/float64(numRequests)*100)

		// Verify performance
		assert.Greater(t, successCount, numRequests*4/5) // At least 80% success rate
		assert.Greater(t, throughput, 1.0)               // At least 1 request per second
	})

	t.Run("TestLatency", func(t *testing.T) {
		// Test latency for different request sizes
		testCases := []struct {
			name      string
			maxTokens int
			expected  time.Duration
		}{
			{"Small", 20, 5 * time.Second},
			{"Medium", 100, 15 * time.Second},
			{"Large", 500, 60 * time.Second},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &api.InferenceRequest{
					Model:  "llama3.2:1b",
					Prompt: "Test latency for " + tc.name + " response",
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  tc.maxTokens,
					},
				}

				startTime := time.Now()
				response, err := cluster.GetLeader().ProcessInference(context.Background(), req)
				latency := time.Since(startTime)

				require.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Response)

				t.Logf("%s request latency: %v", tc.name, latency)
				assert.Less(t, latency, tc.expected)
			})
		}
	})
}

// BenchmarkDistributedInference benchmarks distributed inference performance
func BenchmarkDistributedInference(b *testing.B) {
	// Create test cluster
	cluster, err := NewTestCluster(3)
	require.NoError(b, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(b, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	b.ResetTimer()

	b.Run("SingleRequest", func(b *testing.B) {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "Benchmark test",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := cluster.GetLeader().ProcessInference(context.Background(), req)
				if err != nil {
					b.Errorf("Inference failed: %v", err)
				}
			}
		})
	})

	b.Run("ConcurrentRequests", func(b *testing.B) {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "Concurrent benchmark test",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  50,
			},
		}

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := cluster.GetLeader().ProcessInference(context.Background(), req)
				if err != nil {
					b.Errorf("Inference failed: %v", err)
				}
			}
		})
	})
}
