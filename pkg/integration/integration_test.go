package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/pkg/distributed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDistributedSystemIntegration tests the integration of all distributed components
func TestDistributedSystemIntegration(t *testing.T) {
	// Setup distributed components
	loadBalancer := distributed.NewSmartLoadBalancer()
	layerStrategy := distributed.NewLayerPartitionStrategy()
	
	// Mock nodes
	nodes := []distributed.NodeInfo{
		{
			ID:      "node1",
			Address: "http://localhost:8081",
			Status:  "active",
			Models:  []string{"test-model"},
			Capacity: &distributed.NodeCapacity{
				CPU:    2.0,
				Memory: 8 * 1024 * 1024 * 1024, // 8GB
				GPU:    1,
			},
		},
		{
			ID:      "node2", 
			Address: "http://localhost:8082",
			Status:  "active",
			Models:  []string{"test-model"},
			Capacity: &distributed.NodeCapacity{
				CPU:    4.0,
				Memory: 16 * 1024 * 1024 * 1024, // 16GB
				GPU:    2,
			},
		},
		{
			ID:      "node3",
			Address: "http://localhost:8083", 
			Status:  "active",
			Models:  []string{"test-model"},
			Capacity: &distributed.NodeCapacity{
				CPU:    1.0,
				Memory: 4 * 1024 * 1024 * 1024, // 4GB
				GPU:    0,
			},
		},
	}
	
	ctx := context.Background()
	
	// Test load balancer node selection
	request := &distributed.InferenceRequest{
		ID:    "integration-test-1",
		Model: "test-model",
		Prompt: "This is a test inference request for integration testing",
		Priority: 5,
	}
	
	selectedNode, err := loadBalancer.SelectNode(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, selectedNode)
	assert.Contains(t, []string{"node1", "node2", "node3"}, selectedNode.ID)
	
	// Test partition strategy
	plan, err := layerStrategy.Partition(ctx, request, nodes)
	if err != nil {
		// Expected for layer strategy without mapping
		assert.Contains(t, err.Error(), "no layer mapping found")
	}
	
	// Update load balancer metrics
	metrics := &distributed.NodeMetrics{
		NodeID:         selectedNode.ID,
		RequestCount:   1,
		SuccessCount:   1,
		ErrorCount:     0,
		AverageLatency: 100 * time.Millisecond,
		CurrentLoad:    0.5,
		LastUpdated:    time.Now(),
	}
	
	loadBalancer.UpdateMetrics(selectedNode.ID, metrics)
	
	// Verify metrics were stored
	allMetrics := loadBalancer.GetMetrics()
	assert.Contains(t, allMetrics, selectedNode.ID)
	assert.Equal(t, int64(1), allMetrics[selectedNode.ID].RequestCount)
}

func TestFaultToleranceIntegration(t *testing.T) {
	// Create fault tolerance configuration
	config := &distributed.FaultToleranceConfig{
		DetectionConfig: &distributed.DetectionConfig{
			Thresholds: &distributed.HealthThresholds{
				CPUThreshold:    80.0,
				MemoryThreshold: 85.0,
				DiskThreshold:   90.0,
				LatencyThreshold: 1000 * time.Millisecond,
				ErrorRateThreshold: 0.05,
			},
			CheckInterval:    10 * time.Second,
			AlertHandlers:    []distributed.AlertHandler{},
			MetricsRetention: 1 * time.Hour,
		},
		RecoveryConfig: &distributed.RecoveryConfig{
			Strategies:    make(map[string]distributed.RecoveryStrategy),
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
			HistorySize:   50,
		},
		ReplicationConfig: &distributed.ReplicationConfig{
			ReplicationFactor: 2,
			Consistency:      "strong",
			SyncInterval:     1 * time.Second,
		},
		CircuitConfig: &distributed.CircuitConfig{
			Thresholds: &distributed.CircuitThresholds{
				FailureThreshold: 3,
				TimeoutThreshold: 10 * time.Second,
				RetryInterval:    30 * time.Second,
			},
		},
		CheckpointConfig: &distributed.CheckpointConfig{
			Storage:       &MockCheckpointStorage{},
			Interval:      5 * time.Minute,
			MaxCheckpoints: 5,
		},
	}
	
	ftManager := distributed.NewFaultToleranceManager(config)
	require.NotNil(t, ftManager)
	
	ctx := context.Background()
	
	// Start fault tolerance
	err := ftManager.Start(ctx)
	require.NoError(t, err)
	
	// Test health checking for unknown nodes
	healthy := ftManager.IsHealthy("unknown-node")
	assert.False(t, healthy)
	
	// Stop fault tolerance
	err = ftManager.Stop()
	require.NoError(t, err)
	
	// After stopping, nodes should be considered healthy
	healthy = ftManager.IsHealthy("any-node")
	assert.True(t, healthy)
}

