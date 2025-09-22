package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/ollama-distributed/pkg/p2p/host"
)

// MessageRouter handles routing of messages between peers in the distributed system
type MessageRouter struct {
	config *RouterConfig

	// Local peer identity
	localPeerID peer.ID

	// Protocol handlers
	handlers   map[protocol.ID]ProtocolHandler
	handlersMu sync.RWMutex

	// Message queues
	outboundQueue     *MessageQueue
	inboundQueue      *MessageQueue
	tensorStreamQueue *TensorStreamQueue

	// Connection management
	connections   map[peer.ID]*PeerConnection
	connectionsMu sync.RWMutex

	// Message tracking
	pendingMessages map[string]*PendingMessage
	pendingMu       sync.RWMutex

	// Routing table
	routingTable *RoutingTable

	// Metrics and monitoring
	metrics *RouterMetrics

	// Tensor streaming support
	tensorMetrics    *TensorStreamMetrics
	bandwidthManager *host.BandwidthManager

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TensorStreamQueue extends MessageQueue with tensor streaming optimizations
type TensorStreamQueue struct {
	*MessageQueue
	maxTensorSize    int64
	compressionRatio float64
	priorityLevels   map[string]int
	streamingMetrics *TensorStreamMetrics
}

// TensorStreamMetrics tracks tensor streaming performance in the router
type TensorStreamMetrics struct {
	TotalTensorMessages     int64
	TensorBytesRouted       int64
	AverageCompressionRatio float64
	StreamingThroughput     float64
	QueueBackpressure       int
	RoutingLatency          time.Duration
	mutex                   sync.RWMutex
}

// BandwidthManager is now unified to use host.BandwidthManager

// RouterConfig configures the message router
type RouterConfig struct {
	// Queue settings
	MaxQueueSize int
	QueueTimeout time.Duration

	// Connection settings
	MaxConnections    int
	ConnectionTimeout time.Duration
	KeepAliveInterval time.Duration

	// Message settings
	MaxMessageSize int
	MessageTimeout time.Duration
	RetryAttempts  int
	RetryBackoff   time.Duration

	// Routing settings
	RoutingTableSize     int
	RouteRefreshInterval time.Duration

	// Performance settings
	WorkerCount       int
	BufferSize        int
	EnableCompression bool

	// Reliability settings
	EnableAcknowledgments    bool
	AckTimeout               time.Duration
	EnableDuplicateDetection bool
}

// Message represents a message in the distributed system
type Message struct {
	// Message identification
	ID       string      `json:"id"`
	Type     MessageType `json:"type"`
	Protocol protocol.ID `json:"protocol"`

	// Routing information
	Source      peer.ID   `json:"source"`
	Destination peer.ID   `json:"destination"`
	Route       []peer.ID `json:"route,omitempty"`

	// Message content
	Payload []byte            `json:"payload"`
	Headers map[string]string `json:"headers"`

	// Message metadata
	Timestamp time.Time       `json:"timestamp"`
	TTL       time.Duration   `json:"ttl"`
	Priority  MessagePriority `json:"priority"`

	// Reliability
	RequiresAck bool `json:"requires_ack"`
	RetryCount  int  `json:"retry_count"`

	// Compression
	Compressed   bool `json:"compressed"`
	OriginalSize int  `json:"original_size,omitempty"`
}

// PendingMessage tracks messages awaiting acknowledgment
type PendingMessage struct {
	Message      *Message
	SentAt       time.Time
	RetryCount   int
	AckReceived  bool
	AckChan      chan bool
	TimeoutTimer *time.Timer
}

// PeerConnection represents a connection to a peer
type PeerConnection struct {
	PeerID       peer.ID
	Protocol     protocol.ID
	Connected    bool
	ConnectedAt  time.Time
	LastActivity time.Time

	// Connection state
	SendQueue    chan *Message
	ReceiveQueue chan *Message

	// Statistics
	MessagesSent     int64
	MessagesReceived int64
	BytesSent        int64
	BytesReceived    int64

	// Reliability
	LastPing time.Time
	RTT      time.Duration

	mu sync.RWMutex
}

// RoutingTable manages routing information for the network
type RoutingTable struct {
	routes   map[peer.ID]*RouteEntry
	routesMu sync.RWMutex

	// Network topology
	neighbors   map[peer.ID]bool
	neighborsMu sync.RWMutex

	// Route discovery
	discoveryQueue chan peer.ID
	lastDiscovery  map[peer.ID]time.Time
	discoveryMu    sync.RWMutex
}

// RouteEntry represents a route to a peer
type RouteEntry struct {
	Destination peer.ID
	NextHop     peer.ID
	HopCount    int
	Cost        int
	LastUpdated time.Time
	Valid       bool
}

// RouterMetrics tracks router performance
type RouterMetrics struct {
	// Message metrics
	TotalMessages    int64
	MessagesSent     int64
	MessagesReceived int64
	MessagesDropped  int64
	MessagesRetried  int64

	// Queue metrics
	OutboundQueueSize int64
	InboundQueueSize  int64
	QueueOverflows    int64

	// Connection metrics
	ActiveConnections  int64
	ConnectionFailures int64
	ConnectionTimeouts int64

	// Routing metrics
	RoutingTableSize int64
	RouteDiscoveries int64
	RoutingFailures  int64

	// Performance metrics
	AverageLatency    time.Duration
	MessageThroughput float64

	// Last updated
	LastUpdated time.Time
	mu          sync.RWMutex
}

// MessageQueue implements a thread-safe message queue
type MessageQueue struct {
	messages  chan *Message
	maxSize   int
	timeout   time.Duration
	overflows int64
	mu        sync.RWMutex
}

// Enums and constants
type MessageType string

const (
	MessageTypeConsensus MessageType = "consensus"
	MessageTypeScheduler MessageType = "scheduler"
	MessageTypeModel     MessageType = "model"
	MessageTypeDiscovery MessageType = "discovery"
	MessageTypeHealth    MessageType = "health"
	MessageTypeData      MessageType = "data"
	MessageTypeControl   MessageType = "control"
	MessageTypeAck       MessageType = "ack"
	MessageTypeTensorStart MessageType = "tensor_start"
	MessageTypeTensorChunk MessageType = "tensor_chunk"
	MessageTypeTensorComplete MessageType = "tensor_complete"
	MessageTypeTensorStream MessageType = "tensor_stream"
)

type MessagePriority int

const (
	PriorityLow      MessagePriority = 1
	PriorityNormal   MessagePriority = 5
	PriorityHigh     MessagePriority = 8
	PriorityCritical MessagePriority = 10
)

// Protocol IDs
const (
	ProtocolConsensus protocol.ID = "/ollama-distributed/consensus/1.0.0"
	ProtocolScheduler protocol.ID = "/ollama-distributed/scheduler/1.0.0"
	ProtocolModel     protocol.ID = "/ollama-distributed/model/1.0.0"
	ProtocolDiscovery protocol.ID = "/ollama-distributed/discovery/1.0.0"
	ProtocolHealth    protocol.ID = "/ollama-distributed/health/1.0.0"
	ProtocolData      protocol.ID = "/ollama-distributed/data/1.0.0"
	ProtocolTensorStream protocol.ID = "/ollama-distributed/tensor-stream/1.0.0"
)

// Interfaces
type ProtocolHandler interface {
	HandleMessage(ctx context.Context, msg *Message) error
	GetProtocol() protocol.ID
	GetMessageTypes() []MessageType
}

type MessageSerializer interface {
	Serialize(msg *Message) ([]byte, error)
	Deserialize(data []byte) (*Message, error)
}

// NewMessageRouter creates a new message router
func NewMessageRouter(config *RouterConfig, localPeerID peer.ID) *MessageRouter {
	if config == nil {
		config = &RouterConfig{
			MaxQueueSize:             10000,
			QueueTimeout:             30 * time.Second,
			MaxConnections:           1000,
			ConnectionTimeout:        30 * time.Second,
			KeepAliveInterval:        30 * time.Second,
			MaxMessageSize:           10 * 1024 * 1024, // 10MB
			MessageTimeout:           30 * time.Second,
			RetryAttempts:            3,
			RetryBackoff:             time.Second,
			RoutingTableSize:         10000,
			RouteRefreshInterval:     5 * time.Minute,
			WorkerCount:              10,
			BufferSize:               1024,
			EnableCompression:        true,
			EnableAcknowledgments:    true,
			AckTimeout:               10 * time.Second,
			EnableDuplicateDetection: true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	router := &MessageRouter{
		config:          config,
		localPeerID:     localPeerID,
		handlers:        make(map[protocol.ID]ProtocolHandler),
		connections:     make(map[peer.ID]*PeerConnection),
		pendingMessages: make(map[string]*PendingMessage),
		metrics: &RouterMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize message queues
	router.outboundQueue = NewMessageQueue(config.MaxQueueSize, config.QueueTimeout)
	router.inboundQueue = NewMessageQueue(config.MaxQueueSize, config.QueueTimeout)

	// Initialize tensor stream queue with larger buffer for high-throughput streaming
	tensorQueueSize := config.MaxQueueSize * 2 // Larger buffer for tensor data
	tensorBaseQueue := NewMessageQueue(tensorQueueSize, config.QueueTimeout)
	router.tensorStreamQueue = NewTensorStreamQueue(tensorBaseQueue, 1024*1024*1024) // 1GB max tensor size

	// Initialize tensor metrics
	router.tensorMetrics = &TensorStreamMetrics{}

	// Initialize bandwidth manager with proper configuration
	bandwidthConfig := &host.BandwidthConfig{
		GlobalLimit:             1024 * 1024 * 1024, // 1GB/s total
		DefaultPeerLimit:        100 * 1024 * 1024,  // 100MB/s per peer
		TensorStreamingPriority: true,
		TensorStreamingLimit:    512 * 1024 * 1024, // 512MB/s for tensor streaming
		WindowSize:              time.Second,
		BurstSize:               50 * 1024 * 1024, // 50MB burst
		UpdateInterval:          time.Second,
		ProtocolLimits:          make(map[string]int64),
		PriorityProtocols:       []string{"/ollama/tensor-stream/1.0.0"},
		PriorityMultiplier:      2.0,
	}
	router.bandwidthManager = host.NewBandwidthManager(bandwidthConfig)

	// Initialize routing table
	router.routingTable = &RoutingTable{
		routes:         make(map[peer.ID]*RouteEntry),
		neighbors:      make(map[peer.ID]bool),
		discoveryQueue: make(chan peer.ID, 1000),
		lastDiscovery:  make(map[peer.ID]time.Time),
	}

	return router
}

// RegisterTensorStreamHandler registers the tensor stream handler with the router
func (mr *MessageRouter) RegisterTensorStreamHandler() {
	handler := NewTensorStreamHandler(mr.localPeerID)
	mr.RegisterHandler(handler)
}

// Start starts the message router
func (mr *MessageRouter) Start() error {
	// Start worker goroutines
	for i := 0; i < mr.config.WorkerCount; i++ {
		mr.wg.Add(1)
		go mr.outboundWorker()

		mr.wg.Add(1)
		go mr.inboundWorker()
	}

	// Start connection manager
	mr.wg.Add(1)
	go mr.connectionManager()

	// Start routing table manager
	mr.wg.Add(1)
	go mr.routingTableManager()

	// Start metrics collector
	mr.wg.Add(1)
	go mr.metricsCollector()

	// Start acknowledgment handler
	if mr.config.EnableAcknowledgments {
		mr.wg.Add(1)
		go mr.acknowledgmentHandler()
	}

	// Start tensor stream processor for high-throughput processing
	mr.wg.Add(1)
	go mr.tensorStreamProcessor()

	// Start bandwidth monitor
	mr.wg.Add(1)
	go mr.bandwidthMonitor()

	return nil
}

// Stop stops the message router
func (mr *MessageRouter) Stop() error {
	mr.cancel()
	mr.wg.Wait()

	// Close queues
	mr.outboundQueue.Close()
	mr.inboundQueue.Close()

	// Close connections
	mr.connectionsMu.Lock()
	for _, conn := range mr.connections {
		close(conn.SendQueue)
		close(conn.ReceiveQueue)
	}
	mr.connectionsMu.Unlock()

	return nil
}

// RegisterHandler registers a protocol handler
func (mr *MessageRouter) RegisterHandler(handler ProtocolHandler) {
	mr.handlersMu.Lock()
	defer mr.handlersMu.Unlock()
	mr.handlers[handler.GetProtocol()] = handler
}

// ensureHeaders ensures that message headers are initialized
func ensureHeaders(msg *Message) {
	if msg.Headers == nil {
		msg.Headers = make(map[string]string)
	}
}

// SendMessage sends a message to a peer
func (mr *MessageRouter) SendMessage(msg *Message) error {
	// Ensure headers are initialized
	ensureHeaders(msg)

	// Validate message
	if err := mr.validateMessage(msg); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Set message metadata
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	if msg.TTL == 0 {
		msg.TTL = mr.config.MessageTimeout
	}

	// Compress if enabled and beneficial
	if mr.config.EnableCompression && len(msg.Payload) > 1024 {
		if err := mr.compressMessage(msg); err != nil {
			return fmt.Errorf("failed to compress message: %w", err)
		}
	}

	// Add to outbound queue
	select {
	case mr.outboundQueue.messages <- msg:
		mr.metrics.mu.Lock()
		mr.metrics.TotalMessages++
		mr.metrics.mu.Unlock()
		return nil
	case <-time.After(mr.config.QueueTimeout):
		mr.metrics.mu.Lock()
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
		return fmt.Errorf("outbound queue timeout")
	}
}

// BroadcastMessage broadcasts a message to all connected peers
func (mr *MessageRouter) BroadcastMessage(msg *Message) error {
	// Ensure headers are initialized
	ensureHeaders(msg)

	mr.connectionsMu.RLock()
	peers := make([]peer.ID, 0, len(mr.connections))
	for peerID := range mr.connections {
		peers = append(peers, peerID)
	}
	mr.connectionsMu.RUnlock()

	for _, peerID := range peers {
		msgCopy := *msg
		msgCopy.Destination = peerID
		msgCopy.ID = generateMessageID()

		if err := mr.SendMessage(&msgCopy); err != nil {
			// Log error but continue with other peers
			continue
		}
	}

	return nil
}

// validateMessage validates a message before sending
func (mr *MessageRouter) validateMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}
	if msg.Destination == "" {
		return fmt.Errorf("destination is required")
	}
	if len(msg.Payload) > mr.config.MaxMessageSize {
		return fmt.Errorf("message too large: %d > %d", len(msg.Payload), mr.config.MaxMessageSize)
	}
	return nil
}

// compressMessage compresses a message payload
func (mr *MessageRouter) compressMessage(msg *Message) error {
	// Implementation would compress the payload
	// For now, this is a placeholder
	msg.Compressed = true
	msg.OriginalSize = len(msg.Payload)
	return nil
}

// decompressMessage decompresses a message payload
func (mr *MessageRouter) decompressMessage(msg *Message) error {
	// Implementation would decompress the payload
	// For now, this is a placeholder
	msg.Compressed = false
	return nil
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(maxSize int, timeout time.Duration) *MessageQueue {
	return &MessageQueue{
		messages: make(chan *Message, maxSize),
		maxSize:  maxSize,
		timeout:  timeout,
	}
}

// Enqueue adds a message to the queue
func (mq *MessageQueue) Enqueue(msg *Message) error {
	select {
	case mq.messages <- msg:
		return nil
	default:
		mq.mu.Lock()
		mq.overflows++
		mq.mu.Unlock()
		return fmt.Errorf("queue full")
	}
}

// Dequeue removes a message from the queue
func (mq *MessageQueue) Dequeue() (*Message, error) {
	select {
	case msg := <-mq.messages:
		return msg, nil
	case <-time.After(mq.timeout):
		return nil, fmt.Errorf("queue timeout")
	}
}

// Size returns the current queue size
func (mq *MessageQueue) Size() int {
	return len(mq.messages)
}

// Close closes the queue
func (mq *MessageQueue) Close() {
	close(mq.messages)
}

// Worker functions

// outboundWorker processes outbound messages
func (mr *MessageRouter) outboundWorker() {
	defer mr.wg.Done()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case msg := <-mr.outboundQueue.messages:
			mr.processOutboundMessage(msg)
		}
	}
}

// inboundWorker processes inbound messages
func (mr *MessageRouter) inboundWorker() {
	defer mr.wg.Done()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case msg := <-mr.inboundQueue.messages:
			mr.processInboundMessage(msg)
		}
	}
}

