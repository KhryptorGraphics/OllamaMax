package loadbalancer

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/types"
)

// LoadBalancer manages advanced load balancing across nodes
type LoadBalancer struct {
	mu sync.RWMutex

	// Node management
	nodes       map[string]*LoadBalancedNode
	nodeMetrics map[string]*NodeLoadMetrics

	// Load balancing strategies
	strategies      map[string]LoadBalancingStrategy
	currentStrategy string

	// Configuration
	config *LoadBalancerConfig

	// Metrics and monitoring
	metrics     *AdvancedLoadBalancerMetrics
	loadHistory []*LoadSnapshot

	// Predictive modeling
	predictor *LoadPredictor

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// LoadBalancedNode represents a node with load balancing information
type LoadBalancedNode struct {
	NodeID      string              `json:"node_id"`
	Capacity    *types.NodeCapacity `json:"capacity"`
	CurrentLoad *NodeLoadMetrics    `json:"current_load"`

	// Load balancing weights
	Weight   float64 `json:"weight"`
	Priority int     `json:"priority"`

	// Performance characteristics
	Throughput  float64       `json:"throughput"`
	Latency     time.Duration `json:"latency"`
	Reliability float64       `json:"reliability"`

	// Load distribution
	AssignedTasks  int   `json:"assigned_tasks"`
	CompletedTasks int64 `json:"completed_tasks"`
	FailedTasks    int64 `json:"failed_tasks"`

	// Health and availability
	HealthScore float64   `json:"health_score"`
	Available   bool      `json:"available"`
	LastUpdate  time.Time `json:"last_update"`

	// Load prediction
	PredictedLoad float64   `json:"predicted_load"`
	LoadTrend     LoadTrend `json:"load_trend"`
}

// NodeLoadMetrics represents detailed load metrics for a node
type NodeLoadMetrics struct {
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`

	// Resource utilization
	CPUUtilization     float64 `json:"cpu_utilization"`
	MemoryUtilization  float64 `json:"memory_utilization"`
	DiskUtilization    float64 `json:"disk_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`

	// Composite load scores
	OverallLoad    float64 `json:"overall_load"`
	WeightedLoad   float64 `json:"weighted_load"`
	NormalizedLoad float64 `json:"normalized_load"`

	// Performance metrics
	RequestRate  float64       `json:"request_rate"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorRate    float64       `json:"error_rate"`

	// Queue metrics
	QueueLength   int           `json:"queue_length"`
	QueueWaitTime time.Duration `json:"queue_wait_time"`

	// Trend indicators
	LoadVelocity     float64 `json:"load_velocity"`
	LoadAcceleration float64 `json:"load_acceleration"`
}

// LoadTrend represents the trend of load over time
type LoadTrend string

const (
	LoadTrendIncreasing LoadTrend = "increasing"
	LoadTrendDecreasing LoadTrend = "decreasing"
	LoadTrendStable     LoadTrend = "stable"
	LoadTrendVolatile   LoadTrend = "volatile"
)

// LoadSnapshot represents a snapshot of system load at a point in time
type LoadSnapshot struct {
	Timestamp     time.Time          `json:"timestamp"`
	TotalNodes    int                `json:"total_nodes"`
	ActiveNodes   int                `json:"active_nodes"`
	AverageLoad   float64            `json:"average_load"`
	LoadVariance  float64            `json:"load_variance"`
	LoadImbalance float64            `json:"load_imbalance"`
	NodeLoads     map[string]float64 `json:"node_loads"`
}

// LoadBalancingStrategy defines how load is balanced across nodes
type LoadBalancingStrategy interface {
	Name() string
	SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error)
	CalculateWeight(node *LoadBalancedNode) float64
	ShouldRebalance(nodes []*LoadBalancedNode) bool
}

// LoadBalancerConfig configures the load balancer
type LoadBalancerConfig struct {
	// Strategy settings
	DefaultStrategy        string
	RebalanceThreshold     float64
	LoadImbalanceThreshold float64

	// Monitoring settings
	MetricsInterval  time.Duration
	HistoryRetention time.Duration
	MaxHistorySize   int

	// Prediction settings
	EnablePrediction   bool
	PredictionWindow   time.Duration
	PredictionAccuracy float64

	// Performance settings
	MaxRebalanceFrequency time.Duration
	RebalanceBatchSize    int
	GracefulRebalance     bool

	// Thresholds
	HighLoadThreshold     float64
	LowLoadThreshold      float64
	CriticalLoadThreshold float64

	// Weights
	CPUWeight     float64
	MemoryWeight  float64
	DiskWeight    float64
	NetworkWeight float64
}

// AdvancedLoadBalancerMetrics tracks advanced load balancer performance
type AdvancedLoadBalancerMetrics struct {
	// Balancing metrics
	TotalRebalances      int64         `json:"total_rebalances"`
	SuccessfulRebalances int64         `json:"successful_rebalances"`
	FailedRebalances     int64         `json:"failed_rebalances"`
	AverageRebalanceTime time.Duration `json:"average_rebalance_time"`

	// Load distribution metrics
	CurrentImbalance float64 `json:"current_imbalance"`
	AverageImbalance float64 `json:"average_imbalance"`
	MaxImbalance     float64 `json:"max_imbalance"`
	LoadVariance     float64 `json:"load_variance"`

	// Performance metrics
	ThroughputImprovement float64 `json:"throughput_improvement"`
	LatencyReduction      float64 `json:"latency_reduction"`
	ResourceUtilization   float64 `json:"resource_utilization"`

	// Strategy metrics
	StrategyUsage         map[string]int64   `json:"strategy_usage"`
	StrategyEffectiveness map[string]float64 `json:"strategy_effectiveness"`

	// Prediction metrics
	PredictionAccuracy float64 `json:"prediction_accuracy"`
	PredictionErrors   int64   `json:"prediction_errors"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// LoadPredictor predicts future load based on historical data
type LoadPredictor struct {
	mu       sync.RWMutex
	enabled  bool
	window   time.Duration
	accuracy float64

	// Historical data
	loadHistory []*LoadSnapshot
	predictions map[string]*LoadPrediction

	// Model parameters
	trendWeight    float64
	seasonalWeight float64
	noiseThreshold float64
}

// LoadPrediction represents a load prediction for a node
type LoadPrediction struct {
	NodeID        string             `json:"node_id"`
	PredictedLoad float64            `json:"predicted_load"`
	Confidence    float64            `json:"confidence"`
	TimeHorizon   time.Duration      `json:"time_horizon"`
	CreatedAt     time.Time          `json:"created_at"`
	Factors       map[string]float64 `json:"factors"`
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(config *LoadBalancerConfig) *LoadBalancer {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &LoadBalancerConfig{
			DefaultStrategy:        "weighted_round_robin",
			RebalanceThreshold:     0.2,
			LoadImbalanceThreshold: 0.3,
			MetricsInterval:        10 * time.Second,
			HistoryRetention:       24 * time.Hour,
			MaxHistorySize:         1000,
			EnablePrediction:       true,
			PredictionWindow:       5 * time.Minute,
			PredictionAccuracy:     0.8,
			MaxRebalanceFrequency:  30 * time.Second,
			RebalanceBatchSize:     10,
			GracefulRebalance:      true,
			HighLoadThreshold:      0.8,
			LowLoadThreshold:       0.2,
			CriticalLoadThreshold:  0.95,
			CPUWeight:              0.4,
			MemoryWeight:           0.3,
			DiskWeight:             0.2,
			NetworkWeight:          0.1,
		}
	}

	lb := &LoadBalancer{
		nodes:           make(map[string]*LoadBalancedNode),
		nodeMetrics:     make(map[string]*NodeLoadMetrics),
		strategies:      make(map[string]LoadBalancingStrategy),
		currentStrategy: config.DefaultStrategy,
		config:          config,
		metrics: &AdvancedLoadBalancerMetrics{
			StrategyUsage:         make(map[string]int64),
			StrategyEffectiveness: make(map[string]float64),
		},
		loadHistory: make([]*LoadSnapshot, 0),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize predictor
	if config.EnablePrediction {
		lb.predictor = &LoadPredictor{
			enabled:        true,
			window:         config.PredictionWindow,
			accuracy:       config.PredictionAccuracy,
			loadHistory:    make([]*LoadSnapshot, 0),
			predictions:    make(map[string]*LoadPrediction),
			trendWeight:    0.6,
			seasonalWeight: 0.3,
			noiseThreshold: 0.1,
		}
	}

	// Register default strategies
	lb.registerDefaultStrategies()

	// Start background tasks
	lb.wg.Add(3)
	go lb.monitoringLoop()
	go lb.rebalancingLoop()
	go lb.predictionLoop()

	return lb
}

// RegisterNode registers a node for load balancing
func (lb *LoadBalancer) RegisterNode(nodeID string, capacity *types.NodeCapacity) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	node := &LoadBalancedNode{
		NodeID:      nodeID,
		Capacity:    capacity,
		Weight:      1.0,
		Priority:    1,
		Reliability: 1.0,
		Available:   true,
		LastUpdate:  time.Now(),
		LoadTrend:   LoadTrendStable,
	}

	lb.nodes[nodeID] = node
	lb.nodeMetrics[nodeID] = &NodeLoadMetrics{
		NodeID:    nodeID,
		Timestamp: time.Now(),
	}
}

// UpdateNodeMetrics updates load metrics for a node
func (lb *LoadBalancer) UpdateNodeMetrics(nodeID string, resourceMetrics *types.ResourceMetrics) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	node, exists := lb.nodes[nodeID]
	if !exists {
		return
	}

	// Calculate load metrics
	loadMetrics := lb.calculateLoadMetrics(resourceMetrics)
	lb.nodeMetrics[nodeID] = loadMetrics

	// Update node information
	node.CurrentLoad = loadMetrics
	node.LastUpdate = time.Now()

	// Update load trend
	lb.updateLoadTrend(node, loadMetrics)

	// Update predictions if enabled
	if lb.predictor != nil && lb.predictor.enabled {
		lb.updatePredictions(nodeID, loadMetrics)
	}
}

// calculateLoadMetrics calculates comprehensive load metrics from resource metrics
func (lb *LoadBalancer) calculateLoadMetrics(resourceMetrics *types.ResourceMetrics) *NodeLoadMetrics {
	loadMetrics := &NodeLoadMetrics{
		NodeID:             resourceMetrics.NodeID,
		Timestamp:          resourceMetrics.Timestamp,
		CPUUtilization:     resourceMetrics.CPUUsagePercent / 100.0,
		MemoryUtilization:  resourceMetrics.MemoryUsagePercent / 100.0,
		DiskUtilization:    resourceMetrics.DiskUsagePercent / 100.0,
		NetworkUtilization: lb.calculateNetworkUtilization(resourceMetrics),
	}

	// Calculate overall load using weighted average
	loadMetrics.OverallLoad = (loadMetrics.CPUUtilization * lb.config.CPUWeight) +
		(loadMetrics.MemoryUtilization * lb.config.MemoryWeight) +
		(loadMetrics.DiskUtilization * lb.config.DiskWeight) +
		(loadMetrics.NetworkUtilization * lb.config.NetworkWeight)

	// Calculate weighted load (considering node capacity)
	loadMetrics.WeightedLoad = loadMetrics.OverallLoad

	// Normalize load (0.0 to 1.0)
	loadMetrics.NormalizedLoad = math.Min(loadMetrics.OverallLoad, 1.0)

	return loadMetrics
}

// calculateNetworkUtilization calculates network utilization from metrics
func (lb *LoadBalancer) calculateNetworkUtilization(metrics *types.ResourceMetrics) float64 {
	// Simplified network utilization calculation
	// In a real implementation, you would have baseline network capacity
	totalBytes := metrics.NetworkInBytes + metrics.NetworkOutBytes

	// Assume 1 Gbps baseline capacity (125 MB/s)
	baselineCapacity := int64(125 * 1024 * 1024) // 125 MB/s

	if baselineCapacity == 0 {
		return 0.0
	}

	utilization := float64(totalBytes) / float64(baselineCapacity)
	return math.Min(utilization, 1.0)
}

// updateLoadTrend updates the load trend for a node
func (lb *LoadBalancer) updateLoadTrend(node *LoadBalancedNode, currentMetrics *NodeLoadMetrics) {
	if node.CurrentLoad == nil {
		node.LoadTrend = LoadTrendStable
		return
	}

	previousLoad := node.CurrentLoad.OverallLoad
	currentLoad := currentMetrics.OverallLoad

	loadDiff := currentLoad - previousLoad
	threshold := 0.05 // 5% threshold

	if math.Abs(loadDiff) < threshold {
		node.LoadTrend = LoadTrendStable
	} else if loadDiff > threshold {
		node.LoadTrend = LoadTrendIncreasing
	} else {
		node.LoadTrend = LoadTrendDecreasing
	}

	// Calculate load velocity and acceleration
	timeDiff := currentMetrics.Timestamp.Sub(node.CurrentLoad.Timestamp).Seconds()
	if timeDiff > 0 {
		currentMetrics.LoadVelocity = loadDiff / timeDiff

		if node.CurrentLoad.LoadVelocity != 0 {
			velocityDiff := currentMetrics.LoadVelocity - node.CurrentLoad.LoadVelocity
			currentMetrics.LoadAcceleration = velocityDiff / timeDiff
		}
	}
}

// SelectNode selects the best node for a new task based on current strategy
func (lb *LoadBalancer) SelectNode(taskLoad float64) (*LoadBalancedNode, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Get available nodes
	availableNodes := lb.getAvailableNodes()
	if len(availableNodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	// Get current strategy
	strategy := lb.getStrategy(lb.currentStrategy)
	if strategy == nil {
		return nil, fmt.Errorf("no load balancing strategy available")
	}

	// Select node using strategy
	selectedNode, err := strategy.SelectNode(availableNodes, taskLoad)
	if err != nil {
		return nil, fmt.Errorf("failed to select node: %w", err)
	}

	// Update metrics
	lb.metrics.StrategyUsage[strategy.Name()]++

	return selectedNode, nil
}

// getAvailableNodes returns nodes that are available for task assignment
func (lb *LoadBalancer) getAvailableNodes() []*LoadBalancedNode {
	var available []*LoadBalancedNode

	for _, node := range lb.nodes {
		if node.Available &&
			node.HealthScore > 0.5 &&
			time.Since(node.LastUpdate) < 5*time.Minute {
			available = append(available, node)
		}
	}

	return available
}

// getStrategy returns the load balancing strategy by name
func (lb *LoadBalancer) getStrategy(name string) LoadBalancingStrategy {
	if strategy, exists := lb.strategies[name]; exists {
		return strategy
	}

	// Return first available strategy as fallback
	for _, strategy := range lb.strategies {
		return strategy
	}

	return nil
}

// CalculateLoadImbalance calculates the current load imbalance across nodes
func (lb *LoadBalancer) CalculateLoadImbalance() float64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.nodes) < 2 {
		return 0.0
	}

	loads := make([]float64, 0, len(lb.nodes))
	totalLoad := 0.0

	for _, node := range lb.nodes {
		if node.Available && node.CurrentLoad != nil {
			load := node.CurrentLoad.OverallLoad
			loads = append(loads, load)
			totalLoad += load
		}
	}

	if len(loads) == 0 {
		return 0.0
	}

	avgLoad := totalLoad / float64(len(loads))

	// Calculate variance
	variance := 0.0
	for _, load := range loads {
		diff := load - avgLoad
		variance += diff * diff
	}
	variance /= float64(len(loads))

	// Return coefficient of variation as imbalance measure
	if avgLoad > 0 {
		return math.Sqrt(variance) / avgLoad
	}

	return 0.0
}

