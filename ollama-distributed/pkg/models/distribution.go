package models

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/errors"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/logging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Manager manages model distribution across the network
type Manager struct {
	config *config.StorageConfig
	p2p    *p2p.Node

	// Local model storage
	models   map[string]*Model
	modelsMu sync.RWMutex

	// Transfer management
	transfers   map[string]*Transfer
	transfersMu sync.RWMutex

	// Download queue
	downloadQueue chan *DownloadRequest
	uploadQueue   chan *UploadRequest

	// Workers
	downloadWorkers []*DownloadWorker
	uploadWorkers   []*UploadWorker

	// Advanced transfer components
	p2pEngine           *P2PTransferEngine
	verifier            *IntegrityVerifier
	versionManager      *VersionManager
	replicationManager  *ReplicationManager
	advancedReplication *AdvancedReplicationManager
	lifecycleManager    *LifecycleManager
	advancedCAS         *AdvancedCAS
	syncEngine          *SyncEngine

	// Observability components
	logger           *logging.StructuredLogger
	errorHandler     *errors.ErrorHandler
	metricsCollector *observability.MetricsCollector
	tracer           *observability.Tracer

	started bool
	mu      sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// Model represents a model in the distributed system
type Model struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Size         int64             `json:"size"`
	Checksum     string            `json:"checksum"`
	Path         string            `json:"path"`
	Status       ModelStatus       `json:"status"`
	Replicas     []string          `json:"replicas"` // Node IDs that have this model
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	AccessCount  int64             `json:"access_count"`
	LastAccessed time.Time         `json:"last_accessed"`
}

// ModelStatus represents the status of a model
type ModelStatus string

const (
	ModelStatusDownloading ModelStatus = "downloading"
	ModelStatusAvailable   ModelStatus = "available"
	ModelStatusCorrupted   ModelStatus = "corrupted"
	ModelStatusDeleted     ModelStatus = "deleted"
)

// Transfer represents a model transfer operation
type Transfer struct {
	ID          string         `json:"id"`
	ModelName   string         `json:"model_name"`
	Type        TransferType   `json:"type"`
	Status      TransferStatus `json:"status"`
	Progress    float64        `json:"progress"`
	BytesTotal  int64          `json:"bytes_total"`
	BytesDone   int64          `json:"bytes_done"`
	Speed       int64          `json:"speed"` // bytes per second
	PeerID      string         `json:"peer_id"`
	StartedAt   time.Time      `json:"started_at"`
	CompletedAt time.Time      `json:"completed_at"`
	Error       string         `json:"error,omitempty"`

	// Internal fields
	ctx    context.Context
	cancel context.CancelFunc
}

// TransferType represents the type of transfer
type TransferType string

const (
	TransferTypeDownload TransferType = "download"
	TransferTypeUpload   TransferType = "upload"
)

// Note: TransferStatus and related constants are now defined in p2p_transfer.go

// DownloadRequest represents a request to download a model
type DownloadRequest struct {
	ModelName  string
	PeerID     string
	Priority   int
	ResponseCh chan *DownloadResponse
}

// DownloadResponse represents a response to a download request
type DownloadResponse struct {
	Success  bool
	Model    *Model
	Error    string
	Duration time.Duration
}

// UploadRequest represents a request to upload a model
type UploadRequest struct {
	ModelName  string
	PeerID     string
	ResponseCh chan *UploadResponse
}

// UploadResponse represents a response to an upload request
type UploadResponse struct {
	Success  bool
	Error    string
	Duration time.Duration
}

// DownloadWorker handles download operations
type DownloadWorker struct {
	ID      int
	manager *Manager
	stopCh  chan struct{}
}

// UploadWorker handles upload operations
type UploadWorker struct {
	ID      int
	manager *Manager
	stopCh  chan struct{}
}

