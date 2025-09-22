package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// TensorStreamMonitor provides comprehensive performance monitoring for tensor streaming operations
type TensorStreamMonitor struct {
	metrics           map[string]*StreamMetrics
	metricsMutex      sync.RWMutex
	analyzer          *StreamingPerformanceAnalyzer
	alertManager      *StreamAlertManager
	historicalData    *HistoricalMetrics
	config            *MonitorConfig
	ctx               context.Context
	cancel            context.CancelFunc
	lastReport        time.Time
}

// StreamMetrics contains detailed metrics for a tensor stream
type StreamMetrics struct {
	StreamID          string                 `json:"stream_id"`
	InferenceID       string                 `json:"inference_id"`
	SourcePeer        peer.ID                `json:"source_peer"`
	TargetPeer        peer.ID                `json:"target_peer"`
	StartTime         time.Time              `json:"start_time"`
	EndTime           time.Time              `json:"end_time"`
	Duration          time.Duration          `json:"duration"`
	TotalBytes        int64                  `json:"total_bytes"`
	CompressedBytes   int64                  `json:"compressed_bytes"`
	ChunksTransferred int32                  `json:"chunks_transferred"`
	TotalChunks       int32                  `json:"total_chunks"`
	ThroughputMBps    float64                `json:"throughput_mbps"`
	LatencyMs         float64                `json:"latency_ms"`
	CompressionRatio  float64                `json:"compression_ratio"`
	ErrorCount        int                    `json:"error_count"`
	RetryCount        int                    `json:"retry_count"`
	PacketLoss        float64                `json:"packet_loss"`
	NetworkUtilization float64               `json:"network_utilization"`
	CPUUtilization    float64                `json:"cpu_utilization"`
	MemoryUsage       int64                  `json:"memory_usage"`
	QualityScore      float64                `json:"quality_score"`
	Bottlenecks       []string               `json:"bottlenecks"`
	CustomMetrics     map[string]interface{} `json:"custom_metrics"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// StreamingPerformanceAnalyzer analyzes streaming performance and identifies bottlenecks
type StreamingPerformanceAnalyzer struct {
	analysisRules     map[string]*AnalysisRule
	bottleneckDetector *BottleneckDetector
	trendAnalyzer     *TrendAnalyzer
	mutex             sync.RWMutex
}

// AnalysisRule defines a rule for performance analysis
type AnalysisRule struct {
	Name        string
	Description string
	Condition   func(*StreamMetrics) bool
	Severity    AlertSeverity
	Action      func(*StreamMetrics) *PerformanceInsight
	Enabled     bool
}

// PerformanceInsight provides actionable insights about stream performance
type PerformanceInsight struct {
	Type          InsightType               `json:"type"`
	Severity      AlertSeverity             `json:"severity"`
	Title         string                    `json:"title"`
	Description   string                    `json:"description"`
	Recommendation string                   `json:"recommendation"`
	Impact        string                    `json:"impact"`
	Metrics       map[string]interface{}    `json:"metrics"`
	Timestamp     time.Time                 `json:"timestamp"`
}

// InsightType categorizes performance insights
type InsightType uint8

const (
	InsightTypeBottleneck InsightType = iota
	InsightTypeOptimization
	InsightTypeAnomaly
	InsightTypeTrend
	InsightTypeAlert
)

// AlertSeverity defines the severity level of alerts
type AlertSeverity uint8

const (
	AlertSeverityInfo AlertSeverity = iota
	AlertSeverityWarning
	AlertSeverityError
	AlertSeverityCritical
)

// BottleneckDetector identifies performance bottlenecks in streaming
type BottleneckDetector struct {
	detectionRules map[string]*BottleneckRule
	thresholds     *BottleneckThresholds
}

// BottleneckRule defines a rule for bottleneck detection
type BottleneckRule struct {
	Name        string
	Component   string // network, cpu, memory, compression, etc.
	Condition   func(*StreamMetrics) bool
	Impact      float64 // 0.0 to 1.0
	Description string
}

// BottleneckThresholds defines thresholds for bottleneck detection
type BottleneckThresholds struct {
	MinThroughputMBps    float64
	MaxLatencyMs         float64
	MaxCPUUtilization    float64
	MaxMemoryUsageMB     int64
	MinCompressionRatio  float64
	MaxErrorRate         float64
	MaxPacketLoss        float64
}

// TrendAnalyzer analyzes performance trends over time
type TrendAnalyzer struct {
	trendData    map[string]*TrendData
	trendMutex   sync.RWMutex
	windowSize   time.Duration
	sampleRate   time.Duration
}

// TrendData stores trend information for a metric
type TrendData struct {
	MetricName   string
	Values       []float64
	Timestamps   []time.Time
	Trend        TrendDirection
	Slope        float64
	Confidence   float64
	LastUpdated  time.Time
}

// TrendDirection indicates the direction of a trend
type TrendDirection uint8

const (
	TrendDirectionStable TrendDirection = iota
	TrendDirectionIncreasing
	TrendDirectionDecreasing
	TrendDirectionVolatile
)

// StreamAlertManager manages alerts for streaming issues
type StreamAlertManager struct {
	alerts       map[string]*StreamAlert
	alertsMutex  sync.RWMutex
	alertRules   map[string]*AlertRule
	subscribers  []AlertSubscriber
}

// StreamAlert represents an alert about streaming performance
type StreamAlert struct {
	ID          string        `json:"id"`
	StreamID    string        `json:"stream_id"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Metrics     map[string]interface{} `json:"metrics"`
	Timestamp   time.Time     `json:"timestamp"`
	Resolved    bool          `json:"resolved"`
	ResolvedAt  time.Time     `json:"resolved_at,omitempty"`
}

