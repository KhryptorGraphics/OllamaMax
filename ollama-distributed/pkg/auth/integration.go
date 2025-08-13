package auth

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

// Integration provides easy integration with the existing API server
type Integration struct {
	AuthManager       *Manager
	JWTManager        *JWTManager
	MiddlewareManager *MiddlewareManager
	Routes            *Routes
}

// NewIntegration creates a complete authentication integration
func NewIntegration(cfg *config.AuthConfig) (*Integration, error) {
	// Create auth manager
	authManager, err := NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	// Create JWT manager
	jwtManager, err := NewJWTManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	// Create middleware manager
	middlewareManager := NewMiddlewareManager(authManager, jwtManager, cfg)

	// Create routes
	routes := NewRoutes(authManager, jwtManager, middlewareManager)

	return &Integration{
		AuthManager:       authManager,
		JWTManager:        jwtManager,
		MiddlewareManager: middlewareManager,
		Routes:            routes,
	}, nil
}

// SetupRouter configures a Gin router with authentication
func (i *Integration) SetupRouter() *gin.Engine {
	router := gin.New()

	// Register authentication routes
	i.Routes.RegisterRoutes(router)

	return router
}

// ProtectAPIRoutes adds authentication to existing API routes
func (i *Integration) ProtectAPIRoutes(router *gin.Engine) {
	// Apply authentication middleware to protected API routes
	api := router.Group("/api/v1")
	api.Use(i.MiddlewareManager.AuthRequired())

	// Node management - requires node permissions
	nodeRoutes := api.Group("/nodes")
	nodeRoutes.Use(i.MiddlewareManager.RequireAnyPermission(
		PermissionNodeRead,
		PermissionNodeWrite,
		PermissionNodeAdmin,
	))

	// Model management - requires model permissions
	modelRoutes := api.Group("/models")
	modelRoutes.Use(i.MiddlewareManager.RequireAnyPermission(
		PermissionModelRead,
		PermissionModelWrite,
		PermissionModelAdmin,
	))

	// Cluster management - requires cluster permissions
	clusterRoutes := api.Group("/cluster")
	clusterRoutes.Use(i.MiddlewareManager.RequireAnyPermission(
		PermissionClusterRead,
		PermissionClusterWrite,
		PermissionClusterAdmin,
	))

	// Inference - requires inference permissions
	inferenceRoutes := api.Group("/")
	inferenceRoutes.Use(i.MiddlewareManager.RequireAnyPermission(
		PermissionInferenceRead,
		PermissionInferenceWrite,
	))

	// Metrics - requires metrics permissions
	metricsRoutes := api.Group("/metrics")
	metricsRoutes.Use(i.MiddlewareManager.RequirePermission(PermissionMetricsRead))
}

// CreateServiceToken creates a service token for internal communication
func (i *Integration) CreateServiceToken(serviceID, serviceName string) (string, error) {
	permissions := []string{
		PermissionNodeRead,
		PermissionModelRead,
		PermissionInferenceWrite,
		PermissionClusterRead,
	}

	return i.JWTManager.GenerateServiceToken(serviceID, serviceName, permissions)
}

// CreateAdminToken creates an admin token for administrative tasks
func (i *Integration) CreateAdminToken(adminID, adminName string) (string, error) {
	permissions := DefaultRolePermissions[RoleAdmin]
	return i.JWTManager.GenerateServiceToken(adminID, adminName, permissions)
}

// Close gracefully shuts down the authentication system
func (i *Integration) Close() {
	i.AuthManager.Close()
}

// Example integration with existing server
func ExampleIntegration() {
	// Load configuration
	cfg := &config.AuthConfig{
		Enabled:     true,
		Method:      "jwt",
		TokenExpiry: 24 * 3600, // 24 hours in seconds
		SecretKey:   "your-secret-key",
		Issuer:      "ollama-distributed",
		Audience:    "ollama-api",
	}

	// Create authentication integration
	authIntegration, err := NewIntegration(cfg)
	if err != nil {
		log.Fatalf("Failed to create auth integration: %v", err)
	}
	defer authIntegration.Close()

	// Setup router with authentication
	router := gin.New()

	// Register authentication routes
	authIntegration.Routes.RegisterRoutes(router)

	// Protect existing API routes
	authIntegration.ProtectAPIRoutes(router)

	// Example: Add a protected endpoint
	protected := router.Group("/api/v1/protected")
	protected.Use(authIntegration.MiddlewareManager.AuthRequired())
	protected.Use(authIntegration.MiddlewareManager.RequirePermission(PermissionSystemAdmin))
	{
		protected.GET("/admin-only", func(c *gin.Context) {
			user := GetCurrentUser(c)
			c.JSON(200, gin.H{
				"message": "This is an admin-only endpoint",
				"user":    user.Username,
				"role":    user.Role,
			})
		})
	}

	// Example: Create a service token
	serviceToken, err := authIntegration.CreateServiceToken("node-1", "Ollama Node 1")
	if err != nil {
		log.Printf("Failed to create service token: %v", err)
	} else {
		log.Printf("Service token created: %s", serviceToken)
	}

	// Start server
	log.Println("Starting server with authentication on :8080")
	router.Run(":8080")
}

