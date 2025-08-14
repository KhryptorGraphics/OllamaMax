package models

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// BandwidthManager manages bandwidth allocation and throttling
type BandwidthManager struct {
	mu sync.RWMutex

	maxBandwidth    int64
	currentUsage    int64
	logger          *slog.Logger

	// Bandwidth allocation tracking
	allocations     map[string]*BandwidthAllocation
	reservations    map[string]*BandwidthReservation

	// Rate limiting
	rateLimiter     *TokenBucket
	
	// Monitoring
	usageHistory    []*BandwidthUsageEntry
	maxHistorySize  int
	
	// Adaptive throttling
	adaptiveThrottle *AdaptiveThrottle
}

// BandwidthAllocation represents an active bandwidth allocation
type BandwidthAllocation struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"task_id"`
	AllocatedBytes  int64         `json:"allocated_bytes"`
	UsedBytes       int64         `json:"used_bytes"`
	StartTime       time.Time     `json:"start_time"`
	LastUsed        time.Time     `json:"last_used"`
	Priority        int           `json:"priority"`
	MaxDuration     time.Duration `json:"max_duration"`
}

// BandwidthReservation represents a future bandwidth reservation
type BandwidthReservation struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"task_id"`
	RequestedBytes  int64         `json:"requested_bytes"`
	StartTime       time.Time     `json:"start_time"`
	Duration        time.Duration `json:"duration"`
	Priority        int           `json:"priority"`
	CreatedAt       time.Time     `json:"created_at"`
}

// BandwidthUsageEntry records bandwidth usage over time
type BandwidthUsageEntry struct {
	Timestamp       time.Time `json:"timestamp"`
	TotalUsage      int64     `json:"total_usage"`
	AllocatedUsage  int64     `json:"allocated_usage"`
	PeakUsage       int64     `json:"peak_usage"`
	ActiveTasks     int       `json:"active_tasks"`
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	mu sync.Mutex
	
	capacity    int64
	tokens      int64
	refillRate  int64
	lastRefill  time.Time
}

// AdaptiveThrottle implements adaptive bandwidth throttling
type AdaptiveThrottle struct {
	mu sync.RWMutex
	
	enabled         bool
	currentFactor   float64
	targetUtilization float64
	adjustmentRate  float64
	lastAdjustment  time.Time
}

