package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// Prometheus metrics for query-level instrumentation
var (
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		},
		[]string{"operation", "table"},
	)

	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries executed",
		},
		[]string{"operation", "table", "status"},
	)

	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "table"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type", "table"},
	)

	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 10), // 0.1ms to ~100ms
		},
		[]string{"operation", "cache_type", "table"},
	)
)

// recordQueryMetrics wraps a query execution with metrics recording
func recordQueryMetrics(operation, table string, queryFunc func() error) error {
	start := time.Now()
	err := queryFunc()
	duration := time.Since(start).Seconds()

	// Record duration histogram
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)

	// Record counter with status
	status := "success"
	if err != nil {
		status = "error"
	}
	dbQueriesTotal.WithLabelValues(operation, table, status).Inc()

	return err
}

// recordCacheOperation wraps a cache operation with metrics recording
func recordCacheOperation(operation, cacheType, table string, cacheFunc func() error) error {
	start := time.Now()
	err := cacheFunc()
	duration := time.Since(start).Seconds()

	// Record duration histogram
	cacheOperationDuration.WithLabelValues(operation, cacheType, table).Observe(duration)

	return err
}

// recordCacheHit increments cache hit counter
func recordCacheHit(cacheType, table string) {
	cacheHitsTotal.WithLabelValues(cacheType, table).Inc()
}

// recordCacheMiss increments cache miss counter
func recordCacheMiss(cacheType, table string) {
	cacheMissesTotal.WithLabelValues(cacheType, table).Inc()
}

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

	// Wrap query execution with metrics
	err := recordQueryMetrics("create", "models", func() error {
		_, err := r.db.NamedExecContext(ctx, query, model)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Cache the model in Redis for 1 hour with metrics
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", model.ID.String())
		modelJSON, err := json.Marshal(model)
		if err != nil {
			r.logger.Warn("Failed to marshal model for cache", "error", err, "model_id", model.ID)
		} else {
			recordCacheOperation("set", "redis", "models", func() error {
				return r.redis.Set(ctx, key, modelJSON, time.Hour).Err()
			})
		}
	}

	return nil
}

func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*Model, error) {
	// Try to get from cache first with metrics
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		var modelJSON []byte
		var cacheErr error

		recordCacheOperation("get", "redis", "models", func() error {
			result := r.redis.Get(ctx, key)
			modelJSON, cacheErr = result.Bytes()
			return cacheErr
		})

		if cacheErr == nil {
			// Cache hit
			recordCacheHit("redis", "models")
			var model Model
			if err := json.Unmarshal(modelJSON, &model); err == nil {
				return &model, nil
			}
		} else if cacheErr == redis.Nil {
			// Cache miss
			recordCacheMiss("redis", "models")
		}
	}

	var model Model
	query := `SELECT * FROM models WHERE id = $1`

	err := recordQueryMetrics("get", "models", func() error {
		return r.db.GetContext(ctx, &model, query, id)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// Cache the model with metrics
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		modelJSON, err := json.Marshal(model)
		if err == nil {
			recordCacheOperation("set", "redis", "models", func() error {
				return r.redis.Set(ctx, key, modelJSON, time.Hour).Err()
			})
		}
	}

	return &model, nil
}

func (r *ModelRepository) GetByName(ctx context.Context, name string) (*Model, error) {
	var model Model
	query := `SELECT * FROM models WHERE name = $1 ORDER BY created_at DESC LIMIT 1`

	err := recordQueryMetrics("get", "models", func() error {
		return r.db.GetContext(ctx, &model, query, name)
	})

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
	var rows *sqlx.Rows

	err := recordQueryMetrics("list", "models", func() error {
		var err error
		rows, err = r.db.NamedQueryContext(ctx, query, args)
		return err
	})

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

	var result sql.Result
	err := recordQueryMetrics("update", "models", func() error {
		var err error
		result, err = r.db.NamedExecContext(ctx, query, model)
		return err
	})

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

	// Invalidate cache with metrics
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", model.ID.String())
		recordCacheOperation("delete", "redis", "models", func() error {
			return r.redis.Del(ctx, key).Err()
		})
	}

	return nil
}

