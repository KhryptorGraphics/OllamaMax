package pkg

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor tracks system and application performance metrics
type PerformanceMonitor struct {
	mutex          sync.RWMutex
	metrics        *PerformanceMetrics
	collectors     []MetricCollector
	interval       time.Duration
	stopCh         chan struct{}
	wg             sync.WaitGroup
	alertThresholds *AlertThresholds
}

// PerformanceMetrics contains comprehensive performance data
type PerformanceMetrics struct {
	Timestamp time.Time `json:"timestamp"`
	
	// System metrics
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsageBytes   int64   `json:"memory_usage_bytes"`
	MemoryTotalBytes   int64   `json:"memory_total_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	GoroutineCount     int     `json:"goroutine_count"`
	
	// Application metrics
	RequestsPerSecond     float64       `json:"requests_per_second"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	ErrorRate             float64       `json:"error_rate"`
	ActiveConnections     int           `json:"active_connections"`
	
	// Model serving metrics
	ModelsLoaded          int     `json:"models_loaded"`
	InferenceLatency      time.Duration `json:"inference_latency"`
	InferencesThroughput  float64 `json:"inferences_throughput"`
	ModelMemoryUsage      int64   `json:"model_memory_usage"`
	
	// Distributed system metrics
	ConnectedPeers        int     `json:"connected_peers"`
	NetworkLatency        time.Duration `json:"network_latency"`
	ConsensusLatency      time.Duration `json:"consensus_latency"`
	ReplicationLag        time.Duration `json:"replication_lag"`
	PartitionCount        int     `json:"partition_count"`
}

// AlertThresholds defines thresholds for performance alerts
type AlertThresholds struct {
	CPUUsagePercent       float64       `yaml:"cpu_usage_percent" json:"cpu_usage_percent"`
	MemoryUsagePercent    float64       `yaml:"memory_usage_percent" json:"memory_usage_percent"`
	AverageResponseTime   time.Duration `yaml:"average_response_time" json:"average_response_time"`
	ErrorRate             float64       `yaml:"error_rate" json:"error_rate"`
	InferenceLatency      time.Duration `yaml:"inference_latency" json:"inference_latency"`
	NetworkLatency        time.Duration `yaml:"network_latency" json:"network_latency"`
}

// MetricCollector interface for pluggable metric collection
type MetricCollector interface {
	CollectMetrics(ctx context.Context) (*PerformanceMetrics, error)
	Name() string
}