// NewManager creates a new model distribution manager
func NewManager(config *config.StorageConfig, p2pNode *p2p.Node) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:        config,
		p2p:           p2pNode,
		models:        make(map[string]*Model),
		transfers:     make(map[string]*Transfer),
		downloadQueue: make(chan *DownloadRequest, 100),
		uploadQueue:   make(chan *UploadRequest, 100),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Initialize P2P transfer engine
	manager.p2pEngine = NewP2PTransferEngine(nil)

	// Initialize integrity verifier
	manager.verifier = NewIntegrityVerifier(nil)

	// Initialize version manager
	manager.versionManager = NewVersionManager(nil)

	// Initialize base replication manager
	// Note: This would normally be initialized with proper config and dependencies
	// For now, we'll initialize the advanced replication manager without the base manager
	manager.advancedReplication = NewAdvancedReplicationManager(nil, nil)

	// Initialize lifecycle manager
	manager.lifecycleManager = NewLifecycleManager(manager.versionManager, manager.advancedReplication, nil)

	// Initialize advanced CAS
	advancedCAS, err := NewAdvancedCAS(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize advanced CAS: %w", err)
	}
	manager.advancedCAS = advancedCAS

	// Initialize sync engine
	manager.syncEngine = NewSyncEngine(manager.versionManager, manager.advancedReplication, nil)

	// Initialize observability components
	logger, err := logging.NewStructuredLogger(&logging.LoggerConfig{
		Level:            logging.LevelInfo,
		Format:           logging.FormatJSON,
		EnableStructured: true,
		EnableCaller:     true,
		ServiceName:      "ollama-distributed",
		ServiceVersion:   "1.0.0",
		Environment:      "development",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	manager.logger = logger

	// Initialize error handler
	manager.errorHandler = errors.NewErrorHandler(&errors.ErrorHandlerConfig{
		EnableStackTrace:   true,
		EnableReporting:    true,
		ReportingThreshold: errors.SeverityHigh,
	})

	// Initialize metrics collector
	manager.metricsCollector = observability.NewMetricsCollector(&observability.MetricsConfig{
		Namespace:    "ollama_distributed",
		Subsystem:    "models",
		EnableExport: false,
	})

	// Initialize tracer
	manager.tracer = observability.NewTracer(&observability.TracerConfig{
		ServiceName:    "ollama-distributed",
		ServiceVersion: "1.0.0",
		SamplingRate:   1.0,
		EnableExport:   false,
	})

	// Create download workers
	manager.downloadWorkers = make([]*DownloadWorker, 3)
	for i := 0; i < 3; i++ {
		manager.downloadWorkers[i] = &DownloadWorker{
			ID:      i,
			manager: manager,
			stopCh:  make(chan struct{}),
		}
	}

	// Create upload workers
	manager.uploadWorkers = make([]*UploadWorker, 3)
	for i := 0; i < 3; i++ {
		manager.uploadWorkers[i] = &UploadWorker{
			ID:      i,
			manager: manager,
			stopCh:  make(chan struct{}),
		}
	}

	return manager, nil
}

// Start starts the model distribution manager
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("model manager already started")
	}

	// Create storage directories
	if err := os.MkdirAll(m.config.ModelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	// Load existing models
	if err := m.loadModels(); err != nil {
		return fmt.Errorf("failed to load models: %w", err)
	}

	// Start workers
	for _, worker := range m.downloadWorkers {
		go worker.start()
	}

	for _, worker := range m.uploadWorkers {
		go worker.start()
	}

	// Start cleanup routine
	go m.cleanupRoutine()

	m.started = true
	return nil
}

// loadModels loads existing models from disk
func (m *Manager) loadModels() error {
	return filepath.Walk(m.config.ModelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's a model file (you might want to add more sophisticated detection)
		if filepath.Ext(path) == ".gguf" || filepath.Ext(path) == ".bin" {
			if err := m.registerLocalModel(path); err != nil {
				// Log error but continue
				fmt.Printf("Failed to register model %s: %v\n", path, err)
			}
		}

		return nil
	})
}

