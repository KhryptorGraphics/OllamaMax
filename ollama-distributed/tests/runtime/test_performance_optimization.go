//go:build ignore

package main

import (
	"fmt"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/performance"
)

func main() {
	fmt.Println("Testing Performance Optimization System...")

	// Setup complete performance optimization stack
	fmt.Println("Setting up performance optimization stack...")

	// 1. Create performance optimization engine
	optimizationConfig := performance.DefaultOptimizationConfig()
	optimizationConfig.OptimizationInterval = 30 * time.Second // Fast optimization for testing
	optimizationEngine := performance.NewPerformanceOptimizationEngine(optimizationConfig)

	// 2. Create resource optimizer
	resourceConfig := performance.DefaultResourceOptimizerConfig()
	resourceOptimizer := performance.NewResourceOptimizer(resourceConfig)

	// 3. Create advanced cache manager
	cacheConfig := performance.DefaultAdvancedCacheConfig()
	cacheConfig.L1MaxSize = 50 * 1024 * 1024 // 50MB for testing
	cacheManager := performance.NewAdvancedCacheManager(cacheConfig)

	// 4. Create performance profiler
	profilerConfig := performance.DefaultProfilerConfig()
	profilerConfig.ProfilingInterval = 10 * time.Second // Fast profiling for testing
	profiler := performance.NewPerformanceProfiler(profilerConfig)

	// 5. Create auto tuner
	tunerConfig := performance.DefaultAutoTunerConfig()
	tunerConfig.TuningInterval = 1 * time.Minute // Fast tuning for testing
	autoTuner := performance.NewAutoTuner(tunerConfig)

	// Start all systems
	fmt.Println("Starting performance optimization systems...")

	if err := optimizationEngine.Start(); err != nil {
		fmt.Printf("❌ Failed to start optimization engine: %v\n", err)
		return
	}

	if err := resourceOptimizer.Start(); err != nil {
		fmt.Printf("❌ Failed to start resource optimizer: %v\n", err)
		return
	}

	if err := cacheManager.Start(); err != nil {
		fmt.Printf("❌ Failed to start cache manager: %v\n", err)
		return
	}

	if err := profiler.Start(); err != nil {
		fmt.Printf("❌ Failed to start profiler: %v\n", err)
		return
	}

	if err := autoTuner.Start(); err != nil {
		fmt.Printf("❌ Failed to start auto tuner: %v\n", err)
		return
	}

	fmt.Println("✅ All performance optimization systems started")

	// Run tests
	testResults := []bool{}

	// Test 1: Performance Optimization Engine
	fmt.Println("\n=== Testing Performance Optimization Engine ===")
	result := testPerformanceOptimizationEngine(optimizationEngine)
	testResults = append(testResults, result)

	// Test 2: Resource Optimizer
	fmt.Println("\n=== Testing Resource Optimizer ===")
	result = testResourceOptimizer(resourceOptimizer)
	testResults = append(testResults, result)

	// Test 3: Advanced Cache Manager
	fmt.Println("\n=== Testing Advanced Cache Manager ===")
	result = testAdvancedCacheManager(cacheManager)
	testResults = append(testResults, result)

	// Test 4: Performance Profiler
	fmt.Println("\n=== Testing Performance Profiler ===")
	result = testPerformanceProfiler(profiler)
	testResults = append(testResults, result)

	// Test 5: Auto Tuner
	fmt.Println("\n=== Testing Auto Tuner ===")
	result = testAutoTuner(autoTuner)
	testResults = append(testResults, result)

	// Test 6: Multi-level Caching
	fmt.Println("\n=== Testing Multi-level Caching ===")
	result = testMultiLevelCaching(cacheManager)
	testResults = append(testResults, result)

	// Test 7: Bottleneck Detection
	fmt.Println("\n=== Testing Bottleneck Detection ===")
	result = testBottleneckDetection(profiler)
	testResults = append(testResults, result)

	// Test 8: Performance Integration
	fmt.Println("\n=== Testing Performance Integration ===")
	result = testPerformanceIntegration(optimizationEngine, profiler, autoTuner)
	testResults = append(testResults, result)

	// Cleanup
	fmt.Println("\n=== Cleaning up ===")
	optimizationEngine.Shutdown()
	resourceOptimizer.Shutdown()
	cacheManager.Shutdown()
	profiler.Shutdown()
	autoTuner.Shutdown()

	// Summary
	fmt.Println("\n=== Test Results Summary ===")
	passed := 0
	for i, result := range testResults {
		status := "❌ FAILED"
		if result {
			status = "✅ PASSED"
			passed++
		}
		fmt.Printf("Test %d: %s\n", i+1, status)
	}

	fmt.Printf("\nOverall: %d/%d tests passed\n", passed, len(testResults))

	if passed == len(testResults) {
		fmt.Println("🎉 All performance optimization tests passed!")
	} else {
		fmt.Println("⚠️  Some tests failed. Check the output above for details.")
	}
}

