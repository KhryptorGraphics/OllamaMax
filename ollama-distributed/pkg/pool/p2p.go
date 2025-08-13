package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// P2PConnectionPool manages P2P connections to peers
type P2PConnectionPool struct {
	config      *P2PConfig
	connections map[string]*P2PConnection
	mu          sync.RWMutex
	stats       *P2PStats

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// P2PConfig holds P2P connection pool configuration
type P2PConfig struct {
	// Pool settings
	MaxConnectionsPerPeer int           `yaml:"max_connections_per_peer"`
	ConnectionTimeout     time.Duration `yaml:"connection_timeout"`
	KeepAliveInterval     time.Duration `yaml:"keep_alive_interval"`

	// Health checking
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout"`

	// Retry settings
	MaxRetries      int           `yaml:"max_retries"`
	RetryBackoff    time.Duration `yaml:"retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`

	// Connection management
	IdleTimeout time.Duration `yaml:"idle_timeout"`
	MaxLifetime time.Duration `yaml:"max_lifetime"`
}

// DefaultP2PConfig returns default P2P connection pool configuration
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		MaxConnectionsPerPeer: 5,
		ConnectionTimeout:     10 * time.Second,
		KeepAliveInterval:     30 * time.Second,
		HealthCheckInterval:   60 * time.Second,
		HealthCheckTimeout:    5 * time.Second,
		MaxRetries:            3,
		RetryBackoff:          1 * time.Second,
		MaxRetryBackoff:       30 * time.Second,
		IdleTimeout:           5 * time.Minute,
		MaxLifetime:           30 * time.Minute,
	}
}

// P2PConnection represents a connection to a peer
type P2PConnection struct {
	PeerID           string
	Address          string
	Connected        bool
	LastUsed         time.Time
	CreatedAt        time.Time
	MessagesSent     int64
	MessagesReceived int64
	BytesSent        int64
	BytesReceived    int64
	Errors           int64

	// Connection state
	mu sync.RWMutex
}

// P2PStats holds P2P connection pool statistics
type P2PStats struct {
	// Connection statistics
	TotalConnections  int   `json:"total_connections"`
	ActiveConnections int   `json:"active_connections"`
	IdleConnections   int   `json:"idle_connections"`
	FailedConnections int64 `json:"failed_connections"`

	// Message statistics
	MessagesSent     int64 `json:"messages_sent"`
	MessagesReceived int64 `json:"messages_received"`
	BytesSent        int64 `json:"bytes_sent"`
	BytesReceived    int64 `json:"bytes_received"`

	// Performance metrics
	AverageLatency     time.Duration `json:"average_latency"`
	AverageConnectTime time.Duration `json:"average_connect_time"`

	// Error statistics
	ConnectionErrors int64 `json:"connection_errors"`
	TimeoutErrors    int64 `json:"timeout_errors"`
	ProtocolErrors   int64 `json:"protocol_errors"`

	// Timestamps
	LastActivity time.Time `json:"last_activity"`
	StartTime    time.Time `json:"start_time"`
}

// NewP2PConnectionPool creates a new P2P connection pool
func NewP2PConnectionPool(config *P2PConfig) *P2PConnectionPool {
	if config == nil {
		config = DefaultP2PConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &P2PConnectionPool{
		config:      config,
		connections: make(map[string]*P2PConnection),
		stats:       &P2PStats{StartTime: time.Now()},
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool
}

// Start initializes the P2P connection pool
func (p *P2PConnectionPool) Start() error {
	// Start health check routine
	p.wg.Add(1)
	go p.runHealthCheck()

	// Start cleanup routine
	p.wg.Add(1)
	go p.runCleanup()

	return nil
}

// Stop shuts down the P2P connection pool
func (p *P2PConnectionPool) Stop() error {
	p.cancel()
	p.wg.Wait()

	// Close all connections
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.connections {
		p.closeConnection(conn)
	}

	p.connections = make(map[string]*P2PConnection)

	return nil
}

// GetConnection returns a connection to the specified peer
func (p *P2PConnectionPool) GetConnection(peerID string) (*P2PConnection, error) {
	p.mu.RLock()
	conn, exists := p.connections[peerID]
	p.mu.RUnlock()

	if exists && conn.Connected {
		conn.mu.Lock()
		conn.LastUsed = time.Now()
		conn.mu.Unlock()

		p.updateStats(func(s *P2PStats) {
			s.LastActivity = time.Now()
		})

		return conn, nil
	}

	// Create new connection
	return p.createConnection(peerID)
}

// SendMessage sends a message to a peer
func (p *P2PConnectionPool) SendMessage(peerID string, message []byte) error {
	conn, err := p.GetConnection(peerID)
	if err != nil {
		return fmt.Errorf("failed to get connection to peer %s: %w", peerID, err)
	}

	// Simulate sending message
	// In a real implementation, this would use the actual P2P protocol

	conn.mu.Lock()
	conn.MessagesSent++
	conn.BytesSent += int64(len(message))
	conn.LastUsed = time.Now()
	conn.mu.Unlock()

	p.updateStats(func(s *P2PStats) {
		s.MessagesSent++
		s.BytesSent += int64(len(message))
		s.LastActivity = time.Now()
	})

	return nil
}

// ReceiveMessage simulates receiving a message from a peer
func (p *P2PConnectionPool) ReceiveMessage(peerID string, message []byte) {
	p.mu.RLock()
	conn, exists := p.connections[peerID]
	p.mu.RUnlock()

	if !exists {
		// Create connection for incoming message
		conn, _ = p.createConnection(peerID)
	}

	if conn != nil {
		conn.mu.Lock()
		conn.MessagesReceived++
		conn.BytesReceived += int64(len(message))
		conn.LastUsed = time.Now()
		conn.mu.Unlock()
	}

	p.updateStats(func(s *P2PStats) {
		s.MessagesReceived++
		s.BytesReceived += int64(len(message))
		s.LastActivity = time.Now()
	})
}

// Stats returns current P2P pool statistics
func (p *P2PConnectionPool) Stats() P2PStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return *p.stats
}

// GetConnectionStats returns statistics for all connections
func (p *P2PConnectionPool) GetConnectionStats() map[string]*P2PConnection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[string]*P2PConnection)
	for peerID, conn := range p.connections {
		// Return a copy to avoid race conditions
		connCopy := *conn
		stats[peerID] = &connCopy
	}

	return stats
}