func TestLoadBalancingStrategies(t *testing.T) {
	strategies := []struct {
		name     string
		balancer distributed.LoadBalancer
	}{
		{"round-robin", distributed.NewRoundRobinBalancer()},
		{"least-connections", distributed.NewLeastConnectionsBalancer()},
		{"latency-based", distributed.NewLatencyBasedBalancer()},
		{"smart", distributed.NewSmartLoadBalancer()},
	}
	
	nodes := []distributed.NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	request := &distributed.InferenceRequest{
		ID:    "strategy-test",
		Model: "test-model",
		Prompt: "test prompt for strategy testing",
		Priority: 3,
	}
	
	ctx := context.Background()
	
	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			// Test node selection
			node, err := strategy.balancer.SelectNode(ctx, request, nodes)
			require.NoError(t, err, "Strategy %s failed", strategy.name)
			require.NotNil(t, node, "Strategy %s returned nil node", strategy.name)
			assert.Contains(t, []string{"node1", "node2", "node3"}, node.ID)
			
			// Test metrics update
			metrics := &distributed.NodeMetrics{
				NodeID:         node.ID,
				RequestCount:   10,
				SuccessCount:   9,
				ErrorCount:     1,
				AverageLatency: 75 * time.Millisecond,
				CurrentLoad:    0.6,
				LastUpdated:    time.Now(),
			}
			
			strategy.balancer.UpdateMetrics(node.ID, metrics)
			
			// Verify metrics retrieval
			allMetrics := strategy.balancer.GetMetrics()
			assert.Contains(t, allMetrics, node.ID)
		})
	}
}

func TestPartitioningStrategies(t *testing.T) {
	strategies := []struct {
		name     string
		strategy distributed.PartitionStrategy
	}{
		{"layer", distributed.NewLayerPartitionStrategy()},
		{"tensor", distributed.NewTensorPartitionStrategy()},
		{"pipeline", distributed.NewPipelinePartitionStrategy()},
		{"data", distributed.NewDataPartitionStrategy(50)},
	}
	
	nodes := []distributed.NodeInfo{
		{ID: "node1", Address: "192.168.1.1"},
		{ID: "node2", Address: "192.168.1.2"},
		{ID: "node3", Address: "192.168.1.3"},
	}
	
	request := &distributed.InferenceRequest{
		ID:    "partition-test",
		Model: "test-model",
		Prompt: "This is a long test prompt that can be used for partitioning strategies testing and validation",
	}
	
	ctx := context.Background()
	
	for _, strategy := range strategies {
		t.Run(strategy.name, func(t *testing.T) {
			plan, err := strategy.strategy.Partition(ctx, request, nodes)
			
			if strategy.name == "data" {
				// Data partitioning should always work
				require.NoError(t, err)
				require.NotNil(t, plan)
				assert.Equal(t, strategy.name, plan.Strategy)
				assert.True(t, len(plan.Partitions) > 0)
				
				// Test validation
				err = strategy.strategy.Validate(plan)
				assert.NoError(t, err)
				
				// Test estimates
				latency := strategy.strategy.EstimateLatency(plan)
				assert.True(t, latency > 0)
				
				memory := strategy.strategy.EstimateMemoryUsage(plan)
				assert.True(t, memory > 0)
				
			} else {
				// Other strategies need mappings, so they should fail
				assert.Error(t, err)
			}
		})
	}
}

