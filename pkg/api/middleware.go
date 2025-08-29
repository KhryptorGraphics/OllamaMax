package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// loggingMiddleware provides structured request logging
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		s.logger.Info("HTTP request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)
		return ""
	})
}

// corsMiddleware configures CORS based on application configuration
func (s *Server) corsMiddleware() gin.HandlerFunc {
	if !s.config.API.Cors.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:     s.config.API.Cors.AllowedOrigins,
		AllowMethods:     s.config.API.Cors.AllowedMethods,
		AllowHeaders:     s.config.API.Cors.AllowedHeaders,
		AllowCredentials: s.config.API.Cors.AllowCredentials,
		MaxAge:           time.Duration(s.config.API.Cors.MaxAge) * time.Second,
	}

	// Handle wildcard origins properly
	if len(corsConfig.AllowOrigins) == 1 && corsConfig.AllowOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
	}

	return cors.New(corsConfig)
}

// securityMiddleware adds security headers
func (s *Server) securityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// Remove server information
		c.Header("Server", "OllamaMax")

		c.Next()
	}
}

// rateLimitMiddleware implements rate limiting per IP
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	// Create rate limiter map for different IPs
	limiters := make(map[string]*rate.Limiter)

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Get or create limiter for this IP
		limiter, exists := limiters[clientIP]
		if !exists {
			// Create new limiter: requests per duration with burst size
			limiter = rate.NewLimiter(
				rate.Limit(s.config.API.RateLimit.RequestsPer)/rate.Limit(s.config.API.RateLimit.Duration.Seconds()),
				s.config.API.RateLimit.BurstSize,
			)
			limiters[clientIP] = limiter
		}

		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
				"retry_after": int(s.config.API.RateLimit.Duration.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// requestSizeMiddleware limits request body size
func (s *Server) requestSizeMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.config.API.MaxBodySize)
		c.Next()
	})
}

// auditMiddleware logs all requests for audit purposes
func (s *Server) auditMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Get user ID if authenticated
		userID, exists := c.Get("user_id")
		var userIDStr *string
		if exists {
			if uid, ok := userID.(string); ok {
				userIDStr = &uid
			}
		}

		// Create audit log entry
		auditEntry := &database.AuditLogEntry{
			Operation: strings.ToUpper(c.Request.Method),
			TableName: "api_requests",
			UserID:    userIDStr,
			IPAddress: &c.ClientIP,
			UserAgent: &c.Request.UserAgent,
			NewValues: &database.JSONMap{
				"path":        c.Request.URL.Path,
				"method":      c.Request.Method,
				"status_code": c.Writer.Status(),
				"duration_ms": time.Since(start).Milliseconds(),
			},
			Timestamp: time.Now(),
		}

		// Log to audit repository (async to not block requests)
		go func() {
			if err := s.db.Audit.Create(c.Request.Context(), auditEntry); err != nil {
				s.logger.Error("Failed to create audit log", "error", err)
			}
		}()
	})
}

// contentTypeMiddleware ensures proper content type handling
func (s *Server) contentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For API endpoints, ensure JSON content type for POST/PUT/PATCH
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			method := c.Request.Method
			if method == "POST" || method == "PUT" || method == "PATCH" {
				contentType := c.GetHeader("Content-Type")
				if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "multipart/form-data") {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "invalid_content_type",
						"message": "Content-Type must be application/json for API endpoints",
					})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}

// versionMiddleware adds API version information to responses
func (s *Server) versionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", "v1")
		c.Header("X-Service-Version", "1.0.0")
		c.Next()
	}
}

// compressionMiddleware handles response compression
func (s *Server) compressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client accepts gzip
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			c.Header("Content-Encoding", "gzip")
			c.Header("Vary", "Accept-Encoding")
		}
		c.Next()
	}
}

// healthCheckMiddleware provides detailed health information
func (s *Server) healthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			health, err := s.db.Health(c.Request.Context())
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "unhealthy",
					"error":  err.Error(),
				})
				c.Abort()
				return
			}

			status := http.StatusOK
			if health.Overall != "healthy" {
				status = http.StatusServiceUnavailable
			}

			c.JSON(status, gin.H{
				"status":    health.Overall,
				"timestamp": time.Now(),
				"services":  health,
				"version":   "1.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
