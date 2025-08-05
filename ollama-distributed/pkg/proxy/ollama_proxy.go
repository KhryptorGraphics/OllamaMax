package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler/loadbalancer"
)

// OllamaProxy manages distributed Ollama request routing
type OllamaProxy struct {
	// Core components
	scheduler    *scheduler.Engine
	loadBalancer *loadbalancer.LoadBalancer

	// Instance management
	instances   map[string]*OllamaInstance
	instancesMu sync.RWMutex

	// Request routing
	router *RequestRouter

	// Health monitoring
	healthChecker *InstanceHealthChecker

	// Metrics
	metrics *ProxyMetrics

	// Configuration
	config *ProxyConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// OllamaInstance represents a single Ollama instance
type OllamaInstance struct {
	ID       string
	NodeID   string
	Endpoint string
	Status   InstanceStatus
	Models   []string
	Load     *InstanceLoad
	Health   *InstanceHealth

	// HTTP client for this instance
	client *http.Client
	proxy  *httputil.ReverseProxy

	// Metrics
	RequestCount    int64
	ErrorCount      int64
	AverageLatency  time.Duration
	LastRequestTime time.Time

	mu sync.RWMutex
}

// InstanceStatus represents the status of an Ollama instance
type InstanceStatus string

const (
	InstanceStatusHealthy     InstanceStatus = "healthy"
	InstanceStatusUnhealthy   InstanceStatus = "unhealthy"
	InstanceStatusUnavailable InstanceStatus = "unavailable"
	InstanceStatusStarting    InstanceStatus = "starting"
	InstanceStatusStopping    InstanceStatus = "stopping"
	InstanceStatusDraining    InstanceStatus = "draining"
	InstanceStatusUnknown     InstanceStatus = "unknown"
)

// InstanceLoad represents the current load of an instance
type InstanceLoad struct {
	ActiveRequests int
	QueuedRequests int
	CPUUsage       float64
	MemoryUsage    float64
	GPUUsage       float64
	LastUpdated    time.Time
}

// InstanceHealth represents the health status of an instance
type InstanceHealth struct {
	IsHealthy       bool
	LastHealthCheck time.Time
	ResponseTime    time.Duration
	ErrorRate       float64
	Uptime          time.Duration
}

// ProxyConfig configures the Ollama proxy
type ProxyConfig struct {
	// Load balancing
	LoadBalancingStrategy string
	HealthCheckInterval   time.Duration
	RequestTimeout        time.Duration

	// Retry configuration
	MaxRetries   int
	RetryDelay   time.Duration
	RetryBackoff float64

	// Circuit breaker
	CircuitBreakerThreshold int
	CircuitBreakerTimeout   time.Duration

	// Model synchronization
	EnableModelSync   bool
	ModelSyncInterval time.Duration

	// Request routing
	EnableRequestLogging bool
	EnableMetrics        bool
}

// ProxyMetrics tracks proxy performance
type ProxyMetrics struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	AverageLatency     time.Duration
	RequestsPerSecond  float64

	// Per-instance metrics
	InstanceMetrics map[string]*InstanceMetrics

	// Load balancing metrics
	LoadBalancingDecisions int64
	LoadBalancingErrors    int64

	mu sync.RWMutex
}

// InstanceMetrics tracks metrics for a specific instance
type InstanceMetrics struct {
	Requests       int64
	Errors         int64
	AverageLatency time.Duration
	LastRequest    time.Time
}

// NewOllamaProxy creates a new Ollama proxy
func NewOllamaProxy(scheduler *scheduler.Engine, loadBalancer *loadbalancer.LoadBalancer, config *ProxyConfig) (*OllamaProxy, error) {
	if config == nil {
		config = DefaultProxyConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	proxy := &OllamaProxy{
		scheduler:    scheduler,
		loadBalancer: loadBalancer,
		instances:    make(map[string]*OllamaInstance),
		config:       config,
		metrics: &ProxyMetrics{
			InstanceMetrics: make(map[string]*InstanceMetrics),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize components
	if err := proxy.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize proxy components: %w", err)
	}

	return proxy, nil
}

// DefaultProxyConfig returns default proxy configuration
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		LoadBalancingStrategy:   "least_loaded",
		HealthCheckInterval:     30 * time.Second,
		RequestTimeout:          60 * time.Second,
		MaxRetries:              3,
		RetryDelay:              1 * time.Second,
		RetryBackoff:            2.0,
		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
		EnableModelSync:         true,
		ModelSyncInterval:       5 * time.Minute,
		EnableRequestLogging:    true,
		EnableMetrics:           true,
	}
}

