package network

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Optimizer provides network performance optimizations
type Optimizer struct {
	config *Config

	// Compression
	compressor *Compressor

	// Keep-alive management
	keepAlive *KeepAliveManager

	// Protocol optimization
	protocol *ProtocolOptimizer

	// Statistics
	stats *Stats
	mu    sync.RWMutex
}

// Config holds network optimization configuration
type Config struct {
	// Compression settings
	EnableCompression    bool `yaml:"enable_compression"`
	CompressionLevel     int  `yaml:"compression_level"`
	CompressionThreshold int  `yaml:"compression_threshold"`

	// Keep-alive settings
	EnableKeepAlive      bool          `yaml:"enable_keep_alive"`
	KeepAliveTimeout     time.Duration `yaml:"keep_alive_timeout"`
	KeepAliveInterval    time.Duration `yaml:"keep_alive_interval"`
	MaxKeepAliveRequests int           `yaml:"max_keep_alive_requests"`

	// Protocol optimization
	EnableHTTP2      bool          `yaml:"enable_http2"`
	EnableTCPNoDelay bool          `yaml:"enable_tcp_no_delay"`
	TCPKeepAlive     time.Duration `yaml:"tcp_keep_alive"`

	// Buffer settings
	ReadBufferSize  int `yaml:"read_buffer_size"`
	WriteBufferSize int `yaml:"write_buffer_size"`

	// Timeout settings
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
}

// DefaultConfig returns default network optimization configuration
func DefaultConfig() *Config {
	return &Config{
		EnableCompression:    true,
		CompressionLevel:     6,    // Balanced compression
		CompressionThreshold: 1024, // 1KB
		EnableKeepAlive:      true,
		KeepAliveTimeout:     30 * time.Second,
		KeepAliveInterval:    15 * time.Second,
		MaxKeepAliveRequests: 100,
		EnableHTTP2:          true,
		EnableTCPNoDelay:     true,
		TCPKeepAlive:         30 * time.Second,
		ReadBufferSize:       32 * 1024, // 32KB
		WriteBufferSize:      32 * 1024, // 32KB
		ConnectTimeout:       10 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
	}
}

// Stats holds network optimization statistics
type Stats struct {
	// Compression statistics
	BytesCompressed   int64         `json:"bytes_compressed"`
	BytesUncompressed int64         `json:"bytes_uncompressed"`
	CompressionRatio  float64       `json:"compression_ratio"`
	CompressionTime   time.Duration `json:"compression_time"`

	// Keep-alive statistics
	KeepAliveConnections int64 `json:"keep_alive_connections"`
	KeepAliveReuses      int64 `json:"keep_alive_reuses"`
	KeepAliveTimeouts    int64 `json:"keep_alive_timeouts"`

	// Protocol statistics
	HTTP2Connections int64 `json:"http2_connections"`
	HTTP1Connections int64 `json:"http1_connections"`
	TCPConnections   int64 `json:"tcp_connections"`

	// Performance metrics
	AverageLatency    time.Duration `json:"average_latency"`
	AverageThroughput float64       `json:"average_throughput"` // bytes/sec

	// Error statistics
	CompressionErrors int64 `json:"compression_errors"`
	ProtocolErrors    int64 `json:"protocol_errors"`
	TimeoutErrors     int64 `json:"timeout_errors"`

	// Timestamps
	LastActivity time.Time `json:"last_activity"`
	StartTime    time.Time `json:"start_time"`
}

// NewOptimizer creates a new network optimizer
func NewOptimizer(config *Config) *Optimizer {
	if config == nil {
		config = DefaultConfig()
	}

	optimizer := &Optimizer{
		config:     config,
		compressor: NewCompressor(config),
		keepAlive:  NewKeepAliveManager(config),
		protocol:   NewProtocolOptimizer(config),
		stats:      &Stats{StartTime: time.Now()},
	}

	return optimizer
}

// OptimizeHTTPTransport optimizes an HTTP transport
func (o *Optimizer) OptimizeHTTPTransport(transport *http.Transport) {
	// Configure keep-alive
	if o.config.EnableKeepAlive {
		transport.DisableKeepAlives = false
		transport.MaxIdleConns = 100
		transport.MaxIdleConnsPerHost = 10
		transport.IdleConnTimeout = o.config.KeepAliveTimeout
	} else {
		transport.DisableKeepAlives = true
	}

	// Configure timeouts
	transport.DialContext = (&net.Dialer{
		Timeout:   o.config.ConnectTimeout,
		KeepAlive: o.config.TCPKeepAlive,
	}).DialContext

	transport.ResponseHeaderTimeout = o.config.ReadTimeout
	transport.ExpectContinueTimeout = 1 * time.Second

	// Configure HTTP/2
	if !o.config.EnableHTTP2 {
		transport.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	}

	// Configure compression
	transport.DisableCompression = !o.config.EnableCompression
}

// OptimizeTCPConnection optimizes a TCP connection
func (o *Optimizer) OptimizeTCPConnection(conn net.Conn) error {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Enable TCP_NODELAY for low latency
		if o.config.EnableTCPNoDelay {
			if err := tcpConn.SetNoDelay(true); err != nil {
				return fmt.Errorf("failed to set TCP_NODELAY: %w", err)
			}
		}

		// Configure keep-alive
		if o.config.EnableKeepAlive {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				return fmt.Errorf("failed to enable keep-alive: %w", err)
			}

			if err := tcpConn.SetKeepAlivePeriod(o.config.TCPKeepAlive); err != nil {
				return fmt.Errorf("failed to set keep-alive period: %w", err)
			}
		}

		o.updateStats(func(s *Stats) {
			s.TCPConnections++
			s.LastActivity = time.Now()
		})
	}

	return nil
}

