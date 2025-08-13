package nat

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// Benchmark tests to validate NAT traversal performance optimizations

func BenchmarkNATTraversalManager_Creation(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewNATTraversalManager(ctx, nil)
		manager.Close()
	}
}

func BenchmarkNATTraversalManager_AddSTUNServers(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.AddSTUNServer("stun.example.com", 19302)
	}
}

func BenchmarkNATTraversalManager_AddTURNServers(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.AddTURNServer("turn.example.com", 3478, "user", "pass", "realm", "udp")
	}
}

func BenchmarkNATTraversalManager_DiscoveryWithCache(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Pre-populate cache
	manager.cacheMux.Lock()
	manager.discoveryCache["nat_type"] = &DiscoveryResult{
		NATType:   NATTypeFullCone,
		Timestamp: time.Now(),
		RTT:       100 * time.Millisecond,
	}
	manager.cacheMux.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.DiscoverNATType(ctx)
	}
}

func BenchmarkNATTraversalManager_TURNServerSelection(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Add multiple TURN servers
	for i := 0; i < 10; i++ {
		manager.AddTURNServer("turn.example.com", 3478+i, "user", "pass", "realm", "udp")
		manager.turnServers[i].Priority = 100 - i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.selectBestTURNServer()
	}
}

func BenchmarkNATTraversalManager_HolePunchingOptimized(b *testing.B) {
	ctx := context.Background()

	config := DefaultTraversalConfig()
	config.HolePunchRetries = 2                   // Reduce for benchmark
	config.BackoffInitial = 10 * time.Millisecond // Faster backoff

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

func BenchmarkNATTraversalManager_ConnectionPooling(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Pre-populate connection pool
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("server%d:3478", i)
		manager.relayConnections[key] = &RelayConnection{
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
			InUse:     false,
			Conn:      &mockNetConn{},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.cleanupConnectionPool()
	}
}

func BenchmarkNATTraversalManager_MetricsUpdate(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.metrics.STUNRequests++
		manager.metrics.STUNSuccesses++
		manager.metrics.TURNRequests++
		manager.metrics.RelayConnections++
		manager.metrics.LastDiscovery = time.Now()
	}
}

func BenchmarkNATTraversalManager_CacheOperations(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	result := &DiscoveryResult{
		NATType:   NATTypeRestrictedCone,
		Timestamp: time.Now(),
		RTT:       150 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.cacheMux.Lock()
		manager.discoveryCache["test"] = result
		_, exists := manager.discoveryCache["test"]
		if exists {
			delete(manager.discoveryCache, "test")
		}
		manager.cacheMux.Unlock()
	}
}

func BenchmarkNATTraversalManager_BackoffCalculation(b *testing.B) {
	config := DefaultTraversalConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backoff := config.BackoffInitial
		for j := 0; j < 10; j++ {
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.BackoffMax {
				backoff = config.BackoffMax
			}
		}
	}
}

func BenchmarkNATTraversalManager_ConcurrentAccess(b *testing.B) {
	ctx := context.Background()
	manager := NewNATTraversalManager(ctx, nil)
	defer manager.Close()

	// Add some servers
	manager.AddSTUNServer("stun1.example.com", 19302)
	manager.AddSTUNServer("stun2.example.com", 19302)
	manager.AddTURNServer("turn1.example.com", 3478, "user", "pass", "realm", "udp")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent operations
			_ = manager.GetNATType()
			_ = manager.GetPublicAddress()
			_ = manager.GetMetrics()
			_ = manager.IsRelayRequired()
			_ = manager.selectBestTURNServer()
		}
	})
}

// Memory allocation benchmarks
func BenchmarkNATTraversalManager_MemoryAllocation(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := NewNATTraversalManager(ctx, nil)
		manager.AddSTUNServer("stun.example.com", 19302)
		manager.AddTURNServer("turn.example.com", 3478, "user", "pass", "realm", "udp")
		manager.Close()
	}
}

func BenchmarkNATType_String(b *testing.B) {
	natTypes := []NATType{
		NATTypeOpen,
		NATTypeFullCone,
		NATTypeRestrictedCone,
		NATTypePortRestrictedCone,
		NATTypeSymmetric,
		NATTypeBlocked,
		NATTypeUnknown,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, natType := range natTypes {
			_ = natType.String()
		}
	}
}

// Performance comparison benchmarks
func BenchmarkNATTraversalManager_OptimizedVsBasic(b *testing.B) {
	ctx := context.Background()

	b.Run("Optimized", func(b *testing.B) {
		config := DefaultTraversalConfig()
		config.ConnectTimeout = 5 * time.Second // Optimized
		config.ParallelAttempts = 3             // Parallel
		config.EarlySuccessDelay = 200 * time.Millisecond

		manager := NewNATTraversalManager(ctx, config)
		defer manager.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate optimized operations
			_ = manager.GetNATType()
		}
	})

	b.Run("Basic", func(b *testing.B) {
		config := DefaultTraversalConfig()
		config.ConnectTimeout = 30 * time.Second // Traditional
		config.ParallelAttempts = 1              // Sequential
		config.EarlySuccessDelay = 0             // No early success

		manager := NewNATTraversalManager(ctx, config)
		defer manager.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate basic operations
			_ = manager.GetNATType()
		}
	})
}

// Example usage for documentation
func ExampleNATTraversalManager_optimizedUsage() {
	ctx := context.Background()

	// Create manager with optimized configuration
	config := DefaultTraversalConfig()
	config.ConnectTimeout = 5 * time.Second // Reduced from 30s
	config.ParallelAttempts = 3             // Parallel connections
	config.EarlySuccessDelay = 200 * time.Millisecond

	manager := NewNATTraversalManager(ctx, config)
	defer manager.Close()

	// Add STUN servers for NAT discovery
	manager.AddSTUNServer("stun.l.google.com", 19302)
	manager.AddSTUNServer("stun1.l.google.com", 19302)

	// Add TURN servers for relay when needed
	manager.AddTURNServer("turn.example.com", 3478, "username", "password", "realm", "udp")

	// Discover NAT type
	_, err := manager.DiscoverNATType(ctx)
	if err != nil {
		return
	}

	// Check if relay is required
	if manager.IsRelayRequired() {
		// Use TURN relay for connections
		_, _ = manager.EstablishRelayConnection(ctx, "peer-id")
	} else {
		// Direct connection or hole punching possible
		// Use optimized connection strategies
	}

	// Monitor performance
	metrics := manager.GetMetrics()
	_ = metrics.STUNRequests
	_ = metrics.RelayConnections
	_ = metrics.SuccessfulHoles
}
