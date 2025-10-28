package api

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
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

// limiterEntry tracks a rate limiter with its last access time for LRU eviction
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// rateLimitMiddleware implements rate limiting per IP with endpoint-specific limits
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	// SECURITY FIX (ISSUE-007): Endpoint-specific rate limiting for auth routes
	// MEMORY FIX (ISSUE-008): LRU cache with TTL eviction to prevent unbounded growth

	// Create separate limiter maps for general and auth-specific endpoints
	generalLimiters := make(map[string]*limiterEntry)
	loginLimiters := make(map[string]*limiterEntry)
	registerLimiters := make(map[string]*limiterEntry)
	resetPasswordLimiters := make(map[string]*limiterEntry)

	// Mutex to protect concurrent access to limiter maps
	var mu sync.RWMutex

	// Configuration for LRU cache with TTL
	const (
		maxIPEntries  = 10000           // Maximum number of IP entries to store
		ipEntryTTL    = 1 * time.Hour   // TTL for inactive IP entries
		cleanupPeriod = 10 * time.Minute // How often to run cleanup
	)

	// Background goroutine for periodic cleanup of expired entries
	go func() {
		ticker := time.NewTicker(cleanupPeriod)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			mu.Lock()

			// Cleanup function for a single map
			cleanupMap := func(limiters map[string]*limiterEntry) int {
				removed := 0
				for ip, entry := range limiters {
					if now.Sub(entry.lastAccess) > ipEntryTTL {
						delete(limiters, ip)
						removed++
					}
				}
				return removed
			}

			// Clean up all limiter maps
			generalRemoved := cleanupMap(generalLimiters)
			loginRemoved := cleanupMap(loginLimiters)
			registerRemoved := cleanupMap(registerLimiters)
			resetRemoved := cleanupMap(resetPasswordLimiters)

			totalRemoved := generalRemoved + loginRemoved + registerRemoved + resetRemoved
			if totalRemoved > 0 {
				s.logger.Info("Rate limiter cache cleanup",
					"removed_entries", totalRemoved,
					"general", generalRemoved,
					"login", loginRemoved,
					"register", registerRemoved,
					"reset_password", resetRemoved,
				)
			}

			mu.Unlock()
		}
	}()

	// Helper function to get or create limiter with LRU+TTL
	getLimiter := func(limiters map[string]*limiterEntry, clientIP string, requestsPer int, duration time.Duration, burstSize int) *rate.Limiter {
		mu.RLock()
		entry, exists := limiters[clientIP]
		mu.RUnlock()

		if exists {
			// Update last access time
			mu.Lock()
			entry.lastAccess = time.Now()
			mu.Unlock()
			return entry.limiter
		}

		// Create new limiter
		limiter := rate.NewLimiter(
			rate.Limit(requestsPer)/rate.Limit(duration.Seconds()),
			burstSize,
		)

		mu.Lock()
		defer mu.Unlock()

		// Check if we need to evict entries (LRU eviction)
		if len(limiters) >= maxIPEntries {
			// Find and remove oldest entry
			var oldestIP string
			var oldestTime time.Time = time.Now()

			for ip, entry := range limiters {
				if entry.lastAccess.Before(oldestTime) {
					oldestTime = entry.lastAccess
					oldestIP = ip
				}
			}

			if oldestIP != "" {
				delete(limiters, oldestIP)
				s.logger.Warn("Rate limiter cache eviction (LRU)",
					"evicted_ip", oldestIP,
					"cache_size", len(limiters),
					"max_entries", maxIPEntries,
				)
			}
		}

		// Store new limiter with current timestamp
		limiters[clientIP] = &limiterEntry{
			limiter:    limiter,
			lastAccess: time.Now(),
		}

		return limiter
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.FullPath()

		var limiter *rate.Limiter
		var retryAfter int

		// SECURITY: Apply stricter rate limits for authentication endpoints
		switch path {
		case "/api/auth/login":
			limiter = getLimiter(loginLimiters, clientIP,
				s.config.API.RateLimit.LoginRequestsPer,
				time.Minute,
				1) // Burst of 1 for login
			retryAfter = 60 // 1 minute

		case "/api/auth/register":
			limiter = getLimiter(registerLimiters, clientIP,
				s.config.API.RateLimit.RegisterRequestsPer,
				time.Minute,
				1) // Burst of 1 for registration
			retryAfter = 60 // 1 minute

		case "/api/auth/reset-password", "/api/auth/forgot-password":
			limiter = getLimiter(resetPasswordLimiters, clientIP,
				s.config.API.RateLimit.ResetPasswordRequestsPer,
				time.Minute,
				1) // Burst of 1 for password reset
			retryAfter = 60 // 1 minute

		default:
			// General rate limiting for all other endpoints
			limiter = getLimiter(generalLimiters, clientIP,
				s.config.API.RateLimit.RequestsPer,
				s.config.API.RateLimit.Duration,
				s.config.API.RateLimit.BurstSize)
			retryAfter = int(s.config.API.RateLimit.Duration.Seconds())
		}

		// Check if request is allowed
		if !limiter.Allow() {
			s.logger.Warn("Rate limit exceeded",
				"client_ip", clientIP,
				"path", path,
				"method", c.Request.Method,
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": retryAfter,
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
// NOTE: This is a placeholder. For production, use github.com/gin-contrib/gzip middleware
func (s *Server) compressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement actual gzip compression using github.com/gin-contrib/gzip
		// For now, this is a placeholder that doesn't compress
		// PERFORMANCE: Add gzip middleware to save 70-85% bandwidth
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
