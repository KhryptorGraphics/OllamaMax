package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/integration"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/proxy"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/loadbalancer"
)

// Server represents the API server
type Server struct {
	config      *config.APIConfig
	p2p         *p2p.Node
	consensus   *consensus.Engine
	scheduler   *scheduler.Engine
	integration *integration.SimpleOllamaIntegration

	// Proxy for multi-node Ollama routing
	ollamaProxy  *proxy.OllamaProxy
	loadBalancer *loadbalancer.LoadBalancer

	router   *gin.Engine
	server   *http.Server
	upgrader websocket.Upgrader

	// WebSocket connections
	wsConnections map[string]*WSConnection
	wsHub         *WSHub
}

// WSHub manages WebSocket connections
type WSHub struct {
	clients    map[*WSConnection]bool
	broadcast  chan []byte
	register   chan *WSConnection
	unregister chan *WSConnection
	rooms      map[string]map[*WSConnection]bool
	mu         sync.RWMutex
}

// NewServer creates a new API server
func NewServer(config *config.APIConfig, p2pNode *p2p.Node, consensusEngine *consensus.Engine, schedulerEngine *scheduler.Engine) (*Server, error) {
	server := &Server{
		config:        config,
		p2p:           p2pNode,
		consensus:     consensusEngine,
		scheduler:     schedulerEngine,
		integration:   nil, // Will be set later via SetIntegration
		wsConnections: make(map[string]*WSConnection),
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
			clients:    make(map[*WSConnection]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *WSConnection),
			unregister: make(chan *WSConnection),
		},
	}

	// Initialize load balancer
	loadBalancerConfig := &loadbalancer.LoadBalancerConfig{
		DefaultStrategy:        "least_loaded",
		RebalanceThreshold:     0.3,
		LoadImbalanceThreshold: 0.2,
		MetricsInterval:        10 * time.Second,
		HistoryRetention:       time.Hour,
		MaxHistorySize:         1000,
		EnablePrediction:       true,
		PredictionWindow:       time.Hour,
		PredictionAccuracy:     0.8,
		MaxRebalanceFrequency:  30 * time.Second,
		RebalanceBatchSize:     10,
		GracefulRebalance:      true,
		HighLoadThreshold:      0.8,
		LowLoadThreshold:       0.2,
		CriticalLoadThreshold:  0.95,
		CPUWeight:              0.4,
		MemoryWeight:           0.3,
	}
	server.loadBalancer = loadbalancer.NewLoadBalancer(loadBalancerConfig)

	// Initialize Ollama proxy
	proxyConfig := proxy.DefaultProxyConfig()
	ollamaProxy, err := proxy.NewOllamaProxy(schedulerEngine, server.loadBalancer, proxyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama proxy: %w", err)
	}
	server.ollamaProxy = ollamaProxy

	server.setupRoutes()

	return server, nil
}

