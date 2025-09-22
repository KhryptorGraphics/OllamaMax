package partitioning

import (
	"context"
	"fmt"
	"strings"
	"time"

	oldtypes "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// LegacyPartitionManager manages legacy workload partitioning strategies
// This is a compatibility wrapper for old code that will be deprecated
type LegacyPartitionManager struct {
	config     *Config
	strategies map[string]PartitionStrategy
}

// Config holds partitioning configuration
type Config struct {
	DefaultStrategy string `json:"default_strategy"`
	LayerThreshold  int    `json:"layer_threshold"`
	BatchSizeLimit  int    `json:"batch_size_limit"`
}

// LegacyPartitionTask represents a legacy task to be partitioned
// This is kept for backward compatibility and will be converted to PartitionRequest
type LegacyPartitionTask struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Model     *oldtypes.OllamaModel  `json:"model"`
	Options   map[string]interface{} `json:"options"`
	Nodes     []*LegacyNodeInfo      `json:"nodes"`
	Metadata  map[string]interface{} `json:"metadata"`
	Priority  int                    `json:"priority"`
	Timeout   time.Duration          `json:"timeout"`
	CreatedAt time.Time              `json:"created_at"`
}

// Helper functions for safe options access
func (pt *LegacyPartitionTask) GetNumCtx() int {
	if val, ok := pt.Options["num_ctx"]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if floatVal, ok := val.(float64); ok {
			return int(floatVal)
		}
	}
	return 2048 // default context length
}

// LegacyNodeInfo represents legacy node information for partitioning
type LegacyNodeInfo struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Capacity     *ResourceCapacity      `json:"capacity"`
	Usage        *ResourceUsage         `json:"usage"`
	GPUs         []GPUInfo              `json:"gpus"`
	Latency      time.Duration          `json:"latency"`
	Bandwidth    int64                  `json:"bandwidth"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Memory    int64  `json:"memory"`
	Compute   string `json:"compute"`
	Driver    string `json:"driver"`
	Available bool   `json:"available"`
}

// ResourceCapacity represents node resource capacity
type ResourceCapacity struct {
	CPUCores         int64   `json:"cpu_cores"`
	MemoryBytes      int64   `json:"memory_bytes"`
	GPUCount         int     `json:"gpu_count"`
	GPUMemoryBytes   int64   `json:"gpu_memory_bytes"`
	NetworkBandwidth int64   `json:"network_bandwidth"`
	StorageBytes     int64   `json:"storage_bytes"`
	Utilization      float64 `json:"utilization"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    int64     `json:"memory_usage"`
	GPUUsage       float64   `json:"gpu_usage"`
	GPUMemoryUsage int64     `json:"gpu_memory_usage"`
	NetworkUsage   int64     `json:"network_usage"`
	StorageUsage   int64     `json:"storage_usage"`
	ActiveTasks    int       `json:"active_tasks"`
	LastUpdated    time.Time `json:"last_updated"`
}

