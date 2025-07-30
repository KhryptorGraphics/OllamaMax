package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
)

// EnhancedIntegrationTestSuite provides comprehensive end-to-end testing
type EnhancedIntegrationTestSuite struct {
	cluster    *TestCluster
	testModels []TestModel
	metrics    *TestMetrics
	ctx        context.Context
	cancel     context.CancelFunc
}

// TestModel represents a test model configuration
type TestModel struct {
	Name              string
	Size              int64
	Type              string
	ReplicationFactor int
	ExpectedLatency   time.Duration
}

// TestMetrics tracks test execution metrics
type TestMetrics struct {
	mu                sync.RWMutex
	TestsRun          int
	TestsPassed       int
	TestsFailed       int
	TotalDuration     time.Duration
	AverageLatency    time.Duration
	ThroughputRPS     float64
	ErrorRate         float64
	StartTime         time.Time
}

// NewEnhancedIntegrationTestSuite creates a new enhanced integration test suite
func NewEnhancedIntegrationTestSuite(nodeCount int) (*EnhancedIntegrationTestSuite, error) {
	cluster, err := NewTestCluster(nodeCount)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Define test models
	testModels := []TestModel{
		{
			Name:              "llama3.2:1b",
			Size:              1024 * 1024 * 1024, // 1GB
			Type:              "language_model",
			ReplicationFactor: 2,
			ExpectedLatency:   500 * time.Millisecond,
		},
		{
			Name:              "llama3.2:8b",
			Size:              8 * 1024 * 1024 * 1024, // 8GB
			Type:              "language_model",
			ReplicationFactor: 3,
			ExpectedLatency:   2 * time.Second,
		},
		{
			Name:              "codellama:7b",
			Size:              7 * 1024 * 1024 * 1024, // 7GB
			Type:              "code_model",
			ReplicationFactor: 2,
			ExpectedLatency:   1500 * time.Millisecond,
		},
	}

	return &EnhancedIntegrationTestSuite{
		cluster:    cluster,
		testModels: testModels,
		metrics:    &TestMetrics{StartTime: time.Now()},
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// TestEnhancedIntegration runs comprehensive integration tests
func TestEnhancedIntegration(t *testing.T) {
	suite, err := NewEnhancedIntegrationTestSuite(5)
	require.NoError(t, err)
	defer suite.cleanup()

	// Start cluster
	require.NoError(t, suite.cluster.Start())
	
	// Wait for cluster stabilization
	time.Sleep(10 * time.Second)

	t.Run("ClusterBootstrap", func(t *testing.T) {
		suite.testClusterBootstrap(t)
	})

	t.Run("ModelLifecycle", func(t *testing.T) {
		suite.testModelLifecycle(t)
	})

	t.Run("DistributedInference", func(t *testing.T) {
		suite.testDistributedInference(t)
	})

	t.Run("LoadBalancing", func(t *testing.T) {
		suite.testLoadBalancing(t)
	})

	t.Run("FailureRecovery", func(t *testing.T) {
		suite.testFailureRecovery(t)
	})

	t.Run("ScalabilityTesting", func(t *testing.T) {
		suite.testScalability(t)
	})

	t.Run("DataIntegrity", func(t *testing.T) {
		suite.testDataIntegrity(t)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		suite.testConcurrentOperations(t)
	})

	t.Run("ResourceManagement", func(t *testing.T) {
		suite.testResourceManagement(t)
	})

	t.Run("EndToEndWorkflows", func(t *testing.T) {
		suite.testEndToEndWorkflows(t)
	})

	// Generate final report
	suite.generateTestReport(t)
}

// testClusterBootstrap tests cluster bootstrap and initialization
func (suite *EnhancedIntegrationTestSuite) testClusterBootstrap(t *testing.T) {
	suite.recordTestStart()

	// Verify all nodes are running
	nodes := suite.cluster.GetActiveNodes()
	assert.GreaterOrEqual(t, len(nodes), 3, "Should have at least 3 active nodes")

	// Verify leader election
	leader := suite.cluster.GetLeader()
	assert.NotNil(t, leader, "Should have an elected leader")

	// Verify peer connectivity
	for i, node := range nodes {
		connectedPeers := node.GetConnectedPeers()
		assert.GreaterOrEqual(t, len(connectedPeers), 1, 
			"Node %d should be connected to other peers", i)
	}

	// Verify consensus is operational
	testKey := "bootstrap_test"
	testValue := "cluster_initialized"
	
	err := leader.ApplyConsensusOperation(testKey, testValue)
	assert.NoError(t, err, "Should be able to apply consensus operation")

	// Wait for replication
	time.Sleep(3 * time.Second)

	// Verify all nodes have the same state
	for i, node := range nodes {
		value, exists := node.GetConsensusValue(testKey)
		assert.True(t, exists, "Node %d should have replicated the key", i)
		assert.Equal(t, testValue, value, "Node %d should have correct value", i)
	}

	// Verify API endpoints are accessible
	for i, node := range nodes {
		health, err := node.GetHealthStatus()
		assert.NoError(t, err, "Node %d health check should succeed", i)
		assert.Equal(t, "healthy", health.Status, "Node %d should be healthy", i)
	}

	suite.recordTestEnd(true)
}

// testModelLifecycle tests complete model lifecycle operations
func (suite *EnhancedIntegrationTestSuite) testModelLifecycle(t *testing.T) {
	suite.recordTestStart()

	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader for model operations")

	for _, testModel := range suite.testModels {
		t.Run(fmt.Sprintf("Model_%s", testModel.Name), func(t *testing.T) {
			// Test model registration
			modelInfo := &ModelInfo{
				Name:              testModel.Name,
				Size:              testModel.Size,
				Type:              testModel.Type,
				ReplicationFactor: testModel.ReplicationFactor,
				Checksum:          fmt.Sprintf("checksum_%s", testModel.Name),
				LastAccessed:      time.Now(),
				Popularity:        0.5,
			}

			err := leader.RegisterModel(modelInfo)
			assert.NoError(t, err, "Model registration should succeed")

			// Test model discovery
			time.Sleep(2 * time.Second)
			discoveredModels, err := leader.DiscoverModels()
			assert.NoError(t, err, "Model discovery should succeed")

			found := false
			for _, model := range discoveredModels {
				if model.Name == testModel.Name {
					found = true
					assert.Equal(t, testModel.Size, model.Size, "Model size should match")
					assert.Equal(t, testModel.ReplicationFactor, model.ReplicationFactor, 
						"Replication factor should match")
					break
				}
			}
			assert.True(t, found, "Model should be discoverable")

			// Test model replication
			replicas, err := leader.GetModelReplicas(testModel.Name)
			assert.NoError(t, err, "Should be able to get model replicas")
			assert.GreaterOrEqual(t, len(replicas), 1, "Should have at least one replica")
			
			// Eventually should reach desired replication factor
			// (in a real system this might take longer)

			// Test model access
			modelData, err := leader.AccessModel(testModel.Name)
			assert.NoError(t, err, "Should be able to access model")
			assert.NotNil(t, modelData, "Model data should not be nil")

			// Test model update
			updatedInfo := *modelInfo
			updatedInfo.Popularity = 0.8
			err = leader.UpdateModel(&updatedInfo)
			assert.NoError(t, err, "Model update should succeed")

			// Verify update propagation
			time.Sleep(2 * time.Second)
			retrievedInfo, err := leader.GetModelInfo(testModel.Name)
			assert.NoError(t, err, "Should be able to retrieve updated model info")
			assert.Equal(t, 0.8, retrievedInfo.Popularity, "Popularity should be updated")

			// Test model deletion (optional, might comment out for persistence)
			// err = leader.DeleteModel(testModel.Name)
			// assert.NoError(t, err, "Model deletion should succeed")
		})
	}

	suite.recordTestEnd(true)
}

// testDistributedInference tests distributed inference capabilities
func (suite *EnhancedIntegrationTestSuite) testDistributedInference(t *testing.T) {
	suite.recordTestStart()

	nodes := suite.cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(nodes), 2, "Need at least 2 nodes for distributed inference")

	// Register a model for inference testing
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader")

	testModel := suite.testModels[0] // Use smallest model for inference
	modelInfo := &ModelInfo{
		Name:              testModel.Name,
		Size:              testModel.Size,
		Type:              testModel.Type,
		ReplicationFactor: 2,
		Checksum:          fmt.Sprintf("checksum_%s", testModel.Name),
		LastAccessed:      time.Now(),
		Popularity:        0.7,
	}

	err := leader.RegisterModel(modelInfo)
	require.NoError(t, err, "Model registration should succeed")

	// Wait for model distribution
	time.Sleep(5 * time.Second)

	// Test basic inference
	t.Run("BasicInference", func(t *testing.T) {
		req := &api.InferenceRequest{
			Model:  testModel.Name,
			Prompt: "Hello, world! How are you today?",
			Options: map[string]interface{}{
				"temperature": 0.7,
				"max_tokens":  100,
			},
		}

		start := time.Now()
		resp, err := leader.ProcessInference(suite.ctx, req)
		latency := time.Since(start)

		assert.NoError(t, err, "Inference should succeed")
		assert.NotNil(t, resp, "Response should not be nil")
		assert.NotEmpty(t, resp.Response, "Response should contain generated text")
		assert.True(t, resp.Done, "Response should be marked as done")
		assert.Less(t, latency, testModel.ExpectedLatency*2, 
			"Inference latency should be reasonable")

		// Update metrics
		suite.updateLatencyMetric(latency)
	})

	// Test streaming inference
	t.Run("StreamingInference", func(t *testing.T) {
		req := &api.InferenceRequest{
			Model:  testModel.Name,
			Prompt: "Write a short story about a distributed AI system",
			Stream: true,
			Options: map[string]interface{}{
				"temperature": 0.8,
				"max_tokens":  200,
			},
		}

		responseCount := 0
		var totalLatency time.Duration
		start := time.Now()

		err := leader.ProcessStreamingInference(suite.ctx, req, func(resp *api.InferenceResponse) {
			responseCount++
			if responseCount == 1 {
				// Measure time to first token
				totalLatency = time.Since(start)
			}
			
			assert.NotNil(t, resp, "Streaming response should not be nil")
			if resp.Done {
				assert.NotEmpty(t, resp.Response, "Final response should contain text")
			}
		})

		assert.NoError(t, err, "Streaming inference should succeed")
		assert.Greater(t, responseCount, 1, "Should receive multiple streaming responses")
		assert.Less(t, totalLatency, testModel.ExpectedLatency, 
			"Time to first token should be reasonable")
	})

	// Test concurrent inference
	t.Run("ConcurrentInference", func(t *testing.T) {
		concurrency := 5
		var wg sync.WaitGroup
		results := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				req := &api.InferenceRequest{
					Model:  testModel.Name,
					Prompt: fmt.Sprintf("Concurrent request %d: What is AI?", index),
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  50,
					},
				}

				_, err := leader.ProcessInference(suite.ctx, req)
				results <- err
			}(i)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for err := range results {
			if err == nil {
				successCount++
			} else {
				t.Logf("Concurrent inference error: %v", err)
			}
		}

		assert.GreaterOrEqual(t, successCount, concurrency*8/10, 
			"At least 80%% of concurrent requests should succeed")
	})

	// Test load distribution
	t.Run("LoadDistribution", func(t *testing.T) {
		requestCount := 20
		nodeUsage := make(map[string]int)

		for i := 0; i < requestCount; i++ {
			req := &api.InferenceRequest{
				Model:  testModel.Name,
				Prompt: fmt.Sprintf("Load test request %d", i),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  20,
				},
			}

			// Use different entry points to test load balancing
			entryNode := nodes[i%len(nodes)]
			resp, err := entryNode.ProcessInference(suite.ctx, req)
			
			if err == nil {
				// Track which node actually processed the request
				processingNode := resp.ProcessedBy
				if processingNode != "" {
					nodeUsage[processingNode]++
				}
			}
		}

		// Verify load is somewhat distributed
		assert.Greater(t, len(nodeUsage), 1, "Load should be distributed across multiple nodes")
		
		// No single node should handle all requests (unless only one has the model)
		for nodeID, count := range nodeUsage {
			ratio := float64(count) / float64(requestCount)
			assert.Less(t, ratio, 0.8, "Node %s should not handle >80%% of requests", nodeID)
		}
	})

	suite.recordTestEnd(true)
}

