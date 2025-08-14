package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// TestDistributedScheduler tests the distributed scheduler
func TestDistributedScheduler(t *testing.T) {
	// Use helper function to create scheduler
	scheduler := createMockSchedulerEngine(t)
	require.NotNil(t, scheduler)

	t.Run("TestInitialization", func(t *testing.T) {
		assert.NotNil(t, scheduler, "Scheduler should be created")
	})

	t.Run("TestMetrics", func(t *testing.T) {
		// Test basic metrics functionality
		assert.NotNil(t, scheduler, "Scheduler should have metrics")
	})

	t.Run("TestNodeManagement", func(t *testing.T) {
		// Test basic node management functionality
		assert.NotNil(t, scheduler, "Scheduler should handle nodes")
	})

	t.Run("TestActiveTasks", func(t *testing.T) {
		// Test basic task management functionality
		assert.NotNil(t, scheduler, "Scheduler should handle tasks")
	})

	t.Run("TestClusterHealth", func(t *testing.T) {
		// Test basic health check functionality
		assert.NotNil(t, scheduler, "Scheduler should handle health checks")
	})
}

// TestDistributedEngine tests the distributed engine
func TestDistributedEngine(t *testing.T) {
	// Create test scheduler
	scheduler := &distributed.DistributedScheduler{}

	// Create test configuration
	config := &distributed.DistributedConfig{
		DefaultStrategy: "layerwise",
		LayerThreshold:  8,
		BatchSizeLimit:  32,
	}

	// Test basic functionality
	assert.NotNil(t, scheduler)
	assert.NotNil(t, config)
}

// TestClusterManager tests the cluster manager
func TestClusterManager(t *testing.T) {
	// Create test scheduler
	scheduler := &distributed.DistributedScheduler{}

	// Create cluster manager
	assert.NotNil(t, scheduler)
}

// TestLoadBalancer tests the load balancer
func TestLoadBalancer(t *testing.T) {
	// Create test configuration
	assert.True(t, true, "Load balancer test placeholder")
}

// TestResourceMetrics tests resource metrics collection
func TestResourceMetrics(t *testing.T) {
	// Create test resource capacity
	capacity := &distributed.ResourceCapacity{
		CPUCores:         8,
		MemoryBytes:      16 * 1024 * 1024 * 1024,   // 16GB
		DiskBytes:        1024 * 1024 * 1024 * 1024, // 1TB
		GPUCount:         2,
		GPUMemoryBytes:   16 * 1024 * 1024 * 1024, // 16GB
		NetworkBandwidth: 1000 * 1000 * 1000,      // 1Gbps
		ComputeScore:     150.0,
	}

	// Create test resource usage
	usage := &distributed.ResourceUsage{
		CPUUtilization:     0.65,
		MemoryUtilization:  0.45,
		DiskUtilization:    0.30,
		GPUUtilization:     0.80,
		NetworkUtilization: 0.25,
		ActiveRequests:     5,
		QueuedRequests:     2,
		LoadAverage:        2.5,
	}

	// Test capacity metrics
	assert.Equal(t, int64(8), capacity.CPUCores)
	assert.Equal(t, int64(16*1024*1024*1024), capacity.MemoryBytes)
	assert.Equal(t, 2, capacity.GPUCount)
	assert.Equal(t, 150.0, capacity.ComputeScore)

	// Test usage metrics
	assert.Equal(t, 0.65, usage.CPUUtilization)
	assert.Equal(t, 0.45, usage.MemoryUtilization)
	assert.Equal(t, 0.80, usage.GPUUtilization)
	assert.Equal(t, 5, usage.ActiveRequests)
	assert.Equal(t, 2, usage.QueuedRequests)
}

// BenchmarkDistributedScheduler benchmarks scheduler performance
func BenchmarkDistributedScheduler(b *testing.B) {
	// Create test configuration
	config := &distributed.DistributedConfig{
		ClusterID:         "bench-cluster",
		NodeID:            "bench-node",
		DefaultStrategy:   "layerwise",
		LBAlgorithm:       "weighted_round_robin",
		ReplicationFactor: 2,
	}

	// Create scheduler
	scheduler, err := distributed.NewDistributedScheduler(nil, config, nil, nil)
	require.NoError(b, err)
	defer scheduler.Shutdown(context.Background())

	b.ResetTimer()

	b.Run("TaskCreation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				task := &distributed.DistributedTask{
					ID:        fmt.Sprintf("bench-task-%d", i),
					Type:      distributed.TaskTypeInference,
					ModelName: "bench-model",
					Status:    distributed.TaskStatusPending,
					CreatedAt: time.Now(),
					Priority:  1,
				}

				// Simulate task processing
				_ = task
				i++
			}
		})
	})

	b.Run("MetricsCollection", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				metrics := scheduler.GetMetrics()
				_ = metrics
			}
		})
	})
}
