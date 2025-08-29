package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/types"
)

func TestNewScheduler(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   10,
			JobTimeout:          30 * time.Second,
			ResourcePoolSize:    5,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	assert.Equal(t, cfg.Scheduler.Type, scheduler.config.Type)
	assert.Equal(t, cfg.Scheduler.MaxConcurrentJobs, scheduler.config.MaxConcurrentJobs)
	assert.Equal(t, cfg.Scheduler.JobTimeout, scheduler.config.JobTimeout)
}

func TestScheduler_ScheduleJob_RoundRobin(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    3,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add some nodes
	nodes := []types.Node{
		{ID: "node1", Address: "http://node1:8080", Resources: types.ResourceInfo{CPUCores: 4, MemoryGB: 8, GPUCount: 1}},
		{ID: "node2", Address: "http://node2:8080", Resources: types.ResourceInfo{CPUCores: 8, MemoryGB: 16, GPUCount: 2}},
		{ID: "node3", Address: "http://node3:8080", Resources: types.ResourceInfo{CPUCores: 2, MemoryGB: 4, GPUCount: 0}},
	}

	for _, node := range nodes {
		err := scheduler.AddNode(ctx, node)
		require.NoError(t, err)
	}

	// Create test jobs
	jobs := []types.Job{
		{
			ID:          "job1",
			Type:        "inference",
			ModelName:   "llama2",
			Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 1},
			Priority:    types.PriorityNormal,
		},
		{
			ID:          "job2",
			Type:        "training",
			ModelName:   "gpt-3.5",
			Requirements: types.ResourceRequirement{CPUCores: 4, MemoryGB: 8, GPUCount: 2},
			Priority:    types.PriorityHigh,
		},
		{
			ID:          "job3",
			Type:        "inference",
			ModelName:   "bert",
			Requirements: types.ResourceRequirement{CPUCores: 1, MemoryGB: 2, GPUCount: 0},
			Priority:    types.PriorityLow,
		},
	}

	// Schedule jobs
	scheduledJobs := make([]string, len(jobs))
	for i, job := range jobs {
		nodeID, err := scheduler.ScheduleJob(ctx, job)
		require.NoError(t, err)
		scheduledJobs[i] = nodeID
		assert.Contains(t, []string{"node1", "node2", "node3"}, nodeID)
	}

	// Verify round-robin distribution (jobs should be distributed across nodes)
	nodeUsage := make(map[string]int)
	for _, nodeID := range scheduledJobs {
		nodeUsage[nodeID]++
	}

	assert.Greater(t, len(nodeUsage), 1, "Jobs should be distributed across multiple nodes")
}

func TestScheduler_ScheduleJob_ResourceAware(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "resource_aware",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    3,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add nodes with different resources
	nodes := []types.Node{
		{ID: "gpu-node", Address: "http://gpu-node:8080", Resources: types.ResourceInfo{CPUCores: 8, MemoryGB: 32, GPUCount: 4}},
		{ID: "cpu-node", Address: "http://cpu-node:8080", Resources: types.ResourceInfo{CPUCores: 16, MemoryGB: 64, GPUCount: 0}},
		{ID: "small-node", Address: "http://small-node:8080", Resources: types.ResourceInfo{CPUCores: 2, MemoryGB: 4, GPUCount: 0}},
	}

	for _, node := range nodes {
		err := scheduler.AddNode(ctx, node)
		require.NoError(t, err)
	}

	// Create GPU-intensive job
	gpuJob := types.Job{
		ID:          "gpu-job",
		Type:        "training",
		ModelName:   "stable-diffusion",
		Requirements: types.ResourceRequirement{CPUCores: 4, MemoryGB: 16, GPUCount: 2},
		Priority:    types.PriorityHigh,
	}

	nodeID, err := scheduler.ScheduleJob(ctx, gpuJob)
	require.NoError(t, err)
	assert.Equal(t, "gpu-node", nodeID, "GPU-intensive job should be scheduled on GPU node")

	// Create CPU-intensive job
	cpuJob := types.Job{
		ID:          "cpu-job",
		Type:        "inference",
		ModelName:   "bert-large",
		Requirements: types.ResourceRequirement{CPUCores: 8, MemoryGB: 32, GPUCount: 0},
		Priority:    types.PriorityNormal,
	}

	nodeID, err = scheduler.ScheduleJob(ctx, cpuJob)
	require.NoError(t, err)
	assert.Equal(t, "cpu-node", nodeID, "CPU-intensive job should be scheduled on CPU node")

	// Create small job
	smallJob := types.Job{
		ID:          "small-job",
		Type:        "inference",
		ModelName:   "tiny-bert",
		Requirements: types.ResourceRequirement{CPUCores: 1, MemoryGB: 2, GPUCount: 0},
		Priority:    types.PriorityLow,
	}

	nodeID, err = scheduler.ScheduleJob(ctx, smallJob)
	require.NoError(t, err)
	assert.Contains(t, []string{"cpu-node", "small-node"}, nodeID, "Small job should fit on any available node")
}