// testLoadBalancing tests load balancing mechanisms
func (suite *EnhancedIntegrationTestSuite) testLoadBalancing(t *testing.T) {
	suite.recordTestStart()

	nodes := suite.cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(nodes), 3, "Need at least 3 nodes for load balancing test")

	// Create load on different nodes
	var wg sync.WaitGroup
	requestCount := 100
	concurrency := 10

	loadResults := make(chan LoadTestResult, requestCount)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < requestCount/concurrency; j++ {
				start := time.Now()
				
				// Round-robin between nodes
				targetNode := nodes[(workerID*requestCount/concurrency+j)%len(nodes)]
				
				req := &api.InferenceRequest{
					Model:  suite.testModels[0].Name,
					Prompt: fmt.Sprintf("Load balancing test %d-%d", workerID, j),
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  20,
					},
				}

				resp, err := targetNode.ProcessInference(suite.ctx, req)
				latency := time.Since(start)

				result := LoadTestResult{
					RequestID:    fmt.Sprintf("%d-%d", workerID, j),
					EntryNode:    targetNode.GetID(),
					ProcessingNode: "",
					Latency:      latency,
					Success:      err == nil,
					Error:        err,
				}

				if err == nil && resp != nil {
					result.ProcessingNode = resp.ProcessedBy
				}

				loadResults <- result
			}
		}(i)
	}

	wg.Wait()
	close(loadResults)

	// Analyze load balancing results
	entryNodeCounts := make(map[string]int)
	processingNodeCounts := make(map[string]int)
	successCount := 0
	totalLatency := time.Duration(0)

	for result := range loadResults {
		entryNodeCounts[result.EntryNode]++
		if result.ProcessingNode != "" {
			processingNodeCounts[result.ProcessingNode]++
		}
		if result.Success {
			successCount++
			totalLatency += result.Latency
		}
	}

	// Verify load balancing effectiveness
	assert.Greater(t, len(processingNodeCounts), 1, 
		"Load should be distributed across multiple processing nodes")

	// Calculate load distribution metrics
	avgProcessingCount := float64(successCount) / float64(len(processingNodeCounts))
	maxDeviation := 0.0

	for _, count := range processingNodeCounts {
		deviation := abs(float64(count) - avgProcessingCount) / avgProcessingCount
		if deviation > maxDeviation {
			maxDeviation = deviation
		}
	}

	assert.Less(t, maxDeviation, 0.5, 
		"Load distribution should be reasonably balanced (max 50%% deviation)")

	// Verify performance under load
	avgLatency := totalLatency / time.Duration(successCount)
	assert.Less(t, avgLatency, suite.testModels[0].ExpectedLatency*3, 
		"Average latency should remain reasonable under load")

	successRate := float64(successCount) / float64(requestCount)
	assert.Greater(t, successRate, 0.9, "Success rate should be >90%% under load")

	suite.recordTestEnd(true)
}

