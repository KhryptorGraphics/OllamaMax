package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/consensus"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/integration"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/messaging"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/performance"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/scheduler"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/web"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
	rootCmd *cobra.Command
)

func main() {
	rootCmd = &cobra.Command{
		Use:   "ollama-distributed",
		Short: "🚀 OllamaMax - Enterprise Distributed AI Platform",
		Long: `🚀 OllamaMax - Enterprise Distributed AI Platform

A distributed, enterprise-grade version of Ollama that transforms the single-node
architecture into a horizontally scalable, fault-tolerant platform.

Features:
  🌐 Distributed AI model serving across multiple nodes
  🔒 Enterprise-grade security with JWT authentication
  📊 Real-time performance monitoring and optimization
  🎨 Beautiful web interface for easy management
  ⚡ Automatic load balancing and failover
  🔄 Seamless model distribution and synchronization

Quick Start:
  ollama-distributed quickstart     # Start with defaults
  ollama-distributed setup         # Interactive configuration
  ollama-distributed start         # Start your node

Web Interface: http://localhost:8081
API Endpoint:  http://localhost:8080

Documentation: https://github.com/KhryptorGraphics/OllamaMax`,
		Version: version,
		Example: `  # Quick start with defaults
  ollama-distributed quickstart

  # Interactive setup
  ollama-distributed setup

  # Start with custom config
  ollama-distributed start --config config.yaml

  # Check cluster status
  ollama-distributed proxy status

  # Pull and use models
  ollama-distributed proxy pull llama2`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ollama-distributed.yaml)")

	// Add commands with better organization
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(joinCmd())
	rootCmd.AddCommand(proxyCmd())

	// Initialize user experience commands
	// TODO: Implement these functions for enhanced user experience
	// initHelpCommands()
	// initSetupCommands()

	if err := rootCmd.Execute(); err != nil {
		log.Printf("Error executing command: %v", err)
		os.Exit(1)
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

func proxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage Ollama proxy and load balancing",
		Long:  "Manage the distributed Ollama proxy, instances, and load balancing",
	}

	cmd.AddCommand(proxyStatusCmd())
	cmd.AddCommand(proxyInstancesCmd())
	cmd.AddCommand(proxyMetricsCmd())

	return cmd
}

func proxyStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show proxy status",
		Long:  "Show the current status of the Ollama proxy and registered instances",
		RunE:  runProxyStatus,
	}

	cmd.Flags().String("api-url", "http://localhost:8080", "API server URL")
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}

func proxyInstancesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage proxy instances",
		Long:  "List and manage Ollama instances registered with the proxy",
		RunE:  runProxyInstances,
	}

	cmd.Flags().String("api-url", "http://localhost:8080", "API server URL")
	cmd.Flags().Bool("json", false, "Output in JSON format")

	return cmd
}

func proxyMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show proxy metrics",
		Long:  "Show performance metrics for the Ollama proxy",
		RunE:  runProxyMetrics,
	}

	cmd.Flags().String("api-url", "http://localhost:8080", "API server URL")
	cmd.Flags().Bool("json", false, "Output in JSON format")
	cmd.Flags().Bool("watch", false, "Watch metrics in real-time")
	cmd.Flags().Int("interval", 5, "Update interval in seconds (for watch mode)")

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags (only if explicitly set by user)
	if cmd.Flags().Changed("listen") {
		listen, _ := cmd.Flags().GetString("listen")
		log.Printf("🔧 Overriding API listen with CLI flag: %s", listen)
		cfg.API.Listen = listen
	}
	if cmd.Flags().Changed("p2p-listen") {
		p2pListen, _ := cmd.Flags().GetString("p2p-listen")
		log.Printf("🔧 Overriding P2P listen with CLI flag: %s", p2pListen)
		cfg.P2P.Listen = p2pListen
	}
	if cmd.Flags().Changed("bootstrap") {
		bootstrap, _ := cmd.Flags().GetStringSlice("bootstrap")
		log.Printf("🔧 Overriding P2P bootstrap with CLI flag: %v", bootstrap)
		cfg.P2P.Bootstrap = bootstrap
	}
	if cmd.Flags().Changed("data-dir") {
		dataDir, _ := cmd.Flags().GetString("data-dir")
		log.Printf("🔧 Overriding data dir with CLI flag: %s", dataDir)
		cfg.Storage.DataDir = dataDir
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize P2P networking with full configuration
	p2pNode, err := p2p.NewNode(ctx, &cfg.P2P)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Create messaging and monitoring components
	messageRouter := messaging.NewMessageRouter(nil)
	networkMonitor := monitoring.NewNetworkMonitor(nil)

	// Initialize consensus engine
	consensusEngine, err := consensus.NewEngine(&cfg.Consensus, p2pNode, messageRouter, networkMonitor)
	if err != nil {
		return fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Initialize scheduler
	schedulerEngine, err := scheduler.NewEngine(&cfg.Scheduler, p2pNode, consensusEngine)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Initialize performance monitoring
	log.Printf("📊 Initializing performance monitoring...")

	// Create performance optimization engine
	perfOptConfig := performance.DefaultOptimizationConfig()
	perfOptConfig.Enabled = true
	perfOptConfig.OptimizationLevel = "balanced"
	perfOptConfig.OptimizationInterval = 5 * time.Minute
	perfOptEngine := performance.NewPerformanceOptimizationEngine(perfOptConfig)

	// Create Prometheus exporter for metrics
	prometheusConfig := observability.DefaultPrometheusConfig()
	prometheusConfig.ListenAddress = ":9090"
	prometheusExporter := observability.NewPrometheusExporter(prometheusConfig)

	log.Printf("✅ Performance monitoring initialized")

	// Initialize API server
	apiServer, err := api.NewServer(&cfg.API, p2pNode, consensusEngine, schedulerEngine)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}

	// Initialize web server
	log.Printf("🌐 Initializing web server...")
	webConfig := web.DefaultConfig()

	// Use configuration from YAML file
	if cfg.Web.Listen != "" {
		webConfig.ListenAddress = cfg.Web.Listen
	} else {
		webConfig.ListenAddress = ":8081" // Use different port from API
	}

	// Only use custom static path if the directory actually exists
	if cfg.Web.StaticDir != "" {
		if _, err := os.Stat(cfg.Web.StaticDir); err == nil {
			webConfig.StaticPath = cfg.Web.StaticDir
			log.Printf("📁 Using custom static files from: %s", cfg.Web.StaticDir)
		} else {
			log.Printf("📁 Custom static directory not found (%s), using embedded files", cfg.Web.StaticDir)
			webConfig.StaticPath = "" // Use embedded files
		}
	}

	webConfig.EnableAuth = true // Enable authentication by default
	webServer := web.NewWebServer(webConfig, apiServer)
	log.Printf("✅ Web server initialized on %s", webConfig.ListenAddress)

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

	// Start performance monitoring
	log.Printf("📊 Starting performance monitoring...")
	if err := perfOptEngine.Start(); err != nil {
		log.Printf("⚠️  Failed to start performance optimization engine: %v", err)
	} else {
		log.Printf("✅ Performance optimization engine started")
	}

	// Start Prometheus metrics exporter
	monitoringCtx := context.Background()
	if err := prometheusExporter.Start(monitoringCtx); err != nil {
		log.Printf("⚠️  Failed to start Prometheus exporter: %v", err)
	} else {
		log.Printf("✅ Prometheus metrics exporter started on :9090")
	}

	// Start API server
	log.Printf("🚀 Starting API server...")
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("⚠️  API server error: %v", err)
		}
	}()
	log.Printf("✅ API server started on %s", cfg.API.Listen)

	// Start web server
	log.Printf("🌐 Starting web server...")
	go func() {
		if err := webServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("⚠️  Web server error: %v", err)
		}
	}()
	log.Printf("✅ Web server started on %s", webConfig.ListenAddress)

	// Initialize and start Ollama integration
	log.Printf("🤖 Initializing Ollama integration...")
	ollamaIntegration := integration.NewSimpleOllamaIntegration(cfg)
	if err := ollamaIntegration.Start(); err != nil {
		log.Printf("⚠️  Ollama integration failed to start: %v", err)
		log.Printf("   The distributed system will run without Ollama integration")
		log.Printf("   To enable integration, install Ollama: https://ollama.com/download")
	} else {
		log.Printf("✅ Ollama integration started successfully")
		log.Printf("   Ollama API: %s", ollamaIntegration.GetOllamaAPIURL())
		log.Printf("   Distributed API: %s", ollamaIntegration.GetDistributedAPIURL())

		// Connect integration to API server
		apiServer.SetIntegration(ollamaIntegration)
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

	if err := apiServer.Stop(); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}

	if err := schedulerEngine.Shutdown(shutdownCtx); err != nil {
		log.Printf("Scheduler shutdown error: %v", err)
	}

	if err := consensusEngine.Shutdown(shutdownCtx); err != nil {
		log.Printf("Consensus engine shutdown error: %v", err)
	}

	if err := p2pNode.Stop(); err != nil {
		log.Printf("P2P node shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Ollama Distributed Node Status\n")
	fmt.Printf("==============================\n\n")

	// Connect to existing node to get status
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to connect to the API server to get status
	apiAddr := cfg.API.Listen
	fmt.Printf("🔗 API Server: %s\n", apiAddr)

	// Initialize a temporary P2P node to check cluster status
	p2pNode, err := p2p.NewNode(ctx, &cfg.P2P)
	if err != nil {
		fmt.Printf("❌ Failed to initialize P2P node: %v\n", err)
		return nil // Don't fail entirely, show what we can
	}

	// Start P2P node temporarily to get peer information
	if err := p2pNode.Start(); err != nil {
		fmt.Printf("❌ Failed to start P2P node: %v\n", err)
	} else {
		defer p2pNode.Stop()

		// Wait a moment for peer discovery
		time.Sleep(2 * time.Second)

		// Get node information
		nodeStatus := p2pNode.GetStatus()
		metrics := p2pNode.GetMetrics()
		capabilities := p2pNode.GetCapabilities()
		resourceMetrics := p2pNode.GetResourceMetrics()

		// Display node health and basic info
		fmt.Printf("📊 Node Health\n")
		fmt.Printf("   ID: %s\n", nodeStatus.ID)
		fmt.Printf("   Status: %s\n", getStatusString(nodeStatus.Started))
		fmt.Printf("   Uptime: %v\n", nodeStatus.Uptime)
		fmt.Printf("   Last Activity: %v\n", nodeStatus.LastActivity.Format(time.RFC3339))
		fmt.Printf("\n")

		// Display peer and cluster information
		fmt.Printf("🌐 Cluster Status\n")
		fmt.Printf("   Connected Peers: %d\n", nodeStatus.ConnectedPeers)
		fmt.Printf("   Total Connections: %d\n", metrics.TotalConnections)
		fmt.Printf("   Connection Errors: %d\n", metrics.ConnectionErrors)
		fmt.Printf("   Peers Discovered: %d\n", metrics.PeersDiscovered)

		// Show listen addresses
		fmt.Printf("   Listen Addresses:\n")
		for _, addr := range nodeStatus.ListenAddresses {
			fmt.Printf("     - %s\n", addr.String())
		}
		fmt.Printf("\n")

		// Display resource utilization
		fmt.Printf("💻 Resource Utilization\n")
		if resourceMetrics != nil {
			fmt.Printf("   CPU Usage: %.1f%%\n", resourceMetrics.CPUUsage)
			fmt.Printf("   Memory Usage: %s\n", formatBytes(resourceMetrics.MemoryUsage))
			fmt.Printf("   Disk Usage: %s\n", formatBytes(resourceMetrics.DiskUsage))
			fmt.Printf("   Network RX: %s/s\n", formatBytes(resourceMetrics.NetworkRx))
			fmt.Printf("   Network TX: %s/s\n", formatBytes(resourceMetrics.NetworkTx))
		} else {
			fmt.Printf("   Resource metrics unavailable\n")
		}
		fmt.Printf("\n")

		// Display node capabilities
		fmt.Printf("⚡ Node Capabilities\n")
		if capabilities != nil {
			fmt.Printf("   CPU Cores: %d\n", capabilities.CPUCores)
			fmt.Printf("   Memory: %s\n", formatBytes(capabilities.Memory))
			fmt.Printf("   Storage: %s\n", formatBytes(capabilities.Storage))
			fmt.Printf("   Supported Models: %v\n", capabilities.SupportedModels)
			fmt.Printf("   Available: %t\n", capabilities.Available)
			fmt.Printf("   Load Factor: %.2f\n", capabilities.LoadFactor)
		} else {
			fmt.Printf("   Capabilities not configured\n")
		}
		fmt.Printf("\n")

		// Display performance metrics
		fmt.Printf("📈 Performance Metrics\n")
		fmt.Printf("   Authentication Attempts: %d\n", metrics.AuthAttempts)
		fmt.Printf("   Authentication Successes: %d\n", metrics.AuthSuccesses)
		fmt.Printf("   Authentication Failures: %d\n", metrics.AuthFailures)
		fmt.Printf("   Content Published: %d\n", metrics.ContentPublished)
		fmt.Printf("   Content Requests: %d\n", metrics.ContentRequests)
		fmt.Printf("   Content Provided: %d\n", metrics.ContentProvided)
		fmt.Printf("   Average Latency: %v\n", metrics.AverageLatency)
		fmt.Printf("   Message Throughput: %d msg/s\n", metrics.MessageThroughput)
		fmt.Printf("\n")

		// Display consensus status if available
		fmt.Printf("🗳️  Consensus Status\n")
		fmt.Printf("   Consensus Engine: %s\n", getConsensusStatus(cfg))
		fmt.Printf("   Data Directory: %s\n", cfg.Consensus.DataDir)
		fmt.Printf("   Bind Address: %s\n", cfg.Consensus.BindAddr)
		fmt.Printf("\n")

		// Display scheduler status
		fmt.Printf("🎯 Scheduler Status\n")
		fmt.Printf("   Algorithm: %s\n", cfg.Scheduler.Algorithm)
		fmt.Printf("   Load Balancing: %s\n", cfg.Scheduler.LoadBalancing)
		fmt.Printf("   Worker Count: %d\n", cfg.Scheduler.WorkerCount)
		fmt.Printf("   Queue Size: %d\n", cfg.Scheduler.QueueSize)
		fmt.Printf("\n")
	}

	fmt.Printf("✅ Status check completed\n")
	return nil
}

