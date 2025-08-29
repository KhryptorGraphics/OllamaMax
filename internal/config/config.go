package config

import (
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

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return DefaultConfig()
}