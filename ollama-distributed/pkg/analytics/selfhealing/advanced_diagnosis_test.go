package selfhealing

import (
	"context"
	"testing"
	"time"
)

func TestAdvancedDiagnosisEngine(t *testing.T) {
	// Create diagnosis configuration
	config := &DiagnosisConfig{
		AnalysisTimeout:       time.Minute,
		LogRetentionPeriod:    time.Hour * 24,
		PatternMatchThreshold: 0.8,
		ConfidenceThreshold:   0.6,
		MaxConcurrentAnalysis: 5,
		EnableMLDiagnosis:     true,
		EnableLogAnalysis:     true,
		EnablePatternMatching: true,
	}

	// Create diagnosis engine
	engine, err := NewAdvancedDiagnosisEngine(config)
	if err != nil {
		t.Fatalf("Failed to create diagnosis engine: %v", err)
	}
	defer engine.Stop()

	// Test incident diagnosis
	t.Run("IncidentDiagnosis", func(t *testing.T) {
		testIncidentDiagnosis(t, engine)
	})

	// Test feature extraction
	t.Run("FeatureExtraction", func(t *testing.T) {
		testFeatureExtraction(t, engine)
	})

	// Test log analysis
	t.Run("LogAnalysis", func(t *testing.T) {
		testLogAnalysis(t, engine)
	})

	// Test pattern recognition
	t.Run("PatternRecognition", func(t *testing.T) {
		testPatternRecognition(t, engine)
	})

	// Test root cause analysis
	t.Run("RootCauseAnalysis", func(t *testing.T) {
		testRootCauseAnalysis(t, engine)
	})
}

func testIncidentDiagnosis(t *testing.T, engine *AdvancedDiagnosisEngine) {
	// Create test incident
	incident := &SystemIncident{
		ID:          "incident-1",
		Type:        "performance_degradation",
		Description: "High CPU usage detected",
		Severity:    "high",
		NodeID:      "node-1",
		Symptoms:    []string{"high_cpu", "slow_response"},
		Metrics: map[string]float64{
			"cpu_utilization":    0.95,
			"memory_utilization": 0.7,
			"error_rate":         0.05,
			"response_time":      2.5,
		},
		Logs: []*LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "ERROR",
				Source:    "inference-service",
				Message:   "High CPU usage detected",
				NodeID:    "node-1",
			},
		},
		Events: []*SystemEvent{
			{
				ID:        "event-1",
				Type:      "resource_alert",
				Source:    "monitoring",
				Message:   "CPU threshold exceeded",
				Severity:  "high",
				Timestamp: time.Now(),
			},
		},
		StartTime:  time.Now().Add(-time.Minute * 5),
		DetectedAt: time.Now(),
	}

	// Perform diagnosis
	ctx := context.Background()
	result, err := engine.DiagnoseIncident(ctx, incident)
	if err != nil {
		t.Fatalf("Diagnosis failed: %v", err)
	}

	// Verify diagnosis result
	if result.IncidentID != incident.ID {
		t.Errorf("Expected incident ID %s, got %s", incident.ID, result.IncidentID)
	}

	if result.RootCause == "" {
		t.Error("Expected root cause to be identified")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}

	if len(result.Evidence) == 0 {
		t.Error("Expected evidence to be provided")
	}

	if len(result.RecommendedActions) == 0 {
		t.Error("Expected recommended actions")
	}

	if result.DiagnosisTime <= 0 {
		t.Error("Expected positive diagnosis time")
	}

	// Verify analysis details
	if result.AnalysisDetails == nil {
		t.Error("Expected analysis details")
	} else {
		if result.AnalysisDetails.MLAnalysis == nil {
			t.Error("Expected ML analysis")
		}
		if result.AnalysisDetails.LogAnalysis == nil {
			t.Error("Expected log analysis")
		}
		if result.AnalysisDetails.PatternAnalysis == nil {
			t.Error("Expected pattern analysis")
		}
		if result.AnalysisDetails.RootCauseAnalysis == nil {
			t.Error("Expected root cause analysis")
		}
	}

	// Test CPU exhaustion diagnosis
	if result.RootCause == "cpu_exhaustion" {
		expectedActions := []string{"Scale up CPU resources", "Optimize CPU-intensive processes"}
		for _, expected := range expectedActions {
			found := false
			for _, action := range result.RecommendedActions {
				if action == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected action '%s' not found", expected)
			}
		}
	}
}

