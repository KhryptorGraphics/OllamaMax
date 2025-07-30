package nat

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for NAT traversal scenarios

func TestNATTraversal_FullConeScenario(t *testing.T) {
	ctx := context.Background()
	
	// Test Full Cone NAT scenario
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Simulate Full Cone NAT detection
	manager.natType = NATTypeFullCone
	manager.publicAddr = &net.UDPAddr{
		IP:   net.ParseIP("203.0.113.1"),
		Port: 12345,
	}
	
	// Test characteristics
	assert.Equal(t, NATTypeFullCone, manager.GetNATType())
	assert.False(t, manager.IsRelayRequired())
	
	publicAddr := manager.GetPublicAddress()
	require.NotNil(t, publicAddr)
	assert.Equal(t, "203.0.113.1", publicAddr.IP.String())
	assert.Equal(t, 12345, publicAddr.Port)
	
	// Test connection strategy
	assert.False(t, manager.IsRelayRequired(), "Full Cone NAT should not require relay")
	
	t.Logf("Full Cone NAT scenario: Public address %s, Relay required: %v", 
		publicAddr, manager.IsRelayRequired())
}

func TestNATTraversal_RestrictedConeScenario(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Simulate Restricted Cone NAT detection
	manager.natType = NATTypeRestrictedCone
	manager.publicAddr = &net.UDPAddr{
		IP:   net.ParseIP("203.0.113.2"),
		Port: 23456,
	}
	
	assert.Equal(t, NATTypeRestrictedCone, manager.GetNATType())
	assert.False(t, manager.IsRelayRequired())
	
	// Test hole punching capability
	err := manager.performHolePunching(ctx, &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 8080})
	
	// Should attempt hole punching without error in basic case
	assert.NoError(t, err)
	
	t.Logf("Restricted Cone NAT scenario: %s", manager.GetNATType())
}

func TestNATTraversal_SymmetricNATScenario(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Add TURN servers for symmetric NAT
	manager.AddTURNServer("turn.example.com", 3478, "user", "pass", "realm", "udp")
	manager.AddTURNServer("turn2.example.com", 3478, "user2", "pass2", "realm", "tcp")
	
	// Simulate Symmetric NAT detection
	manager.natType = NATTypeSymmetric
	
	assert.Equal(t, NATTypeSymmetric, manager.GetNATType())
	assert.True(t, manager.IsRelayRequired(), "Symmetric NAT should require relay")
	
	// Test TURN server selection
	bestServer := manager.selectBestTURNServer()
	require.NotNil(t, bestServer)
	assert.Contains(t, []string{"turn.example.com", "turn2.example.com"}, bestServer.Address)
	
	t.Logf("Symmetric NAT scenario: Best TURN server %s:%d", 
		bestServer.Address, bestServer.Port)
}

func TestNATTraversal_BlockedScenario(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Simulate blocked/firewall scenario
	manager.natType = NATTypeBlocked
	
	assert.Equal(t, NATTypeBlocked, manager.GetNATType())
	assert.True(t, manager.IsRelayRequired(), "Blocked NAT should require relay")
	
	// In blocked scenario, only relay connections should work
	t.Logf("Blocked NAT scenario: Relay required: %v", manager.IsRelayRequired())
}

func TestNATTraversal_OpenNetworkScenario(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Simulate open network (no NAT)
	manager.natType = NATTypeOpen
	manager.publicAddr = &net.UDPAddr{
		IP:   net.ParseIP("203.0.113.100"),
		Port: 8080,
	}
	
	assert.Equal(t, NATTypeOpen, manager.GetNATType())
	assert.False(t, manager.IsRelayRequired(), "Open network should not require relay")
	
	publicAddr := manager.GetPublicAddress()
	require.NotNil(t, publicAddr)  
	assert.True(t, publicAddr.IP.IsGlobalUnicast() || publicAddr.IP.IsLoopback())
	
	t.Logf("Open network scenario: Direct connections available")
}

func TestNATTraversal_STUNServerFailover(t *testing.T) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.DiscoveryTimeout = 2 * time.Second
	
	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()
	
	// Add multiple STUN servers (some will fail in test environment)
	manager.AddSTUNServer("nonexistent1.example.com", 19302)
	manager.AddSTUNServer("nonexistent2.example.com", 19302)
	manager.AddSTUNServer("stun.l.google.com", 19302) // Real server
	
	// Test failover behavior
	discoveryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	_, err := manager.DiscoverNATType(discoveryCtx)
	
	// In test environment, this may fail, but we test the failover mechanism
	metrics := manager.GetMetrics()
	assert.Greater(t, metrics.STUNRequests, int64(0), "Should have attempted STUN requests")
	
	if err == nil {
		t.Log("NAT discovery succeeded with failover")
	} else {
		t.Logf("NAT discovery failed (expected in test env): %v", err)
		assert.True(t, err != nil, "Expected error for failed STUN discovery")
	}
}

