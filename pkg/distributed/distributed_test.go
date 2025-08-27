package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinBalancer(t *testing.T) {
	balancer := NewRoundRobinBalancer()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	request := &InferenceRequest{
		ID:    "test-1",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	// Test round-robin selection
	selections := make(map[string]int)
	for i := 0; i < 6; i++ {
		node, err := balancer.SelectNode(ctx, request, nodes)
		require.NoError(t, err)
		require.NotNil(t, node)
		selections[node.ID]++
	}
	
	// Each node should be selected twice
	assert.Equal(t, 2, selections["node1"])
	assert.Equal(t, 2, selections["node2"])
	assert.Equal(t, 2, selections["node3"])
}

func TestWeightedRoundRobinBalancer(t *testing.T) {
	weights := map[string]int{
		"node1": 1,
		"node2": 2,
		"node3": 3,
	}
	
	balancer := NewWeightedRoundRobinBalancer(weights)
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	request := &InferenceRequest{
		ID:    "test-1",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	// Test weighted selection
	selections := make(map[string]int)
	for i := 0; i < 12; i++ {
		node, err := balancer.SelectNode(ctx, request, nodes)
		require.NoError(t, err)
		require.NotNil(t, node)
		selections[node.ID]++
	}
	
	// node3 should be selected most (weight 3)
	assert.True(t, selections["node3"] >= selections["node2"])
	assert.True(t, selections["node2"] >= selections["node1"])
}

func TestLeastConnectionsBalancer(t *testing.T) {
	balancer := NewLeastConnectionsBalancer()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	// Set up different connection counts
	balancer.UpdateMetrics("node1", &NodeMetrics{NodeID: "node1", RequestCount: 10})
	balancer.UpdateMetrics("node2", &NodeMetrics{NodeID: "node2", RequestCount: 5})
	balancer.UpdateMetrics("node3", &NodeMetrics{NodeID: "node3", RequestCount: 15})
	
	request := &InferenceRequest{
		ID:    "test-1",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	// node2 should be selected (lowest connections)
	node, err := balancer.SelectNode(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, "node2", node.ID)
}

func TestLatencyBasedBalancer(t *testing.T) {
	balancer := NewLatencyBasedBalancer()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
	}
	
	// Set up different latencies
	balancer.UpdateMetrics("node1", &NodeMetrics{
		NodeID: "node1", 
		AverageLatency: 100 * time.Millisecond,
	})
	balancer.UpdateMetrics("node2", &NodeMetrics{
		NodeID: "node2", 
		AverageLatency: 50 * time.Millisecond,
	})
	balancer.UpdateMetrics("node3", &NodeMetrics{
		NodeID: "node3", 
		AverageLatency: 200 * time.Millisecond,
	})
	
	request := &InferenceRequest{
		ID:    "test-1",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	// node2 should be selected (lowest latency)
	node, err := balancer.SelectNode(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, "node2", node.ID)
}

func TestSmartLoadBalancer(t *testing.T) {
	balancer := NewSmartLoadBalancer()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
	}
	
	ctx := context.Background()
	
	// High priority request should use latency strategy
	highPriorityRequest := &InferenceRequest{
		ID:    "test-high",
		Model: "test-model",
		Prompt: "test prompt",
		Priority: 8,
	}
	
	node, err := balancer.SelectNode(ctx, highPriorityRequest, nodes)
	require.NoError(t, err)
	require.NotNil(t, node)
	
	// Low priority request should use round-robin
	lowPriorityRequest := &InferenceRequest{
		ID:    "test-low",
		Model: "test-model",
		Prompt: "test prompt",
		Priority: 1,
	}
	
	node, err = balancer.SelectNode(ctx, lowPriorityRequest, nodes)
	require.NoError(t, err)
	require.NotNil(t, node)
}

func TestLayerPartitionStrategy(t *testing.T) {
	strategy := NewLayerPartitionStrategy()
	
	// Mock layer mapping
	layers := []LayerInfo{
		{Name: "layer1", Parameters: 1000, MemoryUsage: 100 * 1024 * 1024},
		{Name: "layer2", Parameters: 2000, MemoryUsage: 200 * 1024 * 1024},
		{Name: "layer3", Parameters: 1500, MemoryUsage: 150 * 1024 * 1024},
		{Name: "layer4", Parameters: 1200, MemoryUsage: 120 * 1024 * 1024},
	}
	
	strategy.mu.Lock()
	strategy.layerMapping["test-model"] = layers
	strategy.mu.Unlock()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1"},
		{ID: "node2", Address: "192.168.1.2"},
	}
	
	request := &InferenceRequest{
		ID:    "test-partition",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	plan, err := strategy.Partition(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, plan)
	
	// Should have partitions for both nodes
	assert.Len(t, plan.Partitions, 2)
	assert.Equal(t, "layer", plan.Strategy)
	assert.Equal(t, "test-partition", plan.RequestID)
	
	// Validate the plan
	err = strategy.Validate(plan)
	assert.NoError(t, err)
	
	// Test latency estimation
	latency := strategy.EstimateLatency(plan)
	assert.True(t, latency > 0)
	
	// Test memory estimation
	memory := strategy.EstimateMemoryUsage(plan)
	assert.True(t, memory > 0)
}

func TestTensorPartitionStrategy(t *testing.T) {
	strategy := NewTensorPartitionStrategy()
	
	// Mock tensor mapping
	tensors := []TensorInfo{
		{Name: "tensor1", Shape: []int{100, 100}, Size: 40000},
		{Name: "tensor2", Shape: []int{200, 200}, Size: 160000},
	}
	
	strategy.mu.Lock()
	strategy.tensorMapping["test-model"] = tensors
	strategy.mu.Unlock()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1"},
		{ID: "node2", Address: "192.168.1.2"},
	}
	
	request := &InferenceRequest{
		ID:    "test-tensor-partition",
		Model: "test-model",
		Prompt: "test prompt",
	}
	
	ctx := context.Background()
	
	plan, err := strategy.Partition(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, plan)
	
	assert.Equal(t, "tensor", plan.Strategy)
	assert.Len(t, plan.Partitions, 2) // One per tensor
	
	err = strategy.Validate(plan)
	assert.NoError(t, err)
}

func TestDataPartitionStrategy(t *testing.T) {
	strategy := NewDataPartitionStrategy(10) // 10 character chunks
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1"},
		{ID: "node2", Address: "192.168.1.2"},
	}
	
	request := &InferenceRequest{
		ID:    "test-data-partition",
		Model: "test-model",
		Prompt: "This is a long test prompt that should be partitioned",
	}
	
	ctx := context.Background()
	
	plan, err := strategy.Partition(ctx, request, nodes)
	require.NoError(t, err)
	require.NotNil(t, plan)
	
	assert.Equal(t, "data", plan.Strategy)
	assert.True(t, len(plan.Partitions) > 0)
	
	// Check that data is actually partitioned
	totalDataSize := 0
	for _, partition := range plan.Partitions {
		if partition.Data != nil {
			totalDataSize += len(partition.Data)
		}
	}
	
	assert.Equal(t, len(request.Prompt), totalDataSize)
	
	err = strategy.Validate(plan)
	assert.NoError(t, err)
}