// ShouldRebalance determines if load rebalancing is needed
func (lb *LoadBalancer) ShouldRebalance() bool {
	imbalance := lb.CalculateLoadImbalance()
	return imbalance > lb.config.LoadImbalanceThreshold
}

// registerDefaultStrategies registers default load balancing strategies
func (lb *LoadBalancer) registerDefaultStrategies() {
	lb.strategies["round_robin"] = &RoundRobinLoadBalancer{}
	lb.strategies["least_loaded"] = &LeastLoadedBalancer{}
	lb.strategies["weighted_round_robin"] = &WeightedRoundRobinBalancer{}
	lb.strategies["resource_aware"] = &ResourceAwareBalancer{}
	lb.strategies["predictive"] = &PredictiveBalancer{predictor: lb.predictor}
}

// updatePredictions updates load predictions for a node
func (lb *LoadBalancer) updatePredictions(nodeID string, metrics *NodeLoadMetrics) {
	if lb.predictor == nil {
		return
	}

	lb.predictor.mu.Lock()
	defer lb.predictor.mu.Unlock()

	// Simple trend-based prediction
	prediction := &LoadPrediction{
		NodeID:      nodeID,
		TimeHorizon: lb.predictor.window,
		CreatedAt:   time.Now(),
		Confidence:  lb.predictor.accuracy,
		Factors:     make(map[string]float64),
	}

	// Predict based on current load and trend
	currentLoad := metrics.OverallLoad
	velocity := metrics.LoadVelocity

	// Simple linear prediction
	futureSeconds := lb.predictor.window.Seconds()
	prediction.PredictedLoad = currentLoad + (velocity * futureSeconds)

	// Clamp to valid range
	prediction.PredictedLoad = math.Max(0.0, math.Min(1.0, prediction.PredictedLoad))

	// Store prediction
	lb.predictor.predictions[nodeID] = prediction

	// Update node's predicted load
	if node, exists := lb.nodes[nodeID]; exists {
		node.PredictedLoad = prediction.PredictedLoad
	}
}

