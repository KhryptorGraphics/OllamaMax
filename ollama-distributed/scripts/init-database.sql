-- OllamaMax Database Initialization Script
-- This script sets up the initial database structure

-- Create database if it doesn't exist (run this manually if needed)
-- CREATE DATABASE ollamamax;

-- Connect to the database
\c ollamamax;

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create default admin user (password: admin123)
-- This will be created by the migration system, but included here for reference
-- INSERT INTO users (id, username, email, password_hash, full_name, role, is_active, is_verified)
-- VALUES (
--     uuid_generate_v4(),
--     'admin',
--     'admin@ollamamax.local',
--     crypt('admin123', gen_salt('bf')),
--     'System Administrator',
--     'admin',
--     true,
--     true
-- ) ON CONFLICT (username) DO NOTHING;

-- Create default viewer user (password: viewer123)
-- INSERT INTO users (id, username, email, password_hash, full_name, role, is_active, is_verified)
-- VALUES (
--     uuid_generate_v4(),
--     'viewer',
--     'viewer@ollamamax.local',
--     crypt('viewer123', gen_salt('bf')),
--     'System Viewer',
--     'viewer',
--     true,
--     true
-- ) ON CONFLICT (username) DO NOTHING;

-- Grant necessary permissions
GRANT ALL PRIVILEGES ON DATABASE ollamamax TO ollamamax;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ollamamax;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ollamamax;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ollamamax;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ollamamax;