// AlertType categorizes different types of alerts
type AlertType uint8

const (
	AlertTypeThroughput AlertType = iota
	AlertTypeLatency
	AlertTypeError
	AlertTypeBottleneck
	AlertTypeAnomaly
)

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name        string
	Type        AlertType
	Severity    AlertSeverity
	Condition   func(*StreamMetrics) bool
	Description string
	Enabled     bool
	Cooldown    time.Duration
	LastFired   time.Time
}

// AlertSubscriber receives alert notifications
type AlertSubscriber interface {
	OnAlert(alert *StreamAlert) error
}

// HistoricalMetrics stores historical performance data
type HistoricalMetrics struct {
	data        map[string][]*StreamMetrics
	dataMutex   sync.RWMutex
	retention   time.Duration
	maxEntries  int
}

// MonitorConfig configures the tensor stream monitor
type MonitorConfig struct {
	MetricsInterval     time.Duration
	AnalysisInterval    time.Duration
	AlertCheckInterval  time.Duration
	HistoryRetention    time.Duration
	MaxHistoryEntries   int
	EnableTrendAnalysis bool
	EnableAlerting      bool
	EnableInsights      bool
}

// NewTensorStreamMonitor creates a new tensor stream monitor
func NewTensorStreamMonitor(config *MonitorConfig) *TensorStreamMonitor {
	if config == nil {
		config = &MonitorConfig{
			MetricsInterval:     1 * time.Second,
			AnalysisInterval:    10 * time.Second,
			AlertCheckInterval:  5 * time.Second,
			HistoryRetention:    24 * time.Hour,
			MaxHistoryEntries:   10000,
			EnableTrendAnalysis: true,
			EnableAlerting:      true,
			EnableInsights:      true,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	monitor := &TensorStreamMonitor{
		metrics:        make(map[string]*StreamMetrics),
		analyzer:       NewStreamingPerformanceAnalyzer(),
		alertManager:   NewStreamAlertManager(),
		historicalData: NewHistoricalMetrics(config.HistoryRetention, config.MaxHistoryEntries),
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
		lastReport:     time.Now(),
	}
	
	// Start background workers
	go monitor.metricsCollectionWorker()
	go monitor.performanceAnalysisWorker()
	go monitor.alertCheckWorker()
	
	return monitor
}

// NewStreamingPerformanceAnalyzer creates a new performance analyzer
func NewStreamingPerformanceAnalyzer() *StreamingPerformanceAnalyzer {
	analyzer := &StreamingPerformanceAnalyzer{
		analysisRules:      make(map[string]*AnalysisRule),
		bottleneckDetector: NewBottleneckDetector(),
		trendAnalyzer:      NewTrendAnalyzer(),
	}
	
	analyzer.addDefaultAnalysisRules()
	return analyzer
}

// NewBottleneckDetector creates a new bottleneck detector
func NewBottleneckDetector() *BottleneckDetector {
	detector := &BottleneckDetector{
		detectionRules: make(map[string]*BottleneckRule),
		thresholds: &BottleneckThresholds{
			MinThroughputMBps:   10.0,
			MaxLatencyMs:        500.0,
			MaxCPUUtilization:   80.0,
			MaxMemoryUsageMB:    1024,
			MinCompressionRatio: 0.3,
			MaxErrorRate:        0.05,
			MaxPacketLoss:       0.01,
		},
	}
	
	detector.addDefaultRules()
	return detector
}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{
		trendData:  make(map[string]*TrendData),
		windowSize: 10 * time.Minute,
		sampleRate: 30 * time.Second,
	}
}

// NewStreamAlertManager creates a new alert manager
func NewStreamAlertManager() *StreamAlertManager {
	manager := &StreamAlertManager{
		alerts:      make(map[string]*StreamAlert),
		alertRules:  make(map[string]*AlertRule),
		subscribers: make([]AlertSubscriber, 0),
	}
	
	manager.addDefaultAlertRules()
	return manager
}

// NewHistoricalMetrics creates a new historical metrics store
func NewHistoricalMetrics(retention time.Duration, maxEntries int) *HistoricalMetrics {
	return &HistoricalMetrics{
		data:       make(map[string][]*StreamMetrics),
		retention:  retention,
		maxEntries: maxEntries,
	}
}

// addDefaultAnalysisRules adds default performance analysis rules
func (spa *StreamingPerformanceAnalyzer) addDefaultAnalysisRules() {
	// Low throughput rule
	spa.analysisRules["low_throughput"] = &AnalysisRule{
		Name:        "Low Throughput Detection",
		Description: "Detects streams with throughput below expected levels",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.ThroughputMBps < 10.0
		},
		Severity: AlertSeverityWarning,
		Action: func(metrics *StreamMetrics) *PerformanceInsight {
			return &PerformanceInsight{
				Type:           InsightTypeBottleneck,
				Severity:       AlertSeverityWarning,
				Title:          "Low Throughput Detected",
				Description:    fmt.Sprintf("Stream %s has low throughput: %.2f MB/s", metrics.StreamID, metrics.ThroughputMBps),
				Recommendation: "Consider increasing chunk size or checking network conditions",
				Impact:         "May cause inference delays",
				Metrics: map[string]interface{}{
					"throughput_mbps": metrics.ThroughputMBps,
					"target_mbps":     10.0,
				},
				Timestamp: time.Now(),
			}
		},
		Enabled: true,
	}
	
	// High latency rule
	spa.analysisRules["high_latency"] = &AnalysisRule{
		Name:        "High Latency Detection",
		Description: "Detects streams with high latency",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.LatencyMs > 500.0
		},
		Severity: AlertSeverityError,
		Action: func(metrics *StreamMetrics) *PerformanceInsight {
			return &PerformanceInsight{
				Type:           InsightTypeBottleneck,
				Severity:       AlertSeverityError,
				Title:          "High Latency Detected",
				Description:    fmt.Sprintf("Stream %s has high latency: %.2f ms", metrics.StreamID, metrics.LatencyMs),
				Recommendation: "Check network path and consider reducing chunk size",
				Impact:         "Significantly impacts inference performance",
				Metrics: map[string]interface{}{
					"latency_ms": metrics.LatencyMs,
					"target_ms":  500.0,
				},
				Timestamp: time.Now(),
			}
		},
		Enabled: true,
	}
}