// monitoringLoop periodically monitors load and updates metrics
func (lb *LoadBalancer) monitoringLoop() {
	defer lb.wg.Done()

	ticker := time.NewTicker(lb.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case <-ticker.C:
			lb.updateMetrics()
			lb.createLoadSnapshot()
		}
	}
}

// updateMetrics updates load balancer metrics
func (lb *LoadBalancer) updateMetrics() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Update current imbalance
	lb.metrics.CurrentImbalance = lb.CalculateLoadImbalance()

	// Update max imbalance
	if lb.metrics.CurrentImbalance > lb.metrics.MaxImbalance {
		lb.metrics.MaxImbalance = lb.metrics.CurrentImbalance
	}

	// Calculate average resource utilization
	totalUtilization := 0.0
	activeNodes := 0

	for _, node := range lb.nodes {
		if node.Available && node.CurrentLoad != nil {
			totalUtilization += node.CurrentLoad.OverallLoad
			activeNodes++
		}
	}

	if activeNodes > 0 {
		lb.metrics.ResourceUtilization = totalUtilization / float64(activeNodes)
	}

	lb.metrics.LastUpdated = time.Now()
}

// createLoadSnapshot creates a snapshot of current system load
func (lb *LoadBalancer) createLoadSnapshot() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	snapshot := &LoadSnapshot{
		Timestamp:  time.Now(),
		TotalNodes: len(lb.nodes),
		NodeLoads:  make(map[string]float64),
	}

	totalLoad := 0.0
	activeNodes := 0
	loads := make([]float64, 0, len(lb.nodes))

	for nodeID, node := range lb.nodes {
		if node.Available && node.CurrentLoad != nil {
			load := node.CurrentLoad.OverallLoad
			snapshot.NodeLoads[nodeID] = load
			loads = append(loads, load)
			totalLoad += load
			activeNodes++
		}
	}

	snapshot.ActiveNodes = activeNodes

	if activeNodes > 0 {
		snapshot.AverageLoad = totalLoad / float64(activeNodes)

		// Calculate variance
		variance := 0.0
		for _, load := range loads {
			diff := load - snapshot.AverageLoad
			variance += diff * diff
		}
		snapshot.LoadVariance = variance / float64(activeNodes)

		// Calculate imbalance
		snapshot.LoadImbalance = lb.CalculateLoadImbalance()
	}

	// Add to history
	lb.loadHistory = append(lb.loadHistory, snapshot)

	// Limit history size
	if len(lb.loadHistory) > lb.config.MaxHistorySize {
		lb.loadHistory = lb.loadHistory[1:]
	}

	// Update predictor history
	if lb.predictor != nil {
		lb.predictor.mu.Lock()
		lb.predictor.loadHistory = append(lb.predictor.loadHistory, snapshot)
		if len(lb.predictor.loadHistory) > lb.config.MaxHistorySize {
			lb.predictor.loadHistory = lb.predictor.loadHistory[1:]
		}
		lb.predictor.mu.Unlock()
	}
}

