package orchestration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// OrchestrationEngine manages distributed task orchestration
type OrchestrationEngine struct {
	config         *Config
	scheduler      *DistributedScheduler
	coordinator    *RequestCoordinator
	aggregator     *ResponseAggregator
	monitor        *OrchestrationMonitor
	activeTasks    map[string]*OrchestrationTask
	activeTasksMu  sync.RWMutex
	metrics        *OrchestrationMetrics
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	started        bool
}

// Config holds orchestration configuration
type Config struct {
	ClusterManager interface{} `json:"cluster_manager"`
	LoadBalancer   interface{} `json:"load_balancer"`
	FaultTolerance interface{} `json:"fault_tolerance"`
	MaxConcurrentTasks int     `json:"max_concurrent_tasks"`
	TaskTimeout        time.Duration `json:"task_timeout"`
	RetryPolicy        *RetryPolicy  `json:"retry_policy"`
	CoordinationMode   string        `json:"coordination_mode"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DistributedScheduler interface for scheduler operations
type DistributedScheduler interface {
	GetNodes() []interface{}
	GetMetrics() interface{}
}

// RequestCoordinator handles request coordination
type RequestCoordinator struct {
	engine      *OrchestrationEngine
	router      *RequestRouter
	partitioner *RequestPartitioner
	synchronizer *RequestSynchronizer
	stateManager *SessionStateManager
}

// RequestRouter handles request routing
type RequestRouter struct {
	coordinator *RequestCoordinator
	rules       []RoutingRule
	balancer    interface{}
	fallback    *FallbackStrategy
	metrics     *RoutingMetrics
}

// RoutingRule defines routing rules
type RoutingRule struct {
	ID        string                 `json:"id"`
	Condition string                 `json:"condition"`
	Action    string                 `json:"action"`
	Target    string                 `json:"target"`
	Priority  int                    `json:"priority"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// FallbackStrategy defines fallback behavior
type FallbackStrategy struct {
	Type        string                 `json:"type"`
	Target      string                 `json:"target"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RoutingMetrics tracks routing performance
type RoutingMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRoutes   int64         `json:"successful_routes"`
	FailedRoutes       int64         `json:"failed_routes"`
	AverageLatency     time.Duration `json:"average_latency"`
	RuleHitRate        float64       `json:"rule_hit_rate"`
	FallbackUsage      int64         `json:"fallback_usage"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// RequestPartitioner handles request partitioning
type RequestPartitioner struct {
	coordinator *RequestCoordinator
	strategies  map[string]PartitioningStrategy
	optimizer   *PartitionOptimizer
}

// PartitioningStrategy interface for partitioning
type PartitioningStrategy interface {
	Partition(request *OrchestrationRequest) (*PartitionPlan, error)
	GetName() string
}

// PartitionOptimizer optimizes partitioning decisions
type PartitionOptimizer struct {
	partitioner *RequestPartitioner
	history     []*PartitionResult
	historyMu   sync.RWMutex
	optimizationWeights map[string]float64
}

// PartitionPlan represents a partitioning plan
type PartitionPlan struct {
	ID          string                 `json:"id"`
	Strategy    string                 `json:"strategy"`
	Partitions  []*TaskPartition       `json:"partitions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// TaskPartition represents a task partition
type TaskPartition struct {
	ID           string                 `json:"id"`
	NodeID       string                 `json:"node_id"`
	Type         string                 `json:"type"`
	Data         interface{}            `json:"data"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PartitionResult represents partitioning results
type PartitionResult struct {
	Plan         *PartitionPlan `json:"plan"`
	Success      bool           `json:"success"`
	Latency      time.Duration  `json:"latency"`
	Throughput   float64        `json:"throughput"`
	Timestamp    time.Time      `json:"timestamp"`
}

// RequestSynchronizer handles request synchronization
type RequestSynchronizer struct {
	coordinator   *RequestCoordinator
	syncPoints    map[string]*SyncPoint
	syncPointsMu  sync.RWMutex
	barriers      map[string]*SyncBarrier
	barriersMu    sync.RWMutex
}

// SyncPoint represents a synchronization point
type SyncPoint struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Condition   string                 `json:"condition"`
	WaitingTasks []string              `json:"waiting_tasks"`
	Completed   bool                   `json:"completed"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SyncBarrier represents a synchronization barrier
type SyncBarrier struct {
	ID            string                 `json:"id"`
	RequiredTasks []string               `json:"required_tasks"`
	CompletedTasks []string              `json:"completed_tasks"`
	WaitingTasks  []string               `json:"waiting_tasks"`
	Released      bool                   `json:"released"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
}

// SessionStateManager manages session state
type SessionStateManager struct {
	coordinator *RequestCoordinator
	sessions    map[string]*SessionState
	sessionsMu  sync.RWMutex
	persistence SessionPersistence
}

// SessionState represents session state
type SessionState struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	Requests    []string               `json:"requests"`
	Responses   []string               `json:"responses"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SessionPersistence interface for session persistence
type SessionPersistence interface {
	Save(session *SessionState) error
	Load(id string) (*SessionState, error)
	Delete(id string) error
	List() ([]*SessionState, error)
}

// ResponseAggregator handles response aggregation
type ResponseAggregator struct {
	engine      *OrchestrationEngine
	strategies  map[string]AggregationStrategy
	pendingResults map[string]*AggregationContext
	pendingMu   sync.RWMutex
}

// AggregationStrategy interface for aggregation
type AggregationStrategy interface {
	Aggregate(context *AggregationContext) (*AggregatedResponse, error)
	GetName() string
}

// AggregationContext holds aggregation context
type AggregationContext struct {
	TaskID      string                 `json:"task_id"`
	Strategy    string                 `json:"strategy"`
	PartialResults []PartialResult     `json:"partial_results"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// PartialResult represents a partial result
type PartialResult struct {
	PartitionID string                 `json:"partition_id"`
	NodeID      string                 `json:"node_id"`
	Data        interface{}            `json:"data"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AggregatedResponse represents an aggregated response
type AggregatedResponse struct {
	TaskID      string                 `json:"task_id"`
	Strategy    string                 `json:"strategy"`
	Data        interface{}            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Latency     time.Duration          `json:"latency"`
	Timestamp   time.Time              `json:"timestamp"`
}

// OrchestrationMonitor monitors orchestration performance
type OrchestrationMonitor struct {
	engine    *OrchestrationEngine
	metrics   *OrchestrationMetrics
	monitors  []Monitor
	interval  time.Duration
	stopCh    chan struct{}
}

// Monitor interface for monitoring
type Monitor interface {
	Collect() (map[string]interface{}, error)
	GetName() string
}

// OrchestrationMetrics tracks orchestration metrics
type OrchestrationMetrics struct {
	TotalTasks         int64         `json:"total_tasks"`
	ActiveTasks        int64         `json:"active_tasks"`
	CompletedTasks     int64         `json:"completed_tasks"`
	FailedTasks        int64         `json:"failed_tasks"`
	AverageLatency     time.Duration `json:"average_latency"`
	Throughput         float64       `json:"throughput"`
	ResourceUtilization float64      `json:"resource_utilization"`
	ErrorRate          float64       `json:"error_rate"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// OrchestrationTask represents a task being orchestrated
type OrchestrationTask struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Request         *OrchestrationRequest  `json:"request"`
	PartitionPlan   *PartitionPlan         `json:"partition_plan"`
	PartialResults  []PartialResult        `json:"partial_results"`
	AggregatedResult *AggregatedResponse   `json:"aggregated_result"`
	Status          TaskStatus             `json:"status"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at"`
	Metadata        map[string]interface{} `json:"metadata"`
	RetryCount      int                    `json:"retry_count"`
	LastError       string                 `json:"last_error"`
}

// TaskStatus represents task status
type TaskStatus string

const (
	TaskStatusPending     TaskStatus = "pending"
	TaskStatusPartitioned TaskStatus = "partitioned"
	TaskStatusExecuting   TaskStatus = "executing"
	TaskStatusAggregating TaskStatus = "aggregating"
	TaskStatusCompleted   TaskStatus = "completed"
	TaskStatusFailed      TaskStatus = "failed"
	TaskStatusRetrying    TaskStatus = "retrying"
)

// OrchestrationRequest represents a request for orchestration
type OrchestrationRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Payload     interface{}            `json:"payload"`
	Options     map[string]interface{} `json:"options"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewOrchestrationEngine creates a new orchestration engine
func NewOrchestrationEngine(config *Config) *OrchestrationEngine {
	ctx, cancel := context.WithCancel(context.Background())
	
	oe := &OrchestrationEngine{
		config:      config,
		activeTasks: make(map[string]*OrchestrationTask),
		metrics:     &OrchestrationMetrics{LastUpdated: time.Now()},
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Initialize components
	oe.initializeComponents()
	
	return oe
}

// initializeComponents initializes all orchestration components
func (oe *OrchestrationEngine) initializeComponents() {
	// Initialize request coordinator
	oe.coordinator = &RequestCoordinator{
		engine: oe,
	}
	
	// Initialize request router
	oe.coordinator.router = &RequestRouter{
		coordinator: oe.coordinator,
		rules:       make([]RoutingRule, 0),
		fallback: &FallbackStrategy{
			Type:    "local",
			Timeout: 30 * time.Second,
		},
		metrics: &RoutingMetrics{LastUpdated: time.Now()},
	}
	
	// Initialize request partitioner
	oe.coordinator.partitioner = &RequestPartitioner{
		coordinator: oe.coordinator,
		strategies:  make(map[string]PartitioningStrategy),
		optimizer: &PartitionOptimizer{
			history: make([]*PartitionResult, 0),
			optimizationWeights: map[string]float64{
				"latency":    0.4,
				"throughput": 0.3,
				"resource":   0.2,
				"reliability": 0.1,
			},
		},
	}
	
	// Initialize request synchronizer
	oe.coordinator.synchronizer = &RequestSynchronizer{
		coordinator: oe.coordinator,
		syncPoints:  make(map[string]*SyncPoint),
		barriers:    make(map[string]*SyncBarrier),
	}
	
	// Initialize session state manager
	oe.coordinator.stateManager = &SessionStateManager{
		coordinator: oe.coordinator,
		sessions:    make(map[string]*SessionState),
	}
	
	// Initialize response aggregator
	oe.aggregator = &ResponseAggregator{
		engine:         oe,
		strategies:     make(map[string]AggregationStrategy),
		pendingResults: make(map[string]*AggregationContext),
	}
	
	// Initialize orchestration monitor
	oe.monitor = &OrchestrationMonitor{
		engine:   oe,
		metrics:  oe.metrics,
		monitors: make([]Monitor, 0),
		interval: 10 * time.Second,
		stopCh:   make(chan struct{}),
	}
	
	// Register default strategies
	oe.registerDefaultStrategies()
}

// registerDefaultStrategies registers default strategies
func (oe *OrchestrationEngine) registerDefaultStrategies() {
	// Register aggregation strategies
	oe.aggregator.strategies["concat"] = &ConcatAggregationStrategy{}
	oe.aggregator.strategies["average"] = &AverageAggregationStrategy{}
	oe.aggregator.strategies["weighted"] = &WeightedAggregationStrategy{}
	
	// Register partitioning strategies
	oe.coordinator.partitioner.strategies["round_robin"] = &RoundRobinPartitioningStrategy{}
	oe.coordinator.partitioner.strategies["load_based"] = &LoadBasedPartitioningStrategy{}
	oe.coordinator.partitioner.strategies["capability_based"] = &CapabilityBasedPartitioningStrategy{}
}

// Start starts the orchestration engine
func (oe *OrchestrationEngine) Start(ctx context.Context) error {
	oe.mu.Lock()
	defer oe.mu.Unlock()
	
	if oe.started {
		return fmt.Errorf("orchestration engine already started")
	}
	
	// Start monitoring
	go oe.monitor.start(oe.ctx)
	
	// Start request processing
	go oe.processRequests(oe.ctx)
	
	// Start result aggregation
	go oe.processAggregation(oe.ctx)
	
	oe.started = true
	
	slog.Info("orchestration engine started",
		"max_concurrent_tasks", oe.config.MaxConcurrentTasks,
		"coordination_mode", oe.config.CoordinationMode)
	
	return nil
}

// ExecuteTask executes a distributed task
func (oe *OrchestrationEngine) ExecuteTask(ctx context.Context, task interface{}) error {
	// Convert task to orchestration request
	request := oe.convertToOrchestrationRequest(task)
	
	// Create orchestration task
	orchTask := &OrchestrationTask{
		ID:         request.ID,
		Type:       request.Type,
		Request:    request,
		Status:     TaskStatusPending,
		StartedAt:  time.Now(),
		Metadata:   make(map[string]interface{}),
		RetryCount: 0,
	}
	
	// Store active task
	oe.activeTasksMu.Lock()
	oe.activeTasks[orchTask.ID] = orchTask
	oe.activeTasksMu.Unlock()
	
	// Update metrics
	oe.metrics.TotalTasks++
	oe.metrics.ActiveTasks++
	
	// Execute task asynchronously
	go oe.executeTaskAsync(ctx, orchTask)
	
	return nil
}

// convertToOrchestrationRequest converts a task to orchestration request
func (oe *OrchestrationEngine) convertToOrchestrationRequest(task interface{}) *OrchestrationRequest {
	return &OrchestrationRequest{
		ID:        fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Type:      "distributed_inference",
		Payload:   task,
		Options:   make(map[string]interface{}),
		Priority:  1,
		Timeout:   oe.config.TaskTimeout,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
}

// executeTaskAsync executes a task asynchronously
func (oe *OrchestrationEngine) executeTaskAsync(ctx context.Context, task *OrchestrationTask) {
	defer func() {
		// Clean up task
		oe.activeTasksMu.Lock()
		delete(oe.activeTasks, task.ID)
		oe.activeTasksMu.Unlock()
		
		// Update metrics
		oe.metrics.ActiveTasks--
		if task.Status == TaskStatusCompleted {
			oe.metrics.CompletedTasks++
		} else {
			oe.metrics.FailedTasks++
		}
	}()
	
	for {
		switch task.Status {
		case TaskStatusPending:
			if err := oe.partitionTask(ctx, task); err != nil {
				if oe.shouldRetry(task, err) {
					oe.retryTask(task, err)
					continue
				}
				oe.failTask(task, err)
				return
			}
			task.Status = TaskStatusPartitioned
			
		case TaskStatusPartitioned:
			if err := oe.executePartitions(ctx, task); err != nil {
				if oe.shouldRetry(task, err) {
					oe.retryTask(task, err)
					continue
				}
				oe.failTask(task, err)
				return
			}
			task.Status = TaskStatusExecuting
			
		case TaskStatusExecuting:
			if oe.arePartitionsComplete(task) {
				task.Status = TaskStatusAggregating
			} else {
				// Wait for partitions to complete
				time.Sleep(100 * time.Millisecond)
			}
			
		case TaskStatusAggregating:
			if err := oe.aggregateResults(ctx, task); err != nil {
				if oe.shouldRetry(task, err) {
					oe.retryTask(task, err)
					continue
				}
				oe.failTask(task, err)
				return
			}
			task.Status = TaskStatusCompleted
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			
		case TaskStatusCompleted:
			slog.Info("task completed", "task_id", task.ID, "duration", time.Since(task.StartedAt))
			return
			
		case TaskStatusFailed:
			slog.Error("task failed", "task_id", task.ID, "error", task.LastError)
			return
			
		case TaskStatusRetrying:
			// Wait before retrying
			delay := oe.calculateRetryDelay(task.RetryCount)
			time.Sleep(delay)
			task.Status = TaskStatusPending
			
		default:
			slog.Error("unknown task status", "task_id", task.ID, "status", task.Status)
			return
		}
	}
}

// partitionTask partitions a task for distributed execution
func (oe *OrchestrationEngine) partitionTask(ctx context.Context, task *OrchestrationTask) error {
	// Select partitioning strategy
	strategy, err := oe.selectPartitioningStrategy(task)
	if err != nil {
		return fmt.Errorf("failed to select partitioning strategy: %v", err)
	}
	
	// Partition the task
	plan, err := strategy.Partition(task.Request)
	if err != nil {
		return fmt.Errorf("failed to partition task: %v", err)
	}
	
	task.PartitionPlan = plan
	return nil
}

// selectPartitioningStrategy selects the best partitioning strategy
func (oe *OrchestrationEngine) selectPartitioningStrategy(task *OrchestrationTask) (PartitioningStrategy, error) {
	// Simple strategy selection based on task type
	switch task.Type {
	case "distributed_inference":
		return oe.coordinator.partitioner.strategies["load_based"], nil
	case "batch_processing":
		return oe.coordinator.partitioner.strategies["round_robin"], nil
	default:
		return oe.coordinator.partitioner.strategies["capability_based"], nil
	}
}

// executePartitions executes task partitions
func (oe *OrchestrationEngine) executePartitions(ctx context.Context, task *OrchestrationTask) error {
	// Execute partitions in parallel
	for _, partition := range task.PartitionPlan.Partitions {
		go oe.executePartition(ctx, task, partition)
	}
	
	return nil
}

// executePartition executes a single partition
func (oe *OrchestrationEngine) executePartition(ctx context.Context, task *OrchestrationTask, partition *TaskPartition) {
	start := time.Now()
	
	// Simulate partition execution
	time.Sleep(100 * time.Millisecond)
	
	// Create partial result
	result := PartialResult{
		PartitionID: partition.ID,
		NodeID:      partition.NodeID,
		Data:        "mock_result",
		Metadata:    make(map[string]interface{}),
		Timestamp:   time.Now(),
	}
	
	// Store partial result
	task.PartialResults = append(task.PartialResults, result)
	
	slog.Debug("partition executed", "task_id", task.ID, "partition_id", partition.ID, "duration", time.Since(start))
}

// arePartitionsComplete checks if all partitions are complete
func (oe *OrchestrationEngine) arePartitionsComplete(task *OrchestrationTask) bool {
	return len(task.PartialResults) >= len(task.PartitionPlan.Partitions)
}

// aggregateResults aggregates partial results
func (oe *OrchestrationEngine) aggregateResults(ctx context.Context, task *OrchestrationTask) error {
	// Create aggregation context
	aggCtx := &AggregationContext{
		TaskID:         task.ID,
		Strategy:       "concat", // Default strategy
		PartialResults: task.PartialResults,
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
	}
	
	// Select aggregation strategy
	strategy := oe.aggregator.strategies[aggCtx.Strategy]
	if strategy == nil {
		return fmt.Errorf("aggregation strategy not found: %s", aggCtx.Strategy)
	}
	
	// Aggregate results
	aggregated, err := strategy.Aggregate(aggCtx)
	if err != nil {
		return fmt.Errorf("failed to aggregate results: %v", err)
	}
	
	task.AggregatedResult = aggregated
	return nil
}

// shouldRetry determines if a task should be retried
func (oe *OrchestrationEngine) shouldRetry(task *OrchestrationTask, err error) bool {
	if oe.config.RetryPolicy == nil {
		return false
	}
	
	return task.RetryCount < oe.config.RetryPolicy.MaxRetries
}

// retryTask prepares a task for retry
func (oe *OrchestrationEngine) retryTask(task *OrchestrationTask, err error) {
	task.RetryCount++
	task.LastError = err.Error()
	task.Status = TaskStatusRetrying
	
	slog.Warn("retrying task", "task_id", task.ID, "retry_count", task.RetryCount, "error", err)
}

// failTask marks a task as failed
func (oe *OrchestrationEngine) failTask(task *OrchestrationTask, err error) {
	task.Status = TaskStatusFailed
	task.LastError = err.Error()
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	
	slog.Error("task failed", "task_id", task.ID, "error", err)
}

// calculateRetryDelay calculates retry delay with exponential backoff
func (oe *OrchestrationEngine) calculateRetryDelay(retryCount int) time.Duration {
	if oe.config.RetryPolicy == nil {
		return time.Second
	}
	
	delay := oe.config.RetryPolicy.InitialDelay
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * oe.config.RetryPolicy.BackoffFactor)
		if delay > oe.config.RetryPolicy.MaxDelay {
			delay = oe.config.RetryPolicy.MaxDelay
			break
		}
	}
	
	return delay
}

// processRequests processes incoming requests
func (oe *OrchestrationEngine) processRequests(ctx context.Context) {
	slog.Info("request processor started")
	
	for {
		select {
		case <-ctx.Done():
			slog.Info("request processor shutting down")
			return
		default:
			// Process any pending requests
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processAggregation processes result aggregation
func (oe *OrchestrationEngine) processAggregation(ctx context.Context) {
	slog.Info("aggregation processor started")
	
	for {
		select {
		case <-ctx.Done():
			slog.Info("aggregation processor shutting down")
			return
		default:
			// Process any pending aggregations
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// GetMetrics returns orchestration metrics
func (oe *OrchestrationEngine) GetMetrics() *OrchestrationMetrics {
	oe.mu.RLock()
	defer oe.mu.RUnlock()
	
	// Update metrics
	if oe.metrics.TotalTasks > 0 {
		oe.metrics.ErrorRate = float64(oe.metrics.FailedTasks) / float64(oe.metrics.TotalTasks)
		oe.metrics.Throughput = float64(oe.metrics.CompletedTasks) / time.Since(oe.metrics.LastUpdated).Seconds()
	}
	
	return oe.metrics
}

// GetActiveTasks returns active tasks
func (oe *OrchestrationEngine) GetActiveTasks() []*OrchestrationTask {
	oe.activeTasksMu.RLock()
	defer oe.activeTasksMu.RUnlock()
	
	tasks := make([]*OrchestrationTask, 0, len(oe.activeTasks))
	for _, task := range oe.activeTasks {
		tasks = append(tasks, task)
	}
	
	return tasks
}

// Shutdown gracefully shuts down the orchestration engine
func (oe *OrchestrationEngine) Shutdown(ctx context.Context) error {
	oe.mu.Lock()
	defer oe.mu.Unlock()
	
	if !oe.started {
		return nil
	}
	
	slog.Info("shutting down orchestration engine")
	
	// Stop monitoring
	close(oe.monitor.stopCh)
	
	// Cancel context
	oe.cancel()
	
	// Wait for active tasks to complete
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	for {
		oe.activeTasksMu.RLock()
		activeCount := len(oe.activeTasks)
		oe.activeTasksMu.RUnlock()
		
		if activeCount == 0 {
			break
		}
		
		select {
		case <-shutdownCtx.Done():
			slog.Warn("shutdown timeout, forcing shutdown with active tasks", "active_tasks", activeCount)
			return nil
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	oe.started = false
	
	return nil
}

// Monitor methods

// start starts the orchestration monitor
func (om *OrchestrationMonitor) start(ctx context.Context) {
	ticker := time.NewTicker(om.interval)
	defer ticker.Stop()
	
	slog.Info("orchestration monitor started")
	
	for {
		select {
		case <-ctx.Done():
			slog.Info("orchestration monitor shutting down")
			return
		case <-om.stopCh:
			return
		case <-ticker.C:
			om.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all monitors
func (om *OrchestrationMonitor) collectMetrics() {
	for _, monitor := range om.monitors {
		metrics, err := monitor.Collect()
		if err != nil {
			slog.Warn("failed to collect metrics", "monitor", monitor.GetName(), "error", err)
			continue
		}
		
		// Process metrics
		om.processMetrics(monitor.GetName(), metrics)
	}
	
	// Update last updated time
	om.metrics.LastUpdated = time.Now()
}

// processMetrics processes collected metrics
func (om *OrchestrationMonitor) processMetrics(monitorName string, metrics map[string]interface{}) {
	// Process metrics based on monitor type
	switch monitorName {
	case "resource_monitor":
		if util, ok := metrics["resource_utilization"].(float64); ok {
			om.metrics.ResourceUtilization = util
		}
	case "performance_monitor":
		if latency, ok := metrics["average_latency"].(time.Duration); ok {
			om.metrics.AverageLatency = latency
		}
	case "throughput_monitor":
		if throughput, ok := metrics["throughput"].(float64); ok {
			om.metrics.Throughput = throughput
		}
	}
}
