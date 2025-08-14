package discovery

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
)

// DiscoveryConfig represents configuration needed by the discovery engine
type DiscoveryConfig interface {
	GetBootstrapPeers() []string
	GetRendezvousString() string
	IsAutoDiscoveryEnabled() bool
}

// DiscoveryEngine manages multi-strategy peer discovery
type DiscoveryEngine struct {
	host       host.Host
	config     DiscoveryConfig
	dht        *dht.IpfsDHT
	mdns       mdns.Service
	bootstrap  *BootstrapDiscovery
	rendezvous *RendezvousDiscovery

	// Discovery strategies
	strategies []DiscoveryStrategy

	// Hybrid strategy (combines all strategies)
	hybridStrategy *HybridDiscoveryStrategy

	// Peer cache
	peerCache *PeerCache

	// Geographic detection
	geoDetector *GeographicDetector

	// Health checking
	healthChecker *HealthChecker

	// Events
	peerFound chan peer.AddrInfo
	peerLost  chan peer.ID

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	metrics *DiscoveryMetrics
}

// DiscoveryStrategy defines interface for discovery strategies
type DiscoveryStrategy interface {
	Name() string
	FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error)
	Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error)
}

// DiscoveryMetrics tracks discovery performance
type DiscoveryMetrics struct {
	PeersFound      int
	PeersLost       int
	ActivePeers     int
	DiscoveryErrors int
	LastDiscovery   time.Time
	StartTime       time.Time

	// Strategy metrics
	StrategyMetrics map[string]*StrategyMetrics
}

// StrategyMetrics tracks metrics for individual strategies
type StrategyMetrics struct {
	PeersFound     int
	Errors         int
	LastSuccess    time.Time
	AverageLatency time.Duration
}

// PeerCache manages discovered peers with LRU eviction and performance scoring
type PeerCache struct {
	peers    map[peer.ID]*CachedPeer
	peersMux sync.RWMutex

	// Cache settings
	maxSize int
	ttl     time.Duration

	// LRU tracking
	accessOrder []peer.ID
	orderMap    map[peer.ID]int

	// Performance scoring
	performanceScores map[peer.ID]float64

	// Cleanup
	cleanupInterval time.Duration
}

// CachedPeer represents a cached peer
type CachedPeer struct {
	AddrInfo     peer.AddrInfo
	DiscoveredAt time.Time
	LastSeen     time.Time
	Source       string
	Quality      *PeerQuality
}

// PeerQuality represents peer quality metrics
type PeerQuality struct {
	Latency     time.Duration
	Reliability float64
	Bandwidth   int64
	LastTest    time.Time
}

// NewDiscoveryEngine creates a new discovery engine
func NewDiscoveryEngine(ctx context.Context, h host.Host, config DiscoveryConfig) (*DiscoveryEngine, error) {
	ctx, cancel := context.WithCancel(ctx)

	engine := &DiscoveryEngine{
		host:      h,
		config:    config,
		peerFound: make(chan peer.AddrInfo, 100),
		peerLost:  make(chan peer.ID, 100),
		ctx:       ctx,
		cancel:    cancel,
		metrics: &DiscoveryMetrics{
			StartTime:       time.Now(),
			StrategyMetrics: make(map[string]*StrategyMetrics),
		},
	}

	// Initialize peer cache
	engine.peerCache = NewPeerCache(1000, 5*time.Minute)

	// Initialize geographic detector
	engine.geoDetector = NewGeographicDetector()

	// Initialize health checker
	engine.healthChecker = NewHealthChecker(h)

	// Initialize discovery strategies
	if err := engine.initializeStrategies(); err != nil {
		return nil, fmt.Errorf("failed to initialize discovery strategies: %w", err)
	}

	return engine, nil
}

