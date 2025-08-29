// 04-api-integration/comprehensive-api-client.go
// Comprehensive API client implementation for Ollama Distributed Training
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	_ "strconv"
	"strings"
	"time"
	"log"
	"os"
)

// OllamaDistributedClient provides a comprehensive client for the Ollama Distributed API
type OllamaDistributedClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	UserAgent  string
	Debug      bool
}

// Configuration for client initialization
type ClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	APIKey     string
	UserAgent  string
	Debug      bool
	MaxRetries int
	RetryDelay time.Duration
}

// Response structures matching the API

// Health Response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime,omitempty"`
	Version   string    `json:"version,omitempty"`
}

// Model structures
type Model struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Digest       string    `json:"digest,omitempty"`
	ModifiedAt   time.Time `json:"modified_at,omitempty"`
	Details      *ModelDetails `json:"details,omitempty"`
}

type ModelDetails struct {
	Format            string `json:"format"`
	Family            string `json:"family"`
	ParameterSize     string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

type ModelsResponse struct {
	Models []Model `json:"models"`
}

// Node structures
type Node struct {
	ID           string                 `json:"id"`
	Address      string                 `json:"address"`
	Status       string                 `json:"status"`
	LastSeen     time.Time             `json:"last_seen"`
	Capabilities NodeCapabilities      `json:"capabilities"`
	Metrics      NodeMetrics           `json:"metrics"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type NodeCapabilities struct {
	CPUCores        int      `json:"cpu_cores"`
	Memory          int64    `json:"memory"`
	Storage         int64    `json:"storage"`
	SupportedModels []string `json:"supported_models"`
	Available       bool     `json:"available"`
	LoadFactor      float64  `json:"load_factor"`
}

type NodeMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	DiskUsage   int64   `json:"disk_usage"`
	NetworkRx   int64   `json:"network_rx"`
	NetworkTx   int64   `json:"network_tx"`
}

type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}

// Cluster structures
type ClusterStatus struct {
	NodeID         string    `json:"node_id"`
	ClusterSize    int       `json:"cluster_size"`
	ConnectedPeers int       `json:"connected_peers"`
	Leader         string    `json:"leader"`
	IsLeader       bool      `json:"is_leader"`
	Health         string    `json:"health"`
	LastUpdate     time.Time `json:"last_update"`
	Nodes          []Node    `json:"nodes,omitempty"`
}

// Generation structures
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Format   string                 `json:"format,omitempty"`
}

type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Context  []int  `json:"context,omitempty"`
	
	// Performance metrics
	TotalDuration     int64 `json:"total_duration,omitempty"`
	LoadDuration      int64 `json:"load_duration,omitempty"`
	PromptEvalCount   int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount         int   `json:"eval_count,omitempty"`
	EvalDuration      int64 `json:"eval_duration,omitempty"`
}

// Chat structures
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ChatResponse struct {
	Model   string      `json:"model"`
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
	
	// Performance metrics (similar to GenerateResponse)
	TotalDuration     int64 `json:"total_duration,omitempty"`
	LoadDuration      int64 `json:"load_duration,omitempty"`
	PromptEvalCount   int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount         int   `json:"eval_count,omitempty"`
	EvalDuration      int64 `json:"eval_duration,omitempty"`
}

// Metrics structures
type SystemMetrics struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage int64     `json:"memory_usage"`
	DiskUsage   int64     `json:"disk_usage"`
	NetworkRx   int64     `json:"network_rx"`
	NetworkTx   int64     `json:"network_tx"`
	
	// Ollama-specific metrics
	ActiveModels    int     `json:"active_models"`
	RequestsPerSec  float64 `json:"requests_per_sec"`
	AvgResponseTime float64 `json:"avg_response_time"`
	ErrorRate       float64 `json:"error_rate"`
}

// Error response structure
type ErrorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// NewClient creates a new Ollama Distributed client
func NewClient(config ClientConfig) *OllamaDistributedClient {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	
	if config.UserAgent == "" {
		config.UserAgent = "ollama-distributed-client/1.0"
	}
	
	return &OllamaDistributedClient{
		BaseURL: strings.TrimRight(config.BaseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: config.Timeout,
		},
		APIKey:    config.APIKey,
		UserAgent: config.UserAgent,
		Debug:     config.Debug,
	}
}

// Helper methods for making HTTP requests

func (c *OllamaDistributedClient) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
		
		if c.Debug {
			log.Printf("Request body: %s", string(jsonData))
		}
	}
	
	url := c.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	
	if c.Debug {
		log.Printf("Making request: %s %s", method, url)
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	if c.Debug {
		log.Printf("Response status: %s", resp.Status)
	}
	
	return resp, nil
}

func (c *OllamaDistributedClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	if c.Debug {
		log.Printf("Response body: %s", string(body))
	}
	
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}
	
	return nil
}

// Health and Status Methods

func (c *OllamaDistributedClient) Health(ctx context.Context) (*HealthResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	
	var health HealthResponse
	if err := c.parseResponse(resp, &health); err != nil {
		return nil, err
	}
	
	return &health, nil
}

func (c *OllamaDistributedClient) GetClusterStatus(ctx context.Context) (*ClusterStatus, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/distributed/status", nil)
	if err != nil {
		return nil, err
	}
	
	var status ClusterStatus
	if err := c.parseResponse(resp, &status); err != nil {
		return nil, err
	}
	
	return &status, nil
}

func (c *OllamaDistributedClient) GetSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/distributed/metrics", nil)
	if err != nil {
		return nil, err
	}
	
	var metrics SystemMetrics
	if err := c.parseResponse(resp, &metrics); err != nil {
		return nil, err
	}
	
	return &metrics, nil
}

// Model Management Methods

func (c *OllamaDistributedClient) ListModels(ctx context.Context) (*ModelsResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/tags", nil)
	if err != nil {
		return nil, err
	}
	
	var models ModelsResponse
	if err := c.parseResponse(resp, &models); err != nil {
		return nil, err
	}
	
	return &models, nil
}

func (c *OllamaDistributedClient) PullModel(ctx context.Context, modelName string) error {
	body := map[string]string{"name": modelName}
	
	resp, err := c.makeRequest(ctx, "POST", "/api/pull", body)
	if err != nil {
		return err
	}
	
	return c.parseResponse(resp, nil)
}

func (c *OllamaDistributedClient) DeleteModel(ctx context.Context, modelName string) error {
	body := map[string]string{"name": modelName}
	
	resp, err := c.makeRequest(ctx, "DELETE", "/api/delete", body)
	if err != nil {
		return err
	}
	
	return c.parseResponse(resp, nil)
}

// Generation Methods

func (c *OllamaDistributedClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/generate", req)
	if err != nil {
		return nil, err
	}
	
	var genResp GenerateResponse
	if err := c.parseResponse(resp, &genResp); err != nil {
		return nil, err
	}
	
	return &genResp, nil
}

func (c *OllamaDistributedClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/chat", req)
	if err != nil {
		return nil, err
	}
	
	var chatResp ChatResponse
	if err := c.parseResponse(resp, &chatResp); err != nil {
		return nil, err
	}
	
	return &chatResp, nil
}

// Node Management Methods

func (c *OllamaDistributedClient) ListNodes(ctx context.Context) (*NodesResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/distributed/nodes", nil)
	if err != nil {
		return nil, err
	}
	
	var nodes NodesResponse
	if err := c.parseResponse(resp, &nodes); err != nil {
		return nil, err
	}
	
	return &nodes, nil
}

func (c *OllamaDistributedClient) GetNode(ctx context.Context, nodeID string) (*Node, error) {
	endpoint := "/api/distributed/nodes/" + url.PathEscape(nodeID)
	
	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	
	var node Node
	if err := c.parseResponse(resp, &node); err != nil {
		return nil, err
	}
	
	return &node, nil
}

func (c *OllamaDistributedClient) DrainNode(ctx context.Context, nodeID string) error {
	endpoint := "/api/distributed/nodes/" + url.PathEscape(nodeID) + "/drain"
	
	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	
	return c.parseResponse(resp, nil)
}

func (c *OllamaDistributedClient) UncordonNode(ctx context.Context, nodeID string) error {
	endpoint := "/api/distributed/nodes/" + url.PathEscape(nodeID) + "/uncordon"
	
	resp, err := c.makeRequest(ctx, "POST", endpoint, nil)
	if err != nil {
		return err
	}
	
	return c.parseResponse(resp, nil)
}

// Advanced Features

// StreamingCallback defines a callback function for streaming responses
type StreamingCallback func(chunk []byte) error

func (c *OllamaDistributedClient) GenerateStreaming(ctx context.Context, req *GenerateRequest, callback StreamingCallback) error {
	req.Stream = true
	
	resp, err := c.makeRequest(ctx, "POST", "/api/generate", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return c.parseResponse(resp, nil)
	}
	
	// Read streaming response
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if callbackErr := callback(buffer[:n]); callbackErr != nil {
				return callbackErr
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading stream: %w", err)
		}
	}
	
	return nil
}

// Utility Methods

func (c *OllamaDistributedClient) IsHealthy(ctx context.Context) bool {
	health, err := c.Health(ctx)
	return err == nil && health.Status == "healthy"
}

func (c *OllamaDistributedClient) WaitForHealthy(ctx context.Context, maxWait time.Duration, checkInterval time.Duration) error {
	deadline := time.Now().Add(maxWait)
	
	for time.Now().Before(deadline) {
		if c.IsHealthy(ctx) {
			return nil
		}
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(checkInterval):
			// Continue checking
		}
	}
	
	return fmt.Errorf("timeout waiting for service to become healthy")
}

func (c *OllamaDistributedClient) GetVersion(ctx context.Context) (string, error) {
	health, err := c.Health(ctx)
	if err != nil {
		return "", err
	}
	return health.Version, nil
}

// Example usage and testing functions

func ExampleBasicUsage() {
	// Create client
	config := ClientConfig{
		BaseURL: "http://localhost:8080",
		Timeout: 30 * time.Second,
		Debug:   true,
	}
	client := NewClient(config)
	
	ctx := context.Background()
	
	// Check health
	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	log.Printf("Service health: %s", health.Status)
	
	// List models
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Printf("Failed to list models: %v", err)
	} else {
		log.Printf("Found %d models", len(models.Models))
		for _, model := range models.Models {
			log.Printf("  - %s (%d bytes)", model.Name, model.Size)
		}
	}
	
	// Check cluster status
	cluster, err := client.GetClusterStatus(ctx)
	if err != nil {
		log.Printf("Failed to get cluster status: %v", err)
	} else {
		log.Printf("Cluster: %d nodes, %d connected peers", cluster.ClusterSize, cluster.ConnectedPeers)
	}
}

func ExampleGenerateText() {
	config := ClientConfig{
		BaseURL: "http://localhost:8080",
		Debug:   false,
	}
	client := NewClient(config)
	
	ctx := context.Background()
	
	// Generate text
	req := &GenerateRequest{
		Model:  "llama2:7b",
		Prompt: "Explain quantum computing in simple terms:",
		Options: map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  100,
		},
	}
	
	response, err := client.Generate(ctx, req)
	if err != nil {
		log.Printf("Generation failed: %v", err)
		return
	}
	
	log.Printf("Generated response: %s", response.Response)
	log.Printf("Total duration: %d ms", response.TotalDuration/1000000)
}

func ExampleStreamingGeneration() {
	config := ClientConfig{
		BaseURL: "http://localhost:8080",
		Debug:   false,
	}
	client := NewClient(config)
	
	ctx := context.Background()
	
	req := &GenerateRequest{
		Model:  "llama2:7b",
		Prompt: "Write a short story about artificial intelligence:",
	}
	
	log.Println("Starting streaming generation...")
	
	err := client.GenerateStreaming(ctx, req, func(chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	})
	
	if err != nil {
		log.Printf("Streaming failed: %v", err)
	} else {
		log.Println("\nStreaming completed successfully")
	}
}

func ExampleMonitoringDashboard() {
	config := ClientConfig{
		BaseURL: "http://localhost:8080",
		Timeout: 10 * time.Second,
		Debug:   false,
	}
	client := NewClient(config)
	
	ctx := context.Background()
	
	fmt.Println("Ollama Distributed Monitoring Dashboard")
	fmt.Println("======================================")
	
	// Health check
	if health, err := client.Health(ctx); err == nil {
		fmt.Printf("Health: %s (Uptime: %s)\n", health.Status, health.Uptime)
	} else {
		fmt.Printf("Health: ERROR - %v\n", err)
	}
	
	// Cluster status
	if cluster, err := client.GetClusterStatus(ctx); err == nil {
		fmt.Printf("Cluster: %d/%d nodes connected\n", cluster.ConnectedPeers, cluster.ClusterSize)
		fmt.Printf("Leader: %s (This node is leader: %v)\n", cluster.Leader, cluster.IsLeader)
	} else {
		fmt.Printf("Cluster: ERROR - %v\n", err)
	}
	
	// System metrics
	if metrics, err := client.GetSystemMetrics(ctx); err == nil {
		fmt.Printf("CPU: %.1f%% | Memory: %d MB | Disk: %d MB\n",
			metrics.CPUUsage, metrics.MemoryUsage/(1024*1024), metrics.DiskUsage/(1024*1024))
		fmt.Printf("Active Models: %d | Requests/sec: %.2f | Avg Response: %.2f ms\n",
			metrics.ActiveModels, metrics.RequestsPerSec, metrics.AvgResponseTime)
	} else {
		fmt.Printf("Metrics: ERROR - %v\n", err)
	}
	
	// Models
	if models, err := client.ListModels(ctx); err == nil {
		fmt.Printf("Models: %d loaded\n", len(models.Models))
		for _, model := range models.Models {
			sizeStr := fmt.Sprintf("%.1f MB", float64(model.Size)/(1024*1024))
			fmt.Printf("  - %s (%s)\n", model.Name, sizeStr)
		}
	} else {
		fmt.Printf("Models: ERROR - %v\n", err)
	}
	
	// Nodes
	if nodes, err := client.ListNodes(ctx); err == nil {
		fmt.Printf("Nodes: %d discovered\n", len(nodes.Nodes))
		for _, node := range nodes.Nodes {
			fmt.Printf("  - %s (%s) - Status: %s\n", node.ID, node.Address, node.Status)
		}
	} else {
		fmt.Printf("Nodes: ERROR - %v\n", err)
	}
}

// Command-line interface for testing
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run comprehensive-api-client.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  basic      - Run basic usage example")
		fmt.Println("  generate   - Run text generation example")
		fmt.Println("  stream     - Run streaming generation example")
		fmt.Println("  monitor    - Run monitoring dashboard")
		fmt.Println("  health     - Check service health")
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "basic":
		ExampleBasicUsage()
	case "generate":
		ExampleGenerateText()
	case "stream":
		ExampleStreamingGeneration()
	case "monitor":
		ExampleMonitoringDashboard()
	case "health":
		config := ClientConfig{
			BaseURL: "http://localhost:8080",
			Timeout: 5 * time.Second,
		}
		client := NewClient(config)
		
		ctx := context.Background()
		if client.IsHealthy(ctx) {
			fmt.Println("✅ Service is healthy")
		} else {
			fmt.Println("❌ Service is not healthy")
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}