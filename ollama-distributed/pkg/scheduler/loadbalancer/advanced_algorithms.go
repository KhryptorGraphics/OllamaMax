package loadbalancer

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// PredictiveLoadBalancingAlgorithm implements predictive load balancing
type PredictiveLoadBalancingAlgorithm struct {
	predictor *PerformancePredictor
	metrics   *AlgorithmMetrics
	history   *PredictionHistory
}

// PredictionHistory tracks prediction accuracy
type PredictionHistory struct {
	records []*PredictionRecord
	mu      sync.RWMutex
}

// PredictionRecord represents a prediction record
type PredictionRecord struct {
	NodeID           string
	TaskType         string
	PredictedLatency time.Duration
	ActualLatency    time.Duration
	PredictedThroughput float64
	ActualThroughput float64
	Accuracy         float64
	Timestamp        time.Time
}

// NewPredictiveLoadBalancingAlgorithm creates a new predictive load balancing algorithm
func NewPredictiveLoadBalancingAlgorithm(predictor *PerformancePredictor) *PredictiveLoadBalancingAlgorithm {
	return &PredictiveLoadBalancingAlgorithm{
		predictor: predictor,
		metrics: &AlgorithmMetrics{
			LastUsed: time.Now(),
		},
		history: &PredictionHistory{
			records: make([]*PredictionRecord, 0),
		},
	}
}

// GetName returns the algorithm name
func (plba *PredictiveLoadBalancingAlgorithm) GetName() string {
	return "predictive"
}

// GetMetrics returns algorithm metrics
func (plba *PredictiveLoadBalancingAlgorithm) GetMetrics() *AlgorithmMetrics {
	return plba.metrics
}

// UpdateMetrics updates algorithm metrics
func (plba *PredictiveLoadBalancingAlgorithm) UpdateMetrics(result *SelectionResult) {
	plba.metrics.Selections++
	if result.Successful {
		plba.metrics.SuccessRate = (plba.metrics.SuccessRate*float64(plba.metrics.Selections-1) + 1.0) / float64(plba.metrics.Selections)
		plba.metrics.AverageLatency = (plba.metrics.AverageLatency*time.Duration(plba.metrics.Selections-1) + 
			result.ExecutionLatency) / time.Duration(plba.metrics.Selections)
		plba.metrics.Throughput = (plba.metrics.Throughput*float64(plba.metrics.Selections-1) + result.Throughput) / float64(plba.metrics.Selections)
	} else {
		plba.metrics.SuccessRate = (plba.metrics.SuccessRate*float64(plba.metrics.Selections-1)) / float64(plba.metrics.Selections)
	}
	plba.metrics.LastUsed = time.Now()
}

// SelectNodes selects nodes based on predictive performance
func (plba *PredictiveLoadBalancingAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	start := time.Now()
	
	// Get task type
	taskType := "general"
	if taskMap, ok := task.(map[string]interface{}); ok {
		if t, exists := taskMap["type"]; exists {
			if tStr, ok := t.(string); ok {
				taskType = tStr
			}
		}
	}
	
	// Predict performance for each node
	nodeScores := make([]NodeScore, len(nodes))
	
	for i, node := range nodes {
		// Predict performance
		predictedLatency, predictedThroughput := plba.predictor.PredictPerformance(node, taskType)
		
		// Calculate score based on predictions and current load
		score := plba.calculateScore(node, predictedLatency, predictedThroughput)
		
		nodeScores[i] = NodeScore{
			Node:     node,
			Score:    score,
			Latency:  predictedLatency,
			Throughput: predictedThroughput,
		}
	}
	
	// Sort nodes by score (higher is better)
	sort.Slice(nodeScores, func(i, j int) bool {
		return nodeScores[i].Score > nodeScores[j].Score
	})
	
	// Select top nodes (at least 1, up to 3)
	selectedCount := 1
	if len(nodeScores) > 3 {
		selectedCount = 3
	} else if len(nodeScores) > 1 {
		selectedCount = len(nodeScores)
	}
	
	selectedNodes := make([]*NodeInfo, selectedCount)
	for i := 0; i < selectedCount; i++ {
		selectedNodes[i] = nodeScores[i].Node
	}
	
	// Record selection for learning
	selectionTime := time.Since(start)
	plba.recordSelection(selectedNodes, taskType, selectionTime)
	
	return selectedNodes, nil
}

