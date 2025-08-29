package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// GetMigrations returns all available migrations
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Initial schema",
			Up: `
				-- Enable UUID extension
				CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
				CREATE EXTENSION IF NOT EXISTS "pgcrypto";

				-- Users table
				CREATE TABLE users (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					username VARCHAR(50) UNIQUE NOT NULL,
					email VARCHAR(255) UNIQUE NOT NULL,
					password_hash TEXT NOT NULL,
					full_name VARCHAR(255),
					avatar_url TEXT,
					role VARCHAR(50) NOT NULL DEFAULT 'user',
					is_active BOOLEAN DEFAULT true,
					is_verified BOOLEAN DEFAULT false,
					last_login TIMESTAMP WITH TIME ZONE,
					failed_login_attempts INTEGER DEFAULT 0,
					locked_until TIMESTAMP WITH TIME ZONE,
					metadata JSONB DEFAULT '{}',
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					CONSTRAINT valid_role CHECK (role IN ('admin', 'operator', 'user', 'viewer'))
				);

				-- Nodes table
				CREATE TABLE nodes (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					name VARCHAR(255) NOT NULL,
					address VARCHAR(255) NOT NULL,
					port INTEGER NOT NULL,
					status VARCHAR(50) DEFAULT 'offline',
					role VARCHAR(50) DEFAULT 'worker',
					region VARCHAR(100),
					zone VARCHAR(100),
					capabilities JSONB DEFAULT '{}',
					resources JSONB DEFAULT '{}',
					metadata JSONB DEFAULT '{}',
					last_seen TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					CONSTRAINT valid_status CHECK (status IN ('online', 'offline', 'maintenance', 'error')),
					CONSTRAINT valid_role CHECK (role IN ('leader', 'worker', 'observer'))
				);

				-- AI models registry
				CREATE TABLE models (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					name VARCHAR(255) NOT NULL,
					version VARCHAR(50) NOT NULL,
					family VARCHAR(100),
					size_bytes BIGINT,
					parameters BIGINT,
					format VARCHAR(50),
					quantization VARCHAR(50),
					content_hash VARCHAR(255) UNIQUE,
					config JSONB DEFAULT '{}',
					metadata JSONB DEFAULT '{}',
					is_public BOOLEAN DEFAULT false,
					created_by UUID REFERENCES users(id),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(name, version)
				);

				-- Model replicas on nodes
				CREATE TABLE model_replicas (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
					node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
					status VARCHAR(50) DEFAULT 'downloading',
					path TEXT,
					size_bytes BIGINT,
					checksum VARCHAR(255),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					UNIQUE(model_id, node_id),
					CONSTRAINT valid_replica_status CHECK (status IN ('downloading', 'available', 'error', 'deleted'))
				);

				-- Inference requests
				CREATE TABLE inference_requests (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					user_id UUID REFERENCES users(id),
					model_id UUID REFERENCES models(id),
					node_id UUID REFERENCES nodes(id),
					request_type VARCHAR(50) NOT NULL,
					prompt TEXT,
					parameters JSONB DEFAULT '{}',
					status VARCHAR(50) DEFAULT 'queued',
					priority INTEGER DEFAULT 5,
					queue_position INTEGER,
					started_at TIMESTAMP WITH TIME ZONE,
					completed_at TIMESTAMP WITH TIME ZONE,
					error_message TEXT,
					tokens_input INTEGER,
					tokens_output INTEGER,
					latency_ms INTEGER,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					CONSTRAINT valid_request_type CHECK (request_type IN ('completion', 'chat', 'embedding', 'classification')),
					CONSTRAINT valid_request_status CHECK (status IN ('queued', 'processing', 'completed', 'failed', 'cancelled'))
				);

				-- Inference results
				CREATE TABLE inference_results (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					request_id UUID NOT NULL REFERENCES inference_requests(id) ON DELETE CASCADE,
					response TEXT,
					embeddings REAL[],
					metadata JSONB DEFAULT '{}',
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				-- API keys
				CREATE TABLE api_keys (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					name VARCHAR(255) NOT NULL,
					key_hash TEXT NOT NULL UNIQUE,
					permissions TEXT[] DEFAULT '{}',
					metadata JSONB DEFAULT '{}',
					last_used TIMESTAMP WITH TIME ZONE,
					expires_at TIMESTAMP WITH TIME ZONE,
					is_active BOOLEAN DEFAULT true,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				-- Sessions
				CREATE TABLE sessions (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					token TEXT NOT NULL UNIQUE,
					ip_address INET,
					user_agent TEXT,
					metadata JSONB DEFAULT '{}',
					expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				-- Audit logs
				CREATE TABLE audit_logs (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					user_id UUID REFERENCES users(id),
					action VARCHAR(100) NOT NULL,
					resource VARCHAR(100),
					details JSONB DEFAULT '{}',
					ip_address INET,
					user_agent TEXT,
					success BOOLEAN DEFAULT true,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);

				-- Create indexes for better performance
				CREATE INDEX idx_users_username ON users(username);
				CREATE INDEX idx_users_email ON users(email);
				CREATE INDEX idx_users_role ON users(role);
				CREATE INDEX idx_users_created ON users(created_at);

				CREATE INDEX idx_nodes_status ON nodes(status);
				CREATE INDEX idx_nodes_role ON nodes(role);
				CREATE INDEX idx_nodes_region ON nodes(region);
				CREATE INDEX idx_nodes_last_seen ON nodes(last_seen);

				CREATE INDEX idx_models_name ON models(name);
				CREATE INDEX idx_models_family ON models(family);
				CREATE INDEX idx_models_created ON models(created_at);
				CREATE INDEX idx_models_public ON models(is_public);

				CREATE INDEX idx_model_replicas_model ON model_replicas(model_id);
				CREATE INDEX idx_model_replicas_node ON model_replicas(node_id);
				CREATE INDEX idx_model_replicas_status ON model_replicas(status);

				CREATE INDEX idx_inference_user ON inference_requests(user_id);
				CREATE INDEX idx_inference_model ON inference_requests(model_id);
				CREATE INDEX idx_inference_node ON inference_requests(node_id);
				CREATE INDEX idx_inference_status ON inference_requests(status);
				CREATE INDEX idx_inference_created ON inference_requests(created_at);

				CREATE INDEX idx_results_request ON inference_results(request_id);

				CREATE INDEX idx_api_keys_user ON api_keys(user_id);
				CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
				CREATE INDEX idx_api_keys_active ON api_keys(is_active);

				CREATE INDEX idx_sessions_user ON sessions(user_id);
				CREATE INDEX idx_sessions_token ON sessions(token);
				CREATE INDEX idx_sessions_expires ON sessions(expires_at);

				CREATE INDEX idx_audit_user ON audit_logs(user_id);
				CREATE INDEX idx_audit_action ON audit_logs(action);
				CREATE INDEX idx_audit_created ON audit_logs(created_at);
			`,
			Down: `
				DROP TABLE IF EXISTS audit_logs;
				DROP TABLE IF EXISTS sessions;
				DROP TABLE IF EXISTS api_keys;
				DROP TABLE IF EXISTS inference_results;
				DROP TABLE IF EXISTS inference_requests;
				DROP TABLE IF EXISTS model_replicas;
				DROP TABLE IF EXISTS models;
				DROP TABLE IF EXISTS nodes;
				DROP TABLE IF EXISTS users;
				DROP EXTENSION IF EXISTS "pgcrypto";
				DROP EXTENSION IF EXISTS "uuid-ossp";
			`,
		},
		{
			Version:     2,
			Description: "Add migration tracking table",
			Up: `
				-- Migration tracking table
				CREATE TABLE schema_migrations (
					version INTEGER PRIMARY KEY,
					description TEXT NOT NULL,
					applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
				);
			`,
			Down: `
				DROP TABLE IF EXISTS schema_migrations;
			`,
		},
	}
}

