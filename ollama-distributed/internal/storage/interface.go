package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Storage defines the interface for distributed storage operations
type Storage interface {
	// Core storage operations
	Store(ctx context.Context, key string, data io.Reader, metadata *ObjectMetadata) error
	Retrieve(ctx context.Context, key string) (io.ReadCloser, *ObjectMetadata, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// Metadata operations
	GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error)
	SetMetadata(ctx context.Context, key string, metadata *ObjectMetadata) error
	UpdateMetadata(ctx context.Context, key string, updates map[string]interface{}) error
	
	// Batch operations
	BatchStore(ctx context.Context, operations []BatchStoreOperation) error
	BatchDelete(ctx context.Context, keys []string) error
	
	// Listing and iteration
	List(ctx context.Context, prefix string, options *ListOptions) (*ListResult, error)
	ListKeys(ctx context.Context, prefix string) ([]string, error)
	
	// Health and monitoring
	HealthCheck(ctx context.Context) (*HealthStatus, error)
	GetStats(ctx context.Context) (*StorageStats, error)
	
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Close() error
}

// DistributedStorage extends Storage with distributed-specific operations
type DistributedStorage interface {
	Storage
	
	// Replication operations
	Replicate(ctx context.Context, key string, targetNodes []string) error
	GetReplicationStatus(ctx context.Context, key string) (*ReplicationStatus, error)
	SetReplicationPolicy(ctx context.Context, key string, policy *ReplicationPolicy) error
	
	// Consensus and coordination
	ProposeWrite(ctx context.Context, key string, data io.Reader, metadata *ObjectMetadata) error
	ProposeDelete(ctx context.Context, key string) error
	GetConsensusState(ctx context.Context) (*ConsensusState, error)
	
	// Node management
	AddNode(ctx context.Context, nodeID string, nodeInfo *NodeInfo) error
	RemoveNode(ctx context.Context, nodeID string) error
	GetNodes(ctx context.Context) ([]*NodeInfo, error)
	
	// Distributed coordination
	AcquireLock(ctx context.Context, lockID string, timeout time.Duration) (Lock, error)
	GetDistributedMetrics(ctx context.Context) (*DistributedMetrics, error)
}

// ModelStorage defines storage operations specific to AI models
type ModelStorage interface {
	Storage
	
	// Model-specific operations
	StoreModel(ctx context.Context, modelID string, modelData io.Reader, config *ModelConfig) error
	RetrieveModel(ctx context.Context, modelID string) (io.ReadCloser, *ModelConfig, error)
	DeleteModel(ctx context.Context, modelID string) error
	
	// Model metadata and versioning
	GetModelVersions(ctx context.Context, modelID string) ([]*ModelVersion, error)
	GetModelConfig(ctx context.Context, modelID string) (*ModelConfig, error)
	SetModelConfig(ctx context.Context, modelID string, config *ModelConfig) error
	
	// Model lifecycle
	ArchiveModel(ctx context.Context, modelID string) error
	RestoreModel(ctx context.Context, modelID string) error
	GetArchivedModels(ctx context.Context) ([]*ArchivedModel, error)
}

// BackupStorage defines backup and recovery operations
type BackupStorage interface {
	// Backup operations
	CreateBackup(ctx context.Context, backupID string, options *BackupOptions) error
	RestoreBackup(ctx context.Context, backupID string, options *RestoreOptions) error
	DeleteBackup(ctx context.Context, backupID string) error
	
	// Backup management
	ListBackups(ctx context.Context) ([]*BackupInfo, error)
	GetBackupInfo(ctx context.Context, backupID string) (*BackupInfo, error)
	VerifyBackup(ctx context.Context, backupID string) (*BackupVerification, error)
	
	// Incremental backup support
	CreateIncrementalBackup(ctx context.Context, backupID string, baseBackupID string, options *BackupOptions) error
	GetBackupChain(ctx context.Context, backupID string) ([]*BackupInfo, error)
}

// Lock represents a distributed lock
type Lock interface {
	Release() error
	Renew(timeout time.Duration) error
	IsHeld() bool
	GetOwner() string
	GetExpiration() time.Time
}

