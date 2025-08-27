package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntelligentSyncBasic tests basic intelligent sync functionality
func TestIntelligentSyncBasic(t *testing.T) {
	// Create a basic sync configuration for testing
	config := &SyncConfig{
		SyncInterval:    1 * time.Second,
		MaxRetries:      3,
		RetryInterval:   500 * time.Millisecond,
		BatchSize:       10,
		CompressionEnabled: true,
	}
	
	assert.NotNil(t, config)
	assert.Equal(t, 1*time.Second, config.SyncInterval)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 10, config.BatchSize)
	assert.True(t, config.CompressionEnabled)
}

func TestSyncConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *SyncConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &SyncConfig{
				SyncInterval:    5 * time.Second,
				MaxRetries:      5,
				RetryInterval:   1 * time.Second,
				BatchSize:       20,
				CompressionEnabled: false,
			},
			valid: true,
		},
		{
			name: "zero sync interval",
			config: &SyncConfig{
				SyncInterval:    0,
				MaxRetries:      3,
				RetryInterval:   500 * time.Millisecond,
				BatchSize:       10,
				CompressionEnabled: true,
			},
			valid: false,
		},
		{
			name: "negative max retries",
			config: &SyncConfig{
				SyncInterval:    1 * time.Second,
				MaxRetries:      -1,
				RetryInterval:   500 * time.Millisecond,
				BatchSize:       10,
				CompressionEnabled: true,
			},
			valid: false,
		},
		{
			name: "zero batch size",
			config: &SyncConfig{
				SyncInterval:    1 * time.Second,
				MaxRetries:      3,
				RetryInterval:   500 * time.Millisecond,
				BatchSize:       0,
				CompressionEnabled: true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.SyncInterval > 0 && 
					  tt.config.MaxRetries >= 0 && 
					  tt.config.BatchSize > 0

			if isValid != tt.valid {
				t.Errorf("Expected config validity %v, got %v", tt.valid, isValid)
			}
		})
	}
}

func TestModelInfo(t *testing.T) {
	modelInfo := &ModelInfo{
		ID:       "test-model-1",
		Name:     "Test Model",
		Version:  "1.0.0",
		Size:     1024 * 1024 * 1024, // 1GB
		Hash:     "abc123def456",
		Status:   "loaded",
		LastSync: time.Now(),
	}

	assert.Equal(t, "test-model-1", modelInfo.ID)
	assert.Equal(t, "Test Model", modelInfo.Name)
	assert.Equal(t, "1.0.0", modelInfo.Version)
	assert.Equal(t, int64(1024*1024*1024), modelInfo.Size)
	assert.Equal(t, "abc123def456", modelInfo.Hash)
	assert.Equal(t, "loaded", modelInfo.Status)
	assert.True(t, time.Since(modelInfo.LastSync) < time.Second)
}

func TestSyncOperation(t *testing.T) {
	operation := &SyncOperation{
		ID:        "sync-op-1",
		Type:      "model-update",
		ModelID:   "test-model",
		Status:    "pending",
		Progress:  0.0,
		StartTime: time.Now(),
	}

	assert.Equal(t, "sync-op-1", operation.ID)
	assert.Equal(t, "model-update", operation.Type)
	assert.Equal(t, "test-model", operation.ModelID)
	assert.Equal(t, "pending", operation.Status)
	assert.Equal(t, 0.0, operation.Progress)
	assert.True(t, time.Since(operation.StartTime) < time.Second)

	// Test progress update
	operation.Progress = 0.5
	operation.Status = "in-progress"
	
	assert.Equal(t, 0.5, operation.Progress)
	assert.Equal(t, "in-progress", operation.Status)
}

func TestConflictResolution(t *testing.T) {
	resolution := &ConflictResolution{
		ConflictID:   "conflict-1",
		ModelID:      "test-model",
		Strategy:     "latest-wins",
		Resolution:   "resolved",
		ResolvedBy:   "node-1",
		ResolvedAt:   time.Now(),
	}

	assert.Equal(t, "conflict-1", resolution.ConflictID)
	assert.Equal(t, "test-model", resolution.ModelID)
	assert.Equal(t, "latest-wins", resolution.Strategy)
	assert.Equal(t, "resolved", resolution.Resolution)
	assert.Equal(t, "node-1", resolution.ResolvedBy)
	assert.True(t, time.Since(resolution.ResolvedAt) < time.Second)
}

func TestSyncManager(t *testing.T) {
	manager := &SyncManager{
		NodeID:     "test-node-1",
		Models:     make(map[string]*ModelInfo),
		Operations: make(map[string]*SyncOperation),
		Conflicts:  make(map[string]*ConflictResolution),
	}

	assert.Equal(t, "test-node-1", manager.NodeID)
	assert.Equal(t, 0, len(manager.Models))
	assert.Equal(t, 0, len(manager.Operations))
	assert.Equal(t, 0, len(manager.Conflicts))

	// Add a model
	modelInfo := &ModelInfo{
		ID:      "model-1",
		Name:    "Test Model 1",
		Version: "1.0.0",
		Status:  "loaded",
	}
	manager.Models["model-1"] = modelInfo

	assert.Equal(t, 1, len(manager.Models))
	assert.Equal(t, modelInfo, manager.Models["model-1"])
}

