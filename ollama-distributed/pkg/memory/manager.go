package memory

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Manager handles memory management across the distributed system
type Manager struct {
	config *Config

	// Memory monitoring
	monitor   *Monitor
	collector *SimpleGCOptimizer

	// Cache management
	caches   map[string]Cache
	cachesMu sync.RWMutex

	// Memory pools
	pools   map[string]*Pool
	poolsMu sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Config holds memory management configuration
type Config struct {
	// Memory limits
	MaxMemoryMB         int64 `yaml:"max_memory_mb"`
	WarningThresholdMB  int64 `yaml:"warning_threshold_mb"`
	CriticalThresholdMB int64 `yaml:"critical_threshold_mb"`

	// GC optimization
	GCTargetPercent int           `yaml:"gc_target_percent"`
	GCInterval      time.Duration `yaml:"gc_interval"`

	// Cache settings
	DefaultCacheSize int           `yaml:"default_cache_size"`
	CacheTTL         time.Duration `yaml:"cache_ttl"`

	// Pool settings
	PoolCleanupInterval time.Duration `yaml:"pool_cleanup_interval"`

	// Monitoring
	MonitorInterval time.Duration `yaml:"monitor_interval"`
	EnableProfiling bool          `yaml:"enable_profiling"`
}

// DefaultConfig returns default memory management configuration
func DefaultConfig() *Config {
	return &Config{
		MaxMemoryMB:         8192, // 8GB
		WarningThresholdMB:  6144, // 6GB
		CriticalThresholdMB: 7168, // 7GB
		GCTargetPercent:     100,
		GCInterval:          30 * time.Second,
		DefaultCacheSize:    1000,
		CacheTTL:            5 * time.Minute,
		PoolCleanupInterval: 1 * time.Minute,
		MonitorInterval:     10 * time.Second,
		EnableProfiling:     false,
	}
}

// NewManager creates a new memory manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:    config,
		monitor:   NewMonitor(config),
		collector: NewSimpleGCOptimizer(config),
		caches:    make(map[string]Cache),
		pools:     make(map[string]*Pool),
		ctx:       ctx,
		cancel:    cancel,
	}

	return manager
}

// Start starts the memory manager
func (m *Manager) Start() error {
	// Start memory monitoring
	m.wg.Add(1)
	go m.runMonitoring()

	// Start GC optimization
	m.wg.Add(1)
	go m.runGCOptimization()

	// Start cache cleanup
	m.wg.Add(1)
	go m.runCacheCleanup()

	// Start pool cleanup
	m.wg.Add(1)
	go m.runPoolCleanup()

	return nil
}

// Stop stops the memory manager
func (m *Manager) Stop() error {
	m.cancel()
	m.wg.Wait()

	// Clear all caches
	m.cachesMu.Lock()
	for name, cache := range m.caches {
		cache.Clear()
		delete(m.caches, name)
	}
	m.cachesMu.Unlock()

	// Clear all pools
	m.poolsMu.Lock()
	for name, pool := range m.pools {
		pool.Clear()
		delete(m.pools, name)
	}
	m.poolsMu.Unlock()

	return nil
}

// GetCache returns or creates a cache with the given name
func (m *Manager) GetCache(name string) Cache {
	m.cachesMu.RLock()
	if cache, exists := m.caches[name]; exists {
		m.cachesMu.RUnlock()
		return cache
	}
	m.cachesMu.RUnlock()

	m.cachesMu.Lock()
	defer m.cachesMu.Unlock()

	// Double-check after acquiring write lock
	if cache, exists := m.caches[name]; exists {
		return cache
	}

	// Create new LRU cache
	cache := NewLRUCache(m.config.DefaultCacheSize, m.config.CacheTTL)
	m.caches[name] = cache

	return cache
}

// GetPool returns or creates a memory pool with the given name
func (m *Manager) GetPool(name string, itemSize int) *Pool {
	m.poolsMu.RLock()
	if pool, exists := m.pools[name]; exists {
		m.poolsMu.RUnlock()
		return pool
	}
	m.poolsMu.RUnlock()

	m.poolsMu.Lock()
	defer m.poolsMu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := m.pools[name]; exists {
		return pool
	}

	// Create new memory pool
	pool := NewPool(itemSize)
	m.pools[name] = pool

	return pool
}

// GetMemoryStats returns current memory statistics
func (m *Manager) GetMemoryStats() *Stats {
	return m.monitor.GetStats()
}

// ForceGC forces garbage collection
func (m *Manager) ForceGC() {
	runtime.GC()
	runtime.GC() // Run twice for better cleanup
}

// GetCacheStats returns statistics for all caches
func (m *Manager) GetCacheStats() map[string]CacheStats {
	m.cachesMu.RLock()
	defer m.cachesMu.RUnlock()

	stats := make(map[string]CacheStats)
	for name, cache := range m.caches {
		stats[name] = cache.Stats()
	}

	return stats
}

// GetPoolStats returns statistics for all pools
func (m *Manager) GetPoolStats() map[string]PoolStats {
	m.poolsMu.RLock()
	defer m.poolsMu.RUnlock()

	stats := make(map[string]PoolStats)
	for name, pool := range m.pools {
		stats[name] = pool.Stats()
	}

	return stats
}

// runMonitoring runs the memory monitoring loop
func (m *Manager) runMonitoring() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.monitor.Update()

			// Check memory thresholds
			stats := m.monitor.GetStats()
			if stats.UsedMB > m.config.CriticalThresholdMB {
				m.handleCriticalMemory()
			} else if stats.UsedMB > m.config.WarningThresholdMB {
				m.handleWarningMemory()
			}
		}
	}
}

// runGCOptimization runs the GC optimization loop
func (m *Manager) runGCOptimization() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.GCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collector.Optimize()
		}
	}
}

// runCacheCleanup runs the cache cleanup loop
func (m *Manager) runCacheCleanup() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.CacheTTL / 2) // Cleanup twice per TTL
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupCaches()
		}
	}
}

// runPoolCleanup runs the pool cleanup loop
func (m *Manager) runPoolCleanup() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.PoolCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupPools()
		}
	}
}

// handleCriticalMemory handles critical memory situations
func (m *Manager) handleCriticalMemory() {
	fmt.Printf("CRITICAL: Memory usage above threshold, forcing cleanup\n")

	// Aggressive cache cleanup
	m.cachesMu.Lock()
	for _, cache := range m.caches {
		cache.Clear()
	}
	m.cachesMu.Unlock()

	// Force GC
	m.ForceGC()

	// TODO: Implement additional emergency measures
	// - Reject new requests
	// - Scale down operations
	// - Alert monitoring systems
}

// handleWarningMemory handles warning memory situations
func (m *Manager) handleWarningMemory() {
	fmt.Printf("WARNING: Memory usage above warning threshold\n")

	// Gentle cache cleanup
	m.cleanupCaches()

	// Trigger GC
	runtime.GC()
}

// cleanupCaches cleans up expired cache entries
func (m *Manager) cleanupCaches() {
	m.cachesMu.RLock()
	defer m.cachesMu.RUnlock()

	for _, cache := range m.caches {
		cache.Cleanup()
	}
}

// cleanupPools cleans up unused pool items
func (m *Manager) cleanupPools() {
	m.poolsMu.RLock()
	defer m.poolsMu.RUnlock()

	for _, pool := range m.pools {
		pool.Cleanup()
	}
}