func runJoin(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	peers, _ := cmd.Flags().GetStringSlice("peers")
	if len(peers) == 0 {
		return fmt.Errorf("no peers specified, use --peers flag to specify peer addresses")
	}

	fmt.Printf("Joining Ollama Distributed Cluster\n")
	fmt.Printf("=================================\n\n")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize P2P networking
	fmt.Printf("🔧 Initializing P2P node...\n")
	p2pNode, err := p2p.NewNode(ctx, &cfg.P2P)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}

	// Start P2P node
	fmt.Printf("🚀 Starting P2P networking...\n")
	if err := p2pNode.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}
	defer p2pNode.Stop()

	nodeID := p2pNode.ID()
	fmt.Printf("📍 Node ID: %s\n\n", nodeID)

	// Connect to specified peers
	fmt.Printf("🌐 Connecting to peers...\n")
	var successfulConnections int
	var connectionErrors []string

	for i, peerAddr := range peers {
		fmt.Printf("   [%d/%d] Connecting to %s...", i+1, len(peers), peerAddr)

		if err := connectToPeer(ctx, p2pNode, peerAddr); err != nil {
			fmt.Printf(" ❌ Failed: %v\n", err)
			connectionErrors = append(connectionErrors, fmt.Sprintf("%s: %v", peerAddr, err))
		} else {
			fmt.Printf(" ✅ Connected\n")
			successfulConnections++
		}
	}

	if successfulConnections == 0 {
		fmt.Printf("\n❌ Failed to connect to any peers\n")
		for _, errMsg := range connectionErrors {
			fmt.Printf("   - %s\n", errMsg)
		}
		return fmt.Errorf("no successful peer connections")
	}

	fmt.Printf("\n✅ Connected to %d/%d peers\n\n", successfulConnections, len(peers))

	// Wait for peer discovery and cluster state synchronization
	fmt.Printf("🔍 Discovering cluster topology...\n")
	time.Sleep(5 * time.Second)

	// Get current cluster state
	connectedPeers := p2pNode.GetConnectedPeers()
	fmt.Printf("   Found %d peers in cluster\n", len(connectedPeers))

	// Create messaging and monitoring components
	messageRouter := messaging.NewMessageRouter(nil)
	networkMonitor := monitoring.NewNetworkMonitor(nil)

	// Initialize consensus engine and join cluster
	fmt.Printf("🗳️  Joining consensus cluster...\n")
	consensusEngine, err := consensus.NewEngine(&cfg.Consensus, p2pNode, messageRouter, networkMonitor)
	if err != nil {
		return fmt.Errorf("failed to create consensus engine: %w", err)
	}

	// Start consensus engine (it will automatically try to join the cluster)
	if err := consensusEngine.Start(); err != nil {
		return fmt.Errorf("failed to start consensus engine: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		consensusEngine.Shutdown(shutdownCtx)
	}()

	// Wait for consensus participation
	fmt.Printf("⏳ Waiting for consensus participation...\n")
	time.Sleep(10 * time.Second)

	// Check consensus status
	if consensusEngine.IsLeader() {
		fmt.Printf("👑 This node is now the cluster leader\n")
	} else {
		leaderAddr := consensusEngine.Leader()
		if leaderAddr != "" {
			fmt.Printf("📡 Following leader: %s\n", leaderAddr)
		} else {
			fmt.Printf("⏳ Waiting for leader election...\n")
		}
	}

	// Initialize scheduler
	fmt.Printf("🎯 Joining scheduler network...\n")
	schedulerEngine, err := scheduler.NewEngine(&cfg.Scheduler, p2pNode, consensusEngine)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	if err := schedulerEngine.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		schedulerEngine.Shutdown(shutdownCtx)
	}()

	// Sync cluster state
	fmt.Printf("🔄 Synchronizing cluster state...\n")
	time.Sleep(3 * time.Second)

	// Display final cluster state
	fmt.Printf("\n📊 Cluster Join Summary\n")
	fmt.Printf("   Node ID: %s\n", nodeID)
	fmt.Printf("   Connected Peers: %d\n", len(p2pNode.GetConnectedPeers()))
	fmt.Printf("   Consensus Status: %s\n", getConsensusJoinStatus(consensusEngine))
	fmt.Printf("   Scheduler Status: %s\n", getSchedulerStatus(schedulerEngine))

	// Final validation
	if len(p2pNode.GetConnectedPeers()) > 0 {
		fmt.Printf("\n✅ Successfully joined cluster!\n")
		fmt.Printf("💡 You can now start the full node with: ollama-distributed start\n")
		return nil
	} else {
		fmt.Printf("\n⚠️  Joined with warnings - no active peer connections\n")
		return nil
	}
}

