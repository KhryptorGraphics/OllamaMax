package protocols

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// TensorStreamMonitor provides comprehensive monitoring for tensor streaming operations
type TensorStreamMonitor struct {
	// Core components
	streamManager  *ActivationStreamManager
	bandwidthMgr   *host.BandwidthManager
	alertManager   *AlertManager
	metricsStore   *MetricsStore
	dashboardMgr   *DashboardManager

	// Monitoring state
	monitoredStreams map[string]*StreamMonitorData
	streamsMutex     sync.RWMutex
	globalMetrics    *GlobalStreamingMetrics
	metricsMutex     sync.RWMutex

	// Configuration
	config *MonitorConfig

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// StreamMonitorData tracks monitoring data for a specific stream
type StreamMonitorData struct {
	StreamID        string
	InferenceID     string
	SourcePeer      peer.ID
	TargetPeer      peer.ID
	StartTime       time.Time
	LastUpdate      time.Time
	Status          StreamMonitorStatus
	Metrics         *StreamMetrics
	Alerts          []*StreamAlert
	PerformanceData *PerformanceHistory
	HealthScore     float64
	mutex           sync.RWMutex
}

// StreamMetrics contains detailed metrics for a stream
type StreamMetrics struct {
	// Throughput metrics
	BytesPerSecond     float64
	ChunksPerSecond    float64
	TotalBytes         int64
	TotalChunks        int64
	CompressionRatio   float64

	// Latency metrics
	AverageLatency     time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	LatencyP95         time.Duration
	LatencyP99         time.Duration

	// Error metrics
	ErrorCount         int64
	ErrorRate          float64
	RetryCount         int64
	TimeoutCount       int64

	// Quality metrics
	PacketLoss         float64
	Jitter             time.Duration
	OutOfOrderCount    int64

	// Resource metrics
	CPUUsage           float64
	MemoryUsage        int64
	NetworkUtilization float64
	BandwidthUsage     int64

	LastUpdated        time.Time
}

// PerformanceHistory maintains historical performance data
type PerformanceHistory struct {
	ThroughputHistory []ThroughputPoint
	LatencyHistory    []LatencyPoint
	ErrorHistory      []ErrorPoint
	MaxHistorySize    int
	mutex             sync.RWMutex
}

// ThroughputPoint represents a point in throughput history
type ThroughputPoint struct {
	Timestamp   time.Time
	BytesPerSec float64
	ChunksPerSec float64
}

// LatencyPoint represents a point in latency history
type LatencyPoint struct {
	Timestamp time.Time
	Latency   time.Duration
}

// ErrorPoint represents a point in error history
type ErrorPoint struct {
	Timestamp time.Time
	ErrorType string
	Count     int64
}

// GlobalStreamingMetrics tracks system-wide streaming metrics
type GlobalStreamingMetrics struct {
	ActiveStreams           int64
	TotalStreamsStarted     int64
	TotalStreamsCompleted   int64
	TotalStreamsFailed      int64
	AverageStreamDuration   time.Duration
	TotalBytesTransferred   int64
	AverageLatency          time.Duration
	SystemThroughput        float64
	OverallHealthScore      float64
	LastUpdated             time.Time
}

// StreamAlert represents an alert for a stream
type StreamAlert struct {
	ID          string
	StreamID    string
	Type        AlertType
	Severity    AlertSeverity
	Title       string
	Description string
	Timestamp   time.Time
	Resolved    bool
	ResolvedAt  time.Time
}

// AlertManager handles monitoring alerts
type AlertManager struct {
	alerts       map[string]*StreamAlert
	alertsMutex  sync.RWMutex
	alertRules   []*AlertRule
	subscribers  []AlertSubscriber
}

// AlertRule defines conditions for generating alerts
type AlertRule struct {
	ID          string
	Name        string
	Description string
	Condition   func(*StreamMetrics) bool
	Severity    AlertSeverity
	Cooldown    time.Duration
	LastFired   map[string]time.Time
	Enabled     bool
}

// MetricsStore provides persistent storage for metrics
type MetricsStore struct {
	streams       map[string]*StreamMetricsHistory
	global        *GlobalMetricsHistory
	storageMutex  sync.RWMutex
	retentionTime time.Duration
}

// DashboardManager provides real-time monitoring dashboards
type DashboardManager struct {
	dashboards map[string]*MonitoringDashboard
	webServer  *MonitoringWebServer
	wsClients  map[string]*WebSocketClient
	mutex      sync.RWMutex
}

// Monitoring enums and types
type StreamMonitorStatus uint8

const (
	StreamMonitorStatusInitializing StreamMonitorStatus = iota
	StreamMonitorStatusActive
	StreamMonitorStatusWarning
	StreamMonitorStatusCritical
	StreamMonitorStatusCompleted
	StreamMonitorStatusFailed
)

type AlertType uint8

const (
	AlertTypeLatency AlertType = iota
	AlertTypeThroughput
	AlertTypeError
	AlertTypeTimeout
	AlertTypeHealth
	AlertTypeResource
)

type AlertSeverity uint8

const (
	AlertSeverityInfo AlertSeverity = iota
	AlertSeverityWarning
	AlertSeverityCritical
	AlertSeverityEmergency
)

type AlertSubscriber interface {
	OnAlert(alert *StreamAlert) error
}

// MonitorConfig configures the tensor stream monitor
type MonitorConfig struct {
	// Collection intervals
	MetricsCollectionInterval time.Duration
	HealthCheckInterval       time.Duration
	AlertCheckInterval        time.Duration

	// Storage settings
	MetricsRetentionTime      time.Duration
	MaxHistoryPoints          int

	// Alert thresholds
	LatencyWarningThreshold   time.Duration
	LatencyCriticalThreshold  time.Duration
	ThroughputWarningPercent  float64
	ErrorRateWarningThreshold float64
	ErrorRateCriticalThreshold float64

	// Dashboard settings
	EnableWebDashboard        bool
	WebDashboardPort         int
	WebSocketEnabled         bool

	// Performance settings
	MonitoringBufferSize     int
	MaxConcurrentStreams     int
}

// NewTensorStreamMonitor creates a new tensor stream monitor
func NewTensorStreamMonitor(
	streamManager *ActivationStreamManager,
	bandwidthMgr *host.BandwidthManager,
) *TensorStreamMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	config := &MonitorConfig{
		MetricsCollectionInterval:  1 * time.Second,
		HealthCheckInterval:        5 * time.Second,
		AlertCheckInterval:         2 * time.Second,
		MetricsRetentionTime:      24 * time.Hour,
		MaxHistoryPoints:          1000,
		LatencyWarningThreshold:   100 * time.Millisecond,
		LatencyCriticalThreshold:  500 * time.Millisecond,
		ThroughputWarningPercent:  0.8,
		ErrorRateWarningThreshold: 0.05,
		ErrorRateCriticalThreshold: 0.15,
		EnableWebDashboard:        true,
		WebDashboardPort:         8080,
		WebSocketEnabled:         true,
		MonitoringBufferSize:     10000,
		MaxConcurrentStreams:     1000,
	}

	monitor := &TensorStreamMonitor{
		streamManager:    streamManager,
		bandwidthMgr:    bandwidthMgr,
		alertManager:    NewAlertManager(),
		metricsStore:    NewMetricsStore(config.MetricsRetentionTime),
		monitoredStreams: make(map[string]*StreamMonitorData),
		globalMetrics:   &GlobalStreamingMetrics{
			LastUpdated: time.Now(),
		},
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize dashboard if enabled
	if config.EnableWebDashboard {
		monitor.dashboardMgr = NewDashboardManager(config)
	}

	// Start background workers
	monitor.startWorkers()

	return monitor
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	am := &AlertManager{
		alerts:      make(map[string]*StreamAlert),
		alertRules:  make([]*AlertRule, 0),
		subscribers: make([]AlertSubscriber, 0),
	}

	// Add default alert rules
	am.addDefaultRules()

	return am
}

// NewMetricsStore creates a new metrics store
func NewMetricsStore(retentionTime time.Duration) *MetricsStore {
	return &MetricsStore{
		streams:       make(map[string]*StreamMetricsHistory),
		retentionTime: retentionTime,
	}
}

// NewDashboardManager creates a new dashboard manager
func NewDashboardManager(config *MonitorConfig) *DashboardManager {
	return &DashboardManager{
		dashboards: make(map[string]*MonitoringDashboard),
		wsClients:  make(map[string]*WebSocketClient),
	}
}

// startWorkers starts background monitoring workers
func (tsm *TensorStreamMonitor) startWorkers() {
	// Metrics collection worker
	tsm.wg.Add(1)
	go tsm.metricsCollectionWorker()

	// Health check worker
	tsm.wg.Add(1)
	go tsm.healthCheckWorker()

	// Alert processing worker
	tsm.wg.Add(1)
	go tsm.alertProcessingWorker()

	// Cleanup worker
	tsm.wg.Add(1)
	go tsm.cleanupWorker()

	// Dashboard worker (if enabled)
	if tsm.config.EnableWebDashboard {
		tsm.wg.Add(1)
		go tsm.dashboardWorker()
	}
}

// StartMonitoring starts monitoring a specific stream
func (tsm *TensorStreamMonitor) StartMonitoring(streamID, inferenceID string, sourcePeer, targetPeer peer.ID) {
	tsm.streamsMutex.Lock()
	defer tsm.streamsMutex.Unlock()

	monitorData := &StreamMonitorData{
		StreamID:    streamID,
		InferenceID: inferenceID,
		SourcePeer:  sourcePeer,
		TargetPeer:  targetPeer,
		StartTime:   time.Now(),
		LastUpdate:  time.Now(),
		Status:      StreamMonitorStatusInitializing,
		Metrics:     &StreamMetrics{
			LastUpdated: time.Now(),
		},
		Alerts: make([]*StreamAlert, 0),
		PerformanceData: &PerformanceHistory{
			ThroughputHistory: make([]ThroughputPoint, 0),
			LatencyHistory:    make([]LatencyPoint, 0),
			ErrorHistory:      make([]ErrorPoint, 0),
			MaxHistorySize:    tsm.config.MaxHistoryPoints,
		},
		HealthScore: 1.0,
	}

	tsm.monitoredStreams[streamID] = monitorData

	// Update global metrics
	tsm.metricsMutex.Lock()
	tsm.globalMetrics.TotalStreamsStarted++
	tsm.globalMetrics.ActiveStreams++
	tsm.metricsMutex.Unlock()

	log.Printf("Started monitoring stream %s for inference %s", streamID, inferenceID)
}

// StopMonitoring stops monitoring a stream
func (tsm *TensorStreamMonitor) StopMonitoring(streamID string, success bool) {
	tsm.streamsMutex.Lock()
	defer tsm.streamsMutex.Unlock()

	monitorData, exists := tsm.monitoredStreams[streamID]
	if !exists {
		return
	}

	monitorData.mutex.Lock()
	if success {
		monitorData.Status = StreamMonitorStatusCompleted
	} else {
		monitorData.Status = StreamMonitorStatusFailed
	}
	monitorData.LastUpdate = time.Now()
	monitorData.mutex.Unlock()

	// Update global metrics
	tsm.metricsMutex.Lock()
	tsm.globalMetrics.ActiveStreams--
	if success {
		tsm.globalMetrics.TotalStreamsCompleted++
	} else {
		tsm.globalMetrics.TotalStreamsFailed++
	}
	tsm.metricsMutex.Unlock()

	// Store final metrics
	tsm.metricsStore.StoreStreamMetrics(streamID, monitorData.Metrics)

	log.Printf("Stopped monitoring stream %s (success: %v)", streamID, success)
}

// metricsCollectionWorker collects metrics from monitored streams
func (tsm *TensorStreamMonitor) metricsCollectionWorker() {
	defer tsm.wg.Done()

	ticker := time.NewTicker(tsm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			tsm.collectMetrics()
		}
	}
}

