package auth

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.AuthConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid config",
			config: &config.AuthConfig{
				JWTSecret:   "test-issuer",
				TokenExpiry: 1 * time.Hour,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewJWTService(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.NotNil(t, service.privateKey)
				assert.NotNil(t, service.publicKey)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, err := NewJWTService(nil)
	require.NoError(t, err)
	require.NotNil(t, service)

	tests := []struct {
		name        string
		userID      string
		username    string
		role        string
		permissions []string
		expectError bool
	}{
		{
			name:        "valid token generation",
			userID:      "user123",
			username:    "testuser",
			role:        RoleUser,
			permissions: []string{PermissionModelRead},
			expectError: false,
		},
		{
			name:        "admin token generation",
			userID:      "admin123",
			username:    "admin",
			role:        RoleAdmin,
			permissions: GetRolePermissions(RoleAdmin),
			expectError: false,
		},
		{
			name:        "empty user data",
			userID:      "",
			username:    "",
			role:        "",
			permissions: []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenPair, err := service.GenerateToken(tt.userID, tt.username, tt.role, tt.permissions)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tokenPair)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tokenPair)
				assert.NotEmpty(t, tokenPair.AccessToken)
				assert.NotEmpty(t, tokenPair.RefreshToken)
				assert.Equal(t, "Bearer", tokenPair.TokenType)
				assert.True(t, tokenPair.ExpiresAt.After(time.Now()))
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	service, err := NewJWTService(nil)
	require.NoError(t, err)

	// Generate a test token
	tokenPair, err := service.GenerateToken("test123", "testuser", RoleUser, []string{PermissionModelRead})
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
		checkClaims func(t *testing.T, claims *Claims)
	}{
		{
			name:        "valid token",
			token:       tokenPair.AccessToken,
			expectError: false,
			checkClaims: func(t *testing.T, claims *Claims) {
				assert.Equal(t, "test123", claims.UserID)
				assert.Equal(t, "testuser", claims.Username)
				assert.Equal(t, RoleUser, claims.Role)
				assert.Contains(t, claims.Permissions, PermissionModelRead)
			},
		},
		{
			name:        "invalid token",
			token:       "invalid.token.here",
			expectError: true,
			checkClaims: nil,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
			checkClaims: nil,
		},
		{
			name:        "malformed token",
			token:       "not.a.jwt",
			expectError: true,
			checkClaims: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				if tt.checkClaims != nil {
					tt.checkClaims(t, claims)
				}
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	service, err := NewJWTService(nil)
	require.NoError(t, err)

	// Generate initial tokens
	tokenPair, err := service.GenerateToken("test123", "testuser", RoleUser, []string{PermissionModelRead})
	require.NoError(t, err)

	tests := []struct {
		name         string
		refreshToken string
		expectError  bool
	}{
		{
			name:         "valid refresh token",
			refreshToken: tokenPair.RefreshToken,
			expectError:  false,
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid.token",
			expectError:  true,
		},
		{
			name:         "access token instead of refresh",
			refreshToken: tokenPair.AccessToken,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newTokenPair, err := service.RefreshToken(tt.refreshToken)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, newTokenPair)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newTokenPair)
				assert.NotEmpty(t, newTokenPair.AccessToken)
				assert.NotEmpty(t, newTokenPair.RefreshToken)
				
				// Verify new token is different
				assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
			}
		})
	}
}

func TestClaimsPermissions(t *testing.T) {
	claims := &Claims{
		Role:        RoleAdmin,
		Permissions: GetRolePermissions(RoleAdmin),
		Metadata:    make(map[string]string),
	}

	// Test permission checking
	assert.True(t, claims.HasPermission(PermissionModelManage))
	assert.True(t, claims.HasPermission(PermissionSystemManage))
	assert.False(t, claims.HasPermission("non-existent-permission"))

	// Test role checking
	assert.True(t, claims.IsAdmin())
	assert.True(t, claims.IsOperator())

	// Test metadata operations
	claims.SetMetadata("test-key", "test-value")
	value, exists := claims.GetMetadata("test-key")
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)

	_, exists = claims.GetMetadata("non-existent-key")
	assert.False(t, exists)
}

func TestGetRolePermissions(t *testing.T) {
	tests := []struct {
		role        string
		expectedLen int
		shouldHave  []string
	}{
		{
			role:        RoleAdmin,
			expectedLen: 9,
			shouldHave:  []string{PermissionModelManage, PermissionSystemManage},
		},
		{
			role:        RoleOperator,
			expectedLen: 5,
			shouldHave:  []string{PermissionModelRead, PermissionInferenceRun},
		},
		{
			role:        RoleUser,
			expectedLen: 2,
			shouldHave:  []string{PermissionModelRead, PermissionInferenceRun},
		},
		{
			role:        RoleReadonly,
			expectedLen: 4,
			shouldHave:  []string{PermissionModelRead, PermissionMetricsRead},
		},
		{
			role:        "unknown-role",
			expectedLen: 0,
			shouldHave:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			permissions := GetRolePermissions(tt.role)
			assert.Len(t, permissions, tt.expectedLen)
			
			for _, expectedPerm := range tt.shouldHave {
				assert.Contains(t, permissions, expectedPerm)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// Create service with short expiration
	config := &config.AuthConfig{
		TokenExpiry: 1 * time.Millisecond,
	}
	service, err := NewJWTService(config)
	require.NoError(t, err)

	// Generate token
	tokenPair, err := service.GenerateToken("test", "test", RoleUser, []string{})
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	claims, err := service.ValidateToken(tokenPair.AccessToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestRevokeToken(t *testing.T) {
	service, err := NewJWTService(nil)
	require.NoError(t, err)

	tokenPair, err := service.GenerateToken("test", "test", RoleUser, []string{})
	require.NoError(t, err)

	// Test revoke valid token
	err = service.RevokeToken(tokenPair.AccessToken)
	assert.NoError(t, err)

	// Test revoke invalid token
	err = service.RevokeToken("invalid.token")
	assert.Error(t, err)
}

func TestPublicKeyAccess(t *testing.T) {
	service, err := NewJWTService(nil)
	require.NoError(t, err)

	publicKey := service.GetPublicKey()
	assert.NotNil(t, publicKey)
	assert.Equal(t, service.publicKey, publicKey)
}

func BenchmarkGenerateToken(b *testing.B) {
	service, err := NewJWTService(nil)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateToken("user123", "testuser", RoleUser, GetRolePermissions(RoleUser))
		require.NoError(b, err)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	service, err := NewJWTService(nil)
	require.NoError(b, err)

	tokenPair, err := service.GenerateToken("user123", "testuser", RoleUser, GetRolePermissions(RoleUser))
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateToken(tokenPair.AccessToken)
		require.NoError(b, err)
	}
}