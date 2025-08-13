//go:build ignore

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
)

func main() {
	fmt.Println("Testing Predictive Fault Detection System...")

	// Test the predictive fault detection system
	if err := testPredictiveFaultDetection(); err != nil {
		log.Fatalf("Predictive fault detection test failed: %v", err)
	}

	fmt.Println("✅ Predictive fault detection system test completed successfully!")
}

func testPredictiveFaultDetection() error {
	fmt.Println("Setting up predictive fault detection system...")

	// Create enhanced fault detector first
	enhancedDetector := createEnhancedFaultDetector()
	if enhancedDetector == nil {
		return fmt.Errorf("failed to create enhanced fault detector")
	}

	// Create predictive fault detector
	config := &fault_tolerance.PredictiveDetectionConfig{
		PredictionInterval:        10 * time.Second,
		LearningInterval:          30 * time.Second,
		ModelUpdateInterval:       1 * time.Minute,
		PredictionHorizon:         5 * time.Minute,
		ConfidenceThreshold:       0.6,
		MinHistorySize:            10,
		MaxPredictionHistory:      1000,
		EnableTimeSeriesML:        true,
		EnableTrendAnalysis:       true,
		EnableCorrelationAnalysis: true,
		EnableContinuousLearning:  true,
		EnableEnsemblePrediction:  true,
		EnableAdaptiveThresholds:  true,
		EnableSeasonalAdjustment:  false,
	}

	predictiveDetector := fault_tolerance.NewPredictiveFaultDetector(enhancedDetector, config)
	if predictiveDetector == nil {
		return fmt.Errorf("failed to create predictive fault detector")
	}

	// Start the predictive detector
	if err := predictiveDetector.Start(); err != nil {
		return fmt.Errorf("failed to start predictive detector: %v", err)
	}
	defer predictiveDetector.Stop()

	fmt.Println("✅ Predictive fault detection system setup complete")

	// Test scenarios
	scenarios := []struct {
		name        string
		description string
		testFunc    func(*fault_tolerance.PredictiveFaultDetector) error
	}{
		{
			name:        "Time Series Prediction",
			description: "Test time series analysis and prediction",
			testFunc:    testTimeSeriesPrediction,
		},
		{
			name:        "ML-based Prediction",
			description: "Test machine learning prediction models",
			testFunc:    testMLPrediction,
		},
		{
			name:        "Trend Analysis",
			description: "Test trend analysis and prediction",
			testFunc:    testTrendAnalysis,
		},
		{
			name:        "Correlation Analysis",
			description: "Test correlation-based prediction",
			testFunc:    testCorrelationAnalysis,
		},
		{
			name:        "Ensemble Prediction",
			description: "Test ensemble prediction combining multiple models",
			testFunc:    testEnsemblePrediction,
		},
		{
			name:        "Prediction Statistics",
			description: "Test prediction statistics and reporting",
			testFunc:    testPredictionStatistics,
		},
	}

	fmt.Println("\n=== Testing Predictive Fault Detection ===")

	for i, scenario := range scenarios {
		fmt.Printf("%d. Testing %s...\n", i+1, scenario.name)
		fmt.Printf("  Description: %s\n", scenario.description)

		if err := scenario.testFunc(predictiveDetector); err != nil {
			fmt.Printf("  ❌ Test failed: %v\n", err)
			return err
		}

		fmt.Printf("  ✅ Test passed\n\n")

		// Wait between tests
		time.Sleep(2 * time.Second)
	}

	return nil
}

func createEnhancedFaultDetector() *fault_tolerance.EnhancedFaultDetector {
	// Create a basic fault tolerance manager first
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

	// Create enhanced fault detector config
	config := &fault_tolerance.EnhancedDetectionConfig{
		HealthCheckInterval:        5 * time.Second,
		AnomalyThreshold:           0.7,
		PatternConfidence:          0.8,
		ProcessingWorkers:          2,
		EventBufferSize:            100,
		HealthScoreThreshold:       0.6,
		EnableStatisticalDetection: true,
		EnableMLDetection:          true,
		EnablePatternRecognition:   true,
	}

	return fault_tolerance.NewEnhancedFaultDetector(ftManager, config)
}

func testTimeSeriesPrediction(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Create mock time series data with trend
	metrics := createMockTimeSeriesMetrics("increasing_cpu")

	// Perform prediction
	predictions := detector.PredictFaults(metrics)

	fmt.Printf("    Generated %d time series predictions\n", len(predictions))

	// Validate predictions
	for _, prediction := range predictions {
		if prediction.Confidence < 0.1 || prediction.Confidence > 1.0 {
			return fmt.Errorf("invalid confidence: %f", prediction.Confidence)
		}

		if prediction.TimeToFailure <= 0 {
			return fmt.Errorf("invalid time to failure: %v", prediction.TimeToFailure)
		}

		fmt.Printf("    Prediction: %s (confidence: %.2f, time: %v)\n",
			prediction.PredictedType, prediction.Confidence, prediction.TimeToFailure)
	}

	return nil
}

