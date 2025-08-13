package performance

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/auth"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/stretchr/testify/require"
)

// BenchmarkSuite represents a complete benchmark test suite
type BenchmarkSuite struct {
	cluster        *TestCluster
	workloadConfig *WorkloadConfig
}

// TestCluster represents a cluster for performance testing
type TestCluster struct {
	Nodes  []*TestNode
	Size   int
	Config *ClusterConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// TestNode represents a single node for benchmarking
type TestNode struct {
	ID              string
	P2PNode         *p2p.Node
	ConsensusEngine *consensus.Engine
	SchedulerEngine *scheduler.Engine
	ModelManager    *models.Manager
	AuthManager     *auth.Manager
	APIServer       *api.Server
	Config          *NodeConfig
}

// NodeConfig holds node configuration
type NodeConfig struct {
	DataDir   string
	P2PPort   int
	APIPort   int
	RaftPort  int
	Bootstrap bool
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	Size           int
	NetworkLatency time.Duration
	DiskIOLatency  time.Duration
	MemoryLimitMB  int
}

// WorkloadConfig defines benchmark workload parameters
type WorkloadConfig struct {
	ConcurrentClients int
	RequestsPerSecond int
	Duration          time.Duration
	DataSize          int
	ModelSize         int64
}

// PerformanceMetrics holds benchmark results
type PerformanceMetrics struct {
	Throughput    float64
	Latency       LatencyMetrics
	ResourceUsage ResourceMetrics
	ErrorRate     float64
	NetworkIO     NetworkMetrics
	DiskIO        DiskMetrics
}

// LatencyMetrics holds latency statistics
type LatencyMetrics struct {
	Mean  time.Duration
	P50   time.Duration
	P95   time.Duration
	P99   time.Duration
	P99_9 time.Duration
	Max   time.Duration
}

// ResourceMetrics holds resource usage statistics
type ResourceMetrics struct {
	CPUUsage       float64
	MemoryUsage    int64
	GoroutineCount int
	GCPauses       []time.Duration
}

// NetworkMetrics holds network I/O statistics
type NetworkMetrics struct {
	BytesIn    int64
	BytesOut   int64
	PacketsIn  int64
	PacketsOut int64
}

// DiskMetrics holds disk I/O statistics
type DiskMetrics struct {
	BytesRead    int64
	BytesWritten int64
	IOOperations int64
}

// Comprehensive benchmark tests

func BenchmarkConsensusOperations(b *testing.B) {
	suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
		ConcurrentClients: 10,
		Duration:          30 * time.Second,
		DataSize:          1024,
	})
	defer suite.cleanup(b)

	b.Run("SingleThreaded", func(b *testing.B) {
		suite.benchmarkConsensusOperations(b, 1)
	})

	b.Run("MultiThreaded", func(b *testing.B) {
		suite.benchmarkConsensusOperations(b, runtime.NumCPU())
	})

	b.Run("HighConcurrency", func(b *testing.B) {
		suite.benchmarkConsensusOperations(b, 100)
	})
}

func BenchmarkP2PNetworking(b *testing.B) {
	suite := setupBenchmarkSuite(b, 5, &WorkloadConfig{
		ConcurrentClients: 20,
		Duration:          30 * time.Second,
		DataSize:          4096,
	})
	defer suite.cleanup(b)

	b.Run("PeerDiscovery", func(b *testing.B) {
		suite.benchmarkPeerDiscovery(b)
	})

	b.Run("MessageBroadcast", func(b *testing.B) {
		suite.benchmarkMessageBroadcast(b)
	})

	b.Run("ContentRouting", func(b *testing.B) {
		suite.benchmarkContentRouting(b)
	})
}

func BenchmarkModelDistribution(b *testing.B) {
	suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
		ConcurrentClients: 5,
		Duration:          60 * time.Second,
		ModelSize:         100 * 1024 * 1024, // 100MB
	})
	defer suite.cleanup(b)

	b.Run("ModelDownload", func(b *testing.B) {
		suite.benchmarkModelDownload(b)
	})

	b.Run("ModelReplication", func(b *testing.B) {
		suite.benchmarkModelReplication(b)
	})

	b.Run("ConcurrentAccess", func(b *testing.B) {
		suite.benchmarkConcurrentModelAccess(b)
	})
}

