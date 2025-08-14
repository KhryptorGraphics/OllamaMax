package discovery

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

// HybridDiscoveryStrategy combines multiple discovery strategies with weighted selection
type HybridDiscoveryStrategy struct {
	host   host.Host
	dht    *dht.IpfsDHT
	config DiscoveryConfig

	// Individual strategies
	dhtStrategy        *DHTStrategy
	mdnsStrategy       *MDNSStrategy
	bootstrapStrategy  *BootstrapDiscovery
	rendezvousStrategy *RendezvousDiscovery

	// Performance tracking
	performanceMetrics map[string]*StrategyPerformance
	metricsMutex       sync.RWMutex

	// Strategy weights (0.0 to 1.0)
	weights      map[string]float64
	weightsMutex sync.RWMutex

	// Last used timestamps
	lastUsed      map[string]time.Time
	lastUsedMutex sync.RWMutex
}

// StrategyPerformance tracks performance metrics for each strategy
type StrategyPerformance struct {
	SuccessCount   int
	FailureCount   int
	AverageLatency time.Duration
	LastSuccess    time.Time
	LastFailure    time.Time
}

// NewHybridDiscoveryStrategy creates a new hybrid discovery strategy
func NewHybridDiscoveryStrategy(ctx context.Context, host host.Host, config DiscoveryConfig, dhtInstance *dht.IpfsDHT) (*HybridDiscoveryStrategy, error) {
	hds := &HybridDiscoveryStrategy{
		host:               host,
		config:             config,
		dht:                dhtInstance,
		performanceMetrics: make(map[string]*StrategyPerformance),
		weights:            make(map[string]float64),
		lastUsed:           make(map[string]time.Time),
	}

	// Initialize individual strategies
	if err := hds.initializeStrategies(); err != nil {
		return nil, fmt.Errorf("failed to initialize strategies: %w", err)
	}

	// Set initial weights (can be adjusted dynamically based on performance)
	hds.weights["dht"] = 0.4
	hds.weights["mdns"] = 0.2
	hds.weights["bootstrap"] = 0.2
	hds.weights["rendezvous"] = 0.2

	return hds, nil
}

// initializeStrategies initializes all individual discovery strategies
func (hds *HybridDiscoveryStrategy) initializeStrategies() error {
	// Initialize DHT strategy
	if hds.dht != nil {
		hds.dhtStrategy = &DHTStrategy{
			dht:     hds.dht,
			routing: routing.NewRoutingDiscovery(hds.dht),
		}
		hds.performanceMetrics["dht"] = &StrategyPerformance{}
	}

	// Initialize mDNS strategy
	notifee := &mdnsNotifee{
		peerFound: make(chan peer.AddrInfo, 100), // Buffered channel
	}
	mdnsService := mdns.NewMdnsService(hds.host, hds.config.GetRendezvousString(), notifee)
	hds.mdnsStrategy = &MDNSStrategy{
		service: mdnsService,
	}
	hds.performanceMetrics["mdns"] = &StrategyPerformance{}

	// Initialize bootstrap strategy
	bootstrapPeers, err := parseBootstrapPeers(hds.config)
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}
	hds.bootstrapStrategy = NewBootstrapDiscovery(hds.host, bootstrapPeers, 5, 20)
	hds.performanceMetrics["bootstrap"] = &StrategyPerformance{}

	// Initialize rendezvous strategy
	if hds.dht != nil {
		hds.rendezvousStrategy = NewRendezvousDiscovery(hds.host, hds.dht)
		hds.performanceMetrics["rendezvous"] = &StrategyPerformance{}
	}

	return nil
}

// Name returns the strategy name
func (hds *HybridDiscoveryStrategy) Name() string {
	return "hybrid"
}

