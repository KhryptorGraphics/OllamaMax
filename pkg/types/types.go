package types

import (
	"time"
)

// ResourceRequirement defines resource requirements for model execution
type ResourceRequirement struct {
	CPU       int64  `json:"cpu"`       // CPU cores required
	Memory    int64  `json:"memory"`    // Memory in bytes
	GPU       int64  `json:"gpu"`       // GPU memory in bytes
	Storage   int64  `json:"storage"`   // Storage in bytes
	Bandwidth int64  `json:"bandwidth"` // Network bandwidth required
}

// NodeCapacity defines the capacity of a node
type NodeCapacity struct {
	CPU       int64  `json:"cpu"`       // Available CPU cores
	Memory    int64  `json:"memory"`    // Available memory in bytes
	GPU       int64  `json:"gpu"`       // Available GPU memory in bytes
	Storage   int64  `json:"storage"`   // Available storage in bytes
	Bandwidth int64  `json:"bandwidth"` // Available network bandwidth
}

// NodeInfo represents information about a cluster node
type NodeInfo struct {
	ID           string        `json:"id"`
	Address      string        `json:"address"`
	Status       string        `json:"status"`
	Capacity     NodeCapacity  `json:"capacity"`
	UsedResource ResourceRequirement `json:"used_resource"`
	LastSeen     time.Time     `json:"last_seen"`
	Latency      time.Duration `json:"latency"`
	Load         float64       `json:"load"`
}

// InferenceRequest represents a request for model inference
type InferenceRequest struct {
	ID           string              `json:"id"`
	ModelID      string              `json:"model_id"`
	UserID       string              `json:"user_id"`
	Input        interface{}         `json:"input"`
	Requirements ResourceRequirement `json:"requirements"`
	Priority     int                 `json:"priority"`
	CreatedAt    time.Time           `json:"created_at"`
	Timeout      time.Duration       `json:"timeout"`
}

// InferenceResponse represents the response from model inference
type InferenceResponse struct {
	RequestID string      `json:"request_id"`
	Output    interface{} `json:"output"`
	Error     string      `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	NodeID    string      `json:"node_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// LayerInfo represents information about a model layer
type LayerInfo struct {
	Name        string `json:"name"`
	Parameters  int64  `json:"parameters"`
	MemoryUsage int64  `json:"memory_usage"`
	Size        int64  `json:"size"`
}

// BandwidthUsage represents bandwidth usage statistics
type BandwidthUsage struct {
	Used      int64     `json:"used"`
	Available int64     `json:"available"`
	Peak      int64     `json:"peak"`
	Average   int64     `json:"average"`
	Timestamp time.Time `json:"timestamp"`
}

// AdaptiveBandwidthConfig represents adaptive bandwidth configuration
type AdaptiveBandwidthConfig struct {
	MinBandwidth     int64         `json:"min_bandwidth"`
	MaxBandwidth     int64         `json:"max_bandwidth"`
	AdaptationRate   float64       `json:"adaptation_rate"`
	ThresholdHigh    float64       `json:"threshold_high"`
	ThresholdLow     float64       `json:"threshold_low"`
	MonitorInterval  time.Duration `json:"monitor_interval"`
	Enabled          bool          `json:"enabled"`
}

// DistributedNode represents a node in the distributed system
type DistributedNode struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Status   string    `json:"status"`
	LastSeen time.Time `json:"last_seen"`
}

// Status constants
const (
	StatusActive  = "active"
	StatusFailed  = "failed"
	StatusPending = "pending"
	StatusStopped = "stopped"
)