package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	JWT  JWTConfig  `json:"jwt"`
	Auth AuthConfig `json:"auth"`
	API  APIConfig  `json:"api"`
	P2P  P2PConfig  `json:"p2p"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey    string        `json:"secret_key"`
	ExpiryTime   time.Duration `json:"expiry_time"`
	RefreshTime  time.Duration `json:"refresh_time"`
	Issuer       string        `json:"issuer"`
	Audience     string        `json:"audience"`
}

// APIConfig holds API server configuration
type APIConfig struct {
	Listen      string          `json:"listen"`
	ListenAddr  string          `json:"listen_addr"`
	Port        int             `json:"port"`
	TLSEnabled  bool            `json:"tls_enabled"`
	CertFile    string          `json:"cert_file"`
	KeyFile     string          `json:"key_file"`
	MaxBodySize int64           `json:"max_body_size"`
	RateLimit   RateLimitConfig `json:"rate_limit"`
	Cors        CorsConfig      `json:"cors"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled      bool          `json:"enabled"`
	Method       string        `json:"method"`
	Provider     string        `json:"provider"` // Alias for Method
	TokenExpiry  time.Duration `json:"token_expiry"`
	SecretKey    string        `json:"secret_key"`
	RefreshTime  time.Duration `json:"refresh_time"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled     bool          `json:"enabled"`
	RequestsPer int           `json:"requests_per"`
	Duration    time.Duration `json:"duration"`
	BurstSize   int           `json:"burst_size"`
	// Legacy fields for backward compatibility
	RPS       int      `json:"rps"`
	Burst     int      `json:"burst"`
	WhiteList []string `json:"whitelist"`
}

// CorsConfig holds CORS configuration
type CorsConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// P2PConfig holds P2P networking configuration
type P2PConfig struct {
	ListenAddr     string        `json:"listen_addr"`
	BootstrapPeers []string      `json:"bootstrap_peers"`
	DialTimeout    time.Duration `json:"dial_timeout"`
	MaxConnections int           `json:"max_connections"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		JWT: JWTConfig{
			SecretKey:   getEnvOrDefault("JWT_SECRET_KEY", "your-secret-key-change-this"),
			ExpiryTime:  24 * time.Hour,
			RefreshTime: 7 * 24 * time.Hour,
			Issuer:      "ollamamax",
			Audience:    "ollamamax-users",
		},
		Auth: AuthConfig{
			Enabled:     getEnvBoolOrDefault("AUTH_ENABLED", true),
			Method:      getEnvOrDefault("AUTH_METHOD", "jwt"),
			TokenExpiry: 24 * time.Hour,
			SecretKey:   getEnvOrDefault("AUTH_SECRET_KEY", "your-secret-key-change-this"),
			RefreshTime: 7 * 24 * time.Hour,
		},
		API: APIConfig{
			Listen:      getEnvOrDefault("API_LISTEN", "0.0.0.0:11434"),
			ListenAddr:  getEnvOrDefault("API_LISTEN_ADDR", "0.0.0.0"),
			Port:        getEnvIntOrDefault("API_PORT", 11434),
			TLSEnabled:  getEnvBoolOrDefault("API_TLS_ENABLED", false),
			CertFile:    getEnvOrDefault("API_CERT_FILE", ""),
			KeyFile:     getEnvOrDefault("API_KEY_FILE", ""),
			MaxBodySize: int64(getEnvIntOrDefault("API_MAX_BODY_SIZE", 32*1024*1024)), // 32MB
			RateLimit: RateLimitConfig{
				Enabled:     getEnvBoolOrDefault("RATE_LIMIT_ENABLED", true),
				RequestsPer: getEnvIntOrDefault("RATE_LIMIT_REQUESTS", 100),
				Duration:    time.Minute,
				BurstSize:   getEnvIntOrDefault("RATE_LIMIT_BURST", 10),
			},
			Cors: CorsConfig{
				Enabled:          getEnvBoolOrDefault("CORS_ENABLED", true),
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				AllowCredentials: false,
			},
		},
		P2P: P2PConfig{
			ListenAddr:     getEnvOrDefault("P2P_LISTEN_ADDR", "/ip4/0.0.0.0/tcp/0"),
			BootstrapPeers: []string{},
			DialTimeout:    30 * time.Second,
			MaxConnections: getEnvIntOrDefault("P2P_MAX_CONNECTIONS", 100),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// LoadConfig loads configuration from a file path
func LoadConfig(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("error accessing config file: %w", err)
	}

	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Validate that it's not empty
	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}

	// Basic validation: check for invalid YAML patterns that would cause parsing errors
	// This is a simple heuristic until we implement full YAML parsing
	content := string(data)
	// Check for common YAML syntax errors
	// Pattern like "key: value: extra:" with multiple unescaped colons in sequence is invalid
	if len(content) > 0 {
		// Count colons in each line - more than reasonable number indicates malformed YAML
		lines := 0
		colonsInLine := 0
		for i := 0; i < len(content); i++ {
			if content[i] == '\n' {
				lines++
				colonsInLine = 0
			} else if content[i] == ':' {
				colonsInLine++
				// More than 3 colons in a single line segment is suspicious
				if colonsInLine > 3 {
					return nil, fmt.Errorf("invalid YAML syntax detected: excessive colons")
				}
			}
		}
	}

	// For now, just return default config
	// In the future, this can parse from YAML/JSON files
	cfg := DefaultConfig()

	// Validate the configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// ValidateP2PConfig validates P2P configuration