// addDefaultRules adds default bottleneck detection rules
func (bd *BottleneckDetector) addDefaultRules() {
	// Network bottleneck
	bd.detectionRules["network_bottleneck"] = &BottleneckRule{
		Name:      "Network Bottleneck",
		Component: "network",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.ThroughputMBps < bd.thresholds.MinThroughputMBps ||
				metrics.LatencyMs > bd.thresholds.MaxLatencyMs ||
				metrics.PacketLoss > bd.thresholds.MaxPacketLoss
		},
		Impact:      0.8,
		Description: "Network performance is limiting streaming throughput",
	}
	
	// CPU bottleneck
	bd.detectionRules["cpu_bottleneck"] = &BottleneckRule{
		Name:      "CPU Bottleneck",
		Component: "cpu",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.CPUUtilization > bd.thresholds.MaxCPUUtilization
		},
		Impact:      0.6,
		Description: "High CPU utilization is affecting streaming performance",
	}
	
	// Compression bottleneck
	bd.detectionRules["compression_bottleneck"] = &BottleneckRule{
		Name:      "Compression Bottleneck",
		Component: "compression",
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.CompressionRatio < bd.thresholds.MinCompressionRatio
		},
		Impact:      0.4,
		Description: "Poor compression ratio is increasing data transfer overhead",
	}
}

