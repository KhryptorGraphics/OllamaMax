package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/network"
)

// Example of how to integrate peer selection into an existing scheduler

// EnhancedScheduler extends a basic scheduler with adaptive peer selection
type EnhancedScheduler struct {
	peerTracker *PeerPerformanceTracker
	// ... other scheduler fields
}

// NewEnhancedScheduler creates a new scheduler with adaptive peer selection
func NewEnhancedScheduler() *EnhancedScheduler {
	return &EnhancedScheduler{
		peerTracker: NewPeerPerformanceTracker(nil), // Use default config
	}
}

// SelectOptimalPeersForTask selects the best peers for a given task
func (s *EnhancedScheduler) SelectOptimalPeersForTask(
	availablePeers []peer.ID,
	net network.Network,
) ([]peer.ID, error) {
	return s.peerTracker.selectOptimalPeers(availablePeers, net)
}

// UpdatePeerMetrics updates performance metrics for a peer
func (s *EnhancedScheduler) UpdatePeerMetrics(peerID peer.ID, metrics *PeerMetrics) {
	s.peerTracker.UpdatePeerMetrics(peerID, metrics)
}

// RecordTaskCompletion records the completion of a task for performance tracking
func (s *EnhancedScheduler) RecordTaskCompletion(
	peerID peer.ID,
	success bool,
	processingTime time.Duration,
	taskSize int64,
) {
	s.peerTracker.RecordTaskCompletion(peerID, success, processingTime, taskSize)
}

// RecordLatencyMeasurement records a latency measurement for a peer
func (s *EnhancedScheduler) RecordLatencyMeasurement(peerID peer.ID, latency time.Duration) {
	s.peerTracker.RecordLatencyMeasurement(peerID, latency)
}

// GetPeerMetrics returns the current metrics for a peer
func (s *EnhancedScheduler) GetPeerMetrics(peerID peer.ID) *PeerMetrics {
	return s.peerTracker.GetPeerMetrics(peerID)
}

// GetAllPeerScores returns utility scores for all tracked peers
func (s *EnhancedScheduler) GetAllPeerScores() []*PeerUtilityScore {
	return s.peerTracker.GetAllPeerScores()
}

// Example of how to use the enhanced scheduler in a task assignment scenario
func (s *EnhancedScheduler) AssignTaskToOptimalPeer(
	task *Task, // Assuming Task is defined elsewhere
	net network.Network,
) (peer.ID, error) {

	// Get all available peers from the network
	allPeers := net.Peers()

	// Select optimal peers using utility scoring
	selectedPeers, err := s.SelectOptimalPeersForTask(allPeers, net)
	if err != nil {
		return "", fmt.Errorf("failed to select optimal peers: %w", err)
	}

	if len(selectedPeers) == 0 {
		return "", fmt.Errorf("no suitable peers available")
	}

	// Return the best peer (first in the sorted list)
	return selectedPeers[0], nil
}

// Example of periodic metrics update from a monitoring system
func (s *EnhancedScheduler) UpdatePeerMetricsFromMonitoring() {
	// This would typically be called periodically to update peer metrics
	// based on actual performance data

	// Example: Update metrics for a peer based on monitoring data
	/*
	peerMetrics := &PeerMetrics{
		EffectiveFLOPs:      measuredFLOPs,
		AvailableMemory:     availableMemory,
		Latency:             measuredLatency,
		EstimatedCongestion: congestionEstimate,
		SuccessRate:         calculatedSuccessRate,
		// ... other metrics
	}
	s.UpdatePeerMetrics(peerID, peerMetrics)
	*/
}

// Integration with P2P discovery system
// This shows how to extend the OptimizedBootstrapDiscovery to use utility scoring

// ExtendedOptimizedBootstrapDiscovery adds utility scoring to the existing discovery
type ExtendedOptimizedBootstrapDiscovery struct {
	// Embed the existing discovery
	*OptimizedBootstrapDiscovery

	// Add peer performance tracking
	peerTracker *PeerPerformanceTracker
}

// NewExtendedOptimizedBootstrapDiscovery creates an extended discovery with utility scoring
func NewExtendedOptimizedBootstrapDiscovery(
	host network.Host,
	bootstrapPeers []peer.AddrInfo,
	minPeers, maxPeers int,
) *ExtendedOptimizedBootstrapDiscovery {

	baseDiscovery := NewOptimizedBootstrapDiscovery(host, bootstrapPeers, minPeers, maxPeers)

	return &ExtendedOptimizedBootstrapDiscovery{
		OptimizedBootstrapDiscovery: baseDiscovery,
		peerTracker:                NewPeerPerformanceTracker(nil),
	}
}

// selectOptimalPeers overrides the base implementation to use utility scoring
func (e *ExtendedOptimizedBootstrapDiscovery) selectOptimalPeers(peers []peer.AddrInfo) []peer.AddrInfo {
	if len(peers) == 0 {
		return peers
	}

	// Convert peer.AddrInfo to peer.ID for selection
	peerIDs := make([]peer.ID, len(peers))
	for i, peerInfo := range peers {
		peerIDs[i] = peerInfo.ID
	}

	// Use utility scoring to select optimal peers
	selectedPeerIDs, err := e.peerTracker.selectOptimalPeers(peerIDs, e.host.Network())
	if err != nil {
		// Fallback to original selection method
		return e.OptimizedBootstrapDiscovery.selectOptimalPeers(peers)
	}

	// Convert back to peer.AddrInfo
	selectedPeers := make([]peer.AddrInfo, len(selectedPeerIDs))
	selectedIndex := 0

	for _, peerInfo := range peers {
		for _, selectedID := range selectedPeerIDs {
			if peerInfo.ID == selectedID {
				selectedPeers[selectedIndex] = peerInfo
				selectedIndex++
				break
			}
		}
	}

	return selectedPeers
}

// UpdatePeerMetricsForDiscovery updates peer metrics from the discovery system
func (e *ExtendedOptimizedBootstrapDiscovery) UpdatePeerMetricsForDiscovery(peerID peer.ID) {
	// Get connection info from the base discovery
	connInfo := e.GetConnectionInfo(peerID)
	if connInfo == nil {
		return
	}

	// Convert OptimizedConnectionInfo to PeerMetrics
	peerMetrics := &PeerMetrics{
		EffectiveFLOPs:      connInfo.EffectiveFLOPs,
		AvailableMemory:     connInfo.AvailableMemory,
		ActiveTasks:         connInfo.ActiveTasks,
		QueueSize:           connInfo.QueueSize,
		EstimatedCongestion: connInfo.EstimatedCongestion,
		SuccessRate:         connInfo.SuccessRate,
		TotalTasks:          int64(connInfo.Attempts),
		SuccessfulTasks:     int64(connInfo.Attempts - connInfo.Failures),
		FailedTasks:         int64(connInfo.Failures),
		Latency:             connInfo.RTT,
		LatencyPenalty:      e.calculateLatencyPenaltyForMetrics(connInfo.RTT),
	}

	e.peerTracker.UpdatePeerMetrics(peerID, peerMetrics)
}

// calculateLatencyPenaltyForMetrics calculates latency penalty for metrics conversion
func (e *ExtendedOptimizedBootstrapDiscovery) calculateLatencyPenaltyForMetrics(latency time.Duration) float64 {
	// Convert RTT to latency penalty using the same formula as PeerPerformanceTracker
	latencyMs := float64(latency.Milliseconds())
	penalty := 1.0

	if latencyMs > 0 {
		penalty = 1.0 + (latencyMs / 100.0) // Simple linear penalty
	}

	return penalty
}