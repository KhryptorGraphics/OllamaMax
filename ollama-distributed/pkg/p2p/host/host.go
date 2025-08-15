package host

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/multiformats/go-multiaddr"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/nat"
)

// P2PHost wraps libp2p host with enhanced functionality
type P2PHost struct {
	host.Host
	config       *config.NodeConfig
	capabilities *config.NodeCapabilities
	relayService *relay.Relay

	// Protocol handlers
	protocols map[protocol.ID]network.StreamHandler

	// Event handlers
	connectHandler    func(network.Network, network.Conn)
	disconnectHandler func(network.Network, network.Conn)

	// Metrics
	metrics *HostMetrics

	// NAT traversal
	natManager *nat.NATTraversalManager

	// Connection management
	connectionTracker *ConnectionTracker
	connectionPool    *ConnectionPool
	bandwidthManager  *BandwidthManager

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// HostMetrics tracks host performance metrics
type HostMetrics struct {
	ConnectionCount  int
	StreamCount      int
	BytesReceived    int64
	BytesSent        int64
	ProtocolHandlers int
	LastActivity     time.Time
	StartTime        time.Time

	// NAT traversal metrics
	NATType            string
	STUNRequests       int64
	TURNConnections    int64
	HolePunchAttempts  int64
	HolePunchSuccesses int64

	// Connection optimization metrics
	ParallelConnections int
	EarlySuccesses      int64
	ConnectionTimeouts  int64
	BackoffRetries      int64
}

// ConnectionTracker manages optimized connection attempts
type ConnectionTracker struct {
	host           host.Host
	activeAttempts map[peer.ID]*ConnectionAttempt
	mux            sync.RWMutex
	config         *ConnectionConfig
}

// ConnectionAttempt tracks individual connection attempts
type ConnectionAttempt struct {
	PeerID      peer.ID
	StartTime   time.Time
	Attempts    int
	LastBackoff time.Duration
	Ctx         context.Context
	Cancel      context.CancelFunc
	ResultChan  chan error
}

// ConnectionConfig configures connection behavior
type ConnectionConfig struct {
	Timeout           time.Duration
	ParallelAttempts  int
	EarlySuccessDelay time.Duration
	BackoffInitial    time.Duration
	BackoffMax        time.Duration
	BackoffMultiplier float64
	MaxRetries        int
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
		libp2p.Transport(libp2pwebtransport.New),
	}

	// Configure security
	security := []libp2p.Option{}
	if config.EnableNoise {
		security = append(security, libp2p.Security(noise.ID, noise.New))
	}
	if config.EnableTLS {
		security = append(security, libp2p.Security(libp2ptls.ID, libp2ptls.New))
	}

	// Configure NAT traversal with enhanced capabilities
	natOptions := []libp2p.Option{}
	if config.EnableNATService {
		natOptions = append(natOptions, libp2p.EnableNATService())
	}
	if config.EnableHolePunching {
		natOptions = append(natOptions, libp2p.EnableHolePunching())
	}
	if config.EnableAutoRelay {
		// Parse StaticRelays []string into []peer.AddrInfo
		var staticInfos []peer.AddrInfo
		for _, addr := range config.StaticRelays {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("Invalid static relay address %s: %v", addr, err)
				continue
			}
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				log.Printf("Failed to parse static relay %s: %v", addr, err)
				continue
			}
			staticInfos = append(staticInfos, *info)
		}
		if len(staticInfos) > 0 {
			natOptions = append(natOptions, libp2p.EnableAutoRelayWithStaticRelays(staticInfos))
		}
	}

	// Create NAT traversal manager
	natManager := nat.NewNATTraversalManager(ctx, nil)

	// Add STUN servers for NAT discovery
	natManager.AddSTUNServer("stun.l.google.com", 19302)
	natManager.AddSTUNServer("stun1.l.google.com", 19302)
	natManager.AddSTUNServer("stun2.l.google.com", 19302)

	// Add TURN servers if configured
	if len(config.TURNServers) > 0 {
		for _, turnServer := range config.TURNServers {
			natManager.AddTURNServer(
				turnServer.Address,
				turnServer.Port,
				turnServer.Username,
				turnServer.Password,
				turnServer.Realm,
				turnServer.Transport,
			)
		}
	}

	// Create connection tracker with optimized settings
	connTracker := &ConnectionTracker{
		activeAttempts: make(map[peer.ID]*ConnectionAttempt),
		config: &ConnectionConfig{
			Timeout:           5 * time.Second, // Reduced from 30s
			ParallelAttempts:  3,
			EarlySuccessDelay: 200 * time.Millisecond,
			BackoffInitial:    1 * time.Second,
			BackoffMax:        30 * time.Second,
			BackoffMultiplier: 2.0,
			MaxRetries:        5,
		},
	}

	// Create connection pool
	poolConfig := &PoolConfig{
		MaxConnections:       100,
		MaxIdleTime:          5 * time.Minute,
		MaxConnectionAge:     30 * time.Minute,
		CleanupInterval:      1 * time.Minute,
		QualityCheckInterval: 30 * time.Second,
		MinQualityThreshold:  0.7,
		MaxStreamsPerConn:    10,
		StreamIdleTimeout:    2 * time.Minute,
	}

	// Create bandwidth manager
	bandwidthConfig := &BandwidthConfig{
		GlobalLimit:        100 * 1024 * 1024, // 100 MB/s
		DefaultPeerLimit:   10 * 1024 * 1024,  // 10 MB/s per peer
		WindowSize:         time.Second,
		BurstSize:          10 * 1024 * 1024, // 10 MB burst
		UpdateInterval:     time.Second,
		ProtocolLimits:     make(map[string]int64),
		PriorityProtocols:  []string{"/ollama/consensus/1.0.0", "/ollama/health/1.0.0"},
		PriorityMultiplier: 2.0,
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
		natManager:        natManager,
		connectionTracker: connTracker,
		connectionPool:    NewConnectionPool(libp2pHost, poolConfig),
		bandwidthManager:  NewBandwidthManager(bandwidthConfig),
		ctx:               ctx,
		cancel:            cancel,
	}

	// Set host reference for connection tracker
	connTracker.host = libp2pHost

	// Setup network event handlers
	p2pHost.setupEventHandlers()

	// Start metrics collection
	go p2pHost.collectMetrics()

	// Start NAT discovery
	go p2pHost.performNATDiscovery()

	log.Printf("P2P host created with ID: %s", libp2pHost.ID())
	log.Printf("Listen addresses: %v", libp2pHost.Addrs())

	return p2pHost, nil
}