func (r *ModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM models WHERE id = $1`

	var result sql.Result
	err := recordQueryMetrics("delete", "models", func() error {
		var err error
		result, err = r.db.ExecContext(ctx, query, id)
		return err
	})

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

	// Invalidate cache with metrics
	if r.redis != nil {
		key := fmt.Sprintf("model:%s", id.String())
		recordCacheOperation("delete", "redis", "models", func() error {
			return r.redis.Del(ctx, key).Err()
		})
	}

	return nil
}

func (r *ModelRepository) GetReplicas(ctx context.Context, modelID uuid.UUID) ([]*ModelReplica, error) {
	var replicas []*ModelReplica
	query := `SELECT * FROM model_replicas WHERE model_id = $1 AND status = 'active' ORDER BY created_at`

	err := recordQueryMetrics("get", "model_replicas", func() error {
		return r.db.SelectContext(ctx, &replicas, query, modelID)
	})

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

	err = recordQueryMetrics("create", "users", func() error {
		_, err := r.db.NamedExecContext(ctx, query, user)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = $1`

	err := recordQueryMetrics("get", "users", func() error {
		return r.db.GetContext(ctx, &user, query, id)
	})

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

	err := recordQueryMetrics("get", "users", func() error {
		return r.db.GetContext(ctx, &user, query, username)
	})

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

	recordQueryMetrics("update", "users", func() error {
		_, err := r.db.ExecContext(ctx, query, userID)
		if err != nil {
			r.logger.Error("Failed to increment failed login attempts", "error", err, "user_id", userID)
		}
		return err
	})
}

func (r *UserRepository) resetFailedAttempts(ctx context.Context, userID uuid.UUID) {
	query := `UPDATE users SET failed_login_attempts = 0 WHERE id = $1`

	recordQueryMetrics("update", "users", func() error {
		_, err := r.db.ExecContext(ctx, query, userID)
		if err != nil {
			r.logger.Error("Failed to reset failed login attempts", "error", err, "user_id", userID)
		}
		return err
	})
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

	var result sql.Result
	err := recordQueryMetrics("update", "users", func() error {
		var err error
		result, err = r.db.NamedExecContext(ctx, query, user)
		return err
	})

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

	err := recordQueryMetrics("create", "audit_log_entries", func() error {
		_, err := r.db.NamedExecContext(ctx, query, entry)
		return err
	})

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
	var rows *sqlx.Rows

	err := recordQueryMetrics("list", "nodes", func() error {
		var err error
		rows, err = r.db.NamedQueryContext(ctx, query, args)
		return err
	})

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

func (r *NodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*Node, error) {
	var node Node
	query := `SELECT * FROM nodes WHERE id = $1`

	err := recordQueryMetrics("get", "nodes", func() error {
		return r.db.GetContext(ctx, &node, query, id)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found")
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return &node, nil
}

func (r *NodeRepository) Create(ctx context.Context, node *Node) error {
	node.ID = uuid.New()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	query := `
		INSERT INTO nodes (id, name, url, region, status, capabilities, metrics, metadata, created_at, updated_at)
		VALUES (:id, :name, :url, :region, :status, :capabilities, :metrics, :metadata, :created_at, :updated_at)`

	err := recordQueryMetrics("create", "nodes", func() error {
		_, err := r.db.NamedExecContext(ctx, query, node)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to create node: %w", err)
	}

	return nil
}

func (r *NodeRepository) Update(ctx context.Context, node *Node) error {
	node.UpdatedAt = time.Now()

	query := `
		UPDATE nodes
		SET name = :name, url = :url, region = :region, status = :status,
		    capabilities = :capabilities, metrics = :metrics, metadata = :metadata, updated_at = :updated_at
		WHERE id = :id`

	var result sql.Result
	err := recordQueryMetrics("update", "nodes", func() error {
		var err error
		result, err = r.db.NamedExecContext(ctx, query, node)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("node not found")
	}

	return nil
}

func (r *NodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM nodes WHERE id = $1`

	var result sql.Result
	err := recordQueryMetrics("delete", "nodes", func() error {
		var err error
		result, err = r.db.ExecContext(ctx, query, id)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("node not found")
	}

	return nil
}
