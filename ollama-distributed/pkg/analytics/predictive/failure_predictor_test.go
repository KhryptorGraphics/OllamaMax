package predictive

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestFailurePredictor(t *testing.T) {
	// Create predictor configuration
	config := &PredictorConfig{
		PredictionWindow:    time.Minute,
		FeatureWindow:       time.Hour * 24,
		ModelUpdateInterval: time.Minute,
		ConfidenceThreshold: 0.6,
		AnomalyThreshold:    0.7,
		MaxPredictions:      1000,
		EnableRetraining:    true,
	}

	// Create failure predictor
	predictor, err := NewFailurePredictor(config)
	if err != nil {
		t.Fatalf("Failed to create failure predictor: %v", err)
	}
	defer predictor.Stop()

	// Test failure prediction
	t.Run("FailurePrediction", func(t *testing.T) {
		testFailurePrediction(t, predictor)
	})

	// Test training data addition
	t.Run("TrainingData", func(t *testing.T) {
		testTrainingData(t, predictor)
	})

	// Test anomaly detection
	t.Run("AnomalyDetection", func(t *testing.T) {
		testAnomalyDetection(t, predictor)
	})

	// Test prediction accuracy
	t.Run("PredictionAccuracy", func(t *testing.T) {
		testPredictionAccuracy(t, predictor)
	})
}

func testFailurePrediction(t *testing.T, predictor *FailurePredictor) {
	// Test normal metrics (should not predict failure)
	normalMetrics := map[string]float64{
		"cpu_utilization":    0.5,
		"memory_utilization": 0.6,
		"disk_utilization":   0.4,
		"error_rate":         0.01,
		"response_time":      0.1,
		"throughput":         100.0,
	}

	prediction, err := predictor.PredictFailure("node-1", normalMetrics)
	if err != nil {
		t.Fatalf("Normal prediction failed: %v", err)
	}

	// Should not predict failure for normal metrics
	if prediction != nil {
		t.Logf("Normal metrics prediction: probability=%f, confidence=%f", prediction.Probability, prediction.Confidence)
		if prediction.Probability > predictor.config.ConfidenceThreshold {
			t.Errorf("Unexpected failure prediction for normal metrics: %f", prediction.Probability)
		}
	}

	// Test high-risk metrics (should predict failure)
	highRiskMetrics := map[string]float64{
		"cpu_utilization":    0.95, // Very high CPU
		"memory_utilization": 0.92, // Very high memory
		"disk_utilization":   0.88,
		"error_rate":         0.15, // High error rate
		"response_time":      2.0,  // High response time
		"throughput":         10.0, // Low throughput
	}

	prediction, err = predictor.PredictFailure("node-2", highRiskMetrics)
	if err != nil {
		t.Fatalf("High-risk prediction failed: %v", err)
	}

	if prediction == nil {
		t.Error("Expected failure prediction for high-risk metrics")
	} else {
		// Verify prediction structure
		if prediction.NodeID != "node-2" {
			t.Errorf("Expected node ID 'node-2', got '%s'", prediction.NodeID)
		}

		if prediction.Probability < 0 || prediction.Probability > 1 {
			t.Errorf("Invalid probability: %f", prediction.Probability)
		}

		if prediction.Confidence < 0 || prediction.Confidence > 1 {
			t.Errorf("Invalid confidence: %f", prediction.Confidence)
		}

		if prediction.TimeToFailure <= 0 {
			t.Errorf("Invalid time to failure: %v", prediction.TimeToFailure)
		}

		if prediction.FailureType == "" {
			t.Error("Expected failure type")
		}

		if len(prediction.Recommendations) == 0 {
			t.Error("Expected recommendations")
		}

		// Check severity classification
		expectedSeverities := []string{"low", "medium", "high", "critical"}
		validSeverity := false
		for _, severity := range expectedSeverities {
			if prediction.Severity == severity {
				validSeverity = true
				break
			}
		}
		if !validSeverity {
			t.Errorf("Invalid severity: %s", prediction.Severity)
		}
	}
}

