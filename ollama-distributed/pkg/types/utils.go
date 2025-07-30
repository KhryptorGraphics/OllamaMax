package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Utility functions for type operations

// ID generation utilities

// GenerateNodeID generates a unique node ID
func GenerateNodeID() NodeID {
	return NodeID(generateID("node"))
}

// GenerateTaskID generates a unique task ID
func GenerateTaskID() TaskID {
	return TaskID(generateID("task"))
}

// GenerateModelID generates a unique model ID
func GenerateModelID() ModelID {
	return ModelID(generateID("model"))
}

// GenerateClusterID generates a unique cluster ID
func GenerateClusterID() ClusterID {
	return ClusterID(generateID("cluster"))
}

// generateID generates a unique ID with a prefix
func generateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixNano(), hex.EncodeToString(bytes))
}

// Validation utilities

// ValidateNodeID validates a node ID format
func ValidateNodeID(id NodeID) bool {
	return validateID(string(id), "node")
}

// ValidateTaskID validates a task ID format
func ValidateTaskID(id TaskID) bool {
	return validateID(string(id), "task")
}

// ValidateModelID validates a model ID format
func ValidateModelID(id ModelID) bool {
	return validateID(string(id), "model")
}

// ValidateClusterID validates a cluster ID format
func ValidateClusterID(id ClusterID) bool {
	return validateID(string(id), "cluster")
}

// validateID validates an ID format
func validateID(id, prefix string) bool {
	if id == "" {
		return false
	}
	pattern := fmt.Sprintf(`^%s-\d+-[a-f0-9]{16}$`, prefix)
	matched, _ := regexp.MatchString(pattern, id)
	return matched
}

// ValidateNodeStatus validates a node status
func ValidateNodeStatus(status NodeStatus) bool {
	switch status {
	case NodeStatusOnline, NodeStatusOffline, NodeStatusDraining, NodeStatusMaintenance:
		return true
	default:
		return false
	}
}

// ValidateTaskStatus validates a task status
func ValidateTaskStatus(status TaskStatus) bool {
	switch status {
	case TaskStatusPending, TaskStatusRunning, TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// ValidateModelStatus validates a model status
func ValidateModelStatus(status ModelStatus) bool {
	switch status {
	case ModelStatusAvailable, ModelStatusLoading, ModelStatusUnavailable, ModelStatusError:
		return true
	default:
		return false
	}
}

// ValidateClusterStatus validates a cluster status
func ValidateClusterStatus(status ClusterStatus) bool {
	switch status {
	case ClusterStatusHealthy, ClusterStatusDegraded, ClusterStatusUnavailable:
		return true
	default:
		return false
	}
}

// Conversion utilities

// NodeToNodeInfo converts a Node to a simplified NodeInfo structure
func NodeToNodeInfo(node *Node) *NodeInfo {
	if node == nil {
		return nil
	}
	
	return &NodeInfo{
		ID:       string(node.ID),
		Address:  node.Address,
		Status:   convertNodeStatus(node.Status),
		Capacity: convertNodeCapacity(node.Capabilities),
		Usage:    convertNodeUsage(node.Metrics),
		Models:   []string{}, // TODO: Extract from node metadata
		LastSeen: node.LastSeen,
		Metadata: convertMetadata(node.Metadata),
	}
}

// convertNodeStatus converts NodeStatus to the legacy format
func convertNodeStatus(status NodeStatus) NodeStatus {
	// For now, return as-is since we're using the same enum
	return status
}

// convertNodeCapacity converts NodeCapabilities to NodeCapacity
func convertNodeCapacity(capabilities *NodeCapabilities) NodeCapacity {
	if capabilities == nil {
		return NodeCapacity{}
	}
	
	return NodeCapacity{
		CPU:    int64(capabilities.Hardware.CPU),
		Memory: capabilities.Hardware.Memory,
		GPU:    capabilities.Hardware.GPUMemory,
		Disk:   capabilities.Hardware.Storage,
	}
}

// convertNodeUsage converts NodeMetrics to NodeUsage
func convertNodeUsage(metrics *NodeMetrics) NodeUsage {
	if metrics == nil {
		return NodeUsage{}
	}
	
	return NodeUsage{
		CPU:    metrics.CPUUsage,
		Memory: metrics.MemoryUsage,
		GPU:    metrics.GPUUsage,
		Disk:   metrics.StorageUsage,
	}
}

// convertMetadata converts map[string]interface{} to map[string]string
func convertMetadata(metadata map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// Resource calculation utilities

// CalculateResourceUtilization calculates resource utilization percentage
func CalculateResourceUtilization(usage, capacity float64) float64 {
	if capacity == 0 {
		return 0
	}
	return (usage / capacity) * 100
}

// HasSufficientResources checks if a node has sufficient resources for requirements
func HasSufficientResources(node *Node, requirements *ResourceRequirements) bool {
	if node == nil || node.Capabilities == nil || node.Metrics == nil || requirements == nil {
		return false
	}
	
	// Check CPU
	availableCPU := float64(node.Capabilities.Hardware.CPU) * (1.0 - node.Metrics.CPUUsage/100.0)
	if availableCPU < requirements.MinCPU {
		return false
	}
	
	// Check Memory
	availableMemory := float64(node.Capabilities.Hardware.Memory) * (1.0 - node.Metrics.MemoryUsage/100.0)
	if availableMemory < float64(requirements.MinMemory) {
		return false
	}
	
	// Check GPU if required
	if requirements.RequiresGPU {
		if node.Capabilities.Hardware.GPU == 0 {
			return false
		}
		availableGPUMemory := float64(node.Capabilities.Hardware.GPUMemory) * (1.0 - node.Metrics.GPUUsage/100.0)
		if availableGPUMemory < float64(requirements.MinGPUMemory) {
			return false
		}
	}
	
	return true
}

// String utilities

// SanitizeModelName sanitizes a model name for use as an identifier
func SanitizeModelName(name string) string {
	// Replace invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_\.]`)
	sanitized := reg.ReplaceAllString(name, "_")
	
	// Remove multiple consecutive underscores
	reg = regexp.MustCompile(`_{2,}`)
	sanitized = reg.ReplaceAllString(sanitized, "_")
	
	// Trim underscores from start and end
	sanitized = strings.Trim(sanitized, "_")
	
	return sanitized
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FormatBytes formats bytes in a human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Time utilities

// IsExpired checks if a timestamp is expired based on a TTL
func IsExpired(timestamp time.Time, ttl time.Duration) bool {
	return time.Since(timestamp) > ttl
}

// GetAge returns the age of a timestamp
func GetAge(timestamp time.Time) time.Duration {
	return time.Since(timestamp)
}

// IsRecent checks if a timestamp is within a recent time window
func IsRecent(timestamp time.Time, window time.Duration) bool {
	return time.Since(timestamp) <= window
}
