package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	p2phost "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/host"
)

// createTestHost creates a test P2P host for discovery testing
func createTestHost(t *testing.T, ctx context.Context) host.Host {
	nodeConfig := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	p2pHost, err := p2phost.NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)

	return p2pHost
}

// createTestPeerID creates a test peer ID
func createTestPeerID(t *testing.T) peer.ID {
	// Create a simple test peer ID
	id, err := peer.Decode("12D3KooWTest" + string(rune(65+t.Name()[len(t.Name())-1])))
	if err != nil {
		// Fallback to a known valid peer ID
		id, _ = peer.Decode("12D3KooWGRUVh6fXBzD3KuRbVoNBrZw3gKHiSF7F7Gv8Z8Z8Z8Z8")
	}
	return id
}

func TestOptimizedBootstrapDiscovery_Creation(t *testing.T) {
	ctx := context.Background()

	// Create test host
	host := createTestHost(t, ctx)
	defer host.Close()

	// Create bootstrap peers
	bootstrapPeers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, bootstrapPeers, 2, 5)
	require.NotNil(t, discovery)

	assert.Equal(t, "optimized_bootstrap", discovery.Name())
	assert.Equal(t, 2, discovery.minPeers)
	assert.Equal(t, 5, discovery.maxPeers)
	assert.Len(t, discovery.bootstrapPeers, 2)
	assert.NotNil(t, discovery.config)
	assert.NotNil(t, discovery.metrics)
}

func TestDefaultOptimizedDiscoveryConfig(t *testing.T) {
	config := DefaultOptimizedDiscoveryConfig()
	require.NotNil(t, config)

	// Test optimized default values
	assert.Equal(t, 5*time.Second, config.ConnectTimeout) // Reduced from 30s
	assert.Equal(t, 3, config.ParallelAttempts)           // Parallel connections
	assert.Equal(t, 200*time.Millisecond, config.EarlySuccessDelay)
	assert.Equal(t, 1*time.Second, config.BackoffInitial)
	assert.Equal(t, 30*time.Second, config.BackoffMax)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
	assert.Equal(t, 10*time.Second, config.DiscoveryInterval) // More frequent
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, "adaptive", config.PeerSelectionStrategy)
	assert.Equal(t, 3, config.MaxFailuresBeforeBackoff)
	assert.Equal(t, 500*time.Millisecond, config.RTTThreshold)
	assert.Equal(t, 0.7, config.SuccessRateThreshold)
}

func TestOptimizedBootstrapDiscovery_FindPeers(t *testing.T) {
	ctx := context.Background()

	// Create test host
	host := createTestHost(t, ctx)
	defer host.Close()

	// Create bootstrap peers
	bootstrapPeers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, bootstrapPeers, 2, 5)

	// Test FindPeers
	peerChan, err := discovery.FindPeers(ctx, "test", nil)
	assert.NoError(t, err)

	// Collect peers from channel
	var foundPeers []peer.AddrInfo
	for peer := range peerChan {
		foundPeers = append(foundPeers, peer)
	}

	assert.Len(t, foundPeers, 3)
}

func TestOptimizedBootstrapDiscovery_Advertise(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)

	ttl, err := discovery.Advertise(ctx, "test", nil)
	assert.NoError(t, err)
	assert.Equal(t, discovery.config.DiscoveryInterval, ttl)
}

func TestOptimizedConnectionInfo_Creation(t *testing.T) {
	info := &OptimizedConnectionInfo{
		ConnectedAt:  time.Now(),
		LastSeen:     time.Now(),
		Attempts:     3,
		Failures:     1,
		LastBackoff:  2 * time.Second,
		RTT:          100 * time.Millisecond,
		SuccessRate:  0.67,
		Priority:     100,
		IsConnecting: false,
	}

	assert.Equal(t, 3, info.Attempts)
	assert.Equal(t, 1, info.Failures)
	assert.Equal(t, 2*time.Second, info.LastBackoff)
	assert.Equal(t, 100*time.Millisecond, info.RTT)
	assert.Equal(t, 0.67, info.SuccessRate)
	assert.False(t, info.IsConnecting)
}

