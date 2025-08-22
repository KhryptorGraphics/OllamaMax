package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"log/slog"
)

// HealthChecker provides comprehensive health monitoring
type HealthChecker struct {
	logger *slog.Logger
	config *HealthConfig

	// Health check functions
	checks map[string]HealthCheck
	mu     sync.RWMutex

	// Status tracking
	status     HealthStatus
	lastCheck  time.Time
	statusMu   sync.RWMutex

	// Component health
	components map[string]*ComponentHealth
	compMu     sync.RWMutex

	// Metrics
	metrics *HealthMetrics
}

// HealthConfig holds health checker configuration
type HealthConfig struct {
	Enabled          bool          `yaml:"enabled" json:"enabled"`
	CheckInterval    time.Duration `yaml:"check_interval" json:"check_interval"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout"`
	GracePeriod      time.Duration `yaml:"grace_period" json:"grace_period"`
	FailureThreshold int           `yaml:"failure_threshold" json:"failure_threshold"`
	Port             int           `yaml:"port" json:"port"`
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// HealthStatus represents overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth tracks health of individual components
type ComponentHealth struct {
	Name           string        `json:"name"`
	Status         HealthStatus  `json:"status"`
	LastCheck      time.Time     `json:"last_check"`
	LastError      string        `json:"last_error,omitempty"`
	FailureCount   int           `json:"failure_count"`
	ResponseTime   time.Duration `json:"response_time"`
	ConsecutiveFails int         `json:"consecutive_fails"`
	Details        interface{}   `json:"details,omitempty"`
}

// HealthMetrics tracks health check metrics
type HealthMetrics struct {
	TotalChecks      int64         `json:"total_checks"`
	SuccessfulChecks int64         `json:"successful_checks"`
	FailedChecks     int64         `json:"failed_checks"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastCheckTime    time.Time     `json:"last_check_time"`
	Uptime           time.Duration `json:"uptime"`
	StartTime        time.Time     `json:"start_time"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     HealthStatus                    `json:"status"`
	Timestamp  time.Time                       `json:"timestamp"`
	Uptime     time.Duration                   `json:"uptime"`
	Version    string                          `json:"version"`
	Components map[string]*ComponentHealth     `json:"components"`
	System     *SystemHealth                   `json:"system"`
	Metrics    *HealthMetrics                  `json:"metrics"`
	Details    map[string]interface{}          `json:"details,omitempty"`
}

// SystemHealth represents system-level health information
type SystemHealth struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	GoroutineCount int   `json:"goroutine_count"`
	HeapSize     uint64  `json:"heap_size_bytes"`
	HeapObjects  uint64  `json:"heap_objects"`
	GCPauses     int64   `json:"gc_pauses_total"`
}

// DefaultHealthConfig returns default health checker configuration
func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		Enabled:          true,
		CheckInterval:    30 * time.Second,
		Timeout:          10 * time.Second,
		GracePeriod:      2 * time.Minute,
		FailureThreshold: 3,
		Port:             8082,
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *HealthConfig, logger *slog.Logger) *HealthChecker {
	if config == nil {
		config = DefaultHealthConfig()
	}

	hc := &HealthChecker{
		logger:     logger,
		config:     config,
		checks:     make(map[string]HealthCheck),
		status:     HealthStatusUnknown,
		components: make(map[string]*ComponentHealth),
		metrics: &HealthMetrics{
			StartTime: time.Now(),
		},
	}

	// Register default system checks
	hc.registerDefaultChecks()

	return hc
}

// registerDefaultChecks registers built-in health checks
func (hc *HealthChecker) registerDefaultChecks() {
	// System memory check
	hc.RegisterCheck("memory", func(ctx context.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		// Check if memory usage is too high (>90%)
		memUsage := float64(m.Sys) / float64(1024*1024*1024) // Convert to GB
		if memUsage > 8.0 { // Assuming 8GB max
			return fmt.Errorf("high memory usage: %.2f GB", memUsage)
		}
		return nil
	})

	// Goroutine leak check
	hc.RegisterCheck("goroutines", func(ctx context.Context) error {
		count := runtime.NumGoroutine()
		if count > 1000 { // Arbitrary threshold
			return fmt.Errorf("high goroutine count: %d", count)
		}
		return nil
	})

	// Runtime check
	hc.RegisterCheck("runtime", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	})
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = check
	
	// Initialize component health
	hc.compMu.Lock()
	hc.components[name] = &ComponentHealth{
		Name:   name,
		Status: HealthStatusUnknown,
	}
	hc.compMu.Unlock()

	hc.logger.Info("Health check registered", "name", name)
}

