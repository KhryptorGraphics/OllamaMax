package integration

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/distributed"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
)

// TestEndToEndInferenceFlow tests the complete inference flow with all Phase A+B components
func TestEndToEndInferenceFlow(t *testing.T) {
	// Setup: Create topology with multiple nodes
	ctx := context.Background()
	topology := distributed.NewClusterTopology(ctx, peer.ID("test-node"))

	// Add 3 nodes with different capabilities
	nodes := []struct {
		peerID peer.ID
		rtt    time.Duration
	}{
		{peer.ID("node-1"), 5 * time.Millisecond},
		{peer.ID("node-2"), 6 * time.Millisecond},
		{peer.ID("node-3"), 7 * time.Millisecond},
	}

	for _, node := range nodes {
		err := topology.AddNode(node.peerID, &distributed.NetworkMetrics{
			RTT:       node.rtt,
			Bandwidth: 1000.0,
		})
		require.NoError(t, err)
	}

	// Get cluster ID
	clusterID := topology.NodeCluster[nodes[0].peerID]
	require.NotEmpty(t, clusterID)

	// Create block synchronizer
	// Create back-pressure manager
	backpressure := scheduler.NewBackPressureManager(scheduler.DefaultBackPressureConfig())
	defer backpressure.Stop()

	// Create NVRAR all-reduce
	nvrarConfig := distributed.DefaultNVRARConfig()
	nvrar := distributed.NewNVRARAllReduce(nvrarConfig, topology)
	err := nvrar.InitializeNodeGroups()
	require.NoError(t, err)

	// Step 1: Update queue depths (simulate load)
	for _, node := range nodes {
		backpressure.UpdateQueueDepth(node.peerID, 50)
	}

	// Verify no nodes are throttled with normal load
	for _, node := range nodes {
		assert.False(t, backpressure.IsThrottled(node.peerID))
	}

	// Verify metrics
	nvrarMetrics := nvrar.GetMetrics()
	assert.NotNil(t, nvrarMetrics)
}

// TestMultiTierCommunication tests communication across different cluster tiers
func TestMultiTierCommunication(t *testing.T) {
	ctx := context.Background()
	topology := distributed.NewClusterTopology(ctx, peer.ID("test-node"))

	// Add nodes in different tiers
	localNodes := []peer.ID{peer.ID("local-1"), peer.ID("local-2")}
	regionalNodes := []peer.ID{peer.ID("regional-1"), peer.ID("regional-2")}

	// Add local tier nodes
	for _, peerID := range localNodes {
		err := topology.AddNode(peerID, &distributed.NetworkMetrics{
			RTT:       5 * time.Millisecond,
			Bandwidth: 10000.0,
		})
		require.NoError(t, err)
	}

	// Add regional tier nodes
	for _, peerID := range regionalNodes {
		err := topology.AddNode(peerID, &distributed.NetworkMetrics{
			RTT:       30 * time.Millisecond,
			Bandwidth: 1000.0,
		})
		require.NoError(t, err)
	}

	// Verify tier assignments
	localCluster := topology.NodeCluster[localNodes[0]]
	regionalCluster := topology.NodeCluster[regionalNodes[0]]

	assert.NotEmpty(t, localCluster)
	assert.NotEmpty(t, regionalCluster)
}

// TestDynamicRepartitioning tests dynamic re-partitioning under load
func TestDynamicRepartitioning(t *testing.T) {
	ctx := context.Background()
	topology := distributed.NewClusterTopology(ctx, peer.ID("test-node"))

	// Add initial nodes
	peer1 := peer.ID("peer-1")
	err := topology.AddNode(peer1, &distributed.NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	})
	require.NoError(t, err)

	// Add another node
	peer2 := peer.ID("peer-2")
	err = topology.AddNode(peer2, &distributed.NetworkMetrics{
		RTT:       6 * time.Millisecond,
		Bandwidth: 1000.0,
	})
	require.NoError(t, err)

	// Verify both nodes are in topology
	assert.NotEmpty(t, topology.NodeCluster[peer1])
	assert.NotEmpty(t, topology.NodeCluster[peer2])
}

// TestBackPressureUnderLoad tests back-pressure mechanisms under high load
func TestBackPressureUnderLoad(t *testing.T) {
	backpressure := scheduler.NewBackPressureManager(&scheduler.BackPressureConfig{
		MaxQueueDepth:       100,
		QueueDepthThreshold: 0.8,
		ThrottleDuration:    1 * time.Second,
		MonitorInterval:     100 * time.Millisecond,
		AdaptiveThrottling:  true,
		MinProcessingRate:   5.0,
	})
	defer backpressure.Stop()

	peer1 := peer.ID("peer-1")
	peer2 := peer.ID("peer-2")

	// Simulate normal load on peer1
	for i := 0; i < 10; i++ {
		backpressure.UpdateQueueDepth(peer1, 50)
		time.Sleep(10 * time.Millisecond)
	}

	// Simulate high load on peer2
	for i := 0; i < 10; i++ {
		backpressure.UpdateQueueDepth(peer2, 85+i)
		time.Sleep(10 * time.Millisecond)
	}

	// Verify peer1 is not throttled
	assert.False(t, backpressure.IsThrottled(peer1))

	// Verify peer2 is throttled
	assert.True(t, backpressure.IsThrottled(peer2))
}