// processOutboundMessage processes an outbound message
func (mr *MessageRouter) processOutboundMessage(msg *Message) {
	// Find route to destination
	route := mr.findRoute(msg.Destination)
	if route == nil {
		mr.metrics.mu.Lock()
		mr.metrics.RoutingFailures++
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
		return
	}

	// Get connection to next hop
	nextHop := route.NextHop
	if route.HopCount == 1 {
		nextHop = msg.Destination
	}

	conn := mr.getConnection(nextHop)
	if conn == nil {
		mr.metrics.mu.Lock()
		mr.metrics.ConnectionFailures++
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
		return
	}

	// Send message
	select {
	case conn.SendQueue <- msg:
		mr.metrics.mu.Lock()
		mr.metrics.MessagesSent++
		mr.metrics.mu.Unlock()

		// Track pending message if acknowledgment required
		if msg.RequiresAck {
			mr.trackPendingMessage(msg)
		}

	case <-time.After(mr.config.MessageTimeout):
		mr.metrics.mu.Lock()
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
	}
}

// processInboundMessage processes an inbound message
func (mr *MessageRouter) processInboundMessage(msg *Message) {
	// Send acknowledgment if required
	if msg.RequiresAck {
		mr.sendAcknowledgment(msg)
	}

	// Check if message is for this node
	if msg.Destination == mr.getLocalPeerID() {
		mr.handleLocalMessage(msg)
		return
	}

	// Forward message if TTL allows
	if time.Since(msg.Timestamp) < msg.TTL {
		mr.forwardMessage(msg)
	} else {
		mr.metrics.mu.Lock()
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
	}
}

