package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/cluster"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
)

// TestPhase2BasicTypes tests the Phase 2 type definitions and structures
func TestPhase2BasicTypes(t *testing.T) {
	t.Run("NodeInfoStructure", func(t *testing.T) {
		testNodeInfoStructure(t)
	})

	t.Run("HealthCheckSystem", func(t *testing.T) {
		testHealthCheckSystem(t)
	})

	t.Run("LoadBalancingTypes", func(t *testing.T) {
		testLoadBalancingTypes(t)
	})

	t.Run("PerformanceMetrics", func(t *testing.T) {
		testPerformanceMetrics(t)
	})

	t.Run("ScalingPolicies", func(t *testing.T) {
		testScalingPolicies(t)
	})

	t.Run("RegionManagement", func(t *testing.T) {
		testRegionManagement(t)
	})
}

func testNodeInfoStructure(t *testing.T) {
	// Test comprehensive node information
	nodeInfo := &cluster.NodeInfo{
		ID:       "node-001",
		Name:     "Primary Inference Node",
		Address:  "10.0.1.100:8080",
		Region:   "us-west-2",
		Zone:     "us-west-2a",
		Status:   cluster.NodeStatusHealthy,
		JoinedAt: time.Now().Add(-2 * time.Hour),
		LastSeen: time.Now().Add(-30 * time.Second),
		Capabilities: cluster.NodeCapabilities{
			Inference:    true,
			Storage:      true,
			Coordination: false,
			Gateway:      true,
			Models:       []string{"llama2-7b", "llama2-13b", "codellama-7b"},
		},
		Resources: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Used:      6.5,
				Available: 1.5,
				Total:     8.0,
				Percent:   81.25,
			},
			Memory: cluster.ResourceUsage{
				Used:      12.0,
				Available: 4.0,
				Total:     16.0,
				Percent:   75.0,
			},
			GPU: cluster.ResourceUsage{
				Used:      1.0,
				Available: 1.0,
				Total:     2.0,
				Percent:   50.0,
			},
			Disk: cluster.ResourceUsage{
				Used:      450.0,
				Available: 550.0,
				Total:     1000.0,
				Percent:   45.0,
			},
			Network: cluster.NetworkUsage{
				BytesIn:    1024 * 1024 * 100, // 100 MB
				BytesOut:   1024 * 1024 * 80,  // 80 MB
				PacketsIn:  50000,
				PacketsOut: 45000,
			},
		},
		Metadata: map[string]string{
			"instance_type": "c5.2xlarge",
			"ami_id":        "ami-12345678",
			"environment":   "production",
		},
	}

	// Verify node info structure
	assert.Equal(t, "node-001", nodeInfo.ID)
	assert.Equal(t, "Primary Inference Node", nodeInfo.Name)
	assert.Equal(t, cluster.NodeStatusHealthy, nodeInfo.Status)
	assert.Equal(t, "us-west-2", nodeInfo.Region)
	assert.Equal(t, "us-west-2a", nodeInfo.Zone)

	// Verify capabilities
	assert.True(t, nodeInfo.Capabilities.Inference)
	assert.True(t, nodeInfo.Capabilities.Storage)
	assert.False(t, nodeInfo.Capabilities.Coordination)
	assert.True(t, nodeInfo.Capabilities.Gateway)
	assert.Len(t, nodeInfo.Capabilities.Models, 3)
	assert.Contains(t, nodeInfo.Capabilities.Models, "llama2-7b")

	// Verify resource usage
	assert.Equal(t, 81.25, nodeInfo.Resources.CPU.Percent)
	assert.Equal(t, 75.0, nodeInfo.Resources.Memory.Percent)
	assert.Equal(t, 50.0, nodeInfo.Resources.GPU.Percent)
	assert.Equal(t, 45.0, nodeInfo.Resources.Disk.Percent)

	// Verify network usage
	assert.Equal(t, uint64(1024*1024*100), nodeInfo.Resources.Network.BytesIn)
	assert.Equal(t, uint64(50000), nodeInfo.Resources.Network.PacketsIn)

	// Verify metadata
	assert.Equal(t, "c5.2xlarge", nodeInfo.Metadata["instance_type"])
	assert.Equal(t, "production", nodeInfo.Metadata["environment"])
}

