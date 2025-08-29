package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Default(t *testing.T) {
	// Test loading default configuration
	cfg, err := Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values
	assert.NotEmpty(t, cfg.Node.ID)
	assert.Equal(t, []string{"/ip4/0.0.0.0/tcp/8080"}, cfg.Node.Listen)
	assert.True(t, cfg.Node.EnableNoise)
	assert.Equal(t, 10, cfg.Node.ConnMgrLow)
	assert.Equal(t, 100, cfg.Node.ConnMgrHigh)
	assert.Equal(t, time.Minute, cfg.Node.ConnMgrGrace)
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("OLLAMA_NODE_ID", "test-node-123")
	os.Setenv("OLLAMA_LISTEN", "/ip4/127.0.0.1/tcp/9090")
	os.Setenv("OLLAMA_ENABLE_NAT", "true")
	defer func() {
		os.Unsetenv("OLLAMA_NODE_ID")
		os.Unsetenv("OLLAMA_LISTEN")
		os.Unsetenv("OLLAMA_ENABLE_NAT")
	}()

	cfg, err := Load("")
	require.NoError(t, err)

	// Verify environment variables are used
	assert.Equal(t, "test-node-123", cfg.Node.ID)
	assert.Contains(t, cfg.Node.Listen, "/ip4/127.0.0.1/tcp/9090")
	assert.True(t, cfg.Node.EnableNATService)
}

func TestValidateConfig_Valid(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
		Auth: &AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: 24 * time.Hour,
			SecretKey:   "test-secret-key-32-characters!",
			Issuer:      "ollama",
			Audience:    "ollama-api",
		},
		API: &APIConfig{
			Host: "localhost",
			Port: 8080,
			TLS: &TLSConfig{
				Enabled:  false,
				CertFile: "",
				KeyFile:  "",
			},
		},
		Database: &DatabaseConfig{
			Driver: "sqlite",
			DSN:    "./ollama.db",
		},
	}

	err := ValidateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_Invalid_NodeID(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "", // Empty ID should fail
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node ID cannot be empty")
}

func TestValidateConfig_Invalid_Listen(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{}, // Empty listen addresses should fail
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one listen address required")
}

func TestValidateConfig_Invalid_ConnectionManager(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   100, // Low > High should fail
			ConnMgrHigh:  50,
			ConnMgrGrace: time.Minute,
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection manager low must be less than high")
}

func TestValidateConfig_Auth_SecretKey(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
		Auth: &AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: 24 * time.Hour,
			SecretKey:   "too-short", // Too short secret key
			Issuer:      "ollama",
			Audience:    "ollama-api",
		},
	}

	err := ValidateConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret key must be at least 32 characters")
}

func TestValidateConfig_Database(t *testing.T) {
	testCases := []struct {
		name     string
		dbConfig *DatabaseConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid SQLite",
			dbConfig: &DatabaseConfig{
				Driver: "sqlite",
				DSN:    "./test.db",
			},
			wantErr: false,
		},
		{
			name: "Valid PostgreSQL",
			dbConfig: &DatabaseConfig{
				Driver: "postgres",
				DSN:    "postgres://user:pass@localhost/db?sslmode=disable",
			},
			wantErr: false,
		},
		{
			name: "Invalid Driver",
			dbConfig: &DatabaseConfig{
				Driver: "invalid",
				DSN:    "./test.db",
			},
			wantErr: true,
			errMsg:  "unsupported database driver",
		},
		{
			name: "Empty DSN",
			dbConfig: &DatabaseConfig{
				Driver: "sqlite",
				DSN:    "",
			},
			wantErr: true,
			errMsg:  "database DSN cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				Node: &NodeConfig{
					ID:           "test-node",
					Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
					EnableNoise:  true,
					ConnMgrLow:   10,
					ConnMgrHigh:  100,
					ConnMgrGrace: time.Minute,
				},
				Database: tc.dbConfig,
			}

			err := ValidateConfig(cfg)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetNodeConfig(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
	}

	nodeConfig := cfg.GetNodeConfig()
	assert.NotNil(t, nodeConfig)
	assert.Equal(t, "test-node", nodeConfig.ID)
	assert.Equal(t, cfg.Node.Listen, nodeConfig.Listen)
}

