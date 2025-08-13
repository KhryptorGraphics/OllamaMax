package network

import (
	"context"
	"sync"
	"time"
)

// BandwidthManager manages network bandwidth allocation and throttling
type BandwidthManager struct {
	config *BandwidthConfig

	// Rate limiters
	inboundLimiter  *RateLimiter
	outboundLimiter *RateLimiter

	// Per-connection limiters
	connectionLimiters map[string]*RateLimiter
	mu                 sync.RWMutex

	// Statistics
	stats *BandwidthStats

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// BandwidthConfig holds bandwidth management configuration
type BandwidthConfig struct {
	// Global limits (bytes per second)
	MaxInboundBPS  int64 `yaml:"max_inbound_bps"`
	MaxOutboundBPS int64 `yaml:"max_outbound_bps"`

	// Per-connection limits
	MaxConnectionBPS int64 `yaml:"max_connection_bps"`

	// Burst settings
	InboundBurst  int64 `yaml:"inbound_burst"`
	OutboundBurst int64 `yaml:"outbound_burst"`

	// QoS settings
	EnableQoS       bool           `yaml:"enable_qos"`
	PriorityClasses map[string]int `yaml:"priority_classes"`
	DefaultPriority int            `yaml:"default_priority"`

	// Monitoring
	MonitorInterval time.Duration `yaml:"monitor_interval"`
	StatsRetention  time.Duration `yaml:"stats_retention"`
}

// DefaultBandwidthConfig returns default bandwidth configuration
func DefaultBandwidthConfig() *BandwidthConfig {
	return &BandwidthConfig{
		MaxInboundBPS:    100 * 1024 * 1024, // 100 MB/s
		MaxOutboundBPS:   100 * 1024 * 1024, // 100 MB/s
		MaxConnectionBPS: 10 * 1024 * 1024,  // 10 MB/s per connection
		InboundBurst:     10 * 1024 * 1024,  // 10 MB burst
		OutboundBurst:    10 * 1024 * 1024,  // 10 MB burst
		EnableQoS:        true,
		PriorityClasses: map[string]int{
			"critical": 1,
			"high":     2,
			"normal":   3,
			"low":      4,
		},
		DefaultPriority: 3,
		MonitorInterval: 1 * time.Second,
		StatsRetention:  1 * time.Hour,
	}
}

// BandwidthStats holds bandwidth usage statistics
type BandwidthStats struct {
	// Current usage
	InboundBPS  int64 `json:"inbound_bps"`
	OutboundBPS int64 `json:"outbound_bps"`

	// Total transferred
	TotalInbound  int64 `json:"total_inbound"`
	TotalOutbound int64 `json:"total_outbound"`

	// Peak usage
	PeakInboundBPS  int64 `json:"peak_inbound_bps"`
	PeakOutboundBPS int64 `json:"peak_outbound_bps"`

	// Connection statistics
	ActiveConnections int   `json:"active_connections"`
	ThrottledRequests int64 `json:"throttled_requests"`

	// QoS statistics
	HighPriorityBytes int64 `json:"high_priority_bytes"`
	LowPriorityBytes  int64 `json:"low_priority_bytes"`

	// Timestamps
	LastUpdate time.Time `json:"last_update"`
	StartTime  time.Time `json:"start_time"`
}

// NewBandwidthManager creates a new bandwidth manager
func NewBandwidthManager(config *BandwidthConfig) *BandwidthManager {
	if config == nil {
		config = DefaultBandwidthConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &BandwidthManager{
		config:             config,
		inboundLimiter:     NewRateLimiter(config.MaxInboundBPS, config.InboundBurst),
		outboundLimiter:    NewRateLimiter(config.MaxOutboundBPS, config.OutboundBurst),
		connectionLimiters: make(map[string]*RateLimiter),
		stats:              &BandwidthStats{StartTime: time.Now()},
		ctx:                ctx,
		cancel:             cancel,
	}

	return manager
}

// Start starts the bandwidth manager
func (bm *BandwidthManager) Start() error {
	// Start monitoring routine
	bm.wg.Add(1)
	go bm.runMonitoring()

	// Start cleanup routine
	bm.wg.Add(1)
	go bm.runCleanup()

	return nil
}

// Stop stops the bandwidth manager
func (bm *BandwidthManager) Stop() error {
	bm.cancel()
	bm.wg.Wait()
	return nil
}

// AllowInbound checks if inbound traffic is allowed
func (bm *BandwidthManager) AllowInbound(connectionID string, bytes int64) bool {
	// Check global inbound limit
	if !bm.inboundLimiter.Allow(bytes) {
		bm.updateStats(func(s *BandwidthStats) {
			s.ThrottledRequests++
		})
		return false
	}

	// Check per-connection limit
	if connectionID != "" {
		limiter := bm.getConnectionLimiter(connectionID)
		if !limiter.Allow(bytes) {
			bm.updateStats(func(s *BandwidthStats) {
				s.ThrottledRequests++
			})
			return false
		}
	}

	// Update statistics
	bm.updateStats(func(s *BandwidthStats) {
		s.TotalInbound += bytes
		s.LastUpdate = time.Now()
	})

	return true
}

// AllowOutbound checks if outbound traffic is allowed
func (bm *BandwidthManager) AllowOutbound(connectionID string, bytes int64) bool {
	// Check global outbound limit
	if !bm.outboundLimiter.Allow(bytes) {
		bm.updateStats(func(s *BandwidthStats) {
			s.ThrottledRequests++
		})
		return false
	}

	// Check per-connection limit
	if connectionID != "" {
		limiter := bm.getConnectionLimiter(connectionID)
		if !limiter.Allow(bytes) {
			bm.updateStats(func(s *BandwidthStats) {
				s.ThrottledRequests++
			})
			return false
		}
	}

	// Update statistics
	bm.updateStats(func(s *BandwidthStats) {
		s.TotalOutbound += bytes
		s.LastUpdate = time.Now()
	})

	return true
}

// WaitForInbound waits for inbound bandwidth to become available
func (bm *BandwidthManager) WaitForInbound(ctx context.Context, connectionID string, bytes int64) error {
	// Wait for global inbound capacity
	if err := bm.inboundLimiter.Wait(ctx, bytes); err != nil {
		return err
	}

	// Wait for per-connection capacity
	if connectionID != "" {
		limiter := bm.getConnectionLimiter(connectionID)
		if err := limiter.Wait(ctx, bytes); err != nil {
			return err
		}
	}

	// Update statistics
	bm.updateStats(func(s *BandwidthStats) {
		s.TotalInbound += bytes
		s.LastUpdate = time.Now()
	})

	return nil
}

// WaitForOutbound waits for outbound bandwidth to become available
func (bm *BandwidthManager) WaitForOutbound(ctx context.Context, connectionID string, bytes int64) error {
	// Wait for global outbound capacity
	if err := bm.outboundLimiter.Wait(ctx, bytes); err != nil {
		return err
	}

	// Wait for per-connection capacity
	if connectionID != "" {
		limiter := bm.getConnectionLimiter(connectionID)
		if err := limiter.Wait(ctx, bytes); err != nil {
			return err
		}
	}

	// Update statistics
	bm.updateStats(func(s *BandwidthStats) {
		s.TotalOutbound += bytes
		s.LastUpdate = time.Now()
	})

	return nil
}

// Stats returns current bandwidth statistics
func (bm *BandwidthManager) Stats() BandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	return *bm.stats
}

// SetInboundLimit updates the inbound bandwidth limit
func (bm *BandwidthManager) SetInboundLimit(bps int64) {
	bm.inboundLimiter.SetRate(bps)
	bm.config.MaxInboundBPS = bps
}

// SetOutboundLimit updates the outbound bandwidth limit
func (bm *BandwidthManager) SetOutboundLimit(bps int64) {
	bm.outboundLimiter.SetRate(bps)
	bm.config.MaxOutboundBPS = bps
}

// getConnectionLimiter returns or creates a rate limiter for a connection
func (bm *BandwidthManager) getConnectionLimiter(connectionID string) *RateLimiter {
	bm.mu.RLock()
	limiter, exists := bm.connectionLimiters[connectionID]
	bm.mu.RUnlock()

	if exists {
		return limiter
	}

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = bm.connectionLimiters[connectionID]; exists {
		return limiter
	}

	// Create new limiter for this connection
	limiter = NewRateLimiter(bm.config.MaxConnectionBPS, bm.config.MaxConnectionBPS/10)
	bm.connectionLimiters[connectionID] = limiter

	return limiter
}

// updateStats safely updates bandwidth statistics
func (bm *BandwidthManager) updateStats(fn func(*BandwidthStats)) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	fn(bm.stats)
}

