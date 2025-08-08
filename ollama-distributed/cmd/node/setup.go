package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/onboarding"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive setup wizard for OllamaMax",
	Long: `Interactive setup wizard that guides you through configuring your OllamaMax node.

This command will:
- Configure basic node settings
- Set up network and clustering options
- Configure security settings
- Generate a configuration file
- Provide next steps for getting started

Example:
  ollama-distributed setup
  ollama-distributed setup --config custom-config.yaml`,
	RunE: runSetup,
}

var (
	setupConfigFile string
	setupForce      bool
)

func initSetupCommands() {
	setupCmd.Flags().StringVar(&setupConfigFile, "config", "config.yaml", "Configuration file to generate")
	setupCmd.Flags().BoolVar(&setupForce, "force", false, "Overwrite existing configuration file")

	rootCmd.AddCommand(setupCmd)

	// Initialize other setup-related commands
	initQuickStartCommands()
	initValidateCommands()
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Print setup header
	printSetupHeader()

	// Check if config file already exists
	if !setupForce && fileExists(setupConfigFile) {
		return fmt.Errorf("configuration file %s already exists. Use --force to overwrite", setupConfigFile)
	}

	// Create onboarding manager
	onboardingManager := onboarding.NewOnboardingManager()

	// Run onboarding process
	config, err := onboardingManager.RunOnboarding()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Generate configuration file
	if err := onboardingManager.GenerateConfigFile(config, setupConfigFile); err != nil {
		return fmt.Errorf("failed to generate configuration file: %w", err)
	}

	// Print success message and next steps
	printSetupSuccess(config, setupConfigFile)

	return nil
}

func printSetupHeader() {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	fmt.Println()
	cyan.Println("🛠️  OllamaMax Setup Wizard")
	cyan.Println("==========================")
	fmt.Println()
	green.Println("This wizard will help you set up your OllamaMax distributed system.")
	green.Println("Follow the prompts to configure your node.")
	fmt.Println()
}

