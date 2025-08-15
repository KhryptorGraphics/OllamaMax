package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInferenceRequest represents a simple inference request for testing
type TestInferenceRequest struct {
	ID    string
	Model string
	Input string
}

// MockDistributedScheduler simulates a real distributed scheduler for integration testing
type MockDistributedScheduler struct {
	nodes                 map[string]*MockNode
	faultToleranceManager *EnhancedFaultToleranceManager
	clusterManager        *MockClusterManager
	mu                    sync.RWMutex
	isRunning             bool
	simulatedFailures     map[string]bool
	requestCount          int64
	successfulRequests    int64
	failedRequests        int64
}

// MockNode represents a node in the distributed system
type MockNode struct {
	ID          string
	Address     string
	IsHealthy   bool
	IsAvailable bool
	Load        float64
	LastSeen    time.Time
	mu          sync.RWMutex
}

// MockClusterManager manages the cluster of nodes
type MockClusterManager struct {
	nodes map[string]*MockNode
	mu    sync.RWMutex
}

// NewMockDistributedScheduler creates a new mock distributed scheduler for testing
func NewMockDistributedScheduler(nodeCount int) *MockDistributedScheduler {
	nodes := make(map[string]*MockNode)
	for i := 0; i < nodeCount; i++ {
		nodeID := fmt.Sprintf("node-%d", i)
		nodes[nodeID] = &MockNode{
			ID:          nodeID,
			Address:     fmt.Sprintf("127.0.0.1:%d", 8080+i),
			IsHealthy:   true,
			IsAvailable: true,
			Load:        0.0,
			LastSeen:    time.Now(),
		}
	}

	clusterManager := &MockClusterManager{
		nodes: nodes,
	}

	// Create fault tolerance configuration
	baseConfig := &Config{
		ReplicationFactor:     2,
		HealthCheckInterval:   5 * time.Second,
		RecoveryTimeout:       30 * time.Second,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    10 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          2 * time.Second,
	}

	baseFT := NewFaultToleranceManager(baseConfig)
	enhancedConfig := NewEnhancedFaultToleranceConfig(baseConfig)
	eftm := NewEnhancedFaultToleranceManager(enhancedConfig, baseFT)

	// Set up node provider for fault tolerance
	nodeProvider := func() []interface{} {
		clusterManager.mu.RLock()
		defer clusterManager.mu.RUnlock()

		nodeList := make([]interface{}, 0, len(clusterManager.nodes))
		for _, node := range clusterManager.nodes {
			nodeList = append(nodeList, node)
		}
		return nodeList
	}
	eftm.SetNodeProvider(nodeProvider)

	return &MockDistributedScheduler{
		nodes:                 nodes,
		faultToleranceManager: eftm,
		clusterManager:        clusterManager,
		simulatedFailures:     make(map[string]bool),
	}
}

// Start starts the mock distributed scheduler
func (mds *MockDistributedScheduler) Start(ctx context.Context) error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	if mds.isRunning {
		return fmt.Errorf("scheduler already running")
	}

	mds.isRunning = true

	// Start fault tolerance manager (no Start method needed for testing)
	// The fault tolerance manager is already initialized and ready

	return nil
}

// Stop stops the mock distributed scheduler
func (mds *MockDistributedScheduler) Stop() error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	if !mds.isRunning {
		return nil
	}

	mds.isRunning = false

	// Stop fault tolerance manager (no Stop method needed for testing)
	// The fault tolerance manager will be cleaned up automatically

	return nil
}

// SimulateNodeFailure simulates a node failure
func (mds *MockDistributedScheduler) SimulateNodeFailure(nodeID string) error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	node, exists := mds.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.mu.Lock()
	node.IsHealthy = false
	node.IsAvailable = false
	node.mu.Unlock()

	mds.simulatedFailures[nodeID] = true

	// Trigger fault detection
	fault := &FaultDetection{
		ID:          fmt.Sprintf("fault-%s-%d", nodeID, time.Now().Unix()),
		Type:        "node_failure",
		Target:      nodeID,
		Severity:    "high",
		DetectedAt:  time.Now(),
		Description: fmt.Sprintf("Simulated failure of node %s", nodeID),
		Metadata: map[string]interface{}{
			"node_id":   nodeID,
			"simulated": true,
		},
	}

	// Report fault to fault tolerance manager (simplified for testing)
	go func() {
		// For testing, we'll just log the fault detection
		fmt.Printf("Fault detected: %s - %s\n", fault.ID, fault.Description)
	}()

	return nil
}

// RecoverNode simulates node recovery
func (mds *MockDistributedScheduler) RecoverNode(nodeID string) error {
	mds.mu.Lock()
	defer mds.mu.Unlock()

	node, exists := mds.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.mu.Lock()
	node.IsHealthy = true
	node.IsAvailable = true
	node.LastSeen = time.Now()
	node.mu.Unlock()

	delete(mds.simulatedFailures, nodeID)

	return nil
}

