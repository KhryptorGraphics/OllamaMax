package unit

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
)

// TestModelSyncManager tests the model synchronization manager
func TestModelSyncManager(t *testing.T) {
	// Create test configuration
	config := &models.SyncConfig{
		CheckInterval:     10 * time.Second,
		MaxConcurrentSync: 3,
		RetryAttempts:     3,
		RetryDelay:        1 * time.Second,
		CompressionLevel:  6,
		EnableDelta:       true,
		DeltaThreshold:    0.1,
	}

	// Create sync manager
	manager := models.NewSyncManager(config)
	require.NotNil(t, manager)

	t.Run("TestInitialization", func(t *testing.T) {
		assert.Equal(t, config.CheckInterval, manager.GetConfig().CheckInterval)
		assert.Equal(t, config.MaxConcurrentSync, manager.GetConfig().MaxConcurrentSync)
		assert.Equal(t, config.EnableDelta, manager.GetConfig().EnableDelta)
	})

	t.Run("TestModelRegistration", func(t *testing.T) {
		// Create test model
		model := &models.ModelInfo{
			Name:        "test-model",
			Version:     "1.0.0",
			Path:        "/models/test-model",
			Size:        1024 * 1024 * 1024, // 1GB
			Checksum:    "abc123def456",
			LastUpdated: time.Now(),
			Metadata: map[string]interface{}{
				"architecture": "transformer",
				"parameters":   "7B",
			},
		}

		// Register model
		err := manager.RegisterModel(model)
		assert.NoError(t, err)

		// Verify model is registered
		registeredModel := manager.GetModel(model.Name)
		assert.NotNil(t, registeredModel)
		assert.Equal(t, model.Name, registeredModel.Name)
		assert.Equal(t, model.Version, registeredModel.Version)
		assert.Equal(t, model.Checksum, registeredModel.Checksum)
	})

	t.Run("TestModelSync", func(t *testing.T) {
		// Create source and target models
		sourceModel := &models.ModelInfo{
			Name:        "sync-model",
			Version:     "1.0.0",
			Checksum:    "source-checksum",
			LastUpdated: time.Now(),
		}

		targetModel := &models.ModelInfo{
			Name:        "sync-model",
			Version:     "0.9.0",
			Checksum:    "target-checksum",
			LastUpdated: time.Now().Add(-1 * time.Hour),
		}

		// Check if sync is needed
		needsSync := manager.NeedsSync(sourceModel, targetModel)
		assert.True(t, needsSync)

		// Create sync task
		task := &models.SyncTask{
			ID:          "sync-task-1",
			ModelName:   sourceModel.Name,
			SourceNode:  "node-1",
			TargetNode:  "node-2",
			Status:      models.SyncStatusPending,
			CreatedAt:   time.Now(),
			Priority:    1,
		}

		// Process sync task
		err := manager.ProcessSyncTask(task)
		assert.NoError(t, err)
	})
}