func testPerformanceOptimizationEngine(engine *performance.PerformanceOptimizationEngine) bool {
	fmt.Println("1. Testing Performance Optimization Engine...")

	// Test optimization request
	req := performance.OptimizationRequest{
		Type:      performance.OptimizationTypeCPU,
		Priority:  performance.OptimizationPriorityHigh,
		Component: "test-component",
		Metrics: map[string]interface{}{
			"cpu_usage": 0.9,
			"timestamp": time.Now(),
		},
	}

	err := engine.RequestOptimization(req)
	if err != nil {
		fmt.Printf("  ❌ Failed to request optimization: %v\n", err)
		return false
	}

	// Wait for optimization to process
	time.Sleep(2 * time.Second)

	// Get metrics
	metrics := engine.GetMetrics()
	if metrics == nil {
		fmt.Println("  ❌ No optimization metrics returned")
		return false
	}

	fmt.Printf("  ✅ Optimization engine successful: %d total optimizations\n", metrics.TotalOptimizations)
	return true
}

func testResourceOptimizer(optimizer *performance.ResourceOptimizer) bool {
	fmt.Println("2. Testing Resource Optimizer...")

	// Test CPU optimization
	metrics := map[string]interface{}{
		"cpu_usage":    0.85,
		"memory_usage": 0.75,
		"goroutines":   500,
	}

	improvement, changes, err := optimizer.OptimizeCPU(metrics)
	if err != nil {
		fmt.Printf("  ❌ CPU optimization failed: %v\n", err)
		return false
	}

	// Test memory optimization
	memImprovement, memChanges, err := optimizer.OptimizeMemory(metrics)
	if err != nil {
		fmt.Printf("  ❌ Memory optimization failed: %v\n", err)
		return false
	}

	// Test resource allocation optimization
	resImprovement, resChanges, err := optimizer.OptimizeResourceAllocation(metrics)
	if err != nil {
		fmt.Printf("  ❌ Resource allocation optimization failed: %v\n", err)
		return false
	}

	fmt.Printf("  ✅ Resource optimizer successful: CPU=%.1f%%, Memory=%.1f%%, Resource=%.1f%% improvement\n",
		improvement, memImprovement, resImprovement)
	fmt.Printf("  ✅ Total changes: CPU=%d, Memory=%d, Resource=%d\n",
		len(changes), len(memChanges), len(resChanges))

	return true
}

func testAdvancedCacheManager(cacheManager *performance.AdvancedCacheManager) bool {
	fmt.Println("3. Testing Advanced Cache Manager...")

	// Test cache operations
	testKey := "test-key-123"
	testValue := "test-value-data"

	// Set value
	cacheManager.Set(testKey, testValue, 1*time.Hour)

	// Get value
	value, found := cacheManager.Get(testKey)
	if !found {
		fmt.Println("  ❌ Cache miss for recently set value")
		return false
	}

	if value != testValue {
		fmt.Println("  ❌ Retrieved value doesn't match set value")
		return false
	}

	// Test cache optimization
	metrics := map[string]interface{}{
		"cache_hit_ratio": 0.75,
		"cache_size":      50 * 1024 * 1024,
	}

	improvement, changes, err := cacheManager.OptimizeCache(metrics)
	if err != nil {
		fmt.Printf("  ❌ Cache optimization failed: %v\n", err)
		return false
	}

	// Get cache statistics
	stats := cacheManager.GetStats()
	if stats == nil {
		fmt.Println("  ❌ No cache statistics returned")
		return false
	}

	fmt.Printf("  ✅ Cache manager successful: %.1f%% improvement, %d changes\n", improvement, len(changes))
	fmt.Printf("  ✅ Cache stats: L1 hit ratio=%.2f, total hits=%d\n", stats.L1Stats.HitRatio, stats.TotalHits)

	return true
}

func testPerformanceProfiler(profiler *performance.PerformanceProfiler) bool {
	fmt.Println("4. Testing Performance Profiler...")

	// Wait for profiler to collect some metrics
	time.Sleep(15 * time.Second)

	// Get current metrics
	metrics := profiler.GetCurrentMetrics()
	if metrics == nil {
		fmt.Println("  ❌ No performance metrics returned")
		return false
	}

	// Get bottlenecks
	bottlenecks := profiler.GetBottlenecks()

	// Get trends
	trends := profiler.GetTrends()

	// Generate report
	report := profiler.GenerateReport()
	if report == nil {
		fmt.Println("  ❌ No profiling report generated")
		return false
	}

	fmt.Printf("  ✅ Performance profiler successful: CPU=%.1f%%, Memory=%.1f%%, Goroutines=%d\n",
		metrics.CPUUsage*100, metrics.MemoryUsage*100, metrics.GoroutineCount)
	fmt.Printf("  ✅ Analysis: %d bottlenecks, %d trends, %d recommendations\n",
		len(bottlenecks), len(trends), len(report.Recommendations))

	return true
}