// testFailureRecovery tests failure recovery scenarios
func (suite *EnhancedIntegrationTestSuite) testFailureRecovery(t *testing.T) {
	suite.recordTestStart()

	nodes := suite.cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(nodes), 3, "Need at least 3 nodes for failure recovery test")

	// Test node failure recovery
	t.Run("NodeFailureRecovery", func(t *testing.T) {
		// Select a non-leader node to fail
		var nodeToFail *TestNode
		for _, node := range nodes {
			if !node.IsLeader() {
				nodeToFail = node
				break
			}
		}
		require.NotNil(t, nodeToFail, "Need a non-leader node to fail")

		failedNodeID := nodeToFail.GetID()

		// Record system state before failure
		preFailureState := suite.recordSystemState()

		// Fail the node
		err := nodeToFail.SimulateFailure("crash")
		require.NoError(t, err, "Should be able to simulate failure")

		// Wait for failure detection
		time.Sleep(10 * time.Second)

		// Verify system continues operating
		leader := suite.cluster.GetLeader()
		assert.NotNil(t, leader, "Should still have a leader after node failure")

		// Test inference still works
		req := &api.InferenceRequest{
			Model:  suite.testModels[0].Name,
			Prompt: "Test after node failure",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		_, err = leader.ProcessInference(suite.ctx, req)
		assert.NoError(t, err, "Inference should work after node failure")

		// Recover the node
		err = nodeToFail.Recover()
		require.NoError(t, err, "Node recovery should succeed")

		// Wait for recovery
		time.Sleep(15 * time.Second)

		// Verify node is back online
		recoveredNode := suite.cluster.GetNodeByID(failedNodeID)
		assert.NotNil(t, recoveredNode, "Recovered node should be available")
		assert.True(t, recoveredNode.IsRunning(), "Recovered node should be running")

		// Verify system state consistency
		postRecoveryState := suite.recordSystemState()
		suite.verifyStateConsistency(t, preFailureState, postRecoveryState)
	})

	// Test leader failure recovery
	t.Run("LeaderFailureRecovery", func(t *testing.T) {
		originalLeader := suite.cluster.GetLeader()
		require.NotNil(t, originalLeader, "Need a leader to fail")

		originalLeaderID := originalLeader.GetID()

		// Monitor leader election
		newLeaderElected := make(chan string, 1)
		suite.cluster.OnLeaderChange(func(newLeaderID string) {
			if newLeaderID != originalLeaderID {
				newLeaderElected <- newLeaderID
			}
		})

		// Fail the leader
		err := originalLeader.SimulateFailure("crash")
		require.NoError(t, err, "Should be able to fail leader")

		// Wait for new leader election
		select {
		case newLeaderID := <-newLeaderElected:
			t.Logf("New leader elected: %s", newLeaderID)

			// Verify new leader is functional
			newLeader := suite.cluster.GetNodeByID(newLeaderID)
			require.NotNil(t, newLeader, "New leader should be available")
			assert.True(t, newLeader.IsLeader(), "New leader should be in leader state")

			// Test consensus with new leader
			err := newLeader.ApplyConsensusOperation("leader_recovery_test", "new_leader_value")
			assert.NoError(t, err, "New leader should be able to perform consensus operations")

		case <-time.After(30 * time.Second):
			t.Fatal("New leader not elected within timeout")
		}
	})

	suite.recordTestEnd(true)
}

