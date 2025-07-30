package loadbalancer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/loadbalancer"
)

// TestLoadBalancingAlgorithms tests various load balancing algorithms
func TestLoadBalancingAlgorithms(t *testing.T) {
	t.Run("RoundRobin", testRoundRobinBalancer)
	t.Run("WeightedRoundRobin", testWeightedRoundRobinBalancer)
	t.Run("LeastConnections", testLeastConnectionsBalancer)
	t.Run("WeightedLeastConnections", testWeightedLeastConnectionsBalancer)
	t.Run("IPHash", testIPHashBalancer)
	t.Run("ConsistentHash", testConsistentHashBalancer)
	t.Run("ResourceBased", testResourceBasedBalancer)
	t.Run("AdaptiveBalancer", testAdaptiveBalancer)
}

// testRoundRobinBalancer tests round-robin load balancing
func testRoundRobinBalancer(t *testing.T) {
	// Create nodes
	nodes := createTestNodes(4)
	
	// Create round-robin balancer
	balancer := loadbalancer.NewRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test round-robin distribution
	requestCount := 100
	nodeSelections := make(map[string]int)

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)

		nodeSelections[selectedNode.ID]++
	}

	// Verify even distribution
	expectedPerNode := requestCount / len(nodes)
	for nodeID, count := range nodeSelections {
		assert.Equal(t, expectedPerNode, count, "Node %s should receive equal requests", nodeID)
	}

	// Test with node removal
	removedNode := nodes[0]
	err := balancer.RemoveNode(removedNode.ID)
	assert.NoError(t, err)

	// Test distribution with remaining nodes
	nodeSelections = make(map[string]int)
	remainingNodes := len(nodes) - 1

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("req-after-removal-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotEqual(t, removedNode.ID, selectedNode.ID)

		nodeSelections[selectedNode.ID]++
	}

	// Verify removed node is not selected
	assert.Equal(t, 0, nodeSelections[removedNode.ID])

	// Verify even distribution among remaining nodes
	expectedPerNode = requestCount / remainingNodes
	for _, node := range nodes[1:] { // Skip removed node
		assert.Equal(t, expectedPerNode, nodeSelections[node.ID])
	}
}

// testWeightedRoundRobinBalancer tests weighted round-robin load balancing
func testWeightedRoundRobinBalancer(t *testing.T) {
	// Create nodes with different weights
	nodes := []*loadbalancer.Node{
		{ID: "node-1", Weight: 1, Status: loadbalancer.NodeStatusHealthy},
		{ID: "node-2", Weight: 2, Status: loadbalancer.NodeStatusHealthy},
		{ID: "node-3", Weight: 3, Status: loadbalancer.NodeStatusHealthy},
		{ID: "node-4", Weight: 4, Status: loadbalancer.NodeStatusHealthy},
	}

	balancer := loadbalancer.NewWeightedRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test weighted distribution
	requestCount := 1000 // Use larger number for better distribution accuracy
	nodeSelections := make(map[string]int)

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("weighted-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)

		nodeSelections[selectedNode.ID]++
	}

	// Calculate total weight
	totalWeight := 0
	for _, node := range nodes {
		totalWeight += node.Weight
	}

	// Verify distribution matches weights (with some tolerance)
	tolerance := 0.05 // 5% tolerance
	for _, node := range nodes {
		expectedRatio := float64(node.Weight) / float64(totalWeight)
		expectedCount := int(expectedRatio * float64(requestCount))
		actualCount := nodeSelections[node.ID]
		
		diff := math.Abs(float64(actualCount-expectedCount)) / float64(expectedCount)
		assert.Less(t, diff, tolerance, "Node %s distribution should match weight", node.ID)
	}
}

