package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// Server represents the API server
type Server struct {
	config    *config.APIConfig
	p2p       *p2p.Node
	consensus *consensus.Engine
	scheduler *scheduler.Engine
	
	router   *gin.Engine
	server   *http.Server
	upgrader websocket.Upgrader
	
	// WebSocket connections
	wsConnections map[string]*websocket.Conn
	wsHub         *WSHub
}

// WSHub manages WebSocket connections
type WSHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewServer creates a new API server
func NewServer(config *config.APIConfig, p2pNode *p2p.Node, consensusEngine *consensus.Engine, schedulerEngine *scheduler.Engine) (*Server, error) {
	server := &Server{
		config:        config,
		p2p:           p2pNode,
		consensus:     consensusEngine,
		scheduler:     schedulerEngine,
		wsConnections: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Check if origin is in allowed list
			allowedOrigins := []string{"http://localhost:8080", "https://localhost:8080"}
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
			},
		},
		wsHub: &WSHub{
			clients:    make(map[*websocket.Conn]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *websocket.Conn),
			unregister: make(chan *websocket.Conn),
		},
	}
	
	server.setupRoutes()
	
	return server, nil
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}
	
	s.router = gin.New()
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.rateLimitMiddleware())
	s.router.Use(s.inputValidationMiddleware())
	
	// CORS middleware with security-first configuration
	s.router.Use(func(c *gin.Context) {
		// Get origin from config CORS settings
		allowedOrigins := s.config.Cors.AllowedOrigins
		if len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
			// Default to secure localhost for development if wildcard is configured
			allowedOrigins = []string{"http://localhost:8080", "https://localhost:8080"}
		}
		
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "")
		}
		
		c.Header("Access-Control-Allow-Methods", strings.Join(s.config.Cors.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(s.config.Cors.AllowedHeaders, ", "))
		if s.config.Cors.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", s.config.Cors.MaxAge))
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// API routes with authentication
	api := s.router.Group("/api/v1")
	api.Use(s.authMiddleware())
	{
		// Node management
		api.GET("/nodes", s.getNodes)
		api.GET("/nodes/:id", s.getNode)
		api.POST("/nodes/:id/drain", s.drainNode)
		api.POST("/nodes/:id/undrain", s.undrainNode)
		
		// Model management
		api.GET("/models", s.getModels)
		api.GET("/models/:name", s.getModel)
		api.POST("/models/:name/download", s.downloadModel)
		api.DELETE("/models/:name", s.deleteModel)
		
		// Distribution management
		api.POST("/distribution/auto-configure", s.handleAutoDistribution)
		
		// Cluster management
		api.GET("/cluster/status", s.getClusterStatus)
		api.GET("/cluster/leader", s.getLeader)
		api.POST("/cluster/join", s.joinCluster)
		api.POST("/cluster/leave", s.leaveCluster)
		
		// Inference
		api.POST("/generate", s.generate)
		api.POST("/chat", s.chat)
		api.POST("/embeddings", s.embeddings)
		
		// Monitoring
		api.GET("/metrics", s.getMetrics)
		api.GET("/health", s.healthCheck)
		api.GET("/transfers", s.getTransfers)
		api.GET("/transfers/:id", s.getTransfer)
		
		// WebSocket endpoint
		api.GET("/ws", s.handleWebSocket)
	}
	
	// Serve static files for web UI
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/index.html")
	s.router.StaticFile("/favicon.ico", "./web/favicon.ico")
	
	// Catch-all for SPA routing
	s.router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})
}

// Start starts the API server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:    s.config.Listen,
		Handler: s.router,
	}
	
	// Start WebSocket hub
	go s.wsHub.run()
	
	// Start metrics broadcasting
	s.StartMetricsBroadcasting()
	
	// Start server
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()
	
	fmt.Printf("API server started on %s\n", s.config.Listen)
	return nil
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Node management handlers