func testTrainingData(t *testing.T, predictor *FailurePredictor) {
	// Add training data for failed node
	failedMetrics := map[string]float64{
		"cpu_utilization":    0.98,
		"memory_utilization": 0.95,
		"error_rate":         0.2,
	}

	err := predictor.AddTrainingData("failed-node", failedMetrics, true)
	if err != nil {
		t.Fatalf("Failed to add training data for failed node: %v", err)
	}

	// Add training data for healthy node
	healthyMetrics := map[string]float64{
		"cpu_utilization":    0.3,
		"memory_utilization": 0.4,
		"error_rate":         0.001,
	}

	err = predictor.AddTrainingData("healthy-node", healthyMetrics, false)
	if err != nil {
		t.Fatalf("Failed to add training data for healthy node: %v", err)
	}

	// Verify training data was added
	if len(predictor.model.trainingData) < 2 {
		t.Error("Expected training data to be added")
	}
}

func testAnomalyDetection(t *testing.T, predictor *FailurePredictor) {
	// Test normal features (should not be anomaly) - normalized values
	normalFeatures := []float64{0.5, 0.6, 0.4, 0.3, 0.1, 0.2, 0.8, 0.7, 0.5, 0.4, 0.0, 0.0, 0.0, 0.5, 0.3}

	anomalyScore := predictor.anomalyDetector.DetectAnomaly(normalFeatures)
	if anomalyScore.IsAnomaly {
		t.Errorf("Normal features incorrectly detected as anomaly: score=%f", anomalyScore.Score)
	}

	// Test extreme features (should be anomaly) - many extreme values
	extremeFeatures := []float64{0.99, 0.98, 0.97, 0.96, 0.95, 0.94, 0.93, 0.92, 0.91, 0.90, 0.89, 0.88, 0.87, 0.86, 0.85}

	anomalyScore = predictor.anomalyDetector.DetectAnomaly(extremeFeatures)
	if !anomalyScore.IsAnomaly {
		t.Errorf("Extreme features not detected as anomaly: score=%f", anomalyScore.Score)
	}

	// Verify anomaly score structure
	if anomalyScore.Score < 0 {
		t.Errorf("Invalid anomaly score: %f", anomalyScore.Score)
	}

	if anomalyScore.Threshold != predictor.config.AnomalyThreshold {
		t.Errorf("Expected threshold %f, got %f", predictor.config.AnomalyThreshold, anomalyScore.Threshold)
	}
}

func testPredictionAccuracy(t *testing.T, predictor *FailurePredictor) {
	// Add multiple training examples
	trainingCases := []struct {
		metrics map[string]float64
		failed  bool
	}{
		{map[string]float64{"cpu_utilization": 0.95, "memory_utilization": 0.9, "error_rate": 0.1}, true},
		{map[string]float64{"cpu_utilization": 0.98, "memory_utilization": 0.95, "error_rate": 0.15}, true},
		{map[string]float64{"cpu_utilization": 0.3, "memory_utilization": 0.4, "error_rate": 0.001}, false},
		{map[string]float64{"cpu_utilization": 0.2, "memory_utilization": 0.3, "error_rate": 0.002}, false},
		{map[string]float64{"cpu_utilization": 0.5, "memory_utilization": 0.6, "error_rate": 0.01}, false},
	}

	for i, tc := range trainingCases {
		err := predictor.AddTrainingData(fmt.Sprintf("node-%d", i), tc.metrics, tc.failed)
		if err != nil {
			t.Fatalf("Failed to add training case %d: %v", i, err)
		}
	}

	// Test predictions on similar data
	testCases := []struct {
		metrics  map[string]float64
		expected bool
	}{
		{map[string]float64{"cpu_utilization": 0.94, "memory_utilization": 0.88, "error_rate": 0.12}, true},
		{map[string]float64{"cpu_utilization": 0.25, "memory_utilization": 0.35, "error_rate": 0.003}, false},
	}

	correct := 0
	total := len(testCases)

	for i, tc := range testCases {
		prediction, err := predictor.PredictFailure(fmt.Sprintf("test-node-%d", i), tc.metrics)
		if err != nil {
			t.Fatalf("Prediction failed for test case %d: %v", i, err)
		}

		predicted := prediction != nil && prediction.Probability > predictor.config.ConfidenceThreshold
		if predicted == tc.expected {
			correct++
		}
	}

	accuracy := float64(correct) / float64(total)
	if accuracy < 0.5 { // At least 50% accuracy expected
		t.Errorf("Low prediction accuracy: %f (%d/%d)", accuracy, correct, total)
	}
}

