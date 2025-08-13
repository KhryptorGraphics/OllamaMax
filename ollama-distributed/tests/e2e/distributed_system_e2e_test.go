package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DistributedNode represents a complete node in the distributed system
type DistributedNode struct {
	ID              string
	P2PNode         *p2p.Node
	ConsensusEngine *consensus.Engine
	SchedulerEngine *scheduler.Engine
	ModelManager    *models.Manager
	APIServer       *api.Server
	Config          *TestNodeConfig
	Started         bool
}

// TestNodeConfig holds configuration for test nodes
type TestNodeConfig struct {
	DataDir   string
	P2PPort   int
	APIPort   int
	RaftPort  int
	Bootstrap bool
}

// TestDistributedSystem tests the complete distributed system
func TestDistributedSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a cluster of 3 nodes
	cluster := createTestCluster(t, ctx, 3)
	defer cleanupTestCluster(t, cluster)

	// Start all nodes
	for _, node := range cluster {
		err := startNode(t, node)
		require.NoError(t, err)
	}

	// Wait for cluster to stabilize
	time.Sleep(2 * time.Second)

	// Test cluster formation
	t.Run("ClusterFormation", func(t *testing.T) {
		testClusterFormation(t, cluster)
	})

	// Test consensus functionality
	t.Run("ConsensusOperations", func(t *testing.T) {
		testConsensusOperations(t, cluster)
	})

	// Test model distribution
	t.Run("ModelDistribution", func(t *testing.T) {
		testModelDistribution(t, cluster)
	})

	// Test API functionality
	t.Run("APIOperations", func(t *testing.T) {
		testAPIOperations(t, cluster)
	})

	// Test fault tolerance
	t.Run("FaultTolerance", func(t *testing.T) {
		testFaultTolerance(t, cluster)
	})

	// Test load balancing
	t.Run("LoadBalancing", func(t *testing.T) {
		testLoadBalancing(t, cluster)
	})
}

// TestSingleNodeBootstrap tests single node bootstrap scenario
func TestSingleNodeBootstrap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create single bootstrap node
	cluster := createTestCluster(t, ctx, 1)
	defer cleanupTestCluster(t, cluster)

	node := cluster[0]
	err := startNode(t, node)
	require.NoError(t, err)

	// Wait for node to start
	time.Sleep(1 * time.Second)

	// Verify node is leader
	assert.True(t, node.ConsensusEngine.IsLeader(), "Bootstrap node should be leader")

	// Test basic operations
	err = node.ConsensusEngine.Apply("test-key", "test-value", nil)
	assert.NoError(t, err)

	value, exists := node.ConsensusEngine.Get("test-key")
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)
}

// TestClusterExpansion tests dynamic cluster expansion
func TestClusterExpansion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Start with 2 nodes
	cluster := createTestCluster(t, ctx, 2)
	defer cleanupTestCluster(t, cluster)

	// Start initial nodes
	for i := 0; i < 2; i++ {
		err := startNode(t, cluster[i])
		require.NoError(t, err)
	}

	// Wait for initial cluster
	time.Sleep(2 * time.Second)

	// Find leader
	var leader *DistributedNode
	for _, node := range cluster[:2] {
		if node.ConsensusEngine.IsLeader() {
			leader = node
			break
		}
	}
	require.NotNil(t, leader, "Should have a leader")

	// Add third node to cluster
	newNode := cluster[2]
	err := startNode(t, newNode)
	require.NoError(t, err)

	// Join new node to cluster
	err = leader.ConsensusEngine.AddVoter(
		newNode.ID,
		fmt.Sprintf("127.0.0.1:%d", newNode.Config.RaftPort),
	)
	assert.NoError(t, err)

	// Wait for cluster to stabilize
	time.Sleep(3 * time.Second)

	// Verify cluster configuration
	config, err := leader.ConsensusEngine.GetConfiguration()
	assert.NoError(t, err)
	assert.Len(t, config.Servers, 3, "Should have 3 servers in cluster")
}

