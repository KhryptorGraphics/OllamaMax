package models

import (
	"context"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// BandwidthAwareStrategy optimizes for bandwidth efficiency
type BandwidthAwareStrategy struct {
	name     string
	priority int
	logger   *slog.Logger
}

func NewBandwidthAwareStrategy(logger *slog.Logger) *BandwidthAwareStrategy {
	return &BandwidthAwareStrategy{
		name:     "bandwidth_aware",
		priority: 80,
		logger:   logger,
	}
}

func (bas *BandwidthAwareStrategy) Optimize(ctx context.Context, task *IntelligentSyncTask, peers []peer.ID) (*OptimizationResult, error) {
	// Sort peers by bandwidth (highest first)
	sortedPeers := make([]peer.ID, len(peers))
	copy(sortedPeers, peers)
	
	// For now, use a simple heuristic - in real implementation, we'd use actual bandwidth data
	// Limit to top bandwidth peers
	maxPeers := int(math.Min(float64(len(sortedPeers)), 3))
	optimizedPeers := sortedPeers[:maxPeers]
	
	// Optimize for bandwidth efficiency
	chunkSize := int64(2 * 1024 * 1024) // 2MB chunks for bandwidth efficiency
	parallelism := maxPeers             // One connection per peer
	compressionLevel := 9               // High compression to save bandwidth
	
	retryStrategy := &RetryStrategy{
		MaxRetries:    3,
		InitialDelay:  2 * time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      30 * time.Second,
		JitterEnabled: true,
	}
	
	estimatedTime := time.Duration(float64(task.Options.ChunkSize) / (1024 * 1024)) * time.Second // Rough estimate
	
	return &OptimizationResult{
		OptimizedPeers:   optimizedPeers,
		ChunkSize:        chunkSize,
		Parallelism:      parallelism,
		CompressionLevel: compressionLevel,
		RetryStrategy:    retryStrategy,
		EstimatedTime:    estimatedTime,
		EstimatedCost:    float64(len(optimizedPeers)) * 0.1, // Cost per peer
		Metadata: map[string]interface{}{
			"strategy":     bas.name,
			"optimization": "bandwidth_efficiency",
			"peer_count":   len(optimizedPeers),
		},
	}, nil
}

func (bas *BandwidthAwareStrategy) GetName() string { return bas.name }
func (bas *BandwidthAwareStrategy) GetPriority() int { return bas.priority }

// LatencyOptimizedStrategy optimizes for low latency
type LatencyOptimizedStrategy struct {
	name     string
	priority int
	logger   *slog.Logger
}

func NewLatencyOptimizedStrategy(logger *slog.Logger) *LatencyOptimizedStrategy {
	return &LatencyOptimizedStrategy{
		name:     "latency_optimized",
		priority: 85,
		logger:   logger,
	}
}

func (los *LatencyOptimizedStrategy) Optimize(ctx context.Context, task *IntelligentSyncTask, peers []peer.ID) (*OptimizationResult, error) {
	// Sort peers by latency (lowest first)
	sortedPeers := make([]peer.ID, len(peers))
	copy(sortedPeers, peers)
	
	// Select low-latency peers
	maxPeers := int(math.Min(float64(len(sortedPeers)), 2))
	optimizedPeers := sortedPeers[:maxPeers]
	
	// Optimize for low latency
	chunkSize := int64(512 * 1024)    // Smaller chunks for faster response
	parallelism := maxPeers * 2       // More parallel connections
	compressionLevel := 3             // Lower compression for speed
	
	retryStrategy := &RetryStrategy{
		MaxRetries:    5,
		InitialDelay:  500 * time.Millisecond,
		BackoffFactor: 1.5,
		MaxDelay:      5 * time.Second,
		JitterEnabled: true,
	}
	
	estimatedTime := time.Duration(float64(task.Options.ChunkSize) / (2 * 1024 * 1024)) * time.Second
	
	return &OptimizationResult{
		OptimizedPeers:   optimizedPeers,
		ChunkSize:        chunkSize,
		Parallelism:      parallelism,
		CompressionLevel: compressionLevel,
		RetryStrategy:    retryStrategy,
		EstimatedTime:    estimatedTime,
		EstimatedCost:    float64(len(optimizedPeers)) * 0.15, // Higher cost for speed
		Metadata: map[string]interface{}{
			"strategy":     los.name,
			"optimization": "latency_minimization",
			"peer_count":   len(optimizedPeers),
		},
	}, nil
}

func (los *LatencyOptimizedStrategy) GetName() string { return los.name }
func (los *LatencyOptimizedStrategy) GetPriority() int { return los.priority }

// ReliabilityFocusedStrategy optimizes for reliability and fault tolerance
type ReliabilityFocusedStrategy struct {
	name     string
	priority int
	logger   *slog.Logger
}

func NewReliabilityFocusedStrategy(logger *slog.Logger) *ReliabilityFocusedStrategy {
	return &ReliabilityFocusedStrategy{
		name:     "reliability_focused",
		priority: 90,
		logger:   logger,
	}
}

func (rfs *ReliabilityFocusedStrategy) Optimize(ctx context.Context, task *IntelligentSyncTask, peers []peer.ID) (*OptimizationResult, error) {
	// Sort peers by reliability (highest first)
	sortedPeers := make([]peer.ID, len(peers))
	copy(sortedPeers, peers)
	
	// Select more peers for redundancy
	maxPeers := int(math.Min(float64(len(sortedPeers)), 5))
	optimizedPeers := sortedPeers[:maxPeers]
	
	// Optimize for reliability
	chunkSize := int64(1024 * 1024)   // Medium chunks for balance
	parallelism := 2                  // Conservative parallelism
	compressionLevel := 6             // Balanced compression
	
	retryStrategy := &RetryStrategy{
		MaxRetries:    10,
		InitialDelay:  1 * time.Second,
		BackoffFactor: 1.8,
		MaxDelay:      60 * time.Second,
		JitterEnabled: true,
	}
	
	estimatedTime := time.Duration(float64(task.Options.ChunkSize) / (1024 * 1024)) * time.Second * 2 // Conservative estimate
	
	return &OptimizationResult{
		OptimizedPeers:   optimizedPeers,
		ChunkSize:        chunkSize,
		Parallelism:      parallelism,
		CompressionLevel: compressionLevel,
		RetryStrategy:    retryStrategy,
		EstimatedTime:    estimatedTime,
		EstimatedCost:    float64(len(optimizedPeers)) * 0.08, // Lower cost for reliability
		Metadata: map[string]interface{}{
			"strategy":     rfs.name,
			"optimization": "reliability_maximization",
			"peer_count":   len(optimizedPeers),
			"redundancy":   true,
		},
	}, nil
}

func (rfs *ReliabilityFocusedStrategy) GetName() string { return rfs.name }
func (rfs *ReliabilityFocusedStrategy) GetPriority() int { return rfs.priority }

// AdaptiveStrategy adapts based on current conditions and historical performance
type AdaptiveStrategy struct {
	name     string
	priority int
	logger   *slog.Logger
}

func NewAdaptiveStrategy(logger *slog.Logger) *AdaptiveStrategy {
	return &AdaptiveStrategy{
		name:     "adaptive",
		priority: 95,
		logger:   logger,
	}
}

func (as *AdaptiveStrategy) Optimize(ctx context.Context, task *IntelligentSyncTask, peers []peer.ID) (*OptimizationResult, error) {
	// Analyze task characteristics
	taskSize := task.Options.ChunkSize
	taskPriority := task.Priority
	
	// Adaptive peer selection based on task characteristics
	var maxPeers int
	var chunkSize int64
	var parallelism int
	var compressionLevel int
	
	switch {
	case taskSize > 100*1024*1024: // Large tasks (>100MB)
		maxPeers = int(math.Min(float64(len(peers)), 4))
		chunkSize = 4 * 1024 * 1024 // 4MB chunks
		parallelism = maxPeers
		compressionLevel = 7
	case taskSize > 10*1024*1024: // Medium tasks (10-100MB)
		maxPeers = int(math.Min(float64(len(peers)), 3))
		chunkSize = 2 * 1024 * 1024 // 2MB chunks
		parallelism = maxPeers
		compressionLevel = 6
	default: // Small tasks (<10MB)
		maxPeers = int(math.Min(float64(len(peers)), 2))
		chunkSize = 1 * 1024 * 1024 // 1MB chunks
		parallelism = maxPeers
		compressionLevel = 5
	}
	
	// Adjust based on priority
	if taskPriority > 5 {
		parallelism = int(float64(parallelism) * 1.5) // Increase parallelism for high priority
		compressionLevel = int(math.Max(float64(compressionLevel-2), 1)) // Reduce compression for speed
	}
	
	// Select peers (in real implementation, this would use actual performance data)
	sortedPeers := make([]peer.ID, len(peers))
	copy(sortedPeers, peers)
	optimizedPeers := sortedPeers[:maxPeers]
	
	retryStrategy := &RetryStrategy{
		MaxRetries:    5,
		InitialDelay:  1 * time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      30 * time.Second,
		JitterEnabled: true,
	}
	
	// Estimate time based on adaptive parameters
	estimatedBandwidth := float64(1024 * 1024) // 1MB/s baseline
	estimatedTime := time.Duration(float64(taskSize) / estimatedBandwidth) * time.Second
	
	return &OptimizationResult{
		OptimizedPeers:   optimizedPeers,
		ChunkSize:        chunkSize,
		Parallelism:      parallelism,
		CompressionLevel: compressionLevel,
		RetryStrategy:    retryStrategy,
		EstimatedTime:    estimatedTime,
		EstimatedCost:    float64(len(optimizedPeers)) * 0.12,
		Metadata: map[string]interface{}{
			"strategy":      as.name,
			"optimization":  "adaptive_balanced",
			"peer_count":    len(optimizedPeers),
			"task_size":     taskSize,
			"task_priority": taskPriority,
			"adaptation":    "size_and_priority_based",
		},
	}, nil
}

func (as *AdaptiveStrategy) GetName() string { return as.name }
func (as *AdaptiveStrategy) GetPriority() int { return as.priority }

// PeerSorter helps sort peers by different criteria
type PeerSorter struct {
	peers   []peer.ID
	metrics map[peer.ID]*PeerPerformanceMetrics
}

func NewPeerSorter(peers []peer.ID, metrics map[peer.ID]*PeerPerformanceMetrics) *PeerSorter {
	return &PeerSorter{
		peers:   peers,
		metrics: metrics,
	}
}

func (ps *PeerSorter) SortByBandwidth() []peer.ID {
	sorted := make([]peer.ID, len(ps.peers))
	copy(sorted, ps.peers)
	
	sort.Slice(sorted, func(i, j int) bool {
		metricsI := ps.metrics[sorted[i]]
		metricsJ := ps.metrics[sorted[j]]
		
		if metricsI == nil && metricsJ == nil {
			return false
		}
		if metricsI == nil {
			return false
		}
		if metricsJ == nil {
			return true
		}
		
		return metricsI.Bandwidth > metricsJ.Bandwidth
	})
	
	return sorted
}

func (ps *PeerSorter) SortByLatency() []peer.ID {
	sorted := make([]peer.ID, len(ps.peers))
	copy(sorted, ps.peers)
	
	sort.Slice(sorted, func(i, j int) bool {
		metricsI := ps.metrics[sorted[i]]
		metricsJ := ps.metrics[sorted[j]]
		
		if metricsI == nil && metricsJ == nil {
			return false
		}
		if metricsI == nil {
			return false
		}
		if metricsJ == nil {
			return true
		}
		
		return metricsI.AverageLatency < metricsJ.AverageLatency
	})
	
	return sorted
}

func (ps *PeerSorter) SortByReliability() []peer.ID {
	sorted := make([]peer.ID, len(ps.peers))
	copy(sorted, ps.peers)
	
	sort.Slice(sorted, func(i, j int) bool {
		metricsI := ps.metrics[sorted[i]]
		metricsJ := ps.metrics[sorted[j]]
		
		if metricsI == nil && metricsJ == nil {
			return false
		}
		if metricsI == nil {
			return false
		}
		if metricsJ == nil {
			return true
		}
		
		return metricsI.Reliability > metricsJ.Reliability
	})
	
	return sorted
}
