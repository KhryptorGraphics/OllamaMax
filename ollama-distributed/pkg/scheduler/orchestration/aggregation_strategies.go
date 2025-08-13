package orchestration

import (
	"fmt"
	"time"
)

// ConcatAggregationStrategy concatenates partial results
type ConcatAggregationStrategy struct {
	name string
}

func (cas *ConcatAggregationStrategy) GetName() string {
	return "concat"
}

func (cas *ConcatAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Concatenate all partial results
	results := make([]interface{}, 0)
	for _, partial := range context.PartialResults {
		if partial.Error == "" {
			results = append(results, partial.Data)
		}
	}

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: cas.GetName(),
		Data:     results,
		Metadata: map[string]interface{}{
			"concatenated_count": len(results),
			"total_partitions":   len(context.PartialResults),
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// AverageAggregationStrategy averages partial results
type AverageAggregationStrategy struct {
	name string
}

func (aas *AverageAggregationStrategy) GetName() string {
	return "average"
}

func (aas *AverageAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Average numeric results
	var total float64
	var count int

	for _, partial := range context.PartialResults {
		if partial.Error == "" {
			if value, ok := partial.Data.(float64); ok {
				total += value
				count++
			}
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("no valid numeric results to average")
	}

	average := total / float64(count)

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: aas.GetName(),
		Data:     average,
		Metadata: map[string]interface{}{
			"total_sum":        total,
			"count":            count,
			"average":          average,
			"total_partitions": len(context.PartialResults),
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// WeightedAggregationStrategy performs weighted aggregation
type WeightedAggregationStrategy struct {
	name string
}

func (was *WeightedAggregationStrategy) GetName() string {
	return "weighted"
}

func (was *WeightedAggregationStrategy) Aggregate(context *AggregationContext) (*AggregatedResponse, error) {
	start := time.Now()

	// Perform weighted aggregation
	var weightedSum float64
	var totalWeight float64

	for _, partial := range context.PartialResults {
		if partial.Error == "" {
			if value, ok := partial.Data.(float64); ok {
				// Get weight from metadata, default to 1.0
				weight := 1.0
				if w, exists := partial.Metadata["weight"]; exists {
					if weightVal, ok := w.(float64); ok {
						weight = weightVal
					}
				}

				weightedSum += value * weight
				totalWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("no valid weighted results to aggregate")
	}

	weightedAverage := weightedSum / totalWeight

	return &AggregatedResponse{
		TaskID:   context.TaskID,
		Strategy: was.GetName(),
		Data:     weightedAverage,
		Metadata: map[string]interface{}{
			"weighted_sum":     weightedSum,
			"total_weight":     totalWeight,
			"weighted_average": weightedAverage,
			"total_partitions": len(context.PartialResults),
		},
		Latency:   time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// RoundRobinPartitioningStrategy implements round-robin partitioning
type RoundRobinPartitioningStrategy struct {
	name    string
	counter int
}

func (rrps *RoundRobinPartitioningStrategy) GetName() string {
	return "round_robin"
}

func (rrps *RoundRobinPartitioningStrategy) Partition(request *OrchestrationRequest) (*PartitionPlan, error) {
	// Simple round-robin partitioning
	nodeCount := 3 // Mock node count
	partitions := make([]*TaskPartition, nodeCount)

	for i := 0; i < nodeCount; i++ {
		partitions[i] = &TaskPartition{
			ID:           fmt.Sprintf("partition_%d", i),
			NodeID:       fmt.Sprintf("node_%d", i),
			Type:         "round_robin",
			Data:         fmt.Sprintf("partition_data_%d", i),
			Dependencies: []string{},
			Metadata: map[string]interface{}{
				"partition_index":  i,
				"total_partitions": nodeCount,
			},
		}
	}

	return &PartitionPlan{
		ID:         fmt.Sprintf("plan_%d", time.Now().UnixNano()),
		Strategy:   rrps.GetName(),
		Partitions: partitions,
		Metadata: map[string]interface{}{
			"strategy":   "round_robin",
			"node_count": nodeCount,
		},
		CreatedAt: time.Now(),
	}, nil
}

// LoadBasedPartitioningStrategy implements load-based partitioning
type LoadBasedPartitioningStrategy struct {
	name string
}

func (lbps *LoadBasedPartitioningStrategy) GetName() string {
	return "load_based"
}

func (lbps *LoadBasedPartitioningStrategy) Partition(request *OrchestrationRequest) (*PartitionPlan, error) {
	// Mock load-based partitioning
	nodeLoads := map[string]float64{
		"node_0": 0.3,
		"node_1": 0.7,
		"node_2": 0.5,
	}

	partitions := make([]*TaskPartition, 0)
	partitionIndex := 0

	for nodeID, load := range nodeLoads {
		// Assign more partitions to nodes with lower load
		partitionCount := int((1.0-load)*3) + 1

		for i := 0; i < partitionCount; i++ {
			partitions = append(partitions, &TaskPartition{
				ID:           fmt.Sprintf("partition_%d", partitionIndex),
				NodeID:       nodeID,
				Type:         "load_based",
				Data:         fmt.Sprintf("partition_data_%d", partitionIndex),
				Dependencies: []string{},
				Metadata: map[string]interface{}{
					"node_load":        load,
					"partition_weight": 1.0 - load,
				},
			})
			partitionIndex++
		}
	}

	return &PartitionPlan{
		ID:         fmt.Sprintf("plan_%d", time.Now().UnixNano()),
		Strategy:   lbps.GetName(),
		Partitions: partitions,
		Metadata: map[string]interface{}{
			"strategy":         "load_based",
			"node_loads":       nodeLoads,
			"total_partitions": len(partitions),
		},
		CreatedAt: time.Now(),
	}, nil
}

// CapabilityBasedPartitioningStrategy implements capability-based partitioning
type CapabilityBasedPartitioningStrategy struct {
	name string
}

func (cbps *CapabilityBasedPartitioningStrategy) GetName() string {
	return "capability_based"
}

func (cbps *CapabilityBasedPartitioningStrategy) Partition(request *OrchestrationRequest) (*PartitionPlan, error) {
	// Mock capability-based partitioning
	nodeCapabilities := map[string][]string{
		"node_0": {"cpu", "memory"},
		"node_1": {"gpu", "memory"},
		"node_2": {"cpu", "gpu", "memory"},
	}

	partitions := make([]*TaskPartition, 0)
	partitionIndex := 0

	for nodeID, capabilities := range nodeCapabilities {
		// Assign partitions based on capabilities
		for _, capability := range capabilities {
			partitions = append(partitions, &TaskPartition{
				ID:           fmt.Sprintf("partition_%d", partitionIndex),
				NodeID:       nodeID,
				Type:         "capability_based",
				Data:         fmt.Sprintf("partition_data_%s_%s", nodeID, capability),
				Dependencies: []string{},
				Metadata: map[string]interface{}{
					"required_capability": capability,
					"node_capabilities":   capabilities,
				},
			})
			partitionIndex++
		}
	}

	return &PartitionPlan{
		ID:         fmt.Sprintf("plan_%d", time.Now().UnixNano()),
		Strategy:   cbps.GetName(),
		Partitions: partitions,
		Metadata: map[string]interface{}{
			"strategy":          "capability_based",
			"node_capabilities": nodeCapabilities,
			"total_partitions":  len(partitions),
		},
		CreatedAt: time.Now(),
	}, nil
}