// NodeScore represents a node with its score
type NodeScore struct {
	Node      *NodeInfo
	Score     float64
	Latency   time.Duration
	Throughput float64
}

// calculateScore calculates a score for a node based on various factors
func (plba *PredictiveLoadBalancingAlgorithm) calculateScore(node *NodeInfo, predictedLatency time.Duration, predictedThroughput float64) float64 {
	// Factors affecting score:
	// 1. Predicted latency (lower is better)
	// 2. Predicted throughput (higher is better)
	// 3. Current load (lower is better)
	// 4. Health score (higher is better)
	// 5. Performance score (higher is better)
	
	// Normalize latency (0-1, where 1 is best)
	latencyScore := 1.0 - math.Min(float64(predictedLatency)/float64(10*time.Second), 1.0)
	
	// Normalize throughput (0-1, where 1 is best)
	throughputScore := math.Min(predictedThroughput/1000.0, 1.0) // Assuming 1000 ops/sec is good
	
	// Load score (0-1, where 1 is best - lower load)
	loadScore := 1.0 - (node.Usage.CPUUtilization+node.Usage.MemoryUtilization)/2.0/100.0
	
	// Health score (0-1)
	healthScore := node.HealthScore / 100.0
	
	// Performance score (0-1)
	performanceScore := node.PerformanceScore / 100.0
	
	// Weighted combination
	score := 0.3*latencyScore + 0.2*throughputScore + 0.2*loadScore + 0.15*healthScore + 0.15*performanceScore
	
	return score
}

// recordSelection records a node selection for learning
func (plba *PredictiveLoadBalancingAlgorithm) recordSelection(nodes []*NodeInfo, taskType string, selectionTime time.Duration) {
	// In a real implementation, this would record the actual performance
	// For now, we'll just update the metrics
	plba.metrics.Selections++
	plba.metrics.LastUsed = time.Now()
}

// AdaptiveLoadBalancingAlgorithm implements adaptive load balancing that learns from history
type AdaptiveLoadBalancingAlgorithm struct {
	history       *RequestHistory
	metrics       *AlgorithmMetrics
	learningRate  float64
	weightFactors map[string]float64
}

// NewAdaptiveLoadBalancingAlgorithm creates a new adaptive load balancing algorithm
func NewAdaptiveLoadBalancingAlgorithm(history *RequestHistory) *AdaptiveLoadBalancingAlgorithm {
	return &AdaptiveLoadBalancingAlgorithm{
		history:      history,
		metrics:      &AlgorithmMetrics{LastUsed: time.Now()},
		learningRate: 0.1,
		weightFactors: map[string]float64{
			"latency":     0.3,
			"throughput":  0.25,
			"load":        0.2,
			"health":      0.15,
			"performance": 0.1,
		},
	}
}

// GetName returns the algorithm name
func (alba *AdaptiveLoadBalancingAlgorithm) GetName() string {
	return "adaptive"
}

// GetMetrics returns algorithm metrics
func (alba *AdaptiveLoadBalancingAlgorithm) GetMetrics() *AlgorithmMetrics {
	return alba.metrics
}

