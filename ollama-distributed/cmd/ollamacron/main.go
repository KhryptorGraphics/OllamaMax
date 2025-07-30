package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/metrics"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/models"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Build information - set during build
var (
	version   = "dev"
	commit    = "unknown"
	date      = "unknown"
	goVersion = runtime.Version()
)

// Application state
type Application struct {
	Config          *config.Config
	P2PNode         *p2p.Node
	ConsensusEngine *consensus.Engine
	SchedulerEngine *scheduler.Engine
	ModelManager    *models.DistributedModelManager
	APIServer       *api.Server
	MetricsServer   *metrics.Server
	Logger          zerolog.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

func main() {
	// Initialize application
	app := &Application{}
	app.ctx, app.cancel = context.WithCancel(context.Background())

	// Build root command
	rootCmd := &cobra.Command{
		Use:   "ollamacron",
		Short: "Ollamacron - Distributed Ollama Platform",
		Long: `Ollamacron is a distributed, enterprise-grade version of Ollama that transforms 
the single-node architecture into a horizontally scalable, fault-tolerant platform.

Features:
• Peer-to-peer networking with automatic discovery
• Distributed consensus and coordination
• Intelligent load balancing and scheduling
• Model synchronization across nodes
• Advanced security and encryption
• Real-time monitoring and metrics
• Web-based management interface`,
		Version: buildVersion(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return app.initializeLogging()
		},
		SilenceUsage: true,
	}

	// Add global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default: ./config/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "json", "log format (json, console)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")

	// Add subcommands
	rootCmd.AddCommand(
		buildNodeCmd(app),
		buildCoordinatorCmd(app),
		buildStandaloneCmd(app),
		buildStatusCmd(app),
		buildJoinCmd(app),
		buildVersionCmd(),
		buildConfigCmd(app),
		buildHealthCmd(app),
		buildMetricsCmd(app),
	)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

// buildNodeCmd creates the node subcommand
func buildNodeCmd(app *Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Start as a distributed node",
		Long:  "Start Ollamacron as a distributed node that can join an existing cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runNode(cmd, args)
		},
	}

	cmd.Flags().String("listen", "0.0.0.0:11434", "API server listen address")
	cmd.Flags().String("p2p-listen", "/ip4/0.0.0.0/tcp/4001", "P2P listen address")
	cmd.Flags().StringSlice("bootstrap", []string{}, "Bootstrap peer addresses")
	cmd.Flags().String("data-dir", "./data", "Data directory")
	cmd.Flags().String("model-dir", "./models", "Model directory")
	cmd.Flags().Bool("enable-web", true, "Enable web interface")
	cmd.Flags().String("web-listen", "0.0.0.0:8080", "Web interface listen address")
	cmd.Flags().Bool("enable-metrics", true, "Enable metrics collection")
	cmd.Flags().String("metrics-listen", "0.0.0.0:9090", "Metrics server listen address")
	cmd.Flags().String("node-id", "", "Node ID (auto-generated if empty)")
	cmd.Flags().String("node-name", "", "Node name (hostname if empty)")
	cmd.Flags().String("region", "", "Node region")
	cmd.Flags().String("zone", "", "Node zone")

	return cmd
}

// buildCoordinatorCmd creates the coordinator subcommand
func buildCoordinatorCmd(app *Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coordinator",
		Short: "Start as a cluster coordinator",
		Long:  "Start Ollamacron as a cluster coordinator that manages other nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runCoordinator(cmd, args)
		},
	}

	cmd.Flags().String("listen", "0.0.0.0:11434", "API server listen address")
	cmd.Flags().String("p2p-listen", "/ip4/0.0.0.0/tcp/4001", "P2P listen address")
	cmd.Flags().String("consensus-listen", "0.0.0.0:7000", "Consensus server listen address")
	cmd.Flags().String("data-dir", "./data", "Data directory")
	cmd.Flags().Bool("bootstrap", false, "Bootstrap new cluster")
	cmd.Flags().Bool("enable-web", true, "Enable web interface")
	cmd.Flags().String("web-listen", "0.0.0.0:8080", "Web interface listen address")
	cmd.Flags().Bool("enable-metrics", true, "Enable metrics collection")
	cmd.Flags().String("metrics-listen", "0.0.0.0:9090", "Metrics server listen address")

	return cmd
}

