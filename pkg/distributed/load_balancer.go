package distributed

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// LoadBalancer interfaces for different load balancing strategies
type LoadBalancer interface {
	SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error)
	UpdateMetrics(nodeID string, metrics *NodeMetrics)
	GetMetrics() map[string]*NodeMetrics
	GetPrometheusRegistry() *prometheus.Registry
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	ID       string
	Model    string
	Prompt   string
	Options  map[string]interface{}
	Priority int
}

// NodeInfo contains information about a node
type NodeInfo struct {
	ID               string
	Address          string
	Status           string
	Models           []string
	Capacity         *NodeCapacity
	CPUCores         int
	TotalMemory      int64
	AvailableMemory  int64
	GPUCount         int
	NetworkBandwidth float64
	NetworkLatency   time.Duration
}

// NodeCapacity represents node resource capacity
type NodeCapacity struct {
	CPU    float64
	Memory int64
	GPU    int
}

// NodeMetrics tracks node performance metrics
type NodeMetrics struct {
	NodeID           string
	RequestCount     int64
	SuccessCount     int64
	ErrorCount       int64
	AverageLatency   time.Duration
	CurrentLoad      float64
	LastUpdated      time.Time
}

// RoundRobinBalancer implements round-robin load balancing
type RoundRobinBalancer struct {
	counter          int64
	metrics          map[string]*NodeMetrics
	mu               sync.RWMutex
	promRegistry     *prometheus.Registry
	requestsTotal    *prometheus.CounterVec
	selectionDuration prometheus.Histogram
	nodeUtilization  *prometheus.GaugeVec
}

// WeightedRoundRobinBalancer implements weighted round-robin
type WeightedRoundRobinBalancer struct {
	weights           map[string]int
	current           map[string]int
	metrics           map[string]*NodeMetrics
	mu                sync.RWMutex
	promRegistry      *prometheus.Registry
	requestsTotal     *prometheus.CounterVec
	selectionDuration prometheus.Histogram
	nodeUtilization   *prometheus.GaugeVec
}

// LeastConnectionsBalancer selects node with fewest active connections
type LeastConnectionsBalancer struct {
	connections       map[string]int64
	metrics           map[string]*NodeMetrics
	mu                sync.RWMutex
	promRegistry      *prometheus.Registry
	requestsTotal     *prometheus.CounterVec
	selectionDuration prometheus.Histogram
	nodeUtilization   *prometheus.GaugeVec
}

// LatencyBasedBalancer selects node with lowest latency
type LatencyBasedBalancer struct {
	metrics           map[string]*NodeMetrics
	mu                sync.RWMutex
	promRegistry      *prometheus.Registry
	requestsTotal     *prometheus.CounterVec
	selectionDuration prometheus.Histogram
	nodeUtilization   *prometheus.GaugeVec
}

// NewRoundRobinBalancer creates a new round-robin balancer
func NewRoundRobinBalancer() *RoundRobinBalancer {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of load balancer requests",
		},
		[]string{"strategy", "node_id"},
	)

	selectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "lb_node_selection_duration_seconds",
			Help:    "Time taken to select a node",
			Buckets: prometheus.DefBuckets,
		},
	)

	nodeUtilization := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_node_utilization",
			Help: "Current utilization of each node",
		},
		[]string{"node_id"},
	)

	registry.MustRegister(requestsTotal)
	registry.MustRegister(selectionDuration)
	registry.MustRegister(nodeUtilization)

	return &RoundRobinBalancer{
		metrics:           make(map[string]*NodeMetrics),
		promRegistry:      registry,
		requestsTotal:     requestsTotal,
		selectionDuration: selectionDuration,
		nodeUtilization:   nodeUtilization,
	}
}

// SelectNode selects next node using round-robin
func (rb *RoundRobinBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	timer := prometheus.NewTimer(rb.selectionDuration)
	defer timer.ObserveDuration()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	rb.mu.Lock()
	rb.counter++
	index := int(rb.counter) % len(nodes)
	rb.mu.Unlock()

	selectedNode := &nodes[index]

	// Record request metrics
	rb.requestsTotal.WithLabelValues("round-robin", selectedNode.ID).Inc()

	return selectedNode, nil
}

// UpdateMetrics updates node metrics
func (rb *RoundRobinBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.metrics[nodeID] = metrics

	// Update Prometheus gauge for node utilization
	rb.nodeUtilization.WithLabelValues(nodeID).Set(metrics.CurrentLoad)
}