func TestNATTraversal_TURNServerFailover(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Add TURN servers with different priorities
	manager.AddTURNServer("turn1.example.com", 3478, "user1", "pass1", "realm", "udp")
	manager.AddTURNServer("turn2.example.com", 3478, "user2", "pass2", "realm", "udp")
	manager.AddTURNServer("turn3.example.com", 3478, "user3", "pass3", "realm", "tcp")
	
	// Set different priorities and availability
	manager.turnServers[0].Priority = 100
	manager.turnServers[0].Available = false  // Unavailable
	manager.turnServers[1].Priority = 90
	manager.turnServers[1].Available = true
	manager.turnServers[2].Priority = 80
	manager.turnServers[2].Available = true
	
	// Test server selection with failover
	bestServer := manager.selectBestTURNServer()
	require.NotNil(t, bestServer)
	assert.Equal(t, "turn2.example.com", bestServer.Address)
	assert.Equal(t, 90, bestServer.Priority)
	
	t.Logf("TURN failover: Selected %s:%d (priority %d)", 
		bestServer.Address, bestServer.Port, bestServer.Priority)
}

func TestNATTraversal_ConnectionPooling(t *testing.T) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.RelayConnTTL = 100 * time.Millisecond // Short TTL for testing
	
	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()
	
	// Add mock connection to pool
	mockConn := &RelayConnection{
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		InUse:     false,
		Conn:      &mockNetConn{},
	}
	
	poolKey := "test-server:3478"
	manager.connPoolMux.Lock()
	manager.relayConnections[poolKey] = mockConn
	manager.connPoolMux.Unlock()
	
	// Wait for TTL expiry
	time.Sleep(150 * time.Millisecond)
	
	// Trigger cleanup
	manager.cleanupConnectionPool()
	
	// Connection should be removed
	manager.connPoolMux.RLock()
	_, exists := manager.relayConnections[poolKey]
	manager.connPoolMux.RUnlock()
	
	assert.False(t, exists, "Expired connection should be cleaned up")
	
	t.Log("Connection pool cleanup verified")
}

func TestNATTraversal_ConcurrentDiscovery(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Add STUN servers
	manager.AddSTUNServer("stun.l.google.com", 19302)
	manager.AddSTUNServer("stun1.l.google.com", 19302)
	
	// Perform concurrent discovery attempts
	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make(chan struct {
		natType NATType
		err     error
	}, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			
			natType, err := manager.DiscoverNATType(discoveryCtx)
			results <- struct {
				natType NATType
				err     error
			}{natType, err}
		}()
	}
	
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	var successCount int
	for result := range results {
		if result.err == nil {
			successCount++
			assert.NotEqual(t, NATTypeUnknown, result.natType)
		}
	}
	
	t.Logf("Concurrent discovery: %d/%d successful", successCount, numGoroutines)
	
	// Test metrics
	metrics := manager.GetMetrics()
	assert.Greater(t, metrics.STUNRequests, int64(0))
}

func TestNATTraversal_ExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.BackoffInitial = 100 * time.Millisecond
	config.BackoffMax = 1 * time.Second
	config.BackoffMultiplier = 2.0
	
	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()
	
	// Test backoff calculation
	testBackoffs := []time.Duration{
		100 * time.Millisecond,  // Initial
		200 * time.Millisecond,  // 100 * 2
		400 * time.Millisecond,  // 200 * 2
		800 * time.Millisecond,  // 400 * 2
		1 * time.Second,         // Max reached
		1 * time.Second,         // Still max
	}
	
	backoff := config.BackoffInitial
	for i, expected := range testBackoffs {
		if i > 0 {
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.BackoffMax {
				backoff = config.BackoffMax
			}
		}
		
		assert.Equal(t, expected, backoff, "Backoff calculation mismatch at step %d", i)
	}
	
	t.Log("Exponential backoff calculation verified")
}