// rebalancingLoop periodically checks if rebalancing is needed
func (lb *LoadBalancer) rebalancingLoop() {
	defer lb.wg.Done()

	ticker := time.NewTicker(lb.config.MaxRebalanceFrequency)
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case <-ticker.C:
			if lb.ShouldRebalance() {
				lb.performRebalancing()
			}
		}
	}
}

// performRebalancing performs load rebalancing
func (lb *LoadBalancer) performRebalancing() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	startTime := time.Now()

	// Get current strategy
	strategy := lb.getStrategy(lb.currentStrategy)
	if strategy == nil {
		return
	}

	// Check if strategy supports rebalancing
	availableNodes := lb.getAvailableNodes()
	if !strategy.ShouldRebalance(availableNodes) {
		return
	}

	// Perform rebalancing (simplified)
	// In a real implementation, you would:
	// 1. Identify overloaded and underloaded nodes
	// 2. Calculate optimal task redistribution
	// 3. Migrate tasks gracefully
	// 4. Monitor the rebalancing process

	lb.metrics.TotalRebalances++
	lb.metrics.SuccessfulRebalances++

	duration := time.Since(startTime)
	totalTime := time.Duration(lb.metrics.TotalRebalances-1)*lb.metrics.AverageRebalanceTime + duration
	lb.metrics.AverageRebalanceTime = totalTime / time.Duration(lb.metrics.TotalRebalances)
}