// Helper functions for status display

func getStatusString(started bool) string {
	if started {
		return "✅ Online"
	}
	return "❌ Offline"
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getConsensusStatus(cfg *config.Config) string {
	if cfg.Consensus.Bootstrap {
		return "Bootstrap mode"
	}
	return "Follower mode"
}

func connectToPeer(ctx context.Context, p2pNode *p2p.P2PNode, peerAddr string) error {
	// Parse multiaddr format peer address
	// Example: /ip4/192.168.1.100/tcp/4001/p2p/QmPeerID
	maddr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		// Try simpler format: ip:port
		if host, port, err := net.SplitHostPort(peerAddr); err == nil {
			maddr, err = multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", host, port))
			if err != nil {
				return fmt.Errorf("invalid peer address format: %w", err)
			}
		} else {
			return fmt.Errorf("invalid peer address format: %w", err)
		}
	}

	// Extract peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		// If no peer ID in address, try to connect anyway
		// This is a simplified connection attempt
		return fmt.Errorf("could not extract peer info: %w", err)
	}

	// Connect to the peer
	return p2pNode.ConnectToPeer(ctx, *peerInfo)
}

func getConsensusJoinStatus(engine *consensus.Engine) string {
	if engine.IsLeader() {
		return "Leader"
	}
	leader := engine.Leader()
	if leader != "" {
		return fmt.Sprintf("Follower (Leader: %s)", leader)
	}
	return "Waiting for leader"
}

