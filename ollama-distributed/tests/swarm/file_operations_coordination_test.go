//go:build ignore

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FileOperationCoordinator manages coordinated file operations across swarm agents
type FileOperationCoordinator struct {
	workDir       string
	lockManager   *FileLockManager
	changeTracker *FileChangeTracker
	backupManager *BackupManager
	agents        map[string]*FileOperationAgent
	operations    []*FileOperation
	mu            sync.RWMutex
}

// FileOperationAgent represents an agent that performs file operations
type FileOperationAgent struct {
	ID               string
	WorkingDirectory string
	ActiveOperations map[string]*FileOperation
	CompletedOps     int
	FailedOps        int
	LastActivity     time.Time
	Lock             sync.RWMutex
}

// FileOperation represents a coordinated file operation
type FileOperation struct {
	ID           string
	Type         FileOperationType
	Path         string
	Content      []byte
	AgentID      string
	Status       OperationStatus
	StartTime    time.Time
	EndTime      time.Time
	Error        error
	Checksum     string
	BackupPath   string
	Dependencies []string
}

// FileOperationType defines types of file operations
type FileOperationType int

const (
	CreateFile FileOperationType = iota
	UpdateFile
	DeleteFile
	MoveFile
	CopyFile
	ReadFile
	AppendFile
)

// OperationStatus defines the status of file operations
type OperationStatus int

const (
	Pending OperationStatus = iota
	InProgress
	Completed
	Failed
	Cancelled
)

// FileLockManager manages file locking across agents
type FileLockManager struct {
	locks map[string]*FileLock
	mu    sync.RWMutex
}

// FileLock represents a file lock
type FileLock struct {
	Path      string
	AgentID   string
	LockType  LockType
	Timestamp time.Time
}

// LockType defines types of file locks
type LockType int

const (
	ReadLock LockType = iota
	WriteLock
	ExclusiveLock
)

// FileChangeTracker tracks file changes and conflicts
type FileChangeTracker struct {
	changes map[string][]*FileChange
	mu      sync.RWMutex
}

// FileChange represents a change to a file
type FileChange struct {
	Path      string
	AgentID   string
	Operation FileOperationType
	Timestamp time.Time
	Checksum  string
	Size      int64
}

// BackupManager manages file backups during operations
type BackupManager struct {
	backupDir string
	backups   map[string]*FileBackup
	mu        sync.RWMutex
}

// FileBackup represents a file backup
type FileBackup struct {
	OriginalPath string
	BackupPath   string
	Timestamp    time.Time
	Size         int64
	Checksum     string
}

// NewFileOperationCoordinator creates a new file operation coordinator
func NewFileOperationCoordinator(workDir string) (*FileOperationCoordinator, error) {
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	backupDir := filepath.Join(workDir, ".backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &FileOperationCoordinator{
		workDir:       workDir,
		lockManager:   NewFileLockManager(),
		changeTracker: NewFileChangeTracker(),
		backupManager: NewBackupManager(backupDir),
		agents:        make(map[string]*FileOperationAgent),
		operations:    make([]*FileOperation, 0),
	}, nil
}

// NewFileLockManager creates a new file lock manager
func NewFileLockManager() *FileLockManager {
	return &FileLockManager{
		locks: make(map[string]*FileLock),
	}
}

// NewFileChangeTracker creates a new file change tracker
func NewFileChangeTracker() *FileChangeTracker {
	return &FileChangeTracker{
		changes: make(map[string][]*FileChange),
	}
}

// NewBackupManager creates a new backup manager
func NewBackupManager(backupDir string) *BackupManager {
	return &BackupManager{
		backupDir: backupDir,
		backups:   make(map[string]*FileBackup),
	}
}

// RegisterAgent registers a new file operation agent
func (foc *FileOperationCoordinator) RegisterAgent(agentID string) *FileOperationAgent {
	foc.mu.Lock()
	defer foc.mu.Unlock()

	agent := &FileOperationAgent{
		ID:               agentID,
		WorkingDirectory: foc.workDir,
		ActiveOperations: make(map[string]*FileOperation),
		LastActivity:     time.Now(),
	}

	foc.agents[agentID] = agent
	return agent
}

