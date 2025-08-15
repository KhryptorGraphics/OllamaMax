package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

// TestPhase1BasicFunctionality tests the core Phase 1 implementation
func TestPhase1BasicFunctionality(t *testing.T) {
	t.Run("ConfigurationSystem", func(t *testing.T) {
		testConfigurationSystem(t)
	})

	t.Run("JWTTokenGeneration", func(t *testing.T) {
		testJWTTokenGeneration(t)
	})

	t.Run("ConfigurationValidation", func(t *testing.T) {
		testConfigurationValidation(t)
	})

	t.Run("ConfigurationDefaults", func(t *testing.T) {
		testConfigurationDefaults(t)
	})
}

func testConfigurationSystem(t *testing.T) {
	// Test configuration structure creation
	config := &config.DistributedConfig{}
	require.NotNil(t, config)

	// Set defaults
	config.SetDefaults()

	// Verify core defaults are set
	assert.Equal(t, "0.0.0.0", config.API.Host)
	assert.Equal(t, 8080, config.API.Port)
	assert.Equal(t, 9000, config.P2P.Port)
	assert.Equal(t, "raft", config.Consensus.Algorithm)
	assert.Equal(t, "development", config.Node.Environment)
	assert.Equal(t, "jwt", config.Auth.Provider)
	assert.Equal(t, 24*time.Hour, config.Auth.SessionTimeout)

	// Test configuration methods
	assert.True(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())

	// Test address getters
	assert.Equal(t, "0.0.0.0:8080", config.GetAPIAddress())
	assert.Equal(t, ":9000", config.GetP2PAddress())
	assert.Equal(t, ":9090", config.GetMetricsAddress())

	// Test node tags
	tags := config.GetNodeTags()
	assert.NotNil(t, tags)
}

func testJWTTokenGeneration(t *testing.T) {
	// Test basic JWT token generation
	token, err := security.GenerateJWT("test@example.com", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test admin token generation
	adminToken, err := security.GenerateJWT("admin@example.com", "admin")
	assert.NoError(t, err)
	assert.NotEmpty(t, adminToken)

	// Tokens should be different
	assert.NotEqual(t, token, adminToken)

	// Test with empty values
	emptyToken, err := security.GenerateJWT("", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, emptyToken)
}

func testConfigurationValidation(t *testing.T) {
	// Test valid configuration
	config := &config.DistributedConfig{}
	config.SetDefaults()
	config.Node.ID = "test-node-1"

	err := config.Validate()
	assert.NoError(t, err)

	// Test invalid API port
	config.API.Port = 0
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api.port must be between 1 and 65535")

	// Test invalid API port (too high)
	config.API.Port = 70000
	err = config.Validate()
	assert.Error(t, err)

	// Reset to valid port
	config.API.Port = 8080

	// Test invalid P2P port
	config.P2P.Port = -1
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "p2p.port must be between 1 and 65535")

	// Reset to valid port
	config.P2P.Port = 9000

	// Test missing consensus algorithm
	config.Consensus.Algorithm = ""
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "consensus.algorithm is required")

	// Test missing node ID
	config.Consensus.Algorithm = "raft"
	config.Node.ID = ""
	err = config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node.id is required")
}

func testConfigurationDefaults(t *testing.T) {
	config := &config.DistributedConfig{}
	config.SetDefaults()

	// Test Node defaults
	assert.Equal(t, "development", config.Node.Environment)

	// Test API defaults
	assert.Equal(t, "0.0.0.0", config.API.Host)
	assert.Equal(t, 8080, config.API.Port)
	assert.Equal(t, 30*time.Second, config.API.Timeout)

	// Test P2P defaults
	assert.Equal(t, 9000, config.P2P.Port)
	assert.Equal(t, 50, config.P2P.MaxPeers)

	// Test Consensus defaults
	assert.Equal(t, "raft", config.Consensus.Algorithm)
	assert.Equal(t, 5*time.Second, config.Consensus.ElectionTimeout)
	assert.Equal(t, 1*time.Second, config.Consensus.HeartbeatInterval)

	// Test Scheduler defaults
	assert.Equal(t, "round_robin", config.Scheduler.Strategy)
	assert.Equal(t, 100, config.Scheduler.MaxConcurrent)
	assert.Equal(t, 60*time.Second, config.Scheduler.Timeout)

	// Test Auth defaults
	assert.Equal(t, "jwt", config.Auth.Provider)
	assert.Equal(t, 24*time.Hour, config.Auth.SessionTimeout)

	// Test WebSocket defaults
	assert.Equal(t, "/ws", config.WebSocket.Path)
	assert.Equal(t, 1000, config.WebSocket.MaxConnections)

	// Test Metrics defaults
	assert.Equal(t, 9090, config.Metrics.Port)
	assert.Equal(t, "/metrics", config.Metrics.Path)
	assert.Equal(t, 15*time.Second, config.Metrics.Interval)

	// Test Logging defaults
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, "stdout", config.Logging.Output)
}

