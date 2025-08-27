-- OllamaMax Database Functions and Triggers
-- Version 1.0
-- Created: 2025-08-24

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
CREATE TRIGGER update_models_updated_at BEFORE UPDATE ON models
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_nodes_updated_at BEFORE UPDATE ON nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_model_replicas_updated_at BEFORE UPDATE ON model_replicas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_config_updated_at BEFORE UPDATE ON system_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function for audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    -- INSERT operation
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            table_name, 
            operation, 
            row_id, 
            new_values, 
            user_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            NEW.id,
            to_jsonb(NEW),
            NULLIF(current_setting('app.current_user_id', true), '')::UUID
        );
        RETURN NEW;
    
    -- UPDATE operation
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            table_name, 
            operation, 
            row_id, 
            old_values, 
            new_values, 
            user_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            NEW.id,
            to_jsonb(OLD),
            to_jsonb(NEW),
            NULLIF(current_setting('app.current_user_id', true), '')::UUID
        );
        RETURN NEW;
    
    -- DELETE operation
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (
            table_name, 
            operation, 
            row_id, 
            old_values, 
            user_id
        ) VALUES (
            TG_TABLE_NAME,
            TG_OP,
            OLD.id,
            to_jsonb(OLD),
            NULLIF(current_setting('app.current_user_id', true), '')::UUID
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit triggers for important tables
CREATE TRIGGER audit_models_trigger
    AFTER INSERT OR UPDATE OR DELETE ON models
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_users_trigger
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_nodes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON nodes
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_model_replicas_trigger
    AFTER INSERT OR UPDATE OR DELETE ON model_replicas
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions 
    WHERE expires_at < CURRENT_TIMESTAMP 
       OR (refresh_expires_at IS NOT NULL AND refresh_expires_at < CURRENT_TIMESTAMP)
       OR revoked = true;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to update model usage statistics
CREATE OR REPLACE FUNCTION update_model_usage_stats(
    p_model_id UUID,
    p_request_count INTEGER DEFAULT 1,
    p_tokens_processed INTEGER DEFAULT 0,
    p_response_time_ms INTEGER DEFAULT 0,
    p_user_id UUID DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO model_usage_stats (
        model_id,
        date,
        request_count,
        total_tokens,
        average_response_time_ms,
        unique_users
    )
    VALUES (
        p_model_id,
        CURRENT_DATE,
        p_request_count,
        p_tokens_processed,
        p_response_time_ms,
        CASE WHEN p_user_id IS NOT NULL THEN 1 ELSE 0 END
    )
    ON CONFLICT (model_id, date)
    DO UPDATE SET
        request_count = model_usage_stats.request_count + p_request_count,
        total_tokens = model_usage_stats.total_tokens + p_tokens_processed,
        average_response_time_ms = (
            (model_usage_stats.average_response_time_ms * model_usage_stats.request_count + p_response_time_ms * p_request_count) 
            / (model_usage_stats.request_count + p_request_count)
        );
END;
$$ LANGUAGE plpgsql;

-- Function to get model health score
CREATE OR REPLACE FUNCTION calculate_model_health_score(p_model_id UUID)
RETURNS FLOAT AS $$
DECLARE
    avg_health FLOAT;
    replica_count INTEGER;
    ready_replicas INTEGER;
BEGIN
    SELECT 
        AVG(health_score),
        COUNT(*),
        COUNT(CASE WHEN status = 'ready' THEN 1 END)
    INTO avg_health, replica_count, ready_replicas
    FROM model_replicas 
    WHERE model_id = p_model_id;
    
    -- If no replicas, return 0
    IF replica_count = 0 THEN
        RETURN 0.0;
    END IF;
    
    -- Calculate weighted health score
    -- 50% based on average health score of replicas
    -- 50% based on percentage of ready replicas
    RETURN (
        COALESCE(avg_health, 0) * 0.5 + 
        (ready_replicas::FLOAT / replica_count::FLOAT) * 0.5
    );
END;
$$ LANGUAGE plpgsql;

-- Function to get node capacity utilization
CREATE OR REPLACE FUNCTION get_node_capacity_utilization(p_node_id UUID)
RETURNS JSONB AS $$
DECLARE
    node_resources JSONB;
    model_count INTEGER;
    total_model_size BIGINT;
    result JSONB;
BEGIN
    SELECT resources INTO node_resources FROM nodes WHERE id = p_node_id;
    
    SELECT 
        COUNT(*),
        COALESCE(SUM(m.size), 0)
    INTO model_count, total_model_size
    FROM model_replicas mr
    JOIN models m ON mr.model_id = m.id
    WHERE mr.node_id = p_node_id AND mr.status = 'ready';
    
    result := jsonb_build_object(
        'model_count', model_count,
        'total_model_size_bytes', total_model_size,
        'total_model_size_gb', ROUND((total_model_size / (1024.0^3))::NUMERIC, 2)
    );
    
    -- Add storage utilization if available
    IF node_resources ? 'storage_bytes' THEN
        result := result || jsonb_build_object(
            'storage_utilization_percent',
            ROUND((total_model_size::FLOAT / (node_resources->>'storage_bytes')::BIGINT * 100)::NUMERIC, 2)
        );
    END IF;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Function to find optimal nodes for model placement
CREATE OR REPLACE FUNCTION find_optimal_nodes_for_model(
    p_model_id UUID,
    p_required_replicas INTEGER DEFAULT 2
)
RETURNS TABLE (
    node_id UUID,
    peer_id VARCHAR,
    name VARCHAR,
    region VARCHAR,
    zone VARCHAR,
    utilization_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    WITH node_utilization AS (
        SELECT 
            n.id,
            n.peer_id,
            n.name,
            n.region,
            n.zone,
            COUNT(mr.id) as current_models,
            COALESCE(SUM(m.size), 0) as used_storage,
            COALESCE((n.resources->>'storage_bytes')::BIGINT, 0) as total_storage
        FROM nodes n
        LEFT JOIN model_replicas mr ON n.id = mr.node_id AND mr.status IN ('syncing', 'ready')
        LEFT JOIN models m ON mr.model_id = m.id
        WHERE n.status = 'active' 
          AND n.last_heartbeat > NOW() - INTERVAL '5 minutes'
          AND n.id NOT IN (
              SELECT node_id FROM model_replicas 
              WHERE model_id = p_model_id AND status IN ('syncing', 'ready')
          )
        GROUP BY n.id, n.peer_id, n.name, n.region, n.zone, n.resources
    )
    SELECT 
        nu.id,
        nu.peer_id,
        nu.name,
        nu.region,
        nu.zone,
        CASE 
            WHEN nu.total_storage > 0 THEN 
                (nu.used_storage::FLOAT / nu.total_storage::FLOAT) * 0.7 + 
                (nu.current_models::FLOAT / 10.0) * 0.3
            ELSE nu.current_models::FLOAT / 10.0
        END as utilization_score
    FROM node_utilization nu
    ORDER BY utilization_score ASC
    LIMIT p_required_replicas;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old audit logs
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(p_days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_log 
    WHERE timestamp < CURRENT_TIMESTAMP - (p_days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup old performance metrics
CREATE OR REPLACE FUNCTION cleanup_old_metrics(p_days_to_keep INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM performance_metrics 
    WHERE timestamp < CURRENT_TIMESTAMP - (p_days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;