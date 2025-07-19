package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama-distributed/internal/config"
)

// MiddlewareManager handles HTTP middleware for authentication and authorization
type MiddlewareManager struct {
	authManager *Manager
	jwtManager  *JWTManager
	config      *config.AuthConfig
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager(authManager *Manager, jwtManager *JWTManager, config *config.AuthConfig) *MiddlewareManager {
	return &MiddlewareManager{
		authManager: authManager,
		jwtManager:  jwtManager,
		config:      config,
	}
}

// AuthRequired middleware that requires authentication
func (mm *MiddlewareManager) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth if disabled
		if !mm.config.Enabled {
			c.Next()
			return
		}
		
		// Skip auth for certain paths
		if mm.shouldSkipAuth(c.Request.URL.Path, c.Request.Method) {
			c.Next()
			return
		}
		
		// Try to authenticate
		authCtx, err := mm.authenticate(c)
		if err != nil {
			mm.handleAuthError(c, err)
			return
		}
		
		// Store auth context
		mm.setAuthContext(c, authCtx)
		c.Next()
	}
}

// RequirePermission middleware that requires specific permissions
func (mm *MiddlewareManager) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := mm.getAuthContext(c)
		if authCtx == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		if !mm.authManager.HasPermission(authCtx, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required_permission": permission,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireRole middleware that requires a specific role
func (mm *MiddlewareManager) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := mm.getAuthContext(c)
		if authCtx == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		if authCtx.User.Role != role && authCtx.User.Role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient role",
				"required_role": role,
				"user_role": authCtx.User.Role,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (mm *MiddlewareManager) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := mm.getAuthContext(c)
		if authCtx == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		// Admin always has access
		if authCtx.User.Role == RoleAdmin {
			c.Next()
			return
		}
		
		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if authCtx.User.Role == role {
				hasRole = true
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient role",
				"required_roles": roles,
				"user_role": authCtx.User.Role,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAnyPermission middleware that requires any of the specified permissions
func (mm *MiddlewareManager) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := mm.getAuthContext(c)
		if authCtx == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		
		// Check if user has any of the required permissions
		hasPermission := false
		for _, permission := range permissions {
			if mm.authManager.HasPermission(authCtx, permission) {
				hasPermission = true
				break
			}
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
				"required_permissions": permissions,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Optional middleware that attempts authentication but doesn't require it
func (mm *MiddlewareManager) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !mm.config.Enabled {
			c.Next()
			return
		}
		
		// Try to authenticate but don't fail if unsuccessful
		authCtx, _ := mm.authenticate(c)
		if authCtx != nil {
			mm.setAuthContext(c, authCtx)
		}
		
		c.Next()
	}
}

// RateLimit middleware for API rate limiting
func (mm *MiddlewareManager) RateLimit() gin.HandlerFunc {
	// This is a simplified rate limiter
	// In production, use a proper rate limiting library like tollbooth or redis-based limiter
	requestCounts := make(map[string]map[int64]int)
	
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID if authenticated)
		clientID := c.ClientIP()
		if authCtx := mm.getAuthContext(c); authCtx != nil {
			clientID = authCtx.User.ID
		}
		
		// Current minute window
		currentMinute := time.Now().Unix() / 60
		
		if requestCounts[clientID] == nil {
			requestCounts[clientID] = make(map[int64]int)
		}
		
		// Clean old entries
		for minute := range requestCounts[clientID] {
			if currentMinute-minute > 5 { // Keep last 5 minutes
				delete(requestCounts[clientID], minute)
			}
		}
		
		// Count requests in current minute
		requestCounts[clientID][currentMinute]++
		
		// Check limit (100 requests per minute)
		if requestCounts[clientID][currentMinute] > 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": 60,
			})
			c.Abort()
			return
		}
		
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", string(rune(100-requestCounts[clientID][currentMinute])))
		c.Header("X-RateLimit-Reset", string(rune((currentMinute+1)*60)))
		
		c.Next()
	}
}

// CORS middleware with authentication-aware settings
func (mm *MiddlewareManager) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Default allowed origins
		allowedOrigins := []string{"http://localhost:8080", "https://localhost:8080"}
		
		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "3600")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// SecurityHeaders middleware that adds security headers
func (mm *MiddlewareManager) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		c.Next()
	}
}

