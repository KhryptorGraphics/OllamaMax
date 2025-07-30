package consensus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

// TestConsensusEngine tests the consensus engine lifecycle
func TestConsensusEngine(t *testing.T) {
	t.Run("EngineCreation", testEngineCreation)
	t.Run("EngineStartup", testEngineStartup)
	t.Run("EngineShutdown", testEngineShutdown)
	t.Run("SingleNodeConsensus", testSingleNodeConsensus)
}

// testEngineCreation tests consensus engine creation
func testEngineCreation(t *testing.T) {
	config := &consensus.Config{
		NodeID:          "test-node-1",
		DataDir:         "/tmp/consensus-test-1",
		BindAddr:        "127.0.0.1:0",
		BootstrapExpect: 1,
		HeartbeatTimeout: 1000 * time.Millisecond,
		ElectionTimeout:  5000 * time.Millisecond,
		CommitTimeout:    500 * time.Millisecond,
		MaxAppendEntries: 64,
		SnapshotInterval: 1000,
		SnapshotThreshold: 8192,
	}

	engine, err := consensus.NewEngine(config, nil)
	assert.NoError(t, err)
	assert.NotNil(t, engine)
	assert.Equal(t, config.NodeID, engine.GetNodeID())
	assert.False(t, engine.IsStarted())
	assert.False(t, engine.IsLeader())

	// Clean up
	engine.Close()
}

// testEngineStartup tests consensus engine startup
func testEngineStartup(t *testing.T) {
	engine := createTestEngine(t, "startup-test", 1)
	defer engine.Close()

	// Start engine
	err := engine.Start()
	assert.NoError(t, err)
	assert.True(t, engine.IsStarted())

	// Wait for leadership in single node cluster
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 10*time.Second, 100*time.Millisecond, "Node should become leader")

	// Test idempotent startup
	err = engine.Start()
	assert.NoError(t, err)
	assert.True(t, engine.IsStarted())
}

// testEngineShutdown tests graceful engine shutdown
func testEngineShutdown(t *testing.T) {
	engine := createTestEngine(t, "shutdown-test", 1)

	// Start and then shutdown
	err := engine.Start()
	require.NoError(t, err)

	// Wait for startup
	time.Sleep(500 * time.Millisecond)

	// Shutdown
	err = engine.Close()
	assert.NoError(t, err)
	assert.False(t, engine.IsStarted())

	// Multiple shutdowns should be safe
	err = engine.Close()
	assert.NoError(t, err)
}

// testSingleNodeConsensus tests consensus operations with single node
func testSingleNodeConsensus(t *testing.T) {
	engine := createTestEngine(t, "single-node", 1)
	defer engine.Close()

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 10*time.Second, 100*time.Millisecond)

	// Test state operations
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": []string{"a", "b", "c"},
	}

	for key, value := range testData {
		err = engine.Set(key, value)
		assert.NoError(t, err)

		retrievedValue, exists := engine.Get(key)
		assert.True(t, exists)
		assert.Equal(t, value, retrievedValue)
	}

	// Test deletion
	err = engine.Delete("key1")
	assert.NoError(t, err)

	_, exists := engine.Get("key1")
	assert.False(t, exists)
}

// TestMultiNodeConsensus tests consensus with multiple nodes
func TestMultiNodeConsensus(t *testing.T) {
	t.Run("ThreeNodeCluster", testThreeNodeCluster)
	t.Run("FiveNodeCluster", testFiveNodeCluster)
	t.Run("ConsistentState", testConsistentState)
	t.Run("LeaderElection", testLeaderElection)
}

// testThreeNodeCluster tests 3-node consensus cluster
func testThreeNodeCluster(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to form and elect leader
	cluster.WaitForLeader(10 * time.Second)

	leader := cluster.GetLeader()
	assert.NotNil(t, leader)

	followers := cluster.GetFollowers()
	assert.Equal(t, 2, len(followers))

	// Test basic operations through leader
	err = leader.Set("test-key", "test-value")
	assert.NoError(t, err)

	// Verify consistency across all nodes
	cluster.WaitForConsistency(5 * time.Second)

	for i, engine := range cluster.engines {
		value, exists := engine.Get("test-key")
		assert.True(t, exists, "Node %d should have the key", i)
		assert.Equal(t, "test-value", value, "Node %d should have correct value", i)
	}
}