// buildStandaloneCmd creates the standalone subcommand
func buildStandaloneCmd(app *Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "standalone",
		Short: "Start in standalone mode",
		Long:  "Start Ollamacron in standalone mode (single node, no clustering)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runStandalone(cmd, args)
		},
	}

	cmd.Flags().String("listen", "0.0.0.0:11434", "API server listen address")
	cmd.Flags().String("data-dir", "./data", "Data directory")
	cmd.Flags().String("model-dir", "./models", "Model directory")
	cmd.Flags().Bool("enable-web", true, "Enable web interface")
	cmd.Flags().String("web-listen", "0.0.0.0:8080", "Web interface listen address")
	cmd.Flags().Bool("enable-metrics", true, "Enable metrics collection")
	cmd.Flags().String("metrics-listen", "0.0.0.0:9090", "Metrics server listen address")

	return cmd
}

// buildStatusCmd creates the status subcommand
func buildStatusCmd(app *Application) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show node status",
		Long:  "Show the current status of the Ollamacron node",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runStatus(cmd, args)
		},
	}
}

// buildJoinCmd creates the join subcommand
func buildJoinCmd(app *Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join an existing cluster",
		Long:  "Join an existing Ollamacron cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runJoin(cmd, args)
		},
	}

	cmd.Flags().StringSlice("peers", []string{}, "Peer addresses to join")
	cmd.MarkFlagRequired("peers")

	return cmd
}

// buildVersionCmd creates the version subcommand
func buildVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Show detailed version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Ollamacron %s\n", buildVersion())
			fmt.Printf("  Version: %s\n", version)
			fmt.Printf("  Commit: %s\n", commit)
			fmt.Printf("  Date: %s\n", date)
			fmt.Printf("  Go version: %s\n", goVersion)
			fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			
			if info, ok := debug.ReadBuildInfo(); ok {
				fmt.Printf("  Module: %s\n", info.Main.Path)
				if info.Main.Version != "" {
					fmt.Printf("  Module version: %s\n", info.Main.Version)
				}
			}
		},
	}
}

// buildConfigCmd creates the config subcommand
func buildConfigCmd(app *Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Generate and manage configuration files",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate default configuration",
		Long:  "Generate a default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runConfigGenerate(cmd, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		Long:  "Validate the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runConfigValidate(cmd, args)
		},
	})

	return cmd
}

// buildHealthCmd creates the health subcommand
func buildHealthCmd(app *Application) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check system health",
		Long:  "Check the health of the Ollamacron system",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runHealth(cmd, args)
		},
	}
}

// buildMetricsCmd creates the metrics subcommand
func buildMetricsCmd(app *Application) *cobra.Command {
	return &cobra.Command{
		Use:   "metrics",
		Short: "Show system metrics",
		Long:  "Show current system metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.runMetrics(cmd, args)
		},
	}
}

// initializeLogging initializes the logging system
func (app *Application) initializeLogging() error {
	// Get log level from flags
	logLevel := viper.GetString("log-level")
	logFormat := viper.GetString("log-format")
	debug := viper.GetBool("debug")

	// Set log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	
	if debug {
		level = zerolog.DebugLevel
	}

	// Configure logger
	zerolog.SetGlobalLevel(level)
	
	if logFormat == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Set as application logger
	app.Logger = log.With().Str("component", "ollamacron").Logger()

	return nil
}

// loadConfig loads the configuration
func (app *Application) loadConfig(configFile string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app.Config = cfg
	app.Logger.Info().
		Str("config_file", viper.ConfigFileUsed()).
		Msg("Configuration loaded")

	return nil
}

