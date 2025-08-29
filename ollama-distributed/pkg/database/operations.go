package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// User operations

// CreateUser creates a new user
func (m *Manager) CreateUser(ctx context.Context, user *User) (*User, error) {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	metadataJSON, _ := json.Marshal(user.Metadata)

	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, avatar_url, role, is_active, is_verified, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	err := m.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.FullName,
		user.AvatarURL, user.Role, user.IsActive, user.IsVerified,
		metadataJSON, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (m *Manager) GetUser(ctx context.Context, id string) (*User, error) {
	user := &User{}
	var metadataJSON []byte

	query := `
		SELECT id, username, email, password_hash, full_name, avatar_url, role, is_active, is_verified,
		       last_login, failed_login_attempts, locked_until, metadata, created_at, updated_at
		FROM users WHERE id = $1`

	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.AvatarURL, &user.Role, &user.IsActive, &user.IsVerified,
		&user.LastLogin, &user.FailedLoginAttempts, &user.LockedUntil,
		&metadataJSON, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (m *Manager) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	var metadataJSON []byte

	query := `
		SELECT id, username, email, password_hash, full_name, avatar_url, role, is_active, is_verified,
		       last_login, failed_login_attempts, locked_until, metadata, created_at, updated_at
		FROM users WHERE username = $1`

	err := m.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.AvatarURL, &user.Role, &user.IsActive, &user.IsVerified,
		&user.LastLogin, &user.FailedLoginAttempts, &user.LockedUntil,
		&metadataJSON, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &user.Metadata)
	}

	return user, nil
}

// Model operations

// CreateModel creates a new model
func (m *Manager) CreateModel(ctx context.Context, model *Model) (*Model, error) {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	configJSON, _ := json.Marshal(model.Config)
	metadataJSON, _ := json.Marshal(model.Metadata)

	query := `
		INSERT INTO models (id, name, version, family, size_bytes, parameters, format, quantization, content_hash, config, metadata, is_public, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	err := m.db.QueryRowContext(ctx, query,
		model.ID, model.Name, model.Version, model.Family, model.SizeBytes,
		model.Parameters, model.Format, model.Quantization, model.ContentHash,
		configJSON, metadataJSON, model.IsPublic, model.CreatedBy,
		model.CreatedAt, model.UpdatedAt,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	return model, nil
}

// GetModel retrieves a model by ID
func (m *Manager) GetModel(ctx context.Context, id string) (*Model, error) {
	model := &Model{}
	var configJSON, metadataJSON []byte

	query := `
		SELECT id, name, version, family, size_bytes, parameters, format, quantization, content_hash, config, metadata, is_public, created_by, created_at, updated_at
		FROM models WHERE id = $1`

	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID, &model.Name, &model.Version, &model.Family, &model.SizeBytes,
		&model.Parameters, &model.Format, &model.Quantization, &model.ContentHash,
		&configJSON, &metadataJSON, &model.IsPublic, &model.CreatedBy,
		&model.CreatedAt, &model.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("model not found")
		}
		return nil, fmt.Errorf("failed to get model: %w", err)
	}

	if len(configJSON) > 0 {
		json.Unmarshal(configJSON, &model.Config)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &model.Metadata)
	}

	return model, nil
}

// ListModels retrieves all models with pagination
func (m *Manager) ListModels(ctx context.Context, limit, offset int) ([]*Model, error) {
	query := `
		SELECT id, name, version, family, size_bytes, parameters, format, quantization, content_hash, config, metadata, is_public, created_by, created_at, updated_at
		FROM models ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := m.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer rows.Close()

	var models []*Model
	for rows.Next() {
		model := &Model{}
		var configJSON, metadataJSON []byte

		err := rows.Scan(
			&model.ID, &model.Name, &model.Version, &model.Family, &model.SizeBytes,
			&model.Parameters, &model.Format, &model.Quantization, &model.ContentHash,
			&configJSON, &metadataJSON, &model.IsPublic, &model.CreatedBy,
			&model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model: %w", err)
		}

		if len(configJSON) > 0 {
			json.Unmarshal(configJSON, &model.Config)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &model.Metadata)
		}

		models = append(models, model)
	}

	return models, nil
}

