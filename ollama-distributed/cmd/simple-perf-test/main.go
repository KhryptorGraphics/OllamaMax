package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("🚀 OllamaMax Simple Performance Test")
	fmt.Println("====================================")

	runMemoryBenchmark()
	runConcurrencyBenchmark()
	runHTTPBenchmark()
	runGCBenchmark()
	
	fmt.Println("\n✅ Performance tests completed!")
	printPerformanceSummary()
}

func runMemoryBenchmark() {
	fmt.Println("\n🧠 Memory Performance Benchmark")
	fmt.Println("--------------------------------")

	var m1, m2 runtime.MemStats
	
	// Measure baseline
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	start := time.Now()
	
	// Simulate memory-intensive workload
	const iterations = 50000
	data := make([][]byte, iterations)
	
	for i := 0; i < iterations; i++ {
		data[i] = make([]byte, 2048) // 2KB allocations
		if i%1000 == 0 {
			// Fill with some data to prevent optimization
			for j := range data[i] {
				data[i][j] = byte(j % 256)
			}
		}
	}
	
	// Force GC and measure
	runtime.GC()
	runtime.ReadMemStats(&m2)
	elapsed := time.Since(start)
	
	// Calculate metrics
	allocatedMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
	memoryMB := float64(m2.Alloc) / 1024 / 1024
	gcCycles := m2.NumGC - m1.NumGC
	gcPauseMS := float64(m2.PauseTotalNs-m1.PauseTotalNs) / 1e6
	
	fmt.Printf("  Iterations:      %d\n", iterations)
	fmt.Printf("  Total time:      %v\n", elapsed)
	fmt.Printf("  Allocated:       %.2f MB\n", allocatedMB)
	fmt.Printf("  Current memory:  %.2f MB\n", memoryMB)
	fmt.Printf("  GC cycles:       %d\n", gcCycles)
	fmt.Printf("  GC pause total:  %.2f ms\n", gcPauseMS)
	fmt.Printf("  Allocation rate: %.2f MB/s\n", allocatedMB/elapsed.Seconds())
	
	if gcCycles > 0 {
		fmt.Printf("  Avg GC pause:    %.2f ms\n", gcPauseMS/float64(gcCycles))
	}
	
	// Cleanup
	data = nil
	runtime.GC()
}

func runConcurrencyBenchmark() {
	fmt.Println("\n⚡ Concurrency Performance Benchmark")
	fmt.Println("-------------------------------------")

	concurrencyLevels := []int{1, 10, 50, 100, 250, 500}
	
	for _, concurrency := range concurrencyLevels {
		var operations int64
		var wg sync.WaitGroup
		
		start := time.Now()
		
		// Launch workers
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				// Each worker performs 1000 operations
				for j := 0; j < 1000; j++ {
					// Simulate CPU-bound work
					_ = fmt.Sprintf("operation_%d_%d", i, j)
					
					// Simulate some computation
					sum := 0
					for k := 0; k < 100; k++ {
						sum += k * k
					}
					_ = sum
					
					atomic.AddInt64(&operations, 1)
				}
			}()
		}
		
		wg.Wait()
		elapsed := time.Since(start)
		
		opsPerSecond := float64(operations) / elapsed.Seconds()
		avgLatency := elapsed / time.Duration(operations)
		
		fmt.Printf("  Concurrency %3d: %8.0f ops/sec, %6s avg latency, %8s total\n", 
			concurrency, opsPerSecond, avgLatency, elapsed)
	}
}

func runHTTPBenchmark() {
	fmt.Println("\n🌐 HTTP Performance Benchmark")
	fmt.Println("------------------------------")

	// Create optimized client for external testing
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	// Use httpbin.org for reliable testing
	baseURL := "http://httpbin.org"
	fmt.Printf("  Using test endpoint: %s\n", baseURL)
	
	// Test different load levels
	loadLevels := []struct {
		name        string
		concurrency int
		requests    int
	}{
		{"Light", 5, 50},
		{"Medium", 10, 100},
		{"Heavy", 20, 200},
	}
	
	for _, load := range loadLevels {
		var success, failed int64
		var totalLatency int64
		var wg sync.WaitGroup
		
		start := time.Now()
		
		// Launch concurrent requests
		for i := 0; i < load.concurrency; i++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				
				requestsPerWorker := load.requests / load.concurrency
				for j := 0; j < requestsPerWorker; j++ {
					reqStart := time.Now()
					resp, err := client.Get(baseURL + "/status/200")
					reqLatency := time.Since(reqStart)
					
					if err != nil {
						atomic.AddInt64(&failed, 1)
						continue
					}
					
					resp.Body.Close()
					if resp.StatusCode == 200 {
						atomic.AddInt64(&success, 1)
						atomic.AddInt64(&totalLatency, int64(reqLatency))
					} else {
						atomic.AddInt64(&failed, 1)
					}
				}
			}(i)
		}
		
		wg.Wait()
		elapsed := time.Since(start)
		
		// Calculate metrics
		totalRequests := success + failed
		if totalRequests > 0 {
			successRate := float64(success) / float64(totalRequests) * 100
			throughput := float64(success) / elapsed.Seconds()
			avgLatency := time.Duration(totalLatency / success)
			
			fmt.Printf("  %s Load:  %3d/%3d success (%.1f%%), %6.0f req/s, %6s avg latency\n",
				load.name, success, totalRequests, successRate, throughput, avgLatency)
		}
	}
}

