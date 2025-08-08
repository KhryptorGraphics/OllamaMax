package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
)

// APIGatewayManager manages the complete API gateway system
type APIGatewayManager struct {
	config *APIGatewayConfig

	// Core components
	httpServer      *HTTPServer
	webSocketServer *WebSocketServer
	authManager     *AuthManager
	rateLimiter     *RateLimiter
	requestRouter   *RequestRouter
	healthChecker   *HealthChecker

	// Integration components
	p2pNode          *p2p.Node
	consensusManager *consensus.ConsensusManager
	schedulerManager *scheduler.SchedulerManager
	messageRouter    *messaging.MessageRouter
	networkMonitor   *monitoring.NetworkMonitor

	// Legacy integration
	routeIntegration  *RouteIntegration
	distributedRoutes *DistributedRoutes
	integrationLayer  *IntegrationLayer

	// State management
	state   *GatewayState
	stateMu sync.RWMutex

	// Metrics and monitoring
	metrics *GatewayMetrics

	// Lifecycle
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	started   bool
	startedMu sync.RWMutex
}

// APIGatewayConfig configures the API gateway
type APIGatewayConfig struct {
	// Basic settings
	Listen     string
	TLSEnabled bool
	CertFile   string
	KeyFile    string

	// API settings
	APIConfig *config.APIConfig

	// Authentication settings
	AuthEnabled bool
	JWTSecret   string
	TokenExpiry time.Duration

	// Rate limiting settings
	RateLimitEnabled  bool
	RequestsPerSecond int
	BurstSize         int

	// Request routing settings
	RoutingAlgorithm    string
	HealthCheckInterval time.Duration
	RequestTimeout      time.Duration

	// WebSocket settings
	WSEnabled         bool
	WSReadBufferSize  int
	WSWriteBufferSize int

	// Performance settings
	MaxConnections int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration

	// Integration settings
	EnableDistributed bool
	EnableFallback    bool
	LocalOllamaURL    string

	// Monitoring settings
	MetricsEnabled  bool
	MetricsInterval time.Duration
}

