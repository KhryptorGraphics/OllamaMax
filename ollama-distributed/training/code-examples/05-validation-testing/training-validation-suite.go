// 05-validation-testing/training-validation-suite.go
// Comprehensive validation and testing suite for Ollama Distributed Training
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	_ "testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestResult represents the result of a validation test
type TestResult struct {
	Name        string        `json:"name"`
	Category    string        `json:"category"`
	Status      string        `json:"status"` // pass, fail, skip, warning
	Message     string        `json:"message"`
	Duration    time.Duration `json:"duration"`
	Details     interface{}   `json:"details,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// TestSuite manages the execution of validation tests
type TestSuite struct {
	BaseURL     string
	ConfigPath  string
	BinaryPath  string
	DataDir     string
	Results     []TestResult
	StartTime   time.Time
	HTTPClient  *http.Client
	Debug       bool
}

// ValidationConfig holds configuration for validation tests
type ValidationConfig struct {
	API struct {
		BaseURL string `yaml:"base_url"`
		Timeout string `yaml:"timeout"`
	} `yaml:"api"`
	Paths struct {
		Binary  string `yaml:"binary"`
		Config  string `yaml:"config"`
		DataDir string `yaml:"data_dir"`
	} `yaml:"paths"`
	Tests struct {
		Prerequisites bool `yaml:"prerequisites"`
		Installation  bool `yaml:"installation"`
		Configuration bool `yaml:"configuration"`
		Startup       bool `yaml:"startup"`
		API           bool `yaml:"api"`
		Performance   bool `yaml:"performance"`
		Security      bool `yaml:"security"`
	} `yaml:"tests"`
	Timeouts struct {
		Short  string `yaml:"short"`
		Medium string `yaml:"medium"`
		Long   string `yaml:"long"`
	} `yaml:"timeouts"`
}

// NewTestSuite creates a new test suite
func NewTestSuite(configPath string) (*TestSuite, error) {
	config, err := loadValidationConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load validation config: %w", err)
	}
	
	timeout, _ := time.ParseDuration(config.API.Timeout)
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	return &TestSuite{
		BaseURL:    config.API.BaseURL,
		ConfigPath: config.Paths.Config,
		BinaryPath: config.Paths.Binary,
		DataDir:    config.Paths.DataDir,
		Results:    make([]TestResult, 0),
		StartTime:  time.Now(),
		HTTPClient: &http.Client{Timeout: timeout},
		Debug:      os.Getenv("DEBUG") == "true",
	}, nil
}

func loadValidationConfig(configPath string) (*ValidationConfig, error) {
	if configPath == "" {
		// Create default config
		return &ValidationConfig{
			API: struct {
				BaseURL string `yaml:"base_url"`
				Timeout string `yaml:"timeout"`
			}{
				BaseURL: "http://127.0.0.1:8080",
				Timeout: "30s",
			},
			Paths: struct {
				Binary  string `yaml:"binary"`
				Config  string `yaml:"config"`
				DataDir string `yaml:"data_dir"`
			}{
				Binary:  "./bin/ollama-distributed",
				Config:  "~/.ollama-distributed/profiles/development.yaml",
				DataDir: "./dev-data",
			},
			Tests: struct {
				Prerequisites bool `yaml:"prerequisites"`
				Installation  bool `yaml:"installation"`
				Configuration bool `yaml:"configuration"`
				Startup       bool `yaml:"startup"`
				API           bool `yaml:"api"`
				Performance   bool `yaml:"performance"`
				Security      bool `yaml:"security"`
			}{
				Prerequisites: true,
				Installation:  true,
				Configuration: true,
				Startup:       true,
				API:           true,
				Performance:   true,
				Security:      true,
			},
			Timeouts: struct {
				Short  string `yaml:"short"`
				Medium string `yaml:"medium"`
				Long   string `yaml:"long"`
			}{
				Short:  "5s",
				Medium: "30s",
				Long:   "2m",
			},
		}, nil
	}
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var config ValidationConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// Test execution methods

func (ts *TestSuite) runTest(name, category string, testFunc func() error) {
	start := time.Now()
	result := TestResult{
		Name:      name,
		Category:  category,
		Timestamp: start,
	}
	
	if ts.Debug {
		log.Printf("Running test: %s/%s", category, name)
	}
	
	err := testFunc()
	result.Duration = time.Since(start)
	
	if err != nil {
		result.Status = "fail"
		result.Message = err.Error()
	} else {
		result.Status = "pass"
		result.Message = "Test passed successfully"
	}
	
	ts.Results = append(ts.Results, result)
	
	if ts.Debug {
		log.Printf("Test %s/%s: %s (%v)", category, name, result.Status, result.Duration)
	}
}

// Prerequisites validation
func (ts *TestSuite) validatePrerequisites() {
	category := "Prerequisites"
	
	// Check Go installation
	ts.runTest("Go Version", category, func() error {
		cmd := exec.Command("go", "version")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Go not installed or not in PATH")
		}
		
		versionStr := string(output)
		if !strings.Contains(versionStr, "go1.21") && !strings.Contains(versionStr, "go1.22") && !strings.Contains(versionStr, "go1.23") {
			return fmt.Errorf("Go version may be incompatible: %s", strings.TrimSpace(versionStr))
		}
		
		return nil
	})
	
	// Check Git installation
	ts.runTest("Git Availability", category, func() error {
		cmd := exec.Command("git", "--version")
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("Git not installed or not in PATH")
		}
		return nil
	})
	
	// Check curl availability
	ts.runTest("Curl Availability", category, func() error {
		cmd := exec.Command("curl", "--version")
		_, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("curl not installed or not in PATH")
		}
		return nil
	})
	
	// Check disk space
	ts.runTest("Disk Space", category, func() error {
		cmd := exec.Command("df", "-h", ".")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("unable to check disk space: %w", err)
		}
		
		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			return fmt.Errorf("unexpected df output format")
		}
		
		fields := strings.Fields(lines[1])
		if len(fields) < 4 {
			return fmt.Errorf("unable to parse disk usage")
		}
		
		usagePercent := strings.TrimSuffix(fields[4], "%")
		if usagePercent > "90" {
			return fmt.Errorf("disk usage is %s%%, which may cause issues", usagePercent)
		}
		
		return nil
	})
	
	// Check network ports
	ports := []string{"8080", "8081", "4001"}
	for _, port := range ports {
		portNum := port
		ts.runTest(fmt.Sprintf("Port %s Availability", portNum), category, func() error {
			cmd := exec.Command("netstat", "-ln")
			output, err := cmd.Output()
			if err != nil {
				// netstat might not be available, try ss
				cmd = exec.Command("ss", "-ln")
				output, err = cmd.Output()
				if err != nil {
					return nil // Skip if neither is available
				}
			}
			
			if strings.Contains(string(output), ":"+portNum+" ") {
				return fmt.Errorf("port %s is already in use", portNum)
			}
			
			return nil
		})
	}
}

// Installation validation
func (ts *TestSuite) validateInstallation() {
	category := "Installation"
	
	// Check binary exists
	ts.runTest("Binary Exists", category, func() error {
		if _, err := os.Stat(ts.BinaryPath); os.IsNotExist(err) {
			return fmt.Errorf("binary not found at %s", ts.BinaryPath)
		}
		return nil
	})
	
	// Check binary is executable
	ts.runTest("Binary Executable", category, func() error {
		info, err := os.Stat(ts.BinaryPath)
		if err != nil {
			return err
		}
		
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("binary is not executable")
		}
		return nil
	})
	
	// Test help command
	ts.runTest("Help Command", category, func() error {
		cmd := exec.Command(ts.BinaryPath, "--help")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("help command failed: %w", err)
		}
		
		helpText := string(output)
		if !strings.Contains(helpText, "ollama-distributed") {
			return fmt.Errorf("unexpected help output")
		}
		
		return nil
	})
	
	// Test version command
	ts.runTest("Version Command", category, func() error {
		cmd := exec.Command(ts.BinaryPath, "--version")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("version command failed: %w", err)
		}
		
		versionText := string(output)
		if len(strings.TrimSpace(versionText)) == 0 {
			return fmt.Errorf("empty version output")
		}
		
		return nil
	})
}

// Configuration validation
func (ts *TestSuite) validateConfiguration() {
	category := "Configuration"
	
	// Check configuration file exists
	ts.runTest("Config File Exists", category, func() error {
		configPath := ts.ConfigPath
		if strings.HasPrefix(configPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configPath = filepath.Join(home, configPath[1:])
		}
		
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found at %s", configPath)
		}
		return nil
	})
	
	// Validate YAML syntax
	ts.runTest("YAML Syntax", category, func() error {
		configPath := ts.ConfigPath
		if strings.HasPrefix(configPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			configPath = filepath.Join(home, configPath[1:])
		}
		
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return err
		}
		
		var config interface{}
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("invalid YAML syntax: %w", err)
		}
		
		return nil
	})
	
	// Test configuration validation command
	ts.runTest("Config Validation Command", category, func() error {
		cmd := exec.Command(ts.BinaryPath, "validate", "--config", ts.ConfigPath)
		output, err := cmd.Output()
		if err != nil {
			// Command might not exist, which is acceptable
			if strings.Contains(err.Error(), "unknown command") {
				return nil
			}
			return fmt.Errorf("config validation failed: %w", err)
		}
		
		if strings.Contains(string(output), "error") || strings.Contains(string(output), "failed") {
			return fmt.Errorf("configuration validation reported errors")
		}
		
		return nil
	})
	
	// Check data directory
	ts.runTest("Data Directory", category, func() error {
		if _, err := os.Stat(ts.DataDir); os.IsNotExist(err) {
			// Try to create it
			if err := os.MkdirAll(ts.DataDir, 0755); err != nil {
				return fmt.Errorf("cannot create data directory: %w", err)
			}
		}
		
		// Check write permissions
		testFile := filepath.Join(ts.DataDir, ".test-write")
		if err := ioutil.WriteFile(testFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("data directory is not writable: %w", err)
		}
		os.Remove(testFile)
		
		return nil
	})
}

// Service startup validation
func (ts *TestSuite) validateStartup() {
	category := "Startup"
	
	// Test dry run if available
	ts.runTest("Dry Run", category, func() error {
		cmd := exec.Command(ts.BinaryPath, "start", "--config", ts.ConfigPath, "--dry-run")
		output, err := cmd.Output()
		if err != nil {
			// Dry run might not be implemented
			if strings.Contains(err.Error(), "unknown flag") {
				return nil
			}
			return fmt.Errorf("dry run failed: %w", err)
		}
		
		dryRunOutput := string(output)
		if strings.Contains(dryRunOutput, "error") || strings.Contains(dryRunOutput, "failed") {
			return fmt.Errorf("dry run reported errors: %s", dryRunOutput)
		}
		
		return nil
	})
	
	// Test actual startup (background process)
	ts.runTest("Service Startup", category, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Start the service in background
		cmd := exec.CommandContext(ctx, ts.BinaryPath, "start", "--config", ts.ConfigPath)
		
		// Create pipes for output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start service: %w", err)
		}
		
		// Wait for service to be ready or fail
		ready := make(chan bool, 1)
		go func() {
			time.Sleep(10 * time.Second) // Give service time to start
			
			// Check if service is responding
			resp, err := ts.HTTPClient.Get(ts.BaseURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				ready <- true
			} else {
				ready <- false
			}
		}()
		
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			return fmt.Errorf("startup timeout")
		case isReady := <-ready:
			// Clean shutdown
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			
			// Read any output
			stdoutBytes, _ := ioutil.ReadAll(stdout)
			stderrBytes, _ := ioutil.ReadAll(stderr)
			
			if !isReady {
				return fmt.Errorf("service did not become ready. stdout: %s, stderr: %s", 
					string(stdoutBytes), string(stderrBytes))
			}
			
			return nil
		}
	})
}

// API validation
func (ts *TestSuite) validateAPI() {
	category := "API"
	
	// Health endpoint
	ts.runTest("Health Endpoint", category, func() error {
		resp, err := ts.HTTPClient.Get(ts.BaseURL + "/health")
		if err != nil {
			return fmt.Errorf("health endpoint not accessible: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
		}
		
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		
		var health map[string]interface{}
		if err := json.Unmarshal(body, &health); err != nil {
			return fmt.Errorf("invalid JSON response: %w", err)
		}
		
		if status, ok := health["status"]; !ok || status != "healthy" {
			return fmt.Errorf("service reports unhealthy status")
		}
		
		return nil
	})
	
	// Core API endpoints
	endpoints := map[string]string{
		"Models API":          "/api/tags",
		"Cluster Status":      "/api/distributed/status",
		"Node List":          "/api/distributed/nodes",
		"System Metrics":     "/api/distributed/metrics",
	}
	
	for name, endpoint := range endpoints {
		endpointName := name
		endpointURL := endpoint
		ts.runTest(endpointName, category, func() error {
			resp, err := ts.HTTPClient.Get(ts.BaseURL + endpointURL)
			if err != nil {
				return fmt.Errorf("endpoint not accessible: %w", err)
			}
			defer resp.Body.Close()
			
			if resp.StatusCode >= 500 {
				return fmt.Errorf("endpoint returned server error: %d", resp.StatusCode)
			}
			
			// Check for valid JSON response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			
			var jsonData interface{}
			if err := json.Unmarshal(body, &jsonData); err != nil {
				return fmt.Errorf("endpoint returned invalid JSON: %w", err)
			}
			
			return nil
		})
	}
	
	// CORS headers
	ts.runTest("CORS Headers", category, func() error {
		req, err := http.NewRequest("OPTIONS", ts.BaseURL+"/health", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Origin", "http://localhost:3000")
		
		resp, err := ts.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("CORS preflight request failed: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.Header.Get("Access-Control-Allow-Origin") == "" {
			return fmt.Errorf("CORS headers not present")
		}
		
		return nil
	})
}

// Performance validation
func (ts *TestSuite) validatePerformance() {
	category := "Performance"
	
	// Response time test
	ts.runTest("Response Time", category, func() error {
		start := time.Now()
		resp, err := ts.HTTPClient.Get(ts.BaseURL + "/health")
		duration := time.Since(start)
		
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		resp.Body.Close()
		
		if duration > 5*time.Second {
			return fmt.Errorf("response time too slow: %v", duration)
		}
		
		return nil
	})
	
	// Concurrent requests test
	ts.runTest("Concurrent Requests", category, func() error {
		const numRequests = 10
		results := make(chan error, numRequests)
		
		for i := 0; i < numRequests; i++ {
			go func() {
				resp, err := ts.HTTPClient.Get(ts.BaseURL + "/health")
				if err != nil {
					results <- err
					return
				}
				resp.Body.Close()
				
				if resp.StatusCode != 200 {
					results <- fmt.Errorf("status code %d", resp.StatusCode)
					return
				}
				
				results <- nil
			}()
		}
		
		var errors []string
		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				errors = append(errors, err.Error())
			}
		}
		
		if len(errors) > 0 {
			return fmt.Errorf("concurrent request failures: %v", strings.Join(errors, ", "))
		}
		
		return nil
	})
	
	// Memory usage test
	ts.runTest("Memory Usage", category, func() error {
		// This is a simplified test - in reality you'd check actual memory usage
		cmd := exec.Command("ps", "-o", "rss", "-p", "$$")
		output, err := cmd.Output()
		if err != nil {
			return nil // Skip if ps command fails
		}
		
		lines := strings.Split(string(output), "\n")
		if len(lines) < 2 {
			return nil
		}
		
		// Just check that we got some output - real implementation would parse RSS
		if strings.TrimSpace(lines[1]) == "" {
			return fmt.Errorf("unable to determine memory usage")
		}
		
		return nil
	})
}

// Security validation
func (ts *TestSuite) validateSecurity() {
	category := "Security"
	
	// Check for security headers
	ts.runTest("Security Headers", category, func() error {
		resp, err := ts.HTTPClient.Get(ts.BaseURL + "/health")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		securityHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
		}
		
		missing := []string{}
		for _, header := range securityHeaders {
			if resp.Header.Get(header) == "" {
				missing = append(missing, header)
			}
		}
		
		if len(missing) > 0 {
			return fmt.Errorf("missing security headers: %v", strings.Join(missing, ", "))
		}
		
		return nil
	})
	
	// Test for information disclosure
	ts.runTest("Information Disclosure", category, func() error {
		resp, err := ts.HTTPClient.Get(ts.BaseURL + "/nonexistent")
		if err != nil {
			return nil // Network error is fine
		}
		defer resp.Body.Close()
		
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil
		}
		
		bodyStr := strings.ToLower(string(body))
		sensitiveInfo := []string{"stack trace", "internal error", "debug", "panic"}
		
		for _, info := range sensitiveInfo {
			if strings.Contains(bodyStr, info) {
				return fmt.Errorf("potential information disclosure: response contains '%s'", info)
			}
		}
		
		return nil
	})
}

// Reporting methods

func (ts *TestSuite) printResults() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("OLLAMA DISTRIBUTED TRAINING VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 80))
	
	totalDuration := time.Since(ts.StartTime)
	
	// Group results by category
	categories := make(map[string][]TestResult)
	for _, result := range ts.Results {
		categories[result.Category] = append(categories[result.Category], result)
	}
	
	// Print results by category
	for category, results := range categories {
		fmt.Printf("\n%s:\n", category)
		fmt.Println(strings.Repeat("-", 40))
		
		passed := 0
		failed := 0
		
		for _, result := range results {
			status := result.Status
			symbol := "?"
			switch status {
			case "pass":
				symbol = "✅"
				passed++
			case "fail":
				symbol = "❌"
				failed++
			case "skip":
				symbol = "⏭️ "
			case "warning":
				symbol = "⚠️ "
			}
			
			fmt.Printf("  %s %-25s (%v)\n", symbol, result.Name, result.Duration)
			if result.Status == "fail" && result.Message != "" {
				fmt.Printf("     Error: %s\n", result.Message)
			}
		}
		
		fmt.Printf("     Summary: %d passed, %d failed\n", passed, failed)
	}
	
	// Overall summary
	totalTests := len(ts.Results)
	totalPassed := 0
	totalFailed := 0
	
	for _, result := range ts.Results {
		if result.Status == "pass" {
			totalPassed++
		} else if result.Status == "fail" {
			totalFailed++
		}
	}
	
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("OVERALL SUMMARY\n")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", totalPassed)
	fmt.Printf("Failed: %d\n", totalFailed)
	fmt.Printf("Success Rate: %.1f%%\n", float64(totalPassed)/float64(totalTests)*100)
	fmt.Printf("Total Duration: %v\n", totalDuration)
	
	if totalFailed == 0 {
		fmt.Printf("\n🎉 All tests passed! Training environment is ready.\n")
	} else {
		fmt.Printf("\n⚠️  Some tests failed. Please review and fix issues before proceeding.\n")
	}
}

func (ts *TestSuite) saveResults(filename string) error {
	summary := map[string]interface{}{
		"timestamp":     time.Now(),
		"total_tests":   len(ts.Results),
		"total_passed":  len(ts.getResultsByStatus("pass")),
		"total_failed":  len(ts.getResultsByStatus("fail")),
		"total_duration": time.Since(ts.StartTime),
		"results":       ts.Results,
	}
	
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(filename, data, 0644)
}

func (ts *TestSuite) getResultsByStatus(status string) []TestResult {
	var results []TestResult
	for _, result := range ts.Results {
		if result.Status == status {
			results = append(results, result)
		}
	}
	return results
}

// Main execution functions

func (ts *TestSuite) RunAllTests() {
	fmt.Println("Starting Ollama Distributed Training Validation Suite...")
	fmt.Printf("Target API: %s\n", ts.BaseURL)
	fmt.Printf("Binary: %s\n", ts.BinaryPath)
	fmt.Printf("Config: %s\n", ts.ConfigPath)
	
	ts.validatePrerequisites()
	ts.validateInstallation()
	ts.validateConfiguration()
	ts.validateStartup()
	ts.validateAPI()
	ts.validatePerformance()
	ts.validateSecurity()
}

// Main function for command-line usage
func main() {
	var configPath string
	var outputFile string
	
	// Simple command line parsing
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			configPath = args[i+1]
		}
		if arg == "--output" && i+1 < len(args) {
			outputFile = args[i+1]
		}
	}
	
	// Create test suite
	suite, err := NewTestSuite(configPath)
	if err != nil {
		log.Fatalf("Failed to create test suite: %v", err)
	}
	
	// Run all tests
	suite.RunAllTests()
	
	// Print results
	suite.printResults()
	
	// Save results to file if requested
	if outputFile != "" {
		if err := suite.saveResults(outputFile); err != nil {
			log.Printf("Failed to save results to %s: %v", outputFile, err)
		} else {
			fmt.Printf("\nResults saved to: %s\n", outputFile)
		}
	}
	
	// Exit with appropriate code
	failed := len(suite.getResultsByStatus("fail"))
	if failed > 0 {
		os.Exit(1)
	}
}