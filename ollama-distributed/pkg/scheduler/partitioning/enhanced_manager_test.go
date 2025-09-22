package partitioning

import (
	"context"
	"testing"

	api_types "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

func TestEnhancedManager_ConsumeNodeCapabilities(t *testing.T) {
	// Create enhanced manager
	manager := NewEnhancedPartitionManager().(*EnhancedPartitionManagerImpl)

	// Test case 1: With preferred_nodes_capabilities metadata
	t.Run("WithNodeCapabilities", func(t *testing.T) {
		task := &api_types.DistributedTask{
			ID:        "test-task-1",
			ModelName: "llama-7b",
			Metadata: map[string]interface{}{
				"preferred_node_ids": []string{"node-1", "node-2", "node-3"},
				"preferred_nodes_capabilities": []map[string]interface{}{
					{
						"id":           "node-1",
						"cpu_cores":    16,
						"mem_total":    int64(64 * 1024 * 1024 * 1024), // 64GB
						"mem_available": int64(48 * 1024 * 1024 * 1024), // 48GB
						"gpu_count":    2,
						"net_bw":       100.0, // 100 Gbps
						"net_latency":  0.5,   // 0.5ms
					},
					{
						"id":           "node-2",
						"cpu_cores":    32,
						"mem_total":    int64(128 * 1024 * 1024 * 1024), // 128GB
						"mem_available": int64(96 * 1024 * 1024 * 1024),  // 96GB
						"gpu_count":    4,
						"net_bw":       200.0, // 200 Gbps
						"net_latency":  0.3,   // 0.3ms
					},
					{
						"id":           "node-3",
						"cpu_cores":    8,
						"mem_total":    int64(32 * 1024 * 1024 * 1024), // 32GB
						"mem_available": int64(16 * 1024 * 1024 * 1024), // 16GB
						"gpu_count":    0,
						"net_bw":       10.0, // 10 Gbps
						"net_latency":  2.0,  // 2ms
					},
				},
			},
		}

		nodes, err := manager.getAvailableNodes(task)
		if err != nil {
			t.Fatalf("Failed to get available nodes: %v", err)
		}

		if len(nodes) != 3 {
			t.Errorf("Expected 3 nodes, got %d", len(nodes))
		}

		// Verify node-1 capabilities
		if nodes[0].ID != "node-1" {
			t.Errorf("Expected first node ID to be 'node-1', got '%s'", nodes[0].ID)
		}
		if nodes[0].Capabilities.CPU.Cores != 16 {
			t.Errorf("Expected node-1 to have 16 CPU cores, got %d", nodes[0].Capabilities.CPU.Cores)
		}
		if nodes[0].Capabilities.Memory.TotalBytes != 64*1024*1024*1024 {
			t.Errorf("Expected node-1 to have 64GB total memory, got %d", nodes[0].Capabilities.Memory.TotalBytes)
		}
		if nodes[0].Capabilities.GPU == nil || nodes[0].Capabilities.GPU.Count != 2 {
			t.Errorf("Expected node-1 to have 2 GPUs")
		}
		if nodes[0].Capabilities.Network.Bandwidth != 100.0 {
			t.Errorf("Expected node-1 to have 100Gbps bandwidth, got %f", nodes[0].Capabilities.Network.Bandwidth)
		}

		// Verify node-2 capabilities
		if nodes[1].ID != "node-2" {
			t.Errorf("Expected second node ID to be 'node-2', got '%s'", nodes[1].ID)
		}
		if nodes[1].Capabilities.CPU.Cores != 32 {
			t.Errorf("Expected node-2 to have 32 CPU cores, got %d", nodes[1].Capabilities.CPU.Cores)
		}
		if nodes[1].Capabilities.Memory.TotalBytes != 128*1024*1024*1024 {
			t.Errorf("Expected node-2 to have 128GB total memory, got %d", nodes[1].Capabilities.Memory.TotalBytes)
		}
		if nodes[1].Capabilities.GPU == nil || nodes[1].Capabilities.GPU.Count != 4 {
			t.Errorf("Expected node-2 to have 4 GPUs")
		}

		// Verify node-3 capabilities (no GPU)
		if nodes[2].ID != "node-3" {
			t.Errorf("Expected third node ID to be 'node-3', got '%s'", nodes[2].ID)
		}
		if nodes[2].Capabilities.GPU != nil && nodes[2].Capabilities.GPU.Count > 0 {
			t.Errorf("Expected node-3 to have no GPUs")
		}
		if nodes[2].Capabilities.Network.Latency != 2.0 {
			t.Errorf("Expected node-3 to have 2ms latency, got %f", nodes[2].Capabilities.Network.Latency)
		}
	})

	// Test case 2: With []interface{} type for capabilities (JSON unmarshaling)
	t.Run("WithInterfaceCapabilities", func(t *testing.T) {
		task := &api_types.DistributedTask{
			ID:        "test-task-2",
			ModelName: "llama-13b",
			Metadata: map[string]interface{}{
				"preferred_nodes_capabilities": []interface{}{
					map[string]interface{}{
						"id":           "node-a",
						"cpu_cores":    float64(24), // JSON numbers come as float64
						"mem_total":    float64(96 * 1024 * 1024 * 1024),
						"mem_available": float64(72 * 1024 * 1024 * 1024),
						"gpu_count":    float64(1),
						"net_bw":       50.0,
						"net_latency":  1.5,
					},
				},
			},
		}

		nodes, err := manager.getAvailableNodes(task)
		if err != nil {
			t.Fatalf("Failed to get available nodes: %v", err)
		}

		if len(nodes) != 1 {
			t.Errorf("Expected 1 node, got %d", len(nodes))
		}

		if nodes[0].ID != "node-a" {
			t.Errorf("Expected node ID to be 'node-a', got '%s'", nodes[0].ID)
		}
		if nodes[0].Capabilities.CPU.Cores != 24 {
			t.Errorf("Expected node to have 24 CPU cores, got %d", nodes[0].Capabilities.CPU.Cores)
		}
	})

	// Test case 3: Fallback to preferred_node_ids only
	t.Run("FallbackToNodeIDs", func(t *testing.T) {
		task := &api_types.DistributedTask{
			ID:        "test-task-3",
			ModelName: "llama-70b",
			Metadata: map[string]interface{}{
				"preferred_node_ids": []string{"fallback-1", "fallback-2"},
			},
		}

		nodes, err := manager.getAvailableNodes(task)
		if err != nil {
			t.Fatalf("Failed to get available nodes: %v", err)
		}

		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(nodes))
		}

		// Should have default capabilities
		if nodes[0].ID != "fallback-1" {
			t.Errorf("Expected first node ID to be 'fallback-1', got '%s'", nodes[0].ID)
		}
		// Default capabilities should be applied
		if nodes[0].Capabilities.CPU.Cores != 8 {
			t.Errorf("Expected default 8 CPU cores, got %d", nodes[0].Capabilities.CPU.Cores)
		}
	})

	// Test case 4: Order preservation with both IDs and capabilities
	t.Run("OrderPreservation", func(t *testing.T) {
		task := &api_types.DistributedTask{
			ID:        "test-task-4",
			ModelName: "mixtral-8x7b",
			Metadata: map[string]interface{}{
				"preferred_node_ids": []string{"node-z", "node-y", "node-x"},
				"preferred_nodes_capabilities": []map[string]interface{}{
					{"id": "node-y", "cpu_cores": 10, "mem_total": int64(10 * 1024 * 1024 * 1024), "mem_available": int64(8 * 1024 * 1024 * 1024)},
					{"id": "node-x", "cpu_cores": 20, "mem_total": int64(20 * 1024 * 1024 * 1024), "mem_available": int64(15 * 1024 * 1024 * 1024)},
					{"id": "node-z", "cpu_cores": 30, "mem_total": int64(30 * 1024 * 1024 * 1024), "mem_available": int64(25 * 1024 * 1024 * 1024)},
				},
			},
		}

		nodes, err := manager.getAvailableNodes(task)
		if err != nil {
			t.Fatalf("Failed to get available nodes: %v", err)
		}

		if len(nodes) != 3 {
			t.Errorf("Expected 3 nodes, got %d", len(nodes))
		}

		// Verify order matches preferred_node_ids
		expectedOrder := []string{"node-z", "node-y", "node-x"}
		for i, expectedID := range expectedOrder {
			if nodes[i].ID != expectedID {
				t.Errorf("Expected node at position %d to be '%s', got '%s'", i, expectedID, nodes[i].ID)
			}
		}

		// Verify capabilities are correctly mapped
		if nodes[0].Capabilities.CPU.Cores != 30 { // node-z
			t.Errorf("Expected node-z to have 30 cores, got %d", nodes[0].Capabilities.CPU.Cores)
		}
		if nodes[1].Capabilities.CPU.Cores != 10 { // node-y
			t.Errorf("Expected node-y to have 10 cores, got %d", nodes[1].Capabilities.CPU.Cores)
		}
		if nodes[2].Capabilities.CPU.Cores != 20 { // node-x
			t.Errorf("Expected node-x to have 20 cores, got %d", nodes[2].Capabilities.CPU.Cores)
		}
	})
}

