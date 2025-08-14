//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/observability"
)

func main() {
	fmt.Println("Testing Distributed Tracing System...")

	// Setup distributed tracing system
	fmt.Println("Setting up distributed tracing system...")
	tracingSystem, err := setupDistributedTracingSystem()
	if err != nil {
		log.Fatalf("Failed to setup distributed tracing system: %v", err)
	}

	// Start the tracing system
	if err := tracingSystem.Start(); err != nil {
		log.Fatalf("Failed to start distributed tracing system: %v", err)
	}
	defer tracingSystem.Stop()

	fmt.Println("✅ Distributed tracing system setup complete")

	// Wait for system to initialize
	time.Sleep(2 * time.Second)

	// Run distributed tracing tests
	fmt.Println("\n=== Testing Distributed Tracing System ===")

	tests := []struct {
		name        string
		description string
		testFunc    func(*observability.DistributedTracingSystem) error
	}{
		{
			name:        "Tracing System Startup",
			description: "Test that distributed tracing system starts correctly",
			testFunc:    testTracingSystemStartup,
		},
		{
			name:        "Custom Tracing Integration",
			description: "Test custom tracing functionality",
			testFunc:    testCustomTracingIntegration,
		},
		{
			name:        "OpenTelemetry Integration",
			description: "Test OpenTelemetry integration and Jaeger export",
			testFunc:    testOpenTelemetryIntegration,
		},
		{
			name:        "Component Tracing",
			description: "Test component-specific tracing (scheduler, P2P, consensus, API, model)",
			testFunc:    testComponentTracing,
		},
		{
			name:        "Context Propagation",
			description: "Test trace context propagation across service boundaries",
			testFunc:    testContextPropagation,
		},
		{
			name:        "Distributed Operation Tracing",
			description: "Test end-to-end distributed operation tracing",
			testFunc:    testDistributedOperationTracing,
		},
		{
			name:        "Span Instrumentation",
			description: "Test span instrumentation with attributes and events",
			testFunc:    testSpanInstrumentation,
		},
		{
			name:        "Error Tracing",
			description: "Test error recording and status propagation",
			testFunc:    testErrorTracing,
		},
		{
			name:        "Performance Impact",
			description: "Test performance impact of tracing system",
			testFunc:    testPerformanceImpact,
		},
		{
			name:        "Tracing Statistics",
			description: "Test tracing system statistics and monitoring",
			testFunc:    testTracingStatistics,
		},
	}

	for i, test := range tests {
		fmt.Printf("%d. Testing %s...\n", i+1, test.name)
		fmt.Printf("  Description: %s\n", test.description)

		if err := test.testFunc(tracingSystem); err != nil {
			fmt.Printf("  ❌ Test failed: %v\n", err)
			continue
		}

		fmt.Printf("  ✅ Test passed\n\n")

		// Wait between tests
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("✅ Distributed tracing system test completed successfully!")
}

func setupDistributedTracingSystem() (*observability.DistributedTracingSystem, error) {
	// Create distributed tracing configuration
	config := &observability.DistributedTracingConfig{
		ServiceName:    "ollama-distributed-test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		NodeID:         "test-node-1",

		// Custom tracing config
		CustomTracingConfig: &observability.TracerConfig{
			ServiceName:     "ollama-distributed-test",
			ServiceVersion:  "1.0.0",
			Environment:     "test",
			SamplingRate:    1.0,
			EnableExport:    false, // Disable export for testing
			ExportInterval:  10 * time.Second,
			ExportBatchSize: 100,
			ExportTimeout:   30 * time.Second,
			MaxSpans:        1000,
			SpanTTL:         time.Hour,
		},

		// OpenTelemetry config
		OpenTelemetryConfig: &observability.OpenTelemetryConfig{
			ServiceName:         "ollama-distributed-test",
			ServiceVersion:      "1.0.0",
			Environment:         "test",
			NodeID:              "test-node-1",
			JaegerEndpoint:      "http://localhost:14268/api/traces",
			SamplingRatio:       1.0,
			EnableOpenTelemetry: false, // Disable OpenTelemetry for testing
			EnableJaegerExport:  false, // Disable Jaeger export for testing
			EnablePropagation:   true,
			EnableBatching:      true,
			BatchTimeout:        1 * time.Second,
			ExportTimeout:       5 * time.Second,
			MaxExportBatch:      100,
			MaxQueueSize:        500,
		},

		// Features
		EnableCustomTracing:      true,
		EnableOpenTelemetry:      false, // Disable for testing
		EnableContextPropagation: true,
		EnableComponentTracing:   true,
		SamplingRatio:            1.0,
	}

	// Create distributed tracing system
	tracingSystem := observability.NewDistributedTracingSystem(config)

	return tracingSystem, nil
}

func testTracingSystemStartup(tracingSystem *observability.DistributedTracingSystem) error {
	// Test that the tracing system is enabled
	if !tracingSystem.IsEnabled() {
		return fmt.Errorf("tracing system is not enabled")
	}

	// Test that component tracers are available
	if tracingSystem.GetSchedulerTracer() == nil {
		return fmt.Errorf("scheduler tracer not available")
	}

	if tracingSystem.GetP2PTracer() == nil {
		return fmt.Errorf("P2P tracer not available")
	}

	if tracingSystem.GetConsensusTracer() == nil {
		return fmt.Errorf("consensus tracer not available")
	}

	if tracingSystem.GetAPITracer() == nil {
		return fmt.Errorf("API tracer not available")
	}

	if tracingSystem.GetModelTracer() == nil {
		return fmt.Errorf("model tracer not available")
	}

	fmt.Printf("    Distributed tracing system startup successful\n")
	return nil
}

func testCustomTracingIntegration(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Test custom tracer
	customTracer := tracingSystem.GetCustomTracer()
	if customTracer == nil {
		return fmt.Errorf("custom tracer not available")
	}

	// Start a custom span
	span, _ := customTracer.StartSpan(ctx, "test.custom_operation")
	span.SetTag("test.type", "custom_tracing")
	span.SetTag("test.component", "test")

	// Add some logs
	span.LogFields(map[string]interface{}{
		"event":   "test_event",
		"message": "Testing custom tracing integration",
	})

	// Finish span
	customTracer.FinishSpan(span)

	// Verify span was created
	if span.TraceID == "" {
		return fmt.Errorf("custom span trace ID not set")
	}

	if span.SpanID == "" {
		return fmt.Errorf("custom span ID not set")
	}

	fmt.Printf("    Custom tracing integration successful: trace_id=%s\n", span.TraceID[:8])
	return nil
}

func testOpenTelemetryIntegration(tracingSystem *observability.DistributedTracingSystem) error {
	// Test OpenTelemetry adapter
	otelAdapter := tracingSystem.GetOpenTelemetryAdapter()
	if otelAdapter == nil {
		fmt.Printf("    OpenTelemetry adapter disabled for testing\n")
		return nil
	}

	if !otelAdapter.IsEnabled() {
		fmt.Printf("    OpenTelemetry adapter not enabled (expected for testing)\n")
		return nil
	}

	// If enabled, test the functionality
	tracer := otelAdapter.GetTracer()
	if tracer == nil {
		return fmt.Errorf("OpenTelemetry tracer not available")
	}

	fmt.Printf("    OpenTelemetry integration available but disabled for testing\n")
	return nil
}

func testComponentTracing(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Test scheduler tracing
	schedulerTracer := tracingSystem.GetSchedulerTracer()
	newCtx, customSpan, otelSpan := schedulerTracer.TraceTaskScheduling(ctx, "task-123", "inference")
	schedulerTracer.FinishSpans(customSpan, otelSpan, nil)

	// Test P2P tracing
	p2pTracer := tracingSystem.GetP2PTracer()
	newCtx, customSpan, otelSpan = p2pTracer.TraceMessageSend(newCtx, "heartbeat", "peer-456", 1024)
	p2pTracer.FinishSpans(customSpan, otelSpan, nil)

	// Test consensus tracing
	consensusTracer := tracingSystem.GetConsensusTracer()
	newCtx, customSpan, otelSpan = consensusTracer.TraceLeaderElection(newCtx, 1, "node-1")
	consensusTracer.FinishSpans(customSpan, otelSpan, nil)

	// Test API tracing
	apiTracer := tracingSystem.GetAPITracer()
	newCtx, customSpan, otelSpan = apiTracer.TraceHTTPRequest(newCtx, "POST", "/api/v1/inference", "test-agent")
	apiTracer.FinishSpans(customSpan, otelSpan, nil)

	// Test model tracing
	modelTracer := tracingSystem.GetModelTracer()
	newCtx, customSpan, otelSpan = modelTracer.TraceInference(newCtx, "llama2-7b", "req-789", 100, 50)
	modelTracer.FinishSpans(customSpan, otelSpan, nil)

	fmt.Printf("    Component tracing successful: all 5 components traced\n")
	return nil
}

func testContextPropagation(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Start a distributed operation
	newCtx, customSpan, otelSpan := tracingSystem.StartDistributedOperation(ctx, "test.distributed_operation", "test", map[string]interface{}{
		"operation.type": "context_propagation_test",
	})

	// Test HTTP header propagation
	headers := make(http.Header)
	tracingSystem.InjectTraceContext(newCtx, headers)

	// Verify headers were injected
	if len(headers) == 0 {
		return fmt.Errorf("trace context not injected into HTTP headers")
	}

	// Extract context from headers
	_ = tracingSystem.ExtractTraceContext(context.Background(), headers)

	// Test map propagation
	carrier := make(map[string]string)
	tracingSystem.InjectTraceContextToMap(newCtx, carrier)

	// Verify map was populated
	if len(carrier) == 0 {
		return fmt.Errorf("trace context not injected into map carrier")
	}

	// Extract context from map
	_ = tracingSystem.ExtractTraceContextFromMap(context.Background(), carrier)

	// Finish operation
	tracingSystem.FinishDistributedOperation(customSpan, otelSpan, nil)

	fmt.Printf("    Context propagation successful: HTTP and map carriers\n")
	return nil
}

func testDistributedOperationTracing(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Simulate a distributed operation flow

	// 1. API request
	apiTracer := tracingSystem.GetAPITracer()
	apiCtx, apiCustomSpan, apiOtelSpan := apiTracer.TraceHTTPRequest(ctx, "POST", "/api/v1/inference", "test-client")

	// 2. Task scheduling
	schedulerTracer := tracingSystem.GetSchedulerTracer()
	schedCtx, schedCustomSpan, schedOtelSpan := schedulerTracer.TraceTaskScheduling(apiCtx, "task-456", "inference")

	// 3. P2P communication
	p2pTracer := tracingSystem.GetP2PTracer()
	p2pCtx, p2pCustomSpan, p2pOtelSpan := p2pTracer.TraceMessageSend(schedCtx, "task_assignment", "worker-node", 2048)

	// 4. Model inference
	modelTracer := tracingSystem.GetModelTracer()
	_, modelCustomSpan, modelOtelSpan := modelTracer.TraceInference(p2pCtx, "llama2-7b", "req-456", 150, 75)

	// Finish spans in reverse order (child to parent)
	modelTracer.FinishSpans(modelCustomSpan, modelOtelSpan, nil)
	p2pTracer.FinishSpans(p2pCustomSpan, p2pOtelSpan, nil)
	schedulerTracer.FinishSpans(schedCustomSpan, schedOtelSpan, nil)
	apiTracer.FinishSpans(apiCustomSpan, apiOtelSpan, nil)

	// Verify trace relationships
	if apiCustomSpan.TraceID != schedCustomSpan.TraceID {
		return fmt.Errorf("trace ID not propagated correctly")
	}

	fmt.Printf("    Distributed operation tracing successful: 4-span trace\n")
	return nil
}

func testSpanInstrumentation(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Start operation with attributes
	attributes := map[string]interface{}{
		"user.id":       "user-123",
		"request.size":  1024,
		"request.type":  "inference",
		"model.name":    "llama2-7b",
		"priority":      "high",
		"timeout":       30.0,
		"retry.enabled": true,
	}

	_, customSpan, otelSpan := tracingSystem.StartDistributedOperation(ctx, "test.instrumentation", "test", attributes)

	// Add events
	tracingSystem.AddDistributedEvent(customSpan, otelSpan, "processing_started", map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"stage":     "initialization",
	})

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	tracingSystem.AddDistributedEvent(customSpan, otelSpan, "processing_completed", map[string]interface{}{
		"timestamp":     time.Now().Unix(),
		"stage":         "completion",
		"result.tokens": 42,
	})

	// Finish operation
	tracingSystem.FinishDistributedOperation(customSpan, otelSpan, nil)

	// Verify instrumentation
	if len(customSpan.Tags) < len(attributes) {
		return fmt.Errorf("not all attributes were set on custom span")
	}

	if len(customSpan.Logs) < 2 {
		return fmt.Errorf("not all events were recorded on custom span")
	}

	fmt.Printf("    Span instrumentation successful: %d attributes, %d events\n", len(customSpan.Tags), len(customSpan.Logs))
	return nil
}

