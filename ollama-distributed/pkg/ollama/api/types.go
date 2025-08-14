package api

import (
	"time"
)

// GenerateRequest represents a request to generate text
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse represents a response from text generation
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

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Images  []byte `json:"images,omitempty"`
}

// ListResponse represents a list of models response
type ListResponse struct {
	Models []ModelResponse `json:"models"`
}

// ModelResponse represents information about a model
type ModelResponse struct {
	Name       string       `json:"name"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	ModifiedAt time.Time    `json:"modified_at"`
	Details    ModelDetails `json:"details,omitempty"`
}

// ModelDetails represents detailed information about a model
type ModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// PullRequest represents a request to pull a model
type PullRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

// PushRequest represents a request to push a model
type PushRequest struct {
	Name     string `json:"name"`
	Insecure bool   `json:"insecure,omitempty"`
	Stream   bool   `json:"stream,omitempty"`
}

// CreateRequest represents a request to create a model
type CreateRequest struct {
	Name      string `json:"name"`
	Modelfile string `json:"modelfile"`
	Stream    bool   `json:"stream,omitempty"`
}

// DeleteRequest represents a request to delete a model
type DeleteRequest struct {
	Name string `json:"name"`
}

// CopyRequest represents a request to copy a model
type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// ShowRequest represents a request to show model information
type ShowRequest struct {
	Name string `json:"name"`
}

// ShowResponse represents a response with model information
type ShowResponse struct {
	License    string       `json:"license"`
	Modelfile  string       `json:"modelfile"`
	Parameters string       `json:"parameters"`
	Template   string       `json:"template"`
	System     string       `json:"system"`
	Details    ModelDetails `json:"details"`
}

// EmbeddingRequest represents a request for embeddings
type EmbeddingRequest struct {
	Model  string   `json:"model"`
	Prompt []string `json:"prompt"`
}

// EmbeddingResponse represents a response with embeddings
type EmbeddingResponse struct {
	Embedding [][]float64 `json:"embedding"`
}

// ProgressResponse represents a progress response for long-running operations
type ProgressResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse represents a status response
type StatusResponse struct {
	Status string `json:"status"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// VersionResponse represents a version response
type VersionResponse struct {
	Version string `json:"version"`
}

// Options represents generation options
type Options map[string]interface{}

// Common option keys
const (
	OptionTemperature   = "temperature"
	OptionTopP          = "top_p"
	OptionTopK          = "top_k"
	OptionRepeatPenalty = "repeat_penalty"
	OptionSeed          = "seed"
	OptionNumPredict    = "num_predict"
	OptionNumCtx        = "num_ctx"
	OptionNumBatch      = "num_batch"
	OptionNumGQA        = "num_gqa"
	OptionNumGPU        = "num_gpu"
	OptionMainGPU       = "main_gpu"
	OptionLowVRAM       = "low_vram"
	OptionF16KV         = "f16_kv"
	OptionLogitsAll     = "logits_all"
	OptionVocabOnly     = "vocab_only"
	OptionUseMMap       = "use_mmap"
	OptionUseMlock      = "use_mlock"
	OptionEmbeddingOnly = "embedding_only"
	OptionRopeFreqBase  = "rope_frequency_base"
	OptionRopeFreqScale = "rope_frequency_scale"
	OptionNumThread     = "num_thread"
)

// Role constants for chat messages
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Status constants
const (
	StatusSuccess = "success"
	StatusError   = "error"
	StatusPending = "pending"
)

// Format constants
const (
	FormatJSON = "json"
)

// Helper functions

// NewGenerateRequest creates a new generate request
func NewGenerateRequest(model, prompt string) *GenerateRequest {
	return &GenerateRequest{
		Model:  model,
		Prompt: prompt,
	}
}

// NewChatRequest creates a new chat request
func NewChatRequest(model string, messages []Message) *ChatRequest {
	return &ChatRequest{
		Model:    model,
		Messages: messages,
	}
}

// NewMessage creates a new message
func NewMessage(role, content string) Message {
	return Message{
		Role:    role,
		Content: content,
	}
}

// WithOptions adds options to a generate request
func (r *GenerateRequest) WithOptions(options Options) *GenerateRequest {
	r.Options = options
	return r
}

// WithStream enables streaming for a generate request
func (r *GenerateRequest) WithStream(stream bool) *GenerateRequest {
	r.Stream = stream
	return r
}

// WithSystem sets the system prompt for a generate request
func (r *GenerateRequest) WithSystem(system string) *GenerateRequest {
	r.System = system
	return r
}

// WithContext sets the context for a generate request
func (r *GenerateRequest) WithContext(context []int) *GenerateRequest {
	r.Context = context
	return r
}

// WithOptions adds options to a chat request
func (r *ChatRequest) WithOptions(options Options) *ChatRequest {
	r.Options = options
	return r
}

// WithStream enables streaming for a chat request
func (r *ChatRequest) WithStream(stream bool) *ChatRequest {
	r.Stream = stream
	return r
}

// AddMessage adds a message to a chat request
func (r *ChatRequest) AddMessage(role, content string) *ChatRequest {
	r.Messages = append(r.Messages, NewMessage(role, content))
	return r
}

// IsComplete returns true if the response is complete
func (r *GenerateResponse) IsComplete() bool {
	return r.Done
}

// IsComplete returns true if the response is complete
func (r *ChatResponse) IsComplete() bool {
	return r.Done
}

// GetContent returns the content of a chat response
func (r *ChatResponse) GetContent() string {
	return r.Message.Content
}

// GetRole returns the role of a chat response
func (r *ChatResponse) GetRole() string {
	return r.Message.Role
}