// GetMetrics returns all metrics
func (rb *RoundRobinBalancer) GetMetrics() map[string]*NodeMetrics {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range rb.metrics {
		result[k] = v
	}
	return result
}

// GetPrometheusRegistry returns the Prometheus registry
func (rb *RoundRobinBalancer) GetPrometheusRegistry() *prometheus.Registry {
	return rb.promRegistry
}

// NewWeightedRoundRobinBalancer creates a weighted round-robin balancer
func NewWeightedRoundRobinBalancer(weights map[string]int) *WeightedRoundRobinBalancer {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of load balancer requests",
		},
		[]string{"strategy", "node_id"},
	)

	selectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "lb_node_selection_duration_seconds",
			Help:    "Time taken to select a node",
			Buckets: prometheus.DefBuckets,
		},
	)

	nodeUtilization := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_node_utilization",
			Help: "Current utilization of each node",
		},
		[]string{"node_id"},
	)

	registry.MustRegister(requestsTotal)
	registry.MustRegister(selectionDuration)
	registry.MustRegister(nodeUtilization)

	return &WeightedRoundRobinBalancer{
		weights:           weights,
		current:           make(map[string]int),
		metrics:           make(map[string]*NodeMetrics),
		promRegistry:      registry,
		requestsTotal:     requestsTotal,
		selectionDuration: selectionDuration,
		nodeUtilization:   nodeUtilization,
	}
}

// SelectNode selects node using weighted round-robin
func (wrb *WeightedRoundRobinBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	timer := prometheus.NewTimer(wrb.selectionDuration)
	defer timer.ObserveDuration()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	// Find node with highest current weight
	var selectedNode *NodeInfo
	maxWeight := -1

	for i, node := range nodes {
		weight := wrb.weights[node.ID]
		if weight == 0 {
			weight = 1 // Default weight
		}

		wrb.current[node.ID] += weight

		if wrb.current[node.ID] > maxWeight {
			maxWeight = wrb.current[node.ID]
			selectedNode = &nodes[i]
		}
	}

	if selectedNode != nil {
		// Decrease selected node's current weight
		totalWeight := 0
		for _, weight := range wrb.weights {
			totalWeight += weight
		}
		wrb.current[selectedNode.ID] -= totalWeight

		// Record request metrics
		wrb.requestsTotal.WithLabelValues("weighted-round-robin", selectedNode.ID).Inc()
	}

	return selectedNode, nil
}

// UpdateMetrics updates node metrics for weighted balancer
func (wrb *WeightedRoundRobinBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	wrb.mu.Lock()
	defer wrb.mu.Unlock()

	wrb.metrics[nodeID] = metrics

	// Update Prometheus gauge for node utilization
	wrb.nodeUtilization.WithLabelValues(nodeID).Set(metrics.CurrentLoad)
}

// GetMetrics returns all metrics for weighted balancer
func (wrb *WeightedRoundRobinBalancer) GetMetrics() map[string]*NodeMetrics {
	wrb.mu.RLock()
	defer wrb.mu.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range wrb.metrics {
		result[k] = v
	}
	return result
}

// GetPrometheusRegistry returns the Prometheus registry
func (wrb *WeightedRoundRobinBalancer) GetPrometheusRegistry() *prometheus.Registry {
	return wrb.promRegistry
}

// NewLeastConnectionsBalancer creates a least connections balancer
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of load balancer requests",
		},
		[]string{"strategy", "node_id"},
	)

	selectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "lb_node_selection_duration_seconds",
			Help:    "Time taken to select a node",
			Buckets: prometheus.DefBuckets,
		},
	)

	nodeUtilization := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_node_utilization",
			Help: "Current utilization of each node",
		},
		[]string{"node_id"},
	)

	registry.MustRegister(requestsTotal)
	registry.MustRegister(selectionDuration)
	registry.MustRegister(nodeUtilization)

	return &LeastConnectionsBalancer{
		connections:       make(map[string]int64),
		metrics:           make(map[string]*NodeMetrics),
		promRegistry:      registry,
		requestsTotal:     requestsTotal,
		selectionDuration: selectionDuration,
		nodeUtilization:   nodeUtilization,
	}
}

// SelectNode selects node with fewest connections (based on request count from metrics)
func (lcb *LeastConnectionsBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	timer := prometheus.NewTimer(lcb.selectionDuration)
	defer timer.ObserveDuration()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	lcb.mu.RLock()
	defer lcb.mu.RUnlock()

	var selectedNode *NodeInfo
	minConnections := int64(math.MaxInt64)

	for i, node := range nodes {
		// Use RequestCount from metrics if available, otherwise use connections map
		var connections int64
		if metrics, exists := lcb.metrics[node.ID]; exists {
			connections = metrics.RequestCount
		} else {
			connections = lcb.connections[node.ID]
		}

		if connections < minConnections {
			minConnections = connections
			selectedNode = &nodes[i]
		}
	}

	if selectedNode != nil {
		// Record request metrics
		lcb.requestsTotal.WithLabelValues("least-connections", selectedNode.ID).Inc()
	}

	return selectedNode, nil
}

