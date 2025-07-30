package scheduler

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// LoadBalancerConfig configures the load balancer
type LoadBalancerConfig struct {
	Algorithm        string
	Interval         time.Duration
	WeightingFactors map[string]float64
	HealthThreshold  float64
	LoadThreshold    float64
}

// TaskLoadBalancer manages task assignment to workers
type TaskLoadBalancer struct {
	config        *LoadBalancerConfig
	workerManager *WorkerManager
	metrics       *LoadBalancerMetrics
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// LoadBalancerMetrics tracks load balancer performance
type LoadBalancerMetrics struct {
	TotalSelections      int64         `json:"total_selections"`
	SuccessfulSelections int64         `json:"successful_selections"`
	FailedSelections     int64         `json:"failed_selections"`
	AverageSelectionTime time.Duration `json:"average_selection_time"`
	LastSelectionTime    time.Time     `json:"last_selection_time"`
	AlgorithmUsed        string        `json:"algorithm_used"`
	LastUpdated          time.Time     `json:"last_updated"`
	mu                   sync.RWMutex
}

// WorkerScore represents a worker's suitability score for a task
type WorkerScore struct {
	Worker  *WorkerNode
	Score   float64
	Factors map[string]float64
	Reason  string
}

// NewTaskLoadBalancer creates a new task load balancer
func NewTaskLoadBalancer(config *LoadBalancerConfig, workerManager *WorkerManager) (*TaskLoadBalancer, error) {
	if config == nil {
		config = &LoadBalancerConfig{
			Algorithm: "least_loaded",
			Interval:  10 * time.Second,
			WeightingFactors: map[string]float64{
				"cpu_usage":    0.3,
				"memory_usage": 0.2,
				"task_count":   0.3,
				"health_score": 0.2,
			},
			HealthThreshold: 0.7,
			LoadThreshold:   0.8,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	balancer := &TaskLoadBalancer{
		config:        config,
		workerManager: workerManager,
		metrics: &LoadBalancerMetrics{
			AlgorithmUsed: config.Algorithm,
			LastUpdated:   time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return balancer, nil
}

// Start starts the load balancer
func (lb *TaskLoadBalancer) Start() error {
	// Start metrics collection
	lb.wg.Add(1)
	go lb.metricsLoop()

	return nil
}

// Stop stops the load balancer
func (lb *TaskLoadBalancer) Stop() error {
	lb.cancel()
	lb.wg.Wait()
	return nil
}

// SelectWorker selects the best worker for a given task
func (lb *TaskLoadBalancer) SelectWorker(task *Task) (*WorkerNode, error) {
	startTime := time.Now()

	// Get available workers (this would be injected in a real implementation)
	workers := lb.getAvailableWorkers(task)
	if len(workers) == 0 {
		lb.updateMetrics(false, time.Since(startTime))
		return nil, fmt.Errorf("no available workers")
	}

	// Select worker based on algorithm
	var selectedWorker *WorkerNode
	var err error

	switch lb.config.Algorithm {
	case "round_robin":
		selectedWorker, err = lb.selectRoundRobin(workers)
	case "least_loaded":
		selectedWorker, err = lb.selectLeastLoaded(workers)
	case "weighted_round_robin":
		selectedWorker, err = lb.selectWeightedRoundRobin(workers)
	case "resource_aware":
		selectedWorker, err = lb.selectResourceAware(task, workers)
	case "capability_based":
		selectedWorker, err = lb.selectCapabilityBased(task, workers)
	default:
		selectedWorker, err = lb.selectLeastLoaded(workers)
	}

	if err != nil {
		lb.updateMetrics(false, time.Since(startTime))
		return nil, err
	}

	lb.updateMetrics(true, time.Since(startTime))
	return selectedWorker, nil
}

// selectRoundRobin implements round-robin selection
func (lb *TaskLoadBalancer) selectRoundRobin(workers []*WorkerNode) (*WorkerNode, error) {
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers available")
	}

	// Simple round-robin based on current time
	index := int(time.Now().UnixNano()) % len(workers)
	return workers[index], nil
}

// selectLeastLoaded implements least-loaded selection
func (lb *TaskLoadBalancer) selectLeastLoaded(workers []*WorkerNode) (*WorkerNode, error) {
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers available")
	}

	var bestWorker *WorkerNode
	var lowestLoad float64 = math.MaxFloat64

	for _, worker := range workers {
		load := lb.calculateWorkerLoad(worker)
		if load < lowestLoad {
			lowestLoad = load
			bestWorker = worker
		}
	}

	if bestWorker == nil {
		return workers[0], nil // Fallback to first worker
	}

	return bestWorker, nil
}

// selectWeightedRoundRobin implements weighted round-robin selection
func (lb *TaskLoadBalancer) selectWeightedRoundRobin(workers []*WorkerNode) (*WorkerNode, error) {
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers available")
	}

	// Calculate weights based on worker capacity
	var totalWeight float64
	weights := make([]float64, len(workers))

	for i, worker := range workers {
		weight := lb.calculateWorkerWeight(worker)
		weights[i] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return lb.selectRoundRobin(workers)
	}

	// Select based on weighted probability
	target := float64(time.Now().UnixNano()%int64(totalWeight*1000)) / 1000.0
	var cumulative float64

	for i, weight := range weights {
		cumulative += weight
		if cumulative >= target {
			return workers[i], nil
		}
	}

	return workers[len(workers)-1], nil
}

// selectResourceAware implements resource-aware selection
func (lb *TaskLoadBalancer) selectResourceAware(task *Task, workers []*WorkerNode) (*WorkerNode, error) {
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers available")
	}

	scores := make([]*WorkerScore, 0, len(workers))

	for _, worker := range workers {
		score := lb.calculateResourceScore(task, worker)
		scores = append(scores, score)
	}

	// Sort by score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	if len(scores) == 0 || scores[0].Score <= 0 {
		return nil, fmt.Errorf("no suitable workers found")
	}

	return scores[0].Worker, nil
}

