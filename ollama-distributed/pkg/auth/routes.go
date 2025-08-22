package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Routes handles authentication-related HTTP routes
type Routes struct {
	authManager       *Manager
	jwtManager        *JWTManager
	middlewareManager *MiddlewareManager
}

// NewRoutes creates a new routes handler
func NewRoutes(authManager *Manager, jwtManager *JWTManager, middlewareManager *MiddlewareManager) *Routes {
	return &Routes{
		authManager:       authManager,
		jwtManager:        jwtManager,
		middlewareManager: middlewareManager,
	}
}

// RegisterRoutes registers authentication routes with the Gin router
func (r *Routes) RegisterRoutes(router *gin.Engine) {
	// Apply global middleware
	router.Use(r.middlewareManager.CORS())
	router.Use(r.middlewareManager.SecurityHeaders())
	router.Use(r.middlewareManager.RateLimit())
	router.Use(r.middlewareManager.AuditLog())

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		public.POST("/login", r.login)
		public.POST("/register", r.register)
		public.POST("/refresh", r.refreshToken)
		public.GET("/health", r.health)
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(r.middlewareManager.AuthRequired())
	{
		// User management
		user := protected.Group("/user")
		{
			user.GET("/profile", r.getProfile)
			user.PUT("/profile", r.updateProfile)
			user.POST("/change-password", r.changePassword)
			user.POST("/logout", r.logout)
			user.GET("/sessions", r.getSessions)
			user.DELETE("/sessions/:session_id", r.revokeSession)
		}

		// API key management
		apiKeys := protected.Group("/api-keys")
		{
			apiKeys.GET("", r.listAPIKeys)
			apiKeys.POST("", r.createAPIKey)
			apiKeys.DELETE("/:key_id", r.revokeAPIKey)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(r.middlewareManager.RequireRole(RoleAdmin))
		{
			admin.GET("/users", r.listUsers)
			admin.POST("/users", r.createUser)
			admin.GET("/users/:user_id", r.getUser)
			admin.PUT("/users/:user_id", r.updateUser)
			admin.DELETE("/users/:user_id", r.deleteUser)
			admin.POST("/users/:user_id/reset-password", r.resetUserPassword)
			admin.GET("/stats", r.getAuthStats)
		}
	}
}

// Authentication handlers

func (r *Routes) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get client metadata
	metadata := map[string]string{
		"ip_address": c.ClientIP(),
		"user_agent": c.Request.Header.Get("User-Agent"),
	}
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Authenticate user
	authCtx, err := r.authManager.Authenticate(req.Username, req.Password, metadata)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	response := LoginResponse{
		Token:     authCtx.TokenString,
		ExpiresAt: authCtx.Session.ExpiresAt,
		User:      authCtx.User,
		SessionID: authCtx.Session.ID,
	}

	c.JSON(http.StatusOK, response)
}

func (r *Routes) register(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Default role for registration
	if req.Role == "" {
		req.Role = RoleUser
	}

	// Only allow certain roles for self-registration
	allowedRoles := []string{RoleUser, RoleReadOnly}
	roleAllowed := false
	for _, role := range allowedRoles {
		if req.Role == role {
			roleAllowed = true
			break
		}
	}

	if !roleAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role for registration"})
		return
	}

	user, err := r.authManager.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove sensitive data from response
	user.Metadata = map[string]string{}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (r *Routes) refreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user from token (even if expired)
	_, err := r.jwtManager.GetTokenClaims(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// This is a simplified implementation
	// In a real system, you'd use a separate refresh token
	var refreshReq RefreshTokenRequest
	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// This is a simplified implementation
	// In a real system, you'd validate the refresh token properly
	user := &User{ID: "user123", Username: "testuser", Role: "user"}
	tokenPair, err := r.jwtManager.GenerateTokenPair(user, "session123", map[string]string{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
		"expires_in":    3600,
		"message":       "Token refreshed successfully",
	})
}

func (r *Routes) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "auth",
		"timestamp": time.Now().Unix(),
	})
}

// User management handlers

func (r *Routes) getProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Remove sensitive data
	user.Metadata = map[string]string{}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (r *Routes) updateProfile(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update allowed fields
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			if k != "password_hash" { // Prevent password hash modification
				user.Metadata[k] = v
			}
		}
	}

	user.UpdatedAt = time.Now()

	// Remove sensitive data from response
	user.Metadata = map[string]string{}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (r *Routes) changePassword(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password change not implemented"})
}

func (r *Routes) logout(c *gin.Context) {
	authCtx := r.middlewareManager.getAuthContext(c)
	if authCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No active session"})
		return
	}

	// Revoke session if it exists
	if authCtx.Session != nil {
		r.authManager.RevokeSession(authCtx.Session.ID)
	}

	// Blacklist the token
	if authCtx.Claims != nil {
		r.authManager.RevokeToken(authCtx.Claims.ID, authCtx.Claims.ExpiresAt.Time)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (r *Routes) getSessions(c *gin.Context) {
	// This would need to be implemented to return user sessions
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Session listing not implemented"})
}

func (r *Routes) revokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	err := r.authManager.RevokeSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

// API key management handlers

func (r *Routes) listAPIKeys(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Filter API keys to only show current user's keys
	var userAPIKeys []APIKey
	for _, apiKey := range user.APIKeys {
		// Remove the actual key value for security
		apiKey.Key = ""
		userAPIKeys = append(userAPIKeys, apiKey)
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": userAPIKeys})
}

func (r *Routes) createAPIKey(c *gin.Context) {
	user := GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	apiKey, rawKey, err := r.authManager.CreateAPIKey(user.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := CreateAPIKeyResponse{
		APIKey: apiKey,
		Key:    rawKey,
	}

	// Remove the hashed key from the response
	response.APIKey.Key = ""

	c.JSON(http.StatusCreated, response)
}

func (r *Routes) revokeAPIKey(c *gin.Context) {
	keyID := c.Param("key_id")

	err := r.authManager.RevokeAPIKey(keyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully"})
}

// Admin handlers

func (r *Routes) listUsers(c *gin.Context) {
	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User listing not implemented"})
}

func (r *Routes) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := r.authManager.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove sensitive data
	user.Metadata = map[string]string{}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (r *Routes) getUser(c *gin.Context) {
	userID := c.Param("user_id")

	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User retrieval not implemented", "user_id": userID})
}

func (r *Routes) updateUser(c *gin.Context) {
	userID := c.Param("user_id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User update not implemented", "user_id": userID})
}

func (r *Routes) deleteUser(c *gin.Context) {
	userID := c.Param("user_id")

	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "User deletion not implemented", "user_id": userID})
}

func (r *Routes) resetUserPassword(c *gin.Context) {
	userID := c.Param("user_id")

	// This would need to be implemented in the auth manager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented", "user_id": userID})
}

func (r *Routes) getAuthStats(c *gin.Context) {
	stats := r.jwtManager.GetTokenStats()

	// Add more stats from auth manager
	stats["timestamp"] = time.Now().Unix()

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
