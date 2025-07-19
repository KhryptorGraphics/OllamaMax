package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	ollamaapi "github.com/ollama/ollama/api"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// Stub types for missing llm package
type CompletionRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream,omitempty"`
}

type CompletionResponse struct {
	Text     string `json:"text"`
	Finished bool   `json:"finished"`
	Error    string `json:"error,omitempty"`
}

// DistributedRunner extends the existing runner functionality for distributed execution
type DistributedRunner struct {
	scheduler       *scheduler.Engine
	integrationLayer *IntegrationLayer
	
	// Runner state
	mu              sync.RWMutex
	activeRunners   map[string]*RunnerInstance
	runnerPool      chan *RunnerInstance
	maxRunners      int
	
	// Metrics
	requestCount    int64
	successCount    int64
	failureCount    int64
	averageLatency  time.Duration
}

// RunnerInstance represents a runner instance on a specific node
type RunnerInstance struct {
	ID          string
	NodeID      string
	ModelName   string
	Status      string
	Created     time.Time
	LastUsed    time.Time
	RequestChan chan *RunnerRequest
	ResponseChan chan *RunnerResponse
}

// RunnerRequest represents a request to a runner
type RunnerRequest struct {
	ID       string
	Type     string
	Context  context.Context
	Payload  interface{}
	Response chan *RunnerResponse
}

// RunnerResponse represents a response from a runner
type RunnerResponse struct {
	ID      string
	Success bool
	Data    interface{}
	Error   string
	NodeID  string
	Latency time.Duration
}

// NewDistributedRunner creates a new distributed runner
func NewDistributedRunner(scheduler *scheduler.Engine, integrationLayer *IntegrationLayer) *DistributedRunner {
	return &DistributedRunner{
		scheduler:        scheduler,
		integrationLayer: integrationLayer,
		activeRunners:    make(map[string]*RunnerInstance),
		runnerPool:       make(chan *RunnerInstance, 100),
		maxRunners:       10,
	}
}

// Start starts the distributed runner
func (dr *DistributedRunner) Start(ctx context.Context) error {
	// Start runner pool manager
	go dr.manageRunnerPool(ctx)
	
	// Start metrics collection
	go dr.collectMetrics(ctx)
	
	return nil
}

// ExecuteRequest executes a request using distributed runners
func (dr *DistributedRunner) ExecuteRequest(req *RunnerRequest) (*RunnerResponse, error) {
	startTime := time.Now()
	
	// Get or create runner instance
	runner, err := dr.getRunner(req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get runner: %w", err)
	}
	
	// Send request to runner
	select {
	case runner.RequestChan <- req:
		// Request sent successfully
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("runner request timeout")
	}
	
	// Wait for response
	select {
	case resp := <-req.Response:
		resp.Latency = time.Since(startTime)
		dr.updateMetrics(resp)
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("runner response timeout")
	}
}

