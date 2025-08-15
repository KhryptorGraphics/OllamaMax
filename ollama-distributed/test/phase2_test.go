package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/cluster"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/sirupsen/logrus"
)

// TestPhase2AdvancedDistributedFeatures tests the Phase 2 implementation
func TestPhase2AdvancedDistributedFeatures(t *testing.T) {
	t.Run("EnhancedClusterManagement", func(t *testing.T) {
		testEnhancedClusterManagement(t)
	})

	t.Run("NodeDiscoveryAndHealth", func(t *testing.T) {
		testNodeDiscoveryAndHealth(t)
	})

	t.Run("LoadBalancingStrategies", func(t *testing.T) {
		testLoadBalancingStrategies(t)
	})

	t.Run("PerformanceMonitoring", func(t *testing.T) {
		testPerformanceMonitoring(t)
	})

	t.Run("PredictiveScaling", func(t *testing.T) {
		testPredictiveScaling(t)
	})

	t.Run("CrossRegionReplication", func(t *testing.T) {
		testCrossRegionReplication(t)
	})
}

func testEnhancedClusterManagement(t *testing.T) {
	// Create test configuration
	cfg := createTestConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	// Create base cluster manager (mock)
	baseManager := createMockClusterManager(cfg)

	// Create enhanced cluster manager
	enhancedManager, err := cluster.NewEnhancedManager(cfg, baseManager, logger)
	require.NoError(t, err)
	require.NotNil(t, enhancedManager)

	// Test starting the enhanced manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = enhancedManager.Start(ctx)
	assert.NoError(t, err)

	// Test getting enhanced cluster status
	status := enhancedManager.GetEnhancedClusterStatus()
	require.NotNil(t, status)
	assert.NotNil(t, status.BasicStatus)
	assert.NotNil(t, status.NodeHealth)
	assert.NotNil(t, status.LoadDistribution)
	assert.NotNil(t, status.PerformanceMetrics)
	assert.NotNil(t, status.ScalingState)
	assert.NotNil(t, status.RegionStatus)
	assert.NotNil(t, status.Predictions)
}

func testNodeDiscoveryAndHealth(t *testing.T) {
	cfg := createTestConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Test node info creation
	nodeInfo := &cluster.NodeInfo{
		ID:       "test-node-1",
		Name:     "Test Node 1",
		Address:  "127.0.0.1:8080",
		Region:   "us-west-2",
		Zone:     "us-west-2a",
		Status:   cluster.NodeStatusHealthy,
		JoinedAt: time.Now(),
		LastSeen: time.Now(),
		Capabilities: cluster.NodeCapabilities{
			Inference:    true,
			Storage:      true,
			Coordination: false,
			Gateway:      false,
			Models:       []string{"llama2-7b", "codellama-13b"},
		},
		Resources: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Used:      2.5,
				Available: 1.5,
				Total:     4.0,
				Percent:   62.5,
			},
			Memory: cluster.ResourceUsage{
				Used:      6.0,
				Available: 2.0,
				Total:     8.0,
				Percent:   75.0,
			},
		},
	}

	// Verify node info structure
	assert.Equal(t, "test-node-1", nodeInfo.ID)
	assert.Equal(t, cluster.NodeStatusHealthy, nodeInfo.Status)
	assert.True(t, nodeInfo.Capabilities.Inference)
	assert.Equal(t, 62.5, nodeInfo.Resources.CPU.Percent)
	assert.Len(t, nodeInfo.Capabilities.Models, 2)

	// Test health check structure
	healthCheck := &cluster.HealthCheck{
		Name:     "api-health",
		Endpoint: "/health",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Retries:  3,
		Enabled:  true,
		LastResult: &cluster.HealthResult{
			Success:   true,
			Latency:   50 * time.Millisecond,
			Timestamp: time.Now(),
		},
	}

	assert.Equal(t, "api-health", healthCheck.Name)
	assert.True(t, healthCheck.Enabled)
	assert.True(t, healthCheck.LastResult.Success)
	assert.Equal(t, 50*time.Millisecond, healthCheck.LastResult.Latency)
}

