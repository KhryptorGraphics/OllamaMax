package chaos

import (
	"context"
	"testing"
	"time"
)

func TestChaosEngineeringFramework(t *testing.T) {
	// Create chaos configuration
	config := &ChaosConfig{
		EnableContinuousTesting:  false, // Disable for testing
		ExperimentInterval:       time.Hour,
		MaxConcurrentExperiments: 3,
		SafetyThreshold:          0.1,
		AutoRollbackEnabled:      true,
		RollbackTimeout:          time.Minute * 5,
		MetricsRetention:         time.Hour * 24,
		ReportingEnabled:         false, // Disable for testing
		IntegrationEnabled:       true,
	}

	// Create chaos framework
	framework, err := NewChaosEngineeringFramework(config)
	if err != nil {
		t.Fatalf("Failed to create chaos framework: %v", err)
	}
	defer framework.Stop()

	// Test experiment creation from template
	t.Run("ExperimentCreation", func(t *testing.T) {
		testExperimentCreation(t, framework)
	})

	// Test experiment execution
	t.Run("ExperimentExecution", func(t *testing.T) {
		testExperimentExecution(t, framework)
	})

	// Test safety monitoring
	t.Run("SafetyMonitoring", func(t *testing.T) {
		testSafetyMonitoring(t, framework)
	})

	// Test resilience scoring
	t.Run("ResilienceScoring", func(t *testing.T) {
		testResilienceScoring(t, framework)
	})

	// Test experiment templates
	t.Run("ExperimentTemplates", func(t *testing.T) {
		testExperimentTemplates(t, framework)
	})
}

func testExperimentCreation(t *testing.T, framework *ChaosEngineeringFramework) {
	// Test creating experiment from template
	experiment, err := framework.CreateExperimentFromTemplate(
		"network_latency",
		"test-service",
		[]string{"node-1", "node-2"},
	)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	// Verify experiment properties
	if experiment.ID == "" {
		t.Error("Expected experiment ID to be set")
	}

	if experiment.Name != "Network Latency" {
		t.Errorf("Expected name 'Network Latency', got %s", experiment.Name)
	}

	if experiment.TargetService != "test-service" {
		t.Errorf("Expected target service 'test-service', got %s", experiment.TargetService)
	}

	if len(experiment.TargetNodes) != 2 {
		t.Errorf("Expected 2 target nodes, got %d", len(experiment.TargetNodes))
	}

	if len(experiment.FailureScenarios) == 0 {
		t.Error("Expected failure scenarios to be set")
	}

	if experiment.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	if experiment.SafetyLimits == nil {
		t.Error("Expected safety limits to be set")
	}

	if len(experiment.SuccessCriteria) == 0 {
		t.Error("Expected success criteria to be set")
	}

	// Test invalid template
	_, err = framework.CreateExperimentFromTemplate("invalid_template", "service", []string{"node"})
	if err == nil {
		t.Error("Expected error for invalid template")
	}
}

func testExperimentExecution(t *testing.T, framework *ChaosEngineeringFramework) {
	// Create test experiment
	experiment, err := framework.CreateExperimentFromTemplate(
		"service_restart",
		"test-service",
		[]string{"node-1"},
	)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	// Set short duration for testing
	experiment.Duration = time.Second * 2

	// Run experiment
	ctx := context.Background()
	result, err := framework.RunExperiment(ctx, experiment)
	if err != nil {
		t.Fatalf("Failed to run experiment: %v", err)
	}

	// Verify initial result
	if result.ExperimentID != experiment.ID {
		t.Errorf("Expected experiment ID %s, got %s", experiment.ID, result.ExperimentID)
	}

	if result.Status != "started" {
		t.Errorf("Expected status 'started', got %s", result.Status)
	}

	// Wait for experiment to complete
	time.Sleep(time.Second * 3)

	// Check experiment history
	history := framework.GetExperimentHistory()
	if len(history) == 0 {
		t.Error("Expected experiment to be recorded in history")
	}

	// Verify completed experiment
	completedResult := history[len(history)-1]
	if completedResult.ExperimentID != experiment.ID {
		t.Error("Expected experiment in history")
	}

	if completedResult.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", completedResult.Status)
	}

	if completedResult.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Test concurrent experiment limit
	for i := 0; i < framework.config.MaxConcurrentExperiments+1; i++ {
		exp, _ := framework.CreateExperimentFromTemplate("service_restart", "test-service", []string{"node-1"})
		exp.Duration = time.Second * 10 // Longer duration
		exp.ID = exp.ID + string(rune(i)) // Make unique
		
		_, err := framework.RunExperiment(ctx, exp)
		if i >= framework.config.MaxConcurrentExperiments && err == nil {
			t.Error("Expected error when exceeding concurrent experiment limit")
		}
	}
}

