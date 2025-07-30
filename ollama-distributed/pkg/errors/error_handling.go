package errors

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization ErrorType = "authorization"
	ErrorTypeNotFound      ErrorType = "not_found"
	ErrorTypeConflict      ErrorType = "conflict"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeExternal      ErrorType = "external"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
	ErrorTypeUnavailable   ErrorType = "unavailable"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// DistributedError represents a comprehensive error with context and metadata
type DistributedError struct {
	// Core error information
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Type        ErrorType              `json:"type"`
	Severity    ErrorSeverity          `json:"severity"`
	
	// Context information
	Service     string                 `json:"service"`
	Operation   string                 `json:"operation"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	
	// Technical details
	Cause       error                  `json:"cause,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	
	// Timing information
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration,omitempty"`
	
	// Additional metadata
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	
	// Retry information
	Retryable   bool                   `json:"retryable"`
	RetryAfter  time.Duration          `json:"retry_after,omitempty"`
	
	// HTTP status code (if applicable)
	HTTPStatus  int                    `json:"http_status,omitempty"`
}

// Error implements the error interface
func (e *DistributedError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *DistributedError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target
func (e *DistributedError) Is(target error) bool {
	if t, ok := target.(*DistributedError); ok {
		return e.Code == t.Code && e.Type == t.Type
	}
	return false
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	err *DistributedError
}

// NewError creates a new error builder
func NewError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &DistributedError{
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}
}

// WithType sets the error type
func (eb *ErrorBuilder) WithType(errorType ErrorType) *ErrorBuilder {
	eb.err.Type = errorType
	return eb
}

// WithSeverity sets the error severity
func (eb *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	eb.err.Severity = severity
	return eb
}

// WithService sets the service name
func (eb *ErrorBuilder) WithService(service string) *ErrorBuilder {
	eb.err.Service = service
	return eb
}

// WithOperation sets the operation name
func (eb *ErrorBuilder) WithOperation(operation string) *ErrorBuilder {
	eb.err.Operation = operation
	return eb
}

// WithCause sets the underlying cause
func (eb *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	eb.err.Cause = cause
	return eb
}

// WithContext extracts information from context
func (eb *ErrorBuilder) WithContext(ctx context.Context) *ErrorBuilder {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			eb.err.RequestID = id
		}
	}
	
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			eb.err.UserID = id
		}
	}
	
	return eb
}

// WithMetadata adds metadata
func (eb *ErrorBuilder) WithMetadata(key string, value interface{}) *ErrorBuilder {
	eb.err.Metadata[key] = value
	return eb
}

// WithHTTPStatus sets the HTTP status code
func (eb *ErrorBuilder) WithHTTPStatus(status int) *ErrorBuilder {
	eb.err.HTTPStatus = status
	return eb
}

// WithRetry marks the error as retryable
func (eb *ErrorBuilder) WithRetry(retryable bool, retryAfter time.Duration) *ErrorBuilder {
	eb.err.Retryable = retryable
	eb.err.RetryAfter = retryAfter
	return eb
}

// WithStackTrace captures the stack trace
func (eb *ErrorBuilder) WithStackTrace() *ErrorBuilder {
	eb.err.StackTrace = captureStackTrace()
	return eb
}

// Build creates the final error
func (eb *ErrorBuilder) Build() *DistributedError {
	// Set default values
	if eb.err.Type == "" {
		eb.err.Type = ErrorTypeInternal
	}
	
	if eb.err.Severity == "" {
		eb.err.Severity = SeverityMedium
	}
	
	// Auto-capture stack trace for high severity errors
	if eb.err.Severity == SeverityHigh || eb.err.Severity == SeverityCritical {
		if eb.err.StackTrace == "" {
			eb.err.StackTrace = captureStackTrace()
		}
	}
	
	return eb.err
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	config    *ErrorHandlerConfig
	reporters []ErrorReporter
	mu        sync.RWMutex
}

// ErrorHandlerConfig configures the error handler
type ErrorHandlerConfig struct {
	EnableStackTrace    bool
	EnableReporting     bool
	ReportingThreshold  ErrorSeverity
	MaxStackDepth       int
	SampleRate          float64
}

