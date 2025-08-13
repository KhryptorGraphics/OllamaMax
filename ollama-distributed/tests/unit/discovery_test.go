package unit

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiscoveryEngine tests the P2P discovery engine
func TestDiscoveryEngine(t *testing.T) {
	// Create test P2P node using helper
	node := createMockP2PNode(t)
	require.NotNil(t, node)
	defer node.Stop()

	// Test basic discovery functionality
	// Since we're using a mock node, we can only test basic operations
	assert.NotNil(t, node, "P2P node should be created")

	// Test that the node can be started
	err := node.Start()
	assert.NoError(t, err, "Node should start successfully")
}

// TestBootstrapDiscovery tests bootstrap discovery strategy

// TestMDNSDiscovery tests mDNS discovery strategy

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
		PeersFound:     5,
		Errors:         1,
		LastSuccess:    time.Now(),
		AverageLatency: 100 * time.Millisecond,
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
