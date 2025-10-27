package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

// JSON metrics handler for backward compatibility
func (s *Server) metricsJSONHandler(c *gin.Context) {
	stats := s.db.Stats()
	c.JSON(http.StatusOK, gin.H{
		"database": stats,
		"timestamp": time.Now(),
	})
}

// Authentication handlers
func (s *Server) loginHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("ollamamax-api")

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

	// Create span for user authentication
	ctx, authSpan := tracer.Start(ctx, "db.users.authenticate",
		trace.WithAttributes(
			attribute.String("db.operation", "authenticate"),
			attribute.String("db.table", "users"),
			attribute.String("username", req.Username),
		),
	)
	defer authSpan.End()

	// Authenticate user
	user, err := s.db.Users.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		authSpan.SetStatus(codes.Error, "Authentication failed")
		authSpan.RecordError(err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication_failed",
			"message": "Invalid username or password",
		})
		return
	}

	authSpan.SetAttributes(attribute.String("user.id", user.ID.String()))
	authSpan.SetStatus(codes.Ok, "Authentication successful")

	// Generate JWT tokens
	ctx, tokenSpan := tracer.Start(ctx, "jwt.generate_tokens",
		trace.WithAttributes(
			attribute.String("user.id", user.ID.String()),
			attribute.String("username", user.Username),
		),
	)
	accessToken, refreshToken, err := s.jwtSvc.GenerateTokens(user.ID.String(), user.Username, user.Roles)
	tokenSpan.End()

	if err != nil {
		tokenSpan.SetStatus(codes.Error, "Token generation failed")
		tokenSpan.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_generation_failed",
			"message": "Failed to generate authentication tokens",
		})
		return
	}

	// Create session with span
	ctx, sessionSpan := tracer.Start(ctx, "db.sessions.create",
		trace.WithAttributes(
			attribute.String("db.operation", "create"),
			attribute.String("db.table", "sessions"),
			attribute.String("user.id", user.ID.String()),
		),
	)

	session := &database.UserSession{
		UserID:           user.ID,
		TokenID:          accessToken[:32], // Use first 32 chars as token ID
		ExpiresAt:        time.Now().Add(s.config.Auth.TokenExpiry),
		IPAddress:        &c.ClientIP,
		UserAgent:        &c.Request.UserAgent,
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}

	if err := s.db.Sessions.Create(ctx, session); err != nil {
		sessionSpan.SetStatus(codes.Error, "Session creation failed")
		sessionSpan.RecordError(err)
		s.logger.Error("Failed to create session", "error", err)
	} else {
		sessionSpan.SetStatus(codes.Ok, "Session created")
	}
	sessionSpan.End()

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
	ctx := c.Request.Context()
	tracer := otel.Tracer("ollamamax-api")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
		return
	}

	uid, _ := uuid.Parse(userID.(string))

	// Create span for database query
	ctx, dbSpan := tracer.Start(ctx, "db.users.get_by_id",
		trace.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", uid.String()),
		),
	)
	defer dbSpan.End()

	user, err := s.db.Users.GetByID(ctx, uid)
	if err != nil {
		dbSpan.SetStatus(codes.Error, "User not found")
		dbSpan.RecordError(err)
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "user_not_found",
			"message": "User not found",
		})
		return
	}

	dbSpan.SetStatus(codes.Ok, "User retrieved successfully")
	dbSpan.SetAttributes(attribute.String("user.username", user.Username))

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
	ctx := c.Request.Context()
	tracer := otel.Tracer("ollamamax-api")

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

	// Create span for database query
	ctx, dbSpan := tracer.Start(ctx, "db.models.list",
		trace.WithAttributes(
			attribute.String("db.operation", "select"),
			attribute.String("db.table", "models"),
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer dbSpan.End()

	if status != "" {
		dbSpan.SetAttributes(attribute.String("query.status", status))
	}

	models, err := s.db.Models.List(ctx, filters)
	if err != nil {
		dbSpan.SetStatus(codes.Error, "Failed to list models")
		dbSpan.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "list_failed",
			"message": "Failed to list models",
		})
		return
	}

	dbSpan.SetStatus(codes.Ok, "Models retrieved successfully")
	dbSpan.SetAttributes(attribute.Int("result.count", len(models)))

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