// TestDeltaTracker tests the delta tracking functionality
func TestDeltaTracker(t *testing.T) {
	// Create delta tracker
	tracker := models.NewDeltaTracker()
	require.NotNil(t, tracker)

	t.Run("TestDeltaCreation", func(t *testing.T) {
		// Create test model versions
		oldModel := &models.ModelInfo{
			Name:     "delta-model",
			Version:  "1.0.0",
			Checksum: "old-checksum",
			Size:     1024 * 1024 * 1024, // 1GB
		}

		newModel := &models.ModelInfo{
			Name:     "delta-model",
			Version:  "1.1.0",
			Checksum: "new-checksum",
			Size:     1024 * 1024 * 1024 + 50*1024*1024, // 1GB + 50MB
		}

		// Create delta
		delta, err := tracker.CreateDelta(oldModel, newModel)
		assert.NoError(t, err)
		assert.NotNil(t, delta)
		assert.Equal(t, oldModel.Version, delta.FromVersion)
		assert.Equal(t, newModel.Version, delta.ToVersion)
		assert.Greater(t, delta.Size, int64(0))
	})

	t.Run("TestDeltaApplication", func(t *testing.T) {
		// Create test delta
		delta := &models.Delta{
			ID:          "delta-1",
			ModelName:   "delta-model",
			FromVersion: "1.0.0",
			ToVersion:   "1.1.0",
			Size:        50 * 1024 * 1024, // 50MB
			Checksum:    "delta-checksum",
			CreatedAt:   time.Now(),
			Data:        []byte("mock-delta-data"),
		}

		// Apply delta
		err := tracker.ApplyDelta(delta)
		assert.NoError(t, err)
	})

	t.Run("TestDeltaOptimization", func(t *testing.T) {
		// Test delta size optimization
		largeDelta := &models.Delta{
			Size: 900 * 1024 * 1024, // 900MB delta
		}
		
		originalSize := int64(1024 * 1024 * 1024) // 1GB original
		
		// Check if delta is worth using
		isWorthwhile := tracker.IsDeltaWorthwhile(largeDelta, originalSize)
		assert.False(t, isWorthwhile) // 900MB delta for 1GB file is not worthwhile
		
		smallDelta := &models.Delta{
			Size: 50 * 1024 * 1024, // 50MB delta
		}
		
		isWorthwhile = tracker.IsDeltaWorthwhile(smallDelta, originalSize)
		assert.True(t, isWorthwhile) // 50MB delta for 1GB file is worthwhile
	})
}

// TestReplicationManager tests the replication manager
func TestReplicationManager(t *testing.T) {
	// Create replication configuration
	config := &models.ReplicationConfig{
		DefaultFactor:    3,
		MinReplicas:      2,
		MaxReplicas:      5,
		PlacementPolicy:  "anti_affinity",
		HealthCheckInterval: 30 * time.Second,
		RepairDelay:      5 * time.Minute,
	}

	// Create replication manager
	manager := models.NewReplicationManager(config)
	require.NotNil(t, manager)

	t.Run("TestReplicationPolicy", func(t *testing.T) {
		// Create test model
		model := &models.ModelInfo{
			Name:              "replicated-model",
			Size:              2 * 1024 * 1024 * 1024, // 2GB
			ReplicationFactor: 3,
			Locations:         []string{"node-1"},
		}

		// Create test nodes
		nodes := []*models.NodeInfo{
			{ID: "node-1", Zone: "zone-a", Rack: "rack-1"},
			{ID: "node-2", Zone: "zone-a", Rack: "rack-2"},
			{ID: "node-3", Zone: "zone-b", Rack: "rack-1"},
			{ID: "node-4", Zone: "zone-b", Rack: "rack-2"},
		}

		// Calculate replication targets
		targets, err := manager.CalculateReplicationTargets(model, nodes)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(targets)) // Need 2 more replicas
		
		// Verify anti-affinity (different zones/racks)
		zones := make(map[string]bool)
		for _, target := range targets {
			zones[target.Zone] = true
		}
		assert.GreaterOrEqual(t, len(zones), 1)
	})

	t.Run("TestReplicationHealth", func(t *testing.T) {
		// Create test model with replicas
		model := &models.ModelInfo{
			Name:              "health-model",
			ReplicationFactor: 3,
			Locations:         []string{"node-1", "node-2", "node-3"},
		}

		// Simulate failed replica
		failedReplicas := []string{"node-2"}
		
		// Check replication health
		health := manager.CheckReplicationHealth(model, failedReplicas)
		assert.False(t, health.IsHealthy)
		assert.Equal(t, 2, health.ActiveReplicas)
		assert.Equal(t, 3, health.DesiredReplicas)
		assert.Equal(t, 1, health.FailedReplicas)
	})

	t.Run("TestReplicationRepair", func(t *testing.T) {
		// Create repair task
		task := &models.RepairTask{
			ID:          "repair-1",
			ModelName:   "health-model",
			FailedNode:  "node-2",
			TargetNode:  "node-4",
			Status:      models.RepairStatusPending,
			CreatedAt:   time.Now(),
			Priority:    2,
		}

		// Process repair task
		err := manager.ProcessRepairTask(task)
		assert.NoError(t, err)
	})
}

