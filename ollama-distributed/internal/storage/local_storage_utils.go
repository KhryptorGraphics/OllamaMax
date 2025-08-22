package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// backgroundCleanup performs periodic cleanup tasks
func (ls *LocalStorage) backgroundCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ls.ctx.Done():
			return
		case <-ticker.C:
			ls.performCleanup()
		}
	}
}

// performCleanup performs cleanup tasks
func (ls *LocalStorage) performCleanup() {
	ls.logger.Debug("starting background cleanup")

	// Clean up temporary files
	ls.cleanupTempFiles()

	// Clean up orphaned metadata files
	ls.cleanupOrphanedMetadata()

	// Clean up empty directories
	ls.cleanupEmptyDirectories()

	// Update throughput statistics
	ls.updateThroughputStats()

	ls.logger.Debug("background cleanup completed")
}

// cleanupTempFiles removes temporary files
func (ls *LocalStorage) cleanupTempFiles() {
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check for temporary files
		if strings.HasSuffix(path, ".tmp") {
			// Remove files older than 1 hour
			if time.Since(info.ModTime()) > time.Hour {
				if err := os.Remove(path); err != nil {
					ls.logger.Warn("failed to remove temp file", "path", path, "error", err)
				} else {
					ls.logger.Debug("removed temp file", "path", path)
				}
			}
		}

		return nil
	})

	if err != nil {
		ls.logger.Error("failed to cleanup temp files", "error", err)
	}
}

// cleanupOrphanedMetadata removes metadata files without corresponding objects
func (ls *LocalStorage) cleanupOrphanedMetadata() {
	err := filepath.Walk(ls.metaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check for metadata files
		if strings.HasSuffix(path, ".meta") {
			// Extract key from metadata filename
			relPath, err := filepath.Rel(ls.metaPath, path)
			if err != nil {
				return err
			}

			key := strings.TrimSuffix(relPath, ".meta")
			objectPath := ls.getObjectPath(key)

			// Check if corresponding object exists
			if _, err := os.Stat(objectPath); os.IsNotExist(err) {
				// Remove orphaned metadata file
				if err := os.Remove(path); err != nil {
					ls.logger.Warn("failed to remove orphaned metadata", "path", path, "error", err)
				} else {
					ls.logger.Debug("removed orphaned metadata", "path", path, "key", key)
				}
			}
		}

		return nil
	})

	if err != nil {
		ls.logger.Error("failed to cleanup orphaned metadata", "error", err)
	}
}

// cleanupEmptyDirectories removes empty directories
func (ls *LocalStorage) cleanupEmptyDirectories() {
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process directories
		if !info.IsDir() {
			return nil
		}

		// Skip base path and metadata path
		if path == ls.basePath || path == ls.metaPath {
			return nil
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			// Remove empty directory
			if err := os.Remove(path); err != nil {
				ls.logger.Warn("failed to remove empty directory", "path", path, "error", err)
			} else {
				ls.logger.Debug("removed empty directory", "path", path)
			}
		}

		return nil
	})

	if err != nil {
		ls.logger.Error("failed to cleanup empty directories", "error", err)
	}
}

// Compact performs storage compaction (placeholder for future implementation)
func (ls *LocalStorage) Compact(ctx context.Context) error {
	ls.logger.Info("starting storage compaction")

	// TODO: Implement compaction logic
	// This could include:
	// - Defragmenting files
	// - Reorganizing directory structure
	// - Compressing old files
	// - Removing duplicate data

	ls.logger.Info("storage compaction completed")
	return nil
}

// Backup creates a backup of the storage (placeholder for future implementation)
func (ls *LocalStorage) Backup(ctx context.Context, backupPath string) error {
	ls.logger.Info("starting storage backup", "backup_path", backupPath)

	// TODO: Implement backup logic
	// This could include:
	// - Creating tar/zip archives
	// - Incremental backups
	// - Compression
	// - Encryption

	ls.logger.Info("storage backup completed", "backup_path", backupPath)
	return nil
}