// ExecuteOperation executes a coordinated file operation
func (foc *FileOperationCoordinator) ExecuteOperation(ctx context.Context, op *FileOperation) error {
	// Acquire lock
	if err := foc.lockManager.AcquireLock(op.Path, op.AgentID, WriteLock); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer foc.lockManager.ReleaseLock(op.Path, op.AgentID)

	// Create backup if needed
	if op.Type == UpdateFile || op.Type == DeleteFile {
		backup, err := foc.backupManager.CreateBackup(op.Path)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		op.BackupPath = backup.BackupPath
	}

	// Execute operation
	op.Status = InProgress
	op.StartTime = time.Now()

	err := foc.performOperation(ctx, op)

	op.EndTime = time.Now()
	if err != nil {
		op.Status = Failed
		op.Error = err
		// Restore from backup if needed
		if op.BackupPath != "" {
			foc.backupManager.RestoreBackup(op.BackupPath, op.Path)
		}
		return err
	}

	op.Status = Completed

	// Track change
	change := &FileChange{
		Path:      op.Path,
		AgentID:   op.AgentID,
		Operation: op.Type,
		Timestamp: time.Now(),
		Checksum:  op.Checksum,
	}
	foc.changeTracker.TrackChange(change)

	return nil
}

// performOperation performs the actual file operation
func (foc *FileOperationCoordinator) performOperation(ctx context.Context, op *FileOperation) error {
	switch op.Type {
	case CreateFile:
		return foc.createFile(op)
	case UpdateFile:
		return foc.updateFile(op)
	case DeleteFile:
		return foc.deleteFile(op)
	case ReadFile:
		return foc.readFile(op)
	case AppendFile:
		return foc.appendFile(op)
	case CopyFile:
		return foc.copyFile(op)
	case MoveFile:
		return foc.moveFile(op)
	default:
		return fmt.Errorf("unsupported operation type: %v", op.Type)
	}
}

// createFile creates a new file
func (foc *FileOperationCoordinator) createFile(op *FileOperation) error {
	fullPath := filepath.Join(foc.workDir, op.Path)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := ioutil.WriteFile(fullPath, op.Content, 0644); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	op.Checksum = calculateChecksum(op.Content)
	return nil
}

// updateFile updates an existing file
func (foc *FileOperationCoordinator) updateFile(op *FileOperation) error {
	fullPath := filepath.Join(foc.workDir, op.Path)

	if err := ioutil.WriteFile(fullPath, op.Content, 0644); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	op.Checksum = calculateChecksum(op.Content)
	return nil
}

// deleteFile deletes a file
func (foc *FileOperationCoordinator) deleteFile(op *FileOperation) error {
	fullPath := filepath.Join(foc.workDir, op.Path)

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// readFile reads a file
func (foc *FileOperationCoordinator) readFile(op *FileOperation) error {
	fullPath := filepath.Join(foc.workDir, op.Path)

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	op.Content = content
	op.Checksum = calculateChecksum(content)
	return nil
}

// appendFile appends content to a file
func (foc *FileOperationCoordinator) appendFile(op *FileOperation) error {
	fullPath := filepath.Join(foc.workDir, op.Path)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for append: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(op.Content); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}

	return nil
}