// addDefaultAlertRules adds default alert rules
func (sam *StreamAlertManager) addDefaultAlertRules() {
	// Throughput alert
	sam.alertRules["throughput_alert"] = &AlertRule{
		Name:     "Throughput Alert",
		Type:     AlertTypeThroughput,
		Severity: AlertSeverityWarning,
		Condition: func(metrics *StreamMetrics) bool {
			return metrics.ThroughputMBps < 5.0
		},
		Description: "Stream throughput is below acceptable levels",
		Enabled:     true,
		Cooldown:    5 * time.Minute,
	}
	
	// Error rate alert
	sam.alertRules["error_rate_alert"] = &AlertRule{
		Name:     "Error Rate Alert",
		Type:     AlertTypeError,
		Severity: AlertSeverityError,
		Condition: func(metrics *StreamMetrics) bool {
			totalTransfers := float64(metrics.ChunksTransferred)
			if totalTransfers == 0 {
				return false
			}
			errorRate := float64(metrics.ErrorCount) / totalTransfers
			return errorRate > 0.05 // 5% error rate
		},
		Description: "Stream error rate is too high",
		Enabled:     true,
		Cooldown:    2 * time.Minute,
	}
}

// RecordMetrics records metrics for a tensor stream
func (tsm *TensorStreamMonitor) RecordMetrics(metrics *StreamMetrics) {
	tsm.metricsMutex.Lock()
	defer tsm.metricsMutex.Unlock()
	
	metrics.LastUpdated = time.Now()
	tsm.metrics[metrics.StreamID] = metrics
	
	// Store in historical data
	tsm.historicalData.AddMetrics(metrics)
	
	log.Printf("Recorded metrics for stream %s: %.2f MB/s, %.2f ms latency", 
		metrics.StreamID, metrics.ThroughputMBps, metrics.LatencyMs)
}

// AddMetrics adds metrics to historical storage
func (hm *HistoricalMetrics) AddMetrics(metrics *StreamMetrics) {
	hm.dataMutex.Lock()
	defer hm.dataMutex.Unlock()
	
	streamHistory, exists := hm.data[metrics.StreamID]
	if !exists {
		streamHistory = make([]*StreamMetrics, 0)
	}
	
	// Add new metrics
	streamHistory = append(streamHistory, metrics)
	
	// Enforce retention policy
	cutoff := time.Now().Add(-hm.retention)
	filtered := make([]*StreamMetrics, 0)
	for _, m := range streamHistory {
		if m.LastUpdated.After(cutoff) {
			filtered = append(filtered, m)
		}
	}
	
	// Enforce max entries
	if len(filtered) > hm.maxEntries {
		filtered = filtered[len(filtered)-hm.maxEntries:]
	}
	
	hm.data[metrics.StreamID] = filtered
}