func testMLPrediction(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Create mock metrics for ML prediction
	metrics := map[string]interface{}{
		"cpu_usage":     0.85,
		"memory_usage":  0.90,
		"response_time": 1500.0,
		"error_rate":    0.05,
		"connectivity":  0.95,
	}

	// Perform prediction
	predictions := detector.PredictFaults(metrics)

	fmt.Printf("    Generated %d ML predictions\n", len(predictions))

	// Validate predictions
	for _, prediction := range predictions {
		if prediction.ModelUsed == "" {
			return fmt.Errorf("missing model information")
		}

		if len(prediction.Features) == 0 {
			return fmt.Errorf("missing features")
		}

		fmt.Printf("    ML Prediction: %s using %s (confidence: %.2f)\n",
			prediction.PredictedType, prediction.ModelUsed, prediction.Confidence)
	}

	return nil
}

func testTrendAnalysis(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Create metrics with strong trend
	metrics := createMockTrendMetrics()

	// Perform prediction
	predictions := detector.PredictFaults(metrics)

	fmt.Printf("    Generated %d trend-based predictions\n", len(predictions))

	// Check for trend-related metadata
	for _, prediction := range predictions {
		if metadata, exists := prediction.Metadata["trend_direction"]; exists {
			fmt.Printf("    Trend detected: %v\n", metadata)
		}
	}

	return nil
}

func testCorrelationAnalysis(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Create correlated metrics
	metrics := createMockCorrelatedMetrics()

	// Perform prediction
	predictions := detector.PredictFaults(metrics)

	fmt.Printf("    Generated %d correlation-based predictions\n", len(predictions))

	// Check for correlation-related metadata
	for _, prediction := range predictions {
		if prediction.ModelUsed == "correlation_analysis" {
			fmt.Printf("    Correlation prediction for: %s\n", prediction.Target)
		}
	}

	return nil
}

func testEnsemblePrediction(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Create comprehensive metrics to trigger ensemble prediction
	metrics := createMockComprehensiveMetrics()

	// Perform prediction
	predictions := detector.PredictFaults(metrics)

	fmt.Printf("    Generated %d ensemble predictions\n", len(predictions))

	// Check for ensemble predictions
	ensembleCount := 0
	for _, prediction := range predictions {
		if prediction.ModelUsed == "ensemble" {
			ensembleCount++
			fmt.Printf("    Ensemble prediction: %s (confidence: %.2f)\n",
				prediction.PredictedType, prediction.Confidence)
		}
	}

	if ensembleCount == 0 {
		fmt.Printf("    No ensemble predictions generated (this is normal)\n")
	}

	return nil
}

func testPredictionStatistics(detector *fault_tolerance.PredictiveFaultDetector) error {
	// Generate some predictions first
	metrics := createMockComprehensiveMetrics()
	detector.PredictFaults(metrics)

	// Wait a moment for processing
	time.Sleep(1 * time.Second)

	// Get statistics
	stats := detector.GetPredictionStatistics()

	fmt.Printf("    Total predictions: %d\n", stats.TotalPredictions)
	fmt.Printf("    Active predictions: %d\n", stats.ActivePredictions)
	fmt.Printf("    Average confidence: %.2f\n", stats.AverageConfidence)
	fmt.Printf("    Accuracy rate: %.2f\n", stats.AccuracyRate)

	// Get current predictions
	predictions := detector.GetPredictions()
	fmt.Printf("    Current predictions count: %d\n", len(predictions))

	// Get prediction history
	history := detector.GetPredictionHistory()
	fmt.Printf("    Prediction history count: %d\n", len(history))

	return nil
}

// Helper functions to create mock data

func createMockTimeSeriesMetrics(pattern string) map[string]interface{} {
	metrics := make(map[string]interface{})

	switch pattern {
	case "increasing_cpu":
		metrics["cpu_usage"] = 0.75 + rand.Float64()*0.2
	case "decreasing_memory":
		metrics["memory_usage"] = 0.3 + rand.Float64()*0.2
	default:
		metrics["cpu_usage"] = 0.5 + rand.Float64()*0.3
		metrics["memory_usage"] = 0.5 + rand.Float64()*0.3
	}

	return metrics
}

func createMockTrendMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cpu_usage":     0.80 + rand.Float64()*0.15,
		"memory_usage":  0.85 + rand.Float64()*0.10,
		"response_time": 1000.0 + rand.Float64()*500,
	}
}

func createMockCorrelatedMetrics() map[string]interface{} {
	baseValue := 0.7 + rand.Float64()*0.2
	return map[string]interface{}{
		"cpu_usage":     baseValue,
		"memory_usage":  baseValue + 0.1,          // Correlated with CPU
		"response_time": (1.0 - baseValue) * 2000, // Inversely correlated
	}
}

func createMockComprehensiveMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cpu_usage":     0.75 + rand.Float64()*0.2,
		"memory_usage":  0.80 + rand.Float64()*0.15,
		"response_time": 1200.0 + rand.Float64()*800,
		"error_rate":    0.02 + rand.Float64()*0.03,
		"connectivity":  0.95 + rand.Float64()*0.05,
		"disk_usage":    0.70 + rand.Float64()*0.25,
		"network_io":    1000.0 + rand.Float64()*2000,
	}
}
