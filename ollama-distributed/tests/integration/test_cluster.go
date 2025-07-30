package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/distributed"
)

// TestCluster represents a test cluster for integration testing
type TestCluster struct {
	nodes      []*TestNode
	dataDir    string
	mu         sync.RWMutex
	started    bool
	shutdownCh chan struct{}
}

// TestNode represents a test node in the cluster
type TestNode struct {
	config    *config.Config
	p2pNode   *p2p.Node
	consensus *consensus.Engine
	scheduler *distributed.DistributedScheduler
	apiServer *api.Server
	active    bool
	mu        sync.RWMutex
}

// NewTestCluster creates a new test cluster with the specified number of nodes
func NewTestCluster(nodeCount int) (*TestCluster, error) {
	if nodeCount < 1 {
		return nil, fmt.Errorf("node count must be at least 1")
	}

	// Create temporary directory for test data
	dataDir := filepath.Join(os.TempDir(), fmt.Sprintf("ollama-test-cluster-%d", time.Now().UnixNano()))
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create test data directory: %w", err)
	}

	cluster := &TestCluster{
		nodes:      make([]*TestNode, nodeCount),
		dataDir:    dataDir,
		shutdownCh: make(chan struct{}),
	}

	// Create nodes
	for i := 0; i < nodeCount; i++ {
		node, err := cluster.createNode(i)
		if err != nil {
			cluster.cleanup()
			return nil, fmt.Errorf("failed to create node %d: %w", i, err)
		}
		cluster.nodes[i] = node
	}

	return cluster, nil
}

// createNode creates a test node with the given index
func (tc *TestCluster) createNode(index int) (*TestNode, error) {
	// Create node-specific directory
	nodeDir := filepath.Join(tc.dataDir, fmt.Sprintf("node-%d", index))
	err := os.MkdirAll(nodeDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create node directory: %w", err)
	}

	// Create configuration
	cfg := tc.createNodeConfig(nodeDir, index)

	// Create P2P node
	p2pNode, err := p2p.NewNode(context.Background(), &cfg.P2P)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Create consensus engine
	consensusEngine, err := consensus.NewEngine(&cfg.Consensus, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Create distributed scheduler
	schedulerConfig := &distributed.DistributedConfig{
		ClusterID:         cfg.ClusterID,
		NodeID:            cfg.NodeID,
		MaxNodes:          len(tc.nodes),
		HeartbeatInterval: 10 * time.Second,
		DefaultStrategy:   "layerwise",
		LayerThreshold:    8,
		BatchSizeLimit:    32,
		LBAlgorithm:       "weighted_round_robin",
		LatencyTarget:     100 * time.Millisecond,
		WeightFactors: map[string]float64{
			"cpu":    0.3,
			"memory": 0.3,
			"gpu":    0.4,
		},
		ReplicationFactor:     2,
		HealthCheckInterval:   5 * time.Second,
		RecoveryTimeout:       30 * time.Second,
		CommunicationProtocol: "libp2p",
		Encryption:            false, // Disable for testing
		Compression:           true,
	}

	scheduler, err := distributed.NewDistributedScheduler(nil, schedulerConfig, p2pNode, consensusEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create distributed scheduler: %w", err)
	}

	// Create API server
	apiServer, err := api.NewServer(&cfg.API, scheduler)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	return &TestNode{
		config:    cfg,
		p2pNode:   p2pNode,
		consensus: consensusEngine,
		scheduler: scheduler,
		apiServer: apiServer,
		active:    false,
	}, nil
}

// createNodeConfig creates a configuration for a test node
func (tc *TestCluster) createNodeConfig(nodeDir string, index int) *config.Config {
	cfg := config.DefaultConfig()

	// Base ports for each node
	basePort := 20000 + index*100

	// Node identification
	cfg.NodeID = fmt.Sprintf("test-node-%d", index)
	cfg.ClusterID = "test-cluster"

	// Network configuration
	cfg.API.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+1)
	cfg.P2P.Listen = fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", basePort+2)
	cfg.Consensus.BindAddr = fmt.Sprintf("127.0.0.1:%d", basePort+3)
	cfg.Web.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+4)
	cfg.Metrics.Listen = fmt.Sprintf("127.0.0.1:%d", basePort+5)

	// Storage configuration
	cfg.Storage.DataDir = filepath.Join(nodeDir, "data")
	cfg.Storage.ModelDir = filepath.Join(nodeDir, "models")
	cfg.Storage.CacheDir = filepath.Join(nodeDir, "cache")
	cfg.Consensus.DataDir = filepath.Join(nodeDir, "consensus")

	// Security (disable for testing)
	cfg.Security.TLS.Enabled = false
	cfg.Security.Auth.Enabled = false
	cfg.Web.TLS.Enabled = false

	// Cluster configuration
	cfg.Consensus.Bootstrap = (index == 0) // First node is bootstrap
	cfg.Scheduler.WorkerCount = 2
	cfg.P2P.ConnMgrLow = 1
	cfg.P2P.ConnMgrHigh = 10

	// Bootstrap peers (first node for others)
	if index > 0 {
		bootstrapPort := 20000 + 2 // First node's P2P port
		cfg.P2P.Bootstrap = []string{
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", bootstrapPort),
		}
	}

	return cfg
}

