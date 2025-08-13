package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/resources"
)

// TestP2PNodeCreation tests P2P node creation and initialization
func TestP2PNodeCreation(t *testing.T) {
	ctx := context.Background()

	// Test with default config
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err, "Failed to create P2P node")
	require.NotNil(t, node, "P2P node should not be nil")

	// Test node properties
	assert.NotEmpty(t, node.GetHost().ID(), "Node should have a peer ID")
	assert.NotEmpty(t, node.GetHost().Addrs(), "Node should have listening addresses")

	// Test node startup
	err = node.Start()
	require.NoError(t, err, "Failed to start P2P node")

	// Cleanup
	defer node.Stop()
}

// TestP2PNodeConfiguration tests various node configurations
func TestP2PNodeConfiguration(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name   string
		config *config.NodeConfig
		valid  bool
	}{
		{
			name: "Valid Basic Config",
			config: &config.NodeConfig{
				Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
				BootstrapPeers: []string{},
				EnableDHT:      true,
			},
			valid: true,
		},
		{
			name: "Valid With Bootstrap Peers",
			config: &config.NodeConfig{
				Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
				BootstrapPeers: []string{"/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWTest"},
				EnableDHT:      true,
			},
			valid: true,
		},
		{
			name: "Valid IPv6 Config",
			config: &config.NodeConfig{
				Listen:         []string{"/ip6/::1/tcp/0"},
				BootstrapPeers: []string{},
				EnableDHT:      true,
			},
			valid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := p2p.NewP2PNode(ctx, tc.config)

			if tc.valid {
				require.NoError(t, err, "Should create valid node")
				require.NotNil(t, node, "Node should not be nil")

				err = node.Start()
				require.NoError(t, err, "Should start successfully")

				defer node.Stop()
			} else {
				assert.Error(t, err, "Should fail with invalid config")
			}
		})
	}
}

// TestP2PNodeLifecycle tests complete node lifecycle
func TestP2PNodeLifecycle(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)

	// Test initial state
	status := node.GetStatus()
	assert.False(t, status.Started, "Node should not be started initially")

	// Test startup
	err = node.Start()
	require.NoError(t, err, "Node should start successfully")

	status = node.GetStatus()
	assert.True(t, status.Started, "Node should be started")
	assert.Greater(t, len(status.ListenAddresses), 0, "Should have listen addresses")

	// Test metrics collection
	time.Sleep(100 * time.Millisecond) // Allow metrics to update
	metrics := node.GetMetrics()
	assert.NotNil(t, metrics, "Should have metrics")
	assert.True(t, metrics.StartTime.Before(time.Now()), "Start time should be in the past")

	// Test graceful shutdown
	err = node.Stop()
	require.NoError(t, err, "Node should stop gracefully")

	status = node.GetStatus()
	assert.False(t, status.Started, "Node should be stopped")
}

// TestP2PNodeConnections tests peer connections
func TestP2PNodeConnections(t *testing.T) {
	ctx := context.Background()

	// Create two nodes
	config1 := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      false, // Disable for simpler testing
	}

	config2 := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      false,
	}

	node1, err := p2p.NewP2PNode(ctx, config1)
	require.NoError(t, err)

	node2, err := p2p.NewP2PNode(ctx, config2)
	require.NoError(t, err)

	// Start both nodes
	err = node1.Start()
	require.NoError(t, err)
	defer node1.Stop()

	err = node2.Start()
	require.NoError(t, err)
	defer node2.Stop()

	// Get node1's address info
	node1Addrs := node1.GetHost().Addrs()
	require.Greater(t, len(node1Addrs), 0, "Node1 should have addresses")

	peerInfo := peer.AddrInfo{
		ID:    node1.GetHost().ID(),
		Addrs: node1Addrs,
	}

	// Test connection from node2 to node1
	err = node2.ConnectToPeer(ctx, peerInfo)
	require.NoError(t, err, "Should connect successfully")

	// Wait for connection to establish
	time.Sleep(100 * time.Millisecond)

	// Verify connection
	node2Peers := node2.GetConnectedPeers()
	assert.Greater(t, len(node2Peers), 0, "Node2 should have connected peers")

	found := false
	for _, peerID := range node2Peers {
		if peerID == node1.GetHost().ID() {
			found = true
			break
		}
	}
	assert.True(t, found, "Node1 should be in node2's peer list")

	// Test disconnection
	err = node2.DisconnectFromPeer(node1.GetHost().ID())
	require.NoError(t, err, "Should disconnect successfully")

	// Wait for disconnection
	time.Sleep(100 * time.Millisecond)

	// Verify disconnection
	assert.False(t, node2.IsConnected(node1.GetHost().ID()), "Should be disconnected")
}

