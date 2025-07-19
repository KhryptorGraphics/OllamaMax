package loadbalancer

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// WeightedRoundRobinAlgorithm implements weighted round-robin with prediction
type WeightedRoundRobinAlgorithm struct {
	name    string
	metrics *AlgorithmMetrics
	counter int
	weights map[string]float64
}

// NewWeightedRoundRobinAlgorithm creates a new weighted round-robin algorithm
func NewWeightedRoundRobinAlgorithm() *WeightedRoundRobinAlgorithm {
	return &WeightedRoundRobinAlgorithm{
		name:    "weighted_round_robin",
		metrics: &AlgorithmMetrics{LastUsed: time.Now()},
		weights: make(map[string]float64),
	}
}

func (wrr *WeightedRoundRobinAlgorithm) GetName() string {
	return wrr.name
}

func (wrr *WeightedRoundRobinAlgorithm) GetMetrics() *AlgorithmMetrics {
	return wrr.metrics
}

func (wrr *WeightedRoundRobinAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Update weights based on current node performance
	for _, node := range nodes {
		weight := wrr.calculateWeight(node)
		wrr.weights[node.ID] = weight
	}
	
	// Select node using weighted round-robin
	totalWeight := 0.0
	for _, weight := range wrr.weights {
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		// Fallback to simple round-robin
		selectedIndex := wrr.counter % len(nodes)
		wrr.counter++
		return []*NodeInfo{nodes[selectedIndex]}, nil
	}
	
	// Weighted selection
	target := rand.Float64() * totalWeight
	currentSum := 0.0
	
	for _, node := range nodes {
		currentSum += wrr.weights[node.ID]
		if currentSum >= target {
			return []*NodeInfo{node}, nil
		}
	}
	
	// Fallback to first node
	return []*NodeInfo{nodes[0]}, nil
}

func (wrr *WeightedRoundRobinAlgorithm) calculateWeight(node *NodeInfo) float64 {
	// Weight based on capacity and inverse of utilization
	capacityScore := node.PerformanceScore
	utilizationPenalty := (node.Usage.CPUUtilization + node.Usage.MemoryUtilization) / 2.0
	healthBonus := node.HealthScore
	
	weight := capacityScore * healthBonus * (1.0 - utilizationPenalty)
	return math.Max(weight, 0.1) // Minimum weight
}

func (wrr *WeightedRoundRobinAlgorithm) UpdateMetrics(result *SelectionResult) {
	wrr.metrics.Selections++
	wrr.metrics.LastUsed = time.Now()
	
	if result.Successful {
		wrr.metrics.SuccessRate = float64(wrr.metrics.Selections) / float64(wrr.metrics.Selections)
		wrr.metrics.AverageLatency = (wrr.metrics.AverageLatency + result.ExecutionLatency) / 2
		wrr.metrics.Throughput = (wrr.metrics.Throughput + result.Throughput) / 2
	}
}

// LeastEffectiveLoadAlgorithm implements least effective load balancing
type LeastEffectiveLoadAlgorithm struct {
	name    string
	metrics *AlgorithmMetrics
}

// NewLeastEffectiveLoadAlgorithm creates a new least effective load algorithm
func NewLeastEffectiveLoadAlgorithm() *LeastEffectiveLoadAlgorithm {
	return &LeastEffectiveLoadAlgorithm{
		name:    "least_effective_load",
		metrics: &AlgorithmMetrics{LastUsed: time.Now()},
	}
}

func (lel *LeastEffectiveLoadAlgorithm) GetName() string {
	return lel.name
}

func (lel *LeastEffectiveLoadAlgorithm) GetMetrics() *AlgorithmMetrics {
	return lel.metrics
}

func (lel *LeastEffectiveLoadAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Calculate effective load for each node
	type nodeScore struct {
		node         *NodeInfo
		effectiveLoad float64
	}
	
	scores := make([]nodeScore, len(nodes))
	for i, node := range nodes {
		scores[i] = nodeScore{
			node:         node,
			effectiveLoad: lel.calculateEffectiveLoad(node),
		}
	}
	
	// Sort by effective load (ascending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].effectiveLoad < scores[j].effectiveLoad
	})
	
	// Select the node with least effective load
	return []*NodeInfo{scores[0].node}, nil
}