// setupEventHandlers configures network event handlers
func (h *P2PHost) setupEventHandlers() {
	net := h.Host.Network()

	// Connection events
	notifee := &network.NotifyBundle{
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
	}

	net.Notify(notifee)
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

	// Update NAT traversal metrics
	if h.natManager != nil {
		natMetrics := h.natManager.GetMetrics()
		h.metrics.STUNRequests = natMetrics.STUNRequests
		h.metrics.TURNConnections = natMetrics.RelayConnections
		h.metrics.HolePunchAttempts = natMetrics.SuccessfulHoles + natMetrics.FailedHoles
		h.metrics.HolePunchSuccesses = natMetrics.SuccessfulHoles
	}

	// Update connection optimization metrics
	if h.connectionTracker != nil {
		h.connectionTracker.mux.RLock()
		h.metrics.ParallelConnections = len(h.connectionTracker.activeAttempts)
		h.connectionTracker.mux.RUnlock()
	}

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
func (h *P2PHost) SetCapabilities(caps interface{}) {
	// Accept both config.NodeCapabilities and resources.NodeCapabilities
	switch v := caps.(type) {
	case *config.NodeCapabilities:
		h.capabilities = v
	default:
		// For resources.NodeCapabilities, we don't store them directly
		// as they're managed by the resource layer
		log.Printf("Setting capabilities of type %T", caps)
	}
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

	// Close NAT manager
	if h.natManager != nil {
		h.natManager.Close()
	}

	// Close connection pool
	if h.connectionPool != nil {
		h.connectionPool.Close()
	}

	// Close bandwidth manager
	if h.bandwidthManager != nil {
		h.bandwidthManager.Close()
	}

	// Cancel active connection attempts
	h.connectionTracker.mux.Lock()
	for _, attempt := range h.connectionTracker.activeAttempts {
		attempt.Cancel()
	}
	h.connectionTracker.mux.Unlock()

	return h.Host.Close()
}

// performNATDiscovery performs NAT type discovery in background
func (h *P2PHost) performNATDiscovery() {
	ctx, cancel := context.WithTimeout(h.ctx, 15*time.Second)
	defer cancel()

	natType, err := h.natManager.DiscoverNATType(ctx)
	if err != nil {
		log.Printf("NAT discovery failed: %v", err)
		return
	}

	// Update metrics
	h.metrics.NATType = natType.String()
	log.Printf("NAT type discovered: %s", natType)

	// Update connection strategy based on NAT type
	h.updateConnectionStrategy(natType)
}

// updateConnectionStrategy updates connection strategy based on NAT type
func (h *P2PHost) updateConnectionStrategy(natType nat.NATType) {
	switch natType {
	case nat.NATTypeSymmetric, nat.NATTypeBlocked:
		// Prefer relay connections for difficult NAT types
		h.connectionTracker.config.ParallelAttempts = 5
		h.connectionTracker.config.Timeout = 10 * time.Second
		log.Printf("Using relay-preferred strategy for %s NAT", natType)

	case nat.NATTypeOpen, nat.NATTypeFullCone:
		// Direct connections work well
		h.connectionTracker.config.ParallelAttempts = 2
		h.connectionTracker.config.Timeout = 3 * time.Second
		log.Printf("Using direct connection strategy for %s NAT", natType)

	default:
		// Hole punching may work
		h.connectionTracker.config.ParallelAttempts = 3
		h.connectionTracker.config.Timeout = 5 * time.Second
		log.Printf("Using hole-punching strategy for %s NAT", natType)
	}
}

// ConnectWithOptimization connects to a peer with optimized strategy
func (h *P2PHost) ConnectWithOptimization(ctx context.Context, peerInfo peer.AddrInfo) error {
	return h.connectionTracker.ConnectOptimized(ctx, peerInfo)
}

// ConnectOptimized performs optimized connection with early success detection
func (c *ConnectionTracker) ConnectOptimized(ctx context.Context, peerInfo peer.AddrInfo) error {
	c.mux.Lock()

	// Check if already attempting connection
	if attempt, exists := c.activeAttempts[peerInfo.ID]; exists {
		c.mux.Unlock()
		// Wait for existing attempt
		select {
		case err := <-attempt.ResultChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Create new connection attempt
	attemptCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	attempt := &ConnectionAttempt{
		PeerID:      peerInfo.ID,
		StartTime:   time.Now(),
		Attempts:    1,
		LastBackoff: c.config.BackoffInitial,
		Ctx:         attemptCtx,
		Cancel:      cancel,
		ResultChan:  make(chan error, 1),
	}

	c.activeAttempts[peerInfo.ID] = attempt
	c.mux.Unlock()

	// Start parallel connection attempts
	go c.performParallelConnection(attempt, peerInfo)

	// Wait for result
	select {
	case err := <-attempt.ResultChan:
		c.mux.Lock()
		delete(c.activeAttempts, peerInfo.ID)
		c.mux.Unlock()
		return err
	case <-ctx.Done():
		attempt.Cancel()
		c.mux.Lock()
		delete(c.activeAttempts, peerInfo.ID)
		c.mux.Unlock()
		return ctx.Err()
	}
}

// performParallelConnection performs parallel connection attempts
func (c *ConnectionTracker) performParallelConnection(attempt *ConnectionAttempt, peerInfo peer.AddrInfo) {
	defer close(attempt.ResultChan)

	// Create multiple connection attempts in parallel
	resultChan := make(chan error, c.config.ParallelAttempts)
	successChan := make(chan struct{}, 1)

	// Start parallel connection attempts
	for i := 0; i < c.config.ParallelAttempts; i++ {
		go func(attemptNum int) {
			// Add slight delay for subsequent attempts
			if attemptNum > 0 {
				time.Sleep(time.Duration(attemptNum) * 100 * time.Millisecond)
			}

			err := c.host.Connect(attempt.Ctx, peerInfo)
			resultChan <- err

			// Signal early success
			if err == nil {
				select {
				case successChan <- struct{}{}:
				default:
				}
			}
		}(i)
	}

	// Wait for early success or all attempts to complete
	earlySuccessTimer := time.NewTimer(c.config.EarlySuccessDelay)
	defer earlySuccessTimer.Stop()

	var lastErr error
	completedAttempts := 0

	for {
		select {
		case <-successChan:
			// Early success detected
			attempt.ResultChan <- nil
			return

		case err := <-resultChan:
			completedAttempts++
			if err == nil {
				// Success
				attempt.ResultChan <- nil
				return
			}
			lastErr = err

			// Check if all attempts completed
			if completedAttempts >= c.config.ParallelAttempts {
				// All attempts failed, try with exponential backoff
				if attempt.Attempts < c.config.MaxRetries {
					go c.retryWithBackoff(attempt, peerInfo)
					return
				}

				// Final failure
				attempt.ResultChan <- lastErr
				return
			}

		case <-attempt.Ctx.Done():
			// Timeout
			attempt.ResultChan <- attempt.Ctx.Err()
			return
		}
	}
}

// retryWithBackoff retries connection with exponential backoff
func (c *ConnectionTracker) retryWithBackoff(attempt *ConnectionAttempt, peerInfo peer.AddrInfo) {
	// Wait for backoff period
	time.Sleep(attempt.LastBackoff)

	// Update backoff for next retry
	attempt.LastBackoff = time.Duration(float64(attempt.LastBackoff) * c.config.BackoffMultiplier)
	if attempt.LastBackoff > c.config.BackoffMax {
		attempt.LastBackoff = c.config.BackoffMax
	}
	attempt.Attempts++

	// Retry connection
	c.performParallelConnection(attempt, peerInfo)
}

// GetNATManager returns the NAT traversal manager
func (h *P2PHost) GetNATManager() *nat.NATTraversalManager {
	return h.natManager
}

// GetConnectionTracker returns the connection tracker
func (h *P2PHost) GetConnectionTracker() *ConnectionTracker {
	return h.connectionTracker
}

// GetConnectionPool returns the connection pool
func (h *P2PHost) GetConnectionPool() *ConnectionPool {
	return h.connectionPool
}

// GetBandwidthManager returns the bandwidth manager
func (h *P2PHost) GetBandwidthManager() *BandwidthManager {
	return h.bandwidthManager
}

// GetPooledConnection gets a connection from the pool
func (h *P2PHost) GetPooledConnection(ctx context.Context, peerID peer.ID) (*PooledConnection, error) {
	return h.connectionPool.GetConnection(ctx, peerID)
}

// GetPooledStream gets a stream from the connection pool
func (h *P2PHost) GetPooledStream(ctx context.Context, peerID peer.ID, protocolID protocol.ID) (network.Stream, error) {
	return h.connectionPool.GetStream(ctx, peerID, protocolID)
}

// ReturnPooledStream returns a stream to the pool for reuse
func (h *P2PHost) ReturnPooledStream(stream network.Stream, protocolID protocol.ID) {
	h.connectionPool.ReturnStream(stream, protocolID)
}

// CheckBandwidth checks if a data transfer is allowed
func (h *P2PHost) CheckBandwidth(peerID peer.ID, protocol string, bytes int64) bool {
	return h.bandwidthManager.CheckBandwidth(peerID, protocol, bytes)
}

// RecordBandwidthUsage records bandwidth usage
func (h *P2PHost) RecordBandwidthUsage(peerID peer.ID, protocol string, bytesSent, bytesReceived int64) {
	h.bandwidthManager.RecordUsage(peerID, protocol, bytesSent, bytesReceived)
}

// loadOrGenerateKey loads existing key or generates new one
func loadOrGenerateKey(config *config.NodeConfig) (crypto.PrivKey, error) {
	// Try to load existing key material from config if present in future
	// For now, always generate ephemeral key to avoid config dependency
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return priv, nil
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