// testLeastConnectionsBalancer tests least connections load balancing
func testLeastConnectionsBalancer(t *testing.T) {
	nodes := createTestNodes(3)
	
	balancer := loadbalancer.NewLeastConnectionsBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Simulate varying connection loads
	// Add connections to nodes
	balancer.AddConnection(nodes[0].ID, "conn-1")
	balancer.AddConnection(nodes[0].ID, "conn-2")
	balancer.AddConnection(nodes[0].ID, "conn-3") // node-1: 3 connections

	balancer.AddConnection(nodes[1].ID, "conn-4") // node-2: 1 connection
	// node-3: 0 connections

	// Next request should go to node with least connections (node-3)
	request := &loadbalancer.Request{
		ID:   "least-conn-req-1",
		Type: "inference",
	}

	selectedNode, err := balancer.SelectNode(request)
	assert.NoError(t, err)
	assert.Equal(t, nodes[2].ID, selectedNode.ID) // Should select node-3

	// Add connection to selected node
	balancer.AddConnection(selectedNode.ID, "new-conn-1")

	// Next request should go to node-2 (1 connection) or node-3 (1 connection)
	request = &loadbalancer.Request{
		ID:   "least-conn-req-2",
		Type: "inference",
	}

	selectedNode, err = balancer.SelectNode(request)
	assert.NoError(t, err)
	assert.True(t, selectedNode.ID == nodes[1].ID || selectedNode.ID == nodes[2].ID)
	assert.NotEqual(t, nodes[0].ID, selectedNode.ID) // Should not select node-1 (3 connections)

	// Test connection removal
	balancer.RemoveConnection(nodes[0].ID, "conn-1")
	balancer.RemoveConnection(nodes[0].ID, "conn-2") // node-1 now has 1 connection

	// Now all nodes have similar connection counts
	connectionCounts := make(map[string]int)
	for i := 0; i < 30; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("balanced-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		connectionCounts[selectedNode.ID]++
	}

	// Should have relatively even distribution
	for _, count := range connectionCounts {
		assert.Greater(t, count, 5) // Each node should get some requests
	}
}

// testWeightedLeastConnectionsBalancer tests weighted least connections
func testWeightedLeastConnectionsBalancer(t *testing.T) {
	nodes := []*loadbalancer.Node{
		{ID: "node-1", Weight: 1, Status: loadbalancer.NodeStatusHealthy},
		{ID: "node-2", Weight: 2, Status: loadbalancer.NodeStatusHealthy},
		{ID: "node-3", Weight: 3, Status: loadbalancer.NodeStatusHealthy},
	}

	balancer := loadbalancer.NewWeightedLeastConnectionsBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Add different connection loads
	balancer.AddConnection(nodes[0].ID, "conn-1") // node-1: 1 conn, weight 1, ratio 1.0
	balancer.AddConnection(nodes[1].ID, "conn-2") // node-2: 1 conn, weight 2, ratio 0.5
	balancer.AddConnection(nodes[1].ID, "conn-3") // node-2: 2 conn, weight 2, ratio 1.0
	// node-3: 0 conn, weight 3, ratio 0.0

	// Next request should go to node-3 (lowest connection/weight ratio)
	request := &loadbalancer.Request{
		ID:   "weighted-least-req-1",
		Type: "inference",
	}

	selectedNode, err := balancer.SelectNode(request)
	assert.NoError(t, err)
	assert.Equal(t, nodes[2].ID, selectedNode.ID)

	// Test multiple requests to verify balanced distribution
	nodeSelections := make(map[string]int)
	for i := 0; i < 60; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("weighted-least-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		nodeSelections[selectedNode.ID]++

		// Simulate connection lifecycle
		balancer.AddConnection(selectedNode.ID, fmt.Sprintf("temp-conn-%d", i))
		if i%10 == 0 && i > 0 {
			// Remove some connections periodically
			balancer.RemoveConnection(selectedNode.ID, fmt.Sprintf("temp-conn-%d", i-10))
		}
	}

	// Higher weight nodes should receive more requests
	assert.Greater(t, nodeSelections[nodes[2].ID], nodeSelections[nodes[0].ID])
	assert.Greater(t, nodeSelections[nodes[1].ID], nodeSelections[nodes[0].ID])
}

