package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

var (
	configFile string
	outputFile string
	environment string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "config-tool",
		Short: "OllamaMax configuration management tool",
		Long:  "A tool for managing OllamaMax distributed configuration files",
	}

	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  "Validate a configuration file for syntax and semantic errors",
		RunE:  validateConfig,
	}

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate default configuration",
		Long:  "Generate a default configuration file for the specified environment",
		RunE:  generateConfig,
	}

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show configuration",
		Long:  "Display the current configuration with environment variable substitution",
		RunE:  showConfig,
	}

	// Add flags
	validateCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file to validate")
	validateCmd.MarkFlagRequired("config")

	generateCmd.Flags().StringVarP(&environment, "env", "e", "development", "Environment (development, testing, production)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")

	showCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file to show")

	// Add commands
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(showCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func validateConfig(cmd *cobra.Command, args []string) error {
	fmt.Printf("Validating configuration file: %s\n", configFile)

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Basic validation
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// Extended validation
	if err := cfg.ValidateExtended(); err != nil {
		return fmt.Errorf("extended validation failed: %w", err)
	}

	fmt.Println("✅ Configuration is valid!")
	return nil
}

func generateConfig(cmd *cobra.Command, args []string) error {
	fmt.Printf("Generating %s configuration...\n", environment)

	// Get default configuration
	cfg := config.DefaultConfig()
	
	// Customize for environment
	switch environment {
	case "development":
		cfg.Node.Environment = "development"
		cfg.Security.Auth.Enabled = false
		cfg.API.TLS.Enabled = false
		cfg.Logging.Level = "debug"
		cfg.Metrics.Enabled = true
	case "testing":
		cfg.Node.Environment = "testing"
		cfg.Security.Auth.Enabled = false
		cfg.API.TLS.Enabled = false
		cfg.Logging.Level = "error"
		cfg.Metrics.Enabled = false
	case "production":
		cfg.Node.Environment = "production"
		cfg.Security.Auth.Enabled = true
		cfg.API.TLS.Enabled = true
		cfg.Logging.Level = "info"
		cfg.Logging.Format = "json"
		cfg.Metrics.Enabled = true
	default:
		return fmt.Errorf("unsupported environment: %s", environment)
	}

	// Save or print configuration
	if outputFile != "" {
		if err := cfg.Save(outputFile); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		fmt.Printf("Configuration saved to: %s\n", outputFile)
	} else {
		// Print to stdout (simplified YAML output)
		fmt.Printf(`# %s Configuration for OllamaMax Distributed
node:
  id: "%s"
  name: "%s"
  environment: "%s"

api:
  listen: "%s"
  tls:
    enabled: %t

security:
  auth:
    enabled: %t

logging:
  level: "%s"
  format: "%s"

metrics:
  enabled: %t
`,
			environment,
			cfg.Node.ID,
			cfg.Node.Name,
			cfg.Node.Environment,
			cfg.API.Listen,
			cfg.API.TLS.Enabled,
			cfg.Security.Auth.Enabled,
			cfg.Logging.Level,
			cfg.Logging.Format,
			cfg.Metrics.Enabled,
		)
	}

	return nil
}

func showConfig(cmd *cobra.Command, args []string) error {
	var cfg *config.Config
	var err error

	if configFile != "" {
		fmt.Printf("Loading configuration from: %s\n", configFile)
		cfg, err = config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	} else {
		fmt.Println("Using default configuration")
		cfg = config.DefaultConfig()
	}

	// Display key configuration values
	fmt.Printf(`
Configuration Summary:
=====================
Node ID:          %s
Environment:      %s
API Listen:       %s
TLS Enabled:      %t
Auth Enabled:     %t
P2P Listen:       %s
Data Directory:   %s
Log Level:        %s
Metrics Enabled:  %t
`,
		cfg.Node.ID,
		cfg.Node.Environment,
		cfg.API.Listen,
		cfg.API.TLS.Enabled,
		cfg.Security.Auth.Enabled,
		cfg.P2P.Listen,
		cfg.Storage.DataDir,
		cfg.Logging.Level,
		cfg.Metrics.Enabled,
	)

	return nil
}