func (s *Server) getNodes(c *gin.Context) {
	nodes := s.scheduler.GetNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func (s *Server) getNode(c *gin.Context) {
	nodeID := c.Param("id")
	
	nodes := s.scheduler.GetNodes()
	if node, exists := nodes[nodeID]; exists {
		c.JSON(http.StatusOK, gin.H{"node": node})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
	}
}

func (s *Server) drainNode(c *gin.Context) {
	nodeID := c.Param("id")
	
	// TODO: Implement node draining
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node %s is being drained", nodeID)})
}

func (s *Server) undrainNode(c *gin.Context) {
	nodeID := c.Param("id")
	
	// TODO: Implement node undraining
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Node %s is no longer draining", nodeID)})
}

// Model management handlers

func (s *Server) getModels(c *gin.Context) {
	// Get models from scheduler
	models := s.scheduler.GetAllModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (s *Server) getModel(c *gin.Context) {
	modelName := c.Param("name")
	
	// Get specific model from scheduler
	if model, exists := s.scheduler.GetModel(modelName); exists {
		c.JSON(http.StatusOK, gin.H{"model": model})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
	}
}

func (s *Server) downloadModel(c *gin.Context) {
	modelName := c.Param("name")
	
	// Initiate model download via scheduler
	nodes := s.scheduler.GetAvailableNodes()
	if len(nodes) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available nodes for model download"})
		return
	}
	
	// Select the best node for download (first available for now)
	targetNode := nodes[0]
	
	// Register the model as being downloaded
	err := s.scheduler.RegisterModel(modelName, 0, "", targetNode.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register model: %v", err)})
		return
	}
	
	// TODO: Implement actual model download to node
	// For now, just simulate the download process
	go func() {
		time.Sleep(2 * time.Second) // Simulate download time
		// Update model status in scheduler
		s.scheduler.RegisterModel(modelName, 1024*1024*100, "simulated_checksum", targetNode.ID)
	}()
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Started downloading model %s to node %s", modelName, targetNode.ID),
		"target_node": targetNode.ID,
	})
}

func (s *Server) deleteModel(c *gin.Context) {
	modelName := c.Param("name")
	
	// Get model info from scheduler
	model, exists := s.scheduler.GetModel(modelName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}
	
	// Delete model from scheduler registry
	err := s.scheduler.DeleteModel(modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Send deletion commands to nodes that have the model
	// For now, just remove from registry
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Deleted model %s from %d nodes", modelName, len(model.Locations)),
		"nodes_affected": model.Locations,
	})
}

// Cluster management handlers

func (s *Server) getClusterStatus(c *gin.Context) {
	status := gin.H{
		"node_id":   s.p2p.ID().String(),
		"is_leader": s.consensus.IsLeader(),
		"leader":    s.consensus.Leader(),
		"peers":     len(s.p2p.ConnectedPeers()),
		"status":    "healthy",
	}
	
	c.JSON(http.StatusOK, status)
}

func (s *Server) getLeader(c *gin.Context) {
	leader := s.consensus.Leader()
	c.JSON(http.StatusOK, gin.H{"leader": leader})
}

