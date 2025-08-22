package maintenance

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/analytics/predictive"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/autoscaling"
)

func TestProactiveMaintenanceEngine(t *testing.T) {
	// Create dependencies
	predictorConfig := &predictive.PredictorConfig{
		PredictionWindow:    time.Minute,
		FeatureWindow:       time.Hour,
		ModelUpdateInterval: time.Minute,
		ConfidenceThreshold: 0.6,
		AnomalyThreshold:    0.7,
		MaxPredictions:      1000,
		EnableRetraining:    false,
	}

	failurePredictor, err := predictive.NewFailurePredictor(predictorConfig)
	if err != nil {
		t.Fatalf("Failed to create failure predictor: %v", err)
	}
	defer failurePredictor.Stop()

	autoScalerConfig := &autoscaling.AutoScalingConfig{
		PredictionWindow:   time.Minute,
		ScalingCooldown:    time.Minute,
		MetricsInterval:    time.Second * 30,
		MaxInstances:       10,
		MinInstances:       1,
		TargetUtilization:  0.7,
		ScaleUpThreshold:   0.8,
		ScaleDownThreshold: 0.3,
		EnablePredictive:   false,
	}

	autoScaler, err := autoscaling.NewAutoScalingEngine(autoScalerConfig)
	if err != nil {
		t.Fatalf("Failed to create auto scaler: %v", err)
	}
	defer autoScaler.Stop()

	// Create maintenance engine configuration
	maintenanceConfig := &MaintenanceConfig{
		PredictionThreshold:    0.7,
		MaintenanceWindow:      time.Hour,
		MaxConcurrentTasks:     3,
		WorkloadMigrationDelay: time.Minute * 5,
		SelfHealingEnabled:     true,
		AutoScalingEnabled:     true,
		MaintenanceInterval:    time.Minute,
		EmergencyThreshold:     0.9,
	}

	// Create maintenance engine
	engine, err := NewProactiveMaintenanceEngine(failurePredictor, autoScaler, maintenanceConfig)
	if err != nil {
		t.Fatalf("Failed to create maintenance engine: %v", err)
	}
	defer engine.Stop()

	// Test maintenance scheduling
	t.Run("MaintenanceScheduling", func(t *testing.T) {
		testMaintenanceScheduling(t, engine)
	})

	// Test self-healing
	t.Run("SelfHealing", func(t *testing.T) {
		testSelfHealing(t, engine)
	})

	// Test workload migration
	t.Run("WorkloadMigration", func(t *testing.T) {
		testWorkloadMigration(t, engine)
	})

	// Test maintenance recommendations
	t.Run("MaintenanceRecommendations", func(t *testing.T) {
		testMaintenanceRecommendations(t, engine, failurePredictor)
	})
}

func testMaintenanceScheduling(t *testing.T, engine *ProactiveMaintenanceEngine) {
	// Create a failure prediction
	prediction := &predictive.FailurePrediction{
		NodeID:          "test-node-1",
		FailureType:     "cpu_exhaustion",
		Probability:     0.8,
		Confidence:      0.9,
		TimeToFailure:   time.Hour,
		PredictedTime:   time.Now().Add(time.Hour),
		RootCause:       "High CPU utilization detected",
		Severity:        "high",
		Recommendations: []string{"Scale up CPU resources", "Optimize CPU-intensive processes"},
		Timestamp:       time.Now(),
	}

	// Schedule maintenance
	task, err := engine.ScheduleMaintenance("test-node-1", prediction)
	if err != nil {
		t.Fatalf("Failed to schedule maintenance: %v", err)
	}

	// Verify task properties
	if task.NodeID != "test-node-1" {
		t.Errorf("Expected node ID 'test-node-1', got '%s'", task.NodeID)
	}

	if task.TaskType != "resource_optimization" {
		t.Errorf("Expected task type 'resource_optimization', got '%s'", task.TaskType)
	}

	if task.Priority != 2 { // High priority for 0.8 probability
		t.Errorf("Expected priority 2, got %d", task.Priority)
	}

	if task.Status != "scheduled" {
		t.Errorf("Expected status 'scheduled', got '%s'", task.Status)
	}

	if len(task.Actions) == 0 {
		t.Error("Expected maintenance actions")
	}

	if len(task.Prerequisites) == 0 {
		t.Error("Expected prerequisites")
	}

	// Test maintenance execution
	err = engine.ExecuteMaintenance(task.ID)
	if err != nil {
		t.Fatalf("Failed to execute maintenance: %v", err)
	}

	// Verify task completion
	if task.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", task.Status)
	}

	if !task.Success {
		t.Error("Expected successful maintenance")
	}

	if task.ActualDuration <= 0 {
		t.Error("Expected positive actual duration")
	}
}