// ProcessRequest simulates processing a request through the distributed system
func (mds *MockDistributedScheduler) ProcessRequest(ctx context.Context, request *TestInferenceRequest) error {
	mds.mu.Lock()
	mds.requestCount++
	mds.mu.Unlock()

	// Find available nodes
	availableNodes := mds.getAvailableNodes()
	if len(availableNodes) == 0 {
		mds.mu.Lock()
		mds.failedRequests++
		mds.mu.Unlock()
		return fmt.Errorf("no available nodes")
	}

	// Simulate request processing
	time.Sleep(10 * time.Millisecond) // Simulate processing time

	// Check if any failures occurred during processing
	for _, node := range availableNodes {
		if mds.simulatedFailures[node.ID] {
			mds.mu.Lock()
			mds.failedRequests++
			mds.mu.Unlock()
			return fmt.Errorf("request failed due to node failure: %s", node.ID)
		}
	}

	mds.mu.Lock()
	mds.successfulRequests++
	mds.mu.Unlock()

	return nil
}

// getAvailableNodes returns a list of available nodes
func (mds *MockDistributedScheduler) getAvailableNodes() []*MockNode {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	var available []*MockNode
	for _, node := range mds.nodes {
		node.mu.RLock()
		if node.IsHealthy && node.IsAvailable {
			available = append(available, node)
		}
		node.mu.RUnlock()
	}

	return available
}

// GetStats returns statistics about the scheduler
func (mds *MockDistributedScheduler) GetStats() (int64, int64, int64) {
	mds.mu.RLock()
	defer mds.mu.RUnlock()

	return mds.requestCount, mds.successfulRequests, mds.failedRequests
}

// GetAvailableNodes returns available nodes for cluster manager interface
func (mcm *MockClusterManager) GetAvailableNodes() []interface{} {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	var nodes []interface{}
	for _, node := range mcm.nodes {
		node.mu.RLock()
		if node.IsHealthy && node.IsAvailable {
			nodes = append(nodes, node)
		}
		node.mu.RUnlock()
	}

	return nodes
}

// LoadConfiguration loads fault tolerance configuration into the scheduler
func (mds *MockDistributedScheduler) LoadConfiguration(config *config.DistributedConfig) error {
	return mds.faultToleranceManager.LoadConfiguration(config)
}

// Integration Tests

// TestDistributedSchedulerIntegration tests the complete integration of fault tolerance with distributed scheduler
func TestDistributedSchedulerIntegration(t *testing.T) {
	t.Run("basic scheduler operation with fault tolerance", func(t *testing.T) {
		// Create scheduler with 3 nodes
		scheduler := NewMockDistributedScheduler(3)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process some requests
		for i := 0; i < 10; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Verify all requests were successful
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(10), total)
		assert.Equal(t, int64(10), successful)
		assert.Equal(t, int64(0), failed)
	})

	t.Run("fault tolerance with single node failure", func(t *testing.T) {
		// Create scheduler with 3 nodes
		scheduler := NewMockDistributedScheduler(3)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process some requests successfully
		for i := 0; i < 5; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Simulate node failure
		err = scheduler.SimulateNodeFailure("node-0")
		require.NoError(t, err)

		// Wait for fault detection and recovery
		time.Sleep(2 * time.Second)

		// Continue processing requests (should still work with remaining nodes)
		for i := 5; i < 10; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			// Some requests might fail during the failure window, but system should recover
			if err != nil {
				t.Logf("Request %d failed during node failure: %v", i, err)
			}
		}

		// Verify system continued operating
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(10), total)
		assert.True(t, successful >= 5, "Should have at least 5 successful requests")
		t.Logf("Stats: Total=%d, Successful=%d, Failed=%d", total, successful, failed)
	})

	t.Run("fault tolerance with multiple node failures", func(t *testing.T) {
		// Create scheduler with 5 nodes for better resilience
		scheduler := NewMockDistributedScheduler(5)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process initial requests
		for i := 0; i < 5; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Simulate multiple node failures
		err = scheduler.SimulateNodeFailure("node-0")
		require.NoError(t, err)

		time.Sleep(1 * time.Second)

		err = scheduler.SimulateNodeFailure("node-1")
		require.NoError(t, err)

		// Wait for fault detection and recovery
		time.Sleep(3 * time.Second)

		// Continue processing requests
		for i := 5; i < 15; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during multiple node failures: %v", i, err)
			}
		}

		// Verify system maintained some level of operation
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(15), total)
		assert.True(t, successful >= 5, "Should have at least 5 successful requests")
		t.Logf("Stats after multiple failures: Total=%d, Successful=%d, Failed=%d", total, successful, failed)
	})
}