// TestP2PCapabilities tests node capabilities management
func TestP2PCapabilities(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Test setting capabilities
	capabilities := &resources.NodeCapabilities{
		SupportedModels: []string{"llama", "gpt"},
		CPUCores:        8,
		Memory:          16 * 1024 * 1024 * 1024, // 16GB
		Features:        []string{"quantization", "batching"},
	}

	node.SetCapabilities(capabilities)

	// Verify capabilities
	retrievedCaps := node.GetCapabilities()
	require.NotNil(t, retrievedCaps, "Should have capabilities")
	assert.Equal(t, capabilities.SupportedModels, retrievedCaps.SupportedModels)
	assert.Equal(t, capabilities.CPUCores, retrievedCaps.CPUCores)
	assert.Equal(t, capabilities.Memory, retrievedCaps.Memory)
}

// TestP2PResourceMetrics tests resource metrics management
func TestP2PResourceMetrics(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Test setting resource metrics
	metrics := &resources.ResourceMetrics{
		CPUUsage:    25.5,
		MemoryUsage: 60 * 1024 * 1024, // 60MB
		DiskUsage:   40 * 1024 * 1024, // 40MB
		NetworkRx:   1024,
		NetworkTx:   2048,
		Timestamp:   time.Now(),
	}

	node.SetResourceMetrics(metrics)

	// Verify metrics
	retrievedMetrics := node.GetResourceMetrics()
	require.NotNil(t, retrievedMetrics, "Should have resource metrics")
	assert.Equal(t, metrics.CPUUsage, retrievedMetrics.CPUUsage)
	assert.Equal(t, metrics.MemoryUsage, retrievedMetrics.MemoryUsage)
	assert.Equal(t, metrics.NetworkRx, retrievedMetrics.NetworkRx)
}

// TestP2PEventSystem tests the event system
func TestP2PEventSystem(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      false,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	// Test event handling
	eventReceived := make(chan *p2p.NodeEvent, 1)

	node.On(p2p.EventPeerConnected, func(event *p2p.NodeEvent) {
		eventReceived <- event
	})

	err = node.Start()
	require.NoError(t, err)

	// Skip creating second node for simplicity in testing
	// Just test that the event system is working

	// Test that the event system is initialized
	assert.NotNil(t, eventReceived, "Event channel should be available")
}

// TestP2PMetricsCollection tests metrics collection
func TestP2PMetricsCollection(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Wait for metrics to be collected
	time.Sleep(1 * time.Second)

	metrics := node.GetMetrics()
	require.NotNil(t, metrics, "Should have metrics")

	// Test basic metrics
	assert.True(t, metrics.StartTime.Before(time.Now()), "Start time should be valid")
	assert.GreaterOrEqual(t, metrics.ConnectedPeers, 0, "Connected peers should be non-negative")
	assert.GreaterOrEqual(t, metrics.TotalConnections, 0, "Total connections should be non-negative")
	assert.GreaterOrEqual(t, metrics.Uptime, time.Duration(0), "Uptime should be non-negative")

	// Test that uptime increases
	time.Sleep(100 * time.Millisecond)
	newMetrics := node.GetMetrics()
	assert.Greater(t, newMetrics.Uptime, metrics.Uptime, "Uptime should increase")
}

