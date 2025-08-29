# Ollama Distributed Developer Track Certification (Professional Level)
# 💻 Complete Training & Assessment Guide

## 📋 Certification Overview

**Track**: Developer Track  
**Level**: Professional  
**Duration**: 120 minutes instruction + 60 minutes assessment  
**Target Audience**: Software developers, API engineers, integration specialists  
**Prerequisites**: User Track certification + Programming experience (Go/Python/JavaScript)  

### Certification Objectives
Upon successful completion, certified professionals will demonstrate:
- ✅ Proficiency in API client development and integration
- ✅ Expertise in developing extensions and tools
- ✅ Advanced testing and validation capabilities
- ✅ Understanding of distributed system development
- ✅ Ability to contribute to the open source project

---

## 📚 Module 2.1: Development Environment (20 minutes)

### Learning Objectives
- Configure complete development environment
- Set up testing infrastructure  
- Navigate and understand codebase structure
- Use development tools effectively

### Pre-Module Requirements
```bash
# Verify prerequisites
go version           # Should be 1.21+
git --version       # Should be 2.0+
docker --version    # Optional but recommended
make --version      # Build automation tool
```

### Hands-On Exercise 2.1.A: Development Setup

```bash
#!/bin/bash
# Development environment setup script

echo "🔧 Setting up Ollama Distributed Development Environment"

# 1. Fork and clone the repository
echo "Step 1: Repository setup"
git clone https://github.com/KhryptorGraphics/ollamamax.git
cd ollamamax/ollama-distributed

# 2. Set up Git hooks
echo "Step 2: Installing Git hooks"
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for code quality
echo "Running pre-commit checks..."

# Format check
gofmt -l . > /tmp/gofmt_output
if [ -s /tmp/gofmt_output ]; then
    echo "❌ Code formatting issues found:"
    cat /tmp/gofmt_output
    exit 1
fi

# Lint check
golint ./... > /tmp/golint_output
if [ -s /tmp/golint_output ]; then
    echo "⚠️ Linting warnings:"
    cat /tmp/golint_output
fi

# Test check
go test ./... -short
if [ $? -ne 0 ]; then
    echo "❌ Tests failed"
    exit 1
fi

echo "✅ Pre-commit checks passed"
EOF
chmod +x .git/hooks/pre-commit

# 3. Install development dependencies
echo "Step 3: Installing development tools"
go install golang.org/x/lint/golint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/goreleaser/goreleaser@latest

# 4. Set up workspace
echo "Step 4: Creating development workspace"
mkdir -p dev/{scripts,tests,docs}

# 5. Create development configuration
cat > dev/config.dev.yaml << 'EOF'
# Development Configuration
node:
  name: "dev-node"
  environment: "development"
  
api:
  port: 8080
  debug: true
  verbose: true
  
logging:
  level: "debug"
  format: "json"
  output: "stdout"
  
monitoring:
  enabled: true
  metrics_port: 9090
  
development:
  hot_reload: true
  profiling: true
  trace: true
EOF

echo "✅ Development environment ready!"
```

### Hands-On Exercise 2.1.B: Codebase Navigation

```go
// Understanding the codebase structure
/*
ollama-distributed/
├── cmd/                    // Entry points
│   ├── distributed-ollama/ // Main CLI application
│   └── node/              // Node-specific commands
├── pkg/                   // Public packages
│   ├── api/              // API server implementation
│   ├── p2p/              // P2P networking
│   ├── config/           // Configuration management
│   ├── model/            // Model management (future)
│   └── utils/            // Utility functions
├── internal/             // Private packages
│   ├── core/            // Core business logic
│   ├── storage/         // Data persistence
│   └── metrics/         // Monitoring and metrics
├── tests/               // Test suites
│   ├── unit/           // Unit tests
│   ├── integration/    // Integration tests
│   └── e2e/           // End-to-end tests
└── docs/              // Documentation
*/

// Key files to understand:
// main.go - Entry point
// pkg/api/server.go - API server
// pkg/p2p/node.go - P2P networking
// pkg/config/config.go - Configuration
```

### Development Workflow

```makefile
# Makefile for common development tasks
.PHONY: build test lint clean run

# Build the project
build:
	go build -o bin/ollama-distributed ./cmd/distributed-ollama

# Run tests
test:
	go test ./... -v -cover

# Run linter
lint:
	golint ./...
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/ *.log

# Run in development mode
run-dev:
	go run ./cmd/distributed-ollama --config dev/config.dev.yaml

# Generate code coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test ./... -bench=. -benchmem
```

### Knowledge Check 2.1
1. **What Go version is required?** 1.21+
2. **Where are public packages located?** `pkg/` directory
3. **What tool formats Go code?** `gofmt` or `goimports`
4. **How to run tests with coverage?** `go test ./... -cover`
5. **What's the purpose of internal packages?** Private implementation details

