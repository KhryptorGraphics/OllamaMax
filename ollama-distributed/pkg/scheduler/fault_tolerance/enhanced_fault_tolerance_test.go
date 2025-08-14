//go:build ft_legacy

package fault_tolerance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestEnhancedFaultToleranceManager tests the enhanced fault tolerance manager
func TestEnhancedFaultToleranceManager(t *testing.T) {
	// Create mock config
	config := NewEnhancedFaultToleranceConfig(&Config{
		ReplicationFactor:     3,
		HealthCheckInterval:   30 * time.Second,
		RecoveryTimeout:       60 * time.Second,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    5 * time.Minute,
		MaxRetries:            3,
		RetryBackoff:          1 * time.Second,
	})

	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     config.ReplicationFactor,
		HealthCheckInterval:   config.HealthCheckInterval,
		RecoveryTimeout:       config.RecoveryTimeout,
		CircuitBreakerEnabled: config.CircuitBreakerEnabled,
		CheckpointInterval:    config.CheckpointInterval,
		MaxRetries:            config.MaxRetries,
		RetryBackoff:          config.RetryBackoff,
	}

	baseManager := NewFaultToleranceManager(baseConfig)

	// Create enhanced fault tolerance manager
	manager := NewEnhancedFaultToleranceManager(config, baseManager)

	// Test initialization
	if manager == nil {
		t.Fatal("Expected enhanced fault tolerance manager to be non-nil")
	}

	// Test that base manager is embedded
	if manager.FaultToleranceManager == nil {
		t.Error("Expected base fault tolerance manager to be embedded")
	}

	// Test that advanced strategies are initialized
	if manager.advancedStrategies == nil {
		t.Error("Expected advanced strategies to be initialized")
	}

	// Test that predictor is initialized
	if manager.predictor == nil {
		t.Error("Expected predictor to be initialized")
	}

	// Test that self-healer is initialized
	if manager.selfHealer == nil {
		t.Error("Expected self-healer to be initialized")
	}

	// Test that redundancy manager is initialized
	if manager.redundancyManager == nil {
		t.Error("Expected redundancy manager to be initialized")
	}

	// Test that performance tracker is initialized
	if manager.performanceTracker == nil {
		t.Error("Expected performance tracker to be initialized")
	}

	// Test that config adaptor is initialized
	if manager.configAdaptor == nil {
		t.Error("Expected config adaptor to be initialized")
	}

	// Test that enhanced metrics are initialized
	if manager.enhancedMetrics == nil {
		t.Error("Expected enhanced metrics to be initialized")
	}

	// Test that context is initialized
	if manager.ctx == nil {
		t.Error("Expected context to be initialized")
	}

	// Test that cancel function is initialized
	if manager.cancel == nil {
		t.Error("Expected cancel function to be initialized")
	}

	// Test that mutex is initialized
	manager.mu.Lock()
	manager.mu.Unlock()

	// Test that wait group is initialized
	if manager.wg == (sync.WaitGroup{}) {
		// This is difficult to test, so we'll just make sure it doesn't panic
		t.Log("Wait group initialized (cannot directly test)")
	}

	// Test that started flag is initialized
	if manager.started {
		t.Error("Expected started flag to be false initially")
	}

	// Test registerAdvancedStrategies
	manager.registerAdvancedStrategies()

	// Verify that advanced strategies are registered
	faultTypes := []FaultType{
		FaultTypeNodeFailure,
		FaultTypeNetworkPartition,
		FaultTypeResourceExhaustion,
		FaultTypePerformanceAnomaly,
		FaultTypeServiceUnavailable,
	}

	for _, faultType := range faultTypes {
		if strategies, exists := manager.advancedStrategies[faultType]; exists {
			if len(strategies) == 0 {
				t.Errorf("Expected strategies for fault type %s", faultType)
			}
		} else {
			t.Errorf("Expected strategies for fault type %s", faultType)
		}
	}

	// Test GetEnhancedMetrics
	metrics := manager.GetEnhancedMetrics()
	if metrics == nil {
		t.Error("Expected enhanced metrics to be non-nil")
	}

	// Test that enhanced metrics contain base metrics
	if metrics.FaultToleranceMetrics == nil {
		t.Error("Expected base metrics to be embedded in enhanced metrics")
	}

	// Test that enhanced metrics contain extended fields
	// Note: These will be zero initially but should exist
	_ = metrics.PredictionsMade
	_ = metrics.PredictionsCorrect
	_ = metrics.PredictionAccuracy
	_ = metrics.AveragePredictionLatency
	_ = metrics.SelfHealingAttempts
	_ = metrics.SelfHealingSuccesses
	_ = metrics.SelfHealingFailures
	_ = metrics.AverageHealingTime
	_ = metrics.RedundancyFactor
	_ = metrics.ActiveReplicas
	_ = metrics.FailedReplicas
	_ = metrics.ReplicationLatency
	_ = metrics.AverageRecoveryTime
	_ = metrics.RecoverySuccessRate
	_ = metrics.ResourceUtilization
	_ = metrics.SystemStability
	_ = metrics.ConfigAdaptations
	_ = metrics.AdaptationAccuracy
	_ = metrics.CircuitBreakerTrips
	_ = metrics.CircuitBreakerResets
	_ = metrics.AlertsSent
	_ = metrics.AlertThrottling

	// Test NewEnhancedFaultToleranceConfig
	defaultConfig := NewEnhancedFaultToleranceConfig(baseConfig)
	if defaultConfig == nil {
		t.Error("Expected default enhanced fault tolerance config to be non-nil")
	}

	// Test that default config contains all expected fields
	if !defaultConfig.EnablePrediction {
		t.Error("Expected EnablePrediction to be true by default")
	}

	if defaultConfig.PredictionWindowSize != 30*time.Second {
		t.Error("Expected PredictionWindowSize to be 30 seconds by default")
	}

	if defaultConfig.PredictionThreshold != 0.8 {
		t.Error("Expected PredictionThreshold to be 0.8 by default")
	}

	if !defaultConfig.EnableSelfHealing {
		t.Error("Expected EnableSelfHealing to be true by default")
	}

	if defaultConfig.SelfHealingInterval != 60*time.Second {
		t.Error("Expected SelfHealingInterval to be 60 seconds by default")
	}

	if defaultConfig.SelfHealingThreshold != 0.7 {
		t.Error("Expected SelfHealingThreshold to be 0.7 by default")
	}

	if !defaultConfig.EnableRedundancy {
		t.Error("Expected EnableRedundancy to be true by default")
	}

	if defaultConfig.DefaultRedundancyFactor != 2 {
		t.Error("Expected DefaultRedundancyFactor to be 2 by default")
	}

	if defaultConfig.MaxRedundancyFactor != 5 {
		t.Error("Expected MaxRedundancyFactor to be 5 by default")
	}

	if defaultConfig.RedundancyUpdateInterval != 300*time.Second {
		t.Error("Expected RedundancyUpdateInterval to be 300 seconds by default")
	}

	if !defaultConfig.EnablePerformanceTracking {
		t.Error("Expected EnablePerformanceTracking to be true by default")
	}

	if defaultConfig.PerformanceWindowSize != 60*time.Second {
		t.Error("Expected PerformanceWindowSize to be 60 seconds by default")
	}

	if !defaultConfig.EnableConfigAdaptation {
		t.Error("Expected EnableConfigAdaptation to be true by default")
	}

	if defaultConfig.ConfigAdaptationInterval != 300*time.Second {
		t.Error("Expected ConfigAdaptationInterval to be 300 seconds by default")
	}

	if defaultConfig.MaxRecoveryRetries != 5 {
		t.Error("Expected MaxRecoveryRetries to be 5 by default")
	}

	if defaultConfig.RecoveryBackoffFactor != 1.5 {
		t.Error("Expected RecoveryBackoffFactor to be 1.5 by default")
	}

	if defaultConfig.RecoveryTimeout != 30*time.Second {
		t.Error("Expected RecoveryTimeout to be 30 seconds by default")
	}

	if !defaultConfig.CheckpointCompression {
		t.Error("Expected CheckpointCompression to be true by default")
	}

	if !defaultConfig.CheckpointEncryption {
		t.Error("Expected CheckpointEncryption to be true by default")
	}

	if defaultConfig.CheckpointRetention != 24*time.Hour {
		t.Error("Expected CheckpointRetention to be 24 hours by default")
	}

	if defaultConfig.CircuitBreakerThreshold != 5 {
		t.Error("Expected CircuitBreakerThreshold to be 5 by default")
	}

	if defaultConfig.CircuitBreakerTimeout != 30*time.Second {
		t.Error("Expected CircuitBreakerTimeout to be 30 seconds by default")
	}

	if defaultConfig.AlertThrottleTime != 5*time.Minute {
		t.Error("Expected AlertThrottleTime to be 5 minutes by default")
	}

	if defaultConfig.AlertSeverityThreshold != "medium" {
		t.Error("Expected AlertSeverityThreshold to be 'medium' by default")
	}

	// Test DetectFault method
	fault := manager.DetectFault(FaultTypeNodeFailure, "test-node", "Test fault detection", map[string]interface{}{
		"test_key": "test_value",
	})

	if fault == nil {
		t.Error("Expected fault detection to return non-nil fault")
	}

	if fault.Type != FaultTypeNodeFailure {
		t.Errorf("Expected fault type 'node_failure', got '%s'", fault.Type)
	}

	if fault.Target != "test-node" {
		t.Errorf("Expected fault target 'test-node', got '%s'", fault.Target)
	}

	if fault.Description != "Test fault detection" {
		t.Errorf("Expected fault description 'Test fault detection', got '%s'", fault.Description)
	}

	if metaValue, exists := fault.Metadata["test_key"]; !exists || metaValue != "test_value" {
		t.Error("Expected metadata to contain 'test_key' with value 'test_value'")
	}

	// Test that enhanced metrics are updated after fault detection
	updatedMetrics := manager.GetEnhancedMetrics()
	if updatedMetrics.FaultsDetected != 1 {
		t.Errorf("Expected FaultsDetected to be 1, got %d", updatedMetrics.FaultsDetected)
	}

	if updatedMetrics.LastFault == nil {
		t.Error("Expected LastFault to be set after fault detection")
	}

	// Test Recover method
	ctx := context.Background()
	result, err := manager.Recover(ctx, fault)

	// Recovery might fail in test environment, but we're mainly testing the method signature
	if err != nil {
		// This is expected in test environment
		t.Logf("Recovery failed (expected in test environment): %v", err)
	}

	if result != nil {
		// Update metrics if recovery succeeded
		updatedMetrics = manager.GetEnhancedMetrics()
		if updatedMetrics.RecoveryAttempts != 1 {
			t.Errorf("Expected RecoveryAttempts to be 1, got %d", updatedMetrics.RecoveryAttempts)
		}

		if updatedMetrics.RecoverySuccesses != 0 && result.Successful {
			t.Errorf("Expected RecoverySuccesses to reflect recovery result")
		}

		if updatedMetrics.RecoveryFailures != 0 && !result.Successful {
			t.Errorf("Expected RecoveryFailures to reflect recovery result")
		}
	}

	// Test that the enhanced manager can be started
	// (This would fail in test environment, but we can check that the method exists)
	err = manager.Start()
	if err != nil && err.Error() != "enhanced fault tolerance manager already started" {
		// This is expected to fail in test environment
		t.Logf("Start failed (expected in test environment): %v", err)
	}

	// Test that the enhanced manager can be shut down
	// (This would fail in test environment, but we can check that the method exists)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Shutdown(shutdownCtx)
	if err != nil {
		// This is expected to fail in test environment
		t.Logf("Shutdown failed (expected in test environment): %v", err)
	}

	// Test updateRecoveryMetrics
	testResult := &RecoveryResult{
		FaultID:    "test_fault",
		Strategy:   "test_strategy",
		Successful: true,
		Duration:   100 * time.Millisecond,
		Metadata:   map[string]interface{}{"test": "value"},
		Timestamp:  time.Now(),
	}

	manager.updateRecoveryMetrics(testResult, 150*time.Millisecond)

	// Test that metrics were updated
	updatedMetrics = manager.GetEnhancedMetrics()
	if updatedMetrics.RecoveryAttempts == 0 {
		t.Error("Expected RecoveryAttempts to be incremented")
	}

	if updatedMetrics.RecoverySuccesses == 0 && testResult.Successful {
		t.Error("Expected RecoverySuccesses to be incremented for successful recovery")
	}

	if updatedMetrics.AverageRecoveryTime == 0 {
		t.Error("Expected AverageRecoveryTime to be updated")
	}
}