// ObjectMetadata contains metadata for stored objects
type ObjectMetadata struct {
	Key         string                 `json:"key"`
	Size        int64                  `json:"size"`
	ContentType string                 `json:"content_type"`
	Hash        string                 `json:"hash"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
	Version     string                 `json:"version"`
	Attributes  map[string]interface{} `json:"attributes"`
	
	// Replication metadata
	ReplicationPolicy *ReplicationPolicy `json:"replication_policy,omitempty"`
	ReplicationNodes  []string           `json:"replication_nodes,omitempty"`
	
	// Model-specific metadata
	ModelInfo *ModelMetadata `json:"model_info,omitempty"`
}

// ModelMetadata contains AI model specific metadata
type ModelMetadata struct {
	ModelID     string            `json:"model_id"`
	ModelType   string            `json:"model_type"`
	Format      string            `json:"format"`
	Parameters  map[string]string `json:"parameters"`
	Tags        []string          `json:"tags"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
}

// BatchStoreOperation represents a batch store operation
type BatchStoreOperation struct {
	Key      string          `json:"key"`
	Data     io.Reader       `json:"-"`
	Metadata *ObjectMetadata `json:"metadata"`
}

// ListOptions contains options for listing operations
type ListOptions struct {
	Limit        int    `json:"limit"`
	Continuation string `json:"continuation"`
	Recursive    bool   `json:"recursive"`
	IncludeSize  bool   `json:"include_size"`
	SortBy       string `json:"sort_by"`
	SortOrder    string `json:"sort_order"`
}

// ListResult contains the result of a list operation
type ListResult struct {
	Items        []*ObjectMetadata `json:"items"`
	Continuation string            `json:"continuation"`
	Total        int64             `json:"total"`
	HasMore      bool              `json:"has_more"`
}

// HealthStatus represents the health status of storage
type HealthStatus struct {
	Status     string                 `json:"status"`
	Healthy    bool                   `json:"healthy"`
	LastCheck  time.Time              `json:"last_check"`
	Checks     map[string]CheckResult `json:"checks"`
	NodeHealth map[string]NodeHealth  `json:"node_health,omitempty"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Latency int64     `json:"latency_ms"`
	Time    time.Time `json:"time"`
}

// NodeHealth represents the health of a storage node
type NodeHealth struct {
	NodeID      string    `json:"node_id"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	Latency     int64     `json:"latency_ms"`
	StorageUsed int64     `json:"storage_used"`
	StorageTotal int64    `json:"storage_total"`
}

// StorageStats contains storage statistics
type StorageStats struct {
	TotalObjects    int64             `json:"total_objects"`
	TotalSize       int64             `json:"total_size"`
	UsedSpace       int64             `json:"used_space"`
	AvailableSpace  int64             `json:"available_space"`
	OperationCounts map[string]int64  `json:"operation_counts"`
	Performance     *PerformanceStats `json:"performance"`
	Replication     *ReplicationStats `json:"replication,omitempty"`
}

// PerformanceStats contains performance metrics
type PerformanceStats struct {
	ReadLatency   *LatencyStats `json:"read_latency"`
	WriteLatency  *LatencyStats `json:"write_latency"`
	DeleteLatency *LatencyStats `json:"delete_latency"`
	Throughput    *Throughput   `json:"throughput"`
}

// LatencyStats contains latency statistics
type LatencyStats struct {
	Min     int64   `json:"min_ms"`
	Max     int64   `json:"max_ms"`
	Mean    float64 `json:"mean_ms"`
	Median  int64   `json:"median_ms"`
	P95     int64   `json:"p95_ms"`
	P99     int64   `json:"p99_ms"`
	Samples int64   `json:"samples"`
}

// Throughput contains throughput metrics
type Throughput struct {
	ReadOpsPerSec   float64 `json:"read_ops_per_sec"`
	WriteOpsPerSec  float64 `json:"write_ops_per_sec"`
	DeleteOpsPerSec float64 `json:"delete_ops_per_sec"`
	ReadBytesPerSec int64   `json:"read_bytes_per_sec"`
	WriteBytesPerSec int64  `json:"write_bytes_per_sec"`
}

