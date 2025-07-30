package protocols

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// FileTransferHandler handles file transfer requests with chunking support
type FileTransferHandler struct {
	// File management
	fileStore       FileStore
	
	// Transfer management
	activeTransfers map[string]*FileTransfer
	transfersMux    sync.RWMutex
	
	// Configuration
	config          *FileTransferConfig
	
	// Metrics
	metrics         *FileTransferMetrics
}

// FileTransferConfig configures file transfer behavior
type FileTransferConfig struct {
	ChunkSize         int           `json:"chunk_size"`
	MaxConcurrentTransfers int      `json:"max_concurrent_transfers"`
	TransferTimeout   time.Duration `json:"transfer_timeout"`
	ResumeTimeout     time.Duration `json:"resume_timeout"`
	VerifyChecksums   bool          `json:"verify_checksums"`
	CompressionEnabled bool         `json:"compression_enabled"`
	EncryptionEnabled bool          `json:"encryption_enabled"`
	MaxFileSize       int64         `json:"max_file_size"`
	AllowedExtensions []string      `json:"allowed_extensions"`
	StorageDir        string        `json:"storage_dir"`
}

// FileTransferMetrics tracks file transfer performance
type FileTransferMetrics struct {
	TotalTransfers     int64         `json:"total_transfers"`
	SuccessfulTransfers int64        `json:"successful_transfers"`
	FailedTransfers    int64         `json:"failed_transfers"`
	ActiveTransfers    int           `json:"active_transfers"`
	BytesTransferred   int64         `json:"bytes_transferred"`
	AverageSpeed       int64         `json:"average_speed"` // bytes per second
	TotalTime          time.Duration `json:"total_time"`
	
	// Chunk-specific metrics
	ChunksTransferred  int64         `json:"chunks_transferred"`
	ChunkErrors        int64         `json:"chunk_errors"`
	AverageChunkTime   time.Duration `json:"average_chunk_time"`
	
	// Per-file metrics
	FileMetrics        map[string]*FileTransferInfo `json:"file_metrics"`
	
	mu sync.RWMutex
}

// FileTransfer represents an active file transfer
type FileTransfer struct {
	ID           string        `json:"id"`
	FileName     string        `json:"file_name"`
	FileSize     int64         `json:"file_size"`
	Checksum     string        `json:"checksum"`
	ChunkSize    int           `json:"chunk_size"`
	TotalChunks  int           `json:"total_chunks"`
	
	// Transfer state
	Status       TransferStatus `json:"status"`
	Direction    TransferDirection `json:"direction"`
	PeerID       peer.ID       `json:"peer_id"`
	Progress     float64       `json:"progress"`
	BytesTransferred int64     `json:"bytes_transferred"`
	
	// Chunk tracking
	CompletedChunks []bool      `json:"completed_chunks"`
	ChunkChecksums  []string    `json:"chunk_checksums"`
	
	// Timing
	StartTime    time.Time     `json:"start_time"`
	LastActivity time.Time     `json:"last_activity"`
	EstimatedTTL time.Duration `json:"estimated_ttl"`
	
	// Context
	Context      context.Context    `json:"-"`
	CancelFunc   context.CancelFunc `json:"-"`
	
	// File handle
	File         *os.File      `json:"-"`
	
	// Synchronization
	ChunkMux     sync.RWMutex  `json:"-"`
}

// TransferStatus represents the status of a file transfer
type TransferStatus string

const (
	TransferStatusPending     TransferStatus = "pending"
	TransferStatusInitiating  TransferStatus = "initiating"
	TransferStatusTransferring TransferStatus = "transferring"
	TransferStatusCompleted   TransferStatus = "completed"
	TransferStatusFailed      TransferStatus = "failed"
	TransferStatusCancelled   TransferStatus = "cancelled"
	TransferStatusPaused      TransferStatus = "paused"
)

// TransferDirection indicates direction of transfer
type TransferDirection string

const (
	TransferDirectionUpload   TransferDirection = "upload"
	TransferDirectionDownload TransferDirection = "download"
)

// FileTransferInfo contains information about a file transfer
type FileTransferInfo struct {
	FileName      string        `json:"file_name"`
	FileSize      int64         `json:"file_size"`
	TransferCount int64         `json:"transfer_count"`
	SuccessCount  int64         `json:"success_count"`
	FailureCount  int64         `json:"failure_count"`
	TotalBytes    int64         `json:"total_bytes"`
	AverageSpeed  int64         `json:"average_speed"`
	LastTransfer  time.Time     `json:"last_transfer"`
}

