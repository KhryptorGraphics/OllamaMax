package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	ollamaapi "github.com/ollama/ollama/api"
	"github.com/ollama/ollama-distributed/pkg/models"
	"github.com/ollama/ollama-distributed/pkg/scheduler"
)

// IntegrationLayer provides transparent API layer for distributed Ollama
type IntegrationLayer struct {
	scheduler    *scheduler.Engine
	localProxy   *httputil.ReverseProxy
	localURL     *url.URL
	
	// Distributed mode settings
	distributedMode bool
	fallbackMode    bool
	
	// Request tracking
	requestTracker *RequestTracker
	
	// Model distribution
	modelDistribution *models.Manager
}

// RequestTracker tracks ongoing requests for failover
type RequestTracker struct {
	mu       sync.RWMutex
	requests map[string]*TrackedRequest
}

type TrackedRequest struct {
	ID       string
	Started  time.Time
	NodeID   string
	Model    string
	Retries  int
	Response chan *scheduler.Response
}

// NewIntegrationLayer creates a new API integration layer
func NewIntegrationLayer(scheduler *scheduler.Engine, localAddr string, modelDist *models.Manager) (*IntegrationLayer, error) {
	localURL, err := url.Parse(localAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid local Ollama URL: %w", err)
	}
	
	proxy := httputil.NewSingleHostReverseProxy(localURL)
	
	// Customize proxy to handle errors and add distributed headers
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("X-Ollama-Distributed-Error", "local-proxy-failed")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(gin.H{"error": "Local Ollama instance unavailable"})
	}
	
	return &IntegrationLayer{
		scheduler:         scheduler,
		localProxy:        proxy,
		localURL:          localURL,
		distributedMode:   true,
		fallbackMode:      true,
		requestTracker:    &RequestTracker{requests: make(map[string]*TrackedRequest)},
		modelDistribution: modelDist,
	}, nil
}

// HandleRequest processes API requests with transparent distributed routing
func (il *IntegrationLayer) HandleRequest(c *gin.Context) {
	path := c.Request.URL.Path
	
	// Add distributed headers
	c.Header("X-Ollama-Distributed", "true")
	c.Header("X-Ollama-Mode", il.getMode())
	
	// Route based on endpoint
	switch {
	case strings.HasPrefix(path, "/api/generate"):
		il.handleGenerate(c)
	case strings.HasPrefix(path, "/api/chat"):
		il.handleChat(c)
	case strings.HasPrefix(path, "/api/embed"):
		il.handleEmbed(c)
	case strings.HasPrefix(path, "/api/embeddings"):
		il.handleEmbeddings(c)
	case strings.HasPrefix(path, "/api/pull"):
		il.handlePull(c)
	case strings.HasPrefix(path, "/api/push"):
		il.handlePush(c)
	case strings.HasPrefix(path, "/api/show"):
		il.handleShow(c)
	case strings.HasPrefix(path, "/api/tags"):
		il.handleTags(c)
	case strings.HasPrefix(path, "/api/delete"):
		il.handleDelete(c)
	case strings.HasPrefix(path, "/api/copy"):
		il.handleCopy(c)
	case strings.HasPrefix(path, "/api/ps"):
		il.handlePs(c)
	case strings.HasPrefix(path, "/api/create"):
		il.handleCreate(c)
	case strings.HasPrefix(path, "/api/version"):
		il.handleVersion(c)
	case strings.HasPrefix(path, "/v1/"):
		// OpenAI compatibility - handle distributed
		il.handleOpenAI(c)
	default:
		// Unknown endpoint - fallback to local
		il.proxyToLocal(c)
	}
}

