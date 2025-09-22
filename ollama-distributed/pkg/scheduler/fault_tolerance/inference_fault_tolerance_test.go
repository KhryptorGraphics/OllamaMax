package fault_tolerance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInferenceCheckpointManager tests the checkpoint manager functionality
func TestInferenceCheckpointManager(t *testing.T) {
	t.Run("CreateAndRestoreCheckpoint", func(t *testing.T) {
		config := &CheckpointConfig{
			CheckpointInterval:   30 * time.Second,
			StorageBackend:       "memory",
			CompressionEnabled:   true,
			EncryptionEnabled:    false,
			RetentionPeriod:      24 * time.Hour,
			IncrementalSnapshots: true,
			MaxCheckpointSize:    1024 * 1024, // 1MB
		}

		manager := NewInferenceCheckpointManager(config)
		ctx := context.Background()

		// Create a checkpoint
		checkpoint := &InferenceCheckpoint{
			ID:        "test-checkpoint-1",
			SessionID: "session-1",
			Timestamp: time.Now(),
			ModelState: map[string]interface{}{
				"layer1": []float32{1.0, 2.0, 3.0},
				"layer2": []float32{4.0, 5.0, 6.0},
			},
			InferenceState: map[string]interface{}{
				"progress": 0.5,
				"tokens":   100,
			},
		}

		// Store checkpoint
		err := manager.storeCheckpoint(ctx, checkpoint)
		require.NoError(t, err)

		// Retrieve checkpoint
		retrieved, err := manager.getCheckpoint("test-checkpoint-1")
		require.NoError(t, err)
		assert.Equal(t, checkpoint.ID, retrieved.ID)
		assert.Equal(t, checkpoint.SessionID, retrieved.SessionID)
	})

	t.Run("IncrementalCheckpoints", func(t *testing.T) {
		config := &CheckpointConfig{
			IncrementalSnapshots: true,
			StorageBackend:       "memory",
		}

		manager := NewInferenceCheckpointManager(config)
		ctx := context.Background()

		// Create base checkpoint
		base := &InferenceCheckpoint{
			ID:        "base-1",
			SessionID: "session-2",
			ModelState: map[string]interface{}{
				"weights": []float32{1.0, 2.0, 3.0},
			},
		}
		err := manager.storeCheckpoint(ctx, base)
		require.NoError(t, err)

		// Create incremental checkpoint
		incremental := &InferenceCheckpoint{
			ID:               "incremental-1",
			SessionID:        "session-2",
			BaseCheckpointID: "base-1",
			ModelState: map[string]interface{}{
				"weights": []float32{1.1, 2.1, 3.1}, // Only changed values
			},
		}
		err = manager.storeCheckpoint(ctx, incremental)
		require.NoError(t, err)

		// Verify incremental checkpoint
		retrieved, err := manager.getCheckpoint("incremental-1")
		require.NoError(t, err)
		assert.Equal(t, "base-1", retrieved.BaseCheckpointID)
	})

	t.Run("CheckpointScheduling", func(t *testing.T) {
		config := &CheckpointConfig{
			CheckpointInterval: 100 * time.Millisecond,
			StorageBackend:     "memory",
		}

		manager := NewInferenceCheckpointManager(config)
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		scheduler := manager.scheduler
		scheduler.Start(ctx)

		// Wait for at least 2 checkpoint intervals
		time.Sleep(250 * time.Millisecond)

		// Verify checkpoints were created
		checkpoints := manager.ListCheckpoints("", 10)
		assert.GreaterOrEqual(t, len(checkpoints), 2)
	})
}

