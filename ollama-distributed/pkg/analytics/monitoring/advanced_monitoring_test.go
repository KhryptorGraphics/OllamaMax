package monitoring

import (
	"testing"
	"time"
)

func TestAdvancedMonitoringEngine(t *testing.T) {
	// Create monitoring configuration
	config := &MonitoringConfig{
		CollectionInterval:      time.Second * 30,
		AlertEvaluationInterval: time.Second * 10,
		MetricsRetention:        time.Hour * 24,
		AlertRetention:          time.Hour * 48,
		MaxAlertsPerMinute:      100,
		EnablePredictiveAlerts:  true,
		DashboardRefreshRate:    time.Minute,
		IntegrationTimeout:      time.Second * 30,
	}

	// Create monitoring engine
	engine, err := NewAdvancedMonitoringEngine(config)
	if err != nil {
		t.Fatalf("Failed to create monitoring engine: %v", err)
	}
	defer engine.Stop()

	// Test alert creation
	t.Run("AlertCreation", func(t *testing.T) {
		testAlertCreation(t, engine)
	})

	// Test alert prioritization
	t.Run("AlertPrioritization", func(t *testing.T) {
		testAlertPrioritization(t, engine)
	})

	// Test alert correlation
	t.Run("AlertCorrelation", func(t *testing.T) {
		testAlertCorrelation(t, engine)
	})

	// Test fatigue reduction
	t.Run("FatigueReduction", func(t *testing.T) {
		testFatigueReduction(t, engine)
	})

	// Test predictive insights
	t.Run("PredictiveInsights", func(t *testing.T) {
		testPredictiveInsights(t, engine)
	})
}

func testAlertCreation(t *testing.T, engine *AdvancedMonitoringEngine) {
	// Create an alert rule
	rule := &AlertRule{
		ID:          "cpu-high",
		Name:        "High CPU Usage",
		Description: "CPU utilization is above threshold",
		MetricName:  "cpu_utilization",
		Condition:   "gt",
		Threshold:   0.8,
		Duration:    time.Minute * 5,
		Severity:    "high",
		Enabled:     true,
		Tags:        []string{"cpu", "performance"},
		Actions:     []string{"scale_up", "notify"},
	}

	// Add rule to alert manager
	engine.alertManager.alertRules[rule.ID] = rule

	// Create a metric that triggers the alert
	metric := &Metric{
		Name:      "cpu_utilization",
		Value:     0.9, // Above threshold
		Timestamp: time.Now(),
		Labels: map[string]string{
			"node_id": "test-node-1",
		},
		Metadata: make(map[string]interface{}),
	}

	// Create alert
	alert, err := engine.CreateAlert(rule.ID, metric)
	if err != nil {
		t.Fatalf("Failed to create alert: %v", err)
	}

	// Verify alert properties
	if alert.RuleID != rule.ID {
		t.Errorf("Expected rule ID %s, got %s", rule.ID, alert.RuleID)
	}

	if alert.Severity != rule.Severity {
		t.Errorf("Expected severity %s, got %s", rule.Severity, alert.Severity)
	}

	if alert.Status != "active" {
		t.Errorf("Expected status 'active', got %s", alert.Status)
	}

	if alert.MetricName != metric.Name {
		t.Errorf("Expected metric name %s, got %s", metric.Name, alert.MetricName)
	}

	if alert.MetricValue != metric.Value {
		t.Errorf("Expected metric value %f, got %f", metric.Value, alert.MetricValue)
	}

	if alert.NodeID != metric.Labels["node_id"] {
		t.Errorf("Expected node ID %s, got %s", metric.Labels["node_id"], alert.NodeID)
	}

	if len(alert.RecommendedActions) == 0 {
		t.Error("Expected recommended actions")
	}

	if alert.Priority < 0 || alert.Priority > 1 {
		t.Errorf("Invalid priority: %f", alert.Priority)
	}
}

