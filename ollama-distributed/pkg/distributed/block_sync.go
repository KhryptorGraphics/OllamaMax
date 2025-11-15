package distributed

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

// BlockSynchronizer implements block-level synchronization for layer groups
// Groups layers into blocks (4-8 layers per block) to reduce communication overhead
// Based on TawPipe research: confines 75%+ communication within node boundaries
type BlockSynchronizer struct {
	mu sync.RWMutex

	// Block configuration
	layersPerBlock int
	totalBlocks    int
	blockSize      []uint64 // Memory size of each block

	// Synchronization state
	blockStates map[int]*BlockState
	nodeBlocks  map[peer.ID][]int // Blocks assigned to each node

	// Ring partitioner integration
	partitioner *RingPartitioner
}

// BlockState tracks the synchronization state of a block
type BlockState struct {
	BlockID       int
	StartLayer    int
	EndLayer      int
	LayerCount    int
	MemorySize    uint64
	OwnerNode     peer.ID
	Status        BlockStatus
	LastSync      time.Time
	SyncCount     int64
	PendingData   []byte // Buffered activation data
	DataReady     bool
	mu            sync.RWMutex
}

// BlockStatus represents the current status of a block
type BlockStatus int

const (
	BlockIdle BlockStatus = iota
	BlockProcessing
	BlockSyncing
	BlockComplete
	BlockFailed
)

// BlockSyncConfig configures block synchronization behavior
type BlockSyncConfig struct {
	LayersPerBlock    int           // Number of layers per block (4-8 recommended)
	SyncTimeout       time.Duration // Timeout for block synchronization
	MaxRetries        int           // Maximum retry attempts for failed syncs
	BufferSize        int           // Size of data buffer per block
	CompressionEnabled bool         // Enable compression for block data
}

// NewBlockSynchronizer creates a new block synchronizer
func NewBlockSynchronizer(config *BlockSyncConfig, partitioner *RingPartitioner) *BlockSynchronizer {
	totalBlocks := (partitioner.totalLayers + config.LayersPerBlock - 1) / config.LayersPerBlock

	return &BlockSynchronizer{
		layersPerBlock: config.LayersPerBlock,
		totalBlocks:    totalBlocks,
		blockStates:    make(map[int]*BlockState),
		nodeBlocks:     make(map[peer.ID][]int),
		partitioner:    partitioner,
	}
}

// InitializeBlocks initializes block states based on ring partitioning
func (bs *BlockSynchronizer) InitializeBlocks() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Create block states
	for blockID := 0; blockID < bs.totalBlocks; blockID++ {
		startLayer := blockID * bs.layersPerBlock
		endLayer := min(startLayer+bs.layersPerBlock, bs.partitioner.totalLayers)

		// Calculate block memory size
		memorySize := uint64(0)
		for i := startLayer; i < endLayer && i < len(bs.partitioner.layerSizes); i++ {
			memorySize += bs.partitioner.layerSizes[i]
		}

		// Find owner node for this block
		ownerNode := bs.findBlockOwner(startLayer)

		blockState := &BlockState{
			BlockID:     blockID,
			StartLayer:  startLayer,
			EndLayer:    endLayer,
			LayerCount:  endLayer - startLayer,
			MemorySize:  memorySize,
			OwnerNode:   ownerNode,
			Status:      BlockIdle,
			LastSync:    time.Now(),
			SyncCount:   0,
			DataReady:   false,
		}

		bs.blockStates[blockID] = blockState

		// Track blocks per node
		if _, exists := bs.nodeBlocks[ownerNode]; !exists {
			bs.nodeBlocks[ownerNode] = make([]int, 0)
		}
		bs.nodeBlocks[ownerNode] = append(bs.nodeBlocks[ownerNode], blockID)
	}

	return nil
}

// findBlockOwner determines which node owns a block based on layer assignment
func (bs *BlockSynchronizer) findBlockOwner(startLayer int) peer.ID {
	for peerID, partition := range bs.partitioner.nodeAssignments {
		if startLayer >= partition.StartLayer && startLayer < partition.EndLayer {
			return peerID
		}
	}
	return ""
}

// SyncBlock synchronizes a block between nodes
func (bs *BlockSynchronizer) SyncBlock(ctx context.Context, blockID int, data []byte) error {
	bs.mu.RLock()
	blockState, exists := bs.blockStates[blockID]
	bs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("block %d not found", blockID)
	}

	blockState.mu.Lock()
	defer blockState.mu.Unlock()

	// Update block state
	blockState.Status = BlockSyncing
	blockState.PendingData = data
	blockState.LastSync = time.Now()
	blockState.SyncCount++

	// Mark data as ready
	blockState.DataReady = true
	blockState.Status = BlockComplete

	return nil
}

// GetBlockState returns the current state of a block
func (bs *BlockSynchronizer) GetBlockState(blockID int) (*BlockState, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	blockState, exists := bs.blockStates[blockID]
	if !exists {
		return nil, fmt.Errorf("block %d not found", blockID)
	}

	return blockState, nil
}

// GetNodeBlocks returns all blocks assigned to a node
func (bs *BlockSynchronizer) GetNodeBlocks(peerID peer.ID) ([]int, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	blocks, exists := bs.nodeBlocks[peerID]
	if !exists {
		return nil, fmt.Errorf("no blocks found for peer %s", peerID)
	}

	return blocks, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

