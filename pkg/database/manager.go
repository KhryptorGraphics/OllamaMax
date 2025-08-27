package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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
}

// NewDatabaseManager creates a new database manager with all repositories
func NewDatabaseManager(config *DatabaseConfig, logger *slog.Logger) (*DatabaseManager, error) {
	// Set defaults
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
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
	dm.Models = NewModelRepository(dm.DB, dm.Redis, dm.logger)
	dm.Nodes = NewNodeRepository(dm.DB, dm.Redis, dm.logger)
	dm.Users = NewUserRepository(dm.DB, dm.Redis, dm.logger)
	dm.Sessions = NewSessionRepository(dm.DB, dm.Redis, dm.logger)
	dm.Inference = NewInferenceRepository(dm.DB, dm.Redis, dm.logger)
	dm.Audit = NewAuditRepository(dm.DB, dm.logger)
	dm.Config = NewConfigRepository(dm.DB, dm.Redis, dm.logger)
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