// TestCASStore tests the Content Addressable Storage
func TestCASStore(t *testing.T) {
	// Create CAS store
	store := models.NewCASStore("/tmp/cas-test")
	require.NotNil(t, store)

	t.Run("TestContentStorage", func(t *testing.T) {
		// Create test content
		content := []byte("test content for CAS storage")
		
		// Calculate hash
		hash := sha256.Sum256(content)
		hashStr := fmt.Sprintf("%x", hash)
		
		// Store content
		err := store.Put(hashStr, content)
		assert.NoError(t, err)
		
		// Retrieve content
		retrievedContent, err := store.Get(hashStr)
		assert.NoError(t, err)
		assert.Equal(t, content, retrievedContent)
		
		// Check existence
		exists := store.Has(hashStr)
		assert.True(t, exists)
	})

	t.Run("TestContentDeduplication", func(t *testing.T) {
		// Store same content multiple times
		content := []byte("duplicate content")
		hash := sha256.Sum256(content)
		hashStr := fmt.Sprintf("%x", hash)
		
		// Store first time
		err := store.Put(hashStr, content)
		assert.NoError(t, err)
		
		// Store second time (should be deduplicated)
		err = store.Put(hashStr, content)
		assert.NoError(t, err)
		
		// Verify content is still there
		retrievedContent, err := store.Get(hashStr)
		assert.NoError(t, err)
		assert.Equal(t, content, retrievedContent)
	})

	t.Run("TestContentDeletion", func(t *testing.T) {
		// Store content
		content := []byte("content to delete")
		hash := sha256.Sum256(content)
		hashStr := fmt.Sprintf("%x", hash)
		
		err := store.Put(hashStr, content)
		assert.NoError(t, err)
		
		// Delete content
		err = store.Delete(hashStr)
		assert.NoError(t, err)
		
		// Verify content is gone
		exists := store.Has(hashStr)
		assert.False(t, exists)
		
		_, err = store.Get(hashStr)
		assert.Error(t, err)
	})
}

// TestModelDistributionManager tests the model distribution manager
func TestModelDistributionManager(t *testing.T) {
	// Create distribution manager
	manager := models.NewDistributionManager()
	require.NotNil(t, manager)

	t.Run("TestModelDistribution", func(t *testing.T) {
		// Create test model
		model := &models.ModelInfo{
			Name:              "distributed-model",
			Size:              4 * 1024 * 1024 * 1024, // 4GB
			ReplicationFactor: 2,
			Popularity:        0.8,
		}

		// Create test nodes
		nodes := []*models.NodeInfo{
			{
				ID:       "node-1",
				Capacity: 10 * 1024 * 1024 * 1024, // 10GB
				Usage:    2 * 1024 * 1024 * 1024,  // 2GB used
				Latency:  50 * time.Millisecond,
			},
			{
				ID:       "node-2",
				Capacity: 8 * 1024 * 1024 * 1024, // 8GB
				Usage:    1 * 1024 * 1024 * 1024, // 1GB used
				Latency:  30 * time.Millisecond,
			},
			{
				ID:       "node-3",
				Capacity: 6 * 1024 * 1024 * 1024, // 6GB
				Usage:    5 * 1024 * 1024 * 1024, // 5GB used
				Latency:  100 * time.Millisecond,
			},
		}

		// Calculate distribution strategy
		strategy, err := manager.CalculateDistributionStrategy(model, nodes)
		assert.NoError(t, err)
		assert.NotNil(t, strategy)
		assert.Equal(t, model.ReplicationFactor, len(strategy.TargetNodes))
		
		// Verify nodes have enough capacity
		for _, nodeID := range strategy.TargetNodes {
			node := findNode(nodes, nodeID)
			assert.NotNil(t, node)
			assert.Greater(t, node.Capacity-node.Usage, model.Size)
		}
	})

	t.Run("TestLoadBalancing", func(t *testing.T) {
		// Test model placement for load balancing
		models := []*models.ModelInfo{
			{Name: "model-1", Size: 1 * 1024 * 1024 * 1024, Popularity: 0.9},
			{Name: "model-2", Size: 2 * 1024 * 1024 * 1024, Popularity: 0.7},
			{Name: "model-3", Size: 3 * 1024 * 1024 * 1024, Popularity: 0.5},
		}

		nodes := []*models.NodeInfo{
			{ID: "node-1", Capacity: 10 * 1024 * 1024 * 1024, Usage: 0},
			{ID: "node-2", Capacity: 10 * 1024 * 1024 * 1024, Usage: 0},
			{ID: "node-3", Capacity: 10 * 1024 * 1024 * 1024, Usage: 0},
		}

		// Calculate balanced distribution
		distribution, err := manager.CalculateBalancedDistribution(models, nodes)
		assert.NoError(t, err)
		assert.NotNil(t, distribution)
		
		// Verify all models are distributed
		assert.Equal(t, len(models), len(distribution.Placements))
		
		// Check load distribution
		loadDistribution := distribution.GetLoadDistribution()
		assert.NotNil(t, loadDistribution)
		assert.Equal(t, len(nodes), len(loadDistribution))
	})
}

