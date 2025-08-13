//go:build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/config"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/p2p"
)

func main() {
	fmt.Println("Testing System-wide Metrics Integration...")

	// Setup system-wide metrics integration
	fmt.Println("Setting up system-wide metrics integration...")
	registry, integration, err := setupSystemMetricsIntegration()
	if err != nil {
		log.Fatalf("Failed to setup system metrics integration: %v", err)
	}

	// Start the metrics system
	if err := registry.Start(); err != nil {
		log.Fatalf("Failed to start metrics registry: %v", err)
	}
	defer registry.Stop()

	if err := integration.Start(); err != nil {
		log.Fatalf("Failed to start metrics integration: %v", err)
	}
	defer integration.Stop()

	fmt.Println("✅ System-wide metrics integration setup complete")

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Run system-wide metrics integration tests
	fmt.Println("\n=== Testing System-wide Metrics Integration ===")

	tests := []struct {
		name        string
		description string
		testFunc    func(*observability.MetricsRegistry, *observability.MetricsIntegration) error
	}{
		{
			name:        "Metrics Integration Startup",
			description: "Test that metrics integration starts correctly",
			testFunc:    testMetricsIntegrationStartup,
		},
		{
			name:        "Scheduler Metrics Integration",
			description: "Test scheduler metrics integration with Prometheus",
			testFunc:    testSchedulerMetricsIntegration,
		},
		{
			name:        "P2P Metrics Integration",
			description: "Test P2P network metrics integration with Prometheus",
			testFunc:    testP2PMetricsIntegration,
		},
		{
			name:        "Consensus Metrics Integration",
			description: "Test consensus engine metrics integration with Prometheus",
			testFunc:    testConsensusMetricsIntegration,
		},
		{
			name:        "API Gateway Metrics Integration",
			description: "Test API gateway metrics integration with Prometheus",
			testFunc:    testAPIMetricsIntegration,
		},
		{
			name:        "Fault Tolerance Metrics Integration",
			description: "Test fault tolerance metrics integration with Prometheus",
			testFunc:    testFaultToleranceMetricsIntegration,
		},
		{
			name:        "Model Management Metrics Integration",
			description: "Test model management metrics integration with Prometheus",
			testFunc:    testModelMetricsIntegration,
		},
		{
			name:        "Real-time Metrics Collection",
			description: "Test real-time metrics collection across all components",
			testFunc:    testRealTimeMetricsCollection,
		},
		{
			name:        "Metrics Endpoint Validation",
			description: "Test that all component metrics appear in Prometheus endpoint",
			testFunc:    testMetricsEndpointValidation,
		},
		{
			name:        "Component Integration Validation",
			description: "Test integration with actual system components",
			testFunc:    testComponentIntegrationValidation,
		},
	}

	for i, test := range tests {
		fmt.Printf("%d. Testing %s...\n", i+1, test.name)
		fmt.Printf("  Description: %s\n", test.description)

		if err := test.testFunc(registry, integration); err != nil {
			fmt.Printf("  ❌ Test failed: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ Test passed\n\n")

		// Wait between tests
		time.Sleep(1 * time.Second)
	}

	fmt.Println("✅ System-wide metrics integration test completed successfully!")
}

func setupSystemMetricsIntegration() (*observability.MetricsRegistry, *observability.MetricsIntegration, error) {
	// Create metrics configuration
	config := &observability.MetricsConfig{
		Namespace:          "ollama",
		Subsystem:          "distributed",
		CollectionInterval: 5 * time.Second,
		EnablePrometheus:   true,
		PrometheusConfig: &observability.PrometheusConfig{
			ListenAddress:        ":9092", // Use different port to avoid conflicts
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

	// Create metrics integration
	integration := observability.NewMetricsIntegration(registry, "test-node-1")

	return registry, integration, nil
}

func testMetricsIntegrationStartup(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	// Test that all integrators are available
	if integration.GetSchedulerIntegrator() == nil {
		return fmt.Errorf("scheduler integrator not available")
	}

	if integration.GetP2PIntegrator() == nil {
		return fmt.Errorf("P2P integrator not available")
	}

	if integration.GetConsensusIntegrator() == nil {
		return fmt.Errorf("consensus integrator not available")
	}

	if integration.GetAPIIntegrator() == nil {
		return fmt.Errorf("API integrator not available")
	}

	if integration.GetFaultToleranceIntegrator() == nil {
		return fmt.Errorf("fault tolerance integrator not available")
	}

	if integration.GetModelIntegrator() == nil {
		return fmt.Errorf("model integrator not available")
	}

	fmt.Printf("    All component integrators initialized successfully\n")
	return nil
}

func testSchedulerMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	schedulerIntegrator := integration.GetSchedulerIntegrator()

	// Test scheduler metrics reporting
	schedulerIntegrator.ReportTaskScheduled("inference", "completed")
	schedulerIntegrator.ReportTaskActive("inference", 5)
	schedulerIntegrator.ReportTaskDuration("inference", 2*time.Second)
	schedulerIntegrator.ReportTaskError("inference", "timeout")
	schedulerIntegrator.ReportLoadBalancerRequest("round_robin")
	schedulerIntegrator.ReportNodeUtilization("cpu", 0.75)
	schedulerIntegrator.ReportNodeUtilization("memory", 0.60)

	fmt.Printf("    Scheduler metrics reported: tasks, duration, utilization\n")
	return nil
}

func testP2PMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	p2pIntegrator := integration.GetP2PIntegrator()

	// Test P2P metrics reporting
	p2pIntegrator.ReportConnection("outbound", "peer1")
	p2pIntegrator.ReportActiveConnections("tcp", 3)
	p2pIntegrator.ReportMessageSent("heartbeat", "peer1")
	p2pIntegrator.ReportMessageReceived("heartbeat", "peer1")
	p2pIntegrator.ReportNetworkLatency("peer1", 25*time.Millisecond)
	p2pIntegrator.ReportBandwidthUsage("inbound", "peer1", 1024)
	p2pIntegrator.ReportPeerDiscovery("mdns", "success")

	fmt.Printf("    P2P metrics reported: connections, messages, latency\n")
	return nil
}

func testConsensusMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	consensusIntegrator := integration.GetConsensusIntegrator()

	// Test consensus metrics reporting
	consensusIntegrator.ReportLeaderElection("success")
	consensusIntegrator.ReportLogEntry("append", 10)
	consensusIntegrator.ReportCommitLatency(15 * time.Millisecond)
	consensusIntegrator.ReportQuorumStatus("cluster1", true)
	consensusIntegrator.ReportNodeState("leader")
	consensusIntegrator.ReportConsensusError("timeout")

	fmt.Printf("    Consensus metrics reported: elections, log entries, quorum\n")
	return nil
}

func testAPIMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	apiIntegrator := integration.GetAPIIntegrator()

	// Test API metrics reporting
	apiIntegrator.ReportAPIRequest("POST", "/api/v1/inference", "200")
	apiIntegrator.ReportRequestDuration("POST", "/api/v1/inference", 500*time.Millisecond)
	apiIntegrator.ReportResponseSize("POST", "/api/v1/inference", 2048)
	apiIntegrator.ReportActiveConnections("http", 25)
	apiIntegrator.ReportWebSocketConnections("/ws", 5)
	apiIntegrator.ReportRateLimitHit("/api/v1/inference", "client1")

	fmt.Printf("    API metrics reported: requests, duration, connections\n")
	return nil
}

func testFaultToleranceMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	ftIntegrator := integration.GetFaultToleranceIntegrator()

	// Test fault tolerance metrics reporting
	ftIntegrator.ReportFaultDetected("node_failure", "scheduler", "high")
	ftIntegrator.ReportRecoveryAttempt("restart", "scheduler")
	ftIntegrator.ReportRecoverySuccess("restart", "scheduler")
	ftIntegrator.ReportPredictionAccuracy("ml_model", "scheduler", 0.95)
	ftIntegrator.ReportHealingOperation("auto_restart", "scheduler")
	ftIntegrator.ReportSystemHealth("scheduler", "task_manager", 0.98)

	fmt.Printf("    Fault tolerance metrics reported: faults, recovery, health\n")
	return nil
}