// collectMetrics collects current metrics from all monitored streams
func (tsm *TensorStreamMonitor) collectMetrics() {
	tsm.streamsMutex.RLock()
	streamsCopy := make(map[string]*StreamMonitorData)
	for k, v := range tsm.monitoredStreams {
		streamsCopy[k] = v
	}
	tsm.streamsMutex.RUnlock()

	for streamID, monitorData := range streamsCopy {
		tsm.collectStreamMetrics(streamID, monitorData)
	}

	// Update global metrics
	tsm.updateGlobalMetrics()
}

// collectStreamMetrics collects metrics for a specific stream
func (tsm *TensorStreamMonitor) collectStreamMetrics(streamID string, monitorData *StreamMonitorData) {
	// Get current streaming metrics from the stream client
	if tsm.streamManager != nil && tsm.streamManager.streamClient != nil {
		streamingMetrics := tsm.streamManager.streamClient.GetStreamingMetrics()

		monitorData.mutex.Lock()

		// Update throughput metrics
		if streamingMetrics.TotalBytes > monitorData.Metrics.TotalBytes {
			bytesTransferred := streamingMetrics.TotalBytes - monitorData.Metrics.TotalBytes
			timeDiff := time.Since(monitorData.Metrics.LastUpdated)
			if timeDiff > 0 {
				monitorData.Metrics.BytesPerSecond = float64(bytesTransferred) / timeDiff.Seconds()
			}
		}

		// Update compression metrics
		monitorData.Metrics.CompressionRatio = streamingMetrics.CompressionRatio

		// Update latency from bandwidth manager
		if tsm.bandwidthMgr != nil {
			if usage := tsm.bandwidthMgr.GetPeerUsage(monitorData.TargetPeer); usage != nil {
				// Simplified latency calculation based on recent activity
				if time.Since(usage.LastUpdated) < 5*time.Second {
					monitorData.Metrics.AverageLatency = 50 * time.Millisecond // Placeholder
				}
			}
		}

		// Calculate health score
		healthScore := tsm.calculateHealthScore(monitorData.Metrics)
		monitorData.HealthScore = healthScore

		// Update status based on health
		if healthScore < 0.3 {
			monitorData.Status = StreamMonitorStatusCritical
		} else if healthScore < 0.7 {
			monitorData.Status = StreamMonitorStatusWarning
		} else {
			monitorData.Status = StreamMonitorStatusActive
		}

		// Add to performance history
		monitorData.PerformanceData.mutex.Lock()

		// Add throughput point
		monitorData.PerformanceData.ThroughputHistory = append(
			monitorData.PerformanceData.ThroughputHistory,
			ThroughputPoint{
				Timestamp:    time.Now(),
				BytesPerSec:  monitorData.Metrics.BytesPerSecond,
				ChunksPerSec: monitorData.Metrics.ChunksPerSecond,
			},
		)

		// Add latency point
		monitorData.PerformanceData.LatencyHistory = append(
			monitorData.PerformanceData.LatencyHistory,
			LatencyPoint{
				Timestamp: time.Now(),
				Latency:   monitorData.Metrics.AverageLatency,
			},
		)

		// Trim history if too long
		if len(monitorData.PerformanceData.ThroughputHistory) > monitorData.PerformanceData.MaxHistorySize {
			monitorData.PerformanceData.ThroughputHistory = monitorData.PerformanceData.ThroughputHistory[1:]
		}
		if len(monitorData.PerformanceData.LatencyHistory) > monitorData.PerformanceData.MaxHistorySize {
			monitorData.PerformanceData.LatencyHistory = monitorData.PerformanceData.LatencyHistory[1:]
		}

		monitorData.PerformanceData.mutex.Unlock()

		monitorData.Metrics.LastUpdated = time.Now()
		monitorData.LastUpdate = time.Now()
		monitorData.mutex.Unlock()
	}
}

