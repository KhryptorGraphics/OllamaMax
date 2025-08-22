package models

import (
	"time"
)

// ReplicaStatus represents the status of a replica
type ReplicaStatus string

const (
	ReplicaStatusHealthy     ReplicaStatus = "healthy"
	ReplicaStatusSyncing     ReplicaStatus = "syncing"
	ReplicaStatusOutOfSync   ReplicaStatus = "out_of_sync"
	ReplicaStatusUnhealthy   ReplicaStatus = "unhealthy"
	ReplicaStatusUnreachable ReplicaStatus = "unreachable"
)

// ReplicaHealth represents the health status of a replica
type ReplicaHealth string

const (
	HealthGood    ReplicaHealth = "good"
	HealthWarning ReplicaHealth = "warning"
	HealthError   ReplicaHealth = "error"
)

// ReplicaInfoV2 contains information about a model replica (enhanced version)
type ReplicaInfoV2 struct {
	ModelName    string            `json:"model_name"`
	PeerID       string            `json:"peer_id"`
	Status       ReplicaStatus     `json:"status"`
	LastSync     time.Time         `json:"last_sync"`
	SyncAttempts int               `json:"sync_attempts"`
	Health       ReplicaHealth     `json:"health"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ReplicationPolicyV2 defines how a model should be replicated (enhanced version)
type ReplicationPolicyV2 struct {
	ModelName         string            `json:"model_name"`
	MinReplicas       int               `json:"min_replicas"`
	MaxReplicas       int               `json:"max_replicas"`
	PreferredPeers    []string          `json:"preferred_peers"`
	ExcludedPeers     []string          `json:"excluded_peers"`
	ReplicationFactor int               `json:"replication_factor"`
	SyncInterval      time.Duration     `json:"sync_interval"`
	Priority          int               `json:"priority"`
	Constraints       map[string]string `json:"constraints"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ReplicationTaskV2 represents a replication task (enhanced version)
type ReplicationTaskV2 struct {
	ID          string                 `json:"id"`
	Type        ReplicationTaskType    `json:"type"`
	ModelName   string                 `json:"model_name"`
	SourcePeer  string                 `json:"source_peer"`
	TargetPeer  string                 `json:"target_peer"`
	Status      ReplicationTaskStatus  `json:"status"`
	Progress    float64                `json:"progress"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// ReplicationTaskType represents the type of replication task
type ReplicationTaskType string

const (
	TaskTypeSyncV2   ReplicationTaskType = "sync"
	TaskTypeCopyV2   ReplicationTaskType = "copy"
	TaskTypeDeleteV2 ReplicationTaskType = "delete"
	TaskTypeVerifyV2 ReplicationTaskType = "verify"
)

// ReplicationTaskStatus represents the status of a replication task
type ReplicationTaskStatus string

const (
	TaskStatusPending    ReplicationTaskStatus = "pending"
	TaskStatusRunning    ReplicationTaskStatus = "running"
	TaskStatusCompleted  ReplicationTaskStatus = "completed"
	TaskStatusFailed     ReplicationTaskStatus = "failed"
	TaskStatusCancelled  ReplicationTaskStatus = "cancelled"
)

// ReplicationWorkerV2 manages replication tasks (enhanced version)
type ReplicationWorkerV2 struct {
	ID       string
	Busy     bool
	LastTask time.Time
}

// SyncManagerV2 manages synchronization tasks (enhanced version)
type SyncManagerV2 struct {
	Workers []string
	Queue   chan *ReplicationTaskV2
}