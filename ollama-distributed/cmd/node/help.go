package main

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// helpCmd provides enhanced help and guidance
var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "📚 Get help and guidance for OllamaMax",
	Long: `📚 Get comprehensive help and guidance for OllamaMax

This command provides detailed help, examples, and guidance for using OllamaMax.
It includes quick start guides, common tasks, and troubleshooting tips.`,
	RunE: runHelp,
}

var (
	helpExamples bool
	helpQuick    bool
	helpTrouble  bool
)

func initHelpCommands() {
	helpCmd.Flags().BoolVar(&helpExamples, "examples", false, "Show common usage examples")
	helpCmd.Flags().BoolVar(&helpQuick, "quick", false, "Show quick start guide")
	helpCmd.Flags().BoolVar(&helpTrouble, "troubleshoot", false, "Show troubleshooting guide")

	rootCmd.AddCommand(helpCmd)
	rootCmd.AddCommand(versionCmd)
}

func runHelp(cmd *cobra.Command, args []string) error {
	if helpQuick {
		return showQuickStartGuide()
	}

	if helpExamples {
		return showExamples()
	}

	if helpTrouble {
		return showTroubleshooting()
	}

	// Default comprehensive help
	return showComprehensiveHelp()
}

func showComprehensiveHelp() error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow, color.Bold)
	blue := color.New(color.FgBlue)

	fmt.Println()
	cyan.Println("🚀 OllamaMax - Enterprise Distributed AI Platform")
	cyan.Println("================================================")
	fmt.Println()

	green.Println("OllamaMax transforms Ollama into a distributed, enterprise-grade AI platform")
	green.Println("with automatic scaling, load balancing, and high availability.")
	fmt.Println()

	yellow.Println("🎯 Quick Actions:")
	fmt.Println("  ollama-distributed quickstart          # Start with defaults (fastest)")
	fmt.Println("  ollama-distributed setup              # Interactive configuration")
	fmt.Println("  ollama-distributed start              # Start your node")
	fmt.Println("  ollama-distributed proxy status       # Check cluster status")
	fmt.Println()

	yellow.Println("📋 Main Commands:")
	fmt.Println("  setup         Interactive setup wizard")
	fmt.Println("  quickstart    Quick start with defaults")
	fmt.Println("  start         Start OllamaMax node")
	fmt.Println("  status        Show node status")
	fmt.Println("  join          Join existing cluster")
	fmt.Println("  proxy         Proxy commands to Ollama")
	fmt.Println("  validate      Validate configuration")
	fmt.Println()

	yellow.Println("🌐 Access Points:")
	fmt.Printf("  Web UI:       http://localhost:8081\n")
	fmt.Printf("  API:          http://localhost:8080\n")
	fmt.Printf("  Metrics:      http://localhost:9090\n")
	fmt.Println()

	yellow.Println("📚 Get More Help:")
	fmt.Println("  ollama-distributed help --quick           # Quick start guide")
	fmt.Println("  ollama-distributed help --examples        # Usage examples")
	fmt.Println("  ollama-distributed help --troubleshoot    # Troubleshooting")
	fmt.Println("  ollama-distributed [command] --help       # Command-specific help")
	fmt.Println()

	blue.Println("📖 Documentation:")
	fmt.Println("  GETTING_STARTED.md     - User-friendly getting started guide")
	fmt.Println("  README.md              - Project overview and features")
	fmt.Println("  docs/                  - Complete documentation")
	fmt.Println()

	green.Println("🎉 Ready to get started? Run: ollama-distributed quickstart")
	fmt.Println()

	return nil
}

