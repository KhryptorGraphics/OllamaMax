package types

import (
	"context"
	"time"
)

// Core distributed system types

// NodeID represents a unique identifier for a node in the cluster
type NodeID string

// ClusterID represents a unique identifier for a cluster
type ClusterID string

// TaskID represents a unique identifier for a distributed task
type TaskID string

// ModelID represents a unique identifier for a model
type ModelID string

// Node represents a node in the distributed cluster
type Node struct {
	ID           NodeID                 `json:"id"`
	Address      string                 `json:"address"`
	Status       NodeStatus             `json:"status"`
	Capabilities *NodeCapabilities      `json:"capabilities"`
	Metrics      *NodeMetrics           `json:"metrics"`
	LastSeen     time.Time              `json:"last_seen"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"
	NodeStatusOffline     NodeStatus = "offline"
	NodeStatusDraining    NodeStatus = "draining"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// NodeCapabilities represents the capabilities of a node
type NodeCapabilities struct {
	ModelTypes              []string             `json:"model_types"`
	MaxConcurrentInferences int                  `json:"max_concurrent_inferences"`
	Hardware                HardwareCapabilities `json:"hardware"`
	SupportedFormats        []string             `json:"supported_formats"`
}

// HardwareCapabilities represents hardware capabilities of a node
type HardwareCapabilities struct {
	CPU       int   `json:"cpu_cores"`
	Memory    int64 `json:"memory_bytes"`
	GPU       int   `json:"gpu_count"`
	GPUMemory int64 `json:"gpu_memory_bytes"`
	Storage   int64 `json:"storage_bytes"`
}

// NodeMetrics represents performance metrics for a node
type NodeMetrics struct {
	CPUUsage     float64       `json:"cpu_usage"`
	MemoryUsage  float64       `json:"memory_usage"`
	GPUUsage     float64       `json:"gpu_usage"`
	StorageUsage float64       `json:"storage_usage"`
	ActiveTasks  int           `json:"active_tasks"`
	QueueLength  int           `json:"queue_length"`
	Latency      time.Duration `json:"latency"`
	Throughput   float64       `json:"throughput"`
	LastUpdated  time.Time     `json:"last_updated"`
}

// DistributedTask represents a task that can be distributed across nodes
type DistributedTask struct {
	ID           TaskID                 `json:"id"`
	Type         TaskType               `json:"type"`
	ModelName    string                 `json:"model_name"`
	Input        interface{}            `json:"input"`
	Output       interface{}            `json:"output,omitempty"`
	Status       TaskStatus             `json:"status"`
	AssignedNode NodeID                 `json:"assigned_node,omitempty"`
	Priority     int                    `json:"priority"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Requirements *ResourceRequirements  `json:"requirements,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TaskType represents the type of distributed task
type TaskType string

const (
	TaskTypeInference   TaskType = "inference"
	TaskTypeModelLoad   TaskType = "model_load"
	TaskTypeModelUnload TaskType = "model_unload"
	TaskTypeHealthCheck TaskType = "health_check"
	TaskTypeReplication TaskType = "replication"
)

// TaskStatus represents the status of a distributed task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ResourceRequirements represents resource requirements for a task
type ResourceRequirements struct {
	MinCPU       float64       `json:"min_cpu"`
	MinMemory    int64         `json:"min_memory"`
	RequiresGPU  bool          `json:"requires_gpu"`
	MinGPUMemory int64         `json:"min_gpu_memory,omitempty"`
	MaxLatency   time.Duration `json:"max_latency,omitempty"`
}

// Model represents a distributed model
type Model struct {
	ID        ModelID                `json:"id"`
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Size      int64                  `json:"size"`
	Format    string                 `json:"format"`
	Checksum  string                 `json:"checksum"`
	Replicas  []ModelReplica         `json:"replicas"`
	Status    ModelStatus            `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ModelReplica represents a replica of a model on a specific node
type ModelReplica struct {
	NodeID   NodeID        `json:"node_id"`
	Status   ReplicaStatus `json:"status"`
	LastSync time.Time     `json:"last_sync"`
	Health   ReplicaHealth `json:"health"`
}

// ModelStatus represents the status of a model in the cluster
type ModelStatus string

const (
	ModelStatusAvailable   ModelStatus = "available"
	ModelStatusLoading     ModelStatus = "loading"
	ModelStatusUnavailable ModelStatus = "unavailable"
	ModelStatusError       ModelStatus = "error"
)

// ReplicaStatus represents the status of a model replica
type ReplicaStatus string

const (
	ReplicaStatusHealthy   ReplicaStatus = "healthy"
	ReplicaStatusSyncing   ReplicaStatus = "syncing"
	ReplicaStatusOutOfSync ReplicaStatus = "out_of_sync"
	ReplicaStatusFailed    ReplicaStatus = "failed"
)

// ReplicaHealth represents the health status of a replica
type ReplicaHealth string

const (
	ReplicaHealthGood     ReplicaHealth = "good"
	ReplicaHealthWarning  ReplicaHealth = "warning"
	ReplicaHealthCritical ReplicaHealth = "critical"
)

// ClusterState represents the overall state of the cluster
type ClusterState struct {
	ID          ClusterID              `json:"id"`
	Leader      NodeID                 `json:"leader"`
	Nodes       []Node                 `json:"nodes"`
	Models      []Model                `json:"models"`
	ActiveTasks []DistributedTask      `json:"active_tasks"`
	Status      ClusterStatus          `json:"status"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ClusterStatus represents the status of the cluster
type ClusterStatus string

const (
	ClusterStatusHealthy     ClusterStatus = "healthy"
	ClusterStatusDegraded    ClusterStatus = "degraded"
	ClusterStatusUnavailable ClusterStatus = "unavailable"
)

// Scheduler interface for task scheduling
type Scheduler interface {
	ScheduleTask(ctx context.Context, task *DistributedTask) error
	GetTaskStatus(ctx context.Context, taskID TaskID) (*DistributedTask, error)
	CancelTask(ctx context.Context, taskID TaskID) error
	GetMetrics(ctx context.Context) (*SchedulerMetrics, error)
}

// SchedulerMetrics represents metrics for the scheduler
type SchedulerMetrics struct {
	TotalTasks     int64         `json:"total_tasks"`
	CompletedTasks int64         `json:"completed_tasks"`
	FailedTasks    int64         `json:"failed_tasks"`
	AverageLatency time.Duration `json:"average_latency"`
	QueueLength    int           `json:"queue_length"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// ModelManager interface for model management
type ModelManager interface {
	LoadModel(ctx context.Context, modelName string) error
	UnloadModel(ctx context.Context, modelName string) error
	GetModel(ctx context.Context, modelName string) (*Model, error)
	ListModels(ctx context.Context) ([]Model, error)
	ReplicateModel(ctx context.Context, modelName string, targetNodes []NodeID) error
}

// P2PNode interface for P2P networking
type P2PNode interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetID() NodeID
	GetPeers() []NodeID
	SendMessage(ctx context.Context, target NodeID, message interface{}) error
	BroadcastMessage(ctx context.Context, message interface{}) error
}

// ConsensusEngine interface for distributed consensus
type ConsensusEngine interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsLeader() bool
	GetLeader() NodeID
	ProposeChange(ctx context.Context, change interface{}) error
	GetState() *ClusterState
}

// Legacy compatibility types for API compatibility

// NodeInfo represents simplified node information for API compatibility
type NodeInfo struct {
	ID       string            `json:"id"`
	Address  string            `json:"address"`
	Status   NodeStatus        `json:"status"`
	Capacity NodeCapacity      `json:"capacity"`
	Usage    NodeUsage         `json:"usage"`
	Models   []string          `json:"models"`
	LastSeen time.Time         `json:"last_seen"`
	Metadata map[string]string `json:"metadata"`
}

// NodeCapacity represents simplified node resource capacity
type NodeCapacity struct {
	CPU    int64 `json:"cpu"`
	Memory int64 `json:"memory"`
	GPU    int64 `json:"gpu,omitempty"`
	Disk   int64 `json:"disk"`
}

// NodeUsage represents simplified current node resource usage
type NodeUsage struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	GPU    float64 `json:"gpu,omitempty"`
	Disk   float64 `json:"disk"`
}
