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
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

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

// NodeRepository handles node-related database operations
type NodeRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// InferenceRepository handles inference request database operations
type InferenceRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// ConfigRepository handles system configuration database operations
type ConfigRepository struct {
	db     *sqlx.DB
	redis  *redis.Client
	logger *slog.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, redis: redis, logger: logger}
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *SessionRepository {
	return &SessionRepository{db: db, redis: redis, logger: logger}
}

// NewNodeRepository creates a new node repository
func NewNodeRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *NodeRepository {
	return &NodeRepository{db: db, redis: redis, logger: logger}
}

// NewInferenceRepository creates a new inference repository
func NewInferenceRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *InferenceRepository {
	return &InferenceRepository{db: db, redis: redis, logger: logger}
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sqlx.DB, logger *slog.Logger) *AuditRepository {
	return &AuditRepository{db: db, logger: logger}
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *sqlx.DB, redis *redis.Client, logger *slog.Logger) *ConfigRepository {
	return &ConfigRepository{db: db, redis: redis, logger: logger}
}

// UserRepository methods

func (r *UserRepository) Create(ctx context.Context, user *User) error {
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	query := `
		INSERT INTO users (username, email, password_hash, roles, permissions, active, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err = r.db.QueryRowxContext(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Roles,
		user.Permissions, user.Active, user.Metadata).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.logger.Info("User created", "user_id", user.ID, "username", user.Username)
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE username = $1 AND active = true`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	query := `SELECT * FROM users WHERE id = $1 AND active = true`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ipAddress string) error {
	query := `
		UPDATE users 
		SET last_login_at = CURRENT_TIMESTAMP, last_login_ip = $1, failed_login_attempts = 0
		WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, ipAddress, userID)
	return err
}

func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1,
			locked_until = CASE 
				WHEN failed_login_attempts + 1 >= 5 THEN CURRENT_TIMESTAMP + INTERVAL '30 minutes'
				ELSE locked_until
			END
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// SessionRepository methods

func (r *SessionRepository) Create(ctx context.Context, session *UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, token_id, refresh_token_hash, expires_at, 
			refresh_expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, last_used_at`

	err := r.db.QueryRowxContext(ctx, query,
		session.UserID, session.TokenID, session.RefreshTokenHash,
		session.ExpiresAt, session.RefreshExpiresAt, session.IPAddress, session.UserAgent).
		Scan(&session.ID, &session.CreatedAt, &session.LastUsedAt)

	return err
}

