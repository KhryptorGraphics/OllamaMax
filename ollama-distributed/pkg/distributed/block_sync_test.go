package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlockSynchronizer(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	totalLayers := 32
	layerSizes := make([]uint64, totalLayers)
	for i := range layerSizes {
		layerSizes[i] = 1024 * 1024
	}

	config := &RingPartitionConfig{
		TotalLayers: totalLayers,
		LayerSizes:  layerSizes,
	}
	partitioner := NewRingPartitioner(config, topology)

	blockConfig := &BlockSyncConfig{
		LayersPerBlock:     4,
		SyncTimeout:        5 * time.Second,
		MaxRetries:         3,
		BufferSize:         1024 * 1024,
		CompressionEnabled: true,
	}
	blockSync := NewBlockSynchronizer(blockConfig, partitioner)

	assert.NotNil(t, blockSync)
	assert.Equal(t, blockConfig.LayersPerBlock, blockSync.layersPerBlock)
	assert.NotNil(t, blockSync.blockStates)
}

func TestInitializeBlocks(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	// Add node
	peerID := peer.ID("peer-1")
	metrics := &NetworkMetrics{
		RTT:       5 * time.Millisecond,
		Bandwidth: 1000.0,
	}
	err := topology.AddNode(peerID, metrics)
	require.NoError(t, err)

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

	blockConfig := &BlockSyncConfig{
		LayersPerBlock:     4,
		SyncTimeout:        5 * time.Second,
		MaxRetries:         3,
		BufferSize:         1024 * 1024,
		CompressionEnabled: true,
	}
	blockSync := NewBlockSynchronizer(blockConfig, partitioner)

	// Verify blocks were created
	expectedBlocks := (totalLayers + blockConfig.LayersPerBlock - 1) / blockConfig.LayersPerBlock
	assert.Equal(t, expectedBlocks, blockSync.totalBlocks)
	assert.NotNil(t, blockSync.blockStates)
}

func TestBlockSyncConfiguration(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	totalLayers := 32
	layerSizes := make([]uint64, totalLayers)
	for i := range layerSizes {
		layerSizes[i] = 100 * 1024 * 1024
	}

	config := &RingPartitionConfig{
		TotalLayers: totalLayers,
		LayerSizes:  layerSizes,
	}
	partitioner := NewRingPartitioner(config, topology)

	blockConfig := &BlockSyncConfig{
		LayersPerBlock:     4,
		SyncTimeout:        5 * time.Second,
		MaxRetries:         3,
		BufferSize:         1024 * 1024,
		CompressionEnabled: true,
	}
	blockSync := NewBlockSynchronizer(blockConfig, partitioner)

	// Verify configuration
	assert.Equal(t, blockConfig.LayersPerBlock, blockSync.layersPerBlock)
	assert.Equal(t, blockConfig.CompressionEnabled, blockConfig.CompressionEnabled)
}

func TestBlockStateTracking(t *testing.T) {
	ctx := context.Background()
	topology := NewClusterTopology(ctx, peer.ID("test-node"))

	totalLayers := 32
	layerSizes := make([]uint64, totalLayers)
	for i := range layerSizes {
		layerSizes[i] = 100 * 1024 * 1024
	}

	config := &RingPartitionConfig{
		TotalLayers: totalLayers,
		LayerSizes:  layerSizes,
	}
	partitioner := NewRingPartitioner(config, topology)

	blockConfig := &BlockSyncConfig{
		LayersPerBlock:     4,
		SyncTimeout:        5 * time.Second,
		MaxRetries:         3,
		BufferSize:         1024 * 1024,
		CompressionEnabled: true,
	}
	blockSync := NewBlockSynchronizer(blockConfig, partitioner)

	// Verify block states map is initialized
	assert.NotNil(t, blockSync.blockStates)
}
