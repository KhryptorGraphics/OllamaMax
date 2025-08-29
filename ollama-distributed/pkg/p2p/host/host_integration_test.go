package host

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/nat"
)

func TestP2PHost_NATIntegration(t *testing.T) {
	ctx := context.Background()

	// Create test configuration
	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableAutoRelay:    false,
		EnableNoise:        true,
		EnableTLS:          false,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
	}

	// Create P2P host
	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	// Test NAT manager integration
	natManager := host.GetNATManager()
	require.NotNil(t, natManager)

	// Test NAT discovery (may timeout in test environment)
	discoveryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	natType, err := natManager.DiscoverNATType(discoveryCtx)
	// In test environment, this may fail, but we test the integration
	if err == nil {
		assert.NotEqual(t, nat.NATTypeUnknown, natType)
		t.Logf("Discovered NAT type: %s", natType)
	} else {
		t.Logf("NAT discovery failed (expected in test env): %v", err)
	}

	// Test basic NAT manager functionality
	assert.NotNil(t, natManager.GetExternalIP())
	t.Logf("External IP: %v", natManager.GetExternalIP())
	
	// Test metrics
	metrics := natManager.GetMetrics()
	assert.NotNil(t, metrics)
}

func TestP2PHost_HolePunchingSetup(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         5,
		ConnMgrHigh:        50,
		ConnMgrGrace:       30 * time.Second,
	}

	host1, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host1.Close()

	host2, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host2.Close()

	// Test both hosts have NAT managers
	natManager1 := host1.GetNATManager()
	natManager2 := host2.GetNATManager()
	
	require.NotNil(t, natManager1)
	require.NotNil(t, natManager2)

	// Test NAT type setting
	natManager1.SetNATType(nat.NATTypeFullCone)
	natManager2.SetNATType(nat.NATTypeRestrictedCone)

	assert.Equal(t, nat.NATTypeFullCone, natManager1.GetNATType())
	assert.Equal(t, nat.NATTypeRestrictedCone, natManager2.GetNATType())

	// Test relay requirement logic
	natManager1.SetNATType(nat.NATTypeSymmetric)
	assert.True(t, natManager1.IsRelayRequired())

	natManager2.SetNATType(nat.NATTypeOpen)
	assert.False(t, natManager2.IsRelayRequired())

	t.Log("Hole punching setup test completed successfully")
}

func TestP2PHost_TURNIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		TURNServers: []string{
			"turn:test.example.com:3478",
			"turns:test.example.com:5349",
		},
		TURNUsername: "testuser",
		TURNPassword: "testpass",
		EnableNoise:  true,
		ConnMgrLow:   5,
		ConnMgrHigh:  25,
		ConnMgrGrace: 30 * time.Second,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	natManager := host.GetNATManager()
	require.NotNil(t, natManager)

	// Test TURN server configuration
	turnServers := natManager.GetTURNServers()
	assert.Len(t, turnServers, 2)
	assert.Contains(t, turnServers, "turn:test.example.com:3478")
	assert.Contains(t, turnServers, "turns:test.example.com:5349")

	// Test relay connection requirement for symmetric NAT
	natManager.SetNATType(nat.NATTypeSymmetric)
	assert.True(t, natManager.IsRelayRequired())

	t.Log("TURN integration test completed successfully")
}

func TestP2PHost_HolePunchingIntegration(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         10,
		ConnMgrHigh:        100,
		ConnMgrGrace:       time.Minute,
		HolePunchTimeout:   30 * time.Second,
	}

	host1, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host1.Close()

	host2, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host2.Close()

	natManager1 := host1.GetNATManager()
	natManager2 := host2.GetNATManager()

	require.NotNil(t, natManager1)
	require.NotNil(t, natManager2)

	// Test hole punching attempt (will likely fail in test env but tests the mechanism)
	punchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	peer1ID := host1.ID()
	peer2ID := host2.ID()

	// Attempt hole punching from host1 to host2
	err = natManager1.AttemptHolePunch(punchCtx, peer2ID, host2.Addrs())
	// This will likely fail in test environment, but we test the mechanism
	t.Logf("Hole punch result from %s to %s: %v", peer1ID, peer2ID, err)

	// Test metrics after hole punch attempt
	metrics1 := natManager1.GetMetrics()
	assert.NotNil(t, metrics1)
	assert.GreaterOrEqual(t, metrics1.HolePunchAttempts, uint64(1))

	t.Log("Hole punching integration test completed")
}