// registerLocalModel registers a local model file
func (m *Manager) registerLocalModel(path string) error {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat model file: %w", err)
	}

	// Calculate checksum
	checksum, err := m.calculateChecksum(path)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Extract model name from path
	name := filepath.Base(path)
	name = name[:len(name)-len(filepath.Ext(name))]

	// Create model entry
	model := &Model{
		Name:         name,
		Version:      "1.0.0", // TODO: Extract version from filename or metadata
		Size:         info.Size(),
		Checksum:     checksum,
		Path:         path,
		Status:       ModelStatusAvailable,
		Replicas:     []string{m.p2p.ID().String()},
		Metadata:     make(map[string]string),
		CreatedAt:    info.ModTime(),
		UpdatedAt:    info.ModTime(),
		AccessCount:  0,
		LastAccessed: time.Now(),
	}

	m.modelsMu.Lock()
	m.models[name] = model
	m.modelsMu.Unlock()

	return nil
}

// calculateChecksum calculates SHA256 checksum of a file
func (m *Manager) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// cleanupRoutine runs periodic cleanup tasks
func (m *Manager) cleanupRoutine() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

// cleanup performs cleanup tasks
func (m *Manager) cleanup() {
	// Remove old transfers
	m.transfersMu.Lock()
	for id, transfer := range m.transfers {
		if transfer.Status == TransferStatusCompleted || transfer.Status == TransferStatusFailed {
			if time.Since(transfer.CompletedAt) > time.Hour {
				delete(m.transfers, id)
			}
		}
	}
	m.transfersMu.Unlock()

	// Clean up old model files based on cleanup age
	m.modelsMu.Lock()
	for name, model := range m.models {
		if time.Since(model.LastAccessed) > m.config.CleanupAge {
			// Remove model file
			if err := os.Remove(model.Path); err == nil {
				delete(m.models, name)
			}
		}
	}
	m.modelsMu.Unlock()
}

// DownloadModel downloads a model from peers
func (m *Manager) DownloadModel(modelName string, peerID string) (*Model, error) {
	responseCh := make(chan *DownloadResponse, 1)

	req := &DownloadRequest{
		ModelName:  modelName,
		PeerID:     peerID,
		Priority:   1,
		ResponseCh: responseCh,
	}

	select {
	case m.downloadQueue <- req:
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("download queue full")
	}

	select {
	case response := <-responseCh:
		if response.Success {
			return response.Model, nil
		}
		return nil, fmt.Errorf("download failed: %s", response.Error)
	case <-time.After(10 * time.Minute):
		return nil, fmt.Errorf("download timeout")
	}
}

// GetModel returns a model by name
func (m *Manager) GetModel(name string) (*Model, bool) {
	m.modelsMu.RLock()
	defer m.modelsMu.RUnlock()

	model, exists := m.models[name]
	if exists {
		// Update access statistics
		model.AccessCount++
		model.LastAccessed = time.Now()
	}

	return model, exists
}

// GetAllModels returns all models
func (m *Manager) GetAllModels() map[string]*Model {
	m.modelsMu.RLock()
	defer m.modelsMu.RUnlock()

	models := make(map[string]*Model)
	for k, v := range m.models {
		models[k] = v
	}

	return models
}

// GetTransfer returns a transfer by ID
func (m *Manager) GetTransfer(id string) (*Transfer, bool) {
	m.transfersMu.RLock()
	defer m.transfersMu.RUnlock()

	transfer, exists := m.transfers[id]
	return transfer, exists
}

// GetAllTransfers returns all transfers
func (m *Manager) GetAllTransfers() map[string]*Transfer {
	m.transfersMu.RLock()
	defer m.transfersMu.RUnlock()

	transfers := make(map[string]*Transfer)
	for k, v := range m.transfers {
		transfers[k] = v
	}

	return transfers
}

