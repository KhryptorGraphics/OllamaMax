package discovery

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// OptimizedBootstrapDiscovery implements optimized bootstrap discovery with parallel connection attempts
type OptimizedBootstrapDiscovery struct {
	host            host.Host
	bootstrapPeers  []peer.AddrInfo
	minPeers        int
	maxPeers        int
	
	// Connection tracking with optimization
	connections     map[peer.ID]*OptimizedConnectionInfo
	connectionsMux  sync.RWMutex
	
	// Performance optimization
	config          *OptimizedDiscoveryConfig
	metrics         *OptimizedDiscoveryMetrics
}

// OptimizedConnectionInfo tracks optimized connection information
type OptimizedConnectionInfo struct {
	ConnectedAt       time.Time
	LastSeen          time.Time
	Attempts          int
	Failures          int
	LastBackoff       time.Duration
	
	// Performance tracking
	RTT               time.Duration
	SuccessRate       float64
	Priority          int
	
	// Connection state
	IsConnecting      bool
	ConnectionCtx     context.Context
	ConnectionCancel  context.CancelFunc
}

// OptimizedDiscoveryConfig configures optimized discovery behavior
type OptimizedDiscoveryConfig struct {
	// Connection optimization
	ConnectTimeout       time.Duration  // Reduced from 30s to 5s
	ParallelAttempts     int           // Number of parallel connection attempts
	EarlySuccessDelay    time.Duration // Delay for early success detection
	
	// Backoff configuration
	BackoffInitial       time.Duration
	BackoffMax           time.Duration
	BackoffMultiplier    float64
	
	// Discovery optimization
	DiscoveryInterval    time.Duration
	HealthCheckInterval  time.Duration
	PeerSelectionStrategy string
	
	// Performance thresholds
	MaxFailuresBeforeBackoff int
	RTTThreshold            time.Duration
	SuccessRateThreshold    float64
}

// OptimizedDiscoveryMetrics tracks discovery performance
type OptimizedDiscoveryMetrics struct {
	// Connection metrics
	TotalAttempts         int64
	SuccessfulConnections int64
	FailedConnections     int64
	ParallelConnections   int64
	EarlySuccesses        int64
	
	// Performance metrics
	AverageRTT           time.Duration
	AverageConnectTime   time.Duration
	BackoffRetries       int64
	
	// Timing metrics
	LastDiscovery        time.Time
	LastSuccessfulConnect time.Time
	
	// Optimization metrics
	TimeoutReductions    int64
	ParallelEfficiency   float64
}

// DefaultOptimizedDiscoveryConfig returns optimized default configuration
func DefaultOptimizedDiscoveryConfig() *OptimizedDiscoveryConfig {
	return &OptimizedDiscoveryConfig{
		ConnectTimeout:        5 * time.Second,   // Reduced from 30s
		ParallelAttempts:      3,                 // Parallel connection attempts
		EarlySuccessDelay:     200 * time.Millisecond,
		BackoffInitial:        1 * time.Second,
		BackoffMax:           30 * time.Second,
		BackoffMultiplier:    2.0,
		DiscoveryInterval:    10 * time.Second,   // More frequent discovery
		HealthCheckInterval:  30 * time.Second,
		PeerSelectionStrategy: "adaptive",
		MaxFailuresBeforeBackoff: 3,
		RTTThreshold:         500 * time.Millisecond,
		SuccessRateThreshold: 0.7,
	}
}

// NewOptimizedBootstrapDiscovery creates an optimized bootstrap discovery strategy
func NewOptimizedBootstrapDiscovery(host host.Host, bootstrapPeers []peer.AddrInfo, minPeers, maxPeers int) *OptimizedBootstrapDiscovery {
	return &OptimizedBootstrapDiscovery{
		host:           host,
		bootstrapPeers: bootstrapPeers,
		minPeers:       minPeers,
		maxPeers:       maxPeers,
		connections:    make(map[peer.ID]*OptimizedConnectionInfo),
		config:         DefaultOptimizedDiscoveryConfig(),
		metrics:        &OptimizedDiscoveryMetrics{},
	}
}

// Name returns the strategy name
func (o *OptimizedBootstrapDiscovery) Name() string {
	return "optimized_bootstrap"
}

