package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

// StateSynchronizer manages efficient state synchronization between nodes
type StateSynchronizer struct {
	engine          *Engine
	mu              sync.RWMutex
	
	// State tracking
	stateVersions   map[string]*StateVersion
	syncRequests    map[string]*SyncRequest
	
	// Configuration
	config          *SyncConfig
	
	// Metrics
	metrics         *SyncMetrics
	
	// Lifecycle
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// StateVersion represents a versioned state snapshot
type StateVersion struct {
	Version         int64             `json:"version"`
	Hash            string            `json:"hash"`
	Timestamp       time.Time         `json:"timestamp"`
	Size            int64             `json:"size"`
	Checksum        string            `json:"checksum"`
	Metadata        map[string]interface{} `json:"metadata"`
	
	// Delta information
	PreviousVersion int64             `json:"previous_version,omitempty"`
	DeltaSize       int64             `json:"delta_size,omitempty"`
	Changes         []*StateChange    `json:"changes,omitempty"`
}

// StateChange represents a single state change
type StateChange struct {
	Type            ChangeType        `json:"type"`
	Key             string            `json:"key"`
	OldValue        interface{}       `json:"old_value,omitempty"`
	NewValue        interface{}       `json:"new_value,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// ChangeType represents the type of state change
type ChangeType string

const (
	ChangeTypeSet    ChangeType = "set"
	ChangeTypeDelete ChangeType = "delete"
	ChangeTypeUpdate ChangeType = "update"
)

// SyncRequest represents a state synchronization request
type SyncRequest struct {
	RequestID       string            `json:"request_id"`
	FromNode        raft.ServerID     `json:"from_node"`
	ToNode          raft.ServerID     `json:"to_node"`
	FromVersion     int64             `json:"from_version"`
	ToVersion       int64             `json:"to_version"`
	SyncType        SyncType          `json:"sync_type"`
	Priority        SyncPriority      `json:"priority"`
	StartTime       time.Time         `json:"start_time"`
	Status          SyncStatus        `json:"status"`
	Progress        float64           `json:"progress"`
	Error           string            `json:"error,omitempty"`
}

// SyncType represents the type of synchronization
type SyncType string

const (
	SyncTypeFull        SyncType = "full"
	SyncTypeDelta       SyncType = "delta"
	SyncTypeIncremental SyncType = "incremental"
)

// SyncPriority represents the priority of synchronization
type SyncPriority string

const (
	SyncPriorityLow    SyncPriority = "low"
	SyncPriorityNormal SyncPriority = "normal"
	SyncPriorityHigh   SyncPriority = "high"
	SyncPriorityUrgent SyncPriority = "urgent"
)

// SyncStatus represents the status of synchronization
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusInProgress SyncStatus = "in_progress"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusCancelled  SyncStatus = "cancelled"
)

// SyncConfig configures state synchronization
type SyncConfig struct {
	// Delta sync settings
	EnableDeltaSync     bool
	MaxDeltaSize        int64
	DeltaCompressionThreshold int64
	
	// Batch settings
	BatchSize           int
	MaxConcurrentSyncs  int
	SyncTimeout         time.Duration
	
	// Optimization settings
	EnableCompression   bool
	EnableChecksums     bool
	EnableDeduplication bool
	
	// Performance settings
	SyncInterval        time.Duration
	CleanupInterval     time.Duration
	MaxVersionHistory   int
}

// SyncMetrics tracks synchronization performance
type SyncMetrics struct {
	TotalSyncs          int64         `json:"total_syncs"`
	SuccessfulSyncs     int64         `json:"successful_syncs"`
	FailedSyncs         int64         `json:"failed_syncs"`
	AverageSyncTime     time.Duration `json:"average_sync_time"`
	TotalBytesTransferred int64       `json:"total_bytes_transferred"`
	DeltaSyncs          int64         `json:"delta_syncs"`
	FullSyncs           int64         `json:"full_syncs"`
	LastSync            time.Time     `json:"last_sync"`
}

// NewStateSynchronizer creates a new state synchronizer
func NewStateSynchronizer(engine *Engine, config *SyncConfig) *StateSynchronizer {
	ctx, cancel := context.WithCancel(context.Background())
	
	if config == nil {
		config = &SyncConfig{
			EnableDeltaSync:           true,
			MaxDeltaSize:              10 * 1024 * 1024, // 10MB
			DeltaCompressionThreshold: 1024 * 1024,      // 1MB
			BatchSize:                 100,
			MaxConcurrentSyncs:        5,
			SyncTimeout:               30 * time.Second,
			EnableCompression:         true,
			EnableChecksums:           true,
			EnableDeduplication:       true,
			SyncInterval:              10 * time.Second,
			CleanupInterval:           5 * time.Minute,
			MaxVersionHistory:         100,
		}
	}
	
	ss := &StateSynchronizer{
		engine:        engine,
		stateVersions: make(map[string]*StateVersion),
		syncRequests:  make(map[string]*SyncRequest),
		config:        config,
		metrics:       &SyncMetrics{},
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Start background tasks
	ss.wg.Add(2)
	go ss.syncLoop()
	go ss.cleanupLoop()
	
	return ss
}

// CreateStateVersion creates a new state version snapshot
func (ss *StateSynchronizer) CreateStateVersion() (*StateVersion, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	// Get current state
	state := ss.engine.GetState()
	
	// Serialize state
	stateData, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize state: %w", err)
	}
	
	// Calculate hash and checksum
	hash := sha256.Sum256(stateData)
	hashStr := hex.EncodeToString(hash[:])
	
	// Create version
	version := &StateVersion{
		Version:   time.Now().UnixNano(),
		Hash:      hashStr,
		Timestamp: time.Now(),
		Size:      int64(len(stateData)),
		Checksum:  hashStr, // Using same as hash for simplicity
		Metadata: map[string]interface{}{
			"node_id": ss.engine.GetNodeID(),
			"keys":    len(state),
		},
	}
	
	// Store version
	versionKey := fmt.Sprintf("version_%d", version.Version)
	ss.stateVersions[versionKey] = version
	
	// Cleanup old versions
	ss.cleanupOldVersions()
	
	return version, nil
}

// SyncWithPeer synchronizes state with a specific peer
func (ss *StateSynchronizer) SyncWithPeer(peerID raft.ServerID, syncType SyncType, priority SyncPriority) (*SyncRequest, error) {
	requestID := fmt.Sprintf("sync_%d_%s", time.Now().UnixNano(), peerID)
	
	request := &SyncRequest{
		RequestID:   requestID,
		FromNode:    raft.ServerID(ss.engine.GetNodeID()),
		ToNode:      peerID,
		SyncType:    syncType,
		Priority:    priority,
		StartTime:   time.Now(),
		Status:      SyncStatusPending,
		Progress:    0.0,
	}
	
	ss.mu.Lock()
	ss.syncRequests[requestID] = request
	ss.mu.Unlock()
	
	// Start sync in background
	go ss.performSync(request)
	
	return request, nil
}

// performSync performs the actual synchronization
func (ss *StateSynchronizer) performSync(request *SyncRequest) {
	ss.updateSyncStatus(request.RequestID, SyncStatusInProgress, 0.0, "")
	
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		ss.updateSyncMetrics(request, duration)
	}()
	
	switch request.SyncType {
	case SyncTypeFull:
		ss.performFullSync(request)
	case SyncTypeDelta:
		ss.performDeltaSync(request)
	case SyncTypeIncremental:
		ss.performIncrementalSync(request)
	default:
		ss.updateSyncStatus(request.RequestID, SyncStatusFailed, 0.0, "unknown sync type")
	}
}

// performFullSync performs a full state synchronization
func (ss *StateSynchronizer) performFullSync(request *SyncRequest) {
	// Create current state version
	version, err := ss.CreateStateVersion()
	if err != nil {
		ss.updateSyncStatus(request.RequestID, SyncStatusFailed, 0.0, err.Error())
		return
	}
	
	// In a real implementation, you would:
	// 1. Send the full state to the peer
	// 2. Wait for acknowledgment
	// 3. Handle any errors or retries
	
	// For now, simulate the sync
	ss.updateSyncStatus(request.RequestID, SyncStatusInProgress, 0.5, "transferring state")
	time.Sleep(100 * time.Millisecond) // Simulate network delay
	ss.updateSyncStatus(request.RequestID, SyncStatusCompleted, 1.0, "")
	
	ss.metrics.FullSyncs++
	ss.metrics.TotalBytesTransferred += version.Size
}

// performDeltaSync performs a delta state synchronization
func (ss *StateSynchronizer) performDeltaSync(request *SyncRequest) {
	// Calculate delta between versions
	changes, err := ss.calculateStateDelta(request.FromVersion, request.ToVersion)
	if err != nil {
		ss.updateSyncStatus(request.RequestID, SyncStatusFailed, 0.0, err.Error())
		return
	}
	
	// Check if delta is too large
	deltaSize := ss.estimateDeltaSize(changes)
	if deltaSize > ss.config.MaxDeltaSize {
		// Fall back to full sync
		request.SyncType = SyncTypeFull
		ss.performFullSync(request)
		return
	}
	
	// Send delta
	ss.updateSyncStatus(request.RequestID, SyncStatusInProgress, 0.5, "transferring delta")
	time.Sleep(50 * time.Millisecond) // Simulate network delay
	ss.updateSyncStatus(request.RequestID, SyncStatusCompleted, 1.0, "")
	
	ss.metrics.DeltaSyncs++
	ss.metrics.TotalBytesTransferred += deltaSize
}

// performIncrementalSync performs an incremental state synchronization
func (ss *StateSynchronizer) performIncrementalSync(request *SyncRequest) {
	// Similar to delta sync but with smaller increments
	ss.performDeltaSync(request)
}

// calculateStateDelta calculates the delta between two state versions
func (ss *StateSynchronizer) calculateStateDelta(fromVersion, toVersion int64) ([]*StateChange, error) {
	// In a real implementation, you would:
	// 1. Load the two state versions
	// 2. Compare them to find differences
	// 3. Generate a list of changes
	
	// For now, return empty changes
	return []*StateChange{}, nil
}

// estimateDeltaSize estimates the size of a delta
func (ss *StateSynchronizer) estimateDeltaSize(changes []*StateChange) int64 {
	size := int64(0)
	for _, change := range changes {
		// Rough estimation based on JSON serialization
		changeData, _ := json.Marshal(change)
		size += int64(len(changeData))
	}
	return size
}

// updateSyncStatus updates the status of a sync request
func (ss *StateSynchronizer) updateSyncStatus(requestID string, status SyncStatus, progress float64, errorMsg string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	if request, exists := ss.syncRequests[requestID]; exists {
		request.Status = status
		request.Progress = progress
		request.Error = errorMsg
	}
}

// updateSyncMetrics updates synchronization metrics
func (ss *StateSynchronizer) updateSyncMetrics(request *SyncRequest, duration time.Duration) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	ss.metrics.TotalSyncs++
	if request.Status == SyncStatusCompleted {
		ss.metrics.SuccessfulSyncs++
	} else {
		ss.metrics.FailedSyncs++
	}
	
	// Update average sync time
	totalTime := time.Duration(ss.metrics.TotalSyncs-1) * ss.metrics.AverageSyncTime + duration
	ss.metrics.AverageSyncTime = totalTime / time.Duration(ss.metrics.TotalSyncs)
	
	ss.metrics.LastSync = time.Now()
}

// GetSyncRequest returns a sync request by ID
func (ss *StateSynchronizer) GetSyncRequest(requestID string) *SyncRequest {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	
	if request, exists := ss.syncRequests[requestID]; exists {
		// Return a copy
		requestCopy := *request
		return &requestCopy
	}
	return nil
}

// GetSyncMetrics returns synchronization metrics
func (ss *StateSynchronizer) GetSyncMetrics() *SyncMetrics {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	
	metrics := *ss.metrics
	return &metrics
}

// cleanupOldVersions removes old state versions
func (ss *StateSynchronizer) cleanupOldVersions() {
	if len(ss.stateVersions) <= ss.config.MaxVersionHistory {
		return
	}
	
	// Convert to slice and sort by version
	versions := make([]*StateVersion, 0, len(ss.stateVersions))
	for _, version := range ss.stateVersions {
		versions = append(versions, version)
	}
	
	// Sort by version (oldest first)
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].Version > versions[j].Version {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}
	
	// Remove oldest versions
	toRemove := len(versions) - ss.config.MaxVersionHistory
	for i := 0; i < toRemove; i++ {
		versionKey := fmt.Sprintf("version_%d", versions[i].Version)
		delete(ss.stateVersions, versionKey)
	}
}

// syncLoop periodically performs synchronization tasks
func (ss *StateSynchronizer) syncLoop() {
	defer ss.wg.Done()
	
	ticker := time.NewTicker(ss.config.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ss.ctx.Done():
			return
		case <-ticker.C:
			ss.performPeriodicSync()
		}
	}
}

// performPeriodicSync performs periodic synchronization tasks
func (ss *StateSynchronizer) performPeriodicSync() {
	// Create periodic state version
	ss.CreateStateVersion()
	
	// In a real implementation, you would also:
	// 1. Check for peers that need synchronization
	// 2. Initiate sync requests as needed
	// 3. Monitor sync progress
}

// cleanupLoop periodically cleans up old data
func (ss *StateSynchronizer) cleanupLoop() {
	defer ss.wg.Done()
	
	ticker := time.NewTicker(ss.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ss.ctx.Done():
			return
		case <-ticker.C:
			ss.performCleanup()
		}
	}
}

// performCleanup performs cleanup tasks
func (ss *StateSynchronizer) performCleanup() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	
	// Clean up completed sync requests older than 1 hour
	cutoff := time.Now().Add(-time.Hour)
	for requestID, request := range ss.syncRequests {
		if request.StartTime.Before(cutoff) && 
		   (request.Status == SyncStatusCompleted || request.Status == SyncStatusFailed) {
			delete(ss.syncRequests, requestID)
		}
	}
	
	// Clean up old state versions
	ss.cleanupOldVersions()
}

// Close closes the state synchronizer
func (ss *StateSynchronizer) Close() error {
	ss.cancel()
	ss.wg.Wait()
	return nil
}