// Shutdown gracefully shuts down the model manager
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	// Stop workers
	for _, worker := range m.downloadWorkers {
		close(worker.stopCh)
	}

	for _, worker := range m.uploadWorkers {
		close(worker.stopCh)
	}

	// Cancel ongoing transfers
	m.transfersMu.Lock()
	for _, transfer := range m.transfers {
		if transfer.Status == TransferStatusActive {
			transfer.cancel()
		}
	}
	m.transfersMu.Unlock()

	// Cancel context
	m.cancel()

	m.started = false
	return nil
}

// DownloadWorker methods

// start starts the download worker
func (w *DownloadWorker) start() {
	for {
		select {
		case <-w.stopCh:
			return
		case req := <-w.manager.downloadQueue:
			w.processDownload(req)
		}
	}
}

// processDownload processes a download request
func (w *DownloadWorker) processDownload(req *DownloadRequest) {
	start := time.Now()

	// Create transfer entry
	transferID := fmt.Sprintf("download_%s_%d", req.ModelName, time.Now().UnixNano())
	ctx, cancel := context.WithCancel(w.manager.ctx)

	transfer := &Transfer{
		ID:        transferID,
		ModelName: req.ModelName,
		Type:      TransferTypeDownload,
		Status:    TransferStatusPending,
		Progress:  0,
		PeerID:    req.PeerID,
		StartedAt: start,
		ctx:       ctx,
		cancel:    cancel,
	}

	w.manager.transfersMu.Lock()
	w.manager.transfers[transferID] = transfer
	w.manager.transfersMu.Unlock()

	// TODO: Implement actual download from peer
	// For now, simulate download
	model, err := w.simulateDownload(transfer)

	response := &DownloadResponse{
		Success:  err == nil,
		Model:    model,
		Duration: time.Since(start),
	}

	if err != nil {
		response.Error = err.Error()
		transfer.Status = TransferStatusFailed
		transfer.Error = err.Error()
	} else {
		transfer.Status = TransferStatusCompleted
	}

	transfer.CompletedAt = time.Now()

	select {
	case req.ResponseCh <- response:
	case <-time.After(5 * time.Second):
		// Response channel blocked
	}
}

// simulateDownload simulates downloading a model
func (w *DownloadWorker) simulateDownload(transfer *Transfer) (*Model, error) {
	// Update transfer status
	transfer.Status = TransferStatusActive
	transfer.BytesTotal = 1024 * 1024 * 100 // 100MB

	// Simulate download progress
	for i := 0; i < 10; i++ {
		select {
		case <-transfer.ctx.Done():
			return nil, fmt.Errorf("download cancelled")
		default:
		}

		transfer.BytesDone = int64(i+1) * (transfer.BytesTotal / 10)
		transfer.Progress = float64(transfer.BytesDone) / float64(transfer.BytesTotal) * 100

		time.Sleep(100 * time.Millisecond)
	}

	// Create model file path
	modelPath := filepath.Join(w.manager.config.ModelDir, transfer.ModelName+".gguf")

	// Create dummy model file
	file, err := os.Create(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create model file: %w", err)
	}

	// Write some dummy data
	if _, err := file.WriteString("dummy model data"); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write model file: %w", err)
	}
	file.Close()

	// Calculate checksum
	checksum, err := w.manager.calculateChecksum(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Create model entry
	model := &Model{
		Name:         transfer.ModelName,
		Version:      "1.0.0",
		Size:         transfer.BytesTotal,
		Checksum:     checksum,
		Path:         modelPath,
		Status:       ModelStatusAvailable,
		Replicas:     []string{w.manager.p2p.ID().String()},
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		AccessCount:  0,
		LastAccessed: time.Now(),
	}

	// Store model
	w.manager.modelsMu.Lock()
	w.manager.models[transfer.ModelName] = model
	w.manager.modelsMu.Unlock()

	return model, nil
}

