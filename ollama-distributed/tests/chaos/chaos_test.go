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

	"github.com/ollama/ollama-distributed/pkg/api"
	"github.com/ollama/ollama-distributed/tests/integration"
)

// ChaosTestSuite represents a chaos testing suite
type ChaosTestSuite struct {
	cluster     *integration.TestCluster
	running     bool
	stopCh      chan struct{}
	mu          sync.RWMutex
	events      []ChaosEvent
	scenarios   []ChaosScenario
}

// ChaosEvent represents a chaos event
type ChaosEvent struct {
	Type      string    `json:"type"`
	Target    string    `json:"target"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// ChaosScenario represents a chaos testing scenario
type ChaosScenario struct {
	Name        string
	Description string
	Duration    time.Duration
	Execute     func(*ChaosTestSuite, *testing.T) error
	Validate    func(*ChaosTestSuite, *testing.T) error
}

// NewChaosTestSuite creates a new chaos testing suite
func NewChaosTestSuite(nodeCount int) (*ChaosTestSuite, error) {
	cluster, err := integration.NewTestCluster(nodeCount)
	if err != nil {
		return nil, err
	}

	suite := &ChaosTestSuite{
		cluster:   cluster,
		stopCh:    make(chan struct{}),
		events:    make([]ChaosEvent, 0),
		scenarios: make([]ChaosScenario, 0),
	}

	// Register chaos scenarios
	suite.registerScenarios()

	return suite, nil
}

// registerScenarios registers all chaos scenarios
func (cts *ChaosTestSuite) registerScenarios() {
	cts.scenarios = []ChaosScenario{
		{
			Name:        "RandomNodeFailure",
			Description: "Randomly fail nodes during operation",
			Duration:    5 * time.Minute,
			Execute:     cts.randomNodeFailureScenario,
			Validate:    cts.validateSystemRecovery,
		},
		{
			Name:        "NetworkPartition",
			Description: "Simulate network partitions",
			Duration:    3 * time.Minute,
			Execute:     cts.networkPartitionScenario,
			Validate:    cts.validateConsistency,
		},
		{
			Name:        "ResourceExhaustion",
			Description: "Simulate resource exhaustion",
			Duration:    4 * time.Minute,
			Execute:     cts.resourceExhaustionScenario,
			Validate:    cts.validateResourceRecovery,
		},
		{
			Name:        "ByzantineFailure",
			Description: "Simulate Byzantine failures",
			Duration:    6 * time.Minute,
			Execute:     cts.byzantineFailureScenario,
			Validate:    cts.validateByzantineRecovery,
		},
		{
			Name:        "CascadingFailure",
			Description: "Simulate cascading failures",
			Duration:    4 * time.Minute,
			Execute:     cts.cascadingFailureScenario,
			Validate:    cts.validateCascadingRecovery,
		},
		{
			Name:        "HighLoadStorm",
			Description: "Simulate sudden high load",
			Duration:    3 * time.Minute,
			Execute:     cts.highLoadStormScenario,
			Validate:    cts.validateLoadRecovery,
		},
	}
}

// Start starts the chaos testing suite
func (cts *ChaosTestSuite) Start() error {
	cts.mu.Lock()
	defer cts.mu.Unlock()

	if cts.running {
		return fmt.Errorf("chaos suite already running")
	}

	err := cts.cluster.Start()
	if err != nil {
		return err
	}

	// Wait for cluster to stabilize
	time.Sleep(15 * time.Second)

	cts.running = true
	return nil
}

// Stop stops the chaos testing suite
func (cts *ChaosTestSuite) Stop() {
	cts.mu.Lock()
	defer cts.mu.Unlock()

	if !cts.running {
		return
	}

	close(cts.stopCh)
	cts.cluster.Shutdown()
	cts.running = false
}

// logEvent logs a chaos event
func (cts *ChaosTestSuite) logEvent(eventType, target, action, details string) {
	cts.mu.Lock()
	defer cts.mu.Unlock()

	event := ChaosEvent{
		Type:      eventType,
		Target:    target,
		Action:    action,
		Timestamp: time.Now(),
		Details:   details,
	}

	cts.events = append(cts.events, event)
}

// TestChaosEngineering runs the complete chaos engineering test suite
func TestChaosEngineering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	suite, err := NewChaosTestSuite(5)
	require.NoError(t, err)
	defer suite.Stop()

	err = suite.Start()
	require.NoError(t, err)

	// Run each chaos scenario
	for _, scenario := range suite.scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Logf("Starting chaos scenario: %s", scenario.Description)
			
			// Execute the chaos scenario
			err := scenario.Execute(suite, t)
			require.NoError(t, err)
			
			// Validate the results
			err = scenario.Validate(suite, t)
			require.NoError(t, err)
			
			t.Logf("Completed chaos scenario: %s", scenario.Name)
		})
	}

	// Generate chaos report
	suite.generateReport(t)
}

// randomNodeFailureScenario simulates random node failures
func (cts *ChaosTestSuite) randomNodeFailureScenario(t *testing.T) error {
	t.Log("Starting random node failure scenario")

	// Start background load
	loadCtx, loadCancel := context.WithCancel(context.Background())
	defer loadCancel()

	go cts.generateBackgroundLoad(loadCtx, t)

	// Random node failures
	duration := 3 * time.Minute
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		nodes := cts.cluster.GetActiveNodes()
		if len(nodes) <= 2 {
			// Don't fail more nodes if we have too few
			time.Sleep(10 * time.Second)
			continue
		}

		// Pick a random non-leader node
		var targetNode *integration.TestNode
		for _, node := range nodes {
			if !node.IsLeader() {
				targetNode = node
				break
			}
		}

		if targetNode != nil {
			t.Logf("Failing node: %s", targetNode.GetID())
			cts.logEvent("node_failure", targetNode.GetID(), "shutdown", "Random failure")
			
			err := targetNode.Shutdown()
			if err != nil {
				t.Logf("Failed to shutdown node %s: %v", targetNode.GetID(), err)
			}

			// Wait for failure detection
			time.Sleep(15 * time.Second)

			// Restart the node after some time
			if rand.Float64() < 0.7 { // 70% chance to restart
				t.Logf("Restarting node: %s", targetNode.GetID())
				cts.logEvent("node_recovery", targetNode.GetID(), "restart", "Recovery after failure")
				
				err := targetNode.Start()
				if err != nil {
					t.Logf("Failed to restart node %s: %v", targetNode.GetID(), err)
				}
			}
		}

		// Wait before next failure
		time.Sleep(time.Duration(rand.Intn(30)+15) * time.Second)
	}

	t.Log("Completed random node failure scenario")
	return nil
}

// networkPartitionScenario simulates network partitions
func (cts *ChaosTestSuite) networkPartitionScenario(t *testing.T) error {
	t.Log("Starting network partition scenario")

	// Start background load
	loadCtx, loadCancel := context.WithCancel(context.Background())
	defer loadCancel()

	go cts.generateBackgroundLoad(loadCtx, t)

	// Simulate network partitions
	nodes := cts.cluster.GetActiveNodes()
	if len(nodes) < 3 {
		return fmt.Errorf("need at least 3 nodes for network partition test")
	}

	// Create partition: isolate one node
	partitionNode := nodes[len(nodes)-1]
	t.Logf("Creating network partition: isolating node %s", partitionNode.GetID())
	cts.logEvent("network_partition", partitionNode.GetID(), "isolate", "Network partition created")

	// Simulate partition by shutting down the node temporarily
	err := partitionNode.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to create partition: %v", err)
	}

	// Wait for partition duration
	time.Sleep(90 * time.Second)

	// Heal the partition
	t.Logf("Healing network partition: reconnecting node %s", partitionNode.GetID())
	cts.logEvent("network_partition", partitionNode.GetID(), "heal", "Network partition healed")

	err = partitionNode.Start()
	if err != nil {
		return fmt.Errorf("failed to heal partition: %v", err)
	}

	// Wait for recovery
	time.Sleep(60 * time.Second)

	t.Log("Completed network partition scenario")
	return nil
}

// resourceExhaustionScenario simulates resource exhaustion
func (cts *ChaosTestSuite) resourceExhaustionScenario(t *testing.T) error {
	t.Log("Starting resource exhaustion scenario")

	leader := cts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Fill up the cluster with many models
	modelCount := 50
	for i := 0; i < modelCount; i++ {
		modelInfo := &integration.ModelInfo{
			Name:              fmt.Sprintf("resource-exhaustion-model-%d", i),
			Size:              200 * 1024 * 1024, // 200MB each
			Checksum:          fmt.Sprintf("checksum-%d", i),
			ReplicationFactor: 2,
			LastAccessed:      time.Now(),
			Popularity:        rand.Float64(),
		}

		err := leader.RegisterModel(modelInfo)
		if err != nil {
			t.Logf("Failed to register model %d: %v", i, err)
		}
	}

	cts.logEvent("resource_exhaustion", "cluster", "models", fmt.Sprintf("Registered %d models", modelCount))

	// Generate high load
	highLoadCtx, highLoadCancel := context.WithCancel(context.Background())
	defer highLoadCancel()

	go cts.generateHighLoad(highLoadCtx, t, 100, 20)

	// Wait for resource exhaustion effects
	time.Sleep(2 * time.Minute)

	t.Log("Completed resource exhaustion scenario")
	return nil
}

// byzantineFailureScenario simulates Byzantine failures
func (cts *ChaosTestSuite) byzantineFailureScenario(t *testing.T) error {
	t.Log("Starting Byzantine failure scenario")

	// Start background load
	loadCtx, loadCancel := context.WithCancel(context.Background())
	defer loadCancel()

	go cts.generateBackgroundLoad(loadCtx, t)

	// Simulate Byzantine behavior by having a node send conflicting messages
	nodes := cts.cluster.GetActiveNodes()
	if len(nodes) < 4 {
		return fmt.Errorf("need at least 4 nodes for Byzantine failure test")
	}

	// Pick a non-leader node for Byzantine behavior
	var byzantineNode *integration.TestNode
	for _, node := range nodes {
		if !node.IsLeader() {
			byzantineNode = node
			break
		}
	}

	if byzantineNode == nil {
		return fmt.Errorf("no suitable node for Byzantine failure")
	}

	t.Logf("Simulating Byzantine failure on node: %s", byzantineNode.GetID())
	cts.logEvent("byzantine_failure", byzantineNode.GetID(), "start", "Byzantine behavior initiated")

	// Simulate Byzantine behavior by sending conflicting consensus operations
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("byzantine_key_%d", i)
		value1 := fmt.Sprintf("value1_%d", i)
		value2 := fmt.Sprintf("value2_%d", i)

		// Send conflicting values
		go func() {
			err := byzantineNode.ApplyConsensusOperation(key, value1)
			if err != nil {
				t.Logf("Byzantine operation 1 failed: %v", err)
			}
		}()

		go func() {
			err := byzantineNode.ApplyConsensusOperation(key, value2)
			if err != nil {
				t.Logf("Byzantine operation 2 failed: %v", err)
			}
		}()

		time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
	}

	// Wait for Byzantine behavior duration
	time.Sleep(2 * time.Minute)

	// Stop Byzantine behavior
	t.Logf("Stopping Byzantine behavior on node: %s", byzantineNode.GetID())
	cts.logEvent("byzantine_failure", byzantineNode.GetID(), "stop", "Byzantine behavior stopped")

	// Restart the node to return to normal behavior
	err := byzantineNode.Shutdown()
	if err != nil {
		t.Logf("Failed to shutdown Byzantine node: %v", err)
	}

	time.Sleep(10 * time.Second)

	err = byzantineNode.Start()
	if err != nil {
		t.Logf("Failed to restart Byzantine node: %v", err)
	}

	// Wait for recovery
	time.Sleep(60 * time.Second)

	t.Log("Completed Byzantine failure scenario")
	return nil
}

// cascadingFailureScenario simulates cascading failures
func (cts *ChaosTestSuite) cascadingFailureScenario(t *testing.T) error {
	t.Log("Starting cascading failure scenario")

	// Start background load
	loadCtx, loadCancel := context.WithCancel(context.Background())
	defer loadCancel()

	go cts.generateBackgroundLoad(loadCtx, t)

	nodes := cts.cluster.GetActiveNodes()
	if len(nodes) < 4 {
		return fmt.Errorf("need at least 4 nodes for cascading failure test")
	}

	// Start cascading failures
	failureDelay := 20 * time.Second
	failedNodes := make([]*integration.TestNode, 0)

	for i := 0; i < 2; i++ { // Fail 2 nodes in cascade
		// Find a non-leader node
		var targetNode *integration.TestNode
		for _, node := range nodes {
			if !node.IsLeader() {
				found := false
				for _, failed := range failedNodes {
					if failed.GetID() == node.GetID() {
						found = true
						break
					}
				}
				if !found {
					targetNode = node
					break
				}
			}
		}

		if targetNode == nil {
			break
		}

		t.Logf("Cascading failure %d: failing node %s", i+1, targetNode.GetID())
		cts.logEvent("cascading_failure", targetNode.GetID(), "fail", fmt.Sprintf("Cascading failure %d", i+1))

		err := targetNode.Shutdown()
		if err != nil {
			t.Logf("Failed to shutdown node %s: %v", targetNode.GetID(), err)
		}

		failedNodes = append(failedNodes, targetNode)

		// Wait for failure effects to propagate
		time.Sleep(failureDelay)
	}

	// Wait for system to handle cascading failures
	time.Sleep(90 * time.Second)

	// Recovery phase: restart nodes one by one
	for i, node := range failedNodes {
		t.Logf("Recovering from cascading failure %d: restarting node %s", i+1, node.GetID())
		cts.logEvent("cascading_recovery", node.GetID(), "restart", fmt.Sprintf("Recovery %d", i+1))

		err := node.Start()
		if err != nil {
			t.Logf("Failed to restart node %s: %v", node.GetID(), err)
		}

		// Wait between recoveries
		time.Sleep(30 * time.Second)
	}

	t.Log("Completed cascading failure scenario")
	return nil
}

// highLoadStormScenario simulates sudden high load
func (cts *ChaosTestSuite) highLoadStormScenario(t *testing.T) error {
	t.Log("Starting high load storm scenario")

	// Start with normal load
	normalLoadCtx, normalLoadCancel := context.WithCancel(context.Background())
	go cts.generateBackgroundLoad(normalLoadCtx, t)

	// Wait for baseline
	time.Sleep(30 * time.Second)

	// Cancel normal load
	normalLoadCancel()

	// Start high load storm
	t.Log("Initiating high load storm")
	cts.logEvent("load_storm", "cluster", "start", "High load storm initiated")

	stormCtx, stormCancel := context.WithCancel(context.Background())
	defer stormCancel()

	go cts.generateHighLoad(stormCtx, t, 200, 50)

	// Maintain high load for duration
	time.Sleep(90 * time.Second)

	// Stop high load
	t.Log("Stopping high load storm")
	cts.logEvent("load_storm", "cluster", "stop", "High load storm stopped")
	stormCancel()

	// Wait for recovery
	time.Sleep(60 * time.Second)

	t.Log("Completed high load storm scenario")
	return nil
}

// generateBackgroundLoad generates background load for testing
func (cts *ChaosTestSuite) generateBackgroundLoad(ctx context.Context, t *testing.T) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	requestCount := 0
	for {
		select {
		case <-ctx.Done():
			t.Logf("Background load generator stopped after %d requests", requestCount)
			return
		case <-ticker.C:
			leader := cts.cluster.GetLeader()
			if leader == nil {
				continue
			}

			req := &api.InferenceRequest{
				Model:  "llama3.2:1b",
				Prompt: fmt.Sprintf("Background load request %d", requestCount),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  20,
				},
			}

			_, err := leader.ProcessInference(context.Background(), req)
			if err != nil {
				t.Logf("Background load request failed: %v", err)
			}

			requestCount++
		}
	}
}

// generateHighLoad generates high load for testing
func (cts *ChaosTestSuite) generateHighLoad(ctx context.Context, t *testing.T, requests, concurrency int) {
	requestChan := make(chan int, requests)
	responseChan := make(chan bool, requests)
	errorChan := make(chan error, requests)

	// Generate requests
	for i := 0; i < requests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Process requests concurrently
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case reqNum, ok := <-requestChan:
					if !ok {
						return
					}

					leader := cts.cluster.GetLeader()
					if leader == nil {
						errorChan <- fmt.Errorf("no leader available")
						continue
					}

					req := &api.InferenceRequest{
						Model:  "llama3.2:1b",
						Prompt: fmt.Sprintf("High load request %d", reqNum),
						Options: map[string]interface{}{
							"temperature": 0.1,
							"max_tokens":  30,
						},
					}

					_, err := leader.ProcessInference(context.Background(), req)
					if err != nil {
						errorChan <- err
					} else {
						responseChan <- true
					}
				}
			}
		}()
	}

	wg.Wait()

	// Count results
	successCount := 0
	errorCount := 0

	for len(responseChan) > 0 || len(errorChan) > 0 {
		select {
		case <-responseChan:
			successCount++
		case <-errorChan:
			errorCount++
		default:
			break
		}
	}

	t.Logf("High load test: %d successful, %d failed", successCount, errorCount)
}

// Validation functions

// validateSystemRecovery validates system recovery after chaos
func (cts *ChaosTestSuite) validateSystemRecovery(t *testing.T) error {
	t.Log("Validating system recovery")

	// Check cluster health
	activeNodes := cts.cluster.GetActiveNodes()
	assert.GreaterOrEqual(t, len(activeNodes), 1, "Should have at least 1 active node")

	// Check leader election
	leader := cts.cluster.GetLeader()
	assert.NotNil(t, leader, "Should have a leader")

	// Test basic functionality
	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "System recovery validation test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err := leader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Basic inference should work after recovery")

	return nil
}

// validateConsistency validates data consistency
func (cts *ChaosTestSuite) validateConsistency(t *testing.T) error {
	t.Log("Validating data consistency")

	// Test consensus consistency
	leader := cts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Apply a consensus operation
	testKey := "consistency_test"
	testValue := "consistency_value"

	err := leader.ApplyConsensusOperation(testKey, testValue)
	require.NoError(t, err)

	// Wait for replication
	time.Sleep(10 * time.Second)

	// Check consistency across nodes
	activeNodes := cts.cluster.GetActiveNodes()
	for _, node := range activeNodes {
		value, exists := node.GetConsensusValue(testKey)
		assert.True(t, exists, "Node %s should have the key", node.GetID())
		assert.Equal(t, testValue, value, "Node %s should have correct value", node.GetID())
	}

	return nil
}

// validateResourceRecovery validates resource recovery
func (cts *ChaosTestSuite) validateResourceRecovery(t *testing.T) error {
	t.Log("Validating resource recovery")

	// Test that system can handle requests after resource exhaustion
	leader := cts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Send multiple requests to test resource recovery
	successCount := 0
	totalRequests := 20

	for i := 0; i < totalRequests; i++ {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: fmt.Sprintf("Resource recovery test %d", i),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		_, err := leader.ProcessInference(context.Background(), req)
		if err == nil {
			successCount++
		}
	}

	successRate := float64(successCount) / float64(totalRequests)
	assert.Greater(t, successRate, 0.5, "Should have > 50% success rate after resource recovery")

	return nil
}

// validateByzantineRecovery validates recovery from Byzantine failures
func (cts *ChaosTestSuite) validateByzantineRecovery(t *testing.T) error {
	t.Log("Validating Byzantine failure recovery")

	// Test that consensus still works correctly
	leader := cts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Apply multiple consensus operations
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("byzantine_recovery_%d", i)
		value := fmt.Sprintf("recovery_value_%d", i)

		err := leader.ApplyConsensusOperation(key, value)
		assert.NoError(t, err, "Consensus should work after Byzantine recovery")
	}

	// Test inference functionality
	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "Byzantine recovery test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err := leader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Inference should work after Byzantine recovery")

	return nil
}

// validateCascadingRecovery validates recovery from cascading failures
func (cts *ChaosTestSuite) validateCascadingRecovery(t *testing.T) error {
	t.Log("Validating cascading failure recovery")

	// Wait for full recovery
	time.Sleep(30 * time.Second)

	// Check that majority of nodes are active
	activeNodes := cts.cluster.GetActiveNodes()
	totalNodes := len(cts.cluster.GetNodes())
	
	assert.GreaterOrEqual(t, len(activeNodes), totalNodes/2, "Should have majority of nodes active")

	// Test system functionality
	leader := cts.cluster.GetLeader()
	assert.NotNil(t, leader, "Should have a leader after cascading recovery")

	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "Cascading recovery test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err := leader.ProcessInference(context.Background(), req)
	assert.NoError(t, err, "Inference should work after cascading recovery")

	return nil
}

// validateLoadRecovery validates recovery from high load
func (cts *ChaosTestSuite) validateLoadRecovery(t *testing.T) error {
	t.Log("Validating load recovery")

	// Test that system responds normally after load storm
	leader := cts.cluster.GetLeader()
	require.NotNil(t, leader)

	// Measure response time
	start := time.Now()
	req := &api.InferenceRequest{
		Model:  "llama3.2:1b",
		Prompt: "Load recovery test",
		Options: map[string]interface{}{
			"temperature": 0.1,
			"max_tokens":  20,
		},
	}

	_, err := leader.ProcessInference(context.Background(), req)
	responseTime := time.Since(start)

	assert.NoError(t, err, "Inference should work after load recovery")
	assert.Less(t, responseTime, 10*time.Second, "Response time should be reasonable after load recovery")

	return nil
}

// generateReport generates a chaos testing report
func (cts *ChaosTestSuite) generateReport(t *testing.T) {
	t.Log("=== CHAOS TESTING REPORT ===")
	t.Logf("Total chaos events: %d", len(cts.events))

	// Group events by type
	eventTypes := make(map[string]int)
	for _, event := range cts.events {
		eventTypes[event.Type]++
	}

	t.Log("Event breakdown:")
	for eventType, count := range eventTypes {
		t.Logf("  %s: %d", eventType, count)
	}

	// Final cluster state
	activeNodes := cts.cluster.GetActiveNodes()
	totalNodes := len(cts.cluster.GetNodes())
	
	t.Logf("Final cluster state: %d/%d nodes active", len(activeNodes), totalNodes)
	
	leader := cts.cluster.GetLeader()
	if leader != nil {
		t.Logf("Current leader: %s", leader.GetID())
	} else {
		t.Log("No leader elected")
	}

	// Test final system health
	if leader != nil {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: "Final health check",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  10,
			},
		}

		_, err := leader.ProcessInference(context.Background(), req)
		if err != nil {
			t.Logf("Final health check FAILED: %v", err)
		} else {
			t.Log("Final health check PASSED")
		}
	}

	t.Log("=== END CHAOS TESTING REPORT ===")
}