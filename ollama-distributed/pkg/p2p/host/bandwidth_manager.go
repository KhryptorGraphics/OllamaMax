package host

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// BandwidthManager manages bandwidth allocation and throttling
type BandwidthManager struct {
	mu             sync.RWMutex
	globalLimit    int64 // bytes per second
	peerLimits     map[peer.ID]int64
	peerUsage      map[peer.ID]*BandwidthUsage
	protocolLimits map[string]int64
	protocolUsage  map[string]*BandwidthUsage

	// Token bucket for rate limiting
	globalBucket    *TokenBucket
	peerBuckets     map[peer.ID]*TokenBucket
	protocolBuckets map[string]*TokenBucket

	// Configuration
	config *BandwidthConfig

	// Metrics
	metrics *BandwidthMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// BandwidthUsage tracks bandwidth usage over time
type BandwidthUsage struct {
	BytesSent     int64
	BytesReceived int64
	LastUpdated   time.Time
	WindowStart   time.Time
	WindowBytes   int64
	PeakUsage     int64
}

// BandwidthConfig configures bandwidth management
type BandwidthConfig struct {
	GlobalLimit             int64         // Global bandwidth limit (bytes/sec)
	DefaultPeerLimit        int64         // Default per-peer limit (bytes/sec)
	TensorStreamingPriority bool          // Prioritize tensor streaming traffic
	TensorStreamingLimit    int64         // Dedicated bandwidth for tensor streaming (bytes/sec)
	WindowSize              time.Duration // Time window for rate limiting
	BurstSize               int64         // Maximum burst size
	UpdateInterval          time.Duration // How often to update usage stats

	// Protocol-specific limits
	ProtocolLimits map[string]int64

	// Quality of Service
	PriorityProtocols  []string
	PriorityMultiplier float64
}

// BandwidthMetrics tracks bandwidth usage metrics
type BandwidthMetrics struct {
	TotalBytesSent     int64
	TotalBytesReceived int64
	CurrentUsage       int64
	PeakUsage          int64
	ThrottledRequests  int64
	LastUpdated        time.Time
}

// TokenBucket implements a token bucket for rate limiting
type TokenBucket struct {
	mu         sync.Mutex
	capacity   int64
	tokens     int64
	refillRate int64 // tokens per second
	lastRefill time.Time
}

// NewBandwidthManager creates a new bandwidth manager
func NewBandwidthManager(config *BandwidthConfig) *BandwidthManager {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &BandwidthConfig{
			GlobalLimit:        100 * 1024 * 1024, // 100 MB/s
			DefaultPeerLimit:   10 * 1024 * 1024,  // 10 MB/s per peer
			WindowSize:         time.Second,
			BurstSize:          10 * 1024 * 1024, // 10 MB burst
			UpdateInterval:     time.Second,
			ProtocolLimits:     make(map[string]int64),
			PriorityProtocols:  []string{"/ollama/consensus/1.0.0", "/ollama/health/1.0.0"},
			PriorityMultiplier: 2.0,
		}
	}

	bm := &BandwidthManager{
		globalLimit:     config.GlobalLimit,
		peerLimits:      make(map[peer.ID]int64),
		peerUsage:       make(map[peer.ID]*BandwidthUsage),
		protocolLimits:  config.ProtocolLimits,
		protocolUsage:   make(map[string]*BandwidthUsage),
		globalBucket:    NewTokenBucket(config.GlobalLimit, config.BurstSize),
		peerBuckets:     make(map[peer.ID]*TokenBucket),
		protocolBuckets: make(map[string]*TokenBucket),
		config:          config,
		metrics:         &BandwidthMetrics{},
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize protocol buckets
	for protocol, limit := range config.ProtocolLimits {
		bm.protocolBuckets[protocol] = NewTokenBucket(limit, config.BurstSize)
	}

	// Start background tasks
	bm.wg.Add(1)
	go bm.updateLoop()

	return bm
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TryConsume attempts to consume tokens from the bucket
func (tb *TokenBucket) TryConsume(tokens int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate

	tb.tokens += tokensToAdd
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now

	// Check if we have enough tokens
	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

// CheckBandwidth checks if a data transfer is allowed
func (bm *BandwidthManager) CheckBandwidth(peerID peer.ID, protocol string, bytes int64) bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	// Check global limit
	if !bm.globalBucket.TryConsume(bytes) {
		bm.metrics.ThrottledRequests++
		return false
	}

	// Check peer limit
	peerBucket, exists := bm.peerBuckets[peerID]
	if !exists {
		limit := bm.config.DefaultPeerLimit
		if peerLimit, hasCustomLimit := bm.peerLimits[peerID]; hasCustomLimit {
			limit = peerLimit
		}
		peerBucket = NewTokenBucket(limit, bm.config.BurstSize)
		bm.peerBuckets[peerID] = peerBucket
	}

	if !peerBucket.TryConsume(bytes) {
		bm.metrics.ThrottledRequests++
		return false
	}

	// Check protocol limit
	if protocolBucket, exists := bm.protocolBuckets[protocol]; exists {
		if !protocolBucket.TryConsume(bytes) {
			bm.metrics.ThrottledRequests++
			return false
		}
	}

	return true
}

// RecordUsage records bandwidth usage for a peer and protocol
func (bm *BandwidthManager) RecordUsage(peerID peer.ID, protocol string, bytesSent, bytesReceived int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()

	// Update peer usage
	if usage, exists := bm.peerUsage[peerID]; exists {
		usage.BytesSent += bytesSent
		usage.BytesReceived += bytesReceived
		usage.LastUpdated = now

		// Update window usage
		if now.Sub(usage.WindowStart) >= bm.config.WindowSize {
			usage.WindowStart = now
			usage.WindowBytes = bytesSent + bytesReceived
		} else {
			usage.WindowBytes += bytesSent + bytesReceived
		}

		if usage.WindowBytes > usage.PeakUsage {
			usage.PeakUsage = usage.WindowBytes
		}
	} else {
		bm.peerUsage[peerID] = &BandwidthUsage{
			BytesSent:     bytesSent,
			BytesReceived: bytesReceived,
			LastUpdated:   now,
			WindowStart:   now,
			WindowBytes:   bytesSent + bytesReceived,
			PeakUsage:     bytesSent + bytesReceived,
		}
	}

	// Update protocol usage
	if usage, exists := bm.protocolUsage[protocol]; exists {
		usage.BytesSent += bytesSent
		usage.BytesReceived += bytesReceived
		usage.LastUpdated = now
	} else {
		bm.protocolUsage[protocol] = &BandwidthUsage{
			BytesSent:     bytesSent,
			BytesReceived: bytesReceived,
			LastUpdated:   now,
			WindowStart:   now,
			WindowBytes:   bytesSent + bytesReceived,
		}
	}

	// Update global metrics
	bm.metrics.TotalBytesSent += bytesSent
	bm.metrics.TotalBytesReceived += bytesReceived
	bm.metrics.LastUpdated = now
}

// SetPeerLimit sets a custom bandwidth limit for a specific peer
func (bm *BandwidthManager) SetPeerLimit(peerID peer.ID, limit int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.peerLimits[peerID] = limit

	// Update or create bucket with new limit
	bm.peerBuckets[peerID] = NewTokenBucket(limit, bm.config.BurstSize)
}

// GetPeerUsage returns bandwidth usage for a specific peer
func (bm *BandwidthManager) GetPeerUsage(peerID peer.ID) *BandwidthUsage {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if usage, exists := bm.peerUsage[peerID]; exists {
		// Return a copy to avoid race conditions
		usageCopy := *usage
		return &usageCopy
	}

	return nil
}

// GetProtocolUsage returns bandwidth usage for a specific protocol
func (bm *BandwidthManager) GetProtocolUsage(protocol string) *BandwidthUsage {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if usage, exists := bm.protocolUsage[protocol]; exists {
		// Return a copy to avoid race conditions
		usageCopy := *usage
		return &usageCopy
	}

	return nil
}

// GetCurrentUsage calculates current bandwidth usage across all peers
func (bm *BandwidthManager) GetCurrentUsage() int64 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	now := time.Now()
	totalUsage := int64(0)

	for _, usage := range bm.peerUsage {
		// Only count usage from the current window
		if now.Sub(usage.WindowStart) < bm.config.WindowSize {
			totalUsage += usage.WindowBytes
		}
	}

	return totalUsage
}

// GetMetrics returns bandwidth metrics
func (bm *BandwidthManager) GetMetrics() *BandwidthMetrics {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	metrics := *bm.metrics
	metrics.CurrentUsage = bm.GetCurrentUsage()

	return &metrics
}

// updateLoop periodically updates bandwidth statistics
func (bm *BandwidthManager) updateLoop() {
	defer bm.wg.Done()

	ticker := time.NewTicker(bm.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.updateStats()
		}
	}
}