// testFiveNodeCluster tests 5-node consensus cluster
func testFiveNodeCluster(t *testing.T) {
	cluster := createTestCluster(t, 5)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	// Wait for cluster formation
	cluster.WaitForLeader(15 * time.Second)

	leader := cluster.GetLeader()
	assert.NotNil(t, leader)

	followers := cluster.GetFollowers()
	assert.Equal(t, 4, len(followers))

	// Test concurrent operations
	var wg sync.WaitGroup
	numOperations := 50

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-key-%d", index)
			value := fmt.Sprintf("concurrent-value-%d", index)
			
			err := leader.Set(key, value)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Wait for all operations to propagate
	cluster.WaitForConsistency(10 * time.Second)

	// Verify all operations succeeded
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("concurrent-key-%d", i)
		expectedValue := fmt.Sprintf("concurrent-value-%d", i)

		for j, engine := range cluster.engines {
			value, exists := engine.Get(key)
			assert.True(t, exists, "Node %d should have key %s", j, key)
			assert.Equal(t, expectedValue, value, "Node %d should have correct value for %s", j, key)
		}
	}
}

// testConsistentState tests state consistency across nodes
func testConsistentState(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()

	// Perform multiple operations
	operations := []struct {
		op    string
		key   string
		value interface{}
	}{
		{"set", "user:1", map[string]interface{}{"name": "Alice", "age": 30}},
		{"set", "user:2", map[string]interface{}{"name": "Bob", "age": 25}},
		{"set", "config:timeout", 5000},
		{"set", "config:enabled", true},
		{"delete", "user:1", nil},
		{"set", "user:3", map[string]interface{}{"name": "Charlie", "age": 35}},
	}

	for _, op := range operations {
		switch op.op {
		case "set":
			err = leader.Set(op.key, op.value)
			assert.NoError(t, err)
		case "delete":
			err = leader.Delete(op.key)
			assert.NoError(t, err)
		}
	}

	// Wait for consistency
	cluster.WaitForConsistency(5 * time.Second)

	// Verify final state on all nodes
	expectedState := map[string]interface{}{
		"user:2":         map[string]interface{}{"name": "Bob", "age": 25},
		"config:timeout": 5000,
		"config:enabled": true,
		"user:3":         map[string]interface{}{"name": "Charlie", "age": 35},
	}

	for i, engine := range cluster.engines {
		for key, expectedValue := range expectedState {
			value, exists := engine.Get(key)
			assert.True(t, exists, "Node %d should have key %s", i, key)
			assert.Equal(t, expectedValue, value, "Node %d has incorrect value for %s", i, key)
		}

		// Verify deleted key is not present
		_, exists := engine.Get("user:1")
		assert.False(t, exists, "Node %d should not have deleted key user:1", i)
	}
}

// testLeaderElection tests leader election process
func testLeaderElection(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	// Wait for initial leader election
	cluster.WaitForLeader(10 * time.Second)
	
	initialLeader := cluster.GetLeader()
	assert.NotNil(t, initialLeader)
	initialLeaderID := initialLeader.GetNodeID()

	// Verify only one leader
	leaders := cluster.GetLeaders()
	assert.Equal(t, 1, len(leaders))

	// Simulate leader failure by stopping the leader
	err = initialLeader.Close()
	assert.NoError(t, err)

	// Wait for new leader election
	cluster.WaitForLeaderElection(10 * time.Second)

	newLeader := cluster.GetLeader()
	assert.NotNil(t, newLeader)
	assert.NotEqual(t, initialLeaderID, newLeader.GetNodeID())

	// Verify cluster is still functional
	err = newLeader.Set("post-election-key", "post-election-value")
	assert.NoError(t, err)

	// Wait for consistency among remaining nodes
	remainingNodes := cluster.GetRunningEngines()
	assert.Equal(t, 2, len(remainingNodes))

	time.Sleep(2 * time.Second) // Allow propagation

	for i, engine := range remainingNodes {
		value, exists := engine.Get("post-election-key")
		assert.True(t, exists, "Remaining node %d should have the key", i)
		assert.Equal(t, "post-election-value", value, "Remaining node %d should have correct value", i)
	}
}

