package memory

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// BoundedMemoryManager provides memory bounds enforcement and optimization
type BoundedMemoryManager struct {
	// Configuration
	config *MemoryConfig
	
	// Memory tracking
	currentUsage   int64
	peakUsage      int64
	allocations    int64
	deallocations  int64
	
	// Memory pools for frequent operations
	bufferPools    map[int]*sync.Pool
	objectPools    map[string]*sync.Pool
	
	// Garbage collection optimization
	gcStats        *GCStats
	gcOptimizer    *GCOptimizer
	
	// Monitoring and alerts
	monitors       []MemoryMonitor
	alerts         chan MemoryAlert
	
	// Lifecycle
	stopCh         chan struct{}
	mu             sync.RWMutex
}

// MemoryConfig holds memory management configuration
type MemoryConfig struct {
	// Limits
	MaxMemoryUsage     int64         `json:"max_memory_usage"`      // bytes
	WarningThreshold   float64       `json:"warning_threshold"`     // percentage (0-1)
	CriticalThreshold  float64       `json:"critical_threshold"`    // percentage (0-1)
	
	// Request size limits
	MaxRequestSize     int64         `json:"max_request_size"`      // bytes
	MaxResponseSize    int64         `json:"max_response_size"`     // bytes
	MaxConcurrentReqs  int           `json:"max_concurrent_reqs"`
	
	// Memory pools
	EnablePools        bool          `json:"enable_pools"`
	PoolSizes          []int         `json:"pool_sizes"`            // Buffer sizes to pool
	MaxPoolObjects     int           `json:"max_pool_objects"`      // Per pool
	
	// GC optimization
	EnableGCOptimization bool        `json:"enable_gc_optimization"`
	GCTargetPercent      int         `json:"gc_target_percent"`
	GCMaxPause           time.Duration `json:"gc_max_pause"`
	
	// Monitoring
	MonitoringInterval   time.Duration `json:"monitoring_interval"`
	EnableAlerts         bool          `json:"enable_alerts"`
	AlertCooldown        time.Duration `json:"alert_cooldown"`
}

// GCStats tracks garbage collection statistics
type GCStats struct {
	TotalGCPauses      time.Duration `json:"total_gc_pauses"`
	GCCount            int64         `json:"gc_count"`
	LastGCPause        time.Duration `json:"last_gc_pause"`
	AverageGCPause     time.Duration `json:"average_gc_pause"`
	MaxGCPause         time.Duration `json:"max_gc_pause"`
	ForcedGCs          int64         `json:"forced_gcs"`
	mu sync.RWMutex
}

// GCOptimizer manages garbage collection optimization
type GCOptimizer struct {
	config              *MemoryConfig
	lastOptimization    time.Time
	optimizationCount   int64
	effectivenessScore  float64
}

// MemoryMonitor interface for memory monitoring components
type MemoryMonitor interface {
	OnMemoryUpdate(usage int64, percentage float64)
	OnMemoryAlert(alert MemoryAlert)
	OnGCEvent(stats *GCStats)
}