// FindPeers finds peers from bootstrap list with optimization
func (o *OptimizedBootstrapDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo, len(o.bootstrapPeers))
	
	go func() {
		defer close(peerChan)
		
		// Sort peers by priority for optimal discovery
		sortedPeers := o.selectOptimalPeers(o.bootstrapPeers)
		
		for _, peer := range sortedPeers {
			select {
			case peerChan <- peer:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return peerChan, nil
}

// Advertise advertises to bootstrap peers
func (o *OptimizedBootstrapDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	return o.config.DiscoveryInterval, nil
}

// Start starts the optimized bootstrap discovery process
func (o *OptimizedBootstrapDiscovery) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	ticker := time.NewTicker(o.config.DiscoveryInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.ensureOptimizedConnections(ctx)
		}
	}
}

// ensureOptimizedConnections ensures minimum connections with optimization
func (o *OptimizedBootstrapDiscovery) ensureOptimizedConnections(ctx context.Context) {
	connected := len(o.host.Network().Peers())
	
	if connected < o.minPeers {
		// Select best peers for connection
		candidatePeers := o.selectConnectionCandidates()
		
		// Perform parallel connection attempts
		o.performParallelConnections(ctx, candidatePeers)
	}
	
	o.metrics.LastDiscovery = time.Now()
}

// selectConnectionCandidates selects the best peers for connection attempts
func (o *OptimizedBootstrapDiscovery) selectConnectionCandidates() []peer.AddrInfo {
	var candidates []peer.AddrInfo
	
	for _, peer := range o.bootstrapPeers {
		// Skip if already connected
		if o.host.Network().Connectedness(peer.ID) == network.Connected {
			continue
		}
		
		// Check connection info
		o.connectionsMux.RLock()
		connInfo, exists := o.connections[peer.ID]
		o.connectionsMux.RUnlock()
		
		// Skip if currently connecting
		if exists && connInfo.IsConnecting {
			continue
		}
		
		// Skip if too many recent failures with backoff
		if exists && o.shouldSkipDueToFailures(connInfo) {
			continue
		}
		
		candidates = append(candidates, peer)
		
		// Limit candidates to avoid overwhelming
		if len(candidates) >= o.config.ParallelAttempts*2 {
			break
		}
	}
	
	return o.selectOptimalPeers(candidates)
}

// selectOptimalPeers selects optimal peers based on performance metrics
func (o *OptimizedBootstrapDiscovery) selectOptimalPeers(peers []peer.AddrInfo) []peer.AddrInfo {
	if len(peers) <= o.config.ParallelAttempts {
		return peers
	}
	
	// Sort by priority/performance
	peerPriorities := make(map[peer.ID]float64)
	
	o.connectionsMux.RLock()
	for _, peer := range peers {
		priority := 1.0 // Base priority
		
		if connInfo, exists := o.connections[peer.ID]; exists {
			// Boost priority for successful peers
			priority += connInfo.SuccessRate
			
			// Penalize high RTT
			if connInfo.RTT > o.config.RTTThreshold {
				priority *= 0.5
			}
			
			// Penalize recent failures
			if connInfo.Failures > o.config.MaxFailuresBeforeBackoff {
				priority *= 0.2
			}
		}
		
		peerPriorities[peer.ID] = priority
	}
	o.connectionsMux.RUnlock()
	
	// Select top peers
	selected := make([]peer.AddrInfo, 0, o.config.ParallelAttempts)
	for i := 0; i < o.config.ParallelAttempts && i < len(peers); i++ {
		bestPeer := peers[0]
		bestPriority := peerPriorities[bestPeer.ID]
		bestIndex := 0
		
		for j, peer := range peers {
			if peerPriorities[peer.ID] > bestPriority {
				bestPeer = peer
				bestPriority = peerPriorities[peer.ID]
				bestIndex = j
			}
		}
		
		selected = append(selected, bestPeer)
		// Remove from candidates
		peers = append(peers[:bestIndex], peers[bestIndex+1:]...)
	}
	
	return selected
}