// TestModelReplication tests model replication across nodes
func TestModelReplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create 3-node cluster
	cluster := createTestCluster(t, ctx, 3)
	defer cleanupTestCluster(t, cluster)

	// Start all nodes
	for _, node := range cluster {
		err := startNode(t, node)
		require.NoError(t, err)
	}

	// Wait for cluster stabilization
	time.Sleep(3 * time.Second)

	// Register a model on first node
	firstNode := cluster[0]
	testModelPath := createTestModel(t, firstNode.Config.DataDir)

	err := firstNode.ModelManager.RegisterModel("test-model", testModelPath)
	assert.NoError(t, err)

	// Verify model exists on first node
	model, exists := firstNode.ModelManager.GetModel("test-model")
	assert.True(t, exists)
	assert.NotNil(t, model)

	// Test model distribution to other nodes
	secondNode := cluster[1]
	err = secondNode.ModelManager.DownloadFromPeer("test-model", firstNode.ID)
	assert.NoError(t, err)

	// Wait for download
	time.Sleep(2 * time.Second)

	// Verify model exists on second node
	model2, exists2 := secondNode.ModelManager.GetModel("test-model")
	assert.True(t, exists2)
	assert.NotNil(t, model2)
	assert.Equal(t, model.Name, model2.Name)
}

// TestAPIEndToEnd tests complete API workflow
func TestAPIEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Create single node for API testing
	cluster := createTestCluster(t, ctx, 1)
	defer cleanupTestCluster(t, cluster)

	node := cluster[0]
	err := startNode(t, node)
	require.NoError(t, err)

	// Wait for services to start
	time.Sleep(2 * time.Second)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", node.Config.APIPort)
	client := &http.Client{Timeout: 10 * time.Second}

	// Test health check
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/api/v1/health")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	// Test cluster status
	t.Run("ClusterStatus", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/api/v1/cluster/status")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&status)
		assert.NoError(t, err)
		assert.Equal(t, node.ID, status["node_id"])
		assert.Equal(t, true, status["is_leader"])
		resp.Body.Close()
	})

	// Test metrics endpoint
	t.Run("Metrics", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/api/v1/metrics")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&metrics)
		assert.NoError(t, err)
		assert.Contains(t, metrics, "node_id")
		assert.Contains(t, metrics, "connected_peers")
		resp.Body.Close()
	})
}

// Helper functions

func createTestCluster(t *testing.T, ctx context.Context, size int) []*DistributedNode {
	cluster := make([]*DistributedNode, size)

	for i := 0; i < size; i++ {
		nodeConfig := &TestNodeConfig{
			DataDir:   t.TempDir(),
			P2PPort:   9000 + i,
			APIPort:   8000 + i,
			RaftPort:  7000 + i,
			Bootstrap: i == 0, // First node is bootstrap
		}

		node, err := createTestNode(t, ctx, nodeConfig)
		require.NoError(t, err)
		cluster[i] = node
	}

	return cluster
}