// testIPHashBalancer tests IP hash-based load balancing
func testIPHashBalancer(t *testing.T) {
	nodes := createTestNodes(4)
	
	balancer := loadbalancer.NewIPHashBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test that same IP always goes to same node
	testIPs := []string{
		"192.168.1.100",
		"192.168.1.101",
		"192.168.1.102",
		"10.0.0.1",
		"10.0.0.2",
	}

	ipToNode := make(map[string]string)

	// First pass - record which node each IP gets
	for _, ip := range testIPs {
		request := &loadbalancer.Request{
			ID:       fmt.Sprintf("ip-hash-req-%s", ip),
			Type:     "inference",
			ClientIP: ip,
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)

		ipToNode[ip] = selectedNode.ID
	}

	// Second pass - verify same IPs get same nodes
	for i := 0; i < 10; i++ {
		for _, ip := range testIPs {
			request := &loadbalancer.Request{
				ID:       fmt.Sprintf("ip-hash-req-%s-%d", ip, i),
				Type:     "inference",
				ClientIP: ip,
			}

			selectedNode, err := balancer.SelectNode(request)
			assert.NoError(t, err)
			assert.Equal(t, ipToNode[ip], selectedNode.ID, "IP %s should always go to same node", ip)
		}
	}

	// Test distribution across different IPs
	nodeSelections := make(map[string]int)
	for i := 0; i < 100; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i+1)
		request := &loadbalancer.Request{
			ID:       fmt.Sprintf("dist-req-%d", i),
			Type:     "inference",
			ClientIP: ip,
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		nodeSelections[selectedNode.ID]++
	}

	// Should have reasonable distribution across nodes
	for _, count := range nodeSelections {
		assert.Greater(t, count, 5) // Each node should get some requests
	}
}

// testConsistentHashBalancer tests consistent hash load balancing
func testConsistentHashBalancer(t *testing.T) {
	nodes := createTestNodes(4)
	
	config := &loadbalancer.ConsistentHashConfig{
		VirtualNodes: 100, // Number of virtual nodes per real node
		HashFunction: "sha256",
	}

	balancer := loadbalancer.NewConsistentHashBalancer(config)
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test that same key always goes to same node
	testKeys := []string{
		"model:llama-7b",
		"model:gpt-3.5",
		"model:claude-2",
		"model:mistral-7b",
		"user:alice",
		"user:bob",
		"session:12345",
	}

	keyToNode := make(map[string]string)

	// First pass - record which node each key gets
	for _, key := range testKeys {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("consistent-req-%s", key),
			Type: "inference",
			Key:  key,
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)

		keyToNode[key] = selectedNode.ID
	}

	// Verify consistency across multiple requests
	for i := 0; i < 20; i++ {
		for _, key := range testKeys {
			request := &loadbalancer.Request{
				ID:   fmt.Sprintf("consistent-req-%s-%d", key, i),
				Type: "inference",
				Key:  key,
			}

			selectedNode, err := balancer.SelectNode(request)
			assert.NoError(t, err)
			assert.Equal(t, keyToNode[key], selectedNode.ID, "Key %s should always go to same node", key)
		}
	}

	// Test node addition
	newNode := &loadbalancer.Node{
		ID:     "node-5",
		Status: loadbalancer.NodeStatusHealthy,
		Weight: 1,
	}

	err := balancer.AddNode(newNode)
	assert.NoError(t, err)

	// Some keys may be redistributed, but most should stay
	stableKeys := 0
	for _, key := range testKeys {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("after-add-req-%s", key),
			Type: "inference",
			Key:  key,
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)

		if selectedNode.ID == keyToNode[key] {
			stableKeys++
		}
	}

	// Most keys should remain on their original nodes (minimal disruption)
	assert.GreaterOrEqual(t, stableKeys, len(testKeys)/2)

	// Test node removal
	err = balancer.RemoveNode(nodes[0].ID)
	assert.NoError(t, err)

	// Keys that were on removed node should be redistributed
	for _, key := range testKeys {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("after-remove-req-%s", key),
			Type: "inference",
			Key:  key,
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotEqual(t, nodes[0].ID, selectedNode.ID)
	}
}

