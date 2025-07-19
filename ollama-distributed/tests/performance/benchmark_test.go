package performance

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ollama/ollama-distributed/pkg/api"
	"github.com/ollama/ollama-distributed/tests/integration"
)

// BenchmarkInferenceLatency benchmarks inference latency
func BenchmarkInferenceLatency(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test different prompt sizes
	testCases := []struct {
		name      string
		prompt    string
		maxTokens int
	}{
		{"Small", "Hello", 20},
		{"Medium", "Explain quantum computing", 100},
		{"Large", "Write a detailed explanation of machine learning algorithms and their applications", 500},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			req := &api.InferenceRequest{
				Model:  "llama3.2:1b",
				Prompt: tc.prompt,
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  tc.maxTokens,
				},
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					start := time.Now()
					_, err := leader.ProcessInference(context.Background(), req)
					if err != nil {
						b.Errorf("Inference failed: %v", err)
					}
					latency := time.Since(start)
					b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
				}
			})
		})
	}
}

// BenchmarkInferenceThroughput benchmarks inference throughput
func BenchmarkInferenceThroughput(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test different concurrency levels
	concurrencyLevels := []int{1, 5, 10, 20, 50}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency%d", concurrency), func(b *testing.B) {
			req := &api.InferenceRequest{
				Model:  "llama3.2:1b",
				Prompt: "Benchmark throughput test",
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  50,
				},
			}

			b.ResetTimer()
			b.SetParallelism(concurrency)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := leader.ProcessInference(context.Background(), req)
					if err != nil {
						b.Errorf("Inference failed: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkDistributedInference benchmarks distributed inference
func BenchmarkDistributedInference(b *testing.B) {
	// Test with different cluster sizes
	clusterSizes := []int{1, 3, 5}

	for _, size := range clusterSizes {
		b.Run(fmt.Sprintf("Cluster%d", size), func(b *testing.B) {
			cluster := setupBenchmarkCluster(b, size)
			defer cluster.Shutdown()

			leader := cluster.GetLeader()
			require.NotNil(b, leader)

			// Use large model to trigger distribution
			req := &api.InferenceRequest{
				Model:  "llama3.2:8b",
				Prompt: "Benchmark distributed inference across multiple nodes",
				Options: map[string]interface{}{
					"temperature": 0.5,
					"max_tokens":  200,
				},
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := leader.ProcessInference(context.Background(), req)
					if err != nil {
						b.Errorf("Distributed inference failed: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkModelSync benchmarks model synchronization
func BenchmarkModelSync(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test different model sizes
	modelSizes := []int64{
		100 * 1024 * 1024,  // 100MB
		500 * 1024 * 1024,  // 500MB
		1024 * 1024 * 1024, // 1GB
	}

	for _, size := range modelSizes {
		b.Run(fmt.Sprintf("Model%dMB", size/(1024*1024)), func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				modelInfo := &integration.ModelInfo{
					Name:              fmt.Sprintf("benchmark-model-%d-%d", size, i),
					Size:              size,
					Checksum:          fmt.Sprintf("checksum-%d-%d", size, i),
					ReplicationFactor: 2,
					LastAccessed:      time.Now(),
					Popularity:        rand.Float64(),
				}

				start := time.Now()
				err := leader.RegisterModel(modelInfo)
				if err != nil {
					b.Errorf("Model registration failed: %v", err)
				}
				
				// Wait for replication
				time.Sleep(time.Duration(size/1024/1024/10) * time.Millisecond) // Rough estimate
				
				syncTime := time.Since(start)
				b.ReportMetric(float64(syncTime.Nanoseconds()), "ns/op")
			}
		})
	}
}

// BenchmarkConsensusOperations benchmarks consensus operations
func BenchmarkConsensusOperations(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test different operation types
	testCases := []struct {
		name      string
		operation func(int) (string, string)
	}{
		{
			name: "SimpleKV",
			operation: func(i int) (string, string) {
				return fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i)
			},
		},
		{
			name: "LargeValue",
			operation: func(i int) (string, string) {
				value := make([]byte, 1024) // 1KB value
				for j := range value {
					value[j] = byte(i % 256)
				}
				return fmt.Sprintf("large-key-%d", i), string(value)
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key, value := tc.operation(i)
					start := time.Now()
					err := leader.ApplyConsensusOperation(key, value)
					if err != nil {
						b.Errorf("Consensus operation failed: %v", err)
					}
					duration := time.Since(start)
					b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
					i++
				}
			})
		})
	}
}

// BenchmarkLoadBalancing benchmarks load balancing performance
func BenchmarkLoadBalancing(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 5)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test load balancing with varying request patterns
	testCases := []struct {
		name        string
		requests    int
		concurrency int
	}{
		{"LowLoad", 100, 5},
		{"MediumLoad", 500, 10},
		{"HighLoad", 1000, 20},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				start := time.Now()
				
				// Run load balancing test
				runLoadBalancingTest(b, leader, tc.requests, tc.concurrency)
				
				duration := time.Since(start)
				b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
			}
		})
	}
}