// initializeStrategies initializes all discovery strategies
func (d *DiscoveryEngine) initializeStrategies() error {
	// Initialize DHT if enabled
	if isDHTEnabled(d.config) {
		if err := d.initializeDHT(); err != nil {
			return fmt.Errorf("failed to initialize DHT: %w", err)
		}
	}

	// Initialize mDNS
	if err := d.initializeMDNS(); err != nil {
		log.Printf("Failed to initialize mDNS: %v", err)
	}

	// Initialize bootstrap discovery
	if err := d.initializeBootstrap(); err != nil {
		return fmt.Errorf("failed to initialize bootstrap: %w", err)
	}

	// Initialize rendezvous discovery
	if err := d.initializeRendezvous(); err != nil {
		log.Printf("Failed to initialize rendezvous: %v", err)
	}

	// Initialize hybrid discovery strategy
	if err := d.initializeHybridStrategy(); err != nil {
		log.Printf("Failed to initialize hybrid strategy: %v", err)
	}

	return nil
}

// initializeDHT initializes DHT discovery
func (d *DiscoveryEngine) initializeDHT() error {
	// Configure DHT mode
	var mode dht.ModeOpt
	switch getDHTMode(d.config) {
	case "client":
		mode = dht.ModeClient
	case "server":
		mode = dht.ModeServer
	default:
		mode = dht.ModeAuto
	}

	// Create DHT
	kadDHT, err := dht.New(d.ctx, d.host, dht.Mode(mode))
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	d.dht = kadDHT

	// Bootstrap DHT
	if err := d.bootstrapDHT(); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Add DHT strategy
	dhtStrategy := &DHTStrategy{
		dht:     kadDHT,
		routing: routing.NewRoutingDiscovery(kadDHT),
	}
	d.strategies = append(d.strategies, dhtStrategy)
	d.metrics.StrategyMetrics["dht"] = &StrategyMetrics{}

	log.Printf("DHT initialized in %s mode", getDHTMode(d.config))
	return nil
}

// bootstrapDHT bootstraps the DHT with configured peers
func (d *DiscoveryEngine) bootstrapDHT() error {
	bootstrapPeers, err := parseBootstrapPeers(d.config)
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}

	if len(bootstrapPeers) == 0 {
		log.Printf("No bootstrap peers configured")
		return nil
	}

	// Connect to bootstrap peers
	for _, peerInfo := range bootstrapPeers {
		go func(p peer.AddrInfo) {
			ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
			defer cancel()

			if err := d.host.Connect(ctx, p); err != nil {
				log.Printf("Failed to connect to bootstrap peer %s: %v", p.ID, err)
				return
			}

			log.Printf("Connected to bootstrap peer: %s", p.ID)
		}(peerInfo)
	}

	// Bootstrap the DHT
	if err := d.dht.Bootstrap(d.ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	log.Printf("DHT bootstrapped with %d peers", len(bootstrapPeers))
	return nil
}

// initializeMDNS initializes mDNS discovery
func (d *DiscoveryEngine) initializeMDNS() error {
	notifee := &mdnsNotifee{
		peerFound: d.peerFound,
	}

	mdnsService := mdns.NewMdnsService(d.host, d.config.GetRendezvousString(), notifee)
	d.mdns = mdnsService

	// Add mDNS strategy
	mdnsStrategy := &MDNSStrategy{
		service: mdnsService,
	}
	d.strategies = append(d.strategies, mdnsStrategy)
	d.metrics.StrategyMetrics["mdns"] = &StrategyMetrics{}

	log.Printf("mDNS discovery initialized")
	return nil
}

// initializeBootstrap initializes bootstrap discovery
func (d *DiscoveryEngine) initializeBootstrap() error {
	bootstrapPeers, err := parseBootstrapPeers(d.config)
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}

	d.bootstrap = NewBootstrapDiscovery(d.host, bootstrapPeers, 5, 20)

	// Add bootstrap strategy
	d.strategies = append(d.strategies, d.bootstrap)
	d.metrics.StrategyMetrics["bootstrap"] = &StrategyMetrics{}

	log.Printf("Bootstrap discovery initialized with %d peers", len(bootstrapPeers))
	return nil
}

// initializeRendezvous initializes rendezvous discovery
func (d *DiscoveryEngine) initializeRendezvous() error {
	if d.dht == nil {
		return fmt.Errorf("rendezvous requires DHT")
	}

	d.rendezvous = NewRendezvousDiscovery(d.host, d.dht)

	// Add rendezvous strategy
	d.strategies = append(d.strategies, d.rendezvous)
	d.metrics.StrategyMetrics["rendezvous"] = &StrategyMetrics{}

	log.Printf("Rendezvous discovery initialized")
	return nil
}

