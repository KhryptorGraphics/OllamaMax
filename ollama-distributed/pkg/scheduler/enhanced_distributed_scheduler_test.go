package scheduler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

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
			"large_model":     5.0 * 1024 * 1024 * 1024, // 5GB
			"large_context":   2048,
			"many_layers":     20,
			"high_parallelism": 0.8,
		},
		LearningRate:                   0.1,
		IntelligentLoadBalancingAlgorithm: "predictive",
		LoadBalancingWeightFactors: map[string]float64{
			"latency":    0.4,
			"throughput": 0.3,
			"reliability": 0.2,
			"capacity":   0.1,
		},
		AdvancedFaultToleranceStrategy: "hybrid",
		FaultRecoveryTimeout:           30 * time.Second,
		AdvisorDecisionTimeout:         5 * time.Second,
		AdvisorLearningRate:           0.1,
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
	
	// Test available strategies
	strategies := scheduler.GetAvailableStrategies()
	if len(strategies) == 0 {
		t.Error("Expected available strategies to be non-empty")
	}
	
	// Test strategy metrics
	metrics := scheduler.GetStrategyMetrics()
	if metrics == nil {
		t.Error("Expected strategy metrics to be non-nil")
	}
	
	// Test selection history
	history := scheduler.GetSelectionHistory()
	if history == nil {
		t.Error("Expected selection history to be non-nil")
	}
	
	// Test metrics
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

// TestPerformanceTracker tests the performance tracker
func TestPerformanceTracker(t *testing.T) {
	tracker := NewPerformanceTracker(100, 30*time.Second, true)
	
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
	
	if !tracker.learning {
		t.Error("Expected learning to be enabled")
	}
	
	// Test recording performance
	record := &PerformanceRecord{
		Timestamp: time.Now(),
		TaskID:    "test_task",
		ModelName: "test_model",
		Latency:   100 * time.Millisecond,
		Success:   true,
		Metadata:  make(map[string]interface{}),
		ResourceUsage: &NodeResourceState{
			CPUUtilization:    50.0,
			MemoryUtilization: 60.0,
			GPUUtilization:    40.0,
			NetworkUtilization: 30.0,
			ActiveRequests:    5,
			QueuedRequests:    2,
			LoadAverage:       1.5,
		},
	}
	
	tracker.RecordPerformance(record)
	
	if len(tracker.history) != 1 {
		t.Errorf("Expected history size 1, got %d", len(tracker.history))
	}
	
	// Test metrics retrieval
	metrics := tracker.GetAggregatedMetrics()
	if metrics == nil {
		t.Error("Expected aggregated metrics to be non-nil")
	}
	
	if metrics.TotalTasks != 1 {
		t.Errorf("Expected total tasks 1, got %d", metrics.TotalTasks)
	}
	
	if metrics.SuccessfulTasks != 1 {
		t.Errorf("Expected successful tasks 1, got %d", metrics.SuccessfulTasks)
	}
	
	if metrics.FailedTasks != 0 {
		t.Errorf("Expected failed tasks 0, got %d", metrics.FailedTasks)
	}
	
	if metrics.AverageLatency != 100*time.Millisecond {
		t.Errorf("Expected average latency 100ms, got %v", metrics.AverageLatency)
	}
	
	if metrics.SuccessRate != 1.0 {
		t.Errorf("Expected success rate 1.0, got %f", metrics.SuccessRate)
	}
	
	if metrics.ErrorRate != 0.0 {
		t.Errorf("Expected error rate 0.0, got %f", metrics.ErrorRate)
	}
	
	// Test multiple records
	for i := 0; i < 5; i++ {
		record := &PerformanceRecord{
			Timestamp: time.Now(),
			TaskID:    fmt.Sprintf("test_task_%d", i),
			ModelName: "test_model",
			Latency:   time.Duration(50+i*10) * time.Millisecond,
			Success:   i%2 == 0, // Alternate success/failure
			Metadata:  make(map[string]interface{}),
			ResourceUsage: &NodeResourceState{
				CPUUtilization:    float64(30 + i*5),
				MemoryUtilization: float64(40 + i*5),
				GPUUtilization:    float64(20 + i*5),
				NetworkUtilization: float64(10 + i*5),
				ActiveRequests:    3 + i,
				QueuedRequests:    1 + i%2,
				LoadAverage:       0.5 + float64(i)*0.2,
			},
		}
		
		tracker.RecordPerformance(record)
	}
	
	// Test updated metrics
	metrics = tracker.GetAggregatedMetrics()
	if metrics == nil {
		t.Error("Expected aggregated metrics to be non-nil")
	}
	
	if metrics.TotalTasks != 6 {
		t.Errorf("Expected total tasks 6, got %d", metrics.TotalTasks)
	}
	
	if metrics.SuccessfulTasks != 3 {
		t.Errorf("Expected successful tasks 3, got %d", metrics.SuccessfulTasks)
	}
	
	if metrics.FailedTasks != 3 {
		t.Errorf("Expected failed tasks 3, got %d", metrics.FailedTasks)
	}
	
	if metrics.SuccessRate != 0.5 {
		t.Errorf("Expected success rate 0.5, got %f", metrics.SuccessRate)
	}
	
	if metrics.ErrorRate != 0.5 {
		t.Errorf("Expected error rate 0.5, got %f", metrics.ErrorRate)
	}
	
	if metrics.LastUpdated.IsZero() {
		t.Error("Expected last updated time to be set")
	}
}

