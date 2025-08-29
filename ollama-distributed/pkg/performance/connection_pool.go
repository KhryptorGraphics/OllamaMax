package performance

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

// ConnectionPool provides optimized connection pooling for network operations
type ConnectionPool struct {
	config *OptimizerConfig

	// Connection management
	connections chan *PooledConnection
	factory     ConnectionFactory
	validator   ConnectionValidator

	// Pool state
	activeConnections int32
	totalConnections  int32
	maxConnections    int32

	// Statistics
	stats *PoolStats

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
	closed bool
}

// PooledConnection represents a connection in the pool
type PooledConnection struct {
	conn        net.Conn
	createdAt   time.Time
	lastUsedAt  time.Time
	usageCount  int32
	pool        *ConnectionPool
}

// PoolStats tracks connection pool statistics
type PoolStats struct {
	// Connection counts
	ActiveConnections   int   `json:"active_connections"`
	IdleConnections     int   `json:"idle_connections"`
	TotalConnections    int   `json:"total_connections"`
	PeakConnections     int   `json:"peak_connections"`

	// Usage statistics
	ConnectionsCreated  int64 `json:"connections_created"`
	ConnectionsDestroyed int64 `json:"connections_destroyed"`
	ConnectionsReused   int64 `json:"connections_reused"`
	ConnectionsTimedOut int64 `json:"connections_timed_out"`

	// Performance metrics
	AverageWaitTime     time.Duration `json:"average_wait_time"`
	AverageUsageCount   float64       `json:"average_usage_count"`
	PoolUtilization     float64       `json:"pool_utilization"`

	// Error tracking
	Errors              int64 `json:"errors"`
	ValidationFailures  int64 `json:"validation_failures"`
	TimeoutErrors       int64 `json:"timeout_errors"`

	LastUpdated         time.Time `json:"last_updated"`
}

// ConnectionFactory creates new connections
type ConnectionFactory func() (net.Conn, error)

// ConnectionValidator validates connection health
type ConnectionValidator func(net.Conn) bool

// NewConnectionPool creates a new optimized connection pool
func NewConnectionPool(config *OptimizerConfig) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:         config,
		connections:    make(chan *PooledConnection, config.MaxConnections),
		maxConnections: int32(config.MaxConnections),
		stats:          &PoolStats{LastUpdated: time.Now()},
		ctx:            ctx,
		cancel:         cancel,
	}

	// Set default factory and validator
	pool.factory = pool.defaultFactory
	pool.validator = pool.defaultValidator

	return pool
}

// Start starts the connection pool
func (cp *ConnectionPool) Start() error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.closed {
		return errors.New("connection pool is closed")
	}

	// Start connection maintenance goroutine
	cp.wg.Add(1)
	go cp.runMaintenance()

	// Start statistics updater
	cp.wg.Add(1)
	go cp.runStatsUpdater()

	// Pre-warm the pool with idle connections
	go cp.prewarmPool()

	log.Info().
		Int("max_connections", int(cp.maxConnections)).
		Msg("Connection pool started")

	return nil
}

// Stop stops the connection pool
func (cp *ConnectionPool) Stop() error {
	cp.mu.Lock()
	if cp.closed {
		cp.mu.Unlock()
		return nil
	}
	cp.closed = true
	cp.mu.Unlock()

	cp.cancel()
	cp.wg.Wait()

	// Close all connections in the pool
	close(cp.connections)
	for conn := range cp.connections {
		conn.Close()
	}

	log.Info().Msg("Connection pool stopped")
	return nil
}

// Get retrieves a connection from the pool
func (cp *ConnectionPool) Get() (*PooledConnection, error) {
	return cp.GetWithTimeout(cp.config.ConnectionTimeout)
}