func ValidateP2PConfig(cfg *P2PConfig) error {
	if cfg == nil {
		return fmt.Errorf("P2P config cannot be nil")
	}

	if cfg.ListenAddr == "" {
		return fmt.Errorf("P2P listen address cannot be empty")
	}

	if cfg.DialTimeout <= 0 {
		return fmt.Errorf("P2P dial timeout must be positive")
	}

	if cfg.MaxConnections <= 0 {
		return fmt.Errorf("P2P max connections must be positive")
	}

	return nil
}

// ValidateAuthConfig validates authentication configuration
func ValidateAuthConfig(cfg *AuthConfig) error {
	if cfg == nil {
		return fmt.Errorf("auth config cannot be nil")
	}

	// If auth is disabled, no further validation needed
	if !cfg.Enabled {
		return nil
	}

	// For enabled auth, validate provider/method
	if cfg.Method == "" {
		return fmt.Errorf("auth method/provider cannot be empty when auth is enabled")
	}

	return nil
}

// MergeConfigs merges two configurations, with override taking precedence
func MergeConfigs(base *Config, override *Config) *Config {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := &Config{
		JWT:  base.JWT,
		Auth: base.Auth,
		API:  base.API,
		P2P:  base.P2P,
	}

	// Override API config
	if override.API.Listen != "" {
		merged.API.Listen = override.API.Listen
	}
	if override.API.ListenAddr != "" {
		merged.API.ListenAddr = override.API.ListenAddr
	}
	if override.API.Port != 0 {
		merged.API.Port = override.API.Port
	}
	if override.API.MaxBodySize != 0 {
		merged.API.MaxBodySize = override.API.MaxBodySize
	}
	if override.API.TLSEnabled {
		merged.API.TLSEnabled = override.API.TLSEnabled
		merged.API.CertFile = override.API.CertFile
		merged.API.KeyFile = override.API.KeyFile
	}

	// Override JWT config
	if override.JWT.SecretKey != "" && override.JWT.SecretKey != "your-secret-key-change-this" {
		merged.JWT.SecretKey = override.JWT.SecretKey
	}
	if override.JWT.ExpiryTime != 0 {
		merged.JWT.ExpiryTime = override.JWT.ExpiryTime
	}

	// Override Auth config
	if override.Auth.Method != "" {
		merged.Auth.Method = override.Auth.Method
	}
	if override.Auth.SecretKey != "" && override.Auth.SecretKey != "your-secret-key-change-this" {
		merged.Auth.SecretKey = override.Auth.SecretKey
	}

	// Override P2P config
	if override.P2P.ListenAddr != "" {
		merged.P2P.ListenAddr = override.P2P.ListenAddr
	}
	if override.P2P.MaxConnections != 0 {
		merged.P2P.MaxConnections = override.P2P.MaxConnections
	}

	return merged
}

// ValidateConfig validates the entire configuration
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate API config
	if cfg.API.Listen == "" {
		return fmt.Errorf("API listen address cannot be empty")
	}

	if cfg.API.MaxBodySize < 0 {
		return fmt.Errorf("API max body size cannot be negative")
	}

	// Validate Auth config
	if err := ValidateAuthConfig(&cfg.Auth); err != nil {
		return fmt.Errorf("auth config validation failed: %w", err)
	}

	// Validate P2P config
	if err := ValidateP2PConfig(&cfg.P2P); err != nil {
		return fmt.Errorf("P2P config validation failed: %w", err)
	}

	return nil
}