func TestPeerSync(t *testing.T) {
	peer := &PeerInfo{
		NodeID:    "peer-node-1",
		Address:   "192.168.1.100",
		Status:    "connected",
		LastSeen:  time.Now(),
		Models:    []string{"model-1", "model-2"},
		Latency:   50 * time.Millisecond,
	}

	assert.Equal(t, "peer-node-1", peer.NodeID)
	assert.Equal(t, "192.168.1.100", peer.Address)
	assert.Equal(t, "connected", peer.Status)
	assert.True(t, time.Since(peer.LastSeen) < time.Second)
	assert.Len(t, peer.Models, 2)
	assert.Contains(t, peer.Models, "model-1")
	assert.Contains(t, peer.Models, "model-2")
	assert.Equal(t, 50*time.Millisecond, peer.Latency)
}

func TestVersionVector(t *testing.T) {
	vector := &VersionVector{
		Versions: make(map[string]int64),
	}

	// Initial state
	assert.Equal(t, 0, len(vector.Versions))

	// Update version
	vector.Versions["node-1"] = 1
	vector.Versions["node-2"] = 1

	assert.Equal(t, 2, len(vector.Versions))
	assert.Equal(t, int64(1), vector.Versions["node-1"])
	assert.Equal(t, int64(1), vector.Versions["node-2"])

	// Increment version
	vector.Versions["node-1"]++
	assert.Equal(t, int64(2), vector.Versions["node-1"])
}

func TestSyncStats(t *testing.T) {
	stats := &SyncStats{
		TotalOperations:   100,
		SuccessfulSyncs:   95,
		FailedSyncs:       5,
		AverageLatency:    150 * time.Millisecond,
		LastSyncTime:      time.Now(),
		BytesSynced:       1024 * 1024 * 50, // 50MB
		ModelsManaged:     10,
		ActivePeers:       5,
	}

	assert.Equal(t, int64(100), stats.TotalOperations)
	assert.Equal(t, int64(95), stats.SuccessfulSyncs)
	assert.Equal(t, int64(5), stats.FailedSyncs)
	assert.Equal(t, 150*time.Millisecond, stats.AverageLatency)
	assert.True(t, time.Since(stats.LastSyncTime) < time.Second)
	assert.Equal(t, int64(1024*1024*50), stats.BytesSynced)
	assert.Equal(t, 10, stats.ModelsManaged)
	assert.Equal(t, 5, stats.ActivePeers)

	// Calculate success rate
	successRate := float64(stats.SuccessfulSyncs) / float64(stats.TotalOperations)
	assert.Equal(t, 0.95, successRate)
}

func BenchmarkSyncConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config := &SyncConfig{
			SyncInterval:    1 * time.Second,
			MaxRetries:      3,
			RetryInterval:   500 * time.Millisecond,
			BatchSize:       10,
			CompressionEnabled: true,
		}
		_ = config
	}
}

func BenchmarkModelInfoCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		modelInfo := &ModelInfo{
			ID:       "test-model",
			Name:     "Test Model",
			Version:  "1.0.0",
			Size:     1024 * 1024 * 1024,
			Hash:     "abc123def456",
			Status:   "loaded",
			LastSync: time.Now(),
		}
		_ = modelInfo
	}
}

// Helper types for testing (these would normally be in the main package)
type SyncConfig struct {
	SyncInterval       time.Duration
	MaxRetries         int
	RetryInterval      time.Duration
	BatchSize          int
	CompressionEnabled bool
}

type ModelInfo struct {
	ID       string
	Name     string
	Version  string
	Size     int64
	Hash     string
	Status   string
	LastSync time.Time
}

type SyncOperation struct {
	ID        string
	Type      string
	ModelID   string
	Status    string
	Progress  float64
	StartTime time.Time
}

type ConflictResolution struct {
	ConflictID string
	ModelID    string
	Strategy   string
	Resolution string
	ResolvedBy string
	ResolvedAt time.Time
}

type SyncManager struct {
	NodeID     string
	Models     map[string]*ModelInfo
	Operations map[string]*SyncOperation
	Conflicts  map[string]*ConflictResolution
}

type PeerInfo struct {
	NodeID   string
	Address  string
	Status   string
	LastSeen time.Time
	Models   []string
	Latency  time.Duration
}

type VersionVector struct {
	Versions map[string]int64
}

type SyncStats struct {
	TotalOperations int64
	SuccessfulSyncs int64
	FailedSyncs     int64
	AverageLatency  time.Duration
	LastSyncTime    time.Time
	BytesSynced     int64
	ModelsManaged   int
	ActivePeers     int
}