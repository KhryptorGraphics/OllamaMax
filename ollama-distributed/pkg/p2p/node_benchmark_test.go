package p2p

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BenchmarkEventEmission tests event emission performance with bounded goroutine pool
func BenchmarkEventEmission(b *testing.B) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}

	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		b.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Stop()

	// Register test event handler
	handlerCalled := int64(0)
	node.On("test_event", func(event *NodeEvent) {
		// Simulate some work
		time.Sleep(1 * time.Millisecond)
		handlerCalled++
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			node.emitEvent("test_event", "test_data", "")
		}
	})
}

// BenchmarkConnectionPoolOperations tests connection pool performance
func BenchmarkConnectionPoolOperations(b *testing.B) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen:      []string{"/ip4/127.0.0.1/tcp/0"},
		ConnMgrHigh: 1000,
	}

	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		b.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Stop()

	// Generate test peer IDs
	testPeers := make([]peer.ID, 100)
	for i := 0; i < 100; i++ {
		testPeers[i] = peer.ID("test-peer-" + string(rune(i)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			peerID := testPeers[i%len(testPeers)]
			
			// Test connection pool operations
			node.canAcceptConnection(peerID)
			node.addToConnectionPool(peerID)
			node.removeFromConnectionPool(peerID)
			
			i++
		}
	})
}

// BenchmarkResourceMetricsUpdate tests optimized resource monitoring
func BenchmarkResourceMetricsUpdate(b *testing.B) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}

	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		b.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.updateResourceMetrics()
	}
}

// TestMemoryUsageOptimization tests memory usage improvements
func TestMemoryUsageOptimization(t *testing.T) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen:      []string{"/ip4/127.0.0.1/tcp/0"},
		ConnMgrHigh: 100,
	}

	// Measure memory before
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create and use P2P node
	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		t.Fatalf("Failed to create P2P node: %v", err)
	}

	// Simulate heavy event activity
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.emitEvent("test_event", "data", "")
		}()
	}
	wg.Wait()

	// Simulate connection activity
	for i := 0; i < 50; i++ {
		peerID := peer.ID("test-peer-" + string(rune(i)))
		node.addToConnectionPool(peerID)
	}

	// Cleanup
	node.Stop()

	// Measure memory after
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Memory usage should be reasonable
	memoryIncrease := m2.Alloc - m1.Alloc
	t.Logf("Memory increase: %d bytes", memoryIncrease)

	// Should not exceed 10MB for this test
	if memoryIncrease > 10*1024*1024 {
		t.Errorf("Memory usage too high: %d bytes", memoryIncrease)
	}
}

// TestConnectionPoolLimits tests connection pool bounds
func TestConnectionPoolLimits(t *testing.T) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen:      []string{"/ip4/127.0.0.1/tcp/0"},
		ConnMgrHigh: 10, // Small limit for testing
	}

	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		t.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Stop()

	// Add connections up to limit
	for i := 0; i < 10; i++ {
		peerID := peer.ID("test-peer-" + string(rune(i)))
		if !node.canAcceptConnection(peerID) {
			t.Errorf("Should accept connection %d", i)
		}
		node.addToConnectionPool(peerID)
	}

	// Should reject additional connections
	extraPeer := peer.ID("extra-peer")
	if node.canAcceptConnection(extraPeer) {
		t.Error("Should reject connection beyond limit")
	}

	// Cleanup should allow new connections
	node.cleanupConnectionPool()
	
	// After cleanup, should be able to add new connections
	if !node.canAcceptConnection(extraPeer) {
		t.Error("Should accept connection after cleanup")
	}
}

// TestEventHandlerBounds tests bounded goroutine pool for events
func TestEventHandlerBounds(t *testing.T) {
	ctx := context.Background()
	nodeConfig := &config.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	}

	node, err := NewP2PNode(ctx, nodeConfig)
	if err != nil {
		t.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Stop()

	// Register slow event handler
	handlerCount := int64(0)
	node.On("slow_event", func(event *NodeEvent) {
		time.Sleep(100 * time.Millisecond) // Slow handler
		handlerCount++
	})

	// Emit many events rapidly
	start := time.Now()
	for i := 0; i < 100; i++ {
		node.emitEvent("slow_event", "data", "")
	}

	// Should not block indefinitely due to bounded pool
	elapsed := time.Since(start)
	if elapsed > 5*time.Second {
		t.Errorf("Event emission took too long: %v", elapsed)
	}

	t.Logf("Event emission completed in: %v", elapsed)
}