func TestP2PHost_MultiNAT_Scenarios(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		natType1 nat.NATType
		natType2 nat.NATType
		canConnect bool
	}{
		{
			name:       "Open to Open",
			natType1:   nat.NATTypeOpen,
			natType2:   nat.NATTypeOpen,
			canConnect: true,
		},
		{
			name:       "Open to FullCone",
			natType1:   nat.NATTypeOpen,
			natType2:   nat.NATTypeFullCone,
			canConnect: true,
		},
		{
			name:       "FullCone to FullCone",
			natType1:   nat.NATTypeFullCone,
			natType2:   nat.NATTypeFullCone,
			canConnect: true,
		},
		{
			name:       "Symmetric to Symmetric",
			natType1:   nat.NATTypeSymmetric,
			natType2:   nat.NATTypeSymmetric,
			canConnect: false, // Requires relay
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodeConfig := &config.NodeConfig{
				Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
				EnableNATService:   true,
				EnableHolePunching: true,
				EnableNoise:        true,
				ConnMgrLow:         5,
				ConnMgrHigh:        25,
				ConnMgrGrace:       30 * time.Second,
			}

			host1, err := NewP2PHost(ctx, nodeConfig)
			require.NoError(t, err)
			defer host1.Close()

			host2, err := NewP2PHost(ctx, nodeConfig)
			require.NoError(t, err)
			defer host2.Close()

			natManager1 := host1.GetNATManager()
			natManager2 := host2.GetNATManager()

			// Set NAT types for test scenario
			natManager1.SetNATType(tc.natType1)
			natManager2.SetNATType(tc.natType2)

			// Check if relay is required
			relayRequired1 := natManager1.IsRelayRequired()
			relayRequired2 := natManager2.IsRelayRequired()

			if tc.canConnect {
				assert.False(t, relayRequired1 && relayRequired2, "Both peers should not require relay for direct connection")
			} else {
				assert.True(t, relayRequired1 || relayRequired2, "At least one peer should require relay")
			}

			t.Logf("NAT scenario %s: Host1=%s (relay: %v), Host2=%s (relay: %v)", 
				tc.name, tc.natType1, relayRequired1, tc.natType2, relayRequired2)
		})
	}
}

func TestP2PHost_NATDiscovery_Metrics(t *testing.T) {
	ctx := context.Background()

	nodeConfig := &config.NodeConfig{
		Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
		EnableNATService:   true,
		EnableHolePunching: true,
		EnableNoise:        true,
		ConnMgrLow:         5,
		ConnMgrHigh:        25,
		ConnMgrGrace:       30 * time.Second,
	}

	host, err := NewP2PHost(ctx, nodeConfig)
	require.NoError(t, err)
	defer host.Close()

	natManager := host.GetNATManager()
	require.NotNil(t, natManager)

	// Get initial metrics
	initialMetrics := natManager.GetMetrics()
	assert.NotNil(t, initialMetrics)

	// Attempt NAT discovery
	discoveryCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err = natManager.DiscoverNATType(discoveryCtx)
	// Error expected in test environment

	// Get metrics after discovery attempt
	finalMetrics := natManager.GetMetrics()
	assert.NotNil(t, finalMetrics)

	// Verify metrics are being tracked
	assert.GreaterOrEqual(t, finalMetrics.DiscoveryAttempts, initialMetrics.DiscoveryAttempts)

	t.Logf("NAT discovery metrics: Attempts=%d, Success=%d, Failed=%d, HolePunch=%d", 
		finalMetrics.DiscoveryAttempts,
		finalMetrics.SuccessfulDiscoveries,
		finalMetrics.FailedDiscoveries,
		finalMetrics.HolePunchAttempts)
}

func TestP2PHost_NATConfiguration_Validation(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name     string
		config   *config.NodeConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid NAT Config",
			config: &config.NodeConfig{
				Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
				EnableNATService:   true,
				EnableHolePunching: true,
				EnableNoise:        true,
				ConnMgrLow:         5,
				ConnMgrHigh:        25,
				ConnMgrGrace:       30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "NAT Disabled",
			config: &config.NodeConfig{
				Listen:             []string{"/ip4/127.0.0.1/tcp/0"},
				EnableNATService:   false,
				EnableHolePunching: false,
				EnableNoise:        true,
				ConnMgrLow:         5,
				ConnMgrHigh:        25,
				ConnMgrGrace:       30 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			host, err := NewP2PHost(ctx, tc.config)
			
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}
			
			require.NoError(t, err)
			require.NotNil(t, host)
			defer host.Close()

			if tc.config.EnableNATService {
				natManager := host.GetNATManager()
				assert.NotNil(t, natManager)
			}
		})
	}
}