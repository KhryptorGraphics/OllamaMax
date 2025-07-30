package security

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// SecurityMonitor monitors overall security status and events
type SecurityMonitor struct {
	mu        sync.RWMutex
	events    []*SecurityEvent
	maxEvents int

	// Components
	ids         *IntrusionDetectionSystem
	auditLogger *AuditLogger

	// Metrics
	metrics *MonitorMetrics

	// Configuration
	config *MonitorConfig

	// Event handlers
	eventHandlers []SecurityEventHandler

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      SecurityEventType      `json:"type"`
	Severity  ThreatSeverity         `json:"severity"`
	Source    string                 `json:"source"`
	PeerID    peer.ID                `json:"peer_id,omitempty"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Action    string                 `json:"action,omitempty"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	EventTypeAuthentication     SecurityEventType = "authentication"
	EventTypeAuthorization      SecurityEventType = "authorization"
	EventTypeIntrusion          SecurityEventType = "intrusion"
	EventTypeAnomalyDetection   SecurityEventType = "anomaly_detection"
	EventTypeKeyRotation        SecurityEventType = "key_rotation"
	EventTypeConnectionSecurity SecurityEventType = "connection_security"
	EventTypeDataIntegrity      SecurityEventType = "data_integrity"
	EventTypeSystemSecurity     SecurityEventType = "system_security"
)

// SecurityEventHandler handles security events
type SecurityEventHandler func(*SecurityEvent)

// MonitorMetrics tracks security monitoring metrics
type MonitorMetrics struct {
	TotalEvents           int64                       `json:"total_events"`
	EventsByType          map[SecurityEventType]int64 `json:"events_by_type"`
	EventsBySeverity      map[ThreatSeverity]int64    `json:"events_by_severity"`
	ActiveThreats         int64                       `json:"active_threats"`
	BlockedPeers          int64                       `json:"blocked_peers"`
	FailedAuthentications int64                       `json:"failed_authentications"`
	KeyRotations          int64                       `json:"key_rotations"`
	LastUpdate            time.Time                   `json:"last_update"`
}

// MonitorConfig configures the security monitor
type MonitorConfig struct {
	MaxEvents          int
	EventRetention     time.Duration
	MetricsInterval    time.Duration
	AlertThreshold     int64
	EnableAuditLogging bool
	AuditLogPath       string
}

// AuditLogger logs security events for compliance and forensics
type AuditLogger struct {
	mu      sync.Mutex
	logPath string
	enabled bool
}