// updateStats updates bandwidth statistics
func (bm *BandwidthManager) updateStats() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()
	currentUsage := int64(0)

	// Clean up old usage data and calculate current usage
	for peerID, usage := range bm.peerUsage {
		if now.Sub(usage.LastUpdated) > 5*time.Minute {
			// Remove stale usage data
			delete(bm.peerUsage, peerID)
			delete(bm.peerBuckets, peerID)
		} else if now.Sub(usage.WindowStart) < bm.config.WindowSize {
			currentUsage += usage.WindowBytes
		}
	}

	bm.metrics.CurrentUsage = currentUsage
	if currentUsage > bm.metrics.PeakUsage {
		bm.metrics.PeakUsage = currentUsage
	}
}

// IsPriorityProtocol checks if a protocol has priority
func (bm *BandwidthManager) IsPriorityProtocol(protocol string) bool {
	for _, priorityProtocol := range bm.config.PriorityProtocols {
		if protocol == priorityProtocol {
			return true
		}
	}
	return false
}

// AllocateTensorStreamingBandwidth allocates bandwidth specifically for tensor streaming
func (bm *BandwidthManager) AllocateTensorStreamingBandwidth(peerID peer.ID, bytes int64) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Check if tensor streaming has dedicated bandwidth
	if bm.config.TensorStreamingLimit > 0 {
		// Use dedicated tensor streaming bandwidth
		if bucket, exists := bm.protocolBuckets["tensor-stream"]; exists {
			return bucket.TryConsume(bytes)
		}
	}

	// Fall back to regular bandwidth allocation with priority
	if bm.config.TensorStreamingPriority {
		// Give tensor streaming higher priority
		return bm.allocateBandwidthWithPriority(peerID, "tensor-stream", bytes)
	}

	return bm.CheckBandwidth(peerID, "tensor-stream", bytes)
}