// TestConsensusFailures tests failure scenarios
func TestConsensusFailures(t *testing.T) {
	t.Run("MinorityFailure", testMinorityFailure)
	t.Run("MajorityFailure", testMajorityFailure)
	t.Run("NetworkPartition", testNetworkPartition)
	t.Run("LeaderFailureRecovery", testLeaderFailureRecovery)
	t.Run("FollowerFailureRecovery", testFollowerFailureRecovery)
}

// testMinorityFailure tests cluster behavior when minority of nodes fail
func testMinorityFailure(t *testing.T) {
	cluster := createTestCluster(t, 5)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()

	// Set initial state
	err = leader.Set("failure-test", "initial-value")
	assert.NoError(t, err)

	cluster.WaitForConsistency(2 * time.Second)

	// Stop 2 nodes (minority in 5-node cluster)
	followers := cluster.GetFollowers()
	assert.GreaterOrEqual(t, len(followers), 2)

	for i := 0; i < 2; i++ {
		err = followers[i].Close()
		assert.NoError(t, err)
	}

	// Cluster should still be functional
	err = leader.Set("after-failure", "still-working")
	assert.NoError(t, err)

	// Wait for consensus among remaining nodes
	time.Sleep(2 * time.Second)

	runningEngines := cluster.GetRunningEngines()
	assert.Equal(t, 3, len(runningEngines))

	for i, engine := range runningEngines {
		value, exists := engine.Get("after-failure")
		assert.True(t, exists, "Running node %d should have the key", i)
		assert.Equal(t, "still-working", value, "Running node %d should have correct value", i)
	}
}

// testMajorityFailure tests cluster behavior when majority of nodes fail
func testMajorityFailure(t *testing.T) {
	cluster := createTestCluster(t, 5)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()

	// Set initial state
	err = leader.Set("majority-failure-test", "initial-value")
	assert.NoError(t, err)

	cluster.WaitForConsistency(2 * time.Second)

	// Stop 3 nodes (majority in 5-node cluster)
	allEngines := cluster.engines
	var remainingEngines []*consensus.Engine

	for i := 0; i < 3; i++ {
		err = allEngines[i].Close()
		assert.NoError(t, err)
	}

	// Keep track of remaining engines
	for i := 3; i < 5; i++ {
		remainingEngines = append(remainingEngines, allEngines[i])
	}

	// Cluster should not be able to achieve consensus
	// Remaining nodes should not be able to elect a leader
	time.Sleep(5 * time.Second)

	hasLeader := false
	for _, engine := range remainingEngines {
		if engine.IsLeader() {
			hasLeader = true
			break
		}
	}

	assert.False(t, hasLeader, "No node should be leader with majority failure")

	// Operations should fail or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, engine := range remainingEngines {
		err = engine.SetWithContext(ctx, "should-fail", "value")
		assert.Error(t, err, "Operations should fail without majority")
	}
}

