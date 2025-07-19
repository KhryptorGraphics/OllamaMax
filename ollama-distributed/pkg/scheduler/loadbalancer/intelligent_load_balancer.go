package loadbalancer

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"
)

// IntelligentLoadBalancer implements intelligent load balancing with prediction
type IntelligentLoadBalancer struct {
	config      *Config
	algorithms  map[string]LoadBalancingAlgorithm
	predictor   *PerformancePredictor
	history     *RequestHistory
	constraints []LoadBalancingConstraint
	metrics     *LoadBalancerMetrics
	mu          sync.RWMutex
}

// Config holds load balancer configuration
type Config struct {
	Algorithm     string                 `json:"algorithm"`
	LatencyTarget time.Duration          `json:"latency_target"`
	WeightFactors map[string]float64     `json:"weight_factors"`
	Adaptive      bool                   `json:"adaptive"`
	PredictionEnabled bool               `json:"prediction_enabled"`
	HistorySize   int                    `json:"history_size"`
}

// LoadBalancingAlgorithm defines the interface for load balancing algorithms
type LoadBalancingAlgorithm interface {
	SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error)
	GetName() string
	GetMetrics() *AlgorithmMetrics
	UpdateMetrics(result *SelectionResult)
}

// NodeInfo represents node information for load balancing
type NodeInfo struct {
	ID               string                 `json:"id"`
	Address          string                 `json:"address"`
	Capacity         *ResourceCapacity      `json:"capacity"`
	Usage            *ResourceUsage         `json:"usage"`
	Latency          time.Duration          `json:"latency"`
	Bandwidth        int64                  `json:"bandwidth"`
	HealthScore      float64                `json:"health_score"`
	LoadScore        float64                `json:"load_score"`
	PerformanceScore float64                `json:"performance_score"`
	Capabilities     []string               `json:"capabilities"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ResourceCapacity represents node resource capacity
type ResourceCapacity struct {
	CPUCores         int64   `json:"cpu_cores"`
	MemoryBytes      int64   `json:"memory_bytes"`
	GPUCount         int     `json:"gpu_count"`
	GPUMemoryBytes   int64   `json:"gpu_memory_bytes"`
	NetworkBandwidth int64   `json:"network_bandwidth"`
	ComputeScore     float64 `json:"compute_score"`
}

// ResourceUsage represents node resource usage
type ResourceUsage struct {
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
	GPUUtilization    float64 `json:"gpu_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`
	ActiveRequests    int     `json:"active_requests"`
	QueuedRequests    int     `json:"queued_requests"`
	LoadAverage       float64 `json:"load_average"`
}

// LoadBalancingConstraint represents a constraint for load balancing
type LoadBalancingConstraint struct {
	Type     string      `json:"type"`     // "memory", "gpu", "latency", "cost"
	Value    interface{} `json:"value"`
	Operator string      `json:"operator"` // "<", ">", "=", "<=", ">="
	Priority int         `json:"priority"`
}

// PerformancePredictor predicts node performance
type PerformancePredictor struct {
	models     map[string]*PredictionModel
	history    []*PerformanceSample
	historyMu  sync.RWMutex
	learning   bool
	accuracy   float64
}

// PredictionModel represents a performance prediction model
type PredictionModel struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Weights    map[string]float64     `json:"weights"`
	Accuracy   float64                `json:"accuracy"`
	LastTrained time.Time             `json:"last_trained"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// PerformanceSample represents a performance sample for learning
type PerformanceSample struct {
	NodeID          string        `json:"node_id"`
	TaskType        string        `json:"task_type"`
	ResourceState   *ResourceUsage `json:"resource_state"`
	PredictedLatency time.Duration `json:"predicted_latency"`
	ActualLatency   time.Duration `json:"actual_latency"`
	PredictedThroughput float64   `json:"predicted_throughput"`
	ActualThroughput float64      `json:"actual_throughput"`
	Timestamp       time.Time     `json:"timestamp"`
}

// RequestHistory tracks request history for patterns
type RequestHistory struct {
	requests   []*RequestRecord
	requestsMu sync.RWMutex
	patterns   map[string]*RequestPattern
	patternsMu sync.RWMutex
}

// RequestRecord represents a request record
type RequestRecord struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	SelectedNodes    []string               `json:"selected_nodes"`
	Latency          time.Duration          `json:"latency"`
	Throughput       float64                `json:"throughput"`
	ResourceUsage    map[string]interface{} `json:"resource_usage"`
	Timestamp        time.Time              `json:"timestamp"`
	Successful       bool                   `json:"successful"`
}

// RequestPattern represents a request pattern
type RequestPattern struct {
	Type              string        `json:"type"`
	AverageLatency    time.Duration `json:"average_latency"`
	AverageThroughput float64       `json:"average_throughput"`
	PreferredNodes    []string      `json:"preferred_nodes"`
	ResourceProfile   map[string]float64 `json:"resource_profile"`
	Confidence        float64       `json:"confidence"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// LoadBalancerMetrics represents load balancer metrics
type LoadBalancerMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	Throughput         float64       `json:"throughput"`
	AlgorithmMetrics   map[string]*AlgorithmMetrics `json:"algorithm_metrics"`
	LastUpdated        time.Time     `json:"last_updated"`
}

// AlgorithmMetrics represents metrics for a specific algorithm
type AlgorithmMetrics struct {
	Selections        int64         `json:"selections"`
	SuccessRate       float64       `json:"success_rate"`
	AverageLatency    time.Duration `json:"average_latency"`
	Throughput        float64       `json:"throughput"`
	LastUsed          time.Time     `json:"last_used"`
}

// SelectionResult represents the result of a node selection
type SelectionResult struct {
	Nodes            []*NodeInfo   `json:"nodes"`
	Algorithm        string        `json:"algorithm"`
	SelectionTime    time.Duration `json:"selection_time"`
	ExecutionLatency time.Duration `json:"execution_latency"`
	Throughput       float64       `json:"throughput"`
	Successful       bool          `json:"successful"`
	Timestamp        time.Time     `json:"timestamp"`
}

// NewIntelligentLoadBalancer creates a new intelligent load balancer
func NewIntelligentLoadBalancer(config *Config) *IntelligentLoadBalancer {
	ilb := &IntelligentLoadBalancer{
		config:      config,
		algorithms:  make(map[string]LoadBalancingAlgorithm),
		constraints: make([]LoadBalancingConstraint, 0),
		metrics: &LoadBalancerMetrics{
			AlgorithmMetrics: make(map[string]*AlgorithmMetrics),
			LastUpdated:      time.Now(),
		},
	}
	
	// Initialize performance predictor
	ilb.predictor = &PerformancePredictor{
		models:   make(map[string]*PredictionModel),
		history:  make([]*PerformanceSample, 0),
		learning: config.PredictionEnabled,
		accuracy: 0.7, // Initial accuracy
	}
	
	// Initialize request history
	ilb.history = &RequestHistory{
		requests: make([]*RequestRecord, 0),
		patterns: make(map[string]*RequestPattern),
	}
	
	// Register algorithms
	ilb.RegisterAlgorithm(NewWeightedRoundRobinAlgorithm())
	ilb.RegisterAlgorithm(NewLeastEffectiveLoadAlgorithm())
	ilb.RegisterAlgorithm(NewLocalityAwareAlgorithm())
	ilb.RegisterAlgorithm(NewPredictiveAlgorithm(ilb.predictor))
	ilb.RegisterAlgorithm(NewAdaptiveAlgorithm(ilb.history))
	
	return ilb
}

// RegisterAlgorithm registers a load balancing algorithm
func (ilb *IntelligentLoadBalancer) RegisterAlgorithm(algorithm LoadBalancingAlgorithm) {
	ilb.mu.Lock()
	defer ilb.mu.Unlock()
	
	ilb.algorithms[algorithm.GetName()] = algorithm
	ilb.metrics.AlgorithmMetrics[algorithm.GetName()] = &AlgorithmMetrics{
		LastUsed: time.Now(),
	}
}

// SelectNodes selects the best nodes for a task
func (ilb *IntelligentLoadBalancer) SelectNodes(task interface{}, availableNodes []*NodeInfo) ([]*NodeInfo, error) {
	start := time.Now()
	
	// Update metrics
	ilb.metrics.TotalRequests++
	
	// Apply constraints
	constrainedNodes := ilb.applyConstraints(availableNodes)
	if len(constrainedNodes) == 0 {
		ilb.metrics.FailedRequests++
		return nil, fmt.Errorf("no nodes satisfy constraints")
	}
	
	// Select algorithm
	algorithm, err := ilb.selectAlgorithm(task, constrainedNodes)
	if err != nil {
		ilb.metrics.FailedRequests++
		return nil, fmt.Errorf("failed to select algorithm: %v", err)
	}
	
	// Select nodes using the chosen algorithm
	selectedNodes, err := algorithm.SelectNodes(task, constrainedNodes)
	if err != nil {
		ilb.metrics.FailedRequests++
		return nil, fmt.Errorf("algorithm selection failed: %v", err)
	}
	
	// Update metrics
	ilb.metrics.SuccessfulRequests++
	selectionTime := time.Since(start)
	
	// Record selection result
	result := &SelectionResult{
		Nodes:         selectedNodes,
		Algorithm:     algorithm.GetName(),
		SelectionTime: selectionTime,
		Successful:    true,
		Timestamp:     time.Now(),
	}
	
	// Update algorithm metrics
	algorithm.UpdateMetrics(result)
	
	slog.Debug("node selection completed",
		"algorithm", algorithm.GetName(),
		"selected_nodes", len(selectedNodes),
		"selection_time", selectionTime,
		"available_nodes", len(availableNodes))
	
	return selectedNodes, nil
}