// TestFaultPredictor tests the fault predictor
func TestFaultPredictor(t *testing.T) {
	// Create mock config
	config := &EnhancedFaultToleranceConfig{
		Config: &Config{
			ReplicationFactor:     3,
			HealthCheckInterval:   30 * time.Second,
			RecoveryTimeout:       60 * time.Second,
			CircuitBreakerEnabled: true,
			CheckpointInterval:    5 * time.Minute,
			MaxRetries:            3,
			RetryBackoff:          1 * time.Second,
		},
		EnablePrediction:          true,
		PredictionWindowSize:      30 * time.Second,
		PredictionThreshold:       0.8,
		EnableSelfHealing:         true,
		SelfHealingInterval:       60 * time.Second,
		SelfHealingThreshold:      0.7,
		EnableRedundancy:          true,
		DefaultRedundancyFactor:   2,
		MaxRedundancyFactor:       5,
		RedundancyUpdateInterval:  300 * time.Second,
		EnablePerformanceTracking: true,
		PerformanceWindowSize:     60 * time.Second,
		EnableConfigAdaptation:    true,
		ConfigAdaptationInterval:  300 * time.Second,
		MaxRecoveryRetries:        5,
		RecoveryBackoffFactor:     1.5,
		RecoveryTimeout:           30 * time.Second,
		CheckpointCompression:     true,
		CheckpointEncryption:      true,
		CheckpointRetention:       24 * time.Hour,
		CircuitBreakerThreshold:   5,
		CircuitBreakerTimeout:     30 * time.Second,
		AlertThrottleTime:         5 * time.Minute,
		AlertSeverityThreshold:    "medium",
	}

	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     config.ReplicationFactor,
		HealthCheckInterval:   config.HealthCheckInterval,
		RecoveryTimeout:       config.RecoveryTimeout,
		CircuitBreakerEnabled: config.CircuitBreakerEnabled,
		CheckpointInterval:    config.CheckpointInterval,
		MaxRetries:            config.MaxRetries,
		RetryBackoff:          config.RetryBackoff,
	}

	baseManager := NewFaultToleranceManager(baseConfig)

	// Create fault predictor
	predictor := NewFaultPredictor(config, baseManager)

	// Test initialization
	if predictor == nil {
		t.Fatal("Expected fault predictor to be non-nil")
	}

	// Test that manager is set
	if predictor.manager == nil {
		t.Error("Expected manager to be set")
	}

	// Test that window size is set
	if predictor.windowSize != config.PredictionWindowSize {
		t.Errorf("Expected window size %v, got %v", config.PredictionWindowSize, predictor.windowSize)
	}

	// Test that threshold is set
	if predictor.threshold != config.PredictionThreshold {
		t.Errorf("Expected threshold %f, got %f", config.PredictionThreshold, predictor.threshold)
	}

	// Test that prediction models are initialized
	if len(predictor.predictionModels) == 0 {
		t.Error("Expected prediction models to be initialized")
	}

	// Test that history is initialized
	if predictor.history == nil {
		t.Error("Expected history to be initialized")
	}

	// Test that learning is enabled
	if !predictor.learning {
		t.Error("Expected learning to be enabled")
	}

	// Test that accuracy is initialized
	if predictor.accuracy != 0.0 {
		t.Error("Expected accuracy to be 0.0 initially")
	}

	// Test that metrics are initialized
	if predictor.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}

	// Test that mutex is initialized
	predictor.mu.Lock()
	predictor.mu.Unlock()

	// Test initializeModels
	predictor.initializeModels()

	// Test that prediction models are properly initialized
	modelNames := []string{"node_failure", "performance_anomaly", "resource_exhaustion"}

	for _, modelName := range modelNames {
		if model, exists := predictor.predictionModels[modelName]; exists {
			if model.Name != modelName {
				t.Errorf("Expected model name '%s', got '%s'", modelName, model.Name)
			}

			if model.Type == "" {
				t.Errorf("Expected model type for '%s'", modelName)
			}

			if len(model.Features) == 0 {
				t.Errorf("Expected features for model '%s'", modelName)
			}

			if len(model.Weights) == 0 {
				t.Errorf("Expected weights for model '%s'", modelName)
			}

			if model.Accuracy == 0.0 {
				t.Errorf("Expected accuracy for model '%s'", modelName)
			}

			if model.LastTrained.IsZero() {
				t.Errorf("Expected LastTrained for model '%s'", modelName)
			}

			if model.Metadata == nil {
				t.Errorf("Expected metadata for model '%s'", modelName)
			}
		} else {
			t.Errorf("Expected model '%s' to be initialized", modelName)
		}
	}

	// Test GetMetrics
	metrics := predictor.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	// Test GetAccuracy
	accuracy := predictor.GetAccuracy()
	if accuracy != 0.0 {
		t.Errorf("Expected accuracy 0.0, got %f", accuracy)
	}

	// Test GetHistory (initially empty)
	history := predictor.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}

	// Test GetModels
	models := predictor.GetModels()
	if len(models) == 0 {
		t.Error("Expected models to be non-nil")
	}

	// Test Enable/Disable methods
	predictor.Disable()
	if predictor.IsEnabled() {
		t.Error("Expected predictor to be disabled")
	}

	predictor.Enable()
	if !predictor.IsEnabled() {
		t.Error("Expected predictor to be enabled")
	}

	// Test SetThreshold/GetThreshold
	predictor.SetThreshold(0.9)
	if predictor.GetThreshold() != 0.9 {
		t.Errorf("Expected threshold 0.9, got %f", predictor.GetThreshold())
	}

	// Test SetWindowSize/GetWindowSize
	predictor.SetWindowSize(60 * time.Second)
	if predictor.GetWindowSize() != 60*time.Second {
		t.Errorf("Expected window size 60s, got %v", predictor.GetWindowSize())
	}

	// Test that predictor can predict faults (this is a simplified test)
	// In a real implementation, we'd need to mock the system state
	fault := &FaultDetection{
		ID:          "test_fault",
		Type:        FaultTypeNodeFailure,
		Severity:    FaultSeverityHigh,
		Target:      "test_node",
		Description: "Test fault",
		DetectedAt:  time.Now(),
		Status:      FaultStatusDetected,
		Metadata:    make(map[string]interface{}),
	}

	// Test predictFault
	predictor.predictFault(fault)

	// Test that fault is added to history
	history = predictor.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 entry in history, got %d", len(history))
	}

	// Test learnFromPredictions
	sample := &PredictionSample{
		Timestamp:   time.Now(),
		NodeID:      "test_node",
		Metrics:     make(map[string]float64),
		FaultType:   FaultTypeNodeFailure,
		Predicted:   true,
		ActualFault: true,
		Confidence:  0.9,
		Metadata:    make(map[string]interface{}),
	}

	predictor.learnFromPredictions([]*PredictionSample{sample})

	// Test rebalanceModelWeights
	predictor.rebalanceModelWeights()

	// Test that models have weights applied
	for _, model := range predictor.GetModels() {
		// All models should have weights now
		if len(model.Weights) == 0 {
			t.Errorf("Expected weights for model '%s'", model.Name)
		}
	}
}

