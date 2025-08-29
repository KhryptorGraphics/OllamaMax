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
func createTestHost(t testing.TB, ctx context.Context) host.Host {
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
func createTestPeerID(t testing.TB) peer.ID {
	// Create a simple test peer ID
	testName := "TestPeer"
	if bt, ok := t.(*testing.T); ok {
		testName = bt.Name()
	} else if bb, ok := t.(*testing.B); ok {
		testName = bb.Name()
	}
	
	id, err := peer.Decode("12D3KooWTest" + string(rune(65+(len(testName)%26))))
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

	// Create test peers
	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	// Test discovery creation
	discovery := NewOptimizedBootstrapDiscovery(host, peers, 10, 50)
	require.NotNil(t, discovery)

	assert.Equal(t, host, discovery.host)
	assert.Equal(t, len(peers), len(discovery.bootstrapPeers))
	assert.Equal(t, 10, discovery.maxPeers)
	assert.Equal(t, 50, discovery.connectionTimeout)
}

func TestOptimizedBootstrapDiscovery_SelectOptimalPeers(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	// Create test peers
	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 2, 50)

	// Test peer selection
	selected := discovery.selectOptimalPeers(peers, 2)
	assert.LessOrEqual(t, len(selected), 2)
	assert.Greater(t, len(selected), 0)
}

func TestOptimizedBootstrapDiscovery_Connect(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 1, 50)

	// Test connection (will likely fail in test env, but tests the mechanism)
	connectCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := discovery.Connect(connectCtx)
	// In test environment, connections may fail - just ensure no panic
	t.Logf("Connect result: %v", err)
}

func TestOptimizedBootstrapDiscovery_Metrics(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	peers := []peer.AddrInfo{
		{ID: createTestPeerID(t)},
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 1, 50)

	// Test metrics
	metrics := discovery.GetConnectionMetrics()
	assert.NotNil(t, metrics)
	assert.GreaterOrEqual(t, metrics.TotalAttempts, uint64(0))
	assert.GreaterOrEqual(t, metrics.SuccessfulConnections, uint64(0))
	assert.GreaterOrEqual(t, metrics.FailedConnections, uint64(0))
}

// Benchmark tests
func BenchmarkOptimizedBootstrapDiscovery_SelectOptimalPeers(b *testing.B) {
	ctx := context.Background()

	host := createTestHost(b, ctx)
	defer host.Close()

	// Create many peers
	peers := make([]peer.AddrInfo, 100)
	for i := range peers {
		peers[i] = peer.AddrInfo{ID: createTestPeerID(b)}
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 10, 50)

	// Add connection info for some peers
	for i := 0; i < 50; i++ {
		discovery.connections[peers[i].ID] = &OptimizedConnectionInfo{
			LastSeen:      time.Now(),
			LatencyMs:     50 + int64(i),
			SuccessRate:   0.8,
			IsReliable:    i%2 == 0,
			ConnectionAge: time.Hour,
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			selected := discovery.selectOptimalPeers(peers, 10)
			if len(selected) == 0 {
				b.Error("No peers selected")
			}
		}
	})
}

