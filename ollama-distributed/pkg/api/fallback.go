package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama-distributed/pkg/integration"
)

// FallbackManager manages fallback mechanisms for distributed requests
type FallbackManager struct {
	// Local Ollama instance details
	localURL     *url.URL
	localClient  *http.Client
	localHealthy bool
	
	// Fallback configuration
	fallbackEnabled     bool
	fallbackTimeout     time.Duration
	healthCheckInterval time.Duration
	maxRetries          int
	
	// Health monitoring
	healthMu           sync.RWMutex
	lastHealthCheck    time.Time
	consecutiveFailures int
	
	// Fallback statistics
	statsMu         sync.RWMutex
	fallbackCount   int64
	successCount    int64
	failureCount    int64
	averageLatency  time.Duration
}

// NewFallbackManager creates a new fallback manager
func NewFallbackManager(localAddr string) (*FallbackManager, error) {
	localURL, err := url.Parse(localAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid local URL: %w", err)
	}
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	fm := &FallbackManager{
		localURL:            localURL,
		localClient:         client,
		localHealthy:        true,
		fallbackEnabled:     true,
		fallbackTimeout:     30 * time.Second,
		healthCheckInterval: 30 * time.Second,
		maxRetries:          3,
	}
	
	// Start health monitoring
	go fm.startHealthMonitoring()
	
	return fm, nil
}

// startHealthMonitoring starts continuous health monitoring of local instance
func (fm *FallbackManager) startHealthMonitoring() {
	ticker := time.NewTicker(fm.healthCheckInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		fm.checkLocalHealth()
	}
}

// checkLocalHealth checks the health of the local Ollama instance
func (fm *FallbackManager) checkLocalHealth() {
	fm.healthMu.Lock()
	defer fm.healthMu.Unlock()
	
	fm.lastHealthCheck = time.Now()
	
	// Create health check request
	healthURL := fmt.Sprintf("%s/api/version", fm.localURL.String())
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		fm.recordHealthFailure()
		return
	}
	
	resp, err := fm.localClient.Do(req)
	if err != nil {
		fm.recordHealthFailure()
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		fm.recordHealthSuccess()
	} else {
		fm.recordHealthFailure()
	}
}

// recordHealthSuccess records a successful health check
func (fm *FallbackManager) recordHealthSuccess() {
	fm.consecutiveFailures = 0
	if !fm.localHealthy {
		fm.localHealthy = true
		slog.Info("Local Ollama instance is healthy again")
	}
}

// recordHealthFailure records a failed health check
func (fm *FallbackManager) recordHealthFailure() {
	fm.consecutiveFailures++
	if fm.consecutiveFailures >= 3 && fm.localHealthy {
		fm.localHealthy = false
		slog.Warn("Local Ollama instance is unhealthy", "failures", fm.consecutiveFailures)
	}
}

// IsLocalHealthy returns whether the local instance is healthy
func (fm *FallbackManager) IsLocalHealthy() bool {
	fm.healthMu.RLock()
	defer fm.healthMu.RUnlock()
	return fm.localHealthy
}

// ShouldFallback determines if a request should fallback to local
func (fm *FallbackManager) ShouldFallback(reason string) bool {
	if !fm.fallbackEnabled {
		return false
	}
	
	// Always fallback if local is unhealthy and we have no other options
	if !fm.IsLocalHealthy() && reason != "local-unhealthy" {
		return false
	}
	
	// Check specific fallback reasons
	switch reason {
	case "scheduler-error":
		return true
	case "execution-error":
		return true
	case "timeout":
		return true
	case "no-nodes":
		return true
	case "model-not-found":
		return true
	case "distributed-error":
		return true
	default:
		return false
	}
}