// UpdateMetrics updates metrics and connection count
func (lcb *LeastConnectionsBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	lcb.mu.Lock()
	defer lcb.mu.Unlock()

	lcb.metrics[nodeID] = metrics
	lcb.connections[nodeID] = metrics.RequestCount - metrics.SuccessCount - metrics.ErrorCount

	// Update Prometheus gauge for node utilization
	lcb.nodeUtilization.WithLabelValues(nodeID).Set(metrics.CurrentLoad)
}

// GetMetrics returns all metrics for least connections balancer
func (lcb *LeastConnectionsBalancer) GetMetrics() map[string]*NodeMetrics {
	lcb.mu.RLock()
	defer lcb.mu.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range lcb.metrics {
		result[k] = v
	}
	return result
}

// GetPrometheusRegistry returns the Prometheus registry
func (lcb *LeastConnectionsBalancer) GetPrometheusRegistry() *prometheus.Registry {
	return lcb.promRegistry
}

// NewLatencyBasedBalancer creates a latency-based balancer
func NewLatencyBasedBalancer() *LatencyBasedBalancer {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of load balancer requests",
		},
		[]string{"strategy", "node_id"},
	)

	selectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "lb_node_selection_duration_seconds",
			Help:    "Time taken to select a node",
			Buckets: prometheus.DefBuckets,
		},
	)

	nodeUtilization := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_node_utilization",
			Help: "Current utilization of each node",
		},
		[]string{"node_id"},
	)

	registry.MustRegister(requestsTotal)
	registry.MustRegister(selectionDuration)
	registry.MustRegister(nodeUtilization)

	return &LatencyBasedBalancer{
		metrics:           make(map[string]*NodeMetrics),
		promRegistry:      registry,
		requestsTotal:     requestsTotal,
		selectionDuration: selectionDuration,
		nodeUtilization:   nodeUtilization,
	}
}

// SelectNode selects node with lowest latency
func (lbb *LatencyBasedBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	timer := prometheus.NewTimer(lbb.selectionDuration)
	defer timer.ObserveDuration()

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	lbb.mu.RLock()
	defer lbb.mu.RUnlock()

	// Sort nodes by latency
	nodeLatencies := make([]struct {
		node    *NodeInfo
		latency time.Duration
	}, 0, len(nodes))

	for i, node := range nodes {
		metrics := lbb.metrics[node.ID]
		latency := time.Hour // Default high latency for unknown nodes
		if metrics != nil {
			latency = metrics.AverageLatency
		}

		nodeLatencies = append(nodeLatencies, struct {
			node    *NodeInfo
			latency time.Duration
		}{
			node:    &nodes[i],
			latency: latency,
		})
	}

	sort.Slice(nodeLatencies, func(i, j int) bool {
		return nodeLatencies[i].latency < nodeLatencies[j].latency
	})

	selectedNode := nodeLatencies[0].node

	// Record request metrics
	lbb.requestsTotal.WithLabelValues("latency-based", selectedNode.ID).Inc()

	return selectedNode, nil
}

// UpdateMetrics updates node metrics for latency balancer
func (lbb *LatencyBasedBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	lbb.mu.Lock()
	defer lbb.mu.Unlock()

	lbb.metrics[nodeID] = metrics

	// Update Prometheus gauge for node utilization
	lbb.nodeUtilization.WithLabelValues(nodeID).Set(metrics.CurrentLoad)
}

// GetMetrics returns all metrics for latency balancer
func (lbb *LatencyBasedBalancer) GetMetrics() map[string]*NodeMetrics {
	lbb.mu.RLock()
	defer lbb.mu.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range lbb.metrics {
		result[k] = v
	}
	return result
}

// GetPrometheusRegistry returns the Prometheus registry
func (lbb *LatencyBasedBalancer) GetPrometheusRegistry() *prometheus.Registry {
	return lbb.promRegistry
}