// AuditLog middleware that logs authentication events
func (mm *MiddlewareManager) AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Log after request completion
		duration := time.Since(start)
		authCtx := mm.getAuthContext(c)
		
		logData := map[string]interface{}{
			"timestamp":    start.Unix(),
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       c.Writer.Status(),
			"duration_ms":  duration.Milliseconds(),
			"ip":           c.ClientIP(),
			"user_agent":   c.Request.Header.Get("User-Agent"),
		}
		
		if authCtx != nil {
			logData["user_id"] = authCtx.User.ID
			logData["username"] = authCtx.User.Username
			logData["auth_method"] = string(authCtx.Method)
			if authCtx.Session != nil {
				logData["session_id"] = authCtx.Session.ID
			}
			if authCtx.APIKey != nil {
				logData["api_key_id"] = authCtx.APIKey.ID
			}
		}
		
		// In production, send this to a proper logging system
		// fmt.Printf("AUDIT: %+v\n", logData)
	}
}

// Helper methods

func (mm *MiddlewareManager) authenticate(c *gin.Context) (*AuthContext, error) {
	// Try API key authentication first
	if apiKey := mm.extractAPIKey(c); apiKey != "" {
		return mm.authManager.ValidateAPIKey(apiKey)
	}
	
	// Try JWT token authentication
	if token := mm.extractBearerToken(c); token != "" {
		return mm.authManager.ValidateToken(token)
	}
	
	return nil, ErrInvalidCredentials
}

func (mm *MiddlewareManager) extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	
	return parts[1]
}

func (mm *MiddlewareManager) extractAPIKey(c *gin.Context) string {
	// Check Authorization header with API key
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "apikey ") {
		return strings.TrimPrefix(authHeader, "ApiKey ")
	}
	
	// Check X-API-Key header
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}
	
	// Check query parameter
	if apiKey := c.Query("api_key"); apiKey != "" {
		return apiKey
	}
	
	return ""
}

func (mm *MiddlewareManager) shouldSkipAuth(path, method string) bool {
	// Public endpoints that don't require authentication
	publicPaths := []string{
		"/api/v1/health",
		"/api/v1/login",
		"/api/v1/register",
		"/metrics",
		"/favicon.ico",
	}
	
	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
		if strings.HasPrefix(path, "/static/") {
			return true
		}
	}
	
	// Always allow OPTIONS requests for CORS
	if method == "OPTIONS" {
		return true
	}
	
	return false
}

func (mm *MiddlewareManager) handleAuthError(c *gin.Context, err error) {
	var status int
	var response gin.H
	
	switch err.(type) {
	case AuthError:
		authErr := err.(AuthError)
		switch authErr.Code {
		case "TOKEN_EXPIRED":
			status = http.StatusUnauthorized
		case "TOKEN_INVALID":
			status = http.StatusUnauthorized
		case "TOKEN_BLACKLISTED":
			status = http.StatusUnauthorized
		case "INSUFFICIENT_PERMISSIONS":
			status = http.StatusForbidden
		case "USER_NOT_FOUND":
			status = http.StatusUnauthorized
		case "USER_INACTIVE":
			status = http.StatusUnauthorized
		default:
			status = http.StatusUnauthorized
		}
		response = gin.H{
			"error": authErr.Message,
			"code":  authErr.Code,
		}
	default:
		status = http.StatusUnauthorized
		response = gin.H{
			"error": "Authentication required",
		}
	}
	
	c.JSON(status, response)
	c.Abort()
}

func (mm *MiddlewareManager) setAuthContext(c *gin.Context, authCtx *AuthContext) {
	c.Set("auth_context", authCtx)
	c.Set("user", authCtx.User)
	c.Set("user_id", authCtx.User.ID)
	c.Set("username", authCtx.User.Username)
	c.Set("role", authCtx.User.Role)
	c.Set("permissions", authCtx.User.Permissions)
	if authCtx.Session != nil {
		c.Set("session", authCtx.Session)
		c.Set("session_id", authCtx.Session.ID)
	}
	if authCtx.APIKey != nil {
		c.Set("api_key", authCtx.APIKey)
		c.Set("api_key_id", authCtx.APIKey.ID)
	}
}

func (mm *MiddlewareManager) getAuthContext(c *gin.Context) *AuthContext {
	if authCtx, exists := c.Get("auth_context"); exists {
		if ctx, ok := authCtx.(*AuthContext); ok {
			return ctx
		}
	}
	return nil
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) *User {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*User); ok {
			return u
		}
	}
	return nil
}

// GetCurrentUserID helper function to get current user ID from context
func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// HasPermission helper function to check permissions in handlers
func HasPermission(c *gin.Context, permission string) bool {
	if permissions, exists := c.Get("permissions"); exists {
		if perms, ok := permissions.([]string); ok {
			for _, perm := range perms {
				if perm == permission || perm == PermissionSystemAdmin {
					return true
				}
			}
		}
	}
	return false
}