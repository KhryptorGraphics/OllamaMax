package performance

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/tests/integration"
)

// PerformanceTestSuite provides comprehensive performance testing
type PerformanceTestSuite struct {
	cluster  *integration.TestCluster
	metrics  *PerformanceMetrics
	baseline *BaselineMetrics
	ctx      context.Context
	cancel   context.CancelFunc
}

// PerformanceMetrics tracks comprehensive performance data
type PerformanceMetrics struct {
	mu              sync.RWMutex
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedReqs      int64
	TotalLatency    time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	ThroughputRPS   float64
	MemoryUsageMB   float64
	CPUUsagePercent float64
	NetworkBytesIn  int64
	NetworkBytesOut int64
	ConcurrentUsers int
	ErrorRate       float64
	StartTime       time.Time
	EndTime         time.Time
	LatencyHistory  []time.Duration
}

// BaselineMetrics stores baseline performance expectations
type BaselineMetrics struct {
	MaxLatencyP95      time.Duration
	MinThroughputRPS   float64
	MaxMemoryUsageMB   float64
	MaxCPUUsagePercent float64
	MaxErrorRate       float64
}

// LoadTestConfig configures load testing parameters
type LoadTestConfig struct {
	ConcurrentUsers  int
	RequestsPerUser  int
	RampUpDuration   time.Duration
	SustainDuration  time.Duration
	RampDownDuration time.Duration
	RequestInterval  time.Duration
	Model            string
	PromptTemplate   string
}

// NewPerformanceTestSuite creates a new performance test suite
func NewPerformanceTestSuite(nodeCount int) (*PerformanceTestSuite, error) {
	cluster, err := integration.NewTestCluster(nodeCount)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Define baseline performance expectations
	baseline := &BaselineMetrics{
		MaxLatencyP95:      2 * time.Second,
		MinThroughputRPS:   10.0,
		MaxMemoryUsageMB:   1024.0, // 1GB per node
		MaxCPUUsagePercent: 80.0,
		MaxErrorRate:       0.05, // 5%
	}

	return &PerformanceTestSuite{
		cluster:  cluster,
		metrics:  &PerformanceMetrics{StartTime: time.Now()},
		baseline: baseline,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// TestPerformance runs comprehensive performance tests
func TestPerformance(t *testing.T) {
	suite, err := NewPerformanceTestSuite(3)
	require.NoError(t, err)
	defer suite.cleanup()

	// Start cluster
	require.NoError(t, suite.cluster.Start())
	time.Sleep(10 * time.Second)

	t.Run("BaselinePerformance", func(t *testing.T) {
		suite.testBaselinePerformance(t)
	})

	t.Run("LoadTesting", func(t *testing.T) {
		suite.testLoadPerformance(t)
	})

	t.Run("StressTesting", func(t *testing.T) {
		suite.testStressPerformance(t)
	})

	t.Run("ScalabilityTesting", func(t *testing.T) {
		suite.testScalabilityPerformance(t)
	})

	t.Run("MemoryLeakTesting", func(t *testing.T) {
		suite.testMemoryLeaks(t)
	})

	t.Run("ConcurrencyTesting", func(t *testing.T) {
		suite.testConcurrencyPerformance(t)
	})

	t.Run("NetworkThroughput", func(t *testing.T) {
		suite.testNetworkThroughput(t)
	})

	// Generate performance report
	suite.generatePerformanceReport(t)
}

// testBaselinePerformance establishes baseline performance metrics
func (suite *PerformanceTestSuite) testBaselinePerformance(t *testing.T) {
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader for baseline testing")

	// Register test model
	err := suite.registerTestModel(leader)
	require.NoError(t, err, "Model registration should succeed")

	// Single request baseline
	t.Run("SingleRequestBaseline", func(t *testing.T) {
		req := &api.InferenceRequest{
			Model:  "test-model-1b",
			Prompt: "Hello, world! This is a baseline performance test.",
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  50,
			},
		}

		start := time.Now()
		resp, err := leader.ProcessInference(suite.ctx, req)
		latency := time.Since(start)

		assert.NoError(t, err, "Baseline request should succeed")
		assert.NotNil(t, resp, "Response should not be nil")
		assert.Less(t, latency, 5*time.Second, "Baseline latency should be reasonable")

		t.Logf("Baseline single request latency: %v", latency)
		suite.updateMetrics(latency, err == nil)
	})

	// Sequential requests baseline
	t.Run("SequentialRequestsBaseline", func(t *testing.T) {
		requestCount := 10
		var totalLatency time.Duration

		for i := 0; i < requestCount; i++ {
			req := &api.InferenceRequest{
				Model:  "test-model-1b",
				Prompt: fmt.Sprintf("Sequential baseline test request %d", i),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  30,
				},
			}

			start := time.Now()
			_, err := leader.ProcessInference(suite.ctx, req)
			latency := time.Since(start)
			totalLatency += latency

			suite.updateMetrics(latency, err == nil)
		}

		avgLatency := totalLatency / time.Duration(requestCount)
		t.Logf("Baseline sequential average latency: %v", avgLatency)
		assert.Less(t, avgLatency, 3*time.Second, "Sequential average latency should be reasonable")
	})
}