// FindPeers finds peers using a weighted combination of all strategies
func (hds *HybridDiscoveryStrategy) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo, 200) // Larger buffer for hybrid approach

	go func() {
		defer close(peerChan)

		var wg sync.WaitGroup

		// Get current weights
		hds.weightsMutex.RLock()
		weights := make(map[string]float64)
		for k, v := range hds.weights {
			weights[k] = v
		}
		hds.weightsMutex.RUnlock()

		// Calculate how many peers to get from each strategy based on weights
		totalWeight := 0.0
		for _, w := range weights {
			totalWeight += w
		}

		if totalWeight <= 0 {
			totalWeight = 1.0 // Fallback to equal distribution
		}

		// Limit to 50 peers total
		limit := discovery.Limit(50)
		opts = append(opts, limit)

		// Use strategies based on weights
		if hds.dhtStrategy != nil && weights["dht"] > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				hds.executeStrategy(ctx, "dht", hds.dhtStrategy, ns, peerChan, opts...)
			}()
		}

		if hds.mdnsStrategy != nil && weights["mdns"] > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				hds.executeStrategy(ctx, "mdns", hds.mdnsStrategy, ns, peerChan, opts...)
			}()
		}

		if hds.bootstrapStrategy != nil && weights["bootstrap"] > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				hds.executeStrategy(ctx, "bootstrap", hds.bootstrapStrategy, ns, peerChan, opts...)
			}()
		}

		if hds.rendezvousStrategy != nil && weights["rendezvous"] > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				hds.executeStrategy(ctx, "rendezvous", hds.rendezvousStrategy, ns, peerChan, opts...)
			}()
		}

		// Wait for all strategies to complete
		wg.Wait()
	}()

	return peerChan, nil
}

// executeStrategy executes a single discovery strategy and tracks performance
func (hds *HybridDiscoveryStrategy) executeStrategy(ctx context.Context, strategyName string, strategy DiscoveryStrategy, ns string, peerChan chan<- peer.AddrInfo, opts ...discovery.Option) {
	start := time.Now()

	// Update last used timestamp
	hds.lastUsedMutex.Lock()
	hds.lastUsed[strategyName] = time.Now()
	hds.lastUsedMutex.Unlock()

	// Execute the strategy
	peerStream, err := strategy.FindPeers(ctx, ns, opts...)
	if err != nil {
		log.Printf("Hybrid discovery: Strategy %s failed: %v", strategyName, err)
		hds.updatePerformanceMetrics(strategyName, false, time.Since(start))
		return
	}

	peersFound := 0
	for peer := range peerStream {
		// Skip ourselves
		if peer.ID == hds.host.ID() {
			continue
		}

		// Add a random delay to simulate varying network conditions for testing
		// In production, this would be replaced with actual performance metrics
		delay := time.Duration(rand.Intn(100)) * time.Millisecond

		select {
		case peerChan <- peer:
			peersFound++
		case <-ctx.Done():
			hds.updatePerformanceMetrics(strategyName, true, time.Since(start))
			return
		case <-time.After(delay):
			// Simulate processing delay
		}
	}

	hds.updatePerformanceMetrics(strategyName, true, time.Since(start))
	log.Printf("Hybrid discovery: Found %d peers using %s strategy", peersFound, strategyName)
}

// updatePerformanceMetrics updates performance metrics for a strategy
func (hds *HybridDiscoveryStrategy) updatePerformanceMetrics(strategyName string, success bool, latency time.Duration) {
	hds.metricsMutex.Lock()
	defer hds.metricsMutex.Unlock()

	metrics, exists := hds.performanceMetrics[strategyName]
	if !exists {
		metrics = &StrategyPerformance{}
		hds.performanceMetrics[strategyName] = metrics
	}

	if success {
		metrics.SuccessCount++
		metrics.LastSuccess = time.Now()

		// Update average latency (exponential moving average)
		if metrics.AverageLatency == 0 {
			metrics.AverageLatency = latency
		} else {
			// Simple moving average
			totalLatency := int64(metrics.AverageLatency)*int64(metrics.SuccessCount-1) + int64(latency)
			metrics.AverageLatency = time.Duration(totalLatency / int64(metrics.SuccessCount))
		}
	} else {
		metrics.FailureCount++
		metrics.LastFailure = time.Now()
	}
}