func BenchmarkAPIEndpoints(b *testing.B) {
	suite := setupBenchmarkSuite(b, 1, &WorkloadConfig{
		ConcurrentClients: 50,
		Duration:          30 * time.Second,
		DataSize:          1024,
	})
	defer suite.cleanup(b)

	b.Run("HealthCheck", func(b *testing.B) {
		suite.benchmarkAPIEndpoint(b, "GET", "/api/v1/health", nil)
	})

	b.Run("ClusterStatus", func(b *testing.B) {
		suite.benchmarkAPIEndpoint(b, "GET", "/api/v1/cluster/status", nil)
	})

	b.Run("NodeOperations", func(b *testing.B) {
		suite.benchmarkAPIEndpoint(b, "GET", "/api/v1/nodes", nil)
	})

	b.Run("ModelOperations", func(b *testing.B) {
		suite.benchmarkAPIEndpoint(b, "GET", "/api/v1/models", nil)
	})
}

func BenchmarkAuthenticationSystem(b *testing.B) {
	suite := setupBenchmarkSuite(b, 1, &WorkloadConfig{
		ConcurrentClients: 100,
		Duration:          30 * time.Second,
	})
	defer suite.cleanup(b)

	b.Run("JWTValidation", func(b *testing.B) {
		suite.benchmarkJWTValidation(b)
	})

	b.Run("APIKeyValidation", func(b *testing.B) {
		suite.benchmarkAPIKeyValidation(b)
	})

	b.Run("PermissionCheck", func(b *testing.B) {
		suite.benchmarkPermissionCheck(b)
	})
}

func BenchmarkSchedulerEngine(b *testing.B) {
	suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
		ConcurrentClients: 20,
		Duration:          30 * time.Second,
	})
	defer suite.cleanup(b)

	b.Run("TaskScheduling", func(b *testing.B) {
		suite.benchmarkTaskScheduling(b)
	})

	b.Run("LoadBalancing", func(b *testing.B) {
		suite.benchmarkLoadBalancing(b)
	})

	b.Run("ResourceAllocation", func(b *testing.B) {
		suite.benchmarkResourceAllocation(b)
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
		ConcurrentClients: 50,
		Duration:          60 * time.Second,
	})
	defer suite.cleanup(b)

	b.Run("MemoryEfficiency", func(b *testing.B) {
		suite.benchmarkMemoryEfficiency(b)
	})

	b.Run("GarbageCollection", func(b *testing.B) {
		suite.benchmarkGarbageCollection(b)
	})
}

func BenchmarkConcurrentOperations(b *testing.B) {
	suite := setupBenchmarkSuite(b, 5, &WorkloadConfig{
		ConcurrentClients: 100,
		Duration:          45 * time.Second,
	})
	defer suite.cleanup(b)

	b.Run("MixedWorkload", func(b *testing.B) {
		suite.benchmarkMixedWorkload(b)
	})

	b.Run("StressTest", func(b *testing.B) {
		suite.benchmarkStressTest(b)
	})
}

// Setup and utility functions

func setupBenchmarkSuite(b *testing.B, clusterSize int, workloadConfig *WorkloadConfig) *BenchmarkSuite {
	clusterConfig := &ClusterConfig{
		Size:           clusterSize,
		NetworkLatency: 1 * time.Millisecond,
		DiskIOLatency:  5 * time.Millisecond,
		MemoryLimitMB:  512,
	}

	cluster := createBenchmarkCluster(b, clusterConfig)
	err := cluster.start(b)
	require.NoError(b, err)

	// Wait for cluster stabilization
	time.Sleep(2 * time.Second)

	return &BenchmarkSuite{
		cluster:        cluster,
		workloadConfig: workloadConfig,
	}
}

func (s *BenchmarkSuite) cleanup(b *testing.B) {
	if s.cluster != nil {
		s.cluster.cleanup(b)
	}
}