// UpdateMetrics updates algorithm metrics
func (alba *AdaptiveLoadBalancingAlgorithm) UpdateMetrics(result *SelectionResult) {
	alba.metrics.Selections++
	if result.Successful {
		alba.metrics.SuccessRate = (alba.metrics.SuccessRate*float64(alba.metrics.Selections-1) + 1.0) / float64(alba.metrics.Selections)
		alba.metrics.AverageLatency = (alba.metrics.AverageLatency*time.Duration(alba.metrics.Selections-1) + 
			result.ExecutionLatency) / time.Duration(alba.metrics.Selections)
		alba.metrics.Throughput = (alba.metrics.Throughput*float64(alba.metrics.Selections-1) + result.Throughput) / float64(alba.metrics.Selections)
	} else {
		alba.metrics.SuccessRate = (alba.metrics.SuccessRate*float64(alba.metrics.Selections-1)) / float64(alba.metrics.Selections)
	}
	alba.metrics.LastUsed = time.Now()
	
	// Learn from result
	alba.learnFromResult(result)
}

// SelectNodes selects nodes based on adaptive scoring
func (alba *AdaptiveLoadBalancingAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Get historical patterns for this task type
	taskType := "general"
	if taskMap, ok := task.(map[string]interface{}); ok {
		if t, exists := taskMap["type"]; exists {
			if tStr, ok := t.(string); ok {
				taskType = tStr
			}
		}
	}
	
	pattern := alba.history.getPattern(taskType)
	
	// Calculate scores for each node
	nodeScores := make([]NodeScore, len(nodes))
	
	for i, node := range nodes {
		score := alba.calculateAdaptiveScore(node, pattern)
		nodeScores[i] = NodeScore{
			Node:  node,
			Score: score,
		}
	}
	
	// Sort nodes by score (higher is better)
	sort.Slice(nodeScores, func(i, j int) bool {
		return nodeScores[i].Score > nodeScores[j].Score
	})
	
	// Select top node
	selectedNodes := []*NodeInfo{nodeScores[0].Node}
	
	return selectedNodes, nil
}

// calculateAdaptiveScore calculates a score using adaptive weights
func (alba *AdaptiveLoadBalancingAlgorithm) calculateAdaptiveScore(node *NodeInfo, pattern *RequestPattern) float64 {
	// Base score calculation
	latencyScore := 1.0 - math.Min(float64(node.Latency)/float64(10*time.Second), 1.0)
	// Calculate throughput score based on active requests and load average
	// We'll approximate throughput as inverse of load (lower load = higher throughput)
	throughputScore := 1.0 - (node.Usage.LoadAverage / 100.0)
	if throughputScore < 0 {
		throughputScore = 0
	}
	loadScore := 1.0 - (node.Usage.CPUUtilization+node.Usage.MemoryUtilization)/2.0/100.0
	healthScore := node.HealthScore / 100.0
	performanceScore := node.PerformanceScore / 100.0
	
	// Apply weights
	score := alba.weightFactors["latency"]*latencyScore +
		alba.weightFactors["throughput"]*throughputScore +
		alba.weightFactors["load"]*loadScore +
		alba.weightFactors["health"]*healthScore +
		alba.weightFactors["performance"]*performanceScore
	
	// Adjust based on pattern if available
	if pattern != nil && pattern.Confidence > 0.5 {
		// Check if this node is preferred in the pattern
		isPreferred := false
		for _, preferredNode := range pattern.PreferredNodes {
			if preferredNode == node.ID {
				isPreferred = true
				break
			}
		}
		
		if isPreferred {
			score *= (1.0 + pattern.Confidence*0.2) // Boost score for preferred nodes
		}
	}
	
	return score
}

// learnFromResult learns from a selection result to improve future decisions
func (alba *AdaptiveLoadBalancingAlgorithm) learnFromResult(result *SelectionResult) {
	if !result.Successful {
		// Negative reinforcement - reduce weights for factors that likely contributed to failure
		alba.weightFactors["latency"] *= (1.0 - alba.learningRate*0.5)
		alba.weightFactors["load"] *= (1.0 - alba.learningRate*0.3)
	} else {
		// Positive reinforcement based on performance
		if result.ExecutionLatency < 100*time.Millisecond {
			alba.weightFactors["latency"] *= (1.0 + alba.learningRate*0.3)
		}
		if result.Throughput > 50.0 {
			alba.weightFactors["throughput"] *= (1.0 + alba.learningRate*0.2)
		}
	}
	
	// Normalize weights
	alba.normalizeWeights()
}

