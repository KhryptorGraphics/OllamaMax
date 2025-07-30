package protocols

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// InferenceHandler handles model inference requests
type InferenceHandler struct {
	// Model registry and execution
	modelRegistry   *ModelRegistry
	executionEngine InferenceExecutor

	// Request tracking
	activeRequests map[string]*InferenceRequest
	requestsMux    sync.RWMutex

	// Configuration
	config *InferenceConfig

	// Metrics
	metrics *InferenceMetrics
}

// InferenceConfig configures inference handling
type InferenceConfig struct {
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	ModelLoadTimeout      time.Duration `json:"model_load_timeout"`
	MaxPromptLength       int           `json:"max_prompt_length"`
	MaxResponseLength     int           `json:"max_response_length"`
	EnableCaching         bool          `json:"enable_caching"`
	CacheSize             int           `json:"cache_size"`
	CacheTTL              time.Duration `json:"cache_ttl"`
}

// InferenceMetrics tracks inference performance
type InferenceMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	ActiveRequests     int           `json:"active_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	ModelLoadTime      time.Duration `json:"model_load_time"`
	InferenceTime      time.Duration `json:"inference_time"`
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`

	// Per-model metrics
	ModelMetrics map[string]*ModelInferenceMetrics `json:"model_metrics"`

	mu sync.RWMutex
}

// ModelInferenceMetrics tracks metrics for specific models
type ModelInferenceMetrics struct {
	RequestCount   int64         `json:"request_count"`
	SuccessCount   int64         `json:"success_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
	LoadCount      int64         `json:"load_count"`
	LastLoadTime   time.Duration `json:"last_load_time"`
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	ID          string                 `json:"id"`
	ModelName   string                 `json:"model_name"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	RequesterID peer.ID                `json:"requester_id"`

	// Execution context
	Context    context.Context    `json:"-"`
	CancelFunc context.CancelFunc `json:"-"`

	// Progress tracking
	Status      RequestStatus `json:"status"`
	Progress    float64       `json:"progress"`
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt time.Time     `json:"completed_at"`

	// Results
	Response *InferenceResponse `json:"response,omitempty"`
	Error    string             `json:"error,omitempty"`
}

// RequestStatus represents the status of an inference request
type RequestStatus string

const (
	StatusPending   RequestStatus = "pending"
	StatusLoading   RequestStatus = "loading"
	StatusExecuting RequestStatus = "executing"
	StatusCompleted RequestStatus = "completed"
	StatusFailed    RequestStatus = "failed"
	StatusCancelled RequestStatus = "cancelled"
)

// InferenceResponse represents an inference response
type InferenceResponse struct {
	RequestID    string                 `json:"request_id"`
	ModelName    string                 `json:"model_name"`
	Response     string                 `json:"response"`
	TokensUsed   int                    `json:"tokens_used"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata"`
	GeneratedAt  time.Time              `json:"generated_at"`

	// Performance metrics
	LoadTime      time.Duration `json:"load_time"`
	InferenceTime time.Duration `json:"inference_time"`
	TotalTime     time.Duration `json:"total_time"`
}

// InferenceExecutor defines the interface for executing model inference
type InferenceExecutor interface {
	LoadModel(ctx context.Context, modelName string) error
	UnloadModel(modelName string) error
	IsModelLoaded(modelName string) bool
	ExecuteInference(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error)
	GetModelInfo(modelName string) (*ModelInfo, error)
	ListLoadedModels() []string
}

// ModelInfo represents information about a model
type ModelInfo struct {
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	Type         string            `json:"type"`
	Architecture string            `json:"architecture"`
	Parameters   int64             `json:"parameters"`
	IsLoaded     bool              `json:"is_loaded"`
	LoadTime     time.Duration     `json:"load_time"`
	MemoryUsage  int64             `json:"memory_usage"`
	Metadata     map[string]string `json:"metadata"`
}

// ModelRegistry manages available models
type ModelRegistry struct {
	models    map[string]*ModelInfo
	modelsMux sync.RWMutex
}

// NewInferenceHandler creates a new inference handler
func NewInferenceHandler(executor InferenceExecutor, config *InferenceConfig) *InferenceHandler {
	if config == nil {
		config = DefaultInferenceConfig()
	}

	return &InferenceHandler{
		modelRegistry:   NewModelRegistry(),
		executionEngine: executor,
		activeRequests:  make(map[string]*InferenceRequest),
		config:          config,
		metrics: &InferenceMetrics{
			ModelMetrics: make(map[string]*ModelInferenceMetrics),
		},
	}
}

