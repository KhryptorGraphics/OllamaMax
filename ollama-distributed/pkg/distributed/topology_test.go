package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClusterTopology(t *testing.T) {
	ctx := context.Background()
	localNodeID := peer.ID("test-node")
	topology := NewClusterTopology(ctx, localNodeID)

	assert.NotNil(t, topology)
	assert.NotNil(t, topology.Clusters)
	assert.NotNil(t, topology.NodeCluster)
	assert.Equal(t, 10*time.Millisecond, topology.LocalLatencyThreshold)
	assert.Equal(t, 50*time.Millisecond, topology.RegionalLatencyThreshold)
}

func TestDetermineTier(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	tests := []struct {
		name     string
		rtt      time.Duration
		expected ClusterTier
	}{
		{"Local tier", 5 * time.Millisecond, TierLocal},
		{"Regional tier", 30 * time.Millisecond, TierRegional},
		{"Global tier", 100 * time.Millisecond, TierGlobal},
		{"Boundary local", 10 * time.Millisecond, TierRegional},
		{"Boundary regional", 50 * time.Millisecond, TierGlobal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := topology.determineTier(tt.rtt)
			assert.Equal(t, tt.expected, tier)
		})
	}
}

func TestAddNode(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peerID := peer.ID("test-peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peerID, metrics)
	require.NoError(t, err)

	// Verify node was added
	clusterID, exists := topology.NodeCluster[peerID]
	assert.True(t, exists)
	assert.NotEmpty(t, clusterID)

	// Verify cluster was created
	topology.ClustersMux.RLock()
	cluster, exists := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()
	assert.True(t, exists)
	assert.Equal(t, TierLocal, cluster.Tier)
	assert.Len(t, cluster.Nodes, 1)
}

func TestAddMultipleNodesToSameCluster(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add first node
	peer1 := peer.ID("test-peer-1")
	metrics1 := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peer1, metrics1)
	require.NoError(t, err)

	// Add second node with similar latency (within 20% threshold)
	peer2 := peer.ID("test-peer-2")
	metrics2 := &NetworkMetrics{
		RTT:       5500 * time.Microsecond, // 5.5ms, which is 10% difference from 5ms
		Bandwidth: 1000.0,
	}

	err = topology.AddNode(peer2, metrics2)
	require.NoError(t, err)

	// Verify both nodes are in the same cluster
	cluster1ID := topology.NodeCluster[peer1]
	cluster2ID := topology.NodeCluster[peer2]
	assert.Equal(t, cluster1ID, cluster2ID)

	topology.ClustersMux.RLock()
	cluster := topology.Clusters[cluster1ID]
	topology.ClustersMux.RUnlock()
	assert.Len(t, cluster.Nodes, 2)
}

func TestNodeTierAssignment(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add node with local latency
	peerLocal := peer.ID("test-peer-local")
	metricsLocal := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peerLocal, metricsLocal)
	require.NoError(t, err)

	// Add node with regional latency
	peerRegional := peer.ID("test-peer-regional")
	metricsRegional := &NetworkMetrics{
		RTT:       30 * time.Millisecond,
		Bandwidth: 500.0,
	}

	err = topology.AddNode(peerRegional, metricsRegional)
	require.NoError(t, err)

	// Verify they are in different clusters
	clusterLocal := topology.NodeCluster[peerLocal]
	clusterRegional := topology.NodeCluster[peerRegional]
	assert.NotEqual(t, clusterLocal, clusterRegional)
}

func TestUpdateNodeMetrics(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peerID := peer.ID("test-peer-1")
	initialMetrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peerID, initialMetrics)
	require.NoError(t, err)

	// Update metrics
	newMetrics := &NetworkMetrics{
		RTT:       10 * time.Millisecond,
		Bandwidth: 2000.0,
	}

	err = topology.UpdateNodeMetrics(peerID, newMetrics)
	require.NoError(t, err)

	// Verify metrics were updated
	clusterID := topology.NodeCluster[peerID]
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()

	var node *ClusterNode
	cluster.NodesMux.RLock()
	for _, n := range cluster.Nodes {
		if n.PeerID == peerID {
			node = n
			break
		}
	}
	cluster.NodesMux.RUnlock()

	require.NotNil(t, node)
	// Verify that metrics were updated
	assert.Equal(t, node.Metrics.RTT, newMetrics.RTT)
	assert.Equal(t, node.Metrics.Bandwidth, newMetrics.Bandwidth)
	assert.Greater(t, node.Metrics.MeasureCount, 0)
}

func TestClusterMetrics(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add multiple nodes to same cluster
	peer1 := peer.ID("test-peer-1")
	peer2 := peer.ID("test-peer-2")

	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	err = topology.AddNode(peer2, metrics)
	require.NoError(t, err)

	// Verify both nodes are in same cluster
	clusterID1 := topology.NodeCluster[peer1]
	clusterID2 := topology.NodeCluster[peer2]
	assert.Equal(t, clusterID1, clusterID2)

	// Verify cluster has both nodes
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID1]
	topology.ClustersMux.RUnlock()

	assert.Len(t, cluster.Nodes, 2)
}
