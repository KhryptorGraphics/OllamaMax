package auth

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama-distributed/internal/config"
)

// ExampleServerWithAuth demonstrates how to integrate the authentication system
// with the existing Ollama distributed server
func ExampleServerWithAuth() {
	// Load main configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Ensure auth is enabled
	if !cfg.Security.Auth.Enabled {
		log.Println("WARNING: Authentication is disabled. Enable it for production!")
		cfg.Security.Auth.Enabled = true
		cfg.Security.Auth.Method = "jwt"
		cfg.Security.Auth.TokenExpiry = 24 * time.Hour
		cfg.Security.Auth.SecretKey = "demo-secret-key-change-in-production"
		cfg.Security.Auth.Issuer = "ollama-distributed"
		cfg.Security.Auth.Audience = "ollama-api"
	}
	
	// Create authentication integration
	authIntegration, err := NewIntegration(&cfg.Security.Auth)
	if err != nil {
		log.Fatalf("Failed to create auth integration: %v", err)
	}
	defer authIntegration.Close()
	
	// Create Gin router
	router := gin.New()
	
	// Apply global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(authIntegration.MiddlewareManager.SecurityHeaders())
	router.Use(authIntegration.MiddlewareManager.CORS())
	router.Use(authIntegration.MiddlewareManager.RateLimit())
	router.Use(authIntegration.MiddlewareManager.AuditLog())
	
	// Register authentication routes
	authIntegration.Routes.RegisterRoutes(router)
	
	// Setup protected API routes
	setupProtectedAPIRoutes(router, authIntegration)
	
	// Setup public routes
	setupPublicRoutes(router)
	
	// Start server
	log.Printf("Starting Ollama Distributed Server with Authentication on %s", cfg.API.Listen)
	if err := router.Run(cfg.API.Listen); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupProtectedAPIRoutes configures the protected API endpoints
func setupProtectedAPIRoutes(router *gin.Engine, authIntegration *Integration) {
	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(authIntegration.MiddlewareManager.AuthRequired())
	
	// Create middleware helpers
	helpers := NewMiddlewareHelpers(authIntegration)
	
	// Node management endpoints
	setupNodeRoutes(api, helpers)
	
	// Model management endpoints
	setupModelRoutes(api, helpers)
	
	// Cluster management endpoints
	setupClusterRoutes(api, helpers)
	
	// Inference endpoints
	setupInferenceRoutes(api, helpers)
	
	// Monitoring endpoints
	setupMonitoringRoutes(api, helpers)
	
	// Distribution management endpoints
	setupDistributionRoutes(api, helpers)
}

// setupNodeRoutes configures node management routes
func setupNodeRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	nodes := api.Group("/nodes")
	
	// List nodes - requires read permission
	nodes.GET("", helpers.RequireNodePermission("read"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s requested node list", user.Username)
		
		// Mock response - in real implementation, this would call the scheduler
		c.JSON(200, gin.H{
			"nodes": []map[string]interface{}{
				{
					"id":     "node-1",
					"status": "online",
					"cpu":    "50%",
					"memory": "60%",
				},
				{
					"id":     "node-2",
					"status": "online",
					"cpu":    "30%",
					"memory": "40%",
				},
			},
			"total": 2,
		})
	})
	
	// Get specific node - requires read permission
	nodes.GET("/:id", helpers.RequireNodePermission("read"), func(c *gin.Context) {
		nodeID := c.Param("id")
		user := GetCurrentUser(c)
		log.Printf("User %s requested details for node %s", user.Username, nodeID)
		
		c.JSON(200, gin.H{
			"node": map[string]interface{}{
				"id":       nodeID,
				"status":   "online",
				"cpu":      "50%",
				"memory":   "60%",
				"models":   []string{"llama2", "codellama"},
				"requests": 150,
			},
		})
	})
	
	// Drain node - requires write permission
	nodes.POST("/:id/drain", helpers.RequireNodePermission("write"), func(c *gin.Context) {
		nodeID := c.Param("id")
		user := GetCurrentUser(c)
		log.Printf("User %s initiated drain for node %s", user.Username, nodeID)
		
		c.JSON(200, gin.H{
			"message": "Node drain initiated",
			"node_id": nodeID,
			"status":  "draining",
		})
	})
	
	// Delete node - requires admin permission
	nodes.DELETE("/:id", helpers.RequireNodePermission("admin"), func(c *gin.Context) {
		nodeID := c.Param("id")
		user := GetCurrentUser(c)
		log.Printf("User %s deleted node %s", user.Username, nodeID)
		
		c.JSON(200, gin.H{
			"message": "Node deleted successfully",
			"node_id": nodeID,
		})
	})
}