// GetWithTimeout retrieves a connection with a timeout
func (cp *ConnectionPool) GetWithTimeout(timeout time.Duration) (*PooledConnection, error) {
	if cp.closed {
		return nil, errors.New("connection pool is closed")
	}

	start := time.Now()
	
	// Try to get an existing connection from the pool
	select {
	case conn := <-cp.connections:
		// Validate the connection
		if cp.validator(conn.conn) {
			conn.lastUsedAt = time.Now()
			atomic.AddInt32(&conn.usageCount, 1)
			atomic.AddInt32(&cp.activeConnections, 1)
			atomic.AddInt64(&cp.stats.ConnectionsReused, 1)
			return conn, nil
		} else {
			// Connection is invalid, close it and try to create a new one
			conn.Close()
			atomic.AddInt64(&cp.stats.ValidationFailures, 1)
		}
	case <-time.After(timeout):
		atomic.AddInt64(&cp.stats.TimeoutErrors, 1)
		return nil, errors.New("timeout waiting for connection")
	default:
		// No connections available, try to create a new one
	}

	// Try to create a new connection if under the limit
	if atomic.LoadInt32(&cp.totalConnections) < cp.maxConnections {
		conn, err := cp.createConnection()
		if err != nil {
			atomic.AddInt64(&cp.stats.Errors, 1)
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}
		
		atomic.AddInt32(&cp.activeConnections, 1)
		cp.updateWaitTime(time.Since(start))
		return conn, nil
	}

	// Wait for a connection to become available
	ctx, cancel := context.WithTimeout(cp.ctx, timeout)
	defer cancel()

	select {
	case conn := <-cp.connections:
		if cp.validator(conn.conn) {
			conn.lastUsedAt = time.Now()
			atomic.AddInt32(&conn.usageCount, 1)
			atomic.AddInt32(&cp.activeConnections, 1)
			atomic.AddInt64(&cp.stats.ConnectionsReused, 1)
			cp.updateWaitTime(time.Since(start))
			return conn, nil
		} else {
			conn.Close()
			atomic.AddInt64(&cp.stats.ValidationFailures, 1)
			return nil, errors.New("all connections are invalid")
		}
	case <-ctx.Done():
		atomic.AddInt64(&cp.stats.TimeoutErrors, 1)
		return nil, errors.New("timeout waiting for connection")
	}
}

// Put returns a connection to the pool
func (cp *ConnectionPool) Put(conn *PooledConnection) {
	if cp.closed || conn == nil {
		if conn != nil {
			conn.Close()
		}
		return
	}

	atomic.AddInt32(&cp.activeConnections, -1)

	// Check if connection is still valid and not too old
	if !cp.validator(conn.conn) || time.Since(conn.createdAt) > cp.config.IdleTimeout*10 {
		conn.Close()
		return
	}

	// Try to return the connection to the pool
	select {
	case cp.connections <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

// Scale adjusts the maximum number of connections in the pool
func (cp *ConnectionPool) Scale(newMaxConnections int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	oldMax := cp.maxConnections
	cp.maxConnections = int32(newMaxConnections)

	if newMaxConnections > int(oldMax) {
		// Expanding pool - create a new channel with larger capacity
		newConnections := make(chan *PooledConnection, newMaxConnections)
		
		// Transfer existing connections
		close(cp.connections)
		for conn := range cp.connections {
			select {
			case newConnections <- conn:
			default:
				conn.Close()
			}
		}
		
		cp.connections = newConnections
		
		log.Info().
			Int("old_max", int(oldMax)).
			Int("new_max", newMaxConnections).
			Msg("Scaled up connection pool")
	} else {
		// Shrinking pool - close excess connections
		excess := int(oldMax) - newMaxConnections
		for i := 0; i < excess; i++ {
			select {
			case conn := <-cp.connections:
				conn.Close()
			default:
				break
			}
		}
		
		log.Info().
			Int("old_max", int(oldMax)).
			Int("new_max", newMaxConnections).
			Msg("Scaled down connection pool")
	}
}

// GetStats returns current pool statistics
func (cp *ConnectionPool) GetStats() *PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := *cp.stats
	stats.ActiveConnections = int(atomic.LoadInt32(&cp.activeConnections))
	stats.TotalConnections = int(atomic.LoadInt32(&cp.totalConnections))
	stats.IdleConnections = stats.TotalConnections - stats.ActiveConnections
	
	if cp.maxConnections > 0 {
		stats.PoolUtilization = float64(stats.ActiveConnections) / float64(cp.maxConnections) * 100
	}
	
	return &stats
}

// createConnection creates a new pooled connection
func (cp *ConnectionPool) createConnection() (*PooledConnection, error) {
	conn, err := cp.factory()
	if err != nil {
		return nil, err
	}

	pooledConn := &PooledConnection{
		conn:       conn,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		usageCount: 1,
		pool:       cp,
	}

	atomic.AddInt32(&cp.totalConnections, 1)
	atomic.AddInt64(&cp.stats.ConnectionsCreated, 1)

	// Update peak connections
	current := atomic.LoadInt32(&cp.totalConnections)
	if int(current) > cp.stats.PeakConnections {
		cp.stats.PeakConnections = int(current)
	}

	return pooledConn, nil
}

// runMaintenance runs connection pool maintenance
func (cp *ConnectionPool) runMaintenance() {
	defer cp.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cp.ctx.Done():
			return
		case <-ticker.C:
			cp.performMaintenance()
		}
	}
}