// HandleMessage handles inference protocol messages
func (ih *InferenceHandler) HandleMessage(ctx context.Context, stream network.Stream, msg *Message) error {
	switch msg.Type {
	case MsgTypeInferenceRequest:
		return ih.handleInferenceRequest(ctx, stream, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleInferenceRequest handles incoming inference requests
func (ih *InferenceHandler) handleInferenceRequest(ctx context.Context, stream network.Stream, msg *Message) error {
	start := time.Now()

	// Parse inference request
	req, err := ih.parseInferenceRequest(msg)
	if err != nil {
		return ih.sendErrorResponse(stream, msg.ID, "invalid_request", err.Error())
	}

	// Validate request
	if err := ih.validateInferenceRequest(req); err != nil {
		return ih.sendErrorResponse(stream, msg.ID, "validation_error", err.Error())
	}

	// Check capacity
	if !ih.checkCapacity() {
		return ih.sendErrorResponse(stream, msg.ID, "capacity_exceeded", "Too many concurrent requests")
	}

	// Create request context
	req.Context, req.CancelFunc = context.WithTimeout(ctx, req.Timeout)
	req.CreatedAt = time.Now()
	req.Status = StatusPending
	req.RequesterID = stream.Conn().RemotePeer()

	// Track request
	ih.trackRequest(req)
	defer ih.untrackRequest(req.ID)

	// Update metrics
	ih.updateRequestMetrics()

	// Execute inference
	response, err := ih.executeInference(req)
	if err != nil {
		ih.updateErrorMetrics(req.ModelName)
		return ih.sendErrorResponse(stream, msg.ID, "execution_error", err.Error())
	}

	// Send response
	if err := ih.sendInferenceResponse(stream, msg.ID, response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	// Update success metrics
	ih.updateSuccessMetrics(req.ModelName, time.Since(start))

	log.Printf("Completed inference request %s for model %s (latency: %v)", req.ID, req.ModelName, time.Since(start))
	return nil
}

// parseInferenceRequest parses an inference request from message data
func (ih *InferenceHandler) parseInferenceRequest(msg *Message) (*InferenceRequest, error) {
	data := msg.Data

	req := &InferenceRequest{
		ID: msg.ID,
	}

	// Extract required fields
	if modelName, ok := data["model_name"].(string); ok {
		req.ModelName = modelName
	} else {
		return nil, fmt.Errorf("model_name is required")
	}

	if prompt, ok := data["prompt"].(string); ok {
		req.Prompt = prompt
	} else {
		return nil, fmt.Errorf("prompt is required")
	}

	// Extract optional fields
	if params, ok := data["parameters"].(map[string]interface{}); ok {
		req.Parameters = params
	} else {
		req.Parameters = make(map[string]interface{})
	}

	if priority, ok := data["priority"].(float64); ok {
		req.Priority = int(priority)
	} else {
		req.Priority = 0
	}

	if timeoutStr, ok := data["timeout"].(string); ok {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			req.Timeout = timeout
		} else {
			req.Timeout = ih.config.RequestTimeout
		}
	} else {
		req.Timeout = ih.config.RequestTimeout
	}

	return req, nil
}

// validateInferenceRequest validates an inference request
func (ih *InferenceHandler) validateInferenceRequest(req *InferenceRequest) error {
	if req.ModelName == "" {
		return fmt.Errorf("model name is required")
	}

	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	if len(req.Prompt) > ih.config.MaxPromptLength {
		return fmt.Errorf("prompt length exceeds maximum of %d characters", ih.config.MaxPromptLength)
	}

	if req.Timeout <= 0 || req.Timeout > ih.config.RequestTimeout {
		req.Timeout = ih.config.RequestTimeout
	}

	return nil
}

// checkCapacity checks if we can handle another request
func (ih *InferenceHandler) checkCapacity() bool {
	ih.requestsMux.RLock()
	defer ih.requestsMux.RUnlock()

	return len(ih.activeRequests) < ih.config.MaxConcurrentRequests
}

// trackRequest tracks an active request
func (ih *InferenceHandler) trackRequest(req *InferenceRequest) {
	ih.requestsMux.Lock()
	defer ih.requestsMux.Unlock()

	ih.activeRequests[req.ID] = req
}

// untrackRequest removes a request from tracking
func (ih *InferenceHandler) untrackRequest(requestID string) {
	ih.requestsMux.Lock()
	defer ih.requestsMux.Unlock()

	delete(ih.activeRequests, requestID)
}

// executeInference executes the inference request
func (ih *InferenceHandler) executeInference(req *InferenceRequest) (*InferenceResponse, error) {
	start := time.Now()

	// Check if model is loaded
	if !ih.executionEngine.IsModelLoaded(req.ModelName) {
		req.Status = StatusLoading
		loadStart := time.Now()

		if err := ih.executionEngine.LoadModel(req.Context, req.ModelName); err != nil {
			return nil, fmt.Errorf("failed to load model %s: %w", req.ModelName, err)
		}

		loadTime := time.Since(loadStart)
		ih.updateModelLoadMetrics(req.ModelName, loadTime)
		log.Printf("Loaded model %s in %v", req.ModelName, loadTime)
	}

	// Execute inference
	req.Status = StatusExecuting
	req.StartedAt = time.Now()

	response, err := ih.executionEngine.ExecuteInference(req.Context, req)
	if err != nil {
		req.Status = StatusFailed
		req.Error = err.Error()
		return nil, fmt.Errorf("inference execution failed: %w", err)
	}

	// Update response timing
	response.TotalTime = time.Since(start)
	response.InferenceTime = time.Since(req.StartedAt)
	response.GeneratedAt = time.Now()

	req.Status = StatusCompleted
	req.CompletedAt = time.Now()
	req.Response = response

	return response, nil
}

// sendInferenceResponse sends an inference response
func (ih *InferenceHandler) sendInferenceResponse(stream network.Stream, requestID string, response *InferenceResponse) error {
	responseMsg := &Message{
		Type:      MsgTypeInferenceResponse,
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":     requestID,
			"model_name":     response.ModelName,
			"response":       response.Response,
			"tokens_used":    response.TokensUsed,
			"finish_reason":  response.FinishReason,
			"metadata":       response.Metadata,
			"generated_at":   response.GeneratedAt,
			"load_time":      response.LoadTime.String(),
			"inference_time": response.InferenceTime.String(),
			"total_time":     response.TotalTime.String(),
		},
	}

	handler := NewProtocolHandler(InferenceProtocol)
	return handler.SendMessage(stream, responseMsg)
}

// sendErrorResponse sends an error response
func (ih *InferenceHandler) sendErrorResponse(stream network.Stream, requestID, errorCode, errorMessage string) error {
	errorMsg := &Message{
		Type:      "error",
		ID:        generateMessageID(),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"request_id":    requestID,
			"error_code":    errorCode,
			"error_message": errorMessage,
		},
	}

	handler := NewProtocolHandler(InferenceProtocol)
	return handler.SendMessage(stream, errorMsg)
}

// Metrics update methods

func (ih *InferenceHandler) updateRequestMetrics() {
	ih.metrics.mu.Lock()
	defer ih.metrics.mu.Unlock()

	ih.metrics.TotalRequests++
	ih.metrics.ActiveRequests = len(ih.activeRequests)
}

func (ih *InferenceHandler) updateSuccessMetrics(modelName string, latency time.Duration) {
	ih.metrics.mu.Lock()
	defer ih.metrics.mu.Unlock()

	ih.metrics.SuccessfulRequests++
	ih.metrics.ActiveRequests = len(ih.activeRequests)

	// Update average latency
	if ih.metrics.SuccessfulRequests > 0 {
		totalLatency := ih.metrics.AverageLatency * time.Duration(ih.metrics.SuccessfulRequests-1)
		ih.metrics.AverageLatency = (totalLatency + latency) / time.Duration(ih.metrics.SuccessfulRequests)
	} else {
		ih.metrics.AverageLatency = latency
	}

	// Update model-specific metrics
	if ih.metrics.ModelMetrics[modelName] == nil {
		ih.metrics.ModelMetrics[modelName] = &ModelInferenceMetrics{}
	}

	modelMetrics := ih.metrics.ModelMetrics[modelName]
	modelMetrics.RequestCount++
	modelMetrics.SuccessCount++
	modelMetrics.LastUsed = time.Now()

	// Update model average latency
	if modelMetrics.SuccessCount > 0 {
		totalLatency := modelMetrics.AverageLatency * time.Duration(modelMetrics.SuccessCount-1)
		modelMetrics.AverageLatency = (totalLatency + latency) / time.Duration(modelMetrics.SuccessCount)
	} else {
		modelMetrics.AverageLatency = latency
	}
}

func (ih *InferenceHandler) updateErrorMetrics(modelName string) {
	ih.metrics.mu.Lock()
	defer ih.metrics.mu.Unlock()

	ih.metrics.FailedRequests++
	ih.metrics.ActiveRequests = len(ih.activeRequests)

	// Update model-specific error metrics
	if ih.metrics.ModelMetrics[modelName] == nil {
		ih.metrics.ModelMetrics[modelName] = &ModelInferenceMetrics{}
	}

	modelMetrics := ih.metrics.ModelMetrics[modelName]
	modelMetrics.RequestCount++
	modelMetrics.ErrorCount++
}

func (ih *InferenceHandler) updateModelLoadMetrics(modelName string, loadTime time.Duration) {
	ih.metrics.mu.Lock()
	defer ih.metrics.mu.Unlock()

	ih.metrics.ModelLoadTime = loadTime

	// Update model-specific load metrics
	if ih.metrics.ModelMetrics[modelName] == nil {
		ih.metrics.ModelMetrics[modelName] = &ModelInferenceMetrics{}
	}

	modelMetrics := ih.metrics.ModelMetrics[modelName]
	modelMetrics.LoadCount++
	modelMetrics.LastLoadTime = loadTime
}

// GetMetrics returns a copy of the current metrics
func (ih *InferenceHandler) GetMetrics() *InferenceMetrics {
	ih.metrics.mu.RLock()
	defer ih.metrics.mu.RUnlock()

	// Create deep copy
	metricsCopy := &InferenceMetrics{
		TotalRequests:      ih.metrics.TotalRequests,
		SuccessfulRequests: ih.metrics.SuccessfulRequests,
		FailedRequests:     ih.metrics.FailedRequests,
		ActiveRequests:     ih.metrics.ActiveRequests,
		AverageLatency:     ih.metrics.AverageLatency,
		ModelLoadTime:      ih.metrics.ModelLoadTime,
		InferenceTime:      ih.metrics.InferenceTime,
		CacheHits:          ih.metrics.CacheHits,
		CacheMisses:        ih.metrics.CacheMisses,
		ModelMetrics:       make(map[string]*ModelInferenceMetrics),
	}

	// Copy model metrics
	for modelName, modelMetrics := range ih.metrics.ModelMetrics {
		metricsCopy.ModelMetrics[modelName] = &ModelInferenceMetrics{
			RequestCount:   modelMetrics.RequestCount,
			SuccessCount:   modelMetrics.SuccessCount,
			ErrorCount:     modelMetrics.ErrorCount,
			AverageLatency: modelMetrics.AverageLatency,
			LastUsed:       modelMetrics.LastUsed,
			LoadCount:      modelMetrics.LoadCount,
			LastLoadTime:   modelMetrics.LastLoadTime,
		}
	}

	return metricsCopy
}

// GetActiveRequests returns currently active requests
func (ih *InferenceHandler) GetActiveRequests() map[string]*InferenceRequest {
	ih.requestsMux.RLock()
	defer ih.requestsMux.RUnlock()

	requests := make(map[string]*InferenceRequest)
	for id, req := range ih.activeRequests {
		// Create copy without context
		reqCopy := *req
		reqCopy.Context = nil
		reqCopy.CancelFunc = nil
		requests[id] = &reqCopy
	}

	return requests
}

// CancelRequest cancels an active request
func (ih *InferenceHandler) CancelRequest(requestID string) error {
	ih.requestsMux.Lock()
	defer ih.requestsMux.Unlock()

	req, exists := ih.activeRequests[requestID]
	if !exists {
		return fmt.Errorf("request %s not found", requestID)
	}

	if req.CancelFunc != nil {
		req.CancelFunc()
		req.Status = StatusCancelled
	}

	return nil
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelInfo),
	}
}