// handleLocalMessage handles a message destined for this node
func (mr *MessageRouter) handleLocalMessage(msg *Message) {
	mr.handlersMu.RLock()
	handler, exists := mr.handlers[msg.Protocol]
	mr.handlersMu.RUnlock()

	if !exists {
		mr.metrics.mu.Lock()
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
		return
	}

	// Handle message in goroutine to avoid blocking
	go func() {
		ctx, cancel := context.WithTimeout(mr.ctx, mr.config.MessageTimeout)
		defer cancel()

		if err := handler.HandleMessage(ctx, msg); err != nil {
			// Log error but don't fail
		}

		mr.metrics.mu.Lock()
		mr.metrics.MessagesReceived++
		mr.metrics.mu.Unlock()
	}()
}

// forwardMessage forwards a message to its destination
func (mr *MessageRouter) forwardMessage(msg *Message) {
	// Add this node to the route to prevent loops
	for _, hop := range msg.Route {
		if hop == mr.getLocalPeerID() {
			// Loop detected, drop message
			mr.metrics.mu.Lock()
			mr.metrics.MessagesDropped++
			mr.metrics.mu.Unlock()
			return
		}
	}

	msg.Route = append(msg.Route, mr.getLocalPeerID())

	// Forward the message
	mr.processOutboundMessage(msg)
}

