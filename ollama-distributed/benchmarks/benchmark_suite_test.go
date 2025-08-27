package benchmarks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestLogger implements Logger interface for testing
type TestLogger struct{}

func (l *TestLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *TestLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *TestLogger) Debug(msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

// BenchmarkComprehensivePerformance runs the full benchmark suite
func BenchmarkComprehensivePerformance(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping comprehensive benchmark in short mode")
	}

	logger := &TestLogger{}
	config := &BenchmarkConfig{
		Duration:          2 * time.Minute,
		WarmupDuration:    10 * time.Second,
		ConcurrentWorkers: 4,
		ClusterSizes:      []int{1, 3},
		ModelSizes:        []int{100, 500},
		RequestSizes:      []int{64, 1024},
		Categories: []string{
			"consensus",
			"p2p_networking",
			"api_endpoints",
			"memory_usage",
		},
		OutputDir:    "./benchmark-results",
		ReportFormat: "yaml",
	}

	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		err := runner.Run(ctx)
		if err != nil {
			b.Fatalf("Benchmark run failed: %v", err)
		}
	}
}

// BenchmarkConsensusOperations benchmarks consensus performance
func BenchmarkConsensusOperations(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	config.Duration = 30 * time.Second
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	b.Run("SingleThreaded", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkConsensusOperations(ctx, 1)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MultiThreaded", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkConsensusOperations(ctx, 4)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("HighConcurrency", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkConsensusOperations(ctx, 50)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkP2PNetworking benchmarks P2P networking performance
func BenchmarkP2PNetworking(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	b.Run("PeerDiscovery", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkPeerDiscovery(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MessageBroadcast", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkMessageBroadcast(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ContentRouting", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkContentRouting(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkModelDistribution benchmarks model distribution performance
func BenchmarkModelDistribution(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	modelSizes := []int{100, 500, 1000}

	for _, size := range modelSizes {
		b.Run(fmt.Sprintf("Download_%dMB", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runner.benchmarkModelDownload(ctx, size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("Replication_%dMB", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runner.benchmarkModelReplication(ctx, size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkAPIEndpoints benchmarks API endpoint performance
func BenchmarkAPIEndpoints(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	endpoints := []struct {
		name   string
		method string
		path   string
	}{
		{"HealthCheck", "GET", "/health"},
		{"ClusterStatus", "GET", "/api/v1/cluster/status"},
		{"ListNodes", "GET", "/api/v1/nodes"},
		{"ListModels", "GET", "/api/v1/models"},
	}

	for _, endpoint := range endpoints {
		b.Run(endpoint.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runner.benchmarkAPIEndpoint(ctx, endpoint.method, endpoint.path)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	b.Run("MemoryEfficiency", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkMemoryEfficiency(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GarbageCollection", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkGarbageCollection(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MemoryLeakDetection", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkMemoryLeaks(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent operation performance
func BenchmarkConcurrentOperations(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	concurrencyLevels := []int{1, 10, 50, 100}

	for _, level := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", level), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runner.benchmarkConcurrentOperations(ctx, level)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLoadBalancing benchmarks load balancing performance
func BenchmarkLoadBalancing(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	loadPatterns := []struct {
		name     string
		requests int
		workers  int
	}{
		{"LightLoad", 100, 5},
		{"MediumLoad", 500, 20},
		{"HeavyLoad", 1000, 50},
	}

	for _, pattern := range loadPatterns {
		b.Run(pattern.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := runner.benchmarkLoadBalancing(ctx, pattern.requests, pattern.workers)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkScalability tests scalability across different cluster sizes
func BenchmarkScalability(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping scalability benchmark in short mode")
	}

	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	config.Duration = 1 * time.Minute
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	clusterSizes := []int{1, 3, 5}

	for _, size := range clusterSizes {
		b.Run(fmt.Sprintf("Cluster_%d_Nodes", size), func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				// Measure system metrics for this cluster size
				metrics, err := runner.measureSystemMetrics(ctx, size)
				if err != nil {
					b.Fatal(err)
				}
				
				// Report key metrics
				b.ReportMetric(metrics.RequestsPerSecond, "requests/sec")
				b.ReportMetric(metrics.LatencyMean, "latency_ms")
				b.ReportMetric(metrics.CPUUsagePercent, "cpu_percent")
				b.ReportMetric(metrics.MemoryUsageMB, "memory_mb")
			}
		})
	}
}

// BenchmarkResourceUtilization measures resource usage patterns
func BenchmarkResourceUtilization(b *testing.B) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	b.Run("CPUIntensive", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate CPU-intensive operations
			err := runner.benchmarkConsensusOperations(ctx, 100)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MemoryIntensive", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkMemoryEfficiency(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NetworkIntensive", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := runner.benchmarkMessageBroadcast(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestBenchmarkRunner tests the benchmark runner functionality
func TestBenchmarkRunner(t *testing.T) {
	logger := &TestLogger{}
	config := &BenchmarkConfig{
		Duration:          5 * time.Second,
		WarmupDuration:    1 * time.Second,
		ConcurrentWorkers: 2,
		ClusterSizes:      []int{1},
		Categories:        []string{"consensus"},
		OutputDir:         "./test-results",
		ReportFormat:      "yaml",
	}

	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	// Test baseline establishment
	err := runner.EstablishBaseline(ctx)
	if err != nil {
		t.Fatalf("Failed to establish baseline: %v", err)
	}

	// Verify baseline was created
	if runner.baseline == nil {
		t.Fatal("Baseline was not established")
	}

	// Test category execution
	err = runner.runConsensusBenchmarks(ctx)
	if err != nil {
		t.Fatalf("Failed to run consensus benchmarks: %v", err)
	}

	// Verify results were created
	if len(runner.results.Categories) == 0 {
		t.Fatal("No category results were generated")
	}

	// Verify consensus category exists
	consensusResults, exists := runner.results.Categories["consensus"]
	if !exists {
		t.Fatal("Consensus category results not found")
	}

	if len(consensusResults.Tests) == 0 {
		t.Fatal("No test results were generated for consensus category")
	}

	// Test results saving
	err = runner.SaveResults()
	if err != nil {
		t.Fatalf("Failed to save results: %v", err)
	}

	// Verify results file was created
	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		t.Fatal("Output directory was not created")
	}

	// Cleanup
	os.RemoveAll(config.OutputDir)
}

// TestPerformanceRegression checks for performance regressions
func TestPerformanceRegression(t *testing.T) {
	logger := &TestLogger{}
	config := DefaultBenchmarkConfig()
	config.Duration = 10 * time.Second
	config.Categories = []string{"consensus"}

	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	// Establish baseline
	err := runner.EstablishBaseline(ctx)
	if err != nil {
		t.Fatalf("Failed to establish baseline: %v", err)
	}

	baseline := runner.baseline.SingleNode

	// Run current performance test
	current, err := runner.measureSystemMetrics(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to measure current performance: %v", err)
	}

	// Check for regressions (allowing for 10% variation)
	tolerance := 0.10

	if current.RequestsPerSecond < baseline.RequestsPerSecond*(1-tolerance) {
		t.Errorf("Throughput regression detected: current=%.2f, baseline=%.2f", 
			current.RequestsPerSecond, baseline.RequestsPerSecond)
	}

	if current.LatencyMean > baseline.LatencyMean*(1+tolerance) {
		t.Errorf("Latency regression detected: current=%.2f, baseline=%.2f", 
			current.LatencyMean, baseline.LatencyMean)
	}

	if current.MemoryUsageMB > baseline.MemoryUsageMB*(1+tolerance) {
		t.Errorf("Memory usage regression detected: current=%.2f, baseline=%.2f", 
			current.MemoryUsageMB, baseline.MemoryUsageMB)
	}

	t.Logf("Performance check passed - no regressions detected")
	t.Logf("Throughput: %.2f req/s (baseline: %.2f)", current.RequestsPerSecond, baseline.RequestsPerSecond)
	t.Logf("Latency: %.2f ms (baseline: %.2f)", current.LatencyMean, baseline.LatencyMean)
	t.Logf("Memory: %.2f MB (baseline: %.2f)", current.MemoryUsageMB, baseline.MemoryUsageMB)
}

// ExampleBenchmarkRunner demonstrates how to use the benchmark runner
func ExampleBenchmarkRunner() {
	// Create a simple logger
	logger := &TestLogger{}
	
	// Configure benchmark
	config := &BenchmarkConfig{
		Duration:          30 * time.Second,
		WarmupDuration:    5 * time.Second,
		ConcurrentWorkers: 4,
		ClusterSizes:      []int{1, 3},
		Categories:        []string{"consensus", "api_endpoints"},
		OutputDir:         "./benchmark-results",
		ReportFormat:      "yaml",
	}

	// Create and run benchmark
	runner := NewBenchmarkRunner(config, logger)
	ctx := context.Background()

	err := runner.Run(ctx)
	if err != nil {
		fmt.Printf("Benchmark failed: %v\n", err)
		return
	}

	// Results are automatically saved to the output directory
	fmt.Printf("Benchmark completed successfully\n")
	fmt.Printf("Results saved to: %s\n", config.OutputDir)
	
	// Output:
	// Benchmark completed successfully
	// Results saved to: ./benchmark-results
}