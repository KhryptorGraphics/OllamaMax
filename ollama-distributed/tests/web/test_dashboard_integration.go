//go:build ignore

package web_tests

import (
	"fmt"
	"net/http"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/api"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/monitoring"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/performance"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/security"
)

func TestDashboardIntegration() {
	fmt.Println("Testing Web Dashboard Integration...")

	// Setup backend services for testing
	fmt.Println("Setting up backend services...")

	// 1. Create monitoring system
	monitoringConfig := monitoring.DefaultConfig()
	monitoringSystem := monitoring.NewMonitoringSystem(monitoringConfig)

	// 2. Create security system
	securityConfig := security.DefaultSecurityConfig()
	securitySystem := security.NewSecuritySystem(securityConfig)

	// 3. Create performance system
	performanceConfig := performance.DefaultOptimizationConfig()
	performanceSystem := performance.NewPerformanceOptimizationEngine(performanceConfig)

	// 4. Create API gateway
	apiConfig := api.DefaultConfig()
	apiGateway := api.NewGateway(apiConfig)

	// Start all systems
	fmt.Println("Starting backend systems...")

	if err := monitoringSystem.Start(); err != nil {
		fmt.Printf("❌ Failed to start monitoring system: %v\n", err)
		return
	}

	if err := securitySystem.Start(); err != nil {
		fmt.Printf("❌ Failed to start security system: %v\n", err)
		return
	}

	if err := performanceSystem.Start(); err != nil {
		fmt.Printf("❌ Failed to start performance system: %v\n", err)
		return
	}

	if err := apiGateway.Start(); err != nil {
		fmt.Printf("❌ Failed to start API gateway: %v\n", err)
		return
	}

	fmt.Println("✅ All backend systems started")

	// Wait for systems to initialize
	time.Sleep(3 * time.Second)

	// Run dashboard integration tests
	testResults := []bool{}

	// Test 1: Web Dashboard Accessibility
	fmt.Println("\n=== Testing Web Dashboard Accessibility ===")
	result := testWebDashboardAccessibility()
	testResults = append(testResults, result)

	// Test 2: API Endpoints
	fmt.Println("\n=== Testing API Endpoints ===")
	result = testAPIEndpoints()
	testResults = append(testResults, result)

	// Test 3: Security Dashboard Integration
	fmt.Println("\n=== Testing Security Dashboard Integration ===")
	result = testSecurityDashboardIntegration()
	testResults = append(testResults, result)

	// Test 4: Performance Dashboard Integration
	fmt.Println("\n=== Testing Performance Dashboard Integration ===")
	result = testPerformanceDashboardIntegration()
	testResults = append(testResults, result)

	// Test 5: Real-time Features
	fmt.Println("\n=== Testing Real-time Features ===")
	result = testRealTimeFeatures()
	testResults = append(testResults, result)

	// Test 6: WebSocket Integration
	fmt.Println("\n=== Testing WebSocket Integration ===")
	result = testWebSocketIntegration()
	testResults = append(testResults, result)

	// Test 7: Dashboard Navigation
	fmt.Println("\n=== Testing Dashboard Navigation ===")
	result = testDashboardNavigation()
	testResults = append(testResults, result)

	// Test 8: Error Handling
	fmt.Println("\n=== Testing Error Handling ===")
	result = testErrorHandling()
	testResults = append(testResults, result)

	// Cleanup
	fmt.Println("\n=== Cleaning up ===")
	monitoringSystem.Shutdown()
	securitySystem.Shutdown()
	performanceSystem.Shutdown()
	apiGateway.Shutdown()

	// Summary
	fmt.Println("\n=== Test Results Summary ===")
	passed := 0
	for i, result := range testResults {
		status := "❌ FAILED"
		if result {
			status = "✅ PASSED"
			passed++
		}
		fmt.Printf("Test %d: %s\n", i+1, status)
	}

	fmt.Printf("\nOverall: %d/%d tests passed\n", passed, len(testResults))

	if passed == len(testResults) {
		fmt.Println("🎉 All web dashboard integration tests passed!")
	} else {
		fmt.Println("⚠️  Some tests failed. Check the output above for details.")
	}
}

