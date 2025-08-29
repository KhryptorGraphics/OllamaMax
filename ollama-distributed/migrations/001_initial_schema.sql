-- OllamaMax Distributed AI Platform
-- Initial Database Schema Migration
-- Version: 1.0.0
-- Date: 2025-08-27

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create schema
CREATE SCHEMA IF NOT EXISTS ollamamax;
SET search_path TO ollamamax, public;

-- =====================================================
-- USERS AND AUTHENTICATION
-- =====================================================

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

-- Session management
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_token (token_hash),
    INDEX idx_sessions_expires (expires_at)
);

-- API keys for programmatic access
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    permissions JSONB DEFAULT '[]',
    rate_limit INTEGER DEFAULT 1000,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_api_keys_user_id (user_id),
    INDEX idx_api_keys_hash (key_hash)
);

-- =====================================================
-- CLUSTER AND NODES
-- =====================================================

-- Cluster configuration
CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    version VARCHAR(50),
    status VARCHAR(50) DEFAULT 'initializing',
    config JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_cluster_status CHECK (status IN ('initializing', 'healthy', 'degraded', 'maintenance', 'error'))
);

-- Nodes in the cluster
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cluster_id UUID REFERENCES clusters(id) ON DELETE CASCADE,
    node_id VARCHAR(255) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    port INTEGER NOT NULL,
    role VARCHAR(50) DEFAULT 'worker',
    status VARCHAR(50) DEFAULT 'offline',
    version VARCHAR(50),
    capabilities JSONB DEFAULT '{}',
    resources JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_node_role CHECK (role IN ('coordinator', 'worker', 'storage', 'compute')),
    CONSTRAINT valid_node_status CHECK (status IN ('online', 'offline', 'draining', 'maintenance', 'error')),
    INDEX idx_nodes_cluster (cluster_id),
    INDEX idx_nodes_status (status),
    INDEX idx_nodes_heartbeat (last_heartbeat)
);

-- Node health history
CREATE TABLE node_health_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    disk_usage DECIMAL(5,2),
    network_usage DECIMAL(10,2),
    temperature DECIMAL(5,2),
    metrics JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_health_node_id (node_id),
    INDEX idx_health_recorded (recorded_at)
);

-- =====================================================
-- MODELS AND VERSIONS
-- =====================================================

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
    UNIQUE(name, version),
    INDEX idx_models_name (name),
    INDEX idx_models_family (family),
    INDEX idx_models_created (created_at)
);

-- Model replicas across nodes
CREATE TABLE model_replicas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending',
    progress DECIMAL(5,2) DEFAULT 0,
    location TEXT,
    size_bytes BIGINT,
    last_verified TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_replica_status CHECK (status IN ('pending', 'downloading', 'ready', 'error', 'deleted')),
    UNIQUE(model_id, node_id),
    INDEX idx_replicas_model (model_id),
    INDEX idx_replicas_node (node_id),
    INDEX idx_replicas_status (status)
);

-- =====================================================
-- TASKS AND INFERENCE
-- =====================================================

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
    CONSTRAINT valid_request_status CHECK (status IN ('queued', 'processing', 'completed', 'failed', 'cancelled')),
    INDEX idx_inference_user (user_id),
    INDEX idx_inference_model (model_id),
    INDEX idx_inference_node (node_id),
    INDEX idx_inference_status (status),
    INDEX idx_inference_created (created_at)
);

-- Inference results
CREATE TABLE inference_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES inference_requests(id) ON DELETE CASCADE,
    response TEXT,
    embeddings VECTOR(1536),  -- For embedding storage (requires pgvector)
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_results_request (request_id)
);

-- =====================================================
-- TRANSFERS AND SYNCHRONIZATION
-- =====================================================

-- File transfers between nodes
CREATE TABLE transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_node_id UUID REFERENCES nodes(id),
    target_node_id UUID REFERENCES nodes(id),
    model_id UUID REFERENCES models(id),
    transfer_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    progress DECIMAL(5,2) DEFAULT 0,
    size_bytes BIGINT,
    speed_mbps DECIMAL(10,2),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_transfer_type CHECK (transfer_type IN ('model', 'checkpoint', 'dataset', 'config')),
    CONSTRAINT valid_transfer_status CHECK (status IN ('pending', 'active', 'paused', 'completed', 'failed', 'cancelled')),
    INDEX idx_transfers_source (source_node_id),
    INDEX idx_transfers_target (target_node_id),
    INDEX idx_transfers_status (status),
    INDEX idx_transfers_created (created_at)
);

-- =====================================================
-- AUDIT AND MONITORING
-- =====================================================

-- Audit log for all actions
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_created (created_at)
);

-- System metrics
CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_name VARCHAR(255) NOT NULL,
    metric_value DECIMAL(20,4),
    tags JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_metrics_name (metric_name),
    INDEX idx_metrics_recorded (recorded_at)
);

-- Alerts and notifications
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    source VARCHAR(100),
    entity_type VARCHAR(50),
    entity_id UUID,
    status VARCHAR(50) DEFAULT 'active',
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_severity CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')),
    CONSTRAINT valid_alert_status CHECK (status IN ('active', 'acknowledged', 'resolved', 'ignored')),
    INDEX idx_alerts_type (alert_type),
    INDEX idx_alerts_severity (severity),
    INDEX idx_alerts_status (status),
    INDEX idx_alerts_created (created_at)
);

-- =====================================================
-- FUNCTIONS AND TRIGGERS
-- =====================================================

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply update trigger to all relevant tables
CREATE TRIGGER update_users_timestamp BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_clusters_timestamp BEFORE UPDATE ON clusters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_nodes_timestamp BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_models_timestamp BEFORE UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_model_replicas_timestamp BEFORE UPDATE ON model_replicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Additional performance indexes
CREATE INDEX idx_users_email_lower ON users(LOWER(email));
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;
CREATE INDEX idx_nodes_online ON nodes(status) WHERE status = 'online';
CREATE INDEX idx_models_public ON models(is_public) WHERE is_public = true;
CREATE INDEX idx_inference_recent ON inference_requests(created_at DESC) WHERE created_at > CURRENT_DATE - INTERVAL '7 days';
CREATE INDEX idx_transfers_active ON transfers(status) WHERE status IN ('active', 'pending');

-- =====================================================
-- INITIAL DATA
-- =====================================================

-- Insert default cluster
INSERT INTO clusters (name, description, version, status) 
VALUES ('default', 'Default OllamaMax Cluster', '1.0.0', 'initializing');

-- Insert default admin user (password: admin123 - should be changed immediately)
INSERT INTO users (username, email, password_hash, full_name, role, is_active, is_verified)
VALUES ('admin', 'admin@ollamamax.io', crypt('admin123', gen_salt('bf')), 'System Administrator', 'admin', true, true);

-- =====================================================
-- PERMISSIONS
-- =====================================================

-- Create read-only role
CREATE ROLE ollamamax_readonly;
GRANT CONNECT ON DATABASE postgres TO ollamamax_readonly;
GRANT USAGE ON SCHEMA ollamamax TO ollamamax_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA ollamamax TO ollamamax_readonly;

-- Create application role
CREATE ROLE ollamamax_app;
GRANT CONNECT ON DATABASE postgres TO ollamamax_app;
GRANT USAGE ON SCHEMA ollamamax TO ollamamax_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ollamamax TO ollamamax_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ollamamax TO ollamamax_app;

-- Migration complete
-- Version: 001
-- Status: SUCCESS