package protocols

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// MessageType represents the type of a protocol message
type MessageType uint8

const (
	// Protocol ID for shard protocol
	ShardProtocolID = protocol.ID("/ollama-distributed/shard/1.0.0")

	// Message types for shard protocol
	MsgTypeShardAnnounce            MessageType = 0x40
	MsgTypeShardRequest             MessageType = 0x41
	MsgTypeShardTransferCoordination MessageType = 0x42
	MsgTypeShardStatus              MessageType = 0x43
	MsgTypeShardLocate              MessageType = 0x44
	MsgTypeShardLocateResponse      MessageType = 0x45
)

// ShardAnnounceMessage announces available shards
type ShardAnnounceMessage struct {
	NodeID    string      `json:"node_id"`
	ShardIDs  []string    `json:"shard_ids"`
	ModelName string      `json:"model_name"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// ShardRequestMessage requests specific shards
type ShardRequestMessage struct {
	RequestID string   `json:"request_id"`
	ShardIDs  []string `json:"shard_ids"`
	ModelName string   `json:"model_name"`
	Priority  int      `json:"priority"`
}

// ShardTransferCoordinationMessage coordinates shard transfers
type ShardTransferCoordinationMessage struct {
	TransferID   string    `json:"transfer_id"`
	ShardID      string    `json:"shard_id"`
	SourceNode   string    `json:"source_node"`
	TargetNode   string    `json:"target_node"`
	Action       string    `json:"action"` // "initiate", "progress", "complete", "cancel"
	Progress     float64   `json:"progress,omitempty"`
	BytesTransferred int64 `json:"bytes_transferred,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// ShardStatusMessage reports shard status
type ShardStatusMessage struct {
	ShardID      string    `json:"shard_id"`
	ModelName    string    `json:"model_name"`
	NodeID       string    `json:"node_id"`
	IsAvailable  bool      `json:"is_available"`
	IsLoaded     bool      `json:"is_loaded"`
	StoragePath  string    `json:"storage_path"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	LastAccessed time.Time `json:"last_accessed"`
}

// ShardLocateMessage queries for shard locations
type ShardLocateMessage struct {
	RequestID string `json:"request_id"`
	ShardID   string `json:"shard_id"`
	ModelName string `json:"model_name"`
}

// ShardLocateResponseMessage responds with shard locations
type ShardLocateResponseMessage struct {
	RequestID string              `json:"request_id"`
	ShardID   string              `json:"shard_id"`
	Locations []ShardNodeLocation `json:"locations"`
}

// ShardNodeLocation represents a shard's location on a node
type ShardNodeLocation struct {
	NodeID      string    `json:"node_id"`
	PeerID      peer.ID   `json:"peer_id"`
	IsAvailable bool      `json:"is_available"`
	IsLoaded    bool      `json:"is_loaded"`
	IsLocal     bool      `json:"is_local"`
	StoragePath string    `json:"storage_path"`
	LastSeen    time.Time `json:"last_seen"`
}

// ShardProtocolHandler handles shard protocol messages
type ShardProtocolHandler struct {
	host             host.Host
	logger           *slog.Logger
	shardRegistry    ShardRegistry
	announcements    map[string]*ShardAnnounceMessage
	announcementsMux sync.RWMutex
	requests         map[string]*ShardRequestMessage
	requestsMux      sync.RWMutex
	transfers        map[string]*ShardTransferCoordinationMessage
	transfersMux     sync.RWMutex
}

// ShardRegistry interface for shard management
type ShardRegistry interface {
	RegisterShard(shardID, modelName string, location ShardNodeLocation) error
	LocateShard(shardID string) ([]ShardNodeLocation, error)
	GetLocalShards() ([]string, error)
	UpdateShardStatus(shardID string, status ShardStatusMessage) error
}

// NewShardProtocolHandler creates a new shard protocol handler
func NewShardProtocolHandler(h host.Host, logger *slog.Logger, registry ShardRegistry) *ShardProtocolHandler {
	handler := &ShardProtocolHandler{
		host:          h,
		logger:        logger,
		shardRegistry: registry,
		announcements: make(map[string]*ShardAnnounceMessage),
		requests:      make(map[string]*ShardRequestMessage),
		transfers:     make(map[string]*ShardTransferCoordinationMessage),
	}

	// Register stream handler
	h.SetStreamHandler(ShardProtocolID, handler.handleStream)

	return handler
}

// HandleMessage implements MessageHandler interface
func (h *ShardProtocolHandler) HandleMessage(ctx context.Context, msgType MessageType, data []byte, stream network.Stream) error {
	switch msgType {
	case MsgTypeShardAnnounce:
		return h.handleShardAnnounce(ctx, data, stream)
	case MsgTypeShardRequest:
		return h.handleShardRequest(ctx, data, stream)
	case MsgTypeShardTransferCoordination:
		return h.handleShardTransferCoordination(ctx, data, stream)
	case MsgTypeShardStatus:
		return h.handleShardStatus(ctx, data, stream)
	case MsgTypeShardLocate:
		return h.handleShardLocate(ctx, data, stream)
	case MsgTypeShardLocateResponse:
		return h.handleShardLocateResponse(ctx, data, stream)
	default:
		return fmt.Errorf("unknown message type: %d", msgType)
	}
}

// handleStream handles incoming streams
func (h *ShardProtocolHandler) handleStream(stream network.Stream) {
	defer stream.Close()

	// Read message header
	header := make([]byte, 5) // 1 byte type + 4 bytes length
	if _, err := io.ReadFull(stream, header); err != nil {
		h.logger.Error("Failed to read message header", "error", err)
		return
	}

	msgType := MessageType(header[0])
	length := uint32(header[1])<<24 | uint32(header[2])<<16 | uint32(header[3])<<8 | uint32(header[4])

	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		h.logger.Error("Failed to read message data", "error", err)
		return
	}

	// Handle message
	ctx := context.Background()
	if err := h.HandleMessage(ctx, msgType, data, stream); err != nil {
		h.logger.Error("Failed to handle message", "type", msgType, "error", err)
	}
}

// handleShardAnnounce handles shard announcements
func (h *ShardProtocolHandler) handleShardAnnounce(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardAnnounceMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal announce message: %w", err)
	}

	h.logger.Info("Received shard announcement",
		"node_id", msg.NodeID,
		"model", msg.ModelName,
		"shard_count", len(msg.ShardIDs))

	// Store announcement
	h.announcementsMux.Lock()
	h.announcements[msg.NodeID] = &msg
	h.announcementsMux.Unlock()

	// Register shards in registry
	peerID := stream.Conn().RemotePeer()
	for _, shardID := range msg.ShardIDs {
		location := ShardNodeLocation{
			NodeID:      msg.NodeID,
			PeerID:      peerID,
			IsAvailable: true,
			IsLoaded:    false,
			IsLocal:     false, // Remote shard announcement
			StoragePath: "", // Will be set by the registry
			LastSeen:    time.Now(),
		}
		if err := h.shardRegistry.RegisterShard(shardID, msg.ModelName, location); err != nil {
			h.logger.Error("Failed to register shard", "shard_id", shardID, "error", err)
		}
	}

	return nil
}

// handleShardRequest handles shard requests
func (h *ShardProtocolHandler) handleShardRequest(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardRequestMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal request message: %w", err)
	}

	h.logger.Info("Received shard request",
		"request_id", msg.RequestID,
		"model", msg.ModelName,
		"shard_count", len(msg.ShardIDs))

	// Store request
	h.requestsMux.Lock()
	h.requests[msg.RequestID] = &msg
	h.requestsMux.Unlock()

	// Check local availability and respond
	localShards, err := h.shardRegistry.GetLocalShards()
	if err != nil {
		return fmt.Errorf("failed to get local shards: %w", err)
	}

	localSet := make(map[string]bool)
	for _, id := range localShards {
		localSet[id] = true
	}

	// Send status for available shards
	for _, shardID := range msg.ShardIDs {
		if localSet[shardID] {
			status := ShardStatusMessage{
				ShardID:     shardID,
				ModelName:   msg.ModelName,
				NodeID:      h.host.ID().String(),
				IsAvailable: true,
				IsLoaded:    false, // Would check actual load status
			}

			statusData, err := json.Marshal(status)
			if err != nil {
				continue
			}

			// Send status back
			if err := h.sendMessage(stream.Conn().RemotePeer(), MsgTypeShardStatus, statusData); err != nil {
				h.logger.Error("Failed to send shard status", "error", err)
			}
		}
	}

	return nil
}

// handleShardTransferCoordination handles transfer coordination messages
func (h *ShardProtocolHandler) handleShardTransferCoordination(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardTransferCoordinationMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal transfer coordination message: %w", err)
	}

	h.logger.Info("Received transfer coordination",
		"transfer_id", msg.TransferID,
		"shard_id", msg.ShardID,
		"action", msg.Action)

	// Store transfer state
	h.transfersMux.Lock()
	h.transfers[msg.TransferID] = &msg
	h.transfersMux.Unlock()

	// Handle based on action
	switch msg.Action {
	case "initiate":
		// Initiate transfer if we're the source
		if msg.SourceNode == h.host.ID().String() {
			// Would trigger actual transfer here
			h.logger.Info("Initiating shard transfer", "shard_id", msg.ShardID, "target", msg.TargetNode)
		}
	case "progress":
		// Update progress tracking
		h.logger.Debug("Transfer progress", "transfer_id", msg.TransferID, "progress", msg.Progress)
	case "complete":
		// Mark transfer as complete
		h.logger.Info("Transfer completed", "transfer_id", msg.TransferID)
	case "cancel":
		// Cancel transfer
		h.logger.Info("Transfer cancelled", "transfer_id", msg.TransferID, "error", msg.Error)
	}

	return nil
}

// handleShardStatus handles shard status messages
func (h *ShardProtocolHandler) handleShardStatus(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardStatusMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal status message: %w", err)
	}

	h.logger.Debug("Received shard status",
		"shard_id", msg.ShardID,
		"node_id", msg.NodeID,
		"available", msg.IsAvailable)

	// Update registry
	if err := h.shardRegistry.UpdateShardStatus(msg.ShardID, msg); err != nil {
		return fmt.Errorf("failed to update shard status: %w", err)
	}

	return nil
}

// handleShardLocate handles shard locate queries
func (h *ShardProtocolHandler) handleShardLocate(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardLocateMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal locate message: %w", err)
	}

	h.logger.Debug("Received shard locate request",
		"request_id", msg.RequestID,
		"shard_id", msg.ShardID)

	// Query registry for locations
	locations, err := h.shardRegistry.LocateShard(msg.ShardID)
	if err != nil {
		h.logger.Error("Failed to locate shard", "shard_id", msg.ShardID, "error", err)
		locations = []ShardNodeLocation{} // Send empty response
	}

	// Send response
	response := ShardLocateResponseMessage{
		RequestID: msg.RequestID,
		ShardID:   msg.ShardID,
		Locations: locations,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal locate response: %w", err)
	}

	return h.sendMessage(stream.Conn().RemotePeer(), MsgTypeShardLocateResponse, responseData)
}

// handleShardLocateResponse handles locate responses
func (h *ShardProtocolHandler) handleShardLocateResponse(ctx context.Context, data []byte, stream network.Stream) error {
	var msg ShardLocateResponseMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal locate response: %w", err)
	}

	h.logger.Debug("Received locate response",
		"request_id", msg.RequestID,
		"shard_id", msg.ShardID,
		"location_count", len(msg.Locations))

	// Would handle response based on request context
	// This would typically update a waiting request or trigger next steps

	return nil
}

// sendMessage sends a message to a peer
func (h *ShardProtocolHandler) sendMessage(peerID peer.ID, msgType MessageType, data []byte) error {
	stream, err := h.host.NewStream(context.Background(), peerID, ShardProtocolID)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Create header
	length := uint32(len(data))
	header := []byte{
		byte(msgType),
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	// Send header and data
	if _, err := stream.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// AnnounceShards announces available shards to the network
func (h *ShardProtocolHandler) AnnounceShards(modelName string, shardIDs []string) error {
	msg := ShardAnnounceMessage{
		NodeID:    h.host.ID().String(),
		ShardIDs:  shardIDs,
		ModelName: modelName,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal announce message: %w", err)
	}

	// Broadcast to all connected peers
	for _, peer := range h.host.Network().Peers() {
		if err := h.sendMessage(peer, MsgTypeShardAnnounce, data); err != nil {
			h.logger.Error("Failed to announce to peer", "peer", peer, "error", err)
		}
	}

	return nil
}

// RequestShards requests specific shards from the network
func (h *ShardProtocolHandler) RequestShards(requestID, modelName string, shardIDs []string, priority int) error {
	msg := ShardRequestMessage{
		RequestID: requestID,
		ShardIDs:  shardIDs,
		ModelName: modelName,
		Priority:  priority,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal request message: %w", err)
	}

	// Broadcast request to all peers
	for _, peer := range h.host.Network().Peers() {
		if err := h.sendMessage(peer, MsgTypeShardRequest, data); err != nil {
			h.logger.Error("Failed to send request to peer", "peer", peer, "error", err)
		}
	}

	return nil
}

// CoordinateTransfer coordinates a shard transfer
func (h *ShardProtocolHandler) CoordinateTransfer(transferID, shardID, sourceNode, targetNode, action string) error {
	msg := ShardTransferCoordinationMessage{
		TransferID: transferID,
		ShardID:    shardID,
		SourceNode: sourceNode,
		TargetNode: targetNode,
		Action:     action,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal coordination message: %w", err)
	}

	// Send to relevant peers
	for _, peer := range h.host.Network().Peers() {
		if err := h.sendMessage(peer, MsgTypeShardTransferCoordination, data); err != nil {
			h.logger.Error("Failed to send coordination to peer", "peer", peer, "error", err)
		}
	}

	return nil
}

// LocateShard queries the network for shard locations
func (h *ShardProtocolHandler) LocateShard(requestID, shardID, modelName string) error {
	msg := ShardLocateMessage{
		RequestID: requestID,
		ShardID:   shardID,
		ModelName: modelName,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal locate message: %w", err)
	}

	// Query all peers
	for _, peer := range h.host.Network().Peers() {
		if err := h.sendMessage(peer, MsgTypeShardLocate, data); err != nil {
			h.logger.Error("Failed to send locate query to peer", "peer", peer, "error", err)
		}
	}

	return nil
}

// ShardProtocolClient provides client interface for shard protocol
type ShardProtocolClient struct {
	handler *ShardProtocolHandler
}

// NewShardProtocolClient creates a new shard protocol client
func NewShardProtocolClient(handler *ShardProtocolHandler) *ShardProtocolClient {
	return &ShardProtocolClient{
		handler: handler,
	}
}

// AnnounceShards announces shards
func (c *ShardProtocolClient) AnnounceShards(modelName string, shardIDs []string) error {
	return c.handler.AnnounceShards(modelName, shardIDs)
}

// RequestShards requests shards
func (c *ShardProtocolClient) RequestShards(requestID, modelName string, shardIDs []string, priority int) error {
	return c.handler.RequestShards(requestID, modelName, shardIDs, priority)
}

// CoordinateTransfer coordinates transfer
func (c *ShardProtocolClient) CoordinateTransfer(transferID, shardID, sourceNode, targetNode, action string) error {
	return c.handler.CoordinateTransfer(transferID, shardID, sourceNode, targetNode, action)
}

// LocateShard locates a shard
func (c *ShardProtocolClient) LocateShard(requestID, shardID, modelName string) error {
	return c.handler.LocateShard(requestID, shardID, modelName)
}