func (r *SessionRepository) GetByTokenID(ctx context.Context, tokenID string) (*UserSession, error) {
	var session UserSession
	query := `SELECT * FROM user_sessions WHERE token_id = $1 AND revoked = false AND expires_at > CURRENT_TIMESTAMP`

	err := r.db.GetContext(ctx, &session, query, tokenID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) UpdateLastUsed(ctx context.Context, tokenID string) error {
	query := `UPDATE user_sessions SET last_used_at = CURRENT_TIMESTAMP WHERE token_id = $1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	return err
}

func (r *SessionRepository) Revoke(ctx context.Context, tokenID string) error {
	query := `UPDATE user_sessions SET revoked = true WHERE token_id = $1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	return err
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) (int, error) {
	var deletedCount int
	query := `SELECT cleanup_expired_sessions()`
	err := r.db.QueryRowContext(ctx, query).Scan(&deletedCount)
	return deletedCount, err
}

// NodeRepository methods

func (r *NodeRepository) Upsert(ctx context.Context, node *Node) error {
	if err := node.Validate(); err != nil {
		return fmt.Errorf("node validation failed: %w", err)
	}

	query := `
		INSERT INTO nodes (peer_id, name, region, zone, address, port, capabilities, 
			resources, status, last_heartbeat, version, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (peer_id) DO UPDATE SET
			name = EXCLUDED.name,
			region = EXCLUDED.region,
			zone = EXCLUDED.zone,
			address = EXCLUDED.address,
			port = EXCLUDED.port,
			capabilities = EXCLUDED.capabilities,
			resources = EXCLUDED.resources,
			status = EXCLUDED.status,
			last_heartbeat = EXCLUDED.last_heartbeat,
			version = EXCLUDED.version,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		node.PeerID, node.Name, node.Region, node.Zone, node.Address, node.Port,
		node.Capabilities, node.Resources, node.Status, node.LastHeartbeat,
		node.Version, node.Metadata).
		Scan(&node.ID, &node.CreatedAt, &node.UpdatedAt)

	return err
}

func (r *NodeRepository) GetByPeerID(ctx context.Context, peerID string) (*Node, error) {
	var node Node
	query := `SELECT * FROM nodes WHERE peer_id = $1`

	err := r.db.GetContext(ctx, &node, query, peerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found: %s", peerID)
		}
		return nil, err
	}

	return &node, nil
}

func (r *NodeRepository) List(ctx context.Context, filters *NodeFilters) ([]*Node, error) {
	if filters == nil {
		filters = &NodeFilters{Limit: 100}
	}

	query := `
		SELECT n.*, 
			CASE 
				WHEN n.last_heartbeat > NOW() - INTERVAL '5 minutes' THEN 'healthy'
				WHEN n.last_heartbeat > NOW() - INTERVAL '15 minutes' THEN 'degraded'
				ELSE 'unhealthy'
			END as health_status,
			COUNT(mr.id) as model_count,
			COUNT(CASE WHEN mr.status = 'ready' THEN 1 END) as ready_models
		FROM nodes n
		LEFT JOIN model_replicas mr ON n.id = mr.node_id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if filters.Region != nil {
		query += fmt.Sprintf(" AND n.region = $%d", argIndex)
		args = append(args, *filters.Region)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND n.status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.HealthyOnly {
		query += " AND n.last_heartbeat > NOW() - INTERVAL '5 minutes'"
	}

	query += " GROUP BY n.id, n.peer_id, n.name, n.region, n.zone, n.address, n.port, n.capabilities, n.resources, n.status, n.last_heartbeat, n.version, n.metadata, n.created_at, n.updated_at"
	query += " ORDER BY n.last_heartbeat DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.PeerID, &node.Name, &node.Region, &node.Zone,
			&node.Address, &node.Port, &node.Capabilities, &node.Resources,
			&node.Status, &node.LastHeartbeat, &node.Version, &node.Metadata,
			&node.CreatedAt, &node.UpdatedAt, &node.HealthStatus,
			&node.ModelCount, &node.ReadyModels,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, nil
}

func (r *NodeRepository) UpdateHeartbeat(ctx context.Context, peerID string) error {
	query := `UPDATE nodes SET last_heartbeat = CURRENT_TIMESTAMP WHERE peer_id = $1`
	_, err := r.db.ExecContext(ctx, query, peerID)
	return err
}

// InferenceRepository methods

func (r *InferenceRepository) Create(ctx context.Context, request *InferenceRequest) error {
	query := `
		INSERT INTO inference_requests (request_id, user_id, model_id, model_name, 
			prompt_hash, prompt_length, nodes_used, partition_strategy, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := r.db.QueryRowxContext(ctx, query,
		request.RequestID, request.UserID, request.ModelID, request.ModelName,
		request.PromptHash, request.PromptLength, request.NodesUsed,
		request.PartitionStrategy, request.Status, request.Metadata).
		Scan(&request.ID, &request.CreatedAt)

	return err
}

func (r *InferenceRepository) UpdateStatus(ctx context.Context, requestID string, status string, errorMessage *string) error {
	query := `
		UPDATE inference_requests 
		SET status = $1, error_message = $2,
			started_at = CASE WHEN status = 'processing' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,
			completed_at = CASE WHEN $1 IN ('completed', 'failed', 'cancelled') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE request_id = $3`

	_, err := r.db.ExecContext(ctx, query, status, errorMessage, requestID)
	return err
}

func (r *InferenceRepository) UpdateMetrics(ctx context.Context, requestID string, tokens, responseLength, executionTimeMs int) error {
	query := `
		UPDATE inference_requests 
		SET tokens_processed = $1, response_length = $2, execution_time_ms = $3,
			total_time_ms = EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - created_at))::INTEGER * 1000
		WHERE request_id = $4`

	_, err := r.db.ExecContext(ctx, query, tokens, responseLength, executionTimeMs, requestID)
	return err
}

// ConfigRepository methods

func (r *ConfigRepository) Get(ctx context.Context, key string) (*SystemConfig, error) {
	// Try cache first
	cachedValue, err := r.redis.Get(ctx, fmt.Sprintf("config:%s", key)).Result()
	if err == nil {
		var config SystemConfig
		if err := json.Unmarshal([]byte(cachedValue), &config); err == nil {
			return &config, nil
		}
	}

	var config SystemConfig
	query := `SELECT * FROM system_config WHERE key = $1`
	err = r.db.GetContext(ctx, &config, query, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config not found: %s", key)
		}
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(config); err == nil {
		r.redis.Set(ctx, fmt.Sprintf("config:%s", key), data, 5*time.Minute)
	}

	return &config, nil
}

func (r *ConfigRepository) Set(ctx context.Context, key string, value interface{}, description *string, userID *uuid.UUID) error {
	query := `
		INSERT INTO system_config (key, value, description, updated_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			description = COALESCE(EXCLUDED.description, system_config.description),
			updated_by = EXCLUDED.updated_by,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at`

	var config SystemConfig
	err := r.db.QueryRowxContext(ctx, query, key, value, description, userID).
		Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		return err
	}

	// Invalidate cache
	r.redis.Del(ctx, fmt.Sprintf("config:%s", key))

	return nil
}

// AuditRepository methods

func (r *AuditRepository) Log(ctx context.Context, entry *AuditLogEntry) error {
	query := `
		INSERT INTO audit_log (table_name, operation, row_id, old_values, new_values, 
			user_id, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, timestamp`

	err := r.db.QueryRowxContext(ctx, query,
		entry.TableName, entry.Operation, entry.RowID, entry.OldValues,
		entry.NewValues, entry.UserID, entry.IPAddress, entry.UserAgent).
		Scan(&entry.ID, &entry.Timestamp)

	return err
}

func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]*AuditLogEntry, error) {
	query := `
		SELECT * FROM audit_log 
		ORDER BY timestamp DESC 
		LIMIT $1 OFFSET $2`

	var entries []*AuditLogEntry
	err := r.db.SelectContext(ctx, &entries, query, limit, offset)
	return entries, err
}

func (r *AuditRepository) Cleanup(ctx context.Context, daysToKeep int) (int, error) {
	var deletedCount int
	query := `SELECT cleanup_old_audit_logs($1)`
	err := r.db.QueryRowContext(ctx, query, daysToKeep).Scan(&deletedCount)
	return deletedCount, err
}