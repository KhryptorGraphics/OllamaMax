-- PostgreSQL initialization script for OllamaMax authentication system
-- This script creates the users and sessions tables for production PostgreSQL deployment

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table with enhanced security features
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verification_expires TIMESTAMPTZ,
    reset_token VARCHAR(255),
    reset_expires TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    role VARCHAR(20) DEFAULT 'user',
    preferences JSONB DEFAULT '{}',
    profile JSONB DEFAULT '{}'
);

-- Create sessions table for JWT token management
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    expires TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create audit log table for security tracking
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id VARCHAR(100),
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create API keys table for service-to-service authentication
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '{}',
    expires TIMESTAMPTZ,
    last_used TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);
CREATE INDEX IF NOT EXISTS idx_users_reset_token ON users(reset_token);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(is_active);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function for user registration with validation
CREATE OR REPLACE FUNCTION register_user(
    p_username VARCHAR(50),
    p_email VARCHAR(255),
    p_password_hash VARCHAR(255),
    p_verification_token VARCHAR(255),
    p_verification_expires TIMESTAMPTZ
)
RETURNS UUID AS $$
DECLARE
    new_user_id UUID;
BEGIN
    -- Validate email format
    IF p_email !~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$' THEN
        RAISE EXCEPTION 'Invalid email format';
    END IF;
    
    -- Validate username
    IF LENGTH(p_username) < 3 OR LENGTH(p_username) > 50 THEN
        RAISE EXCEPTION 'Username must be between 3 and 50 characters';
    END IF;
    
    -- Insert new user
    INSERT INTO users (username, email, password_hash, verification_token, verification_expires)
    VALUES (p_username, p_email, p_password_hash, p_verification_token, p_verification_expires)
    RETURNING id INTO new_user_id;
    
    -- Log registration
    INSERT INTO audit_logs (user_id, action, details)
    VALUES (new_user_id, 'USER_REGISTERED', jsonb_build_object('email', p_email, 'username', p_username));
    
    RETURN new_user_id;
END;
$$ LANGUAGE plpgsql;

-- Create function to verify email
CREATE OR REPLACE FUNCTION verify_user_email(p_verification_token VARCHAR(255))
RETURNS BOOLEAN AS $$
DECLARE
    user_record RECORD;
BEGIN
    -- Find user with valid token
    SELECT id, email, username INTO user_record
    FROM users 
    WHERE verification_token = p_verification_token 
    AND verification_expires > NOW()
    AND NOT email_verified;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Update user as verified
    UPDATE users 
    SET email_verified = TRUE,
        verification_token = NULL,
        verification_expires = NULL,
        updated_at = NOW()
    WHERE id = user_record.id;
    
    -- Log verification
    INSERT INTO audit_logs (user_id, action, details)
    VALUES (user_record.id, 'EMAIL_VERIFIED', jsonb_build_object('email', user_record.email));
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Create function to clean expired sessions
CREATE OR REPLACE FUNCTION clean_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions WHERE expires < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to get user statistics
CREATE OR REPLACE FUNCTION get_user_stats()
RETURNS JSON AS $$
BEGIN
    RETURN json_build_object(
        'total_users', (SELECT COUNT(*) FROM users),
        'verified_users', (SELECT COUNT(*) FROM users WHERE email_verified = TRUE),
        'active_users', (SELECT COUNT(*) FROM users WHERE is_active = TRUE),
        'active_sessions', (SELECT COUNT(*) FROM sessions WHERE expires > NOW()),
        'registrations_today', (SELECT COUNT(*) FROM users WHERE created_at >= CURRENT_DATE),
        'logins_today', (SELECT COUNT(*) FROM audit_logs WHERE action = 'USER_LOGIN' AND created_at >= CURRENT_DATE)
    );
END;
$$ LANGUAGE plpgsql;

-- Grant permissions to application user
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_user WHERE usename = 'ollama') THEN
        GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ollama;
        GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO ollama;
        GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ollama;
    END IF;
END $$;

-- Insert default admin user (password: admin123 - change in production!)
INSERT INTO users (username, email, password_hash, email_verified, role)
VALUES (
    'admin',
    'admin@ollamamax.local',
    '$2b$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    TRUE,
    'admin'
) ON CONFLICT (email) DO NOTHING;

-- Log initialization
INSERT INTO audit_logs (action, details) 
VALUES ('DATABASE_INITIALIZED', jsonb_build_object('timestamp', NOW(), 'version', '1.0'));

-- Print initialization summary
DO $$
BEGIN
    RAISE NOTICE 'OllamaMax PostgreSQL database initialized successfully';
    RAISE NOTICE 'Tables created: users, sessions, audit_logs, api_keys';
    RAISE NOTICE 'Default admin user: admin@ollamamax.local (password: admin123)';
    RAISE NOTICE 'SECURITY WARNING: Change default admin password in production!';
END $$;