// TestSchedulingAdvisor tests the scheduling advisor
func TestSchedulingAdvisor(t *testing.T) {
	advisor := NewSchedulingAdvisor(0.1, 5*time.Second, true)
	
	if advisor == nil {
		t.Fatal("Expected scheduling advisor to be non-nil")
	}
	
	if advisor.history == nil {
		t.Error("Expected history to be non-nil")
	}
	
	if advisor.patterns == nil {
		t.Error("Expected patterns to be non-nil")
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
	
	if !advisor.learning {
		t.Error("Expected learning to be enabled")
	}
	
	// Test recommendation
	req := &Request{
		ID:        "test_request",
		ModelName: "llama2",
		Type:      "inference",
		Priority:  1,
		Timeout:   30 * time.Second,
		Metadata:  make(map[string]string),
		Payload:   make(map[string]interface{}),
		ResponseCh: make(chan *Response, 1),
		CreatedAt: time.Now(),
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
	
	// Test with classification request
	req.Type = "classification"
	recommendation = advisor.GetRecommendation(req)
	if recommendation != "task_parallel" {
		t.Errorf("Expected 'task_parallel' recommendation for classification, got '%s'", recommendation)
	}
	
	// Test with unknown request type
	req.Type = "unknown"
	recommendation = advisor.GetRecommendation(req)
	if recommendation != "round_robin" {
		t.Errorf("Expected 'round_robin' recommendation for unknown type, got '%s'", recommendation)
	}
}

// TestRecommendationEngine tests the recommendation engine
func TestRecommendationEngine(t *testing.T) {
	engine := NewRecommendationEngine()
	
	if engine == nil {
		t.Fatal("Expected recommendation engine to be non-nil")
	}
	
	if engine.patterns == nil {
		t.Error("Expected patterns to be non-nil")
	}
	
	if engine.algorithms == nil {
		t.Error("Expected algorithms to be non-nil")
	}
	
	if !engine.learning {
		t.Error("Expected learning to be enabled")
	}
	
	if engine.accuracy != 0.7 {
		t.Errorf("Expected accuracy 0.7, got %f", engine.accuracy)
	}
	
	// Test generating recommendation with empty patterns
	req := &Request{
		ID:        "test_request",
		ModelName: "test_model",
		Type:      "inference",
		Priority:  1,
		Timeout:   30 * time.Second,
		Metadata:  make(map[string]string),
		Payload:   make(map[string]interface{}),
		ResponseCh: make(chan *Response, 1),
		CreatedAt: time.Now(),
	}
	
	patterns := make([]*SchedulingPattern, 0)
	
	_, err := engine.GenerateRecommendation(req, patterns)
	if err == nil {
		t.Error("Expected error for empty patterns")
	}
	
	// Test with valid patterns
	patterns = append(patterns, &SchedulingPattern{
		ID:          "pattern_layerwise",
		Name:        "layerwise",
		Description: "Layerwise partitioning pattern",
		Conditions: map[string]interface{}{
			"model_type": "transformer",
			"layer_count": ">20",
		},
		Strategies: []string{"layerwise"},
		Confidence: 0.8,
		LastUpdated: time.Now(),
		Metadata: make(map[string]interface{}),
	})
	
	patterns = append(patterns, &SchedulingPattern{
		ID:          "pattern_data_split",
		Name:        "data_split",
		Description: "Data split partitioning pattern",
		Conditions: map[string]interface{}{
			"context_length": ">2048",
		},
		Strategies: []string{"data_split"},
		Confidence: 0.9,
		LastUpdated: time.Now(),
		Metadata: make(map[string]interface{}),
	})
	
	// Test generating recommendation
	strategy, err := engine.GenerateRecommendation(req, patterns)
	if err != nil {
		t.Errorf("Unexpected error generating recommendation: %v", err)
	}
	
	if strategy != "data_split" {
		t.Errorf("Expected strategy 'data_split', got '%s'", strategy)
	}
	
	// Test accuracy methods
	if engine.GetAccuracy() != 0.7 {
		t.Errorf("Expected accuracy 0.7, got %f", engine.GetAccuracy())
	}
	
	// Test updating accuracy
	engine.UpdateAccuracy(true)
	if engine.accuracy <= 0.7 {
		t.Errorf("Expected accuracy to increase after successful update, got %f", engine.accuracy)
	}
	
	engine.UpdateAccuracy(false)
	// Accuracy may decrease or stay the same depending on implementation
	if engine.accuracy > 1.0 {
		t.Errorf("Expected accuracy to be <= 1.0, got %f", engine.accuracy)
	}
}

// mockP2PNode implements a mock P2P node for testing
type mockP2PNode struct{}

func (m *mockP2PNode) ID() string { return "mock-node-id" }

// mockConsensusEngine implements a mock consensus engine for testing
type mockConsensusEngine struct{}

func (m *mockConsensusEngine) IsLeader() bool { return true }
func (m *mockConsensusEngine) Leader() string { return "mock-leader-id" }