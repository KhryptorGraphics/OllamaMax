package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PerformanceProfiler provides real-time performance monitoring and bottleneck detection
type PerformanceProfiler struct {
	config             *ProfilerConfig
	bottleneckDetector *BottleneckDetector
	resourceMonitor    *ResourceMonitor
	performanceTracker *PerformanceTracker

	// State management
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// ProfilerConfig configures the performance profiler
type ProfilerConfig struct {
	Enabled             bool          `json:"enabled"`
	RealTimeMonitoring  bool          `json:"real_time_monitoring"`
	BottleneckDetection bool          `json:"bottleneck_detection"`
	ProfilingInterval   time.Duration `json:"profiling_interval"`

	// Monitoring thresholds
	CPUThreshold     float64       `json:"cpu_threshold"`
	MemoryThreshold  float64       `json:"memory_threshold"`
	LatencyThreshold time.Duration `json:"latency_threshold"`

	// Detection settings
	DetectionWindow time.Duration `json:"detection_window"`
	MinSampleSize   int           `json:"min_sample_size"`
	ConfidenceLevel float64       `json:"confidence_level"`

	// Profiling settings
	EnableCPUProfiling       bool `json:"enable_cpu_profiling"`
	EnableMemoryProfiling    bool `json:"enable_memory_profiling"`
	EnableGoroutineProfiling bool `json:"enable_goroutine_profiling"`
}

// BottleneckDetector detects performance bottlenecks
type BottleneckDetector struct {
	config      *ProfilerConfig
	bottlenecks map[string]*Bottleneck
	samples     []PerformanceSample
	mu          sync.RWMutex
}

// ResourceMonitor monitors system resources
type ResourceMonitor struct {
	config  *ProfilerConfig
	metrics *ResourceMetrics
	mu      sync.RWMutex
}

// PerformanceTracker tracks performance trends
type PerformanceTracker struct {
	config *ProfilerConfig
	trends map[string]*PerformanceTrend
	mu     sync.RWMutex
}

// Bottleneck represents a detected performance bottleneck
type Bottleneck struct {
	ID          string                 `json:"id"`
	Type        BottleneckType         `json:"type"`
	Severity    BottleneckSeverity     `json:"severity"`
	Component   string                 `json:"component"`
	Description string                 `json:"description"`
	Impact      float64                `json:"impact"`
	Confidence  float64                `json:"confidence"`
	DetectedAt  time.Time              `json:"detected_at"`
	Evidence    []PerformanceSample    `json:"evidence"`
	Suggestions []string               `json:"suggestions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// BottleneckType represents the type of bottleneck
type BottleneckType string

const (
	BottleneckTypeCPU        BottleneckType = "cpu"
	BottleneckTypeMemory     BottleneckType = "memory"
	BottleneckTypeNetwork    BottleneckType = "network"
	BottleneckTypeDisk       BottleneckType = "disk"
	BottleneckTypeGoroutine  BottleneckType = "goroutine"
	BottleneckTypeGC         BottleneckType = "gc"
	BottleneckTypeLatency    BottleneckType = "latency"
	BottleneckTypeThroughput BottleneckType = "throughput"
)

// BottleneckSeverity represents the severity of a bottleneck
type BottleneckSeverity string

const (
	BottleneckSeverityLow      BottleneckSeverity = "low"
	BottleneckSeverityMedium   BottleneckSeverity = "medium"
	BottleneckSeverityHigh     BottleneckSeverity = "high"
	BottleneckSeverityCritical BottleneckSeverity = "critical"
)

// PerformanceSample represents a performance measurement sample
type PerformanceSample struct {
	Timestamp time.Time              `json:"timestamp"`
	Component string                 `json:"component"`
	Metric    string                 `json:"metric"`
	Value     float64                `json:"value"`
	Unit      string                 `json:"unit"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ResourceMetrics represents current resource usage metrics
type ResourceMetrics struct {
	CPUUsage       float64        `json:"cpu_usage"`
	MemoryUsage    float64        `json:"memory_usage"`
	GoroutineCount int            `json:"goroutine_count"`
	GCStats        GCMetrics      `json:"gc_stats"`
	NetworkStats   NetworkMetrics `json:"network_stats"`
	DiskStats      DiskMetrics    `json:"disk_stats"`
	Timestamp      time.Time      `json:"timestamp"`
}

// GCMetrics represents garbage collection metrics
type GCMetrics struct {
	NumGC       uint32        `json:"num_gc"`
	PauseTotal  time.Duration `json:"pause_total"`
	LastPause   time.Duration `json:"last_pause"`
	HeapObjects uint64        `json:"heap_objects"`
	HeapAlloc   uint64        `json:"heap_alloc"`
	HeapSys     uint64        `json:"heap_sys"`
}

// NetworkMetrics represents network performance metrics
type NetworkMetrics struct {
	BytesSent       uint64 `json:"bytes_sent"`
	BytesReceived   uint64 `json:"bytes_received"`
	PacketsSent     uint64 `json:"packets_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	Connections     int    `json:"connections"`
}

// DiskMetrics represents disk performance metrics
type DiskMetrics struct {
	ReadBytes   uint64  `json:"read_bytes"`
	WriteBytes  uint64  `json:"write_bytes"`
	ReadOps     uint64  `json:"read_ops"`
	WriteOps    uint64  `json:"write_ops"`
	Utilization float64 `json:"utilization"`
}

// PerformanceTrend represents a performance trend
type PerformanceTrend struct {
	Metric     string    `json:"metric"`
	Component  string    `json:"component"`
	Trend      string    `json:"trend"` // improving, degrading, stable
	Slope      float64   `json:"slope"`
	Confidence float64   `json:"confidence"`
	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`
	Samples    []float64 `json:"samples"`
}

// ProfilingReport represents a comprehensive profiling report
type ProfilingReport struct {
	GeneratedAt     time.Time                    `json:"generated_at"`
	Duration        time.Duration                `json:"duration"`
	ResourceMetrics *ResourceMetrics             `json:"resource_metrics"`
	Bottlenecks     []Bottleneck                 `json:"bottlenecks"`
	Trends          map[string]*PerformanceTrend `json:"trends"`
	Recommendations []string                     `json:"recommendations"`
	Summary         string                       `json:"summary"`
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler(config *ProfilerConfig) *PerformanceProfiler {
	if config == nil {
		config = DefaultProfilerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pp := &PerformanceProfiler{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize components
	if config.BottleneckDetection {
		pp.bottleneckDetector = &BottleneckDetector{
			config:      config,
			bottlenecks: make(map[string]*Bottleneck),
			samples:     make([]PerformanceSample, 0),
		}
	}

	pp.resourceMonitor = &ResourceMonitor{
		config:  config,
		metrics: &ResourceMetrics{},
	}

	pp.performanceTracker = &PerformanceTracker{
		config: config,
		trends: make(map[string]*PerformanceTrend),
	}

	return pp
}

// Start starts the performance profiler
func (pp *PerformanceProfiler) Start() error {
	if !pp.config.Enabled {
		log.Info().Msg("Performance profiler disabled")
		return nil
	}

	// Start monitoring loops
	if pp.config.RealTimeMonitoring {
		go pp.monitoringLoop()
	}

	if pp.bottleneckDetector != nil {
		go pp.bottleneckDetectionLoop()
	}

	go pp.trendAnalysisLoop()

	log.Info().
		Bool("real_time_monitoring", pp.config.RealTimeMonitoring).
		Bool("bottleneck_detection", pp.config.BottleneckDetection).
		Dur("profiling_interval", pp.config.ProfilingInterval).
		Msg("Performance profiler started")

	return nil
}

// monitoringLoop performs real-time resource monitoring
func (pp *PerformanceProfiler) monitoringLoop() {
	ticker := time.NewTicker(pp.config.ProfilingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pp.ctx.Done():
			return
		case <-ticker.C:
			pp.collectResourceMetrics()
		}
	}
}

// collectResourceMetrics collects current resource metrics
func (pp *PerformanceProfiler) collectResourceMetrics() {
	pp.resourceMonitor.mu.Lock()
	defer pp.resourceMonitor.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pp.resourceMonitor.metrics = &ResourceMetrics{
		CPUUsage:       pp.getCPUUsage(),
		MemoryUsage:    float64(m.Alloc) / float64(m.Sys),
		GoroutineCount: runtime.NumGoroutine(),
		GCStats: GCMetrics{
			NumGC:       m.NumGC,
			PauseTotal:  time.Duration(m.PauseTotalNs),
			LastPause:   time.Duration(m.PauseNs[(m.NumGC+255)%256]),
			HeapObjects: m.HeapObjects,
			HeapAlloc:   m.HeapAlloc,
			HeapSys:     m.HeapSys,
		},
		NetworkStats: pp.getNetworkStats(),
		DiskStats:    pp.getDiskStats(),
		Timestamp:    time.Now(),
	}

	// Record samples for trend analysis
	pp.recordSample("cpu_usage", "system", pp.resourceMonitor.metrics.CPUUsage, "%")
	pp.recordSample("memory_usage", "system", pp.resourceMonitor.metrics.MemoryUsage, "%")
	pp.recordSample("goroutine_count", "runtime", float64(pp.resourceMonitor.metrics.GoroutineCount), "count")
}

// getCPUUsage gets current CPU usage (simplified implementation)
func (pp *PerformanceProfiler) getCPUUsage() float64 {
	// This would typically use system calls or external libraries
	// For now, return a placeholder value based on goroutine count
	goroutines := runtime.NumGoroutine()
	cpuCores := runtime.NumCPU()
	usage := float64(goroutines) / float64(cpuCores*100)
	if usage > 1.0 {
		usage = 1.0
	}
	return usage
}

// getNetworkStats gets current network statistics (placeholder)
func (pp *PerformanceProfiler) getNetworkStats() NetworkMetrics {
	// This would typically read from /proc/net/dev or use system APIs
	return NetworkMetrics{
		BytesSent:       1000000,
		BytesReceived:   2000000,
		PacketsSent:     1000,
		PacketsReceived: 2000,
		Connections:     10,
	}
}

// getDiskStats gets current disk statistics (placeholder)
func (pp *PerformanceProfiler) getDiskStats() DiskMetrics {
	// This would typically read from /proc/diskstats or use system APIs
	return DiskMetrics{
		ReadBytes:   500000,
		WriteBytes:  300000,
		ReadOps:     100,
		WriteOps:    50,
		Utilization: 0.3,
	}
}

// recordSample records a performance sample
func (pp *PerformanceProfiler) recordSample(metric, component string, value float64, unit string) {
	sample := PerformanceSample{
		Timestamp: time.Now(),
		Component: component,
		Metric:    metric,
		Value:     value,
		Unit:      unit,
		Metadata:  make(map[string]interface{}),
	}

	// Add to bottleneck detector
	if pp.bottleneckDetector != nil {
		pp.bottleneckDetector.mu.Lock()
		pp.bottleneckDetector.samples = append(pp.bottleneckDetector.samples, sample)

		// Keep only recent samples
		cutoff := time.Now().Add(-pp.config.DetectionWindow)
		var recentSamples []PerformanceSample
		for _, s := range pp.bottleneckDetector.samples {
			if s.Timestamp.After(cutoff) {
				recentSamples = append(recentSamples, s)
			}
		}
		pp.bottleneckDetector.samples = recentSamples
		pp.bottleneckDetector.mu.Unlock()
	}

	// Add to performance tracker
	pp.performanceTracker.mu.Lock()
	trendKey := fmt.Sprintf("%s_%s", component, metric)
	trend, exists := pp.performanceTracker.trends[trendKey]
	if !exists {
		trend = &PerformanceTrend{
			Metric:    metric,
			Component: component,
			Trend:     "stable",
			StartTime: time.Now(),
			Samples:   make([]float64, 0),
		}
		pp.performanceTracker.trends[trendKey] = trend
	}

	trend.Samples = append(trend.Samples, value)
	trend.LastUpdate = time.Now()

	// Keep only recent samples for trend analysis
	if len(trend.Samples) > 100 {
		trend.Samples = trend.Samples[len(trend.Samples)-100:]
	}
	pp.performanceTracker.mu.Unlock()
}

// bottleneckDetectionLoop performs bottleneck detection
func (pp *PerformanceProfiler) bottleneckDetectionLoop() {
	ticker := time.NewTicker(pp.config.ProfilingInterval * 2)
	defer ticker.Stop()

	for {
		select {
		case <-pp.ctx.Done():
			return
		case <-ticker.C:
			pp.detectBottlenecks()
		}
	}
}

// detectBottlenecks detects performance bottlenecks
func (pp *PerformanceProfiler) detectBottlenecks() {
	pp.bottleneckDetector.mu.Lock()
	defer pp.bottleneckDetector.mu.Unlock()

	if len(pp.bottleneckDetector.samples) < pp.config.MinSampleSize {
		return
	}

	// Analyze samples for bottlenecks
	metricGroups := make(map[string][]PerformanceSample)
	for _, sample := range pp.bottleneckDetector.samples {
		key := fmt.Sprintf("%s_%s", sample.Component, sample.Metric)
		metricGroups[key] = append(metricGroups[key], sample)
	}

	for key, samples := range metricGroups {
		if bottleneck := pp.analyzeForBottleneck(key, samples); bottleneck != nil {
			pp.bottleneckDetector.bottlenecks[bottleneck.ID] = bottleneck

			log.Warn().
				Str("bottleneck_id", bottleneck.ID).
				Str("type", string(bottleneck.Type)).
				Str("severity", string(bottleneck.Severity)).
				Float64("impact", bottleneck.Impact).
				Msg("Performance bottleneck detected")
		}
	}
}

// analyzeForBottleneck analyzes samples for bottleneck patterns
func (pp *PerformanceProfiler) analyzeForBottleneck(key string, samples []PerformanceSample) *Bottleneck {
	if len(samples) < pp.config.MinSampleSize {
		return nil
	}

	// Calculate statistics
	var sum, max float64
	for _, sample := range samples {
		sum += sample.Value
		if sample.Value > max {
			max = sample.Value
		}
	}
	avg := sum / float64(len(samples))

	// Determine bottleneck type and severity
	var bottleneckType BottleneckType
	var severity BottleneckSeverity
	var threshold float64

	if samples[0].Metric == "cpu_usage" {
		bottleneckType = BottleneckTypeCPU
		threshold = pp.config.CPUThreshold
	} else if samples[0].Metric == "memory_usage" {
		bottleneckType = BottleneckTypeMemory
		threshold = pp.config.MemoryThreshold
	} else {
		return nil
	}

	// Check if threshold is exceeded
	if avg <= threshold {
		return nil
	}

	// Determine severity
	if avg > threshold*1.5 {
		severity = BottleneckSeverityCritical
	} else if avg > threshold*1.2 {
		severity = BottleneckSeverityHigh
	} else {
		severity = BottleneckSeverityMedium
	}

	// Calculate impact and confidence
	impact := (avg - threshold) / threshold * 100
	confidence := float64(len(samples)) / float64(pp.config.MinSampleSize*2)
	if confidence > 1.0 {
		confidence = 1.0
	}

	bottleneck := &Bottleneck{
		ID:          fmt.Sprintf("bottleneck-%s-%d", key, time.Now().Unix()),
		Type:        bottleneckType,
		Severity:    severity,
		Component:   samples[0].Component,
		Description: fmt.Sprintf("%s usage is %.1f%% (threshold: %.1f%%)", samples[0].Metric, avg*100, threshold*100),
		Impact:      impact,
		Confidence:  confidence,
		DetectedAt:  time.Now(),
		Evidence:    samples,
		Suggestions: pp.generateSuggestions(bottleneckType, avg, threshold),
		Metadata:    make(map[string]interface{}),
	}

	return bottleneck
}

// generateSuggestions generates optimization suggestions for bottlenecks
func (pp *PerformanceProfiler) generateSuggestions(bottleneckType BottleneckType, value, threshold float64) []string {
	suggestions := make([]string, 0)

	switch bottleneckType {
	case BottleneckTypeCPU:
		suggestions = append(suggestions, "Consider optimizing CPU-intensive algorithms")
		suggestions = append(suggestions, "Implement goroutine pooling to reduce context switching")
		suggestions = append(suggestions, "Profile CPU usage to identify hot spots")
	case BottleneckTypeMemory:
		suggestions = append(suggestions, "Implement memory pooling for frequent allocations")
		suggestions = append(suggestions, "Optimize garbage collection settings")
		suggestions = append(suggestions, "Review memory usage patterns and reduce allocations")
	}

	return suggestions
}

// trendAnalysisLoop performs trend analysis
func (pp *PerformanceProfiler) trendAnalysisLoop() {
	ticker := time.NewTicker(pp.config.ProfilingInterval * 5)
	defer ticker.Stop()

	for {
		select {
		case <-pp.ctx.Done():
			return
		case <-ticker.C:
			pp.analyzeTrends()
		}
	}
}

// analyzeTrends analyzes performance trends
func (pp *PerformanceProfiler) analyzeTrends() {
	pp.performanceTracker.mu.Lock()
	defer pp.performanceTracker.mu.Unlock()

	for _, trend := range pp.performanceTracker.trends {
		if len(trend.Samples) < 10 {
			continue
		}

		// Simple linear regression to detect trend
		n := float64(len(trend.Samples))
		var sumX, sumY, sumXY, sumX2 float64

		for i, y := range trend.Samples {
			x := float64(i)
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}

		// Calculate slope
		slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
		trend.Slope = slope

		// Determine trend direction
		if slope > 0.01 {
			trend.Trend = "degrading"
		} else if slope < -0.01 {
			trend.Trend = "improving"
		} else {
			trend.Trend = "stable"
		}

		// Calculate confidence (simplified)
		trend.Confidence = minFloat(n/50.0, 1.0)
	}
}

// GetCurrentMetrics returns current resource metrics
func (pp *PerformanceProfiler) GetCurrentMetrics() *ResourceMetrics {
	pp.resourceMonitor.mu.RLock()
	defer pp.resourceMonitor.mu.RUnlock()
	return pp.resourceMonitor.metrics
}

// GetBottlenecks returns detected bottlenecks
func (pp *PerformanceProfiler) GetBottlenecks() map[string]*Bottleneck {
	if pp.bottleneckDetector == nil {
		return make(map[string]*Bottleneck)
	}

	pp.bottleneckDetector.mu.RLock()
	defer pp.bottleneckDetector.mu.RUnlock()

	bottlenecks := make(map[string]*Bottleneck)
	for k, v := range pp.bottleneckDetector.bottlenecks {
		bottlenecks[k] = v
	}

	return bottlenecks
}

// GetTrends returns performance trends
func (pp *PerformanceProfiler) GetTrends() map[string]*PerformanceTrend {
	pp.performanceTracker.mu.RLock()
	defer pp.performanceTracker.mu.RUnlock()

	trends := make(map[string]*PerformanceTrend)
	for k, v := range pp.performanceTracker.trends {
		trends[k] = v
	}

	return trends
}

// GenerateReport generates a comprehensive profiling report
func (pp *PerformanceProfiler) GenerateReport() *ProfilingReport {
	report := &ProfilingReport{
		GeneratedAt:     time.Now(),
		Duration:        pp.config.ProfilingInterval,
		ResourceMetrics: pp.GetCurrentMetrics(),
		Trends:          pp.GetTrends(),
		Recommendations: make([]string, 0),
	}

	// Get bottlenecks
	bottlenecks := pp.GetBottlenecks()
	for _, bottleneck := range bottlenecks {
		report.Bottlenecks = append(report.Bottlenecks, *bottleneck)
		report.Recommendations = append(report.Recommendations, bottleneck.Suggestions...)
	}

	// Generate summary
	report.Summary = pp.generateSummary(report)

	return report
}

// generateSummary generates a summary of the profiling report
func (pp *PerformanceProfiler) generateSummary(report *ProfilingReport) string {
	summary := fmt.Sprintf("Performance Report Summary\n")
	summary += fmt.Sprintf("CPU Usage: %.1f%%\n", report.ResourceMetrics.CPUUsage*100)
	summary += fmt.Sprintf("Memory Usage: %.1f%%\n", report.ResourceMetrics.MemoryUsage*100)
	summary += fmt.Sprintf("Goroutines: %d\n", report.ResourceMetrics.GoroutineCount)
	summary += fmt.Sprintf("Bottlenecks Detected: %d\n", len(report.Bottlenecks))

	if len(report.Bottlenecks) > 0 {
		summary += "Critical Issues:\n"
		for _, bottleneck := range report.Bottlenecks {
			if bottleneck.Severity == BottleneckSeverityCritical {
				summary += fmt.Sprintf("- %s: %s\n", bottleneck.Type, bottleneck.Description)
			}
		}
	}

	return summary
}

// Shutdown gracefully shuts down the performance profiler
func (pp *PerformanceProfiler) Shutdown() error {
	pp.cancel()
	log.Info().Msg("Performance profiler stopped")
	return nil
}

// Helper function for minFloat
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// DefaultProfilerConfig returns default profiler configuration
func DefaultProfilerConfig() *ProfilerConfig {
	return &ProfilerConfig{
		Enabled:                  true,
		RealTimeMonitoring:       true,
		BottleneckDetection:      true,
		ProfilingInterval:        30 * time.Second,
		CPUThreshold:             0.8,
		MemoryThreshold:          0.85,
		LatencyThreshold:         100 * time.Millisecond,
		DetectionWindow:          5 * time.Minute,
		MinSampleSize:            10,
		ConfidenceLevel:          0.8,
		EnableCPUProfiling:       true,
		EnableMemoryProfiling:    true,
		EnableGoroutineProfiling: true,
	}
}
