package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/integration"
)

// ByzantineTestSuite represents a Byzantine failure testing suite
type ByzantineTestSuite struct {
	cluster       *integration.TestCluster
	byzantineNode *integration.TestNode
	honestNodes   []*integration.TestNode
	faultCount    int
	tolerance     int
}

// NewByzantineTestSuite creates a new Byzantine failure testing suite
func NewByzantineTestSuite(nodeCount int) (*ByzantineTestSuite, error) {
	cluster, err := integration.NewTestCluster(nodeCount)
	if err != nil {
		return nil, err
	}

	// Calculate Byzantine fault tolerance
	// For n nodes, we can tolerate at most (n-1)/3 Byzantine failures
	tolerance := (nodeCount - 1) / 3
	faultCount := min(tolerance, 1) // Start with 1 Byzantine node

	return &ByzantineTestSuite{
		cluster:    cluster,
		faultCount: faultCount,
		tolerance:  tolerance,
	}, nil
}

// TestByzantineFailureTolerance tests Byzantine failure tolerance
func TestByzantineFailureTolerance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Byzantine test in short mode")
	}

	// Need at least 4 nodes for Byzantine fault tolerance (3f+1 where f=1)
	suite, err := NewByzantineTestSuite(4)
	require.NoError(t, err)
	defer suite.cluster.Shutdown()

	err = suite.cluster.Start()
	require.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(15 * time.Second)

	t.Run("SingleByzantineNode", func(t *testing.T) {
		suite.testSingleByzantineNode(t)
	})

	t.Run("ByzantineConsensus", func(t *testing.T) {
		suite.testByzantineConsensus(t)
	})

	t.Run("ByzantineInference", func(t *testing.T) {
		suite.testByzantineInference(t)
	})

	t.Run("ByzantineRecovery", func(t *testing.T) {
		suite.testByzantineRecovery(t)
	})

	t.Run("MultipleByzantineNodes", func(t *testing.T) {
		if suite.tolerance > 1 {
			suite.testMultipleByzantineNodes(t)
		} else {
			t.Skip("Not enough nodes for multiple Byzantine failures")
		}
	})
}

// testSingleByzantineNode tests behavior with a single Byzantine node
func (bts *ByzantineTestSuite) testSingleByzantineNode(t *testing.T) {
	t.Log("Testing single Byzantine node behavior")

	// Select a Byzantine node (non-leader)
	nodes := bts.cluster.GetActiveNodes()
	var byzantineNode *integration.TestNode
	var honestNodes []*integration.TestNode

	for _, node := range nodes {
		if !node.IsLeader() && byzantineNode == nil {
			byzantineNode = node
		} else {
			honestNodes = append(honestNodes, node)
		}
	}

	require.NotNil(t, byzantineNode, "Should have a non-leader node for Byzantine behavior")
	require.GreaterOrEqual(t, len(honestNodes), 3, "Should have at least 3 honest nodes")

	bts.byzantineNode = byzantineNode
	bts.honestNodes = honestNodes

	t.Logf("Byzantine node: %s", byzantineNode.GetID())
	t.Logf("Honest nodes: %d", len(honestNodes))

	// Start Byzantine behavior
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	go bts.simulateByzantineBehavior(ctx, byzantineNode, t)

	// Test that honest nodes can still reach consensus
	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Apply consensus operations
	consensusOps := 10
	successCount := 0

	for i := 0; i < consensusOps; i++ {
		key := fmt.Sprintf("byzantine_test_%d", i)
		value := fmt.Sprintf("test_value_%d", i)

		err := leader.ApplyConsensusOperation(key, value)
		if err == nil {
			successCount++
		} else {
			t.Logf("Consensus operation %d failed: %v", i, err)
		}

		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
	}

	t.Logf("Consensus operations: %d/%d successful", successCount, consensusOps)
	assert.Greater(t, successCount, consensusOps*2/3, "Should have > 2/3 successful consensus operations")

	// Verify consistency among honest nodes
	time.Sleep(10 * time.Second)
	bts.verifyConsistency(t, honestNodes)
}

