package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AlgorithmCache provides intelligent caching for frequently used algorithms
type AlgorithmCache struct {
	// In-memory cache for hot data
	memory   map[string]*CacheEntry
	memoryMu sync.RWMutex

	// Configuration
	config *CacheConfig

	// Statistics
	stats *CacheStats

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	stopCh chan struct{}
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	// Memory cache settings
	MaxMemoryEntries int           `json:"max_memory_entries"`
	MemoryTTL        time.Duration `json:"memory_ttl"`

	// Cache policies
	EvictionPolicy string `json:"eviction_policy"` // LRU, LFU, TTL
	MaxKeySize     int    `json:"max_key_size"`
	MaxValueSize   int    `json:"max_value_size"`

	// Performance settings
	EnableCompression bool          `json:"enable_compression"`
	EnableMetrics     bool          `json:"enable_metrics"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`

	// Redis settings (optional external cache)
	RedisEnabled  bool   `json:"redis_enabled"`
	RedisAddr     string `json:"redis_addr"`
	RedisPassword string `json:"redis_password"`
	RedisDB       int    `json:"redis_db"`
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key            string                 `json:"key"`
	Value          interface{}            `json:"value"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessedAt time.Time              `json:"last_accessed_at"`
	AccessCount    int64                  `json:"access_count"`
	TTL            time.Duration          `json:"ttl"`
	Compressed     bool                   `json:"compressed"`
	Size           int64                  `json:"size"`
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	Evictions      int64         `json:"evictions"`
	Errors         int64         `json:"errors"`
	TotalRequests  int64         `json:"total_requests"`
	AverageLatency time.Duration `json:"average_latency"`
	MemoryUsage    int64         `json:"memory_usage"`
	LastResetTime  time.Time     `json:"last_reset_time"`
	mu             sync.RWMutex
}

// NewAlgorithmCache creates a new algorithm cache instance
func NewAlgorithmCache(config *CacheConfig) (*AlgorithmCache, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &AlgorithmCache{
		memory: make(map[string]*CacheEntry),
		config: config,
		stats:  &CacheStats{LastResetTime: time.Now()},
		ctx:    ctx,
		cancel: cancel,
		stopCh: make(chan struct{}),
	}

	// Start background cleanup routine
	go cache.cleanupRoutine()

	log.Info().
		Int("max_entries", config.MaxMemoryEntries).
		Dur("ttl", config.MemoryTTL).
		Str("eviction_policy", config.EvictionPolicy).
		Msg("Algorithm cache initialized")

	return cache, nil
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxMemoryEntries:  10000,
		MemoryTTL:         30 * time.Minute,
		EvictionPolicy:    "LRU",
		MaxKeySize:        256,
		MaxValueSize:      1024 * 1024, // 1MB
		EnableCompression: true,
		EnableMetrics:     true,
		CleanupInterval:   5 * time.Minute,
		RedisEnabled:      false,
		RedisAddr:         os.Getenv("REDIS_ADDR"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           0,
	}
}

// Get retrieves a value from cache
func (c *AlgorithmCache) Get(key string) (interface{}, bool, error) {
	start := time.Now()
	defer func() {
		c.updateLatency(time.Since(start))
	}()

	c.stats.mu.Lock()
	c.stats.TotalRequests++
	c.stats.mu.Unlock()

	// Check memory cache first
	c.memoryMu.RLock()
	entry, exists := c.memory[key]
	c.memoryMu.RUnlock()

	if exists {
		// Check if entry has expired
		if c.isExpired(entry) {
			c.Delete(key)
			c.recordMiss()
			return nil, false, nil
		}

		// Update access statistics
		c.updateEntryAccess(entry)
		c.recordHit()

		return entry.Value, true, nil
	}

	c.recordMiss()
	return nil, false, nil
}

// Set stores a value in cache
func (c *AlgorithmCache) Set(key string, value interface{}, ttl time.Duration) error {
	if len(key) > c.config.MaxKeySize {
		return fmt.Errorf("key size exceeds maximum: %d > %d", len(key), c.config.MaxKeySize)
	}

	// Serialize value for size calculation
	valueBytes, err := json.Marshal(value)
	if err != nil {
		c.recordError()
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	if len(valueBytes) > c.config.MaxValueSize {
		return fmt.Errorf("value size exceeds maximum: %d > %d", len(valueBytes), c.config.MaxValueSize)
	}

	// Create cache entry
	entry := &CacheEntry{
		Key:            key,
		Value:          value,
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
		LastAccessedAt: time.Now(),
		AccessCount:    1,
		TTL:            ttl,
		Size:           int64(len(valueBytes)),
	}

	// Check if we need to evict entries
	c.memoryMu.Lock()
	if len(c.memory) >= c.config.MaxMemoryEntries {
		if err := c.evictEntry(); err != nil {
			c.memoryMu.Unlock()
			c.recordError()
			return fmt.Errorf("failed to evict entry: %w", err)
		}
	}

	c.memory[key] = entry
	c.memoryMu.Unlock()

	log.Debug().
		Str("key", key).
		Int64("size", entry.Size).
		Dur("ttl", ttl).
		Msg("Cache entry stored")

	return nil
}

// Delete removes a value from cache
func (c *AlgorithmCache) Delete(key string) error {
	c.memoryMu.Lock()
	delete(c.memory, key)
	c.memoryMu.Unlock()

	log.Debug().Str("key", key).Msg("Cache entry deleted")
	return nil
}

// Clear removes all entries from cache
func (c *AlgorithmCache) Clear() error {
	c.memoryMu.Lock()
	c.memory = make(map[string]*CacheEntry)
	c.memoryMu.Unlock()

	log.Info().Msg("Cache cleared")
	return nil
}

// GetStats returns cache statistics
func (c *AlgorithmCache) GetStats() *CacheStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()

	// Calculate memory usage
	c.memoryMu.RLock()
	var memUsage int64
	for _, entry := range c.memory {
		memUsage += entry.Size
	}
	c.memoryMu.RUnlock()

	statsCopy := *c.stats
	statsCopy.MemoryUsage = memUsage

	return &statsCopy
}

// Close gracefully shuts down the cache
func (c *AlgorithmCache) Close() error {
	c.cancel()
	close(c.stopCh)

	log.Info().Msg("Algorithm cache closed")
	return nil
}

// Helper methods

func (c *AlgorithmCache) isExpired(entry *CacheEntry) bool {
	if entry.TTL == 0 {
		return false // No expiration
	}
	return time.Since(entry.CreatedAt) > entry.TTL
}

func (c *AlgorithmCache) updateEntryAccess(entry *CacheEntry) {
	entry.LastAccessedAt = time.Now()
	entry.AccessCount++
}

func (c *AlgorithmCache) evictEntry() error {
	switch c.config.EvictionPolicy {
	case "LRU":
		return c.evictLRU()
	case "LFU":
		return c.evictLFU()
	case "TTL":
		return c.evictTTL()
	default:
		return c.evictLRU()
	}
}

func (c *AlgorithmCache) evictLRU() error {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, entry := range c.memory {
		if entry.LastAccessedAt.Before(oldestTime) {
			oldestTime = entry.LastAccessedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.memory, oldestKey)
		c.recordEviction()
		log.Debug().Str("key", oldestKey).Msg("LRU eviction")
	}

	return nil
}

func (c *AlgorithmCache) evictLFU() error {
	var leastUsedKey string
	var leastCount int64 = -1

	for key, entry := range c.memory {
		if leastCount == -1 || entry.AccessCount < leastCount {
			leastCount = entry.AccessCount
			leastUsedKey = key
		}
	}

	if leastUsedKey != "" {
		delete(c.memory, leastUsedKey)
		c.recordEviction()
		log.Debug().Str("key", leastUsedKey).Msg("LFU eviction")
	}

	return nil
}

func (c *AlgorithmCache) evictTTL() error {
	start := time.Now()
	for key, entry := range c.memory {
		if c.isExpired(entry) {
			delete(c.memory, key)
			c.recordEviction()
			log.Debug().
				Str("key", key).
				Dur("eviction_time", time.Since(start)).
				Msg("TTL eviction")
			return nil
		}
	}

	// If no expired entries, fall back to LRU
	return c.evictLRU()
}

func (c *AlgorithmCache) cleanupRoutine() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

func (c *AlgorithmCache) cleanup() {
	start := time.Now()
	c.memoryMu.Lock()
	defer c.memoryMu.Unlock()

	expiredKeys := make([]string, 0)

	for key, entry := range c.memory {
		if c.isExpired(entry) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.memory, key)
		c.recordEviction()
	}

	if len(expiredKeys) > 0 {
		log.Debug().
			Int("expired_count", len(expiredKeys)).
			Dur("cleanup_time", time.Since(start)).
			Msg("Cache cleanup completed")
	}
}

// Statistics helpers

func (c *AlgorithmCache) recordHit() {
	c.stats.mu.Lock()
	c.stats.Hits++
	c.stats.mu.Unlock()
}

func (c *AlgorithmCache) recordMiss() {
	c.stats.mu.Lock()
	c.stats.Misses++
	c.stats.mu.Unlock()
}

func (c *AlgorithmCache) recordEviction() {
	c.stats.mu.Lock()
	c.stats.Evictions++
	c.stats.mu.Unlock()
}

func (c *AlgorithmCache) recordError() {
	c.stats.mu.Lock()
	c.stats.Errors++
	c.stats.mu.Unlock()
}

func (c *AlgorithmCache) updateLatency(duration time.Duration) {
	c.stats.mu.Lock()
	// Simple moving average
	if c.stats.TotalRequests > 0 {
		c.stats.AverageLatency = (c.stats.AverageLatency*time.Duration(c.stats.TotalRequests-1) + duration) / time.Duration(c.stats.TotalRequests)
	} else {
		c.stats.AverageLatency = duration
	}
	c.stats.mu.Unlock()
}

// Validate validates cache configuration
func (cfg *CacheConfig) Validate() error {
	if cfg.MaxMemoryEntries <= 0 {
		return fmt.Errorf("max_memory_entries must be positive")
	}

	if cfg.MaxKeySize <= 0 {
		return fmt.Errorf("max_key_size must be positive")
	}

	if cfg.MaxValueSize <= 0 {
		return fmt.Errorf("max_value_size must be positive")
	}

	validPolicies := map[string]bool{"LRU": true, "LFU": true, "TTL": true}
	if !validPolicies[cfg.EvictionPolicy] {
		return fmt.Errorf("invalid eviction_policy: %s", cfg.EvictionPolicy)
	}

	return nil
}