// initializeHybridStrategy initializes the hybrid discovery strategy
func (d *DiscoveryEngine) initializeHybridStrategy() error {
	hybrid, err := NewHybridDiscoveryStrategy(d.ctx, d.host, d.config, d.dht)
	if err != nil {
		return fmt.Errorf("failed to create hybrid discovery strategy: %w", err)
	}

	d.hybridStrategy = hybrid

	// Add hybrid strategy
	d.strategies = append(d.strategies, d.hybridStrategy)
	d.metrics.StrategyMetrics["hybrid"] = &StrategyMetrics{}

	log.Printf("Hybrid discovery strategy initialized")
	return nil
}

// Start starts the discovery engine
func (d *DiscoveryEngine) Start() {
	log.Printf("Starting discovery engine with %d strategies", len(d.strategies))

	// Start peer cache cleanup
	d.wg.Add(1)
	go d.peerCache.start(d.ctx, &d.wg)

	// Start hybrid discovery strategy (it combines all others)
	if d.hybridStrategy != nil {
		d.wg.Add(1)
		go d.runHybridStrategy()
	} else {
		// Fallback to individual strategies if hybrid is not available
		for _, strategy := range d.strategies {
			d.wg.Add(1)
			go d.runStrategy(strategy)
		}
	}

	// Start bootstrap discovery
	if d.bootstrap != nil {
		d.wg.Add(1)
		go d.bootstrap.Start(d.ctx, &d.wg)
	}

	// Start health checker
	d.healthChecker.Start()

	// Start geographic detection for local node
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		if err := d.geoDetector.DetectLocalLocation(d.ctx); err != nil {
			log.Printf("Failed to detect local geographic location: %v", err)
		} else {
			log.Printf("Local geographic location detected")
		}
	}()

	// Start metrics collection
	d.wg.Add(1)
	go d.collectMetrics()

	// Start event processing
	d.wg.Add(1)
	go d.processEvents()

	// Start weight adjustment for hybrid strategy
	if d.hybridStrategy != nil {
		d.wg.Add(1)
		go d.adjustHybridWeights()
	}

	log.Printf("Discovery engine started")
}

// runStrategy runs a discovery strategy
func (d *DiscoveryEngine) runStrategy(strategy DiscoveryStrategy) {
	defer d.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.runDiscovery(strategy)
		}
	}
}

// runHybridStrategy runs the hybrid discovery strategy
func (d *DiscoveryEngine) runHybridStrategy() {
	defer d.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.runDiscovery(d.hybridStrategy)
		}
	}
}

// adjustHybridWeights periodically adjusts the weights of the hybrid strategy based on performance
func (d *DiscoveryEngine) adjustHybridWeights() {
	defer d.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if d.hybridStrategy != nil {
				d.hybridStrategy.AdjustWeights()
				log.Printf("Adjusted hybrid discovery strategy weights: %v", d.hybridStrategy.GetWeights())
			}
		}
	}
}

// runDiscovery runs discovery for a strategy
func (d *DiscoveryEngine) runDiscovery(strategy DiscoveryStrategy) {
	start := time.Now()

	// Advertise our presence
	if _, err := strategy.Advertise(d.ctx, d.config.GetRendezvousString()); err != nil {
		log.Printf("Failed to advertise on %s: %v", strategy.Name(), err)
		d.metrics.StrategyMetrics[strategy.Name()].Errors++
		return
	}

	// Find peers
	peerChan, err := strategy.FindPeers(d.ctx, d.config.GetRendezvousString(), discovery.Limit(50))
	if err != nil {
		log.Printf("Failed to find peers on %s: %v", strategy.Name(), err)
		d.metrics.StrategyMetrics[strategy.Name()].Errors++
		return
	}

	// Process found peers
	peersFound := 0
	for peer := range peerChan {
		// Skip ourselves
		if peer.ID == d.host.ID() {
			continue
		}

		// Add to cache
		d.peerCache.Add(peer, strategy.Name())

		// Send to event channel
		select {
		case d.peerFound <- peer:
			peersFound++
		case <-d.ctx.Done():
			return
		}
	}

	// Update metrics
	metrics := d.metrics.StrategyMetrics[strategy.Name()]
	metrics.PeersFound += peersFound
	metrics.LastSuccess = time.Now()
	metrics.AverageLatency = time.Since(start)

	if peersFound > 0 {
		log.Printf("Found %d peers using %s strategy", peersFound, strategy.Name())
	}
}