// testLoadPerformance tests performance under various load conditions
func (suite *PerformanceTestSuite) testLoadPerformance(t *testing.T) {
	loadConfigs := []LoadTestConfig{
		{
			ConcurrentUsers:  5,
			RequestsPerUser:  10,
			RampUpDuration:   10 * time.Second,
			SustainDuration:  30 * time.Second,
			RampDownDuration: 5 * time.Second,
			RequestInterval:  500 * time.Millisecond,
			Model:            "test-model-1b",
			PromptTemplate:   "Load test request %d from user %d",
		},
		{
			ConcurrentUsers:  10,
			RequestsPerUser:  15,
			RampUpDuration:   15 * time.Second,
			SustainDuration:  45 * time.Second,
			RampDownDuration: 10 * time.Second,
			RequestInterval:  300 * time.Millisecond,
			Model:            "test-model-1b",
			PromptTemplate:   "Medium load test request %d from user %d",
		},
		{
			ConcurrentUsers:  20,
			RequestsPerUser:  20,
			RampUpDuration:   20 * time.Second,
			SustainDuration:  60 * time.Second,
			RampDownDuration: 15 * time.Second,
			RequestInterval:  200 * time.Millisecond,
			Model:            "test-model-1b",
			PromptTemplate:   "High load test request %d from user %d",
		},
	}

	for i, config := range loadConfigs {
		t.Run(fmt.Sprintf("LoadTest_%d_Users_%d", i+1, config.ConcurrentUsers), func(t *testing.T) {
			suite.runLoadTest(t, config)
		})
	}
}

// runLoadTest executes a load test with the given configuration
func (suite *PerformanceTestSuite) runLoadTest(t *testing.T, config LoadTestConfig) {
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader for load testing")

	// Reset metrics for this test
	suite.resetMetrics()

	// Track resource usage before test
	initialMemory := suite.getMemoryUsage()
	initialCPU := suite.getCPUUsage()

	var wg sync.WaitGroup
	results := make(chan LoadTestResult, config.ConcurrentUsers*config.RequestsPerUser)

	startTime := time.Now()

	// Ramp up phase
	t.Logf("Ramping up %d users over %v", config.ConcurrentUsers, config.RampUpDuration)
	userStartInterval := config.RampUpDuration / time.Duration(config.ConcurrentUsers)

	for userID := 0; userID < config.ConcurrentUsers; userID++ {
		time.Sleep(userStartInterval) // Gradual ramp-up

		wg.Add(1)
		go suite.loadTestUser(userID, config, results, &wg)
	}

	// Monitor system resources during test
	resourceMonitorCtx, resourceCancel := context.WithCancel(suite.ctx)
	go suite.monitorResources(resourceMonitorCtx)

	// Wait for all users to complete
	wg.Wait()
	close(results)
	resourceCancel()

	// Collect and analyze results
	suite.analyzeLoadTestResults(t, results, startTime, config)

	// Check resource usage after test
	finalMemory := suite.getMemoryUsage()
	finalCPU := suite.getCPUUsage()

	t.Logf("Memory usage: %0.2f MB -> %.2f MB (delta: %.2f MB)",
		initialMemory, finalMemory, finalMemory-initialMemory)
	t.Logf("CPU usage: %.2f%% -> %.2f%% (delta: %.2f%%)",
		initialCPU, finalCPU, finalCPU-initialCPU)

	// Validate against baseline
	suite.validateAgainstBaseline(t, config)
}

