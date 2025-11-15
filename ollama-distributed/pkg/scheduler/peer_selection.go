package scheduler

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/network"
)

// PeerPerformanceTracker tracks peer performance metrics for utility scoring
type PeerPerformanceTracker struct {
	// Performance metrics
	peerMetrics map[peer.ID]*PeerMetrics
	metricsMux  sync.RWMutex

	// Configuration
	config *PeerSelectionConfig

	// Context for periodic operations
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// PeerMetrics tracks performance metrics for a single peer
type PeerMetrics struct {
	// Resource metrics
	EffectiveFLOPs     float64 // Effective FLOPs based on processing speed
	AvailableMemory    uint64  // Available memory in bytes
	TotalMemory        uint64  // Total memory in bytes

	// Network metrics
	Latency        time.Duration // Average latency to peer
	LatencyHistory []time.Duration // Recent latency measurements
	LatencyPenalty float64       // Calculated latency penalty

	// Congestion metrics
	ActiveTasks        int           // Number of active tasks
	QueueSize          int           // Task queue size
	EstimatedCongestion float64      // Estimated network congestion
	CongestionHistory  []float64     // Recent congestion measurements

	// Utility scoring
	LastUtilityScore float64       // Last calculated utility score
	LastUpdated      time.Time     // Last time metrics were updated

	// Performance history
	SuccessRate      float64       // Task success rate
	TotalTasks       int64         // Total tasks processed
	SuccessfulTasks  int64         // Successfully completed tasks
	FailedTasks      int64         // Failed tasks

	// Timestamps
	LastSeen         time.Time     // Last time peer was active
	LastTaskComplete time.Time     // Last time a task completed
}

// PeerSelectionConfig configures peer selection behavior
type PeerSelectionConfig struct {
	// Utility scoring parameters
	LatencyWeight        float64       // Weight for latency penalty in utility calculation
	CongestionWeight     float64       // Weight for congestion in utility calculation
	MemoryWeight         float64       // Weight for memory factor in utility calculation
	FLOPsWeight          float64       // Weight for FLOPs factor in utility calculation

	// Selection parameters
	TopKPeers            int           // Number of top peers to select
	ReevaluationInterval time.Duration // How often to re-evaluate peer rankings
	LatencyThreshold     time.Duration // Maximum acceptable latency
	LatencyPenaltyFactor float64       // Factor for latency penalty calculation

	// Back-pressure parameters
	MaxLatencyThreshold  time.Duration // Threshold for back-pressure mechanism
	LatencyIncreaseLimit float64       // Maximum latency increase allowed

	// Historical data parameters
	LatencyHistorySize   int           // Size of latency history buffer
	CongestionHistorySize int          // Size of congestion history buffer
}

// PeerUtilityScore represents a peer with its utility score
type PeerUtilityScore struct {
	PeerID       peer.ID
	UtilityScore float64
	Metrics      *PeerMetrics
	Details      map[string]float64 // Detailed breakdown of score components
}

// DefaultPeerSelectionConfig returns default configuration for peer selection
func DefaultPeerSelectionConfig() *PeerSelectionConfig {
	return &PeerSelectionConfig{
		LatencyWeight:        1.0,
		CongestionWeight:     1.0,
		MemoryWeight:         1.0,
		FLOPsWeight:          1.0,
		TopKPeers:            5,
		ReevaluationInterval: 30 * time.Second,
		LatencyThreshold:     500 * time.Millisecond,
		LatencyPenaltyFactor: 1.5,
		MaxLatencyThreshold:  1000 * time.Millisecond,
		LatencyIncreaseLimit: 1.2,
		LatencyHistorySize:   10,
		CongestionHistorySize: 10,
	}
}

// NewPeerPerformanceTracker creates a new peer performance tracker
func NewPeerPerformanceTracker(config *PeerSelectionConfig) *PeerPerformanceTracker {
	if config == nil {
		config = DefaultPeerSelectionConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tracker := &PeerPerformanceTracker{
		peerMetrics: make(map[peer.ID]*PeerMetrics),
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start periodic re-evaluation
	tracker.wg.Add(1)
	go tracker.periodicReevaluation()

	return tracker
}

// Stop stops the peer performance tracker
func (ppt *PeerPerformanceTracker) Stop() {
	ppt.cancel()
	ppt.wg.Wait()
}

// periodicReevaluation periodically re-evaluates peer rankings
func (ppt *PeerPerformanceTracker) periodicReevaluation() {
	defer ppt.wg.Done()

	ticker := time.NewTicker(ppt.config.ReevaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ppt.ctx.Done():
			return
		case <-ticker.C:
			// Re-evaluate all peer utility scores
			ppt.recalculateAllUtilityScores()
		}
	}
}

// recalculateAllUtilityScores recalculates utility scores for all tracked peers
func (ppt *PeerPerformanceTracker) recalculateAllUtilityScores() {
	ppt.metricsMux.Lock()
	defer ppt.metricsMux.Unlock()

	for _, metrics := range ppt.peerMetrics {
		ppt.calculateUtilityScore(metrics)
	}
}

// UpdatePeerMetrics updates metrics for a peer
func (ppt *PeerPerformanceTracker) UpdatePeerMetrics(peerID peer.ID, metrics *PeerMetrics) {
	ppt.metricsMux.Lock()
	defer ppt.metricsMux.Unlock()

	// Update or create metrics entry
	if existing, exists := ppt.peerMetrics[peerID]; exists {
		// Update existing metrics
		existing.EffectiveFLOPs = metrics.EffectiveFLOPs
		existing.AvailableMemory = metrics.AvailableMemory
		existing.TotalMemory = metrics.TotalMemory
		existing.ActiveTasks = metrics.ActiveTasks
		existing.QueueSize = metrics.QueueSize
		existing.SuccessRate = metrics.SuccessRate
		existing.TotalTasks = metrics.TotalTasks
		existing.SuccessfulTasks = metrics.SuccessfulTasks
		existing.FailedTasks = metrics.FailedTasks
		existing.LastSeen = metrics.LastSeen
		existing.LastTaskComplete = metrics.LastTaskComplete

		// Update latency history
		if metrics.Latency > 0 {
			existing.LatencyHistory = append(existing.LatencyHistory, metrics.Latency)
			if len(existing.LatencyHistory) > ppt.config.LatencyHistorySize {
				existing.LatencyHistory = existing.LatencyHistory[1:]
			}
			existing.Latency = ppt.calculateAverageLatency(existing.LatencyHistory)
			existing.LatencyPenalty = ppt.calculateLatencyPenalty(existing.Latency)
		}

		// Update congestion history
		existing.CongestionHistory = append(existing.CongestionHistory, metrics.EstimatedCongestion)
		if len(existing.CongestionHistory) > ppt.config.CongestionHistorySize {
			existing.CongestionHistory = existing.CongestionHistory[1:]
		}
		existing.EstimatedCongestion = ppt.calculateAverageCongestion(existing.CongestionHistory)

		// Recalculate utility score
		ppt.calculateUtilityScore(existing)
	} else {
		// Create new metrics entry
		newMetrics := &PeerMetrics{
			EffectiveFLOPs:       metrics.EffectiveFLOPs,
			AvailableMemory:      metrics.AvailableMemory,
			TotalMemory:          metrics.TotalMemory,
			Latency:              metrics.Latency,
			LatencyHistory:       []time.Duration{metrics.Latency},
			ActiveTasks:          metrics.ActiveTasks,
			QueueSize:            metrics.QueueSize,
			EstimatedCongestion:  metrics.EstimatedCongestion,
			CongestionHistory:    []float64{metrics.EstimatedCongestion},
			SuccessRate:          metrics.SuccessRate,
			TotalTasks:           metrics.TotalTasks,
			SuccessfulTasks:      metrics.SuccessfulTasks,
			FailedTasks:          metrics.FailedTasks,
			LastSeen:             metrics.LastSeen,
			LastTaskComplete:     metrics.LastTaskComplete,
			LastUpdated:          time.Now(),
		}

		// Calculate initial values
		newMetrics.LatencyPenalty = ppt.calculateLatencyPenalty(newMetrics.Latency)
		ppt.calculateUtilityScore(newMetrics)

		ppt.peerMetrics[peerID] = newMetrics
	}
}

// calculateAverageLatency calculates average latency from history
func (ppt *PeerPerformanceTracker) calculateAverageLatency(history []time.Duration) time.Duration {
	if len(history) == 0 {
		return 0
	}

	var total time.Duration
	for _, latency := range history {
		total += latency
	}

	return time.Duration(int64(total) / int64(len(history)))
}

// calculateAverageCongestion calculates average congestion from history
func (ppt *PeerPerformanceTracker) calculateAverageCongestion(history []float64) float64 {
	if len(history) == 0 {
		return 0.0
	}

	var total float64
	for _, congestion := range history {
		total += congestion
	}

	return total / float64(len(history))
}

// calculateLatencyPenalty calculates latency penalty based on configured factor
func (ppt *PeerPerformanceTracker) calculateLatencyPenalty(latency time.Duration) float64 {
	// Convert latency to milliseconds for calculation
	latencyMs := float64(latency.Milliseconds())

	// Apply penalty factor (exponential penalty for high latency)
	penalty := math.Pow(latencyMs/100.0, ppt.config.LatencyPenaltyFactor)

	// Ensure minimum penalty of 1.0
	if penalty < 1.0 {
		penalty = 1.0
	}

	return penalty
}

// calculateUtilityScore calculates the utility score for a peer using the formula:
// U = (effective_FLOPs × available_memory) / (latency_penalty × estimated_congestion)
func (ppt *PeerPerformanceTracker) calculateUtilityScore(metrics *PeerMetrics) {
	// Calculate numerator: effective_FLOPs × available_memory
	numerator := metrics.EffectiveFLOPs * float64(metrics.AvailableMemory)

	// Calculate denominator: latency_penalty × estimated_congestion
	denominator := metrics.LatencyPenalty * metrics.EstimatedCongestion

	// Avoid division by zero
	if denominator <= 0 {
		denominator = 1.0
	}

	// Calculate utility score
	metrics.LastUtilityScore = numerator / denominator
	metrics.LastUpdated = time.Now()
}

// RecordTaskCompletion records task completion for a peer
func (ppt *PeerPerformanceTracker) RecordTaskCompletion(peerID peer.ID, success bool, processingTime time.Duration, taskSize int64) {
	ppt.metricsMux.Lock()
	defer ppt.metricsMux.Unlock()

	metrics, exists := ppt.peerMetrics[peerID]
	if !exists {
		return
	}

	// Update task counters
	metrics.TotalTasks++
	if success {
		metrics.SuccessfulTasks++
	} else {
		metrics.FailedTasks++
	}

	// Update success rate
	if metrics.TotalTasks > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulTasks) / float64(metrics.TotalTasks)
	}

	// Update FLOPs estimation based on task processing
	if processingTime > 0 && taskSize > 0 {
		// Estimate FLOPs based on task size and processing time
		estimatedFLOPs := float64(taskSize) / processingTime.Seconds()

		// Use exponential moving average to smooth FLOPs estimation
		if metrics.EffectiveFLOPs == 0 {
			metrics.EffectiveFLOPs = estimatedFLOPs
		} else {
			metrics.EffectiveFLOPs = 0.7*metrics.EffectiveFLOPs + 0.3*estimatedFLOPs
		}
	}

	metrics.LastTaskComplete = time.Now()
	metrics.LastUpdated = time.Now()
}

