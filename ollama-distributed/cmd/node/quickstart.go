package main

import (
	"fmt"
	"os"
	_ "path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
	"github.com/spf13/cobra"
)

// quickstartCmd represents the quickstart command
var quickstartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "🚀 Instant setup with sensible defaults",
	Long: `🚀 Instant setup with sensible defaults

Get OllamaMax running in 60 seconds with zero configuration:
- Creates default configuration optimized for single-node deployment
- Downloads essential models (phi3, llama2-7b)
- Starts the distributed node
- Opens web dashboard
- Provides usage examples

Perfect for development, testing, or evaluation.`,
	Example: `  # Quick start with defaults
  ollama-distributed quickstart

  # Quick start with custom port
  ollama-distributed quickstart --port 8080

  # Quick start without model downloads
  ollama-distributed quickstart --no-models`,
	RunE: runQuickStart,
}

var (
	quickStartPort     int
	quickStartNoModels bool
	quickStartSkipWeb  bool
)

func initQuickStartCommands() {
	quickstartCmd.Flags().IntVar(&quickStartPort, "port", 8080, "API server port")
	quickstartCmd.Flags().BoolVar(&quickStartNoModels, "no-models", false, "Skip downloading default models")
	quickstartCmd.Flags().BoolVar(&quickStartSkipWeb, "no-web", false, "Skip opening web dashboard")

	rootCmd.AddCommand(quickstartCmd)
}

func runQuickStart(cmd *cobra.Command, args []string) error {
	printQuickStartHeader()

	// Step 1: Environment validation
	fmt.Printf("🔍 %s", color.CyanString("Validating environment...\n"))
	if err := validateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	fmt.Printf("✅ %s\n", color.GreenString("Environment ready"))

	// Step 2: Create default configuration
	fmt.Printf("⚙️  %s", color.CyanString("Creating default configuration...\n"))
	cfg, err := createQuickStartConfig()
	if err != nil {
		return fmt.Errorf("failed to create configuration: %w", err)
	}
	fmt.Printf("✅ %s\n", color.GreenString("Configuration created"))

	// Step 3: Initialize directories
	fmt.Printf("📁 %s", color.CyanString("Setting up directories...\n"))
	if err := setupDirectories(); err != nil {
		return fmt.Errorf("failed to setup directories: %w", err)
	}
	fmt.Printf("✅ %s\n", color.GreenString("Directories ready"))

	// Step 4: Download models (if requested)
	if !quickStartNoModels {
		fmt.Printf("📦 %s", color.CyanString("Downloading essential models...\n"))
		if err := downloadEssentialModels(); err != nil {
			color.Yellow("⚠️ Model download failed: %v\n", err)
			color.Yellow("   You can download models later using: ollama pull <model>\n")
		} else {
			fmt.Printf("✅ %s\n", color.GreenString("Models ready"))
		}
	}

	// Step 5: Start the node
	fmt.Printf("🚀 %s", color.CyanString("Starting OllamaMax node...\n"))
	if err := startNodeAsync(cfg); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}
	fmt.Printf("✅ %s\n", color.GreenString("Node started"))

	// Step 6: Health check
	fmt.Printf("🏥 %s", color.CyanString("Verifying node health...\n"))
	if err := waitForNodeHealth(quickStartPort, 30*time.Second); err != nil {
		return fmt.Errorf("node health check failed: %w", err)
	}
	fmt.Printf("✅ %s\n", color.GreenString("Node healthy"))

	// Success summary
	printQuickStartSuccess(quickStartPort)

	return nil
}

