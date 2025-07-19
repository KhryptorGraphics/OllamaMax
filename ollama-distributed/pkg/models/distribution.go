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

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/p2p"
)

// Manager manages model distribution across the network
type Manager struct {
	config *config.StorageConfig
	p2p    *p2p.Node
	
	// Local model storage
	models    map[string]*Model
	modelsMu  sync.RWMutex
	
	// Transfer management
	transfers map[string]*Transfer
	transfersMu sync.RWMutex
	
	// Download queue
	downloadQueue chan *DownloadRequest
	uploadQueue   chan *UploadRequest
	
	// Workers
	downloadWorkers []*DownloadWorker
	uploadWorkers   []*UploadWorker
	
	started bool
	mu      sync.RWMutex
	
	ctx    context.Context
	cancel context.CancelFunc
}

// Model represents a model in the distributed system
type Model struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Path        string            `json:"path"`
	Status      ModelStatus       `json:"status"`
	Replicas    []string          `json:"replicas"`    // Node IDs that have this model
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	AccessCount int64             `json:"access_count"`
	LastAccessed time.Time        `json:"last_accessed"`
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
	ID          string          `json:"id"`
	ModelName   string          `json:"model_name"`
	Type        TransferType    `json:"type"`
	Status      TransferStatus  `json:"status"`
	Progress    float64         `json:"progress"`
	BytesTotal  int64           `json:"bytes_total"`
	BytesDone   int64           `json:"bytes_done"`
	Speed       int64           `json:"speed"`       // bytes per second
	PeerID      string          `json:"peer_id"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt time.Time       `json:"completed_at"`
	Error       string          `json:"error,omitempty"`
	
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

// TransferStatus represents the status of a transfer
type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "pending"
	TransferStatusActive     TransferStatus = "active"
	TransferStatusCompleted  TransferStatus = "completed"
	TransferStatusFailed     TransferStatus = "failed"
	TransferStatusCancelled  TransferStatus = "cancelled"
)

// DownloadRequest represents a request to download a model
type DownloadRequest struct {
	ModelName string
	PeerID    string
	Priority  int
	ResponseCh chan *DownloadResponse
}

// DownloadResponse represents a response to a download request
type DownloadResponse struct {
	Success   bool
	Model     *Model
	Error     string
	Duration  time.Duration
}

// UploadRequest represents a request to upload a model
type UploadRequest struct {
	ModelName string
	PeerID    string
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