func TestFailurePredictionModel(t *testing.T) {
	model := NewFailurePredictionModel()

	// Test model initialization
	if len(model.layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(model.layers))
	}

	if model.layers[0] != 15 || model.layers[1] != 10 || model.layers[2] != 1 {
		t.Errorf("Unexpected layer sizes: %v", model.layers)
	}

	// Test prediction
	features := make([]float64, 15)
	for i := range features {
		features[i] = 0.5
	}

	prediction, err := model.Predict(features)
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if prediction.Probability < 0 || prediction.Probability > 1 {
		t.Errorf("Invalid probability: %f", prediction.Probability)
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", prediction.Confidence)
	}

	// Test training example addition
	example := &TrainingExample{
		Features: features,
		Label:    1.0,
		Weight:   1.0,
		NodeID:   "test-node",
		Time:     time.Now(),
	}

	err = model.AddTrainingExample(example)
	if err != nil {
		t.Fatalf("Failed to add training example: %v", err)
	}

	if len(model.trainingData) != 1 {
		t.Errorf("Expected 1 training example, got %d", len(model.trainingData))
	}
}

func TestAnomalyDetector(t *testing.T) {
	detector := NewAnomalyDetector(0.7)

	// Test normal features
	normalFeatures := []float64{0.5, 0.5, 0.5, 0.5, 0.5}
	score := detector.DetectAnomaly(normalFeatures)

	if score.IsAnomaly {
		t.Errorf("Normal features detected as anomaly: %f", score.Score)
	}

	// Test anomalous features - many extreme values
	anomalousFeatures := []float64{0.99, 0.98, 0.97, 0.96, 0.95, 0.94, 0.93, 0.92, 0.91, 0.90}
	score = detector.DetectAnomaly(anomalousFeatures)

	if !score.IsAnomaly {
		t.Errorf("Anomalous features not detected: %f", score.Score)
	}

	// Test history tracking
	if len(detector.history) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(detector.history))
	}
}

func TestFeatureExtraction(t *testing.T) {
	config := &PredictorConfig{
		PredictionWindow:    time.Hour,
		ConfidenceThreshold: 0.6,
		AnomalyThreshold:    0.7,
	}

	predictor, err := NewFailurePredictor(config)
	if err != nil {
		t.Fatalf("Failed to create predictor: %v", err)
	}
	defer predictor.Stop()

	metrics := map[string]float64{
		"cpu_utilization":     0.8,
		"memory_utilization":  0.7,
		"disk_utilization":    0.6,
		"network_utilization": 0.5,
		"response_time":       0.2,
		"error_rate":          0.05,
		"throughput":          150.0,
		"load_average":        2.5,
		"connection_count":    100.0,
		"queue_depth":         10.0,
	}

	features := predictor.extractPredictionFeatures("test-node", metrics)

	expectedFeatureCount := 15 // As defined in getFeatureNames
	if len(features) != expectedFeatureCount {
		t.Errorf("Expected %d features, got %d", expectedFeatureCount, len(features))
	}

	// Verify feature values are reasonable
	for i, feature := range features {
		if math.IsNaN(feature) || math.IsInf(feature, 0) {
			t.Errorf("Invalid feature value at index %d: %f", i, feature)
		}
	}
}

func BenchmarkFailurePrediction(b *testing.B) {
	config := &PredictorConfig{
		PredictionWindow:    time.Hour,
		ConfidenceThreshold: 0.6,
		AnomalyThreshold:    0.7,
		EnableRetraining:    false, // Disable for benchmark
	}

	predictor, err := NewFailurePredictor(config)
	if err != nil {
		b.Fatalf("Failed to create predictor: %v", err)
	}
	defer predictor.Stop()

	metrics := map[string]float64{
		"cpu_utilization":    0.8,
		"memory_utilization": 0.7,
		"error_rate":         0.05,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := predictor.PredictFailure("bench-node", metrics)
		if err != nil {
			b.Fatalf("Prediction failed: %v", err)
		}
	}
}
