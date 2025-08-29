package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"


	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/memory"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/network"
	"github.com/rs/zerolog/log"
)

// SystemOptimizer provides comprehensive system-wide performance optimizations
type SystemOptimizer struct {
	config *OptimizerConfig

	// Component optimizers
	memoryManager    *memory.Manager
	networkOptimizer *network.Optimizer
	cacheManager     *CacheManager
	gcOptimizer      *GCOptimizer
	connectionPool   *ConnectionPool

	// Performance monitoring
	metrics *PerformanceMetrics
	monitor *PerformanceMonitor

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// OptimizerConfig holds performance optimization configuration
type OptimizerConfig struct {
	// System optimization settings
	EnableGCTuning       bool          `yaml:"enable_gc_tuning"`
	EnableMemoryPools    bool          `yaml:"enable_memory_pools"`
	EnableConnectionPool bool          `yaml:"enable_connection_pool"`
	EnableRequestBatching bool         `yaml:"enable_request_batching"`

	// Performance targets
	TargetThroughputOPS int           `yaml:"target_throughput_ops"`    // ops/sec
	TargetLatencyP99MS  int           `yaml:"target_latency_p99_ms"`    // milliseconds
	MaxMemoryUsageMB    int           `yaml:"max_memory_usage_mb"`      // megabytes
	MaxCPUUsagePercent  float64       `yaml:"max_cpu_usage_percent"`    // percentage

	// GC optimization settings
	GCTargetPercent     int           `yaml:"gc_target_percent"`
	GCMaxPause          time.Duration `yaml:"gc_max_pause"`
	GCMemoryLimit       int64         `yaml:"gc_memory_limit"`          // bytes

	// Connection pool settings
	MaxConnections      int           `yaml:"max_connections"`
	MaxIdleConnections  int           `yaml:"max_idle_connections"`
	ConnectionTimeout   time.Duration `yaml:"connection_timeout"`
	IdleTimeout         time.Duration `yaml:"idle_timeout"`

	// Batch processing settings
	BatchSize           int           `yaml:"batch_size"`
	BatchTimeout        time.Duration `yaml:"batch_timeout"`
	MaxConcurrentBatch  int           `yaml:"max_concurrent_batch"`

	// Monitoring settings
	MetricsInterval     time.Duration `yaml:"metrics_interval"`
	PerformanceLogging  bool          `yaml:"performance_logging"`
}

// DefaultOptimizerConfig returns default performance optimization configuration
func DefaultOptimizerConfig() *OptimizerConfig {
	return &OptimizerConfig{
		EnableGCTuning:       true,
		EnableMemoryPools:    true,
		EnableConnectionPool: true,
		EnableRequestBatching: true,

		TargetThroughputOPS: 500,
		TargetLatencyP99MS:  10,
		MaxMemoryUsageMB:    1024,
		MaxCPUUsagePercent:  25.0,

		GCTargetPercent: 50,
		GCMaxPause:      5 * time.Millisecond,
		GCMemoryLimit:   1 << 30, // 1GB

		MaxConnections:     100,
		MaxIdleConnections: 25,
		ConnectionTimeout:  5 * time.Second,
		IdleTimeout:        30 * time.Second,

		BatchSize:          100,
		BatchTimeout:       10 * time.Millisecond,
		MaxConcurrentBatch: 10,

		MetricsInterval:    10 * time.Second,
		PerformanceLogging: true,
	}
}

// PerformanceMetrics tracks system performance metrics
type PerformanceMetrics struct {
	// Throughput metrics
	RequestsPerSecond   float64   `json:"requests_per_second"`
	OperationsPerSecond float64   `json:"operations_per_second"`
	BytesPerSecond      float64   `json:"bytes_per_second"`

	// Latency metrics
	AverageLatency time.Duration `json:"average_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	MaxLatency     time.Duration `json:"max_latency"`

	// Resource metrics
	CPUUsagePercent   float64 `json:"cpu_usage_percent"`
	MemoryUsageMB     float64 `json:"memory_usage_mb"`
	GoroutineCount    int     `json:"goroutine_count"`
	GCPauseMS         float64 `json:"gc_pause_ms"`

	// Connection metrics
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	ConnectionErrors  int `json:"connection_errors"`

	// Cache metrics
	CacheHitRate     float64 `json:"cache_hit_rate"`
	CacheMissRate    float64 `json:"cache_miss_rate"`
	CacheEvictions   int64   `json:"cache_evictions"`

	// Batch processing metrics
	BatchesProcessed int64 `json:"batches_processed"`
	BatchProcessTime time.Duration `json:"batch_process_time"`
	QueueDepth       int   `json:"queue_depth"`

	LastUpdated time.Time `json:"last_updated"`
	mu          sync.RWMutex
}

// NewSystemOptimizer creates a new system performance optimizer
func NewSystemOptimizer(config *OptimizerConfig) (*SystemOptimizer, error) {
	if config == nil {
		config = DefaultOptimizerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize component optimizers
	memConfig := memory.DefaultConfig()
	memConfig.MaxMemoryMB = int64(config.MaxMemoryUsageMB)
	memConfig.GCTargetPercent = config.GCTargetPercent
	memoryManager := memory.NewManager(memConfig)

	netConfig := network.DefaultConfig()
	networkOptimizer := network.NewOptimizer(netConfig)

	cacheManager := NewCacheManager(config)
	gcOptimizer := NewGCOptimizer(config)
	connectionPool := NewConnectionPool(config)

	optimizer := &SystemOptimizer{
		config:           config,
		memoryManager:    memoryManager,
		networkOptimizer: networkOptimizer,
		cacheManager:     cacheManager,
		gcOptimizer:      gcOptimizer,
		connectionPool:   connectionPool,
		metrics:          &PerformanceMetrics{LastUpdated: time.Now()},
		monitor:          NewPerformanceMonitor(config),
		ctx:              ctx,
		cancel:           cancel,
	}

	return optimizer, nil
}

// Start starts the performance optimizer
func (so *SystemOptimizer) Start() error {
	log.Info().Msg("Starting system performance optimizer")

	// Start component optimizers
	if err := so.memoryManager.Start(); err != nil {
		return fmt.Errorf("failed to start memory manager: %w", err)
	}

	if err := so.cacheManager.Start(); err != nil {
		return fmt.Errorf("failed to start cache manager: %w", err)
	}

	if err := so.connectionPool.Start(); err != nil {
		return fmt.Errorf("failed to start connection pool: %w", err)
	}

	// Start performance monitoring
	so.wg.Add(1)
	go so.runPerformanceMonitoring()

	// Start GC optimization
	if so.config.EnableGCTuning {
		so.wg.Add(1)
		go so.runGCOptimization()
	}

	// Start automatic optimization
	so.wg.Add(1)
	go so.runAutoOptimization()

	log.Info().
		Int("target_throughput", so.config.TargetThroughputOPS).
		Int("target_p99_latency", so.config.TargetLatencyP99MS).
		Msg("System optimizer started")

	return nil
}

// Stop stops the performance optimizer
func (so *SystemOptimizer) Stop() error {
	log.Info().Msg("Stopping system performance optimizer")

	so.cancel()
	so.wg.Wait()

	// Stop component optimizers
	if so.memoryManager != nil {
		so.memoryManager.Stop()
	}
	if so.cacheManager != nil {
		so.cacheManager.Stop()
	}
	if so.connectionPool != nil {
		so.connectionPool.Stop()
	}

	return nil
}

// OptimizeRequest optimizes a single request processing
func (so *SystemOptimizer) OptimizeRequest(ctx context.Context, requestData []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		so.updateLatencyMetrics(time.Since(start))
	}()

	// Use connection pool for network operations
	conn, err := so.connectionPool.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	defer so.connectionPool.Put(conn)

	// Use memory pool for temporary allocations
	bufferSize := len(requestData) * 2
	if bufferSize < 1024 {
		bufferSize = 1024
	}
	bufferPool := so.memoryManager.GetPool("request-buffer", bufferSize)
	buffer := bufferPool.Get()
	defer bufferPool.Put(buffer)

	// Use cache for frequently accessed data
	cacheKey := fmt.Sprintf("request_%x", requestData[:min(32, len(requestData))])
	if cached, found := so.cacheManager.Get("request-cache", cacheKey); found {
		return cached.([]byte), nil
	}

	// Process request with optimizations
	result, err := so.processOptimizedRequest(ctx, requestData, buffer)
	if err != nil {
		return nil, err
	}

	// Cache result for future requests
	so.cacheManager.Set("request-cache", cacheKey, result, 5*time.Minute)

	return result, nil
}

// GetMetrics returns current performance metrics
func (so *SystemOptimizer) GetMetrics() *PerformanceMetrics {
	so.metrics.mu.RLock()
	defer so.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *so.metrics
	return &metrics
}

// runPerformanceMonitoring runs the performance monitoring loop
func (so *SystemOptimizer) runPerformanceMonitoring() {
	defer so.wg.Done()

	ticker := time.NewTicker(so.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-so.ctx.Done():
			return
		case <-ticker.C:
			so.updateMetrics()
		}
	}
}

// updateMetrics updates performance metrics
func (so *SystemOptimizer) updateMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	so.metrics.mu.Lock()
	defer so.metrics.mu.Unlock()

	// Update memory metrics
	so.metrics.MemoryUsageMB = float64(m.Alloc) / 1024 / 1024
	so.metrics.GoroutineCount = runtime.NumGoroutine()
	so.metrics.GCPauseMS = float64(m.PauseTotalNs) / 1e6

	// Update cache metrics
	cacheStats := so.cacheManager.GetStats()
	totalRequests := cacheStats.TotalHits + cacheStats.TotalMisses
	if totalRequests > 0 {
		so.metrics.CacheHitRate = float64(cacheStats.TotalHits) / float64(totalRequests) * 100
		so.metrics.CacheMissRate = float64(cacheStats.TotalMisses) / float64(totalRequests) * 100
	}
	so.metrics.CacheEvictions = cacheStats.TotalEvictions

	// Update connection metrics
	poolStats := so.connectionPool.GetStats()
	so.metrics.ActiveConnections = poolStats.ActiveConnections
	so.metrics.IdleConnections = poolStats.IdleConnections
	so.metrics.ConnectionErrors = int(poolStats.Errors)

	so.metrics.LastUpdated = time.Now()

	// Log performance metrics if enabled
	if so.config.PerformanceLogging {
		log.Info().
			Float64("memory_mb", so.metrics.MemoryUsageMB).
			Int("goroutines", so.metrics.GoroutineCount).
			Float64("gc_pause_ms", so.metrics.GCPauseMS).
			Float64("cache_hit_rate", so.metrics.CacheHitRate).
			Int("active_connections", so.metrics.ActiveConnections).
			Msg("Performance metrics updated")
	}
}

// updateLatencyMetrics updates request latency metrics
func (so *SystemOptimizer) updateLatencyMetrics(latency time.Duration) {
	so.metrics.mu.Lock()
	defer so.metrics.mu.Unlock()

	// Simple moving average for average latency
	if so.metrics.AverageLatency == 0 {
		so.metrics.AverageLatency = latency
	} else {
		so.metrics.AverageLatency = (so.metrics.AverageLatency + latency) / 2
	}

	// Update max latency
	if latency > so.metrics.MaxLatency {
		so.metrics.MaxLatency = latency
	}

	// Update P95 and P99 (simplified implementation)
	if latency > so.metrics.P99Latency {
		so.metrics.P99Latency = latency
		so.metrics.P95Latency = latency * 95 / 99
	}
}

// runGCOptimization runs garbage collection optimization
func (so *SystemOptimizer) runGCOptimization() {
	defer so.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-so.ctx.Done():
			return
		case <-ticker.C:
			so.gcOptimizer.Optimize()
		}
	}
}

// runAutoOptimization runs automatic performance optimization
func (so *SystemOptimizer) runAutoOptimization() {
	defer so.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-so.ctx.Done():
			return
		case <-ticker.C:
			so.performAutoOptimization()
		}
	}
}

// performAutoOptimization performs automatic system optimization
func (so *SystemOptimizer) performAutoOptimization() {
	metrics := so.GetMetrics()

	// Optimize based on current metrics
	if metrics.MemoryUsageMB > float64(so.config.MaxMemoryUsageMB)*0.8 {
		log.Info().
			Float64("current_memory", metrics.MemoryUsageMB).
			Int("max_memory", so.config.MaxMemoryUsageMB).
			Msg("High memory usage detected, triggering optimization")

		// Trigger aggressive garbage collection
		so.memoryManager.ForceGC()
		
		// Clear cache entries to free memory
		so.cacheManager.ClearExpired()
	}

	// Adjust GC target if latency is too high
	if metrics.P99Latency > time.Duration(so.config.TargetLatencyP99MS)*time.Millisecond {
		log.Info().
			Dur("current_p99", metrics.P99Latency).
			Int("target_p99", so.config.TargetLatencyP99MS).
			Msg("High latency detected, adjusting GC settings")

		so.gcOptimizer.AdjustForLowLatency()
	}

	// Scale connection pool based on usage
	if metrics.ActiveConnections > so.config.MaxConnections*8/10 {
		so.connectionPool.Scale(so.config.MaxConnections + 10)
	}
}

// processOptimizedRequest processes a request with optimizations applied
func (so *SystemOptimizer) processOptimizedRequest(ctx context.Context, requestData, buffer []byte) ([]byte, error) {
	// Compress request data if beneficial
	compressed, wasCompressed, err := so.networkOptimizer.CompressData(requestData)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	dataToProcess := requestData
	if wasCompressed {
		dataToProcess = compressed
	}

	// Simulate request processing
	if len(buffer) >= len(dataToProcess) {
		copy(buffer, dataToProcess)
	}
	
	// Simulate some processing time
	time.Sleep(time.Microsecond * 100)

	// Return processed data
	result := make([]byte, len(dataToProcess))
	copy(result, dataToProcess)

	return result, nil
}

// IsHealthy returns true if the optimizer is performing within targets
func (so *SystemOptimizer) IsHealthy() bool {
	metrics := so.GetMetrics()

	// Check if we're meeting performance targets
	return metrics.P99Latency <= time.Duration(so.config.TargetLatencyP99MS)*time.Millisecond &&
		metrics.MemoryUsageMB <= float64(so.config.MaxMemoryUsageMB) &&
		metrics.CPUUsagePercent <= so.config.MaxCPUUsagePercent
}

// GetOptimizationRecommendations returns performance optimization recommendations
func (so *SystemOptimizer) GetOptimizationRecommendations() []string {
	metrics := so.GetMetrics()
	recommendations := make([]string, 0)

	if metrics.P99Latency > time.Duration(so.config.TargetLatencyP99MS)*time.Millisecond {
		recommendations = append(recommendations,
			"Consider increasing connection pool size or reducing GC frequency")
	}

	if metrics.MemoryUsageMB > float64(so.config.MaxMemoryUsageMB)*0.9 {
		recommendations = append(recommendations,
			"Memory usage is high - consider reducing cache size or increasing GC frequency")
	}

	if metrics.CacheHitRate < 80 && metrics.CacheHitRate > 0 {
		recommendations = append(recommendations,
			"Low cache hit rate - consider increasing cache size or adjusting TTL")
	}

	if metrics.GoroutineCount > 1000 {
		recommendations = append(recommendations,
			"High goroutine count detected - check for goroutine leaks")
	}

	return recommendations
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}