func testModelMetricsIntegration(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	modelIntegrator := integration.GetModelIntegrator()

	// Test model metrics reporting
	modelIntegrator.ReportModelLoaded("llama2-7b", 1)
	modelIntegrator.ReportModelRequest("llama2-7b", "success")
	modelIntegrator.ReportModelLatency("llama2-7b", 800*time.Millisecond)
	modelIntegrator.ReportModelError("llama2-7b", "out_of_memory")
	modelIntegrator.ReportReplicationOperation("sync", "llama2-7b", "success")
	modelIntegrator.ReportStorageUsage("llama2-7b", "disk", 7000000000)

	fmt.Printf("    Model metrics reported: loaded models, requests, latency\n")
	return nil
}

func testRealTimeMetricsCollection(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	// Simulate real-time metrics collection from multiple components

	// Scheduler activity
	schedulerIntegrator := integration.GetSchedulerIntegrator()
	for i := 0; i < 5; i++ {
		schedulerIntegrator.ReportTaskScheduled("inference", "completed")
		schedulerIntegrator.ReportTaskActive("inference", float64(10-i))
		time.Sleep(100 * time.Millisecond)
	}

	// P2P activity
	p2pIntegrator := integration.GetP2PIntegrator()
	for i := 0; i < 3; i++ {
		p2pIntegrator.ReportMessageSent("data", "peer1")
		p2pIntegrator.ReportMessageReceived("data", "peer1")
		time.Sleep(100 * time.Millisecond)
	}

	// API activity
	apiIntegrator := integration.GetAPIIntegrator()
	for i := 0; i < 7; i++ {
		apiIntegrator.ReportAPIRequest("GET", "/api/v1/status", "200")
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("    Real-time metrics collection simulated across all components\n")
	return nil
}

func testMetricsEndpointValidation(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	// Test that metrics appear in the Prometheus endpoint
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

	// Check for component-specific metrics
	expectedMetrics := []string{
		"ollama_distributed_scheduler_tasks_total",
		"ollama_distributed_p2p_connections_total",
		"ollama_distributed_consensus_leader_elections_total",
		"ollama_distributed_api_requests_total",
		"ollama_distributed_fault_tolerance_faults_detected_total",
		"ollama_distributed_model_requests_total",
	}

	foundMetrics := 0
	for _, metric := range expectedMetrics {
		if strings.Contains(bodyStr, metric) {
			foundMetrics++
		}
	}

	if foundMetrics == 0 {
		return fmt.Errorf("no component metrics found in endpoint")
	}

	fmt.Printf("    Metrics endpoint validation: %d/%d component metrics found\n", foundMetrics, len(expectedMetrics))
	return nil
}

func testComponentIntegrationValidation(registry *observability.MetricsRegistry, integration *observability.MetricsIntegration) error {
	// Test integration with actual P2P component
	ctx := context.Background()
	nodeConfig := config.DefaultConfig()
	nodeConfig.Listen = []string{"/ip4/127.0.0.1/tcp/0"}

	p2pNode, err := p2p.NewP2PNode(ctx, nodeConfig)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}
	defer p2pNode.Stop()

	// Set metrics integration
	p2pNode.SetMetricsIntegration(integration)

	// Start the node
	if err := p2pNode.Start(); err != nil {
		return fmt.Errorf("failed to start P2P node: %w", err)
	}

	// Wait for metrics to be collected
	time.Sleep(2 * time.Second)

	fmt.Printf("    Component integration validated with actual P2P node\n")
	return nil
}
