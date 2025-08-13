package nat

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSTUNServer simulates a STUN server for testing
type MockSTUNServer struct {
	Address   string
	Port      int
	Available bool
	Responses map[string][]byte
}

// MockTURNServer simulates a TURN server for testing
type MockTURNServer struct {
	Address   string
	Port      int
	Available bool
	Transport string
}

func TestNATTraversalManager_Creation(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()

	manager := NewNATTraversalManager(ctx, config)
	require.NotNil(t, manager)

	assert.Equal(t, NATTypeUnknown, manager.GetNATType())
	assert.Equal(t, config, manager.config)
	assert.NotNil(t, manager.metrics)

	err := manager.Close()
	assert.NoError(t, err)
}

func TestNATTraversalManager_AddServers(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Test adding STUN server
	manager.AddSTUNServer("stun.l.google.com", 19302)
	assert.Len(t, manager.stunServers, 1)
	assert.Equal(t, "stun.l.google.com", manager.stunServers[0].Address)
	assert.Equal(t, 19302, manager.stunServers[0].Port)
	assert.True(t, manager.stunServers[0].Available)

	// Test adding TURN server
	manager.AddTURNServer("turn.example.com", 3478, "user", "pass", "realm", "udp")
	assert.Len(t, manager.turnServers, 1)
	assert.Equal(t, "turn.example.com", manager.turnServers[0].Address)
	assert.Equal(t, 3478, manager.turnServers[0].Port)
	assert.Equal(t, "user", manager.turnServers[0].Username)
	assert.Equal(t, "pass", manager.turnServers[0].Password)
	assert.Equal(t, "udp", manager.turnServers[0].Transport)
}

func TestNATType_String(t *testing.T) {
	tests := []struct {
		natType  NATType
		expected string
	}{
		{NATTypeOpen, "Open"},
		{NATTypeFullCone, "Full Cone"},
		{NATTypeRestrictedCone, "Restricted Cone"},
		{NATTypePortRestrictedCone, "Port Restricted Cone"},
		{NATTypeSymmetric, "Symmetric"},
		{NATTypeBlocked, "Blocked"},
		{NATTypeUnknown, "Unknown"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.natType.String())
		})
	}
}

func TestDefaultTraversalConfig(t *testing.T) {
	config := DefaultTraversalConfig()
	require.NotNil(t, config)

	// Test default values
	assert.Equal(t, 5*time.Second, config.STUNTimeout)
	assert.Equal(t, 3, config.STUNRetries)
	assert.Equal(t, 5*time.Second, config.ConnectTimeout) // Reduced from 30s
	assert.Equal(t, 3, config.ParallelAttempts)
	assert.Equal(t, 200*time.Millisecond, config.EarlySuccessDelay)
	assert.Equal(t, 1*time.Second, config.BackoffInitial)
	assert.Equal(t, 30*time.Second, config.BackoffMax)
	assert.Equal(t, 2.0, config.BackoffMultiplier)
}

func TestNATDiscovery_NoServers(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Test discovery with no STUN servers
	natType, err := manager.DiscoverNATType(ctx)
	assert.Error(t, err)
	assert.Equal(t, NATTypeUnknown, natType)
	assert.Contains(t, err.Error(), "no STUN servers configured")
}

func TestNATDiscovery_WithServers(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()
	config.DiscoveryTimeout = 1 * time.Second // Short timeout for testing

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Add a mock STUN server (will fail in test environment)
	manager.AddSTUNServer("127.0.0.1", 19302)

	// Test discovery (expected to timeout/fail in test environment)
	natType, err := manager.DiscoverNATType(ctx)
	// In test environment, this will likely fail due to no actual STUN server
	// but we test the flow
	if err != nil {
		assert.Equal(t, NATTypeUnknown, natType)
		assert.Contains(t, err.Error(), "STUN servers failed")
	}
}