// connectionManager manages peer connections
func (mr *MessageRouter) connectionManager() {
	defer mr.wg.Done()

	ticker := time.NewTicker(mr.config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.maintainConnections()
		}
	}
}

// maintainConnections maintains active connections
func (mr *MessageRouter) maintainConnections() {
	mr.connectionsMu.Lock()
	defer mr.connectionsMu.Unlock()

	now := time.Now()
	for peerID, conn := range mr.connections {
		// Check if connection is stale
		if now.Sub(conn.LastActivity) > mr.config.KeepAliveInterval*2 {
			// Send ping message
			mr.sendPing(peerID)
		}

		// Remove dead connections
		if now.Sub(conn.LastActivity) > mr.config.ConnectionTimeout {
			delete(mr.connections, peerID)
			close(conn.SendQueue)
			close(conn.ReceiveQueue)
		}
	}
}

// routingTableManager manages the routing table
func (mr *MessageRouter) routingTableManager() {
	defer mr.wg.Done()

	ticker := time.NewTicker(mr.config.RouteRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.refreshRoutingTable()
		case peerID := <-mr.routingTable.discoveryQueue:
			mr.discoverRoute(peerID)
		}
	}
}

// refreshRoutingTable refreshes the routing table
func (mr *MessageRouter) refreshRoutingTable() {
	mr.routingTable.routesMu.Lock()
	defer mr.routingTable.routesMu.Unlock()

	now := time.Now()
	for peerID, route := range mr.routingTable.routes {
		// Mark old routes as invalid
		if now.Sub(route.LastUpdated) > mr.config.RouteRefreshInterval*2 {
			route.Valid = false
		}

		// Remove very old routes
		if now.Sub(route.LastUpdated) > mr.config.RouteRefreshInterval*5 {
			delete(mr.routingTable.routes, peerID)
		}
	}
}