// normalizeWeights normalizes the weight factors to sum to 1.0
func (alba *AdaptiveLoadBalancingAlgorithm) normalizeWeights() {
	sum := 0.0
	for _, weight := range alba.weightFactors {
		sum += weight
	}
	
	if sum > 0 {
		for key, weight := range alba.weightFactors {
			alba.weightFactors[key] = weight / sum
		}
	}
}

// getPattern gets a pattern for a task type
func (rh *RequestHistory) getPattern(taskType string) *RequestPattern {
	rh.patternsMu.RLock()
	defer rh.patternsMu.RUnlock()
	
	if pattern, exists := rh.patterns[taskType]; exists {
		return pattern
	}
	
	return nil
}

// ResourceAwareLoadBalancingAlgorithm implements resource-aware load balancing
type ResourceAwareLoadBalancingAlgorithm struct {
	metrics *AlgorithmMetrics
}

// NewResourceAwareLoadBalancingAlgorithm creates a new resource-aware load balancing algorithm
func NewResourceAwareLoadBalancingAlgorithm() *ResourceAwareLoadBalancingAlgorithm {
	return &ResourceAwareLoadBalancingAlgorithm{
		metrics: &AlgorithmMetrics{LastUsed: time.Now()},
	}
}

// GetName returns the algorithm name
func (ralba *ResourceAwareLoadBalancingAlgorithm) GetName() string {
	return "resource_aware"
}

// GetMetrics returns algorithm metrics
func (ralba *ResourceAwareLoadBalancingAlgorithm) GetMetrics() *AlgorithmMetrics {
	return ralba.metrics
}

// UpdateMetrics updates algorithm metrics
func (ralba *ResourceAwareLoadBalancingAlgorithm) UpdateMetrics(result *SelectionResult) {
	ralba.metrics.Selections++
	if result.Successful {
		ralba.metrics.SuccessRate = (ralba.metrics.SuccessRate*float64(ralba.metrics.Selections-1) + 1.0) / float64(ralba.metrics.Selections)
		ralba.metrics.AverageLatency = (ralba.metrics.AverageLatency*time.Duration(ralba.metrics.Selections-1) + 
			result.ExecutionLatency) / time.Duration(ralba.metrics.Selections)
		ralba.metrics.Throughput = (ralba.metrics.Throughput*float64(ralba.metrics.Selections-1) + result.Throughput) / float64(ralba.metrics.Selections)
	} else {
		ralba.metrics.SuccessRate = (ralba.metrics.SuccessRate*float64(ralba.metrics.Selections-1)) / float64(ralba.metrics.Selections)
	}
	ralba.metrics.LastUsed = time.Now()
}

// SelectNodes selects nodes based on resource availability
func (ralba *ResourceAwareLoadBalancingAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Estimate resource requirements from task
	taskRequirements := ralba.estimateTaskRequirements(task)
	
	// Score nodes based on resource availability
	nodeScores := make([]NodeScore, len(nodes))
	
	for i, node := range nodes {
		score := ralba.calculateResourceScore(node, taskRequirements)
		nodeScores[i] = NodeScore{
			Node:  node,
			Score: score,
		}
	}
	
	// Sort nodes by score (higher is better)
	sort.Slice(nodeScores, func(i, j int) bool {
		return nodeScores[i].Score > nodeScores[j].Score
	})
	
	// Select top nodes (1-3 depending on availability)
	selectedCount := 1
	if len(nodeScores) > 3 {
		selectedCount = 3
	} else if len(nodeScores) > 1 {
		selectedCount = len(nodeScores)
	}
	
	selectedNodes := make([]*NodeInfo, selectedCount)
	for i := 0; i < selectedCount; i++ {
		selectedNodes[i] = nodeScores[i].Node
	}
	
	return selectedNodes, nil
}

