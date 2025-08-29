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

	// Get the actual generated admin password from the manager
	manager.mu.RLock()
	adminUser, exists := manager.users["admin"]
	manager.mu.RUnlock()
	require.True(t, exists, "Admin user should exist")

	// Create a test user with known password for testing
	testPassword := "testpassword123"
	hashedPassword, err := manager.hashPassword(testPassword)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.users["testuser"] = &User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Permissions:  []Permission{PermissionRead},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	manager.mu.Unlock()

	// Test authentication with test user
	authCtx, err := manager.Authenticate("testuser", testPassword, map[string]string{
		"ip_address": "127.0.0.1",
		"user_agent": "test-agent",
	})

	require.NoError(t, err)
	require.NotNil(t, authCtx)
	assert.Equal(t, "testuser", authCtx.User.Username)
	assert.Equal(t, RoleUser, authCtx.User.Role)
	assert.Equal(t, AuthMethodJWT, authCtx.Method)
	assert.NotEmpty(t, authCtx.TokenString)

	// Test authentication with wrong password
	_, err = manager.Authenticate("testuser", "wrongpassword", map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_CREDENTIALS")

	// Test authentication with non-existent user
	_, err = manager.Authenticate("nonexistent", "password", map[string]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_CREDENTIALS")

	// Verify admin user exists but use a proper test for admin login
	assert.NotNil(t, adminUser)
	assert.Equal(t, RoleAdmin, adminUser.Role)
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

	// Create a test user with known password
	testPassword := "testpassword123"
	hashedPassword, err := manager.hashPassword(testPassword)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.users["testuser"] = &User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Permissions:  []Permission{PermissionRead},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	manager.mu.Unlock()

	// Authenticate to get a token
	authCtx, err := manager.Authenticate("testuser", testPassword, map[string]string{})
	require.NoError(t, err)

	// Validate the token
	validatedCtx, err := manager.ValidateToken(authCtx.TokenString)
	require.NoError(t, err)
	require.NotNil(t, validatedCtx)
	assert.Equal(t, "testuser", validatedCtx.User.Username)
	assert.Equal(t, RoleUser, validatedCtx.User.Role)

	// Test invalid token
	_, err = manager.ValidateToken("invalid.token.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_TOKEN")

	// Test empty token
	_, err = manager.ValidateToken("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_TOKEN")
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

	// Test creating a new user
	user, err := manager.CreateUser("testuser", "password123", RoleUser, []Permission{PermissionRead})
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, RoleUser, user.Role)
	assert.True(t, user.IsActive)

	// Test creating duplicate user
	_, err = manager.CreateUser("testuser", "password123", RoleUser, []Permission{PermissionRead})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "USER_ALREADY_EXISTS")

	// Test creating user with invalid username
	_, err = manager.CreateUser("", "password123", RoleUser, []Permission{PermissionRead})
	assert.Error(t, err)
}

func TestUserManagement(t *testing.T) {
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

	// Create a test user
	user, err := manager.CreateUser("testuser", "password123", RoleUser, []Permission{PermissionRead})
	require.NoError(t, err)

	// Test getting user
	retrievedUser, err := manager.GetUser("testuser")
	require.NoError(t, err)
	assert.Equal(t, user.Username, retrievedUser.Username)

	// Test updating user
	err = manager.UpdateUser("testuser", &UserUpdate{
		Role:        &RoleAdmin,
		Permissions: &[]Permission{PermissionRead, PermissionWrite},
	})
	require.NoError(t, err)

	updatedUser, err := manager.GetUser("testuser")
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, updatedUser.Role)
	assert.Contains(t, updatedUser.Permissions, PermissionWrite)

	// Test deactivating user
	err = manager.DeactivateUser("testuser")
	require.NoError(t, err)

	deactivatedUser, err := manager.GetUser("testuser")
	require.NoError(t, err)
	assert.False(t, deactivatedUser.IsActive)

	// Test deleting user
	err = manager.DeleteUser("testuser")
	require.NoError(t, err)

	_, err = manager.GetUser("testuser")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "USER_NOT_FOUND")
}