**✅ Module 2.1 Complete**: Development environment configured

---

## 📚 Module 2.2: API Development (30 minutes)

### Learning Objectives
- Build comprehensive API clients
- Implement error handling and resilience
- Create monitoring and logging systems
- Master JSON processing and validation

### Hands-On Exercise 2.2.A: API Client Implementation

```go
// api_client.go - Complete API client implementation
package ollamaclient

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// Client represents an Ollama Distributed API client
type Client struct {
    baseURL    string
    httpClient *http.Client
    apiKey     string
    debug      bool
}

// NewClient creates a new API client
func NewClient(baseURL string, options ...ClientOption) *Client {
    client := &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
    
    // Apply options
    for _, opt := range options {
        opt(client)
    }
    
    return client
}

// ClientOption is a functional option for configuring the client
type ClientOption func(*Client)

// WithTimeout sets custom timeout
func WithTimeout(timeout time.Duration) ClientOption {
    return func(c *Client) {
        c.httpClient.Timeout = timeout
    }
}

// WithAPIKey sets API key for authentication
func WithAPIKey(key string) ClientOption {
    return func(c *Client) {
        c.apiKey = key
    }
}

// WithDebug enables debug mode
func WithDebug(debug bool) ClientOption {
    return func(c *Client) {
        c.debug = debug
    }
}

// Health checks the health of the node
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
    resp, err := c.doRequest(ctx, "GET", "/health", nil)
    if err != nil {
        return nil, fmt.Errorf("health check failed: %w", err)
    }
    defer resp.Body.Close()
    
    var health HealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &health, nil
}

// Status gets the current node status
func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
    resp, err := c.doRequest(ctx, "GET", "/status", nil)
    if err != nil {
        return nil, fmt.Errorf("status check failed: %w", err)
    }
    defer resp.Body.Close()
    
    var status StatusResponse
    if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &status, nil
}

// Metrics retrieves current metrics
func (c *Client) Metrics(ctx context.Context) (*MetricsResponse, error) {
    resp, err := c.doRequest(ctx, "GET", "/metrics", nil)
    if err != nil {
        return nil, fmt.Errorf("metrics retrieval failed: %w", err)
    }
    defer resp.Body.Close()
    
    var metrics MetricsResponse
    if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &metrics, nil
}

// UpdateConfig updates the node configuration
func (c *Client) UpdateConfig(ctx context.Context, config *NodeConfig) error {
    resp, err := c.doRequest(ctx, "POST", "/config", config)
    if err != nil {
        return fmt.Errorf("config update failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("config update failed with status %d: %s", 
            resp.StatusCode, string(body))
    }
    
    return nil
}

// doRequest performs an HTTP request with proper error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
    url := c.baseURL + path
    
    var bodyReader io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        bodyReader = bytes.NewReader(jsonBody)
        
        if c.debug {
            fmt.Printf("Request: %s %s\nBody: %s\n", method, url, string(jsonBody))
        }
    }
    
    req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    if c.apiKey != "" {
        req.Header.Set("Authorization", "Bearer " + c.apiKey)
    }
    
    // Perform request with retry logic
    var resp *http.Response
    retries := 3
    for i := 0; i < retries; i++ {
        resp, err = c.httpClient.Do(req)
        if err == nil && resp.StatusCode < 500 {
            break
        }
        if i < retries-1 {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    if err != nil {
        return nil, fmt.Errorf("request failed after %d retries: %w", retries, err)
    }
    
    if c.debug {
        fmt.Printf("Response: %d %s\n", resp.StatusCode, resp.Status)
    }
    
    return resp, nil
}

// Response types
type HealthResponse struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version"`
}

type StatusResponse struct {
    NodeID      string                 `json:"node_id"`
    State       string                 `json:"state"`
    Connections int                    `json:"connections"`
    Uptime      time.Duration         `json:"uptime"`
    Resources   map[string]interface{} `json:"resources"`
}

type MetricsResponse struct {
    CPU       float64           `json:"cpu_usage"`
    Memory    float64           `json:"memory_usage"`
    Disk      float64           `json:"disk_usage"`
    Network   NetworkMetrics    `json:"network"`
    Custom    map[string]float64 `json:"custom"`
}

type NetworkMetrics struct {
    BytesIn  int64 `json:"bytes_in"`
    BytesOut int64 `json:"bytes_out"`
    Latency  float64 `json:"latency_ms"`
}

type NodeConfig struct {
    Name        string            `json:"name"`
    APIPort     int              `json:"api_port"`
    WebPort     int              `json:"web_port"`
    MaxMemory   int              `json:"max_memory"`
    EnableGPU   bool             `json:"enable_gpu"`
    Custom      map[string]string `json:"custom"`
}
```

### Hands-On Exercise 2.2.B: Error Handling & Resilience

```go
// error_handling.go - Advanced error handling patterns
package ollamaclient

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"
)

