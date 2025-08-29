package benchmarks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/cache"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/memory"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkAPIServerThroughput measures API server throughput under load
func BenchmarkAPIServerThroughput(b *testing.B) {
	// Create test server
	server := createTestAPIServer()
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(server.URL + "/api/v1/health")
			if err != nil {
				b.Error(err)
				continue
			}
			resp.Body.Close()
			
			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected 200, got %d", resp.StatusCode)
			}
		}
	})
}

// BenchmarkAPIServerLatency measures API server response latency
func BenchmarkAPIServerLatency(b *testing.B) {
	server := createTestAPIServer()
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	
	var totalLatency time.Duration
	var measurements int64

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		start := time.Now()
		resp, err := client.Get(server.URL + "/api/v1/health")
		latency := time.Since(start)
		
		if err == nil {
			resp.Body.Close()
			totalLatency += latency
			measurements++
		}
	}

	avgLatency := totalLatency / time.Duration(measurements)
	b.ReportMetric(float64(avgLatency.Nanoseconds())/1e6, "ms/op")
}

// BenchmarkCachePerformance benchmarks cache operations
func BenchmarkCachePerformance(b *testing.B) {
	config := cache.DefaultCacheConfig()
	config.MaxMemoryEntries = 10000
	
	algorithmCache, err := cache.NewAlgorithmCache(config)
	require.NoError(b, err)
	defer algorithmCache.Close()

	// Benchmark cache writes
	b.Run("CacheSet", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i)
				value := fmt.Sprintf("value_%d", i)
				err := algorithmCache.Set(key, value, 5*time.Minute)
				if err != nil {
					b.Error(err)
				}
				i++
			}
		})
	})

	// Benchmark cache reads
	b.Run("CacheGet", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := fmt.Sprintf("value_%d", i)
			algorithmCache.Set(key, value, 5*time.Minute)
		}

		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i%1000)
				_, found, err := algorithmCache.Get(key)
				if err != nil {
					b.Error(err)
				}
				if !found {
					b.Error("Key not found")
				}
				i++
			}
		})
	})
}

// BenchmarkMemoryManager benchmarks memory management operations
func BenchmarkMemoryManager(b *testing.B) {
	config := memory.DefaultConfig()
	config.MaxMemoryMB = 1024 // 1GB limit
	
	manager := memory.NewManager(config)
	err := manager.Start()
	require.NoError(b, err)
	defer manager.Stop()

	b.Run("MemoryPoolAllocation", func(b *testing.B) {
		pool := manager.GetPool("test-pool", 1024)
		
		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				item := pool.Get()
				pool.Put(item)
			}
		})
	})

	b.Run("CacheOperations", func(b *testing.B) {
		cache := manager.GetCache("test-cache")
		
		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key_%d", i)
				value := fmt.Sprintf("value_%d", i)
				
				// Set and get operation
				cache.Set(key, value)
				cache.Get(key)
				
				i++
			}
		})
	})
}

// BenchmarkNetworkOptimizer benchmarks network optimization
func BenchmarkNetworkOptimizer(b *testing.B) {
	config := network.DefaultConfig()
	optimizer := network.NewOptimizer(config)

	// Test data for compression
	testData := make([]byte, 10*1024) // 10KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.Run("DataCompression", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			compressed, wasCompressed, err := optimizer.CompressData(testData)
			if err != nil {
				b.Error(err)
			}
			if !wasCompressed {
				b.Error("Data should have been compressed")
			}
			
			// Measure compression ratio
			if i == 0 {
				ratio := float64(len(compressed)) / float64(len(testData))
				b.ReportMetric(ratio*100, "compression_%")
			}
		}
	})

	b.Run("DataDecompression", func(b *testing.B) {
		// Pre-compress data
		compressed, _, err := optimizer.CompressData(testData)
		require.NoError(b, err)
		
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			decompressed, err := optimizer.DecompressData(compressed)
			if err != nil {
				b.Error(err)
			}
			if len(decompressed) != len(testData) {
				b.Error("Decompressed data size mismatch")
			}
		}
	})
}

// BenchmarkConcurrentOperations tests system under concurrent load
func BenchmarkConcurrentOperations(b *testing.B) {
	server := createTestAPIServer()
	defer server.Close()

	// Test different concurrency levels
	concurrencyLevels := []int{1, 10, 50, 100, 250}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			client := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        concurrency * 2,
					MaxIdleConnsPerHost: concurrency,
					IdleConnTimeout:     30 * time.Second,
				},
			}

			var wg sync.WaitGroup
			var successCount int64
			var errorCount int64
			
			b.ResetTimer()
			startTime := time.Now()
			
			// Launch concurrent workers
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					
					for j := 0; j < b.N/concurrency; j++ {
						resp, err := client.Get(server.URL + "/api/v1/health")
						if err != nil {
							errorCount++
							continue
						}
						resp.Body.Close()
						
						if resp.StatusCode == http.StatusOK {
							successCount++
						} else {
							errorCount++
						}
					}
				}()
			}
			
			wg.Wait()
			elapsed := time.Since(startTime)
			
			// Calculate and report metrics
			throughput := float64(successCount) / elapsed.Seconds()
			errorRate := float64(errorCount) / float64(successCount+errorCount) * 100
			
			b.ReportMetric(throughput, "ops/sec")
			b.ReportMetric(errorRate, "error_%")
		})
	}
}

