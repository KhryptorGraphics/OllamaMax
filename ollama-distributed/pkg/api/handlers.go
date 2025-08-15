package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// health returns the health status of the API server
func (s *Server) health(c *gin.Context) {
	// Get node ID from P2P node if available
	nodeID := "unknown"
	if s.p2p != nil {
		nodeID = string(s.p2p.ID())
	}

	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"node_id":   nodeID,
		"services": gin.H{
			"p2p":       s.p2p != nil,
			"consensus": s.consensus != nil,
			"scheduler": s.scheduler != nil,
		},
	}

	// Check if services are actually healthy
	if s.p2p != nil {
		status["services"].(gin.H)["p2p_peers"] = 0 // TODO: Implement GetPeers method
	}

	if s.consensus != nil {
		status["services"].(gin.H)["consensus_leader"] = s.consensus.IsLeader()
	}

	if s.scheduler != nil {
		status["services"].(gin.H)["available_nodes"] = len(s.scheduler.GetAvailableNodes())
	}

	c.JSON(http.StatusOK, status)
}

// version returns the API version information
func (s *Server) version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    "1.0.0",
		"build_date": "2024-01-01",
		"git_commit": "unknown",
		"go_version": "1.21+",
	})
}

// getModels returns all available models
func (s *Server) getModels(c *gin.Context) {
	models := s.scheduler.GetAllModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

// getModel returns a specific model
func (s *Server) getModel(c *gin.Context) {
	modelName := c.Param("name")

	// Validate model name for security
	if err := security.ValidateModelName(modelName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

	// Get specific model from scheduler
	if model, exists := s.scheduler.GetModel(modelName); exists {
		c.JSON(http.StatusOK, gin.H{"model": model})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
	}
}

// downloadModel initiates model download
func (s *Server) downloadModel(c *gin.Context) {
	modelName := c.Param("name")

	// Validate model name for security
	if err := security.ValidateModelName(modelName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

	// Get available nodes from scheduler
	nodes := s.scheduler.GetNodes()
	if len(nodes) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available nodes for model download"})
		return
	}

	// Select the first available node for download
	var targetNodeID string
	for nodeID, node := range nodes {
		if node.Status == "online" || node.Status == "active" {
			targetNodeID = nodeID
			break
		}
	}

	if targetNodeID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No online nodes available for download"})
		return
	}

	// TODO: Implement actual model download logic through distributed scheduler
	// For now, simulate successful download initiation
	c.JSON(http.StatusOK, gin.H{
		"message":     "Model download initiated",
		"model_name":  modelName,
		"target_node": targetNodeID,
		"status":      "downloading",
		"progress":    0.0,
	})
}

// deleteModel removes a model
func (s *Server) deleteModel(c *gin.Context) {
	modelName := c.Param("name")

	// Validate model name for security
	if err := security.ValidateModelName(modelName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

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

	// Broadcast model update
	s.BroadcastModelUpdate(modelName, "deleted", 100.0)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Model deleted successfully",
		"model_name": modelName,
		"model":      model,
	})
}

// getNodes returns all available nodes
func (s *Server) getNodes(c *gin.Context) {
	nodes := s.scheduler.GetAvailableNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

// getNode returns a specific node
func (s *Server) getNode(c *gin.Context) {
	nodeID := c.Param("id")

	// Validate node ID for security
	if err := security.ValidateNodeID(nodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid node ID: %v", err)})
		return
	}

	// Get all nodes and find the specific one
	nodes := s.scheduler.GetNodes()
	node, exists := nodes[nodeID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"node": node})
}