// Custom error types
var (
    ErrConnection     = errors.New("connection error")
    ErrTimeout       = errors.New("request timeout")
    ErrNotFound      = errors.New("resource not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrRateLimit     = errors.New("rate limit exceeded")
    ErrServerError   = errors.New("server error")
)

// ErrorHandler provides structured error handling
type ErrorHandler struct {
    logger     *log.Logger
    metrics    *MetricsCollector
    retryPolicy RetryPolicy
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
    MaxRetries     int
    InitialDelay   time.Duration
    MaxDelay       time.Duration
    BackoffFactor  float64
}

// DefaultRetryPolicy returns sensible defaults
func DefaultRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxRetries:    3,
        InitialDelay:  1 * time.Second,
        MaxDelay:      30 * time.Second,
        BackoffFactor: 2.0,
    }
}

// HandleWithRetry executes a function with retry logic
func (eh *ErrorHandler) HandleWithRetry(ctx context.Context, fn func() error) error {
    var lastErr error
    delay := eh.retryPolicy.InitialDelay
    
    for attempt := 0; attempt <= eh.retryPolicy.MaxRetries; attempt++ {
        if attempt > 0 {
            eh.logger.Printf("Retry attempt %d after %v", attempt, delay)
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }
        }
        
        err := fn()
        if err == nil {
            if attempt > 0 {
                eh.logger.Printf("Succeeded after %d retries", attempt)
            }
            return nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryable(err) {
            eh.logger.Printf("Non-retryable error: %v", err)
            return err
        }
        
        // Calculate next delay with exponential backoff
        delay = time.Duration(float64(delay) * eh.retryPolicy.BackoffFactor)
        if delay > eh.retryPolicy.MaxDelay {
            delay = eh.retryPolicy.MaxDelay
        }
    }
    
    return fmt.Errorf("failed after %d retries: %w", 
        eh.retryPolicy.MaxRetries, lastErr)
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
    switch {
    case errors.Is(err, ErrConnection):
        return true
    case errors.Is(err, ErrTimeout):
        return true
    case errors.Is(err, ErrRateLimit):
        return true
    case errors.Is(err, ErrServerError):
        return true
    default:
        return false
    }
}

// Circuit Breaker implementation
type CircuitBreaker struct {
    maxFailures      int
    resetTimeout     time.Duration
    failureCount     int
    lastFailureTime  time.Time
    state           CircuitState
}

type CircuitState int

const (
    StateClosed CircuitState = iota
    StateOpen
    StateHalfOpen
)

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:       StateClosed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == StateOpen {
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.state = StateHalfOpen
            cb.failureCount = 0
        } else {
            return errors.New("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failureCount++
        cb.lastFailureTime = time.Now()
        
        if cb.failureCount >= cb.maxFailures {
            cb.state = StateOpen
            return fmt.Errorf("circuit breaker opened: %w", err)
        }
        return err
    }
    
    if cb.state == StateHalfOpen {
        cb.state = StateClosed
    }
    cb.failureCount = 0
    return nil
}
```

### Hands-On Exercise 2.2.C: Monitoring & Logging

```go
// monitoring.go - Comprehensive monitoring implementation
package ollamaclient

import (
    "fmt"
    "sync"
    "time"
)

// MetricsCollector collects and reports metrics
type MetricsCollector struct {
    mu          sync.RWMutex
    requests    map[string]*RequestMetrics
    startTime   time.Time
}

type RequestMetrics struct {
    Count       int64
    TotalTime   time.Duration
    MinTime     time.Duration
    MaxTime     time.Duration
    Errors      int64
    LastError   error
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        requests:  make(map[string]*RequestMetrics),
        startTime: time.Now(),
    }
}

func (mc *MetricsCollector) RecordRequest(endpoint string, duration time.Duration, err error) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    
    metrics, exists := mc.requests[endpoint]
    if !exists {
        metrics = &RequestMetrics{
            MinTime: duration,
            MaxTime: duration,
        }
        mc.requests[endpoint] = metrics
    }
    
    metrics.Count++
    metrics.TotalTime += duration
    
    if duration < metrics.MinTime {
        metrics.MinTime = duration
    }
    if duration > metrics.MaxTime {
        metrics.MaxTime = duration
    }
    
    if err != nil {
        metrics.Errors++
        metrics.LastError = err
    }
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    
    result := make(map[string]interface{})
    result["uptime"] = time.Since(mc.startTime).String()
    
    endpoints := make(map[string]interface{})
    for endpoint, metrics := range mc.requests {
        avgTime := time.Duration(0)
        if metrics.Count > 0 {
            avgTime = metrics.TotalTime / time.Duration(metrics.Count)
        }
        
        endpoints[endpoint] = map[string]interface{}{
            "count":      metrics.Count,
            "errors":     metrics.Errors,
            "avg_time":   avgTime.String(),
            "min_time":   metrics.MinTime.String(),
            "max_time":   metrics.MaxTime.String(),
            "error_rate": fmt.Sprintf("%.2f%%", 
                float64(metrics.Errors)/float64(metrics.Count)*100),
        }
    }
    result["endpoints"] = endpoints
    
    return result
}