// copyFile copies a file
func (foc *FileOperationCoordinator) copyFile(op *FileOperation) error {
	srcPath := filepath.Join(foc.workDir, op.Path)
	dstPath := filepath.Join(foc.workDir, string(op.Content)) // Content contains destination path

	content, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := ioutil.WriteFile(dstPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	op.Checksum = calculateChecksum(content)
	return nil
}

// moveFile moves a file
func (foc *FileOperationCoordinator) moveFile(op *FileOperation) error {
	srcPath := filepath.Join(foc.workDir, op.Path)
	dstPath := filepath.Join(foc.workDir, string(op.Content)) // Content contains destination path

	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// AcquireLock acquires a file lock
func (flm *FileLockManager) AcquireLock(path, agentID string, lockType LockType) error {
	flm.mu.Lock()
	defer flm.mu.Unlock()

	if existingLock, exists := flm.locks[path]; exists {
		// Check if lock is compatible
		if lockType == WriteLock || existingLock.LockType == WriteLock {
			return fmt.Errorf("file %s is already locked by agent %s", path, existingLock.AgentID)
		}
		if lockType == ReadLock && existingLock.LockType == ReadLock {
			// Multiple read locks are allowed
		} else {
			return fmt.Errorf("incompatible lock requested for file %s", path)
		}
	}

	lock := &FileLock{
		Path:      path,
		AgentID:   agentID,
		LockType:  lockType,
		Timestamp: time.Now(),
	}

	flm.locks[path] = lock
	return nil
}

// ReleaseLock releases a file lock
func (flm *FileLockManager) ReleaseLock(path, agentID string) {
	flm.mu.Lock()
	defer flm.mu.Unlock()

	if lock, exists := flm.locks[path]; exists && lock.AgentID == agentID {
		delete(flm.locks, path)
	}
}

// TrackChange tracks a file change
func (fct *FileChangeTracker) TrackChange(change *FileChange) {
	fct.mu.Lock()
	defer fct.mu.Unlock()

	if fct.changes[change.Path] == nil {
		fct.changes[change.Path] = make([]*FileChange, 0)
	}

	fct.changes[change.Path] = append(fct.changes[change.Path], change)
}

// GetChanges returns all changes for a file
func (fct *FileChangeTracker) GetChanges(path string) []*FileChange {
	fct.mu.RLock()
	defer fct.mu.RUnlock()

	changes := fct.changes[path]
	result := make([]*FileChange, len(changes))
	copy(result, changes)
	return result
}

// DetectConflicts detects potential conflicts in file operations
func (fct *FileChangeTracker) DetectConflicts(path string, agentID string) bool {
	fct.mu.RLock()
	defer fct.mu.RUnlock()

	changes := fct.changes[path]
	if len(changes) < 2 {
		return false
	}

	// Check for concurrent modifications
	recentChanges := 0
	cutoff := time.Now().Add(-5 * time.Second) // Last 5 seconds

	for _, change := range changes {
		if change.Timestamp.After(cutoff) && change.AgentID != agentID {
			recentChanges++
		}
	}

	return recentChanges > 0
}

// CreateBackup creates a backup of a file
func (bm *BackupManager) CreateBackup(filePath string) (*FileBackup, error) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, no backup needed
		return &FileBackup{
			OriginalPath: filePath,
			BackupPath:   "",
			Timestamp:    time.Now(),
		}, nil
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for backup: %w", err)
	}

	backupName := fmt.Sprintf("%s_%d.backup", filepath.Base(filePath), time.Now().Unix())
	backupPath := filepath.Join(bm.backupDir, backupName)

	if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}

	backup := &FileBackup{
		OriginalPath: filePath,
		BackupPath:   backupPath,
		Timestamp:    time.Now(),
		Size:         int64(len(content)),
		Checksum:     calculateChecksum(content),
	}

	bm.backups[filePath] = backup
	return backup, nil
}

