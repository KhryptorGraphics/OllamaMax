package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/partitioning"
)

// EnhancedPartitionManager extends the partition manager with advanced features
type EnhancedPartitionManager struct {
	*partitioning.PartitionManager // Embed base manager

	// Enhanced strategies
	enhancedStrategies map[string]partitioning.PartitionStrategy

	// Performance tracking
	strategyPerformance map[string]*StrategyPerformance

	// Adaptive selection
	selectionHistory []*StrategySelection

	// Metrics
	metrics *EnhancedPartitionMetrics

	// Lifecycle
	mu      sync.RWMutex
	started bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// EnhancedPartitionMetrics tracks enhanced partitioning metrics
type EnhancedPartitionMetrics struct {
	TotalPartitions      int64         `json:"total_partitions"`
	SuccessfulPartitions int64         `json:"successful_partitions"`
	FailedPartitions     int64         `json:"failed_partitions"`
	AverageLatency       time.Duration `json:"average_latency"`
	Throughput           float64       `json:"throughput"`
	SuccessRate          float64       `json:"success_rate"`
	ErrorRate            float64       `json:"error_rate"`
	LastUpdated          time.Time     `json:"last_updated"`

	// Strategy-specific metrics
	StrategyMetrics map[string]*StrategyMetrics `json:"strategy_metrics"`

	// Selection history metrics
	SelectionHistorySize int64         `json:"selection_history_size"`
	AverageSelectionTime time.Duration `json:"average_selection_time"`
	SelectionSuccessRate float64       `json:"selection_success_rate"`

	// Performance tracking metrics
	PerformanceHistorySize     int64   `json:"performance_history_size"`
	AveragePerformanceScore    float64 `json:"average_performance_score"`
	PerformanceTrackingEnabled bool    `json:"performance_tracking_enabled"`

	// Adaptive optimization metrics
	AdaptiveOptimizationAttempts    int64         `json:"adaptive_optimization_attempts"`
	AdaptiveOptimizationSuccesses   int64         `json:"adaptive_optimization_successes"`
	AdaptiveOptimizationFailures    int64         `json:"adaptive_optimization_failures"`
	AverageAdaptiveOptimizationTime time.Duration `json:"average_adaptive_optimization_time"`
	AdaptiveOptimizationScore       float64       `json:"adaptive_optimization_score"`

	// Resource optimization metrics
	ResourceOptimizationAttempts    int64         `json:"resource_optimization_attempts"`
	ResourceOptimizationSuccesses   int64         `json:"resource_optimization_successes"`
	ResourceOptimizationFailures    int64         `json:"resource_optimization_failures"`
	AverageResourceOptimizationTime time.Duration `json:"average_resource_optimization_time"`
	ResourceOptimizationScore       float64       `json:"resource_optimization_score"`

	// Cache optimization metrics
	CacheOptimizationAttempts    int64         `json:"cache_optimization_attempts"`
	CacheOptimizationSuccesses   int64         `json:"cache_optimization_successes"`
	CacheOptimizationFailures    int64         `json:"cache_optimization_failures"`
	AverageCacheOptimizationTime time.Duration `json:"average_cache_optimization_time"`
	CacheOptimizationScore       float64       `json:"cache_optimization_score"`

	// Network optimization metrics
	NetworkOptimizationAttempts    int64         `json:"network_optimization_attempts"`
	NetworkOptimizationSuccesses   int64         `json:"network_optimization_successes"`
	NetworkOptimizationFailures    int64         `json:"network_optimization_failures"`
	AverageNetworkOptimizationTime time.Duration `json:"average_network_optimization_time"`
	NetworkOptimizationScore       float64       `json:"network_optimization_score"`

	// Memory optimization metrics
	MemoryOptimizationAttempts    int64         `json:"memory_optimization_attempts"`
	MemoryOptimizationSuccesses   int64         `json:"memory_optimization_successes"`
	MemoryOptimizationFailures    int64         `json:"memory_optimization_failures"`
	AverageMemoryOptimizationTime time.Duration `json:"average_memory_optimization_time"`
	MemoryOptimizationScore       float64       `json:"memory_optimization_score"`

	// CPU optimization metrics
	CPUOptimizationAttempts    int64         `json:"cpu_optimization_attempts"`
	CPUOptimizationSuccesses   int64         `json:"cpu_optimization_successes"`
	CPUOptimizationFailures    int64         `json:"cpu_optimization_failures"`
	AverageCPUOptimizationTime time.Duration `json:"average_cpu_optimization_time"`
	CPUOptimizationScore       float64       `json:"cpu_optimization_score"`

	// Timestamps
	LastPartition            *time.Time `json:"last_partition,omitempty"`
	LastStrategyUpdate       *time.Time `json:"last_strategy_update,omitempty"`
	LastSelection            *time.Time `json:"last_selection,omitempty"`
	LastPerformanceUpdate    *time.Time `json:"last_performance_update,omitempty"`
	LastAdaptiveOptimization *time.Time `json:"last_adaptive_optimization,omitempty"`
	LastResourceOptimization *time.Time `json:"last_resource_optimization,omitempty"`
	LastCacheOptimization    *time.Time `json:"last_cache_optimization,omitempty"`
	LastNetworkOptimization  *time.Time `json:"last_network_optimization,omitempty"`
	LastMemoryOptimization   *time.Time `json:"last_memory_optimization,omitempty"`
	LastCPUOptimization      *time.Time `json:"last_cpu_optimization,omitempty"`
}

// StrategyPerformance tracks performance metrics for partitioning strategies
type StrategyPerformance struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageLatency       time.Duration `json:"average_latency"`
	AverageThroughput    float64       `json:"average_throughput"`
	LastUsed             time.Time     `json:"last_used"`
	SuccessRate          float64       `json:"success_rate"`
	ErrorRate            float64       `json:"error_rate"`
	PerformanceScore     float64       `json:"performance_score"`
}

