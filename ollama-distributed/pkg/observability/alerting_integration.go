package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AlertingIntegration integrates alerting with metrics and health systems
type AlertingIntegration struct {
	metricsRegistry *MetricsRegistry
	healthManager   *HealthCheckManager
	notificationSys *NotificationSystem
	dashboard       *MonitoringDashboard
	
	// Alert rules and thresholds
	alertRules    map[string]*AlertRule
	alertHistory  map[string]*AlertHistory
	
	// Configuration
	config *AlertingConfig
	
	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// AlertingConfig configures the alerting integration
type AlertingConfig struct {
	Enabled           bool          `json:"enabled"`
	EvaluationInterval time.Duration `json:"evaluation_interval"`
	AlertTimeout      time.Duration `json:"alert_timeout"`
	
	// Thresholds
	HighErrorRateThreshold    float64 `json:"high_error_rate_threshold"`
	HighLatencyThreshold      float64 `json:"high_latency_threshold"`
	LowConnectivityThreshold  int     `json:"low_connectivity_threshold"`
	HighMemoryThreshold       float64 `json:"high_memory_threshold"`
	HighCPUThreshold          float64 `json:"high_cpu_threshold"`
	
	// Rate limiting
	AlertCooldown time.Duration `json:"alert_cooldown"`
}

// AlertRule defines an alerting rule
type AlertRule struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Query       string                 `json:"query"`
	Threshold   float64                `json:"threshold"`
	Operator    string                 `json:"operator"` // >, <, >=, <=, ==, !=
	Duration    time.Duration          `json:"duration"`
	Severity    string                 `json:"severity"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Enabled     bool                   `json:"enabled"`
	LastFired   time.Time              `json:"last_fired"`
}

// AlertHistory tracks alert firing history
type AlertHistory struct {
	RuleName    string    `json:"rule_name"`
	FiredAt     time.Time `json:"fired_at"`
	ResolvedAt  time.Time `json:"resolved_at"`
	Duration    time.Duration `json:"duration"`
	Status      string    `json:"status"` // firing, resolved
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
}

// NewAlertingIntegration creates a new alerting integration
func NewAlertingIntegration(
	config *AlertingConfig,
	metricsRegistry *MetricsRegistry,
	healthManager *HealthCheckManager,
	notificationSys *NotificationSystem,
	dashboard *MonitoringDashboard,
) *AlertingIntegration {
	if config == nil {
		config = DefaultAlertingConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	ai := &AlertingIntegration{
		metricsRegistry: metricsRegistry,
		healthManager:   healthManager,
		notificationSys: notificationSys,
		dashboard:       dashboard,
		config:          config,
		alertRules:      make(map[string]*AlertRule),
		alertHistory:    make(map[string]*AlertHistory),
		ctx:             ctx,
		cancel:          cancel,
	}
	
	// Initialize default alert rules
	ai.initializeDefaultRules()
	
	return ai
}

// initializeDefaultRules sets up default alerting rules
func (ai *AlertingIntegration) initializeDefaultRules() {
	// High error rate alert
	ai.alertRules["high_error_rate"] = &AlertRule{
		Name:        "HighErrorRate",
		Description: "High error rate detected",
		Threshold:   ai.config.HighErrorRateThreshold,
		Operator:    ">",
		Duration:    5 * time.Minute,
		Severity:    "warning",
		Enabled:     true,
		Labels: map[string]string{
			"component": "api",
			"type":      "error_rate",
		},
		Annotations: map[string]string{
			"summary":     "High error rate detected",
			"description": "Error rate is above threshold",
		},
	}
	
	// High latency alert
	ai.alertRules["high_latency"] = &AlertRule{
		Name:        "HighLatency",
		Description: "High latency detected",
		Threshold:   ai.config.HighLatencyThreshold,
		Operator:    ">",
		Duration:    5 * time.Minute,
		Severity:    "warning",
		Enabled:     true,
		Labels: map[string]string{
			"component": "api",
			"type":      "latency",
		},
		Annotations: map[string]string{
			"summary":     "High latency detected",
			"description": "Response latency is above threshold",
		},
	}
	
	// Low P2P connectivity alert
	ai.alertRules["low_connectivity"] = &AlertRule{
		Name:        "LowP2PConnectivity",
		Description: "Low P2P connectivity",
		Threshold:   float64(ai.config.LowConnectivityThreshold),
		Operator:    "<",
		Duration:    2 * time.Minute,
		Severity:    "warning",
		Enabled:     true,
		Labels: map[string]string{
			"component": "p2p",
			"type":      "connectivity",
		},
		Annotations: map[string]string{
			"summary":     "Low P2P connectivity",
			"description": "Number of P2P connections is below threshold",
		},
	}
	
	// Service health alert
	ai.alertRules["service_unhealthy"] = &AlertRule{
		Name:        "ServiceUnhealthy",
		Description: "Service health check failing",
		Threshold:   0,
		Operator:    "==",
		Duration:    1 * time.Minute,
		Severity:    "critical",
		Enabled:     true,
		Labels: map[string]string{
			"component": "health",
			"type":      "service_health",
		},
		Annotations: map[string]string{
			"summary":     "Service is unhealthy",
			"description": "Service health check is failing",
		},
	}
	
	log.Info().
		Int("rules_count", len(ai.alertRules)).
		Msg("Default alert rules initialized")
}

// Start starts the alerting integration
func (ai *AlertingIntegration) Start() error {
	if !ai.config.Enabled {
		log.Info().Msg("Alerting integration disabled")
		return nil
	}
	
	// Start alert evaluation loop
	go ai.evaluationLoop()
	
	log.Info().
		Dur("evaluation_interval", ai.config.EvaluationInterval).
		Msg("Alerting integration started")
	
	return nil
}

// evaluationLoop continuously evaluates alert rules
func (ai *AlertingIntegration) evaluationLoop() {
	ticker := time.NewTicker(ai.config.EvaluationInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ai.ctx.Done():
			return
		case <-ticker.C:
			ai.evaluateRules()
		}
	}
}

// evaluateRules evaluates all enabled alert rules
func (ai *AlertingIntegration) evaluateRules() {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	
	for ruleName, rule := range ai.alertRules {
		if !rule.Enabled {
			continue
		}
		
		// Check cooldown period
		if time.Since(rule.LastFired) < ai.config.AlertCooldown {
			continue
		}
		
		// Evaluate rule
		shouldFire, value := ai.evaluateRule(rule)
		
		if shouldFire {
			ai.fireAlert(ruleName, rule, value)
		}
	}
}

// evaluateRule evaluates a single alert rule
func (ai *AlertingIntegration) evaluateRule(rule *AlertRule) (bool, float64) {
	var value float64
	
	switch rule.Labels["type"] {
	case "error_rate":
		value = ai.getErrorRate()
	case "latency":
		value = ai.getLatency()
	case "connectivity":
		value = ai.getP2PConnectivity()
	case "service_health":
		value = ai.getServiceHealth()
	default:
		return false, 0
	}
	
	// Evaluate threshold
	switch rule.Operator {
	case ">":
		return value > rule.Threshold, value
	case "<":
		return value < rule.Threshold, value
	case ">=":
		return value >= rule.Threshold, value
	case "<=":
		return value <= rule.Threshold, value
	case "==":
		return value == rule.Threshold, value
	case "!=":
		return value != rule.Threshold, value
	default:
		return false, value
	}
}

// getErrorRate calculates current error rate
func (ai *AlertingIntegration) getErrorRate() float64 {
	if ai.metricsRegistry == nil {
		return 0
	}
	
	// Get error rate from metrics
	totalRequests := ai.metricsRegistry.GetMetricValue("ollama_distributed_api_requests_total")
	errorRequests := ai.metricsRegistry.GetMetricValue("ollama_distributed_api_requests_total{status=~\"5..\"}")
	
	if totalRequests > 0 {
		return errorRequests / totalRequests
	}
	
	return 0
}

// getLatency calculates current latency
func (ai *AlertingIntegration) getLatency() float64 {
	if ai.metricsRegistry == nil {
		return 0
	}
	
	// Get 95th percentile latency
	return ai.metricsRegistry.GetMetricValue("ollama_distributed_api_request_duration_seconds{quantile=\"0.95\"}")
}

// getP2PConnectivity gets current P2P connectivity
func (ai *AlertingIntegration) getP2PConnectivity() float64 {
	if ai.metricsRegistry == nil {
		return 0
	}
	
	return ai.metricsRegistry.GetMetricValue("ollama_distributed_p2p_connections_active")
}

// getServiceHealth gets overall service health
func (ai *AlertingIntegration) getServiceHealth() float64 {
	if ai.healthManager == nil {
		return 1 // Assume healthy if no health manager
	}
	
	health := ai.healthManager.GetOverallHealth()
	if health.Status == "healthy" {
		return 1
	}
	
	return 0
}

// fireAlert fires an alert
func (ai *AlertingIntegration) fireAlert(ruleName string, rule *AlertRule, value float64) {
	now := time.Now()
	
	// Update rule last fired time
	rule.LastFired = now
	
	// Create alert history entry
	history := &AlertHistory{
		RuleName:  ruleName,
		FiredAt:   now,
		Status:    "firing",
		Value:     value,
		Threshold: rule.Threshold,
	}
	ai.alertHistory[fmt.Sprintf("%s-%d", ruleName, now.Unix())] = history
	
	// Create notification
	notification := &Notification{
		ID:        fmt.Sprintf("alert-%s-%d", ruleName, now.Unix()),
		Title:     rule.Annotations["summary"],
		Message:   fmt.Sprintf("%s: Current value %.2f, threshold %.2f", rule.Annotations["description"], value, rule.Threshold),
		Severity:  rule.Severity,
		Component: rule.Labels["component"],
		Timestamp: now,
		Labels:    rule.Labels,
		Annotations: rule.Annotations,
		Metadata: map[string]interface{}{
			"rule_name": ruleName,
			"value":     value,
			"threshold": rule.Threshold,
			"operator":  rule.Operator,
		},
	}
	
	// Send notification
	if ai.notificationSys != nil {
		if err := ai.notificationSys.SendNotification(notification); err != nil {
			log.Error().
				Err(err).
				Str("rule_name", ruleName).
				Msg("Failed to send alert notification")
		}
	}
	
	log.Warn().
		Str("rule_name", ruleName).
		Float64("value", value).
		Float64("threshold", rule.Threshold).
		Str("severity", rule.Severity).
		Msg("Alert fired")
}

// AddRule adds a custom alert rule
func (ai *AlertingIntegration) AddRule(name string, rule *AlertRule) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	
	ai.alertRules[name] = rule
	log.Info().
		Str("rule_name", name).
		Msg("Alert rule added")
}

// RemoveRule removes an alert rule
func (ai *AlertingIntegration) RemoveRule(name string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	
	delete(ai.alertRules, name)
	log.Info().
		Str("rule_name", name).
		Msg("Alert rule removed")
}

// GetAlertHistory returns alert history
func (ai *AlertingIntegration) GetAlertHistory() map[string]*AlertHistory {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	
	// Return a copy
	history := make(map[string]*AlertHistory)
	for k, v := range ai.alertHistory {
		history[k] = v
	}
	
	return history
}

// Shutdown gracefully shuts down the alerting integration
func (ai *AlertingIntegration) Shutdown() error {
	ai.cancel()
	log.Info().Msg("Alerting integration stopped")
	return nil
}

// DefaultAlertingConfig returns a default alerting configuration
func DefaultAlertingConfig() *AlertingConfig {
	return &AlertingConfig{
		Enabled:                   true,
		EvaluationInterval:        30 * time.Second,
		AlertTimeout:              5 * time.Minute,
		HighErrorRateThreshold:    0.05, // 5%
		HighLatencyThreshold:      10.0, // 10 seconds
		LowConnectivityThreshold:  2,    // 2 connections
		HighMemoryThreshold:       0.85, // 85%
		HighCPUThreshold:          0.80, // 80%
		AlertCooldown:             5 * time.Minute,
	}
}
