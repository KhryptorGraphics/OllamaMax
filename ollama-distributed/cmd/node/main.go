package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ollama/ollama-distributed/internal/config"
	"github.com/ollama/ollama-distributed/pkg/api"
	"github.com/ollama/ollama-distributed/pkg/consensus"
	"github.com/ollama/ollama-distributed/pkg/p2p"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ollama-distributed",
		Short: "Distributed Ollama Platform",
		Long: `A distributed, enterprise-grade version of Ollama that transforms 
the single-node architecture into a horizontally scalable, fault-tolerant platform.`,
		Version: version,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ollama-distributed.yaml)")
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(joinCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a distributed Ollama node",
		Long:  "Start a distributed Ollama node with P2P networking and consensus",
		RunE:  runStart,
	}

	cmd.Flags().String("listen", "0.0.0.0:11434", "Address to listen on")
	cmd.Flags().String("p2p-listen", "0.0.0.0:4001", "P2P listen address")
	cmd.Flags().StringSlice("bootstrap", []string{}, "Bootstrap peers")
	cmd.Flags().String("data-dir", "./data", "Data directory")
	cmd.Flags().Bool("enable-web", true, "Enable web control panel")
	cmd.Flags().String("web-listen", "0.0.0.0:8080", "Web panel listen address")

	return cmd
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show node status",
		Long:  "Show the current status of the distributed Ollama node",
		RunE:  runStatus,
	}
}

func joinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join an existing cluster",
		Long:  "Join an existing distributed Ollama cluster",
		RunE:  runJoin,
	}

	cmd.Flags().StringSlice("peers", []string{}, "Peer addresses to join")
	cmd.MarkFlagRequired("peers")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	if listen, _ := cmd.Flags().GetString("listen"); listen != "" {
		cfg.API.Listen = listen
	}
	if p2pListen, _ := cmd.Flags().GetString("p2p-listen"); p2pListen != "" {
		cfg.P2P.Listen = p2pListen
	}
	if bootstrap, _ := cmd.Flags().GetStringSlice("bootstrap"); len(bootstrap) > 0 {
		cfg.P2P.Bootstrap = bootstrap
	}
	if dataDir, _ := cmd.Flags().GetString("data-dir"); dataDir != "" {
		cfg.Storage.DataDir = dataDir
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize P2P networking
	p2pNode, err := p2p.NewNode(ctx, cfg.P2P)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Initialize consensus engine
	consensusEngine, err := consensus.NewEngine(cfg.Consensus, p2pNode)
	if err != nil {
		return fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Initialize scheduler
	schedulerEngine, err := scheduler.NewEngine(cfg.Scheduler, p2pNode, consensusEngine)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Initialize API server
	apiServer, err := api.NewServer(cfg.API, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// Start all services
	if err := p2pNode.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	if err := consensusEngine.Start(); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	if err := schedulerEngine.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	if err := apiServer.Start(); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	log.Printf("Distributed Ollama node started successfully")
	log.Printf("API server listening on: %s", cfg.API.Listen)
	log.Printf("P2P node listening on: %s", cfg.P2P.Listen)
	log.Printf("Node ID: %s", p2pNode.ID())

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}

	if err := schedulerEngine.Shutdown(shutdownCtx); err != nil {
		log.Printf("Scheduler shutdown error: %v", err)
	}

	if err := consensusEngine.Shutdown(shutdownCtx); err != nil {
		log.Printf("Consensus engine shutdown error: %v", err)
	}

	if err := p2pNode.Shutdown(shutdownCtx); err != nil {
		log.Printf("P2P node shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	// TODO: Implement status command
	// This would connect to a running node and display its status
	fmt.Println("Status command not yet implemented")
	return nil
}

func runJoin(cmd *cobra.Command, args []string) error {
	// TODO: Implement join command
	// This would connect to existing peers and join the cluster
	fmt.Println("Join command not yet implemented")
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ollama-distributed")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}