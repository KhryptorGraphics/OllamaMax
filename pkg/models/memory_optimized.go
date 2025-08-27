package models

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryOptimizedBandwidthManager fixes the memory leaks in the original BandwidthManager
type MemoryOptimizedBandwidthManager struct {
	mu sync.RWMutex
	
	// Context for proper cleanup
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Original fields
	totalBandwidth  int64
	allocations     map[string]*BandwidthAllocation
	usageHistory    []BandwidthUsage
	adaptiveConfig  *AdaptiveBandwidthConfig
	
	// Optimization: bounded channels
	cleanupCh chan struct{}
	updateCh  chan *BandwidthUsage
}

// NewMemoryOptimizedBandwidthManager creates a new bandwidth manager with proper cleanup
func NewMemoryOptimizedBandwidthManager(totalBandwidth int64) *MemoryOptimizedBandwidthManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	bm := &MemoryOptimizedBandwidthManager{
		totalBandwidth:  totalBandwidth,
		allocations:     make(map[string]*BandwidthAllocation),
		usageHistory:    make([]BandwidthUsage, 0, 1000), // Pre-allocate capacity
		adaptiveConfig:  DefaultAdaptiveBandwidthConfig(),
		ctx:            ctx,
		cancel:         cancel,
		cleanupCh:      make(chan struct{}, 1),
		updateCh:       make(chan *BandwidthUsage, 100), // Bounded channel
	}
	
	// Start background workers with proper cleanup
	bm.startWorkers()
	
	return bm
}

// startWorkers starts background workers with proper lifecycle management
func (bm *MemoryOptimizedBandwidthManager) startWorkers() {
	// Cleanup worker
	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		bm.cleanupExpiredAllocationsOptimized()
	}()
	
	// Update history worker
	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		bm.updateUsageHistoryOptimized()
	}()
	
	// Adaptive throttling worker
	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		bm.adaptiveThrottlingOptimized()
	}()
}

// cleanupExpiredAllocationsOptimized runs with proper cancellation
func (bm *MemoryOptimizedBandwidthManager) cleanupExpiredAllocationsOptimized() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.performCleanup()
		}
	}
}

// updateUsageHistoryOptimized updates history with bounded growth
func (bm *MemoryOptimizedBandwidthManager) updateUsageHistoryOptimized() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case usage := <-bm.updateCh:
			bm.addUsageHistory(usage)
		case <-ticker.C:
			// Periodic history trimming
			bm.trimHistory()
		}
	}
}

// adaptiveThrottlingOptimized performs adaptive throttling with proper termination
func (bm *MemoryOptimizedBandwidthManager) adaptiveThrottlingOptimized() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			bm.performAdaptiveThrottling()
		}
	}
}

// performCleanup removes expired allocations
func (bm *MemoryOptimizedBandwidthManager) performCleanup() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	now := time.Now()
	for id, alloc := range bm.allocations {
		if alloc.ExpiresAt.Before(now) {
			delete(bm.allocations, id)
		}
	}
}

// addUsageHistory adds usage with bounded growth
func (bm *MemoryOptimizedBandwidthManager) addUsageHistory(usage *BandwidthUsage) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Use ring buffer pattern to prevent unbounded growth
	const maxHistory = 1000
	if len(bm.usageHistory) >= maxHistory {
		// Remove oldest entry efficiently
		copy(bm.usageHistory[0:], bm.usageHistory[1:])
		bm.usageHistory[len(bm.usageHistory)-1] = *usage
	} else {
		bm.usageHistory = append(bm.usageHistory, *usage)
	}
}

// trimHistory removes old history entries
func (bm *MemoryOptimizedBandwidthManager) trimHistory() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	cutoff := time.Now().Add(-1 * time.Hour)
	
	// Find first non-expired entry
	idx := 0
	for idx < len(bm.usageHistory) && bm.usageHistory[idx].Timestamp.Before(cutoff) {
		idx++
	}
	
	// Remove expired entries
	if idx > 0 {
		bm.usageHistory = bm.usageHistory[idx:]
	}
}

