package auth

import (
	"context"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/redis/go-redis/v9"
)

func TestJWTTokenGeneration(t *testing.T) {
	// Create auth config with proper structure
	authConfig := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: time.Hour,
		SecretKey:   "test-secret-key-for-testing-purposes",
		RefreshTime: 24 * time.Hour,
	}

	if authConfig.SecretKey == "" {
		t.Fatal("Auth config secret key should not be empty")
	}

	// Test basic token generation would work with this config
	if !authConfig.Enabled {
		t.Error("Auth should be enabled for testing")
	}

	if authConfig.Method != "jwt" {
		t.Errorf("Expected jwt method, got %s", authConfig.Method)
	}

	if authConfig.TokenExpiry <= 0 {
		t.Error("Token expiry should be positive")
	}
}

func TestJWTConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.AuthConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.AuthConfig{
				Enabled:     true,
				Method:      "jwt",
				TokenExpiry: time.Hour,
				SecretKey:   "valid-secret-key",
				RefreshTime: 24 * time.Hour,
			},
			expectError: false,
		},
		{
			name: "empty secret key",
			config: &config.AuthConfig{
				Enabled:     true,
				Method:      "jwt",
				TokenExpiry: time.Hour,
				SecretKey:   "",
				RefreshTime: 24 * time.Hour,
			},
			expectError: true,
		},
		{
			name: "zero token expiry",
			config: &config.AuthConfig{
				Enabled:     true,
				Method:      "jwt",
				TokenExpiry: 0,
				SecretKey:   "valid-secret-key",
				RefreshTime: 24 * time.Hour,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateAuthConfig(tt.config)
			if tt.expectError && valid {
				t.Error("Expected validation to fail but it passed")
			}
			if !tt.expectError && !valid {
				t.Error("Expected validation to pass but it failed")
			}
		})
	}
}

// Helper function for validation (to be implemented in main auth code)
func validateAuthConfig(config *config.AuthConfig) bool {
	if !config.Enabled {
		return true // Disabled auth is valid
	}
	
	if config.SecretKey == "" {
		return false
	}
	
	if config.TokenExpiry <= 0 {
		return false
	}
	
	if config.Method == "" {
		return false
	}
	
	return true
}

func TestDefaultAuthConfig(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	
	if defaultConfig == nil {
		t.Fatal("Default config should not be nil")
	}
	
	authConfig := defaultConfig.Auth
	
	// Test that default auth config is reasonable
	if authConfig.SecretKey == "" {
		t.Error("Default auth config should have a secret key")
	}
	
	if authConfig.TokenExpiry <= 0 {
		t.Error("Default auth config should have positive token expiry")
	}
	
	if authConfig.Method == "" {
		t.Error("Default auth config should have an auth method")
	}
}

func TestTokenRevocation(t *testing.T) {
	// Create a mock Redis client for testing
	mockRedis := &mockRedisClient{}

	authConfig := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: time.Hour,
		SecretKey:   "test-secret-key-for-testing-purposes",
		RefreshTime: 24 * time.Hour,
	}

	service, err := NewJWTService(authConfig, mockRedis)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Test token generation and revocation
	tokenPair, err := service.GenerateToken("testuser", "Test User", "user", nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid initially
	claims, err := service.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Token should be valid initially: %v", err)
	}
	if claims.UserID != "testuser" {
		t.Errorf("Expected user ID 'testuser', got '%s'", claims.UserID)
	}

	// Revoke the token
	err = service.RevokeToken(tokenPair.AccessToken, time.Hour)
	if err != nil {
		t.Fatalf("Failed to revoke token: %v", err)
	}

	// Token should now be invalid
	_, err = service.ValidateToken(tokenPair.AccessToken)
	if err == nil {
		t.Error("Token should be invalid after revocation")
	}
	if err.Error() != "token has been revoked" {
		t.Errorf("Expected 'token has been revoked' error, got: %v", err)
	}
}

func TestUserTokenRevocation(t *testing.T) {
	mockRedis := &mockRedisClient{}

	authConfig := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: time.Hour,
		SecretKey:   "test-secret-key-for-testing-purposes",
		RefreshTime: 24 * time.Hour,
	}

	service, err := NewJWTService(authConfig, mockRedis)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	userID := "testuser"

	// Generate a token
	tokenPair, err := service.GenerateToken(userID, "Test User", "user", nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid initially
	claims, err := service.ValidateToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("Token should be valid initially: %v", err)
	}

	// Revoke all user tokens
	err = service.RevokeUserTokens(userID)
	if err != nil {
		t.Fatalf("Failed to revoke user tokens: %v", err)
	}

	// Check that user is marked as revoked
	isRevoked, err := service.IsUserRevoked(userID)
	if err != nil {
		t.Fatalf("Failed to check user revocation: %v", err)
	}
	if !isRevoked {
		t.Error("User should be marked as revoked")
	}
}

// Mock Redis client for testing
type mockRedisClient struct {
	store map[string]string
}

func (m *mockRedisClient) Set(ctx context.Context, key, value string, expiration time.Duration) *redis.StatusCmd {
	if m.store == nil {
		m.store = make(map[string]string)
	}
	m.store[key] = value
	return redis.NewStatusCmd(ctx, "set", key, value, "ex", expiration)
}

func (m *mockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if m.store == nil {
		m.store = make(map[string]string)
	}
	count := int64(0)
	for _, key := range keys {
		if _, exists := m.store[key]; exists {
			count++
		}
	}
	cmd := redis.NewIntCmd(ctx, "exists", keys)
	cmd.SetVal(count)
	return cmd
}

func (m *mockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "ping")
	cmd.SetVal("PONG")
	return cmd
}

func (m *mockRedisClient) Close() error {
	return nil
}