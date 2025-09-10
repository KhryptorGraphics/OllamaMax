package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/internal/server"
	"github.com/khryptorgraphics/ollamamax/pkg/auth"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

// createDatabaseConfig creates a database configuration with proper environment variable handling
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
	dbPort, err := getEnvInt("DB_PORT", 15432)
	if err != nil {
		return nil, err
	}

	// Parse Redis port
	redisPort, err := getEnvInt("REDIS_PORT", 16379)
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

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	logger.Info("Starting OllamaMax distributed inference platform")
	
	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Configuration loaded", "listen_addr", cfg.API.Listen)
	
	// Initialize database with environment variables and proper error handling
	dbConfig, err := createDatabaseConfig(logger)
	if err != nil {
		logger.Error("Failed to create database configuration", "error", err)
		os.Exit(1)
	}
	
	db, err := database.NewDatabaseManager(dbConfig, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		logger.Warn("Continuing without database - some features will be unavailable")
		db = nil
	} else {
		defer db.Close()
		logger.Info("Database initialized successfully")
	}
	
	// Initialize authentication service
	jwtService, err := auth.NewJWTService(&cfg.Auth)
	if err != nil {
		logger.Error("Failed to initialize JWT service", "error", err)
		os.Exit(1)
	}
	
	logger.Info("JWT service initialized")
	_ = jwtService // Avoid unused variable error
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-signalChan
		logger.Info("Shutdown signal received, initiating graceful shutdown")
		cancel()
	}()
	
	// Initialize API server
	logger.Info("Initializing API server...")
	apiServer := server.NewServer(cfg, db, logger)
	
	logger.Info("API server initialized successfully")
	
	// Start health check routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if db != nil {
					health, err := db.Health(ctx)
					if err != nil {
						logger.Error("Health check failed", "error", err)
						continue
					}
					logger.Info("System health check", "status", health.Overall)
				} else {
					logger.Info("System health check", "status", "degraded", "database", "unavailable")
				}
			}
		}
	}()
	
	logger.Info("OllamaMax server starting",
		"version", "1.0.0",
		"api_listen", cfg.API.Listen)
	
	// Start the server (this blocks until shutdown)
	if err := apiServer.Start(ctx); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
	
	logger.Info("OllamaMax server stopped")
	fmt.Println("Server stopped successfully")
}