// MemoryAlert represents a memory-related alert
type MemoryAlert struct {
	Level       AlertLevel    `json:"level"`
	Message     string        `json:"message"`
	CurrentUsage int64        `json:"current_usage"`
	MaxUsage    int64         `json:"max_usage"`
	Percentage  float64       `json:"percentage"`
	Timestamp   time.Time     `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertLevel represents the severity of a memory alert
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
	AlertLevelEmergency
)

// NewBoundedMemoryManager creates a new bounded memory manager
func NewBoundedMemoryManager(config *MemoryConfig) (*BoundedMemoryManager, error) {
	if config == nil {
		config = DefaultMemoryConfig()
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid memory config: %w", err)
	}
	
	manager := &BoundedMemoryManager{
		config:      config,
		bufferPools: make(map[int]*sync.Pool),
		objectPools: make(map[string]*sync.Pool),
		gcStats:     &GCStats{},
		alerts:      make(chan MemoryAlert, 100),
		stopCh:      make(chan struct{}),
	}
	
	// Initialize GC optimizer
	if config.EnableGCOptimization {
		manager.gcOptimizer = &GCOptimizer{
			config: config,
			lastOptimization: time.Now(),
		}
		manager.optimizeGC()
	}
	
	// Initialize memory pools
	if config.EnablePools {
		manager.initializePools()
	}
	
	// Start monitoring routine
	go manager.monitoringRoutine()
	
	log.Info().
		Int64("max_memory", config.MaxMemoryUsage).
		Float64("warning_threshold", config.WarningThreshold).
		Float64("critical_threshold", config.CriticalThreshold).
		Msg("Bounded memory manager initialized")
	
	return manager, nil
}

// DefaultMemoryConfig returns default memory configuration
func DefaultMemoryConfig() *MemoryConfig {
	return &MemoryConfig{
		MaxMemoryUsage:       2 * 1024 * 1024 * 1024, // 2GB
		WarningThreshold:     0.75,                     // 75%
		CriticalThreshold:    0.90,                     // 90%
		MaxRequestSize:       100 * 1024 * 1024,       // 100MB
		MaxResponseSize:      100 * 1024 * 1024,       // 100MB
		MaxConcurrentReqs:    1000,
		EnablePools:          true,
		PoolSizes:           []int{1024, 4096, 16384, 65536, 262144}, // Common buffer sizes
		MaxPoolObjects:       100,
		EnableGCOptimization: true,
		GCTargetPercent:     50,
		GCMaxPause:          10 * time.Millisecond,
		MonitoringInterval:  30 * time.Second,
		EnableAlerts:        true,
		AlertCooldown:       5 * time.Minute,
	}
}

// AllocateBuffer allocates a buffer of specified size, using pools if available
func (m *BoundedMemoryManager) AllocateBuffer(size int) []byte {
	// Check memory limits
	if !m.checkAllocationLimit(int64(size)) {
		log.Warn().
			Int("requested_size", size).
			Int64("current_usage", atomic.LoadInt64(&m.currentUsage)).
			Msg("Memory allocation denied - would exceed limits")
		return nil
	}
	
	// Try to get from pool first
	if m.config.EnablePools {
		if buffer := m.getBufferFromPool(size); buffer != nil {
			m.trackAllocation(int64(len(buffer)))
			return buffer
		}
	}
	
	// Allocate new buffer
	buffer := make([]byte, size)
	m.trackAllocation(int64(size))
	
	return buffer
}

// ReleaseBuffer returns a buffer to the appropriate pool
func (m *BoundedMemoryManager) ReleaseBuffer(buffer []byte) {
	if buffer == nil {
		return
	}
	
	size := len(buffer)
	m.trackDeallocation(int64(size))
	
	// Return to pool if enabled and size matches
	if m.config.EnablePools {
		if m.returnBufferToPool(buffer) {
			return
		}
	}
	
	// Buffer will be garbage collected normally
}

// GetMemoryStats returns current memory statistics
func (m *BoundedMemoryManager) GetMemoryStats() *MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	current := atomic.LoadInt64(&m.currentUsage)
	peak := atomic.LoadInt64(&m.peakUsage)
	
	return &MemoryStats{
		CurrentUsage:      current,
		PeakUsage:        peak,
		MaxUsage:         m.config.MaxMemoryUsage,
		UsagePercentage:  float64(current) / float64(m.config.MaxMemoryUsage),
		Allocations:      atomic.LoadInt64(&m.allocations),
		Deallocations:    atomic.LoadInt64(&m.deallocations),
		SystemMemory:     &SystemMemoryStats{
			Alloc:        memStats.Alloc,
			TotalAlloc:   memStats.TotalAlloc,
			Sys:          memStats.Sys,
			NumGC:        memStats.NumGC,
			PauseTotalNs: memStats.PauseTotalNs,
		},
		GCStats: m.getGCStats(),
		Timestamp: time.Now(),
	}
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	CurrentUsage     int64                `json:"current_usage"`
	PeakUsage        int64                `json:"peak_usage"`
	MaxUsage         int64                `json:"max_usage"`
	UsagePercentage  float64              `json:"usage_percentage"`
	Allocations      int64                `json:"allocations"`
	Deallocations    int64                `json:"deallocations"`
	SystemMemory     *SystemMemoryStats   `json:"system_memory"`
	GCStats          *GCStats             `json:"gc_stats"`
	Timestamp        time.Time            `json:"timestamp"`
}

// SystemMemoryStats represents Go runtime memory statistics
type SystemMemoryStats struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"num_gc"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
}

// CheckRequestSize validates if a request size is within limits
func (m *BoundedMemoryManager) CheckRequestSize(size int64) error {
	if size > m.config.MaxRequestSize {
		return fmt.Errorf("request size %d exceeds maximum %d", size, m.config.MaxRequestSize)
	}
	return nil
}

// CheckResponseSize validates if a response size is within limits  
func (m *BoundedMemoryManager) CheckResponseSize(size int64) error {
	if size > m.config.MaxResponseSize {
		return fmt.Errorf("response size %d exceeds maximum %d", size, m.config.MaxResponseSize)
	}
	return nil
}

// ForceGC triggers garbage collection and optimizes if enabled
func (m *BoundedMemoryManager) ForceGC() {
	start := time.Now()
	runtime.GC()
	pause := time.Since(start)
	
	// Update GC stats
	m.gcStats.mu.Lock()
	m.gcStats.ForcedGCs++
	m.gcStats.GCCount++
	m.gcStats.LastGCPause = pause
	m.gcStats.TotalGCPauses += pause
	
	if pause > m.gcStats.MaxGCPause {
		m.gcStats.MaxGCPause = pause
	}
	
	m.gcStats.AverageGCPause = m.gcStats.TotalGCPauses / time.Duration(m.gcStats.GCCount)
	m.gcStats.mu.Unlock()
	
	log.Debug().
		Dur("pause", pause).
		Int64("forced_gcs", m.gcStats.ForcedGCs).
		Msg("Forced garbage collection completed")
}

// AddMonitor adds a memory monitor
func (m *BoundedMemoryManager) AddMonitor(monitor MemoryMonitor) {
	m.mu.Lock()
	m.monitors = append(m.monitors, monitor)
	m.mu.Unlock()
}

// Close gracefully shuts down the memory manager
func (m *BoundedMemoryManager) Close() error {
	close(m.stopCh)
	close(m.alerts)
	
	log.Info().Msg("Bounded memory manager closed")
	return nil
}

// Helper methods

func (m *BoundedMemoryManager) checkAllocationLimit(size int64) bool {
	current := atomic.LoadInt64(&m.currentUsage)
	if current+size > m.config.MaxMemoryUsage {
		return false
	}
	return true
}

func (m *BoundedMemoryManager) trackAllocation(size int64) {
	current := atomic.AddInt64(&m.currentUsage, size)
	atomic.AddInt64(&m.allocations, 1)
	
	// Update peak usage
	for {
		peak := atomic.LoadInt64(&m.peakUsage)
		if current <= peak || atomic.CompareAndSwapInt64(&m.peakUsage, peak, current) {
			break
		}
	}
	
	// Check thresholds
	percentage := float64(current) / float64(m.config.MaxMemoryUsage)
	m.checkThresholds(current, percentage)
}

func (m *BoundedMemoryManager) trackDeallocation(size int64) {
	atomic.AddInt64(&m.currentUsage, -size)
	atomic.AddInt64(&m.deallocations, 1)
}

func (m *BoundedMemoryManager) checkThresholds(usage int64, percentage float64) {
	if percentage >= m.config.CriticalThreshold {
		m.sendAlert(AlertLevelCritical, "Memory usage critical", usage, percentage)
		if m.config.EnableGCOptimization {
			m.ForceGC()
		}
	} else if percentage >= m.config.WarningThreshold {
		m.sendAlert(AlertLevelWarning, "Memory usage warning", usage, percentage)
	}
}

func (m *BoundedMemoryManager) sendAlert(level AlertLevel, message string, usage int64, percentage float64) {
	alert := MemoryAlert{
		Level:        level,
		Message:      message,
		CurrentUsage: usage,
		MaxUsage:     m.config.MaxMemoryUsage,
		Percentage:   percentage,
		Timestamp:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}
	
	select {
	case m.alerts <- alert:
		// Alert sent
	default:
		// Alert queue full, log directly
		log.Warn().
			Str("level", fmt.Sprintf("%d", level)).
			Str("message", message).
			Float64("percentage", percentage*100).
			Msg("Memory alert (queue full)")
	}
}

func (m *BoundedMemoryManager) initializePools() {
	for _, size := range m.config.PoolSizes {
		m.bufferPools[size] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		}
	}
	
	log.Debug().
		Ints("pool_sizes", m.config.PoolSizes).
		Msg("Memory pools initialized")
}

func (m *BoundedMemoryManager) getBufferFromPool(requestedSize int) []byte {
	// Find the smallest pool that can accommodate the request
	for _, poolSize := range m.config.PoolSizes {
		if poolSize >= requestedSize {
			if pool, exists := m.bufferPools[poolSize]; exists {
				if buffer := pool.Get().([]byte); buffer != nil {
					return buffer[:requestedSize] // Return slice of requested size
				}
			}
		}
	}
	return nil
}

func (m *BoundedMemoryManager) returnBufferToPool(buffer []byte) bool {
	originalSize := cap(buffer)
	
	// Check if this buffer came from a pool
	if pool, exists := m.bufferPools[originalSize]; exists {
		// Reset buffer before returning to pool
		for i := range buffer {
			buffer[i] = 0
		}
		pool.Put(buffer[:originalSize]) // Return full capacity buffer
		return true
	}
	
	return false
}

func (m *BoundedMemoryManager) optimizeGC() {
	if m.gcOptimizer == nil {
		return
	}
	
	// Set GC target percentage using debug.SetGCPercent
	debug.SetGCPercent(m.config.GCTargetPercent)
	
	log.Debug().
		Int("gc_target_percent", m.config.GCTargetPercent).
		Msg("GC optimization applied")
}

func (m *BoundedMemoryManager) monitoringRoutine() {
	ticker := time.NewTicker(m.config.MonitoringInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.performMonitoring()
		case alert := <-m.alerts:
			m.handleAlert(alert)
		case <-m.stopCh:
			return
		}
	}
}

func (m *BoundedMemoryManager) performMonitoring() {
	current := atomic.LoadInt64(&m.currentUsage)
	percentage := float64(current) / float64(m.config.MaxMemoryUsage)
	
	// Notify monitors
	m.mu.RLock()
	for _, monitor := range m.monitors {
		monitor.OnMemoryUpdate(current, percentage)
	}
	m.mu.RUnlock()
	
	// Log periodic stats
	log.Debug().
		Int64("current_usage", current).
		Float64("percentage", percentage*100).
		Int64("allocations", atomic.LoadInt64(&m.allocations)).
		Int64("deallocations", atomic.LoadInt64(&m.deallocations)).
		Msg("Memory monitoring update")
}

func (m *BoundedMemoryManager) handleAlert(alert MemoryAlert) {
	// Notify monitors
	m.mu.RLock()
	for _, monitor := range m.monitors {
		monitor.OnMemoryAlert(alert)
	}
	m.mu.RUnlock()
	
	// Log alert
	log.Warn().
		Str("level", fmt.Sprintf("%d", alert.Level)).
		Str("message", alert.Message).
		Float64("percentage", alert.Percentage*100).
		Msg("Memory alert triggered")
}

func (m *BoundedMemoryManager) getGCStats() *GCStats {
	m.gcStats.mu.RLock()
	defer m.gcStats.mu.RUnlock()
	
	statsCopy := *m.gcStats
	return &statsCopy
}

// Validate validates memory configuration
func (cfg *MemoryConfig) Validate() error {
	if cfg.MaxMemoryUsage <= 0 {
		return fmt.Errorf("max_memory_usage must be positive")
	}
	
	if cfg.WarningThreshold <= 0 || cfg.WarningThreshold > 1 {
		return fmt.Errorf("warning_threshold must be between 0 and 1")
	}
	
	if cfg.CriticalThreshold <= cfg.WarningThreshold || cfg.CriticalThreshold > 1 {
		return fmt.Errorf("critical_threshold must be between warning_threshold and 1")
	}
	
	if cfg.MaxRequestSize <= 0 {
		return fmt.Errorf("max_request_size must be positive")
	}
	
	if cfg.MaxResponseSize <= 0 {
		return fmt.Errorf("max_response_size must be positive")
	}
	
	return nil
}