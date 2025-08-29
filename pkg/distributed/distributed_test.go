package distributed

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLoadBalancer(t *testing.T) {
	// Test basic load balancer functionality
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active", Models: []string{"model1"}},
		{ID: "node2", Address: "node2:11434", Status: "active", Models: []string{"model1"}},
		{ID: "node3", Address: "node3:11434", Status: "active", Models: []string{"model1"}},
	}

	lb := NewRoundRobinBalancer()

	// Test node selection
	ctx := context.Background()
	request := &InferenceRequest{ID: "test", Model: "model1", Prompt: "test"}
	
	selected, err := lb.SelectNode(ctx, request, nodes)
	if err != nil {
		t.Fatalf("Should select a node when nodes are available: %v", err)
	}

	if selected == nil {
		t.Error("Should return a selected node")
	}

	if selected.Status != "active" {
		t.Error("Should only select active nodes")
	}
}

func TestRoundRobinSelection(t *testing.T) {
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active"},
		{ID: "node2", Address: "node2:11434", Status: "active"},
		{ID: "node3", Address: "node3:11434", Status: "active"},
	}

	lb := NewRoundRobinBalancer()
	ctx := context.Background()
	request := &InferenceRequest{ID: "test", Model: "model1", Prompt: "test"}

	// Test that selections rotate through nodes
	selections := make(map[string]int)
	for i := 0; i < 9; i++ {
		selected, err := lb.SelectNode(ctx, request, nodes)
		if err != nil {
			t.Fatalf("Selection failed: %v", err)
		}
		selections[selected.ID]++
	}

	// Should distribute across all nodes
	if len(selections) != 3 {
		t.Errorf("Expected 3 nodes selected, got %d", len(selections))
	}

	// Each node should be selected equally
	for nodeID, count := range selections {
		if count != 3 {
			t.Errorf("Node %s selected %d times, expected 3", nodeID, count)
		}
	}
}

func TestLeastConnectionsSelection(t *testing.T) {
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active"},
		{ID: "node2", Address: "node2:11434", Status: "active"},
		{ID: "node3", Address: "node3:11434", Status: "active"},
	}

	lb := NewLeastConnectionsBalancer()
	ctx := context.Background()
	request := &InferenceRequest{ID: "test", Model: "model1", Prompt: "test"}

	// Update metrics to simulate different load levels
	lb.UpdateMetrics("node1", &NodeMetrics{RequestCount: 10, SuccessCount: 8, ErrorCount: 2})
	lb.UpdateMetrics("node2", &NodeMetrics{RequestCount: 5, SuccessCount: 4, ErrorCount: 1})
	lb.UpdateMetrics("node3", &NodeMetrics{RequestCount: 15, SuccessCount: 12, ErrorCount: 3})

	selected, err := lb.SelectNode(ctx, request, nodes)
	if err != nil {
		t.Fatalf("Selection failed: %v", err)
	}

	// Should select node2 (least connections)
	if selected.ID != "node2" {
		t.Errorf("Expected node2 to be selected (least connections), got %s", selected.ID)
	}
}

func TestLatencyBasedSelection(t *testing.T) {
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active"},
		{ID: "node2", Address: "node2:11434", Status: "active"},
		{ID: "node3", Address: "node3:11434", Status: "active"},
	}

	lb := NewLatencyBasedBalancer()
	ctx := context.Background()
	request := &InferenceRequest{ID: "test", Model: "model1", Prompt: "test"}

	// Update metrics with different latencies
	lb.UpdateMetrics("node1", &NodeMetrics{AverageLatency: 100 * time.Millisecond})
	lb.UpdateMetrics("node2", &NodeMetrics{AverageLatency: 50 * time.Millisecond})
	lb.UpdateMetrics("node3", &NodeMetrics{AverageLatency: 150 * time.Millisecond})

	selected, err := lb.SelectNode(ctx, request, nodes)
	if err != nil {
		t.Fatalf("Selection failed: %v", err)
	}

	// Should select node2 (lowest latency)
	if selected.ID != "node2" {
		t.Errorf("Expected node2 to be selected (lowest latency), got %s", selected.ID)
	}
}