// selectCapabilityBased implements capability-based selection
func (lb *TaskLoadBalancer) selectCapabilityBased(task *Task, workers []*WorkerNode) (*WorkerNode, error) {
	if len(workers) == 0 {
		return nil, fmt.Errorf("no workers available")
	}

	// Filter workers by required capabilities
	var suitableWorkers []*WorkerNode

	if task.Requirements != nil && len(task.Requirements.SpecialHardware) > 0 {
		for _, worker := range workers {
			if lb.hasRequiredCapabilities(worker, task.Requirements.SpecialHardware) {
				suitableWorkers = append(suitableWorkers, worker)
			}
		}
	} else {
		suitableWorkers = workers
	}

	if len(suitableWorkers) == 0 {
		return nil, fmt.Errorf("no workers with required capabilities")
	}

	// Use least-loaded among suitable workers
	return lb.selectLeastLoaded(suitableWorkers)
}

// calculateWorkerLoad calculates the current load of a worker
func (lb *TaskLoadBalancer) calculateWorkerLoad(worker *WorkerNode) float64 {
	if worker.Load == nil {
		return 0.0
	}

	// Weighted combination of different load factors
	cpuWeight := lb.config.WeightingFactors["cpu_usage"]
	memoryWeight := lb.config.WeightingFactors["memory_usage"]
	taskWeight := lb.config.WeightingFactors["task_count"]

	cpuLoad := worker.Load.CPUUsage * cpuWeight
	memoryLoad := worker.Load.MemoryUsage * memoryWeight
	taskLoad := float64(worker.Load.ActiveTasks) / 10.0 * taskWeight // Normalize to 0-1

	return cpuLoad + memoryLoad + taskLoad
}

// calculateWorkerWeight calculates the weight of a worker for weighted selection
func (lb *TaskLoadBalancer) calculateWorkerWeight(worker *WorkerNode) float64 {
	if worker.Resources == nil {
		return 1.0
	}

	// Weight based on available resources
	cpuWeight := worker.Resources.AvailableCPU / worker.Resources.TotalCPU
	memoryWeight := float64(worker.Resources.AvailableMemory) / float64(worker.Resources.TotalMemory)

	// Health factor
	healthFactor := worker.HealthScore
	if healthFactor < lb.config.HealthThreshold {
		healthFactor = 0.1 // Heavily penalize unhealthy workers
	}

	return (cpuWeight + memoryWeight) / 2.0 * healthFactor
}

