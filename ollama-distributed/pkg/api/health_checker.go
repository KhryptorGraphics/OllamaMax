package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthCheckConfig configures health checking
type HealthCheckConfig struct {
	CheckInterval    time.Duration
	Timeout          time.Duration
	FailureThreshold int
	SuccessThreshold int
	RetryInterval    time.Duration
}

// HealthMetrics tracks health checking performance
type HealthMetrics struct {
	TotalChecks       int64         `json:"total_checks"`
	SuccessfulChecks  int64         `json:"successful_checks"`
	FailedChecks      int64         `json:"failed_checks"`
	HealthyServices   int64         `json:"healthy_services"`
	UnhealthyServices int64         `json:"unhealthy_services"`
	AverageCheckTime  time.Duration `json:"average_check_time"`
	LastUpdated       time.Time     `json:"last_updated"`
	mu                sync.RWMutex
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	URL          string                 `json:"url"`
	Status       ServiceHealthStatus    `json:"status"`
	LastCheck    time.Time              `json:"last_check"`
	LastSuccess  time.Time              `json:"last_success"`
	LastFailure  time.Time              `json:"last_failure"`
	FailureCount int                    `json:"failure_count"`
	SuccessCount int                    `json:"success_count"`
	ResponseTime time.Duration          `json:"response_time"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Config       *HealthCheckConfig     `json:"config"`
}

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus string

const (
	ServiceHealthStatusHealthy   ServiceHealthStatus = "healthy"
	ServiceHealthStatusUnhealthy ServiceHealthStatus = "unhealthy"
	ServiceHealthStatusUnknown   ServiceHealthStatus = "unknown"
	ServiceHealthStatusChecking  ServiceHealthStatus = "checking"
)

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *HealthCheckConfig) (*HealthChecker, error) {
	if config == nil {
		config = &HealthCheckConfig{
			CheckInterval:    30 * time.Second,
			Timeout:          10 * time.Second,
			FailureThreshold: 3,
			SuccessThreshold: 2,
			RetryInterval:    5 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	checker := &HealthChecker{
		config:   config,
		services: make(map[string]*ServiceHealth),
		metrics: &HealthMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return checker, nil
}

// Start starts the health checker
func (hc *HealthChecker) Start() error {
	// Start health checking loop
	hc.wg.Add(1)
	go hc.healthCheckLoop()

	// Start metrics collection
	hc.wg.Add(1)
	go hc.metricsLoop()

	return nil
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() error {
	hc.cancel()
	hc.wg.Wait()
	return nil
}

// AddService adds a service to health check
func (hc *HealthChecker) AddService(service *ServiceHealth) error {
	if service == nil {
		return fmt.Errorf("service cannot be nil")
	}

	if service.ID == "" {
		service.ID = generateServiceID()
	}

	if service.Config == nil {
		service.Config = hc.config
	}

	service.Status = ServiceHealthStatusUnknown
	service.Metadata = make(map[string]interface{})

	hc.servicesMu.Lock()
	defer hc.servicesMu.Unlock()

	hc.services[service.ID] = service
	return nil
}

// RemoveService removes a service from health checking
func (hc *HealthChecker) RemoveService(serviceID string) error {
	hc.servicesMu.Lock()
	defer hc.servicesMu.Unlock()

	if _, exists := hc.services[serviceID]; !exists {
		return fmt.Errorf("service not found")
	}

	delete(hc.services, serviceID)
	return nil
}

// GetService returns a service by ID
func (hc *HealthChecker) GetService(serviceID string) (*ServiceHealth, bool) {
	hc.servicesMu.RLock()
	defer hc.servicesMu.RUnlock()

	service, exists := hc.services[serviceID]
	return service, exists
}

// GetAllServices returns all services
func (hc *HealthChecker) GetAllServices() []*ServiceHealth {
	hc.servicesMu.RLock()
	defer hc.servicesMu.RUnlock()

	services := make([]*ServiceHealth, 0, len(hc.services))
	for _, service := range hc.services {
		services = append(services, service)
	}

	return services
}

// GetHealthyServices returns all healthy services
func (hc *HealthChecker) GetHealthyServices() []*ServiceHealth {
	hc.servicesMu.RLock()
	defer hc.servicesMu.RUnlock()

	var healthy []*ServiceHealth
	for _, service := range hc.services {
		if service.Status == ServiceHealthStatusHealthy {
			healthy = append(healthy, service)
		}
	}

	return healthy
}

// IsServiceHealthy checks if a service is healthy
func (hc *HealthChecker) IsServiceHealthy(serviceID string) bool {
	hc.servicesMu.RLock()
	defer hc.servicesMu.RUnlock()

	service, exists := hc.services[serviceID]
	if !exists {
		return false
	}

	return service.Status == ServiceHealthStatusHealthy
}

// CheckService performs a health check on a specific service
func (hc *HealthChecker) CheckService(serviceID string) error {
	hc.servicesMu.RLock()
	service, exists := hc.services[serviceID]
	hc.servicesMu.RUnlock()

	if !exists {
		return fmt.Errorf("service not found")
	}

	return hc.performHealthCheck(service)
}

// performHealthCheck performs a health check on a service
func (hc *HealthChecker) performHealthCheck(service *ServiceHealth) error {
	start := time.Now()

	// Update metrics
	hc.metrics.mu.Lock()
	hc.metrics.TotalChecks++
	hc.metrics.mu.Unlock()

	// Set status to checking
	service.Status = ServiceHealthStatusChecking
	service.LastCheck = start

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: service.Config.Timeout,
	}

	// Perform health check
	resp, err := client.Get(service.URL)
	duration := time.Since(start)
	service.ResponseTime = duration

	if err != nil {
		// Health check failed
		service.Status = ServiceHealthStatusUnhealthy
		service.LastFailure = time.Now()
		service.FailureCount++
		service.ErrorMessage = err.Error()

		hc.metrics.mu.Lock()
		hc.metrics.FailedChecks++
		hc.metrics.LastUpdated = time.Now()
		hc.metrics.mu.Unlock()

		return err
	}

	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Health check succeeded
		service.SuccessCount++
		service.ErrorMessage = ""

		// Update status based on success threshold
		if service.SuccessCount >= service.Config.SuccessThreshold {
			service.Status = ServiceHealthStatusHealthy
			service.LastSuccess = time.Now()
			service.FailureCount = 0 // Reset failure count on success
		}

		hc.metrics.mu.Lock()
		hc.metrics.SuccessfulChecks++
		if hc.metrics.SuccessfulChecks == 1 {
			hc.metrics.AverageCheckTime = duration
		} else {
			hc.metrics.AverageCheckTime = (hc.metrics.AverageCheckTime + duration) / 2
		}
		hc.metrics.LastUpdated = time.Now()
		hc.metrics.mu.Unlock()

		return nil
	} else {
		// Health check failed due to bad status code
		service.Status = ServiceHealthStatusUnhealthy
		service.LastFailure = time.Now()
		service.FailureCount++
		service.ErrorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)

		hc.metrics.mu.Lock()
		hc.metrics.FailedChecks++
		hc.metrics.LastUpdated = time.Now()
		hc.metrics.mu.Unlock()

		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}
}

// GetMetrics returns health checking metrics
func (hc *HealthChecker) GetMetrics() *HealthMetrics {
	hc.metrics.mu.RLock()
	defer hc.metrics.mu.RUnlock()

	// Create a copy
	metrics := *hc.metrics
	return &metrics
}

// healthCheckLoop runs the health checking loop
func (hc *HealthChecker) healthCheckLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.checkAllServices()
		}
	}
}

// checkAllServices performs health checks on all services
func (hc *HealthChecker) checkAllServices() {
	hc.servicesMu.RLock()
	services := make([]*ServiceHealth, 0, len(hc.services))
	for _, service := range hc.services {
		services = append(services, service)
	}
	hc.servicesMu.RUnlock()

	// Check services concurrently
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(s *ServiceHealth) {
			defer wg.Done()
			hc.performHealthCheck(s)
		}(service)
	}

	wg.Wait()
}

// metricsLoop runs the metrics collection loop
func (hc *HealthChecker) metricsLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hc.ctx.Done():
			return
		case <-ticker.C:
			hc.updateMetrics()
		}
	}
}

// updateMetrics updates health checking metrics
func (hc *HealthChecker) updateMetrics() {
	hc.metrics.mu.Lock()
	defer hc.metrics.mu.Unlock()

	hc.servicesMu.RLock()
	var healthy, unhealthy int64
	for _, service := range hc.services {
		switch service.Status {
		case ServiceHealthStatusHealthy:
			healthy++
		case ServiceHealthStatusUnhealthy:
			unhealthy++
		}
	}
	hc.servicesMu.RUnlock()

	hc.metrics.HealthyServices = healthy
	hc.metrics.UnhealthyServices = unhealthy
	hc.metrics.LastUpdated = time.Now()
}

// generateServiceID generates a unique service ID
func generateServiceID() string {
	return fmt.Sprintf("service_%d", time.Now().UnixNano())
}

// UpdateService updates a service configuration
func (hc *HealthChecker) UpdateService(serviceID string, updates *ServiceHealth) error {
	hc.servicesMu.Lock()
	defer hc.servicesMu.Unlock()

	service, exists := hc.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found")
	}

	// Update fields
	if updates.Name != "" {
		service.Name = updates.Name
	}
	if updates.URL != "" {
		service.URL = updates.URL
	}
	if updates.Config != nil {
		service.Config = updates.Config
	}
	if updates.Metadata != nil {
		service.Metadata = updates.Metadata
	}

	return nil
}

// Reset resets the health checker
func (hc *HealthChecker) Reset() {
	hc.servicesMu.Lock()
	defer hc.servicesMu.Unlock()

	hc.services = make(map[string]*ServiceHealth)

	hc.metrics.mu.Lock()
	hc.metrics.TotalChecks = 0
	hc.metrics.SuccessfulChecks = 0
	hc.metrics.FailedChecks = 0
	hc.metrics.HealthyServices = 0
	hc.metrics.UnhealthyServices = 0
	hc.metrics.AverageCheckTime = 0
	hc.metrics.LastUpdated = time.Now()
	hc.metrics.mu.Unlock()
}