// loadTestUser simulates a single user's load testing behavior
func (suite *PerformanceTestSuite) loadTestUser(userID int, config LoadTestConfig, results chan<- LoadTestResult, wg *sync.WaitGroup) {
	defer wg.Done()

	leader := suite.cluster.GetLeader()
	if leader == nil {
		return
	}

	for reqID := 0; reqID < config.RequestsPerUser; reqID++ {
		start := time.Now()

		req := &api.InferenceRequest{
			Model:  config.Model,
			Prompt: fmt.Sprintf(config.PromptTemplate, reqID, userID),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  40,
			},
		}

		resp, err := leader.ProcessInference(suite.ctx, req)
		latency := time.Since(start)

		result := LoadTestResult{
			UserID:    userID,
			RequestID: reqID,
			Latency:   latency,
			Success:   err == nil,
			Error:     err,
			Timestamp: start,
		}

		if err == nil && resp != nil {
			result.ResponseSize = len(resp.Response)
		}

		results <- result
		suite.updateMetrics(latency, err == nil)

		// Wait before next request (if not the last one)
		if reqID < config.RequestsPerUser-1 {
			time.Sleep(config.RequestInterval)
		}
	}
}

// analyzeLoadTestResults analyzes the results of a load test
func (suite *PerformanceTestSuite) analyzeLoadTestResults(t *testing.T, results <-chan LoadTestResult, startTime time.Time, config LoadTestConfig) {
	var (
		totalRequests     int
		successfulReqs    int
		failedReqs        int
		latencies         []time.Duration
		totalLatency      time.Duration
		minLatency        = time.Duration(math.MaxInt64)
		maxLatency        time.Duration
		totalResponseSize int64
	)

	// Collect all results
	for result := range results {
		totalRequests++
		totalLatency += result.Latency
		latencies = append(latencies, result.Latency)
		totalResponseSize += int64(result.ResponseSize)

		if result.Success {
			successfulReqs++
		} else {
			failedReqs++
			if result.Error != nil {
				t.Logf("Request error: %v", result.Error)
			}
		}

		if result.Latency < minLatency {
			minLatency = result.Latency
		}
		if result.Latency > maxLatency {
			maxLatency = result.Latency
		}
	}

	// Calculate metrics
	testDuration := time.Since(startTime)
	avgLatency := totalLatency / time.Duration(totalRequests)
	throughput := float64(successfulReqs) / testDuration.Seconds()
	errorRate := float64(failedReqs) / float64(totalRequests)

	// Calculate percentiles
	p50, p95, p99 := calculatePercentiles(latencies)

	// Update suite metrics
	suite.metrics.mu.Lock()
	suite.metrics.TotalRequests = int64(totalRequests)
	suite.metrics.SuccessfulReqs = int64(successfulReqs)
	suite.metrics.FailedReqs = int64(failedReqs)
	suite.metrics.MinLatency = minLatency
	suite.metrics.MaxLatency = maxLatency
	suite.metrics.P50Latency = p50
	suite.metrics.P95Latency = p95
	suite.metrics.P99Latency = p99
	suite.metrics.ThroughputRPS = throughput
	suite.metrics.ErrorRate = errorRate
	suite.metrics.ConcurrentUsers = config.ConcurrentUsers
	suite.metrics.EndTime = time.Now()
	suite.metrics.mu.Unlock()

	// Log results
	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Successful: %d (%.1f%%)", successfulReqs, float64(successfulReqs)/float64(totalRequests)*100)
	t.Logf("  Failed: %d (%.1f%%)", failedReqs, errorRate*100)
	t.Logf("  Test Duration: %v", testDuration)
	t.Logf("  Throughput: %.2f RPS", throughput)
	t.Logf("  Latency - Min: %v, Avg: %v, Max: %v", minLatency, avgLatency, maxLatency)
	t.Logf("  Latency Percentiles - P50: %v, P95: %v, P99: %v", p50, p95, p99)
	t.Logf("  Average Response Size: %.2f KB", float64(totalResponseSize)/float64(successfulReqs)/1024)

	// Assertions
	assert.Greater(t, throughput, 1.0, "Throughput should be > 1 RPS")
	assert.Less(t, errorRate, 0.1, "Error rate should be < 10%")
	assert.Less(t, p95, 10*time.Second, "P95 latency should be < 10s")
}