// Structured logging
type Logger struct {
    level LogLevel
    mu    sync.Mutex
}

type LogLevel int

const (
    LogDebug LogLevel = iota
    LogInfo
    LogWarn
    LogError
)

func (l *Logger) Log(level LogLevel, format string, args ...interface{}) {
    if level < l.level {
        return
    }
    
    l.mu.Lock()
    defer l.mu.Unlock()
    
    timestamp := time.Now().Format(time.RFC3339)
    levelStr := []string{"DEBUG", "INFO", "WARN", "ERROR"}[level]
    
    message := fmt.Sprintf(format, args...)
    fmt.Printf("[%s] %s: %s\n", timestamp, levelStr, message)
}
```

### Knowledge Check 2.2
1. **What pattern is used for client configuration?** Functional options
2. **How many retries are attempted by default?** 3 retries
3. **What determines if an error is retryable?** Error type (connection, timeout, rate limit, server)
4. **What is a circuit breaker?** Pattern to prevent cascading failures
5. **How are metrics collected?** Thread-safe map with mutex protection

**✅ Module 2.2 Complete**: API client development mastered

---

## 📚 Module 2.3: Integration Patterns (30 minutes)

### Learning Objectives
- Implement common integration patterns
- Build middleware and proxy components
- Create authentication and authorization layers
- Master performance optimization

### Hands-On Exercise 2.3.A: Middleware Development

```go
// middleware.go - HTTP middleware implementation
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/google/uuid"
)

// Middleware type for chaining
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middleware
func Chain(middlewares ...Middleware) Middleware {
    return func(next http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            next = middlewares[i](next)
        }
        return next
    }
}

// RequestID adds a unique request ID
func RequestID() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = uuid.New().String()
            }
            
            ctx := context.WithValue(r.Context(), "requestID", requestID)
            w.Header().Set("X-Request-ID", requestID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Logging middleware for request/response logging
func Logging(logger *Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            wrapped := &responseWriter{
                ResponseWriter: w,
                statusCode:    http.StatusOK,
            }
            
            next.ServeHTTP(wrapped, r)
            
            duration := time.Since(start)
            logger.Info(
                "Request completed",
                "method", r.Method,
                "path", r.URL.Path,
                "status", wrapped.statusCode,
                "duration", duration,
                "request_id", r.Context().Value("requestID"),
            )
        })
    }
}

// RateLimit implements rate limiting
func RateLimit(requests int, window time.Duration) Middleware {
    limiter := NewRateLimiter(requests, window)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)
            
            if !limiter.Allow(clientIP) {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Authentication middleware
func Authentication(validator TokenValidator) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "Missing authentication", http.StatusUnauthorized)
                return
            }
            
            claims, err := validator.Validate(token)
            if err != nil {
                http.Error(w, "Invalid authentication", http.StatusUnauthorized)
                return
            }
            
            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// CORS middleware for cross-origin requests
func CORS(origins []string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            for _, allowed := range origins {
                if allowed == "*" || allowed == origin {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    w.Header().Set("Access-Control-Allow-Methods", 
                        "GET, POST, PUT, DELETE, OPTIONS")
                    w.Header().Set("Access-Control-Allow-Headers", 
                        "Content-Type, Authorization")
                    break
                }
            }
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Compression middleware
func Compression() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !shouldCompress(r) {
                next.ServeHTTP(w, r)
                return
            }
            
            gzw := newGzipResponseWriter(w)
            defer gzw.Close()
            
            w.Header().Set("Content-Encoding", "gzip")
            next.ServeHTTP(gzw, r)
        })
    }
}

// Timeout middleware
func Timeout(duration time.Duration) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), duration)
            defer cancel()
            
            done := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()
            
            select {
            case <-done:
                return
            case <-ctx.Done():
                http.Error(w, "Request timeout", http.StatusRequestTimeout)
            }
        })
    }
}

// Recovery middleware for panic recovery
func Recovery(logger *Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logger.Error("Panic recovered", 
                        "error", err,
                        "path", r.URL.Path,
                        "method", r.Method,
                    )
                    http.Error(w, "Internal server error", 
                        http.StatusInternalServerError)
                }
            }()
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### Hands-On Exercise 2.3.B: Proxy Pattern Implementation

