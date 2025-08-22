package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// BenchmarkSuite provides comprehensive performance benchmarking
type BenchmarkSuite struct {
	results map[string]*BenchmarkResult
	mutex   sync.RWMutex
}

// BenchmarkResult contains the results of a performance benchmark
type BenchmarkResult struct {
	Name            string        `json:"name"`
	Duration        time.Duration `json:"duration"`
	Operations      int64         `json:"operations"`
	OperationsPerSec float64       `json:"operations_per_sec"`
	MemoryUsed      int64         `json:"memory_used_bytes"`
	GCPauses        int64         `json:"gc_pauses"`
	Goroutines      int           `json:"goroutines"`
	Timestamp       time.Time     `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		results: make(map[string]*BenchmarkResult),
	}
}

// RunBenchmark executes a benchmark and records the results
func (bs *BenchmarkSuite) RunBenchmark(name string, operation func() error, iterations int) *BenchmarkResult {
	// Record initial state
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)
	
	startTime := time.Now()
	goroutinesBefore := runtime.NumGoroutine()
	
	// Run the benchmark
	var successful int64
	var wg sync.WaitGroup
	
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := operation(); err == nil {
				atomic.AddInt64(&successful, 1)
			}
		}()
	}
	
	wg.Wait()
	
	// Record final state
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	// Calculate metrics
	opsPerSec := float64(successful) / duration.Seconds()
	memoryUsed := int64(memAfter.Alloc - memBefore.Alloc)
	gcPauses := int64(memAfter.NumGC - memBefore.NumGC)
	
	result := &BenchmarkResult{
		Name:            name,
		Duration:        duration,
		Operations:      successful,
		OperationsPerSec: opsPerSec,
		MemoryUsed:      memoryUsed,
		GCPauses:        gcPauses,
		Goroutines:      goroutinesAfter - goroutinesBefore,
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"iterations":         iterations,
			"success_rate":       float64(successful) / float64(iterations),
			"memory_alloc_mb":    float64(memoryUsed) / 1024 / 1024,
			"avg_operation_time": duration.Nanoseconds() / int64(successful),
		},
	}
	
	// Store result
	bs.mutex.Lock()
	bs.results[name] = result
	bs.mutex.Unlock()
	
	return result
}

// RunConcurrentBenchmark executes a benchmark with controlled concurrency
func (bs *BenchmarkSuite) RunConcurrentBenchmark(name string, operation func(ctx context.Context) error, concurrency int, duration time.Duration) *BenchmarkResult {
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)
	
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	startTime := time.Now()
	goroutinesBefore := runtime.NumGoroutine()
	
	var operations int64
	var wg sync.WaitGroup
	
	// Launch concurrent workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := operation(ctx); err == nil {
						atomic.AddInt64(&operations, 1)
					}
				}
			}
		}()
	}
	
	// Wait for completion
	<-ctx.Done()
	wg.Wait()
	
	endTime := time.Now()
	actualDuration := endTime.Sub(startTime)
	runtime.ReadMemStats(&memAfter)
	goroutinesAfter := runtime.NumGoroutine()
	
	opsPerSec := float64(operations) / actualDuration.Seconds()
	memoryUsed := int64(memAfter.Alloc - memBefore.Alloc)
	gcPauses := int64(memAfter.NumGC - memBefore.NumGC)
	
	result := &BenchmarkResult{
		Name:            name,
		Duration:        actualDuration,
		Operations:      operations,
		OperationsPerSec: opsPerSec,
		MemoryUsed:      memoryUsed,
		GCPauses:        gcPauses,
		Goroutines:      goroutinesAfter - goroutinesBefore,
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"concurrency":      concurrency,
			"target_duration":  duration,
			"actual_duration":  actualDuration,
			"memory_alloc_mb":  float64(memoryUsed) / 1024 / 1024,
			"ops_per_worker":   float64(operations) / float64(concurrency),
		},
	}
	
	bs.mutex.Lock()
	bs.results[name] = result
	bs.mutex.Unlock()
	
	return result
}

// GetResults returns all benchmark results
func (bs *BenchmarkSuite) GetResults() map[string]*BenchmarkResult {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	results := make(map[string]*BenchmarkResult)
	for k, v := range bs.results {
		results[k] = v
	}
	return results
}

// GetResult returns a specific benchmark result
func (bs *BenchmarkSuite) GetResult(name string) (*BenchmarkResult, bool) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	result, exists := bs.results[name]
	return result, exists
}

// ClearResults clears all benchmark results
func (bs *BenchmarkSuite) ClearResults() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()
	
	bs.results = make(map[string]*BenchmarkResult)
}

// PrintSummary prints a summary of all benchmark results
func (bs *BenchmarkSuite) PrintSummary() {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	fmt.Println("=== Performance Benchmark Summary ===")
	fmt.Printf("Timestamp: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("Active Goroutines: %d\n", runtime.NumGoroutine())
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory Usage: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Println()
	
	for name, result := range bs.results {
		fmt.Printf("Benchmark: %s\n", name)
		fmt.Printf("  Duration: %v\n", result.Duration)
		fmt.Printf("  Operations: %d\n", result.Operations)
		fmt.Printf("  Ops/sec: %.2f\n", result.OperationsPerSec)
		fmt.Printf("  Memory Used: %.2f MB\n", float64(result.MemoryUsed)/1024/1024)
		fmt.Printf("  GC Pauses: %d\n", result.GCPauses)
		fmt.Printf("  Goroutines Delta: %d\n", result.Goroutines)
		if successRate, ok := result.Metadata["success_rate"]; ok {
			fmt.Printf("  Success Rate: %.1f%%\n", successRate.(float64)*100)
		}
		fmt.Println()
	}
}

// CompareResults compares two benchmark results and returns improvement metrics
func (bs *BenchmarkSuite) CompareResults(baseline, current string) map[string]float64 {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()
	
	baseResult, baseExists := bs.results[baseline]
	currResult, currExists := bs.results[current]
	
	if !baseExists || !currExists {
		return nil
	}
	
	comparison := make(map[string]float64)
	
	// Operations per second improvement
	if baseResult.OperationsPerSec > 0 {
		comparison["ops_improvement"] = (currResult.OperationsPerSec - baseResult.OperationsPerSec) / baseResult.OperationsPerSec * 100
	}
	
	// Duration improvement (negative means faster)
	if baseResult.Duration > 0 {
		comparison["duration_improvement"] = -(float64(currResult.Duration-baseResult.Duration) / float64(baseResult.Duration) * 100)
	}
	
	// Memory usage change
	if baseResult.MemoryUsed != 0 {
		comparison["memory_change"] = float64(currResult.MemoryUsed-baseResult.MemoryUsed) / float64(baseResult.MemoryUsed) * 100
	}
	
	return comparison
}

// MemoryProfiler provides memory usage profiling
type MemoryProfiler struct {
	snapshots []MemorySnapshot
	mutex     sync.RWMutex
}

// MemorySnapshot represents a point-in-time memory snapshot
type MemorySnapshot struct {
	Timestamp     time.Time `json:"timestamp"`
	Alloc         uint64    `json:"alloc_bytes"`
	TotalAlloc    uint64    `json:"total_alloc_bytes"`
	Sys           uint64    `json:"sys_bytes"`
	NumGC         uint32    `json:"num_gc"`
	HeapObjects   uint64    `json:"heap_objects"`
	StackInUse    uint64    `json:"stack_in_use_bytes"`
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler() *MemoryProfiler {
	return &MemoryProfiler{}
}

// TakeSnapshot takes a memory snapshot
func (mp *MemoryProfiler) TakeSnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	snapshot := MemorySnapshot{
		Timestamp:   time.Now(),
		Alloc:       m.Alloc,
		TotalAlloc:  m.TotalAlloc,
		Sys:         m.Sys,
		NumGC:       m.NumGC,
		HeapObjects: m.HeapObjects,
		StackInUse:  m.StackInuse,
	}
	
	mp.mutex.Lock()
	mp.snapshots = append(mp.snapshots, snapshot)
	mp.mutex.Unlock()
	
	return snapshot
}

// GetSnapshots returns all memory snapshots
func (mp *MemoryProfiler) GetSnapshots() []MemorySnapshot {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	
	snapshots := make([]MemorySnapshot, len(mp.snapshots))
	copy(snapshots, mp.snapshots)
	return snapshots
}

// ClearSnapshots clears all memory snapshots
func (mp *MemoryProfiler) ClearSnapshots() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()
	
	mp.snapshots = mp.snapshots[:0]
}

// GetMemoryTrend analyzes memory usage trends
func (mp *MemoryProfiler) GetMemoryTrend() map[string]interface{} {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	
	if len(mp.snapshots) < 2 {
		return map[string]interface{}{"error": "insufficient data"}
	}
	
	first := mp.snapshots[0]
	last := mp.snapshots[len(mp.snapshots)-1]
	duration := last.Timestamp.Sub(first.Timestamp)
	
	allocGrowth := float64(last.Alloc-first.Alloc) / float64(first.Alloc) * 100
	sysGrowth := float64(last.Sys-first.Sys) / float64(first.Sys) * 100
	gcCount := last.NumGC - first.NumGC
	
	return map[string]interface{}{
		"duration_minutes":    duration.Minutes(),
		"alloc_growth_pct":    allocGrowth,
		"sys_growth_pct":      sysGrowth,
		"gc_count":            gcCount,
		"gc_frequency_per_min": float64(gcCount) / duration.Minutes(),
		"current_alloc_mb":    float64(last.Alloc) / 1024 / 1024,
		"current_sys_mb":      float64(last.Sys) / 1024 / 1024,
		"heap_objects":        last.HeapObjects,
		"stack_usage_mb":      float64(last.StackInUse) / 1024 / 1024,
	}
}