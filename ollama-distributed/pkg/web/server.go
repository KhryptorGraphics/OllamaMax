package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
)

//go:embed static/*
var staticFiles embed.FS

// WebServer provides the web UI for OllamaMax
type WebServer struct {
	config     *Config
	apiServer  *api.Server
	apiBaseURL string
	router     *gin.Engine
	server     *http.Server
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	httpClient *http.Client
}

// Config holds web server configuration
type Config struct {
	ListenAddress string `yaml:"listen_address" json:"listen_address"`
	EnableTLS     bool   `yaml:"enable_tls" json:"enable_tls"`
	TLSCertFile   string `yaml:"tls_cert_file" json:"tls_cert_file"`
	TLSKeyFile    string `yaml:"tls_key_file" json:"tls_key_file"`
	StaticPath    string `yaml:"static_path" json:"static_path"`
	EnableAuth    bool   `yaml:"enable_auth" json:"enable_auth"`
	APIBaseURL    string `yaml:"api_base_url" json:"api_base_url"`
}

// DefaultConfig returns default web server configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddress: ":8081",
		EnableTLS:     false,
		StaticPath:    "./web",
		EnableAuth:    true,
		APIBaseURL:    "http://localhost:8080",
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
		apiBaseURL: config.APIBaseURL,
		upgrader:   upgrader,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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
	// ws.router.Use(observability.GinMetricsMiddleware()) // Temporarily disabled

	// WebSocket endpoint
	ws.router.GET("/ws", ws.handleWebSocket)

	// API proxy endpoints
	api := ws.router.Group("/api")
	{
		// Debug route to test API group
		api.GET("/debug", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "API group is working",
				"path":    c.Request.URL.Path,
			})
		})

		// Core API endpoints
		api.GET("v1/health", ws.proxyToAPI)
		api.GET("v1/version", ws.proxyToAPI)
		api.GET("v1/nodes", ws.proxyToAPI)
		api.GET("v1/models", ws.proxyToAPI)
		api.POST("v1/models/pull", ws.proxyToAPI)
		api.DELETE("v1/models/:name", ws.proxyToAPI)
		api.GET("v1/cluster/status", ws.proxyToAPI)
		api.GET("v1/cluster/leader", ws.proxyToAPI)
		api.GET("v1/tasks", ws.proxyToAPI)
		api.GET("v1/tasks/queue", ws.proxyToAPI)
		api.POST("v1/inference", ws.proxyToAPI)

		// Metrics endpoints
		api.GET("v1/metrics", ws.proxyToAPI)
		api.GET("v1/metrics/resources", ws.proxyToAPI)
		api.GET("v1/metrics/performance", ws.proxyToAPI)

		// Security endpoints
		api.GET("v1/security/status", ws.proxyToAPI)
		api.GET("v1/security/threats", ws.proxyToAPI)
		api.GET("v1/security/alerts", ws.proxyToAPI)
		api.GET("v1/security/audit", ws.proxyToAPI)

		// Performance endpoints
		api.GET("v1/performance/metrics", ws.proxyToAPI)
		api.GET("v1/performance/optimizations", ws.proxyToAPI)
		api.GET("v1/performance/bottlenecks", ws.proxyToAPI)
		api.GET("v1/performance/report", ws.proxyToAPI)

		// Model sync endpoints
		api.GET("v1/models/sync/status", ws.proxyToAPI)
		api.POST("v1/models/sync/start", ws.proxyToAPI)

		// Transfer endpoints
		api.GET("v1/transfers", ws.proxyToAPI)
		api.POST("v1/transfers", ws.proxyToAPI)
		api.DELETE("v1/transfers/:id", ws.proxyToAPI)
	}

	// Health check endpoint
	ws.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "ollama-distributed-web",
		})
	})

	// Diagnostic endpoint for troubleshooting
	ws.router.GET("/diagnostic", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"web_server": gin.H{
				"status":      "running",
				"listen_addr": ws.config.ListenAddress,
				"api_base":    ws.config.APIBaseURL,
			},
			"static_files": gin.H{
				"embed_available": true,
				"static_path":     ws.config.StaticPath,
			},
			"routes": gin.H{
				"health":     "/health",
				"diagnostic": "/diagnostic",
				"main":       "/",
				"api_proxy":  "/api/v1/*",
			},
			"troubleshooting": gin.H{
				"common_issues": []string{
					"CDN resources not loading (React, Bootstrap, etc.)",
					"JavaScript errors in browser console",
					"API proxy not working",
					"CORS issues",
					"Missing static files",
				},
				"next_steps": []string{
					"Check browser console for errors (F12)",
					"Test API connectivity: curl http://localhost:8081/api/v1/health",
					"Verify CDN access: check network tab in browser",
					"Test basic HTML: curl http://localhost:8081/",
				},
			},
		})
	})

	// Serve static files
	ws.setupStaticFiles()

	// Optional: redirect legacy /login to new /v2/auth/login when enabled
	if os.Getenv("LOGIN_REDIRECT") == "1" {
		ws.router.GET("/login", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/v2/auth/login")
		})
	}

	// Dev proxy or prod static for Vite-based app under /v2
	ws.setupV2Routes()
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

	// Serve test page for diagnostics
	ws.router.GET("/test", ws.serveTest)
	ws.router.GET("/test.html", ws.serveTest)

	// Serve web application for all other routes (SPA routing)
	// Only serve index for non-API routes
	ws.router.NoRoute(func(c *gin.Context) {
		// Don't serve index for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
				"path":  c.Request.URL.Path,
			})
			return
		}
		// Serve index for all other routes (SPA routing)
		ws.serveIndex(c)
	})
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

