-- OllamaMax Default Data
-- Version 1.0
-- Created: 2025-08-24

-- Insert default admin user
INSERT INTO users (
    username, 
    email, 
    password_hash, 
    roles, 
    permissions,
    active
) VALUES (
    'admin',
    'admin@ollamamax.local',
    '$2a$10$rT1vZ5ZxJxJYQWkYBzKYDeX.BsKKGqHoKvdaUXqKYk7TYnFQUqJAa', -- password: admin123
    ARRAY['admin'],
    ARRAY[
        'model:manage', 'model:read',
        'cluster:manage', 'cluster:read',
        'node:manage', 'node:read',
        'inference:run', 'metrics:read',
        'system:manage'
    ],
    true
);

-- Insert default system user for internal operations
INSERT INTO users (
    username, 
    email, 
    password_hash, 
    roles, 
    permissions,
    active
) VALUES (
    'system',
    'system@ollamamax.local',
    '$2a$10$system.internal.hash.placeholder.secure',
    ARRAY['system'],
    ARRAY['*'],
    true
);

-- Insert default operator user
INSERT INTO users (
    username, 
    email, 
    password_hash, 
    roles, 
    permissions,
    active
) VALUES (
    'operator',
    'operator@ollamamax.local',
    '$2a$10$operator.default.hash.change.in.production',
    ARRAY['operator'],
    ARRAY[
        'model:read', 'cluster:read',
        'node:read', 'inference:run',
        'metrics:read'
    ],
    true
);

-- Insert default system configuration
INSERT INTO system_config (key, value, description, category) VALUES
    ('cluster.min_replicas', '2', 'Minimum number of replicas for each model', 'replication'),
    ('cluster.max_replicas', '5', 'Maximum number of replicas for each model', 'replication'),
    ('cluster.replication_factor', '2', 'Default replication factor for new models', 'replication'),
    ('cluster.health_check_interval', '30', 'Health check interval in seconds', 'monitoring'),
    ('cluster.heartbeat_timeout', '60', 'Node heartbeat timeout in seconds', 'monitoring'),
    ('inference.max_concurrent_requests', '100', 'Maximum concurrent inference requests', 'performance'),
    ('inference.default_timeout', '300', 'Default inference timeout in seconds', 'performance'),
    ('inference.queue_size', '1000', 'Maximum inference queue size', 'performance'),
    ('auth.session_timeout', '3600', 'User session timeout in seconds', 'security'),
    ('auth.max_failed_attempts', '5', 'Maximum failed login attempts before lockout', 'security'),
    ('auth.lockout_duration', '1800', 'Account lockout duration in seconds', 'security'),
    ('storage.cleanup_interval', '86400', 'Storage cleanup interval in seconds', 'storage'),
    ('storage.max_disk_usage', '85', 'Maximum disk usage percentage before cleanup', 'storage'),
    ('metrics.retention_days', '30', 'Metrics retention period in days', 'monitoring'),
    ('audit.retention_days', '90', 'Audit log retention period in days', 'security'),
    ('p2p.connection_timeout', '30', 'P2P connection timeout in seconds', 'networking'),
    ('p2p.max_peers', '50', 'Maximum number of P2P peers', 'networking'),
    ('api.rate_limit_rpm', '1000', 'API rate limit per minute per user', 'api'),
    ('api.max_request_size', '33554432', 'Maximum API request size in bytes (32MB)', 'api'),
    ('scheduler.algorithm', 'round_robin', 'Default scheduling algorithm', 'scheduling'),
    ('scheduler.load_balancing', 'least_connections', 'Load balancing strategy', 'scheduling');

