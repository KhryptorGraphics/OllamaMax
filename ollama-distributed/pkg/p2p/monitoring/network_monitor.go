package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// NetworkMonitor provides comprehensive monitoring of P2P network performance
type NetworkMonitor struct {
	config *MonitorConfig

	// Metrics collectors
	connectionMetrics  *ConnectionMetrics
	messageMetrics     *MessageMetrics
	protocolMetrics    *ProtocolMetrics
	performanceMetrics *PerformanceMetrics
	securityMetrics    *SecurityMetrics

	// Real-time tracking
	activeConnections map[peer.ID]*ConnectionTracker
	connectionsMu     sync.RWMutex

	// Protocol tracking
	protocolStats map[protocol.ID]*ProtocolStats
	protocolMu    sync.RWMutex

	// Performance tracking
	latencyTracker    *LatencyTracker
	throughputTracker *ThroughputTracker

	// Event tracking
	eventBuffer  *EventBuffer
	alertManager *AlertManager

	// Lifecycle
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	runningMu sync.RWMutex
}

// MonitorConfig configures the network monitor
type MonitorConfig struct {
	// Collection intervals
	MetricsInterval     time.Duration
	HealthCheckInterval time.Duration
	CleanupInterval     time.Duration

	// Buffer sizes
	EventBufferSize   int
	MetricsBufferSize int

	// Retention periods
	MetricsRetention time.Duration
	EventRetention   time.Duration

	// Performance settings
	MaxTrackedConnections int
	MaxTrackedProtocols   int

	// Alert settings
	EnableAlerts    bool
	AlertThresholds *AlertThresholds

	// Export settings
	EnablePrometheus bool
	PrometheusPort   int
	EnableInfluxDB   bool
	InfluxDBConfig   *InfluxDBConfig
}