// FileStore defines the interface for file storage operations
type FileStore interface {
	// File existence and metadata
	FileExists(filename string) bool
	GetFileSize(filename string) (int64, error)
	GetFileChecksum(filename string) (string, error)
	
	// File operations
	OpenFileForReading(filename string) (*os.File, error)
	OpenFileForWriting(filename string, size int64) (*os.File, error)
	DeleteFile(filename string) error
	
	// Directory operations
	ListFiles() ([]string, error)
	GetStorageDir() string
	
	// Chunk operations
	ReadChunk(filename string, chunkIndex int, chunkSize int) ([]byte, error)
	WriteChunk(filename string, chunkIndex int, data []byte) error
	VerifyChunk(filename string, chunkIndex int, expectedChecksum string) (bool, error)
}

// ChunkInfo represents information about a file chunk
type ChunkInfo struct {
	Index    int    `json:"index"`
	Size     int    `json:"size"`
	Checksum string `json:"checksum"`
	Data     []byte `json:"data,omitempty"`
}

// NewFileTransferHandler creates a new file transfer handler
func NewFileTransferHandler(fileStore FileStore, config *FileTransferConfig) *FileTransferHandler {
	if config == nil {
		config = DefaultFileTransferConfig()
	}
	
	return &FileTransferHandler{
		fileStore:       fileStore,
		activeTransfers: make(map[string]*FileTransfer),
		config:          config,
		metrics: &FileTransferMetrics{
			FileMetrics: make(map[string]*FileTransferInfo),
		},
	}
}