// Start starts all nodes in the cluster
func (tc *TestCluster) Start() error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.started {
		return fmt.Errorf("cluster already started")
	}

	// Start nodes sequentially with delays
	for i, node := range tc.nodes {
		err := node.Start()
		if err != nil {
			return fmt.Errorf("failed to start node %d: %w", i, err)
		}

		// Wait between node starts
		time.Sleep(2 * time.Second)
	}

	// Wait for cluster to stabilize
	time.Sleep(10 * time.Second)

	tc.started = true
	return nil
}

// Shutdown shuts down all nodes and cleans up
func (tc *TestCluster) Shutdown() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.started {
		return
	}

	close(tc.shutdownCh)

	// Shutdown nodes
	for _, node := range tc.nodes {
		if node != nil {
			node.Shutdown()
		}
	}

	// Cleanup data directory
	tc.cleanup()
	tc.started = false
}

// cleanup removes the test data directory
func (tc *TestCluster) cleanup() {
	if tc.dataDir != "" {
		os.RemoveAll(tc.dataDir)
	}
}

// GetNodes returns all nodes in the cluster
func (tc *TestCluster) GetNodes() []*TestNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	nodes := make([]*TestNode, len(tc.nodes))
	copy(nodes, tc.nodes)
	return nodes
}

// GetActiveNodes returns only active nodes
func (tc *TestCluster) GetActiveNodes() []*TestNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	var activeNodes []*TestNode
	for _, node := range tc.nodes {
		if node.IsActive() {
			activeNodes = append(activeNodes, node)
		}
	}
	return activeNodes
}

// GetLeader returns the cluster leader
func (tc *TestCluster) GetLeader() *TestNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	for _, node := range tc.nodes {
		if node.IsActive() && node.IsLeader() {
			return node
		}
	}
	return nil
}

// GetNodeLoads returns the current load on each node
func (tc *TestCluster) GetNodeLoads() map[string]int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	loads := make(map[string]int)
	for _, node := range tc.nodes {
		if node.IsActive() {
			loads[node.GetID()] = node.GetLoad()
		}
	}
	return loads
}

// TestNode methods

// Start starts the test node
func (tn *TestNode) Start() error {
	tn.mu.Lock()
	defer tn.mu.Unlock()

	if tn.active {
		return fmt.Errorf("node already started")
	}

	// Start P2P node
	err := tn.p2pNode.Start()
	if err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Start consensus engine
	err = tn.consensus.Start()
	if err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start scheduler
	err = tn.scheduler.Start()
	if err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start API server
	err = tn.apiServer.Start()
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	tn.active = true
	return nil
}