// performAdaptiveThrottling adjusts bandwidth based on usage patterns
func (bm *MemoryOptimizedBandwidthManager) performAdaptiveThrottling() {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	if len(bm.usageHistory) < 10 {
		return
	}
	
	// Calculate average usage
	var totalUsage int64
	for i := len(bm.usageHistory) - 10; i < len(bm.usageHistory); i++ {
		totalUsage += bm.usageHistory[i].BytesTransferred
	}
	avgUsage := totalUsage / 10
	
	// Adjust bandwidth allocation based on usage
	if avgUsage > bm.totalBandwidth*8/10 {
		// High usage, consider increasing limits
		atomic.AddInt64(&bm.totalBandwidth, bm.totalBandwidth/10)
	} else if avgUsage < bm.totalBandwidth*2/10 {
		// Low usage, consider decreasing limits
		atomic.AddInt64(&bm.totalBandwidth, -bm.totalBandwidth/10)
	}
}

// Stop gracefully shuts down the bandwidth manager
func (bm *MemoryOptimizedBandwidthManager) Stop() error {
	// Signal all workers to stop
	bm.cancel()
	
	// Wait for all workers to complete
	bm.wg.Wait()
	
	// Clean up channels
	close(bm.cleanupCh)
	close(bm.updateCh)
	
	return nil
}

// BoundedConflictQueue implements a bounded queue for conflict resolution
// This prevents unbounded channel growth
type BoundedConflictQueue struct {
	ch      chan *ConflictResolutionTask
	size    int32
	maxSize int32
	mu      sync.RWMutex
}

// NewBoundedConflictQueue creates a new bounded queue
func NewBoundedConflictQueue(maxSize int) *BoundedConflictQueue {
	return &BoundedConflictQueue{
		ch:      make(chan *ConflictResolutionTask, maxSize),
		maxSize: int32(maxSize),
		size:    0,
	}
}

// Enqueue attempts to add a task to the queue
func (bq *BoundedConflictQueue) Enqueue(task *ConflictResolutionTask) bool {
	// Check if queue is full
	if atomic.LoadInt32(&bq.size) >= bq.maxSize {
		return false // Drop task rather than block
	}
	
	select {
	case bq.ch <- task:
		atomic.AddInt32(&bq.size, 1)
		return true
	default:
		return false // Channel full
	}
}

// Dequeue retrieves a task from the queue
func (bq *BoundedConflictQueue) Dequeue() (*ConflictResolutionTask, bool) {
	select {
	case task := <-bq.ch:
		atomic.AddInt32(&bq.size, -1)
		return task, true
	default:
		return nil, false
	}
}

// Size returns the current queue size
func (bq *BoundedConflictQueue) Size() int {
	return int(atomic.LoadInt32(&bq.size))
}

// Close closes the queue
func (bq *BoundedConflictQueue) Close() {
	close(bq.ch)
}

// StringInternPool provides string interning to reduce memory usage
// This reduces duplicate string allocations
type StringInternPool struct {
	cache map[string]string
	mu    sync.RWMutex
}

// NewStringInternPool creates a new string intern pool
func NewStringInternPool() *StringInternPool {
	return &StringInternPool{
		cache: make(map[string]string),
	}
}

// Intern returns an interned version of the string
func (sip *StringInternPool) Intern(s string) string {
	sip.mu.RLock()
	if cached, exists := sip.cache[s]; exists {
		sip.mu.RUnlock()
		return cached
	}
	sip.mu.RUnlock()
	
	// Double-check after acquiring write lock
	sip.mu.Lock()
	defer sip.mu.Unlock()
	
	if cached, exists := sip.cache[s]; exists {
		return cached
	}
	
	sip.cache[s] = s
	return s
}

// Size returns the number of interned strings
func (sip *StringInternPool) Size() int {
	sip.mu.RLock()
	defer sip.mu.RUnlock()
	return len(sip.cache)
}

// Clear removes all interned strings
func (sip *StringInternPool) Clear() {
	sip.mu.Lock()
	defer sip.mu.Unlock()
	sip.cache = make(map[string]string)
}