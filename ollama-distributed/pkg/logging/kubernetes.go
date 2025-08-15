package logging

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// KubernetesLogger extends StructuredLogger with Kubernetes-specific features
type KubernetesLogger struct {
	*StructuredLogger
}

// NewKubernetesLogger creates a new Kubernetes-compatible logger
func NewKubernetesLogger(config LoggerConfig) *KubernetesLogger {
	// Ensure JSON format for Kubernetes compatibility
	cfg := config
	cfg.Format = FormatJSON

	structuredLogger, _ := NewStructuredLogger(&cfg)

	return &KubernetesLogger{
		StructuredLogger: structuredLogger,
	}
}

// CorrelationIDKey is the context key for correlation IDs
type CorrelationIDKey string

const (
	// CorrelationID is the key for correlation IDs in context
	CorrelationID CorrelationIDKey = "correlation_id"
	// RequestID is the key for request IDs in context
	RequestID CorrelationIDKey = "request_id"
	// TraceID is the key for trace IDs in context
	TraceID CorrelationIDKey = "trace_id"
)

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	return context.WithValue(ctx, CorrelationID, correlationID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return context.WithValue(ctx, RequestID, requestID)
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	return context.WithValue(ctx, TraceID, traceID)
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(CorrelationID).(string); ok {
		return id
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(RequestID).(string); ok {
		return id
	}
	return ""
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(TraceID).(string); ok {
		return id
	}
	return ""
}

// WithKubernetesContext creates a logger with Kubernetes-specific context
func (kl *KubernetesLogger) WithKubernetesContext(ctx context.Context) *ContextLogger {
	return kl.WithContext(ctx)
}

// LogHealthCheck logs health check events
func (kl *KubernetesLogger) LogHealthCheck(ctx context.Context, checkType, result string, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component":   "health_check",
		"check_type":  checkType,
		"result":      result,
		"duration_ms": duration.Milliseconds(),
		"success":     result == "healthy" || result == "ready" || result == "started",
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)

	if result == "healthy" || result == "ready" || result == "started" {
		logger.Info("Health check passed")
	} else {
		logger.Warn("Health check failed")
	}
}

// LogProbeEvent logs Kubernetes probe events
func (kl *KubernetesLogger) LogProbeEvent(ctx context.Context, probeType string, success bool, httpStatus int, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component":   "kubernetes_probe",
		"probe_type":  probeType,
		"success":     success,
		"http_status": httpStatus,
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)

	if success {
		logger.Debug("Kubernetes probe successful")
	} else {
		logger.Warn("Kubernetes probe failed")
	}
}

// LogFaultToleranceEvent logs fault tolerance events with Kubernetes context
func (kl *KubernetesLogger) LogFaultToleranceEvent(ctx context.Context, event string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["component"] = "fault_tolerance"
	fields["event"] = event

	logger := kl.WithKubernetesContext(ctx)
	logger.Info("Fault tolerance event")
}

// LogHealingAttempt logs healing attempts with Kubernetes context
func (kl *KubernetesLogger) LogHealingAttempt(ctx context.Context, nodeID, strategy, result string, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component":        "fault_tolerance",
		"event":            "healing_attempt",
		"target_node_id":   nodeID,
		"healing_strategy": strategy,
		"result":           result,
		"duration_ms":      duration.Milliseconds(),
		"success":          result == "success",
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)

	if result == "success" {
		logger.Info("Healing attempt successful")
	} else {
		logger.Warn("Healing attempt failed")
	}
}

// LogConfigurationEvent logs configuration events with Kubernetes context
func (kl *KubernetesLogger) LogConfigurationEvent(ctx context.Context, action, component string, success bool, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component": "configuration",
		"action":    action,
		"target":    component,
		"success":   success,
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)

	if success {
		logger.Info("Configuration operation successful")
	} else {
		logger.Error("Configuration operation failed", nil)
	}
}

// HTTPMiddleware creates HTTP middleware for Kubernetes-compatible request logging
func (kl *KubernetesLogger) HTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate correlation ID if not present
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Generate request ID
			requestID := uuid.New().String()

			// Add IDs to context
			ctx := WithCorrelationID(r.Context(), correlationID)
			ctx = WithRequestID(ctx, requestID)
			r = r.WithContext(ctx)

			// Add IDs to response headers
			w.Header().Set("X-Correlation-ID", correlationID)
			w.Header().Set("X-Request-ID", requestID)

			// Wrap response writer to capture status code and size
			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request with Kubernetes context
			duration := time.Since(start)
			kl.LogHTTPRequest(ctx, r.Method, r.URL.Path, wrapped.statusCode, duration, wrapped.size)
		})
	}
}

// LogHTTPRequest logs HTTP requests with Kubernetes context
func (kl *KubernetesLogger) LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, size int64) {
	logger := kl.WithKubernetesContext(ctx)

	if statusCode >= 500 {
		logger.Error("HTTP request failed", nil)
	} else if statusCode >= 400 {
		logger.Warn("HTTP request error")
	} else {
		logger.Info("HTTP request completed")
	}
}

// responseWriter wraps http.ResponseWriter to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

// LogStartup logs application startup events
func (kl *KubernetesLogger) LogStartup(ctx context.Context, phase string, success bool, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component":   "startup",
		"phase":       phase,
		"success":     success,
		"duration_ms": duration.Milliseconds(),
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)

	if success {
		logger.Info("Startup phase completed")
	} else {
		logger.Error("Startup phase failed", nil)
	}
}

// LogShutdown logs application shutdown events
func (kl *KubernetesLogger) LogShutdown(ctx context.Context, phase string, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"component":   "shutdown",
		"phase":       phase,
		"duration_ms": duration.Milliseconds(),
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	logger := kl.WithKubernetesContext(ctx)
	logger.Info("Shutdown phase completed")
}
