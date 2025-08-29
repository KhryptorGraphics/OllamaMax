package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/internal/config"
	"github.com/khryptorgraphics/ollamamax/pkg/auth"
	"github.com/khryptorgraphics/ollamamax/pkg/database"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	logger.Info("Starting OllamaMax distributed inference platform")
	
	// Load configuration
	cfg := config.LoadConfig()
	logger.Info("Configuration loaded", "listen_addr", cfg.API.Listen)
	
	// Initialize database
	dbConfig := &database.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     5432,
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  "prefer",
		RedisHost: os.Getenv("REDIS_HOST"),
		RedisPort: 6379,
	}
	
	// Set defaults if environment variables are not set
	if dbConfig.Host == "" {
		dbConfig.Host = "localhost"
	}
	if dbConfig.Name == "" {
		dbConfig.Name = "ollamamax"
	}
	if dbConfig.User == "" {
		dbConfig.User = "ollama"
	}
	if dbConfig.RedisHost == "" {
		dbConfig.RedisHost = "localhost"
	}
	
	db, err := database.NewDatabaseManager(dbConfig, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	
	logger.Info("Database initialized successfully")
	
	// Initialize authentication service
	_, err = auth.NewJWTService(&cfg.Auth)
	if err != nil {
		logger.Error("Failed to initialize JWT service", "error", err)
		os.Exit(1)
	}
	
	logger.Info("JWT service initialized")
	
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
	
	// Health check endpoint
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				health, err := db.Health(ctx)
				if err != nil {
					logger.Error("Health check failed", "error", err)
					continue
				}
				logger.Info("System health check", "status", health.Overall)
			}
		}
	}()
	
	logger.Info("OllamaMax server started", "version", "1.0.0", "port", cfg.API.Port)
	
	// Keep the application running
	<-ctx.Done()
	
	// Graceful shutdown
	logger.Info("Shutting down gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Cleanup resources
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded")
	default:
		logger.Info("Cleanup completed")
	}
	
	logger.Info("OllamaMax server stopped")
	fmt.Println("Server stopped successfully")
}