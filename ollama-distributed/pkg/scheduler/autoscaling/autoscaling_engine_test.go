package autoscaling

import (
	"testing"
	"time"
)

func TestAutoScalingEngine(t *testing.T) {
	// Create auto-scaling configuration
	config := &AutoScalingConfig{
		PredictionWindow:   time.Hour,
		ScalingCooldown:    5 * time.Minute,
		MetricsInterval:    30 * time.Second,
		MaxInstances:       10,
		MinInstances:       1,
		TargetUtilization:  0.7,
		ScaleUpThreshold:   0.8,
		ScaleDownThreshold: 0.3,
		PredictionAccuracy: 0.85,
		EnablePredictive:   true,
	}

	// Create auto-scaling engine
	engine, err := NewAutoScalingEngine(config)
	if err != nil {
		t.Fatalf("Failed to create auto-scaling engine: %v", err)
	}
	defer engine.Stop()

	// Test load prediction
	t.Run("LoadPrediction", func(t *testing.T) {
		testLoadPrediction(t, engine)
	})

	// Test resource scaling
	t.Run("ResourceScaling", func(t *testing.T) {
		testResourceScaling(t, engine)
	})

	// Test scaling policies
	t.Run("ScalingPolicies", func(t *testing.T) {
		testScalingPolicies(t, engine)
	})

	// Test metrics collection
	t.Run("MetricsCollection", func(t *testing.T) {
		testMetricsCollection(t, engine)
	})
}

func testLoadPrediction(t *testing.T, engine *AutoScalingEngine) {
	// Test load prediction
	prediction, err := engine.PredictLoad(time.Hour)
	if err != nil {
		t.Fatalf("Load prediction failed: %v", err)
	}

	// Verify prediction result
	if prediction.PredictedLoad < 0 || prediction.PredictedLoad > 1 {
		t.Errorf("Invalid predicted load: %f", prediction.PredictedLoad)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", prediction.Confidence)
	}

	if prediction.TimeHorizon != time.Hour {
		t.Errorf("Expected time horizon %v, got %v", time.Hour, prediction.TimeHorizon)
	}

	if prediction.RecommendedAction == "" {
		t.Error("Expected recommended action")
	}
}

func testResourceScaling(t *testing.T, engine *AutoScalingEngine) {
	// Test scale up scenario
	t.Run("ScaleUp", func(t *testing.T) {
		metrics := map[string]float64{
			"cpu_utilization":    0.9, // High CPU usage
			"memory_utilization": 0.7,
		}

		err := engine.ScaleResource("test-service", metrics)
		if err != nil {
			t.Fatalf("Scale up failed: %v", err)
		}

		// Check scaling history
		history := engine.GetScalingHistory()
		if len(history) == 0 {
			t.Error("Expected scaling event in history")
		}

		lastEvent := history[len(history)-1]
		if lastEvent.ResourceName != "test-service" {
			t.Errorf("Expected resource name 'test-service', got '%s'", lastEvent.ResourceName)
		}

		if !lastEvent.Success {
			t.Error("Expected successful scaling event")
		}
	})

	// Test scale down scenario
	t.Run("ScaleDown", func(t *testing.T) {
		// Use a different service name to avoid cooldown
		metrics := map[string]float64{
			"cpu_utilization":    0.2, // Low CPU usage
			"memory_utilization": 0.3,
		}

		err := engine.ScaleResource("test-service-scale-down", metrics)
		if err != nil {
			t.Fatalf("Scale down failed: %v", err)
		}
	})

	// Test cooldown period
	t.Run("CooldownPeriod", func(t *testing.T) {
		metrics := map[string]float64{
			"cpu_utilization": 0.9,
		}

		// First scaling should work
		err := engine.ScaleResource("test-service-2", metrics)
		if err != nil {
			t.Fatalf("First scaling failed: %v", err)
		}

		// Immediate second scaling should be blocked by cooldown
		err = engine.ScaleResource("test-service-2", metrics)
		if err == nil {
			t.Error("Expected cooldown error")
		}
	})
}

func testScalingPolicies(t *testing.T, engine *AutoScalingEngine) {
	// Test that default policies are created
	if len(engine.policies) == 0 {
		t.Error("Expected default scaling policies")
	}

	// Check CPU policy
	cpuPolicy, exists := engine.policies["cpu"]
	if !exists {
		t.Error("Expected CPU scaling policy")
	} else {
		if cpuPolicy.MetricType != "cpu_utilization" {
			t.Errorf("Expected CPU metric type, got %s", cpuPolicy.MetricType)
		}

		if !cpuPolicy.Enabled {
			t.Error("Expected CPU policy to be enabled")
		}

		if cpuPolicy.MaxReplicas <= cpuPolicy.MinReplicas {
			t.Error("Max replicas should be greater than min replicas")
		}
	}

	// Check memory policy
	memoryPolicy, exists := engine.policies["memory"]
	if !exists {
		t.Error("Expected memory scaling policy")
	} else {
		if memoryPolicy.MetricType != "memory_utilization" {
			t.Errorf("Expected memory metric type, got %s", memoryPolicy.MetricType)
		}
	}
}

