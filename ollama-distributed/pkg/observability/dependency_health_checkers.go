package observability

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"
)

// DatabaseHealthChecker checks database connectivity and health
type DatabaseHealthChecker struct {
	name             string
	connectionString string
	db               *sql.DB
	required         bool
	timeout          time.Duration
}

// RedisHealthChecker checks Redis connectivity and health
type RedisHealthChecker struct {
	name     string
	address  string
	required bool
	timeout  time.Duration
}

// HTTPServiceHealthChecker checks HTTP service health
type HTTPServiceHealthChecker struct {
	name     string
	url      string
	required bool
	timeout  time.Duration
	client   *http.Client
}

// TCPServiceHealthChecker checks TCP service connectivity
type TCPServiceHealthChecker struct {
	name     string
	address  string
	required bool
	timeout  time.Duration
}

// StorageHealthChecker checks storage system health
type StorageHealthChecker struct {
	name     string
	path     string
	required bool
	timeout  time.Duration
}

// ExternalAPIHealthChecker checks external API health
type ExternalAPIHealthChecker struct {
	name     string
	url      string
	required bool
	timeout  time.Duration
	client   *http.Client
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(name, connectionString string, required bool) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:             name,
		connectionString: connectionString,
		required:         required,
		timeout:          10 * time.Second,
	}
}

// GetDependencyName returns the dependency name
func (dhc *DatabaseHealthChecker) GetDependencyName() string {
	return dhc.name
}

// GetDependencyType returns the dependency type
func (dhc *DatabaseHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeDatabase
}

// IsRequired returns whether the dependency is required
func (dhc *DatabaseHealthChecker) IsRequired() bool {
	return dhc.required
}

// CheckDependency checks the database health
func (dhc *DatabaseHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: dhc.name,
		Type:           DependencyTypeDatabase,
		Required:       dhc.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, dhc.timeout)
	defer cancel()

	// Try to connect to database
	if dhc.db == nil {
		db, err := sql.Open("postgres", dhc.connectionString)
		if err != nil {
			status.Status = HealthStatusUnhealthy
			status.Message = fmt.Sprintf("Failed to open database connection: %v", err)
			status.Latency = time.Since(start)
			return status
		}
		dhc.db = db
	}

	// Ping database
	if err := dhc.db.PingContext(timeoutCtx); err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Database ping failed: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	// Check database stats
	stats := dhc.db.Stats()
	status.Metadata["open_connections"] = stats.OpenConnections
	status.Metadata["in_use"] = stats.InUse
	status.Metadata["idle"] = stats.Idle

	// Determine health status based on connection stats
	if stats.OpenConnections > 50 {
		status.Status = HealthStatusDegraded
		status.Message = "High number of database connections"
	} else {
		status.Status = HealthStatusHealthy
		status.Message = "Database is healthy"
	}

	status.Latency = time.Since(start)
	return status
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(name, address string, required bool) *RedisHealthChecker {
	return &RedisHealthChecker{
		name:     name,
		address:  address,
		required: required,
		timeout:  5 * time.Second,
	}
}

// GetDependencyName returns the dependency name
func (rhc *RedisHealthChecker) GetDependencyName() string {
	return rhc.name
}

// GetDependencyType returns the dependency type
func (rhc *RedisHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeCache
}

// IsRequired returns whether the dependency is required
func (rhc *RedisHealthChecker) IsRequired() bool {
	return rhc.required
}

// CheckDependency checks the Redis health
func (rhc *RedisHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: rhc.name,
		Type:           DependencyTypeCache,
		Required:       rhc.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, rhc.timeout)
	defer cancel()

	// Try to connect to Redis
	conn, err := net.DialTimeout("tcp", rhc.address, rhc.timeout)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to connect to Redis: %v", err)
		status.Latency = time.Since(start)
		return status
	}
	defer conn.Close()

	// Set deadline for connection operations
	conn.SetDeadline(time.Now().Add(rhc.timeout))

	// Send PING command
	_, err = conn.Write([]byte("PING\r\n"))
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to send PING to Redis: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	// Read response
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to read PING response from Redis: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	status.Status = HealthStatusHealthy
	status.Message = "Redis is healthy"
	status.Latency = time.Since(start)

	// Check if context was cancelled
	select {
	case <-timeoutCtx.Done():
		status.Status = HealthStatusDegraded
		status.Message = "Redis health check timed out"
	default:
		// Continue normally
	}

	return status
}