// Shutdown shuts down the test node
func (tn *TestNode) Shutdown() error {
	tn.mu.Lock()
	defer tn.mu.Unlock()

	if !tn.active {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown components in reverse order
	if tn.apiServer != nil {
		tn.apiServer.Shutdown(ctx)
	}

	if tn.scheduler != nil {
		tn.scheduler.Shutdown(ctx)
	}

	if tn.consensus != nil {
		tn.consensus.Shutdown(ctx)
	}

	if tn.p2pNode != nil {
		tn.p2pNode.Shutdown(ctx)
	}

	tn.active = false
	return nil
}

// IsActive returns whether the node is active
func (tn *TestNode) IsActive() bool {
	tn.mu.RLock()
	defer tn.mu.RUnlock()
	return tn.active
}

// IsLeader returns whether the node is the cluster leader
func (tn *TestNode) IsLeader() bool {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.consensus == nil {
		return false
	}

	return tn.consensus.IsLeader()
}

// GetID returns the node ID
func (tn *TestNode) GetID() string {
	return tn.config.NodeID
}

// GetLoad returns the current load on the node
func (tn *TestNode) GetLoad() int {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return 0
	}

	tasks := tn.scheduler.GetActiveTasks()
	return len(tasks)
}

// ProcessInference processes an inference request
func (tn *TestNode) ProcessInference(ctx context.Context, req *api.InferenceRequest) (*api.InferenceResponse, error) {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return nil, fmt.Errorf("node not active")
	}

	// For testing, return a mock response
	return &api.InferenceResponse{
		Response:      fmt.Sprintf("Mock response from %s to: %s", tn.GetID(), req.Prompt),
		Done:          true,
		TotalDuration: 1000000000, // 1 second in nanoseconds
	}, nil
}

// RegisterModel registers a model on the node
func (tn *TestNode) RegisterModel(model *distributed.ModelInfo) error {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return fmt.Errorf("node not active")
	}

	// For testing, just record the model
	return nil
}

// UpdateModel updates a model on the node
func (tn *TestNode) UpdateModel(model *distributed.ModelInfo) error {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return fmt.Errorf("node not active")
	}

	// For testing, just update the model
	return nil
}

// HasModel returns whether the node has the specified model
func (tn *TestNode) HasModel(modelName string) bool {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return false
	}

	// For testing, simulate model presence
	return true
}

// GetModel returns the model information
func (tn *TestNode) GetModel(modelName string) *distributed.ModelInfo {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return nil
	}

	// For testing, return a mock model
	return &distributed.ModelInfo{
		Name:         modelName,
		Version:      "1.0.0",
		Size:         1024 * 1024 * 1024,
		Checksum:     "mock-checksum",
		LastAccessed: time.Now(),
		Popularity:   0.5,
	}
}

// GetModels returns all models on the node
func (tn *TestNode) GetModels() []*distributed.ModelInfo {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.scheduler == nil {
		return nil
	}

	// For testing, return mock models
	return []*distributed.ModelInfo{
		{
			Name:         "llama3.2:1b",
			Version:      "1.0.0",
			Size:         1024 * 1024 * 1024,
			Checksum:     "llama32-1b-checksum",
			LastAccessed: time.Now(),
			Popularity:   0.8,
		},
		{
			Name:         "llama3.2:8b",
			Version:      "1.0.0",
			Size:         8 * 1024 * 1024 * 1024,
			Checksum:     "llama32-8b-checksum",
			LastAccessed: time.Now(),
			Popularity:   0.9,
		},
	}
}

// ApplyConsensusOperation applies a consensus operation
func (tn *TestNode) ApplyConsensusOperation(key, value string) error {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.consensus == nil {
		return fmt.Errorf("node not active")
	}

	return tn.consensus.Apply(key, value, nil)
}

// GetConsensusValue gets a value from consensus
func (tn *TestNode) GetConsensusValue(key string) (string, bool) {
	tn.mu.RLock()
	defer tn.mu.RUnlock()

	if !tn.active || tn.consensus == nil {
		return "", false
	}

	return tn.consensus.Get(key)
}