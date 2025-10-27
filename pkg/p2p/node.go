package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, from string, data []byte) error

// PeerInfo represents detailed information about a peer
type PeerInfo struct {
	ID       string        `json:"id"`
	Address  string        `json:"address"`
	Latency  time.Duration `json:"latency"`
	LastSeen time.Time     `json:"last_seen"`
}

// NodeStatus represents the status of a node
type NodeStatus struct {
	ID          string      `json:"id"`
	Address     string      `json:"address"`
	Connected   bool        `json:"connected"`
	PeerCount   int         `json:"peer_count"`
	LastUpdate  time.Time   `json:"last_update"`
	Protocols   []string    `json:"protocols"`
}

// Node represents a P2P network node
type Node interface {
	// Start starts the node
	Start(ctx context.Context) error

	// Stop stops the node
	Stop(ctx context.Context) error

	// ID returns the node's unique identifier
	ID() string

	// Connect connects to a peer
	Connect(ctx context.Context, peerAddr string) error

	// Disconnect disconnects from a peer
	Disconnect(ctx context.Context, peerID string) error

	// Broadcast broadcasts a message to all connected peers
	Broadcast(ctx context.Context, topic string, data []byte) error

	// Subscribe subscribes to a topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error

	// GetPeers returns the list of connected peers
	GetPeers() []PeerInfo

	// GetStatus returns the node status
	GetStatus() NodeStatus

	// GetPrometheusRegistry returns the Prometheus registry for metrics
	GetPrometheusRegistry() *prometheus.Registry
}

// BasicNode implements the Node interface
type BasicNode struct {
	id       string
	address  string
	config   *NodeConfig
	peers    map[string]*PeerInfo
	handlers map[string]MessageHandler
	status   NodeStatus

	// Prometheus metrics
	registry             *prometheus.Registry
	connectedPeers       prometheus.Gauge
	messagesSent         *prometheus.CounterVec
	messagesReceived     *prometheus.CounterVec
	messageLatency       prometheus.Histogram
	connectionErrors     prometheus.Counter
}

// NewBasicNode creates a new basic P2P node
func NewBasicNode(id, address string, config *NodeConfig) *BasicNode {
	if config == nil {
		config = DefaultNodeConfig()
	}

	// Create Prometheus registry
	registry := prometheus.NewRegistry()

	// Initialize metrics
	connectedPeers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "p2p_connected_peers",
		Help: "Number of currently connected peers",
	})

	messagesSent := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "p2p_messages_sent_total",
			Help: "Total number of messages sent by topic",
		},
		[]string{"topic"},
	)

	messagesReceived := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "p2p_messages_received_total",
			Help: "Total number of messages received by topic",
		},
		[]string{"topic"},
	)

	messageLatency := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "p2p_message_latency_seconds",
		Help:    "Message processing latency in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
	})

	connectionErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "p2p_connection_errors_total",
		Help: "Total number of connection errors",
	})

	// Register metrics
	registry.MustRegister(connectedPeers)
	registry.MustRegister(messagesSent)
	registry.MustRegister(messagesReceived)
	registry.MustRegister(messageLatency)
	registry.MustRegister(connectionErrors)

	node := &BasicNode{
		id:       id,
		address:  address,
		config:   config,
		peers:    make(map[string]*PeerInfo),
		handlers: make(map[string]MessageHandler),
		status: NodeStatus{
			ID:         id,
			Address:    address,
			Connected:  false,
			PeerCount:  0,
			LastUpdate: time.Now(),
			Protocols:  []string{"ollamamax/1.0.0"},
		},
		registry:         registry,
		connectedPeers:   connectedPeers,
		messagesSent:     messagesSent,
		messagesReceived: messagesReceived,
		messageLatency:   messageLatency,
		connectionErrors: connectionErrors,
	}

	// Initialize connected peers gauge
	node.connectedPeers.Set(0)

	return node
}

// Start implements the Node interface
func (n *BasicNode) Start(ctx context.Context) error {
	n.status.Connected = true
	n.status.LastUpdate = time.Now()

	// Update connected peers gauge
	n.connectedPeers.Set(float64(len(n.peers)))

	return nil
}

// Stop implements the Node interface
func (n *BasicNode) Stop(ctx context.Context) error {
	n.status.Connected = false
	n.status.LastUpdate = time.Now()

	// Reset connected peers gauge
	n.connectedPeers.Set(0)

	return nil
}

// ID implements the Node interface
func (n *BasicNode) ID() string {
	return n.id
}

// Connect implements the Node interface
func (n *BasicNode) Connect(ctx context.Context, peerAddr string) error {
	// Simulate connection logic
	// In a real implementation, this would establish a network connection

	// For this example, we'll create a mock peer
	peerID := fmt.Sprintf("peer-%s", peerAddr)

	// Simulate potential connection error
	if peerAddr == "" {
		n.connectionErrors.Inc()
		return fmt.Errorf("invalid peer address")
	}

	// Add peer to the peers map
	n.peers[peerID] = &PeerInfo{
		ID:       peerID,
		Address:  peerAddr,
		Latency:  time.Millisecond * 50, // Mock latency
		LastSeen: time.Now(),
	}

	// Update status
	n.status.PeerCount = len(n.peers)
	n.status.LastUpdate = time.Now()

	// Update connected peers gauge on success
	n.connectedPeers.Set(float64(len(n.peers)))

	return nil
}

// Disconnect implements the Node interface
func (n *BasicNode) Disconnect(ctx context.Context, peerID string) error {
	delete(n.peers, peerID)
	n.status.PeerCount = len(n.peers)
	n.status.LastUpdate = time.Now()

	// Decrement connected peers gauge
	n.connectedPeers.Set(float64(len(n.peers)))

	return nil
}

// Broadcast implements the Node interface
func (n *BasicNode) Broadcast(ctx context.Context, topic string, data []byte) error {
	// Simulate broadcast logic
	// In a real implementation, this would send the message to all connected peers

	// Increment messages sent counter with topic label
	n.messagesSent.WithLabelValues(topic).Inc()

	return nil
}

// Subscribe implements the Node interface
func (n *BasicNode) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	// Wrap the handler to add metrics
	wrappedHandler := func(ctx context.Context, from string, data []byte) error {
		// Start timer for latency measurement
		start := time.Now()

		// Increment messages received counter with topic label
		n.messagesReceived.WithLabelValues(topic).Inc()

		// Call the original handler
		err := handler(ctx, from, data)

		// Observe message processing latency
		duration := time.Since(start).Seconds()
		n.messageLatency.Observe(duration)

		return err
	}

	n.handlers[topic] = wrappedHandler
	return nil
}

// GetPeers implements the Node interface
func (n *BasicNode) GetPeers() []PeerInfo {
	peers := make([]PeerInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, *peer)
	}
	return peers
}

// GetStatus implements the Node interface
func (n *BasicNode) GetStatus() NodeStatus {
	n.status.LastUpdate = time.Now()
	return n.status
}

// GetPrometheusRegistry implements the Node interface
func (n *BasicNode) GetPrometheusRegistry() *prometheus.Registry {
	return n.registry
}
