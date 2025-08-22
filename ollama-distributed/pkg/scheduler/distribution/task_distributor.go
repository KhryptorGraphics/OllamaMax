package distribution

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/types"
)

// TaskDistributor manages intelligent task distribution across nodes
type TaskDistributor struct {
	mu sync.RWMutex

	// Node management
	nodes          map[string]*NodeInfo
	nodeCapacities map[string]*types.NodeCapacity
	nodeMetrics    map[string]*types.ResourceMetrics

	// Task management
	pendingTasks   []*Task
	runningTasks   map[string]*Task
	completedTasks map[string]*Task

	// Distribution strategies
	strategies      map[string]DistributionStrategy
	defaultStrategy string

	// Configuration
	config *DistributorConfig

	// Metrics
	metrics *DistributionMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Task represents a task to be distributed and executed
type Task struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	Priority types.TaskPriority `json:"priority"`
	Status   types.TaskStatus   `json:"status"`

	// Resource requirements
	Requirements *types.ResourceRequirement `json:"requirements"`

	// Scheduling constraints
	Constraints *TaskConstraints `json:"constraints"`

	// Execution details
	AssignedNodeID string        `json:"assigned_node_id,omitempty"`
	StartTime      time.Time     `json:"start_time,omitempty"`
	EndTime        time.Time     `json:"end_time,omitempty"`
	Duration       time.Duration `json:"duration,omitempty"`

	// Task data
	Payload      map[string]interface{} `json:"payload"`
	Dependencies []string               `json:"dependencies"`

	// Retry and timeout
	MaxRetries int           `json:"max_retries"`
	RetryCount int           `json:"retry_count"`
	Timeout    time.Duration `json:"timeout"`

	// Metadata
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TaskConstraints represents scheduling constraints for a task
type TaskConstraints struct {
	// Node selection
	RequiredNodes  []string `json:"required_nodes,omitempty"`
	ExcludedNodes  []string `json:"excluded_nodes,omitempty"`
	PreferredNodes []string `json:"preferred_nodes,omitempty"`

	// Affinity rules
	NodeAffinity map[string]string `json:"node_affinity,omitempty"`
	TaskAffinity []string          `json:"task_affinity,omitempty"`
	AntiAffinity []string          `json:"anti_affinity,omitempty"`

	// Timing constraints
	EarliestStart time.Time `json:"earliest_start,omitempty"`
	LatestStart   time.Time `json:"latest_start,omitempty"`
	Deadline      time.Time `json:"deadline,omitempty"`

	// Resource constraints
	MinResources *types.ResourceRequirement `json:"min_resources,omitempty"`
	MaxResources *types.ResourceRequirement `json:"max_resources,omitempty"`

	// Quality constraints
	MinReliability float64       `json:"min_reliability,omitempty"`
	MaxLatency     time.Duration `json:"max_latency,omitempty"`
}

// NodeInfo represents information about a node
type NodeInfo struct {
	NodeID   string     `json:"node_id"`
	Status   NodeStatus `json:"status"`
	LastSeen time.Time  `json:"last_seen"`

	// Capabilities
	Capacity *types.NodeCapacity `json:"capacity"`
	Features []string            `json:"features"`

	// Performance characteristics
	Reliability    float64       `json:"reliability"`
	AverageLatency time.Duration `json:"average_latency"`
	Throughput     float64       `json:"throughput"`

	// Current load
	RunningTasks int     `json:"running_tasks"`
	QueuedTasks  int     `json:"queued_tasks"`
	LoadScore    float64 `json:"load_score"`

	// Health
	HealthScore     float64   `json:"health_score"`
	LastHealthCheck time.Time `json:"last_health_check"`
}

// NodeStatus represents the status of a node
type NodeStatus string

const (
	NodeStatusActive      NodeStatus = "active"
	NodeStatusIdle        NodeStatus = "idle"
	NodeStatusBusy        NodeStatus = "busy"
	NodeStatusDraining    NodeStatus = "draining"
	NodeStatusUnavailable NodeStatus = "unavailable"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// DistributionStrategy defines how tasks are distributed to nodes
type DistributionStrategy interface {
	Name() string
	SelectNode(task *Task, availableNodes []*NodeInfo) (*NodeInfo, error)
	CalculateScore(task *Task, node *NodeInfo) float64
}

// DistributorConfig configures the task distributor
type DistributorConfig struct {
	// Distribution settings
	DefaultStrategy    string
	MaxConcurrentTasks int
	TaskTimeout        time.Duration
	RetryDelay         time.Duration

	// Node management
	NodeHealthCheckInterval time.Duration
	NodeTimeout             time.Duration
	MaxUnhealthyNodes       int

	// Performance settings
	DistributionInterval  time.Duration
	MetricsUpdateInterval time.Duration
	CleanupInterval       time.Duration

	// Quality settings
	MinNodeReliability   float64
	MaxNodeLatency       time.Duration
	LoadBalanceThreshold float64
}

// DistributionMetrics tracks distribution performance
type DistributionMetrics struct {
	// Task metrics
	TotalTasks     int64 `json:"total_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`
	RetryTasks     int64 `json:"retry_tasks"`

	// Distribution metrics
	AverageDistributionTime time.Duration `json:"average_distribution_time"`
	AverageExecutionTime    time.Duration `json:"average_execution_time"`
	TaskThroughput          float64       `json:"task_throughput"`

	// Node metrics
	ActiveNodes     int     `json:"active_nodes"`
	AverageNodeLoad float64 `json:"average_node_load"`
	LoadImbalance   float64 `json:"load_imbalance"`

	// Strategy metrics
	StrategyUsage   map[string]int64   `json:"strategy_usage"`
	StrategySuccess map[string]float64 `json:"strategy_success"`

	// Quality metrics
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	SLACompliance  float64       `json:"sla_compliance"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewTaskDistributor creates a new task distributor
func NewTaskDistributor(config *DistributorConfig) *TaskDistributor {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &DistributorConfig{
			DefaultStrategy:         "resource_aware",
			MaxConcurrentTasks:      1000,
			TaskTimeout:             30 * time.Minute,
			RetryDelay:              5 * time.Second,
			NodeHealthCheckInterval: 30 * time.Second,
			NodeTimeout:             5 * time.Minute,
			MaxUnhealthyNodes:       3,
			DistributionInterval:    time.Second,
			MetricsUpdateInterval:   10 * time.Second,
			CleanupInterval:         5 * time.Minute,
			MinNodeReliability:      0.8,
			MaxNodeLatency:          100 * time.Millisecond,
			LoadBalanceThreshold:    0.8,
		}
	}

	td := &TaskDistributor{
		nodes:           make(map[string]*NodeInfo),
		nodeCapacities:  make(map[string]*types.NodeCapacity),
		nodeMetrics:     make(map[string]*types.ResourceMetrics),
		pendingTasks:    make([]*Task, 0),
		runningTasks:    make(map[string]*Task),
		completedTasks:  make(map[string]*Task),
		strategies:      make(map[string]DistributionStrategy),
		defaultStrategy: config.DefaultStrategy,
		config:          config,
		metrics: &DistributionMetrics{
			StrategyUsage:   make(map[string]int64),
			StrategySuccess: make(map[string]float64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Register default strategies
	td.registerDefaultStrategies()

	// Start background tasks
	td.wg.Add(3)
	go td.distributionLoop()
	go td.metricsLoop()
	go td.cleanupLoop()

	return td
}

// SubmitTask submits a new task for distribution
func (td *TaskDistributor) SubmitTask(task *Task) error {
	td.mu.Lock()
	defer td.mu.Unlock()

	// Validate task
	if err := td.validateTask(task); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	// Set initial status and timestamps
	task.Status = types.TaskStatusPending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Add to pending queue
	td.pendingTasks = append(td.pendingTasks, task)

	// Sort by priority
	td.sortPendingTasks()

	td.metrics.TotalTasks++

	return nil
}

// RegisterNode registers a new node for task distribution
func (td *TaskDistributor) RegisterNode(nodeInfo *NodeInfo) {
	td.mu.Lock()
	defer td.mu.Unlock()

	nodeInfo.LastSeen = time.Now()
	nodeInfo.LastHealthCheck = time.Now()

	td.nodes[nodeInfo.NodeID] = nodeInfo

	if nodeInfo.Capacity != nil {
		td.nodeCapacities[nodeInfo.NodeID] = nodeInfo.Capacity
	}
}

// UpdateNodeMetrics updates resource metrics for a node
func (td *TaskDistributor) UpdateNodeMetrics(nodeID string, metrics *types.ResourceMetrics) {
	td.mu.Lock()
	defer td.mu.Unlock()

	td.nodeMetrics[nodeID] = metrics

	// Update node info
	if node, exists := td.nodes[nodeID]; exists {
		node.LastSeen = time.Now()
		td.updateNodeLoadScore(node, metrics)
	}
}

// updateNodeLoadScore updates the load score for a node
func (td *TaskDistributor) updateNodeLoadScore(node *NodeInfo, metrics *types.ResourceMetrics) {
	// Calculate load score based on resource utilization
	cpuLoad := metrics.CPUUsagePercent / 100.0
	memoryLoad := metrics.MemoryUsagePercent / 100.0
	diskLoad := metrics.DiskUsagePercent / 100.0

	// Weighted average (CPU and memory are more important)
	node.LoadScore = (cpuLoad*0.4 + memoryLoad*0.4 + diskLoad*0.2)

	// Update node status based on load
	if node.LoadScore < 0.3 {
		node.Status = NodeStatusIdle
	} else if node.LoadScore < 0.8 {
		node.Status = NodeStatusActive
	} else {
		node.Status = NodeStatusBusy
	}
}

// DistributeTasks distributes pending tasks to available nodes
func (td *TaskDistributor) DistributeTasks() error {
	td.mu.Lock()
	defer td.mu.Unlock()

	if len(td.pendingTasks) == 0 {
		return nil
	}

	// Get available nodes
	availableNodes := td.getAvailableNodes()
	if len(availableNodes) == 0 {
		return fmt.Errorf("no available nodes for task distribution")
	}

	// Distribute tasks
	distributed := 0
	for i := len(td.pendingTasks) - 1; i >= 0; i-- {
		task := td.pendingTasks[i]

		// Check if we can distribute more tasks
		if len(td.runningTasks) >= td.config.MaxConcurrentTasks {
			break
		}

		// Select strategy
		strategy, err := td.getStrategy(task)
		if err != nil {
			continue // Try next task
		}

		// Select node
		selectedNode, err := strategy.SelectNode(task, availableNodes)
		if err != nil {
			continue // Try next task
		}

		// Assign task to node
		if err := td.assignTaskToNode(task, selectedNode); err != nil {
			continue // Try next task
		}

		// Remove from pending queue
		td.pendingTasks = append(td.pendingTasks[:i], td.pendingTasks[i+1:]...)
		distributed++

		// Update metrics
		td.metrics.StrategyUsage[strategy.Name()]++
	}

	return nil
}

// assignTaskToNode assigns a task to a specific node
func (td *TaskDistributor) assignTaskToNode(task *Task, node *NodeInfo) error {
	// Check resource availability
	if !td.hasAvailableResources(node, task.Requirements) {
		return fmt.Errorf("insufficient resources on node %s", node.NodeID)
	}

	// Update task
	task.Status = types.TaskStatusScheduled
	task.AssignedNodeID = node.NodeID
	task.StartTime = time.Now()
	task.UpdatedAt = time.Now()

	// Add to running tasks
	td.runningTasks[task.ID] = task

	// Update node info
	node.RunningTasks++

	// Reserve resources (simplified)
	if node.Capacity != nil && task.Requirements != nil {
		node.Capacity.AvailableCPUCores -= task.Requirements.CPUCores
		node.Capacity.AvailableMemoryBytes -= task.Requirements.MemoryBytes
		node.Capacity.AvailableDiskBytes -= task.Requirements.DiskBytes
	}

	return nil
}

// hasAvailableResources checks if a node has available resources for a task
func (td *TaskDistributor) hasAvailableResources(node *NodeInfo, requirements *types.ResourceRequirement) bool {
	if node.Capacity == nil || requirements == nil {
		return true // Assume available if no capacity info
	}

	return node.Capacity.AvailableCPUCores >= requirements.CPUCores &&
		node.Capacity.AvailableMemoryBytes >= requirements.MemoryBytes &&
		node.Capacity.AvailableDiskBytes >= requirements.DiskBytes
}

// getAvailableNodes returns nodes that are available for task assignment
func (td *TaskDistributor) getAvailableNodes() []*NodeInfo {
	var available []*NodeInfo

	for _, node := range td.nodes {
		if td.isNodeAvailable(node) {
			available = append(available, node)
		}
	}

	return available
}

// isNodeAvailable checks if a node is available for task assignment
func (td *TaskDistributor) isNodeAvailable(node *NodeInfo) bool {
	// Check node status
	if node.Status == NodeStatusUnavailable ||
		node.Status == NodeStatusMaintenance ||
		node.Status == NodeStatusDraining {
		return false
	}

	// Check if node is responsive
	if time.Since(node.LastSeen) > td.config.NodeTimeout {
		return false
	}

	// Check reliability
	if node.Reliability < td.config.MinNodeReliability {
		return false
	}

	// Check latency
	if node.AverageLatency > td.config.MaxNodeLatency {
		return false
	}

	// Check load
	if node.LoadScore > td.config.LoadBalanceThreshold {
		return false
	}

	return true
}

// getStrategy returns the appropriate distribution strategy for a task
func (td *TaskDistributor) getStrategy(task *Task) (DistributionStrategy, error) {
	// Check if task specifies a strategy
	if strategyName, exists := task.Metadata["strategy"]; exists {
		if strategy, found := td.strategies[strategyName.(string)]; found {
			return strategy, nil
		}
	}

	// Use default strategy
	if strategy, exists := td.strategies[td.defaultStrategy]; exists {
		return strategy, nil
	}

	// Fallback to first available strategy
	for _, strategy := range td.strategies {
		return strategy, nil
	}

	// Return error instead of panic for better error handling
	return nil, fmt.Errorf("no distribution strategies available - this indicates a configuration error")
}

// validateTask validates a task before submission
func (td *TaskDistributor) validateTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("task ID is required")
	}

	if task.Name == "" {
		return fmt.Errorf("task name is required")
	}

	if task.Type == "" {
		return fmt.Errorf("task type is required")
	}

	// Check for duplicate task ID
	if _, exists := td.runningTasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	if _, exists := td.completedTasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already completed", task.ID)
	}

	return nil
}

// sortPendingTasks sorts pending tasks by priority and creation time
func (td *TaskDistributor) sortPendingTasks() {
	sort.Slice(td.pendingTasks, func(i, j int) bool {
		taskI, taskJ := td.pendingTasks[i], td.pendingTasks[j]

		// First sort by priority
		priorityI := td.getPriorityValue(taskI.Priority)
		priorityJ := td.getPriorityValue(taskJ.Priority)

		if priorityI != priorityJ {
			return priorityI > priorityJ // Higher priority first
		}

		// Then by creation time (older first)
		return taskI.CreatedAt.Before(taskJ.CreatedAt)
	})
}

// getPriorityValue returns numeric value for priority comparison
func (td *TaskDistributor) getPriorityValue(priority types.TaskPriority) int {
	switch priority {
	case types.TaskPriorityUrgent:
		return 5
	case types.TaskPriorityCritical:
		return 4
	case types.TaskPriorityHigh:
		return 3
	case types.TaskPriorityNormal:
		return 2
	case types.TaskPriorityLow:
		return 1
	default:
		return 0
	}
}

// registerDefaultStrategies registers default distribution strategies
func (td *TaskDistributor) registerDefaultStrategies() {
	// Register strategies (implementations will be added in separate files)
	td.strategies["round_robin"] = &RoundRobinStrategy{}
	td.strategies["least_loaded"] = &LeastLoadedStrategy{}
	td.strategies["resource_aware"] = &ResourceAwareStrategy{}
	td.strategies["latency_based"] = &LatencyBasedStrategy{}
}

// distributionLoop periodically distributes pending tasks
func (td *TaskDistributor) distributionLoop() {
	defer td.wg.Done()

	ticker := time.NewTicker(td.config.DistributionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-td.ctx.Done():
			return
		case <-ticker.C:
			td.DistributeTasks()
		}
	}
}

// metricsLoop periodically updates metrics
func (td *TaskDistributor) metricsLoop() {
	defer td.wg.Done()

	ticker := time.NewTicker(td.config.MetricsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-td.ctx.Done():
			return
		case <-ticker.C:
			td.updateMetrics()
		}
	}
}

// updateMetrics updates distribution metrics
func (td *TaskDistributor) updateMetrics() {
	td.mu.RLock()
	defer td.mu.RUnlock()

	// Update node metrics
	td.metrics.ActiveNodes = len(td.nodes)

	// Calculate average node load
	totalLoad := 0.0
	activeNodes := 0
	for _, node := range td.nodes {
		if td.isNodeAvailable(node) {
			totalLoad += node.LoadScore
			activeNodes++
		}
	}

	if activeNodes > 0 {
		td.metrics.AverageNodeLoad = totalLoad / float64(activeNodes)
	}

	// Calculate task throughput
	if td.metrics.CompletedTasks > 0 {
		duration := time.Since(td.metrics.LastUpdated)
		if duration > 0 {
			td.metrics.TaskThroughput = float64(td.metrics.CompletedTasks) / duration.Seconds()
		}
	}

	td.metrics.LastUpdated = time.Now()
}

// cleanupLoop periodically cleans up completed tasks and stale data
func (td *TaskDistributor) cleanupLoop() {
	defer td.wg.Done()

	ticker := time.NewTicker(td.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-td.ctx.Done():
			return
		case <-ticker.C:
			td.performCleanup()
		}
	}
}

// performCleanup performs cleanup tasks
func (td *TaskDistributor) performCleanup() {
	td.mu.Lock()
	defer td.mu.Unlock()

	// Clean up old completed tasks (keep last 1000)
	if len(td.completedTasks) > 1000 {
		// Convert to slice and sort by completion time
		tasks := make([]*Task, 0, len(td.completedTasks))
		for _, task := range td.completedTasks {
			tasks = append(tasks, task)
		}

		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].EndTime.After(tasks[j].EndTime)
		})

		// Keep only the most recent 1000
		td.completedTasks = make(map[string]*Task)
		for i := 0; i < 1000 && i < len(tasks); i++ {
			td.completedTasks[tasks[i].ID] = tasks[i]
		}
	}

	// Remove stale nodes
	cutoff := time.Now().Add(-td.config.NodeTimeout)
	for nodeID, node := range td.nodes {
		if node.LastSeen.Before(cutoff) {
			delete(td.nodes, nodeID)
			delete(td.nodeCapacities, nodeID)
			delete(td.nodeMetrics, nodeID)
		}
	}
}

// GetMetrics returns current distribution metrics
func (td *TaskDistributor) GetMetrics() *DistributionMetrics {
	td.mu.RLock()
	defer td.mu.RUnlock()

	metrics := *td.metrics
	return &metrics
}

// Close closes the task distributor
func (td *TaskDistributor) Close() error {
	td.cancel()
	td.wg.Wait()
	return nil
}

// RoundRobinStrategy implements round-robin task distribution
type RoundRobinStrategy struct {
	lastNodeIndex int
}

func (s *RoundRobinStrategy) Name() string {
	return "round_robin"
}

func (s *RoundRobinStrategy) SelectNode(task *Task, availableNodes []*NodeInfo) (*NodeInfo, error) {
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	s.lastNodeIndex = (s.lastNodeIndex + 1) % len(availableNodes)
	return availableNodes[s.lastNodeIndex], nil
}

func (s *RoundRobinStrategy) CalculateScore(task *Task, node *NodeInfo) float64 {
	return 1.0 // All nodes have equal score in round-robin
}

// LeastLoadedStrategy selects the node with the lowest load
type LeastLoadedStrategy struct{}

func (s *LeastLoadedStrategy) Name() string {
	return "least_loaded"
}

func (s *LeastLoadedStrategy) SelectNode(task *Task, availableNodes []*NodeInfo) (*NodeInfo, error) {
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *NodeInfo
	bestScore := -1.0

	for _, node := range availableNodes {
		score := s.CalculateScore(task, node)
		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *LeastLoadedStrategy) CalculateScore(task *Task, node *NodeInfo) float64 {
	// Higher score for lower load (inverted load score)
	return 1.0 - node.LoadScore
}

// ResourceAwareStrategy selects nodes based on resource availability and requirements
type ResourceAwareStrategy struct{}

func (s *ResourceAwareStrategy) Name() string {
	return "resource_aware"
}

func (s *ResourceAwareStrategy) SelectNode(task *Task, availableNodes []*NodeInfo) (*NodeInfo, error) {
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *NodeInfo
	bestScore := -1.0

	for _, node := range availableNodes {
		score := s.CalculateScore(task, node)
		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *ResourceAwareStrategy) CalculateScore(task *Task, node *NodeInfo) float64 {
	if node.Capacity == nil || task.Requirements == nil {
		return 0.5 // Neutral score if no resource info
	}

	// Calculate resource utilization after task assignment
	cpuUtil := (node.Capacity.TotalCPUCores - node.Capacity.AvailableCPUCores + task.Requirements.CPUCores) / node.Capacity.TotalCPUCores
	memUtil := float64(node.Capacity.TotalMemoryBytes-node.Capacity.AvailableMemoryBytes+task.Requirements.MemoryBytes) / float64(node.Capacity.TotalMemoryBytes)

	// Prefer nodes with lower utilization after assignment
	avgUtil := (cpuUtil + memUtil) / 2.0
	return 1.0 - avgUtil
}

// LatencyBasedStrategy selects nodes based on latency characteristics
type LatencyBasedStrategy struct{}

func (s *LatencyBasedStrategy) Name() string {
	return "latency_based"
}

func (s *LatencyBasedStrategy) SelectNode(task *Task, availableNodes []*NodeInfo) (*NodeInfo, error) {
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *NodeInfo
	bestScore := -1.0

	for _, node := range availableNodes {
		score := s.CalculateScore(task, node)
		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *LatencyBasedStrategy) CalculateScore(task *Task, node *NodeInfo) float64 {
	// Higher score for lower latency
	maxLatency := 1000.0 // 1 second in milliseconds
	latencyMs := float64(node.AverageLatency.Milliseconds())

	if latencyMs >= maxLatency {
		return 0.0
	}

	return 1.0 - (latencyMs / maxLatency)
}
