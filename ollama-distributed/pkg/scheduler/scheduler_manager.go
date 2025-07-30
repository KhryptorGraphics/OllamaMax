package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/libp2p/go-libp2p/core/peer"
)

// SchedulerManager manages the complete distributed scheduling system
type SchedulerManager struct {
	config *SchedulerManagerConfig

	// Core components
	engine        *Engine
	taskQueue     *TaskQueue
	workerManager *WorkerManager
	loadBalancer  *TaskLoadBalancer
	taskTracker   *TaskTracker

	// Integration components
	p2pNode          *p2p.Node
	consensusManager *consensus.ConsensusManager
	messageRouter    *messaging.MessageRouter
	networkMonitor   *monitoring.NetworkMonitor

	// Messaging handlers
	schedulerHandler *messaging.SchedulerHandler

	// State management
	state   *SchedulerState
	stateMu sync.RWMutex

	// Metrics and monitoring
	metrics *SchedulerMetrics

	// Lifecycle
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	started   bool
	startedMu sync.RWMutex
}

// SchedulerManagerConfig configures the scheduler manager
type SchedulerManagerConfig struct {
	// Basic settings
	NodeID    string
	ClusterID string

	// Scheduler settings
	SchedulerConfig *config.SchedulerConfig

	// Task queue settings
	MaxQueueSize int
	QueueTimeout time.Duration

	// Worker settings
	MaxWorkers          int
	WorkerTimeout       time.Duration
	HealthCheckInterval time.Duration

	// Load balancing settings
	LoadBalanceAlgorithm string
	LoadBalanceInterval  time.Duration

	// Performance settings
	MetricsInterval  time.Duration
	EnableMonitoring bool

	// Integration settings
	EnableConsensus    bool
	EnableP2PMessaging bool
}

// SchedulerState represents the current state of the scheduler
type SchedulerState struct {
	// Scheduler status
	Status   SchedulerStatus `json:"status"`
	IsLeader bool            `json:"is_leader"`
	LeaderID peer.ID         `json:"leader_id"`

	// Task statistics
	TotalTasks     int64 `json:"total_tasks"`
	QueuedTasks    int64 `json:"queued_tasks"`
	RunningTasks   int64 `json:"running_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`

	// Worker statistics
	TotalWorkers  int64 `json:"total_workers"`
	ActiveWorkers int64 `json:"active_workers"`
	IdleWorkers   int64 `json:"idle_workers"`

	// Performance metrics
	AverageTaskTime time.Duration `json:"average_task_time"`
	TaskThroughput  float64       `json:"task_throughput"`
	SystemLoad      float64       `json:"system_load"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// SchedulerMetrics tracks scheduler performance
type SchedulerMetrics struct {
	// Task metrics
	TasksScheduled int64 `json:"tasks_scheduled"`
	TasksCompleted int64 `json:"tasks_completed"`
	TasksFailed    int64 `json:"tasks_failed"`
	TasksRetried   int64 `json:"tasks_retried"`

	// Performance metrics
	AverageLatency      time.Duration `json:"average_latency"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	QueueUtilization    float64       `json:"queue_utilization"`
	WorkerUtilization   float64       `json:"worker_utilization"`

	// Error metrics
	SchedulingErrors    int64 `json:"scheduling_errors"`
	WorkerErrors        int64 `json:"worker_errors"`
	CommunicationErrors int64 `json:"communication_errors"`

	// Resource metrics
	MemoryUsage  int64   `json:"memory_usage"`
	CPUUsage     float64 `json:"cpu_usage"`
	NetworkUsage int64   `json:"network_usage"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
	mu          sync.RWMutex
}

// TaskQueue manages the task queue with priority support
type TaskQueue struct {
	config *TaskQueueConfig

	// Queue storage
	highPriorityQueue   chan *Task
	normalPriorityQueue chan *Task
	lowPriorityQueue    chan *Task

	// Queue metrics
	metrics *QueueMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TaskQueueConfig configures the task queue
type TaskQueueConfig struct {
	MaxSize             int
	Timeout             time.Duration
	EnablePriority      bool
	HighPriorityRatio   float64
	NormalPriorityRatio float64
	LowPriorityRatio    float64
}

// WorkerManager manages worker nodes and their capabilities
type WorkerManager struct {
	config *WorkerManagerConfig

	// Worker registry
	workers   map[peer.ID]*WorkerNode
	workersMu sync.RWMutex

	// Capability tracking
	capabilities   map[string][]*WorkerNode
	capabilitiesMu sync.RWMutex

	// Health monitoring
	healthChecker *WorkerHealthChecker

	// Metrics
	metrics *WorkerMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// WorkerManagerConfig configures the worker manager
type WorkerManagerConfig struct {
	MaxWorkers          int
	HealthCheckInterval time.Duration
	WorkerTimeout       time.Duration
	CapabilityRefresh   time.Duration
}

// TaskTracker tracks task lifecycle and results
type TaskTracker struct {
	config *TaskTrackerConfig

	// Task tracking
	activeTasks   map[string]*TrackedTask
	activeTasksMu sync.RWMutex

	// Result collection
	results chan *TaskResult

	// Metrics
	metrics *TaskMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TaskTrackerConfig configures the task tracker
type TaskTrackerConfig struct {
	MaxActiveTasks   int
	TaskTimeout      time.Duration
	ResultBufferSize int
	CleanupInterval  time.Duration
}

// Data structures

// Task represents a schedulable task
type Task struct {
	ID           string                 `json:"id"`
	Type         TaskType               `json:"type"`
	Priority     TaskPriority           `json:"priority"`
	ModelName    string                 `json:"model_name"`
	Payload      []byte                 `json:"payload"`
	Requirements *ResourceRequirements  `json:"requirements"`
	Constraints  *TaskConstraints       `json:"constraints"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Lifecycle
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	// Assignment
	AssignedWorker peer.ID `json:"assigned_worker"`
	AssignedNode   string  `json:"assigned_node"`

	// Status
	Status     TaskStatus `json:"status"`
	Progress   float64    `json:"progress"`
	Error      string     `json:"error,omitempty"`
	RetryCount int        `json:"retry_count"`
	MaxRetries int        `json:"max_retries"`
}