// metricsCollectionWorker runs the metrics collection loop
func (tsm *TensorStreamMonitor) metricsCollectionWorker() {
	ticker := time.NewTicker(tsm.config.MetricsInterval)
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

// collectMetrics collects current metrics from active streams
func (tsm *TensorStreamMonitor) collectMetrics() {
	// This would integrate with the actual streaming components
	// For now, this is a placeholder for the collection logic
	log.Printf("Collecting tensor stream metrics...")
}

// performanceAnalysisWorker runs the performance analysis loop
func (tsm *TensorStreamMonitor) performanceAnalysisWorker() {
	ticker := time.NewTicker(tsm.config.AnalysisInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			if tsm.config.EnableInsights {
				tsm.performAnalysis()
			}
		}
	}
}

// performAnalysis performs performance analysis on current metrics
func (tsm *TensorStreamMonitor) performAnalysis() {
	tsm.metricsMutex.RLock()
	metrics := make([]*StreamMetrics, 0, len(tsm.metrics))
	for _, m := range tsm.metrics {
		metrics = append(metrics, m)
	}
	tsm.metricsMutex.RUnlock()
	
	for _, metric := range metrics {
		insights := tsm.analyzer.AnalyzeStream(metric)
		for _, insight := range insights {
			log.Printf("Performance insight for stream %s: %s", metric.StreamID, insight.Title)
		}
	}
}

// AnalyzeStream analyzes a stream's performance and returns insights
func (spa *StreamingPerformanceAnalyzer) AnalyzeStream(metrics *StreamMetrics) []*PerformanceInsight {
	spa.mutex.RLock()
	defer spa.mutex.RUnlock()
	
	insights := make([]*PerformanceInsight, 0)
	
	for _, rule := range spa.analysisRules {
		if rule.Enabled && rule.Condition(metrics) {
			insight := rule.Action(metrics)
			insights = append(insights, insight)
		}
	}
	
	// Detect bottlenecks
	bottlenecks := spa.bottleneckDetector.DetectBottlenecks(metrics)
	for _, bottleneck := range bottlenecks {
		insight := &PerformanceInsight{
			Type:           InsightTypeBottleneck,
			Severity:       AlertSeverityWarning,
			Title:          fmt.Sprintf("%s Bottleneck", bottleneck.Component),
			Description:    bottleneck.Description,
			Recommendation: fmt.Sprintf("Address %s performance issues", bottleneck.Component),
			Impact:         fmt.Sprintf("Performance impact: %.1f%%", bottleneck.Impact*100),
			Timestamp:      time.Now(),
		}
		insights = append(insights, insight)
	}
	
	return insights
}

// DetectBottlenecks detects performance bottlenecks in a stream
func (bd *BottleneckDetector) DetectBottlenecks(metrics *StreamMetrics) []*BottleneckRule {
	bottlenecks := make([]*BottleneckRule, 0)
	
	for _, rule := range bd.detectionRules {
		if rule.Condition(metrics) {
			bottlenecks = append(bottlenecks, rule)
		}
	}
	
	return bottlenecks
}

// alertCheckWorker runs the alert checking loop
func (tsm *TensorStreamMonitor) alertCheckWorker() {
	ticker := time.NewTicker(tsm.config.AlertCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-tsm.ctx.Done():
			return
		case <-ticker.C:
			if tsm.config.EnableAlerting {
				tsm.checkAlerts()
			}
		}
	}
}

// checkAlerts checks for alert conditions
func (tsm *TensorStreamMonitor) checkAlerts() {
	tsm.metricsMutex.RLock()
	metrics := make([]*StreamMetrics, 0, len(tsm.metrics))
	for _, m := range tsm.metrics {
		metrics = append(metrics, m)
	}
	tsm.metricsMutex.RUnlock()
	
	for _, metric := range metrics {
		alerts := tsm.alertManager.CheckAlerts(metric)
		for _, alert := range alerts {
			tsm.alertManager.TriggerAlert(alert)
		}
	}
}

