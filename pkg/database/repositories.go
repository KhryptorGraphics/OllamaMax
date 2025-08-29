package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Repository interfaces for better testability and abstraction

// ModelRepository handles model-related database operations
type ModelRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// NodeRepository handles node-related database operations
type NodeRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// UserRepository handles user-related database operations
type UserRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// SessionRepository handles session-related database operations
type SessionRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// InferenceRepository handles inference-related database operations
type InferenceRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// AuditRepository handles audit log operations
type AuditRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// ConfigRepository handles system configuration operations
type ConfigRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// NewModelRepository creates a new model repository
func NewModelRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *ModelRepository {
	return &ModelRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// Model repository methods
func (r *ModelRepository) Create(ctx context.Context, model *Model) error {
	if err := model.Validate(); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	model.ID = uuid.New()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	query := `
		INSERT INTO models (id, name, version, size, hash, content_type, description, tags, parameters, 
		                   model_file_path, quantization_level, parameter_size, family, status, created_by, 
		                   created_at, updated_at)
		VALUES (:id, :name, :version, :size, :hash, :content_type, :description, :tags, :parameters, 
		        :model_file_path, :quantization_level, :parameter_size, :family, :status, :created_by, 
		        :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Cache the model in Redis for 1 hour
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", model.ID.String())
		if err := r.redis.Set(ctx, key, model, time.Hour).Err(); err != nil {
			r.logger.Warn("Failed to cache model in Redis", "error", err, "model_id", model.ID)
		}
	}

	return nil
}

func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*Model, error) {
	// Try to get from cache first
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		var model Model
		err := r.redis.Get(ctx, key).Scan(&model)
		if err == nil {
			return &model, nil
		}
	}

	var model Model
	query := `SELECT * FROM models WHERE id = $1`
	
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// Cache the model
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		if err := r.redis.Set(ctx, key, model, time.Hour).Err(); err != nil {
			r.logger.Warn("Failed to cache model in Redis", "error", err, "model_id", id)
		}
	}

	return &model, nil
}

func (r *ModelRepository) GetByName(ctx context.Context, name string) (*Model, error) {
	var model Model
	query := `SELECT * FROM models WHERE name = $1 ORDER BY created_at DESC LIMIT 1`
	
	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model by name: %w", err)
	}

	return &model, nil
}

func (r *ModelRepository) List(ctx context.Context, filters *ModelFilters) ([]*Model, error) {
	query := `SELECT * FROM models WHERE 1=1`
	args := make(map[string]interface{})

	if filters != nil {
		if filters.Status != nil {
			query += ` AND status = :status`
			args["status"] = *filters.Status
		}
		if filters.Family != nil {
			query += ` AND family = :family`
			args["family"] = *filters.Family
		}
		if filters.CreatedBy != nil {
			query += ` AND created_by = :created_by`
			args["created_by"] = *filters.CreatedBy
		}
		if len(filters.Tags) > 0 {
			query += ` AND tags && :tags`
			args["tags"] = filters.Tags
		}

		query += ` ORDER BY created_at DESC`

		if filters.Limit > 0 {
			query += ` LIMIT :limit`
			args["limit"] = filters.Limit
		}
		if filters.Offset > 0 {
			query += ` OFFSET :offset`
			args["offset"] = filters.Offset
		}
	}

	var models []*Model
	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var model Model
		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, &model)
	}

	return models, nil
}

func (r *ModelRepository) Update(ctx context.Context, model *Model) error {
	if err := model.Validate(); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	model.UpdatedAt = time.Now()

	query := `
		UPDATE models 
		SET name = :name, version = :version, description = :description, tags = :tags, 
		    parameters = :parameters, status = :status, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	// Invalidate cache
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", model.ID.String())
		if err := r.redis.Del(ctx, key).Err(); err != nil {
			r.logger.Warn("Failed to invalidate model cache", "error", err, "model_id", model.ID)
		}
	}

	return nil
}