-- Insert some example models (these would normally be populated when models are pulled)
INSERT INTO models (
    name, 
    version, 
    size, 
    hash, 
    content_type, 
    description, 
    tags, 
    parameters,
    quantization_level,
    parameter_size,
    family,
    status
) VALUES 
    (
        'llama2:7b-chat',
        '1.0.0',
        3825819519, -- ~3.8GB
        'sha256:dummy_hash_llama2_7b_chat_replace_with_actual',
        'application/octet-stream',
        'Llama 2 7B Chat model optimized for conversational AI',
        '["llama2", "chat", "7b", "meta"]',
        '{"context_length": 4096, "temperature": 0.8, "top_p": 0.95}',
        'Q4_0',
        '7B',
        'llama',
        'active'
    ),
    (
        'mistral:7b-instruct',
        '1.0.0',
        4109453312, -- ~4.1GB
        'sha256:dummy_hash_mistral_7b_instruct_replace_with_actual',
        'application/octet-stream',
        'Mistral 7B Instruct model for instruction following',
        '["mistral", "instruct", "7b"]',
        '{"context_length": 8192, "temperature": 0.7, "top_p": 0.9}',
        'Q4_0',
        '7B',
        'mistral',
        'active'
    ),
    (
        'codellama:7b-code',
        '1.0.0',
        3825819519, -- ~3.8GB
        'sha256:dummy_hash_codellama_7b_code_replace_with_actual',
        'application/octet-stream',
        'Code Llama 7B specialized for code generation',
        '["codellama", "code", "7b", "meta"]',
        '{"context_length": 16384, "temperature": 0.1, "top_p": 0.95}',
        'Q4_0',
        '7B',
        'llama',
        'active'
    );

-- Create some performance baselines
INSERT INTO performance_metrics (metric_name, metric_value, tags) VALUES
    ('system.startup_time_ms', '2500', '{"component": "system", "version": "1.0.0"}'),
    ('database.connection_pool_size', '25', '{"component": "database", "backend": "postgresql"}'),
    ('cache.hit_ratio', '0.85', '{"component": "cache", "backend": "redis"}'),
    ('p2p.discovery_time_ms', '1200', '{"component": "p2p", "protocol": "libp2p"}'),
    ('api.avg_response_time_ms', '150', '{"component": "api", "endpoint": "health"}');

-- Create cleanup job schedules (these would be run by a scheduler)
INSERT INTO system_config (key, value, description, category) VALUES
    ('job.cleanup_sessions.enabled', 'true', 'Enable automatic session cleanup', 'jobs'),
    ('job.cleanup_sessions.schedule', '0 */6 * * *', 'Cleanup sessions every 6 hours', 'jobs'),
    ('job.cleanup_audit_logs.enabled', 'true', 'Enable automatic audit log cleanup', 'jobs'),
    ('job.cleanup_audit_logs.schedule', '0 2 * * *', 'Cleanup audit logs daily at 2 AM', 'jobs'),
    ('job.cleanup_metrics.enabled', 'true', 'Enable automatic metrics cleanup', 'jobs'),
    ('job.cleanup_metrics.schedule', '0 3 * * *', 'Cleanup metrics daily at 3 AM', 'jobs'),
    ('job.update_statistics.enabled', 'true', 'Enable automatic statistics updates', 'jobs'),
    ('job.update_statistics.schedule', '*/15 * * * *', 'Update statistics every 15 minutes', 'jobs');

-- Add some initial model usage statistics for the example models
INSERT INTO model_usage_stats (model_id, date, request_count, total_tokens, average_response_time_ms, unique_users, success_rate)
SELECT 
    m.id,
    CURRENT_DATE - INTERVAL '1 day',
    FLOOR(RANDOM() * 100 + 10)::INTEGER,
    FLOOR(RANDOM() * 10000 + 1000)::INTEGER,
    FLOOR(RANDOM() * 2000 + 500)::INTEGER,
    FLOOR(RANDOM() * 20 + 5)::INTEGER,
    0.95 + (RANDOM() * 0.05)
FROM models m
WHERE m.status = 'active';

-- Create system health check entry
INSERT INTO performance_metrics (metric_name, metric_value, tags) VALUES
    ('system.health_status', '"healthy"', '{"component": "system", "check": "overall"}'),
    ('system.uptime_seconds', '0', '{"component": "system", "metric": "uptime"}');

-- Add welcome message in system config
INSERT INTO system_config (key, value, description, category) VALUES
    ('system.welcome_message', '"Welcome to OllamaMax Distributed LLM Platform v1.0"', 'Welcome message for API', 'ui'),
    ('system.version', '"1.0.0"', 'System version', 'system'),
    ('system.build_time', '"2025-08-24T16:00:00Z"', 'System build timestamp', 'system'),
    ('system.environment', '"production"', 'System environment', 'system');