// RegisterModel registers a model in the registry
func (mr *ModelRegistry) RegisterModel(info *ModelInfo) {
	mr.modelsMux.Lock()
	defer mr.modelsMux.Unlock()

	mr.models[info.Name] = info
}

// GetModel retrieves model information
func (mr *ModelRegistry) GetModel(name string) (*ModelInfo, bool) {
	mr.modelsMux.RLock()
	defer mr.modelsMux.RUnlock()

	info, exists := mr.models[name]
	return info, exists
}

// ListModels returns all registered models
func (mr *ModelRegistry) ListModels() []*ModelInfo {
	mr.modelsMux.RLock()
	defer mr.modelsMux.RUnlock()

	models := make([]*ModelInfo, 0, len(mr.models))
	for _, info := range mr.models {
		models = append(models, info)
	}

	return models
}

// DefaultInferenceConfig returns default inference configuration
func DefaultInferenceConfig() *InferenceConfig {
	return &InferenceConfig{
		MaxConcurrentRequests: 10,
		RequestTimeout:        5 * time.Minute,
		ModelLoadTimeout:      2 * time.Minute,
		MaxPromptLength:       50000,
		MaxResponseLength:     100000,
		EnableCaching:         true,
		CacheSize:             1000,
		CacheTTL:              1 * time.Hour,
	}
}