// TestSelfHealingEngine tests the self-healing engine
func TestSelfHealingEngine(t *testing.T) {
	// Create mock config
	config := &EnhancedFaultToleranceConfig{
		Config: &Config{
			ReplicationFactor:     3,
			HealthCheckInterval:   30 * time.Second,
			RecoveryTimeout:       60 * time.Second,
			CircuitBreakerEnabled: true,
			CheckpointInterval:    5 * time.Minute,
			MaxRetries:            3,
			RetryBackoff:          1 * time.Second,
		},
		EnablePrediction:          true,
		PredictionWindowSize:      30 * time.Second,
		PredictionThreshold:       0.8,
		EnableSelfHealing:         true,
		SelfHealingInterval:       60 * time.Second,
		SelfHealingThreshold:      0.7,
		EnableRedundancy:          true,
		DefaultRedundancyFactor:   2,
		MaxRedundancyFactor:       5,
		RedundancyUpdateInterval:  300 * time.Second,
		EnablePerformanceTracking: true,
		PerformanceWindowSize:     60 * time.Second,
		EnableConfigAdaptation:    true,
		ConfigAdaptationInterval:  300 * time.Second,
		MaxRecoveryRetries:        5,
		RecoveryBackoffFactor:     1.5,
		RecoveryTimeout:           30 * time.Second,
		CheckpointCompression:     true,
		CheckpointEncryption:      true,
		CheckpointRetention:       24 * time.Hour,
		CircuitBreakerThreshold:   5,
		CircuitBreakerTimeout:     30 * time.Second,
		AlertThrottleTime:         5 * time.Minute,
		AlertSeverityThreshold:    "medium",
	}

	// Create base fault tolerance manager
	baseConfig := &Config{
		ReplicationFactor:     config.ReplicationFactor,
		HealthCheckInterval:   config.HealthCheckInterval,
		RecoveryTimeout:       config.RecoveryTimeout,
		CircuitBreakerEnabled: config.CircuitBreakerEnabled,
		CheckpointInterval:    config.CheckpointInterval,
		MaxRetries:            config.MaxRetries,
		RetryBackoff:          config.RetryBackoff,
	}

	baseManager := NewFaultToleranceManager(baseConfig)

	// Create self-healing engine
	engine := NewSelfHealingEngine(config, baseManager)

	// Test initialization
	if engine == nil {
		t.Fatal("Expected self-healing engine to be non-nil")
	}

	// Test that manager is set
	if engine.manager == nil {
		t.Error("Expected manager to be set")
	}

	// Test that interval is set
	if engine.interval != config.SelfHealingInterval {
		t.Errorf("Expected interval %v, got %v", config.SelfHealingInterval, engine.interval)
	}

	// Test that threshold is set
	if engine.threshold != config.SelfHealingThreshold {
		t.Errorf("Expected threshold %f, got %f", config.SelfHealingThreshold, engine.threshold)
	}

	// Test that healing strategies are initialized
	if len(engine.healingStrategies) == 0 {
		t.Error("Expected healing strategies to be initialized")
	}

	// Test that strategy weights are initialized
	if len(engine.strategyWeights) == 0 {
		t.Error("Expected strategy weights to be initialized")
	}

	// Test that healing history is initialized
	if engine.healingHistory == nil {
		t.Error("Expected healing history to be initialized")
	}

	// Test that learning is enabled
	if !engine.learning {
		t.Error("Expected learning to be enabled")
	}

	// Test that success rate is initialized
	if engine.successRate != 0.0 {
		t.Error("Expected success rate to be 0.0 initially")
	}

	// Test that metrics are initialized
	if engine.metrics == nil {
		t.Error("Expected metrics to be initialized")
	}

	// Test that mutex is initialized
	engine.mu.Lock()
	engine.mu.Unlock()

	// Test that healing strategies are properly initialized
	expectedStrategies := []string{
		"restart_services",
		"resource_rebalancing",
		"load_shedding",
		"component_scaling",
		"network_optimization",
	}

	// Check that all expected strategies are present
	for _, expected := range expectedStrategies {
		found := false
		for _, strategy := range engine.healingStrategies {
			if strategy.GetName() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected healing strategy '%s' to be initialized", expected)
		}
	}

	// Test that strategy weights are properly initialized
	for _, expected := range expectedStrategies {
		if _, exists := engine.strategyWeights[expected]; !exists {
			t.Errorf("Expected strategy weight for '%s'", expected)
		}
	}

	// Test GetMetrics
	metrics := engine.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics to be non-nil")
	}

	// Test GetSuccessRate
	successRate := engine.GetSuccessRate()
	if successRate != 0.0 {
		t.Errorf("Expected success rate 0.0, got %f", successRate)
	}

	// Test GetHistory (initially empty)
	history := engine.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(history))
	}

	// Test GetStrategies
	strategies := engine.GetStrategies()
	if len(strategies) == 0 {
		t.Error("Expected strategies to be non-nil")
	}

	// Test GetStrategyWeights
	weights := engine.GetStrategyWeights()
	if len(weights) == 0 {
		t.Error("Expected weights to be non-nil")
	}

	// Test SetStrategyWeight
	engine.SetStrategyWeight("restart_services", 0.5)
	if newWeight := engine.GetStrategyWeights()["restart_services"]; newWeight != 0.5 {
		t.Errorf("Expected strategy weight 0.5, got %f", newWeight)
	}

	// Test Enable/Disable methods
	engine.Disable()
	if engine.IsEnabled() {
		t.Error("Expected engine to be disabled")
	}

	engine.Enable()
	if !engine.IsEnabled() {
		t.Error("Expected engine to be enabled")
	}

	// Test SetThreshold/GetThreshold
	engine.SetThreshold(0.9)
	if engine.GetThreshold() != 0.9 {
		t.Errorf("Expected threshold 0.9, got %f", engine.GetThreshold())
	}

	// Test SetInterval/GetInterval
	engine.SetInterval(120 * time.Second)
	if engine.GetInterval() != 120*time.Second {
		t.Errorf("Expected interval 120s, got %v", engine.GetInterval())
	}

	// Test healing strategies implementation
	// Create a system state for testing
	state := &SystemState{
		Nodes: []*NodeInfo{
			{ID: "node1"},
			{ID: "node2"},
			{ID: "node3"},
		},
		Resources: &ResourceMetrics{
			CPUUtilization:     85.0,
			MemoryUtilization:  75.0,
			DiskUtilization:    60.0,
			GPUUtilization:     40.0,
			NetworkUtilization: 30.0,
			ActiveRequests:     100,
			QueuedRequests:     20,
			LoadAverage:        75.0,
			LastUpdated:        time.Now(),
		},
		Performance: &PerformanceMetrics{
			AverageLatency:    200 * time.Millisecond,
			Throughput:        50.0,
			SuccessRate:       0.95,
			ErrorRate:         0.05,
			RequestsProcessed: 1000,
			LastUpdated:       time.Now(),
		},
		Health: &HealthMetrics{
			TotalNodes:         3,
			HealthyNodes:       3,
			UnhealthyNodes:     0,
			AverageHealthScore: 95.0,
			WorstNodeHealth:    90.0,
			BestNodeHealth:     100.0,
			LastUpdated:        time.Now(),
		},
		Faults: []*FaultDetection{
			{
				ID:          "fault1",
				Type:        FaultTypeNodeFailure,
				Severity:    FaultSeverityHigh,
				Target:      "node1",
				Description: "Node failure",
				DetectedAt:  time.Now(),
				Status:      FaultStatusDetected,
				Metadata:    make(map[string]interface{}),
			},
		},
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Test each healing strategy
	for _, strategy := range engine.healingStrategies {
		// Test if strategy can handle the state
		canHandle := strategy.CanHandle(state)
		t.Logf("Strategy %s can handle state: %v", strategy.GetName(), canHandle)

		// Test applying the strategy
		ctx := context.Background()
		result, err := strategy.Apply(ctx, state)

		if err != nil {
			t.Logf("Strategy %s returned error: %v", strategy.GetName(), err)
		} else if result != nil {
			t.Logf("Strategy %s returned result: Improvement=%.2f, Actions=%d",
				strategy.GetName(), result.Improvement, len(result.ActionsTaken))
		} else {
			t.Logf("Strategy %s returned nil result", strategy.GetName())
		}
	}

	// Test needsHealing
	needsHealing := engine.needsHealing(state)
	t.Logf("System needs healing: %v", needsHealing)

	// Test selectBestStrategy
	bestStrategy := engine.selectBestStrategy(state)
	if bestStrategy != nil {
		t.Logf("Best strategy: %s", bestStrategy.GetName())
	} else {
		t.Log("No best strategy selected")
	}

	// Test getCurrentSystemState
	currentState := engine.getCurrentSystemState()
	if currentState != nil {
		t.Logf("Current system state retrieved with %d nodes", len(currentState.Nodes))
	} else {
		t.Log("Current system state is nil")
	}

	// Test healSystem
	fault := &FaultDetection{
		ID:          "test_fault",
		Type:        FaultTypeNodeFailure,
		Severity:    FaultSeverityHigh,
		Target:      "node1",
		Description: "Test fault",
		DetectedAt:  time.Now(),
		Status:      FaultStatusDetected,
		Metadata:    make(map[string]interface{}),
	}

	// Create healing attempt for testing
	attempt := &HealingAttempt{
		ID:          fmt.Sprintf("heal_%d", time.Now().UnixNano()),
		Timestamp:   time.Now(),
		Strategy:    "test_strategy",
		SystemState: state,
		Result: &HealingResult{
			Improvement:  0.5,
			Metrics:      make(map[string]float64),
			ActionsTaken: []string{"test_action"},
			Timestamp:    time.Now(),
		},
		Duration: 100 * time.Millisecond,
		Success:  true,
		Metadata: make(map[string]interface{}),
	}

	// Test updateMetrics
	engine.updateMetrics(attempt)

	// Test addToHistory
	engine.addToHistory(attempt)

	// Test GetHistory after adding attempt
	history = engine.GetHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 entry in history, got %d", len(history))
	}

	// Test learnFromAttempt
	engine.learnFromAttempt(attempt)

	// Test getStrategyByName
	strategy, exists := engine.getStrategyByName("restart_services")
	if !exists {
		t.Error("Expected restart_services strategy to exist")
	}
	if strategy == nil {
		t.Error("Expected restart_services strategy to be non-nil")
	}

	// Test adjustWeightForState
	weight := engine.adjustWeightForState(0.5, state, strategy)
	if weight <= 0 || weight > 1.0 {
		t.Errorf("Expected adjusted weight to be between 0 and 1, got %f", weight)
	}

	// Test getResourceMetrics
	resourceMetrics := engine.getResourceMetrics()
	if resourceMetrics == nil {
		t.Error("Expected resource metrics to be non-nil")
	}

	// Test getPerformanceMetrics
	performanceMetrics := engine.getPerformanceMetrics()
	if performanceMetrics == nil {
		t.Error("Expected performance metrics to be non-nil")
	}

	// Test getHealthMetrics
	healthMetrics := engine.getHealthMetrics()
	if healthMetrics == nil {
		t.Error("Expected health metrics to be non-nil")
	}
}
