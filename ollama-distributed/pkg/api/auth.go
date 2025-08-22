package api

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SecretKey     string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// AuthMiddleware provides JWT authentication middleware
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for health check and public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		claims, err := s.validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func (s *Server) RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No roles found"})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid roles format"})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole || role == "admin" {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware implements rate limiting
func (s *Server) RateLimitMiddleware() gin.HandlerFunc {
	// Simple in-memory rate limiter
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old entries
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// Check rate limit (100 requests per minute)
		if len(clients[clientIP]) >= 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)
		c.Next()
	}
}

// CORSMiddleware handles CORS headers with secure origin validation
func (s *Server) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from config or default to localhost for development
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080", "https://localhost:8443"}
		if s.config != nil && len(s.config.API.Cors.AllowedOrigins) > 0 {
			allowedOrigins = s.config.API.Cors.AllowedOrigins
		}
		
		origin := c.GetHeader("Origin")
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			// Deny credentials for non-allowed origins
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "false")
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func (s *Server) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// LoggingMiddleware logs requests
func (s *Server) LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// generateToken creates a new JWT token
func (s *Server) generateToken(userID, username string, roles []string) (string, error) {
	// Get JWT secret from config with fallback
	secret := s.getSecretKey()
	
	// Calculate token expiry from config with fallback
	expiry := 24 * time.Hour
	if s.config != nil && s.config.Security.Auth.TokenExpiry > 0 {
		expiry = s.config.Security.Auth.TokenExpiry
	}

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.getJWTIssuer(),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// validateToken validates a JWT token
func (s *Server) validateToken(tokenString string) (*JWTClaims, error) {
	secret := s.getSecretKey()
	
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Verify issuer if configured
		expectedIssuer := s.getJWTIssuer()
		if claims.Issuer != expectedIssuer {
			return nil, fmt.Errorf("invalid token issuer")
		}
		
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// extractToken extracts JWT token from request
func extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Check query parameter
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Check cookie
	cookie, err := c.Cookie("auth_token")
	if err == nil {
		return cookie
	}

	return ""
}

// isPublicEndpoint checks if endpoint is public
func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/health",
		"/api/version",
		"/api/auth/login",
		"/api/auth/register",
		"/api/v1/health",
		"/api/v1/version",
		"/metrics",
		"/",
		"/static/",
	}

	for _, endpoint := range publicEndpoints {
		if strings.HasPrefix(path, endpoint) {
			return true
		}
	}

	return false
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
}