func createBenchmarkCluster(b *testing.B, config *ClusterConfig) *TestCluster {
	ctx, cancel := context.WithCancel(context.Background())

	cluster := &TestCluster{
		Nodes:  make([]*TestNode, config.Size),
		Size:   config.Size,
		Config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < config.Size; i++ {
		node := createBenchmarkNode(b, i, config)
		cluster.Nodes[i] = node
	}

	return cluster
}

func createBenchmarkNode(b *testing.B, index int, config *ClusterConfig) *TestNode {
	nodeConfig := &NodeConfig{
		DataDir:   b.TempDir(),
		P2PPort:   9000 + index,
		APIPort:   8000 + index,
		RaftPort:  7000 + index,
		Bootstrap: index == 0,
	}

	nodeID := fmt.Sprintf("bench-node-%d", index)

	// Create P2P node
	ctx := context.Background()
	p2pNode, err := p2p.NewP2PNode(ctx, nil)
	require.NoError(b, err)

	// Create consensus engine
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
		LogLevel:          "ERROR",
	}

	consensusEngine, err := consensus.NewEngine(consensusConfig, p2pNode)
	require.NoError(b, err)

	// Create scheduler engine
	schedulerEngine, err := scheduler.NewEngine(p2pNode, consensusEngine)
	require.NoError(b, err)

	// Create model manager
	storageConfig := &config.StorageConfig{
		ModelDir:   nodeConfig.DataDir + "/models",
		CleanupAge: time.Hour,
	}

	modelManager, err := models.NewManager(storageConfig, p2pNode)
	require.NoError(b, err)

	// Create auth manager
	authConfig := &config.AuthConfig{
		SecretKey:   "benchmark-secret-key",
		TokenExpiry: time.Hour,
		Issuer:      "benchmark-issuer",
		Audience:    "benchmark-audience",
	}

	authManager, err := auth.NewManager(authConfig)
	require.NoError(b, err)

	// Create API server
	apiConfig := &config.APIConfig{
		Listen:      fmt.Sprintf("127.0.0.1:%d", nodeConfig.APIPort),
		MaxBodySize: 1024 * 1024,
		RateLimit: config.RateLimitConfig{
			RPS: 1000, // High limit for benchmarks
		},
		Cors: config.CorsConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: false,
			MaxAge:           3600,
		},
	}

	apiServer, err := api.NewServer(apiConfig, p2pNode, consensusEngine, schedulerEngine)
	require.NoError(b, err)

	return &TestNode{
		ID:              nodeID,
		P2PNode:         p2pNode,
		ConsensusEngine: consensusEngine,
		SchedulerEngine: schedulerEngine,
		ModelManager:    modelManager,
		AuthManager:     authManager,
		APIServer:       apiServer,
		Config:          nodeConfig,
	}
}

func (c *TestCluster) start(b *testing.B) error {
	for _, node := range c.Nodes {
		err := node.start(b)
		if err != nil {
			return fmt.Errorf("failed to start node %s: %w", node.ID, err)
		}
	}
	return nil
}

func (c *TestCluster) cleanup(b *testing.B) {
	c.cancel()

	for _, node := range c.Nodes {
		if node != nil {
			node.stop(b)
		}
	}
}

func (n *TestNode) start(b *testing.B) error {
	err := n.P2PNode.Start()
	if err != nil {
		return err
	}

	err = n.ConsensusEngine.Start()
	if err != nil {
		return err
	}

	err = n.SchedulerEngine.Start()
	if err != nil {
		return err
	}

	err = n.ModelManager.Start()
	if err != nil {
		return err
	}

	err = n.APIServer.Start()
	if err != nil {
		return err
	}

	return nil
}

func (n *TestNode) stop(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if n.APIServer != nil {
		n.APIServer.Shutdown(ctx)
	}

	if n.ModelManager != nil {
		n.ModelManager.Shutdown(ctx)
	}

	if n.SchedulerEngine != nil {
		n.SchedulerEngine.Shutdown(ctx)
	}

	if n.ConsensusEngine != nil {
		n.ConsensusEngine.Shutdown(ctx)
	}

	if n.P2PNode != nil {
		n.P2PNode.Stop()
	}

	if n.AuthManager != nil {
		n.AuthManager.Close()
	}
}

// Benchmark implementations

func (s *BenchmarkSuite) benchmarkConsensusOperations(b *testing.B, concurrency int) {
	leader := s.findLeader()
	if leader == nil {
		b.Fatal("No leader found")
	}

	b.ResetTimer()
	b.SetParallelism(concurrency)
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			key := fmt.Sprintf("bench-key-%d-%d", b.N, counter)
			value := fmt.Sprintf("bench-value-%d", counter)

			err := leader.ConsensusEngine.Apply(key, value, nil)
			if err != nil {
				b.Fatal(err)
			}
			counter++
		}
	})
}