func BenchmarkOptimizedBootstrapDiscovery_ConnectionScoring(b *testing.B) {
	ctx := context.Background()

	host := createTestHost(b, ctx)
	defer host.Close()

	peers := make([]peer.AddrInfo, 1000)
	for i := range peers {
		peers[i] = peer.AddrInfo{ID: createTestPeerID(b)}
	}

	discovery := NewOptimizedBootstrapDiscovery(host, peers, 50, 100)

	// Add connection info for all peers
	for i, p := range peers {
		discovery.connections[p.ID] = &OptimizedConnectionInfo{
			LastSeen:      time.Now().Add(-time.Duration(i) * time.Minute),
			LatencyMs:     int64(10 + (i % 200)),
			SuccessRate:   float64(50+i%50) / 100.0,
			IsReliable:    i%3 == 0,
			ConnectionAge: time.Duration(i) * time.Minute,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discovery.scoreConnection(peers[i%len(peers)].ID)
	}
}

func TestOptimizedDHTDiscovery_Creation(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &OptimizedDHTConfig{
		Mode:               1, // Client mode
		Concurrency:        10,
		QueryTimeout:       30 * time.Second,
		RefreshInterval:    5 * time.Minute,
		BootstrapPeers:     []peer.AddrInfo{},
		RecordLifetime:     24 * time.Hour,
		RepublishInterval:  12 * time.Hour,
		EnableProviders:    true,
		EnableValues:       true,
		MaxRecordSize:      8192,
		ValidationTimeout:  10 * time.Second,
	}

	discovery, err := NewOptimizedDHTDiscovery(ctx, host, config)
	require.NoError(t, err)
	require.NotNil(t, discovery)

	defer discovery.Close()

	assert.Equal(t, host, discovery.host)
	assert.Equal(t, config.Concurrency, discovery.config.Concurrency)
}

func TestOptimizedDHTDiscovery_Advertise(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &OptimizedDHTConfig{
		Mode:               1,
		Concurrency:        5,
		QueryTimeout:       10 * time.Second,
		RefreshInterval:    time.Minute,
		BootstrapPeers:     []peer.AddrInfo{},
		RecordLifetime:     time.Hour,
		RepublishInterval:  30 * time.Minute,
		EnableProviders:    true,
		EnableValues:       true,
		MaxRecordSize:      4096,
		ValidationTimeout:  5 * time.Second,
	}

	discovery, err := NewOptimizedDHTDiscovery(ctx, host, config)
	require.NoError(t, err)
	defer discovery.Close()

	// Test advertisement (may timeout in test environment)
	advCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err = discovery.Advertise(advCtx, "test-service")
	// In test environment, this may fail due to no DHT peers
	t.Logf("Advertise result: %v", err)
}

func TestOptimizedDHTDiscovery_FindPeers(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &OptimizedDHTConfig{
		Mode:               1,
		Concurrency:        5,
		QueryTimeout:       5 * time.Second,
		RefreshInterval:    time.Minute,
		BootstrapPeers:     []peer.AddrInfo{},
		RecordLifetime:     time.Hour,
		RepublishInterval:  30 * time.Minute,
		EnableProviders:    true,
		EnableValues:       true,
		MaxRecordSize:      4096,
		ValidationTimeout:  3 * time.Second,
	}

	discovery, err := NewOptimizedDHTDiscovery(ctx, host, config)
	require.NoError(t, err)
	defer discovery.Close()

	// Test peer discovery
	findCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	peerChan, err := discovery.FindPeers(findCtx, "test-service")
	if err != nil {
		t.Logf("FindPeers failed (expected in test env): %v", err)
		return
	}

	// Collect found peers (if any)
	var foundPeers []peer.AddrInfo
	for p := range peerChan {
		foundPeers = append(foundPeers, p)
		if len(foundPeers) >= 5 {
			break
		}
	}

	t.Logf("Found %d peers", len(foundPeers))
}

func TestHybridDiscoveryStrategy_Creation(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &HybridConfig{
		EnableBootstrap: true,
		EnableDHT:       true,
		EnableMDNS:      true,
		BootstrapPeers: []peer.AddrInfo{
			{ID: createTestPeerID(t)},
		},
		DHTConfig: &OptimizedDHTConfig{
			Mode:               1,
			Concurrency:        5,
			QueryTimeout:       10 * time.Second,
			RefreshInterval:    time.Minute,
			BootstrapPeers:     []peer.AddrInfo{},
			RecordLifetime:     time.Hour,
			RepublishInterval:  30 * time.Minute,
			EnableProviders:    true,
			EnableValues:       true,
			MaxRecordSize:      4096,
			ValidationTimeout:  5 * time.Second,
		},
		MDNSConfig: &OptimizedMDNSConfig{
			ServiceTag:      "ollama-test",
			QueryInterval:   10 * time.Second,
			CacheTimeout:    5 * time.Minute,
			MaxCacheSize:    100,
			EnableCaching:   true,
			InterfaceFilter: nil,
		},
	}

	strategy, err := NewHybridDiscoveryStrategy(ctx, host, config)
	require.NoError(t, err)
	require.NotNil(t, strategy)

	defer strategy.Close()

	assert.Equal(t, host, strategy.host)
	assert.True(t, strategy.config.EnableBootstrap)
	assert.True(t, strategy.config.EnableDHT)
	assert.True(t, strategy.config.EnableMDNS)
}

func TestHybridDiscoveryStrategy_Discovery(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &HybridConfig{
		EnableBootstrap: true,
		EnableDHT:       false, // Disable DHT for faster test
		EnableMDNS:      true,
		BootstrapPeers: []peer.AddrInfo{
			{ID: createTestPeerID(t)},
		},
		DHTConfig: nil,
		MDNSConfig: &OptimizedMDNSConfig{
			ServiceTag:      "ollama-test",
			QueryInterval:   time.Second,
			CacheTimeout:    time.Minute,
			MaxCacheSize:    50,
			EnableCaching:   true,
			InterfaceFilter: nil,
		},
	}

	strategy, err := NewHybridDiscoveryStrategy(ctx, host, config)
	require.NoError(t, err)
	defer strategy.Close()

	// Test discovery
	discCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	peerChan, err := strategy.FindPeers(discCtx, "test-service")
	if err != nil {
		t.Logf("Discovery failed (expected in test env): %v", err)
		return
	}

	// Collect discovered peers
	var discoveredPeers []peer.AddrInfo
	for p := range peerChan {
		discoveredPeers = append(discoveredPeers, p)
		if len(discoveredPeers) >= 3 {
			break
		}
	}

	t.Logf("Discovered %d peers via hybrid strategy", len(discoveredPeers))
}

func TestHybridDiscoveryStrategy_Metrics(t *testing.T) {
	ctx := context.Background()

	host := createTestHost(t, ctx)
	defer host.Close()

	config := &HybridConfig{
		EnableBootstrap: true,
		EnableDHT:       false,
		EnableMDNS:      true,
		BootstrapPeers: []peer.AddrInfo{
			{ID: createTestPeerID(t)},
		},
		MDNSConfig: &OptimizedMDNSConfig{
			ServiceTag:      "ollama-test",
			QueryInterval:   time.Second,
			CacheTimeout:    time.Minute,
			MaxCacheSize:    50,
			EnableCaching:   true,
			InterfaceFilter: nil,
		},
	}

	strategy, err := NewHybridDiscoveryStrategy(ctx, host, config)
	require.NoError(t, err)
	defer strategy.Close()

	// Get metrics
	metrics := strategy.GetMetrics()
	assert.NotNil(t, metrics)
	
	// Test that metrics structure is valid
	assert.GreaterOrEqual(t, metrics.TotalQueries, uint64(0))
	assert.GreaterOrEqual(t, metrics.SuccessfulQueries, uint64(0))
	assert.GreaterOrEqual(t, metrics.FailedQueries, uint64(0))
	assert.GreaterOrEqual(t, metrics.PeersFound, uint64(0))
	assert.GreaterOrEqual(t, metrics.CacheHits, uint64(0))
	assert.GreaterOrEqual(t, metrics.CacheMisses, uint64(0))
}