// serveTest serves the diagnostic test page
func (ws *WebServer) serveTest(c *gin.Context) {
	// Try to serve from custom static path first
	if ws.config.StaticPath != "" {
		testPath := filepath.Join(ws.config.StaticPath, "test.html")
		c.File(testPath)
		return
	}

	// Serve from embedded files
	testContent, err := staticFiles.ReadFile("static/test.html")
	if err != nil {
		c.String(http.StatusNotFound, "Test page not found")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", testContent)
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
	// Keep the full path including /api prefix for the API server
	apiPath := c.Request.URL.Path

	// Build target URL
	targetURL, err := url.Parse(ws.apiBaseURL + apiPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid API URL",
		})
		return
	}

	// Add query parameters
	if c.Request.URL.RawQuery != "" {
		targetURL.RawQuery = c.Request.URL.RawQuery
	}

	// Create request
	var reqBody io.Reader
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read request body",
			})
			return
		}
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(c.Request.Method, targetURL.String(), reqBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create request",
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make request
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Failed to reach API server",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response",
		})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
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

	// Send initial data
	ws.sendInitialData(conn)

	// Handle messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Handle incoming messages
		ws.handleWebSocketMessage(conn, message)
	}

	// Unregister client
	ws.unregister <- conn
}

// sendInitialData sends initial data to a new WebSocket client
func (ws *WebServer) sendInitialData(conn *websocket.Conn) {
	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type":      "welcome",
		"message":   "Connected to OllamaMax Distributed",
		"timestamp": time.Now().UTC(),
	}

	if data, err := json.Marshal(welcomeMsg); err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

// handleWebSocketMessage handles incoming WebSocket messages
func (ws *WebServer) handleWebSocketMessage(conn *websocket.Conn, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Respond to ping with pong
		pongMsg := map[string]interface{}{
			"type":      "pong",
			"timestamp": msg["timestamp"],
		}
		if data, err := json.Marshal(pongMsg); err == nil {
			conn.WriteMessage(websocket.TextMessage, data)
		}
	case "subscribe":
		// Handle subscription requests
		ws.handleSubscription(conn, msg)
	case "unsubscribe":
		// Handle unsubscription requests
		ws.handleUnsubscription(conn, msg)
	}
}

// handleSubscription handles subscription requests
func (ws *WebServer) handleSubscription(conn *websocket.Conn, msg map[string]interface{}) {
	// Implementation for handling subscriptions to specific data streams
	// This could include metrics, logs, node status, etc.
}

// handleUnsubscription handles unsubscription requests
func (ws *WebServer) handleUnsubscription(conn *websocket.Conn, msg map[string]interface{}) {
	// Implementation for handling unsubscriptions
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

// setupV2Routes configures /v2 to point to Vite (dev or prod)
func (ws *WebServer) setupV2Routes() {
	// Environment-driven behavior
	viteDev := os.Getenv("VITE_DEV_PROXY") == "1"
	viteDevURL := os.Getenv("VITE_DEV_URL")
	if viteDevURL == "" {
		viteDevURL = "http://localhost:5173"
	}

	if viteDev {
		// Proxy /v2/* to Vite dev server
		ws.router.Any("/v2/*path", func(c *gin.Context) {
			target, _ := url.Parse(viteDevURL)
			// Rebuild target URL
			target.Path = c.Request.URL.Path
			target.RawQuery = c.Request.URL.RawQuery

			req, err := http.NewRequest(c.Request.Method, target.String(), c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "proxy request create failed"})
				return
			}
			for k, vs := range c.Request.Header {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
			resp, err := ws.httpClient.Do(req)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "vite dev unreachable", "details": err.Error()})
				return
			}
			defer resp.Body.Close()
			for k, vs := range resp.Header {
				for _, v := range vs {
					c.Header(k, v)
				}
			}
			body, _ := io.ReadAll(resp.Body)
			c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
		})
		return
	}

	// Production: serve built Vite app (web/frontend/dist) under /v2
	viteDist := filepath.Join("./web/frontend", "dist")
	// Static assets
	ws.router.Static("/v2/assets", filepath.Join(viteDist, "assets"))
	// Index for /v2 and client-side routes
	ws.router.GET("/v2", func(c *gin.Context) {
		c.File(filepath.Join(viteDist, "index.html"))
	})
	ws.router.GET("/v2/*path", func(c *gin.Context) {
		// Do not shadow API
		if strings.HasPrefix(c.Param("path"), "/api/") {
			c.Next()
			return
		}
		c.File(filepath.Join(viteDist, "index.html"))
	})
}

		// Channel is full, skip this message
	}
}