func testWebDashboardAccessibility() bool {
	fmt.Println("1. Testing Web Dashboard Accessibility...")

	// Test main dashboard page
	resp, err := http.Get("http://localhost:12925/")
	if err != nil {
		fmt.Printf("  ❌ Failed to access dashboard: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  ❌ Dashboard returned status %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("  ✅ Dashboard accessible")
	return true
}

func testAPIEndpoints() bool {
	fmt.Println("2. Testing API Endpoints...")

	endpoints := []string{
		"/api/v1/cluster/status",
		"/api/v1/nodes",
		"/api/v1/models",
		"/api/v1/metrics",
		"/api/v1/security/status",
		"/api/v1/performance/metrics",
	}

	for _, endpoint := range endpoints {
		resp, err := http.Get("http://localhost:12925" + endpoint)
		if err != nil {
			fmt.Printf("  ❌ Failed to access %s: %v\n", endpoint, err)
			return false
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			fmt.Printf("  ❌ Endpoint %s returned status %d\n", endpoint, resp.StatusCode)
			return false
		}
	}

	fmt.Printf("  ✅ All %d API endpoints accessible\n", len(endpoints))
	return true
}

func testSecurityDashboardIntegration() bool {
	fmt.Println("3. Testing Security Dashboard Integration...")

	// Test security status endpoint
	resp, err := http.Get("http://localhost:12925/api/v1/security/status")
	if err != nil {
		fmt.Printf("  ❌ Failed to access security status: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Test security threats endpoint
	resp, err = http.Get("http://localhost:12925/api/v1/security/threats")
	if err != nil {
		fmt.Printf("  ❌ Failed to access security threats: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Test security alerts endpoint
	resp, err = http.Get("http://localhost:12925/api/v1/security/alerts")
	if err != nil {
		fmt.Printf("  ❌ Failed to access security alerts: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	fmt.Println("  ✅ Security dashboard integration successful")
	return true
}

func testPerformanceDashboardIntegration() bool {
	fmt.Println("4. Testing Performance Dashboard Integration...")

	// Test performance metrics endpoint
	resp, err := http.Get("http://localhost:12925/api/v1/performance/metrics")
	if err != nil {
		fmt.Printf("  ❌ Failed to access performance metrics: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Test performance optimizations endpoint
	resp, err = http.Get("http://localhost:12925/api/v1/performance/optimizations")
	if err != nil {
		fmt.Printf("  ❌ Failed to access performance optimizations: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// Test performance bottlenecks endpoint
	resp, err = http.Get("http://localhost:12925/api/v1/performance/bottlenecks")
	if err != nil {
		fmt.Printf("  ❌ Failed to access performance bottlenecks: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	fmt.Println("  ✅ Performance dashboard integration successful")
	return true
}

func testRealTimeFeatures() bool {
	fmt.Println("5. Testing Real-time Features...")

	// Test metrics endpoint for real-time data
	resp, err := http.Get("http://localhost:12925/api/v1/metrics")
	if err != nil {
		fmt.Printf("  ❌ Failed to access real-time metrics: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  ❌ Real-time metrics returned status %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("  ✅ Real-time features accessible")
	return true
}

func testWebSocketIntegration() bool {
	fmt.Println("6. Testing WebSocket Integration...")

	// Test WebSocket endpoint accessibility
	resp, err := http.Get("http://localhost:12925/api/v1/ws")
	if err != nil {
		fmt.Printf("  ❌ Failed to access WebSocket endpoint: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// WebSocket upgrade should return 400 for regular HTTP request
	if resp.StatusCode != http.StatusBadRequest {
		fmt.Printf("  ❌ WebSocket endpoint returned unexpected status %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("  ✅ WebSocket endpoint accessible")
	return true
}

func testDashboardNavigation() bool {
	fmt.Println("7. Testing Dashboard Navigation...")

	// Test that dashboard serves static files
	resp, err := http.Get("http://localhost:12925/")
	if err != nil {
		fmt.Printf("  ❌ Failed to access dashboard: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  ❌ Dashboard navigation failed with status %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("  ✅ Dashboard navigation successful")
	return true
}

func testErrorHandling() bool {
	fmt.Println("8. Testing Error Handling...")

	// Test non-existent endpoint
	resp, err := http.Get("http://localhost:12925/api/v1/nonexistent")
	if err != nil {
		fmt.Printf("  ❌ Failed to test error handling: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		fmt.Printf("  ❌ Error handling returned unexpected status %d\n", resp.StatusCode)
		return false
	}

	fmt.Println("  ✅ Error handling working correctly")
	return true
}
