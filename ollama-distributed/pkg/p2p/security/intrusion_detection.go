package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// IntrusionDetectionSystem monitors network activity for suspicious behavior
type IntrusionDetectionSystem struct {
	mu                sync.RWMutex
	peerBehavior      map[peer.ID]*PeerBehavior
	threatRules       []ThreatRule
	alertHandlers     []AlertHandler
	config            *IDSConfig
	
	// Metrics
	metrics           *IDSMetrics
	
	// Lifecycle
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// PeerBehavior tracks behavior patterns for a peer
type PeerBehavior struct {
	PeerID            peer.ID
	ConnectionCount   int64
	MessageCount      int64
	BytesTransferred  int64
	FailedAttempts    int64
	LastActivity      time.Time
	FirstSeen         time.Time
	
	// Behavioral patterns
	ConnectionRate    float64  // connections per minute
	MessageRate       float64  // messages per minute
	DataRate          float64  // bytes per minute
	FailureRate       float64  // failures per minute
	
	// Anomaly scores
	AnomalyScore      float64
	ThreatLevel       ThreatLevel
	
	// Historical data
	ActivityWindows   []*ActivityWindow
}

// ActivityWindow represents activity in a time window
type ActivityWindow struct {
	StartTime         time.Time
	EndTime           time.Time
	Connections       int64
	Messages          int64
	Bytes             int64
	Failures          int64
}

// ThreatRule defines a rule for detecting threats
type ThreatRule struct {
	ID                string
	Name              string
	Description       string
	Severity          ThreatSeverity
	Condition         func(*PeerBehavior) bool
	Action            ThreatAction
	Enabled           bool
}

// ThreatLevel represents the threat level of a peer
type ThreatLevel string

const (
	ThreatLevelNone     ThreatLevel = "none"
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// ThreatSeverity represents the severity of a threat
type ThreatSeverity string

const (
	SeverityInfo     ThreatSeverity = "info"
	SeverityWarning  ThreatSeverity = "warning"
	SeverityError    ThreatSeverity = "error"
	SeverityCritical ThreatSeverity = "critical"
)

// ThreatAction represents actions to take when a threat is detected
type ThreatAction string

const (
	ActionLog        ThreatAction = "log"
	ActionAlert      ThreatAction = "alert"
	ActionThrottle   ThreatAction = "throttle"
	ActionBlock      ThreatAction = "block"
	ActionDisconnect ThreatAction = "disconnect"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string
	Timestamp   time.Time
	PeerID      peer.ID
	RuleID      string
	Severity    ThreatSeverity
	Message     string
	Details     map[string]interface{}
	Action      ThreatAction
}

// AlertHandler handles security alerts
type AlertHandler func(*SecurityAlert)

// IDSConfig configures the intrusion detection system
type IDSConfig struct {
	WindowSize        time.Duration
	MaxWindows        int
	UpdateInterval    time.Duration
	AnomalyThreshold  float64
	
	// Rate limits for anomaly detection
	MaxConnectionRate float64  // connections per minute
	MaxMessageRate    float64  // messages per minute
	MaxDataRate       float64  // bytes per minute
	MaxFailureRate    float64  // failures per minute
	
	// Behavioral analysis
	LearningPeriod    time.Duration
	BaselineWindow    time.Duration
}

// IDSMetrics tracks IDS performance
type IDSMetrics struct {
	TotalAlerts       int64
	AlertsBySeverity  map[ThreatSeverity]int64
	PeersMonitored    int64
	ThreatsBlocked    int64
	LastUpdate        time.Time
}

// NewIntrusionDetectionSystem creates a new IDS
func NewIntrusionDetectionSystem(config *IDSConfig) *IntrusionDetectionSystem {
	ctx, cancel := context.WithCancel(context.Background())
	
	if config == nil {
		config = &IDSConfig{
			WindowSize:        time.Minute,
			MaxWindows:        60, // 1 hour of history
			UpdateInterval:    10 * time.Second,
			AnomalyThreshold:  0.8,
			MaxConnectionRate: 10.0,  // 10 connections per minute
			MaxMessageRate:    100.0, // 100 messages per minute
			MaxDataRate:       1024 * 1024, // 1MB per minute
			MaxFailureRate:    5.0,   // 5 failures per minute
			LearningPeriod:    24 * time.Hour,
			BaselineWindow:    time.Hour,
		}
	}
	
	ids := &IntrusionDetectionSystem{
		peerBehavior:  make(map[peer.ID]*PeerBehavior),
		threatRules:   getDefaultThreatRules(),
		alertHandlers: make([]AlertHandler, 0),
		config:        config,
		metrics: &IDSMetrics{
			AlertsBySeverity: make(map[ThreatSeverity]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Start monitoring
	ids.wg.Add(1)
	go ids.monitoringLoop()
	
	return ids
}

// RecordActivity records activity for a peer
func (ids *IntrusionDetectionSystem) RecordActivity(peerID peer.ID, activityType string, data map[string]interface{}) {
	ids.mu.Lock()
	defer ids.mu.Unlock()
	
	behavior, exists := ids.peerBehavior[peerID]
	if !exists {
		behavior = &PeerBehavior{
			PeerID:          peerID,
			FirstSeen:       time.Now(),
			ActivityWindows: make([]*ActivityWindow, 0),
			ThreatLevel:     ThreatLevelNone,
		}
		ids.peerBehavior[peerID] = behavior
	}
	
	behavior.LastActivity = time.Now()
	
	// Update counters based on activity type
	switch activityType {
	case "connection":
		behavior.ConnectionCount++
	case "message":
		behavior.MessageCount++
		if bytes, ok := data["bytes"].(int64); ok {
			behavior.BytesTransferred += bytes
		}
	case "failure":
		behavior.FailedAttempts++
	}
	
	// Update current window
	ids.updateActivityWindow(behavior)
}

// updateActivityWindow updates the current activity window
func (ids *IntrusionDetectionSystem) updateActivityWindow(behavior *PeerBehavior) {
	now := time.Now()
	windowStart := now.Truncate(ids.config.WindowSize)
	
	// Find or create current window
	var currentWindow *ActivityWindow
	if len(behavior.ActivityWindows) > 0 {
		lastWindow := behavior.ActivityWindows[len(behavior.ActivityWindows)-1]
		if lastWindow.StartTime.Equal(windowStart) {
			currentWindow = lastWindow
		}
	}
	
	if currentWindow == nil {
		currentWindow = &ActivityWindow{
			StartTime: windowStart,
			EndTime:   windowStart.Add(ids.config.WindowSize),
		}
		behavior.ActivityWindows = append(behavior.ActivityWindows, currentWindow)
		
		// Limit window history
		if len(behavior.ActivityWindows) > ids.config.MaxWindows {
			behavior.ActivityWindows = behavior.ActivityWindows[1:]
		}
	}
	
	// Update window counters (simplified - in reality you'd track increments)
	currentWindow.Connections = behavior.ConnectionCount
	currentWindow.Messages = behavior.MessageCount
	currentWindow.Bytes = behavior.BytesTransferred
	currentWindow.Failures = behavior.FailedAttempts
}

// AnalyzeBehavior analyzes peer behavior for anomalies
func (ids *IntrusionDetectionSystem) AnalyzeBehavior(peerID peer.ID) {
	ids.mu.RLock()
	behavior, exists := ids.peerBehavior[peerID]
	ids.mu.RUnlock()
	
	if !exists {
		return
	}
	
	// Calculate rates
	ids.calculateRates(behavior)
	
	// Calculate anomaly score
	anomalyScore := ids.calculateAnomalyScore(behavior)
	
	ids.mu.Lock()
	behavior.AnomalyScore = anomalyScore
	behavior.ThreatLevel = ids.calculateThreatLevel(anomalyScore)
	ids.mu.Unlock()
	
	// Check threat rules
	ids.checkThreatRules(behavior)
}

// calculateRates calculates activity rates for a peer
func (ids *IntrusionDetectionSystem) calculateRates(behavior *PeerBehavior) {
	if len(behavior.ActivityWindows) < 2 {
		return
	}
	
	// Calculate rates based on recent windows
	recentWindows := behavior.ActivityWindows
	if len(recentWindows) > 10 {
		recentWindows = recentWindows[len(recentWindows)-10:] // Last 10 windows
	}
	
	totalTime := float64(len(recentWindows)) * ids.config.WindowSize.Minutes()
	if totalTime == 0 {
		return
	}
	
	// Calculate average rates
	var totalConnections, totalMessages, totalBytes, totalFailures int64
	for _, window := range recentWindows {
		totalConnections += window.Connections
		totalMessages += window.Messages
		totalBytes += window.Bytes
		totalFailures += window.Failures
	}
	
	behavior.ConnectionRate = float64(totalConnections) / totalTime
	behavior.MessageRate = float64(totalMessages) / totalTime
	behavior.DataRate = float64(totalBytes) / totalTime
	behavior.FailureRate = float64(totalFailures) / totalTime
}

// calculateAnomalyScore calculates an anomaly score for a peer
func (ids *IntrusionDetectionSystem) calculateAnomalyScore(behavior *PeerBehavior) float64 {
	score := 0.0
	
	// Check rate anomalies
	if behavior.ConnectionRate > ids.config.MaxConnectionRate {
		score += 0.3
	}
	if behavior.MessageRate > ids.config.MaxMessageRate {
		score += 0.2
	}
	if behavior.DataRate > ids.config.MaxDataRate {
		score += 0.2
	}
	if behavior.FailureRate > ids.config.MaxFailureRate {
		score += 0.3
	}
	
	// Normalize score
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// calculateThreatLevel calculates threat level based on anomaly score
func (ids *IntrusionDetectionSystem) calculateThreatLevel(anomalyScore float64) ThreatLevel {
	switch {
	case anomalyScore >= 0.9:
		return ThreatLevelCritical
	case anomalyScore >= 0.7:
		return ThreatLevelHigh
	case anomalyScore >= 0.5:
		return ThreatLevelMedium
	case anomalyScore >= 0.3:
		return ThreatLevelLow
	default:
		return ThreatLevelNone
	}
}

// checkThreatRules checks all threat rules against a peer's behavior
func (ids *IntrusionDetectionSystem) checkThreatRules(behavior *PeerBehavior) {
	for _, rule := range ids.threatRules {
		if !rule.Enabled {
			continue
		}
		
		if rule.Condition(behavior) {
			alert := &SecurityAlert{
				ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
				Timestamp: time.Now(),
				PeerID:    behavior.PeerID,
				RuleID:    rule.ID,
				Severity:  rule.Severity,
				Message:   fmt.Sprintf("Threat detected: %s", rule.Name),
				Details: map[string]interface{}{
					"anomaly_score": behavior.AnomalyScore,
					"threat_level":  behavior.ThreatLevel,
					"description":   rule.Description,
				},
				Action: rule.Action,
			}
			
			ids.handleAlert(alert)
		}
	}
}

// handleAlert handles a security alert
func (ids *IntrusionDetectionSystem) handleAlert(alert *SecurityAlert) {
	ids.metrics.TotalAlerts++
	ids.metrics.AlertsBySeverity[alert.Severity]++
	
	// Call alert handlers
	for _, handler := range ids.alertHandlers {
		go handler(alert)
	}
}

// AddAlertHandler adds an alert handler
func (ids *IntrusionDetectionSystem) AddAlertHandler(handler AlertHandler) {
	ids.mu.Lock()
	defer ids.mu.Unlock()
	ids.alertHandlers = append(ids.alertHandlers, handler)
}

// GetPeerThreatLevel returns the threat level for a peer
func (ids *IntrusionDetectionSystem) GetPeerThreatLevel(peerID peer.ID) ThreatLevel {
	ids.mu.RLock()
	defer ids.mu.RUnlock()
	
	if behavior, exists := ids.peerBehavior[peerID]; exists {
		return behavior.ThreatLevel
	}
	return ThreatLevelNone
}

// GetMetrics returns IDS metrics
func (ids *IntrusionDetectionSystem) GetMetrics() *IDSMetrics {
	ids.mu.RLock()
	defer ids.mu.RUnlock()
	
	metrics := *ids.metrics
	metrics.PeersMonitored = int64(len(ids.peerBehavior))
	metrics.LastUpdate = time.Now()
	
	return &metrics
}

// monitoringLoop runs the main monitoring loop
func (ids *IntrusionDetectionSystem) monitoringLoop() {
	defer ids.wg.Done()
	
	ticker := time.NewTicker(ids.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ids.ctx.Done():
			return
		case <-ticker.C:
			ids.performAnalysis()
		}
	}
}

// performAnalysis performs behavioral analysis on all peers
func (ids *IntrusionDetectionSystem) performAnalysis() {
	ids.mu.RLock()
	peers := make([]peer.ID, 0, len(ids.peerBehavior))
	for peerID := range ids.peerBehavior {
		peers = append(peers, peerID)
	}
	ids.mu.RUnlock()
	
	for _, peerID := range peers {
		ids.AnalyzeBehavior(peerID)
	}
}

// Close closes the IDS
func (ids *IntrusionDetectionSystem) Close() error {
	ids.cancel()
	ids.wg.Wait()
	return nil
}

// getDefaultThreatRules returns default threat detection rules
func getDefaultThreatRules() []ThreatRule {
	return []ThreatRule{
		{
			ID:          "high-connection-rate",
			Name:        "High Connection Rate",
			Description: "Peer is making too many connections",
			Severity:    SeverityWarning,
			Condition: func(behavior *PeerBehavior) bool {
				return behavior.ConnectionRate > 20.0 // 20 connections per minute
			},
			Action:  ActionThrottle,
			Enabled: true,
		},
		{
			ID:          "high-failure-rate",
			Name:        "High Failure Rate",
			Description: "Peer has too many failed attempts",
			Severity:    SeverityError,
			Condition: func(behavior *PeerBehavior) bool {
				return behavior.FailureRate > 10.0 // 10 failures per minute
			},
			Action:  ActionBlock,
			Enabled: true,
		},
		{
			ID:          "critical-anomaly",
			Name:        "Critical Anomaly Score",
			Description: "Peer behavior is highly anomalous",
			Severity:    SeverityCritical,
			Condition: func(behavior *PeerBehavior) bool {
				return behavior.AnomalyScore >= 0.9
			},
			Action:  ActionDisconnect,
			Enabled: true,
		},
	}
}
