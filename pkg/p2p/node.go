package p2p

import (
	"context"
	"time"
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
}

// BasicNode implements the Node interface
type BasicNode struct {
	id       string
	address  string
	config   *NodeConfig
	peers    map[string]*PeerInfo
	handlers map[string]MessageHandler
	status   NodeStatus
}

// NewBasicNode creates a new basic P2P node
func NewBasicNode(id, address string, config *NodeConfig) *BasicNode {
	if config == nil {
		config = DefaultNodeConfig()
	}
	
	return &BasicNode{
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
	}
}

// Start implements the Node interface
func (n *BasicNode) Start(ctx context.Context) error {
	n.status.Connected = true
	n.status.LastUpdate = time.Now()
	return nil
}

// Stop implements the Node interface
func (n *BasicNode) Stop(ctx context.Context) error {
	n.status.Connected = false
	n.status.LastUpdate = time.Now()
	return nil
}

// ID implements the Node interface
func (n *BasicNode) ID() string {
	return n.id
}

// Connect implements the Node interface
func (n *BasicNode) Connect(ctx context.Context, peerAddr string) error {
	// Implementation would go here
	return nil
}

// Disconnect implements the Node interface
func (n *BasicNode) Disconnect(ctx context.Context, peerID string) error {
	delete(n.peers, peerID)
	n.status.PeerCount = len(n.peers)
	n.status.LastUpdate = time.Now()
	return nil
}

// Broadcast implements the Node interface
func (n *BasicNode) Broadcast(ctx context.Context, topic string, data []byte) error {
	// Implementation would go here
	return nil
}

// Subscribe implements the Node interface
func (n *BasicNode) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	n.handlers[topic] = handler
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