// ReplicationPolicy defines how objects should be replicated
type ReplicationPolicy struct {
	MinReplicas      int                    `json:"min_replicas"`
	MaxReplicas      int                    `json:"max_replicas"`
	PreferredNodes   []string               `json:"preferred_nodes"`
	ExcludedNodes    []string               `json:"excluded_nodes"`
	ConsistencyLevel string                 `json:"consistency_level"` // strong, eventual, weak
	Strategy         string                 `json:"strategy"`          // eager, lazy, on_demand
	Priority         int                    `json:"priority"`
	Constraints      map[string]interface{} `json:"constraints"`
}

// ReplicationStatus represents the status of object replication
type ReplicationStatus struct {
	Key              string            `json:"key"`
	Policy           *ReplicationPolicy `json:"policy"`
	CurrentReplicas  int               `json:"current_replicas"`
	HealthyReplicas  int               `json:"healthy_replicas"`
	ReplicaNodes     []string          `json:"replica_nodes"`
	SyncStatus       map[string]string `json:"sync_status"`
	LastSync         time.Time         `json:"last_sync"`
	ConsistencyCheck time.Time         `json:"consistency_check"`
}

// ReplicationStats contains replication statistics
type ReplicationStats struct {
	TotalReplicas     int64             `json:"total_replicas"`
	HealthyReplicas   int64             `json:"healthy_replicas"`
	OutOfSyncReplicas int64             `json:"out_of_sync_replicas"`
	ReplicationLag    map[string]int64  `json:"replication_lag_ms"`
	SyncOperations    *SyncStats        `json:"sync_operations"`
}

// SyncStats contains synchronization statistics
type SyncStats struct {
	SuccessfulSyncs int64 `json:"successful_syncs"`
	FailedSyncs     int64 `json:"failed_syncs"`
	PendingSyncs    int64 `json:"pending_syncs"`
	AverageSyncTime int64 `json:"average_sync_time_ms"`
}

// ConsensusState represents the state of distributed consensus
type ConsensusState struct {
	LeaderID       string            `json:"leader_id"`
	Term           int64             `json:"term"`
	CommitIndex    int64             `json:"commit_index"`
	LastApplied    int64             `json:"last_applied"`
	Nodes          map[string]string `json:"nodes"` // nodeID -> status
	QuorumSize     int               `json:"quorum_size"`
	IsHealthy      bool              `json:"is_healthy"`
	LastHeartbeat  time.Time         `json:"last_heartbeat"`
}