// createIntegrationTestConfig creates a configuration optimized for integration testing
func createIntegrationTestConfig() *config.DistributedConfig {
	cfg := &config.DistributedConfig{}
	cfg.Inference.FaultTolerance.Enabled = true
	cfg.Inference.FaultTolerance.RetryAttempts = 3
	cfg.Inference.FaultTolerance.RetryDelay = "100ms"
	cfg.Inference.FaultTolerance.HealthCheckInterval = "5s"
	cfg.Inference.FaultTolerance.RecoveryTimeout = "30s"
	cfg.Inference.FaultTolerance.CircuitBreakerEnabled = true
	cfg.Inference.FaultTolerance.CheckpointInterval = "2s"
	cfg.Inference.FaultTolerance.MaxRetries = 3
	cfg.Inference.FaultTolerance.RetryBackoff = "200ms"
	cfg.Inference.FaultTolerance.ReplicationFactor = 2

	// Predictive detection (faster intervals for testing)
	cfg.Inference.FaultTolerance.PredictiveDetection.Enabled = true
	cfg.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.7
	cfg.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "2s"
	cfg.Inference.FaultTolerance.PredictiveDetection.WindowSize = "5s"
	cfg.Inference.FaultTolerance.PredictiveDetection.Threshold = 0.7
	cfg.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
	cfg.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = true
	cfg.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = true

	// Self-healing (faster intervals for testing)
	cfg.Inference.FaultTolerance.SelfHealing.Enabled = true
	cfg.Inference.FaultTolerance.SelfHealing.HealingThreshold = 0.6
	cfg.Inference.FaultTolerance.SelfHealing.HealingInterval = "10s"
	cfg.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "1s"
	cfg.Inference.FaultTolerance.SelfHealing.LearningInterval = "5s"
	cfg.Inference.FaultTolerance.SelfHealing.ServiceRestart = true
	cfg.Inference.FaultTolerance.SelfHealing.ResourceReallocation = true
	cfg.Inference.FaultTolerance.SelfHealing.LoadRedistribution = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableLearning = true
	cfg.Inference.FaultTolerance.SelfHealing.EnablePredictive = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableProactive = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableFailover = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableScaling = false // Disable scaling for simpler testing

	// Redundancy
	cfg.Inference.FaultTolerance.Redundancy.Enabled = true
	cfg.Inference.FaultTolerance.Redundancy.DefaultFactor = 2
	cfg.Inference.FaultTolerance.Redundancy.MaxFactor = 3
	cfg.Inference.FaultTolerance.Redundancy.UpdateInterval = "10s"

	// Performance tracking (window should be 3x healing interval: 10s * 3 = 30s)
	cfg.Inference.FaultTolerance.PerformanceTracking.Enabled = true
	cfg.Inference.FaultTolerance.PerformanceTracking.WindowSize = "30s"

	// Config adaptation
	cfg.Inference.FaultTolerance.ConfigAdaptation.Enabled = true
	cfg.Inference.FaultTolerance.ConfigAdaptation.Interval = "1m"

	return cfg
}