// runMonitoring runs the bandwidth monitoring loop
func (bm *BandwidthManager) runMonitoring() {
	defer bm.wg.Done()

	ticker := time.NewTicker(bm.config.MonitorInterval)
	defer ticker.Stop()

	var lastInbound, lastOutbound int64
	var lastTime time.Time = time.Now()

	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.updateBandwidthStats(&lastInbound, &lastOutbound, &lastTime)
		}
	}
}

// updateBandwidthStats calculates current bandwidth usage
func (bm *BandwidthManager) updateBandwidthStats(lastInbound, lastOutbound *int64, lastTime *time.Time) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(*lastTime).Seconds()

	if elapsed > 0 {
		// Calculate current bandwidth usage
		inboundDelta := bm.stats.TotalInbound - *lastInbound
		outboundDelta := bm.stats.TotalOutbound - *lastOutbound

		bm.stats.InboundBPS = int64(float64(inboundDelta) / elapsed)
		bm.stats.OutboundBPS = int64(float64(outboundDelta) / elapsed)

		// Update peak values
		if bm.stats.InboundBPS > bm.stats.PeakInboundBPS {
			bm.stats.PeakInboundBPS = bm.stats.InboundBPS
		}

		if bm.stats.OutboundBPS > bm.stats.PeakOutboundBPS {
			bm.stats.PeakOutboundBPS = bm.stats.OutboundBPS
		}

		// Update connection count
		bm.stats.ActiveConnections = len(bm.connectionLimiters)
	}

	*lastInbound = bm.stats.TotalInbound
	*lastOutbound = bm.stats.TotalOutbound
	*lastTime = now
}

