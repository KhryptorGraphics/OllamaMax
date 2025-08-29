package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

// Node management handlers

func (s *Server) listNodesHandler(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	region := c.Query("region")
	healthyOnly := c.DefaultQuery("healthy_only", "false") == "true"

	filters := &database.NodeFilters{
		Limit:       limit,
		Offset:      offset,
		HealthyOnly: healthyOnly,
	}
	if status != "" {
		filters.Status = &status
	}
	if region != "" {
		filters.Region = &region
	}

	nodes, err := s.db.Nodes.List(c.Request.Context(), filters)
	if err != nil {
		s.logger.Error("Failed to list nodes", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_failed",
			"message": "Failed to list nodes",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(nodes),
		},
	})
}

func (s *Server) getNodeHandler(c *gin.Context) {
	id := c.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid node ID format",
		})
		return
	}

	node, err := s.db.Nodes.GetByID(c.Request.Context(), nodeID)
	if err != nil {
		s.logger.Error("Failed to get node", "node_id", nodeID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "node_not_found",
			"message": "Node not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node": node,
	})
}

func (s *Server) updateNodeHandler(c *gin.Context) {
	id := c.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid node ID format",
		})
		return
	}

	var req struct {
		Name         *string                `json:"name,omitempty"`
		Region       *string                `json:"region,omitempty"`
		Zone         *string                `json:"zone,omitempty"`
		Status       *string                `json:"status,omitempty"`
		Capabilities map[string]interface{} `json:"capabilities,omitempty"`
		Resources    map[string]interface{} `json:"resources,omitempty"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	node, err := s.db.Nodes.GetByID(c.Request.Context(), nodeID)
	if err != nil {
		s.logger.Error("Failed to get node for update", "node_id", nodeID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "node_not_found",
			"message": "Node not found",
		})
		return
	}

	// Update fields
	if req.Name != nil {
		node.Name = req.Name
	}
	if req.Region != nil {
		node.Region = req.Region
	}
	if req.Zone != nil {
		node.Zone = req.Zone
	}
	if req.Status != nil {
		node.Status = *req.Status
	}
	if req.Capabilities != nil {
		node.Capabilities = database.JSONMap(req.Capabilities)
	}
	if req.Resources != nil {
		node.Resources = database.JSONMap(req.Resources)
	}
	if req.Metadata != nil {
		node.Metadata = database.JSONMap(req.Metadata)
	}

	if err := s.db.Nodes.Update(c.Request.Context(), node); err != nil {
		s.logger.Error("Failed to update node", "node_id", nodeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": "Failed to update node",
		})
		return
	}

	// Broadcast node status update via WebSocket
	s.websocket.BroadcastNodeStatus(nodeID, node.Status, gin.H{
		"name":   node.Name,
		"region": node.Region,
		"zone":   node.Zone,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Node updated successfully",
		"node":    node,
	})
}

func (s *Server) deleteNodeHandler(c *gin.Context) {
	id := c.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid node ID format",
		})
		return
	}

	if err := s.db.Nodes.Delete(c.Request.Context(), nodeID); err != nil {
		s.logger.Error("Failed to delete node", "node_id", nodeID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "deletion_failed",
			"message": "Failed to delete node",
		})
		return
	}

	// Broadcast node removal via WebSocket
	s.websocket.BroadcastNodeStatus(nodeID, "deleted", gin.H{
		"action": "node_removed",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Node deleted successfully",
	})
}

func (s *Server) getNodeHealthHandler(c *gin.Context) {
	id := c.Param("id")
	nodeID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid node ID format",
		})
		return
	}

	node, err := s.db.Nodes.GetByID(c.Request.Context(), nodeID)
	if err != nil {
		s.logger.Error("Failed to get node for health check", "node_id", nodeID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "node_not_found",
			"message": "Node not found",
		})
		return
	}

	// Calculate health status based on last heartbeat and status
	healthStatus := "unhealthy"
	healthScore := 0.0

	if node.Status == "active" && node.LastHeartbeat != nil {
		timeSinceHeartbeat := time.Since(*node.LastHeartbeat)
		if timeSinceHeartbeat < 5*time.Minute {
			healthStatus = "healthy"
			// Calculate health score based on heartbeat freshness
			if timeSinceHeartbeat < 1*time.Minute {
				healthScore = 1.0
			} else {
				// Linearly decrease score from 1.0 to 0.2 over 5 minutes
				healthScore = 1.0 - (0.8 * (timeSinceHeartbeat.Seconds() / (5 * 60)))
			}
		} else if timeSinceHeartbeat < 10*time.Minute {
			healthStatus = "degraded"
			healthScore = 0.1
		}
	}

	// Get model replicas hosted on this node
	replicas, err := s.db.Models.GetReplicasByNodeID(c.Request.Context(), nodeID)
	if err != nil {
		s.logger.Error("Failed to get node replicas", "node_id", nodeID, "error", err)
		// Continue without replica information
		replicas = []*database.ModelReplica{}
	}

	// Calculate replica health
	readyReplicas := 0
	totalReplicas := len(replicas)
	for _, replica := range replicas {
		if replica.Status == "ready" && replica.HealthScore > 0.7 {
			readyReplicas++
		}
	}

	replicaHealthRatio := 0.0
	if totalReplicas > 0 {
		replicaHealthRatio = float64(readyReplicas) / float64(totalReplicas)
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id":     nodeID,
		"status":      healthStatus,
		"score":       healthScore,
		"last_seen":   node.LastHeartbeat,
		"node_status": node.Status,
		"replicas": gin.H{
			"total":       totalReplicas,
			"ready":       readyReplicas,
			"health_ratio": replicaHealthRatio,
		},
		"resources":    node.Resources,
		"capabilities": node.Capabilities,
		"metadata":     node.Metadata,
	})
}

// System management handlers

func (s *Server) getSystemConfigHandler(c *gin.Context) {
	configs, err := s.db.Config.GetAll(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get system config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "config_fetch_failed",
			"message": "Failed to fetch system configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": configs,
	})
}

func (s *Server) updateSystemConfigHandler(c *gin.Context) {
	var req struct {
		Config map[string]interface{} `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Get user ID for audit trail
	userID, exists := c.Get("user_id")
	var userUUID *uuid.UUID
	if exists {
		if uid, err := uuid.Parse(userID.(string)); err == nil {
			userUUID = &uid
		}
	}

	// Update each configuration item
	for key, value := range req.Config {
		config := &database.SystemConfig{
			Key:       key,
			Value:     database.JSONValue{"value": value},
			UpdatedBy: userUUID,
		}

		if err := s.db.Config.Set(c.Request.Context(), config); err != nil {
			s.logger.Error("Failed to update config", "key", key, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "config_update_failed",
				"message": fmt.Sprintf("Failed to update configuration: %s", key),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "System configuration updated successfully",
	})
}

func (s *Server) getSystemStatsHandler(c *gin.Context) {
	// Get database statistics
	dbStats := s.db.Stats()

	// Get WebSocket connection count
	wsConnections := s.websocket.GetConnectedClients()

	// Get system health
	health, err := s.db.Health(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get system health", "error", err)
		health = &database.HealthStatus{
			Overall: "unknown",
		}
	}

	// Get dashboard statistics (if materialized view exists)
	dashboardStats, err := s.db.GetDashboardStats(c.Request.Context())
	if err != nil {
		s.logger.Warn("Failed to get dashboard stats", "error", err)
		dashboardStats = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"system": gin.H{
			"health":            health,
			"websocket_clients": wsConnections,
			"database":          dbStats,
			"dashboard":         dashboardStats,
		},
		"timestamp": time.Now(),
	})
}

