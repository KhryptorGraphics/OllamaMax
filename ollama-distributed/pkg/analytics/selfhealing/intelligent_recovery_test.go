package selfhealing

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestIntelligentRecoveryEngine(t *testing.T) {
	// Create recovery configuration
	config := &RecoveryConfig{
		MaxConcurrentRecoveries: 3,
		RecoveryTimeout:         time.Minute * 10,
		RollbackTimeout:         time.Minute * 5,
		MaxRetries:              3,
		RetryDelay:              time.Second * 30,
		EnableLearning:          true,
		EnableRollback:          true,
		SuccessThreshold:        0.8,
	}

	// Create recovery engine
	engine, err := NewIntelligentRecoveryEngine(config)
	if err != nil {
		t.Fatalf("Failed to create recovery engine: %v", err)
	}
	defer engine.Stop()

	// Test recovery from incident
	t.Run("RecoveryFromIncident", func(t *testing.T) {
		testRecoveryFromIncident(t, engine)
	})

	// Test strategy selection
	t.Run("StrategySelection", func(t *testing.T) {
		testStrategySelection(t, engine)
	})

	// Test recovery execution
	t.Run("RecoveryExecution", func(t *testing.T) {
		testRecoveryExecution(t, engine)
	})

	// Test rollback functionality
	t.Run("RollbackFunctionality", func(t *testing.T) {
		testRollbackFunctionality(t, engine)
	})

	// Test recovery strategies
	t.Run("RecoveryStrategies", func(t *testing.T) {
		testRecoveryStrategies(t, engine)
	})
}

func testRecoveryFromIncident(t *testing.T, engine *IntelligentRecoveryEngine) {
	// Create test incident
	incident := &SystemIncident{
		ID:          "incident-recovery-1",
		Type:        "service_failure",
		Description: "Service degradation detected",
		Severity:    "high",
		NodeID:      "node-1",
		Symptoms:    []string{"high_error_rate", "slow_response"},
		Metrics: map[string]float64{
			"cpu_utilization": 0.8,
			"error_rate":      0.15,
			"response_time":   3.0,
		},
		StartTime:  time.Now().Add(-time.Minute * 5),
		DetectedAt: time.Now(),
	}

	// Create diagnosis result
	diagnosis := &DiagnosticResult{
		IncidentID:         incident.ID,
		RootCause:          "service_degradation",
		Confidence:         0.9,
		Evidence:           []string{"High error rate detected", "Response time exceeded threshold"},
		RecommendedActions: []string{"Restart service", "Check configuration"},
	}

	// Perform recovery
	ctx := context.Background()
	operation, err := engine.RecoverFromIncident(ctx, incident, diagnosis)
	if err != nil {
		t.Fatalf("Recovery failed: %v", err)
	}

	// Verify recovery operation
	if operation.IncidentID != incident.ID {
		t.Errorf("Expected incident ID %s, got %s", incident.ID, operation.IncidentID)
	}

	if operation.Status != "scheduled" {
		t.Errorf("Expected status 'scheduled', got %s", operation.Status)
	}

	if operation.Progress != 0.0 {
		t.Errorf("Expected initial progress 0.0, got %f", operation.Progress)
	}

	// Wait for recovery to start
	time.Sleep(time.Millisecond * 100)

	// Check that recovery is in progress or completed
	if operation.Status != "in_progress" && operation.Status != "completed" {
		t.Errorf("Expected recovery to be in progress or completed, got %s", operation.Status)
	}

	// Test duplicate recovery prevention
	operation2, err := engine.RecoverFromIncident(ctx, incident, diagnosis)
	if err != nil {
		t.Fatalf("Duplicate recovery check failed: %v", err)
	}

	if operation2.ID != operation.ID {
		t.Error("Expected same operation for duplicate recovery request")
	}
}

func testStrategySelection(t *testing.T, engine *IntelligentRecoveryEngine) {
	// Test service restart strategy selection
	incident := &SystemIncident{
		ID:       "incident-strategy-1",
		Severity: "high",
	}

	diagnosis := &DiagnosticResult{
		RootCause:  "service_degradation",
		Confidence: 0.9,
	}

	strategy, err := engine.selectBestStrategy(incident, diagnosis)
	if err != nil {
		t.Fatalf("Strategy selection failed: %v", err)
	}

	if strategy.GetName() != "service_restart" {
		t.Errorf("Expected service_restart strategy, got %s", strategy.GetName())
	}

	// Test resource scaling strategy selection
	diagnosis.RootCause = "cpu_exhaustion"
	strategy, err = engine.selectBestStrategy(incident, diagnosis)
	if err != nil {
		t.Fatalf("Strategy selection failed: %v", err)
	}

	if strategy.GetName() != "resource_scaling" {
		t.Errorf("Expected resource_scaling strategy, got %s", strategy.GetName())
	}

	// Test no applicable strategy
	diagnosis.RootCause = "unknown_issue"
	_, err = engine.selectBestStrategy(incident, diagnosis)
	if err == nil {
		t.Error("Expected error for unknown issue")
	}
}