// TestDynamicRepartitioningManager tests dynamic repartitioning functionality
func TestDynamicRepartitioningManager(t *testing.T) {
	t.Run("HandleNodeFailure", func(t *testing.T) {
		config := &RepartitioningConfig{
			Strategy:             "conservative",
			MaxPartitionSize:     1024 * 1024,
			MinPartitionSize:     1024,
			RebalanceThreshold:   0.2,
			MigrationBandwidth:   1024 * 1024,
			EnableCompression:    true,
			EnableParallelTransfer: true,
		}

		manager := NewDynamicRepartitioningManager(config, nil)
		ctx := context.Background()

		// Simulate node failure
		operation, err := manager.HandleNodeFailure(ctx, "node-1", []string{"node-2", "node-3"}, []string{"shard-1", "shard-2"})
		require.NoError(t, err)
		assert.NotNil(t, operation)
		assert.Equal(t, "node-1", operation.FailedNodeID)
		assert.Len(t, operation.TargetNodes, 2)
	})

	t.Run("RepartitioningStrategies", func(t *testing.T) {
		strategies := []string{"conservative", "aggressive", "adaptive"}

		for _, strategy := range strategies {
			t.Run(strategy, func(t *testing.T) {
				config := &RepartitioningConfig{
					Strategy:         strategy,
					MaxPartitionSize: 1024 * 1024,
					MinPartitionSize: 1024,
				}

				manager := NewDynamicRepartitioningManager(config, nil)

				// Test strategy creation
				plan := manager.createPartitionPlan(context.Background(), "node-1", []string{"node-2", "node-3"})
				assert.NotNil(t, plan)
				assert.Equal(t, strategy, plan.Strategy)
			})
		}
	})

	t.Run("MigrationProcess", func(t *testing.T) {
		config := &RepartitioningConfig{
			Strategy:               "adaptive",
			MigrationBandwidth:     1024 * 1024,
			EnableParallelTransfer: true,
		}

		manager := NewDynamicRepartitioningManager(config, nil)
		ctx := context.Background()

		migration := &ShardMigration{
			ShardID:    "shard-1",
			SourceNode: "node-1",
			TargetNode: "node-2",
			Size:       1024 * 100, // 100KB
			Priority:   1,
		}

		// Execute migration (mock)
		err := manager.executeMigration(ctx, migration)
		assert.NoError(t, err)
	})
}

// TestGracefulDegradationManager tests graceful degradation functionality
func TestGracefulDegradationManager(t *testing.T) {
	t.Run("ApplyDegradationStrategies", func(t *testing.T) {
		config := &DegradationConfig{
			QualityThreshold:     0.8,
			MinAcceptableQuality: 0.6,
			MaxDegradationLevel:  3,
			RecoveryThreshold:    0.9,
			MonitoringInterval:   100 * time.Millisecond,
			EnableAutoRecovery:   true,
		}

		manager := NewInferenceGracefulDegradationManager(config)
		ctx := context.Background()

		inferenceState := &InferenceState{
			SessionID:      "session-1",
			ModelID:        "model-1",
			CurrentQuality: 0.7,
			Precision:      "fp32",
			BatchSize:      32,
		}

		resourceState := &ResourceState{
			AvailableMemory: 1024 * 1024 * 512, // 512MB
			AvailableCPU:    0.5,
			AvailableGPU:    0.3,
			NetworkLatency:  50 * time.Millisecond,
		}

		// Apply degradation
		result, err := manager.ApplyDegradation(ctx, inferenceState, resourceState)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Applied)
	})

	t.Run("QualityMonitoring", func(t *testing.T) {
		config := &DegradationConfig{
			QualityThreshold:   0.8,
			MonitoringInterval: 50 * time.Millisecond,
		}

		manager := NewInferenceGracefulDegradationManager(config)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// Start monitoring
		go manager.MonitorQuality(ctx)

		// Simulate quality changes
		manager.UpdateQualityMetrics("session-1", 0.75)
		time.Sleep(60 * time.Millisecond)
		manager.UpdateQualityMetrics("session-1", 0.85)
		time.Sleep(60 * time.Millisecond)

		// Check if degradation was triggered
		metrics := manager.GetMetrics()
		assert.NotNil(t, metrics)
	})

	t.Run("AutoRecovery", func(t *testing.T) {
		config := &DegradationConfig{
			QualityThreshold:    0.8,
			RecoveryThreshold:   0.9,
			EnableAutoRecovery:  true,
			MonitoringInterval:  50 * time.Millisecond,
		}

		manager := NewInferenceGracefulDegradationManager(config)
		ctx := context.Background()

		// Apply degradation
		inferenceState := &InferenceState{
			SessionID:      "session-2",
			CurrentQuality: 0.7,
		}
		resourceState := &ResourceState{
			AvailableMemory: 1024 * 1024 * 256, // Low memory
		}

		result, err := manager.ApplyDegradation(ctx, inferenceState, resourceState)
		require.NoError(t, err)
		assert.True(t, result.Applied)

		// Simulate resource recovery
		resourceState.AvailableMemory = 1024 * 1024 * 1024 // 1GB
		inferenceState.CurrentQuality = 0.95

		// Trigger recovery
		recovered, err := manager.AttemptRecovery(ctx, inferenceState, resourceState)
		require.NoError(t, err)
		assert.True(t, recovered)
	})
}