// RunMigrations runs all pending migrations
func (m *Manager) RunMigrations(ctx context.Context) error {
	log.Info().Msg("Starting database migrations")

	// Ensure migration table exists
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	migrations := GetMigrations()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	for _, migration := range migrations {
		applied, err := m.isMigrationApplied(ctx, migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if applied {
			log.Debug().Int("version", migration.Version).Msg("Migration already applied")
			continue
		}

		log.Info().
			Int("version", migration.Version).
			Str("description", migration.Description).
			Msg("Applying migration")

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		log.Info().Int("version", migration.Version).Msg("Migration applied successfully")
	}

	log.Info().Msg("All migrations completed successfully")
	return nil
}

// ensureMigrationTable creates the migration tracking table if it doesn't exist
func (m *Manager) ensureMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// isMigrationApplied checks if a migration has been applied
func (m *Manager) isMigrationApplied(ctx context.Context, version int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`
	err := m.db.QueryRowContext(ctx, query, version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// applyMigration applies a single migration
func (m *Manager) applyMigration(ctx context.Context, migration Migration) error {
	return m.ExecuteInTransaction(ctx, func(tx *sql.Tx) error {
		// Execute the migration
		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}

		// Record the migration
		query := `INSERT INTO schema_migrations (version, description, applied_at) VALUES ($1, $2, $3)`
		if _, err := tx.ExecContext(ctx, query, migration.Version, migration.Description, time.Now()); err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		return nil
	})
}

// GetAppliedMigrations returns all applied migrations
func (m *Manager) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	query := `SELECT version, description FROM schema_migrations ORDER BY version`
	
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var migration Migration
		if err := rows.Scan(&migration.Version, &migration.Description); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}