func testHealthCheckSystem(t *testing.T) {
	// Test health check configuration
	healthCheck := &cluster.HealthCheck{
		Name:     "inference-api-health",
		Endpoint: "/api/v1/health",
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
		Retries:  3,
		Enabled:  true,
		LastResult: &cluster.HealthResult{
			Success:    true,
			Latency:    45 * time.Millisecond,
			Timestamp:  time.Now(),
			StatusCode: 200,
			Response:   `{"status":"healthy","uptime":"2h30m"}`,
		},
	}

	assert.Equal(t, "inference-api-health", healthCheck.Name)
	assert.Equal(t, "/api/v1/health", healthCheck.Endpoint)
	assert.Equal(t, 30*time.Second, healthCheck.Interval)
	assert.Equal(t, 5*time.Second, healthCheck.Timeout)
	assert.Equal(t, 3, healthCheck.Retries)
	assert.True(t, healthCheck.Enabled)

	// Verify health result
	require.NotNil(t, healthCheck.LastResult)
	assert.True(t, healthCheck.LastResult.Success)
	assert.Equal(t, 45*time.Millisecond, healthCheck.LastResult.Latency)
	assert.Equal(t, 200, healthCheck.LastResult.StatusCode)
	assert.Contains(t, healthCheck.LastResult.Response, "healthy")

	// Test alert creation
	alert := &cluster.Alert{
		ID:       "alert-001",
		Type:     cluster.AlertTypeHighLatency,
		Severity: cluster.AlertSeverityWarning,
		Title:    "High API Latency Detected",
		Message:  "API response time exceeded 500ms threshold",
		NodeID:   "node-001",
		Metadata: map[string]string{
			"threshold": "500ms",
			"actual":    "750ms",
			"endpoint":  "/api/v1/inference",
		},
		CreatedAt: time.Now(),
	}

	assert.Equal(t, "alert-001", alert.ID)
	assert.Equal(t, cluster.AlertTypeHighLatency, alert.Type)
	assert.Equal(t, cluster.AlertSeverityWarning, alert.Severity)
	assert.Equal(t, "node-001", alert.NodeID)
	assert.Equal(t, "500ms", alert.Metadata["threshold"])
	assert.Nil(t, alert.ResolvedAt)
}

func testLoadBalancingTypes(t *testing.T) {
	// Test request context
	requestCtx := &cluster.RequestContext{
		Method:    "POST",
		Path:      "/api/v1/inference",
		ModelName: "llama2-13b",
		Priority:  2,
		Timeout:   60 * time.Second,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
			"User-Agent":    "ollama-client/1.0",
		},
		Metadata: map[string]string{
			"client_id":    "client-001",
			"session_id":   "session-123",
			"request_size": "2048",
		},
	}

	assert.Equal(t, "POST", requestCtx.Method)
	assert.Equal(t, "/api/v1/inference", requestCtx.Path)
	assert.Equal(t, "llama2-13b", requestCtx.ModelName)
	assert.Equal(t, 2, requestCtx.Priority)
	assert.Equal(t, 60*time.Second, requestCtx.Timeout)
	assert.Equal(t, "application/json", requestCtx.Headers["Content-Type"])
	assert.Equal(t, "client-001", requestCtx.Metadata["client_id"])

	// Test load metrics
	loadMetrics := &cluster.LoadMetrics{
		NodeID:              "node-001",
		RequestsPerSecond:   25.5,
		AverageLatency:      180.0,
		ErrorRate:           0.015,
		CPUUtilization:      78.5,
		MemoryUtilization:   82.0,
		ActiveConnections:   45,
		QueueLength:         8,
		LastUpdated:         time.Now(),
	}

	assert.Equal(t, "node-001", loadMetrics.NodeID)
	assert.Equal(t, 25.5, loadMetrics.RequestsPerSecond)
	assert.Equal(t, 180.0, loadMetrics.AverageLatency)
	assert.Equal(t, 0.015, loadMetrics.ErrorRate)
	assert.Equal(t, 78.5, loadMetrics.CPUUtilization)
	assert.Equal(t, 82.0, loadMetrics.MemoryUtilization)
	assert.Equal(t, 45, loadMetrics.ActiveConnections)
	assert.Equal(t, 8, loadMetrics.QueueLength)
}

