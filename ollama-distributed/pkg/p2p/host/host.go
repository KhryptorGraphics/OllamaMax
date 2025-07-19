package host

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/multiformats/go-multiaddr"
	
	"github.com/ollama/ollama-distributed/pkg/config"
)

// P2PHost wraps libp2p host with enhanced functionality
type P2PHost struct {
	host.Host
	config         *config.NodeConfig
	capabilities   *config.NodeCapabilities
	relayService   *relay.Relay
	
	// Protocol handlers
	protocols      map[protocol.ID]network.StreamHandler
	
	// Event handlers
	connectHandler    func(network.Network, network.Conn)
	disconnectHandler func(network.Network, network.Conn)
	
	// Metrics
	metrics        *HostMetrics
	
	// Lifecycle
	ctx            context.Context
	cancel         context.CancelFunc
}

// HostMetrics tracks host performance metrics
type HostMetrics struct {
	ConnectionCount    int
	StreamCount        int
	BytesReceived      int64
	BytesSent          int64
	ProtocolHandlers   int
	LastActivity       time.Time
	StartTime          time.Time
}

// NewP2PHost creates a new enhanced P2P host
func NewP2PHost(ctx context.Context, config *config.NodeConfig) (*P2PHost, error) {
	// Load or generate private key
	priv, err := loadOrGenerateKey(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key: %w", err)
	}
	
	// Build listen addresses
	listenAddrs := make([]multiaddr.Multiaddr, 0, len(config.Listen))
	for _, addr := range config.Listen {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("Invalid listen address %s: %v", addr, err)
			continue
		}
		listenAddrs = append(listenAddrs, maddr)
	}
	
	// Configure transports
	transports := []libp2p.Option{
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.Transport(webtransport.New),
	}
	
	// Configure security
	security := []libp2p.Option{}
	if config.EnableNoise {
		security = append(security, libp2p.Security(noise.ID, noise.New))
	}
	if config.EnableTLS {
		security = append(security, libp2p.Security(tls.ID, tls.New))
	}
	
	// Configure NAT traversal
	natOptions := []libp2p.Option{}
	if config.EnableNATService {
		natOptions = append(natOptions, libp2p.EnableNATService())
	}
	if config.EnableHolePunching {
		natOptions = append(natOptions, libp2p.EnableHolePunching())
	}
	if config.EnableAutoRelay {
		staticRelays, err := config.ParseStaticRelays()
		if err != nil {
			return nil, fmt.Errorf("failed to parse static relays: %w", err)
		}
		natOptions = append(natOptions, libp2p.EnableAutoRelayWithStaticRelays(staticRelays))
	}
	
	// Configure connection manager
	connMgr, err := connmgr.NewConnManager(
		config.ConnMgrLow,
		config.ConnMgrHigh,
		connmgr.WithGracePeriod(config.ConnMgrGrace),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}
	
	// Build host options
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.ConnectionManager(connMgr),
		libp2p.EnableRelay(),
	}
	
	// Add transport options
	opts = append(opts, transports...)
	opts = append(opts, security...)
	opts = append(opts, natOptions...)
	
	// Add announce addresses
	if len(config.AnnounceAddresses) > 0 {
		announceAddrs := make([]multiaddr.Multiaddr, 0, len(config.AnnounceAddresses))
		for _, addr := range config.AnnounceAddresses {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("Invalid announce address %s: %v", addr, err)
				continue
			}
			announceAddrs = append(announceAddrs, maddr)
		}
		opts = append(opts, libp2p.AddrsFactory(func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
			return announceAddrs
		}))
	}
	
	// Create host
	libp2pHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	
	// Create enhanced host wrapper
	ctx, cancel := context.WithCancel(ctx)
	p2pHost := &P2PHost{
		Host:      libp2pHost,
		config:    config,
		protocols: make(map[protocol.ID]network.StreamHandler),
		metrics: &HostMetrics{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Setup network event handlers
	p2pHost.setupEventHandlers()
	
	// Start metrics collection
	go p2pHost.collectMetrics()
	
	log.Printf("P2P host created with ID: %s", libp2pHost.ID())
	log.Printf("Listen addresses: %v", libp2pHost.Addrs())
	
	return p2pHost, nil
}

// setupEventHandlers configures network event handlers
func (h *P2PHost) setupEventHandlers() {
	network := h.Host.Network()
	
	// Connection events
	network.Notify(&network.NotifyBundle{
		ConnectedF: func(net network.Network, conn network.Conn) {
			h.metrics.ConnectionCount++
			h.metrics.LastActivity = time.Now()
			log.Printf("Connected to peer: %s", conn.RemotePeer())
			
			if h.connectHandler != nil {
				h.connectHandler(net, conn)
			}
		},
		DisconnectedF: func(net network.Network, conn network.Conn) {
			h.metrics.ConnectionCount--
			h.metrics.LastActivity = time.Now()
			log.Printf("Disconnected from peer: %s", conn.RemotePeer())
			
			if h.disconnectHandler != nil {
				h.disconnectHandler(net, conn)
			}
		},
	})
}

// collectMetrics periodically collects host metrics
func (h *P2PHost) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.updateMetrics()
		}
	}
}