// runNode runs the node mode
func (app *Application) runNode(cmd *cobra.Command, args []string) error {
	// Load configuration
	configFile, _ := cmd.Flags().GetString("config")
	if err := app.loadConfig(configFile); err != nil {
		return err
	}

	// Override config with CLI flags
	app.overrideConfigFromFlags(cmd)

	app.Logger.Info().
		Str("mode", "node").
		Str("version", version).
		Msg("Starting Ollamacron")

	// Initialize and start services
	if err := app.initializeServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := app.startServices(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for shutdown signal
	return app.waitForShutdown()
}

// runCoordinator runs the coordinator mode
func (app *Application) runCoordinator(cmd *cobra.Command, args []string) error {
	// Load configuration
	configFile, _ := cmd.Flags().GetString("config")
	if err := app.loadConfig(configFile); err != nil {
		return err
	}

	// Override config with CLI flags
	app.overrideConfigFromFlags(cmd)

	// Set coordinator-specific settings
	app.Config.Consensus.Bootstrap = true

	app.Logger.Info().
		Str("mode", "coordinator").
		Str("version", version).
		Msg("Starting Ollamacron")

	// Initialize and start services
	if err := app.initializeServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := app.startServices(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for shutdown signal
	return app.waitForShutdown()
}

// runStandalone runs the standalone mode
func (app *Application) runStandalone(cmd *cobra.Command, args []string) error {
	// Load configuration
	configFile, _ := cmd.Flags().GetString("config")
	if err := app.loadConfig(configFile); err != nil {
		return err
	}

	// Override config with CLI flags
	app.overrideConfigFromFlags(cmd)

	// Disable clustering features
	app.Config.P2P.EnableDHT = false
	app.Config.P2P.EnablePubSub = false
	app.Config.Consensus.Bootstrap = false

	app.Logger.Info().
		Str("mode", "standalone").
		Str("version", version).
		Msg("Starting Ollamacron")

	// Initialize and start services (without clustering)
	if err := app.initializeStandaloneServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := app.startStandaloneServices(); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}

	// Wait for shutdown signal
	return app.waitForShutdown()
}

// runStatus runs the status command
func (app *Application) runStatus(cmd *cobra.Command, args []string) error {
	// TODO: Implement status checking
	// This would connect to a running node and display its status
	fmt.Println("Status command not yet implemented")
	return nil
}

// runJoin runs the join command
func (app *Application) runJoin(cmd *cobra.Command, args []string) error {
	// TODO: Implement join functionality
	// This would connect to existing peers and join the cluster
	fmt.Println("Join command not yet implemented")
	return nil
}

// runConfigGenerate generates a default configuration file
func (app *Application) runConfigGenerate(cmd *cobra.Command, args []string) error {
	cfg := config.DefaultConfig()
	
	// Generate config file
	filename := "config.yaml"
	if len(args) > 0 {
		filename = args[0]
	}

	if err := cfg.Save(filename); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Generated configuration file: %s\n", filename)
	return nil
}

// runConfigValidate validates the configuration
func (app *Application) runConfigValidate(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	if err := app.loadConfig(configFile); err != nil {
		return err
	}

	fmt.Println("Configuration is valid")
	return nil
}

// runHealth checks system health
func (app *Application) runHealth(cmd *cobra.Command, args []string) error {
	// TODO: Implement health checking
	fmt.Println("Health check not yet implemented")
	return nil
}

// runMetrics shows system metrics
func (app *Application) runMetrics(cmd *cobra.Command, args []string) error {
	// TODO: Implement metrics display
	fmt.Println("Metrics display not yet implemented")
	return nil
}

// overrideConfigFromFlags overrides configuration with CLI flags
func (app *Application) overrideConfigFromFlags(cmd *cobra.Command) {
	// API settings
	if listen, _ := cmd.Flags().GetString("listen"); listen != "" {
		app.Config.API.Listen = listen
	}

	// P2P settings
	if p2pListen, _ := cmd.Flags().GetString("p2p-listen"); p2pListen != "" {
		app.Config.P2P.Listen = p2pListen
	}
	if bootstrap, _ := cmd.Flags().GetStringSlice("bootstrap"); len(bootstrap) > 0 {
		app.Config.P2P.Bootstrap = bootstrap
	}

	// Storage settings
	if dataDir, _ := cmd.Flags().GetString("data-dir"); dataDir != "" {
		app.Config.Storage.DataDir = dataDir
	}
	if modelDir, _ := cmd.Flags().GetString("model-dir"); modelDir != "" {
		app.Config.Storage.ModelDir = modelDir
	}

	// Web settings
	if webListen, _ := cmd.Flags().GetString("web-listen"); webListen != "" {
		app.Config.Web.Listen = webListen
	}
	if enableWeb, _ := cmd.Flags().GetBool("enable-web"); cmd.Flags().Changed("enable-web") {
		app.Config.Web.Enabled = enableWeb
	}

	// Metrics settings
	if metricsListen, _ := cmd.Flags().GetString("metrics-listen"); metricsListen != "" {
		app.Config.Metrics.Listen = metricsListen
	}
	if enableMetrics, _ := cmd.Flags().GetBool("enable-metrics"); cmd.Flags().Changed("enable-metrics") {
		app.Config.Metrics.Enabled = enableMetrics
	}

	// Node settings
	if nodeID, _ := cmd.Flags().GetString("node-id"); nodeID != "" {
		app.Config.Node.ID = nodeID
	}
	if nodeName, _ := cmd.Flags().GetString("node-name"); nodeName != "" {
		app.Config.Node.Name = nodeName
	}
	if region, _ := cmd.Flags().GetString("region"); region != "" {
		app.Config.Node.Region = region
	}
	if zone, _ := cmd.Flags().GetString("zone"); zone != "" {
		app.Config.Node.Zone = zone
	}

	// Consensus settings
	if consensusListen, _ := cmd.Flags().GetString("consensus-listen"); consensusListen != "" {
		app.Config.Consensus.BindAddr = consensusListen
	}
	if bootstrap, _ := cmd.Flags().GetBool("bootstrap"); cmd.Flags().Changed("bootstrap") {
		app.Config.Consensus.Bootstrap = bootstrap
	}
}

// initializeServices initializes all services
func (app *Application) initializeServices() error {
	var err error

	app.Logger.Info().Msg("Initializing services...")

	// Initialize security
	if err := security.Initialize(app.Config.Security); err != nil {
		return fmt.Errorf("failed to initialize security: %w", err)
	}

	// Initialize P2P networking
	app.P2PNode, err = p2p.NewNode(app.ctx, app.Config.P2P)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Initialize consensus engine
	app.ConsensusEngine, err = consensus.NewEngine(app.Config.Consensus, app.P2PNode)
	if err != nil {
		return fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Initialize model manager
	app.ModelManager, err = models.NewDistributedModelManager(app.Config.Storage, app.P2PNode)
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	// Initialize scheduler
	app.SchedulerEngine, err = scheduler.NewEngine(app.Config.Scheduler, app.P2PNode, app.ConsensusEngine)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Initialize API server
	app.APIServer, err = api.NewServer(app.Config.API, app.P2PNode, app.ConsensusEngine, app.SchedulerEngine)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// Initialize metrics server
	if app.Config.Metrics.Enabled {
		app.MetricsServer, err = metrics.NewServer(app.Config.Metrics)
		if err != nil {
			return fmt.Errorf("failed to create metrics server: %w", err)
		}
	}

	app.Logger.Info().Msg("Services initialized successfully")
	return nil
}

// initializeStandaloneServices initializes services for standalone mode
func (app *Application) initializeStandaloneServices() error {
	var err error

	app.Logger.Info().Msg("Initializing standalone services...")

	// Initialize security
	if err := security.Initialize(app.Config.Security); err != nil {
		return fmt.Errorf("failed to initialize security: %w", err)
	}

	// Initialize model manager (local only)
	app.ModelManager, err = models.NewDistributedModelManager(app.Config.Storage, nil)
	if err != nil {
		return fmt.Errorf("failed to create model manager: %w", err)
	}

	// Initialize scheduler (local only)
	app.SchedulerEngine, err = scheduler.NewEngine(app.Config.Scheduler, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Initialize API server (without clustering)
	app.APIServer, err = api.NewServer(app.Config.API, nil, nil, app.SchedulerEngine)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// Initialize metrics server
	if app.Config.Metrics.Enabled {
		app.MetricsServer, err = metrics.NewServer(app.Config.Metrics)
		if err != nil {
			return fmt.Errorf("failed to create metrics server: %w", err)
		}
	}

	app.Logger.Info().Msg("Standalone services initialized successfully")
	return nil
}

// startServices starts all services
func (app *Application) startServices() error {
	app.Logger.Info().Msg("Starting services...")

	// Start P2P node
	if err := app.P2PNode.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Start consensus engine
	if err := app.ConsensusEngine.Start(); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}

	// Start model manager
	if err := app.ModelManager.Start(); err != nil {
		return fmt.Errorf("failed to start model manager: %w", err)
	}

	// Start scheduler
	if err := app.SchedulerEngine.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start API server
	if err := app.APIServer.Start(); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	// Start metrics server
	if app.MetricsServer != nil {
		if err := app.MetricsServer.Start(); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
	}

	app.Logger.Info().
		Str("api_listen", app.Config.API.Listen).
		Str("p2p_listen", app.Config.P2P.Listen).
		Str("node_id", app.P2PNode.ID()).
		Msg("All services started successfully")

	return nil
}

// startStandaloneServices starts services for standalone mode
func (app *Application) startStandaloneServices() error {
	app.Logger.Info().Msg("Starting standalone services...")

	// Start model manager
	if err := app.ModelManager.Start(); err != nil {
		return fmt.Errorf("failed to start model manager: %w", err)
	}

	// Start scheduler
	if err := app.SchedulerEngine.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// Start API server
	if err := app.APIServer.Start(); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}

	// Start metrics server
	if app.MetricsServer != nil {
		if err := app.MetricsServer.Start(); err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
	}

	app.Logger.Info().
		Str("api_listen", app.Config.API.Listen).
		Msg("All standalone services started successfully")

	return nil
}

// waitForShutdown waits for shutdown signal and performs graceful shutdown
func (app *Application) waitForShutdown() error {
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	app.Logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")

	// Perform graceful shutdown
	return app.shutdown()
}

// shutdown performs graceful shutdown of all services
func (app *Application) shutdown() error {
	app.Logger.Info().Msg("Shutting down...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown services in reverse order
	var shutdownErrors []error

	// Stop metrics server
	if app.MetricsServer != nil {
		if err := app.MetricsServer.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("metrics server shutdown: %w", err))
		}
	}

	// Stop API server
	if app.APIServer != nil {
		if err := app.APIServer.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("API server shutdown: %w", err))
		}
	}

	// Stop scheduler
	if app.SchedulerEngine != nil {
		if err := app.SchedulerEngine.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("scheduler shutdown: %w", err))
		}
	}

	// Stop model manager
	if app.ModelManager != nil {
		if err := app.ModelManager.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("model manager shutdown: %w", err))
		}
	}

	// Stop consensus engine
	if app.ConsensusEngine != nil {
		if err := app.ConsensusEngine.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("consensus engine shutdown: %w", err))
		}
	}

	// Stop P2P node
	if app.P2PNode != nil {
		if err := app.P2PNode.Shutdown(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("P2P node shutdown: %w", err))
		}
	}

	// Cancel application context
	app.cancel()

	// Log shutdown errors if any
	if len(shutdownErrors) > 0 {
		for _, err := range shutdownErrors {
			app.Logger.Error().Err(err).Msg("Shutdown error")
		}
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	}

	app.Logger.Info().Msg("Shutdown completed successfully")
	return nil
}

// buildVersion returns the build version string
func buildVersion() string {
	if version == "dev" {
		return "dev"
	}
	return fmt.Sprintf("v%s", version)
}

// init initializes the application
func init() {
	// Configure viper
	viper.SetEnvPrefix("OLLAMACRON")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set default config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.ollamacron")
	viper.AddConfigPath("/etc/ollamacron")

	// Set default config name
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
}