func testPerformanceMetrics(t *testing.T) {
	// Test comprehensive performance metrics
	perfMetrics := &cluster.PerformanceMetrics{
		Timestamp:           time.Now(),
		TotalRequests:       50000,
		RequestsPerSecond:   125.5,
		AverageLatency:      150.0,
		P95Latency:          280.0,
		P99Latency:          450.0,
		ErrorRate:           0.008,
		ThroughputMBps:      45.2,
		ActiveConnections:   200,
		ClusterUtilization:  68.5,
	}

	assert.Equal(t, uint64(50000), perfMetrics.TotalRequests)
	assert.Equal(t, 125.5, perfMetrics.RequestsPerSecond)
	assert.Equal(t, 150.0, perfMetrics.AverageLatency)
	assert.Equal(t, 280.0, perfMetrics.P95Latency)
	assert.Equal(t, 450.0, perfMetrics.P99Latency)
	assert.Equal(t, 0.008, perfMetrics.ErrorRate)
	assert.Equal(t, 45.2, perfMetrics.ThroughputMBps)
	assert.Equal(t, 200, perfMetrics.ActiveConnections)
	assert.Equal(t, 68.5, perfMetrics.ClusterUtilization)

	// Test performance insights
	insights := &cluster.PerformanceInsights{
		OverallHealth: 0.92,
		Bottlenecks:   []string{"memory_pressure_node_2", "network_latency_region_cross"},
		Recommendations: []string{
			"Scale up memory on node-2",
			"Optimize cross-region network routing",
			"Enable request caching for frequent queries",
		},
		TrendAnalysis: &cluster.TrendAnalysis{
			LatencyTrend:     "stable",
			ThroughputTrend:  "improving",
			ErrorRateTrend:   "improving",
			UtilizationTrend: "increasing",
			Confidence:       0.89,
		},
		ResourceEfficiency: map[string]float64{
			"cpu":     0.85,
			"memory":  0.78,
			"gpu":     0.92,
			"network": 0.71,
			"disk":    0.88,
		},
		PredictedIssues: []*cluster.PredictedIssue{
			{
				Type:        "resource_exhaustion",
				Severity:    "warning",
				Description: "Node-2 memory utilization will exceed 90% in 2 hours",
				ETA:         time.Now().Add(2 * time.Hour),
				Confidence:  0.87,
				Mitigation:  "Scale up node-2 or redistribute load",
			},
		},
	}

	assert.Equal(t, 0.92, insights.OverallHealth)
	assert.Len(t, insights.Bottlenecks, 2)
	assert.Len(t, insights.Recommendations, 3)
	assert.Equal(t, "stable", insights.TrendAnalysis.LatencyTrend)
	assert.Equal(t, "improving", insights.TrendAnalysis.ThroughputTrend)
	assert.Equal(t, 0.89, insights.TrendAnalysis.Confidence)
	assert.Equal(t, 0.85, insights.ResourceEfficiency["cpu"])
	assert.Len(t, insights.PredictedIssues, 1)
	assert.Equal(t, "resource_exhaustion", insights.PredictedIssues[0].Type)
}

func testScalingPolicies(t *testing.T) {
	// Test comprehensive scaling policy
	scalingPolicy := &cluster.ScalingPolicy{
		Name:    "adaptive-cpu-memory-scaling",
		Enabled: true,
		Triggers: []cluster.ScalingTrigger{
			{
				Metric:    "cpu_utilization",
				Operator:  ">",
				Threshold: 75.0,
				Duration:  3 * time.Minute,
			},
			{
				Metric:    "memory_utilization",
				Operator:  ">",
				Threshold: 80.0,
				Duration:  3 * time.Minute,
			},
		},
		Actions: []cluster.ScalingAction{
			{
				Type:     cluster.ScalingActionScaleUp,
				Count:    2,
				NodeType: "inference",
				Region:   "us-west-2",
				Zone:     "us-west-2a",
			},
		},
		Cooldown: 15 * time.Minute,
		MinNodes: 3,
		MaxNodes: 20,
	}

	assert.Equal(t, "adaptive-cpu-memory-scaling", scalingPolicy.Name)
	assert.True(t, scalingPolicy.Enabled)
	assert.Len(t, scalingPolicy.Triggers, 2)
	assert.Equal(t, "cpu_utilization", scalingPolicy.Triggers[0].Metric)
	assert.Equal(t, ">", scalingPolicy.Triggers[0].Operator)
	assert.Equal(t, 75.0, scalingPolicy.Triggers[0].Threshold)
	assert.Equal(t, 3*time.Minute, scalingPolicy.Triggers[0].Duration)
	assert.Len(t, scalingPolicy.Actions, 1)
	assert.Equal(t, cluster.ScalingActionScaleUp, scalingPolicy.Actions[0].Type)
	assert.Equal(t, 2, scalingPolicy.Actions[0].Count)
	assert.Equal(t, 15*time.Minute, scalingPolicy.Cooldown)
	assert.Equal(t, 3, scalingPolicy.MinNodes)
	assert.Equal(t, 20, scalingPolicy.MaxNodes)

	// Test scaling prediction
	prediction := &cluster.ScalingPrediction{
		Timestamp:        time.Now(),
		PredictedLoad:    88.5,
		RecommendedNodes: 8,
		Confidence:       0.91,
		Horizon:          45 * time.Minute,
		Reasoning:        "CPU and memory utilization trending upward, expected peak in 45 minutes based on historical patterns",
	}

	assert.Equal(t, 88.5, prediction.PredictedLoad)
	assert.Equal(t, 8, prediction.RecommendedNodes)
	assert.Equal(t, 0.91, prediction.Confidence)
	assert.Equal(t, 45*time.Minute, prediction.Horizon)
	assert.Contains(t, prediction.Reasoning, "trending upward")
}