func testRecoveryExecution(t *testing.T, engine *IntelligentRecoveryEngine) {
	// Create recovery plan
	plan := &RecoveryPlan{
		ID:                 "plan-test-1",
		StrategyName:       "service_restart",
		IncidentID:         "incident-test-1",
		EstimatedDuration:  time.Minute * 2,
		SuccessProbability: 0.9,
		RiskLevel:          "low",
		Steps: []*RecoveryStep{
			{
				ID:          "step-1",
				Name:        "Test Step",
				Description: "Test recovery step",
				Action:      "test_action",
				Timeout:     time.Second * 30,
				Critical:    true,
				Order:       1,
			},
		},
		CreatedAt: time.Now(),
	}

	// Create recovery operation
	operation := &RecoveryOperation{
		ID:          "op-test-1",
		PlanID:      plan.ID,
		IncidentID:  "incident-test-1",
		Status:      "scheduled",
		StartTime:   time.Now(),
		Progress:    0.0,
		StepResults: make([]*StepResult, 0),
	}

	// Execute recovery
	ctx := context.Background()
	engine.executeRecovery(ctx, operation, plan)

	// Verify execution results
	if operation.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", operation.Status)
	}

	if operation.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", operation.Progress)
	}

	if len(operation.StepResults) != len(plan.Steps) {
		t.Errorf("Expected %d step results, got %d", len(plan.Steps), len(operation.StepResults))
	}

	// Verify step result
	if len(operation.StepResults) > 0 {
		result := operation.StepResults[0]
		if !result.Success {
			t.Error("Expected step to succeed")
		}
		if result.Duration <= 0 {
			t.Error("Expected positive step duration")
		}
	}
}

func testRollbackFunctionality(t *testing.T, engine *IntelligentRecoveryEngine) {
	// Create operation that should trigger rollback
	operation := &RecoveryOperation{
		ID:     "op-rollback-test",
		Status: "failed",
		StepResults: []*StepResult{
			{Success: false, ErrorMessage: "Step failed"},
			{Success: false, ErrorMessage: "Step failed"},
		},
	}

	// Test rollback decision
	shouldRollback := engine.shouldRollback(operation)
	if !shouldRollback {
		t.Error("Expected rollback to be triggered for failed operation")
	}

	// Test operation with low failure rate
	operation.StepResults = []*StepResult{
		{Success: true},
		{Success: true},
		{Success: false, ErrorMessage: "Step failed"},
	}

	shouldRollback = engine.shouldRollback(operation)
	if shouldRollback {
		t.Error("Expected no rollback for low failure rate")
	}
}

func testRecoveryStrategies(t *testing.T, engine *IntelligentRecoveryEngine) {
	// Test service restart strategy
	t.Run("ServiceRestartStrategy", func(t *testing.T) {
		strategy := &ServiceRestartStrategy{}

		incident := &SystemIncident{ID: "test-incident"}
		diagnosis := &DiagnosticResult{RootCause: "service_degradation", Confidence: 0.9}

		if !strategy.CanRecover(incident, diagnosis) {
			t.Error("Service restart strategy should handle service degradation")
		}

		if strategy.GetName() != "service_restart" {
			t.Errorf("Expected name 'service_restart', got %s", strategy.GetName())
		}

		if strategy.GetPriority() <= 0 {
			t.Error("Expected positive priority")
		}

		duration := strategy.EstimateRecoveryTime(incident, diagnosis)
		if duration <= 0 {
			t.Error("Expected positive recovery time")
		}

		probability := strategy.EstimateSuccessProbability(incident, diagnosis)
		if probability < 0 || probability > 1 {
			t.Errorf("Invalid success probability: %f", probability)
		}

		plan, err := strategy.CreateRecoveryPlan(incident, diagnosis)
		if err != nil {
			t.Fatalf("Failed to create recovery plan: %v", err)
		}

		if len(plan.Steps) == 0 {
			t.Error("Expected recovery steps")
		}
	})

	// Test resource scaling strategy
	t.Run("ResourceScalingStrategy", func(t *testing.T) {
		strategy := &ResourceScalingStrategy{}

		incident := &SystemIncident{ID: "test-incident"}
		diagnosis := &DiagnosticResult{RootCause: "cpu_exhaustion", Confidence: 0.9}

		if !strategy.CanRecover(incident, diagnosis) {
			t.Error("Resource scaling strategy should handle CPU exhaustion")
		}

		plan, err := strategy.CreateRecoveryPlan(incident, diagnosis)
		if err != nil {
			t.Fatalf("Failed to create scaling plan: %v", err)
		}

		if plan.RiskLevel != "medium" {
			t.Errorf("Expected medium risk level, got %s", plan.RiskLevel)
		}
	})

	// Test cache clearing strategy
	t.Run("CacheClearingStrategy", func(t *testing.T) {
		strategy := &CacheClearingStrategy{}

		incident := &SystemIncident{ID: "test-incident"}
		diagnosis := &DiagnosticResult{RootCause: "memory_exhaustion", Confidence: 0.8}

		if !strategy.CanRecover(incident, diagnosis) {
			t.Error("Cache clearing strategy should handle memory exhaustion")
		}

		plan, err := strategy.CreateRecoveryPlan(incident, diagnosis)
		if err != nil {
			t.Fatalf("Failed to create cache clearing plan: %v", err)
		}

		if plan.RiskLevel != "low" {
			t.Errorf("Expected low risk level, got %s", plan.RiskLevel)
		}
	})

	// Test network recovery strategy
	t.Run("NetworkRecoveryStrategy", func(t *testing.T) {
		strategy := &NetworkRecoveryStrategy{}

		incident := &SystemIncident{ID: "test-incident"}
		diagnosis := &DiagnosticResult{RootCause: "network_issues", Confidence: 0.7}

		if !strategy.CanRecover(incident, diagnosis) {
			t.Error("Network recovery strategy should handle network issues")
		}

		plan, err := strategy.CreateRecoveryPlan(incident, diagnosis)
		if err != nil {
			t.Fatalf("Failed to create network recovery plan: %v", err)
		}

		if plan.RiskLevel != "high" {
			t.Errorf("Expected high risk level, got %s", plan.RiskLevel)
		}
	})
}

