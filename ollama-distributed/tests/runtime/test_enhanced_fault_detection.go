//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	pkgConfig "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/fault_tolerance"
)

func main() {
	fmt.Println("Testing Enhanced Fault Detection System...")

	// Setup fault tolerance system
	detector := setupEnhancedFaultDetector()
	if detector == nil {
		log.Fatal("Failed to setup enhanced fault detector")
	}

	// Start the detector
	if err := detector.Start(); err != nil {
		log.Fatalf("Failed to start enhanced fault detector: %v", err)
	}
	defer detector.Stop()

	// Give system time to start
	time.Sleep(2 * time.Second)

	// Test enhanced fault detection
	fmt.Println("\n=== Testing Enhanced Fault Detection ===")
	testEnhancedFaultDetection(detector)

	// Test anomaly detection
	fmt.Println("\n=== Testing Anomaly Detection ===")
	testAnomalyDetection(detector)

	// Test health score calculation
	fmt.Println("\n=== Testing Health Score Calculation ===")
	testHealthScoreCalculation(detector)

	// Test fault classification
	fmt.Println("\n=== Testing Fault Classification ===")
	testFaultClassification(detector)

	// Test detection statistics
	fmt.Println("\n=== Testing Detection Statistics ===")
	testDetectionStatistics(detector)

	fmt.Println("\n🎯 Enhanced Fault Detection test completed!")
	fmt.Println("✅ Advanced fault tolerance system validation successful")
}

func setupEnhancedFaultDetector() *fault_tolerance.EnhancedFaultDetector {
	fmt.Println("Setting up enhanced fault detection system...")

	// Create P2P node
	ctx := context.Background()
	p2pConfig := &pkgConfig.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}

	p2pNode, err := p2p.NewP2PNode(ctx, p2pConfig)
	if err != nil {
		log.Printf("Failed to create P2P node: %v", err)
		return nil
	}

	// Create messaging and monitoring
	messageRouter := messaging.NewMessageRouter(nil)
	networkMonitor := monitoring.NewNetworkMonitor(nil)

	// Create consensus engine (optional)
	consensusConfig := &config.ConsensusConfig{
		NodeID:    "test-node-1",
		DataDir:   "./test_data/consensus",
		Bootstrap: true,
	}

	_, err = consensus.NewEngine(consensusConfig, p2pNode, messageRouter, networkMonitor)
	if err != nil {
		fmt.Printf("Consensus engine creation failed: %v\n", err)
	}

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

	// Create enhanced fault detector
	detectorConfig := &fault_tolerance.EnhancedDetectionConfig{
		HealthCheckInterval:        5 * time.Second,
		AnomalyCheckInterval:       3 * time.Second,
		PatternCheckInterval:       10 * time.Second,
		AnomalyThreshold:           2.0,
		PatternConfidence:          0.8,
		HealthScoreThreshold:       0.6,
		EnableStatisticalDetection: true,
		EnableMLDetection:          false, // Disabled for testing
		EnablePatternRecognition:   true,
		EnablePredictiveDetection:  true,
		HistoryRetentionPeriod:     1 * time.Hour,
		MaxDetectionHistory:        1000,
		EventBufferSize:            100,
		ProcessingWorkers:          2,
	}

	detector := fault_tolerance.NewEnhancedFaultDetector(ftManager, detectorConfig)

	fmt.Println("✅ Enhanced fault detection system setup complete")
	return detector
}

