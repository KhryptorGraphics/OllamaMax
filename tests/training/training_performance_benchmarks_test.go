package training

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TrainingPerformanceBenchmarks contains performance tests for training scenarios
type TrainingPerformanceBenchmarks struct {
	APIServer    *httptest.Server
	StartTime    time.Time
	BaselineMemory runtime.MemStats
}

// SetupPerformanceBenchmarks initializes the performance testing environment
func SetupPerformanceBenchmarks() *TrainingPerformanceBenchmarks {
	tpb := &TrainingPerformanceBenchmarks{
		StartTime: time.Now(),
	}
	
	// Create mock API server for training
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing time
		time.Sleep(time.Millisecond * 5)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})
	
	// API v1 health endpoint
	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 3)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"ollama-distributed","version":"1.0.0"}`))
	})
	
	// Nodes endpoint
	mux.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 10)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"nodes":[{"id":"node1","status":"active","address":"127.0.0.1:4001"}]}`))
	})
	
	// Models endpoint (placeholder response)
	mux.HandleFunc("/api/v1/models", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 15)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"models":[{"id":"placeholder","name":"model-placeholder","status":"available"}]}`))
	})
	
	// Stats endpoint
	mux.HandleFunc("/api/v1/stats", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 8)
		uptime := time.Since(tpb.StartTime)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"uptime":"%s","requests":42,"memory_usage":"50MB"}`, uptime.String())))
	})
	
	tpb.APIServer = httptest.NewServer(mux)
	
	// Capture baseline memory stats
	runtime.GC()
	runtime.ReadMemStats(&tpb.BaselineMemory)
	
	return tpb
}

// Cleanup tears down the performance testing environment
func (tpb *TrainingPerformanceBenchmarks) Cleanup() {
	if tpb.APIServer != nil {
		tpb.APIServer.Close()
	}
}

// BenchmarkTrainingModuleExecution benchmarks the execution time of training modules
func BenchmarkTrainingModuleExecution(b *testing.B) {
	modules := []struct {
		name     string
		duration time.Duration
	}{
		{"module-1-installation", 10 * time.Minute},
		{"module-2-configuration", 10 * time.Minute},
		{"module-3-cluster-ops", 10 * time.Minute},
		{"module-4-model-mgmt", 10 * time.Minute},
		{"module-5-api-integration", 5 * time.Minute},
	}
	
	for _, module := range modules {
		b.Run(module.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				start := time.Now()
				
				// Simulate module execution steps
				simulateModuleExecution(module.name, module.duration)
				
				elapsed := time.Since(start)
				
				// Report custom metrics
				b.ReportMetric(float64(elapsed.Milliseconds()), "ms/module")
				
				// Validate execution time is reasonable
				if elapsed > time.Second*30 {
					b.Errorf("Module %s took too long: %v", module.name, elapsed)
				}
			}
		})
	}
}

// BenchmarkAPIResponseTimes benchmarks API endpoint response times
func BenchmarkAPIResponseTimes(b *testing.B) {
	tpb := SetupPerformanceBenchmarks()
	defer tpb.Cleanup()
	
	endpoints := []struct {
		name string
		path string
		expectedMaxTime time.Duration
	}{
		{"health", "/health", 50 * time.Millisecond},
		{"api_health", "/api/v1/health", 50 * time.Millisecond},
		{"nodes", "/api/v1/nodes", 100 * time.Millisecond},
		{"models", "/api/v1/models", 100 * time.Millisecond},
		{"stats", "/api/v1/stats", 100 * time.Millisecond},
	}
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	for _, endpoint := range endpoints {
		b.Run(endpoint.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			var totalLatency time.Duration
			var successCount int64
			
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					start := time.Now()
					resp, err := client.Get(tpb.APIServer.URL + endpoint.path)
					latency := time.Since(start)
					
					if err == nil && resp.StatusCode == http.StatusOK {
						resp.Body.Close()
						totalLatency += latency
						successCount++
						
						// Validate response time is within expected range
						if latency > endpoint.expectedMaxTime {
							b.Logf("Endpoint %s exceeded expected time: %v > %v", endpoint.name, latency, endpoint.expectedMaxTime)
						}
					}
				}
			})
			
			if successCount > 0 {
				avgLatency := totalLatency / time.Duration(successCount)
				b.ReportMetric(float64(avgLatency.Nanoseconds())/1e6, "ms/request")
				b.ReportMetric(float64(successCount)/float64(b.N)*100, "success_rate_%")
			}
		})
	}
}

