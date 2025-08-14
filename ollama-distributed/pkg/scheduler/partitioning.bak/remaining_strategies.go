package partitioning

import (
	"context"
	"fmt"
	"time"
)

// TaskParallelismStrategy implements task parallelism partitioning
type TaskParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// NewTaskParallelismStrategy creates a new task parallelism strategy
func NewTaskParallelismStrategy() *TaskParallelismStrategy {
	return &TaskParallelismStrategy{
		name:    "task_parallel",
		metrics: &StrategyMetrics{LastUsed: time.Now()},
	}
}

func (tps *TaskParallelismStrategy) GetName() string {
	return tps.name
}

func (tps *TaskParallelismStrategy) GetMetrics() *StrategyMetrics {
	return tps.metrics
}

func (tps *TaskParallelismStrategy) CanHandle(task *PartitionTask) bool {
	return len(task.Nodes) >= 2 && task.Type == "multi_modal"
}

func (tps *TaskParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	tps.metrics.TotalPartitions++
	tps.metrics.LastUsed = time.Now()

	// Create simple task parallel partitions
	partitions := make([]*Partition, len(task.Nodes))
	for i, node := range task.Nodes {
		partitions[i] = &Partition{
			ID:               fmt.Sprintf("task_%d", i),
			NodeID:           node.ID,
			Type:             PartitionTypeTask,
			Data:             map[string]interface{}{"task_id": i},
			Dependencies:     []string{},
			EstimatedLatency: 100 * time.Millisecond,
			EstimatedMemory:  1024 * 1024 * 1024, // 1GB
			Metadata:         map[string]interface{}{"strategy": "task_parallel"},
		}
	}

	tps.metrics.SuccessfulPartitions++
	return &PartitionPlan{
		ID:                  fmt.Sprintf("task_parallel_%d", time.Now().UnixNano()),
		Strategy:            tps.name,
		Partitions:          partitions,
		EstimatedLatency:    100 * time.Millisecond,
		EstimatedThroughput: float64(len(partitions)) * 10.0,
		CreatedAt:           time.Now(),
		Metadata:            map[string]interface{}{"parallel_tasks": len(partitions)},
	}, nil
}

// SequenceParallelismStrategy implements sequence parallelism partitioning
type SequenceParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// NewSequenceParallelismStrategy creates a new sequence parallelism strategy
func NewSequenceParallelismStrategy() *SequenceParallelismStrategy {
	return &SequenceParallelismStrategy{
		name:    "sequence_parallel",
		metrics: &StrategyMetrics{LastUsed: time.Now()},
	}
}

func (sps *SequenceParallelismStrategy) GetName() string {
	return sps.name
}

func (sps *SequenceParallelismStrategy) GetMetrics() *StrategyMetrics {
	return sps.metrics
}

func (sps *SequenceParallelismStrategy) CanHandle(task *PartitionTask) bool {
	return len(task.Nodes) >= 2 && task.GetNumCtx() > 1024
}

func (sps *SequenceParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	sps.metrics.TotalPartitions++
	sps.metrics.LastUsed = time.Now()

	// Create sequence parallel partitions
	sequenceLength := task.GetNumCtx()
	partitionsCount := len(task.Nodes)
	sequencePerPartition := sequenceLength / partitionsCount

	partitions := make([]*Partition, partitionsCount)
	for i, node := range task.Nodes {
		startSeq := i * sequencePerPartition
		endSeq := startSeq + sequencePerPartition
		if i == partitionsCount-1 {
			endSeq = sequenceLength // Last partition gets remaining sequence
		}

		partitions[i] = &Partition{
			ID:               fmt.Sprintf("seq_%d", i),
			NodeID:           node.ID,
			Type:             PartitionTypeSequence,
			Data:             map[string]interface{}{"start_seq": startSeq, "end_seq": endSeq},
			Dependencies:     sps.getSequenceDependencies(i, partitionsCount),
			EstimatedLatency: 200 * time.Millisecond,
			EstimatedMemory:  int64(sequencePerPartition * 4 * 1024), // 4KB per token
			Metadata:         map[string]interface{}{"sequence_length": endSeq - startSeq},
		}
	}

	sps.metrics.SuccessfulPartitions++
	return &PartitionPlan{
		ID:                  fmt.Sprintf("sequence_parallel_%d", time.Now().UnixNano()),
		Strategy:            sps.name,
		Partitions:          partitions,
		EstimatedLatency:    time.Duration(partitionsCount) * 200 * time.Millisecond,
		EstimatedThroughput: float64(sequenceLength) / 2.0,
		CreatedAt:           time.Now(),
		Metadata:            map[string]interface{}{"sequence_length": sequenceLength, "partitions": partitionsCount},
	}, nil
}

func (sps *SequenceParallelismStrategy) getSequenceDependencies(index, total int) []string {
	if index == 0 {
		return []string{} // First partition has no dependencies
	}
	return []string{fmt.Sprintf("seq_%d", index-1)} // Depends on previous partition
}

// AttentionParallelismStrategy implements attention parallelism partitioning
type AttentionParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// NewAttentionParallelismStrategy creates a new attention parallelism strategy
func NewAttentionParallelismStrategy() *AttentionParallelismStrategy {
	return &AttentionParallelismStrategy{
		name:    "attention_parallel",
		metrics: &StrategyMetrics{LastUsed: time.Now()},
	}
}

func (aps *AttentionParallelismStrategy) GetName() string {
	return aps.name
}

func (aps *AttentionParallelismStrategy) GetMetrics() *StrategyMetrics {
	return aps.metrics
}

func (aps *AttentionParallelismStrategy) CanHandle(task *PartitionTask) bool {
	return len(task.Nodes) >= 2 && task.GetNumCtx() > 512
}

func (aps *AttentionParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	aps.metrics.TotalPartitions++
	aps.metrics.LastUsed = time.Now()

	// Create attention parallel partitions
	headCount := len(task.Nodes)
	headsPerPartition := 8 / headCount // Assume 8 attention heads
	if headsPerPartition < 1 {
		headsPerPartition = 1
	}

	partitions := make([]*Partition, headCount)
	for i, node := range task.Nodes {
		startHead := i * headsPerPartition
		endHead := startHead + headsPerPartition
		if i == headCount-1 {
			endHead = 8 // Last partition gets remaining heads
		}

		partitions[i] = &Partition{
			ID:               fmt.Sprintf("attn_%d", i),
			NodeID:           node.ID,
			Type:             PartitionTypeAttention,
			Data:             map[string]interface{}{"start_head": startHead, "end_head": endHead},
			Dependencies:     []string{}, // Attention heads can be computed in parallel
			EstimatedLatency: 150 * time.Millisecond,
			EstimatedMemory:  int64(task.GetNumCtx() * task.GetNumCtx() * 4), // O(n^2) for attention
			Metadata:         map[string]interface{}{"attention_heads": endHead - startHead},
		}
	}

	aps.metrics.SuccessfulPartitions++
	return &PartitionPlan{
		ID:                  fmt.Sprintf("attention_parallel_%d", time.Now().UnixNano()),
		Strategy:            aps.name,
		Partitions:          partitions,
		EstimatedLatency:    150 * time.Millisecond,
		EstimatedThroughput: float64(headCount) * 5.0,
		CreatedAt:           time.Now(),
		Metadata:            map[string]interface{}{"attention_heads": 8, "partitions": headCount},
	}, nil
}
