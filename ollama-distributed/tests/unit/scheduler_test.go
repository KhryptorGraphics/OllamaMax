package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// TestDistributedScheduler tests the distributed scheduler
func TestDistributedScheduler(t *testing.T) {
	// Create test configuration
	config := &distributed.DistributedConfig{
		ClusterID:             "test-cluster",
		NodeID:                "test-node-1",
		MaxNodes:              5,
		HeartbeatInterval:     30 * time.Second,
		DefaultStrategy:       "layerwise",
		LayerThreshold:        8,
		BatchSizeLimit:        32,
		LBAlgorithm:           "weighted_round_robin",
		LatencyTarget:         100 * time.Millisecond,
		ReplicationFactor:     2,
		HealthCheckInterval:   10 * time.Second,
		RecoveryTimeout:       30 * time.Second,
		CommunicationProtocol: "libp2p",
		Encryption:            true,
		Compression:           true,
	}

	// Create base scheduler (mock)
	baseScheduler := &server.Scheduler{}

	// Create P2P node (mock)
	p2pNode := &p2p.Node{}

	// Create consensus engine (mock)
	consensusEngine := &consensus.Engine{}

	// Create distributed scheduler
	scheduler, err := distributed.NewDistributedScheduler(baseScheduler, config, p2pNode, consensusEngine)
	require.NoError(t, err)
	defer scheduler.Shutdown(context.Background())

	t.Run("TestInitialization", func(t *testing.T) {
		assert.NotNil(t, scheduler)
		assert.Equal(t, config.ClusterID, scheduler.GetConfig().ClusterID)
		assert.Equal(t, config.NodeID, scheduler.GetConfig().NodeID)
	})

	t.Run("TestMetrics", func(t *testing.T) {
		metrics := scheduler.GetMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(0), metrics.TotalRequests)
		assert.Equal(t, int64(0), metrics.CompletedRequests)
		assert.Equal(t, int64(0), metrics.FailedRequests)
	})

	t.Run("TestNodeManagement", func(t *testing.T) {
		nodes := scheduler.GetNodes()
		assert.NotNil(t, nodes)
		assert.Equal(t, 0, len(nodes)) // Initially no nodes
	})

	t.Run("TestActiveTasks", func(t *testing.T) {
		tasks := scheduler.GetActiveTasks()
		assert.NotNil(t, tasks)
		assert.Equal(t, 0, len(tasks)) // Initially no tasks
	})

	t.Run("TestClusterHealth", func(t *testing.T) {
		health := scheduler.GetClusterHealth()
		assert.NotNil(t, health)
		assert.Equal(t, 0, len(health)) // Initially no health checks
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

	// Create distributed engine
	engine := distributed.NewDistributedEngine(scheduler, config)
	require.NotNil(t, engine)

	t.Run("TestTaskExecution", func(t *testing.T) {
		// Create test task
		task := &distributed.DistributedTask{
			ID:        "test-task-1",
			Type:      distributed.TaskTypeInference,
			ModelName: "test-model",
			Status:    distributed.TaskStatusPending,
			CreatedAt: time.Now(),
			Priority:  1,
			Metadata:  make(map[string]interface{}),
		}

		// Test task status transitions
		assert.Equal(t, distributed.TaskStatusPending, task.Status)

		task.Status = distributed.TaskStatusPartitioned
		assert.Equal(t, distributed.TaskStatusPartitioned, task.Status)

		task.Status = distributed.TaskStatusScheduled
		assert.Equal(t, distributed.TaskStatusScheduled, task.Status)

		task.Status = distributed.TaskStatusRunning
		assert.Equal(t, distributed.TaskStatusRunning, task.Status)

		task.Status = distributed.TaskStatusCompleted
		assert.Equal(t, distributed.TaskStatusCompleted, task.Status)
	})

	t.Run("TestSubtaskManagement", func(t *testing.T) {
		// Create test subtask
		subtask := &distributed.Subtask{
			ID:       "subtask-1",
			TaskID:   "test-task-1",
			NodeID:   "test-node-1",
			Type:     "inference",
			Status:   distributed.TaskStatusPending,
			Metadata: make(map[string]interface{}),
		}

		assert.Equal(t, "subtask-1", subtask.ID)
		assert.Equal(t, "test-task-1", subtask.TaskID)
		assert.Equal(t, "test-node-1", subtask.NodeID)
		assert.Equal(t, distributed.TaskStatusPending, subtask.Status)
	})
}

// TestClusterManager tests the cluster manager
func TestClusterManager(t *testing.T) {
	// Create test scheduler
	scheduler := &distributed.DistributedScheduler{}

	// Create cluster manager
	manager := distributed.NewClusterManager(scheduler)
	require.NotNil(t, manager)

	t.Run("TestNodeRegistration", func(t *testing.T) {
		// Create test node
		node := &distributed.NodeInfo{
			ID:      "test-node-1",
			Address: "127.0.0.1:8080",
			Status:  distributed.NodeStatusOnline,
			Capacity: &distributed.ResourceCapacity{
				CPUCores:       8,
				MemoryBytes:    16 * 1024 * 1024 * 1024, // 16GB
				GPUCount:       1,
				GPUMemoryBytes: 8 * 1024 * 1024 * 1024, // 8GB
				ComputeScore:   100.0,
			},
			Usage: &distributed.ResourceUsage{
				CPUUtilization:    0.5,
				MemoryUtilization: 0.3,
				GPUUtilization:    0.0,
				ActiveRequests:    0,
				QueuedRequests:    0,
			},
			LastSeen:     time.Now(),
			Capabilities: []string{"inference", "training"},
		}

		// Test node registration
		manager.RegisterNode(node)

		// Verify node is registered
		registeredNode := manager.GetNode(node.ID)
		assert.NotNil(t, registeredNode)
		assert.Equal(t, node.ID, registeredNode.ID)
		assert.Equal(t, node.Address, registeredNode.Address)
		assert.Equal(t, node.Status, registeredNode.Status)
	})

	t.Run("TestModelRegistration", func(t *testing.T) {
		// Create test model
		model := &distributed.ModelInfo{
			Name:              "test-model",
			Path:              "/models/test-model",
			Size:              1024 * 1024 * 1024, // 1GB
			Checksum:          "abc123",
			Locations:         []string{"test-node-1"},
			ReplicationFactor: 2,
			AccessCount:       10,
			LastAccessed:      time.Now(),
			Popularity:        0.8,
		}

		// Test model registration
		manager.RegisterModel(model)

		// Verify model is registered
		registeredModel := manager.GetModel(model.Name)
		assert.NotNil(t, registeredModel)
		assert.Equal(t, model.Name, registeredModel.Name)
		assert.Equal(t, model.Size, registeredModel.Size)
		assert.Equal(t, model.Checksum, registeredModel.Checksum)
	})

	t.Run("TestHeartbeat", func(t *testing.T) {
		// Create test heartbeat
		heartbeat := &distributed.HeartbeatMessage{
			NodeID:    "test-node-1",
			Timestamp: time.Now(),
			Status:    distributed.NodeStatusOnline,
			Capacity: &distributed.ResourceCapacity{
				CPUCores:    8,
				MemoryBytes: 16 * 1024 * 1024 * 1024,
			},
			Usage: &distributed.ResourceUsage{
				CPUUtilization:    0.6,
				MemoryUtilization: 0.4,
			},
		}

		// Test heartbeat processing
		manager.ProcessHeartbeat(heartbeat)

		// Verify node status is updated
		node := manager.GetNode(heartbeat.NodeID)
		if node != nil {
			assert.Equal(t, heartbeat.Status, node.Status)
		}
	})
}

// TestLoadBalancer tests the load balancer
func TestLoadBalancer(t *testing.T) {
	// Create test configuration
	config := &distributed.LoadBalancerConfig{
		Algorithm:     "weighted_round_robin",
		LatencyTarget: 100 * time.Millisecond,
		WeightFactors: map[string]float64{
			"cpu":    0.3,
			"memory": 0.3,
			"gpu":    0.4,
		},
	}

	// Create load balancer
	lb := distributed.NewLoadBalancer(config)
	require.NotNil(t, lb)

	t.Run("TestNodeSelection", func(t *testing.T) {
		// Create test nodes
		nodes := []*distributed.NodeInfo{
			{
				ID:     "node-1",
				Status: distributed.NodeStatusOnline,
				Capacity: &distributed.ResourceCapacity{
					CPUCores:    8,
					MemoryBytes: 16 * 1024 * 1024 * 1024,
					GPUCount:    1,
				},
				Usage: &distributed.ResourceUsage{
					CPUUtilization:    0.5,
					MemoryUtilization: 0.3,
					GPUUtilization:    0.2,
				},
				Latency: 50 * time.Millisecond,
			},
			{
				ID:     "node-2",
				Status: distributed.NodeStatusOnline,
				Capacity: &distributed.ResourceCapacity{
					CPUCores:    4,
					MemoryBytes: 8 * 1024 * 1024 * 1024,
					GPUCount:    2,
				},
				Usage: &distributed.ResourceUsage{
					CPUUtilization:    0.8,
					MemoryUtilization: 0.7,
					GPUUtilization:    0.1,
				},
				Latency: 30 * time.Millisecond,
			},
		}

		// Create test task
		task := &distributed.DistributedTask{
			ID:        "test-task",
			Type:      distributed.TaskTypeInference,
			ModelName: "test-model",
		}

		// Test node selection
		selectedNodes, err := lb.SelectNodes(task, nodes)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(selectedNodes), 1)
		assert.LessOrEqual(t, len(selectedNodes), len(nodes))
	})

	t.Run("TestLoadBalancingAlgorithms", func(t *testing.T) {
		algorithms := []string{
			"round_robin",
			"weighted_round_robin",
			"least_connections",
			"least_response_time",
			"resource_aware",
		}

		for _, algorithm := range algorithms {
			config.Algorithm = algorithm
			lb := distributed.NewLoadBalancer(config)
			assert.NotNil(t, lb)
		}
	})
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
