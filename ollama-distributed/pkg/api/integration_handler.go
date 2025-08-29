package api

import (
	_ "context"
	_ "encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	// "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api/middleware"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/database"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/monitoring"
)

// Prometheus metrics for API integration
var (
	apiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollama_api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	apiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ollama_api_request_duration_seconds",
			Help:    "Duration of API requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	databaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ollama_database_connections",
			Help: "Number of active database connections",
		},
		[]string{"database", "state"},
	)

	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollama_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollama_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)
)

// IntegrationHandler handles all external integration endpoints
type IntegrationHandler struct {
	server   *Server
	tracer   trace.Tracer
	database *database.Manager
	monitor  *monitoring.MetricsCollector
}

// NewIntegrationHandler creates a new integration handler
func NewIntegrationHandler(server *Server, db *database.Manager, monitor *monitoring.MetricsCollector) *IntegrationHandler {
	return &IntegrationHandler{
		server:   server,
		tracer:   otel.Tracer("ollama-api-integration"),
		database: db,
		monitor:  monitor,
	}
}

// SetupIntegrationRoutes sets up all integration API routes
func (h *IntegrationHandler) SetupIntegrationRoutes(r *gin.Engine) {
	// Apply common middleware
	// r.Use(middleware.RequestID())
	// r.Use(middleware.Logger())
	// r.Use(middleware.CORS())
	// r.Use(middleware.RateLimit())
	// r.Use(middleware.Authentication())
	r.Use(h.prometheusMiddleware())
	r.Use(h.tracingMiddleware())

	// Health and readiness endpoints
	health := r.Group("/")
	{
		health.GET("/health", h.healthCheck)
		health.GET("/ready", h.readinessCheck)
		health.GET("/metrics", h.metricsHandler)
	}

	// API versioned routes
	v1 := r.Group("/api/v1")
	{
		// Model management
		models := v1.Group("/models")
		{
			models.GET("", h.listModels)
			models.POST("", h.createModel)
			models.GET("/:id", h.getModel)
			models.PUT("/:id", h.updateModel)
			models.DELETE("/:id", h.deleteModel)
			models.POST("/:id/download", h.downloadModel)
			models.GET("/:id/status", h.getModelStatus)
		}

		// Inference endpoints
		inference := v1.Group("/inference")
		{
			inference.POST("/completions", h.createCompletion)
			inference.POST("/chat/completions", h.createChatCompletion)
			inference.POST("/embeddings", h.createEmbedding)
			inference.GET("/jobs/:id", h.getInferenceJob)
			inference.DELETE("/jobs/:id", h.cancelInferenceJob)
		}

		// Cluster management
		cluster := v1.Group("/cluster")
		{
			cluster.GET("/status", h.getClusterStatus)
			cluster.GET("/nodes", h.getNodes)
			cluster.GET("/nodes/:id", h.getNode)
			cluster.POST("/nodes/:id/drain", h.drainNode)
			cluster.POST("/nodes/:id/uncordon", h.uncordonNode)
		}

		// System administration
		admin := v1.Group("/admin")
		// admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", h.listUsers)
			admin.POST("/users", h.createUser)
			admin.GET("/users/:id", h.getUser)
			admin.PUT("/users/:id", h.updateUser)
			admin.DELETE("/users/:id", h.deleteUser)
			admin.GET("/audit-logs", h.getAuditLogs)
			admin.GET("/system-metrics", h.getSystemMetrics)
		}

		// Integration endpoints
		integrations := v1.Group("/integrations")
		{
			integrations.POST("/webhooks", h.createWebhook)
			integrations.GET("/webhooks", h.listWebhooks)
			integrations.DELETE("/webhooks/:id", h.deleteWebhook)
			integrations.POST("/external-api", h.proxyExternalAPI)
		}
	}

	// WebSocket endpoints
	ws := r.Group("/ws")
	{
		ws.GET("/inference/:id", h.inferenceWebSocket)
		ws.GET("/cluster-events", h.clusterEventsWebSocket)
		ws.GET("/logs", h.logsWebSocket)
	}

	// GraphQL endpoint (optional)
	r.POST("/graphql", h.graphqlHandler)

	// gRPC gateway (optional)
	r.Any("/grpc/*path", h.grpcGatewayHandler)
}

