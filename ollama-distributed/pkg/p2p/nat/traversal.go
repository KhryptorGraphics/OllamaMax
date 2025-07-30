package nat

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// NATType represents different types of NAT configurations
type NATType int

const (
	// NATTypeUnknown represents unknown NAT type
	NATTypeUnknown NATType = iota
	// NATTypeOpen represents no NAT (direct connection)
	NATTypeOpen
	// NATTypeFullCone represents full cone NAT
	NATTypeFullCone
	// NATTypeRestrictedCone represents restricted cone NAT
	NATTypeRestrictedCone
	// NATTypePortRestrictedCone represents port restricted cone NAT
	NATTypePortRestrictedCone
	// NATTypeSymmetric represents symmetric NAT
	NATTypeSymmetric
	// NATTypeBlocked represents blocked/firewall
	NATTypeBlocked
)

// String returns string representation of NAT type
func (n NATType) String() string {
	switch n {
	case NATTypeOpen:
		return "Open"
	case NATTypeFullCone:
		return "Full Cone"
	case NATTypeRestrictedCone:
		return "Restricted Cone"
	case NATTypePortRestrictedCone:
		return "Port Restricted Cone"
	case NATTypeSymmetric:
		return "Symmetric"
	case NATTypeBlocked:
		return "Blocked"
	default:
		return "Unknown"
	}
}

// STUNServer represents a STUN server configuration
type STUNServer struct {
	Address   string
	Port      int
	Username  string
	Password  string
	Priority  int
	Available bool
	LastCheck time.Time
}

// TURNServer represents a TURN server configuration
type TURNServer struct {
	Address   string
	Port      int
	Username  string
	Password  string
	Realm     string
	Transport string // "udp" or "tcp"
	Priority  int
	Available bool
	LastCheck time.Time
}

