package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/auth"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Server represents the API server
type Server struct {
	config   *config.Config
	db       *database.DatabaseManager
	jwtSvc   *auth.JWTService
	logger   *slog.Logger
	server   *http.Server
	websocket *WebSocketHub
	registry *prometheus.Registry
	tracerProvider *sdktrace.TracerProvider

	// Prometheus metrics
	httpRequestsTotal          *prometheus.CounterVec
	httpRequestDuration        *prometheus.HistogramVec
	httpRequestsInFlight       prometheus.Gauge
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, db *database.DatabaseManager, logger *slog.Logger) (*Server, error) {
	// Initialize JWT service
	jwtSvc, err := auth.NewJWTService(&cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT service: %w", err)
	}

	// Initialize WebSocket hub
	websocketHub := NewWebSocketHub(logger)

	// Initialize Prometheus registry
	registry := prometheus.NewRegistry()

	// Define Prometheus metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollamamax_api_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ollamamax_api_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ollamamax_api_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// Register metrics with the registry
	if err := registry.Register(httpRequestsTotal); err != nil {
		return nil, fmt.Errorf("failed to register http_requests_total: %w", err)
	}
	if err := registry.Register(httpRequestDuration); err != nil {
		return nil, fmt.Errorf("failed to register http_request_duration_seconds: %w", err)
	}
	if err := registry.Register(httpRequestsInFlight); err != nil {
		return nil, fmt.Errorf("failed to register http_requests_in_flight: %w", err)
	}

	// Register database metrics to the main registry
	if db != nil {
		if err := db.RegisterTo(registry); err != nil {
			logger.Warn("Failed to register database metrics", "error", err)
		}
	}

	// Note: When load balancer is available, create it with the shared registry:
	// loadBalancer := distributed.NewRoundRobinBalancerWithRegistry(registry)
	// or
	// loadBalancer := distributed.NewSmartLoadBalancerWithRegistry(registry)

	// Initialize OpenTelemetry/Jaeger tracing
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://localhost:14268/api/traces"
	}

	// Create Jaeger exporter
	jaegerExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
	)
	if err != nil {
		logger.Warn("Failed to create Jaeger exporter, tracing disabled", "error", err)
	}

	// Create TracerProvider
	var tracerProvider *sdktrace.TracerProvider
	if jaegerExporter != nil {
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(jaegerExporter),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName("ollamamax-api"),
				semconv.ServiceVersion("1.0.0"),
				attribute.String("environment", "production"),
			)),
		)

		// Set global TracerProvider
		otel.SetTracerProvider(tracerProvider)

		// Set global TextMapPropagator for context propagation
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		logger.Info("OpenTelemetry tracing initialized", "jaeger_endpoint", jaegerEndpoint)
	}

	server := &Server{
		config:              cfg,
		db:                  db,
		jwtSvc:              jwtSvc,
		logger:              logger,
		websocket:           websocketHub,
		registry:            registry,
		tracerProvider:      tracerProvider,
		httpRequestsTotal:   httpRequestsTotal,
		httpRequestDuration: httpRequestDuration,
		httpRequestsInFlight: httpRequestsInFlight,
	}

	return server, nil
}