// testStressPerformance tests system behavior under stress conditions
func (suite *PerformanceTestSuite) testStressPerformance(t *testing.T) {
	t.Run("HighConcurrencyStress", func(t *testing.T) {
		config := LoadTestConfig{
			ConcurrentUsers:  50,
			RequestsPerUser:  10,
			RampUpDuration:   30 * time.Second,
			SustainDuration:  60 * time.Second,
			RampDownDuration: 20 * time.Second,
			RequestInterval:  100 * time.Millisecond,
			Model:            "test-model-1b",
			PromptTemplate:   "Stress test request %d from user %d",
		}

		suite.runLoadTest(t, config)

		// Stress test should still maintain reasonable performance
		assert.Greater(t, suite.metrics.ThroughputRPS, 5.0, "Throughput should remain > 5 RPS under stress")
		assert.Less(t, suite.metrics.ErrorRate, 0.2, "Error rate should be < 20% under stress")
	})

	t.Run("SustainedLoadStress", func(t *testing.T) {
		config := LoadTestConfig{
			ConcurrentUsers:  25,
			RequestsPerUser:  50,
			RampUpDuration:   20 * time.Second,
			SustainDuration:  5 * time.Minute,
			RampDownDuration: 15 * time.Second,
			RequestInterval:  200 * time.Millisecond,
			Model:            "test-model-1b",
			PromptTemplate:   "Sustained stress test request %d from user %d",
		}

		initialMemory := suite.getMemoryUsage()
		suite.runLoadTest(t, config)
		finalMemory := suite.getMemoryUsage()

		// Check for memory growth (potential leaks)
		memoryGrowth := finalMemory - initialMemory
		assert.Less(t, memoryGrowth, 500.0, "Memory growth should be < 500MB during sustained load")
	})
}

// testScalabilityPerformance tests system scalability
func (suite *PerformanceTestSuite) testScalabilityPerformance(t *testing.T) {
	concurrencyLevels := []int{1, 5, 10, 20, 30}
	results := make([]ScalabilityResult, 0, len(concurrencyLevels))

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			config := LoadTestConfig{
				ConcurrentUsers:  concurrency,
				RequestsPerUser:  20,
				RampUpDuration:   10 * time.Second,
				SustainDuration:  30 * time.Second,
				RampDownDuration: 5 * time.Second,
				RequestInterval:  250 * time.Millisecond,
				Model:            "test-model-1b",
				PromptTemplate:   "Scalability test request %d from user %d",
			}

			suite.runLoadTest(t, config)

			result := ScalabilityResult{
				ConcurrentUsers: concurrency,
				ThroughputRPS:   suite.metrics.ThroughputRPS,
				AvgLatency:      suite.metrics.P50Latency,
				P95Latency:      suite.metrics.P95Latency,
				ErrorRate:       suite.metrics.ErrorRate,
				MemoryUsage:     suite.getMemoryUsage(),
				CPUUsage:        suite.getCPUUsage(),
			}

			results = append(results, result)
			t.Logf("Scalability Result - Users: %d, Throughput: %.2f RPS, P95: %v, Errors: %.2f%%",
				concurrency, result.ThroughputRPS, result.P95Latency, result.ErrorRate*100)
		})
	}

	// Analyze scalability trends
	suite.analyzeScalabilityTrends(t, results)
}

// testMemoryLeaks tests for memory leaks during extended operation
func (suite *PerformanceTestSuite) testMemoryLeaks(t *testing.T) {
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader for memory leak testing")

	// Take initial memory snapshot
	initialMemory := suite.getMemoryUsage()
	memorySnapshots := []MemorySnapshot{
		{Timestamp: time.Now(), MemoryMB: initialMemory},
	}

	// Run continuous load for extended period
	testDuration := 5 * time.Minute
	snapshotInterval := 30 * time.Second
	requestInterval := 100 * time.Millisecond

	endTime := time.Now().Add(testDuration)
	lastSnapshot := time.Now()

	requestCount := 0
	for time.Now().Before(endTime) {
		// Send request
		req := &api.InferenceRequest{
			Model:  "test-model-1b",
			Prompt: fmt.Sprintf("Memory leak test request %d", requestCount),
			Options: map[string]interface{}{
				"temperature": 0.1,
				"max_tokens":  30,
			},
		}

		_, err := leader.ProcessInference(suite.ctx, req)
		if err != nil {
			t.Logf("Request %d failed: %v", requestCount, err)
		}

		requestCount++

		// Take memory snapshot
		if time.Since(lastSnapshot) >= snapshotInterval {
			currentMemory := suite.getMemoryUsage()
			memorySnapshots = append(memorySnapshots, MemorySnapshot{
				Timestamp: time.Now(),
				MemoryMB:  currentMemory,
			})
			lastSnapshot = time.Now()
		}

		time.Sleep(requestInterval)
	}

	// Analyze memory usage trend
	suite.analyzeMemoryTrend(t, memorySnapshots)

	t.Logf("Memory leak test completed with %d requests", requestCount)
}