// processEvents processes discovery events
func (d *DiscoveryEngine) processEvents() {
	defer d.wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			return
		case peer := <-d.peerFound:
			d.handlePeerFound(peer)
		case peerID := <-d.peerLost:
			d.handlePeerLost(peerID)
		}
	}
}

// handlePeerFound handles peer discovery events
func (d *DiscoveryEngine) handlePeerFound(peer peer.AddrInfo) {
	// Add to health monitoring
	d.healthChecker.AddPeer(peer.ID)

	// Detect geographic location
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := d.geoDetector.DetectPeerLocation(ctx, peer.ID, peer.Addrs); err != nil {
			log.Printf("Failed to detect location for peer %s: %v", peer.ID, err)
		}
	}()

	// Connect to peer if not already connected
	if d.host.Network().Connectedness(peer.ID) != network.Connected {
		go func() {
			ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
			defer cancel()

			start := time.Now()
			if err := d.host.Connect(ctx, peer); err != nil {
				log.Printf("Failed to connect to discovered peer %s: %v", peer.ID, err)
				// Update performance score - decrease for failed connection
				if d.peerCache != nil {
					currentScore, _ := d.peerCache.GetPerformanceScore(peer.ID)
					newScore := currentScore * 0.8 // Reduce score by 20%
					if newScore < 0.1 {
						newScore = 0.1
					}
					d.peerCache.UpdatePerformanceScore(peer.ID, newScore)
				}
				return
			}

			// Update performance score - increase for successful connection
			if d.peerCache != nil {
				currentScore, _ := d.peerCache.GetPerformanceScore(peer.ID)
				if currentScore == 0 {
					currentScore = 1.0 // Default for new peers
				}
				connectionTime := time.Since(start)
				// Faster connections get higher scores
				timeFactor := 1.0
				if connectionTime < 100*time.Millisecond {
					timeFactor = 1.2
				} else if connectionTime < 500*time.Millisecond {
					timeFactor = 1.1
				} else if connectionTime > 2*time.Second {
					timeFactor = 0.9
				}
				newScore := currentScore * timeFactor
				if newScore > 10.0 {
					newScore = 10.0
				}
				d.peerCache.UpdatePerformanceScore(peer.ID, newScore)
			}

			log.Printf("Connected to discovered peer: %s", peer.ID)
		}()
	}

	d.metrics.PeersFound++
	d.metrics.LastDiscovery = time.Now()
}

// handlePeerLost handles peer loss events
func (d *DiscoveryEngine) handlePeerLost(peerID peer.ID) {
	d.peerCache.Remove(peerID)
	d.healthChecker.RemovePeer(peerID)
	d.metrics.PeersLost++
	log.Printf("Peer lost: %s", peerID)
}

// collectMetrics collects discovery metrics
func (d *DiscoveryEngine) collectMetrics() {
	defer d.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.updateMetrics()
		}
	}
}

// updateMetrics updates discovery metrics
func (d *DiscoveryEngine) updateMetrics() {
	d.metrics.ActivePeers = len(d.host.Network().Peers())
}

// GetMetrics returns discovery metrics
func (d *DiscoveryEngine) GetMetrics() *DiscoveryMetrics {
	return d.metrics
}

// GetDHT returns the DHT instance
func (d *DiscoveryEngine) GetDHT() *dht.IpfsDHT {
	return d.dht
}

// GetGeographicDetector returns the geographic detector
func (d *DiscoveryEngine) GetGeographicDetector() *GeographicDetector {
	return d.geoDetector
}

// GetHealthChecker returns the health checker
func (d *DiscoveryEngine) GetHealthChecker() *HealthChecker {
	return d.healthChecker
}

// GetNearbyPeers returns peers within a geographic distance
func (d *DiscoveryEngine) GetNearbyPeers(maxDistanceKm float64) []peer.ID {
	return d.geoDetector.GetNearbyPeers(maxDistanceKm)
}

// GetHealthyPeers returns a list of healthy peers
func (d *DiscoveryEngine) GetHealthyPeers() []peer.ID {
	return d.healthChecker.GetHealthyPeers()
}