// applyConstraints applies load balancing constraints to nodes
func (ilb *IntelligentLoadBalancer) applyConstraints(nodes []*NodeInfo) []*NodeInfo {
	if len(ilb.constraints) == 0 {
		return nodes
	}
	
	constrained := make([]*NodeInfo, 0)
	
	for _, node := range nodes {
		satisfies := true
		
		for _, constraint := range ilb.constraints {
			if !ilb.satisfiesConstraint(node, constraint) {
				satisfies = false
				break
			}
		}
		
		if satisfies {
			constrained = append(constrained, node)
		}
	}
	
	return constrained
}

// satisfiesConstraint checks if a node satisfies a constraint
func (ilb *IntelligentLoadBalancer) satisfiesConstraint(node *NodeInfo, constraint LoadBalancingConstraint) bool {
	switch constraint.Type {
	case "memory":
		memoryUtilization := node.Usage.MemoryUtilization
		threshold := constraint.Value.(float64)
		return ilb.compareValues(memoryUtilization, threshold, constraint.Operator)
		
	case "gpu":
		gpuUtilization := node.Usage.GPUUtilization
		threshold := constraint.Value.(float64)
		return ilb.compareValues(gpuUtilization, threshold, constraint.Operator)
		
	case "latency":
		latency := node.Latency
		threshold := constraint.Value.(time.Duration)
		return ilb.compareLatency(latency, threshold, constraint.Operator)
		
	case "cost":
		// Cost constraint implementation would go here
		return true
		
	default:
		return true
	}
}

// compareValues compares two float64 values using an operator
func (ilb *IntelligentLoadBalancer) compareValues(value, threshold float64, operator string) bool {
	switch operator {
	case "<":
		return value < threshold
	case ">":
		return value > threshold
	case "=":
		return math.Abs(value-threshold) < 0.001
	case "<=":
		return value <= threshold
	case ">=":
		return value >= threshold
	default:
		return true
	}
}

// compareLatency compares two latency values using an operator
func (ilb *IntelligentLoadBalancer) compareLatency(latency, threshold time.Duration, operator string) bool {
	switch operator {
	case "<":
		return latency < threshold
	case ">":
		return latency > threshold
	case "=":
		return latency == threshold
	case "<=":
		return latency <= threshold
	case ">=":
		return latency >= threshold
	default:
		return true
	}
}

// selectAlgorithm selects the best algorithm for a task
func (ilb *IntelligentLoadBalancer) selectAlgorithm(task interface{}, nodes []*NodeInfo) (LoadBalancingAlgorithm, error) {
	// If adaptive mode is disabled, use configured algorithm
	if !ilb.config.Adaptive {
		if algorithm, exists := ilb.algorithms[ilb.config.Algorithm]; exists {
			return algorithm, nil
		}
		return nil, fmt.Errorf("algorithm not found: %s", ilb.config.Algorithm)
	}
	
	// Adaptive algorithm selection
	return ilb.selectAdaptiveAlgorithm(task, nodes)
}

// selectAdaptiveAlgorithm selects an algorithm adaptively based on context
func (ilb *IntelligentLoadBalancer) selectAdaptiveAlgorithm(task interface{}, nodes []*NodeInfo) (LoadBalancingAlgorithm, error) {
	// Analyze task characteristics
	taskType := ilb.getTaskType(task)
	nodeCount := len(nodes)
	loadVariance := ilb.calculateLoadVariance(nodes)
	
	// Select algorithm based on context
	if nodeCount <= 2 {
		// Simple round-robin for small clusters
		return ilb.algorithms["weighted_round_robin"], nil
	}
	
	if loadVariance > 0.5 {
		// Use least effective load for unbalanced clusters
		return ilb.algorithms["least_effective_load"], nil
	}
	
	if taskType == "latency_sensitive" {
		// Use locality-aware for latency-sensitive tasks
		return ilb.algorithms["locality_aware"], nil
	}
	
	if ilb.config.PredictionEnabled {
		// Use predictive algorithm when prediction is enabled
		return ilb.algorithms["predictive"], nil
	}
	
	// Default to adaptive algorithm
	return ilb.algorithms["adaptive"], nil
}

