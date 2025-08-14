package pool

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnectionPool manages a pool of network connections
type ConnectionPool struct {
	config *Config

	// Connection management
	connections chan net.Conn
	factory     ConnectionFactory

	// Statistics
	stats *Stats
	mu    sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Config holds connection pool configuration
type Config struct {
	// Pool sizing
	MinConnections int `yaml:"min_connections"`
	MaxConnections int `yaml:"max_connections"`

	// Connection settings
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxLifetime    time.Duration `yaml:"max_lifetime"`

	// Health checking
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `yaml:"health_check_timeout"`

	// Retry settings
	MaxRetries    int           `yaml:"max_retries"`
	RetryInterval time.Duration `yaml:"retry_interval"`
}

// DefaultConfig returns default connection pool configuration
func DefaultConfig() *Config {
	return &Config{
		MinConnections:      5,
		MaxConnections:      50,
		ConnectTimeout:      10 * time.Second,
		IdleTimeout:         5 * time.Minute,
		MaxLifetime:         30 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		MaxRetries:          3,
		RetryInterval:       1 * time.Second,
	}
}

// ConnectionFactory creates new connections
type ConnectionFactory interface {
	Create() (net.Conn, error)
	Validate(conn net.Conn) error
	Close(conn net.Conn) error
}

// Stats holds connection pool statistics
type Stats struct {
	// Pool state
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	TotalConnections  int `json:"total_connections"`

	// Usage statistics
	ConnectionsCreated int64 `json:"connections_created"`
	ConnectionsClosed  int64 `json:"connections_closed"`
	ConnectionsReused  int64 `json:"connections_reused"`

	// Error statistics
	ConnectionErrors  int64 `json:"connection_errors"`
	HealthCheckErrors int64 `json:"health_check_errors"`
	TimeoutErrors     int64 `json:"timeout_errors"`

	// Performance metrics
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	AverageConnectTime time.Duration `json:"average_connect_time"`

	// Timestamps
	LastActivity time.Time `json:"last_activity"`
	StartTime    time.Time `json:"start_time"`
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(config *Config, factory ConnectionFactory) *ConnectionPool {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		config:      config,
		connections: make(chan net.Conn, config.MaxConnections),
		factory:     factory,
		stats:       &Stats{StartTime: time.Now()},
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool
}

// Start initializes the connection pool
func (p *ConnectionPool) Start() error {
	// Pre-populate with minimum connections
	for i := 0; i < p.config.MinConnections; i++ {
		conn, err := p.createConnection()
		if err != nil {
			return fmt.Errorf("failed to create initial connection: %w", err)
		}

		select {
		case p.connections <- conn:
		default:
			p.factory.Close(conn)
		}
	}

	// Start health check routine
	p.wg.Add(1)
	go p.runHealthCheck()

	// Start cleanup routine
	p.wg.Add(1)
	go p.runCleanup()

	return nil
}

// Stop shuts down the connection pool
func (p *ConnectionPool) Stop() error {
	p.cancel()
	p.wg.Wait()

	// Close all connections
	close(p.connections)
	for conn := range p.connections {
		p.factory.Close(conn)
		p.updateStats(func(s *Stats) {
			s.ConnectionsClosed++
			s.TotalConnections--
		})
	}

	return nil
}

// Get retrieves a connection from the pool
func (p *ConnectionPool) Get(ctx context.Context) (net.Conn, error) {
	start := time.Now()
	defer func() {
		p.updateStats(func(s *Stats) {
			s.AverageWaitTime = time.Since(start)
			s.LastActivity = time.Now()
		})
	}()

	// Try to get an existing connection
	select {
	case conn := <-p.connections:
		// Validate connection
		if err := p.factory.Validate(conn); err == nil {
			p.updateStats(func(s *Stats) {
				s.ConnectionsReused++
				s.ActiveConnections++
				s.IdleConnections--
			})
			return conn, nil
		}

		// Connection is invalid, close it
		p.factory.Close(conn)
		p.updateStats(func(s *Stats) {
			s.ConnectionsClosed++
			s.TotalConnections--
		})

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// No idle connections available
	}

	// Create new connection if under limit
	p.mu.RLock()
	canCreate := p.stats.TotalConnections < p.config.MaxConnections
	p.mu.RUnlock()

	if canCreate {
		conn, err := p.createConnection()
		if err != nil {
			p.updateStats(func(s *Stats) {
				s.ConnectionErrors++
			})
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}

		p.updateStats(func(s *Stats) {
			s.ActiveConnections++
		})

		return conn, nil
	}

	// Wait for a connection to become available
	select {
	case conn := <-p.connections:
		if err := p.factory.Validate(conn); err == nil {
			p.updateStats(func(s *Stats) {
				s.ConnectionsReused++
				s.ActiveConnections++
				s.IdleConnections--
			})
			return conn, nil
		}

		// Connection is invalid, try again
		p.factory.Close(conn)
		p.updateStats(func(s *Stats) {
			s.ConnectionsClosed++
			s.TotalConnections--
		})

		return p.Get(ctx)

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn net.Conn) error {
	if conn == nil {
		return fmt.Errorf("cannot put nil connection")
	}

	// Validate connection before putting back
	if err := p.factory.Validate(conn); err != nil {
		p.factory.Close(conn)
		p.updateStats(func(s *Stats) {
			s.ConnectionsClosed++
			s.TotalConnections--
			s.ActiveConnections--
		})
		return err
	}

	select {
	case p.connections <- conn:
		p.updateStats(func(s *Stats) {
			s.ActiveConnections--
			s.IdleConnections++
			s.LastActivity = time.Now()
		})
		return nil

	default:
		// Pool is full, close the connection
		p.factory.Close(conn)
		p.updateStats(func(s *Stats) {
			s.ConnectionsClosed++
			s.TotalConnections--
			s.ActiveConnections--
		})
		return nil
	}
}

// Stats returns current pool statistics
func (p *ConnectionPool) Stats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy
	return *p.stats
}

// createConnection creates a new connection
func (p *ConnectionPool) createConnection() (net.Conn, error) {
	start := time.Now()

	conn, err := p.factory.Create()
	if err != nil {
		return nil, err
	}

	p.updateStats(func(s *Stats) {
		s.ConnectionsCreated++
		s.TotalConnections++
		s.AverageConnectTime = time.Since(start)
	})

	return conn, nil
}

// updateStats safely updates pool statistics
func (p *ConnectionPool) updateStats(fn func(*Stats)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fn(p.stats)
}

// runHealthCheck performs periodic health checks on idle connections
func (p *ConnectionPool) runHealthCheck() {
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

// performHealthCheck checks the health of idle connections
func (p *ConnectionPool) performHealthCheck() {
	// Get current idle connections
	var idleConns []net.Conn

	for {
		select {
		case conn := <-p.connections:
			idleConns = append(idleConns, conn)
		default:
			goto checkConnections
		}
	}

checkConnections:
	// Check each connection and put back healthy ones
	for _, conn := range idleConns {
		if err := p.factory.Validate(conn); err == nil {
			select {
			case p.connections <- conn:
			default:
				// Pool is full, close connection
				p.factory.Close(conn)
				p.updateStats(func(s *Stats) {
					s.ConnectionsClosed++
					s.TotalConnections--
					s.IdleConnections--
				})
			}
		} else {
			// Connection is unhealthy, close it
			p.factory.Close(conn)
			p.updateStats(func(s *Stats) {
				s.ConnectionsClosed++
				s.TotalConnections--
				s.IdleConnections--
				s.HealthCheckErrors++
			})
		}
	}
}

// runCleanup performs periodic cleanup of old connections
func (p *ConnectionPool) runCleanup() {
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

// performCleanup removes old or excess connections
func (p *ConnectionPool) performCleanup() {
	// TODO: Implement connection age tracking and cleanup
	// For now, just ensure we don't exceed max connections

	p.mu.RLock()
	excess := p.stats.TotalConnections - p.config.MaxConnections
	p.mu.RUnlock()

	for i := 0; i < excess; i++ {
		select {
		case conn := <-p.connections:
			p.factory.Close(conn)
			p.updateStats(func(s *Stats) {
				s.ConnectionsClosed++
				s.TotalConnections--
				s.IdleConnections--
			})
		default:
			return
		}
	}
}
