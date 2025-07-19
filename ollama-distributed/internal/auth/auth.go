package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ollama/ollama-distributed/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// Manager handles all authentication operations
type Manager struct {
	config *config.AuthConfig
	
	// JWT signing key
	signingKey []byte
	
	// In-memory stores (in production, these would be backed by persistent storage)
	users          map[string]*User
	apiKeys        map[string]*APIKey
	sessions       map[string]*Session
	blacklistCache map[string]time.Time
	
	// Password hasher
	bcryptCost int
	
	// Mutex for thread safety
	mu sync.RWMutex
	
	// Background cleanup
	stopCleanup chan struct{}
}

// NewManager creates a new authentication manager
func NewManager(cfg *config.AuthConfig) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("auth config is required")
	}
	
	// Generate or use provided signing key
	signingKey := []byte(cfg.SecretKey)
	if len(signingKey) == 0 {
		// Generate a random signing key
		signingKey = make([]byte, 32)
		if _, err := rand.Read(signingKey); err != nil {
			return nil, fmt.Errorf("failed to generate signing key: %w", err)
		}
	}
	
	manager := &Manager{
		config:         cfg,
		signingKey:     signingKey,
		users:          make(map[string]*User),
		apiKeys:        make(map[string]*APIKey),
		sessions:       make(map[string]*Session),
		blacklistCache: make(map[string]time.Time),
		bcryptCost:     bcrypt.DefaultCost,
		stopCleanup:    make(chan struct{}),
	}
	
	// Create default admin user if none exists
	if err := manager.createDefaultAdmin(); err != nil {
		return nil, fmt.Errorf("failed to create default admin: %w", err)
	}
	
	// Start background cleanup routines
	go manager.cleanupExpiredSessions()
	go manager.cleanupBlacklist()
	
	return manager, nil
}

// Close gracefully shuts down the auth manager
func (m *Manager) Close() {
	close(m.stopCleanup)
}

// createDefaultAdmin creates a default admin user if no users exist
func (m *Manager) createDefaultAdmin() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if any users exist
	if len(m.users) > 0 {
		return nil
	}
	
	// Create default admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), m.bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash default password: %w", err)
	}
	
	adminUser := &User{
		ID:          generateID(),
		Username:    "admin",
		Email:       "admin@localhost",
		Role:        RoleAdmin,
		Permissions: DefaultRolePermissions[RoleAdmin],
		Metadata: map[string]string{
			"password_hash": string(hashedPassword),
			"created_by":    "system",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}
	
	m.users[adminUser.ID] = adminUser
	
	fmt.Printf("Created default admin user (username: admin, password: admin123)\n")
	fmt.Printf("WARNING: Please change the default password immediately!\n")
	
	return nil
}

// Authenticate validates credentials and returns an auth context
func (m *Manager) Authenticate(username, password string, metadata map[string]string) (*AuthContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Find user by username
	var user *User
	for _, u := range m.users {
		if u.Username == username && u.Active {
			user = u
			break
		}
	}
	
	if user == nil {
		return nil, ErrInvalidCredentials
	}
	
	// Verify password
	passwordHash := user.Metadata["password_hash"]
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	
	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	
	// Create session
	session := &Session{
		ID:        generateID(),
		UserID:    user.ID,
		IPAddress: metadata["ip_address"],
		UserAgent: metadata["user_agent"],
		Metadata:  metadata,
		CreatedAt: now,
		ExpiresAt: now.Add(m.config.TokenExpiry),
		Active:    true,
	}
	
	m.sessions[session.ID] = session
	
	// Generate JWT token
	claims := &Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: user.Permissions,
		SessionID:   session.ID,
		Metadata:    user.Metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(session.ExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Subject:   user.ID,
			ID:        generateID(),
			Audience:  []string{m.config.Audience},
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.signingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}
	
	session.TokenID = claims.ID
	
	return &AuthContext{
		User:        user,
		Session:     session,
		Claims:      claims,
		TokenString: tokenString,
		Method:      AuthMethodJWT,
	}, nil
}

// ValidateToken validates a JWT token and returns the auth context
func (m *Manager) ValidateToken(tokenString string) (*AuthContext, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.signingKey, nil
	})
	
	if err != nil {
		return nil, ErrTokenInvalid
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	
	// Check if token is blacklisted
	if m.isTokenBlacklisted(claims.ID) {
		return nil, ErrTokenBlacklisted
	}
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Get user
	user, exists := m.users[claims.UserID]
	if !exists || !user.Active {
		return nil, ErrUserNotFound
	}
	
	// Get session if available
	var session *Session
	if claims.SessionID != "" {
		if s, exists := m.sessions[claims.SessionID]; exists && s.Active {
			if time.Now().After(s.ExpiresAt) {
				return nil, ErrSessionExpired
			}
			session = s
		}
	}
	
	return &AuthContext{
		User:        user,
		Session:     session,
		Claims:      claims,
		TokenString: tokenString,
		Method:      AuthMethodJWT,
	}, nil
}

