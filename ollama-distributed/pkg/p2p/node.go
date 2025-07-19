package p2p

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	
	"github.com/ollama/ollama-distributed/pkg/config"
	"github.com/ollama/ollama-distributed/pkg/p2p/discovery"
	p2phost "github.com/ollama/ollama-distributed/pkg/p2p/host"
	"github.com/ollama/ollama-distributed/pkg/p2p/resources"
	"github.com/ollama/ollama-distributed/pkg/p2p/routing"
	"github.com/ollama/ollama-distributed/pkg/p2p/security"
)

// P2PNode represents a complete P2P node implementation
type P2PNode struct {
	// Core components
	host              *p2phost.P2PHost
	config            *config.NodeConfig
	
	// Network components
	discoveryEngine   *discovery.DiscoveryEngine
	securityManager   *security.SecurityManager
	resourceAdvertiser *resources.ResourceAdvertiser
	contentRouter     *routing.ContentRouter
	
	// Node state
	capabilities      *config.NodeCapabilities
	resourceMetrics   *ResourceMetrics
	
	// Event handlers
	eventHandlers     map[string][]EventHandler
	eventMux          sync.RWMutex
	
	// Metrics
	metrics           *NodeMetrics
	
	// Lifecycle
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	started           bool
	startedMux        sync.RWMutex
}

// NodeMetrics tracks node performance metrics
type NodeMetrics struct {
	// Connection metrics
	ConnectedPeers    int
	TotalConnections  int
	ConnectionErrors  int
	
	// Discovery metrics
	PeersDiscovered   int
	DiscoveryErrors   int
	
	// Security metrics
	AuthAttempts      int
	AuthSuccesses     int
	AuthFailures      int
	
	// Resource metrics
	ResourcesAdvertised int
	ResourcesDiscovered int
	
	// Content metrics
	ContentPublished  int
	ContentRequests   int
	ContentProvided   int
	
	// Performance metrics
	AverageLatency    time.Duration
	MessageThroughput int64
	
	// Timestamps
	StartTime         time.Time
	LastActivity      time.Time
	Uptime            time.Duration
}

// EventHandler defines event handler interface
type EventHandler func(event *NodeEvent)

// NodeEvent represents a node event
type NodeEvent struct {
	Type      string
	Data      interface{}
	PeerID    peer.ID
	Timestamp time.Time
}

// Event types
const (
	EventPeerConnected    = "peer_connected"
	EventPeerDisconnected = "peer_disconnected"
	EventPeerDiscovered   = "peer_discovered"
	EventResourceUpdated  = "resource_updated"
	EventContentPublished = "content_published"
	EventContentRequested = "content_requested"
	EventAuthSuccess      = "auth_success"
	EventAuthFailure      = "auth_failure"
	EventError            = "error"
)

