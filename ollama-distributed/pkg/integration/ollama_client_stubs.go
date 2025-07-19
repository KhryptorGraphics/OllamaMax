package integration

import (
	"context"
	"fmt"
	"net/http"
)

// Registry stub to replace github.com/ollama/ollama/server/internal/client/ollama.Registry
type Registry struct {
	// Stub implementation for compatibility
	models map[string]interface{}
}

// NewRegistry creates a new registry stub
func NewRegistry() *Registry {
	return &Registry{
		models: make(map[string]interface{}),
	}
}

// RegisterModel registers a model in the stub registry
func (r *Registry) RegisterModel(name string, model interface{}) {
	r.models[name] = model
}

// GetModel retrieves a model from the stub registry
func (r *Registry) GetModel(name string) (interface{}, bool) {
	model, exists := r.models[name]
	return model, exists
}

// ListModels returns all registered models
func (r *Registry) ListModels() map[string]interface{} {
	return r.models
}

// ClientInterface provides methods that would be available from ollama client
type ClientInterface interface {
	Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error)
	Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error)
	Embed(ctx context.Context, request EmbedRequest) (*EmbedResponse, error)
	List(ctx context.Context) (*ListResponse, error)
	Show(ctx context.Context, request ShowRequest) (*ShowResponse, error)
	Pull(ctx context.Context, request PullRequest) error
	Delete(ctx context.Context, request DeleteRequest) error
	Version(ctx context.Context) (*VersionResponse, error)
}

// Client stub implementation
type Client struct {
	baseURL string
	client  *http.Client
}

// NewClient creates a new ollama client stub
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// Request/Response types for ollama API

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream,omitempty"`
}

type GenerateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Model   string      `json:"model"`
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

type EmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

type ListResponse struct {
	Models []ModelInfo `json:"models"`
}

type ModelInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Digest     string `json:"digest"`
	ModifiedAt string `json:"modified_at"`
}

type ShowRequest struct {
	Name string `json:"name"`
}

type ShowResponse struct {
	License    string                 `json:"license"`
	Modelfile  string                 `json:"modelfile"`
	Parameters map[string]interface{} `json:"parameters"`
	Template   string                 `json:"template"`
	Details    map[string]interface{} `json:"details"`
}

type PullRequest struct {
	Name   string `json:"name"`
	Stream bool   `json:"stream,omitempty"`
}

type DeleteRequest struct {
	Name string `json:"name"`
}

type VersionResponse struct {
	Version string `json:"version"`
}

// Stub implementations of client methods

func (c *Client) Generate(ctx context.Context, request GenerateRequest) (*GenerateResponse, error) {
	// Stub implementation - would make HTTP request to ollama server
	return &GenerateResponse{
		Model:    request.Model,
		Response: fmt.Sprintf("Generated response for: %s", request.Prompt),
		Done:     true,
	}, nil
}

func (c *Client) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// Stub implementation
	lastMessage := request.Messages[len(request.Messages)-1]
	return &ChatResponse{
		Model: request.Model,
		Message: ChatMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("Chat response to: %s", lastMessage.Content),
		},
		Done: true,
	}, nil
}

func (c *Client) Embed(ctx context.Context, request EmbedRequest) (*EmbedResponse, error) {
	// Stub implementation - return dummy embedding
	embedding := make([]float64, 768) // Standard embedding size
	for i := range embedding {
		embedding[i] = 0.1 * float64(i%10)
	}
	
	return &EmbedResponse{
		Embedding: embedding,
	}, nil
}

func (c *Client) List(ctx context.Context) (*ListResponse, error) {
	// Stub implementation
	return &ListResponse{
		Models: []ModelInfo{
			{
				Name:       "llama2:7b",
				Size:       3825819519,
				Digest:     "abc123",
				ModifiedAt: "2024-01-01T00:00:00Z",
			},
		},
	}, nil
}

func (c *Client) Show(ctx context.Context, request ShowRequest) (*ShowResponse, error) {
	// Stub implementation
	return &ShowResponse{
		License:    "MIT",
		Modelfile:  fmt.Sprintf("FROM %s", request.Name),
		Parameters: map[string]interface{}{"temperature": 0.7},
		Template:   "{{ .System }} {{ .Prompt }}",
		Details:    map[string]interface{}{"family": "llama"},
	}, nil
}

func (c *Client) Pull(ctx context.Context, request PullRequest) error {
	// Stub implementation
	return nil
}

func (c *Client) Delete(ctx context.Context, request DeleteRequest) error {
	// Stub implementation
	return nil
}

func (c *Client) Version(ctx context.Context) (*VersionResponse, error) {
	// Stub implementation
	return &VersionResponse{
		Version: "0.1.0-distributed",
	}, nil
}

// Server interface stub to replace ollama server interface
type Server interface {
	GenerateRoutes(registry *Registry) (http.Handler, error)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ServerStub provides a stub implementation of the Server interface
type ServerStub struct {
	client *Client
}

// NewServerStub creates a new server stub
func NewServerStub(baseURL string) *ServerStub {
	return &ServerStub{
		client: NewClient(baseURL),
	}
}

func (s *ServerStub) GenerateRoutes(registry *Registry) (http.Handler, error) {
	// Stub implementation
	return http.DefaultServeMux, nil
}

func (s *ServerStub) Start(ctx context.Context) error {
	// Stub implementation
	return nil
}

func (s *ServerStub) Stop(ctx context.Context) error {
	// Stub implementation
	return nil
}

// Additional stub types and functions that might be needed

type ModelManifest struct {
	SchemaVersion int                    `json:"schemaVersion"`
	MediaType     string                 `json:"mediaType"`
	Config        ModelConfig            `json:"config"`
	Layers        []ModelLayer           `json:"layers"`
	Annotations   map[string]interface{} `json:"annotations,omitempty"`
}

type ModelConfig struct {
	MediaType string                 `json:"mediaType"`
	Digest    string                 `json:"digest"`
	Size      int64                  `json:"size"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type ModelLayer struct {
	MediaType   string                 `json:"mediaType"`
	Digest      string                 `json:"digest"`
	Size        int64                  `json:"size"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// Progress tracking for long-running operations
type ProgressCallback func(status string, current, total int64)

// Additional utility functions
func ParseModelName(name string) (registry, namespace, model, tag string) {
	// Stub implementation for model name parsing
	return "", "", name, "latest"
}

func ValidateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	return nil
}

// Health check function
func (c *Client) Health(ctx context.Context) error {
	// Stub implementation
	return nil
}

// Error types
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// Constants
const (
	DefaultBaseURL = "http://localhost:11434"
	DefaultTimeout = 30 // seconds
)