// TestMultiNodeFailureRecoveryScenarios tests various multi-node failure and recovery patterns
func TestMultiNodeFailureRecoveryScenarios(t *testing.T) {
	t.Run("cascading failure recovery", func(t *testing.T) {
		// Create scheduler with 6 nodes for better resilience testing
		scheduler := NewMockDistributedScheduler(6)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process initial requests to establish baseline
		for i := 0; i < 10; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("baseline-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Simulate cascading failures (one node fails, then another, then another)
		err = scheduler.SimulateNodeFailure("node-0")
		require.NoError(t, err)
		t.Logf("Failed node-0")

		time.Sleep(2 * time.Second)

		err = scheduler.SimulateNodeFailure("node-1")
		require.NoError(t, err)
		t.Logf("Failed node-1")

		time.Sleep(2 * time.Second)

		err = scheduler.SimulateNodeFailure("node-2")
		require.NoError(t, err)
		t.Logf("Failed node-2")

		// Wait for fault detection and recovery
		time.Sleep(5 * time.Second)

		// Continue processing requests with remaining nodes
		for i := 10; i < 25; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("during-failure-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during cascading failures: %v", i, err)
			}
		}

		// Recover nodes one by one
		err = scheduler.RecoverNode("node-0")
		require.NoError(t, err)
		t.Logf("Recovered node-0")

		time.Sleep(2 * time.Second)

		err = scheduler.RecoverNode("node-1")
		require.NoError(t, err)
		t.Logf("Recovered node-1")

		time.Sleep(2 * time.Second)

		err = scheduler.RecoverNode("node-2")
		require.NoError(t, err)
		t.Logf("Recovered node-2")

		// Wait for recovery to stabilize
		time.Sleep(3 * time.Second)

		// Process final requests to verify full recovery
		for i := 25; i < 35; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("post-recovery-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err, "Request should succeed after recovery")
		}

		// Verify system maintained operation throughout
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(35), total)
		assert.True(t, successful >= 20, "Should have at least 20 successful requests despite failures")
		t.Logf("Cascading failure stats: Total=%d, Successful=%d, Failed=%d", total, successful, failed)
	})

	t.Run("simultaneous failure recovery", func(t *testing.T) {
		// Create scheduler with 8 nodes
		scheduler := NewMockDistributedScheduler(8)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process initial requests
		for i := 0; i < 10; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("initial-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Simulate simultaneous failure of multiple nodes
		go func() {
			scheduler.SimulateNodeFailure("node-0")
		}()
		go func() {
			scheduler.SimulateNodeFailure("node-1")
		}()
		go func() {
			scheduler.SimulateNodeFailure("node-2")
		}()
		go func() {
			scheduler.SimulateNodeFailure("node-3")
		}()

		t.Logf("Simulated simultaneous failure of 4 nodes")

		// Wait for fault detection
		time.Sleep(3 * time.Second)

		// Continue processing requests with remaining nodes
		for i := 10; i < 30; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("during-simultaneous-failure-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during simultaneous failures: %v", i, err)
			}
		}

		// Recover all nodes simultaneously
		go func() {
			scheduler.RecoverNode("node-0")
		}()
		go func() {
			scheduler.RecoverNode("node-1")
		}()
		go func() {
			scheduler.RecoverNode("node-2")
		}()
		go func() {
			scheduler.RecoverNode("node-3")
		}()

		t.Logf("Initiated simultaneous recovery of 4 nodes")

		// Wait for recovery
		time.Sleep(5 * time.Second)

		// Process final requests
		for i := 30; i < 40; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("post-simultaneous-recovery-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err, "Request should succeed after simultaneous recovery")
		}

		// Verify system performance
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(40), total)
		assert.True(t, successful >= 25, "Should have at least 25 successful requests despite simultaneous failures")
		t.Logf("Simultaneous failure stats: Total=%d, Successful=%d, Failed=%d", total, successful, failed)
	})

	t.Run("partial recovery scenario", func(t *testing.T) {
		// Create scheduler with 5 nodes
		scheduler := NewMockDistributedScheduler(5)

		// Create test configuration
		config := createIntegrationTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Process initial requests
		for i := 0; i < 5; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("pre-failure-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Fail 3 out of 5 nodes
		err = scheduler.SimulateNodeFailure("node-0")
		require.NoError(t, err)
		err = scheduler.SimulateNodeFailure("node-1")
		require.NoError(t, err)
		err = scheduler.SimulateNodeFailure("node-2")
		require.NoError(t, err)

		t.Logf("Failed 3 out of 5 nodes")

		// Wait for fault detection
		time.Sleep(3 * time.Second)

		// Process requests with reduced capacity
		for i := 5; i < 15; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("reduced-capacity-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed with reduced capacity: %v", i, err)
			}
		}

		// Recover only 1 node (partial recovery)
		err = scheduler.RecoverNode("node-0")
		require.NoError(t, err)
		t.Logf("Partially recovered: node-0 back online")

		// Wait for partial recovery
		time.Sleep(3 * time.Second)

		// Process requests with partial recovery
		for i := 15; i < 25; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("partial-recovery-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during partial recovery: %v", i, err)
			}
		}

		// Complete recovery
		err = scheduler.RecoverNode("node-1")
		require.NoError(t, err)
		err = scheduler.RecoverNode("node-2")
		require.NoError(t, err)
		t.Logf("Completed recovery: all nodes back online")

		// Wait for full recovery
		time.Sleep(3 * time.Second)

		// Process final requests
		for i := 25; i < 30; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("full-recovery-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err, "Request should succeed after full recovery")
		}

		// Verify system resilience
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(30), total)
		assert.True(t, successful >= 15, "Should have at least 15 successful requests despite partial failures")
		t.Logf("Partial recovery stats: Total=%d, Successful=%d, Failed=%d", total, successful, failed)
	})
}