// CheckAlerts checks if any alert rules are triggered for the given metrics
func (sam *StreamAlertManager) CheckAlerts(metrics *StreamMetrics) []*StreamAlert {
	alerts := make([]*StreamAlert, 0)
	
	for _, rule := range sam.alertRules {
		if !rule.Enabled {
			continue
		}
		
		// Check cooldown
		if time.Since(rule.LastFired) < rule.Cooldown {
			continue
		}
		
		if rule.Condition(metrics) {
			alert := &StreamAlert{
				ID:          fmt.Sprintf("alert-%s-%d", metrics.StreamID, time.Now().UnixNano()),
				StreamID:    metrics.StreamID,
				Type:        rule.Type,
				Severity:    rule.Severity,
				Title:       rule.Name,
				Description: rule.Description,
				Metrics: map[string]interface{}{
					"throughput_mbps": metrics.ThroughputMBps,
					"latency_ms":      metrics.LatencyMs,
					"error_count":     metrics.ErrorCount,
				},
				Timestamp: time.Now(),
				Resolved:  false,
			}
			
			alerts = append(alerts, alert)
			rule.LastFired = time.Now()
		}
	}
	
	return alerts
}

// TriggerAlert triggers an alert and notifies subscribers
func (sam *StreamAlertManager) TriggerAlert(alert *StreamAlert) {
	sam.alertsMutex.Lock()
	sam.alerts[alert.ID] = alert
	sam.alertsMutex.Unlock()
	
	// Notify subscribers
	for _, subscriber := range sam.subscribers {
		go func(sub AlertSubscriber) {
			if err := sub.OnAlert(alert); err != nil {
				log.Printf("Failed to notify alert subscriber: %v", err)
			}
		}(subscriber)
	}
	
	log.Printf("ALERT [%s]: %s - %s", alert.Severity, alert.Title, alert.Description)
}

// GetStreamMetrics returns current metrics for all streams
func (tsm *TensorStreamMonitor) GetStreamMetrics() map[string]*StreamMetrics {
	tsm.metricsMutex.RLock()
	defer tsm.metricsMutex.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*StreamMetrics)
	for k, v := range tsm.metrics {
		result[k] = v
	}
	
	return result
}

// GenerateReport generates a comprehensive performance report
func (tsm *TensorStreamMonitor) GenerateReport() ([]byte, error) {
	report := map[string]interface{}{
		"timestamp":       time.Now(),
		"metrics":         tsm.GetStreamMetrics(),
		"active_alerts":   tsm.alertManager.GetActiveAlerts(),
		"performance_summary": tsm.generatePerformanceSummary(),
	}
	
	return json.MarshalIndent(report, "", "  ")
}

// GetActiveAlerts returns all active alerts
func (sam *StreamAlertManager) GetActiveAlerts() map[string]*StreamAlert {
	sam.alertsMutex.RLock()
	defer sam.alertsMutex.RUnlock()
	
	active := make(map[string]*StreamAlert)
	for id, alert := range sam.alerts {
		if !alert.Resolved {
			active[id] = alert
		}
	}
	
	return active
}

// generatePerformanceSummary generates a summary of overall performance
func (tsm *TensorStreamMonitor) generatePerformanceSummary() map[string]interface{} {
	tsm.metricsMutex.RLock()
	defer tsm.metricsMutex.RUnlock()
	
	if len(tsm.metrics) == 0 {
		return map[string]interface{}{
			"total_streams": 0,
		}
	}
	
	var totalThroughput, totalLatency, totalCompressionRatio float64
	var totalBytes int64
	var totalErrors int
	
	for _, metrics := range tsm.metrics {
		totalThroughput += metrics.ThroughputMBps
		totalLatency += metrics.LatencyMs
		totalCompressionRatio += metrics.CompressionRatio
		totalBytes += metrics.TotalBytes
		totalErrors += metrics.ErrorCount
	}
	
	count := float64(len(tsm.metrics))
	
	return map[string]interface{}{
		"total_streams":         len(tsm.metrics),
		"average_throughput":    totalThroughput / count,
		"average_latency":       totalLatency / count,
		"average_compression":   totalCompressionRatio / count,
		"total_bytes":           totalBytes,
		"total_errors":          totalErrors,
		"last_report_time":      tsm.lastReport,
	}
}

// Close shuts down the tensor stream monitor
func (tsm *TensorStreamMonitor) Close() error {
	tsm.cancel()
	return nil
}
