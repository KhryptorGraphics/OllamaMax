package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// OpenTelemetryAdapter adapts the custom tracing system to OpenTelemetry
type OpenTelemetryAdapter struct {
	config *OpenTelemetryConfig

	// OpenTelemetry components
	tracerProvider *trace.TracerProvider
	tracer         oteltrace.Tracer

	// Custom tracer integration
	customTracer *Tracer

	// Lifecycle
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// OpenTelemetryConfig configures OpenTelemetry integration
type OpenTelemetryConfig struct {
	// Service information
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	Environment    string `json:"environment"`
	NodeID         string `json:"node_id"`

	// Jaeger configuration
	JaegerEndpoint string `json:"jaeger_endpoint"`
	JaegerUser     string `json:"jaeger_user"`
	JaegerPassword string `json:"jaeger_password"`

	// Sampling configuration
	SamplingRatio float64 `json:"sampling_ratio"`

	// Features
	EnableOpenTelemetry bool `json:"enable_opentelemetry"`
	EnableJaegerExport  bool `json:"enable_jaeger_export"`
	EnablePropagation   bool `json:"enable_propagation"`
	EnableBatching      bool `json:"enable_batching"`

	// Performance settings
	BatchTimeout   time.Duration `json:"batch_timeout"`
	ExportTimeout  time.Duration `json:"export_timeout"`
	MaxExportBatch int           `json:"max_export_batch"`
	MaxQueueSize   int           `json:"max_queue_size"`
}

// NewOpenTelemetryAdapter creates a new OpenTelemetry adapter
func NewOpenTelemetryAdapter(config *OpenTelemetryConfig, customTracer *Tracer) *OpenTelemetryAdapter {
	if config == nil {
		config = DefaultOpenTelemetryConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &OpenTelemetryAdapter{
		config:       config,
		customTracer: customTracer,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// DefaultOpenTelemetryConfig returns default OpenTelemetry configuration
func DefaultOpenTelemetryConfig() *OpenTelemetryConfig {
	return &OpenTelemetryConfig{
		ServiceName:         "ollama-distributed",
		ServiceVersion:      "1.0.0",
		Environment:         "development",
		NodeID:              "node-1",
		JaegerEndpoint:      "http://localhost:14268/api/traces",
		SamplingRatio:       1.0, // Sample all traces in development
		EnableOpenTelemetry: true,
		EnableJaegerExport:  true,
		EnablePropagation:   true,
		EnableBatching:      true,
		BatchTimeout:        5 * time.Second,
		ExportTimeout:       30 * time.Second,
		MaxExportBatch:      512,
		MaxQueueSize:        2048,
	}
}

// Start starts the OpenTelemetry adapter
func (ota *OpenTelemetryAdapter) Start() error {
	ota.mu.Lock()
	defer ota.mu.Unlock()

	if ota.started {
		return nil
	}

	if !ota.config.EnableOpenTelemetry {
		log.Info().Msg("OpenTelemetry integration disabled")
		return nil
	}

	// Create resource
	res, err := ota.createResource()
	if err != nil {
		return fmt.Errorf("failed to create OpenTelemetry resource: %w", err)
	}

	// Create tracer provider options
	opts := []trace.TracerProviderOption{
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(ota.config.SamplingRatio)),
	}

	// Add Jaeger exporter if enabled
	if ota.config.EnableJaegerExport {
		exporter, err := ota.createJaegerExporter()
		if err != nil {
			return fmt.Errorf("failed to create Jaeger exporter: %w", err)
		}

		if ota.config.EnableBatching {
			opts = append(opts, trace.WithBatcher(exporter,
				trace.WithBatchTimeout(ota.config.BatchTimeout),
				trace.WithExportTimeout(ota.config.ExportTimeout),
				trace.WithMaxExportBatchSize(ota.config.MaxExportBatch),
				trace.WithMaxQueueSize(ota.config.MaxQueueSize),
			))
		} else {
			opts = append(opts, trace.WithSyncer(exporter))
		}
	}

	// Create tracer provider
	ota.tracerProvider = trace.NewTracerProvider(opts...)

	// Set global tracer provider
	otel.SetTracerProvider(ota.tracerProvider)

	// Set global propagator
	if ota.config.EnablePropagation {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	// Create tracer
	ota.tracer = ota.tracerProvider.Tracer(
		ota.config.ServiceName,
		oteltrace.WithInstrumentationVersion(ota.config.ServiceVersion),
	)

	ota.started = true
	log.Info().
		Str("service", ota.config.ServiceName).
		Str("jaeger_endpoint", ota.config.JaegerEndpoint).
		Float64("sampling_ratio", ota.config.SamplingRatio).
		Bool("jaeger_export", ota.config.EnableJaegerExport).
		Msg("OpenTelemetry adapter started")

	return nil
}

// Stop stops the OpenTelemetry adapter
func (ota *OpenTelemetryAdapter) Stop() error {
	ota.mu.Lock()
	defer ota.mu.Unlock()

	if !ota.started {
		return nil
	}

	// Shutdown tracer provider
	if ota.tracerProvider != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := ota.tracerProvider.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown OpenTelemetry tracer provider")
			return err
		}
	}

	ota.cancel()
	ota.started = false
	log.Info().Msg("OpenTelemetry adapter stopped")
	return nil
}

// AdaptSpan adapts a custom span to OpenTelemetry
func (ota *OpenTelemetryAdapter) AdaptSpan(ctx context.Context, customSpan *Span) (context.Context, oteltrace.Span) {
	if !ota.started || ota.tracer == nil {
		return ctx, oteltrace.SpanFromContext(ctx)
	}

	// Start OpenTelemetry span
	otelCtx, otelSpan := ota.tracer.Start(ctx, customSpan.OperationName)

	// Copy attributes from custom span
	attrs := make([]attribute.KeyValue, 0, len(customSpan.Tags)+3)
	attrs = append(attrs,
		attribute.String("service.name", customSpan.ServiceName),
		attribute.String("trace.id", customSpan.TraceID),
		attribute.String("span.id", customSpan.SpanID),
	)

	if customSpan.ParentID != "" {
		attrs = append(attrs, attribute.String("parent.id", customSpan.ParentID))
	}

	// Convert custom tags to OpenTelemetry attributes
	for key, value := range customSpan.Tags {
		switch v := value.(type) {
		case string:
			attrs = append(attrs, attribute.String(key, v))
		case int:
			attrs = append(attrs, attribute.Int(key, v))
		case int64:
			attrs = append(attrs, attribute.Int64(key, v))
		case float64:
			attrs = append(attrs, attribute.Float64(key, v))
		case bool:
			attrs = append(attrs, attribute.Bool(key, v))
		default:
			attrs = append(attrs, attribute.String(key, fmt.Sprintf("%v", v)))
		}
	}

	otelSpan.SetAttributes(attrs...)

	// Copy logs as events
	for _, logEntry := range customSpan.Logs {
		eventAttrs := make([]attribute.KeyValue, 0, len(logEntry.Fields))
		for key, value := range logEntry.Fields {
			switch v := value.(type) {
			case string:
				eventAttrs = append(eventAttrs, attribute.String(key, v))
			case int:
				eventAttrs = append(eventAttrs, attribute.Int(key, v))
			case int64:
				eventAttrs = append(eventAttrs, attribute.Int64(key, v))
			case float64:
				eventAttrs = append(eventAttrs, attribute.Float64(key, v))
			case bool:
				eventAttrs = append(eventAttrs, attribute.Bool(key, v))
			default:
				eventAttrs = append(eventAttrs, attribute.String(key, fmt.Sprintf("%v", v)))
			}
		}

		otelSpan.AddEvent("log", oteltrace.WithAttributes(eventAttrs...), oteltrace.WithTimestamp(logEntry.Timestamp))
	}

	// Set status
	if customSpan.Status.Code == SpanStatusCodeError {
		// otelSpan.SetStatus(codes.Error, customSpan.Status.Message)
		// Disabled due to missing OpenTelemetry dependencies
	} else {
		// otelSpan.SetStatus(codes.Ok, customSpan.Status.Message)
		// Disabled due to missing OpenTelemetry dependencies
	}

	return otelCtx, otelSpan
}

// CreateSpanFromOtel creates a custom span from OpenTelemetry span
func (ota *OpenTelemetryAdapter) CreateSpanFromOtel(otelSpan oteltrace.Span, operationName, serviceName string) *Span {
	spanContext := otelSpan.SpanContext()

	customSpan := &Span{
		TraceID:       spanContext.TraceID().String(),
		SpanID:        spanContext.SpanID().String(),
		OperationName: operationName,
		ServiceName:   serviceName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		BaggageItems:  make(map[string]string),
		Status: SpanStatus{
			Code:    SpanStatusCodeOK,
			Message: "",
		},
	}

	return customSpan
}

// GetTracer returns the OpenTelemetry tracer
func (ota *OpenTelemetryAdapter) GetTracer() oteltrace.Tracer {
	return ota.tracer
}

// IsEnabled returns whether OpenTelemetry is enabled
func (ota *OpenTelemetryAdapter) IsEnabled() bool {
	ota.mu.RLock()
	defer ota.mu.RUnlock()
	return ota.started && ota.config.EnableOpenTelemetry
}

// createResource creates the OpenTelemetry resource
func (ota *OpenTelemetryAdapter) createResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ota.config.ServiceName),
			semconv.ServiceVersion(ota.config.ServiceVersion),
			semconv.DeploymentEnvironment(ota.config.Environment),
			attribute.String("node.id", ota.config.NodeID),
		),
	)
}

// createJaegerExporter creates the Jaeger exporter
func (ota *OpenTelemetryAdapter) createJaegerExporter() (trace.SpanExporter, error) {
	opts := []jaeger.CollectorEndpointOption{
		jaeger.WithEndpoint(ota.config.JaegerEndpoint),
	}

	if ota.config.JaegerUser != "" && ota.config.JaegerPassword != "" {
		opts = append(opts, jaeger.WithUsername(ota.config.JaegerUser))
		opts = append(opts, jaeger.WithPassword(ota.config.JaegerPassword))
	}

	return jaeger.New(jaeger.WithCollectorEndpoint(opts...))
}