// getTaskType determines the type of task
func (ilb *IntelligentLoadBalancer) getTaskType(task interface{}) string {
	// This would analyze the task to determine its type
	// For now, return a default type
	return "general"
}

// calculateLoadVariance calculates the variance in load across nodes
func (ilb *IntelligentLoadBalancer) calculateLoadVariance(nodes []*NodeInfo) float64 {
	if len(nodes) == 0 {
		return 0.0
	}
	
	// Calculate average load
	totalLoad := 0.0
	for _, node := range nodes {
		totalLoad += node.LoadScore
	}
	averageLoad := totalLoad / float64(len(nodes))
	
	// Calculate variance
	variance := 0.0
	for _, node := range nodes {
		deviation := node.LoadScore - averageLoad
		variance += deviation * deviation
	}
	
	return variance / float64(len(nodes))
}

// AddConstraint adds a load balancing constraint
func (ilb *IntelligentLoadBalancer) AddConstraint(constraint LoadBalancingConstraint) {
	ilb.mu.Lock()
	defer ilb.mu.Unlock()
	
	ilb.constraints = append(ilb.constraints, constraint)
	
	// Sort constraints by priority
	sort.Slice(ilb.constraints, func(i, j int) bool {
		return ilb.constraints[i].Priority > ilb.constraints[j].Priority
	})
}

// RemoveConstraint removes a load balancing constraint
func (ilb *IntelligentLoadBalancer) RemoveConstraint(constraintType string) {
	ilb.mu.Lock()
	defer ilb.mu.Unlock()
	
	for i, constraint := range ilb.constraints {
		if constraint.Type == constraintType {
			ilb.constraints = append(ilb.constraints[:i], ilb.constraints[i+1:]...)
			break
		}
	}
}

// GetMetrics returns load balancer metrics
func (ilb *IntelligentLoadBalancer) GetMetrics() *LoadBalancerMetrics {
	ilb.mu.RLock()
	defer ilb.mu.RUnlock()
	
	// Calculate average latency
	if ilb.metrics.SuccessfulRequests > 0 {
		// This would be calculated from actual measurements
		ilb.metrics.AverageLatency = 100 * time.Millisecond
	}
	
	// Calculate throughput
	if ilb.metrics.SuccessfulRequests > 0 {
		// This would be calculated from actual measurements
		ilb.metrics.Throughput = float64(ilb.metrics.SuccessfulRequests) / time.Since(ilb.metrics.LastUpdated).Seconds()
	}
	
	return ilb.metrics
}

// GetAvailableAlgorithms returns all available algorithms
func (ilb *IntelligentLoadBalancer) GetAvailableAlgorithms() []string {
	ilb.mu.RLock()
	defer ilb.mu.RUnlock()
	
	algorithms := make([]string, 0, len(ilb.algorithms))
	for name := range ilb.algorithms {
		algorithms = append(algorithms, name)
	}
	
	return algorithms
}

// UpdateConfig updates the load balancer configuration
func (ilb *IntelligentLoadBalancer) UpdateConfig(config *Config) {
	ilb.mu.Lock()
	defer ilb.mu.Unlock()
	
	ilb.config = config
	ilb.predictor.learning = config.PredictionEnabled
	
	slog.Info("load balancer configuration updated",
		"algorithm", config.Algorithm,
		"adaptive", config.Adaptive,
		"prediction_enabled", config.PredictionEnabled)
}

// RecordResult records the result of a load balancing decision
func (ilb *IntelligentLoadBalancer) RecordResult(result *SelectionResult) {
	// Record in history
	ilb.history.recordRequest(&RequestRecord{
		ID:            fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Type:          "load_balancing",
		SelectedNodes: ilb.getNodeIDs(result.Nodes),
		Latency:       result.ExecutionLatency,
		Throughput:    result.Throughput,
		Timestamp:     result.Timestamp,
		Successful:    result.Successful,
	})
	
	// Update predictor if enabled
	if ilb.config.PredictionEnabled {
		ilb.predictor.recordSample(&PerformanceSample{
			TaskType:         "load_balancing",
			ActualLatency:    result.ExecutionLatency,
			ActualThroughput: result.Throughput,
			Timestamp:        result.Timestamp,
		})
	}
}

// getNodeIDs extracts node IDs from a slice of nodes
func (ilb *IntelligentLoadBalancer) getNodeIDs(nodes []*NodeInfo) []string {
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = node.ID
	}
	return ids
}