// UploadWorker methods

// start starts the upload worker
func (w *UploadWorker) start() {
	for {
		select {
		case <-w.stopCh:
			return
		case req := <-w.manager.uploadQueue:
			w.processUpload(req)
		}
	}
}

// processUpload processes an upload request
func (w *UploadWorker) processUpload(req *UploadRequest) {
	start := time.Now()

	// TODO: Implement actual upload to peer
	// For now, simulate upload
	err := w.simulateUpload(req)

	response := &UploadResponse{
		Success:  err == nil,
		Duration: time.Since(start),
	}

	if err != nil {
		response.Error = err.Error()
	}

	select {
	case req.ResponseCh <- response:
	case <-time.After(5 * time.Second):
		// Response channel blocked
	}
}

// simulateUpload simulates uploading a model
func (w *UploadWorker) simulateUpload(req *UploadRequest) error {
	// Check if model exists
	model, exists := w.manager.GetModel(req.ModelName)
	if !exists {
		return fmt.Errorf("model %s not found", req.ModelName)
	}

	// Simulate upload time
	time.Sleep(time.Duration(model.Size/1024/1024) * time.Millisecond)

	return nil
}

// ShouldDistribute determines if a model should be distributed
func (m *Manager) ShouldDistribute(modelName string) bool {
	// Check if model exists and is suitable for distribution
	_, exists := m.GetModel(modelName)
	return exists
}

// IsDistributed checks if a model is already distributed
func (m *Manager) IsDistributed(modelName string) bool {
	// Check if model exists in the distributed network
	_, exists := m.GetModel(modelName)
	return exists
}

// GetModelInfo returns information about a model
func (m *Manager) GetModelInfo(modelName string) map[string]interface{} {
	if model, exists := m.GetModel(modelName); exists {
		return map[string]interface{}{
			"name":     model.Name,
			"version":  model.Version,
			"size":     model.Size,
			"checksum": model.Checksum,
			"path":     model.Path,
			"created":  model.CreatedAt,
			"accessed": model.LastAccessed,
		}
	}
	return nil
}

// GetDistributedModels returns all distributed models as API responses
func (m *Manager) GetDistributedModels() []interface{} {
	m.modelsMu.RLock()
	defer m.modelsMu.RUnlock()

	var models []interface{}
	for _, model := range m.models {
		models = append(models, map[string]interface{}{
			"name":     model.Name,
			"version":  model.Version,
			"size":     model.Size,
			"checksum": model.Checksum,
			"path":     model.Path,
			"created":  model.CreatedAt,
			"accessed": model.LastAccessed,
		})
	}
	return models
}

// DeleteModel deletes a model from the distributed system
func (m *Manager) DeleteModel(modelName string) error {
	m.modelsMu.Lock()
	defer m.modelsMu.Unlock()

	if _, exists := m.models[modelName]; !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	delete(m.models, modelName)
	return nil
}

// GetDistributedModelCount returns the count of distributed models
func (m *Manager) GetDistributedModelCount() int {
	m.modelsMu.RLock()
	defer m.modelsMu.RUnlock()
	return len(m.models)
}

// DownloadFromPeer downloads a model from a peer
func (m *Manager) DownloadFromPeer(modelName, peerID string) error {
	// This is a wrapper around the existing DownloadModel method
	_, err := m.DownloadModel(modelName, peerID)
	return err
}

// RegisterModel registers a model in the distributed system
func (m *Manager) RegisterModel(modelName, modelPath string) error {
	return m.registerLocalModel(modelPath)
}

// Rebalance rebalances models across the distributed network
func (m *Manager) Rebalance() error {
	// Stub implementation for rebalancing logic
	return nil
}

// MigrateModel migrates a model to a different node
func (m *Manager) MigrateModel(modelName, targetNodeID string) error {
	// Stub implementation for model migration
	return nil
}