// TaskRequirements represents the resource requirements of a task
type TaskRequirements struct {
	CPU     int64   // CPU cores required
	Memory  int64   // Memory in bytes required
	GPU     int     // GPU count required
	Network int64   // Network bandwidth required
}

// estimateTaskRequirements estimates resource requirements for a task
func (ralba *ResourceAwareLoadBalancingAlgorithm) estimateTaskRequirements(task interface{}) *TaskRequirements {
	// Default requirements
	requirements := &TaskRequirements{
		CPU:     2,
		Memory:  4 * 1024 * 1024 * 1024, // 4GB
		GPU:     1,
		Network: 100 * 1024 * 1024,     // 100MB/s
	}
	
	// Try to extract more specific requirements from task
	if taskMap, ok := task.(map[string]interface{}); ok {
		// Extract model size if available
		if modelInfo, exists := taskMap["model"]; exists {
			if modelMap, ok := modelInfo.(map[string]interface{}); ok {
				if size, exists := modelMap["size"]; exists {
					if sizeInt, ok := size.(int64); ok {
						// Estimate memory requirement as 1.5x model size
						requirements.Memory = int64(float64(sizeInt) * 1.5)
						// Estimate CPU requirement based on model size
						if sizeInt > 10*1024*1024*1024 { // 10GB+
							requirements.CPU = 8
						} else if sizeInt > 5*1024*1024*1024 { // 5GB+
							requirements.CPU = 4
						}
					}
				}
			}
		}
		
		// Extract context length if available
		if options, exists := taskMap["options"]; exists {
			if optionsMap, ok := options.(map[string]interface{}); ok {
				if ctxLen, exists := optionsMap["num_ctx"]; exists {
					if ctxInt, ok := ctxLen.(int); ok {
						// Increase memory requirement for longer contexts
						if ctxInt > 2048 {
							requirements.Memory += int64(ctxInt) * 2 * 1024 // 2KB per token
						}
					}
				}
			}
		}
	}
	
	return requirements
}

// calculateResourceScore calculates a score based on resource availability
func (ralba *ResourceAwareLoadBalancingAlgorithm) calculateResourceScore(node *NodeInfo, requirements *TaskRequirements) float64 {
	// Check if node has sufficient resources
	if node.Capacity.CPUCores < requirements.CPU {
		return 0.0 // Insufficient CPU
	}
	
	if node.Capacity.MemoryBytes < requirements.Memory {
		return 0.0 // Insufficient memory
	}
	
	if node.Capacity.GPUCount < requirements.GPU {
		return 0.0 // Insufficient GPU
	}
	
	if node.Capacity.NetworkBandwidth < requirements.Network {
		return 0.0 // Insufficient network
	}
	
	// Calculate resource utilization scores (0-1, where 1 is best - low utilization)
	cpuUtilization := node.Usage.CPUUtilization / 100.0
	memoryUtilization := node.Usage.MemoryUtilization / 100.0
	gpuUtilization := node.Usage.GPUUtilization / 100.0
	networkUtilization := node.Usage.NetworkUtilization / 100.0
	
	// Resource availability scores (higher is better)
	cpuScore := 1.0 - cpuUtilization
	memoryScore := 1.0 - memoryUtilization
	gpuScore := 1.0 - gpuUtilization
	networkScore := 1.0 - networkUtilization
	
	// Weighted combination
	score := 0.3*cpuScore + 0.3*memoryScore + 0.2*gpuScore + 0.2*networkScore
	
	// Boost score for nodes with plenty of free resources
	freeCPURatio := float64(node.Capacity.CPUCores-requirements.CPU) / float64(node.Capacity.CPUCores)
	freeMemoryRatio := float64(node.Capacity.MemoryBytes-requirements.Memory) / float64(node.Capacity.MemoryBytes)
	
	if freeCPURatio > 0.5 && freeMemoryRatio > 0.5 {
		score *= 1.2 // 20% boost for nodes with plenty of free resources
	}
	
	return score
}