```go
// proxy.go - Reverse proxy implementation
package proxy

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"
)

// LoadBalancer distributes requests across backend servers
type LoadBalancer struct {
    backends    []*Backend
    current     int
    mu          sync.RWMutex
    healthCheck time.Duration
}

type Backend struct {
    URL       *url.URL
    Alive     bool
    Weight    int
    mu        sync.RWMutex
}

func NewLoadBalancer(urls []string, healthCheck time.Duration) (*LoadBalancer, error) {
    lb := &LoadBalancer{
        backends:    make([]*Backend, 0, len(urls)),
        healthCheck: healthCheck,
    }
    
    for _, u := range urls {
        parsed, err := url.Parse(u)
        if err != nil {
            return nil, fmt.Errorf("invalid URL %s: %w", u, err)
        }
        
        lb.backends = append(lb.backends, &Backend{
            URL:    parsed,
            Alive:  true,
            Weight: 1,
        })
    }
    
    // Start health checking
    go lb.healthCheckLoop()
    
    return lb, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    backend := lb.nextBackend()
    if backend == nil {
        http.Error(w, "No available backends", http.StatusServiceUnavailable)
        return
    }
    
    proxy := &ReverseProxy{
        Director: func(req *http.Request) {
            req.URL.Scheme = backend.URL.Scheme
            req.URL.Host = backend.URL.Host
            req.URL.Path = singleJoiningSlash(backend.URL.Path, req.URL.Path)
            req.Host = backend.URL.Host
            
            // Add forwarding headers
            if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
                req.Header.Set("X-Forwarded-For", clientIP)
            }
            req.Header.Set("X-Forwarded-Proto", "http")
            req.Header.Set("X-Forwarded-Host", r.Host)
        },
        ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
            backend.mu.Lock()
            backend.Alive = false
            backend.mu.Unlock()
            
            http.Error(w, "Backend error", http.StatusBadGateway)
        },
    }
    
    proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) nextBackend() *Backend {
    lb.mu.Lock()
    defer lb.mu.Unlock()
    
    // Round-robin with health checking
    start := lb.current
    for {
        lb.current = (lb.current + 1) % len(lb.backends)
        
        backend := lb.backends[lb.current]
        backend.mu.RLock()
        alive := backend.Alive
        backend.mu.RUnlock()
        
        if alive {
            return backend
        }
        
        if lb.current == start {
            return nil // No healthy backends
        }
    }
}

func (lb *LoadBalancer) healthCheckLoop() {
    ticker := time.NewTicker(lb.healthCheck)
    defer ticker.Stop()
    
    for range ticker.C {
        for _, backend := range lb.backends {
            go lb.checkHealth(backend)
        }
    }
}

func (lb *LoadBalancer) checkHealth(backend *Backend) {
    timeout := 5 * time.Second
    client := &http.Client{Timeout: timeout}
    
    resp, err := client.Get(backend.URL.String() + "/health")
    
    backend.mu.Lock()
    defer backend.mu.Unlock()
    
    if err != nil || resp.StatusCode != http.StatusOK {
        backend.Alive = false
    } else {
        backend.Alive = true
    }
}

// ReverseProxy implementation
type ReverseProxy struct {
    Director     func(*http.Request)
    Transport    http.RoundTripper
    ErrorHandler func(http.ResponseWriter, *http.Request, error)
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    transport := p.Transport
    if transport == nil {
        transport = http.DefaultTransport
    }
    
    outreq := r.Clone(r.Context())
    if p.Director != nil {
        p.Director(outreq)
    }
    
    res, err := transport.RoundTrip(outreq)
    if err != nil {
        if p.ErrorHandler != nil {
            p.ErrorHandler(w, r, err)
        } else {
            http.Error(w, err.Error(), http.StatusBadGateway)
        }
        return
    }
    defer res.Body.Close()
    
    // Copy headers
    for key, values := range res.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    
    w.WriteHeader(res.StatusCode)
    io.Copy(w, res.Body)
}

func singleJoiningSlash(a, b string) string {
    aslash := strings.HasSuffix(a, "/")
    bslash := strings.HasPrefix(b, "/")
    switch {
    case aslash && bslash:
        return a + b[1:]
    case !aslash && !bslash:
        return a + "/" + b
    }
    return a + b
}
```

### Knowledge Check 2.3
1. **What is middleware chaining?** Combining multiple middleware functions in sequence
2. **How does the circuit breaker pattern work?** Opens after failures, prevents cascading
3. **What is round-robin load balancing?** Distributing requests evenly across backends
4. **What headers are added by a reverse proxy?** X-Forwarded-For, X-Forwarded-Proto, X-Forwarded-Host
5. **How are health checks performed?** Periodic HTTP requests to /health endpoint

