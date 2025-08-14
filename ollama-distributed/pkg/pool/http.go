package pool

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

// HTTPConnectionPool manages HTTP client connections with pooling
type HTTPConnectionPool struct {
	config    *HTTPConfig
	transport *http.Transport
	clients   map[string]*http.Client
	mu        sync.RWMutex
	stats     *HTTPStats
}

// HTTPConfig holds HTTP connection pool configuration
type HTTPConfig struct {
	// Connection pooling
	MaxIdleConns        int `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost     int `yaml:"max_conns_per_host"`

	// Timeouts
	DialTimeout           time.Duration `yaml:"dial_timeout"`
	KeepAlive             time.Duration `yaml:"keep_alive"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"`

	// HTTP/2 settings
	DisableHTTP2      bool `yaml:"disable_http2"`
	ForceAttemptHTTP2 bool `yaml:"force_attempt_http2"`

	// Compression
	DisableCompression bool `yaml:"disable_compression"`

	// Keep-alive
	DisableKeepAlives bool `yaml:"disable_keep_alives"`
}

// DefaultHTTPConfig returns default HTTP connection pool configuration
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       50,
		DialTimeout:           10 * time.Second,
		KeepAlive:             30 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableHTTP2:          false,
		ForceAttemptHTTP2:     true,
		DisableCompression:    false,
		DisableKeepAlives:     false,
	}
}

// HTTPStats holds HTTP connection pool statistics
type HTTPStats struct {
	// Request statistics
	TotalRequests      int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests     int64 `json:"failed_requests"`

	// Connection statistics
	ConnectionsCreated int64 `json:"connections_created"`
	ConnectionsReused  int64 `json:"connections_reused"`
	ConnectionsClosed  int64 `json:"connections_closed"`

	// Performance metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	AverageConnectTime  time.Duration `json:"average_connect_time"`

	// Error statistics
	TimeoutErrors    int64 `json:"timeout_errors"`
	ConnectionErrors int64 `json:"connection_errors"`
	DNSErrors        int64 `json:"dns_errors"`

	// Timestamps
	LastActivity time.Time `json:"last_activity"`
	StartTime    time.Time `json:"start_time"`
}

// NewHTTPConnectionPool creates a new HTTP connection pool
func NewHTTPConnectionPool(config *HTTPConfig) *HTTPConnectionPool {
	if config == nil {
		config = DefaultHTTPConfig()
	}

	// Create custom dialer
	dialer := &net.Dialer{
		Timeout:   config.DialTimeout,
		KeepAlive: config.KeepAlive,
	}

	// Create custom transport
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		DisableCompression:    config.DisableCompression,
		DisableKeepAlives:     config.DisableKeepAlives,
		ForceAttemptHTTP2:     config.ForceAttemptHTTP2,
	}

	// Disable HTTP/2 if requested
	if config.DisableHTTP2 {
		transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	}

	pool := &HTTPConnectionPool{
		config:    config,
		transport: transport,
		clients:   make(map[string]*http.Client),
		stats:     &HTTPStats{StartTime: time.Now()},
	}

	return pool
}

// GetClient returns an HTTP client for the specified host
func (p *HTTPConnectionPool) GetClient(host string) *http.Client {
	p.mu.RLock()
	client, exists := p.clients[host]
	p.mu.RUnlock()

	if exists {
		return client
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists = p.clients[host]; exists {
		return client
	}

	// Create new client for this host
	client = &http.Client{
		Transport: p.transport,
		Timeout:   30 * time.Second, // Default request timeout
	}

	p.clients[host] = client
	return client
}

// Do performs an HTTP request with connection pooling
func (p *HTTPConnectionPool) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()

	client := p.GetClient(req.URL.Host)

	resp, err := client.Do(req)

	// Update statistics
	p.updateStats(func(s *HTTPStats) {
		s.TotalRequests++
		s.AverageResponseTime = time.Since(start)
		s.LastActivity = time.Now()

		if err != nil {
			s.FailedRequests++
			// Categorize error types
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				s.TimeoutErrors++
			} else {
				s.ConnectionErrors++
			}
		} else {
			s.SuccessfulRequests++
		}
	})

	return resp, err
}