// allocateBandwidthWithPriority allocates bandwidth with priority for tensor streaming
func (bm *BandwidthManager) allocateBandwidthWithPriority(peerID peer.ID, protocol string, bytes int64) bool {
	// Try global bucket first
	if !bm.globalBucket.TryConsume(bytes) {
		return false
	}

	// Try peer bucket with relaxed limits for tensor streaming
	if bucket, exists := bm.peerBuckets[peerID]; exists {
		if protocol == "tensor-stream" {
			// Allow burst for tensor streaming
			burstBytes := bytes * 2 // Allow 2x burst for tensor streaming
			if bucket.TryConsume(burstBytes) {
				return true
			}
		}
		return bucket.TryConsume(bytes)
	}

	return true
}

// GetTensorStreamingMetrics returns metrics specific to tensor streaming
func (bm *BandwidthManager) GetTensorStreamingMetrics() map[string]interface{} {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	metrics := make(map[string]interface{})

	// Get tensor streaming protocol usage
	if usage, exists := bm.protocolUsage["tensor-stream"]; exists {
		metrics["bytes_sent"] = usage.BytesSent
		metrics["bytes_received"] = usage.BytesReceived
		metrics["peak_usage"] = usage.PeakUsage
		metrics["last_updated"] = usage.LastUpdated
	}

	// Get tensor streaming bandwidth allocation
	if bm.config.TensorStreamingLimit > 0 {
		metrics["dedicated_limit"] = bm.config.TensorStreamingLimit
		metrics["priority_enabled"] = bm.config.TensorStreamingPriority
	}

	return metrics
}

// OptimizeForTensorStreaming optimizes bandwidth allocation for tensor streaming workloads
func (bm *BandwidthManager) OptimizeForTensorStreaming() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Enable tensor streaming priority
	bm.config.TensorStreamingPriority = true

	// Allocate 70% of global bandwidth to tensor streaming if not already set
	if bm.config.TensorStreamingLimit == 0 {
		bm.config.TensorStreamingLimit = int64(float64(bm.config.GlobalLimit) * 0.7)
	}

	// Create dedicated bucket for tensor streaming
	if _, exists := bm.protocolBuckets["tensor-stream"]; !exists {
		bm.protocolBuckets["tensor-stream"] = NewTokenBucket(
			bm.config.TensorStreamingLimit,
			bm.config.BurstSize*2, // Allow larger bursts for tensor streaming
		)
	}

	// Initialize usage tracking for tensor streaming
	if _, exists := bm.protocolUsage["tensor-stream"]; !exists {
		bm.protocolUsage["tensor-stream"] = &BandwidthUsage{
			LastUpdated: time.Now(),
			WindowStart: time.Now(),
		}
	}
}

// Close closes the bandwidth manager
func (bm *BandwidthManager) Close() error {
	bm.cancel()
	bm.wg.Wait()
	return nil
}