func TestSelectOptimalPeers(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	// Create test peers
	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 2, 5)

	// Add connection info to simulate peer performance
	discovery.connections[peers[0].ID] = &OptimizedConnectionInfo{
		SuccessRate: 0.9,
		RTT:         100 * time.Millisecond,
		Failures:    0,
	}

	discovery.connections[peers[1].ID] = &OptimizedConnectionInfo{
		SuccessRate: 0.5,
		RTT:         800 * time.Millisecond, // High RTT
		Failures:    2,
	}

	discovery.connections[peers[2].ID] = &OptimizedConnectionInfo{
		SuccessRate: 0.8,
		RTT:         200 * time.Millisecond,
		Failures:    1,
	}

	selected := discovery.selectOptimalPeers(peers)

	// Should select up to ParallelAttempts (3) peers
	assert.LessOrEqual(t, len(selected), discovery.config.ParallelAttempts)

	// First selected peer should be the best performing one (peers[0])
	if len(selected) > 0 {
		assert.Equal(t, peers[0].ID, selected[0].ID)
	}
}

func TestShouldSkipDueToFailures(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)

	// Test case 1: Too many failures within backoff period
	connInfo1 := &OptimizedConnectionInfo{
		Failures:    5, // > MaxFailuresBeforeBackoff (3)
		LastSeen:    time.Now().Add(-1 * time.Second),
		LastBackoff: 5 * time.Second,
	}

	assert.True(t, discovery.shouldSkipDueToFailures(connInfo1))

	// Test case 2: Too many failures but backoff period passed
	connInfo2 := &OptimizedConnectionInfo{
		Failures:    5,
		LastSeen:    time.Now().Add(-10 * time.Second),
		LastBackoff: 5 * time.Second,
	}

	assert.False(t, discovery.shouldSkipDueToFailures(connInfo2))

	// Test case 3: Low success rate
	connInfo3 := &OptimizedConnectionInfo{
		Attempts:    10,
		Failures:    7,
		SuccessRate: 0.3, // < SuccessRateThreshold (0.7)
	}

	assert.True(t, discovery.shouldSkipDueToFailures(connInfo3))

	// Test case 4: Good performance
	connInfo4 := &OptimizedConnectionInfo{
		Attempts:    10,
		Failures:    1,
		SuccessRate: 0.9,
		LastSeen:    time.Now(),
	}

	assert.False(t, discovery.shouldSkipDueToFailures(connInfo4))
}

func TestUpdateConnectionMetrics(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)
	peer := peer.AddrInfo{ID: createTestPeerID(t)}

	// Test successful connection metrics update
	discovery.updateConnectionMetrics(peer, nil, 150*time.Millisecond)
	assert.Equal(t, 150*time.Millisecond, discovery.metrics.AverageRTT)

	// Test second successful connection (moving average)
	discovery.updateConnectionMetrics(peer, nil, 250*time.Millisecond)
	expectedAvg := time.Duration(0.8*150+0.2*250) * time.Millisecond
	assert.Equal(t, expectedAvg, discovery.metrics.AverageRTT)

	// Test timeout metric
	timeoutErr := context.DeadlineExceeded
	discovery.updateConnectionMetrics(peer, timeoutErr, 6*time.Second)
	assert.Equal(t, int64(1), discovery.metrics.TimeoutReductions)
}

func TestGetConnectionInfo(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)
	peerID := createTestPeerID(t)

	// Test non-existent peer
	info := discovery.GetConnectionInfo(peerID)
	assert.Nil(t, info)

	// Add connection info
	originalInfo := &OptimizedConnectionInfo{
		ConnectedAt: time.Now(),
		Attempts:    5,
		Failures:    1,
		RTT:         200 * time.Millisecond,
		SuccessRate: 0.8,
	}

	discovery.connections[peerID] = originalInfo

	// Test retrieving connection info
	retrievedInfo := discovery.GetConnectionInfo(peerID)
	require.NotNil(t, retrievedInfo)

	assert.Equal(t, originalInfo.Attempts, retrievedInfo.Attempts)
	assert.Equal(t, originalInfo.Failures, retrievedInfo.Failures)
	assert.Equal(t, originalInfo.RTT, retrievedInfo.RTT)
	assert.Equal(t, originalInfo.SuccessRate, retrievedInfo.SuccessRate)

	// Ensure it's a copy (modification doesn't affect original)
	retrievedInfo.Attempts = 10
	assert.Equal(t, 5, originalInfo.Attempts)
}

func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)

	// Test config update
	newConfig := &OptimizedDiscoveryConfig{
		ConnectTimeout:    3 * time.Second,
		ParallelAttempts:  5,
		EarlySuccessDelay: 100 * time.Millisecond,
		BackoffInitial:    500 * time.Millisecond,
	}

	discovery.UpdateConfig(newConfig)

	assert.Equal(t, newConfig, discovery.config)
	assert.Equal(t, 3*time.Second, discovery.config.ConnectTimeout)
	assert.Equal(t, 5, discovery.config.ParallelAttempts)
}