// NewP2PNode creates a new P2P node
func NewP2PNode(ctx context.Context, config *config.NodeConfig) (*P2PNode, error) {
	if config == nil {
		config = config.DefaultConfig()
	}
	
	// Generate key if not provided
	if config.PrivateKey == "" {
		if err := config.GenerateKey(); err != nil {
			return nil, fmt.Errorf("failed to generate node key: %w", err)
		}
	}
	
	ctx, cancel := context.WithCancel(ctx)
	
	node := &P2PNode{
		config:        config,
		eventHandlers: make(map[string][]EventHandler),
		metrics: &NodeMetrics{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize components
	if err := node.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}
	
	// Setup event handlers
	node.setupEventHandlers()
	
	log.Printf("P2P node initialized with ID: %s", node.host.ID())
	return node, nil
}

// initializeComponents initializes all node components
func (n *P2PNode) initializeComponents() error {
	var err error
	
	// Initialize libp2p host
	n.host, err = p2phost.NewP2PHost(n.ctx, n.config)
	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	
	// Initialize discovery engine
	n.discoveryEngine, err = discovery.NewDiscoveryEngine(n.ctx, n.host, n.config)
	if err != nil {
		return fmt.Errorf("failed to create discovery engine: %w", err)
	}
	
	// Initialize security manager
	securityConfig := security.DefaultSecurityConfig()
	n.securityManager, err = security.NewSecurityManager(n.ctx, n.host, securityConfig)
	if err != nil {
		return fmt.Errorf("failed to create security manager: %w", err)
	}
	
	// Initialize resource advertiser
	// Note: We'll need to get the DHT from discovery engine
	dht := n.discoveryEngine.GetDHT()
	if dht != nil {
		advertiserConfig := resources.DefaultAdvertiserConfig()
		n.resourceAdvertiser, err = resources.NewResourceAdvertiser(n.ctx, n.host, dht, advertiserConfig)
		if err != nil {
			return fmt.Errorf("failed to create resource advertiser: %w", err)
		}
	}
	
	// Initialize content router
	if dht != nil {
		routerConfig := routing.DefaultContentRouterConfig()
		n.contentRouter, err = routing.NewContentRouter(n.ctx, n.host, dht, routerConfig)
		if err != nil {
			return fmt.Errorf("failed to create content router: %w", err)
		}
	}
	
	return nil
}

// setupEventHandlers sets up internal event handlers
func (n *P2PNode) setupEventHandlers() {
	// Connection events
	n.host.OnConnect(func(net network.Network, conn network.Conn) {
		n.metrics.ConnectedPeers++
		n.metrics.TotalConnections++
		n.metrics.LastActivity = time.Now()
		
		n.emitEvent(EventPeerConnected, map[string]interface{}{
			"peer_id": conn.RemotePeer(),
			"addr":    conn.RemoteMultiaddr(),
		}, conn.RemotePeer())
	})
	
	n.host.OnDisconnect(func(net network.Network, conn network.Conn) {
		n.metrics.ConnectedPeers--
		n.metrics.LastActivity = time.Now()
		
		n.emitEvent(EventPeerDisconnected, map[string]interface{}{
			"peer_id": conn.RemotePeer(),
			"addr":    conn.RemoteMultiaddr(),
		}, conn.RemotePeer())
	})
}

// Start starts the P2P node
func (n *P2PNode) Start() error {
	n.startedMux.Lock()
	defer n.startedMux.Unlock()
	
	if n.started {
		return fmt.Errorf("node already started")
	}
	
	log.Printf("Starting P2P node...")
	
	// Start discovery engine
	n.discoveryEngine.Start()
	
	// Start resource advertiser
	if n.resourceAdvertiser != nil {
		n.resourceAdvertiser.Start()
	}
	
	// Start content router
	if n.contentRouter != nil {
		n.contentRouter.Start()
	}
	
	// Start metrics collection
	n.wg.Add(1)
	go n.metricsTask()
	
	// Start resource monitoring
	n.wg.Add(1)
	go n.resourceMonitoringTask()
	
	n.started = true
	log.Printf("P2P node started successfully")
	log.Printf("Node ID: %s", n.host.ID())
	log.Printf("Listen addresses: %v", n.host.Addrs())
	
	return nil
}

// Stop stops the P2P node
func (n *P2PNode) Stop() error {
	n.startedMux.Lock()
	defer n.startedMux.Unlock()
	
	if !n.started {
		return fmt.Errorf("node not started")
	}
	
	log.Printf("Stopping P2P node...")
	
	// Cancel context
	n.cancel()
	
	// Wait for background tasks
	n.wg.Wait()
	
	// Stop components
	if n.discoveryEngine != nil {
		n.discoveryEngine.Stop()
	}
	
	if n.resourceAdvertiser != nil {
		n.resourceAdvertiser.Stop()
	}
	
	if n.contentRouter != nil {
		n.contentRouter.Stop()
	}
	
	if n.securityManager != nil {
		n.securityManager.Close()
	}
	
	// Close host
	if n.host != nil {
		n.host.Close()
	}
	
	n.started = false
	log.Printf("P2P node stopped")
	
	return nil
}

// ConnectToPeer connects to a specific peer
func (n *P2PNode) ConnectToPeer(ctx context.Context, peerInfo peer.AddrInfo) error {
	if err := n.host.Connect(ctx, peerInfo); err != nil {
		n.metrics.ConnectionErrors++
		return fmt.Errorf("failed to connect to peer: %w", err)
	}
	
	log.Printf("Connected to peer: %s", peerInfo.ID)
	return nil
}

// DisconnectFromPeer disconnects from a specific peer
func (n *P2PNode) DisconnectFromPeer(peerID peer.ID) error {
	if err := n.host.Network().ClosePeer(peerID); err != nil {
		return fmt.Errorf("failed to disconnect from peer: %w", err)
	}
	
	log.Printf("Disconnected from peer: %s", peerID)
	return nil
}

// GetConnectedPeers returns list of connected peers
func (n *P2PNode) GetConnectedPeers() []peer.ID {
	return n.host.GetConnectedPeers()
}

// GetPeerCount returns number of connected peers
func (n *P2PNode) GetPeerCount() int {
	return n.host.GetPeerCount()
}

// IsConnected checks if peer is connected
func (n *P2PNode) IsConnected(peerID peer.ID) bool {
	return n.host.IsConnected(peerID)
}

// SetCapabilities sets node capabilities
func (n *P2PNode) SetCapabilities(caps *NodeCapabilities) {
	n.capabilities = caps
	n.host.SetCapabilities(caps)
	
	// Update advertiser
	if n.resourceAdvertiser != nil {
		n.resourceAdvertiser.SetCapabilities(caps)
	}
	
	n.emitEvent(EventResourceUpdated, caps, "")
}

// GetCapabilities returns node capabilities
func (n *P2PNode) GetCapabilities() *NodeCapabilities {
	return n.capabilities
}

// SetResourceMetrics sets resource metrics
func (n *P2PNode) SetResourceMetrics(metrics *ResourceMetrics) {
	n.resourceMetrics = metrics
	
	// Update advertiser
	if n.resourceAdvertiser != nil {
		n.resourceAdvertiser.SetResourceMetrics(metrics)
	}
	
	n.emitEvent(EventResourceUpdated, metrics, "")
}

// GetResourceMetrics returns resource metrics
func (n *P2PNode) GetResourceMetrics() *ResourceMetrics {
	return n.resourceMetrics
}

// PublishContent publishes content to the network
func (n *P2PNode) PublishContent(ctx context.Context, content *routing.ContentMetadata) error {
	if n.contentRouter == nil {
		return fmt.Errorf("content router not available")
	}
	
	if err := n.contentRouter.PublishContent(ctx, content); err != nil {
		return fmt.Errorf("failed to publish content: %w", err)
	}
	
	n.metrics.ContentPublished++
	n.emitEvent(EventContentPublished, content, "")
	
	return nil
}

// RequestContent requests content from the network
func (n *P2PNode) RequestContent(ctx context.Context, contentID string, priority int) (*routing.ContentRequest, error) {
	if n.contentRouter == nil {
		return nil, fmt.Errorf("content router not available")
	}
	
	request, err := n.contentRouter.RequestContent(ctx, contentID, priority)
	if err != nil {
		return nil, fmt.Errorf("failed to request content: %w", err)
	}
	
	n.metrics.ContentRequests++
	n.emitEvent(EventContentRequested, request, "")
	
	return request, nil
}

// FindContent finds content in the network
func (n *P2PNode) FindContent(ctx context.Context, contentID string) (*routing.ContentMetadata, []peer.ID, error) {
	if n.contentRouter == nil {
		return nil, nil, fmt.Errorf("content router not available")
	}
	
	return n.contentRouter.FindContent(ctx, contentID)
}

// EstablishSecureChannel establishes a secure channel with a peer
func (n *P2PNode) EstablishSecureChannel(ctx context.Context, peerID peer.ID) (*security.SecureChannel, error) {
	if n.securityManager == nil {
		return nil, fmt.Errorf("security manager not available")
	}
	
	return n.securityManager.EstablishSecureChannel(ctx, peerID)
}

// Event system

// On registers an event handler
func (n *P2PNode) On(eventType string, handler EventHandler) {
	n.eventMux.Lock()
	defer n.eventMux.Unlock()
	
	n.eventHandlers[eventType] = append(n.eventHandlers[eventType], handler)
}

// Off removes an event handler
func (n *P2PNode) Off(eventType string, handler EventHandler) {
	n.eventMux.Lock()
	defer n.eventMux.Unlock()
	
	handlers := n.eventHandlers[eventType]
	for i, h := range handlers {
		// Note: In Go, we can't directly compare functions
		// This is a simplified implementation
		if &h == &handler {
			n.eventHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// emitEvent emits an event
func (n *P2PNode) emitEvent(eventType string, data interface{}, peerID peer.ID) {
	n.eventMux.RLock()
	handlers := n.eventHandlers[eventType]
	n.eventMux.RUnlock()
	
	if len(handlers) == 0 {
		return
	}
	
	event := &NodeEvent{
		Type:      eventType,
		Data:      data,
		PeerID:    peerID,
		Timestamp: time.Now(),
	}
	
	// Call handlers in separate goroutines
	for _, handler := range handlers {
		go handler(event)
	}
}

// Background tasks

// metricsTask collects node metrics
func (n *P2PNode) metricsTask() {
	defer n.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.updateMetrics()
		}
	}
}

// resourceMonitoringTask monitors resource usage
func (n *P2PNode) resourceMonitoringTask() {
	defer n.wg.Done()
	
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.updateResourceMetrics()
		}
	}
}

