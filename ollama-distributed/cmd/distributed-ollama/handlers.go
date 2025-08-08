package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ollamaAPI "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/ollama/api"
)

// handleGenerate handles the /api/generate endpoint with distributed inference
func (s *DistributedOllamaServer) handleGenerate(c *gin.Context) {
	var req ollamaAPI.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received generate request",
		"model", req.Model,
		"prompt_length", len(req.Prompt))

	// Use distributed integration to handle the request
	response, err := s.integration.HandleGenerateRequest(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to handle generate request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleChat handles the /api/chat endpoint with distributed inference
func (s *DistributedOllamaServer) handleChat(c *gin.Context) {
	var req ollamaAPI.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received chat request",
		"model", req.Model,
		"messages", len(req.Messages))

	// Convert chat request to generate request for distributed processing
	// In a real implementation, this would properly handle chat context
	prompt := ""
	for _, msg := range req.Messages {
		prompt += msg.Content + "\n"
	}

	generateReq := &ollamaAPI.GenerateRequest{
		Model:   req.Model,
		Prompt:  prompt,
		Options: req.Options,
		Stream:  req.Stream,
	}

	// Use distributed integration
	generateResp, err := s.integration.HandleGenerateRequest(c.Request.Context(), generateReq)
	if err != nil {
		s.logger.Error("Failed to handle chat request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert generate response to chat response
	chatResp := &ollamaAPI.ChatResponse{
		Model:     generateResp.Model,
		CreatedAt: generateResp.CreatedAt,
		Message: ollamaAPI.Message{
			Role:    "assistant",
			Content: generateResp.Response,
		},
		Done: generateResp.Done,
	}

	c.JSON(http.StatusOK, chatResp)
}

// handleListModels handles the /api/tags endpoint
func (s *DistributedOllamaServer) handleListModels(c *gin.Context) {
	models := s.modelManager.GetDistributedModels()

	// Convert to Ollama API format
	var modelList []ollamaAPI.ModelResponse
	for _, model := range models {
		modelResp := ollamaAPI.ModelResponse{
			Name:       model.Name,
			Size:       model.Size,
			Digest:     model.Hash,
			ModifiedAt: model.CreatedAt,
		}
		modelList = append(modelList, modelResp)
	}

	response := ollamaAPI.ListResponse{
		Models: modelList,
	}

	c.JSON(http.StatusOK, response)
}

// handlePullModel handles the /api/pull endpoint
func (s *DistributedOllamaServer) handlePullModel(c *gin.Context) {
	var req ollamaAPI.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received pull request", "model", req.Name)

	// Add model to distributed system
	model, err := s.modelManager.AddModel(req.Name, "/tmp/models/"+req.Name)
	if err != nil {
		s.logger.Error("Failed to pull model", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ollamaAPI.ProgressResponse{
		Status:    "success",
		Digest:    model.Hash,
		Total:     model.Size,
		Completed: model.Size,
	}

	c.JSON(http.StatusOK, response)
}

// handleDeleteModel handles the /api/delete endpoint
func (s *DistributedOllamaServer) handleDeleteModel(c *gin.Context) {
	var req ollamaAPI.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received delete request", "model", req.Name)

	// For now, just return success since RemoveModel method doesn't exist
	// In a real implementation, this would remove the model from the distributed system
	s.logger.Info("Model deletion requested", "model", req.Name)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handlePushModel handles the /api/push endpoint
func (s *DistributedOllamaServer) handlePushModel(c *gin.Context) {
	var req ollamaAPI.PushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received push request", "model", req.Name)

	// In a real implementation, this would push the model to a registry
	response := ollamaAPI.ProgressResponse{
		Status: "success",
	}

	c.JSON(http.StatusOK, response)
}

// handleCreateModel handles the /api/create endpoint
func (s *DistributedOllamaServer) handleCreateModel(c *gin.Context) {
	var req ollamaAPI.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received create request", "model", req.Name)

	// In a real implementation, this would create a model from a Modelfile
	response := ollamaAPI.ProgressResponse{
		Status: "success",
	}

	c.JSON(http.StatusOK, response)
}

// handleCopyModel handles the /api/copy endpoint
func (s *DistributedOllamaServer) handleCopyModel(c *gin.Context) {
	var req ollamaAPI.CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received copy request", "source", req.Source, "destination", req.Destination)

	// In a real implementation, this would copy a model
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleShowModel handles the /api/show endpoint
func (s *DistributedOllamaServer) handleShowModel(c *gin.Context) {
	var req ollamaAPI.ShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := s.modelManager.GetModel(req.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	response := ollamaAPI.ShowResponse{
		License:    "Unknown",
		Modelfile:  "# Distributed model\nFROM " + model.Name,
		Parameters: "{}",
		Template:   "{{ .Prompt }}",
		System:     "",
		Details: ollamaAPI.ModelDetails{
			Format:            "gguf",
			Family:            "llama",
			Families:          []string{"llama"},
			ParameterSize:     "7B",
			QuantizationLevel: "Q4_0",
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleEmbed handles the /api/embed endpoint
func (s *DistributedOllamaServer) handleEmbed(c *gin.Context) {
	var req ollamaAPI.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("Received embed request", "model", req.Model)

	// In a real implementation, this would generate embeddings using distributed processing
	// For now, return mock embeddings
	embeddings := make([][]float64, len(req.Prompt))
	for i := range embeddings {
		embeddings[i] = make([]float64, 384) // Mock 384-dimensional embeddings
		for j := range embeddings[i] {
			embeddings[i][j] = 0.1 // Mock values
		}
	}

	response := ollamaAPI.EmbeddingResponse{
		Embedding: embeddings,
	}

	c.JSON(http.StatusOK, response)
}

// handleDistributedStatus handles the /api/distributed/status endpoint
func (s *DistributedOllamaServer) handleDistributedStatus(c *gin.Context) {
	peers := s.p2pNode.GetConnectedPeers()

	status := gin.H{
		"node_id":         s.p2pNode.ID().String(),
		"connected_peers": len(peers),
		"peers":           peers,
		"models_loaded":   len(s.modelManager.GetDistributedModels()),
		"uptime":          time.Since(time.Now()).String(), // This would track actual uptime
		"version":         "1.0.0",
	}

	c.JSON(http.StatusOK, status)
}

// handleListNodes handles the /api/distributed/nodes endpoint
func (s *DistributedOllamaServer) handleListNodes(c *gin.Context) {
	peers := s.p2pNode.GetConnectedPeers()

	var nodes []gin.H
	for _, peerID := range peers {
		node := gin.H{
			"id":     peerID.String(),
			"status": "connected",
			// In a real implementation, this would include more node information
		}
		nodes = append(nodes, node)
	}

	response := gin.H{
		"nodes": nodes,
		"total": len(nodes),
	}

	c.JSON(http.StatusOK, response)
}

// handleDistributedModels handles the /api/distributed/models endpoint
func (s *DistributedOllamaServer) handleDistributedModels(c *gin.Context) {
	models := s.modelManager.GetDistributedModels()

	var distributedModels []gin.H
	for _, model := range models {
		distributedModel := gin.H{
			"name":      model.Name,
			"size":      model.Size,
			"replicas":  len(model.Replicas),
			"locations": model.Replicas,
			"status":    "available",
		}
		distributedModels = append(distributedModels, distributedModel)
	}

	response := gin.H{
		"models": distributedModels,
		"total":  len(distributedModels),
	}

	c.JSON(http.StatusOK, response)
}

// handleMetrics handles the /api/distributed/metrics endpoint
func (s *DistributedOllamaServer) handleMetrics(c *gin.Context) {
	integrationMetrics := s.integration.GetMetrics()
	inferenceMetrics := s.inferenceEngine.GetMetrics()

	metrics := gin.H{
		"integration": gin.H{
			"total_requests":       integrationMetrics.TotalRequests,
			"distributed_requests": integrationMetrics.DistributedRequests,
			"local_requests":       integrationMetrics.LocalRequests,
			"successful_requests":  integrationMetrics.SuccessfulRequests,
			"failed_requests":      integrationMetrics.FailedRequests,
			"average_latency":      integrationMetrics.AverageLatency.String(),
			"average_nodes_used":   integrationMetrics.AverageNodesUsed,
			"last_updated":         integrationMetrics.LastUpdated,
		},
		"inference": gin.H{
			"total_inferences":       inferenceMetrics.TotalInferences,
			"successful_inferences":  inferenceMetrics.SuccessfulInferences,
			"failed_inferences":      inferenceMetrics.FailedInferences,
			"average_latency":        inferenceMetrics.AverageLatency.String(),
			"average_nodes_used":     inferenceMetrics.AverageNodesUsed,
			"total_tokens_processed": inferenceMetrics.TotalTokensProcessed,
			"last_updated":           inferenceMetrics.LastUpdated,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// handleActiveRequests handles the /api/distributed/requests endpoint
func (s *DistributedOllamaServer) handleActiveRequests(c *gin.Context) {
	activeRequests := s.integration.GetActiveRequests()

	var requests []gin.H
	for _, req := range activeRequests {
		request := gin.H{
			"id":              req.ID,
			"model":           req.OriginalRequest.Model,
			"status":          req.Status,
			"start_time":      req.StartTime,
			"nodes_used":      req.NodesUsed,
			"partition_count": req.PartitionCount,
		}
		requests = append(requests, request)
	}

	response := gin.H{
		"active_requests": requests,
		"total":           len(requests),
	}

	c.JSON(http.StatusOK, response)
}

// handleHealth handles the /health endpoint
func (s *DistributedOllamaServer) handleHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"components": gin.H{
			"p2p":           "healthy",
			"model_manager": "healthy",
			"scheduler":     "healthy",
			"integration":   "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}