func TestConfigurationCloning(t *testing.T) {
	// Create original config
	original := &config.DistributedConfig{}
	original.SetDefaults()
	original.Node.ID = "original-node"
	original.Node.Name = "Original Node"
	original.API.Port = 8080

	// Clone the configuration
	cloned := original.Clone()
	require.NotNil(t, cloned)

	// Verify clone has same values
	assert.Equal(t, original.Node.ID, cloned.Node.ID)
	assert.Equal(t, original.Node.Name, cloned.Node.Name)
	assert.Equal(t, original.API.Port, cloned.API.Port)

	// Modify clone and verify original is unchanged
	cloned.Node.ID = "cloned-node"
	cloned.API.Port = 9080

	assert.Equal(t, "original-node", original.Node.ID)
	assert.Equal(t, 8080, original.API.Port)
	assert.Equal(t, "cloned-node", cloned.Node.ID)
	assert.Equal(t, 9080, cloned.API.Port)
}

func TestConfigurationMerging(t *testing.T) {
	// Create base config
	base := &config.DistributedConfig{}
	base.SetDefaults()
	base.Node.ID = "base-node"
	base.API.Port = 8080

	// Create override config
	override := &config.DistributedConfig{}
	override.Node.ID = "override-node"
	override.Node.Name = "Override Node"
	override.API.Port = 9080

	// Merge configurations
	base.MergeConfig(override)

	// Verify merge results
	assert.Equal(t, "override-node", base.Node.ID)
	assert.Equal(t, "Override Node", base.Node.Name)
	assert.Equal(t, 9080, base.API.Port)
}

func TestNodeCapabilities(t *testing.T) {
	config := &config.DistributedConfig{}
	config.SetDefaults()

	// Test default capabilities
	caps := config.Node.Capabilities
	assert.False(t, caps.Inference)
	assert.False(t, caps.Storage)
	assert.False(t, caps.Coordination)
	assert.False(t, caps.Gateway)

	// Test setting capabilities
	config.Node.Capabilities.Inference = true
	config.Node.Capabilities.Storage = true

	assert.True(t, config.Node.Capabilities.Inference)
	assert.True(t, config.Node.Capabilities.Storage)
	assert.False(t, config.Node.Capabilities.Coordination)
	assert.False(t, config.Node.Capabilities.Gateway)
}

func TestStaticRelaysParsing(t *testing.T) {
	config := &config.DistributedConfig{}
	config.SetDefaults()

	// Test empty static relays
	relays, err := config.Node.ParseStaticRelays()
	assert.NoError(t, err)
	assert.Empty(t, relays)

	// Test with static relays
	config.Node.StaticRelays = []string{
		"/ip4/127.0.0.1/tcp/4001/p2p/QmRelay1",
		"/ip4/127.0.0.1/tcp/4002/p2p/QmRelay2",
	}

	relays, err = config.Node.ParseStaticRelays()
	assert.NoError(t, err)
	assert.Len(t, relays, 2)
	assert.Equal(t, "/ip4/127.0.0.1/tcp/4001/p2p/QmRelay1", relays[0])
	assert.Equal(t, "/ip4/127.0.0.1/tcp/4002/p2p/QmRelay2", relays[1])
}
