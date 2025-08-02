package integration

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/internal/config"
)

// SimpleOllamaIntegration provides basic Ollama integration
type SimpleOllamaIntegration struct {
	config    *config.Config
	ollamaCmd *exec.Cmd
	started   bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewSimpleOllamaIntegration creates a new simple Ollama integration
func NewSimpleOllamaIntegration(cfg *config.Config) *SimpleOllamaIntegration {
	ctx, cancel := context.WithCancel(context.Background())

	return &SimpleOllamaIntegration{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts the Ollama integration
func (soi *SimpleOllamaIntegration) Start() error {
	soi.mu.Lock()
	defer soi.mu.Unlock()

	if soi.started {
		return fmt.Errorf("Ollama integration already started")
	}

	// Check if Ollama is available
	if !soi.isOllamaAvailable() {
		fmt.Printf("⚠️  Ollama not found in PATH. Please install Ollama first.\n")
		fmt.Printf("   Visit: https://ollama.com/download\n")
		return fmt.Errorf("Ollama not available")
	}

	// Start Ollama server
	if err := soi.startOllamaServer(); err != nil {
		return fmt.Errorf("failed to start Ollama server: %w", err)
	}

	soi.started = true
	fmt.Printf("✅ Ollama integration started successfully\n")
	fmt.Printf("   Ollama API: http://localhost:11434\n")
	fmt.Printf("   Distributed API: %s\n", soi.GetDistributedAPIURL())

	return nil
}

// isOllamaAvailable checks if Ollama is available in the system
func (soi *SimpleOllamaIntegration) isOllamaAvailable() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

// startOllamaServer starts the Ollama server
func (soi *SimpleOllamaIntegration) startOllamaServer() error {
	// Check if Ollama is already running
	if soi.isOllamaRunning() {
		fmt.Printf("ℹ️  Ollama server already running\n")
		return nil
	}

	// Start Ollama serve command
	soi.ollamaCmd = exec.CommandContext(soi.ctx, "ollama", "serve")

	// Set environment variables for Ollama
	soi.ollamaCmd.Env = append(soi.ollamaCmd.Env,
		"OLLAMA_HOST=127.0.0.1:11434",
		"OLLAMA_KEEP_ALIVE=5m",
	)

	if err := soi.ollamaCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Ollama server: %w", err)
	}

	// Wait for Ollama to be ready
	if err := soi.waitForOllamaReady(); err != nil {
		return fmt.Errorf("Ollama server failed to start: %w", err)
	}

	fmt.Printf("✅ Ollama server started successfully\n")

	// Monitor Ollama process
	go soi.monitorOllamaProcess()

	return nil
}

// isOllamaRunning checks if Ollama is already running
func (soi *SimpleOllamaIntegration) isOllamaRunning() bool {
	// Try to connect to Ollama API
	cmd := exec.Command("curl", "-s", "http://localhost:11434/api/tags")
	err := cmd.Run()
	return err == nil
}

// waitForOllamaReady waits for Ollama to be ready
func (soi *SimpleOllamaIntegration) waitForOllamaReady() error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Ollama to be ready")
		case <-ticker.C:
			if soi.isOllamaRunning() {
				return nil
			}
		}
	}
}

// monitorOllamaProcess monitors the Ollama process
func (soi *SimpleOllamaIntegration) monitorOllamaProcess() {
	if soi.ollamaCmd == nil {
		return
	}

	// Wait for the process to exit
	err := soi.ollamaCmd.Wait()
	if err != nil {
		fmt.Printf("⚠️  Ollama process exited with error: %v\n", err)
	} else {
		fmt.Printf("ℹ️  Ollama process exited normally\n")
	}

	soi.mu.Lock()
	soi.started = false
	soi.mu.Unlock()
}

// GetStatus returns the integration status
func (soi *SimpleOllamaIntegration) GetStatus() map[string]interface{} {
	soi.mu.RLock()
	defer soi.mu.RUnlock()

	status := map[string]interface{}{
		"integration_started": soi.started,
		"ollama_available":    soi.isOllamaAvailable(),
		"ollama_running":      soi.isOllamaRunning(),
		"timestamp":           time.Now(),
	}

	if soi.ollamaCmd != nil && soi.ollamaCmd.Process != nil {
		status["ollama_pid"] = soi.ollamaCmd.Process.Pid
	}

	return status
}