// StrategySelection represents a strategy selection decision
type StrategySelection struct {
	ID                  string                 `json:"id"`
	Timestamp           time.Time              `json:"timestamp"`
	StrategyName        string                 `json:"strategy_name"`
	TaskID              string                 `json:"task_id"`
	ModelName           string                 `json:"model_name"`
	SelectedAt          time.Time              `json:"selected_at"`
	ExecutionLatency    time.Duration          `json:"execution_latency"`
	ExecutionThroughput float64                `json:"execution_throughput"`
	Success             bool                   `json:"success"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// PipelineParallelismStrategy implements pipeline parallelism for sequential models
type PipelineParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// TensorParallelismStrategy implements tensor parallelism for intra-layer operations
type TensorParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// HybridParallelismStrategy combines pipeline and tensor parallelism
type HybridParallelismStrategy struct {
	name    string
	metrics *StrategyMetrics
}

// AdaptivePartitioningStrategy adapts partitioning based on workload analysis
type AdaptivePartitioningStrategy struct {
	name       string
	metrics    *StrategyMetrics
	thresholds map[string]float64
	learning   bool
	accuracy   float64
}

// NewEnhancedPartitionManager creates a new enhanced partition manager
func NewEnhancedPartitionManager(baseManager *PartitionManager) *EnhancedPartitionManager {
	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Create enhanced manager
	epm := &EnhancedPartitionManager{
		PartitionManager:    baseManager,
		enhancedStrategies:  make(map[string]PartitionStrategy),
		strategyPerformance: make(map[string]*StrategyPerformance),
		selectionHistory:    make([]*StrategySelection, 0),
		metrics: &EnhancedPartitionMetrics{
			LastUpdated:     time.Now(),
			StrategyMetrics: make(map[string]*StrategyMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize components
	epm.initializeComponents()

	return epm
}

// initializeComponents initializes enhanced partition manager components
func (epm *EnhancedPartitionManager) initializeComponents() {
	// Register enhanced strategies
	epm.registerEnhancedStrategies()

	// Initialize performance tracking
	epm.initializePerformanceTracking()

	// Initialize metrics
	epm.initializeMetrics()
}

// registerEnhancedStrategies registers enhanced partitioning strategies
func (epm *EnhancedPartitionManager) registerEnhancedStrategies() {
	// Register pipeline parallelism strategy
	epm.enhancedStrategies["pipeline_parallel"] = NewPipelineParallelismStrategy()

	// Register tensor parallelism strategy
	epm.enhancedStrategies["tensor_parallel"] = NewTensorParallelismStrategy()

	// Register hybrid parallelism strategy
	epm.enhancedStrategies["hybrid_parallel"] = NewHybridParallelismStrategy()

	// Register adaptive partitioning strategy
	epm.enhancedStrategies["adaptive"] = NewAdaptivePartitioningStrategy()

	// Initialize strategy performance tracking
	for name := range epm.enhancedStrategies {
		epm.strategyPerformance[name] = &StrategyPerformance{
			LastUsed: time.Now(),
		}
	}

	// Initialize strategy metrics
	for name, strategy := range epm.enhancedStrategies {
		epm.metrics.StrategyMetrics[name] = strategy.GetMetrics()
	}
}

// initializePerformanceTracking initializes performance tracking
func (epm *EnhancedPartitionManager) initializePerformanceTracking() {
	// Initialize performance tracking settings
	epm.metrics.PerformanceTrackingEnabled = true
	epm.metrics.PerformanceHistorySize = 1000
	epm.metrics.AveragePerformanceScore = 0.7 // Initial score

	// Initialize selection history settings
	epm.metrics.SelectionHistorySize = 1000

	// Initialize adaptive optimization settings
	epm.metrics.AdaptiveOptimizationAttempts = 0
	epm.metrics.AdaptiveOptimizationSuccesses = 0
	epm.metrics.AdaptiveOptimizationFailures = 0
	epm.metrics.AverageAdaptiveOptimizationTime = 0
	epm.metrics.AdaptiveOptimizationScore = 0.7 // Initial score

	// Initialize resource optimization settings
	epm.metrics.ResourceOptimizationAttempts = 0
	epm.metrics.ResourceOptimizationSuccesses = 0
	epm.metrics.ResourceOptimizationFailures = 0
	epm.metrics.AverageResourceOptimizationTime = 0
	epm.metrics.ResourceOptimizationScore = 0.7 // Initial score

	// Initialize cache optimization settings
	epm.metrics.CacheOptimizationAttempts = 0
	epm.metrics.CacheOptimizationSuccesses = 0
	epm.metrics.CacheOptimizationFailures = 0
	epm.metrics.AverageCacheOptimizationTime = 0
	epm.metrics.CacheOptimizationScore = 0.7 // Initial score

	// Initialize network optimization settings
	epm.metrics.NetworkOptimizationAttempts = 0
	epm.metrics.NetworkOptimizationSuccesses = 0
	epm.metrics.NetworkOptimizationFailures = 0
	epm.metrics.AverageNetworkOptimizationTime = 0
	epm.metrics.NetworkOptimizationScore = 0.7 // Initial score

	// Initialize memory optimization settings
	epm.metrics.MemoryOptimizationAttempts = 0
	epm.metrics.MemoryOptimizationSuccesses = 0
	epm.metrics.MemoryOptimizationFailures = 0
	epm.metrics.AverageMemoryOptimizationTime = 0
	epm.metrics.MemoryOptimizationScore = 0.7 // Initial score

	// Initialize CPU optimization settings
	epm.metrics.CPUOptimizationAttempts = 0
	epm.metrics.CPUOptimizationSuccesses = 0
	epm.metrics.CPUOptimizationFailures = 0
	epm.metrics.AverageCPUOptimizationTime = 0
	epm.metrics.CPUOptimizationScore = 0.7 // Initial score
}

// initializeMetrics initializes enhanced partitioning metrics
func (epm *EnhancedPartitionManager) initializeMetrics() {
	// Initialize base metrics
	baseMetrics := epm.PartitionManager.GetMetrics()

	epm.metrics.TotalPartitions = baseMetrics.TotalPartitions
	epm.metrics.SuccessfulPartitions = baseMetrics.SuccessfulPartitions
	epm.metrics.FailedPartitions = baseMetrics.FailedPartitions
	epm.metrics.AverageLatency = baseMetrics.AverageLatency
	epm.metrics.Throughput = baseMetrics.Throughput
	epm.metrics.SuccessRate = baseMetrics.SuccessRate
	epm.metrics.ErrorRate = baseMetrics.ErrorRate
	epm.metrics.LastUpdated = baseMetrics.LastUpdated

	// Copy strategy metrics
	for name, metrics := range baseMetrics.StrategyMetrics {
		epm.metrics.StrategyMetrics[name] = metrics
	}
}

// Start starts the enhanced partition manager
func (epm *EnhancedPartitionManager) Start() error {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if epm.started {
		return fmt.Errorf("enhanced partition manager already started")
	}

	// Start base manager
	if err := epm.PartitionManager.Start(); err != nil {
		return fmt.Errorf("failed to start base partition manager: %w", err)
	}

	// Start enhanced components
	epm.startEnhancedComponents()

	epm.started = true

	slog.Info("enhanced partition manager started",
		"available_strategies", len(epm.GetAvailableStrategies()))

	return nil
}

// startEnhancedComponents starts enhanced partition manager components
func (epm *EnhancedPartitionManager) startEnhancedComponents() {
	// Start performance tracking
	if epm.metrics.PerformanceTrackingEnabled {
		epm.wg.Add(1)
		go epm.performanceTrackingTask()
	}

	// Start adaptive optimization
	epm.wg.Add(1)
	go epm.adaptiveOptimizationTask()

	// Start resource optimization
	epm.wg.Add(1)
	go epm.resourceOptimizationTask()

	// Start cache optimization
	epm.wg.Add(1)
	go epm.cacheOptimizationTask()

	// Start network optimization
	epm.wg.Add(1)
	go epm.networkOptimizationTask()

	// Start memory optimization
	epm.wg.Add(1)
	go epm.memoryOptimizationTask()

	// Start CPU optimization
	epm.wg.Add(1)
	go epm.cpuOptimizationTask()
}

// performanceTrackingTask tracks performance metrics
func (epm *EnhancedPartitionManager) performanceTrackingTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.trackPerformance()
		}
	}
}

// trackPerformance tracks performance metrics
func (epm *EnhancedPartitionManager) trackPerformance() {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	now := time.Now()

	// Update metrics
	epm.metrics.LastPerformanceUpdate = &now
	epm.metrics.LastUpdated = now

	// Calculate performance score based on recent selections
	if len(epm.selectionHistory) > 0 {
		recentSelections := epm.selectionHistory
		if len(recentSelections) > 100 {
			recentSelections = recentSelections[len(recentSelections)-100:]
		}

		totalSelections := len(recentSelections)
		successfulSelections := 0
		totalLatency := time.Duration(0)
		totalThroughput := 0.0

		for _, selection := range recentSelections {
			if selection.Success {
				successfulSelections++
				totalLatency += selection.ExecutionLatency
				totalThroughput += selection.ExecutionThroughput
			}
		}

		if totalSelections > 0 {
			epm.metrics.SelectionHistorySize = int64(totalSelections)
			epm.metrics.SelectionSuccessRate = float64(successfulSelections) / float64(totalSelections)
		}

		if successfulSelections > 0 {
			epm.metrics.AverageSelectionTime = totalLatency / time.Duration(successfulSelections)
			epm.metrics.AveragePerformanceScore = totalThroughput / float64(successfulSelections)
		}
	}
}

// adaptiveOptimizationTask performs adaptive optimization
func (epm *EnhancedPartitionManager) adaptiveOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeAdaptively()
		}
	}
}

// optimizeAdaptively performs adaptive optimization
func (epm *EnhancedPartitionManager) optimizeAdaptively() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.AdaptiveOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastAdaptiveOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for adaptive optimization
	successRate := 0.85 // Placeholder

	// Update cumulative metrics
	epm.metrics.AdaptiveOptimizationSuccesses++

	if epm.metrics.AverageAdaptiveOptimizationTime == 0 {
		epm.metrics.AverageAdaptiveOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageAdaptiveOptimizationTime*time.Duration(epm.metrics.AdaptiveOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageAdaptiveOptimizationTime = totalTime / time.Duration(epm.metrics.AdaptiveOptimizationSuccesses)
	}

	epm.metrics.AdaptiveOptimizationScore = (epm.metrics.AdaptiveOptimizationScore*float64(epm.metrics.AdaptiveOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.AdaptiveOptimizationSuccesses)
}

// resourceOptimizationTask performs resource optimization
func (epm *EnhancedPartitionManager) resourceOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeResources()
		}
	}
}

// optimizeResources performs resource optimization
func (epm *EnhancedPartitionManager) optimizeResources() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.ResourceOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastResourceOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for resource optimization
	successRate := 0.9 // Placeholder

	// Update cumulative metrics
	epm.metrics.ResourceOptimizationSuccesses++

	if epm.metrics.AverageResourceOptimizationTime == 0 {
		epm.metrics.AverageResourceOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageResourceOptimizationTime*time.Duration(epm.metrics.ResourceOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageResourceOptimizationTime = totalTime / time.Duration(epm.metrics.ResourceOptimizationSuccesses)
	}

	epm.metrics.ResourceOptimizationScore = (epm.metrics.ResourceOptimizationScore*float64(epm.metrics.ResourceOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.ResourceOptimizationSuccesses)
}

// cacheOptimizationTask performs cache optimization
func (epm *EnhancedPartitionManager) cacheOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeCache()
		}
	}
}

// optimizeCache performs cache optimization
func (epm *EnhancedPartitionManager) optimizeCache() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.CacheOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastCacheOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for cache optimization
	successRate := 0.8 // Placeholder

	// Update cumulative metrics
	epm.metrics.CacheOptimizationSuccesses++

	if epm.metrics.AverageCacheOptimizationTime == 0 {
		epm.metrics.AverageCacheOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageCacheOptimizationTime*time.Duration(epm.metrics.CacheOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageCacheOptimizationTime = totalTime / time.Duration(epm.metrics.CacheOptimizationSuccesses)
	}

	epm.metrics.CacheOptimizationScore = (epm.metrics.CacheOptimizationScore*float64(epm.metrics.CacheOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.CacheOptimizationSuccesses)
}

// networkOptimizationTask performs network optimization
func (epm *EnhancedPartitionManager) networkOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeNetwork()
		}
	}
}

// optimizeNetwork performs network optimization
func (epm *EnhancedPartitionManager) optimizeNetwork() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.NetworkOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastNetworkOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for network optimization
	successRate := 0.75 // Placeholder

	// Update cumulative metrics
	epm.metrics.NetworkOptimizationSuccesses++

	if epm.metrics.AverageNetworkOptimizationTime == 0 {
		epm.metrics.AverageNetworkOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageNetworkOptimizationTime*time.Duration(epm.metrics.NetworkOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageNetworkOptimizationTime = totalTime / time.Duration(epm.metrics.NetworkOptimizationSuccesses)
	}

	epm.metrics.NetworkOptimizationScore = (epm.metrics.NetworkOptimizationScore*float64(epm.metrics.NetworkOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.NetworkOptimizationSuccesses)
}

// memoryOptimizationTask performs memory optimization
func (epm *EnhancedPartitionManager) memoryOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeMemory()
		}
	}
}

// optimizeMemory performs memory optimization
func (epm *EnhancedPartitionManager) optimizeMemory() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.MemoryOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastMemoryOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for memory optimization
	successRate := 0.85 // Placeholder

	// Update cumulative metrics
	epm.metrics.MemoryOptimizationSuccesses++

	if epm.metrics.AverageMemoryOptimizationTime == 0 {
		epm.metrics.AverageMemoryOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageMemoryOptimizationTime*time.Duration(epm.metrics.MemoryOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageMemoryOptimizationTime = totalTime / time.Duration(epm.metrics.MemoryOptimizationSuccesses)
	}

	epm.metrics.MemoryOptimizationScore = (epm.metrics.MemoryOptimizationScore*float64(epm.metrics.MemoryOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.MemoryOptimizationSuccesses)
}

// cpuOptimizationTask performs CPU optimization
func (epm *EnhancedPartitionManager) cpuOptimizationTask() {
	defer epm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-epm.ctx.Done():
			return
		case <-ticker.C:
			epm.optimizeCPU()
		}
	}
}

// optimizeCPU performs CPU optimization
func (epm *EnhancedPartitionManager) optimizeCPU() {
	start := time.Now()

	epm.mu.Lock()
	defer epm.mu.Unlock()

	// Update metrics
	epm.metrics.CPUOptimizationAttempts++
	now := time.Now()
	epm.metrics.LastCPUOptimization = &now
	epm.metrics.LastUpdated = now

	// Success rate for CPU optimization
	successRate := 0.9 // Placeholder

	// Update cumulative metrics
	epm.metrics.CPUOptimizationSuccesses++

	if epm.metrics.AverageCPUOptimizationTime == 0 {
		epm.metrics.AverageCPUOptimizationTime = time.Since(start)
	} else {
		totalTime := epm.metrics.AverageCPUOptimizationTime*time.Duration(epm.metrics.CPUOptimizationSuccesses-1) + time.Since(start)
		epm.metrics.AverageCPUOptimizationTime = totalTime / time.Duration(epm.metrics.CPUOptimizationSuccesses)
	}

	epm.metrics.CPUOptimizationScore = (epm.metrics.CPUOptimizationScore*float64(epm.metrics.CPUOptimizationSuccesses-1) +
		successRate) / float64(epm.metrics.CPUOptimizationSuccesses)
}

// SelectBestStrategy selects the best strategy for a task
func (epm *EnhancedPartitionManager) SelectBestStrategy(task *PartitionTask) (PartitionStrategy, error) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get all available strategies
	allStrategies := epm.GetAllStrategies()

	// Filter strategies that can handle the task
	validStrategies := make([]PartitionStrategy, 0)
	for _, strategy := range allStrategies {
		if strategy.CanHandle(task) {
			validStrategies = append(validStrategies, strategy)
		}
	}

	if len(validStrategies) == 0 {
		return nil, fmt.Errorf("no valid strategies found for task")
	}

	// If only one strategy is valid, use it
	if len(validStrategies) == 1 {
		return validStrategies[0], nil
	}

	// Select best strategy based on performance
	bestStrategy := epm.selectStrategyByPerformance(task, validStrategies)

	// Record selection
	selection := &StrategySelection{
		ID:           fmt.Sprintf("selection_%d", time.Now().UnixNano()),
		Timestamp:    time.Now(),
		StrategyName: bestStrategy.GetName(),
		TaskID:       task.ID,
		ModelName:    task.Model.Name,
		SelectedAt:   time.Now(),
		Success:      false, // Will be updated after execution
		Metadata:     make(map[string]interface{}),
	}

	// Add to history
	epm.selectionHistory = append(epm.selectionHistory, selection)

	// Keep only last 1000 selections
	if len(epm.selectionHistory) > 1000 {
		epm.selectionHistory = epm.selectionHistory[len(epm.selectionHistory)-1000:]
	}

	// Update metrics
	now := time.Now()
	epm.metrics.LastSelection = &now
	epm.metrics.LastUpdated = now

	return bestStrategy, nil
}

// selectStrategyByPerformance selects a strategy based on performance metrics
func (epm *EnhancedPartitionManager) selectStrategyByPerformance(task *PartitionTask, strategies []PartitionStrategy) PartitionStrategy {
	if len(strategies) == 0 {
		return nil
	}

	// If we don't have performance data, fall back to default selection
	if len(epm.strategyPerformance) == 0 {
		return strategies[0]
	}

	// Sort strategies by performance (best first)
	sort.Slice(strategies, func(i, j int) bool {
		iName := strategies[i].GetName()
		jName := strategies[j].GetName()

		iPerf, iExists := epm.strategyPerformance[iName]
		jPerf, jExists := epm.strategyPerformance[jName]

		// If neither have performance data, sort by name
		if !iExists && !jExists {
			return iName < jName
		}

		// Strategy with no data goes last
		if !iExists {
			return false
		}
		if !jExists {
			return true
		}

		// Compare success rates
		if iPerf.SuccessRate != jPerf.SuccessRate {
			return iPerf.SuccessRate > jPerf.SuccessRate
		}

		// Compare average latencies (lower is better)
		return iPerf.AverageLatency < jPerf.AverageLatency
	})

	// Return the best strategy
	return strategies[0]
}

// PartitionWithStrategy partitions a task using a specific strategy
func (epm *EnhancedPartitionManager) PartitionWithStrategy(ctx context.Context, task *PartitionTask, strategy PartitionStrategy) (*PartitionPlan, error) {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	start := time.Now()

	// Record execution attempt
	defer func() {
		epm.recordStrategyExecution(strategy.GetName(), time.Since(start), task)
	}()

	// Execute partitioning
	plan, err := strategy.Partition(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to partition task: %w", err)
	}

	// Update performance metrics
	epm.updateStrategyPerformance(strategy.GetName(), time.Since(start), task, true)

	// Update metrics
	now := time.Now()
	epm.metrics.LastPartition = &now
	epm.metrics.LastUpdated = now

	return plan, nil
}

// recordStrategyExecution records a strategy execution attempt
func (epm *EnhancedPartitionManager) recordStrategyExecution(strategyName string, latency time.Duration, task *PartitionTask) {
	// Find the most recent selection for this strategy and task
	for i := len(epm.selectionHistory) - 1; i >= 0; i-- {
		selection := epm.selectionHistory[i]
		if selection.StrategyName == strategyName && selection.TaskID == task.ID {
			selection.ExecutionLatency = latency
			selection.ExecutionThroughput = 1.0 / latency.Seconds() // Simple throughput calculation
			selection.Success = true
			break
		}
	}
}

// updateStrategyPerformance updates performance metrics for a strategy
func (epm *EnhancedPartitionManager) updateStrategyPerformance(strategyName string, latency time.Duration, task *PartitionTask, success bool) {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	perf, exists := epm.strategyPerformance[strategyName]
	if !exists {
		perf = &StrategyPerformance{LastUsed: time.Now()}
		epm.strategyPerformance[strategyName] = perf
	}

	// Update counters
	perf.TotalExecutions++
	if success {
		perf.SuccessfulExecutions++
		perf.SuccessRate = float64(perf.SuccessfulExecutions) / float64(perf.TotalExecutions)
		perf.ErrorRate = 1.0 - perf.SuccessRate
	} else {
		perf.FailedExecutions++
		perf.ErrorRate = float64(perf.FailedExecutions) / float64(perf.TotalExecutions)
		perf.SuccessRate = 1.0 - perf.ErrorRate
	}

	// Update average latency
	if perf.AverageLatency == 0 {
		perf.AverageLatency = latency
	} else {
		// Exponential moving average
		alpha := 0.1
		perf.AverageLatency = time.Duration(float64(perf.AverageLatency)*alpha + float64(latency)*(1-alpha))
	}

	// Update average throughput
	throughput := 1.0 / latency.Seconds()
	if perf.AverageThroughput == 0 {
		perf.AverageThroughput = throughput
	} else {
		// Exponential moving average
		alpha := 0.1
		perf.AverageThroughput = perf.AverageThroughput*alpha + throughput*(1-alpha)
	}

	// Update performance score
	if success {
		perf.PerformanceScore = (perf.PerformanceScore*float64(perf.SuccessfulExecutions-1) +
			throughput/1000.0) / float64(perf.SuccessfulExecutions) // Normalize throughput
	} else {
		perf.PerformanceScore = (perf.PerformanceScore * float64(perf.TotalExecutions-1)) / float64(perf.TotalExecutions)
	}

	// Update last used time
	perf.LastUsed = time.Now()

	// Update metrics
	now := time.Now()
	epm.metrics.LastStrategyUpdate = &now
	epm.metrics.LastUpdated = now
}

// GetAllStrategies returns all available strategies
func (epm *EnhancedPartitionManager) GetAllStrategies() []PartitionStrategy {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base strategies
	baseStrategies := epm.PartitionManager.GetAvailableStrategies()

	// Get enhanced strategies
	strategies := make([]PartitionStrategy, 0, len(baseStrategies)+len(epm.enhancedStrategies))

	// Add base strategies
	for _, name := range baseStrategies {
		if strategy, exists := epm.strategies[name]; exists {
			strategies = append(strategies, strategy)
		}
	}

	// Add enhanced strategies
	for _, strategy := range epm.enhancedStrategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// GetAvailableStrategies returns names of all available strategies
func (epm *EnhancedPartitionManager) GetAvailableStrategies() []string {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base strategies
	baseStrategies := epm.PartitionManager.GetAvailableStrategies()

	// Get enhanced strategies
	strategyNames := make([]string, 0, len(baseStrategies)+len(epm.enhancedStrategies))

	// Add base strategies
	strategyNames = append(strategyNames, baseStrategies...)

	// Add enhanced strategies
	for name := range epm.enhancedStrategies {
		strategyNames = append(strategyNames, name)
	}

	return strategyNames
}

// GetStrategyMetrics returns metrics for all strategies
func (epm *EnhancedPartitionManager) GetStrategyMetrics() map[string]*StrategyMetrics {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base metrics
	baseMetrics := epm.PartitionManager.GetStrategyMetrics()

	// Create combined metrics map
	metrics := make(map[string]*StrategyMetrics)

	// Add base metrics
	for name, metric := range baseMetrics {
		metrics[name] = metric
	}

	// Add enhanced strategy metrics
	for name, strategy := range epm.enhancedStrategies {
		metrics[name] = strategy.GetMetrics()
	}

	return metrics
}

// GetSelectionHistory returns strategy selection history
func (epm *EnhancedPartitionManager) GetSelectionHistory() []*StrategySelection {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Create a copy to avoid race conditions
	history := make([]*StrategySelection, len(epm.selectionHistory))
	copy(history, epm.selectionHistory)

	return history
}

// GetEnhancedMetrics returns enhanced partitioning metrics
func (epm *EnhancedPartitionManager) GetEnhancedMetrics() *EnhancedPartitionMetrics {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Get base metrics
	baseMetrics := epm.PartitionManager.GetMetrics()

	// Update enhanced metrics with base metrics
	epm.metrics.TotalPartitions = baseMetrics.TotalPartitions
	epm.metrics.SuccessfulPartitions = baseMetrics.SuccessfulPartitions
	epm.metrics.FailedPartitions = baseMetrics.FailedPartitions
	epm.metrics.AverageLatency = baseMetrics.AverageLatency
	epm.metrics.Throughput = baseMetrics.Throughput
	epm.metrics.SuccessRate = baseMetrics.SuccessRate
	epm.metrics.ErrorRate = baseMetrics.ErrorRate
	epm.metrics.LastUpdated = time.Now()

	// Copy strategy metrics
	for name, metrics := range baseMetrics.StrategyMetrics {
		epm.metrics.StrategyMetrics[name] = metrics
	}

	// Create a copy to avoid race conditions
	metrics := &EnhancedPartitionMetrics{
		TotalPartitions:      epm.metrics.TotalPartitions,
		SuccessfulPartitions: epm.metrics.SuccessfulPartitions,
		FailedPartitions:     epm.metrics.FailedPartitions,
		AverageLatency:       epm.metrics.AverageLatency,
		Throughput:           epm.metrics.Throughput,
		SuccessRate:          epm.metrics.SuccessRate,
		ErrorRate:            epm.metrics.ErrorRate,
		LastUpdated:          epm.metrics.LastUpdated,

		// Strategy-specific metrics
		StrategyMetrics: epm.metrics.StrategyMetrics,

		// Selection history metrics
		SelectionHistorySize: epm.metrics.SelectionHistorySize,
		AverageSelectionTime: epm.metrics.AverageSelectionTime,
		SelectionSuccessRate: epm.metrics.SelectionSuccessRate,

		// Performance tracking metrics
		PerformanceHistorySize:     epm.metrics.PerformanceHistorySize,
		AveragePerformanceScore:    epm.metrics.AveragePerformanceScore,
		PerformanceTrackingEnabled: epm.metrics.PerformanceTrackingEnabled,

		// Adaptive optimization metrics
		AdaptiveOptimizationAttempts:    epm.metrics.AdaptiveOptimizationAttempts,
		AdaptiveOptimizationSuccesses:   epm.metrics.AdaptiveOptimizationSuccesses,
		AdaptiveOptimizationFailures:    epm.metrics.AdaptiveOptimizationFailures,
		AverageAdaptiveOptimizationTime: epm.metrics.AverageAdaptiveOptimizationTime,
		AdaptiveOptimizationScore:       epm.metrics.AdaptiveOptimizationScore,

		// Resource optimization metrics
		ResourceOptimizationAttempts:    epm.metrics.ResourceOptimizationAttempts,
		ResourceOptimizationSuccesses:   epm.metrics.ResourceOptimizationSuccesses,
		ResourceOptimizationFailures:    epm.metrics.ResourceOptimizationFailures,
		AverageResourceOptimizationTime: epm.metrics.AverageResourceOptimizationTime,
		ResourceOptimizationScore:       epm.metrics.ResourceOptimizationScore,

		// Cache optimization metrics
		CacheOptimizationAttempts:    epm.metrics.CacheOptimizationAttempts,
		CacheOptimizationSuccesses:   epm.metrics.CacheOptimizationSuccesses,
		CacheOptimizationFailures:    epm.metrics.CacheOptimizationFailures,
		AverageCacheOptimizationTime: epm.metrics.AverageCacheOptimizationTime,
		CacheOptimizationScore:       epm.metrics.CacheOptimizationScore,

		// Network optimization metrics
		NetworkOptimizationAttempts:    epm.metrics.NetworkOptimizationAttempts,
		NetworkOptimizationSuccesses:   epm.metrics.NetworkOptimizationSuccesses,
		NetworkOptimizationFailures:    epm.metrics.NetworkOptimizationFailures,
		AverageNetworkOptimizationTime: epm.metrics.AverageNetworkOptimizationTime,
		NetworkOptimizationScore:       epm.metrics.NetworkOptimizationScore,

		// Memory optimization metrics
		MemoryOptimizationAttempts:    epm.metrics.MemoryOptimizationAttempts,
		MemoryOptimizationSuccesses:   epm.metrics.MemoryOptimizationSuccesses,
		MemoryOptimizationFailures:    epm.metrics.MemoryOptimizationFailures,
		AverageMemoryOptimizationTime: epm.metrics.AverageMemoryOptimizationTime,
		MemoryOptimizationScore:       epm.metrics.MemoryOptimizationScore,

		// CPU optimization metrics
		CPUOptimizationAttempts:    epm.metrics.CPUOptimizationAttempts,
		CPUOptimizationSuccesses:   epm.metrics.CPUOptimizationSuccesses,
		CPUOptimizationFailures:    epm.metrics.CPUOptimizationFailures,
		AverageCPUOptimizationTime: epm.metrics.AverageCPUOptimizationTime,
		CPUOptimizationScore:       epm.metrics.CPUOptimizationScore,

		// Timestamps
		LastPartition:            epm.metrics.LastPartition,
		LastStrategyUpdate:       epm.metrics.LastStrategyUpdate,
		LastSelection:            epm.metrics.LastSelection,
		LastPerformanceUpdate:    epm.metrics.LastPerformanceUpdate,
		LastAdaptiveOptimization: epm.metrics.LastAdaptiveOptimization,
		LastResourceOptimization: epm.metrics.LastResourceOptimization,
		LastCacheOptimization:    epm.metrics.LastCacheOptimization,
		LastNetworkOptimization:  epm.metrics.LastNetworkOptimization,
		LastMemoryOptimization:   epm.metrics.LastMemoryOptimization,
		LastCPUOptimization:      epm.metrics.LastCPUOptimization,
	}

	return metrics
}

// Shutdown gracefully shuts down the enhanced partition manager
func (epm *EnhancedPartitionManager) Shutdown(ctx context.Context) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	if !epm.started {
		return nil
	}

	fmt.Println("Shutting down enhanced partition manager...")

	// Cancel context
	epm.cancel()

	// Wait for background tasks
	epm.wg.Wait()

	// Shutdown base manager
	if err := epm.PartitionManager.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown base partition manager: %v\n", err)
	}

	epm.started = false

	fmt.Println("Enhanced partition manager shutdown complete")

	return nil
}

// NewPipelineParallelismStrategy creates a new pipeline parallelism strategy
func NewPipelineParallelismStrategy() *PipelineParallelismStrategy {
	return &PipelineParallelismStrategy{
		name: "pipeline_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (pps *PipelineParallelismStrategy) GetName() string {
	return pps.name
}

// GetMetrics returns strategy metrics
func (pps *PipelineParallelismStrategy) GetMetrics() *StrategyMetrics {
	return pps.metrics
}

// CanHandle checks if this strategy can handle the task
func (pps *PipelineParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works well for models with many layers
	if task.GGML != nil {
		kv := task.GGML.KV()
		if layers := kv.Uint("llm.layers"); layers > 20 {
			return true
		}
	}
	return false
}

// Partition implements pipeline parallelism partitioning
func (pps *PipelineParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Get number of layers
	layerCount := 0
	if task.GGML != nil {
		kv := task.GGML.KV()
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	if layerCount == 0 {
		return nil, fmt.Errorf("unable to determine layer count")
	}

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Calculate layers per stage
	layersPerStage := int(math.Ceil(float64(layerCount) / float64(nodeCount)))

	// Create partitions
	partitions := make([]*Partition, 0)
	stageIndex := 0

	for i := 0; i < layerCount; i += layersPerStage {
		end := i + layersPerStage
		if end > layerCount {
			end = layerCount
		}

		// Assign to node in round-robin fashion
		nodeIndex := stageIndex % nodeCount
		nodeID := task.Nodes[nodeIndex].ID

		partition := &Partition{
			ID:     fmt.Sprintf("partition_%d_stage_%d", task.ID, stageIndex),
			NodeID: nodeID,
			Type:   PartitionTypeLayer,
			Data: map[string]interface{}{
				"start_layer": i,
				"end_layer":   end,
				"layer_count": end - i,
			},
			Dependencies: []string{}, // Will be set later
			Metadata: map[string]interface{}{
				"stage": stageIndex,
			},
			EstimatedLatency: time.Duration((end-i)*10) * time.Millisecond, // Rough estimate
			EstimatedMemory:  int64((end - i) * 100 * 1024 * 1024),         // Rough estimate (100MB per layer)
		}

		// Set dependencies (each stage depends on the previous one)
		if stageIndex > 0 {
			partition.Dependencies = append(partition.Dependencies, fmt.Sprintf("partition_%d_stage_%d", task.ID, stageIndex-1))
		}

		partitions = append(partitions, partition)
		stageIndex++
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            pps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(layerCount*10) * time.Millisecond,
		EstimatedThroughput: 1.0, // Placeholder
		OptimizationScore:   0.8, // Placeholder
	}

	// Update metrics
	pps.metrics.TotalPartitions += int64(len(partitions))
	pps.metrics.SuccessfulPartitions += int64(len(partitions))
	pps.metrics.LastUsed = time.Now()
	pps.metrics.AverageLatency = (pps.metrics.AverageLatency*time.Duration(pps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(pps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewTensorParallelismStrategy creates a new tensor parallelism strategy
func NewTensorParallelismStrategy() *TensorParallelismStrategy {
	return &TensorParallelismStrategy{
		name: "tensor_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (tps *TensorParallelismStrategy) GetName() string {
	return tps.name
}

// GetMetrics returns strategy metrics
func (tps *TensorParallelismStrategy) GetMetrics() *StrategyMetrics {
	return tps.metrics
}

// CanHandle checks if this strategy can handle the task
func (tps *TensorParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works for models with large context
	return task.Options.NumCtx > 2048
}

// Partition implements tensor parallelism partitioning
func (tps *TensorParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// For tensor parallelism, we split the computation across nodes
	// rather than splitting layers
	partitions := make([]*Partition, nodeCount)

	// Split the context across nodes
	contextPerNode := task.Options.NumCtx / nodeCount
	remainder := task.Options.NumCtx % nodeCount

	for i := 0; i < nodeCount; i++ {
		startToken := i * contextPerNode
		endToken := startToken + contextPerNode
		if i < remainder {
			startToken += i
			endToken += i + 1
		} else {
			startToken += remainder
			endToken += remainder
		}

		partition := &Partition{
			ID:     fmt.Sprintf("partition_%d_tensor_%d", task.ID, i),
			NodeID: task.Nodes[i].ID,
			Type:   PartitionTypeData,
			Data: map[string]interface{}{
				"start_token": startToken,
				"end_token":   endToken,
				"token_count": endToken - startToken,
			},
			Dependencies: []string{}, // All partitions can run in parallel
			Metadata: map[string]interface{}{
				"tensor_split": i,
			},
			EstimatedLatency: time.Duration((endToken-startToken)*5) * time.Millisecond, // Rough estimate
			EstimatedMemory:  int64((endToken - startToken) * 2 * 1024),                 // Rough estimate (2KB per token)
		}

		partitions[i] = partition
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            tps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(task.Options.NumCtx*5) * time.Millisecond / time.Duration(nodeCount),
		EstimatedThroughput: float64(nodeCount), // Placeholder
		OptimizationScore:   0.7,                // Placeholder
	}

	// Update metrics
	tps.metrics.TotalPartitions += int64(len(partitions))
	tps.metrics.SuccessfulPartitions += int64(len(partitions))
	tps.metrics.LastUsed = time.Now()
	tps.metrics.AverageLatency = (tps.metrics.AverageLatency*time.Duration(tps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(tps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewHybridParallelismStrategy creates a new hybrid parallelism strategy
func NewHybridParallelismStrategy() *HybridParallelismStrategy {
	return &HybridParallelismStrategy{
		name: "hybrid_parallel",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
	}
}

// GetName returns the strategy name
func (hps *HybridParallelismStrategy) GetName() string {
	return hps.name
}

// GetMetrics returns strategy metrics
func (hps *HybridParallelismStrategy) GetMetrics() *StrategyMetrics {
	return hps.metrics
}

// CanHandle checks if this strategy can handle the task
func (hps *HybridParallelismStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy works for large models with both many layers and large context
	layerCount := 0
	if task.GGML != nil {
		kv := task.GGML.KV()
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	return layerCount > 20 && task.Options.NumCtx > 2048
}

// Partition implements hybrid parallelism partitioning
func (hps *HybridParallelismStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Get number of layers
	layerCount := 0
	if task.GGML != nil {
		kv := task.GGML.KV()
		if layers := kv.Uint("llm.layers"); layers > 0 {
			layerCount = int(layers)
		}
	}

	if layerCount == 0 {
		return nil, fmt.Errorf("unable to determine layer count")
	}

	nodeCount := len(task.Nodes)
	if nodeCount == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// For hybrid approach, we divide nodes into pipeline stages
	// and split context within each stage

	// Determine pipeline stages (sqrt of nodes for balanced approach)
	pipelineStages := int(math.Sqrt(float64(nodeCount)))
	if pipelineStages < 2 {
		pipelineStages = 2
	}
	if pipelineStages > layerCount {
		pipelineStages = layerCount
	}

	nodesPerStage := nodeCount / pipelineStages
	if nodesPerStage == 0 {
		nodesPerStage = 1
	}

	// Calculate layers per stage
	layersPerStage := layerCount / pipelineStages
	if layersPerStage == 0 {
		layersPerStage = 1
	}

	partitions := make([]*Partition, 0)
	stageIndex := 0

	// Create pipeline stages
	for i := 0; i < layerCount; i += layersPerStage {
		endLayer := i + layersPerStage
		if endLayer > layerCount {
			endLayer = layerCount
		}

		// For each pipeline stage, split context across nodes in that stage
		stageNodes := make([]*NodeInfo, 0)
		for j := 0; j < nodesPerStage && (stageIndex*nodesPerStage+j) < nodeCount; j++ {
			nodeIndex := stageIndex*nodesPerStage + j
			if nodeIndex < len(task.Nodes) {
				stageNodes = append(stageNodes, task.Nodes[nodeIndex])
			}
		}

		if len(stageNodes) == 0 {
			continue
		}

		// Split context across nodes in this stage
		contextPerNode := task.Options.NumCtx / len(stageNodes)
		remainder := task.Options.NumCtx % len(stageNodes)

		for j, node := range stageNodes {
			startToken := j * contextPerNode
			endToken := startToken + contextPerNode
			if j < remainder {
				startToken += j
				endToken += j + 1
			} else {
				startToken += remainder
				endToken += remainder
			}

			partition := &Partition{
				ID:     fmt.Sprintf("partition_%d_hybrid_%d_%d", task.ID, stageIndex, j),
				NodeID: node.ID,
				Type:   PartitionTypeLayer,
				Data: map[string]interface{}{
					"start_layer": i,
					"end_layer":   endLayer,
					"layer_count": endLayer - i,
					"start_token": startToken,
					"end_token":   endToken,
					"token_count": endToken - startToken,
				},
				Dependencies: []string{}, // Will be set later
				Metadata: map[string]interface{}{
					"pipeline_stage": stageIndex,
					"tensor_split":   j,
				},
				EstimatedLatency: time.Duration((endLayer-i)*(endToken-startToken)*2) * time.Millisecond,
				EstimatedMemory:  int64((endLayer - i) * (endToken - startToken) * 2 * 1024), // Rough estimate
			}

			// Set dependencies (depends on previous pipeline stage)
			if stageIndex > 0 {
				// Depend on all partitions from previous stage
				for k := 0; k < len(stageNodes); k++ {
					partition.Dependencies = append(partition.Dependencies,
						fmt.Sprintf("partition_%d_hybrid_%d_%d", task.ID, stageIndex-1, k))
				}
			}

			partitions = append(partitions, partition)
		}

		stageIndex++
	}

	// Create plan
	plan := &PartitionPlan{
		ID:                  fmt.Sprintf("plan_%s", task.ID),
		Strategy:            hps.GetName(),
		Partitions:          partitions,
		Metadata:            make(map[string]interface{}),
		CreatedAt:           time.Now(),
		EstimatedLatency:    time.Duration(layerCount*task.Options.NumCtx*2) * time.Millisecond / time.Duration(nodeCount),
		EstimatedThroughput: float64(nodeCount), // Placeholder
		OptimizationScore:   0.9,                // High score for hybrid approach
	}

	// Update metrics
	hps.metrics.TotalPartitions += int64(len(partitions))
	hps.metrics.SuccessfulPartitions += int64(len(partitions))
	hps.metrics.LastUsed = time.Now()
	hps.metrics.AverageLatency = (hps.metrics.AverageLatency*time.Duration(hps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(hps.metrics.SuccessfulPartitions)

	return plan, nil
}

// NewAdaptivePartitioningStrategy creates a new adaptive partitioning strategy
func NewAdaptivePartitioningStrategy() *AdaptivePartitioningStrategy {
	return &AdaptivePartitioningStrategy{
		name: "adaptive",
		metrics: &StrategyMetrics{
			LastUsed: time.Now(),
		},
		thresholds: map[string]float64{
			"large_model":      5.0 * 1024 * 1024 * 1024, // 5GB
			"large_context":    2048,
			"many_layers":      20,
			"high_parallelism": 0.8,
		},
		learning: true,
		accuracy: 0.7, // Initial accuracy
	}
}

// GetName returns the strategy name
func (aps *AdaptivePartitioningStrategy) GetName() string {
	return aps.name
}

// GetMetrics returns strategy metrics
func (aps *AdaptivePartitioningStrategy) GetMetrics() *StrategyMetrics {
	return aps.metrics
}

// CanHandle checks if this strategy can handle the task
func (aps *AdaptivePartitioningStrategy) CanHandle(task *PartitionTask) bool {
	// This strategy can handle any task
	return true
}

// Partition implements adaptive partitioning based on workload analysis
func (aps *AdaptivePartitioningStrategy) Partition(ctx context.Context, task *PartitionTask) (*PartitionPlan, error) {
	start := time.Now()

	// Analyze workload characteristics
	modelSize := aps.estimateModelSize(task)
	contextLength := task.Options.NumCtx
	layerCount := aps.estimateLayerCount(task)
	parallelizability := aps.estimateParallelizability(task)
	nodeCount := len(task.Nodes)

	// Select the best strategy based on workload analysis
	var plan *PartitionPlan
	var err error

	// For very large models, use pipeline parallelism
	if modelSize > aps.thresholds["large_model"] && layerCount > int(aps.thresholds["many_layers"]) {
		strategy := NewPipelineParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if contextLength > int(aps.thresholds["large_context"]) && parallelizability > aps.thresholds["high_parallelism"] {
		// For large context with high parallelizability, use tensor parallelism
		strategy := NewTensorParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if modelSize > aps.thresholds["large_model"] && contextLength > int(aps.thresholds["large_context"]) {
		// For both large model and large context, use hybrid parallelism
		strategy := NewHybridParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else if nodeCount > 1 && layerCount > int(aps.thresholds["many_layers"]) {
		// For multi-node setups with sufficient layers, use pipeline parallelism
		strategy := NewPipelineParallelismStrategy()
		plan, err = strategy.Partition(ctx, task)
	} else {
		// Default to layerwise partitioning
		strategy := NewLayerwiseStrategy()
		plan, err = strategy.Partition(ctx, task)
	}

	if err != nil {
		aps.metrics.FailedPartitions++
		return nil, fmt.Errorf("failed to partition task: %w", err)
	}

	// Update metrics
	aps.metrics.TotalPartitions += int64(len(plan.Partitions))
	aps.metrics.SuccessfulPartitions += int64(len(plan.Partitions))
	aps.metrics.LastUsed = time.Now()
	aps.metrics.AverageLatency = (aps.metrics.AverageLatency*time.Duration(aps.metrics.SuccessfulPartitions-1) +
		time.Since(start)) / time.Duration(aps.metrics.SuccessfulPartitions)

	// Record result for learning
	result := &PartitionResult{
		Plan:             plan,
		ActualLatency:    time.Since(start),
		ActualThroughput: 1.0, // Placeholder
		Success:          true,
		Timestamp:        time.Now(),
	}
	aps.recordResult(result)

	return plan, nil
}

// estimateModelSize estimates the size of a model
func (aps *AdaptivePartitioningStrategy) estimateModelSize(task *PartitionTask) float64 {
	if task.GGML != nil {
		return float64(task.GGML.Length)
	}
	// Fallback estimation based on model name patterns
	return 4.0 * 1024 * 1024 * 1024 // 4GB default
}

// estimateLayerCount estimates the number of layers in a model
func (aps *AdaptivePartitioningStrategy) estimateLayerCount(task *PartitionTask) int {
	if task.GGML != nil {
		kv := task.GGML.KV()
		if layers := kv.Uint("llm.layers"); layers > 0 {
			return int(layers)
		}
	}
	// Fallback estimation
	return 24 // Default for many transformer models
}

// estimateParallelizability estimates how parallelizable a task is
func (aps *AdaptivePartitioningStrategy) estimateParallelizability(task *PartitionTask) float64 {
	// Factors that affect parallelizability:
	// 1. Model architecture (transformers are more parallelizable)
	// 2. Context length (longer contexts are more parallelizable)
	// 3. Batch size (larger batches are more parallelizable)

	contextLength := float64(task.Options.NumCtx)
	batchSize := float64(1) // Default batch size

	// Base parallelizability on context length and batch size
	parallelizability := math.Min((contextLength/2048.0)*(batchSize/4.0), 1.0)

	// Adjust based on model type
	if task.Model != nil {
		// Check if model is a transformer (more parallelizable)
		if aps.isTransformerModel(task.Model) {
			parallelizability *= 1.2
		}
	}

	return math.Min(parallelizability, 1.0)
}

// isTransformerModel checks if a model is a transformer-based model
func (aps *AdaptivePartitioningStrategy) isTransformerModel(model interface{}) bool {
	// Check model family or architecture
	if modelStruct, ok := model.(struct{ Name string }); ok {
		modelName := modelStruct.Name
		if modelName != "" {
			// Check if model is a transformer (more parallelizable)
			transformerModels := []string{"llama", "mistral", "gpt", "gemma"}
			for _, transformer := range transformerModels {
				if len(modelName) >= len(transformer) &&
					(modelName[:len(transformer)] == transformer ||
						modelName[len(modelName)-len(transformer):] == transformer) {
					return true
				}
			}
		}
	}

	return false
}

// recordResult records a partitioning result for learning
func (aps *AdaptivePartitioningStrategy) recordResult(result *PartitionResult) {
	// In a real implementation, this would record the actual performance
	// For now, we'll just update the accuracy

	if aps.learning {
		if result.Success {
			aps.accuracy = (aps.accuracy*0.9 + 0.1*1.0) // Increase accuracy for successful result
		} else {
			aps.accuracy = (aps.accuracy*0.9 + 0.1*0.0) // Decrease accuracy for failed result
		}
	}
}

// GetAccuracy returns the strategy accuracy
func (aps *AdaptivePartitioningStrategy) GetAccuracy() float64 {
	return aps.accuracy
}

// SetAccuracy sets the strategy accuracy
func (aps *AdaptivePartitioningStrategy) SetAccuracy(accuracy float64) {
	aps.accuracy = accuracy
}

// GetThresholds returns strategy thresholds
func (aps *AdaptivePartitioningStrategy) GetThresholds() map[string]float64 {
	return aps.thresholds
}

// SetThresholds sets strategy thresholds
func (aps *AdaptivePartitioningStrategy) SetThresholds(thresholds map[string]float64) {
	aps.thresholds = thresholds
}

// GetLearning returns whether learning is enabled
func (aps *AdaptivePartitioningStrategy) GetLearning() bool {
	return aps.learning
}

// SetLearning sets whether learning is enabled
func (aps *AdaptivePartitioningStrategy) SetLearning(learning bool) {
	aps.learning = learning
}