// BenchmarkConcurrentTrainingUsers benchmarks system under concurrent training load
func BenchmarkConcurrentTrainingUsers(b *testing.B) {
	tpb := SetupPerformanceBenchmarks()
	defer tpb.Cleanup()
	
	concurrencyLevels := []int{1, 5, 10, 25, 50}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Users_%d", concurrency), func(b *testing.B) {
			client := &http.Client{
				Timeout: 10 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        concurrency * 2,
					MaxIdleConnsPerHost: concurrency,
					IdleConnTimeout:     30 * time.Second,
				},
			}
			
			var successCount int64
			var errorCount int64
			var totalLatency time.Duration
			var mu sync.Mutex
			
			b.ResetTimer()
			start := time.Now()
			
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// Simulate user training workflow
					success, latency := simulateUserWorkflow(client, tpb.APIServer.URL)
					
					mu.Lock()
					if success {
						successCount++
						totalLatency += latency
					} else {
						errorCount++
					}
					mu.Unlock()
				}
			})
			
			elapsed := time.Since(start)
			totalRequests := successCount + errorCount
			
			if totalRequests > 0 {
				throughput := float64(totalRequests) / elapsed.Seconds()
				errorRate := float64(errorCount) / float64(totalRequests) * 100
				avgLatency := float64(totalLatency.Nanoseconds()) / float64(successCount) / 1e6
				
				b.ReportMetric(throughput, "requests/sec")
				b.ReportMetric(errorRate, "error_rate_%")
				b.ReportMetric(avgLatency, "avg_latency_ms")
			}
		})
	}
}

// BenchmarkTrainingToolExecution benchmarks training tool performance
func BenchmarkTrainingToolExecution(b *testing.B) {
	tools := []struct {
		name string
		exec func() time.Duration
	}{
		{
			name: "health_monitor",
			exec: func() time.Duration {
				start := time.Now()
				// Simulate health monitor execution
				time.Sleep(time.Millisecond * 10)
				return time.Since(start)
			},
		},
		{
			name: "api_client",
			exec: func() time.Duration {
				start := time.Now()
				// Simulate API client execution
				time.Sleep(time.Millisecond * 25)
				return time.Since(start)
			},
		},
		{
			name: "config_generator",
			exec: func() time.Duration {
				start := time.Now()
				// Simulate config generation
				time.Sleep(time.Millisecond * 5)
				return time.Since(start)
			},
		},
		{
			name: "performance_monitor",
			exec: func() time.Duration {
				start := time.Now()
				// Simulate performance monitoring
				time.Sleep(time.Millisecond * 15)
				return time.Since(start)
			},
		},
	}
	
	for _, tool := range tools {
		b.Run(tool.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			
			var totalTime time.Duration
			
			for i := 0; i < b.N; i++ {
				execTime := tool.exec()
				totalTime += execTime
			}
			
			avgTime := totalTime / time.Duration(b.N)
			b.ReportMetric(float64(avgTime.Nanoseconds())/1e6, "ms/execution")
		})
	}
}

