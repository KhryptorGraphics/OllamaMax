package messaging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// MessageRouter handles routing of messages between peers in the distributed system
type MessageRouter struct {
	config *RouterConfig

	// Protocol handlers
	handlers   map[protocol.ID]ProtocolHandler
	handlersMu sync.RWMutex

	// Message queues
	outboundQueue *MessageQueue
	inboundQueue  *MessageQueue

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

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

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
func NewMessageRouter(config *RouterConfig) *MessageRouter {
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

	// Initialize routing table
	router.routingTable = &RoutingTable{
		routes:         make(map[peer.ID]*RouteEntry),
		neighbors:      make(map[peer.ID]bool),
		discoveryQueue: make(chan peer.ID, 1000),
		lastDiscovery:  make(map[peer.ID]time.Time),
	}

	return router
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

// SendMessage sends a message to a peer
func (mr *MessageRouter) SendMessage(msg *Message) error {
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
	// Implementation would return the actual local peer ID
	// For now, return empty
	return ""
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
