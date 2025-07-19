package distributed

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/discover"
)

// LoadBalancer interfaces for different load balancing strategies
type LoadBalancer interface {
	SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error)
	UpdateMetrics(nodeID string, metrics *NodeMetrics)
	GetMetrics() map[string]*NodeMetrics
	SetConstraints(constraints []LoadBalancingConstraint)
}

// IntelligentLoadBalancer implements ML-based load balancing
type IntelligentLoadBalancer struct {
	algorithm     string
	metrics       *MetricsCollector
	predictor     *PerformancePredictor
	history       *RequestHistory
	constraints   []LoadBalancingConstraint
	nodeMetrics   map[string]*NodeMetrics
	metricsMutex  sync.RWMutex
	roundRobinIdx int
	rrMutex       sync.Mutex
}

// LoadBalancingConstraint defines constraints for load balancing decisions
type LoadBalancingConstraint struct {
	Type     string      // "memory", "gpu", "latency", "cost", "locality"
	Value    interface{} // constraint value
	Priority int         // higher priority = more important
	Weight   float64     // weight in scoring function
}

// NodeMetrics tracks performance metrics for each node
type NodeMetrics struct {
	NodeID           string
	RequestCount     int64
	ActiveRequests   int64
	AverageLatency   time.Duration
	SuccessRate      float64
	ErrorRate        float64
	CPUUtilization   float64
	GPUUtilization   float64
	MemoryUtilization float64
	NetworkBandwidth int64
	QueueLength      int
	LastUpdated      time.Time
	PredictedLatency time.Duration
	LoadScore        float64
}

// PerformancePredictor predicts node performance based on historical data
type PerformancePredictor struct {
	models        map[string]*PredictionModel
	features      []string
	updateTicker  *time.Ticker
	trainingData  map[string][]TrainingExample
	mutex         sync.RWMutex
}

type PredictionModel struct {
	Weights    []float64
	Bias       float64
	Accuracy   float64
	LastTrained time.Time
}

type TrainingExample struct {
	Features []float64
	Target   float64
	Timestamp time.Time
}

// RequestHistory maintains history of requests for learning
type RequestHistory struct {
	requests      []HistoricalRequest
	maxSize       int
	mutex         sync.RWMutex
	aggregateData map[string]*AggregateMetrics
}

type HistoricalRequest struct {
	ID            string
	NodeID        string
	ModelName     string
	BatchSize     int
	SequenceLength int
	Latency       time.Duration
	Success       bool
	Timestamp     time.Time
	Features      []float64
}

type AggregateMetrics struct {
	AverageLatency time.Duration
	SuccessRate    float64
	Throughput     float64
	LastUpdated    time.Time
}

// MetricsCollector collects and aggregates metrics
type MetricsCollector struct {
	nodeMetrics   map[string]*NodeMetrics
	systemMetrics *SystemMetrics
	collectors    []MetricCollector
	mutex         sync.RWMutex
}

type SystemMetrics struct {
	TotalRequests    int64
	TotalLatency     time.Duration
	AverageLatency   time.Duration
	SystemThroughput float64
	ErrorRate        float64
	ActiveNodes      int
	LastUpdated      time.Time
}

type MetricCollector interface {
	CollectMetrics(nodeID string) (*NodeMetrics, error)
	GetSystemMetrics() (*SystemMetrics, error)
}

// NewIntelligentLoadBalancer creates a new intelligent load balancer
func NewIntelligentLoadBalancer(algorithm string) *IntelligentLoadBalancer {
	lb := &IntelligentLoadBalancer{
		algorithm:   algorithm,
		metrics:     NewMetricsCollector(),
		predictor:   NewPerformancePredictor(),
		history:     NewRequestHistory(10000),
		constraints: make([]LoadBalancingConstraint, 0),
		nodeMetrics: make(map[string]*NodeMetrics),
	}

	// Start background metric collection
	go lb.metricsCollectionLoop()

	return lb
}

// SelectNode selects the best node for a request
func (lb *IntelligentLoadBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Filter nodes based on constraints
	candidateNodes := lb.filterNodes(nodes, request)
	if len(candidateNodes) == 0 {
		return nil, fmt.Errorf("no nodes meet the constraints")
	}

	// Apply load balancing algorithm
	switch lb.algorithm {
	case "intelligent":
		return lb.intelligentSelection(ctx, request, candidateNodes)
	case "weighted_round_robin":
		return lb.weightedRoundRobin(candidateNodes)
	case "least_effective_load":
		return lb.leastEffectiveLoad(candidateNodes)
	case "locality_aware":
		return lb.localityAwareSelection(request, candidateNodes)
	case "predictive":
		return lb.predictiveSelection(request, candidateNodes)
	default:
		return lb.intelligentSelection(ctx, request, candidateNodes)
	}
}