// predictionLoop periodically updates load predictions
func (lb *LoadBalancer) predictionLoop() {
	defer lb.wg.Done()

	if lb.predictor == nil || !lb.predictor.enabled {
		return
	}

	ticker := time.NewTicker(lb.config.PredictionWindow / 4) // Update 4 times per window
	defer ticker.Stop()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case <-ticker.C:
			lb.updateAllPredictions()
		}
	}
}

// updateAllPredictions updates predictions for all nodes
func (lb *LoadBalancer) updateAllPredictions() {
	lb.mu.RLock()
	nodeIDs := make([]string, 0, len(lb.nodes))
	for nodeID := range lb.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	lb.mu.RUnlock()

	for _, nodeID := range nodeIDs {
		if metrics, exists := lb.nodeMetrics[nodeID]; exists {
			lb.updatePredictions(nodeID, metrics)
		}
	}
}

// GetMetrics returns current load balancer metrics
func (lb *LoadBalancer) GetMetrics() *AdvancedLoadBalancerMetrics {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	metrics := *lb.metrics
	return &metrics
}

// GetLoadHistory returns load history snapshots
func (lb *LoadBalancer) GetLoadHistory(limit int) []*LoadSnapshot {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if limit <= 0 || limit > len(lb.loadHistory) {
		limit = len(lb.loadHistory)
	}

	start := len(lb.loadHistory) - limit
	history := make([]*LoadSnapshot, limit)
	copy(history, lb.loadHistory[start:])

	return history
}

