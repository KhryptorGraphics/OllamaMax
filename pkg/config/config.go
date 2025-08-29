package config

import (
	"time"
)

// SyncConfig holds synchronization configuration
type SyncConfig struct {
	Interval            time.Duration `json:"interval" yaml:"interval"`
	BatchSize          int           `json:"batch_size" yaml:"batch_size"`
	MaxRetries         int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay" yaml:"retry_delay"`
	CompressionEnabled bool          `json:"compression_enabled" yaml:"compression_enabled"`
	EncryptionEnabled  bool          `json:"encryption_enabled" yaml:"encryption_enabled"`
	ConflictResolution string        `json:"conflict_resolution" yaml:"conflict_resolution"`
	PeerTimeout        time.Duration `json:"peer_timeout" yaml:"peer_timeout"`
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Algorithm           string        `json:"algorithm" yaml:"algorithm"`
	MaxConcurrency      int           `json:"max_concurrency" yaml:"max_concurrency"`
	LoadBalanceStrategy string        `json:"load_balance_strategy" yaml:"load_balance_strategy"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	ResourceThreshold   float64       `json:"resource_threshold" yaml:"resource_threshold"`
	PreemptionEnabled   bool          `json:"preemption_enabled" yaml:"preemption_enabled"`
	PriorityClasses     []string      `json:"priority_classes" yaml:"priority_classes"`
	NodeSelector        map[string]string `json:"node_selector" yaml:"node_selector"`
}

// DefaultSyncConfig returns a default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		Interval:            30 * time.Second,
		BatchSize:          100,
		MaxRetries:         3,
		RetryDelay:         5 * time.Second,
		CompressionEnabled: true,
		EncryptionEnabled:  true,
		ConflictResolution: "last-write-wins",
		PeerTimeout:        10 * time.Second,
	}
}

// DefaultSchedulerConfig returns a default scheduler configuration
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		Algorithm:           "round-robin",
		MaxConcurrency:      10,
		LoadBalanceStrategy: "least-loaded",
		HealthCheckInterval: 30 * time.Second,
		ResourceThreshold:   0.8,
		PreemptionEnabled:   true,
		PriorityClasses:     []string{"high", "medium", "low"},
		NodeSelector:        make(map[string]string),
	}
}