// testByzantineConsensus tests consensus with Byzantine failures
func (bts *ByzantineTestSuite) testByzantineConsensus(t *testing.T) {
	t.Log("Testing Byzantine consensus behavior")

	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Concurrent consensus operations with Byzantine interference
	numOperations := 50
	concurrency := 10

	operationChan := make(chan int, numOperations)
	resultChan := make(chan bool, numOperations)
	errorChan := make(chan error, numOperations)

	// Generate operations
	for i := 0; i < numOperations; i++ {
		operationChan <- i
	}
	close(operationChan)

	// Start Byzantine interference
	byzantineCtx, byzantineCancel := context.WithCancel(context.Background())
	defer byzantineCancel()

	if bts.byzantineNode != nil {
		go bts.simulateByzantineBehavior(byzantineCtx, bts.byzantineNode, t)
	}

	// Execute concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for opNum := range operationChan {
				key := fmt.Sprintf("concurrent_byzantine_%d", opNum)
				value := fmt.Sprintf("value_%d", opNum)

				err := leader.ApplyConsensusOperation(key, value)
				if err != nil {
					errorChan <- err
				} else {
					resultChan <- true
				}
			}
		}()
	}

	wg.Wait()
	byzantineCancel()

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numOperations; i++ {
		select {
		case <-resultChan:
			successCount++
		case err := <-errorChan:
			t.Logf("Consensus error: %v", err)
			errorCount++
		case <-time.After(1 * time.Second):
			errorCount++
		}
	}

	t.Logf("Byzantine consensus test: %d successful, %d failed", successCount, errorCount)
	assert.Greater(t, successCount, numOperations/2, "Should have > 50% successful operations despite Byzantine interference")

	// Verify final consistency
	time.Sleep(15 * time.Second)
	bts.verifyConsistency(t, bts.honestNodes)
}

// testByzantineInference tests inference with Byzantine failures
func (bts *ByzantineTestSuite) testByzantineInference(t *testing.T) {
	t.Log("Testing Byzantine inference behavior")

	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Start Byzantine behavior
	byzantineCtx, byzantineCancel := context.WithCancel(context.Background())
	defer byzantineCancel()

	if bts.byzantineNode != nil {
		go bts.simulateByzantineBehavior(byzantineCtx, bts.byzantineNode, t)
	}

	// Test inference operations
	numInferences := 20
	successCount := 0

	for i := 0; i < numInferences; i++ {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: fmt.Sprintf("Byzantine inference test %d", i),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  30,
			},
		}

		_, err := leader.ProcessInference(context.Background(), req)
		if err == nil {
			successCount++
		} else {
			t.Logf("Inference %d failed: %v", i, err)
		}

		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
	}

	t.Logf("Byzantine inference test: %d/%d successful", successCount, numInferences)
	assert.Greater(t, successCount, numInferences*2/3, "Should have > 2/3 successful inferences despite Byzantine interference")
}

// testByzantineRecovery tests recovery from Byzantine failures
func (bts *ByzantineTestSuite) testByzantineRecovery(t *testing.T) {
	t.Log("Testing Byzantine failure recovery")

	if bts.byzantineNode == nil {
		t.Skip("No Byzantine node to recover")
	}

	// Stop Byzantine behavior by restarting the node
	t.Logf("Stopping Byzantine behavior on node: %s", bts.byzantineNode.GetID())
	
	err := bts.byzantineNode.Shutdown()
	require.NoError(t, err)

	time.Sleep(10 * time.Second)

	err = bts.byzantineNode.Start()
	require.NoError(t, err)

	// Wait for recovery
	time.Sleep(30 * time.Second)

	// Test that system works normally after recovery
	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Test consensus operations
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("recovery_test_%d", i)
		value := fmt.Sprintf("recovery_value_%d", i)

		err := leader.ApplyConsensusOperation(key, value)
		assert.NoError(t, err, "Consensus should work after Byzantine recovery")
	}

	// Test inference operations
	for i := 0; i < 5; i++ {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: fmt.Sprintf("Recovery test %d", i),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		_, err := leader.ProcessInference(context.Background(), req)
		assert.NoError(t, err, "Inference should work after Byzantine recovery")
	}

	// Verify consistency across all nodes
	time.Sleep(10 * time.Second)
	allNodes := bts.cluster.GetActiveNodes()
	bts.verifyConsistency(t, allNodes)
}

