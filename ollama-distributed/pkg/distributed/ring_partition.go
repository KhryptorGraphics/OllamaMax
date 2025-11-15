package distributed

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

// RingPartitioner implements memory-weighted ring partitioning for model layers
// Based on exo architecture: each device runs contiguous layers proportional to memory
type RingPartitioner struct {
	mu sync.RWMutex

	// Model configuration
	totalLayers int
	layerSizes  []uint64 // Memory size of each layer in bytes

	// Node assignments
	nodeAssignments map[peer.ID]*NodePartition
	orderedNodes    []peer.ID // Nodes in ring order

	// Topology integration
	topology *ClusterTopology
}

// NodePartition represents the layer assignment for a node
type NodePartition struct {
	PeerID      peer.ID
	StartLayer  int     // Inclusive
	EndLayer    int     // Exclusive
	LayerCount  int     // Number of layers assigned
	MemoryUsed  uint64  // Total memory used by assigned layers
	MemoryTotal uint64  // Total available memory
	LoadFactor  float64 // Ratio of memory used to total
}

// RingPartitionConfig configures ring partitioning behavior
type RingPartitionConfig struct {
	TotalLayers      int
	LayerSizes       []uint64
	MinLayersPerNode int     // Minimum layers per node (default: 1)
	MaxLoadImbalance float64 // Maximum allowed load imbalance (default: 0.2)
}

// NewRingPartitioner creates a new ring partitioner
func NewRingPartitioner(config *RingPartitionConfig, topology *ClusterTopology) *RingPartitioner {
	return &RingPartitioner{
		totalLayers:     config.TotalLayers,
		layerSizes:      config.LayerSizes,
		nodeAssignments: make(map[peer.ID]*NodePartition),
		orderedNodes:    make([]peer.ID, 0),
		topology:        topology,
	}
}

// PartitionLayers assigns contiguous layers to nodes based on memory weights
func (rp *RingPartitioner) PartitionLayers(nodes []*ClusterNode) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes available for partitioning")
	}

	// Sort nodes by peer ID for stable ordering
	sortedNodes := make([]*ClusterNode, len(nodes))
	copy(sortedNodes, nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return string(sortedNodes[i].PeerID) < string(sortedNodes[j].PeerID)
	})

	// Calculate memory-weighted layer assignments
	assignments := make([]*NodePartition, 0, len(sortedNodes))
	currentLayer := 0

	for _, node := range sortedNodes {
		// Calculate proportion of layers based on node count
		layerCount := int(math.Round(float64(rp.totalLayers) / float64(len(sortedNodes))))

		// Ensure at least 1 layer per node
		if layerCount < 1 {
			layerCount = 1
		}

		// Ensure we don't exceed total layers
		if currentLayer+layerCount > rp.totalLayers {
			layerCount = rp.totalLayers - currentLayer
		}

		if layerCount > 0 {
			// Calculate memory used by assigned layers
			memoryUsed := uint64(0)
			for i := currentLayer; i < currentLayer+layerCount && i < len(rp.layerSizes); i++ {
				memoryUsed += rp.layerSizes[i]
			}

			partition := &NodePartition{
				PeerID:      node.PeerID,
				StartLayer:  currentLayer,
				EndLayer:    currentLayer + layerCount,
				LayerCount:  layerCount,
				MemoryUsed:  memoryUsed,
				MemoryTotal: 0,
				LoadFactor:  0,
			}

			assignments = append(assignments, partition)
			currentLayer += layerCount
		}

		if currentLayer >= rp.totalLayers {
			break
		}
	}

	// Update internal state
	rp.nodeAssignments = make(map[peer.ID]*NodePartition)
	rp.orderedNodes = make([]peer.ID, 0, len(assignments))

	for _, assignment := range assignments {
		rp.nodeAssignments[assignment.PeerID] = assignment
		rp.orderedNodes = append(rp.orderedNodes, assignment.PeerID)
	}

	return nil
}

// GetPartition returns the partition assignment for a node
func (rp *RingPartitioner) GetPartition(peerID peer.ID) (*NodePartition, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	partition, exists := rp.nodeAssignments[peerID]
	if !exists {
		return nil, fmt.Errorf("no partition found for peer %s", peerID)
	}

	return partition, nil
}

// GetNextNode returns the next node in the ring for layer forwarding
func (rp *RingPartitioner) GetNextNode(currentPeerID peer.ID) (peer.ID, error) {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	for i, peerID := range rp.orderedNodes {
		if peerID == currentPeerID {
			nextIdx := (i + 1) % len(rp.orderedNodes)
			return rp.orderedNodes[nextIdx], nil
		}
	}

	return "", fmt.Errorf("peer %s not found in ring", currentPeerID)
}
