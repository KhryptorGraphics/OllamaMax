package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
)

//go:embed static/*
var staticFiles embed.FS

// WebServer provides the web UI for OllamaMax
type WebServer struct {
	config     *Config
	apiServer  *api.Server
	router     *gin.Engine
	server     *http.Server
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// Config holds web server configuration
type Config struct {
	ListenAddress string `yaml:"listen_address" json:"listen_address"`
	EnableTLS     bool   `yaml:"enable_tls" json:"enable_tls"`
	TLSCertFile   string `yaml:"tls_cert_file" json:"tls_cert_file"`
	TLSKeyFile    string `yaml:"tls_key_file" json:"tls_key_file"`
	StaticPath    string `yaml:"static_path" json:"static_path"`
	EnableAuth    bool   `yaml:"enable_auth" json:"enable_auth"`
}

// DefaultConfig returns default web server configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddress: ":8080",
		EnableTLS:     false,
		StaticPath:    "./web",
		EnableAuth:    true,
	}
}

// NewWebServer creates a new web server instance
func NewWebServer(config *Config, apiServer *api.Server) *WebServer {
	if config == nil {
		config = DefaultConfig()
	}

	// Configure WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow connections from same origin
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws := &WebServer{
		config:     config,
		apiServer:  apiServer,
		upgrader:   upgrader,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}

	ws.setupRouter()
	return ws
}

// setupRouter configures the web server routes
func (ws *WebServer) setupRouter() {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)
	
	ws.router = gin.New()
	
	// Add middleware
	ws.router.Use(gin.Logger())
	ws.router.Use(gin.Recovery())
	ws.router.Use(ws.corsMiddleware())
	
	// Add security headers
	ws.router.Use(ws.securityHeadersMiddleware())
	
	// Add metrics middleware
	ws.router.Use(observability.GinMetricsMiddleware())

	// WebSocket endpoint
	ws.router.GET("/ws", ws.handleWebSocket)

	// API proxy endpoints
	api := ws.router.Group("/api")
	{
		api.GET("/v1/proxy/status", ws.proxyToAPI)
		api.GET("/v1/proxy/instances", ws.proxyToAPI)
		api.GET("/v1/proxy/metrics", ws.proxyToAPI)
		api.GET("/v1/nodes", ws.proxyToAPI)
		api.GET("/v1/models", ws.proxyToAPI)
		api.POST("/v1/models/pull", ws.proxyToAPI)
		api.DELETE("/v1/models/:name", ws.proxyToAPI)
		api.GET("/v1/cluster/status", ws.proxyToAPI)
		api.GET("/v1/analytics/performance", ws.proxyToAPI)
		api.GET("/v1/security/status", ws.proxyToAPI)
	}

	// Health check endpoint
	ws.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "ollama-distributed-web",
		})
	})

	// Serve static files
	ws.setupStaticFiles()
}

// setupStaticFiles configures static file serving
func (ws *WebServer) setupStaticFiles() {
	// Try to serve from embedded files first
	staticFS, err := fs.Sub(staticFiles, "static")
	if err == nil {
		ws.router.StaticFS("/static", http.FS(staticFS))
	}

	// Serve main web application
	ws.router.GET("/", ws.serveIndex)
	ws.router.GET("/index.html", ws.serveIndex)
	
	// Serve web application for all other routes (SPA routing)
	ws.router.NoRoute(ws.serveIndex)
}

// serveIndex serves the main web application
func (ws *WebServer) serveIndex(c *gin.Context) {
	// Try to serve from custom static path first
	if ws.config.StaticPath != "" {
		indexPath := filepath.Join(ws.config.StaticPath, "index.html")
		c.File(indexPath)
		return
	}

	// Serve from embedded files
	indexContent, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		c.String(http.StatusNotFound, "Web UI not found")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
}

// corsMiddleware adds CORS headers
func (ws *WebServer) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// securityHeadersMiddleware adds security headers
func (ws *WebServer) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		c.Next()
	}
}

// proxyToAPI proxies requests to the API server
func (ws *WebServer) proxyToAPI(c *gin.Context) {
	// Extract the API path
	apiPath := strings.TrimPrefix(c.Request.URL.Path, "/api")
	
	// Forward the request to the API server
	// This is a simplified proxy - in production, you might want to use a proper reverse proxy
	c.JSON(http.StatusOK, gin.H{
		"message": "API proxy not fully implemented",
		"path":    apiPath,
		"method":  c.Request.Method,
	})
}

// handleWebSocket handles WebSocket connections
func (ws *WebServer) handleWebSocket(c *gin.Context) {
	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// Register client
	ws.register <- conn

	// Handle messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Echo message back (simple implementation)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Printf("WebSocket write error: %v\n", err)
			break
		}
	}

	// Unregister client
	ws.unregister <- conn
}

// handleWebSocketHub manages WebSocket connections
func (ws *WebServer) handleWebSocketHub() {
	for {
		select {
		case client := <-ws.register:
			ws.clients[client] = true
			fmt.Printf("WebSocket client connected. Total: %d\n", len(ws.clients))

		case client := <-ws.unregister:
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				client.Close()
				fmt.Printf("WebSocket client disconnected. Total: %d\n", len(ws.clients))
			}

		case message := <-ws.broadcast:
			for client := range ws.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					delete(ws.clients, client)
					client.Close()
				}
			}
		}
	}
}

// Start starts the web server
func (ws *WebServer) Start() error {
	// Start WebSocket hub
	go ws.handleWebSocketHub()

	// Create HTTP server
	ws.server = &http.Server{
		Addr:         ws.config.ListenAddress,
		Handler:      ws.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("🌐 Starting web server on %s\n", ws.config.ListenAddress)

	// Start server
	if ws.config.EnableTLS {
		return ws.server.ListenAndServeTLS(ws.config.TLSCertFile, ws.config.TLSKeyFile)
	}
	return ws.server.ListenAndServe()
}

// Stop stops the web server
func (ws *WebServer) Stop() error {
	if ws.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return ws.server.Shutdown(ctx)
}

// BroadcastMessage sends a message to all connected WebSocket clients
func (ws *WebServer) BroadcastMessage(message []byte) {
	select {
	case ws.broadcast <- message:
	default:
		// Channel is full, skip this message
	}
}