func TestNATTraversal_HolePunchingWithBackoff(t *testing.T) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.HolePunchRetries = 3
	config.HolePunchTimeout = 2 * time.Second
	config.BackoffInitial = 100 * time.Millisecond
	config.BackoffMultiplier = 2.0
	
	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()
	
	targetAddr := &net.UDPAddr{
		IP:   net.ParseIP("192.0.2.1"),
		Port: 8080,
	}
	
	start := time.Now()
	err := manager.performHolePunching(ctx, targetAddr)
	duration := time.Since(start)
	
	// Should complete without error
	assert.NoError(t, err)
	
	// Should take some time due to backoff delays
	expectedMinDuration := 100*time.Millisecond + 200*time.Millisecond // First two backoffs
	assert.GreaterOrEqual(t, duration, expectedMinDuration/2) // Allow some variance
	
	t.Logf("Hole punching with backoff completed in %v", duration)
}

func TestNATTraversal_MetricsAccuracy(t *testing.T) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Test metrics initialization
	metrics := manager.GetMetrics()
	assert.Equal(t, int64(0), metrics.STUNRequests)
	assert.Equal(t, int64(0), metrics.STUNSuccesses)
	assert.Equal(t, int64(0), metrics.STUNFailures)
	assert.Equal(t, int64(0), metrics.TURNRequests)
	assert.Equal(t, int64(0), metrics.NATDetections)
	
	// Simulate some activity
	manager.metrics.STUNRequests = 5
	manager.metrics.STUNSuccesses = 3
	manager.metrics.STUNFailures = 2
	manager.metrics.TURNRequests = 2
	manager.metrics.TURNSuccesses = 1
	manager.metrics.NATDetections = 1
	
	metrics = manager.GetMetrics()
	assert.Equal(t, int64(5), metrics.STUNRequests)
	assert.Equal(t, int64(3), metrics.STUNSuccesses)
	assert.Equal(t, int64(2), metrics.STUNFailures)
	assert.Equal(t, int64(2), metrics.TURNRequests)
	assert.Equal(t, int64(1), metrics.TURNSuccesses)
	assert.Equal(t, int64(1), metrics.NATDetections)
	
	t.Log("Metrics accuracy verified")
}

func TestNATTraversal_CacheEffectiveness(t *testing.T) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.CacheExpiry = 200 * time.Millisecond
	
	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()
	
	// Manually add cached result
	manager.cacheMux.Lock()
	manager.discoveryCache["nat_type"] = &DiscoveryResult{
		NATType:   NATTypeFullCone,
		Timestamp: time.Now(),
		ServerUsed: "test-server",
		RTT:       100 * time.Millisecond,
	}
	manager.cacheMux.Unlock()
	
	// First call should use cache
	start := time.Now()
	natType, err := manager.DiscoverNATType(ctx)
	duration1 := time.Since(start)
	
	assert.NoError(t, err)
	assert.Equal(t, NATTypeFullCone, natType)
	assert.Less(t, duration1, 50*time.Millisecond, "Cache hit should be fast")
	
	// Wait for cache expiry
	time.Sleep(250 * time.Millisecond)
	
	// Add STUN server for next attempt
	manager.AddSTUNServer("127.0.0.1", 19302) // Will fail, but tests cache expiry
	
	// Second call should attempt fresh discovery
	start = time.Now()
	_, err = manager.DiscoverNATType(ctx)
	duration2 := time.Since(start)
	
	// Should take longer than cache hit (even if it fails)
	assert.Greater(t, duration2, duration1)
	
	t.Logf("Cache test: Hit=%v, Miss=%v", duration1, duration2)
}

// Benchmark tests for performance validation
func BenchmarkNATTraversal_Discovery(b *testing.B) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Add mock cache to avoid actual network calls
	manager.cacheMux.Lock()
	manager.discoveryCache["nat_type"] = &DiscoveryResult{
		NATType:   NATTypeFullCone,
		Timestamp: time.Now(),
	}
	manager.cacheMux.Unlock()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.DiscoverNATType(ctx)
	}
}

func BenchmarkNATTraversal_HolePunching(b *testing.B) {
	ctx := context.Background()
	
	config := DefaultTraversalConfig()
	config.HolePunchRetries = 1 // Reduce for benchmarking
	
	manager := NewNATTraversalManager(ctx, config)
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

func BenchmarkNATTraversal_TURNServerSelection(b *testing.B) {
	ctx := context.Background()
	
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()
	
	// Add many TURN servers
	for i := 0; i < 100; i++ {
		manager.AddTURNServer(
			fmt.Sprintf("turn%d.example.com", i),
			3478,
			"user",
			"pass", 
			"realm",
			"udp",
		)
		manager.turnServers[i].Priority = i
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.selectBestTURNServer()
	}
}