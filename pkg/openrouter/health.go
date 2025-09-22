package openrouter

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// HealthChecker manages health checking for OpenRouter
type HealthChecker struct {
	client    *Client
	logger    *logrus.Logger
	healthy   bool
	mu        sync.RWMutex
	ticker    *time.Ticker
	stopCh    chan struct{}
	lastCheck time.Time
	error     error
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(client *Client, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		client:  client,
		logger:  logger,
		healthy: true,
		stopCh:  make(chan struct{}),
	}
}

// Start starts the health checker
func (h *HealthChecker) Start() {
	h.ticker = time.NewTicker(60 * time.Second) // Check every minute
	
	go func() {
		// Initial health check
		h.check()
		
		for {
			select {
			case <-h.ticker.C:
				h.check()
			case <-h.stopCh:
				return
			}
		}
	}()
	
	h.logger.Info("Health checker started")
}

// Stop stops the health checker
func (h *HealthChecker) Stop() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
	close(h.stopCh)
	h.logger.Info("Health checker stopped")
}

// check performs a health check
func (h *HealthChecker) check() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	err := h.client.Health(ctx)
	
	h.mu.Lock()
	h.lastCheck = time.Now()
	h.error = err
	h.healthy = err == nil
	h.mu.Unlock()
	
	if err != nil {
		h.logger.WithError(err).Warn("Health check failed")
	} else {
		h.logger.Debug("Health check passed")
	}
}

// IsHealthy returns the current health status
func (h *HealthChecker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.healthy
}

// GetLastCheck returns the last check time and error
func (h *HealthChecker) GetLastCheck() (time.Time, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastCheck, h.error
}

// LoadBalancer manages load balancing for OpenRouter requests
type LoadBalancer struct {
	config *IntegrationConfig
	logger *logrus.Logger
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(config *IntegrationConfig) *LoadBalancer {
	return &LoadBalancer{
		config: config,
		logger: logrus.New(),
	}
}

// CostTracker tracks costs and usage for OpenRouter requests
type CostTracker struct {
	config       *IntegrationConfig
	logger       *logrus.Logger
	mu           sync.RWMutex
	totalCost    float64
	totalTokens  int
	requestCount int
	usageStats   map[string]interface{}
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(config *IntegrationConfig) *CostTracker {
	return &CostTracker{
		config:     config,
		logger:     logrus.New(),
		usageStats: make(map[string]interface{}),
	}
}

// TrackRequest tracks a request for cost calculation
func (c *CostTracker) TrackRequest(req *ChatRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.requestCount++
	
	c.logger.WithFields(logrus.Fields{
		"model": req.Model,
		"request_count": c.requestCount,
	}).Debug("Tracking request")
}

// TrackResponse tracks a response for cost calculation
func (c *CostTracker) TrackResponse(resp *ChatResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.totalTokens += resp.Usage.TotalTokens
	
	// Calculate cost (this would use actual model pricing)
	promptCost := float64(resp.Usage.PromptTokens) * 0.000003
	completionCost := float64(resp.Usage.CompletionTokens) * 0.000015
	c.totalCost += promptCost + completionCost
	
	c.logger.WithFields(logrus.Fields{
		"model": resp.Model,
		"prompt_tokens": resp.Usage.PromptTokens,
		"completion_tokens": resp.Usage.CompletionTokens,
		"total_cost": c.totalCost,
	}).Debug("Tracking response")
}

// GetStats returns cost statistics
func (c *CostTracker) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return map[string]interface{}{
		"total_cost":    c.totalCost,
		"total_tokens":  c.totalTokens,
		"request_count": c.requestCount,
	}
}

// GetUsageStats returns usage statistics
func (c *CostTracker) GetUsageStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.usageStats
}

// CacheManager manages caching for OpenRouter responses
type CacheManager struct {
	config *IntegrationConfig
	logger *logrus.Logger
	cache  map[string]*ChatResponse
	mu     sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *IntegrationConfig) *CacheManager {
	return &CacheManager{
		config: config,
		logger: logrus.New(),
		cache:  make(map[string]*ChatResponse),
	}
}

// Get retrieves a cached response
func (c *CacheManager) Get(req *ChatRequest) *ChatResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	key := c.generateKey(req)
	return c.cache[key]
}

// Set stores a response in cache
func (c *CacheManager) Set(req *ChatRequest, resp *ChatResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := c.generateKey(req)
	c.cache[key] = resp
	
	c.logger.WithField("key", key).Debug("Cached response")
}

// generateKey generates a cache key for a request
func (c *CacheManager) generateKey(req *ChatRequest) string {
	// This is a simple implementation - in practice you'd want a more sophisticated key
	return fmt.Sprintf("%s:%s", req.Model, req.Messages[len(req.Messages)-1].Content)
}