func TestGetAuthConfig(t *testing.T) {
	cfg := &Config{
		Auth: &AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: 24 * time.Hour,
			SecretKey:   "test-secret-key-32-characters!",
			Issuer:      "ollama",
			Audience:    "ollama-api",
		},
	}

	authConfig := cfg.GetAuthConfig()
	assert.NotNil(t, authConfig)
	assert.True(t, authConfig.Enabled)
	assert.Equal(t, "jwt", authConfig.Method)
}

func TestMergeConfigs(t *testing.T) {
	base := &Config{
		Node: &NodeConfig{
			ID:           "base-node",
			Listen:       []string{"/ip4/0.0.0.0/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
		Auth: &AuthConfig{
			Enabled: false,
		},
	}

	override := &Config{
		Node: &NodeConfig{
			ID:     "override-node",
			Listen: []string{"/ip4/127.0.0.1/tcp/9090"},
		},
		Auth: &AuthConfig{
			Enabled: true,
			Method:  "jwt",
		},
	}

	merged := MergeConfigs(base, override)

	// Node config should be merged
	assert.Equal(t, "override-node", merged.Node.ID)
	assert.Equal(t, []string{"/ip4/127.0.0.1/tcp/9090"}, merged.Node.Listen)
	assert.True(t, merged.Node.EnableNoise) // From base
	assert.Equal(t, 10, merged.Node.ConnMgrLow) // From base

	// Auth config should be merged
	assert.True(t, merged.Auth.Enabled)
	assert.Equal(t, "jwt", merged.Auth.Method)
}

func TestNodeConfig_DefaultBootstrapPeers(t *testing.T) {
	cfg := &NodeConfig{}
	cfg.SetDefaultBootstrapPeers()

	assert.NotEmpty(t, cfg.BootstrapPeers)
	assert.Contains(t, cfg.BootstrapPeers, "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN")
}

func TestTLSConfig_Validation(t *testing.T) {
	testCases := []struct {
		name      string
		tlsConfig *TLSConfig
		wantErr   bool
		errMsg    string
	}{
		{
			name: "TLS Disabled",
			tlsConfig: &TLSConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "TLS Enabled with Certs",
			tlsConfig: &TLSConfig{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
			wantErr: false,
		},
		{
			name: "TLS Enabled without Cert",
			tlsConfig: &TLSConfig{
				Enabled:  true,
				CertFile: "",
				KeyFile:  "key.pem",
			},
			wantErr: true,
			errMsg:  "cert file required when TLS is enabled",
		},
		{
			name: "TLS Enabled without Key",
			tlsConfig: &TLSConfig{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "",
			},
			wantErr: true,
			errMsg:  "key file required when TLS is enabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTLSConfig(tc.tlsConfig)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	cfg := &Config{
		Node: &NodeConfig{
			ID:           "test-node",
			Listen:       []string{"/ip4/127.0.0.1/tcp/8080"},
			EnableNoise:  true,
			ConnMgrLow:   10,
			ConnMgrHigh:  100,
			ConnMgrGrace: time.Minute,
		},
		Auth: &AuthConfig{
			Enabled:     true,
			Method:      "jwt",
			TokenExpiry: 24 * time.Hour,
			SecretKey:   "test-secret-key-32-characters!",
			Issuer:      "ollama",
			Audience:    "ollama-api",
		},
	}

	tempFile := "/tmp/test-config.yaml"
	defer os.Remove(tempFile)

	err := SaveConfig(cfg, tempFile)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tempFile)
	assert.NoError(t, err)

	// Load and verify
	loadedCfg, err := Load(tempFile)
	require.NoError(t, err)
	assert.Equal(t, cfg.Node.ID, loadedCfg.Node.ID)
	assert.Equal(t, cfg.Auth.Method, loadedCfg.Auth.Method)
}

func TestConfigDefaults(t *testing.T) {
	cfg := &Config{}
	ApplyDefaults(cfg)

	// Verify defaults are applied
	assert.NotNil(t, cfg.Node)
	assert.NotEmpty(t, cfg.Node.ID)
	assert.NotEmpty(t, cfg.Node.Listen)
	assert.True(t, cfg.Node.EnableNoise)
	assert.Greater(t, cfg.Node.ConnMgrHigh, cfg.Node.ConnMgrLow)

	assert.NotNil(t, cfg.API)
	assert.NotEmpty(t, cfg.API.Host)
	assert.Greater(t, cfg.API.Port, 0)

	assert.NotNil(t, cfg.Database)
	assert.NotEmpty(t, cfg.Database.Driver)
	assert.NotEmpty(t, cfg.Database.DSN)
}