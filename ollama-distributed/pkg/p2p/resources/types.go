package resources

import (
	"time"
)

// NodeCapabilities represents the capabilities of a P2P node
type NodeCapabilities struct {
	// Compute resources
	CPUCores int        `json:"cpu_cores"`
	Memory   int64      `json:"memory"`
	Storage  int64      `json:"storage"`
	GPUs     []*GPUInfo `json:"gpus"`

	// AI capabilities
	SupportedModels []string `json:"supported_models"`
	ModelFormats    []string `json:"model_formats"`
	Quantizations   []string `json:"quantizations"`

	// Network capabilities
	Bandwidth     int64         `json:"bandwidth"`
	Latency       time.Duration `json:"latency"`
	Reliability   float64       `json:"reliability"`
	PricePerToken float64       `json:"price_per_token"`

	// Node state
	Available  bool      `json:"available"`
	LoadFactor float64   `json:"load_factor"`
	Priority   int       `json:"priority"`
	LastSeen   time.Time `json:"last_seen"`

	// Version information
	Version         string   `json:"version"`
	ProtocolVersion string   `json:"protocol_version"`
	Features        []string `json:"features"`
}

// GPUInfo represents information about a GPU
type GPUInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Memory      int64             `json:"memory"`
	Compute     string            `json:"compute"`
	Available   bool              `json:"available"`
	Utilization float64           `json:"utilization"`
	Properties  map[string]string `json:"properties"`
}

// ResourceMetrics contains real-time resource usage metrics
type ResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage" yaml:"cpu_usage"`       // CPU usage percentage (0-100)
	MemoryUsage int64   `json:"memory_usage" yaml:"memory_usage"` // Memory usage in bytes
	DiskUsage   int64   `json:"disk_usage" yaml:"disk_usage"`     // Disk usage in bytes
	NetworkRx   int64   `json:"network_rx" yaml:"network_rx"`     // Network received bytes/sec
	NetworkTx   int64   `json:"network_tx" yaml:"network_tx"`     // Network transmitted bytes/sec

	// GPU metrics
	GPUUsage  []float64 `json:"gpu_usage" yaml:"gpu_usage"`   // GPU usage percentage per GPU
	GPUMemory []int64   `json:"gpu_memory" yaml:"gpu_memory"` // GPU memory usage per GPU
	GPUTemp   []float64 `json:"gpu_temp" yaml:"gpu_temp"`     // GPU temperature per GPU

	// Performance metrics
	RequestsPerSec float64       `json:"requests_per_sec" yaml:"requests_per_sec"`
	AvgLatency     time.Duration `json:"avg_latency" yaml:"avg_latency"`
	ErrorRate      float64       `json:"error_rate" yaml:"error_rate"`

	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}
