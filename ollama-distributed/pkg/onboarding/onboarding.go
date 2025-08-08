package onboarding

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// OnboardingManager handles user onboarding experience
type OnboardingManager struct {
	reader *bufio.Reader
}

// NewOnboardingManager creates a new onboarding manager
func NewOnboardingManager() *OnboardingManager {
	return &OnboardingManager{
		reader: bufio.NewReader(os.Stdin),
	}
}

// OnboardingConfig holds onboarding configuration
type OnboardingConfig struct {
	NodeName        string
	ListenPort      int
	EnableWebUI     bool
	EnableSecurity  bool
	JoinExisting    bool
	BootstrapPeers  []string
	ModelDirectory  string
	LogLevel        string
}

// RunOnboarding runs the interactive onboarding process
func (om *OnboardingManager) RunOnboarding() (*OnboardingConfig, error) {
	config := &OnboardingConfig{}

	// Welcome message
	om.printWelcome()

	// Basic configuration
	if err := om.configureBasics(config); err != nil {
		return nil, err
	}

	// Network configuration
	if err := om.configureNetwork(config); err != nil {
		return nil, err
	}

	// Security configuration
	if err := om.configureSecurity(config); err != nil {
		return nil, err
	}

	// Advanced configuration
	if err := om.configureAdvanced(config); err != nil {
		return nil, err
	}

	// Summary and confirmation
	if err := om.confirmConfiguration(config); err != nil {
		return nil, err
	}

	return config, nil
}

// printWelcome prints the welcome message
func (om *OnboardingManager) printWelcome() {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	
	fmt.Println()
	cyan.Println("🚀 Welcome to OllamaMax Distributed System!")
	fmt.Println("==========================================")
	fmt.Println()
	green.Println("This setup wizard will help you configure your OllamaMax node.")
	green.Println("You can press Ctrl+C at any time to exit.")
	fmt.Println()
	
	om.waitForEnter("Press Enter to continue...")
}

// configureBasics configures basic settings
func (om *OnboardingManager) configureBasics(config *OnboardingConfig) error {
	blue := color.New(color.FgBlue, color.Bold)
	
	blue.Println("📝 Basic Configuration")
	blue.Println("=====================")
	fmt.Println()

	// Node name
	nodeName, err := om.promptString("Enter a name for this node", "ollama-node-1")
	if err != nil {
		return err
	}
	config.NodeName = nodeName

	// Listen port
	port, err := om.promptInt("Enter the port to listen on", 8080)
	if err != nil {
		return err
	}
	config.ListenPort = port

	// Enable Web UI
	enableWebUI, err := om.promptBool("Enable Web UI", true)
	if err != nil {
		return err
	}
	config.EnableWebUI = enableWebUI

	// Model directory
	modelDir, err := om.promptString("Enter model storage directory", "./models")
	if err != nil {
		return err
	}
	config.ModelDirectory = modelDir

	// Log level
	logLevel, err := om.promptChoice("Select log level", []string{"debug", "info", "warn", "error"}, "info")
	if err != nil {
		return err
	}
	config.LogLevel = logLevel

	fmt.Println()
	return nil
}

// configureNetwork configures network settings
func (om *OnboardingManager) configureNetwork(config *OnboardingConfig) error {
	blue := color.New(color.FgBlue, color.Bold)
	
	blue.Println("🌐 Network Configuration")
	blue.Println("=======================")
	fmt.Println()

	// Join existing cluster
	joinExisting, err := om.promptBool("Join an existing cluster", false)
	if err != nil {
		return err
	}
	config.JoinExisting = joinExisting

	if joinExisting {
		// Bootstrap peers
		fmt.Println("Enter bootstrap peer addresses (one per line, empty line to finish):")
		peers := []string{}
		for {
			peer, err := om.promptString("Peer address", "")
			if err != nil {
				return err
			}
			if peer == "" {
				break
			}
			peers = append(peers, peer)
		}
		config.BootstrapPeers = peers
	}

	fmt.Println()
	return nil
}

// configureSecurity configures security settings
func (om *OnboardingManager) configureSecurity(config *OnboardingConfig) error {
	blue := color.New(color.FgBlue, color.Bold)
	
	blue.Println("🔒 Security Configuration")
	blue.Println("========================")
	fmt.Println()

	// Enable security
	enableSecurity, err := om.promptBool("Enable security features (recommended)", true)
	if err != nil {
		return err
	}
	config.EnableSecurity = enableSecurity

	if enableSecurity {
		yellow := color.New(color.FgYellow)
		yellow.Println("Security features will be automatically configured with secure defaults.")
		yellow.Println("You can customize security settings later in the configuration file.")
	}

	fmt.Println()
	return nil
}