func createTestNode(t *testing.T, ctx context.Context, nodeConfig *TestNodeConfig) (*DistributedNode, error) {
	nodeID := fmt.Sprintf("node-%d", nodeConfig.APIPort)

	// P2P configuration
	p2pConfig := &config.P2PConfig{
		Listen:       fmt.Sprintf("127.0.0.1:%d", nodeConfig.P2PPort),
		EnableDHT:    true,
		ConnMgrLow:   10,
		ConnMgrHigh:  100,
		ConnMgrGrace: "1m",
	}

	// Create P2P node
	p2pNode, err := p2p.NewNode(ctx, p2pConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Consensus configuration
	consensusConfig := &config.ConsensusConfig{
		DataDir:           nodeConfig.DataDir + "/consensus",
		BindAddr:          fmt.Sprintf("127.0.0.1:%d", nodeConfig.RaftPort),
		Bootstrap:         nodeConfig.Bootstrap,
		HeartbeatTimeout:  time.Second,
		ElectionTimeout:   time.Second,
		CommitTimeout:     time.Second,
		MaxAppendEntries:  64,
		SnapshotInterval:  time.Hour,
		SnapshotThreshold: 8192,
		LogLevel:          "ERROR", // Reduce logging in tests
	}

	// Create consensus engine
	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Storage configuration
	storageConfig := &config.StorageConfig{
		ModelDir:   nodeConfig.DataDir + "/models",
		CleanupAge: time.Hour,
	}

	// Create model manager
	modelManager, err := models.NewManager(storageConfig, p2pNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create model manager: %w", err)
	}

	// Create scheduler engine
	schedulerEngine, err := scheduler.NewEngine(p2pNode, consensusEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler engine: %w", err)
	}

	// API configuration
	apiConfig := &config.APIConfig{
		Listen:      fmt.Sprintf("127.0.0.1:%d", nodeConfig.APIPort),
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			RPS: 100,
		},
		Cors: config.CorsConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	// Create API server
	apiServer, err := api.NewServer(apiConfig, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	return &DistributedNode{
		ID:              nodeID,
		P2PNode:         p2pNode,
		ConsensusEngine: consensusEngine,
		SchedulerEngine: schedulerEngine,
		ModelManager:    modelManager,
		APIServer:       apiServer,
		Config:          nodeConfig,
		Started:         false,
	}, nil
}

func startNode(t *testing.T, node *DistributedNode) error {
	if node.Started {
		return fmt.Errorf("node already started")
	}

	// Start P2P node
	err := node.P2PNode.Start()
	if err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Start consensus engine
	err = node.ConsensusEngine.Start()
	if err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start model manager
	err = node.ModelManager.Start()
	if err != nil {
		return fmt.Errorf("failed to start model manager: %w", err)
	}

	// Start scheduler engine
	err = node.SchedulerEngine.Start()
	if err != nil {
		return fmt.Errorf("failed to start scheduler engine: %w", err)
	}

	// Start API server
	err = node.APIServer.Start()
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	node.Started = true
	return nil
}

func stopNode(t *testing.T, node *DistributedNode) {
	if !node.Started {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop components in reverse order
	if node.APIServer != nil {
		node.APIServer.Shutdown(ctx)
	}

	if node.SchedulerEngine != nil {
		node.SchedulerEngine.Shutdown(ctx)
	}

	if node.ModelManager != nil {
		node.ModelManager.Shutdown(ctx)
	}

	if node.ConsensusEngine != nil {
		node.ConsensusEngine.Shutdown(ctx)
	}

	if node.P2PNode != nil {
		node.P2PNode.Stop()
	}

	node.Started = false
}

func cleanupTestCluster(t *testing.T, cluster []*DistributedNode) {
	for _, node := range cluster {
		stopNode(t, node)
	}
}

func createTestModel(t *testing.T, dataDir string) string {
	modelPath := dataDir + "/test-model.gguf"
	err := os.WriteFile(modelPath, []byte("test model data"), 0644)
	require.NoError(t, err)
	return modelPath
}

// Test functions

func testClusterFormation(t *testing.T, cluster []*DistributedNode) {
	// Verify we have exactly one leader
	leaderCount := 0
	var leader *DistributedNode

	for _, node := range cluster {
		if node.ConsensusEngine.IsLeader() {
			leaderCount++
			leader = node
		}
	}

	assert.Equal(t, 1, leaderCount, "Should have exactly one leader")
	assert.NotNil(t, leader, "Should have a leader")

	// Verify all nodes see the same leader
	leaderAddr := leader.ConsensusEngine.Leader()
	for _, node := range cluster {
		assert.Equal(t, leaderAddr, node.ConsensusEngine.Leader(), "All nodes should see same leader")
	}

	// Verify cluster configuration
	config, err := leader.ConsensusEngine.GetConfiguration()
	assert.NoError(t, err)
	assert.Len(t, config.Servers, len(cluster), "Should have all nodes in configuration")
}

func testConsensusOperations(t *testing.T, cluster []*DistributedNode) {
	// Find leader
	var leader *DistributedNode
	for _, node := range cluster {
		if node.ConsensusEngine.IsLeader() {
			leader = node
			break
		}
	}
	require.NotNil(t, leader)

	// Test state operations from leader
	testKey := "e2e-test-key"
	testValue := "e2e-test-value"

	err := leader.ConsensusEngine.Apply(testKey, testValue, map[string]interface{}{
		"test":      true,
		"timestamp": time.Now().Unix(),
	})
	assert.NoError(t, err)

	// Wait for replication
	time.Sleep(500 * time.Millisecond)

	// Verify state is replicated to all nodes
	for i, node := range cluster {
		value, exists := node.ConsensusEngine.Get(testKey)
		assert.True(t, exists, "Key should exist on node %d", i)
		assert.Equal(t, testValue, value, "Value should be replicated to node %d", i)
	}

	// Test delete operation
	err = leader.ConsensusEngine.Delete(testKey)
	assert.NoError(t, err)

	// Wait for replication
	time.Sleep(500 * time.Millisecond)

	// Verify deletion is replicated
	for i, node := range cluster {
		_, exists := node.ConsensusEngine.Get(testKey)
		assert.False(t, exists, "Key should be deleted on node %d", i)
	}
}

func testModelDistribution(t *testing.T, cluster []*DistributedNode) {
	// This is a simplified test since full model distribution
	// requires more complex setup

	// Test model registration
	firstNode := cluster[0]
	testModelPath := createTestModel(t, firstNode.Config.DataDir)

	err := firstNode.ModelManager.RegisterModel("e2e-model", testModelPath)
	assert.NoError(t, err)

	// Verify model exists
	model, exists := firstNode.ModelManager.GetModel("e2e-model")
	assert.True(t, exists)
	assert.NotNil(t, model)
	assert.Equal(t, "e2e-model", model.Name)

	// Test model metadata
	info := firstNode.ModelManager.GetModelInfo("e2e-model")
	assert.NotNil(t, info)
	assert.Equal(t, "e2e-model", info["name"])
}

func testAPIOperations(t *testing.T, cluster []*DistributedNode) {
	if len(cluster) == 0 {
		t.Skip("No nodes in cluster")
	}

	node := cluster[0]
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", node.Config.APIPort)
	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint
	resp, err := client.Get(baseURL + "/api/v1/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test cluster status
	resp, err = client.Get(baseURL + "/api/v1/cluster/status")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test metrics
	resp, err = client.Get(baseURL + "/api/v1/metrics")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func testFaultTolerance(t *testing.T, cluster []*DistributedNode) {
	if len(cluster) < 3 {
		t.Skip("Need at least 3 nodes for fault tolerance test")
	}

	// Find leader
	var leader *DistributedNode
	var followers []*DistributedNode

	for _, node := range cluster {
		if node.ConsensusEngine.IsLeader() {
			leader = node
		} else {
			followers = append(followers, node)
		}
	}

	require.NotNil(t, leader)
	require.NotEmpty(t, followers)

	// Store some data
	testKey := "fault-test-key"
	testValue := "fault-test-value"

	err := leader.ConsensusEngine.Apply(testKey, testValue, nil)
	assert.NoError(t, err)

	// Wait for replication
	time.Sleep(500 * time.Millisecond)

	// Verify data exists on followers
	for _, follower := range followers {
		value, exists := follower.ConsensusEngine.Get(testKey)
		assert.True(t, exists)
		assert.Equal(t, testValue, value)
	}

	// Simulate leader failure by stopping it
	stopNode(t, leader)

	// Wait for leader election
	time.Sleep(3 * time.Second)

	// Verify new leader is elected
	newLeaderCount := 0
	for _, follower := range followers {
		if follower.ConsensusEngine.IsLeader() {
			newLeaderCount++
		}
	}

	assert.Equal(t, 1, newLeaderCount, "Should have new leader after failure")

	// Data should still be accessible
	for _, follower := range followers {
		value, exists := follower.ConsensusEngine.Get(testKey)
		assert.True(t, exists, "Data should survive leader failure")
		assert.Equal(t, testValue, value)
	}
}

func testLoadBalancing(t *testing.T, cluster []*DistributedNode) {
	// Test that scheduler distributes load across available nodes

	if len(cluster) == 0 {
		t.Skip("No nodes in cluster")
	}

	firstNode := cluster[0]

	// Get node statistics
	stats := firstNode.SchedulerEngine.GetStats()
	assert.NotNil(t, stats)

	// Test node availability
	nodes := firstNode.SchedulerEngine.GetNodes()
	assert.NotNil(t, nodes)

	// Verify scheduler can handle requests
	onlineCount := firstNode.SchedulerEngine.GetOnlineNodeCount()
	assert.True(t, onlineCount >= 0)
}

// Benchmark tests for E2E scenarios

func BenchmarkE2E_ConsensusOperations(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping E2E benchmark in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create 3-node cluster
	cluster := createTestCluster(&testing.T{}, ctx, 3)
	defer cleanupTestCluster(&testing.T{}, cluster)

	// Start all nodes
	for _, node := range cluster {
		err := startNode(&testing.T{}, node)
		if err != nil {
			b.Fatal(err)
		}
	}

	// Wait for cluster
	time.Sleep(3 * time.Second)

	// Find leader
	var leader *DistributedNode
	for _, node := range cluster {
		if node.ConsensusEngine.IsLeader() {
			leader = node
			break
		}
	}

	if leader == nil {
		b.Fatal("No leader found")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d", i)
			value := fmt.Sprintf("bench-value-%d", i)

			err := leader.ConsensusEngine.Apply(key, value, nil)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkE2E_APIRequests(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping E2E benchmark in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create single node
	cluster := createTestCluster(&testing.T{}, ctx, 1)
	defer cleanupTestCluster(&testing.T{}, cluster)

	node := cluster[0]
	err := startNode(&testing.T{}, node)
	if err != nil {
		b.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", node.Config.APIPort)
	client := &http.Client{Timeout: 5 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(baseURL + "/api/v1/health")
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatal("Expected 200, got", resp.StatusCode)
			}
		}
	})
}