func (s *Server) getAuditLogsHandler(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	tableName := c.Query("table")
	operation := c.Query("operation")
	userIDStr := c.Query("user_id")

	var userID *uuid.UUID
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	filters := &database.AuditFilters{
		TableName: &tableName,
		Operation: &operation,
		UserID:    userID,
		Limit:     limit,
		Offset:    offset,
	}

	auditLogs, err := s.db.Audit.List(c.Request.Context(), filters)
	if err != nil {
		s.logger.Error("Failed to get audit logs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "audit_fetch_failed",
			"message": "Failed to fetch audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": auditLogs,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(auditLogs),
		},
	})
}

// Inference handlers

func (s *Server) chatHandler(c *gin.Context) {
	var req struct {
		ModelName string                 `json:"model" binding:"required"`
		Messages  []map[string]string    `json:"messages" binding:"required"`
		Options   map[string]interface{} `json:"options,omitempty"`
		Stream    bool                   `json:"stream,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Get user ID for inference tracking
	userID, exists := c.Get("user_id")
	var userUUID *uuid.UUID
	if exists {
		if uid, err := uuid.Parse(userID.(string)); err == nil {
			userUUID = &uid
		}
	}

	// Find the model
	model, err := s.db.Models.GetByName(c.Request.Context(), req.ModelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "model_not_found",
			"message": fmt.Sprintf("Model not found: %s", req.ModelName),
		})
		return
	}

	// Create inference request
	inferenceReq := &database.InferenceRequest{
		RequestID: uuid.New().String(),
		UserID:    userUUID,
		ModelID:   model.ID,
		ModelName: req.ModelName,
		Status:    "pending",
		Metadata: database.JSONMap{
			"type":     "chat",
			"messages": req.Messages,
			"options":  req.Options,
			"stream":   req.Stream,
		},
	}

	if err := s.db.Inference.Create(c.Request.Context(), inferenceReq); err != nil {
		s.logger.Error("Failed to create inference request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "inference_creation_failed",
			"message": "Failed to create inference request",
		})
		return
	}

	// Broadcast inference start via WebSocket
	s.websocket.BroadcastInferenceUpdate(inferenceReq.ID, "started", gin.H{
		"model":      req.ModelName,
		"request_id": inferenceReq.RequestID,
	})

	if req.Stream {
		// Handle streaming response
		c.JSON(http.StatusAccepted, gin.H{
			"request_id": inferenceReq.RequestID,
			"status":     "streaming",
			"websocket":  fmt.Sprintf("/ws/inference/%s", inferenceReq.ID.String()),
		})
	} else {
		// Handle non-streaming response
		c.JSON(http.StatusAccepted, gin.H{
			"request_id": inferenceReq.RequestID,
			"status":     "processing",
		})
	}
}

func (s *Server) generateHandler(c *gin.Context) {
	var req struct {
		ModelName string                 `json:"model" binding:"required"`
		Prompt    string                 `json:"prompt" binding:"required"`
		Options   map[string]interface{} `json:"options,omitempty"`
		Stream    bool                   `json:"stream,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Get user ID for inference tracking
	userID, exists := c.Get("user_id")
	var userUUID *uuid.UUID
	if exists {
		if uid, err := uuid.Parse(userID.(string)); err == nil {
			userUUID = &uid
		}
	}

	// Find the model
	model, err := s.db.Models.GetByName(c.Request.Context(), req.ModelName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "model_not_found",
			"message": fmt.Sprintf("Model not found: %s", req.ModelName),
		})
		return
	}

	// Create inference request
	inferenceReq := &database.InferenceRequest{
		RequestID:    uuid.New().String(),
		UserID:       userUUID,
		ModelID:      model.ID,
		ModelName:    req.ModelName,
		PromptLength: &len(req.Prompt),
		Status:       "pending",
		Metadata: database.JSONMap{
			"type":    "generate",
			"prompt":  req.Prompt,
			"options": req.Options,
			"stream":  req.Stream,
		},
	}

	if err := s.db.Inference.Create(c.Request.Context(), inferenceReq); err != nil {
		s.logger.Error("Failed to create inference request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "inference_creation_failed",
			"message": "Failed to create inference request",
		})
		return
	}

	// Broadcast inference start via WebSocket
	s.websocket.BroadcastInferenceUpdate(inferenceReq.ID, "started", gin.H{
		"model":      req.ModelName,
		"request_id": inferenceReq.RequestID,
	})

	if req.Stream {
		// Handle streaming response
		c.JSON(http.StatusAccepted, gin.H{
			"request_id": inferenceReq.RequestID,
			"status":     "streaming",
			"websocket":  fmt.Sprintf("/ws/inference/%s", inferenceReq.ID.String()),
		})
	} else {
		// Handle non-streaming response
		c.JSON(http.StatusAccepted, gin.H{
			"request_id": inferenceReq.RequestID,
			"status":     "processing",
		})
	}
}

func (s *Server) listInferenceRequestsHandler(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")
	modelName := c.Query("model")
	userIDStr := c.Query("user_id")

	var userID *uuid.UUID
	if userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			userID = &uid
		}
	}

	filters := &database.InferenceFilters{
		Status:    &status,
		ModelName: &modelName,
		UserID:    userID,
		Limit:     limit,
		Offset:    offset,
	}

	requests, err := s.db.Inference.List(c.Request.Context(), filters)
	if err != nil {
		s.logger.Error("Failed to list inference requests", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_failed",
			"message": "Failed to list inference requests",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(requests),
		},
	})
}

func (s *Server) getInferenceRequestHandler(c *gin.Context) {
	id := c.Param("id")
	requestID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid request ID format",
		})
		return
	}

	request, err := s.db.Inference.GetByID(c.Request.Context(), requestID)
	if err != nil {
		s.logger.Error("Failed to get inference request", "request_id", requestID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "request_not_found",
			"message": "Inference request not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"request": request,
	})
}