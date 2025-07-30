package protocols

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Protocol IDs for the distributed Ollama system
const (
	// Core communication protocols
	InferenceProtocol    = protocol.ID("/ollama-distributed/inference/1.0.0")
	HealthCheckProtocol  = protocol.ID("/ollama-distributed/health/1.0.0")
	FileTransferProtocol = protocol.ID("/ollama-distributed/file-transfer/1.0.0")
	ModelSyncProtocol    = protocol.ID("/ollama-distributed/model-sync/1.0.0")

	// BitTorrent-style distribution protocols
	ModelChunkProtocol    = protocol.ID("/ollama-distributed/model-chunk/1.0.0")
	ChunkRequestProtocol  = protocol.ID("/ollama-distributed/chunk-request/1.0.0")
	ChunkAnnounceProtocol = protocol.ID("/ollama-distributed/chunk-announce/1.0.0")

	// Coordination protocols
	ConsensusProtocol = protocol.ID("/ollama-distributed/consensus/1.0.0")
	SchedulerProtocol = protocol.ID("/ollama-distributed/scheduler/1.0.0")
)

// Message types for different protocols
const (
	// Inference message types
	MsgTypeInferenceRequest  = "inference_request"
	MsgTypeInferenceResponse = "inference_response"

	// Health check message types
	MsgTypeHealthPing           = "health_ping"
	MsgTypeHealthPong           = "health_pong"
	MsgTypeCapabilitiesRequest  = "capabilities_request"
	MsgTypeCapabilitiesResponse = "capabilities_response"

	// File transfer message types
	MsgTypeFileRequest  = "file_request"
	MsgTypeFileResponse = "file_response"
	MsgTypeFileChunk    = "file_chunk"
	MsgTypeFileComplete = "file_complete"
	MsgTypeFileError    = "file_error"

	// Model sync message types
	MsgTypeModelAnnounce = "model_announce"
	MsgTypeModelRequest  = "model_request"
	MsgTypeModelList     = "model_list"

	// Chunk distribution message types
	MsgTypeChunkHave   = "chunk_have"
	MsgTypeChunkWant   = "chunk_want"
	MsgTypeChunkData   = "chunk_data"
	MsgTypeChunkCancel = "chunk_cancel"
)

// Maximum message sizes to prevent memory exhaustion
const (
	MaxMessageSize  = 64 * 1024 * 1024 // 64MB max message
	MaxChunkSize    = 1024 * 1024      // 1MB chunk size
	MaxMetadataSize = 1024 * 1024      // 1MB metadata
	MaxHeaderSize   = 4096             // 4KB header
)

// ProtocolHandler manages protocol message handling
type ProtocolHandler struct {
	protocolID   protocol.ID
	messageTypes map[string]MessageHandler
	metrics      *ProtocolMetrics
	mu           sync.RWMutex
}

// MessageHandler defines the interface for handling protocol messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, stream network.Stream, msg *Message) error
}

// Message represents a protocol message
type Message struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// ProtocolMetrics tracks protocol performance
type ProtocolMetrics struct {
	MessagesReceived  int64         `json:"messages_received"`
	MessagesSent      int64         `json:"messages_sent"`
	BytesReceived     int64         `json:"bytes_received"`
	BytesSent         int64         `json:"bytes_sent"`
	ErrorCount        int64         `json:"error_count"`
	AverageLatency    time.Duration `json:"average_latency"`
	LastActivity      time.Time     `json:"last_activity"`
	ActiveConnections int           `json:"active_connections"`

	// Message type specific metrics
	MessageTypeMetrics map[string]*MessageTypeMetrics `json:"message_type_metrics"`

	mu sync.RWMutex
}

// MessageTypeMetrics tracks metrics for specific message types
type MessageTypeMetrics struct {
	Count        int64         `json:"count"`
	Errors       int64         `json:"errors"`
	TotalLatency time.Duration `json:"total_latency"`
	LastSeen     time.Time     `json:"last_seen"`
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler(protocolID protocol.ID) *ProtocolHandler {
	return &ProtocolHandler{
		protocolID:   protocolID,
		messageTypes: make(map[string]MessageHandler),
		metrics: &ProtocolMetrics{
			MessageTypeMetrics: make(map[string]*MessageTypeMetrics),
		},
	}
}

// RegisterMessageHandler registers a handler for a specific message type
func (ph *ProtocolHandler) RegisterMessageHandler(msgType string, handler MessageHandler) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.messageTypes[msgType] = handler

	// Initialize metrics for this message type
	if ph.metrics.MessageTypeMetrics[msgType] == nil {
		ph.metrics.MessageTypeMetrics[msgType] = &MessageTypeMetrics{}
	}
}