// metricsCollector collects and updates metrics
func (mr *MessageRouter) metricsCollector() {
	defer mr.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.updateMetrics()
		}
	}
}

// updateMetrics updates router metrics
func (mr *MessageRouter) updateMetrics() {
	mr.metrics.mu.Lock()
	defer mr.metrics.mu.Unlock()

	// Update queue sizes
	mr.metrics.OutboundQueueSize = int64(mr.outboundQueue.Size())
	mr.metrics.InboundQueueSize = int64(mr.inboundQueue.Size())

	// Update connection count
	mr.connectionsMu.RLock()
	mr.metrics.ActiveConnections = int64(len(mr.connections))
	mr.connectionsMu.RUnlock()

	// Update routing table size
	mr.routingTable.routesMu.RLock()
	mr.metrics.RoutingTableSize = int64(len(mr.routingTable.routes))
	mr.routingTable.routesMu.RUnlock()

	mr.metrics.LastUpdated = time.Now()
}

// acknowledgmentHandler handles message acknowledgments
func (mr *MessageRouter) acknowledgmentHandler() {
	defer mr.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.checkPendingAcknowledgments()
		}
	}
}

// Helper functions

// findRoute finds a route to the destination peer
func (mr *MessageRouter) findRoute(destination peer.ID) *RouteEntry {
	mr.routingTable.routesMu.RLock()
	defer mr.routingTable.routesMu.RUnlock()

	if route, exists := mr.routingTable.routes[destination]; exists && route.Valid {
		return route
	}

	// Trigger route discovery
	select {
	case mr.routingTable.discoveryQueue <- destination:
	default:
		// Discovery queue full
	}

	return nil
}

// getConnection gets a connection to a peer
func (mr *MessageRouter) getConnection(peerID peer.ID) *PeerConnection {
	mr.connectionsMu.RLock()
	defer mr.connectionsMu.RUnlock()

	if conn, exists := mr.connections[peerID]; exists && conn.Connected {
		return conn
	}

	return nil
}

// getLocalPeerID returns the local peer ID
func (mr *MessageRouter) getLocalPeerID() peer.ID {
	return mr.localPeerID
}

// SetLocalPeerID sets the local peer ID (for cases where it needs to be updated)
func (mr *MessageRouter) SetLocalPeerID(peerID peer.ID) {
	mr.localPeerID = peerID
}

// trackPendingMessage tracks a message awaiting acknowledgment
func (mr *MessageRouter) trackPendingMessage(msg *Message) {
	mr.pendingMu.Lock()
	defer mr.pendingMu.Unlock()

	pending := &PendingMessage{
		Message:     msg,
		SentAt:      time.Now(),
		RetryCount:  0,
		AckReceived: false,
		AckChan:     make(chan bool, 1),
	}

	// Set timeout timer
	pending.TimeoutTimer = time.AfterFunc(mr.config.AckTimeout, func() {
		mr.handleAckTimeout(msg.ID)
	})

	mr.pendingMessages[msg.ID] = pending
}

// sendAcknowledgment sends an acknowledgment for a message
func (mr *MessageRouter) sendAcknowledgment(msg *Message) {
	ack := &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeAck,
		Protocol:    msg.Protocol,
		Source:      mr.getLocalPeerID(),
		Destination: msg.Source,
		Headers:     map[string]string{"ack_for": msg.ID},
		Timestamp:   time.Now(),
		TTL:         mr.config.MessageTimeout,
	}

	mr.SendMessage(ack)
}

// sendPing sends a ping message to a peer
func (mr *MessageRouter) sendPing(peerID peer.ID) {
	ping := &Message{
		ID:          generateMessageID(),
		Type:        MessageTypeHealth,
		Protocol:    ProtocolHealth,
		Source:      mr.getLocalPeerID(),
		Destination: peerID,
		Headers:     map[string]string{"type": "ping"},
		Timestamp:   time.Now(),
		TTL:         mr.config.MessageTimeout,
	}

	mr.SendMessage(ping)
}

// discoverRoute discovers a route to a peer
func (mr *MessageRouter) discoverRoute(peerID peer.ID) {
	// Implementation would perform route discovery
	// For now, this is a placeholder
}

// checkPendingAcknowledgments checks for timed out acknowledgments
func (mr *MessageRouter) checkPendingAcknowledgments() {
	mr.pendingMu.Lock()
	defer mr.pendingMu.Unlock()

	now := time.Now()
	for msgID, pending := range mr.pendingMessages {
		if !pending.AckReceived && now.Sub(pending.SentAt) > mr.config.AckTimeout {
			mr.handleAckTimeout(msgID)
		}
	}
}

