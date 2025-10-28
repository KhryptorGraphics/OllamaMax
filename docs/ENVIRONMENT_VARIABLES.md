# Environment Variables Documentation

## Critical Security Variables (REQUIRED)

These environment variables **MUST** be set in production for security:

### JWT Configuration

```bash
# JWT Secret Key - REQUIRED (no default)
# Generate with: openssl rand -base64 32
JWT_SECRET=your-strong-random-secret-here
# or
JWT_SECRET_KEY=your-strong-random-secret-here

# Auth Secret Key (optional, falls back to JWT_SECRET)
AUTH_SECRET_KEY=your-strong-random-secret-here
```

### SMTP Configuration

```bash
# SMTP Password - REQUIRED for email functionality (no default)
SMTP_PASSWORD=your-smtp-password-here

# SMTP Configuration (defaults provided)
SMTP_HOST=smtp.gmail.com          # Default: smtp.gmail.com
SMTP_PORT=587                       # Default: 587
SMTP_USER=noreply@giggatek.com     # Default: noreply@giggatek.com
SMTP_FROM=noreply@giggatek.com     # Default: noreply@giggatek.com
```

### Database Credentials

```bash
# PostgreSQL (defaults shown)
POSTGRES_DB=ollamamax              # Default: ollamamax
POSTGRES_USER=ollama               # Default: ollama
POSTGRES_PASSWORD=secure_password  # Default: secure_password (CHANGE IN PRODUCTION)

# Redis (defaults shown)
REDIS_PASSWORD=ollama_redis_pass   # Default: ollama_redis_pass (CHANGE IN PRODUCTION)
```

## CORS Security Configuration

```bash
# CORS Allowed Origins - REQUIRED for security
# Comma-separated list of allowed origins
# Default (development only): http://localhost:3000,http://localhost:8080
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# Enable/disable CORS
CORS_ENABLED=true                  # Default: true
```

## API Server Configuration

```bash
# Server listening configuration
API_LISTEN=0.0.0.0:11434          # Default: 0.0.0.0:11434
API_LISTEN_ADDR=0.0.0.0            # Default: 0.0.0.0
API_PORT=11434                     # Default: 11434

# TLS/SSL Configuration
API_TLS_ENABLED=false              # Default: false
API_CERT_FILE=/path/to/cert.pem    # Required if TLS enabled
API_KEY_FILE=/path/to/key.pem      # Required if TLS enabled

# Request size limit (in bytes)
API_MAX_BODY_SIZE=33554432         # Default: 32MB (32*1024*1024)
```

## Rate Limiting Configuration

```bash
# Rate limiting settings
RATE_LIMIT_ENABLED=true            # Default: true
RATE_LIMIT_REQUESTS=100            # Default: 100 requests
RATE_LIMIT_BURST=10                # Default: 10 burst size
```

## Database Connection Pool

```bash
# PostgreSQL Connection Pool (optimized for production)
OLLAMA_DB_HOST=localhost           # Default: from config
OLLAMA_DB_PORT=5432                # Default: 5432
OLLAMA_DB_NAME=ollamamax           # Default: ollamamax
OLLAMA_DB_USER=ollama              # Default: ollama
OLLAMA_DB_PASSWORD=secure_pass     # Default: (none)
OLLAMA_DB_SSL_MODE=prefer          # Default: prefer

# Connection pool settings (production-optimized)
OLLAMA_DB_MAX_OPEN_CONNS=100       # Default: 100 (increased from 25)
OLLAMA_DB_MAX_IDLE_CONNS=20        # Default: 20 (increased from 5)
OLLAMA_DB_CONN_MAX_LIFETIME=5m     # Default: 5 minutes
```

## Redis Configuration

```bash
# Redis Connection Settings
OLLAMA_REDIS_HOST=localhost        # Default: from config
OLLAMA_REDIS_PORT=6379             # Default: 6379
OLLAMA_REDIS_PASSWORD=pass         # Default: (none)
OLLAMA_REDIS_DB=0                  # Default: 0

# Redis Pool Settings
OLLAMA_REDIS_POOL_SIZE=10          # Default: 10
OLLAMA_REDIS_MIN_IDLE_CONNS=5      # Default: 5
OLLAMA_REDIS_DIAL_TIMEOUT=5s       # Default: 5 seconds
OLLAMA_REDIS_READ_TIMEOUT=3s       # Default: 3 seconds
OLLAMA_REDIS_WRITE_TIMEOUT=3s      # Default: 3 seconds
```

## Authentication Configuration

```bash
# Auth settings
AUTH_ENABLED=true                  # Default: true
AUTH_METHOD=jwt                    # Default: jwt
```

## P2P Networking

