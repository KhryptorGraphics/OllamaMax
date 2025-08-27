package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/libp2p/go-libp2p/core/peer"
)

// SyncOptimizer optimizes synchronization operations for performance and efficiency
type SyncOptimizer struct {
	mu sync.RWMutex

	config *config.SyncConfig
	logger *slog.Logger

	// Performance tracking
	peerPerformance   map[peer.ID]*PeerPerformanceMetrics
	networkConditions *NetworkConditions

	// Optimization strategies
	strategies map[string]OptimizationStrategy

	// Adaptive parameters
	adaptiveParams *AdaptiveParameters

	// Historical data
	syncHistory    []*SyncHistoryEntry
	maxHistorySize int
}

// PeerPerformanceMetrics tracks performance metrics for each peer
type PeerPerformanceMetrics struct {
	PeerID           peer.ID       `json:"peer_id"`
	AverageLatency   time.Duration `json:"average_latency"`
	Bandwidth        int64         `json:"bandwidth"`
	Reliability      float64       `json:"reliability"`
	SuccessRate      float64       `json:"success_rate"`
	LastSyncTime     time.Time     `json:"last_sync_time"`
	TotalSyncs       int64         `json:"total_syncs"`
	FailedSyncs      int64         `json:"failed_syncs"`
	BytesTransferred int64         `json:"bytes_transferred"`

	// Recent performance window
	RecentLatencies  []time.Duration `json:"recent_latencies"`
	RecentBandwidths []int64         `json:"recent_bandwidths"`
	WindowSize       int             `json:"window_size"`
}

