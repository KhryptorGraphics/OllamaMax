package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/rs/zerolog/log"
)

// Server represents a metrics server
type Server struct {
	config *config.MetricsConfig
	server *http.Server
}

// NewServer creates a new metrics server
func NewServer(config config.MetricsConfig) (*Server, error) {
	mux := http.NewServeMux()

	// Add basic metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "# Ollamacron Metrics\n")
		fmt.Fprintf(w, "# TYPE ollamacron_info gauge\n")
		fmt.Fprintf(w, "ollamacron_info{version=\"dev\"} 1\n")
		fmt.Fprintf(w, "# TYPE ollamacron_uptime_seconds counter\n")
		fmt.Fprintf(w, "ollamacron_uptime_seconds %d\n", time.Now().Unix())
	})

	// Add health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	})

	server := &http.Server{
		Addr:         config.Listen,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &Server{
		config: &config,
		server: server,
	}, nil
}

// Start starts the metrics server
func (s *Server) Start() error {
	log.Info().Str("address", s.config.Listen).Msg("Starting metrics server")

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Metrics server error")
		}
	}()

	return nil
}

// Shutdown shuts down the metrics server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info().Msg("Shutting down metrics server")
	return s.server.Shutdown(ctx)
}