// testConcurrencyPerformance tests concurrent access patterns
func (suite *PerformanceTestSuite) testConcurrencyPerformance(t *testing.T) {
	t.Run("ReadWriteConcurrency", func(t *testing.T) {
		// Test concurrent reads and writes to consensus
		leader := suite.cluster.GetLeader()
		require.NotNil(t, leader, "Need a leader")

		var wg sync.WaitGroup
		errorChan := make(chan error, 100)

		// Concurrent writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(writerID int) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					key := fmt.Sprintf("perf_test_%d_%d", writerID, j)
					value := fmt.Sprintf("value_%d_%d", writerID, j)
					err := leader.ApplyConsensusOperation(key, value)
					if err != nil {
						errorChan <- err
					}
				}
			}(i)
		}

		// Concurrent readers
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(readerID int) {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					key := fmt.Sprintf("perf_test_%d_%d", readerID%10, j%20)
					_, exists := leader.GetConsensusValue(key)
					if !exists {
						// This is expected for some keys due to timing
					}
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for errors
		errorCount := 0
		for err := range errorChan {
			errorCount++
			t.Logf("Concurrency error: %v", err)
		}

		assert.Less(t, errorCount, 10, "Should have minimal errors during concurrent operations")
	})
}

// testNetworkThroughput tests network throughput capabilities
func (suite *PerformanceTestSuite) testNetworkThroughput(t *testing.T) {
	leader := suite.cluster.GetLeader()
	require.NotNil(t, leader, "Need a leader")

	// Test with various payload sizes
	payloadSizes := []int{1024, 10240, 102400} // 1KB, 10KB, 100KB

	for _, size := range payloadSizes {
		t.Run(fmt.Sprintf("PayloadSize_%dB", size), func(t *testing.T) {
			// Generate payload of specified size
			payload := generatePayload(size)

			req := &api.InferenceRequest{
				Model:  "test-model-1b",
				Prompt: payload,
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  50,
				},
			}

			start := time.Now()
			resp, err := leader.ProcessInference(suite.ctx, req)
			latency := time.Since(start)

			assert.NoError(t, err, "Large payload request should succeed")
			assert.NotNil(t, resp, "Response should not be nil")

			throughputMBps := float64(size+len(resp.Response)) / (1024 * 1024) / latency.Seconds()
			t.Logf("Payload size: %d bytes, Latency: %v, Throughput: %.2f MB/s",
				size, latency, throughputMBps)

			// Validate reasonable throughput
			assert.Greater(t, throughputMBps, 0.1, "Throughput should be > 0.1 MB/s")
		})
	}
}

// Helper functions and types

type LoadTestResult struct {
	UserID       int
	RequestID    int
	Latency      time.Duration
	Success      bool
	Error        error
	ResponseSize int
	Timestamp    time.Time
}

type ScalabilityResult struct {
	ConcurrentUsers int
	ThroughputRPS   float64
	AvgLatency      time.Duration
	P95Latency      time.Duration
	ErrorRate       float64
	MemoryUsage     float64
	CPUUsage        float64
}

type MemorySnapshot struct {
	Timestamp time.Time
	MemoryMB  float64
}

// Implementation of helper methods

func (suite *PerformanceTestSuite) registerTestModel(leader *integration.TestNode) error {
	modelInfo := &integration.ModelInfo{
		Name:              "test-model-1b",
		Size:              1024 * 1024 * 1024, // 1GB
		Type:              "language_model",
		ReplicationFactor: 2,
		Checksum:          "test-checksum-1b",
		LastAccessed:      time.Now(),
		Popularity:        0.8,
	}

	return leader.RegisterModel(modelInfo)
}

func (suite *PerformanceTestSuite) updateMetrics(latency time.Duration, success bool) {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()

	suite.metrics.TotalRequests++
	suite.metrics.TotalLatency += latency
	suite.metrics.LatencyHistory = append(suite.metrics.LatencyHistory, latency)

	if success {
		suite.metrics.SuccessfulReqs++
	} else {
		suite.metrics.FailedReqs++
	}

	// Keep latency history bounded
	if len(suite.metrics.LatencyHistory) > 1000 {
		suite.metrics.LatencyHistory = suite.metrics.LatencyHistory[len(suite.metrics.LatencyHistory)-1000:]
	}
}