// testNetworkPartition tests network partition scenarios
func testNetworkPartition(t *testing.T) {
	cluster := createTestCluster(t, 5)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)

	// Create network partition: 3 nodes vs 2 nodes
	partition1 := cluster.engines[0:3] // Majority partition
	partition2 := cluster.engines[3:5] // Minority partition

	// Simulate network partition by isolating partitions
	cluster.SimulatePartition(partition1, partition2)

	// Wait for partition effects
	time.Sleep(5 * time.Second)

	// Majority partition should elect leader and be functional
	var majorityLeader *consensus.Engine
	for _, engine := range partition1 {
		if engine.IsLeader() {
			majorityLeader = engine
			break
		}
	}
	assert.NotNil(t, majorityLeader, "Majority partition should have a leader")

	// Minority partition should not have a leader
	minorityHasLeader := false
	for _, engine := range partition2 {
		if engine.IsLeader() {
			minorityHasLeader = true
			break
		}
	}
	assert.False(t, minorityHasLeader, "Minority partition should not have a leader")

	// Majority partition should accept writes
	err = majorityLeader.Set("partition-test", "majority-value")
	assert.NoError(t, err)

	// Minority partition should reject writes
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, engine := range partition2 {
		err = engine.SetWithContext(ctx, "should-fail", "value")
		assert.Error(t, err, "Minority partition should reject writes")
	}

	// Heal partition
	cluster.HealPartition()

	// Wait for cluster to reunify
	time.Sleep(5 * time.Second)

	// All nodes should converge to the same state
	cluster.WaitForConsistency(5 * time.Second)

	for i, engine := range cluster.engines {
		if !engine.IsRunning() {
			continue // Skip stopped engines
		}

		value, exists := engine.Get("partition-test")
		assert.True(t, exists, "Node %d should have partition-test key", i)
		assert.Equal(t, "majority-value", value, "Node %d should have majority value", i)
	}
}

// testLeaderFailureRecovery tests recovery from leader failure
func testLeaderFailureRecovery(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	originalLeader := cluster.GetLeader()
	originalLeaderID := originalLeader.GetNodeID()

	// Set some state
	err = originalLeader.Set("pre-failure", "value")
	assert.NoError(t, err)

	cluster.WaitForConsistency(2 * time.Second)

	// Stop the leader
	err = originalLeader.Close()
	assert.NoError(t, err)

	// Wait for new leader election
	cluster.WaitForLeaderElection(10 * time.Second)

	newLeader := cluster.GetLeader()
	assert.NotNil(t, newLeader)
	assert.NotEqual(t, originalLeaderID, newLeader.GetNodeID())

	// Verify state is preserved
	value, exists := newLeader.Get("pre-failure")
	assert.True(t, exists)
	assert.Equal(t, "value", value)

	// New leader should be able to process operations
	err = newLeader.Set("post-failure", "recovery-value")
	assert.NoError(t, err)

	// Verify consistency among remaining nodes
	remainingEngines := cluster.GetRunningEngines()
	for _, engine := range remainingEngines {
		value, exists := engine.Get("post-failure")
		assert.True(t, exists)
		assert.Equal(t, "recovery-value", value)
	}
}

// testFollowerFailureRecovery tests recovery from follower failure
func testFollowerFailureRecovery(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()
	followers := cluster.GetFollowers()
	assert.Equal(t, 2, len(followers))

	// Stop one follower
	failedFollower := followers[0]
	err = failedFollower.Close()
	assert.NoError(t, err)

	// Cluster should remain functional
	err = leader.Set("during-failure", "still-working")
	assert.NoError(t, err)

	// Restart the failed follower
	err = failedFollower.Start()
	assert.NoError(t, err)

	// Wait for rejoining
	time.Sleep(3 * time.Second)

	// Failed follower should catch up
	cluster.WaitForConsistency(5 * time.Second)

	value, exists := failedFollower.Get("during-failure")
	assert.True(t, exists)
	assert.Equal(t, "still-working", value)
}

// TestConsensusPerformance tests performance characteristics
func TestConsensusPerformance(t *testing.T) {
	t.Run("HighThroughput", testHighThroughput)
	t.Run("LargeCluster", testLargeCluster)
	t.Run("ConcurrentOperations", testConcurrentOperations)
}

// testHighThroughput tests high throughput operations
func testHighThroughput(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()

	// Measure throughput
	numOperations := 1000
	start := time.Now()

	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("throughput-key-%d", i)
		value := fmt.Sprintf("throughput-value-%d", i)
		
		err = leader.Set(key, value)
		assert.NoError(t, err)
	}

	duration := time.Since(start)
	throughput := float64(numOperations) / duration.Seconds()

	t.Logf("Throughput: %.2f operations/second for %d operations", throughput, numOperations)
	
	// Verify minimum acceptable throughput (this may vary by system)
	assert.Greater(t, throughput, 100.0, "Throughput should be at least 100 ops/sec")

	// Verify final consistency
	cluster.WaitForConsistency(10 * time.Second)

	// Spot check some values
	checkIndices := []int{0, numOperations/2, numOperations - 1}
	for _, i := range checkIndices {
		key := fmt.Sprintf("throughput-key-%d", i)
		expectedValue := fmt.Sprintf("throughput-value-%d", i)

		for j, engine := range cluster.engines {
			value, exists := engine.Get(key)
			assert.True(t, exists, "Node %d should have key %s", j, key)
			assert.Equal(t, expectedValue, value, "Node %d should have correct value for %s", j, key)
		}
	}
}

