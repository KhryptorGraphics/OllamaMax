package integration

import (
	"testing"
	"time"
)

// TestComprehensiveIntegration runs the complete integration test suite
func TestComprehensiveIntegration(t *testing.T) {
	// Initialize test framework
	framework := NewIntegrationTestFramework("./ollama-distributed", "http://localhost:8080")
	defer framework.Teardown(t)

	// Setup test environment
	if err := framework.Setup(t); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	// Track test results
	results := make(map[string]bool)

	// Run test suites
	t.Run("BasicFunctionality", func(t *testing.T) {
		results["BasicFunctionality"] = t.Run("TestBasicCLI", func(t *testing.T) {
			testBasicCLI(t, framework)
		})
	})

	t.Run("ProxyCommands", func(t *testing.T) {
		results["ProxyStatus"] = t.Run("TestProxyStatus", func(t *testing.T) {
			framework.TestProxyCommand(t, "status")
		})
		
		results["ProxyInstances"] = t.Run("TestProxyInstances", func(t *testing.T) {
			framework.TestProxyCommand(t, "instances")
		})
		
		results["ProxyMetrics"] = t.Run("TestProxyMetrics", func(t *testing.T) {
			framework.TestProxyCommand(t, "metrics")
		})
	})

	t.Run("APIEndpoints", func(t *testing.T) {
		results["HealthEndpoint"] = t.Run("TestHealthEndpoint", func(t *testing.T) {
			framework.TestAPIEndpoint(t, "/health", 200)
		})
		
		results["ProxyStatusAPI"] = t.Run("TestProxyStatusAPI", func(t *testing.T) {
			framework.TestAPIEndpoint(t, "/api/v1/proxy/status", 200)
		})
	})

	t.Run("JSONOutput", func(t *testing.T) {
		results["JSONValidation"] = t.Run("TestJSONValidation", func(t *testing.T) {
			testJSONValidation(t, framework)
		})
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		results["ErrorScenarios"] = t.Run("TestErrorScenarios", func(t *testing.T) {
			testErrorScenarios(t, framework)
		})
	})

	t.Run("Performance", func(t *testing.T) {
		results["PerformanceTests"] = t.Run("TestPerformance", func(t *testing.T) {
			testPerformance(t, framework)
		})
	})

	// Generate comprehensive report
	framework.GenerateTestReport(t, results)
}

// testBasicCLI tests basic CLI functionality
func testBasicCLI(t *testing.T, framework *IntegrationTestFramework) {
	t.Log("🎛️ Testing basic CLI functionality...")

	// Test main help
	output, err := framework.RunCLICommand("--help")
	if err != nil {
		t.Errorf("Main help command failed: %v", err)
	}
	if !containsAll(output, []string{"OllamaMax", "proxy", "start"}) {
		t.Errorf("Main help output missing expected content")
	}

	// Test proxy help
	output, err = framework.RunCLICommand("proxy", "--help")
	if err != nil {
		t.Errorf("Proxy help command failed: %v", err)
	}
	if !containsAll(output, []string{"status", "instances", "metrics"}) {
		t.Errorf("Proxy help output missing expected content")
	}

	t.Log("✅ Basic CLI functionality working")
}

// testJSONValidation tests JSON output validation
func testJSONValidation(t *testing.T, framework *IntegrationTestFramework) {
	t.Log("📋 Testing JSON output validation...")

	commands := [][]string{
		{"proxy", "status", "--json", "--api-url", framework.APIBaseURL},
		{"proxy", "instances", "--json", "--api-url", framework.APIBaseURL},
		{"proxy", "metrics", "--json", "--api-url", framework.APIBaseURL},
	}

	for _, cmd := range commands {
		output, err := framework.RunCLICommand(cmd...)
		if err != nil {
			t.Errorf("JSON command failed: %v, command: %v", err, cmd)
			continue
		}

		data := framework.ValidateJSONOutput(t, output)
		if len(data) == 0 {
			t.Errorf("JSON output is empty for command: %v", cmd)
		}
	}

	t.Log("✅ JSON output validation passed")
}