func testSafetyMonitoring(t *testing.T, framework *ChaosEngineeringFramework) {
	// Create experiment with strict safety limits
	experiment, err := framework.CreateExperimentFromTemplate(
		"memory_pressure",
		"test-service",
		[]string{"node-1"},
	)
	if err != nil {
		t.Fatalf("Failed to create experiment: %v", err)
	}

	// Set very strict safety limits for testing
	experiment.SafetyLimits.MaxErrorRate = 0.001 // Very low threshold
	experiment.SafetyLimits.MonitoringWindow = time.Millisecond * 100

	// Test safety violation check
	violation := framework.checkSafetyViolation(experiment)
	// Should potentially trigger violation due to random metrics
	t.Logf("Safety violation detected: %v", violation)

	// Test metrics collection
	metrics := framework.getCurrentMetrics("test-service")
	if metrics == nil {
		t.Error("Expected metrics to be returned")
	}

	if metrics.AvailabilityScore < 0 || metrics.AvailabilityScore > 1 {
		t.Errorf("Invalid availability score: %f", metrics.AvailabilityScore)
	}

	if metrics.ErrorRate < 0 || metrics.ErrorRate > 1 {
		t.Errorf("Invalid error rate: %f", metrics.ErrorRate)
	}

	if metrics.ResilienceScore < 0 || metrics.ResilienceScore > 1 {
		t.Errorf("Invalid resilience score: %f", metrics.ResilienceScore)
	}
}

func testResilienceScoring(t *testing.T, framework *ChaosEngineeringFramework) {
	// Test resilience score calculation
	score, err := framework.GetResilienceScore()
	if err != nil {
		t.Fatalf("Failed to get resilience score: %v", err)
	}

	if score < 0 || score > 1 {
		t.Errorf("Invalid resilience score: %f", score)
	}

	// Test validator directly
	validator := framework.resilienceValidator
	directScore, err := validator.CalculateResilienceScore()
	if err != nil {
		t.Fatalf("Failed to calculate resilience score: %v", err)
	}

	if directScore != score {
		t.Errorf("Expected same score from both methods, got %f vs %f", score, directScore)
	}
}

func testExperimentTemplates(t *testing.T, framework *ChaosEngineeringFramework) {
	// Test all default templates
	templates := []string{"node_failure", "network_latency", "memory_pressure", "service_restart"}

	for _, templateID := range templates {
		t.Run(templateID, func(t *testing.T) {
			experiment, err := framework.CreateExperimentFromTemplate(
				templateID,
				"test-service",
				[]string{"node-1"},
			)
			if err != nil {
				t.Fatalf("Failed to create experiment from template %s: %v", templateID, err)
			}

			// Validate experiment
			err = framework.validateExperiment(experiment)
			if err != nil {
				t.Errorf("Template %s created invalid experiment: %v", templateID, err)
			}

			// Check template-specific properties
			switch templateID {
			case "node_failure":
				if experiment.Type != "infrastructure" {
					t.Errorf("Expected type 'infrastructure', got %s", experiment.Type)
				}
			case "network_latency":
				if experiment.Type != "network" {
					t.Errorf("Expected type 'network', got %s", experiment.Type)
				}
			case "memory_pressure":
				if experiment.Type != "resource" {
					t.Errorf("Expected type 'resource', got %s", experiment.Type)
				}
			case "service_restart":
				if experiment.Type != "service" {
					t.Errorf("Expected type 'service', got %s", experiment.Type)
				}
			}
		})
	}
}