func TestGetPerformanceStats(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 1, 3)

	// Add some metrics
	discovery.metrics.TotalAttempts = 10
	discovery.metrics.SuccessfulConnections = 7
	discovery.metrics.FailedConnections = 3
	discovery.metrics.ParallelConnections = 5
	discovery.metrics.EarlySuccesses = 2

	// Add connection info
	discovery.connections[peers[0].ID] = &OptimizedConnectionInfo{
		Attempts:    3,
		RTT:         150 * time.Millisecond,
		SuccessRate: 0.8,
	}

	discovery.connections[peers[1].ID] = &OptimizedConnectionInfo{
		Attempts:    2,
		RTT:         250 * time.Millisecond,
		SuccessRate: 0.6,
	}

	stats := discovery.GetPerformanceStats()

	assert.Equal(t, 2, stats["total_peers"])
	assert.Equal(t, int64(10), stats["total_attempts"])
	assert.Equal(t, int64(7), stats["successful_connections"])
	assert.Equal(t, int64(3), stats["failed_connections"])
	assert.Equal(t, 0.7, stats["success_rate"]) // 7/10
	assert.Equal(t, int64(5), stats["parallel_connections"])
	assert.Equal(t, int64(2), stats["early_successes"])

	// Check average calculations
	avgRTT := stats["average_rtt"].(time.Duration)
	assert.Equal(t, 200*time.Millisecond, avgRTT) // (150+250)/2

	avgSuccessRate := stats["average_success_rate"].(float64)
	assert.Equal(t, 0.7, avgSuccessRate) // (0.8+0.6)/2
}

func TestSelectConnectionCandidates(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 1, 5)

	// Mark one peer as connecting
	discovery.connections[peers[0].ID] = &OptimizedConnectionInfo{
		IsConnecting: true,
	}

	// Mark one peer with many failures
	discovery.connections[peers[1].ID] = &OptimizedConnectionInfo{
		Failures:    10,
		LastSeen:    time.Now(),
		LastBackoff: 5 * time.Second,
	}

	candidates := discovery.selectConnectionCandidates()

	// Should only return the third peer
	assert.Len(t, candidates, 1)
	assert.Equal(t, peers[2].ID, candidates[0].ID)
}

// Benchmark tests
func BenchmarkOptimizedBootstrapDiscovery_SelectOptimalPeers(b *testing.B) {
	ctx := context.Background()

	host := createTestHost(&testing.T{}, ctx)
	defer host.Close()

	// Create many peers
	peers := make([]peer.AddrInfo, 100)
	for i := range peers {
		peers[i] = peer.AddrInfo{ID: createTestPeerID(&testing.T{})}
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 10, 50)

	// Add connection info for some peers
	for i := 0; i < 50; i++ {
		discovery.connections[peers[i].ID] = &OptimizedConnectionInfo{
			SuccessRate: float64(i) / 100.0,
			RTT:         time.Duration(i*10) * time.Millisecond,
			Failures:    i % 5,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = discovery.selectOptimalPeers(peers)
	}
}

func BenchmarkOptimizedBootstrapDiscovery_UpdateConnectionMetrics(b *testing.B) {
	ctx := context.Background()

	host := createTestHost(&testing.T{}, ctx)
	defer host.Close()

	discovery := NewOptimizedBootstrapDiscovery(host, nil, 1, 3)
	peer := peer.AddrInfo{ID: createTestPeerID(&testing.T{})}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discovery.updateConnectionMetrics(peer, nil, time.Duration(i)*time.Millisecond)
	}
}

func TestOptimizedDiscoveryMetrics_Initialization(t *testing.T) {
	metrics := &OptimizedDiscoveryMetrics{}

	// Test zero initialization
	assert.Equal(t, int64(0), metrics.TotalAttempts)
	assert.Equal(t, int64(0), metrics.SuccessfulConnections)
	assert.Equal(t, int64(0), metrics.FailedConnections)
	assert.Equal(t, int64(0), metrics.ParallelConnections)
	assert.Equal(t, int64(0), metrics.EarlySuccesses)
	assert.Equal(t, time.Duration(0), metrics.AverageRTT)
	assert.Equal(t, time.Duration(0), metrics.AverageConnectTime)
	assert.Equal(t, int64(0), metrics.BackoffRetries)
	assert.Equal(t, int64(0), metrics.TimeoutReductions)
	assert.Equal(t, 0.0, metrics.ParallelEfficiency)
}