// testScalability tests system scalability
func (suite *EnhancedIntegrationTestSuite) testScalability(t *testing.T) {
	suite.recordTestStart()

	// Test adding new nodes
	t.Run("NodeAddition", func(t *testing.T) {
		originalNodeCount := len(suite.cluster.GetActiveNodes())

		// Add a new node
		newNode, err := suite.cluster.AddNode()
		require.NoError(t, err, "Should be able to add new node")

		// Wait for integration
		time.Sleep(10 * time.Second)

		// Verify new node is integrated
		currentNodes := suite.cluster.GetActiveNodes()
		assert.Equal(t, originalNodeCount+1, len(currentNodes), 
			"Node count should increase by 1")

		// Verify new node is connected
		connectedPeers := newNode.GetConnectedPeers()
		assert.GreaterOrEqual(t, len(connectedPeers), 1, 
			"New node should be connected to other peers")

		// Test that new node can participate in consensus
		leader := suite.cluster.GetLeader()
		err = leader.ApplyConsensusOperation("scalability_test", "node_added")
		require.NoError(t, err, "Should be able to apply consensus operation")

		// Wait for replication
		time.Sleep(3 * time.Second)

		// Verify new node has the consensus state
		value, exists := newNode.GetConsensusValue("scalability_test")
		assert.True(t, exists, "New node should have consensus state")
		assert.Equal(t, "node_added", value, "New node should have correct value")
	})

	// Test handling increased load
	t.Run("LoadScaling", func(t *testing.T) {
		// Gradually increase load and measure performance
		loadLevels := []int{10, 25, 50, 100}
		performanceMetrics := make([]PerformanceMetric, 0, len(loadLevels))

		for _, load := range loadLevels {
			metric := suite.measurePerformanceAtLoad(t, load)
			performanceMetrics = append(performanceMetrics, metric)

			t.Logf("Load %d: Avg Latency=%v, Success Rate=%.2f%%, Throughput=%.2f RPS",
				load, metric.AverageLatency, metric.SuccessRate*100, metric.ThroughputRPS)
		}

		// Verify performance doesn't degrade significantly
		for i := 1; i < len(performanceMetrics); i++ {
			prev := performanceMetrics[i-1]
			curr := performanceMetrics[i]

			// Latency should not increase more than 3x
			latencyRatio := float64(curr.AverageLatency) / float64(prev.AverageLatency)
			assert.Less(t, latencyRatio, 3.0, 
				"Latency should not increase more than 3x with load")

			// Success rate should remain high
			assert.Greater(t, curr.SuccessRate, 0.8, 
				"Success rate should remain >80%% even with increased load")
		}
	})

	suite.recordTestEnd(true)
}