// NewSecurityMonitor creates a new security monitor
func NewSecurityMonitor(config *MonitorConfig) *SecurityMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &MonitorConfig{
			MaxEvents:          10000,
			EventRetention:     24 * time.Hour,
			MetricsInterval:    time.Minute,
			AlertThreshold:     100,
			EnableAuditLogging: true,
			AuditLogPath:       "/var/log/ollama/security.log",
		}
	}

	monitor := &SecurityMonitor{
		events:        make([]*SecurityEvent, 0),
		maxEvents:     config.MaxEvents,
		config:        config,
		eventHandlers: make([]SecurityEventHandler, 0),
		metrics: &MonitorMetrics{
			EventsByType:     make(map[SecurityEventType]int64),
			EventsBySeverity: make(map[ThreatSeverity]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize audit logger
	if config.EnableAuditLogging {
		monitor.auditLogger = &AuditLogger{
			logPath: config.AuditLogPath,
			enabled: true,
		}
	}

	// Initialize IDS
	monitor.ids = NewIntrusionDetectionSystem(nil)

	// Add IDS alert handler
	monitor.ids.AddAlertHandler(func(alert *SecurityAlert) {
		monitor.RecordEvent(&SecurityEvent{
			ID:        alert.ID,
			Timestamp: alert.Timestamp,
			Type:      EventTypeIntrusion,
			Severity:  alert.Severity,
			Source:    "IDS",
			PeerID:    alert.PeerID,
			Message:   alert.Message,
			Details:   alert.Details,
			Action:    string(alert.Action),
		})
	})

	// Start background tasks
	monitor.wg.Add(2)
	go monitor.metricsLoop()
	go monitor.cleanupLoop()

	return monitor
}

// RecordEvent records a security event
func (sm *SecurityMonitor) RecordEvent(event *SecurityEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Add timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Generate ID if not set
	if event.ID == "" {
		event.ID = fmt.Sprintf("event-%d", time.Now().UnixNano())
	}

	// Add to events list
	sm.events = append(sm.events, event)

	// Limit events list size
	if len(sm.events) > sm.maxEvents {
		sm.events = sm.events[1:]
	}

	// Update metrics
	sm.metrics.TotalEvents++
	sm.metrics.EventsByType[event.Type]++
	sm.metrics.EventsBySeverity[event.Severity]++

	// Log to audit log
	if sm.auditLogger != nil && sm.auditLogger.enabled {
		sm.auditLogger.LogEvent(event)
	}

	// Call event handlers
	for _, handler := range sm.eventHandlers {
		go handler(event)
	}
}

// RecordAuthenticationEvent records an authentication event
func (sm *SecurityMonitor) RecordAuthenticationEvent(peerID peer.ID, success bool, details map[string]interface{}) {
	severity := SeverityInfo
	message := "Authentication successful"

	if !success {
		severity = SeverityWarning
		message = "Authentication failed"
		sm.mu.Lock()
		sm.metrics.FailedAuthentications++
		sm.mu.Unlock()
	}

	sm.RecordEvent(&SecurityEvent{
		Type:     EventTypeAuthentication,
		Severity: severity,
		Source:   "AuthManager",
		PeerID:   peerID,
		Message:  message,
		Details:  details,
	})
}

// RecordKeyRotationEvent records a key rotation event
func (sm *SecurityMonitor) RecordKeyRotationEvent(keyType string, success bool) {
	severity := SeverityInfo
	message := fmt.Sprintf("Key rotation successful for %s", keyType)

	if !success {
		severity = SeverityError
		message = fmt.Sprintf("Key rotation failed for %s", keyType)
	} else {
		sm.mu.Lock()
		sm.metrics.KeyRotations++
		sm.mu.Unlock()
	}

	sm.RecordEvent(&SecurityEvent{
		Type:     EventTypeKeyRotation,
		Severity: severity,
		Source:   "KeyManager",
		Message:  message,
		Details: map[string]interface{}{
			"key_type": keyType,
			"success":  success,
		},
	})
}

// RecordConnectionSecurityEvent records a connection security event
func (sm *SecurityMonitor) RecordConnectionSecurityEvent(peerID peer.ID, eventType string, details map[string]interface{}) {
	severity := SeverityInfo
	message := fmt.Sprintf("Connection security event: %s", eventType)

	// Determine severity based on event type
	switch eventType {
	case "tls_handshake_failed", "certificate_invalid", "encryption_failed":
		severity = SeverityError
	case "weak_cipher", "certificate_expiring":
		severity = SeverityWarning
	}

	sm.RecordEvent(&SecurityEvent{
		Type:     EventTypeConnectionSecurity,
		Severity: severity,
		Source:   "ConnectionManager",
		PeerID:   peerID,
		Message:  message,
		Details:  details,
	})
}

// GetEvents returns recent security events
func (sm *SecurityMonitor) GetEvents(limit int, eventType SecurityEventType) []*SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var filteredEvents []*SecurityEvent

	// Filter by type if specified
	for i := len(sm.events) - 1; i >= 0 && len(filteredEvents) < limit; i-- {
		event := sm.events[i]
		if eventType == "" || event.Type == eventType {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents
}

// GetMetrics returns security metrics
func (sm *SecurityMonitor) GetMetrics() *MonitorMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := *sm.metrics
	metrics.LastUpdate = time.Now()

	// Get IDS metrics
	if sm.ids != nil {
		idsMetrics := sm.ids.GetMetrics()
		metrics.ActiveThreats = idsMetrics.TotalAlerts
	}

	return &metrics
}

// GetSecurityStatus returns overall security status
func (sm *SecurityMonitor) GetSecurityStatus() map[string]interface{} {
	metrics := sm.GetMetrics()

	// Calculate security score (simplified)
	securityScore := 100.0

	// Deduct points for security issues
	if metrics.FailedAuthentications > 10 {
		securityScore -= 10.0
	}
	if metrics.ActiveThreats > 5 {
		securityScore -= 20.0
	}
	if metrics.BlockedPeers > 0 {
		securityScore -= 5.0
	}

	// Determine status
	status := "healthy"
	if securityScore < 70 {
		status = "degraded"
	}
	if securityScore < 50 {
		status = "critical"
	}

	return map[string]interface{}{
		"status":         status,
		"security_score": securityScore,
		"metrics":        metrics,
		"last_updated":   time.Now(),
	}
}

// AddEventHandler adds a security event handler
func (sm *SecurityMonitor) AddEventHandler(handler SecurityEventHandler) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.eventHandlers = append(sm.eventHandlers, handler)
}

// GetIntrusionDetectionSystem returns the IDS
func (sm *SecurityMonitor) GetIntrusionDetectionSystem() *IntrusionDetectionSystem {
	return sm.ids
}

// metricsLoop periodically updates metrics
func (sm *SecurityMonitor) metricsLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

// updateMetrics updates security metrics
func (sm *SecurityMonitor) updateMetrics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Update metrics based on current state
	sm.metrics.LastUpdate = time.Now()

	// Count blocked peers (simplified - would need actual blocked peer tracking)
	sm.metrics.BlockedPeers = 0

	// Check for alert threshold
	if sm.metrics.TotalEvents > sm.config.AlertThreshold {
		sm.RecordEvent(&SecurityEvent{
			Type:     EventTypeSystemSecurity,
			Severity: SeverityWarning,
			Source:   "SecurityMonitor",
			Message:  "High number of security events detected",
			Details: map[string]interface{}{
				"total_events": sm.metrics.TotalEvents,
				"threshold":    sm.config.AlertThreshold,
			},
		})
	}
}

// cleanupLoop periodically cleans up old events
func (sm *SecurityMonitor) cleanupLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.cleanupOldEvents()
		}
	}
}

// cleanupOldEvents removes old events based on retention policy
func (sm *SecurityMonitor) cleanupOldEvents() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-sm.config.EventRetention)

	// Find first event to keep
	keepIndex := 0
	for i, event := range sm.events {
		if event.Timestamp.After(cutoff) {
			keepIndex = i
			break
		}
	}

	// Remove old events
	if keepIndex > 0 {
		sm.events = sm.events[keepIndex:]
	}
}

// LogEvent logs an event to the audit log
func (al *AuditLogger) LogEvent(event *SecurityEvent) {
	if !al.enabled {
		return
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return
	}

	// In a real implementation, you would write to a file or external logging system
	// For now, we'll just log to stdout
	fmt.Printf("AUDIT: %s\n", string(eventJSON))
}

// Close closes the security monitor
func (sm *SecurityMonitor) Close() error {
	sm.cancel()
	sm.wg.Wait()

	if sm.ids != nil {
		sm.ids.Close()
	}

	return nil
}