// handleAckTimeout handles acknowledgment timeout
func (mr *MessageRouter) handleAckTimeout(msgID string) {
	mr.pendingMu.Lock()
	defer mr.pendingMu.Unlock()

	pending, exists := mr.pendingMessages[msgID]
	if !exists || pending.AckReceived {
		return
	}

	// Retry if attempts remaining
	if pending.RetryCount < mr.config.RetryAttempts {
		pending.RetryCount++
		pending.SentAt = time.Now()

		// Resend message
		go mr.SendMessage(pending.Message)

		mr.metrics.mu.Lock()
		mr.metrics.MessagesRetried++
		mr.metrics.mu.Unlock()
	} else {
		// Give up
		delete(mr.pendingMessages, msgID)
		close(pending.AckChan)

		mr.metrics.mu.Lock()
		mr.metrics.MessagesDropped++
		mr.metrics.mu.Unlock()
	}
}

// GetMetrics returns router metrics
func (mr *MessageRouter) GetMetrics() *RouterMetrics {
	mr.metrics.mu.RLock()
	defer mr.metrics.mu.RUnlock()

	// Create a copy without the mutex
	return &RouterMetrics{
		TotalMessages:      mr.metrics.TotalMessages,
		MessagesSent:       mr.metrics.MessagesSent,
		MessagesReceived:   mr.metrics.MessagesReceived,
		MessagesDropped:    mr.metrics.MessagesDropped,
		MessagesRetried:    mr.metrics.MessagesRetried,
		OutboundQueueSize:  mr.metrics.OutboundQueueSize,
		InboundQueueSize:   mr.metrics.InboundQueueSize,
		QueueOverflows:     mr.metrics.QueueOverflows,
		ActiveConnections:  mr.metrics.ActiveConnections,
		ConnectionFailures: mr.metrics.ConnectionFailures,
		ConnectionTimeouts: mr.metrics.ConnectionTimeouts,
		RoutingTableSize:   mr.metrics.RoutingTableSize,
		RouteDiscoveries:   mr.metrics.RouteDiscoveries,
		RoutingFailures:    mr.metrics.RoutingFailures,
		AverageLatency:     mr.metrics.AverageLatency,
		MessageThroughput:  mr.metrics.MessageThroughput,
		LastUpdated:        mr.metrics.LastUpdated,
	}
}

// NewTensorStreamQueue creates a new tensor stream queue
func NewTensorStreamQueue(baseQueue *MessageQueue, maxTensorSize int64) *TensorStreamQueue {
	return &TensorStreamQueue{
		MessageQueue:     baseQueue,
		maxTensorSize:    maxTensorSize,
		compressionRatio: 0.7, // Default compression ratio
		priorityLevels:   make(map[string]int),
		streamingMetrics: &TensorStreamMetrics{},
	}
}

// RouteTensorMessage routes tensor streaming messages with optimizations
func (mr *MessageRouter) RouteTensorMessage(msg *Message) error {
	// Ensure headers are initialized
	ensureHeaders(msg)

	// Check if this is a tensor streaming message
	if !mr.isTensorStreamingMessage(msg) {
		return mr.SendMessage(msg)
	}

	// Apply tensor-specific routing optimizations
	mr.optimizeTensorRouting(msg)

	// Update tensor metrics
	mr.updateTensorMetrics(msg)

	// Route through tensor stream queue
	return mr.tensorStreamQueue.Enqueue(msg)
}

// isTensorStreamingMessage checks if a message is tensor streaming related
func (mr *MessageRouter) isTensorStreamingMessage(msg *Message) bool {
	return msg.Protocol == ProtocolTensorStream ||
		msg.Type == MessageTypeTensorStart ||
		msg.Type == MessageTypeTensorChunk ||
		msg.Type == MessageTypeTensorComplete
}

// optimizeTensorRouting applies tensor-specific routing optimizations
func (mr *MessageRouter) optimizeTensorRouting(msg *Message) {
	// Apply compression if beneficial
	if len(msg.Payload) > 1024 { // Only compress larger payloads
		mr.applyTensorCompression(msg)
	}

	// Set priority based on tensor type
	mr.setTensorPriority(msg)

	// Apply bandwidth management
	mr.manageTensorBandwidth(msg)
}

// applyTensorCompression applies compression to tensor messages
func (mr *MessageRouter) applyTensorCompression(msg *Message) {
	// Placeholder for compression logic
	// In real implementation, would use the tensor compressor
	msg.Headers["compressed"] = "true"
}

// setTensorPriority sets priority for tensor messages
func (mr *MessageRouter) setTensorPriority(msg *Message) {
	switch msg.Type {
	case "activation_start":
		msg.Headers["priority"] = "high"
	case "activation_chunk":
		msg.Headers["priority"] = "normal"
	case "activation_complete":
		msg.Headers["priority"] = "low"
	default:
		msg.Headers["priority"] = "normal"
	}
}

// manageTensorBandwidth manages bandwidth allocation for tensor messages
func (mr *MessageRouter) manageTensorBandwidth(msg *Message) {
	if mr.bandwidthManager != nil {
		// Use unified bandwidth manager with proper peer and protocol tracking
		peerID := msg.Destination // Use destination as peer ID
		bytesUsed := int64(len(msg.Payload))
		mr.bandwidthManager.RecordUsage(peerID, "/ollama/tensor-stream/1.0.0", bytesUsed, 0)
	}
}