// BenchmarkFaultTolerance benchmarks fault tolerance performance
func BenchmarkFaultTolerance(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 5)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		// Simulate node failure and recovery
		nodes := cluster.GetActiveNodes()
		if len(nodes) > 2 {
			// Fail a non-leader node
			var nodeToFail *integration.TestNode
			for _, node := range nodes {
				if !node.IsLeader() {
					nodeToFail = node
					break
				}
			}
			
			if nodeToFail != nil {
				nodeToFail.Shutdown()
				
				// Wait for failure detection
				time.Sleep(5 * time.Second)
				
				// Test that system is still operational
				req := &api.InferenceRequest{
					Model:  "llama3.2:1b",
					Prompt: "Fault tolerance test",
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  20,
					},
				}
				
				_, err := leader.ProcessInference(context.Background(), req)
				if err != nil {
					b.Errorf("Inference failed after node failure: %v", err)
				}
			}
		}
		
		duration := time.Since(start)
		b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	// Test different memory usage patterns
	testCases := []struct {
		name      string
		models    int
		requests  int
	}{
		{"SmallFootprint", 1, 10},
		{"MediumFootprint", 5, 50},
		{"LargeFootprint", 10, 100},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Register models
				for j := 0; j < tc.models; j++ {
					modelInfo := &integration.ModelInfo{
						Name:              fmt.Sprintf("memory-test-model-%d-%d", i, j),
						Size:              100 * 1024 * 1024, // 100MB
						Checksum:          fmt.Sprintf("checksum-%d-%d", i, j),
						ReplicationFactor: 2,
						LastAccessed:      time.Now(),
						Popularity:        rand.Float64(),
					}
					
					err := leader.RegisterModel(modelInfo)
					if err != nil {
						b.Errorf("Model registration failed: %v", err)
					}
				}
				
				// Generate inference requests
				for j := 0; j < tc.requests; j++ {
					req := &api.InferenceRequest{
						Model:  fmt.Sprintf("memory-test-model-%d-%d", i, j%tc.models),
						Prompt: fmt.Sprintf("Memory test request %d", j),
						Options: map[string]interface{}{
							"temperature": 0.1,
							"max_tokens":  20,
						},
					}
					
					_, err := leader.ProcessInference(context.Background(), req)
					if err != nil {
						b.Errorf("Inference failed: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkNetworkLatency benchmarks network latency between nodes
func BenchmarkNetworkLatency(b *testing.B) {
	cluster := setupBenchmarkCluster(b, 3)
	defer cluster.Shutdown()

	leader := cluster.GetLeader()
	require.NotNil(b, leader)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now()
			
			// Perform a distributed operation that requires network communication
			err := leader.ApplyConsensusOperation(
				fmt.Sprintf("network-test-%d", time.Now().UnixNano()),
				"network latency test value",
			)
			
			if err != nil {
				b.Errorf("Network operation failed: %v", err)
			}
			
			latency := time.Since(start)
			b.ReportMetric(float64(latency.Nanoseconds()), "ns/op")
		}
	})
}

// Helper function to setup benchmark cluster
func setupBenchmarkCluster(b *testing.B, nodeCount int) *integration.TestCluster {
	cluster, err := integration.NewTestCluster(nodeCount)
	require.NoError(b, err)

	err = cluster.Start()
	require.NoError(b, err)

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	return cluster
}

// Helper function to run load balancing test
func runLoadBalancingTest(b *testing.B, leader *integration.TestNode, requests, concurrency int) {
	requestChan := make(chan int, requests)
	responseChan := make(chan bool, requests)
	errorChan := make(chan error, requests)

	// Generate requests
	for i := 0; i < requests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Process requests concurrently
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for reqNum := range requestChan {
				req := &api.InferenceRequest{
					Model:  "llama3.2:1b",
					Prompt: fmt.Sprintf("Load balancing test request %d", reqNum),
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  20,
					},
				}

				_, err := leader.ProcessInference(context.Background(), req)
				if err != nil {
					errorChan <- err
				} else {
					responseChan <- true
				}
			}
		}()
	}

	// Wait for completion
	wg.Wait()

	// Count results
	successCount := 0
	errorCount := 0

	for len(responseChan) > 0 || len(errorChan) > 0 {
		select {
		case <-responseChan:
			successCount++
		case err := <-errorChan:
			b.Logf("Load balancing test error: %v", err)
			errorCount++
		default:
			break
		}
	}

	if successCount+errorCount != requests {
		b.Errorf("Expected %d total responses, got %d", requests, successCount+errorCount)
	}
}

// Performance test for specific scenarios
func TestPerformanceScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	cluster := setupBenchmarkCluster(t, 3)
	defer cluster.Shutdown()

	t.Run("HighThroughputScenario", func(t *testing.T) {
		testHighThroughputScenario(t, cluster)
	})

	t.Run("LowLatencyScenario", func(t *testing.T) {
		testLowLatencyScenario(t, cluster)
	})

	t.Run("ResourceConstrainedScenario", func(t *testing.T) {
		testResourceConstrainedScenario(t, cluster)
	})
}

