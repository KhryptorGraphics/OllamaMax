package api

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSConnection represents a WebSocket connection
type WSConnection struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
	Roles    []string
	LastPing time.Time
	mu       sync.Mutex
}

// WSHub manages WebSocket connections
type WSHub struct {
	clients    map[string]*WSConnection
	broadcast  chan WSMessage
	register   chan *WSConnection
	unregister chan *WSConnection
	mu         sync.RWMutex
	running    bool
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WSMessageType constants
const (
	WSMsgTypeHeartbeat     = "heartbeat"
	WSMsgTypeNotification  = "notification"
	WSMsgTypeStatus        = "status"
	WSMsgTypeMetrics       = "metrics"
	WSMsgTypeTaskUpdate    = "task_update"
	WSMsgTypeNodeUpdate    = "node_update"
	WSMsgTypeModelUpdate   = "model_update"
	WSMsgTypeClusterUpdate = "cluster_update"
	WSMsgTypeError         = "error"
	WSMsgTypeAuth          = "auth"
	WSMsgTypeSubscribe     = "subscribe"
	WSMsgTypeUnsubscribe   = "unsubscribe"
)

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[string]*WSConnection),
		broadcast:  make(chan WSMessage, 256),
		register:   make(chan *WSConnection, 32),
		unregister: make(chan *WSConnection, 32),
	}
}

// Run starts the WebSocket hub
func (h *WSHub) Run() {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	// Start heartbeat routine
	go h.heartbeatRoutine()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// Stop stops the WebSocket hub
func (h *WSHub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.running = false

	// Close all connections
	for _, conn := range h.clients {
		conn.Conn.Close()
	}

	// Clear clients
	h.clients = make(map[string]*WSConnection)
}

// IsHealthy returns whether the WebSocket hub is healthy
func (h *WSHub) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.running
}

// GetClientCount returns the number of connected clients
func (h *WSHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast sends a message to all connected clients
func (h *WSHub) Broadcast(msgType string, data interface{}) {
	message := WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("WebSocket broadcast channel full, dropping message")
	}
}

// registerConnection registers a new WebSocket connection
func (h *WSHub) registerConnection(conn *WSConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove existing connection for this user if any
	if existing, exists := h.clients[conn.UserID]; exists {
		existing.Conn.Close()
	}

	h.clients[conn.UserID] = conn
	log.Printf("WebSocket client registered: %s (%s)", conn.Username, conn.UserID)

	// Send welcome message
	welcomeMsg := WSMessage{
		Type: WSMsgTypeStatus,
		Data: map[string]interface{}{
			"status":    "connected",
			"timestamp": time.Now(),
			"user_id":   conn.UserID,
			"username":  conn.Username,
		},
		Timestamp: time.Now(),
	}
	conn.sendMessage(welcomeMsg)
}

// unregisterConnection unregisters a WebSocket connection
func (h *WSHub) unregisterConnection(conn *WSConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[conn.UserID]; exists {
		delete(h.clients, conn.UserID)
		conn.Conn.Close()
		log.Printf("WebSocket client unregistered: %s (%s)", conn.Username, conn.UserID)
	}
}

// broadcastMessage broadcasts a message to all connected clients
func (h *WSHub) broadcastMessage(message WSMessage) {
	h.mu.RLock()
	clients := make([]*WSConnection, 0, len(h.clients))
	for _, conn := range h.clients {
		clients = append(clients, conn)
	}
	h.mu.RUnlock()

	// Send to all clients (or specific user if UserID is set)
	for _, conn := range clients {
		if message.UserID == "" || message.UserID == conn.UserID {
			// Check if user has permission to receive this message type
			if h.hasPermission(conn, message.Type) {
				conn.sendMessage(message)
			}
		}
	}
}

// hasPermission checks if a user has permission to receive a message type
func (h *WSHub) hasPermission(conn *WSConnection, msgType string) bool {
	// Admin users can receive all messages
	for _, role := range conn.Roles {
		if role == "admin" {
			return true
		}
	}

	// Regular users can receive most messages except sensitive ones
	switch msgType {
	case WSMsgTypeHeartbeat, WSMsgTypeNotification, WSMsgTypeStatus, WSMsgTypeTaskUpdate:
		return true
	case WSMsgTypeMetrics, WSMsgTypeNodeUpdate, WSMsgTypeClusterUpdate:
		// Only admin users can receive system metrics and updates
		return false
	default:
		return true
	}
}

// heartbeatRoutine sends periodic heartbeat messages
func (h *WSHub) heartbeatRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends heartbeat to all connected clients
func (h *WSHub) sendHeartbeat() {
	h.Broadcast(WSMsgTypeHeartbeat, map[string]interface{}{
		"timestamp":    time.Now(),
		"server_time":  time.Now().Unix(),
		"client_count": h.GetClientCount(),
	})
}

// sendMessage sends a message through the WebSocket connection
func (conn *WSConnection) sendMessage(message WSMessage) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	// Set write deadline
	conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err := conn.Conn.WriteJSON(message); err != nil {
		log.Printf("WebSocket write error for user %s: %v", conn.UserID, err)
		conn.Conn.Close()
	}
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// Authenticate the WebSocket connection
	wsConn, err := s.authenticateWebSocket(c, conn)
	if err != nil {
		s.logger.Error("WebSocket authentication failed", "error", err)
		conn.WriteMessage(websocket.CloseMessage, 
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Authentication failed"))
		conn.Close()
		return
	}

	// Register the connection
	s.wsHub.register <- wsConn

	// Store connection for cleanup
	s.wsConnections[wsConn.UserID] = wsConn

	// Start goroutines for reading and writing
	go s.handleWebSocketRead(wsConn)
	go s.handleWebSocketWrite(wsConn)
}

