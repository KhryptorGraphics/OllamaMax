package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	HealthCheckProtocol = "/ollama/health/1.0.0"
	PingProtocol        = "/ollama/ping/1.0.0"
)

// PeerHealth represents the health status of a peer
type PeerHealth struct {
	PeerID           peer.ID       `json:"peer_id"`
	Status           HealthStatus  `json:"status"`
	LastCheck        time.Time     `json:"last_check"`
	LastSeen         time.Time     `json:"last_seen"`
	Latency          time.Duration `json:"latency"`
	ConsecutiveFails int           `json:"consecutive_fails"`
	TotalChecks      int           `json:"total_checks"`
	SuccessRate      float64       `json:"success_rate"`

	// Extended health metrics
	ConnectionCount int     `json:"connection_count"`
	Bandwidth       int64   `json:"bandwidth"`
	CPUUsage        float64 `json:"cpu_usage"`
	MemoryUsage     float64 `json:"memory_usage"`
	ModelCount      int     `json:"model_count"`
}

// HealthStatus represents the health status of a peer
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthChecker monitors peer health and connectivity
type HealthChecker struct {
	host       host.Host
	mu         sync.RWMutex
	peerHealth map[peer.ID]*PeerHealth

	// Configuration
	checkInterval time.Duration
	timeout       time.Duration
	maxFailures   int

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	totalChecks   int64
	totalFailures int64
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(h host.Host) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())

	hc := &HealthChecker{
		host:          h,
		peerHealth:    make(map[peer.ID]*PeerHealth),
		checkInterval: 30 * time.Second,
		timeout:       10 * time.Second,
		maxFailures:   3,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Set up protocol handlers
	h.SetStreamHandler(HealthCheckProtocol, hc.handleHealthCheck)
	h.SetStreamHandler(PingProtocol, hc.handlePing)

	return hc
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.healthCheckLoop()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	hc.cancel()
	hc.wg.Wait()
}

// AddPeer adds a peer to health monitoring
func (hc *HealthChecker) AddPeer(peerID peer.ID) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if _, exists := hc.peerHealth[peerID]; !exists {
		hc.peerHealth[peerID] = &PeerHealth{
			PeerID:   peerID,
			Status:   HealthStatusUnknown,
			LastSeen: time.Now(),
		}
	}
}

// RemovePeer removes a peer from health monitoring
func (hc *HealthChecker) RemovePeer(peerID peer.ID) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.peerHealth, peerID)
}

// GetPeerHealth returns the health status of a peer
func (hc *HealthChecker) GetPeerHealth(peerID peer.ID) *PeerHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	if health, exists := hc.peerHealth[peerID]; exists {
		// Return a copy to avoid race conditions
		healthCopy := *health
		return &healthCopy
	}
	return nil
}

// GetHealthyPeers returns a list of healthy peers
func (hc *HealthChecker) GetHealthyPeers() []peer.ID {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var healthy []peer.ID
	for peerID, health := range hc.peerHealth {
		if health.Status == HealthStatusHealthy {
			healthy = append(healthy, peerID)
		}
	}
	return healthy
}

// GetAllPeerHealth returns health status for all monitored peers
func (hc *HealthChecker) GetAllPeerHealth() map[peer.ID]*PeerHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	result := make(map[peer.ID]*PeerHealth)
	for peerID, health := range hc.peerHealth {
		healthCopy := *health
		result[peerID] = &healthCopy
	}
	return result
}

