package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Testing database configuration...")

	// Create database configuration using the same logic as main.go
	dbConfig, err := createDatabaseConfig(logger)
	if err != nil {
		logger.Error("Failed to create database configuration", "error", err)
		os.Exit(1)
	}

	// Test database connection
	logger.Info("Attempting to connect to database...")
	db, err := database.NewDatabaseManager(dbConfig, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test database health
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := db.Health(ctx)
	if err != nil {
		logger.Error("Failed to check database health", "error", err)
		os.Exit(1)
	}

	// Print health status
	logger.Info("Database health check completed",
		"overall", health.Overall,
		"postgresql_status", health.PostgreSQL.Status,
		"postgresql_response_time", health.PostgreSQL.ResponseTime,
		"redis_status", health.Redis.Status,
		"redis_response_time", health.Redis.ResponseTime,
	)

	if health.Overall == "healthy" {
		logger.Info("✅ Database configuration test PASSED - all connections healthy")
	} else {
		logger.Error("❌ Database configuration test FAILED - unhealthy connections detected")
		if health.PostgreSQL.Error != "" {
			logger.Error("PostgreSQL error", "error", health.PostgreSQL.Error)
		}
		if health.Redis.Error != "" {
			logger.Error("Redis error", "error", health.Redis.Error)
		}
		os.Exit(1)
	}

	// Test database stats
	stats := db.Stats()
	logger.Info("Database connection statistics",
		"postgresql_open_connections", stats.PostgreSQL.OpenConnections,
		"postgresql_in_use", stats.PostgreSQL.InUse,
		"postgresql_idle", stats.PostgreSQL.Idle,
		"redis_pool_size", stats.Redis.PoolSize,
		"redis_min_idle_conns", stats.Redis.MinIdleConns,
	)

	logger.Info("Database configuration test completed successfully!")
}

// createDatabaseConfig creates a database configuration with proper environment variable handling
// This function is duplicated from main.go for testing purposes
func createDatabaseConfig(logger *slog.Logger) (*database.DatabaseConfig, error) {
	// Helper function to get integer environment variable with default
	getEnvInt := func(key string, defaultValue int) (int, error) {
		if value := os.Getenv(key); value != "" {
			if intValue, err := strconv.Atoi(value); err != nil {
				return 0, fmt.Errorf("invalid integer value for %s: %s", key, value)
			} else {
				return intValue, nil
			}
		}
		return defaultValue, nil
	}

	// Parse database port
	dbPort, err := getEnvInt("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}

	// Parse Redis port
	redisPort, err := getEnvInt("REDIS_PORT", 6379)
	if err != nil {
		return nil, err
	}

	config := &database.DatabaseConfig{
		// PostgreSQL configuration
		Host:     getEnvWithDefault("DB_HOST", "localhost"),
		Port:     dbPort,
		Name:     getEnvWithDefault("DB_NAME", "ollamamax"),
		User:     getEnvWithDefault("DB_USER", "ollama"),
		Password: os.Getenv("DB_PASSWORD"), // No default for password - should be explicitly set
		SSLMode:  getEnvWithDefault("DB_SSL_MODE", "prefer"),
		
		// Redis configuration
		RedisHost:     getEnvWithDefault("REDIS_HOST", "localhost"),
		RedisPort:     redisPort,
		RedisPassword: os.Getenv("REDIS_PASSWORD"), // No default for password - should be explicitly set
		RedisDB:       0, // Default Redis DB
	}

	// Validate critical configuration
	if config.Password == "" {
		logger.Warn("DB_PASSWORD environment variable not set - using empty password")
	}
	
	// Log configuration (without sensitive data)
	logger.Info("Database configuration created",
		"host", config.Host,
		"port", config.Port,
		"database", config.Name,
		"user", config.User,
		"ssl_mode", config.SSLMode,
		"redis_host", config.RedisHost,
		"redis_port", config.RedisPort,
		"redis_auth", config.RedisPassword != "",
	)

	return config, nil
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}