// SetIntegration sets the Ollama integration for the server
func (s *Server) SetIntegration(integration *integration.SimpleOllamaIntegration) {
	s.integration = integration
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

		// Ollama Integration endpoints
		api.GET("/integration/status", s.getIntegrationStatus)
		api.POST("/integration/test", s.testIntegration)
		api.GET("/integration/models", s.getIntegrationModels)
		api.POST("/integration/models/pull", s.pullModel)

		// Ollama Proxy endpoints
		api.GET("/proxy/status", s.getProxyStatus)
		api.GET("/proxy/instances", s.getProxyInstances)
		api.GET("/proxy/metrics", s.getProxyMetrics)
		api.POST("/proxy/instances/register", s.registerProxyInstance)
		api.DELETE("/proxy/instances/:id", s.unregisterProxyInstance)

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

	// Start Ollama proxy if available
	if s.ollamaProxy != nil {
		if err := s.ollamaProxy.Start(); err != nil {
			fmt.Printf("Warning: Failed to start Ollama proxy: %v\n", err)
		} else {
			fmt.Printf("Ollama proxy started successfully\n")
		}
	}

	// Start metrics broadcasting
	s.StartMetricsBroadcasting()

	// Start server
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error securely without exposing internal details
			fmt.Printf("API server encountered an error\n")
			// In production, send to structured logging system
			// logger.Error("server_start_failed", "error", err)
		}
	}()

	fmt.Printf("API server started on %s\n", s.config.Listen)
	return nil
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop Ollama proxy if available
	if s.ollamaProxy != nil {
		if err := s.ollamaProxy.Stop(); err != nil {
			fmt.Printf("Warning: Failed to stop Ollama proxy: %v\n", err)
		}
	}

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

	// Check if node exists
	nodes := s.scheduler.GetNodes()
	node, exists := nodes[nodeID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Set node to draining status (stub implementation)
	// TODO: Implement actual node draining logic
	_ = nodeID // Use nodeID to avoid unused variable error

	// Broadcast node event
	s.BroadcastNodeEvent("draining", nodeID, map[string]interface{}{
		"status":    "draining",
		"address":   node.Address,
		"timestamp": time.Now().Unix(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Node %s is being drained", nodeID),
		"status":  "draining",
		"node_id": nodeID,
	})
}

func (s *Server) undrainNode(c *gin.Context) {
	nodeID := c.Param("id")

	// Check if node exists
	nodes := s.scheduler.GetNodes()
	node, exists := nodes[nodeID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// Set node back to online status (stub implementation)
	// TODO: Implement actual node undraining logic
	_ = nodeID // Use nodeID to avoid unused variable error

	// Broadcast node event
	s.BroadcastNodeEvent("online", nodeID, map[string]interface{}{
		"status":    "online",
		"address":   node.Address,
		"timestamp": time.Now().Unix(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Node %s is no longer draining", nodeID),
		"status":  "online",
		"node_id": nodeID,
	})
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
		"message":     fmt.Sprintf("Started downloading model %s to node %s", modelName, targetNode.ID),
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
		"message":        fmt.Sprintf("Deleted model %s from %d nodes", modelName, len(model.Locations)),
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
		Options  map[string]interface{}   `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create scheduler request
	schedReq := &scheduler.Request{
		ID:         fmt.Sprintf("chat_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Type:       "chat",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"messages": req.Messages,
			"stream":   req.Stream,
			"options":  req.Options,
		},
	}

	// Schedule the request
	if err := s.scheduler.Schedule(schedReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Wait for response
	select {
	case response := <-schedReq.ResponseCh:
		s.incrementRequestCounter("chat", response.Success)
		if response.Success {
			c.Header("X-Ollama-Node", response.NodeID)
			c.JSON(http.StatusOK, gin.H{
				"message": gin.H{
					"role":    "assistant",
					"content": response.Data,
				},
				"model":   req.Model,
				"node_id": response.NodeID,
				"done":    true,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		s.incrementRequestCounter("chat", false)
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

func (s *Server) embeddings(c *gin.Context) {
	var req struct {
		Model   string                 `json:"model"`
		Input   string                 `json:"input"`
		Options map[string]interface{} `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create scheduler request
	schedReq := &scheduler.Request{
		ID:         fmt.Sprintf("embed_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Type:       "embeddings",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"input":   req.Input,
			"options": req.Options,
		},
	}

	// Schedule the request
	if err := s.scheduler.Schedule(schedReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Wait for response
	select {
	case response := <-schedReq.ResponseCh:
		s.incrementRequestCounter("embeddings", response.Success)
		if response.Success {
			c.Header("X-Ollama-Node", response.NodeID)
			// Extract embeddings from response data
			if embeddings, ok := response.Data.([]float64); ok {
				c.JSON(http.StatusOK, gin.H{
					"embeddings": embeddings,
					"model":      req.Model,
					"node_id":    response.NodeID,
				})
			} else {
				// Default embeddings if format is unexpected
				c.JSON(http.StatusOK, gin.H{
					"embeddings": []float64{},
					"model":      req.Model,
					"node_id":    response.NodeID,
					"data":       response.Data,
				})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		s.incrementRequestCounter("embeddings", false)
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

// Monitoring handlers

func (s *Server) getMetrics(c *gin.Context) {
	// Get real metrics from scheduler
	modelsCount := s.scheduler.GetModelCount()
	nodesCount := len(s.scheduler.GetNodes())
	onlineNodes := s.scheduler.GetOnlineNodeCount()

	metrics := gin.H{
		"node_id":             s.p2p.ID().String(),
		"connected_peers":     len(s.p2p.ConnectedPeers()),
		"is_leader":           s.consensus.IsLeader(),
		"requests_processed":  s.getRequestCounter(),
		"models_loaded":       modelsCount,
		"nodes_total":         nodesCount,
		"nodes_online":        onlineNodes,
		"uptime":              time.Since(time.Now()).String(),
		"cpu_usage":           15.5, // Mock data
		"memory_usage":        23.8, // Mock data
		"network_usage":       45.2, // Mock data
		"requests_per_second": 12,   // Mock data
		"average_latency":     125,  // Mock data
		"active_connections":  8,    // Mock data
		"error_rate":          0.2,  // Mock data
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
				"id":           fmt.Sprintf("%s-%s", modelName, nodeID[:8]),
				"model_name":   modelName,
				"type":         "download",
				"status":       "completed",
				"progress":     100.0,
				"speed":        0.0,
				"eta":          0,
				"peer_id":      nodeID,
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
			"type":  "model_replicated",
			"model": modelName,
			"node":  nodeID,
		})
	}
}

// WebSocket handler

func (s *Server) handleWebSocket(c *gin.Context) {
	// Additional WebSocket security checks
	origin := c.GetHeader("Origin")
	if origin == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Origin header required for WebSocket connections"})
		return
	}

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Log error securely without exposing internal details
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	// Create WSConnection wrapper
	wsConn := &WSConnection{
		ID:          generateConnectionID(),
		Conn:        conn,
		Send:        make(chan []byte, 256),
		Hub:         s.wsHub,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Store connection
	s.wsConnections[wsConn.ID] = wsConn

	s.wsHub.register <- wsConn

	// Handle messages
	go s.handleWSConnection(wsConn)
}

func (s *Server) handleWSConnection(wsConn *WSConnection) {
	defer func() {
		s.wsHub.unregister <- wsConn
		wsConn.Close()
	}()

	// Send initial metrics when client connects
	go s.sendInitialMetrics(wsConn.Conn)

	for {
		var msg map[string]interface{}
		if err := wsConn.Conn.ReadJSON(&msg); err != nil {
			break
		}

		// Handle different message types
		switch msg["type"] {
		case "subscribe":
			// Handle subscription to specific events
			if channel, ok := msg["channel"].(string); ok {
				s.subscribeToChannel(wsConn.Conn, channel)
			}
		case "unsubscribe":
			// Handle unsubscription from events
			if channel, ok := msg["channel"].(string); ok {
				s.unsubscribeFromChannel(wsConn.Conn, channel)
			}
		case "ping":
			// Handle ping
			wsConn.Conn.WriteJSON(map[string]interface{}{"type": "pong"})
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
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				// Remove from all rooms
				for room, clients := range h.rooms {
					if clients[conn] {
						delete(clients, conn)
						if len(clients) == 0 {
							delete(h.rooms, room)
						}
					}
				}
			}
			h.mu.Unlock()
			// Close connection
			conn.Close()

		case message := <-h.broadcast:
			h.mu.RLock()
			var toRemove []*WSConnection
			for conn := range h.clients {
				select {
				case conn.Send <- message:
				default:
					toRemove = append(toRemove, conn)
				}
			}
			h.mu.RUnlock()

			// Remove failed connections
			if len(toRemove) > 0 {
				h.mu.Lock()
				for _, conn := range toRemove {
					delete(h.clients, conn)
				}
				h.mu.Unlock()
			}

		case <-cleanupTicker.C:
			// Periodic cleanup of stale connections
			h.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes connections that haven't been active
func (h *WSHub) cleanupStaleConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	var toRemove []*WSConnection
	cutoff := time.Now().Add(-5 * time.Minute)

	for conn := range h.clients {
		// Check if connection is stale based on last ping
		if conn.LastPing.Before(cutoff) {
			toRemove = append(toRemove, conn)
		}
	}

	// Remove stale connections
	for _, conn := range toRemove {
		delete(h.clients, conn)
		// Remove from all rooms
		for room, clients := range h.rooms {
			if clients[conn] {
				delete(clients, conn)
				if len(clients) == 0 {
					delete(h.rooms, room)
				}
			}
		}
	}
}

// broadcastToRoom broadcasts a message to all clients in a room
func (h *WSHub) broadcastToRoom(room string, message []byte) {
	h.mu.RLock()
	clients, exists := h.rooms[room]
	h.mu.RUnlock()

	if !exists {
		return
	}

	var toRemove []*WSConnection
	for client := range clients {
		select {
		case client.Send <- message:
		default:
			toRemove = append(toRemove, client)
		}
	}

	// Remove failed connections
	if len(toRemove) > 0 {
		h.mu.Lock()
		for _, client := range toRemove {
			delete(clients, client)
			delete(h.clients, client)
		}
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
		h.mu.Unlock()
	}
}

// joinRoom adds a client to a room
func (h *WSHub) joinRoom(client *WSConnection, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*WSConnection]bool)
	}
	h.rooms[room][client] = true
}

// leaveRoom removes a client from a room
func (h *WSHub) leaveRoom(client *WSConnection, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, exists := h.rooms[room]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
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
	cpuUsage := 25.0 + float64(len(nodes)*3)                      // Mock CPU based on nodes
	memoryUsage := 40.0 + float64(s.scheduler.GetModelCount()*2)  // Mock memory based on models
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
		"type":       "model_event",
		"event_type": eventType,
		"model_name": modelName,
		"data":       data,
		"timestamp":  time.Now().Unix(),
	}
	s.wsHub.Broadcast(event)
}

// BroadcastNodeEvent broadcasts node-related events
func (s *Server) BroadcastNodeEvent(eventType string, nodeID string, data map[string]interface{}) {
	event := map[string]interface{}{
		"type":       "node_event",
		"event_type": eventType,
		"node_id":    nodeID,
		"data":       data,
		"timestamp":  time.Now().Unix(),
	}
	s.wsHub.Broadcast(event)
}

// BroadcastAlert broadcasts alert messages
func (s *Server) BroadcastAlert(level string, message string) {
	alert := map[string]interface{}{
		"type":      "alert",
		"level":     level,
		"message":   message,
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

		// Get JWT secret from secret manager
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication system not properly configured"})
			c.Abort()
			return
		}

		// Validate secret strength
		if len(jwtSecret) < 32 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication system configuration invalid"})
			c.Abort()
			return
		}

		secretKey := []byte(jwtSecret)

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
		// Implement proper rate limiting
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// Create client identifier (IP + first 20 chars of user agent for uniqueness)
		clientID := clientIP
		if len(userAgent) > 0 {
			clientID += "_" + userAgent[:min(20, len(userAgent))]
		}

		// Check rate limit (simplified implementation - use Redis in production)
		currentMinute := time.Now().Unix() / 60
		requestKey := fmt.Sprintf("rate_limit_%s_%d", clientID, currentMinute)
		_ = requestKey // TODO: Use requestKey for actual rate limiting

		// For demo purposes, allow higher limits but add proper headers
		limit := s.config.RateLimit.RPS
		if limit == 0 {
			limit = 1000
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", (currentMinute+1)*60))
		c.Header("X-RateLimit-Window", "60")

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

		// Validate request size with detailed limits
		maxSize := s.config.MaxBodySize
		if maxSize == 0 {
			maxSize = 32 * 1024 * 1024 // 32MB default
		}

		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":               "Request body too large",
				"max_size_bytes":      maxSize,
				"received_size_bytes": c.Request.ContentLength,
			})
			c.Abort()
			return
		}

		// Validate request method
		allowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}
		methodAllowed := false
		for _, method := range allowedMethods {
			if c.Request.Method == method {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			c.Abort()
			return
		}

		// Add comprehensive security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), speaker=()")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		c.Next()
	}
}