func testAlertPrioritization(t *testing.T, engine *AdvancedMonitoringEngine) {
	prioritizer := engine.alertManager.prioritizer

	// Test critical alert priority
	criticalAlert := &Alert{
		Severity:    "critical",
		MetricValue: 0.95,
		Threshold:   0.8,
		CreatedAt:   time.Now(),
	}

	priority := prioritizer.CalculatePriority(criticalAlert)
	if priority < 0.6 { // Critical alerts should have high priority
		t.Errorf("Critical alert priority too low: %f", priority)
	}

	// Test low severity alert priority
	lowAlert := &Alert{
		Severity:    "low",
		MetricValue: 0.3,
		Threshold:   0.5,
		CreatedAt:   time.Now(),
	}

	priority = prioritizer.CalculatePriority(lowAlert)
	if priority > 0.6 { // Low alerts should have lower priority
		t.Errorf("Low alert priority too high: %f", priority)
	}

	// Test feature extraction
	features := prioritizer.features.ExtractFeatures(criticalAlert)
	if len(features) != 6 { // Should have 6 features
		t.Errorf("Expected 6 features, got %d", len(features))
	}

	// Verify severity feature
	if features[0] != 1.0 { // Critical severity should be 1.0
		t.Errorf("Expected critical severity feature 1.0, got %f", features[0])
	}
}

func testAlertCorrelation(t *testing.T, engine *AdvancedMonitoringEngine) {
	correlationEngine := engine.alertManager.correlationEngine

	// Create test alert
	alert := &Alert{
		ID:         "test-alert-1",
		NodeID:     "test-node-1",
		MetricName: "cpu_utilization",
		CreatedAt:  time.Now(),
	}

	// Find correlations (should return empty for now)
	correlations := correlationEngine.FindCorrelations(alert)
	if len(correlations) != 0 {
		t.Errorf("Expected no correlations, got %d", len(correlations))
	}

	// Verify correlation engine initialization
	if correlationEngine.correlationRules == nil {
		t.Error("Correlation rules not initialized")
	}

	if correlationEngine.activeGroups == nil {
		t.Error("Active groups not initialized")
	}
}

func testFatigueReduction(t *testing.T, engine *AdvancedMonitoringEngine) {
	fatigueReducer := engine.alertManager.fatigueReducer

	// Create test alert
	alert := &Alert{
		ID:         "test-alert-fatigue",
		NodeID:     "test-node-fatigue",
		MetricName: "memory_utilization",
		CreatedAt:  time.Now(),
	}

	// First alert should not be suppressed
	shouldSuppress := fatigueReducer.ShouldSuppress(alert)
	if shouldSuppress {
		t.Error("First alert should not be suppressed")
	}

	// Verify fatigue metrics initialization
	if fatigueReducer.fatigueMetrics == nil {
		t.Error("Fatigue metrics not initialized")
	}

	if fatigueReducer.suppressionRules == nil {
		t.Error("Suppression rules not initialized")
	}

	if fatigueReducer.suppressedAlerts == nil {
		t.Error("Suppressed alerts not initialized")
	}
}

func testPredictiveInsights(t *testing.T, engine *AdvancedMonitoringEngine) {
	// Get predictive insights
	insights, err := engine.GetPredictiveInsights()
	if err != nil {
		t.Fatalf("Failed to get predictive insights: %v", err)
	}

	// Verify insights structure
	if insights.PredictedFailures == nil {
		t.Error("Predicted failures should not be nil")
	}

	if insights.RecommendedActions == nil {
		t.Error("Recommended actions should not be nil")
	}

	if insights.ConfidenceScore < 0 || insights.ConfidenceScore > 1 {
		t.Errorf("Invalid confidence score: %f", insights.ConfidenceScore)
	}

	if insights.TimeHorizon <= 0 {
		t.Errorf("Invalid time horizon: %v", insights.TimeHorizon)
	}
}

func TestIntelligentAlertManager(t *testing.T) {
	alertManager := NewIntelligentAlertManager()

	// Test initialization
	if alertManager.alertRules == nil {
		t.Error("Alert rules not initialized")
	}

	if alertManager.activeAlerts == nil {
		t.Error("Active alerts not initialized")
	}

	if alertManager.prioritizer == nil {
		t.Error("Prioritizer not initialized")
	}

	if alertManager.correlationEngine == nil {
		t.Error("Correlation engine not initialized")
	}

	if alertManager.fatigueReducer == nil {
		t.Error("Fatigue reducer not initialized")
	}

	// Test getting active alerts (should be empty initially)
	alerts := alertManager.GetActiveAlerts()
	if len(alerts) != 0 {
		t.Errorf("Expected no active alerts, got %d", len(alerts))
	}
}