// intelligentSelection uses ML-based node selection
func (lb *IntelligentLoadBalancer) intelligentSelection(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	scores := make([]NodeScore, 0, len(nodes))

	for _, node := range nodes {
		score := lb.calculateNodeScore(request, &node)
		scores = append(scores, NodeScore{
			Node:  &node,
			Score: score,
		})
	}

	// Sort by score (higher is better)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Select top node with some randomization to avoid hotspots
	if len(scores) > 1 && scores[0].Score-scores[1].Score < 0.1 {
		// Scores are close, add randomization
		if rand.Float64() < 0.3 {
			return scores[1].Node, nil
		}
	}

	return scores[0].Node, nil
}

type NodeScore struct {
	Node  *NodeInfo
	Score float64
}

// calculateNodeScore calculates a comprehensive score for a node
func (lb *IntelligentLoadBalancer) calculateNodeScore(request *InferenceRequest, node *NodeInfo) float64 {
	lb.metricsMutex.RLock()
	metrics, exists := lb.nodeMetrics[node.ID]
	lb.metricsMutex.RUnlock()

	if !exists {
		// New node, give it a chance
		return 0.5
	}

	score := 0.0

	// Latency factor (lower is better)
	latencyFactor := 1.0
	if metrics.AverageLatency > 0 {
		latencyFactor = 1.0 / (1.0 + float64(metrics.AverageLatency.Milliseconds())/1000.0)
	}
	score += latencyFactor * 0.3

	// Load factor (lower is better)
	loadFactor := 1.0 - (float64(metrics.ActiveRequests) / 100.0) // Assume max 100 concurrent requests
	if loadFactor < 0 {
		loadFactor = 0
	}
	score += loadFactor * 0.25

	// Success rate factor
	score += metrics.SuccessRate * 0.2

	// Resource utilization factor (prefer nodes with available resources)
	resourceFactor := 1.0 - (metrics.CPUUtilization + metrics.GPUUtilization + metrics.MemoryUtilization) / 3.0
	if resourceFactor < 0 {
		resourceFactor = 0
	}
	score += resourceFactor * 0.15

	// Prediction factor
	if metrics.PredictedLatency > 0 {
		predictionFactor := 1.0 / (1.0 + float64(metrics.PredictedLatency.Milliseconds())/1000.0)
		score += predictionFactor * 0.1
	}

	// Apply constraints
	for _, constraint := range lb.constraints {
		constraintScore := lb.evaluateConstraint(constraint, node, metrics)
		score += constraintScore * constraint.Weight
	}

	return score
}

// evaluateConstraint evaluates a constraint against a node
func (lb *IntelligentLoadBalancer) evaluateConstraint(constraint LoadBalancingConstraint, node *NodeInfo, metrics *NodeMetrics) float64 {
	switch constraint.Type {
	case "memory":
		if threshold, ok := constraint.Value.(float64); ok {
			if metrics.MemoryUtilization <= threshold {
				return 1.0
			}
			return 0.0
		}
	case "gpu":
		if threshold, ok := constraint.Value.(float64); ok {
			if metrics.GPUUtilization <= threshold {
				return 1.0
			}
			return 0.0
		}
	case "latency":
		if threshold, ok := constraint.Value.(time.Duration); ok {
			if metrics.AverageLatency <= threshold {
				return 1.0
			}
			return 0.0
		}
	case "locality":
		// Implement locality-based scoring
		return 0.5 // Placeholder
	}
	return 0.0
}

// weightedRoundRobin implements weighted round-robin load balancing
func (lb *IntelligentLoadBalancer) weightedRoundRobin(nodes []NodeInfo) (*NodeInfo, error) {
	lb.rrMutex.Lock()
	defer lb.rrMutex.Unlock()

	// Calculate weights based on node capabilities
	weights := make([]float64, len(nodes))
	totalWeight := 0.0

	for i, node := range nodes {
		weight := lb.calculateNodeWeight(&node)
		weights[i] = weight
		totalWeight += weight
	}

	// Weighted selection
	target := rand.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if cumulative >= target {
			return &nodes[i], nil
		}
	}

	// Fallback to round-robin
	idx := lb.roundRobinIdx % len(nodes)
	lb.roundRobinIdx++
	return &nodes[idx], nil
}

// calculateNodeWeight calculates weight for weighted round-robin
func (lb *IntelligentLoadBalancer) calculateNodeWeight(node *NodeInfo) float64 {
	lb.metricsMutex.RLock()
	metrics, exists := lb.nodeMetrics[node.ID]
	lb.metricsMutex.RUnlock()

	if !exists {
		return 1.0 // Default weight
	}

	// Weight based on capacity and current load
	capacity := float64(node.Capacity.GPU + node.Capacity.CPU + node.Capacity.Memory)
	utilization := (metrics.CPUUtilization + metrics.GPUUtilization + metrics.MemoryUtilization) / 3.0
	
	weight := capacity * (1.0 - utilization) * metrics.SuccessRate
	if weight < 0.1 {
		weight = 0.1
	}

	return weight
}

