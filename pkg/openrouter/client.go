package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Client represents an OpenRouter API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	config     *Config
}

// Config holds OpenRouter client configuration
type Config struct {
	APIKey      string        `yaml:"api_key"`
	BaseURL     string        `yaml:"base_url"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxRetries  int           `yaml:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
	UserAgent   string        `yaml:"user_agent"`
	DebugMode   bool          `yaml:"debug_mode"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []Message      `json:"messages"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
	TopP        float64        `json:"top_p,omitempty"`
	TopK        int            `json:"top_k,omitempty"`
	Stream      bool           `json:"stream,omitempty"`
	PresencePenalty float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	Stop        []string       `json:"stop,omitempty"`
	Seed        int            `json:"seed,omitempty"`
	LogitBias   map[string]float64 `json:"logit_bias,omitempty"`
	Transforms  []string       `json:"transforms,omitempty"`
	Models      []string       `json:"models,omitempty"`
	Route       string         `json:"route,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Index        int      `json:"index"`
	Message      Message  `json:"message"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("OpenRouter API error: %s (type: %s)", e.Message, e.Type)
}

// NewClient creates a new OpenRouter client
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://openrouter.ai/api/v1"
	}
	
	if config.Timeout == 0 {
		config.Timeout = 300 * time.Second
	}
	
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}
	
	if config.UserAgent == "" {
		config.UserAgent = "OllamaMax/1.0.0"
	}
	
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}
	
	logger := logrus.New()
	if config.DebugMode {
		logger.SetLevel(logrus.DebugLevel)
	}
	
	return &Client{
		apiKey:     config.APIKey,
		baseURL:    config.BaseURL,
		httpClient: httpClient,
		logger:     logger,
		config:     config,
	}
}

// ChatCompletion sends a chat completion request to OpenRouter
func (c *Client) ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	c.logger.WithFields(logrus.Fields{
		"model": req.Model,
		"messages_count": len(req.Messages),
		"max_tokens": req.MaxTokens,
	}).Debug("Sending chat completion request")
	
	// Default to Sonoma Sky Alpha if no model specified
	if req.Model == "" {
		req.Model = "alpindale/sonoma-sky-alpha"
	}
	
	var response *ChatResponse
	var err error
	
	// Retry logic
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		response, err = c.doRequest(ctx, req)
		if err == nil {
			break
		}
		
		c.logger.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"max_retries": c.config.MaxRetries,
			"error": err,
		}).Warn("Request failed, retrying")
		
		if attempt < c.config.MaxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay * time.Duration(attempt+1)):
				// Exponential backoff
			}
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed after %d attempts: %w", c.config.MaxRetries+1, err)
	}
	
	c.logger.WithFields(logrus.Fields{
		"response_id": response.ID,
		"model": response.Model,
		"usage": response.Usage,
	}).Debug("Received chat completion response")
	
	return response, nil
}

// doRequest performs the actual HTTP request
func (c *Client) doRequest(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("User-Agent", c.config.UserAgent)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/khryptorgraphics/ollamamax")
	httpReq.Header.Set("X-Title", "OllamaMax Distributed AI Platform")
	
	if c.config.DebugMode {
		c.logger.WithFields(logrus.Fields{
			"url": httpReq.URL.String(),
			"headers": httpReq.Header,
			"body": string(body),
		}).Debug("Sending HTTP request")
	}
	
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if c.config.DebugMode {
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"headers": resp.Header,
			"body": string(respBody),
		}).Debug("Received HTTP response")
	}
	
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// Check for API errors
	if chatResp.Error != nil {
		return nil, chatResp.Error
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return &chatResp, nil
}

// GetModels retrieves available models from OpenRouter
func (c *Client) GetModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.config.UserAgent)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return modelsResp.Data, nil
}

// Model represents an OpenRouter model
type Model struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Pricing     ModelPricing      `json:"pricing"`
	ContextLength int             `json:"context_length"`
	Architecture ModelArchitecture `json:"architecture"`
	TopProvider  ModelProvider     `json:"top_provider"`
	PerRequestLimits map[string]int `json:"per_request_limits"`
}

// ModelPricing represents model pricing information
type ModelPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Image      string `json:"image,omitempty"`
	Request    string `json:"request,omitempty"`
}

// ModelArchitecture represents model architecture information
type ModelArchitecture struct {
	Modality    string `json:"modality"`
	Tokenizer   string `json:"tokenizer"`
	InstructType string `json:"instruct_type,omitempty"`
}

// ModelProvider represents model provider information
type ModelProvider struct {
	MaxCompletionTokens int    `json:"max_completion_tokens"`
	IsModerationEnabled bool   `json:"is_moderation_enabled"`
}

// ModelsResponse represents the models API response
type ModelsResponse struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

// ValidateRequest validates a chat request
func (c *Client) ValidateRequest(req *ChatRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}
	
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}
		
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return fmt.Errorf("message %d: invalid role %s", i, msg.Role)
		}
	}
	
	if req.Temperature < 0 || req.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	if req.TopP < 0 || req.TopP > 1 {
		return fmt.Errorf("top_p must be between 0 and 1")
	}
	
	if req.MaxTokens < 0 || req.MaxTokens > 131072 {
		return fmt.Errorf("max_tokens must be between 0 and 131072")
	}
	
	return nil
}

// Health checks the health of the OpenRouter API
func (c *Client) Health(ctx context.Context) error {
	req := &ChatRequest{
		Model: "alpindale/sonoma-sky-alpha",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 5,
	}
	
	_, err := c.ChatCompletion(ctx, req)
	return err
}
