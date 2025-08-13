package discovery

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

// DHTStrategy implements DHT-based discovery
type DHTStrategy struct {
	dht     *dht.IpfsDHT
	routing *routing.RoutingDiscovery
}

// Name returns the strategy name
func (d *DHTStrategy) Name() string {
	return "dht"
}

// FindPeers finds peers using DHT
func (d *DHTStrategy) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	return d.routing.FindPeers(ctx, ns, opts...)
}

// Advertise advertises our presence in DHT
func (d *DHTStrategy) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	return d.routing.Advertise(ctx, ns, opts...)
}

// MDNSStrategy implements mDNS-based discovery
type MDNSStrategy struct {
	service mdns.Service
}

// Name returns the strategy name
func (m *MDNSStrategy) Name() string {
	return "mdns"
}

// FindPeers finds peers using mDNS
func (m *MDNSStrategy) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	// mDNS discovery is passive - peers are found via notifications
	peerChan := make(chan peer.AddrInfo)
	close(peerChan)
	return peerChan, nil
}

// Advertise advertises our presence via mDNS
func (m *MDNSStrategy) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	// mDNS advertising is handled by the service itself
	return 5 * time.Minute, nil
}

// BootstrapDiscovery implements bootstrap peer discovery
type BootstrapDiscovery struct {
	host           host.Host
	bootstrapPeers []peer.AddrInfo
	minPeers       int
	maxPeers       int

	// Connection tracking
	connections    map[peer.ID]*ConnectionInfo
	connectionsMux sync.RWMutex
}

// ConnectionInfo tracks connection information
type ConnectionInfo struct {
	ConnectedAt time.Time
	LastSeen    time.Time
	Attempts    int
	Failures    int
}

// NewBootstrapDiscovery creates a new bootstrap discovery strategy
func NewBootstrapDiscovery(host host.Host, bootstrapPeers []peer.AddrInfo, minPeers, maxPeers int) *BootstrapDiscovery {
	return &BootstrapDiscovery{
		host:           host,
		bootstrapPeers: bootstrapPeers,
		minPeers:       minPeers,
		maxPeers:       maxPeers,
		connections:    make(map[peer.ID]*ConnectionInfo),
	}
}

// Name returns the strategy name
func (b *BootstrapDiscovery) Name() string {
	return "bootstrap"
}

// FindPeers finds peers from bootstrap list
func (b *BootstrapDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo, len(b.bootstrapPeers))

	go func() {
		defer close(peerChan)

		for _, peer := range b.bootstrapPeers {
			select {
			case peerChan <- peer:
			case <-ctx.Done():
				return
			}
		}
	}()

	return peerChan, nil
}

// Advertise advertises to bootstrap peers
func (b *BootstrapDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	// Bootstrap peers don't need advertisement
	return 5 * time.Minute, nil
}

// Start starts the bootstrap discovery process
func (b *BootstrapDiscovery) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.ensureConnections(ctx)
		}
	}
}

// ensureConnections ensures minimum connection requirements
func (b *BootstrapDiscovery) ensureConnections(ctx context.Context) {
	connected := len(b.host.Network().Peers())

	if connected < b.minPeers {
		// Connect to more bootstrap peers
		for _, peer := range b.bootstrapPeers {
			if connected >= b.maxPeers {
				break
			}

			// Skip if already connected
			if b.host.Network().Connectedness(peer.ID) == network.Connected {
				continue
			}

			// Skip if too many recent failures
			b.connectionsMux.RLock()
			connInfo, exists := b.connections[peer.ID]
			b.connectionsMux.RUnlock()

			if exists && connInfo.Failures > 5 && time.Since(connInfo.LastSeen) < 5*time.Minute {
				continue
			}

			go b.connectToPeer(ctx, peer)
			connected++
		}
	}
}