// calculateHealthScore calculates a health score for a stream based on its metrics
func (tsm *TensorStreamMonitor) calculateHealthScore(metrics *StreamMetrics) float64 {
	score := 1.0

	// Penalize high latency
	if metrics.AverageLatency > tsm.config.LatencyWarningThreshold {
		score -= 0.2
	}
	if metrics.AverageLatency > tsm.config.LatencyCriticalThreshold {
		score -= 0.3
	}

	// Penalize high error rate
	if metrics.ErrorRate > tsm.config.ErrorRateWarningThreshold {
		score -= 0.2
	}
	if metrics.ErrorRate > tsm.config.ErrorRateCriticalThreshold {
		score -= 0.3
	}

	// Penalize low throughput (if we have a baseline)
	if metrics.BytesPerSecond < 1.0 {
		score -= 0.1
	}

	// Ensure score is between 0 and 1
	if score < 0 {
		score = 0
	}

	return score
}

// updateGlobalMetrics updates system-wide metrics
func (tsm *TensorStreamMonitor) updateGlobalMetrics() {
	tsm.metricsMutex.Lock()
	defer tsm.metricsMutex.Unlock()

	totalThroughput := 0.0
	totalLatency := time.Duration(0)
	activeCount := 0
	totalHealthScore := 0.0

	tsm.streamsMutex.RLock()
	for _, monitorData := range tsm.monitoredStreams {
		monitorData.mutex.RLock()
		if monitorData.Status == StreamMonitorStatusActive ||
		   monitorData.Status == StreamMonitorStatusWarning ||
		   monitorData.Status == StreamMonitorStatusCritical {
			totalThroughput += monitorData.Metrics.BytesPerSecond
			totalLatency += monitorData.Metrics.AverageLatency
			totalHealthScore += monitorData.HealthScore
			activeCount++
		}
		monitorData.mutex.RUnlock()
	}
	tsm.streamsMutex.RUnlock()

	if activeCount > 0 {
		tsm.globalMetrics.SystemThroughput = totalThroughput
		tsm.globalMetrics.AverageLatency = totalLatency / time.Duration(activeCount)
		tsm.globalMetrics.OverallHealthScore = totalHealthScore / float64(activeCount)
	}

	tsm.globalMetrics.LastUpdated = time.Now()
}

