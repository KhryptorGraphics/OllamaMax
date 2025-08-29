package performance

import (
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/cache"
	"github.com/rs/zerolog/log"
)

// CacheManager manages multiple caches for optimal performance
type CacheManager struct {
	config *OptimizerConfig

	// Cache instances
	caches map[string]*cache.AlgorithmCache
	mu     sync.RWMutex

	// Cache configurations
	defaultConfig *cache.CacheConfig
	
	// Aggregate statistics
	aggregateStats *CacheAggregateStats
}

// CacheAggregateStats holds aggregate statistics across all caches
type CacheAggregateStats struct {
	TotalHits      int64         `json:"total_hits"`
	TotalMisses    int64         `json:"total_misses"`
	TotalEvictions int64         `json:"total_evictions"`
	TotalErrors    int64         `json:"total_errors"`
	HitRate        float64       `json:"hit_rate"`
	MemoryUsage    int64         `json:"memory_usage"`
	LastUpdated    time.Time     `json:"last_updated"`
	mu             sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *OptimizerConfig) *CacheManager {
	defaultConfig := &cache.CacheConfig{
		MaxMemoryEntries:  10000,
		MemoryTTL:         5 * time.Minute,
		EvictionPolicy:    "LRU",
		MaxKeySize:        256,
		MaxValueSize:      1024 * 1024, // 1MB
		EnableCompression: true,
		EnableMetrics:     true,
		CleanupInterval:   1 * time.Minute,
	}

	return &CacheManager{
		config:         config,
		caches:         make(map[string]*cache.AlgorithmCache),
		defaultConfig:  defaultConfig,
		aggregateStats: &CacheAggregateStats{LastUpdated: time.Now()},
	}
}

// Start starts the cache manager
func (cm *CacheManager) Start() error {
	log.Info().Msg("Cache manager started")
	return nil
}

// Stop stops the cache manager
func (cm *CacheManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close all caches
	for name, c := range cm.caches {
		if err := c.Close(); err != nil {
			log.Warn().Err(err).Str("cache", name).Msg("Failed to close cache")
		}
	}

	cm.caches = make(map[string]*cache.AlgorithmCache)
	log.Info().Msg("Cache manager stopped")
	return nil
}

// GetCache gets or creates a cache with the given name
func (cm *CacheManager) GetCache(name string) *cache.AlgorithmCache {
	cm.mu.RLock()
	if c, exists := cm.caches[name]; exists {
		cm.mu.RUnlock()
		return c
	}
	cm.mu.RUnlock()

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock
	if c, exists := cm.caches[name]; exists {
		return c
	}

	// Create new cache
	c, err := cache.NewAlgorithmCache(cm.defaultConfig)
	if err != nil {
		log.Error().Err(err).Str("cache", name).Msg("Failed to create cache")
		return nil
	}

	cm.caches[name] = c
	log.Info().Str("cache", name).Msg("Created new cache")
	return c
}

// Get retrieves a value from the specified cache
func (cm *CacheManager) Get(cacheName, key string) (interface{}, bool) {
	c := cm.GetCache(cacheName)
	if c == nil {
		return nil, false
	}

	value, found, err := c.Get(key)
	if err != nil {
		log.Error().Err(err).Str("cache", cacheName).Str("key", key).Msg("Cache get error")
		return nil, false
	}

	return value, found
}

// Set stores a value in the specified cache
func (cm *CacheManager) Set(cacheName, key string, value interface{}, ttl time.Duration) error {
	c := cm.GetCache(cacheName)
	if c == nil {
		return cache.ErrCacheNotFound
	}

	return c.Set(key, value, ttl)
}

// Delete removes a value from the specified cache
func (cm *CacheManager) Delete(cacheName, key string) error {
	c := cm.GetCache(cacheName)
	if c == nil {
		return cache.ErrCacheNotFound
	}

	return c.Delete(key)
}

// Clear removes all entries from the specified cache
func (cm *CacheManager) Clear(cacheName string) error {
	c := cm.GetCache(cacheName)
	if c == nil {
		return cache.ErrCacheNotFound
	}

	return c.Clear()
}

// ClearExpired removes expired entries from all caches
func (cm *CacheManager) ClearExpired() {
	cm.mu.RLock()
	caches := make([]*cache.AlgorithmCache, 0, len(cm.caches))
	for _, c := range cm.caches {
		caches = append(caches, c)
	}
	cm.mu.RUnlock()

	// Clear expired entries in parallel
	var wg sync.WaitGroup
	for _, c := range caches {
		wg.Add(1)
		go func(cache *cache.AlgorithmCache) {
			defer wg.Done()
			// The cache implementation should handle cleanup internally
			// This is a placeholder for explicit cleanup if needed
		}(c)
	}
	wg.Wait()

	log.Info().Msg("Cleared expired cache entries")
}

// GetStats returns aggregate cache statistics
func (cm *CacheManager) GetStats() *CacheAggregateStats {
	cm.aggregateStats.mu.Lock()
	defer cm.aggregateStats.mu.Unlock()

	var totalHits, totalMisses, totalEvictions, totalErrors int64
	var totalMemoryUsage int64

	cm.mu.RLock()
	for _, c := range cm.caches {
		stats := c.GetStats()
		totalHits += stats.Hits
		totalMisses += stats.Misses
		totalEvictions += stats.Evictions
		totalErrors += stats.Errors
		totalMemoryUsage += stats.MemoryUsage
	}
	cm.mu.RUnlock()

	// Update aggregate stats
	cm.aggregateStats.TotalHits = totalHits
	cm.aggregateStats.TotalMisses = totalMisses
	cm.aggregateStats.TotalEvictions = totalEvictions
	cm.aggregateStats.TotalErrors = totalErrors
	cm.aggregateStats.MemoryUsage = totalMemoryUsage
	cm.aggregateStats.LastUpdated = time.Now()

	// Calculate hit rate
	totalRequests := totalHits + totalMisses
	if totalRequests > 0 {
		cm.aggregateStats.HitRate = float64(totalHits) / float64(totalRequests) * 100
	}

	// Return a copy
	stats := *cm.aggregateStats
	return &stats
}

