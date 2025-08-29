package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Model represents a distributed AI model
type Model struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	Name               string     `db:"name" json:"name"`
	Version            string     `db:"version" json:"version"`
	Size               int64      `db:"size" json:"size"`
	Hash               string     `db:"hash" json:"hash"`
	ContentType        string     `db:"content_type" json:"content_type"`
	Description        *string    `db:"description" json:"description,omitempty"`
	Tags               JSONArray  `db:"tags" json:"tags"`
	Parameters         JSONMap    `db:"parameters" json:"parameters"`
	ModelFilePath      *string    `db:"model_file_path" json:"model_file_path,omitempty"`
	QuantizationLevel  *string    `db:"quantization_level" json:"quantization_level,omitempty"`
	ParameterSize      *string    `db:"parameter_size" json:"parameter_size,omitempty"`
	Family             *string    `db:"family" json:"family,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy          *string    `db:"created_by" json:"created_by,omitempty"`
	Status             string     `db:"status" json:"status"`

	// Computed fields (not stored in DB)
	ReplicaCount     int                `json:"replica_count,omitempty"`
	ReadyReplicas    int                `json:"ready_replicas,omitempty"`
	HealthScore      float64            `json:"health_score,omitempty"`
	Replicas         []*ModelReplica    `json:"replicas,omitempty"`
}

// Node represents a distributed system node
type Node struct {
	ID            uuid.UUID `db:"id" json:"id"`
	PeerID        string    `db:"peer_id" json:"peer_id"`
	Name          *string   `db:"name" json:"name,omitempty"`
	Region        *string   `db:"region" json:"region,omitempty"`
	Zone          *string   `db:"zone" json:"zone,omitempty"`
	Address       *string   `db:"address" json:"address,omitempty"`
	Port          *int      `db:"port" json:"port,omitempty"`
	Capabilities  JSONMap   `db:"capabilities" json:"capabilities"`
	Resources     JSONMap   `db:"resources" json:"resources"`
	Status        string    `db:"status" json:"status"`
	LastHeartbeat *time.Time `db:"last_heartbeat" json:"last_heartbeat,omitempty"`
	Version       *string   `db:"version" json:"version,omitempty"`
	Metadata      JSONMap   `db:"metadata" json:"metadata"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

	// Computed fields
	HealthStatus     string  `json:"health_status,omitempty"`
	ModelCount       int     `json:"model_count,omitempty"`
	ReadyModels      int     `json:"ready_models,omitempty"`
	Utilization      JSONMap `json:"utilization,omitempty"`
}

// ModelReplica represents a model replica on a specific node
type ModelReplica struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	ModelID       uuid.UUID  `db:"model_id" json:"model_id"`
	NodeID        uuid.UUID  `db:"node_id" json:"node_id"`
	ReplicaPath   string     `db:"replica_path" json:"replica_path"`
	ReplicaHash   *string    `db:"replica_hash" json:"replica_hash,omitempty"`
	ReplicaSize   *int64     `db:"replica_size" json:"replica_size,omitempty"`
	Status        string     `db:"status" json:"status"`
	HealthScore   float64    `db:"health_score" json:"health_score"`
	LastVerified  *time.Time `db:"last_verified" json:"last_verified,omitempty"`
	SyncProgress  float64    `db:"sync_progress" json:"sync_progress"`
	ErrorMessage  *string    `db:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`

	// Related objects
	Model *Model `json:"model,omitempty"`
	Node  *Node  `json:"node,omitempty"`
}

// User represents a system user
type User struct {
	ID                   uuid.UUID  `db:"id" json:"id"`
	Username             string     `db:"username" json:"username"`
	Email                *string    `db:"email" json:"email,omitempty"`
	PasswordHash         string     `db:"password_hash" json:"-"`
	Roles                StringArray `db:"roles" json:"roles"`
	Permissions          StringArray `db:"permissions" json:"permissions"`
	Active               bool       `db:"active" json:"active"`
	Metadata             JSONMap    `db:"metadata" json:"metadata"`
	LastLoginAt          *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	LastLoginIP          *string    `db:"last_login_ip" json:"last_login_ip,omitempty"`
	FailedLoginAttempts  int        `db:"failed_login_attempts" json:"failed_login_attempts"`
	LockedUntil          *time.Time `db:"locked_until" json:"locked_until,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// UserSession represents a user session/token
type UserSession struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	UserID             uuid.UUID  `db:"user_id" json:"user_id"`
	TokenID            string     `db:"token_id" json:"token_id"`
	RefreshTokenHash   *string    `db:"refresh_token_hash" json:"-"`
	ExpiresAt          time.Time  `db:"expires_at" json:"expires_at"`
	RefreshExpiresAt   *time.Time `db:"refresh_expires_at" json:"refresh_expires_at,omitempty"`
	IPAddress          *string    `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent          *string    `db:"user_agent" json:"user_agent,omitempty"`
	Revoked            bool       `db:"revoked" json:"revoked"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	LastUsedAt         time.Time  `db:"last_used_at" json:"last_used_at"`
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	ID               uuid.UUID   `db:"id" json:"id"`
	RequestID        string      `db:"request_id" json:"request_id"`
	UserID           *uuid.UUID  `db:"user_id" json:"user_id,omitempty"`
	ModelID          uuid.UUID   `db:"model_id" json:"model_id"`
	ModelName        string      `db:"model_name" json:"model_name"`
	PromptHash       *string     `db:"prompt_hash" json:"prompt_hash,omitempty"`
	PromptLength     *int        `db:"prompt_length" json:"prompt_length,omitempty"`
	ResponseLength   *int        `db:"response_length" json:"response_length,omitempty"`
	TokensProcessed  *int        `db:"tokens_processed" json:"tokens_processed,omitempty"`
	NodesUsed        StringArray `db:"nodes_used" json:"nodes_used"`
	PartitionStrategy *string    `db:"partition_strategy" json:"partition_strategy,omitempty"`
	ExecutionTimeMs  *int        `db:"execution_time_ms" json:"execution_time_ms,omitempty"`
	QueueTimeMs      *int        `db:"queue_time_ms" json:"queue_time_ms,omitempty"`
	TotalTimeMs      *int        `db:"total_time_ms" json:"total_time_ms,omitempty"`
	Status           string      `db:"status" json:"status"`
	ErrorMessage     *string     `db:"error_message" json:"error_message,omitempty"`
	Metadata         JSONMap     `db:"metadata" json:"metadata"`
	CreatedAt        time.Time   `db:"created_at" json:"created_at"`
	StartedAt        *time.Time  `db:"started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time  `db:"completed_at" json:"completed_at,omitempty"`
}

