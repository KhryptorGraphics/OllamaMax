package api

import (
	"time"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]interface{} `json:"services"`
}

// VersionResponse represents the version information response
type VersionResponse struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
}

// NodeInfo represents information about a cluster node
type NodeInfo struct {
	ID       string            `json:"id"`
	Address  string            `json:"address"`
	Status   string            `json:"status"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NodesResponse represents the response for listing cluster nodes
type NodesResponse struct {
	Nodes []NodeInfo `json:"nodes"`
	Total int        `json:"total"`
}

// NodeResponse represents the response for a single node
type NodeResponse struct {
	Node NodeInfo `json:"node"`
}

// InferenceRequest represents a request for model inference
type InferenceRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Template string                 `json:"template,omitempty"`
}

// InferenceResponse represents a response from model inference
type InferenceResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Config represents the API configuration for tests
type Config struct {
	Listen      string `json:"listen"`
	EnableCORS  bool   `json:"enable_cors"`
	RateLimit   int    `json:"rate_limit"`
	Timeout     int    `json:"timeout"`
	MaxBodySize int    `json:"max_body_size"`
}

// Note: GenerateRequest is defined in handlers.go to avoid duplication

// ChatMessage represents a single chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelInfo represents information about a model
type ModelInfo struct {
	Name       string            `json:"name"`
	Size       int64             `json:"size"`
	Digest     string            `json:"digest"`
	ModifiedAt time.Time         `json:"modified_at"`
	Details    map[string]string `json:"details,omitempty"`
}

// ModelsResponse represents the response for listing models
type ModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// PullRequest represents a request to pull a model
type PullRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

// PullResponse represents a response from pulling a model
type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// PushRequest represents a request to push a model
type PushRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

// PushResponse represents a response from pushing a model
type PushResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// DeleteRequest represents a request to delete a model
type DeleteRequest struct {
	Name string `json:"name"`
}

// DeleteResponse represents a response from deleting a model
type DeleteResponse struct {
	Status string `json:"status"`
}

// ShowRequest represents a request to show model information
type ShowRequest struct {
	Name string `json:"name"`
}

// ShowResponse represents a response with model information
type ShowResponse struct {
	License    string            `json:"license,omitempty"`
	Modelfile  string            `json:"modelfile,omitempty"`
	Parameters string            `json:"parameters,omitempty"`
	Template   string            `json:"template,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
}

// CopyRequest represents a request to copy a model
type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// CopyResponse represents a response from copying a model
type CopyResponse struct {
	Status string `json:"status"`
}

// CreateRequest represents a request to create a model
type CreateRequest struct {
	Name      string `json:"name"`
	Modelfile string `json:"modelfile"`
	Stream    bool   `json:"stream,omitempty"`
}

// CreateResponse represents a response from creating a model
type CreateResponse struct {
	Status string `json:"status"`
}

// EmbeddingsRequest represents a request for embeddings
type EmbeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingsResponse represents a response with embeddings
type EmbeddingsResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// TagsResponse represents the response for listing tags (alias for ModelsResponse)
type TagsResponse = ModelsResponse

// StatusResponse represents a generic status response
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// OllamaGenerateResponse represents a response from Ollama generate endpoint
type OllamaGenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// OllamaChatResponse represents a response from Ollama chat endpoint
type OllamaChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt time.Time   `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}