// testResourceBasedBalancer tests resource-based load balancing
func testResourceBasedBalancer(t *testing.T) {
	// Create nodes with different resource capacities
	nodes := []*loadbalancer.Node{
		{
			ID:     "node-1",
			Status: loadbalancer.NodeStatusHealthy,
			Resources: &loadbalancer.Resources{
				CPU:    4.0,  // 4 cores
				Memory: 8192, // 8GB
				GPU:    1,    // 1 GPU
			},
			Usage: &loadbalancer.Resources{
				CPU:    2.0,  // 50% CPU used
				Memory: 2048, // 25% memory used
				GPU:    0,    // 0% GPU used
			},
		},
		{
			ID:     "node-2",
			Status: loadbalancer.NodeStatusHealthy,
			Resources: &loadbalancer.Resources{
				CPU:    8.0,   // 8 cores
				Memory: 16384, // 16GB
				GPU:    2,     // 2 GPUs
			},
			Usage: &loadbalancer.Resources{
				CPU:    1.0,  // 12.5% CPU used
				Memory: 4096, // 25% memory used
				GPU:    0,    // 0% GPU used
			},
		},
		{
			ID:     "node-3",
			Status: loadbalancer.NodeStatusHealthy,
			Resources: &loadbalancer.Resources{
				CPU:    2.0,  // 2 cores
				Memory: 4096, // 4GB
				GPU:    0,    // No GPU
			},
			Usage: &loadbalancer.Resources{
				CPU:    1.8,  // 90% CPU used
				Memory: 3072, // 75% memory used
				GPU:    0,
			},
		},
	}

	balancer := loadbalancer.NewResourceBasedBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test CPU-intensive requests
	cpuRequest := &loadbalancer.Request{
		ID:   "cpu-intensive-req",
		Type: "inference",
		ResourceRequirements: &loadbalancer.Resources{
			CPU:    2.0, // Requires 2 CPU cores
			Memory: 1024,
		},
	}

	selectedNode, err := balancer.SelectNode(cpuRequest)
	assert.NoError(t, err)
	// Should prefer node-2 (has most available CPU)
	assert.Equal(t, "node-2", selectedNode.ID)

	// Test GPU-required requests
	gpuRequest := &loadbalancer.Request{
		ID:   "gpu-required-req",
		Type: "inference",
		ResourceRequirements: &loadbalancer.Resources{
			GPU: 1, // Requires 1 GPU
		},
	}

	selectedNode, err = balancer.SelectNode(gpuRequest)
	assert.NoError(t, err)
	// Should go to node with available GPU (node-1 or node-2)
	assert.True(t, selectedNode.ID == "node-1" || selectedNode.ID == "node-2")
	assert.NotEqual(t, "node-3", selectedNode.ID) // node-3 has no GPU

	// Test memory-intensive requests
	memoryRequest := &loadbalancer.Request{
		ID:   "memory-intensive-req",
		Type: "inference",
		ResourceRequirements: &loadbalancer.Resources{
			Memory: 8192, // Requires 8GB
		},
	}

	selectedNode, err = balancer.SelectNode(memoryRequest)
	assert.NoError(t, err)
	// Should prefer node-2 (has most available memory)
	assert.Equal(t, "node-2", selectedNode.ID)

	// Test requests that exceed capacity
	oversizedRequest := &loadbalancer.Request{
		ID:   "oversized-req",
		Type: "inference",
		ResourceRequirements: &loadbalancer.Resources{
			CPU:    10.0, // Requires 10 CPU cores (more than any node has)
			Memory: 32768,
		},
	}

	selectedNode, err = balancer.SelectNode(oversizedRequest)
	assert.Error(t, err) // Should fail - no node can handle this request
	assert.Nil(t, selectedNode)
}