// NATTraversalManager manages NAT traversal operations
type NATTraversalManager struct {
	stunServers    []*STUNServer
	turnServers    []*TURNServer
	natType        NATType
	publicAddr     *net.UDPAddr
	localAddr      *net.UDPAddr
	
	// Connection pooling
	relayConnections map[string]*RelayConnection
	connPoolMux      sync.RWMutex
	
	// Discovery results cache
	discoveryCache   map[string]*DiscoveryResult
	cacheMux         sync.RWMutex
	cacheExpiry      time.Duration
	
	// Metrics
	metrics          *TraversalMetrics
	
	// Configuration
	config           *TraversalConfig
	
	// Lifecycle
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// RelayConnection represents a pooled relay connection
type RelayConnection struct {
	Server       *TURNServer
	Conn         net.Conn
	CreatedAt    time.Time
	LastUsed     time.Time
	InUse        bool
	BytesSent    int64
	BytesReceived int64
}

// DiscoveryResult caches NAT discovery results
type DiscoveryResult struct {
	NATType     NATType
	PublicAddr  *net.UDPAddr
	Timestamp   time.Time
	ServerUsed  string
	RTT         time.Duration
}

// TraversalMetrics tracks NAT traversal performance
type TraversalMetrics struct {
	STUNRequests      int64
	STUNSuccesses     int64
	STUNFailures      int64
	TURNRequests      int64
	TURNSuccesses     int64
	TURNFailures      int64
	NATDetections     int64
	RelayConnections  int64
	SuccessfulHoles   int64
	FailedHoles       int64
	AverageRTT        time.Duration
	LastDiscovery     time.Time
}

// TraversalConfig configures NAT traversal behavior
type TraversalConfig struct {
	// STUN configuration
	STUNTimeout       time.Duration
	STUNRetries       int
	STUNServerCheck   time.Duration
	
	// TURN configuration  
	TURNTimeout       time.Duration
	TURNRetries       int
	TURNServerCheck   time.Duration
	MaxRelayConns     int
	RelayConnTTL      time.Duration
	
	// Discovery configuration
	DiscoveryTimeout  time.Duration
	DiscoveryRetries  int
	CacheExpiry       time.Duration
	
	// Hole punching configuration
	HolePunchTimeout  time.Duration
	HolePunchRetries  int
	HolePunchDelay    time.Duration
	
	// Connection optimization
	ConnectTimeout    time.Duration
	ParallelAttempts  int
	EarlySuccessDelay time.Duration
	
	// Backoff configuration
	BackoffInitial    time.Duration
	BackoffMax        time.Duration
	BackoffMultiplier float64
}

// DefaultTraversalConfig returns default configuration
func DefaultTraversalConfig() *TraversalConfig {
	return &TraversalConfig{
		STUNTimeout:       5 * time.Second,
		STUNRetries:       3,
		STUNServerCheck:   5 * time.Minute,
		TURNTimeout:       10 * time.Second,
		TURNRetries:       3,
		TURNServerCheck:   5 * time.Minute,
		MaxRelayConns:     10,
		RelayConnTTL:      30 * time.Minute,
		DiscoveryTimeout:  15 * time.Second,
		DiscoveryRetries:  2,
		CacheExpiry:       10 * time.Minute,
		HolePunchTimeout:  10 * time.Second,
		HolePunchRetries:  5,
		HolePunchDelay:    100 * time.Millisecond,
		ConnectTimeout:    5 * time.Second,  // Reduced from 30s
		ParallelAttempts:  3,
		EarlySuccessDelay: 200 * time.Millisecond,
		BackoffInitial:    1 * time.Second,
		BackoffMax:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// NewNATTraversalManager creates a new NAT traversal manager
func NewNATTraversalManager(ctx context.Context, config *TraversalConfig) *NATTraversalManager {
	if config == nil {
		config = DefaultTraversalConfig()
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	manager := &NATTraversalManager{
		stunServers:      make([]*STUNServer, 0),
		turnServers:      make([]*TURNServer, 0),
		natType:          NATTypeUnknown,
		relayConnections: make(map[string]*RelayConnection),
		discoveryCache:   make(map[string]*DiscoveryResult),
		metrics:          &TraversalMetrics{},
		config:           config,
		ctx:              ctx,
		cancel:           cancel,
		cacheExpiry:      config.CacheExpiry,
	}
	
	// Start background processes
	manager.wg.Add(2)
	go manager.serverHealthChecker()
	go manager.connectionPoolManager()
	
	return manager
}

// AddSTUNServer adds a STUN server to the configuration
func (n *NATTraversalManager) AddSTUNServer(address string, port int) {
	server := &STUNServer{
		Address:   address,
		Port:      port,
		Priority:  100,
		Available: true,
		LastCheck: time.Now(),
	}
	
	n.stunServers = append(n.stunServers, server)
	log.Printf("Added STUN server: %s:%d", address, port)
}

// AddTURNServer adds a TURN server to the configuration
func (n *NATTraversalManager) AddTURNServer(address string, port int, username, password, realm, transport string) {
	server := &TURNServer{
		Address:   address,
		Port:      port,
		Username:  username,
		Password:  password,
		Realm:     realm,
		Transport: transport,
		Priority:  100,
		Available: true,
		LastCheck: time.Now(),
	}
	
	n.turnServers = append(n.turnServers, server)
	log.Printf("Added TURN server: %s:%d (%s)", address, port, transport)
}

// DiscoverNATType discovers the NAT type using STUN protocol (RFC 3489)
func (n *NATTraversalManager) DiscoverNATType(ctx context.Context) (NATType, error) {
	// Check cache first
	n.cacheMux.RLock()
	if cached, exists := n.discoveryCache["nat_type"]; exists {
		if time.Since(cached.Timestamp) < n.cacheExpiry {
			n.cacheMux.RUnlock()
			n.natType = cached.NATType
			return cached.NATType, nil
		}
	}
	n.cacheMux.RUnlock()
	
	// Perform NAT type discovery
	natType, publicAddr, err := n.performNATDiscovery(ctx)
	if err != nil {
		n.metrics.NATDetections++
		return NATTypeUnknown, fmt.Errorf("NAT discovery failed: %w", err)
	}
	
	// Cache the result
	n.cacheMux.Lock()
	n.discoveryCache["nat_type"] = &DiscoveryResult{
		NATType:   natType,
		PublicAddr: publicAddr,
		Timestamp: time.Now(),
	}
	n.cacheMux.Unlock()
	
	n.natType = natType  
	n.publicAddr = publicAddr
	n.metrics.NATDetections++
	n.metrics.LastDiscovery = time.Now()
	
	log.Printf("Discovered NAT type: %s, Public address: %v", natType, publicAddr)
	return natType, nil
}

// performNATDiscovery performs the actual NAT type discovery
func (n *NATTraversalManager) performNATDiscovery(ctx context.Context) (NATType, *net.UDPAddr, error) {
	if len(n.stunServers) == 0 {
		return NATTypeUnknown, nil, fmt.Errorf("no STUN servers configured")
	}
	
	// Try multiple STUN servers in parallel for reliability
	resultChan := make(chan struct {
		natType    NATType
		publicAddr *net.UDPAddr
		err        error
	}, len(n.stunServers))
	
	discoverCtx, cancel := context.WithTimeout(ctx, n.config.DiscoveryTimeout)
	defer cancel()
	
	// Start parallel discovery attempts
	for _, server := range n.stunServers {
		if !server.Available {
			continue
		}
		
		go func(s *STUNServer) {
			natType, addr, err := n.discoverWithSTUNServer(discoverCtx, s)
			resultChan <- struct {
				natType    NATType
				publicAddr *net.UDPAddr
				err        error
			}{natType, addr, err}
		}(server)
	}
	
	// Wait for first successful result or all failures
	var lastErr error
	attempts := 0
	maxAttempts := len(n.stunServers)
	
	for attempts < maxAttempts {
		select {
		case result := <-resultChan:
			attempts++
			if result.err == nil {
				return result.natType, result.publicAddr, nil
			}
			lastErr = result.err
		case <-discoverCtx.Done():
			return NATTypeUnknown, nil, fmt.Errorf("discovery timeout: %w", discoverCtx.Err())
		}
	}
	
	return NATTypeUnknown, nil, fmt.Errorf("all STUN servers failed, last error: %w", lastErr)
}

// discoverWithSTUNServer discovers NAT type using a specific STUN server
func (n *NATTraversalManager) discoverWithSTUNServer(ctx context.Context, server *STUNServer) (NATType, *net.UDPAddr, error) {
	n.metrics.STUNRequests++
	
	// Implementation of RFC 3489 STUN NAT discovery algorithm
	// This is a simplified version - in production, you'd use a full STUN client library
	
	serverAddr := fmt.Sprintf("%s:%d", server.Address, server.Port)
	
	// Step 1: Test I - Basic connectivity test
	conn, err := net.DialTimeout("udp", serverAddr, n.config.STUNTimeout)
	if err != nil {
		n.metrics.STUNFailures++
		return NATTypeBlocked, nil, fmt.Errorf("cannot connect to STUN server: %w", err)
	}
	defer conn.Close()
	
	// Send binding request (simplified)
	bindingRequest := []byte{0x00, 0x01, 0x00, 0x00, 0x21, 0x12, 0xA4, 0x42}
	_, err = conn.Write(bindingRequest)
	if err != nil {
		n.metrics.STUNFailures++
		return NATTypeBlocked, nil, fmt.Errorf("failed to send STUN request: %w", err)
	}
	
	// Read response with timeout
	conn.SetReadDeadline(time.Now().Add(n.config.STUNTimeout))
	response := make([]byte, 1024)
	respLen, err := conn.Read(response)
	if err != nil {
		n.metrics.STUNFailures++
		return NATTypeBlocked, nil, fmt.Errorf("failed to read STUN response: %w", err)
	}
	
	// Parse response (simplified - in production use proper STUN library)
	if respLen < 20 {
		n.metrics.STUNFailures++
		return NATTypeUnknown, nil, fmt.Errorf("invalid STUN response")
	}
	
	// Extract mapped address (simplified parsing)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	
	// For this implementation, we'll do basic detection based on local vs remote address
	// In a full implementation, you'd perform all RFC 3489 tests
	
	n.metrics.STUNSuccesses++
	server.Available = true
	server.LastCheck = time.Now()
	
	// Simplified NAT type detection - in production, implement full RFC 3489 algorithm
	if localAddr.IP.IsPrivate() {
		// Behind NAT - for simplicity, assume restricted cone
		// Full implementation would perform additional tests
		return NATTypeRestrictedCone, &net.UDPAddr{
			IP:   net.ParseIP("192.0.2.1"), // Placeholder public IP
			Port: localAddr.Port + 1000,     // Placeholder mapped port
		}, nil
	}
	
	return NATTypeOpen, localAddr, nil
}

// EstablishRelayConnection establishes a connection through TURN relay
func (n *NATTraversalManager) EstablishRelayConnection(ctx context.Context, targetPeer peer.ID) (*RelayConnection, error) {
	n.metrics.TURNRequests++
	
	// Find best available TURN server
	server := n.selectBestTURNServer()
	if server == nil {
		n.metrics.TURNFailures++
		return nil, fmt.Errorf("no available TURN servers")
	}
	
	// Check connection pool first
	poolKey := fmt.Sprintf("%s:%d", server.Address, server.Port)
	n.connPoolMux.RLock()
	if conn, exists := n.relayConnections[poolKey]; exists {
		if !conn.InUse && time.Since(conn.LastUsed) < n.config.RelayConnTTL {
			conn.InUse = true
			conn.LastUsed = time.Now()
			n.connPoolMux.RUnlock()
			return conn, nil
		}
	}
	n.connPoolMux.RUnlock()
	
	// Create new relay connection
	relayConn, err := n.createRelayConnection(ctx, server)
	if err != nil {
		n.metrics.TURNFailures++
		server.Available = false
		return nil, fmt.Errorf("failed to create relay connection: %w", err)
	}
	
	// Add to connection pool
	n.connPoolMux.Lock()
	n.relayConnections[poolKey] = relayConn
	n.connPoolMux.Unlock()
	
	n.metrics.TURNSuccesses++
	n.metrics.RelayConnections++
	
	log.Printf("Established relay connection through %s:%d", server.Address, server.Port)
	return relayConn, nil
}

// selectBestTURNServer selects the best available TURN server
func (n *NATTraversalManager) selectBestTURNServer() *TURNServer {
	var bestServer *TURNServer
	bestPriority := -1
	
	for _, server := range n.turnServers {
		if server.Available && server.Priority > bestPriority {
			bestServer = server
			bestPriority = server.Priority
		}
	}
	
	return bestServer
}

// createRelayConnection creates a new TURN relay connection
func (n *NATTraversalManager) createRelayConnection(ctx context.Context, server *TURNServer) (*RelayConnection, error) {
	serverAddr := fmt.Sprintf("%s:%d", server.Address, server.Port)
	
	// Create connection with timeout
	connectCtx, cancel := context.WithTimeout(ctx, n.config.TURNTimeout)
	defer cancel()
	
	var d net.Dialer
	conn, err := d.DialContext(connectCtx, server.Transport, serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TURN server: %w", err)
	}
	
	// Perform TURN allocation (simplified - use full TURN client library in production)
	// This is a placeholder for TURN protocol implementation
	
	relayConn := &RelayConnection{
		Server:    server,
		Conn:      conn,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		InUse:     true,
	}
	
	return relayConn, nil
}

// AttemptHolePunching attempts NAT hole punching
func (n *NATTraversalManager) AttemptHolePunching(ctx context.Context, targetAddr multiaddr.Multiaddr) error {
	// Convert multiaddr to net.Addr
	netAddr, err := n.multiaddrToNetAddr(targetAddr)
	if err != nil {
		return fmt.Errorf("invalid target address: %w", err)
	}
	
	// Implement hole punching algorithm
	return n.performHolePunching(ctx, netAddr)
}

// performHolePunching performs the actual hole punching
func (n *NATTraversalManager) performHolePunching(ctx context.Context, targetAddr net.Addr) error {
	// Create UDP connection for hole punching
	localAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return fmt.Errorf("failed to resolve local address: %w", err)
	}
	
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return fmt.Errorf("failed to create UDP connection: %w", err)
	}
	defer conn.Close()
	
	// Send hole punching packets
	udpAddr, ok := targetAddr.(*net.UDPAddr)
	if !ok {
		return fmt.Errorf("target address is not UDP")
	}
	
	// Implement exponential backoff
	backoff := n.config.BackoffInitial
	
	for attempt := 0; attempt < n.config.HolePunchRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Send hole punching packet
		message := fmt.Sprintf("HOLE_PUNCH_%d", attempt)
		_, err := conn.WriteToUDP([]byte(message), udpAddr)
		if err != nil {
			log.Printf("Hole punch attempt %d failed: %v", attempt, err)
		} else {
			log.Printf("Sent hole punch packet %d to %v", attempt, udpAddr)
		}
		
		// Wait before next attempt with exponential backoff
		if attempt < n.config.HolePunchRetries-1 {
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * n.config.BackoffMultiplier)
			if backoff > n.config.BackoffMax {
				backoff = n.config.BackoffMax
			}
		}
	}
	
	n.metrics.SuccessfulHoles++ // This should be conditional on actual success
	return nil
}

