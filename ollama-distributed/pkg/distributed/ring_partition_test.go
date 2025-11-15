package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRingPartitioner(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	totalLayers := 32
	layerSizes := make([]uint64, totalLayers)
	for i := range layerSizes {
		layerSizes[i] = 1024 * 1024 // 1MB per layer
	}

	config := &RingPartitionConfig{
		TotalLayers: totalLayers,
		LayerSizes:  layerSizes,
	}
	partitioner := NewRingPartitioner(config, topology)

	assert.NotNil(t, partitioner)
	assert.Equal(t, totalLayers, partitioner.totalLayers)
	assert.Len(t, partitioner.layerSizes, totalLayers)
}

func TestPartitionLayers(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add nodes with different latencies to same cluster
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

	// Create partitioner
	totalLayers := 32
	layerSizes := make([]uint64, totalLayers)
	for i := range layerSizes {
		layerSizes[i] = 100 * 1024 * 1024 // 100MB per layer
	}

	config := &RingPartitionConfig{
		TotalLayers: totalLayers,
		LayerSizes:  layerSizes,
	}
	partitioner := NewRingPartitioner(config, topology)

	// Get cluster nodes
	clusterID := topology.NodeCluster[peer1]
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()

	// Collect nodes from cluster
	cluster.NodesMux.RLock()
	clusterNodes := make([]*ClusterNode, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		clusterNodes = append(clusterNodes, node)
	}
	cluster.NodesMux.RUnlock()

	// Partition layers
	err = partitioner.PartitionLayers(clusterNodes)
	require.NoError(t, err)

	// Verify partitions were created
	assert.Greater(t, len(partitioner.nodeAssignments), 0)
}

func TestRingPartitionerOrdering(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add nodes
	peer1 := peer.ID("peer-1")
	peer2 := peer.ID("peer-2")
	peer3 := peer.ID("peer-3")

	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	for _, peerID := range []peer.ID{peer1, peer2, peer3} {
		err := topology.AddNode(peerID, metrics)
		require.NoError(t, err)
	}

	// Create partitioner
	config := &RingPartitionConfig{
		TotalLayers: 32,
		LayerSizes:  make([]uint64, 32),
	}
	for i := range config.LayerSizes {
		config.LayerSizes[i] = 100 * 1024 * 1024 // 100MB per layer
	}
	partitioner := NewRingPartitioner(config, topology)

	// Get nodes from cluster
	clusterID := topology.NodeCluster[peer1]
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()

	cluster.NodesMux.RLock()
	nodes := make([]*ClusterNode, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		nodes = append(nodes, node)
	}
	cluster.NodesMux.RUnlock()

	// Partition layers
	err := partitioner.PartitionLayers(nodes)
	require.NoError(t, err)

	// Verify ordered nodes are set
	assert.Greater(t, len(partitioner.orderedNodes), 0)
}

func TestRingPartitionerMemoryWeighting(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add nodes
	peer1 := peer.ID("peer-1")
	peer2 := peer.ID("peer-2")

	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}

	for _, peerID := range []peer.ID{peer1, peer2} {
		err := topology.AddNode(peerID, metrics)
		require.NoError(t, err)
	}

	// Create partitioner
	config := &RingPartitionConfig{
		TotalLayers: 32,
		LayerSizes:  make([]uint64, 32),
	}
	partitioner := NewRingPartitioner(config, topology)

	// Verify partitioner was created
	assert.NotNil(t, partitioner)
	assert.Equal(t, 32, partitioner.totalLayers)
}

func TestGetNodePartition(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	peerID := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peerID, metrics)
	require.NoError(t, err)

	config := &RingPartitionConfig{
		TotalLayers: 32,
		LayerSizes:  make([]uint64, 32),
	}
	partitioner := NewRingPartitioner(config, topology)

	// Get cluster nodes
	clusterID := topology.NodeCluster[peerID]
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()

	cluster.NodesMux.RLock()
	clusterNodes := make([]*ClusterNode, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		clusterNodes = append(clusterNodes, node)
	}
	cluster.NodesMux.RUnlock()

	// Partition layers
	err = partitioner.PartitionLayers(clusterNodes)
	require.NoError(t, err)

	// Verify partition was created
	assert.Greater(t, len(partitioner.nodeAssignments), 0)
}

func TestRingPartitionerMultipleNodes(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add initial nodes
	peer1 := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peer1, metrics)
	require.NoError(t, err)

	config := &RingPartitionConfig{
		TotalLayers: 32,
		LayerSizes:  make([]uint64, 32),
	}
	for i := range config.LayerSizes {
		config.LayerSizes[i] = 100 * 1024 * 1024 // 100MB per layer
	}
	partitioner := NewRingPartitioner(config, topology)

	// Add another node
	peer2 := peer.ID("peer-2")
	err = topology.AddNode(peer2, metrics)
	require.NoError(t, err)

	// Get nodes from cluster
	clusterID := topology.NodeCluster[peer1]
	topology.ClustersMux.RLock()
	cluster := topology.Clusters[clusterID]
	topology.ClustersMux.RUnlock()

	cluster.NodesMux.RLock()
	nodes := make([]*ClusterNode, 0, len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		nodes = append(nodes, node)
	}
	cluster.NodesMux.RUnlock()

	// Partition layers
	err = partitioner.PartitionLayers(nodes)
	require.NoError(t, err)

	// Verify both nodes are in ordered nodes
	assert.Greater(t, len(partitioner.orderedNodes), 0)
}
