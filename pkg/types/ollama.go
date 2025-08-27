package types

import (
	"time"
)

// OllamaModel represents an Ollama model
type OllamaModel struct {
	Name       string                 `json:"name"`
	Size       int64                  `json:"size"`
	Digest     string                 `json:"digest"`
	ModifiedAt time.Time              `json:"modified_at"`
	Details    OllamaModelDetails     `json:"details"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// OllamaModelDetails contains detailed information about a model
type OllamaModelDetails struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// ModelInfo represents basic model information
type ModelInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	Digest       string    `json:"digest"`
	ModifiedAt   time.Time `json:"modified_at"`
	Status       string    `json:"status"`
	DownloadedAt time.Time `json:"downloaded_at"`
}

// ModelStatus represents the status of a model
type ModelStatus string

const (
	ModelStatusAvailable   ModelStatus = "available"
	ModelStatusDownloading ModelStatus = "downloading"
	ModelStatusCorrupted   ModelStatus = "corrupted"
	ModelStatusMissing     ModelStatus = "missing"
)

// GenerateRequest represents a request for text generation
type GenerateRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream,omitempty"`
	Raw      bool                   `json:"raw,omitempty"`
	Format   string                 `json:"format,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	KeepAlive interface{}           `json:"keep_alive,omitempty"`
}

// GenerateResponse represents a response from text generation
type GenerateResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	
	// Metrics
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
}