package host

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// ConnectionPool manages a pool of reusable connections
type ConnectionPool struct {
	host        host.Host
	mu          sync.RWMutex
	connections map[peer.ID]*PooledConnection
	config      *PoolConfig

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Metrics
	metrics *PoolMetrics
}

// PooledConnection represents a connection in the pool
type PooledConnection struct {
	Conn     network.Conn
	PeerID   peer.ID
	LastUsed time.Time
	UseCount int64
	Created  time.Time
	Quality  *ConnectionQuality

	// Stream management
	activeStreams map[protocol.ID][]network.Stream
	streamMu      sync.RWMutex
}

// ConnectionQuality tracks connection quality metrics
type ConnectionQuality struct {
	Latency      time.Duration
	Bandwidth    int64 // bytes per second
	PacketLoss   float64
	Jitter       time.Duration
	Reliability  float64 // 0.0 to 1.0
	LastMeasured time.Time
}

// PoolConfig configures the connection pool
type PoolConfig struct {
	MaxConnections       int
	MaxIdleTime          time.Duration
	MaxConnectionAge     time.Duration
	CleanupInterval      time.Duration
	QualityCheckInterval time.Duration
	MinQualityThreshold  float64

	// Stream pooling
	MaxStreamsPerConn int
	StreamIdleTimeout time.Duration
}