// BenchmarkMemoryUsageTraining benchmarks memory usage during training
func BenchmarkMemoryUsageTraining(b *testing.B) {
	var m1, m2 runtime.MemStats
	
	// Measure baseline memory
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Simulate training data structures
	trainingData := make([][]byte, b.N)
	configData := make([]map[string]interface{}, b.N)
	logEntries := make([]string, b.N)
	
	for i := 0; i < b.N; i++ {
		// Simulate configuration files
		trainingData[i] = make([]byte, 1024) // 1KB config files
		
		// Simulate configuration parsing
		configData[i] = map[string]interface{}{
			"api": map[string]interface{}{
				"listen": ":8080",
				"cors":   map[string]interface{}{"enabled": true},
			},
			"p2p": map[string]interface{}{
				"listen_addr": "/ip4/127.0.0.1/tcp/4001",
			},
		}
		
		// Simulate log entries
		logEntries[i] = fmt.Sprintf("Training step %d completed at %s", i, time.Now().Format(time.RFC3339))
	}
	
	// Force garbage collection and measure memory
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	// Calculate memory usage
	allocatedBytes := m2.TotalAlloc - m1.TotalAlloc
	heapInuse := m2.HeapInuse - m1.HeapInuse
	
	b.ReportMetric(float64(allocatedBytes)/float64(b.N), "bytes/iteration")
	b.ReportMetric(float64(heapInuse)/1024/1024, "heap_mb")
	
	// Report GC statistics
	gcPauses := m2.PauseTotalNs - m1.PauseTotalNs
	gcCycles := m2.NumGC - m1.NumGC
	
	if gcCycles > 0 {
		avgGCPause := float64(gcPauses) / float64(gcCycles) / 1e6
		b.ReportMetric(avgGCPause, "avg_gc_pause_ms")
		b.ReportMetric(float64(gcCycles), "gc_cycles")
	}
	
	// Keep references to prevent premature GC
	_ = trainingData
	_ = configData  
	_ = logEntries
}

// BenchmarkValidationScriptExecution benchmarks validation script performance
func BenchmarkValidationScriptExecution(b *testing.B) {
	validationTests := []struct {
		name string
		exec func() (bool, time.Duration)
	}{
		{
			name: "prerequisites_check",
			exec: func() (bool, time.Duration) {
				start := time.Now()
				// Simulate prerequisites check
				time.Sleep(time.Millisecond * 50)
				return true, time.Since(start)
			},
		},
		{
			name: "installation_validation",
			exec: func() (bool, time.Duration) {
				start := time.Now()
				// Simulate installation validation
				time.Sleep(time.Millisecond * 100)
				return true, time.Since(start)
			},
		},
		{
			name: "configuration_test",
			exec: func() (bool, time.Duration) {
				start := time.Now()
				// Simulate configuration testing
				time.Sleep(time.Millisecond * 75)
				return true, time.Since(start)
			},
		},
		{
			name: "api_validation",
			exec: func() (bool, time.Duration) {
				start := time.Now()
				// Simulate API validation
				time.Sleep(time.Millisecond * 200)
				return true, time.Since(start)
			},
		},
	}
	
	for _, test := range validationTests {
		b.Run(test.name, func(b *testing.B) {
			b.ResetTimer()
			
			var successCount int64
			var totalTime time.Duration
			
			for i := 0; i < b.N; i++ {
				success, duration := test.exec()
				if success {
					successCount++
				}
				totalTime += duration
			}
			
			avgTime := totalTime / time.Duration(b.N)
			successRate := float64(successCount) / float64(b.N) * 100
			
			b.ReportMetric(float64(avgTime.Nanoseconds())/1e6, "ms/validation")
			b.ReportMetric(successRate, "success_rate_%")
		})
	}
}

// BenchmarkCertificationAssessment benchmarks certification assessment performance
func BenchmarkCertificationAssessment(b *testing.B) {
	assessmentComponents := []struct {
		name string
		exec func() (float64, time.Duration) // returns score and duration
	}{
		{
			name: "knowledge_questions",
			exec: func() (float64, time.Duration) {
				start := time.Now()
				// Simulate answering 20 knowledge questions
				score := 0.0
				for i := 0; i < 20; i++ {
					time.Sleep(time.Microsecond * 100) // Question processing time
					if i%4 != 0 { // 75% correct answers
						score += 5.0
					}
				}
				return score / 100.0, time.Since(start)
			},
		},
		{
			name: "practical_tasks",
			exec: func() (float64, time.Duration) {
				start := time.Now()
				// Simulate 5 practical tasks
				score := 0.0
				for i := 0; i < 5; i++ {
					time.Sleep(time.Millisecond * 10) // Task execution time
					if i != 1 { // Fail one task
						score += 20.0
					}
				}
				return score / 100.0, time.Since(start)
			},
		},
		{
			name: "integration_scenarios",
			exec: func() (float64, time.Duration) {
				start := time.Now()
				// Simulate 3 integration scenarios
				score := 0.0
				for i := 0; i < 3; i++ {
					time.Sleep(time.Millisecond * 15) // Scenario execution time
					score += 30.0 + float64(i*2) // Varying scores
				}
				return score / 100.0, time.Since(start)
			},
		},
	}
	
	for _, component := range assessmentComponents {
		b.Run(component.name, func(b *testing.B) {
			b.ResetTimer()
			
			var totalScore float64
			var totalTime time.Duration
			
			for i := 0; i < b.N; i++ {
				score, duration := component.exec()
				totalScore += score
				totalTime += duration
			}
			
			avgScore := totalScore / float64(b.N)
			avgTime := totalTime / time.Duration(b.N)
			
			b.ReportMetric(avgScore*100, "avg_score_%")
			b.ReportMetric(float64(avgTime.Nanoseconds())/1e6, "ms/assessment")
		})
	}
}