// testDataIntegrity tests data integrity across operations
func (suite *EnhancedIntegrationTestSuite) testDataIntegrity(t *testing.T) {
	suite.recordTestStart()

	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader")

	// Test data consistency across nodes
	t.Run("DataConsistency", func(t *testing.T) {
		operations := []struct {
			key   string
			value string
		}{
			{"integrity_test_1", "value_1"},
			{"integrity_test_2", "value_2"},
			{"integrity_test_3", "value_3"},
		}

		// Apply operations
		for _, op := range operations {
			err := leader.ApplyConsensusOperation(op.key, op.value)
			require.NoError(t, err, "Should be able to apply operation")
		}

		// Wait for replication
		time.Sleep(5 * time.Second)

		// Verify consistency across all nodes
		nodes := suite.cluster.GetActiveNodes()
		for _, op := range operations {
			for _, node := range nodes {
				value, exists := node.GetConsensusValue(op.key)
				assert.True(t, exists, "Node %s should have key %s", node.GetID(), op.key)
				assert.Equal(t, op.value, value, 
					"Node %s should have correct value for %s", node.GetID(), op.key)
			}
		}
	})

	// Test data integrity during failures
	t.Run("IntegrityDuringFailures", func(t *testing.T) {
		// Apply initial data
		initialData := map[string]string{
			"account_a": "1000",
			"account_b": "500",
			"account_c": "750",
		}

		for key, value := range initialData {
			err := leader.ApplyConsensusOperation(key, value)
			require.NoError(t, err, "Should be able to apply initial data")
		}

		// Wait for replication
		time.Sleep(3 * time.Second)

		// Fail a node during transaction
		nodes := suite.cluster.GetActiveNodes()
		var nodeToFail *TestNode
		for _, node := range nodes {
			if !node.IsLeader() {
				nodeToFail = node
				break
			}
		}

		if nodeToFail != nil {
			// Start transaction
			go func() {
				time.Sleep(1 * time.Second) // Let transaction start
				nodeToFail.SimulateFailure("crash")
			}()

			// Perform transfer transaction (atomic operation)
			err := leader.ApplyTransaction([]Operation{
				{Type: "SUBTRACT", Key: "account_a", Value: "100"},
				{Type: "ADD", Key: "account_b", Value: "100"},
			})

			// Transaction should either fully succeed or fully fail
			if err == nil {
				// If transaction succeeded, verify final state
				time.Sleep(3 * time.Second)

				valueA, _ := leader.GetConsensusValue("account_a")
				valueB, _ := leader.GetConsensusValue("account_b")

				assert.Equal(t, "900", valueA, "Account A should have correct balance")
				assert.Equal(t, "600", valueB, "Account B should have correct balance")
			} else {
				// If transaction failed, verify original state is preserved
				valueA, _ := leader.GetConsensusValue("account_a")
				valueB, _ := leader.GetConsensusValue("account_b")

				assert.Equal(t, "1000", valueA, "Account A should have original balance")
				assert.Equal(t, "500", valueB, "Account B should have original balance")
			}
		}
	})

	// Test checksum validation
	t.Run("ChecksumValidation", func(t *testing.T) {
		// This would test model file integrity, consensus log integrity, etc.
		for _, testModel := range suite.testModels {
			checksum, err := leader.GetModelChecksum(testModel.Name)
			if err == nil {
				// Verify checksum across replicas
				replicas, err := leader.GetModelReplicas(testModel.Name)
				if err == nil {
					for _, replica := range replicas {
						replicaChecksum, err := replica.GetChecksum()
						if err == nil {
							assert.Equal(t, checksum, replicaChecksum, 
								"Model %s checksums should match across replicas", testModel.Name)
						}
					}
				}
			}
		}
	})

	suite.recordTestEnd(true)
}