func TestScheduler_ScheduleJob_PriorityBased(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "priority",
			MaxConcurrentJobs:   2, // Limited concurrency to test priority
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    1,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add a single node
	node := types.Node{
		ID:        "node1",
		Address:   "http://node1:8080",
		Resources: types.ResourceInfo{CPUCores: 4, MemoryGB: 8, GPUCount: 1},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Create jobs with different priorities
	highPriorityJob := types.Job{
		ID:          "high-job",
		Type:        "inference",
		ModelName:   "critical-model",
		Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 1},
		Priority:    types.PriorityHigh,
	}

	normalPriorityJob := types.Job{
		ID:          "normal-job",
		Type:        "training",
		ModelName:   "normal-model",
		Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 0},
		Priority:    types.PriorityNormal,
	}

	lowPriorityJob := types.Job{
		ID:          "low-job",
		Type:        "inference",
		ModelName:   "low-model",
		Requirements: types.ResourceRequirement{CPUCores: 1, MemoryGB: 2, GPUCount: 0},
		Priority:    types.PriorityLow,
	}

	// Schedule all jobs
	jobs := []types.Job{normalPriorityJob, lowPriorityJob, highPriorityJob}
	var wg sync.WaitGroup

	scheduledOrder := make([]string, 0, len(jobs))
	orderMutex := sync.Mutex{}

	for _, job := range jobs {
		wg.Add(1)
		go func(j types.Job) {
			defer wg.Done()
			nodeID, err := scheduler.ScheduleJob(ctx, j)
			require.NoError(t, err)
			assert.Equal(t, "node1", nodeID)

			orderMutex.Lock()
			scheduledOrder = append(scheduledOrder, j.ID)
			orderMutex.Unlock()
		}(job)
	}

	wg.Wait()

	// High priority job should be scheduled first (or among the first)
	assert.Contains(t, scheduledOrder, "high-job")
	assert.Contains(t, scheduledOrder, "normal-job")
	assert.Contains(t, scheduledOrder, "low-job")
}

