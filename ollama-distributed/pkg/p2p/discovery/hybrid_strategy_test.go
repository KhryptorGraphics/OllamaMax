package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/discovery"
)

// mockDiscoveryConfig implements DiscoveryConfig for testing
type mockDiscoveryConfig struct {
	bootstrapPeers []string
	rendezvousStr  string
	autoDiscovery  bool
}

func (m *mockDiscoveryConfig) GetBootstrapPeers() []string {
	return m.bootstrapPeers
}

func (m *mockDiscoveryConfig) GetRendezvousString() string {
	return m.rendezvousStr
}

func (m *mockDiscoveryConfig) IsAutoDiscoveryEnabled() bool {
	return m.autoDiscovery
}

// mockDiscoveryStrategy implements DiscoveryStrategy for testing
type mockDiscoveryStrategy struct {
	name      string
	peers     []peer.AddrInfo
	advertiseFunc func(context.Context, string, ...discovery.Option) (time.Duration, error)
	findFunc  func(context.Context, string, ...discovery.Option) (<-chan peer.AddrInfo, error)
}

func (m *mockDiscoveryStrategy) Name() string {
	return m.name
}

func (m *mockDiscoveryStrategy) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, ns, opts...)
	}
	
	peerChan := make(chan peer.AddrInfo, len(m.peers))
	go func() {
		defer close(peerChan)
		for _, p := range m.peers {
			select {
			case peerChan <- p:
			case <-ctx.Done():
				return
			}
		}
	}()
	return peerChan, nil
}

func (m *mockDiscoveryStrategy) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	if m.advertiseFunc != nil {
		return m.advertiseFunc(ctx, ns, opts...)
	}
	return time.Minute, nil
}

func TestHybridDiscoveryStrategy(t *testing.T) {
	// Create mock config
	config := &mockDiscoveryConfig{
		bootstrapPeers: []string{},
		rendezvousStr:  "test",
		autoDiscovery:  true,
	}
	
	// Create hybrid strategy
	hybrid := &HybridDiscoveryStrategy{
		config: config,
		performanceMetrics: make(map[string]*StrategyPerformance),
		weights: map[string]float64{
			"strategy1": 0.5,
			"strategy2": 0.5,
		},
		lastUsed: make(map[string]time.Time),
		dhtStrategy: &DHTStrategy{}, // Empty for testing
	}
	
	// Test name
	if name := hybrid.Name(); name != "hybrid" {
		t.Errorf("Expected name 'hybrid', got '%s'", name)
	}
	
	// Test weights
	weights := hybrid.GetWeights()
	if len(weights) != 2 {
		t.Errorf("Expected 2 weights, got %d", len(weights))
	}
	
	// Test performance metrics
	metrics := hybrid.GetPerformanceMetrics()
	// The metrics map is initialized empty, so we expect 0 entries initially
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics entries, got %d", len(metrics))
	}
	
	// Test last used
	lastUsed := hybrid.GetLastUsed()
	if len(lastUsed) != 0 {
		t.Errorf("Expected 0 last used entries, got %d", len(lastUsed))
	}
	
	// Test adjust weights
	hybrid.performanceMetrics["strategy1"] = &StrategyPerformance{
		SuccessCount: 10,
		FailureCount: 0,
	}
	hybrid.performanceMetrics["strategy2"] = &StrategyPerformance{
		SuccessCount: 5,
		FailureCount: 5,
	}
	
	hybrid.AdjustWeights()
	weights = hybrid.GetWeights()
	
	// Strategy1 should have higher weight due to better success rate
	if weights["strategy1"] <= weights["strategy2"] {
		t.Errorf("Expected strategy1 to have higher weight than strategy2")
	}
}

func TestPeerCache(t *testing.T) {
	cache := NewPeerCache(3, time.Minute)
	
	peer1 := peer.AddrInfo{ID: "peer1"}
	peer2 := peer.AddrInfo{ID: "peer2"}
	peer3 := peer.AddrInfo{ID: "peer3"}
	
	// Add peers
	cache.Add(peer1, "test")
	cache.Add(peer2, "test")
	cache.Add(peer3, "test")
	
	// Check cache size
	if len(cache.peers) != 3 {
		t.Errorf("Expected cache size 3, got %d", len(cache.peers))
	}
	
	// Get peer
	if cachedPeer, exists := cache.Get("peer1"); !exists {
		t.Error("Expected peer1 to exist")
	} else if cachedPeer.AddrInfo.ID != "peer1" {
		t.Error("Expected peer1 ID to match")
	}
	
	// Update performance score
	cache.UpdatePerformanceScore("peer1", 2.0)
	if score, exists := cache.GetPerformanceScore("peer1"); !exists {
		t.Error("Expected performance score to exist")
	} else if score != 2.0 {
		t.Errorf("Expected score 2.0, got %f", score)
	}
	
	// Get top performers
	performers := cache.GetTopPerformers(2)
	// We expect 1 performer since only peer1 has a score
	// Note: GetTopPerformers might return additional peers if they exist in the cache
	// but don't have explicit scores (defaulting to 0.0), so we check that peer1 is in the list
	found := false
	for _, p := range performers {
		if p == "peer1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected peer1 in top performers")
	}
}