// testMultipleByzantineNodes tests multiple Byzantine nodes
func (bts *ByzantineTestSuite) testMultipleByzantineNodes(t *testing.T) {
	t.Log("Testing multiple Byzantine nodes")

	nodes := bts.cluster.GetActiveNodes()
	byzantineNodes := make([]*integration.TestNode, 0)
	honestNodes := make([]*integration.TestNode, 0)

	// Select Byzantine nodes (up to tolerance limit)
	byzantineCount := 0
	for _, node := range nodes {
		if !node.IsLeader() && byzantineCount < bts.tolerance {
			byzantineNodes = append(byzantineNodes, node)
			byzantineCount++
		} else {
			honestNodes = append(honestNodes, node)
		}
	}

	require.Equal(t, bts.tolerance, len(byzantineNodes), "Should have exactly %d Byzantine nodes", bts.tolerance)
	require.GreaterOrEqual(t, len(honestNodes), 2*bts.tolerance+1, "Should have at least 2f+1 honest nodes")

	t.Logf("Byzantine nodes: %d", len(byzantineNodes))
	t.Logf("Honest nodes: %d", len(honestNodes))

	// Start Byzantine behavior on all Byzantine nodes
	byzantineCtx, byzantineCancel := context.WithCancel(context.Background())
	defer byzantineCancel()

	for _, node := range byzantineNodes {
		go bts.simulateByzantineBehavior(byzantineCtx, node, t)
	}

	// Test consensus with multiple Byzantine nodes
	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	consensusOps := 20
	successCount := 0

	for i := 0; i < consensusOps; i++ {
		key := fmt.Sprintf("multi_byzantine_%d", i)
		value := fmt.Sprintf("multi_value_%d", i)

		err := leader.ApplyConsensusOperation(key, value)
		if err == nil {
			successCount++
		} else {
			t.Logf("Multi-Byzantine consensus operation %d failed: %v", i, err)
		}

		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
	}

	t.Logf("Multi-Byzantine consensus: %d/%d successful", successCount, consensusOps)
	
	// With f Byzantine nodes, we should still be able to make progress
	// as long as we have at least 2f+1 honest nodes
	expectedSuccessRate := 0.5 // At least 50% should succeed
	actualSuccessRate := float64(successCount) / float64(consensusOps)
	
	assert.Greater(t, actualSuccessRate, expectedSuccessRate, 
		"Should have > %.0f%% successful operations with %d Byzantine nodes", 
		expectedSuccessRate*100, len(byzantineNodes))

	// Verify consistency among honest nodes
	time.Sleep(15 * time.Second)
	bts.verifyConsistency(t, honestNodes)
}

// simulateByzantineBehavior simulates Byzantine behavior on a node
func (bts *ByzantineTestSuite) simulateByzantineBehavior(ctx context.Context, node *integration.TestNode, t *testing.T) {
	t.Logf("Starting Byzantine behavior on node: %s", node.GetID())

	behaviors := []func(context.Context, *integration.TestNode, *testing.T){
		bts.sendConflictingMessages,
		bts.sendRandomMessages,
		bts.delayMessages,
		bts.duplicateMessages,
	}

	ticker := time.NewTicker(time.Duration(rand.Intn(5)+2) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Logf("Stopping Byzantine behavior on node: %s", node.GetID())
			return
		case <-ticker.C:
			if !node.IsActive() {
				continue
			}

			// Randomly select a Byzantine behavior
			behavior := behaviors[rand.Intn(len(behaviors))]
			behavior(ctx, node, t)
		}
	}
}