// leastEffectiveLoad selects node with least effective load
func (lb *IntelligentLoadBalancer) leastEffectiveLoad(nodes []NodeInfo) (*NodeInfo, error) {
	bestNode := &nodes[0]
	minLoad := math.MaxFloat64

	for _, node := range nodes {
		effectiveLoad := lb.calculateEffectiveLoad(&node)
		if effectiveLoad < minLoad {
			minLoad = effectiveLoad
			bestNode = &node
		}
	}

	return bestNode, nil
}

// calculateEffectiveLoad calculates effective load considering heterogeneous hardware
func (lb *IntelligentLoadBalancer) calculateEffectiveLoad(node *NodeInfo) float64 {
	lb.metricsMutex.RLock()
	metrics, exists := lb.nodeMetrics[node.ID]
	lb.metricsMutex.RUnlock()

	if !exists {
		return 0.0
	}

	// Normalize by node capacity
	gpuLoad := metrics.GPUUtilization * (float64(node.Capacity.GPU) / 1e9)
	cpuLoad := metrics.CPUUtilization * (float64(node.Capacity.CPU) / 1e3)
	memLoad := metrics.MemoryUtilization * (float64(node.Capacity.Memory) / 1e9)

	return gpuLoad + cpuLoad + memLoad + float64(metrics.QueueLength)
}

// localityAwareSelection prioritizes nodes with cached models
func (lb *IntelligentLoadBalancer) localityAwareSelection(request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	// Separate nodes with and without the model
	withModel := make([]NodeInfo, 0)
	withoutModel := make([]NodeInfo, 0)

	for _, node := range nodes {
		hasModel := false
		for _, modelName := range node.Models {
			if modelName == request.Model.Name {
				hasModel = true
				break
			}
		}

		if hasModel {
			withModel = append(withModel, node)
		} else {
			withoutModel = append(withoutModel, node)
		}
	}

	// Prefer nodes with the model
	if len(withModel) > 0 {
		return lb.leastEffectiveLoad(withModel)
	}

	// Fall back to nodes without the model
	return lb.leastEffectiveLoad(withoutModel)
}

// predictiveSelection uses ML predictions for node selection
func (lb *IntelligentLoadBalancer) predictiveSelection(request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	bestNode := &nodes[0]
	minPredictedLatency := time.Duration(math.MaxInt64)

	for _, node := range nodes {
		predictedLatency := lb.predictor.PredictLatency(request, &node)
		if predictedLatency < minPredictedLatency {
			minPredictedLatency = predictedLatency
			bestNode = &node
		}
	}

	return bestNode, nil
}