// ExecuteFallback executes a fallback request to local Ollama
func (fm *FallbackManager) ExecuteFallback(c *gin.Context, reason string) error {
	if !fm.ShouldFallback(reason) {
		return fmt.Errorf("fallback not allowed for reason: %s", reason)
	}
	
	startTime := time.Now()
	
	// Increment fallback count
	fm.statsMu.Lock()
	fm.fallbackCount++
	fm.statsMu.Unlock()
	
	// Log fallback
	slog.Info("Executing fallback to local", "reason", reason, "path", c.Request.URL.Path)
	
	// Add fallback headers
	c.Header("X-Ollama-Fallback", "true")
	c.Header("X-Ollama-Fallback-Reason", reason)
	c.Header("X-Ollama-Fallback-Time", startTime.Format(time.RFC3339))
	
	// Execute fallback request
	err := fm.executeLocalRequest(c)
	
	// Update statistics
	latency := time.Since(startTime)
	fm.updateStats(err == nil, latency)
	
	return err
}

// executeLocalRequest executes a request against the local Ollama instance
func (fm *FallbackManager) executeLocalRequest(c *gin.Context) error {
	// Read request body
	var body []byte
	var err error
	
	if c.Request.Body != nil {
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}
		// Reset body for potential retries
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	
	// Construct local URL
	localURL := fmt.Sprintf("%s%s", fm.localURL.String(), c.Request.URL.Path)
	if c.Request.URL.RawQuery != "" {
		localURL += "?" + c.Request.URL.RawQuery
	}
	
	// Create request with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), fm.fallbackTimeout)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, c.Request.Method, localURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	
	// Execute request with retries
	var resp *http.Response
	var lastErr error
	
	for attempt := 0; attempt < fm.maxRetries; attempt++ {
		if attempt > 0 {
			slog.Debug("Retrying fallback request", "attempt", attempt+1, "maxRetries", fm.maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		
		resp, lastErr = fm.localClient.Do(req)
		if lastErr == nil {
			break
		}
		
		// Reset request body for retry
		if body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
		}
	}
	
	if lastErr != nil {
		return fmt.Errorf("fallback request failed after %d attempts: %w", fm.maxRetries, lastErr)
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	
	// Set status code
	c.Status(resp.StatusCode)
	
	// Stream response body
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy response body: %w", err)
	}
	
	return nil
}

// FallbackToLocal provides a convenience method for fallback
func (fm *FallbackManager) FallbackToLocal(c *gin.Context, reason string) {
	if err := fm.ExecuteFallback(c, reason); err != nil {
		slog.Error("Fallback execution failed", "error", err, "reason", reason)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Fallback to local failed",
			"reason": reason,
			"details": err.Error(),
		})
	}
}

// HandleDistributedError handles errors from distributed execution
func (fm *FallbackManager) HandleDistributedError(c *gin.Context, err error, operation string) {
	// Determine fallback reason based on error
	reason := "distributed-error"
	
	if err != nil {
		switch {
		case bytes.Contains([]byte(err.Error()), []byte("timeout")):
			reason = "timeout"
		case bytes.Contains([]byte(err.Error()), []byte("no nodes")):
			reason = "no-nodes"
		case bytes.Contains([]byte(err.Error()), []byte("model not found")):
			reason = "model-not-found"
		case bytes.Contains([]byte(err.Error()), []byte("scheduler")):
			reason = "scheduler-error"
		}
	}
	
	// Log error
	slog.Error("Distributed operation failed", "error", err, "operation", operation, "reason", reason)
	
	// Add error headers
	c.Header("X-Ollama-Distributed-Error", err.Error())
	c.Header("X-Ollama-Error-Operation", operation)
	
	// Try fallback
	if fm.ShouldFallback(reason) {
		fm.FallbackToLocal(c, reason)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Distributed operation failed",
			"operation": operation,
			"reason": reason,
			"details": err.Error(),
			"fallback_available": false,
		})
	}
}

// updateStats updates fallback statistics
func (fm *FallbackManager) updateStats(success bool, latency time.Duration) {
	fm.statsMu.Lock()
	defer fm.statsMu.Unlock()
	
	if success {
		fm.successCount++
	} else {
		fm.failureCount++
	}
	
	// Update average latency
	totalRequests := fm.successCount + fm.failureCount
	if totalRequests == 1 {
		fm.averageLatency = latency
	} else {
		fm.averageLatency = (fm.averageLatency*time.Duration(totalRequests-1) + latency) / time.Duration(totalRequests)
	}
}