// testLargeCluster tests consensus with larger cluster
func testLargeCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large cluster test in short mode")
	}

	clusterSize := 7
	cluster := createTestCluster(t, clusterSize)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	// Larger clusters may take longer to elect leader
	cluster.WaitForLeader(20 * time.Second)

	leader := cluster.GetLeader()
	assert.NotNil(t, leader)

	followers := cluster.GetFollowers()
	assert.Equal(t, clusterSize-1, len(followers))

	// Test basic operation
	err = leader.Set("large-cluster-test", "working")
	assert.NoError(t, err)

	// Wait for consensus across all nodes
	cluster.WaitForConsistency(10 * time.Second)

	for i, engine := range cluster.engines {
		value, exists := engine.Get("large-cluster-test")
		assert.True(t, exists, "Node %d should have the key", i)
		assert.Equal(t, "working", value, "Node %d should have correct value", i)
	}

	// Test failure resilience - stop minority of nodes
	failureCount := (clusterSize - 1) / 2 // Less than majority
	for i := 0; i < failureCount; i++ {
		err = followers[i].Close()
		assert.NoError(t, err)
	}

	// Cluster should still work
	err = leader.Set("after-partial-failure", "still-working")
	assert.NoError(t, err)

	// Wait for consensus among remaining nodes
	time.Sleep(3 * time.Second)

	runningEngines := cluster.GetRunningEngines()
	assert.Equal(t, clusterSize-failureCount, len(runningEngines))

	for _, engine := range runningEngines {
		value, exists := engine.Get("after-partial-failure")
		assert.True(t, exists)
		assert.Equal(t, "still-working", value)
	}
}

// testConcurrentOperations tests concurrent operations on consensus
func testConcurrentOperations(t *testing.T) {
	cluster := createTestCluster(t, 3)
	defer cluster.Shutdown()

	err := cluster.Start()
	require.NoError(t, err)

	cluster.WaitForLeader(10 * time.Second)
	leader := cluster.GetLeader()

	// Test concurrent operations
	numGoroutines := 10
	operationsPerGoroutine := 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]string)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				key := fmt.Sprintf("concurrent-%d-%d", goroutineID, j)
				value := fmt.Sprintf("value-%d-%d", goroutineID, j)

				err := leader.Set(key, value)
				assert.NoError(t, err)

				mu.Lock()
				results[key] = value
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Wait for all operations to propagate
	cluster.WaitForConsistency(10 * time.Second)

	// Verify all operations succeeded and are consistent
	totalOperations := numGoroutines * operationsPerGoroutine
	assert.Equal(t, totalOperations, len(results))

	for key, expectedValue := range results {
		for i, engine := range cluster.engines {
			value, exists := engine.Get(key)
			assert.True(t, exists, "Node %d should have key %s", i, key)
			assert.Equal(t, expectedValue, value, "Node %d should have correct value for %s", i, key)
		}
	}
}

// TestConsensusSnapshots tests snapshot functionality
func TestConsensusSnapshots(t *testing.T) {
	t.Run("SnapshotCreation", testSnapshotCreation)
	t.Run("SnapshotRestore", testSnapshotRestore)
	t.Run("SnapshotCompaction", testSnapshotCompaction)
}