// updateMetrics updates host metrics
func (h *P2PHost) updateMetrics() {
	network := h.Host.Network()
	
	// Update connection count
	h.metrics.ConnectionCount = len(network.Peers())
	
	// Update stream count
	streamCount := 0
	for _, peer := range network.Peers() {
		streams := network.ConnsToPeer(peer)
		for _, conn := range streams {
			streamCount += len(conn.GetStreams())
		}
	}
	h.metrics.StreamCount = streamCount
	
	// Update protocol handler count
	h.metrics.ProtocolHandlers = len(h.protocols)
	
	h.metrics.LastActivity = time.Now()
}

// RegisterProtocol registers a protocol handler
func (h *P2PHost) RegisterProtocol(protocolID protocol.ID, handler network.StreamHandler) {
	h.Host.SetStreamHandler(protocolID, handler)
	h.protocols[protocolID] = handler
	log.Printf("Registered protocol: %s", protocolID)
}

// UnregisterProtocol unregisters a protocol handler
func (h *P2PHost) UnregisterProtocol(protocolID protocol.ID) {
	h.Host.RemoveStreamHandler(protocolID)
	delete(h.protocols, protocolID)
	log.Printf("Unregistered protocol: %s", protocolID)
}

// GetMetrics returns current host metrics
func (h *P2PHost) GetMetrics() *HostMetrics {
	return h.metrics
}

// GetConfig returns host configuration
func (h *P2PHost) GetConfig() *config.NodeConfig {
	return h.config
}

// SetCapabilities sets node capabilities
func (h *P2PHost) SetCapabilities(caps *config.NodeCapabilities) {
	h.capabilities = caps
}

// GetCapabilities returns node capabilities
func (h *P2PHost) GetCapabilities() *config.NodeCapabilities {
	return h.capabilities
}

// OnConnect sets connection event handler
func (h *P2PHost) OnConnect(handler func(network.Network, network.Conn)) {
	h.connectHandler = handler
}

// OnDisconnect sets disconnection event handler
func (h *P2PHost) OnDisconnect(handler func(network.Network, network.Conn)) {
	h.disconnectHandler = handler
}

// Close closes the host and releases resources
func (h *P2PHost) Close() error {
	h.cancel()
	return h.Host.Close()
}

// loadOrGenerateKey loads existing key or generates new one
func loadOrGenerateKey(config *config.NodeConfig) (crypto.PrivKey, error) {
	// Try to load existing key
	if config.PrivateKey != "" {
		return config.GetPrivateKey()
	}
	
	// Generate new key
	if err := config.GenerateKey(); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	
	return config.GetPrivateKey()
}

// GetPeerCount returns number of connected peers
func (h *P2PHost) GetPeerCount() int {
	return len(h.Host.Network().Peers())
}

// GetConnectedPeers returns list of connected peers
func (h *P2PHost) GetConnectedPeers() []peer.ID {
	return h.Host.Network().Peers()
}

// IsConnected checks if peer is connected
func (h *P2PHost) IsConnected(peerID peer.ID) bool {
	return h.Host.Network().Connectedness(peerID) == network.Connected
}

// GetProtocols returns registered protocols
func (h *P2PHost) GetProtocols() []protocol.ID {
	protocols := make([]protocol.ID, 0, len(h.protocols))
	for id := range h.protocols {
		protocols = append(protocols, id)
	}
	return protocols
}