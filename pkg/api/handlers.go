package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

// Health check handler
func (s *Server) healthHandler(c *gin.Context) {
	health, err := s.db.Health(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	status := http.StatusOK
	if health.Overall != "healthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":    health.Overall,
		"timestamp": time.Now(),
		"services":  health,
		"version":   "1.0.0",
	})
}

// Metrics handler
func (s *Server) metricsHandler(c *gin.Context) {
	stats := s.db.Stats()
	c.JSON(http.StatusOK, gin.H{
		"database": stats,
		"timestamp": time.Now(),
	})
}

// Authentication handlers
func (s *Server) loginHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Authenticate user
	user, err := s.db.Users.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_failed",
			"message": "Invalid username or password",
		})
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.jwtSvc.GenerateTokens(user.ID.String(), user.Username, user.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_generation_failed",
			"message": "Failed to generate authentication tokens",
		})
		return
	}

	// Create session
	session := &database.UserSession{
		UserID:           user.ID,
		TokenID:          accessToken[:32], // Use first 32 chars as token ID
		ExpiresAt:        time.Now().Add(s.config.Auth.TokenExpiry),
		IPAddress:        &c.ClientIP,
		UserAgent:        &c.Request.UserAgent,
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}

	if err := s.db.Sessions.Create(c.Request.Context(), session); err != nil {
		s.logger.Error("Failed to create session", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(s.config.Auth.TokenExpiry.Seconds()),
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		},
	})
}

func (s *Server) registerHandler(c *gin.Context) {
	var req struct {
		Username string   `json:"username" binding:"required,min=3,max=50"`
		Email    string   `json:"email" binding:"required,email"`
		Password string   `json:"password" binding:"required,min=8"`
		Roles    []string `json:"roles,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Default roles if none provided
	if len(req.Roles) == 0 {
		req.Roles = []string{"user"}
	}

	// Create user
	user := &database.User{
		Username: req.Username,
		Email:    &req.Email,
		Roles:    database.StringArray(req.Roles),
		Active:   true,
	}

	if err := s.db.Users.Create(c.Request.Context(), user, req.Password); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "registration_failed",
			"message": "Username or email already exists",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		},
	})
}

func (s *Server) refreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// Validate refresh token and get user info
	claims, err := s.jwtSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid_refresh_token",
			"message": "Refresh token is invalid or expired",
		})
		return
	}

	// Get user from database
	userID, _ := uuid.Parse(claims.Subject)
	user, err := s.db.Users.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := s.jwtSvc.GenerateTokens(user.ID.String(), user.Username, user.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_generation_failed",
			"message": "Failed to generate new tokens",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(s.config.Auth.TokenExpiry.Seconds()),
	})
}

func (s *Server) logoutHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	// Revoke all user sessions
	uid, _ := uuid.Parse(userID.(string))
	if err := s.db.Sessions.RevokeUserSessions(c.Request.Context(), uid); err != nil {
		s.logger.Error("Failed to revoke user sessions", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// User management handlers
func (s *Server) getUserProfileHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	uid, _ := uuid.Parse(userID.(string))
	user, err := s.db.Users.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"roles":      user.Roles,
			"active":     user.Active,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (s *Server) updateUserProfileHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	var req struct {
		Email *string `json:"email,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	uid, _ := uuid.Parse(userID.(string))
	user, err := s.db.Users.GetByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	// Update fields
	if req.Email != nil {
		user.Email = req.Email
	}

	if err := s.db.Users.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": "Failed to update user profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"roles":    user.Roles,
		},
	})
}

// Model management handlers
func (s *Server) listModelsHandler(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")

	filters := &database.ModelFilters{
		Limit:  limit,
		Offset: offset,
	}
	if status != "" {
		filters.Status = &status
	}

	models, err := s.db.Models.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_failed",
			"message": "Failed to list models",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(models),
		},
	})
}

func (s *Server) createModelHandler(c *gin.Context) {
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Version     string                 `json:"version" binding:"required"`
		Size        int64                  `json:"size" binding:"required,min=1"`
		Hash        string                 `json:"hash" binding:"required"`
		ContentType string                 `json:"content_type"`
		Description *string                `json:"description,omitempty"`
		Tags        []interface{}          `json:"tags,omitempty"`
		Parameters  map[string]interface{} `json:"parameters,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	model := &database.Model{
		Name:        req.Name,
		Version:     req.Version,
		Size:        req.Size,
		Hash:        req.Hash,
		ContentType: req.ContentType,
		Description: req.Description,
		Tags:        database.JSONArray(req.Tags),
		Parameters:  database.JSONMap(req.Parameters),
		Status:      "pending",
	}

	if err := s.db.Models.Create(c.Request.Context(), model); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "creation_failed",
			"message": "Model already exists or creation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Model created successfully",
		"model":   model,
	})
}

func (s *Server) getModelHandler(c *gin.Context) {
	id := c.Param("id")
	modelID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid model ID format",
		})
		return
	}

	model, err := s.db.Models.GetByID(c.Request.Context(), modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "model_not_found",
			"message": "Model not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model": model,
	})
}

func (s *Server) updateModelHandler(c *gin.Context) {
	id := c.Param("id")
	modelID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid model ID format",
		})
		return
	}

	var req struct {
		Description *string                `json:"description,omitempty"`
		Tags        []interface{}          `json:"tags,omitempty"`
		Parameters  map[string]interface{} `json:"parameters,omitempty"`
		Status      *string                `json:"status,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	model, err := s.db.Models.GetByID(c.Request.Context(), modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "model_not_found",
			"message": "Model not found",
		})
		return
	}

	// Update fields
	if req.Description != nil {
		model.Description = req.Description
	}
	if req.Tags != nil {
		model.Tags = database.JSONArray(req.Tags)
	}
	if req.Parameters != nil {
		model.Parameters = database.JSONMap(req.Parameters)
	}
	if req.Status != nil {
		model.Status = *req.Status
	}

	if err := s.db.Models.Update(c.Request.Context(), model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update_failed",
			"message": "Failed to update model",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model updated successfully",
		"model":   model,
	})
}

func (s *Server) deleteModelHandler(c *gin.Context) {
	id := c.Param("id")
	modelID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid model ID format",
		})
		return
	}

	if err := s.db.Models.Delete(c.Request.Context(), modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "deletion_failed",
			"message": "Failed to delete model",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model deleted successfully",
	})
}

func (s *Server) getModelReplicasHandler(c *gin.Context) {
	id := c.Param("id")
	modelID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_id",
			"message": "Invalid model ID format",
		})
		return
	}

	replicas, err := s.db.Models.GetReplicasByModelID(c.Request.Context(), modelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "fetch_failed",
			"message": "Failed to fetch model replicas",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"replicas": replicas,
	})
}