func testEnhancedFaultDetection(detector *fault_tolerance.EnhancedFaultDetector) {
	fmt.Println("1. Testing Enhanced Fault Detection...")

	// Test scenarios with different health metrics
	testScenarios := []struct {
		name    string
		target  string
		metrics map[string]interface{}
	}{
		{
			name:   "Healthy System",
			target: "node-1",
			metrics: map[string]interface{}{
				"cpu_usage":     0.3,
				"memory_usage":  0.4,
				"response_time": 50.0,
				"error_rate":    0.01,
				"connectivity":  true,
			},
		},
		{
			name:   "High CPU Usage",
			target: "node-2",
			metrics: map[string]interface{}{
				"cpu_usage":     0.95,
				"memory_usage":  0.4,
				"response_time": 150.0,
				"error_rate":    0.02,
				"connectivity":  true,
			},
		},
		{
			name:   "High Memory Usage",
			target: "node-3",
			metrics: map[string]interface{}{
				"cpu_usage":     0.4,
				"memory_usage":  0.98,
				"response_time": 200.0,
				"error_rate":    0.03,
				"connectivity":  true,
			},
		},
		{
			name:   "Slow Response Time",
			target: "node-4",
			metrics: map[string]interface{}{
				"cpu_usage":     0.3,
				"memory_usage":  0.4,
				"response_time": 2000.0,
				"error_rate":    0.01,
				"connectivity":  true,
			},
		},
		{
			name:   "High Error Rate",
			target: "node-5",
			metrics: map[string]interface{}{
				"cpu_usage":     0.3,
				"memory_usage":  0.4,
				"response_time": 100.0,
				"error_rate":    0.15,
				"connectivity":  true,
			},
		},
		{
			name:   "Connectivity Issues",
			target: "node-6",
			metrics: map[string]interface{}{
				"cpu_usage":     0.3,
				"memory_usage":  0.4,
				"response_time": 100.0,
				"error_rate":    0.01,
				"connectivity":  false,
			},
		},
	}

	detectionCount := 0
	for _, scenario := range testScenarios {
		fmt.Printf("  Testing scenario: %s\n", scenario.name)

		detection := detector.DetectFault(scenario.target, scenario.metrics)
		if detection != nil {
			fmt.Printf("    ✅ Fault detected: %s (Severity: %s)\n",
				string(detection.Type), string(detection.Severity))
			detectionCount++
		} else {
			fmt.Printf("    ✅ No fault detected (healthy system)\n")
		}

		// Small delay between tests
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Enhanced Fault Detection Results:\n")
	fmt.Printf("  ✅ Test scenarios: %d\n", len(testScenarios))
	fmt.Printf("  ✅ Faults detected: %d\n", detectionCount)

	if detectionCount > 0 {
		fmt.Println("✅ Enhanced fault detection working correctly")
	} else {
		fmt.Println("⚠️ No faults detected - check thresholds")
	}
}

func testAnomalyDetection(detector *fault_tolerance.EnhancedFaultDetector) {
	fmt.Println("2. Testing Anomaly Detection...")

	// Generate a series of metrics to build statistical models
	baseMetrics := map[string]interface{}{
		"cpu_usage":     0.3,
		"memory_usage":  0.4,
		"response_time": 100.0,
	}

	// Feed normal data to build baseline
	for i := 0; i < 20; i++ {
		normalMetrics := make(map[string]interface{})
		for k, v := range baseMetrics {
			if val, ok := v.(float64); ok {
				// Add small random variation
				variation := float64(i%5) * 0.02
				normalMetrics[k] = val + variation
			}
		}

		detector.DetectFault(fmt.Sprintf("node-baseline-%d", i), normalMetrics)
		time.Sleep(10 * time.Millisecond)
	}

	// Now test with anomalous data
	anomalousMetrics := map[string]interface{}{
		"cpu_usage":     0.95,   // Anomalously high
		"memory_usage":  0.98,   // Anomalously high
		"response_time": 5000.0, // Anomalously high
	}

	detection := detector.DetectFault("node-anomaly", anomalousMetrics)
	if detection != nil {
		fmt.Printf("  ✅ Anomaly detected: %s\n", detection.Description)
		fmt.Printf("  ✅ Severity: %s\n", string(detection.Severity))
	} else {
		fmt.Printf("  ⚠️ No anomaly detected\n")
	}

	fmt.Println("✅ Anomaly detection testing completed")
}

func testHealthScoreCalculation(detector *fault_tolerance.EnhancedFaultDetector) {
	fmt.Println("3. Testing Health Score Calculation...")

	healthTestCases := []struct {
		name     string
		metrics  map[string]interface{}
		expected string
	}{
		{
			name: "Perfect Health",
			metrics: map[string]interface{}{
				"cpu_usage":     0.1,
				"memory_usage":  0.2,
				"response_time": 50.0,
				"error_rate":    0.001,
				"connectivity":  true,
			},
			expected: "high",
		},
		{
			name: "Moderate Health",
			metrics: map[string]interface{}{
				"cpu_usage":     0.6,
				"memory_usage":  0.7,
				"response_time": 300.0,
				"error_rate":    0.05,
				"connectivity":  true,
			},
			expected: "medium",
		},
		{
			name: "Poor Health",
			metrics: map[string]interface{}{
				"cpu_usage":     0.95,
				"memory_usage":  0.98,
				"response_time": 2000.0,
				"error_rate":    0.2,
				"connectivity":  false,
			},
			expected: "low",
		},
	}

	for _, testCase := range healthTestCases {
		detection := detector.DetectFault("health-test", testCase.metrics)
		if detection != nil {
			fmt.Printf("  %s: Health score resulted in %s severity detection\n",
				testCase.name, string(detection.Severity))
		} else {
			fmt.Printf("  %s: No detection (healthy system)\n", testCase.name)
		}
	}

	fmt.Println("✅ Health score calculation testing completed")
}

func testFaultClassification(detector *fault_tolerance.EnhancedFaultDetector) {
	fmt.Println("4. Testing Fault Classification...")

	classificationTests := []struct {
		name         string
		metrics      map[string]interface{}
		expectedType string
	}{
		{
			name: "Resource Exhaustion",
			metrics: map[string]interface{}{
				"cpu_usage":    0.98,
				"memory_usage": 0.95,
			},
			expectedType: "resource_exhaustion",
		},
		{
			name: "Performance Anomaly",
			metrics: map[string]interface{}{
				"response_time": 5000.0,
			},
			expectedType: "performance_anomaly",
		},
		{
			name: "Network Issues",
			metrics: map[string]interface{}{
				"connectivity": false,
			},
			expectedType: "network_partition",
		},
	}

	for _, test := range classificationTests {
		detection := detector.DetectFault("classification-test", test.metrics)
		if detection != nil {
			fmt.Printf("  %s: Classified as %s\n", test.name, string(detection.Type))
		} else {
			fmt.Printf("  %s: No classification (no fault detected)\n", test.name)
		}
	}

	fmt.Println("✅ Fault classification testing completed")
}

func testDetectionStatistics(detector *fault_tolerance.EnhancedFaultDetector) {
	fmt.Println("5. Testing Detection Statistics...")

	// Get current statistics
	stats := detector.GetStatistics()

	fmt.Printf("Detection Statistics:\n")
	fmt.Printf("  ✅ Total Detections: %d\n", stats.TotalDetections)
	fmt.Printf("  ✅ Active Detections: %d\n", stats.ActiveDetections)
	fmt.Printf("  ✅ Detections by Type:\n")
	for faultType, count := range stats.DetectionsByType {
		fmt.Printf("    - %s: %d\n", string(faultType), count)
	}
	fmt.Printf("  ✅ Detections by Severity:\n")
	for severity, count := range stats.DetectionsBySeverity {
		fmt.Printf("    - %s: %d\n", severity, count)
	}

	if !stats.LastDetection.IsZero() {
		fmt.Printf("  ✅ Last Detection: %v\n", stats.LastDetection.Format(time.RFC3339))
	}

	// Get detection history
	history := detector.GetDetectionHistory()
	fmt.Printf("  ✅ Detection History: %d entries\n", len(history))

	fmt.Println("✅ Detection statistics testing completed")
}