// Close closes the load balancer
func (lb *LoadBalancer) Close() error {
	lb.cancel()
	lb.wg.Wait()
	return nil
}

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	lastIndex int
}

func (s *RoundRobinLoadBalancer) Name() string {
	return "round_robin"
}

func (s *RoundRobinLoadBalancer) SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	s.lastIndex = (s.lastIndex + 1) % len(nodes)
	return nodes[s.lastIndex], nil
}

func (s *RoundRobinLoadBalancer) CalculateWeight(node *LoadBalancedNode) float64 {
	return 1.0
}

func (s *RoundRobinLoadBalancer) ShouldRebalance(nodes []*LoadBalancedNode) bool {
	return false // Round-robin doesn't need rebalancing
}

// LeastLoadedBalancer selects the node with the lowest current load
type LeastLoadedBalancer struct{}

func (s *LeastLoadedBalancer) Name() string {
	return "least_loaded"
}

func (s *LeastLoadedBalancer) SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *LoadBalancedNode
	bestLoad := math.Inf(1)

	for _, node := range nodes {
		currentLoad := 0.0
		if node.CurrentLoad != nil {
			currentLoad = node.CurrentLoad.OverallLoad
		}

		if currentLoad < bestLoad {
			bestLoad = currentLoad
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *LeastLoadedBalancer) CalculateWeight(node *LoadBalancedNode) float64 {
	if node.CurrentLoad == nil {
		return 1.0
	}
	return 1.0 - node.CurrentLoad.OverallLoad
}

func (s *LeastLoadedBalancer) ShouldRebalance(nodes []*LoadBalancedNode) bool {
	if len(nodes) < 2 {
		return false
	}

	loads := make([]float64, 0, len(nodes))
	for _, node := range nodes {
		if node.CurrentLoad != nil {
			loads = append(loads, node.CurrentLoad.OverallLoad)
		}
	}

	if len(loads) < 2 {
		return false
	}

	sort.Float64s(loads)
	return (loads[len(loads)-1] - loads[0]) > 0.3 // 30% difference threshold
}

// WeightedRoundRobinBalancer implements weighted round-robin load balancing
type WeightedRoundRobinBalancer struct {
	currentWeights map[string]float64
}

func (s *WeightedRoundRobinBalancer) Name() string {
	return "weighted_round_robin"
}

func (s *WeightedRoundRobinBalancer) SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	if s.currentWeights == nil {
		s.currentWeights = make(map[string]float64)
	}

	var bestNode *LoadBalancedNode
	bestWeight := -1.0

	totalWeight := 0.0
	for _, node := range nodes {
		totalWeight += s.CalculateWeight(node)
	}

	for _, node := range nodes {
		nodeWeight := s.CalculateWeight(node)
		s.currentWeights[node.NodeID] += nodeWeight

		if s.currentWeights[node.NodeID] > bestWeight {
			bestWeight = s.currentWeights[node.NodeID]
			bestNode = node
		}
	}

	if bestNode != nil {
		s.currentWeights[bestNode.NodeID] -= totalWeight
	}

	return bestNode, nil
}

