package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	// PostgreSQL configuration
	Host     string `yaml:"host" env:"OLLAMA_DB_HOST"`
	Port     int    `yaml:"port" env:"OLLAMA_DB_PORT"`
	Name     string `yaml:"name" env:"OLLAMA_DB_NAME"`
	User     string `yaml:"user" env:"OLLAMA_DB_USER"`
	Password string `yaml:"password" env:"OLLAMA_DB_PASSWORD"`
	SSLMode  string `yaml:"ssl_mode" env:"OLLAMA_DB_SSL_MODE"`
	
	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" env:"OLLAMA_DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"OLLAMA_DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"OLLAMA_DB_CONN_MAX_LIFETIME"`
	
	// Redis configuration
	RedisHost     string `yaml:"redis_host" env:"OLLAMA_REDIS_HOST"`
	RedisPort     int    `yaml:"redis_port" env:"OLLAMA_REDIS_PORT"`
	RedisPassword string `yaml:"redis_password" env:"OLLAMA_REDIS_PASSWORD"`
	RedisDB       int    `yaml:"redis_db" env:"OLLAMA_REDIS_DB"`
	
	// Redis connection settings
	RedisPoolSize     int           `yaml:"redis_pool_size" env:"OLLAMA_REDIS_POOL_SIZE"`
	RedisMinIdleConns int           `yaml:"redis_min_idle_conns" env:"OLLAMA_REDIS_MIN_IDLE_CONNS"`
	RedisDialTimeout  time.Duration `yaml:"redis_dial_timeout" env:"OLLAMA_REDIS_DIAL_TIMEOUT"`
	RedisReadTimeout  time.Duration `yaml:"redis_read_timeout" env:"OLLAMA_REDIS_READ_TIMEOUT"`
	RedisWriteTimeout time.Duration `yaml:"redis_write_timeout" env:"OLLAMA_REDIS_WRITE_TIMEOUT"`
}

// DatabaseManager manages database connections and provides access to repositories
type DatabaseManager struct {
	DB     *sqlx.DB
	Redis  *redis.Client
	config *DatabaseConfig
	logger *slog.Logger

	// Repositories
	Models    *ModelRepository
	Nodes     *NodeRepository
	Users     *UserRepository
	Sessions  *SessionRepository
	Inference *InferenceRepository
	Audit     *AuditRepository
	Config    *ConfigRepository

	// Prometheus metrics
	registry                    *prometheus.Registry
	dbConnectionsOpen           prometheus.Gauge
	dbConnectionsInUse          prometheus.Gauge
	dbConnectionsIdle           prometheus.Gauge
	dbConnectionsWaitCount      prometheus.Gauge
	dbConnectionsWaitDuration   prometheus.Gauge
	dbConnectionsMax            prometheus.Gauge
	dbQueriesTotal              *prometheus.CounterVec
	dbQueryDuration             *prometheus.HistogramVec
	redisPoolSize               prometheus.Gauge
	redisCommandsTotal          *prometheus.CounterVec
	redisCommandDuration        *prometheus.HistogramVec
	cacheHitsTotal              prometheus.Counter
	cacheMissesTotal            prometheus.Counter
	cacheOperationDuration      prometheus.Histogram

	// Lifecycle management
	metricsCancel context.CancelFunc
}

// NewDatabaseManager creates a new database manager with all repositories
func NewDatabaseManager(config *DatabaseConfig, logger *slog.Logger) (*DatabaseManager, error) {
	// Set defaults for production-level performance
	if config.MaxOpenConns == 0 {
		// PERFORMANCE: Increased from 25 to 100 for better scalability (supports 10,000+ RPS)
		config.MaxOpenConns = 100
	}
	if config.MaxIdleConns == 0 {
		// PERFORMANCE: Increased idle connections to maintain pool efficiency
		config.MaxIdleConns = 20
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 5 * time.Minute
	}
	if config.SSLMode == "" {
		config.SSLMode = "prefer"
	}
	if config.RedisPoolSize == 0 {
		config.RedisPoolSize = 10
	}
	if config.RedisMinIdleConns == 0 {
		config.RedisMinIdleConns = 5
	}
	if config.RedisDialTimeout == 0 {
		config.RedisDialTimeout = 5 * time.Second
	}
	if config.RedisReadTimeout == 0 {
		config.RedisReadTimeout = 3 * time.Second
	}
	if config.RedisWriteTimeout == 0 {
		config.RedisWriteTimeout = 3 * time.Second
	}

	dm := &DatabaseManager{
		config: config,
		logger: logger,
	}

	// Initialize PostgreSQL connection
	if err := dm.initializePostgreSQL(); err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Initialize Redis connection
	if err := dm.initializeRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize repositories
	dm.initializeRepositories()

	// Initialize Prometheus metrics
	dm.initializeMetrics()

	// Start periodic metrics collection
	dm.startMetricsCollection()

	logger.Info("Database manager initialized successfully",
		"postgres_host", config.Host,
		"postgres_port", config.Port,
		"postgres_db", config.Name,
		"redis_host", config.RedisHost,
		"redis_port", config.RedisPort)

	return dm, nil
}

