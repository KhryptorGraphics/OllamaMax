package types

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// API types compatible with Ollama API
// These types provide compatibility with the original Ollama API while supporting distributed operations
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Stream   *bool                  `json:"stream,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
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

type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   *bool                  `json:"stream,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ChatResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PullRequest struct {
	Name     string `json:"name"`
	Model    string `json:"model,omitempty"`
	Insecure bool   `json:"insecure,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Stream   *bool  `json:"stream,omitempty"`
}

type PullResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type PushRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Stream   *bool  `json:"stream,omitempty"`
}

type PushResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type ListResponse struct {
	Models []ModelResponse `json:"models"`
}

type ModelResponse struct {
	Model      string       `json:"model"`
	Name       string       `json:"name"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty"`
	SizeVRAM   int64        `json:"size_vram,omitempty"`
}

type ListModelResponse = ModelResponse

type ProgressResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

type Options map[string]interface{}

type Duration struct {
	Duration time.Duration
}

// Server types to replace ollama/server imports
// Note: Model type is defined in distributed_types.go to avoid duplication
type OllamaModel struct {
	Name     string
	Size     int64
	Digest   string
	Path     string
	Template string
	System   string
	Options  Options
}

type RunnerRef struct {
	Model    *Model
	Adapter  *Model
	Sequence int

	// Mock implementation for now
	mu sync.Mutex
}

// Client interface for distributed operations
type Client interface {
	Generate(ctx context.Context, req *GenerateRequest, fn func(GenerateResponse)) error
	Chat(ctx context.Context, req *ChatRequest, fn func(ChatResponse)) error
	Pull(ctx context.Context, req *PullRequest, fn func(PullResponse)) error
	Push(ctx context.Context, req *PushRequest, fn func(PushResponse)) error
	List(ctx context.Context) (*ListResponse, error)
	Show(ctx context.Context, req *ShowRequest) (*ShowResponse, error)
	Copy(ctx context.Context, req *CopyRequest) error
	Delete(ctx context.Context, req *DeleteRequest) error
	Heartbeat(ctx context.Context) error
}

type ShowRequest struct {
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

type ShowResponse struct {
	License    string       `json:"license,omitempty"`
	Modelfile  string       `json:"modelfile,omitempty"`
	Parameters string       `json:"parameters,omitempty"`
	Template   string       `json:"template,omitempty"`
	System     string       `json:"system,omitempty"`
	Details    ModelDetails `json:"details,omitempty"`
}

type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

type DeleteRequest struct {
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

type EmbedRequest struct {
	Model     string                 `json:"model"`
	Input     string                 `json:"input"`
	Inputs    []string               `json:"inputs,omitempty"`
	Truncate  bool                   `json:"truncate,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	KeepAlive *Duration              `json:"keep_alive,omitempty"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

type EmbeddingRequest struct {
	Model     string                 `json:"model"`
	Input     string                 `json:"input"`
	Prompt    string                 `json:"prompt,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
	KeepAlive *Duration              `json:"keep_alive,omitempty"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type ProcessModelResponse struct {
	Status    string     `json:"status"`
	Digest    string     `json:"digest,omitempty"`
	Total     int64      `json:"total,omitempty"`
	Completed int64      `json:"completed,omitempty"`
	Model     string     `json:"model,omitempty"`
	Name      string     `json:"name,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	SizeVRAM  int64      `json:"size_vram,omitempty"`
	Size      int64      `json:"size,omitempty"`
}

type ProcessResponse struct {
	Status    string                 `json:"status"`
	Digest    string                 `json:"digest,omitempty"`
	Total     int64                  `json:"total,omitempty"`
	Completed int64                  `json:"completed,omitempty"`
	Models    []ProcessModelResponse `json:"models,omitempty"`
}

// GPU and hardware types
type GpuInfo struct {
	ID      string `json:"id"`
	Library string `json:"library"`
	Name    string `json:"name"`
	Compute string `json:"compute"`
	Driver  string `json:"driver"`
	Memory  uint64 `json:"memory"`
	Free    uint64 `json:"free"`
}

type GpuInfoList []GpuInfo

// Quantization types
type QuantizationType string

const (
	QuantizationQ4_0 QuantizationType = "q4_0"
	QuantizationQ4_1 QuantizationType = "q4_1"
	QuantizationQ5_0 QuantizationType = "q5_0"
	QuantizationQ5_1 QuantizationType = "q5_1"
	QuantizationQ8_0 QuantizationType = "q8_0"
	QuantizationF16  QuantizationType = "f16"
	QuantizationF32  QuantizationType = "f32"
)

// Error types
type StatusError struct {
	StatusCode   int
	Status       string
	ErrorMessage string `json:"error"`
}

func (e StatusError) Error() string {
	switch {
	case e.Status != "" && e.ErrorMessage != "":
		return fmt.Sprintf("%s: %s", e.Status, e.ErrorMessage)
	case e.Status != "":
		return e.Status
	case e.ErrorMessage != "":
		return e.ErrorMessage
	default:
		return fmt.Sprintf("status %d", e.StatusCode)
	}
}

// Streaming response handler
type ResponseFunc func([]byte) error

// Blob and file types
type Blob struct {
	Digest string
	Size   int64
}

type LayerReader struct {
	Layer
	io.ReadCloser
}

type Layer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// Model name parsing
type Name struct {
	Host      string
	Namespace string
	Model     string
	Tag       string
	RawTag    string
	Build     string
}

func (n Name) String() string {
	var sb strings.Builder
	if n.Host != "" {
		sb.WriteString(n.Host)
		sb.WriteString("/")
	}
	if n.Namespace != "" {
		sb.WriteString(n.Namespace)
		sb.WriteString("/")
	}
	sb.WriteString(n.Model)
	if n.Tag != "" {
		sb.WriteString(":")
		sb.WriteString(n.Tag)
	}
	return sb.String()
}

func ParseName(s string) Name {
	// Simple implementation - in production this would be more robust
	parts := strings.Split(s, "/")
	name := Name{}

	if len(parts) > 1 {
		name.Host = parts[0]
		if len(parts) > 2 {
			name.Namespace = parts[1]
			name.Model = parts[2]
		} else {
			name.Model = parts[1]
		}
	} else {
		name.Model = parts[0]
	}

	// Handle tag
	if strings.Contains(name.Model, ":") {
		modelParts := strings.Split(name.Model, ":")
		name.Model = modelParts[0]
		name.Tag = modelParts[1]
		name.RawTag = name.Tag
	}

	return name
}

// Note: Distributed system types (Scheduler, Node, Task, etc.) are defined in distributed_types.go
// This file focuses on Ollama API compatibility types