// testAdaptiveBalancer tests adaptive load balancing
func testAdaptiveBalancer(t *testing.T) {
	nodes := createTestNodes(3)
	
	config := &loadbalancer.AdaptiveConfig{
		MetricsWindow:    5 * time.Second,
		AdjustmentFactor: 0.1,
		MinWeight:        0.1,
		MaxWeight:        2.0,
	}

	balancer := loadbalancer.NewAdaptiveBalancer(config)
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Simulate different response times for nodes
	responseTimeSimulator := map[string]time.Duration{
		"node-1": 100 * time.Millisecond, // Fast node
		"node-2": 300 * time.Millisecond, // Medium node
		"node-3": 500 * time.Millisecond, // Slow node
	}

	// Send requests and simulate responses
	requestCount := 100
	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("adaptive-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)

		// Simulate request completion with response time
		responseTime := responseTimeSimulator[selectedNode.ID]
		success := true
		if i%20 == 0 && selectedNode.ID == "node-3" {
			// Simulate occasional failures on slow node
			success = false
		}

		balancer.RecordResponse(selectedNode.ID, request.ID, responseTime, success)
	}

	// Wait for adaptation
	time.Sleep(6 * time.Second)

	// After adaptation, fast nodes should receive more traffic
	nodeSelections := make(map[string]int)
	for i := 0; i < 60; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("post-adaptation-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		nodeSelections[selectedNode.ID]++
	}

	// Fast node should get more requests than slow node
	assert.Greater(t, nodeSelections["node-1"], nodeSelections["node-3"])
	// Medium node should get more requests than slow node
	assert.Greater(t, nodeSelections["node-2"], nodeSelections["node-3"])
}

// TestLoadBalancerHealth tests health checking functionality
func TestLoadBalancerHealth(t *testing.T) {
	t.Run("HealthChecking", testHealthChecking)
	t.Run("UnhealthyNodeExclusion", testUnhealthyNodeExclusion)
	t.Run("NodeRecovery", testNodeRecovery)
	t.Run("HealthCheckFailover", testHealthCheckFailover)
}

// testHealthChecking tests health checking mechanisms
func testHealthChecking(t *testing.T) {
	nodes := createTestNodes(3)
	
	healthConfig := &loadbalancer.HealthConfig{
		CheckInterval:    1 * time.Second,
		Timeout:          500 * time.Millisecond,
		FailureThreshold: 3,
		SuccessThreshold: 2,
	}

	balancer := loadbalancer.NewHealthAwareBalancer(healthConfig)
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Start health checking
	err := balancer.StartHealthChecking()
	assert.NoError(t, err)
	defer balancer.StopHealthChecking()

	// Initially all nodes should be healthy
	time.Sleep(2 * time.Second)
	
	for _, node := range nodes {
		status := balancer.GetNodeStatus(node.ID)
		assert.Equal(t, loadbalancer.NodeStatusHealthy, status)
	}

	// Simulate node failure
	balancer.SimulateNodeFailure(nodes[0].ID)

	// Wait for health check to detect failure
	time.Sleep(4 * time.Second)

	status := balancer.GetNodeStatus(nodes[0].ID)
	assert.Equal(t, loadbalancer.NodeStatusUnhealthy, status)

	// Other nodes should remain healthy
	assert.Equal(t, loadbalancer.NodeStatusHealthy, balancer.GetNodeStatus(nodes[1].ID))
	assert.Equal(t, loadbalancer.NodeStatusHealthy, balancer.GetNodeStatus(nodes[2].ID))
}

// testUnhealthyNodeExclusion tests exclusion of unhealthy nodes
func testUnhealthyNodeExclusion(t *testing.T) {
	nodes := createTestNodes(3)
	
	balancer := loadbalancer.NewRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Mark one node as unhealthy
	err := balancer.SetNodeStatus(nodes[0].ID, loadbalancer.NodeStatusUnhealthy)
	assert.NoError(t, err)

	// Test that unhealthy node is not selected
	nodeSelections := make(map[string]int)
	requestCount := 60

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("exclusion-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)
		assert.NotEqual(t, nodes[0].ID, selectedNode.ID, "Unhealthy node should not be selected")

		nodeSelections[selectedNode.ID]++
	}

	// Verify unhealthy node received no requests
	assert.Equal(t, 0, nodeSelections[nodes[0].ID])

	// Verify healthy nodes received requests
	assert.Greater(t, nodeSelections[nodes[1].ID], 0)
	assert.Greater(t, nodeSelections[nodes[2].ID], 0)

	// Verify even distribution among healthy nodes
	expectedPerNode := requestCount / 2 // Only 2 healthy nodes
	assert.Equal(t, expectedPerNode, nodeSelections[nodes[1].ID])
	assert.Equal(t, expectedPerNode, nodeSelections[nodes[2].ID])
}