// RecordLatencyMeasurement records a latency measurement for a peer
func (ppt *PeerPerformanceTracker) RecordLatencyMeasurement(peerID peer.ID, latency time.Duration) {
	ppt.metricsMux.Lock()
	defer ppt.metricsMux.Unlock()

	metrics, exists := ppt.peerMetrics[peerID]
	if !exists {
		return
	}

	// Add to latency history
	metrics.LatencyHistory = append(metrics.LatencyHistory, latency)
	if len(metrics.LatencyHistory) > ppt.config.LatencyHistorySize {
		metrics.LatencyHistory = metrics.LatencyHistory[1:]
	}

	// Update average latency
	metrics.Latency = ppt.calculateAverageLatency(metrics.LatencyHistory)

	// Update latency penalty
	metrics.LatencyPenalty = ppt.calculateLatencyPenalty(metrics.Latency)

	// Recalculate utility score
	ppt.calculateUtilityScore(metrics)
	metrics.LastUpdated = time.Now()
}

// selectOptimalPeers selects optimal peers based on utility scoring with dynamic K selection
func (ppt *PeerPerformanceTracker) selectOptimalPeers(availablePeers []peer.ID, network network.Network) ([]peer.ID, error) {
	if len(availablePeers) == 0 {
		return nil, fmt.Errorf("no available peers")
	}

	ppt.metricsMux.RLock()
	defer ppt.metricsMux.RUnlock()

	// Calculate utility scores for all available peers
	peerScores := make([]*PeerUtilityScore, 0, len(availablePeers))

	for _, peerID := range availablePeers {
		// Skip if not connected
		connectedness := network.Connectedness(peerID)
		if connectedness.String() == "NotConnected" {
			continue
		}

		metrics, exists := ppt.peerMetrics[peerID]
		if !exists {
			// Create default metrics for unknown peers
			metrics = &PeerMetrics{
				EffectiveFLOPs:      1.0,
				AvailableMemory:     1024 * 1024 * 1024, // 1GB default
				Latency:             100 * time.Millisecond,
				LatencyPenalty:      1.0,
				EstimatedCongestion: 1.0,
				LastUtilityScore:    1.0,
			}
		}

		// Apply back-pressure mechanism
		if metrics.Latency > ppt.config.MaxLatencyThreshold {
			continue // Skip peers with excessive latency
		}

		// Check if latency has increased beyond limit
		if len(metrics.LatencyHistory) > 1 {
			previousLatency := metrics.LatencyHistory[len(metrics.LatencyHistory)-2]
			if metrics.Latency > time.Duration(float64(previousLatency)*ppt.config.LatencyIncreaseLimit) {
				continue // Skip peers with rapidly increasing latency
			}
		}

		score := &PeerUtilityScore{
			PeerID:       peerID,
			UtilityScore: metrics.LastUtilityScore,
			Metrics:      metrics,
			Details: map[string]float64{
				"effective_flops":      metrics.EffectiveFLOPs,
				"available_memory_gb":  float64(metrics.AvailableMemory) / (1024 * 1024 * 1024),
				"latency_ms":           float64(metrics.Latency.Milliseconds()),
				"latency_penalty":      metrics.LatencyPenalty,
				"estimated_congestion": metrics.EstimatedCongestion,
			},
		}

		peerScores = append(peerScores, score)
	}

	if len(peerScores) == 0 {
		return nil, fmt.Errorf("no suitable peers found after filtering")
	}

	// Sort by utility score (highest first)
	sort.Slice(peerScores, func(i, j int) bool {
		return peerScores[i].UtilityScore > peerScores[j].UtilityScore
	})

	// Dynamic K selection: select top-K peers that yield net speedup
	k := ppt.config.TopKPeers
	if k > len(peerScores) {
		k = len(peerScores)
	}

	// Apply dynamic K selection based on diminishing returns
	k = ppt.calculateOptimalK(peerScores, k)

	// Extract peer IDs
	selectedPeers := make([]peer.ID, 0, k)
	for i := 0; i < k; i++ {
		selectedPeers = append(selectedPeers, peerScores[i].PeerID)
	}

	return selectedPeers, nil
}