// updateMetrics updates node metrics
func (n *P2PNode) updateMetrics() {
	// Update uptime
	n.metrics.Uptime = time.Since(n.metrics.StartTime)
	
	// Update peer count
	n.metrics.ConnectedPeers = n.host.GetPeerCount()
	
	// Update last activity
	n.metrics.LastActivity = time.Now()
	
	// Aggregate metrics from components
	if n.discoveryEngine != nil {
		discoveryMetrics := n.discoveryEngine.GetMetrics()
		n.metrics.PeersDiscovered = discoveryMetrics.PeersFound
		n.metrics.DiscoveryErrors = discoveryMetrics.DiscoveryErrors
	}
	
	if n.securityManager != nil {
		securityMetrics := n.securityManager.GetMetrics()
		n.metrics.AuthAttempts = securityMetrics.AuthAttempts
		n.metrics.AuthSuccesses = securityMetrics.AuthAttempts - securityMetrics.AuthFailures
		n.metrics.AuthFailures = securityMetrics.AuthFailures
	}
	
	if n.resourceAdvertiser != nil {
		advertiserMetrics := n.resourceAdvertiser.GetMetrics()
		n.metrics.ResourcesAdvertised = advertiserMetrics.AdvertisementsSent
		n.metrics.ResourcesDiscovered = advertiserMetrics.AdvertisementsReceived
	}
	
	if n.contentRouter != nil {
		routerMetrics := n.contentRouter.GetMetrics()
		n.metrics.ContentPublished = routerMetrics.ContentPublished
		n.metrics.ContentRequests = routerMetrics.ContentRequests
		n.metrics.ContentProvided = routerMetrics.ContentProvided
	}
}