func testSelfHealing(t *testing.T, engine *ProactiveMaintenanceEngine) {
	// Test self-healing trigger
	action, err := engine.TriggerSelfHealing("test-node-2", "service_degradation")
	if err != nil {
		t.Fatalf("Failed to trigger self-healing: %v", err)
	}

	// Verify healing action
	if action.NodeID != "test-node-2" {
		t.Errorf("Expected node ID 'test-node-2', got '%s'", action.NodeID)
	}

	if action.Issue != "service_degradation" {
		t.Errorf("Expected issue 'service_degradation', got '%s'", action.Issue)
	}

	if action.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", action.Status)
	}

	if !action.Success {
		t.Error("Expected successful healing")
	}

	if len(action.Actions) == 0 {
		t.Error("Expected healing actions")
	}
}

func testWorkloadMigration(t *testing.T, engine *ProactiveMaintenanceEngine) {
	// Test workload migration scheduling
	migrationTask, err := engine.workloadMigrator.ScheduleMigration("source-node", "inference")
	if err != nil {
		t.Fatalf("Failed to schedule workload migration: %v", err)
	}

	// Verify migration task
	if migrationTask.SourceNodeID != "source-node" {
		t.Errorf("Expected source node 'source-node', got '%s'", migrationTask.SourceNodeID)
	}

	if migrationTask.WorkloadType != "inference" {
		t.Errorf("Expected workload type 'inference', got '%s'", migrationTask.WorkloadType)
	}

	if migrationTask.Status != "scheduled" {
		t.Errorf("Expected status 'scheduled', got '%s'", migrationTask.Status)
	}

	if migrationTask.TargetNodeID == "" {
		t.Error("Expected target node ID")
	}
}

func testMaintenanceRecommendations(t *testing.T, engine *ProactiveMaintenanceEngine, predictor *predictive.FailurePredictor) {
	// Add some training data to create predictions
	highRiskMetrics := map[string]float64{
		"cpu_utilization":    0.95,
		"memory_utilization": 0.90,
		"error_rate":         0.1,
	}

	err := predictor.AddTrainingData("test-node-3", highRiskMetrics, true)
	if err != nil {
		t.Fatalf("Failed to add training data: %v", err)
	}

	// Create a prediction
	prediction, err := predictor.PredictFailure("test-node-3", highRiskMetrics)
	if err != nil {
		t.Fatalf("Failed to predict failure: %v", err)
	}

	// Only proceed if we have a prediction
	if prediction != nil {
		// Get maintenance recommendations
		recommendations, err := engine.GetMaintenanceRecommendations()
		if err != nil {
			t.Fatalf("Failed to get maintenance recommendations: %v", err)
		}

		// Verify recommendations
		if len(recommendations) == 0 {
			t.Error("Expected maintenance recommendations")
		} else {
			rec := recommendations[0]
			
			if rec.NodeID == "" {
				t.Error("Expected node ID in recommendation")
			}

			if rec.MaintenanceType == "" {
				t.Error("Expected maintenance type in recommendation")
			}

			if rec.Priority <= 0 {
				t.Error("Expected positive priority")
			}

			if rec.EstimatedDuration <= 0 {
				t.Error("Expected positive estimated duration")
			}

			if len(rec.Actions) == 0 {
				t.Error("Expected recommended actions")
			}

			if rec.Confidence < 0 || rec.Confidence > 1 {
				t.Errorf("Invalid confidence: %f", rec.Confidence)
			}
		}
	}
}

func TestWorkloadMigrator(t *testing.T) {
	migrator := NewWorkloadMigrator()

	// Test migration scheduling
	task, err := migrator.ScheduleMigration("node-1", "inference")
	if err != nil {
		t.Fatalf("Failed to schedule migration: %v", err)
	}

	if task.SourceNodeID != "node-1" {
		t.Errorf("Expected source node 'node-1', got '%s'", task.SourceNodeID)
	}

	if task.WorkloadType != "inference" {
		t.Errorf("Expected workload type 'inference', got '%s'", task.WorkloadType)
	}

	if task.Status != "scheduled" {
		t.Errorf("Expected status 'scheduled', got '%s'", task.Status)
	}
}

func TestSelfHealingEngine(t *testing.T) {
	engine := NewSelfHealingEngine()

	// Test healing action
	action, err := engine.Heal(nil, "test-node", "memory_leak")
	if err != nil {
		t.Fatalf("Failed to perform healing: %v", err)
	}

	if action.NodeID != "test-node" {
		t.Errorf("Expected node ID 'test-node', got '%s'", action.NodeID)
	}

	if action.Issue != "memory_leak" {
		t.Errorf("Expected issue 'memory_leak', got '%s'", action.Issue)
	}

	if action.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", action.Status)
	}

	if !action.Success {
		t.Error("Expected successful healing")
	}

	// Verify healing history
	if len(engine.healingHistory) != 1 {
		t.Errorf("Expected 1 healing action in history, got %d", len(engine.healingHistory))
	}
}