// GatewayState represents the current state of the API gateway
type GatewayState struct {
	// Gateway status
	Status    GatewayStatus `json:"status"`
	StartedAt time.Time     `json:"started_at"`

	// Connection statistics
	ActiveConnections    int64 `json:"active_connections"`
	TotalConnections     int64 `json:"total_connections"`
	WebSocketConnections int64 `json:"websocket_connections"`

	// Request statistics
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`

	// Health status
	HealthStatus    HealthStatus `json:"health_status"`
	LastHealthCheck time.Time    `json:"last_health_check"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// GatewayMetrics tracks API gateway performance
type GatewayMetrics struct {
	// Request metrics
	RequestsTotal   int64         `json:"requests_total"`
	RequestsSuccess int64         `json:"requests_success"`
	RequestsError   int64         `json:"requests_error"`
	RequestDuration time.Duration `json:"request_duration"`

	// Connection metrics
	ConnectionsActive int64 `json:"connections_active"`
	ConnectionsTotal  int64 `json:"connections_total"`
	ConnectionErrors  int64 `json:"connection_errors"`

	// Rate limiting metrics
	RateLimitHits    int64 `json:"rate_limit_hits"`
	RateLimitBlocked int64 `json:"rate_limit_blocked"`

	// Authentication metrics
	AuthAttempts int64 `json:"auth_attempts"`
	AuthSuccess  int64 `json:"auth_success"`
	AuthFailures int64 `json:"auth_failures"`

	// Routing metrics
	RoutingDecisions int64 `json:"routing_decisions"`
	RoutingErrors    int64 `json:"routing_errors"`
	FallbackRequests int64 `json:"fallback_requests"`

	// Performance metrics
	ResponseTime time.Duration `json:"response_time"`
	Throughput   float64       `json:"throughput"`
	ErrorRate    float64       `json:"error_rate"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`
	mu          sync.RWMutex
}

// HTTPServer manages HTTP/REST API endpoints
type HTTPServer struct {
	config *HTTPServerConfig
	router *gin.Engine
	server *http.Server

	// Middleware
	authMiddleware      gin.HandlerFunc
	rateLimitMiddleware gin.HandlerFunc
	corsMiddleware      gin.HandlerFunc

	// Metrics
	metrics *HTTPMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// WebSocketServer manages WebSocket connections
type WebSocketServer struct {
	config   *WebSocketConfig
	upgrader websocket.Upgrader
	hub      *WSHub

	// Connection management
	connections   map[string]*WSConnection
	connectionsMu sync.RWMutex

	// Metrics
	metrics *WebSocketMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// FailedAttempts tracks failed authentication attempts for rate limiting
type FailedAttempts struct {
	Count        int
	LastAttempt  time.Time
	BlockedUntil time.Time
}

// AuthManager handles authentication and authorization
type AuthManager struct {
	config *AuthConfig

	// JWT management
	jwtSecret   []byte
	tokenExpiry time.Duration

	// User management
	users   map[string]*User
	usersMu sync.RWMutex

	// Session management
	sessions   map[string]*Session
	sessionsMu sync.RWMutex

	// Rate limiting
	failedAttempts   map[string]*FailedAttempts
	failedAttemptsMu sync.RWMutex

	// Metrics
	metrics *AuthMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// RateLimiter handles request rate limiting
type RateLimiter struct {
	config *RateLimitConfig

	// Rate limiting state
	buckets   map[string]*TokenBucket
	bucketsMu sync.RWMutex

	// Metrics
	metrics *RateLimitMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// RequestRouter handles request routing and load balancing
type RequestRouter struct {
	config *RouterConfig

	// Routing state
	routes   map[string]*Route
	routesMu sync.RWMutex

	// Load balancing
	loadBalancer LoadBalancingStrategy

	// Health checking
	healthChecker *HealthChecker

	// Metrics
	metrics *RouterMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// HealthChecker monitors backend service health
type HealthChecker struct {
	config *HealthCheckConfig

	// Health state
	services   map[string]*ServiceHealth
	servicesMu sync.RWMutex

	// Metrics
	metrics *HealthMetrics

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Enums and constants
type GatewayStatus string

const (
	GatewayStatusStarting GatewayStatus = "starting"
	GatewayStatusRunning  GatewayStatus = "running"
	GatewayStatusStopping GatewayStatus = "stopping"
	GatewayStatusStopped  GatewayStatus = "stopped"
	GatewayStatusError    GatewayStatus = "error"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin  LoadBalancingStrategy = "round_robin"
	LoadBalancingLeastLoaded LoadBalancingStrategy = "least_loaded"
	LoadBalancingWeighted    LoadBalancingStrategy = "weighted"
	LoadBalancingIPHash      LoadBalancingStrategy = "ip_hash"
)

// NewAPIGatewayManager creates a new API gateway manager
func NewAPIGatewayManager(config *APIGatewayConfig, p2pNode *p2p.Node, consensusManager *consensus.ConsensusManager, schedulerManager *scheduler.SchedulerManager, messageRouter *messaging.MessageRouter, networkMonitor *monitoring.NetworkMonitor) (*APIGatewayManager, error) {
	if config == nil {
		config = &APIGatewayConfig{
			Listen:              ":8080",
			AuthEnabled:         true,
			TokenExpiry:         24 * time.Hour,
			RateLimitEnabled:    true,
			RequestsPerSecond:   100,
			BurstSize:           200,
			RoutingAlgorithm:    "least_loaded",
			HealthCheckInterval: 30 * time.Second,
			RequestTimeout:      30 * time.Second,
			WSEnabled:           true,
			WSReadBufferSize:    1024,
			WSWriteBufferSize:   1024,
			MaxConnections:      10000,
			ReadTimeout:         30 * time.Second,
			WriteTimeout:        30 * time.Second,
			IdleTimeout:         120 * time.Second,
			EnableDistributed:   true,
			EnableFallback:      true,
			LocalOllamaURL:      "http://localhost:11434",
			MetricsEnabled:      true,
			MetricsInterval:     30 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &APIGatewayManager{
		config:           config,
		p2pNode:          p2pNode,
		consensusManager: consensusManager,
		schedulerManager: schedulerManager,
		messageRouter:    messageRouter,
		networkMonitor:   networkMonitor,
		state: &GatewayState{
			Status:      GatewayStatusStopped,
			StartedAt:   time.Now(),
			LastUpdated: time.Now(),
		},
		metrics: &GatewayMetrics{
			LastUpdated: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Create core components
	if err := manager.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Setup legacy integration
	if err := manager.setupLegacyIntegration(); err != nil {
		return nil, fmt.Errorf("failed to setup legacy integration: %w", err)
	}

	return manager, nil
}

// initializeComponents initializes the core API gateway components
func (gm *APIGatewayManager) initializeComponents() error {
	// Create HTTP server
	httpServer, err := NewHTTPServer(&HTTPServerConfig{
		Listen:       gm.config.Listen,
		TLSEnabled:   gm.config.TLSEnabled,
		CertFile:     gm.config.CertFile,
		KeyFile:      gm.config.KeyFile,
		ReadTimeout:  gm.config.ReadTimeout,
		WriteTimeout: gm.config.WriteTimeout,
		IdleTimeout:  gm.config.IdleTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create HTTP server: %w", err)
	}
	gm.httpServer = httpServer

	// Create WebSocket server if enabled
	if gm.config.WSEnabled {
		wsServer, err := NewWebSocketServer(&WebSocketConfig{
			ReadBufferSize:  gm.config.WSReadBufferSize,
			WriteBufferSize: gm.config.WSWriteBufferSize,
		})
		if err != nil {
			return fmt.Errorf("failed to create WebSocket server: %w", err)
		}
		gm.webSocketServer = wsServer
	}

	// Create auth manager if enabled
	if gm.config.AuthEnabled {
		authManager, err := NewAuthManager(&AuthConfig{
			JWTSecret:   gm.config.JWTSecret,
			TokenExpiry: gm.config.TokenExpiry,
		})
		if err != nil {
			return fmt.Errorf("failed to create auth manager: %w", err)
		}
		gm.authManager = authManager
	}

	// Create rate limiter if enabled
	if gm.config.RateLimitEnabled {
		rateLimiter, err := NewRateLimiter(&RateLimitConfig{
			RequestsPerSecond: gm.config.RequestsPerSecond,
			BurstSize:         gm.config.BurstSize,
		})
		if err != nil {
			return fmt.Errorf("failed to create rate limiter: %w", err)
		}
		gm.rateLimiter = rateLimiter
	}

	// Create request router
	requestRouter, err := NewRequestRouter(&RouterConfig{
		Algorithm:           LoadBalancingStrategy(gm.config.RoutingAlgorithm),
		HealthCheckInterval: gm.config.HealthCheckInterval,
		RequestTimeout:      gm.config.RequestTimeout,
	})
	if err != nil {
		return fmt.Errorf("failed to create request router: %w", err)
	}
	gm.requestRouter = requestRouter

	// Create health checker
	healthChecker, err := NewHealthChecker(&HealthCheckConfig{
		CheckInterval: gm.config.HealthCheckInterval,
		Timeout:       10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create health checker: %w", err)
	}
	gm.healthChecker = healthChecker

	return nil
}

// setupLegacyIntegration sets up integration with existing API components
func (gm *APIGatewayManager) setupLegacyIntegration() error {
	// This would integrate with the existing RouteIntegration, DistributedRoutes, etc.
	// For now, we'll create placeholder implementations
	return nil
}

// Start starts the API gateway manager
func (gm *APIGatewayManager) Start() error {
	gm.startedMu.Lock()
	defer gm.startedMu.Unlock()

	if gm.started {
		return nil
	}

	gm.state.Status = GatewayStatusStarting

	// Start core components
	if gm.authManager != nil {
		if err := gm.authManager.Start(); err != nil {
			return fmt.Errorf("failed to start auth manager: %w", err)
		}
	}

	if gm.rateLimiter != nil {
		if err := gm.rateLimiter.Start(); err != nil {
			return fmt.Errorf("failed to start rate limiter: %w", err)
		}
	}

	if err := gm.requestRouter.Start(); err != nil {
		return fmt.Errorf("failed to start request router: %w", err)
	}

	if err := gm.healthChecker.Start(); err != nil {
		return fmt.Errorf("failed to start health checker: %w", err)
	}

	if gm.webSocketServer != nil {
		if err := gm.webSocketServer.Start(); err != nil {
			return fmt.Errorf("failed to start WebSocket server: %w", err)
		}
	}

	if err := gm.httpServer.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start monitoring if enabled
	if gm.config.MetricsEnabled {
		gm.wg.Add(1)
		go gm.monitoringLoop()
	}

	// Update state
	gm.stateMu.Lock()
	gm.state.Status = GatewayStatusRunning
	gm.state.StartedAt = time.Now()
	gm.state.LastUpdated = time.Now()
	gm.stateMu.Unlock()

	gm.started = true
	return nil
}

// Stop stops the API gateway manager
func (gm *APIGatewayManager) Stop() error {
	gm.startedMu.Lock()
	defer gm.startedMu.Unlock()

	if !gm.started {
		return nil
	}

	gm.stateMu.Lock()
	gm.state.Status = GatewayStatusStopping
	gm.state.LastUpdated = time.Now()
	gm.stateMu.Unlock()

	gm.cancel()

	// Stop components
	if gm.httpServer != nil {
		gm.httpServer.Stop()
	}

	if gm.webSocketServer != nil {
		gm.webSocketServer.Stop()
	}

	if gm.healthChecker != nil {
		gm.healthChecker.Stop()
	}

	if gm.requestRouter != nil {
		gm.requestRouter.Stop()
	}

	if gm.rateLimiter != nil {
		gm.rateLimiter.Stop()
	}

	if gm.authManager != nil {
		gm.authManager.Stop()
	}

	gm.wg.Wait()

	gm.stateMu.Lock()
	gm.state.Status = GatewayStatusStopped
	gm.state.LastUpdated = time.Now()
	gm.stateMu.Unlock()

	gm.started = false
	return nil
}

// GetState returns the current gateway state
func (gm *APIGatewayManager) GetState() *GatewayState {
	gm.stateMu.RLock()
	defer gm.stateMu.RUnlock()

	// Create a copy
	state := *gm.state
	return &state
}

// GetMetrics returns the current gateway metrics
func (gm *APIGatewayManager) GetMetrics() *GatewayMetrics {
	gm.metrics.mu.RLock()
	defer gm.metrics.mu.RUnlock()

	// Create a copy
	metrics := *gm.metrics
	return &metrics
}

// IsHealthy returns whether the gateway is healthy
func (gm *APIGatewayManager) IsHealthy() bool {
	gm.stateMu.RLock()
	defer gm.stateMu.RUnlock()

	return gm.state.HealthStatus == HealthStatusHealthy
}

// monitoringLoop runs the monitoring loop
func (gm *APIGatewayManager) monitoringLoop() {
	defer gm.wg.Done()

	ticker := time.NewTicker(gm.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-gm.ctx.Done():
			return
		case <-ticker.C:
			gm.collectMetrics()
		}
	}
}

// collectMetrics collects and updates metrics
func (gm *APIGatewayManager) collectMetrics() {
	// Update gateway metrics
	gm.updateGatewayMetrics()

	// Report to network monitor if available
	if gm.networkMonitor != nil {
		gm.reportToNetworkMonitor()
	}
}

// updateGatewayMetrics updates overall gateway metrics
func (gm *APIGatewayManager) updateGatewayMetrics() {
	// Get metrics from components
	var httpMetrics *HTTPMetrics
	if gm.httpServer != nil {
		httpMetrics = gm.httpServer.GetMetrics()
	}

	var wsMetrics *WebSocketMetrics
	if gm.webSocketServer != nil {
		wsMetrics = gm.webSocketServer.GetMetrics()
	}

	var authMetrics *AuthMetrics
	if gm.authManager != nil {
		authMetrics = gm.authManager.GetMetrics()
	}

	var rateLimitMetrics *RateLimitMetrics
	if gm.rateLimiter != nil {
		rateLimitMetrics = gm.rateLimiter.GetMetrics()
	}

	var routerMetrics *RouterMetrics
	if gm.requestRouter != nil {
		routerMetrics = gm.requestRouter.GetMetrics()
	}

	// Update gateway metrics
	gm.metrics.mu.Lock()
	defer gm.metrics.mu.Unlock()

	if httpMetrics != nil {
		gm.metrics.RequestsTotal = httpMetrics.RequestsTotal
		gm.metrics.RequestsSuccess = httpMetrics.RequestsSuccess
		gm.metrics.RequestsError = httpMetrics.RequestsError
		gm.metrics.ResponseTime = httpMetrics.AverageResponseTime
		gm.metrics.ConnectionsActive = httpMetrics.ConnectionsActive
		gm.metrics.ConnectionsTotal = httpMetrics.ConnectionsTotal
	}

	if wsMetrics != nil {
		// Add WebSocket connection metrics
		gm.metrics.ConnectionsActive += wsMetrics.ConnectionsActive
		gm.metrics.ConnectionsTotal += wsMetrics.ConnectionsTotal
	}

	if authMetrics != nil {
		gm.metrics.AuthAttempts = authMetrics.AuthAttempts
		gm.metrics.AuthSuccess = authMetrics.AuthSuccess
		gm.metrics.AuthFailures = authMetrics.AuthFailures
	}

	if rateLimitMetrics != nil {
		gm.metrics.RateLimitHits = rateLimitMetrics.RequestsAllowed
		gm.metrics.RateLimitBlocked = rateLimitMetrics.RequestsBlocked
	}

	if routerMetrics != nil {
		gm.metrics.RoutingDecisions = routerMetrics.RoutingDecisions
		gm.metrics.RoutingErrors = routerMetrics.RoutingErrors
	}

	// Calculate derived metrics
	if gm.metrics.RequestsTotal > 0 {
		gm.metrics.ErrorRate = float64(gm.metrics.RequestsError) / float64(gm.metrics.RequestsTotal)
		gm.metrics.Throughput = float64(gm.metrics.RequestsTotal) / time.Since(gm.state.StartedAt).Seconds()
	}

	gm.metrics.LastUpdated = time.Now()

	// Update state
	gm.stateMu.Lock()
	gm.state.TotalRequests = gm.metrics.RequestsTotal
	gm.state.SuccessfulRequests = gm.metrics.RequestsSuccess
	gm.state.FailedRequests = gm.metrics.RequestsError
	gm.state.ActiveConnections = gm.metrics.ConnectionsActive
	gm.state.TotalConnections = gm.metrics.ConnectionsTotal
	gm.state.AverageResponseTime = gm.metrics.ResponseTime
	gm.state.RequestsPerSecond = gm.metrics.Throughput
	gm.state.LastUpdated = time.Now()

	// Update health status based on metrics
	if gm.metrics.ErrorRate > 0.1 { // More than 10% error rate
		gm.state.HealthStatus = HealthStatusDegraded
	} else if gm.metrics.ErrorRate > 0.05 { // More than 5% error rate
		gm.state.HealthStatus = HealthStatusDegraded
	} else {
		gm.state.HealthStatus = HealthStatusHealthy
	}

	gm.stateMu.Unlock()
}

// reportToNetworkMonitor reports metrics to the network monitor
func (gm *APIGatewayManager) reportToNetworkMonitor() {
	// This would integrate with the network monitor to report API gateway metrics
	// Implementation would depend on the network monitor interface
}

// GetHTTPServer returns the HTTP server
func (gm *APIGatewayManager) GetHTTPServer() *HTTPServer {
	return gm.httpServer
}

// GetWebSocketServer returns the WebSocket server
func (gm *APIGatewayManager) GetWebSocketServer() *WebSocketServer {
	return gm.webSocketServer
}

// GetAuthManager returns the authentication manager
func (gm *APIGatewayManager) GetAuthManager() *AuthManager {
	return gm.authManager
}

// GetRateLimiter returns the rate limiter
func (gm *APIGatewayManager) GetRateLimiter() *RateLimiter {
	return gm.rateLimiter
}

// GetRequestRouter returns the request router
func (gm *APIGatewayManager) GetRequestRouter() *RequestRouter {
	return gm.requestRouter
}

// GetHealthChecker returns the health checker
func (gm *APIGatewayManager) GetHealthChecker() *HealthChecker {
	return gm.healthChecker
}

// GetSchedulerManager returns the scheduler manager
func (gm *APIGatewayManager) GetSchedulerManager() *scheduler.SchedulerManager {
	return gm.schedulerManager
}

// GetConsensusManager returns the consensus manager
func (gm *APIGatewayManager) GetConsensusManager() *consensus.ConsensusManager {
	return gm.consensusManager
}

// IsLeader returns whether this node is the API gateway leader
func (gm *APIGatewayManager) IsLeader() bool {
	if gm.consensusManager == nil {
		return true // Single node mode
	}

	return gm.consensusManager.IsLeader()
}
