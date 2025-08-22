package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// HealthCheck performs a health check on local storage
func (ls *LocalStorage) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	checks := make(map[string]CheckResult)
	healthy := true

	// Check disk space
	diskCheck := ls.checkDiskSpace()
	checks["disk_space"] = diskCheck
	if diskCheck.Status != "healthy" {
		healthy = false
	}

	// Check write permissions
	writeCheck := ls.checkWritePermissions()
	checks["write_permissions"] = writeCheck
	if writeCheck.Status != "healthy" {
		healthy = false
	}

	// Check read permissions
	readCheck := ls.checkReadPermissions()
	checks["read_permissions"] = readCheck
	if readCheck.Status != "healthy" {
		healthy = false
	}

	// Check directory structure
	dirCheck := ls.checkDirectoryStructure()
	checks["directory_structure"] = dirCheck
	if dirCheck.Status != "healthy" {
		healthy = false
	}

	// Check metadata consistency
	metaCheck := ls.checkMetadataConsistency()
	checks["metadata_consistency"] = metaCheck
	if metaCheck.Status != "healthy" {
		healthy = false
	}

	ls.lastHealthCheck = time.Now()
	ls.healthy = healthy

	return &HealthStatus{
		Healthy:   healthy,
		LastCheck: ls.lastHealthCheck,
		Checks:    checks,
	}, nil
}

// checkDiskSpace checks available disk space
func (ls *LocalStorage) checkDiskSpace() CheckResult {
	stat, err := ls.getDiskUsage(ls.basePath)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Failed to get disk usage: %v", err),
			Time:    time.Now(),
		}
	}

	usagePercent := float64(stat.Used) / float64(stat.Total) * 100
	availableGB := float64(stat.Available) / (1024 * 1024 * 1024)

	if usagePercent > 95 {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Disk usage critical: %.1f%% used, %.1f GB available", usagePercent, availableGB),
			Time:    time.Now(),
		}
	}

	if usagePercent > 85 {
		return CheckResult{
			Status:  "warning",
			Message: fmt.Sprintf("Disk usage warning: %.1f%% used, %.1f GB available", usagePercent, availableGB),
			Time:    time.Now(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Disk usage normal: %.1f%% used, %.1f GB available", usagePercent, availableGB),
		Time:    time.Now(),
	}
}

// checkWritePermissions checks write permissions
func (ls *LocalStorage) checkWritePermissions() CheckResult {
	testFile := filepath.Join(ls.basePath, ".write_test")

	// Try to create a test file
	file, err := os.Create(testFile)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Cannot write to storage directory: %v", err),
			Time:    time.Now(),
		}
	}
	file.Close()

	// Clean up test file
	os.Remove(testFile)

	return CheckResult{
		Status:  "healthy",
		Message: "Write permissions OK",
		Time:    time.Now(),
	}
}

// checkReadPermissions checks read permissions
func (ls *LocalStorage) checkReadPermissions() CheckResult {
	// Try to read the base directory
	_, err := os.ReadDir(ls.basePath)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Cannot read storage directory: %v", err),
			Time:    time.Now(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Read permissions OK",
		Time:    time.Now(),
	}
}

// checkDirectoryStructure checks directory structure integrity
func (ls *LocalStorage) checkDirectoryStructure() CheckResult {
	// Check if base path exists and is a directory
	info, err := os.Stat(ls.basePath)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Base path not accessible: %v", err),
			Time:    time.Now(),
		}
	}

	if !info.IsDir() {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Base path is not a directory",
			Time:    time.Now(),
		}
	}

	// Check metadata directory
	info, err = os.Stat(ls.metaPath)
	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Metadata path not accessible: %v", err),
			Time:    time.Now(),
		}
	}

	if !info.IsDir() {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Metadata path is not a directory",
			Time:    time.Now(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Directory structure OK",
		Time:    time.Now(),
	}
}

// checkMetadataConsistency checks metadata consistency
func (ls *LocalStorage) checkMetadataConsistency() CheckResult {
	inconsistencies := 0
	totalChecked := 0

	// Walk through objects and check metadata
	err := filepath.Walk(ls.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and metadata directory
		if info.IsDir() || filepath.Dir(path) == ls.metaPath {
			return nil
		}

		// Get relative path as key
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(relPath)
		totalChecked++

		// Check if metadata exists and is consistent
		metadata, err := ls.getMetadata(key)
		if err != nil {
			inconsistencies++
			return nil
		}

		// Check size consistency
		if metadata.Size != info.Size() {
			inconsistencies++
		}

		return nil
	})

	if err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Failed to check metadata consistency: %v", err),
			Time:    time.Now(),
		}
	}

	if inconsistencies > 0 {
		return CheckResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Found %d metadata inconsistencies out of %d objects", inconsistencies, totalChecked),
			Time:    time.Now(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Metadata consistency OK (%d objects checked)", totalChecked),
		Time:    time.Now(),
	}
}