func (s *Server) joinCluster(c *gin.Context) {
	var req struct {
		NodeID  string `json:"node_id"`
		Address string `json:"address"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := s.consensus.AddVoter(req.NodeID, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Node joined cluster"})
}

func (s *Server) leaveCluster(c *gin.Context) {
	var req struct {
		NodeID string `json:"node_id"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := s.consensus.RemoveServer(req.NodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Node left cluster"})
}

// Inference handlers

func (s *Server) generate(c *gin.Context) {
	var req struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create scheduler request
	schedReq := &scheduler.Request{
		ID:         fmt.Sprintf("req_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
	}
	
	// Schedule the request
	if err := s.scheduler.Schedule(schedReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Wait for response
	select {
	case response := <-schedReq.ResponseCh:
		if response.Success {
			c.JSON(http.StatusOK, gin.H{
				"response": "Generated response would be here",
				"model":    req.Model,
				"node_id":  response.NodeID,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

func (s *Server) chat(c *gin.Context) {
	var req struct {
		Model    string                   `json:"model"`
		Messages []map[string]interface{} `json:"messages"`
		Stream   bool                     `json:"stream,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement chat completion
	c.JSON(http.StatusOK, gin.H{
		"message": gin.H{
			"role":    "assistant",
			"content": "This is a placeholder response from the distributed Ollama system",
		},
	})
}

func (s *Server) embeddings(c *gin.Context) {
	var req struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: Implement embeddings
	c.JSON(http.StatusOK, gin.H{
		"embeddings": []float64{0.1, 0.2, 0.3, 0.4, 0.5},
	})
}

// Monitoring handlers

func (s *Server) getMetrics(c *gin.Context) {
	// Get real metrics from scheduler
	modelsCount := s.scheduler.GetModelCount()
	nodesCount := len(s.scheduler.GetNodes())
	onlineNodes := s.scheduler.GetOnlineNodeCount()
	
	metrics := gin.H{
		"node_id":           s.p2p.ID().String(),
		"connected_peers":   len(s.p2p.ConnectedPeers()),
		"is_leader":         s.consensus.IsLeader(),
		"requests_processed": 0, // TODO: Add request counter to scheduler
		"models_loaded":     modelsCount,
		"nodes_total":       nodesCount,
		"nodes_online":      onlineNodes,
		"uptime":            time.Since(time.Now()).String(),
		"cpu_usage":         15.5,  // Mock data
		"memory_usage":      23.8,  // Mock data
		"network_usage":     45.2,  // Mock data
		"requests_per_second": 12,   // Mock data
		"average_latency":   125,   // Mock data
		"active_connections": 8,    // Mock data
		"error_rate":        0.2,   // Mock data
	}
	
	c.JSON(http.StatusOK, metrics)
}

func (s *Server) healthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"node_id":   s.p2p.ID().String(),
		"services": gin.H{
			"p2p":       "healthy",
			"consensus": "healthy",
			"scheduler": "healthy",
		},
	}
	
	c.JSON(http.StatusOK, health)
}

func (s *Server) getTransfers(c *gin.Context) {
	// Get transfer information from scheduler
	// For now, return mock transfer data based on models
	models := s.scheduler.GetAllModels()
	
	transfers := []map[string]interface{}{}
	for modelName, model := range models {
		for _, nodeID := range model.Locations {
			transfers = append(transfers, map[string]interface{}{
				"id": fmt.Sprintf("%s-%s", modelName, nodeID[:8]),
				"model_name": modelName,
				"type": "download",
				"status": "completed",
				"progress": 100.0,
				"speed": 0.0,
				"eta": 0,
				"peer_id": nodeID,
				"completed_at": model.LastAccessed,
			})
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}

func (s *Server) getTransfer(c *gin.Context) {
	transferID := c.Param("id")
	
	// TODO: Get specific transfer
	c.JSON(http.StatusOK, gin.H{"transfer": gin.H{"id": transferID}})
}

// Distribution management handlers

func (s *Server) handleAutoDistribution(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Store auto-distribution setting in consensus
	if err := s.consensus.Apply("auto_distribution_enabled", req.Enabled, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update auto-distribution setting: %v", err)})
		return
	}
	
	// If enabling auto-distribution, trigger redistribution
	if req.Enabled {
		go s.triggerModelRedistribution()
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Auto-distribution %s", map[bool]string{true: "enabled", false: "disabled"}[req.Enabled]),
		"enabled": req.Enabled,
	})
}

// triggerModelRedistribution triggers redistribution of models across nodes
func (s *Server) triggerModelRedistribution() {
	// Get all available nodes
	nodes := s.scheduler.GetAvailableNodes()
	if len(nodes) < 2 {
		return // Need at least 2 nodes for redistribution
	}
	
	// Get all models
	models := s.scheduler.GetAllModels()
	
	// Redistribute models for better load balancing
	for modelName, model := range models {
		// If model is only on one node, try to replicate it
		if len(model.Locations) == 1 {
			// Find a node that doesn't have this model
			for _, node := range nodes {
				if !contains(model.Locations, node.ID) {
					// Replicate model to this node
					go s.replicateModel(modelName, node.ID)
					break
				}
			}
		}
	}
}

// replicateModel replicates a model to a specific node
func (s *Server) replicateModel(modelName, nodeID string) {
	// TODO: Implement actual model replication logic
	// This would involve P2P transfer of the model file
	
	// For now, just simulate the replication
	time.Sleep(5 * time.Second) // Simulate replication time
	
	// Register the model on the new node
	s.scheduler.RegisterModel(modelName, 0, "", nodeID)
	
	// Broadcast update via WebSocket
	if s.wsHub != nil {
		s.wsHub.Broadcast(map[string]interface{}{
			"type": "model_replicated",
			"model": modelName,
			"node": nodeID,
		})
	}
}

// WebSocket handler

func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	
	s.wsHub.register <- conn
	
	// Handle messages
	go s.handleWSConnection(conn)
}

func (s *Server) handleWSConnection(conn *websocket.Conn) {
	defer func() {
		s.wsHub.unregister <- conn
		conn.Close()
	}()
	
	// Send initial metrics when client connects
	go s.sendInitialMetrics(conn)
	
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		
		// Handle different message types
		switch msg["type"] {
		case "subscribe":
			// Handle subscription to specific events
			if channel, ok := msg["channel"].(string); ok {
				s.subscribeToChannel(conn, channel)
			}
		case "unsubscribe":
			// Handle unsubscription from events
			if channel, ok := msg["channel"].(string); ok {
				s.unsubscribeFromChannel(conn, channel)
			}
		case "ping":
			// Handle ping
			conn.WriteJSON(map[string]interface{}{"type": "pong"})
		}
	}
}

// WebSocket Hub methods

func (h *WSHub) run() {
	// Use a cleanup ticker to prevent memory leaks
	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
			
		case conn := <-h.unregister:
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				// Ensure connection is properly closed
				if err := conn.Close(); err != nil {
					fmt.Printf("Error closing WebSocket connection: %v\n", err)
				}
			}
			
		case message := <-h.broadcast:
			// Create a list of connections to remove to avoid concurrent map access
			var toRemove []*websocket.Conn
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					toRemove = append(toRemove, conn)
				}
			}
			// Remove failed connections
			for _, conn := range toRemove {
				delete(h.clients, conn)
				conn.Close()
			}
			
		case <-cleanupTicker.C:
			// Periodic cleanup of stale connections
			h.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes connections that haven't been active
func (h *WSHub) cleanupStaleConnections() {
	var toRemove []*websocket.Conn
	
	for conn := range h.clients {
		// Send ping to check if connection is alive
		if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			toRemove = append(toRemove, conn)
		}
	}
	
	// Remove dead connections
	for _, conn := range toRemove {
		delete(h.clients, conn)
		conn.Close()
	}
}

func (h *WSHub) Broadcast(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	
	select {
	case h.broadcast <- data:
	default:
		// Channel full, drop message
	}
}

// WebSocket helper methods

func (s *Server) sendInitialMetrics(conn *websocket.Conn) {
	// Send initial cluster status
	clusterStatus := map[string]interface{}{
		"type": "cluster_status",
		"data": map[string]interface{}{
			"node_id":   s.p2p.ID().String(),
			"is_leader": s.consensus.IsLeader(),
			"leader":    s.consensus.Leader(),
			"peers":     len(s.p2p.ConnectedPeers()),
			"status":    "healthy",
		},
	}
	conn.WriteJSON(clusterStatus)
	
	// Send initial metrics
	metrics := s.getSystemMetrics()
	metricsMsg := map[string]interface{}{
		"type": "metrics",
		"data": metrics,
	}
	conn.WriteJSON(metricsMsg)
}

func (s *Server) subscribeToChannel(conn *websocket.Conn, channel string) {
	// Implementation for channel subscriptions
	// For now, just acknowledge the subscription
	response := map[string]interface{}{
		"type":    "subscription_ack",
		"channel": channel,
		"status":  "subscribed",
	}
	conn.WriteJSON(response)
}

func (s *Server) unsubscribeFromChannel(conn *websocket.Conn, channel string) {
	// Implementation for channel unsubscriptions
	response := map[string]interface{}{
		"type":    "unsubscription_ack",
		"channel": channel,
		"status":  "unsubscribed",
	}
	conn.WriteJSON(response)
}

func (s *Server) getSystemMetrics() map[string]interface{} {
	stats := s.scheduler.GetStats()
	nodes := s.scheduler.GetNodes()
	
	// Calculate CPU, memory, and network usage
	cpuUsage := 25.0 + float64(len(nodes)*3) // Mock CPU based on nodes
	memoryUsage := 40.0 + float64(s.scheduler.GetModelCount()*2) // Mock memory based on models
	networkUsage := 15.0 + float64(len(s.p2p.ConnectedPeers())*5) // Mock network based on peers
	
	return map[string]interface{}{
		"cpu":     cpuUsage,
		"memory":  memoryUsage,
		"network": networkUsage,
		"nodes":   len(nodes),
		"models":  s.scheduler.GetModelCount(),
		"peers":   len(s.p2p.ConnectedPeers()),
		"stats":   stats,
	}
}

// StartMetricsBroadcasting starts broadcasting metrics to WebSocket clients
func (s *Server) StartMetricsBroadcasting() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Broadcast metrics to all connected clients
				metrics := s.getSystemMetrics()
				metricsMsg := map[string]interface{}{
					"type": "metrics",
					"data": metrics,
				}
				s.wsHub.Broadcast(metricsMsg)
			}
		}
	}()
}

// BroadcastModelEvent broadcasts model-related events
func (s *Server) BroadcastModelEvent(eventType string, modelName string, data map[string]interface{}) {
	event := map[string]interface{}{
		"type": "model_event",
		"event_type": eventType,
		"model_name": modelName,
		"data": data,
		"timestamp": time.Now().Unix(),
	}
	s.wsHub.Broadcast(event)
}

// BroadcastNodeEvent broadcasts node-related events
func (s *Server) BroadcastNodeEvent(eventType string, nodeID string, data map[string]interface{}) {
	event := map[string]interface{}{
		"type": "node_event",
		"event_type": eventType,
		"node_id": nodeID,
		"data": data,
		"timestamp": time.Now().Unix(),
	}
	s.wsHub.Broadcast(event)
}

// BroadcastAlert broadcasts alert messages
func (s *Server) BroadcastAlert(level string, message string) {
	alert := map[string]interface{}{
		"type": "alert",
		"level": level,
		"message": message,
		"timestamp": time.Now().Unix(),
	}
	s.wsHub.Broadcast(alert)
}

// Helper functions

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Security middleware implementations

// authMiddleware provides JWT-based authentication
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check and options
		if c.Request.URL.Path == "/api/v1/health" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		
		// TODO: Get secret from config
		secretKey := []byte("your-secret-key") // This should come from config
		
		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user_id", claims["user_id"])
			c.Set("username", claims["username"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
	}
}

// rateLimitMiddleware provides rate limiting
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with redis or in-memory store
		// For now, just add headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", s.config.RateLimit.RPS))
		c.Header("X-RateLimit-Remaining", "999") // Mock value
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		c.Next()
	}
}

// inputValidationMiddleware provides input validation and sanitization
func (s *Server) inputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate content type for POST/PUT requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.Contains(contentType, "application/json") && 
			   !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
			   !strings.Contains(contentType, "multipart/form-data") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported content type"})
				c.Abort()
				return
			}
		}

		// Validate request size
		if c.Request.ContentLength > s.config.MaxBodySize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request body too large"})
			c.Abort()
			return
		}

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}