func TestSmartLoadBalancer(t *testing.T) {
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active"},
		{ID: "node2", Address: "node2:11434", Status: "active"},
		{ID: "node3", Address: "node3:11434", Status: "active"},
	}

	slb := NewSmartLoadBalancer()
	ctx := context.Background()

	// Test high priority request (should use latency-based)
	highPriorityRequest := &InferenceRequest{
		ID:       "test-high",
		Model:    "model1",
		Prompt:   "test",
		Priority: 10,
	}

	selected, err := slb.SelectNode(ctx, highPriorityRequest, nodes)
	if err != nil {
		t.Fatalf("High priority selection failed: %v", err)
	}

	if selected == nil {
		t.Error("Should select a node for high priority request")
	}

	// Test normal priority request
	normalRequest := &InferenceRequest{
		ID:       "test-normal",
		Model:    "model1",
		Prompt:   "test",
		Priority: 3,
	}

	selected, err = slb.SelectNode(ctx, normalRequest, nodes)
	if err != nil {
		t.Fatalf("Normal priority selection failed: %v", err)
	}

	if selected == nil {
		t.Error("Should select a node for normal priority request")
	}
}

func TestEmptyNodeList(t *testing.T) {
	lb := NewRoundRobinBalancer()
	ctx := context.Background()
	request := &InferenceRequest{ID: "test", Model: "model1", Prompt: "test"}

	selected, err := lb.SelectNode(ctx, request, []NodeInfo{})
	if err == nil {
		t.Error("Should return error for empty node list")
	}

	if selected != nil {
		t.Error("Should not return a node when list is empty")
	}
}

func TestMetricsUpdate(t *testing.T) {
	lb := NewRoundRobinBalancer()
	
	metrics := &NodeMetrics{
		NodeID:           "test-node",
		RequestCount:     100,
		SuccessCount:     95,
		ErrorCount:       5,
		AverageLatency:   50 * time.Millisecond,
		CurrentLoad:      0.7,
		LastUpdated:      time.Now(),
	}

	lb.UpdateMetrics("test-node", metrics)
	
	allMetrics := lb.GetMetrics()
	
	retrievedMetrics, exists := allMetrics["test-node"]
	if !exists {
		t.Error("Metrics should exist after update")
	}

	if retrievedMetrics.RequestCount != 100 {
		t.Errorf("Expected 100 requests, got %d", retrievedMetrics.RequestCount)
	}

	if retrievedMetrics.SuccessCount != 95 {
		t.Errorf("Expected 95 successes, got %d", retrievedMetrics.SuccessCount)
	}
}

func TestConcurrentAccess(t *testing.T) {
	nodes := []NodeInfo{
		{ID: "node1", Address: "node1:11434", Status: "active"},
		{ID: "node2", Address: "node2:11434", Status: "active"},
	}

	lb := NewRoundRobinBalancer()
	ctx := context.Background()

	var wg sync.WaitGroup
	const numGoroutines = 10
	const selectionsPerGoroutine = 100

	// Test concurrent selections
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < selectionsPerGoroutine; j++ {
				request := &InferenceRequest{
					ID:     fmt.Sprintf("req-%d-%d", goroutineID, j),
					Model:  "model1",
					Prompt: "test",
				}
				
				selected, err := lb.SelectNode(ctx, request, nodes)
				if err != nil {
					t.Errorf("Concurrent selection failed: %v", err)
					return
				}
				
				if selected == nil {
					t.Error("Concurrent selection returned nil")
					return
				}
			}
		}(i)
	}

	// Test concurrent metrics updates
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < selectionsPerGoroutine; j++ {
				metrics := &NodeMetrics{
					NodeID:         fmt.Sprintf("node-%d", goroutineID%2+1),
					RequestCount:   int64(j + 1),
					SuccessCount:   int64(j),
					ErrorCount:     1,
					AverageLatency: time.Duration(j) * time.Millisecond,
					CurrentLoad:    float64(j) / 100.0,
					LastUpdated:    time.Now(),
				}
				
				lb.UpdateMetrics(fmt.Sprintf("node-%d", goroutineID%2+1), metrics)
			}
		}(i)
	}

	wg.Wait()
	
	// Verify state is consistent
	allMetrics := lb.GetMetrics()
	if len(allMetrics) == 0 {
		t.Error("Should have metrics after concurrent updates")
	}
}