// UnregisterCheck removes a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.checks, name)
	
	hc.compMu.Lock()
	delete(hc.components, name)
	hc.compMu.Unlock()

	hc.logger.Info("Health check unregistered", "name", name)
}

// Start starts the health checker
func (hc *HealthChecker) Start(ctx context.Context) error {
	if !hc.config.Enabled {
		hc.logger.Info("Health checker disabled")
		return nil
	}

	hc.logger.Info("Starting health checker", "interval", hc.config.CheckInterval)

	// Start periodic health checks
	go hc.runPeriodicChecks(ctx)

	// Start health HTTP server
	go hc.startHealthServer(ctx)

	return nil
}

// runPeriodicChecks runs health checks periodically
func (hc *HealthChecker) runPeriodicChecks(ctx context.Context) {
	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()

	// Run initial check
	hc.runHealthChecks(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.runHealthChecks(ctx)
		}
	}
}

// runHealthChecks executes all registered health checks
func (hc *HealthChecker) runHealthChecks(ctx context.Context) {
	start := time.Now()
	
	hc.mu.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mu.RUnlock()

	// Run checks concurrently
	type checkResult struct {
		name     string
		err      error
		duration time.Duration
	}

	resultChan := make(chan checkResult, len(checks))
	
	for name, check := range checks {
		go func(name string, check HealthCheck) {
			checkStart := time.Now()
			checkCtx, cancel := context.WithTimeout(ctx, hc.config.Timeout)
			defer cancel()

			err := check(checkCtx)
			duration := time.Since(checkStart)

			resultChan <- checkResult{
				name:     name,
				err:      err,
				duration: duration,
			}
		}(name, check)
	}

	// Collect results
	healthyCount := 0
	totalCount := len(checks)

	for i := 0; i < totalCount; i++ {
		result := <-resultChan
		
		hc.compMu.Lock()
		component := hc.components[result.name]
		if component == nil {
			component = &ComponentHealth{Name: result.name}
			hc.components[result.name] = component
		}

		component.LastCheck = time.Now()
		component.ResponseTime = result.duration

		if result.err != nil {
			component.Status = HealthStatusUnhealthy
			component.LastError = result.err.Error()
			component.FailureCount++
			component.ConsecutiveFails++
		} else {
			component.Status = HealthStatusHealthy
			component.LastError = ""
			component.ConsecutiveFails = 0
			healthyCount++
		}
		hc.compMu.Unlock()
	}

	// Update overall status
	hc.updateOverallStatus(healthyCount, totalCount)
	
	// Update metrics
	hc.updateMetrics(time.Since(start), totalCount-healthyCount == 0)

	hc.statusMu.Lock()
	hc.lastCheck = time.Now()
	hc.statusMu.Unlock()
}

// updateOverallStatus updates the overall health status
func (hc *HealthChecker) updateOverallStatus(healthyCount, totalCount int) {
	hc.statusMu.Lock()
	defer hc.statusMu.Unlock()

	if totalCount == 0 {
		hc.status = HealthStatusUnknown
		return
	}

	healthRatio := float64(healthyCount) / float64(totalCount)
	
	switch {
	case healthRatio == 1.0:
		hc.status = HealthStatusHealthy
	case healthRatio >= 0.8:
		hc.status = HealthStatusDegraded
	default:
		hc.status = HealthStatusUnhealthy
	}
}

