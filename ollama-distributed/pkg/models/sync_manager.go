package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// SyncManager manages model synchronization across the distributed network
type SyncManager struct {
	config      *config.SyncConfig
	p2p         *p2p.Node
	manager     *Manager
	logger      *slog.Logger
	
	// Sync state
	syncStates  map[string]*SyncState
	syncMutex   sync.RWMutex
	
	// Version tracking
	modelVersions map[string]*ModelVersion
	versionMutex  sync.RWMutex
	
	// Sync workers
	syncWorkers    []*SyncWorker
	syncQueue      chan *SyncRequest
	
	// Delta tracking
	deltaTracker   *DeltaTracker
	
	// Content-addressed storage
	casStore       *ContentAddressedStore
	
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// SyncState represents the synchronization state of a model
type SyncState struct {
	ModelName      string                 `json:"model_name"`
	LocalVersion   string                 `json:"local_version"`
	RemoteVersions map[string]string      `json:"remote_versions"` // peerID -> version
	Status         SyncStatus             `json:"status"`
	LastSyncTime   time.Time              `json:"last_sync_time"`
	PendingDeltas  []string               `json:"pending_deltas"`
	Conflicts      []SyncConflict         `json:"conflicts"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// SyncStatus represents the status of model synchronization
type SyncStatus string

const (
	SyncStatusInSync     SyncStatus = "in_sync"
	SyncStatusOutOfSync  SyncStatus = "out_of_sync"
	SyncStatusSyncing    SyncStatus = "syncing"
	SyncStatusConflict   SyncStatus = "conflict"
	SyncStatusError      SyncStatus = "error"
)

// SyncConflict represents a synchronization conflict
type SyncConflict struct {
	Type        ConflictType `json:"type"`
	LocalHash   string       `json:"local_hash"`
	RemoteHash  string       `json:"remote_hash"`
	PeerID      string       `json:"peer_id"`
	Description string       `json:"description"`
	Timestamp   time.Time    `json:"timestamp"`
}

// ConflictType represents the type of synchronization conflict
type ConflictType string

const (
	ConflictTypeVersion    ConflictType = "version"
	ConflictTypeChecksum   ConflictType = "checksum"
	ConflictTypeTimestamp  ConflictType = "timestamp"
	ConflictTypeMetadata   ConflictType = "metadata"
)

// ModelVersion represents a versioned model with content addressing
type ModelVersion struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Hash        string            `json:"hash"`        // Content-addressed hash
	ParentHash  string            `json:"parent_hash"` // Previous version hash
	Size        int64             `json:"size"`
	Chunks      []ChunkInfo       `json:"chunks"`
	Metadata    map[string]string `json:"metadata"`
	Timestamp   time.Time         `json:"timestamp"`
	Author      string            `json:"author"`      // Node ID that created this version
	Signature   string            `json:"signature"`   // Cryptographic signature
}

// ChunkInfo represents a chunk of a model file
type ChunkInfo struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
}

// SyncRequest represents a request to synchronize a model
type SyncRequest struct {
	ModelName    string      `json:"model_name"`
	PeerID       string      `json:"peer_id"`
	SyncType     SyncType    `json:"sync_type"`
	Priority     int         `json:"priority"`
	ResponseChan chan error  `json:"-"`
}

// SyncType represents the type of synchronization
type SyncType string

const (
	SyncTypeFull        SyncType = "full"
	SyncTypeIncremental SyncType = "incremental"
	SyncTypeDelta       SyncType = "delta"
)

// SyncWorker handles synchronization operations
type SyncWorker struct {
	ID         int
	manager    *SyncManager
	stopChan   chan struct{}
}

// NewSyncManager creates a new model synchronization manager
func NewSyncManager(config *config.SyncConfig, p2pNode *p2p.Node, manager *Manager, logger *slog.Logger) (*SyncManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	sm := &SyncManager{
		config:        config,
		p2p:           p2pNode,
		manager:       manager,
		logger:        logger,
		syncStates:    make(map[string]*SyncState),
		modelVersions: make(map[string]*ModelVersion),
		syncQueue:     make(chan *SyncRequest, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
	
	// Initialize delta tracker
	deltaTracker, err := NewDeltaTracker(config.DeltaDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create delta tracker: %w", err)
	}
	sm.deltaTracker = deltaTracker
	
	// Initialize content-addressed store
	casStore, err := NewContentAddressedStore(config.CASDir, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CAS store: %w", err)
	}
	sm.casStore = casStore
	
	// Create sync workers
	sm.syncWorkers = make([]*SyncWorker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		sm.syncWorkers[i] = &SyncWorker{
			ID:       i,
			manager:  sm,
			stopChan: make(chan struct{}),
		}
	}
	
	return sm, nil
}

// Start starts the synchronization manager
func (sm *SyncManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.started {
		return fmt.Errorf("sync manager already started")
	}
	
	// Create necessary directories
	if err := os.MkdirAll(sm.config.DeltaDir, 0755); err != nil {
		return fmt.Errorf("failed to create delta directory: %w", err)
	}
	
	if err := os.MkdirAll(sm.config.CASDir, 0755); err != nil {
		return fmt.Errorf("failed to create CAS directory: %w", err)
	}
	
	// Load existing sync states
	if err := sm.loadSyncStates(); err != nil {
		return fmt.Errorf("failed to load sync states: %w", err)
	}
	
	// Start sync workers
	for _, worker := range sm.syncWorkers {
		go worker.start()
	}
	
	// Start periodic sync routine
	go sm.periodicSyncRoutine()
	
	// Start version tracking routine
	go sm.versionTrackingRoutine()
	
	sm.started = true
	sm.logger.Info("sync manager started", "workers", len(sm.syncWorkers))
	
	return nil
}

// SynchronizeModel synchronizes a model with a specific peer
func (sm *SyncManager) SynchronizeModel(modelName, peerID string, syncType SyncType) error {
	req := &SyncRequest{
		ModelName:    modelName,
		PeerID:       peerID,
		SyncType:     syncType,
		Priority:     1,
		ResponseChan: make(chan error, 1),
	}
	
	select {
	case sm.syncQueue <- req:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("sync queue full")
	}
	
	select {
	case err := <-req.ResponseChan:
		return err
	case <-time.After(10 * time.Minute):
		return fmt.Errorf("sync timeout")
	}
}

// GetSyncState returns the synchronization state of a model
func (sm *SyncManager) GetSyncState(modelName string) (*SyncState, bool) {
	sm.syncMutex.RLock()
	defer sm.syncMutex.RUnlock()
	
	state, exists := sm.syncStates[modelName]
	return state, exists
}

// GetModelVersion returns the version information for a model
func (sm *SyncManager) GetModelVersion(modelName string) (*ModelVersion, bool) {
	sm.versionMutex.RLock()
	defer sm.versionMutex.RUnlock()
	
	version, exists := sm.modelVersions[modelName]
	return version, exists
}

// CreateModelVersion creates a new version for a model
func (sm *SyncManager) CreateModelVersion(modelName, modelPath string) (*ModelVersion, error) {
	// Calculate content hash
	hash, err := sm.calculateModelHash(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate model hash: %w", err)
	}
	
	// Get file info
	info, err := os.Stat(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat model file: %w", err)
	}
	
	// Create chunks
	chunks, err := sm.createChunks(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunks: %w", err)
	}
	
	// Get previous version if exists
	var parentHash string
	if prevVersion, exists := sm.GetModelVersion(modelName); exists {
		parentHash = prevVersion.Hash
	}
	
	// Create version
	version := &ModelVersion{
		Name:       modelName,
		Version:    sm.generateVersion(),
		Hash:       hash,
		ParentHash: parentHash,
		Size:       info.Size(),
		Chunks:     chunks,
		Metadata:   make(map[string]string),
		Timestamp:  time.Now(),
		Author:     sm.p2p.ID().String(),
	}
	
	// Store in content-addressed store
	if err := sm.casStore.Store(hash, modelPath); err != nil {
		return nil, fmt.Errorf("failed to store in CAS: %w", err)
	}
	
	// Update version tracking
	sm.versionMutex.Lock()
	sm.modelVersions[modelName] = version
	sm.versionMutex.Unlock()
	
	sm.logger.Info("created model version", "model", modelName, "version", version.Version, "hash", hash)
	
	return version, nil
}

// loadSyncStates loads existing synchronization states
func (sm *SyncManager) loadSyncStates() error {
	stateFile := filepath.Join(sm.config.DeltaDir, "sync_states.json")
	
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return nil // No existing states
	}
	
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read sync states: %w", err)
	}
	
	var states map[string]*SyncState
	if err := json.Unmarshal(data, &states); err != nil {
		return fmt.Errorf("failed to unmarshal sync states: %w", err)
	}
	
	sm.syncMutex.Lock()
	sm.syncStates = states
	sm.syncMutex.Unlock()
	
	return nil
}

// saveSyncStates saves synchronization states to disk
func (sm *SyncManager) saveSyncStates() error {
	sm.syncMutex.RLock()
	defer sm.syncMutex.RUnlock()
	
	stateFile := filepath.Join(sm.config.DeltaDir, "sync_states.json")
	
	data, err := json.MarshalIndent(sm.syncStates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync states: %w", err)
	}
	
	return os.WriteFile(stateFile, data, 0644)
}

// calculateModelHash calculates the content hash of a model file
func (sm *SyncManager) calculateModelHash(modelPath string) (string, error) {
	file, err := os.Open(modelPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createChunks creates chunks for a model file
func (sm *SyncManager) createChunks(modelPath string) ([]ChunkInfo, error) {
	file, err := os.Open(modelPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	chunkSize := sm.config.ChunkSize
	if chunkSize == 0 {
		chunkSize = 1024 * 1024 // Default 1MB chunks
	}
	
	var chunks []ChunkInfo
	buffer := make([]byte, chunkSize)
	offset := int64(0)
	
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		
		if n == 0 {
			break
		}
		
		// Calculate chunk hash
		hash := sha256.Sum256(buffer[:n])
		
		chunks = append(chunks, ChunkInfo{
			Hash:   hex.EncodeToString(hash[:]),
			Size:   int64(n),
			Offset: offset,
		})
		
		offset += int64(n)
	}
	
	return chunks, nil
}

// generateVersion generates a new version string
func (sm *SyncManager) generateVersion() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// periodicSyncRoutine runs periodic synchronization tasks
func (sm *SyncManager) periodicSyncRoutine() {
	ticker := time.NewTicker(sm.config.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.performPeriodicSync()
		}
	}
}

// performPeriodicSync performs periodic synchronization tasks
func (sm *SyncManager) performPeriodicSync() {
	// Save sync states
	if err := sm.saveSyncStates(); err != nil {
		sm.logger.Error("failed to save sync states", "error", err)
	}
	
	// Check for models that need synchronization
	models := sm.manager.GetAllModels()
	for modelName := range models {
		state, exists := sm.GetSyncState(modelName)
		if !exists {
			// Create initial sync state
			state = &SyncState{
				ModelName:      modelName,
				LocalVersion:   "1.0.0",
				RemoteVersions: make(map[string]string),
				Status:         SyncStatusInSync,
				LastSyncTime:   time.Now(),
				PendingDeltas:  []string{},
				Conflicts:      []SyncConflict{},
				Metadata:       make(map[string]interface{}),
			}
			
			sm.syncMutex.Lock()
			sm.syncStates[modelName] = state
			sm.syncMutex.Unlock()
		}
		
		// Check if sync is needed
		if time.Since(state.LastSyncTime) > sm.config.SyncInterval {
			// Queue for synchronization
			go func(name string) {
				sm.SynchronizeModel(name, "", SyncTypeIncremental)
			}(modelName)
		}
	}
}

// versionTrackingRoutine runs version tracking tasks
func (sm *SyncManager) versionTrackingRoutine() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.trackModelVersions()
		}
	}
}

// trackModelVersions tracks changes in model versions
func (sm *SyncManager) trackModelVersions() {
	models := sm.manager.GetAllModels()
	
	for modelName, model := range models {
		// Check if version has changed
		currentVersion, exists := sm.GetModelVersion(modelName)
		if !exists {
			// Create initial version
			if _, err := sm.CreateModelVersion(modelName, model.Path); err != nil {
				sm.logger.Error("failed to create initial version", "model", modelName, "error", err)
			}
			continue
		}
		
		// Calculate current hash
		currentHash, err := sm.calculateModelHash(model.Path)
		if err != nil {
			sm.logger.Error("failed to calculate model hash", "model", modelName, "error", err)
			continue
		}
		
		// Check if model has changed
		if currentHash != currentVersion.Hash {
			// Create new version
			if _, err := sm.CreateModelVersion(modelName, model.Path); err != nil {
				sm.logger.Error("failed to create new version", "model", modelName, "error", err)
			} else {
				sm.logger.Info("model version updated", "model", modelName, "old_hash", currentVersion.Hash, "new_hash", currentHash)
			}
		}
	}
}

// Shutdown gracefully shuts down the synchronization manager
func (sm *SyncManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.started {
		return nil
	}
	
	// Stop workers
	for _, worker := range sm.syncWorkers {
		close(worker.stopChan)
	}
	
	// Save final state
	if err := sm.saveSyncStates(); err != nil {
		sm.logger.Error("failed to save final sync states", "error", err)
	}
	
	// Close delta tracker
	if sm.deltaTracker != nil {
		sm.deltaTracker.Close()
	}
	
	// Close CAS store
	if sm.casStore != nil {
		sm.casStore.Close()
	}
	
	sm.cancel()
	sm.started = false
	
	sm.logger.Info("sync manager shutdown complete")
	return nil
}

// SyncWorker methods

// start starts the sync worker
func (w *SyncWorker) start() {
	w.manager.logger.Info("sync worker started", "worker_id", w.ID)
	
	for {
		select {
		case <-w.stopChan:
			w.manager.logger.Info("sync worker stopped", "worker_id", w.ID)
			return
		case req := <-w.manager.syncQueue:
			w.processSyncRequest(req)
		}
	}
}

// processSyncRequest processes a synchronization request
func (w *SyncWorker) processSyncRequest(req *SyncRequest) {
	w.manager.logger.Info("processing sync request", "worker_id", w.ID, "model", req.ModelName, "peer", req.PeerID, "type", req.SyncType)
	
	var err error
	
	switch req.SyncType {
	case SyncTypeFull:
		err = w.performFullSync(req)
	case SyncTypeIncremental:
		err = w.performIncrementalSync(req)
	case SyncTypeDelta:
		err = w.performDeltaSync(req)
	default:
		err = fmt.Errorf("unknown sync type: %s", req.SyncType)
	}
	
	if err != nil {
		w.manager.logger.Error("sync request failed", "worker_id", w.ID, "model", req.ModelName, "error", err)
	} else {
		w.manager.logger.Info("sync request completed", "worker_id", w.ID, "model", req.ModelName)
	}
	
	// Send response
	select {
	case req.ResponseChan <- err:
	case <-time.After(time.Second):
		// Response channel blocked
	}
}

// performFullSync performs a full synchronization
func (w *SyncWorker) performFullSync(req *SyncRequest) error {
	// TODO: Implement full sync logic
	// This would involve:
	// 1. Getting the complete model from peer
	// 2. Verifying integrity
	// 3. Replacing local model
	// 4. Updating sync state
	
	time.Sleep(100 * time.Millisecond) // Simulate work
	
	// Update sync state
	w.manager.syncMutex.Lock()
	if state, exists := w.manager.syncStates[req.ModelName]; exists {
		state.Status = SyncStatusInSync
		state.LastSyncTime = time.Now()
		if req.PeerID != "" {
			state.RemoteVersions[req.PeerID] = "1.0.0"
		}
	}
	w.manager.syncMutex.Unlock()
	
	return nil
}

// performIncrementalSync performs an incremental synchronization
func (w *SyncWorker) performIncrementalSync(req *SyncRequest) error {
	// TODO: Implement incremental sync logic
	// This would involve:
	// 1. Comparing model versions
	// 2. Identifying differences
	// 3. Downloading only changed parts
	// 4. Applying changes
	// 5. Updating sync state
	
	time.Sleep(50 * time.Millisecond) // Simulate work
	
	// Update sync state
	w.manager.syncMutex.Lock()
	if state, exists := w.manager.syncStates[req.ModelName]; exists {
		state.Status = SyncStatusInSync
		state.LastSyncTime = time.Now()
		if req.PeerID != "" {
			state.RemoteVersions[req.PeerID] = "1.0.1"
		}
	}
	w.manager.syncMutex.Unlock()
	
	return nil
}

// performDeltaSync performs a delta synchronization
func (w *SyncWorker) performDeltaSync(req *SyncRequest) error {
	// TODO: Implement delta sync logic
	// This would involve:
	// 1. Getting delta information from peer
	// 2. Applying deltas to local model
	// 3. Verifying result integrity
	// 4. Updating sync state
	
	time.Sleep(25 * time.Millisecond) // Simulate work
	
	// Update sync state
	w.manager.syncMutex.Lock()
	if state, exists := w.manager.syncStates[req.ModelName]; exists {
		state.Status = SyncStatusInSync
		state.LastSyncTime = time.Now()
		if req.PeerID != "" {
			state.RemoteVersions[req.PeerID] = "1.0.2"
		}
	}
	w.manager.syncMutex.Unlock()
	
	return nil
}