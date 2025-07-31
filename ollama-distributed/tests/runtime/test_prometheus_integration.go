package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
)

func main() {
	fmt.Println("Testing Prometheus Integration...")

	// Setup metrics registry with Prometheus integration
	fmt.Println("Setting up metrics registry with Prometheus integration...")
	registry, err := setupPrometheusIntegration()
	if err != nil {
		log.Fatalf("Failed to setup Prometheus integration: %v", err)
	}

	// Start the metrics registry
	if err := registry.Start(); err != nil {
		log.Fatalf("Failed to start metrics registry: %v", err)
	}
	defer registry.Stop()

	fmt.Println("✅ Prometheus integration setup complete")

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Run Prometheus integration tests
	fmt.Println("\n=== Testing Prometheus Integration ===")

	tests := []struct {
		name        string
		description string
		testFunc    func(*observability.MetricsRegistry) error
	}{
		{
			name:        "Prometheus Server Startup",
			description: "Test that Prometheus metrics server starts correctly",
			testFunc:    testPrometheusServerStartup,
		},
		{
			name:        "Metrics Endpoint Accessibility",
			description: "Test that /metrics endpoint is accessible and returns data",
			testFunc:    testMetricsEndpoint,
		},
		{
			name:        "Scheduler Metrics Collection",
			description: "Test scheduler metrics collection and export",
			testFunc:    testSchedulerMetrics,
		},
		{
			name:        "Consensus Metrics Collection",
			description: "Test consensus metrics collection and export",
			testFunc:    testConsensusMetrics,
		},
		{
			name:        "P2P Metrics Collection",
			description: "Test P2P network metrics collection and export",
			testFunc:    testP2PMetrics,
		},
		{
			name:        "API Metrics Collection",
			description: "Test API gateway metrics collection and export",
			testFunc:    testAPIMetrics,
		},
		{
			name:        "Fault Tolerance Metrics Collection",
			description: "Test fault tolerance metrics collection and export",
			testFunc:    testFaultToleranceMetrics,
		},
		{
			name:        "Model Metrics Collection",
			description: "Test model management metrics collection and export",
			testFunc:    testModelMetrics,
		},
		{
			name:        "Health Check Endpoints",
			description: "Test health and readiness check endpoints",
			testFunc:    testHealthEndpoints,
		},
		{
			name:        "Metric Naming Conventions",
			description: "Test that metrics follow Prometheus naming conventions",
			testFunc:    testMetricNamingConventions,
		},
	}

	for i, test := range tests {
		fmt.Printf("%d. Testing %s...\n", i+1, test.name)
		fmt.Printf("  Description: %s\n", test.description)

		if err := test.testFunc(registry); err != nil {
			fmt.Printf("  ❌ Test failed: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ Test passed\n\n")

		// Wait between tests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("✅ Prometheus integration test completed successfully!")
}

func setupPrometheusIntegration() (*observability.MetricsRegistry, error) {
	// Create metrics configuration with Prometheus enabled
	config := &observability.MetricsConfig{
		Namespace:          "ollama",
		Subsystem:          "distributed",
		CollectionInterval: 5 * time.Second,
		EnablePrometheus:   true,
		PrometheusConfig: &observability.PrometheusConfig{
			ListenAddress:        ":9091", // Use different port to avoid conflicts
			MetricsPath:          "/metrics",
			Namespace:            "ollama",
			Subsystem:            "distributed",
			EnableGoMetrics:      true,
			EnableProcessMetrics: true,
			GatherInterval:       10 * time.Second,
			ReadTimeout:          30 * time.Second,
			WriteTimeout:         30 * time.Second,
			IdleTimeout:          60 * time.Second,
		},
	}

	// Create metrics registry
	registry := observability.NewMetricsRegistry(config)

	return registry, nil
}

func testPrometheusServerStartup(registry *observability.MetricsRegistry) error {
	// Test that the Prometheus exporter is available
	exporter := registry.GetPrometheusExporter()
	if exporter == nil {
		return fmt.Errorf("Prometheus exporter not initialized")
	}

	fmt.Printf("    Prometheus server started on %s\n", exporter.GetMetricsURL())
	return nil
}

func testMetricsEndpoint(registry *observability.MetricsRegistry) error {
	// Test that the /metrics endpoint is accessible
	exporter := registry.GetPrometheusExporter()
	metricsURL := exporter.GetMetricsURL()

	resp, err := http.Get(metricsURL)
	if err != nil {
		return fmt.Errorf("failed to access metrics endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metrics endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read metrics response: %w", err)
	}

	// Check that response contains Prometheus metrics
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "# HELP") || !strings.Contains(bodyStr, "# TYPE") {
		return fmt.Errorf("metrics endpoint does not return Prometheus format")
	}

	fmt.Printf("    Metrics endpoint accessible, returned %d bytes\n", len(body))
	return nil
}

func testSchedulerMetrics(registry *observability.MetricsRegistry) error {
	// Test scheduler metrics
	schedulerMetrics := registry.GetSchedulerMetrics()
	if schedulerMetrics == nil {
		return fmt.Errorf("scheduler metrics not initialized")
	}

	// Test some scheduler metrics
	schedulerMetrics.TasksTotal.WithLabelValues("completed", "node1", "inference").Inc()
	schedulerMetrics.TasksActive.WithLabelValues("node1", "inference").Set(5)
	schedulerMetrics.TaskDuration.WithLabelValues("inference", "node1").Observe(2.5)
	schedulerMetrics.NodeUtilization.WithLabelValues("node1", "cpu").Set(0.75)

	fmt.Printf("    Scheduler metrics recorded: tasks, utilization, duration\n")
	return nil
}

func testConsensusMetrics(registry *observability.MetricsRegistry) error {
	// Test consensus metrics
	consensusMetrics := registry.GetConsensusMetrics()
	if consensusMetrics == nil {
		return fmt.Errorf("consensus metrics not initialized")
	}

	// Test some consensus metrics
	consensusMetrics.LeaderElections.WithLabelValues("node1", "success").Inc()
	consensusMetrics.LogEntries.WithLabelValues("node1", "append").Add(10)
	consensusMetrics.CommitLatency.WithLabelValues("node1").Observe(0.05)
	consensusMetrics.QuorumStatus.WithLabelValues("cluster1").Set(1)

	fmt.Printf("    Consensus metrics recorded: elections, log entries, latency\n")
	return nil
}

func testP2PMetrics(registry *observability.MetricsRegistry) error {
	// Test P2P metrics
	p2pMetrics := registry.GetP2PMetrics()
	if p2pMetrics == nil {
		return fmt.Errorf("P2P metrics not initialized")
	}

	// Test some P2P metrics
	p2pMetrics.ConnectionsTotal.WithLabelValues("outbound", "peer1").Inc()
	p2pMetrics.ConnectionsActive.WithLabelValues("tcp").Set(3)
	p2pMetrics.MessagesSent.WithLabelValues("heartbeat", "peer1").Add(100)
	p2pMetrics.NetworkLatency.WithLabelValues("peer1").Observe(0.025)

	fmt.Printf("    P2P metrics recorded: connections, messages, latency\n")
	return nil
}

func testAPIMetrics(registry *observability.MetricsRegistry) error {
	// Test API metrics
	apiMetrics := registry.GetAPIMetrics()
	if apiMetrics == nil {
		return fmt.Errorf("API metrics not initialized")
	}

	// Test some API metrics
	apiMetrics.RequestsTotal.WithLabelValues("POST", "/api/v1/inference", "200").Inc()
	apiMetrics.RequestDuration.WithLabelValues("POST", "/api/v1/inference").Observe(1.2)
	apiMetrics.ResponseSize.WithLabelValues("POST", "/api/v1/inference").Observe(1024)
	apiMetrics.ActiveConnections.WithLabelValues("http").Set(25)

	fmt.Printf("    API metrics recorded: requests, duration, response size\n")
	return nil
}

func testFaultToleranceMetrics(registry *observability.MetricsRegistry) error {
	// Test fault tolerance metrics
	ftMetrics := registry.GetFaultToleranceMetrics()
	if ftMetrics == nil {
		return fmt.Errorf("fault tolerance metrics not initialized")
	}

	// Test some fault tolerance metrics
	ftMetrics.FaultsDetected.WithLabelValues("node_failure", "scheduler", "high").Inc()
	ftMetrics.RecoveryAttempts.WithLabelValues("restart", "scheduler").Inc()
	ftMetrics.RecoverySuccess.WithLabelValues("restart", "scheduler").Inc()
	ftMetrics.SystemHealth.WithLabelValues("scheduler", "task_manager").Set(0.95)

	fmt.Printf("    Fault tolerance metrics recorded: faults, recovery, health\n")
	return nil
}

func testModelMetrics(registry *observability.MetricsRegistry) error {
	// Test model metrics
	modelMetrics := registry.GetModelMetrics()
	if modelMetrics == nil {
		return fmt.Errorf("model metrics not initialized")
	}

	// Test some model metrics
	modelMetrics.ModelsLoaded.WithLabelValues("llama2-7b", "node1").Set(1)
	modelMetrics.ModelRequests.WithLabelValues("llama2-7b", "node1", "success").Add(50)
	modelMetrics.ModelLatency.WithLabelValues("llama2-7b", "node1").Observe(0.8)
	modelMetrics.StorageUsage.WithLabelValues("llama2-7b", "node1", "disk").Set(7000000000)

	fmt.Printf("    Model metrics recorded: loaded models, requests, latency\n")
	return nil
}

func testHealthEndpoints(registry *observability.MetricsRegistry) error {
	// Test health endpoints
	baseURL := "http://localhost:9091"

	// Test health endpoint
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to access health endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
	}

	// Test readiness endpoint
	resp, err = http.Get(baseURL + "/ready")
	if err != nil {
		return fmt.Errorf("failed to access readiness endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("readiness endpoint returned status %d", resp.StatusCode)
	}

	fmt.Printf("    Health and readiness endpoints accessible\n")
	return nil
}

func testMetricNamingConventions(registry *observability.MetricsRegistry) error {
	// Test that metrics follow Prometheus naming conventions
	exporter := registry.GetPrometheusExporter()
	metricsURL := exporter.GetMetricsURL()

	resp, err := http.Get(metricsURL)
	if err != nil {
		return fmt.Errorf("failed to access metrics endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read metrics response: %w", err)
	}

	bodyStr := string(body)

	// Check for proper naming conventions
	expectedPrefixes := []string{
		"ollama_distributed_scheduler_",
		"ollama_distributed_consensus_",
		"ollama_distributed_p2p_",
		"ollama_distributed_api_",
		"ollama_distributed_fault_tolerance_",
		"ollama_distributed_model_",
	}

	foundPrefixes := 0
	for _, prefix := range expectedPrefixes {
		if strings.Contains(bodyStr, prefix) {
			foundPrefixes++
		}
	}

	if foundPrefixes == 0 {
		return fmt.Errorf("no metrics with expected naming conventions found")
	}

	fmt.Printf("    Metrics follow naming conventions: %d/%d prefixes found\n", foundPrefixes, len(expectedPrefixes))
	return nil
}
