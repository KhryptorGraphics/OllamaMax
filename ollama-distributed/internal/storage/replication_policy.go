package storage

import (
	"fmt"
	"sort"
	"time"
)

// ReplicationStrategy defines the interface for replication strategies
type ReplicationStrategy interface {
	GetName() string
	SelectTargetNodes(sourceNode string, nodes []*StorageNode, policy *ReplicationPolicy) ([]*StorageNode, error)
	GetReplicationOrder(sourceNode string, targetNodes []*StorageNode) []*StorageNode
	ShouldReplicate(key string, metadata *ObjectMetadata, policy *ReplicationPolicy) bool
	GetConsistencyLevel() string
}

// EagerReplicationStrategy implements eager replication
type EagerReplicationStrategy struct {
	config *ReplicationConfig
}

// LazyReplicationStrategy implements lazy replication
type LazyReplicationStrategy struct {
	config *ReplicationConfig
}

// GeographicReplicationStrategy implements geographic replication
type GeographicReplicationStrategy struct {
	config         *ReplicationConfig
	regionPriority map[string]int
}

// EagerReplicationStrategy implementation

// GetName returns the strategy name
func (ers *EagerReplicationStrategy) GetName() string {
	return "eager"
}

// SelectTargetNodes selects target nodes for eager replication
func (ers *EagerReplicationStrategy) SelectTargetNodes(sourceNode string, nodes []*StorageNode, policy *ReplicationPolicy) ([]*StorageNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Sort nodes by health and capacity
	sortedNodes := make([]*StorageNode, len(nodes))
	copy(sortedNodes, nodes)

	sort.Slice(sortedNodes, func(i, j int) bool {
		scoreI := ers.calculateNodeScore(sortedNodes[i])
		scoreJ := ers.calculateNodeScore(sortedNodes[j])
		return scoreI > scoreJ
	})

	// Select top nodes up to max replicas
	count := policy.MaxReplicas
	if count > len(sortedNodes) {
		count = len(sortedNodes)
	}

	return sortedNodes[:count], nil
}

// GetReplicationOrder returns the order for replication
func (ers *EagerReplicationStrategy) GetReplicationOrder(sourceNode string, targetNodes []*StorageNode) []*StorageNode {
	// For eager replication, replicate to all nodes in parallel
	return targetNodes
}

// ShouldReplicate determines if an object should be replicated
func (ers *EagerReplicationStrategy) ShouldReplicate(key string, metadata *ObjectMetadata, policy *ReplicationPolicy) bool {
	// Always replicate in eager strategy
	return true
}

// GetConsistencyLevel returns the consistency level
func (ers *EagerReplicationStrategy) GetConsistencyLevel() string {
	return "strong"
}

// calculateNodeScore calculates a score for node selection
func (ers *EagerReplicationStrategy) calculateNodeScore(node *StorageNode) float64 {
	score := 0.0

	// Health score
	if node.Health.Status == "healthy" {
		score += 100.0
	} else if node.Health.Status == "degraded" {
		score += 50.0
	}

	// Capacity score
	if node.Capacity != nil {
		availablePercent := float64(node.Capacity.AvailableBytes) / float64(node.Capacity.TotalBytes)
		score += availablePercent * 50.0
	}

	// Load factor score (lower is better)
	score -= node.LoadFactor * 10.0

	return score
}

// LazyReplicationStrategy implementation

// GetName returns the strategy name
func (lrs *LazyReplicationStrategy) GetName() string {
	return "lazy"
}

// SelectTargetNodes selects target nodes for lazy replication
func (lrs *LazyReplicationStrategy) SelectTargetNodes(sourceNode string, nodes []*StorageNode, policy *ReplicationPolicy) ([]*StorageNode, error) {
	// For lazy replication, select fewer nodes initially
	count := policy.MinReplicas / 2
	if count == 0 {
		count = 1
	}

	selectedNodes := make([]*StorageNode, 0, count)
	for i, node := range nodes {
		if i >= count {
			break
		}
		selectedNodes = append(selectedNodes, node)
	}

	return selectedNodes, nil
}

// GetReplicationOrder returns the order for replication
func (lrs *LazyReplicationStrategy) GetReplicationOrder(sourceNode string, targetNodes []*StorageNode) []*StorageNode {
	return targetNodes
}

// ShouldReplicate determines if an object should be replicated
func (lrs *LazyReplicationStrategy) ShouldReplicate(key string, metadata *ObjectMetadata, policy *ReplicationPolicy) bool {
	// Replicate based on access patterns or other criteria
	return time.Since(metadata.CreatedAt) > time.Hour
}

// GetConsistencyLevel returns the consistency level
func (lrs *LazyReplicationStrategy) GetConsistencyLevel() string {
	return "eventual"
}

// GeographicReplicationStrategy implementation

// GetName returns the strategy name
func (grs *GeographicReplicationStrategy) GetName() string {
	return "geographic"
}