// initializePostgreSQL sets up PostgreSQL connection with connection pooling
func (dm *DatabaseManager) initializePostgreSQL() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dm.config.Host, dm.config.Port, dm.config.User, dm.config.Password, dm.config.Name, dm.config.SSLMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(dm.config.MaxOpenConns)
	db.SetMaxIdleConns(dm.config.MaxIdleConns)
	db.SetConnMaxLifetime(dm.config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	dm.DB = db
	return nil
}

// initializeRedis sets up Redis connection with proper configuration
func (dm *DatabaseManager) initializeRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", dm.config.RedisHost, dm.config.RedisPort),
		Password:     dm.config.RedisPassword,
		DB:           dm.config.RedisDB,
		PoolSize:     dm.config.RedisPoolSize,
		MinIdleConns: dm.config.RedisMinIdleConns,
		DialTimeout:  dm.config.RedisDialTimeout,
		ReadTimeout:  dm.config.RedisReadTimeout,
		WriteTimeout: dm.config.RedisWriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	dm.Redis = rdb
	return nil
}

// initializeRepositories creates all repository instances
func (dm *DatabaseManager) initializeRepositories() {
	dm.Models = NewModelRepository(dm.DB, dm.Redis, dm.logger, dm)
	dm.Nodes = NewNodeRepository(dm.DB, dm.Redis, dm.logger, dm)
	dm.Users = NewUserRepository(dm.DB, dm.Redis, dm.logger, dm)
	dm.Sessions = NewSessionRepository(dm.DB, dm.Redis, dm.logger, dm)
	dm.Inference = NewInferenceRepository(dm.DB, dm.Redis, dm.logger, dm)
	dm.Audit = NewAuditRepository(dm.DB, dm.logger, dm)
	dm.Config = NewConfigRepository(dm.DB, dm.Redis, dm.logger, dm)
}

// initializeMetrics creates and registers Prometheus metrics
func (dm *DatabaseManager) initializeMetrics() {
	dm.registry = prometheus.NewRegistry()

	// Database connection pool metrics with ollamamax_database_ namespace
	dm.dbConnectionsOpen = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_open",
		Help: "Number of open database connections",
	})

	dm.dbConnectionsInUse = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_active",
		Help: "Number of database connections currently in use",
	})

	dm.dbConnectionsIdle = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_idle",
		Help: "Number of idle database connections",
	})

	// Additional connection pool metrics
	dm.dbConnectionsWaitCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_wait_count",
		Help: "Total number of connections waited for",
	})

	dm.dbConnectionsWaitDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_wait_duration_seconds",
		Help: "Total time waited for connections in seconds",
	})

	dm.dbConnectionsMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_db_connections_max",
		Help: "Maximum number of open connections to the database",
	})

	// Query metrics
	dm.dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollamamax_database_db_queries_total",
			Help: "Total number of database queries executed",
		},
		[]string{"operation", "table"},
	)

	dm.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ollamamax_database_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"operation", "table"},
	)

	// Redis pool metrics
	dm.redisPoolSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ollamamax_database_redis_pool_size",
		Help: "Configured Redis connection pool size",
	})

	dm.redisCommandsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ollamamax_database_redis_commands_total",
			Help: "Total number of Redis commands executed",
		},
		[]string{"command"},
	)

	dm.redisCommandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ollamamax_database_redis_command_duration_seconds",
			Help:    "Redis command duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"command"},
	)

	// Cache metrics
	dm.cacheHitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ollamamax_database_cache_hits_total",
		Help: "Total number of cache hits",
	})

	dm.cacheMissesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ollamamax_database_cache_misses_total",
		Help: "Total number of cache misses",
	})

	dm.cacheOperationDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "ollamamax_database_cache_operation_duration_seconds",
		Help:    "Cache operation duration in seconds",
		Buckets: []float64{0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1},
	})

	// Register all metrics
	dm.registry.MustRegister(
		dm.dbConnectionsOpen,
		dm.dbConnectionsInUse,
		dm.dbConnectionsIdle,
		dm.dbConnectionsWaitCount,
		dm.dbConnectionsWaitDuration,
		dm.dbConnectionsMax,
		dm.dbQueriesTotal,
		dm.dbQueryDuration,
		dm.redisPoolSize,
		dm.redisCommandsTotal,
		dm.redisCommandDuration,
		dm.cacheHitsTotal,
		dm.cacheMissesTotal,
		dm.cacheOperationDuration,
	)

	// Set initial static values
	dm.dbConnectionsMax.Set(float64(dm.config.MaxOpenConns))
	dm.redisPoolSize.Set(float64(dm.config.RedisPoolSize))

	dm.logger.Info("Prometheus metrics initialized and registered")
}

