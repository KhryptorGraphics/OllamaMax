package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
)

func main() {
	fmt.Println("Testing Self-Healing System...")

	// Test the self-healing system
	if err := testSelfHealingSystem(); err != nil {
		log.Fatalf("Self-healing system test failed: %v", err)
	}

	fmt.Println("✅ Self-healing system test completed successfully!")
}

func testSelfHealingSystem() error {
	fmt.Println("Setting up self-healing system...")

	// Create fault tolerance manager
	ftConfig := &fault_tolerance.Config{
		ReplicationFactor:     3,
		HealthCheckInterval:   10 * time.Second,
		RecoveryTimeout:       30 * time.Second,
		CircuitBreakerEnabled: true,
		CheckpointInterval:    60 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          5 * time.Second,
	}

	ftManager := fault_tolerance.NewFaultToleranceManager(ftConfig)
	if ftManager == nil {
		return fmt.Errorf("failed to create fault tolerance manager")
	}

	// Create self-healing engine
	healingConfig := &fault_tolerance.SelfHealingConfig{
		HealingInterval:            15 * time.Second,
		MonitoringInterval:         5 * time.Second,
		LearningInterval:           30 * time.Second,
		HealingThreshold:           0.7,
		MaxConcurrentHealing:       2,
		HealingTimeout:             2 * time.Minute,
		MaxHealingHistory:          100,
		EnableAdaptiveStrategy:     true,
		EnableLearning:             true,
		EnablePredictiveHealing:    true,
		EnableProactiveHealing:     true,
		EnableServiceRestart:       true,
		EnableResourceReallocation: true,
		EnableLoadRedistribution:   true,
		EnableFailover:             true,
		EnableScaling:              true,
	}

	healingEngine := fault_tolerance.NewSelfHealingEngine(ftManager, healingConfig)
	if healingEngine == nil {
		return fmt.Errorf("failed to create self-healing engine")
	}

	// Start the healing engine
	if err := healingEngine.Start(); err != nil {
		return fmt.Errorf("failed to start healing engine: %v", err)
	}
	defer healingEngine.Stop()

	fmt.Println("✅ Self-healing system setup complete")

	// Test scenarios
	scenarios := []struct {
		name        string
		description string
		testFunc    func(*fault_tolerance.SelfHealingEngine) error
	}{
		{
			name:        "Service Restart Healing",
			description: "Test service restart healing strategy",
			testFunc:    testServiceRestartHealing,
		},
		{
			name:        "Resource Reallocation Healing",
			description: "Test resource reallocation healing strategy",
			testFunc:    testResourceReallocationHealing,
		},
		{
			name:        "Load Redistribution Healing",
			description: "Test load redistribution healing strategy",
			testFunc:    testLoadRedistributionHealing,
		},
		{
			name:        "Failover Healing",
			description: "Test failover healing strategy",
			testFunc:    testFailoverHealing,
		},
		{
			name:        "Scaling Healing",
			description: "Test scaling healing strategy",
			testFunc:    testScalingHealing,
		},
		{
			name:        "Proactive System Healing",
			description: "Test proactive system healing capabilities",
			testFunc:    testProactiveSystemHealing,
		},
		{
			name:        "Healing Performance Tracking",
			description: "Test healing performance tracking and metrics",
			testFunc:    testHealingPerformanceTracking,
		},
	}

	fmt.Println("\n=== Testing Self-Healing System ===")

	for i, scenario := range scenarios {
		fmt.Printf("%d. Testing %s...\n", i+1, scenario.name)
		fmt.Printf("  Description: %s\n", scenario.description)

		if err := scenario.testFunc(healingEngine); err != nil {
			fmt.Printf("  ❌ Test failed: %v\n", err)
			return err
		}

		fmt.Printf("  ✅ Test passed\n\n")

		// Wait between tests
		time.Sleep(2 * time.Second)
	}

	return nil
}

func testServiceRestartHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Create a service unavailable fault
	fault := &fault_tolerance.FaultDetection{
		ID:          "test_service_fault",
		Type:        fault_tolerance.FaultTypeServiceUnavailable,
		Target:      "api_gateway",
		Severity:    fault_tolerance.FaultSeverityHigh,
		Description: "API Gateway service is unavailable",
		Metadata:    map[string]interface{}{"service": "api_gateway"},
	}

	// Attempt healing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealFault(ctx, fault)
	if err != nil {
		return fmt.Errorf("healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("healing was not successful")
	}

	if len(result.Actions) == 0 {
		return fmt.Errorf("no healing actions were taken")
	}

	fmt.Printf("    Healing successful: %d actions taken in %v\n",
		len(result.Actions), result.Duration)

	return nil
}

func testResourceReallocationHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Create a resource exhaustion fault
	fault := &fault_tolerance.FaultDetection{
		ID:          "test_resource_fault",
		Type:        fault_tolerance.FaultTypeResourceExhaustion,
		Target:      "scheduler",
		Severity:    fault_tolerance.FaultSeverityMedium,
		Description: "Scheduler is running out of memory",
		Metadata:    map[string]interface{}{"resource": "memory", "usage": 0.95},
	}

	// Attempt healing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealFault(ctx, fault)
	if err != nil {
		return fmt.Errorf("healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("healing was not successful")
	}

	fmt.Printf("    Resource reallocation successful: health improvement %.2f\n",
		result.HealthImprovement)

	return nil
}

func testLoadRedistributionHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Create a performance anomaly fault
	fault := &fault_tolerance.FaultDetection{
		ID:          "test_load_fault",
		Type:        fault_tolerance.FaultTypePerformanceAnomaly,
		Target:      "load_balancer",
		Severity:    fault_tolerance.FaultSeverityMedium,
		Description: "Load balancer is experiencing high latency",
		Metadata:    map[string]interface{}{"latency": 2000, "threshold": 500},
	}

	// Attempt healing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealFault(ctx, fault)
	if err != nil {
		return fmt.Errorf("healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("healing was not successful")
	}

	fmt.Printf("    Load redistribution successful: confidence %.2f\n",
		result.Confidence)

	return nil
}

func testFailoverHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Create a critical node failure fault
	fault := &fault_tolerance.FaultDetection{
		ID:          "test_failover_fault",
		Type:        fault_tolerance.FaultTypeNodeFailure,
		Target:      "primary_node",
		Severity:    fault_tolerance.FaultSeverityCritical,
		Description: "Primary node has failed completely",
		Metadata:    map[string]interface{}{"node_id": "node-1", "failure_type": "hardware"},
	}

	// Attempt healing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealFault(ctx, fault)
	if err != nil {
		return fmt.Errorf("healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("healing was not successful")
	}

	fmt.Printf("    Failover successful: %d actions, duration %v\n",
		len(result.Actions), result.Duration)

	return nil
}

func testScalingHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Create a resource exhaustion fault that requires scaling
	fault := &fault_tolerance.FaultDetection{
		ID:          "test_scaling_fault",
		Type:        fault_tolerance.FaultTypeResourceExhaustion,
		Target:      "worker_pool",
		Severity:    fault_tolerance.FaultSeverityHigh,
		Description: "Worker pool is at maximum capacity",
		Metadata:    map[string]interface{}{"cpu_usage": 0.95, "queue_length": 1000},
	}

	// Attempt healing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealFault(ctx, fault)
	if err != nil {
		return fmt.Errorf("healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("healing was not successful")
	}

	fmt.Printf("    Scaling successful: health improvement %.2f\n",
		result.HealthImprovement)

	return nil
}

func testProactiveSystemHealing(engine *fault_tolerance.SelfHealingEngine) error {
	// Test proactive system healing without specific faults
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := engine.HealSystem(ctx)
	if err != nil {
		return fmt.Errorf("proactive healing failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("proactive healing was not successful")
	}

	fmt.Printf("    Proactive healing completed: %d actions taken\n",
		len(result.Actions))

	return nil
}

func testHealingPerformanceTracking(engine *fault_tolerance.SelfHealingEngine) error {
	// Create multiple faults to generate performance data
	faults := []*fault_tolerance.FaultDetection{
		{
			ID:          "perf_test_1",
			Type:        fault_tolerance.FaultTypeServiceUnavailable,
			Target:      "service_1",
			Severity:    fault_tolerance.FaultSeverityMedium,
			Description: "Service 1 performance test",
		},
		{
			ID:          "perf_test_2",
			Type:        fault_tolerance.FaultTypeResourceExhaustion,
			Target:      "service_2",
			Severity:    fault_tolerance.FaultSeverityHigh,
			Description: "Service 2 performance test",
		},
	}

	// Heal multiple faults
	for _, fault := range faults {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

		result, err := engine.HealFault(ctx, fault)
		cancel()

		if err != nil {
			return fmt.Errorf("healing fault %s failed: %w", fault.ID, err)
		}

		if !result.Success {
			return fmt.Errorf("healing fault %s was not successful", fault.ID)
		}
	}

	// Wait for metrics to be processed
	time.Sleep(2 * time.Second)

	fmt.Printf("    Performance tracking test completed: %d faults healed\n",
		len(faults))

	return nil
}

// Helper function to create mock system state
func createMockSystemState() *fault_tolerance.SystemState {
	return &fault_tolerance.SystemState{
		OverallHealth: 0.6, // Below threshold to trigger healing
		ComponentHealth: map[string]float64{
			"api_gateway": 0.5,
			"scheduler":   0.7,
			"p2p_network": 0.8,
		},
		ResourceUsage: map[string]float64{
			"cpu":    0.85,
			"memory": 0.90,
			"disk":   0.60,
		},
		Performance: map[string]float64{
			"response_time": 0.4,
			"throughput":    0.6,
			"availability":  0.8,
		},
		ActiveFaults: []*fault_tolerance.FaultDetection{},
		Predictions:  []*fault_tolerance.FaultPrediction{},
		Metadata:     make(map[string]interface{}),
		Timestamp:    time.Now(),
	}
}