// MiddlewareHelpers provides helper functions for common middleware patterns
type MiddlewareHelpers struct {
	integration *Integration
}

// NewMiddlewareHelpers creates middleware helpers
func NewMiddlewareHelpers(integration *Integration) *MiddlewareHelpers {
	return &MiddlewareHelpers{integration: integration}
}

// RequireNodePermission creates middleware for node operations
func (mh *MiddlewareHelpers) RequireNodePermission(operation string) gin.HandlerFunc {
	switch operation {
	case "read":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionNodeRead)
	case "write":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionNodeWrite)
	case "admin":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionNodeAdmin)
	default:
		return mh.integration.MiddlewareManager.RequirePermission(PermissionNodeRead)
	}
}

// RequireModelPermission creates middleware for model operations
func (mh *MiddlewareHelpers) RequireModelPermission(operation string) gin.HandlerFunc {
	switch operation {
	case "read":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionModelRead)
	case "write":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionModelWrite)
	case "admin":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionModelAdmin)
	default:
		return mh.integration.MiddlewareManager.RequirePermission(PermissionModelRead)
	}
}

// RequireClusterPermission creates middleware for cluster operations
func (mh *MiddlewareHelpers) RequireClusterPermission(operation string) gin.HandlerFunc {
	switch operation {
	case "read":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionClusterRead)
	case "write":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionClusterWrite)
	case "admin":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionClusterAdmin)
	default:
		return mh.integration.MiddlewareManager.RequirePermission(PermissionClusterRead)
	}
}

// RequireInferencePermission creates middleware for inference operations
func (mh *MiddlewareHelpers) RequireInferencePermission(operation string) gin.HandlerFunc {
	switch operation {
	case "read":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionInferenceRead)
	case "write":
		return mh.integration.MiddlewareManager.RequirePermission(PermissionInferenceWrite)
	default:
		return mh.integration.MiddlewareManager.RequirePermission(PermissionInferenceRead)
	}
}

// Example usage in existing API handlers
func ExampleAPIIntegration(authIntegration *Integration) {
	router := gin.New()
	helpers := NewMiddlewareHelpers(authIntegration)

	// Register auth routes
	authIntegration.Routes.RegisterRoutes(router)

	// Protected API routes
	api := router.Group("/api/v1")
	api.Use(authIntegration.MiddlewareManager.AuthRequired())

	// Node management with granular permissions
	nodes := api.Group("/nodes")
	{
		nodes.GET("", helpers.RequireNodePermission("read"), func(c *gin.Context) {
			// Get nodes logic
			c.JSON(200, gin.H{"nodes": []string{}})
		})

		nodes.POST("", helpers.RequireNodePermission("write"), func(c *gin.Context) {
			// Create node logic
			c.JSON(201, gin.H{"message": "Node created"})
		})

		nodes.DELETE("/:id", helpers.RequireNodePermission("admin"), func(c *gin.Context) {
			// Delete node logic
			c.JSON(200, gin.H{"message": "Node deleted"})
		})
	}

	// Model management with granular permissions
	models := api.Group("/models")
	{
		models.GET("", helpers.RequireModelPermission("read"), func(c *gin.Context) {
			// Get models logic
			c.JSON(200, gin.H{"models": []string{}})
		})

		models.POST("/:name/download", helpers.RequireModelPermission("write"), func(c *gin.Context) {
			// Download model logic
			c.JSON(200, gin.H{"message": "Download started"})
		})

		models.DELETE("/:name", helpers.RequireModelPermission("admin"), func(c *gin.Context) {
			// Delete model logic
			c.JSON(200, gin.H{"message": "Model deleted"})
		})
	}

	// Inference endpoints
	inference := api.Group("/")
	{
		inference.POST("/generate", helpers.RequireInferencePermission("write"), func(c *gin.Context) {
			// Generate logic
			user := GetCurrentUser(c)
			c.JSON(200, gin.H{
				"response": "Generated text",
				"user":     user.Username,
			})
		})

		inference.POST("/chat", helpers.RequireInferencePermission("write"), func(c *gin.Context) {
			// Chat logic
			c.JSON(200, gin.H{"response": "Chat response"})
		})
	}

	// Metrics (read-only)
	api.GET("/metrics", helpers.RequireInferencePermission("read"), func(c *gin.Context) {
		// Metrics logic
		c.JSON(200, gin.H{"metrics": map[string]interface{}{}})
	})
}