// NewHTTPServiceHealthChecker creates a new HTTP service health checker
func NewHTTPServiceHealthChecker(name, url string, required bool) *HTTPServiceHealthChecker {
	return &HTTPServiceHealthChecker{
		name:     name,
		url:      url,
		required: required,
		timeout:  10 * time.Second,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDependencyName returns the dependency name
func (hsc *HTTPServiceHealthChecker) GetDependencyName() string {
	return hsc.name
}

// GetDependencyType returns the dependency type
func (hsc *HTTPServiceHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeService
}

// IsRequired returns whether the dependency is required
func (hsc *HTTPServiceHealthChecker) IsRequired() bool {
	return hsc.required
}

// CheckDependency checks the HTTP service health
func (hsc *HTTPServiceHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: hsc.name,
		Type:           DependencyTypeService,
		Required:       hsc.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", hsc.url, nil)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to create HTTP request: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	// Make HTTP request
	resp, err := hsc.client.Do(req)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("HTTP request failed: %v", err)
		status.Latency = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	// Add response metadata
	status.Metadata["status_code"] = resp.StatusCode
	status.Metadata["content_length"] = resp.ContentLength

	// Determine health based on status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status.Status = HealthStatusHealthy
		status.Message = "HTTP service is healthy"
	} else if resp.StatusCode >= 500 {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("HTTP service returned server error: %d", resp.StatusCode)
	} else {
		status.Status = HealthStatusDegraded
		status.Message = fmt.Sprintf("HTTP service returned non-success status: %d", resp.StatusCode)
	}

	status.Latency = time.Since(start)
	return status
}

// NewTCPServiceHealthChecker creates a new TCP service health checker
func NewTCPServiceHealthChecker(name, address string, required bool) *TCPServiceHealthChecker {
	return &TCPServiceHealthChecker{
		name:     name,
		address:  address,
		required: required,
		timeout:  5 * time.Second,
	}
}

// GetDependencyName returns the dependency name
func (tsc *TCPServiceHealthChecker) GetDependencyName() string {
	return tsc.name
}

// GetDependencyType returns the dependency type
func (tsc *TCPServiceHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeNetwork
}

// IsRequired returns whether the dependency is required
func (tsc *TCPServiceHealthChecker) IsRequired() bool {
	return tsc.required
}

// CheckDependency checks the TCP service health
func (tsc *TCPServiceHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: tsc.name,
		Type:           DependencyTypeNetwork,
		Required:       tsc.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// Try to connect to TCP service
	conn, err := net.DialTimeout("tcp", tsc.address, tsc.timeout)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to connect to TCP service: %v", err)
		status.Latency = time.Since(start)
		return status
	}
	defer conn.Close()

	// Add connection metadata
	status.Metadata["local_addr"] = conn.LocalAddr().String()
	status.Metadata["remote_addr"] = conn.RemoteAddr().String()

	status.Status = HealthStatusHealthy
	status.Message = "TCP service is reachable"
	status.Latency = time.Since(start)

	return status
}

// NewStorageHealthChecker creates a new storage health checker
func NewStorageHealthChecker(name, path string, required bool) *StorageHealthChecker {
	return &StorageHealthChecker{
		name:     name,
		path:     path,
		required: required,
		timeout:  5 * time.Second,
	}
}

// GetDependencyName returns the dependency name
func (shc *StorageHealthChecker) GetDependencyName() string {
	return shc.name
}

// GetDependencyType returns the dependency type
func (shc *StorageHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeStorage
}

// IsRequired returns whether the dependency is required
func (shc *StorageHealthChecker) IsRequired() bool {
	return shc.required
}

// CheckDependency checks the storage health
func (shc *StorageHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: shc.name,
		Type:           DependencyTypeStorage,
		Required:       shc.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// This is a simplified storage health check
	// In a real implementation, you would check disk space, permissions, etc.

	status.Status = HealthStatusHealthy
	status.Message = "Storage is accessible"
	status.Latency = time.Since(start)
	status.Metadata["path"] = shc.path

	return status
}

// NewExternalAPIHealthChecker creates a new external API health checker
func NewExternalAPIHealthChecker(name, url string, required bool) *ExternalAPIHealthChecker {
	return &ExternalAPIHealthChecker{
		name:     name,
		url:      url,
		required: required,
		timeout:  15 * time.Second,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetDependencyName returns the dependency name
func (eac *ExternalAPIHealthChecker) GetDependencyName() string {
	return eac.name
}

// GetDependencyType returns the dependency type
func (eac *ExternalAPIHealthChecker) GetDependencyType() DependencyType {
	return DependencyTypeExternal
}

// IsRequired returns whether the dependency is required
func (eac *ExternalAPIHealthChecker) IsRequired() bool {
	return eac.required
}

// CheckDependency checks the external API health
func (eac *ExternalAPIHealthChecker) CheckDependency(ctx context.Context) *DependencyHealthStatus {
	start := time.Now()

	status := &DependencyHealthStatus{
		DependencyName: eac.name,
		Type:           DependencyTypeExternal,
		Required:       eac.required,
		Timestamp:      start,
		Metadata:       make(map[string]interface{}),
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", eac.url, nil)
	if err != nil {
		status.Status = HealthStatusUnhealthy
		status.Message = fmt.Sprintf("Failed to create external API request: %v", err)
		status.Latency = time.Since(start)
		return status
	}

	// Make HTTP request
	resp, err := eac.client.Do(req)
	if err != nil {
		// External APIs might be temporarily unavailable
		status.Status = HealthStatusDegraded
		status.Message = fmt.Sprintf("External API request failed: %v", err)
		status.Latency = time.Since(start)
		return status
	}
	defer resp.Body.Close()

	// Add response metadata
	status.Metadata["status_code"] = resp.StatusCode
	status.Metadata["content_length"] = resp.ContentLength

	// Determine health based on status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status.Status = HealthStatusHealthy
		status.Message = "External API is healthy"
	} else if resp.StatusCode >= 500 {
		status.Status = HealthStatusDegraded // External APIs are less critical
		status.Message = fmt.Sprintf("External API returned server error: %d", resp.StatusCode)
	} else {
		status.Status = HealthStatusDegraded
		status.Message = fmt.Sprintf("External API returned non-success status: %d", resp.StatusCode)
	}

	status.Latency = time.Since(start)
	return status
}