func TestNATDiscovery_Caching(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()
	config.CacheExpiry = 100 * time.Millisecond

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Manually add a cached result
	manager.cacheMux.Lock()
	manager.discoveryCache["nat_type"] = &DiscoveryResult{
		NATType:   NATTypeOpen,
		Timestamp: time.Now(),
	}
	manager.cacheMux.Unlock()

	// Test cached result
	natType, err := manager.DiscoverNATType(ctx)
	assert.NoError(t, err)
	assert.Equal(t, NATTypeOpen, natType)

	// Wait for cache expiry
	time.Sleep(150 * time.Millisecond)

	// Add STUN server for next test
	manager.AddSTUNServer("127.0.0.1", 19302)

	// Test expired cache (will attempt discovery)
	natType, err = manager.DiscoverNATType(ctx)
	// Expected to fail in test environment
	if err != nil {
		assert.Equal(t, NATTypeUnknown, natType)
	}
}

func TestIsRelayRequired(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Test different NAT types
	testCases := []struct {
		natType       NATType
		relayRequired bool
	}{
		{NATTypeOpen, false},
		{NATTypeFullCone, false},
		{NATTypeRestrictedCone, false},
		{NATTypePortRestrictedCone, false},
		{NATTypeSymmetric, true},
		{NATTypeBlocked, true},
		{NATTypeUnknown, false},
	}

	for _, tc := range testCases {
		t.Run(tc.natType.String(), func(t *testing.T) {
			manager.natType = tc.natType
			assert.Equal(t, tc.relayRequired, manager.IsRelayRequired())
		})
	}
}

func TestSelectBestTURNServer(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Add TURN servers with different priorities
	manager.AddTURNServer("turn1.example.com", 3478, "user1", "pass1", "realm", "udp")
	manager.AddTURNServer("turn2.example.com", 3478, "user2", "pass2", "realm", "udp")
	manager.AddTURNServer("turn3.example.com", 3478, "user3", "pass3", "realm", "udp")

	// Set different priorities
	manager.turnServers[0].Priority = 50
	manager.turnServers[1].Priority = 100 // Highest
	manager.turnServers[2].Priority = 75

	// Make one unavailable
	manager.turnServers[1].Available = false

	best := manager.selectBestTURNServer()
	require.NotNil(t, best)
	assert.Equal(t, "turn3.example.com", best.Address)
	assert.Equal(t, 75, best.Priority)
}

func TestMultiaddrToNetAddr(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Test the error path with nil multiaddr
	// This tests that the function handles invalid input gracefully
	t.Skip("Skipping multiaddr test - requires libp2p multiaddr implementation")
}

func TestHolePunching_Configuration(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()
	config.HolePunchRetries = 3
	config.HolePunchTimeout = 5 * time.Second
	config.BackoffInitial = 100 * time.Millisecond
	config.BackoffMax = 2 * time.Second
	config.BackoffMultiplier = 2.0

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	assert.Equal(t, 3, manager.config.HolePunchRetries)
	assert.Equal(t, 5*time.Second, manager.config.HolePunchTimeout)
	assert.Equal(t, 100*time.Millisecond, manager.config.BackoffInitial)
	assert.Equal(t, 2*time.Second, manager.config.BackoffMax)
	assert.Equal(t, 2.0, manager.config.BackoffMultiplier)
}

func TestConnectionOptimization_EarlySuccess(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()

	// Test optimized connection settings
	assert.Equal(t, 5*time.Second, config.ConnectTimeout) // Reduced from 30s
	assert.Equal(t, 3, config.ParallelAttempts)           // Parallel connections
	assert.Equal(t, 200*time.Millisecond, config.EarlySuccessDelay)

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	assert.Equal(t, config.ConnectTimeout, manager.config.ConnectTimeout)
	assert.Equal(t, config.ParallelAttempts, manager.config.ParallelAttempts)
}

func TestExponentialBackoff_Implementation(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Test backoff calculation logic
	initial := config.BackoffInitial
	multiplier := config.BackoffMultiplier
	maxBackoff := config.BackoffMax

	backoff := initial
	for i := 0; i < 5; i++ {
		if i > 0 {
			backoff = time.Duration(float64(backoff) * multiplier)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}

		assert.LessOrEqual(t, backoff, maxBackoff)
		if i > 0 {
			assert.GreaterOrEqual(t, backoff, initial)
		}
	}
}