// SystemConfig represents system configuration
type SystemConfig struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Key         string     `db:"key" json:"key"`
	Value       JSONValue  `db:"value" json:"value"`
	Description *string    `db:"description" json:"description,omitempty"`
	Category    *string    `db:"category" json:"category,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	UpdatedBy   *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	TableName string     `db:"table_name" json:"table_name"`
	Operation string     `db:"operation" json:"operation"`
	RowID     *uuid.UUID `db:"row_id" json:"row_id,omitempty"`
	OldValues *JSONMap   `db:"old_values" json:"old_values,omitempty"`
	NewValues *JSONMap   `db:"new_values" json:"new_values,omitempty"`
	UserID    *uuid.UUID `db:"user_id" json:"user_id,omitempty"`
	IPAddress *string    `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent *string    `db:"user_agent" json:"user_agent,omitempty"`
	Timestamp time.Time  `db:"timestamp" json:"timestamp"`
}

// ModelUsageStats represents model usage statistics
type ModelUsageStats struct {
	ID                      uuid.UUID `db:"id" json:"id"`
	ModelID                 uuid.UUID `db:"model_id" json:"model_id"`
	Date                    time.Time `db:"date" json:"date"`
	RequestCount            int       `db:"request_count" json:"request_count"`
	TotalTokens             int       `db:"total_tokens" json:"total_tokens"`
	AverageResponseTimeMs   float64   `db:"average_response_time_ms" json:"average_response_time_ms"`
	UniqueUsers             int       `db:"unique_users" json:"unique_users"`
	SuccessRate             float64   `db:"success_rate" json:"success_rate"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
}

// Custom types for handling JSON and array fields in PostgreSQL

// JSONMap represents a JSON object stored as JSONB in PostgreSQL
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONArray represents a JSON array stored as JSONB in PostgreSQL
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArray, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONArray", value)
	}

	return json.Unmarshal(bytes, j)
}

// JSONValue represents any JSON value stored as JSONB in PostgreSQL
type JSONValue map[string]interface{}

func (j JSONValue) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONValue) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONValue", value)
	}

	return json.Unmarshal(bytes, j)
}

// StringArray represents a TEXT[] array in PostgreSQL
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	
	// PostgreSQL array format: {item1,item2,item3}
	result := "{"
	for i, item := range s {
		if i > 0 {
			result += ","
		}
		// Escape quotes and backslashes
		escaped := fmt.Sprintf("%q", item)
		result += escaped
	}
	result += "}"
	
	return result, nil
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = make(StringArray, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot scan %T into StringArray", value)
		}
		bytes = []byte(str)
	}

	// Simple PostgreSQL array parser
	str := string(bytes)
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("invalid PostgreSQL array format: %s", str)
	}

	if str == "{}" {
		*s = make(StringArray, 0)
		return nil
	}

	// Parse array elements (simple implementation)
	content := str[1 : len(str)-1]
	var result []string
	var current string
	inQuotes := false
	
	for i, r := range content {
		switch r {
		case '"':
			if i == 0 || content[i-1] != '\\' {
				inQuotes = !inQuotes
			} else {
				current += string(r)
			}
		case ',':
			if !inQuotes {
				result = append(result, current)
				current = ""
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}

	*s = StringArray(result)
	return nil
}

// Filter types for queries
type ModelFilters struct {
	NameFilter *string
	Tags       JSONArray
	Status     *string
	Family     *string
	MinSize    *int64
	MaxSize    *int64
	Limit      int
	Offset     int
}

type NodeFilters struct {
	Region      *string
	Zone        *string
	Status      *string
	MinModels   *int
	MaxModels   *int
	HealthyOnly bool
	Limit       int
	Offset      int
}

type InferenceFilters struct {
	UserID        *uuid.UUID
	ModelID       *uuid.UUID
	Status        *string
	FromDate      *time.Time
	ToDate        *time.Time
	MinDuration   *time.Duration
	MaxDuration   *time.Duration
	Limit         int
	Offset        int
}

// Validation methods

func (m *Model) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if m.Size <= 0 {
		return fmt.Errorf("model size must be positive")
	}
	if m.Hash == "" {
		return fmt.Errorf("model hash is required")
	}
	if m.Version == "" {
		return fmt.Errorf("model version is required")
	}
	return nil
}

func (n *Node) Validate() error {
	if n.PeerID == "" {
		return fmt.Errorf("node peer ID is required")
	}
	return nil
}

func (u *User) Validate() error {
	if u.Username == "" {
		return fmt.Errorf("username is required")
	}
	if u.PasswordHash == "" {
		return fmt.Errorf("password hash is required")
	}
	if len(u.Roles) == 0 {
		return fmt.Errorf("user must have at least one role")
	}
	return nil
}