func (s *BenchmarkSuite) benchmarkPeerDiscovery(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate peer discovery operations
		peers := node.P2PNode.GetConnectedPeers()
		_ = len(peers)

		status := node.P2PNode.GetStatus()
		_ = status.ConnectedPeers
	}
}

func (s *BenchmarkSuite) benchmarkMessageBroadcast(b *testing.B) {
	node := s.cluster.Nodes[0]
	testData := make([]byte, s.workloadConfig.DataSize)
	rand.Read(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate message broadcasting
		// This would normally use P2P messaging protocols
		for _, peer := range node.P2PNode.GetConnectedPeers() {
			_ = peer
			// Simulate message send
		}
	}
}

func (s *BenchmarkSuite) benchmarkContentRouting(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contentID := fmt.Sprintf("content-%d", i)
		// Simulate content routing
		_, _, err := node.P2PNode.FindContent(context.Background(), contentID)
		// Error is expected for non-existent content
		_ = err
	}
}

func (s *BenchmarkSuite) benchmarkModelDownload(b *testing.B) {
	sourceNode := s.cluster.Nodes[0]
	targetNode := s.cluster.Nodes[1]

	// Create test model
	testModelPath := s.createTestModel(b, sourceNode)
	err := sourceNode.ModelManager.RegisterModel("bench-model", testModelPath)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modelName := fmt.Sprintf("bench-model-%d", i)
		err := targetNode.ModelManager.DownloadFromPeer(modelName, sourceNode.ID)
		// Error is expected for non-existent models in this benchmark
		_ = err
	}
}

func (s *BenchmarkSuite) benchmarkModelReplication(b *testing.B) {
	sourceNode := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modelName := fmt.Sprintf("replication-model-%d", i)

		// Simulate model replication across cluster
		for _, node := range s.cluster.Nodes[1:] {
			err := node.ModelManager.DownloadFromPeer(modelName, sourceNode.ID)
			_ = err
		}
	}
}

func (s *BenchmarkSuite) benchmarkConcurrentModelAccess(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			modelName := fmt.Sprintf("concurrent-model-%d", rand.Intn(100))
			_, exists := node.ModelManager.GetModel(modelName)
			_ = exists
		}
	})
}

func (s *BenchmarkSuite) benchmarkAPIEndpoint(b *testing.B, method, path string, body interface{}) {
	// This would implement HTTP client benchmarking
	// For now, simulate API operations
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate API call processing
			switch path {
			case "/api/v1/health":
				// Health check simulation
				_ = node.P2PNode.GetStatus()
			case "/api/v1/cluster/status":
				// Cluster status simulation
				_ = node.ConsensusEngine.IsLeader()
			case "/api/v1/nodes":
				// Node listing simulation
				_ = node.SchedulerEngine.GetNodes()
			case "/api/v1/models":
				// Model listing simulation
				_ = node.ModelManager.GetAllModels()
			}
		}
	})
}

func (s *BenchmarkSuite) benchmarkJWTValidation(b *testing.B) {
	node := s.cluster.Nodes[0]

	// Create test user and get token
	authCtx, err := node.AuthManager.Authenticate("admin", "admin123", map[string]string{
		"ip_address": "127.0.0.1",
		"user_agent": "benchmark",
	})
	require.NoError(b, err)

	token := authCtx.TokenString

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := node.AuthManager.ValidateToken(token)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (s *BenchmarkSuite) benchmarkAPIKeyValidation(b *testing.B) {
	node := s.cluster.Nodes[0]

	// Create test API key
	user := &auth.CreateUserRequest{
		Username: "apiuser",
		Password: "password123",
		Email:    "api@test.com",
		Role:     auth.RoleUser,
	}

	createdUser, err := node.AuthManager.CreateUser(user)
	require.NoError(b, err)

	apiKeyReq := &auth.CreateAPIKeyRequest{
		Name: "benchmark-key",
	}

	_, rawKey, err := node.AuthManager.CreateAPIKey(createdUser.ID, apiKeyReq)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := node.AuthManager.ValidateAPIKey(rawKey)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func (s *BenchmarkSuite) benchmarkPermissionCheck(b *testing.B) {
	node := s.cluster.Nodes[0]

	// Create auth context
	authCtx, err := node.AuthManager.Authenticate("admin", "admin123", nil)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			hasPermission := node.AuthManager.HasPermission(authCtx, auth.PermissionSystemAdmin)
			_ = hasPermission
		}
	})
}

