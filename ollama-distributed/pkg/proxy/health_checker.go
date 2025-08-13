package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// InstanceHealthChecker monitors the health of Ollama instances
type InstanceHealthChecker struct {
	proxy    *OllamaProxy
	interval time.Duration

	// Health check configuration
	timeout       time.Duration
	retryAttempts int
	retryDelay    time.Duration

	// Circuit breaker
	circuitBreaker map[string]*CircuitBreaker

	// Metrics
	metrics *HealthCheckerMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// CircuitBreaker implements circuit breaker pattern for health checks
type CircuitBreaker struct {
	FailureCount    int
	LastFailureTime time.Time
	State           CircuitBreakerState
	Threshold       int
	Timeout         time.Duration

	mu sync.RWMutex
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// HealthCheckerMetrics tracks health checker performance
type HealthCheckerMetrics struct {
	TotalChecks      int64
	SuccessfulChecks int64
	FailedChecks     int64
	AverageCheckTime time.Duration
	LastCheckTime    time.Time

	// Per-instance metrics
	InstanceChecks map[string]*InstanceHealthMetrics

	mu sync.RWMutex
}

// InstanceHealthMetrics tracks health metrics for a specific instance
type InstanceHealthMetrics struct {
	TotalChecks      int64
	SuccessfulChecks int64
	FailedChecks     int64
	LastCheckTime    time.Time
	LastCheckResult  bool
	AverageLatency   time.Duration
	Uptime           time.Duration
	DowntimeStart    *time.Time
}

// NewInstanceHealthChecker creates a new health checker
func NewInstanceHealthChecker(proxy *OllamaProxy, interval time.Duration) *InstanceHealthChecker {
	ctx, cancel := context.WithCancel(context.Background())

	return &InstanceHealthChecker{
		proxy:          proxy,
		interval:       interval,
		timeout:        10 * time.Second,
		retryAttempts:  3,
		retryDelay:     2 * time.Second,
		circuitBreaker: make(map[string]*CircuitBreaker),
		metrics: &HealthCheckerMetrics{
			InstanceChecks: make(map[string]*InstanceHealthMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the health checker
func (hc *InstanceHealthChecker) Start() {
	defer hc.wg.Done()

	log.Printf("Starting instance health checker (interval: %v)", hc.interval)

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			log.Printf("Health checker stopped")
			return
		case <-ticker.C:
			hc.checkAllInstances()
		}
	}
}

// Stop stops the health checker
func (hc *InstanceHealthChecker) Stop() {
	hc.cancel()
	hc.wg.Wait()
}

// checkAllInstances checks the health of all registered instances
func (hc *InstanceHealthChecker) checkAllInstances() {
	instances := hc.proxy.GetInstances()

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *OllamaInstance) {
			defer wg.Done()
			hc.CheckInstance(inst)
		}(instance)
	}

	wg.Wait()
}

// CheckInstance performs a health check on a specific instance
func (hc *InstanceHealthChecker) CheckInstance(instance *OllamaInstance) {
	startTime := time.Now()

	// Update metrics
	hc.updateCheckMetrics(instance.ID, true)

	// Check circuit breaker
	if hc.isCircuitBreakerOpen(instance.ID) {
		log.Printf("Circuit breaker open for instance %s, skipping health check", instance.ID)
		hc.updateInstanceStatus(instance, InstanceStatusUnavailable, startTime, fmt.Errorf("circuit breaker open"))
		return
	}

	// Perform health check with retries
	var lastErr error
	for attempt := 0; attempt < hc.retryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(hc.retryDelay)
		}

		err := hc.performHealthCheck(instance)
		if err == nil {
			// Health check successful
			hc.updateInstanceStatus(instance, InstanceStatusHealthy, startTime, nil)
			hc.resetCircuitBreaker(instance.ID)
			hc.updateCheckMetrics(instance.ID, false)
			return
		}

		lastErr = err
		log.Printf("Health check attempt %d failed for instance %s: %v", attempt+1, instance.ID, err)
	}

	// All attempts failed
	hc.updateInstanceStatus(instance, InstanceStatusUnhealthy, startTime, lastErr)
	hc.recordCircuitBreakerFailure(instance.ID)
	hc.updateCheckMetrics(instance.ID, false)
}

// performHealthCheck performs the actual health check
func (hc *InstanceHealthChecker) performHealthCheck(instance *OllamaInstance) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: hc.timeout,
	}

	// Perform health check request
	healthURL := instance.Endpoint + "/api/tags"
	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// updateInstanceStatus updates the status of an instance
func (hc *InstanceHealthChecker) updateInstanceStatus(instance *OllamaInstance, status InstanceStatus, checkTime time.Time, err error) {
	instance.mu.Lock()
	defer instance.mu.Unlock()

	previousStatus := instance.Status
	instance.Status = status

	// Update health information
	if instance.Health == nil {
		instance.Health = &InstanceHealth{}
	}

	instance.Health.LastHealthCheck = checkTime
	instance.Health.ResponseTime = time.Since(checkTime)
	instance.Health.IsHealthy = (status == InstanceStatusHealthy)

	// Log status changes
	if previousStatus != status {
		if err != nil {
			log.Printf("Instance %s status changed: %s -> %s (error: %v)", instance.ID, previousStatus, status, err)
		} else {
			log.Printf("Instance %s status changed: %s -> %s", instance.ID, previousStatus, status)
		}
	}

	// Update uptime/downtime tracking
	hc.updateUptimeTracking(instance, status)
}