func (lel *LeastEffectiveLoadAlgorithm) calculateEffectiveLoad(node *NodeInfo) float64 {
	// Effective load considers both utilization and capacity
	cpuLoad := node.Usage.CPUUtilization / math.Max(float64(node.Capacity.CPUCores), 1.0)
	memoryLoad := node.Usage.MemoryUtilization
	gpuLoad := node.Usage.GPUUtilization
	networkLoad := node.Usage.NetworkUtilization
	
	// Queue load
	queueLoad := float64(node.Usage.ActiveRequests+node.Usage.QueuedRequests) / 10.0
	
	// Weighted effective load
	effectiveLoad := 0.3*cpuLoad + 0.3*memoryLoad + 0.2*gpuLoad + 0.1*networkLoad + 0.1*queueLoad
	
	// Adjust for health score
	effectiveLoad = effectiveLoad / math.Max(node.HealthScore, 0.1)
	
	return effectiveLoad
}

func (lel *LeastEffectiveLoadAlgorithm) UpdateMetrics(result *SelectionResult) {
	lel.metrics.Selections++
	lel.metrics.LastUsed = time.Now()
	
	if result.Successful {
		lel.metrics.SuccessRate = float64(lel.metrics.Selections) / float64(lel.metrics.Selections)
		lel.metrics.AverageLatency = (lel.metrics.AverageLatency + result.ExecutionLatency) / 2
		lel.metrics.Throughput = (lel.metrics.Throughput + result.Throughput) / 2
	}
}

// LocalityAwareAlgorithm implements locality-aware scheduling
type LocalityAwareAlgorithm struct {
	name    string
	metrics *AlgorithmMetrics
	cache   map[string][]string // model -> preferred nodes
}

// NewLocalityAwareAlgorithm creates a new locality-aware algorithm
func NewLocalityAwareAlgorithm() *LocalityAwareAlgorithm {
	return &LocalityAwareAlgorithm{
		name:    "locality_aware",
		metrics: &AlgorithmMetrics{LastUsed: time.Now()},
		cache:   make(map[string][]string),
	}
}

func (laa *LocalityAwareAlgorithm) GetName() string {
	return laa.name
}

func (laa *LocalityAwareAlgorithm) GetMetrics() *AlgorithmMetrics {
	return laa.metrics
}

func (laa *LocalityAwareAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Try to determine model name from task
	modelName := laa.getModelName(task)
	
	// Check cache for preferred nodes
	if preferredNodes, exists := laa.cache[modelName]; exists {
		for _, preferredNodeID := range preferredNodes {
			for _, node := range nodes {
				if node.ID == preferredNodeID {
					// Check if node is still suitable
					if laa.isSuitableNode(node) {
						return []*NodeInfo{node}, nil
					}
				}
			}
		}
	}
	
	// No cached preference or preferred nodes unavailable
	// Select based on locality factors
	type nodeScore struct {
		node         *NodeInfo
		localityScore float64
	}
	
	scores := make([]nodeScore, len(nodes))
	for i, node := range nodes {
		scores[i] = nodeScore{
			node:         node,
			localityScore: laa.calculateLocalityScore(node, modelName),
		}
	}
	
	// Sort by locality score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].localityScore > scores[j].localityScore
	})
	
	// Update cache with selected node
	selectedNode := scores[0].node
	laa.updateCache(modelName, selectedNode.ID)
	
	return []*NodeInfo{selectedNode}, nil
}

func (laa *LocalityAwareAlgorithm) getModelName(task interface{}) string {
	// Extract model name from task
	// This is a simplified implementation
	return "default_model"
}

func (laa *LocalityAwareAlgorithm) isSuitableNode(node *NodeInfo) bool {
	// Check if node is suitable for execution
	return node.HealthScore > 0.5 && 
		   node.Usage.CPUUtilization < 0.9 && 
		   node.Usage.MemoryUtilization < 0.9
}

func (laa *LocalityAwareAlgorithm) calculateLocalityScore(node *NodeInfo, modelName string) float64 {
	// Locality score based on:
	// 1. Model cache hit (if model is already loaded)
	// 2. Network latency
	// 3. Data locality
	// 4. Session affinity
	
	// Check if model is cached on this node
	modelCacheHit := laa.hasModelCached(node, modelName)
	
	// Network latency score (inverse of latency)
	latencyScore := 1.0 / (1.0 + float64(node.Latency)/float64(time.Millisecond))
	
	// Data locality score (placeholder)
	dataLocalityScore := 0.8
	
	// Session affinity score (placeholder)
	sessionAffinityScore := 0.7
	
	// Combine scores
	localityScore := 0.4*modelCacheHit + 0.3*latencyScore + 0.2*dataLocalityScore + 0.1*sessionAffinityScore
	
	return localityScore
}