// calculateOptimalK calculates the optimal number of peers to select based on diminishing returns
func (ppt *PeerPerformanceTracker) calculateOptimalK(peerScores []*PeerUtilityScore, maxK int) int {
	if len(peerScores) <= 1 {
		return len(peerScores)
	}

	// Start with at least 1 peer
	optimalK := 1

	// Calculate diminishing returns threshold (5% improvement)
	threshold := 0.05

	for i := 1; i < maxK && i < len(peerScores); i++ {
		// Calculate improvement from adding this peer
		currentScore := peerScores[i-1].UtilityScore
		nextScore := peerScores[i].UtilityScore

		if currentScore > 0 {
			improvement := (currentScore - nextScore) / currentScore

			// If improvement is above threshold, continue adding peers
			if improvement <= threshold {
				break // Diminishing returns, stop here
			}
		}

		optimalK = i + 1
	}

	return optimalK
}

// GetPeerMetrics returns metrics for a specific peer
func (ppt *PeerPerformanceTracker) GetPeerMetrics(peerID peer.ID) *PeerMetrics {
	ppt.metricsMux.RLock()
	defer ppt.metricsMux.RUnlock()

	if metrics, exists := ppt.peerMetrics[peerID]; exists {
		// Return a copy to avoid race conditions
		return &PeerMetrics{
			EffectiveFLOPs:       metrics.EffectiveFLOPs,
			AvailableMemory:      metrics.AvailableMemory,
			TotalMemory:          metrics.TotalMemory,
			Latency:              metrics.Latency,
			LatencyHistory:       append([]time.Duration(nil), metrics.LatencyHistory...),
			LatencyPenalty:       metrics.LatencyPenalty,
			ActiveTasks:          metrics.ActiveTasks,
			QueueSize:            metrics.QueueSize,
			EstimatedCongestion:  metrics.EstimatedCongestion,
			CongestionHistory:    append([]float64(nil), metrics.CongestionHistory...),
			LastUtilityScore:     metrics.LastUtilityScore,
			LastUpdated:          metrics.LastUpdated,
			SuccessRate:          metrics.SuccessRate,
			TotalTasks:           metrics.TotalTasks,
			SuccessfulTasks:      metrics.SuccessfulTasks,
			FailedTasks:          metrics.FailedTasks,
			LastSeen:             metrics.LastSeen,
			LastTaskComplete:     metrics.LastTaskComplete,
		}
	}

	return nil
}