func (r *ModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM models WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found")
	}

	// Invalidate cache
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		if err := r.redis.Del(ctx, key).Err(); err != nil {
			r.logger.Warn("Failed to invalidate model cache", "error", err, "model_id", id)
		}
	}

	return nil
}

func (r *ModelRepository) GetReplicas(ctx context.Context, modelID uuid.UUID) ([]*ModelReplica, error) {
	var replicas []*ModelReplica
	query := `SELECT * FROM model_replicas WHERE model_id = $1 AND status = 'active' ORDER BY created_at`
	
	err := r.db.SelectContext(ctx, &replicas, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model replicas: %w", err)
	}

	return replicas, nil
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// User repository methods
func (r *UserRepository) Create(ctx context.Context, user *User, password string) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.ID = uuid.New()
	user.PasswordHash = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password_hash, roles, permissions, active, metadata, 
		                  failed_login_attempts, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :roles, :permissions, :active, :metadata,
		        :failed_login_attempts, :created_at, :updated_at)`

	_, err = r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = $1`
	
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE username = $1`
	
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Authenticate(ctx context.Context, username, password string) (*User, error) {
	user, err := r.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Increment failed login attempts
		r.incrementFailedAttempts(ctx, user.ID)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login attempts on successful authentication
	r.resetFailedAttempts(ctx, user.ID)

	return user, nil
}

func (r *UserRepository) incrementFailedAttempts(ctx context.Context, userID uuid.UUID) {
	query := `UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to increment failed login attempts", "error", err, "user_id", userID)
	}
}

func (r *UserRepository) resetFailedAttempts(ctx context.Context, userID uuid.UUID) {
	query := `UPDATE users SET failed_login_attempts = 0 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to reset failed login attempts", "error", err, "user_id", userID)
	}
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET username = :username, email = :email, roles = :roles, permissions = :permissions,
		    active = :active, metadata = :metadata, updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// Session repository methods would go here...

// NewInferenceRepository creates a new inference repository
func NewInferenceRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *InferenceRepository {
	return &InferenceRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// Inference repository methods would go here...

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sqlx.DB, logger *slog.Logger) *AuditRepository {
	return &AuditRepository{
		db:     db,
		logger: logger,
	}
}

// Audit repository methods
func (r *AuditRepository) Create(ctx context.Context, entry *AuditLogEntry) error {
	entry.ID = uuid.New()
	entry.Timestamp = time.Now()

	query := `
		INSERT INTO audit_log_entries (id, table_name, operation, row_id, old_values, new_values,
		                              user_id, ip_address, user_agent, timestamp)
		VALUES (:id, :table_name, :operation, :row_id, :old_values, :new_values,
		        :user_id, :ip_address, :user_agent, :timestamp)`

	_, err := r.db.NamedExecContext(ctx, query, entry)
	if err != nil {
		return fmt.Errorf("failed to create audit log entry: %w", err)
	}

	return nil
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *ConfigRepository {
	return &ConfigRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// NewNodeRepository creates a new node repository
func NewNodeRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *NodeRepository {
	return &NodeRepository{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// NodeRepository methods
func (r *NodeRepository) List(ctx context.Context, filters *NodeFilters) ([]*Node, error) {
	query := `SELECT * FROM nodes WHERE 1=1`
	args := make(map[string]interface{})

	if filters != nil {
		if filters.Status != nil {
			query += ` AND status = :status`
			args["status"] = *filters.Status
		}
		if filters.Region != nil {
			query += ` AND region = :region`
			args["region"] = *filters.Region
		}
		if filters.HealthyOnly {
			query += ` AND status = 'healthy'`
		}

		query += ` ORDER BY created_at DESC`

		if filters.Limit > 0 {
			query += ` LIMIT :limit`
			args["limit"] = filters.Limit
		}
		if filters.Offset > 0 {
			query += ` OFFSET :offset`
			args["offset"] = filters.Offset
		}
	} else {
		query += ` ORDER BY created_at DESC LIMIT 50` // Default limit
	}

	var nodes []*Node
	rows, err := r.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var node Node
		if err := rows.StructScan(&node); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}