// NetworkConditions represents current network conditions
type NetworkConditions struct {
	AverageBandwidth int64         `json:"average_bandwidth"`
	AverageLatency   time.Duration `json:"average_latency"`
	PacketLoss       float64       `json:"packet_loss"`
	Congestion       float64       `json:"congestion"`
	PeakHours        bool          `json:"peak_hours"`
	NetworkQuality   string        `json:"network_quality"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// OptimizationStrategy interface for different optimization approaches
type OptimizationStrategy interface {
	Optimize(ctx context.Context, task *IntelligentSyncTask, peers []peer.ID) (*OptimizationResult, error)
	GetName() string
	GetPriority() int
}

// OptimizationResult contains the result of optimization
type OptimizationResult struct {
	OptimizedPeers   []peer.ID              `json:"optimized_peers"`
	ChunkSize        int64                  `json:"chunk_size"`
	Parallelism      int                    `json:"parallelism"`
	CompressionLevel int                    `json:"compression_level"`
	RetryStrategy    *RetryStrategy         `json:"retry_strategy"`
	Metadata         map[string]interface{} `json:"metadata"`
	EstimatedTime    time.Duration          `json:"estimated_time"`
	EstimatedCost    float64                `json:"estimated_cost"`
}

// RetryStrategy defines how to handle retries
type RetryStrategy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	MaxDelay      time.Duration `json:"max_delay"`
	JitterEnabled bool          `json:"jitter_enabled"`
}

// AdaptiveParameters contains parameters that adapt based on conditions
type AdaptiveParameters struct {
	ChunkSize         int64     `json:"chunk_size"`
	Parallelism       int       `json:"parallelism"`
	CompressionLevel  int       `json:"compression_level"`
	TimeoutMultiplier float64   `json:"timeout_multiplier"`
	LastUpdated       time.Time `json:"last_updated"`
}

// SyncHistoryEntry records historical sync performance
type SyncHistoryEntry struct {
	Timestamp        time.Time     `json:"timestamp"`
	ModelName        string        `json:"model_name"`
	PeerID           peer.ID       `json:"peer_id"`
	Duration         time.Duration `json:"duration"`
	BytesTransferred int64         `json:"bytes_transferred"`
	Success          bool          `json:"success"`
	ChunkSize        int64         `json:"chunk_size"`
	Parallelism      int           `json:"parallelism"`
	NetworkQuality   string        `json:"network_quality"`
}

// NewSyncOptimizer creates a new sync optimizer
func NewSyncOptimizer(config *config.SyncConfig, logger *slog.Logger) *SyncOptimizer {
	optimizer := &SyncOptimizer{
		config:            config,
		logger:            logger,
		peerPerformance:   make(map[peer.ID]*PeerPerformanceMetrics),
		networkConditions: &NetworkConditions{},
		strategies:        make(map[string]OptimizationStrategy),
		adaptiveParams: &AdaptiveParameters{
			ChunkSize:         1024 * 1024, // 1MB default
			Parallelism:       4,
			CompressionLevel:  6,
			TimeoutMultiplier: 1.0,
			LastUpdated:       time.Now(),
		},
		maxHistorySize: 1000,
	}

	// Initialize optimization strategies
	optimizer.initializeStrategies()

	return optimizer
}

// initializeStrategies sets up the optimization strategies
func (so *SyncOptimizer) initializeStrategies() {
	so.strategies["bandwidth_aware"] = NewBandwidthAwareStrategy(so.logger)
	so.strategies["latency_optimized"] = NewLatencyOptimizedStrategy(so.logger)
	so.strategies["reliability_focused"] = NewReliabilityFocusedStrategy(so.logger)
	so.strategies["adaptive"] = NewAdaptiveStrategy(so.logger)
}

// OptimizeSync optimizes a synchronization task
func (so *SyncOptimizer) OptimizeSync(ctx context.Context, task *IntelligentSyncTask, availablePeers []peer.ID) (*OptimizationResult, error) {
	so.mu.Lock()
	defer so.mu.Unlock()

	// Update network conditions
	so.updateNetworkConditions()

	// Select best optimization strategy based on current conditions
	strategy := so.selectOptimizationStrategy(task, availablePeers)

	// Apply optimization
	result, err := strategy.Optimize(ctx, task, availablePeers)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	// Update adaptive parameters based on result
	so.updateAdaptiveParameters(result)

	so.logger.Info("sync optimization completed",
		"strategy", strategy.GetName(),
		"peers", len(result.OptimizedPeers),
		"chunk_size", result.ChunkSize,
		"parallelism", result.Parallelism,
		"estimated_time", result.EstimatedTime)

	return result, nil
}

// UpdatePeerPerformance updates performance metrics for a peer
func (so *SyncOptimizer) UpdatePeerPerformance(peerID peer.ID, latency time.Duration, bandwidth int64, success bool, bytesTransferred int64) {
	so.mu.Lock()
	defer so.mu.Unlock()

	metrics, exists := so.peerPerformance[peerID]
	if !exists {
		metrics = &PeerPerformanceMetrics{
			PeerID:           peerID,
			WindowSize:       10,
			RecentLatencies:  make([]time.Duration, 0, 10),
			RecentBandwidths: make([]int64, 0, 10),
		}
		so.peerPerformance[peerID] = metrics
	}

	// Update counters
	metrics.TotalSyncs++
	if !success {
		metrics.FailedSyncs++
	}
	metrics.BytesTransferred += bytesTransferred
	metrics.LastSyncTime = time.Now()

	// Update recent performance windows
	metrics.RecentLatencies = append(metrics.RecentLatencies, latency)
	if len(metrics.RecentLatencies) > metrics.WindowSize {
		metrics.RecentLatencies = metrics.RecentLatencies[1:]
	}

	metrics.RecentBandwidths = append(metrics.RecentBandwidths, bandwidth)
	if len(metrics.RecentBandwidths) > metrics.WindowSize {
		metrics.RecentBandwidths = metrics.RecentBandwidths[1:]
	}

	// Calculate averages
	metrics.AverageLatency = so.calculateAverageLatency(metrics.RecentLatencies)
	metrics.Bandwidth = so.calculateAverageBandwidth(metrics.RecentBandwidths)
	metrics.SuccessRate = float64(metrics.TotalSyncs-metrics.FailedSyncs) / float64(metrics.TotalSyncs)

	// Calculate reliability (combination of success rate and consistency)
	metrics.Reliability = so.calculateReliability(metrics)

	// Record in history
	so.recordSyncHistory(peerID, latency, bandwidth, success, bytesTransferred)
}

// selectOptimizationStrategy selects the best strategy for current conditions
func (so *SyncOptimizer) selectOptimizationStrategy(task *IntelligentSyncTask, peers []peer.ID) OptimizationStrategy {
	// Analyze current conditions
	avgBandwidth := so.networkConditions.AverageBandwidth
	avgLatency := so.networkConditions.AverageLatency
	congestion := so.networkConditions.Congestion

	// Select strategy based on conditions and task type
	switch {
	case congestion > 0.7:
		// High congestion - focus on reliability
		return so.strategies["reliability_focused"]
	case avgLatency > 100*time.Millisecond:
		// High latency - optimize for latency
		return so.strategies["latency_optimized"]
	case avgBandwidth < 1024*1024: // Less than 1MB/s
		// Low bandwidth - optimize for bandwidth
		return so.strategies["bandwidth_aware"]
	default:
		// Normal conditions - use adaptive strategy
		return so.strategies["adaptive"]
	}
}

// updateNetworkConditions updates current network conditions
func (so *SyncOptimizer) updateNetworkConditions() {
	// Calculate average metrics from peer performance
	var totalBandwidth int64
	var totalLatency time.Duration
	var peerCount int

	for _, metrics := range so.peerPerformance {
		if metrics.LastSyncTime.After(time.Now().Add(-5 * time.Minute)) {
			totalBandwidth += metrics.Bandwidth
			totalLatency += metrics.AverageLatency
			peerCount++
		}
	}

	if peerCount > 0 {
		so.networkConditions.AverageBandwidth = totalBandwidth / int64(peerCount)
		so.networkConditions.AverageLatency = totalLatency / time.Duration(peerCount)
	}

	// Estimate congestion based on performance degradation
	so.networkConditions.Congestion = so.estimateCongestion()

	// Determine network quality
	so.networkConditions.NetworkQuality = so.determineNetworkQuality()

	// Check if it's peak hours (simplified heuristic)
	hour := time.Now().Hour()
	so.networkConditions.PeakHours = hour >= 9 && hour <= 17

	so.networkConditions.LastUpdated = time.Now()
}

// estimateCongestion estimates network congestion based on performance trends
func (so *SyncOptimizer) estimateCongestion() float64 {
	if len(so.syncHistory) < 10 {
		return 0.0
	}

	// Look at recent performance vs historical average
	recentEntries := so.syncHistory[len(so.syncHistory)-10:]
	var recentAvgDuration time.Duration
	var recentAvgBandwidth int64

	for _, entry := range recentEntries {
		recentAvgDuration += entry.Duration
		if entry.BytesTransferred > 0 && entry.Duration > 0 {
			bandwidth := entry.BytesTransferred / int64(entry.Duration.Seconds())
			recentAvgBandwidth += bandwidth
		}
	}

	recentAvgDuration /= time.Duration(len(recentEntries))
	recentAvgBandwidth /= int64(len(recentEntries))

	// Compare with historical average
	var historicalAvgDuration time.Duration
	var historicalAvgBandwidth int64
	validEntries := 0

	for _, entry := range so.syncHistory {
		if entry.Success {
			historicalAvgDuration += entry.Duration
			if entry.BytesTransferred > 0 && entry.Duration > 0 {
				bandwidth := entry.BytesTransferred / int64(entry.Duration.Seconds())
				historicalAvgBandwidth += bandwidth
			}
			validEntries++
		}
	}

	if validEntries == 0 {
		return 0.0
	}

	historicalAvgDuration /= time.Duration(validEntries)
	historicalAvgBandwidth /= int64(validEntries)

	// Calculate congestion score (0.0 to 1.0)
	durationRatio := float64(recentAvgDuration) / float64(historicalAvgDuration)
	bandwidthRatio := float64(historicalAvgBandwidth) / float64(recentAvgBandwidth)

	congestion := (durationRatio + bandwidthRatio - 2.0) / 2.0
	if congestion < 0 {
		congestion = 0
	}
	if congestion > 1 {
		congestion = 1
	}

	return congestion
}

// determineNetworkQuality determines overall network quality
func (so *SyncOptimizer) determineNetworkQuality() string {
	avgBandwidth := so.networkConditions.AverageBandwidth
	avgLatency := so.networkConditions.AverageLatency
	congestion := so.networkConditions.Congestion

	score := 0.0

	// Bandwidth score (0-40 points)
	if avgBandwidth >= 10*1024*1024 { // 10MB/s
		score += 40
	} else if avgBandwidth >= 1*1024*1024 { // 1MB/s
		score += 30
	} else if avgBandwidth >= 100*1024 { // 100KB/s
		score += 20
	} else {
		score += 10
	}

	// Latency score (0-30 points)
	if avgLatency <= 50*time.Millisecond {
		score += 30
	} else if avgLatency <= 100*time.Millisecond {
		score += 25
	} else if avgLatency <= 200*time.Millisecond {
		score += 20
	} else {
		score += 10
	}

	// Congestion score (0-30 points)
	score += (1.0 - congestion) * 30

	switch {
	case score >= 80:
		return "excellent"
	case score >= 60:
		return "good"
	case score >= 40:
		return "fair"
	default:
		return "poor"
	}
}

// Helper functions

func (so *SyncOptimizer) calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}
	return total / time.Duration(len(latencies))
}

func (so *SyncOptimizer) calculateAverageBandwidth(bandwidths []int64) int64 {
	if len(bandwidths) == 0 {
		return 0
	}

	var total int64
	for _, bandwidth := range bandwidths {
		total += bandwidth
	}
	return total / int64(len(bandwidths))
}

func (so *SyncOptimizer) calculateReliability(metrics *PeerPerformanceMetrics) float64 {
	if metrics.TotalSyncs == 0 {
		return 0.0
	}

	// Base reliability on success rate
	reliability := metrics.SuccessRate

	// Adjust for consistency (lower variance in latency = higher reliability)
	if len(metrics.RecentLatencies) > 1 {
		variance := so.calculateLatencyVariance(metrics.RecentLatencies)
		consistencyFactor := 1.0 / (1.0 + variance)
		reliability = (reliability + consistencyFactor) / 2.0
	}

	return reliability
}

func (so *SyncOptimizer) calculateLatencyVariance(latencies []time.Duration) float64 {
	if len(latencies) <= 1 {
		return 0.0
	}

	mean := so.calculateAverageLatency(latencies)
	var variance float64

	for _, latency := range latencies {
		diff := float64(latency - mean)
		variance += diff * diff
	}

	return variance / float64(len(latencies))
}

func (so *SyncOptimizer) recordSyncHistory(peerID peer.ID, latency time.Duration, bandwidth int64, success bool, bytesTransferred int64) {
	entry := &SyncHistoryEntry{
		Timestamp:        time.Now(),
		PeerID:           peerID,
		Duration:         latency,
		BytesTransferred: bytesTransferred,
		Success:          success,
		ChunkSize:        so.adaptiveParams.ChunkSize,
		Parallelism:      so.adaptiveParams.Parallelism,
		NetworkQuality:   so.networkConditions.NetworkQuality,
	}

	so.syncHistory = append(so.syncHistory, entry)

	// Limit history size
	if len(so.syncHistory) > so.maxHistorySize {
		so.syncHistory = so.syncHistory[1:]
	}
}

func (so *SyncOptimizer) updateAdaptiveParameters(result *OptimizationResult) {
	// Update parameters based on optimization result
	so.adaptiveParams.ChunkSize = result.ChunkSize
	so.adaptiveParams.Parallelism = result.Parallelism
	so.adaptiveParams.CompressionLevel = result.CompressionLevel
	so.adaptiveParams.LastUpdated = time.Now()
}

// GetPeerPerformance returns performance metrics for a peer
func (so *SyncOptimizer) GetPeerPerformance(peerID peer.ID) *PeerPerformanceMetrics {
	so.mu.RLock()
	defer so.mu.RUnlock()

	if metrics, exists := so.peerPerformance[peerID]; exists {
		return metrics
	}
	return nil
}

// GetNetworkConditions returns current network conditions
func (so *SyncOptimizer) GetNetworkConditions() *NetworkConditions {
	so.mu.RLock()
	defer so.mu.RUnlock()

	return so.networkConditions
}

// GetAdaptiveParameters returns current adaptive parameters
func (so *SyncOptimizer) GetAdaptiveParameters() *AdaptiveParameters {
	so.mu.RLock()
	defer so.mu.RUnlock()

	return so.adaptiveParams
}