**✅ Module 2.3 Complete**: Integration patterns implemented

---

## 📚 Module 2.4: Testing & Quality (25 minutes)

### Learning Objectives
- Develop comprehensive test suites
- Implement automated validation
- Create performance benchmarks
- Master quality assurance processes

### Hands-On Exercise 2.4.A: Unit Testing

```go
// client_test.go - Comprehensive unit tests
package ollamaclient

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
    tests := []struct {
        name    string
        baseURL string
        options []ClientOption
        want    *Client
    }{
        {
            name:    "default client",
            baseURL: "http://localhost:8080",
            options: nil,
            want: &Client{
                baseURL: "http://localhost:8080",
                httpClient: &http.Client{
                    Timeout: 30 * time.Second,
                },
            },
        },
        {
            name:    "client with options",
            baseURL: "http://localhost:8080",
            options: []ClientOption{
                WithTimeout(10 * time.Second),
                WithAPIKey("test-key"),
                WithDebug(true),
            },
            want: &Client{
                baseURL: "http://localhost:8080",
                apiKey:  "test-key",
                debug:   true,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := NewClient(tt.baseURL, tt.options...)
            
            assert.Equal(t, tt.want.baseURL, client.baseURL)
            assert.Equal(t, tt.want.apiKey, client.apiKey)
            assert.Equal(t, tt.want.debug, client.debug)
        })
    }
}

func TestClient_Health(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/health", r.URL.Path)
        assert.Equal(t, "GET", r.Method)
        
        response := HealthResponse{
            Status:    "healthy",
            Timestamp: time.Now(),
            Version:   "1.0.0",
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    // Test client
    client := NewClient(server.URL)
    ctx := context.Background()
    
    health, err := client.Health(ctx)
    
    require.NoError(t, err)
    assert.Equal(t, "healthy", health.Status)
    assert.Equal(t, "1.0.0", health.Version)
}

func TestClient_HealthError(t *testing.T) {
    // Create test server that returns error
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()
    
    client := NewClient(server.URL)
    ctx := context.Background()
    
    _, err := client.Health(ctx)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "health check failed")
}

func TestClient_WithRetry(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        
        response := HealthResponse{
            Status:    "healthy",
            Timestamp: time.Now(),
            Version:   "1.0.0",
        }
        
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    client := NewClient(server.URL)
    ctx := context.Background()
    
    health, err := client.Health(ctx)
    
    require.NoError(t, err)
    assert.Equal(t, "healthy", health.Status)
    assert.Equal(t, 3, attempts)
}

// Benchmark tests
func BenchmarkClient_Health(b *testing.B) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        response := HealthResponse{
            Status:    "healthy",
            Timestamp: time.Now(),
            Version:   "1.0.0",
        }
        json.NewEncoder(w).Encode(response)
    }))
    defer server.Close()
    
    client := NewClient(server.URL)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = client.Health(ctx)
    }
}

// Table-driven tests
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        statusCode  int
        wantError   bool
        errorContains string
    }{
        {
            name:       "success",
            statusCode: http.StatusOK,
            wantError:  false,
        },
        {
            name:       "not found",
            statusCode: http.StatusNotFound,
            wantError:  true,
            errorContains: "not found",
        },
        {
            name:       "unauthorized",
            statusCode: http.StatusUnauthorized,
            wantError:  true,
            errorContains: "unauthorized",
        },
        {
            name:       "server error",
            statusCode: http.StatusInternalServerError,
            wantError:  true,
            errorContains: "server error",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(tt.statusCode)
            }))
            defer server.Close()
            
            client := NewClient(server.URL)
            ctx := context.Background()
            
            _, err := client.Health(ctx)
            
            if tt.wantError {
                assert.Error(t, err)
                if tt.errorContains != "" {
                    assert.Contains(t, err.Error(), tt.errorContains)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Hands-On Exercise 2.4.B: Integration Testing

```go
// integration_test.go - Integration test suite
// +build integration

package integration

import (
    "context"
    "os"
    "testing"
    "time"
    
    "github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
    suite.Suite
    client  *Client
    baseURL string
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Get test environment configuration
    suite.baseURL = os.Getenv("TEST_BASE_URL")
    if suite.baseURL == "" {
        suite.baseURL = "http://localhost:8080"
    }
    
    // Create client with test configuration
    suite.client = NewClient(
        suite.baseURL,
        WithTimeout(30 * time.Second),
        WithDebug(true),
    )
    
    // Wait for service to be ready
    suite.waitForService()
}

func (suite *IntegrationTestSuite) waitForService() {
    ctx := context.Background()
    maxRetries := 30
    
    for i := 0; i < maxRetries; i++ {
        health, err := suite.client.Health(ctx)
        if err == nil && health.Status == "healthy" {
            return
        }
        time.Sleep(1 * time.Second)
    }
    
    suite.T().Fatal("Service not ready after 30 seconds")
}

