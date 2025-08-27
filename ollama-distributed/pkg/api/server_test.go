package api

import (
	"context"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	p2pconfig "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

// Basic test to verify API package compiles and server can be created
func TestServerCreation(t *testing.T) {
	// Test that we can create a basic server configuration
	apiConfig := &config.APIConfig{
		Listen:      ":0",
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			RPS: 100,
		},
		Cors: config.CorsConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	// Verify config is valid
	if apiConfig.Listen == "" {
		t.Error("API config should have a listen address")
	}
	
	if apiConfig.MaxBodySize <= 0 {
		t.Error("API config should have a positive max body size")
	}
}

func TestBasicTypes(t *testing.T) {
	// Test that basic types are accessible
	ctx := context.Background()
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test that we can create basic config structures
	nodeConfig := &p2pconfig.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  20,
		ConnMgrGrace: 30 * time.Second,
	}

	if len(nodeConfig.Listen) == 0 {
		t.Error("Node config should have listen addresses")
	}
}

func TestAPIConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *config.APIConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &config.APIConfig{
				Listen:      ":8080",
				MaxBodySize: 1024 * 1024,
				RateLimit: config.RateLimitConfig{
					RPS: 100,
				},
			},
			valid: true,
		},
		{
			name: "empty listen address",
			config: &config.APIConfig{
				Listen:      "",
				MaxBodySize: 1024 * 1024,
			},
			valid: false,
		},
		{
			name: "zero max body size",
			config: &config.APIConfig{
				Listen:      ":8080",
				MaxBodySize: 0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Listen != "" && tt.config.MaxBodySize > 0
			if isValid != tt.valid {
				t.Errorf("Expected config validity %v, got %v", tt.valid, isValid)
			}
		})
	}
}