// configureAdvanced configures advanced settings
func (om *OnboardingManager) configureAdvanced(config *OnboardingConfig) error {
	blue := color.New(color.FgBlue, color.Bold)
	
	blue.Println("⚙️  Advanced Configuration")
	blue.Println("=========================")
	fmt.Println()

	// Ask if user wants advanced configuration
	wantAdvanced, err := om.promptBool("Configure advanced settings", false)
	if err != nil {
		return err
	}

	if !wantAdvanced {
		fmt.Println("Using default advanced settings.")
		fmt.Println()
		return nil
	}

	// Advanced settings would go here
	fmt.Println("Advanced configuration options will be available in future versions.")
	fmt.Println()
	return nil
}

// confirmConfiguration shows configuration summary and asks for confirmation
func (om *OnboardingManager) confirmConfiguration(config *OnboardingConfig) error {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	
	green.Println("📋 Configuration Summary")
	green.Println("=======================")
	fmt.Println()

	cyan.Printf("Node Name: %s\n", config.NodeName)
	cyan.Printf("Listen Port: %d\n", config.ListenPort)
	cyan.Printf("Web UI: %s\n", boolToString(config.EnableWebUI))
	cyan.Printf("Security: %s\n", boolToString(config.EnableSecurity))
	cyan.Printf("Join Existing Cluster: %s\n", boolToString(config.JoinExisting))
	if len(config.BootstrapPeers) > 0 {
		cyan.Printf("Bootstrap Peers: %s\n", strings.Join(config.BootstrapPeers, ", "))
	}
	cyan.Printf("Model Directory: %s\n", config.ModelDirectory)
	cyan.Printf("Log Level: %s\n", config.LogLevel)
	fmt.Println()

	// Confirm
	confirmed, err := om.promptBool("Is this configuration correct", true)
	if err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("configuration cancelled by user")
	}

	return nil
}

// Helper methods for prompting user input

func (om *OnboardingManager) promptString(prompt, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := om.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func (om *OnboardingManager) promptInt(prompt string, defaultValue int) (int, error) {
	defaultStr := strconv.Itoa(defaultValue)
	input, err := om.promptString(prompt, defaultStr)
	if err != nil {
		return 0, err
	}

	if input == defaultStr {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}
	return value, nil
}

func (om *OnboardingManager) promptBool(prompt string, defaultValue bool) (bool, error) {
	defaultStr := "y"
	if !defaultValue {
		defaultStr = "n"
	}

	input, err := om.promptString(prompt+" (y/n)", defaultStr)
	if err != nil {
		return false, err
	}

	input = strings.ToLower(input)
	switch input {
	case "y", "yes", "true", "1":
		return true, nil
	case "n", "no", "false", "0":
		return false, nil
	default:
		return defaultValue, nil
	}
}

func (om *OnboardingManager) promptChoice(prompt string, choices []string, defaultValue string) (string, error) {
	fmt.Printf("%s (%s) [%s]: ", prompt, strings.Join(choices, "/"), defaultValue)

	input, err := om.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	// Validate choice
	for _, choice := range choices {
		if strings.EqualFold(input, choice) {
			return choice, nil
		}
	}

	return "", fmt.Errorf("invalid choice: %s (must be one of: %s)", input, strings.Join(choices, ", "))
}

func (om *OnboardingManager) waitForEnter(prompt string) {
	fmt.Print(prompt)
	om.reader.ReadString('\n')
}

func boolToString(b bool) string {
	if b {
		return "Enabled"
	}
	return "Disabled"
}

// GenerateConfigFile generates a configuration file from onboarding config
func (om *OnboardingManager) GenerateConfigFile(config *OnboardingConfig, filename string) error {
	configContent := fmt.Sprintf(`# OllamaMax Configuration
# Generated by onboarding wizard on %s

node:
  name: "%s"
  data_dir: "./data"

api:
  listen_address: ":%d"
  enable_cors: true
  request_timeout: "30s"

p2p:
  listen_address: "/ip4/0.0.0.0/tcp/9000"
  bootstrap_peers: [%s]

consensus:
  algorithm: "raft"
  election_timeout: "5s"
  heartbeat_interval: "1s"

scheduler:
  algorithm: "round_robin"
  health_check_interval: "30s"
  max_retries: 3

models:
  storage_path: "%s"
  cache_size: "1GB"
  auto_pull: true

logging:
  level: "%s"
  format: "json"
  output: "stdout"

security:
  enabled: %t
  jwt_secret: "auto-generated"
  session_timeout: "24h"

web:
  enabled: %t
  listen_address: ":8081"
  static_path: "./web"
`,
		time.Now().Format("2006-01-02 15:04:05"),
		config.NodeName,
		config.ListenPort,
		formatBootstrapPeers(config.BootstrapPeers),
		config.ModelDirectory,
		config.LogLevel,
		config.EnableSecurity,
		config.EnableWebUI,
	)

	return os.WriteFile(filename, []byte(configContent), 0644)
}

func formatBootstrapPeers(peers []string) string {
	if len(peers) == 0 {
		return ""
	}

	quoted := make([]string, len(peers))
	for i, peer := range peers {
		quoted[i] = fmt.Sprintf(`"%s"`, peer)
	}
	return strings.Join(quoted, ", ")
}