// testHighThroughputScenario tests high throughput scenario
func testHighThroughputScenario(t *testing.T, cluster *integration.TestCluster) {
	leader := cluster.GetLeader()
	require.NotNil(t, leader)

	requests := 1000
	concurrency := 50
	
	startTime := time.Now()
	
	requestChan := make(chan int, requests)
	responseChan := make(chan bool, requests)
	errorChan := make(chan error, requests)

	// Generate requests
	for i := 0; i < requests; i++ {
		requestChan <- i
	}
	close(requestChan)

	// Process requests
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for reqNum := range requestChan {
				req := &api.InferenceRequest{
					Model:  "llama3.2:1b",
					Prompt: fmt.Sprintf("High throughput test %d", reqNum),
					Options: map[string]interface{}{
						"temperature": 0.1,
						"max_tokens":  20,
					},
				}

				_, err := leader.ProcessInference(context.Background(), req)
				if err != nil {
					errorChan <- err
				} else {
					responseChan <- true
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Count results
	successCount := 0
	errorCount := 0

	for len(responseChan) > 0 || len(errorChan) > 0 {
		select {
		case <-responseChan:
			successCount++
		case <-errorChan:
			errorCount++
		default:
			break
		}
	}

	throughput := float64(successCount) / duration.Seconds()
	
	t.Logf("High throughput test results:")
	t.Logf("  Total requests: %d", requests)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", errorCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Throughput: %.2f req/s", throughput)
	
	// Verify performance requirements
	require.Greater(t, successCount, requests*8/10, "Should have at least 80% success rate")
	require.Greater(t, throughput, 10.0, "Should achieve at least 10 req/s throughput")
}

// testLowLatencyScenario tests low latency scenario
func testLowLatencyScenario(t *testing.T, cluster *integration.TestCluster) {
	leader := cluster.GetLeader()
	require.NotNil(t, leader)

	requests := 100
	latencies := make([]time.Duration, requests)

	for i := 0; i < requests; i++ {
		req := &api.InferenceRequest{
			Model:  "llama3.2:1b",
			Prompt: fmt.Sprintf("Low latency test %d", i),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  10,
			},
		}

		start := time.Now()
		_, err := leader.ProcessInference(context.Background(), req)
		latency := time.Since(start)
		
		require.NoError(t, err)
		latencies[i] = latency
	}

	// Calculate statistics
	var totalLatency time.Duration
	minLatency := latencies[0]
	maxLatency := latencies[0]

	for _, latency := range latencies {
		totalLatency += latency
		if latency < minLatency {
			minLatency = latency
		}
		if latency > maxLatency {
			maxLatency = latency
		}
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	t.Logf("Low latency test results:")
	t.Logf("  Requests: %d", requests)
	t.Logf("  Min latency: %v", minLatency)
	t.Logf("  Max latency: %v", maxLatency)
	t.Logf("  Avg latency: %v", avgLatency)

	// Verify latency requirements
	require.Less(t, avgLatency, 5*time.Second, "Average latency should be less than 5 seconds")
	require.Less(t, maxLatency, 15*time.Second, "Max latency should be less than 15 seconds")
}

// testResourceConstrainedScenario tests resource constrained scenario
func testResourceConstrainedScenario(t *testing.T, cluster *integration.TestCluster) {
	leader := cluster.GetLeader()
	require.NotNil(t, leader)

	// Fill up the cluster with models
	modelCount := 20
	for i := 0; i < modelCount; i++ {
		modelInfo := &integration.ModelInfo{
			Name:              fmt.Sprintf("resource-test-model-%d", i),
			Size:              100 * 1024 * 1024, // 100MB each
			Checksum:          fmt.Sprintf("checksum-%d", i),
			ReplicationFactor: 2,
			LastAccessed:      time.Now(),
			Popularity:        rand.Float64(),
		}

		err := leader.RegisterModel(modelInfo)
		require.NoError(t, err)
	}

	// Wait for models to be distributed
	time.Sleep(30 * time.Second)

	// Now test inference under resource constraints
	requests := 50
	successCount := 0
	errorCount := 0

	for i := 0; i < requests; i++ {
		req := &api.InferenceRequest{
			Model:  fmt.Sprintf("resource-test-model-%d", i%modelCount),
			Prompt: fmt.Sprintf("Resource constrained test %d", i),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  20,
			},
		}

		_, err := leader.ProcessInference(context.Background(), req)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	t.Logf("Resource constrained test results:")
	t.Logf("  Total requests: %d", requests)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Failed: %d", errorCount)
	t.Logf("  Success rate: %.2f%%", float64(successCount)/float64(requests)*100)

	// Verify that system handles resource constraints gracefully
	require.Greater(t, successCount, requests/2, "Should handle at least 50% of requests under resource constraints")
}