// connectToPeer connects to a bootstrap peer
func (b *BootstrapDiscovery) connectToPeer(ctx context.Context, peer peer.AddrInfo) {
	b.connectionsMux.Lock()
	connInfo, exists := b.connections[peer.ID]
	if !exists {
		connInfo = &ConnectionInfo{}
		b.connections[peer.ID] = connInfo
	}
	connInfo.Attempts++
	b.connectionsMux.Unlock()

	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := b.host.Connect(connectCtx, peer); err != nil {
		log.Printf("Failed to connect to bootstrap peer %s: %v", peer.ID, err)

		b.connectionsMux.Lock()
		connInfo.Failures++
		connInfo.LastSeen = time.Now()
		b.connectionsMux.Unlock()
		return
	}

	log.Printf("Connected to bootstrap peer: %s", peer.ID)

	b.connectionsMux.Lock()
	connInfo.ConnectedAt = time.Now()
	connInfo.LastSeen = time.Now()
	b.connectionsMux.Unlock()
}

// RendezvousDiscovery implements rendezvous-based discovery
type RendezvousDiscovery struct {
	host    host.Host
	dht     *dht.IpfsDHT
	routing *routing.RoutingDiscovery

	// Rendezvous points
	rendezvous map[string]*RendezvousPoint
	rendMux    sync.RWMutex
}

// RendezvousPoint represents a rendezvous point
type RendezvousPoint struct {
	Namespace  string
	TTL        time.Duration
	LastUpdate time.Time
	PeerCount  int
}

// NewRendezvousDiscovery creates a new rendezvous discovery strategy
func NewRendezvousDiscovery(host host.Host, dht *dht.IpfsDHT) *RendezvousDiscovery {
	return &RendezvousDiscovery{
		host:       host,
		dht:        dht,
		routing:    routing.NewRoutingDiscovery(dht),
		rendezvous: make(map[string]*RendezvousPoint),
	}
}

// Name returns the strategy name
func (r *RendezvousDiscovery) Name() string {
	return "rendezvous"
}