// NodeInfo contains information about a storage node
type NodeInfo struct {
	NodeID       string                 `json:"node_id"`
	Address      string                 `json:"address"`
	Port         int                    `json:"port"`
	Region       string                 `json:"region"`
	Zone         string                 `json:"zone"`
	Capacity     int64                  `json:"capacity"`
	Used         int64                  `json:"used"`
	Available    int64                  `json:"available"`
	Status       string                 `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
	JoinedAt     time.Time              `json:"joined_at"`
	LastSeen     time.Time              `json:"last_seen"`
}

// DistributedMetrics contains distributed storage metrics
type DistributedMetrics struct {
	ClusterSize       int                    `json:"cluster_size"`
	HealthyNodes      int                    `json:"healthy_nodes"`
	TotalCapacity     int64                  `json:"total_capacity"`
	UsedCapacity      int64                  `json:"used_capacity"`
	ReplicationFactor float64                `json:"replication_factor"`
	DataDistribution  map[string]int64       `json:"data_distribution"`
	NetworkMetrics    *NetworkMetrics        `json:"network_metrics"`
	ConsensusMetrics  *ConsensusMetrics      `json:"consensus_metrics"`
}

// NetworkMetrics contains network-related metrics
type NetworkMetrics struct {
	AverageLatency    int64            `json:"average_latency_ms"`
	TotalBandwidth    int64            `json:"total_bandwidth_bps"`
	UsedBandwidth     int64            `json:"used_bandwidth_bps"`
	NetworkErrors     int64            `json:"network_errors"`
	ConnectionCounts  map[string]int64 `json:"connection_counts"`
}

// ConsensusMetrics contains consensus-related metrics
type ConsensusMetrics struct {
	LeaderElections    int64 `json:"leader_elections"`
	TermsCompleted     int64 `json:"terms_completed"`
	ProposalsSubmitted int64 `json:"proposals_submitted"`
	ProposalsCommitted int64 `json:"proposals_committed"`
	ConsensuLatency    int64 `json:"consensus_latency_ms"`
}

// ModelConfig contains configuration for AI models
type ModelConfig struct {
	ModelID      string                 `json:"model_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Format       string                 `json:"format"`
	Size         int64                  `json:"size"`
	Hash         string                 `json:"hash"`
	Parameters   map[string]interface{} `json:"parameters"`
	Metadata     map[string]interface{} `json:"metadata"`
	Dependencies []string               `json:"dependencies"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ModelVersion represents a version of a model
type ModelVersion struct {
	Version   string    `json:"version"`
	Hash      string    `json:"hash"`
	Size      int64     `json:"size"`
	Changes   []string  `json:"changes"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// ArchivedModel represents an archived model
type ArchivedModel struct {
	ModelID     string    `json:"model_id"`
	ArchiveID   string    `json:"archive_id"`
	ArchivePath string    `json:"archive_path"`
	OriginalSize int64    `json:"original_size"`
	CompressedSize int64  `json:"compressed_size"`
	ArchivedAt  time.Time `json:"archived_at"`
	Reason      string    `json:"reason"`
}

// BackupOptions contains options for backup operations
type BackupOptions struct {
	Compression    string            `json:"compression"`
	Encryption     bool              `json:"encryption"`
	IncludeIndex   bool              `json:"include_index"`
	IncludeMetadata bool             `json:"include_metadata"`
	Filters        []string          `json:"filters"`
	ChunkSize      int64             `json:"chunk_size"`
	Parallel       bool              `json:"parallel"`
	Metadata       map[string]string `json:"metadata"`
}

// RestoreOptions contains options for restore operations
type RestoreOptions struct {
	OverwriteExisting bool              `json:"overwrite_existing"`
	VerifyIntegrity   bool              `json:"verify_integrity"`
	RestoreMetadata   bool              `json:"restore_metadata"`
	RestoreIndex      bool              `json:"restore_index"`
	Filters           []string          `json:"filters"`
	TargetPath        string            `json:"target_path"`
	Parallel          bool              `json:"parallel"`
	Metadata          map[string]string `json:"metadata"`
}

// BackupInfo contains information about a backup
type BackupInfo struct {
	BackupID      string    `json:"backup_id"`
	BaseBackupID  string    `json:"base_backup_id,omitempty"`
	Type          string    `json:"type"` // full, incremental
	Status        string    `json:"status"`
	Size          int64     `json:"size"`
	CompressedSize int64    `json:"compressed_size"`
	ObjectCount   int64     `json:"object_count"`
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   time.Time `json:"completed_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Metadata      map[string]string `json:"metadata"`
	Checksums     map[string]string `json:"checksums"`
}

// BackupVerification contains the result of backup verification
type BackupVerification struct {
	BackupID    string            `json:"backup_id"`
	IsValid     bool              `json:"is_valid"`
	Errors      []string          `json:"errors"`
	Warnings    []string          `json:"warnings"`
	CheckedAt   time.Time         `json:"checked_at"`
	Checksums   map[string]string `json:"checksums"`
	ObjectCount int64             `json:"object_count"`
	TotalSize   int64             `json:"total_size"`
}

// StorageError represents storage operation errors
type StorageError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Operation string `json:"operation"`
	Key       string `json:"key,omitempty"`
	Cause     error  `json:"-"`
}

func (e *StorageError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *StorageError) Unwrap() error {
	return e.Cause
}

// Common error codes
const (
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeAlreadyExists    = "ALREADY_EXISTS"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
	ErrCodeQuotaExceeded    = "QUOTA_EXCEEDED"
	ErrCodeInvalidArgument  = "INVALID_ARGUMENT"
	ErrCodeInternal         = "INTERNAL_ERROR"
	ErrCodeUnavailable      = "UNAVAILABLE"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeCorrupted        = "CORRUPTED"
	ErrCodeConsistency      = "CONSISTENCY_ERROR"
)