func TestAlertPrioritizer(t *testing.T) {
	prioritizer := NewAlertPrioritizer()

	// Test model initialization
	if prioritizer.model == nil {
		t.Error("Priority model not initialized")
	}

	if len(prioritizer.model.weights) != 6 {
		t.Errorf("Expected 6 weights, got %d", len(prioritizer.model.weights))
	}

	if prioritizer.features == nil {
		t.Error("Feature extractor not initialized")
	}

	if len(prioritizer.features.featureNames) != 6 {
		t.Errorf("Expected 6 feature names, got %d", len(prioritizer.features.featureNames))
	}

	// Test priority calculation
	testAlert := &Alert{
		Severity:    "high",
		MetricValue: 0.8,
		Threshold:   0.7,
		CreatedAt:   time.Now(),
	}

	priority := prioritizer.CalculatePriority(testAlert)
	if priority < 0 || priority > 1 {
		t.Errorf("Invalid priority: %f", priority)
	}
}

func TestDashboardEngine(t *testing.T) {
	dashboardEngine := NewDashboardEngine()

	// Test initialization
	if dashboardEngine.dashboards == nil {
		t.Error("Dashboards not initialized")
	}

	if dashboardEngine.widgets == nil {
		t.Error("Widgets not initialized")
	}

	if dashboardEngine.dataProviders == nil {
		t.Error("Data providers not initialized")
	}
}

func TestEnhancedMetricsCollector(t *testing.T) {
	collector := NewEnhancedMetricsCollector()

	// Test initialization
	if collector.collectors == nil {
		t.Error("Collectors not initialized")
	}

	if collector.processors == nil {
		t.Error("Processors not initialized")
	}

	if collector.aggregators == nil {
		t.Error("Aggregators not initialized")
	}
}

func TestIntegrationManager(t *testing.T) {
	integrationManager := NewIntegrationManager()

	// Test initialization
	if integrationManager.integrations == nil {
		t.Error("Integrations not initialized")
	}

	if integrationManager.config == nil {
		t.Error("Integration config not initialized")
	}

	// Test default configuration
	config := integrationManager.config
	if config.PrometheusEnabled {
		t.Error("Prometheus should be disabled by default")
	}

	if config.GrafanaEnabled {
		t.Error("Grafana should be disabled by default")
	}

	if config.PagerDutyEnabled {
		t.Error("PagerDuty should be disabled by default")
	}

	if config.SlackEnabled {
		t.Error("Slack should be disabled by default")
	}
}

func TestAlertRecommendations(t *testing.T) {
	alertManager := NewIntelligentAlertManager()

	// Test CPU utilization recommendations
	rule := &AlertRule{
		Name: "CPU Alert",
	}

	metric := &Metric{
		Name: "cpu_utilization",
	}

	actions := alertManager.generateRecommendedActions(rule, metric)
	if len(actions) == 0 {
		t.Error("Expected recommended actions for CPU utilization")
	}

	expectedActions := []string{"Scale up CPU resources", "Optimize CPU-intensive processes"}
	for _, expected := range expectedActions {
		found := false
		for _, action := range actions {
			if action == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected action '%s' not found", expected)
		}
	}

	// Test memory utilization recommendations
	metric.Name = "memory_utilization"
	actions = alertManager.generateRecommendedActions(rule, metric)
	if len(actions) == 0 {
		t.Error("Expected recommended actions for memory utilization")
	}

	// Test unknown metric recommendations
	metric.Name = "unknown_metric"
	actions = alertManager.generateRecommendedActions(rule, metric)
	if len(actions) == 0 {
		t.Error("Expected default recommended actions for unknown metric")
	}
}

func BenchmarkAlertCreation(b *testing.B) {
	config := &MonitoringConfig{
		CollectionInterval:      time.Minute,
		AlertEvaluationInterval: time.Minute,
		MetricsRetention:        time.Hour,
		AlertRetention:          time.Hour,
		MaxAlertsPerMinute:      1000,
		EnablePredictiveAlerts:  false, // Disable for benchmark
		DashboardRefreshRate:    time.Minute,
		IntegrationTimeout:      time.Second * 30,
	}

	engine, err := NewAdvancedMonitoringEngine(config)
	if err != nil {
		b.Fatalf("Failed to create monitoring engine: %v", err)
	}
	defer engine.Stop()

	// Add test rule
	rule := &AlertRule{
		ID:        "bench-rule",
		Name:      "Benchmark Rule",
		Threshold: 0.8,
		Severity:  "high",
		Enabled:   true,
	}
	engine.alertManager.alertRules[rule.ID] = rule

	metric := &Metric{
		Name:      "cpu_utilization",
		Value:     0.9,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"node_id": "bench-node",
		},
		Metadata: make(map[string]interface{}),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := engine.CreateAlert(rule.ID, metric)
		if err != nil {
			b.Fatalf("Alert creation failed: %v", err)
		}
	}
}
