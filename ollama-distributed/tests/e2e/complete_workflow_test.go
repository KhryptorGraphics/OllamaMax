package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ollama/ollama-distributed/tests/integration"
)

// TestCompleteWorkflow tests a complete end-to-end workflow
func TestCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create test cluster
	cluster, err := integration.NewTestCluster(3)
	require.NoError(t, err)
	defer cluster.Shutdown()

	// Start cluster
	err = cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(15 * time.Second)

	t.Run("CompleteInferenceWorkflow", func(t *testing.T) {
		testCompleteInferenceWorkflow(t, cluster)
	})

	t.Run("ModelManagementWorkflow", func(t *testing.T) {
		testModelManagementWorkflow(t, cluster)
	})

	t.Run("FaultToleranceWorkflow", func(t *testing.T) {
		testFaultToleranceWorkflow(t, cluster)
	})

	t.Run("ScalingWorkflow", func(t *testing.T) {
		testScalingWorkflow(t, cluster)
	})
}

// testCompleteInferenceWorkflow tests a complete inference workflow
func testCompleteInferenceWorkflow(t *testing.T, cluster *integration.TestCluster) {
	// Step 1: Verify cluster is healthy
	leader := cluster.GetLeader()
	require.NotNil(t, leader, "Cluster should have a leader")

	activeNodes := cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(activeNodes), 2, "Should have at least 2 active nodes")

	// Step 2: Test API health endpoint
	apiEndpoint := fmt.Sprintf("http://%s/health", leader.GetAPIAddress())
	resp, err := http.Get(apiEndpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 3: Test model listing
	modelsEndpoint := fmt.Sprintf("http://%s/api/tags", leader.GetAPIAddress())
	resp, err = http.Get(modelsEndpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()
	
	assert.Contains(t, string(body), "models")

	// Step 4: Test simple inference
	inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", leader.GetAPIAddress())
	requestBody := `{
		"model": "llama3.2:1b",
		"prompt": "What is the capital of France?",
		"stream": false,
		"options": {
			"temperature": 0.1,
			"max_tokens": 50
		}
	}`

	resp, err = http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Contains(t, string(body), "response")
	assert.Contains(t, string(body), "done")

	// Step 5: Test streaming inference
	streamRequestBody := `{
		"model": "llama3.2:1b",
		"prompt": "Tell me a short story",
		"stream": true,
		"options": {
			"temperature": 0.7,
			"max_tokens": 200
		}
	}`

	resp, err = http.Post(inferenceEndpoint, "application/json", strings.NewReader(streamRequestBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Read streaming response
	streamData, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.NotEmpty(t, streamData)
	assert.Contains(t, string(streamData), "response")

	// Step 6: Test distributed inference (large model)
	largeModelRequest := `{
		"model": "llama3.2:8b",
		"prompt": "Explain quantum computing in detail",
		"stream": false,
		"options": {
			"temperature": 0.5,
			"max_tokens": 300
		}
	}`

	resp, err = http.Post(inferenceEndpoint, "application/json", strings.NewReader(largeModelRequest))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Contains(t, string(body), "response")
	assert.Contains(t, string(body), "done")

	// Step 7: Verify load distribution
	nodeLoads := cluster.GetNodeLoads()
	assert.Greater(t, len(nodeLoads), 1, "Load should be distributed across nodes")

	totalLoad := 0
	for _, load := range nodeLoads {
		totalLoad += load
	}
	assert.Greater(t, totalLoad, 0, "Should have some load on the cluster")

	// Step 8: Test metrics endpoint
	metricsEndpoint := fmt.Sprintf("http://%s/api/v1/metrics", leader.GetAPIAddress())
	resp, err = http.Get(metricsEndpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Contains(t, string(body), "metrics")
}

// testModelManagementWorkflow tests model management workflow
func testModelManagementWorkflow(t *testing.T, cluster *integration.TestCluster) {
	leader := cluster.GetLeader()
	require.NotNil(t, leader)

	// Step 1: Test model upload simulation
	testModelPath := filepath.Join(os.TempDir(), "test-model.bin")
	err := os.WriteFile(testModelPath, []byte("mock model data"), 0644)
	require.NoError(t, err)
	defer os.Remove(testModelPath)

	// Step 2: Test model registration
	modelInfo := &integration.ModelInfo{
		Name:              "test-e2e-model",
		Path:              testModelPath,
		Size:              1024 * 1024 * 1024, // 1GB
		Checksum:          "e2e-test-checksum",
		ReplicationFactor: 2,
		LastAccessed:      time.Now(),
		Popularity:        0.6,
	}

	err = leader.RegisterModel(modelInfo)
	require.NoError(t, err)

	// Step 3: Wait for model replication
	time.Sleep(30 * time.Second)

	// Step 4: Verify model is available on multiple nodes
	replicaCount := 0
	for _, node := range cluster.GetActiveNodes() {
		if node.HasModel(modelInfo.Name) {
			replicaCount++
		}
	}
	assert.GreaterOrEqual(t, replicaCount, modelInfo.ReplicationFactor)

	// Step 5: Test model update
	modelInfo.Version = "1.1.0"
	modelInfo.Checksum = "e2e-test-checksum-updated"

	err = leader.UpdateModel(modelInfo)
	require.NoError(t, err)

	// Step 6: Wait for update propagation
	time.Sleep(20 * time.Second)

	// Step 7: Verify all replicas have the updated version
	for _, node := range cluster.GetActiveNodes() {
		if node.HasModel(modelInfo.Name) {
			model := node.GetModel(modelInfo.Name)
			assert.Equal(t, "1.1.0", model.Version)
			assert.Equal(t, "e2e-test-checksum-updated", model.Checksum)
		}
	}

	// Step 8: Test model usage tracking
	// Use the model in an inference request
	inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", leader.GetAPIAddress())
	requestBody := fmt.Sprintf(`{
		"model": "%s",
		"prompt": "Test model usage",
		"stream": false,
		"options": {
			"temperature": 0.1,
			"max_tokens": 20
		}
	}`, modelInfo.Name)

	resp, err := http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 9: Verify usage statistics are updated
	time.Sleep(5 * time.Second)
	
	for _, node := range cluster.GetActiveNodes() {
		if node.HasModel(modelInfo.Name) {
			model := node.GetModel(modelInfo.Name)
			assert.True(t, model.LastAccessed.After(modelInfo.LastAccessed))
		}
	}
}

// testFaultToleranceWorkflow tests fault tolerance workflow
func testFaultToleranceWorkflow(t *testing.T, cluster *integration.TestCluster) {
	// Step 1: Verify initial cluster state
	initialNodes := cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(initialNodes), 3, "Need at least 3 nodes for fault tolerance test")

	initialLeader := cluster.GetLeader()
	require.NotNil(t, initialLeader)

	// Step 2: Start a long-running inference task
	inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", initialLeader.GetAPIAddress())
	requestBody := `{
		"model": "llama3.2:8b",
		"prompt": "Write a long story about fault tolerance in distributed systems",
		"stream": false,
		"options": {
			"temperature": 0.7,
			"max_tokens": 500
		}
	}`

	// Start inference in background
	responseChan := make(chan *http.Response, 1)
	errorChan := make(chan error, 1)

	go func() {
		resp, err := http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
		if err != nil {
			errorChan <- err
		} else {
			responseChan <- resp
		}
	}()

	// Step 3: Wait a bit then simulate node failure
	time.Sleep(5 * time.Second)

	// Find a non-leader node to fail
	var nodeToFail *integration.TestNode
	for _, node := range initialNodes {
		if !node.IsLeader() {
			nodeToFail = node
			break
		}
	}
	require.NotNil(t, nodeToFail, "Should have a non-leader node to fail")

	// Fail the node
	err := nodeToFail.Shutdown()
	require.NoError(t, err)

	// Step 4: Wait for failure detection and recovery
	time.Sleep(20 * time.Second)

	// Step 5: Verify cluster is still operational
	remainingNodes := cluster.GetActiveNodes()
	assert.Equal(t, len(initialNodes)-1, len(remainingNodes))

	currentLeader := cluster.GetLeader()
	assert.NotNil(t, currentLeader)

	// Step 6: Test new inference request
	testEndpoint := fmt.Sprintf("http://%s/api/generate", currentLeader.GetAPIAddress())
	testRequest := `{
		"model": "llama3.2:1b",
		"prompt": "Test after node failure",
		"stream": false,
		"options": {
			"temperature": 0.1,
			"max_tokens": 50
		}
	}`

	resp, err := http.Post(testEndpoint, "application/json", strings.NewReader(testRequest))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Step 7: Check if original inference completed or failed gracefully
	select {
	case resp := <-responseChan:
		// If it completed, verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Contains(t, string(body), "response")
		t.Log("Original inference completed successfully despite node failure")
	case err := <-errorChan:
		// If it failed, that's also acceptable
		t.Logf("Original inference failed after node failure: %v", err)
	case <-time.After(60 * time.Second):
		t.Log("Original inference timed out (acceptable)")
	}

	// Step 8: Test leader failure
	if len(remainingNodes) >= 3 {
		// Fail the current leader
		err := currentLeader.Shutdown()
		require.NoError(t, err)

		// Wait for leader election
		time.Sleep(30 * time.Second)

		// Verify new leader is elected
		newLeader := cluster.GetLeader()
		assert.NotNil(t, newLeader)
		assert.NotEqual(t, currentLeader.GetID(), newLeader.GetID())

		// Test inference with new leader
		newLeaderEndpoint := fmt.Sprintf("http://%s/api/generate", newLeader.GetAPIAddress())
		leaderTestRequest := `{
			"model": "llama3.2:1b",
			"prompt": "Test after leader failure",
			"stream": false,
			"options": {
				"temperature": 0.1,
				"max_tokens": 50
			}
		}`

		resp, err := http.Post(newLeaderEndpoint, "application/json", strings.NewReader(leaderTestRequest))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
}

// testScalingWorkflow tests scaling workflow
func testScalingWorkflow(t *testing.T, cluster *integration.TestCluster) {
	// Step 1: Measure baseline performance
	leader := cluster.GetLeader()
	require.NotNil(t, leader)

	// Send multiple requests to establish baseline
	baselineRequests := 10
	baselineStart := time.Now()
	
	for i := 0; i < baselineRequests; i++ {
		requestBody := fmt.Sprintf(`{
			"model": "llama3.2:1b",
			"prompt": "Baseline request %d",
			"stream": false,
			"options": {
				"temperature": 0.1,
				"max_tokens": 50
			}
		}`, i)

		inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", leader.GetAPIAddress())
		resp, err := http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	baselineDuration := time.Since(baselineStart)
	baselineThroughput := float64(baselineRequests) / baselineDuration.Seconds()

	t.Logf("Baseline performance: %d requests in %v (%.2f req/s)", 
		baselineRequests, baselineDuration, baselineThroughput)

	// Step 2: Test high load scenario
	highLoadRequests := 50
	concurrency := 10
	
	requests := make(chan int, highLoadRequests)
	responses := make(chan bool, highLoadRequests)
	errors := make(chan error, highLoadRequests)

	// Generate request numbers
	for i := 0; i < highLoadRequests; i++ {
		requests <- i
	}
	close(requests)

	// Start high load test
	highLoadStart := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			for reqNum := range requests {
				requestBody := fmt.Sprintf(`{
					"model": "llama3.2:1b",
					"prompt": "High load request %d",
					"stream": false,
					"options": {
						"temperature": 0.1,
						"max_tokens": 50
					}
				}`, reqNum)

				inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", leader.GetAPIAddress())
				resp, err := http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
				
				if err != nil {
					errors <- err
				} else {
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						responses <- true
					} else {
						errors <- fmt.Errorf("HTTP %d", resp.StatusCode)
					}
				}
			}
		}()
	}

	// Collect results
	successCount := 0
	errorCount := 0
	timeout := time.After(300 * time.Second) // 5 minutes

	for i := 0; i < highLoadRequests; i++ {
		select {
		case <-responses:
			successCount++
		case err := <-errors:
			t.Logf("High load request error: %v", err)
			errorCount++
		case <-timeout:
			t.Fatal("Timeout waiting for high load responses")
		}
	}

	highLoadDuration := time.Since(highLoadStart)
	highLoadThroughput := float64(successCount) / highLoadDuration.Seconds()

	t.Logf("High load performance: %d/%d requests succeeded in %v (%.2f req/s)", 
		successCount, highLoadRequests, highLoadDuration, highLoadThroughput)

	// Step 3: Verify performance characteristics
	assert.Greater(t, successCount, highLoadRequests*7/10, "Should have at least 70% success rate under high load")
	assert.Greater(t, highLoadThroughput, baselineThroughput*0.5, "High load throughput should not be less than 50% of baseline")

	// Step 4: Test resource utilization
	nodeLoads := cluster.GetNodeLoads()
	assert.Greater(t, len(nodeLoads), 1, "Load should be distributed across nodes")

	// Verify load is reasonably distributed
	if len(nodeLoads) > 1 {
		totalLoad := 0
		maxLoad := 0
		for _, load := range nodeLoads {
			totalLoad += load
			if load > maxLoad {
				maxLoad = load
			}
		}

		if totalLoad > 0 {
			maxLoadRatio := float64(maxLoad) / float64(totalLoad)
			assert.Less(t, maxLoadRatio, 0.8, "No single node should handle more than 80% of the load")
		}
	}

	// Step 5: Test recovery after high load
	time.Sleep(30 * time.Second)

	// Send a few more requests to verify system recovered
	recoveryRequests := 5
	for i := 0; i < recoveryRequests; i++ {
		requestBody := fmt.Sprintf(`{
			"model": "llama3.2:1b",
			"prompt": "Recovery request %d",
			"stream": false,
			"options": {
				"temperature": 0.1,
				"max_tokens": 50
			}
		}`, i)

		inferenceEndpoint := fmt.Sprintf("http://%s/api/generate", leader.GetAPIAddress())
		resp, err := http.Post(inferenceEndpoint, "application/json", strings.NewReader(requestBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	t.Log("System recovered successfully after high load test")
}

// Helper method to get API address from a test node
func (tn *integration.TestNode) GetAPIAddress() string {
	return tn.GetConfig().API.Listen
}

// Helper method to get node configuration
func (tn *integration.TestNode) GetConfig() *config.Config {
	return tn.config
}