func (s *BenchmarkSuite) benchmarkTaskScheduling(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &scheduler.Request{
			ID:         fmt.Sprintf("bench-task-%d", i),
			ModelName:  "test-model",
			Priority:   1,
			Timeout:    30 * time.Second,
			ResponseCh: make(chan *scheduler.Response, 1),
		}

		err := node.SchedulerEngine.Schedule(req)
		_ = err // Might fail due to no available nodes
	}
}

func (s *BenchmarkSuite) benchmarkLoadBalancing(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate load balancing decisions
		nodes := node.SchedulerEngine.GetNodes()
		availableNodes := node.SchedulerEngine.GetAvailableNodes()
		_ = len(nodes)
		_ = len(availableNodes)
	}
}

func (s *BenchmarkSuite) benchmarkResourceAllocation(b *testing.B) {
	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate resource allocation
		stats := node.SchedulerEngine.GetStats()
		onlineNodes := node.SchedulerEngine.GetOnlineNodeCount()
		_ = stats
		_ = onlineNodes
	}
}

func (s *BenchmarkSuite) benchmarkMemoryEfficiency(b *testing.B) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	initialAlloc := memStats.Alloc

	node := s.cluster.Nodes[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Perform memory-intensive operations
		key := fmt.Sprintf("memory-test-%d", i)
		value := make([]byte, 1024) // 1KB per operation

		err := node.ConsensusEngine.Apply(key, string(value), nil)
		if err != nil {
			b.Fatal(err)
		}
	}

	runtime.ReadMemStats(&memStats)
	finalAlloc := memStats.Alloc

	b.ReportMetric(float64(finalAlloc-initialAlloc)/float64(b.N), "bytes/op")
}

func (s *BenchmarkSuite) benchmarkGarbageCollection(b *testing.B) {
	var memStats runtime.MemStats

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate memory that will be GC'd
		data := make([]byte, 10*1024) // 10KB
		_ = data

		if i%1000 == 0 {
			runtime.GC()
			runtime.ReadMemStats(&memStats)
		}
	}

	runtime.ReadMemStats(&memStats)
	b.ReportMetric(float64(memStats.NumGC), "gc-cycles")
	b.ReportMetric(float64(memStats.PauseTotalNs)/1e6, "gc-pause-ms")
}

func (s *BenchmarkSuite) benchmarkMixedWorkload(b *testing.B) {
	leader := s.findLeader()
	if leader == nil {
		b.Fatal("No leader found")
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			operation := counter % 4
			counter++

			switch operation {
			case 0: // Consensus operation
				key := fmt.Sprintf("mixed-key-%d", counter)
				err := leader.ConsensusEngine.Apply(key, "value", nil)
				_ = err
			case 1: // P2P operation
				_ = leader.P2PNode.GetConnectedPeers()
			case 2: // Model operation
				_, _ = leader.ModelManager.GetModel("nonexistent")
			case 3: // Scheduler operation
				_ = leader.SchedulerEngine.GetStats()
			}
		}
	})
}

func (s *BenchmarkSuite) benchmarkStressTest(b *testing.B) {
	// High-intensity stress test
	var wg sync.WaitGroup
	numWorkers := s.workloadConfig.ConcurrentClients

	b.ResetTimer()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			node := s.cluster.Nodes[workerID%len(s.cluster.Nodes)]

			for i := 0; i < b.N/numWorkers; i++ {
				// Mixed high-intensity operations
				key := fmt.Sprintf("stress-key-%d-%d", workerID, i)
				value := fmt.Sprintf("stress-value-%d", i)

				if node.ConsensusEngine.IsLeader() {
					err := node.ConsensusEngine.Apply(key, value, nil)
					_ = err
				} else {
					_, _ = node.ConsensusEngine.Get(key)
				}

				// Add small delay to prevent overwhelming
				if i%100 == 0 {
					time.Sleep(time.Microsecond)
				}
			}
		}(w)
	}

	wg.Wait()
}

// Utility functions

func (s *BenchmarkSuite) findLeader() *TestNode {
	for _, node := range s.cluster.Nodes {
		if node.ConsensusEngine.IsLeader() {
			return node
		}
	}
	return nil
}

