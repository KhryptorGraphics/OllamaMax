package partitioning

import (
	"context"
	"fmt"
	"time"

	"github.com/khryptorgraphics/ollamamax/pkg/types"
	api_types "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
)

// PartitionStrategy defines the interface for different partitioning strategies
type PartitionStrategy interface {
	// Partition creates a partition plan for the given request
	Partition(ctx context.Context, req *PartitionRequest, nodes []*NodeInfo) (*PartitionPlan, error)
	
	// CanHandle determines if this strategy can handle the given request
	CanHandle(req *PartitionRequest) bool
	
	// GetName returns the strategy name
	GetName() string
	
	// GetMetrics returns strategy-specific metrics
	GetMetrics() *StrategyMetrics
}

// PartitionRequest contains information needed for partitioning
type PartitionRequest struct {
	TaskID        string                     `json:"task_id"`
	Model         *ModelInfo                 `json:"model"`
	ModelAnalysis *ModelAnalysis             `json:"model_analysis,omitempty"`
	Options       *PartitionOptions          `json:"options"`
	Constraints   *PartitionConstraints      `json:"constraints"`
}

// ModelInfo contains basic model information
type ModelInfo struct {
	Name         string                     `json:"name"`
	Path         string                     `json:"path"`
	Size         int64                      `json:"size"`         // Size in bytes
	Parameters   int64                      `json:"parameters"`   // Parameter count
	Details      *types.OllamaModelDetails  `json:"details,omitempty"`
	Family       string                     `json:"family"`       // Model family (llama, gpt, etc.)
}

// PartitionOptions contains options for partitioning
type PartitionOptions struct {
	Strategy         string            `json:"strategy"`          // Preferred strategy name
	MaxNodes         int               `json:"max_nodes"`         // Maximum nodes to use
	MinNodes         int               `json:"min_nodes"`         // Minimum nodes to use
	MemoryThreshold  float64           `json:"memory_threshold"`  // Memory utilization threshold (0.0-1.0)
	LoadBalance      bool              `json:"load_balance"`      // Enable load balancing
	OptimizeFor      OptimizationTarget `json:"optimize_for"`     // What to optimize for
	CustomParams     map[string]interface{} `json:"custom_params"` // Strategy-specific parameters
}

// PartitionConstraints contains constraints for partitioning
type PartitionConstraints struct {
	MaxMemoryPerNode int64             `json:"max_memory_per_node"` // Maximum memory per node in bytes
	MaxLatency       time.Duration     `json:"max_latency"`         // Maximum acceptable latency
	RequiredNodes    []string          `json:"required_nodes"`      // Specific nodes that must be used
	ExcludedNodes    []string          `json:"excluded_nodes"`      // Nodes that should not be used
	Locality         *LocalityConstraints `json:"locality,omitempty"` // Locality constraints
}

// LocalityConstraints defines locality requirements
type LocalityConstraints struct {
	PreferSameRegion bool   `json:"prefer_same_region"`
	PreferSameZone   bool   `json:"prefer_same_zone"`
	MaxNetworkHops   int    `json:"max_network_hops"`
}

// OptimizationTarget defines what to optimize for
type OptimizationTarget string

const (
	OptimizeLatency    OptimizationTarget = "latency"
	OptimizeThroughput OptimizationTarget = "throughput"
	OptimizeMemory     OptimizationTarget = "memory"
	OptimizeBalance    OptimizationTarget = "balance"
)

// NodeInfo contains information about available nodes
type NodeInfo struct {
	ID           string                `json:"id"`
	Address      string                `json:"address"`
	Status       NodeStatus            `json:"status"`
	Capabilities *NodeCapabilities     `json:"capabilities"`
	CurrentLoad  *NodeLoad             `json:"current_load"`
	Metadata     map[string]string     `json:"metadata"`
}

// NodeStatus represents the current status of a node
type NodeStatus string

const (
	NodeStatusActive      NodeStatus = "active"
	NodeStatusInactive    NodeStatus = "inactive"
	NodeStatusMaintenance NodeStatus = "maintenance"
	NodeStatusOverloaded  NodeStatus = "overloaded"
)

// NodeCapabilities contains detailed node capability information
type NodeCapabilities struct {
	CPU          *CPUCapabilities     `json:"cpu"`
	Memory       *MemoryCapabilities  `json:"memory"`
	GPU          *GPUCapabilities     `json:"gpu,omitempty"`
	Network      *NetworkCapabilities `json:"network"`
	Storage      *StorageCapabilities `json:"storage"`
}

