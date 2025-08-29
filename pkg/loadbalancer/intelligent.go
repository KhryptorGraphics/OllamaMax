package loadbalancer

import (
	"context"
	"time"
	
	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// IntelligentLoadBalancer provides intelligent load balancing capabilities
type IntelligentLoadBalancer interface {
	// SelectNode selects the best node for a given request
	SelectNode(ctx context.Context, request *types.InferenceRequest, nodes []types.NodeInfo) (*types.NodeInfo, error)
	
	// UpdateNodeMetrics updates metrics for a node
	UpdateNodeMetrics(nodeID string, metrics *types.Metrics) error
	
	// GetNodeScore calculates the score for a node
	GetNodeScore(ctx context.Context, node types.NodeInfo, request *types.InferenceRequest) float64
	
	// GetStrategy returns the current load balancing strategy
	GetStrategy() string
	
	// SetStrategy sets the load balancing strategy
	SetStrategy(strategy string) error
}

// IntelligentNodeMetrics represents metrics for intelligent load balancing
type IntelligentNodeMetrics struct {
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	DiskUsage    float64   `json:"disk_usage"`
	NetworkIn    int64     `json:"network_in"`
	NetworkOut   int64     `json:"network_out"`
	ResponseTime time.Duration `json:"response_time"`
	RequestCount int64     `json:"request_count"`
	ErrorRate    float64   `json:"error_rate"`
	Timestamp    time.Time `json:"timestamp"`
}

// SmartLoadBalancer implements intelligent load balancing
type SmartLoadBalancer struct {
	strategy string
	metrics  map[string]IntelligentNodeMetrics
}

// NewSmartLoadBalancer creates a new smart load balancer
func NewSmartLoadBalancer(strategy string) *SmartLoadBalancer {
	return &SmartLoadBalancer{
		strategy: strategy,
		metrics:  make(map[string]IntelligentNodeMetrics),
	}
}

// SelectNode implements IntelligentLoadBalancer
func (lb *SmartLoadBalancer) SelectNode(ctx context.Context, request *types.InferenceRequest, nodes []types.NodeInfo) (*types.NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, ErrNoAvailableNodes
	}
	
	// Simple implementation - select first available node
	// In a real implementation, this would use the configured strategy
	for i := range nodes {
		if nodes[i].Status == "active" {
			return &nodes[i], nil
		}
	}
	
	return nil, ErrNoAvailableNodes
}

// UpdateNodeMetrics implements IntelligentLoadBalancer
func (lb *SmartLoadBalancer) UpdateNodeMetrics(nodeID string, metrics *types.Metrics) error {
	// Convert types.Metrics to IntelligentNodeMetrics
	intelligentMetrics := IntelligentNodeMetrics{
		CPUUsage:     metrics.CPUUsage,
		MemoryUsage:  metrics.MemoryUsage,
		DiskUsage:    metrics.DiskUsage,
		NetworkIn:    metrics.NetworkIn,
		NetworkOut:   metrics.NetworkOut,
		Timestamp:    metrics.Timestamp,
	}
	lb.metrics[nodeID] = intelligentMetrics
	return nil
}

// GetNodeScore implements IntelligentLoadBalancer
func (lb *SmartLoadBalancer) GetNodeScore(ctx context.Context, node types.NodeInfo, request *types.InferenceRequest) float64 {
	metrics, exists := lb.metrics[node.ID]
	if !exists {
		return 0.0
	}
	
	// Simple scoring - lower CPU and memory usage = higher score
	score := 100.0 - (metrics.CPUUsage + metrics.MemoryUsage) / 2
	return score
}

// GetStrategy implements IntelligentLoadBalancer
func (lb *SmartLoadBalancer) GetStrategy() string {
	return lb.strategy
}

// SetStrategy implements IntelligentLoadBalancer
func (lb *SmartLoadBalancer) SetStrategy(strategy string) error {
	lb.strategy = strategy
	return nil
}

// Common errors
var (
	ErrNoAvailableNodes = &LoadBalancerError{Code: "NO_AVAILABLE_NODES", Message: "No available nodes for load balancing"}
	ErrInvalidStrategy  = &LoadBalancerError{Code: "INVALID_STRATEGY", Message: "Invalid load balancing strategy"}
)

// LoadBalancerError represents a load balancer error
type LoadBalancerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *LoadBalancerError) Error() string {
	return e.Message
}