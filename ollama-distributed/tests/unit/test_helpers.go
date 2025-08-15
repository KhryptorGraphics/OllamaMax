package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	internalConfig "github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/types"
	"github.com/stretchr/testify/require"
)

// MockDistributedScheduler is a mock implementation of the distributed scheduler
type MockDistributedScheduler struct {
	nodes []types.NodeInfo
}

func (m *MockDistributedScheduler) GetAvailableNodes() []types.NodeInfo {
	return m.nodes
}

func (m *MockDistributedScheduler) ScheduleTask(task *types.DistributedTask) error {
	return nil
}

func (m *MockDistributedScheduler) GetNodeLoad(nodeID string) (float64, error) {
	return 0.5, nil
}

func (m *MockDistributedScheduler) Shutdown(ctx context.Context) error {
	return nil
}

// createMockP2PNode creates a mock P2P node for testing
func createMockP2PNode(t *testing.T) *p2p.P2PNode {
	ctx := context.Background()

	// Use the pkg/config NodeConfig which has the correct fields
	pkgConfig := &config.NodeConfig{
		Listen:       []string{"/ip4/127.0.0.1/tcp/0"},
		StaticRelays: []string{},
		EnableNoise:  false,
	}

	node, err := p2p.NewP2PNode(ctx, pkgConfig)
	require.NoError(t, err)
	return node
}

// createMockConsensusEngine creates a mock consensus engine for testing
func createMockConsensusEngine(t *testing.T) *consensus.Engine {
	tempDir := t.TempDir()

	// Use internal config for consensus
	consensusConfig := &internalConfig.ConsensusConfig{
		DataDir:           tempDir,
		NodeID:            "test-node",
		BindAddr:          "127.0.0.1:0",
		AdvertiseAddr:     "127.0.0.1:0",
		Bootstrap:         false,
		BootstrapExpect:   1,
		HeartbeatTimeout:  1000 * time.Millisecond,
		ElectionTimeout:   1000 * time.Millisecond,
		CommitTimeout:     50 * time.Millisecond,
		MaxAppendEntries:  64,
		SnapshotInterval:  120 * time.Second,
		SnapshotThreshold: 8192,
	}

	p2pNode := createMockP2PNode(t)

	// Create real message router and network monitor with minimal config
	messageRouter := messaging.NewMessageRouter(nil)    // Uses default config
	networkMonitor := monitoring.NewNetworkMonitor(nil) // Uses default config

	engine, err := consensus.NewEngine(consensusConfig, p2pNode, messageRouter, networkMonitor)
	require.NoError(t, err)
	return engine
}

// createMockSchedulerEngine creates a mock scheduler engine for testing
func createMockSchedulerEngine(t *testing.T) *scheduler.Engine {
	schedulerConfig := &internalConfig.SchedulerConfig{
		Algorithm:           "round_robin",
		LoadBalancing:       "round_robin",
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:          3,
		RetryDelay:          5 * time.Second,
		QueueSize:           100,
		WorkerCount:         10,
	}

	p2pNode := createMockP2PNode(t)
	consensusEngine := createMockConsensusEngine(t)

	engine, err := scheduler.NewEngine(schedulerConfig, p2pNode, consensusEngine)
	require.NoError(t, err)

	// Add a test node for testing purposes
	testNode := &scheduler.NodeInfo{
		ID:      "test-node-id",
		Address: "127.0.0.1:8080",
		Status:  scheduler.NodeStatusOnline,
		Capacity: scheduler.NodeCapacity{
			CPU:    8,
			Memory: 16 * 1024 * 1024 * 1024,   // 16GB
			Disk:   1024 * 1024 * 1024 * 1024, // 1TB
			GPU:    1,
		},
		Usage: scheduler.NodeUsage{
			CPU:    0.5,
			Memory: 0.3,
			Disk:   0.2,
			GPU:    0.0,
		},
		Models:   []string{},
		LastSeen: time.Now(),
		Metadata: map[string]string{},
	}

	// Directly add to the nodes map for testing (accessing private field)
	// This is acceptable in test code
	engine.AddTestNode(testNode)

	return engine
}

// getAuthToken gets an authentication token for testing
func getAuthToken(t *testing.T, server *api.Server) string {
	loginRequest := map[string]interface{}{
		"username": "testuser",
		"password": "testpass",
	}

	jsonData, err := json.Marshal(loginRequest)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Login failed with status %d: %s", w.Code, w.Body.String())
		return ""
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	token, ok := response["token"].(string)
	require.True(t, ok, "Token not found in login response")

	return token
}

