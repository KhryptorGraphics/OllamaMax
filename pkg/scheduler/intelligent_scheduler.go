package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/pkg/loadbalancer"
	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// IntelligentScheduler provides advanced scheduling with ML-based optimization
type IntelligentScheduler struct {
	// Fine-grained locking for better concurrency
	taskQueueMu    sync.RWMutex // Lock for task queue operations
	nodeStateMu    sync.RWMutex // Lock for node state updates
	runningTasksMu sync.RWMutex // Lock for running tasks map
	historyMu      sync.RWMutex // Lock for task history

	// Core components
	config    *IntelligentSchedulerConfig
	p2p       *p2p.Node
	consensus *consensus.Engine
	logger    *slog.Logger

	// Enhanced scheduling components
	loadBalancer      *loadbalancer.IntelligentLoadBalancer
	resourcePredictor *ResourcePredictor
	taskAnalyzer      *TaskAnalyzer
	performanceModel  *PerformanceModel
	scalingManager    *DynamicScalingManager

	// Task management
	taskQueue      *PriorityTaskQueue
	runningTasks   map[string]*ScheduledTask
	completedTasks map[string]*TaskResult
	taskHistory    []*TaskExecutionRecord

	// Node management
	nodeManager     *IntelligentNodeManager
	nodePerformance map[string]*NodePerformanceProfile
	clusterTopology *ClusterTopology

	// Optimization
	optimizer      *SchedulingOptimizer
	adaptiveParams *AdaptiveSchedulingParams

	// Metrics and monitoring
	metrics            *IntelligentSchedulerMetrics
	performanceTracker *PerformanceTracker

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	workers []*SchedulingWorker
	started bool
}

// IntelligentSchedulerConfig holds configuration for the intelligent scheduler
type IntelligentSchedulerConfig struct {
	*config.SchedulerConfig

	// Machine Learning features
	EnableMLOptimization     bool `json:"enable_ml_optimization"`
	EnablePredictiveScaling  bool `json:"enable_predictive_scaling"`
	EnableAdaptiveScheduling bool `json:"enable_adaptive_scheduling"`

	// Performance tuning
	OptimizationInterval time.Duration `json:"optimization_interval"`
	PredictionWindow     time.Duration `json:"prediction_window"`
	AdaptationRate       float64       `json:"adaptation_rate"`

	// Resource management
	ResourcePredictionDepth int     `json:"resource_prediction_depth"`
	ScalingThreshold        float64 `json:"scaling_threshold"`
	MaxScalingFactor        float64 `json:"max_scaling_factor"`

	// Task prioritization
	PriorityLevels    int     `json:"priority_levels"`
	PriorityDecayRate float64 `json:"priority_decay_rate"`
	DeadlineWeight    float64 `json:"deadline_weight"`
}

