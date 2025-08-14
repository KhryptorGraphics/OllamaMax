package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// TraceContext represents tracing context information
type TraceContext struct {
	TraceID  string `json:"trace_id"`
	SpanID   string `json:"span_id"`
	ParentID string `json:"parent_id,omitempty"`
	Sampled  bool   `json:"sampled"`
}

// Span represents a single span in a trace
type Span struct {
	TraceID       string `json:"trace_id"`
	SpanID        string `json:"span_id"`
	ParentID      string `json:"parent_id,omitempty"`
	OperationName string `json:"operation_name"`
	ServiceName   string `json:"service_name"`

	// Timing
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`

	// Status
	Status     SpanStatus     `json:"status"`
	StatusCode SpanStatusCode `json:"status_code"`

	// Metadata
	Tags map[string]interface{} `json:"tags"`
	Logs []SpanLog              `json:"logs"`

	// Context
	BaggageItems map[string]string `json:"baggage_items"`

	// Internal
	finished bool
	mu       sync.RWMutex
}

// SpanStatus represents the status of a span
type SpanStatus struct {
	Code    SpanStatusCode `json:"code"`
	Message string         `json:"message,omitempty"`
}

// SpanStatusCode represents span status codes
type SpanStatusCode int

const (
	SpanStatusCodeUnset SpanStatusCode = iota
	SpanStatusCodeOK
	SpanStatusCodeError
)

// SpanLog represents a log entry within a span
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// Tracer provides distributed tracing capabilities
type Tracer struct {
	serviceName string
	config      *TracerConfig
	spans       map[string]*Span
	exporters   []SpanExporter
	sampler     Sampler
	mu          sync.RWMutex

	// Background processing
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TracerConfig configures the tracer
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Sampling
	SamplingRate  float64
	SamplingRules []SamplingRule

	// Export
	EnableExport    bool
	ExportInterval  time.Duration
	ExportBatchSize int
	ExportTimeout   time.Duration

	// Resource limits
	MaxSpans int
	SpanTTL  time.Duration
}

// SamplingRule defines sampling rules
type SamplingRule struct {
	ServicePattern   string
	OperationPattern string
	SamplingRate     float64
}

// Sampler determines if a trace should be sampled
type Sampler interface {
	ShouldSample(ctx context.Context, traceID string, operationName string) bool
}

// SpanExporter exports spans to external systems
type SpanExporter interface {
	Export(ctx context.Context, spans []*Span) error
	Shutdown(ctx context.Context) error
}

// NewTracer creates a new tracer
func NewTracer(config *TracerConfig) *Tracer {
	if config == nil {
		config = &TracerConfig{
			ServiceName:     "ollama-distributed",
			SamplingRate:    1.0,
			EnableExport:    false,
			ExportInterval:  10 * time.Second,
			ExportBatchSize: 100,
			ExportTimeout:   30 * time.Second,
			MaxSpans:        10000,
			SpanTTL:         time.Hour,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	tracer := &Tracer{
		serviceName: config.ServiceName,
		config:      config,
		spans:       make(map[string]*Span),
		exporters:   make([]SpanExporter, 0),
		sampler:     NewProbabilitySampler(config.SamplingRate),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background tasks
	tracer.wg.Add(2)
	go tracer.exportLoop()
	go tracer.cleanupLoop()

	return tracer
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, operationName string) (*Span, context.Context) {
	// Extract parent context
	parentSpan := SpanFromContext(ctx)

	// Generate IDs
	var traceID, parentID string
	if parentSpan != nil {
		traceID = parentSpan.TraceID
		parentID = parentSpan.SpanID
	} else {
		traceID = generateID()
	}

	spanID := generateID()

	// Check sampling
	sampled := t.sampler.ShouldSample(ctx, traceID, operationName)

	span := &Span{
		TraceID:       traceID,
		SpanID:        spanID,
		ParentID:      parentID,
		OperationName: operationName,
		ServiceName:   t.serviceName,
		StartTime:     time.Now(),
		Status: SpanStatus{
			Code: SpanStatusCodeUnset,
		},
		Tags:         make(map[string]interface{}),
		Logs:         make([]SpanLog, 0),
		BaggageItems: make(map[string]string),
	}

	// Store span if sampled
	if sampled {
		t.mu.Lock()
		t.spans[spanID] = span
		t.mu.Unlock()
	}

	// Create new context with span
	newCtx := ContextWithSpan(ctx, span)

	return span, newCtx
}

// FinishSpan finishes a span
func (t *Tracer) FinishSpan(span *Span) {
	span.mu.Lock()
	defer span.mu.Unlock()

	if span.finished {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.finished = true

	// Set default status if not set
	if span.Status.Code == SpanStatusCodeUnset {
		span.Status.Code = SpanStatusCodeOK
	}
}

// AddExporter adds a span exporter
func (t *Tracer) AddExporter(exporter SpanExporter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.exporters = append(t.exporters, exporter)
}

// exportLoop periodically exports spans
func (t *Tracer) exportLoop() {
	defer t.wg.Done()

	if !t.config.EnableExport {
		return
	}

	ticker := time.NewTicker(t.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			t.exportAllSpans()
			return
		case <-ticker.C:
			t.exportFinishedSpans()
		}
	}
}

// exportFinishedSpans exports finished spans
func (t *Tracer) exportFinishedSpans() {
	t.mu.Lock()

	var finishedSpans []*Span
	for spanID, span := range t.spans {
		span.mu.RLock()
		if span.finished {
			finishedSpans = append(finishedSpans, span)
			delete(t.spans, spanID)
		}
		span.mu.RUnlock()

		if len(finishedSpans) >= t.config.ExportBatchSize {
			break
		}
	}

	exporters := make([]SpanExporter, len(t.exporters))
	copy(exporters, t.exporters)
	t.mu.Unlock()

	if len(finishedSpans) == 0 {
		return
	}

	// Export to all exporters
	for _, exporter := range exporters {
		go func(exp SpanExporter) {
			ctx, cancel := context.WithTimeout(context.Background(), t.config.ExportTimeout)
			defer cancel()

			if err := exp.Export(ctx, finishedSpans); err != nil {
				// Log export error
				fmt.Printf("Failed to export spans: %v\n", err)
			}
		}(exporter)
	}
}

// exportAllSpans exports all remaining spans
func (t *Tracer) exportAllSpans() {
	t.mu.Lock()

	var allSpans []*Span
	for _, span := range t.spans {
		allSpans = append(allSpans, span)
	}

	exporters := make([]SpanExporter, len(t.exporters))
	copy(exporters, t.exporters)
	t.mu.Unlock()

	if len(allSpans) == 0 {
		return
	}

	// Export to all exporters
	for _, exporter := range exporters {
		ctx, cancel := context.WithTimeout(context.Background(), t.config.ExportTimeout)
		exporter.Export(ctx, allSpans)
		cancel()
	}
}

// cleanupLoop periodically cleans up old spans
func (t *Tracer) cleanupLoop() {
	defer t.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.cleanupOldSpans()
		}
	}
}

// cleanupOldSpans removes old spans
func (t *Tracer) cleanupOldSpans() {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-t.config.SpanTTL)

	for spanID, span := range t.spans {
		span.mu.RLock()
		if span.StartTime.Before(cutoff) {
			delete(t.spans, spanID)
		}
		span.mu.RUnlock()
	}
}

// Close closes the tracer
func (t *Tracer) Close() error {
	t.cancel()
	t.wg.Wait()

	// Shutdown all exporters
	for _, exporter := range t.exporters {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		exporter.Shutdown(ctx)
		cancel()
	}

	return nil
}

// Span methods

// SetTag sets a tag on the span
func (s *Span) SetTag(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tags[key] = value
}

// SetStatus sets the span status
func (s *Span) SetStatus(code SpanStatusCode, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = SpanStatus{
		Code:    code,
		Message: message,
	}
}

// LogFields adds a log entry to the span
func (s *Span) LogFields(fields map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log := SpanLog{
		Timestamp: time.Now(),
		Fields:    fields,
	}

	s.Logs = append(s.Logs, log)
}

// SetBaggageItem sets a baggage item
func (s *Span) SetBaggageItem(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BaggageItems[key] = value
}

// GetBaggageItem gets a baggage item
func (s *Span) GetBaggageItem(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BaggageItems[key]
}

// Context utilities

type spanContextKey struct{}

// ContextWithSpan returns a context with the span
func ContextWithSpan(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// SpanFromContext extracts a span from context
func SpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value(spanContextKey{}).(*Span); ok {
		return span
	}
	return nil
}

// Sampling implementations

// ProbabilitySampler samples based on probability
type ProbabilitySampler struct {
	rate float64
}

// NewProbabilitySampler creates a new probability sampler
func NewProbabilitySampler(rate float64) *ProbabilitySampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}

	return &ProbabilitySampler{rate: rate}
}

// ShouldSample determines if a trace should be sampled
func (ps *ProbabilitySampler) ShouldSample(ctx context.Context, traceID string, operationName string) bool {
	if ps.rate == 0 {
		return false
	}
	if ps.rate == 1 {
		return true
	}

	// Use trace ID for consistent sampling decisions
	// This is a simplified implementation
	return len(traceID)%100 < int(ps.rate*100)
}

// Utility functions

// generateID generates a random ID
func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Helper functions for common operations

// StartOperation starts a new span for an operation
func StartOperation(ctx context.Context, tracer *Tracer, operationName string) (*Span, context.Context) {
	return tracer.StartSpan(ctx, operationName)
}

// FinishOperation finishes an operation span
func FinishOperation(span *Span, tracer *Tracer, err error) {
	if err != nil {
		span.SetStatus(SpanStatusCodeError, err.Error())
		span.SetTag("error", true)
		span.LogFields(map[string]interface{}{
			"error.message": err.Error(),
			"error.type":    fmt.Sprintf("%T", err),
		})
	}

	tracer.FinishSpan(span)
}

// TraceFunction traces a function execution
func TraceFunction(ctx context.Context, tracer *Tracer, functionName string, fn func(context.Context) error) error {
	span, newCtx := tracer.StartSpan(ctx, functionName)
	defer func() {
		tracer.FinishSpan(span)
	}()

	err := fn(newCtx)
	if err != nil {
		span.SetStatus(SpanStatusCodeError, err.Error())
		span.SetTag("error", true)
	}

	return err
}
