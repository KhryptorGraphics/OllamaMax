package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Manager handles database operations
type Manager struct {
	db     *sql.DB
	config *Config
}

// Config holds database configuration
type Config struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	SSLMode  string `yaml:"ssl_mode" json:"ssl_mode"`
	
	// Connection pool settings
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            15432, // Non-standard port as requested
		Database:        "ollamamax",
		Username:        "ollamamax",
		Password:        "ollamamax_secure_password",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// NewManager creates a new database manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", config.Host).
		Int("port", config.Port).
		Str("database", config.Database).
		Msg("Connected to database")

	return &Manager{
		db:     db,
		config: config,
	}, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// Health checks database health
func (m *Manager) Health(ctx context.Context) error {
	return m.db.PingContext(ctx)
}

// GetDB returns the underlying database connection
func (m *Manager) GetDB() *sql.DB {
	return m.db
}

// BeginTx starts a new transaction
func (m *Manager) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return m.db.BeginTx(ctx, nil)
}

// ExecuteInTransaction executes a function within a transaction
func (m *Manager) ExecuteInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := m.BeginTx(ctx)
	if err != nil {
		return err
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

// MigrationsComplete checks if all migrations have been applied
func (m *Manager) MigrationsComplete() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if migration table exists
	var exists bool
	query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')`
	err := m.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil || !exists {
		return false
	}

	// Check if we have any migrations applied
	var count int
	query = `SELECT COUNT(*) FROM schema_migrations`
	err = m.db.QueryRowContext(ctx, query).Scan(&count)
	return err == nil && count > 0
}

// GetStats returns database connection statistics
func (m *Manager) GetStats() sql.DBStats {
	return m.db.Stats()
}