func testFeatureExtraction(t *testing.T, engine *AdvancedDiagnosisEngine) {
	// Create test incident
	incident := &SystemIncident{
		ID:       "incident-feature-test",
		Severity: "critical",
		Metrics: map[string]float64{
			"cpu_utilization":    0.9,
			"memory_utilization": 0.8,
			"disk_utilization":   0.7,
			"error_rate":         0.1,
			"response_time":      3.0,
			"throughput":         500.0,
		},
		Symptoms:   []string{"high_cpu", "slow_response", "errors"},
		Logs:       []*LogEntry{{Level: "ERROR"}, {Level: "INFO"}, {Level: "ERROR"}},
		Events:     []*SystemEvent{{}, {}},
		DetectedAt: time.Now(),
	}

	// Extract features
	features := engine.diagnosticModel.featureExtractor.ExtractFeatures(incident)

	// Verify feature count
	expectedFeatureCount := len(engine.diagnosticModel.featureExtractor.featureNames)
	if len(features) != expectedFeatureCount {
		t.Errorf("Expected %d features, got %d", expectedFeatureCount, len(features))
	}

	// Verify feature values are reasonable
	for i, feature := range features {
		if feature < 0 || feature > 1 {
			featureName := ""
			if i < len(engine.diagnosticModel.featureExtractor.featureNames) {
				featureName = engine.diagnosticModel.featureExtractor.featureNames[i]
			}
			t.Errorf("Feature %d (%s) out of range [0,1]: %f", i, featureName, feature)
		}
	}

	// Test specific feature values
	featureMap := make(map[string]float64)
	for i, name := range engine.diagnosticModel.featureExtractor.featureNames {
		if i < len(features) {
			featureMap[name] = features[i]
		}
	}

	// CPU utilization should be high
	if cpu, exists := featureMap["cpu_utilization"]; exists {
		if cpu != 0.9 {
			t.Errorf("Expected CPU utilization 0.9, got %f", cpu)
		}
	}

	// Severity score should be 1.0 for critical
	if severity, exists := featureMap["severity_score"]; exists {
		if severity != 1.0 {
			t.Errorf("Expected severity score 1.0 for critical, got %f", severity)
		}
	}
}

func testLogAnalysis(t *testing.T, engine *AdvancedDiagnosisEngine) {
	// Create test logs
	logs := []*LogEntry{
		{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Source:    "service-a",
			Message:   "Connection timeout",
			NodeID:    "node-1",
		},
		{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Source:    "service-a",
			Message:   "Connection timeout",
			NodeID:    "node-1",
		},
		{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Source:    "service-b",
			Message:   "Memory allocation failed",
			NodeID:    "node-1",
		},
		{
			Timestamp: time.Now(),
			Level:     "INFO",
			Source:    "service-a",
			Message:   "Processing request",
			NodeID:    "node-1",
		},
	}

	// Analyze logs
	ctx := context.Background()
	result, err := engine.logAnalyzer.AnalyzeLogs(ctx, logs)
	if err != nil {
		t.Fatalf("Log analysis failed: %v", err)
	}

	// Verify analysis result
	if result.ProcessedLogs != len(logs) {
		t.Errorf("Expected %d processed logs, got %d", len(logs), result.ProcessedLogs)
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}

	// Should detect high error rate
	if len(result.Anomalies) == 0 {
		t.Error("Expected anomalies to be detected")
	} else {
		anomaly := result.Anomalies[0]
		if anomaly.Type != "high_error_rate" {
			t.Errorf("Expected high_error_rate anomaly, got %s", anomaly.Type)
		}
	}

	// Should find patterns
	if len(result.Patterns) == 0 {
		t.Error("Expected patterns to be found")
	}

	// Should find correlations
	if len(result.Correlations) == 0 {
		t.Error("Expected correlations to be found")
	}

	// Should generate insights
	if len(result.Insights) == 0 {
		t.Error("Expected insights to be generated")
	}
}

func testPatternRecognition(t *testing.T, engine *AdvancedDiagnosisEngine) {
	// Create test incident
	incident := &SystemIncident{
		ID:       "incident-pattern-test",
		Symptoms: []string{"high_cpu", "memory_leak", "slow_response"},
	}

	// Analyze patterns
	ctx := context.Background()
	result, err := engine.patternRecognizer.AnalyzePatterns(ctx, incident)
	if err != nil {
		t.Fatalf("Pattern analysis failed: %v", err)
	}

	// Verify pattern analysis result
	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}

	// Matched patterns should be empty initially (no patterns in knowledge base)
	if len(result.MatchedPatterns) != 0 {
		t.Errorf("Expected no matched patterns initially, got %d", len(result.MatchedPatterns))
	}

	// New patterns should be empty (not implemented in this version)
	if len(result.NewPatterns) != 0 {
		t.Errorf("Expected no new patterns, got %d", len(result.NewPatterns))
	}
}