func TestAPIEndpointIntegration(t *testing.T) {
	// Create a mock API server
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"healthy","timestamp":1234567890}`)
	})
	
	// Inference endpoint
	mux.HandleFunc("/inference", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"result":"test response","latency_ms":150}`)
	})
	
	// Metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"requests_total":100,"avg_latency_ms":125}`)
	})
	
	server := httptest.NewServer(mux)
	defer server.Close()
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Test health endpoint
	resp, err := client.Get(server.URL + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// Test metrics endpoint
	resp, err = client.Get(server.URL + "/metrics")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
	
	// Test inference endpoint with POST
	resp, err = client.Post(server.URL+"/inference", "application/json", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestConcurrentLoadBalancing(t *testing.T) {
	balancer := distributed.NewRoundRobinBalancer()
	
	nodes := []distributed.NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	ctx := context.Background()
	concurrentRequests := 100
	results := make(chan string, concurrentRequests)
	
	// Launch concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		go func(requestID int) {
			request := &distributed.InferenceRequest{
				ID:    fmt.Sprintf("concurrent-%d", requestID),
				Model: "test-model",
				Prompt: fmt.Sprintf("concurrent test %d", requestID),
			}
			
			node, err := balancer.SelectNode(ctx, request, nodes)
			if err != nil {
				results <- "ERROR"
			} else {
				results <- node.ID
			}
		}(i)
	}
	
	// Collect results
	selections := make(map[string]int)
	for i := 0; i < concurrentRequests; i++ {
		result := <-results
		if result != "ERROR" {
			selections[result]++
		}
	}
	
	// Verify all requests succeeded
	totalSelections := selections["node1"] + selections["node2"] + selections["node3"]
	assert.Equal(t, concurrentRequests, totalSelections)
	
	// Verify distribution is roughly even (within 20% tolerance)
	expected := concurrentRequests / 3
	tolerance := expected / 5
	
	for nodeID, count := range selections {
		assert.True(t, count >= expected-tolerance && count <= expected+tolerance,
			"Node %s: expected ~%d selections (±%d), got %d", nodeID, expected, tolerance, count)
	}
}

func TestSystemResilience(t *testing.T) {
	balancer := distributed.NewSmartLoadBalancer()
	
	// Initially healthy nodes
	nodes := []distributed.NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "failed"}, // One failed node
	}
	
	ctx := context.Background()
	request := &distributed.InferenceRequest{
		ID:    "resilience-test",
		Model: "test-model",
		Prompt: "resilience test",
	}
	
	// Should still work with some failed nodes
	node, err := balancer.SelectNode(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.NotEqual(t, "node3", node.ID) // Should not select failed node
	
	// Test with no nodes
	emptyNodes := []distributed.NodeInfo{}
	_, err = balancer.SelectNode(ctx, request, emptyNodes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no nodes available")
}

// MockCheckpointStorage implements CheckpointStorage for testing
type MockCheckpointStorage struct{}

func (mcs *MockCheckpointStorage) Save(ctx context.Context, checkpoint *distributed.Checkpoint) error {
	return nil
}

func (mcs *MockCheckpointStorage) Load(ctx context.Context, id string) (*distributed.Checkpoint, error) {
	return &distributed.Checkpoint{
		ID:        id,
		NodeID:    "test-node",
		Timestamp: time.Now(),
		Data:      []byte("test checkpoint data"),
		Hash:      "mock-hash",
	}, nil
}

func (mcs *MockCheckpointStorage) Delete(ctx context.Context, id string) error {
	return nil
}

func (mcs *MockCheckpointStorage) List(ctx context.Context, nodeID string) ([]*distributed.Checkpoint, error) {
	return []*distributed.Checkpoint{
		{
			ID:        "checkpoint-1",
			NodeID:    nodeID,
			Timestamp: time.Now().Add(-1 * time.Hour),
			Hash:      "mock-hash-1",
		},
		{
			ID:        "checkpoint-2",
			NodeID:    nodeID,
			Timestamp: time.Now().Add(-30 * time.Minute),
			Hash:      "mock-hash-2",
		},
	}, nil
}

func BenchmarkIntegratedSystem(b *testing.B) {
	balancer := distributed.NewSmartLoadBalancer()
	strategy := distributed.NewDataPartitionStrategy(100)
	
	nodes := []distributed.NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
		{ID: "node4", Address: "192.168.1.4", Status: "active"},
		{ID: "node5", Address: "192.168.1.5", Status: "active"},
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &distributed.InferenceRequest{
			ID:    fmt.Sprintf("bench-%d", i),
			Model: "benchmark-model",
			Prompt: "This is a benchmark test prompt that simulates a real inference request with sufficient length for partitioning",
			Priority: i % 10,
		}
		
		// Load balancing
		_, err := balancer.SelectNode(ctx, request, nodes)
		if err != nil {
			b.Fatal(err)
		}
		
		// Partitioning
		_, err = strategy.Partition(ctx, request, nodes)
		if err != nil {
			b.Fatal(err)
		}
	}
}