// GetStats returns fallback statistics
func (fm *FallbackManager) GetStats() map[string]interface{} {
	fm.statsMu.RLock()
	defer fm.statsMu.RUnlock()
	
	fm.healthMu.RLock()
	defer fm.healthMu.RUnlock()
	
	successRate := float64(0)
	totalRequests := fm.successCount + fm.failureCount
	if totalRequests > 0 {
		successRate = float64(fm.successCount) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"enabled":               fm.fallbackEnabled,
		"local_healthy":         fm.localHealthy,
		"local_url":             fm.localURL.String(),
		"fallback_count":        fm.fallbackCount,
		"success_count":         fm.successCount,
		"failure_count":         fm.failureCount,
		"success_rate":          successRate,
		"average_latency":       fm.averageLatency.String(),
		"consecutive_failures":  fm.consecutiveFailures,
		"last_health_check":     fm.lastHealthCheck.Format(time.RFC3339),
		"health_check_interval": fm.healthCheckInterval.String(),
		"timeout":               fm.fallbackTimeout.String(),
		"max_retries":           fm.maxRetries,
	}
}

// SetEnabled enables or disables fallback
func (fm *FallbackManager) SetEnabled(enabled bool) {
	fm.fallbackEnabled = enabled
	slog.Info("Fallback mode changed", "enabled", enabled)
}

// SetTimeout sets the fallback timeout
func (fm *FallbackManager) SetTimeout(timeout time.Duration) {
	fm.fallbackTimeout = timeout
}

// SetMaxRetries sets the maximum number of retries
func (fm *FallbackManager) SetMaxRetries(retries int) {
	fm.maxRetries = retries
}

// SetHealthCheckInterval sets the health check interval
func (fm *FallbackManager) SetHealthCheckInterval(interval time.Duration) {
	fm.healthCheckInterval = interval
}

// Reset resets fallback statistics
func (fm *FallbackManager) Reset() {
	fm.statsMu.Lock()
	defer fm.statsMu.Unlock()
	
	fm.fallbackCount = 0
	fm.successCount = 0
	fm.failureCount = 0
	fm.averageLatency = 0
	
	slog.Info("Fallback statistics reset")
}

// StandaloneMode represents standalone mode configuration
type StandaloneMode struct {
	enabled      bool
	reason       string
	enabledAt    time.Time
	fallbackMgr  *FallbackManager
}

// NewStandaloneMode creates a new standalone mode manager
func NewStandaloneMode(fallbackMgr *FallbackManager) *StandaloneMode {
	return &StandaloneMode{
		enabled:     false,
		fallbackMgr: fallbackMgr,
	}
}

// Enable enables standalone mode
func (sm *StandaloneMode) Enable(reason string) {
	sm.enabled = true
	sm.reason = reason
	sm.enabledAt = time.Now()
	
	slog.Info("Standalone mode enabled", "reason", reason)
}

// Disable disables standalone mode
func (sm *StandaloneMode) Disable() {
	sm.enabled = false
	sm.reason = ""
	
	slog.Info("Standalone mode disabled")
}

// IsEnabled returns whether standalone mode is enabled
func (sm *StandaloneMode) IsEnabled() bool {
	return sm.enabled
}

// GetReason returns the reason for standalone mode
func (sm *StandaloneMode) GetReason() string {
	return sm.reason
}

// HandleRequest handles a request in standalone mode
func (sm *StandaloneMode) HandleRequest(c *gin.Context) {
	if !sm.enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Standalone mode is not enabled",
		})
		return
	}
	
	// Add standalone headers
	c.Header("X-Ollama-Standalone", "true")
	c.Header("X-Ollama-Standalone-Reason", sm.reason)
	c.Header("X-Ollama-Standalone-Since", sm.enabledAt.Format(time.RFC3339))
	
	// Execute request using fallback manager
	if err := sm.fallbackMgr.ExecuteFallback(c, "standalone-mode"); err != nil {
		slog.Error("Standalone request failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Standalone request failed",
			"details": err.Error(),
		})
	}
}