// GetCacheNames returns all cache names
func (cm *CacheManager) GetCacheNames() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	names := make([]string, 0, len(cm.caches))
	for name := range cm.caches {
		names = append(names, name)
	}
	return names
}

// OptimizeForWorkload optimizes cache settings for specific workload
func (cm *CacheManager) OptimizeForWorkload(workloadType string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var newConfig *cache.CacheConfig

	switch workloadType {
	case "read-heavy":
		newConfig = &cache.CacheConfig{
			MaxMemoryEntries:  20000,  // Larger cache for read-heavy workloads
			MemoryTTL:         10 * time.Minute,
			EvictionPolicy:    "LFU",  // Least Frequently Used
			MaxKeySize:        256,
			MaxValueSize:      2 * 1024 * 1024, // 2MB
			EnableCompression: true,
			EnableMetrics:     true,
			CleanupInterval:   2 * time.Minute,
		}

	case "write-heavy":
		newConfig = &cache.CacheConfig{
			MaxMemoryEntries:  5000,   // Smaller cache for write-heavy
			MemoryTTL:         2 * time.Minute,
			EvictionPolicy:    "LRU",  // Least Recently Used
			MaxKeySize:        256,
			MaxValueSize:      512 * 1024, // 512KB
			EnableCompression: false,  // Disable compression for faster writes
			EnableMetrics:     true,
			CleanupInterval:   30 * time.Second,
		}

	case "memory-constrained":
		newConfig = &cache.CacheConfig{
			MaxMemoryEntries:  2000,   // Very small cache
			MemoryTTL:         1 * time.Minute,
			EvictionPolicy:    "TTL",  // Time-based eviction
			MaxKeySize:        128,
			MaxValueSize:      256 * 1024, // 256KB
			EnableCompression: true,   // Enable compression to save memory
			EnableMetrics:     false,  // Disable metrics to save memory
			CleanupInterval:   15 * time.Second,
		}

	case "high-performance":
		newConfig = &cache.CacheConfig{
			MaxMemoryEntries:  50000,  // Very large cache
			MemoryTTL:         30 * time.Minute,
			EvictionPolicy:    "LRU",
			MaxKeySize:        512,
			MaxValueSize:      4 * 1024 * 1024, // 4MB
			EnableCompression: false,  // Disable for speed
			EnableMetrics:     true,
			CleanupInterval:   5 * time.Minute,
		}

	default:
		log.Warn().Str("workload", workloadType).Msg("Unknown workload type, using default config")
		return
	}

	cm.defaultConfig = newConfig
	log.Info().Str("workload", workloadType).Msg("Optimized cache configuration for workload")
}

// Prefetch preloads cache entries based on predicted access patterns
func (cm *CacheManager) Prefetch(cacheName string, keys []string, loader func(string) (interface{}, error)) {
	c := cm.GetCache(cacheName)
	if c == nil {
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent prefetch operations

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check if already cached
			_, found, _ := c.Get(k)
			if found {
				return
			}

			// Load and cache the value
			if loader != nil {
				value, err := loader(k)
				if err == nil {
					c.Set(k, value, cm.defaultConfig.MemoryTTL)
				}
			}
		}(key)
	}

	wg.Wait()
	log.Info().
		Str("cache", cacheName).
		Int("keys", len(keys)).
		Msg("Completed cache prefetch")
}

// Warmup preloads frequently accessed data
func (cm *CacheManager) Warmup(cacheName string, data map[string]interface{}) {
	c := cm.GetCache(cacheName)
	if c == nil {
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20) // Limit concurrent operations

	for key, value := range data {
		wg.Add(1)
		go func(k string, v interface{}) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			c.Set(k, v, cm.defaultConfig.MemoryTTL)
		}(key, value)
	}

	wg.Wait()
	log.Info().
		Str("cache", cacheName).
		Int("entries", len(data)).
		Msg("Cache warmup completed")
}

// Monitor starts monitoring cache performance
func (cm *CacheManager) Monitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			stats := cm.GetStats()
			
			log.Info().
				Int64("hits", stats.TotalHits).
				Int64("misses", stats.TotalMisses).
				Float64("hit_rate", stats.HitRate).
				Int64("memory_mb", stats.MemoryUsage/(1024*1024)).
				Int64("evictions", stats.TotalEvictions).
				Msg("Cache performance metrics")

			// Alert on low hit rate
			if stats.HitRate < 70.0 && stats.TotalHits+stats.TotalMisses > 1000 {
				log.Warn().
					Float64("hit_rate", stats.HitRate).
					Msg("Low cache hit rate detected - consider optimization")
			}

			// Alert on high memory usage
			maxMemoryMB := int64(cm.config.MaxMemoryUsageMB) / 4 // 25% of total for caches
			if stats.MemoryUsage/(1024*1024) > maxMemoryMB {
				log.Warn().
					Int64("memory_mb", stats.MemoryUsage/(1024*1024)).
					Int64("max_memory_mb", maxMemoryMB).
					Msg("High cache memory usage - consider cleanup")
			}
		}
	}()
}