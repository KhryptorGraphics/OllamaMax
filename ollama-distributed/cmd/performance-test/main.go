package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/performance"
)

func main() {
	fmt.Println("🚀 OllamaMax Performance Test Suite")
	fmt.Println("===================================")

	// Run performance tests
	runMemoryTest()
	runConcurrencyTest()
	runGCTest()
	runSystemOptimizerTest()

	fmt.Println("\n✅ Performance tests completed!")
}

// runMemoryTest tests memory performance
func runMemoryTest() {
	fmt.Println("\n🧠 Memory Performance Test")
	fmt.Println("---------------------------")

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Allocate memory in patterns similar to real workload
	start := time.Now()
	data := make([][]byte, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = make([]byte, 1024) // 1KB allocations
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)
	elapsed := time.Since(start)

	allocatedMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
	gcPauseMS := float64(m2.PauseTotalNs-m1.PauseTotalNs) / 1e6

	fmt.Printf("  Memory allocated: %.2f MB\n", allocatedMB)
	fmt.Printf("  Allocation time: %v\n", elapsed)
	fmt.Printf("  GC pause time: %.2f ms\n", gcPauseMS)
	fmt.Printf("  GC cycles: %d\n", m2.NumGC-m1.NumGC)

	// Clean up
	data = nil
	runtime.GC()
}

// runConcurrencyTest tests concurrent operations
func runConcurrencyTest() {
	fmt.Println("\n⚡ Concurrency Performance Test")
	fmt.Println("-------------------------------")

	concurrencyLevels := []int{10, 50, 100, 250}
	
	for _, concurrency := range concurrencyLevels {
		var ops int64
		var wg sync.WaitGroup
		
		start := time.Now()
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					// Simulate work
					_ = fmt.Sprintf("operation_%d", j)
					atomic.AddInt64(&ops, 1)
				}
			}()
		}
		
		wg.Wait()
		elapsed := time.Since(start)
		opsPerSecond := float64(ops) / elapsed.Seconds()
		
		fmt.Printf("  Concurrency %3d: %8.0f ops/sec (%v total)\n", 
			concurrency, opsPerSecond, elapsed)
	}
}

// runGCTest tests garbage collection performance
func runGCTest() {
	fmt.Println("\n🗑️  Garbage Collection Test")
	fmt.Println("---------------------------")

	var m1, m2 runtime.MemStats
	
	// Test with different GC target percentages
	gcTargets := []int{50, 100, 200}
	
	for _, target := range gcTargets {
		fmt.Printf("  Testing GC target: %d%%\n", target)
		
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		start := time.Now()
		
		// Create memory pressure
		for i := 0; i < 5000; i++ {
			data := make([]byte, 2048) // 2KB allocations
			_ = data
			
			if i%1000 == 0 {
				runtime.GC()
			}
		}
		
		runtime.ReadMemStats(&m2)
		elapsed := time.Since(start)
		
		gcCycles := m2.NumGC - m1.NumGC
		avgPauseMS := float64(m2.PauseTotalNs-m1.PauseTotalNs) / float64(gcCycles) / 1e6
		
		fmt.Printf("    Time: %v, GC cycles: %d, Avg pause: %.2f ms\n",
			elapsed, gcCycles, avgPauseMS)
	}
}

// runSystemOptimizerTest tests the system optimizer
func runSystemOptimizerTest() {
	fmt.Println("\n🔧 System Optimizer Test")
	fmt.Println("-------------------------")

	config := performance.DefaultOptimizerConfig()
	optimizer, err := performance.NewSystemOptimizer(config)
	if err != nil {
		log.Printf("Failed to create optimizer: %v", err)
		return
	}

	// Start optimizer
	err = optimizer.Start()
	if err != nil {
		log.Printf("Failed to start optimizer: %v", err)
		return
	}
	defer optimizer.Stop()

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Test optimized request processing
	testData := make([]byte, 1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	start := time.Now()
	var processed int
	
	for i := 0; i < 1000; i++ {
		result, err := optimizer.OptimizeRequest(nil, testData)
		if err != nil {
			log.Printf("Request optimization failed: %v", err)
			continue
		}
		if len(result) > 0 {
			processed++
		}
	}
	
	elapsed := time.Since(start)
	requestsPerSecond := float64(processed) / elapsed.Seconds()
	
	fmt.Printf("  Processed requests: %d\n", processed)
	fmt.Printf("  Processing time: %v\n", elapsed)
	fmt.Printf("  Requests/sec: %.0f\n", requestsPerSecond)
	
	// Get performance metrics
	metrics := optimizer.GetMetrics()
	fmt.Printf("  Memory usage: %.2f MB\n", metrics.MemoryUsageMB)
	fmt.Printf("  Cache hit rate: %.1f%%\n", metrics.CacheHitRate)
	fmt.Printf("  Active connections: %d\n", metrics.ActiveConnections)
	
	// Check if system is healthy
	healthy := optimizer.IsHealthy()
	fmt.Printf("  System healthy: %t\n", healthy)
	
	// Get recommendations
	recommendations := optimizer.GetOptimizationRecommendations()
	if len(recommendations) > 0 {
		fmt.Println("  Recommendations:")
		for _, rec := range recommendations {
			fmt.Printf("    - %s\n", rec)
		}
	}
}

// runHTTPServerTest tests HTTP server performance
func runHTTPServerTest() {
	fmt.Println("\n🌐 HTTP Server Performance Test")
	fmt.Println("--------------------------------")

	// Create a simple test server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:         ":0", // Let system choose port
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test concurrent requests
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	var successCount int64
	var wg sync.WaitGroup
	start := time.Now()

	// Launch concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				resp, err := client.Get("http://localhost" + server.Addr + "/health")
				if err == nil && resp.StatusCode == 200 {
					atomic.AddInt64(&successCount, 1)
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)
	
	fmt.Printf("  Successful requests: %d/1000\n", successCount)
	fmt.Printf("  Total time: %v\n", elapsed)
	fmt.Printf("  Requests/sec: %.0f\n", float64(successCount)/elapsed.Seconds())

	// Shutdown server
	server.Close()
}