func (s *BenchmarkSuite) createTestModel(b *testing.B, node *TestNode) string {
	modelPath := node.Config.DataDir + "/test-model.gguf"
	modelData := make([]byte, s.workloadConfig.ModelSize)
	rand.Read(modelData)

	err := os.WriteFile(modelPath, modelData, 0644)
	require.NoError(b, err)

	return modelPath
}

func (s *BenchmarkSuite) collectMetrics(b *testing.B) *PerformanceMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &PerformanceMetrics{
		ResourceUsage: ResourceMetrics{
			MemoryUsage:    int64(memStats.Alloc),
			GoroutineCount: runtime.NumGoroutine(),
		},
	}

	// Collect GC pause times
	for i := 0; i < 256 && i < len(memStats.PauseNs); i++ {
		if memStats.PauseNs[i] > 0 {
			metrics.ResourceUsage.GCPauses = append(
				metrics.ResourceUsage.GCPauses,
				time.Duration(memStats.PauseNs[i]),
			)
		}
	}

	return metrics
}

// Specialized benchmark tests

func BenchmarkConsensusScalability(b *testing.B) {
	clusterSizes := []int{1, 3, 5, 7}

	for _, size := range clusterSizes {
		b.Run(fmt.Sprintf("Nodes_%d", size), func(b *testing.B) {
			suite := setupBenchmarkSuite(b, size, &WorkloadConfig{
				ConcurrentClients: 10,
				DataSize:          1024,
			})
			defer suite.cleanup(b)

			leader := suite.findLeader()
			if leader == nil {
				b.Fatal("No leader found")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("scale-key-%d", i)
				value := fmt.Sprintf("scale-value-%d", i)

				err := leader.ConsensusEngine.Apply(key, value, nil)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDataSizeImpact(b *testing.B) {
	dataSizes := []int{64, 256, 1024, 4096, 16384}

	for _, size := range dataSizes {
		b.Run(fmt.Sprintf("Size_%dB", size), func(b *testing.B) {
			suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
				DataSize: size,
			})
			defer suite.cleanup(b)

			leader := suite.findLeader()
			if leader == nil {
				b.Fatal("No leader found")
			}

			testData := make([]byte, size)
			rand.Read(testData)

			b.SetBytes(int64(size))
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("data-key-%d", i)
				value := string(testData)

				err := leader.ConsensusEngine.Apply(key, value, nil)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkConcurrencyLevels(b *testing.B) {
	concurrencyLevels := []int{1, 5, 10, 25, 50, 100}

	for _, level := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", level), func(b *testing.B) {
			suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
				ConcurrentClients: level,
			})
			defer suite.cleanup(b)

			suite.benchmarkConsensusOperations(b, level)
		})
	}
}

// Memory and resource benchmarks

func BenchmarkMemoryLeaks(b *testing.B) {
	suite := setupBenchmarkSuite(b, 1, &WorkloadConfig{})
	defer suite.cleanup(b)

	node := suite.cluster.Nodes[0]

	var initialMemStats, finalMemStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMemStats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Operations that might cause memory leaks
		key := fmt.Sprintf("leak-test-%d", i)
		value := make([]byte, 1024)

		err := node.ConsensusEngine.Apply(key, string(value), nil)
		if err != nil {
			b.Fatal(err)
		}

		// Force cleanup
		if i%1000 == 0 {
			runtime.GC()
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&finalMemStats)

	memGrowth := finalMemStats.Alloc - initialMemStats.Alloc
	b.ReportMetric(float64(memGrowth), "memory-growth-bytes")
	b.ReportMetric(float64(finalMemStats.NumGC-initialMemStats.NumGC), "gc-cycles")
}

func BenchmarkGoroutineUsage(b *testing.B) {
	suite := setupBenchmarkSuite(b, 3, &WorkloadConfig{
		ConcurrentClients: 50,
	})
	defer suite.cleanup(b)

	initialGoroutines := runtime.NumGoroutine()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Operations that create goroutines
			node := suite.cluster.Nodes[0]
			_ = node.P2PNode.GetStatus()
		}
	})

	finalGoroutines := runtime.NumGoroutine()
	b.ReportMetric(float64(finalGoroutines-initialGoroutines), "goroutine-growth")
}
