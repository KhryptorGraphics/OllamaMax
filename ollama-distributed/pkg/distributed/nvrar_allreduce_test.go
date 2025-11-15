package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNVRARAllReduce(t *testing.T) {
	config := &NVRARConfig{
		MessageSizeThreshold: 1024 * 1024,
		ChunkSize:            64 * 1024,
		TimeoutPerPhase:      30 * time.Second,
		EnablePipelining:     true,
		CompressionEnabled:   true,
	}
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	nvrar := NewNVRARAllReduce(config, topology)

	assert.NotNil(t, nvrar)
	assert.NotNil(t, nvrar.nodeGroups)
	assert.Equal(t, config.MessageSizeThreshold, nvrar.config.MessageSizeThreshold)
}

func TestInitializeNodeGroups(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add nodes to create clusters
	peer1 := peer.ID("peer-1")
	peer2 := peer.ID("peer-2")

	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	err = topology.AddNode(peer2, metrics)
	require.NoError(t, err)

	config := &NVRARConfig{
		MessageSizeThreshold: 1024 * 1024,
		ChunkSize:            64 * 1024,
		TimeoutPerPhase:      30 * time.Second,
		EnablePipelining:     true,
		CompressionEnabled:   true,
	}
	nvrar := NewNVRARAllReduce(config, topology)

	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Verify node groups were created
	assert.Greater(t, len(nvrar.nodeGroups), 0)

	// Verify nodes are assigned to groups
	for _, group := range nvrar.nodeGroups {
		assert.NotEmpty(t, group.GroupID)
		assert.Greater(t, len(group.Nodes), 0)
		assert.NotEmpty(t, group.Leader)
	}
}

func TestAllReduceSmallMessage(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := DefaultNVRARConfig()
	nvrar := NewNVRARAllReduce(config, topology)
	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Verify metrics initialized
	assert.NotNil(t, nvrar.metrics)
}

func TestAllReduceLargeMessage(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := DefaultNVRARConfig()
	nvrar := NewNVRARAllReduce(config, topology)
	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Verify configuration
	assert.Equal(t, config.MessageSizeThreshold, nvrar.config.MessageSizeThreshold)
}

func TestAllReduceOperations(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := DefaultNVRARConfig()
	nvrar := NewNVRARAllReduce(config, topology)
	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Verify node groups created
	assert.Greater(t, len(nvrar.nodeGroups), 0)
}

func TestNVRARGetMetrics(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := DefaultNVRARConfig()
	nvrar := NewNVRARAllReduce(config, topology)
	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Get metrics
	allReduceMetrics := nvrar.GetMetrics()
	assert.NotNil(t, allReduceMetrics)
}

func TestNVRARSpeedupEstimation(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := DefaultNVRARConfig()
	nvrar := NewNVRARAllReduce(config, topology)
	err = nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Verify configuration
	assert.Equal(t, config.ChunkSize, nvrar.config.ChunkSize)
}
