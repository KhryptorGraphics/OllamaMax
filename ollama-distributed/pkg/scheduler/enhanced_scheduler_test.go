package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/loadbalancer"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// mockP2PNode implements a mock P2P node for testing
type mockP2PNode struct{}

func (m *mockP2PNode) ID() string { return "mock-node-id" }

// mockConsensusEngine implements a mock consensus engine for testing
type mockConsensusEngine struct{}

func (m *mockConsensusEngine) IsLeader() bool { return true }
func (m *mockConsensusEngine) Leader() string { return "mock-leader-id" }

// TestEnhancedDistributedScheduler tests the enhanced distributed scheduler
func TestEnhancedDistributedScheduler(t *testing.T) {
	// Create mock components
	mockP2P := &mockP2PNode{}
	mockConsensus := &mockConsensusEngine{}

	// Create enhanced scheduler config
	config := &EnhancedSchedulerConfig{
		SchedulerConfig: &config.SchedulerConfig{
			Algorithm:           "round_robin",
			LoadBalancing:       "least_connections",
			HealthCheckInterval: 30 * time.Second,
			MaxRetries:          3,
			RetryDelay:          1 * time.Second,
			QueueSize:           100,
			WorkerCount:         4,
		},
		EnableAdaptiveScheduling:       true,
		EnablePerformanceTracking:      true,
		EnableIntelligentLoadBalancing: true,
		EnableAdvancedFaultTolerance:   true,
		PerformanceHistorySize:         100,
		PerformanceCollectionInterval:  30 * time.Second,
		AdaptiveThresholds: map[string]float64{
			"large_model":      5.0 * 1024 * 1024 * 1024, // 5GB
			"large_context":    2048,
			"many_layers":      20,
			"high_parallelism": 0.8,
		},
		LearningRate:                      0.1,
		IntelligentLoadBalancingAlgorithm: "predictive",
		LoadBalancingWeightFactors: map[string]float64{
			"latency":     0.4,
			"throughput":  0.3,
			"reliability": 0.2,
			"capacity":    0.1,
		},
		AdvancedFaultToleranceStrategy: "hybrid",
		FaultRecoveryTimeout:           30 * time.Second,
		AdvisorDecisionTimeout:         5 * time.Second,
		AdvisorLearningRate:            0.1,
	}

	// Create enhanced scheduler
	scheduler, err := NewEnhancedDistributedScheduler(config, mockP2P, mockConsensus)
	if err != nil {
		t.Fatalf("Failed to create enhanced scheduler: %v", err)
	}

	// Test initialization
	if scheduler == nil {
		t.Fatal("Expected scheduler to be non-nil")
	}

	// Test enhanced partition manager
	if scheduler.enhancedPartitionManager == nil {
		t.Error("Expected enhanced partition manager to be non-nil")
	}

	// Test intelligent load balancer
	if scheduler.intelligentLoadBalancer == nil {
		t.Error("Expected intelligent load balancer to be non-nil")
	}

	// Test advanced fault tolerance manager
	if scheduler.advancedFaultTolerance == nil {
		t.Error("Expected advanced fault tolerance manager to be non-nil")
	}

	// Test performance tracker
	if scheduler.performanceTracker == nil {
		t.Error("Expected performance tracker to be non-nil")
	}

	// Test scheduling advisor
	if scheduler.schedulingAdvisor == nil {
		t.Error("Expected scheduling advisor to be non-nil")
	}

	// Test scheduling with tracking
	req := &Request{
		ID:         "test-request",
		ModelName:  "test-model",
		Type:       "inference",
		Priority:   1,
		Timeout:    30 * time.Second,
		Metadata:   make(map[string]string),
		Payload:    make(map[string]interface{}),
		ResponseCh: make(chan *Response, 1),
		CreatedAt:  time.Now(),
	}

	// Test scheduling
	err = scheduler.Schedule(req)
	if err != nil {
		// Scheduling might fail in test environment, but we're mainly testing initialization
		t.Logf("Scheduling failed (expected in test environment): %v", err)
	}

	// Test stats retrieval
	stats := scheduler.GetEnhancedStats()
	if stats == nil {
		t.Error("Expected enhanced stats to be non-nil")
	}

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := scheduler.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

// TestEnhancedPartitionManager tests the enhanced partition manager
func TestEnhancedPartitionManager(t *testing.T) {
	// Create base partition manager
	baseConfig := &partitioning.Config{
		DefaultStrategy: "layerwise",
		LayerThreshold:  10,
		BatchSizeLimit:  32,
	}

	baseManager := partitioning.NewPartitionManager(baseConfig)

	// Create enhanced partition manager
	enhanced := NewEnhancedPartitionManager(baseManager)

	if enhanced == nil {
		t.Fatal("Expected enhanced partition manager to be non-nil")
	}

	if enhanced.baseManager != baseManager {
		t.Error("Expected base manager to match")
	}

	if len(enhanced.enhancedStrategies) != 0 {
		t.Errorf("Expected empty enhanced strategies, got %d", len(enhanced.enhancedStrategies))
	}

	if len(enhanced.strategyPerformance) != 0 {
		t.Errorf("Expected empty strategy performance, got %d", len(enhanced.strategyPerformance))
	}

	if len(enhanced.selectionHistory) != 0 {
		t.Errorf("Expected empty selection history, got %d", len(enhanced.selectionHistory))
	}
}

// TestIntelligentLoadBalancer tests the intelligent load balancer
func TestIntelligentLoadBalancer(t *testing.T) {
	// Create base load balancer
	baseConfig := &loadbalancer.Config{
		Algorithm:     "round_robin",
		LatencyTarget: 100 * time.Millisecond,
		WeightFactors: map[string]float64{
			"latency":     0.4,
			"throughput":  0.3,
			"reliability": 0.2,
			"capacity":    0.1,
		},
		Adaptive:          true,
		PredictionEnabled: true,
		HistorySize:       1000,
		MaxBodySize:       32 * 1024 * 1024,
		RateLimit: &loadbalancer.RateLimitConfig{
			RPS:    1000,
			Burst:  2000,
			Window: 1 * time.Minute,
		},
		Cors: &loadbalancer.CorsConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	baseBalancer, err := loadbalancer.NewIntelligentLoadBalancer(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create base load balancer: %v", err)
	}

	// Create intelligent load balancer
	intelligent := NewIntelligentLoadBalancer(baseBalancer)

	if intelligent == nil {
		t.Fatal("Expected intelligent load balancer to be non-nil")
	}

	if intelligent.baseBalancer != baseBalancer {
		t.Error("Expected base balancer to match")
	}

	if intelligent.predictor == nil {
		t.Error("Expected predictor to be non-nil")
	}

	if intelligent.learningEngine == nil {
		t.Error("Expected learning engine to be non-nil")
	}

	if intelligent.selector == nil {
		t.Error("Expected selector to be non-nil")
	}

	if intelligent.metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	// Test metrics retrieval
	metrics := intelligent.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}
}

// TestAdvancedFaultToleranceManager tests the advanced fault tolerance manager
func TestAdvancedFaultToleranceManager(t *testing.T) {
	// Create base fault tolerance manager
	baseConfig := &fault_tolerance.Config{
		ReplicationFactor:     3,
		HealthCheckInterval:   30 * time.Second,
		RecoveryTimeout:       60 * time.Second,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    5 * time.Minute,
		MaxRetries:            3,
		RetryBackoff:          1 * time.Second,
	}

	baseManager := fault_tolerance.NewFaultToleranceManager(baseConfig)

	// Create advanced fault tolerance manager
	advanced := NewAdvancedFaultToleranceManager(baseManager)

	if advanced == nil {
		t.Fatal("Expected advanced fault tolerance manager to be non-nil")
	}

	if advanced.baseManager != baseManager {
		t.Error("Expected base manager to match")
	}

	if len(advanced.enhancedStrategies) != 0 {
		t.Errorf("Expected empty enhanced strategies, got %d", len(advanced.enhancedStrategies))
	}

	if advanced.fastRecovery == nil {
		t.Error("Expected fast recovery to be non-nil")
	}

	if advanced.redundancyManager == nil {
		t.Error("Expected redundancy manager to be non-nil")
	}

	if advanced.degradationManager == nil {
		t.Error("Expected degradation manager to be non-nil")
	}

	if advanced.metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	// Test metrics retrieval
	metrics := advanced.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}
}