// HandleGeneration handles generation requests
func (dr *DistributedRunner) HandleGeneration(c *gin.Context) {
	var req ollamaapi.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create runner request
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("gen_%d", time.Now().UnixNano()),
		Type:     "generate",
		Context:  c.Request.Context(),
		Payload:  req,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dr.ExecuteRequest(runnerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if resp.Success {
		c.Header("X-Ollama-Node", resp.NodeID)
		c.Header("X-Ollama-Latency", resp.Latency.String())
		c.JSON(http.StatusOK, resp.Data)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
	}
}

// HandleChat handles chat requests
func (dr *DistributedRunner) HandleChat(c *gin.Context) {
	var req ollamaapi.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create runner request
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("chat_%d", time.Now().UnixNano()),
		Type:     "chat",
		Context:  c.Request.Context(),
		Payload:  req,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dr.ExecuteRequest(runnerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if resp.Success {
		c.Header("X-Ollama-Node", resp.NodeID)
		c.Header("X-Ollama-Latency", resp.Latency.String())
		c.JSON(http.StatusOK, resp.Data)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
	}
}

// HandleEmbedding handles embedding requests
func (dr *DistributedRunner) HandleEmbedding(c *gin.Context) {
	var req ollamaapi.EmbedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Create runner request
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("embed_%d", time.Now().UnixNano()),
		Type:     "embed",
		Context:  c.Request.Context(),
		Payload:  req,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dr.ExecuteRequest(runnerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	if resp.Success {
		c.Header("X-Ollama-Node", resp.NodeID)
		c.Header("X-Ollama-Latency", resp.Latency.String())
		c.JSON(http.StatusOK, resp.Data)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": resp.Error})
	}
}

// getRunner gets or creates a runner instance
func (dr *DistributedRunner) getRunner(requestType string) (*RunnerInstance, error) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	// Try to get from pool
	select {
	case runner := <-dr.runnerPool:
		runner.LastUsed = time.Now()
		return runner, nil
	default:
		// Create new runner if pool is empty
		if len(dr.activeRunners) < dr.maxRunners {
			return dr.createRunner(requestType)
		}
		return nil, fmt.Errorf("max runners reached")
	}
}

// createRunner creates a new runner instance
func (dr *DistributedRunner) createRunner(requestType string) (*RunnerInstance, error) {
	// Get available node from scheduler
	nodes := dr.scheduler.GetAvailableNodes()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes")
	}
	
	// Select best node (for now, just use first)
	nodeID := nodes[0].ID
	
	runner := &RunnerInstance{
		ID:           fmt.Sprintf("runner_%d", time.Now().UnixNano()),
		NodeID:       nodeID,
		ModelName:    "auto", // Will be set based on request
		Status:       "active",
		Created:      time.Now(),
		LastUsed:     time.Now(),
		RequestChan:  make(chan *RunnerRequest, 10),
		ResponseChan: make(chan *RunnerResponse, 10),
	}
	
	// Start runner goroutine
	go dr.runRunner(runner)
	
	dr.activeRunners[runner.ID] = runner
	
	return runner, nil
}

// runRunner runs a runner instance
func (dr *DistributedRunner) runRunner(runner *RunnerInstance) {
	defer func() {
		dr.mu.Lock()
		delete(dr.activeRunners, runner.ID)
		dr.mu.Unlock()
	}()
	
	for {
		select {
		case req := <-runner.RequestChan:
			// Process request
			resp := dr.processRequest(runner, req)
			
			// Send response
			select {
			case req.Response <- resp:
				// Response sent
			case <-time.After(5 * time.Second):
				// Response timeout
			}
			
		case <-time.After(30 * time.Second):
			// Runner idle timeout
			return
		}
	}
}

// processRequest processes a request on a runner
func (dr *DistributedRunner) processRequest(runner *RunnerInstance, req *RunnerRequest) *RunnerResponse {
	startTime := time.Now()
	
	// Create scheduler request
	// Convert interface{} payload to map[string]interface{}
	var payload map[string]interface{}
	if req.Payload != nil {
		if p, ok := req.Payload.(map[string]interface{}); ok {
			payload = p
		} else {
			// If payload is not a map, wrap it
			payload = map[string]interface{}{
				"data": req.Payload,
			}
		}
	}
	
	schedReq := &scheduler.Request{
		ID:         req.ID,
		Type:       req.Type,
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload:    payload,
	}
	
	// Schedule request
	if err := dr.scheduler.Schedule(schedReq); err != nil {
		return &RunnerResponse{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
			NodeID:  runner.NodeID,
			Latency: time.Since(startTime),
		}
	}
	
	// Wait for response
	select {
	case schedResp := <-schedReq.ResponseCh:
		return &RunnerResponse{
			ID:      req.ID,
			Success: schedResp.Success,
			Data:    schedResp.Data,
			Error:   schedResp.Error,
			NodeID:  schedResp.NodeID,
			Latency: time.Since(startTime),
		}
	case <-time.After(30 * time.Second):
		return &RunnerResponse{
			ID:      req.ID,
			Success: false,
			Error:   "scheduler timeout",
			NodeID:  runner.NodeID,
			Latency: time.Since(startTime),
		}
	}
}

// manageRunnerPool manages the runner pool
func (dr *DistributedRunner) manageRunnerPool(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dr.cleanupRunners()
		}
	}
}

// cleanupRunners cleans up inactive runners
func (dr *DistributedRunner) cleanupRunners() {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	now := time.Now()
	for id, runner := range dr.activeRunners {
		if now.Sub(runner.LastUsed) > 5*time.Minute {
			delete(dr.activeRunners, id)
		}
	}
}

// collectMetrics collects performance metrics
func (dr *DistributedRunner) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dr.calculateMetrics()
		}
	}
}

// calculateMetrics calculates performance metrics
func (dr *DistributedRunner) calculateMetrics() {
	// TODO: Implement metrics calculation
	// This would calculate average latency, throughput, etc.
}

// updateMetrics updates metrics with response data
func (dr *DistributedRunner) updateMetrics(resp *RunnerResponse) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	dr.requestCount++
	if resp.Success {
		dr.successCount++
	} else {
		dr.failureCount++
	}
	
	// Update average latency (simple moving average)
	if dr.requestCount == 1 {
		dr.averageLatency = resp.Latency
	} else {
		dr.averageLatency = (dr.averageLatency*time.Duration(dr.requestCount-1) + resp.Latency) / time.Duration(dr.requestCount)
	}
}