func (suite *PerformanceTestSuite) resetMetrics() {
	suite.metrics.mu.Lock()
	defer suite.metrics.mu.Unlock()

	suite.metrics.TotalRequests = 0
	suite.metrics.SuccessfulReqs = 0
	suite.metrics.FailedReqs = 0
	suite.metrics.TotalLatency = 0
	suite.metrics.LatencyHistory = nil
	suite.metrics.StartTime = time.Now()
}

func (suite *PerformanceTestSuite) getMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024 // Convert to MB
}

func (suite *PerformanceTestSuite) getCPUUsage() float64 {
	// Simplified CPU usage calculation
	// In a real implementation, this would use more sophisticated CPU monitoring
	return float64(runtime.NumGoroutine()) * 0.1 // Mock calculation
}

func (suite *PerformanceTestSuite) monitorResources(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			suite.metrics.mu.Lock()
			suite.metrics.MemoryUsageMB = suite.getMemoryUsage()
			suite.metrics.CPUUsagePercent = suite.getCPUUsage()
			suite.metrics.mu.Unlock()
		}
	}
}

func (suite *PerformanceTestSuite) validateAgainstBaseline(t *testing.T, config LoadTestConfig) {
	suite.metrics.mu.RLock()
	defer suite.metrics.mu.RUnlock()

	// Validate against baseline metrics
	if suite.metrics.P95Latency > suite.baseline.MaxLatencyP95 {
		t.Errorf("P95 latency %v exceeds baseline %v",
			suite.metrics.P95Latency, suite.baseline.MaxLatencyP95)
	}

	if suite.metrics.ThroughputRPS < suite.baseline.MinThroughputRPS {
		t.Errorf("Throughput %.2f RPS below baseline %.2f RPS",
			suite.metrics.ThroughputRPS, suite.baseline.MinThroughputRPS)
	}

	if suite.metrics.ErrorRate > suite.baseline.MaxErrorRate {
		t.Errorf("Error rate %.2f%% exceeds baseline %.2f%%",
			suite.metrics.ErrorRate*100, suite.baseline.MaxErrorRate*100)
	}

	if suite.metrics.MemoryUsageMB > suite.baseline.MaxMemoryUsageMB {
		t.Errorf("Memory usage %.2f MB exceeds baseline %.2f MB",
			suite.metrics.MemoryUsageMB, suite.baseline.MaxMemoryUsageMB)
	}
}

func (suite *PerformanceTestSuite) analyzeScalabilityTrends(t *testing.T, results []ScalabilityResult) {
	if len(results) < 2 {
		return
	}

	t.Log("Scalability Analysis:")

	// Analyze throughput scaling
	for i := 1; i < len(results); i++ {
		prev := results[i-1]
		curr := results[i]

		userRatio := float64(curr.ConcurrentUsers) / float64(prev.ConcurrentUsers)
		throughputRatio := curr.ThroughputRPS / prev.ThroughputRPS

		efficiency := throughputRatio / userRatio
		t.Logf("  Users %d->%d: Throughput ratio %.2f, User ratio %.2f, Efficiency %.2f",
			prev.ConcurrentUsers, curr.ConcurrentUsers, throughputRatio, userRatio, efficiency)

		// Good scalability should have efficiency > 0.5
		if efficiency < 0.3 {
			t.Logf("    WARNING: Poor scaling efficiency (%.2f)", efficiency)
		}
	}

	// Check if latency degrades acceptably
	for _, result := range results {
		if result.P95Latency > 15*time.Second {
			t.Errorf("P95 latency %v too high at %d users", result.P95Latency, result.ConcurrentUsers)
		}
	}
}