// CPUCapabilities contains CPU-related capabilities
type CPUCapabilities struct {
	Cores       int     `json:"cores"`
	ThreadsPerCore int  `json:"threads_per_core"`
	Architecture string `json:"architecture"`
	Frequency   float64 `json:"frequency_ghz"`
	CacheSize   int64   `json:"cache_size_bytes"`
}

// MemoryCapabilities contains memory-related capabilities
type MemoryCapabilities struct {
	TotalBytes    int64   `json:"total_bytes"`
	AvailableBytes int64  `json:"available_bytes"`
	Type          string  `json:"type"`          // DDR4, DDR5, etc.
	Speed         int     `json:"speed_mhz"`
	Bandwidth     float64 `json:"bandwidth_gbps"`
}

// GPUCapabilities contains GPU-related capabilities
type GPUCapabilities struct {
	Count          int      `json:"count"`
	Model          string   `json:"model"`
	MemoryPerGPU   int64    `json:"memory_per_gpu_bytes"`
	TotalMemory    int64    `json:"total_memory_bytes"`
	ComputeCapability string `json:"compute_capability"`
	CudaCores      int      `json:"cuda_cores,omitempty"`
	TensorCores    int      `json:"tensor_cores,omitempty"`
}

// NetworkCapabilities contains network-related capabilities
type NetworkCapabilities struct {
	Bandwidth     float64 `json:"bandwidth_gbps"`
	Latency       float64 `json:"latency_ms"`
	InterconnectType string `json:"interconnect_type"` // InfiniBand, Ethernet, etc.
	Topology      string  `json:"topology"`           // Fat tree, mesh, etc.
}

// StorageCapabilities contains storage-related capabilities
type StorageCapabilities struct {
	TotalBytes     int64   `json:"total_bytes"`
	AvailableBytes int64   `json:"available_bytes"`
	Type           string  `json:"type"`           // SSD, NVMe, HDD
	ReadSpeed      float64 `json:"read_speed_mbps"`
	WriteSpeed     float64 `json:"write_speed_mbps"`
}

// NodeLoad contains current load information for a node
type NodeLoad struct {
	CPUUtilization    float64 `json:"cpu_utilization"`    // 0.0-1.0
	MemoryUtilization float64 `json:"memory_utilization"` // 0.0-1.0
	GPUUtilization    float64 `json:"gpu_utilization"`    // 0.0-1.0
	NetworkUtilization float64 `json:"network_utilization"` // 0.0-1.0
	ActiveTasks       int     `json:"active_tasks"`
	QueuedTasks       int     `json:"queued_tasks"`
}