func showQuickStartGuide() error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow, color.Bold)

	fmt.Println()
	cyan.Println("⚡ OllamaMax Quick Start Guide")
	cyan.Println("=============================")
	fmt.Println()

	yellow.Println("🚀 Option 1: Super Quick (30 seconds)")
	fmt.Println("  1. ollama-distributed quickstart")
	fmt.Println("  2. ollama-distributed start --config quickstart-config.yaml")
	fmt.Println("  3. Open http://localhost:8081")
	fmt.Println()

	yellow.Println("🛠️  Option 2: Custom Setup (2 minutes)")
	fmt.Println("  1. ollama-distributed setup")
	fmt.Println("  2. Follow the interactive prompts")
	fmt.Println("  3. ollama-distributed start --config config.yaml")
	fmt.Println("  4. Open http://localhost:8081")
	fmt.Println()

	yellow.Println("🤖 Using AI Models:")
	fmt.Println("  ollama-distributed proxy pull llama2      # Pull a model")
	fmt.Println("  ollama-distributed proxy list             # List models")
	fmt.Println("  curl -X POST http://localhost:8080/api/generate \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"model\":\"llama2\",\"prompt\":\"Hello!\"}'")
	fmt.Println()

	yellow.Println("📊 Monitoring:")
	fmt.Println("  ollama-distributed proxy status           # Cluster status")
	fmt.Println("  ollama-distributed proxy instances        # Node health")
	fmt.Println("  http://localhost:8081                     # Web dashboard")
	fmt.Println()

	green.Println("🎯 Next Steps:")
	fmt.Println("  - Read GETTING_STARTED.md for detailed guide")
	fmt.Println("  - Explore the web interface at http://localhost:8081")
	fmt.Println("  - Join multiple nodes to create a cluster")
	fmt.Println()

	return nil
}

func showExamples() error {
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("💡 OllamaMax Usage Examples")
	cyan.Println("===========================")
	fmt.Println()

	yellow.Println("🏗️  Setup and Configuration:")
	fmt.Println("  # Interactive setup with custom configuration")
	fmt.Println("  ollama-distributed setup")
	fmt.Println()
	fmt.Println("  # Quick start for development")
	fmt.Println("  ollama-distributed quickstart --name dev-node --port 9000")
	fmt.Println()
	fmt.Println("  # Validate existing configuration")
	fmt.Println("  ollama-distributed validate config.yaml")
	fmt.Println()

	yellow.Println("🚀 Starting Nodes:")
	fmt.Println("  # Start with default config")
	fmt.Println("  ollama-distributed start")
	fmt.Println()
	fmt.Println("  # Start with custom config")
	fmt.Println("  ollama-distributed start --config production.yaml")
	fmt.Println()
	fmt.Println("  # Start with debug logging")
	fmt.Println("  ollama-distributed start --log-level debug")
	fmt.Println()

	yellow.Println("🌐 Cluster Management:")
	fmt.Println("  # Join existing cluster")
	fmt.Println("  ollama-distributed join --bootstrap-peer 192.168.1.100:9000")
	fmt.Println()
	fmt.Println("  # Check cluster status")
	fmt.Println("  ollama-distributed proxy status")
	fmt.Println()
	fmt.Println("  # View cluster nodes")
	fmt.Println("  curl http://localhost:8080/api/v1/nodes")
	fmt.Println()

	yellow.Println("🤖 Model Management:")
	fmt.Println("  # Pull popular models")
	fmt.Println("  ollama-distributed proxy pull llama2")
	fmt.Println("  ollama-distributed proxy pull codellama")
	fmt.Println("  ollama-distributed proxy pull mistral")
	fmt.Println()
	fmt.Println("  # List available models")
	fmt.Println("  ollama-distributed proxy list")
	fmt.Println()
	fmt.Println("  # Remove unused models")
	fmt.Println("  ollama-distributed proxy rm old-model")
	fmt.Println()

	yellow.Println("📊 Monitoring and Debugging:")
	fmt.Println("  # Check node health")
	fmt.Println("  ollama-distributed proxy instances")
	fmt.Println()
	fmt.Println("  # View performance metrics")
	fmt.Println("  curl http://localhost:8080/api/v1/proxy/metrics")
	fmt.Println()
	fmt.Println("  # Monitor model transfers")
	fmt.Println("  curl http://localhost:8080/api/v1/transfers")
	fmt.Println()

	yellow.Println("🔌 API Usage:")
	fmt.Println("  # Generate text")
	fmt.Println("  curl -X POST http://localhost:8080/api/generate \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"model\":\"llama2\",\"prompt\":\"Explain AI\"}'")
	fmt.Println()
	fmt.Println("  # Chat completion")
	fmt.Println("  curl -X POST http://localhost:8080/api/chat \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"model\":\"llama2\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}'")
	fmt.Println()

	green.Println("💡 Pro Tips:")
	fmt.Println("  - Use the web interface at http://localhost:8081 for easy management")
	fmt.Println("  - Monitor performance with Grafana dashboards")
	fmt.Println("  - Set up alerts for production deployments")
	fmt.Println("  - Use configuration files for reproducible deployments")
	fmt.Println()

	return nil
}

