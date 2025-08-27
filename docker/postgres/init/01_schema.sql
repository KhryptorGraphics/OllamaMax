-- OllamaMax Database Schema
-- Version 1.0
-- Created: 2025-08-24

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Models table for centralized model information
CREATE TABLE models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    version VARCHAR(50) NOT NULL DEFAULT '1.0.0',
    size BIGINT NOT NULL,
    hash VARCHAR(64) NOT NULL,
    content_type VARCHAR(100) DEFAULT 'application/octet-stream',
    description TEXT,
    tags JSONB DEFAULT '[]',
    parameters JSONB DEFAULT '{}',
    model_file_path VARCHAR(500),
    quantization_level VARCHAR(20),
    parameter_size VARCHAR(20),
    family VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'deprecated', 'deleted'))
);

-- Nodes table for cluster management
CREATE TABLE nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    peer_id VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    region VARCHAR(100),
    zone VARCHAR(100),
    address INET,
    port INTEGER,
    capabilities JSONB DEFAULT '{}',
    resources JSONB DEFAULT '{}', -- CPU, memory, storage
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'draining', 'offline', 'failed')),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    version VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Model replicas for distributed storage tracking
CREATE TABLE model_replicas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    replica_path VARCHAR(500) NOT NULL,
    replica_hash VARCHAR(64),
    replica_size BIGINT,
    status VARCHAR(20) DEFAULT 'syncing' CHECK (status IN ('syncing', 'ready', 'failed', 'deleted')),
    health_score FLOAT DEFAULT 1.0 CHECK (health_score >= 0 AND health_score <= 1),
    last_verified TIMESTAMP WITH TIME ZONE,
    sync_progress FLOAT DEFAULT 0.0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(model_id, node_id)
);

-- Users table for authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(320) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    roles TEXT[] DEFAULT ARRAY['user'],
    permissions TEXT[] DEFAULT ARRAY[],
    active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Sessions table for JWT token management
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_id VARCHAR(255) NOT NULL UNIQUE,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    user_agent TEXT,
    revoked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Inference requests for analytics and monitoring
CREATE TABLE inference_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(255) NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id),
    model_id UUID NOT NULL REFERENCES models(id),
    model_name VARCHAR(255) NOT NULL,
    prompt_hash VARCHAR(64),
    prompt_length INTEGER,
    response_length INTEGER,
    tokens_processed INTEGER,
    nodes_used TEXT[],
    partition_strategy VARCHAR(50),
    execution_time_ms INTEGER,
    queue_time_ms INTEGER,
    total_time_ms INTEGER,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Audit log for security and compliance
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(10) NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    row_id UUID,
    old_values JSONB,
    new_values JSONB,
    user_id UUID REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- System configuration table
CREATE TABLE system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(255) NOT NULL UNIQUE,
    value JSONB NOT NULL,
    description TEXT,
    category VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID REFERENCES users(id)
);

-- Performance metrics table
CREATE TABLE performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(255) NOT NULL,
    metric_value JSONB NOT NULL,
    tags JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    node_id UUID REFERENCES nodes(id)
);

-- Model usage statistics
CREATE TABLE model_usage_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    request_count INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    average_response_time_ms FLOAT DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    success_rate FLOAT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(model_id, date)
);

-- Create indexes for performance

-- Models indexes
CREATE INDEX idx_models_name ON models(name);
CREATE INDEX idx_models_status ON models(status);
CREATE INDEX idx_models_created_at ON models(created_at);
CREATE INDEX idx_models_tags ON models USING gin(tags);
CREATE INDEX idx_models_size ON models(size);
CREATE INDEX idx_models_hash ON models(hash);

-- Nodes indexes
CREATE INDEX idx_nodes_peer_id ON nodes(peer_id);
CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_last_heartbeat ON nodes(last_heartbeat);
CREATE INDEX idx_nodes_region_zone ON nodes(region, zone);
CREATE INDEX idx_nodes_capabilities ON nodes USING gin(capabilities);