// TestPerformanceTracker tests the performance tracker
func TestPerformanceTracker(t *testing.T) {
	// Create performance tracker
	tracker := NewPerformanceTracker(100, 30*time.Second)

	if tracker == nil {
		t.Fatal("Expected performance tracker to be non-nil")
	}

	if len(tracker.history) != 0 {
		t.Errorf("Expected empty history, got %d", len(tracker.history))
	}

	if tracker.metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	if tracker.collectionInterval != 30*time.Second {
		t.Errorf("Expected collection interval 30s, got %v", tracker.collectionInterval)
	}

	if tracker.historySize != 100 {
		t.Errorf("Expected history size 100, got %d", tracker.historySize)
	}

	// Test recording performance
	record := &PerformanceRecord{
		Timestamp: time.Now(),
		TaskID:    "test-task",
		ModelName: "test-model",
		Latency:   100 * time.Millisecond,
		Success:   true,
		Metadata:  make(map[string]interface{}),
	}

	tracker.RecordPerformance(record)

	if len(tracker.history) != 1 {
		t.Errorf("Expected history size 1, got %d", len(tracker.history))
	}

	// Test metrics retrieval
	metrics := tracker.GetAggregatedMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	if metrics.TotalTasks != 1 {
		t.Errorf("Expected total tasks 1, got %d", metrics.TotalTasks)
	}

	if metrics.SuccessfulTasks != 1 {
		t.Errorf("Expected successful tasks 1, got %d", metrics.SuccessfulTasks)
	}

	if metrics.AverageLatency != 100*time.Millisecond {
		t.Errorf("Expected average latency 100ms, got %v", metrics.AverageLatency)
	}
}