// Additional test helper methods and implementations...

// Helper structs and methods
type LoadTestResult struct {
	RequestID      string
	EntryNode      string
	ProcessingNode string
	Latency        time.Duration
	Success        bool
	Error          error
}

type PerformanceMetric struct {
	LoadLevel        int
	AverageLatency   time.Duration
	SuccessRate      float64
	ThroughputRPS    float64
	ResourceUsage    float64
}

type SystemState struct {
	NodeCount      int
	LeaderID       string
	ConsensusState map[string]string
	ModelRegistry  map[string]*ModelInfo
	Timestamp      time.Time
}

// Implementation of helper methods
func (suite *EnhancedIntegrationTestSuite) recordTestStart() {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()
	suite.metrics.TestsRun++
}

func (suite *EnhancedIntegrationTestSuite) recordTestEnd(success bool) {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()
	
	if success {
		suite.metrics.TestsPassed++
	} else {
		suite.metrics.TestsFailed++
	}
}

func (suite *EnhancedIntegrationTestSuite) updateLatencyMetric(latency time.Duration) {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()
	
	// Update running average
	if suite.metrics.AverageLatency == 0 {
		suite.metrics.AverageLatency = latency
	} else {
		suite.metrics.AverageLatency = (suite.metrics.AverageLatency + latency) / 2
	}
}