// GetAllPeerScores returns utility scores for all tracked peers
func (ppt *PeerPerformanceTracker) GetAllPeerScores() []*PeerUtilityScore {
	ppt.metricsMux.RLock()
	defer ppt.metricsMux.RUnlock()

	scores := make([]*PeerUtilityScore, 0, len(ppt.peerMetrics))

	for peerID, metrics := range ppt.peerMetrics {
		score := &PeerUtilityScore{
			PeerID:       peerID,
			UtilityScore: metrics.LastUtilityScore,
			Metrics:      metrics,
			Details: map[string]float64{
				"effective_flops":      metrics.EffectiveFLOPs,
				"available_memory_gb":  float64(metrics.AvailableMemory) / (1024 * 1024 * 1024),
				"latency_ms":           float64(metrics.Latency.Milliseconds()),
				"latency_penalty":      metrics.LatencyPenalty,
				"estimated_congestion": metrics.EstimatedCongestion,
			},
		}
		scores = append(scores, score)
	}

	// Sort by utility score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].UtilityScore > scores[j].UtilityScore
	})

	return scores
}

// RemovePeer removes tracking for a peer
func (ppt *PeerPerformanceTracker) RemovePeer(peerID peer.ID) {
	ppt.metricsMux.Lock()
	defer ppt.metricsMux.Unlock()

	delete(ppt.peerMetrics, peerID)
}