// testSnapshotCreation tests snapshot creation
func testSnapshotCreation(t *testing.T) {
	engine := createTestEngine(t, "snapshot-creation", 1)
	defer engine.Close()

	err := engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 10*time.Second, 100*time.Millisecond)

	// Add enough data to trigger snapshot
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("snapshot-key-%d", i)
		value := fmt.Sprintf("snapshot-value-%d", i)
		err = engine.Set(key, value)
		assert.NoError(t, err)
	}

	// Force snapshot creation
	err = engine.CreateSnapshot()
	assert.NoError(t, err)

	// Verify snapshot exists
	snapshots, err := engine.ListSnapshots()
	assert.NoError(t, err)
	assert.NotEmpty(t, snapshots)

	// Verify data integrity after snapshot
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("snapshot-key-%d", i)
		expectedValue := fmt.Sprintf("snapshot-value-%d", i)
		
		value, exists := engine.Get(key)
		assert.True(t, exists)
		assert.Equal(t, expectedValue, value)
	}
}

// testSnapshotRestore tests snapshot restoration
func testSnapshotRestore(t *testing.T) {
	engine1 := createTestEngine(t, "snapshot-restore-1", 1)
	defer engine1.Close()

	err := engine1.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine1.IsLeader()
	}, 10*time.Second, 100*time.Millisecond)

	// Add test data
	testData := map[string]string{
		"restore-key-1": "restore-value-1",
		"restore-key-2": "restore-value-2",
		"restore-key-3": "restore-value-3",
	}

	for key, value := range testData {
		err = engine1.Set(key, value)
		assert.NoError(t, err)
	}

	// Create snapshot
	err = engine1.CreateSnapshot()
	assert.NoError(t, err)

	snapshots, err := engine1.ListSnapshots()
	assert.NoError(t, err)
	assert.NotEmpty(t, snapshots)

	latestSnapshot := snapshots[len(snapshots)-1]

	// Create new engine and restore from snapshot
	engine2 := createTestEngine(t, "snapshot-restore-2", 1)
	defer engine2.Close()

	err = engine2.RestoreSnapshot(latestSnapshot.ID)
	assert.NoError(t, err)

	err = engine2.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine2.IsLeader()
	}, 10*time.Second, 100*time.Millisecond)

	// Verify restored data
	for key, expectedValue := range testData {
		value, exists := engine2.Get(key)
		assert.True(t, exists, "Restored engine should have key %s", key)
		assert.Equal(t, expectedValue, value, "Restored engine should have correct value for %s", key)
	}
}

// testSnapshotCompaction tests log compaction through snapshots
func testSnapshotCompaction(t *testing.T) {
	config := &consensus.Config{
		NodeID:            "compaction-test",
		DataDir:           "/tmp/consensus-compaction-test",
		BindAddr:          "127.0.0.1:0",
		BootstrapExpect:   1,
		SnapshotInterval:  10,  // Take snapshot every 10 operations
		SnapshotThreshold: 100, // Keep max 100 log entries
	}

	engine, err := consensus.NewEngine(config, nil)
	require.NoError(t, err)
	defer engine.Close()

	err = engine.Start()
	require.NoError(t, err)

	// Wait for leadership
	assert.Eventually(t, func() bool {
		return engine.IsLeader()
	}, 10*time.Second, 100*time.Millisecond)

	// Add many operations to trigger compaction
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("compaction-key-%d", i)
		value := fmt.Sprintf("compaction-value-%d", i)
		err = engine.Set(key, value)
		assert.NoError(t, err)
	}

	// Wait for compaction
	time.Sleep(2 * time.Second)

	// Verify snapshots were created
	snapshots, err := engine.ListSnapshots()
	assert.NoError(t, err)
	assert.NotEmpty(t, snapshots)

	// Verify log was compacted
	logSize := engine.GetLogSize()
	assert.Less(t, logSize, int64(200), "Log should be compacted")

	// Verify data integrity after compaction
	for i := 0; i < 200; i++ {
		key := fmt.Sprintf("compaction-key-%d", i)
		expectedValue := fmt.Sprintf("compaction-value-%d", i)
		
		value, exists := engine.Get(key)
		assert.True(t, exists, "Key %s should exist after compaction", key)
		assert.Equal(t, expectedValue, value, "Value should be correct after compaction")
	}
}

// Helper types and functions

