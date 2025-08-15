package cluster

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/sirupsen/logrus"
)

// Discovery Strategies

// MDNSDiscovery implements mDNS-based node discovery
type MDNSDiscovery struct {
	config *config.DistributedConfig
	logger *logrus.Logger
	mu     sync.RWMutex
}

func (m *MDNSDiscovery) Discover() ([]*NodeInfo, error) {
	// Simulate mDNS discovery
	nodes := []*NodeInfo{
		{
			ID:       "mdns-node-1",
			Name:     "MDNS Discovered Node 1",
			Address:  "192.168.1.100:8080",
			Region:   "local",
			Zone:     "local-a",
			Status:   NodeStatusHealthy,
			LastSeen: time.Now(),
			Capabilities: NodeCapabilities{
				Inference: true,
				Storage:   false,
				Models:    []string{"llama2-7b"},
			},
		},
	}
	return nodes, nil
}

func (m *MDNSDiscovery) GetName() string {
	return "mdns"
}

// P2PDiscovery implements P2P network-based node discovery
type P2PDiscovery struct {
	config *config.DistributedConfig
	logger *logrus.Logger
	mu     sync.RWMutex
}

func (p *P2PDiscovery) Discover() ([]*NodeInfo, error) {
	// Simulate P2P discovery
	nodes := []*NodeInfo{
		{
			ID:       "p2p-node-1",
			Name:     "P2P Discovered Node 1",
			Address:  "10.0.1.50:8080",
			Region:   "us-west-2",
			Zone:     "us-west-2b",
			Status:   NodeStatusHealthy,
			LastSeen: time.Now(),
			Capabilities: NodeCapabilities{
				Inference:    true,
				Storage:      true,
				Coordination: false,
				Models:       []string{"llama2-13b", "codellama-7b"},
			},
		},
	}
	return nodes, nil
}

func (p *P2PDiscovery) GetName() string {
	return "p2p"
}

// Load Balancing Strategies

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	counter uint64
	mu      sync.Mutex
}

func (r *RoundRobinStrategy) SelectNode(nodes []*NodeInfo, request *RequestContext) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	// Filter healthy nodes
	healthyNodes := make([]*NodeInfo, 0)
	for _, node := range nodes {
		if node.Status == NodeStatusHealthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	if len(healthyNodes) == 0 {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	r.mu.Lock()
	index := r.counter % uint64(len(healthyNodes))
	r.counter++
	r.mu.Unlock()

	return healthyNodes[index], nil
}

func (r *RoundRobinStrategy) GetName() string {
	return "round_robin"
}

// LeastLoadedStrategy selects the node with the lowest load
type LeastLoadedStrategy struct {
	loadMetrics map[string]*LoadMetrics
	mu          sync.RWMutex
}

func (l *LeastLoadedStrategy) SelectNode(nodes []*NodeInfo, request *RequestContext) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	var bestNode *NodeInfo
	var lowestLoad float64 = math.MaxFloat64

	for _, node := range nodes {
		if node.Status != NodeStatusHealthy {
			continue
		}

		// Calculate load score (CPU + Memory utilization)
		cpuLoad := node.Resources.CPU.Percent
		memoryLoad := node.Resources.Memory.Percent
		totalLoad := (cpuLoad + memoryLoad) / 2.0

		if totalLoad < lowestLoad {
			lowestLoad = totalLoad
			bestNode = node
		}
	}

	if bestNode == nil {
		return nil, fmt.Errorf("no healthy nodes available")
	}

	return bestNode, nil
}

func (l *LeastLoadedStrategy) GetName() string {
	return "least_loaded"
}

// AffinityStrategy implements model affinity-based load balancing
type AffinityStrategy struct {
	affinityMap map[string]string // model -> preferred node
	mu          sync.RWMutex
}