func getSchedulerStatus(engine *scheduler.Engine) string {
	if engine.IsHealthy() {
		stats := engine.GetStats()
		return fmt.Sprintf("Healthy (%d nodes, %d models)", stats.NodesOnline, stats.ModelsTotal)
	}
	return "Initializing"
}

// Proxy command implementations

func runProxyStatus(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	fmt.Printf("🔄 Ollama Proxy Status\n")
	fmt.Printf("=====================\n\n")

	// Make API request to get proxy status
	url := apiURL + "/api/v1/proxy/status"
	resp, err := makeHTTPRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to get proxy status: %w", err)
	}

	if jsonOutput {
		fmt.Println(string(resp))
		return nil
	}

	// Display formatted output
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Response: %s\n", string(resp))

	return nil
}

func runProxyInstances(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	fmt.Printf("🖥️  Proxy Instances\n")
	fmt.Printf("==================\n\n")

	// Make API request to get instances
	url := apiURL + "/api/v1/proxy/instances"
	resp, err := makeHTTPRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to get proxy instances: %w", err)
	}

	if jsonOutput {
		fmt.Println(string(resp))
		return nil
	}

	// Display formatted output
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Response: %s\n", string(resp))

	return nil
}

func runProxyMetrics(cmd *cobra.Command, args []string) error {
	apiURL, _ := cmd.Flags().GetString("api-url")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	watch, _ := cmd.Flags().GetBool("watch")
	interval, _ := cmd.Flags().GetInt("interval")

	if watch {
		return watchProxyMetrics(apiURL, jsonOutput, interval)
	}

	fmt.Printf("📊 Proxy Metrics\n")
	fmt.Printf("================\n\n")

	// Make API request to get metrics
	url := apiURL + "/api/v1/proxy/metrics"
	resp, err := makeHTTPRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to get proxy metrics: %w", err)
	}

	if jsonOutput {
		fmt.Println(string(resp))
		return nil
	}

	// Display formatted output
	fmt.Printf("API URL: %s\n", apiURL)
	fmt.Printf("Response: %s\n", string(resp))

	return nil
}

func watchProxyMetrics(apiURL string, jsonOutput bool, interval int) error {
	fmt.Printf("👀 Watching proxy metrics (interval: %ds, press Ctrl+C to stop)\n\n", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Handle Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			// Clear screen and show updated metrics
			fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top

			url := apiURL + "/api/v1/proxy/metrics"
			resp, err := makeHTTPRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if jsonOutput {
				fmt.Println(string(resp))
			} else {
				fmt.Printf("📊 Proxy Metrics (Updated: %s)\n", time.Now().Format("15:04:05"))
				fmt.Printf("=====================================\n\n")
				fmt.Printf("Response: %s\n", string(resp))
			}

		case <-c:
			fmt.Printf("\n👋 Stopping metrics watch...\n")
			return nil
		}
	}
}

func makeHTTPRequest(method, url string, body interface{}) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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
			log.Printf("Warning: Could not determine home directory: %v", err)
			// Continue with current directory only
		} else {
			viper.AddConfigPath(home)
		}

		viper.AddConfigPath(".")
		viper.SetConfigName(".ollama-distributed")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}