func TestDefaultActionExecutor(t *testing.T) {
	executor := &DefaultActionExecutor{action: "test_action"}

	// Test execution
	step := &RecoveryStep{
		ID:      "test-step",
		Action:  "test_action",
		Timeout: time.Second * 30,
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, step)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Error("Expected successful execution")
	}

	if result.StepID != step.ID {
		t.Errorf("Expected step ID %s, got %s", step.ID, result.StepID)
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Test capabilities
	if !executor.CanExecute("any_action") {
		t.Error("Default executor should handle any action")
	}

	if executor.GetName() != "default_executor" {
		t.Errorf("Expected name 'default_executor', got %s", executor.GetName())
	}

	if executor.GetTimeout() <= 0 {
		t.Error("Expected positive timeout")
	}
}

func TestRecoveryResult(t *testing.T) {
	// Create test operation
	operation := &RecoveryOperation{
		ID:         "test-op",
		IncidentID: "test-incident",
		StartTime:  time.Now().Add(-time.Minute),
		StepResults: []*StepResult{
			{Success: true},
			{Success: true},
			{Success: false},
		},
		Status: "completed",
	}

	plan := &RecoveryPlan{
		StrategyName: "test_strategy",
	}

	// Create recovery engine to test result recording
	config := &RecoveryConfig{
		EnableLearning: true,
	}

	engine, err := NewIntelligentRecoveryEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Stop()

	// Record result
	engine.recordRecoveryResult(operation, plan)

	// Verify result was recorded
	if len(engine.recoveryHistory) != 1 {
		t.Errorf("Expected 1 recovery result, got %d", len(engine.recoveryHistory))
	}

	result := engine.recoveryHistory[0]
	if result.OperationID != operation.ID {
		t.Errorf("Expected operation ID %s, got %s", operation.ID, result.OperationID)
	}

	if result.StepsCompleted != 2 {
		t.Errorf("Expected 2 completed steps, got %d", result.StepsCompleted)
	}

	if result.StepsFailed != 1 {
		t.Errorf("Expected 1 failed step, got %d", result.StepsFailed)
	}

	expectedRecoveryRate := 2.0 / 3.0
	if result.RecoveryRate != expectedRecoveryRate {
		t.Errorf("Expected recovery rate %f, got %f", expectedRecoveryRate, result.RecoveryRate)
	}
}

func BenchmarkRecovery(b *testing.B) {
	config := &RecoveryConfig{
		MaxConcurrentRecoveries: 10,
		RecoveryTimeout:         time.Minute,
		EnableLearning:          false, // Disable for benchmark
		EnableRollback:          false, // Disable for benchmark
	}

	engine, err := NewIntelligentRecoveryEngine(config)
	if err != nil {
		b.Fatalf("Failed to create recovery engine: %v", err)
	}
	defer engine.Stop()

	incident := &SystemIncident{
		ID:       "bench-incident",
		Severity: "high",
	}

	diagnosis := &DiagnosticResult{
		RootCause:  "service_degradation",
		Confidence: 0.9,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create unique incident ID for each iteration
		incident.ID = fmt.Sprintf("bench-incident-%d", i)

		_, err := engine.RecoverFromIncident(context.Background(), incident, diagnosis)
		if err != nil {
			b.Fatalf("Recovery failed: %v", err)
		}
	}
}