// RestoreBackup restores a file from backup
func (bm *BackupManager) RestoreBackup(backupPath, originalPath string) error {
	content, err := ioutil.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := ioutil.WriteFile(originalPath, content, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	return nil
}

// Test functions

// TestFileOperationCoordination tests coordinated file operations
func TestFileOperationCoordination(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "swarm_file_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	coordinator, err := NewFileOperationCoordinator(tempDir)
	require.NoError(t, err)

	t.Run("Single Agent File Operations", func(t *testing.T) {
		agent := coordinator.RegisterAgent("agent_1")
		assert.Equal(t, "agent_1", agent.ID)

		// Test file creation
		createOp := &FileOperation{
			ID:      "create_test_1",
			Type:    CreateFile,
			Path:    "test1.txt",
			Content: []byte("Hello, World!"),
			AgentID: agent.ID,
			Status:  Pending,
		}

		err := coordinator.ExecuteOperation(context.Background(), createOp)
		require.NoError(t, err)
		assert.Equal(t, Completed, createOp.Status)

		// Verify file exists
		fullPath := filepath.Join(tempDir, "test1.txt")
		content, err := ioutil.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(content))
	})

	t.Run("Multi-Agent Coordination", func(t *testing.T) {
		agent1 := coordinator.RegisterAgent("agent_1")
		agent2 := coordinator.RegisterAgent("agent_2")

		var wg sync.WaitGroup
		var errors []error
		var errorMu sync.Mutex

		// Agent 1 creates file
		wg.Add(1)
		go func() {
			defer wg.Done()
			op := &FileOperation{
				ID:      "create_test_2",
				Type:    CreateFile,
				Path:    "shared_file.txt",
				Content: []byte("Initial content"),
				AgentID: agent1.ID,
				Status:  Pending,
			}
			if err := coordinator.ExecuteOperation(context.Background(), op); err != nil {
				errorMu.Lock()
				errors = append(errors, err)
				errorMu.Unlock()
			}
		}()

		// Agent 2 tries to update the same file (should be blocked until agent 1 completes)
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // Slight delay
			op := &FileOperation{
				ID:      "update_test_2",
				Type:    UpdateFile,
				Path:    "shared_file.txt",
				Content: []byte("Updated content"),
				AgentID: agent2.ID,
				Status:  Pending,
			}
			if err := coordinator.ExecuteOperation(context.Background(), op); err != nil {
				errorMu.Lock()
				errors = append(errors, err)
				errorMu.Unlock()
			}
		}()

		wg.Wait()

		// Should have no errors - operations should be serialized
		assert.Empty(t, errors)

		// Final content should be from the last operation
		fullPath := filepath.Join(tempDir, "shared_file.txt")
		content, err := ioutil.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, "Updated content", string(content))
	})

	t.Run("Conflict Detection", func(t *testing.T) {
		agent1 := coordinator.RegisterAgent("conflict_agent_1")
		agent2 := coordinator.RegisterAgent("conflict_agent_2")

		// Both agents try to create the same file
		op1 := &FileOperation{
			ID:      "conflict_test_1",
			Type:    CreateFile,
			Path:    "conflict_file.txt",
			Content: []byte("Agent 1 content"),
			AgentID: agent1.ID,
			Status:  Pending,
		}

		op2 := &FileOperation{
			ID:      "conflict_test_2",
			Type:    CreateFile,
			Path:    "conflict_file.txt",
			Content: []byte("Agent 2 content"),
			AgentID: agent2.ID,
			Status:  Pending,
		}

		// Execute first operation
		err1 := coordinator.ExecuteOperation(context.Background(), op1)
		require.NoError(t, err1)

		// Simulate concurrent access
		time.Sleep(10 * time.Millisecond)

		// Check for conflicts
		hasConflict := coordinator.changeTracker.DetectConflicts("conflict_file.txt", agent2.ID)
		assert.True(t, hasConflict, "Should detect conflict for concurrent file operations")

		// Second operation should still succeed (it will overwrite)
		err2 := coordinator.ExecuteOperation(context.Background(), op2)
		require.NoError(t, err2)
	})

	t.Run("Backup and Recovery", func(t *testing.T) {
		agent := coordinator.RegisterAgent("backup_agent")

		// Create initial file
		createOp := &FileOperation{
			ID:      "backup_create",
			Type:    CreateFile,
			Path:    "backup_test.txt",
			Content: []byte("Original content"),
			AgentID: agent.ID,
			Status:  Pending,
		}

		err := coordinator.ExecuteOperation(context.Background(), createOp)
		require.NoError(t, err)

		// Update file (should create backup)
		updateOp := &FileOperation{
			ID:      "backup_update",
			Type:    UpdateFile,
			Path:    "backup_test.txt",
			Content: []byte("Updated content"),
			AgentID: agent.ID,
			Status:  Pending,
		}

		err = coordinator.ExecuteOperation(context.Background(), updateOp)
		require.NoError(t, err)
		assert.NotEmpty(t, updateOp.BackupPath, "Backup should be created for update operation")

		// Verify backup exists
		_, err = os.Stat(updateOp.BackupPath)
		assert.NoError(t, err, "Backup file should exist")

		// Verify updated content
		fullPath := filepath.Join(tempDir, "backup_test.txt")
		content, err := ioutil.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, "Updated content", string(content))
	})
}