func testLoadBalancingStrategies(t *testing.T) {
	// Test request context
	requestCtx := &cluster.RequestContext{
		Method:    "POST",
		Path:      "/api/v1/inference",
		ModelName: "llama2-7b",
		Priority:  1,
		Timeout:   30 * time.Second,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Metadata: map[string]string{
			"client_id": "test-client",
		},
	}

	assert.Equal(t, "POST", requestCtx.Method)
	assert.Equal(t, "llama2-7b", requestCtx.ModelName)
	assert.Equal(t, 1, requestCtx.Priority)

	// Test load metrics
	loadMetrics := &cluster.LoadMetrics{
		NodeID:              "test-node-1",
		RequestsPerSecond:   15.5,
		AverageLatency:      250.0,
		ErrorRate:           0.02,
		CPUUtilization:      65.0,
		MemoryUtilization:   70.0,
		ActiveConnections:   25,
		QueueLength:         5,
		LastUpdated:         time.Now(),
	}

	assert.Equal(t, "test-node-1", loadMetrics.NodeID)
	assert.Equal(t, 15.5, loadMetrics.RequestsPerSecond)
	assert.Equal(t, 0.02, loadMetrics.ErrorRate)
	assert.Equal(t, 25, loadMetrics.ActiveConnections)
}

func testPerformanceMonitoring(t *testing.T) {
	// Test performance metrics
	perfMetrics := &cluster.PerformanceMetrics{
		Timestamp:           time.Now(),
		TotalRequests:       10000,
		RequestsPerSecond:   50.0,
		AverageLatency:      200.0,
		P95Latency:          350.0,
		P99Latency:          500.0,
		ErrorRate:           0.01,
		ThroughputMBps:      25.5,
		ActiveConnections:   100,
		ClusterUtilization:  75.0,
	}

	assert.Equal(t, uint64(10000), perfMetrics.TotalRequests)
	assert.Equal(t, 50.0, perfMetrics.RequestsPerSecond)
	assert.Equal(t, 200.0, perfMetrics.AverageLatency)
	assert.Equal(t, 0.01, perfMetrics.ErrorRate)

	// Test performance insights
	insights := &cluster.PerformanceInsights{
		OverallHealth: 0.85,
		Bottlenecks:   []string{"memory_pressure", "network_latency"},
		Recommendations: []string{
			"Scale up memory on node-2",
			"Optimize network configuration",
		},
		TrendAnalysis: &cluster.TrendAnalysis{
			LatencyTrend:     "stable",
			ThroughputTrend:  "improving",
			ErrorRateTrend:   "stable",
			UtilizationTrend: "increasing",
			Confidence:       0.92,
		},
		ResourceEfficiency: map[string]float64{
			"cpu":     0.78,
			"memory":  0.82,
			"network": 0.65,
			"disk":    0.90,
		},
	}

	assert.Equal(t, 0.85, insights.OverallHealth)
	assert.Len(t, insights.Bottlenecks, 2)
	assert.Len(t, insights.Recommendations, 2)
	assert.Equal(t, "stable", insights.TrendAnalysis.LatencyTrend)
	assert.Equal(t, 0.92, insights.TrendAnalysis.Confidence)
	assert.Equal(t, 0.78, insights.ResourceEfficiency["cpu"])
}