// SelectTargetNodes selects target nodes for geographic replication
func (grs *GeographicReplicationStrategy) SelectTargetNodes(sourceNode string, nodes []*StorageNode, policy *ReplicationPolicy) ([]*StorageNode, error) {
	// Group nodes by region
	regionNodes := make(map[string][]*StorageNode)
	for _, node := range nodes {
		region := node.Region
		if region == "" {
			region = "default"
		}
		regionNodes[region] = append(regionNodes[region], node)
	}

	// Select nodes from different regions
	selectedNodes := make([]*StorageNode, 0, policy.MaxReplicas)
	regionsUsed := make(map[string]bool)

	// First pass: select one node from each region
	for region, regionNodeList := range regionNodes {
		if len(selectedNodes) >= policy.MaxReplicas {
			break
		}
		if len(regionNodeList) > 0 {
			// Select best node from this region
			bestNode := regionNodeList[0]
			for _, node := range regionNodeList {
				if grs.calculateNodeScore(node) > grs.calculateNodeScore(bestNode) {
					bestNode = node
				}
			}
			selectedNodes = append(selectedNodes, bestNode)
			regionsUsed[region] = true
		}
	}

	return selectedNodes, nil
}

// GetReplicationOrder returns the order for replication
func (grs *GeographicReplicationStrategy) GetReplicationOrder(sourceNode string, targetNodes []*StorageNode) []*StorageNode {
	return targetNodes
}

// ShouldReplicate determines if an object should be replicated
func (grs *GeographicReplicationStrategy) ShouldReplicate(key string, metadata *ObjectMetadata, policy *ReplicationPolicy) bool {
	return true
}

// GetConsistencyLevel returns the consistency level
func (grs *GeographicReplicationStrategy) GetConsistencyLevel() string {
	return "eventual"
}

// calculateNodeScore calculates a score for node selection
func (grs *GeographicReplicationStrategy) calculateNodeScore(node *StorageNode) float64 {
	score := 0.0

	// Health score
	if node.Health.Status == "healthy" {
		score += 100.0
	} else if node.Health.Status == "degraded" {
		score += 50.0
	}

	// Regional priority score
	if priority, exists := grs.regionPriority[node.Region]; exists {
		score += float64(priority) * 10.0
	}

	return score
}

// Policy helper functions

// NewEagerReplicationStrategy creates a new eager replication strategy
func NewEagerReplicationStrategy(config *ReplicationConfig) *EagerReplicationStrategy {
	return &EagerReplicationStrategy{
		config: config,
	}
}

// NewLazyReplicationStrategy creates a new lazy replication strategy
func NewLazyReplicationStrategy(config *ReplicationConfig) *LazyReplicationStrategy {
	return &LazyReplicationStrategy{
		config: config,
	}
}

// NewGeographicReplicationStrategy creates a new geographic replication strategy
func NewGeographicReplicationStrategy(config *ReplicationConfig) *GeographicReplicationStrategy {
	return &GeographicReplicationStrategy{
		config:         config,
		regionPriority: make(map[string]int),
	}
}

// SetRegionPriority sets the priority for a region in geographic replication
func (grs *GeographicReplicationStrategy) SetRegionPriority(region string, priority int) {
	grs.regionPriority[region] = priority
}

// GetRegionPriority gets the priority for a region
func (grs *GeographicReplicationStrategy) GetRegionPriority(region string) int {
	if priority, exists := grs.regionPriority[region]; exists {
		return priority
	}
	return 0
}

// ValidateReplicationPolicy validates a replication policy
func ValidateReplicationPolicy(policy *ReplicationPolicy) error {
	if policy == nil {
		return fmt.Errorf("replication policy cannot be nil")
	}

	if policy.MinReplicas < 1 {
		return fmt.Errorf("minimum replicas must be at least 1")
	}

	if policy.MaxReplicas < policy.MinReplicas {
		return fmt.Errorf("maximum replicas must be greater than or equal to minimum replicas")
	}

	// Note: ReplicationFactor field doesn't exist in the interface.go definition
	// Using MaxReplicas for validation instead

	validConsistencyLevels := map[string]bool{
		"strong":   true,
		"eventual": true,
		"weak":     true,
	}

	if !validConsistencyLevels[policy.ConsistencyLevel] {
		return fmt.Errorf("invalid consistency level: %s", policy.ConsistencyLevel)
	}

	return nil
}

// GetDefaultReplicationPolicy returns a default replication policy
func GetDefaultReplicationPolicy() *ReplicationPolicy {
	return &ReplicationPolicy{
		MinReplicas:      2,
		MaxReplicas:      5,
		PreferredNodes:   []string{},
		ExcludedNodes:    []string{},
		ConsistencyLevel: "eventual",
		Strategy:         "eager",
		Priority:         1,
		Constraints:      make(map[string]interface{}),
	}
}
