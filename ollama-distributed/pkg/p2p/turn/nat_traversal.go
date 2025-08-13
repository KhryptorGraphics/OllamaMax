package turn

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// NATTraversal manages NAT traversal using TURN and other techniques
type NATTraversal struct {
	config *NATTraversalConfig

	// TURN components
	turnServer *TURNServer
	turnClient *TURNClient

	// STUN client for NAT detection
	stunClient *STUNClient

	// ICE candidate gathering
	iceGatherer *ICEGatherer

	// Connection management
	connections   map[peer.ID]*NATConnection
	connectionsMu sync.RWMutex

	// NAT type detection
	natType     NATType
	publicIP    net.IP
	mappedPorts map[int]int

	// Metrics
	metrics *NATTraversalMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NATTraversalConfig configures NAT traversal
type NATTraversalConfig struct {
	// TURN server settings
	TURNServerConfig *TURNConfig
	EnableTURNServer bool

	// TURN client settings
	TURNClientConfig *TURNClientConfig
	TURNServers      []string

	// STUN settings
	STUNServers []string
	STUNTimeout time.Duration

	// ICE settings
	EnableICE           bool
	ICETimeout          time.Duration
	ICEGatheringTimeout time.Duration

	// Connection settings
	ConnectionTimeout time.Duration
	MaxConnections    int

	// NAT detection
	EnableNATDetection  bool
	NATDetectionTimeout time.Duration
}

// NATConnection represents a NAT-traversed connection
type NATConnection struct {
	PeerID     peer.ID
	LocalAddr  *net.UDPAddr
	RemoteAddr *net.UDPAddr
	RelayAddr  *net.UDPAddr

	// Connection type
	ConnectionType ConnectionType

	// ICE candidates
	LocalCandidates  []*ICECandidate
	RemoteCandidates []*ICECandidate
	SelectedPair     *CandidatePair

	// State
	State        ConnectionState
	CreatedAt    time.Time
	ConnectedAt  time.Time
	LastActivity time.Time

	// Statistics
	BytesSent       int64
	BytesReceived   int64
	PacketsSent     int64
	PacketsReceived int64
	RTT             time.Duration

	mu sync.RWMutex
}

// STUNClient implements a STUN client for NAT detection
type STUNClient struct {
	servers  []string
	timeout  time.Duration
	publicIP net.IP
	natType  NATType
	mu       sync.RWMutex
}

// ICEGatherer gathers ICE candidates for connectivity
type ICEGatherer struct {
	config        *ICEConfig
	candidates    []*ICECandidate
	candidatesMu  sync.RWMutex
	gatheringDone chan struct{}
}

// ICECandidate represents an ICE candidate
type ICECandidate struct {
	Type        CandidateType
	Protocol    string
	Address     *net.UDPAddr
	Priority    uint32
	Foundation  string
	ComponentID uint16
	RelatedAddr *net.UDPAddr
}

// CandidatePair represents a pair of ICE candidates
type CandidatePair struct {
	Local     *ICECandidate
	Remote    *ICECandidate
	Priority  uint64
	State     PairState
	Nominated bool
}

// ICEConfig configures ICE gathering
type ICEConfig struct {
	STUNServers   []string
	TURNServers   []string
	GatherTimeout time.Duration
	CheckTimeout  time.Duration
}

// NATTraversalMetrics tracks NAT traversal performance
type NATTraversalMetrics struct {
	// Connection metrics
	TotalConnections      int64
	SuccessfulConnections int64
	FailedConnections     int64
	ActiveConnections     int64

	// NAT traversal success rates
	DirectConnections int64
	STUNConnections   int64
	TURNConnections   int64
	ICEConnections    int64

	// Performance metrics
	AverageConnectionTime time.Duration
	AverageRTT            time.Duration

	// Error metrics
	NATDetectionErrors int64
	STUNErrors         int64
	TURNErrors         int64
	ICEErrors          int64

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// Enums and constants
type NATType string

const (
	NATTypeNone           NATType = "none"
	NATTypeFullCone       NATType = "full_cone"
	NATTypeRestrictedCone NATType = "restricted_cone"
	NATTypePortRestricted NATType = "port_restricted"
	NATTypeSymmetric      NATType = "symmetric"
	NATTypeUnknown        NATType = "unknown"
)

type ConnectionType string

const (
	ConnectionTypeDirect ConnectionType = "direct"
	ConnectionTypeSTUN   ConnectionType = "stun"
	ConnectionTypeTURN   ConnectionType = "turn"
	ConnectionTypeICE    ConnectionType = "ice"
)

type ConnectionState string

const (
	ConnectionStateConnecting   ConnectionState = "connecting"
	ConnectionStateConnected    ConnectionState = "connected"
	ConnectionStateDisconnected ConnectionState = "disconnected"
	ConnectionStateFailed       ConnectionState = "failed"
)

type CandidateType string

const (
	CandidateTypeHost            CandidateType = "host"
	CandidateTypeServerReflexive CandidateType = "srflx"
	CandidateTypePeerReflexive   CandidateType = "prflx"
	CandidateTypeRelay           CandidateType = "relay"
)

type PairState string

const (
	PairStateWaiting    PairState = "waiting"
	PairStateInProgress PairState = "in_progress"
	PairStateSucceeded  PairState = "succeeded"
	PairStateFailed     PairState = "failed"
)

// NewNATTraversal creates a new NAT traversal manager
func NewNATTraversal(config *NATTraversalConfig) (*NATTraversal, error) {
	if config == nil {
		config = &NATTraversalConfig{
			EnableTURNServer:    false,
			TURNServers:         []string{"stun:stun.l.google.com:19302"},
			STUNServers:         []string{"stun:stun.l.google.com:19302"},
			STUNTimeout:         5 * time.Second,
			EnableICE:           true,
			ICETimeout:          30 * time.Second,
			ICEGatheringTimeout: 10 * time.Second,
			ConnectionTimeout:   30 * time.Second,
			MaxConnections:      1000,
			EnableNATDetection:  true,
			NATDetectionTimeout: 10 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	nt := &NATTraversal{
		config:      config,
		connections: make(map[peer.ID]*NATConnection),
		mappedPorts: make(map[int]int),
		metrics: &NATTraversalMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize STUN client
	nt.stunClient = &STUNClient{
		servers: config.STUNServers,
		timeout: config.STUNTimeout,
	}

	// Initialize ICE gatherer if enabled
	if config.EnableICE {
		nt.iceGatherer = &ICEGatherer{
			config: &ICEConfig{
				STUNServers:   config.STUNServers,
				TURNServers:   config.TURNServers,
				GatherTimeout: config.ICEGatheringTimeout,
				CheckTimeout:  config.ICETimeout,
			},
			candidates:    make([]*ICECandidate, 0),
			gatheringDone: make(chan struct{}),
		}
	}

	// Initialize TURN server if enabled
	if config.EnableTURNServer && config.TURNServerConfig != nil {
		turnServer, err := NewTURNServer(config.TURNServerConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TURN server: %w", err)
		}
		nt.turnServer = turnServer
	}

	// Initialize TURN client if TURN servers are configured
	if len(config.TURNServers) > 0 && config.TURNClientConfig != nil {
		turnClient, err := NewTURNClient(config.TURNClientConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TURN client: %w", err)
		}
		nt.turnClient = turnClient
	}

	return nt, nil
}

// Start starts the NAT traversal manager
func (nt *NATTraversal) Start() error {
	// Start TURN server if enabled
	if nt.turnServer != nil {
		if err := nt.turnServer.Start(); err != nil {
			return fmt.Errorf("failed to start TURN server: %w", err)
		}
	}

	// Connect TURN client if configured
	if nt.turnClient != nil {
		if err := nt.turnClient.Connect(); err != nil {
			return fmt.Errorf("failed to connect TURN client: %w", err)
		}
	}

	// Detect NAT type if enabled
	if nt.config.EnableNATDetection {
		nt.wg.Add(1)
		go nt.detectNATType()
	}

	// Start ICE gathering if enabled
	if nt.iceGatherer != nil {
		nt.wg.Add(1)
		go nt.gatherICECandidates()
	}

	// Start metrics loop
	nt.wg.Add(1)
	go nt.metricsLoop()

	return nil
}

// Stop stops the NAT traversal manager
func (nt *NATTraversal) Stop() error {
	nt.cancel()

	// Stop TURN server
	if nt.turnServer != nil {
		nt.turnServer.Stop()
	}

	// Disconnect TURN client
	if nt.turnClient != nil {
		nt.turnClient.Disconnect()
	}

	// Wait for goroutines
	nt.wg.Wait()

	return nil
}

// EstablishConnection establishes a NAT-traversed connection to a peer
func (nt *NATTraversal) EstablishConnection(peerID peer.ID, remoteAddr *net.UDPAddr) (*NATConnection, error) {
	nt.connectionsMu.Lock()
	defer nt.connectionsMu.Unlock()

	// Check if connection already exists
	if conn, exists := nt.connections[peerID]; exists {
		return conn, nil
	}

	// Create new connection
	conn := &NATConnection{
		PeerID:           peerID,
		RemoteAddr:       remoteAddr,
		State:            ConnectionStateConnecting,
		CreatedAt:        time.Now(),
		LocalCandidates:  make([]*ICECandidate, 0),
		RemoteCandidates: make([]*ICECandidate, 0),
	}

	nt.connections[peerID] = conn

	// Try different connection methods
	go nt.attemptConnection(conn)

	nt.metrics.mu.Lock()
	nt.metrics.TotalConnections++
	nt.metrics.ActiveConnections++
	nt.metrics.mu.Unlock()

	return conn, nil
}

// attemptConnection attempts to establish a connection using various methods
func (nt *NATTraversal) attemptConnection(conn *NATConnection) {
	// Try direct connection first
	if nt.tryDirectConnection(conn) {
		return
	}

	// Try STUN-assisted connection
	if nt.trySTUNConnection(conn) {
		return
	}

	// Try ICE connection
	if nt.iceGatherer != nil && nt.tryICEConnection(conn) {
		return
	}

	// Fall back to TURN relay
	if nt.turnClient != nil && nt.tryTURNConnection(conn) {
		return
	}

	// All methods failed
	conn.mu.Lock()
	conn.State = ConnectionStateFailed
	conn.mu.Unlock()

	nt.metrics.mu.Lock()
	nt.metrics.FailedConnections++
	nt.metrics.ActiveConnections--
	nt.metrics.mu.Unlock()
}

// tryDirectConnection attempts a direct connection
func (nt *NATTraversal) tryDirectConnection(conn *NATConnection) bool {
	// Implementation would attempt direct UDP connection
	// For now, this is a placeholder
	return false
}

// trySTUNConnection attempts a STUN-assisted connection
func (nt *NATTraversal) trySTUNConnection(conn *NATConnection) bool {
	// Implementation would use STUN for NAT traversal
	// For now, this is a placeholder
	return false
}

// tryICEConnection attempts an ICE connection
func (nt *NATTraversal) tryICEConnection(conn *NATConnection) bool {
	// Implementation would use ICE for connectivity establishment
	// For now, this is a placeholder
	return false
}

// tryTURNConnection attempts a TURN relay connection
func (nt *NATTraversal) tryTURNConnection(conn *NATConnection) bool {
	if nt.turnClient == nil {
		return false
	}

	// Get relay address
	relayAddr := nt.turnClient.GetRelayAddress()
	if relayAddr == nil {
		return false
	}

	conn.mu.Lock()
	conn.RelayAddr = relayAddr
	conn.ConnectionType = ConnectionTypeTURN
	conn.State = ConnectionStateConnected
	conn.ConnectedAt = time.Now()
	conn.mu.Unlock()

	nt.metrics.mu.Lock()
	nt.metrics.SuccessfulConnections++
	nt.metrics.TURNConnections++
	nt.metrics.mu.Unlock()

	return true
}

// detectNATType detects the NAT type using STUN
func (nt *NATTraversal) detectNATType() {
	defer nt.wg.Done()

	// Implementation would detect NAT type using STUN
	// For now, set to unknown
	nt.natType = NATTypeUnknown
}

// gatherICECandidates gathers ICE candidates
func (nt *NATTraversal) gatherICECandidates() {
	defer nt.wg.Done()

	if nt.iceGatherer == nil {
		return
	}

	// Implementation would gather ICE candidates
	// For now, this is a placeholder
	close(nt.iceGatherer.gatheringDone)
}

// metricsLoop updates metrics periodically
func (nt *NATTraversal) metricsLoop() {
	defer nt.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-nt.ctx.Done():
			return
		case <-ticker.C:
			nt.updateMetrics()
		}
	}
}

// updateMetrics updates NAT traversal metrics
func (nt *NATTraversal) updateMetrics() {
	nt.metrics.mu.Lock()
	defer nt.metrics.mu.Unlock()

	nt.connectionsMu.RLock()
	nt.metrics.ActiveConnections = int64(len(nt.connections))
	nt.connectionsMu.RUnlock()

	nt.metrics.LastUpdated = time.Now()
}

// GetMetrics returns NAT traversal metrics
func (nt *NATTraversal) GetMetrics() *NATTraversalMetrics {
	nt.metrics.mu.RLock()
	defer nt.metrics.mu.RUnlock()

	// Create a copy without the mutex
	return &NATTraversalMetrics{
		TotalConnections:      nt.metrics.TotalConnections,
		SuccessfulConnections: nt.metrics.SuccessfulConnections,
		FailedConnections:     nt.metrics.FailedConnections,
		ActiveConnections:     nt.metrics.ActiveConnections,
		DirectConnections:     nt.metrics.DirectConnections,
		STUNConnections:       nt.metrics.STUNConnections,
		TURNConnections:       nt.metrics.TURNConnections,
		ICEConnections:        nt.metrics.ICEConnections,
		AverageConnectionTime: nt.metrics.AverageConnectionTime,
		AverageRTT:            nt.metrics.AverageRTT,
		NATDetectionErrors:    nt.metrics.NATDetectionErrors,
		STUNErrors:            nt.metrics.STUNErrors,
		TURNErrors:            nt.metrics.TURNErrors,
		ICEErrors:             nt.metrics.ICEErrors,
		LastUpdated:           nt.metrics.LastUpdated,
	}
}

// GetNATType returns the detected NAT type
func (nt *NATTraversal) GetNATType() NATType {
	return nt.natType
}

// GetPublicIP returns the detected public IP
func (nt *NATTraversal) GetPublicIP() net.IP {
	return nt.publicIP
}