// GetStats returns storage statistics
func (ls *LocalStorage) GetStats(ctx context.Context) (*StorageStats, error) {
	ls.statsMutex.RLock()
	defer ls.statsMutex.RUnlock()

	// Update disk usage
	if diskStat, err := ls.getDiskUsage(ls.basePath); err == nil {
		ls.stats.UsedSpace = diskStat.Used
		ls.stats.AvailableSpace = diskStat.Available
	}

	// Create a copy of stats to return
	statsCopy := *ls.stats
	return &statsCopy, nil
}

// getDiskUsage returns disk usage statistics
func (ls *LocalStorage) getDiskUsage(path string) (*diskStat, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available

	return &diskStat{
		Total:     total,
		Used:      used,
		Available: available,
	}, nil
}

// diskStat represents disk usage statistics
type diskStat struct {
	Total     int64
	Used      int64
	Available int64
}

// updateLatencyStats updates latency statistics
func (ls *LocalStorage) updateLatencyStats(operation string, duration time.Duration) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	switch operation {
	case "read":
		if ls.stats.Performance.ReadLatency != nil {
			ls.stats.Performance.ReadLatency.Mean = float64(duration.Milliseconds())
		}
	case "write":
		if ls.stats.Performance.WriteLatency != nil {
			ls.stats.Performance.WriteLatency.Mean = float64(duration.Milliseconds())
		}
	case "delete":
		if ls.stats.Performance.DeleteLatency != nil {
			ls.stats.Performance.DeleteLatency.Mean = float64(duration.Milliseconds())
		}
	}
}

// incrementOperationCount increments operation count
func (ls *LocalStorage) incrementOperationCount(operation string) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	if ls.stats.OperationCounts == nil {
		ls.stats.OperationCounts = make(map[string]int64)
	}

	ls.stats.OperationCounts[operation]++
}

// incrementStorageSize increments storage size
func (ls *LocalStorage) incrementStorageSize(size int64) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	ls.stats.TotalSize += size
	ls.stats.TotalObjects++
}

// decrementStorageSize decrements storage size
func (ls *LocalStorage) decrementStorageSize(size int64) {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	ls.stats.TotalSize -= size
	if ls.stats.TotalObjects > 0 {
		ls.stats.TotalObjects--
	}
}

// backgroundHealthCheck performs periodic health checks
func (ls *LocalStorage) backgroundHealthCheck() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ls.ctx.Done():
			return
		case <-ticker.C:
			if _, err := ls.HealthCheck(ls.ctx); err != nil {
				ls.logger.Error("health check failed", "error", err)
			}
		}
	}
}

// updateThroughputStats updates throughput statistics
func (ls *LocalStorage) updateThroughputStats() {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	// Calculate operations per second (simple moving average over 1 minute)
	if ls.stats.OperationCounts != nil {
		if readOps, exists := ls.stats.OperationCounts["retrieve"]; exists {
			ls.stats.Performance.Throughput.ReadOpsPerSec = float64(readOps) / 60.0
		}
		if writeOps, exists := ls.stats.OperationCounts["store"]; exists {
			ls.stats.Performance.Throughput.WriteOpsPerSec = float64(writeOps) / 60.0
		}
		if deleteOps, exists := ls.stats.OperationCounts["delete"]; exists {
			ls.stats.Performance.Throughput.DeleteOpsPerSec = float64(deleteOps) / 60.0
		}
	}
}

// GetHealthStatus returns the current health status
func (ls *LocalStorage) GetHealthStatus() bool {
	return ls.healthy
}

// GetLastHealthCheck returns the time of the last health check
func (ls *LocalStorage) GetLastHealthCheck() time.Time {
	return ls.lastHealthCheck
}

// ResetStats resets storage statistics
func (ls *LocalStorage) ResetStats() {
	ls.statsMutex.Lock()
	defer ls.statsMutex.Unlock()

	ls.stats = NewStorageStats()
}

// GetOperationCounts returns operation counts
func (ls *LocalStorage) GetOperationCounts() map[string]int64 {
	ls.statsMutex.RLock()
	defer ls.statsMutex.RUnlock()

	counts := make(map[string]int64)
	for k, v := range ls.stats.OperationCounts {
		counts[k] = v
	}
	return counts
}
