package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
)

// NetworkMonitor monitors network health and performance
type NetworkMonitor struct {
	config   *NetworkMonitorConfig
	metrics  *NetworkMetrics
	peers    map[peer.ID]*PeerMetrics
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// NetworkMonitorConfig configures the network monitor
type NetworkMonitorConfig struct {
	MonitorInterval     time.Duration `json:"monitor_interval"`
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`
	MaxPeers           int           `json:"max_peers"`
	EnableLatencyCheck bool          `json:"enable_latency_check"`
	EnableBandwidthCheck bool         `json:"enable_bandwidth_check"`
}

// NetworkMetrics holds network-wide metrics
type NetworkMetrics struct {
	TotalPeers        int                    `json:"total_peers"`
	ActivePeers       int                    `json:"active_peers"`
	AverageLatency    time.Duration          `json:"average_latency"`
	TotalBandwidth    int64                  `json:"total_bandwidth"`
	PacketLoss        float64                `json:"packet_loss"`
	LastUpdated       time.Time              `json:"last_updated"`
	PeerMetrics       map[peer.ID]*PeerMetrics `json:"peer_metrics"`
}

// PeerMetrics holds metrics for individual peers
type PeerMetrics struct {
	PeerID          peer.ID       `json:"peer_id"`
	Latency         time.Duration `json:"latency"`
	Bandwidth       int64         `json:"bandwidth"`
	PacketLoss      float64       `json:"packet_loss"`
	ConnectionTime  time.Time     `json:"connection_time"`
	LastSeen        time.Time     `json:"last_seen"`
	MessagesSent    int64         `json:"messages_sent"`
	MessagesReceived int64        `json:"messages_received"`
	BytesSent       int64         `json:"bytes_sent"`
	BytesReceived   int64         `json:"bytes_received"`
	IsHealthy       bool          `json:"is_healthy"`
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(config *NetworkMonitorConfig) *NetworkMonitor {
	if config == nil {
		config = &NetworkMonitorConfig{
			MonitorInterval:      30 * time.Second,
			HealthCheckTimeout:   10 * time.Second,
			MaxPeers:            1000,
			EnableLatencyCheck:   true,
			EnableBandwidthCheck: true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &NetworkMonitor{
		config: config,
		metrics: &NetworkMetrics{
			PeerMetrics: make(map[peer.ID]*PeerMetrics),
			LastUpdated: time.Now(),
		},
		peers:  make(map[peer.ID]*PeerMetrics),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the network monitor
func (nm *NetworkMonitor) Start() error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.running {
		return nil
	}

	nm.running = true
	go nm.monitorLoop()

	log.Info().Msg("Network monitor started")
	return nil
}

// Stop stops the network monitor
func (nm *NetworkMonitor) Stop() error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.running {
		return nil
	}

	nm.cancel()
	nm.running = false

	log.Info().Msg("Network monitor stopped")
	return nil
}

// AddPeer adds a peer to monitoring
func (nm *NetworkMonitor) AddPeer(peerID peer.ID) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.peers[peerID]; !exists {
		nm.peers[peerID] = &PeerMetrics{
			PeerID:         peerID,
			ConnectionTime: time.Now(),
			LastSeen:       time.Now(),
			IsHealthy:      true,
		}
		nm.updateMetrics()
	}
}

// RemovePeer removes a peer from monitoring
func (nm *NetworkMonitor) RemovePeer(peerID peer.ID) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	delete(nm.peers, peerID)
	delete(nm.metrics.PeerMetrics, peerID)
	nm.updateMetrics()
}

// UpdatePeerMetrics updates metrics for a specific peer
func (nm *NetworkMonitor) UpdatePeerMetrics(peerID peer.ID, latency time.Duration, bandwidth int64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if peer, exists := nm.peers[peerID]; exists {
		peer.Latency = latency
		peer.Bandwidth = bandwidth
		peer.LastSeen = time.Now()
		peer.IsHealthy = nm.isPeerHealthy(peer)
		nm.updateMetrics()
	}
}

// RecordMessage records a message sent or received
func (nm *NetworkMonitor) RecordMessage(peerID peer.ID, sent bool, bytes int64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if peer, exists := nm.peers[peerID]; exists {
		if sent {
			peer.MessagesSent++
			peer.BytesSent += bytes
		} else {
			peer.MessagesReceived++
			peer.BytesReceived += bytes
		}
		peer.LastSeen = time.Now()
	}
}

// GetMetrics returns current network metrics
func (nm *NetworkMonitor) GetMetrics() *NetworkMetrics {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := &NetworkMetrics{
		TotalPeers:     nm.metrics.TotalPeers,
		ActivePeers:    nm.metrics.ActivePeers,
		AverageLatency: nm.metrics.AverageLatency,
		TotalBandwidth: nm.metrics.TotalBandwidth,
		PacketLoss:     nm.metrics.PacketLoss,
		LastUpdated:    nm.metrics.LastUpdated,
		PeerMetrics:    make(map[peer.ID]*PeerMetrics),
	}

	for id, peer := range nm.metrics.PeerMetrics {
		peerCopy := *peer
		metrics.PeerMetrics[id] = &peerCopy
	}

	return metrics
}

// GetPeerMetrics returns metrics for a specific peer
func (nm *NetworkMonitor) GetPeerMetrics(peerID peer.ID) *PeerMetrics {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	if peer, exists := nm.peers[peerID]; exists {
		peerCopy := *peer
		return &peerCopy
	}
	return nil
}

// IsHealthy returns whether the network is healthy
func (nm *NetworkMonitor) IsHealthy() bool {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	healthyPeers := 0
	for _, peer := range nm.peers {
		if peer.IsHealthy {
			healthyPeers++
		}
	}

	// Consider network healthy if at least 50% of peers are healthy
	return len(nm.peers) == 0 || float64(healthyPeers)/float64(len(nm.peers)) >= 0.5
}

// monitorLoop runs the monitoring loop
func (nm *NetworkMonitor) monitorLoop() {
	ticker := time.NewTicker(nm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all peers
func (nm *NetworkMonitor) performHealthChecks() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	now := time.Now()
	for _, peer := range nm.peers {
		// Check if peer has been seen recently
		if now.Sub(peer.LastSeen) > nm.config.HealthCheckTimeout*2 {
			peer.IsHealthy = false
		} else {
			peer.IsHealthy = nm.isPeerHealthy(peer)
		}
	}

	nm.updateMetrics()
}

// isPeerHealthy determines if a peer is healthy based on its metrics
func (nm *NetworkMonitor) isPeerHealthy(peer *PeerMetrics) bool {
	// Consider peer healthy if:
	// 1. Latency is reasonable (< 1 second)
	// 2. Packet loss is low (< 5%)
	// 3. Has been seen recently
	
	if peer.Latency > time.Second {
		return false
	}
	
	if peer.PacketLoss > 0.05 {
		return false
	}
	
	if time.Since(peer.LastSeen) > nm.config.HealthCheckTimeout {
		return false
	}
	
	return true
}

// updateMetrics updates the aggregate network metrics
func (nm *NetworkMonitor) updateMetrics() {
	nm.metrics.TotalPeers = len(nm.peers)
	nm.metrics.ActivePeers = 0
	nm.metrics.TotalBandwidth = 0
	
	var totalLatency time.Duration
	var totalPacketLoss float64
	
	for id, peer := range nm.peers {
		nm.metrics.PeerMetrics[id] = peer
		
		if peer.IsHealthy {
			nm.metrics.ActivePeers++
		}
		
		nm.metrics.TotalBandwidth += peer.Bandwidth
		totalLatency += peer.Latency
		totalPacketLoss += peer.PacketLoss
	}
	
	if nm.metrics.TotalPeers > 0 {
		nm.metrics.AverageLatency = totalLatency / time.Duration(nm.metrics.TotalPeers)
		nm.metrics.PacketLoss = totalPacketLoss / float64(nm.metrics.TotalPeers)
	}
	
	nm.metrics.LastUpdated = time.Now()
}

// GetHealthStatus returns a detailed health status
func (nm *NetworkMonitor) GetHealthStatus() map[string]interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	status := map[string]interface{}{
		"healthy":         nm.IsHealthy(),
		"total_peers":     nm.metrics.TotalPeers,
		"active_peers":    nm.metrics.ActivePeers,
		"average_latency": nm.metrics.AverageLatency.String(),
		"total_bandwidth": nm.metrics.TotalBandwidth,
		"packet_loss":     nm.metrics.PacketLoss,
		"last_updated":    nm.metrics.LastUpdated,
	}

	return status
}