```bash
# P2P Configuration
P2P_LISTEN_ADDR=/ip4/0.0.0.0/tcp/0 # Default: /ip4/0.0.0.0/tcp/0
P2P_MAX_CONNECTIONS=100            # Default: 100
```

## Monitoring & Observability

```bash
# Jaeger Tracing
JAEGER_ENDPOINT=http://localhost:14268/api/traces  # Default endpoint

# Grafana (for docker-compose)
GRAFANA_USER=admin                 # Default: admin
GRAFANA_PASSWORD=admin_password    # Default: admin_password (CHANGE IN PRODUCTION)
```

## Production Deployment Checklist

Before deploying to production, ensure these are configured:

- [ ] **JWT_SECRET** - Strong random secret (32+ characters)
- [ ] **SMTP_PASSWORD** - Valid SMTP credentials
- [ ] **POSTGRES_PASSWORD** - Strong database password (16+ characters)
- [ ] **REDIS_PASSWORD** - Strong Redis password (16+ characters)
- [ ] **CORS_ALLOWED_ORIGINS** - Specific allowed origins (no wildcards)
- [ ] **GRAFANA_PASSWORD** - Changed from default
- [ ] **API_TLS_ENABLED=true** - Enable TLS for production
- [ ] **API_CERT_FILE** & **API_KEY_FILE** - Valid SSL certificates

## Security Best Practices

### 1. Secret Generation

Generate strong secrets using:

```bash
# For JWT_SECRET (256-bit)
openssl rand -base64 32

# For database passwords (longer)
openssl rand -base64 48
```

### 2. Secret Management

**DO NOT** hardcode secrets in:
- Source code
- Configuration files committed to git
- Docker images
- Environment defaults

**DO** use:
- Environment variables injected at runtime
- Secret management services (AWS Secrets Manager, HashiCorp Vault)
- `.env` files (excluded from git via `.gitignore`)
- Kubernetes Secrets for K8s deployments

### 3. Example `.env` File

Create `.env` file in project root (NOT committed to git):

```bash
# .env - NEVER commit this file to version control
# Add .env to .gitignore

# Critical Security Secrets
JWT_SECRET=$(openssl rand -base64 32)
SMTP_PASSWORD=your-smtp-app-password-here
POSTGRES_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)

# CORS Security
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://api.yourdomain.com

# Production Settings
API_TLS_ENABLED=true
API_CERT_FILE=/etc/ssl/certs/ollamamax.crt
API_KEY_FILE=/etc/ssl/private/ollamamax.key

# Monitoring
GRAFANA_PASSWORD=$(openssl rand -base64 24)
```

### 4. Docker Compose Usage

```bash
# Load environment variables from .env file
docker-compose --env-file .env up -d

# Or export variables before running
export JWT_SECRET=$(openssl rand -base64 32)
export SMTP_PASSWORD=your-password
docker-compose up -d
```

### 5. Kubernetes Secrets

```bash
# Create Kubernetes secrets
kubectl create secret generic ollamamax-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=smtp-password=your-password \
  --from-literal=postgres-password=$(openssl rand -base64 32) \
  --from-literal=redis-password=$(openssl rand -base64 32)
```

## Migration from Hardcoded Defaults

If upgrading from a version with hardcoded defaults:

### Before (Insecure)

```javascript
// ❌ OLD - Hardcoded in code
this.jwtSecret = process.env.JWT_SECRET || 'ollamamax_secret_key_2024';
password: 'teamrsi123teamrsi123'
```

### After (Secure)

```javascript
// ✅ NEW - Required via environment
this.jwtSecret = process.env.JWT_SECRET;
if (!this.jwtSecret) {
    throw new Error('JWT_SECRET environment variable is required');
}

const smtpPassword = process.env.SMTP_PASSWORD;
if (!smtpPassword) {
    console.warn('SMTP_PASSWORD not set. Email functionality disabled.');
}
```

## Troubleshooting

### Error: "JWT_SECRET environment variable is required"

**Cause**: JWT_SECRET not set in environment.

**Solution**:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
# or add to .env file
```

### Error: "SMTP_PASSWORD not set. Email functionality disabled."

**Cause**: SMTP_PASSWORD not configured.

**Solution**:
```bash
export SMTP_PASSWORD=your-smtp-app-password
# For Gmail, create an App Password: https://support.google.com/accounts/answer/185833
```

### Database connection pool exhausted

**Cause**: Connection pool too small for load.

**Solution**: Increase connection pool size:
```bash
export OLLAMA_DB_MAX_OPEN_CONNS=100
export OLLAMA_DB_MAX_IDLE_CONNS=20
```

## References

- [OWASP Secret Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [12-Factor App Config](https://12factor.net/config)
- [OpenSSL Random Generation](https://www.openssl.org/docs/man1.1.1/man1/rand.html)