func testRootCauseAnalysis(t *testing.T, engine *AdvancedDiagnosisEngine) {
	// Create test incident with high CPU
	incident := &SystemIncident{
		ID: "incident-rootcause-test",
		Metrics: map[string]float64{
			"cpu_utilization":    0.95,
			"memory_utilization": 0.6,
			"disk_utilization":   0.5,
			"error_rate":         0.02,
		},
	}

	// Analyze root cause
	ctx := context.Background()
	result, err := engine.rootCauseAnalyzer.AnalyzeRootCause(ctx, incident)
	if err != nil {
		t.Fatalf("Root cause analysis failed: %v", err)
	}

	// Verify root cause analysis result
	if result.PrimaryCause == "" {
		t.Error("Expected primary cause to be identified")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("Invalid confidence: %f", result.Confidence)
	}

	// Should identify CPU exhaustion
	if result.PrimaryCause != "cpu_exhaustion" {
		t.Errorf("Expected cpu_exhaustion, got %s", result.PrimaryCause)
	}

	// Should have alternative causes
	if len(result.AlternativeCauses) == 0 {
		t.Error("Expected alternative causes")
	}

	// Verify alternative cause structure
	for _, alt := range result.AlternativeCauses {
		if alt.Cause == "" {
			t.Error("Alternative cause should have a name")
		}
		if alt.Confidence < 0 || alt.Confidence > 1 {
			t.Errorf("Invalid alternative cause confidence: %f", alt.Confidence)
		}
	}
}

func TestDiagnosticFeatureExtractor(t *testing.T) {
	extractor := NewDiagnosticFeatureExtractor()

	// Test feature names
	if len(extractor.featureNames) == 0 {
		t.Error("Expected feature names to be defined")
	}

	// Test extractors
	if len(extractor.extractors) == 0 {
		t.Error("Expected extractors to be defined")
	}

	// Verify all feature names have extractors
	for _, name := range extractor.featureNames {
		if _, exists := extractor.extractors[name]; !exists {
			t.Errorf("Missing extractor for feature: %s", name)
		}
	}

	// Test feature extraction with empty incident
	emptyIncident := &SystemIncident{
		Metrics:    make(map[string]float64),
		Symptoms:   []string{},
		Logs:       []*LogEntry{},
		Events:     []*SystemEvent{},
		DetectedAt: time.Now(),
	}

	features := extractor.ExtractFeatures(emptyIncident)
	if len(features) != len(extractor.featureNames) {
		t.Errorf("Expected %d features, got %d", len(extractor.featureNames), len(features))
	}

	// All features should be valid (0-1 range)
	for i, feature := range features {
		if feature < 0 || feature > 1 {
			t.Errorf("Feature %d out of range: %f", i, feature)
		}
	}
}

func TestLogAnalyzer(t *testing.T) {
	analyzer := NewLogAnalyzer()

	// Test empty logs
	emptyResult, err := analyzer.AnalyzeLogs(context.Background(), []*LogEntry{})
	if err != nil {
		t.Fatalf("Empty log analysis failed: %v", err)
	}

	if emptyResult.Confidence != 0.0 {
		t.Errorf("Expected 0 confidence for empty logs, got %f", emptyResult.Confidence)
	}

	// Test normal logs
	normalLogs := []*LogEntry{
		{Level: "INFO", Message: "Normal operation"},
		{Level: "INFO", Message: "Processing request"},
		{Level: "DEBUG", Message: "Debug info"},
	}

	normalResult, err := analyzer.AnalyzeLogs(context.Background(), normalLogs)
	if err != nil {
		t.Fatalf("Normal log analysis failed: %v", err)
	}

	if len(normalResult.Anomalies) != 0 {
		t.Error("Expected no anomalies for normal logs")
	}
}

func BenchmarkDiagnosis(b *testing.B) {
	config := &DiagnosisConfig{
		AnalysisTimeout:       time.Minute,
		ConfidenceThreshold:   0.6,
		MaxConcurrentAnalysis: 5,
		EnableMLDiagnosis:     true,
		EnableLogAnalysis:     false, // Disable for benchmark
		EnablePatternMatching: false, // Disable for benchmark
	}

	engine, err := NewAdvancedDiagnosisEngine(config)
	if err != nil {
		b.Fatalf("Failed to create diagnosis engine: %v", err)
	}
	defer engine.Stop()

	incident := &SystemIncident{
		ID:       "bench-incident",
		Severity: "high",
		Metrics: map[string]float64{
			"cpu_utilization":    0.9,
			"memory_utilization": 0.8,
			"error_rate":         0.1,
		},
		Symptoms:   []string{"high_cpu"},
		DetectedAt: time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.DiagnoseIncident(context.Background(), incident)
		if err != nil {
			b.Fatalf("Diagnosis failed: %v", err)
		}
	}
}
