package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MonitoringSystem provides comprehensive system monitoring
type MonitoringSystem struct {
	networkMonitor     *NetworkMonitor
	performanceMonitor *PerformanceMonitor
	healthMonitor      *HealthMonitor
	config             *MonitoringConfig
	mu                 sync.RWMutex
	running            bool
	ctx                context.Context
	cancel             context.CancelFunc
}

// MonitoringConfig configures the monitoring system
type MonitoringConfig struct {
	EnableNetworkMonitoring     bool             `json:"enable_network_monitoring"`
	EnablePerformanceMonitoring bool             `json:"enable_performance_monitoring"`
	EnableHealthMonitoring      bool             `json:"enable_health_monitoring"`
	MetricsInterval             time.Duration    `json:"metrics_interval"`
	RetentionPeriod             time.Duration    `json:"retention_period"`
	AlertThresholds             *AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds defines thresholds for alerts
type AlertThresholds struct {
	CPUUsage       float64       `json:"cpu_usage"`
	MemoryUsage    float64       `json:"memory_usage"`
	DiskUsage      float64       `json:"disk_usage"`
	NetworkLatency time.Duration `json:"network_latency"`
	ErrorRate      float64       `json:"error_rate"`
}

// PerformanceMonitor monitors system performance
type PerformanceMonitor struct {
	config  *PerformanceConfig
	metrics *PerformanceMetrics
	mu      sync.RWMutex
}

// PerformanceConfig configures performance monitoring
type PerformanceConfig struct {
	SampleInterval time.Duration `json:"sample_interval"`
	BufferSize     int           `json:"buffer_size"`
}

// PerformanceMetrics holds performance metrics
type PerformanceMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIO   int64     `json:"network_io"`
	DiskIO      int64     `json:"disk_io"`
	Timestamp   time.Time `json:"timestamp"`
}

// HealthMonitor monitors system health
type HealthMonitor struct {
	config *HealthConfig
	status HealthStatus
	mu     sync.RWMutex
}

// Note: HealthConfig, HealthCheck, and HealthStatus types are defined in health_checker.go