// GetStats returns runner statistics
func (dr *DistributedRunner) GetStats() map[string]interface{} {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	
	successRate := float64(0)
	if dr.requestCount > 0 {
		successRate = float64(dr.successCount) / float64(dr.requestCount) * 100
	}
	
	return map[string]interface{}{
		"active_runners":   len(dr.activeRunners),
		"max_runners":      dr.maxRunners,
		"request_count":    dr.requestCount,
		"success_count":    dr.successCount,
		"failure_count":    dr.failureCount,
		"success_rate":     successRate,
		"average_latency":  dr.averageLatency.String(),
		"pool_size":        len(dr.runnerPool),
	}
}

// GetActiveRunners returns information about active runners
func (dr *DistributedRunner) GetActiveRunners() map[string]*RunnerInstance {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	
	result := make(map[string]*RunnerInstance)
	for k, v := range dr.activeRunners {
		result[k] = v
	}
	return result
}

// SetMaxRunners sets the maximum number of runners
func (dr *DistributedRunner) SetMaxRunners(max int) {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	dr.maxRunners = max
}

// Shutdown gracefully shuts down the distributed runner
func (dr *DistributedRunner) Shutdown(ctx context.Context) error {
	dr.mu.Lock()
	defer dr.mu.Unlock()
	
	// Close all active runners
	for _, runner := range dr.activeRunners {
		close(runner.RequestChan)
		close(runner.ResponseChan)
	}
	
	// Clear runners
	dr.activeRunners = make(map[string]*RunnerInstance)
	
	return nil
}

// DistributedRunnerAdapter adapts the distributed runner to work with existing Ollama server
type DistributedRunnerAdapter struct {
	runner *DistributedRunner
}

// NewDistributedRunnerAdapter creates a new adapter
func NewDistributedRunnerAdapter(runner *DistributedRunner) *DistributedRunnerAdapter {
	return &DistributedRunnerAdapter{
		runner: runner,
	}
}

// Completion implements the LlamaServer interface for compatibility
func (dra *DistributedRunnerAdapter) Completion(ctx context.Context, req CompletionRequest, fn func(CompletionResponse)) error {
	// Convert to RunnerRequest
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("comp_%d", time.Now().UnixNano()),
		Type:     "completion",
		Context:  ctx,
		Payload:  req,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dra.runner.ExecuteRequest(runnerReq)
	if err != nil {
		return err
	}
	
	if !resp.Success {
		return fmt.Errorf("completion failed: %s", resp.Error)
	}
	
	// Convert response back to CompletionResponse
	if completionResp, ok := resp.Data.(CompletionResponse); ok {
		fn(completionResp)
	}
	
	return nil
}

// Embedding implements the LlamaServer interface for compatibility
func (dra *DistributedRunnerAdapter) Embedding(ctx context.Context, prompt string) ([]float32, error) {
	// Convert to RunnerRequest
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("embed_%d", time.Now().UnixNano()),
		Type:     "embedding",
		Context:  ctx,
		Payload:  prompt,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dra.runner.ExecuteRequest(runnerReq)
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("embedding failed: %s", resp.Error)
	}
	
	// Convert response back to embedding
	if embedding, ok := resp.Data.([]float32); ok {
		return embedding, nil
	}
	
	return nil, fmt.Errorf("invalid embedding response")
}

// Tokenize implements the LlamaServer interface for compatibility
func (dra *DistributedRunnerAdapter) Tokenize(ctx context.Context, content string) ([]int, error) {
	// Convert to RunnerRequest
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("tokenize_%d", time.Now().UnixNano()),
		Type:     "tokenize",
		Context:  ctx,
		Payload:  content,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dra.runner.ExecuteRequest(runnerReq)
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("tokenization failed: %s", resp.Error)
	}
	
	// Convert response back to tokens
	if tokens, ok := resp.Data.([]int); ok {
		return tokens, nil
	}
	
	return nil, fmt.Errorf("invalid tokenization response")
}

// Detokenize implements the LlamaServer interface for compatibility
func (dra *DistributedRunnerAdapter) Detokenize(ctx context.Context, tokens []int) (string, error) {
	// Convert to RunnerRequest
	runnerReq := &RunnerRequest{
		ID:       fmt.Sprintf("detokenize_%d", time.Now().UnixNano()),
		Type:     "detokenize",
		Context:  ctx,
		Payload:  tokens,
		Response: make(chan *RunnerResponse, 1),
	}
	
	// Execute request
	resp, err := dra.runner.ExecuteRequest(runnerReq)
	if err != nil {
		return "", err
	}
	
	if !resp.Success {
		return "", fmt.Errorf("detokenization failed: %s", resp.Error)
	}
	
	// Convert response back to string
	if text, ok := resp.Data.(string); ok {
		return text, nil
	}
	
	return "", fmt.Errorf("invalid detokenization response")
}

// Close implements the LlamaServer interface for compatibility
func (dra *DistributedRunnerAdapter) Close() error {
	return dra.runner.Shutdown(context.Background())
}