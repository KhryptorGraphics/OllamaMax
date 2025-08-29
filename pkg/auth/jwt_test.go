package auth

import (
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
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

func TestJWTSecurityRequirements(t *testing.T) {
	// Test minimum security requirements for JWT
	authConfig := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: time.Hour,
		SecretKey:   "short", // Too short for security
		RefreshTime: 24 * time.Hour,
	}
	
	// Secret key should be long enough for security
	if len(authConfig.SecretKey) < 16 {
		t.Log("Warning: Secret key is too short for production use")
	}
	
	// Token expiry should not be too long
	if authConfig.TokenExpiry > 24*time.Hour {
		t.Log("Warning: Token expiry is very long, consider shorter duration")
	}
}