package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

func TestNewMonitor(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: 30 * time.Second,
			DataRetention:  24 * time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)
	require.NotNil(t, monitor)

	assert.Equal(t, cfg.Monitoring.MetricsPort, monitor.config.MetricsPort)
	assert.Equal(t, cfg.Monitoring.UpdateInterval, monitor.config.UpdateInterval)
}

func TestSystemMetrics_Collection(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for metrics collection
	time.Sleep(2 * time.Second)

	// Get system metrics
	metrics := monitor.GetSystemMetrics()
	assert.NotNil(t, metrics)

	// Verify metrics are being collected
	assert.GreaterOrEqual(t, metrics.CPU.UsagePercent, 0.0)
	assert.LessOrEqual(t, metrics.CPU.UsagePercent, 100.0)
	assert.Greater(t, metrics.Memory.TotalBytes, uint64(0))
	assert.GreaterOrEqual(t, metrics.Memory.UsedBytes, uint64(0))
	assert.LessOrEqual(t, metrics.Memory.UsedBytes, metrics.Memory.TotalBytes)
	assert.Greater(t, metrics.Disk.TotalBytes, uint64(0))
	assert.GreaterOrEqual(t, metrics.Disk.UsedBytes, uint64(0))
	assert.LessOrEqual(t, metrics.Disk.UsedBytes, metrics.Disk.TotalBytes)
}

func TestNetworkMetrics_Collection(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for metrics collection
	time.Sleep(1500 * time.Millisecond)

	// Get network metrics
	metrics := monitor.GetNetworkMetrics()
	assert.NotNil(t, metrics)

	// Network metrics should be initialized
	assert.GreaterOrEqual(t, metrics.BytesReceived, uint64(0))
	assert.GreaterOrEqual(t, metrics.BytesSent, uint64(0))
	assert.GreaterOrEqual(t, metrics.PacketsReceived, uint64(0))
	assert.GreaterOrEqual(t, metrics.PacketsSent, uint64(0))
}

func TestP2PMetrics_Collection(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	// Simulate P2P metrics updates
	monitor.UpdateP2PMetrics(P2PMetrics{
		ConnectedPeers:    5,
		TotalConnections:  10,
		MessagesReceived:  100,
		MessagesSent:      85,
		DataReceived:      1024 * 1024, // 1MB
		DataSent:          768 * 1024,  // 768KB
	})

	metrics := monitor.GetP2PMetrics()
	assert.NotNil(t, metrics)

	assert.Equal(t, 5, metrics.ConnectedPeers)
	assert.Equal(t, uint64(10), metrics.TotalConnections)
	assert.Equal(t, uint64(100), metrics.MessagesReceived)
	assert.Equal(t, uint64(85), metrics.MessagesSent)
	assert.Equal(t, uint64(1024*1024), metrics.DataReceived)
	assert.Equal(t, uint64(768*1024), metrics.DataSent)
}

func TestJobMetrics_Collection(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	// Simulate job metrics updates
	monitor.UpdateJobMetrics(JobMetrics{
		TotalJobs:      50,
		CompletedJobs:  45,
		FailedJobs:     3,
		RunningJobs:    2,
		QueuedJobs:     0,
		AverageLatency: 1250 * time.Millisecond,
	})

	metrics := monitor.GetJobMetrics()
	assert.NotNil(t, metrics)

	assert.Equal(t, 50, metrics.TotalJobs)
	assert.Equal(t, 45, metrics.CompletedJobs)
	assert.Equal(t, 3, metrics.FailedJobs)
	assert.Equal(t, 2, metrics.RunningJobs)
	assert.Equal(t, 0, metrics.QueuedJobs)
	assert.Equal(t, 1250*time.Millisecond, metrics.AverageLatency)
}

func TestHealthCheck_Endpoints(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for startup
	time.Sleep(500 * time.Millisecond)

	// Test health endpoint
	health := monitor.GetHealth()
	assert.NotNil(t, health)
	assert.True(t, health.Healthy)
	assert.Equal(t, "healthy", health.Status)
	assert.NotEmpty(t, health.Timestamp)
	assert.NotNil(t, health.Checks)
}

func TestAlerts_Threshold(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
			AlertThresholds: &config.AlertThresholds{
				CPUUsagePercent:    80.0,
				MemoryUsagePercent: 85.0,
				DiskUsagePercent:   90.0,
				ErrorRate:          5.0,
			},
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	// Test CPU alert threshold
	monitor.checkAlerts(SystemMetrics{
		CPU: CPUMetrics{
			UsagePercent: 85.0, // Above threshold
		},
		Memory: MemoryMetrics{
			TotalBytes: 8 * 1024 * 1024 * 1024, // 8GB
			UsedBytes:  4 * 1024 * 1024 * 1024, // 4GB (50% usage)
		},
		Disk: DiskMetrics{
			TotalBytes: 100 * 1024 * 1024 * 1024, // 100GB
			UsedBytes:  50 * 1024 * 1024 * 1024,  // 50GB (50% usage)
		},
	})

	alerts := monitor.GetAlerts()
	assert.NotNil(t, alerts)

	// Should have at least one CPU alert
	cpuAlerts := 0
	for _, alert := range alerts {
		if alert.Type == "cpu_usage" {
			cpuAlerts++
		}
	}
	assert.Greater(t, cpuAlerts, 0, "Should have CPU usage alerts")

	// Test memory alert threshold
	monitor.checkAlerts(SystemMetrics{
		CPU: CPUMetrics{
			UsagePercent: 50.0, // Below threshold
		},
		Memory: MemoryMetrics{
			TotalBytes: 8 * 1024 * 1024 * 1024,   // 8GB
			UsedBytes:  7.5 * 1024 * 1024 * 1024, // 7.5GB (93.75% usage)
		},
		Disk: DiskMetrics{
			TotalBytes: 100 * 1024 * 1024 * 1024, // 100GB
			UsedBytes:  50 * 1024 * 1024 * 1024,  // 50GB (50% usage)
		},
	})

	alerts = monitor.GetAlerts()
	memoryAlerts := 0
	for _, alert := range alerts {
		if alert.Type == "memory_usage" {
			memoryAlerts++
		}
	}
	assert.Greater(t, memoryAlerts, 0, "Should have memory usage alerts")
}