// setupModelRoutes configures model management routes
func setupModelRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	models := api.Group("/models")
	
	// List models - requires read permission
	models.GET("", helpers.RequireModelPermission("read"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s requested model list", user.Username)
		
		c.JSON(200, gin.H{
			"models": []map[string]interface{}{
				{
					"name":      "llama2",
					"size":      "7B",
					"locations": []string{"node-1", "node-2"},
					"status":    "ready",
				},
				{
					"name":      "codellama",
					"size":      "13B",
					"locations": []string{"node-1"},
					"status":    "ready",
				},
			},
		})
	})
	
	// Download model - requires write permission
	models.POST("/:name/download", helpers.RequireModelPermission("write"), func(c *gin.Context) {
		modelName := c.Param("name")
		user := GetCurrentUser(c)
		log.Printf("User %s initiated download for model %s", user.Username, modelName)
		
		c.JSON(200, gin.H{
			"message":    "Model download initiated",
			"model_name": modelName,
			"status":     "downloading",
			"progress":   0,
		})
	})
	
	// Delete model - requires admin permission
	models.DELETE("/:name", helpers.RequireModelPermission("admin"), func(c *gin.Context) {
		modelName := c.Param("name")
		user := GetCurrentUser(c)
		log.Printf("User %s deleted model %s", user.Username, modelName)
		
		c.JSON(200, gin.H{
			"message":    "Model deleted successfully",
			"model_name": modelName,
		})
	})
}

// setupClusterRoutes configures cluster management routes
func setupClusterRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	cluster := api.Group("/cluster")
	
	// Get cluster status - requires read permission
	cluster.GET("/status", helpers.RequireClusterPermission("read"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s requested cluster status", user.Username)
		
		c.JSON(200, gin.H{
			"status": "healthy",
			"nodes":  2,
			"leader": "node-1",
			"peers":  1,
		})
	})
	
	// Join cluster - requires write permission
	cluster.POST("/join", helpers.RequireClusterPermission("write"), func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		
		user := GetCurrentUser(c)
		log.Printf("User %s initiated cluster join", user.Username)
		
		c.JSON(200, gin.H{
			"message": "Node join initiated",
		})
	})
	
	// Leave cluster - requires admin permission
	cluster.POST("/leave", helpers.RequireClusterPermission("admin"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s initiated cluster leave", user.Username)
		
		c.JSON(200, gin.H{
			"message": "Node leave initiated",
		})
	})
}

// setupInferenceRoutes configures inference routes
func setupInferenceRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	// Generate endpoint - requires write permission
	api.POST("/generate", helpers.RequireInferencePermission("write"), func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		
		user := GetCurrentUser(c)
		modelName := req["model"]
		log.Printf("User %s requested generation with model %v", user.Username, modelName)
		
		c.JSON(200, gin.H{
			"response": "This is a generated response from the distributed Ollama system",
			"model":    modelName,
			"node_id":  "node-1",
			"user":     user.Username,
		})
	})
	
	// Chat endpoint - requires write permission
	api.POST("/chat", helpers.RequireInferencePermission("write"), func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		
		user := GetCurrentUser(c)
		log.Printf("User %s initiated chat session", user.Username)
		
		c.JSON(200, gin.H{
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": "Hello! I'm your AI assistant powered by the distributed Ollama system.",
			},
			"user": user.Username,
		})
	})
	
	// Embeddings endpoint - requires write permission
	api.POST("/embeddings", helpers.RequireInferencePermission("write"), func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		
		user := GetCurrentUser(c)
		log.Printf("User %s requested embeddings", user.Username)
		
		c.JSON(200, gin.H{
			"embeddings": []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			"model":      req["model"],
			"user":       user.Username,
		})
	})
}