// GetPeerLocation returns geographic information for a peer
func (d *DiscoveryEngine) GetPeerLocation(peerID peer.ID) *GeographicInfo {
	return d.geoDetector.GetPeerLocation(peerID)
}

// GetPeerHealth returns health information for a peer
func (d *DiscoveryEngine) GetPeerHealth(peerID peer.ID) *PeerHealth {
	return d.healthChecker.GetPeerHealth(peerID)
}

// Stop stops the discovery engine
func (d *DiscoveryEngine) Stop() {
	log.Printf("Stopping discovery engine")
	d.cancel()

	// Stop health checker
	if d.healthChecker != nil {
		d.healthChecker.Stop()
	}

	d.wg.Wait()

	if d.dht != nil {
		d.dht.Close()
	}

	log.Printf("Discovery engine stopped")
}

// mdnsNotifee handles mDNS notifications
type mdnsNotifee struct {
	peerFound chan peer.AddrInfo
}

func (n *mdnsNotifee) HandlePeerFound(peer peer.AddrInfo) {
	select {
	case n.peerFound <- peer:
	default:
		// Channel full, drop event
	}
}

// NewPeerCache creates a new peer cache with LRU eviction and performance scoring
func NewPeerCache(maxSize int, ttl time.Duration) *PeerCache {
	return &PeerCache{
		peers:             make(map[peer.ID]*CachedPeer),
		maxSize:           maxSize,
		ttl:               ttl,
		accessOrder:       make([]peer.ID, 0, maxSize),
		orderMap:          make(map[peer.ID]int),
		performanceScores: make(map[peer.ID]float64),
		cleanupInterval:   ttl / 2,
	}
}

// Add adds a peer to the cache with LRU tracking
func (c *PeerCache) Add(peer peer.AddrInfo, source string) {
	c.peersMux.Lock()
	defer c.peersMux.Unlock()

	now := time.Now()
	cachedPeer := &CachedPeer{
		AddrInfo:     peer,
		DiscoveredAt: now,
		LastSeen:     now,
		Source:       source,
	}

	// Update existing peer or add new one
	if existingPeer, exists := c.peers[peer.ID]; exists {
		// Update existing peer info
		existingPeer.LastSeen = now
		existingPeer.AddrInfo = peer
		// Move to end of access order (most recently used)
		c.moveToEnd(peer.ID)
	} else {
		// Add new peer
		c.peers[peer.ID] = cachedPeer
		c.accessOrder = append(c.accessOrder, peer.ID)
		c.orderMap[peer.ID] = len(c.accessOrder) - 1

		// Initialize performance score
		c.performanceScores[peer.ID] = 1.0

		// Cleanup if cache is full
		if len(c.peers) > c.maxSize {
			c.lruCleanup()
		}
	}
}

// Remove removes a peer from the cache
func (c *PeerCache) Remove(peerID peer.ID) {
	c.peersMux.Lock()
	defer c.peersMux.Unlock()

	delete(c.peers, peerID)
}