func (suite *IntegrationTestSuite) TestHealthEndpoint() {
    ctx := context.Background()
    
    health, err := suite.client.Health(ctx)
    
    suite.NoError(err)
    suite.Equal("healthy", health.Status)
    suite.NotEmpty(health.Version)
}

func (suite *IntegrationTestSuite) TestStatusEndpoint() {
    ctx := context.Background()
    
    status, err := suite.client.Status(ctx)
    
    suite.NoError(err)
    suite.NotEmpty(status.NodeID)
    suite.Equal("running", status.State)
}

func (suite *IntegrationTestSuite) TestMetricsEndpoint() {
    ctx := context.Background()
    
    metrics, err := suite.client.Metrics(ctx)
    
    suite.NoError(err)
    suite.GreaterOrEqual(metrics.CPU, 0.0)
    suite.GreaterOrEqual(metrics.Memory, 0.0)
}

func (suite *IntegrationTestSuite) TestConfigUpdate() {
    ctx := context.Background()
    
    config := &NodeConfig{
        Name:      "test-node",
        APIPort:   8080,
        WebPort:   8081,
        MaxMemory: 2048,
        EnableGPU: false,
    }
    
    err := suite.client.UpdateConfig(ctx, config)
    
    suite.NoError(err)
}

func (suite *IntegrationTestSuite) TestConcurrentRequests() {
    ctx := context.Background()
    concurrency := 10
    errors := make(chan error, concurrency)
    
    for i := 0; i < concurrency; i++ {
        go func() {
            _, err := suite.client.Health(ctx)
            errors <- err
        }()
    }
    
    for i := 0; i < concurrency; i++ {
        err := <-errors
        suite.NoError(err)
    }
}

func TestIntegrationSuite(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }
    
    suite.Run(t, new(IntegrationTestSuite))
}
```

### Hands-On Exercise 2.4.C: Performance Benchmarks

```go
// benchmark_test.go - Performance benchmarking
package benchmark

import (
    "context"
    "fmt"
    "testing"
    "time"
)

func BenchmarkAPICall(b *testing.B) {
    client := setupBenchmarkClient()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = client.Health(ctx)
    }
}

func BenchmarkConcurrentAPICalls(b *testing.B) {
    concurrencyLevels := []int{1, 10, 50, 100}
    
    for _, concurrency := range concurrencyLevels {
        b.Run(fmt.Sprintf("concurrency-%d", concurrency), func(b *testing.B) {
            client := setupBenchmarkClient()
            ctx := context.Background()
            
            b.SetParallelism(concurrency)
            b.ResetTimer()
            
            b.RunParallel(func(pb *testing.PB) {
                for pb.Next() {
                    _, _ = client.Health(ctx)
                }
            })
        })
    }
}

func BenchmarkJSONSerialization(b *testing.B) {
    data := generateTestData()
    
    b.Run("marshal", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, _ = json.Marshal(data)
        }
    })
    
    b.Run("unmarshal", func(b *testing.B) {
        bytes, _ := json.Marshal(data)
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
            var result interface{}
            _ = json.Unmarshal(bytes, &result)
        }
    })
}

func BenchmarkMemoryAllocation(b *testing.B) {
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _ = make([]byte, 1024)
    }
}

// Load testing
func TestLoadScenario(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    duration := 1 * time.Minute
    rps := 100 // requests per second
    
    client := setupBenchmarkClient()
    ctx := context.Background()
    
    ticker := time.NewTicker(time.Second / time.Duration(rps))
    defer ticker.Stop()
    
    timeout := time.After(duration)
    
    var successCount, errorCount int
    
    for {
        select {
        case <-ticker.C:
            go func() {
                _, err := client.Health(ctx)
                if err != nil {
                    errorCount++
                } else {
                    successCount++
                }
            }()
        case <-timeout:
            t.Logf("Load test complete: %d successful, %d errors", 
                successCount, errorCount)
            
            errorRate := float64(errorCount) / float64(successCount+errorCount) * 100
            if errorRate > 1.0 {
                t.Errorf("Error rate too high: %.2f%%", errorRate)
            }
            return
        }
    }
}
```

### Knowledge Check 2.4
1. **What testing framework is commonly used in Go?** testify (assert/require/suite)
2. **How are integration tests tagged?** Build tags like `// +build integration`
3. **What does b.ResetTimer() do?** Excludes setup time from benchmark
4. **How to run parallel benchmarks?** b.RunParallel with b.SetParallelism
5. **What metrics are important in load testing?** Success rate, error rate, latency

**✅ Module 2.4 Complete**: Testing and quality assurance mastered