// updateTensorMetrics updates tensor streaming metrics
func (mr *MessageRouter) updateTensorMetrics(msg *Message) {
	mr.tensorMetrics.mutex.Lock()
	defer mr.tensorMetrics.mutex.Unlock()

	mr.tensorMetrics.TotalTensorMessages++
	mr.tensorMetrics.TensorBytesRouted += int64(len(msg.Payload))

	// Update average compression ratio if compressed
	if msg.Headers["compressed"] == "true" {
		// Simplified compression ratio calculation
		mr.tensorMetrics.AverageCompressionRatio = 0.7
	}
}

// GetTensorMetrics returns current tensor streaming metrics
func (mr *MessageRouter) GetTensorMetrics() *TensorStreamMetrics {
	mr.tensorMetrics.mutex.RLock()
	defer mr.tensorMetrics.mutex.RUnlock()

	return &TensorStreamMetrics{
		TotalTensorMessages:     mr.tensorMetrics.TotalTensorMessages,
		TensorBytesRouted:       mr.tensorMetrics.TensorBytesRouted,
		AverageCompressionRatio: mr.tensorMetrics.AverageCompressionRatio,
		StreamingThroughput:     mr.tensorMetrics.StreamingThroughput,
		QueueBackpressure:       mr.tensorMetrics.QueueBackpressure,
		RoutingLatency:          mr.tensorMetrics.RoutingLatency,
	}
}

// AllocateBandwidth allocates bandwidth for a specific purpose
// BandwidthManager methods removed - now using unified host.BandwidthManager implementation

// tensorStreamProcessor processes tensor streaming messages with high throughput
func (mr *MessageRouter) tensorStreamProcessor() {
	defer mr.wg.Done()

	// Use multiple workers for parallel tensor processing
	numWorkers := mr.config.WorkerCount * 2 // More workers for tensor streams
	for i := 0; i < numWorkers; i++ {
		mr.wg.Add(1)
		go mr.tensorStreamWorker()
	}

	// Monitor tensor queue for backpressure
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.monitorTensorQueueBackpressure()
		}
	}
}

// tensorStreamWorker processes individual tensor messages
func (mr *MessageRouter) tensorStreamWorker() {
	defer mr.wg.Done()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case msg := <-mr.tensorStreamQueue.messages:
			mr.processTensorMessage(msg)
		}
	}
}

// processTensorMessage processes a single tensor streaming message
func (mr *MessageRouter) processTensorMessage(msg *Message) {
	startTime := time.Now()

	// Apply tensor-specific optimizations
	mr.optimizeTensorRouting(msg)

	// Route the message with priority handling
	if err := mr.routeTensorMessageWithPriority(msg); err != nil {
		mr.tensorMetrics.mutex.Lock()
		mr.tensorMetrics.QueueBackpressure++
		mr.tensorMetrics.mutex.Unlock()
		return
	}

	// Update latency metrics
	latency := time.Since(startTime)
	mr.tensorMetrics.mutex.Lock()
	mr.tensorMetrics.RoutingLatency = (mr.tensorMetrics.RoutingLatency + latency) / 2
	mr.tensorMetrics.mutex.Unlock()
}

// routeTensorMessageWithPriority routes tensor messages based on priority
func (mr *MessageRouter) routeTensorMessageWithPriority(msg *Message) error {
	priority := msg.Headers["priority"]

	switch priority {
	case "high":
		// High priority messages bypass normal queue
		return mr.sendDirectly(msg)
	case "normal", "low":
		// Normal and low priority use standard routing
		mr.processOutboundMessage(msg)
		return nil
	default:
		mr.processOutboundMessage(msg)
		return nil
	}
}

// sendDirectly sends a message directly without queuing
func (mr *MessageRouter) sendDirectly(msg *Message) error {
	// Find route to destination
	route := mr.findRoute(msg.Destination)
	if route == nil {
		return fmt.Errorf("no route found to %s", msg.Destination)
	}

	// Get connection and send immediately
	nextHop := route.NextHop
	if route.HopCount == 1 {
		nextHop = msg.Destination
	}

	conn := mr.getConnection(nextHop)
	if conn == nil {
		return fmt.Errorf("no connection to %s", nextHop)
	}

	// Send with timeout
	select {
	case conn.SendQueue <- msg:
		return nil
	case <-time.After(mr.config.MessageTimeout):
		return fmt.Errorf("send timeout")
	}
}

// monitorTensorQueueBackpressure monitors tensor queue for backpressure
func (mr *MessageRouter) monitorTensorQueueBackpressure() {
	queueSize := mr.tensorStreamQueue.Size()
	maxSize := mr.tensorStreamQueue.maxSize

	backpressureRatio := float64(queueSize) / float64(maxSize)

	mr.tensorMetrics.mutex.Lock()
	mr.tensorMetrics.QueueBackpressure = int(backpressureRatio * 100)

	// Update throughput calculation
	currentTime := time.Now().Unix()
	if currentTime > 0 {
		mr.tensorMetrics.StreamingThroughput = float64(mr.tensorMetrics.TensorBytesRouted) / float64(currentTime)
	}
	mr.tensorMetrics.mutex.Unlock()

	// Apply backpressure management if needed
	if backpressureRatio > 0.8 {
		mr.applyBackpressureControl()
	}
}

