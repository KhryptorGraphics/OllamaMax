package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig configures the rate limiter
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
	WindowSize        time.Duration
	CleanupInterval   time.Duration
}

// RateLimitMetrics tracks rate limiting performance
type RateLimitMetrics struct {
	RequestsAllowed int64     `json:"requests_allowed"`
	RequestsBlocked int64     `json:"requests_blocked"`
	ActiveBuckets   int64     `json:"active_buckets"`
	TotalRequests   int64     `json:"total_requests"`
	BlockRate       float64   `json:"block_rate"`
	LastUpdated     time.Time `json:"last_updated"`
	mu              sync.RWMutex
}

// TokenBucket implements a token bucket for rate limiting
type TokenBucket struct {
	capacity     int
	tokens       int
	refillRate   int
	lastRefill   time.Time
	mu           sync.Mutex
}

// BucketInfo represents information about a rate limit bucket
type BucketInfo struct {
	Key         string    `json:"key"`
	Capacity    int       `json:"capacity"`
	Tokens      int       `json:"tokens"`
	RefillRate  int       `json:"refill_rate"`
	LastRefill  time.Time `json:"last_refill"`
	LastAccess  time.Time `json:"last_access"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimitConfig) (*RateLimiter, error) {
	if config == nil {
		config = &RateLimitConfig{
			RequestsPerSecond: 100,
			BurstSize:         200,
			WindowSize:        time.Minute,
			CleanupInterval:   5 * time.Minute,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	limiter := &RateLimiter{
		config:  config,
		buckets: make(map[string]*TokenBucket),
		metrics: &RateLimitMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return limiter, nil
}

// Start starts the rate limiter
func (rl *RateLimiter) Start() error {
	// Start cleanup routine
	rl.wg.Add(1)
	go rl.cleanupLoop()
	
	// Start metrics collection
	rl.wg.Add(1)
	go rl.metricsLoop()
	
	return nil
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() error {
	rl.cancel()
	rl.wg.Wait()
	return nil
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	bucket := rl.getBucket(key)
	
	rl.metrics.mu.Lock()
	rl.metrics.TotalRequests++
	rl.metrics.mu.Unlock()
	
	if bucket.consume() {
		rl.metrics.mu.Lock()
		rl.metrics.RequestsAllowed++
		rl.metrics.LastUpdated = time.Now()
		rl.metrics.mu.Unlock()
		return true
	}
	
	rl.metrics.mu.Lock()
	rl.metrics.RequestsBlocked++
	rl.metrics.LastUpdated = time.Now()
	rl.metrics.mu.Unlock()
	
	return false
}

// getBucket gets or creates a token bucket for a key
func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	rl.bucketsMu.RLock()
	bucket, exists := rl.buckets[key]
	rl.bucketsMu.RUnlock()
	
	if exists {
		return bucket
	}
	
	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()
	
	// Double-check after acquiring write lock
	if bucket, exists := rl.buckets[key]; exists {
		return bucket
	}
	
	// Create new bucket
	bucket = &TokenBucket{
		capacity:   rl.config.BurstSize,
		tokens:     rl.config.BurstSize,
		refillRate: rl.config.RequestsPerSecond,
		lastRefill: time.Now(),
	}
	
	rl.buckets[key] = bucket
	return bucket
}

// GetBucketInfo returns information about a bucket
func (rl *RateLimiter) GetBucketInfo(key string) (*BucketInfo, bool) {
	rl.bucketsMu.RLock()
	bucket, exists := rl.buckets[key]
	rl.bucketsMu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	
	return &BucketInfo{
		Key:        key,
		Capacity:   bucket.capacity,
		Tokens:     bucket.tokens,
		RefillRate: bucket.refillRate,
		LastRefill: bucket.lastRefill,
		LastAccess: time.Now(),
	}, true
}

// GetAllBuckets returns information about all buckets
func (rl *RateLimiter) GetAllBuckets() []*BucketInfo {
	rl.bucketsMu.RLock()
	defer rl.bucketsMu.RUnlock()
	
	buckets := make([]*BucketInfo, 0, len(rl.buckets))
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		info := &BucketInfo{
			Key:        key,
			Capacity:   bucket.capacity,
			Tokens:     bucket.tokens,
			RefillRate: bucket.refillRate,
			LastRefill: bucket.lastRefill,
			LastAccess: time.Now(),
		}
		bucket.mu.Unlock()
		buckets = append(buckets, info)
	}
	
	return buckets
}

// GetMetrics returns rate limiting metrics
func (rl *RateLimiter) GetMetrics() *RateLimitMetrics {
	rl.metrics.mu.RLock()
	defer rl.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *rl.metrics
	return &metrics
}

// Middleware returns a Gin middleware for rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return rl.MiddlewareWithKeyFunc(func(c *gin.Context) string {
		// Default key function uses client IP
		return c.ClientIP()
	})
}

// MiddlewareWithKeyFunc returns a Gin middleware with custom key function
func (rl *RateLimiter) MiddlewareWithKeyFunc(keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		
		if !rl.Allow(key) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "too many requests",
			})
			c.Abort()
			return
		}
		
		// Add rate limit headers
		bucket := rl.getBucket(key)
		bucket.mu.Lock()
		remaining := bucket.tokens
		bucket.mu.Unlock()
		
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
		
		c.Next()
	}
}

// cleanupLoop cleans up old buckets
func (rl *RateLimiter) cleanupLoop() {
	defer rl.wg.Done()
	
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.cleanupOldBuckets()
		}
	}
}

// cleanupOldBuckets removes old unused buckets
func (rl *RateLimiter) cleanupOldBuckets() {
	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()
	
	cutoff := time.Now().Add(-rl.config.WindowSize)
	var keysToDelete []string
	
	for key, bucket := range rl.buckets {
		bucket.mu.Lock()
		if bucket.lastRefill.Before(cutoff) {
			keysToDelete = append(keysToDelete, key)
		}
		bucket.mu.Unlock()
	}
	
	for _, key := range keysToDelete {
		delete(rl.buckets, key)
	}
}

// metricsLoop runs the metrics collection loop
func (rl *RateLimiter) metricsLoop() {
	defer rl.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.updateMetrics()
		}
	}
}

// updateMetrics updates rate limiting metrics
func (rl *RateLimiter) updateMetrics() {
	rl.metrics.mu.Lock()
	defer rl.metrics.mu.Unlock()
	
	rl.bucketsMu.RLock()
	rl.metrics.ActiveBuckets = int64(len(rl.buckets))
	rl.bucketsMu.RUnlock()
	
	// Calculate block rate
	if rl.metrics.TotalRequests > 0 {
		rl.metrics.BlockRate = float64(rl.metrics.RequestsBlocked) / float64(rl.metrics.TotalRequests)
	}
	
	rl.metrics.LastUpdated = time.Now()
}

// TokenBucket methods

// consume tries to consume a token from the bucket
func (tb *TokenBucket) consume() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.refill()
	
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	
	return false
}

// refill refills the token bucket based on elapsed time
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	if elapsed <= 0 {
		return
	}
	
	// Calculate tokens to add
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// getTokens returns the current number of tokens
func (tb *TokenBucket) getTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	
	tb.refill()
	return tb.tokens
}

// Reset resets the token bucket
func (rl *RateLimiter) Reset() {
	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()
	
	rl.buckets = make(map[string]*TokenBucket)
	
	rl.metrics.mu.Lock()
	rl.metrics.RequestsAllowed = 0
	rl.metrics.RequestsBlocked = 0
	rl.metrics.TotalRequests = 0
	rl.metrics.ActiveBuckets = 0
	rl.metrics.BlockRate = 0
	rl.metrics.LastUpdated = time.Now()
	rl.metrics.mu.Unlock()
}

// UpdateConfig updates the rate limiter configuration
func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.config = config
	
	// Update existing buckets
	rl.bucketsMu.Lock()
	defer rl.bucketsMu.Unlock()
	
	for _, bucket := range rl.buckets {
		bucket.mu.Lock()
		bucket.capacity = config.BurstSize
		bucket.refillRate = config.RequestsPerSecond
		if bucket.tokens > bucket.capacity {
			bucket.tokens = bucket.capacity
		}
		bucket.mu.Unlock()
	}
}