// ErrorReporter interface for error reporting
type ErrorReporter interface {
	Report(ctx context.Context, err *DistributedError) error
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(config *ErrorHandlerConfig) *ErrorHandler {
	if config == nil {
		config = &ErrorHandlerConfig{
			EnableStackTrace:   true,
			EnableReporting:    true,
			ReportingThreshold: SeverityHigh,
			MaxStackDepth:      50,
			SampleRate:         1.0,
		}
	}
	
	return &ErrorHandler{
		config:    config,
		reporters: make([]ErrorReporter, 0),
	}
}

// AddReporter adds an error reporter
func (eh *ErrorHandler) AddReporter(reporter ErrorReporter) {
	eh.mu.Lock()
	defer eh.mu.Unlock()
	eh.reporters = append(eh.reporters, reporter)
}

// Handle handles an error with context
func (eh *ErrorHandler) Handle(ctx context.Context, err error) *DistributedError {
	// Convert to DistributedError if needed
	var distErr *DistributedError
	if de, ok := err.(*DistributedError); ok {
		distErr = de
	} else {
		distErr = NewError("UNKNOWN_ERROR", err.Error()).
			WithType(ErrorTypeInternal).
			WithSeverity(SeverityMedium).
			WithCause(err).
			Build()
	}
	
	// Add stack trace if enabled and not already present
	if eh.config.EnableStackTrace && distErr.StackTrace == "" {
		distErr.StackTrace = captureStackTrace()
	}
	
	// Report error if enabled and meets threshold
	if eh.config.EnableReporting && eh.shouldReport(distErr) {
		eh.reportError(ctx, distErr)
	}
	
	return distErr
}

// shouldReport determines if an error should be reported
func (eh *ErrorHandler) shouldReport(err *DistributedError) bool {
	// Check severity threshold
	severityLevels := map[ErrorSeverity]int{
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}
	
	errLevel := severityLevels[err.Severity]
	thresholdLevel := severityLevels[eh.config.ReportingThreshold]
	
	return errLevel >= thresholdLevel
}

// reportError reports an error to all configured reporters
func (eh *ErrorHandler) reportError(ctx context.Context, err *DistributedError) {
	eh.mu.RLock()
	reporters := make([]ErrorReporter, len(eh.reporters))
	copy(reporters, eh.reporters)
	eh.mu.RUnlock()
	
	for _, reporter := range reporters {
		go func(r ErrorReporter) {
			if reportErr := r.Report(ctx, err); reportErr != nil {
				// Log reporting error (avoid infinite recursion)
				fmt.Printf("Failed to report error: %v\n", reportErr)
			}
		}(reporter)
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	
	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])
	
	for {
		frame, more := frames.Next()
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		
		if !more {
			break
		}
	}
	
	return sb.String()
}

// Common error constructors

// ValidationError creates a validation error
func ValidationError(field, message string) *DistributedError {
	return NewError("VALIDATION_ERROR", fmt.Sprintf("Validation failed for field '%s': %s", field, message)).
		WithType(ErrorTypeValidation).
		WithSeverity(SeverityLow).
		WithHTTPStatus(400).
		Build()
}

// NotFoundError creates a not found error
func NotFoundError(resource, id string) *DistributedError {
	return NewError("NOT_FOUND", fmt.Sprintf("%s with id '%s' not found", resource, id)).
		WithType(ErrorTypeNotFound).
		WithSeverity(SeverityLow).
		WithHTTPStatus(404).
		Build()
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *DistributedError {
	return NewError("UNAUTHORIZED", message).
		WithType(ErrorTypeAuthentication).
		WithSeverity(SeverityMedium).
		WithHTTPStatus(401).
		Build()
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string) *DistributedError {
	return NewError("FORBIDDEN", message).
		WithType(ErrorTypeAuthorization).
		WithSeverity(SeverityMedium).
		WithHTTPStatus(403).
		Build()
}

// ConflictError creates a conflict error
func ConflictError(resource, message string) *DistributedError {
	return NewError("CONFLICT", fmt.Sprintf("Conflict with %s: %s", resource, message)).
		WithType(ErrorTypeConflict).
		WithSeverity(SeverityMedium).
		WithHTTPStatus(409).
		Build()
}

// InternalError creates an internal server error
func InternalError(message string, cause error) *DistributedError {
	return NewError("INTERNAL_ERROR", message).
		WithType(ErrorTypeInternal).
		WithSeverity(SeverityHigh).
		WithCause(cause).
		WithHTTPStatus(500).
		WithStackTrace().
		Build()
}

// NetworkError creates a network error
func NetworkError(operation string, cause error) *DistributedError {
	return NewError("NETWORK_ERROR", fmt.Sprintf("Network error during %s", operation)).
		WithType(ErrorTypeNetwork).
		WithSeverity(SeverityMedium).
		WithCause(cause).
		WithRetry(true, 5*time.Second).
		Build()
}

// TimeoutError creates a timeout error
func TimeoutError(operation string, timeout time.Duration) *DistributedError {
	return NewError("TIMEOUT", fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout)).
		WithType(ErrorTypeTimeout).
		WithSeverity(SeverityMedium).
		WithRetry(true, 10*time.Second).
		WithHTTPStatus(408).
		Build()
}

// RateLimitError creates a rate limit error
func RateLimitError(retryAfter time.Duration) *DistributedError {
	return NewError("RATE_LIMIT", "Rate limit exceeded").
		WithType(ErrorTypeRateLimit).
		WithSeverity(SeverityLow).
		WithRetry(true, retryAfter).
		WithHTTPStatus(429).
		Build()
}

// UnavailableError creates a service unavailable error
func UnavailableError(service string, retryAfter time.Duration) *DistributedError {
	return NewError("SERVICE_UNAVAILABLE", fmt.Sprintf("Service '%s' is temporarily unavailable", service)).
		WithType(ErrorTypeUnavailable).
		WithSeverity(SeverityHigh).
		WithRetry(true, retryAfter).
		WithHTTPStatus(503).
		Build()
}