-- Model replicas indexes
CREATE INDEX idx_replicas_model_id ON model_replicas(model_id);
CREATE INDEX idx_replicas_node_id ON model_replicas(node_id);
CREATE INDEX idx_replicas_status ON model_replicas(status);
CREATE INDEX idx_replicas_health_score ON model_replicas(health_score);
CREATE INDEX idx_replicas_last_verified ON model_replicas(last_verified);

-- Users indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(active);
CREATE INDEX idx_users_roles ON users USING gin(roles);
CREATE INDEX idx_users_last_login ON users(last_login_at);

-- Sessions indexes
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token_id ON user_sessions(token_id);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_sessions_revoked ON user_sessions(revoked);

-- Inference requests indexes
CREATE INDEX idx_inference_request_id ON inference_requests(request_id);
CREATE INDEX idx_inference_user_id ON inference_requests(user_id);
CREATE INDEX idx_inference_model_id ON inference_requests(model_id);
CREATE INDEX idx_inference_status ON inference_requests(status);
CREATE INDEX idx_inference_created_at ON inference_requests(created_at);
CREATE INDEX idx_inference_model_name ON inference_requests(model_name);
CREATE INDEX idx_inference_execution_time ON inference_requests(execution_time_ms);

-- Audit log indexes
CREATE INDEX idx_audit_table_name ON audit_log(table_name);
CREATE INDEX idx_audit_operation ON audit_log(operation);
CREATE INDEX idx_audit_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_row_id ON audit_log(row_id);

-- System config indexes
CREATE INDEX idx_system_config_key ON system_config(key);
CREATE INDEX idx_system_config_category ON system_config(category);

-- Performance metrics indexes
CREATE INDEX idx_perf_metrics_name ON performance_metrics(metric_name);
CREATE INDEX idx_perf_metrics_timestamp ON performance_metrics(timestamp);
CREATE INDEX idx_perf_metrics_node_id ON performance_metrics(node_id);
CREATE INDEX idx_perf_metrics_tags ON performance_metrics USING gin(tags);

-- Model usage stats indexes
CREATE INDEX idx_usage_stats_model_id ON model_usage_stats(model_id);
CREATE INDEX idx_usage_stats_date ON model_usage_stats(date);
CREATE INDEX idx_usage_stats_request_count ON model_usage_stats(request_count);

-- Create views for common queries

-- Active models with replica information
CREATE VIEW active_models_with_replicas AS
SELECT 
    m.id,
    m.name,
    m.version,
    m.size,
    m.hash,
    m.description,
    m.tags,
    m.created_at,
    COUNT(mr.id) as replica_count,
    COUNT(CASE WHEN mr.status = 'ready' THEN 1 END) as ready_replicas,
    AVG(mr.health_score) as average_health_score
FROM models m
LEFT JOIN model_replicas mr ON m.id = mr.model_id
WHERE m.status = 'active'
GROUP BY m.id, m.name, m.version, m.size, m.hash, m.description, m.tags, m.created_at;

-- Node health status
CREATE VIEW node_health_status AS
SELECT 
    n.id,
    n.peer_id,
    n.name,
    n.region,
    n.zone,
    n.status,
    n.last_heartbeat,
    CASE 
        WHEN n.last_heartbeat > NOW() - INTERVAL '5 minutes' THEN 'healthy'
        WHEN n.last_heartbeat > NOW() - INTERVAL '15 minutes' THEN 'degraded'
        ELSE 'unhealthy'
    END as health_status,
    COUNT(mr.id) as model_count,
    COUNT(CASE WHEN mr.status = 'ready' THEN 1 END) as ready_models
FROM nodes n
LEFT JOIN model_replicas mr ON n.id = mr.node_id
GROUP BY n.id, n.peer_id, n.name, n.region, n.zone, n.status, n.last_heartbeat;

-- Recent inference activity
CREATE VIEW recent_inference_activity AS
SELECT 
    ir.id,
    ir.request_id,
    u.username,
    ir.model_name,
    ir.status,
    ir.execution_time_ms,
    ir.tokens_processed,
    ir.created_at,
    ir.completed_at
FROM inference_requests ir
LEFT JOIN users u ON ir.user_id = u.id
WHERE ir.created_at > NOW() - INTERVAL '24 hours'
ORDER BY ir.created_at DESC;