// TestCluster represents a test consensus cluster
type TestCluster struct {
	engines []*consensus.Engine
	configs []*consensus.Config
	size    int
}

// createTestEngine creates a single test consensus engine
func createTestEngine(t *testing.T, name string, bootstrapExpected int) *consensus.Engine {
	config := &consensus.Config{
		NodeID:           fmt.Sprintf("%s-node", name),
		DataDir:          fmt.Sprintf("/tmp/consensus-test-%s", name),
		BindAddr:         "127.0.0.1:0",
		BootstrapExpect:  bootstrapExpected,
		HeartbeatTimeout: 500 * time.Millisecond,
		ElectionTimeout:  2000 * time.Millisecond,
		CommitTimeout:    100 * time.Millisecond,
		MaxAppendEntries: 64,
		SnapshotInterval: 1000,
		SnapshotThreshold: 8192,
	}

	engine, err := consensus.NewEngine(config, nil)
	require.NoError(t, err)

	return engine
}

// createTestCluster creates a test consensus cluster
func createTestCluster(t *testing.T, size int) *TestCluster {
	cluster := &TestCluster{
		engines: make([]*consensus.Engine, size),
		configs: make([]*consensus.Config, size),
		size:    size,
	}

	// Create all engines
	for i := 0; i < size; i++ {
		config := &consensus.Config{
			NodeID:           fmt.Sprintf("test-node-%d", i),
			DataDir:          fmt.Sprintf("/tmp/consensus-cluster-test-%d", i),
			BindAddr:         "127.0.0.1:0",
			BootstrapExpect:  size,
			HeartbeatTimeout: 500 * time.Millisecond,
			ElectionTimeout:  2000 * time.Millisecond,
			CommitTimeout:    100 * time.Millisecond,
			MaxAppendEntries: 64,
			SnapshotInterval: 1000,
			SnapshotThreshold: 8192,
		}

		engine, err := consensus.NewEngine(config, nil)
		require.NoError(t, err)

		cluster.engines[i] = engine
		cluster.configs[i] = config
	}

	return cluster
}

// Start starts all engines in the cluster
func (c *TestCluster) Start() error {
	// Get all bind addresses first
	var peers []string
	for i, engine := range c.engines {
		err := engine.Start()
		if err != nil {
			return err
		}
		addr := engine.GetBindAddr()
		c.configs[i].Peers = append(c.configs[i].Peers, addr)
		peers = append(peers, addr)
	}

	// Configure peers for all engines
	for _, engine := range c.engines {
		err := engine.SetPeers(peers)
		if err != nil {
			return err
		}
	}

	return nil
}

// Shutdown shuts down all engines in the cluster
func (c *TestCluster) Shutdown() {
	for _, engine := range c.engines {
		if engine != nil {
			engine.Close()
		}
	}
}