func TestFaultToleranceManager(t *testing.T) {
	config := &FaultToleranceConfig{
		DetectionConfig: &DetectionConfig{
			Thresholds: &HealthThresholds{
				CPUThreshold:    80.0,
				MemoryThreshold: 85.0,
				DiskThreshold:   90.0,
				LatencyThreshold: 1000 * time.Millisecond,
				ErrorRateThreshold: 0.05,
			},
			CheckInterval:    30 * time.Second,
			AlertHandlers:    []AlertHandler{},
			MetricsRetention: 24 * time.Hour,
		},
		RecoveryConfig: &RecoveryConfig{
			Strategies:    make(map[string]RecoveryStrategy),
			MaxRetries:    3,
			RetryInterval: 10 * time.Second,
			HistorySize:   100,
		},
		ReplicationConfig: &ReplicationConfig{
			ReplicationFactor: 3,
			Consistency:      "eventual",
			SyncInterval:     5 * time.Second,
		},
		CircuitConfig: &CircuitConfig{
			Thresholds: &CircuitThresholds{
				FailureThreshold: 5,
				TimeoutThreshold: 30 * time.Second,
				RetryInterval:    60 * time.Second,
			},
		},
		CheckpointConfig: &CheckpointConfig{
			Storage:       &MockCheckpointStorage{},
			Interval:      10 * time.Minute,
			MaxCheckpoints: 10,
		},
	}
	
	ftm := NewFaultToleranceManager(config)
	require.NotNil(t, ftm)
	
	ctx := context.Background()
	
	// Test starting fault tolerance
	err := ftm.Start(ctx)
	assert.NoError(t, err)
	
	// Test health checking (should return true for unknown nodes when enabled)
	healthy := ftm.IsHealthy("unknown-node")
	assert.False(t, healthy) // Unknown nodes are not healthy
	
	// Test stopping
	err = ftm.Stop()
	assert.NoError(t, err)
	
	// After stopping, all nodes should be considered healthy
	healthy = ftm.IsHealthy("any-node")
	assert.True(t, healthy)
}