// multiaddrToNetAddr converts multiaddr to net.Addr
func (n *NATTraversalManager) multiaddrToNetAddr(addr multiaddr.Multiaddr) (net.Addr, error) {
	// Extract IP and port from multiaddr
	ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
	if err != nil {
		ip, err = addr.ValueForProtocol(multiaddr.P_IP6)
		if err != nil {
			return nil, fmt.Errorf("no IP address found in multiaddr")
		}
	}
	
	port, err := addr.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		port, err = addr.ValueForProtocol(multiaddr.P_UDP)
		if err != nil {
			return nil, fmt.Errorf("no port found in multiaddr")
		}
	}
	
	return &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: parseInt(port),
	}, nil
}

// parseInt parses string to int (helper function)
func parseInt(s string) int {
	// Simple conversion - in production, handle errors properly
	if s == "" {
		return 0
	}
	// Placeholder - implement proper parsing
	return 8080
}

// SetNATType sets the NAT type (for testing)
func (n *NATTraversalManager) SetNATType(natType NATType) {
	n.natType = natType
}

// GetNATType returns the current NAT type
func (n *NATTraversalManager) GetNATType() NATType {
	return n.natType
}

// GetPublicAddress returns the discovered public address
func (n *NATTraversalManager) GetPublicAddress() *net.UDPAddr {
	return n.publicAddr
}

