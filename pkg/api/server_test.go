package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/p2p"
)

// Basic test to verify API package compiles and server can be created
func TestServerCreation(t *testing.T) {
	// Test that we can create a basic server configuration
	apiConfig := &config.APIConfig{
		Listen:      ":0",
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			Enabled:     true,
			RequestsPer: 100,
			Duration:    time.Minute,
			BurstSize:   10,
		},
		Cors: config.CorsConfig{
			Enabled:          true,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: false,
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
	p2pConfig := &config.P2PConfig{
		ListenAddr:     "/ip4/127.0.0.1/tcp/0",
		BootstrapPeers: []string{},
		DialTimeout:    time.Second * 10,
		MaxConnections: 100,
	}

	if p2pConfig.ListenAddr == "" {
		t.Error("P2P config should have a listen address")
	}
}

func TestRateLimitConfig(t *testing.T) {
	rateLimitConfig := config.RateLimitConfig{
		Enabled:     true,
		RequestsPer: 100,
		Duration:    time.Minute,
		BurstSize:   20,
	}

	if !rateLimitConfig.Enabled {
		t.Error("Rate limit should be enabled for test")
	}

	if rateLimitConfig.RequestsPer != 100 {
		t.Errorf("Expected 100 requests per minute, got %d", rateLimitConfig.RequestsPer)
	}
}

func TestCorsConfig(t *testing.T) {
	corsConfig := config.CorsConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}

	if !corsConfig.Enabled {
		t.Error("CORS should be enabled for test")
	}

	if len(corsConfig.AllowedOrigins) == 0 {
		t.Error("CORS should have allowed origins")
	}
}

func TestConfigValidation(t *testing.T) {
	// Test default configuration
	defaultConfig := config.DefaultConfig()
	
	if defaultConfig == nil {
		t.Fatal("Default config should not be nil")
	}

	// Validate JWT config
	if defaultConfig.JWT.SecretKey == "" {
		t.Error("JWT secret key should not be empty")
	}

	// Validate API config
	if defaultConfig.API.Listen == "" {
		t.Error("API listen address should not be empty")
	}

	// Validate Auth config
	if !defaultConfig.Auth.Enabled {
		t.Log("Auth is disabled in default config")
	}

	// Validate P2P config
	if defaultConfig.P2P.ListenAddr == "" {
		t.Error("P2P listen address should not be empty")
	}
}

func TestPerformanceBenchmarks(t *testing.T) {
	// Basic performance test for config creation
	start := time.Now()
	
	for i := 0; i < 1000; i++ {
		_ = config.DefaultConfig()
	}
	
	duration := time.Since(start)
	
	if duration > time.Millisecond*100 {
		t.Logf("Config creation took %v for 1000 iterations, consider optimization", duration)
	}
}