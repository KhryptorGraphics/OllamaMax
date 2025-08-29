package database

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                  string                 `json:"id" db:"id"`
	Username            string                 `json:"username" db:"username"`
	Email               string                 `json:"email" db:"email"`
	PasswordHash        string                 `json:"-" db:"password_hash"`
	FullName            string                 `json:"full_name" db:"full_name"`
	AvatarURL           string                 `json:"avatar_url" db:"avatar_url"`
	Role                string                 `json:"role" db:"role"`
	IsActive            bool                   `json:"is_active" db:"is_active"`
	IsVerified          bool                   `json:"is_verified" db:"is_verified"`
	LastLogin           *time.Time             `json:"last_login" db:"last_login"`
	FailedLoginAttempts int                    `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time             `json:"locked_until" db:"locked_until"`
	Metadata            map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// Node represents a cluster node
type Node struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Address      string                 `json:"address" db:"address"`
	Port         int                    `json:"port" db:"port"`
	Status       string                 `json:"status" db:"status"`
	Role         string                 `json:"role" db:"role"`
	Region       string                 `json:"region" db:"region"`
	Zone         string                 `json:"zone" db:"zone"`
	Capabilities map[string]interface{} `json:"capabilities" db:"capabilities"`
	Resources    map[string]interface{} `json:"resources" db:"resources"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	LastSeen     *time.Time             `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Model represents an AI model
type Model struct {
	ID           string                 `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Version      string                 `json:"version" db:"version"`
	Family       string                 `json:"family" db:"family"`
	SizeBytes    int64                  `json:"size_bytes" db:"size_bytes"`
	Parameters   int64                  `json:"parameters" db:"parameters"`
	Format       string                 `json:"format" db:"format"`
	Quantization string                 `json:"quantization" db:"quantization"`
	ContentHash  string                 `json:"content_hash" db:"content_hash"`
	Config       map[string]interface{} `json:"config" db:"config"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	IsPublic     bool                   `json:"is_public" db:"is_public"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// ModelReplica represents a model replica on a node
type ModelReplica struct {
	ID        string    `json:"id" db:"id"`
	ModelID   string    `json:"model_id" db:"model_id"`
	NodeID    string    `json:"node_id" db:"node_id"`
	Status    string    `json:"status" db:"status"`
	Path      string    `json:"path" db:"path"`
	SizeBytes int64     `json:"size_bytes" db:"size_bytes"`
	Checksum  string    `json:"checksum" db:"checksum"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	ID           string                 `json:"id" db:"id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	ModelID      string                 `json:"model_id" db:"model_id"`
	NodeID       string                 `json:"node_id" db:"node_id"`
	RequestType  string                 `json:"request_type" db:"request_type"`
	Prompt       string                 `json:"prompt" db:"prompt"`
	Parameters   map[string]interface{} `json:"parameters" db:"parameters"`
	Status       string                 `json:"status" db:"status"`
	Priority     int                    `json:"priority" db:"priority"`
	QueuePosition int                   `json:"queue_position" db:"queue_position"`
	StartedAt    *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at" db:"completed_at"`
	ErrorMessage string                 `json:"error_message" db:"error_message"`
	TokensInput  int                    `json:"tokens_input" db:"tokens_input"`
	TokensOutput int                    `json:"tokens_output" db:"tokens_output"`
	LatencyMs    int                    `json:"latency_ms" db:"latency_ms"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// InferenceResult represents the result of an inference request
type InferenceResult struct {
	ID         string                 `json:"id" db:"id"`
	RequestID  string                 `json:"request_id" db:"request_id"`
	Response   string                 `json:"response" db:"response"`
	Embeddings []float32              `json:"embeddings" db:"embeddings"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID          string                 `json:"id" db:"id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Name        string                 `json:"name" db:"name"`
	KeyHash     string                 `json:"-" db:"key_hash"`
	Permissions []string               `json:"permissions" db:"permissions"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	LastUsed    *time.Time             `json:"last_used" db:"last_used"`
	ExpiresAt   *time.Time             `json:"expires_at" db:"expires_at"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Token     string                 `json:"-" db:"token"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	ExpiresAt time.Time              `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Action    string                 `json:"action" db:"action"`
	Resource  string                 `json:"resource" db:"resource"`
	Details   map[string]interface{} `json:"details" db:"details"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	Success   bool                   `json:"success" db:"success"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}