// TestFileOperationPerformance tests performance characteristics
func TestFileOperationPerformance(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "swarm_perf_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	coordinator, err := NewFileOperationCoordinator(tempDir)
	require.NoError(t, err)

	t.Run("High Concurrency Operations", func(t *testing.T) {
		numAgents := 10
		operationsPerAgent := 50

		agents := make([]*FileOperationAgent, numAgents)
		for i := 0; i < numAgents; i++ {
			agents[i] = coordinator.RegisterAgent(fmt.Sprintf("perf_agent_%d", i))
		}

		startTime := time.Now()
		var wg sync.WaitGroup
		var errorMu sync.Mutex
		var errors []error

		for i, agent := range agents {
			wg.Add(1)
			go func(agentIndex int, ag *FileOperationAgent) {
				defer wg.Done()

				for j := 0; j < operationsPerAgent; j++ {
					op := &FileOperation{
						ID:      fmt.Sprintf("perf_op_%d_%d", agentIndex, j),
						Type:    CreateFile,
						Path:    fmt.Sprintf("perf_file_%d_%d.txt", agentIndex, j),
						Content: []byte(fmt.Sprintf("Content from agent %d, operation %d", agentIndex, j)),
						AgentID: ag.ID,
						Status:  Pending,
					}

					if err := coordinator.ExecuteOperation(context.Background(), op); err != nil {
						errorMu.Lock()
						errors = append(errors, err)
						errorMu.Unlock()
					}
				}
			}(i, agent)
		}

		wg.Wait()
		duration := time.Since(startTime)

		totalOperations := numAgents * operationsPerAgent
		opsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("Completed %d operations in %v", totalOperations, duration)
		t.Logf("Performance: %.2f ops/second", opsPerSecond)

		assert.Empty(t, errors, "Should have no errors in performance test")
		assert.Greater(t, opsPerSecond, 100.0, "Should achieve at least 100 ops/second")
	})

	t.Run("Memory Usage Under Load", func(t *testing.T) {
		var memBefore runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memBefore)

		// Create many operations
		numOperations := 1000
		agent := coordinator.RegisterAgent("memory_test_agent")

		operations := make([]*FileOperation, numOperations)
		for i := 0; i < numOperations; i++ {
			operations[i] = &FileOperation{
				ID:      fmt.Sprintf("mem_op_%d", i),
				Type:    CreateFile,
				Path:    fmt.Sprintf("memory_test_%d.txt", i),
				Content: make([]byte, 1024), // 1KB each
				AgentID: agent.ID,
				Status:  Pending,
			}
		}

		// Execute operations
		for _, op := range operations {
			err := coordinator.ExecuteOperation(context.Background(), op)
			require.NoError(t, err)
		}

		var memAfter runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memAfter)

		memUsedMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
		t.Logf("Memory used: %.2f MB for %d operations", memUsedMB, numOperations)

		// Should not use more than 50MB for 1000 operations of 1KB each
		assert.Less(t, memUsedMB, 50.0, "Memory usage should be reasonable")
	})
}

// Utility functions

// calculateChecksum calculates a simple checksum for content
func calculateChecksum(content []byte) string {
	hash := 0
	for _, b := range content {
		hash = hash*31 + int(b)
	}
	return fmt.Sprintf("%x", hash)
}
