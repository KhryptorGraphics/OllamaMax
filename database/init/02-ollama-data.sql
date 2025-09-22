-- PostgreSQL initialization script for OllamaMax core data structures
-- This script creates tables for managing Ollama nodes, models, and conversations

-- Create nodes table for distributed Ollama instances
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 11434,
    is_primary BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    last_health_check TIMESTAMPTZ,
    capabilities JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create models table for tracking available models
CREATE TABLE IF NOT EXISTS models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    tag VARCHAR(50) NOT NULL,
    size BIGINT,
    digest VARCHAR(255),
    family VARCHAR(50),
    format VARCHAR(20),
    parameters JSONB DEFAULT '{}',
    template TEXT,
    system TEXT,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(name, tag)
);

-- Create node_models junction table for model availability per node
CREATE TABLE IF NOT EXISTS node_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    is_loaded BOOLEAN DEFAULT FALSE,
    last_used TIMESTAMPTZ,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(node_id, model_id)
);

-- Create conversations table for chat history
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255),
    model_name VARCHAR(100) NOT NULL,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create messages table for conversation messages
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tokens_used INTEGER,
    response_time_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Create model_usage table for analytics
CREATE TABLE IF NOT EXISTS model_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    response_time_ms INTEGER,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create system_settings table for configuration
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create load_balancing_stats table for performance tracking
CREATE TABLE IF NOT EXISTS load_balancing_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    active_connections INTEGER DEFAULT 0,
    total_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    avg_response_time_ms REAL DEFAULT 0,
    cpu_usage REAL DEFAULT 0,
    memory_usage REAL DEFAULT 0,
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_nodes_host_port ON nodes(host, port);
CREATE INDEX IF NOT EXISTS idx_nodes_active ON nodes(is_active);
CREATE INDEX IF NOT EXISTS idx_nodes_primary ON nodes(is_primary);
CREATE INDEX IF NOT EXISTS idx_nodes_health ON nodes(health_status);

CREATE INDEX IF NOT EXISTS idx_models_name_tag ON models(name, tag);
CREATE INDEX IF NOT EXISTS idx_models_available ON models(is_available);

CREATE INDEX IF NOT EXISTS idx_node_models_node_id ON node_models(node_id);
CREATE INDEX IF NOT EXISTS idx_node_models_model_id ON node_models(model_id);
CREATE INDEX IF NOT EXISTS idx_node_models_loaded ON node_models(is_loaded);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_model ON conversations(model_name);
CREATE INDEX IF NOT EXISTS idx_conversations_active ON conversations(is_active);
CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

CREATE INDEX IF NOT EXISTS idx_model_usage_user_id ON model_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_model_id ON model_usage(model_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_node_id ON model_usage(node_id);
CREATE INDEX IF NOT EXISTS idx_model_usage_created_at ON model_usage(created_at);

CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);
CREATE INDEX IF NOT EXISTS idx_system_settings_public ON system_settings(is_public);

CREATE INDEX IF NOT EXISTS idx_load_balancing_stats_node_id ON load_balancing_stats(node_id);
CREATE INDEX IF NOT EXISTS idx_load_balancing_stats_recorded_at ON load_balancing_stats(recorded_at);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_nodes_updated_at BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_models_updated_at BEFORE UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE ON system_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to register a new node
CREATE OR REPLACE FUNCTION register_node(
    p_name VARCHAR(100),
    p_host VARCHAR(255),
    p_port INTEGER DEFAULT 11434,
    p_is_primary BOOLEAN DEFAULT FALSE,
    p_capabilities JSONB DEFAULT '{}'
)
RETURNS UUID AS $$
DECLARE
    new_node_id UUID;
BEGIN
    -- If this is being set as primary, unset other primary nodes
    IF p_is_primary THEN
        UPDATE nodes SET is_primary = FALSE WHERE is_primary = TRUE;
    END IF;
    
    INSERT INTO nodes (name, host, port, is_primary, capabilities, health_status)
    VALUES (p_name, p_host, p_port, p_is_primary, p_capabilities, 'unknown')
    RETURNING id INTO new_node_id;
    
    -- Log node registration
    INSERT INTO audit_logs (action, details)
    VALUES ('NODE_REGISTERED', jsonb_build_object('node_id', new_node_id, 'name', p_name, 'host', p_host, 'port', p_port));
    
    RETURN new_node_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to update node health
CREATE OR REPLACE FUNCTION update_node_health(
    p_node_id UUID,
    p_health_status VARCHAR(20)
)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE nodes 
    SET health_status = p_health_status,
        last_health_check = NOW(),
        updated_at = NOW()
    WHERE id = p_node_id;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Create function to get load balancing information
