package loadbalancer

import (
	"time"
	"github.com/khryptorgraphics/ollamamax/pkg/types"
)

// LoadBalancer interface for load balancing operations
type LoadBalancer interface {
	SelectNode(nodes []*types.Node) (*types.Node, error)
	UpdateNodeMetrics(nodeID string, metrics *types.Metrics) error
	GetAlgorithm() string
}

// Algorithm represents load balancing algorithm
type Algorithm string

const (
	RoundRobin       Algorithm = "round_robin"
	WeightedRoundRobin Algorithm = "weighted_round_robin"
	LeastConnections Algorithm = "least_connections"
	LatencyBased     Algorithm = "latency_based"
)

// NodeMetrics represents node performance metrics
type NodeMetrics struct {
	NodeID         string        `json:"node_id"`
	Connections    int           `json:"connections"`
	ResponseTime   time.Duration `json:"response_time"`
	ErrorRate      float64       `json:"error_rate"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// MockLoadBalancer is a simple mock implementation
type MockLoadBalancer struct {
	algorithm Algorithm
	counter   int
}

func NewMockLoadBalancer(algo Algorithm) *MockLoadBalancer {
	return &MockLoadBalancer{
		algorithm: algo,
		counter:   0,
	}
}

func (m *MockLoadBalancer) SelectNode(nodes []*types.Node) (*types.Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	
	node := nodes[m.counter%len(nodes)]
	m.counter++
	return node, nil
}

func (m *MockLoadBalancer) UpdateNodeMetrics(nodeID string, metrics *types.Metrics) error {
	return nil
}

func (m *MockLoadBalancer) GetAlgorithm() string {
	return string(m.algorithm)
}