// Alert represents a performance alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Message     string                 `json:"message"`
	Threshold   interface{}            `json:"threshold"`
	ActualValue interface{}            `json:"actual_value"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertSeverity represents the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(interval time.Duration, thresholds *AlertThresholds) *PerformanceMonitor {
	if thresholds == nil {
		thresholds = &AlertThresholds{
			CPUUsagePercent:     80.0,
			MemoryUsagePercent:  85.0,
			AverageResponseTime: 5 * time.Second,
			ErrorRate:           0.05, // 5%
			InferenceLatency:    10 * time.Second,
			NetworkLatency:      1 * time.Second,
		}
	}
	
	return &PerformanceMonitor{
		metrics:         &PerformanceMetrics{},
		collectors:      make([]MetricCollector, 0),
		interval:        interval,
		stopCh:          make(chan struct{}),
		alertThresholds: thresholds,
	}
}

// AddCollector adds a metric collector
func (pm *PerformanceMonitor) AddCollector(collector MetricCollector) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.collectors = append(pm.collectors, collector)
}

// Start begins metric collection
func (pm *PerformanceMonitor) Start(ctx context.Context) error {
	pm.wg.Add(1)
	go pm.collectLoop(ctx)
	return nil
}

// Stop stops metric collection
func (pm *PerformanceMonitor) Stop() error {
	close(pm.stopCh)
	pm.wg.Wait()
	return nil
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metricsCopy := *pm.metrics
	return &metricsCopy
}

// collectLoop runs the metric collection loop
func (pm *PerformanceMonitor) collectLoop(ctx context.Context) {
	defer pm.wg.Done()
	
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			pm.collectMetrics(ctx)
		case <-pm.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// collectMetrics collects metrics from all sources
func (pm *PerformanceMonitor) collectMetrics(ctx context.Context) {
	// Collect system metrics
	systemMetrics := pm.collectSystemMetrics()
	
	// Collect metrics from registered collectors
	for _, collector := range pm.collectors {
		collectorMetrics, err := collector.CollectMetrics(ctx)
		if err != nil {
			fmt.Printf("Error collecting metrics from %s: %v\n", collector.Name(), err)
			continue
		}
		
		// Merge collector metrics with system metrics
		pm.mergeMetrics(systemMetrics, collectorMetrics)
	}
	
	// Update stored metrics
	pm.mutex.Lock()
	pm.metrics = systemMetrics
	pm.mutex.Unlock()
	
	// Check for alerts
	pm.checkAlerts(systemMetrics)
}

// collectSystemMetrics collects basic system metrics
func (pm *PerformanceMonitor) collectSystemMetrics() *PerformanceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return &PerformanceMetrics{
		Timestamp:         time.Now(),
		MemoryUsageBytes:  int64(memStats.Alloc),
		MemoryTotalBytes:  int64(memStats.Sys),
		GoroutineCount:    runtime.NumGoroutine(),
		// CPU usage would require additional system calls or libraries
		CPUUsagePercent:   0.0, // Placeholder - would need system-specific implementation
		MemoryUsagePercent: float64(memStats.Alloc) / float64(memStats.Sys) * 100,
	}
}

// mergeMetrics merges metrics from collectors into the main metrics
func (pm *PerformanceMonitor) mergeMetrics(target, source *PerformanceMetrics) {
	if source.RequestsPerSecond > 0 {
		target.RequestsPerSecond = source.RequestsPerSecond
	}
	if source.AverageResponseTime > 0 {
		target.AverageResponseTime = source.AverageResponseTime
	}
	if source.ErrorRate > 0 {
		target.ErrorRate = source.ErrorRate
	}
	if source.ActiveConnections > 0 {
		target.ActiveConnections = source.ActiveConnections
	}
	if source.ModelsLoaded > 0 {
		target.ModelsLoaded = source.ModelsLoaded
	}
	if source.InferenceLatency > 0 {
		target.InferenceLatency = source.InferenceLatency
	}
	if source.InferencesThroughput > 0 {
		target.InferencesThroughput = source.InferencesThroughput
	}
	if source.ModelMemoryUsage > 0 {
		target.ModelMemoryUsage = source.ModelMemoryUsage
	}
	if source.ConnectedPeers > 0 {
		target.ConnectedPeers = source.ConnectedPeers
	}
	if source.NetworkLatency > 0 {
		target.NetworkLatency = source.NetworkLatency
	}
	if source.ConsensusLatency > 0 {
		target.ConsensusLatency = source.ConsensusLatency
	}
	if source.ReplicationLag > 0 {
		target.ReplicationLag = source.ReplicationLag
	}
	if source.PartitionCount > 0 {
		target.PartitionCount = source.PartitionCount
	}
}

// checkAlerts checks metrics against thresholds and generates alerts
func (pm *PerformanceMonitor) checkAlerts(metrics *PerformanceMetrics) {
	alerts := make([]Alert, 0)
	
	// CPU usage alert
	if metrics.CPUUsagePercent > pm.alertThresholds.CPUUsagePercent {
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("cpu_high_%d", time.Now().Unix()),
			Type:        "cpu_usage",
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("High CPU usage: %.2f%%", metrics.CPUUsagePercent),
			Threshold:   pm.alertThresholds.CPUUsagePercent,
			ActualValue: metrics.CPUUsagePercent,
			Timestamp:   time.Now(),
		})
	}
	
	// Memory usage alert
	if metrics.MemoryUsagePercent > pm.alertThresholds.MemoryUsagePercent {
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("memory_high_%d", time.Now().Unix()),
			Type:        "memory_usage",
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("High memory usage: %.2f%%", metrics.MemoryUsagePercent),
			Threshold:   pm.alertThresholds.MemoryUsagePercent,
			ActualValue: metrics.MemoryUsagePercent,
			Timestamp:   time.Now(),
		})
	}
	
	// Response time alert
	if metrics.AverageResponseTime > pm.alertThresholds.AverageResponseTime {
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("response_time_high_%d", time.Now().Unix()),
			Type:        "response_time",
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("High response time: %v", metrics.AverageResponseTime),
			Threshold:   pm.alertThresholds.AverageResponseTime,
			ActualValue: metrics.AverageResponseTime,
			Timestamp:   time.Now(),
		})
	}
	
	// Error rate alert
	if metrics.ErrorRate > pm.alertThresholds.ErrorRate {
		severity := AlertSeverityWarning
		if metrics.ErrorRate > pm.alertThresholds.ErrorRate*2 {
			severity = AlertSeverityCritical
		}
		
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("error_rate_high_%d", time.Now().Unix()),
			Type:        "error_rate",
			Severity:    severity,
			Message:     fmt.Sprintf("High error rate: %.2f%%", metrics.ErrorRate*100),
			Threshold:   pm.alertThresholds.ErrorRate,
			ActualValue: metrics.ErrorRate,
			Timestamp:   time.Now(),
		})
	}
	
	// Inference latency alert
	if metrics.InferenceLatency > pm.alertThresholds.InferenceLatency {
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("inference_latency_high_%d", time.Now().Unix()),
			Type:        "inference_latency",
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("High inference latency: %v", metrics.InferenceLatency),
			Threshold:   pm.alertThresholds.InferenceLatency,
			ActualValue: metrics.InferenceLatency,
			Timestamp:   time.Now(),
		})
	}
	
	// Network latency alert
	if metrics.NetworkLatency > pm.alertThresholds.NetworkLatency {
		alerts = append(alerts, Alert{
			ID:          fmt.Sprintf("network_latency_high_%d", time.Now().Unix()),
			Type:        "network_latency",
			Severity:    AlertSeverityWarning,
			Message:     fmt.Sprintf("High network latency: %v", metrics.NetworkLatency),
			Threshold:   pm.alertThresholds.NetworkLatency,
			ActualValue: metrics.NetworkLatency,
			Timestamp:   time.Now(),
		})
	}
	
	// Process alerts (in a real implementation, these would be sent to an alerting system)
	for _, alert := range alerts {
		fmt.Printf("ALERT [%s] %s: %s\n", alert.Severity, alert.Type, alert.Message)
	}
}

// SystemHealthStatus represents overall system health
type SystemHealthStatus struct {
	Status      string                 `json:"status"`      // "healthy", "degraded", "unhealthy"
	Score       float64               `json:"score"`       // 0-100, higher is better
	Issues      []string              `json:"issues"`
	Metrics     *PerformanceMetrics   `json:"metrics"`
	Timestamp   time.Time             `json:"timestamp"`
	Details     map[string]interface{} `json:"details"`
}

// GetHealthStatus returns overall system health
func (pm *PerformanceMonitor) GetHealthStatus() *SystemHealthStatus {
	metrics := pm.GetMetrics()
	
	// Calculate health score based on various factors
	score := 100.0
	issues := make([]string, 0)
	
	// CPU penalty
	if metrics.CPUUsagePercent > 90 {
		score -= 30
		issues = append(issues, "High CPU usage")
	} else if metrics.CPUUsagePercent > 70 {
		score -= 15
		issues = append(issues, "Elevated CPU usage")
	}
	
	// Memory penalty
	if metrics.MemoryUsagePercent > 90 {
		score -= 30
		issues = append(issues, "High memory usage")
	} else if metrics.MemoryUsagePercent > 75 {
		score -= 15
		issues = append(issues, "Elevated memory usage")
	}
	
	// Error rate penalty
	if metrics.ErrorRate > 0.1 {
		score -= 40
		issues = append(issues, "High error rate")
	} else if metrics.ErrorRate > 0.05 {
		score -= 20
		issues = append(issues, "Elevated error rate")
	}
	
	// Response time penalty
	if metrics.AverageResponseTime > 10*time.Second {
		score -= 25
		issues = append(issues, "High response time")
	} else if metrics.AverageResponseTime > 5*time.Second {
		score -= 10
		issues = append(issues, "Elevated response time")
	}
	
	// Determine status
	status := "healthy"
	if score < 50 {
		status = "unhealthy"
	} else if score < 80 {
		status = "degraded"
	}
	
	return &SystemHealthStatus{
		Status:    status,
		Score:     score,
		Issues:    issues,
		Metrics:   metrics,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"collectors_count": len(pm.collectors),
			"monitoring_interval": pm.interval.String(),
		},
	}
}