package types

import (
	"time"
)

// FaultType represents different types of faults that can occur
type FaultType string

const (
	FaultTypeNodeFailure            FaultType = "node_failure"
	FaultTypeNetworkPartition       FaultType = "network_partition"
	FaultTypeResourceExhaustion     FaultType = "resource_exhaustion"
	FaultTypePerformanceDegradation FaultType = "performance_degradation"
	FaultTypeMemoryLeak             FaultType = "memory_leak"
	FaultTypeDiskFull               FaultType = "disk_full"
	FaultTypeHighLatency            FaultType = "high_latency"
	FaultTypeConnectionLoss         FaultType = "connection_loss"
)

// ResourceMetrics represents resource usage metrics for a node
type ResourceMetrics struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`

	// CPU metrics
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	CPUCores        int     `json:"cpu_cores"`
	LoadAverage     float64 `json:"load_average"`

	// Memory metrics
	MemoryUsedBytes    int64   `json:"memory_used_bytes"`
	MemoryTotalBytes   int64   `json:"memory_total_bytes"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`

	// Storage metrics
	DiskUsedBytes    int64   `json:"disk_used_bytes"`
	DiskTotalBytes   int64   `json:"disk_total_bytes"`
	DiskUsagePercent float64 `json:"disk_usage_percent"`

	// Network metrics
	NetworkInBytes   int64   `json:"network_in_bytes"`
	NetworkOutBytes  int64   `json:"network_out_bytes"`
	NetworkLatencyMs float64 `json:"network_latency_ms"`

	// GPU metrics (if available)
	GPUUsagePercent     float64 `json:"gpu_usage_percent,omitempty"`
	GPUMemoryUsedBytes  int64   `json:"gpu_memory_used_bytes,omitempty"`
	GPUMemoryTotalBytes int64   `json:"gpu_memory_total_bytes,omitempty"`
}

// PerformanceMetrics represents performance metrics for a node
type PerformanceMetrics struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`

	// Throughput metrics
	RequestsPerSecond float64 `json:"requests_per_second"`
	TasksCompleted    int64   `json:"tasks_completed"`
	TasksPerSecond    float64 `json:"tasks_per_second"`

	// Latency metrics
	AverageLatencyMs float64 `json:"average_latency_ms"`
	P95LatencyMs     float64 `json:"p95_latency_ms"`
	P99LatencyMs     float64 `json:"p99_latency_ms"`

	// Error metrics
	ErrorRate   float64 `json:"error_rate"`
	TimeoutRate float64 `json:"timeout_rate"`

	// Quality metrics
	SuccessRate  float64 `json:"success_rate"`
	Availability float64 `json:"availability"`
	Reliability  float64 `json:"reliability"`
}

// HealthMetrics represents health status metrics for a node
type HealthMetrics struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`

	// Overall health
	HealthScore float64      `json:"health_score"` // 0.0 to 1.0
	Status      HealthStatus `json:"status"`

	// Component health
	CPUHealth     float64 `json:"cpu_health"`
	MemoryHealth  float64 `json:"memory_health"`
	DiskHealth    float64 `json:"disk_health"`
	NetworkHealth float64 `json:"network_health"`

	// Service health
	ServiceUptime time.Duration `json:"service_uptime"`
	LastHeartbeat time.Time     `json:"last_heartbeat"`
	ResponseTime  time.Duration `json:"response_time"`

	// Fault indicators
	FaultCount    int           `json:"fault_count"`
	LastFaultTime time.Time     `json:"last_fault_time,omitempty"`
	RecoveryTime  time.Duration `json:"recovery_time,omitempty"`
}

// HealthStatus represents the health status of a node
type HealthStatus string

const (
	HealthStatusHealthy    HealthStatus = "healthy"
	HealthStatusDegraded   HealthStatus = "degraded"
	HealthStatusUnhealthy  HealthStatus = "unhealthy"
	HealthStatusUnknown    HealthStatus = "unknown"
	HealthStatusRecovering HealthStatus = "recovering"
)

// FaultDetection represents fault detection information
type FaultDetection struct {
	FaultID     string        `json:"fault_id"`
	NodeID      string        `json:"node_id"`
	FaultType   FaultType     `json:"fault_type"`
	Severity    FaultSeverity `json:"severity"`
	DetectedAt  time.Time     `json:"detected_at"`
	Description string        `json:"description"`
	Confidence  float64       `json:"confidence"` // 0.0 to 1.0

	// Detection context
	Metrics         map[string]interface{} `json:"metrics"`
	Symptoms        []string               `json:"symptoms"`
	PredictedImpact string                 `json:"predicted_impact"`

	// Resolution
	SuggestedActions      []string      `json:"suggested_actions"`
	AutoResolvable        bool          `json:"auto_resolvable"`
	EstimatedRecoveryTime time.Duration `json:"estimated_recovery_time"`
}

// FaultSeverity represents the severity of a fault
type FaultSeverity string

const (
	FaultSeverityLow      FaultSeverity = "low"
	FaultSeverityMedium   FaultSeverity = "medium"
	FaultSeverityHigh     FaultSeverity = "high"
	FaultSeverityCritical FaultSeverity = "critical"
)

// SystemState represents the overall state of the system
type SystemState struct {
	Timestamp time.Time `json:"timestamp"`

	// Cluster state
	TotalNodes     int `json:"total_nodes"`
	HealthyNodes   int `json:"healthy_nodes"`
	UnhealthyNodes int `json:"unhealthy_nodes"`

	// Resource state
	TotalCPU        int   `json:"total_cpu"`
	AvailableCPU    int   `json:"available_cpu"`
	TotalMemory     int64 `json:"total_memory"`
	AvailableMemory int64 `json:"available_memory"`

	// Performance state
	AverageLatency  float64 `json:"average_latency"`
	TotalThroughput float64 `json:"total_throughput"`
	ErrorRate       float64 `json:"error_rate"`

	// Fault state
	ActiveFaults   int `json:"active_faults"`
	CriticalFaults int `json:"critical_faults"`

	// Load state
	AverageLoad      float64            `json:"average_load"`
	LoadDistribution map[string]float64 `json:"load_distribution"`

	// Additional fields for fault tolerance
	Nodes       []*NodeInfo         `json:"nodes"`
	Resources   *ResourceMetrics    `json:"resources"`
	Performance *PerformanceMetrics `json:"performance"`
	Health      *HealthMetrics      `json:"health"`
	Faults      []*FaultDetection   `json:"faults"`
}

// HealingResult represents the result of a healing operation
type HealingResult struct {
	HealingID string `json:"healing_id"`
	NodeID    string `json:"node_id"`
	FaultID   string `json:"fault_id"`

	// Healing process
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`

	// Result
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	Error        string `json:"error,omitempty"` // Alias for compatibility

	// Actions taken
	ActionsPerformed []string `json:"actions_performed"`
	ActionsTaken     []string `json:"actions_taken"` // Alias for compatibility

	// Additional fields for fault tolerance
	Improvement   float64                `json:"improvement"`
	Metrics       map[string]float64     `json:"metrics"`
	Timestamp     time.Time              `json:"timestamp"`
	ResourcesUsed map[string]interface{} `json:"resources_used"`

	// Impact
	PerformanceImpact float64 `json:"performance_impact"`
	DowntimeSeconds   float64 `json:"downtime_seconds"`

	// Verification
	VerificationPassed bool    `json:"verification_passed"`
	HealthScoreAfter   float64 `json:"health_score_after"`
}

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
	TaskPriorityUrgent   TaskPriority = "urgent"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusRetrying  TaskStatus = "retrying"
)