// HandleStream handles incoming streams for this protocol
func (ph *ProtocolHandler) HandleStream(stream network.Stream) {
	defer stream.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	peerID := stream.Conn().RemotePeer()

	// Update connection metrics
	ph.updateConnectionMetrics(1)
	defer ph.updateConnectionMetrics(-1)

	// Read message with size limit
	msg, err := ph.readMessage(stream)
	if err != nil {
		log.Printf("Failed to read message from peer %s on protocol %s: %v", peerID, ph.protocolID, err)
		ph.updateErrorMetrics()
		return
	}

	// Update receive metrics
	ph.updateReceiveMetrics(msg)

	// Find and execute handler
	ph.mu.RLock()
	handler, exists := ph.messageTypes[msg.Type]
	ph.mu.RUnlock()

	if !exists {
		log.Printf("No handler for message type %s from peer %s", msg.Type, peerID)
		ph.sendErrorResponse(stream, msg.ID, "unknown_message_type", "No handler for message type")
		ph.updateErrorMetrics()
		return
	}

	// Handle message
	if err := handler.HandleMessage(ctx, stream, msg); err != nil {
		log.Printf("Handler error for message type %s from peer %s: %v", msg.Type, peerID, err)
		ph.sendErrorResponse(stream, msg.ID, "handler_error", err.Error())
		ph.updateErrorMetrics()
		return
	}

	// Update latency metrics
	latency := time.Since(start)
	ph.updateLatencyMetrics(msg.Type, latency)

	log.Printf("Successfully handled %s message from peer %s (latency: %v)", msg.Type, peerID, latency)
}

// readMessage reads and validates a message from the stream
func (ph *ProtocolHandler) readMessage(stream network.Stream) (*Message, error) {
	// Set read deadline
	if conn, ok := stream.Conn().(interface{ SetReadDeadline(time.Time) error }); ok {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	}

	reader := bufio.NewReader(stream)

	// Read header to get message size
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, fmt.Errorf("failed to read message header: %w", err)
	}

	// Parse message size from header
	messageSize := uint32(header[0])<<24 | uint32(header[1])<<16 | uint32(header[2])<<8 | uint32(header[3])

	// Validate message size
	if messageSize > MaxMessageSize {
		return nil, fmt.Errorf("message size %d exceeds maximum %d", messageSize, MaxMessageSize)
	}

	if messageSize == 0 {
		return nil, fmt.Errorf("invalid message size: 0")
	}

	// Read message data
	messageData := make([]byte, messageSize)
	if _, err := io.ReadFull(reader, messageData); err != nil {
		return nil, fmt.Errorf("failed to read message data: %w", err)
	}

	// Parse JSON message
	var msg Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Validate message
	if err := ph.validateMessage(&msg); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	return &msg, nil
}

// validateMessage validates a message structure
func (ph *ProtocolHandler) validateMessage(msg *Message) error {
	if msg.Type == "" {
		return fmt.Errorf("message type is required")
	}

	if msg.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if msg.Timestamp.IsZero() {
		return fmt.Errorf("message timestamp is required")
	}

	// Check timestamp is not too old or too far in the future
	now := time.Now()
	if now.Sub(msg.Timestamp) > 5*time.Minute {
		return fmt.Errorf("message timestamp too old")
	}

	if msg.Timestamp.Sub(now) > 1*time.Minute {
		return fmt.Errorf("message timestamp too far in future")
	}

	return nil
}

// SendMessage sends a message through a stream
func (ph *ProtocolHandler) SendMessage(stream network.Stream, msg *Message) error {
	// Set message timestamp if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Validate size
	if len(data) > MaxMessageSize {
		return fmt.Errorf("message size %d exceeds maximum %d", len(data), MaxMessageSize)
	}

	// Set write deadline
	if conn, ok := stream.Conn().(interface{ SetWriteDeadline(time.Time) error }); ok {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	}

	// Write header with message size
	header := make([]byte, 4)
	messageSize := uint32(len(data))
	header[0] = byte(messageSize >> 24)
	header[1] = byte(messageSize >> 16)
	header[2] = byte(messageSize >> 8)
	header[3] = byte(messageSize)

	if _, err := stream.Write(header); err != nil {
		return fmt.Errorf("failed to write message header: %w", err)
	}

	// Write message data
	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}

	// Update send metrics
	ph.updateSendMetrics(msg)

	return nil
}