// testNodeRecovery tests node recovery functionality
func testNodeRecovery(t *testing.T) {
	nodes := createTestNodes(3)
	
	healthConfig := &loadbalancer.HealthConfig{
		CheckInterval:    500 * time.Millisecond,
		Timeout:          200 * time.Millisecond,
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}

	balancer := loadbalancer.NewHealthAwareBalancer(healthConfig)
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	err := balancer.StartHealthChecking()
	assert.NoError(t, err)
	defer balancer.StopHealthChecking()

	// Simulate node failure
	balancer.SimulateNodeFailure(nodes[0].ID)

	// Wait for failure detection
	time.Sleep(2 * time.Second)
	assert.Equal(t, loadbalancer.NodeStatusUnhealthy, balancer.GetNodeStatus(nodes[0].ID))

	// Simulate node recovery
	balancer.SimulateNodeRecovery(nodes[0].ID)

	// Wait for recovery detection
	time.Sleep(2 * time.Second)
	assert.Equal(t, loadbalancer.NodeStatusHealthy, balancer.GetNodeStatus(nodes[0].ID))

	// Verify recovered node receives requests again
	nodeSelections := make(map[string]int)
	for i := 0; i < 30; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("recovery-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		nodeSelections[selectedNode.ID]++
	}

	// All nodes should receive some requests
	for _, node := range nodes {
		assert.Greater(t, nodeSelections[node.ID], 0, "Recovered node should receive requests")
	}
}

// testHealthCheckFailover tests failover during health check failures
func testHealthCheckFailover(t *testing.T) {
	nodes := createTestNodes(2) // Minimal cluster for failover testing
	
	balancer := loadbalancer.NewRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test normal operation
	request := &loadbalancer.Request{
		ID:   "failover-test-1",
		Type: "inference",
	}

	selectedNode, err := balancer.SelectNode(request)
	assert.NoError(t, err)
	assert.NotNil(t, selectedNode)

	// Mark one node as failed
	failedNodeID := selectedNode.ID
	err = balancer.SetNodeStatus(failedNodeID, loadbalancer.NodeStatusUnhealthy)
	assert.NoError(t, err)

	// Subsequent requests should go to healthy node
	var healthyNodeID string
	for _, node := range nodes {
		if node.ID != failedNodeID {
			healthyNodeID = node.ID
			break
		}
	}

	for i := 0; i < 10; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("failover-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.Equal(t, healthyNodeID, selectedNode.ID)
	}

	// Test complete failure scenario
	err = balancer.SetNodeStatus(healthyNodeID, loadbalancer.NodeStatusUnhealthy)
	assert.NoError(t, err)

	// Should fail when no healthy nodes available
	request = &loadbalancer.Request{
		ID:   "no-healthy-nodes",
		Type: "inference",
	}

	selectedNode, err = balancer.SelectNode(request)
	assert.Error(t, err)
	assert.Nil(t, selectedNode)
}

// TestLoadBalancerPerformance tests performance characteristics
func TestLoadBalancerPerformance(t *testing.T) {
	t.Run("HighThroughput", testHighThroughputBalancing)
	t.Run("ConcurrentRequests", testConcurrentRequestBalancing)
	t.Run("ScalabilityTest", testLoadBalancerScalability)
	t.Run("MemoryUsage", testLoadBalancerMemoryUsage)
}

