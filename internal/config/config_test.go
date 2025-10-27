package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test API config defaults
	if config.API.Listen == "" {
		t.Error("API.Listen should not be empty")
	}

	if config.API.MaxBodySize <= 0 {
		t.Error("API.MaxBodySize should be positive")
	}

	// Test JWT config defaults
	if config.JWT.SecretKey == "" {
		t.Error("JWT.SecretKey should not be empty")
	}

	if config.JWT.ExpiryTime == 0 {
		t.Error("JWT.ExpiryTime should not be zero")
	}

	// Test P2P config defaults
	if config.P2P.ListenAddr == "" {
		t.Error("P2P.ListenAddr should not be empty")
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() string
		cleanup   func(string)
		wantError bool
	}{
		{
			name: "valid config file",
			setup: func() string {
				content := `
api:
  listen: ":9090"
  max_body_size: 2097152
jwt:
  secret_key: "test-secret"
  expiration_time: 3600
p2p:
  listen_addr: "/ip4/0.0.0.0/tcp/4001"
  dial_timeout: 10s
  max_connections: 50
`
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := tmpfile.Write([]byte(content)); err != nil {
					t.Fatal(err)
				}
				tmpfile.Close()
				return tmpfile.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			wantError: false,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return "/non/existent/path/config.yaml"
			},
			cleanup:   func(path string) {},
			wantError: true,
		},
		{
			name: "invalid yaml",
			setup: func() string {
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				if err != nil {
					t.Fatal(err)
				}
				// Write truly malformed YAML with excessive colons that triggers validation
				tmpfile.WriteString("invalid: yaml: content: extra:")
				tmpfile.Close()
				return tmpfile.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setup()
			defer tt.cleanup(configPath)

			_, err := LoadConfig(configPath)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestEnvironmentOverrides(t *testing.T) {
	// Save original env vars
	origJWTSecret := os.Getenv("JWT_SECRET_KEY")
	origAPIListen := os.Getenv("API_LISTEN")
	defer func() {
		os.Setenv("JWT_SECRET_KEY", origJWTSecret)
		os.Setenv("API_LISTEN", origAPIListen)
	}()

	// Set environment variables
	os.Setenv("JWT_SECRET_KEY", "env-secret-key")
	os.Setenv("API_LISTEN", ":8888")

	config := DefaultConfig()

	// Environment variables should override defaults
	if config.JWT.SecretKey != "env-secret-key" {
		t.Errorf("JWT.SecretKey = %v, want %v", config.JWT.SecretKey, "env-secret-key")
	}

	if config.API.Listen != ":8888" {
		t.Errorf("API.Listen = %v, want %v", config.API.Listen, ":8888")
	}
}

func TestRateLimitConfig(t *testing.T) {
	config := DefaultConfig()

	if config.API.RateLimit.Enabled {
		if config.API.RateLimit.RequestsPer <= 0 {
			t.Error("RateLimit.RequestsPer should be positive when enabled")
		}

		if config.API.RateLimit.Duration <= 0 {
			t.Error("RateLimit.Duration should be positive when enabled")
		}

		if config.API.RateLimit.BurstSize < 0 {
			t.Error("RateLimit.BurstSize should be non-negative")
		}
	}
}

func TestCorsConfig(t *testing.T) {
	config := DefaultConfig()

	if config.API.Cors.Enabled {
		if len(config.API.Cors.AllowedOrigins) == 0 {
			t.Error("CORS AllowedOrigins should not be empty when enabled")
		}

		if len(config.API.Cors.AllowedMethods) == 0 {
			t.Error("CORS AllowedMethods should not be empty when enabled")
		}
	}
}

func TestP2PConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  P2PConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: P2PConfig{
				ListenAddr:     "/ip4/0.0.0.0/tcp/4001",
				BootstrapPeers: []string{},
				DialTimeout:    10 * time.Second,
				MaxConnections: 100,
			},
			wantErr: false,
		},
		{
			name: "invalid listen address",
			config: P2PConfig{
				ListenAddr:     "",
				BootstrapPeers: []string{},
				DialTimeout:    10 * time.Second,
				MaxConnections: 100,
			},
			wantErr: true,
		},
		{
			name: "zero dial timeout",
			config: P2PConfig{
				ListenAddr:     "/ip4/0.0.0.0/tcp/4001",
				BootstrapPeers: []string{},
				DialTimeout:    0,
				MaxConnections: 100,
			},
			wantErr: true,
		},
		{
			name: "zero max connections",
			config: P2PConfig{
				ListenAddr:     "/ip4/0.0.0.0/tcp/4001",
				BootstrapPeers: []string{},
				DialTimeout:    10 * time.Second,
				MaxConnections: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateP2PConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateP2PConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  AuthConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: AuthConfig{
				Enabled:     true,
				Method:      "jwt",
				Provider:    "jwt",
				TokenExpiry: 3600 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "disabled auth",
			config: AuthConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			config: AuthConfig{
				Enabled:  true,
				Method:   "",
				Provider: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigMerge(t *testing.T) {
	base := DefaultConfig()
	override := &Config{
		API: APIConfig{
			Listen: ":9999",
		},
	}

	merged := MergeConfigs(base, override)

	if merged.API.Listen != ":9999" {
		t.Errorf("Expected merged Listen = :9999, got %v", merged.API.Listen)
	}

	// Non-overridden fields should retain base values
	if merged.JWT.SecretKey != base.JWT.SecretKey {
		t.Error("Non-overridden JWT.SecretKey should retain base value")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty API listen address",
			config: &Config{
				API: APIConfig{
					Listen: "",
				},
			},
			wantErr: true,
		},
		{
			name: "negative max body size",
			config: &Config{
				API: APIConfig{
					Listen:      ":8080",
					MaxBodySize: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