// InferenceClient provides client-side inference functionality
type InferenceClient struct {
	protocolClient *ProtocolClient
}

// NewInferenceClient creates a new inference client
func NewInferenceClient(dialer StreamDialer, timeout time.Duration) *InferenceClient {
	return &InferenceClient{
		protocolClient: NewProtocolClient(dialer, InferenceProtocol, timeout),
	}
}

// SendInferenceRequest sends an inference request to a peer
func (ic *InferenceClient) SendInferenceRequest(ctx context.Context, peerID peer.ID, req *InferenceRequest) (*InferenceResponse, error) {
	// Create request message
	requestMsg := CreateRequestMessage(MsgTypeInferenceRequest, map[string]interface{}{
		"model_name": req.ModelName,
		"prompt":     req.Prompt,
		"parameters": req.Parameters,
		"priority":   req.Priority,
		"timeout":    req.Timeout.String(),
	})

	// Send request and wait for response
	responseMsg, err := ic.protocolClient.SendRequest(ctx, peerID, requestMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send inference request: %w", err)
	}

	// Handle error response
	if responseMsg.Type == "error" {
		errorCode, _ := responseMsg.Data["error_code"].(string)
		errorMessage, _ := responseMsg.Data["error_message"].(string)
		return nil, fmt.Errorf("inference error [%s]: %s", errorCode, errorMessage)
	}

	// Parse inference response
	return ic.parseInferenceResponse(responseMsg)
}