// ResourceRequirement represents resource requirements for a task
type ResourceRequirement struct {
	CPUCores         float64 `json:"cpu_cores"`
	MemoryBytes      int64   `json:"memory_bytes"`
	DiskBytes        int64   `json:"disk_bytes"`
	NetworkBandwidth int64   `json:"network_bandwidth"`
	GPUCores         int     `json:"gpu_cores,omitempty"`
	GPUMemoryBytes   int64   `json:"gpu_memory_bytes,omitempty"`

	// Constraints
	RequiredFeatures []string `json:"required_features,omitempty"`
	PreferredRegion  string   `json:"preferred_region,omitempty"`
	AntiAffinity     []string `json:"anti_affinity,omitempty"` // Node IDs to avoid
}

// NodeCapacity represents the capacity of a node
type NodeCapacity struct {
	NodeID string `json:"node_id"`

	// Hardware capacity
	TotalCPUCores       float64 `json:"total_cpu_cores"`
	TotalMemoryBytes    int64   `json:"total_memory_bytes"`
	TotalDiskBytes      int64   `json:"total_disk_bytes"`
	TotalGPUCores       int     `json:"total_gpu_cores,omitempty"`
	TotalGPUMemoryBytes int64   `json:"total_gpu_memory_bytes,omitempty"`

	// Available capacity
	AvailableCPUCores       float64 `json:"available_cpu_cores"`
	AvailableMemoryBytes    int64   `json:"available_memory_bytes"`
	AvailableDiskBytes      int64   `json:"available_disk_bytes"`
	AvailableGPUCores       int     `json:"available_gpu_cores,omitempty"`
	AvailableGPUMemoryBytes int64   `json:"available_gpu_memory_bytes,omitempty"`

	// Features and constraints
	SupportedFeatures []string `json:"supported_features"`
	Region            string   `json:"region"`
	Zone              string   `json:"zone"`

	// Performance characteristics
	NetworkBandwidth int64  `json:"network_bandwidth"`
	StorageType      string `json:"storage_type"`
	ProcessorType    string `json:"processor_type"`
}

// LoadBalancingStrategy represents different load balancing strategies
type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin         LoadBalancingStrategy = "round_robin"
	LoadBalancingLeastLoaded        LoadBalancingStrategy = "least_loaded"
	LoadBalancingWeightedRoundRobin LoadBalancingStrategy = "weighted_round_robin"
	LoadBalancingResourceAware      LoadBalancingStrategy = "resource_aware"
	LoadBalancingLatencyBased       LoadBalancingStrategy = "latency_based"
	LoadBalancingCapacityBased      LoadBalancingStrategy = "capacity_based"
	LoadBalancingPredictive         LoadBalancingStrategy = "predictive"
)

// NodeInfo represents information about a node
type NodeInfo struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Status       string                 `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Resources    *ResourceMetrics       `json:"resources"`
	Performance  *PerformanceMetrics    `json:"performance"`
	Health       *HealthMetrics         `json:"health"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}