// Generate endpoint with distributed routing
func (il *IntegrationLayer) handleGenerate(c *gin.Context) {
	var req ollamaapi.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if model should be distributed
	if !il.shouldDistribute(req.Model) {
		il.proxyToLocal(c)
		return
	}
	
	// Create distributed request
	distribReq := &scheduler.Request{
		ID:         fmt.Sprintf("gen_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Type:       "generate",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"prompt":  req.Prompt,
			"stream":  req.Stream,
			"options": req.Options,
		},
	}
	
	// Track request
	il.requestTracker.TrackRequest(distribReq)
	defer il.requestTracker.UntrackRequest(distribReq.ID)
	
	// Schedule on distributed cluster
	if err := il.scheduler.Schedule(distribReq); err != nil {
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "scheduler-error")
			il.proxyToLocal(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Wait for response
	select {
	case response := <-distribReq.ResponseCh:
		if response.Success {
			c.Header("X-Ollama-Node", response.NodeID)
			c.JSON(http.StatusOK, response.Data)
		} else {
			if il.fallbackMode {
				c.Header("X-Ollama-Fallback", "execution-error")
				il.proxyToLocal(c)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "timeout")
			il.proxyToLocal(c)
			return
		}
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

// Chat endpoint with distributed routing
func (il *IntegrationLayer) handleChat(c *gin.Context) {
	var req ollamaapi.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if model should be distributed
	if !il.shouldDistribute(req.Model) {
		il.proxyToLocal(c)
		return
	}
	
	// Create distributed request
	distribReq := &scheduler.Request{
		ID:         fmt.Sprintf("chat_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Type:       "chat",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"messages": req.Messages,
			"stream":   req.Stream,
			"options":  req.Options,
		},
	}
	
	// Track and schedule
	il.requestTracker.TrackRequest(distribReq)
	defer il.requestTracker.UntrackRequest(distribReq.ID)
	
	if err := il.scheduler.Schedule(distribReq); err != nil {
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "scheduler-error")
			il.proxyToLocal(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Wait for response
	select {
	case response := <-distribReq.ResponseCh:
		if response.Success {
			c.Header("X-Ollama-Node", response.NodeID)
			c.JSON(http.StatusOK, response.Data)
		} else {
			if il.fallbackMode {
				c.Header("X-Ollama-Fallback", "execution-error")
				il.proxyToLocal(c)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "timeout")
			il.proxyToLocal(c)
		}
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

// Embed endpoint with distributed routing
func (il *IntegrationLayer) handleEmbed(c *gin.Context) {
	var req ollamaapi.EmbedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if !il.shouldDistribute(req.Model) {
		il.proxyToLocal(c)
		return
	}
	
	// Create distributed request
	distribReq := &scheduler.Request{
		ID:         fmt.Sprintf("embed_%d", time.Now().UnixNano()),
		ModelName:  req.Model,
		Type:       "embed",
		Priority:   1,
		Timeout:    30 * time.Second,
		ResponseCh: make(chan *scheduler.Response, 1),
		Payload: map[string]interface{}{
			"input":    req.Input,
			"truncate": req.Truncate,
			"options":  req.Options,
		},
	}
	
	il.requestTracker.TrackRequest(distribReq)
	defer il.requestTracker.UntrackRequest(distribReq.ID)
	
	if err := il.scheduler.Schedule(distribReq); err != nil {
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "scheduler-error")
			il.proxyToLocal(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	select {
	case response := <-distribReq.ResponseCh:
		if response.Success {
			c.Header("X-Ollama-Node", response.NodeID)
			c.JSON(http.StatusOK, response.Data)
		} else {
			if il.fallbackMode {
				c.Header("X-Ollama-Fallback", "execution-error")
				il.proxyToLocal(c)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": response.Error})
		}
	case <-time.After(30 * time.Second):
		if il.fallbackMode {
			c.Header("X-Ollama-Fallback", "timeout")
			il.proxyToLocal(c)
		}
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
	}
}

// Embeddings endpoint (compatibility)
func (il *IntegrationLayer) handleEmbeddings(c *gin.Context) {
	// Convert to EmbedRequest format and handle
	var req ollamaapi.EmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to EmbedRequest
	embedReq := ollamaapi.EmbedRequest{
		Model:     req.Model,
		Input:     req.Prompt,
		Options:   req.Options,
		KeepAlive: req.KeepAlive,
	}
	
	// Replace request body
	body, _ := json.Marshal(embedReq)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	
	il.handleEmbed(c)
}

// Pull endpoint with distributed model management
func (il *IntegrationLayer) handlePull(c *gin.Context) {
	var req ollamaapi.PullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if model should be distributed
	if il.modelDistribution.ShouldDistribute(req.Model) {
		// Handle distributed pull
		il.handleDistributedPull(c, req)
	} else {
		// Proxy to local
		il.proxyToLocal(c)
	}
}

// Push endpoint - proxy to local by default
func (il *IntegrationLayer) handlePush(c *gin.Context) {
	// Push operations are typically done locally
	il.proxyToLocal(c)
}

// Show endpoint with distributed model info
func (il *IntegrationLayer) handleShow(c *gin.Context) {
	var req ollamaapi.ShowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if model is distributed
	if il.modelDistribution.IsDistributed(req.Model) {
		// Get info from distributed cluster
		info := il.modelDistribution.GetModelInfo(req.Model)
		if info == nil {
			if il.fallbackMode {
				c.Header("X-Ollama-Fallback", "distributed-info-error")
				il.proxyToLocal(c)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Model info not found"})
			return
		}
		
		c.Header("X-Ollama-Distributed-Model", "true")
		c.JSON(http.StatusOK, info)
	} else {
		il.proxyToLocal(c)
	}
}

// Tags endpoint with distributed model listing
func (il *IntegrationLayer) handleTags(c *gin.Context) {
	// Get local models
	localModels, err := il.getLocalModels()
	if err != nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Get distributed models
	distributedModels := il.modelDistribution.GetDistributedModels()
	if distributedModels == nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get distributed models"})
		return
	}
	
	// Convert distributed models to ListModelResponse and merge
	var convertedDistributed []ollamaapi.ListModelResponse
	for _, model := range distributedModels {
		if modelMap, ok := model.(map[string]interface{}); ok {
			response := ollamaapi.ListModelResponse{
				Name:  getString(modelMap, "name"),
				Model: getString(modelMap, "name"),
				Size:  getInt64(modelMap, "size"),
			}
			convertedDistributed = append(convertedDistributed, response)
		}
	}
	
	// Merge models
	allModels := append(localModels, convertedDistributed...)
	
	c.Header("X-Ollama-Total-Models", fmt.Sprintf("%d", len(allModels)))
	c.Header("X-Ollama-Local-Models", fmt.Sprintf("%d", len(localModels)))
	c.Header("X-Ollama-Distributed-Models", fmt.Sprintf("%d", len(distributedModels)))
	
	c.JSON(http.StatusOK, ollamaapi.ListResponse{Models: allModels})
}

// Delete endpoint with distributed model cleanup
func (il *IntegrationLayer) handleDelete(c *gin.Context) {
	var req ollamaapi.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if model is distributed
	if il.modelDistribution.IsDistributed(req.Model) {
		// Delete from distributed cluster
		if err := il.modelDistribution.DeleteModel(req.Model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Model deleted from distributed cluster"})
	} else {
		// Delete locally
		il.proxyToLocal(c)
	}
}

// Copy endpoint - proxy to local
func (il *IntegrationLayer) handleCopy(c *gin.Context) {
	il.proxyToLocal(c)
}

// PS endpoint with distributed process info
func (il *IntegrationLayer) handlePs(c *gin.Context) {
	// Get local processes
	localProcs, err := il.getLocalProcesses()
	if err != nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Get distributed processes
	distributedProcs, err := il.getDistributedProcesses()
	if err != nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Merge processes
	allProcs := append(localProcs, distributedProcs...)
	
	c.Header("X-Ollama-Total-Processes", fmt.Sprintf("%d", len(allProcs)))
	c.Header("X-Ollama-Local-Processes", fmt.Sprintf("%d", len(localProcs)))
	c.Header("X-Ollama-Distributed-Processes", fmt.Sprintf("%d", len(distributedProcs)))
	
	c.JSON(http.StatusOK, ollamaapi.ProcessResponse{Models: allProcs})
}

// Create endpoint - proxy to local
func (il *IntegrationLayer) handleCreate(c *gin.Context) {
	il.proxyToLocal(c)
}

// Version endpoint with distributed info
func (il *IntegrationLayer) handleVersion(c *gin.Context) {
	// Get local version first
	localVersion, err := il.getLocalVersion()
	if err != nil {
		localVersion = map[string]interface{}{"version": "unknown"}
	}
	
	// Add distributed info
	version := map[string]interface{}{
		"version":            localVersion["version"],
		"distributed":        true,
		"distributed_mode":   il.distributedMode,
		"fallback_mode":      il.fallbackMode,
		"cluster_size":       il.scheduler.GetClusterSize(),
		"active_nodes":       il.scheduler.GetActiveNodes(),
		"distributed_models": il.modelDistribution.GetDistributedModelCount(),
	}
	
	c.JSON(http.StatusOK, version)
}

// OpenAI compatibility endpoints
func (il *IntegrationLayer) handleOpenAI(c *gin.Context) {
	// For OpenAI compatibility, we need to handle:
	// - /v1/chat/completions -> chat
	// - /v1/completions -> generate
	// - /v1/embeddings -> embed
	// - /v1/models -> tags
	
	path := c.Request.URL.Path
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
		il.handleOpenAIChat(c)
	case strings.HasSuffix(path, "/completions"):
		il.handleOpenAICompletion(c)
	case strings.HasSuffix(path, "/embeddings"):
		il.handleOpenAIEmbeddings(c)
	case strings.HasSuffix(path, "/models"):
		il.handleOpenAIModels(c)
	default:
		il.proxyToLocal(c)
	}
}

// Helper methods

func (il *IntegrationLayer) shouldDistribute(model string) bool {
	return il.distributedMode && il.modelDistribution.ShouldDistribute(model)
}

func (il *IntegrationLayer) getMode() string {
	if il.distributedMode {
		if il.fallbackMode {
			return "distributed-with-fallback"
		}
		return "distributed"
	}
	return "local"
}

func (il *IntegrationLayer) proxyToLocal(c *gin.Context) {
	c.Header("X-Ollama-Proxy", "local")
	il.localProxy.ServeHTTP(c.Writer, c.Request)
}

// Request tracking methods

func (rt *RequestTracker) TrackRequest(req *scheduler.Request) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	rt.requests[req.ID] = &TrackedRequest{
		ID:       req.ID,
		Started:  time.Now(),
		Model:    req.ModelName,
		Retries:  0,
		Response: req.ResponseCh,
	}
}

func (rt *RequestTracker) UntrackRequest(id string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	
	delete(rt.requests, id)
}

func (rt *RequestTracker) GetActiveRequests() map[string]*TrackedRequest {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	
	result := make(map[string]*TrackedRequest)
	for k, v := range rt.requests {
		result[k] = v
	}
	return result
}

// Helper functions for type conversion
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		if i, ok := v.(int64); ok {
			return i
		}
		if f, ok := v.(float64); ok {
			return int64(f)
		}
		if i, ok := v.(int); ok {
			return int64(i)
		}
	}
	return 0
}

// Placeholder methods for implementation

func (il *IntegrationLayer) handleDistributedPull(c *gin.Context, req ollamaapi.PullRequest) {
	// First check if model already exists in distributed cluster
	if il.modelDistribution.IsDistributed(req.Model) {
		c.JSON(http.StatusOK, gin.H{"status": "Model already available in distributed cluster"})
		return
	}

	// Start distributed pull process
	pullCh := make(chan bool, 1)
	go func() {
		// Try to find model on network first
		if modelInfo := il.modelDistribution.GetModelInfo(req.Model); modelInfo != nil {
			// Model exists on network, download from peer
			if err := il.modelDistribution.DownloadFromPeer(req.Model, "default-peer"); err != nil {
				pullCh <- false
				return
			}
			pullCh <- true
			return
		}

		// Model not on network, pull locally and replicate
		if err := il.pullModelLocally(req); err != nil {
			pullCh <- false
			return
		}

		// Register model for distribution
		if err := il.modelDistribution.RegisterModel(req.Model, "/tmp/models/"+req.Model); err != nil {
			pullCh <- false
			return
		}

		pullCh <- true
	}()

	// Wait for pull to complete or timeout
	select {
	case success := <-pullCh:
		if success {
			c.Header("X-Ollama-Distributed-Pull", "true")
			c.JSON(http.StatusOK, gin.H{"status": "success", "model": req.Model})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pull model"})
		}
	case <-time.After(5 * time.Minute):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Pull timeout"})
	}
}

func (il *IntegrationLayer) getLocalModels() ([]ollamaapi.ListModelResponse, error) {
	// Create request to local Ollama instance
	req, err := http.NewRequest("GET", il.localURL.String()+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make request to local instance
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get local models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local Ollama returned status %d", resp.StatusCode)
	}

	// Parse response
	var listResponse ollamaapi.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return listResponse.Models, nil
}

func (il *IntegrationLayer) getLocalProcesses() ([]ollamaapi.ProcessModelResponse, error) {
	// Create request to local Ollama instance
	req, err := http.NewRequest("GET", il.localURL.String()+"/api/ps", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make request to local instance
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get local processes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("local Ollama returned status %d", resp.StatusCode)
	}

	// Parse response
	var processResponse ollamaapi.ProcessResponse
	if err := json.NewDecoder(resp.Body).Decode(&processResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return processResponse.Models, nil
}

func (il *IntegrationLayer) getDistributedProcesses() ([]ollamaapi.ProcessModelResponse, error) {
	// Get active requests from scheduler
	activeRequests := il.requestTracker.GetActiveRequests()

	// Convert to process responses
	var processes []ollamaapi.ProcessModelResponse
	for _, req := range activeRequests {
		process := ollamaapi.ProcessModelResponse{
			Model: req.Model,
			Name:  req.Model,
			ExpiresAt: req.Started.Add(30 * time.Minute), // Default expiry
			SizeVRAM: 0, // TODO: Get actual VRAM usage
			Size: 0,     // TODO: Get actual size
		}
		processes = append(processes, process)
	}

	// Also get processes from distributed nodes via scheduler
	if il.scheduler != nil {
		nodes := il.scheduler.GetAvailableNodes()
		for _, node := range nodes {
			// Query each node for its processes
			nodeProcesses, err := il.getNodeProcesses(node.ID)
			if err != nil {
				// Log error but continue with other nodes
				continue
			}
			processes = append(processes, nodeProcesses...)
		}
	}

	return processes, nil
}

func (il *IntegrationLayer) getLocalVersion() (map[string]interface{}, error) {
	// Create request to local Ollama instance
	req, err := http.NewRequest("GET", il.localURL.String()+"/api/version", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Make request to local instance
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"version": "unknown"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"version": "unknown"}, nil
	}

	// Parse response
	var versionResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&versionResponse); err != nil {
		return map[string]interface{}{"version": "unknown"}, nil
	}

	return versionResponse, nil
}

func (il *IntegrationLayer) handleOpenAIChat(c *gin.Context) {
	// Parse OpenAI chat completion request
	var openAIReq struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream"`
	}

	if err := c.ShouldBindJSON(&openAIReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to Ollama chat request
	ollamaReq := ollamaapi.ChatRequest{
		Model: openAIReq.Model,
		Stream: &openAIReq.Stream,
	}

	// Convert messages
	for _, msg := range openAIReq.Messages {
		ollamaReq.Messages = append(ollamaReq.Messages, ollamaapi.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Replace request body with Ollama format
	body, _ := json.Marshal(ollamaReq)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	c.Request.ContentLength = int64(len(body))

	// Handle as regular chat request
	il.handleChat(c)
}

func (il *IntegrationLayer) handleOpenAICompletion(c *gin.Context) {
	// Parse OpenAI completion request
	var openAIReq struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Stream bool   `json:"stream"`
	}

	if err := c.ShouldBindJSON(&openAIReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to Ollama generate request
	ollamaReq := ollamaapi.GenerateRequest{
		Model:  openAIReq.Model,
		Prompt: openAIReq.Prompt,
		Stream: &openAIReq.Stream,
	}

	// Replace request body with Ollama format
	body, _ := json.Marshal(ollamaReq)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	c.Request.ContentLength = int64(len(body))

	// Handle as regular generate request
	il.handleGenerate(c)
}

func (il *IntegrationLayer) handleOpenAIEmbeddings(c *gin.Context) {
	// Parse OpenAI embeddings request
	var openAIReq struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}

	if err := c.ShouldBindJSON(&openAIReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to Ollama embed request
	ollamaReq := ollamaapi.EmbedRequest{
		Model: openAIReq.Model,
		Input: openAIReq.Input,
	}

	// Replace request body with Ollama format
	body, _ := json.Marshal(ollamaReq)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	c.Request.ContentLength = int64(len(body))

	// Handle as regular embed request
	il.handleEmbed(c)
}

func (il *IntegrationLayer) handleOpenAIModels(c *gin.Context) {
	// Get models using existing tags handler logic
	localModels, err := il.getLocalModels()
	if err != nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	distributedModels := il.modelDistribution.GetDistributedModels()
	if distributedModels == nil && !il.fallbackMode {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get distributed models"})
		return
	}

	// Convert to OpenAI format
	var openAIModels []map[string]interface{}

	// Add local models
	for _, model := range localModels {
		openAIModels = append(openAIModels, map[string]interface{}{
			"id":      model.Name,
			"object":  "model",
			"created": model.ModifiedAt.Unix(),
			"owned_by": "ollama",
		})
	}

	// Add distributed models
	for _, model := range distributedModels {
		if modelMap, ok := model.(map[string]interface{}); ok {
			openAIModels = append(openAIModels, map[string]interface{}{
				"id":      getString(modelMap, "name"),
				"object":  "model",
				"created": time.Now().Unix(), // Use current time since we don't have ModifiedAt
				"owned_by": "ollama-distributed",
			})
		}
	}

	// Return OpenAI-compatible response
	c.Header("X-Ollama-OpenAI-Compatible", "true")
	c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   openAIModels,
	})
}

// SetDistributedMode enables/disables distributed mode
func (il *IntegrationLayer) SetDistributedMode(enabled bool) {
	il.distributedMode = enabled
}

// SetFallbackMode enables/disables fallback to local
func (il *IntegrationLayer) SetFallbackMode(enabled bool) {
	il.fallbackMode = enabled
}

// GetStats returns integration layer statistics
func (il *IntegrationLayer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"distributed_mode": il.distributedMode,
		"fallback_mode":    il.fallbackMode,
		"active_requests":  len(il.requestTracker.GetActiveRequests()),
		"local_url":        il.localURL.String(),
	}
}

// pullModelLocally pulls a model using the local Ollama instance
func (il *IntegrationLayer) pullModelLocally(req ollamaapi.PullRequest) error {
	// Create request to local Ollama instance
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", il.localURL.String()+"/api/pull", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Make request to local instance
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to pull model locally: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("local pull failed with status %d", resp.StatusCode)
	}

	return nil
}

// getNodeProcesses gets processes from a specific node
func (il *IntegrationLayer) getNodeProcesses(nodeID string) ([]ollamaapi.ProcessModelResponse, error) {
	// TODO: Implement P2P communication to get processes from specific node
	// For now, return empty slice
	return []ollamaapi.ProcessModelResponse{}, nil
}