// sendConflictingMessages sends conflicting consensus messages
func (bts *ByzantineTestSuite) sendConflictingMessages(ctx context.Context, node *integration.TestNode, t *testing.T) {
	key := fmt.Sprintf("conflict_%d", time.Now().UnixNano())
	value1 := fmt.Sprintf("conflicting_value_1_%d", rand.Intn(1000))
	value2 := fmt.Sprintf("conflicting_value_2_%d", rand.Intn(1000))

	// Send conflicting values
	go func() {
		err := node.ApplyConsensusOperation(key, value1)
		if err != nil {
			t.Logf("Byzantine conflict 1 failed: %v", err)
		}
	}()

	go func() {
		err := node.ApplyConsensusOperation(key, value2)
		if err != nil {
			t.Logf("Byzantine conflict 2 failed: %v", err)
		}
	}()
}

// sendRandomMessages sends random consensus messages
func (bts *ByzantineTestSuite) sendRandomMessages(ctx context.Context, node *integration.TestNode, t *testing.T) {
	key := fmt.Sprintf("random_%d", time.Now().UnixNano())
	value := fmt.Sprintf("random_value_%d", rand.Intn(10000))

	err := node.ApplyConsensusOperation(key, value)
	if err != nil {
		t.Logf("Byzantine random message failed: %v", err)
	}
}

// delayMessages simulates message delays
func (bts *ByzantineTestSuite) delayMessages(ctx context.Context, node *integration.TestNode, t *testing.T) {
	// Simulate delay by sleeping
	delay := time.Duration(rand.Intn(5)+1) * time.Second
	time.Sleep(delay)

	key := fmt.Sprintf("delayed_%d", time.Now().UnixNano())
	value := fmt.Sprintf("delayed_value_%d", rand.Intn(1000))

	err := node.ApplyConsensusOperation(key, value)
	if err != nil {
		t.Logf("Byzantine delayed message failed: %v", err)
	}
}

// duplicateMessages sends duplicate messages
func (bts *ByzantineTestSuite) duplicateMessages(ctx context.Context, node *integration.TestNode, t *testing.T) {
	key := fmt.Sprintf("duplicate_%d", time.Now().UnixNano())
	value := fmt.Sprintf("duplicate_value_%d", rand.Intn(1000))

	// Send the same message multiple times
	for i := 0; i < 3; i++ {
		go func(index int) {
			err := node.ApplyConsensusOperation(key, value)
			if err != nil {
				t.Logf("Byzantine duplicate %d failed: %v", index, err)
			}
		}(i)
	}
}

// verifyConsistency verifies data consistency across nodes
func (bts *ByzantineTestSuite) verifyConsistency(t *testing.T, nodes []*integration.TestNode) {
	t.Log("Verifying consistency across nodes")

	// Apply a test operation
	leader := bts.cluster.GetLeader()
	require.NotNil(t, leader)

	testKey := fmt.Sprintf("consistency_check_%d", time.Now().UnixNano())
	testValue := fmt.Sprintf("consistency_value_%d", rand.Intn(1000))

	err := leader.ApplyConsensusOperation(testKey, testValue)
	require.NoError(t, err)

	// Wait for replication
	time.Sleep(10 * time.Second)

	// Check consistency across specified nodes
	consistentNodes := 0
	for _, node := range nodes {
		if !node.IsActive() {
			continue
		}

		value, exists := node.GetConsensusValue(testKey)
		if exists && value == testValue {
			consistentNodes++
		} else {
			t.Logf("Node %s is inconsistent: exists=%v, value=%s (expected=%s)", 
				node.GetID(), exists, value, testValue)
		}
	}

	t.Logf("Consistent nodes: %d/%d", consistentNodes, len(nodes))
	
	// Require that majority of nodes are consistent
	majorityThreshold := len(nodes)/2 + 1
	assert.GreaterOrEqual(t, consistentNodes, majorityThreshold, 
		"Should have at least %d consistent nodes", majorityThreshold)
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestByzantineEdgeCases tests edge cases in Byzantine failure scenarios
func TestByzantineEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Byzantine edge case tests in short mode")
	}

	t.Run("ByzantineLeaderFailure", func(t *testing.T) {
		testByzantineLeaderFailure(t)
	})

	t.Run("ByzantineNodeJoin", func(t *testing.T) {
		testByzantineNodeJoin(t)
	})

	t.Run("ByzantineNetworkPartition", func(t *testing.T) {
		testByzantineNetworkPartition(t)
	})
}