// RequestHistory methods

// recordRequest records a request in the history
func (rh *RequestHistory) recordRequest(record *RequestRecord) {
	rh.requestsMu.Lock()
	defer rh.requestsMu.Unlock()
	
	rh.requests = append(rh.requests, record)
	
	// Keep only last 1000 requests
	if len(rh.requests) > 1000 {
		rh.requests = rh.requests[len(rh.requests)-1000:]
	}
	
	// Update patterns
	go rh.updatePatterns(record)
}

// updatePatterns updates request patterns based on new records
func (rh *RequestHistory) updatePatterns(record *RequestRecord) {
	rh.patternsMu.Lock()
	defer rh.patternsMu.Unlock()
	
	pattern, exists := rh.patterns[record.Type]
	if !exists {
		pattern = &RequestPattern{
			Type:            record.Type,
			PreferredNodes:  record.SelectedNodes,
			ResourceProfile: make(map[string]float64),
			Confidence:      0.5,
			LastUpdated:     time.Now(),
		}
		rh.patterns[record.Type] = pattern
	}
	
	// Update pattern with new data
	pattern.AverageLatency = (pattern.AverageLatency + record.Latency) / 2
	pattern.AverageThroughput = (pattern.AverageThroughput + record.Throughput) / 2
	pattern.LastUpdated = time.Now()
	
	// Update confidence based on success rate
	if record.Successful {
		pattern.Confidence = math.Min(pattern.Confidence*1.1, 1.0)
	} else {
		pattern.Confidence = math.Max(pattern.Confidence*0.9, 0.1)
	}
}

// PerformancePredictor methods

// recordSample records a performance sample
func (pp *PerformancePredictor) recordSample(sample *PerformanceSample) {
	pp.historyMu.Lock()
	defer pp.historyMu.Unlock()
	
	pp.history = append(pp.history, sample)
	
	// Keep only last 1000 samples
	if len(pp.history) > 1000 {
		pp.history = pp.history[len(pp.history)-1000:]
	}
	
	// Update models if learning is enabled
	if pp.learning {
		go pp.updateModels()
	}
}

// updateModels updates prediction models based on new samples
func (pp *PerformancePredictor) updateModels() {
	// This would implement actual machine learning model updates
	// For now, just update accuracy based on recent samples
	pp.historyMu.RLock()
	samples := pp.history
	pp.historyMu.RUnlock()
	
	if len(samples) > 10 {
		// Calculate accuracy based on recent samples
		recentSamples := samples[len(samples)-10:]
		correctPredictions := 0
		
		for _, sample := range recentSamples {
			// Simple accuracy calculation
			if sample.PredictedLatency > 0 {
				error := math.Abs(float64(sample.ActualLatency-sample.PredictedLatency)) / float64(sample.ActualLatency)
				if error < 0.2 { // 20% accuracy threshold
					correctPredictions++
				}
			}
		}
		
		pp.accuracy = float64(correctPredictions) / float64(len(recentSamples))
	}
}

// PredictPerformance predicts the performance of a node for a task
func (pp *PerformancePredictor) PredictPerformance(node *NodeInfo, taskType string) (time.Duration, float64) {
	if !pp.learning {
		// Return simple estimates if prediction is disabled
		return 100 * time.Millisecond, 10.0
	}
	
	// Use historical data and models to predict performance
	pp.historyMu.RLock()
	defer pp.historyMu.RUnlock()
	
	// Find similar samples
	similarSamples := make([]*PerformanceSample, 0)
	for _, sample := range pp.history {
		if sample.NodeID == node.ID && sample.TaskType == taskType {
			similarSamples = append(similarSamples, sample)
		}
	}
	
	if len(similarSamples) == 0 {
		// No historical data, return estimates
		return time.Duration(float64(100*time.Millisecond) / math.Max(node.PerformanceScore, 0.1)), node.PerformanceScore * 10.0
	}
	
	// Calculate weighted average based on recent samples
	totalLatency := time.Duration(0)
	totalThroughput := 0.0
	weightSum := 0.0
	
	for i, sample := range similarSamples {
		// Weight recent samples more heavily
		weight := float64(i+1) / float64(len(similarSamples))
		totalLatency += time.Duration(float64(sample.ActualLatency) * weight)
		totalThroughput += sample.ActualThroughput * weight
		weightSum += weight
	}
	
	predictedLatency := time.Duration(float64(totalLatency) / weightSum)
	predictedThroughput := totalThroughput / weightSum
	
	return predictedLatency, predictedThroughput
}