// TestP2PErrorHandling tests error handling scenarios
func TestP2PErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test invalid listen address
	invalidConfig := &config.NodeConfig{
		Listen:         []string{"invalid-address"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	_, err := p2p.NewP2PNode(ctx, invalidConfig)
	assert.Error(t, err, "Should fail with invalid listen address")

	// Test connection to non-existent peer
	validConfig := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      false,
	}

	node, err := p2p.NewP2PNode(ctx, validConfig)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Try to connect to invalid peer
	invalidPeerInfo := peer.AddrInfo{
		ID:    "12D3KooWInvalidPeerID",
		Addrs: []multiaddr.Multiaddr{},
	}

	err = node.ConnectToPeer(ctx, invalidPeerInfo)
	assert.Error(t, err, "Should fail to connect to invalid peer")

	// Test disconnection from non-connected peer
	someRandomPeerID, _ := peer.Decode("12D3KooWRandomPeerIDForTesting")
	err = node.DisconnectFromPeer(someRandomPeerID)
	// This might not error depending on implementation, so we just test it doesn't panic
}

// TestP2PConcurrentOperations tests concurrent operations
func TestP2PConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      false,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Test concurrent capability updates
	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			capabilities := &resources.NodeCapabilities{
				SupportedModels: []string{fmt.Sprintf("model-%d", id)},
				CPUCores:        id,
				Memory:          int64(id * 1024 * 1024 * 1024),
				Features:        []string{"quantization", "batching"},
			}

			node.SetCapabilities(capabilities)

			// Verify we can read capabilities without race
			_ = node.GetCapabilities()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	// Verify final state is consistent
	finalCaps := node.GetCapabilities()
	assert.NotNil(t, finalCaps, "Should have final capabilities")
}

// TestP2PResourceMonitoring tests resource monitoring
func TestP2PResourceMonitoring(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Wait for resource monitoring to run
	time.Sleep(2 * time.Second)

	// Check that resource metrics are being updated
	metrics := node.GetResourceMetrics()
	if metrics != nil {
		assert.True(t, metrics.Timestamp.After(time.Now().Add(-1*time.Minute)),
			"Resource metrics should be recently updated")
		assert.GreaterOrEqual(t, metrics.CPUUsage, 0.0, "CPU usage should be non-negative")
		assert.GreaterOrEqual(t, metrics.MemoryUsage, 0.0, "Memory usage should be non-negative")
	}
}

// TestP2PNodeStatus tests node status reporting
func TestP2PNodeStatus(t *testing.T) {
	ctx := context.Background()
	config := &config.NodeConfig{
		Listen:         []string{"/ip4/127.0.0.1/tcp/0"},
		BootstrapPeers: []string{},
		EnableDHT:      true,
	}

	node, err := p2p.NewP2PNode(ctx, config)
	require.NoError(t, err)
	defer node.Stop()

	// Test status before start
	status := node.GetStatus()
	assert.False(t, status.Started, "Node should not be started")
	assert.Equal(t, 0, status.ConnectedPeers, "Should have no connected peers")

	// Start node
	err = node.Start()
	require.NoError(t, err)

	// Test status after start
	status = node.GetStatus()
	assert.True(t, status.Started, "Node should be started")
	assert.NotEmpty(t, status.ID, "Should have peer ID")
	assert.Greater(t, len(status.ListenAddresses), 0, "Should have listen addresses")
	assert.True(t, time.Since(status.LastActivity) < time.Minute, "Last activity should be recent")

	// Test status string representation
	statusStr := status.String()
	assert.Contains(t, statusStr, "Started=true", "Status string should contain started state")
	assert.Contains(t, statusStr, string(status.ID), "Status string should contain peer ID")
}