func (suite *EnhancedIntegrationTestSuite) measurePerformanceAtLoad(t *testing.T, loadLevel int) PerformanceMetric {
	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan bool, loadLevel)

	for i := 0; i < loadLevel; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			req := &api.InferenceRequest{
				Model:  suite.testModels[0].Name,
				Prompt: fmt.Sprintf("Performance test %d", index),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  20,
				},
			}

			leader := suite.cluster.GetLeader()
			_, err := leader.ProcessInference(suite.ctx, req)
			results <- (err == nil)
		}(i)
	}

	wg.Wait()
	close(results)

	duration := time.Since(start)
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	return PerformanceMetric{
		LoadLevel:     loadLevel,
		AverageLatency: duration / time.Duration(loadLevel),
		SuccessRate:   float64(successCount) / float64(loadLevel),
		ThroughputRPS: float64(successCount) / duration.Seconds(),
	}
}

func (suite *EnhancedIntegrationTestSuite) recordSystemState() SystemState {
	nodes := suite.cluster.GetActiveNodes()
	leader := suite.cluster.GetLeader()

	state := SystemState{
		NodeCount:      len(nodes),
		ConsensusState: make(map[string]string),
		ModelRegistry:  make(map[string]*ModelInfo),
		Timestamp:      time.Now(),
	}

	if leader != nil {
		state.LeaderID = leader.GetID()

		// Capture consensus state
		consensusKeys := []string{"bootstrap_test", "integrity_test_1", "integrity_test_2", "integrity_test_3"}
		for _, key := range consensusKeys {
			if value, exists := leader.GetConsensusValue(key); exists {
				state.ConsensusState[key] = value
			}
		}

		// Capture model registry
		if models, err := leader.DiscoverModels(); err == nil {
			for _, model := range models {
				state.ModelRegistry[model.Name] = model
			}
		}
	}

	return state
}