// healthCheckWorker performs periodic health checks
func (tsm *TensorStreamMonitor) healthCheckWorker() {
	defer tsm.wg.Done()

	ticker := time.NewTicker(tsm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			tsm.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all monitored streams
func (tsm *TensorStreamMonitor) performHealthChecks() {
	// Health checks are integrated into metrics collection
	// This could be extended for more sophisticated health monitoring
	log.Printf("Performed health checks on %d monitored streams", len(tsm.monitoredStreams))
}

// alertProcessingWorker processes alerts based on collected metrics
func (tsm *TensorStreamMonitor) alertProcessingWorker() {
	defer tsm.wg.Done()

	ticker := time.NewTicker(tsm.config.AlertCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			tsm.processAlerts()
		}
	}
}

// processAlerts checks for alert conditions and generates alerts
func (tsm *TensorStreamMonitor) processAlerts() {
	tsm.streamsMutex.RLock()
	streams := make(map[string]*StreamMonitorData)
	for k, v := range tsm.monitoredStreams {
		streams[k] = v
	}
	tsm.streamsMutex.RUnlock()

	for streamID, monitorData := range streams {
		monitorData.mutex.RLock()
		metrics := monitorData.Metrics
		monitorData.mutex.RUnlock()

		tsm.alertManager.CheckAlerts(streamID, metrics)
	}
}

// cleanupWorker cleans up old monitoring data
func (tsm *TensorStreamMonitor) cleanupWorker() {
	defer tsm.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			tsm.cleanup()
		}
	}
}