// PullModel pulls a model using Ollama
func (soi *SimpleOllamaIntegration) PullModel(modelName string) error {
	if !soi.isOllamaRunning() {
		return fmt.Errorf("Ollama is not running")
	}

	fmt.Printf("📥 Pulling model: %s\n", modelName)

	cmd := exec.Command("ollama", "pull", modelName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull model %s: %w", modelName, err)
	}

	fmt.Printf("✅ Model pulled successfully: %s\n", modelName)
	return nil
}

// ListModels lists available models
func (soi *SimpleOllamaIntegration) ListModels() ([]string, error) {
	if !soi.isOllamaRunning() {
		return nil, fmt.Errorf("Ollama is not running")
	}

	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Parse output to extract model names
	// This is a simplified implementation
	models := []string{}
	lines := string(output)
	if len(lines) > 0 {
		// TODO: Parse the actual output format
		models = append(models, "Available models listed via 'ollama list'")
	}

	return models, nil
}

// RunModel runs a model with a prompt
func (soi *SimpleOllamaIntegration) RunModel(modelName, prompt string) (string, error) {
	if !soi.isOllamaRunning() {
		return "", fmt.Errorf("Ollama is not running")
	}

	cmd := exec.Command("ollama", "run", modelName, prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run model: %w", err)
	}

	return string(output), nil
}

// Shutdown gracefully shuts down the integration
func (soi *SimpleOllamaIntegration) Shutdown() error {
	soi.mu.Lock()
	defer soi.mu.Unlock()

	if !soi.started {
		return nil
	}

	fmt.Printf("🛑 Shutting down Ollama integration\n")

	// Cancel context to stop monitoring
	soi.cancel()

	// Stop Ollama process if we started it
	if soi.ollamaCmd != nil && soi.ollamaCmd.Process != nil {
		if err := soi.ollamaCmd.Process.Kill(); err != nil {
			fmt.Printf("⚠️  Failed to kill Ollama process: %v\n", err)
		} else {
			fmt.Printf("✅ Ollama process stopped\n")
		}
	}

	soi.started = false
	return nil
}

// InstallDefaultModel installs a default model for testing
func (soi *SimpleOllamaIntegration) InstallDefaultModel() error {
	fmt.Printf("📦 Installing default model for testing...\n")

	// Try to pull a small model for testing
	defaultModels := []string{
		"llama3.2:1b", // Small 1B model
		"phi3:mini",   // Alternative small model
		"gemma2:2b",   // Another small option
	}

	for _, model := range defaultModels {
		fmt.Printf("   Trying to pull: %s\n", model)
		if err := soi.PullModel(model); err == nil {
			fmt.Printf("✅ Default model installed: %s\n", model)
			return nil
		}
	}

	fmt.Printf("⚠️  Could not install any default model. You can manually install one with:\n")
	fmt.Printf("   ollama pull llama3.2:1b\n")

	return nil
}

// TestIntegration tests the Ollama integration
func (soi *SimpleOllamaIntegration) TestIntegration() error {
	fmt.Printf("🧪 Testing Ollama integration...\n")

	// Check if Ollama is running
	if !soi.isOllamaRunning() {
		return fmt.Errorf("Ollama is not running")
	}

	// List models
	models, err := soi.ListModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	fmt.Printf("✅ Ollama integration test passed\n")
	fmt.Printf("   Available models: %d\n", len(models))

	return nil
}

// GetOllamaAPIURL returns the Ollama API URL
func (soi *SimpleOllamaIntegration) GetOllamaAPIURL() string {
	return "http://localhost:11434"
}

// GetDistributedAPIURL returns the distributed API URL
func (soi *SimpleOllamaIntegration) GetDistributedAPIURL() string {
	return fmt.Sprintf("http://%s", soi.config.API.Listen)
}

// IsIntegrationComplete returns whether the integration is complete
func (soi *SimpleOllamaIntegration) IsIntegrationComplete() bool {
	return soi.isOllamaAvailable() && soi.isOllamaRunning() && soi.started
}
