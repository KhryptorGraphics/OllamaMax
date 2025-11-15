package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// BackPressureManager implements flow control to prevent overwhelming slow peers
type BackPressureManager struct {
	mu sync.RWMutex

	// Queue monitoring
	peerQueues map[peer.ID]*PeerQueue
	
	// Configuration
	config *BackPressureConfig

	// Throttling state
	throttledPeers map[peer.ID]*ThrottleState

	// Context for periodic operations
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// PeerQueue tracks queue state for a peer
type PeerQueue struct {
	PeerID          peer.ID
	QueueDepth      int
	MaxQueueDepth   int
	ProcessingRate  float64 // Tasks per second
	LastUpdate      time.Time
	DepthHistory    []int
	IsThrottled     bool
	ThrottleReason  string
}

// ThrottleState tracks throttling state for a peer
type ThrottleState struct {
	PeerID          peer.ID
	IsThrottled     bool
	ThrottleStart   time.Time
	ThrottleDuration time.Duration
	Reason          string
	ThrottleCount   int64
}

// BackPressureConfig configures back-pressure behavior
type BackPressureConfig struct {
	MaxQueueDepth       int           // Maximum queue depth before throttling
	QueueDepthThreshold float64       // Threshold ratio for warning (0.0-1.0)
	ThrottleDuration    time.Duration // How long to throttle a peer
	MonitorInterval     time.Duration // How often to check queue depths
	AdaptiveThrottling  bool          // Enable adaptive throttling based on peer performance
	MinProcessingRate   float64       // Minimum processing rate (tasks/sec)
}

// DefaultBackPressureConfig returns default configuration
func DefaultBackPressureConfig() *BackPressureConfig {
	return &BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    5 * time.Second,
		MonitorInterval:     1 * time.Second,
		AdaptiveThrottling:  true,
		MinProcessingRate:   1.0,
	}
}

// NewBackPressureManager creates a new back-pressure manager
func NewBackPressureManager(config *BackPressureConfig) *BackPressureManager {
	if config == nil {
		config = DefaultBackPressureConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bpm := &BackPressureManager{
		peerQueues:     make(map[peer.ID]*PeerQueue),
		throttledPeers: make(map[peer.ID]*ThrottleState),
		config:         config,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start monitoring goroutine
	bpm.wg.Add(1)
	go bpm.monitorQueues()

	return bpm
}

// UpdateQueueDepth updates the queue depth for a peer
func (bpm *BackPressureManager) UpdateQueueDepth(peerID peer.ID, depth int) {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()

	queue, exists := bpm.peerQueues[peerID]
	if !exists {
		queue = &PeerQueue{
			PeerID:        peerID,
			MaxQueueDepth: bpm.config.MaxQueueDepth,
			DepthHistory:  make([]int, 0, 100),
		}
		bpm.peerQueues[peerID] = queue
	}

	queue.QueueDepth = depth
	queue.LastUpdate = time.Now()

	// Update history
	queue.DepthHistory = append(queue.DepthHistory, depth)
	if len(queue.DepthHistory) > 100 {
		queue.DepthHistory = queue.DepthHistory[1:]
	}

	// Calculate processing rate
	if len(queue.DepthHistory) >= 2 {
		queue.ProcessingRate = bpm.calculateProcessingRate(queue)
	}

	// Check if throttling is needed
	bpm.checkThrottling(queue)
}

// calculateProcessingRate calculates the processing rate based on queue depth history
func (bpm *BackPressureManager) calculateProcessingRate(queue *PeerQueue) float64 {
	if len(queue.DepthHistory) < 2 {
		return 0.0
	}

	// Simple rate calculation: change in queue depth over time
	recentDepths := queue.DepthHistory[len(queue.DepthHistory)-10:]
	if len(recentDepths) < 2 {
		return 0.0
	}

	totalChange := 0
	for i := 1; i < len(recentDepths); i++ {
		totalChange += recentDepths[i-1] - recentDepths[i]
	}

	// Positive rate means queue is draining (good)
	rate := float64(totalChange) / float64(len(recentDepths)-1)
	return rate
}

// checkThrottling checks if a peer should be throttled
func (bpm *BackPressureManager) checkThrottling(queue *PeerQueue) {
	// Check if queue depth exceeds threshold
	depthRatio := float64(queue.QueueDepth) / float64(queue.MaxQueueDepth)
	
	if depthRatio >= bpm.config.QueueDepthThreshold {
		bpm.throttlePeer(queue.PeerID, "queue depth exceeded threshold")
	} else if bpm.config.AdaptiveThrottling && queue.ProcessingRate < bpm.config.MinProcessingRate {
		bpm.throttlePeer(queue.PeerID, "processing rate too low")
	} else {
		bpm.unthrottlePeer(queue.PeerID)
	}
}

// throttlePeer throttles a peer
func (bpm *BackPressureManager) throttlePeer(peerID peer.ID, reason string) {
	throttle, exists := bpm.throttledPeers[peerID]
	if !exists {
		throttle = &ThrottleState{
			PeerID: peerID,
		}
		bpm.throttledPeers[peerID] = throttle
	}

	if !throttle.IsThrottled {
		throttle.IsThrottled = true
		throttle.ThrottleStart = time.Now()
		throttle.ThrottleDuration = bpm.config.ThrottleDuration
		throttle.Reason = reason
		throttle.ThrottleCount++
	}

	// Update queue state
	if queue, exists := bpm.peerQueues[peerID]; exists {
		queue.IsThrottled = true
		queue.ThrottleReason = reason
	}
}

// unthrottlePeer removes throttling from a peer
func (bpm *BackPressureManager) unthrottlePeer(peerID peer.ID) {
	if throttle, exists := bpm.throttledPeers[peerID]; exists {
		throttle.IsThrottled = false
	}

	if queue, exists := bpm.peerQueues[peerID]; exists {
		queue.IsThrottled = false
		queue.ThrottleReason = ""
	}
}

// IsThrottled checks if a peer is currently throttled
func (bpm *BackPressureManager) IsThrottled(peerID peer.ID) bool {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()

	throttle, exists := bpm.throttledPeers[peerID]
	if !exists {
		return false
	}

	// Check if throttle duration has expired
	if throttle.IsThrottled && time.Since(throttle.ThrottleStart) > throttle.ThrottleDuration {
		return false
	}

	return throttle.IsThrottled
}

// GetQueueDepth returns the current queue depth for a peer
func (bpm *BackPressureManager) GetQueueDepth(peerID peer.ID) (int, error) {
	bpm.mu.RLock()
	defer bpm.mu.RUnlock()

	queue, exists := bpm.peerQueues[peerID]
	if !exists {
		return 0, fmt.Errorf("no queue found for peer %s", peerID)
	}

	return queue.QueueDepth, nil
}

// monitorQueues periodically monitors queue depths and adjusts throttling
func (bpm *BackPressureManager) monitorQueues() {
	defer bpm.wg.Done()

	ticker := time.NewTicker(bpm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bpm.ctx.Done():
			return
		case <-ticker.C:
			bpm.checkAllQueues()
		}
	}
}

// checkAllQueues checks all peer queues and updates throttling
func (bpm *BackPressureManager) checkAllQueues() {
	bpm.mu.Lock()
	defer bpm.mu.Unlock()

	for _, queue := range bpm.peerQueues {
		bpm.checkThrottling(queue)
	}
}

// Stop stops the back-pressure manager
func (bpm *BackPressureManager) Stop() {
	bpm.cancel()
	bpm.wg.Wait()
}