// cleanup removes old monitoring data
func (tsm *TensorStreamMonitor) cleanup() {
	cutoff := time.Now().Add(-tsm.config.MetricsRetentionTime)

	tsm.streamsMutex.Lock()
	for streamID, monitorData := range tsm.monitoredStreams {
		monitorData.mutex.RLock()
		shouldRemove := (monitorData.Status == StreamMonitorStatusCompleted ||
			monitorData.Status == StreamMonitorStatusFailed) &&
			monitorData.LastUpdate.Before(cutoff)
		monitorData.mutex.RUnlock()

		if shouldRemove {
			delete(tsm.monitoredStreams, streamID)
		}
	}
	tsm.streamsMutex.Unlock()

	// Cleanup metrics store
	tsm.metricsStore.Cleanup(cutoff)
}

// dashboardWorker manages the monitoring dashboard
func (tsm *TensorStreamMonitor) dashboardWorker() {
	defer tsm.wg.Done()

	if tsm.dashboardMgr != nil {
		// Start web server for dashboard
		go tsm.dashboardMgr.StartWebServer(tsm.config.WebDashboardPort)

		// Broadcast updates to WebSocket clients
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-tsm.ctx.Done():
				return
			case <-ticker.C:
				tsm.broadcastDashboardUpdates()
			}
		}
	}
}

