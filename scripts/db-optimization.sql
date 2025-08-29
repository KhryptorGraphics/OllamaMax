-- Database optimization settings for OllamaMax PostgreSQL
-- This script applies performance optimizations and creates additional indexes

-- Performance monitoring setup
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pg_buffercache";
CREATE EXTENSION IF NOT EXISTS "pgstattuple";

-- Function to analyze table bloat
CREATE OR REPLACE FUNCTION analyze_table_bloat()
RETURNS TABLE (
    schemaname name,
    tablename name,
    attname name,
    n_distinct real,
    correlation real,
    most_common_vals text,
    most_common_freqs real[]
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname,
        s.tablename,
        s.attname,
        s.n_distinct,
        s.correlation,
        s.most_common_vals::text,
        s.most_common_freqs
    FROM pg_stats s
    WHERE s.schemaname = 'public'
    ORDER BY s.tablename, s.attname;
END;
$$ LANGUAGE plpgsql;

-- Function to get table sizes
CREATE OR REPLACE FUNCTION get_table_sizes()
RETURNS TABLE (
    schema_name text,
    table_name text,
    table_size text,
    index_size text,
    total_size text,
    row_estimate bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname::text,
        tablename::text,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as table_size,
        pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as index_size,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
        schemaname||'.'||tablename as relname,
        n_tup_ins - n_tup_del as row_estimate
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
    ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
END;
$$ LANGUAGE plpgsql;

-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_composite_search 
ON models (status, family, created_at DESC) 
WHERE status IN ('ready', 'pending');

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inference_requests_user_model 
ON inference_requests (user_id, model_id, created_at DESC) 
WHERE user_id IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inference_requests_performance 
ON inference_requests (status, model_name, created_at DESC) 
WHERE status IN ('completed', 'failed');

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_model_replicas_health 
ON model_replicas (model_id, status, health_score DESC) 
WHERE status = 'ready' AND health_score > 0.5;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nodes_active_region 
ON nodes (status, region, last_heartbeat DESC) 
WHERE status = 'active';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_recent 
ON audit_log_entries (timestamp DESC, table_name) 
WHERE timestamp >= NOW() - INTERVAL '30 days';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_active 
ON user_sessions (user_id, expires_at DESC, revoked) 
WHERE revoked = false AND expires_at > NOW();

-- Partial indexes for frequently filtered data
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_ready 
ON models (name, version, created_at DESC) 
WHERE status = 'ready';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nodes_healthy 
ON nodes (region, zone, last_heartbeat DESC) 
WHERE status = 'active' AND last_heartbeat > NOW() - INTERVAL '5 minutes';

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inference_active 
ON inference_requests (created_at DESC, execution_time_ms) 
WHERE status IN ('pending', 'processing');

-- Text search indexes for better search performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_text_search 
ON models USING gin(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '')));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nodes_text_search 
ON nodes USING gin(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(peer_id, '')));