// filterNodes filters nodes based on constraints
func (lb *IntelligentLoadBalancer) filterNodes(nodes []NodeInfo, request *InferenceRequest) []NodeInfo {
	filtered := make([]NodeInfo, 0)

	for _, node := range nodes {
		if lb.nodePassesConstraints(&node, request) {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

// nodePassesConstraints checks if a node meets all constraints
func (lb *IntelligentLoadBalancer) nodePassesConstraints(node *NodeInfo, request *InferenceRequest) bool {
	lb.metricsMutex.RLock()
	metrics, exists := lb.nodeMetrics[node.ID]
	lb.metricsMutex.RUnlock()

	if !exists {
		return true // New node, assume it passes
	}

	for _, constraint := range lb.constraints {
		if !lb.checkConstraint(constraint, node, metrics) {
			return false
		}
	}

	return true
}

// checkConstraint checks if a node meets a specific constraint
func (lb *IntelligentLoadBalancer) checkConstraint(constraint LoadBalancingConstraint, node *NodeInfo, metrics *NodeMetrics) bool {
	switch constraint.Type {
	case "memory":
		if threshold, ok := constraint.Value.(float64); ok {
			return metrics.MemoryUtilization <= threshold
		}
	case "gpu":
		if threshold, ok := constraint.Value.(float64); ok {
			return metrics.GPUUtilization <= threshold
		}
	case "latency":
		if threshold, ok := constraint.Value.(time.Duration); ok {
			return metrics.AverageLatency <= threshold
		}
	case "cost":
		// Implement cost-based constraints
		return true
	}
	return true
}

// UpdateMetrics updates metrics for a node
func (lb *IntelligentLoadBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	lb.metricsMutex.Lock()
	defer lb.metricsMutex.Unlock()

	metrics.LastUpdated = time.Now()
	lb.nodeMetrics[nodeID] = metrics

	// Update predictor with new data
	lb.predictor.UpdateMetrics(nodeID, metrics)
}

// GetMetrics returns current metrics
func (lb *IntelligentLoadBalancer) GetMetrics() map[string]*NodeMetrics {
	lb.metricsMutex.RLock()
	defer lb.metricsMutex.RUnlock()

	metrics := make(map[string]*NodeMetrics)
	for k, v := range lb.nodeMetrics {
		metrics[k] = v
	}

	return metrics
}

// SetConstraints sets load balancing constraints
func (lb *IntelligentLoadBalancer) SetConstraints(constraints []LoadBalancingConstraint) {
	lb.constraints = constraints
}

// metricsCollectionLoop runs the background metrics collection
func (lb *IntelligentLoadBalancer) metricsCollectionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lb.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all nodes
func (lb *IntelligentLoadBalancer) collectMetrics() {
	// Implementation would collect actual metrics from nodes
	// For now, this is a placeholder
}

// PerformancePredictor implementations

// NewPerformancePredictor creates a new performance predictor
func NewPerformancePredictor() *PerformancePredictor {
	return &PerformancePredictor{
		models:       make(map[string]*PredictionModel),
		features:     []string{"batch_size", "sequence_length", "model_size", "gpu_util", "cpu_util", "memory_util"},
		trainingData: make(map[string][]TrainingExample),
	}
}

// PredictLatency predicts latency for a request on a node
func (pp *PerformancePredictor) PredictLatency(request *InferenceRequest, node *NodeInfo) time.Duration {
	pp.mutex.RLock()
	model, exists := pp.models[node.ID]
	pp.mutex.RUnlock()

	if !exists {
		return time.Duration(500) * time.Millisecond // Default prediction
	}

	features := pp.extractFeatures(request, node)
	prediction := pp.predict(model, features)

	return time.Duration(prediction) * time.Millisecond
}

// extractFeatures extracts features from request and node
func (pp *PerformancePredictor) extractFeatures(request *InferenceRequest, node *NodeInfo) []float64 {
	features := make([]float64, len(pp.features))
	
	// Extract feature values
	features[0] = float64(request.BatchSize)
	features[1] = float64(request.SequenceLength)
	features[2] = float64(request.Model.Size)
	// GPU, CPU, Memory utilization would be fetched from current metrics
	features[3] = 0.5 // Placeholder
	features[4] = 0.5 // Placeholder
	features[5] = 0.5 // Placeholder

	return features
}

// predict makes a prediction using the model
func (pp *PerformancePredictor) predict(model *PredictionModel, features []float64) float64 {
	prediction := model.Bias
	for i, feature := range features {
		if i < len(model.Weights) {
			prediction += feature * model.Weights[i]
		}
	}
	return prediction
}

// UpdateMetrics updates the predictor with new metrics
func (pp *PerformancePredictor) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	// Add training example
	if _, exists := pp.trainingData[nodeID]; !exists {
		pp.trainingData[nodeID] = make([]TrainingExample, 0)
	}

	// This would be populated with actual feature extraction
	example := TrainingExample{
		Features:  make([]float64, len(pp.features)),
		Target:    float64(metrics.AverageLatency.Milliseconds()),
		Timestamp: time.Now(),
	}

	pp.trainingData[nodeID] = append(pp.trainingData[nodeID], example)

	// Retrain model periodically
	if len(pp.trainingData[nodeID]) % 100 == 0 {
		pp.retrainModel(nodeID)
	}
}

// retrainModel retrains the prediction model for a node
func (pp *PerformancePredictor) retrainModel(nodeID string) {
	examples := pp.trainingData[nodeID]
	if len(examples) < 10 {
		return
	}

	// Simple linear regression implementation
	model := &PredictionModel{
		Weights:     make([]float64, len(pp.features)),
		Bias:        0.0,
		LastTrained: time.Now(),
	}

	// Implement training algorithm (simplified)
	// This would use proper ML algorithms in production
	model.Weights[0] = 1.0 // Placeholder weights
	model.Bias = 100.0

	pp.models[nodeID] = model
}

// Supporting functions and types

// NewRequestHistory creates a new request history
func NewRequestHistory(maxSize int) *RequestHistory {
	return &RequestHistory{
		requests:      make([]HistoricalRequest, 0, maxSize),
		maxSize:       maxSize,
		aggregateData: make(map[string]*AggregateMetrics),
	}
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		nodeMetrics:   make(map[string]*NodeMetrics),
		systemMetrics: &SystemMetrics{},
		collectors:    make([]MetricCollector, 0),
	}
}

// Additional helper types and functions would be implemented here...