// testErrorScenarios tests error handling scenarios
func testErrorScenarios(t *testing.T, framework *IntegrationTestFramework) {
	t.Log("🚨 Testing error handling scenarios...")

	errorTests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "Invalid Command",
			args:     []string{"invalid-command"},
			contains: []string{"Error", "unknown"},
		},
		{
			name:     "Invalid Proxy Subcommand",
			args:     []string{"proxy", "invalid-subcommand"},
			contains: []string{"Error", "unknown"},
		},
		{
			name:     "Invalid API URL",
			args:     []string{"proxy", "status", "--api-url", "http://invalid:9999"},
			contains: []string{"Error"},
		},
		{
			name:     "Missing Required Flag",
			args:     []string{"proxy", "status", "--api-url"},
			contains: []string{"Error"},
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			output, err := framework.RunCLICommand(test.args...)
			
			// Should return error
			if err == nil {
				t.Errorf("Expected error but command succeeded")
			}

			// Should contain error message
			if !containsAny(output, test.contains) {
				t.Errorf("Error output should contain one of %v\nOutput: %s", test.contains, output)
			}
		})
	}

	t.Log("✅ Error handling scenarios passed")
}

// testPerformance tests performance characteristics
func testPerformance(t *testing.T, framework *IntegrationTestFramework) {
	t.Log("🏃 Testing performance characteristics...")

	// Test command execution speed
	commands := [][]string{
		{"proxy", "--help"},
		{"proxy", "status", "--help"},
		{"proxy", "instances", "--help"},
		{"proxy", "metrics", "--help"},
	}

	for _, cmd := range commands {
		framework.PerformanceTest(t, cmd, 10, 5*time.Second)
	}

	// Test concurrent execution
	framework.StressTest(t, []string{"proxy", "--help"}, 5, 10*time.Second)

	t.Log("✅ Performance tests passed")
}

// TestUserWorkflows tests complete user workflows
func TestUserWorkflows(t *testing.T) {
	framework := NewIntegrationTestFramework("./ollama-distributed", "http://localhost:8080")
	defer framework.Teardown(t)

	if err := framework.Setup(t); err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}

	t.Log("👤 Testing user workflows...")

	// Workflow 1: New user discovers proxy commands
	t.Run("DiscoveryWorkflow", func(t *testing.T) {
		// User runs main help
		output, err := framework.RunCLICommand("--help")
		if err != nil || !containsAll(output, []string{"proxy"}) {
			t.Errorf("User cannot discover proxy commands")
		}

		// User explores proxy help
		output, err = framework.RunCLICommand("proxy", "--help")
		if err != nil || !containsAll(output, []string{"status", "instances", "metrics"}) {
			t.Errorf("User cannot discover proxy subcommands")
		}
	})

	// Workflow 2: User monitors cluster
	t.Run("MonitoringWorkflow", func(t *testing.T) {
		// Check proxy status
		output, err := framework.RunCLICommand("proxy", "status", "--api-url", framework.APIBaseURL)
		if err != nil {
			t.Errorf("User cannot check proxy status: %v", err)
		}

		// List instances
		output, err = framework.RunCLICommand("proxy", "instances", "--api-url", framework.APIBaseURL)
		if err != nil {
			t.Errorf("User cannot list instances: %v", err)
		}

		// View metrics
		output, err = framework.RunCLICommand("proxy", "metrics", "--api-url", framework.APIBaseURL)
		if err != nil {
			t.Errorf("User cannot view metrics: %v", err)
		}
	})

	// Workflow 3: User automates with JSON
	t.Run("AutomationWorkflow", func(t *testing.T) {
		// Get JSON status
		output, err := framework.RunCLICommand("proxy", "status", "--json", "--api-url", framework.APIBaseURL)
		if err != nil {
			t.Errorf("User cannot get JSON status: %v", err)
		}

		// Validate JSON for automation
		framework.ValidateJSONOutput(t, output)
	})

	t.Log("✅ User workflows tested successfully")
}

// Helper functions
func containsAll(text string, substrings []string) bool {
	for _, substring := range substrings {
		if !contains(text, substring) {
			return false
		}
	}
	return true
}

func containsAny(text string, substrings []string) bool {
	for _, substring := range substrings {
		if contains(text, substring) {
			return true
		}
	}
	return false
}

func contains(text, substring string) bool {
	return len(text) > 0 && len(substring) > 0 && 
		   (text == substring || len(text) > len(substring) && 
		    (text[:len(substring)] == substring || 
		     text[len(text)-len(substring):] == substring ||
		     findInString(text, substring)))
}

func findInString(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}