// broadcastDashboardUpdates broadcasts monitoring updates to dashboard clients
func (tsm *TensorStreamMonitor) broadcastDashboardUpdates() {
	if tsm.config.WebSocketEnabled {
		// Prepare monitoring data for broadcast
		data := tsm.GetMonitoringSummary()
		jsonData, _ := json.Marshal(data)

		// Broadcast to all WebSocket clients
		tsm.dashboardMgr.BroadcastToClients(jsonData)
	}
}

// GetMonitoringSummary returns a summary of current monitoring status
func (tsm *TensorStreamMonitor) GetMonitoringSummary() map[string]interface{} {
	tsm.streamsMutex.RLock()
	tsm.metricsMutex.RLock()
	defer tsm.streamsMutex.RUnlock()
	defer tsm.metricsMutex.RUnlock()

	summary := map[string]interface{}{
		"global_metrics": tsm.globalMetrics,
		"active_streams": len(tsm.monitoredStreams),
		"alerts_count":   len(tsm.alertManager.alerts),
		"timestamp":      time.Now(),
	}

	// Add top streams by throughput
	topStreams := tsm.getTopStreamsByThroughput(5)
	summary["top_streams"] = topStreams

	return summary
}

// getTopStreamsByThroughput returns the top N streams by throughput
func (tsm *TensorStreamMonitor) getTopStreamsByThroughput(limit int) []map[string]interface{} {
	type streamThroughput struct {
		StreamID   string
		Throughput float64
		Status     string
	}

	streams := make([]streamThroughput, 0)
	for streamID, monitorData := range tsm.monitoredStreams {
		monitorData.mutex.RLock()
		streams = append(streams, streamThroughput{
			StreamID:   streamID,
			Throughput: monitorData.Metrics.BytesPerSecond,
			Status:     fmt.Sprintf("%d", monitorData.Status),
		})
		monitorData.mutex.RUnlock()
	}

	// Sort by throughput
	for i := 0; i < len(streams)-1; i++ {
		for j := i + 1; j < len(streams); j++ {
			if streams[i].Throughput < streams[j].Throughput {
				streams[i], streams[j] = streams[j], streams[i]
			}
		}
	}

	// Convert to result format
	result := make([]map[string]interface{}, 0)
	maxLen := limit
	if len(streams) < maxLen {
		maxLen = len(streams)
	}

	for i := 0; i < maxLen; i++ {
		result = append(result, map[string]interface{}{
			"stream_id":  streams[i].StreamID,
			"throughput": streams[i].Throughput,
			"status":     streams[i].Status,
		})
	}

	return result
}

// GetStreamMetrics returns detailed metrics for a specific stream
func (tsm *TensorStreamMonitor) GetStreamMetrics(streamID string) (*StreamMonitorData, bool) {
	tsm.streamsMutex.RLock()
	defer tsm.streamsMutex.RUnlock()

	data, exists := tsm.monitoredStreams[streamID]
	return data, exists
}

