package fault_tolerance

import (
	"context"
	"testing"
	"time"
)

// TestSelfHealingRecoveryAdapter_Integration validates the adapter integrates with RecoveryEngine
func TestSelfHealingRecoveryAdapter_Integration(t *testing.T) {
	// Create base fault tolerance manager
	base := NewFaultToleranceManager(&Config{HealthCheckInterval: time.Second})

	// Create enhanced config with self-healing enabled
	cfg := NewEnhancedFaultToleranceConfig(&Config{HealthCheckInterval: time.Second})
	cfg.EnableSelfHealing = true

	// Create enhanced fault tolerance manager
	eftm := NewEnhancedFaultToleranceManager(cfg, base)

	// Verify adapter was registered in recovery engine
	if eftm.FaultToleranceManager == nil || eftm.FaultToleranceManager.recoveryEngine == nil {
		t.Fatal("recovery engine not initialized")
	}

	re := eftm.FaultToleranceManager.recoveryEngine

	// Check that self-healing adapter was added to performance anomaly strategies
	perfStrategies := re.strategies[FaultTypePerformanceAnomaly]
	found := false
	for _, strategy := range perfStrategies {
		if strategy.GetName() == "self_healing" {
			found = true
			break
		}
	}
	if !found {
		t.Error("self-healing adapter not found in performance anomaly strategies")
	}

	// Test adapter can handle faults
	fault := &FaultDetection{
		ID:          "test-fault",
		Type:        FaultTypePerformanceAnomaly,
		Severity:    FaultSeverityMedium,
		Target:      "test-node",
		Description: "Test performance issue",
		DetectedAt:  time.Now(),
		Status:      FaultStatusDetected,
		Metadata:    make(map[string]interface{}),
	}

	// Find the adapter and test it can handle the fault
	var adapter *SelfHealingRecoveryAdapter
	for _, strategy := range perfStrategies {
		if strategy.GetName() == "self_healing" {
			if a, ok := strategy.(*SelfHealingRecoveryAdapter); ok {
				adapter = a
				break
			}
		}
	}

	if adapter == nil {
		t.Fatal("self-healing adapter not found")
	}

	if !adapter.CanHandle(fault) {
		t.Error("adapter should be able to handle performance anomaly")
	}

	// Test recovery (should not panic)
	ctx := context.Background()
	result, err := adapter.Recover(ctx, fault)
	if err != nil {
		t.Errorf("recovery failed: %v", err)
	}
	if result == nil {
		t.Error("recovery result should not be nil")
	}
	if result.Strategy != "self_healing" {
		t.Errorf("expected strategy 'self_healing', got '%s'", result.Strategy)
	}
}

// TestSelfHealingStrategies_BasicFunctionality validates basic healing strategies work
func TestSelfHealingStrategies_BasicFunctionality(t *testing.T) {
	// Test service restart strategy
	restartStrategy := NewServiceRestartStrategy(nil)
	if restartStrategy.Name() != "service_restart" {
		t.Errorf("expected name 'service_restart', got '%s'", restartStrategy.Name())
	}

	fault := &FaultDetection{
		Type:   FaultTypeServiceUnavailable,
		Target: "test-service",
	}

	systemState := &SystemState{
		OverallHealth:   0.8,
		ComponentHealth: map[string]float64{"node1": 0.8},
		ResourceUsage:   map[string]float64{"cpu": 0.5},
		Performance:     map[string]float64{"latency": 100.0},
		ActiveFaults:    []*FaultDetection{fault},
		Metadata:        make(map[string]interface{}),
	}

	if !restartStrategy.CanHeal(fault, systemState) {
		t.Error("service restart should be able to heal service unavailable fault")
	}

	ctx := context.Background()
	result, err := restartStrategy.Heal(ctx, fault, systemState)
	if err != nil {
		t.Errorf("healing failed: %v", err)
	}
	if result == nil || !result.Success {
		t.Error("healing should succeed")
	}

	// Test resource reallocation strategy
	reallocStrategy := NewResourceReallocationStrategy(nil)
	if reallocStrategy.Name() != "resource_reallocation" {
		t.Errorf("expected name 'resource_reallocation', got '%s'", reallocStrategy.Name())
	}

	resourceFault := &FaultDetection{
		Type:   FaultTypeResourceExhaustion,
		Target: "test-node",
	}

	if !reallocStrategy.CanHeal(resourceFault, systemState) {
		t.Error("resource reallocation should be able to heal resource exhaustion fault")
	}

	result, err = reallocStrategy.Heal(ctx, resourceFault, systemState)
	if err != nil {
		t.Errorf("resource healing failed: %v", err)
	}
	if result == nil || !result.Success {
		t.Error("resource healing should succeed")
	}

	// Test load redistribution strategy
	loadStrategy := NewLoadRedistributionStrategy(nil)
	if loadStrategy.Name() != "load_redistribution" {
		t.Errorf("expected name 'load_redistribution', got '%s'", loadStrategy.Name())
	}

	// Load redistribution handles performance/resource issues, not service unavailable
	perfFault := &FaultDetection{
		Type:   FaultTypePerformanceAnomaly,
		Target: "test-node",
	}

	if !loadStrategy.CanHeal(perfFault, systemState) {
		t.Error("load redistribution should be able to heal performance anomaly fault")
	}

	result, err = loadStrategy.Heal(ctx, perfFault, systemState)
	if err != nil {
		t.Errorf("load redistribution healing failed: %v", err)
	}
	if result == nil || !result.Success {
		t.Error("load redistribution healing should succeed")
	}
}