// initializeComponents initializes proxy components
func (p *OllamaProxy) initializeComponents() error {
	// Initialize request router
	p.router = NewRequestRouter(p)

	// Initialize health checker
	p.healthChecker = NewInstanceHealthChecker(p, p.config.HealthCheckInterval)

	return nil
}

// Start starts the Ollama proxy
func (p *OllamaProxy) Start() error {
	log.Printf("Starting Ollama proxy...")

	// Start health checker
	p.wg.Add(1)
	go p.healthChecker.Start()

	// Start model synchronization if enabled
	if p.config.EnableModelSync {
		p.wg.Add(1)
		go p.modelSyncLoop()
	}

	// Start metrics collection if enabled
	if p.config.EnableMetrics {
		p.wg.Add(1)
		go p.metricsLoop()
	}

	// Discover existing Ollama instances
	if err := p.discoverInstances(); err != nil {
		log.Printf("Warning: Failed to discover instances: %v", err)
	}

	// Start periodic discovery
	p.wg.Add(1)
	go p.periodicDiscovery()

	log.Printf("Ollama proxy started successfully")
	return nil
}

// Stop stops the Ollama proxy
func (p *OllamaProxy) Stop() error {
	log.Printf("Stopping Ollama proxy...")

	p.cancel()
	p.wg.Wait()

	log.Printf("Ollama proxy stopped")
	return nil
}

// RegisterInstance registers a new Ollama instance
func (p *OllamaProxy) RegisterInstance(nodeID, endpoint string) error {
	p.instancesMu.Lock()
	defer p.instancesMu.Unlock()

	instanceID := fmt.Sprintf("%s-%s", nodeID, endpoint)

	// Parse endpoint URL
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: p.config.RequestTimeout,
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(endpointURL)

	instance := &OllamaInstance{
		ID:       instanceID,
		NodeID:   nodeID,
		Endpoint: endpoint,
		Status:   InstanceStatusStarting,
		Models:   []string{},
		Load: &InstanceLoad{
			LastUpdated: time.Now(),
		},
		Health: &InstanceHealth{
			LastHealthCheck: time.Now(),
		},
		client: client,
		proxy:  proxy,
	}

	p.instances[instanceID] = instance

	// Initialize metrics for this instance
	p.metrics.mu.Lock()
	p.metrics.InstanceMetrics[instanceID] = &InstanceMetrics{}
	p.metrics.mu.Unlock()

	log.Printf("Registered Ollama instance: %s (Node: %s, Endpoint: %s)", instanceID, nodeID, endpoint)

	// Perform initial health check
	go p.healthChecker.CheckInstance(instance)

	return nil
}