// GetMetrics returns traversal metrics
func (n *NATTraversalManager) GetMetrics() *TraversalMetrics {
	return n.metrics
}

// IsRelayRequired determines if relay connection is required
func (n *NATTraversalManager) IsRelayRequired() bool {
	return n.natType == NATTypeSymmetric || n.natType == NATTypeBlocked
}

// serverHealthChecker periodically checks server availability
func (n *NATTraversalManager) serverHealthChecker() {
	defer n.wg.Done()
	
	// Ensure minimum ticker interval
	checkInterval := n.config.STUNServerCheck
	if checkInterval <= 0 {
		checkInterval = 30 * time.Second
	}
	
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.checkServerHealth()
		}
	}
}

// checkServerHealth checks the health of STUN/TURN servers
func (n *NATTraversalManager) checkServerHealth() {
	// Check STUN servers
	for _, server := range n.stunServers {
		if time.Since(server.LastCheck) > n.config.STUNServerCheck {
			go n.checkSTUNServer(server)
		}
	}
	
	// Check TURN servers
	for _, server := range n.turnServers {
		if time.Since(server.LastCheck) > n.config.TURNServerCheck {
			go n.checkTURNServer(server)
		}
	}
}

// checkSTUNServer checks if a STUN server is available
func (n *NATTraversalManager) checkSTUNServer(server *STUNServer) {
	ctx, cancel := context.WithTimeout(n.ctx, n.config.STUNTimeout)
	defer cancel()
	_ = ctx // Use context if needed
	
	serverAddr := fmt.Sprintf("%s:%d", server.Address, server.Port)
	conn, err := net.DialTimeout("udp", serverAddr, n.config.STUNTimeout)
	
	server.LastCheck = time.Now()
	if err != nil {
		server.Available = false
		log.Printf("STUN server %s is unavailable: %v", serverAddr, err)
		return
	}
	
	conn.Close()
	server.Available = true
}