// updateResourceMetrics updates resource metrics
func (n *P2PNode) updateResourceMetrics() {
	if n.resourceMetrics == nil {
		n.resourceMetrics = &ResourceMetrics{
			LastUpdated: time.Now(),
		}
	}
	
	// Get system resource usage
	cpuUsage, err := n.getCPUUsage()
	if err != nil {
		cpuUsage = 0.0
	}
	
	memoryUsage, totalMemory, err := n.getMemoryUsage()
	if err != nil {
		memoryUsage = 0.0
		totalMemory = 0
	}
	
	diskUsage, totalDisk, err := n.getDiskUsage()
	if err != nil {
		diskUsage = 0.0
		totalDisk = 0
	}
	
	networkBandwidth := n.getNetworkBandwidth()
	
	// Update metrics
	n.resourceMetrics.CPUUsage = cpuUsage
	n.resourceMetrics.MemoryUsage = memoryUsage
	n.resourceMetrics.MemoryTotal = totalMemory
	n.resourceMetrics.DiskUsage = diskUsage
	n.resourceMetrics.DiskTotal = totalDisk
	n.resourceMetrics.NetworkBandwidth = networkBandwidth
	n.resourceMetrics.LastUpdated = time.Now()
	
	// Update advertiser
	if n.resourceAdvertiser != nil {
		n.resourceAdvertiser.SetResourceMetrics(n.resourceMetrics)
	}
}

// Status and information

// GetStatus returns node status
func (n *P2PNode) GetStatus() *NodeStatus {
	n.startedMux.RLock()
	defer n.startedMux.RUnlock()
	
	return &NodeStatus{
		ID:              n.host.ID(),
		Started:         n.started,
		Uptime:          n.metrics.Uptime,
		ConnectedPeers:  n.metrics.ConnectedPeers,
		ListenAddresses: n.host.Addrs(),
		Capabilities:    n.capabilities,
		ResourceMetrics: n.resourceMetrics,
		LastActivity:    n.metrics.LastActivity,
	}
}