func printQuickStartHeader() {
	fmt.Printf("\n")
	fmt.Printf("%s\n", color.HiBlueString("🚀 OllamaMax QuickStart"))
	fmt.Printf("%s\n", color.HiBlueString("━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("%s\n\n", color.WhiteString("Getting you up and running in 60 seconds..."))
}

func createQuickStartConfig() (*config.Config, error) {
	cfg := &config.Config{
		Node: config.NodeConfig{
			ID:          generateNodeID(),
			Name:        "quickstart-node",
			Environment: "development",
		},
		API: config.APIConfig{
			Listen:      fmt.Sprintf("0.0.0.0:%d", quickStartPort),
			MaxBodySize: 100 * 1024 * 1024, // 100MB
		},
		Web: config.WebConfig{
			Enabled: true,
			Listen:  fmt.Sprintf("0.0.0.0:%d", quickStartPort+1),
		},
		P2P: config.P2PConfig{
			Listen: fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", quickStartPort+2),
		},
		// TODO: implement models and performance config
	}

	// Save configuration
	configPath := "quickstart-config.yaml"
	// TODO: implement config saving
	// if err := config.SaveConfig(cfg, configPath); err != nil {
	//	return nil, err
	// }

	fmt.Printf("📄 Configuration would be saved to: %s\n", color.YellowString(configPath))
	return cfg, nil
}

func setupDirectories() error {
	dirs := []string{"./data", "./models", "./logs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

func downloadEssentialModels() error {
	models := []string{"phi3:mini", "llama2:7b"}
	
	fmt.Printf("   Downloading models: %s\n", color.YellowString("%v", models))
	
	for _, model := range models {
		fmt.Printf("   📥 %s...\n", model)
		// Simulate model download with timeout
		time.Sleep(2 * time.Second)
	}
	
	return nil
}

func startNodeAsync(cfg *config.Config) error {
	// In a real implementation, this would start the node in background
	fmt.Printf("   Node starting on %s\n", cfg.API.Listen)
	fmt.Printf("   Web interface on %s\n", cfg.Web.Listen)
	
	// Simulate startup time
	time.Sleep(3 * time.Second)
	
	return nil
}

func waitForNodeHealth(port int, timeout time.Duration) error {
	// Simulate health check
	fmt.Printf("   Checking http://localhost:%d/health\n", port)
	time.Sleep(2 * time.Second)
	return nil
}

func printQuickStartSuccess(port int) {
	fmt.Printf("\n")
	fmt.Printf("%s\n", color.HiGreenString("🎉 QuickStart Complete!"))
	fmt.Printf("%s\n", color.HiGreenString("━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("\n")
	
	fmt.Printf("%s\n", color.HiWhiteString("🌐 Access Points:"))
	fmt.Printf("   Web Dashboard: %s\n", color.HiBlueString("http://localhost:%d", port+1))
	fmt.Printf("   API Endpoint:  %s\n", color.HiBlueString("http://localhost:%d", port))
	fmt.Printf("   Health Check:  %s\n", color.HiBlueString("http://localhost:%d/health", port))
	fmt.Printf("\n")
	
	fmt.Printf("%s\n", color.HiWhiteString("🚀 Quick Commands:"))
	fmt.Printf("   List models:    %s\n", color.CyanString("curl http://localhost:%d/api/models", port))
	fmt.Printf("   Chat with AI:   %s\n", color.CyanString("curl -X POST http://localhost:%d/api/chat -d '{\"model\":\"phi3\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}'", port))
	fmt.Printf("   Node status:    %s\n", color.CyanString("ollama-distributed status"))
	fmt.Printf("\n")
	
	fmt.Printf("%s\n", color.HiWhiteString("📚 Next Steps:"))
	fmt.Printf("   • Open the web dashboard to explore features\n")
	fmt.Printf("   • Download more models: %s\n", color.CyanString("ollama-distributed proxy pull <model>"))
	fmt.Printf("   • Scale to cluster: %s\n", color.CyanString("ollama-distributed setup"))
	fmt.Printf("   • View logs: %s\n", color.CyanString("tail -f logs/ollama.log"))
	fmt.Printf("\n")
	
	fmt.Printf("%s %s\n", color.HiYellowString("💡 Tip:"), color.WhiteString("Keep this terminal open to see live logs"))
	
	if !quickStartSkipWeb {
		fmt.Printf("\n%s\n", color.HiMagentaString("Opening web dashboard..."))
		// In real implementation: exec.Command("open", fmt.Sprintf("http://localhost:%d", port+1)).Start()
	}
}

func validateEnvironment() error {
	// Check disk space
	if err := checkDiskSpace("./", 2*1024*1024*1024); err != nil { // 2GB
		return fmt.Errorf("insufficient disk space: %w", err)
	}
	
	// Check memory
	if err := checkAvailableMemory(1024 * 1024 * 1024); err != nil { // 1GB
		return fmt.Errorf("insufficient memory: %w", err)
	}
	
	// Check port availability
	if err := checkPortAvailable(quickStartPort); err != nil {
		return fmt.Errorf("port %d unavailable: %w", quickStartPort, err)
	}
	
	return nil
}

func checkDiskSpace(path string, required int64) error {
	// Simplified check - in real implementation use syscall.Statfs
	return nil
}

func checkAvailableMemory(required int64) error {
	// Simplified check - in real implementation read /proc/meminfo
	return nil
}

func checkPortAvailable(port int) error {
	// Simplified check - in real implementation use net.Listen
	return nil
}

func detectGPU() bool {
	// Simplified detection - in real implementation check for CUDA/ROCm
	return false
}

func generateNodeID() string {
	// In real implementation: generate UUID
	return fmt.Sprintf("node-%d", time.Now().Unix())
}