// setupMonitoringRoutes configures monitoring routes
func setupMonitoringRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	// Metrics endpoint - requires read permission
	api.GET("/metrics", helpers.RequireInferencePermission("read"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s requested metrics", user.Username)
		
		c.JSON(200, gin.H{
			"metrics": map[string]interface{}{
				"nodes_online":       2,
				"models_loaded":      5,
				"requests_processed": 1500,
				"cpu_usage":          45.2,
				"memory_usage":       62.8,
				"network_usage":      23.1,
			},
			"timestamp": time.Now().Unix(),
		})
	})
	
	// Health check - no authentication required for monitoring
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})
	
	// Transfers endpoint
	api.GET("/transfers", helpers.RequireInferencePermission("read"), func(c *gin.Context) {
		user := GetCurrentUser(c)
		log.Printf("User %s requested transfer status", user.Username)
		
		c.JSON(200, gin.H{
			"transfers": []map[string]interface{}{
				{
					"id":       "transfer-1",
					"model":    "llama2",
					"status":   "completed",
					"progress": 100,
				},
			},
		})
	})
}

// setupDistributionRoutes configures distribution management routes
func setupDistributionRoutes(api *gin.RouterGroup, helpers *MiddlewareHelpers) {
	distribution := api.Group("/distribution")
	
	// Auto-configure distribution - requires admin permission
	distribution.POST("/auto-configure", helpers.RequireClusterPermission("admin"), func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		
		user := GetCurrentUser(c)
		log.Printf("User %s configured auto-distribution", user.Username)
		
		c.JSON(200, gin.H{
			"message": "Auto-distribution configured",
			"enabled": req["enabled"],
		})
	})
}

// setupPublicRoutes configures public routes that don't require authentication
func setupPublicRoutes(router *gin.Engine) {
	// Serve static files for web UI
	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/favicon.ico", "./web/favicon.ico")
	
	// Catch-all for SPA routing
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})
}

// DemoUsage shows how to use the authentication system programmatically
func DemoUsage() {
	// Create auth config
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * time.Hour,
		SecretKey:   "demo-secret-key",
		Issuer:      "ollama-distributed",
		Audience:    "ollama-api",
	}
	
	// Create auth manager
	authManager, err := NewManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create auth manager: %v", err)
	}
	defer authManager.Close()
	
	// Create a new user
	userReq := &CreateUserRequest{
		Username: "demo-user",
		Email:    "demo@example.com",
		Password: "secure-password",
		Role:     RoleUser,
	}
	
	user, err := authManager.CreateUser(userReq)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	
	log.Printf("Created user: %s (ID: %s)", user.Username, user.ID)
	
	// Authenticate user
	authCtx, err := authManager.Authenticate("demo-user", "secure-password", map[string]string{
		"ip_address": "127.0.0.1",
		"user_agent": "demo-client",
	})
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	
	log.Printf("Authentication successful! Token: %s", authCtx.TokenString[:50]+"...")
	
	// Create API key
	apiKeyReq := &CreateAPIKeyRequest{
		Name:        "Demo API Key",
		Permissions: []string{PermissionModelRead, PermissionInferenceWrite},
	}
	
	apiKey, rawKey, err := authManager.CreateAPIKey(user.ID, apiKeyReq)
	if err != nil {
		log.Fatalf("Failed to create API key: %v", err)
	}
	
	log.Printf("Created API key: %s (Key: %s)", apiKey.Name, rawKey[:20]+"...")
	
	// Validate API key
	apiAuthCtx, err := authManager.ValidateAPIKey(rawKey)
	if err != nil {
		log.Fatalf("Failed to validate API key: %v", err)
	}
	
	log.Printf("API key validation successful for user: %s", apiAuthCtx.User.Username)
	
	// Check permissions
	hasModelRead := authManager.HasPermission(apiAuthCtx, PermissionModelRead)
	hasSystemAdmin := authManager.HasPermission(apiAuthCtx, PermissionSystemAdmin)
	
	log.Printf("User has model read permission: %v", hasModelRead)
	log.Printf("User has system admin permission: %v", hasSystemAdmin)
}