// GetStats returns statistics about the distributed system
func (m *Manager) GetStats() map[string]interface{} {
	m.modelsMu.RLock()
	defer m.modelsMu.RUnlock()

	return map[string]interface{}{
		"total_models":     len(m.models),
		"total_transfers":  len(m.transfers),
		"active_downloads": len(m.downloadQueue),
		"active_uploads":   len(m.uploadQueue),
	}
}

// ForceRebalance forces a rebalancing operation
func (m *Manager) ForceRebalance() error {
	// Stub implementation for forced rebalancing
	return m.Rebalance()
}

// StartP2PTransfer starts a P2P model transfer
func (m *Manager) StartP2PTransfer(modelName, modelVersion string, sourcePeer, targetPeer peer.ID, totalSize int64, expectedChecksum string) (*P2PTransfer, error) {
	if m.p2pEngine == nil {
		return nil, fmt.Errorf("P2P transfer engine not initialized")
	}
	return m.p2pEngine.StartTransfer(modelName, modelVersion, sourcePeer, targetPeer, totalSize, expectedChecksum)
}

// GetP2PTransfer returns a P2P transfer by ID
func (m *Manager) GetP2PTransfer(transferID string) (*P2PTransfer, bool) {
	if m.p2pEngine == nil {
		return nil, false
	}
	return m.p2pEngine.GetTransfer(transferID)
}

// GetActiveP2PTransfers returns all active P2P transfers
func (m *Manager) GetActiveP2PTransfers() []*P2PTransfer {
	if m.p2pEngine == nil {
		return nil
	}
	return m.p2pEngine.GetActiveTransfers()
}

// VerifyModelIntegrity verifies the integrity of a model file
func (m *Manager) VerifyModelIntegrity(modelName, modelVersion, filePath string, expectedChecksums map[HashAlgorithm]string) (*VerificationResult, error) {
	if m.verifier == nil {
		return nil, fmt.Errorf("integrity verifier not initialized")
	}
	return m.verifier.VerifyModel(modelName, modelVersion, filePath, expectedChecksums)
}

// QuickVerifyModel performs a quick SHA256 verification
func (m *Manager) QuickVerifyModel(filePath string, expectedSHA256 string) (bool, error) {
	if m.verifier == nil {
		return false, fmt.Errorf("integrity verifier not initialized")
	}
	return m.verifier.QuickVerify(filePath, expectedSHA256)
}

// GetTransferMetrics returns P2P transfer metrics
func (m *Manager) GetTransferMetrics() *TransferMetrics {
	if m.p2pEngine == nil {
		return &TransferMetrics{}
	}
	return m.p2pEngine.GetMetrics()
}

// GetVerificationMetrics returns verification metrics
func (m *Manager) GetVerificationMetrics() *VerificationMetrics {
	if m.verifier == nil {
		return &VerificationMetrics{}
	}
	return m.verifier.GetMetrics()
}