func (laa *LocalityAwareAlgorithm) hasModelCached(node *NodeInfo, modelName string) float64 {
	// Check if model is cached on the node
	// This would interface with the actual model cache
	// For now, return a placeholder value
	return 0.5
}

func (laa *LocalityAwareAlgorithm) updateCache(modelName, nodeID string) {
	if _, exists := laa.cache[modelName]; !exists {
		laa.cache[modelName] = make([]string, 0)
	}
	
	// Add node to preferred list if not already present
	for _, id := range laa.cache[modelName] {
		if id == nodeID {
			return
		}
	}
	
	laa.cache[modelName] = append(laa.cache[modelName], nodeID)
	
	// Keep only top 3 preferred nodes
	if len(laa.cache[modelName]) > 3 {
		laa.cache[modelName] = laa.cache[modelName][:3]
	}
}

func (laa *LocalityAwareAlgorithm) UpdateMetrics(result *SelectionResult) {
	laa.metrics.Selections++
	laa.metrics.LastUsed = time.Now()
	
	if result.Successful {
		laa.metrics.SuccessRate = float64(laa.metrics.Selections) / float64(laa.metrics.Selections)
		laa.metrics.AverageLatency = (laa.metrics.AverageLatency + result.ExecutionLatency) / 2
		laa.metrics.Throughput = (laa.metrics.Throughput + result.Throughput) / 2
	}
}

// PredictiveAlgorithm implements predictive load balancing
type PredictiveAlgorithm struct {
	name      string
	metrics   *AlgorithmMetrics
	predictor *PerformancePredictor
}

// NewPredictiveAlgorithm creates a new predictive algorithm
func NewPredictiveAlgorithm(predictor *PerformancePredictor) *PredictiveAlgorithm {
	return &PredictiveAlgorithm{
		name:      "predictive",
		metrics:   &AlgorithmMetrics{LastUsed: time.Now()},
		predictor: predictor,
	}
}

func (pa *PredictiveAlgorithm) GetName() string {
	return pa.name
}

func (pa *PredictiveAlgorithm) GetMetrics() *AlgorithmMetrics {
	return pa.metrics
}

func (pa *PredictiveAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	taskType := pa.getTaskType(task)
	
	// Get predictions for all nodes
	type nodePrediction struct {
		node               *NodeInfo
		predictedLatency   time.Duration
		predictedThroughput float64
		predictionScore    float64
	}
	
	predictions := make([]nodePrediction, len(nodes))
	for i, node := range nodes {
		latency, throughput := pa.predictor.PredictPerformance(node, taskType)
		score := pa.calculatePredictionScore(latency, throughput)
		
		predictions[i] = nodePrediction{
			node:               node,
			predictedLatency:   latency,
			predictedThroughput: throughput,
			predictionScore:    score,
		}
	}
	
	// Sort by prediction score (descending)
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].predictionScore > predictions[j].predictionScore
	})
	
	// Select the node with best predicted performance
	return []*NodeInfo{predictions[0].node}, nil
}

func (pa *PredictiveAlgorithm) getTaskType(task interface{}) string {
	// Extract task type from task
	// This is a simplified implementation
	return "inference"
}

func (pa *PredictiveAlgorithm) calculatePredictionScore(latency time.Duration, throughput float64) float64 {
	// Score based on predicted performance
	latencyScore := 1.0 / (1.0 + float64(latency)/float64(time.Second))
	throughputScore := throughput / 100.0 // Normalize to 100 ops/sec
	
	// Weighted combination
	return 0.6*latencyScore + 0.4*throughputScore
}

func (pa *PredictiveAlgorithm) UpdateMetrics(result *SelectionResult) {
	pa.metrics.Selections++
	pa.metrics.LastUsed = time.Now()
	
	if result.Successful {
		pa.metrics.SuccessRate = float64(pa.metrics.Selections) / float64(pa.metrics.Selections)
		pa.metrics.AverageLatency = (pa.metrics.AverageLatency + result.ExecutionLatency) / 2
		pa.metrics.Throughput = (pa.metrics.Throughput + result.Throughput) / 2
	}
}

