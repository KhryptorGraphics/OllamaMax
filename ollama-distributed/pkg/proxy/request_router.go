package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RequestRouter handles routing of requests to appropriate instances
type RequestRouter struct {
	proxy *OllamaProxy

	// Route patterns
	routes map[string]*RouteHandler

	// Middleware
	middleware []Middleware

	// Metrics
	metrics *RouterMetrics

	mu sync.RWMutex
}

// RouteHandler handles a specific route pattern
type RouteHandler struct {
	Pattern     string
	Method      string
	Handler     http.HandlerFunc
	Middleware  []Middleware
	RequireAuth bool
}

// Middleware represents a middleware function
type Middleware func(http.Handler) http.Handler

// RouterMetrics tracks router performance
type RouterMetrics struct {
	TotalRequests  int64
	RoutedRequests int64
	FailedRequests int64
	AverageLatency time.Duration
	RouteHitCounts map[string]int64

	mu sync.RWMutex
}

// NewRequestRouter creates a new request router
func NewRequestRouter(proxy *OllamaProxy) *RequestRouter {
	router := &RequestRouter{
		proxy:  proxy,
		routes: make(map[string]*RouteHandler),
		metrics: &RouterMetrics{
			RouteHitCounts: make(map[string]int64),
		},
	}

	// Register default routes
	router.registerDefaultRoutes()

	return router
}

// registerDefaultRoutes registers default Ollama API routes
func (r *RequestRouter) registerDefaultRoutes() {
	// Ollama API routes
	r.RegisterRoute("GET", "/api/tags", r.handleListModels)
	r.RegisterRoute("POST", "/api/generate", r.handleGenerate)
	r.RegisterRoute("POST", "/api/chat", r.handleChat)
	r.RegisterRoute("POST", "/api/embeddings", r.handleEmbeddings)
	r.RegisterRoute("POST", "/api/pull", r.handlePullModel)
	r.RegisterRoute("POST", "/api/push", r.handlePushModel)
	r.RegisterRoute("DELETE", "/api/delete", r.handleDeleteModel)
	r.RegisterRoute("POST", "/api/copy", r.handleCopyModel)
	r.RegisterRoute("POST", "/api/create", r.handleCreateModel)
	r.RegisterRoute("GET", "/api/show", r.handleShowModel)

	// Health and status routes
	r.RegisterRoute("GET", "/api/version", r.handleVersion)
	r.RegisterRoute("GET", "/health", r.handleHealth)

	// Distributed system specific routes
	r.RegisterRoute("GET", "/api/v1/proxy/status", r.handleProxyStatus)
	r.RegisterRoute("GET", "/api/v1/proxy/instances", r.handleProxyInstances)
	r.RegisterRoute("GET", "/api/v1/proxy/metrics", r.handleProxyMetrics)
}

// RegisterRoute registers a new route
func (r *RequestRouter) RegisterRoute(method, pattern string, handler http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := method + ":" + pattern
	r.routes[key] = &RouteHandler{
		Pattern: pattern,
		Method:  method,
		Handler: handler,
	}

	log.Printf("Registered route: %s %s", method, pattern)
}

// ServeHTTP implements http.Handler interface
func (r *RequestRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	// Update metrics
	r.metrics.mu.Lock()
	r.metrics.TotalRequests++
	r.metrics.mu.Unlock()

	// Find matching route
	handler := r.findRoute(req.Method, req.URL.Path)
	if handler == nil {
		// Default to proxy behavior for unmatched routes
		if err := r.proxy.ProxyRequest(w, req); err != nil {
			http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
			r.recordFailure()
			return
		}
		r.recordSuccess(time.Since(startTime))
		return
	}

	// Execute handler
	handler.Handler(w, req)

	// Update metrics
	duration := time.Since(startTime)
	r.recordSuccess(duration)
	r.recordRouteHit(req.Method + ":" + req.URL.Path)
}

// findRoute finds a matching route for the request
func (r *RequestRouter) findRoute(method, path string) *RouteHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Exact match first
	key := method + ":" + path
	if handler, exists := r.routes[key]; exists {
		return handler
	}

	// Pattern matching
	for routeKey, handler := range r.routes {
		if r.matchRoute(routeKey, method, path) {
			return handler
		}
	}

	return nil
}