func TestExperimentValidation(t *testing.T) {
	config := &ChaosConfig{
		MaxConcurrentExperiments: 1,
	}

	framework, err := NewChaosEngineeringFramework(config)
	if err != nil {
		t.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Stop()

	// Test validation with invalid experiments
	testCases := []struct {
		name        string
		experiment  *ChaosExperiment
		expectError bool
	}{
		{
			name: "Valid experiment",
			experiment: &ChaosExperiment{
				ID:            "valid-exp",
				Name:          "Valid Experiment",
				TargetService: "test-service",
				TargetNodes:   []string{"node-1"},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration:     time.Minute,
				SafetyLimits: &SafetyLimits{},
			},
			expectError: false,
		},
		{
			name: "Missing ID",
			experiment: &ChaosExperiment{
				Name:          "Test",
				TargetService: "test-service",
				TargetNodes:   []string{"node-1"},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration:     time.Minute,
				SafetyLimits: &SafetyLimits{},
			},
			expectError: true,
		},
		{
			name: "Missing target service",
			experiment: &ChaosExperiment{
				ID:   "test-exp",
				Name: "Test",
				TargetNodes: []string{"node-1"},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration:     time.Minute,
				SafetyLimits: &SafetyLimits{},
			},
			expectError: true,
		},
		{
			name: "No target nodes",
			experiment: &ChaosExperiment{
				ID:            "test-exp",
				Name:          "Test",
				TargetService: "test-service",
				TargetNodes:   []string{},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration:     time.Minute,
				SafetyLimits: &SafetyLimits{},
			},
			expectError: true,
		},
		{
			name: "No failure scenarios",
			experiment: &ChaosExperiment{
				ID:               "test-exp",
				Name:             "Test",
				TargetService:    "test-service",
				TargetNodes:      []string{"node-1"},
				FailureScenarios: []*FailureScenario{},
				Duration:         time.Minute,
				SafetyLimits:     &SafetyLimits{},
			},
			expectError: true,
		},
		{
			name: "Zero duration",
			experiment: &ChaosExperiment{
				ID:            "test-exp",
				Name:          "Test",
				TargetService: "test-service",
				TargetNodes:   []string{"node-1"},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration:     0,
				SafetyLimits: &SafetyLimits{},
			},
			expectError: true,
		},
		{
			name: "Missing safety limits",
			experiment: &ChaosExperiment{
				ID:            "test-exp",
				Name:          "Test",
				TargetService: "test-service",
				TargetNodes:   []string{"node-1"},
				FailureScenarios: []*FailureScenario{
					{ID: "scenario-1", Type: "test", Enabled: true},
				},
				Duration: time.Minute,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := framework.validateExperiment(tc.experiment)
			if tc.expectError && err == nil {
				t.Error("Expected validation error")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func BenchmarkChaosExperiment(b *testing.B) {
	config := &ChaosConfig{
		EnableContinuousTesting: false,
		ReportingEnabled:        false,
		MaxConcurrentExperiments: 100,
	}

	framework, err := NewChaosEngineeringFramework(config)
	if err != nil {
		b.Fatalf("Failed to create framework: %v", err)
	}
	defer framework.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		experiment, err := framework.CreateExperimentFromTemplate(
			"service_restart",
			"test-service",
			[]string{"node-1"},
		)
		if err != nil {
			b.Fatalf("Failed to create experiment: %v", err)
		}

		experiment.Duration = time.Millisecond * 10 // Very short for benchmark
		experiment.ID = experiment.ID + string(rune(i)) // Make unique

		_, err = framework.RunExperiment(context.Background(), experiment)
		if err != nil {
			b.Fatalf("Failed to run experiment: %v", err)
		}
	}
}