// Inference operations

// CreateInferenceRequest creates a new inference request
func (m *Manager) CreateInferenceRequest(ctx context.Context, req *InferenceRequest) (*InferenceRequest, error) {
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.CreatedAt = time.Now()

	parametersJSON, _ := json.Marshal(req.Parameters)

	query := `
		INSERT INTO inference_requests (id, user_id, model_id, node_id, request_type, prompt, parameters, status, priority, queue_position, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`

	err := m.db.QueryRowContext(ctx, query,
		req.ID, req.UserID, req.ModelID, req.NodeID, req.RequestType,
		req.Prompt, parametersJSON, req.Status, req.Priority,
		req.QueuePosition, req.CreatedAt,
	).Scan(&req.ID, &req.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create inference request: %w", err)
	}

	return req, nil
}

// Node operations

// CreateNode creates a new node
func (m *Manager) CreateNode(ctx context.Context, node *Node) (*Node, error) {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	capabilitiesJSON, _ := json.Marshal(node.Capabilities)
	resourcesJSON, _ := json.Marshal(node.Resources)
	metadataJSON, _ := json.Marshal(node.Metadata)

	query := `
		INSERT INTO nodes (id, name, address, port, status, role, region, zone, capabilities, resources, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	err := m.db.QueryRowContext(ctx, query,
		node.ID, node.Name, node.Address, node.Port, node.Status,
		node.Role, node.Region, node.Zone, capabilitiesJSON,
		resourcesJSON, metadataJSON, node.CreatedAt, node.UpdatedAt,
	).Scan(&node.ID, &node.CreatedAt, &node.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	return node, nil
}

// GetNode retrieves a node by ID
func (m *Manager) GetNode(ctx context.Context, id string) (*Node, error) {
	node := &Node{}
	var capabilitiesJSON, resourcesJSON, metadataJSON []byte

	query := `
		SELECT id, name, address, port, status, role, region, zone, capabilities, resources, metadata, last_seen, created_at, updated_at
		FROM nodes WHERE id = $1`

	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID, &node.Name, &node.Address, &node.Port, &node.Status,
		&node.Role, &node.Region, &node.Zone, &capabilitiesJSON,
		&resourcesJSON, &metadataJSON, &node.LastSeen,
		&node.CreatedAt, &node.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found")
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if len(capabilitiesJSON) > 0 {
		json.Unmarshal(capabilitiesJSON, &node.Capabilities)
	}
	if len(resourcesJSON) > 0 {
		json.Unmarshal(resourcesJSON, &node.Resources)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &node.Metadata)
	}

	return node, nil
}

// ListNodes retrieves all nodes
func (m *Manager) ListNodes(ctx context.Context) ([]*Node, error) {
	query := `
		SELECT id, name, address, port, status, role, region, zone, capabilities, resources, metadata, last_seen, created_at, updated_at
		FROM nodes ORDER BY created_at DESC`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		var capabilitiesJSON, resourcesJSON, metadataJSON []byte

		err := rows.Scan(
			&node.ID, &node.Name, &node.Address, &node.Port, &node.Status,
			&node.Role, &node.Region, &node.Zone, &capabilitiesJSON,
			&resourcesJSON, &metadataJSON, &node.LastSeen,
			&node.CreatedAt, &node.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}

		if len(capabilitiesJSON) > 0 {
			json.Unmarshal(capabilitiesJSON, &node.Capabilities)
		}
		if len(resourcesJSON) > 0 {
			json.Unmarshal(resourcesJSON, &node.Resources)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &node.Metadata)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// UpdateNodeStatus updates a node's status and last seen time
func (m *Manager) UpdateNodeStatus(ctx context.Context, nodeID, status string) error {
	query := `UPDATE nodes SET status = $1, last_seen = $2, updated_at = $3 WHERE id = $4`
	
	now := time.Now()
	_, err := m.db.ExecContext(ctx, query, status, now, now, nodeID)
	if err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	return nil
}