func runGCBenchmark() {
	fmt.Println("\n🗑️ Garbage Collection Benchmark")
	fmt.Println("--------------------------------")

	// Test different allocation patterns
	patterns := []struct {
		name        string
		allocSize   int
		allocCount  int
		holdPercent float64
	}{
		{"Small objects", 128, 100000, 0.1},
		{"Medium objects", 8192, 10000, 0.2},
		{"Large objects", 65536, 1000, 0.5},
	}
	
	for _, pattern := range patterns {
		fmt.Printf("  Testing %s:\n", pattern.name)
		
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		start := time.Now()
		
		// Create allocations with retention pattern
		hold := make([][]byte, 0, int(float64(pattern.allocCount)*pattern.holdPercent))
		
		for i := 0; i < pattern.allocCount; i++ {
			data := make([]byte, pattern.allocSize)
			
			// Fill with data to prevent optimization
			for j := 0; j < len(data); j += 64 {
				if j+8 < len(data) {
					data[j] = byte(i % 256)
				}
			}
			
			// Hold onto some objects to create retention
			if i%int(1.0/pattern.holdPercent) == 0 {
				hold = append(hold, data)
			}
			
			// Trigger GC periodically
			if i%1000 == 0 {
				runtime.GC()
			}
		}
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		elapsed := time.Since(start)
		
		// Calculate GC metrics
		gcCycles := m2.NumGC - m1.NumGC
		gcPauseTotal := time.Duration(m2.PauseTotalNs - m1.PauseTotalNs)
		allocatedMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
		retainedMB := float64(m2.Alloc) / 1024 / 1024
		
		fmt.Printf("    Time: %v, Allocated: %.1f MB, Retained: %.1f MB\n", 
			elapsed, allocatedMB, retainedMB)
		fmt.Printf("    GC cycles: %d, Total pause: %v", gcCycles, gcPauseTotal)
		
		if gcCycles > 0 {
			fmt.Printf(", Avg pause: %v\n", gcPauseTotal/time.Duration(gcCycles))
		} else {
			fmt.Printf("\n")
		}
		
		// Cleanup
		hold = nil
		runtime.GC()
	}
}

func printPerformanceSummary() {
	fmt.Println("\n📊 Performance Summary")
	fmt.Println("======================")
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	fmt.Printf("  Go version:      %s\n", runtime.Version())
	fmt.Printf("  GOMAXPROCS:      %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("  Goroutines:      %d\n", runtime.NumGoroutine())
	fmt.Printf("  Current memory:  %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("  Total allocated: %.2f MB\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Printf("  System memory:   %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("  GC cycles:       %d\n", m.NumGC)
	fmt.Printf("  Last GC:         %v ago\n", time.Since(time.Unix(0, int64(m.LastGC))))
	
	// Performance insights
	fmt.Println("\n💡 Performance Insights:")
	
	if m.Alloc > 100*1024*1024 { // > 100MB
		fmt.Println("  - High memory usage detected - consider optimization")
	} else {
		fmt.Println("  - Memory usage looks good")
	}
	
	if runtime.NumGoroutine() > 100 {
		fmt.Println("  - High goroutine count - check for leaks")
	} else {
		fmt.Println("  - Goroutine count looks healthy")
	}
	
	if m.NumGC > 1000 {
		fmt.Println("  - High GC frequency - consider tuning GC settings")
	} else {
		fmt.Println("  - GC frequency appears normal")
	}
	
	fmt.Println("\n🎯 Optimization Recommendations:")
	fmt.Println("  - Target < 10ms P99 latency for optimal performance")
	fmt.Println("  - Aim for 500+ ops/sec throughput under load")
	fmt.Println("  - Keep memory usage stable and avoid leaks")
	fmt.Println("  - Monitor GC pause times (target < 5ms)")
	fmt.Println("  - Use connection pooling for network operations")
}

// Benchmark helper functions
func benchmarkFunction(name string, iterations int, fn func()) {
	fmt.Printf("  %s: ", name)
	
	start := time.Now()
	for i := 0; i < iterations; i++ {
		fn()
	}
	elapsed := time.Since(start)
	
	nsPerOp := elapsed.Nanoseconds() / int64(iterations)
	opsPerSec := float64(iterations) / elapsed.Seconds()
	
	fmt.Printf("%d iterations, %d ns/op, %.0f ops/sec\n", iterations, nsPerOp, opsPerSec)
}