// CheckPeerHealth performs an immediate health check on a peer
func (hc *HealthChecker) CheckPeerHealth(ctx context.Context, peerID peer.ID) error {
	start := time.Now()

	// Try to ping the peer
	latency, err := hc.pingPeer(ctx, peerID)

	hc.mu.Lock()
	defer hc.mu.Unlock()

	health, exists := hc.peerHealth[peerID]
	if !exists {
		health = &PeerHealth{
			PeerID:   peerID,
			Status:   HealthStatusUnknown,
			LastSeen: time.Now(),
		}
		hc.peerHealth[peerID] = health
	}

	health.LastCheck = start
	health.TotalChecks++

	if err != nil {
		health.ConsecutiveFails++
		hc.totalFailures++

		// Update status based on consecutive failures
		if health.ConsecutiveFails >= hc.maxFailures {
			health.Status = HealthStatusUnhealthy
		} else if health.ConsecutiveFails > 1 {
			health.Status = HealthStatusDegraded
		}

		return fmt.Errorf("health check failed for peer %s: %w", peerID, err)
	}

	// Successful health check
	health.ConsecutiveFails = 0
	health.Latency = latency
	health.LastSeen = time.Now()
	health.Status = HealthStatusHealthy

	// Update success rate
	successfulChecks := health.TotalChecks - health.ConsecutiveFails
	health.SuccessRate = float64(successfulChecks) / float64(health.TotalChecks)

	return nil
}

// healthCheckLoop runs periodic health checks
func (hc *HealthChecker) healthCheckLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.performHealthChecks()
		}
	}
}

// performHealthChecks performs health checks on all monitored peers
func (hc *HealthChecker) performHealthChecks() {
	hc.mu.RLock()
	peers := make([]peer.ID, 0, len(hc.peerHealth))
	for peerID := range hc.peerHealth {
		peers = append(peers, peerID)
	}
	hc.mu.RUnlock()

	// Check peers concurrently
	var wg sync.WaitGroup
	for _, peerID := range peers {
		wg.Add(1)
		go func(pid peer.ID) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(hc.ctx, hc.timeout)
			defer cancel()

			hc.CheckPeerHealth(ctx, pid)
		}(peerID)
	}

	wg.Wait()
	hc.totalChecks++
}

// pingPeer sends a ping to a peer and measures latency
func (hc *HealthChecker) pingPeer(ctx context.Context, peerID peer.ID) (time.Duration, error) {
	start := time.Now()

	// Open a stream to the peer
	stream, err := hc.host.NewStream(ctx, peerID, PingProtocol)
	if err != nil {
		return 0, fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()

	// Send ping message
	pingMsg := []byte("ping")
	if _, err := stream.Write(pingMsg); err != nil {
		return 0, fmt.Errorf("failed to send ping: %w", err)
	}

	// Read pong response
	pongMsg := make([]byte, 4)
	if _, err := stream.Read(pongMsg); err != nil {
		return 0, fmt.Errorf("failed to read pong: %w", err)
	}

	if string(pongMsg) != "pong" {
		return 0, fmt.Errorf("invalid pong response: %s", string(pongMsg))
	}

	return time.Since(start), nil
}

// handlePing handles incoming ping requests
func (hc *HealthChecker) handlePing(stream network.Stream) {
	defer stream.Close()

	// Read ping message
	pingMsg := make([]byte, 4)
	if _, err := stream.Read(pingMsg); err != nil {
		return
	}

	if string(pingMsg) == "ping" {
		// Send pong response
		stream.Write([]byte("pong"))
	}
}

// handleHealthCheck handles incoming health check requests
func (hc *HealthChecker) handleHealthCheck(stream network.Stream) {
	defer stream.Close()

	// Send health response (simplified JSON)
	response := fmt.Sprintf(`{"status":"healthy","timestamp":%d,"peer_id":"%s","connections":%d}`,
		time.Now().Unix(), hc.host.ID().String(), len(hc.host.Network().Peers()))

	stream.Write([]byte(response))
}

// GetMetrics returns health checker metrics
func (hc *HealthChecker) GetMetrics() map[string]interface{} {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, health := range hc.peerHealth {
		switch health.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	return map[string]interface{}{
		"total_peers":     len(hc.peerHealth),
		"healthy_peers":   healthyCount,
		"degraded_peers":  degradedCount,
		"unhealthy_peers": unhealthyCount,
		"total_checks":    hc.totalChecks,
		"total_failures":  hc.totalFailures,
	}
}