// sendErrorResponse sends an error response message
func (ph *ProtocolHandler) sendErrorResponse(stream network.Stream, requestID, errorCode, errorMessage string) {
	errorMsg := &Message{
		Type:      "error",
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":    requestID,
			"error_code":    errorCode,
			"error_message": errorMessage,
		},
	}

	if err := ph.SendMessage(stream, errorMsg); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}

// Metrics update methods

func (ph *ProtocolHandler) updateReceiveMetrics(msg *Message) {
	ph.metrics.mu.Lock()
	defer ph.metrics.mu.Unlock()

	ph.metrics.MessagesReceived++
	ph.metrics.LastActivity = time.Now()

	// Update message type metrics
	if typeMetrics, exists := ph.metrics.MessageTypeMetrics[msg.Type]; exists {
		typeMetrics.Count++
		typeMetrics.LastSeen = time.Now()
	}
}

func (ph *ProtocolHandler) updateSendMetrics(msg *Message) {
	ph.metrics.mu.Lock()
	defer ph.metrics.mu.Unlock()

	ph.metrics.MessagesSent++
	ph.metrics.LastActivity = time.Now()
}

func (ph *ProtocolHandler) updateErrorMetrics() {
	ph.metrics.mu.Lock()
	defer ph.metrics.mu.Unlock()

	ph.metrics.ErrorCount++
}

func (ph *ProtocolHandler) updateConnectionMetrics(delta int) {
	ph.metrics.mu.Lock()
	defer ph.metrics.mu.Unlock()

	ph.metrics.ActiveConnections += delta
	if ph.metrics.ActiveConnections < 0 {
		ph.metrics.ActiveConnections = 0
	}
}

func (ph *ProtocolHandler) updateLatencyMetrics(msgType string, latency time.Duration) {
	ph.metrics.mu.Lock()
	defer ph.metrics.mu.Unlock()

	// Update overall average latency
	if ph.metrics.MessagesReceived > 0 {
		// Simple moving average
		totalLatency := ph.metrics.AverageLatency * time.Duration(ph.metrics.MessagesReceived-1)
		ph.metrics.AverageLatency = (totalLatency + latency) / time.Duration(ph.metrics.MessagesReceived)
	} else {
		ph.metrics.AverageLatency = latency
	}

	// Update message type latency
	if typeMetrics, exists := ph.metrics.MessageTypeMetrics[msgType]; exists {
		typeMetrics.TotalLatency += latency
	}
}

// GetMetrics returns a copy of the current metrics
func (ph *ProtocolHandler) GetMetrics() *ProtocolMetrics {
	ph.metrics.mu.RLock()
	defer ph.metrics.mu.RUnlock()

	// Create a deep copy of metrics
	metricsCopy := &ProtocolMetrics{
		MessagesReceived:   ph.metrics.MessagesReceived,
		MessagesSent:       ph.metrics.MessagesSent,
		BytesReceived:      ph.metrics.BytesReceived,
		BytesSent:          ph.metrics.BytesSent,
		ErrorCount:         ph.metrics.ErrorCount,
		AverageLatency:     ph.metrics.AverageLatency,
		LastActivity:       ph.metrics.LastActivity,
		ActiveConnections:  ph.metrics.ActiveConnections,
		MessageTypeMetrics: make(map[string]*MessageTypeMetrics),
	}

	// Copy message type metrics
	for msgType, typeMetrics := range ph.metrics.MessageTypeMetrics {
		metricsCopy.MessageTypeMetrics[msgType] = &MessageTypeMetrics{
			Count:        typeMetrics.Count,
			Errors:       typeMetrics.Errors,
			TotalLatency: typeMetrics.TotalLatency,
			LastSeen:     typeMetrics.LastSeen,
		}
	}

	return metricsCopy
}

// Utility functions

// generateMessageID generates a unique message ID
func generateMessageID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("msg-%d-%x", time.Now().UnixNano(), hash[:8])
}

// CreateResponseMessage creates a response message for a request
func CreateResponseMessage(requestMsg *Message, responseType string, data map[string]interface{}) *Message {
	return &Message{
		Type:      responseType,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data:      data,
		Metadata: map[string]string{
			"request_id": requestMsg.ID,
		},
	}
}

// CreateRequestMessage creates a new request message
func CreateRequestMessage(messageType string, data map[string]interface{}) *Message {
	return &Message{
		Type:      messageType,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data:      data,
	}
}

// StreamDialer provides an interface for creating streams
type StreamDialer interface {
	NewStream(ctx context.Context, peerID peer.ID, protocolID protocol.ID) (network.Stream, error)
}

// ProtocolClient provides client functionality for protocols
type ProtocolClient struct {
	dialer     StreamDialer
	protocolID protocol.ID
	timeout    time.Duration
	metrics    *ClientMetrics
}