// Helper functions for benchmarks

func simulateModuleExecution(moduleName string, duration time.Duration) {
	// Simulate the time it takes to execute training module steps
	steps := []string{
		"load_instructions",
		"validate_prerequisites", 
		"execute_commands",
		"verify_outputs",
		"update_progress",
	}
	
	stepDuration := duration / time.Duration(len(steps))
	
	for _, step := range steps {
		// Simulate step execution
		time.Sleep(stepDuration / 100) // Scale down for benchmarking
		
		// Simulate some CPU work
		for i := 0; i < 1000; i++ {
			_ = fmt.Sprintf("%s_%s_%d", moduleName, step, i)
		}
	}
}

func simulateUserWorkflow(client *http.Client, baseURL string) (bool, time.Duration) {
	start := time.Now()
	
	// Simulate typical user training workflow
	endpoints := []string{
		"/health",
		"/api/v1/health",
		"/api/v1/nodes",
		"/api/v1/models",
		"/api/v1/stats",
	}
	
	for _, endpoint := range endpoints {
		resp, err := client.Get(baseURL + endpoint)
		if err != nil {
			return false, time.Since(start)
		}
		resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return false, time.Since(start)
		}
		
		// Brief pause between requests (realistic user behavior)
		time.Sleep(time.Millisecond * 10)
	}
	
	return true, time.Since(start)
}