// matchRoute checks if a route matches the request
func (r *RequestRouter) matchRoute(routeKey, method, path string) bool {
	parts := strings.SplitN(routeKey, ":", 2)
	if len(parts) != 2 {
		return false
	}

	routeMethod, routePattern := parts[0], parts[1]

	// Method must match
	if routeMethod != method {
		return false
	}

	// Simple pattern matching (could be enhanced with regex)
	if routePattern == path {
		return true
	}

	// Check for wildcard patterns
	if strings.HasSuffix(routePattern, "*") {
		prefix := strings.TrimSuffix(routePattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// Route handlers

// handleListModels handles GET /api/tags
func (r *RequestRouter) handleListModels(w http.ResponseWriter, req *http.Request) {
	// Route to an instance and aggregate results
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to list models: %v", err), http.StatusInternalServerError)
	}
}

// handleGenerate handles POST /api/generate
func (r *RequestRouter) handleGenerate(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate: %v", err), http.StatusInternalServerError)
	}
}

// handleChat handles POST /api/chat
func (r *RequestRouter) handleChat(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to chat: %v", err), http.StatusInternalServerError)
	}
}

// handleEmbeddings handles POST /api/embeddings
func (r *RequestRouter) handleEmbeddings(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate embeddings: %v", err), http.StatusInternalServerError)
	}
}

// handlePullModel handles POST /api/pull
func (r *RequestRouter) handlePullModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to pull model: %v", err), http.StatusInternalServerError)
	}
}

// handlePushModel handles POST /api/push
func (r *RequestRouter) handlePushModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to push model: %v", err), http.StatusInternalServerError)
	}
}

// handleDeleteModel handles DELETE /api/delete
func (r *RequestRouter) handleDeleteModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete model: %v", err), http.StatusInternalServerError)
	}
}

// handleCopyModel handles POST /api/copy
func (r *RequestRouter) handleCopyModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to copy model: %v", err), http.StatusInternalServerError)
	}
}

// handleCreateModel handles POST /api/create
func (r *RequestRouter) handleCreateModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create model: %v", err), http.StatusInternalServerError)
	}
}

// handleShowModel handles GET /api/show
func (r *RequestRouter) handleShowModel(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to show model: %v", err), http.StatusInternalServerError)
	}
}

// handleVersion handles GET /api/version
func (r *RequestRouter) handleVersion(w http.ResponseWriter, req *http.Request) {
	// Route to best available instance
	if err := r.proxy.ProxyRequest(w, req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get version: %v", err), http.StatusInternalServerError)
	}
}

// handleHealth handles GET /health
func (r *RequestRouter) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","proxy":"running"}`))
}

// handleProxyStatus handles GET /api/v1/proxy/status
func (r *RequestRouter) handleProxyStatus(w http.ResponseWriter, req *http.Request) {
	instances := r.proxy.GetInstances()

	status := map[string]interface{}{
		"status":            "running",
		"instance_count":    len(instances),
		"healthy_instances": r.countHealthyInstances(instances),
		"total_requests":    r.metrics.TotalRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON encoding
	fmt.Fprintf(w, `{"status":"%s","instance_count":%d,"healthy_instances":%d,"total_requests":%d}`,
		status["status"], status["instance_count"], status["healthy_instances"], status["total_requests"])
}

// handleProxyInstances handles GET /api/v1/proxy/instances
func (r *RequestRouter) handleProxyInstances(w http.ResponseWriter, req *http.Request) {
	instances := r.proxy.GetInstances()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	fmt.Fprintf(w, `{"instances":%d}`, len(instances))
}

// handleProxyMetrics handles GET /api/v1/proxy/metrics
func (r *RequestRouter) handleProxyMetrics(w http.ResponseWriter, req *http.Request) {
	metrics := r.proxy.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Simple JSON response
	fmt.Fprintf(w, `{"total_requests":%d,"successful_requests":%d,"failed_requests":%d}`,
		metrics.TotalRequests, metrics.SuccessfulRequests, metrics.FailedRequests)
}

// Helper methods

// countHealthyInstances counts healthy instances
func (r *RequestRouter) countHealthyInstances(instances map[string]*OllamaInstance) int {
	count := 0
	for _, instance := range instances {
		if instance.Status == InstanceStatusHealthy {
			count++
		}
	}
	return count
}

// recordSuccess records a successful request
func (r *RequestRouter) recordSuccess(duration time.Duration) {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.RoutedRequests++

	// Update average latency
	if r.metrics.RoutedRequests == 1 {
		r.metrics.AverageLatency = duration
	} else {
		r.metrics.AverageLatency = (r.metrics.AverageLatency + duration) / 2
	}
}

// recordFailure records a failed request
func (r *RequestRouter) recordFailure() {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.FailedRequests++
}

// recordRouteHit records a hit for a specific route
func (r *RequestRouter) recordRouteHit(route string) {
	r.metrics.mu.Lock()
	defer r.metrics.mu.Unlock()

	r.metrics.RouteHitCounts[route]++
}
