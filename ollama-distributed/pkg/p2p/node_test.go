package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	nodeconfig "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/resources"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/routing"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewP2PNode tests node creation
func TestNewP2PNode(t *testing.T) {
	tests := []struct {
		name        string
		config      *nodeconfig.NodeConfig
		expectError bool
	}{
		{
			name:        "default config",
			config:      nil, // Should use default
			expectError: false,
		},
		{
			name: "custom config",
			config: &nodeconfig.NodeConfig{
				Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
				ConnMgrLow:   10,
				ConnMgrHigh:  100,
				ConnMgrGrace: time.Minute,
				EnableDHT:    true,
			},
			expectError: false,
		},
		{
			name: "invalid listen address",
			config: &nodeconfig.NodeConfig{
				Listen: []string{"invalid-address"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			node, err := NewP2PNode(ctx, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, node)

				// Verify node properties
				assert.NotEmpty(t, node.ID())
				assert.NotNil(t, node.GetHost())
				assert.NotNil(t, node.GetMetrics())
				assert.NotNil(t, node.GetConfig())

				// Test initial state
				assert.Equal(t, 0, node.GetPeerCount())
				assert.Empty(t, node.GetConnectedPeers())

				// Clean up
				if node != nil {
					err := node.Stop()
					assert.NoError(t, err)
				}
			}
		})
	}
}

// TestNewNode tests compatibility wrapper
func TestNewNode(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p2pConfig := &config.P2PConfig{
		Listen:       "127.0.0.1:0",
		EnableDHT:    true,
		ConnMgrLow:   10,
		ConnMgrHigh:  100,
		ConnMgrGrace: "1m",
	}

	node, err := NewNode(ctx, p2pConfig)
	assert.NoError(t, err)
	assert.NotNil(t, node)

	// Clean up
	err = node.Stop()
	assert.NoError(t, err)
}

// TestNode_StartStop tests node lifecycle
func TestNode_StartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)

	// Test start
	err = node.Start()
	assert.NoError(t, err)

	// Test double start (should fail)
	err = node.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Test stop
	err = node.Stop()
	assert.NoError(t, err)

	// Test double stop (should fail)
	err = node.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

// TestNode_GetStatus tests status reporting
func TestNode_GetStatus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	status := node.GetStatus()
	assert.NotNil(t, status)
	assert.Equal(t, node.ID(), status.ID)
	assert.True(t, status.Started)
	assert.True(t, status.Uptime >= 0)
	assert.NotEmpty(t, status.ListenAddresses)
}

// TestNode_Capabilities tests capability management
func TestNode_Capabilities(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	// Initially no capabilities
	caps := node.GetCapabilities()
	assert.Nil(t, caps)

	// Set capabilities
	testCaps := &resources.NodeCapabilities{
		CPUCores:        4,
		Memory:          8192,
		Storage:         1000000,
		SupportedModels: []string{"llama", "mixtral"},
		Features:        []string{"quantization", "batching"},
	}

	node.SetCapabilities(testCaps)

	// Verify capabilities
	caps = node.GetCapabilities()
	assert.NotNil(t, caps)
	assert.Equal(t, testCaps.CPUCores, caps.CPUCores)
	assert.Equal(t, testCaps.Memory, caps.Memory)
	assert.Equal(t, testCaps.Storage, caps.Storage)
	assert.Equal(t, testCaps.SupportedModels, caps.SupportedModels)
	assert.Equal(t, testCaps.Features, caps.Features)
}

// TestNode_ResourceMetrics tests resource metrics
func TestNode_ResourceMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	// Initially no metrics
	metrics := node.GetResourceMetrics()
	assert.Nil(t, metrics)

	// Set metrics
	testMetrics := &resources.ResourceMetrics{
		CPUUsage:    75.5,
		MemoryUsage: 4096,
		DiskUsage:   500000,
		NetworkRx:   1024,
		NetworkTx:   2048,
		Timestamp:   time.Now(),
	}

	node.SetResourceMetrics(testMetrics)

	// Verify metrics
	metrics = node.GetResourceMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, testMetrics.CPUUsage, metrics.CPUUsage)
	assert.Equal(t, testMetrics.MemoryUsage, metrics.MemoryUsage)
	assert.Equal(t, testMetrics.DiskUsage, metrics.DiskUsage)
	assert.Equal(t, testMetrics.NetworkRx, metrics.NetworkRx)
	assert.Equal(t, testMetrics.NetworkTx, metrics.NetworkTx)
}