// RegisterP2PMetrics registers P2P node metrics to the server's Prometheus registry
// This should be called after server creation if a P2P node is available
func (s *Server) RegisterP2PMetrics(p2pNode interface {
	RegisterTo(prometheus.Registerer) error
}) error {
	if err := p2pNode.RegisterTo(s.registry); err != nil {
		s.logger.Warn("Failed to register P2P metrics", "error", err)
		return fmt.Errorf("failed to register P2P metrics: %w", err)
	}
	s.logger.Info("P2P metrics registered successfully")
	return nil
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	// Create Gin router
	router := s.setupRouter()

	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.config.API.Listen,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start WebSocket hub
	go s.websocket.Run()

	s.logger.Info("Starting API server",
		"address", s.config.API.Listen,
		"tls_enabled", s.config.API.TLSEnabled)

	// Start server
	if s.config.API.TLSEnabled {
		return s.server.ListenAndServeTLS(s.config.API.CertFile, s.config.API.KeyFile)
	}
	return s.server.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")

	// Stop WebSocket hub
	s.websocket.Stop()

	// Shutdown TracerProvider to flush remaining spans
	if s.tracerProvider != nil {
		if err := s.tracerProvider.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown TracerProvider", "error", err)
		}
	}

	// Shutdown HTTP server
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// setupRouter configures the Gin router with middleware and routes
func (s *Server) setupRouter() *gin.Engine {
	// Set Gin mode based on environment
	if s.logger.Enabled(context.Background(), slog.LevelDebug) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware (tracing must be first to capture all requests)
	router.Use(s.tracingMiddleware())
	router.Use(s.loggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(s.prometheusMiddleware())
	router.Use(s.corsMiddleware())
	router.Use(s.securityMiddleware())

	// PERFORMANCE: Compression middleware for bandwidth optimization
	// TODO: Uncomment when github.com/gin-contrib/gzip is added to dependencies
	// router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Rate limiting middleware
	if s.config.API.RateLimit.Enabled {
		router.Use(s.rateLimitMiddleware())
	}

	// Health check endpoint (no auth required)
	router.GET("/health", s.healthHandler)

	// Metrics endpoints - all metrics are now in a single registry
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
		s.registry,
		promhttp.HandlerOpts{},
	)))
	router.GET("/metrics.json", s.metricsJSONHandler)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.loginHandler)
			auth.POST("/register", s.registerHandler)
			auth.POST("/refresh", s.refreshTokenHandler)
		}

		// Protected endpoints (require authentication)
		protected := v1.Group("/")
		protected.Use(auth.JWTAuthMiddleware(s.jwtSvc))
		{
			// User management
			users := protected.Group("/users")
			{
				users.GET("/profile", s.getUserProfileHandler)
				users.PUT("/profile", s.updateUserProfileHandler)
				users.POST("/logout", s.logoutHandler)
			}

			// Model management
			models := protected.Group("/models")
			{
				models.GET("/", s.listModelsHandler)
				models.POST("/", s.createModelHandler)
				models.GET("/:id", s.getModelHandler)
				models.PUT("/:id", s.updateModelHandler)
				models.DELETE("/:id", s.deleteModelHandler)
				models.GET("/:id/replicas", s.getModelReplicasHandler)
			}

			// Node management
			nodes := protected.Group("/nodes")
			{
				nodes.GET("/", s.listNodesHandler)
				nodes.GET("/:id", s.getNodeHandler)
				nodes.PUT("/:id", s.updateNodeHandler)
				nodes.DELETE("/:id", s.deleteNodeHandler)
				nodes.GET("/:id/health", s.getNodeHealthHandler)
			}

			// Inference endpoints
			inference := protected.Group("/inference")
			{
				inference.POST("/chat", s.chatHandler)
				inference.POST("/generate", s.generateHandler)
				inference.GET("/requests", s.listInferenceRequestsHandler)
				inference.GET("/requests/:id", s.getInferenceRequestHandler)
			}

			// System configuration
			system := protected.Group("/system")
			{
				system.GET("/config", s.getSystemConfigHandler)
				system.PUT("/config", s.updateSystemConfigHandler)
				system.GET("/stats", s.getSystemStatsHandler)
				system.GET("/audit", s.getAuditLogsHandler)
			}
		}
	}

	// WebSocket endpoints
	router.GET("/ws", s.websocketHandler)
	router.GET("/ws/inference/:id", s.inferenceWebsocketHandler)

	// Static file serving (for admin dashboard)
	router.Static("/static", "./web/dist")
	router.StaticFile("/", "./web/dist/index.html")

	return router
}

// tracingMiddleware extracts trace context and creates spans for HTTP requests
func (s *Server) tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tracing if TracerProvider is not initialized
		if s.tracerProvider == nil {
			c.Next()
			return
		}

		// Skip metrics endpoints to reduce noise
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/metrics.json" {
			c.Next()
			return
		}

		// Get tracer
		tracer := otel.Tracer("ollamamax-api")

		// Extract context from request headers (for distributed tracing)
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start a new span for the request
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if c.FullPath() == "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPTarget(c.Request.URL.Path),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPScheme(c.Request.URL.Scheme),
				semconv.HTTPUserAgent(c.Request.UserAgent()),
				semconv.HTTPClientIP(c.ClientIP()),
			),
		)
		defer span.End()

		// Store span in context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set response attributes
		span.SetAttributes(
			semconv.HTTPStatusCode(c.Writer.Status()),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Inject trace context into response headers for downstream services
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		// Record error if present
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.Bool("error", true))
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
		}
	}
}

// prometheusMiddleware tracks HTTP request metrics
func (s *Server) prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid recursion
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/metrics.json" {
			c.Next()
			return
		}

		// Increment in-flight requests
		s.httpRequestsInFlight.Inc()
		defer s.httpRequestsInFlight.Dec()

		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get endpoint path (normalized to avoid high cardinality)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		s.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		s.httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
	}
}