// parseInferenceResponse parses an inference response message
func (ic *InferenceClient) parseInferenceResponse(msg *Message) (*InferenceResponse, error) {
	data := msg.Data

	response := &InferenceResponse{}

	if requestID, ok := data["request_id"].(string); ok {
		response.RequestID = requestID
	}

	if modelName, ok := data["model_name"].(string); ok {
		response.ModelName = modelName
	}

	if responseText, ok := data["response"].(string); ok {
		response.Response = responseText
	}

	if tokensUsed, ok := data["tokens_used"].(float64); ok {
		response.TokensUsed = int(tokensUsed)
	}

	if finishReason, ok := data["finish_reason"].(string); ok {
		response.FinishReason = finishReason
	}

	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		response.Metadata = metadata
	}

	if generatedAtStr, ok := data["generated_at"].(string); ok {
		if generatedAt, err := time.Parse(time.RFC3339, generatedAtStr); err == nil {
			response.GeneratedAt = generatedAt
		}
	}

	// Parse timing information
	if loadTimeStr, ok := data["load_time"].(string); ok {
		if loadTime, err := time.ParseDuration(loadTimeStr); err == nil {
			response.LoadTime = loadTime
		}
	}

	if inferenceTimeStr, ok := data["inference_time"].(string); ok {
		if inferenceTime, err := time.ParseDuration(inferenceTimeStr); err == nil {
			response.InferenceTime = inferenceTime
		}
	}

	if totalTimeStr, ok := data["total_time"].(string); ok {
		if totalTime, err := time.ParseDuration(totalTimeStr); err == nil {
			response.TotalTime = totalTime
		}
	}

	return response, nil
}

// GetClientMetrics returns client metrics
func (ic *InferenceClient) GetClientMetrics() *ClientMetrics {
	return ic.protocolClient.GetClientMetrics()
}