// UserInfo represents user information
type UserInfo struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Avatar      string            `json:"avatar,omitempty"`
	Roles       []string          `json:"roles"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	LastLogin   time.Time         `json:"last_login,omitempty"`
}

// login handles user login
func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := s.ValidateRequest(c, &req); err != nil {
		return
	}

	// Authenticate user (mock implementation)
	user, err := s.authenticateUser(req.Username, req.Password)
	if err != nil {
		s.HandleError(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID, user.Username, user.Roles)
	if err != nil {
		s.HandleError(c, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// Calculate expiry
	expiry := 24 * time.Hour
	if s.config != nil && s.config.Security.Auth.TokenExpiry > 0 {
		expiry = s.config.Security.Auth.TokenExpiry
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(expiry),
		User:      *user,
	}

	s.LogRequest(c, "user_login")
	c.JSON(http.StatusOK, response)
}

// register handles user registration
func (s *Server) register(c *gin.Context) {
	var req RegisterRequest
	if err := s.ValidateRequest(c, &req); err != nil {
		return
	}

	// Check if user already exists
	if s.userExists(req.Username) {
		s.HandleError(c, http.StatusConflict, "User already exists", nil)
		return
	}

	// Create new user (mock implementation)
	user, err := s.createUser(req.Username, req.Password, req.Email)
	if err != nil {
		s.HandleError(c, http.StatusInternalServerError, "Failed to create user", err)
		return
	}

	// Generate JWT token
	token, err := s.generateToken(user.ID, user.Username, user.Roles)
	if err != nil {
		s.HandleError(c, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// Calculate expiry
	expiry := 24 * time.Hour
	if s.config != nil && s.config.Security.Auth.TokenExpiry > 0 {
		expiry = s.config.Security.Auth.TokenExpiry
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(expiry),
		User:      *user,
	}

	s.LogRequest(c, "user_register")
	c.JSON(http.StatusCreated, response)
}

// refreshToken handles token refresh
func (s *Server) refreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := s.ValidateRequest(c, &req); err != nil {
		return
	}

	// Validate refresh token
	claims, err := s.validateToken(req.RefreshToken)
	if err != nil {
		s.HandleError(c, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Generate new access token
	token, err := s.generateToken(claims.UserID, claims.Username, claims.Roles)
	if err != nil {
		s.HandleError(c, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// Calculate expiry
	expiry := 24 * time.Hour
	if s.config != nil && s.config.Security.Auth.TokenExpiry > 0 {
		expiry = s.config.Security.Auth.TokenExpiry
	}

	response := RefreshTokenResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(expiry),
	}

	s.LogRequest(c, "token_refresh")
	c.JSON(http.StatusOK, response)
}

// logout handles user logout
func (s *Server) logout(c *gin.Context) {
	// In a real implementation, you would blacklist the token
	// For now, just return success
	s.LogRequest(c, "user_logout")
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// profile returns user profile information
func (s *Server) getUserProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	username := c.GetString("username")
	roles, _ := c.Get("roles")

	user := UserInfo{
		ID:       userID,
		Username: username,
		Roles:    roles.([]string),
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// updateUserProfile handles profile updates
func (s *Server) updateUserProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UpdateProfileRequest
	if err := s.ValidateRequest(c, &req); err != nil {
		return
	}

	// Update user profile (mock implementation)
	err := s.updateUser(userID, req)
	if err != nil {
		s.HandleError(c, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	s.LogRequest(c, "profile_update")
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// getSecretKey returns the JWT secret from config with secure fallback
func (s *Server) getSecretKey() []byte {
	// Try to get from config
	if s.config != nil && s.config.Security.Auth.SecretKey != "" {
		return []byte(s.config.Security.Auth.SecretKey)
	}
	
	// Try environment variable
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	
	// Generate a random secret for development (NOT for production)
	if s.devJWTSecret == nil {
		s.devJWTSecret = s.generateDevSecret()
		s.logger.Warn("Using generated JWT secret - NOT suitable for production")
	}
	
	return s.devJWTSecret
}

// getJWTIssuer returns the JWT issuer from config with fallback
func (s *Server) getJWTIssuer() string {
	if s.config != nil && s.config.Security.Auth.Issuer != "" {
		return s.config.Security.Auth.Issuer
	}
	return "ollama-distributed"
}

// generateDevSecret generates a cryptographically secure random secret for development
func (s *Server) generateDevSecret() []byte {
	secret := make([]byte, 32)
	// Use crypto/rand for secure random generation
	if _, err := rand.Read(secret); err != nil {
		// Fallback to time-based generation (less secure)
		s.logger.Error("Failed to generate secure random secret, using fallback", "error", err)
		for i := range secret {
			secret[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	return secret
}

// User management methods (mock implementations)

// authenticateUser authenticates a user with username and password
func (s *Server) authenticateUser(username, password string) (*UserInfo, error) {
	// Mock user database - in real implementation, this would query a database
	mockUsers := map[string]*UserInfo{
		"admin": {
			ID:          "admin-001",
			Username:    "admin",
			Email:       "admin@example.com",
			DisplayName: "System Administrator",
			Roles:       []string{"admin", "user"},
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			LastLogin:   time.Now().Add(-1 * time.Hour),
		},
		"user": {
			ID:          "user-001",
			Username:    "user",
			Email:       "user@example.com",
			DisplayName: "Regular User",
			Roles:       []string{"user"},
			CreatedAt:   time.Now().Add(-7 * 24 * time.Hour),
			LastLogin:   time.Now().Add(-30 * time.Minute),
		},
	}

	user, exists := mockUsers[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// SECURITY: Use proper password hashing in production
	// This is a mock implementation - replace with bcrypt validation
	expectedHash := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewgqjqOpVcCd9YLy" // bcrypt hash of "SecurePassword123!"
	if !s.validatePasswordHash(password, expectedHash) {
		return nil, fmt.Errorf("invalid password")
	}

	// Update last login
	user.LastLogin = time.Now()

	return user, nil
}

// userExists checks if a user exists
func (s *Server) userExists(username string) bool {
	// Mock implementation - in real implementation, query database
	mockUsers := []string{"admin", "user", "test"}
	for _, user := range mockUsers {
		if user == username {
			return true
		}
	}
	return false
}

// createUser creates a new user
func (s *Server) createUser(username, password, email string) (*UserInfo, error) {
	// Mock implementation - in real implementation, save to database
	user := &UserInfo{
		ID:          fmt.Sprintf("user-%d", time.Now().UnixNano()),
		Username:    username,
		Email:       email,
		DisplayName: username,
		Roles:       []string{"user"},
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
		LastLogin:   time.Now(),
	}

	s.logger.Info("user created", "username", username, "user_id", user.ID)
	return user, nil
}

// updateUser updates user information
func (s *Server) updateUser(userID string, req UpdateProfileRequest) error {
	// Mock implementation - in real implementation, update database
	s.logger.Info("user profile updated", "user_id", userID, "email", req.Email)
	return nil
}

// validatePasswordHash validates a password against a bcrypt hash
func (s *Server) validatePasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// hashPassword creates a bcrypt hash of a password
func (s *Server) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