// Get retrieves a peer from the cache and updates access order
func (c *PeerCache) Get(peerID peer.ID) (*CachedPeer, bool) {
	c.peersMux.Lock()
	defer c.peersMux.Unlock()

	peer, exists := c.peers[peerID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(peer.LastSeen) > c.ttl {
		// Remove expired peer
		delete(c.peers, peerID)
		c.removeFromOrderMap(peerID)
		delete(c.performanceScores, peerID)
		return nil, false
	}

	// Update access order (move to end as most recently used)
	c.moveToEnd(peerID)

	// Update last seen time
	peer.LastSeen = time.Now()

	return peer, true
}

// cleanup removes expired peers from cache
func (c *PeerCache) cleanup() {
	now := time.Now()

	for id, peer := range c.peers {
		if now.Sub(peer.LastSeen) > c.ttl {
			delete(c.peers, id)
			c.removeFromOrderMap(id)
			delete(c.performanceScores, id)
		}
	}
}

// lruCleanup removes the least recently used peers when cache is full
func (c *PeerCache) lruCleanup() {
	// Remove expired peers first
	c.cleanup()

	// If still over capacity, remove LRU peers
	for len(c.peers) > c.maxSize && len(c.accessOrder) > 0 {
		// Get least recently used peer (first in accessOrder)
		lruPeerID := c.accessOrder[0]

		// Remove from all structures
		delete(c.peers, lruPeerID)
		c.removeFromOrderMap(lruPeerID)
		delete(c.performanceScores, lruPeerID)
	}
}

// removeFromOrderMap removes a peer from the order tracking map
func (c *PeerCache) removeFromOrderMap(peerID peer.ID) {
	if idx, exists := c.orderMap[peerID]; exists {
		// Remove from accessOrder slice
		copy(c.accessOrder[idx:], c.accessOrder[idx+1:])
		c.accessOrder[len(c.accessOrder)-1] = ""
		c.accessOrder = c.accessOrder[:len(c.accessOrder)-1]

		// Update indices of remaining peers
		for i := idx; i < len(c.accessOrder); i++ {
			c.orderMap[c.accessOrder[i]] = i
		}

		// Remove from map
		delete(c.orderMap, peerID)
	}
}

// moveToEnd moves a peer to the end of the access order (most recently used)
func (c *PeerCache) moveToEnd(peerID peer.ID) {
	if idx, exists := c.orderMap[peerID]; exists {
		// Remove from current position
		copy(c.accessOrder[idx:], c.accessOrder[idx+1:])
		c.accessOrder[len(c.accessOrder)-1] = peerID
		c.accessOrder = c.accessOrder[:len(c.accessOrder)-1]

		// Add to end
		c.accessOrder = append(c.accessOrder, peerID)

		// Update indices
		for i := idx; i < len(c.accessOrder); i++ {
			c.orderMap[c.accessOrder[i]] = i
		}
	}
}

// UpdatePerformanceScore updates the performance score for a peer
func (c *PeerCache) UpdatePerformanceScore(peerID peer.ID, score float64) {
	c.peersMux.Lock()
	defer c.peersMux.Unlock()

	c.performanceScores[peerID] = score
}

// GetPerformanceScore returns the performance score for a peer
func (c *PeerCache) GetPerformanceScore(peerID peer.ID) (float64, bool) {
	c.peersMux.RLock()
	defer c.peersMux.RUnlock()

	score, exists := c.performanceScores[peerID]
	return score, exists
}

// GetTopPerformers returns the top N performing peers
func (c *PeerCache) GetTopPerformers(n int) []peer.ID {
	c.peersMux.RLock()
	defer c.peersMux.RUnlock()

	// Create a slice of peer IDs and scores
	type peerScore struct {
		id    peer.ID
		score float64
	}

	scores := make([]peerScore, 0, len(c.performanceScores))
	for id, score := range c.performanceScores {
		// Only include peers that still exist in the cache
		if _, exists := c.peers[id]; exists {
			scores = append(scores, peerScore{id: id, score: score})
		}
	}

	// Sort by score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Return top N
	result := make([]peer.ID, 0, n)
	for i := 0; i < len(scores) && i < n; i++ {
		result = append(result, scores[i].id)
	}

	return result
}

// start starts the peer cache cleanup routine
func (c *PeerCache) start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.peersMux.Lock()
			c.cleanup()
			c.peersMux.Unlock()
		}
	}
}

// Helper functions to work with DiscoveryConfig interface

// isDHTEnabled checks if DHT is enabled in the configuration
func isDHTEnabled(config DiscoveryConfig) bool {
	// Use reflection to check for EnableDHT field
	// For now, default to true since DHT is generally enabled
	return true
}

// getDHTMode gets the DHT mode from the configuration
func getDHTMode(config DiscoveryConfig) string {
	// Use reflection to check for DHTMode field
	// For now, default to auto mode
	return "auto"
}

// parseBootstrapPeers parses bootstrap peer addresses from the configuration
func parseBootstrapPeers(config DiscoveryConfig) ([]peer.AddrInfo, error) {
	bootstrapAddrs := config.GetBootstrapPeers()
	var peers []peer.AddrInfo

	for _, addr := range bootstrapAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("Invalid bootstrap address %s: %v", addr, err)
			continue
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Printf("Failed to parse peer info from %s: %v", addr, err)
			continue
		}

		peers = append(peers, *peerInfo)
	}

	return peers, nil
}
