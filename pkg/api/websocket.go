package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket message types
const (
	MessageTypeHeartbeat      = "heartbeat"
	MessageTypeNodeStatus     = "node_status"
	MessageTypeModelUpdate    = "model_update"
	MessageTypeInference      = "inference"
	MessageTypeSystemMetrics  = "system_metrics"
	MessageTypeError          = "error"
	MessageTypeSubscribe      = "subscribe"
	MessageTypeUnsubscribe    = "unsubscribe"
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	ID        string      `json:"id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID           string
	Conn         *websocket.Conn
	Send         chan WebSocketMessage
	Hub          *WebSocketHub
	Subscriptions map[string]bool
	UserID       *uuid.UUID
	mu           sync.RWMutex
}

// WebSocketHub maintains WebSocket connections and handles broadcasting
type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	logger     *slog.Logger
	mu         sync.RWMutex
}

// WebSocket upgrader with proper configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (configure for production)
		return true
	},
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(logger *slog.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		logger:     logger,
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	h.logger.Info("WebSocket hub started")

	// Start heartbeat ticker
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Info("WebSocket client connected", "client_id", client.ID)

			// Send welcome message
			client.Send <- WebSocketMessage{
				Type:      "welcome",
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"client_id": client.ID,
					"message":   "Connected to OllamaMax WebSocket",
				},
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			h.logger.Info("WebSocket client disconnected", "client_id", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, disconnect
					delete(h.clients, client)
					close(client.Send)
				}
			}
			h.mu.RUnlock()

		case <-heartbeat.C:
			// Send heartbeat to all clients
			heartbeatMsg := WebSocketMessage{
				Type:      MessageTypeHeartbeat,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"status": "alive",
				},
			}
			h.BroadcastToSubscribers(heartbeatMsg, MessageTypeHeartbeat)
		}
	}
}

// Stop gracefully stops the WebSocket hub
func (h *WebSocketHub) Stop() {
	h.logger.Info("Stopping WebSocket hub")
	h.mu.Lock()
	for client := range h.clients {
		client.Conn.Close()
		close(client.Send)
		delete(h.clients, client)
	}
	h.mu.Unlock()
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message WebSocketMessage) {
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("Broadcast channel full, dropping message")
	}
}

// BroadcastToSubscribers sends a message to clients subscribed to a specific type
func (h *WebSocketHub) BroadcastToSubscribers(message WebSocketMessage, messageType string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		client.mu.RLock()
		if client.Subscriptions[messageType] {
			select {
			case client.Send <- message:
			default:
				// Client's send channel is full, skip
			}
		}
		client.mu.RUnlock()
	}
}

// GetConnectedClients returns the number of connected clients
func (h *WebSocketHub) GetConnectedClients() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// WebSocket handler for general connections
func (s *Server) websocketHandler(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade WebSocket connection", "error", err)
		return
	}

	// Create new client
	client := &WebSocketClient{
		ID:            uuid.New().String(),
		Conn:          conn,
		Send:          make(chan WebSocketMessage, 256),
		Hub:           s.websocket,
		Subscriptions: make(map[string]bool),
	}

	// Get user ID if authenticated
	if userID, exists := c.Get("user_id"); exists {
		if uid, err := uuid.Parse(userID.(string)); err == nil {
			client.UserID = &uid
		}
	}

	// Register client
	s.websocket.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(s)
}

// WebSocket handler for inference-specific connections
func (s *Server) inferenceWebsocketHandler(c *gin.Context) {
	inferenceID := c.Param("id")
	if inferenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing_inference_id",
		})
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade inference WebSocket", "error", err)
		return
	}

	// Create client with inference subscription
	client := &WebSocketClient{
		ID:   uuid.New().String(),
		Conn: conn,
		Send: make(chan WebSocketMessage, 256),
		Hub:  s.websocket,
		Subscriptions: map[string]bool{
			MessageTypeInference: true,
			"inference_" + inferenceID: true,
		},
	}

	// Register and start
	s.websocket.register <- client
	go client.writePump()
	go client.readPump(s)
}

// readPump handles reading messages from the WebSocket connection
func (c *WebSocketClient) readPump(s *Server) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read limits and timeouts
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket read error", "error", err, "client_id", c.ID)
			}
			break
		}

		// Handle different message types
		switch message.Type {
		case MessageTypeSubscribe:
			c.handleSubscribe(message, s)
		case MessageTypeUnsubscribe:
			c.handleUnsubscribe(message, s)
		case MessageTypeHeartbeat:
			// Echo back heartbeat
			c.Send <- WebSocketMessage{
				Type:      MessageTypeHeartbeat,
				Timestamp: time.Now(),
				Data:      map[string]interface{}{"status": "pong"},
			}
		default:
			s.logger.Warn("Unknown WebSocket message type", "type", message.Type, "client_id", c.ID)
		}
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe processes subscription requests
func (c *WebSocketClient) handleSubscribe(message WebSocketMessage, s *Server) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.Send <- WebSocketMessage{
			Type:      MessageTypeError,
			Timestamp: time.Now(),
			Error:     "Invalid subscription data format",
		}
		return
	}

	topics, ok := data["topics"].([]interface{})
	if !ok {
		c.Send <- WebSocketMessage{
			Type:      MessageTypeError,
			Timestamp: time.Now(),
			Error:     "Invalid topics format",
		}
		return
	}

	c.mu.Lock()
	for _, topic := range topics {
		if topicStr, ok := topic.(string); ok {
			c.Subscriptions[topicStr] = true
			s.logger.Info("Client subscribed to topic", "client_id", c.ID, "topic", topicStr)
		}
	}
	c.mu.Unlock()

	c.Send <- WebSocketMessage{
		Type:      "subscription_confirmed",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"subscribed_topics": topics,
		},
	}
}

// handleUnsubscribe processes unsubscription requests
func (c *WebSocketClient) handleUnsubscribe(message WebSocketMessage, s *Server) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		c.Send <- WebSocketMessage{
			Type:      MessageTypeError,
			Timestamp: time.Now(),
			Error:     "Invalid unsubscription data format",
		}
		return
	}

	topics, ok := data["topics"].([]interface{})
	if !ok {
		c.Send <- WebSocketMessage{
			Type:      MessageTypeError,
			Timestamp: time.Now(),
			Error:     "Invalid topics format",
		}
		return
	}

	c.mu.Lock()
	for _, topic := range topics {
		if topicStr, ok := topic.(string); ok {
			delete(c.Subscriptions, topicStr)
			s.logger.Info("Client unsubscribed from topic", "client_id", c.ID, "topic", topicStr)
		}
	}
	c.mu.Unlock()

	c.Send <- WebSocketMessage{
		Type:      "unsubscription_confirmed",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"unsubscribed_topics": topics,
		},
	}
}

// Helper methods for broadcasting specific types of messages

// BroadcastNodeStatus broadcasts node status updates
func (h *WebSocketHub) BroadcastNodeStatus(nodeID uuid.UUID, status string, data interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeNodeStatus,
		ID:        nodeID.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"node_id": nodeID,
			"status":  status,
			"details": data,
		},
	}
	h.BroadcastToSubscribers(message, MessageTypeNodeStatus)
}

// BroadcastModelUpdate broadcasts model update notifications
func (h *WebSocketHub) BroadcastModelUpdate(modelID uuid.UUID, action string, data interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeModelUpdate,
		ID:        modelID.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"model_id": modelID,
			"action":   action,
			"details":  data,
		},
	}
	h.BroadcastToSubscribers(message, MessageTypeModelUpdate)
}

// BroadcastInferenceUpdate broadcasts inference progress updates
func (h *WebSocketHub) BroadcastInferenceUpdate(inferenceID uuid.UUID, status string, progress interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeInference,
		ID:        inferenceID.String(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"inference_id": inferenceID,
			"status":       status,
			"progress":     progress,
		},
	}
	h.BroadcastToSubscribers(message, MessageTypeInference)
	h.BroadcastToSubscribers(message, "inference_"+inferenceID.String())
}

// BroadcastSystemMetrics broadcasts system performance metrics
func (h *WebSocketHub) BroadcastSystemMetrics(metrics interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeSystemMetrics,
		Timestamp: time.Now(),
		Data:      metrics,
	}
	h.BroadcastToSubscribers(message, MessageTypeSystemMetrics)
}