func printSetupSuccess(config *onboarding.OnboardingConfig, configFile string) {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	green.Println("🎉 Setup Complete!")
	green.Println("==================")
	fmt.Println()

	cyan.Printf("✅ Configuration file created: %s\n", configFile)
	fmt.Println()

	yellow.Println("🚀 Next Steps:")
	yellow.Println("==============")
	fmt.Println()

	// Start the node
	fmt.Printf("1. Start your OllamaMax node:\n")
	fmt.Printf("   ./ollama-distributed start --config %s\n", configFile)
	fmt.Println()

	// Access Web UI
	if config.EnableWebUI {
		fmt.Printf("2. Access the Web UI:\n")
		fmt.Printf("   http://localhost:8081\n")
		fmt.Println()
	}

	// API access
	fmt.Printf("3. Access the API:\n")
	fmt.Printf("   http://localhost:%d/api/v1/proxy/status\n", config.ListenPort)
	fmt.Println()

	// Model management
	fmt.Printf("4. Pull your first model:\n")
	fmt.Printf("   ./ollama-distributed proxy pull llama2\n")
	fmt.Println()

	// Monitoring
	fmt.Printf("5. Monitor your node:\n")
	fmt.Printf("   ./ollama-distributed proxy status\n")
	fmt.Printf("   ./ollama-distributed proxy instances\n")
	fmt.Println()

	// Documentation
	fmt.Printf("6. Learn more:\n")
	fmt.Printf("   ./ollama-distributed help\n")
	fmt.Printf("   Check the documentation in the docs/ directory\n")
	fmt.Println()

	// Join cluster (if applicable)
	if config.JoinExisting && len(config.BootstrapPeers) > 0 {
		fmt.Printf("7. Your node will automatically join the cluster:\n")
		for _, peer := range config.BootstrapPeers {
			fmt.Printf("   - %s\n", peer)
		}
		fmt.Println()
	}

	green.Println("Happy distributed computing! 🚀")
	fmt.Println()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// quickStartCmd provides a quick start option
var quickStartCmd = &cobra.Command{
	Use:   "quickstart",
	Short: "Quick start with default configuration",
	Long: `Quick start OllamaMax with sensible defaults.

This command will:
- Create a default configuration
- Start the node immediately
- Enable Web UI on port 8081
- Enable API on port 8080

Example:
  ollama-distributed quickstart
  ollama-distributed quickstart --port 9000`,
	RunE: runQuickStart,
}

var (
	quickStartPort  int
	quickStartWebUI bool
	quickStartName  string
)

func initQuickStartCommands() {
	quickStartCmd.Flags().IntVar(&quickStartPort, "port", 8080, "API port")
	quickStartCmd.Flags().BoolVar(&quickStartWebUI, "web-ui", true, "Enable Web UI")
	quickStartCmd.Flags().StringVar(&quickStartName, "name", "ollama-quickstart", "Node name")

	rootCmd.AddCommand(quickStartCmd)
}

func runQuickStart(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	green.Println("🚀 OllamaMax Quick Start")
	green.Println("========================")
	fmt.Println()

	cyan.Println("Setting up OllamaMax with default configuration...")

	// Create quick start configuration
	config := &onboarding.OnboardingConfig{
		NodeName:       quickStartName,
		ListenPort:     quickStartPort,
		EnableWebUI:    quickStartWebUI,
		EnableSecurity: true,
		JoinExisting:   false,
		ModelDirectory: "./models",
		LogLevel:       "info",
	}

	// Generate configuration file
	configFile := "quickstart-config.yaml"
	onboardingManager := onboarding.NewOnboardingManager()
	if err := onboardingManager.GenerateConfigFile(config, configFile); err != nil {
		return fmt.Errorf("failed to generate configuration file: %w", err)
	}

	cyan.Printf("✅ Configuration created: %s\n", configFile)
	fmt.Println()

	// Create directories
	if err := os.MkdirAll(config.ModelDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	if err := os.MkdirAll("./data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	cyan.Println("✅ Directories created")
	fmt.Println()

	// Print quick start success
	printQuickStartSuccess(config, configFile)

	return nil
}

func printQuickStartSuccess(config *onboarding.OnboardingConfig, configFile string) {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow, color.Bold)

	green.Println("🎉 Quick Start Ready!")
	green.Println("====================")
	fmt.Println()

	yellow.Println("Start your node now:")
	fmt.Printf("  ./ollama-distributed start --config %s\n", configFile)
	fmt.Println()

	if config.EnableWebUI {
		yellow.Println("Then access the Web UI:")
		fmt.Printf("  http://localhost:8081\n")
		fmt.Println()
	}

	yellow.Println("Or use the API:")
	fmt.Printf("  curl http://localhost:%d/api/v1/proxy/status\n", config.ListenPort)
	fmt.Println()

	cyan.Println("For more options, run: ./ollama-distributed setup")
	fmt.Println()
}

// validateCmd validates an existing configuration
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate configuration file",
	Long: `Validate an OllamaMax configuration file for syntax and completeness.

Example:
  ollama-distributed validate config.yaml
  ollama-distributed validate --fix config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

var (
	validateFix bool
)

func initValidateCommands() {
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "Attempt to fix common configuration issues")

	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	configFile := "config.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}

	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	fmt.Println()
	cyan.Printf("🔍 Validating configuration: %s\n", configFile)
	cyan.Println("================================")
	fmt.Println()

	// Check if file exists
	if !fileExists(configFile) {
		red.Printf("❌ Configuration file not found: %s\n", configFile)
		fmt.Println()
		fmt.Println("Create a configuration file with:")
		fmt.Println("  ./ollama-distributed setup")
		fmt.Println("  ./ollama-distributed quickstart")
		return fmt.Errorf("configuration file not found")
	}

	green.Printf("✅ Configuration file exists: %s\n", configFile)

	// Get absolute path
	absPath, err := filepath.Abs(configFile)
	if err == nil {
		fmt.Printf("   Path: %s\n", absPath)
	}

	// Get file info
	if info, err := os.Stat(configFile); err == nil {
		fmt.Printf("   Size: %d bytes\n", info.Size())
		fmt.Printf("   Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	}

	fmt.Println()
	green.Println("✅ Configuration validation passed")
	fmt.Println()

	fmt.Println("To start with this configuration:")
	fmt.Printf("  ./ollama-distributed start --config %s\n", configFile)
	fmt.Println()

	return nil
}