// FindPeers finds peers at rendezvous points
func (r *RendezvousDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	// Create rendezvous namespaces
	rendezvousPoints := []string{
		fmt.Sprintf("%s/general", ns),
		fmt.Sprintf("%s/models", ns),
		fmt.Sprintf("%s/compute", ns),
	}

	peerChan := make(chan peer.AddrInfo, 100)

	go func() {
		defer close(peerChan)

		for _, rdv := range rendezvousPoints {
			peerStream, err := r.routing.FindPeers(ctx, rdv, opts...)
			if err != nil {
				log.Printf("Failed to find peers at rendezvous %s: %v", rdv, err)
				continue
			}

			for peer := range peerStream {
				select {
				case peerChan <- peer:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return peerChan, nil
}

// Advertise advertises at rendezvous points
func (r *RendezvousDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	// Advertise at multiple rendezvous points
	rendezvousPoints := []string{
		fmt.Sprintf("%s/general", ns),
		fmt.Sprintf("%s/models", ns),
		fmt.Sprintf("%s/compute", ns),
	}

	var minTTL time.Duration

	for _, rdv := range rendezvousPoints {
		ttl, err := r.routing.Advertise(ctx, rdv, opts...)
		if err != nil {
			log.Printf("Failed to advertise at rendezvous %s: %v", rdv, err)
			continue
		}

		// Track rendezvous point
		r.rendMux.Lock()
		r.rendezvous[rdv] = &RendezvousPoint{
			Namespace:  rdv,
			TTL:        ttl,
			LastUpdate: time.Now(),
		}
		r.rendMux.Unlock()

		if minTTL == 0 || ttl < minTTL {
			minTTL = ttl
		}

		log.Printf("Advertised at rendezvous point: %s (TTL: %v)", rdv, ttl)
	}

	if minTTL == 0 {
		return 5 * time.Minute, nil
	}

	return minTTL, nil
}

// GetRendezvousPoints returns active rendezvous points
func (r *RendezvousDiscovery) GetRendezvousPoints() map[string]*RendezvousPoint {
	r.rendMux.RLock()
	defer r.rendMux.RUnlock()

	points := make(map[string]*RendezvousPoint)
	for k, v := range r.rendezvous {
		points[k] = v
	}

	return points
}

// CustomDiscovery implements custom discovery strategies
type CustomDiscovery struct {
	host       host.Host
	name       string
	finder     func(context.Context, string, ...discovery.Option) (<-chan peer.AddrInfo, error)
	advertiser func(context.Context, string, ...discovery.Option) (time.Duration, error)
}

// NewCustomDiscovery creates a new custom discovery strategy
func NewCustomDiscovery(
	host host.Host,
	name string,
	finder func(context.Context, string, ...discovery.Option) (<-chan peer.AddrInfo, error),
	advertiser func(context.Context, string, ...discovery.Option) (time.Duration, error),
) *CustomDiscovery {
	return &CustomDiscovery{
		host:       host,
		name:       name,
		finder:     finder,
		advertiser: advertiser,
	}
}

// Name returns the strategy name
func (c *CustomDiscovery) Name() string {
	return c.name
}

// FindPeers finds peers using custom logic
func (c *CustomDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	if c.finder == nil {
		peerChan := make(chan peer.AddrInfo)
		close(peerChan)
		return peerChan, nil
	}

	return c.finder(ctx, ns, opts...)
}

// Advertise advertises using custom logic
func (c *CustomDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	if c.advertiser == nil {
		return 5 * time.Minute, nil
	}

	return c.advertiser(ctx, ns, opts...)
}

// HybridDiscovery combines multiple discovery strategies
type HybridDiscovery struct {
	strategies []DiscoveryStrategy
	weights    map[string]float64

	// Load balancing
	lastUsed   map[string]time.Time
	usageCount map[string]int
}

// NewHybridDiscovery creates a new hybrid discovery strategy
func NewHybridDiscovery(strategies []DiscoveryStrategy) *HybridDiscovery {
	weights := make(map[string]float64)
	lastUsed := make(map[string]time.Time)
	usageCount := make(map[string]int)

	// Initialize equal weights
	for _, strategy := range strategies {
		weights[strategy.Name()] = 1.0
		lastUsed[strategy.Name()] = time.Now()
		usageCount[strategy.Name()] = 0
	}

	return &HybridDiscovery{
		strategies: strategies,
		weights:    weights,
		lastUsed:   lastUsed,
		usageCount: usageCount,
	}
}

// Name returns the strategy name
func (h *HybridDiscovery) Name() string {
	return "hybrid"
}

// FindPeers finds peers using multiple strategies
func (h *HybridDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	peerChan := make(chan peer.AddrInfo, 100)

	go func() {
		defer close(peerChan)

		var wg sync.WaitGroup

		// Run all strategies in parallel
		for _, strategy := range h.strategies {
			wg.Add(1)
			go func(s DiscoveryStrategy) {
				defer wg.Done()

				strategyPeers, err := s.FindPeers(ctx, ns, opts...)
				if err != nil {
					log.Printf("Strategy %s failed: %v", s.Name(), err)
					return
				}

				for peer := range strategyPeers {
					select {
					case peerChan <- peer:
						h.usageCount[s.Name()]++
					case <-ctx.Done():
						return
					}
				}
			}(strategy)
		}

		wg.Wait()
	}()

	return peerChan, nil
}

// Advertise advertises using all strategies
func (h *HybridDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	var minTTL time.Duration
	var lastErr error

	for _, strategy := range h.strategies {
		ttl, err := strategy.Advertise(ctx, ns, opts...)
		if err != nil {
			lastErr = err
			continue
		}

		h.lastUsed[strategy.Name()] = time.Now()

		if minTTL == 0 || ttl < minTTL {
			minTTL = ttl
		}
	}

	if minTTL == 0 {
		return 5 * time.Minute, lastErr
	}

	return minTTL, nil
}

// UpdateWeights updates strategy weights based on performance
func (h *HybridDiscovery) UpdateWeights(performance map[string]float64) {
	for strategy, weight := range performance {
		if _, exists := h.weights[strategy]; exists {
			h.weights[strategy] = weight
		}
	}
}

// GetWeights returns current strategy weights
func (h *HybridDiscovery) GetWeights() map[string]float64 {
	weights := make(map[string]float64)
	for k, v := range h.weights {
		weights[k] = v
	}
	return weights
}

// GetUsageStats returns usage statistics
func (h *HybridDiscovery) GetUsageStats() map[string]interface{} {
	stats := make(map[string]interface{})

	for strategy := range h.weights {
		stats[strategy] = map[string]interface{}{
			"weight":      h.weights[strategy],
			"last_used":   h.lastUsed[strategy],
			"usage_count": h.usageCount[strategy],
		}
	}

	return stats
}