// GetMetrics returns node metrics
func (n *P2PNode) GetMetrics() *NodeMetrics {
	return n.metrics
}

// GetConfig returns node configuration
func (n *P2PNode) GetConfig() *config.NodeConfig {
	return n.config
}

// GetHost returns the underlying host
func (n *P2PNode) GetHost() host.Host {
	return n.host
}

// NodeStatus represents node status
type NodeStatus struct {
	ID              peer.ID
	Started         bool
	Uptime          time.Duration
	ConnectedPeers  int
	ListenAddresses []multiaddr.Multiaddr
	Capabilities    *NodeCapabilities
	ResourceMetrics *ResourceMetrics
	LastActivity    time.Time
}

// String returns string representation of node status
func (s *NodeStatus) String() string {
	return fmt.Sprintf("Node %s: Started=%t, Uptime=%v, Peers=%d, Addrs=%v",
		s.ID, s.Started, s.Uptime, s.ConnectedPeers, s.ListenAddresses)
}

// System resource monitoring methods

// getCPUUsage returns current CPU usage percentage
func (n *P2PNode) getCPUUsage() (float64, error) {
	// Implementation depends on platform
	// This is a simplified cross-platform approach
	var usage float64
	
	// Try to read /proc/stat on Linux
	if data, err := os.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "cpu ") {
			fields := strings.Fields(lines[0])
			if len(fields) >= 8 {
				user, _ := strconv.ParseFloat(fields[1], 64)
				nice, _ := strconv.ParseFloat(fields[2], 64)
				system, _ := strconv.ParseFloat(fields[3], 64)
				idle, _ := strconv.ParseFloat(fields[4], 64)
				
				total := user + nice + system + idle
				if total > 0 {
					usage = ((total - idle) / total) * 100
				}
			}
		}
	} else {
		// Fallback: use runtime statistics
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		// Approximate CPU usage based on GC activity
		usage = float64(m.NumGC % 100)
	}
	
	return usage, nil
}

// getMemoryUsage returns current memory usage and total memory
func (n *P2PNode) getMemoryUsage() (float64, int64, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Get process memory usage
	processMemory := int64(m.Alloc)
	
	// Try to get system memory info
	var totalMemory int64
	var usage float64
	
	// Try to read /proc/meminfo on Linux
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(data), "\n")
		var memTotal, memAvailable int64
		
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						memTotal = val * 1024 // Convert KB to bytes
					}
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						memAvailable = val * 1024 // Convert KB to bytes
					}
				}
			}
		}
		
		if memTotal > 0 {
			totalMemory = memTotal
			if memAvailable > 0 {
				usage = float64(memTotal-memAvailable) / float64(memTotal) * 100
			}
		}
	}
	
	// Fallback to process memory
	if totalMemory == 0 {
		totalMemory = processMemory * 10 // Rough estimate
		usage = 10.0 // Rough estimate
	}
	
	return usage, totalMemory, nil
}

// getDiskUsage returns current disk usage and total disk space
func (n *P2PNode) getDiskUsage() (float64, int64, error) {
	// Get current working directory disk usage
	pwd, err := os.Getwd()
	if err != nil {
		return 0, 0, err
	}
	
	// Try to get disk usage for the current directory
	// This is a cross-platform approach using file stat
	var totalSpace, usedSpace int64
	
	if info, err := os.Stat(pwd); err == nil {
		// This is a rough approximation
		// In a real implementation, you'd use platform-specific syscalls
		totalSpace = 1024 * 1024 * 1024 * 100 // Assume 100GB
		usedSpace = info.Size() * 1000         // Rough estimate
	}
	
	usage := float64(usedSpace) / float64(totalSpace) * 100
	if usage > 100 {
		usage = 100
	}
	
	return usage, totalSpace, nil
}

// getNetworkBandwidth returns current network bandwidth estimate
func (n *P2PNode) getNetworkBandwidth() int64 {
	// Simple bandwidth estimation based on peer connections
	peerCount := n.GetPeerCount()
	
	// Estimate bandwidth based on number of peers
	// This is a rough approximation
	baseBandwidth := int64(1024 * 1024) // 1 MB/s base
	peerBandwidth := int64(peerCount * 100 * 1024) // 100 KB/s per peer
	
	return baseBandwidth + peerBandwidth
}