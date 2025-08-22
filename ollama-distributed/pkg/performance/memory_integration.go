package performance

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/memory"
)

// MemoryIntegrationManager integrates optimized memory management with performance monitoring
type MemoryIntegrationManager struct {
	memoryManager *memory.OptimizedMemoryManager
	gcOptimizer   *memory.GCOptimizer
	profiler      *PerformanceProfiler
	
	// Performance tracking
	metrics *IntegratedMetrics
	
	// Configuration
	config *IntegrationConfig
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// IntegrationConfig holds configuration for memory-performance integration
type IntegrationConfig struct {
	// Memory optimization
	EnableMemoryPools     bool          `yaml:"enable_memory_pools" json:"enable_memory_pools"`
	EnableGCOptimization  bool          `yaml:"enable_gc_optimization" json:"enable_gc_optimization"`
	
	// Performance monitoring
	EnablePerformanceMonitoring bool          `yaml:"enable_performance_monitoring" json:"enable_performance_monitoring"`
	MetricsInterval            time.Duration `yaml:"metrics_interval" json:"metrics_interval"`
	
	// Adaptive thresholds
	MemoryPressureThreshold    float64 `yaml:"memory_pressure_threshold" json:"memory_pressure_threshold"`
	GCLatencyThreshold         time.Duration `yaml:"gc_latency_threshold" json:"gc_latency_threshold"`
	AllocationRateThreshold    uint64  `yaml:"allocation_rate_threshold" json:"allocation_rate_threshold"`
	
	// Auto-tuning
	EnableAutoTuning          bool          `yaml:"enable_auto_tuning" json:"enable_auto_tuning"`
	AutoTuningInterval        time.Duration `yaml:"auto_tuning_interval" json:"auto_tuning_interval"`
	AdaptiveOptimization      bool          `yaml:"adaptive_optimization" json:"adaptive_optimization"`
}

// IntegratedMetrics combines memory and performance metrics
type IntegratedMetrics struct {
	// Memory pool metrics
	PoolMetrics *memory.PoolMetrics `json:"pool_metrics"`
	
	// GC optimization metrics
	GCMetrics *memory.GCOptimizationStats `json:"gc_metrics"`
	
	// Performance metrics
	CPUUsage        float64       `json:"cpu_usage"`
	MemoryUsage     uint64        `json:"memory_usage"`
	AllocationRate  uint64        `json:"allocation_rate"`
	GCLatency       time.Duration `json:"gc_latency"`
	ThroughputRPS   float64       `json:"throughput_rps"`
	
	// Integration metrics
	OptimizationScore     float64   `json:"optimization_score"`
	EfficiencyRating      float64   `json:"efficiency_rating"`
	ResourceUtilization   float64   `json:"resource_utilization"`
	LastOptimization      time.Time `json:"last_optimization"`
	AdaptiveAdjustments   int64     `json:"adaptive_adjustments"`
	
	// Trend analysis
	PerformanceTrend      string    `json:"performance_trend"` // "improving", "stable", "degrading"
	OptimizationImpact    float64   `json:"optimization_impact"`
	RecommendedActions    []string  `json:"recommended_actions"`
	
	Timestamp time.Time `json:"timestamp"`
}

// DefaultIntegrationConfig returns default integration configuration
func DefaultIntegrationConfig() *IntegrationConfig {
	return &IntegrationConfig{
		EnableMemoryPools:           true,
		EnableGCOptimization:        true,
		EnablePerformanceMonitoring: true,
		MetricsInterval:            30 * time.Second,
		MemoryPressureThreshold:    0.8,
		GCLatencyThreshold:         10 * time.Millisecond,
		AllocationRateThreshold:    1000000, // 1M allocations per interval
		EnableAutoTuning:           true,
		AutoTuningInterval:         5 * time.Minute,
		AdaptiveOptimization:       true,
	}
}

// NewMemoryIntegrationManager creates a new integrated memory-performance manager
func NewMemoryIntegrationManager(config *IntegrationConfig) *MemoryIntegrationManager {
	if config == nil {
		config = DefaultIntegrationConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &MemoryIntegrationManager{
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		metrics: &IntegratedMetrics{Timestamp: time.Now()},
	}
	
	// Initialize memory management if enabled
	if config.EnableMemoryPools {
		memConfig := memory.DefaultOptimizedConfig()
		memConfig.EnableGCOptimization = config.EnableGCOptimization
		manager.memoryManager = memory.NewOptimizedMemoryManager(memConfig)
		
		if config.EnableGCOptimization {
			manager.gcOptimizer = memory.NewGCOptimizer(memConfig)
		}
	}
	
	// Initialize performance profiler if enabled
	if config.EnablePerformanceMonitoring {
		manager.profiler = NewPerformanceProfiler()
	}
	
	return manager
}

// Start begins integrated memory and performance optimization
func (m *MemoryIntegrationManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Start memory management
	if m.memoryManager != nil {
		// Memory pools are ready by default
	}
	
	// Start GC optimization
	if m.gcOptimizer != nil && m.config.EnableGCOptimization {
		m.gcOptimizer.AutoTune()
	}
	
	// Start performance monitoring
	if m.profiler != nil {
		m.profiler.Start()
	}
	
	// Start metrics collection
	m.wg.Add(1)
	go m.metricsCollectionLoop()
	
	// Start auto-tuning if enabled
	if m.config.EnableAutoTuning {
		m.wg.Add(1)
		go m.autoTuningLoop()
	}
	
	return nil
}

// Stop stops the integrated manager
func (m *MemoryIntegrationManager) Stop() error {
	m.cancel()
	m.wg.Wait()
	
	if m.profiler != nil {
		m.profiler.Stop()
	}
	
	return nil
}

// GetOptimizedChannel returns a channel from the optimized pool
func (m *MemoryIntegrationManager) GetOptimizedChannel(capacity int) chan interface{} {
	if m.memoryManager != nil {
		return m.memoryManager.GetChannel(capacity)
	}
	return make(chan interface{}, capacity)
}

// ReturnOptimizedChannel returns a channel to the optimized pool
func (m *MemoryIntegrationManager) ReturnOptimizedChannel(ch chan interface{}, capacity int) {
	if m.memoryManager != nil {
		m.memoryManager.ReturnChannel(ch, capacity)
	}
}

// GetOptimizedBuffer returns a buffer from the optimized pool
func (m *MemoryIntegrationManager) GetOptimizedBuffer(size int) []byte {
	if m.memoryManager != nil {
		return m.memoryManager.GetBuffer(size)
	}
	return make([]byte, size)
}

// ReturnOptimizedBuffer returns a buffer to the optimized pool
func (m *MemoryIntegrationManager) ReturnOptimizedBuffer(buf []byte) {
	if m.memoryManager != nil {
		m.memoryManager.ReturnBuffer(buf)
	}
}

// GetMetrics returns current integrated metrics
func (m *MemoryIntegrationManager) GetMetrics() *IntegratedMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Update metrics from components
	m.updateIntegratedMetrics()
	
	return m.metrics
}

// OptimizePerformance triggers immediate performance optimization
func (m *MemoryIntegrationManager) OptimizePerformance() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Trigger GC optimization
	if m.gcOptimizer != nil {
		m.gcOptimizer.OptimizeGC()
	}
	
	// Update memory pressure
	if m.memoryManager != nil {
		m.memoryManager.UpdateMemoryPressure()
	}
	
	// Analyze current performance
	m.analyzePerformance()
	
	// Apply adaptive optimizations
	if m.config.AdaptiveOptimization {
		m.applyAdaptiveOptimizations()
	}
	
	m.metrics.LastOptimization = time.Now()
	m.metrics.AdaptiveAdjustments++
	
	return nil
}