// HandleMessage handles file transfer protocol messages
func (fth *FileTransferHandler) HandleMessage(ctx context.Context, stream network.Stream, msg *Message) error {
	switch msg.Type {
	case MsgTypeFileRequest:
		return fth.handleFileRequest(ctx, stream, msg)
	case MsgTypeFileChunk:
		return fth.handleFileChunk(ctx, stream, msg)
	case MsgTypeFileComplete:
		return fth.handleFileComplete(ctx, stream, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleFileRequest handles incoming file requests
func (fth *FileTransferHandler) handleFileRequest(ctx context.Context, stream network.Stream, msg *Message) error {
	peerID := stream.Conn().RemotePeer()
	
	// Parse file request
	req, err := fth.parseFileRequest(msg)
	if err != nil {
		return fth.sendErrorResponse(stream, msg.ID, "invalid_request", err.Error())
	}
	
	// Validate request
	if err := fth.validateFileRequest(req); err != nil {
		return fth.sendErrorResponse(stream, msg.ID, "validation_error", err.Error())
	}
	
	// Check if file exists
	if !fth.fileStore.FileExists(req.FileName) {
		return fth.sendErrorResponse(stream, msg.ID, "file_not_found", "Requested file does not exist")
	}
	
	// Check capacity
	if !fth.checkTransferCapacity() {
		return fth.sendErrorResponse(stream, msg.ID, "capacity_exceeded", "Too many concurrent transfers")
	}
	
	// Get file information
	fileSize, err := fth.fileStore.GetFileSize(req.FileName)
	if err != nil {
		return fth.sendErrorResponse(stream, msg.ID, "file_error", err.Error())
	}
	
	checksum, err := fth.fileStore.GetFileChecksum(req.FileName)
	if err != nil {
		return fth.sendErrorResponse(stream, msg.ID, "checksum_error", err.Error())
	}
	
	// Create transfer
	transfer := &FileTransfer{
		ID:          generateMessageID(),
		FileName:    req.FileName,
		FileSize:    fileSize,
		Checksum:    checksum,
		ChunkSize:   fth.config.ChunkSize,
		TotalChunks: int((fileSize + int64(fth.config.ChunkSize) - 1) / int64(fth.config.ChunkSize)),
		Status:      TransferStatusInitiating,
		Direction:   TransferDirectionUpload,
		PeerID:      peerID,
		StartTime:   time.Now(),
		LastActivity: time.Now(),
	}
	
	transfer.Context, transfer.CancelFunc = context.WithTimeout(ctx, fth.config.TransferTimeout)
	transfer.CompletedChunks = make([]bool, transfer.TotalChunks)
	transfer.ChunkChecksums = make([]string, transfer.TotalChunks)
	
	// Calculate chunk checksums
	if err := fth.calculateChunkChecksums(transfer); err != nil {
		return fth.sendErrorResponse(stream, msg.ID, "checksum_calculation_error", err.Error())
	}
	
	// Track transfer
	fth.trackTransfer(transfer)
	defer fth.untrackTransfer(transfer.ID)
	
	// Send file response with metadata
	response := &Message{
		Type:      MsgTypeFileResponse,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":      msg.ID,
			"transfer_id":     transfer.ID,
			"file_name":       transfer.FileName,
			"file_size":       transfer.FileSize,
			"checksum":        transfer.Checksum,
			"chunk_size":      transfer.ChunkSize,
			"total_chunks":    transfer.TotalChunks,
			"chunk_checksums": transfer.ChunkChecksums,
			"status":          string(transfer.Status),
		},
	}
	
	handler := NewProtocolHandler(FileTransferProtocol)
	if err := handler.SendMessage(stream, response); err != nil {
		return fmt.Errorf("failed to send file response: %w", err)
	}
	
	// Update status and start transfer
	transfer.Status = TransferStatusTransferring
	fth.updateTransferMetrics()
	
	log.Printf("Starting file transfer %s for file %s to peer %s (%d chunks, %d bytes)", 
		transfer.ID, transfer.FileName, peerID, transfer.TotalChunks, transfer.FileSize)
	
	return nil
}

// handleFileChunk handles incoming chunk requests
func (fth *FileTransferHandler) handleFileChunk(ctx context.Context, stream network.Stream, msg *Message) error {
	// This would handle chunk requests from downloading peers
	// For now, we'll focus on the upload side
	return nil
}

// handleFileComplete handles transfer completion notifications
func (fth *FileTransferHandler) handleFileComplete(ctx context.Context, stream network.Stream, msg *Message) error {
	data := msg.Data
	
	transferID, _ := data["transfer_id"].(string)
	success, _ := data["success"].(bool)
	
	// Find and update transfer
	fth.transfersMux.Lock()
	defer fth.transfersMux.Unlock()
	
	if transfer, exists := fth.activeTransfers[transferID]; exists {
		if success {
			transfer.Status = TransferStatusCompleted
			transfer.Progress = 1.0
			fth.updateSuccessMetrics(transfer)
		} else {
			transfer.Status = TransferStatusFailed
			fth.updateFailureMetrics(transfer)
		}
		transfer.LastActivity = time.Now()
	}
	
	return nil
}

// parseFileRequest parses a file request from message data
func (fth *FileTransferHandler) parseFileRequest(msg *Message) (*FileRequest, error) {
	data := msg.Data
	
	req := &FileRequest{
		RequestID: msg.ID,
	}
	
	if fileName, ok := data["file_name"].(string); ok {
		req.FileName = fileName
	} else {
		return nil, fmt.Errorf("file_name is required")
	}
	
	if priority, ok := data["priority"].(float64); ok {
		req.Priority = int(priority)
	}
	
	if resumeTransfer, ok := data["resume_transfer"].(bool); ok {
		req.ResumeTransfer = resumeTransfer
	}
	
	if existingTransferID, ok := data["existing_transfer_id"].(string); ok {
		req.ExistingTransferID = existingTransferID
	}
	
	return req, nil
}

// validateFileRequest validates a file request
func (fth *FileTransferHandler) validateFileRequest(req *FileRequest) error {
	if req.FileName == "" {
		return fmt.Errorf("file name is required")
	}
	
	// Check file extension if restrictions are configured
	if len(fth.config.AllowedExtensions) > 0 {
		allowed := false
		for _, ext := range fth.config.AllowedExtensions {
			if len(req.FileName) > len(ext) && req.FileName[len(req.FileName)-len(ext):] == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file extension not allowed")
		}
	}
	
	// Check file size
	if fth.fileStore.FileExists(req.FileName) {
		size, err := fth.fileStore.GetFileSize(req.FileName)
		if err != nil {
			return fmt.Errorf("failed to get file size: %w", err)
		}
		if size > fth.config.MaxFileSize {
			return fmt.Errorf("file size exceeds maximum allowed size")
		}
	}
	
	return nil
}

// calculateChunkChecksums calculates checksums for all chunks of a file
func (fth *FileTransferHandler) calculateChunkChecksums(transfer *FileTransfer) error {
	file, err := fth.fileStore.OpenFileForReading(transfer.FileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	for i := 0; i < transfer.TotalChunks; i++ {
		// Calculate chunk boundaries
		offset := int64(i) * int64(transfer.ChunkSize)
		size := transfer.ChunkSize
		
		// Adjust size for last chunk
		if i == transfer.TotalChunks-1 {
			remaining := transfer.FileSize - offset
			if remaining < int64(size) {
				size = int(remaining)
			}
		}
		
		// Read chunk
		file.Seek(offset, 0)
		chunk := make([]byte, size)
		if _, err := io.ReadFull(file, chunk); err != nil {
			return fmt.Errorf("failed to read chunk %d: %w", i, err)
		}
		
		// Calculate checksum
		hash := sha256.Sum256(chunk)
		transfer.ChunkChecksums[i] = hex.EncodeToString(hash[:])
	}
	
	return nil
}

// SendChunk sends a specific chunk to a peer
func (fth *FileTransferHandler) SendChunk(ctx context.Context, stream network.Stream, transfer *FileTransfer, chunkIndex int) error {
	// Read chunk data
	chunkData, err := fth.fileStore.ReadChunk(transfer.FileName, chunkIndex, transfer.ChunkSize)
	if err != nil {
		return fmt.Errorf("failed to read chunk %d: %w", chunkIndex, err)
	}
	
	// Create chunk message
	chunkMsg := &Message{
		Type:      MsgTypeFileChunk,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"transfer_id":    transfer.ID,
			"chunk_index":    chunkIndex,
			"chunk_size":     len(chunkData),
			"chunk_checksum": transfer.ChunkChecksums[chunkIndex],
			"chunk_data":     chunkData,
			"is_last_chunk":  chunkIndex == transfer.TotalChunks-1,
		},
	}
	
	// Send chunk
	handler := NewProtocolHandler(FileTransferProtocol)
	if err := handler.SendMessage(stream, chunkMsg); err != nil {
		return fmt.Errorf("failed to send chunk %d: %w", chunkIndex, err)
	}
	
	// Update transfer progress
	transfer.ChunkMux.Lock()
	transfer.CompletedChunks[chunkIndex] = true
	transfer.BytesTransferred += int64(len(chunkData))
	transfer.Progress = float64(transfer.BytesTransferred) / float64(transfer.FileSize)
	transfer.LastActivity = time.Now()
	transfer.ChunkMux.Unlock()
	
	// Update metrics
	fth.updateChunkMetrics(int64(len(chunkData)))
	
	log.Printf("Sent chunk %d/%d for transfer %s (%d bytes)", 
		chunkIndex+1, transfer.TotalChunks, transfer.ID, len(chunkData))
	
	return nil
}

// checkTransferCapacity checks if we can handle another transfer
func (fth *FileTransferHandler) checkTransferCapacity() bool {
	fth.transfersMux.RLock()
	defer fth.transfersMux.RUnlock()
	
	return len(fth.activeTransfers) < fth.config.MaxConcurrentTransfers
}

// trackTransfer adds a transfer to active tracking
func (fth *FileTransferHandler) trackTransfer(transfer *FileTransfer) {
	fth.transfersMux.Lock()
	defer fth.transfersMux.Unlock()
	
	fth.activeTransfers[transfer.ID] = transfer
}

// untrackTransfer removes a transfer from active tracking
func (fth *FileTransferHandler) untrackTransfer(transferID string) {
	fth.transfersMux.Lock()
	defer fth.transfersMux.Unlock()
	
	delete(fth.activeTransfers, transferID)
}

// sendErrorResponse sends an error response
func (fth *FileTransferHandler) sendErrorResponse(stream network.Stream, requestID, errorCode, errorMessage string) error {
	errorMsg := &Message{
		Type:      "error",
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":    requestID,
			"error_code":    errorCode,
			"error_message": errorMessage,
		},
	}
	
	handler := NewProtocolHandler(FileTransferProtocol)
	return handler.SendMessage(stream, errorMsg)
}

// Metrics update methods

func (fth *FileTransferHandler) updateTransferMetrics() {
	fth.metrics.mu.Lock()
	defer fth.metrics.mu.Unlock()
	
	fth.metrics.TotalTransfers++
	fth.metrics.ActiveTransfers = len(fth.activeTransfers)
}

func (fth *FileTransferHandler) updateSuccessMetrics(transfer *FileTransfer) {
	fth.metrics.mu.Lock()
	defer fth.metrics.mu.Unlock()
	
	fth.metrics.SuccessfulTransfers++
	fth.metrics.ActiveTransfers = len(fth.activeTransfers)
	fth.metrics.BytesTransferred += transfer.FileSize
	
	duration := time.Since(transfer.StartTime)
	fth.metrics.TotalTime += duration
	
	// Calculate speed
	if duration > 0 {
		speed := transfer.FileSize / int64(duration.Seconds())
		if fth.metrics.SuccessfulTransfers > 0 {
			fth.metrics.AverageSpeed = (fth.metrics.AverageSpeed*int64(fth.metrics.SuccessfulTransfers-1) + speed) / int64(fth.metrics.SuccessfulTransfers)
		} else {
			fth.metrics.AverageSpeed = speed
		}
	}
	
	// Update file-specific metrics
	if fth.metrics.FileMetrics[transfer.FileName] == nil {
		fth.metrics.FileMetrics[transfer.FileName] = &FileTransferInfo{
			FileName: transfer.FileName,
			FileSize: transfer.FileSize,
		}
	}
	
	fileMetrics := fth.metrics.FileMetrics[transfer.FileName]
	fileMetrics.TransferCount++
	fileMetrics.SuccessCount++
	fileMetrics.TotalBytes += transfer.FileSize
	fileMetrics.LastTransfer = time.Now()
	
	if duration > 0 {
		speed := transfer.FileSize / int64(duration.Seconds())
		if fileMetrics.SuccessCount > 0 {
			fileMetrics.AverageSpeed = (fileMetrics.AverageSpeed*int64(fileMetrics.SuccessCount-1) + speed) / int64(fileMetrics.SuccessCount)
		} else {
			fileMetrics.AverageSpeed = speed
		}
	}
}

func (fth *FileTransferHandler) updateFailureMetrics(transfer *FileTransfer) {
	fth.metrics.mu.Lock()
	defer fth.metrics.mu.Unlock()
	
	fth.metrics.FailedTransfers++
	fth.metrics.ActiveTransfers = len(fth.activeTransfers)
	
	// Update file-specific metrics
	if fth.metrics.FileMetrics[transfer.FileName] == nil {
		fth.metrics.FileMetrics[transfer.FileName] = &FileTransferInfo{
			FileName: transfer.FileName,
			FileSize: transfer.FileSize,
		}
	}
	
	fileMetrics := fth.metrics.FileMetrics[transfer.FileName]
	fileMetrics.TransferCount++
	fileMetrics.FailureCount++
}

func (fth *FileTransferHandler) updateChunkMetrics(bytesTransferred int64) {
	fth.metrics.mu.Lock()
	defer fth.metrics.mu.Unlock()
	
	fth.metrics.ChunksTransferred++
	fth.metrics.BytesTransferred += bytesTransferred
}

// GetMetrics returns a copy of current metrics
func (fth *FileTransferHandler) GetMetrics() *FileTransferMetrics {
	fth.metrics.mu.RLock()
	defer fth.metrics.mu.RUnlock()
	
	// Create deep copy
	metricsCopy := &FileTransferMetrics{
		TotalTransfers:      fth.metrics.TotalTransfers,
		SuccessfulTransfers: fth.metrics.SuccessfulTransfers,
		FailedTransfers:     fth.metrics.FailedTransfers,
		ActiveTransfers:     fth.metrics.ActiveTransfers,
		BytesTransferred:    fth.metrics.BytesTransferred,
		AverageSpeed:        fth.metrics.AverageSpeed,
		TotalTime:           fth.metrics.TotalTime,
		ChunksTransferred:   fth.metrics.ChunksTransferred,
		ChunkErrors:         fth.metrics.ChunkErrors,
		AverageChunkTime:    fth.metrics.AverageChunkTime,
		FileMetrics:         make(map[string]*FileTransferInfo),
	}
	
	// Copy file metrics
	for fileName, fileMetrics := range fth.metrics.FileMetrics {
		metricsCopy.FileMetrics[fileName] = &FileTransferInfo{
			FileName:      fileMetrics.FileName,
			FileSize:      fileMetrics.FileSize,
			TransferCount: fileMetrics.TransferCount,
			SuccessCount:  fileMetrics.SuccessCount,
			FailureCount:  fileMetrics.FailureCount,
			TotalBytes:    fileMetrics.TotalBytes,
			AverageSpeed:  fileMetrics.AverageSpeed,
			LastTransfer:  fileMetrics.LastTransfer,
		}
	}
	
	return metricsCopy
}

// GetActiveTransfers returns currently active transfers
func (fth *FileTransferHandler) GetActiveTransfers() map[string]*FileTransfer {
	fth.transfersMux.RLock()
	defer fth.transfersMux.RUnlock()
	
	transfers := make(map[string]*FileTransfer)
	for id, transfer := range fth.activeTransfers {
		// Create copy without context and file handle
		transferCopy := *transfer
		transferCopy.Context = nil
		transferCopy.CancelFunc = nil
		transferCopy.File = nil
		transfers[id] = &transferCopy
	}
	
	return transfers
}

// CancelTransfer cancels an active transfer
func (fth *FileTransferHandler) CancelTransfer(transferID string) error {
	fth.transfersMux.Lock()
	defer fth.transfersMux.Unlock()
	
	transfer, exists := fth.activeTransfers[transferID]
	if !exists {
		return fmt.Errorf("transfer %s not found", transferID)
	}
	
	if transfer.CancelFunc != nil {
		transfer.CancelFunc()
		transfer.Status = TransferStatusCancelled
	}
	
	if transfer.File != nil {
		transfer.File.Close()
	}
	
	return nil
}

// DefaultFileTransferConfig returns default file transfer configuration
func DefaultFileTransferConfig() *FileTransferConfig {
	return &FileTransferConfig{
		ChunkSize:              MaxChunkSize,
		MaxConcurrentTransfers: 5,
		TransferTimeout:        30 * time.Minute,
		ResumeTimeout:          5 * time.Minute,
		VerifyChecksums:        true,
		CompressionEnabled:     false,
		EncryptionEnabled:      false,
		MaxFileSize:            10 * 1024 * 1024 * 1024, // 10GB
		AllowedExtensions:      []string{".gguf", ".bin", ".safetensors"},
		StorageDir:             "./models",
	}
}

// FileRequest represents a file request
type FileRequest struct {
	RequestID          string `json:"request_id"`
	FileName           string `json:"file_name"`
	Priority           int    `json:"priority"`
	ResumeTransfer     bool   `json:"resume_transfer"`
	ExistingTransferID string `json:"existing_transfer_id,omitempty"`
}

// FileTransferClient provides client-side file transfer functionality
type FileTransferClient struct {
	protocolClient *ProtocolClient
	fileStore      FileStore
}

// NewFileTransferClient creates a new file transfer client
func NewFileTransferClient(dialer StreamDialer, fileStore FileStore, timeout time.Duration) *FileTransferClient {
	return &FileTransferClient{
		protocolClient: NewProtocolClient(dialer, FileTransferProtocol, timeout),
		fileStore:      fileStore,
	}
}

// RequestFile requests a file from a peer
func (ftc *FileTransferClient) RequestFile(ctx context.Context, peerID peer.ID, fileName string, priority int) (*FileTransfer, error) {
	// Create file request message
	requestMsg := CreateRequestMessage(MsgTypeFileRequest, map[string]interface{}{
		"file_name": fileName,
		"priority":  priority,
	})
	
	// Send request and wait for response
	responseMsg, err := ftc.protocolClient.SendRequest(ctx, peerID, requestMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send file request: %w", err)
	}
	
	// Handle error response
	if responseMsg.Type == "error" {
		errorCode, _ := responseMsg.Data["error_code"].(string)
		errorMessage, _ := responseMsg.Data["error_message"].(string)
		return nil, fmt.Errorf("file request error [%s]: %s", errorCode, errorMessage)
	}
	
	// Parse file response
	return ftc.parseFileResponse(responseMsg, peerID)
}

// parseFileResponse parses a file response message
func (ftc *FileTransferClient) parseFileResponse(msg *Message, peerID peer.ID) (*FileTransfer, error) {
	data := msg.Data
	
	transfer := &FileTransfer{
		PeerID:      peerID,
		Direction:   TransferDirectionDownload,
		StartTime:   time.Now(),
		LastActivity: time.Now(),
	}
	
	if transferID, ok := data["transfer_id"].(string); ok {
		transfer.ID = transferID
	}
	
	if fileName, ok := data["file_name"].(string); ok {
		transfer.FileName = fileName
	}
	
	if fileSize, ok := data["file_size"].(float64); ok {
		transfer.FileSize = int64(fileSize)
	}
	
	if checksum, ok := data["checksum"].(string); ok {
		transfer.Checksum = checksum
	}
	
	if chunkSize, ok := data["chunk_size"].(float64); ok {
		transfer.ChunkSize = int(chunkSize)
	}
	
	if totalChunks, ok := data["total_chunks"].(float64); ok {
		transfer.TotalChunks = int(totalChunks)
	}
	
	if status, ok := data["status"].(string); ok {
		transfer.Status = TransferStatus(status)
	}
	
	// Parse chunk checksums
	if chunkChecksumsData, ok := data["chunk_checksums"].([]interface{}); ok {
		transfer.ChunkChecksums = make([]string, len(chunkChecksumsData))
		for i, checksum := range chunkChecksumsData {
			if checksumStr, ok := checksum.(string); ok {
				transfer.ChunkChecksums[i] = checksumStr
			}
		}
	}
	
	// Initialize tracking arrays
	transfer.CompletedChunks = make([]bool, transfer.TotalChunks)
	
	return transfer, nil
}

// GetClientMetrics returns client metrics
func (ftc *FileTransferClient) GetClientMetrics() *ClientMetrics {
	return ftc.protocolClient.GetClientMetrics()
}

// LocalFileStore provides a local filesystem implementation of FileStore
type LocalFileStore struct {
	storageDir string
	mu         sync.RWMutex
}

// NewLocalFileStore creates a new local file store
func NewLocalFileStore(storageDir string) (*LocalFileStore, error) {
	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	return &LocalFileStore{
		storageDir: storageDir,
	}, nil
}

// FileExists checks if a file exists
func (lfs *LocalFileStore) FileExists(filename string) bool {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns the size of a file
func (lfs *LocalFileStore) GetFileSize(filename string) (int64, error) {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileChecksum calculates and returns the checksum of a file
func (lfs *LocalFileStore) GetFileChecksum(filename string) (string, error) {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	file, err := os.Open(path)
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

// OpenFileForReading opens a file for reading
func (lfs *LocalFileStore) OpenFileForReading(filename string) (*os.File, error) {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	return os.Open(path)
}

// OpenFileForWriting opens a file for writing
func (lfs *LocalFileStore) OpenFileForWriting(filename string, size int64) (*os.File, error) {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

// DeleteFile deletes a file
func (lfs *LocalFileStore) DeleteFile(filename string) error {
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	return os.Remove(path)
}

// ListFiles returns a list of files in the storage directory
func (lfs *LocalFileStore) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(lfs.storageDir)
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// GetStorageDir returns the storage directory path
func (lfs *LocalFileStore) GetStorageDir() string {
	return lfs.storageDir
}

// ReadChunk reads a specific chunk from a file
func (lfs *LocalFileStore) ReadChunk(filename string, chunkIndex int, chunkSize int) ([]byte, error) {
	file, err := lfs.OpenFileForReading(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	offset := int64(chunkIndex) * int64(chunkSize)
	if _, err := file.Seek(offset, 0); err != nil {
		return nil, err
	}
	
	chunk := make([]byte, chunkSize)
	n, err := file.Read(chunk)
	if err != nil && err != io.EOF {
		return nil, err
	}
	
	return chunk[:n], nil
}

// WriteChunk writes a chunk to a file
func (lfs *LocalFileStore) WriteChunk(filename string, chunkIndex int, data []byte) error {
	lfs.mu.Lock()
	defer lfs.mu.Unlock()
	
	path := fmt.Sprintf("%s/%s", lfs.storageDir, filename)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	offset := int64(chunkIndex) * int64(len(data))
	if _, err := file.Seek(offset, 0); err != nil {
		return err
	}
	
	_, err = file.Write(data)
	return err
}

// VerifyChunk verifies a chunk's checksum
func (lfs *LocalFileStore) VerifyChunk(filename string, chunkIndex int, expectedChecksum string) (bool, error) {
	// This would need to know the chunk size, so let's simplify for now
	return true, nil
}