// Advertise advertises our presence using all strategies
func (hds *HybridDiscoveryStrategy) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	var minTTL time.Duration = time.Hour // Default fallback

	var lastErr error
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Advertise using all available strategies
	strategies := []struct {
		name     string
		strategy DiscoveryStrategy
	}{
		{"dht", hds.dhtStrategy},
		{"mdns", hds.mdnsStrategy},
		{"bootstrap", hds.bootstrapStrategy},
		{"rendezvous", hds.rendezvousStrategy},
	}

	for _, s := range strategies {
		if s.strategy != nil {
			wg.Add(1)
			go func(name string, strategy DiscoveryStrategy) {
				defer wg.Done()

				start := time.Now()
				ttl, err := strategy.Advertise(ctx, ns, opts...)
				latency := time.Since(start)

				hds.updatePerformanceMetrics(name, err == nil, latency)

				if err != nil {
					mu.Lock()
					lastErr = err
					mu.Unlock()
					return
				}

				mu.Lock()
				if minTTL == 0 || ttl < minTTL {
					minTTL = ttl
				}
				mu.Unlock()
			}(s.name, s.strategy)
		}
	}

	wg.Wait()

	if minTTL == 0 {
		return 5 * time.Minute, lastErr
	}

	return minTTL, nil
}

// AdjustWeights dynamically adjusts strategy weights based on performance
func (hds *HybridDiscoveryStrategy) AdjustWeights() {
	hds.metricsMutex.RLock()
	defer hds.metricsMutex.RUnlock()

	hds.weightsMutex.Lock()
	defer hds.weightsMutex.Unlock()

	totalSuccess := 0
	performance := make(map[string]float64)

	// Calculate total successes and individual performance scores
	for name, metrics := range hds.performanceMetrics {
		totalSuccess += metrics.SuccessCount
		if metrics.SuccessCount+metrics.FailureCount > 0 {
			successRate := float64(metrics.SuccessCount) / float64(metrics.SuccessCount+metrics.FailureCount)
			latencyScore := 1.0
			if metrics.AverageLatency > 0 {
				// Lower latency is better, so invert the score
				latencyScore = 1.0 / (1.0 + float64(metrics.AverageLatency)/float64(time.Second))
			}
			// Combine success rate and latency score
			performance[name] = successRate * latencyScore
		} else {
			performance[name] = 0.0
		}
	}

	// Normalize weights
	totalPerformance := 0.0
	for _, score := range performance {
		totalPerformance += score
	}

	if totalPerformance > 0 {
		for name, score := range performance {
			hds.weights[name] = score / totalPerformance
		}
	} else {
		// Reset to equal weights if no performance data
		for name := range hds.weights {
			hds.weights[name] = 1.0 / float64(len(hds.weights))
		}
	}
}

// GetWeights returns current strategy weights
func (hds *HybridDiscoveryStrategy) GetWeights() map[string]float64 {
	hds.weightsMutex.RLock()
	defer hds.weightsMutex.RUnlock()

	weights := make(map[string]float64)
	for k, v := range hds.weights {
		weights[k] = v
	}
	return weights
}

// GetPerformanceMetrics returns performance metrics for all strategies
func (hds *HybridDiscoveryStrategy) GetPerformanceMetrics() map[string]*StrategyPerformance {
	hds.metricsMutex.RLock()
	defer hds.metricsMutex.RUnlock()

	metrics := make(map[string]*StrategyPerformance)
	for k, v := range hds.performanceMetrics {
		metrics[k] = v
	}
	return metrics
}

// GetLastUsed returns last used timestamps for all strategies
func (hds *HybridDiscoveryStrategy) GetLastUsed() map[string]time.Time {
	hds.lastUsedMutex.RLock()
	defer hds.lastUsedMutex.RUnlock()

	lastUsed := make(map[string]time.Time)
	for k, v := range hds.lastUsed {
		lastUsed[k] = v
	}
	return lastUsed
}
