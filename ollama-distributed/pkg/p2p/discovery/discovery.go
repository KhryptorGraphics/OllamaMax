package discovery

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	
	"github.com/ollama/ollama-distributed/pkg/config"
)

// DiscoveryConfig represents configuration needed by the discovery engine
type DiscoveryConfig interface {
	GetBootstrapPeers() []string
	GetRendezvousString() string
	IsAutoDiscoveryEnabled() bool
}

// DiscoveryEngine manages multi-strategy peer discovery
type DiscoveryEngine struct {
	host        host.Host
	config      DiscoveryConfig
	dht         *dht.IpfsDHT
	mdns        discovery.Discovery
	bootstrap   *BootstrapDiscovery
	rendezvous  *RendezvousDiscovery
	
	// Discovery strategies
	strategies  []DiscoveryStrategy
	
	// Peer cache
	peerCache   *PeerCache
	
	// Events
	peerFound   chan peer.AddrInfo
	peerLost    chan peer.ID
	
	// Lifecycle
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	
	// Metrics
	metrics     *DiscoveryMetrics
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
	PeersFound      int
	Errors          int
	LastSuccess     time.Time
	AverageLatency  time.Duration
}

// PeerCache manages discovered peers
type PeerCache struct {
	peers    map[peer.ID]*CachedPeer
	peersMux sync.RWMutex
	
	// Cache settings
	maxSize    int
	ttl        time.Duration
	
	// Cleanup
	cleanupInterval time.Duration
}

// CachedPeer represents a cached peer
type CachedPeer struct {
	AddrInfo    peer.AddrInfo
	DiscoveredAt time.Time
	LastSeen    time.Time
	Source      string
	Quality     *PeerQuality
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
		host:       h,
		config:     config,
		peerFound:  make(chan peer.AddrInfo, 100),
		peerLost:   make(chan peer.ID, 100),
		ctx:        ctx,
		cancel:     cancel,
		metrics: &DiscoveryMetrics{
			StartTime:       time.Now(),
			StrategyMetrics: make(map[string]*StrategyMetrics),
		},
	}
	
	// Initialize peer cache
	engine.peerCache = NewPeerCache(1000, 5*time.Minute)
	
	// Initialize discovery strategies
	if err := engine.initializeStrategies(); err != nil {
		return nil, fmt.Errorf("failed to initialize discovery strategies: %w", err)
	}
	
	return engine, nil
}

// initializeStrategies initializes all discovery strategies
func (d *DiscoveryEngine) initializeStrategies() error {
	// Initialize DHT if enabled
	if d.config.EnableDHT {
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
	
	return nil
}

// initializeDHT initializes DHT discovery
func (d *DiscoveryEngine) initializeDHT() error {
	// Configure DHT mode
	var mode dht.ModeOpt
	switch d.config.DHTMode {
	case "client":
		mode = dht.ModeClient
	case "server":
		mode = dht.ModeServer
	default:
		mode = dht.ModeAuto
	}
	
	// Create DHT
	kadDHT, err := dht.New(d.ctx, d.host, mode)
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
	
	log.Printf("DHT initialized in %s mode", d.config.DHTMode)
	return nil
}

// bootstrapDHT bootstraps the DHT with configured peers
func (d *DiscoveryEngine) bootstrapDHT() error {
	bootstrapPeers, err := d.config.ParseBootstrapPeers()
	if err != nil {
		return fmt.Errorf("failed to parse bootstrap peers: %w", err)
	}
	
	if len(bootstrapPeers) == 0 {
		log.Printf("No bootstrap peers configured")
		return nil
	}
	
	// Connect to bootstrap peers
	for _, peer := range bootstrapPeers {
		go func(p peer.AddrInfo) {
			ctx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
			defer cancel()
			
			if err := d.host.Connect(ctx, p); err != nil {
				log.Printf("Failed to connect to bootstrap peer %s: %v", p.ID, err)
				return
			}
			
			log.Printf("Connected to bootstrap peer: %s", p.ID)
		}(peer)
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
	mdnsService := mdns.NewMdnsService(d.host, "ollamacron", &mdnsNotifee{
		peerFound: d.peerFound,
	})
	
	if err := mdnsService.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS service: %w", err)
	}
	
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
	bootstrapPeers, err := d.config.ParseBootstrapPeers()
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

// Start starts the discovery engine
func (d *DiscoveryEngine) Start() {
	log.Printf("Starting discovery engine with %d strategies", len(d.strategies))
	
	// Start peer cache cleanup
	d.wg.Add(1)
	go d.peerCache.start(d.ctx, &d.wg)
	
	// Start discovery strategies
	for _, strategy := range d.strategies {
		d.wg.Add(1)
		go d.runStrategy(strategy)
	}
	
	// Start bootstrap discovery
	if d.bootstrap != nil {
		d.wg.Add(1)
		go d.bootstrap.Start(d.ctx, &d.wg)
	}
	
	// Start metrics collection
	d.wg.Add(1)
	go d.collectMetrics()
	
	// Start event processing
	d.wg.Add(1)
	go d.processEvents()
	
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

// runDiscovery runs discovery for a strategy
func (d *DiscoveryEngine) runDiscovery(strategy DiscoveryStrategy) {
	start := time.Now()
	
	// Advertise our presence
	if _, err := strategy.Advertise(d.ctx, "ollamacron"); err != nil {
		log.Printf("Failed to advertise on %s: %v", strategy.Name(), err)
		d.metrics.StrategyMetrics[strategy.Name()].Errors++
		return
	}
	
	// Find peers
	peerChan, err := strategy.FindPeers(d.ctx, "ollamacron", discovery.Limit(50))
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
	// Connect to peer if not already connected
	if d.host.Network().Connectedness(peer.ID) != network.Connected {
		go func() {
			ctx, cancel := context.WithTimeout(d.ctx, 10*time.Second)
			defer cancel()
			
			if err := d.host.Connect(ctx, peer); err != nil {
				log.Printf("Failed to connect to discovered peer %s: %v", peer.ID, err)
				return
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

// Stop stops the discovery engine
func (d *DiscoveryEngine) Stop() {
	log.Printf("Stopping discovery engine")
	d.cancel()
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

// NewPeerCache creates a new peer cache
func NewPeerCache(maxSize int, ttl time.Duration) *PeerCache {
	return &PeerCache{
		peers:           make(map[peer.ID]*CachedPeer),
		maxSize:         maxSize,
		ttl:             ttl,
		cleanupInterval: ttl / 2,
	}
}

// Add adds a peer to the cache
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
	
	c.peers[peer.ID] = cachedPeer
	
	// Cleanup if cache is full
	if len(c.peers) > c.maxSize {
		c.cleanup()
	}
}

// Remove removes a peer from the cache
func (c *PeerCache) Remove(peerID peer.ID) {
	c.peersMux.Lock()
	defer c.peersMux.Unlock()
	
	delete(c.peers, peerID)
}

// Get retrieves a peer from the cache
func (c *PeerCache) Get(peerID peer.ID) (*CachedPeer, bool) {
	c.peersMux.RLock()
	defer c.peersMux.RUnlock()
	
	peer, exists := c.peers[peerID]
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if time.Since(peer.LastSeen) > c.ttl {
		return nil, false
	}
	
	return peer, true
}

// cleanup removes expired peers from cache
func (c *PeerCache) cleanup() {
	now := time.Now()
	
	for id, peer := range c.peers {
		if now.Sub(peer.LastSeen) > c.ttl {
			delete(c.peers, id)
		}
	}
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