// ClientMetrics tracks client-side protocol metrics
type ClientMetrics struct {
	RequestsSent      int64         `json:"requests_sent"`
	ResponsesReceived int64         `json:"responses_received"`
	Timeouts          int64         `json:"timeouts"`
	Errors            int64         `json:"errors"`
	AverageLatency    time.Duration `json:"average_latency"`

	mu sync.RWMutex
}

// NewProtocolClient creates a new protocol client
func NewProtocolClient(dialer StreamDialer, protocolID protocol.ID, timeout time.Duration) *ProtocolClient {
	return &ProtocolClient{
		dialer:     dialer,
		protocolID: protocolID,
		timeout:    timeout,
		metrics:    &ClientMetrics{},
	}
}

// SendRequest sends a request and waits for response
func (pc *ProtocolClient) SendRequest(ctx context.Context, peerID peer.ID, request *Message) (*Message, error) {
	start := time.Now()

	// Create stream
	stream, err := pc.dialer.NewStream(ctx, peerID, pc.protocolID)
	if err != nil {
		pc.updateErrorMetrics()
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Set deadline
	if pc.timeout > 0 {
		if conn, ok := stream.Conn().(interface{ SetDeadline(time.Time) error }); ok {
			conn.SetDeadline(time.Now().Add(pc.timeout))
		}
	}

	// Create protocol handler for sending/receiving
	handler := NewProtocolHandler(pc.protocolID)

	// Send request
	if err := handler.SendMessage(stream, request); err != nil {
		pc.updateErrorMetrics()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	pc.updateRequestMetrics()

	// Read response
	response, err := handler.readMessage(stream)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			pc.updateTimeoutMetrics()
			return nil, fmt.Errorf("request timeout: %w", err)
		}
		pc.updateErrorMetrics()
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	pc.updateResponseMetrics()
	pc.updateLatencyMetrics(time.Since(start))

	return response, nil
}

// SendOneWay sends a one-way message (no response expected)
func (pc *ProtocolClient) SendOneWay(ctx context.Context, peerID peer.ID, message *Message) error {
	stream, err := pc.dialer.NewStream(ctx, peerID, pc.protocolID)
	if err != nil {
		pc.updateErrorMetrics()
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Set deadline
	if pc.timeout > 0 {
		if conn, ok := stream.Conn().(interface{ SetDeadline(time.Time) error }); ok {
			conn.SetDeadline(time.Now().Add(pc.timeout))
		}
	}

	handler := NewProtocolHandler(pc.protocolID)

	if err := handler.SendMessage(stream, message); err != nil {
		pc.updateErrorMetrics()
		return fmt.Errorf("failed to send message: %w", err)
	}

	pc.updateRequestMetrics()
	return nil
}

// Client metrics update methods
func (pc *ProtocolClient) updateRequestMetrics() {
	pc.metrics.mu.Lock()
	defer pc.metrics.mu.Unlock()
	pc.metrics.RequestsSent++
}

func (pc *ProtocolClient) updateResponseMetrics() {
	pc.metrics.mu.Lock()
	defer pc.metrics.mu.Unlock()
	pc.metrics.ResponsesReceived++
}

func (pc *ProtocolClient) updateErrorMetrics() {
	pc.metrics.mu.Lock()
	defer pc.metrics.mu.Unlock()
	pc.metrics.Errors++
}

func (pc *ProtocolClient) updateTimeoutMetrics() {
	pc.metrics.mu.Lock()
	defer pc.metrics.mu.Unlock()
	pc.metrics.Timeouts++
}

func (pc *ProtocolClient) updateLatencyMetrics(latency time.Duration) {
	pc.metrics.mu.Lock()
	defer pc.metrics.mu.Unlock()

	if pc.metrics.ResponsesReceived > 0 {
		totalLatency := pc.metrics.AverageLatency * time.Duration(pc.metrics.ResponsesReceived-1)
		pc.metrics.AverageLatency = (totalLatency + latency) / time.Duration(pc.metrics.ResponsesReceived)
	} else {
		pc.metrics.AverageLatency = latency
	}
}

// GetClientMetrics returns a copy of client metrics
func (pc *ProtocolClient) GetClientMetrics() *ClientMetrics {
	pc.metrics.mu.RLock()
	defer pc.metrics.mu.RUnlock()

	return &ClientMetrics{
		RequestsSent:      pc.metrics.RequestsSent,
		ResponsesReceived: pc.metrics.ResponsesReceived,
		Timeouts:          pc.metrics.Timeouts,
		Errors:            pc.metrics.Errors,
		AverageLatency:    pc.metrics.AverageLatency,
	}
}