// ConnectionMetrics tracks connection-related metrics
type ConnectionMetrics struct {
	// Connection counts
	TotalConnections   int64
	ActiveConnections  int64
	FailedConnections  int64
	DroppedConnections int64

	// Connection lifecycle
	ConnectionsOpened  int64
	ConnectionsClosed  int64
	ConnectionTimeouts int64

	// Connection quality
	AverageLatency time.Duration
	MedianLatency  time.Duration
	P95Latency     time.Duration
	P99Latency     time.Duration

	// Bandwidth usage
	TotalBytesIn  int64
	TotalBytesOut int64
	BandwidthIn   float64 // bytes/sec
	BandwidthOut  float64 // bytes/sec

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// MessageMetrics tracks message-related metrics
type MessageMetrics struct {
	// Message counts
	TotalMessages    int64
	MessagesSent     int64
	MessagesReceived int64
	MessagesDropped  int64
	MessagesRetried  int64

	// Message types
	MessagesByType     map[string]int64
	MessagesByProtocol map[string]int64

	// Message performance
	AverageMessageSize float64
	MessageThroughput  float64 // messages/sec

	// Bandwidth usage
	TotalBytesIn  int64
	TotalBytesOut int64

	// Queue metrics
	QueueDepth     int64
	QueueOverflows int64

	// Reliability metrics
	DeliveryRate float64 // percentage
	AckRate      float64 // percentage

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// ProtocolMetrics tracks protocol-specific metrics
type ProtocolMetrics struct {
	// Protocol usage
	ProtocolCounts  map[protocol.ID]int64
	ProtocolErrors  map[protocol.ID]int64
	ProtocolLatency map[protocol.ID]time.Duration

	// Protocol performance
	ProtocolThroughput map[protocol.ID]float64
	ProtocolSuccess    map[protocol.ID]float64

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// PerformanceMetrics tracks overall network performance
type PerformanceMetrics struct {
	// Network performance
	NetworkLatency     time.Duration
	NetworkThroughput  float64
	NetworkUtilization float64

	// Resource usage
	CPUUsage       float64
	MemoryUsage    float64
	NetworkIOUsage float64

	// Error rates
	ErrorRate   float64
	TimeoutRate float64
	RetryRate   float64

	// Availability
	Uptime           time.Duration
	AvailabilityRate float64

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	// Authentication metrics
	AuthAttempts  int64
	AuthFailures  int64
	AuthSuccesses int64

	// Encryption metrics
	EncryptedMessages int64
	DecryptionErrors  int64

	// Security violations
	SecurityViolations int64
	BlockedConnections int64
	SuspiciousActivity int64

	// Certificate metrics
	CertificateErrors   int64
	CertificateExpiries int64

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// ConnectionTracker tracks individual connection metrics
type ConnectionTracker struct {
	PeerID       peer.ID
	ConnectedAt  time.Time
	LastActivity time.Time

	// Traffic metrics
	BytesSent        int64
	BytesReceived    int64
	MessagesSent     int64
	MessagesReceived int64

	// Performance metrics
	Latency    time.Duration
	RTT        time.Duration
	PacketLoss float64

	// Quality metrics
	ConnectionQuality float64
	Reliability       float64

	// Error tracking
	Errors   int64
	Timeouts int64
	Retries  int64

	mu sync.RWMutex
}

// ProtocolStats tracks protocol-specific statistics
type ProtocolStats struct {
	Protocol     protocol.ID
	MessageCount int64
	ErrorCount   int64
	TotalLatency time.Duration
	TotalBytes   int64
	LastUsed     time.Time

	mu sync.RWMutex
}

// LatencyTracker tracks network latency measurements
type LatencyTracker struct {
	measurements    []time.Duration
	maxMeasurements int
	currentIndex    int
	mu              sync.RWMutex
}

// ThroughputTracker tracks network throughput measurements
type ThroughputTracker struct {
	measurements    []float64
	timestamps      []time.Time
	maxMeasurements int
	currentIndex    int
	mu              sync.RWMutex
}

// EventBuffer stores network events for analysis
type EventBuffer struct {
	events       []*NetworkEvent
	maxEvents    int
	currentIndex int
	mu           sync.RWMutex
}

// NetworkEvent represents a network event
type NetworkEvent struct {
	Timestamp time.Time
	Type      EventType
	PeerID    peer.ID
	Protocol  protocol.ID
	Message   string
	Severity  EventSeverity
	Metadata  map[string]interface{}
}

// AlertManager manages network alerts
type AlertManager struct {
	config         *AlertThresholds
	activeAlerts   map[string]*Alert
	alertHistory   []*Alert
	alertCallbacks []AlertCallback
	mu             sync.RWMutex
}

// Alert represents a network alert
type Alert struct {
	ID         string
	Type       AlertType
	Severity   AlertSeverity
	Message    string
	Timestamp  time.Time
	Resolved   bool
	ResolvedAt time.Time
	Metadata   map[string]interface{}
}

// AlertThresholds defines alert thresholds
type AlertThresholds struct {
	MaxLatency            time.Duration
	MinThroughput         float64
	MaxErrorRate          float64
	MaxConnectionFailures int64
	MaxMemoryUsage        float64
	MaxCPUUsage           float64
}

// InfluxDBConfig configures InfluxDB export
type InfluxDBConfig struct {
	URL             string
	Database        string
	Username        string
	Password        string
	RetentionPolicy string
}

// Enums and constants
type EventType string

const (
	EventConnectionOpened  EventType = "connection_opened"
	EventConnectionClosed  EventType = "connection_closed"
	EventConnectionFailed  EventType = "connection_failed"
	EventMessageSent       EventType = "message_sent"
	EventMessageReceived   EventType = "message_received"
	EventMessageDropped    EventType = "message_dropped"
	EventProtocolError     EventType = "protocol_error"
	EventSecurityViolation EventType = "security_violation"
	EventPerformanceIssue  EventType = "performance_issue"
)

type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

type AlertType string

const (
	AlertHighLatency      AlertType = "high_latency"
	AlertLowThroughput    AlertType = "low_throughput"
	AlertHighErrorRate    AlertType = "high_error_rate"
	AlertConnectionIssues AlertType = "connection_issues"
	AlertResourceUsage    AlertType = "resource_usage"
	AlertSecurityIssue    AlertType = "security_issue"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// Callback types
type AlertCallback func(alert *Alert)

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(config *MonitorConfig) *NetworkMonitor {
	if config == nil {
		config = &MonitorConfig{
			MetricsInterval:       10 * time.Second,
			HealthCheckInterval:   30 * time.Second,
			CleanupInterval:       5 * time.Minute,
			EventBufferSize:       10000,
			MetricsBufferSize:     1000,
			MetricsRetention:      24 * time.Hour,
			EventRetention:        1 * time.Hour,
			MaxTrackedConnections: 10000,
			MaxTrackedProtocols:   100,
			EnableAlerts:          true,
			AlertThresholds: &AlertThresholds{
				MaxLatency:            1 * time.Second,
				MinThroughput:         1000, // bytes/sec
				MaxErrorRate:          0.05, // 5%
				MaxConnectionFailures: 100,
				MaxMemoryUsage:        0.8, // 80%
				MaxCPUUsage:           0.8, // 80%
			},
			EnablePrometheus: false,
			PrometheusPort:   9090,
			EnableInfluxDB:   false,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &NetworkMonitor{
		config:            config,
		connectionMetrics: &ConnectionMetrics{LastUpdated: time.Now()},
		messageMetrics: &MessageMetrics{
			MessagesByType:     make(map[string]int64),
			MessagesByProtocol: make(map[string]int64),
			LastUpdated:        time.Now(),
		},
		protocolMetrics: &ProtocolMetrics{
			ProtocolCounts:     make(map[protocol.ID]int64),
			ProtocolErrors:     make(map[protocol.ID]int64),
			ProtocolLatency:    make(map[protocol.ID]time.Duration),
			ProtocolThroughput: make(map[protocol.ID]float64),
			ProtocolSuccess:    make(map[protocol.ID]float64),
			LastUpdated:        time.Now(),
		},
		performanceMetrics: &PerformanceMetrics{LastUpdated: time.Now()},
		securityMetrics:    &SecurityMetrics{LastUpdated: time.Now()},
		activeConnections:  make(map[peer.ID]*ConnectionTracker),
		protocolStats:      make(map[protocol.ID]*ProtocolStats),
		latencyTracker:     NewLatencyTracker(1000),
		throughputTracker:  NewThroughputTracker(1000),
		eventBuffer:        NewEventBuffer(config.EventBufferSize),
		alertManager:       NewAlertManager(config.AlertThresholds),
		ctx:                ctx,
		cancel:             cancel,
	}

	return monitor
}

// Start starts the network monitor
func (nm *NetworkMonitor) Start() error {
	nm.runningMu.Lock()
	defer nm.runningMu.Unlock()

	if nm.running {
		return nil
	}

	nm.running = true

	// Start metrics collection
	nm.wg.Add(1)
	go nm.metricsCollector()

	// Start health checker
	nm.wg.Add(1)
	go nm.healthChecker()

	// Start cleanup routine
	nm.wg.Add(1)
	go nm.cleanupRoutine()

	// Start alert processor
	if nm.config.EnableAlerts {
		nm.wg.Add(1)
		go nm.alertProcessor()
	}

	// Start Prometheus exporter if enabled
	if nm.config.EnablePrometheus {
		nm.wg.Add(1)
		go nm.prometheusExporter()
	}

	// Start InfluxDB exporter if enabled
	if nm.config.EnableInfluxDB {
		nm.wg.Add(1)
		go nm.influxDBExporter()
	}

	return nil
}

// Stop stops the network monitor
func (nm *NetworkMonitor) Stop() error {
	nm.runningMu.Lock()
	defer nm.runningMu.Unlock()

	if !nm.running {
		return nil
	}

	nm.running = false
	nm.cancel()
	nm.wg.Wait()

	return nil
}

// Helper constructors
func NewLatencyTracker(maxMeasurements int) *LatencyTracker {
	return &LatencyTracker{
		measurements:    make([]time.Duration, maxMeasurements),
		maxMeasurements: maxMeasurements,
	}
}

func NewThroughputTracker(maxMeasurements int) *ThroughputTracker {
	return &ThroughputTracker{
		measurements:    make([]float64, maxMeasurements),
		timestamps:      make([]time.Time, maxMeasurements),
		maxMeasurements: maxMeasurements,
	}
}

func NewEventBuffer(maxEvents int) *EventBuffer {
	return &EventBuffer{
		events:    make([]*NetworkEvent, maxEvents),
		maxEvents: maxEvents,
	}
}

func NewAlertManager(thresholds *AlertThresholds) *AlertManager {
	return &AlertManager{
		config:         thresholds,
		activeAlerts:   make(map[string]*Alert),
		alertHistory:   make([]*Alert, 0),
		alertCallbacks: make([]AlertCallback, 0),
	}
}

// Monitoring methods

// RecordConnectionOpened records a new connection
func (nm *NetworkMonitor) RecordConnectionOpened(peerID peer.ID) {
	nm.connectionsMu.Lock()
	defer nm.connectionsMu.Unlock()

	tracker := &ConnectionTracker{
		PeerID:       peerID,
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
	}

	nm.activeConnections[peerID] = tracker

	nm.connectionMetrics.mu.Lock()
	nm.connectionMetrics.ConnectionsOpened++
	nm.connectionMetrics.ActiveConnections++
	nm.connectionMetrics.TotalConnections++
	nm.connectionMetrics.LastUpdated = time.Now()
	nm.connectionMetrics.mu.Unlock()

	// Record event
	nm.recordEvent(&NetworkEvent{
		Timestamp: time.Now(),
		Type:      EventConnectionOpened,
		PeerID:    peerID,
		Severity:  SeverityInfo,
		Message:   "Connection opened",
	})
}

// RecordConnectionClosed records a closed connection
func (nm *NetworkMonitor) RecordConnectionClosed(peerID peer.ID) {
	nm.connectionsMu.Lock()
	defer nm.connectionsMu.Unlock()

	delete(nm.activeConnections, peerID)

	nm.connectionMetrics.mu.Lock()
	nm.connectionMetrics.ConnectionsClosed++
	nm.connectionMetrics.ActiveConnections--
	nm.connectionMetrics.LastUpdated = time.Now()
	nm.connectionMetrics.mu.Unlock()

	// Record event
	nm.recordEvent(&NetworkEvent{
		Timestamp: time.Now(),
		Type:      EventConnectionClosed,
		PeerID:    peerID,
		Severity:  SeverityInfo,
		Message:   "Connection closed",
	})
}

// RecordConnectionFailed records a failed connection
func (nm *NetworkMonitor) RecordConnectionFailed(peerID peer.ID, reason string) {
	nm.connectionMetrics.mu.Lock()
	nm.connectionMetrics.FailedConnections++
	nm.connectionMetrics.LastUpdated = time.Now()
	nm.connectionMetrics.mu.Unlock()

	// Record event
	nm.recordEvent(&NetworkEvent{
		Timestamp: time.Now(),
		Type:      EventConnectionFailed,
		PeerID:    peerID,
		Severity:  SeverityWarning,
		Message:   "Connection failed: " + reason,
	})
}

// RecordMessageSent records a sent message
func (nm *NetworkMonitor) RecordMessageSent(peerID peer.ID, protocol protocol.ID, messageType string, size int) {
	// Update connection tracker
	nm.connectionsMu.RLock()
	if tracker, exists := nm.activeConnections[peerID]; exists {
		tracker.mu.Lock()
		tracker.MessagesSent++
		tracker.BytesSent += int64(size)
		tracker.LastActivity = time.Now()
		tracker.mu.Unlock()
	}
	nm.connectionsMu.RUnlock()

	// Update message metrics
	nm.messageMetrics.mu.Lock()
	nm.messageMetrics.MessagesSent++
	nm.messageMetrics.TotalMessages++
	nm.messageMetrics.MessagesByType[messageType]++
	nm.messageMetrics.MessagesByProtocol[string(protocol)]++
	nm.messageMetrics.TotalBytesOut += int64(size)
	nm.messageMetrics.LastUpdated = time.Now()
	nm.messageMetrics.mu.Unlock()

	// Update protocol stats
	nm.updateProtocolStats(protocol, size)
}

// RecordMessageReceived records a received message
func (nm *NetworkMonitor) RecordMessageReceived(peerID peer.ID, protocol protocol.ID, messageType string, size int) {
	// Update connection tracker
	nm.connectionsMu.RLock()
	if tracker, exists := nm.activeConnections[peerID]; exists {
		tracker.mu.Lock()
		tracker.MessagesReceived++
		tracker.BytesReceived += int64(size)
		tracker.LastActivity = time.Now()
		tracker.mu.Unlock()
	}
	nm.connectionsMu.RUnlock()

	// Update message metrics
	nm.messageMetrics.mu.Lock()
	nm.messageMetrics.MessagesReceived++
	nm.messageMetrics.TotalMessages++
	nm.messageMetrics.MessagesByType[messageType]++
	nm.messageMetrics.MessagesByProtocol[string(protocol)]++
	nm.messageMetrics.TotalBytesIn += int64(size)
	nm.messageMetrics.LastUpdated = time.Now()
	nm.messageMetrics.mu.Unlock()

	// Update protocol stats
	nm.updateProtocolStats(protocol, size)
}

// RecordMessageDropped records a dropped message
func (nm *NetworkMonitor) RecordMessageDropped(peerID peer.ID, protocol protocol.ID, reason string) {
	nm.messageMetrics.mu.Lock()
	nm.messageMetrics.MessagesDropped++
	nm.messageMetrics.LastUpdated = time.Now()
	nm.messageMetrics.mu.Unlock()

	// Record event
	nm.recordEvent(&NetworkEvent{
		Timestamp: time.Now(),
		Type:      EventMessageDropped,
		PeerID:    peerID,
		Protocol:  protocol,
		Severity:  SeverityWarning,
		Message:   "Message dropped: " + reason,
	})
}

// RecordLatency records a latency measurement
func (nm *NetworkMonitor) RecordLatency(peerID peer.ID, latency time.Duration) {
	// Update connection tracker
	nm.connectionsMu.RLock()
	if tracker, exists := nm.activeConnections[peerID]; exists {
		tracker.mu.Lock()
		tracker.Latency = latency
		tracker.mu.Unlock()
	}
	nm.connectionsMu.RUnlock()

	// Update latency tracker
	nm.latencyTracker.AddMeasurement(latency)

	// Update connection metrics
	nm.connectionMetrics.mu.Lock()
	nm.connectionMetrics.AverageLatency = nm.latencyTracker.GetAverage()
	nm.connectionMetrics.MedianLatency = nm.latencyTracker.GetMedian()
	nm.connectionMetrics.P95Latency = nm.latencyTracker.GetPercentile(95)
	nm.connectionMetrics.P99Latency = nm.latencyTracker.GetPercentile(99)
	nm.connectionMetrics.LastUpdated = time.Now()
	nm.connectionMetrics.mu.Unlock()
}

// RecordThroughput records a throughput measurement
func (nm *NetworkMonitor) RecordThroughput(throughput float64) {
	nm.throughputTracker.AddMeasurement(throughput)

	nm.performanceMetrics.mu.Lock()
	nm.performanceMetrics.NetworkThroughput = nm.throughputTracker.GetAverage()
	nm.performanceMetrics.LastUpdated = time.Now()
	nm.performanceMetrics.mu.Unlock()
}

// RecordSecurityEvent records a security-related event
func (nm *NetworkMonitor) RecordSecurityEvent(eventType string, peerID peer.ID, details string) {
	nm.securityMetrics.mu.Lock()
	switch eventType {
	case "auth_attempt":
		nm.securityMetrics.AuthAttempts++
	case "auth_failure":
		nm.securityMetrics.AuthFailures++
	case "auth_success":
		nm.securityMetrics.AuthSuccesses++
	case "security_violation":
		nm.securityMetrics.SecurityViolations++
	case "blocked_connection":
		nm.securityMetrics.BlockedConnections++
	}
	nm.securityMetrics.LastUpdated = time.Now()
	nm.securityMetrics.mu.Unlock()

	// Record event
	nm.recordEvent(&NetworkEvent{
		Timestamp: time.Now(),
		Type:      EventSecurityViolation,
		PeerID:    peerID,
		Severity:  SeverityError,
		Message:   eventType + ": " + details,
	})
}

// updateProtocolStats updates protocol-specific statistics
func (nm *NetworkMonitor) updateProtocolStats(protocol protocol.ID, size int) {
	nm.protocolMu.Lock()
	defer nm.protocolMu.Unlock()

	stats, exists := nm.protocolStats[protocol]
	if !exists {
		stats = &ProtocolStats{
			Protocol: protocol,
		}
		nm.protocolStats[protocol] = stats
	}

	stats.mu.Lock()
	stats.MessageCount++
	stats.TotalBytes += int64(size)
	stats.LastUsed = time.Now()
	stats.mu.Unlock()

	// Update protocol metrics
	nm.protocolMetrics.mu.Lock()
	nm.protocolMetrics.ProtocolCounts[protocol]++
	nm.protocolMetrics.LastUpdated = time.Now()
	nm.protocolMetrics.mu.Unlock()
}

// recordEvent records a network event
func (nm *NetworkMonitor) recordEvent(event *NetworkEvent) {
	nm.eventBuffer.AddEvent(event)
}

// Worker functions

// metricsCollector collects and updates metrics periodically
func (nm *NetworkMonitor) metricsCollector() {
	defer nm.wg.Done()

	ticker := time.NewTicker(nm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.collectMetrics()
		}
	}
}

// healthChecker performs periodic health checks
func (nm *NetworkMonitor) healthChecker() {
	defer nm.wg.Done()

	ticker := time.NewTicker(nm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.performHealthCheck()
		}
	}
}

// cleanupRoutine performs periodic cleanup
func (nm *NetworkMonitor) cleanupRoutine() {
	defer nm.wg.Done()

	ticker := time.NewTicker(nm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.performCleanup()
		}
	}
}

// alertProcessor processes alerts
func (nm *NetworkMonitor) alertProcessor() {
	defer nm.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.checkAlerts()
		}
	}
}

// prometheusExporter exports metrics to Prometheus
func (nm *NetworkMonitor) prometheusExporter() {
	defer nm.wg.Done()

	// Implementation would start Prometheus HTTP server
	// For now, this is a placeholder
	<-nm.ctx.Done()
}

// influxDBExporter exports metrics to InfluxDB
func (nm *NetworkMonitor) influxDBExporter() {
	defer nm.wg.Done()

	// Implementation would export metrics to InfluxDB
	// For now, this is a placeholder
	<-nm.ctx.Done()
}

// Tracker method implementations

// AddMeasurement adds a latency measurement
func (lt *LatencyTracker) AddMeasurement(latency time.Duration) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.measurements[lt.currentIndex] = latency
	lt.currentIndex = (lt.currentIndex + 1) % lt.maxMeasurements
}

// GetAverage returns the average latency
func (lt *LatencyTracker) GetAverage() time.Duration {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	var total time.Duration
	count := 0

	for _, measurement := range lt.measurements {
		if measurement > 0 {
			total += measurement
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / time.Duration(count)
}

// GetMedian returns the median latency
func (lt *LatencyTracker) GetMedian() time.Duration {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	var validMeasurements []time.Duration
	for _, measurement := range lt.measurements {
		if measurement > 0 {
			validMeasurements = append(validMeasurements, measurement)
		}
	}

	if len(validMeasurements) == 0 {
		return 0
	}

	// Simple median calculation (would use sort in production)
	if len(validMeasurements) == 1 {
		return validMeasurements[0]
	}

	return validMeasurements[len(validMeasurements)/2]
}

// GetPercentile returns the specified percentile
func (lt *LatencyTracker) GetPercentile(percentile int) time.Duration {
	lt.mu.RLock()
	defer lt.mu.RUnlock()

	var validMeasurements []time.Duration
	for _, measurement := range lt.measurements {
		if measurement > 0 {
			validMeasurements = append(validMeasurements, measurement)
		}
	}

	if len(validMeasurements) == 0 {
		return 0
	}

	// Simple percentile calculation (would use proper sorting in production)
	index := (percentile * len(validMeasurements)) / 100
	if index >= len(validMeasurements) {
		index = len(validMeasurements) - 1
	}

	return validMeasurements[index]
}

// AddMeasurement adds a throughput measurement
func (tt *ThroughputTracker) AddMeasurement(throughput float64) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	tt.measurements[tt.currentIndex] = throughput
	tt.timestamps[tt.currentIndex] = time.Now()
	tt.currentIndex = (tt.currentIndex + 1) % tt.maxMeasurements
}

// GetAverage returns the average throughput
func (tt *ThroughputTracker) GetAverage() float64 {
	tt.mu.RLock()
	defer tt.mu.RUnlock()

	var total float64
	count := 0

	for _, measurement := range tt.measurements {
		if measurement > 0 {
			total += measurement
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

// AddEvent adds an event to the buffer
func (eb *EventBuffer) AddEvent(event *NetworkEvent) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.events[eb.currentIndex] = event
	eb.currentIndex = (eb.currentIndex + 1) % eb.maxEvents
}

// GetEvents returns recent events
func (eb *EventBuffer) GetEvents(limit int) []*NetworkEvent {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var events []*NetworkEvent
	count := 0

	for i := 0; i < eb.maxEvents && count < limit; i++ {
		index := (eb.currentIndex - 1 - i + eb.maxEvents) % eb.maxEvents
		if eb.events[index] != nil {
			events = append(events, eb.events[index])
			count++
		}
	}

	return events
}

// Implementation placeholders for missing methods

// collectMetrics collects current metrics
func (nm *NetworkMonitor) collectMetrics() {
	// Update bandwidth calculations
	nm.connectionMetrics.mu.Lock()
	// Calculate bandwidth based on recent measurements
	nm.connectionMetrics.BandwidthIn = float64(nm.messageMetrics.TotalBytesIn) / time.Since(nm.connectionMetrics.LastUpdated).Seconds()
	nm.connectionMetrics.BandwidthOut = float64(nm.messageMetrics.TotalBytesOut) / time.Since(nm.connectionMetrics.LastUpdated).Seconds()
	nm.connectionMetrics.LastUpdated = time.Now()
	nm.connectionMetrics.mu.Unlock()

	// Update message throughput
	nm.messageMetrics.mu.Lock()
	nm.messageMetrics.MessageThroughput = float64(nm.messageMetrics.TotalMessages) / time.Since(nm.messageMetrics.LastUpdated).Seconds()
	nm.messageMetrics.LastUpdated = time.Now()
	nm.messageMetrics.mu.Unlock()
}

// performHealthCheck performs health checks
func (nm *NetworkMonitor) performHealthCheck() {
	// Check connection health
	nm.connectionsMu.RLock()
	staleConnections := 0
	for _, tracker := range nm.activeConnections {
		if time.Since(tracker.LastActivity) > 5*time.Minute {
			staleConnections++
		}
	}
	nm.connectionsMu.RUnlock()

	// Update performance metrics
	nm.performanceMetrics.mu.Lock()
	nm.performanceMetrics.NetworkUtilization = float64(staleConnections) / float64(len(nm.activeConnections))
	nm.performanceMetrics.LastUpdated = time.Now()
	nm.performanceMetrics.mu.Unlock()
}

// performCleanup performs periodic cleanup
func (nm *NetworkMonitor) performCleanup() {
	// Clean up old events
	// Clean up stale connections
	// Clean up old metrics
}

// checkAlerts checks for alert conditions
func (nm *NetworkMonitor) checkAlerts() {
	if !nm.config.EnableAlerts {
		return
	}

	// Check latency alerts
	if nm.connectionMetrics.AverageLatency > nm.config.AlertThresholds.MaxLatency {
		nm.alertManager.TriggerAlert("high_latency", string(AlertSeverityHigh), "Average latency exceeded threshold")
	}

	// Check error rate alerts
	errorRate := float64(nm.messageMetrics.MessagesDropped) / float64(nm.messageMetrics.TotalMessages)
	if errorRate > nm.config.AlertThresholds.MaxErrorRate {
		nm.alertManager.TriggerAlert("high_error_rate", string(AlertSeverityHigh), "Error rate exceeded threshold")
	}
}

// TriggerAlert triggers an alert
func (am *AlertManager) TriggerAlert(alertType, severity, message string) {
	// Implementation would create and process alerts
	// For now, this is a placeholder
}

// Public API methods

// GetConnectionMetrics returns connection metrics
func (nm *NetworkMonitor) GetConnectionMetrics() *ConnectionMetrics {
	nm.connectionMetrics.mu.RLock()
	defer nm.connectionMetrics.mu.RUnlock()

	// Return a copy
	metrics := *nm.connectionMetrics
	return &metrics
}

// GetMessageMetrics returns message metrics
func (nm *NetworkMonitor) GetMessageMetrics() *MessageMetrics {
	nm.messageMetrics.mu.RLock()
	defer nm.messageMetrics.mu.RUnlock()

	// Return a copy
	metrics := *nm.messageMetrics
	// Copy maps
	metrics.MessagesByType = make(map[string]int64)
	metrics.MessagesByProtocol = make(map[string]int64)
	for k, v := range nm.messageMetrics.MessagesByType {
		metrics.MessagesByType[k] = v
	}
	for k, v := range nm.messageMetrics.MessagesByProtocol {
		metrics.MessagesByProtocol[k] = v
	}
	return &metrics
}

// GetPerformanceMetrics returns performance metrics
func (nm *NetworkMonitor) GetPerformanceMetrics() *PerformanceMetrics {
	nm.performanceMetrics.mu.RLock()
	defer nm.performanceMetrics.mu.RUnlock()

	// Return a copy
	metrics := *nm.performanceMetrics
	return &metrics
}

// GetSecurityMetrics returns security metrics
func (nm *NetworkMonitor) GetSecurityMetrics() *SecurityMetrics {
	nm.securityMetrics.mu.RLock()
	defer nm.securityMetrics.mu.RUnlock()

	// Return a copy
	metrics := *nm.securityMetrics
	return &metrics
}