// ScheduledTask represents a task in the scheduling system
type ScheduledTask struct {
	ID          string                     `json:"id"`
	Type        string                     `json:"type"`
	Priority    int                        `json:"priority"`
	Deadline    time.Time                  `json:"deadline"`
	ResourceReq *types.ResourceRequirement `json:"resource_requirements"`
	Constraints *TaskConstraints           `json:"constraints"`
	Metadata    map[string]interface{}     `json:"metadata"`

	// Scheduling information
	ScheduledAt      time.Time     `json:"scheduled_at"`
	AssignedNode     string        `json:"assigned_node"`
	EstimatedRuntime time.Duration `json:"estimated_runtime"`
	ActualRuntime    time.Duration `json:"actual_runtime"`

	// Status tracking
	Status      TaskStatus `json:"status"`
	Progress    float64    `json:"progress"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt time.Time  `json:"completed_at"`

	// Performance data
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	NetworkIO   int64   `json:"network_io"`
	DiskIO      int64   `json:"disk_io"`
}

// TaskConstraints defines constraints for task scheduling
type TaskConstraints struct {
	RequiredCapabilities []string             `json:"required_capabilities"`
	PreferredNodes       []string             `json:"preferred_nodes"`
	ExcludedNodes        []string             `json:"excluded_nodes"`
	AffinityRules        []*AffinityRule      `json:"affinity_rules"`
	AntiAffinityRules    []*AntiAffinityRule  `json:"anti_affinity_rules"`
	ResourceLimits       *ResourceLimits      `json:"resource_limits"`
	NetworkRequirements  *NetworkRequirements `json:"network_requirements"`
}

// AffinityRule defines task affinity preferences
type AffinityRule struct {
	Type     string            `json:"type"`   // "node", "task", "zone"
	Target   string            `json:"target"` // Target identifier
	Weight   float64           `json:"weight"` // Preference weight (0-1)
	Metadata map[string]string `json:"metadata"`
}

// AntiAffinityRule defines task anti-affinity preferences
type AntiAffinityRule struct {
	Type     string            `json:"type"`     // "node", "task", "zone"
	Target   string            `json:"target"`   // Target identifier
	Strength string            `json:"strength"` // "required", "preferred"
	Metadata map[string]string `json:"metadata"`
}

// ResourceLimits defines resource usage limits
type ResourceLimits struct {
	MaxCPU     float64 `json:"max_cpu"`
	MaxMemory  int64   `json:"max_memory"`
	MaxDisk    int64   `json:"max_disk"`
	MaxNetwork int64   `json:"max_network"`
}

// NetworkRequirements defines network-related requirements
type NetworkRequirements struct {
	MinBandwidth    int64             `json:"min_bandwidth"`
	MaxLatency      time.Duration     `json:"max_latency"`
	RequiredPorts   []int             `json:"required_ports"`
	SecurityGroups  []string          `json:"security_groups"`
	NetworkPolicies map[string]string `json:"network_policies"`
}

// TaskResult represents the result of task execution
type TaskResult struct {
	TaskID          string                 `json:"task_id"`
	Success         bool                   `json:"success"`
	Result          interface{}            `json:"result"`
	Error           string                 `json:"error,omitempty"`
	ExecutionTime   time.Duration          `json:"execution_time"`
	ResourceUsage   *ResourceUsageStats    `json:"resource_usage"`
	PerformanceData map[string]interface{} `json:"performance_data"`
	CompletedAt     time.Time              `json:"completed_at"`
}

// ResourceUsageStats tracks resource usage during task execution
type ResourceUsageStats struct {
	PeakCPU    float64 `json:"peak_cpu"`
	AvgCPU     float64 `json:"avg_cpu"`
	PeakMemory int64   `json:"peak_memory"`
	AvgMemory  int64   `json:"avg_memory"`
	NetworkIn  int64   `json:"network_in"`
	NetworkOut int64   `json:"network_out"`
	DiskRead   int64   `json:"disk_read"`
	DiskWrite  int64   `json:"disk_write"`
}

// TaskExecutionRecord records historical task execution data
type TaskExecutionRecord struct {
	TaskID           string                 `json:"task_id"`
	TaskType         string                 `json:"task_type"`
	NodeID           string                 `json:"node_id"`
	ScheduledAt      time.Time              `json:"scheduled_at"`
	StartedAt        time.Time              `json:"started_at"`
	CompletedAt      time.Time              `json:"completed_at"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	QueueTime        time.Duration          `json:"queue_time"`
	Success          bool                   `json:"success"`
	ResourceUsage    *ResourceUsageStats    `json:"resource_usage"`
	PerformanceScore float64                `json:"performance_score"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NodePerformanceProfile tracks performance characteristics of nodes
type NodePerformanceProfile struct {
	NodeID              string                   `json:"node_id"`
	TaskTypePerformance map[string]*TaskTypePerf `json:"task_type_performance"`
	ResourceEfficiency  *ResourceEfficiencyStats `json:"resource_efficiency"`
	ReliabilityScore    float64                  `json:"reliability_score"`
	AvailabilityScore   float64                  `json:"availability_score"`
	LastUpdated         time.Time                `json:"last_updated"`
}

// TaskTypePerf tracks performance for specific task types
type TaskTypePerf struct {
	TaskType         string        `json:"task_type"`
	AvgExecutionTime time.Duration `json:"avg_execution_time"`
	SuccessRate      float64       `json:"success_rate"`
	ThroughputScore  float64       `json:"throughput_score"`
	LatencyScore     float64       `json:"latency_score"`
	SampleCount      int           `json:"sample_count"`
}

// ResourceEfficiencyStats tracks resource utilization efficiency
type ResourceEfficiencyStats struct {
	CPUEfficiency     float64 `json:"cpu_efficiency"`
	MemoryEfficiency  float64 `json:"memory_efficiency"`
	NetworkEfficiency float64 `json:"network_efficiency"`
	DiskEfficiency    float64 `json:"disk_efficiency"`
	OverallEfficiency float64 `json:"overall_efficiency"`
}

// ClusterTopology represents the cluster network topology
type ClusterTopology struct {
	Nodes        map[string]*TopologyNode `json:"nodes"`
	Zones        map[string]*Zone         `json:"zones"`
	NetworkLinks []*NetworkLink           `json:"network_links"`
	LastUpdated  time.Time                `json:"last_updated"`
}

// TopologyNode represents a node in the cluster topology
type TopologyNode struct {
	ID           string            `json:"id"`
	Zone         string            `json:"zone"`
	Capabilities []string          `json:"capabilities"`
	Coordinates  *NodeCoordinates  `json:"coordinates"`
	Neighbors    []string          `json:"neighbors"`
	Metadata     map[string]string `json:"metadata"`
}

// Zone represents a logical zone in the cluster
type Zone struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Nodes       []string            `json:"nodes"`
	Capacity    *types.NodeCapacity `json:"capacity"`
	Utilization float64             `json:"utilization"`
}

// NetworkLink represents a network connection between nodes
type NetworkLink struct {
	SourceNode  string        `json:"source_node"`
	TargetNode  string        `json:"target_node"`
	Bandwidth   int64         `json:"bandwidth"`
	Latency     time.Duration `json:"latency"`
	Reliability float64       `json:"reliability"`
	Cost        float64       `json:"cost"`
}

// NodeCoordinates represents the logical coordinates of a node
type NodeCoordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// IntelligentSchedulerMetrics tracks scheduler performance metrics
type IntelligentSchedulerMetrics struct {
	TotalTasksScheduled    int64                    `json:"total_tasks_scheduled"`
	SuccessfulTasks        int64                    `json:"successful_tasks"`
	FailedTasks            int64                    `json:"failed_tasks"`
	AvgSchedulingTime      time.Duration            `json:"avg_scheduling_time"`
	AvgExecutionTime       time.Duration            `json:"avg_execution_time"`
	AvgQueueTime           time.Duration            `json:"avg_queue_time"`
	ResourceUtilization    map[string]float64       `json:"resource_utilization"`
	NodeUtilization        map[string]float64       `json:"node_utilization"`
	TaskTypeDistribution   map[string]int64         `json:"task_type_distribution"`
	SchedulingAlgorithmUse map[string]int64         `json:"scheduling_algorithm_use"`
	OptimizationImpact     *OptimizationImpactStats `json:"optimization_impact"`
	LastUpdated            time.Time                `json:"last_updated"`
}

// OptimizationImpactStats tracks the impact of optimization
type OptimizationImpactStats struct {
	PerformanceImprovement float64 `json:"performance_improvement"`
	ResourceSavings        float64 `json:"resource_savings"`
	CostReduction          float64 `json:"cost_reduction"`
	LatencyReduction       float64 `json:"latency_reduction"`
	ThroughputIncrease     float64 `json:"throughput_increase"`
}

// NewIntelligentScheduler creates a new intelligent scheduler
func NewIntelligentScheduler(
	config *IntelligentSchedulerConfig,
	p2pNode *p2p.Node,
	consensusEngine *consensus.Engine,
	logger *slog.Logger,
) *IntelligentScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	is := &IntelligentScheduler{
		config:          config,
		p2p:             p2pNode,
		consensus:       consensusEngine,
		logger:          logger,
		runningTasks:    make(map[string]*ScheduledTask),
		completedTasks:  make(map[string]*TaskResult),
		taskHistory:     make([]*TaskExecutionRecord, 0),
		nodePerformance: make(map[string]*NodePerformanceProfile),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize components
	is.initializeComponents()

	return is
}

// initializeComponents initializes all scheduler components
func (is *IntelligentScheduler) initializeComponents() {
	// Initialize load balancer
	lbConfig := &loadbalancer.Config{
		Algorithm:           is.config.LoadBalancing,
		HealthCheckInterval: is.config.HealthCheckInterval,
		MaxRetries:          3,
		Timeout:             30 * time.Second,
	}
	is.loadBalancer = loadbalancer.NewIntelligentLoadBalancer(lbConfig)

	// Initialize resource predictor
	is.resourcePredictor = NewResourcePredictor(is.config, is.logger)

	// Initialize task analyzer
	is.taskAnalyzer = NewTaskAnalyzer(is.config, is.logger)

	// Initialize performance model
	is.performanceModel = NewPerformanceModel(is.config, is.logger)

	// Initialize scaling manager
	is.scalingManager = NewDynamicScalingManager(is.config, is.logger)

	// Initialize task queue
	is.taskQueue = NewPriorityTaskQueue(is.config.QueueSize)

	// Initialize node manager
	is.nodeManager = NewIntelligentNodeManager(is.config, is.p2p, is.logger)

	// Initialize cluster topology
	is.clusterTopology = &ClusterTopology{
		Nodes:        make(map[string]*TopologyNode),
		Zones:        make(map[string]*Zone),
		NetworkLinks: make([]*NetworkLink, 0),
		LastUpdated:  time.Now(),
	}

	// Initialize optimizer
	is.optimizer = NewSchedulingOptimizer(is.config, is.logger)

	// Initialize adaptive parameters
	is.adaptiveParams = &AdaptiveSchedulingParams{
		LearningRate:     is.config.AdaptationRate,
		OptimizationGoal: "balanced", // balanced, performance, efficiency
		LastUpdated:      time.Now(),
	}

	// Initialize metrics
	is.metrics = &IntelligentSchedulerMetrics{
		ResourceUtilization:    make(map[string]float64),
		NodeUtilization:        make(map[string]float64),
		TaskTypeDistribution:   make(map[string]int64),
		SchedulingAlgorithmUse: make(map[string]int64),
		OptimizationImpact:     &OptimizationImpactStats{},
		LastUpdated:            time.Now(),
	}

	// Initialize performance tracker
	is.performanceTracker = NewPerformanceTracker(is.config, is.logger)
}

// ScheduleTask schedules a task using intelligent algorithms
func (is *IntelligentScheduler) ScheduleTask(task *ScheduledTask) error {
	// Use fine-grained locking instead of global lock
	is.taskQueueMu.Lock()
	defer is.taskQueueMu.Unlock()

	startTime := time.Now()

	// Analyze task requirements
	analysis, err := is.taskAnalyzer.AnalyzeTask(task)
	if err != nil {
		return fmt.Errorf("task analysis failed: %w", err)
	}

	// Predict resource requirements
	prediction, err := is.resourcePredictor.PredictRequirements(task, analysis)
	if err != nil {
		return fmt.Errorf("resource prediction failed: %w", err)
	}

	// Get available nodes
	availableNodes := is.nodeManager.GetAvailableNodes()
	if len(availableNodes) == 0 {
		return fmt.Errorf("no available nodes")
	}

	// Apply constraints and filters
	candidateNodes := is.applyConstraints(task, availableNodes)
	if len(candidateNodes) == 0 {
		return fmt.Errorf("no nodes satisfy task constraints")
	}

	// Select optimal node using intelligent load balancing
	selectedNode, err := is.selectOptimalNode(task, candidateNodes, prediction)
	if err != nil {
		return fmt.Errorf("node selection failed: %w", err)
	}

	// Assign task to selected node
	task.AssignedNode = selectedNode.ID
	task.ScheduledAt = time.Now()
	task.Status = TaskStatusScheduled
	task.EstimatedRuntime = prediction.EstimatedRuntime

	// Add to running tasks
	is.runningTasks[task.ID] = task

	// Update metrics
	is.updateSchedulingMetrics(task, selectedNode, time.Since(startTime))

	// Start task execution monitoring
	go is.monitorTaskExecution(task)

	is.logger.Info("task scheduled successfully",
		"task_id", task.ID,
		"task_type", task.Type,
		"assigned_node", selectedNode.ID,
		"estimated_runtime", prediction.EstimatedRuntime,
		"scheduling_time", time.Since(startTime))

	return nil
}

// selectOptimalNode selects the best node for a task
func (is *IntelligentScheduler) selectOptimalNode(task *ScheduledTask, nodes []*IntelligentNode, prediction *ResourcePrediction) (*IntelligentNode, error) {
	// Convert to load balancer format
	lbNodes := make([]*loadbalancer.NodeInfo, len(nodes))
	for i, node := range nodes {
		lbNodes[i] = &loadbalancer.NodeInfo{
			ID:       node.ID,
			Address:  node.Address,
			Capacity: node.Capacity,
			Usage:    node.CurrentUsage,
			Metadata: node.Metadata,
		}
	}

	// Use intelligent load balancer to select nodes
	selectedLBNodes, err := is.loadBalancer.SelectNodes(task, lbNodes)
	if err != nil {
		return nil, err
	}

	if len(selectedLBNodes) == 0 {
		return nil, fmt.Errorf("no nodes selected by load balancer")
	}

	// Find the corresponding intelligent node
	selectedLBNode := selectedLBNodes[0] // Take the first (best) node
	for _, node := range nodes {
		if node.ID == selectedLBNode.ID {
			return node, nil
		}
	}

	return nil, fmt.Errorf("selected node not found in candidate list")
}

// applyConstraints filters nodes based on task constraints
func (is *IntelligentScheduler) applyConstraints(task *ScheduledTask, nodes []*IntelligentNode) []*IntelligentNode {
	if task.Constraints == nil {
		return nodes
	}

	var candidateNodes []*IntelligentNode

	for _, node := range nodes {
		if is.nodeMatchesConstraints(node, task.Constraints) {
			candidateNodes = append(candidateNodes, node)
		}
	}

	return candidateNodes
}

// nodeMatchesConstraints checks if a node satisfies task constraints
func (is *IntelligentScheduler) nodeMatchesConstraints(node *IntelligentNode, constraints *TaskConstraints) bool {
	// Check required capabilities
	for _, requiredCap := range constraints.RequiredCapabilities {
		if !node.HasCapability(requiredCap) {
			return false
		}
	}

	// Check excluded nodes
	for _, excludedNode := range constraints.ExcludedNodes {
		if node.ID == excludedNode {
			return false
		}
	}

	// Check resource limits
	if constraints.ResourceLimits != nil {
		if !node.CanAccommodateResources(constraints.ResourceLimits) {
			return false
		}
	}

	// Check network requirements
	if constraints.NetworkRequirements != nil {
		if !node.MeetsNetworkRequirements(constraints.NetworkRequirements) {
			return false
		}
	}

	return true
}

// monitorTaskExecution monitors task execution and updates performance data
func (is *IntelligentScheduler) monitorTaskExecution(task *ScheduledTask) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-is.ctx.Done():
			return
		case <-ticker.C:
			// Update task progress and resource usage
			is.updateTaskProgress(task)

			// Check if task is completed
			if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
				is.handleTaskCompletion(task)
				return
			}
		}
	}
}

// updateTaskProgress updates task progress and resource usage
func (is *IntelligentScheduler) updateTaskProgress(task *ScheduledTask) {
	// In a real implementation, this would query the node for actual progress
	// For now, we'll simulate progress based on elapsed time

	if task.Status != TaskStatusRunning {
		return
	}

	elapsed := time.Since(task.StartedAt)
	if task.EstimatedRuntime > 0 {
		task.Progress = math.Min(float64(elapsed)/float64(task.EstimatedRuntime), 1.0)
	}

	// Simulate resource usage (in real implementation, get from node)
	task.CPUUsage = 0.5 + 0.3*task.Progress // Simulate increasing CPU usage
	task.MemoryUsage = int64(float64(task.ResourceReq.Memory) * (0.3 + 0.7*task.Progress))
}

// handleTaskCompletion handles task completion and updates metrics
func (is *IntelligentScheduler) handleTaskCompletion(task *ScheduledTask) {
	is.mu.Lock()
	defer is.mu.Unlock()

	// Calculate actual runtime
	task.ActualRuntime = time.Since(task.StartedAt)
	task.CompletedAt = time.Now()

	// Create task result
	result := &TaskResult{
		TaskID:        task.ID,
		Success:       task.Status == TaskStatusCompleted,
		ExecutionTime: task.ActualRuntime,
		ResourceUsage: &ResourceUsageStats{
			PeakCPU:    task.CPUUsage,
			AvgCPU:     task.CPUUsage * 0.8, // Simulate average
			PeakMemory: task.MemoryUsage,
			AvgMemory:  int64(float64(task.MemoryUsage) * 0.7),
		},
		CompletedAt: task.CompletedAt,
	}

	if task.Status == TaskStatusFailed {
		result.Error = "Task execution failed"
	}

	// Store result
	is.completedTasks[task.ID] = result

	// Remove from running tasks
	delete(is.runningTasks, task.ID)

	// Record execution history
	is.recordTaskExecution(task, result)

	// Update node performance profile
	is.updateNodePerformance(task, result)

	// Update performance model
	is.performanceModel.UpdateModel(task, result)

	// Update metrics
	is.updateCompletionMetrics(task, result)

	is.logger.Info("task completed",
		"task_id", task.ID,
		"success", result.Success,
		"execution_time", result.ExecutionTime,
		"estimated_time", task.EstimatedRuntime)
}

// recordTaskExecution records task execution in history
func (is *IntelligentScheduler) recordTaskExecution(task *ScheduledTask, result *TaskResult) {
	record := &TaskExecutionRecord{
		TaskID:           task.ID,
		TaskType:         task.Type,
		NodeID:           task.AssignedNode,
		ScheduledAt:      task.ScheduledAt,
		StartedAt:        task.StartedAt,
		CompletedAt:      task.CompletedAt,
		ExecutionTime:    result.ExecutionTime,
		QueueTime:        task.StartedAt.Sub(task.ScheduledAt),
		Success:          result.Success,
		ResourceUsage:    result.ResourceUsage,
		PerformanceScore: is.calculatePerformanceScore(task, result),
		Metadata:         make(map[string]interface{}),
	}

	is.taskHistory = append(is.taskHistory, record)

	// Limit history size
	if len(is.taskHistory) > 10000 {
		is.taskHistory = is.taskHistory[1:]
	}
}

// updateNodePerformance updates node performance profile
func (is *IntelligentScheduler) updateNodePerformance(task *ScheduledTask, result *TaskResult) {
	profile, exists := is.nodePerformance[task.AssignedNode]
	if !exists {
		profile = &NodePerformanceProfile{
			NodeID:              task.AssignedNode,
			TaskTypePerformance: make(map[string]*TaskTypePerf),
			ResourceEfficiency:  &ResourceEfficiencyStats{},
			ReliabilityScore:    1.0,
			AvailabilityScore:   1.0,
			LastUpdated:         time.Now(),
		}
		is.nodePerformance[task.AssignedNode] = profile
	}

	// Update task type performance
	taskTypePerf, exists := profile.TaskTypePerformance[task.Type]
	if !exists {
		taskTypePerf = &TaskTypePerf{
			TaskType:         task.Type,
			AvgExecutionTime: result.ExecutionTime,
			SuccessRate:      1.0,
			ThroughputScore:  1.0,
			LatencyScore:     1.0,
			SampleCount:      1,
		}
		profile.TaskTypePerformance[task.Type] = taskTypePerf
	} else {
		// Update averages
		taskTypePerf.AvgExecutionTime = (taskTypePerf.AvgExecutionTime*time.Duration(taskTypePerf.SampleCount) + result.ExecutionTime) / time.Duration(taskTypePerf.SampleCount+1)
		taskTypePerf.SuccessRate = (taskTypePerf.SuccessRate*float64(taskTypePerf.SampleCount) + map[bool]float64{true: 1.0, false: 0.0}[result.Success]) / float64(taskTypePerf.SampleCount+1)
		taskTypePerf.SampleCount++
	}

	// Update reliability score
	if result.Success {
		profile.ReliabilityScore = (profile.ReliabilityScore*0.9 + 1.0*0.1)
	} else {
		profile.ReliabilityScore = (profile.ReliabilityScore*0.9 + 0.0*0.1)
	}

	profile.LastUpdated = time.Now()
}

// updateSchedulingMetrics updates scheduling metrics
func (is *IntelligentScheduler) updateSchedulingMetrics(task *ScheduledTask, node *IntelligentNode, schedulingTime time.Duration) {
	is.metrics.TotalTasksScheduled++
	is.metrics.AvgSchedulingTime = (is.metrics.AvgSchedulingTime*time.Duration(is.metrics.TotalTasksScheduled-1) + schedulingTime) / time.Duration(is.metrics.TotalTasksScheduled)
	is.metrics.TaskTypeDistribution[task.Type]++
	is.metrics.LastUpdated = time.Now()
}

// updateCompletionMetrics updates completion metrics
func (is *IntelligentScheduler) updateCompletionMetrics(task *ScheduledTask, result *TaskResult) {
	if result.Success {
		is.metrics.SuccessfulTasks++
	} else {
		is.metrics.FailedTasks++
	}

	totalTasks := is.metrics.SuccessfulTasks + is.metrics.FailedTasks
	if totalTasks > 0 {
		is.metrics.AvgExecutionTime = (is.metrics.AvgExecutionTime*time.Duration(totalTasks-1) + result.ExecutionTime) / time.Duration(totalTasks)
	}

	is.metrics.LastUpdated = time.Now()
}

// calculatePerformanceScore calculates a performance score for task execution
func (is *IntelligentScheduler) calculatePerformanceScore(task *ScheduledTask, result *TaskResult) float64 {
	score := 1.0

	// Factor in success
	if !result.Success {
		score *= 0.1
	}

	// Factor in execution time vs estimate
	if task.EstimatedRuntime > 0 {
		timeRatio := float64(result.ExecutionTime) / float64(task.EstimatedRuntime)
		if timeRatio <= 1.0 {
			score *= 1.0 // Perfect or better than expected
		} else {
			score *= 1.0 / timeRatio // Penalty for taking longer
		}
	}

	// Factor in resource efficiency
	if result.ResourceUsage != nil && task.ResourceReq != nil {
		if task.ResourceReq.Memory > 0 {
			memoryEfficiency := float64(result.ResourceUsage.AvgMemory) / float64(task.ResourceReq.Memory)
			score *= (0.5 + 0.5*memoryEfficiency) // Reward efficient memory usage
		}
	}

	return score
}

// GetMetrics returns current scheduler metrics
func (is *IntelligentScheduler) GetMetrics() *IntelligentSchedulerMetrics {
	is.mu.RLock()
	defer is.mu.RUnlock()

	return is.metrics
}

// GetRunningTasks returns currently running tasks
func (is *IntelligentScheduler) GetRunningTasks() map[string]*ScheduledTask {
	is.mu.RLock()
	defer is.mu.RUnlock()

	// Return a copy to avoid race conditions
	runningTasks := make(map[string]*ScheduledTask)
	for id, task := range is.runningTasks {
		runningTasks[id] = task
	}

	return runningTasks
}

// GetTaskHistory returns task execution history
func (is *IntelligentScheduler) GetTaskHistory(limit int) []*TaskExecutionRecord {
	is.mu.RLock()
	defer is.mu.RUnlock()

	if limit <= 0 || limit > len(is.taskHistory) {
		limit = len(is.taskHistory)
	}

	// Return the most recent entries
	start := len(is.taskHistory) - limit
	history := make([]*TaskExecutionRecord, limit)
	copy(history, is.taskHistory[start:])

	return history
}