// selectInstance selects the best instance for a request
func (p *OllamaProxy) selectInstance(r *http.Request) (*OllamaInstance, error) {
	p.instancesMu.RLock()
	defer p.instancesMu.RUnlock()

	// Get healthy instances
	healthyInstances := make([]*OllamaInstance, 0)
	for _, instance := range p.instances {
		if instance.Status == InstanceStatusHealthy {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, fmt.Errorf("no healthy instances available")
	}

	// Use load balancer to select instance
	return p.loadBalanceInstances(healthyInstances, r)
}

// loadBalanceInstances uses the load balancer to select an instance
func (p *OllamaProxy) loadBalanceInstances(instances []*OllamaInstance, r *http.Request) (*OllamaInstance, error) {
	// Convert instances to load balancer nodes
	nodes := make([]*loadbalancer.LoadBalancedNode, len(instances))
	for i, instance := range instances {
		nodes[i] = &loadbalancer.LoadBalancedNode{
			NodeID: instance.NodeID,
			Weight: 1.0, // Default weight
			CurrentLoad: &loadbalancer.NodeLoadMetrics{
				OverallLoad: float64(instance.Load.ActiveRequests) / 10.0, // Normalize
			},
			Reliability: 1.0, // Default reliability
			Available:   true,
			LastUpdate:  time.Now(),
		}
	}

	// Select node using load balancer
	selectedNode, err := p.loadBalancer.SelectNode(0.1) // Assume light task load
	if err != nil {
		// Fallback to round-robin
		return instances[int(time.Now().UnixNano())%len(instances)], nil
	}

	// Find corresponding instance by NodeID
	for _, instance := range instances {
		if instance.NodeID == selectedNode.NodeID {
			return instance, nil
		}
	}

	// Fallback to first instance
	return instances[0], nil
}

// routeRequest routes a request to a specific instance
func (p *OllamaProxy) routeRequest(w http.ResponseWriter, r *http.Request, instance *OllamaInstance) error {
	// Update instance load
	instance.mu.Lock()
	instance.Load.ActiveRequests++
	instance.RequestCount++
	instance.LastRequestTime = time.Now()
	instance.mu.Unlock()

	// Defer load cleanup
	defer func() {
		instance.mu.Lock()
		instance.Load.ActiveRequests--
		instance.mu.Unlock()
	}()

	// Clone request for modification
	proxyReq := r.Clone(r.Context())

	// Update target URL
	targetURL, err := url.Parse(instance.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid instance endpoint: %w", err)
	}

	proxyReq.URL.Scheme = targetURL.Scheme
	proxyReq.URL.Host = targetURL.Host
	proxyReq.Host = targetURL.Host

	// Use reverse proxy
	instance.proxy.ServeHTTP(w, proxyReq)

	return nil
}

// discoverInstances discovers existing Ollama instances from the scheduler
func (p *OllamaProxy) discoverInstances() error {
	log.Printf("Discovering Ollama instances...")

	// Integrate with scheduler.Engine to get node list
	if p.scheduler != nil {
		if err := p.discoverFromScheduler(); err != nil {
			log.Printf("Warning: Failed to discover from scheduler: %v", err)
		}
	} else {
		log.Printf("Warning: Scheduler not available, using basic discovery")
	}

	// Also try to register local instance if available
	if err := p.registerLocalInstance(); err != nil {
		log.Printf("Warning: Failed to register local instance: %v", err)
	}

	return nil
}

// discoverFromScheduler discovers instances from the scheduler engine
func (p *OllamaProxy) discoverFromScheduler() error {
	// Get available nodes from scheduler
	nodes := p.scheduler.GetAvailableNodes()

	log.Printf("Found %d available nodes from scheduler", len(nodes))

	for _, node := range nodes {
		endpoint := p.buildOllamaEndpoint(node.Address)

		// Parse endpoint URL for reverse proxy
		endpointURL, err := url.Parse(endpoint)
		if err != nil {
			log.Printf("Warning: Invalid endpoint URL %s for node %s: %v", endpoint, node.ID, err)
			continue
		}

		// Create HTTP client
		client := &http.Client{
			Timeout: p.config.RequestTimeout,
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(endpointURL)

		// Create instance from node info
		instance := &OllamaInstance{
			ID:       fmt.Sprintf("node-%s", node.ID),
			NodeID:   node.ID,
			Endpoint: endpoint,
			Status:   p.mapNodeStatusToInstanceStatus(string(node.Status)),
			Models:   []string{}, // Will be populated by health checks
			Load:     &InstanceLoad{},
			Health:   &InstanceHealth{},

			// HTTP components
			client: client,
			proxy:  proxy,

			// Initialize metrics
			RequestCount:    0,
			ErrorCount:      0,
			AverageLatency:  0,
			LastRequestTime: time.Time{},
		}

		// Register the instance
		p.instancesMu.Lock()
		p.instances[instance.ID] = instance
		p.instancesMu.Unlock()

		// Initialize metrics for this instance
		p.metrics.mu.Lock()
		if p.metrics.InstanceMetrics == nil {
			p.metrics.InstanceMetrics = make(map[string]*InstanceMetrics)
		}
		p.metrics.InstanceMetrics[instance.ID] = &InstanceMetrics{
			Requests:       0,
			Errors:         0,
			AverageLatency: 0,
			LastRequest:    time.Time{},
		}
		p.metrics.mu.Unlock()

		log.Printf("Registered instance from scheduler: %s (Node: %s, Endpoint: %s)",
			instance.ID, instance.NodeID, instance.Endpoint)
	}

	return nil
}

// buildOllamaEndpoint builds the Ollama API endpoint from node address
func (p *OllamaProxy) buildOllamaEndpoint(nodeAddress string) string {
	// Parse the node address and construct Ollama endpoint
	// Assume Ollama runs on port 11434 by default
	if strings.Contains(nodeAddress, ":") {
		// Address already has port, replace with Ollama port
		host := strings.Split(nodeAddress, ":")[0]
		return fmt.Sprintf("http://%s:11434", host)
	}

	// No port specified, add Ollama port
	return fmt.Sprintf("http://%s:11434", nodeAddress)
}

// mapNodeStatusToInstanceStatus maps scheduler node status to proxy instance status
func (p *OllamaProxy) mapNodeStatusToInstanceStatus(nodeStatus string) InstanceStatus {
	switch nodeStatus {
	case "online":
		return InstanceStatusHealthy
	case "offline":
		return InstanceStatusUnavailable
	case "draining":
		return InstanceStatusDraining
	default:
		return InstanceStatusUnknown
	}
}

// periodicDiscovery runs periodic discovery of new instances
func (p *OllamaProxy) periodicDiscovery() {
	defer p.wg.Done()

	// Run discovery every 60 seconds
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Stopping periodic discovery")
			return
		case <-ticker.C:
			log.Printf("Running periodic instance discovery...")
			if err := p.discoverInstances(); err != nil {
				log.Printf("Warning: Periodic discovery failed: %v", err)
			}
		}
	}
}

