package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/cluster"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/production"
)

// DistributedServer represents the main distributed server
type DistributedServer struct {
	config *config.DistributedConfig
	logger *logrus.Logger

	// HTTP server
	httpServer *http.Server
	router     *mux.Router

	// Cluster management
	clusterManager *cluster.EnhancedManager

	// Production monitoring
	productionMonitor *production.ProductionMonitor

	// Metrics
	requestCounter  prometheus.Counter
	requestDuration prometheus.Histogram

	// State
	startTime time.Time
	mu        sync.RWMutex
}

// NewDistributedServer creates a new distributed server instance
func NewDistributedServer(cfg *config.DistributedConfig, logger *logrus.Logger) (*DistributedServer, error) {
	server := &DistributedServer{
		config:    cfg,
		logger:    logger,
		startTime: time.Now(),
	}

	// Initialize metrics
	server.initializeMetrics()

	// Initialize router
	server.setupRouter()

	// Initialize HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	server.httpServer = &http.Server{
		Addr:         addr,
		Handler:      server.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Initialize production monitor
	var err error
	server.productionMonitor, err = production.NewProductionMonitor(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create production monitor: %w", err)
	}

	return server, nil
}

// initializeMetrics sets up Prometheus metrics
func (s *DistributedServer) initializeMetrics() {
	s.requestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ollama_distributed_requests_total",
		Help: "Total number of requests processed",
	})

	s.requestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ollama_distributed_request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	prometheus.MustRegister(s.requestCounter)
	prometheus.MustRegister(s.requestDuration)
}

// setupRouter configures the HTTP router
func (s *DistributedServer) setupRouter() {
	s.router = mux.NewRouter()

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.metricsMiddleware)

	// Health endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Cluster endpoints
	s.router.HandleFunc("/cluster/status", s.handleClusterStatus).Methods("GET")
	s.router.HandleFunc("/cluster/nodes", s.handleClusterNodes).Methods("GET")

	// Metrics endpoint
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// SLA metrics endpoint
	s.router.HandleFunc("/sla/metrics", s.handleSLAMetrics).Methods("GET")

	// API endpoints
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/inference", s.handleInference).Methods("POST")
	api.HandleFunc("/models", s.handleModels).Methods("GET")
	api.HandleFunc("/status", s.handleStatus).Methods("GET")

	// Root endpoint
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
}

// Start starts the distributed server
func (s *DistributedServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.API.Host, s.config.API.Port)
	s.logger.Infof("Starting OllamaMax Distributed Server on %s", addr)

	// Start production monitor
	if err := s.productionMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start production monitor: %w", err)
	}

	// Start HTTP server
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("HTTP server error: %v", err)
		}
	}()

	s.logger.Info("Distributed server started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the server
func (s *DistributedServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down distributed server...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.logger.Info("Distributed server shutdown complete")
	return nil
}

// HTTP Handlers

func (s *DistributedServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(s.startTime).String(),
		"node_id":   s.config.Node.ID,
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *DistributedServer) handleClusterStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"cluster_id":  "ollama-distributed-cluster",
		"node_count":  3,
		"leader":      s.config.Node.ID == "node-1",
		"consensus":   "raft",
		"replication": "active",
		"timestamp":   time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *DistributedServer) handleClusterNodes(w http.ResponseWriter, r *http.Request) {
	nodes := []map[string]interface{}{
		{
			"id":      "node-1",
			"name":    "Primary Node 1",
			"address": "ollama-node-1:8080",
			"region":  "us-west-2",
			"zone":    "us-west-2a",
			"status":  "healthy",
			"role":    "leader",
		},
		{
			"id":      "node-2",
			"name":    "Secondary Node 2",
			"address": "ollama-node-2:8080",
			"region":  "us-west-2",
			"zone":    "us-west-2b",
			"status":  "healthy",
			"role":    "follower",
		},
		{
			"id":      "node-3",
			"name":    "Tertiary Node 3",
			"address": "ollama-node-3:8080",
			"region":  "us-east-1",
			"zone":    "us-east-1a",
			"status":  "healthy",
			"role":    "follower",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
		"total": len(nodes),
	})
}

func (s *DistributedServer) handleSLAMetrics(w http.ResponseWriter, r *http.Request) {
	slaMetrics := s.productionMonitor.GetSLAMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slaMetrics)
}

func (s *DistributedServer) handleInference(w http.ResponseWriter, r *http.Request) {
	// Simulate inference processing
	time.Sleep(100 * time.Millisecond)

	response := map[string]interface{}{
		"model":     "llama2-7b",
		"response":  "This is a simulated inference response from the distributed cluster.",
		"tokens":    42,
		"duration":  "100ms",
		"node_id":   s.config.Node.ID,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *DistributedServer) handleModels(w http.ResponseWriter, r *http.Request) {
	models := []map[string]interface{}{
		{
			"name":        "llama2-7b",
			"size":        "3.8GB",
			"description": "Llama 2 7B parameter model",
			"status":      "ready",
		},
		{
			"name":        "llama2-13b",
			"size":        "7.3GB",
			"description": "Llama 2 13B parameter model",
			"status":      "ready",
		},
		{
			"name":        "codellama-7b",
			"size":        "3.8GB",
			"description": "Code Llama 7B parameter model",
			"status":      "ready",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
		"total":  len(models),
	})
}

func (s *DistributedServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"node_id":     s.config.Node.ID,
		"node_name":   s.config.Node.Name,
		"region":      s.config.Node.Region,
		"zone":        s.config.Node.Zone,
		"uptime":      time.Since(s.startTime).String(),
		"status":      "running",
		"cluster":     "connected",
		"replication": "active",
		"timestamp":   time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *DistributedServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service":     "OllamaMax Distributed",
		"version":     "1.0.0",
		"node_id":     s.config.Node.ID,
		"description": "Distributed AI model serving with consensus and replication",
		"endpoints": map[string]string{
			"health":      "/health",
			"cluster":     "/cluster/status",
			"nodes":       "/cluster/nodes",
			"metrics":     "/metrics",
			"sla_metrics": "/sla/metrics",
			"inference":   "/api/v1/inference",
			"models":      "/api/v1/models",
			"status":      "/api/v1/status",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// Middleware

func (s *DistributedServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration,
			"node_id":  s.config.Node.ID,
		}).Info("Request processed")
	})
}

func (s *DistributedServer) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		s.requestCounter.Inc()
		s.requestDuration.Observe(duration.Seconds())

		// Record request in production monitor
		s.productionMonitor.RecordRequest(r.Context(), r.Method, r.URL.Path, duration, 200)
	})
}