// NewBandwidthManager creates a new bandwidth manager
func NewBandwidthManager(maxBandwidth int64, logger *slog.Logger) *BandwidthManager {
	bm := &BandwidthManager{
		maxBandwidth:    maxBandwidth,
		logger:          logger,
		allocations:     make(map[string]*BandwidthAllocation),
		reservations:    make(map[string]*BandwidthReservation),
		maxHistorySize:  1000,
		rateLimiter:     NewTokenBucket(maxBandwidth, maxBandwidth/10), // 10% refill rate
		adaptiveThrottle: &AdaptiveThrottle{
			enabled:           true,
			currentFactor:     1.0,
			targetUtilization: 0.8,
			adjustmentRate:    0.1,
			lastAdjustment:    time.Now(),
		},
	}
	
	// Start background tasks
	go bm.cleanupExpiredAllocations()
	go bm.updateUsageHistory()
	go bm.adaptiveThrottling()
	
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

// AllocateBandwidth allocates bandwidth for a task
func (bm *BandwidthManager) AllocateBandwidth(taskID string, requestedBytes int64, priority int, maxDuration time.Duration) (*BandwidthAllocation, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	// Check if we have enough available bandwidth
	availableBandwidth := bm.maxBandwidth - bm.currentUsage
	
	// Apply adaptive throttling
	throttleFactor := bm.adaptiveThrottle.currentFactor
	effectiveRequest := int64(float64(requestedBytes) * throttleFactor)
	
	if effectiveRequest > availableBandwidth {
		// Try to free up bandwidth by cleaning up expired allocations
		bm.cleanupExpiredAllocationsLocked()
		availableBandwidth = bm.maxBandwidth - bm.currentUsage
		
		if effectiveRequest > availableBandwidth {
			return nil, &BandwidthError{
				Type:      "insufficient_bandwidth",
				Message:   "Not enough bandwidth available",
				Requested: requestedBytes,
				Available: availableBandwidth,
			}
		}
	}
	
	// Create allocation
	allocation := &BandwidthAllocation{
		ID:             generateAllocationID(),
		TaskID:         taskID,
		AllocatedBytes: effectiveRequest,
		UsedBytes:      0,
		StartTime:      time.Now(),
		LastUsed:       time.Now(),
		Priority:       priority,
		MaxDuration:    maxDuration,
	}
	
	bm.allocations[allocation.ID] = allocation
	bm.currentUsage += effectiveRequest
	
	bm.logger.Info("bandwidth allocated",
		"allocation_id", allocation.ID,
		"task_id", taskID,
		"allocated_bytes", effectiveRequest,
		"priority", priority,
		"current_usage", bm.currentUsage,
		"max_bandwidth", bm.maxBandwidth)
	
	return allocation, nil
}

// ReleaseBandwidth releases bandwidth allocation
func (bm *BandwidthManager) ReleaseBandwidth(allocationID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return &BandwidthError{
			Type:    "allocation_not_found",
			Message: "Bandwidth allocation not found",
		}
	}
	
	bm.currentUsage -= allocation.AllocatedBytes
	delete(bm.allocations, allocationID)
	
	bm.logger.Info("bandwidth released",
		"allocation_id", allocationID,
		"task_id", allocation.TaskID,
		"allocated_bytes", allocation.AllocatedBytes,
		"used_bytes", allocation.UsedBytes,
		"current_usage", bm.currentUsage)
	
	return nil
}

// UpdateUsage updates bandwidth usage for an allocation
func (bm *BandwidthManager) UpdateUsage(allocationID string, usedBytes int64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	allocation, exists := bm.allocations[allocationID]
	if !exists {
		return &BandwidthError{
			Type:    "allocation_not_found",
			Message: "Bandwidth allocation not found",
		}
	}
	
	allocation.UsedBytes = usedBytes
	allocation.LastUsed = time.Now()
	
	return nil
}

// ReserveBandwidth reserves bandwidth for future use
func (bm *BandwidthManager) ReserveBandwidth(taskID string, requestedBytes int64, startTime time.Time, duration time.Duration, priority int) (*BandwidthReservation, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	reservation := &BandwidthReservation{
		ID:             generateReservationID(),
		TaskID:         taskID,
		RequestedBytes: requestedBytes,
		StartTime:      startTime,
		Duration:       duration,
		Priority:       priority,
		CreatedAt:      time.Now(),
	}
	
	bm.reservations[reservation.ID] = reservation
	
	bm.logger.Info("bandwidth reserved",
		"reservation_id", reservation.ID,
		"task_id", taskID,
		"requested_bytes", requestedBytes,
		"start_time", startTime,
		"duration", duration)
	
	return reservation, nil
}

// GetAvailableBandwidth returns currently available bandwidth
func (bm *BandwidthManager) GetAvailableBandwidth() int64 {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	return bm.maxBandwidth - bm.currentUsage
}

// GetUsageStats returns current usage statistics
func (bm *BandwidthManager) GetUsageStats() *BandwidthStats {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	
	stats := &BandwidthStats{
		MaxBandwidth:     bm.maxBandwidth,
		CurrentUsage:     bm.currentUsage,
		AvailableBandwidth: bm.maxBandwidth - bm.currentUsage,
		UtilizationRate:  float64(bm.currentUsage) / float64(bm.maxBandwidth),
		ActiveAllocations: len(bm.allocations),
		ActiveReservations: len(bm.reservations),
		ThrottleFactor:   bm.adaptiveThrottle.currentFactor,
	}
	
	return stats
}