func TestAPIKeyManagement(t *testing.T) {
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

	// Create a test user first
	user, err := manager.CreateUser("testuser", "password123", RoleUser, []Permission{PermissionRead})
	require.NoError(t, err)

	// Test creating API key
	apiKey, err := manager.CreateAPIKey("testuser", "test-key", []Permission{PermissionRead}, time.Hour)
	require.NoError(t, err)
	require.NotNil(t, apiKey)
	assert.Equal(t, "test-key", apiKey.Name)
	assert.Equal(t, "testuser", apiKey.Username)
	assert.True(t, apiKey.IsActive)

	// Test validating API key
	authCtx, err := manager.ValidateAPIKey(apiKey.Key)
	require.NoError(t, err)
	require.NotNil(t, authCtx)
	assert.Equal(t, "testuser", authCtx.User.Username)
	assert.Equal(t, AuthMethodAPIKey, authCtx.Method)

	// Test revoking API key
	err = manager.RevokeAPIKey(apiKey.Key)
	require.NoError(t, err)

	_, err = manager.ValidateAPIKey(apiKey.Key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INVALID_API_KEY")
}

func TestRateLimiting(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:        true,
		Method:         "jwt",
		TokenExpiry:    24 * time.Hour,
		SecretKey:      "test-secret-key",
		Issuer:         "ollama-test",
		Audience:       "ollama-api",
		RateLimitRPM:   5,    // 5 requests per minute
		RateLimitBurst: 2,    // burst of 2
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Create a test user
	testPassword := "testpassword123"
	hashedPassword, err := manager.hashPassword(testPassword)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.users["testuser"] = &User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Permissions:  []Permission{PermissionRead},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	manager.mu.Unlock()

	clientIP := "127.0.0.1"
	metadata := map[string]string{"ip_address": clientIP}

	// First few requests should succeed (within burst limit)
	for i := 0; i < 2; i++ {
		_, err := manager.Authenticate("testuser", testPassword, metadata)
		require.NoError(t, err, "Request %d should succeed", i+1)
	}

	// Additional request should be rate limited
	_, err = manager.Authenticate("testuser", testPassword, metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_EXCEEDED")
}

func TestTokenExpiry(t *testing.T) {
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: time.Millisecond, // Very short expiry for testing
		SecretKey:   "test-secret-key",
		Issuer:      "ollama-test",
		Audience:    "ollama-api",
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	defer manager.Close()

	// Create a test user
	testPassword := "testpassword123"
	hashedPassword, err := manager.hashPassword(testPassword)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.users["testuser"] = &User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Permissions:  []Permission{PermissionRead},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	manager.mu.Unlock()

	// Authenticate to get a token
	authCtx, err := manager.Authenticate("testuser", testPassword, map[string]string{})
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should now be invalid
	_, err = manager.ValidateToken(authCtx.TokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TOKEN_EXPIRED")
}

func TestBlacklist(t *testing.T) {
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

	// Create a test user
	testPassword := "testpassword123"
	hashedPassword, err := manager.hashPassword(testPassword)
	require.NoError(t, err)

	manager.mu.Lock()
	manager.users["testuser"] = &User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
		Role:         RoleUser,
		Permissions:  []Permission{PermissionRead},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	manager.mu.Unlock()

	// Authenticate to get a token
	authCtx, err := manager.Authenticate("testuser", testPassword, map[string]string{})
	require.NoError(t, err)

	// Token should be valid initially
	_, err = manager.ValidateToken(authCtx.TokenString)
	require.NoError(t, err)

	// Blacklist the token
	err = manager.BlacklistToken(authCtx.TokenString)
	require.NoError(t, err)

	// Token should now be invalid
	_, err = manager.ValidateToken(authCtx.TokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TOKEN_BLACKLISTED")
}

func TestPermissions(t *testing.T) {
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

	// Create users with different permissions
	readOnlyUser, err := manager.CreateUser("readonly", "password123", RoleUser, []Permission{PermissionRead})
	require.NoError(t, err)

	adminUser, err := manager.CreateUser("admin", "password123", RoleAdmin, []Permission{PermissionRead, PermissionWrite, PermissionAdmin})
	require.NoError(t, err)

	// Test permission checking
	assert.True(t, readOnlyUser.HasPermission(PermissionRead))
	assert.False(t, readOnlyUser.HasPermission(PermissionWrite))
	assert.False(t, readOnlyUser.HasPermission(PermissionAdmin))

	assert.True(t, adminUser.HasPermission(PermissionRead))
	assert.True(t, adminUser.HasPermission(PermissionWrite))
	assert.True(t, adminUser.HasPermission(PermissionAdmin))
}