func TestMaintenanceScheduler(t *testing.T) {
	scheduler := NewMaintenanceScheduler()

	// Verify scheduler initialization
	if scheduler.maintenanceWindow == nil {
		t.Error("Expected maintenance window")
	}

	if scheduler.optimizer == nil {
		t.Error("Expected schedule optimizer")
	}

	// Verify maintenance window configuration
	window := scheduler.maintenanceWindow
	if window.StartHour != 2 {
		t.Errorf("Expected start hour 2, got %d", window.StartHour)
	}

	if window.EndHour != 6 {
		t.Errorf("Expected end hour 6, got %d", window.EndHour)
	}

	if len(window.Days) != 2 {
		t.Errorf("Expected 2 maintenance days, got %d", len(window.Days))
	}

	if window.MaxDuration != time.Hour*4 {
		t.Errorf("Expected max duration 4 hours, got %v", window.MaxDuration)
	}
}

func TestMaintenanceTaskPriority(t *testing.T) {
	config := &MaintenanceConfig{
		PredictionThreshold: 0.6,
		EmergencyThreshold:  0.9,
	}

	engine := &ProactiveMaintenanceEngine{
		config: config,
	}

	// Test critical priority
	criticalPrediction := &predictive.FailurePrediction{
		Probability: 0.95,
	}
	priority := engine.calculatePriority(criticalPrediction)
	if priority != 1 {
		t.Errorf("Expected critical priority 1, got %d", priority)
	}

	// Test high priority
	highPrediction := &predictive.FailurePrediction{
		Probability: 0.8,
	}
	priority = engine.calculatePriority(highPrediction)
	if priority != 2 {
		t.Errorf("Expected high priority 2, got %d", priority)
	}

	// Test medium priority
	mediumPrediction := &predictive.FailurePrediction{
		Probability: 0.6,
	}
	priority = engine.calculatePriority(mediumPrediction)
	if priority != 3 {
		t.Errorf("Expected medium priority 3, got %d", priority)
	}

	// Test low priority
	lowPrediction := &predictive.FailurePrediction{
		Probability: 0.4,
	}
	priority = engine.calculatePriority(lowPrediction)
	if priority != 4 {
		t.Errorf("Expected low priority 4, got %d", priority)
	}
}

func BenchmarkMaintenanceScheduling(b *testing.B) {
	// Create minimal dependencies for benchmark
	predictorConfig := &predictive.PredictorConfig{
		PredictionWindow:    time.Minute,
		ConfidenceThreshold: 0.6,
		AnomalyThreshold:    0.7,
		EnableRetraining:    false,
	}

	failurePredictor, err := predictive.NewFailurePredictor(predictorConfig)
	if err != nil {
		b.Fatalf("Failed to create failure predictor: %v", err)
	}
	defer failurePredictor.Stop()

	autoScalerConfig := &autoscaling.AutoScalingConfig{
		PredictionWindow:  time.Minute,
		ScalingCooldown:   time.Minute,
		MetricsInterval:   time.Second * 30,
		MaxInstances:      10,
		MinInstances:      1,
		TargetUtilization: 0.7,
		EnablePredictive:  false,
	}

	autoScaler, err := autoscaling.NewAutoScalingEngine(autoScalerConfig)
	if err != nil {
		b.Fatalf("Failed to create auto scaler: %v", err)
	}
	defer autoScaler.Stop()

	maintenanceConfig := &MaintenanceConfig{
		PredictionThreshold: 0.7,
		MaintenanceWindow:   time.Hour,
		MaxConcurrentTasks:  3,
		SelfHealingEnabled:  false, // Disable for benchmark
		AutoScalingEnabled:  false, // Disable for benchmark
		MaintenanceInterval: time.Hour,
		EmergencyThreshold:  0.9,
	}

	engine, err := NewProactiveMaintenanceEngine(failurePredictor, autoScaler, maintenanceConfig)
	if err != nil {
		b.Fatalf("Failed to create maintenance engine: %v", err)
	}
	defer engine.Stop()

	prediction := &predictive.FailurePrediction{
		NodeID:          "bench-node",
		FailureType:     "cpu_exhaustion",
		Probability:     0.8,
		Confidence:      0.9,
		TimeToFailure:   time.Hour,
		PredictedTime:   time.Now().Add(time.Hour),
		RootCause:       "High CPU utilization",
		Recommendations: []string{"Scale up resources"},
		Timestamp:       time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.ScheduleMaintenance("bench-node", prediction)
		if err != nil {
			b.Fatalf("Maintenance scheduling failed: %v", err)
		}
	}
}
