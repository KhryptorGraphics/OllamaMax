package auth

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	
	defer manager.Close()
	
	// Should have created default admin user
	assert.Equal(t, 1, len(manager.users))
}

func TestAuthenticate(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Test authentication with default admin user
	authCtx, err := manager.Authenticate("admin", "admin123", map[string]string{
		"ip_address": "127.0.0.1",
		"user_agent": "test-agent",
	})
	
	require.NoError(t, err)
	require.NotNil(t, authCtx)
	assert.Equal(t, "admin", authCtx.User.Username)
	assert.Equal(t, RoleAdmin, authCtx.User.Role)
	assert.Equal(t, AuthMethodJWT, authCtx.Method)
	assert.NotEmpty(t, authCtx.TokenString)
}

func TestValidateToken(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Authenticate to get a token
	authCtx, err := manager.Authenticate("admin", "admin123", map[string]string{})
	require.NoError(t, err)
	
	// Validate the token
	validatedCtx, err := manager.ValidateToken(authCtx.TokenString)
	require.NoError(t, err)
	require.NotNil(t, validatedCtx)
	assert.Equal(t, authCtx.User.ID, validatedCtx.User.ID)
	assert.Equal(t, authCtx.User.Username, validatedCtx.User.Username)
}

func TestCreateUser(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Create a new user
	req := &CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword",
		Role:     RoleUser,
	}
	
	user, err := manager.CreateUser(req)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, RoleUser, user.Role)
	assert.True(t, user.Active)
	
	// Should be able to authenticate with new user
	authCtx, err := manager.Authenticate("testuser", "testpassword", map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, user.ID, authCtx.User.ID)
}

func TestCreateAPIKey(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Create a user first
	req := &CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword",
		Role:     RoleUser,
	}
	
	user, err := manager.CreateUser(req)
	require.NoError(t, err)
	
	// Create API key
	apiKeyReq := &CreateAPIKeyRequest{
		Name:        "Test API Key",
		Permissions: []string{PermissionModelRead, PermissionInferenceWrite},
	}
	
	apiKey, rawKey, err := manager.CreateAPIKey(user.ID, apiKeyReq)
	require.NoError(t, err)
	require.NotNil(t, apiKey)
	assert.NotEmpty(t, rawKey)
	assert.Equal(t, "Test API Key", apiKey.Name)
	assert.Equal(t, user.ID, apiKey.UserID)
	assert.True(t, apiKey.Active)
	
	// Should be able to validate API key
	authCtx, err := manager.ValidateAPIKey(rawKey)
	require.NoError(t, err)
	assert.Equal(t, user.ID, authCtx.User.ID)
	assert.Equal(t, AuthMethodAPIKey, authCtx.Method)
}

func TestHasPermission(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Create user with specific permissions
	req := &CreateUserRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "testpassword",
		Role:        RoleUser,
		Permissions: []string{PermissionModelRead, PermissionInferenceWrite},
	}
	
	user, err := manager.CreateUser(req)
	require.NoError(t, err)
	
	authCtx := &AuthContext{
		User:   user,
		Method: AuthMethodJWT,
	}
	
	// Test permissions
	assert.True(t, manager.HasPermission(authCtx, PermissionModelRead))
	assert.True(t, manager.HasPermission(authCtx, PermissionInferenceWrite))
	assert.False(t, manager.HasPermission(authCtx, PermissionNodeAdmin))
	
	// Admin should have all permissions
	adminAuthCtx, err := manager.Authenticate("admin", "admin123", map[string]string{})
	require.NoError(t, err)
	assert.True(t, manager.HasPermission(adminAuthCtx, PermissionNodeAdmin))
	assert.True(t, manager.HasPermission(adminAuthCtx, PermissionSystemAdmin))
}

func TestJWTManager(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	jwtManager, err := NewJWTManager(cfg)
	require.NoError(t, err)
	require.NotNil(t, jwtManager)
	
	// Create test user
	user := &User{
		ID:          "test-user-id",
		Username:    "testuser",
		Email:       "test@example.com",
		Role:        RoleUser,
		Permissions: []string{PermissionModelRead},
	}
	
	// Generate token pair
	tokenPair, err := jwtManager.GenerateTokenPair(user, "session-id", map[string]string{})
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	
	// Validate access token
	claims, err := jwtManager.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Role, claims.Role)
}

func TestServiceToken(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	jwtManager, err := NewJWTManager(cfg)
	require.NoError(t, err)
	
	// Generate service token
	serviceToken, err := jwtManager.GenerateServiceToken(
		"service-1",
		"Test Service",
		[]string{PermissionNodeRead, PermissionModelRead},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceToken)
	
	// Validate service token
	claims, err := jwtManager.ValidateServiceToken(serviceToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, "service-1", claims.UserID)
	assert.Equal(t, "Test Service", claims.Username)
	assert.Equal(t, RoleService, claims.Role)
	assert.Equal(t, "service", claims.Metadata["token_type"])
}

func TestTokenBlacklist(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}
	
	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()
	
	// Authenticate to get a token
	authCtx, err := manager.Authenticate("admin", "admin123", map[string]string{})
	require.NoError(t, err)
	
	// Token should be valid initially
	_, err = manager.ValidateToken(authCtx.TokenString)
	require.NoError(t, err)
	
	// Blacklist the token
	manager.RevokeToken(authCtx.Claims.ID, authCtx.Claims.ExpiresAt.Time)
	
	// Token should now be invalid
	_, err = manager.ValidateToken(authCtx.TokenString)
	require.Error(t, err)
	assert.Equal(t, ErrTokenBlacklisted, err)
}

func TestRolePermissions(t *testing.T) {
	// Test default role permissions
	adminPerms := DefaultRolePermissions[RoleAdmin]
	assert.Contains(t, adminPerms, PermissionSystemAdmin)
	assert.Contains(t, adminPerms, PermissionUserAdmin)
	
	userPerms := DefaultRolePermissions[RoleUser]
	assert.Contains(t, userPerms, PermissionModelRead)
	assert.Contains(t, userPerms, PermissionInferenceWrite)
	assert.NotContains(t, userPerms, PermissionSystemAdmin)
	
	readOnlyPerms := DefaultRolePermissions[RoleReadOnly]
	assert.Contains(t, readOnlyPerms, PermissionModelRead)
	assert.NotContains(t, readOnlyPerms, PermissionModelWrite)
}