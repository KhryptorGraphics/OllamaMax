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
)

// ModelRepository handles model-related database operations
type ModelRepository struct {
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

// Create creates a new model
func (r *ModelRepository) Create(ctx context.Context, model *Model) error {
	if err := model.Validate(); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	query := `
		INSERT INTO models (name, version, size, hash, content_type, description, tags, parameters, 
			model_file_path, quantization_level, parameter_size, family, created_by, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		model.Name, model.Version, model.Size, model.Hash, model.ContentType,
		model.Description, model.Tags, model.Parameters, model.ModelFilePath,
		model.QuantizationLevel, model.ParameterSize, model.Family,
		model.CreatedBy, model.Status).
		Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Invalidate cache
	r.invalidateCache(model.Name)

	r.logger.Info("Model created",
		"model_id", model.ID,
		"name", model.Name,
		"size", model.Size)

	return nil
}

// GetByName retrieves a model by name
func (r *ModelRepository) GetByName(ctx context.Context, name string) (*Model, error) {
	// Try cache first
	if cached, err := r.getFromCache(ctx, name); err == nil && cached != nil {
		return cached, nil
	}

	var model Model
	query := `SELECT * FROM models WHERE name = $1 AND status = 'active'`

	err := r.db.GetContext(ctx, &model, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// Get replica information
	if err := r.loadReplicas(ctx, &model); err != nil {
		r.logger.Warn("Failed to load replica information", "model", name, "error", err)
	}

	// Cache the result
	r.cacheModel(ctx, &model)

	return &model, nil
}

// GetByID retrieves a model by ID
func (r *ModelRepository) GetByID(ctx context.Context, id uuid.UUID) (*Model, error) {
	var model Model
	query := `SELECT * FROM models WHERE id = $1 AND status = 'active'`

	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	// Get replica information
	if err := r.loadReplicas(ctx, &model); err != nil {
		r.logger.Warn("Failed to load replica information", "model_id", id, "error", err)
	}

	return &model, nil
}

// List retrieves models with optional filtering
func (r *ModelRepository) List(ctx context.Context, filters *ModelFilters) ([]*Model, error) {
	if filters == nil {
		filters = &ModelFilters{Limit: 100}
	}
	if filters.Limit == 0 {
		filters.Limit = 100
	}

	query := `
		SELECT m.*, 
			COUNT(mr.id) as replica_count,
			COUNT(CASE WHEN mr.status = 'ready' THEN 1 END) as ready_replicas,
			AVG(mr.health_score) as health_score
		FROM models m
		LEFT JOIN model_replicas mr ON m.id = mr.model_id
		WHERE m.status = 'active'`

	args := []interface{}{}
	argIndex := 1

	if filters.NameFilter != nil {
		query += fmt.Sprintf(" AND m.name ILIKE $%d", argIndex)
		args = append(args, "%"+*filters.NameFilter+"%")
		argIndex++
	}

	if filters.Tags != nil && len(filters.Tags) > 0 {
		query += fmt.Sprintf(" AND m.tags @> $%d", argIndex)
		args = append(args, filters.Tags)
		argIndex++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND m.status = $%d", argIndex)
		args = append(args, *filters.Status)
		argIndex++
	}

	if filters.Family != nil {
		query += fmt.Sprintf(" AND m.family = $%d", argIndex)
		args = append(args, *filters.Family)
		argIndex++
	}

	if filters.MinSize != nil {
		query += fmt.Sprintf(" AND m.size >= $%d", argIndex)
		args = append(args, *filters.MinSize)
		argIndex++
	}

	if filters.MaxSize != nil {
		query += fmt.Sprintf(" AND m.size <= $%d", argIndex)
		args = append(args, *filters.MaxSize)
		argIndex++
	}

	query += " GROUP BY m.id, m.name, m.version, m.size, m.hash, m.content_type, m.description, m.tags, m.parameters, m.model_file_path, m.quantization_level, m.parameter_size, m.family, m.created_at, m.updated_at, m.created_by, m.status"
	query += " ORDER BY m.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []*Model
	for rows.Next() {
		var model Model
		err := rows.Scan(
			&model.ID, &model.Name, &model.Version, &model.Size, &model.Hash,
			&model.ContentType, &model.Description, &model.Tags, &model.Parameters,
			&model.ModelFilePath, &model.QuantizationLevel, &model.ParameterSize,
			&model.Family, &model.CreatedAt, &model.UpdatedAt, &model.CreatedBy,
			&model.Status, &model.ReplicaCount, &model.ReadyReplicas, &model.HealthScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}
		models = append(models, &model)
	}

	return models, nil
}

// Update updates a model
func (r *ModelRepository) Update(ctx context.Context, model *Model) error {
	if err := model.Validate(); err != nil {
		return fmt.Errorf("model validation failed: %w", err)
	}

	query := `
		UPDATE models 
		SET version = $1, size = $2, hash = $3, content_type = $4, description = $5,
			tags = $6, parameters = $7, model_file_path = $8, quantization_level = $9,
			parameter_size = $10, family = $11, status = $12, updated_at = CURRENT_TIMESTAMP
		WHERE id = $13
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		model.Version, model.Size, model.Hash, model.ContentType, model.Description,
		model.Tags, model.Parameters, model.ModelFilePath, model.QuantizationLevel,
		model.ParameterSize, model.Family, model.Status, model.ID).
		Scan(&model.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	// Invalidate cache
	r.invalidateCache(model.Name)

	r.logger.Info("Model updated",
		"model_id", model.ID,
		"name", model.Name)

	return nil
}

// Delete soft deletes a model
func (r *ModelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE models SET status = 'deleted', updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("model not found: %s", id)
	}

	// Get model name for cache invalidation
	var name string
	r.db.GetContext(ctx, &name, "SELECT name FROM models WHERE id = $1", id)
	r.invalidateCache(name)

	r.logger.Info("Model deleted", "model_id", id)

	return nil
}