// metricsCollectionLoop runs the metrics collection loop
func (m *MemoryIntegrationManager) metricsCollectionLoop() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectMetrics()
		}
	}
}

// autoTuningLoop runs the auto-tuning loop
func (m *MemoryIntegrationManager) autoTuningLoop() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.config.AutoTuningInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.OptimizePerformance()
		}
	}
}

// collectMetrics collects metrics from all components
func (m *MemoryIntegrationManager) collectMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.updateIntegratedMetrics()
	m.analyzePerformance()
	
	// Apply adaptive optimizations based on metrics
	if m.config.AdaptiveOptimization {
		m.applyAdaptiveOptimizations()
	}
}

// updateIntegratedMetrics updates metrics from all components
func (m *MemoryIntegrationManager) updateIntegratedMetrics() {
	// Update memory pool metrics
	if m.memoryManager != nil {
		m.metrics.PoolMetrics = m.memoryManager.GetMetrics()
	}
	
	// Update GC metrics
	if m.gcOptimizer != nil {
		m.metrics.GCMetrics = m.gcOptimizer.GetOptimizationStats()
	}
	
	// Update runtime metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	m.metrics.MemoryUsage = memStats.Alloc
	m.metrics.AllocationRate = memStats.Mallocs - memStats.Frees
	
	if memStats.PauseTotalNs > 0 && memStats.NumGC > 0 {
		m.metrics.GCLatency = time.Duration(memStats.PauseTotalNs / uint64(memStats.NumGC))
	}
	
	m.metrics.Timestamp = time.Now()
}