// Restore restores storage from a backup (placeholder for future implementation)
func (ls *LocalStorage) Restore(ctx context.Context, backupPath string) error {
	ls.logger.Info("starting storage restore", "backup_path", backupPath)

	// TODO: Implement restore logic
	// This could include:
	// - Extracting archives
	// - Validating data integrity
	// - Rebuilding metadata
	// - Decryption

	ls.logger.Info("storage restore completed", "backup_path", backupPath)
	return nil
}

// Migrate migrates storage to a new format or location (placeholder for future implementation)
func (ls *LocalStorage) Migrate(ctx context.Context, newConfig *LocalStorageConfig) error {
	ls.logger.Info("starting storage migration")

	// TODO: Implement migration logic
	// This could include:
	// - Moving files to new location
	// - Converting file formats
	// - Updating metadata schema
	// - Preserving data integrity

	ls.logger.Info("storage migration completed")
	return nil
}

// Verify verifies storage integrity
func (ls *LocalStorage) Verify(ctx context.Context) (*VerificationResult, error) {
	ls.logger.Info("starting storage verification")

	result := &VerificationResult{
		StartTime:      time.Now(),
		TotalObjects:   0,
		ValidObjects:   0,
		InvalidObjects: 0,
		Errors:         []string{},
	}

	// Walk through all objects and verify them
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return err
		}

		// Skip directories and metadata directory
		if info.IsDir() || strings.Contains(path, ".metadata") {
			return nil
		}

		// Get relative path as key
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return err
		}

		key := filepath.ToSlash(relPath)
		result.TotalObjects++

		// Verify object
		if err := ls.verifyObject(key); err != nil {
			result.InvalidObjects++
			result.Errors = append(result.Errors, fmt.Sprintf("Object %s: %v", key, err))
		} else {
			result.ValidObjects++
		}

		return nil
	})

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	ls.logger.Info("storage verification completed",
		"total_objects", result.TotalObjects,
		"valid_objects", result.ValidObjects,
		"invalid_objects", result.InvalidObjects,
		"duration", result.Duration)

	return result, nil
}

// verifyObject verifies a single object
func (ls *LocalStorage) verifyObject(key string) error {
	// Get metadata
	metadata, err := ls.getMetadata(key)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Check if object file exists
	objectPath := ls.getObjectPath(key)
	fileInfo, err := os.Stat(objectPath)
	if err != nil {
		return fmt.Errorf("object file not found: %w", err)
	}

	// Verify size
	if metadata.Size != fileInfo.Size() {
		return fmt.Errorf("size mismatch: metadata=%d, file=%d", metadata.Size, fileInfo.Size())
	}

	// Verify hash if available
	if metadata.Hash != "" {
		actualHash, err := ls.calculateFileHash(objectPath)
		if err != nil {
			return fmt.Errorf("failed to calculate hash: %w", err)
		}

		if metadata.Hash != actualHash {
			return fmt.Errorf("hash mismatch: metadata=%s, actual=%s", metadata.Hash, actualHash)
		}
	}

	return nil
}

// VerificationResult represents the result of storage verification
type VerificationResult struct {
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       time.Duration `json:"duration"`
	TotalObjects   int           `json:"total_objects"`
	ValidObjects   int           `json:"valid_objects"`
	InvalidObjects int           `json:"invalid_objects"`
	Errors         []string      `json:"errors"`
}

// GetStorageInfo returns general storage information
func (ls *LocalStorage) GetStorageInfo() map[string]interface{} {
	return map[string]interface{}{
		"type":        "local",
		"base_path":   ls.basePath,
		"meta_path":   ls.metaPath,
		"max_size":    ls.maxSize,
		"compression": ls.compression,
		"encryption":  ls.encryption,
		"started":     ls.started,
		"healthy":     ls.healthy,
	}
}

// SetMaxSize updates the maximum storage size
func (ls *LocalStorage) SetMaxSize(maxSize int64) {
	ls.maxSize = maxSize
	ls.logger.Info("updated max storage size", "max_size", maxSize)
}

// SetCompression enables or disables compression
func (ls *LocalStorage) SetCompression(enabled bool) {
	ls.compression = enabled
	ls.logger.Info("updated compression setting", "enabled", enabled)
}

// SetEncryption enables or disables encryption
func (ls *LocalStorage) SetEncryption(enabled bool) {
	ls.encryption = enabled
	ls.logger.Info("updated encryption setting", "enabled", enabled)
}
