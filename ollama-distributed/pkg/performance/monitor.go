package performance

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PerformanceMonitor monitors system performance in real-time
type PerformanceMonitor struct {
	config *OptimizerConfig

	// Monitoring state
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Metrics collection
	metrics       *PerformanceMetrics
	metricsSamples []PerformanceSnapshot
	maxSamples    int
	mu            sync.RWMutex

	// Alert thresholds
	alertThresholds *AlertThresholds
	alertHandlers   []AlertHandler
}

// PerformanceSnapshot represents a point-in-time performance snapshot
type PerformanceSnapshot struct {
	Timestamp       time.Time     `json:"timestamp"`
	CPUPercent      float64       `json:"cpu_percent"`
	MemoryMB        float64       `json:"memory_mb"`
	GoroutineCount  int           `json:"goroutine_count"`
	GCPauseMS       float64       `json:"gc_pause_ms"`
	RequestLatency  time.Duration `json:"request_latency"`
	ThroughputRPS   float64       `json:"throughput_rps"`
	ActiveConns     int           `json:"active_connections"`
	ErrorRate       float64       `json:"error_rate"`
}

// AlertThresholds defines performance alert thresholds
type AlertThresholds struct {
	MaxCPUPercent      float64       `json:"max_cpu_percent"`
	MaxMemoryMB        float64       `json:"max_memory_mb"`
	MaxGoroutines      int           `json:"max_goroutines"`
	MaxGCPauseMS       float64       `json:"max_gc_pause_ms"`
	MaxLatencyMS       float64       `json:"max_latency_ms"`
	MinThroughputRPS   float64       `json:"min_throughput_rps"`
	MaxErrorRate       float64       `json:"max_error_rate"`
}