// GetReplicas retrieves all replicas for a model
func (r *ModelRepository) GetReplicas(ctx context.Context, modelID uuid.UUID) ([]*ModelReplica, error) {
	query := `
		SELECT mr.*, n.name as node_name, n.peer_id as node_peer_id
		FROM model_replicas mr
		JOIN nodes n ON mr.node_id = n.id
		WHERE mr.model_id = $1
		ORDER BY mr.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, modelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model replicas: %w", err)
	}
	defer rows.Close()

	var replicas []*ModelReplica
	for rows.Next() {
		var replica ModelReplica
		var nodeName, nodePeerID sql.NullString

		err := rows.Scan(
			&replica.ID, &replica.ModelID, &replica.NodeID, &replica.ReplicaPath,
			&replica.ReplicaHash, &replica.ReplicaSize, &replica.Status, &replica.HealthScore,
			&replica.LastVerified, &replica.SyncProgress, &replica.ErrorMessage,
			&replica.CreatedAt, &replica.UpdatedAt, &nodeName, &nodePeerID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replica: %w", err)
		}

		// Add minimal node information
		if nodeName.Valid && nodePeerID.Valid {
			replica.Node = &Node{
				ID:     replica.NodeID,
				PeerID: nodePeerID.String,
				Name:   &nodeName.String,
			}
		}

		replicas = append(replicas, &replica)
	}

	return replicas, nil
}

// GetReplicasByNode retrieves all replicas on a specific node
func (r *ModelRepository) GetReplicasByNode(ctx context.Context, nodeID uuid.UUID) ([]*ModelReplica, error) {
	query := `
		SELECT mr.*, m.name as model_name
		FROM model_replicas mr
		JOIN models m ON mr.model_id = m.id
		WHERE mr.node_id = $1 AND m.status = 'active'
		ORDER BY mr.created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get replicas by node: %w", err)
	}
	defer rows.Close()

	var replicas []*ModelReplica
	for rows.Next() {
		var replica ModelReplica
		var modelName string

		err := rows.Scan(
			&replica.ID, &replica.ModelID, &replica.NodeID, &replica.ReplicaPath,
			&replica.ReplicaHash, &replica.ReplicaSize, &replica.Status, &replica.HealthScore,
			&replica.LastVerified, &replica.SyncProgress, &replica.ErrorMessage,
			&replica.CreatedAt, &replica.UpdatedAt, &modelName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replica: %w", err)
		}

		// Add minimal model information
		replica.Model = &Model{
			ID:   replica.ModelID,
			Name: modelName,
		}

		replicas = append(replicas, &replica)
	}

	return replicas, nil
}

