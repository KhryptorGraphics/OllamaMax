package api

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WSHub manages WebSocket connections
type WSHub struct {
	clients    map[*WSConnection]bool
	broadcast  chan []byte
	register   chan *WSConnection
	unregister chan *WSConnection
	mutex      sync.RWMutex
}

// WSConnection represents a WebSocket connection
type WSConnection struct {
	conn   *websocket.Conn
	send   chan []byte
	hub    *WSHub
	userID string
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id,omitempty"`
}

// Dashboard update message types
const (
	MsgTypeClusterUpdate   = "cluster_update"
	MsgTypeNodeUpdate      = "node_update"
	MsgTypeMetricsUpdate   = "metrics_update"
	MsgTypeModelUpdate     = "model_update"
	MsgTypeInferenceUpdate = "inference_update"
	MsgTypeNotification    = "notification"
	MsgTypeHeartbeat       = "heartbeat"
	MsgTypeSubscribe       = "subscribe"
	MsgTypeUnsubscribe     = "unsubscribe"
)

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSConnection]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSConnection),
		unregister: make(chan *WSConnection),
	}
}

// Run starts the WebSocket hub
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

			// Send welcome message
			welcome := WSMessage{
				Type:      "welcome",
				Data:      map[string]string{"status": "connected"},
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(welcome); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					h.mutex.Lock()
					delete(h.clients, client)
					h.mutex.Unlock()
				}
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *WSHub) Broadcast(msgType string, data interface{}) {
	message := WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}

	if jsonData, err := json.Marshal(message); err == nil {
		select {
		case h.broadcast <- jsonData:
		default:
			log.Printf("WebSocket broadcast channel full, dropping message")
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *WSHub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Get user ID from context or token
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &WSConnection{
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    s.wsHub,
		userID: userID,
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump handles reading from the WebSocket connection
func (c *WSConnection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err == nil {
			c.handleMessage(&wsMsg)
		}
	}
}

// writePump handles writing to the WebSocket connection
func (c *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *WSConnection) handleMessage(msg *WSMessage) {
	switch msg.Type {
	case "ping":
		response := WSMessage{
			Type:      "pong",
			Data:      map[string]string{"status": "ok"},
			Timestamp: time.Now(),
		}
		if data, err := json.Marshal(response); err == nil {
			select {
			case c.send <- data:
			default:
				close(c.send)
			}
		}

	case "subscribe":
		// Handle subscription requests
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if topic, exists := data["topic"].(string); exists {
				log.Printf("Client %s subscribed to topic: %s", c.userID, topic)
				// TODO: Implement topic-based subscriptions
			}
		}

	case "unsubscribe":
		// Handle unsubscription requests
		if data, ok := msg.Data.(map[string]interface{}); ok {
			if topic, exists := data["topic"].(string); exists {
				log.Printf("Client %s unsubscribed from topic: %s", c.userID, topic)
				// TODO: Implement topic-based unsubscriptions
			}
		}

	default:
		log.Printf("Unknown WebSocket message type: %s", msg.Type)
	}
}

// BroadcastNodeUpdate broadcasts node status updates
func (s *Server) BroadcastNodeUpdate(nodeID string, status string) {
	s.wsHub.Broadcast("node_update", map[string]interface{}{
		"node_id":   nodeID,
		"status":    status,
		"timestamp": time.Now(),
	})
}

// BroadcastModelUpdate broadcasts model status updates
func (s *Server) BroadcastModelUpdate(modelName string, status string, progress float64) {
	s.wsHub.Broadcast("model_update", map[string]interface{}{
		"model_name": modelName,
		"status":     status,
		"progress":   progress,
		"timestamp":  time.Now(),
	})
}

// BroadcastMetricsUpdate broadcasts system metrics updates
func (s *Server) BroadcastMetricsUpdate(metrics map[string]interface{}) {
	s.wsHub.Broadcast("metrics_update", map[string]interface{}{
		"metrics":   metrics,
		"timestamp": time.Now(),
	})
}

// BroadcastDashboardUpdate broadcasts comprehensive dashboard updates
func (s *Server) BroadcastDashboardUpdate(data map[string]interface{}) {
	s.wsHub.Broadcast(MsgTypeClusterUpdate, map[string]interface{}{
		"dashboard": data,
		"timestamp": time.Now(),
	})
}

// BroadcastNotification broadcasts notifications to users
func (s *Server) BroadcastNotification(notification map[string]interface{}) {
	s.wsHub.Broadcast(MsgTypeNotification, map[string]interface{}{
		"notification": notification,
		"timestamp":    time.Now(),
	})
}

// BroadcastInferenceUpdate broadcasts inference request updates
func (s *Server) BroadcastInferenceUpdate(requestID string, status string, progress float64, result interface{}) {
	s.wsHub.Broadcast(MsgTypeInferenceUpdate, map[string]interface{}{
		"request_id": requestID,
		"status":     status,
		"progress":   progress,
		"result":     result,
		"timestamp":  time.Now(),
	})
}

// SendHeartbeat sends periodic heartbeat to maintain connections
func (s *Server) SendHeartbeat() {
	s.wsHub.Broadcast(MsgTypeHeartbeat, map[string]interface{}{
		"timestamp":   time.Now(),
		"server_time": time.Now().Unix(),
	})
}

// StartDashboardUpdates starts periodic dashboard updates
func (s *Server) StartDashboardUpdates() {
	ticker := time.NewTicker(5 * time.Second) // Update every 5 seconds
	go func() {
		for range ticker.C {
			// Get current dashboard data
			clusterSize := 0
			if s.scheduler != nil {
				clusterSize = len(s.scheduler.GetAvailableNodes())
			}

			isLeader := false
			if s.consensus != nil {
				isLeader = s.consensus.IsLeader()
			}

			// Broadcast real-time updates
			s.BroadcastDashboardUpdate(map[string]interface{}{
				"clusterStatus": map[string]interface{}{
					"healthy":   true,
					"size":      clusterSize,
					"leader":    isLeader,
					"consensus": s.consensus != nil,
				},
				"nodeCount":    clusterSize,
				"activeModels": 3, // TODO: Get from model manager
				"timestamp":    time.Now().UTC(),
			})
		}
	}()

	// Send heartbeat every 30 seconds
	heartbeatTicker := time.NewTicker(30 * time.Second)
	go func() {
		for range heartbeatTicker.C {
			s.SendHeartbeat()
		}
	}()
}