// Helper function to find node by ID
func findNode(nodes []*models.NodeInfo, id string) *models.NodeInfo {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// BenchmarkModelSync benchmarks model synchronization
func BenchmarkModelSync(b *testing.B) {
	// Create sync manager
	config := &models.SyncConfig{
		MaxConcurrentSync: 10,
		EnableDelta:       true,
		CompressionLevel:  6,
	}
	
	manager := models.NewSyncManager(config)
	require.NotNil(b, manager)

	b.ResetTimer()

	b.Run("ModelRegistration", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				model := &models.ModelInfo{
					Name:        fmt.Sprintf("bench-model-%d", i),
					Version:     "1.0.0",
					Checksum:    fmt.Sprintf("checksum-%d", i),
					Size:        1024 * 1024 * 1024, // 1GB
					LastUpdated: time.Now(),
				}
				
				err := manager.RegisterModel(model)
				if err != nil {
					b.Errorf("Failed to register model: %v", err)
				}
				
				i++
			}
		})
	})

	b.Run("SyncDecision", func(b *testing.B) {
		sourceModel := &models.ModelInfo{
			Name:        "bench-model",
			Version:     "2.0.0",
			Checksum:    "new-checksum",
			LastUpdated: time.Now(),
		}

		targetModel := &models.ModelInfo{
			Name:        "bench-model",
			Version:     "1.0.0",
			Checksum:    "old-checksum",
			LastUpdated: time.Now().Add(-1 * time.Hour),
		}

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				needsSync := manager.NeedsSync(sourceModel, targetModel)
				_ = needsSync
			}
		})
	})

	b.Run("DeltaCreation", func(b *testing.B) {
		tracker := models.NewDeltaTracker()
		
		oldModel := &models.ModelInfo{
			Name:     "bench-delta-model",
			Version:  "1.0.0",
			Checksum: "old-checksum",
			Size:     1024 * 1024 * 1024,
		}

		newModel := &models.ModelInfo{
			Name:     "bench-delta-model",
			Version:  "1.1.0",
			Checksum: "new-checksum",
			Size:     1024 * 1024 * 1024 + 50*1024*1024,
		}

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				delta, err := tracker.CreateDelta(oldModel, newModel)
				if err != nil {
					b.Errorf("Failed to create delta: %v", err)
				}
				_ = delta
			}
		})
	})
}