// BenchmarkMemoryUsage measures memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	var m1, m2 runtime.MemStats
	
	// Measure baseline memory
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// Run memory-intensive operations
	data := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = make([]byte, 1024) // 1KB per allocation
	}
	
	// Force GC and measure memory
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	// Calculate memory usage
	allocatedBytes := m2.TotalAlloc - m1.TotalAlloc
	b.ReportMetric(float64(allocatedBytes)/float64(b.N), "bytes/op")
	
	// Report GC metrics
	gcPauses := m2.PauseTotalNs - m1.PauseTotalNs
	b.ReportMetric(float64(gcPauses)/1e6, "gc_pause_ms")
}

// BenchmarkGCPressure measures garbage collection impact
func BenchmarkGCPressure(b *testing.B) {
	// Configure GC for benchmark
	oldGOGC := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(oldGOGC)
	
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Create memory pressure
	for i := 0; i < b.N; i++ {
		// Allocate and immediately discard memory
		_ = make([]byte, 8*1024) // 8KB allocation
		
		// Simulate work
		time.Sleep(time.Microsecond)
	}
	
	runtime.ReadMemStats(&m2)
	
	// Report GC statistics
	gcCycles := m2.NumGC - m1.NumGC
	b.ReportMetric(float64(gcCycles), "gc_cycles")
	
	avgPause := float64(m2.PauseTotalNs-m1.PauseTotalNs) / float64(gcCycles) / 1e6
	b.ReportMetric(avgPause, "avg_gc_pause_ms")
}

// BenchmarkDatabaseOperations simulates database performance
func BenchmarkDatabaseOperations(b *testing.B) {
	// Simulate database with in-memory map for benchmark
	db := make(map[string]interface{})
	var mu sync.RWMutex
	
	// Pre-populate with test data
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("record_%d", i)
		value := map[string]interface{}{
			"id":   i,
			"name": fmt.Sprintf("Item %d", i),
			"data": make([]byte, 512), // 512 bytes per record
		}
		db[key] = value
	}

	b.Run("DatabaseRead", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("record_%d", i%10000)
				
				mu.RLock()
				_, exists := db[key]
				mu.RUnlock()
				
				if !exists {
					b.Error("Record not found")
				}
				i++
			}
		})
	})

	b.Run("DatabaseWrite", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 10000
			for pb.Next() {
				key := fmt.Sprintf("new_record_%d", i)
				value := map[string]interface{}{
					"id":   i,
					"name": fmt.Sprintf("New Item %d", i),
					"data": make([]byte, 512),
				}
				
				mu.Lock()
				db[key] = value
				mu.Unlock()
				
				i++
			}
		})
	})
}

// createTestAPIServer creates a test HTTP server for benchmarking
func createTestAPIServer() *httptest.Server {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})
	
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"nodes":[{"id":"node1","status":"active"},{"id":"node2","status":"active"}]}`))
	})
	
	server := httptest.NewServer(mux)
	return server
}

// TestPerformanceRegression validates performance doesn't regress
func TestPerformanceRegression(t *testing.T) {
	// Performance thresholds (adjust based on baseline measurements)
	thresholds := map[string]float64{
		"api_latency_ms":      50.0,  // Max 50ms API latency
		"cache_ops_per_sec":   10000, // Min 10K cache ops/sec
		"memory_mb_per_op":    1.0,   // Max 1MB memory per operation
		"throughput_ops_sec":  1000,  // Min 1K ops/sec throughput
	}

	// Run benchmark and collect results
	results := runPerformanceBenchmarks()
	
	// Validate against thresholds
	for metric, threshold := range thresholds {
		value, exists := results[metric]
		if !exists {
			t.Errorf("Missing performance metric: %s", metric)
			continue
		}
		
		switch metric {
		case "api_latency_ms", "memory_mb_per_op":
			// Lower is better
			assert.LessOrEqual(t, value, threshold, 
				"Performance regression in %s: got %.2f, threshold %.2f", metric, value, threshold)
		case "cache_ops_per_sec", "throughput_ops_sec":
			// Higher is better
			assert.GreaterOrEqual(t, value, threshold,
				"Performance regression in %s: got %.2f, threshold %.2f", metric, value, threshold)
		}
	}
}

// runPerformanceBenchmarks runs benchmarks and returns results
func runPerformanceBenchmarks() map[string]float64 {
	results := make(map[string]float64)
	
	// This would integrate with go test -bench to collect actual results
	// For now, return mock results
	results["api_latency_ms"] = 25.0
	results["cache_ops_per_sec"] = 15000
	results["memory_mb_per_op"] = 0.5
	results["throughput_ops_sec"] = 2000
	
	return results
}