// Stats returns current HTTP pool statistics
func (p *HTTPConnectionPool) Stats() HTTPStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return *p.stats
}

// Close closes all HTTP clients and cleans up resources
func (p *HTTPConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close the transport to clean up connections
	p.transport.CloseIdleConnections()

	// Clear clients
	p.clients = make(map[string]*http.Client)

	return nil
}

// updateStats safely updates HTTP pool statistics
func (p *HTTPConnectionPool) updateStats(fn func(*HTTPStats)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fn(p.stats)
}

// SetTimeout sets the request timeout for all clients
func (p *HTTPConnectionPool) SetTimeout(timeout time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.Timeout = timeout
	}
}

// CloseIdleConnections closes idle connections
func (p *HTTPConnectionPool) CloseIdleConnections() {
	p.transport.CloseIdleConnections()
}

// RoundTripperWithStats wraps a RoundTripper to collect statistics
type RoundTripperWithStats struct {
	base  http.RoundTripper
	stats *HTTPStats
	mu    sync.RWMutex
}

// NewRoundTripperWithStats creates a new RoundTripper with statistics
func NewRoundTripperWithStats(base http.RoundTripper) *RoundTripperWithStats {
	return &RoundTripperWithStats{
		base:  base,
		stats: &HTTPStats{StartTime: time.Now()},
	}
}

// RoundTrip implements the RoundTripper interface
func (rt *RoundTripperWithStats) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := rt.base.RoundTrip(req)

	rt.mu.Lock()
	rt.stats.TotalRequests++
	rt.stats.AverageResponseTime = time.Since(start)
	rt.stats.LastActivity = time.Now()

	if err != nil {
		rt.stats.FailedRequests++
	} else {
		rt.stats.SuccessfulRequests++
	}
	rt.mu.Unlock()

	return resp, err
}

// Stats returns RoundTripper statistics
func (rt *RoundTripperWithStats) Stats() HTTPStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	return *rt.stats
}

// HTTPClientPool manages multiple HTTP clients with different configurations
type HTTPClientPool struct {
	clients map[string]*HTTPConnectionPool
	mu      sync.RWMutex
}

// NewHTTPClientPool creates a new HTTP client pool
func NewHTTPClientPool() *HTTPClientPool {
	return &HTTPClientPool{
		clients: make(map[string]*HTTPConnectionPool),
	}
}

// GetPool returns or creates an HTTP connection pool for the given name
func (hcp *HTTPClientPool) GetPool(name string, config *HTTPConfig) *HTTPConnectionPool {
	hcp.mu.RLock()
	pool, exists := hcp.clients[name]
	hcp.mu.RUnlock()

	if exists {
		return pool
	}

	hcp.mu.Lock()
	defer hcp.mu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists = hcp.clients[name]; exists {
		return pool
	}

	// Create new pool
	pool = NewHTTPConnectionPool(config)
	hcp.clients[name] = pool

	return pool
}

// Close closes all HTTP client pools
func (hcp *HTTPClientPool) Close() error {
	hcp.mu.Lock()
	defer hcp.mu.Unlock()

	for _, pool := range hcp.clients {
		pool.Close()
	}

	hcp.clients = make(map[string]*HTTPConnectionPool)
	return nil
}

// GetAllStats returns statistics for all pools
func (hcp *HTTPClientPool) GetAllStats() map[string]HTTPStats {
	hcp.mu.RLock()
	defer hcp.mu.RUnlock()

	stats := make(map[string]HTTPStats)
	for name, pool := range hcp.clients {
		stats[name] = pool.Stats()
	}

	return stats
}