func showTroubleshooting() error {
	cyan := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("🔧 OllamaMax Troubleshooting Guide")
	cyan.Println("==================================")
	fmt.Println()

	red.Println("❌ Common Issues and Solutions:")
	fmt.Println()

	yellow.Println("🚫 Node Won't Start:")
	fmt.Println("  Problem: Node fails to start or exits immediately")
	fmt.Println("  Solutions:")
	fmt.Println("    • Check configuration: ollama-distributed validate config.yaml")
	fmt.Println("    • Verify ports are available: netstat -tulpn | grep :8080")
	fmt.Println("    • Check permissions: ls -la config.yaml")
	fmt.Println("    • Start with debug logging: --log-level debug")
	fmt.Println()

	yellow.Println("🌐 Can't Access Web UI:")
	fmt.Println("  Problem: Web interface not loading at http://localhost:8081")
	fmt.Println("  Solutions:")
	fmt.Println("    • Check if web server is enabled in config")
	fmt.Println("    • Verify port 8081 is not in use: netstat -tulpn | grep :8081")
	fmt.Println("    • Try different port: --web-port 8082")
	fmt.Println("    • Check firewall settings")
	fmt.Println()

	yellow.Println("🔗 Cluster Connection Issues:")
	fmt.Println("  Problem: Nodes can't join cluster or communicate")
	fmt.Println("  Solutions:")
	fmt.Println("    • Test network connectivity: ping <peer-ip>")
	fmt.Println("    • Check P2P port: telnet <peer-ip> 9000")
	fmt.Println("    • Verify bootstrap peers in config")
	fmt.Println("    • Check NAT/firewall configuration")
	fmt.Println()

	yellow.Println("🤖 Model Issues:")
	fmt.Println("  Problem: Models won't pull, load, or sync")
	fmt.Println("  Solutions:")
	fmt.Println("    • Check disk space: df -h ./models")
	fmt.Println("    • Verify model name: ollama-distributed proxy list")
	fmt.Println("    • Check transfer status: curl http://localhost:8080/api/v1/transfers")
	fmt.Println("    • Clear model cache: rm -rf ./models/cache")
	fmt.Println()

	yellow.Println("📊 Performance Issues:")
	fmt.Println("  Problem: Slow responses or high resource usage")
	fmt.Println("  Solutions:")
	fmt.Println("    • Check system resources: top, htop")
	fmt.Println("    • Monitor metrics: http://localhost:8080/api/v1/proxy/metrics")
	fmt.Println("    • Adjust worker count in config")
	fmt.Println("    • Enable performance optimization")
	fmt.Println()

	green.Println("🔍 Diagnostic Commands:")
	fmt.Println("  # System information")
	fmt.Printf("  OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("  ollama-distributed version")
	fmt.Println()
	fmt.Println("  # Configuration check")
	fmt.Println("  ollama-distributed validate config.yaml")
	fmt.Println()
	fmt.Println("  # Health checks")
	fmt.Println("  curl http://localhost:8080/health")
	fmt.Println("  curl http://localhost:8081/health")
	fmt.Println()
	fmt.Println("  # Detailed status")
	fmt.Println("  ollama-distributed proxy status --verbose")
	fmt.Println("  ollama-distributed proxy instances --detailed")
	fmt.Println()

	green.Println("📞 Getting Help:")
	fmt.Println("  • Documentation: GETTING_STARTED.md, README.md")
	fmt.Println("  • Issues: https://github.com/KhryptorGraphics/OllamaMax/issues")
	fmt.Println("  • Discussions: https://github.com/KhryptorGraphics/OllamaMax/discussions")
	fmt.Println("  • Logs: Check application logs for detailed error messages")
	fmt.Println()

	return nil
}

// versionCmd shows detailed version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Show detailed version and build information for OllamaMax`,
	RunE:  runVersion,
}

// versionCmd is added to rootCmd in initHelpCommands()

func runVersion(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("🚀 OllamaMax Version Information")
	cyan.Println("===============================")
	fmt.Println()

	green.Printf("Version:      %s\n", version)
	green.Printf("Go Version:   %s\n", runtime.Version())
	green.Printf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	green.Printf("Compiler:     %s\n", runtime.Compiler)
	fmt.Println()

	fmt.Println("🌟 OllamaMax - Enterprise Distributed AI Platform")
	fmt.Println("   Transform your AI infrastructure with enterprise-grade")
	fmt.Println("   distributed computing, security, and monitoring.")
	fmt.Println()

	return nil
}