// WorkerNode represents a worker node in the cluster
type WorkerNode struct {
	ID           peer.ID       `json:"id"`
	Address      string        `json:"address"`
	Status       WorkerStatus  `json:"status"`
	Capabilities []string      `json:"capabilities"`
	Resources    *ResourceInfo `json:"resources"`
	Load         *LoadInfo     `json:"load"`

	// Health
	LastSeen    time.Time `json:"last_seen"`
	HealthScore float64   `json:"health_score"`

	// Performance
	TasksCompleted  int64         `json:"tasks_completed"`
	TasksFailed     int64         `json:"tasks_failed"`
	AverageTaskTime time.Duration `json:"average_task_time"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// TrackedTask represents a task being tracked
type TrackedTask struct {
	Task       *Task       `json:"task"`
	Worker     *WorkerNode `json:"worker"`
	StartTime  time.Time   `json:"start_time"`
	LastUpdate time.Time   `json:"last_update"`
	Progress   float64     `json:"progress"`
	Status     TaskStatus  `json:"status"`
	Heartbeats []time.Time `json:"heartbeats"`
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	TaskID      string                `json:"task_id"`
	WorkerID    peer.ID               `json:"worker_id"`
	Status      TaskStatus            `json:"status"`
	Result      []byte                `json:"result"`
	Error       string                `json:"error,omitempty"`
	Metrics     *TaskExecutionMetrics `json:"metrics"`
	CompletedAt time.Time             `json:"completed_at"`
	Duration    time.Duration         `json:"duration"`
}

// ResourceRequirements specifies task resource requirements
type ResourceRequirements struct {
	CPU             float64  `json:"cpu"`
	Memory          int64    `json:"memory"`
	GPU             int      `json:"gpu"`
	Storage         int64    `json:"storage"`
	Bandwidth       int64    `json:"bandwidth"`
	SpecialHardware []string `json:"special_hardware"`
}

// TaskConstraints specifies task constraints
type TaskConstraints struct {
	MaxExecutionTime     time.Duration `json:"max_execution_time"`
	PreferredNodes       []peer.ID     `json:"preferred_nodes"`
	ExcludedNodes        []peer.ID     `json:"excluded_nodes"`
	RequiredCapabilities []string      `json:"required_capabilities"`
	Locality             string        `json:"locality"`
}

// ResourceInfo represents worker resource information
type ResourceInfo struct {
	TotalCPU         float64 `json:"total_cpu"`
	AvailableCPU     float64 `json:"available_cpu"`
	TotalMemory      int64   `json:"total_memory"`
	AvailableMemory  int64   `json:"available_memory"`
	TotalGPU         int     `json:"total_gpu"`
	AvailableGPU     int     `json:"available_gpu"`
	TotalStorage     int64   `json:"total_storage"`
	AvailableStorage int64   `json:"available_storage"`
}

// LoadInfo represents worker load information
type LoadInfo struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	GPUUsage     float64 `json:"gpu_usage"`
	NetworkUsage float64 `json:"network_usage"`
	ActiveTasks  int     `json:"active_tasks"`
	QueuedTasks  int     `json:"queued_tasks"`
}

// Enums and constants
type SchedulerStatus string

const (
	SchedulerStatusStarting SchedulerStatus = "starting"
	SchedulerStatusRunning  SchedulerStatus = "running"
	SchedulerStatusStopping SchedulerStatus = "stopping"
	SchedulerStatusStopped  SchedulerStatus = "stopped"
	SchedulerStatusError    SchedulerStatus = "error"
)

type TaskType string

const (
	TaskTypeInference      TaskType = "inference"
	TaskTypeTraining       TaskType = "training"
	TaskTypeEmbedding      TaskType = "embedding"
	TaskTypeClassification TaskType = "classification"
	TaskTypeGeneration     TaskType = "generation"
	TaskTypeCustom         TaskType = "custom"
)

type TaskPriority int

const (
	TaskPriorityLow      TaskPriority = 1
	TaskPriorityNormal   TaskPriority = 5
	TaskPriorityHigh     TaskPriority = 8
	TaskPriorityCritical TaskPriority = 10
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusScheduled TaskStatus = "scheduled"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusRetrying  TaskStatus = "retrying"
)

type WorkerStatus string

const (
	WorkerStatusOnline      WorkerStatus = "online"
	WorkerStatusOffline     WorkerStatus = "offline"
	WorkerStatusBusy        WorkerStatus = "busy"
	WorkerStatusIdle        WorkerStatus = "idle"
	WorkerStatusMaintenance WorkerStatus = "maintenance"
	WorkerStatusError       WorkerStatus = "error"
)

// NewSchedulerManager creates a new scheduler manager
func NewSchedulerManager(config *SchedulerManagerConfig, p2pNode *p2p.Node, consensusManager *consensus.ConsensusManager, messageRouter *messaging.MessageRouter, networkMonitor *monitoring.NetworkMonitor) (*SchedulerManager, error) {
	if config == nil {
		config = &SchedulerManagerConfig{
			MaxQueueSize:         10000,
			QueueTimeout:         30 * time.Second,
			MaxWorkers:           1000,
			WorkerTimeout:        60 * time.Second,
			HealthCheckInterval:  30 * time.Second,
			LoadBalanceAlgorithm: "least_loaded",
			LoadBalanceInterval:  10 * time.Second,
			MetricsInterval:      30 * time.Second,
			EnableMonitoring:     true,
			EnableConsensus:      true,
			EnableP2PMessaging:   true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &SchedulerManager{
		config:           config,
		p2pNode:          p2pNode,
		consensusManager: consensusManager,
		messageRouter:    messageRouter,
		networkMonitor:   networkMonitor,
		state: &SchedulerState{
			Status:      SchedulerStatusStopped,
			LastUpdated: time.Now(),
		},
		metrics: &SchedulerMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Create core components
	if err := manager.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Setup messaging handlers
	if err := manager.setupMessaging(); err != nil {
		return nil, fmt.Errorf("failed to setup messaging: %w", err)
	}

	return manager, nil
}

// initializeComponents initializes the core scheduler components
func (sm *SchedulerManager) initializeComponents() error {
	// Create task queue
	taskQueue, err := NewTaskQueue(&TaskQueueConfig{
		MaxSize:             sm.config.MaxQueueSize,
		Timeout:             sm.config.QueueTimeout,
		EnablePriority:      true,
		HighPriorityRatio:   0.3,
		NormalPriorityRatio: 0.5,
		LowPriorityRatio:    0.2,
	})
	if err != nil {
		return fmt.Errorf("failed to create task queue: %w", err)
	}
	sm.taskQueue = taskQueue

	// Create worker manager
	workerManager, err := NewWorkerManager(&WorkerManagerConfig{
		MaxWorkers:          sm.config.MaxWorkers,
		HealthCheckInterval: sm.config.HealthCheckInterval,
		WorkerTimeout:       sm.config.WorkerTimeout,
		CapabilityRefresh:   5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("failed to create worker manager: %w", err)
	}
	sm.workerManager = workerManager

	// Create load balancer
	loadBalancer, err := NewTaskLoadBalancer(&LoadBalancerConfig{
		Algorithm: sm.config.LoadBalanceAlgorithm,
		Interval:  sm.config.LoadBalanceInterval,
	}, workerManager)
	if err != nil {
		return fmt.Errorf("failed to create load balancer: %w", err)
	}
	sm.loadBalancer = loadBalancer

	// Create task tracker
	taskTracker, err := NewTaskTracker(&TaskTrackerConfig{
		MaxActiveTasks:   10000,
		TaskTimeout:      30 * time.Minute,
		ResultBufferSize: 1000,
		CleanupInterval:  5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("failed to create task tracker: %w", err)
	}
	sm.taskTracker = taskTracker

	// Create scheduler engine
	engine, err := NewEngine(sm.config.SchedulerConfig, sm.p2pNode, sm.consensusManager.GetEngine())
	if err != nil {
		return fmt.Errorf("failed to create scheduler engine: %w", err)
	}
	sm.engine = engine

	return nil
}

// setupMessaging sets up messaging handlers for scheduler communication
func (sm *SchedulerManager) setupMessaging() error {
	// Create scheduler handler
	sm.schedulerHandler = messaging.NewSchedulerHandler(sm.p2pNode.ID())

	// Register message callbacks
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerTaskAssignment, sm.handleTaskAssignment)
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerTaskResult, sm.handleTaskResult)
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerTaskStatus, sm.handleTaskStatus)
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerResourceUpdate, sm.handleResourceUpdate)
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerWorkerRegister, sm.handleWorkerRegister)
	sm.schedulerHandler.RegisterCallback(messaging.SchedulerWorkerHeartbeat, sm.handleWorkerHeartbeat)

	// Register handler with message router
	sm.messageRouter.RegisterHandler(sm.schedulerHandler)

	return nil
}

// Start starts the scheduler manager
func (sm *SchedulerManager) Start() error {
	sm.startedMu.Lock()
	defer sm.startedMu.Unlock()

	if sm.started {
		return nil
	}

	sm.state.Status = SchedulerStatusStarting

	// Start core components
	if err := sm.taskQueue.Start(); err != nil {
		return fmt.Errorf("failed to start task queue: %w", err)
	}

	if err := sm.workerManager.Start(); err != nil {
		return fmt.Errorf("failed to start worker manager: %w", err)
	}

	if err := sm.loadBalancer.Start(); err != nil {
		return fmt.Errorf("failed to start load balancer: %w", err)
	}

	if err := sm.taskTracker.Start(); err != nil {
		return fmt.Errorf("failed to start task tracker: %w", err)
	}

	if err := sm.engine.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler engine: %w", err)
	}

	// Start monitoring if enabled
	if sm.config.EnableMonitoring {
		sm.wg.Add(1)
		go sm.monitoringLoop()
	}

	// Start main scheduler loop
	sm.wg.Add(1)
	go sm.schedulerLoop()

	// Update state
	sm.stateMu.Lock()
	sm.state.Status = SchedulerStatusRunning
	sm.state.LastUpdated = time.Now()
	sm.stateMu.Unlock()

	sm.started = true
	return nil
}

// Stop stops the scheduler manager
func (sm *SchedulerManager) Stop() error {
	sm.startedMu.Lock()
	defer sm.startedMu.Unlock()

	if !sm.started {
		return nil
	}

	sm.stateMu.Lock()
	sm.state.Status = SchedulerStatusStopping
	sm.state.LastUpdated = time.Now()
	sm.stateMu.Unlock()

	sm.cancel()

	// Stop components
	if sm.engine != nil {
		sm.engine.Shutdown(context.Background())
	}

	if sm.taskTracker != nil {
		sm.taskTracker.Stop()
	}

	if sm.loadBalancer != nil {
		sm.loadBalancer.Stop()
	}

	if sm.workerManager != nil {
		sm.workerManager.Stop()
	}

	if sm.taskQueue != nil {
		sm.taskQueue.Stop()
	}

	sm.wg.Wait()

	sm.stateMu.Lock()
	sm.state.Status = SchedulerStatusStopped
	sm.state.LastUpdated = time.Now()
	sm.stateMu.Unlock()

	sm.started = false
	return nil
}

// ScheduleTask schedules a new task
func (sm *SchedulerManager) ScheduleTask(task *Task) error {
	if !sm.started {
		return fmt.Errorf("scheduler not started")
	}

	// Validate task
	if err := sm.validateTask(task); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Set task metadata
	task.CreatedAt = time.Now()
	task.Status = TaskStatusPending

	// Add to queue
	if err := sm.taskQueue.Enqueue(task); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Update metrics
	sm.metrics.mu.Lock()
	sm.metrics.TasksScheduled++
	sm.metrics.LastUpdated = time.Now()
	sm.metrics.mu.Unlock()

	// Update state
	sm.stateMu.Lock()
	sm.state.TotalTasks++
	sm.state.QueuedTasks++
	sm.state.LastUpdated = time.Now()
	sm.stateMu.Unlock()

	return nil
}

// GetState returns the current scheduler state
func (sm *SchedulerManager) GetState() *SchedulerState {
	sm.stateMu.RLock()
	defer sm.stateMu.RUnlock()

	// Create a copy
	state := *sm.state
	return &state
}

// GetMetrics returns the current scheduler metrics
func (sm *SchedulerManager) GetMetrics() *SchedulerMetrics {
	sm.metrics.mu.RLock()
	defer sm.metrics.mu.RUnlock()

	// Create a copy
	metrics := *sm.metrics
	return &metrics
}

// IsLeader returns whether this node is the scheduler leader
func (sm *SchedulerManager) IsLeader() bool {
	if sm.consensusManager == nil {
		return true // Single node mode
	}

	return sm.consensusManager.IsLeader()
}

// validateTask validates a task before scheduling
func (sm *SchedulerManager) validateTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	if task.Type == "" {
		return fmt.Errorf("task type is required")
	}

	if task.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	if task.Requirements != nil {
		if task.Requirements.CPU < 0 {
			return fmt.Errorf("CPU requirement cannot be negative")
		}
		if task.Requirements.Memory < 0 {
			return fmt.Errorf("memory requirement cannot be negative")
		}
	}

	return nil
}

// schedulerLoop runs the main scheduler loop
func (sm *SchedulerManager) schedulerLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.processScheduling()
		}
	}
}

// processScheduling processes pending tasks
func (sm *SchedulerManager) processScheduling() {
	// Only process if we're the leader
	if !sm.IsLeader() {
		return
	}

	// Get next task from queue
	task, err := sm.taskQueue.Dequeue()
	if err != nil {
		return // No tasks available
	}

	// Find suitable worker
	worker, err := sm.loadBalancer.SelectWorker(task)
	if err != nil {
		// No suitable worker, put task back in queue
		sm.taskQueue.Enqueue(task)
		return
	}

	// Assign task to worker
	if err := sm.assignTask(task, worker); err != nil {
		// Assignment failed, put task back in queue
		sm.taskQueue.Enqueue(task)
		return
	}

	// Track the task
	sm.taskTracker.TrackTask(task, worker)

	// Update metrics and state
	sm.updateSchedulingMetrics(task, worker)
}

// assignTask assigns a task to a worker
func (sm *SchedulerManager) assignTask(task *Task, worker *WorkerNode) error {
	// Update task
	task.AssignedWorker = worker.ID
	task.AssignedNode = worker.Address
	task.ScheduledAt = time.Now()
	task.Status = TaskStatusScheduled

	// Send task assignment message
	if sm.config.EnableP2PMessaging {
		return sm.sendTaskAssignment(task, worker)
	}

	// Fallback to direct assignment
	return sm.directTaskAssignment(task, worker)
}

// monitoringLoop runs the monitoring loop
func (sm *SchedulerManager) monitoringLoop() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.collectMetrics()
		}
	}
}

// collectMetrics collects and updates metrics
func (sm *SchedulerManager) collectMetrics() {
	// Update scheduler metrics
	sm.updateSchedulerMetrics()

	// Report to network monitor if available
	if sm.networkMonitor != nil {
		sm.reportToNetworkMonitor()
	}
}

// updateSchedulingMetrics updates metrics after scheduling a task
func (sm *SchedulerManager) updateSchedulingMetrics(task *Task, worker *WorkerNode) {
	sm.stateMu.Lock()
	defer sm.stateMu.Unlock()

	sm.state.QueuedTasks--
	sm.state.RunningTasks++
	sm.state.LastUpdated = time.Now()
}

// updateSchedulerMetrics updates overall scheduler metrics
func (sm *SchedulerManager) updateSchedulerMetrics() {
	// Get metrics from components
	queueMetrics := sm.taskQueue.GetMetrics()
	workerMetrics := sm.workerManager.GetMetrics()
	loadBalancerMetrics := sm.loadBalancer.GetMetrics()
	taskMetrics := sm.taskTracker.GetMetrics()

	// Update scheduler metrics
	sm.metrics.mu.Lock()
	defer sm.metrics.mu.Unlock()

	sm.metrics.QueueUtilization = float64(queueMetrics.CurrentSize) / float64(sm.config.MaxQueueSize)
	sm.metrics.WorkerUtilization = float64(workerMetrics.ActiveWorkers) / float64(workerMetrics.TotalWorkers)
	sm.metrics.AverageLatency = loadBalancerMetrics.AverageSelectionTime
	sm.metrics.ThroughputPerSecond = float64(taskMetrics.CompletedTasks) / time.Since(sm.metrics.LastUpdated).Seconds()
	sm.metrics.LastUpdated = time.Now()

	// Update state
	sm.stateMu.Lock()
	sm.state.TotalWorkers = workerMetrics.TotalWorkers
	sm.state.ActiveWorkers = workerMetrics.ActiveWorkers
	sm.state.IdleWorkers = workerMetrics.IdleWorkers
	sm.state.RunningTasks = taskMetrics.ActiveTasks
	sm.state.CompletedTasks = taskMetrics.CompletedTasks
	sm.state.FailedTasks = taskMetrics.FailedTasks
	sm.state.AverageTaskTime = taskMetrics.AverageExecutionTime
	sm.state.TaskThroughput = sm.metrics.ThroughputPerSecond
	sm.state.SystemLoad = workerMetrics.AverageLoad
	sm.state.LastUpdated = time.Now()
	sm.stateMu.Unlock()
}

// reportToNetworkMonitor reports metrics to the network monitor
func (sm *SchedulerManager) reportToNetworkMonitor() {
	// This would integrate with the network monitor to report scheduler metrics
	// Implementation would depend on the network monitor interface
}

// sendTaskAssignment sends a task assignment message via P2P
func (sm *SchedulerManager) sendTaskAssignment(task *Task, worker *WorkerNode) error {
	// Create task assignment message
	// This would use the messaging system to send the task to the worker
	return fmt.Errorf("P2P task assignment not implemented yet")
}

// directTaskAssignment assigns a task directly (fallback method)
func (sm *SchedulerManager) directTaskAssignment(task *Task, worker *WorkerNode) error {
	// Direct assignment logic (placeholder)
	return nil
}

// Message handlers for P2P communication

func (sm *SchedulerManager) handleTaskAssignment(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle incoming task assignment messages
	return nil
}

func (sm *SchedulerManager) handleTaskResult(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle task result messages
	return nil
}

func (sm *SchedulerManager) handleTaskStatus(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle task status update messages
	return nil
}

func (sm *SchedulerManager) handleResourceUpdate(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle resource update messages
	return nil
}

func (sm *SchedulerManager) handleWorkerRegister(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle worker registration messages
	return nil
}

func (sm *SchedulerManager) handleWorkerHeartbeat(ctx context.Context, msg *messaging.SchedulerMessage) error {
	// Handle worker heartbeat messages
	return nil
}