// TestPredictiveInferenceManager tests predictive failure detection
func TestPredictiveInferenceManager(t *testing.T) {
	t.Run("FailurePrediction", func(t *testing.T) {
		config := &PredictiveConfig{
			PredictionThreshold:       0.7,
			LookAheadWindow:           5 * time.Minute,
			ModelUpdateInterval:       time.Hour,
			MinDataPoints:             10,
			EnableProactiveScaling:    true,
			EnablePreemptiveMigration: true,
			StandbyNodeRatio:          0.1,
		}

		manager := NewInferencePredictiveManager(config)

		// Train the predictor with sample data
		for i := 0; i < 20; i++ {
			metrics := &NodeHealthMetrics{
				NodeID:          fmt.Sprintf("node-%d", i%3),
				Timestamp:       time.Now().Add(time.Duration(i) * time.Minute),
				CPUUsage:        float64(i) * 0.05,
				MemoryUsage:     float64(i) * 0.04,
				DiskUsage:       0.3,
				NetworkLatency:  time.Duration(i) * time.Millisecond,
				ErrorRate:       float64(i) * 0.01,
				ThroughputMbps:  100 - float64(i),
			}
			manager.UpdateNodeMetrics(metrics)
		}

		// Predict failure probability
		probability := manager.predictor.PredictNodeFailure("node-1")
		assert.GreaterOrEqual(t, probability, 0.0)
		assert.LessOrEqual(t, probability, 1.0)
	})

	t.Run("ProactiveMigration", func(t *testing.T) {
		config := &PredictiveConfig{
			PredictionThreshold:       0.6,
			EnablePreemptiveMigration: true,
		}

		manager := NewInferencePredictiveManager(config)
		ctx := context.Background()

		// Simulate high failure probability
		manager.predictor.failureProbabilities["node-risky"] = 0.8

		// Trigger proactive migration
		migrated := manager.proactiveReplacer.TriggerProactiveMigration(ctx, "node-risky", []string{"node-safe-1", "node-safe-2"})
		assert.True(t, migrated)
	})

	t.Run("StandbyNodeManagement", func(t *testing.T) {
		config := &PredictiveConfig{
			StandbyNodeRatio:       0.2,
			EnableProactiveScaling: true,
		}

		manager := NewInferencePredictiveManager(config)

		// Add nodes to pool
		for i := 0; i < 10; i++ {
			manager.standbyManager.AddNodeToPool(fmt.Sprintf("node-%d", i))
		}

		// Check standby allocation
		standbyCount := len(manager.standbyManager.standbyNodes)
		expectedStandby := int(10 * 0.2)
		assert.GreaterOrEqual(t, standbyCount, expectedStandby)

		// Request standby node
		nodeID := manager.standbyManager.AllocateStandbyNode()
		assert.NotEmpty(t, nodeID)
	})
}