// CompressData compresses data if it meets the threshold
func (o *Optimizer) CompressData(data []byte) ([]byte, bool, error) {
	if !o.config.EnableCompression || len(data) < o.config.CompressionThreshold {
		return data, false, nil
	}

	start := time.Now()
	compressed, err := o.compressor.Compress(data)
	compressionTime := time.Since(start)

	if err != nil {
		o.updateStats(func(s *Stats) {
			s.CompressionErrors++
		})
		return data, false, err
	}

	// Only use compression if it actually reduces size
	if len(compressed) >= len(data) {
		return data, false, nil
	}

	o.updateStats(func(s *Stats) {
		s.BytesCompressed += int64(len(compressed))
		s.BytesUncompressed += int64(len(data))
		s.CompressionTime = compressionTime
		s.CompressionRatio = float64(len(compressed)) / float64(len(data))
		s.LastActivity = time.Now()
	})

	return compressed, true, nil
}

// DecompressData decompresses data
func (o *Optimizer) DecompressData(data []byte) ([]byte, error) {
	start := time.Now()
	decompressed, err := o.compressor.Decompress(data)

	if err != nil {
		o.updateStats(func(s *Stats) {
			s.CompressionErrors++
		})
		return nil, err
	}

	o.updateStats(func(s *Stats) {
		s.CompressionTime = time.Since(start)
		s.LastActivity = time.Now()
	})

	return decompressed, nil
}

// Stats returns current optimization statistics
func (o *Optimizer) Stats() Stats {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return *o.stats
}

// updateStats safely updates optimization statistics
func (o *Optimizer) updateStats(fn func(*Stats)) {
	o.mu.Lock()
	defer o.mu.Unlock()

	fn(o.stats)
}

// Compressor handles data compression
type Compressor struct {
	level int
}

// NewCompressor creates a new compressor
func NewCompressor(config *Config) *Compressor {
	return &Compressor{
		level: config.CompressionLevel,
	}
}

// Compress compresses data using gzip
func (c *Compressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	// Set compression level
	writer.Header.Comment = "ollama-distributed"

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress decompresses gzip data
func (c *Compressor) Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// KeepAliveManager manages connection keep-alive
type KeepAliveManager struct {
	config      *Config
	connections map[string]*KeepAliveConnection
	mu          sync.RWMutex
}

// KeepAliveConnection tracks keep-alive connection state
type KeepAliveConnection struct {
	ID           string
	LastUsed     time.Time
	RequestCount int
	MaxRequests  int
}

// NewKeepAliveManager creates a new keep-alive manager
func NewKeepAliveManager(config *Config) *KeepAliveManager {
	return &KeepAliveManager{
		config:      config,
		connections: make(map[string]*KeepAliveConnection),
	}
}

// RegisterConnection registers a new keep-alive connection
func (kam *KeepAliveManager) RegisterConnection(id string) {
	kam.mu.Lock()
	defer kam.mu.Unlock()

	kam.connections[id] = &KeepAliveConnection{
		ID:           id,
		LastUsed:     time.Now(),
		RequestCount: 0,
		MaxRequests:  kam.config.MaxKeepAliveRequests,
	}
}

// UseConnection marks a connection as used
func (kam *KeepAliveManager) UseConnection(id string) bool {
	kam.mu.Lock()
	defer kam.mu.Unlock()

	conn, exists := kam.connections[id]
	if !exists {
		return false
	}

	// Check if connection has expired
	if time.Since(conn.LastUsed) > kam.config.KeepAliveTimeout {
		delete(kam.connections, id)
		return false
	}

	// Check if connection has reached max requests
	if conn.RequestCount >= conn.MaxRequests {
		delete(kam.connections, id)
		return false
	}

	conn.LastUsed = time.Now()
	conn.RequestCount++

	return true
}

// CleanupConnections removes expired connections
func (kam *KeepAliveManager) CleanupConnections() {
	kam.mu.Lock()
	defer kam.mu.Unlock()

	now := time.Now()
	for id, conn := range kam.connections {
		if now.Sub(conn.LastUsed) > kam.config.KeepAliveTimeout {
			delete(kam.connections, id)
		}
	}
}

// ProtocolOptimizer handles protocol-specific optimizations
type ProtocolOptimizer struct {
	config *Config
}

// NewProtocolOptimizer creates a new protocol optimizer
func NewProtocolOptimizer(config *Config) *ProtocolOptimizer {
	return &ProtocolOptimizer{
		config: config,
	}
}

// OptimizeHTTPRequest optimizes an HTTP request
func (po *ProtocolOptimizer) OptimizeHTTPRequest(req *http.Request) {
	// Set appropriate headers
	if po.config.EnableCompression {
		req.Header.Set("Accept-Encoding", "gzip")
	}

	if po.config.EnableKeepAlive {
		req.Header.Set("Connection", "keep-alive")
	}

	// Set user agent
	req.Header.Set("User-Agent", "ollama-distributed/1.0")
}

// OptimizeHTTPResponse optimizes an HTTP response
func (po *ProtocolOptimizer) OptimizeHTTPResponse(w http.ResponseWriter, req *http.Request) {
	// Set keep-alive headers
	if po.config.EnableKeepAlive {
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Keep-Alive", fmt.Sprintf("timeout=%d, max=%d",
			int(po.config.KeepAliveTimeout.Seconds()),
			po.config.MaxKeepAliveRequests))
	}

	// Set compression headers if client supports it
	if po.config.EnableCompression && req.Header.Get("Accept-Encoding") != "" {
		w.Header().Set("Content-Encoding", "gzip")
	}
}
