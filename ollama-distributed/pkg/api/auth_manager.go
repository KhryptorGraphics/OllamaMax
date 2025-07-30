package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig configures the authentication manager
type AuthConfig struct {
	JWTSecret   string
	TokenExpiry time.Duration
	Issuer      string
	Audience    string
}

// AuthMetrics tracks authentication performance
type AuthMetrics struct {
	AuthAttempts    int64     `json:"auth_attempts"`
	AuthSuccess     int64     `json:"auth_success"`
	AuthFailures    int64     `json:"auth_failures"`
	ActiveSessions  int64     `json:"active_sessions"`
	TokensIssued    int64     `json:"tokens_issued"`
	TokensRevoked   int64     `json:"tokens_revoked"`
	LastUpdated     time.Time `json:"last_updated"`
	mu              sync.RWMutex
}

// Session represents an authenticated session
type Session struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	LastAccess  time.Time              `json:"last_access"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	PasswordHash string   `json:"password_hash"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLogin   time.Time `json:"last_login"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	jwt.RegisteredClaims
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig) (*AuthManager, error) {
	if config == nil {
		config = &AuthConfig{
			TokenExpiry: 24 * time.Hour,
			Issuer:      "ollama-distributed",
			Audience:    "ollama-api",
		}
	}
	
	// Generate JWT secret if not provided
	if config.JWTSecret == "" {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("failed to generate JWT secret: %w", err)
		}
		config.JWTSecret = string(secret)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &AuthManager{
		config:      config,
		jwtSecret:   []byte(config.JWTSecret),
		tokenExpiry: config.TokenExpiry,
		sessions:    make(map[string]*Session),
		metrics: &AuthMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return manager, nil
}

// Start starts the authentication manager
func (am *AuthManager) Start() error {
	// Start session cleanup
	am.wg.Add(1)
	go am.sessionCleanupLoop()
	
	// Start metrics collection
	am.wg.Add(1)
	go am.metricsLoop()
	
	return nil
}

// Stop stops the authentication manager
func (am *AuthManager) Stop() error {
	am.cancel()
	am.wg.Wait()
	return nil
}

// Authenticate authenticates a user with username and password
func (am *AuthManager) Authenticate(username, password string) (*Session, string, error) {
	am.metrics.mu.Lock()
	am.metrics.AuthAttempts++
	am.metrics.mu.Unlock()
	
	// TODO: Implement actual user authentication
	// For now, use a simple check
	if username == "admin" && password == "admin" {
		session, token, err := am.createSession(username, []string{"admin"}, []string{"*"})
		if err != nil {
			am.metrics.mu.Lock()
			am.metrics.AuthFailures++
			am.metrics.mu.Unlock()
			return nil, "", err
		}
		
		am.metrics.mu.Lock()
		am.metrics.AuthSuccess++
		am.metrics.TokensIssued++
		am.metrics.mu.Unlock()
		
		return session, token, nil
	}
	
	am.metrics.mu.Lock()
	am.metrics.AuthFailures++
	am.metrics.mu.Unlock()
	
	return nil, "", fmt.Errorf("invalid credentials")
}

// ValidateToken validates a JWT token and returns the session
func (am *AuthManager) ValidateToken(tokenString string) (*Session, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	// Check if session exists
	am.sessionsMu.RLock()
	session, exists := am.sessions[claims.SessionID]
	am.sessionsMu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	
	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		am.revokeSession(session.ID)
		return nil, fmt.Errorf("session expired")
	}
	
	// Update last access
	session.LastAccess = time.Now()
	
	return session, nil
}

// createSession creates a new authenticated session
func (am *AuthManager) createSession(username string, roles, permissions []string) (*Session, string, error) {
	sessionID := generateSessionID()
	
	session := &Session{
		ID:          sessionID,
		UserID:      username, // TODO: Use actual user ID
		Username:    username,
		Roles:       roles,
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(am.tokenExpiry),
		LastAccess:  time.Now(),
		Metadata:    make(map[string]interface{}),
	}
	
	// Create JWT token
	claims := &Claims{
		UserID:      session.UserID,
		Username:    session.Username,
		Roles:       session.Roles,
		Permissions: session.Permissions,
		SessionID:   session.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(session.CreatedAt),
			NotBefore: jwt.NewNumericDate(session.CreatedAt),
			Issuer:    am.config.Issuer,
			Audience:  []string{am.config.Audience},
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	// Store session
	am.sessionsMu.Lock()
	am.sessions[session.ID] = session
	am.sessionsMu.Unlock()
	
	// Update metrics
	am.metrics.mu.Lock()
	am.metrics.ActiveSessions++
	am.metrics.LastUpdated = time.Now()
	am.metrics.mu.Unlock()
	
	return session, tokenString, nil
}

// RevokeSession revokes a session
func (am *AuthManager) RevokeSession(sessionID string) error {
	return am.revokeSession(sessionID)
}

// revokeSession revokes a session (internal)
func (am *AuthManager) revokeSession(sessionID string) error {
	am.sessionsMu.Lock()
	defer am.sessionsMu.Unlock()
	
	if _, exists := am.sessions[sessionID]; exists {
		delete(am.sessions, sessionID)
		
		am.metrics.mu.Lock()
		am.metrics.ActiveSessions--
		am.metrics.TokensRevoked++
		am.metrics.LastUpdated = time.Now()
		am.metrics.mu.Unlock()
		
		return nil
	}
	
	return fmt.Errorf("session not found")
}

// GetSession returns a session by ID
func (am *AuthManager) GetSession(sessionID string) (*Session, bool) {
	am.sessionsMu.RLock()
	defer am.sessionsMu.RUnlock()
	
	session, exists := am.sessions[sessionID]
	return session, exists
}

// GetActiveSessions returns all active sessions
func (am *AuthManager) GetActiveSessions() []*Session {
	am.sessionsMu.RLock()
	defer am.sessionsMu.RUnlock()
	
	sessions := make([]*Session, 0, len(am.sessions))
	for _, session := range am.sessions {
		sessions = append(sessions, session)
	}
	
	return sessions
}

// GetMetrics returns authentication metrics
func (am *AuthManager) GetMetrics() *AuthMetrics {
	am.metrics.mu.RLock()
	defer am.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *am.metrics
	return &metrics
}

// Middleware returns a Gin middleware for authentication
func (am *AuthManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		
		// Extract token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}
		
		// Validate token
		session, err := am.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		
		// Store session in context
		c.Set("session", session)
		c.Set("user_id", session.UserID)
		c.Set("username", session.Username)
		c.Set("roles", session.Roles)
		c.Set("permissions", session.Permissions)
		
		c.Next()
	}
}

// RequireRole returns a middleware that requires specific roles
func (am *AuthManager) RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, exists := c.Get("session")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no session found"})
			c.Abort()
			return
		}
		
		userSession := session.(*Session)
		
		// Check if user has any of the required roles
		hasRole := false
		for _, userRole := range userSession.Roles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole || userRole == "admin" {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		
		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// sessionCleanupLoop cleans up expired sessions
func (am *AuthManager) sessionCleanupLoop() {
	defer am.wg.Done()
	
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (am *AuthManager) cleanupExpiredSessions() {
	am.sessionsMu.Lock()
	defer am.sessionsMu.Unlock()
	
	now := time.Now()
	var expiredSessions []string
	
	for id, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, id)
		}
	}
	
	for _, id := range expiredSessions {
		delete(am.sessions, id)
		
		am.metrics.mu.Lock()
		am.metrics.ActiveSessions--
		am.metrics.LastUpdated = time.Now()
		am.metrics.mu.Unlock()
	}
}

// metricsLoop runs the metrics collection loop
func (am *AuthManager) metricsLoop() {
	defer am.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.updateMetrics()
		}
	}
}

// updateMetrics updates authentication metrics
func (am *AuthManager) updateMetrics() {
	am.metrics.mu.Lock()
	defer am.metrics.mu.Unlock()
	
	am.sessionsMu.RLock()
	am.metrics.ActiveSessions = int64(len(am.sessions))
	am.sessionsMu.RUnlock()
	
	am.metrics.LastUpdated = time.Now()
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