CREATE OR REPLACE FUNCTION get_load_balancing_info()
RETURNS JSON AS $$
BEGIN
    RETURN json_build_object(
        'total_nodes', (SELECT COUNT(*) FROM nodes WHERE is_active = TRUE),
        'healthy_nodes', (SELECT COUNT(*) FROM nodes WHERE is_active = TRUE AND health_status = 'healthy'),
        'primary_node', (SELECT json_build_object('id', id, 'name', name, 'host', host, 'port', port) FROM nodes WHERE is_primary = TRUE AND is_active = TRUE LIMIT 1),
        'total_models', (SELECT COUNT(*) FROM models WHERE is_available = TRUE),
        'active_conversations', (SELECT COUNT(*) FROM conversations WHERE is_active = TRUE),
        'total_messages_today', (SELECT COUNT(*) FROM messages WHERE created_at >= CURRENT_DATE)
    );
END;
$$ LANGUAGE plpgsql;

-- Create function to get model recommendations
CREATE OR REPLACE FUNCTION get_model_recommendations(p_user_id UUID DEFAULT NULL)
RETURNS TABLE(
    model_name VARCHAR(100),
    model_tag VARCHAR(50),
    usage_count BIGINT,
    avg_response_time REAL,
    success_rate REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.name,
        m.tag,
        COUNT(mu.id) as usage_count,
        AVG(mu.response_time_ms::REAL) as avg_response_time,
        (COUNT(CASE WHEN mu.success THEN 1 END)::REAL / COUNT(*)::REAL) as success_rate
    FROM models m
    LEFT JOIN model_usage mu ON m.id = mu.model_id
    WHERE (p_user_id IS NULL OR mu.user_id = p_user_id)
    AND m.is_available = TRUE
    GROUP BY m.id, m.name, m.tag
    HAVING COUNT(mu.id) > 0
    ORDER BY usage_count DESC, success_rate DESC, avg_response_time ASC
    LIMIT 10;
END;
$$ LANGUAGE plpgsql;

-- Create function to clean old data
CREATE OR REPLACE FUNCTION cleanup_old_data(
    p_days_to_keep INTEGER DEFAULT 90
)
RETURNS JSON AS $$
DECLARE
    deleted_stats INTEGER;
    deleted_usage INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - INTERVAL '1 day' * p_days_to_keep;
    
    -- Clean old load balancing stats
    DELETE FROM load_balancing_stats WHERE recorded_at < cutoff_date;
    GET DIAGNOSTICS deleted_stats = ROW_COUNT;
    
    -- Clean old model usage records
    DELETE FROM model_usage WHERE created_at < cutoff_date;
    GET DIAGNOSTICS deleted_usage = ROW_COUNT;
    
    -- Log cleanup
    INSERT INTO audit_logs (action, details)
    VALUES ('DATA_CLEANUP', jsonb_build_object(
        'cutoff_date', cutoff_date,
        'deleted_stats', deleted_stats,
        'deleted_usage', deleted_usage
    ));
    
    RETURN json_build_object(
        'deleted_stats', deleted_stats,
        'deleted_usage', deleted_usage,
        'cutoff_date', cutoff_date
    );
END;
$$ LANGUAGE plpgsql;

-- Insert default nodes
INSERT INTO nodes (name, host, port, is_primary, health_status) VALUES
    ('ollama-primary', 'ollama-primary', 11434, TRUE, 'unknown'),
    ('ollama-worker-1', 'ollama-worker-1', 11434, FALSE, 'unknown'),
    ('ollama-worker-2', 'ollama-worker-2', 11434, FALSE, 'unknown')
ON CONFLICT DO NOTHING;

-- Insert default system settings
INSERT INTO system_settings (key, value, description, is_public) VALUES
    ('max_conversation_history', '100', 'Maximum number of messages to keep in conversation history', TRUE),
    ('default_model', '"llama2"', 'Default model to use for new conversations', TRUE),
    ('load_balancing_algorithm', '"round_robin"', 'Load balancing algorithm: round_robin, least_connections, weighted', FALSE),
    ('health_check_interval', '30', 'Health check interval in seconds', FALSE),
    ('max_tokens_per_request', '4096', 'Maximum tokens allowed per request', TRUE),
    ('rate_limit_per_minute', '60', 'Rate limit per user per minute', TRUE)
ON CONFLICT (key) DO NOTHING;

-- Insert default models (these will be updated by the application)
INSERT INTO models (name, tag, family, format) VALUES
    ('llama2', 'latest', 'llama', 'gguf'),
    ('codellama', 'latest', 'llama', 'gguf'),
    ('mistral', 'latest', 'mistral', 'gguf')
ON CONFLICT (name, tag) DO NOTHING;

-- Log initialization
INSERT INTO audit_logs (action, details) 
VALUES ('OLLAMA_SCHEMA_INITIALIZED', jsonb_build_object(
    'timestamp', NOW(), 
    'version', '1.0',
    'tables_created', ARRAY['nodes', 'models', 'node_models', 'conversations', 'messages', 'model_usage', 'system_settings', 'load_balancing_stats']
));