func TestPartitionPlanValidation(t *testing.T) {
	strategy := NewLayerPartitionStrategy()
	
	// Test nil plan
	err := strategy.Validate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plan is nil")
	
	// Test empty partitions
	emptyPlan := &PartitionPlan{
		ID:         "empty",
		Partitions: []*Partition{},
	}
	
	err = strategy.Validate(emptyPlan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no partitions")
}

func TestLoadBalancerMetrics(t *testing.T) {
	balancer := NewRoundRobinBalancer()
	
	metrics := &NodeMetrics{
		NodeID:         "test-node",
		RequestCount:   100,
		SuccessCount:   95,
		ErrorCount:     5,
		AverageLatency: 50 * time.Millisecond,
		CurrentLoad:    0.7,
		LastUpdated:    time.Now(),
	}
	
	balancer.UpdateMetrics("test-node", metrics)
	
	allMetrics := balancer.GetMetrics()
	assert.Len(t, allMetrics, 1)
	
	retrievedMetrics := allMetrics["test-node"]
	require.NotNil(t, retrievedMetrics)
	assert.Equal(t, int64(100), retrievedMetrics.RequestCount)
	assert.Equal(t, int64(95), retrievedMetrics.SuccessCount)
	assert.Equal(t, int64(5), retrievedMetrics.ErrorCount)
}

// MockCheckpointStorage for testing
type MockCheckpointStorage struct{}

func (mcs *MockCheckpointStorage) Save(ctx context.Context, checkpoint *Checkpoint) error {
	return nil
}

func (mcs *MockCheckpointStorage) Load(ctx context.Context, id string) (*Checkpoint, error) {
	return &Checkpoint{ID: id}, nil
}

func (mcs *MockCheckpointStorage) Delete(ctx context.Context, id string) error {
	return nil
}

func (mcs *MockCheckpointStorage) List(ctx context.Context, nodeID string) ([]*Checkpoint, error) {
	return []*Checkpoint{}, nil
}

func BenchmarkRoundRobinSelection(b *testing.B) {
	balancer := NewRoundRobinBalancer()
	
	nodes := []NodeInfo{
		{ID: "node1", Address: "192.168.1.1", Status: "active"},
		{ID: "node2", Address: "192.168.1.2", Status: "active"},
		{ID: "node3", Address: "192.168.1.3", Status: "active"},
		{ID: "node4", Address: "192.168.1.4", Status: "active"},
		{ID: "node5", Address: "192.168.1.5", Status: "active"},
	}
	
	request := &InferenceRequest{
		ID:    "bench-test",
		Model: "test-model",
		Prompt: "benchmark test prompt",
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := balancer.SelectNode(ctx, request, nodes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLayerPartitioning(b *testing.B) {
	strategy := NewLayerPartitionStrategy()
	
	// Setup layer mapping
	layers := make([]LayerInfo, 100) // 100 layers
	for i := range layers {
		layers[i] = LayerInfo{
			Name:        fmt.Sprintf("layer%d", i),
			Parameters:  1000,
			MemoryUsage: 100 * 1024 * 1024,
		}
	}
	
	strategy.mu.Lock()
	strategy.layerMapping["bench-model"] = layers
	strategy.mu.Unlock()
	
	nodes := make([]NodeInfo, 10) // 10 nodes
	for i := range nodes {
		nodes[i] = NodeInfo{
			ID:      fmt.Sprintf("node%d", i),
			Address: fmt.Sprintf("192.168.1.%d", i+1),
		}
	}
	
	request := &InferenceRequest{
		ID:    "bench-partition",
		Model: "bench-model",
		Prompt: "benchmark test",
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plan, err := strategy.Partition(ctx, request, nodes)
		if err != nil {
			b.Fatal(err)
		}
		if plan == nil {
			b.Fatal("plan is nil")
		}
	}
}