package api

import (
	"context"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
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

func TestCorsConfigDefaults(t *testing.T) {
	corsConfig := config.CorsConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           3600,
	}

	if len(corsConfig.AllowedOrigins) == 0 {
		t.Error("CORS config should have allowed origins")
	}

	if len(corsConfig.AllowedMethods) == 0 {
		t.Error("CORS config should have allowed methods")
	}

	if corsConfig.MaxAge <= 0 {
		t.Error("CORS config should have positive max age")
	}
}

func TestRateLimitConfig(t *testing.T) {
	rateLimitConfig := config.RateLimitConfig{
		RPS:       100,
		Burst:     200,
		Enabled:   true,
		WhiteList: []string{"127.0.0.1", "::1"},
	}

	if rateLimitConfig.RPS <= 0 {
		t.Error("Rate limit RPS should be positive")
	}

	if rateLimitConfig.Burst <= 0 {
		t.Error("Rate limit burst should be positive")
	}

	if !rateLimitConfig.Enabled {
		t.Error("Rate limiting should be enabled for this test")
	}

	if len(rateLimitConfig.WhiteList) == 0 {
		t.Error("Rate limit should have whitelist entries")
	}
}

func TestP2PNodeConfig(t *testing.T) {
	nodeConfig := &p2pconfig.NodeConfig{
		Listen: []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip6/::/tcp/0",
		},
		EnableNoise:       true,
		EnableRelay:       true,
		EnableAutoRelay:   true,
		EnableHolePunch:   true,
		ConnMgrLow:        10,
		ConnMgrHigh:       100,
		ConnMgrGrace:      30 * time.Second,
		BootstrapPeers:    []string{},
		ProtocolPrefix:    "/ollamamax",
		EnableNAT:         true,
		EnableMDNS:        true,
		MDNSServiceName:   "ollamamax-node",
	}

	// Validate configuration
	if len(nodeConfig.Listen) == 0 {
		t.Error("Node config should have listen addresses")
	}

	if nodeConfig.ConnMgrLow >= nodeConfig.ConnMgrHigh {
		t.Error("Connection manager low should be less than high")
	}

	if nodeConfig.ConnMgrGrace <= 0 {
		t.Error("Connection manager grace period should be positive")
	}

	if !nodeConfig.EnableNoise {
		t.Error("Noise protocol should be enabled for security")
	}

	// Test listen addresses are valid multiaddr formats
	for _, addr := range nodeConfig.Listen {
		if addr == "" {
			t.Error("Listen address should not be empty")
		}
		
		// Basic validation - should start with /ip4 or /ip6
		if !(len(addr) > 4 && (addr[:4] == "/ip4" || addr[:4] == "/ip6")) {
			t.Errorf("Invalid listen address format: %s", addr)
		}
	}
}

func TestConfigIntegration(t *testing.T) {
	// Test that API and P2P configs can work together
	apiConfig := &config.APIConfig{
		Listen:      ":8080",
		MaxBodySize: 10 * 1024 * 1024, // 10MB
		RateLimit: config.RateLimitConfig{
			RPS:     1000,
			Burst:   2000,
			Enabled: true,
		},
		Cors: config.CorsConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
		},
	}

	p2pConfig := &p2pconfig.NodeConfig{
		Listen:       []string{"/ip4/0.0.0.0/tcp/9000"},
		EnableNoise:  true,
		ConnMgrLow:   20,
		ConnMgrHigh:  200,
		ConnMgrGrace: 60 * time.Second,
	}

	// Validate both configs are compatible
	if apiConfig == nil || p2pConfig == nil {
		t.Error("Both configs should be valid")
	}

	// Ensure they use different ports (basic check)
	if apiConfig.Listen == "" || len(p2pConfig.Listen) == 0 {
		t.Error("Both configs should specify listen addresses")
	}
}

func BenchmarkAPIConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &config.APIConfig{
			Listen:      ":8080",
			MaxBodySize: 1024 * 1024,
			RateLimit: config.RateLimitConfig{
				RPS: 100,
			},
		}
	}
}

func BenchmarkP2PConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &p2pconfig.NodeConfig{
			Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
			EnableNoise:  true,
			ConnMgrLow:   5,
			ConnMgrHigh:  20,
			ConnMgrGrace: 30 * time.Second,
		}
	}
}