func (s *WeightedRoundRobinBalancer) CalculateWeight(node *LoadBalancedNode) float64 {
	weight := node.Weight

	// Adjust weight based on current load
	if node.CurrentLoad != nil {
		loadFactor := 1.0 - node.CurrentLoad.OverallLoad
		weight *= loadFactor
	}

	// Adjust weight based on reliability
	weight *= node.Reliability

	return math.Max(0.1, weight) // Minimum weight
}

func (s *WeightedRoundRobinBalancer) ShouldRebalance(nodes []*LoadBalancedNode) bool {
	return false // Weighted round-robin self-balances
}

// ResourceAwareBalancer selects nodes based on resource availability and requirements
type ResourceAwareBalancer struct{}

func (s *ResourceAwareBalancer) Name() string {
	return "resource_aware"
}

func (s *ResourceAwareBalancer) SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *LoadBalancedNode
	bestScore := -1.0

	for _, node := range nodes {
		score := s.CalculateWeight(node)

		// Consider task load impact
		if node.CurrentLoad != nil {
			projectedLoad := node.CurrentLoad.OverallLoad + taskLoad
			if projectedLoad > 1.0 {
				score *= 0.1 // Heavily penalize overload
			} else {
				score *= (1.0 - projectedLoad)
			}
		}

		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *ResourceAwareBalancer) CalculateWeight(node *LoadBalancedNode) float64 {
	if node.Capacity == nil {
		return 0.5
	}

	// Calculate available resource ratios
	cpuAvailable := node.Capacity.AvailableCPUCores / node.Capacity.TotalCPUCores
	memAvailable := float64(node.Capacity.AvailableMemoryBytes) / float64(node.Capacity.TotalMemoryBytes)
	diskAvailable := float64(node.Capacity.AvailableDiskBytes) / float64(node.Capacity.TotalDiskBytes)

	// Weighted average of available resources
	return (cpuAvailable*0.4 + memAvailable*0.4 + diskAvailable*0.2) * node.Reliability
}