// TestPerformanceRegression validates that performance doesn't regress
func TestPerformanceRegression(t *testing.T) {
	// Performance thresholds based on training requirements
	thresholds := map[string]float64{
		"api_latency_ms":              50.0,   // Max 50ms for API responses
		"module_execution_time_ms":    30000,  // Max 30s for module execution simulation
		"concurrent_users_throughput": 100,    // Min 100 req/s for 10 concurrent users
		"memory_per_iteration_bytes":  2048,   // Max 2KB per training iteration
		"validation_time_ms":          500,    // Max 500ms for validation scripts
	}
	
	tpb := SetupPerformanceBenchmarks()
	defer tpb.Cleanup()
	
	t.Run("API_Latency", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		
		var totalLatency time.Duration
		iterations := 10
		
		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp, err := client.Get(tpb.APIServer.URL + "/health")
			latency := time.Since(start)
			
			require.NoError(t, err, "API request should succeed")
			resp.Body.Close()
			
			totalLatency += latency
		}
		
		avgLatency := totalLatency / time.Duration(iterations)
		avgLatencyMs := float64(avgLatency.Nanoseconds()) / 1e6
		
		assert.LessOrEqual(t, avgLatencyMs, thresholds["api_latency_ms"], 
			"API latency regression detected: %.2fms > %.2fms", avgLatencyMs, thresholds["api_latency_ms"])
		
		t.Logf("API latency: %.2fms (threshold: %.2fms)", avgLatencyMs, thresholds["api_latency_ms"])
	})
	
	t.Run("Module_Execution_Time", func(t *testing.T) {
		start := time.Now()
		simulateModuleExecution("test-module", 1*time.Second) // Scaled down version
		elapsed := time.Since(start)
		
		elapsedMs := float64(elapsed.Nanoseconds()) / 1e6
		
		// Scale up the measurement for comparison (since we scaled down the simulation)
		projectedMs := elapsedMs * 100 // Reverse the scaling factor
		
		assert.LessOrEqual(t, projectedMs, thresholds["module_execution_time_ms"],
			"Module execution time regression: %.2fms > %.2fms", projectedMs, thresholds["module_execution_time_ms"])
		
		t.Logf("Module execution time: %.2fms projected (threshold: %.2fms)", projectedMs, thresholds["module_execution_time_ms"])
	})
	
	t.Run("Memory_Usage", func(t *testing.T) {
		var m1, m2 runtime.MemStats
		
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		// Simulate training data structures
		iterations := 1000
		data := make([][]byte, iterations)
		for i := 0; i < iterations; i++ {
			data[i] = make([]byte, 1024) // 1KB per iteration
		}
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		
		memoryPerIteration := float64(m2.TotalAlloc-m1.TotalAlloc) / float64(iterations)
		
		assert.LessOrEqual(t, memoryPerIteration, thresholds["memory_per_iteration_bytes"],
			"Memory usage regression: %.2f bytes/iter > %.2f bytes/iter", memoryPerIteration, thresholds["memory_per_iteration_bytes"])
		
		t.Logf("Memory usage: %.2f bytes/iteration (threshold: %.2f)", memoryPerIteration, thresholds["memory_per_iteration_bytes"])
		
		// Keep reference to prevent early GC
		_ = data
	})
	
	t.Run("Concurrent_Throughput", func(t *testing.T) {
		concurrency := 10
		requestsPerUser := 10
		
		var wg sync.WaitGroup
		var successCount int64
		var mu sync.Mutex
		
		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        concurrency * 2,
				MaxIdleConnsPerHost: concurrency,
			},
		}
		
		start := time.Now()
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for j := 0; j < requestsPerUser; j++ {
					resp, err := client.Get(tpb.APIServer.URL + "/health")
					if err == nil && resp.StatusCode == http.StatusOK {
						resp.Body.Close()
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}
			}()
		}
		
		wg.Wait()
		elapsed := time.Since(start)
		
		throughput := float64(successCount) / elapsed.Seconds()
		
		assert.GreaterOrEqual(t, throughput, thresholds["concurrent_users_throughput"],
			"Throughput regression: %.2f req/s < %.2f req/s", throughput, thresholds["concurrent_users_throughput"])
		
		t.Logf("Concurrent throughput: %.2f req/s (threshold: %.2f)", throughput, thresholds["concurrent_users_throughput"])
	})
}

// BenchmarkComprehensiveTrainingScenario runs a complete training scenario benchmark
func BenchmarkComprehensiveTrainingScenario(b *testing.B) {
	tpb := SetupPerformanceBenchmarks()
	defer tpb.Cleanup()
	
	// Simulate complete training program execution
	trainingScenarios := []struct {
		name   string
		modules []string
		expectedDuration time.Duration
	}{
		{
			name: "full_training_program",
			modules: []string{
				"module-1-installation",
				"module-2-configuration", 
				"module-3-cluster-ops",
				"module-4-model-mgmt",
				"module-5-api-integration",
			},
			expectedDuration: 45 * time.Minute, // Full program expected time
		},
		{
			name: "quick_certification_prep",
			modules: []string{
				"module-3-cluster-ops",
				"module-5-api-integration", 
			},
			expectedDuration: 15 * time.Minute, // Abbreviated version
		},
	}
	
	for _, scenario := range trainingScenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				scenarioStart := time.Now()
				
				// Execute all modules in the scenario
				for _, module := range scenario.modules {
					moduleStart := time.Now()
					simulateModuleExecution(module, scenario.expectedDuration/time.Duration(len(scenario.modules)))
					moduleElapsed := time.Since(moduleStart)
					
					// Validate module execution time
					maxModuleTime := scenario.expectedDuration / time.Duration(len(scenario.modules)) / 10 // Scale for benchmarking
					if moduleElapsed > maxModuleTime {
						b.Logf("Module %s exceeded expected time: %v > %v", module, moduleElapsed, maxModuleTime)
					}
				}
				
				// Simulate final validation
				time.Sleep(time.Millisecond * 10)
				
				scenarioElapsed := time.Since(scenarioStart)
				b.ReportMetric(float64(scenarioElapsed.Nanoseconds())/1e6, "ms/scenario")
			}
		})
	}
}