---

## 📚 Module 2.5: Advanced Development (15 minutes)

### Learning Objectives
- Contribute to open source project
- Implement custom extensions
- Understand advanced architectural patterns
- Master community engagement

### Hands-On Exercise 2.5: Contributing to Open Source

```markdown
# Contributing to Ollama Distributed

## Step 1: Fork and Clone
1. Fork the repository on GitHub
2. Clone your fork locally
3. Add upstream remote

## Step 2: Create Feature Branch
```bash
git checkout -b feature/your-feature-name
```

## Step 3: Make Changes
- Follow code style guidelines
- Add tests for new features
- Update documentation

## Step 4: Run Tests
```bash
make test
make lint
make coverage
```

## Step 5: Commit Changes
```bash
git add .
git commit -m "feat: add new feature description"
```

## Step 6: Push and Create PR
```bash
git push origin feature/your-feature-name
```

## Pull Request Template
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Knowledge Check 2.5
1. **What is a fork?** Personal copy of repository
2. **What is a pull request?** Request to merge changes
3. **What are commit conventions?** Standardized commit message format
4. **What is upstream?** Original repository
5. **What is continuous integration?** Automated testing on commits

**✅ Module 2.5 Complete**: Open source contribution ready

---

## 📝 CERTIFICATION ASSESSMENT

### Part A: Knowledge Assessment (30% - 30 minutes)

#### Technical Concepts (15 questions, 2 points each)

1. **What pattern is used for client configuration in Go?**
   - a) Builder pattern
   - b) Factory pattern
   - c) Functional options ✓
   - d) Singleton pattern

2. **Which HTTP status code indicates rate limiting?**
   - a) 403
   - b) 429 ✓
   - c) 503
   - d) 504

3. **What is the purpose of context.Context in Go?**
   - a) Memory management
   - b) Request cancellation and deadlines ✓
   - c) Error handling
   - d) Logging

4. **How do you make a Go test file for integration tests only?**
   - a) Name it integration_test.go
   - b) Use build tags ✓
   - c) Put it in integration folder
   - d) Use TestIntegration prefix

5. **What does defer do in Go?**
   - a) Delays execution
   - b) Executes at function exit ✓
   - c) Creates goroutine
   - d) Handles errors

[Questions 6-15 continue with similar technical depth...]

### Part B: Practical Assessment (70% - 90 minutes)

#### Task 1: API Client Implementation (25 points)

Create a fully functional API client that:
- Implements all current endpoints
- Includes proper error handling
- Has retry logic with exponential backoff
- Supports context cancellation
- Includes comprehensive logging

#### Task 2: Integration Component (25 points)

Build a middleware component that:
- Implements rate limiting
- Adds request ID tracking
- Provides authentication
- Includes metrics collection
- Has proper error recovery

#### Task 3: Test Suite Creation (20 points)

Develop a test suite with:
- Unit tests with >80% coverage
- Integration tests for API
- Benchmark tests
- Table-driven test cases
- Mock implementations

### Scoring Rubric

```yaml
knowledge_assessment:
  total: 30 points
  passing: 75% (23 points)
  
practical_assessment:
  total: 70 points
  passing: 75% (53 points)
  breakdown:
    - code_quality: 20%
    - functionality: 30%
    - testing: 20%
    - documentation: 15%
    - best_practices: 15%

overall:
  passing_score: 80/100
  certification_levels:
    distinction: 90-100
    merit: 85-89
    pass: 80-84
```

---

## 🎓 CERTIFICATE TEMPLATE

```
════════════════════════════════════════════════════════════════
              OLLAMA DISTRIBUTED CERTIFICATION
                    DEVELOPER TRACK
                  PROFESSIONAL LEVEL

  This certifies that [NAME] has demonstrated professional
  proficiency in developing with Ollama Distributed

  Certification ID: DEV-[UUID]
  Issue Date: [DATE]
  Valid Until: [DATE+24M]
  
  Score: [SCORE]/100
  Level: [DISTINCTION/MERIT/PASS]
  
  Competencies Validated:
  ✓ API Client Development
  ✓ Integration Patterns
  ✓ Testing & Quality
  ✓ Open Source Contribution
  
  _____________________        _____________________
  Training Director             Technical Director
════════════════════════════════════════════════════════════════
```

---

## ✅ Developer Track Complete

You have completed the Developer Track (Professional Level) certification materials. You are now equipped to:
- Build production-ready API clients
- Implement complex integration patterns
- Create comprehensive test suites
- Contribute to the open source project
- Progress to Administrator or Architect tracks

**Next Steps**:
1. Complete the practical assessment
2. Submit code for review
3. Receive professional certification
4. Consider advancing to Expert/Master levels
5. Contribute to the project

Good luck with your certification!