// startMetricsCollection starts a background goroutine to update pool metrics periodically
func (dm *DatabaseManager) startMetricsCollection() {
	ctx, cancel := context.WithCancel(context.Background())
	dm.metricsCancel = cancel

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		// Initial collection
		dm.updatePoolMetrics()

		for {
			select {
			case <-ctx.Done():
				dm.logger.Info("Stopping metrics collection")
				return
			case <-ticker.C:
				dm.updatePoolMetrics()
			}
		}
	}()

	dm.logger.Info("Started periodic metrics collection (every 15 seconds)")
}

// updatePoolMetrics updates database connection pool metrics
func (dm *DatabaseManager) updatePoolMetrics() {
	stats := dm.DB.Stats()

	dm.dbConnectionsOpen.Set(float64(stats.OpenConnections))
	dm.dbConnectionsInUse.Set(float64(stats.InUse))
	dm.dbConnectionsIdle.Set(float64(stats.Idle))
	dm.dbConnectionsWaitCount.Set(float64(stats.WaitCount))
	dm.dbConnectionsWaitDuration.Set(stats.WaitDuration.Seconds())

	dm.logger.Debug("Updated database pool metrics",
		"open", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_duration_seconds", stats.WaitDuration.Seconds(),
	)
}

// GetPrometheusRegistry returns the Prometheus registry for metrics exposure
// Deprecated: Use RegisterTo to register metrics on the main app registry instead
func (dm *DatabaseManager) GetPrometheusRegistry() *prometheus.Registry {
	return dm.registry
}

// RegisterTo registers all database metrics to the provided Prometheus registerer
// This ensures metrics are exposed at the main /metrics endpoint
func (dm *DatabaseManager) RegisterTo(registerer prometheus.Registerer) error {
	// Register all database metrics to the provided registerer
	collectors := []prometheus.Collector{
		dm.dbConnectionsOpen,
		dm.dbConnectionsInUse,
		dm.dbConnectionsIdle,
		dm.dbConnectionsWaitCount,
		dm.dbConnectionsWaitDuration,
		dm.dbConnectionsMax,
		dm.dbQueriesTotal,
		dm.dbQueryDuration,
		dm.redisPoolSize,
		dm.redisCommandsTotal,
		dm.redisCommandDuration,
		dm.cacheHitsTotal,
		dm.cacheMissesTotal,
		dm.cacheOperationDuration,
	}

	for _, collector := range collectors {
		if err := registerer.Register(collector); err != nil {
			// Check if already registered (not an error in our case)
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				return fmt.Errorf("failed to register database metrics: %w", err)
			}
		}
	}

	dm.logger.Info("Database metrics registered to main registry")
	return nil
}