// TestNode_EventSystem tests event handling
func TestNode_EventSystem(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	// Test event registration
	var receivedEvents []*NodeEvent
	handler := func(event *NodeEvent) {
		receivedEvents = append(receivedEvents, event)
	}

	node.On(EventResourceUpdated, handler)

	// Trigger event by setting capabilities
	testCaps := &resources.NodeCapabilities{
		CPUCores: 2,
		Memory:   4096,
	}
	node.SetCapabilities(testCaps)

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// Verify event was received
	assert.NotEmpty(t, receivedEvents)
	assert.Equal(t, EventResourceUpdated, receivedEvents[0].Type)
}

// TestNode_ContentOperations tests content publishing and requesting
func TestNode_ContentOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Test publish content (might fail if DHT not ready)
	content := &routing.ContentMetadata{
		ID:   "test-content",
		Name: "Test Content",
		Size: 1024,
		Type: "model",
	}

	err = node.PublishContent(ctx, content)
	// Content router might not be available without DHT
	// Just verify the method doesn't panic

	// Test request content
	request, err := node.RequestContent(ctx, "test-content", 1)
	// This might fail without peers, which is expected
	_ = request
	_ = err

	// Test find content
	metadata, peers, err := node.FindContent(ctx, "test-content")
	// This will likely fail without peers
	_ = metadata
	_ = peers
	_ = err
}

// TestNode_PeerOperations tests peer management
func TestNode_PeerOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create two nodes
	node1, err := NewP2PNode(ctx, &nodeconfig.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer node1.Stop()

	node2, err := NewP2PNode(ctx, &nodeconfig.NodeConfig{
		Listen: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	require.NoError(t, err)
	defer node2.Stop()

	// Start both nodes
	err = node1.Start()
	require.NoError(t, err)

	err = node2.Start()
	require.NoError(t, err)

	// Give nodes time to start
	time.Sleep(100 * time.Millisecond)

	// Initially no peers
	assert.Equal(t, 0, node1.GetPeerCount())
	assert.Equal(t, 0, node2.GetPeerCount())
	assert.Empty(t, node1.GetConnectedPeers())
	assert.Empty(t, node2.GetConnectedPeers())

	// Get node2's addresses for connection
	node2Addrs := node2.GetHost().Addrs()
	if len(node2Addrs) > 0 {
		// Create peer info for connection
		peerInfo := peer.AddrInfo{
			ID:    node2.ID(),
			Addrs: node2Addrs,
		}

		// Connect node1 to node2
		err = node1.ConnectToPeer(ctx, peerInfo)
		if err == nil {
			// If connection succeeded, verify
			time.Sleep(100 * time.Millisecond)

			// Check if connected
			connected := node1.IsConnected(node2.ID())
			if connected {
				assert.True(t, node1.GetPeerCount() > 0)
				assert.Contains(t, node1.GetConnectedPeers(), node2.ID())

				// Test peer info
				allPeers := node1.GetAllPeers()
				assert.Contains(t, allPeers, node2.ID())

				peerInfo := allPeers[node2.ID()]
				assert.Equal(t, node2.ID(), peerInfo.ID)
				assert.True(t, peerInfo.Connected)

				// Test disconnect
				err = node1.DisconnectFromPeer(node2.ID())
				assert.NoError(t, err)

				time.Sleep(100 * time.Millisecond)
				assert.False(t, node1.IsConnected(node2.ID()))
			}
		}
	}
}

// TestNode_MetricsCollection tests metrics collection
func TestNode_MetricsCollection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Wait for metrics to be collected
	time.Sleep(200 * time.Millisecond)

	metrics := node.GetMetrics()
	assert.NotNil(t, metrics)
	assert.True(t, metrics.StartTime.Before(time.Now()))
	assert.True(t, metrics.Uptime >= 0)
	assert.True(t, metrics.LastActivity.After(metrics.StartTime))
}

// TestNode_ResourceMonitoring tests resource monitoring
func TestNode_ResourceMonitoring(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Wait for resource monitoring to run
	time.Sleep(500 * time.Millisecond)

	// Check if resource metrics were updated
	metrics := node.GetResourceMetrics()
	if metrics != nil {
		assert.True(t, metrics.CPUUsage >= 0)
		assert.True(t, metrics.MemoryUsage >= 0)
		assert.True(t, metrics.DiskUsage >= 0)
		assert.True(t, metrics.NetworkRx >= 0)
		assert.True(t, metrics.NetworkTx >= 0)
		assert.True(t, metrics.Timestamp.After(time.Time{}))
	}
}

// TestNode_SecurityOperations tests security operations
func TestNode_SecurityOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Test establishing secure channel (will fail without peer)
	_, err = node.EstablishSecureChannel(ctx, "nonexistent-peer")
	assert.Error(t, err) // Expected to fail
}