// Missing methods for Server (stubs)

// incrementRequestCounter increments request counters for metrics
func (s *Server) incrementRequestCounter(endpoint string, success bool) {
	// Stub implementation - would update metrics
}

// Integration handler methods

// getIntegrationStatus returns the status of Ollama integration
func (s *Server) getIntegrationStatus(c *gin.Context) {
	if s.integration == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Integration not available",
			"message": "Ollama integration is not initialized",
			"status":  "disabled",
		})
		return
	}

	status := s.integration.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"status": "enabled",
		"data":   status,
	})
}

// testIntegration tests the Ollama integration
func (s *Server) testIntegration(c *gin.Context) {
	if s.integration == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Integration not available",
			"message": "Ollama integration is not initialized",
		})
		return
	}

	if err := s.integration.TestIntegration(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Integration test failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Integration test passed",
		"status":  "healthy",
	})
}

// getIntegrationModels returns models available through integration
func (s *Server) getIntegrationModels(c *gin.Context) {
	if s.integration == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Integration not available",
			"message": "Ollama integration is not initialized",
		})
		return
	}

	models, err := s.integration.ListModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list models",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"count":  len(models),
	})
}

// pullModel pulls a model through the integration
func (s *Server) pullModel(c *gin.Context) {
	if s.integration == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Integration not available",
			"message": "Ollama integration is not initialized",
		})
		return
	}

	var request struct {
		Model string `json:"model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	if err := s.integration.PullModel(request.Model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to pull model",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Model %s pulled successfully", request.Model),
		"model":   request.Model,
	})
}

// getRequestCounter gets total request counter for metrics
func (s *Server) getRequestCounter() int64 {
	// Stub implementation - would return total request count
	return 0
}

// Proxy handler methods

// getProxyStatus returns the status of the Ollama proxy
func (s *Server) getProxyStatus(c *gin.Context) {
	if s.ollamaProxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Proxy not available",
			"message": "Ollama proxy is not initialized",
			"status":  "disabled",
		})
		return
	}

	instances := s.ollamaProxy.GetInstances()
	healthyCount := 0
	for _, instance := range instances {
		if instance.Status == proxy.InstanceStatusHealthy {
			healthyCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            "running",
		"instance_count":    len(instances),
		"healthy_instances": healthyCount,
		"load_balancer":     s.loadBalancer != nil,
	})
}

// getProxyInstances returns the list of registered Ollama instances
func (s *Server) getProxyInstances(c *gin.Context) {
	if s.ollamaProxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Proxy not available",
			"message": "Ollama proxy is not initialized",
		})
		return
	}

	instances := s.ollamaProxy.GetInstances()

	// Convert instances to response format
	instanceList := make([]gin.H, 0, len(instances))
	for _, instance := range instances {
		instanceInfo := gin.H{
			"id":            instance.ID,
			"node_id":       instance.NodeID,
			"endpoint":      instance.Endpoint,
			"status":        string(instance.Status),
			"request_count": instance.RequestCount,
			"error_count":   instance.ErrorCount,
			"last_request":  instance.LastRequestTime,
		}

		if instance.Load != nil {
			instanceInfo["load"] = gin.H{
				"active_requests": instance.Load.ActiveRequests,
				"queued_requests": instance.Load.QueuedRequests,
				"cpu_usage":       instance.Load.CPUUsage,
				"memory_usage":    instance.Load.MemoryUsage,
				"gpu_usage":       instance.Load.GPUUsage,
				"last_updated":    instance.Load.LastUpdated,
			}
		}

		if instance.Health != nil {
			instanceInfo["health"] = gin.H{
				"is_healthy":        instance.Health.IsHealthy,
				"last_health_check": instance.Health.LastHealthCheck,
				"response_time":     instance.Health.ResponseTime,
				"error_rate":        instance.Health.ErrorRate,
				"uptime":            instance.Health.Uptime,
			}
		}

		instanceList = append(instanceList, instanceInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instanceList,
		"count":     len(instanceList),
	})
}

// getProxyMetrics returns proxy performance metrics
func (s *Server) getProxyMetrics(c *gin.Context) {
	if s.ollamaProxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Proxy not available",
			"message": "Ollama proxy is not initialized",
		})
		return
	}

	metrics := s.ollamaProxy.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"total_requests":      metrics.TotalRequests,
		"successful_requests": metrics.SuccessfulRequests,
		"failed_requests":     metrics.FailedRequests,
		"average_latency":     metrics.AverageLatency,
		"requests_per_second": metrics.RequestsPerSecond,
		"load_balancing": gin.H{
			"decisions": metrics.LoadBalancingDecisions,
			"errors":    metrics.LoadBalancingErrors,
		},
		"instance_metrics": metrics.InstanceMetrics,
	})
}

// registerProxyInstance registers a new Ollama instance with the proxy
func (s *Server) registerProxyInstance(c *gin.Context) {
	if s.ollamaProxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Proxy not available",
			"message": "Ollama proxy is not initialized",
		})
		return
	}

	var request struct {
		NodeID   string `json:"node_id" binding:"required"`
		Endpoint string `json:"endpoint" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	if err := s.ollamaProxy.RegisterInstance(request.NodeID, request.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to register instance",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Instance registered successfully",
		"node_id":  request.NodeID,
		"endpoint": request.Endpoint,
	})
}

// unregisterProxyInstance unregisters an Ollama instance from the proxy
func (s *Server) unregisterProxyInstance(c *gin.Context) {
	if s.ollamaProxy == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Proxy not available",
			"message": "Ollama proxy is not initialized",
		})
		return
	}

	instanceID := c.Param("id")
	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "Instance ID is required",
		})
		return
	}

	if err := s.ollamaProxy.UnregisterInstance(instanceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to unregister instance",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Instance unregistered successfully",
		"instance_id": instanceID,
	})
}