// createConnection creates a new connection to a peer
func (p *P2PConnectionPool) createConnection(peerID string) (*P2PConnection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if connection was created while waiting for lock
	if conn, exists := p.connections[peerID]; exists && conn.Connected {
		return conn, nil
	}

	start := time.Now()

	// Simulate connection creation
	// In a real implementation, this would establish actual P2P connection
	conn := &P2PConnection{
		PeerID:    peerID,
		Address:   fmt.Sprintf("peer-%s", peerID), // Placeholder address
		Connected: true,
		LastUsed:  time.Now(),
		CreatedAt: time.Now(),
	}

	p.connections[peerID] = conn

	p.stats.TotalConnections++
	p.stats.ActiveConnections++
	p.stats.AverageConnectTime = time.Since(start)
	p.stats.LastActivity = time.Now()

	return conn, nil
}

// closeConnection closes a connection to a peer
func (p *P2PConnectionPool) closeConnection(conn *P2PConnection) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if !conn.Connected {
		return
	}

	conn.Connected = false

	// Update statistics
	p.stats.ActiveConnections--
	if p.stats.ActiveConnections < 0 {
		p.stats.ActiveConnections = 0
	}
}

// updateStats safely updates P2P pool statistics
func (p *P2PConnectionPool) updateStats(fn func(*P2PStats)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fn(p.stats)
}

// runHealthCheck performs periodic health checks on connections
func (p *P2PConnectionPool) runHealthCheck() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performHealthCheck()
		}
	}
}

// performHealthCheck checks the health of all connections
func (p *P2PConnectionPool) performHealthCheck() {
	p.mu.RLock()
	connections := make([]*P2PConnection, 0, len(p.connections))
	for _, conn := range p.connections {
		connections = append(connections, conn)
	}
	p.mu.RUnlock()

	for _, conn := range connections {
		if !p.isConnectionHealthy(conn) {
			p.mu.Lock()
			delete(p.connections, conn.PeerID)
			p.mu.Unlock()

			p.closeConnection(conn)

			p.updateStats(func(s *P2PStats) {
				s.ConnectionErrors++
			})
		}
	}
}

// isConnectionHealthy checks if a connection is healthy
func (p *P2PConnectionPool) isConnectionHealthy(conn *P2PConnection) bool {
	conn.mu.RLock()
	defer conn.mu.RUnlock()

	// Check if connection is too old
	if time.Since(conn.CreatedAt) > p.config.MaxLifetime {
		return false
	}

	// Check if connection has been idle too long
	if time.Since(conn.LastUsed) > p.config.IdleTimeout {
		return false
	}

	// Check if connection is still connected
	return conn.Connected
}

// runCleanup performs periodic cleanup of old connections
func (p *P2PConnectionPool) runCleanup() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.IdleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.performCleanup()
		}
	}
}

// performCleanup removes old or idle connections
func (p *P2PConnectionPool) performCleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()

	for peerID, conn := range p.connections {
		conn.mu.RLock()
		shouldRemove := !conn.Connected ||
			now.Sub(conn.LastUsed) > p.config.IdleTimeout ||
			now.Sub(conn.CreatedAt) > p.config.MaxLifetime
		conn.mu.RUnlock()

		if shouldRemove {
			p.closeConnection(conn)
			delete(p.connections, peerID)
		}
	}
}
