package consensus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/libp2p/go-libp2p/core/peer"
)

// P2PTransport implements Raft transport over P2P messaging
type P2PTransport struct {
	config *P2PTransportConfig

	// Messaging integration
	messageRouter    *messaging.MessageRouter
	consensusHandler *messaging.ConsensusHandler

	// Local node information
	localAddr raft.ServerAddress
	localID   raft.ServerID
	peerID    peer.ID

	// Connection management
	connections   map[raft.ServerAddress]*P2PConnection
	connectionsMu sync.RWMutex

	// Message handling
	consumer        chan raft.RPC
	heartbeatFn     func(raft.RPC)
	heartbeatFnLock sync.Mutex

	// Lifecycle
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdown   bool
	shutdownMu sync.RWMutex
}

// P2PTransportConfig configures the P2P transport
type P2PTransportConfig struct {
	// Connection settings
	MaxConnections    int
	ConnectionTimeout time.Duration
	HeartbeatTimeout  time.Duration

	// Message settings
	MaxMessageSize int
	MessageTimeout time.Duration

	// Performance settings
	BufferSize  int
	WorkerCount int
}

// P2PConnection represents a connection to a peer
type P2PConnection struct {
	target raft.ServerAddress
	peerID peer.ID

	// Message channels
	sendCh chan *raftMessage

	// Connection state
	connected    bool
	connectedAt  time.Time
	lastActivity time.Time

	// Statistics
	messagesSent     int64
	messagesReceived int64

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// raftMessage represents a Raft message for P2P transport
type raftMessage struct {
	Type       string                `json:"type"`
	Data       []byte                `json:"data"`
	Target     raft.ServerAddress    `json:"target"`
	ResponseCh chan raft.RPCResponse `json:"-"`
	Timestamp  time.Time             `json:"timestamp"`
}

// NewP2PTransport creates a new P2P transport for Raft
func NewP2PTransport(config *P2PTransportConfig, messageRouter *messaging.MessageRouter, peerID peer.ID, localAddr raft.ServerAddress) (*P2PTransport, error) {
	if config == nil {
		config = &P2PTransportConfig{
			MaxConnections:    1000,
			ConnectionTimeout: 30 * time.Second,
			HeartbeatTimeout:  10 * time.Second,
			MaxMessageSize:    10 * 1024 * 1024, // 10MB
			MessageTimeout:    30 * time.Second,
			BufferSize:        1000,
			WorkerCount:       5,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	transport := &P2PTransport{
		config:        config,
		messageRouter: messageRouter,
		localAddr:     localAddr,
		localID:       raft.ServerID(peerID.String()),
		peerID:        peerID,
		connections:   make(map[raft.ServerAddress]*P2PConnection),
		consumer:      make(chan raft.RPC, config.BufferSize),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Create consensus handler
	transport.consensusHandler = messaging.NewConsensusHandler(peerID)

	// Register message callbacks
	transport.setupMessageHandlers()

	// Register handler with message router
	messageRouter.RegisterHandler(transport.consensusHandler)

	return transport, nil
}

// setupMessageHandlers sets up message handlers for different Raft message types
func (t *P2PTransport) setupMessageHandlers() {
	// Handle RequestVote messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusRequestVote, t.handleRequestVote)

	// Handle VoteResponse messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusVoteResponse, t.handleVoteResponse)

	// Handle AppendEntries messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusAppendEntries, t.handleAppendEntries)

	// Handle AppendResponse messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusAppendResponse, t.handleAppendResponse)

	// Handle Heartbeat messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusHeartbeat, t.handleHeartbeat)

	// Handle InstallSnapshot messages
	t.consensusHandler.RegisterCallback(messaging.ConsensusInstallSnapshot, t.handleInstallSnapshot)
}

// Consumer returns the channel for consuming RPC messages
func (t *P2PTransport) Consumer() <-chan raft.RPC {
	return t.consumer
}

// LocalAddr returns the local address
func (t *P2PTransport) LocalAddr() raft.ServerAddress {
	return t.localAddr
}

// AppendEntriesPipeline returns an interface for pipelining AppendEntries requests
func (t *P2PTransport) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	// For simplicity, we'll use a basic pipeline implementation
	// In production, this could be optimized for better performance
	return &P2PPipeline{
		transport: t,
		target:    target,
		peerID:    peer.ID(id),
	}, nil
}

// AppendEntries sends an AppendEntries RPC to the target
func (t *P2PTransport) AppendEntries(id raft.ServerID, target raft.ServerAddress, args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	return t.sendRPC(target, "AppendEntries", args, resp)
}

// RequestVote sends a RequestVote RPC to the target
func (t *P2PTransport) RequestVote(id raft.ServerID, target raft.ServerAddress, args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	return t.sendRPC(target, "RequestVote", args, resp)
}

// InstallSnapshot sends an InstallSnapshot RPC to the target
func (t *P2PTransport) InstallSnapshot(id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader) error {
	// For InstallSnapshot, we need to handle the data stream
	// This is a simplified implementation
	return t.sendRPC(target, "InstallSnapshot", args, resp)
}

// EncodePeer encodes a peer address
func (t *P2PTransport) EncodePeer(id raft.ServerID, addr raft.ServerAddress) []byte {
	return []byte(addr)
}

// DecodePeer decodes a peer address
func (t *P2PTransport) DecodePeer(buf []byte) raft.ServerAddress {
	return raft.ServerAddress(buf)
}

// SetHeartbeatHandler sets the heartbeat handler
func (t *P2PTransport) SetHeartbeatHandler(cb func(raft.RPC)) {
	t.heartbeatFnLock.Lock()
	defer t.heartbeatFnLock.Unlock()
	t.heartbeatFn = cb
}

// TimeoutNow sends a TimeoutNow RPC to the target
func (t *P2PTransport) TimeoutNow(id raft.ServerID, target raft.ServerAddress, args *raft.TimeoutNowRequest, resp *raft.TimeoutNowResponse) error {
	return t.sendRPC(target, "TimeoutNow", args, resp)
}

// CloseConnection closes a connection to a peer
func (t *P2PTransport) CloseConnection(target raft.ServerAddress) error {
	t.connectionsMu.Lock()
	defer t.connectionsMu.Unlock()

	if conn, exists := t.connections[target]; exists {
		conn.Close()
		delete(t.connections, target)
	}

	return nil
}

// Close closes the transport
func (t *P2PTransport) Close() error {
	t.shutdownMu.Lock()
	defer t.shutdownMu.Unlock()

	if t.shutdown {
		return nil
	}

	t.shutdown = true
	t.cancel()

	// Close all connections
	t.connectionsMu.Lock()
	for _, conn := range t.connections {
		conn.Close()
	}
	t.connectionsMu.Unlock()

	// Close consumer channel
	close(t.consumer)

	// Wait for goroutines
	t.wg.Wait()

	return nil
}

// sendRPC sends an RPC message to a target
func (t *P2PTransport) sendRPC(target raft.ServerAddress, rpcType string, args interface{}, resp interface{}) error {
	// Serialize the arguments
	data, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("failed to marshal RPC args: %w", err)
	}

	// Create response channel
	respCh := make(chan raft.RPCResponse, 1)

	// Create Raft message
	msg := &raftMessage{
		Type:       rpcType,
		Data:       data,
		Target:     target,
		ResponseCh: respCh,
		Timestamp:  time.Now(),
	}

	// Get or create connection
	conn, err := t.getConnection(target)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Send message
	select {
	case conn.sendCh <- msg:
		// Wait for response
		select {
		case response := <-respCh:
			if response.Error != nil {
				return response.Error
			}

			// Unmarshal response
			if responseData, ok := response.Response.([]byte); ok {
				if err := json.Unmarshal(responseData, resp); err != nil {
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}
			} else {
				return fmt.Errorf("invalid response type")
			}

			return nil

		case <-time.After(t.config.MessageTimeout):
			return fmt.Errorf("RPC timeout")
		}

	case <-time.After(t.config.MessageTimeout):
		return fmt.Errorf("send timeout")
	}
}

// getConnection gets or creates a connection to a target
func (t *P2PTransport) getConnection(target raft.ServerAddress) (*P2PConnection, error) {
	t.connectionsMu.RLock()
	if conn, exists := t.connections[target]; exists {
		t.connectionsMu.RUnlock()
		return conn, nil
	}
	t.connectionsMu.RUnlock()

	t.connectionsMu.Lock()
	defer t.connectionsMu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := t.connections[target]; exists {
		return conn, nil
	}

	// Create new connection
	conn, err := t.createConnection(target)
	if err != nil {
		return nil, err
	}

	t.connections[target] = conn
	return conn, nil
}

// createConnection creates a new connection to a target
func (t *P2PTransport) createConnection(target raft.ServerAddress) (*P2PConnection, error) {
	// Parse target to get peer ID
	// In a real implementation, you'd have a mapping from ServerAddress to peer.ID
	targetPeerID, err := peer.Decode(string(target))
	if err != nil {
		return nil, fmt.Errorf("failed to decode target peer ID: %w", err)
	}

	ctx, cancel := context.WithCancel(t.ctx)

	conn := &P2PConnection{
		target:       target,
		peerID:       targetPeerID,
		sendCh:       make(chan *raftMessage, t.config.BufferSize),
		connected:    true,
		connectedAt:  time.Now(),
		lastActivity: time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start connection worker
	conn.wg.Add(1)
	go t.connectionWorker(conn)

	return conn, nil
}

// connectionWorker handles sending messages for a connection
func (t *P2PTransport) connectionWorker(conn *P2PConnection) {
	defer conn.wg.Done()

	for {
		select {
		case <-conn.ctx.Done():
			return

		case msg := <-conn.sendCh:
			if err := t.sendMessage(conn, msg); err != nil {
				// Log error but continue
				continue
			}

			conn.mu.Lock()
			conn.messagesSent++
			conn.lastActivity = time.Now()
			conn.mu.Unlock()
		}
	}
}

// sendMessage sends a message through the P2P network
func (t *P2PTransport) sendMessage(conn *P2PConnection, msg *raftMessage) error {
	// Convert to consensus message
	consensusMsg := &messaging.ConsensusMessage{
		Type: messaging.ConsensusMessageType(msg.Type),
		// Add other fields as needed based on message type
	}

	// Create P2P message
	p2pMsg, err := messaging.CreateConsensusMessage(
		messaging.ConsensusMessageType(msg.Type),
		t.peerID,
		conn.peerID,
		consensusMsg,
	)
	if err != nil {
		return fmt.Errorf("failed to create consensus message: %w", err)
	}

	// Send through message router
	return t.messageRouter.SendMessage(p2pMsg)
}

// Message handler implementations

func (t *P2PTransport) handleRequestVote(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Convert to Raft RPC and forward to consumer
	rpc := raft.RPC{
		Command: &raft.RequestVoteRequest{
			RPCHeader:    raft.RPCHeader{},
			Term:         msg.Term,
			Candidate:    []byte(msg.CandidateID),
			LastLogIndex: msg.PrevLogIndex,
			LastLogTerm:  msg.PrevLogTerm,
		},
		RespChan: make(chan raft.RPCResponse, 1),
	}

	select {
	case t.consumer <- rpc:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *P2PTransport) handleVoteResponse(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Handle vote response
	return nil
}

func (t *P2PTransport) handleAppendEntries(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Convert to Raft RPC and forward to consumer
	rpc := raft.RPC{
		Command: &raft.AppendEntriesRequest{
			RPCHeader:         raft.RPCHeader{},
			Term:              msg.Term,
			Leader:            []byte(msg.LeaderID),
			PrevLogEntry:      msg.PrevLogIndex,
			PrevLogTerm:       msg.PrevLogTerm,
			LeaderCommitIndex: msg.LeaderCommit,
		},
		RespChan: make(chan raft.RPCResponse, 1),
	}

	select {
	case t.consumer <- rpc:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *P2PTransport) handleAppendResponse(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Handle append response
	return nil
}

func (t *P2PTransport) handleHeartbeat(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Handle heartbeat
	t.heartbeatFnLock.Lock()
	fn := t.heartbeatFn
	t.heartbeatFnLock.Unlock()

	if fn != nil {
		rpc := raft.RPC{
			Command: &raft.AppendEntriesRequest{
				RPCHeader: raft.RPCHeader{},
				Term:      msg.Term,
				Leader:    []byte(msg.LeaderID),
			},
			RespChan: make(chan raft.RPCResponse, 1),
		}
		fn(rpc)
	}

	return nil
}

func (t *P2PTransport) handleInstallSnapshot(ctx context.Context, msg *messaging.ConsensusMessage) error {
	// Handle install snapshot
	return nil
}

// Close closes a P2P connection
func (c *P2PConnection) Close() error {
	c.cancel()
	c.wg.Wait()
	close(c.sendCh)
	return nil
}

// P2PPipeline implements AppendPipeline for P2P transport
type P2PPipeline struct {
	transport *P2PTransport
	target    raft.ServerAddress
	peerID    peer.ID
}

func (p *P2PPipeline) AppendEntries(args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) (raft.AppendFuture, error) {
	// Create future for async response
	future := &P2PAppendFuture{
		respCh: make(chan error, 1),
		resp:   resp,
	}

	// Send async
	go func() {
		err := p.transport.sendRPC(p.target, "AppendEntries", args, resp)
		future.respCh <- err
	}()

	return future, nil
}

func (p *P2PPipeline) Consumer() <-chan raft.AppendFuture {
	// This would return a channel of futures in a real implementation
	return nil
}

func (p *P2PPipeline) Close() error {
	return nil
}

// P2PAppendFuture implements AppendFuture
type P2PAppendFuture struct {
	respCh chan error
	resp   *raft.AppendEntriesResponse
}

func (f *P2PAppendFuture) Error() error {
	return <-f.respCh
}

func (f *P2PAppendFuture) Start() time.Time {
	return time.Now()
}

func (f *P2PAppendFuture) Request() *raft.AppendEntriesRequest {
	return nil
}

func (f *P2PAppendFuture) Response() *raft.AppendEntriesResponse {
	return f.resp
}