// Health check endpoint
func (h *IntegrationHandler) healthCheck(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "health_check")
	defer span.End()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now().Add(-time.Hour)).Seconds(), // TODO: implement proper start time tracking
		"checks":    make(map[string]interface{}),
	}

	// Check database
	if err := h.database.Health(ctx); err != nil {
		health["status"] = "unhealthy"
		health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		span.RecordError(err)
		span.SetAttributes(attribute.String("health.status", "unhealthy"))
	} else {
		health["checks"].(map[string]interface{})["database"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check Redis
	// TODO: implement Redis health check
	if false { // h.server.redis != nil {
		if err := error(nil); err != nil { // h.server.redis.Ping(ctx).Err(); err != nil {
			health["status"] = "unhealthy"
			health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			span.RecordError(err)
		} else {
			health["checks"].(map[string]interface{})["redis"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	// Check P2P connectivity
	if h.server.p2p != nil {
		peerCount := h.server.p2p.GetPeerCount()
		health["checks"].(map[string]interface{})["p2p"] = map[string]interface{}{
			"status":     "healthy",
			"peer_count": peerCount,
		}
	}

	// Check consensus
	if h.server.consensus != nil {
		isLeader := h.server.consensus.IsLeader()
		health["checks"].(map[string]interface{})["consensus"] = map[string]interface{}{
			"status":    "healthy",
			"is_leader": isLeader,
		}
	}

	span.SetAttributes(
		attribute.String("health.status", health["status"].(string)),
		attribute.Int("health.checks", len(health["checks"].(map[string]interface{}))),
	)

	if health["status"] == "healthy" {
		c.JSON(http.StatusOK, health)
	} else {
		c.JSON(http.StatusServiceUnavailable, health)
	}
}

// Readiness check endpoint
func (h *IntegrationHandler) readinessCheck(c *gin.Context) {
	_, span := h.tracer.Start(c.Request.Context(), "readiness_check")
	defer span.End()

	ready := map[string]interface{}{
		"ready":     true,
		"timestamp": time.Now().UTC(),
		"checks":    make(map[string]interface{}),
	}

	// Check if database migrations are complete
	if !h.database.MigrationsComplete() {
		ready["ready"] = false
		ready["checks"].(map[string]interface{})["migrations"] = map[string]interface{}{
			"ready":  false,
			"reason": "database migrations not complete",
		}
	}

	// Check if models are loaded
	// TODO: implement scheduler integration
	if false { // h.server.scheduler != nil {
		modelCount := 0 // h.server.scheduler.GetLoadedModelCount()
		if modelCount == 0 {
			ready["checks"].(map[string]interface{})["models"] = map[string]interface{}{
				"ready":  false,
				"reason": "no models loaded",
			}
		} else {
			ready["checks"].(map[string]interface{})["models"] = map[string]interface{}{
				"ready": true,
				"count": modelCount,
			}
		}
	}

	span.SetAttributes(attribute.Bool("readiness.ready", ready["ready"].(bool)))

	if ready["ready"].(bool) {
		c.JSON(http.StatusOK, ready)
	} else {
		c.JSON(http.StatusServiceUnavailable, ready)
	}
}

// Metrics handler for Prometheus
func (h *IntegrationHandler) metricsHandler(c *gin.Context) {
	// Update database connection metrics
	if h.database != nil {
		stats := h.database.GetStats()
		databaseConnections.WithLabelValues("postgres", "idle").Set(float64(stats.Idle))
		databaseConnections.WithLabelValues("postgres", "in_use").Set(float64(stats.InUse))
		databaseConnections.WithLabelValues("postgres", "open").Set(float64(stats.OpenConnections))
	}

	// Use prometheus default handler
	// TODO: implement metrics handler
	c.JSON(http.StatusOK, gin.H{"message": "metrics endpoint not implemented"})
}

// List models endpoint
func (h *IntegrationHandler) listModels(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "list_models")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	_ = c.Query("search") // TODO: implement search functionality

	models, err := h.database.ListModels(ctx, limit, page*limit)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list models",
			"details": err.Error(),
		})
		return
	}

	span.SetAttributes(
		attribute.Int("models.count", len(models)),
		attribute.Int("models.page", page),
		attribute.Int("models.limit", limit),
	)

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"count": len(models),
		},
	})
}

