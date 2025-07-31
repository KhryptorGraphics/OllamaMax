package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// ComponentTracer provides tracing capabilities for distributed system components
type ComponentTracer struct {
	tracer        *Tracer
	otelAdapter   *OpenTelemetryAdapter
	componentName string
	enableOtel    bool
}

// NewComponentTracer creates a new component tracer
func NewComponentTracer(componentName string, tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *ComponentTracer {
	return &ComponentTracer{
		tracer:        tracer,
		otelAdapter:   otelAdapter,
		componentName: componentName,
		enableOtel:    otelAdapter != nil && otelAdapter.IsEnabled(),
	}
}

// SchedulerTracer provides tracing for scheduler operations
type SchedulerTracer struct {
	*ComponentTracer
}

// P2PTracer provides tracing for P2P operations
type P2PTracer struct {
	*ComponentTracer
}

// ConsensusTracer provides tracing for consensus operations
type ConsensusTracer struct {
	*ComponentTracer
}

// APITracer provides tracing for API operations
type APITracer struct {
	*ComponentTracer
}

// ModelTracer provides tracing for model operations
type ModelTracer struct {
	*ComponentTracer
}

// NewSchedulerTracer creates a new scheduler tracer
func NewSchedulerTracer(tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *SchedulerTracer {
	return &SchedulerTracer{
		ComponentTracer: NewComponentTracer("scheduler", tracer, otelAdapter),
	}
}

// NewP2PTracer creates a new P2P tracer
func NewP2PTracer(tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *P2PTracer {
	return &P2PTracer{
		ComponentTracer: NewComponentTracer("p2p", tracer, otelAdapter),
	}
}

// NewConsensusTracer creates a new consensus tracer
func NewConsensusTracer(tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *ConsensusTracer {
	return &ConsensusTracer{
		ComponentTracer: NewComponentTracer("consensus", tracer, otelAdapter),
	}
}

// NewAPITracer creates a new API tracer
func NewAPITracer(tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *APITracer {
	return &APITracer{
		ComponentTracer: NewComponentTracer("api", tracer, otelAdapter),
	}
}

// NewModelTracer creates a new model tracer
func NewModelTracer(tracer *Tracer, otelAdapter *OpenTelemetryAdapter) *ModelTracer {
	return &ModelTracer{
		ComponentTracer: NewComponentTracer("model", tracer, otelAdapter),
	}
}

// Scheduler tracing methods

// TraceTaskScheduling traces task scheduling operations
func (st *SchedulerTracer) TraceTaskScheduling(ctx context.Context, taskID, taskType string) (context.Context, *Span, oteltrace.Span) {
	operationName := "scheduler.schedule_task"

	// Start custom span
	customSpan, newCtx := st.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "scheduler")
	customSpan.SetTag("operation", "schedule_task")
	customSpan.SetTag("task.id", taskID)
	customSpan.SetTag("task.type", taskType)
	customSpan.SetTag("node.id", st.componentName)

	var otelSpan oteltrace.Span
	if st.enableOtel {
		newCtx, otelSpan = st.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TraceTaskExecution traces task execution
func (st *SchedulerTracer) TraceTaskExecution(ctx context.Context, taskID, workerID string) (context.Context, *Span, oteltrace.Span) {
	operationName := "scheduler.execute_task"

	customSpan, newCtx := st.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "scheduler")
	customSpan.SetTag("operation", "execute_task")
	customSpan.SetTag("task.id", taskID)
	customSpan.SetTag("worker.id", workerID)

	var otelSpan oteltrace.Span
	if st.enableOtel {
		newCtx, otelSpan = st.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TraceLoadBalancing traces load balancing decisions
func (st *SchedulerTracer) TraceLoadBalancing(ctx context.Context, algorithm string, candidateCount int) (context.Context, *Span, oteltrace.Span) {
	operationName := "scheduler.load_balance"

	customSpan, newCtx := st.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "scheduler")
	customSpan.SetTag("operation", "load_balance")
	customSpan.SetTag("algorithm", algorithm)
	customSpan.SetTag("candidate.count", candidateCount)

	var otelSpan oteltrace.Span
	if st.enableOtel {
		newCtx, otelSpan = st.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// P2P tracing methods

// TraceMessageSend traces P2P message sending
func (pt *P2PTracer) TraceMessageSend(ctx context.Context, messageType, peerID string, messageSize int) (context.Context, *Span, oteltrace.Span) {
	operationName := "p2p.send_message"

	customSpan, newCtx := pt.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "p2p")
	customSpan.SetTag("operation", "send_message")
	customSpan.SetTag("message.type", messageType)
	customSpan.SetTag("peer.id", peerID)
	customSpan.SetTag("message.size", messageSize)

	var otelSpan oteltrace.Span
	if pt.enableOtel {
		newCtx, otelSpan = pt.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TraceMessageReceive traces P2P message receiving
func (pt *P2PTracer) TraceMessageReceive(ctx context.Context, messageType, peerID string, messageSize int) (context.Context, *Span, oteltrace.Span) {
	operationName := "p2p.receive_message"

	customSpan, newCtx := pt.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "p2p")
	customSpan.SetTag("operation", "receive_message")
	customSpan.SetTag("message.type", messageType)
	customSpan.SetTag("peer.id", peerID)
	customSpan.SetTag("message.size", messageSize)

	var otelSpan oteltrace.Span
	if pt.enableOtel {
		newCtx, otelSpan = pt.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TracePeerDiscovery traces peer discovery operations
func (pt *P2PTracer) TracePeerDiscovery(ctx context.Context, discoveryType string, peerCount int) (context.Context, *Span, oteltrace.Span) {
	operationName := "p2p.peer_discovery"

	customSpan, newCtx := pt.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "p2p")
	customSpan.SetTag("operation", "peer_discovery")
	customSpan.SetTag("discovery.type", discoveryType)
	customSpan.SetTag("peer.count", peerCount)

	var otelSpan oteltrace.Span
	if pt.enableOtel {
		newCtx, otelSpan = pt.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// Consensus tracing methods

// TraceLeaderElection traces leader election process
func (ct *ConsensusTracer) TraceLeaderElection(ctx context.Context, term int64, candidateID string) (context.Context, *Span, oteltrace.Span) {
	operationName := "consensus.leader_election"

	customSpan, newCtx := ct.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "consensus")
	customSpan.SetTag("operation", "leader_election")
	customSpan.SetTag("term", term)
	customSpan.SetTag("candidate.id", candidateID)

	var otelSpan oteltrace.Span
	if ct.enableOtel {
		newCtx, otelSpan = ct.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TraceLogReplication traces log replication
func (ct *ConsensusTracer) TraceLogReplication(ctx context.Context, logIndex int64, entryCount int) (context.Context, *Span, oteltrace.Span) {
	operationName := "consensus.log_replication"

	customSpan, newCtx := ct.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "consensus")
	customSpan.SetTag("operation", "log_replication")
	customSpan.SetTag("log.index", logIndex)
	customSpan.SetTag("entry.count", entryCount)

	var otelSpan oteltrace.Span
	if ct.enableOtel {
		newCtx, otelSpan = ct.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// API tracing methods

// TraceHTTPRequest traces HTTP API requests
func (at *APITracer) TraceHTTPRequest(ctx context.Context, method, path, userAgent string) (context.Context, *Span, oteltrace.Span) {
	operationName := fmt.Sprintf("api.%s %s", method, path)

	customSpan, newCtx := at.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "api")
	customSpan.SetTag("operation", "http_request")
	customSpan.SetTag("http.method", method)
	customSpan.SetTag("http.path", path)
	customSpan.SetTag("http.user_agent", userAgent)

	var otelSpan oteltrace.Span
	if at.enableOtel {
		newCtx, otelSpan = at.otelAdapter.AdaptSpan(newCtx, customSpan)

		// Add standard HTTP attributes for OpenTelemetry
		if otelSpan.IsRecording() {
			otelSpan.SetAttributes(
				attribute.String("http.method", method),
				attribute.String("http.route", path),
				attribute.String("http.user_agent", userAgent),
			)
		}
	}

	return newCtx, customSpan, otelSpan
}

// TraceWebSocketConnection traces WebSocket connections
func (at *APITracer) TraceWebSocketConnection(ctx context.Context, connectionID, clientIP string) (context.Context, *Span, oteltrace.Span) {
	operationName := "api.websocket_connection"

	customSpan, newCtx := at.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "api")
	customSpan.SetTag("operation", "websocket_connection")
	customSpan.SetTag("connection.id", connectionID)
	customSpan.SetTag("client.ip", clientIP)

	var otelSpan oteltrace.Span
	if at.enableOtel {
		newCtx, otelSpan = at.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// Model tracing methods

// TraceModelLoad traces model loading operations
func (mt *ModelTracer) TraceModelLoad(ctx context.Context, modelName, modelVersion string, modelSize int64) (context.Context, *Span, oteltrace.Span) {
	operationName := "model.load"

	customSpan, newCtx := mt.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "model")
	customSpan.SetTag("operation", "load")
	customSpan.SetTag("model.name", modelName)
	customSpan.SetTag("model.version", modelVersion)
	customSpan.SetTag("model.size", modelSize)

	var otelSpan oteltrace.Span
	if mt.enableOtel {
		newCtx, otelSpan = mt.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// TraceInference traces model inference operations
func (mt *ModelTracer) TraceInference(ctx context.Context, modelName, requestID string, inputTokens, outputTokens int) (context.Context, *Span, oteltrace.Span) {
	operationName := "model.inference"

	customSpan, newCtx := mt.tracer.StartSpan(ctx, operationName)
	customSpan.SetTag("component", "model")
	customSpan.SetTag("operation", "inference")
	customSpan.SetTag("model.name", modelName)
	customSpan.SetTag("request.id", requestID)
	customSpan.SetTag("input.tokens", inputTokens)
	customSpan.SetTag("output.tokens", outputTokens)

	var otelSpan oteltrace.Span
	if mt.enableOtel {
		newCtx, otelSpan = mt.otelAdapter.AdaptSpan(newCtx, customSpan)
	}

	return newCtx, customSpan, otelSpan
}

// Utility methods for all tracers

// FinishSpans finishes both custom and OpenTelemetry spans
func (ct *ComponentTracer) FinishSpans(customSpan *Span, otelSpan oteltrace.Span, err error) {
	// Finish custom span
	if customSpan != nil {
		if err != nil {
			customSpan.SetStatus(SpanStatusCodeError, err.Error())
			customSpan.SetTag("error", true)
			customSpan.LogFields(map[string]interface{}{
				"error.message": err.Error(),
				"error.type":    fmt.Sprintf("%T", err),
			})
		}
		ct.tracer.FinishSpan(customSpan)
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

// AddEvent adds an event to both spans
func (ct *ComponentTracer) AddEvent(customSpan *Span, otelSpan oteltrace.Span, eventName string, attributes map[string]interface{}) {
	// Add to custom span
	if customSpan != nil {
		customSpan.LogFields(attributes)
	}

	// Add to OpenTelemetry span
	if otelSpan != nil && otelSpan.IsRecording() {
		attrs := make([]attribute.KeyValue, 0, len(attributes))
		for key, value := range attributes {
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
		otelSpan.AddEvent(eventName, oteltrace.WithAttributes(attrs...))
	}
}