// BandwidthStats contains bandwidth usage statistics
type BandwidthStats struct {
	MaxBandwidth       int64   `json:"max_bandwidth"`
	CurrentUsage       int64   `json:"current_usage"`
	AvailableBandwidth int64   `json:"available_bandwidth"`
	UtilizationRate    float64 `json:"utilization_rate"`
	ActiveAllocations  int     `json:"active_allocations"`
	ActiveReservations int     `json:"active_reservations"`
	ThrottleFactor     float64 `json:"throttle_factor"`
}

// BandwidthError represents bandwidth-related errors
type BandwidthError struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Requested int64  `json:"requested,omitempty"`
	Available int64  `json:"available,omitempty"`
}

func (be *BandwidthError) Error() string {
	return be.Message
}

// Background tasks

func (bm *BandwidthManager) cleanupExpiredAllocations() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		bm.mu.Lock()
		bm.cleanupExpiredAllocationsLocked()
		bm.mu.Unlock()
	}
}

func (bm *BandwidthManager) cleanupExpiredAllocationsLocked() {
	now := time.Now()
	
	for id, allocation := range bm.allocations {
		if now.Sub(allocation.StartTime) > allocation.MaxDuration {
			bm.currentUsage -= allocation.AllocatedBytes
			delete(bm.allocations, id)
			
			bm.logger.Info("expired allocation cleaned up",
				"allocation_id", id,
				"task_id", allocation.TaskID,
				"allocated_bytes", allocation.AllocatedBytes)
		}
	}
}

func (bm *BandwidthManager) updateUsageHistory() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		bm.mu.Lock()
		
		entry := &BandwidthUsageEntry{
			Timestamp:      time.Now(),
			TotalUsage:     bm.currentUsage,
			AllocatedUsage: bm.currentUsage,
			ActiveTasks:    len(bm.allocations),
		}
		
		bm.usageHistory = append(bm.usageHistory, entry)
		
		// Limit history size
		if len(bm.usageHistory) > bm.maxHistorySize {
			bm.usageHistory = bm.usageHistory[1:]
		}
		
		bm.mu.Unlock()
	}
}

func (bm *BandwidthManager) adaptiveThrottling() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		bm.adaptiveThrottle.mu.Lock()
		
		if !bm.adaptiveThrottle.enabled {
			bm.adaptiveThrottle.mu.Unlock()
			continue
		}
		
		// Calculate current utilization
		bm.mu.RLock()
		utilization := float64(bm.currentUsage) / float64(bm.maxBandwidth)
		bm.mu.RUnlock()
		
		// Adjust throttle factor based on utilization
		target := bm.adaptiveThrottle.targetUtilization
		rate := bm.adaptiveThrottle.adjustmentRate
		
		if utilization > target {
			// Reduce throttle factor to decrease load
			bm.adaptiveThrottle.currentFactor *= (1.0 - rate)
		} else if utilization < target*0.8 {
			// Increase throttle factor to allow more load
			bm.adaptiveThrottle.currentFactor *= (1.0 + rate)
		}
		
		// Keep throttle factor within reasonable bounds
		if bm.adaptiveThrottle.currentFactor < 0.1 {
			bm.adaptiveThrottle.currentFactor = 0.1
		} else if bm.adaptiveThrottle.currentFactor > 1.0 {
			bm.adaptiveThrottle.currentFactor = 1.0
		}
		
		bm.adaptiveThrottle.lastAdjustment = time.Now()
		bm.adaptiveThrottle.mu.Unlock()
	}
}

// Helper functions

func generateAllocationID() string {
	return "alloc_" + generateRandomID()
}

func generateReservationID() string {
	return "resv_" + generateRandomID()
}

func generateRandomID() string {
	// Simple random ID generation - in production, use crypto/rand
	return time.Now().Format("20060102150405") + "_" + "random"
}