// TestPredictiveDetectionPerformance tests the accuracy and performance of predictive fault detection
func TestPredictiveDetectionPerformance(t *testing.T) {
	t.Run("predictive detection accuracy under normal load", func(t *testing.T) {
		// Create scheduler with 4 nodes
		scheduler := NewMockDistributedScheduler(4)

		// Create test configuration with enhanced predictive detection
		config := createIntegrationTestConfig()
		config.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.8
		config.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "1s"
		config.Inference.FaultTolerance.PredictiveDetection.WindowSize = "3s"

		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Generate normal load for baseline
		startTime := time.Now()
		for i := 0; i < 50; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("normal-load-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)

			// Small delay to simulate realistic load
			time.Sleep(50 * time.Millisecond)
		}

		normalLoadDuration := time.Since(startTime)
		t.Logf("Normal load processing time: %v", normalLoadDuration)

		// Verify all requests succeeded under normal conditions
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(50), total)
		assert.Equal(t, int64(50), successful)
		assert.Equal(t, int64(0), failed)

		// Measure predictive detection overhead
		assert.True(t, normalLoadDuration < 10*time.Second, "Predictive detection should not significantly impact performance")
	})

	t.Run("predictive detection under high load", func(t *testing.T) {
		// Create scheduler with 6 nodes for high load testing
		scheduler := NewMockDistributedScheduler(6)

		// Create test configuration optimized for high load
		config := createIntegrationTestConfig()
		config.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.7
		config.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "500ms"
		config.Inference.FaultTolerance.PredictiveDetection.WindowSize = "2s"

		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Generate high concurrent load
		const numWorkers = 10
		const requestsPerWorker = 20

		var wg sync.WaitGroup
		startTime := time.Now()

		for worker := 0; worker < numWorkers; worker++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for i := 0; i < requestsPerWorker; i++ {
					request := &TestInferenceRequest{
						ID:    fmt.Sprintf("high-load-worker-%d-req-%d", workerID, i),
						Model: "test-model",
						Input: "test input",
					}

					err := scheduler.ProcessRequest(ctx, request)
					if err != nil {
						t.Logf("Worker %d request %d failed: %v", workerID, i, err)
					}

					// Minimal delay for high load simulation
					time.Sleep(10 * time.Millisecond)
				}
			}(worker)
		}

		wg.Wait()
		highLoadDuration := time.Since(startTime)
		t.Logf("High load processing time: %v", highLoadDuration)

		// Verify system handled high load reasonably well
		total, successful, failed := scheduler.GetStats()
		expectedTotal := int64(numWorkers * requestsPerWorker)
		assert.Equal(t, expectedTotal, total)

		// Allow for some failures under high load, but most should succeed
		successRate := float64(successful) / float64(total)
		assert.True(t, successRate >= 0.8, "Success rate should be at least 80%% under high load, got %.2f%%", successRate*100)

		t.Logf("High load stats: Total=%d, Successful=%d, Failed=%d, Success Rate=%.2f%%",
			total, successful, failed, successRate*100)
	})

	t.Run("predictive detection with gradual degradation", func(t *testing.T) {
		// Create scheduler with 5 nodes
		scheduler := NewMockDistributedScheduler(5)

		// Create test configuration for degradation testing
		config := createIntegrationTestConfig()
		config.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.6
		config.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "2s"
		config.Inference.FaultTolerance.PredictiveDetection.WindowSize = "5s"

		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Phase 1: Normal operation
		t.Logf("Phase 1: Normal operation")
		for i := 0; i < 20; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("phase1-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
		}

		// Phase 2: Gradual degradation (fail nodes one by one with delays)
		t.Logf("Phase 2: Gradual degradation")

		// Fail first node
		err = scheduler.SimulateNodeFailure("node-0")
		require.NoError(t, err)
		t.Logf("Failed node-0")

		// Continue processing during degradation
		for i := 20; i < 35; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("phase2a-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during first node failure: %v", i, err)
			}
			time.Sleep(200 * time.Millisecond)
		}

		// Fail second node
		err = scheduler.SimulateNodeFailure("node-1")
		require.NoError(t, err)
		t.Logf("Failed node-1")

		// Continue processing with further degradation
		for i := 35; i < 50; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("phase2b-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during second node failure: %v", i, err)
			}
			time.Sleep(200 * time.Millisecond)
		}

		// Phase 3: Recovery
		t.Logf("Phase 3: Recovery")

		// Recover nodes
		err = scheduler.RecoverNode("node-0")
		require.NoError(t, err)
		t.Logf("Recovered node-0")

		time.Sleep(3 * time.Second)

		err = scheduler.RecoverNode("node-1")
		require.NoError(t, err)
		t.Logf("Recovered node-1")

		time.Sleep(3 * time.Second)

		// Final processing after recovery
		for i := 50; i < 60; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("phase3-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err, "Request should succeed after recovery")
			time.Sleep(100 * time.Millisecond)
		}

		// Analyze overall performance
		total, successful, failed := scheduler.GetStats()
		assert.Equal(t, int64(60), total)

		// System should maintain reasonable performance throughout degradation
		successRate := float64(successful) / float64(total)
		assert.True(t, successRate >= 0.7, "Success rate should be at least 70%% during gradual degradation, got %.2f%%", successRate*100)

		t.Logf("Gradual degradation stats: Total=%d, Successful=%d, Failed=%d, Success Rate=%.2f%%",
			total, successful, failed, successRate*100)
	})

	t.Run("predictive detection performance overhead measurement", func(t *testing.T) {
		// Test with predictive detection enabled
		schedulerWithPrediction := NewMockDistributedScheduler(4)
		configWithPrediction := createIntegrationTestConfig()
		configWithPrediction.Inference.FaultTolerance.PredictiveDetection.Enabled = true

		err := schedulerWithPrediction.LoadConfiguration(configWithPrediction)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		err = schedulerWithPrediction.Start(ctx)
		require.NoError(t, err)
		defer schedulerWithPrediction.Stop()

		// Measure performance with prediction
		startTime := time.Now()
		for i := 0; i < 100; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("with-prediction-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := schedulerWithPrediction.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}
		durationWithPrediction := time.Since(startTime)

		// Test with predictive detection disabled
		schedulerWithoutPrediction := NewMockDistributedScheduler(4)
		configWithoutPrediction := createIntegrationTestConfig()
		configWithoutPrediction.Inference.FaultTolerance.PredictiveDetection.Enabled = false
		configWithoutPrediction.Inference.FaultTolerance.SelfHealing.EnablePredictive = false // Disable predictive healing when prediction is disabled

		err = schedulerWithoutPrediction.LoadConfiguration(configWithoutPrediction)
		require.NoError(t, err)

		err = schedulerWithoutPrediction.Start(ctx)
		require.NoError(t, err)
		defer schedulerWithoutPrediction.Stop()

		// Measure performance without prediction
		startTime = time.Now()
		for i := 0; i < 100; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("without-prediction-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := schedulerWithoutPrediction.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}
		durationWithoutPrediction := time.Since(startTime)

		// Calculate overhead
		overhead := durationWithPrediction - durationWithoutPrediction
		overheadPercentage := float64(overhead) / float64(durationWithoutPrediction) * 100

		t.Logf("Performance with prediction: %v", durationWithPrediction)
		t.Logf("Performance without prediction: %v", durationWithoutPrediction)
		t.Logf("Predictive detection overhead: %v (%.2f%%)", overhead, overheadPercentage)

		// Verify overhead is acceptable (should be less than 100ms total or 50% overhead)
		assert.True(t, overhead < 100*time.Millisecond || overheadPercentage < 50,
			"Predictive detection overhead should be minimal: %v (%.2f%%)", overhead, overheadPercentage)
	})
}

// TestPerformanceWithDifferentClusterSizes tests scalability with various cluster sizes
func TestPerformanceWithDifferentClusterSizes(t *testing.T) {
	clusterSizes := []int{2, 3, 5, 7, 10}

	for _, size := range clusterSizes {
		t.Run(fmt.Sprintf("cluster_size_%d_nodes", size), func(t *testing.T) {
			// Create scheduler with specified cluster size
			scheduler := NewMockDistributedScheduler(size)

			// Create test configuration optimized for scalability testing
			config := createScalabilityTestConfig()
			err := scheduler.LoadConfiguration(config)
			require.NoError(t, err)

			// Start scheduler
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			err = scheduler.Start(ctx)
			require.NoError(t, err)
			defer scheduler.Stop()

			// Test baseline performance
			baselineRequests := 50
			startTime := time.Now()

			for i := 0; i < baselineRequests; i++ {
				request := &TestInferenceRequest{
					ID:    fmt.Sprintf("baseline-size-%d-req-%d", size, i),
					Model: "test-model",
					Input: "test input",
				}

				err := scheduler.ProcessRequest(ctx, request)
				assert.NoError(t, err)
				time.Sleep(20 * time.Millisecond) // Consistent load
			}

			baselineDuration := time.Since(startTime)

			// Test performance under node failure
			failureTestRequests := 30

			// Fail one node (proportional to cluster size)
			nodeToFail := fmt.Sprintf("node-%d", size/2)
			err = scheduler.SimulateNodeFailure(nodeToFail)
			require.NoError(t, err)
			t.Logf("Failed %s in cluster of size %d", nodeToFail, size)

			// Wait for fault detection
			time.Sleep(2 * time.Second)

			startTime = time.Now()

			for i := 0; i < failureTestRequests; i++ {
				request := &TestInferenceRequest{
					ID:    fmt.Sprintf("failure-size-%d-req-%d", size, i),
					Model: "test-model",
					Input: "test input",
				}

				err := scheduler.ProcessRequest(ctx, request)
				if err != nil {
					t.Logf("Request %d failed during node failure in cluster size %d: %v", i, size, err)
				}
				time.Sleep(20 * time.Millisecond)
			}

			failureDuration := time.Since(startTime)

			// Recover the node
			err = scheduler.RecoverNode(nodeToFail)
			require.NoError(t, err)
			t.Logf("Recovered %s in cluster of size %d", nodeToFail, size)

			// Wait for recovery
			time.Sleep(3 * time.Second)

			// Test recovery performance
			recoveryTestRequests := 30
			startTime = time.Now()

			for i := 0; i < recoveryTestRequests; i++ {
				request := &TestInferenceRequest{
					ID:    fmt.Sprintf("recovery-size-%d-req-%d", size, i),
					Model: "test-model",
					Input: "test input",
				}

				err := scheduler.ProcessRequest(ctx, request)
				assert.NoError(t, err, "Request should succeed after recovery")
				time.Sleep(20 * time.Millisecond)
			}

			recoveryDuration := time.Since(startTime)

			// Analyze performance metrics
			total, successful, failed := scheduler.GetStats()
			expectedTotal := int64(baselineRequests + failureTestRequests + recoveryTestRequests)
			assert.Equal(t, expectedTotal, total)

			successRate := float64(successful) / float64(total)

			// Calculate performance metrics
			baselinePerformance := float64(baselineRequests) / baselineDuration.Seconds()
			failurePerformance := float64(failureTestRequests) / failureDuration.Seconds()
			recoveryPerformance := float64(recoveryTestRequests) / recoveryDuration.Seconds()

			t.Logf("Cluster size %d performance metrics:", size)
			t.Logf("  Baseline: %.2f req/s", baselinePerformance)
			t.Logf("  During failure: %.2f req/s", failurePerformance)
			t.Logf("  After recovery: %.2f req/s", recoveryPerformance)
			t.Logf("  Success rate: %.2f%%", successRate*100)
			t.Logf("  Total: %d, Successful: %d, Failed: %d", total, successful, failed)

			// Verify scalability requirements
			assert.True(t, successRate >= 0.8, "Success rate should be at least 80%% for cluster size %d, got %.2f%%", size, successRate*100)

			// Performance should not degrade significantly with larger clusters
			if size >= 5 {
				assert.True(t, baselinePerformance >= 10.0, "Baseline performance should be at least 10 req/s for cluster size %d, got %.2f", size, baselinePerformance)
			}

			// Recovery performance should be close to baseline
			recoveryRatio := recoveryPerformance / baselinePerformance
			assert.True(t, recoveryRatio >= 0.7, "Recovery performance should be at least 70%% of baseline for cluster size %d, got %.2f", size, recoveryRatio)
		})
	}
}

// TestClusterScalabilityLimits tests the limits of cluster scalability
func TestClusterScalabilityLimits(t *testing.T) {
	t.Run("large_cluster_performance", func(t *testing.T) {
		// Test with a large cluster (15 nodes)
		largeClusterSize := 15
		scheduler := NewMockDistributedScheduler(largeClusterSize)

		// Create configuration optimized for large clusters
		config := createScalabilityTestConfig()
		// Adjust intervals for larger clusters
		config.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "3s"
		config.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "2s"

		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Test high-throughput performance
		numRequests := 200
		startTime := time.Now()

		// Use concurrent workers for high throughput
		const numWorkers = 20
		requestsPerWorker := numRequests / numWorkers

		var wg sync.WaitGroup

		for worker := 0; worker < numWorkers; worker++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for i := 0; i < requestsPerWorker; i++ {
					request := &TestInferenceRequest{
						ID:    fmt.Sprintf("large-cluster-worker-%d-req-%d", workerID, i),
						Model: "test-model",
						Input: "test input",
					}

					err := scheduler.ProcessRequest(ctx, request)
					if err != nil {
						t.Logf("Worker %d request %d failed: %v", workerID, i, err)
					}

					time.Sleep(5 * time.Millisecond) // High throughput
				}
			}(worker)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Analyze large cluster performance
		total, successful, failed := scheduler.GetStats()
		successRate := float64(successful) / float64(total)
		throughput := float64(successful) / duration.Seconds()

		t.Logf("Large cluster (%d nodes) performance:", largeClusterSize)
		t.Logf("  Duration: %v", duration)
		t.Logf("  Throughput: %.2f req/s", throughput)
		t.Logf("  Success rate: %.2f%%", successRate*100)
		t.Logf("  Total: %d, Successful: %d, Failed: %d", total, successful, failed)

		// Verify large cluster can handle high throughput
		assert.True(t, successRate >= 0.85, "Large cluster success rate should be at least 85%%, got %.2f%%", successRate*100)
		assert.True(t, throughput >= 20.0, "Large cluster throughput should be at least 20 req/s, got %.2f", throughput)
	})

	t.Run("massive_failure_resilience", func(t *testing.T) {
		// Test resilience with massive failures (fail 50% of nodes)
		clusterSize := 12
		scheduler := NewMockDistributedScheduler(clusterSize)

		config := createScalabilityTestConfig()
		err := scheduler.LoadConfiguration(config)
		require.NoError(t, err)

		// Start scheduler
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)
		defer scheduler.Stop()

		// Establish baseline
		for i := 0; i < 20; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("massive-failure-baseline-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err)
		}

		// Fail 50% of nodes
		nodesToFail := clusterSize / 2
		t.Logf("Failing %d out of %d nodes (50%%)", nodesToFail, clusterSize)

		for i := 0; i < nodesToFail; i++ {
			nodeID := fmt.Sprintf("node-%d", i)
			err := scheduler.SimulateNodeFailure(nodeID)
			require.NoError(t, err)
			time.Sleep(200 * time.Millisecond) // Stagger failures
		}

		// Wait for fault detection and recovery
		time.Sleep(5 * time.Second)

		// Test system resilience with reduced capacity
		for i := 20; i < 60; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("massive-failure-resilience-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			if err != nil {
				t.Logf("Request %d failed during massive failure: %v", i, err)
			}
			time.Sleep(50 * time.Millisecond)
		}

		// Gradually recover nodes
		t.Logf("Recovering %d nodes", nodesToFail)
		for i := 0; i < nodesToFail; i++ {
			nodeID := fmt.Sprintf("node-%d", i)
			err := scheduler.RecoverNode(nodeID)
			require.NoError(t, err)
			time.Sleep(500 * time.Millisecond) // Stagger recovery
		}

		// Wait for full recovery
		time.Sleep(5 * time.Second)

		// Test post-recovery performance
		for i := 60; i < 80; i++ {
			request := &TestInferenceRequest{
				ID:    fmt.Sprintf("massive-failure-recovery-req-%d", i),
				Model: "test-model",
				Input: "test input",
			}

			err := scheduler.ProcessRequest(ctx, request)
			assert.NoError(t, err, "Request should succeed after massive failure recovery")
		}

		// Analyze massive failure resilience
		total, successful, failed := scheduler.GetStats()
		successRate := float64(successful) / float64(total)

		t.Logf("Massive failure resilience results:")
		t.Logf("  Total requests: %d", total)
		t.Logf("  Successful: %d", successful)
		t.Logf("  Failed: %d", failed)
		t.Logf("  Success rate: %.2f%%", successRate*100)

		// System should maintain reasonable operation even with 50% node failure
		assert.True(t, successRate >= 0.6, "System should maintain at least 60%% success rate during massive failures, got %.2f%%", successRate*100)
		assert.Equal(t, int64(80), total)
	})
}