func (a *AffinityStrategy) SelectNode(nodes []*NodeInfo, request *RequestContext) (*NodeInfo, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	a.mu.RLock()
	preferredNodeID, hasAffinity := a.affinityMap[request.ModelName]
	a.mu.RUnlock()

	// If we have affinity, try to use the preferred node
	if hasAffinity {
		for _, node := range nodes {
			if node.ID == preferredNodeID && node.Status == NodeStatusHealthy {
				// Check if node has the model
				for _, model := range node.Capabilities.Models {
					if model == request.ModelName {
						return node, nil
					}
				}
			}
		}
	}

	// Fallback: find any node with the required model
	for _, node := range nodes {
		if node.Status != NodeStatusHealthy {
			continue
		}
		for _, model := range node.Capabilities.Models {
			if model == request.ModelName {
				return node, nil
			}
		}
	}

	// Final fallback: use round-robin
	fallback := &RoundRobinStrategy{}
	return fallback.SelectNode(nodes, request)
}

func (a *AffinityStrategy) GetName() string {
	return "affinity"
}

// Alert Channels

// LogAlertChannel sends alerts to logs
type LogAlertChannel struct {
	logger *logrus.Logger
}

func (l *LogAlertChannel) SendAlert(alert *Alert) error {
	switch alert.Severity {
	case AlertSeverityCritical:
		l.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"type":     alert.Type,
			"node_id":  alert.NodeID,
		}).Error(alert.Message)
	case AlertSeverityError:
		l.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"type":     alert.Type,
			"node_id":  alert.NodeID,
		}).Error(alert.Message)
	case AlertSeverityWarning:
		l.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"type":     alert.Type,
			"node_id":  alert.NodeID,
		}).Warn(alert.Message)
	default:
		l.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"type":     alert.Type,
			"node_id":  alert.NodeID,
		}).Info(alert.Message)
	}
	return nil
}

func (l *LogAlertChannel) GetName() string {
	return "log"
}

// WebhookAlertChannel sends alerts via HTTP webhooks
type WebhookAlertChannel struct {
	url    string
	logger *logrus.Logger
}

func (w *WebhookAlertChannel) SendAlert(alert *Alert) error {
	// Simulate webhook sending
	w.logger.Infof("Sending alert %s to webhook %s", alert.ID, w.url)
	return nil
}

func (w *WebhookAlertChannel) GetName() string {
	return "webhook"
}

// Helper functions for component implementations

// Start methods for components
func (nd *NodeDiscovery) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nd.discoverNodes()
		}
	}
}

func (nd *NodeDiscovery) discoverNodes() {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	for _, strategy := range nd.strategies {
		nodes, err := strategy.Discover()
		if err != nil {
			nd.logger.Errorf("Discovery strategy %s failed: %v", strategy.GetName(), err)
			continue
		}

		for _, node := range nodes {
			nd.nodes[node.ID] = node
			nd.logger.Debugf("Discovered node %s via %s", node.ID, strategy.GetName())
		}
	}
}

func (hm *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.performHealthChecks()
		}
	}
}

func (hm *HealthMonitor) performHealthChecks() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	for nodeID, check := range hm.healthChecks {
		if !check.Enabled {
			continue
		}

		start := time.Now()
		success := hm.performSingleHealthCheck(check)
		latency := time.Since(start)

		check.LastResult = &HealthResult{
			Success:   success,
			Latency:   latency,
			Timestamp: time.Now(),
		}

		// Update health score
		if success {
			hm.healthScores[nodeID] = 1.0
		} else {
			hm.healthScores[nodeID] = 0.0
		}

		hm.lastChecked[nodeID] = time.Now()
	}
}

func (hm *HealthMonitor) performSingleHealthCheck(check *HealthCheck) bool {
	// Simulate health check - in real implementation, this would make HTTP requests
	// For now, randomly succeed/fail to simulate real behavior
	return rand.Float64() > 0.1 // 90% success rate
}

func (hm *HealthMonitor) GetHealthScore(nodeID string) float64 {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if score, exists := hm.healthScores[nodeID]; exists {
		return score
	}
	return 0.5 // Default neutral score
}

func (hm *HealthMonitor) GetAllHealthScores() map[string]float64 {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result := make(map[string]float64)
	for nodeID, score := range hm.healthScores {
		result[nodeID] = score
	}
	return result
}