// applyBackpressureControl applies backpressure control mechanisms
func (mr *MessageRouter) applyBackpressureControl() {
	// Temporarily reduce tensor bandwidth allocation
	mr.bandwidthManager.mutex.Lock()
	mr.bandwidthManager.tensorBandwidth = mr.bandwidthManager.tensorBandwidth * 8 / 10 // Reduce by 20%
	mr.bandwidthManager.mutex.Unlock()

	// Could also implement:
	// - Drop low priority messages
	// - Increase compression ratios
	// - Throttle message acceptance
}

// bandwidthMonitor monitors and manages bandwidth allocation
func (mr *MessageRouter) bandwidthMonitor() {
	defer mr.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mr.ctx.Done():
			return
		case <-ticker.C:
			mr.manageBandwidthAllocation()
		}
	}
}

// manageBandwidthAllocation manages dynamic bandwidth allocation
func (mr *MessageRouter) manageBandwidthAllocation() {
	mr.bandwidthManager.mutex.Lock()
	defer mr.bandwidthManager.mutex.Unlock()

	// Calculate current usage
	totalUsage := int64(0)
	for _, usage := range mr.bandwidthManager.currentUsage {
		totalUsage += usage
	}

	// Adjust allocations based on actual usage patterns
	tensorUsage := mr.bandwidthManager.currentUsage["tensor"]
	regularUsage := totalUsage - tensorUsage

	// Dynamic rebalancing: give more bandwidth to the service that needs it
	if tensorUsage > mr.bandwidthManager.tensorBandwidth &&
	   regularUsage < mr.bandwidthManager.regularBandwidth/2 {
		// Reallocate some regular bandwidth to tensor
		reallocation := min(mr.bandwidthManager.regularBandwidth/4, tensorUsage-mr.bandwidthManager.tensorBandwidth)
		mr.bandwidthManager.tensorBandwidth += reallocation
		mr.bandwidthManager.regularBandwidth -= reallocation
	}

	// Reset usage counters for next measurement period
	mr.bandwidthManager.currentUsage = make(map[string]int64)
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// EnhancedMessageCoalescing coalesces small tensor messages to reduce overhead
func (mr *MessageRouter) EnhancedMessageCoalescing(msgs []*Message) []*Message {
	if len(msgs) <= 1 {
		return msgs
	}

	var coalesced []*Message
	var currentBatch []*Message
	var currentBatchSize int

	const maxBatchSize = 64 * 1024 // 64KB batch size

	for _, msg := range msgs {
		if mr.isTensorStreamingMessage(msg) &&
		   msg.Type == MessageTypeTensorChunk &&
		   currentBatchSize + len(msg.Payload) <= maxBatchSize {
			// Add to current batch
			currentBatch = append(currentBatch, msg)
			currentBatchSize += len(msg.Payload)
		} else {
			// Flush current batch if it has messages
			if len(currentBatch) > 1 {
				coalescedMsg := mr.createCoalescedMessage(currentBatch)
				coalesced = append(coalesced, coalescedMsg)
			} else if len(currentBatch) == 1 {
				coalesced = append(coalesced, currentBatch[0])
			}

			// Start new batch
			currentBatch = []*Message{msg}
			currentBatchSize = len(msg.Payload)
		}
	}

	// Handle remaining batch
	if len(currentBatch) > 1 {
		coalescedMsg := mr.createCoalescedMessage(currentBatch)
		coalesced = append(coalesced, coalescedMsg)
	} else if len(currentBatch) == 1 {
		coalesced = append(coalesced, currentBatch[0])
	}

	return coalesced
}

// createCoalescedMessage creates a single coalesced message from multiple chunks
func (mr *MessageRouter) createCoalescedMessage(msgs []*Message) *Message {
	if len(msgs) == 0 {
		return nil
	}

	// Calculate total payload size
	totalSize := 0
	for _, msg := range msgs {
		totalSize += len(msg.Payload)
	}

	// Create coalesced payload
	coalescedPayload := make([]byte, 0, totalSize)
	for _, msg := range msgs {
		coalescedPayload = append(coalescedPayload, msg.Payload...)
	}

	// Create coalesced message based on first message
	baseMsg := msgs[0]
	coalescedMsg := &Message{
		ID:          generateMessageID(),
		Type:        baseMsg.Type,
		Protocol:    baseMsg.Protocol,
		Source:      baseMsg.Source,
		Destination: baseMsg.Destination,
		Payload:     coalescedPayload,
		Headers:     make(map[string]string),
		Timestamp:   time.Now(),
		TTL:         baseMsg.TTL,
		Priority:    baseMsg.Priority,
		RequiresAck: baseMsg.RequiresAck,
		Compressed:  false,
	}

	// Copy headers from base message
	for key, value := range baseMsg.Headers {
		coalescedMsg.Headers[key] = value
	}

	// Mark as coalesced
	coalescedMsg.Headers["coalesced"] = "true"
	coalescedMsg.Headers["original_count"] = fmt.Sprintf("%d", len(msgs))

	return coalescedMsg
}