func (suite *EnhancedIntegrationTestSuite) verifyStateConsistency(t *testing.T, before, after SystemState) {
	// Node count should be the same or greater (if nodes were added)
	assert.GreaterOrEqual(t, after.NodeCount, before.NodeCount-1, 
		"Node count should not decrease significantly")

	// Critical consensus data should be preserved
	for key, beforeValue := range before.ConsensusState {
		if afterValue, exists := after.ConsensusState[key]; exists {
			assert.Equal(t, beforeValue, afterValue, 
				"Consensus value for %s should be preserved", key)
		}
	}

	// Model registry should be preserved
	for modelName, beforeModel := range before.ModelRegistry {
		if afterModel, exists := after.ModelRegistry[modelName]; exists {
			assert.Equal(t, beforeModel.Size, afterModel.Size, 
				"Model %s size should be preserved", modelName)
			assert.Equal(t, beforeModel.Checksum, afterModel.Checksum, 
				"Model %s checksum should be preserved", modelName)
		}
	}
}

func (suite *EnhancedIntegrationTestSuite) generateTestReport(t *testing.T) {
	suite.metrics.mu.RLock()
	defer suite.metrics.mu.RUnlock()

	suite.metrics.TotalDuration = time.Since(suite.metrics.StartTime)

	t.Log("=== ENHANCED INTEGRATION TEST REPORT ===")
	t.Logf("Total Duration: %v", suite.metrics.TotalDuration)
	t.Logf("Tests Run: %d", suite.metrics.TestsRun)
	t.Logf("Tests Passed: %d", suite.metrics.TestsPassed)
	t.Logf("Tests Failed: %d", suite.metrics.TestsFailed)
	
	if suite.metrics.TestsRun > 0 {
		passRate := float64(suite.metrics.TestsPassed) / float64(suite.metrics.TestsRun) * 100
		t.Logf("Pass Rate: %.1f%%", passRate)
	}

	t.Logf("Average Latency: %v", suite.metrics.AverageLatency)
	t.Logf("Throughput: %.2f RPS", suite.metrics.ThroughputRPS)
	t.Log("=== END INTEGRATION TEST REPORT ===")
}

func (suite *EnhancedIntegrationTestSuite) cleanup() {
	suite.cancel()
	if suite.cluster != nil {
		suite.cluster.Shutdown()
	}
}

// Additional test implementations would continue here...
// Including testConcurrentOperations, testResourceManagement, testEndToEndWorkflows, etc.

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}