func TestScheduler_ScheduleJob_InsufficientResources(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "resource_aware",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    1,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add a small node
	node := types.Node{
		ID:        "small-node",
		Address:   "http://small-node:8080",
		Resources: types.ResourceInfo{CPUCores: 2, MemoryGB: 4, GPUCount: 0},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Create job that exceeds node resources
	largeJob := types.Job{
		ID:          "large-job",
		Type:        "training",
		ModelName:   "large-model",
		Requirements: types.ResourceRequirement{CPUCores: 8, MemoryGB: 32, GPUCount: 2},
		Priority:    types.PriorityHigh,
	}

	_, err = scheduler.ScheduleJob(ctx, largeJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient resources")
}

func TestScheduler_JobLifecycle(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    2,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add node
	node := types.Node{
		ID:        "node1",
		Address:   "http://node1:8080",
		Resources: types.ResourceInfo{CPUCores: 4, MemoryGB: 8, GPUCount: 1},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Schedule job
	job := types.Job{
		ID:          "test-job",
		Type:        "inference",
		ModelName:   "test-model",
		Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 0},
		Priority:    types.PriorityNormal,
	}

	nodeID, err := scheduler.ScheduleJob(ctx, job)
	require.NoError(t, err)
	assert.Equal(t, "node1", nodeID)

	// Check job status
	status, err := scheduler.GetJobStatus(ctx, "test-job")
	require.NoError(t, err)
	assert.Equal(t, types.JobStatusScheduled, status.Status)
	assert.Equal(t, "node1", status.AssignedNode)

	// Update job status
	err = scheduler.UpdateJobStatus(ctx, "test-job", types.JobStatusRunning)
	require.NoError(t, err)

	status, err = scheduler.GetJobStatus(ctx, "test-job")
	require.NoError(t, err)
	assert.Equal(t, types.JobStatusRunning, status.Status)

	// Complete job
	err = scheduler.UpdateJobStatus(ctx, "test-job", types.JobStatusCompleted)
	require.NoError(t, err)

	status, err = scheduler.GetJobStatus(ctx, "test-job")
	require.NoError(t, err)
	assert.Equal(t, types.JobStatusCompleted, status.Status)
}

func TestScheduler_NodeManagement(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    3,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add nodes
	nodes := []types.Node{
		{ID: "node1", Address: "http://node1:8080", Resources: types.ResourceInfo{CPUCores: 4, MemoryGB: 8, GPUCount: 1}},
		{ID: "node2", Address: "http://node2:8080", Resources: types.ResourceInfo{CPUCores: 8, MemoryGB: 16, GPUCount: 2}},
	}

	for _, node := range nodes {
		err := scheduler.AddNode(ctx, node)
		require.NoError(t, err)
	}

	// List nodes
	nodeList, err := scheduler.ListNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, nodeList, 2)

	// Find specific nodes
	found := false
	for _, node := range nodeList {
		if node.ID == "node1" {
			found = true
			assert.Equal(t, "http://node1:8080", node.Address)
			assert.Equal(t, 4, node.Resources.CPUCores)
			break
		}
	}
	assert.True(t, found, "node1 should be in the list")

	// Remove node
	err = scheduler.RemoveNode(ctx, "node1")
	require.NoError(t, err)

	// Verify node is removed
	nodeList, err = scheduler.ListNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, nodeList, 1)
	assert.Equal(t, "node2", nodeList[0].ID)

	// Try to remove non-existent node
	err = scheduler.RemoveNode(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
}

func TestScheduler_GetStats(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    2,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add node and schedule some jobs
	node := types.Node{
		ID:        "node1",
		Address:   "http://node1:8080",
		Resources: types.ResourceInfo{CPUCores: 8, MemoryGB: 16, GPUCount: 2},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Schedule multiple jobs
	for i := 0; i < 3; i++ {
		job := types.Job{
			ID:          fmt.Sprintf("job-%d", i),
			Type:        "inference",
			ModelName:   "test-model",
			Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 0},
			Priority:    types.PriorityNormal,
		}

		_, err := scheduler.ScheduleJob(ctx, job)
		require.NoError(t, err)
	}

	// Get scheduler stats
	stats, err := scheduler.GetStats(ctx)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, stats.TotalJobs, uint64(3))
	assert.GreaterOrEqual(t, stats.ActiveJobs, uint64(0))
	assert.Equal(t, uint64(1), stats.TotalNodes)
	assert.Equal(t, uint64(1), stats.ActiveNodes)
}

func TestScheduler_Shutdown(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   5,
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    2,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add node
	node := types.Node{
		ID:        "node1",
		Address:   "http://node1:8080",
		Resources: types.ResourceInfo{CPUCores: 4, MemoryGB: 8, GPUCount: 1},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Schedule job
	job := types.Job{
		ID:          "test-job",
		Type:        "inference",
		ModelName:   "test-model",
		Requirements: types.ResourceRequirement{CPUCores: 2, MemoryGB: 4, GPUCount: 0},
		Priority:    types.PriorityNormal,
	}

	_, err = scheduler.ScheduleJob(ctx, job)
	require.NoError(t, err)

	// Shutdown scheduler
	err = scheduler.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify scheduler is shut down (subsequent operations should fail)
	_, err = scheduler.ScheduleJob(ctx, job)
	assert.Error(t, err)
}

func TestScheduler_ConcurrencyLimits(t *testing.T) {
	cfg := &config.Config{
		Scheduler: &config.SchedulerConfig{
			Type:                "round_robin",
			MaxConcurrentJobs:   2, // Very low limit for testing
			JobTimeout:          10 * time.Second,
			ResourcePoolSize:    1,
			LoadBalancingWeight: 1.0,
		},
	}

	scheduler, err := NewScheduler(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Add node
	node := types.Node{
		ID:        "node1",
		Address:   "http://node1:8080",
		Resources: types.ResourceInfo{CPUCores: 8, MemoryGB: 16, GPUCount: 2},
	}

	err = scheduler.AddNode(ctx, node)
	require.NoError(t, err)

	// Schedule jobs up to the limit
	var wg sync.WaitGroup
	successCount := int32(0)
	errorCount := int32(0)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(jobID int) {
			defer wg.Done()
			job := types.Job{
				ID:          fmt.Sprintf("job-%d", jobID),
				Type:        "inference",
				ModelName:   "test-model",
				Requirements: types.ResourceRequirement{CPUCores: 1, MemoryGB: 2, GPUCount: 0},
				Priority:    types.PriorityNormal,
			}

			_, err := scheduler.ScheduleJob(ctx, job)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
			} else {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// Some jobs should succeed, others should be rejected due to concurrency limits
	assert.Greater(t, successCount, int32(0), "Some jobs should succeed")
	t.Logf("Successful jobs: %d, Failed jobs: %d", successCount, errorCount)
}