// RecordQuery records a database query for metrics
func (dm *DatabaseManager) RecordQuery(operation, table string, duration time.Duration) {
	dm.dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dm.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordCacheHit records a cache hit
func (dm *DatabaseManager) RecordCacheHit(duration time.Duration) {
	dm.cacheHitsTotal.Inc()
	dm.cacheOperationDuration.Observe(duration.Seconds())
}

// RecordCacheMiss records a cache miss
func (dm *DatabaseManager) RecordCacheMiss(duration time.Duration) {
	dm.cacheMissesTotal.Inc()
	dm.cacheOperationDuration.Observe(duration.Seconds())
}

// RecordRedisCommand records a Redis command execution
func (dm *DatabaseManager) RecordRedisCommand(command string, duration time.Duration) {
	dm.redisCommandsTotal.WithLabelValues(command).Inc()
	dm.redisCommandDuration.WithLabelValues(command).Observe(duration.Seconds())
}

// Health returns the health status of database connections
func (dm *DatabaseManager) Health(ctx context.Context) (*HealthStatus, error) {
	health := &HealthStatus{
		PostgreSQL: &ComponentHealth{Status: "healthy"},
		Redis:      &ComponentHealth{Status: "healthy"},
	}

	// Check PostgreSQL
	pgStart := time.Now()
	if err := dm.DB.PingContext(ctx); err != nil {
		health.PostgreSQL.Status = "unhealthy"
		health.PostgreSQL.Error = err.Error()
	}
	health.PostgreSQL.ResponseTime = time.Since(pgStart)

	// Check Redis
	redisStart := time.Now()
	if err := dm.Redis.Ping(ctx).Err(); err != nil {
		health.Redis.Status = "unhealthy"
		health.Redis.Error = err.Error()
	}
	health.Redis.ResponseTime = time.Since(redisStart)

	// Overall status
	if health.PostgreSQL.Status == "healthy" && health.Redis.Status == "healthy" {
		health.Overall = "healthy"
	} else {
		health.Overall = "degraded"
	}

	return health, nil
}

// Stats returns database connection statistics
func (dm *DatabaseManager) Stats() *DatabaseStats {
	dbStats := dm.DB.Stats()
	
	return &DatabaseStats{
		PostgreSQL: &PostgreSQLStats{
			OpenConnections:     dbStats.OpenConnections,
			InUse:              dbStats.InUse,
			Idle:               dbStats.Idle,
			WaitCount:          dbStats.WaitCount,
			WaitDuration:       dbStats.WaitDuration,
			MaxIdleClosed:      dbStats.MaxIdleClosed,
			MaxLifetimeClosed:  dbStats.MaxLifetimeClosed,
			MaxOpenConnections: dm.config.MaxOpenConns,
			MaxIdleConnections: dm.config.MaxIdleConns,
		},
		Redis: &RedisStats{
			PoolSize:     dm.config.RedisPoolSize,
			MinIdleConns: dm.config.RedisMinIdleConns,
		},
	}
}

// Close gracefully closes all database connections
func (dm *DatabaseManager) Close() error {
	var errors []error

	// Stop metrics collection goroutine
	if dm.metricsCancel != nil {
		dm.metricsCancel()
		dm.logger.Info("Metrics collection stopped")
	}

	// Close PostgreSQL connection
	if dm.DB != nil {
		if err := dm.DB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	// Close Redis connection
	if dm.Redis != nil {
		if err := dm.Redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errors)
	}

	dm.logger.Info("Database connections closed successfully")
	return nil
}

// WithTransaction executes a function within a database transaction
func (dm *DatabaseManager) WithTransaction(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := dm.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// SetUserContext sets the current user ID for audit logging
func (dm *DatabaseManager) SetUserContext(ctx context.Context, userID string) context.Context {
	_, err := dm.DB.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, false)", userID)
	if err != nil {
		dm.logger.Warn("Failed to set user context for audit logging", "error", err)
	}
	return ctx
}

// Health and stats types
type HealthStatus struct {
	Overall    string           `json:"overall"`
	PostgreSQL *ComponentHealth `json:"postgresql"`
	Redis      *ComponentHealth `json:"redis"`
}

type ComponentHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

type DatabaseStats struct {
	PostgreSQL *PostgreSQLStats `json:"postgresql"`
	Redis      *RedisStats      `json:"redis"`
}

type PostgreSQLStats struct {
	OpenConnections     int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
	MaxOpenConnections int           `json:"max_open_connections"`
	MaxIdleConnections int           `json:"max_idle_connections"`
}

type RedisStats struct {
	PoolSize     int `json:"pool_size"`
	MinIdleConns int `json:"min_idle_conns"`
}