func (s *ResourceAwareBalancer) ShouldRebalance(nodes []*LoadBalancedNode) bool {
	// Check for resource imbalance
	cpuUtils := make([]float64, 0, len(nodes))
	memUtils := make([]float64, 0, len(nodes))

	for _, node := range nodes {
		if node.CurrentLoad != nil {
			cpuUtils = append(cpuUtils, node.CurrentLoad.CPUUtilization)
			memUtils = append(memUtils, node.CurrentLoad.MemoryUtilization)
		}
	}

	if len(cpuUtils) < 2 {
		return false
	}

	sort.Float64s(cpuUtils)
	sort.Float64s(memUtils)

	cpuImbalance := cpuUtils[len(cpuUtils)-1] - cpuUtils[0]
	memImbalance := memUtils[len(memUtils)-1] - memUtils[0]

	return cpuImbalance > 0.4 || memImbalance > 0.4 // 40% imbalance threshold
}

// PredictiveBalancer uses load predictions for balancing decisions
type PredictiveBalancer struct {
	predictor *LoadPredictor
}

func (s *PredictiveBalancer) Name() string {
	return "predictive"
}

func (s *PredictiveBalancer) SelectNode(nodes []*LoadBalancedNode, taskLoad float64) (*LoadBalancedNode, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}

	var bestNode *LoadBalancedNode
	bestScore := -1.0

	for _, node := range nodes {
		score := s.CalculateWeight(node)

		// Consider predicted load
		projectedLoad := node.PredictedLoad + taskLoad
		if projectedLoad > 1.0 {
			score *= 0.1
		} else {
			score *= (1.0 - projectedLoad)
		}

		// Bonus for stable load trend
		if node.LoadTrend == LoadTrendStable {
			score *= 1.1
		} else if node.LoadTrend == LoadTrendDecreasing {
			score *= 1.05
		}

		if score > bestScore {
			bestScore = score
			bestNode = node
		}
	}

	return bestNode, nil
}

func (s *PredictiveBalancer) CalculateWeight(node *LoadBalancedNode) float64 {
	weight := 1.0

	// Base weight on current load
	if node.CurrentLoad != nil {
		weight *= (1.0 - node.CurrentLoad.OverallLoad)
	}

	// Adjust for predicted load
	weight *= (1.0 - node.PredictedLoad)

	// Adjust for reliability
	weight *= node.Reliability

	return math.Max(0.1, weight)
}

func (s *PredictiveBalancer) ShouldRebalance(nodes []*LoadBalancedNode) bool {
	// Check if any node is predicted to be overloaded
	for _, node := range nodes {
		if node.PredictedLoad > 0.9 {
			return true
		}
	}

	// Check for predicted load imbalance
	loads := make([]float64, 0, len(nodes))
	for _, node := range nodes {
		loads = append(loads, node.PredictedLoad)
	}

	if len(loads) < 2 {
		return false
	}

	sort.Float64s(loads)
	return (loads[len(loads)-1] - loads[0]) > 0.3
}