func testRegionManagement(t *testing.T) {
	// Test region information
	regionInfo := &cluster.RegionInfo{
		Name:   "us-west-2",
		Status: cluster.RegionStatusHealthy,
		Latency: map[string]float64{
			"us-east-1":    78.5,
			"eu-west-1":    145.2,
			"ap-southeast": 185.7,
		},
		Capacity: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Total: 64.0,
			},
			Memory: cluster.ResourceUsage{
				Total: 256.0,
			},
			GPU: cluster.ResourceUsage{
				Total: 16.0,
			},
		},
		Utilization: cluster.ResourceInfo{
			CPU: cluster.ResourceUsage{
				Used:    45.5,
				Percent: 71.1,
			},
			Memory: cluster.ResourceUsage{
				Used:    180.0,
				Percent: 70.3,
			},
			GPU: cluster.ResourceUsage{
				Used:    12.0,
				Percent: 75.0,
			},
		},
	}

	assert.Equal(t, "us-west-2", regionInfo.Name)
	assert.Equal(t, cluster.RegionStatusHealthy, regionInfo.Status)
	assert.Equal(t, 78.5, regionInfo.Latency["us-east-1"])
	assert.Equal(t, 145.2, regionInfo.Latency["eu-west-1"])
	assert.Equal(t, 64.0, regionInfo.Capacity.CPU.Total)
	assert.Equal(t, 71.1, regionInfo.Utilization.CPU.Percent)

	// Test replication state
	replicationState := &cluster.ReplicationState{
		SourceRegion: "us-west-2",
		TargetRegion: "us-east-1",
		Status:       "active",
		Progress:     0.95,
		LastSync:     time.Now().Add(-2 * time.Minute),
		Lag:          45 * time.Second,
	}

	assert.Equal(t, "us-west-2", replicationState.SourceRegion)
	assert.Equal(t, "us-east-1", replicationState.TargetRegion)
	assert.Equal(t, "active", replicationState.Status)
	assert.Equal(t, 0.95, replicationState.Progress)
	assert.Equal(t, 45*time.Second, replicationState.Lag)
}

func TestConfigurationIntegration(t *testing.T) {
	// Test that our enhanced cluster types work with the configuration system
	cfg := &config.DistributedConfig{}
	cfg.SetDefaults()
	cfg.Node.ID = "enhanced-test-node"
	cfg.Node.Region = "us-west-2"
	cfg.Node.Zone = "us-west-2a"

	// Verify configuration integration
	assert.Equal(t, "enhanced-test-node", cfg.Node.ID)
	assert.Equal(t, "us-west-2", cfg.Node.Region)
	assert.Equal(t, "us-west-2a", cfg.Node.Zone)

	// Test node capabilities configuration
	cfg.Node.Capabilities.Inference = true
	cfg.Node.Capabilities.Storage = true
	cfg.Node.Capabilities.Gateway = false

	assert.True(t, cfg.Node.Capabilities.Inference)
	assert.True(t, cfg.Node.Capabilities.Storage)
	assert.False(t, cfg.Node.Capabilities.Gateway)
}