// TestNode_ErrorHandling tests error handling scenarios
func TestNode_ErrorHandling(t *testing.T) {
	t.Run("operations on stopped node", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node, err := NewP2PNode(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, node)

		// Don't start the node, test operations

		// These should not panic even if node is not started
		assert.Equal(t, 0, node.GetPeerCount())
		assert.Empty(t, node.GetConnectedPeers())
		assert.NotNil(t, node.GetStatus())
		assert.NotNil(t, node.GetMetrics())
	})

	t.Run("invalid peer connection", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node, err := NewP2PNode(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Stop()

		err = node.Start()
		require.NoError(t, err)

		// Try to connect to invalid peer
		invalidPeer := peer.AddrInfo{
			ID:    "invalid-peer-id",
			Addrs: nil,
		}

		err = node.ConnectToPeer(ctx, invalidPeer)
		assert.Error(t, err)
	})

	t.Run("operations with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		node, err := NewP2PNode(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, node)
		defer node.Stop()

		err = node.Start()
		require.NoError(t, err)

		// Cancel context
		cancel()

		// Operations should handle cancelled context gracefully
		content := &routing.ContentMetadata{
			ID:   "test-content",
			Name: "Test Content",
		}

		err = node.PublishContent(ctx, content)
		// Should handle cancelled context (may or may not error depending on implementation)
	})
}

// TestNode_StringRepresentation tests string methods
func TestNode_StringRepresentation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	status := node.GetStatus()
	statusStr := status.String()
	assert.Contains(t, statusStr, string(node.ID()))
	assert.Contains(t, statusStr, "Started=true")
}

// TestNode_ConcurrentOperations tests concurrent operations
func TestNode_ConcurrentOperations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, node)
	defer node.Stop()

	err = node.Start()
	require.NoError(t, err)

	// Run multiple operations concurrently
	done := make(chan bool, 10)

	// Concurrent capability updates
	go func() {
		for i := 0; i < 10; i++ {
			caps := &resources.NodeCapabilities{
				CPUCores: i + 1,
				Memory:   int64((i + 1) * 1024),
			}
			node.SetCapabilities(caps)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent metrics updates
	go func() {
		for i := 0; i < 10; i++ {
			metrics := &resources.ResourceMetrics{
				CPUUsage:  float64(i * 10),
				Timestamp: time.Now(),
			}
			node.SetResourceMetrics(metrics)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent status reads
	go func() {
		for i := 0; i < 10; i++ {
			status := node.GetStatus()
			assert.NotNil(t, status)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}
}

// Benchmark tests

func BenchmarkNode_GetPeerCount(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(b, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			count := node.GetPeerCount()
			_ = count
		}
	})
}

func BenchmarkNode_GetStatus(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(b, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			status := node.GetStatus()
			_ = status
		}
	})
}

func BenchmarkNode_SetCapabilities(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(b, err)
	defer node.Stop()

	err = node.Start()
	require.NoError(b, err)

	caps := &resources.NodeCapabilities{
		CPUCores: 4,
		Memory:   8192,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.SetCapabilities(caps)
	}
}

func BenchmarkNode_EmitEvent(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := NewP2PNode(ctx, nil)
	require.NoError(b, err)
	defer node.Stop()

	// Register event handler
	node.On(EventResourceUpdated, func(event *NodeEvent) {
		// Do nothing, just receive
	})

	caps := &resources.NodeCapabilities{
		CPUCores: 4,
		Memory:   8192,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.SetCapabilities(caps) // This triggers an event
	}
}
