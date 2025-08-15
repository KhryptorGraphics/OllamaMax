package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	var configPath = flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		// Use default config if file doesn't exist
		cfg = &config.DistributedConfig{}
		cfg.SetDefaults()
		cfg.Node.ID = "test-node"
		cfg.Node.Name = "Test Node"
		cfg.API.Host = "0.0.0.0"
		cfg.API.Port = 8080
		log.Printf("Using default configuration")
	}

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create and start server
	srv, err := server.NewDistributedServer(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}()

	logger.Infof("OllamaMax Distributed started on %s:%d", cfg.API.Host, cfg.API.Port)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Shutdown error: %v", err)
	}

	logger.Info("Server stopped")
}
