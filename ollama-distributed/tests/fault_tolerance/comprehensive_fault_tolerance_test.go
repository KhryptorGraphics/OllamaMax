package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
)

// TestFaultToleranceManager tests the fault tolerance manager
func TestFaultToleranceManager(t *testing.T) {
	t.Run("ManagerCreation", testManagerCreation)
	t.Run("NodeFailureDetection", testNodeFailureDetection)
	t.Run("NodeRecovery", testNodeRecovery)
	t.Run("AutoFailover", testAutoFailover)
	t.Run("ManualFailover", testManualFailover)
}

// testManagerCreation tests fault tolerance manager creation
func testManagerCreation(t *testing.T) {
	config := &fault_tolerance.Config{
		HeartbeatInterval:    1 * time.Second,
		FailureTimeout:       5 * time.Second,
		RecoveryTimeout:      10 * time.Second,
		MaxRetries:           3,
		BackoffMultiplier:    2.0,
		EnableAutoFailover:   true,
		EnableAutoRecovery:   true,
		HealthCheckEndpoint:  "/health",
		HealthCheckTimeout:   2 * time.Second,
	}

	manager, err := fault_tolerance.NewManager(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.False(t, manager.IsStarted())
}

// testNodeFailureDetection tests node failure detection mechanisms
func testNodeFailureDetection(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Stop()

	// Add test nodes
	nodes := createTestNodes(3)
	for _, node := range nodes {
		err := manager.AddNode(node)
		assert.NoError(t, err)
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for initial health checks
	time.Sleep(2 * time.Second)

	// All nodes should be healthy initially
	for _, node := range nodes {
		status := manager.GetNodeStatus(node.ID)
		assert.Equal(t, fault_tolerance.NodeStatusHealthy, status)
	}

	// Simulate node failure
	manager.SimulateNodeFailure(nodes[0].ID)

	// Wait for failure detection
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(nodes[0].ID)
		return status == fault_tolerance.NodeStatusFailed
	}, 10*time.Second, 500*time.Millisecond, "Failed node should be detected")

	// Other nodes should remain healthy
	assert.Equal(t, fault_tolerance.NodeStatusHealthy, manager.GetNodeStatus(nodes[1].ID))
	assert.Equal(t, fault_tolerance.NodeStatusHealthy, manager.GetNodeStatus(nodes[2].ID))

	// Verify failure event was recorded
	events := manager.GetFailureEvents()
	assert.NotEmpty(t, events)
	
	var failureEvent *fault_tolerance.FailureEvent
	for _, event := range events {
		if event.NodeID == nodes[0].ID && event.Type == fault_tolerance.EventTypeNodeFailure {
			failureEvent = event
			break
		}
	}
	assert.NotNil(t, failureEvent)
	assert.Equal(t, nodes[0].ID, failureEvent.NodeID)
}

// testNodeRecovery tests node recovery mechanisms
func testNodeRecovery(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Stop()

	nodes := createTestNodes(2)
	for _, node := range nodes {
		err := manager.AddNode(node)
		assert.NoError(t, err)
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for initial health checks
	time.Sleep(2 * time.Second)

	// Simulate node failure
	manager.SimulateNodeFailure(nodes[0].ID)

	// Wait for failure detection
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(nodes[0].ID)
		return status == fault_tolerance.NodeStatusFailed
	}, 10*time.Second, 500*time.Millisecond)

	// Simulate node recovery
	manager.SimulateNodeRecovery(nodes[0].ID)

	// Wait for recovery detection
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(nodes[0].ID)
		return status == fault_tolerance.NodeStatusHealthy
	}, 15*time.Second, 500*time.Millisecond, "Recovered node should be detected as healthy")

	// Verify recovery event was recorded
	events := manager.GetFailureEvents()
	var recoveryEvent *fault_tolerance.FailureEvent
	for _, event := range events {
		if event.NodeID == nodes[0].ID && event.Type == fault_tolerance.EventTypeNodeRecovery {
			recoveryEvent = event
			break
		}
	}
	assert.NotNil(t, recoveryEvent)
	assert.Equal(t, nodes[0].ID, recoveryEvent.NodeID)
}

