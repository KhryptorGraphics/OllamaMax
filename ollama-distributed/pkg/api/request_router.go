package api

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RouterConfig configures the request router
type RouterConfig struct {
	Algorithm           LoadBalancingStrategy
	HealthCheckInterval time.Duration
	RequestTimeout      time.Duration
	MaxRetries          int
}

// RouterMetrics tracks request routing performance
type RouterMetrics struct {
	RoutingDecisions    int64     `json:"routing_decisions"`
	RoutingErrors       int64     `json:"routing_errors"`
	SuccessfulRoutes    int64     `json:"successful_routes"`
	FailedRoutes        int64     `json:"failed_routes"`
	AverageRouteTime    time.Duration `json:"average_route_time"`
	LastUpdated         time.Time `json:"last_updated"`
	mu                  sync.RWMutex
}

// Route represents a routing destination
type Route struct {
	ID          string                 `json:"id"`
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Target      string                 `json:"target"`
	Weight      int                    `json:"weight"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	HealthCheck *HealthCheckConfig     `json:"health_check"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// NewRequestRouter creates a new request router
func NewRequestRouter(config *RouterConfig) (*RequestRouter, error) {
	if config == nil {
		config = &RouterConfig{
			Algorithm:           LoadBalancingRoundRobin,
			HealthCheckInterval: 30 * time.Second,
			RequestTimeout:      30 * time.Second,
			MaxRetries:          3,
		}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	router := &RequestRouter{
		config:        config,
		routes:        make(map[string]*Route),
		loadBalancer:  config.Algorithm,
		metrics: &RouterMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return router, nil
}

// Start starts the request router
func (rr *RequestRouter) Start() error {
	// Start metrics collection
	rr.wg.Add(1)
	go rr.metricsLoop()
	
	return nil
}

// Stop stops the request router
func (rr *RequestRouter) Stop() error {
	rr.cancel()
	rr.wg.Wait()
	return nil
}

// AddRoute adds a new route
func (rr *RequestRouter) AddRoute(route *Route) error {
	if route == nil {
		return fmt.Errorf("route cannot be nil")
	}
	
	if route.ID == "" {
		route.ID = generateRouteID()
	}
	
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	
	rr.routesMu.Lock()
	defer rr.routesMu.Unlock()
	
	rr.routes[route.ID] = route
	return nil
}

// RemoveRoute removes a route
func (rr *RequestRouter) RemoveRoute(routeID string) error {
	rr.routesMu.Lock()
	defer rr.routesMu.Unlock()
	
	if _, exists := rr.routes[routeID]; !exists {
		return fmt.Errorf("route not found")
	}
	
	delete(rr.routes, routeID)
	return nil
}

// GetRoute returns a route by ID
func (rr *RequestRouter) GetRoute(routeID string) (*Route, bool) {
	rr.routesMu.RLock()
	defer rr.routesMu.RUnlock()
	
	route, exists := rr.routes[routeID]
	return route, exists
}

// GetAllRoutes returns all routes
func (rr *RequestRouter) GetAllRoutes() []*Route {
	rr.routesMu.RLock()
	defer rr.routesMu.RUnlock()
	
	routes := make([]*Route, 0, len(rr.routes))
	for _, route := range rr.routes {
		routes = append(routes, route)
	}
	
	return routes
}

// SelectRoute selects the best route for a request
func (rr *RequestRouter) SelectRoute(path, method string) (*Route, error) {
	start := time.Now()
	
	rr.metrics.mu.Lock()
	rr.metrics.RoutingDecisions++
	rr.metrics.mu.Unlock()
	
	// Get matching routes
	matchingRoutes := rr.getMatchingRoutes(path, method)
	if len(matchingRoutes) == 0 {
		rr.metrics.mu.Lock()
		rr.metrics.RoutingErrors++
		rr.metrics.mu.Unlock()
		return nil, fmt.Errorf("no matching routes found")
	}
	
	// Filter enabled and healthy routes
	availableRoutes := rr.filterAvailableRoutes(matchingRoutes)
	if len(availableRoutes) == 0 {
		rr.metrics.mu.Lock()
		rr.metrics.RoutingErrors++
		rr.metrics.mu.Unlock()
		return nil, fmt.Errorf("no available routes found")
	}
	
	// Select route based on load balancing algorithm
	selectedRoute, err := rr.selectByAlgorithm(availableRoutes)
	if err != nil {
		rr.metrics.mu.Lock()
		rr.metrics.RoutingErrors++
		rr.metrics.mu.Unlock()
		return nil, err
	}
	
	// Update metrics
	duration := time.Since(start)
	rr.metrics.mu.Lock()
	rr.metrics.SuccessfulRoutes++
	if rr.metrics.SuccessfulRoutes == 1 {
		rr.metrics.AverageRouteTime = duration
	} else {
		rr.metrics.AverageRouteTime = (rr.metrics.AverageRouteTime + duration) / 2
	}
	rr.metrics.LastUpdated = time.Now()
	rr.metrics.mu.Unlock()
	
	return selectedRoute, nil
}

// getMatchingRoutes returns routes that match the path and method
func (rr *RequestRouter) getMatchingRoutes(path, method string) []*Route {
	rr.routesMu.RLock()
	defer rr.routesMu.RUnlock()
	
	var matching []*Route
	for _, route := range rr.routes {
		if rr.routeMatches(route, path, method) {
			matching = append(matching, route)
		}
	}
	
	return matching
}

// routeMatches checks if a route matches the path and method
func (rr *RequestRouter) routeMatches(route *Route, path, method string) bool {
	// Simple matching logic - can be enhanced with pattern matching
	if route.Method != "" && route.Method != method {
		return false
	}
	
	if route.Path != "" && route.Path != path {
		// TODO: Implement pattern matching
		return false
	}
	
	return true
}

// filterAvailableRoutes filters routes that are enabled and healthy
func (rr *RequestRouter) filterAvailableRoutes(routes []*Route) []*Route {
	var available []*Route
	for _, route := range routes {
		if route.Enabled && rr.isRouteHealthy(route) {
			available = append(available, route)
		}
	}
	return available
}

// isRouteHealthy checks if a route is healthy
func (rr *RequestRouter) isRouteHealthy(route *Route) bool {
	// TODO: Implement health checking integration
	return true
}

// selectByAlgorithm selects a route based on the load balancing algorithm
func (rr *RequestRouter) selectByAlgorithm(routes []*Route) (*Route, error) {
	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes available")
	}
	
	switch rr.loadBalancer {
	case LoadBalancingRoundRobin:
		return rr.selectRoundRobin(routes), nil
	case LoadBalancingWeighted:
		return rr.selectWeighted(routes), nil
	case LoadBalancingLeastLoaded:
		return rr.selectLeastLoaded(routes), nil
	case LoadBalancingIPHash:
		return rr.selectIPHash(routes), nil
	default:
		return routes[0], nil
	}
}

// selectRoundRobin implements round-robin selection
func (rr *RequestRouter) selectRoundRobin(routes []*Route) *Route {
	// Simple round-robin based on current time
	index := int(time.Now().UnixNano()) % len(routes)
	return routes[index]
}

// selectWeighted implements weighted selection
func (rr *RequestRouter) selectWeighted(routes []*Route) *Route {
	totalWeight := 0
	for _, route := range routes {
		totalWeight += route.Weight
	}
	
	if totalWeight == 0 {
		return rr.selectRoundRobin(routes)
	}
	
	target := int(time.Now().UnixNano()) % totalWeight
	current := 0
	
	for _, route := range routes {
		current += route.Weight
		if current > target {
			return route
		}
	}
	
	return routes[len(routes)-1]
}

// selectLeastLoaded implements least-loaded selection
func (rr *RequestRouter) selectLeastLoaded(routes []*Route) *Route {
	// TODO: Implement actual load tracking
	// For now, use round-robin
	return rr.selectRoundRobin(routes)
}

// selectIPHash implements IP hash-based selection
func (rr *RequestRouter) selectIPHash(routes []*Route) *Route {
	// TODO: Implement IP hash-based selection
	// For now, use round-robin
	return rr.selectRoundRobin(routes)
}

// UpdateRoute updates an existing route
func (rr *RequestRouter) UpdateRoute(routeID string, updates *Route) error {
	rr.routesMu.Lock()
	defer rr.routesMu.Unlock()
	
	route, exists := rr.routes[routeID]
	if !exists {
		return fmt.Errorf("route not found")
	}
	
	// Update fields
	if updates.Path != "" {
		route.Path = updates.Path
	}
	if updates.Method != "" {
		route.Method = updates.Method
	}
	if updates.Target != "" {
		route.Target = updates.Target
	}
	if updates.Weight > 0 {
		route.Weight = updates.Weight
	}
	if updates.Priority > 0 {
		route.Priority = updates.Priority
	}
	route.Enabled = updates.Enabled
	if updates.HealthCheck != nil {
		route.HealthCheck = updates.HealthCheck
	}
	if updates.Metadata != nil {
		route.Metadata = updates.Metadata
	}
	
	route.UpdatedAt = time.Now()
	return nil
}

// GetMetrics returns request router metrics
func (rr *RequestRouter) GetMetrics() *RouterMetrics {
	rr.metrics.mu.RLock()
	defer rr.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *rr.metrics
	return &metrics
}

// SetLoadBalancingAlgorithm sets the load balancing algorithm
func (rr *RequestRouter) SetLoadBalancingAlgorithm(algorithm LoadBalancingStrategy) {
	rr.loadBalancer = algorithm
}

// GetLoadBalancingAlgorithm returns the current load balancing algorithm
func (rr *RequestRouter) GetLoadBalancingAlgorithm() LoadBalancingStrategy {
	return rr.loadBalancer
}

// metricsLoop runs the metrics collection loop
func (rr *RequestRouter) metricsLoop() {
	defer rr.wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-rr.ctx.Done():
			return
		case <-ticker.C:
			rr.updateMetrics()
		}
	}
}

// updateMetrics updates request router metrics
func (rr *RequestRouter) updateMetrics() {
	rr.metrics.mu.Lock()
	defer rr.metrics.mu.Unlock()
	
	rr.metrics.LastUpdated = time.Now()
}

// generateRouteID generates a unique route ID
func generateRouteID() string {
	return fmt.Sprintf("route_%d", time.Now().UnixNano())
}

// Reset resets the request router
func (rr *RequestRouter) Reset() {
	rr.routesMu.Lock()
	defer rr.routesMu.Unlock()
	
	rr.routes = make(map[string]*Route)
	
	rr.metrics.mu.Lock()
	rr.metrics.RoutingDecisions = 0
	rr.metrics.RoutingErrors = 0
	rr.metrics.SuccessfulRoutes = 0
	rr.metrics.FailedRoutes = 0
	rr.metrics.AverageRouteTime = 0
	rr.metrics.LastUpdated = time.Now()
	rr.metrics.mu.Unlock()
}