// performParallelConnections performs optimized parallel connection attempts
func (o *OptimizedBootstrapDiscovery) performParallelConnections(ctx context.Context, peers []peer.AddrInfo) {
	if len(peers) == 0 {
		return
	}
	
	o.metrics.ParallelConnections++
	
	// Define result type
	type connectionResult struct {
		peer peer.AddrInfo
		err  error
		rtt  time.Duration
	}
	
	// Create channels for coordination
	resultChan := make(chan connectionResult, len(peers))
	
	successChan := make(chan struct{}, 1)
	
	// Start parallel connection attempts
	for i, peerInfo := range peers {
		go func(p peer.AddrInfo, attemptNum int) {
			// Add slight delay for load balancing
			if attemptNum > 0 {
				time.Sleep(time.Duration(attemptNum) * 50 * time.Millisecond)
			}
			
			start := time.Now()
			err := o.connectToPeerOptimized(ctx, p)
			rtt := time.Since(start)
			
			resultChan <- connectionResult{p, err, rtt}
			
			// Signal early success
			if err == nil {
				select {
				case successChan <- struct{}{}:
				default:
				}
			}
		}(peerInfo, i)
	}
	
	// Wait for early success or all attempts
	earlySuccessTimer := time.NewTimer(o.config.EarlySuccessDelay)
	defer earlySuccessTimer.Stop()
	
	completed := 0
	successCount := 0
	
	for completed < len(peers) {
		select {
		case <-successChan:
			// Early success detected, continue to collect remaining results
			o.metrics.EarlySuccesses++
			
		case result := <-resultChan:
			completed++
			o.updateConnectionMetrics(result.peer, result.err, result.rtt)
			
			if result.err == nil {
				successCount++
				o.metrics.SuccessfulConnections++
			} else {
				o.metrics.FailedConnections++
			}
			
		case <-ctx.Done():
			return
		}
	}
	
	// Update efficiency metrics
	if len(peers) > 0 {
		o.metrics.ParallelEfficiency = float64(successCount) / float64(len(peers))
	}
	
	log.Printf("Parallel connection attempt completed: %d/%d successful", successCount, len(peers))
}

// connectToPeerOptimized performs optimized connection to a peer
func (o *OptimizedBootstrapDiscovery) connectToPeerOptimized(ctx context.Context, peer peer.AddrInfo) error {
	o.metrics.TotalAttempts++
	
	// Mark as connecting
	o.connectionsMux.Lock()
	connInfo, exists := o.connections[peer.ID]
	if !exists {
		connInfo = &OptimizedConnectionInfo{
			LastBackoff: o.config.BackoffInitial,
		}
		o.connections[peer.ID] = connInfo
	}
	connInfo.IsConnecting = true
	connInfo.Attempts++
	
	// Create connection context with timeout
	connectCtx, cancel := context.WithTimeout(ctx, o.config.ConnectTimeout)
	connInfo.ConnectionCtx = connectCtx
	connInfo.ConnectionCancel = cancel
	o.connectionsMux.Unlock()
	
	defer func() {
		o.connectionsMux.Lock()
		connInfo.IsConnecting = false
		connInfo.ConnectionCancel = nil
		o.connectionsMux.Unlock()
		cancel()
	}()
	
	// Perform connection
	start := time.Now()
	err := o.host.Connect(connectCtx, peer)
	rtt := time.Since(start)
	
	// Update connection info
	o.connectionsMux.Lock()
	connInfo.LastSeen = time.Now()
	
	if err != nil {
		connInfo.Failures++
		// Update backoff for next attempt
		connInfo.LastBackoff = time.Duration(float64(connInfo.LastBackoff) * o.config.BackoffMultiplier)
		if connInfo.LastBackoff > o.config.BackoffMax {
			connInfo.LastBackoff = o.config.BackoffMax
		}
		
		log.Printf("Failed to connect to peer %s: %v (attempt %d)", peer.ID, err, connInfo.Attempts)
	} else {
		connInfo.ConnectedAt = time.Now()
		connInfo.RTT = rtt
		
		// Update success rate
		if connInfo.Attempts > 0 {
			successRate := 1.0 - (float64(connInfo.Failures) / float64(connInfo.Attempts))
			connInfo.SuccessRate = successRate
		}
		
		o.metrics.LastSuccessfulConnect = time.Now()
		log.Printf("Connected to peer %s (RTT: %v, attempt %d)", peer.ID, rtt, connInfo.Attempts)
	}
	o.connectionsMux.Unlock()
	
	return err
}