// Create model endpoint
func (h *IntegrationHandler) createModel(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_model")
	defer span.End()

	var req struct {
		Name         string                 `json:"name" binding:"required"`
		Version      string                 `json:"version" binding:"required"`
		Family       string                 `json:"family"`
		Format       string                 `json:"format" binding:"required"`
		Source       string                 `json:"source" binding:"required"`
		Config       map[string]interface{} `json:"config"`
		Metadata     map[string]interface{} `json:"metadata"`
		IsPublic     bool                   `json:"is_public"`
		Description  string                 `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// userID := middleware.GetUserID(c)
	userID := "default-user" // TODO: implement proper user ID extraction

	// Create model in database
	model, err := h.database.CreateModel(ctx, &database.Model{
		Name:      req.Name,
		Version:   req.Version,
		Family:    req.Family,
		Format:    req.Format,
		Config:    req.Config,
		Metadata:  req.Metadata,
		IsPublic:  req.IsPublic,
		CreatedBy: userID,
	})

	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create model",
			"details": err.Error(),
		})
		return
	}

	// Trigger model download if source is provided
	if req.Source != "" {
		// TODO: implement model download
		// go h.server.scheduler.DownloadModel(model.ID, req.Source)
	}

	span.SetAttributes(
		attribute.String("model.id", model.ID),
		attribute.String("model.name", model.Name),
		attribute.String("model.version", model.Version),
	)

	c.JSON(http.StatusCreated, model)
}

// Create completion endpoint
func (h *IntegrationHandler) createCompletion(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "create_completion")
	defer span.End()

	var req struct {
		Model       string                 `json:"model" binding:"required"`
		Prompt      string                 `json:"prompt" binding:"required"`
		MaxTokens   int                    `json:"max_tokens"`
		Temperature float64                `json:"temperature"`
		Stream      bool                   `json:"stream"`
		Parameters  map[string]interface{} `json:"parameters"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// userID := middleware.GetUserID(c)
	userID := "default-user" // TODO: implement proper user ID extraction
	
	// Create inference request
	inferenceReq, err := h.database.CreateInferenceRequest(ctx, &database.InferenceRequest{
		UserID:      userID,
		ModelID:     req.Model, // TODO: resolve model name to ID
		RequestType: "completion",
		Prompt:      req.Prompt,
		Parameters:  req.Parameters,
		Status:      "queued",
		Priority:    5,
	})

	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create inference request",
			"details": err.Error(),
		})
		return
	}

	// Submit to scheduler
	// TODO: implement scheduler integration
	// For now, just mark as submitted
	if err := error(nil); err != nil { // placeholder
		span.RecordError(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Failed to submit inference request",
			"details": err.Error(),
		})
		return
	}

	span.SetAttributes(
		attribute.String("inference.id", inferenceReq.ID),
		attribute.String("inference.model", req.Model),
		attribute.Bool("inference.stream", req.Stream),
	)

	if req.Stream {
		// Handle streaming response
		h.handleStreamingCompletion(c, inferenceReq.ID)
	} else {
		// Return job ID for polling
		c.JSON(http.StatusAccepted, gin.H{
			"id":     inferenceReq.ID,
			"status": "queued",
		})
	}
}

// Get cluster status endpoint
func (h *IntegrationHandler) getClusterStatus(c *gin.Context) {
	_, span := h.tracer.Start(c.Request.Context(), "get_cluster_status")
	defer span.End()

	status := gin.H{
		"timestamp": time.Now().UTC(),
		"cluster": gin.H{
			"healthy": true,
			"nodes":   0,
		},
		"consensus": gin.H{
			"enabled": h.server.consensus != nil,
		},
		"p2p": gin.H{
			"enabled": h.server.p2p != nil,
		},
	}

	// Get node information
	// TODO: implement p2p peer discovery
	if false { // h.server.p2p != nil {
		// peers := h.server.p2p.GetPeers()
		status["cluster"].(gin.H)["nodes"] = 1 // len(peers) + 1 // +1 for current node
		status["p2p"].(gin.H)["peer_count"] = 0 // len(peers)
	}

	// Get consensus information
	if h.server.consensus != nil {
		status["consensus"].(gin.H)["is_leader"] = h.server.consensus.IsLeader()
		status["consensus"].(gin.H)["term"] = h.server.consensus.GetCurrentTerm()
	}

	// Get scheduler information
	// TODO: implement scheduler status
	if false { // h.server.scheduler != nil {
		// queueLength := h.server.scheduler.GetQueueLength()
		status["scheduler"] = gin.H{
			"queue_length":   0, // queueLength,
			"loaded_models":  0, // h.server.scheduler.GetLoadedModelCount(),
			"active_jobs":    0, // h.server.scheduler.GetActiveJobCount(),
		}
	}

	c.JSON(http.StatusOK, status)
}

