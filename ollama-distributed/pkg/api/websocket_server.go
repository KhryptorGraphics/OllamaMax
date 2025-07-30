package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketConfig configures the WebSocket server
type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	CheckOrigin     func(r *http.Request) bool
}

// WebSocketMetrics tracks WebSocket server performance
type WebSocketMetrics struct {
	ConnectionsActive int64     `json:"connections_active"`
	ConnectionsTotal  int64     `json:"connections_total"`
	MessagesReceived  int64     `json:"messages_received"`
	MessagesSent      int64     `json:"messages_sent"`
	MessageErrors     int64     `json:"message_errors"`
	LastUpdated       time.Time `json:"last_updated"`
	mu                sync.RWMutex
}

// WSConnection represents a WebSocket connection
type WSConnection struct {
	ID          string
	Conn        *websocket.Conn
	Send        chan []byte
	Hub         *WSHub
	ConnectedAt time.Time
	LastPing    time.Time
	UserID      string
	Metadata    map[string]interface{}
	mu          sync.RWMutex
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	From      string                 `json:"from,omitempty"`
	To        string                 `json:"to,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(config *WebSocketConfig) (*WebSocketServer, error) {
	if config == nil {
		config = &WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     config.CheckOrigin,
	}

	hub := &WSHub{
		clients:    make(map[*WSConnection]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WSConnection),
		unregister: make(chan *WSConnection),
		rooms:      make(map[string]map[*WSConnection]bool),
	}

	server := &WebSocketServer{
		config:      config,
		upgrader:    upgrader,
		hub:         hub,
		connections: make(map[string]*WSConnection),
		metrics: &WebSocketMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return server, nil
}

// Start starts the WebSocket server
func (ws *WebSocketServer) Start() error {
	// Start the hub
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		ws.hub.run()
	}()

	// Start metrics collection
	ws.wg.Add(1)
	go ws.metricsLoop()

	return nil
}

// Stop stops the WebSocket server
func (ws *WebSocketServer) Stop() error {
	ws.cancel()

	// Close all connections
	ws.connectionsMu.Lock()
	for _, conn := range ws.connections {
		conn.Close()
	}
	ws.connectionsMu.Unlock()

	ws.wg.Wait()
	return nil
}

// HandleUpgrade handles WebSocket upgrade requests
func (ws *WebSocketServer) HandleUpgrade(c *gin.Context) {
	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}

	// Create WebSocket connection
	wsConn := &WSConnection{
		ID:          generateConnectionID(),
		Conn:        conn,
		Send:        make(chan []byte, 256),
		Hub:         ws.hub,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Register connection
	ws.connectionsMu.Lock()
	ws.connections[wsConn.ID] = wsConn
	ws.connectionsMu.Unlock()

	// Update metrics
	ws.metrics.mu.Lock()
	ws.metrics.ConnectionsActive++
	ws.metrics.ConnectionsTotal++
	ws.metrics.LastUpdated = time.Now()
	ws.metrics.mu.Unlock()

	// Register with hub
	ws.hub.register <- wsConn

	// Start connection handlers
	go wsConn.writePump()
	go wsConn.readPump(ws)
}

// Broadcast sends a message to all connected clients
func (ws *WebSocketServer) Broadcast(message []byte) {
	ws.hub.broadcast <- message
}

// SendToConnection sends a message to a specific connection
func (ws *WebSocketServer) SendToConnection(connectionID string, message []byte) error {
	ws.connectionsMu.RLock()
	conn, exists := ws.connections[connectionID]
	ws.connectionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found")
	}

	// Send message through WSConnection
	select {
	case conn.Send <- message:
		return nil
	default:
		return fmt.Errorf("connection send buffer full")
	}
}

// SendToRoom sends a message to all connections in a room
func (ws *WebSocketServer) SendToRoom(room string, message []byte) {
	ws.hub.broadcastToRoom(room, message)
}

// JoinRoom adds a connection to a room
func (ws *WebSocketServer) JoinRoom(connectionID, room string) error {
	ws.connectionsMu.RLock()
	conn, exists := ws.connections[connectionID]
	ws.connectionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found")
	}

	ws.hub.joinRoom(conn, room)
	return nil
}

// LeaveRoom removes a connection from a room
func (ws *WebSocketServer) LeaveRoom(connectionID, room string) error {
	ws.connectionsMu.RLock()
	conn, exists := ws.connections[connectionID]
	ws.connectionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found")
	}

	ws.hub.leaveRoom(conn, room)
	return nil
}

// GetConnections returns all active connections
func (ws *WebSocketServer) GetConnections() map[string]*WSConnection {
	ws.connectionsMu.RLock()
	defer ws.connectionsMu.RUnlock()

	connections := make(map[string]*WSConnection)
	for id, conn := range ws.connections {
		connections[id] = conn
	}
	return connections
}

// GetMetrics returns WebSocket server metrics
func (ws *WebSocketServer) GetMetrics() *WebSocketMetrics {
	ws.metrics.mu.RLock()
	defer ws.metrics.mu.RUnlock()

	// Create a copy
	metrics := *ws.metrics
	return &metrics
}

// metricsLoop runs the metrics collection loop
func (ws *WebSocketServer) metricsLoop() {
	defer ws.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.ctx.Done():
			return
		case <-ticker.C:
			ws.updateMetrics()
		}
	}
}

// updateMetrics updates WebSocket metrics
func (ws *WebSocketServer) updateMetrics() {
	ws.metrics.mu.Lock()
	defer ws.metrics.mu.Unlock()

	ws.connectionsMu.RLock()
	ws.metrics.ConnectionsActive = int64(len(ws.connections))
	ws.connectionsMu.RUnlock()

	ws.metrics.LastUpdated = time.Now()
}

// WSConnection methods

// Close closes the WebSocket connection
func (c *WSConnection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn != nil {
		c.Conn.Close()
		close(c.Send)
	}
}

// readPump handles reading from the WebSocket connection
func (c *WSConnection) readPump(server *WebSocketServer) {
	defer func() {
		c.Hub.unregister <- c
		c.Close()

		// Remove from server connections
		server.connectionsMu.Lock()
		delete(server.connections, c.ID)
		server.connectionsMu.Unlock()

		// Update metrics
		server.metrics.mu.Lock()
		server.metrics.ConnectionsActive--
		server.metrics.LastUpdated = time.Now()
		server.metrics.mu.Unlock()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				server.metrics.mu.Lock()
				server.metrics.MessageErrors++
				server.metrics.mu.Unlock()
			}
			break
		}

		// Update metrics
		server.metrics.mu.Lock()
		server.metrics.MessagesReceived++
		server.metrics.LastUpdated = time.Now()
		server.metrics.mu.Unlock()

		// Process message (placeholder)
		_ = message
	}
}

// writePump handles writing to the WebSocket connection
func (c *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}