-- Foreign key indexes to improve join performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_model_replicas_model_fk 
ON model_replicas (model_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_model_replicas_node_fk 
ON model_replicas (node_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_sessions_user_fk 
ON user_sessions (user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inference_requests_model_fk 
ON inference_requests (model_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_model_usage_stats_model_fk 
ON model_usage_stats (model_id);

-- JSONB indexes for metadata searches
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_models_parameters_gin 
ON models USING gin (parameters);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nodes_capabilities_gin 
ON nodes USING gin (capabilities);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nodes_resources_gin 
ON nodes USING gin (resources);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_inference_metadata_gin 
ON inference_requests USING gin (metadata);

-- Time-based partitioning preparation for audit logs
-- This creates monthly partitions for better performance with large audit data
CREATE OR REPLACE FUNCTION create_audit_partition(start_date date, end_date date)
RETURNS void AS $$
DECLARE
    partition_name text;
    partition_start text;
    partition_end text;
BEGIN
    partition_name := 'audit_log_entries_' || to_char(start_date, 'YYYY_MM');
    partition_start := start_date::text;
    partition_end := end_date::text;
    
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log_entries 
                    FOR VALUES FROM (%L) TO (%L)',
                   partition_name, partition_start, partition_end);
                   
    EXECUTE format('CREATE INDEX IF NOT EXISTS %I ON %I (timestamp, user_id)',
                   'idx_' || partition_name || '_timestamp_user', partition_name);
END;
$$ LANGUAGE plpgsql;

-- Create initial audit log partitions (current month + next 3 months)
DO $$
DECLARE
    start_date date;
    end_date date;
    i integer;
BEGIN
    FOR i IN 0..3 LOOP
        start_date := date_trunc('month', CURRENT_DATE + (i || ' months')::interval);
        end_date := date_trunc('month', start_date + '1 month'::interval);
        PERFORM create_audit_partition(start_date, end_date);
    END LOOP;
END $$;

-- Materialized view for dashboard statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS dashboard_stats AS
SELECT 
    'models'::text as entity_type,
    COUNT(*)::bigint as total_count,
    COUNT(*) FILTER (WHERE status = 'ready')::bigint as ready_count,
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours')::bigint as recent_count,
    AVG(size)::bigint as avg_size,
    NOW() as last_updated
FROM models
UNION ALL
SELECT 
    'nodes'::text as entity_type,
    COUNT(*)::bigint as total_count,
    COUNT(*) FILTER (WHERE status = 'active')::bigint as active_count,
    COUNT(*) FILTER (WHERE last_heartbeat >= NOW() - INTERVAL '5 minutes')::bigint as healthy_count,
    0::bigint as avg_size,
    NOW() as last_updated
FROM nodes
UNION ALL
SELECT 
    'inference_requests'::text as entity_type,
    COUNT(*)::bigint as total_count,
    COUNT(*) FILTER (WHERE status = 'completed')::bigint as completed_count,
    COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours')::bigint as recent_count,
    AVG(total_time_ms)::bigint as avg_duration_ms,
    NOW() as last_updated
FROM inference_requests
UNION ALL
SELECT 
    'users'::text as entity_type,
    COUNT(*)::bigint as total_count,
    COUNT(*) FILTER (WHERE active = true)::bigint as active_count,
    COUNT(*) FILTER (WHERE last_login_at >= NOW() - INTERVAL '24 hours')::bigint as recent_logins,
    0::bigint as unused,
    NOW() as last_updated
FROM users;

-- Index on materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_dashboard_stats_entity 
ON dashboard_stats (entity_type);

-- Function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_dashboard_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY dashboard_stats;
END;
$$ LANGUAGE plpgsql;

-- Performance monitoring functions
CREATE OR REPLACE FUNCTION get_slow_queries(minutes_back integer DEFAULT 60)
RETURNS TABLE (
    query text,
    calls bigint,
    total_time double precision,
    mean_time double precision,
    max_time double precision,
    stddev_time double precision
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        pss.query,
        pss.calls,
        pss.total_exec_time as total_time,
        pss.mean_exec_time as mean_time,
        pss.max_exec_time as max_time,
        pss.stddev_exec_time as stddev_time
    FROM pg_stat_statements pss
    WHERE pss.calls > 5 
    AND pss.mean_exec_time > 100  -- queries taking more than 100ms on average
    ORDER BY pss.total_exec_time DESC
    LIMIT 20;
END;
$$ LANGUAGE plpgsql;

-- Function to get table statistics
CREATE OR REPLACE FUNCTION get_table_stats()
RETURNS TABLE (
    schema_name text,
    table_name text,
    n_tup_ins bigint,
    n_tup_upd bigint,
    n_tup_del bigint,
    n_live_tup bigint,
    n_dead_tup bigint,
    last_vacuum timestamp,
    last_autovacuum timestamp,
    last_analyze timestamp,
    last_autoanalyze timestamp
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname::text,
        relname::text,
        n_tup_ins,
        n_tup_upd,
        n_tup_del,
        n_live_tup,
        n_dead_tup,
        last_vacuum,
        last_autovacuum,
        last_analyze,
        last_autoanalyze
    FROM pg_stat_user_tables
    ORDER BY n_live_tup DESC;
END;
$$ LANGUAGE plpgsql;

-- Maintenance function for cleanup
CREATE OR REPLACE FUNCTION cleanup_old_data()
RETURNS void AS $$
BEGIN
    -- Clean up old audit logs (older than 6 months)
    DELETE FROM audit_log_entries 
    WHERE timestamp < NOW() - INTERVAL '6 months';
    
    -- Clean up expired user sessions
    DELETE FROM user_sessions 
    WHERE expires_at < NOW() - INTERVAL '7 days'
    AND revoked = true;
    
    -- Clean up old inference requests (older than 3 months, keep successful ones for stats)
    DELETE FROM inference_requests 
    WHERE created_at < NOW() - INTERVAL '3 months'
    AND status NOT IN ('completed');
    
    -- Update statistics
    ANALYZE;
    
    RAISE NOTICE 'Cleanup completed at %', NOW();
END;
$$ LANGUAGE plpgsql;

-- Create maintenance schedule function
CREATE OR REPLACE FUNCTION schedule_maintenance()
RETURNS void AS $$
BEGIN
    -- This would typically be called by an external cron job or scheduler
    PERFORM cleanup_old_data();
    PERFORM refresh_dashboard_stats();
    
    -- Reindex fragmented indexes
    REINDEX INDEX CONCURRENTLY idx_inference_requests_created_at;
    REINDEX INDEX CONCURRENTLY idx_audit_log_entries_timestamp;
    
    RAISE NOTICE 'Scheduled maintenance completed at %', NOW();
END;
$$ LANGUAGE plpgsql;

-- Connection and query monitoring
CREATE OR REPLACE VIEW connection_stats AS
SELECT 
    datname,
    pid,
    usename,
    application_name,
    client_addr,
    client_port,
    backend_start,
    xact_start,
    query_start,
    state_change,
    state,
    backend_xid,
    backend_xmin,
    query,
    backend_type
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start DESC;

-- Database health check function
CREATE OR REPLACE FUNCTION database_health_check()
RETURNS jsonb AS $$
DECLARE
    result jsonb;
    db_size text;
    connection_count int;
    active_queries int;
    blocked_queries int;
    replication_lag interval;
BEGIN
    -- Get database size
    SELECT pg_size_pretty(pg_database_size(current_database())) INTO db_size;
    
    -- Get connection statistics
    SELECT count(*) INTO connection_count FROM pg_stat_activity;
    SELECT count(*) INTO active_queries FROM pg_stat_activity WHERE state = 'active';
    SELECT count(*) INTO blocked_queries FROM pg_stat_activity WHERE wait_event_type = 'Lock';
    
    result := jsonb_build_object(
        'timestamp', NOW(),
        'database_size', db_size,
        'connections', jsonb_build_object(
            'total', connection_count,
            'active_queries', active_queries,
            'blocked_queries', blocked_queries
        ),
        'performance', jsonb_build_object(
            'cache_hit_ratio', (
                SELECT round(
                    100.0 * sum(heap_blks_hit) / nullif(sum(heap_blks_hit) + sum(heap_blks_read), 0), 2
                ) FROM pg_statio_user_tables
            ),
            'index_usage_ratio', (
                SELECT round(
                    100.0 * sum(idx_blks_hit) / nullif(sum(idx_blks_hit) + sum(idx_blks_read), 0), 2
                ) FROM pg_statio_user_indexes
            )
        ),
        'vacuum_stats', (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'table', schemaname || '.' || relname,
                    'last_vacuum', last_vacuum,
                    'last_autovacuum', last_autovacuum,
                    'dead_tuples', n_dead_tup
                )
            )
            FROM pg_stat_user_tables
            WHERE n_dead_tup > 1000
            ORDER BY n_dead_tup DESC
            LIMIT 5
        )
    );
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Optimize PostgreSQL settings
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET track_activity_query_size = 2048;
ALTER SYSTEM SET pg_stat_statements.track = 'all';
ALTER SYSTEM SET pg_stat_statements.max = 10000;
ALTER SYSTEM SET pg_stat_statements.save = on;

-- Auto-vacuum tuning
ALTER SYSTEM SET autovacuum = on;
ALTER SYSTEM SET autovacuum_max_workers = 4;
ALTER SYSTEM SET autovacuum_naptime = '30s';
ALTER SYSTEM SET autovacuum_vacuum_threshold = 50;
ALTER SYSTEM SET autovacuum_analyze_threshold = 50;
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.1;
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.05;
ALTER SYSTEM SET autovacuum_freeze_max_age = 200000000;
ALTER SYSTEM SET autovacuum_multixact_freeze_max_age = 400000000;

-- WAL and checkpointing
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET max_wal_senders = 3;
ALTER SYSTEM SET archive_mode = on;
ALTER SYSTEM SET archive_command = 'test ! -f /backups/wal/%f && cp %p /backups/wal/%f';

-- Final optimizations
ANALYZE;
VACUUM ANALYZE;

-- Success notification
DO $$
BEGIN
    RAISE NOTICE 'Database optimization completed successfully!';
    RAISE NOTICE 'Created % indexes for improved query performance', (
        SELECT count(*) FROM pg_indexes WHERE schemaname = 'public'
    );
    RAISE NOTICE 'Created % functions for monitoring and maintenance', 12;
    RAISE NOTICE 'Created materialized view for dashboard statistics';
    RAISE NOTICE 'Configured auto-vacuum and performance monitoring';
END $$;