// updateMetrics updates health check metrics
func (hc *HealthChecker) updateMetrics(duration time.Duration, success bool) {
	hc.metrics.TotalChecks++
	hc.metrics.LastCheckTime = time.Now()
	hc.metrics.Uptime = time.Since(hc.metrics.StartTime)

	if success {
		hc.metrics.SuccessfulChecks++
	} else {
		hc.metrics.FailedChecks++
	}

	// Update average latency
	if hc.metrics.TotalChecks == 1 {
		hc.metrics.AverageLatency = duration
	} else {
		totalDuration := hc.metrics.AverageLatency * time.Duration(hc.metrics.TotalChecks-1)
		hc.metrics.AverageLatency = (totalDuration + duration) / time.Duration(hc.metrics.TotalChecks)
	}
}

// startHealthServer starts the health HTTP server
func (hc *HealthChecker) startHealthServer(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", hc.handleHealthCheck)
	mux.HandleFunc("/health/live", hc.handleLivenessCheck)
	mux.HandleFunc("/health/ready", hc.handleReadinessCheck)

	addr := fmt.Sprintf(":%d", hc.config.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	hc.logger.Info("Starting health server", "address", addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hc.logger.Error("Health server error", "error", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(shutdownCtx)
}

// handleHealthCheck handles comprehensive health check requests
func (hc *HealthChecker) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	response := hc.GetHealthStatus()
	
	w.Header().Set("Content-Type", "application/json")
	
	if response.Status == HealthStatusHealthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

// handleLivenessCheck handles Kubernetes liveness probe
func (hc *HealthChecker) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	// Liveness check - is the service running?
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	})
}

// handleReadinessCheck handles Kubernetes readiness probe
func (hc *HealthChecker) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	response := hc.GetHealthStatus()
	
	w.Header().Set("Content-Type", "application/json")
	
	// Ready if healthy or degraded (but not unhealthy)
	if response.Status == HealthStatusHealthy || response.Status == HealthStatusDegraded {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    response.Status,
		"timestamp": response.Timestamp,
	})
}

// GetHealthStatus returns the current health status
func (hc *HealthChecker) GetHealthStatus() *HealthResponse {
	hc.statusMu.RLock()
	status := hc.status
	lastCheck := hc.lastCheck
	hc.statusMu.RUnlock()

	hc.compMu.RLock()
	components := make(map[string]*ComponentHealth)
	for name, comp := range hc.components {
		// Create a copy
		components[name] = &ComponentHealth{
			Name:             comp.Name,
			Status:           comp.Status,
			LastCheck:        comp.LastCheck,
			LastError:        comp.LastError,
			FailureCount:     comp.FailureCount,
			ResponseTime:     comp.ResponseTime,
			ConsecutiveFails: comp.ConsecutiveFails,
			Details:          comp.Details,
		}
	}
	hc.compMu.RUnlock()

	return &HealthResponse{
		Status:     status,
		Timestamp:  time.Now().UTC(),
		Uptime:     time.Since(hc.metrics.StartTime),
		Version:    "1.0.0", // TODO: Get from build info
		Components: components,
		System:     hc.getSystemHealth(),
		Metrics:    hc.metrics,
		Details: map[string]interface{}{
			"last_check": lastCheck,
			"config":     hc.config,
		},
	}
}

// getSystemHealth returns system health information
func (hc *HealthChecker) getSystemHealth() *SystemHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &SystemHealth{
		CPUUsage:       0, // TODO: Implement CPU usage collection
		MemoryUsage:    float64(m.Alloc) / float64(m.Sys) * 100,
		GoroutineCount: runtime.NumGoroutine(),
		HeapSize:       m.HeapAlloc,
		HeapObjects:    m.HeapObjects,
		GCPauses:       int64(m.NumGC),
	}
}

// IsHealthy returns true if the system is healthy
func (hc *HealthChecker) IsHealthy() bool {
	hc.statusMu.RLock()
	defer hc.statusMu.RUnlock()
	return hc.status == HealthStatusHealthy
}

// GetStatus returns the current status
func (hc *HealthChecker) GetStatus() HealthStatus {
	hc.statusMu.RLock()
	defer hc.statusMu.RUnlock()
	return hc.status
}