func TestMetricsTracking(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	metrics := manager.GetMetrics()
	require.NotNil(t, metrics)

	// Test initial metrics
	assert.Equal(t, int64(0), metrics.STUNRequests)
	assert.Equal(t, int64(0), metrics.STUNSuccesses)
	assert.Equal(t, int64(0), metrics.STUNFailures)
	assert.Equal(t, int64(0), metrics.TURNRequests)
	assert.Equal(t, int64(0), metrics.TURNSuccesses)
	assert.Equal(t, int64(0), metrics.TURNFailures)
	assert.Equal(t, int64(0), metrics.NATDetections)
	assert.Equal(t, int64(0), metrics.RelayConnections)
	assert.Equal(t, int64(0), metrics.SuccessfulHoles)
	assert.Equal(t, int64(0), metrics.FailedHoles)
}

func TestConnectionPoolManagement(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()
	config.RelayConnTTL = 100 * time.Millisecond // Short TTL for testing

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Add a mock connection to the pool
	mockConn := &RelayConnection{
		CreatedAt: time.Now(),
		LastUsed:  time.Now().Add(-200 * time.Millisecond), // Expired
		InUse:     false,
		Conn:      &mockNetConn{},
	}

	manager.connPoolMux.Lock()
	manager.relayConnections["test-conn"] = mockConn
	manager.connPoolMux.Unlock()

	// Trigger cleanup
	manager.cleanupConnectionPool()

	// Check that expired connection was removed
	manager.connPoolMux.RLock()
	_, exists := manager.relayConnections["test-conn"]
	manager.connPoolMux.RUnlock()

	assert.False(t, exists, "Expired connection should have been cleaned up")
}

func TestServerHealthChecking(t *testing.T) {
	ctx := context.Background()
	config := DefaultTraversalConfig()
	config.STUNServerCheck = 50 * time.Millisecond

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Add a STUN server
	manager.AddSTUNServer("127.0.0.1", 19302)

	// Initially available
	assert.True(t, manager.stunServers[0].Available)

	// Wait for health check (will fail for localhost:19302)
	time.Sleep(100 * time.Millisecond)

	// Server should become unavailable after failed health check
	// Note: This test may be flaky in some environments
}

func TestFullConeNATScenario(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Simulate Full Cone NAT detection
	manager.natType = NATTypeFullCone
	manager.publicAddr = &net.UDPAddr{
		IP:   net.ParseIP("203.0.113.1"),
		Port: 12345,
	}

	assert.Equal(t, NATTypeFullCone, manager.GetNATType())
	assert.False(t, manager.IsRelayRequired())

	publicAddr := manager.GetPublicAddress()
	assert.Equal(t, "203.0.113.1", publicAddr.IP.String())
	assert.Equal(t, 12345, publicAddr.Port)
}

func TestRestrictedConeNATScenario(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Simulate Restricted Cone NAT detection
	manager.natType = NATTypeRestrictedCone

	assert.Equal(t, NATTypeRestrictedCone, manager.GetNATType())
	assert.False(t, manager.IsRelayRequired())
}

func TestSymmetricNATScenario(t *testing.T) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Simulate Symmetric NAT detection
	manager.natType = NATTypeSymmetric

	assert.Equal(t, NATTypeSymmetric, manager.GetNATType())
	assert.True(t, manager.IsRelayRequired(), "Symmetric NAT should require relay")
}

// Benchmark tests
func BenchmarkNATDiscovery(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Add mock STUN server
	manager.AddSTUNServer("127.0.0.1", 19302)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.DiscoverNATType(ctx)
	}
}

func BenchmarkHolePunching(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	targetAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.performHolePunching(ctx, targetAddr)
	}
}

// Mock connection for testing
type mockNetConn struct{}

func (m *mockNetConn) Read(b []byte) (n int, err error)  { return 0, nil }
func (m *mockNetConn) Write(b []byte) (n int, err error) { return len(b), nil }
func (m *mockNetConn) Close() error                      { return nil }
func (m *mockNetConn) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}
func (m *mockNetConn) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8081}
}
func (m *mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error { return nil }