// runCleanup runs the cleanup loop for inactive connections
func (bm *BandwidthManager) runCleanup() {
	defer bm.wg.Done()

	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.cleanupInactiveConnections()
		}
	}
}

// cleanupInactiveConnections removes limiters for inactive connections
func (bm *BandwidthManager) cleanupInactiveConnections() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Remove limiters that haven't been used recently
	// This is a simplified cleanup - in practice, you'd track last usage
	for connectionID, limiter := range bm.connectionLimiters {
		if limiter.IsIdle() {
			delete(bm.connectionLimiters, connectionID)
		}
	}
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	rate     int64
	burst    int64
	tokens   int64
	lastTime time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate, burst int64) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   burst,
		lastTime: time.Now(),
	}
}

// Allow checks if the given number of bytes can be consumed
func (rl *RateLimiter) Allow(bytes int64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens >= bytes {
		rl.tokens -= bytes
		return true
	}

	return false
}

// Wait waits for the given number of bytes to become available
func (rl *RateLimiter) Wait(ctx context.Context, bytes int64) error {
	for {
		if rl.Allow(bytes) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Continue waiting
		}
	}
}

// SetRate updates the rate limit
func (rl *RateLimiter) SetRate(rate int64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.rate = rate
}

// IsIdle checks if the rate limiter has been idle
func (rl *RateLimiter) IsIdle() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return time.Since(rl.lastTime) > 5*time.Minute
}

// refillTokens adds tokens based on elapsed time
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()

	if elapsed > 0 {
		tokensToAdd := int64(float64(rl.rate) * elapsed)
		rl.tokens += tokensToAdd

		if rl.tokens > rl.burst {
			rl.tokens = rl.burst
		}

		rl.lastTime = now
	}
}