// CreateReplica creates a new model replica
func (r *ModelRepository) CreateReplica(ctx context.Context, replica *ModelReplica) error {
	query := `
		INSERT INTO model_replicas (model_id, node_id, replica_path, replica_hash, 
			replica_size, status, health_score, sync_progress)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		replica.ModelID, replica.NodeID, replica.ReplicaPath, replica.ReplicaHash,
		replica.ReplicaSize, replica.Status, replica.HealthScore, replica.SyncProgress).
		Scan(&replica.ID, &replica.CreatedAt, &replica.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create replica: %w", err)
	}

	r.logger.Info("Model replica created",
		"replica_id", replica.ID,
		"model_id", replica.ModelID,
		"node_id", replica.NodeID)

	return nil
}

// UpdateReplica updates a model replica
func (r *ModelRepository) UpdateReplica(ctx context.Context, replica *ModelReplica) error {
	query := `
		UPDATE model_replicas 
		SET replica_hash = $1, replica_size = $2, status = $3, health_score = $4,
			last_verified = $5, sync_progress = $6, error_message = $7,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		replica.ReplicaHash, replica.ReplicaSize, replica.Status, replica.HealthScore,
		replica.LastVerified, replica.SyncProgress, replica.ErrorMessage, replica.ID).
		Scan(&replica.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update replica: %w", err)
	}

	return nil
}

// DeleteReplica deletes a model replica
func (r *ModelRepository) DeleteReplica(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM model_replicas WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete replica: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("replica not found: %s", id)
	}

	r.logger.Info("Model replica deleted", "replica_id", id)

	return nil
}

// GetOptimalNodesForModel finds the best nodes for model placement
func (r *ModelRepository) GetOptimalNodesForModel(ctx context.Context, modelID uuid.UUID, requiredReplicas int) ([]*Node, error) {
	query := `SELECT * FROM find_optimal_nodes_for_model($1, $2)`

	rows, err := r.db.QueryContext(ctx, query, modelID, requiredReplicas)
	if err != nil {
		return nil, fmt.Errorf("failed to find optimal nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		var utilizationScore float64

		err := rows.Scan(
			&node.ID, &node.PeerID, &node.Name, &node.Region, &node.Zone, &utilizationScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		nodes = append(nodes, &node)
	}

	return nodes, nil
}

// GetModelHealthScore calculates the health score for a model
func (r *ModelRepository) GetModelHealthScore(ctx context.Context, modelID uuid.UUID) (float64, error) {
	var healthScore sql.NullFloat64
	query := `SELECT calculate_model_health_score($1)`

	err := r.db.QueryRowContext(ctx, query, modelID).Scan(&healthScore)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate health score: %w", err)
	}

	if !healthScore.Valid {
		return 0, nil
	}

	return healthScore.Float64, nil
}

// Private helper methods

func (r *ModelRepository) loadReplicas(ctx context.Context, model *Model) error {
	replicas, err := r.GetReplicas(ctx, model.ID)
	if err != nil {
		return err
	}

	model.Replicas = replicas
	model.ReplicaCount = len(replicas)

	// Count ready replicas and calculate health
	readyCount := 0
	var totalHealth float64
	for _, replica := range replicas {
		if replica.Status == "ready" {
			readyCount++
		}
		totalHealth += replica.HealthScore
	}

	model.ReadyReplicas = readyCount
	if len(replicas) > 0 {
		model.HealthScore = totalHealth / float64(len(replicas))
	}

	return nil
}

func (r *ModelRepository) cacheModel(ctx context.Context, model *Model) {
	key := fmt.Sprintf("model:%s", model.Name)
	data, err := json.Marshal(model)
	if err != nil {
		r.logger.Warn("Failed to marshal model for cache", "error", err)
		return
	}

	if err := r.redis.Set(ctx, key, data, 15*time.Minute).Err(); err != nil {
		r.logger.Warn("Failed to cache model", "error", err)
	}
}

func (r *ModelRepository) getFromCache(ctx context.Context, name string) (*Model, error) {
	key := fmt.Sprintf("model:%s", name)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var model Model
	if err := json.Unmarshal([]byte(data), &model); err != nil {
		return nil, err
	}

	return &model, nil
}

func (r *ModelRepository) invalidateCache(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("model:%s", name)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		r.logger.Warn("Failed to invalidate model cache", "name", name, "error", err)
	}
}