// AdaptiveAlgorithm implements adaptive load balancing
type AdaptiveAlgorithm struct {
	name    string
	metrics *AlgorithmMetrics
	history *RequestHistory
	strategies []string
	currentStrategy string
}

// NewAdaptiveAlgorithm creates a new adaptive algorithm
func NewAdaptiveAlgorithm(history *RequestHistory) *AdaptiveAlgorithm {
	return &AdaptiveAlgorithm{
		name:    "adaptive",
		metrics: &AlgorithmMetrics{LastUsed: time.Now()},
		history: history,
		strategies: []string{"round_robin", "least_load", "locality_aware"},
		currentStrategy: "round_robin",
	}
}

func (aa *AdaptiveAlgorithm) GetName() string {
	return aa.name
}

func (aa *AdaptiveAlgorithm) GetMetrics() *AlgorithmMetrics {
	return aa.metrics
}

func (aa *AdaptiveAlgorithm) SelectNodes(task interface{}, nodes []*NodeInfo) ([]*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}
	
	// Adapt strategy based on current conditions
	aa.adaptStrategy(nodes)
	
	// Select node based on current strategy
	switch aa.currentStrategy {
	case "round_robin":
		return aa.selectRoundRobin(nodes)
	case "least_load":
		return aa.selectLeastLoad(nodes)
	case "locality_aware":
		return aa.selectLocalityAware(nodes)
	default:
		return []*NodeInfo{nodes[0]}, nil
	}
}

func (aa *AdaptiveAlgorithm) adaptStrategy(nodes []*NodeInfo) {
	// Analyze current system state
	loadVariance := aa.calculateLoadVariance(nodes)
	latencyVariance := aa.calculateLatencyVariance(nodes)
	
	// Adapt strategy based on conditions
	if loadVariance > 0.5 {
		aa.currentStrategy = "least_load"
	} else if latencyVariance > 0.3 {
		aa.currentStrategy = "locality_aware"
	} else {
		aa.currentStrategy = "round_robin"
	}
}

func (aa *AdaptiveAlgorithm) calculateLoadVariance(nodes []*NodeInfo) float64 {
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

func (aa *AdaptiveAlgorithm) calculateLatencyVariance(nodes []*NodeInfo) float64 {
	if len(nodes) == 0 {
		return 0.0
	}
	
	// Calculate average latency
	totalLatency := time.Duration(0)
	for _, node := range nodes {
		totalLatency += node.Latency
	}
	averageLatency := totalLatency / time.Duration(len(nodes))
	
	// Calculate variance
	variance := 0.0
	for _, node := range nodes {
		deviation := float64(node.Latency - averageLatency)
		variance += deviation * deviation
	}
	
	return variance / float64(len(nodes))
}

func (aa *AdaptiveAlgorithm) selectRoundRobin(nodes []*NodeInfo) ([]*NodeInfo, error) {
	// Simple round-robin selection
	static := struct {
		counter int
	}{}
	selectedIndex := static.counter % len(nodes)
	static.counter++
	return []*NodeInfo{nodes[selectedIndex]}, nil
}

func (aa *AdaptiveAlgorithm) selectLeastLoad(nodes []*NodeInfo) ([]*NodeInfo, error) {
	// Select node with least load
	minLoad := math.MaxFloat64
	var selectedNode *NodeInfo
	
	for _, node := range nodes {
		if node.LoadScore < minLoad {
			minLoad = node.LoadScore
			selectedNode = node
		}
	}
	
	return []*NodeInfo{selectedNode}, nil
}

func (aa *AdaptiveAlgorithm) selectLocalityAware(nodes []*NodeInfo) ([]*NodeInfo, error) {
	// Select node with lowest latency
	minLatency := time.Duration(math.MaxInt64)
	var selectedNode *NodeInfo
	
	for _, node := range nodes {
		if node.Latency < minLatency {
			minLatency = node.Latency
			selectedNode = node
		}
	}
	
	return []*NodeInfo{selectedNode}, nil
}

func (aa *AdaptiveAlgorithm) UpdateMetrics(result *SelectionResult) {
	aa.metrics.Selections++
	aa.metrics.LastUsed = time.Now()
	
	if result.Successful {
		aa.metrics.SuccessRate = float64(aa.metrics.Selections) / float64(aa.metrics.Selections)
		aa.metrics.AverageLatency = (aa.metrics.AverageLatency + result.ExecutionLatency) / 2
		aa.metrics.Throughput = (aa.metrics.Throughput + result.Throughput) / 2
	}
}