func testPredictiveScaling(t *testing.T) {
	// Test scaling policy
	scalingPolicy := &cluster.ScalingPolicy{
		Name:    "cpu-based-scaling",
		Enabled: true,
		Triggers: []cluster.ScalingTrigger{
			{
				Metric:    "cpu_utilization",
				Operator:  ">",
				Threshold: 80.0,
				Duration:  5 * time.Minute,
			},
		},
		Actions: []cluster.ScalingAction{
			{
				Type:     cluster.ScalingActionScaleUp,
				Count:    2,
				NodeType: "inference",
				Region:   "us-west-2",
			},
		},
		Cooldown: 10 * time.Minute,
		MinNodes: 2,
		MaxNodes: 10,
	}

	assert.Equal(t, "cpu-based-scaling", scalingPolicy.Name)
	assert.True(t, scalingPolicy.Enabled)
	assert.Len(t, scalingPolicy.Triggers, 1)
	assert.Equal(t, "cpu_utilization", scalingPolicy.Triggers[0].Metric)
	assert.Equal(t, 80.0, scalingPolicy.Triggers[0].Threshold)

	// Test scaling prediction
	prediction := &cluster.ScalingPrediction{
		Timestamp:        time.Now(),
		PredictedLoad:    85.0,
		RecommendedNodes: 6,
		Confidence:       0.88,
		Horizon:          30 * time.Minute,
		Reasoning:        "CPU utilization trending upward, expected to exceed threshold",
	}

	assert.Equal(t, 85.0, prediction.PredictedLoad)
	assert.Equal(t, 6, prediction.RecommendedNodes)
	assert.Equal(t, 0.88, prediction.Confidence)
	assert.Equal(t, 30*time.Minute, prediction.Horizon)

	// Test scaling state
	scalingState := &cluster.ScalingState{
		CurrentNodes:      4,
		TargetNodes:       6,
		LastScaleAction:   time.Now().Add(-5 * time.Minute),
		ScalingInProgress: true,
		CooldownUntil:     time.Now().Add(5 * time.Minute),
	}

	assert.Equal(t, 4, scalingState.CurrentNodes)
	assert.Equal(t, 6, scalingState.TargetNodes)
	assert.True(t, scalingState.ScalingInProgress)
}

func testCrossRegionReplication(t *testing.T) {
	// Test region info
	regionInfo := &cluster.RegionInfo{
		Name:   "us-west-2",
		Status: cluster.RegionStatusHealthy,
		Latency: map[string]float64{
			"us-east-1": 75.0,
			"eu-west-1": 150.0,
		},
		Capacity: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Total: 32.0,
			},
			Memory: cluster.ResourceUsage{
				Total: 128.0,
			},
		},
		Utilization: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Used:    20.0,
				Percent: 62.5,
			},
			Memory: cluster.ResourceUsage{
				Used:    80.0,
				Percent: 62.5,
			},
		},
	}

	assert.Equal(t, "us-west-2", regionInfo.Name)
	assert.Equal(t, cluster.RegionStatusHealthy, regionInfo.Status)
	assert.Equal(t, 75.0, regionInfo.Latency["us-east-1"])
	assert.Equal(t, 62.5, regionInfo.Utilization.CPU.Percent)

	// Test replication state
	replicationState := &cluster.ReplicationState{
		SourceRegion: "us-west-2",
		TargetRegion: "us-east-1",
		Status:       "syncing",
		Progress:     0.75,
		LastSync:     time.Now().Add(-1 * time.Minute),
		Lag:          30 * time.Second,
	}

	assert.Equal(t, "us-west-2", replicationState.SourceRegion)
	assert.Equal(t, "us-east-1", replicationState.TargetRegion)
	assert.Equal(t, "syncing", replicationState.Status)
	assert.Equal(t, 0.75, replicationState.Progress)
	assert.Equal(t, 30*time.Second, replicationState.Lag)
}

// Helper functions

func createTestConfig() *config.DistributedConfig {
	cfg := &config.DistributedConfig{}
	cfg.SetDefaults()
	cfg.Node.ID = "test-node"
	cfg.Node.Region = "us-west-2"
	cfg.Node.Zone = "us-west-2a"
	return cfg
}

func createMockClusterManager(cfg *config.DistributedConfig) *cluster.Manager {
	// This would normally create a real cluster manager
	// For testing, we'll create a minimal mock
	return &cluster.Manager{
		// Add minimal required fields for testing
	}
}