// testByzantineLeaderFailure tests Byzantine behavior when leader fails
func testByzantineLeaderFailure(t *testing.T) {
	suite, err := NewByzantineTestSuite(5)
	require.NoError(t, err)
	defer suite.cluster.Shutdown()

	err = suite.cluster.Start()
	require.NoError(t, err)

	time.Sleep(15 * time.Second)

	// Get initial leader
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader)

	// Start Byzantine behavior on a non-leader node
	nodes := suite.cluster.GetActiveNodes()
	var byzantineNode *integration.TestNode
	for _, node := range nodes {
		if !node.IsLeader() {
			byzantineNode = node
			break
		}
	}

	require.NotNil(t, byzantineNode)

	byzantineCtx, byzantineCancel := context.WithCancel(context.Background())
	defer byzantineCancel()

	go suite.simulateByzantineBehavior(byzantineCtx, byzantineNode, t)

	// Fail the leader
	t.Logf("Failing leader: %s", leader.GetID())
	err = leader.Shutdown()
	require.NoError(t, err)

	// Wait for leader election
	time.Sleep(30 * time.Second)

	// Verify new leader is elected
	newLeader := suite.cluster.GetLeader()
	assert.NotNil(t, newLeader)
	assert.NotEqual(t, leader.GetID(), newLeader.GetID())

	// Test that system works with new leader despite Byzantine node
	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "Byzantine leader failure test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err = newLeader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Should work with new leader despite Byzantine node")
}

// testByzantineNodeJoin tests Byzantine behavior when new node joins
func testByzantineNodeJoin(t *testing.T) {
	// Start with smaller cluster
	suite, err := NewByzantineTestSuite(3)
	require.NoError(t, err)
	defer suite.cluster.Shutdown()

	err = suite.cluster.Start()
	require.NoError(t, err)

	time.Sleep(15 * time.Second)

	// TODO: Implement node joining functionality
	// This would require extending the test infrastructure to support
	// dynamic node addition to the cluster
	t.Skip("Dynamic node joining not yet implemented in test infrastructure")
}

// testByzantineNetworkPartition tests Byzantine behavior with network partitions
func testByzantineNetworkPartition(t *testing.T) {
	suite, err := NewByzantineTestSuite(5)
	require.NoError(t, err)
	defer suite.cluster.Shutdown()

	err = suite.cluster.Start()
	require.NoError(t, err)

	time.Sleep(15 * time.Second)

	nodes := suite.cluster.GetActiveNodes()
	require.GreaterOrEqual(t, len(nodes), 5)

	// Create partition: 3 nodes on one side, 2 on the other
	// One of the minority nodes will be Byzantine
	majorityNodes := nodes[:3]
	minorityNodes := nodes[3:]

	// Make one minority node Byzantine
	byzantineNode := minorityNodes[0]
	byzantineCtx, byzantineCancel := context.WithCancel(context.Background())
	defer byzantineCancel()

	go suite.simulateByzantineBehavior(byzantineCtx, byzantineNode, t)

	// Simulate partition by shutting down minority nodes
	for _, node := range minorityNodes {
		err := node.Shutdown()
		require.NoError(t, err)
	}

	// Wait for partition effects
	time.Sleep(30 * time.Second)

	// Test that majority partition works
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader)

	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "Byzantine partition test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err = leader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Majority partition should work")

	// Heal partition
	for _, node := range minorityNodes {
		err := node.Start()
		require.NoError(t, err)
	}

	// Wait for healing
	time.Sleep(30 * time.Second)

	// Test that system works after healing
	_, err = leader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Should work after partition healing")
}