// calculateResourceScore calculates a resource-based score for task assignment
func (lb *TaskLoadBalancer) calculateResourceScore(task *Task, worker *WorkerNode) *WorkerScore {
	score := &WorkerScore{
		Worker:  worker,
		Score:   0.0,
		Factors: make(map[string]float64),
		Reason:  "",
	}

	// Check if worker meets minimum requirements
	if task.Requirements != nil && worker.Resources != nil {
		if task.Requirements.CPU > worker.Resources.AvailableCPU {
			score.Reason = "insufficient CPU"
			return score
		}

		if task.Requirements.Memory > worker.Resources.AvailableMemory {
			score.Reason = "insufficient memory"
			return score
		}

		if task.Requirements.GPU > worker.Resources.AvailableGPU {
			score.Reason = "insufficient GPU"
			return score
		}
	}

	// Calculate positive score factors
	if worker.Resources != nil {
		// Resource availability score
		cpuScore := worker.Resources.AvailableCPU / worker.Resources.TotalCPU
		memoryScore := float64(worker.Resources.AvailableMemory) / float64(worker.Resources.TotalMemory)

		score.Factors["cpu_availability"] = cpuScore
		score.Factors["memory_availability"] = memoryScore
	}

	// Health score
	score.Factors["health"] = worker.HealthScore

	// Load score (inverted - lower load is better)
	loadScore := 1.0 - lb.calculateWorkerLoad(worker)
	score.Factors["load"] = loadScore

	// Calculate weighted total score
	totalScore := 0.0
	for factor, value := range score.Factors {
		weight := lb.config.WeightingFactors[factor]
		if weight == 0 {
			weight = 0.25 // Default weight
		}
		totalScore += value * weight
	}

	score.Score = totalScore
	score.Reason = "suitable"

	return score
}

// hasRequiredCapabilities checks if a worker has the required capabilities
func (lb *TaskLoadBalancer) hasRequiredCapabilities(worker *WorkerNode, required []string) bool {
	workerCaps := make(map[string]bool)
	for _, cap := range worker.Capabilities {
		workerCaps[cap] = true
	}

	for _, req := range required {
		if !workerCaps[req] {
			return false
		}
	}

	return true
}

// getAvailableWorkers gets available workers for a task
func (lb *TaskLoadBalancer) getAvailableWorkers(task *Task) []*WorkerNode {
	if lb.workerManager == nil {
		return []*WorkerNode{}
	}

	// Get available workers from worker manager
	return lb.workerManager.GetAvailableWorkers()
}

// updateMetrics updates load balancer metrics
func (lb *TaskLoadBalancer) updateMetrics(success bool, duration time.Duration) {
	lb.metrics.mu.Lock()
	defer lb.metrics.mu.Unlock()

	lb.metrics.TotalSelections++
	if success {
		lb.metrics.SuccessfulSelections++
	} else {
		lb.metrics.FailedSelections++
	}

	// Update average selection time
	if lb.metrics.TotalSelections == 1 {
		lb.metrics.AverageSelectionTime = duration
	} else {
		lb.metrics.AverageSelectionTime = (lb.metrics.AverageSelectionTime + duration) / 2
	}

	lb.metrics.LastSelectionTime = time.Now()
	lb.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current load balancer metrics
func (lb *TaskLoadBalancer) GetMetrics() *LoadBalancerMetrics {
	lb.metrics.mu.RLock()
	defer lb.metrics.mu.RUnlock()

	// Create a copy
	metrics := *lb.metrics
	return &metrics
}

// SetAlgorithm changes the load balancing algorithm
func (lb *TaskLoadBalancer) SetAlgorithm(algorithm string) {
	lb.config.Algorithm = algorithm

	lb.metrics.mu.Lock()
	lb.metrics.AlgorithmUsed = algorithm
	lb.metrics.LastUpdated = time.Now()
	lb.metrics.mu.Unlock()
}

// metricsLoop runs the metrics collection loop
func (lb *TaskLoadBalancer) metricsLoop() {
	defer lb.wg.Done()

	ticker := time.NewTicker(lb.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case <-ticker.C:
			// Periodic metrics collection if needed
		}
	}
}