// updateUptimeTracking updates uptime/downtime tracking for an instance
func (hc *InstanceHealthChecker) updateUptimeTracking(instance *OllamaInstance, status InstanceStatus) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	metrics, exists := hc.metrics.InstanceChecks[instance.ID]
	if !exists {
		metrics = &InstanceHealthMetrics{}
		hc.metrics.InstanceChecks[instance.ID] = metrics
	}

	now := time.Now()

	if status == InstanceStatusHealthy {
		// Instance is healthy
		if metrics.DowntimeStart != nil {
			// Was down, now up
			metrics.DowntimeStart = nil
		}
		metrics.LastCheckResult = true
	} else {
		// Instance is unhealthy
		if metrics.DowntimeStart == nil {
			// Was up, now down
			metrics.DowntimeStart = &now
		}
		metrics.LastCheckResult = false
	}

	metrics.LastCheckTime = now
}

// Circuit breaker methods

// isCircuitBreakerOpen checks if the circuit breaker is open for an instance
func (hc *InstanceHealthChecker) isCircuitBreakerOpen(instanceID string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	cb, exists := hc.circuitBreaker[instanceID]
	if !exists {
		return false
	}

	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.State == CircuitBreakerOpen {
		// Check if timeout has passed
		if time.Since(cb.LastFailureTime) > cb.Timeout {
			// Move to half-open state
			cb.State = CircuitBreakerHalfOpen
			return false
		}
		return true
	}

	return false
}

// recordCircuitBreakerFailure records a failure for the circuit breaker
func (hc *InstanceHealthChecker) recordCircuitBreakerFailure(instanceID string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	cb, exists := hc.circuitBreaker[instanceID]
	if !exists {
		cb = &CircuitBreaker{
			Threshold: 5,
			Timeout:   30 * time.Second,
			State:     CircuitBreakerClosed,
		}
		hc.circuitBreaker[instanceID] = cb
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	// Check if threshold is reached
	if cb.FailureCount >= cb.Threshold && cb.State == CircuitBreakerClosed {
		cb.State = CircuitBreakerOpen
		log.Printf("Circuit breaker opened for instance %s (failures: %d)", instanceID, cb.FailureCount)
	}
}

// resetCircuitBreaker resets the circuit breaker for an instance
func (hc *InstanceHealthChecker) resetCircuitBreaker(instanceID string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	cb, exists := hc.circuitBreaker[instanceID]
	if !exists {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.State != CircuitBreakerClosed {
		log.Printf("Circuit breaker reset for instance %s", instanceID)
	}

	cb.FailureCount = 0
	cb.State = CircuitBreakerClosed
}

// updateCheckMetrics updates health check metrics
func (hc *InstanceHealthChecker) updateCheckMetrics(instanceID string, isStart bool) {
	hc.metrics.mu.Lock()
	defer hc.metrics.mu.Unlock()

	if isStart {
		hc.metrics.TotalChecks++
		hc.metrics.LastCheckTime = time.Now()

		// Update instance metrics
		if metrics, exists := hc.metrics.InstanceChecks[instanceID]; exists {
			metrics.TotalChecks++
		}
	} else {
		hc.metrics.SuccessfulChecks++

		// Update instance metrics
		if metrics, exists := hc.metrics.InstanceChecks[instanceID]; exists {
			metrics.SuccessfulChecks++
		}
	}
}

// GetMetrics returns current health checker metrics
func (hc *InstanceHealthChecker) GetMetrics() *HealthCheckerMetrics {
	hc.metrics.mu.RLock()
	defer hc.metrics.mu.RUnlock()

	// Return a copy of metrics
	metrics := &HealthCheckerMetrics{
		TotalChecks:      hc.metrics.TotalChecks,
		SuccessfulChecks: hc.metrics.SuccessfulChecks,
		FailedChecks:     hc.metrics.FailedChecks,
		AverageCheckTime: hc.metrics.AverageCheckTime,
		LastCheckTime:    hc.metrics.LastCheckTime,
		InstanceChecks:   make(map[string]*InstanceHealthMetrics),
	}

	// Copy instance metrics
	for id, instanceMetrics := range hc.metrics.InstanceChecks {
		metrics.InstanceChecks[id] = &InstanceHealthMetrics{
			TotalChecks:      instanceMetrics.TotalChecks,
			SuccessfulChecks: instanceMetrics.SuccessfulChecks,
			FailedChecks:     instanceMetrics.FailedChecks,
			LastCheckTime:    instanceMetrics.LastCheckTime,
			LastCheckResult:  instanceMetrics.LastCheckResult,
			AverageLatency:   instanceMetrics.AverageLatency,
			Uptime:           instanceMetrics.Uptime,
		}

		if instanceMetrics.DowntimeStart != nil {
			downtime := *instanceMetrics.DowntimeStart
			metrics.InstanceChecks[id].DowntimeStart = &downtime
		}
	}

	return metrics
}