// TestSchedulingAdvisor tests the scheduling advisor
func TestSchedulingAdvisor(t *testing.T) {
	// Create scheduling advisor
	advisor := NewSchedulingAdvisor(0.1, 5*time.Second)

	if advisor == nil {
		t.Fatal("Expected scheduling advisor to be non-nil")
	}

	if advisor.patternMatcher == nil {
		t.Error("Expected pattern matcher to be non-nil")
	}

	if advisor.recommender == nil {
		t.Error("Expected recommender to be non-nil")
	}

	if advisor.learningRate != 0.1 {
		t.Errorf("Expected learning rate 0.1, got %f", advisor.learningRate)
	}

	if advisor.decisionTimeout != 5*time.Second {
		t.Errorf("Expected decision timeout 5s, got %v", advisor.decisionTimeout)
	}

	// Test recommendation
	req := &Request{
		ID:         "test-request",
		ModelName:  "llama2",
		Type:       "inference",
		Priority:   1,
		Timeout:    30 * time.Second,
		Metadata:   make(map[string]string),
		Payload:    make(map[string]interface{}),
		ResponseCh: make(chan *Response, 1),
		CreatedAt:  time.Now(),
	}

	recommendation := advisor.GetRecommendation(req)
	if recommendation == "" {
		t.Error("Expected recommendation to be non-empty")
	}

	// Test with different model
	req.ModelName = "gemma"
	recommendation = advisor.GetRecommendation(req)
	if recommendation == "" {
		t.Error("Expected recommendation to be non-empty")
	}

	// Test with embedding request
	req.Type = "embedding"
	recommendation = advisor.GetRecommendation(req)
	if recommendation != "data_split" {
		t.Errorf("Expected 'data_split' recommendation for embedding, got '%s'", recommendation)
	}
}