// analyzePerformance analyzes current performance and generates insights
func (m *MemoryIntegrationManager) analyzePerformance() {
	// Calculate optimization score based on multiple factors
	score := m.calculateOptimizationScore()
	m.metrics.OptimizationScore = score
	
	// Calculate efficiency rating
	efficiency := m.calculateEfficiencyRating()
	m.metrics.EfficiencyRating = efficiency
	
	// Determine performance trend
	trend := m.analyzePerformanceTrend()
	m.metrics.PerformanceTrend = trend
	
	// Generate recommendations
	recommendations := m.generateRecommendations()
	m.metrics.RecommendedActions = recommendations
}

// calculateOptimizationScore calculates overall optimization score (0-100)
func (m *MemoryIntegrationManager) calculateOptimizationScore() float64 {
	score := 100.0
	
	// Memory pressure impact
	if m.memoryManager != nil {
		memStats := m.metrics.PoolMetrics
		if memStats != nil && memStats.MemoryPressure > int64(m.config.MemoryPressureThreshold*100) {
			score -= 30.0
		}
	}
	
	// GC latency impact
	if m.metrics.GCLatency > m.config.GCLatencyThreshold {
		score -= 25.0
	}
	
	// Pool efficiency impact
	if m.metrics.PoolMetrics != nil {
		poolEfficiency := (m.metrics.PoolMetrics.ChannelReuseRate + m.metrics.PoolMetrics.BufferReuseRate) / 2
		score *= poolEfficiency
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// calculateEfficiencyRating calculates resource efficiency rating (0-1)
func (m *MemoryIntegrationManager) calculateEfficiencyRating() float64 {
	efficiency := 1.0
	
	// Memory efficiency
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	if memStats.Sys > 0 {
		memoryEfficiency := 1.0 - float64(memStats.Alloc)/float64(memStats.Sys)
		efficiency *= memoryEfficiency
	}
	
	// Pool reuse efficiency
	if m.metrics.PoolMetrics != nil {
		poolEfficiency := (m.metrics.PoolMetrics.ChannelReuseRate + m.metrics.PoolMetrics.BufferReuseRate) / 2
		efficiency *= poolEfficiency
	}
	
	return efficiency
}

// analyzePerformanceTrend analyzes performance trend
func (m *MemoryIntegrationManager) analyzePerformanceTrend() string {
	// Simple trend analysis based on optimization score
	if m.metrics.OptimizationScore > 80 {
		return "improving"
	} else if m.metrics.OptimizationScore > 60 {
		return "stable"
	}
	return "degrading"
}

// generateRecommendations generates performance improvement recommendations
func (m *MemoryIntegrationManager) generateRecommendations() []string {
	var recommendations []string
	
	// Memory pressure recommendations
	if m.metrics.PoolMetrics != nil && m.metrics.PoolMetrics.MemoryPressure > int64(m.config.MemoryPressureThreshold*100) {
		recommendations = append(recommendations, "High memory pressure detected - consider increasing memory limits or reducing allocation rate")
	}
	
	// GC latency recommendations
	if m.metrics.GCLatency > m.config.GCLatencyThreshold {
		recommendations = append(recommendations, "High GC latency detected - consider tuning GC parameters or reducing allocation pressure")
	}
	
	// Pool efficiency recommendations
	if m.metrics.PoolMetrics != nil {
		if m.metrics.PoolMetrics.ChannelReuseRate < 0.5 {
			recommendations = append(recommendations, "Low channel pool reuse rate - consider increasing pool sizes or improving usage patterns")
		}
		if m.metrics.PoolMetrics.BufferReuseRate < 0.5 {
			recommendations = append(recommendations, "Low buffer pool reuse rate - consider optimizing buffer size classes")
		}
	}
	
	// Optimization recommendations
	if m.metrics.OptimizationScore < 70 {
		recommendations = append(recommendations, "Performance optimization needed - enable auto-tuning and adaptive optimization")
	}
	
	return recommendations
}

// applyAdaptiveOptimizations applies optimizations based on current metrics
func (m *MemoryIntegrationManager) applyAdaptiveOptimizations() {
	// GC optimization adjustments
	if m.gcOptimizer != nil && m.metrics.GCLatency > m.config.GCLatencyThreshold {
		m.gcOptimizer.OptimizeGC()
	}
	
	// Memory pressure management
	if m.memoryManager != nil && m.metrics.PoolMetrics != nil {
		if m.metrics.PoolMetrics.MemoryPressure > int64(m.config.MemoryPressureThreshold*100) {
			m.memoryManager.Cleanup()
		}
	}
}