func TestMetrics_Persistence(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: 100 * time.Millisecond,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for several metrics collections
	time.Sleep(500 * time.Millisecond)

	// Get historical metrics
	history := monitor.GetMetricsHistory(time.Now().Add(-time.Minute), time.Now())
	assert.NotNil(t, history)

	// Should have collected multiple data points
	assert.Greater(t, len(history), 1, "Should have multiple historical metrics")

	// Verify timestamps are ordered
	for i := 1; i < len(history); i++ {
		assert.True(t, history[i].Timestamp.After(history[i-1].Timestamp) || 
			history[i].Timestamp.Equal(history[i-1].Timestamp),
			"Metrics should be ordered by timestamp")
	}
}

func TestMonitor_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for startup
	time.Sleep(500 * time.Millisecond)

	// Verify monitor is running
	health := monitor.GetHealth()
	assert.True(t, health.Healthy)

	// Shutdown
	err = monitor.Shutdown()
	assert.NoError(t, err)

	// Wait for shutdown
	time.Sleep(100 * time.Millisecond)

	// Verify shutdown
	health = monitor.GetHealth()
	assert.False(t, health.Healthy)
	assert.Equal(t, "shutdown", health.Status)
}

func TestCustomMetrics_Registration(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	// Register custom metrics
	err = monitor.RegisterCustomMetric("inference_requests_total", "counter", "Total number of inference requests")
	require.NoError(t, err)

	err = monitor.RegisterCustomMetric("model_load_duration", "histogram", "Time taken to load models")
	require.NoError(t, err)

	// Update custom metrics
	monitor.UpdateCustomMetric("inference_requests_total", 100)
	monitor.UpdateCustomMetric("model_load_duration", 2.5)

	// Get custom metrics
	customMetrics := monitor.GetCustomMetrics()
	assert.NotNil(t, customMetrics)
	assert.Contains(t, customMetrics, "inference_requests_total")
	assert.Contains(t, customMetrics, "model_load_duration")

	assert.Equal(t, float64(100), customMetrics["inference_requests_total"])
	assert.Equal(t, 2.5, customMetrics["model_load_duration"])
}

func TestPrometheus_Integration(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: time.Second,
			DataRetention:  time.Hour,
			PrometheusEnabled: true,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start monitoring
	go monitor.Start(ctx)

	// Wait for startup
	time.Sleep(500 * time.Millisecond)

	// Test Prometheus metrics endpoint
	promMetrics := monitor.GetPrometheusMetrics()
	assert.NotNil(t, promMetrics)
	assert.NotEmpty(t, promMetrics)

	// Should contain standard Prometheus metrics
	assert.Contains(t, promMetrics, "# HELP")
	assert.Contains(t, promMetrics, "# TYPE")
}

func TestMetrics_Aggregation(t *testing.T) {
	cfg := &config.Config{
		Monitoring: &config.MonitoringConfig{
			Enabled:        true,
			MetricsPort:    9090,
			HealthPort:     8080,
			UpdateInterval: 100 * time.Millisecond,
			DataRetention:  time.Hour,
		},
	}

	monitor, err := NewMonitor(cfg)
	require.NoError(t, err)

	// Add multiple data points
	now := time.Now()
	for i := 0; i < 10; i++ {
		monitor.recordMetrics(SystemMetrics{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			CPU: CPUMetrics{
				UsagePercent: float64(10 + i*5),
			},
			Memory: MemoryMetrics{
				TotalBytes: 8 * 1024 * 1024 * 1024,
				UsedBytes:  uint64((3 + i) * 1024 * 1024 * 1024),
			},
		})
	}

	// Test aggregation
	avg := monitor.GetAverageMetrics(now.Add(-10*time.Minute), now)
	assert.NotNil(t, avg)
	assert.Greater(t, avg.CPU.UsagePercent, 0.0)
	assert.Greater(t, avg.Memory.UsedBytes, uint64(0))

	min := monitor.GetMinMetrics(now.Add(-10*time.Minute), now)
	assert.NotNil(t, min)
	assert.Equal(t, 10.0, min.CPU.UsagePercent)

	max := monitor.GetMaxMetrics(now.Add(-10*time.Minute), now)
	assert.NotNil(t, max)
	assert.Equal(t, 55.0, max.CPU.UsagePercent) // 10 + 9*5
}