// RegisterModelVersion registers a new model version
func (m *Manager) RegisterModelVersion(version *DetailedModelVersion) error {
	if m.versionManager == nil {
		return fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.RegisterModelVersion(version)
}

// GetModelVersion retrieves a specific model version
func (m *Manager) GetModelVersion(modelName, version string) (*DetailedModelVersion, error) {
	if m.versionManager == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.GetModelVersion(modelName, version)
}

// GetLatestModelVersion returns the latest version of a model
func (m *Manager) GetLatestModelVersion(modelName string, stableOnly bool) (*DetailedModelVersion, error) {
	if m.versionManager == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.GetLatestVersion(modelName, stableOnly)
}

// ListModelVersions lists all versions of a model
func (m *Manager) ListModelVersions(modelName string, includeDeprecated bool) ([]*DetailedModelVersion, error) {
	if m.versionManager == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.ListModelVersions(modelName, includeDeprecated)
}

// CreateVersionTag creates a tag for a specific version
func (m *Manager) CreateVersionTag(modelName, version, tag string) error {
	if m.versionManager == nil {
		return fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.CreateVersionTag(modelName, version, tag)
}

// DeprecateModelVersion marks a version as deprecated
func (m *Manager) DeprecateModelVersion(modelName, version string, reason string) error {
	if m.versionManager == nil {
		return fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.DeprecateVersion(modelName, version, reason)
}

// GetModelDependencyGraph returns the dependency graph for a model
func (m *Manager) GetModelDependencyGraph(modelName string) (*DependencyGraph, error) {
	if m.versionManager == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.GetDependencyGraph(modelName)
}

// GetVersionHistory returns the version history for a model
func (m *Manager) GetVersionHistory(modelName string, limit int) ([]*VersionEvent, error) {
	if m.versionManager == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return m.versionManager.GetVersionHistory(modelName, limit)
}

// GetVersionMetrics returns versioning metrics
func (m *Manager) GetVersionMetrics() *VersionMetrics {
	if m.versionManager == nil {
		return &VersionMetrics{}
	}
	return m.versionManager.GetMetrics()
}

// CreateReplicationSet creates a new replication set for a model
func (m *Manager) CreateReplicationSet(modelName, modelVersion string, policy *AdvancedReplicationPolicy) (*ReplicationSet, error) {
	if m.advancedReplication == nil {
		return nil, fmt.Errorf("advanced replication manager not initialized")
	}
	return m.advancedReplication.CreateReplicationSet(modelName, modelVersion, policy)
}

// GetReplicationSet returns a replication set by model name and version
func (m *Manager) GetReplicationSet(modelName, modelVersion string) (*ReplicationSet, error) {
	if m.advancedReplication == nil {
		return nil, fmt.Errorf("advanced replication manager not initialized")
	}
	return m.advancedReplication.GetReplicationSet(modelName, modelVersion)
}

// GetReplicationMetrics returns replication metrics
func (m *Manager) GetReplicationMetrics() *ReplicationMetrics {
	if m.advancedReplication == nil {
		return &ReplicationMetrics{
			StorageUtilization: make(map[peer.ID]float64),
			NetworkUtilization: make(map[peer.ID]float64),
		}
	}
	return m.advancedReplication.GetMetrics()
}

// RegisterModelInLifecycle registers a model in the lifecycle management system
func (m *Manager) RegisterModelInLifecycle(modelName, initialVersion string, metadata map[string]interface{}) (*ModelRegistryEntry, error) {
	if m.lifecycleManager == nil {
		return nil, fmt.Errorf("lifecycle manager not initialized")
	}
	return m.lifecycleManager.RegisterModel(modelName, initialVersion, metadata)
}

// TransitionModelLifecycle transitions a model to a new lifecycle stage
func (m *Manager) TransitionModelLifecycle(modelID string, targetStage ModelLifecycleStage, trigger TransitionTrigger, actor, reason string) error {
	if m.lifecycleManager == nil {
		return fmt.Errorf("lifecycle manager not initialized")
	}
	return m.lifecycleManager.TransitionModel(modelID, targetStage, trigger, actor, reason)
}

// GetModelRegistryEntry returns the model registry entry for a model
func (m *Manager) GetModelRegistryEntry(modelID string) (*ModelRegistryEntry, error) {
	if m.lifecycleManager == nil {
		return nil, fmt.Errorf("lifecycle manager not initialized")
	}
	return m.lifecycleManager.GetModelRegistry(modelID)
}

// GetModelLifecycleState returns the lifecycle state for a model
func (m *Manager) GetModelLifecycleState(modelID string) (*LifecycleState, error) {
	if m.lifecycleManager == nil {
		return nil, fmt.Errorf("lifecycle manager not initialized")
	}
	return m.lifecycleManager.GetLifecycleState(modelID)
}

// GetLifecycleMetrics returns lifecycle metrics
func (m *Manager) GetLifecycleMetrics() *LifecycleMetrics {
	if m.lifecycleManager == nil {
		return &LifecycleMetrics{
			ModelsByStage:      make(map[ModelLifecycleStage]int64),
			ModelsByStatus:     make(map[ModelStatus]int64),
			TransitionsByStage: make(map[ModelLifecycleStage]int64),
			ActionsByType:      make(map[ActionType]int64),
		}
	}
	return m.lifecycleManager.GetMetrics()
}

// StoreInCAS stores content in the content-addressed storage
func (m *Manager) StoreInCAS(content io.Reader, metadata map[string]interface{}) (*ContentEntry, error) {
	if m.advancedCAS == nil {
		return nil, fmt.Errorf("advanced CAS not initialized")
	}
	return m.advancedCAS.Store(content, metadata)
}

// RetrieveFromCAS retrieves content from the content-addressed storage
func (m *Manager) RetrieveFromCAS(hash string) (io.ReadCloser, error) {
	if m.advancedCAS == nil {
		return nil, fmt.Errorf("advanced CAS not initialized")
	}
	return m.advancedCAS.Retrieve(hash)
}

// GetCASMetrics returns CAS metrics
func (m *Manager) GetCASMetrics() *CASMetrics {
	if m.advancedCAS == nil {
		return &CASMetrics{
			BackendMetrics: make(map[string]*BackendMetrics),
		}
	}
	return m.advancedCAS.GetMetrics()
}

// SynchronizeModel synchronizes a model across the distributed system
func (m *Manager) SynchronizeModel(modelName string, targetNodes []peer.ID) error {
	if m.syncEngine == nil {
		return fmt.Errorf("sync engine not initialized")
	}
	return m.syncEngine.SynchronizeModel(modelName, targetNodes)
}

// GetModelSyncState returns the synchronization state for a model
func (m *Manager) GetModelSyncState(modelName string) (*EngineSyncState, error) {
	if m.syncEngine == nil {
		return nil, fmt.Errorf("sync engine not initialized")
	}
	return m.syncEngine.GetSyncState(modelName)
}

// GetSyncMetrics returns synchronization metrics
func (m *Manager) GetSyncMetrics() *SyncMetrics {
	if m.syncEngine == nil {
		return &SyncMetrics{
			ConflictsByType: make(map[EngineConflictType]int64),
			NodeSyncStats:   make(map[peer.ID]*NodeSyncStats),
		}
	}
	return m.syncEngine.GetMetrics()
}

// GetLogger returns the structured logger
func (m *Manager) GetLogger() *logging.StructuredLogger {
	return m.logger
}

// GetErrorHandler returns the error handler
func (m *Manager) GetErrorHandler() *errors.ErrorHandler {
	return m.errorHandler
}

// GetMetricsCollector returns the metrics collector
func (m *Manager) GetMetricsCollector() *observability.MetricsCollector {
	return m.metricsCollector
}

// GetTracer returns the tracer
func (m *Manager) GetTracer() *observability.Tracer {
	return m.tracer
}

// LogInfo logs an info message
func (m *Manager) LogInfo(msg string, fields ...interface{}) {
	if m.logger != nil {
		// Convert fields to slog.Attr format
		// This is a simplified implementation
		m.logger.Info(msg)
	}
}

// LogError logs an error message
func (m *Manager) LogError(msg string, err error, fields ...interface{}) {
	if m.logger != nil {
		m.logger.Error(msg, err)
	}
}

// HandleError handles an error with context
func (m *Manager) HandleError(ctx context.Context, err error) *errors.DistributedError {
	if m.errorHandler != nil {
		return m.errorHandler.Handle(ctx, err)
	}

	// Fallback to basic error handling
	return errors.InternalError("Error handling not initialized", err)
}

// GetTotalModels returns the total number of models in the system
func (m *Manager) GetTotalModels() int {
	return m.GetDistributedModelCount()
}