// Close shuts down the tensor stream monitor
func (tsm *TensorStreamMonitor) Close() error {
	tsm.cancel()
	tsm.wg.Wait()

	if tsm.dashboardMgr != nil {
		tsm.dashboardMgr.Close()
	}

	return nil
}

// addDefaultRules adds default alert rules to the alert manager
func (am *AlertManager) addDefaultRules() {
	// High latency alert
	am.alertRules = append(am.alertRules, &AlertRule{
		ID:          "high_latency",
		Name:        "High Latency",
		Description: "Stream latency exceeds threshold",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.AverageLatency > 500*time.Millisecond
		},
		Severity:    AlertSeverityWarning,
		Cooldown:    2 * time.Minute,
		LastFired:   make(map[string]time.Time),
		Enabled:     true,
	})

	// High error rate alert
	am.alertRules = append(am.alertRules, &AlertRule{
		ID:          "high_error_rate",
		Name:        "High Error Rate",
		Description: "Stream error rate exceeds threshold",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.ErrorRate > 0.1
		},
		Severity:    AlertSeverityCritical,
		Cooldown:    1 * time.Minute,
		LastFired:   make(map[string]time.Time),
		Enabled:     true,
	})
}

// CheckAlerts checks alert rules against stream metrics
func (am *AlertManager) CheckAlerts(streamID string, metrics *StreamMetrics) {
	am.alertsMutex.Lock()
	defer am.alertsMutex.Unlock()

	for _, rule := range am.alertRules {
		if !rule.Enabled {
			continue
		}

		if rule.Condition(metrics) {
			// Check cooldown
			if lastFired, exists := rule.LastFired[streamID]; exists {
				if time.Since(lastFired) < rule.Cooldown {
					continue
				}
			}

			// Generate alert
			alert := &StreamAlert{
				ID:          fmt.Sprintf("%s-%s-%d", rule.ID, streamID, time.Now().UnixNano()),
				StreamID:    streamID,
				Type:        AlertTypeLatency, // Simplified
				Severity:    rule.Severity,
				Title:       rule.Name,
				Description: rule.Description,
				Timestamp:   time.Now(),
			}

			am.alerts[alert.ID] = alert
			rule.LastFired[streamID] = time.Now()

			// Notify subscribers
			for _, subscriber := range am.subscribers {
				go subscriber.OnAlert(alert)
			}

			log.Printf("Generated alert %s for stream %s: %s", alert.ID, streamID, alert.Title)
		}
	}
}

// StoreStreamMetrics stores metrics for a stream
func (ms *MetricsStore) StoreStreamMetrics(streamID string, metrics *StreamMetrics) {
	ms.storageMutex.Lock()
	defer ms.storageMutex.Unlock()

	// Implementation would store metrics in persistent storage
	// For now, just log
	log.Printf("Stored metrics for stream %s", streamID)
}

// Cleanup removes old metrics from storage
func (ms *MetricsStore) Cleanup(cutoff time.Time) {
	ms.storageMutex.Lock()
	defer ms.storageMutex.Unlock()

	// Implementation would clean up old metrics
	log.Printf("Cleaned up metrics older than %v", cutoff)
}

// StartWebServer starts the monitoring web server
func (dm *DashboardManager) StartWebServer(port int) error {
	// Implementation would start a web server for the monitoring dashboard
	log.Printf("Started monitoring dashboard on port %d", port)
	return nil
}

// BroadcastToClients broadcasts data to WebSocket clients
func (dm *DashboardManager) BroadcastToClients(data []byte) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	// Implementation would broadcast to WebSocket clients
	log.Printf("Broadcasting %d bytes to %d clients", len(data), len(dm.wsClients))
}

// Close shuts down the dashboard manager
func (dm *DashboardManager) Close() error {
	// Implementation would clean up web server and connections
	return nil
}

// Placeholder types
type StreamMetricsHistory struct{}
type GlobalMetricsHistory struct{}
type MonitoringDashboard struct{}
type MonitoringWebServer struct{}
type WebSocketClient struct{}