// NewMonitoringSystem creates a new monitoring system
func NewMonitoringSystem(config *MonitoringConfig) *MonitoringSystem {
	if config == nil {
		config = &MonitoringConfig{
			EnableNetworkMonitoring:     true,
			EnablePerformanceMonitoring: true,
			EnableHealthMonitoring:      true,
			MetricsInterval:             30 * time.Second,
			RetentionPeriod:             24 * time.Hour,
			AlertThresholds: &AlertThresholds{
				CPUUsage:       80.0,
				MemoryUsage:    85.0,
				DiskUsage:      90.0,
				NetworkLatency: 500 * time.Millisecond,
				ErrorRate:      5.0,
			},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	ms := &MonitoringSystem{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize monitors
	if config.EnableNetworkMonitoring {
		ms.networkMonitor = NewNetworkMonitor(nil)
	}

	if config.EnablePerformanceMonitoring {
		ms.performanceMonitor = NewPerformanceMonitor(&PerformanceConfig{
			SampleInterval: config.MetricsInterval,
			BufferSize:     1000,
		})
	}

	if config.EnableHealthMonitoring {
		ms.healthMonitor = NewHealthMonitor(&HealthConfig{
			CheckInterval: config.MetricsInterval,
			Timeout:       10 * time.Second,
		})
	}

	return ms
}

// Start starts the monitoring system
func (ms *MonitoringSystem) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.running {
		return nil
	}

	// Start individual monitors
	if ms.networkMonitor != nil {
		if err := ms.networkMonitor.Start(); err != nil {
			return err
		}
	}

	if ms.performanceMonitor != nil {
		if err := ms.performanceMonitor.Start(); err != nil {
			return err
		}
	}

	if ms.healthMonitor != nil {
		if err := ms.healthMonitor.Start(); err != nil {
			return err
		}
	}

	ms.running = true
	log.Info().Msg("Monitoring system started")
	return nil
}

// Stop stops the monitoring system
func (ms *MonitoringSystem) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.running {
		return nil
	}

	// Stop individual monitors
	if ms.networkMonitor != nil {
		ms.networkMonitor.Stop()
	}

	if ms.performanceMonitor != nil {
		ms.performanceMonitor.Stop()
	}

	if ms.healthMonitor != nil {
		ms.healthMonitor.Stop()
	}

	ms.cancel()
	ms.running = false
	log.Info().Msg("Monitoring system stopped")
	return nil
}

// GetNetworkMonitor returns the network monitor
func (ms *MonitoringSystem) GetNetworkMonitor() *NetworkMonitor {
	return ms.networkMonitor
}

// GetPerformanceMonitor returns the performance monitor
func (ms *MonitoringSystem) GetPerformanceMonitor() *PerformanceMonitor {
	return ms.performanceMonitor
}

// GetHealthMonitor returns the health monitor
func (ms *MonitoringSystem) GetHealthMonitor() *HealthMonitor {
	return ms.healthMonitor
}

// GetOverallHealth returns the overall system health
func (ms *MonitoringSystem) GetOverallHealth() HealthStatus {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.running {
		return HealthStatusUnknown
	}

	// Check network health
	if ms.networkMonitor != nil && !ms.networkMonitor.IsHealthy() {
		return HealthStatusDegraded
	}

	// Check performance health
	if ms.performanceMonitor != nil {
		metrics := ms.performanceMonitor.GetMetrics()
		if metrics.CPUUsage > ms.config.AlertThresholds.CPUUsage ||
			metrics.MemoryUsage > ms.config.AlertThresholds.MemoryUsage ||
			metrics.DiskUsage > ms.config.AlertThresholds.DiskUsage {
			return HealthStatusDegraded
		}
	}

	// Check component health
	if ms.healthMonitor != nil {
		status := ms.healthMonitor.GetStatus()
		if status != HealthStatusHealthy {
			return status
		}
	}

	return HealthStatusHealthy
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *PerformanceConfig) *PerformanceMonitor {
	if config == nil {
		config = &PerformanceConfig{
			SampleInterval: 30 * time.Second,
			BufferSize:     1000,
		}
	}

	return &PerformanceMonitor{
		config: config,
		metrics: &PerformanceMetrics{
			Timestamp: time.Now(),
		},
	}
}

// Start starts the performance monitor
func (pm *PerformanceMonitor) Start() error {
	// Implementation would start performance monitoring
	log.Info().Msg("Performance monitor started")
	return nil
}

// Stop stops the performance monitor
func (pm *PerformanceMonitor) Stop() error {
	// Implementation would stop performance monitoring
	log.Info().Msg("Performance monitor stopped")
	return nil
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMonitor) GetMetrics() *PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &PerformanceMetrics{
		CPUUsage:    pm.metrics.CPUUsage,
		MemoryUsage: pm.metrics.MemoryUsage,
		DiskUsage:   pm.metrics.DiskUsage,
		NetworkIO:   pm.metrics.NetworkIO,
		DiskIO:      pm.metrics.DiskIO,
		Timestamp:   pm.metrics.Timestamp,
	}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(config *HealthConfig) *HealthMonitor {
	if config == nil {
		config = &HealthConfig{
			CheckInterval: 30 * time.Second,
			Timeout:       10 * time.Second,
		}
	}

	return &HealthMonitor{
		config: config,
		status: HealthStatusHealthy,
	}
}

// Start starts the health monitor
func (hm *HealthMonitor) Start() error {
	// Implementation would start health monitoring
	log.Info().Msg("Health monitor started")
	return nil
}

// Stop stops the health monitor
func (hm *HealthMonitor) Stop() error {
	// Implementation would stop health monitoring
	log.Info().Msg("Health monitor stopped")
	return nil
}

// GetStatus returns current health status
func (hm *HealthMonitor) GetStatus() HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.status
}

// GetHealth returns current health status (deprecated - use GetStatus)
func (hm *HealthMonitor) GetHealth() HealthStatus {
	return hm.GetStatus()
}