// testHighThroughputBalancing tests high throughput scenarios
func testHighThroughputBalancing(t *testing.T) {
	nodes := createTestNodes(5)
	
	balancer := loadbalancer.NewRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Measure throughput
	requestCount := 10000
	start := time.Now()

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("throughput-req-%d", i),
			Type: "inference",
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)
	}

	duration := time.Since(start)
	throughput := float64(requestCount) / duration.Seconds()

	t.Logf("Load balancer throughput: %.2f requests/second", throughput)
	
	// Should achieve high throughput
	assert.Greater(t, throughput, 10000.0, "Load balancer should handle at least 10k req/sec")
}

// testConcurrentRequestBalancing tests concurrent request handling
func testConcurrentRequestBalancing(t *testing.T) {
	nodes := createTestNodes(4)
	
	balancer := loadbalancer.NewRoundRobinBalancer()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Test concurrent requests
	concurrency := 100
	requestsPerGoroutine := 100
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	nodeSelections := make(map[string]int)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			localSelections := make(map[string]int)
			
			for j := 0; j < requestsPerGoroutine; j++ {
				request := &loadbalancer.Request{
					ID:   fmt.Sprintf("concurrent-req-%d-%d", goroutineID, j),
					Type: "inference",
				}

				selectedNode, err := balancer.SelectNode(request)
				assert.NoError(t, err)
				assert.NotNil(t, selectedNode)

				localSelections[selectedNode.ID]++
			}

			// Merge local results
			mu.Lock()
			for nodeID, count := range localSelections {
				nodeSelections[nodeID] += count
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify even distribution
	totalRequests := concurrency * requestsPerGoroutine
	expectedPerNode := totalRequests / len(nodes)
	tolerance := 0.1 // 10% tolerance

	for _, node := range nodes {
		count := nodeSelections[node.ID]
		diff := math.Abs(float64(count-expectedPerNode)) / float64(expectedPerNode)
		assert.Less(t, diff, tolerance, "Node %s should receive even distribution", node.ID)
	}
}

// testLoadBalancerScalability tests scalability with many nodes
func testLoadBalancerScalability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	// Test with large number of nodes
	nodeCount := 100
	nodes := createTestNodes(nodeCount)
	
	balancer := loadbalancer.NewConsistentHashBalancer(&loadbalancer.ConsistentHashConfig{
		VirtualNodes: 50,
		HashFunction: "sha256",
	})

	// Measure node addition time
	start := time.Now()
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}
	additionTime := time.Since(start)

	t.Logf("Added %d nodes in %v (%.2f nodes/sec)", nodeCount, additionTime, float64(nodeCount)/additionTime.Seconds())

	// Measure selection performance with many nodes
	requestCount := 1000
	start = time.Now()

	for i := 0; i < requestCount; i++ {
		request := &loadbalancer.Request{
			ID:   fmt.Sprintf("scale-req-%d", i),
			Type: "inference",
			Key:  fmt.Sprintf("key-%d", i),
		}

		selectedNode, err := balancer.SelectNode(request)
		assert.NoError(t, err)
		assert.NotNil(t, selectedNode)
	}

	selectionTime := time.Since(start)
	selectionThroughput := float64(requestCount) / selectionTime.Seconds()

	t.Logf("Selection throughput with %d nodes: %.2f req/sec", nodeCount, selectionThroughput)

	// Should maintain reasonable performance even with many nodes
	assert.Greater(t, selectionThroughput, 1000.0, "Should maintain good performance with many nodes")
}

