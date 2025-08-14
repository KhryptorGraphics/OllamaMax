package resource

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/types"
)

// ResourceManager manages efficient resource allocation and monitoring
type ResourceManager struct {
	mu sync.RWMutex

	// Resource tracking
	nodeResources map[string]*NodeResourceState
	resourcePools map[string]*ResourcePool
	allocations   map[string]*ResourceAllocation

	// Quotas and limits
	quotas       map[string]*ResourceQuota
	globalLimits *ResourceLimits

	// Configuration
	config *ResourceManagerConfig

	// Metrics and monitoring
	metrics      *ResourceMetrics
	usageHistory []*ResourceUsageSnapshot

	// Optimization
	optimizer *ResourceOptimizer

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NodeResourceState represents the current resource state of a node
type NodeResourceState struct {
	NodeID    string              `json:"node_id"`
	Capacity  *types.NodeCapacity `json:"capacity"`
	Available *types.NodeCapacity `json:"available"`
	Allocated *types.NodeCapacity `json:"allocated"`
	Reserved  *types.NodeCapacity `json:"reserved"`

	// Resource utilization
	Utilization *ResourceUtilization `json:"utilization"`

	// Performance characteristics
	Performance *ResourcePerformance `json:"performance"`

	// Health and status
	HealthScore float64            `json:"health_score"`
	Status      NodeResourceStatus `json:"status"`
	LastUpdate  time.Time          `json:"last_update"`

	// Constraints and preferences
	Constraints *ResourceConstraints `json:"constraints"`
	Tags        map[string]string    `json:"tags"`
}

// ResourcePool represents a pool of resources that can be shared
type ResourcePool struct {
	PoolID string           `json:"pool_id"`
	Name   string           `json:"name"`
	Type   ResourcePoolType `json:"type"`

	// Pool capacity
	TotalCapacity     *types.NodeCapacity `json:"total_capacity"`
	AvailableCapacity *types.NodeCapacity `json:"available_capacity"`

	// Pool members
	Members []string `json:"members"` // Node IDs

	// Pool policies
	AllocationPolicy AllocationPolicy `json:"allocation_policy"`
	SharingPolicy    SharingPolicy    `json:"sharing_policy"`

	// Pool metrics
	Utilization float64 `json:"utilization"`
	Efficiency  float64 `json:"efficiency"`

	// Pool status
	Status    PoolStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ResourceAllocation represents an active resource allocation
type ResourceAllocation struct {
	AllocationID string `json:"allocation_id"`
	TaskID       string `json:"task_id"`
	NodeID       string `json:"node_id"`
	PoolID       string `json:"pool_id,omitempty"`

	// Allocated resources
	AllocatedResources *types.ResourceRequirement `json:"allocated_resources"`

	// Allocation metadata
	Priority  types.TaskPriority `json:"priority"`
	StartTime time.Time          `json:"start_time"`
	Duration  time.Duration      `json:"duration,omitempty"`

	// Allocation status
	Status AllocationStatus `json:"status"`

	// Performance tracking
	ActualUsage *types.ResourceRequirement `json:"actual_usage,omitempty"`
	Efficiency  float64                    `json:"efficiency"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// ResourceQuota defines resource usage limits
type ResourceQuota struct {
	QuotaID string     `json:"quota_id"`
	Name    string     `json:"name"`
	Scope   QuotaScope `json:"scope"`
	Target  string     `json:"target"` // User, team, or project ID

	// Quota limits
	Limits *types.ResourceRequirement `json:"limits"`

	// Current usage
	CurrentUsage *types.ResourceRequirement `json:"current_usage"`

	// Quota policies
	EnforcementPolicy EnforcementPolicy `json:"enforcement_policy"`
	OveragePolicy     OveragePolicy     `json:"overage_policy"`

	// Quota status
	Status    QuotaStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ResourceLimits defines global resource limits
type ResourceLimits struct {
	MaxCPUCores    float64 `json:"max_cpu_cores"`
	MaxMemoryBytes int64   `json:"max_memory_bytes"`
	MaxDiskBytes   int64   `json:"max_disk_bytes"`
	MaxGPUCores    int     `json:"max_gpu_cores"`

	// Per-node limits
	MaxCPUPerNode    float64 `json:"max_cpu_per_node"`
	MaxMemoryPerNode int64   `json:"max_memory_per_node"`

	// Per-task limits
	MaxCPUPerTask    float64       `json:"max_cpu_per_task"`
	MaxMemoryPerTask int64         `json:"max_memory_per_task"`
	MaxTaskDuration  time.Duration `json:"max_task_duration"`
}

// ResourceUtilization represents resource utilization metrics
type ResourceUtilization struct {
	CPUUtilization     float64 `json:"cpu_utilization"`
	MemoryUtilization  float64 `json:"memory_utilization"`
	DiskUtilization    float64 `json:"disk_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`
	GPUUtilization     float64 `json:"gpu_utilization,omitempty"`

	// Composite metrics
	OverallUtilization float64 `json:"overall_utilization"`
	EfficiencyScore    float64 `json:"efficiency_score"`

	// Trend indicators
	UtilizationTrend   UtilizationTrend `json:"utilization_trend"`
	PeakUtilization    float64          `json:"peak_utilization"`
	AverageUtilization float64          `json:"average_utilization"`
}

// ResourcePerformance represents resource performance characteristics
type ResourcePerformance struct {
	CPUPerformance   float64 `json:"cpu_performance"`
	MemoryBandwidth  float64 `json:"memory_bandwidth"`
	DiskIOPS         float64 `json:"disk_iops"`
	NetworkBandwidth float64 `json:"network_bandwidth"`
	GPUPerformance   float64 `json:"gpu_performance,omitempty"`

	// Performance scores
	OverallPerformance float64           `json:"overall_performance"`
	PerformanceRating  PerformanceRating `json:"performance_rating"`

	// Benchmarks
	BenchmarkScores map[string]float64 `json:"benchmark_scores"`
	LastBenchmark   time.Time          `json:"last_benchmark"`
}

// ResourceConstraints defines resource allocation constraints
type ResourceConstraints struct {
	MinResources       *types.ResourceRequirement `json:"min_resources,omitempty"`
	MaxResources       *types.ResourceRequirement `json:"max_resources,omitempty"`
	PreferredResources *types.ResourceRequirement `json:"preferred_resources,omitempty"`

	// Affinity rules
	NodeAffinity     map[string]string `json:"node_affinity,omitempty"`
	ResourceAffinity []string          `json:"resource_affinity,omitempty"`
	AntiAffinity     []string          `json:"anti_affinity,omitempty"`

	// Placement constraints
	RequiredFeatures  []string `json:"required_features,omitempty"`
	ForbiddenFeatures []string `json:"forbidden_features,omitempty"`
	PreferredZones    []string `json:"preferred_zones,omitempty"`
}

// Enums and constants
type NodeResourceStatus string

const (
	NodeResourceStatusHealthy     NodeResourceStatus = "healthy"
	NodeResourceStatusDegraded    NodeResourceStatus = "degraded"
	NodeResourceStatusOverloaded  NodeResourceStatus = "overloaded"
	NodeResourceStatusMaintenance NodeResourceStatus = "maintenance"
	NodeResourceStatusUnavailable NodeResourceStatus = "unavailable"
)

type ResourcePoolType string

const (
	ResourcePoolTypeShared    ResourcePoolType = "shared"
	ResourcePoolTypeDedicated ResourcePoolType = "dedicated"
	ResourcePoolTypeElastic   ResourcePoolType = "elastic"
	ResourcePoolTypeSpot      ResourcePoolType = "spot"
)

type AllocationPolicy string

const (
	AllocationPolicyFirstFit AllocationPolicy = "first_fit"
	AllocationPolicyBestFit  AllocationPolicy = "best_fit"
	AllocationPolicyWorstFit AllocationPolicy = "worst_fit"
	AllocationPolicyBalanced AllocationPolicy = "balanced"
)

type SharingPolicy string

const (
	SharingPolicyExclusive   SharingPolicy = "exclusive"
	SharingPolicyShared      SharingPolicy = "shared"
	SharingPolicyPreemptible SharingPolicy = "preemptible"
)

type PoolStatus string

const (
	PoolStatusActive      PoolStatus = "active"
	PoolStatusInactive    PoolStatus = "inactive"
	PoolStatusMaintenance PoolStatus = "maintenance"
	PoolStatusDraining    PoolStatus = "draining"
)

type AllocationStatus string

const (
	AllocationStatusPending   AllocationStatus = "pending"
	AllocationStatusActive    AllocationStatus = "active"
	AllocationStatusCompleted AllocationStatus = "completed"
	AllocationStatusFailed    AllocationStatus = "failed"
	AllocationStatusPreempted AllocationStatus = "preempted"
)

type QuotaScope string

const (
	QuotaScopeUser    QuotaScope = "user"
	QuotaScopeTeam    QuotaScope = "team"
	QuotaScopeProject QuotaScope = "project"
	QuotaScopeGlobal  QuotaScope = "global"
)

type EnforcementPolicy string

const (
	EnforcementPolicyStrict   EnforcementPolicy = "strict"
	EnforcementPolicyLenient  EnforcementPolicy = "lenient"
	EnforcementPolicyAdaptive EnforcementPolicy = "adaptive"
)

type OveragePolicy string

const (
	OveragePolicyReject  OveragePolicy = "reject"
	OveragePolicyQueue   OveragePolicy = "queue"
	OveragePolicyPreempt OveragePolicy = "preempt"
	OveragePolicyBorrow  OveragePolicy = "borrow"
)

type QuotaStatus string

const (
	QuotaStatusActive    QuotaStatus = "active"
	QuotaStatusSuspended QuotaStatus = "suspended"
	QuotaStatusExceeded  QuotaStatus = "exceeded"
)

type UtilizationTrend string

const (
	UtilizationTrendIncreasing UtilizationTrend = "increasing"
	UtilizationTrendDecreasing UtilizationTrend = "decreasing"
	UtilizationTrendStable     UtilizationTrend = "stable"
	UtilizationTrendVolatile   UtilizationTrend = "volatile"
)

type PerformanceRating string

const (
	PerformanceRatingExcellent PerformanceRating = "excellent"
	PerformanceRatingGood      PerformanceRating = "good"
	PerformanceRatingAverage   PerformanceRating = "average"
	PerformanceRatingPoor      PerformanceRating = "poor"
)

// ResourceUsageSnapshot represents a snapshot of resource usage
type ResourceUsageSnapshot struct {
	Timestamp          time.Time           `json:"timestamp"`
	TotalCapacity      *types.NodeCapacity `json:"total_capacity"`
	TotalAllocated     *types.NodeCapacity `json:"total_allocated"`
	TotalAvailable     *types.NodeCapacity `json:"total_available"`
	OverallUtilization float64             `json:"overall_utilization"`
	NodeUtilizations   map[string]float64  `json:"node_utilizations"`
	PoolUtilizations   map[string]float64  `json:"pool_utilizations"`
}

// ResourceManagerConfig configures the resource manager
type ResourceManagerConfig struct {
	// Allocation settings
	DefaultAllocationPolicy AllocationPolicy
	AllocationTimeout       time.Duration
	PreemptionEnabled       bool

	// Monitoring settings
	MonitoringInterval time.Duration
	HistoryRetention   time.Duration
	MaxHistorySize     int

	// Optimization settings
	EnableOptimization    bool
	OptimizationInterval  time.Duration
	OptimizationThreshold float64

	// Resource limits
	GlobalLimits *ResourceLimits

	// Performance settings
	PerformanceBenchmarkInterval time.Duration
	UtilizationSmoothingFactor   float64
}

// ResourceMetrics tracks resource management performance
type ResourceMetrics struct {
	// Allocation metrics
	TotalAllocations      int64 `json:"total_allocations"`
	SuccessfulAllocations int64 `json:"successful_allocations"`
	FailedAllocations     int64 `json:"failed_allocations"`
	PreemptedAllocations  int64 `json:"preempted_allocations"`

	// Utilization metrics
	AverageUtilization    float64 `json:"average_utilization"`
	PeakUtilization       float64 `json:"peak_utilization"`
	UtilizationEfficiency float64 `json:"utilization_efficiency"`

	// Performance metrics
	AllocationLatency     time.Duration `json:"allocation_latency"`
	ResourceFragmentation float64       `json:"resource_fragmentation"`
	OptimizationGains     float64       `json:"optimization_gains"`

	// Quota metrics
	QuotaViolations  int64              `json:"quota_violations"`
	QuotaUtilization map[string]float64 `json:"quota_utilization"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// ResourceOptimizer optimizes resource allocation and usage
type ResourceOptimizer struct {
	enabled             bool
	threshold           float64
	lastOptimization    time.Time
	optimizationHistory []*OptimizationResult
}

// OptimizationResult represents the result of a resource optimization
type OptimizationResult struct {
	Timestamp         time.Time `json:"timestamp"`
	TriggerReason     string    `json:"trigger_reason"`
	ActionsPerformed  []string  `json:"actions_performed"`
	UtilizationBefore float64   `json:"utilization_before"`
	UtilizationAfter  float64   `json:"utilization_after"`
	Improvement       float64   `json:"improvement"`
	Success           bool      `json:"success"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ResourceManagerConfig) *ResourceManager {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &ResourceManagerConfig{
			DefaultAllocationPolicy:      AllocationPolicyBestFit,
			AllocationTimeout:            30 * time.Second,
			PreemptionEnabled:            true,
			MonitoringInterval:           10 * time.Second,
			HistoryRetention:             24 * time.Hour,
			MaxHistorySize:               1000,
			EnableOptimization:           true,
			OptimizationInterval:         5 * time.Minute,
			OptimizationThreshold:        0.8,
			PerformanceBenchmarkInterval: time.Hour,
			UtilizationSmoothingFactor:   0.1,
			GlobalLimits: &ResourceLimits{
				MaxCPUCores:      1000,
				MaxMemoryBytes:   1024 * 1024 * 1024 * 1024,      // 1TB
				MaxDiskBytes:     10 * 1024 * 1024 * 1024 * 1024, // 10TB
				MaxGPUCores:      100,
				MaxCPUPerNode:    64,
				MaxMemoryPerNode: 512 * 1024 * 1024 * 1024, // 512GB
				MaxCPUPerTask:    32,
				MaxMemoryPerTask: 128 * 1024 * 1024 * 1024, // 128GB
				MaxTaskDuration:  24 * time.Hour,
			},
		}
	}

	rm := &ResourceManager{
		nodeResources: make(map[string]*NodeResourceState),
		resourcePools: make(map[string]*ResourcePool),
		allocations:   make(map[string]*ResourceAllocation),
		quotas:        make(map[string]*ResourceQuota),
		globalLimits:  config.GlobalLimits,
		config:        config,
		metrics: &ResourceMetrics{
			QuotaUtilization: make(map[string]float64),
		},
		usageHistory: make([]*ResourceUsageSnapshot, 0),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize optimizer
	if config.EnableOptimization {
		rm.optimizer = &ResourceOptimizer{
			enabled:             true,
			threshold:           config.OptimizationThreshold,
			optimizationHistory: make([]*OptimizationResult, 0),
		}
	}

	// Start background tasks
	rm.wg.Add(3)
	go rm.monitoringLoop()
	go rm.optimizationLoop()
	go rm.cleanupLoop()

	return rm
}

// RegisterNode registers a node with the resource manager
func (rm *ResourceManager) RegisterNode(nodeID string, capacity *types.NodeCapacity) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Create available capacity (initially same as total)
	available := &types.NodeCapacity{
		NodeID:                  nodeID,
		TotalCPUCores:           capacity.TotalCPUCores,
		TotalMemoryBytes:        capacity.TotalMemoryBytes,
		TotalDiskBytes:          capacity.TotalDiskBytes,
		TotalGPUCores:           capacity.TotalGPUCores,
		TotalGPUMemoryBytes:     capacity.TotalGPUMemoryBytes,
		AvailableCPUCores:       capacity.TotalCPUCores,
		AvailableMemoryBytes:    capacity.TotalMemoryBytes,
		AvailableDiskBytes:      capacity.TotalDiskBytes,
		AvailableGPUCores:       capacity.TotalGPUCores,
		AvailableGPUMemoryBytes: capacity.TotalGPUMemoryBytes,
		SupportedFeatures:       capacity.SupportedFeatures,
		Region:                  capacity.Region,
		Zone:                    capacity.Zone,
		NetworkBandwidth:        capacity.NetworkBandwidth,
		StorageType:             capacity.StorageType,
		ProcessorType:           capacity.ProcessorType,
	}

	// Create allocated capacity (initially zero)
	allocated := &types.NodeCapacity{
		NodeID: nodeID,
	}

	// Create reserved capacity (initially zero)
	reserved := &types.NodeCapacity{
		NodeID: nodeID,
	}

	nodeState := &NodeResourceState{
		NodeID:      nodeID,
		Capacity:    capacity,
		Available:   available,
		Allocated:   allocated,
		Reserved:    reserved,
		Utilization: &ResourceUtilization{},
		Performance: &ResourcePerformance{
			PerformanceRating: PerformanceRatingAverage,
			BenchmarkScores:   make(map[string]float64),
		},
		HealthScore: 1.0,
		Status:      NodeResourceStatusHealthy,
		LastUpdate:  time.Now(),
		Constraints: &ResourceConstraints{},
		Tags:        make(map[string]string),
	}

	rm.nodeResources[nodeID] = nodeState
}

// AllocateResources allocates resources for a task
func (rm *ResourceManager) AllocateResources(taskID string, requirements *types.ResourceRequirement, constraints *ResourceConstraints) (*ResourceAllocation, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Validate requirements against global limits
	if err := rm.validateRequirements(requirements); err != nil {
		return nil, fmt.Errorf("requirements validation failed: %w", err)
	}

	// Find suitable node
	nodeID, err := rm.findSuitableNode(requirements, constraints)
	if err != nil {
		return nil, fmt.Errorf("no suitable node found: %w", err)
	}

	// Create allocation
	allocation := &ResourceAllocation{
		AllocationID:       fmt.Sprintf("alloc_%s_%d", taskID, time.Now().UnixNano()),
		TaskID:             taskID,
		NodeID:             nodeID,
		AllocatedResources: requirements,
		Priority:           types.TaskPriorityNormal, // Default priority
		StartTime:          time.Now(),
		Status:             AllocationStatusPending,
		Metadata:           make(map[string]interface{}),
	}

	// Reserve resources
	if err := rm.reserveResources(nodeID, requirements); err != nil {
		return nil, fmt.Errorf("failed to reserve resources: %w", err)
	}

	// Store allocation
	rm.allocations[allocation.AllocationID] = allocation
	allocation.Status = AllocationStatusActive

	// Update metrics
	rm.metrics.TotalAllocations++
	rm.metrics.SuccessfulAllocations++

	return allocation, nil
}

// validateRequirements validates resource requirements against global limits
func (rm *ResourceManager) validateRequirements(requirements *types.ResourceRequirement) error {
	if rm.globalLimits == nil {
		return nil
	}

	if requirements.CPUCores > rm.globalLimits.MaxCPUPerTask {
		return fmt.Errorf("CPU requirement exceeds limit: %f > %f", requirements.CPUCores, rm.globalLimits.MaxCPUPerTask)
	}

	if requirements.MemoryBytes > rm.globalLimits.MaxMemoryPerTask {
		return fmt.Errorf("memory requirement exceeds limit: %d > %d", requirements.MemoryBytes, rm.globalLimits.MaxMemoryPerTask)
	}

	return nil
}

// findSuitableNode finds a node that can satisfy the resource requirements
func (rm *ResourceManager) findSuitableNode(requirements *types.ResourceRequirement, constraints *ResourceConstraints) (string, error) {
	var candidates []*NodeResourceState

	// Filter nodes based on availability and constraints
	for _, node := range rm.nodeResources {
		if rm.canSatisfyRequirements(node, requirements, constraints) {
			candidates = append(candidates, node)
		}
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no nodes can satisfy requirements")
	}

	// Select best node based on allocation policy
	selectedNode := rm.selectBestNode(candidates, requirements)
	return selectedNode.NodeID, nil
}

// canSatisfyRequirements checks if a node can satisfy resource requirements
func (rm *ResourceManager) canSatisfyRequirements(node *NodeResourceState, requirements *types.ResourceRequirement, constraints *ResourceConstraints) bool {
	// Check basic resource availability
	if node.Available.AvailableCPUCores < requirements.CPUCores ||
		node.Available.AvailableMemoryBytes < requirements.MemoryBytes ||
		node.Available.AvailableDiskBytes < requirements.DiskBytes {
		return false
	}

	// Check GPU requirements if specified
	if requirements.GPUCores > 0 && node.Available.AvailableGPUCores < requirements.GPUCores {
		return false
	}

	if requirements.GPUMemoryBytes > 0 && node.Available.AvailableGPUMemoryBytes < requirements.GPUMemoryBytes {
		return false
	}

	// Check node status
	if node.Status != NodeResourceStatusHealthy {
		return false
	}

	// Check constraints if provided
	if constraints != nil {
		if !rm.satisfiesConstraints(node, constraints) {
			return false
		}
	}

	return true
}

// satisfiesConstraints checks if a node satisfies placement constraints
func (rm *ResourceManager) satisfiesConstraints(node *NodeResourceState, constraints *ResourceConstraints) bool {
	// Check required features
	for _, feature := range constraints.RequiredFeatures {
		found := false
		for _, nodeFeature := range node.Capacity.SupportedFeatures {
			if nodeFeature == feature {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check forbidden features
	for _, feature := range constraints.ForbiddenFeatures {
		for _, nodeFeature := range node.Capacity.SupportedFeatures {
			if nodeFeature == feature {
				return false
			}
		}
	}

	// Check preferred zones
	if len(constraints.PreferredZones) > 0 {
		found := false
		for _, zone := range constraints.PreferredZones {
			if node.Capacity.Zone == zone {
				found = true
				break
			}
		}
		if !found {
			return false // Strict zone preference
		}
	}

	return true
}

// selectBestNode selects the best node from candidates based on allocation policy
func (rm *ResourceManager) selectBestNode(candidates []*NodeResourceState, requirements *types.ResourceRequirement) *NodeResourceState {
	switch rm.config.DefaultAllocationPolicy {
	case AllocationPolicyFirstFit:
		return candidates[0]

	case AllocationPolicyBestFit:
		return rm.selectBestFitNode(candidates, requirements)

	case AllocationPolicyWorstFit:
		return rm.selectWorstFitNode(candidates, requirements)

	case AllocationPolicyBalanced:
		return rm.selectBalancedNode(candidates, requirements)

	default:
		return candidates[0]
	}
}

// selectBestFitNode selects the node with the least remaining resources after allocation
func (rm *ResourceManager) selectBestFitNode(candidates []*NodeResourceState, requirements *types.ResourceRequirement) *NodeResourceState {
	var bestNode *NodeResourceState
	bestScore := math.Inf(1)

	for _, node := range candidates {
		// Calculate remaining resources after allocation
		remainingCPU := node.Available.AvailableCPUCores - requirements.CPUCores
		remainingMemory := float64(node.Available.AvailableMemoryBytes - requirements.MemoryBytes)

		// Calculate fit score (lower is better for best fit)
		score := remainingCPU + remainingMemory/1024/1024/1024 // Normalize memory to GB

		if score < bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode
}

// selectWorstFitNode selects the node with the most remaining resources after allocation
func (rm *ResourceManager) selectWorstFitNode(candidates []*NodeResourceState, requirements *types.ResourceRequirement) *NodeResourceState {
	var bestNode *NodeResourceState
	bestScore := -1.0

	for _, node := range candidates {
		// Calculate remaining resources after allocation
		remainingCPU := node.Available.AvailableCPUCores - requirements.CPUCores
		remainingMemory := float64(node.Available.AvailableMemoryBytes - requirements.MemoryBytes)

		// Calculate fit score (higher is better for worst fit)
		score := remainingCPU + remainingMemory/1024/1024/1024 // Normalize memory to GB

		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode
}

// selectBalancedNode selects the node that maintains the best resource balance
func (rm *ResourceManager) selectBalancedNode(candidates []*NodeResourceState, requirements *types.ResourceRequirement) *NodeResourceState {
	var bestNode *NodeResourceState
	bestScore := math.Inf(1)

	for _, node := range candidates {
		// Calculate utilization after allocation
		cpuUtil := (node.Capacity.TotalCPUCores - node.Available.AvailableCPUCores + requirements.CPUCores) / node.Capacity.TotalCPUCores
		memUtil := float64(node.Capacity.TotalMemoryBytes-node.Available.AvailableMemoryBytes+requirements.MemoryBytes) / float64(node.Capacity.TotalMemoryBytes)

		// Calculate balance score (variance between resource utilizations)
		avgUtil := (cpuUtil + memUtil) / 2.0
		variance := math.Pow(cpuUtil-avgUtil, 2) + math.Pow(memUtil-avgUtil, 2)

		if variance < bestScore {
			bestScore = variance
			bestNode = node
		}
	}

	return bestNode
}

// reserveResources reserves resources on a node
func (rm *ResourceManager) reserveResources(nodeID string, requirements *types.ResourceRequirement) error {
	node, exists := rm.nodeResources[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Update available resources
	node.Available.AvailableCPUCores -= requirements.CPUCores
	node.Available.AvailableMemoryBytes -= requirements.MemoryBytes
	node.Available.AvailableDiskBytes -= requirements.DiskBytes
	node.Available.AvailableGPUCores -= requirements.GPUCores
	node.Available.AvailableGPUMemoryBytes -= requirements.GPUMemoryBytes

	// Update allocated resources
	node.Allocated.AvailableCPUCores += requirements.CPUCores
	node.Allocated.AvailableMemoryBytes += requirements.MemoryBytes
	node.Allocated.AvailableDiskBytes += requirements.DiskBytes
	node.Allocated.AvailableGPUCores += requirements.GPUCores
	node.Allocated.AvailableGPUMemoryBytes += requirements.GPUMemoryBytes

	// Update utilization
	rm.updateNodeUtilization(node)

	return nil
}

// updateNodeUtilization updates utilization metrics for a node
func (rm *ResourceManager) updateNodeUtilization(node *NodeResourceState) {
	if node.Capacity == nil {
		return
	}

	// Calculate utilization percentages
	cpuUtil := (node.Capacity.TotalCPUCores - node.Available.AvailableCPUCores) / node.Capacity.TotalCPUCores
	memUtil := float64(node.Capacity.TotalMemoryBytes-node.Available.AvailableMemoryBytes) / float64(node.Capacity.TotalMemoryBytes)
	diskUtil := float64(node.Capacity.TotalDiskBytes-node.Available.AvailableDiskBytes) / float64(node.Capacity.TotalDiskBytes)

	// Update utilization
	node.Utilization.CPUUtilization = cpuUtil
	node.Utilization.MemoryUtilization = memUtil
	node.Utilization.DiskUtilization = diskUtil

	// Calculate overall utilization (weighted average)
	node.Utilization.OverallUtilization = (cpuUtil*0.4 + memUtil*0.4 + diskUtil*0.2)

	// Update node status based on utilization
	if node.Utilization.OverallUtilization > 0.95 {
		node.Status = NodeResourceStatusOverloaded
	} else if node.Utilization.OverallUtilization > 0.8 {
		node.Status = NodeResourceStatusDegraded
	} else {
		node.Status = NodeResourceStatusHealthy
	}

	node.LastUpdate = time.Now()
}

// monitoringLoop periodically monitors resource usage
func (rm *ResourceManager) monitoringLoop() {
	defer rm.wg.Done()

	ticker := time.NewTicker(rm.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateMetrics()
			rm.createUsageSnapshot()
		}
	}
}

// updateMetrics updates resource management metrics
func (rm *ResourceManager) updateMetrics() {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Calculate average utilization
	totalUtil := 0.0
	nodeCount := 0

	for _, node := range rm.nodeResources {
		if node.Status == NodeResourceStatusHealthy || node.Status == NodeResourceStatusDegraded {
			totalUtil += node.Utilization.OverallUtilization
			nodeCount++
		}
	}

	if nodeCount > 0 {
		rm.metrics.AverageUtilization = totalUtil / float64(nodeCount)
	}

	// Update peak utilization
	if rm.metrics.AverageUtilization > rm.metrics.PeakUtilization {
		rm.metrics.PeakUtilization = rm.metrics.AverageUtilization
	}

	rm.metrics.LastUpdated = time.Now()
}

// createUsageSnapshot creates a snapshot of current resource usage
func (rm *ResourceManager) createUsageSnapshot() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	snapshot := &ResourceUsageSnapshot{
		Timestamp:        time.Now(),
		NodeUtilizations: make(map[string]float64),
		PoolUtilizations: make(map[string]float64),
	}

	// Aggregate capacity and utilization
	totalCapacity := &types.NodeCapacity{}
	totalAllocated := &types.NodeCapacity{}
	totalAvailable := &types.NodeCapacity{}

	for nodeID, node := range rm.nodeResources {
		// Add to totals
		totalCapacity.TotalCPUCores += node.Capacity.TotalCPUCores
		totalCapacity.TotalMemoryBytes += node.Capacity.TotalMemoryBytes
		totalCapacity.TotalDiskBytes += node.Capacity.TotalDiskBytes

		totalAllocated.AvailableCPUCores += node.Allocated.AvailableCPUCores
		totalAllocated.AvailableMemoryBytes += node.Allocated.AvailableMemoryBytes
		totalAllocated.AvailableDiskBytes += node.Allocated.AvailableDiskBytes

		totalAvailable.AvailableCPUCores += node.Available.AvailableCPUCores
		totalAvailable.AvailableMemoryBytes += node.Available.AvailableMemoryBytes
		totalAvailable.AvailableDiskBytes += node.Available.AvailableDiskBytes

		// Record node utilization
		snapshot.NodeUtilizations[nodeID] = node.Utilization.OverallUtilization
	}

	snapshot.TotalCapacity = totalCapacity
	snapshot.TotalAllocated = totalAllocated
	snapshot.TotalAvailable = totalAvailable

	// Calculate overall utilization
	if totalCapacity.TotalCPUCores > 0 {
		snapshot.OverallUtilization = totalAllocated.AvailableCPUCores / totalCapacity.TotalCPUCores
	}

	// Add to history
	rm.usageHistory = append(rm.usageHistory, snapshot)

	// Limit history size
	if len(rm.usageHistory) > rm.config.MaxHistorySize {
		rm.usageHistory = rm.usageHistory[1:]
	}
}

// optimizationLoop periodically optimizes resource allocation
func (rm *ResourceManager) optimizationLoop() {
	defer rm.wg.Done()

	if rm.optimizer == nil || !rm.optimizer.enabled {
		return
	}

	ticker := time.NewTicker(rm.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.performOptimization()
		}
	}
}

// performOptimization performs resource optimization
func (rm *ResourceManager) performOptimization() {
	if rm.optimizer == nil {
		return
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if optimization is needed
	if rm.metrics.AverageUtilization < rm.optimizer.threshold {
		return
	}

	startTime := time.Now()
	utilizationBefore := rm.metrics.AverageUtilization

	// Perform optimization actions
	actions := []string{}

	// Example optimization: consolidate allocations
	if rm.shouldConsolidate() {
		actions = append(actions, "consolidate_allocations")
		// Implementation would go here
	}

	// Example optimization: rebalance loads
	if rm.shouldRebalance() {
		actions = append(actions, "rebalance_loads")
		// Implementation would go here
	}

	// Record optimization result
	result := &OptimizationResult{
		Timestamp:         startTime,
		TriggerReason:     "scheduled_optimization",
		ActionsPerformed:  actions,
		UtilizationBefore: utilizationBefore,
		UtilizationAfter:  rm.metrics.AverageUtilization,
		Success:           len(actions) > 0,
	}

	if result.Success {
		result.Improvement = utilizationBefore - rm.metrics.AverageUtilization
	}

	rm.optimizer.optimizationHistory = append(rm.optimizer.optimizationHistory, result)
	rm.optimizer.lastOptimization = time.Now()
}

// shouldConsolidate determines if allocation consolidation is beneficial
func (rm *ResourceManager) shouldConsolidate() bool {
	// Simple heuristic: consolidate if fragmentation is high
	return rm.metrics.ResourceFragmentation > 0.3
}

// shouldRebalance determines if load rebalancing is beneficial
func (rm *ResourceManager) shouldRebalance() bool {
	// Simple heuristic: rebalance if utilization variance is high
	if len(rm.nodeResources) < 2 {
		return false
	}

	utils := make([]float64, 0, len(rm.nodeResources))
	for _, node := range rm.nodeResources {
		utils = append(utils, node.Utilization.OverallUtilization)
	}

	sort.Float64s(utils)
	return (utils[len(utils)-1] - utils[0]) > 0.3 // 30% difference
}

// cleanupLoop periodically cleans up old data
func (rm *ResourceManager) cleanupLoop() {
	defer rm.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.performCleanup()
		}
	}
}

// performCleanup performs cleanup of old data
func (rm *ResourceManager) performCleanup() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Clean up old allocations
	cutoff := time.Now().Add(-rm.config.HistoryRetention)
	for allocID, alloc := range rm.allocations {
		if alloc.StartTime.Before(cutoff) &&
			(alloc.Status == AllocationStatusCompleted || alloc.Status == AllocationStatusFailed) {
			delete(rm.allocations, allocID)
		}
	}

	// Clean up old usage history
	if len(rm.usageHistory) > rm.config.MaxHistorySize {
		rm.usageHistory = rm.usageHistory[len(rm.usageHistory)-rm.config.MaxHistorySize:]
	}
}

// GetMetrics returns current resource management metrics
func (rm *ResourceManager) GetMetrics() *ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	metrics := *rm.metrics
	return &metrics
}

// GetNodeResourceState returns the resource state of a specific node
func (rm *ResourceManager) GetNodeResourceState(nodeID string) *NodeResourceState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if state, exists := rm.nodeResources[nodeID]; exists {
		// Return a copy
		stateCopy := *state
		return &stateCopy
	}
	return nil
}

// GetUsageHistory returns resource usage history
func (rm *ResourceManager) GetUsageHistory(limit int) []*ResourceUsageSnapshot {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if limit <= 0 || limit > len(rm.usageHistory) {
		limit = len(rm.usageHistory)
	}

	start := len(rm.usageHistory) - limit
	history := make([]*ResourceUsageSnapshot, limit)
	copy(history, rm.usageHistory[start:])

	return history
}

// Close closes the resource manager
func (rm *ResourceManager) Close() error {
	rm.cancel()
	rm.wg.Wait()
	return nil
}