// shouldSkipDueToFailures determines if a peer should be skipped due to failures
func (o *OptimizedBootstrapDiscovery) shouldSkipDueToFailures(connInfo *OptimizedConnectionInfo) bool {
	// Skip if too many failures and within backoff period
	if connInfo.Failures > o.config.MaxFailuresBeforeBackoff {
		timeSinceLastAttempt := time.Since(connInfo.LastSeen)
		if timeSinceLastAttempt < connInfo.LastBackoff {
			return true
		}
	}
	
	// Skip if success rate is too low
	if connInfo.Attempts > 5 && connInfo.SuccessRate < o.config.SuccessRateThreshold {
		return true
	}
	
	return false
}

// updateConnectionMetrics updates connection performance metrics
func (o *OptimizedBootstrapDiscovery) updateConnectionMetrics(peer peer.AddrInfo, err error, rtt time.Duration) {
	// Update RTT metrics
	if err == nil && o.metrics.AverageRTT == 0 {
		o.metrics.AverageRTT = rtt
	} else if err == nil {
		// Exponential moving average
		o.metrics.AverageRTT = time.Duration(0.8*float64(o.metrics.AverageRTT) + 0.2*float64(rtt))
	}
	
	if err != nil && rtt > o.config.ConnectTimeout {
		o.metrics.TimeoutReductions++
	}
}

// GetOptimizedMetrics returns discovery performance metrics
func (o *OptimizedBootstrapDiscovery) GetOptimizedMetrics() *OptimizedDiscoveryMetrics {
	return o.metrics
}

// GetConnectionInfo returns connection information for a peer
func (o *OptimizedBootstrapDiscovery) GetConnectionInfo(peerID peer.ID) *OptimizedConnectionInfo {
	o.connectionsMux.RLock()
	defer o.connectionsMux.RUnlock()
	
	if info, exists := o.connections[peerID]; exists {
		// Return copy to avoid race conditions
		return &OptimizedConnectionInfo{
			ConnectedAt:  info.ConnectedAt,
			LastSeen:     info.LastSeen,
			Attempts:     info.Attempts,
			Failures:     info.Failures,
			LastBackoff:  info.LastBackoff,
			RTT:          info.RTT,
			SuccessRate:  info.SuccessRate,
			Priority:     info.Priority,
			IsConnecting: info.IsConnecting,
		}
	}
	
	return nil
}

// UpdateConfig updates the discovery configuration
func (o *OptimizedBootstrapDiscovery) UpdateConfig(config *OptimizedDiscoveryConfig) {
	o.config = config
	log.Printf("Updated optimized discovery config: timeout=%v, parallel=%d", 
		config.ConnectTimeout, config.ParallelAttempts)
}

// GetPerformanceStats returns performance statistics
func (o *OptimizedBootstrapDiscovery) GetPerformanceStats() map[string]interface{} {
	o.connectionsMux.RLock()
	defer o.connectionsMux.RUnlock()
	
	totalPeers := len(o.bootstrapPeers)
	connectedPeers := len(o.host.Network().Peers())
	
	var avgRTT time.Duration
	var avgSuccessRate float64
	activePeers := 0
	
	for _, info := range o.connections {
		if info.Attempts > 0 {
			avgRTT += info.RTT
			avgSuccessRate += info.SuccessRate
			activePeers++
		}
	}
	
	if activePeers > 0 {
		avgRTT /= time.Duration(activePeers)
		avgSuccessRate /= float64(activePeers)
	}
	
	return map[string]interface{}{
		"total_peers":         totalPeers,
		"connected_peers":     connectedPeers,
		"connection_rate":     float64(connectedPeers) / float64(totalPeers),
		"total_attempts":      o.metrics.TotalAttempts,
		"successful_connections": o.metrics.SuccessfulConnections,
		"failed_connections":   o.metrics.FailedConnections,
		"success_rate":        float64(o.metrics.SuccessfulConnections) / float64(o.metrics.TotalAttempts),
		"parallel_connections": o.metrics.ParallelConnections,
		"early_successes":     o.metrics.EarlySuccesses,
		"average_rtt":         avgRTT,
		"average_success_rate": avgSuccessRate,
		"parallel_efficiency":  o.metrics.ParallelEfficiency,
		"backoff_retries":     o.metrics.BackoffRetries,
		"timeout_reductions":  o.metrics.TimeoutReductions,
		"last_discovery":      o.metrics.LastDiscovery,
		"last_successful_connect": o.metrics.LastSuccessfulConnect,
	}
}