// GetStats returns standalone mode statistics
func (sm *StandaloneMode) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":    sm.enabled,
		"reason":     sm.reason,
		"enabled_at": sm.enabledAt.Format(time.RFC3339),
		"duration":   time.Since(sm.enabledAt).String(),
	}
}

// FallbackChain represents a chain of fallback mechanisms
type FallbackChain struct {
	fallbacks []FallbackHandler
}

// FallbackHandler represents a fallback handler
type FallbackHandler interface {
	CanHandle(reason string) bool
	Handle(c *gin.Context, reason string) error
	GetName() string
}

// NewFallbackChain creates a new fallback chain
func NewFallbackChain() *FallbackChain {
	return &FallbackChain{
		fallbacks: make([]FallbackHandler, 0),
	}
}

// AddFallback adds a fallback handler to the chain
func (fc *FallbackChain) AddFallback(handler FallbackHandler) {
	fc.fallbacks = append(fc.fallbacks, handler)
}

// Execute executes the fallback chain
func (fc *FallbackChain) Execute(c *gin.Context, reason string) error {
	for _, handler := range fc.fallbacks {
		if handler.CanHandle(reason) {
			slog.Debug("Executing fallback handler", "handler", handler.GetName(), "reason", reason)
			return handler.Handle(c, reason)
		}
	}
	
	return fmt.Errorf("no fallback handler available for reason: %s", reason)
}

// LocalFallbackHandler implements fallback to local Ollama
type LocalFallbackHandler struct {
	fallbackMgr *FallbackManager
}

// NewLocalFallbackHandler creates a new local fallback handler
func NewLocalFallbackHandler(fallbackMgr *FallbackManager) *LocalFallbackHandler {
	return &LocalFallbackHandler{
		fallbackMgr: fallbackMgr,
	}
}

// CanHandle checks if this handler can handle the given reason
func (lfh *LocalFallbackHandler) CanHandle(reason string) bool {
	return lfh.fallbackMgr.ShouldFallback(reason)
}

// Handle handles the fallback request
func (lfh *LocalFallbackHandler) Handle(c *gin.Context, reason string) error {
	return lfh.fallbackMgr.ExecuteFallback(c, reason)
}

// GetName returns the handler name
func (lfh *LocalFallbackHandler) GetName() string {
	return "local-fallback"
}

// CachedResponseHandler implements fallback to cached responses
type CachedResponseHandler struct {
	cache map[string]*api.GenerateResponse
	mu    sync.RWMutex
}

// NewCachedResponseHandler creates a new cached response handler
func NewCachedResponseHandler() *CachedResponseHandler {
	return &CachedResponseHandler{
		cache: make(map[string]*api.GenerateResponse),
	}
}

// CanHandle checks if this handler can handle the given reason
func (crh *CachedResponseHandler) CanHandle(reason string) bool {
	// Only handle timeout and temporary errors
	return reason == "timeout" || reason == "temporary-error"
}

// Handle handles the fallback request
func (crh *CachedResponseHandler) Handle(c *gin.Context, reason string) error {
	// Try to find cached response
	key := crh.generateCacheKey(c)
	
	crh.mu.RLock()
	cached, exists := crh.cache[key]
	crh.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("no cached response available")
	}
	
	// Return cached response
	c.Header("X-Ollama-Cached-Response", "true")
	c.Header("X-Ollama-Cache-Key", key)
	c.JSON(http.StatusOK, cached)
	
	return nil
}

// GetName returns the handler name
func (crh *CachedResponseHandler) GetName() string {
	return "cached-response"
}

// generateCacheKey generates a cache key for the request
func (crh *CachedResponseHandler) generateCacheKey(c *gin.Context) string {
	// Simple cache key based on path and method
	return fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)
}

// CacheResponse caches a response
func (crh *CachedResponseHandler) CacheResponse(key string, response *api.GenerateResponse) {
	crh.mu.Lock()
	defer crh.mu.Unlock()
	
	crh.cache[key] = response
}

// ClearCache clears the cache
func (crh *CachedResponseHandler) ClearCache() {
	crh.mu.Lock()
	defer crh.mu.Unlock()
	
	crh.cache = make(map[string]*api.GenerateResponse)
}