// testAutoFailover tests automatic failover functionality
func testAutoFailover(t *testing.T) {
	config := &fault_tolerance.Config{
		HeartbeatInterval:    500 * time.Millisecond,
		FailureTimeout:       2 * time.Second,
		EnableAutoFailover:   true,
		FailoverStrategy:     fault_tolerance.StrategyFastFailover,
		ReplicationFactor:    2,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	// Create primary and backup nodes
	primaryNode := &fault_tolerance.Node{
		ID:       "primary-1",
		Role:     fault_tolerance.RolePrimary,
		Status:   fault_tolerance.NodeStatusHealthy,
		Address:  "192.168.1.100:8080",
		Capacity: 100,
	}

	backupNode := &fault_tolerance.Node{
		ID:       "backup-1",
		Role:     fault_tolerance.RoleBackup,
		Status:   fault_tolerance.NodeStatusHealthy,
		Address:  "192.168.1.101:8080",
		Capacity: 80,
	}

	err = manager.AddNode(primaryNode)
	require.NoError(t, err)
	
	err = manager.AddNode(backupNode)
	require.NoError(t, err)

	// Create service with replication
	service := &fault_tolerance.Service{
		ID:                "test-service",
		PrimaryNodeID:     primaryNode.ID,
		BackupNodeIDs:     []string{backupNode.ID},
		ReplicationFactor: 2,
		FailoverEnabled:   true,
	}

	err = manager.AddService(service)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Wait for service initialization
	time.Sleep(1 * time.Second)

	// Verify initial setup
	assert.Equal(t, primaryNode.ID, manager.GetServicePrimary(service.ID))

	// Simulate primary node failure
	manager.SimulateNodeFailure(primaryNode.ID)

	// Wait for automatic failover
	assert.Eventually(t, func() bool {
		newPrimary := manager.GetServicePrimary(service.ID)
		return newPrimary == backupNode.ID
	}, 10*time.Second, 500*time.Millisecond, "Backup should become primary after failover")

	// Verify failover event
	events := manager.GetFailoverEvents()
	assert.NotEmpty(t, events)
	
	var failoverEvent *fault_tolerance.FailoverEvent
	for _, event := range events {
		if event.ServiceID == service.ID && event.Type == fault_tolerance.FailoverTypeAutomatic {
			failoverEvent = event
			break
		}
	}
	assert.NotNil(t, failoverEvent)
	assert.Equal(t, primaryNode.ID, failoverEvent.FromNodeID)
	assert.Equal(t, backupNode.ID, failoverEvent.ToNodeID)

	// Verify service is still available
	assert.True(t, manager.IsServiceAvailable(service.ID))
}

// testManualFailover tests manual failover functionality
func testManualFailover(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Stop()

	// Create nodes
	nodes := createTestNodes(3)
	for _, node := range nodes {
		err := manager.AddNode(node)
		require.NoError(t, err)
	}

	// Create service
	service := &fault_tolerance.Service{
		ID:                "manual-service",
		PrimaryNodeID:     nodes[0].ID,
		BackupNodeIDs:     []string{nodes[1].ID, nodes[2].ID},
		ReplicationFactor: 2,
		FailoverEnabled:   true,
	}

	err := manager.AddService(service)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Wait for initialization
	time.Sleep(1 * time.Second)

	// Perform manual failover
	err = manager.ManualFailover(service.ID, nodes[1].ID)
	assert.NoError(t, err)

	// Verify failover completed
	assert.Eventually(t, func() bool {
		primary := manager.GetServicePrimary(service.ID)
		return primary == nodes[1].ID
	}, 5*time.Second, 500*time.Millisecond)

	// Verify manual failover event
	events := manager.GetFailoverEvents()
	var manualEvent *fault_tolerance.FailoverEvent
	for _, event := range events {
		if event.ServiceID == service.ID && event.Type == fault_tolerance.FailoverTypeManual {
			manualEvent = event
			break
		}
	}
	assert.NotNil(t, manualEvent)
	assert.Equal(t, nodes[0].ID, manualEvent.FromNodeID)
	assert.Equal(t, nodes[1].ID, manualEvent.ToNodeID)
}

// TestNetworkPartitions tests network partition handling
func TestNetworkPartitions(t *testing.T) {
	t.Run("SplitBrainPrevention", testSplitBrainPrevention)
	t.Run("PartitionRecovery", testPartitionRecovery)
	t.Run("QuorumMaintenance", testQuorumMaintenance)
	t.Run("NetworkHealing", testNetworkHealing)
}

// testSplitBrainPrevention tests split-brain scenario prevention
func testSplitBrainPrevention(t *testing.T) {
	config := &fault_tolerance.Config{
		EnableQuorum:        true,
		QuorumSize:          3,
		SplitBrainProtection: true,
		HeartbeatInterval:   500 * time.Millisecond,
		PartitionTimeout:    2 * time.Second,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	// Create 5-node cluster
	nodes := createTestNodes(5)
	for _, node := range nodes {
		err = manager.AddNode(node)
		require.NoError(t, err)
	}

	err = manager.Start()
	require.NoError(t, err)

	// Wait for cluster formation
	time.Sleep(2 * time.Second)

	// Verify initial quorum
	assert.True(t, manager.HasQuorum())

	// Simulate network partition: 3 nodes vs 2 nodes
	partition1 := []string{nodes[0].ID, nodes[1].ID, nodes[2].ID} // Majority
	partition2 := []string{nodes[3].ID, nodes[4].ID}             // Minority

	manager.SimulateNetworkPartition(partition1, partition2)

	// Wait for partition detection
	time.Sleep(3 * time.Second)

	// Majority partition should maintain quorum and leadership
	assert.True(t, manager.HasQuorum())
	
	// Only majority partition should accept writes
	err = manager.TestWrite("partition-test", "majority-data")
	assert.NoError(t, err, "Majority partition should accept writes")

	// Minority partition should reject writes (split-brain prevention)
	manager.SwitchToPartition(partition2)
	assert.False(t, manager.HasQuorum())
	
	err = manager.TestWrite("partition-test", "minority-data")
	assert.Error(t, err, "Minority partition should reject writes")

	// Switch back to majority partition
	manager.SwitchToPartition(partition1)
	assert.True(t, manager.HasQuorum())

	// Verify data consistency
	value, err := manager.TestRead("partition-test")
	assert.NoError(t, err)
	assert.Equal(t, "majority-data", value)
}

// testPartitionRecovery tests recovery from network partitions
func testPartitionRecovery(t *testing.T) {
	manager := createTestManagerWithQuorum(t, 5)
	defer manager.Stop()

	// Add test data before partition
	err := manager.TestWrite("pre-partition", "initial-data")
	assert.NoError(t, err)

	// Create partition
	nodes := manager.GetAllNodes()
	partition1 := nodes[0:3] // Majority
	partition2 := nodes[3:5] // Minority

	manager.SimulateNetworkPartition(getNodeIDs(partition1), getNodeIDs(partition2))

	// Wait for partition
	time.Sleep(2 * time.Second)

	// Write data in majority partition
	err = manager.TestWrite("during-partition", "majority-update")
	assert.NoError(t, err)

	// Heal partition
	manager.HealNetworkPartition()

	// Wait for recovery
	assert.Eventually(t, func() bool {
		return manager.IsClusterHealthy()
	}, 10*time.Second, 500*time.Millisecond, "Cluster should recover after partition healing")

	// Verify data consistency across all nodes
	for _, node := range nodes {
		manager.SwitchToNode(node.ID)
		
		value, err := manager.TestRead("pre-partition")
		assert.NoError(t, err)
		assert.Equal(t, "initial-data", value)
		
		value, err = manager.TestRead("during-partition")
		assert.NoError(t, err)
		assert.Equal(t, "majority-update", value)
	}
}

// testQuorumMaintenance tests quorum maintenance during failures
func testQuorumMaintenance(t *testing.T) {
	manager := createTestManagerWithQuorum(t, 5)
	defer manager.Stop()

	// Initial quorum should be established
	assert.True(t, manager.HasQuorum())

	nodes := manager.GetAllNodes()

	// Fail one node (still have quorum: 4/5)
	manager.SimulateNodeFailure(nodes[0].ID)
	time.Sleep(2 * time.Second)
	assert.True(t, manager.HasQuorum(), "Should maintain quorum with 4/5 nodes")

	// Fail another node (still have quorum: 3/5)
	manager.SimulateNodeFailure(nodes[1].ID)
	time.Sleep(2 * time.Second)
	assert.True(t, manager.HasQuorum(), "Should maintain quorum with 3/5 nodes")

	// Fail third node (lose quorum: 2/5)
	manager.SimulateNodeFailure(nodes[2].ID)
	time.Sleep(2 * time.Second)
	assert.False(t, manager.HasQuorum(), "Should lose quorum with 2/5 nodes")

	// Operations should be rejected without quorum
	err := manager.TestWrite("no-quorum", "should-fail")
	assert.Error(t, err, "Operations should fail without quorum")

	// Recover one node (regain quorum: 3/5)
	manager.SimulateNodeRecovery(nodes[2].ID)
	
	assert.Eventually(t, func() bool {
		return manager.HasQuorum()
	}, 10*time.Second, 500*time.Millisecond, "Should regain quorum when node recovers")

	// Operations should work again
	err = manager.TestWrite("quorum-restored", "success")
	assert.NoError(t, err, "Operations should work with restored quorum")
}

// testNetworkHealing tests network healing mechanisms
func testNetworkHealing(t *testing.T) {
	manager := createTestManagerWithQuorum(t, 3)
	defer manager.Stop()

	nodes := manager.GetAllNodes()

	// Create intermittent network issues
	for cycle := 0; cycle < 3; cycle++ {
		t.Logf("Network healing cycle %d", cycle)

		// Create temporary partition
		partition1 := []string{nodes[0].ID}
		partition2 := []string{nodes[1].ID, nodes[2].ID}
		
		manager.SimulateNetworkPartition(partition1, partition2)
		time.Sleep(1 * time.Second)

		// Heal partition
		manager.HealNetworkPartition()
		
		// Wait for healing
		assert.Eventually(t, func() bool {
			return manager.IsClusterHealthy()
		}, 5*time.Second, 200*time.Millisecond, "Network should heal quickly")

		// Verify cluster is operational
		err := manager.TestWrite(fmt.Sprintf("heal-test-%d", cycle), fmt.Sprintf("cycle-%d", cycle))
		assert.NoError(t, err, "Cluster should be operational after healing")
	}

	// Verify all data is consistent
	for cycle := 0; cycle < 3; cycle++ {
		key := fmt.Sprintf("heal-test-%d", cycle)
		expectedValue := fmt.Sprintf("cycle-%d", cycle)
		
		for _, node := range nodes {
			manager.SwitchToNode(node.ID)
			value, err := manager.TestRead(key)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue, value)
		}
	}
}

// TestCascadingFailures tests cascading failure scenarios
func TestCascadingFailures(t *testing.T) {
	t.Run("LoadRedistribution", testLoadRedistribution)
	t.Run("OverloadPrevention", testOverloadPrevention)
	t.Run("GracefulDegradation", testGracefulDegradation)
	t.Run("CircuitBreaker", testCircuitBreaker)
}

// testLoadRedistribution tests load redistribution after failures
func testLoadRedistribution(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Stop()

	// Create nodes with different capacities
	nodes := []*fault_tolerance.Node{
		{ID: "node-1", Capacity: 100, Load: 30, Status: fault_tolerance.NodeStatusHealthy},
		{ID: "node-2", Capacity: 100, Load: 40, Status: fault_tolerance.NodeStatusHealthy},
		{ID: "node-3", Capacity: 100, Load: 50, Status: fault_tolerance.NodeStatusHealthy},
		{ID: "node-4", Capacity: 100, Load: 20, Status: fault_tolerance.NodeStatusHealthy},
	}

	for _, node := range nodes {
		err := manager.AddNode(node)
		require.NoError(t, err)
	}

	err := manager.Start()
	require.NoError(t, err)

	// Record initial load distribution
	initialLoad := make(map[string]int)
	for _, node := range nodes {
		initialLoad[node.ID] = node.Load
	}

	// Simulate failure of heavily loaded node
	manager.SimulateNodeFailure(nodes[2].ID) // node-3 with load 50

	// Wait for load redistribution
	time.Sleep(3 * time.Second)

	// Verify load was redistributed to remaining nodes
	remainingNodes := []*fault_tolerance.Node{nodes[0], nodes[1], nodes[3]}
	totalRedistributedLoad := 0

	for _, node := range remainingNodes {
		currentLoad := manager.GetNodeLoad(node.ID)
		additionalLoad := currentLoad - initialLoad[node.ID]
		totalRedistributedLoad += additionalLoad
		
		// Node should not be overloaded
		assert.LessOrEqual(t, currentLoad, node.Capacity, "Node %s should not be overloaded", node.ID)
	}

	// Total redistributed load should approximately match failed node's load
	assert.GreaterOrEqual(t, totalRedistributedLoad, 40, "Load should be redistributed")

	// Test recovery and load rebalancing
	manager.SimulateNodeRecovery(nodes[2].ID)
	
	assert.Eventually(t, func() bool {
		// Check if load is rebalanced
		maxLoad := 0
		minLoad := 100
		for _, node := range nodes {
			load := manager.GetNodeLoad(node.ID)
			if load > maxLoad {
				maxLoad = load
			}
			if load < minLoad {
				minLoad = load
			}
		}
		// Load should be more evenly distributed
		return (maxLoad - minLoad) < 30
	}, 10*time.Second, 500*time.Millisecond, "Load should rebalance after recovery")
}

// testOverloadPrevention tests overload prevention mechanisms
func testOverloadPrevention(t *testing.T) {
	config := &fault_tolerance.Config{
		OverloadProtection:    true,
		MaxLoadPerNode:        80, // 80% capacity
		LoadBalancingStrategy: fault_tolerance.StrategyLeastLoaded,
		OverloadThreshold:     0.9, // 90% of max load
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	// Create nodes
	nodes := []*fault_tolerance.Node{
		{ID: "node-1", Capacity: 100, Load: 75, Status: fault_tolerance.NodeStatusHealthy}, // Near limit
		{ID: "node-2", Capacity: 100, Load: 10, Status: fault_tolerance.NodeStatusHealthy}, // Low load
		{ID: "node-3", Capacity: 100, Load: 85, Status: fault_tolerance.NodeStatusHealthy}, // Overloaded
	}

	for _, node := range nodes {
		err = manager.AddNode(node)
		require.NoError(t, err)
	}

	err = manager.Start()
	require.NoError(t, err)

	// Attempt to add load that would cause overload
	request := &fault_tolerance.LoadRequest{
		ID:       "overload-test",
		Load:     20, // Would overload node-1 (75+20=95 > 80)
		Priority: fault_tolerance.PriorityNormal,
	}

	// Request should be routed to node-2 (low load)
	assignedNode, err := manager.AssignLoad(request)
	assert.NoError(t, err)
	assert.Equal(t, "node-2", assignedNode.ID, "Load should be assigned to least loaded node")

	// Verify overloaded node is not assigned new load
	request = &fault_tolerance.LoadRequest{
		ID:       "overload-test-2",
		Load:     10,
		Priority: fault_tolerance.PriorityNormal,
	}

	assignedNode, err = manager.AssignLoad(request)
	assert.NoError(t, err)
	assert.NotEqual(t, "node-3", assignedNode.ID, "Overloaded node should not receive new load")

	// Test emergency load shedding
	request = &fault_tolerance.LoadRequest{
		ID:       "emergency-test",
		Load:     100, // Very high load
		Priority: fault_tolerance.PriorityCritical,
	}

	// Should trigger load shedding on overloaded nodes
	assignedNode, err = manager.AssignLoad(request)
	if err == nil {
		// If assigned, verify load shedding occurred
		assert.Eventually(t, func() bool {
			return manager.GetNodeLoad("node-3") < 80
		}, 5*time.Second, 500*time.Millisecond, "Overloaded node should shed load")
	}
}

// testGracefulDegradation tests graceful degradation under failures
func testGracefulDegradation(t *testing.T) {
	manager := createTestManager(t)
	defer manager.Stop()

	// Create tiered service levels
	services := []*fault_tolerance.Service{
		{
			ID:           "critical-service",
			Priority:     fault_tolerance.PriorityCritical,
			MinNodes:     2,
			DesiredNodes: 3,
		},
		{
			ID:           "important-service",
			Priority:     fault_tolerance.PriorityHigh,
			MinNodes:     1,
			DesiredNodes: 2,
		},
		{
			ID:           "optional-service",
			Priority:     fault_tolerance.PriorityLow,
			MinNodes:     1,
			DesiredNodes: 1,
		},
	}

	nodes := createTestNodes(4)
	for _, node := range nodes {
		err := manager.AddNode(node)
		require.NoError(t, err)
	}

	for _, service := range services {
		err := manager.AddService(service)
		require.NoError(t, err)
	}

	err := manager.Start()
	require.NoError(t, err)

	// Wait for initial allocation
	time.Sleep(2 * time.Second)

	// All services should be running at desired levels
	for _, service := range services {
		assert.True(t, manager.IsServiceRunning(service.ID))
		runningNodes := manager.GetServiceNodes(service.ID)
		assert.GreaterOrEqual(t, len(runningNodes), service.MinNodes)
	}

	// Simulate gradual node failures
	for i, node := range nodes[:2] { // Fail 2 out of 4 nodes
		t.Logf("Failing node %s (%d/2)", node.ID, i+1)
		manager.SimulateNodeFailure(node.ID)
		time.Sleep(2 * time.Second)

		// Critical service should maintain minimum nodes
		criticalNodes := manager.GetServiceNodes("critical-service")
		assert.GreaterOrEqual(t, len(criticalNodes), 2, "Critical service should maintain minimum nodes")

		// Lower priority services may be degraded
		if i == 1 { // After second failure
			optionalNodes := manager.GetServiceNodes("optional-service")
			// Optional service might be stopped to preserve resources for critical services
			if len(optionalNodes) == 0 {
				t.Log("Optional service gracefully degraded")
			}
		}
	}

	// At least critical service should still be running
	assert.True(t, manager.IsServiceRunning("critical-service"))

	// Test recovery and service restoration
	manager.SimulateNodeRecovery(nodes[0].ID)
	
	assert.Eventually(t, func() bool {
		// All services should be restored
		for _, service := range services {
			if !manager.IsServiceRunning(service.ID) {
				return false
			}
		}
		return true
	}, 10*time.Second, 500*time.Millisecond, "All services should be restored after recovery")
}

// testCircuitBreaker tests circuit breaker functionality
func testCircuitBreaker(t *testing.T) {
	config := &fault_tolerance.Config{
		CircuitBreaker: &fault_tolerance.CircuitBreakerConfig{
			Enabled:           true,
			FailureThreshold:  5,
			SuccessThreshold:  3,
			Timeout:           2 * time.Second,
			HalfOpenMaxCalls:  2,
		},
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	node := &fault_tolerance.Node{
		ID:     "circuit-test-node",
		Status: fault_tolerance.NodeStatusHealthy,
	}

	err = manager.AddNode(node)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Simulate repeated failures to trip circuit breaker
	for i := 0; i < 6; i++ {
		request := &fault_tolerance.Request{
			ID:       fmt.Sprintf("failing-req-%d", i),
			NodeID:   node.ID,
			Timeout:  1 * time.Second,
		}

		// Simulate failure
		manager.SimulateRequestFailure(request)
	}

	// Circuit breaker should be open (blocking requests)
	assert.Eventually(t, func() bool {
		state := manager.GetCircuitBreakerState(node.ID)
		return state == fault_tolerance.CircuitBreakerStateOpen
	}, 3*time.Second, 100*time.Millisecond, "Circuit breaker should be open")

	// Requests should be rejected immediately
	request := &fault_tolerance.Request{
		ID:     "blocked-request",
		NodeID: node.ID,
	}

	response, err := manager.SendRequest(request)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "circuit breaker")

	// Wait for half-open state
	time.Sleep(3 * time.Second)
	
	state := manager.GetCircuitBreakerState(node.ID)
	assert.Equal(t, fault_tolerance.CircuitBreakerStateHalfOpen, state)

	// Send successful requests to close circuit
	for i := 0; i < 3; i++ {
		request := &fault_tolerance.Request{
			ID:     fmt.Sprintf("success-req-%d", i),
			NodeID: node.ID,
		}

		manager.SimulateRequestSuccess(request)
	}

	// Circuit breaker should be closed (allowing requests)
	assert.Eventually(t, func() bool {
		state := manager.GetCircuitBreakerState(node.ID)
		return state == fault_tolerance.CircuitBreakerStateClosed
	}, 2*time.Second, 100*time.Millisecond, "Circuit breaker should be closed")

	// Requests should work normally
	request = &fault_tolerance.Request{
		ID:     "normal-request",
		NodeID: node.ID,
	}

	response, err = manager.SendRequest(request)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

// TestRecoveryStrategies tests different recovery strategies
func TestRecoveryStrategies(t *testing.T) {
	t.Run("ImmediateRecovery", testImmediateRecovery)
	t.Run("GradualRecovery", testGradualRecovery)
	t.Run("BackoffRecovery", testBackoffRecovery)
	t.Run("HealthBasedRecovery", testHealthBasedRecovery)
}

// testImmediateRecovery tests immediate recovery strategy
func testImmediateRecovery(t *testing.T) {
	config := &fault_tolerance.Config{
		RecoveryStrategy: fault_tolerance.RecoveryStrategyImmediate,
		RecoveryTimeout:  1 * time.Second,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	node := createTestNodes(1)[0]
	err = manager.AddNode(node)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Simulate failure and immediate recovery
	manager.SimulateNodeFailure(node.ID)
	time.Sleep(1 * time.Second)

	assert.Equal(t, fault_tolerance.NodeStatusFailed, manager.GetNodeStatus(node.ID))

	manager.SimulateNodeRecovery(node.ID)

	// Should recover immediately
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(node.ID)
		return status == fault_tolerance.NodeStatusHealthy
	}, 2*time.Second, 100*time.Millisecond, "Node should recover immediately")
}

// testGradualRecovery tests gradual recovery strategy
func testGradualRecovery(t *testing.T) {
	config := &fault_tolerance.Config{
		RecoveryStrategy:    fault_tolerance.RecoveryStrategyGradual,
		GradualRecoveryRate: 0.2, // 20% capacity increase per step
		RecoveryStepInterval: 500 * time.Millisecond,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	node := &fault_tolerance.Node{
		ID:       "gradual-node",
		Capacity: 100,
		Status:   fault_tolerance.NodeStatusHealthy,
	}

	err = manager.AddNode(node)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Simulate failure and recovery
	manager.SimulateNodeFailure(node.ID)
	time.Sleep(1 * time.Second)

	manager.SimulateNodeRecovery(node.ID)

	// Should gradually increase capacity
	time.Sleep(600 * time.Millisecond) // First step
	capacity1 := manager.GetNodeEffectiveCapacity(node.ID)
	assert.Equal(t, 20, capacity1) // 20% of 100

	time.Sleep(500 * time.Millisecond) // Second step
	capacity2 := manager.GetNodeEffectiveCapacity(node.ID)
	assert.Equal(t, 40, capacity2) // 40% of 100

	// Wait for full recovery
	assert.Eventually(t, func() bool {
		capacity := manager.GetNodeEffectiveCapacity(node.ID)
		return capacity == 100
	}, 5*time.Second, 100*time.Millisecond, "Node should reach full capacity")
}

// testBackoffRecovery tests exponential backoff recovery
func testBackoffRecovery(t *testing.T) {
	config := &fault_tolerance.Config{
		RecoveryStrategy:  fault_tolerance.RecoveryStrategyBackoff,
		MaxRetries:        3,
		BackoffMultiplier: 2.0,
		InitialBackoff:    500 * time.Millisecond,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	node := createTestNodes(1)[0]
	err = manager.AddNode(node)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Track recovery attempts
	var recoveryAttempts []time.Time
	manager.SetRecoveryCallback(func(nodeID string, attempt int) {
		if nodeID == node.ID {
			recoveryAttempts = append(recoveryAttempts, time.Now())
		}
	})

	// Simulate failure and failed recovery attempts
	manager.SimulateNodeFailure(node.ID)
	
	// First attempt fails immediately
	manager.SimulateFailedRecovery(node.ID)
	
	// Wait and observe backoff intervals
	time.Sleep(4 * time.Second)

	// Should have multiple attempts with increasing intervals
	assert.GreaterOrEqual(t, len(recoveryAttempts), 2)

	if len(recoveryAttempts) >= 3 {
		// Check backoff intervals
		interval1 := recoveryAttempts[1].Sub(recoveryAttempts[0])
		interval2 := recoveryAttempts[2].Sub(recoveryAttempts[1])
		
		assert.GreaterOrEqual(t, interval2, interval1*2, "Backoff should increase exponentially")
	}

	// Finally succeed recovery
	manager.SimulateNodeRecovery(node.ID)
	
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(node.ID)
		return status == fault_tolerance.NodeStatusHealthy
	}, 2*time.Second, 100*time.Millisecond)
}

// testHealthBasedRecovery tests health-based recovery strategy
func testHealthBasedRecovery(t *testing.T) {
	config := &fault_tolerance.Config{
		RecoveryStrategy:     fault_tolerance.RecoveryStrategyHealthBased,
		HealthCheckInterval:  200 * time.Millisecond,
		HealthCheckThreshold: 3, // Need 3 consecutive healthy checks
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)
	defer manager.Stop()

	node := createTestNodes(1)[0]
	err = manager.AddNode(node)
	require.NoError(t, err)

	err = manager.Start()
	require.NoError(t, err)

	// Simulate failure
	manager.SimulateNodeFailure(node.ID)
	time.Sleep(1 * time.Second)

	// Start recovery with intermittent health
	manager.SimulateHealthyCheck(node.ID, false) // Unhealthy
	time.Sleep(250 * time.Millisecond)
	
	manager.SimulateHealthyCheck(node.ID, true) // Healthy
	time.Sleep(250 * time.Millisecond)
	
	manager.SimulateHealthyCheck(node.ID, false) // Unhealthy (resets counter)
	time.Sleep(250 * time.Millisecond)
	
	// Should not be recovered yet
	assert.Equal(t, fault_tolerance.NodeStatusFailed, manager.GetNodeStatus(node.ID))

	// Now provide consistent healthy checks
	for i := 0; i < 3; i++ {
		manager.SimulateHealthyCheck(node.ID, true)
		time.Sleep(250 * time.Millisecond)
	}

	// Should be recovered now
	assert.Eventually(t, func() bool {
		status := manager.GetNodeStatus(node.ID)
		return status == fault_tolerance.NodeStatusHealthy
	}, 2*time.Second, 100*time.Millisecond, "Node should recover after consistent healthy checks")
}

// Helper functions

// createTestManager creates a test fault tolerance manager
func createTestManager(t *testing.T) *fault_tolerance.Manager {
	config := &fault_tolerance.Config{
		HeartbeatInterval:   500 * time.Millisecond,
		FailureTimeout:      2 * time.Second,
		RecoveryTimeout:     5 * time.Second,
		EnableAutoFailover:  true,
		EnableAutoRecovery:  true,
		HealthCheckTimeout:  1 * time.Second,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)

	return manager
}

// createTestManagerWithQuorum creates a test manager with quorum configuration
func createTestManagerWithQuorum(t *testing.T, nodeCount int) *fault_tolerance.Manager {
	config := &fault_tolerance.Config{
		EnableQuorum:         true,
		QuorumSize:           (nodeCount / 2) + 1,
		SplitBrainProtection: true,
		HeartbeatInterval:    300 * time.Millisecond,
		FailureTimeout:       1 * time.Second,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(t, err)

	// Add nodes
	for i := 0; i < nodeCount; i++ {
		node := &fault_tolerance.Node{
			ID:     fmt.Sprintf("quorum-node-%d", i+1),
			Status: fault_tolerance.NodeStatusHealthy,
		}
		err = manager.AddNode(node)
		require.NoError(t, err)
	}

	err = manager.Start()
	require.NoError(t, err)

	// Wait for quorum establishment
	time.Sleep(1 * time.Second)

	return manager
}

// createTestNodes creates test nodes for fault tolerance testing
func createTestNodes(count int) []*fault_tolerance.Node {
	nodes := make([]*fault_tolerance.Node, count)
	
	for i := 0; i < count; i++ {
		nodes[i] = &fault_tolerance.Node{
			ID:       fmt.Sprintf("test-node-%d", i+1),
			Status:   fault_tolerance.NodeStatusHealthy,
			Address:  fmt.Sprintf("192.168.1.%d:8080", i+1),
			Role:     fault_tolerance.RoleWorker,
			Capacity: 100,
			Load:     0,
			Metrics: &fault_tolerance.NodeMetrics{
				CPUUsage:    0.1,
				MemoryUsage: 0.2,
				DiskUsage:   0.15,
				NetworkIO:   1024,
				LastSeen:    time.Now(),
			},
		}
	}
	
	return nodes
}

// getNodeIDs extracts node IDs from node slice
func getNodeIDs(nodes []*fault_tolerance.Node) []string {
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID
	}
	return ids
}

// BenchmarkFaultTolerance benchmarks fault tolerance operations
func BenchmarkFaultTolerance(b *testing.B) {
	manager := createTestManagerForBench(b)
	defer manager.Stop()

	// Add nodes
	for i := 0; i < 10; i++ {
		node := &fault_tolerance.Node{
			ID:     fmt.Sprintf("bench-node-%d", i),
			Status: fault_tolerance.NodeStatusHealthy,
		}
		manager.AddNode(node)
	}

	manager.Start()
	time.Sleep(500 * time.Millisecond)

	b.Run("FailureDetection", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				nodeID := fmt.Sprintf("bench-node-%d", i%10)
				manager.CheckNodeHealth(nodeID)
				i++
			}
		})
	})

	b.Run("StatusQuery", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				nodeID := fmt.Sprintf("bench-node-%d", i%10)
				manager.GetNodeStatus(nodeID)
				i++
			}
		})
	})

	b.Run("LoadAssignment", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				request := &fault_tolerance.LoadRequest{
					ID:   fmt.Sprintf("bench-req-%d", i),
					Load: 10,
				}
				manager.AssignLoad(request)
				i++
			}
		})
	})
}

// createTestManagerForBench creates a test manager for benchmarking
func createTestManagerForBench(b *testing.B) *fault_tolerance.Manager {
	config := &fault_tolerance.Config{
		HeartbeatInterval:  100 * time.Millisecond,
		FailureTimeout:     1 * time.Second,
		EnableAutoFailover: true,
	}

	manager, err := fault_tolerance.NewManager(config)
	require.NoError(b, err)

	return manager
}