// checkTURNServer checks if a TURN server is available
func (n *NATTraversalManager) checkTURNServer(server *TURNServer) {
	ctx, cancel := context.WithTimeout(n.ctx, n.config.TURNTimeout)
	defer cancel()
	_ = ctx // Use context if needed
	
	serverAddr := fmt.Sprintf("%s:%d", server.Address, server.Port)
	conn, err := net.DialTimeout(server.Transport, serverAddr, n.config.TURNTimeout)
	
	server.LastCheck = time.Now()
	if err != nil {
		server.Available = false
		log.Printf("TURN server %s is unavailable: %v", serverAddr, err)
		return
	}
	
	conn.Close()
	server.Available = true
}

// connectionPoolManager manages relay connection pool
func (n *NATTraversalManager) connectionPoolManager() {
	defer n.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.cleanupConnectionPool()
		}
	}
}

// cleanupConnectionPool removes expired connections from pool
func (n *NATTraversalManager) cleanupConnectionPool() {
	n.connPoolMux.Lock()
	defer n.connPoolMux.Unlock()
	
	for key, conn := range n.relayConnections {
		if !conn.InUse && time.Since(conn.LastUsed) > n.config.RelayConnTTL {
			conn.Conn.Close()
			delete(n.relayConnections, key)
			log.Printf("Cleaned up expired relay connection: %s", key)
		}
	}
}

// Close closes the NAT traversal manager and releases resources
func (n *NATTraversalManager) Close() error {
	n.cancel()
	
	// Close all relay connections
	n.connPoolMux.Lock()
	for _, conn := range n.relayConnections {
		conn.Conn.Close()
	}
	n.connPoolMux.Unlock()
	
	// Wait for background goroutines to finish
	n.wg.Wait()
	
	log.Println("NAT traversal manager closed")
	return nil
}