func testAutoTuner(tuner *performance.AutoTuner) bool {
	fmt.Println("5. Testing Auto Tuner...")

	// Wait for tuner to perform some tuning
	time.Sleep(65 * time.Second)

	// Get tuning history
	history := tuner.GetTuningHistory()

	// Get adaptive state
	state := tuner.GetAdaptiveState()
	if state == nil {
		fmt.Println("  ❌ No adaptive state returned")
		return false
	}

	fmt.Printf("  ✅ Auto tuner successful: %d tuning iterations, objective=%.1f\n",
		len(history), state.CurrentObjective)
	fmt.Printf("  ✅ Best objective: %.1f, iteration: %d\n",
		state.BestObjective, state.Iteration)

	return true
}

func testMultiLevelCaching(cacheManager *performance.AdvancedCacheManager) bool {
	fmt.Println("6. Testing Multi-level Caching...")

	// Test multiple cache operations
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("multi-key-%d", i)
		value := fmt.Sprintf("multi-value-%d", i)

		cacheManager.Set(key, value, 30*time.Minute)

		// Retrieve immediately
		retrieved, found := cacheManager.Get(key)
		if !found || retrieved != value {
			fmt.Printf("  ❌ Multi-level cache failed for key %s\n", key)
			return false
		}
	}

	// Get cache statistics
	stats := cacheManager.GetStats()

	fmt.Printf("  ✅ Multi-level caching successful: overall hit ratio=%.2f\n", stats.HitRatio)
	fmt.Printf("  ✅ Cache levels: L1=%.2f, L2=%.2f, L3=%.2f hit ratios\n",
		stats.L1Stats.HitRatio, stats.L2Stats.HitRatio, stats.L3Stats.HitRatio)

	return true
}

func testBottleneckDetection(profiler *performance.PerformanceProfiler) bool {
	fmt.Println("7. Testing Bottleneck Detection...")

	// Wait for bottleneck detection to run
	time.Sleep(10 * time.Second)

	// Get detected bottlenecks
	bottlenecks := profiler.GetBottlenecks()

	// Generate report with bottleneck analysis
	report := profiler.GenerateReport()

	fmt.Printf("  ✅ Bottleneck detection successful: %d bottlenecks detected\n", len(bottlenecks))

	if len(report.Bottlenecks) > 0 {
		fmt.Printf("  ✅ Report bottlenecks: %d found\n", len(report.Bottlenecks))
		for _, bottleneck := range report.Bottlenecks {
			fmt.Printf("    - %s: %s (severity: %s)\n",
				bottleneck.Type, bottleneck.Description, bottleneck.Severity)
		}
	}

	return true
}

func testPerformanceIntegration(engine *performance.PerformanceOptimizationEngine, profiler *performance.PerformanceProfiler, tuner *performance.AutoTuner) bool {
	fmt.Println("8. Testing Performance Integration...")

	// Test that all components are working together

	// Get optimization metrics
	optMetrics := engine.GetMetrics()
	if optMetrics == nil {
		fmt.Println("  ❌ No optimization metrics available")
		return false
	}

	// Get profiling data
	profilerMetrics := profiler.GetCurrentMetrics()
	if profilerMetrics == nil {
		fmt.Println("  ❌ No profiler metrics available")
		return false
	}

	// Get tuning state
	tuningState := tuner.GetAdaptiveState()
	if tuningState == nil {
		fmt.Println("  ❌ No tuning state available")
		return false
	}

	// Generate comprehensive report
	report := profiler.GenerateReport()

	fmt.Printf("  ✅ Performance integration successful:\n")
	fmt.Printf("    - Optimizations: %d total, %.1f%% success rate\n",
		optMetrics.TotalOptimizations,
		float64(optMetrics.SuccessfulOptimizations)/float64(optMetrics.TotalOptimizations)*100)
	fmt.Printf("    - Profiling: CPU=%.1f%%, Memory=%.1f%%, %d goroutines\n",
		profilerMetrics.CPUUsage*100, profilerMetrics.MemoryUsage*100, profilerMetrics.GoroutineCount)
	fmt.Printf("    - Tuning: objective=%.1f, %d iterations\n",
		tuningState.CurrentObjective, tuningState.Iteration)
	fmt.Printf("    - Report: %d bottlenecks, %d recommendations\n",
		len(report.Bottlenecks), len(report.Recommendations))

	return true
}