// ValidateAPIKey validates an API key and returns the auth context
func (m *Manager) ValidateAPIKey(key string) (*AuthContext, error) {
	keyHash := hashAPIKey(key)
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Find API key
	var apiKey *APIKey
	for _, ak := range m.apiKeys {
		if subtle.ConstantTimeCompare([]byte(ak.Key), []byte(keyHash)) == 1 && ak.Active {
			apiKey = ak
			break
		}
	}
	
	if apiKey == nil {
		return nil, ErrAPIKeyNotFound
	}
	
	// Check expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}
	
	// Get user
	user, exists := m.users[apiKey.UserID]
	if !exists || !user.Active {
		return nil, ErrUserNotFound
	}
	
	// Update last used
	now := time.Now()
	apiKey.LastUsedAt = &now
	
	return &AuthContext{
		User:   user,
		APIKey: apiKey,
		Method: AuthMethodAPIKey,
	}, nil
}

// CreateUser creates a new user
func (m *Manager) CreateUser(req *CreateUserRequest) (*User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if username already exists
	for _, u := range m.users {
		if u.Username == req.Username {
			return nil, fmt.Errorf("username already exists")
		}
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), m.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Set permissions based on role if not provided
	permissions := req.Permissions
	if len(permissions) == 0 {
		if rolePerms, exists := DefaultRolePermissions[req.Role]; exists {
			permissions = rolePerms
		}
	}
	
	// Create user
	user := &User{
		ID:          generateID(),
		Username:    req.Username,
		Email:       req.Email,
		Role:        req.Role,
		Permissions: permissions,
		Metadata: map[string]string{
			"password_hash": string(hashedPassword),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}
	
	// Add custom metadata
	for k, v := range req.Metadata {
		user.Metadata[k] = v
	}
	
	m.users[user.ID] = user
	
	return user, nil
}

// CreateAPIKey creates a new API key for a user
func (m *Manager) CreateAPIKey(userID string, req *CreateAPIKeyRequest) (*APIKey, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if user exists
	user, exists := m.users[userID]
	if !exists || !user.Active {
		return nil, "", ErrUserNotFound
	}
	
	// Generate API key
	rawKey := generateAPIKey()
	keyHash := hashAPIKey(rawKey)
	
	// Set permissions
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = user.Permissions
	}
	
	apiKey := &APIKey{
		ID:          generateID(),
		Name:        req.Name,
		Key:         keyHash,
		UserID:      userID,
		Permissions: permissions,
		Metadata:    req.Metadata,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   time.Now(),
		Active:      true,
	}
	
	if apiKey.Metadata == nil {
		apiKey.Metadata = make(map[string]string)
	}
	
	m.apiKeys[apiKey.ID] = apiKey
	
	return apiKey, rawKey, nil
}

// RevokeToken adds a token to the blacklist
func (m *Manager) RevokeToken(tokenID string, expiry time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.blacklistCache[tokenID] = expiry
}

// RevokeSession revokes a session
func (m *Manager) RevokeSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	session, exists := m.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}
	
	session.Active = false
	
	// Also blacklist the associated token
	if session.TokenID != "" {
		m.blacklistCache[session.TokenID] = session.ExpiresAt
	}
	
	return nil
}

// RevokeAPIKey revokes an API key
func (m *Manager) RevokeAPIKey(keyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	apiKey, exists := m.apiKeys[keyID]
	if !exists {
		return ErrAPIKeyNotFound
	}
	
	apiKey.Active = false
	
	return nil
}

// HasPermission checks if the auth context has a specific permission
func (m *Manager) HasPermission(ctx *AuthContext, permission string) bool {
	if ctx == nil || ctx.User == nil {
		return false
	}
	
	// Admin role has all permissions
	if ctx.User.Role == RoleAdmin {
		return true
	}
	
	// Check user permissions
	for _, perm := range ctx.User.Permissions {
		if perm == permission || perm == PermissionSystemAdmin {
			return true
		}
	}
	
	// Check API key permissions if using API key auth
	if ctx.Method == AuthMethodAPIKey && ctx.APIKey != nil {
		for _, perm := range ctx.APIKey.Permissions {
			if perm == permission || perm == PermissionSystemAdmin {
				return true
			}
		}
	}
	
	return false
}

// isTokenBlacklisted checks if a token is blacklisted
func (m *Manager) isTokenBlacklisted(tokenID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	expiry, exists := m.blacklistCache[tokenID]
	if !exists {
		return false
	}
	
	// Check if blacklist entry has expired
	if time.Now().After(expiry) {
		delete(m.blacklistCache, tokenID)
		return false
	}
	
	return true
}

// Background cleanup routines
func (m *Manager) cleanupExpiredSessions() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for id, session := range m.sessions {
				if now.After(session.ExpiresAt) {
					delete(m.sessions, id)
				}
			}
			m.mu.Unlock()
		case <-m.stopCleanup:
			return
		}
	}
}

func (m *Manager) cleanupBlacklist() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for tokenID, expiry := range m.blacklistCache {
				if now.After(expiry) {
					delete(m.blacklistCache, tokenID)
				}
			}
			m.mu.Unlock()
		case <-m.stopCleanup:
			return
		}
	}
}

// Utility functions
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateAPIKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "ok_" + hex.EncodeToString(bytes)
}

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}