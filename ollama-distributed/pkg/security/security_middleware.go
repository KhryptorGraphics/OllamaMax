package security

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// SecurityMiddleware provides comprehensive security for HTTP endpoints
type SecurityMiddleware struct {
	logger                 *slog.Logger
	httpsEnforcement       *HTTPSEnforcement
	inputValidator         *ComprehensiveInputValidator
	sqlInjectionPrevention *SQLInjectionPrevention

	// Rate limiting
	rateLimiter *RateLimiter

	// Security configuration
	config *SecurityConfig
}

// SecurityConfig contains security middleware configuration
type SecurityConfig struct {
	// HTTPS enforcement
	EnforceHTTPS bool `json:"enforce_https"`

	// Input validation
	ValidateInput  bool  `json:"validate_input"`
	MaxRequestSize int64 `json:"max_request_size"`

	// Rate limiting
	EnableRateLimit   bool `json:"enable_rate_limit"`
	RequestsPerMinute int  `json:"requests_per_minute"`

	// SQL injection prevention
	EnableSQLProtection bool `json:"enable_sql_protection"`

	// Request logging
	LogRequests      bool `json:"log_requests"`
	LogSensitiveData bool `json:"log_sensitive_data"`

	// CORS settings
	EnableCORS     bool     `json:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// RateLimiter provides simple rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config *SecurityConfig, logger *slog.Logger) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	sm := &SecurityMiddleware{
		logger:                 logger,
		httpsEnforcement:       NewHTTPSEnforcement(nil),
		inputValidator:         NewComprehensiveInputValidator(),
		sqlInjectionPrevention: NewSQLInjectionPrevention(),
		config:                 config,
	}

	// Initialize rate limiter if enabled
	if config.EnableRateLimit {
		sm.rateLimiter = &RateLimiter{
			requests: make(map[string][]time.Time),
			limit:    config.RequestsPerMinute,
			window:   time.Minute,
		}
	}

	return sm
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnforceHTTPS:        true,
		ValidateInput:       true,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		EnableRateLimit:     true,
		RequestsPerMinute:   100,
		EnableSQLProtection: true,
		LogRequests:         true,
		LogSensitiveData:    false,
		EnableCORS:          true,
		AllowedOrigins:      []string{"*"},
		AllowedMethods:      []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:      []string{"Content-Type", "Authorization", "X-Requested-With"},
	}
}

// Middleware returns the HTTP middleware function
func (sm *SecurityMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start request processing
			start := time.Now()

			// Log request if enabled
			if sm.config.LogRequests {
				sm.logRequest(r)
			}

			// HTTPS enforcement
			if sm.config.EnforceHTTPS {
				if err := sm.httpsEnforcement.ValidateHTTPSRequest(r); err != nil {
					sm.writeErrorResponse(w, http.StatusBadRequest, "HTTPS required", err)
					return
				}
			}

			// Rate limiting
			if sm.config.EnableRateLimit {
				if !sm.checkRateLimit(r) {
					sm.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", nil)
					return
				}
			}

			// Request size validation
			if r.ContentLength > sm.config.MaxRequestSize {
				sm.writeErrorResponse(w, http.StatusRequestEntityTooLarge, "Request too large", nil)
				return
			}

			// Input validation for specific endpoints
			if sm.config.ValidateInput {
				if err := sm.validateRequestInput(r); err != nil {
					sm.writeErrorResponse(w, http.StatusBadRequest, "Invalid input", err)
					return
				}
			}

			// CORS handling
			if sm.config.EnableCORS {
				sm.handleCORS(w, r)
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			// Add security headers
			sm.addSecurityHeaders(w)

			// Process request
			next.ServeHTTP(w, r)

			// Log response time
			duration := time.Since(start)
			sm.logger.Debug("request processed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", duration,
				"remote_addr", r.RemoteAddr)
		})
	}
}

// validateRequestInput validates request input based on endpoint
func (sm *SecurityMiddleware) validateRequestInput(r *http.Request) error {
	// Validate URL path
	if result := sm.inputValidator.ValidateInput(r.URL.Path, "api_endpoint"); !result.Valid {
		return fmt.Errorf("invalid URL path: %v", result.Errors)
	}

	// Validate query parameters
	for key, values := range r.URL.Query() {
		for _, value := range values {
			if sm.config.EnableSQLProtection {
				if err := sm.sqlInjectionPrevention.ValidateInput(value, "safe_text"); err != nil {
					return fmt.Errorf("invalid query parameter %s: %w", key, err)
				}
			}
		}
	}

	// Validate request body for POST/PUT requests
	if r.Method == "POST" || r.Method == "PUT" {
		if err := sm.validateRequestBody(r); err != nil {
			return err
		}
	}

	return nil
}

// validateRequestBody validates the request body
func (sm *SecurityMiddleware) validateRequestBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Restore body for next handler
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// Validate JSON if content type is JSON
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if result := sm.inputValidator.ValidateJSON(body); !result.Valid {
			return fmt.Errorf("invalid JSON: %v", result.Errors)
		}
	}

	// Check for SQL injection patterns in body
	if sm.config.EnableSQLProtection {
		bodyStr := string(body)
		if err := sm.sqlInjectionPrevention.ValidateInput(bodyStr, "user_input"); err != nil {
			return fmt.Errorf("potentially dangerous content in request body: %w", err)
		}
	}

	return nil
}

// checkRateLimit checks if the request is within rate limits
func (sm *SecurityMiddleware) checkRateLimit(r *http.Request) bool {
	if sm.rateLimiter == nil {
		return true
	}

	clientIP := sm.getClientIP(r)
	now := time.Now()

	// Clean old requests
	if requests, exists := sm.rateLimiter.requests[clientIP]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < sm.rateLimiter.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		sm.rateLimiter.requests[clientIP] = validRequests
	}

	// Check if under limit
	if len(sm.rateLimiter.requests[clientIP]) >= sm.rateLimiter.limit {
		return false
	}

	// Add current request
	sm.rateLimiter.requests[clientIP] = append(sm.rateLimiter.requests[clientIP], now)

	return true
}

// getClientIP extracts the client IP address
func (sm *SecurityMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use remote address
	return r.RemoteAddr
}

// handleCORS handles CORS headers
func (sm *SecurityMiddleware) handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range sm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}
}

// addSecurityHeaders adds security headers to the response
func (sm *SecurityMiddleware) addSecurityHeaders(w http.ResponseWriter) {
	// Use HTTPS enforcement middleware for security headers
	middleware := sm.httpsEnforcement.SecurityHeadersMiddleware()

	// Create a dummy handler to apply headers
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middleware(dummyHandler).ServeHTTP(w, &http.Request{})
}

// logRequest logs the incoming request
func (sm *SecurityMiddleware) logRequest(r *http.Request) {
	fields := []slog.Attr{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
	}

	// Add query parameters if not sensitive
	if !sm.config.LogSensitiveData && len(r.URL.RawQuery) > 0 {
		fields = append(fields, slog.String("query", r.URL.RawQuery))
	}

	sm.logger.LogAttrs(nil, slog.LevelInfo, "incoming request", fields...)
}

// writeErrorResponse writes a standardized error response
func (sm *SecurityMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"error":     message,
		"status":    statusCode,
		"timestamp": time.Now().UTC(),
	}

	// Add error details if not sensitive
	if err != nil && !sm.config.LogSensitiveData {
		errorResponse["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)

	// Log the error
	sm.logger.Warn("security error",
		"status", statusCode,
		"message", message,
		"error", err)
}

// Global instance for easy access
var DefaultSecurityMiddleware = NewSecurityMiddleware(nil, slog.Default())

// Convenience function
func GetSecurityMiddleware() func(http.Handler) http.Handler {
	return DefaultSecurityMiddleware.Middleware()
}