func (suite *PerformanceTestSuite) analyzeMemoryTrend(t *testing.T, snapshots []MemorySnapshot) {
	if len(snapshots) < 3 {
		return
	}

	initial := snapshots[0].MemoryMB
	final := snapshots[len(snapshots)-1].MemoryMB
	maxMemory := initial

	// Find peak memory usage
	for _, snapshot := range snapshots {
		if snapshot.MemoryMB > maxMemory {
			maxMemory = snapshot.MemoryMB
		}
	}

	memoryGrowth := final - initial
	peakGrowth := maxMemory - initial

	t.Logf("Memory Analysis:")
	t.Logf("  Initial: %.2f MB", initial)
	t.Logf("  Final: %.2f MB", final)
	t.Logf("  Peak: %.2f MB", maxMemory)
	t.Logf("  Net Growth: %.2f MB", memoryGrowth)
	t.Logf("  Peak Growth: %.2f MB", peakGrowth)

	// Check for memory leaks
	if memoryGrowth > 200 {
		t.Errorf("Potential memory leak detected: %.2f MB growth", memoryGrowth)
	}

	if peakGrowth > 500 {
		t.Errorf("Excessive peak memory usage: %.2f MB above baseline", peakGrowth)
	}
}

func (suite *PerformanceTestSuite) generatePerformanceReport(t *testing.T) {
	suite.metrics.mu.RLock()
	defer suite.metrics.mu.RUnlock()

	t.Log("=== PERFORMANCE TEST REPORT ===")
	t.Logf("Test Duration: %v", suite.metrics.EndTime.Sub(suite.metrics.StartTime))
	t.Logf("Total Requests: %d", suite.metrics.TotalRequests)
	t.Logf("Successful Requests: %d (%.1f%%)",
		suite.metrics.SuccessfulReqs,
		float64(suite.metrics.SuccessfulReqs)/float64(suite.metrics.TotalRequests)*100)
	t.Logf("Failed Requests: %d (%.1f%%)",
		suite.metrics.FailedReqs,
		float64(suite.metrics.FailedReqs)/float64(suite.metrics.TotalRequests)*100)
	t.Logf("Peak Throughput: %.2f RPS", suite.metrics.ThroughputRPS)
	t.Logf("Latency - P50: %v, P95: %v, P99: %v",
		suite.metrics.P50Latency, suite.metrics.P95Latency, suite.metrics.P99Latency)
	t.Logf("Peak Memory Usage: %.2f MB", suite.metrics.MemoryUsageMB)
	t.Logf("Peak CPU Usage: %.2f%%", suite.metrics.CPUUsagePercent)
	t.Log("=== END PERFORMANCE REPORT ===")
}

func (suite *PerformanceTestSuite) cleanup() {
	suite.cancel()
	if suite.cluster != nil {
		suite.cluster.Shutdown()
	}
}

func calculatePercentiles(latencies []time.Duration) (p50, p95, p99 time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}

	// Sort latencies
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)

	for i := 0; i < len(sortedLatencies); i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	// Calculate percentiles
	p50 = sortedLatencies[len(sortedLatencies)*50/100]
	p95 = sortedLatencies[len(sortedLatencies)*95/100]
	p99 = sortedLatencies[len(sortedLatencies)*99/100]

	return p50, p95, p99
}

func generatePayload(size int) string {
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = byte('a' + (i % 26))
	}
	return string(payload)
}

// Benchmark tests

func BenchmarkSingleInference(b *testing.B) {
	suite, err := NewPerformanceTestSuite(1)
	if err != nil {
		b.Fatal(err)
	}
	defer suite.cleanup()

	suite.cluster.Start()
	time.Sleep(5 * time.Second)

	leader := suite.cluster.GetLeader()
	if leader == nil {
		b.Fatal("No leader available")
	}

	suite.registerTestModel(leader)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		requestID := 0
		for pb.Next() {
			req := &api.InferenceRequest{
				Model:  "test-model-1b",
				Prompt: fmt.Sprintf("Benchmark request %d", requestID),
				Options: map[string]interface{}{
					"temperature": 0.1,
					"max_tokens":  30,
				},
			}

			_, err := leader.ProcessInference(context.Background(), req)
			if err != nil {
				b.Errorf("Request failed: %v", err)
			}
			requestID++
		}
	})
}

func BenchmarkConsensusOperations(b *testing.B) {
	suite, err := NewPerformanceTestSuite(3)
	if err != nil {
		b.Fatal(err)
	}
	defer suite.cleanup()

	suite.cluster.Start()
	time.Sleep(5 * time.Second)

	leader := suite.cluster.GetLeader()
	if leader == nil {
		b.Fatal("No leader available")
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		operationID := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", operationID)
			value := fmt.Sprintf("bench_value_%d", operationID)

			err := leader.ApplyConsensusOperation(key, value)
			if err != nil {
				b.Errorf("Consensus operation failed: %v", err)
			}
			operationID++
		}
	})
}