// createTestAPIServer creates a test API server for testing
func createTestAPIServer(t testing.TB) *api.Server {
	apiConfig := &internalConfig.APIConfig{
		Listen:      "127.0.0.1:0",
		Timeout:     30 * time.Second,
		MaxBodySize: 32 * 1024 * 1024,
	}

	p2pNode := createMockP2PNode(t.(*testing.T))
	consensusEngine := createMockConsensusEngine(t.(*testing.T))
	schedulerEngine := createMockSchedulerEngine(t.(*testing.T))

	server, err := api.NewServer(apiConfig, p2pNode, consensusEngine, schedulerEngine)
	require.NoError(t, err)
	return server
}

// MockMessageRouter is a mock implementation of the message router
type MockMessageRouter struct{}

func (m *MockMessageRouter) Route(message interface{}) error {
	return nil
}

func (m *MockMessageRouter) Subscribe(topic string, handler func(interface{})) error {
	return nil
}

func (m *MockMessageRouter) Unsubscribe(topic string) error {
	return nil
}

func (m *MockMessageRouter) Publish(topic string, message interface{}) error {
	return nil
}

func (m *MockMessageRouter) Start(ctx context.Context) error {
	return nil
}

func (m *MockMessageRouter) Stop() error {
	return nil
}

// MockNetworkMonitor is a mock implementation of the network monitor
type MockNetworkMonitor struct{}

func (m *MockNetworkMonitor) GetPeerCount() int {
	return 0
}

func (m *MockNetworkMonitor) GetConnectedPeers() []string {
	return []string{}
}

func (m *MockNetworkMonitor) GetNetworkStats() map[string]interface{} {
	return map[string]interface{}{
		"peers":       0,
		"connections": 0,
		"bandwidth":   0,
	}
}

func (m *MockNetworkMonitor) Start(ctx context.Context) error {
	return nil
}

func (m *MockNetworkMonitor) Stop() error {
	return nil
}

func (m *MockNetworkMonitor) IsHealthy() bool {
	return true
}

// MockNode is a mock implementation of a cluster node
type MockNode struct {
	id       string
	address  string
	status   string
	metadata map[string]string
}

func (m *MockNode) GetID() string {
	return m.id
}

func (m *MockNode) GetAddress() string {
	return m.address
}

func (m *MockNode) GetStatus() string {
	return m.status
}

func (m *MockNode) GetMetadata() map[string]string {
	return m.metadata
}

func (m *MockNode) IsHealthy() bool {
	return m.status == "healthy"
}

// createMockNodes creates a set of mock nodes for testing
func createMockNodes() []types.NodeInfo {
	return []types.NodeInfo{
		{
			ID:       "node-1",
			Address:  "127.0.0.1:8001",
			Status:   "healthy",
			Metadata: map[string]string{"region": "us-west"},
		},
		{
			ID:       "node-2",
			Address:  "127.0.0.1:8002",
			Status:   "healthy",
			Metadata: map[string]string{"region": "us-east"},
		},
	}
}

// createTestSyncManager creates a test sync manager
func createTestSyncManager(t *testing.T) *models.SyncManager {
	// Create test configuration
	config := &internalConfig.SyncConfig{
		SyncInterval: 10 * time.Second,
		WorkerCount:  2,
		ChunkSize:    1024 * 1024, // 1MB
		DeltaDir:     t.TempDir(),
		CASDir:       t.TempDir(),
	}

	// Create mock components
	p2pNode := createMockP2PNode(t)
	modelManager := createMockModelManager(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create sync manager
	syncManager, err := models.NewSyncManager(config, p2pNode, modelManager, logger)
	require.NoError(t, err)

	return syncManager
}

// createMockModelManager creates a mock model manager
func createMockModelManager(t *testing.T) *models.Manager {
	// Create test configuration
	config := &internalConfig.StorageConfig{
		DataDir:     t.TempDir(),
		MaxDiskSize: 1024 * 1024 * 1024, // 1GB
		ModelDir:    t.TempDir(),
		CacheDir:    t.TempDir(),
		CleanupAge:  24 * time.Hour,
	}

	// Create mock P2P node
	p2pNode := createMockP2PNode(t)

	// Create model manager
	manager, err := models.NewManager(config, p2pNode)
	require.NoError(t, err)

	return manager
}
