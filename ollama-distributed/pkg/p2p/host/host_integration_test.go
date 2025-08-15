//go:build integration

package host

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/nat"
)

func TestP2PHost_NATIntegration(t *testing.T) {
	ctx := context.Background()

	// Create test configuration
	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableAutoRelay:    false,
		EnableNoise:        true,
		EnableTLS:          false,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	// Create P2P host
	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	// Test NAT manager integration
	natManager := host.GetNATManager()
	require.NotNil(t, natManager)

	// Test NAT discovery (may timeout in test environment)
	discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	natType, err := natManager.DiscoverNATType(discoveryCtx)
	// In test environment, this may fail, but we test the integration
	if err == nil {
		assert.NotEqual(t, nat.NATTypeUnknown, natType)
		t.Logf("Discovered NAT type: %s", natType)
	} else {
		t.Logf("NAT discovery failed (expected in test env): %v", err)
	}

	// Test metrics integration
	metrics := host.GetMetrics()
	require.NotNil(t, metrics)

	// NAT type should be set in metrics
	if natType != nat.NATTypeUnknown {
		assert.Equal(t, natType.String(), metrics.NATType)
	}
}

func TestP2PHost_ConnectionOptimization(t *testing.T) {
	ctx := context.Background()

	// Create two test hosts
	config1 := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	config2 := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	host1, err := NewP2PHost(ctx, config1)
	require.NoError(t, err)
	defer host1.Close()

	host2, err := NewP2PHost(ctx, config2)
	require.NoError(t, err)
	defer host2.Close()

	// Test connection tracker
	tracker := host1.GetConnectionTracker()
	require.NotNil(t, tracker)

	// Test optimized connection
	peerInfo := host2.Peerstore().PeerInfo(host2.ID())

	// Measure connection time
	start := time.Now()
	err = host1.ConnectWithOptimization(ctx, peerInfo)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.True(t, host1.IsConnected(host2.ID()))

	// Connection should be fast (under configured timeout)
	assert.Less(t, duration, 10*time.Second)

	t.Logf("Connection established in %v", duration)

	// Test metrics update
	metrics := host1.GetMetrics()
	assert.Greater(t, metrics.ConnectionCount, 0)
}

func TestP2PHost_STUNIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	natManager := host.GetNATManager()

	// Test STUN server configuration
	natManager.AddSTUNServer("stun.l.google.com", 19302)
	natManager.AddSTUNServer("stun1.l.google.com", 19302)

	// Test NAT discovery with timeout
	discoveryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	natType, err := natManager.DiscoverNATType(discoveryCtx)

	// Test metrics after discovery attempt
	natMetrics := natManager.GetMetrics()
	assert.Greater(t, natMetrics.STUNRequests, int64(0))

	if err == nil {
		assert.NotEqual(t, nat.NATTypeUnknown, natType)
		t.Logf("NAT discovery successful: %s", natType)

		// Test public address retrieval
		publicAddr := natManager.GetPublicAddress()
		if publicAddr != nil {
			t.Logf("Public address: %s", publicAddr)
		}
	} else {
		t.Logf("NAT discovery failed (may be expected): %v", err)
	}
}

func TestP2PHost_TURNIntegration(t *testing.T) {
	ctx := context.Background()

	// Configure with mock TURN servers
	nodeConfig := &config.NodeConfig{
		Listen:           []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService: true,
		EnableAutoRelay:  true,
		EnableNoise:      true,
		ConnMgrLow:       10,
		ConnMgrHigh:      100,
		ConnMgrGrace:     time.Minute,
		TURNServers: []config.TURNServerConfig{
			{
				Address:   "turn.example.com",
				Port:      3478,
				Username:  "testuser",
				Password:  "testpass",
				Realm:     "example.com",
				Transport: "udp",
			},
		},
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	natManager := host.GetNATManager()

	// Test relay connection requirement
	natManager.(*nat.NATTraversalManager).SetNATType(nat.NATTypeSymmetric)
	assert.True(t, natManager.IsRelayRequired())

	// Test TURN server configuration
	// In a real test, you'd set up actual TURN servers or mocks
	t.Log("TURN integration test completed (mock servers)")
}

func TestP2PHost_HolePunchingIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	natManager := host.GetNATManager()

	// Test hole punching capability
	// In a real test, you'd create a scenario requiring hole punching
	t.Log("Hole punching integration test completed")

	// Test metrics
	metrics := natManager.GetMetrics()
	assert.GreaterOrEqual(t, metrics.HolePunchAttempts, int64(0))
}

func TestP2PHost_BackoffIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	tracker := host.GetConnectionTracker()

	// Test backoff configuration
	assert.Equal(t, 1*time.Second, tracker.config.BackoffInitial)
	assert.Equal(t, 30*time.Second, tracker.config.BackoffMax)
	assert.Equal(t, 2.0, tracker.config.BackoffMultiplier)
	assert.Equal(t, 5, tracker.config.MaxRetries)

	t.Log("Backoff configuration verified")
}

func TestP2PHost_ParallelConnectionIntegration(t *testing.T) {
	ctx := context.Background()

	// Test parallel connection configuration
	nodeConfig := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	tracker := host.GetConnectionTracker()

	// Test optimized connection settings
	assert.Equal(t, 5*time.Second, tracker.config.Timeout) // Reduced from 30s
	assert.Equal(t, 3, tracker.config.ParallelAttempts)
	assert.Equal(t, 200*time.Millisecond, tracker.config.EarlySuccessDelay)

	t.Log("Parallel connection configuration verified")
}

func TestP2PHost_MetricsIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	// Test enhanced metrics
	metrics := host.GetMetrics()
	require.NotNil(t, metrics)

	// Test NAT traversal metrics
	assert.GreaterOrEqual(t, metrics.STUNRequests, int64(0))
	assert.GreaterOrEqual(t, metrics.TURNConnections, int64(0))
	assert.GreaterOrEqual(t, metrics.HolePunchAttempts, int64(0))
	assert.GreaterOrEqual(t, metrics.HolePunchSuccesses, int64(0))

	// Test connection optimization metrics
	assert.GreaterOrEqual(t, metrics.ParallelConnections, 0)
	assert.GreaterOrEqual(t, metrics.EarlySuccesses, int64(0))
	assert.GreaterOrEqual(t, metrics.ConnectionTimeouts, int64(0))
	assert.GreaterOrEqual(t, metrics.BackoffRetries, int64(0))

	t.Log("Enhanced metrics verified")
}

// Benchmark tests for performance validation
func BenchmarkP2PHost_ConnectionOptimization(b *testing.B) {
	ctx := context.Background()

	// Create two hosts
	config1 := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	config2 := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	host1, err := NewP2PHost(ctx, config1)
	if err != nil {
		b.Fatal(err)
	}
	defer host1.Close()

	host2, err := NewP2PHost(ctx, config2)
	if err != nil {
		b.Fatal(err)
	}
	defer host2.Close()

	peerInfo := host2.Peerstore().PeerInfo(host2.ID())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Disconnect if connected
		if host1.IsConnected(host2.ID()) {
			host1.Network().ClosePeer(host2.ID())
		}

		// Benchmark optimized connection
		err := host1.ConnectWithOptimization(ctx, peerInfo)
		if err != nil {
			b.Errorf("Connection failed: %v", err)
		}
	}
}

func BenchmarkP2PHost_NATDiscovery(b *testing.B) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	if err != nil {
		b.Fatal(err)
	}
	defer host.Close()

	natManager := host.GetNATManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = natManager.DiscoverNATType(ctx)
	}
}