// LegacyPartitionPlan represents the legacy result of partitioning
type LegacyPartitionPlan struct {
	ID               string                 `json:"id"`
	TaskID           string                 `json:"task_id"`
	Strategy         string                 `json:"strategy"`
	Partitions       []LegacyPartition      `json:"partitions"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	EstimatedLatency time.Duration          `json:"estimated_latency"`
	EstimatedCost    float64                `json:"estimated_cost"`
}

// LegacyPartition represents a single legacy partition
type LegacyPartition struct {
	ID               string                 `json:"id"`
	NodeID           string                 `json:"node_id"`
	Type             string                 `json:"type"`
	Data             map[string]interface{} `json:"data"`
	Dependencies     []string               `json:"dependencies"`
	EstimatedLatency time.Duration          `json:"estimated_latency"`
	EstimatedMemory  int64                  `json:"estimated_memory"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Adapter functions to convert between legacy and new types

// ConvertLegacyTaskToRequest converts a legacy task to a PartitionRequest
func ConvertLegacyTaskToRequest(task *LegacyPartitionTask) *PartitionRequest {
	if task == nil {
		return nil
	}

	// Model details are not available in OllamaModel struct
	// This would need to be obtained from elsewhere
	var modelDetails *types.OllamaModelDetails

	modelInfo := &ModelInfo{
		Name: task.Model.Name,
		Path: task.Model.Path,
		Size: task.Model.Size,
		Details: modelDetails,
	}

	// Extract parameters from model name as a heuristic
	if task.Model != nil {
		modelInfo.Parameters = extractParameterCount(task.Model.Name)
	}

	options := &PartitionOptions{
		Strategy: "adaptive", // Default strategy
		MaxNodes: 10,
		MinNodes: 1,
		MemoryThreshold: 0.8,
		OptimizeFor: OptimizeBalance,
		CustomParams: task.Options,
	}

	return &PartitionRequest{
		TaskID: task.ID,
		Model: modelInfo,
		Options: options,
		Constraints: &PartitionConstraints{},
	}
}

// ConvertLegacyNodeToNodeInfo converts legacy node info to new NodeInfo
func ConvertLegacyNodeToNodeInfo(legacyNode *LegacyNodeInfo) *NodeInfo {
	if legacyNode == nil {
		return nil
	}

	var gpuCap *GPUCapabilities
	if len(legacyNode.GPUs) > 0 {
		var totalMemory int64
		for _, gpu := range legacyNode.GPUs {
			totalMemory += gpu.Memory
		}
		gpuCap = &GPUCapabilities{
			Count: len(legacyNode.GPUs),
			TotalMemory: totalMemory,
		}
		if len(legacyNode.GPUs) > 0 {
			gpuCap.Model = legacyNode.GPUs[0].Name
			gpuCap.MemoryPerGPU = legacyNode.GPUs[0].Memory
			gpuCap.ComputeCapability = legacyNode.GPUs[0].Compute
		}
	}

	var cpuCap *CPUCapabilities
	var memCap *MemoryCapabilities
	if legacyNode.Capacity != nil {
		cpuCap = &CPUCapabilities{
			Cores: int(legacyNode.Capacity.CPUCores),
		}
		memCap = &MemoryCapabilities{
			TotalBytes: legacyNode.Capacity.MemoryBytes,
		}
		if legacyNode.Usage != nil {
			memCap.AvailableBytes = legacyNode.Capacity.MemoryBytes - legacyNode.Usage.MemoryUsage
		}
	}

	var currentLoad *NodeLoad
	if legacyNode.Usage != nil {
		currentLoad = &NodeLoad{
			CPUUtilization: legacyNode.Usage.CPUUsage,
			MemoryUtilization: float64(legacyNode.Usage.MemoryUsage) / float64(legacyNode.Capacity.MemoryBytes),
			GPUUtilization: legacyNode.Usage.GPUUsage,
			ActiveTasks: legacyNode.Usage.ActiveTasks,
		}
	}

	metadata := make(map[string]string)
	for k, v := range legacyNode.Metadata {
		metadata[k] = fmt.Sprintf("%v", v)
	}

	return &NodeInfo{
		ID: legacyNode.ID,
		Address: legacyNode.Address,
		Status: NodeStatusActive,
		Capabilities: &NodeCapabilities{
			CPU: cpuCap,
			Memory: memCap,
			GPU: gpuCap,
			Network: &NetworkCapabilities{
				Bandwidth: float64(legacyNode.Bandwidth) / 1e9, // Convert to Gbps
				Latency: float64(legacyNode.Latency.Milliseconds()),
			},
			Storage: &StorageCapabilities{},
		},
		CurrentLoad: currentLoad,
		Metadata: metadata,
	}
}

// Helper function to extract parameter count from string
func extractParameterCountFromString(paramSize string) int64 {
	// Simple extraction logic - can be improved
	var params int64
	if _, err := fmt.Sscanf(paramSize, "%dB", &params); err == nil {
		return params * 1_000_000_000
	}
	if _, err := fmt.Sscanf(paramSize, "%dM", &params); err == nil {
		return params * 1_000_000
	}
	return 0
}

// NewLegacyPartitionManager creates a new legacy partition manager
func NewLegacyPartitionManager(config *Config) *LegacyPartitionManager {
	return &LegacyPartitionManager{
		config:     config,
		strategies: make(map[string]PartitionStrategy),
	}
}

// RegisterStrategy registers a partitioning strategy
func (pm *LegacyPartitionManager) RegisterStrategy(strategy PartitionStrategy) {
	pm.strategies[strategy.GetName()] = strategy
}

// SelectStrategy selects the best partitioning strategy for a task
func (pm *LegacyPartitionManager) SelectStrategy(task interface{}, model *oldtypes.OllamaModel, opts map[string]interface{}) (string, error) {
	// Try to select strategy based on model and options
	if model != nil && model.Name != "" {
		// Determine strategy based on model name patterns
		modelName := strings.ToLower(model.Name)
		if strings.Contains(modelName, "70b") || strings.Contains(modelName, "33b") || strings.Contains(modelName, "65b") {
			return "tensor", nil
		}
		// Medium models work well with pipeline
		if strings.Contains(modelName, "13b") || strings.Contains(modelName, "7b") {
			return "pipeline", nil
		}
	}
	
	// Check if a specific strategy is requested in options
	if opts != nil {
		if strategy, ok := opts["strategy"].(string); ok {
			normalized := NormalizeStrategyName(strategy)
			if IsValidStrategy(normalized) {
				return normalized, nil
			}
		}
	}
	
	// Fall back to configured default
	if pm.config.DefaultStrategy != "" {
		return NormalizeStrategyName(pm.config.DefaultStrategy), nil
	}
	
	// Ultimate fallback to adaptive strategy
	return "adaptive", nil
}

// Partition partitions a task using the specified strategy
// This method converts legacy types to new types and delegates to the new implementation
func (pm *LegacyPartitionManager) Partition(ctx context.Context, task *LegacyPartitionTask, strategyName string) (*PartitionPlan, error) {
	// Convert legacy task to new PartitionRequest
	req := ConvertLegacyTaskToRequest(task)
	
	// Convert legacy nodes to new NodeInfo
	var nodes []*NodeInfo
	for _, legacyNode := range task.Nodes {
		nodes = append(nodes, ConvertLegacyNodeToNodeInfo(legacyNode))
	}

	// Use the strategy if available, otherwise create a real strategy
	strategy, exists := pm.strategies[strategyName]
	if !exists {
		// Try to create a real strategy based on the name
		switch strategyName {
		case "layer", "layerwise":
			strategy = NewLayerwiseStrategy()
		case "tensor", "tensor_parallelism":
			strategy = NewTensorParallelismStrategy()
		case "pipeline", "pipeline_parallelism":
			strategy = NewPipelineParallelismStrategy()
		case "hybrid", "hybrid_parallelism":
			strategy = NewHybridParallelismStrategy()
		case "adaptive", "adaptive_partitioning":
			strategy = NewAdaptivePartitioningStrategy()
		default:
			// Fall back to adaptive strategy as it can learn and adapt
			strategy = NewAdaptivePartitioningStrategy()
		}
		// Cache the created strategy
		pm.strategies[strategyName] = strategy
	}

	return strategy.Partition(ctx, req, nodes)
}

// Stub strategy implementations for backward compatibility
// These will delegate to the new strategies in the factory

// stubStrategy is a simple adapter implementation that converts between legacy and new APIs
type stubStrategy struct {
	name string
}

func (s *stubStrategy) GetName() string {
	return s.name
}

func (s *stubStrategy) Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error) {
	return &PartitionPlan{
		ID:       fmt.Sprintf("plan_%s_%d", s.name, time.Now().Unix()),
		TaskID:   req.TaskID,
		Strategy: s.name,
		Assignments: []*NodeAssignment{
			{
				NodeID: "default-node",
				Role:   RoleWorker,
				WorkType: WorkTypeLayers,
				Assignment: &WorkAssignment{
					CustomData: make(map[string]interface{}),
				},
			},
		},
		CreatedAt: time.Now(),
	}, nil
}

func (s *stubStrategy) CanHandle(req *PartitionRequest) bool {
	return true
}

func (s *stubStrategy) GetMetrics() *StrategyMetrics {
	return &StrategyMetrics{
		UsageCount:  0,
		SuccessCount: 0,
		FailureCount: 0,
		LastUsed:    time.Now(),
	}
}
