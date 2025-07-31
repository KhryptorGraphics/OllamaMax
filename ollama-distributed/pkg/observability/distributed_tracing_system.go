package observability

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// DistributedTracingSystem manages distributed tracing across all components
type DistributedTracingSystem struct {
	config *DistributedTracingConfig

	// Core tracing components
	customTracer *Tracer
	otelAdapter  *OpenTelemetryAdapter

	// Component tracers
	schedulerTracer *SchedulerTracer
	p2pTracer       *P2PTracer
	consensusTracer *ConsensusTracer
	apiTracer       *APITracer
	modelTracer     *ModelTracer

	// Context propagation
	propagator propagation.TextMapPropagator

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// DistributedTracingConfig configures the distributed tracing system
type DistributedTracingConfig struct {
	// Service configuration
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	Environment    string `json:"environment"`
	NodeID         string `json:"node_id"`

	// Custom tracing configuration
	CustomTracingConfig *TracerConfig `json:"custom_tracing_config"`

	// OpenTelemetry configuration
	OpenTelemetryConfig *OpenTelemetryConfig `json:"opentelemetry_config"`

	// Features
	EnableCustomTracing      bool `json:"enable_custom_tracing"`
	EnableOpenTelemetry      bool `json:"enable_opentelemetry"`
	EnableContextPropagation bool `json:"enable_context_propagation"`
	EnableComponentTracing   bool `json:"enable_component_tracing"`

	// Sampling
	SamplingRatio float64 `json:"sampling_ratio"`
}

// DistributedTraceContext represents distributed trace context
type DistributedTraceContext struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Sampled    bool              `json:"sampled"`
	Baggage    map[string]string `json:"baggage,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NewDistributedTracingSystem creates a new distributed tracing system
func NewDistributedTracingSystem(config *DistributedTracingConfig) *DistributedTracingSystem {
	if config == nil {
		config = DefaultDistributedTracingConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DistributedTracingSystem{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// DefaultDistributedTracingConfig returns default distributed tracing configuration
func DefaultDistributedTracingConfig() *DistributedTracingConfig {
	return &DistributedTracingConfig{
		ServiceName:    "ollama-distributed",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		NodeID:         "node-1",
		CustomTracingConfig: &TracerConfig{
			ServiceName:     "ollama-distributed",
			ServiceVersion:  "1.0.0",
			Environment:     "development",
			SamplingRate:    1.0,
			EnableExport:    false,
			ExportInterval:  10 * time.Second,
			ExportBatchSize: 100,
			ExportTimeout:   30 * time.Second,
			MaxSpans:        1000,
			SpanTTL:         time.Hour,
		},
		OpenTelemetryConfig:      DefaultOpenTelemetryConfig(),
		EnableCustomTracing:      true,
		EnableOpenTelemetry:      true,
		EnableContextPropagation: true,
		EnableComponentTracing:   true,
		SamplingRatio:            1.0,
	}
}

// Start starts the distributed tracing system
func (dts *DistributedTracingSystem) Start() error {
	dts.mu.Lock()
	defer dts.mu.Unlock()

	if dts.started {
		return nil
	}

	// Initialize custom tracer
	if dts.config.EnableCustomTracing {
		dts.customTracer = NewTracer(dts.config.CustomTracingConfig)
	}

	// Initialize OpenTelemetry adapter
	if dts.config.EnableOpenTelemetry {
		dts.otelAdapter = NewOpenTelemetryAdapter(dts.config.OpenTelemetryConfig, dts.customTracer)
		if err := dts.otelAdapter.Start(); err != nil {
			return fmt.Errorf("failed to start OpenTelemetry adapter: %w", err)
		}
	}

	// Initialize component tracers
	if dts.config.EnableComponentTracing {
		dts.schedulerTracer = NewSchedulerTracer(dts.customTracer, dts.otelAdapter)
		dts.p2pTracer = NewP2PTracer(dts.customTracer, dts.otelAdapter)
		dts.consensusTracer = NewConsensusTracer(dts.customTracer, dts.otelAdapter)
		dts.apiTracer = NewAPITracer(dts.customTracer, dts.otelAdapter)
		dts.modelTracer = NewModelTracer(dts.customTracer, dts.otelAdapter)
	}

	// Initialize context propagation
	if dts.config.EnableContextPropagation {
		dts.propagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	dts.started = true
	log.Info().
		Str("service", dts.config.ServiceName).
		Bool("custom_tracing", dts.config.EnableCustomTracing).
		Bool("opentelemetry", dts.config.EnableOpenTelemetry).
		Bool("component_tracing", dts.config.EnableComponentTracing).
		Float64("sampling_ratio", dts.config.SamplingRatio).
		Msg("Distributed tracing system started")

	return nil
}

// Stop stops the distributed tracing system
func (dts *DistributedTracingSystem) Stop() error {
	dts.mu.Lock()
	defer dts.mu.Unlock()

	if !dts.started {
		return nil
	}

	// Stop OpenTelemetry adapter
	if dts.otelAdapter != nil {
		if err := dts.otelAdapter.Stop(); err != nil {
			log.Error().Err(err).Msg("Failed to stop OpenTelemetry adapter")
		}
	}

	// Stop custom tracer
	if dts.customTracer != nil {
		// Custom tracer doesn't need explicit stopping
		log.Info().Msg("Custom tracer stopped")
	}

	dts.cancel()
	dts.started = false
	log.Info().Msg("Distributed tracing system stopped")
	return nil
}

// GetSchedulerTracer returns the scheduler tracer
func (dts *DistributedTracingSystem) GetSchedulerTracer() *SchedulerTracer {
	return dts.schedulerTracer
}

// GetP2PTracer returns the P2P tracer
func (dts *DistributedTracingSystem) GetP2PTracer() *P2PTracer {
	return dts.p2pTracer
}

// GetConsensusTracer returns the consensus tracer
func (dts *DistributedTracingSystem) GetConsensusTracer() *ConsensusTracer {
	return dts.consensusTracer
}

// GetAPITracer returns the API tracer
func (dts *DistributedTracingSystem) GetAPITracer() *APITracer {
	return dts.apiTracer
}

// GetModelTracer returns the model tracer
func (dts *DistributedTracingSystem) GetModelTracer() *ModelTracer {
	return dts.modelTracer
}

// GetCustomTracer returns the custom tracer
func (dts *DistributedTracingSystem) GetCustomTracer() *Tracer {
	return dts.customTracer
}

// GetOpenTelemetryAdapter returns the OpenTelemetry adapter
func (dts *DistributedTracingSystem) GetOpenTelemetryAdapter() *OpenTelemetryAdapter {
	return dts.otelAdapter
}

// Context propagation methods

// InjectTraceContext injects trace context into HTTP headers
func (dts *DistributedTracingSystem) InjectTraceContext(ctx context.Context, headers http.Header) {
	if !dts.started || dts.propagator == nil {
		return
	}

	dts.propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractTraceContext extracts trace context from HTTP headers
func (dts *DistributedTracingSystem) ExtractTraceContext(ctx context.Context, headers http.Header) context.Context {
	if !dts.started || dts.propagator == nil {
		return ctx
	}

	return dts.propagator.Extract(ctx, propagation.HeaderCarrier(headers))
}

// InjectTraceContextToMap injects trace context into a map
func (dts *DistributedTracingSystem) InjectTraceContextToMap(ctx context.Context, carrier map[string]string) {
	if !dts.started || dts.propagator == nil {
		return
	}

	dts.propagator.Inject(ctx, propagation.MapCarrier(carrier))
}

// ExtractTraceContextFromMap extracts trace context from a map
func (dts *DistributedTracingSystem) ExtractTraceContextFromMap(ctx context.Context, carrier map[string]string) context.Context {
	if !dts.started || dts.propagator == nil {
		return ctx
	}

	return dts.propagator.Extract(ctx, propagation.MapCarrier(carrier))
}

// Utility methods

// StartDistributedOperation starts a distributed operation with both custom and OpenTelemetry tracing
func (dts *DistributedTracingSystem) StartDistributedOperation(ctx context.Context, operationName, component string, attributes map[string]interface{}) (context.Context, *Span, oteltrace.Span) {
	if !dts.started {
		return ctx, nil, nil
	}

	// Start custom span
	var customSpan *Span
	var newCtx context.Context = ctx

	if dts.customTracer != nil {
		customSpan, newCtx = dts.customTracer.StartSpan(ctx, operationName)
		customSpan.SetTag("component", component)

		// Add attributes
		for key, value := range attributes {
			customSpan.SetTag(key, value)
		}
	}

	// Start OpenTelemetry span
	var otelSpan oteltrace.Span
	if dts.otelAdapter != nil && customSpan != nil {
		newCtx, otelSpan = dts.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// FinishDistributedOperation finishes a distributed operation
func (dts *DistributedTracingSystem) FinishDistributedOperation(customSpan *Span, otelSpan oteltrace.Span, err error) {
	// Finish custom span
	if customSpan != nil && dts.customTracer != nil {
		if err != nil {
			customSpan.SetStatus(SpanStatusCodeError, err.Error())
			customSpan.SetTag("error", true)
			customSpan.LogFields(map[string]interface{}{
				"error.message": err.Error(),
				"error.type":    fmt.Sprintf("%T", err),
			})
		}
		dts.customTracer.FinishSpan(customSpan)
	}

	// Finish OpenTelemetry span
	if otelSpan != nil && otelSpan.IsRecording() {
		if err != nil {
			otelSpan.RecordError(err)
			otelSpan.SetStatus(codes.Error, err.Error())
		} else {
			otelSpan.SetStatus(codes.Ok, "")
		}
		otelSpan.End()
	}
}

// AddDistributedEvent adds an event to both tracing systems
func (dts *DistributedTracingSystem) AddDistributedEvent(customSpan *Span, otelSpan oteltrace.Span, eventName string, attributes map[string]interface{}) {
	// Add to custom span
	if customSpan != nil {
		customSpan.LogFields(attributes)
	}

	// Add to OpenTelemetry span
	if otelSpan != nil && otelSpan.IsRecording() {
		// Convert attributes to OpenTelemetry format
		// (implementation similar to component_tracing.go)
	}
}

// GetTraceContext extracts trace context information
func (dts *DistributedTracingSystem) GetTraceContext(ctx context.Context) *DistributedTraceContext {
	// Extract from custom span
	if customSpan := SpanFromContext(ctx); customSpan != nil {
		return &DistributedTraceContext{
			TraceID:  customSpan.TraceID,
			SpanID:   customSpan.SpanID,
			ParentID: customSpan.ParentID,
			Sampled:  true, // Simplified
			Baggage:  customSpan.BaggageItems,
		}
	}

	// Extract from OpenTelemetry span
	if otelSpan := oteltrace.SpanFromContext(ctx); otelSpan.SpanContext().IsValid() {
		spanContext := otelSpan.SpanContext()
		return &DistributedTraceContext{
			TraceID: spanContext.TraceID().String(),
			SpanID:  spanContext.SpanID().String(),
			Sampled: spanContext.IsSampled(),
		}
	}

	return nil
}

// IsEnabled returns whether distributed tracing is enabled
func (dts *DistributedTracingSystem) IsEnabled() bool {
	dts.mu.RLock()
	defer dts.mu.RUnlock()
	return dts.started
}

// GetTracingStats returns tracing statistics
func (dts *DistributedTracingSystem) GetTracingStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["enabled"] = dts.IsEnabled()
	stats["custom_tracing"] = dts.config.EnableCustomTracing
	stats["opentelemetry"] = dts.config.EnableOpenTelemetry
	stats["component_tracing"] = dts.config.EnableComponentTracing
	stats["context_propagation"] = dts.config.EnableContextPropagation
	stats["sampling_ratio"] = dts.config.SamplingRatio

	if dts.customTracer != nil {
		// Add custom tracer stats if available
		stats["custom_tracer_started"] = true
	}

	if dts.otelAdapter != nil {
		stats["opentelemetry_started"] = dts.otelAdapter.IsEnabled()
	}

	return stats
}