// PoolMetrics tracks pool performance
type PoolMetrics struct {
	TotalConnections   int64
	ActiveConnections  int64
	IdleConnections    int64
	ConnectionHits     int64
	ConnectionMisses   int64
	ConnectionsCreated int64
	ConnectionsRemoved int64
	QualityChecks      int64
	LastCleanup        time.Time
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(h host.Host, config *PoolConfig) *ConnectionPool {
	ctx, cancel := context.WithCancel(context.Background())

	if config == nil {
		config = &PoolConfig{
			MaxConnections:       100,
			MaxIdleTime:          5 * time.Minute,
			MaxConnectionAge:     30 * time.Minute,
			CleanupInterval:      1 * time.Minute,
			QualityCheckInterval: 30 * time.Second,
			MinQualityThreshold:  0.7,
			MaxStreamsPerConn:    10,
			StreamIdleTimeout:    2 * time.Minute,
		}
	}

	pool := &ConnectionPool{
		host:        h,
		connections: make(map[peer.ID]*PooledConnection),
		config:      config,
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &PoolMetrics{},
	}

	// Start background tasks
	pool.wg.Add(2)
	go pool.cleanupLoop()
	go pool.qualityCheckLoop()

	return pool
}

// GetConnection gets a connection from the pool or creates a new one
func (p *ConnectionPool) GetConnection(ctx context.Context, peerID peer.ID) (*PooledConnection, error) {
	p.mu.RLock()
	if conn, exists := p.connections[peerID]; exists {
		if p.isConnectionValid(conn) {
			conn.LastUsed = time.Now()
			conn.UseCount++
			p.metrics.ConnectionHits++
			p.mu.RUnlock()
			return conn, nil
		}
	}
	p.mu.RUnlock()

	// Connection not found or invalid, create new one
	p.metrics.ConnectionMisses++
	return p.createConnection(ctx, peerID)
}

// createConnection creates a new pooled connection
func (p *ConnectionPool) createConnection(ctx context.Context, peerID peer.ID) (*PooledConnection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if we've reached the maximum number of connections
	if len(p.connections) >= p.config.MaxConnections {
		// Remove oldest idle connection
		if err := p.evictOldestConnection(); err != nil {
			return nil, fmt.Errorf("failed to evict connection: %w", err)
		}
	}

	// Get network connection
	conn := p.host.Network().ConnsToPeer(peerID)
	if len(conn) == 0 {
		return nil, fmt.Errorf("no connection to peer %s", peerID)
	}

	// Create pooled connection
	pooledConn := &PooledConnection{
		Conn:          conn[0], // Use first available connection
		PeerID:        peerID,
		LastUsed:      time.Now(),
		UseCount:      1,
		Created:       time.Now(),
		Quality:       &ConnectionQuality{},
		activeStreams: make(map[protocol.ID][]network.Stream),
	}

	// Measure initial quality
	p.measureConnectionQuality(pooledConn)

	p.connections[peerID] = pooledConn
	p.metrics.ConnectionsCreated++
	p.metrics.TotalConnections++

	return pooledConn, nil
}

// GetStream gets a stream from the connection pool
func (p *ConnectionPool) GetStream(ctx context.Context, peerID peer.ID, protocolID protocol.ID) (network.Stream, error) {
	conn, err := p.GetConnection(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Check for existing idle stream
	conn.streamMu.Lock()
	if streams, exists := conn.activeStreams[protocolID]; exists && len(streams) > 0 {
		// Reuse existing stream
		stream := streams[0]
		conn.activeStreams[protocolID] = streams[1:]
		conn.streamMu.Unlock()
		return stream, nil
	}
	conn.streamMu.Unlock()

	// Create new stream
	stream, err := p.host.NewStream(ctx, peerID, protocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return stream, nil
}

// ReturnStream returns a stream to the pool for reuse
func (p *ConnectionPool) ReturnStream(stream network.Stream, protocolID protocol.ID) {
	peerID := stream.Conn().RemotePeer()

	p.mu.RLock()
	conn, exists := p.connections[peerID]
	p.mu.RUnlock()

	if !exists {
		stream.Close()
		return
	}

	conn.streamMu.Lock()
	defer conn.streamMu.Unlock()

	// Check if we can pool this stream
	if len(conn.activeStreams[protocolID]) < p.config.MaxStreamsPerConn {
		conn.activeStreams[protocolID] = append(conn.activeStreams[protocolID], stream)
	} else {
		stream.Close()
	}
}

// isConnectionValid checks if a connection is still valid
func (p *ConnectionPool) isConnectionValid(conn *PooledConnection) bool {
	now := time.Now()

	// Check if connection is too old
	if now.Sub(conn.Created) > p.config.MaxConnectionAge {
		return false
	}

	// Check if connection has been idle too long
	if now.Sub(conn.LastUsed) > p.config.MaxIdleTime {
		return false
	}

	// Check connection quality
	if conn.Quality.Reliability < p.config.MinQualityThreshold {
		return false
	}

	// Check if underlying connection is still open
	if conn.Conn.IsClosed() {
		return false
	}

	return true
}

// evictOldestConnection removes the oldest idle connection
func (p *ConnectionPool) evictOldestConnection() error {
	var oldestConn *PooledConnection
	var oldestPeerID peer.ID
	oldestTime := time.Now()

	for peerID, conn := range p.connections {
		if conn.LastUsed.Before(oldestTime) {
			oldestTime = conn.LastUsed
			oldestConn = conn
			oldestPeerID = peerID
		}
	}

	if oldestConn != nil {
		p.removeConnection(oldestPeerID, oldestConn)
		return nil
	}

	return fmt.Errorf("no connections to evict")
}

// removeConnection removes a connection from the pool
func (p *ConnectionPool) removeConnection(peerID peer.ID, conn *PooledConnection) {
	// Close all pooled streams
	conn.streamMu.Lock()
	for _, streams := range conn.activeStreams {
		for _, stream := range streams {
			stream.Close()
		}
	}
	conn.streamMu.Unlock()

	delete(p.connections, peerID)
	p.metrics.ConnectionsRemoved++
	p.metrics.TotalConnections--
}

// measureConnectionQuality measures the quality of a connection
func (p *ConnectionPool) measureConnectionQuality(conn *PooledConnection) {
	// This is a simplified quality measurement
	// In a real implementation, you would measure actual latency, bandwidth, etc.

	start := time.Now()

	// Simulate ping measurement
	// In reality, you would send a ping message and measure response time
	latency := time.Since(start)

	conn.Quality.Latency = latency
	conn.Quality.LastMeasured = time.Now()

	// Calculate reliability based on connection age and usage
	ageScore := 1.0 - (float64(time.Since(conn.Created)) / float64(p.config.MaxConnectionAge))
	usageScore := 1.0 / (1.0 + float64(conn.UseCount)/100.0) // Diminishing returns

	conn.Quality.Reliability = (ageScore + usageScore) / 2.0
	if conn.Quality.Reliability > 1.0 {
		conn.Quality.Reliability = 1.0
	}
	if conn.Quality.Reliability < 0.0 {
		conn.Quality.Reliability = 0.0
	}
}

// cleanupLoop periodically cleans up expired connections
func (p *ConnectionPool) cleanupLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

// cleanup removes expired and invalid connections
func (p *ConnectionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var toRemove []peer.ID

	for peerID, conn := range p.connections {
		if !p.isConnectionValid(conn) {
			toRemove = append(toRemove, peerID)
		}
	}

	for _, peerID := range toRemove {
		p.removeConnection(peerID, p.connections[peerID])
	}

	p.metrics.LastCleanup = time.Now()
}

// qualityCheckLoop periodically checks connection quality
func (p *ConnectionPool) qualityCheckLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.QualityCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.checkQuality()
		}
	}
}

// checkQuality measures quality for all connections
func (p *ConnectionPool) checkQuality() {
	p.mu.RLock()
	connections := make([]*PooledConnection, 0, len(p.connections))
	for _, conn := range p.connections {
		connections = append(connections, conn)
	}
	p.mu.RUnlock()

	for _, conn := range connections {
		p.measureConnectionQuality(conn)
		p.metrics.QualityChecks++
	}
}

// GetMetrics returns pool metrics
func (p *ConnectionPool) GetMetrics() *PoolMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metrics := *p.metrics
	metrics.ActiveConnections = int64(len(p.connections))

	// Count idle connections
	now := time.Now()
	idleCount := int64(0)
	for _, conn := range p.connections {
		if now.Sub(conn.LastUsed) > time.Minute {
			idleCount++
		}
	}
	metrics.IdleConnections = idleCount

	return &metrics
}

// Close closes the connection pool
func (p *ConnectionPool) Close() error {
	p.cancel()
	p.wg.Wait()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Close all connections
	for peerID, conn := range p.connections {
		p.removeConnection(peerID, conn)
	}

	return nil
}