// AlertHandler handles performance alerts
type AlertHandler func(alert *PerformanceAlert)

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Metric      string                 `json:"metric"`
	Value       interface{}            `json:"value"`
	Threshold   interface{}            `json:"threshold"`
	Timestamp   time.Time              `json:"timestamp"`
	Snapshot    *PerformanceSnapshot   `json:"snapshot"`
	Suggestions []string               `json:"suggestions"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *OptimizerConfig) *PerformanceMonitor {
	thresholds := &AlertThresholds{
		MaxCPUPercent:    config.MaxCPUUsagePercent,
		MaxMemoryMB:      float64(config.MaxMemoryUsageMB),
		MaxGoroutines:    1000,
		MaxGCPauseMS:     float64(config.GCMaxPause.Milliseconds()),
		MaxLatencyMS:     float64(config.TargetLatencyP99MS),
		MinThroughputRPS: float64(config.TargetThroughputOPS) * 0.8, // 80% of target
		MaxErrorRate:     5.0, // 5% error rate
	}

	return &PerformanceMonitor{
		config:          config,
		metrics:         &PerformanceMetrics{LastUpdated: time.Now()},
		metricsSamples:  make([]PerformanceSnapshot, 0),
		maxSamples:      1000, // Keep last 1000 samples
		alertThresholds: thresholds,
		alertHandlers:   make([]AlertHandler, 0),
	}
}

// Start starts the performance monitor
func (pm *PerformanceMonitor) Start() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		return nil
	}

	pm.ctx, pm.cancel = context.WithCancel(context.Background())
	pm.isRunning = true

	// Start monitoring goroutine
	pm.wg.Add(1)
	go pm.runMonitoring()

	// Start alert checking goroutine
	pm.wg.Add(1)
	go pm.runAlertChecking()

	log.Info().
		Dur("interval", pm.config.MetricsInterval).
		Msg("Performance monitor started")

	return nil
}

// Stop stops the performance monitor
func (pm *PerformanceMonitor) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return nil
	}

	pm.cancel()
	pm.wg.Wait()
	pm.isRunning = false

	log.Info().Msg("Performance monitor stopped")
	return nil
}

// GetCurrentMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetCurrentMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy
	metrics := *pm.metrics
	return &metrics
}

// GetRecentSamples returns recent performance samples
func (pm *PerformanceMonitor) GetRecentSamples(duration time.Duration) []PerformanceSnapshot {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	samples := make([]PerformanceSnapshot, 0)

	for _, sample := range pm.metricsSamples {
		if sample.Timestamp.After(cutoff) {
			samples = append(samples, sample)
		}
	}

	return samples
}

// AddAlertHandler adds an alert handler
func (pm *PerformanceMonitor) AddAlertHandler(handler AlertHandler) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.alertHandlers = append(pm.alertHandlers, handler)
}

// SetAlertThresholds updates alert thresholds
func (pm *PerformanceMonitor) SetAlertThresholds(thresholds *AlertThresholds) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.alertThresholds = thresholds
}

// runMonitoring runs the main monitoring loop
func (pm *PerformanceMonitor) runMonitoring() {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.collectMetrics()
		}
	}
}

// collectMetrics collects current performance metrics
func (pm *PerformanceMonitor) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := PerformanceSnapshot{
		Timestamp:      time.Now(),
		MemoryMB:       float64(m.Alloc) / 1024 / 1024,
		GoroutineCount: runtime.NumGoroutine(),
		GCPauseMS:      float64(m.PauseTotalNs) / 1e6,
	}

	// Add CPU monitoring (simplified)
	snapshot.CPUPercent = pm.estimateCPUUsage()

	// Update metrics
	pm.mu.Lock()
	pm.metrics.MemoryUsageMB = snapshot.MemoryMB
	pm.metrics.GoroutineCount = snapshot.GoroutineCount
	pm.metrics.GCPauseMS = snapshot.GCPauseMS
	pm.metrics.CPUUsagePercent = snapshot.CPUPercent
	pm.metrics.LastUpdated = snapshot.Timestamp

	// Store snapshot
	pm.metricsSamples = append(pm.metricsSamples, snapshot)
	if len(pm.metricsSamples) > pm.maxSamples {
		pm.metricsSamples = pm.metricsSamples[1:]
	}
	pm.mu.Unlock()

	if pm.config.PerformanceLogging {
		log.Debug().
			Float64("memory_mb", snapshot.MemoryMB).
			Float64("cpu_percent", snapshot.CPUPercent).
			Int("goroutines", snapshot.GoroutineCount).
			Float64("gc_pause_ms", snapshot.GCPauseMS).
			Msg("Performance metrics collected")
	}
}

// runAlertChecking runs the alert checking loop
func (pm *PerformanceMonitor) runAlertChecking() {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.config.MetricsInterval * 2) // Check alerts less frequently
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.checkAlerts()
		}
	}
}

// checkAlerts checks for performance alerts
func (pm *PerformanceMonitor) checkAlerts() {
	pm.mu.RLock()
	metrics := *pm.metrics
	thresholds := *pm.alertThresholds
	handlers := make([]AlertHandler, len(pm.alertHandlers))
	copy(handlers, pm.alertHandlers)

	var currentSnapshot *PerformanceSnapshot
	if len(pm.metricsSamples) > 0 {
		currentSnapshot = &pm.metricsSamples[len(pm.metricsSamples)-1]
	}
	pm.mu.RUnlock()

	alerts := make([]*PerformanceAlert, 0)

	// Check CPU usage
	if metrics.CPUUsagePercent > thresholds.MaxCPUPercent {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "cpu_usage",
			Severity:  pm.getSeverity(metrics.CPUUsagePercent, thresholds.MaxCPUPercent, "high"),
			Message:   "High CPU usage detected",
			Metric:    "cpu_percent",
			Value:     metrics.CPUUsagePercent,
			Threshold: thresholds.MaxCPUPercent,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Check for CPU-intensive operations",
				"Consider horizontal scaling",
				"Optimize algorithms for better efficiency",
			},
		})
	}

	// Check memory usage
	if metrics.MemoryUsageMB > thresholds.MaxMemoryMB {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "memory_usage",
			Severity:  pm.getSeverity(metrics.MemoryUsageMB, thresholds.MaxMemoryMB, "high"),
			Message:   "High memory usage detected",
			Metric:    "memory_mb",
			Value:     metrics.MemoryUsageMB,
			Threshold: thresholds.MaxMemoryMB,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Check for memory leaks",
				"Optimize data structures",
				"Increase garbage collection frequency",
				"Consider memory pooling",
			},
		})
	}

	// Check goroutine count
	if metrics.GoroutineCount > thresholds.MaxGoroutines {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "goroutine_count",
			Severity:  pm.getSeverity(float64(metrics.GoroutineCount), float64(thresholds.MaxGoroutines), "high"),
			Message:   "High goroutine count detected",
			Metric:    "goroutine_count",
			Value:     metrics.GoroutineCount,
			Threshold: thresholds.MaxGoroutines,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Check for goroutine leaks",
				"Implement goroutine pooling",
				"Review concurrent operations",
			},
		})
	}

	// Check GC pause times
	if metrics.GCPauseMS > thresholds.MaxGCPauseMS {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "gc_pause",
			Severity:  pm.getSeverity(metrics.GCPauseMS, thresholds.MaxGCPauseMS, "high"),
			Message:   "High GC pause time detected",
			Metric:    "gc_pause_ms",
			Value:     metrics.GCPauseMS,
			Threshold: thresholds.MaxGCPauseMS,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Reduce GC target percentage",
				"Optimize memory allocations",
				"Consider reducing heap size",
			},
		})
	}

	// Check latency
	latencyMS := float64(metrics.P99Latency.Milliseconds())
	if latencyMS > thresholds.MaxLatencyMS {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "latency",
			Severity:  pm.getSeverity(latencyMS, thresholds.MaxLatencyMS, "high"),
			Message:   "High response latency detected",
			Metric:    "latency_ms",
			Value:     latencyMS,
			Threshold: thresholds.MaxLatencyMS,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Check for blocking operations",
				"Optimize database queries",
				"Implement caching",
				"Consider connection pooling",
			},
		})
	}

	// Check throughput
	if metrics.OperationsPerSecond > 0 && metrics.OperationsPerSecond < thresholds.MinThroughputRPS {
		alerts = append(alerts, &PerformanceAlert{
			Type:      "throughput",
			Severity:  pm.getSeverity(thresholds.MinThroughputRPS, metrics.OperationsPerSecond, "low"),
			Message:   "Low throughput detected",
			Metric:    "throughput_rps",
			Value:     metrics.OperationsPerSecond,
			Threshold: thresholds.MinThroughputRPS,
			Timestamp: time.Now(),
			Snapshot:  currentSnapshot,
			Suggestions: []string{
				"Check for bottlenecks",
				"Optimize critical path",
				"Consider horizontal scaling",
				"Review resource allocation",
			},
		})
	}

	// Trigger alert handlers
	for _, alert := range alerts {
		for _, handler := range handlers {
			go handler(alert)
		}
	}
}

// getSeverity determines alert severity
func (pm *PerformanceMonitor) getSeverity(value, threshold float64, direction string) string {
	var ratio float64
	if direction == "high" {
		ratio = value / threshold
	} else { // low
		ratio = threshold / value
	}

	if ratio >= 2.0 {
		return "critical"
	} else if ratio >= 1.5 {
		return "high"
	} else if ratio >= 1.2 {
		return "medium"
	}
	return "low"
}

// estimateCPUUsage estimates CPU usage (simplified implementation)
func (pm *PerformanceMonitor) estimateCPUUsage() float64 {
	// This is a simplified CPU estimation
	// In a real implementation, you would use system calls or third-party libraries
	
	// Use number of goroutines as a rough proxy for CPU activity
	goroutines := runtime.NumGoroutine()
	
	// Estimate based on goroutine count (very rough approximation)
	cpuEstimate := float64(goroutines) / 10.0
	if cpuEstimate > 100.0 {
		cpuEstimate = 100.0
	}
	
	return cpuEstimate
}

// GetPerformanceSummary returns a summary of recent performance
func (pm *PerformanceMonitor) GetPerformanceSummary(duration time.Duration) *PerformanceSummary {
	samples := pm.GetRecentSamples(duration)
	if len(samples) == 0 {
		return nil
	}

	summary := &PerformanceSummary{
		Duration:    duration,
		SampleCount: len(samples),
		StartTime:   samples[0].Timestamp,
		EndTime:     samples[len(samples)-1].Timestamp,
	}

	// Calculate statistics
	var totalCPU, totalMemory, totalGC, totalLatency float64
	var maxCPU, maxMemory, maxGC, maxLatency float64
	var minCPU, minMemory, minGC, minLatency float64 = 100, 1000000, 1000000, 1000000

	for _, sample := range samples {
		// CPU
		totalCPU += sample.CPUPercent
		if sample.CPUPercent > maxCPU {
			maxCPU = sample.CPUPercent
		}
		if sample.CPUPercent < minCPU {
			minCPU = sample.CPUPercent
		}

		// Memory
		totalMemory += sample.MemoryMB
		if sample.MemoryMB > maxMemory {
			maxMemory = sample.MemoryMB
		}
		if sample.MemoryMB < minMemory {
			minMemory = sample.MemoryMB
		}

		// GC
		totalGC += sample.GCPauseMS
		if sample.GCPauseMS > maxGC {
			maxGC = sample.GCPauseMS
		}
		if sample.GCPauseMS < minGC {
			minGC = sample.GCPauseMS
		}

		// Latency
		latencyMS := float64(sample.RequestLatency.Milliseconds())
		totalLatency += latencyMS
		if latencyMS > maxLatency {
			maxLatency = latencyMS
		}
		if latencyMS < minLatency {
			minLatency = latencyMS
		}
	}

	sampleCount := float64(len(samples))
	summary.AvgCPUPercent = totalCPU / sampleCount
	summary.MaxCPUPercent = maxCPU
	summary.AvgMemoryMB = totalMemory / sampleCount
	summary.MaxMemoryMB = maxMemory
	summary.AvgGCPauseMS = totalGC / sampleCount
	summary.MaxGCPauseMS = maxGC
	summary.AvgLatencyMS = totalLatency / sampleCount
	summary.MaxLatencyMS = maxLatency

	return summary
}

// PerformanceSummary contains performance statistics over a time period
type PerformanceSummary struct {
	Duration      time.Duration `json:"duration"`
	SampleCount   int           `json:"sample_count"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	
	AvgCPUPercent float64       `json:"avg_cpu_percent"`
	MaxCPUPercent float64       `json:"max_cpu_percent"`
	AvgMemoryMB   float64       `json:"avg_memory_mb"`
	MaxMemoryMB   float64       `json:"max_memory_mb"`
	AvgGCPauseMS  float64       `json:"avg_gc_pause_ms"`
	MaxGCPauseMS  float64       `json:"max_gc_pause_ms"`
	AvgLatencyMS  float64       `json:"avg_latency_ms"`
	MaxLatencyMS  float64       `json:"max_latency_ms"`
}