// PartitionPlan contains the result of partitioning
type PartitionPlan struct {
	ID            string                 `json:"id"`
	TaskID        string                 `json:"task_id"`
	Strategy      string                 `json:"strategy"`
	Assignments   []*NodeAssignment      `json:"assignments"`
	Communication *CommunicationPlan     `json:"communication"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	EstimatedCost *ResourceCost          `json:"estimated_cost,omitempty"`
}

// NodeAssignment represents work assigned to a specific node
type NodeAssignment struct {
	NodeID      string                 `json:"node_id"`
	Role        AssignmentRole         `json:"role"`
	WorkType    WorkType               `json:"work_type"`
	Assignment  *WorkAssignment        `json:"assignment"`
	Resources   *ResourceRequirements  `json:"resources"`
	Dependencies []string              `json:"dependencies"` // Dependencies on other assignments
}

// AssignmentRole defines the role of a node in the partition
type AssignmentRole string

const (
	RolePrimary   AssignmentRole = "primary"
	RoleSecondary AssignmentRole = "secondary"
	RoleWorker    AssignmentRole = "worker"
	RoleCoordinator AssignmentRole = "coordinator"
)

// WorkType defines the type of work assigned
type WorkType string

const (
	WorkTypeLayers         WorkType = "layers"
	WorkTypeTensors        WorkType = "tensors"
	WorkTypePipelineStage  WorkType = "pipeline_stage"
	WorkTypeHybrid         WorkType = "hybrid"
)

// WorkAssignment contains the specific work assignment details
type WorkAssignment struct {
	Layers       []LayerAssignment   `json:"layers,omitempty"`
	Tensors      []TensorAssignment  `json:"tensors,omitempty"`
	PipelineStage *PipelineStageAssignment `json:"pipeline_stage,omitempty"`
	CustomData   map[string]interface{} `json:"custom_data,omitempty"`
}

// LayerAssignment represents layer assignment details
type LayerAssignment struct {
	LayerIndices  []int     `json:"layer_indices"`
	LayerNames    []string  `json:"layer_names"`
	TotalLayers   int       `json:"total_layers"`
	StartIndex    int       `json:"start_index"`
	EndIndex      int       `json:"end_index"`
	MemoryRequired int64    `json:"memory_required"`
	ComputeWeight  float64  `json:"compute_weight"`
}

// TensorAssignment represents tensor assignment details
type TensorAssignment struct {
	TensorIndices []int               `json:"tensor_indices"`
	TensorNames   []string            `json:"tensor_names"`
	SplitInfo     []TensorSplitInfo   `json:"split_info"`
	MemoryRequired int64              `json:"memory_required"`
	ComputeWeight  float64            `json:"compute_weight"`
}

// TensorSplitInfo contains information about how a tensor is split
type TensorSplitInfo struct {
	TensorIndex   int     `json:"tensor_index"`
	TensorName    string  `json:"tensor_name"`
	SplitDimension int    `json:"split_dimension"`
	SplitStart    int     `json:"split_start"`
	SplitEnd      int     `json:"split_end"`
	OriginalShape []int   `json:"original_shape"`
	SplitShape    []int   `json:"split_shape"`
}

// PipelineStageAssignment represents pipeline stage assignment details
type PipelineStageAssignment struct {
	StageIndex      int       `json:"stage_index"`
	LayerRange      []int     `json:"layer_range"` // [start, end]
	InputSpecs      []TensorSpec `json:"input_specs"`
	OutputSpecs     []TensorSpec `json:"output_specs"`
	MemoryRequired  int64     `json:"memory_required"`
	ComputeWeight   float64   `json:"compute_weight"`
	IsFirstStage    bool      `json:"is_first_stage"`
	IsLastStage     bool      `json:"is_last_stage"`
}

// TensorSpec contains tensor specification information
type TensorSpec struct {
	Name  string `json:"name"`
	Shape []int  `json:"shape"`
	Type  string `json:"type"`
}

// ResourceRequirements specifies resource requirements for an assignment
type ResourceRequirements struct {
	CPU          *CPURequirement    `json:"cpu"`
	Memory       *MemoryRequirement `json:"memory"`
	GPU          *GPURequirement    `json:"gpu,omitempty"`
	Network      *NetworkRequirement `json:"network"`
	Storage      *StorageRequirement `json:"storage"`
}

// CPURequirement specifies CPU requirements
type CPURequirement struct {
	Cores     int     `json:"cores"`
	Utilization float64 `json:"utilization"` // Expected utilization (0.0-1.0)
}

// MemoryRequirement specifies memory requirements
type MemoryRequirement struct {
	Bytes       int64   `json:"bytes"`
	Utilization float64 `json:"utilization"` // Expected utilization (0.0-1.0)
}

// GPURequirement specifies GPU requirements
type GPURequirement struct {
	Count       int     `json:"count"`
	MemoryBytes int64   `json:"memory_bytes"`
	Utilization float64 `json:"utilization"` // Expected utilization (0.0-1.0)
}

// NetworkRequirement specifies network requirements
type NetworkRequirement struct {
	BandwidthGbps float64 `json:"bandwidth_gbps"`
	LatencyMs     float64 `json:"latency_ms"`
}

// StorageRequirement specifies storage requirements
type StorageRequirement struct {
	Bytes     int64 `json:"bytes"`
	IOPSRead  int   `json:"iops_read"`
	IOPSWrite int   `json:"iops_write"`
}

// CommunicationPlan defines how nodes communicate
type CommunicationPlan struct {
	Topology      CommunicationTopology  `json:"topology"`
	Connections   []NodeConnection       `json:"connections"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// CommunicationTopology defines the communication topology
type CommunicationTopology string

const (
	TopologyPointToPoint CommunicationTopology = "point_to_point"
	TopologyAllToAll     CommunicationTopology = "all_to_all"
	TopologyRing         CommunicationTopology = "ring"
	TopologyTree         CommunicationTopology = "tree"
	TopologyMesh         CommunicationTopology = "mesh"
)

// NodeConnection represents a connection between two nodes
type NodeConnection struct {
	From          string                 `json:"from"`
	To            string                 `json:"to"`
	ConnectionType ConnectionType        `json:"connection_type"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// ConnectionType defines the type of connection
type ConnectionType string

const (
	ConnectionTypeData     ConnectionType = "data"
	ConnectionTypeControl  ConnectionType = "control"
	ConnectionTypeGradient ConnectionType = "gradient"
	ConnectionTypePipeline ConnectionType = "pipeline"
)

// ResourceCost estimates the cost of executing the partition plan
type ResourceCost struct {
	ComputeCost   float64                `json:"compute_cost"`
	MemoryCost    float64                `json:"memory_cost"`
	NetworkCost   float64                `json:"network_cost"`
	StorageCost   float64                `json:"storage_cost"`
	TotalCost     float64                `json:"total_cost"`
	Details       map[string]interface{} `json:"details"`
}

// StrategyMetrics contains metrics for a partitioning strategy
type StrategyMetrics struct {
	UsageCount        int64                  `json:"usage_count"`
	SuccessCount      int64                  `json:"success_count"`
	FailureCount      int64                  `json:"failure_count"`
	AverageLatency    time.Duration          `json:"average_latency"`
	AverageThroughput float64                `json:"average_throughput"`
	LastUsed          time.Time              `json:"last_used"`
	Performance       *PerformanceMetrics    `json:"performance"`
	CustomMetrics     map[string]interface{} `json:"custom_metrics"`
}

// PerformanceMetrics contains performance-related metrics
type PerformanceMetrics struct {
	ExecutionTimeMs   []float64 `json:"execution_time_ms"`
	MemoryUsageBytes  []int64   `json:"memory_usage_bytes"`
	NetworkBandwidth  []float64 `json:"network_bandwidth_gbps"`
	QualityScore      float64   `json:"quality_score"`      // 0.0-1.0
	EfficiencyScore   float64   `json:"efficiency_score"`   // 0.0-1.0
}

// PartitionValidationResult contains validation results for a partition plan
type PartitionValidationResult struct {
	Valid        bool                   `json:"valid"`
	Errors       []ValidationError      `json:"errors"`
	Warnings     []ValidationWarning    `json:"warnings"`
	Score        float64                `json:"score"`        // Quality score 0.0-1.0
	Details      map[string]interface{} `json:"details"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	NodeID      string `json:"node_id,omitempty"`
	Severity    ErrorSeverity `json:"severity"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	NodeID      string `json:"node_id,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ErrorSeverity defines the severity of validation errors
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
)

// PartitionManager interface defines the main partition management interface
type PartitionManager interface {
	// Partition creates a partition plan for the given task
	Partition(ctx context.Context, task *api_types.DistributedTask) (*PartitionPlan, error)
	
	// SelectStrategy selects the best strategy for the given task
	SelectStrategy(task *api_types.DistributedTask, model *ModelInfo, options *PartitionOptions) (string, error)
	
	// ValidatePartition validates a partition plan
	ValidatePartition(ctx context.Context, plan *PartitionPlan, nodes []*NodeInfo) (*PartitionValidationResult, error)
	
	// GetAvailableStrategies returns available partitioning strategies
	GetAvailableStrategies() []string
	
	// RegisterStrategy registers a new partitioning strategy
	RegisterStrategy(name string, strategy PartitionStrategy) error
	
	// GetStrategy returns a strategy by name
	GetStrategy(name string) (PartitionStrategy, error)
}

// EnhancedPartitionManager interface extends the base partition manager with additional capabilities
type EnhancedPartitionManager interface {
	PartitionManager
	
	// GetStrategyMetrics returns metrics for a specific strategy
	GetStrategyMetrics(strategyName string) (*StrategyMetrics, error)
	
	// GetSelectionHistory returns the history of strategy selections
	GetSelectionHistory(limit int) ([]*StrategySelectionRecord, error)
	
	// OptimizeStrategy optimizes strategy selection based on historical data
	OptimizeStrategy(ctx context.Context, req *PartitionRequest) (string, error)
	
	// UpdateMetrics updates metrics for a strategy based on execution results
	UpdateMetrics(strategyName string, result *PartitionExecutionResult) error
	
	// GetRecommendations returns recommendations for improving partitioning
	GetRecommendations(ctx context.Context, req *PartitionRequest) ([]*PartitionRecommendation, error)
}

// StrategySelectionRecord records strategy selection decisions
type StrategySelectionRecord struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	TaskID          string                 `json:"task_id"`
	SelectedStrategy string                `json:"selected_strategy"`
	AlternativeStrategies []string         `json:"alternative_strategies"`
	SelectionReason string                 `json:"selection_reason"`
	ModelInfo       *ModelInfo             `json:"model_info"`
	NodeCount       int                    `json:"node_count"`
	Success         bool                   `json:"success"`
	ExecutionMetrics *PerformanceMetrics   `json:"execution_metrics,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PartitionExecutionResult contains the result of executing a partition plan
type PartitionExecutionResult struct {
	PlanID          string                 `json:"plan_id"`
	TaskID          string                 `json:"task_id"`
	Strategy        string                 `json:"strategy"`
	Success         bool                   `json:"success"`
	ExecutionTime   time.Duration          `json:"execution_time"`
	ThroughputTPS   float64                `json:"throughput_tps"`
	MemoryUsage     int64                  `json:"memory_usage_bytes"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	Metrics         *PerformanceMetrics    `json:"metrics"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// PartitionRecommendation provides recommendations for partition optimization
type PartitionRecommendation struct {
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Actions     []string               `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecommendationType defines the type of recommendation
type RecommendationType string

const (
	RecommendationStrategy     RecommendationType = "strategy"
	RecommendationResources    RecommendationType = "resources"
	RecommendationTopology     RecommendationType = "topology"
	RecommendationOptimization RecommendationType = "optimization"
)

// RecommendationPriority defines the priority of recommendations
type RecommendationPriority string

const (
	RecommendationPriorityHigh   RecommendationPriority = "high"
	RecommendationPriorityMedium RecommendationPriority = "medium"
	RecommendationPriorityLow    RecommendationPriority = "low"
)

// Helper functions for type validation and conversion

// ValidatePartitionRequest validates a partition request
func ValidatePartitionRequest(req *PartitionRequest) error {
	if req == nil {
		return fmt.Errorf("partition request cannot be nil")
	}
	if req.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if req.Model == nil {
		return fmt.Errorf("model information cannot be nil")
	}
	if req.Options == nil {
		req.Options = &PartitionOptions{
			MaxNodes: 10,
			MinNodes: 1,
			MemoryThreshold: 0.8,
			OptimizeFor: OptimizeBalance,
		}
	}
	return nil
}

// ValidateNodeInfo validates node information
func ValidateNodeInfo(node *NodeInfo) error {
	if node == nil {
		return fmt.Errorf("node info cannot be nil")
	}
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if node.Capabilities == nil {
		return fmt.Errorf("node capabilities cannot be nil")
	}
	// Validate that CPU and Memory are present (required for resource calculations)
	if node.Capabilities.CPU == nil {
		return fmt.Errorf("node capabilities CPU cannot be nil")
	}
	if node.Capabilities.Memory == nil {
		return fmt.Errorf("node capabilities Memory cannot be nil")
	}
	return nil
}

// CalculateResourceUtilization calculates resource utilization percentage
func CalculateResourceUtilization(required *ResourceRequirements, available *NodeCapabilities) float64 {
	if required == nil || available == nil {
		return 0.0
	}

	var totalUtilization float64
	var componentCount int

	// CPU utilization
	if required.CPU != nil && available.CPU != nil {
		cpuUtil := float64(required.CPU.Cores) / float64(available.CPU.Cores)
		totalUtilization += cpuUtil
		componentCount++
	}

	// Memory utilization
	if required.Memory != nil && available.Memory != nil {
		memUtil := float64(required.Memory.Bytes) / float64(available.Memory.TotalBytes)
		totalUtilization += memUtil
		componentCount++
	}

	// GPU utilization
	if required.GPU != nil && available.GPU != nil && available.GPU.Count > 0 {
		gpuUtil := float64(required.GPU.Count) / float64(available.GPU.Count)
		totalUtilization += gpuUtil
		componentCount++
	}

	if componentCount == 0 {
		return 0.0
	}

	return totalUtilization / float64(componentCount)
}