package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ollama/ollama-distributed/pkg/p2p"
	"github.com/ollama/ollama-distributed/pkg/p2p/discovery"
)

// TestDiscoveryEngine tests the P2P discovery engine
func TestDiscoveryEngine(t *testing.T) {
	// Create test configuration
	config := &p2p.NodeConfig{
		EnableDHT:    true,
		DHTMode:      "auto",
		Bootstrap:    []string{},
		ConnMgrLow:   2,
		ConnMgrHigh:  10,
	}

	// Create test host
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	host, err := p2p.NewHost(ctx, config)
	require.NoError(t, err)
	defer host.Close()

	// Create discovery engine
	engine, err := discovery.NewDiscoveryEngine(ctx, host, config)
	require.NoError(t, err)
	defer engine.Stop()

	t.Run("TestInitialization", func(t *testing.T) {
		metrics := engine.GetMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.StartTime.Before(time.Now()))
		assert.Empty(t, metrics.StrategyMetrics)
	})

	t.Run("TestPeerCache", func(t *testing.T) {
		cache := discovery.NewPeerCache(10, 5*time.Minute)
		
		// Create test peer
		testPeer := peer.AddrInfo{
			ID: peer.ID("test-peer"),
		}
		
		// Add peer to cache
		cache.Add(testPeer, "test")
		
		// Retrieve peer
		cachedPeer, exists := cache.Get(testPeer.ID)
		assert.True(t, exists)
		assert.Equal(t, testPeer.ID, cachedPeer.AddrInfo.ID)
		assert.Equal(t, "test", cachedPeer.Source)
		
		// Remove peer
		cache.Remove(testPeer.ID)
		_, exists = cache.Get(testPeer.ID)
		assert.False(t, exists)
	})

	t.Run("TestDiscoveryStrategies", func(t *testing.T) {
		engine.Start()
		time.Sleep(2 * time.Second)
		
		metrics := engine.GetMetrics()
		assert.GreaterOrEqual(t, len(metrics.StrategyMetrics), 1)
		
		// Check that strategies are registered
		for name, strategyMetrics := range metrics.StrategyMetrics {
			assert.NotEmpty(t, name)
			assert.NotNil(t, strategyMetrics)
		}
	})
}

// TestBootstrapDiscovery tests bootstrap discovery strategy
func TestBootstrapDiscovery(t *testing.T) {
	// Create test bootstrap peers
	bootstrapPeers := []peer.AddrInfo{
		{ID: peer.ID("bootstrap-1")},
		{ID: peer.ID("bootstrap-2")},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	host, err := p2p.NewHost(ctx, &p2p.NodeConfig{})
	require.NoError(t, err)
	defer host.Close()

	bootstrap := discovery.NewBootstrapDiscovery(host, bootstrapPeers, 5, 10)
	
	t.Run("TestName", func(t *testing.T) {
		assert.Equal(t, "bootstrap", bootstrap.Name())
	})

	t.Run("TestAdvertise", func(t *testing.T) {
		ttl, err := bootstrap.Advertise(ctx, "test-namespace")
		assert.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("TestFindPeers", func(t *testing.T) {
		peerChan, err := bootstrap.FindPeers(ctx, "test-namespace")
		assert.NoError(t, err)
		
		// Collect peers with timeout
		var peers []peer.AddrInfo
		timeout := time.After(5 * time.Second)
		
		for {
			select {
			case peer := <-peerChan:
				peers = append(peers, peer)
			case <-timeout:
				goto done
			}
		}
		
		done:
		assert.LessOrEqual(t, len(peers), len(bootstrapPeers))
	})
}

// TestMDNSDiscovery tests mDNS discovery strategy
func TestMDNSDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	host, err := p2p.NewHost(ctx, &p2p.NodeConfig{})
	require.NoError(t, err)
	defer host.Close()

	// Create mDNS discovery
	mdnsStrategy := discovery.NewMDNSStrategy(host)
	
	t.Run("TestName", func(t *testing.T) {
		assert.Equal(t, "mdns", mdnsStrategy.Name())
	})

	t.Run("TestAdvertise", func(t *testing.T) {
		ttl, err := mdnsStrategy.Advertise(ctx, "ollamacron")
		assert.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
	})

	t.Run("TestFindPeers", func(t *testing.T) {
		peerChan, err := mdnsStrategy.FindPeers(ctx, "ollamacron")
		assert.NoError(t, err)
		
		// Should not block even if no peers found
		select {
		case <-peerChan:
		case <-time.After(2 * time.Second):
		}
	})
}

// TestDiscoveryMetrics tests discovery metrics collection
func TestDiscoveryMetrics(t *testing.T) {
	metrics := &discovery.DiscoveryMetrics{
		StartTime:       time.Now(),
		StrategyMetrics: make(map[string]*discovery.StrategyMetrics),
	}

	// Test initial state
	assert.Equal(t, 0, metrics.PeersFound)
	assert.Equal(t, 0, metrics.PeersLost)
	assert.Equal(t, 0, metrics.ActivePeers)
	assert.Equal(t, 0, metrics.DiscoveryErrors)

	// Add strategy metrics
	metrics.StrategyMetrics["test"] = &discovery.StrategyMetrics{
		PeersFound:      5,
		Errors:          1,
		LastSuccess:     time.Now(),
		AverageLatency:  100 * time.Millisecond,
	}

	// Test strategy metrics
	strategyMetrics := metrics.StrategyMetrics["test"]
	assert.Equal(t, 5, strategyMetrics.PeersFound)
	assert.Equal(t, 1, strategyMetrics.Errors)
	assert.Equal(t, 100*time.Millisecond, strategyMetrics.AverageLatency)
}

// TestPeerQuality tests peer quality assessment
func TestPeerQuality(t *testing.T) {
	quality := &discovery.PeerQuality{
		Latency:     50 * time.Millisecond,
		Reliability: 0.95,
		Bandwidth:   1000000, // 1MB/s
		LastTest:    time.Now(),
	}

	// Test quality metrics
	assert.Equal(t, 50*time.Millisecond, quality.Latency)
	assert.Equal(t, 0.95, quality.Reliability)
	assert.Equal(t, int64(1000000), quality.Bandwidth)
	assert.True(t, quality.LastTest.Before(time.Now().Add(time.Second)))
}

// BenchmarkDiscoveryEngine benchmarks discovery engine performance
func BenchmarkDiscoveryEngine(b *testing.B) {
	config := &p2p.NodeConfig{
		EnableDHT: false, // Disable DHT for faster benchmarks
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	host, err := p2p.NewHost(ctx, config)
	require.NoError(b, err)
	defer host.Close()

	engine, err := discovery.NewDiscoveryEngine(ctx, host, config)
	require.NoError(b, err)
	defer engine.Stop()

	b.ResetTimer()

	b.Run("PeerCacheOperations", func(b *testing.B) {
		cache := discovery.NewPeerCache(1000, 5*time.Minute)
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				peerID := peer.ID(fmt.Sprintf("peer-%d", i))
				testPeer := peer.AddrInfo{ID: peerID}
				
				// Add peer
				cache.Add(testPeer, "benchmark")
				
				// Get peer
				_, exists := cache.Get(peerID)
				if !exists {
					b.Errorf("Peer should exist in cache")
				}
				
				i++
			}
		})
	})

	b.Run("MetricsUpdate", func(b *testing.B) {
		metrics := engine.GetMetrics()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				metrics.PeersFound++
				metrics.LastDiscovery = time.Now()
			}
		})
	})
}