// testLoadBalancerMemoryUsage tests memory usage characteristics
func testLoadBalancerMemoryUsage(t *testing.T) {
	nodes := createTestNodes(50)
	
	balancer := loadbalancer.NewWeightedRoundRobinBalancer()
	
	// Measure initial memory
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Add nodes
	for _, node := range nodes {
		err := balancer.AddNode(node)
		assert.NoError(t, err)
	}

	// Add many connections to track
	for i := 0; i < 1000; i++ {
		nodeID := nodes[i%len(nodes)].ID
		connID := fmt.Sprintf("conn-%d", i)
		balancer.AddConnection(nodeID, connID)
	}

	// Measure memory after operations
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryUsed := m2.Alloc - m1.Alloc
	t.Logf("Memory used for %d nodes and 1000 connections: %d bytes", len(nodes), memoryUsed)

	// Memory usage should be reasonable
	assert.Less(t, memoryUsed, uint64(10*1024*1024), "Memory usage should be less than 10MB")

	// Test memory cleanup
	for _, node := range nodes[:25] { // Remove half the nodes
		err := balancer.RemoveNode(node.ID)
		assert.NoError(t, err)
	}

	// Memory should be freed (this is hard to test precisely due to GC)
	runtime.GC()
	var m3 runtime.MemStats
	runtime.ReadMemStats(&m3)

	t.Logf("Memory after cleanup: %d bytes", m3.Alloc)
}

// Helper functions

// createTestNodes creates test nodes for load balancer testing
func createTestNodes(count int) []*loadbalancer.Node {
	nodes := make([]*loadbalancer.Node, count)
	
	for i := 0; i < count; i++ {
		nodes[i] = &loadbalancer.Node{
			ID:     fmt.Sprintf("node-%d", i+1),
			Status: loadbalancer.NodeStatusHealthy,
			Weight: 1,
			Address: fmt.Sprintf("192.168.1.%d:8080", i+1),
			Resources: &loadbalancer.Resources{
				CPU:    4.0,
				Memory: 8192,
				GPU:    1,
			},
			Usage: &loadbalancer.Resources{
				CPU:    0.0,
				Memory: 0,
				GPU:    0,
			},
			Metrics: &loadbalancer.NodeMetrics{
				RequestCount:     0,
				ErrorCount:       0,
				AverageLatency:   100 * time.Millisecond,
				LastHealthCheck:  time.Now(),
				ConnectionCount:  0,
			},
		}
	}
	
	return nodes
}

// BenchmarkLoadBalancers benchmarks different load balancing algorithms
func BenchmarkLoadBalancers(b *testing.B) {
	nodes := createTestNodes(10)

	b.Run("RoundRobin", func(b *testing.B) {
		balancer := loadbalancer.NewRoundRobinBalancer()
		for _, node := range nodes {
			balancer.AddNode(node)
		}

		request := &loadbalancer.Request{
			ID:   "bench-req",
			Type: "inference",
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := balancer.SelectNode(request)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("WeightedRoundRobin", func(b *testing.B) {
		balancer := loadbalancer.NewWeightedRoundRobinBalancer()
		for _, node := range nodes {
			balancer.AddNode(node)
		}

		request := &loadbalancer.Request{
			ID:   "bench-req",
			Type: "inference",
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := balancer.SelectNode(request)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("LeastConnections", func(b *testing.B) {
		balancer := loadbalancer.NewLeastConnectionsBalancer()
		for _, node := range nodes {
			balancer.AddNode(node)
		}

		request := &loadbalancer.Request{
			ID:   "bench-req",
			Type: "inference",
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := balancer.SelectNode(request)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})

	b.Run("ConsistentHash", func(b *testing.B) {
		config := &loadbalancer.ConsistentHashConfig{
			VirtualNodes: 100,
			HashFunction: "sha256",
		}
		balancer := loadbalancer.NewConsistentHashBalancer(config)
		for _, node := range nodes {
			balancer.AddNode(node)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				request := &loadbalancer.Request{
					ID:   fmt.Sprintf("bench-req-%d", i),
					Type: "inference",
					Key:  fmt.Sprintf("key-%d", i),
				}
				_, err := balancer.SelectNode(request)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	b.Run("ResourceBased", func(b *testing.B) {
		balancer := loadbalancer.NewResourceBasedBalancer()
		for _, node := range nodes {
			balancer.AddNode(node)
		}

		request := &loadbalancer.Request{
			ID:   "bench-req",
			Type: "inference",
			ResourceRequirements: &loadbalancer.Resources{
				CPU:    1.0,
				Memory: 1024,
			},
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := balancer.SelectNode(request)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}