func testMetricsCollection(t *testing.T, engine *AutoScalingEngine) {
	// Test metrics collection
	metrics, err := engine.GetMetrics()
	if err != nil {
		t.Fatalf("Metrics collection failed: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected metrics data")
	}

	// Check that inference service metrics exist
	inferenceMetrics, exists := metrics["inference-service"]
	if !exists {
		t.Error("Expected inference-service metrics")
	} else {
		// Check required metrics
		requiredMetrics := []string{"cpu_utilization", "memory_utilization", "request_rate"}
		for _, metric := range requiredMetrics {
			if _, exists := inferenceMetrics[metric]; !exists {
				t.Errorf("Missing required metric: %s", metric)
			}
		}

		// Verify metric values are reasonable
		if cpu := inferenceMetrics["cpu_utilization"]; cpu < 0 || cpu > 1 {
			t.Errorf("Invalid CPU utilization: %f", cpu)
		}

		if memory := inferenceMetrics["memory_utilization"]; memory < 0 || memory > 1 {
			t.Errorf("Invalid memory utilization: %f", memory)
		}

		if rate := inferenceMetrics["request_rate"]; rate < 0 {
			t.Errorf("Invalid request rate: %f", rate)
		}
	}
}

func TestLoadPredictor(t *testing.T) {
	predictor := NewLoadPredictor()

	// Test prediction
	prediction, err := predictor.PredictLoad(time.Hour)
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	// Verify prediction structure
	if prediction.PredictedLoad < 0 {
		t.Errorf("Predicted load should be non-negative, got %f", prediction.PredictedLoad)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.TimeHorizon != time.Hour {
		t.Errorf("Expected time horizon %v, got %v", time.Hour, prediction.TimeHorizon)
	}
}

func TestResourceScaler(t *testing.T) {
	scaler := NewResourceScaler()

	// Test getting current replicas
	replicas, err := scaler.GetCurrentReplicas("test-service")
	if err != nil {
		t.Fatalf("Failed to get current replicas: %v", err)
	}

	if replicas <= 0 {
		t.Errorf("Expected positive replica count, got %d", replicas)
	}

	// Test scaling
	err = scaler.ScaleResource("test-service", 5)
	if err != nil {
		t.Fatalf("Scaling failed: %v", err)
	}

	// Check scaling history
	history := scaler.GetScalingHistory()
	if len(history) == 0 {
		t.Error("Expected scaling event in history")
	}

	event := history[len(history)-1]
	if event.ResourceName != "test-service" {
		t.Errorf("Expected resource name 'test-service', got '%s'", event.ResourceName)
	}

	if event.ToReplicas != 5 {
		t.Errorf("Expected 5 target replicas, got %d", event.ToReplicas)
	}

	if !event.Success {
		t.Error("Expected successful scaling event")
	}
}

func TestMetricsMonitor(t *testing.T) {
	monitor := NewMetricsMonitor()

	// Test metrics collection
	err := monitor.CollectMetrics()
	if err != nil {
		t.Fatalf("Metrics collection failed: %v", err)
	}

	// Test getting current metrics
	metrics, err := monitor.GetCurrentMetrics()
	if err != nil {
		t.Fatalf("Failed to get current metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("Expected metrics data")
	}
}

func TestTargetReplicasCalculation(t *testing.T) {
	config := &AutoScalingConfig{
		TargetUtilization:  0.7,
		ScaleUpThreshold:   0.8,
		ScaleDownThreshold: 0.3,
		MaxInstances:       10,
		MinInstances:       1,
	}

	engine, err := NewAutoScalingEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Stop()

	// Test scale up calculation
	metrics := map[string]float64{"cpu_utilization": 0.9}
	target := engine.calculateTargetReplicas(metrics, 3)
	if target <= 3 {
		t.Errorf("Expected scale up, got target replicas %d", target)
	}

	// Test scale down calculation
	metrics = map[string]float64{"cpu_utilization": 0.2}
	target = engine.calculateTargetReplicas(metrics, 5)
	if target >= 5 {
		t.Errorf("Expected scale down, got target replicas %d", target)
	}

	// Test no scaling needed
	metrics = map[string]float64{"cpu_utilization": 0.7}
	target = engine.calculateTargetReplicas(metrics, 3)
	if target != 3 {
		t.Errorf("Expected no scaling, got target replicas %d", target)
	}
}

func BenchmarkAutoScaling(b *testing.B) {
	config := &AutoScalingConfig{
		PredictionWindow:   time.Minute,
		ScalingCooldown:    time.Second,
		MetricsInterval:    time.Second,
		MaxInstances:       10,
		MinInstances:       1,
		TargetUtilization:  0.7,
		ScaleUpThreshold:   0.8,
		ScaleDownThreshold: 0.3,
		EnablePredictive:   false, // Disable for benchmark
	}

	engine, err := NewAutoScalingEngine(config)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Stop()

	metrics := map[string]float64{
		"cpu_utilization":    0.85,
		"memory_utilization": 0.60,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := engine.ScaleResource("bench-service", metrics)
		if err != nil && err.Error() != "scaling cooldown in effect" {
			b.Fatalf("Scaling failed: %v", err)
		}
	}
}