// registerLocalInstance registers the local Ollama instance
func (p *OllamaProxy) registerLocalInstance() error {
	// Check if local Ollama is running
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return fmt.Errorf("local Ollama not available: %w", err)
	}
	resp.Body.Close()

	// Register local instance
	return p.RegisterInstance("local", "http://localhost:11434")
}

// modelSyncLoop synchronizes models across instances
func (p *OllamaProxy) modelSyncLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.ModelSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.synchronizeModels()
		}
	}
}

// synchronizeModels synchronizes models across all instances
func (p *OllamaProxy) synchronizeModels() {
	p.instancesMu.RLock()
	instances := make([]*OllamaInstance, 0, len(p.instances))
	for _, instance := range p.instances {
		if instance.Status == InstanceStatusHealthy {
			instances = append(instances, instance)
		}
	}
	p.instancesMu.RUnlock()

	if len(instances) == 0 {
		return
	}

	// Get models from all instances
	allModels := make(map[string]bool)
	instanceModels := make(map[string][]string)

	for _, instance := range instances {
		models, err := p.getInstanceModels(instance)
		if err != nil {
			log.Printf("Failed to get models from instance %s: %v", instance.ID, err)
			continue
		}

		instanceModels[instance.ID] = models
		for _, model := range models {
			allModels[model] = true
		}
	}

	// TODO: Implement model synchronization logic
	// This would ensure all instances have the same models available
	log.Printf("Model synchronization completed. Found %d unique models across %d instances",
		len(allModels), len(instances))
}

// getInstanceModels gets the list of models from an instance
func (p *OllamaProxy) getInstanceModels(instance *OllamaInstance) ([]string, error) {
	resp, err := instance.client.Get(instance.Endpoint + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	models := make([]string, len(response.Models))
	for i, model := range response.Models {
		models[i] = model.Name
	}

	return models, nil
}

// metricsLoop collects and updates metrics
func (p *OllamaProxy) metricsLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.updateMetrics()
		}
	}
}

// updateMetrics updates proxy metrics
func (p *OllamaProxy) updateMetrics() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	// Calculate requests per second
	// This is a simplified calculation
	if p.metrics.TotalRequests > 0 {
		p.metrics.RequestsPerSecond = float64(p.metrics.TotalRequests) / 60.0 // Last minute
	}

	// Update instance metrics
	p.instancesMu.RLock()
	for instanceID, instance := range p.instances {
		if metrics, exists := p.metrics.InstanceMetrics[instanceID]; exists {
			instance.mu.RLock()
			metrics.Requests = instance.RequestCount
			metrics.Errors = instance.ErrorCount
			metrics.AverageLatency = instance.AverageLatency
			metrics.LastRequest = instance.LastRequestTime
			instance.mu.RUnlock()
		}
	}
	p.instancesMu.RUnlock()
}

// recordSuccess records a successful request
func (p *OllamaProxy) recordSuccess(duration time.Duration) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.SuccessfulRequests++

	// Update average latency
	if p.metrics.SuccessfulRequests == 1 {
		p.metrics.AverageLatency = duration
	} else {
		p.metrics.AverageLatency = (p.metrics.AverageLatency + duration) / 2
	}
}