func TestEnhancedManager_PartitionWithCapabilities(t *testing.T) {
	ctx := context.Background()
	manager := NewEnhancedPartitionManager()

	task := &api_types.DistributedTask{
		ID:        "partition-test",
		ModelName: "llama-13b",
		Metadata: map[string]interface{}{
			"preferred_nodes_capabilities": []map[string]interface{}{
				{
					"id":           "gpu-node-1",
					"cpu_cores":    32,
					"mem_total":    int64(256 * 1024 * 1024 * 1024), // 256GB
					"mem_available": int64(200 * 1024 * 1024 * 1024), // 200GB
					"gpu_count":    8,
					"net_bw":       400.0, // 400 Gbps (high bandwidth for tensor parallelism)
					"net_latency":  0.1,   // 0.1ms (low latency)
				},
				{
					"id":           "gpu-node-2",
					"cpu_cores":    32,
					"mem_total":    int64(256 * 1024 * 1024 * 1024),
					"mem_available": int64(200 * 1024 * 1024 * 1024),
					"gpu_count":    8,
					"net_bw":       400.0,
					"net_latency":  0.1,
				},
			},
		},
	}

	// This should select tensor parallelism due to high bandwidth nodes
	plan, err := manager.Partition(ctx, task)
	if err != nil {
		t.Fatalf("Failed to partition task: %v", err)
	}

	if plan == nil {
		t.Fatal("Expected non-nil partition plan")
	}

	// The strategy selection should consider the high bandwidth
	// (tensor_parallelism needs high bandwidth between nodes)
	t.Logf("Selected strategy: %s", plan.Strategy)
	t.Logf("Number of assignments: %d", len(plan.Assignments))

	// Verify assignments use the provided nodes
	for _, assignment := range plan.Assignments {
		if assignment.NodeID != "gpu-node-1" && assignment.NodeID != "gpu-node-2" {
			t.Errorf("Unexpected node ID in assignment: %s", assignment.NodeID)
		}
	}
}