// TestInferenceFaultToleranceCoordinator tests the main coordinator
func TestInferenceFaultToleranceCoordinator(t *testing.T) {
	t.Run("SessionManagement", func(t *testing.T) {
		config := &InferenceCoordinatorConfig{
			MaxConcurrentSessions: 10,
			SessionTimeout:        30 * time.Minute,
			RecoveryTimeout:       5 * time.Minute,
			HealthCheckInterval:   30 * time.Second,
			EnableAutoRecovery:    true,
		}

		// Create coordinator with all components
		checkpointManager := NewInferenceCheckpointManager(&CheckpointConfig{
			StorageBackend: "memory",
		})
		repartitioningManager := NewDynamicRepartitioningManager(&RepartitioningConfig{
			Strategy: "adaptive",
		}, nil)
		degradationManager := NewInferenceGracefulDegradationManager(&DegradationConfig{
			QualityThreshold: 0.8,
		})
		predictiveManager := NewInferencePredictiveManager(&PredictiveConfig{
			PredictionThreshold: 0.7,
		})

		coordinator := NewInferenceFaultToleranceCoordinator(
			config,
			checkpointManager,
			repartitioningManager,
			degradationManager,
			predictiveManager,
		)

		ctx := context.Background()
		err := coordinator.Start(ctx)
		require.NoError(t, err)

		// Create session
		err = coordinator.startSession("session-1", "model-1", map[string]interface{}{
			"batch_size": 32,
		})
		require.NoError(t, err)

		// Get session
		session := coordinator.getSession("session-1")
		assert.NotNil(t, session)
		assert.Equal(t, "session-1", session.ID)
		assert.Equal(t, SessionStatusActive, session.Status)

		// End session
		err = coordinator.endSession("session-1")
		require.NoError(t, err)
	})

	t.Run("FailureHandling", func(t *testing.T) {
		coordinator := createTestCoordinator()
		ctx := context.Background()

		// Start session
		err := coordinator.startSession("session-2", "model-2", nil)
		require.NoError(t, err)

		// Simulate failure
		err = coordinator.HandleFailure(ctx, "session-2", "node-1", "connection lost")
		require.NoError(t, err)

		// Check recovery operation
		session := coordinator.getSession("session-2")
		assert.NotNil(t, session)
		assert.Equal(t, SessionStatusRecovering, session.Status)
	})

	t.Run("IntegratedRecovery", func(t *testing.T) {
		coordinator := createTestCoordinator()
		ctx := context.Background()

		// Start coordinator
		err := coordinator.Start(ctx)
		require.NoError(t, err)

		// Create session with checkpoint
		err = coordinator.startSession("session-3", "model-3", nil)
		require.NoError(t, err)

		// Create checkpoint
		checkpoint := &InferenceCheckpoint{
			ID:        "checkpoint-1",
			SessionID: "session-3",
			ModelState: map[string]interface{}{
				"progress": 0.5,
			},
		}
		err = coordinator.checkpointManager.storeCheckpoint(ctx, checkpoint)
		require.NoError(t, err)

		// Simulate failure and recovery
		err = coordinator.HandleFailure(ctx, "session-3", "node-2", "node crashed")
		require.NoError(t, err)

		// Verify recovery strategies were applied
		operations := coordinator.activeOperations
		assert.Greater(t, len(operations), 0)
	})
}