// recordError records a failed request
func (p *OllamaProxy) recordError() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.FailedRequests++
}

// GetMetrics returns current proxy metrics
func (p *OllamaProxy) GetMetrics() *ProxyMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Return a copy of metrics
	metrics := &ProxyMetrics{
		TotalRequests:          p.metrics.TotalRequests,
		SuccessfulRequests:     p.metrics.SuccessfulRequests,
		FailedRequests:         p.metrics.FailedRequests,
		AverageLatency:         p.metrics.AverageLatency,
		RequestsPerSecond:      p.metrics.RequestsPerSecond,
		LoadBalancingDecisions: p.metrics.LoadBalancingDecisions,
		LoadBalancingErrors:    p.metrics.LoadBalancingErrors,
		InstanceMetrics:        make(map[string]*InstanceMetrics),
	}

	// Copy instance metrics
	for id, instanceMetrics := range p.metrics.InstanceMetrics {
		metrics.InstanceMetrics[id] = &InstanceMetrics{
			Requests:       instanceMetrics.Requests,
			Errors:         instanceMetrics.Errors,
			AverageLatency: instanceMetrics.AverageLatency,
			LastRequest:    instanceMetrics.LastRequest,
		}
	}

	return metrics
}

// GetInstances returns current instances
func (p *OllamaProxy) GetInstances() map[string]*OllamaInstance {
	p.instancesMu.RLock()
	defer p.instancesMu.RUnlock()

	// Return a copy of instances
	instances := make(map[string]*OllamaInstance)
	for id, instance := range p.instances {
		instances[id] = &OllamaInstance{
			ID:              instance.ID,
			NodeID:          instance.NodeID,
			Endpoint:        instance.Endpoint,
			Status:          instance.Status,
			Models:          append([]string{}, instance.Models...),
			RequestCount:    instance.RequestCount,
			ErrorCount:      instance.ErrorCount,
			AverageLatency:  instance.AverageLatency,
			LastRequestTime: instance.LastRequestTime,
		}

		// Copy load and health info
		if instance.Load != nil {
			instances[id].Load = &InstanceLoad{
				ActiveRequests: instance.Load.ActiveRequests,
				QueuedRequests: instance.Load.QueuedRequests,
				CPUUsage:       instance.Load.CPUUsage,
				MemoryUsage:    instance.Load.MemoryUsage,
				GPUUsage:       instance.Load.GPUUsage,
				LastUpdated:    instance.Load.LastUpdated,
			}
		}

		if instance.Health != nil {
			instances[id].Health = &InstanceHealth{
				IsHealthy:       instance.Health.IsHealthy,
				LastHealthCheck: instance.Health.LastHealthCheck,
				ResponseTime:    instance.Health.ResponseTime,
				ErrorRate:       instance.Health.ErrorRate,
				Uptime:          instance.Health.Uptime,
			}
		}
	}

	return instances
}

// UnregisterInstance unregisters an Ollama instance
func (p *OllamaProxy) UnregisterInstance(instanceID string) error {
	p.instancesMu.Lock()
	defer p.instancesMu.Unlock()

	if _, exists := p.instances[instanceID]; !exists {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	delete(p.instances, instanceID)

	// Clean up metrics
	p.metrics.mu.Lock()
	delete(p.metrics.InstanceMetrics, instanceID)
	p.metrics.mu.Unlock()

	log.Printf("Unregistered Ollama instance: %s", instanceID)
	return nil
}

// ProxyRequest routes a request to an appropriate Ollama instance
func (p *OllamaProxy) ProxyRequest(w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()

	// Update metrics
	p.metrics.mu.Lock()
	p.metrics.TotalRequests++
	p.metrics.mu.Unlock()

	// Select target instance
	instance, err := p.selectInstance(r)
	if err != nil {
		p.recordError()
		return fmt.Errorf("failed to select instance: %w", err)
	}

	// Route request to selected instance
	if err := p.routeRequest(w, r, instance); err != nil {
		p.recordError()
		return fmt.Errorf("failed to route request: %w", err)
	}

	// Update metrics
	duration := time.Since(startTime)
	p.recordSuccess(duration)

	return nil
}