// authenticateWebSocket authenticates a WebSocket connection
func (s *Server) authenticateWebSocket(c *gin.Context, conn *websocket.Conn) (*WSConnection, error) {
	// Try to get token from query parameter or headers
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if token != "" && len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		return nil, fmt.Errorf("no authentication token provided")
	}

	// Validate the token
	claims, err := s.validateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	wsConn := &WSConnection{
		Conn:     conn,
		UserID:   claims.UserID,
		Username: claims.Username,
		Roles:    claims.Roles,
		LastPing: time.Now(),
	}

	s.logger.Info("WebSocket connection authenticated", 
		"user_id", claims.UserID, 
		"username", claims.Username)

	return wsConn, nil
}

// handleWebSocketRead handles reading messages from WebSocket connection
func (s *Server) handleWebSocketRead(wsConn *WSConnection) {
	defer func() {
		s.wsHub.unregister <- wsConn
		delete(s.wsConnections, wsConn.UserID)
	}()

	// Set read deadline and pong handler
	wsConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wsConn.Conn.SetPongHandler(func(string) error {
		wsConn.LastPing = time.Now()
		wsConn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WSMessage
		err := wsConn.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket read error", "user_id", wsConn.UserID, "error", err)
			}
			break
		}

		// Handle incoming message
		s.handleWebSocketMessage(wsConn, message)
	}
}

// handleWebSocketWrite handles writing messages to WebSocket connection
func (s *Server) handleWebSocketWrite(wsConn *WSConnection) {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send ping
			wsConn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := wsConn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.Error("WebSocket ping failed", "user_id", wsConn.UserID, "error", err)
				return
			}
		}
	}
}

// handleWebSocketMessage handles incoming WebSocket messages
func (s *Server) handleWebSocketMessage(wsConn *WSConnection, message WSMessage) {
	s.logger.Debug("WebSocket message received", 
		"user_id", wsConn.UserID, 
		"type", message.Type)

	switch message.Type {
	case WSMsgTypeAuth:
		// Re-authentication request
		s.handleWebSocketAuth(wsConn, message)

	case WSMsgTypeSubscribe:
		// Subscribe to specific events
		s.handleWebSocketSubscribe(wsConn, message)

	case WSMsgTypeUnsubscribe:
		// Unsubscribe from specific events
		s.handleWebSocketUnsubscribe(wsConn, message)

	case WSMsgTypeHeartbeat:
		// Client heartbeat response
		wsConn.LastPing = time.Now()

	default:
		s.logger.Warn("Unknown WebSocket message type", 
			"user_id", wsConn.UserID, 
			"type", message.Type)
	}
}

// handleWebSocketAuth handles authentication messages
func (s *Server) handleWebSocketAuth(wsConn *WSConnection, message WSMessage) {
	// For now, just send success response
	response := WSMessage{
		Type: WSMsgTypeAuth,
		Data: map[string]interface{}{
			"status":    "authenticated",
			"user_id":   wsConn.UserID,
			"username":  wsConn.Username,
			"roles":     wsConn.Roles,
		},
		Timestamp: time.Now(),
	}
	wsConn.sendMessage(response)
}

// handleWebSocketSubscribe handles subscription messages
func (s *Server) handleWebSocketSubscribe(wsConn *WSConnection, message WSMessage) {
	// Parse subscription data
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		return
	}

	events, ok := data["events"].([]interface{})
	if !ok {
		return
	}

	// Store subscription preferences (in a real implementation)
	s.logger.Info("WebSocket subscription", 
		"user_id", wsConn.UserID, 
		"events", events)

	// Send confirmation
	response := WSMessage{
		Type: WSMsgTypeSubscribe,
		Data: map[string]interface{}{
			"status": "subscribed",
			"events": events,
		},
		Timestamp: time.Now(),
	}
	wsConn.sendMessage(response)
}

// handleWebSocketUnsubscribe handles unsubscription messages
func (s *Server) handleWebSocketUnsubscribe(wsConn *WSConnection, message WSMessage) {
	// Parse unsubscription data
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		return
	}

	events, ok := data["events"].([]interface{})
	if !ok {
		return
	}

	// Remove subscription preferences (in a real implementation)
	s.logger.Info("WebSocket unsubscription", 
		"user_id", wsConn.UserID, 
		"events", events)

	// Send confirmation
	response := WSMessage{
		Type: WSMsgTypeUnsubscribe,
		Data: map[string]interface{}{
			"status": "unsubscribed",
			"events": events,
		},
		Timestamp: time.Now(),
	}
	wsConn.sendMessage(response)
}

// WebSocket notification methods

// NotifyModelUpdate sends model update notifications
func (s *Server) NotifyModelUpdate(modelName, action string, data interface{}) {
	s.wsHub.Broadcast(WSMsgTypeModelUpdate, map[string]interface{}{
		"model":  modelName,
		"action": action,
		"data":   data,
	})
}

// NotifyNodeUpdate sends node update notifications
func (s *Server) NotifyNodeUpdate(nodeID, action string, data interface{}) {
	s.wsHub.Broadcast(WSMsgTypeNodeUpdate, map[string]interface{}{
		"node_id": nodeID,
		"action":  action,
		"data":    data,
	})
}

// NotifyTaskUpdate sends task update notifications
func (s *Server) NotifyTaskUpdate(taskID, status string, data interface{}) {
	s.wsHub.Broadcast(WSMsgTypeTaskUpdate, map[string]interface{}{
		"task_id": taskID,
		"status":  status,
		"data":    data,
	})
}

// BroadcastMetrics sends system metrics to all connected clients
func (s *Server) BroadcastMetrics(metrics map[string]interface{}) {
	s.wsHub.Broadcast(WSMsgTypeMetrics, metrics)
}