// TestInferenceRecoveryStrategies tests different recovery strategies
func TestInferenceRecoveryStrategies(t *testing.T) {
	t.Run("CheckpointRecoveryStrategy", func(t *testing.T) {
		strategy := &CheckpointRecoveryStrategy{
			checkpointManager: NewInferenceCheckpointManager(&CheckpointConfig{
				StorageBackend: "memory",
			}),
		}

		ctx := context.Background()
		failure := &InferenceFailure{
			SessionID: "session-1",
			NodeID:    "node-1",
			ErrorType: "node_crash",
		}

		// Store a checkpoint
		checkpoint := &InferenceCheckpoint{
			ID:        "checkpoint-1",
			SessionID: "session-1",
		}
		err := strategy.checkpointManager.storeCheckpoint(ctx, checkpoint)
		require.NoError(t, err)

		// Execute recovery
		result, err := strategy.Execute(ctx, failure)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "checkpoint_recovery", result.Strategy)
	})

	t.Run("RepartitioningRecoveryStrategy", func(t *testing.T) {
		strategy := &RepartitioningRecoveryStrategy{
			repartitioningManager: NewDynamicRepartitioningManager(&RepartitioningConfig{
				Strategy: "aggressive",
			}, nil),
		}

		ctx := context.Background()
		failure := &InferenceFailure{
			SessionID:      "session-2",
			NodeID:         "node-2",
			ErrorType:      "node_failure",
			AvailableNodes: []string{"node-3", "node-4"},
		}

		result, err := strategy.Execute(ctx, failure)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, "repartitioning", result.Strategy)
	})

	t.Run("CompoundRecoveryStrategy", func(t *testing.T) {
		// Create compound strategy with multiple strategies
		checkpointStrategy := &CheckpointRecoveryStrategy{
			checkpointManager: NewInferenceCheckpointManager(&CheckpointConfig{
				StorageBackend: "memory",
			}),
		}
		repartitioningStrategy := &RepartitioningRecoveryStrategy{
			repartitioningManager: NewDynamicRepartitioningManager(&RepartitioningConfig{
				Strategy: "conservative",
			}, nil),
		}

		compoundStrategy := &CompoundRecoveryStrategy{
			strategies: []RecoveryStrategy{checkpointStrategy, repartitioningStrategy},
			mode:       "sequential",
		}

		ctx := context.Background()
		failure := &InferenceFailure{
			SessionID:      "session-3",
			NodeID:         "node-3",
			ErrorType:      "critical_failure",
			AvailableNodes: []string{"node-4", "node-5"},
		}

		result, err := compoundStrategy.Execute(ctx, failure)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestInferenceFaultToleranceMetrics tests metrics collection and reporting
func TestInferenceFaultToleranceMetrics(t *testing.T) {
	t.Run("MetricsCollection", func(t *testing.T) {
		metrics := NewInferenceFaultToleranceMetrics(time.Hour)

		// Record various metrics
		metrics.RecordCheckpoint("session-1", 100*time.Millisecond, true)
		metrics.RecordCheckpoint("session-2", 200*time.Millisecond, false)

		metrics.RecordRecovery("session-1", "checkpoint_recovery", 500*time.Millisecond, true)
		metrics.RecordRecovery("session-2", "repartitioning", 1*time.Second, true)

		metrics.RecordDegradation("session-1", "precision_reduction", 0.95, 0.85)

		metrics.RecordPrediction("node-1", 0.75, true)
		metrics.RecordPrediction("node-2", 0.3, false)

		// Get global metrics
		globalMetrics := metrics.GetGlobalMetrics()
		assert.NotNil(t, globalMetrics)
		assert.Equal(t, int64(2), globalMetrics.TotalCheckpoints)
		assert.Equal(t, int64(2), globalMetrics.TotalRecoveries)
		assert.Equal(t, float64(1.0), globalMetrics.RecoverySuccessRate)
	})

	t.Run("PrometheusMetrics", func(t *testing.T) {
		metrics := NewInferenceFaultToleranceMetrics(time.Hour)

		// Initialize Prometheus metrics
		metrics.InitPrometheus()

		// Record metrics
		metrics.RecordCheckpoint("session-3", 150*time.Millisecond, true)
		metrics.RecordRecovery("session-3", "graceful_degradation", 300*time.Millisecond, true)

		// Verify Prometheus metrics were updated
		assert.NotNil(t, metrics.prometheusMetrics)
	})

	t.Run("ResilienceMonitoring", func(t *testing.T) {
		metrics := NewInferenceFaultToleranceMetrics(time.Hour)

		// Update resilience metrics
		metrics.UpdateResilienceScore(0.95)
		metrics.UpdateSessionAvailability("session-4", 0.99)

		// Get dashboard data
		dashboard := metrics.GetDashboardData()
		assert.NotNil(t, dashboard)
		assert.Equal(t, "healthy", dashboard.SystemStatus.Overall)
	})
}

// Helper function to create a test coordinator
func createTestCoordinator() *InferenceFaultToleranceCoordinator {
	config := &InferenceCoordinatorConfig{
		MaxConcurrentSessions: 5,
		SessionTimeout:        10 * time.Minute,
		RecoveryTimeout:       1 * time.Minute,
		EnableAutoRecovery:    true,
	}

	checkpointManager := NewInferenceCheckpointManager(&CheckpointConfig{
		StorageBackend: "memory",
	})
	repartitioningManager := NewDynamicRepartitioningManager(&RepartitioningConfig{
		Strategy: "adaptive",
	}, nil)
	degradationManager := NewInferenceGracefulDegradationManager(&DegradationConfig{
		QualityThreshold: 0.8,
	})
	predictiveManager := NewInferencePredictiveManager(&PredictiveConfig{
		PredictionThreshold: 0.7,
	})

	return NewInferenceFaultToleranceCoordinator(
		config,
		checkpointManager,
		repartitioningManager,
		degradationManager,
		predictiveManager,
	)
}

// BenchmarkCheckpointing benchmarks checkpoint creation and restoration
func BenchmarkCheckpointing(b *testing.B) {
	config := &CheckpointConfig{
		StorageBackend:     "memory",
		CompressionEnabled: true,
	}
	manager := NewInferenceCheckpointManager(config)
	ctx := context.Background()

	checkpoint := &InferenceCheckpoint{
		ID:        "bench-checkpoint",
		SessionID: "bench-session",
		ModelState: map[string]interface{}{
			"weights": make([]float32, 1000000), // 1M parameters
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.storeCheckpoint(ctx, checkpoint)
		_, _ = manager.getCheckpoint("bench-checkpoint")
	}
}

// BenchmarkRepartitioning benchmarks repartitioning operations
func BenchmarkRepartitioning(b *testing.B) {
	config := &RepartitioningConfig{
		Strategy:         "adaptive",
		MaxPartitionSize: 1024 * 1024,
	}
	manager := NewDynamicRepartitioningManager(config, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.HandleNodeFailure(ctx, "failed-node",
			[]string{"node-1", "node-2", "node-3"},
			[]string{"shard-1", "shard-2", "shard-3", "shard-4"})
	}
}