// performMaintenance performs pool maintenance tasks
func (cp *ConnectionPool) performMaintenance() {
	// Remove idle connections that have exceeded the idle timeout
	var connectionsToClose []*PooledConnection
	
	// Drain connections from pool to check them
	for {
		select {
		case conn := <-cp.connections:
			if time.Since(conn.lastUsedAt) > cp.config.IdleTimeout {
				connectionsToClose = append(connectionsToClose, conn)
			} else {
				// Put valid connection back
				select {
				case cp.connections <- conn:
				default:
					// Pool is full, close this connection
					connectionsToClose = append(connectionsToClose, conn)
				}
			}
		default:
			// No more connections to check
			goto cleanup
		}
	}
	
cleanup:
	// Close expired connections
	for _, conn := range connectionsToClose {
		conn.Close()
		atomic.AddInt64(&cp.stats.ConnectionsTimedOut, 1)
	}

	if len(connectionsToClose) > 0 {
		log.Debug().
			Int("closed_connections", len(connectionsToClose)).
			Msg("Closed idle connections during maintenance")
	}
}

// runStatsUpdater updates pool statistics
func (cp *ConnectionPool) runStatsUpdater() {
	defer cp.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cp.ctx.Done():
			return
		case <-ticker.C:
			cp.updateStats()
		}
	}
}

// updateStats updates pool statistics
func (cp *ConnectionPool) updateStats() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.stats.LastUpdated = time.Now()
	
	// Calculate average usage count
	if cp.stats.ConnectionsCreated > 0 {
		totalUsage := cp.stats.ConnectionsReused + cp.stats.ConnectionsCreated
		cp.stats.AverageUsageCount = float64(totalUsage) / float64(cp.stats.ConnectionsCreated)
	}
}

// updateWaitTime updates the average wait time for connections
func (cp *ConnectionPool) updateWaitTime(waitTime time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	if cp.stats.AverageWaitTime == 0 {
		cp.stats.AverageWaitTime = waitTime
	} else {
		cp.stats.AverageWaitTime = (cp.stats.AverageWaitTime + waitTime) / 2
	}
}

// prewarmPool creates initial connections to warm up the pool
func (cp *ConnectionPool) prewarmPool() {
	// Create initial connections up to 25% of max capacity
	initialConnections := int(cp.maxConnections) / 4
	if initialConnections < 1 {
		initialConnections = 1
	}

	for i := 0; i < initialConnections; i++ {
		conn, err := cp.createConnection()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create initial connection")
			continue
		}

		select {
		case cp.connections <- conn:
		default:
			conn.Close()
			break
		}
	}

	log.Info().
		Int("prewarmed_connections", initialConnections).
		Msg("Connection pool prewarmed")
}

// defaultFactory creates a default TCP connection (placeholder)
func (cp *ConnectionPool) defaultFactory() (net.Conn, error) {
	// This is a placeholder - in real implementation, this would connect to actual services
	// For testing purposes, we create a pipe connection
	client, server := net.Pipe()
	go func() {
		defer server.Close()
		// Simple echo server for testing
		buffer := make([]byte, 1024)
		for {
			n, err := server.Read(buffer)
			if err != nil {
				return
			}
			server.Write(buffer[:n])
		}
	}()
	return client, nil
}

// defaultValidator validates connection health
func (cp *ConnectionPool) defaultValidator(conn net.Conn) bool {
	if conn == nil {
		return false
	}

	// Try to set a read deadline to test if connection is alive
	err := conn.SetReadDeadline(time.Now().Add(time.Millisecond))
	if err != nil {
		return false
	}

	// Reset the deadline
	conn.SetReadDeadline(time.Time{})
	return true
}

// SetFactory sets a custom connection factory
func (cp *ConnectionPool) SetFactory(factory ConnectionFactory) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.factory = factory
}

// SetValidator sets a custom connection validator
func (cp *ConnectionPool) SetValidator(validator ConnectionValidator) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.validator = validator
}

// Close closes the pooled connection
func (pc *PooledConnection) Close() error {
	if pc.conn != nil {
		atomic.AddInt32(&pc.pool.totalConnections, -1)
		atomic.AddInt64(&pc.pool.stats.ConnectionsDestroyed, 1)
		return pc.conn.Close()
	}
	return nil
}

// Read implements the io.Reader interface
func (pc *PooledConnection) Read(b []byte) (n int, err error) {
	return pc.conn.Read(b)
}

// Write implements the io.Writer interface
func (pc *PooledConnection) Write(b []byte) (n int, err error) {
	return pc.conn.Write(b)
}

// SetDeadline sets the read and write deadlines
func (pc *PooledConnection) SetDeadline(t time.Time) error {
	return pc.conn.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (pc *PooledConnection) SetReadDeadline(t time.Time) error {
	return pc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (pc *PooledConnection) SetWriteDeadline(t time.Time) error {
	return pc.conn.SetWriteDeadline(t)
}

// LocalAddr returns the local network address
func (pc *PooledConnection) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

// RemoteAddr returns the remote network address
func (pc *PooledConnection) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}