// SmartLoadBalancer combines multiple strategies
type SmartLoadBalancer struct {
	strategies         map[string]LoadBalancer
	selector           StrategySelector
	metrics            map[string]*NodeMetrics
	mu                 sync.RWMutex
	currentStrategy    string
	promRegistry       *prometheus.Registry
	requestsTotal      *prometheus.CounterVec
	selectionDuration  prometheus.Histogram
	nodeUtilization    *prometheus.GaugeVec
	strategySwitches   prometheus.Counter
}

// StrategySelector selects appropriate balancing strategy
type StrategySelector interface {
	SelectStrategy(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) string
}

// DefaultStrategySelector implements basic strategy selection
type DefaultStrategySelector struct{}

// SelectStrategy chooses strategy based on request characteristics
func (dss *DefaultStrategySelector) SelectStrategy(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) string {
	// High priority requests use latency-based
	if request.Priority > 5 {
		return "latency"
	}

	// Many nodes use least connections
	if len(nodes) > 10 {
		return "least-connections"
	}

	// Default to round-robin
	return "round-robin"
}

// NewSmartLoadBalancer creates a smart load balancer
func NewSmartLoadBalancer() *SmartLoadBalancer {
	registry := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "lb_requests_total",
			Help: "Total number of load balancer requests",
		},
		[]string{"strategy", "node_id"},
	)

	selectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "lb_node_selection_duration_seconds",
			Help:    "Time taken to select a node",
			Buckets: prometheus.DefBuckets,
		},
	)

	nodeUtilization := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lb_node_utilization",
			Help: "Current utilization of each node",
		},
		[]string{"node_id"},
	)

	strategySwitches := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "lb_strategy_switches_total",
			Help: "Total number of load balancer strategy switches",
		},
	)

	registry.MustRegister(requestsTotal)
	registry.MustRegister(selectionDuration)
	registry.MustRegister(nodeUtilization)
	registry.MustRegister(strategySwitches)

	strategies := make(map[string]LoadBalancer)
	strategies["round-robin"] = NewRoundRobinBalancer()
	strategies["least-connections"] = NewLeastConnectionsBalancer()
	strategies["latency"] = NewLatencyBasedBalancer()

	return &SmartLoadBalancer{
		strategies:        strategies,
		selector:          &DefaultStrategySelector{},
		metrics:           make(map[string]*NodeMetrics),
		currentStrategy:   "round-robin",
		promRegistry:      registry,
		requestsTotal:     requestsTotal,
		selectionDuration: selectionDuration,
		nodeUtilization:   nodeUtilization,
		strategySwitches:  strategySwitches,
	}
}

// SelectNode selects node using appropriate strategy
func (slb *SmartLoadBalancer) SelectNode(ctx context.Context, request *InferenceRequest, nodes []NodeInfo) (*NodeInfo, error) {
	timer := prometheus.NewTimer(slb.selectionDuration)
	defer timer.ObserveDuration()

	strategyName := slb.selector.SelectStrategy(ctx, request, nodes)

	// Check if strategy changed
	slb.mu.Lock()
	if strategyName != slb.currentStrategy {
		slb.currentStrategy = strategyName
		slb.strategySwitches.Inc()
	}
	strategy := slb.strategies[strategyName]
	slb.mu.Unlock()

	if strategy == nil {
		strategy = slb.strategies["round-robin"] // Fallback
		strategyName = "round-robin"
	}

	selectedNode, err := strategy.SelectNode(ctx, request, nodes)
	if err != nil {
		return nil, err
	}

	// Record request metrics at smart balancer level
	slb.requestsTotal.WithLabelValues("smart-"+strategyName, selectedNode.ID).Inc()

	return selectedNode, nil
}

// UpdateMetrics updates metrics across all strategies
func (slb *SmartLoadBalancer) UpdateMetrics(nodeID string, metrics *NodeMetrics) {
	slb.mu.Lock()
	defer slb.mu.Unlock()

	slb.metrics[nodeID] = metrics

	// Update Prometheus gauge for node utilization
	slb.nodeUtilization.WithLabelValues(nodeID).Set(metrics.CurrentLoad)

	// Update all strategies
	for _, strategy := range slb.strategies {
		strategy.UpdateMetrics(nodeID, metrics)
	}
}

// GetMetrics returns consolidated metrics
func (slb *SmartLoadBalancer) GetMetrics() map[string]*NodeMetrics {
	slb.mu.RLock()
	defer slb.mu.RUnlock()

	result := make(map[string]*NodeMetrics)
	for k, v := range slb.metrics {
		result[k] = v
	}
	return result
}

// GetPrometheusRegistry returns the Prometheus registry
func (slb *SmartLoadBalancer) GetPrometheusRegistry() *prometheus.Registry {
	return slb.promRegistry
}