// WaitForLeader waits for a leader to be elected
func (c *TestCluster) WaitForLeader(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.GetLeader() != nil {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// WaitForLeaderElection waits for a new leader election after current leader fails
func (c *TestCluster) WaitForLeaderElection(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		leader := c.GetLeader()
		if leader != nil && leader.IsRunning() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// GetLeader returns the current cluster leader
func (c *TestCluster) GetLeader() *consensus.Engine {
	for _, engine := range c.engines {
		if engine.IsRunning() && engine.IsLeader() {
			return engine
		}
	}
	return nil
}

// GetLeaders returns all nodes that think they are leader (should be 0 or 1)
func (c *TestCluster) GetLeaders() []*consensus.Engine {
	var leaders []*consensus.Engine
	for _, engine := range c.engines {
		if engine.IsRunning() && engine.IsLeader() {
			leaders = append(leaders, engine)
		}
	}
	return leaders
}

// GetFollowers returns all follower nodes
func (c *TestCluster) GetFollowers() []*consensus.Engine {
	var followers []*consensus.Engine
	for _, engine := range c.engines {
		if engine.IsRunning() && !engine.IsLeader() {
			followers = append(followers, engine)
		}
	}
	return followers
}

// GetRunningEngines returns all running engines
func (c *TestCluster) GetRunningEngines() []*consensus.Engine {
	var running []*consensus.Engine
	for _, engine := range c.engines {
		if engine.IsRunning() {
			running = append(running, engine)
		}
	}
	return running
}

// WaitForConsistency waits for all running nodes to have consistent state
func (c *TestCluster) WaitForConsistency(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.checkConsistency() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// checkConsistency checks if all running nodes have the same state
func (c *TestCluster) checkConsistency() bool {
	runningEngines := c.GetRunningEngines()
	if len(runningEngines) <= 1 {
		return true
	}

	// Get state from first running engine
	firstEngine := runningEngines[0]
	firstState := firstEngine.GetAllKeys()

	// Compare with all other engines
	for _, engine := range runningEngines[1:] {
		engineState := engine.GetAllKeys()
		if !compareStates(firstState, engineState) {
			return false
		}
	}

	return true
}

// compareStates compares two state maps
func compareStates(state1, state2 map[string]interface{}) bool {
	if len(state1) != len(state2) {
		return false
	}

	for key, value1 := range state1 {
		value2, exists := state2[key]
		if !exists {
			return false
		}
		if fmt.Sprintf("%v", value1) != fmt.Sprintf("%v", value2) {
			return false
		}
	}

	return true
}

// SimulatePartition simulates network partition between two groups
func (c *TestCluster) SimulatePartition(partition1, partition2 []*consensus.Engine) {
	// Implementation would depend on the specific consensus library
	// This is a placeholder for partition simulation
	for _, engine1 := range partition1 {
		for _, engine2 := range partition2 {
			engine1.DisconnectPeer(engine2.GetNodeID())
			engine2.DisconnectPeer(engine1.GetNodeID())
		}
	}
}

// HealPartition heals a network partition
func (c *TestCluster) HealPartition() {
	// Implementation would depend on the specific consensus library
	// This is a placeholder for partition healing
	for i, engine1 := range c.engines {
		for j, engine2 := range c.engines {
			if i != j && engine1.IsRunning() && engine2.IsRunning() {
				engine1.ReconnectPeer(engine2.GetNodeID())
			}
		}
	}
}

// BenchmarkConsensus benchmarks consensus operations
func BenchmarkConsensus(b *testing.B) {
	engine := createTestEngine(b, "benchmark", 1)
	defer engine.Close()

	err := engine.Start()
	require.NoError(b, err)

	// Wait for leadership
	for !engine.IsLeader() {
		time.Sleep(10 * time.Millisecond)
	}

	b.Run("SetOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("bench-key-%d", i)
				value := fmt.Sprintf("bench-value-%d", i)
				
				err := engine.Set(key, value)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	b.Run("GetOperations", func(b *testing.B) {
		// Pre-populate some data
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("get-bench-key-%d", i)
			value := fmt.Sprintf("get-bench-value-%d", i)
			engine.Set(key, value)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("get-bench-key-%d", i%1000)
				_, _ = engine.Get(key)
				i++
			}
		})
	})

	b.Run("MixedOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					key := fmt.Sprintf("mixed-key-%d", i)
					value := fmt.Sprintf("mixed-value-%d", i)
					engine.Set(key, value)
				} else {
					key := fmt.Sprintf("mixed-key-%d", i-1)
					engine.Get(key)
				}
				i++
			}
		})
	})
}

// Helper function for benchmark that accepts testing.B
func createTestEngine(t testing.TB, name string, bootstrapExpected int) *consensus.Engine {
	config := &consensus.Config{
		NodeID:           fmt.Sprintf("%s-node", name),
		DataDir:          fmt.Sprintf("/tmp/consensus-test-%s", name),
		BindAddr:         "127.0.0.1:0",
		BootstrapExpect:  bootstrapExpected,
		HeartbeatTimeout: 500 * time.Millisecond,
		ElectionTimeout:  2000 * time.Millisecond,
		CommitTimeout:    100 * time.Millisecond,
		MaxAppendEntries: 64,
		SnapshotInterval: 1000,
		SnapshotThreshold: 8192,
	}

	engine, err := consensus.NewEngine(config, nil)
	require.NoError(t, err)

	return engine
}