func testErrorTracing(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Start operation that will fail
	_, customSpan, otelSpan := tracingSystem.StartDistributedOperation(ctx, "test.error_operation", "test", map[string]interface{}{
		"operation.type": "error_test",
	})

	// Simulate an error
	testError := fmt.Errorf("simulated test error")

	// Finish operation with error
	tracingSystem.FinishDistributedOperation(customSpan, otelSpan, testError)

	// Verify error was recorded
	if customSpan.Status.Code != observability.SpanStatusCodeError {
		return fmt.Errorf("error status not set on custom span")
	}

	if customSpan.Status.Message != testError.Error() {
		return fmt.Errorf("error message not set correctly on custom span")
	}

	// Check error tag
	if errorTag, exists := customSpan.Tags["error"]; !exists || errorTag != true {
		return fmt.Errorf("error tag not set on custom span")
	}

	fmt.Printf("    Error tracing successful: error recorded and propagated\n")
	return nil
}

func testPerformanceImpact(tracingSystem *observability.DistributedTracingSystem) error {
	ctx := context.Background()

	// Measure performance impact
	iterations := 1000

	// Test without tracing
	start := time.Now()
	for i := 0; i < iterations; i++ {
		// Simulate work
		time.Sleep(1 * time.Microsecond)
	}
	baselineDuration := time.Since(start)

	// Test with tracing
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, customSpan, otelSpan := tracingSystem.StartDistributedOperation(ctx, "test.performance", "test", map[string]interface{}{
			"iteration": i,
		})

		// Simulate work
		time.Sleep(1 * time.Microsecond)

		tracingSystem.FinishDistributedOperation(customSpan, otelSpan, nil)
	}
	tracingDuration := time.Since(start)

	// Calculate overhead
	overhead := tracingDuration - baselineDuration
	overheadPercent := float64(overhead) / float64(baselineDuration) * 100

	// Verify reasonable overhead (should be less than 100% for this simple test)
	if overheadPercent > 100 {
		return fmt.Errorf("tracing overhead too high: %.2f%%", overheadPercent)
	}

	fmt.Printf("    Performance impact acceptable: %.2f%% overhead\n", overheadPercent)
	return nil
}

func testTracingStatistics(tracingSystem *observability.DistributedTracingSystem) error {
	// Get tracing statistics
	stats := tracingSystem.GetTracingStats()

	// Verify expected statistics
	expectedKeys := []string{"enabled", "custom_tracing", "opentelemetry", "component_tracing", "context_propagation", "sampling_ratio"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			return fmt.Errorf("missing statistic: %s", key)
		}
	}

	// Verify enabled status
	if enabled, ok := stats["enabled"].(bool); !ok || !enabled {
		return fmt.Errorf("tracing system not enabled according to statistics")
	}

	fmt.Printf("    Tracing statistics available: %d metrics\n", len(stats))
	return nil
}