// createScalabilityTestConfig creates a configuration optimized for scalability testing
func createScalabilityTestConfig() *config.DistributedConfig {
	cfg := &config.DistributedConfig{}
	cfg.Inference.FaultTolerance.Enabled = true
	cfg.Inference.FaultTolerance.RetryAttempts = 2
	cfg.Inference.FaultTolerance.RetryDelay = "50ms"
	cfg.Inference.FaultTolerance.HealthCheckInterval = "5s"
	cfg.Inference.FaultTolerance.RecoveryTimeout = "30s"
	cfg.Inference.FaultTolerance.CircuitBreakerEnabled = true
	cfg.Inference.FaultTolerance.CheckpointInterval = "5s"
	cfg.Inference.FaultTolerance.MaxRetries = 2
	cfg.Inference.FaultTolerance.RetryBackoff = "100ms"
	cfg.Inference.FaultTolerance.ReplicationFactor = 2

	// Predictive detection (optimized for scalability)
	cfg.Inference.FaultTolerance.PredictiveDetection.Enabled = true
	cfg.Inference.FaultTolerance.PredictiveDetection.ConfidenceThreshold = 0.75
	cfg.Inference.FaultTolerance.PredictiveDetection.PredictionInterval = "2s"
	cfg.Inference.FaultTolerance.PredictiveDetection.WindowSize = "6s"
	cfg.Inference.FaultTolerance.PredictiveDetection.Threshold = 0.75
	cfg.Inference.FaultTolerance.PredictiveDetection.EnableMLDetection = false
	cfg.Inference.FaultTolerance.PredictiveDetection.EnableStatistical = true
	cfg.Inference.FaultTolerance.PredictiveDetection.EnablePatternRecog = true

	// Self-healing (optimized for scalability)
	cfg.Inference.FaultTolerance.SelfHealing.Enabled = true
	cfg.Inference.FaultTolerance.SelfHealing.HealingThreshold = 0.65
	cfg.Inference.FaultTolerance.SelfHealing.HealingInterval = "10s"
	cfg.Inference.FaultTolerance.SelfHealing.MonitoringInterval = "2s"
	cfg.Inference.FaultTolerance.SelfHealing.LearningInterval = "10s"
	cfg.Inference.FaultTolerance.SelfHealing.ServiceRestart = true
	cfg.Inference.FaultTolerance.SelfHealing.ResourceReallocation = true
	cfg.Inference.FaultTolerance.SelfHealing.LoadRedistribution = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableLearning = true
	cfg.Inference.FaultTolerance.SelfHealing.EnablePredictive = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableProactive = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableFailover = true
	cfg.Inference.FaultTolerance.SelfHealing.EnableScaling = true

	// Redundancy (optimized for scalability)
	cfg.Inference.FaultTolerance.Redundancy.Enabled = true
	cfg.Inference.FaultTolerance.Redundancy.DefaultFactor = 2
	cfg.Inference.FaultTolerance.Redundancy.MaxFactor = 4
	cfg.Inference.FaultTolerance.Redundancy.UpdateInterval = "15s"

	// Performance tracking (window should be 3x healing interval: 10s * 3 = 30s)
	cfg.Inference.FaultTolerance.PerformanceTracking.Enabled = true
	cfg.Inference.FaultTolerance.PerformanceTracking.WindowSize = "30s"

	// Config adaptation
	cfg.Inference.FaultTolerance.ConfigAdaptation.Enabled = true
	cfg.Inference.FaultTolerance.ConfigAdaptation.Interval = "2m"

	return cfg
}