// Prometheus middleware
func (h *IntegrationHandler) prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		
		apiRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()
		
		apiRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// Tracing middleware
func (h *IntegrationHandler) tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		ctx, span := h.tracer.Start(ctx, spanName)
		defer span.End()
		
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)
		
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)
		
		if c.Writer.Status() >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// Handle streaming completion
func (h *IntegrationHandler) handleStreamingCompletion(c *gin.Context, requestID string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Set up SSE stream
	clientChan := make(chan string, 10)
	
	// Register for inference updates
	// TODO: implement scheduler callbacks
	// h.server.scheduler.RegisterInferenceCallback(requestID, func(update interface{}) {
	//	data, _ := json.Marshal(update)
	//	clientChan <- string(data)
	// })

	// defer h.server.scheduler.UnregisterInferenceCallback(requestID)
	
	for {
		select {
		case data := <-clientChan:
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
			
		case <-c.Request.Context().Done():
			return
		}
	}
}

// Additional handler stubs that would be fully implemented
func (h *IntegrationHandler) getModel(c *gin.Context) {
	// Implementation for getting a specific model
}

func (h *IntegrationHandler) updateModel(c *gin.Context) {
	// Implementation for updating a model
}

func (h *IntegrationHandler) deleteModel(c *gin.Context) {
	// Implementation for deleting a model
}

func (h *IntegrationHandler) downloadModel(c *gin.Context) {
	// Implementation for downloading/installing a model
}

func (h *IntegrationHandler) getModelStatus(c *gin.Context) {
	// Implementation for getting model download/installation status
}

func (h *IntegrationHandler) createChatCompletion(c *gin.Context) {
	// Implementation for chat completions
}

func (h *IntegrationHandler) createEmbedding(c *gin.Context) {
	// Implementation for creating embeddings
}

func (h *IntegrationHandler) getInferenceJob(c *gin.Context) {
	// Implementation for getting inference job status
}

func (h *IntegrationHandler) cancelInferenceJob(c *gin.Context) {
	// Implementation for canceling inference jobs
}

func (h *IntegrationHandler) getNodes(c *gin.Context) {
	// Implementation for listing cluster nodes
}

func (h *IntegrationHandler) getNode(c *gin.Context) {
	// Implementation for getting specific node details
}

func (h *IntegrationHandler) drainNode(c *gin.Context) {
	// Implementation for draining a node
}

func (h *IntegrationHandler) uncordonNode(c *gin.Context) {
	// Implementation for uncordoning a node
}

func (h *IntegrationHandler) listUsers(c *gin.Context) {
	// Implementation for listing users
}

func (h *IntegrationHandler) createUser(c *gin.Context) {
	// Implementation for creating users
}

func (h *IntegrationHandler) getUser(c *gin.Context) {
	// Implementation for getting user details
}

func (h *IntegrationHandler) updateUser(c *gin.Context) {
	// Implementation for updating users
}

func (h *IntegrationHandler) deleteUser(c *gin.Context) {
	// Implementation for deleting users
}

func (h *IntegrationHandler) getAuditLogs(c *gin.Context) {
	// Implementation for getting audit logs
}

func (h *IntegrationHandler) getSystemMetrics(c *gin.Context) {
	// Implementation for getting system metrics
}

func (h *IntegrationHandler) createWebhook(c *gin.Context) {
	// Implementation for creating webhooks
}

func (h *IntegrationHandler) listWebhooks(c *gin.Context) {
	// Implementation for listing webhooks
}

func (h *IntegrationHandler) deleteWebhook(c *gin.Context) {
	// Implementation for deleting webhooks
}

func (h *IntegrationHandler) proxyExternalAPI(c *gin.Context) {
	// Implementation for proxying external API calls
}

func (h *IntegrationHandler) inferenceWebSocket(c *gin.Context) {
	// Implementation for WebSocket inference streams
}

func (h *IntegrationHandler) clusterEventsWebSocket(c *gin.Context) {
	// Implementation for WebSocket cluster events
}

func (h *IntegrationHandler) logsWebSocket(c *gin.Context) {
	// Implementation for WebSocket log streaming
}

func (h *IntegrationHandler) graphqlHandler(c *gin.Context) {
	// Implementation for GraphQL endpoint
}

func (h *IntegrationHandler) grpcGatewayHandler(c *gin.Context) {
	// Implementation for gRPC gateway
}