// drainNode marks a node for draining (no new tasks)
func (s *Server) drainNode(c *gin.Context) {
	nodeID := c.Param("id")

	// Validate node ID for security
	if err := security.ValidateNodeID(nodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid node ID: %v", err)})
		return
	}

	// Check if node exists
	nodes := s.scheduler.GetNodes()
	_, exists := nodes[nodeID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// For now, just return success - actual draining logic would be implemented in scheduler
	c.JSON(http.StatusOK, gin.H{
		"message": "Node marked for draining",
		"node_id": nodeID,
		"status":  "draining",
	})
}

// undrainNode removes drain status from a node
func (s *Server) undrainNode(c *gin.Context) {
	nodeID := c.Param("id")

	// Validate node ID for security
	if err := security.ValidateNodeID(nodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid node ID: %v", err)})
		return
	}

	// Check if node exists
	nodes := s.scheduler.GetNodes()
	_, exists := nodes[nodeID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	// For now, just return success - actual undraining logic would be implemented in scheduler
	c.JSON(http.StatusOK, gin.H{
		"message": "Node drain status removed",
		"node_id": nodeID,
		"status":  "active",
	})
}

// getMetrics returns system metrics
func (s *Server) getMetrics(c *gin.Context) {
	nodes := s.scheduler.GetNodes()
	healthyNodes := 0
	for _, node := range nodes {
		if node.Status == "online" || node.Status == "active" {
			healthyNodes++
		}
	}

	// Calculate uptime (mock for now)
	uptime := time.Since(time.Now().Add(-time.Hour)).Seconds()

	metrics := map[string]interface{}{
		"timestamp":             time.Now(),
		"node_id":               s.p2p.ID().String(),
		"connected_peers":       len(s.p2p.GetConnectedPeers()),
		"is_leader":             false, // TODO: Get from consensus engine
		"requests_processed":    0,     // TODO: Implement request counter
		"models_loaded":         0,     // TODO: Implement model tracking
		"nodes_total":           len(nodes),
		"nodes_online":          healthyNodes,
		"uptime":                uptime,
		"cpu_usage":             0.0, // TODO: Implement system metrics
		"memory_usage":          0.0, // TODO: Implement system metrics
		"network_usage":         0.0, // TODO: Implement system metrics
		"websocket_connections": s.wsHub.GetClientCount(),
	}

	c.JSON(http.StatusOK, metrics)
}

// GenerateRequest represents a generation request
type GenerateRequest struct {
	Model  string `json:"model" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	Stream bool   `json:"stream,omitempty"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Model    string                   `json:"model" binding:"required"`
	Messages []map[string]interface{} `json:"messages" binding:"required"`
	Stream   bool                     `json:"stream,omitempty"`
	Options  map[string]interface{}   `json:"options,omitempty"`
}

// generate handles text generation requests
func (s *Server) generate(c *gin.Context) {
	var req GenerateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate model name for security
	if err := security.ValidateModelName(req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

	// Validate prompt for security
	if err := security.ValidatePrompt(req.Prompt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid prompt: %v", err)})
		return
	}

	// TODO: Check if model exists when model management is implemented
	// For now, accept any model name for testing

	// Create a simple response for now
	// TODO: Implement proper request routing through scheduler
	response := map[string]interface{}{
		"model":    req.Model,
		"response": "This is a placeholder response. Distributed inference not yet implemented.",
		"done":     true,
	}

	c.JSON(http.StatusOK, response)
}

// chat handles chat completion requests
func (s *Server) chat(c *gin.Context) {
	var req ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate model name for security
	if err := security.ValidateModelName(req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

	// Validate messages for security
	for i, message := range req.Messages {
		if content, ok := message["content"].(string); ok {
			if err := security.ValidatePrompt(content); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid message %d: %v", i, err)})
				return
			}
		}
	}

	// TODO: Check if model exists when model management is implemented
	// For now, accept any model name for testing

	// Convert messages to prompt (simplified)
	prompt := ""
	for _, message := range req.Messages {
		if content, ok := message["content"].(string); ok {
			if role, roleOk := message["role"].(string); roleOk {
				prompt += fmt.Sprintf("%s: %s\n", role, content)
			} else {
				prompt += content + "\n"
			}
		}
	}

	// Create a simple response for now
	// TODO: Implement proper request routing through scheduler
	response := map[string]interface{}{
		"model": req.Model,
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": "This is a placeholder response. Distributed chat inference not yet implemented.",
		},
		"done": true,
	}

	c.JSON(http.StatusOK, response)
}

// embeddings handles embedding generation requests
func (s *Server) embeddings(c *gin.Context) {
	var req EmbeddingsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate model name for security
	if err := security.ValidateModelName(req.Model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid model name: %v", err)})
		return
	}

	// For now, return mock embeddings - actual implementation would use the distributed scheduler
	embeddings := [][]float64{
		make([]float64, 384), // Common embedding dimension
	}
	for j := range embeddings[0] {
		embeddings[0][j] = 0.1 // Mock values
	}

	response := EmbeddingsResponse{
		Embeddings: embeddings,
	}

	c.JSON(http.StatusOK, response)
}

// getClusterStatus returns the current cluster status
func (s *Server) getClusterStatus(c *gin.Context) {
	// Get cluster information from consensus engine
	nodes := s.scheduler.GetNodes()
	peers := make([]string, 0, len(nodes))
	for nodeID := range nodes {
		peers = append(peers, nodeID)
	}

	response := map[string]interface{}{
		"node_id":   "test-node-id", // TODO: Get actual node ID
		"is_leader": false,          // TODO: Get from consensus engine
		"leader":    "unknown",      // TODO: Get from consensus engine
		"peers":     peers,
		"status":    "active", // TODO: Get actual status
	}

	c.JSON(http.StatusOK, response)
}

// getClusterLeader returns the current cluster leader
func (s *Server) getClusterLeader(c *gin.Context) {
	// TODO: Get actual leader from consensus engine
	leader := map[string]interface{}{
		"id":      "unknown",
		"address": "unknown",
		"term":    0,
	}

	c.JSON(http.StatusOK, gin.H{"leader": leader})
}

// joinCluster handles cluster join requests
func (s *Server) joinCluster(c *gin.Context) {
	var req struct {
		NodeID  string `json:"node_id" binding:"required"`
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate node ID for security
	if err := security.ValidateNodeID(req.NodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid node ID: %v", err)})
		return
	}

	// TODO: Implement actual cluster join logic through consensus engine
	c.JSON(http.StatusOK, gin.H{
		"message": "Node join request accepted",
		"node_id": req.NodeID,
		"status":  "joining",
	})
}

// leaveCluster handles cluster leave requests
func (s *Server) leaveCluster(c *gin.Context) {
	var req struct {
		NodeID string `json:"node_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate node ID for security
	if err := security.ValidateNodeID(req.NodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid node ID: %v", err)})
		return
	}

	// TODO: Implement actual cluster leave logic through consensus engine
	c.JSON(http.StatusOK, gin.H{
		"message": "Node leave request accepted",
		"node_id": req.NodeID,
		"status":  "leaving",
	})
}

// getTransfers returns all active transfers
func (s *Server) getTransfers(c *gin.Context) {
	// TODO: Get actual transfers from transfer manager
	transfers := []map[string]interface{}{
		{
			"id":       "transfer-1",
			"model":    "llama2",
			"status":   "active",
			"progress": 0.75,
			"source":   "node-1",
			"target":   "node-2",
		},
	}

	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}

// getTransfer returns a specific transfer by ID
func (s *Server) getTransfer(c *gin.Context) {
	transferID := c.Param("id")

	// Validate transfer ID for security
	if err := security.ValidateTransferID(transferID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid transfer ID: %v", err)})
		return
	}

	// TODO: Get actual transfer from transfer manager
	transfer := map[string]interface{}{
		"id":       transferID,
		"model":    "llama2",
		"status":   "active",
		"progress": 0.75,
		"source":   "node-1",
		"target":   "node-2",
		"started":  time.Now().Add(-10 * time.Minute),
	}

	c.JSON(http.StatusOK, gin.H{"transfer": transfer})
}

// cancelTransfer cancels a specific transfer
func (s *Server) cancelTransfer(c *gin.Context) {
	transferID := c.Param("id")

	// Validate transfer ID for security
	if err := security.ValidateTransferID(transferID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid transfer ID: %v", err)})
		return
	}

	// TODO: Implement actual transfer cancellation
	c.JSON(http.StatusOK, gin.H{
		"message":     "Transfer cancelled",
		"transfer_id": transferID,
		"status":      "cancelled",
	})
}

// autoConfigureDistribution handles automatic distribution configuration
func (s *Server) autoConfigureDistribution(c *gin.Context) {
	var req struct {
		Strategy string   `json:"strategy,omitempty"`
		Models   []string `json:"models,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual auto-configuration logic
	c.JSON(http.StatusOK, gin.H{
		"message":  "Auto-configuration initiated",
		"strategy": req.Strategy,
		"models":   req.Models,
		"status":   "configuring",
	})
}

// getStats returns detailed system statistics
func (s *Server) getStats(c *gin.Context) {
	stats := map[string]interface{}{
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()), // TODO: Track actual uptime
		"requests": map[string]interface{}{
			"total":   0, // TODO: Implement request counting
			"success": 0,
			"errors":  0,
		},
		"performance": map[string]interface{}{
			"avg_response_time": "0ms", // TODO: Implement performance tracking
			"requests_per_sec":  0,
		},
	}

	c.JSON(http.StatusOK, stats)
}

// getConfig returns system configuration (sanitized)
func (s *Server) getConfig(c *gin.Context) {
	config := map[string]interface{}{
		"api": map[string]interface{}{
			"listen": s.config.Listen,
		},
		"features": map[string]interface{}{
			"websocket":      true,
			"authentication": true,
			"rate_limiting":  true,
			"cors":           true,
		},
	}

	c.JSON(http.StatusOK, config)
}

// updateConfig updates system configuration
func (s *Server) updateConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement configuration updates
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}

// Authentication handlers for frontend integration

// login handles user authentication
func (s *Server) login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual authentication with database
	// For now, create a mock user and JWT token
	if req.Email == "admin@ollamamax.com" && req.Password == "admin123" {
		token, err := security.GenerateJWT(req.Email, "admin")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		user := gin.H{
			"id":        "1",
			"email":     req.Email,
			"firstName": "Admin",
			"lastName":  "User",
			"role":      "admin",
			"avatar":    "",
			"createdAt": time.Now(),
		}

		c.JSON(http.StatusOK, gin.H{
			"user":  user,
			"token": token,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

// register handles user registration
func (s *Server) register(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual user registration with database
	// For now, create a mock user
	user := gin.H{
		"id":        "2",
		"email":     req.Email,
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"role":      "user",
		"avatar":    "",
		"createdAt": time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// refreshToken handles JWT token refresh
func (s *Server) refreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual token refresh logic
	// For now, generate a new token
	newToken, err := security.GenerateJWT("admin@ollamamax.com", "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}

// logout handles user logout
func (s *Server) logout(c *gin.Context) {
	// TODO: Implement token blacklisting
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// getUserProfile returns the current user's profile
func (s *Server) getUserProfile(c *gin.Context) {
	// TODO: Get user from JWT token
	user := gin.H{
		"id":        "1",
		"email":     "admin@ollamamax.com",
		"firstName": "Admin",
		"lastName":  "User",
		"role":      "admin",
		"avatar":    "",
		"createdAt": time.Now().AddDate(0, -1, 0), // 1 month ago
		"lastLogin": time.Now().Add(-time.Hour),   // 1 hour ago
	}

	c.JSON(http.StatusOK, user)
}

// updateUserProfile updates the current user's profile
func (s *Server) updateUserProfile(c *gin.Context) {
	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Avatar    string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Update user in database
	user := gin.H{
		"id":        "1",
		"email":     "admin@ollamamax.com",
		"firstName": req.FirstName,
		"lastName":  req.LastName,
		"role":      "admin",
		"avatar":    req.Avatar,
		"updatedAt": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// Dashboard handlers for real-time data

// getDashboardData returns comprehensive dashboard data
func (s *Server) getDashboardData(c *gin.Context) {
	// Get cluster status
	clusterSize := 0
	if s.scheduler != nil {
		clusterSize = len(s.scheduler.GetAvailableNodes())
	}

	isLeader := false
	if s.consensus != nil {
		isLeader = s.consensus.IsLeader()
	}

	// Mock data for demonstration - replace with actual metrics
	data := gin.H{
		"clusterStatus": gin.H{
			"healthy":   true,
			"size":      clusterSize,
			"leader":    isLeader,
			"consensus": s.consensus != nil,
		},
		"nodeCount":       clusterSize,
		"activeModels":    3, // TODO: Get from model manager
		"totalRequests":   1250,
		"avgResponseTime": 245.5,
		"errorRate":       0.02,
		"uptime":          time.Since(time.Now().Add(-24 * time.Hour)).Seconds(),
		"nodes": []gin.H{
			{
				"id":       "node-1",
				"status":   "healthy",
				"cpu":      45.2,
				"memory":   67.8,
				"gpu":      23.1,
				"requests": 423,
			},
			{
				"id":       "node-2",
				"status":   "healthy",
				"cpu":      52.1,
				"memory":   71.3,
				"gpu":      34.7,
				"requests": 387,
			},
			{
				"id":       "node-3",
				"status":   "healthy",
				"cpu":      38.9,
				"memory":   59.2,
				"gpu":      18.5,
				"requests": 440,
			},
		},
		"recentActivity": []gin.H{
			{
				"timestamp": time.Now().Add(-5 * time.Minute),
				"type":      "inference",
				"message":   "Model llama2-7b completed inference request",
				"node":      "node-1",
			},
			{
				"timestamp": time.Now().Add(-8 * time.Minute),
				"type":      "cluster",
				"message":   "Node node-3 joined cluster",
				"node":      "node-3",
			},
			{
				"timestamp": time.Now().Add(-12 * time.Minute),
				"type":      "model",
				"message":   "Model codellama-13b loaded successfully",
				"node":      "node-2",
			},
		},
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, data)
}

// getDashboardMetrics returns performance metrics for charts
func (s *Server) getDashboardMetrics(c *gin.Context) {
	// Get time range from query parameters
	timeRange := c.DefaultQuery("range", "1h")

	// Mock metrics data - replace with actual metrics collection
	metrics := gin.H{
		"timeRange": timeRange,
		"requestsPerSecond": []gin.H{
			{"timestamp": time.Now().Add(-60 * time.Minute), "value": 12.3},
			{"timestamp": time.Now().Add(-50 * time.Minute), "value": 15.7},
			{"timestamp": time.Now().Add(-40 * time.Minute), "value": 18.2},
			{"timestamp": time.Now().Add(-30 * time.Minute), "value": 14.9},
			{"timestamp": time.Now().Add(-20 * time.Minute), "value": 21.1},
			{"timestamp": time.Now().Add(-10 * time.Minute), "value": 19.8},
			{"timestamp": time.Now(), "value": 16.5},
		},
		"responseTime": []gin.H{
			{"timestamp": time.Now().Add(-60 * time.Minute), "value": 234.5},
			{"timestamp": time.Now().Add(-50 * time.Minute), "value": 198.7},
			{"timestamp": time.Now().Add(-40 * time.Minute), "value": 267.3},
			{"timestamp": time.Now().Add(-30 * time.Minute), "value": 245.1},
			{"timestamp": time.Now().Add(-20 * time.Minute), "value": 189.9},
			{"timestamp": time.Now().Add(-10 * time.Minute), "value": 223.4},
			{"timestamp": time.Now(), "value": 245.5},
		},
		"errorRate": []gin.H{
			{"timestamp": time.Now().Add(-60 * time.Minute), "value": 0.01},
			{"timestamp": time.Now().Add(-50 * time.Minute), "value": 0.02},
			{"timestamp": time.Now().Add(-40 * time.Minute), "value": 0.015},
			{"timestamp": time.Now().Add(-30 * time.Minute), "value": 0.025},
			{"timestamp": time.Now().Add(-20 * time.Minute), "value": 0.018},
			{"timestamp": time.Now().Add(-10 * time.Minute), "value": 0.012},
			{"timestamp": time.Now(), "value": 0.02},
		},
		"resourceUsage": gin.H{
			"cpu": []gin.H{
				{"node": "node-1", "value": 45.2},
				{"node": "node-2", "value": 52.1},
				{"node": "node-3", "value": 38.9},
			},
			"memory": []gin.H{
				{"node": "node-1", "value": 67.8},
				{"node": "node-2", "value": 71.3},
				{"node": "node-3", "value": 59.2},
			},
			"gpu": []gin.H{
				{"node": "node-1", "value": 23.1},
				{"node": "node-2", "value": 34.7},
				{"node": "node-3", "value": 18.5},
			},
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// getNotifications returns user notifications
func (s *Server) getNotifications(c *gin.Context) {
	// Mock notifications - replace with actual notification system
	notifications := []gin.H{
		{
			"id":        "1",
			"type":      "info",
			"title":     "Cluster Status Update",
			"message":   "All nodes are healthy and operational",
			"timestamp": time.Now().Add(-10 * time.Minute),
			"read":      false,
		},
		{
			"id":        "2",
			"type":      "warning",
			"title":     "High Memory Usage",
			"message":   "Node-2 memory usage is at 85%",
			"timestamp": time.Now().Add(-30 * time.Minute),
			"read":      false,
		},
		{
			"id":        "3",
			"type":      "success",
			"title":     "Model Loaded",
			"message":   "CodeLlama-13B model loaded successfully",
			"timestamp": time.Now().Add(-1 * time.Hour),
			"read":      true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unreadCount":   2,
	})
}

// markNotificationRead marks a notification as read
func (s *Server) markNotificationRead(c *gin.Context) {
	notificationID := c.Param("id")

	// TODO: Update notification in database
	c.JSON(http.StatusOK, gin.H{
		"message":        "Notification marked as read",
		"notificationId": notificationID,
	})
}
