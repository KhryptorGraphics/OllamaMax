package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// SecurityMonitor monitors security events and detects threats
type SecurityMonitor struct {
	config         *MonitoringConfig
	eventCollector *SecurityEventCollector
	threatDetector *ThreatDetector
	alertManager   *SecurityAlertManager

	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// MonitoringConfig configures security monitoring
type MonitoringConfig struct {
	Enabled                bool          `json:"enabled"`
	CollectionInterval     time.Duration `json:"collection_interval"`
	ThreatDetectionLevel   string        `json:"threat_detection_level"`
	AlertThreshold         int           `json:"alert_threshold"`
	RetentionPeriod        time.Duration `json:"retention_period"`
	EnableRealTimeAlerts   bool          `json:"enable_real_time_alerts"`
	EnableAnomalyDetection bool          `json:"enable_anomaly_detection"`
	EnableBehaviorAnalysis bool          `json:"enable_behavior_analysis"`
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        SecurityEventType      `json:"type"`
	Severity    SeverityLevel          `json:"severity"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	EventTypeAuthentication SecurityEventType = "authentication"
	EventTypeAuthorization  SecurityEventType = "authorization"
	EventTypeDataAccess     SecurityEventType = "data_access"
	EventTypeNetworkAccess  SecurityEventType = "network_access"
	EventTypeSystemAccess   SecurityEventType = "system_access"
	EventTypeConfigChange   SecurityEventType = "config_change"
	EventTypeAnomalous      SecurityEventType = "anomalous"
	EventTypeIntrusion      SecurityEventType = "intrusion"
)

// ThreatIndicator represents a potential security threat
type ThreatIndicator struct {
	ID          string                 `json:"id"`
	Type        ThreatType             `json:"type"`
	Severity    SeverityLevel          `json:"severity"`
	Confidence  float64                `json:"confidence"`
	Description string                 `json:"description"`
	Evidence    []SecurityEvent        `json:"evidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	DetectedAt  time.Time              `json:"detected_at"`
	Status      ThreatStatus           `json:"status"`
}

// ThreatType represents the type of threat
type ThreatType string

const (
	ThreatTypeBruteForce          ThreatType = "brute_force"
	ThreatTypeAnomalousLogin      ThreatType = "anomalous_login"
	ThreatTypeDataExfiltration    ThreatType = "data_exfiltration"
	ThreatTypePrivilegeEscalation ThreatType = "privilege_escalation"
	ThreatTypeDDoS                ThreatType = "ddos"
	ThreatTypeMalware             ThreatType = "malware"
	ThreatTypeInsiderThreat       ThreatType = "insider_threat"
)

// ThreatStatus represents the status of a threat
type ThreatStatus string

const (
	ThreatStatusActive        ThreatStatus = "active"
	ThreatStatusInvestigating ThreatStatus = "investigating"
	ThreatStatusMitigated     ThreatStatus = "mitigated"
	ThreatStatusFalsePositive ThreatStatus = "false_positive"
)

// SecurityEventCollector collects security events
type SecurityEventCollector struct {
	events    []SecurityEvent
	eventChan chan SecurityEvent
	mu        sync.RWMutex
}

// ThreatDetector detects security threats from events
type ThreatDetector struct {
	config     *MonitoringConfig
	indicators map[string]*ThreatIndicator
	patterns   map[ThreatType]*ThreatPattern
	mu         sync.RWMutex
}

// ThreatPattern defines patterns for threat detection
type ThreatPattern struct {
	Type       ThreatType          `json:"type"`
	EventTypes []SecurityEventType `json:"event_types"`
	TimeWindow time.Duration       `json:"time_window"`
	Threshold  int                 `json:"threshold"`
	Conditions []string            `json:"conditions"`
}

// SecurityAlertManager manages security alerts
type SecurityAlertManager struct {
	alerts    []SecurityAlert
	alertChan chan SecurityAlert
	mu        sync.RWMutex
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Severity    SeverityLevel          `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Indicators  []ThreatIndicator      `json:"indicators"`
	Metadata    map[string]interface{} `json:"metadata"`
	Status      string                 `json:"status"`
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(config *MonitoringConfig) *SecurityMonitor {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	sm := &SecurityMonitor{
		config: config,
		eventCollector: &SecurityEventCollector{
			events:    make([]SecurityEvent, 0),
			eventChan: make(chan SecurityEvent, 1000),
		},
		threatDetector: &ThreatDetector{
			config:     config,
			indicators: make(map[string]*ThreatIndicator),
			patterns:   make(map[ThreatType]*ThreatPattern),
		},
		alertManager: &SecurityAlertManager{
			alerts:    make([]SecurityAlert, 0),
			alertChan: make(chan SecurityAlert, 100),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize threat patterns
	sm.initializeThreatPatterns()

	return sm
}

// Start starts the security monitoring
func (sm *SecurityMonitor) Start() error {
	if !sm.config.Enabled {
		log.Info().Msg("Security monitoring disabled")
		return nil
	}

	// Start event collection
	go sm.eventCollectionLoop()

	// Start threat detection
	go sm.threatDetectionLoop()

	// Start alert processing
	go sm.alertProcessingLoop()

	log.Info().Msg("Security monitoring started")
	return nil
}

// RecordEvent records a security event
func (sm *SecurityMonitor) RecordEvent(event SecurityEvent) {
	if !sm.config.Enabled {
		return
	}

	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}

	select {
	case sm.eventCollector.eventChan <- event:
		// Event queued successfully
	default:
		log.Warn().Msg("Security event channel full, dropping event")
	}
}

// eventCollectionLoop processes incoming security events
func (sm *SecurityMonitor) eventCollectionLoop() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case event := <-sm.eventCollector.eventChan:
			sm.processEvent(event)
		}
	}
}

// processEvent processes a single security event
func (sm *SecurityMonitor) processEvent(event SecurityEvent) {
	sm.eventCollector.mu.Lock()
	sm.eventCollector.events = append(sm.eventCollector.events, event)

	// Trim old events based on retention period
	cutoff := time.Now().Add(-sm.config.RetentionPeriod)
	var filteredEvents []SecurityEvent
	for _, e := range sm.eventCollector.events {
		if e.Timestamp.After(cutoff) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	sm.eventCollector.events = filteredEvents
	sm.eventCollector.mu.Unlock()

	// Trigger threat detection
	sm.analyzeEventForThreats(event)

	log.Debug().
		Str("event_id", event.ID).
		Str("event_type", string(event.Type)).
		Str("severity", event.Severity.String()).
		Msg("Security event processed")
}

// threatDetectionLoop runs periodic threat detection
func (sm *SecurityMonitor) threatDetectionLoop() {
	ticker := time.NewTicker(sm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.runThreatDetection()
		}
	}
}

// analyzeEventForThreats analyzes an event for immediate threats
func (sm *SecurityMonitor) analyzeEventForThreats(event SecurityEvent) {
	// Check for immediate threat patterns
	for threatType, pattern := range sm.threatDetector.patterns {
		if sm.matchesPattern(event, pattern) {
			sm.createThreatIndicator(threatType, []SecurityEvent{event})
		}
	}
}

// runThreatDetection runs comprehensive threat detection
func (sm *SecurityMonitor) runThreatDetection() {
	sm.eventCollector.mu.RLock()
	events := make([]SecurityEvent, len(sm.eventCollector.events))
	copy(events, sm.eventCollector.events)
	sm.eventCollector.mu.RUnlock()

	// Analyze events for threat patterns
	for threatType, pattern := range sm.threatDetector.patterns {
		matchingEvents := sm.findMatchingEvents(events, pattern)
		if len(matchingEvents) >= pattern.Threshold {
			sm.createThreatIndicator(threatType, matchingEvents)
		}
	}
}

// matchesPattern checks if an event matches a threat pattern
func (sm *SecurityMonitor) matchesPattern(event SecurityEvent, pattern *ThreatPattern) bool {
	// Check if event type matches
	for _, eventType := range pattern.EventTypes {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

// findMatchingEvents finds events matching a threat pattern within time window
func (sm *SecurityMonitor) findMatchingEvents(events []SecurityEvent, pattern *ThreatPattern) []SecurityEvent {
	var matchingEvents []SecurityEvent
	cutoff := time.Now().Add(-pattern.TimeWindow)

	for _, event := range events {
		if event.Timestamp.After(cutoff) && sm.matchesPattern(event, pattern) {
			matchingEvents = append(matchingEvents, event)
		}
	}

	return matchingEvents
}

// createThreatIndicator creates a new threat indicator
func (sm *SecurityMonitor) createThreatIndicator(threatType ThreatType, evidence []SecurityEvent) {
	sm.threatDetector.mu.Lock()
	defer sm.threatDetector.mu.Unlock()

	indicatorID := fmt.Sprintf("threat-%s-%d", threatType, time.Now().Unix())

	indicator := &ThreatIndicator{
		ID:          indicatorID,
		Type:        threatType,
		Severity:    sm.calculateThreatSeverity(threatType, evidence),
		Confidence:  sm.calculateConfidence(threatType, evidence),
		Description: sm.generateThreatDescription(threatType, evidence),
		Evidence:    evidence,
		DetectedAt:  time.Now(),
		Status:      ThreatStatusActive,
		Metadata:    make(map[string]interface{}),
	}

	sm.threatDetector.indicators[indicatorID] = indicator

	// Create security alert
	alert := SecurityAlert{
		ID:          fmt.Sprintf("alert-%s", indicatorID),
		Timestamp:   time.Now(),
		Type:        "threat_detected",
		Severity:    indicator.Severity,
		Title:       fmt.Sprintf("%s Detected", threatType),
		Description: indicator.Description,
		Source:      "security_monitor",
		Indicators:  []ThreatIndicator{*indicator},
		Status:      "active",
		Metadata:    make(map[string]interface{}),
	}

	sm.alertManager.alertChan <- alert

	log.Warn().
		Str("threat_type", string(threatType)).
		Str("indicator_id", indicatorID).
		Float64("confidence", indicator.Confidence).
		Msg("Security threat detected")
}

// calculateThreatSeverity calculates the severity of a threat
func (sm *SecurityMonitor) calculateThreatSeverity(threatType ThreatType, evidence []SecurityEvent) SeverityLevel {
	switch threatType {
	case ThreatTypeBruteForce:
		if len(evidence) > 10 {
			return SeverityCritical
		}
		return SeverityHigh
	case ThreatTypeDataExfiltration:
		return SeverityCritical
	case ThreatTypePrivilegeEscalation:
		return SeverityCritical
	case ThreatTypeDDoS:
		return SeverityHigh
	default:
		return SeverityMedium
	}
}

// calculateConfidence calculates confidence level for threat detection
func (sm *SecurityMonitor) calculateConfidence(threatType ThreatType, evidence []SecurityEvent) float64 {
	baseConfidence := 0.5
	evidenceWeight := float64(len(evidence)) * 0.1

	confidence := baseConfidence + evidenceWeight
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateThreatDescription generates a description for the threat
func (sm *SecurityMonitor) generateThreatDescription(threatType ThreatType, evidence []SecurityEvent) string {
	switch threatType {
	case ThreatTypeBruteForce:
		return fmt.Sprintf("Potential brute force attack detected with %d failed authentication attempts", len(evidence))
	case ThreatTypeAnomalousLogin:
		return "Anomalous login pattern detected"
	case ThreatTypeDataExfiltration:
		return "Potential data exfiltration detected"
	default:
		return fmt.Sprintf("Security threat of type %s detected", threatType)
	}
}

// alertProcessingLoop processes security alerts
func (sm *SecurityMonitor) alertProcessingLoop() {
	for {
		select {
		case <-sm.ctx.Done():
			return
		case alert := <-sm.alertManager.alertChan:
			sm.processAlert(alert)
		}
	}
}

// processAlert processes a security alert
func (sm *SecurityMonitor) processAlert(alert SecurityAlert) {
	sm.alertManager.mu.Lock()
	sm.alertManager.alerts = append(sm.alertManager.alerts, alert)
	sm.alertManager.mu.Unlock()

	// Send real-time alert if enabled
	if sm.config.EnableRealTimeAlerts {
		sm.sendRealTimeAlert(alert)
	}

	log.Info().
		Str("alert_id", alert.ID).
		Str("alert_type", alert.Type).
		Str("severity", alert.Severity.String()).
		Msg("Security alert processed")
}

// sendRealTimeAlert sends a real-time security alert
func (sm *SecurityMonitor) sendRealTimeAlert(alert SecurityAlert) {
	// This would integrate with notification systems
	log.Warn().
		Str("alert_id", alert.ID).
		Str("title", alert.Title).
		Str("description", alert.Description).
		Msg("SECURITY ALERT")
}

// initializeThreatPatterns initializes default threat detection patterns
func (sm *SecurityMonitor) initializeThreatPatterns() {
	sm.threatDetector.patterns[ThreatTypeBruteForce] = &ThreatPattern{
		Type:       ThreatTypeBruteForce,
		EventTypes: []SecurityEventType{EventTypeAuthentication},
		TimeWindow: 5 * time.Minute,
		Threshold:  5,
		Conditions: []string{"failed_login"},
	}

	sm.threatDetector.patterns[ThreatTypeAnomalousLogin] = &ThreatPattern{
		Type:       ThreatTypeAnomalousLogin,
		EventTypes: []SecurityEventType{EventTypeAuthentication},
		TimeWindow: 1 * time.Hour,
		Threshold:  1,
		Conditions: []string{"unusual_location", "unusual_time"},
	}

	sm.threatDetector.patterns[ThreatTypeDataExfiltration] = &ThreatPattern{
		Type:       ThreatTypeDataExfiltration,
		EventTypes: []SecurityEventType{EventTypeDataAccess},
		TimeWindow: 10 * time.Minute,
		Threshold:  3,
		Conditions: []string{"large_data_transfer"},
	}
}

// GetThreatIndicators returns current threat indicators
func (sm *SecurityMonitor) GetThreatIndicators() map[string]*ThreatIndicator {
	sm.threatDetector.mu.RLock()
	defer sm.threatDetector.mu.RUnlock()

	indicators := make(map[string]*ThreatIndicator)
	for k, v := range sm.threatDetector.indicators {
		indicators[k] = v
	}

	return indicators
}

// GetSecurityAlerts returns recent security alerts
func (sm *SecurityMonitor) GetSecurityAlerts() []SecurityAlert {
	sm.alertManager.mu.RLock()
	defer sm.alertManager.mu.RUnlock()

	alerts := make([]SecurityAlert, len(sm.alertManager.alerts))
	copy(alerts, sm.alertManager.alerts)

	return alerts
}

// Shutdown gracefully shuts down the security monitor
func (sm *SecurityMonitor) Shutdown() error {
	sm.cancel()
	log.Info().Msg("Security monitor stopped")
	return nil
}

// DefaultMonitoringConfig returns default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:                true,
		CollectionInterval:     30 * time.Second,
		ThreatDetectionLevel:   "medium",
		AlertThreshold:         5,
		RetentionPeriod:        24 * time.Hour,
		EnableRealTimeAlerts:   true,
		EnableAnomalyDetection: true,
		EnableBehaviorAnalysis: false,
	}
}
