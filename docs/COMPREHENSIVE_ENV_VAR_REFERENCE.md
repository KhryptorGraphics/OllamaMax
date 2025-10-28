# Comprehensive Environment Variables Reference

**Version**: 2.0.0
**Last Updated**: 2025-10-27
**Target Audience**: DevOps, System Administrators, Developers

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Critical Security Variables](#critical-security-variables)
3. [Database Configuration](#database-configuration)
4. [API Server Configuration](#api-server-configuration)
5. [Authentication & Authorization](#authentication--authorization)
6. [Email & SMTP Configuration](#email--smtp-configuration)
7. [CORS & Security Headers](#cors--security-headers)
8. [Rate Limiting](#rate-limiting)
9. [Monitoring & Observability](#monitoring--observability)
10. [Performance Tuning](#performance-tuning)
11. [P2P Networking](#p2p-networking)
12. [Development vs Production](#development-vs-production)
13. [Complete Examples](#complete-examples)
14. [Validation & Testing](#validation--testing)

---

## Quick Start

### Minimal Production Configuration

```bash
# Generate secrets
export JWT_SECRET=$(openssl rand -base64 32)
export SMTP_PASSWORD=your-smtp-app-password
export POSTGRES_PASSWORD=$(openssl rand -base64 48)
export REDIS_PASSWORD=$(openssl rand -base64 48)

# Configure CORS
export CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Enable TLS
export API_TLS_ENABLED=true
export API_CERT_FILE=/etc/ssl/certs/ollamamax.crt
export API_KEY_FILE=/etc/ssl/private/ollamamax.key

# Deploy
docker-compose up -d
```

### Minimal Development Configuration

```bash
# Generate secrets (still required)
export JWT_SECRET=$(openssl rand -base64 32)

# Optional for local dev
export SMTP_PASSWORD=not-required-for-dev

# Deploy
docker-compose up -d
```

---

## Critical Security Variables

### JWT_SECRET / JWT_SECRET_KEY

**Required**: ✅ Yes (application won't start without it)
**Type**: String (base64-encoded)
**Minimum Length**: 32 characters (256-bit security)
**Default**: None (no default for security)

**Purpose**: Secret key used to sign and verify JSON Web Tokens for authentication.

**Generation**:
```bash
# Recommended: 256-bit (32 bytes base64-encoded)
export JWT_SECRET=$(openssl rand -base64 32)

# Extra strong: 384-bit (48 bytes)
export JWT_SECRET=$(openssl rand -base64 48)

# Maximum: 512-bit (64 bytes)
export JWT_SECRET=$(openssl rand -base64 64)
```

**Used By**:
- `api-server/auth-system.js` - Node.js authentication
- `internal/config/config.go` - Go backend authentication

**Security Notes**:
- ⚠️ **NEVER commit to version control**
- ⚠️ **Rotate quarterly** (every 90 days)
- ⚠️ **Use different secrets** for dev/staging/prod
- ✅ **Store in secret management system** (AWS Secrets Manager, Vault)

**Validation**:
```bash
# Check if set
test -n "$JWT_SECRET" && echo "✅ Set" || echo "❌ Not set"

# Check strength
test ${#JWT_SECRET} -ge 32 && echo "✅ Strong" || echo "❌ Too weak"

# Verify uniqueness (not default)
grep -r "ollamamax_secret_key_2024" . && echo "❌ Using old default!" || echo "✅ Unique"
```

**Error if Not Set**:
```
panic: JWT_SECRET_KEY or JWT_SECRET environment variable is required for security
```

---

### AUTH_SECRET_KEY

**Required**: ⚠️ Optional (falls back to JWT_SECRET)
**Type**: String (base64-encoded)
**Default**: Falls back to `JWT_SECRET`

**Purpose**: Separate secret for authentication middleware (allows independent rotation).

**Generation**:
```bash
export AUTH_SECRET_KEY=$(openssl rand -base64 32)
```

**Use Case**: Independent secret rotation
- Rotate `JWT_SECRET` without breaking active sessions
- Use different secrets for JWT signing vs session management

**Best Practice**:
```bash
# Production: Use separate secrets for easier rotation
export JWT_SECRET=$(openssl rand -base64 32)
export AUTH_SECRET_KEY=$(openssl rand -base64 32)

# Development: Simpler to use one secret
export JWT_SECRET=$(openssl rand -base64 32)
# AUTH_SECRET_KEY automatically uses JWT_SECRET
```

---

### SMTP_PASSWORD

**Required**: ⚠️ Recommended (email functionality disabled without it)
**Type**: String (plain text password or app password)
**Default**: None (graceful degradation to mock transporter)

**Purpose**: Password for SMTP server authentication (email verification, password resets).

**Generation**:
For Gmail (most common):
1. Enable 2FA: https://myaccount.google.com/security
2. Generate App Password: https://myaccount.google.com/apppasswords
3. Copy 16-character password

```bash
export SMTP_PASSWORD=abcdefghijklmnop  # Your app password (no spaces)
```

**Used By**:
- `api-server/auth-system.js` - Email verification system

**Behavior Without Password**:
```javascript
if (!smtpPassword) {
    console.warn('⚠️  SMTP_PASSWORD not set. Email functionality will use mock transporter.');
    // Emails logged to console instead of sent
}
```

**Related Variables**:
```bash
export SMTP_HOST=smtp.gmail.com           # Default: smtp.gmail.com
export SMTP_PORT=587                       # Default: 587 (STARTTLS)
export SMTP_USER=noreply@yourdomain.com    # Default: noreply@giggatek.com
export SMTP_FROM=noreply@yourdomain.com    # Email "From" address
```

**Validation**:
```bash
# Test SMTP configuration
node -e "
const nodemailer = require('nodemailer');
const transporter = nodemailer.createTransport({
  host: process.env.SMTP_HOST || 'smtp.gmail.com',
  port: process.env.SMTP_PORT || 587,
  auth: {
    user: process.env.SMTP_USER,
    pass: process.env.SMTP_PASSWORD
  }
});
transporter.verify().then(console.log).catch(console.error);
"
```

---

### POSTGRES_PASSWORD

**Required**: ⚠️ Must change from default
**Type**: String (alphanumeric + special characters recommended)
**Default**: `secure_password` (⚠️ insecure default for development only)

**Purpose**: PostgreSQL database password.

**Generation**:
```bash
# Strong password (48 bytes = 384-bit)
export POSTGRES_PASSWORD=$(openssl rand -base64 48)

# Or use a password manager
export POSTGRES_PASSWORD=$(pwgen -s 32 1)
```

**Related Variables**:
```bash
export POSTGRES_DB=ollamamax        # Database name (default: ollamamax)
export POSTGRES_USER=ollama          # Database user (default: ollama)
```

**Docker Compose Usage**:
```yaml
postgres:
  image: postgres:15-alpine
  environment:
    POSTGRES_DB: ${POSTGRES_DB:-ollamamax}
    POSTGRES_USER: ${POSTGRES_USER:-ollama}
    POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}  # No default - must be set!
```

**Security Notes**:
- ⚠️ **Change from default** in production
- ✅ **Minimum 16 characters**
- ✅ **Include uppercase, lowercase, numbers, special chars**
- ⚠️ **Rotate every 90 days**

**Validation**:
```bash
# Check if changed from default
test "$POSTGRES_PASSWORD" != "secure_password" && echo "✅ Secure" || echo "❌ Using default!"

# Test connection
docker exec ollamamax-postgres psql -U ollama -d ollamamax -c "SELECT version();"
```

---

### REDIS_PASSWORD

**Required**: ⚠️ Must change from default
**Type**: String (alphanumeric recommended)
**Default**: `ollama_redis_pass` (⚠️ insecure default for development only)

**Purpose**: Redis password for authentication.

**Generation**:
```bash
export REDIS_PASSWORD=$(openssl rand -base64 48)
```

**Docker Compose Usage**:
```yaml
redis:
  image: redis:7-alpine
  command: redis-server --requirepass ${REDIS_PASSWORD}
```

**Application Usage** (Go):
```go
// pkg/database/manager.go
rdb := redis.NewClient(&redis.Options{
    Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
    Password: config.RedisPassword,  // From OLLAMA_REDIS_PASSWORD
    DB:       config.RedisDB,
})
```

**Validation**:
```bash
# Check if changed from default
test "$REDIS_PASSWORD" != "ollama_redis_pass" && echo "✅ Secure" || echo "❌ Using default!"

# Test connection
docker exec ollamamax-redis redis-cli -a ${REDIS_PASSWORD} PING
# Expected: PONG
```

---

## Database Configuration

### PostgreSQL Connection

#### OLLAMA_DB_HOST
**Type**: String (hostname or IP)
**Default**: `localhost` (or from config file)
**Example**: `postgres`, `db.internal`, `10.0.1.5`

#### OLLAMA_DB_PORT
**Type**: Integer
**Default**: `5432`
**Range**: 1-65535

#### OLLAMA_DB_NAME
**Type**: String
**Default**: `ollamamax`

#### OLLAMA_DB_USER
**Type**: String
**Default**: `ollama`

#### OLLAMA_DB_PASSWORD
**Type**: String
**Default**: None (must be set, see `POSTGRES_PASSWORD`)

#### OLLAMA_DB_SSL_MODE
**Type**: Enum
**Default**: `prefer`
**Options**: `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`

**Recommendations**:
- **Development**: `disable` (faster)
- **Production**: `require` or `verify-full` (secure)

```bash
# Production
export OLLAMA_DB_SSL_MODE=require

# With certificate verification
export OLLAMA_DB_SSL_MODE=verify-full
export PGSSLROOTCERT=/path/to/ca-cert.pem
```

### PostgreSQL Connection Pool

#### OLLAMA_DB_MAX_OPEN_CONNS
**Type**: Integer
**Default**: `100` (production-optimized, was 25 before)
**Range**: 10-1000
**Recommendation**:
- Development: 25
- Production (1K-5K RPS): 100
- High-load production (10K+ RPS): 200-500

**Performance Impact**:
| Connections | Max RPS | Memory | Recommendation |
|-------------|---------|--------|----------------|
| 10 | ~200 | ~50MB | Embedded/edge |
| 25 | ~500 | ~120MB | Development |
| 100 | ~10,000 | ~500MB | **Production (default)** |
| 200 | ~20,000 | ~1GB | High-traffic |
| 500 | ~50,000 | ~2.5GB | Extreme load |

**Configuration**:
```bash
# High-load production
export OLLAMA_DB_MAX_OPEN_CONNS=200

# Resource-constrained environment
export OLLAMA_DB_MAX_OPEN_CONNS=25
```

#### OLLAMA_DB_MAX_IDLE_CONNS
**Type**: Integer
**Default**: `20` (was 5 before)
**Range**: 2-100
**Recommendation**: 20-25% of `MAX_OPEN_CONNS`

**Purpose**: Maintain ready connections to reduce latency.

```bash
# Production (keep 20% of max as idle)
export OLLAMA_DB_MAX_IDLE_CONNS=20

# High-load (keep 25% idle)
export OLLAMA_DB_MAX_IDLE_CONNS=50
```

#### OLLAMA_DB_CONN_MAX_LIFETIME
**Type**: Duration
**Default**: `5m` (5 minutes)
**Range**: `1m` - `1h`

**Purpose**: Maximum time a connection can be reused before recycling.

```bash
export OLLAMA_DB_CONN_MAX_LIFETIME=5m   # Default
export OLLAMA_DB_CONN_MAX_LIFETIME=10m  # Longer lifetime (less overhead)
export OLLAMA_DB_CONN_MAX_LIFETIME=1m   # Shorter (more recycling)
```

### Redis Configuration

#### OLLAMA_REDIS_HOST
**Type**: String
**Default**: `localhost`
**Example**: `redis`, `cache.internal`, `10.0.1.6`

#### OLLAMA_REDIS_PORT
**Type**: Integer
**Default**: `6379`

#### OLLAMA_REDIS_PASSWORD
**Type**: String
**Default**: None (see `REDIS_PASSWORD`)

#### OLLAMA_REDIS_DB
**Type**: Integer
**Default**: `0`
**Range**: 0-15 (Redis supports 16 databases by default)

#### OLLAMA_REDIS_POOL_SIZE
**Type**: Integer
**Default**: `10`
**Range**: 5-100

#### OLLAMA_REDIS_MIN_IDLE_CONNS
**Type**: Integer
**Default**: `5`
**Range**: 1-50

#### OLLAMA_REDIS_DIAL_TIMEOUT
**Type**: Duration
**Default**: `5s`

#### OLLAMA_REDIS_READ_TIMEOUT
**Type**: Duration
**Default**: `3s`

#### OLLAMA_REDIS_WRITE_TIMEOUT
**Type**: Duration
**Default**: `3s`

**Complete Redis Configuration**:
```bash
# Connection
export OLLAMA_REDIS_HOST=redis-cluster
export OLLAMA_REDIS_PORT=6379
export OLLAMA_REDIS_PASSWORD=$(openssl rand -base64 32)
export OLLAMA_REDIS_DB=0

# Pool settings
export OLLAMA_REDIS_POOL_SIZE=20
export OLLAMA_REDIS_MIN_IDLE_CONNS=10

# Timeouts
export OLLAMA_REDIS_DIAL_TIMEOUT=5s
export OLLAMA_REDIS_READ_TIMEOUT=3s
export OLLAMA_REDIS_WRITE_TIMEOUT=3s
```

---

## API Server Configuration

### API_LISTEN
**Type**: String (address:port)
**Default**: `0.0.0.0:11434`

**Examples**:
```bash
# Listen on all interfaces (default)
export API_LISTEN=0.0.0.0:11434

# Listen on specific interface
export API_LISTEN=192.168.1.100:11434

# Listen on localhost only (development)
export API_LISTEN=127.0.0.1:11434
```

### API_LISTEN_ADDR
**Type**: String (IP address)
**Default**: `0.0.0.0`

### API_PORT
**Type**: Integer
**Default**: `11434`
**Range**: 1024-65535

### API_TLS_ENABLED
**Type**: Boolean
**Default**: `false`
**Production**: `true` (required)

```bash
# Production
export API_TLS_ENABLED=true
```

### API_CERT_FILE
**Type**: String (file path)
**Required If**: `API_TLS_ENABLED=true`
**Example**: `/etc/ssl/certs/ollamamax.crt`

### API_KEY_FILE
**Type**: String (file path)
**Required If**: `API_TLS_ENABLED=true`
**Example**: `/etc/ssl/private/ollamamax.key`

**Complete TLS Configuration**:
```bash
# Enable TLS
export API_TLS_ENABLED=true

# Certificate paths
export API_CERT_FILE=/etc/letsencrypt/live/api.yourdomain.com/fullchain.pem
export API_KEY_FILE=/etc/letsencrypt/live/api.yourdomain.com/privkey.pem

# Optional: Verify certificate
openssl x509 -in $API_CERT_FILE -text -noout
```

### API_MAX_BODY_SIZE
**Type**: Integer (bytes)
**Default**: `33554432` (32MB)

**Examples**:
```bash
# 64MB (for larger payloads)
export API_MAX_BODY_SIZE=67108864

# 128MB (for file uploads)
export API_MAX_BODY_SIZE=134217728

# Convert from MB to bytes
export API_MAX_BODY_SIZE=$((128 * 1024 * 1024))  # 128MB
```

---

## Authentication & Authorization

### AUTH_ENABLED
**Type**: Boolean
**Default**: `true`

```bash
# Disable authentication (DEVELOPMENT ONLY)
export AUTH_ENABLED=false

# Enable authentication (production)
export AUTH_ENABLED=true
```

### AUTH_METHOD
**Type**: String
**Default**: `jwt`
**Options**: `jwt`, `oauth`, `saml` (future support)

---

## Email & SMTP Configuration

### SMTP_HOST
**Type**: String (hostname)
**Default**: `smtp.gmail.com`

**Common Providers**:
```bash
# Gmail
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587

# SendGrid
export SMTP_HOST=smtp.sendgrid.net
export SMTP_PORT=587

# Mailgun
export SMTP_HOST=smtp.mailgun.org
export SMTP_PORT=587

# AWS SES (us-east-1)
export SMTP_HOST=email-smtp.us-east-1.amazonaws.com
export SMTP_PORT=587

# Office 365
export SMTP_HOST=smtp.office365.com
export SMTP_PORT=587

# Custom SMTP server
export SMTP_HOST=mail.yourdomain.com
export SMTP_PORT=587
```

### SMTP_PORT
**Type**: Integer
**Default**: `587` (STARTTLS)

**Common Ports**:
- `25` - Plain SMTP (insecure, often blocked)
- `587` - STARTTLS (recommended)
- `465` - SSL/TLS (older standard)
- `2525` - Alternative STARTTLS (some providers)

### SMTP_USER
**Type**: String (email address)
**Default**: `noreply@giggatek.com`

### SMTP_FROM
**Type**: String (email address)
**Default**: Same as `SMTP_USER`

**Example**:
```bash
export SMTP_USER=noreply@yourdomain.com
export SMTP_FROM="OllamaMax <noreply@yourdomain.com>"  # With display name
```

---

## CORS & Security Headers

### CORS_ENABLED
**Type**: Boolean
**Default**: `true`

```bash
# Disable CORS (if behind API gateway that handles CORS)
export CORS_ENABLED=false
```

### CORS_ALLOWED_ORIGINS
**Type**: Comma-separated list of URLs
**Default**: `http://localhost:3000,http://localhost:8080` (development only)
**Production**: **MUST** be set to actual domains

**Examples**:
```bash
# Development (default)
export CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Production (single domain)
export CORS_ALLOWED_ORIGINS=https://app.yourdomain.com

# Production (multiple domains)
export CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com,https://admin.yourdomain.com

# Production (multiple environments)
export CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://staging.yourdomain.com,https://dev.yourdomain.com
```

**Security Notes**:
- ⚠️ **NEVER use `*` in production**
- ⚠️ **Do not include `http://` URLs in production**
- ✅ **Use HTTPS only** for production
- ✅ **Include all subdomains** that need access

**Validation**:
```bash
# Test allowed origin
curl -H "Origin: https://app.yourdomain.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS http://localhost:11434/health -v

# Should see: Access-Control-Allow-Origin: https://app.yourdomain.com

# Test blocked origin
curl -H "Origin: https://evil.com" \
     -H "Access-Control-Request-Method: GET" \
     -X OPTIONS http://localhost:11434/health -v

# Should NOT see: Access-Control-Allow-Origin header
```

---

## Rate Limiting

### RATE_LIMIT_ENABLED
**Type**: Boolean
**Default**: `true`

### RATE_LIMIT_REQUESTS
**Type**: Integer (requests per minute)
**Default**: `100`
**Range**: 10-10000

**Recommendations**:
```bash
# Development (relaxed)
export RATE_LIMIT_REQUESTS=1000

# Production (standard)
export RATE_LIMIT_REQUESTS=100

# Production (strict, authenticated API)
export RATE_LIMIT_REQUESTS=60

# High-traffic API (with authentication)
export RATE_LIMIT_REQUESTS=500
```

### RATE_LIMIT_BURST
**Type**: Integer
**Default**: `10`
**Range**: 1-100

**Purpose**: Allow short bursts above the rate limit.

```bash
# Standard (10 requests can burst)
export RATE_LIMIT_BURST=10

# Strict (no bursting)
export RATE_LIMIT_BURST=0

# Permissive (allow larger bursts)
export RATE_LIMIT_BURST=50
```

**How It Works**:
```
Limit: 100 req/min with burst 10
- Sustained: 100 requests per minute
- Burst: Can handle up to 110 requests in short period
- After burst: Must wait for tokens to refill
```

---

## Monitoring & Observability

### JAEGER_ENDPOINT
**Type**: String (URL)
**Default**: `http://localhost:14268/api/traces`

**Examples**:
```bash
# Local Jaeger
export JAEGER_ENDPOINT=http://localhost:14268/api/traces

# Remote Jaeger
export JAEGER_ENDPOINT=http://jaeger-collector.monitoring:14268/api/traces

# Kubernetes
export JAEGER_ENDPOINT=http://jaeger-collector.jaeger-system.svc.cluster.local:14268/api/traces

# Disable tracing (graceful degradation)
unset JAEGER_ENDPOINT
```

### GRAFANA_USER
**Type**: String
**Default**: `admin`

### GRAFANA_PASSWORD
**Type**: String
**Default**: `admin_password` (⚠️ MUST change for production)

**Generation**:
```bash
export GRAFANA_PASSWORD=$(openssl rand -base64 24)
```

---

## Performance Tuning

### Connection Pool Summary

**Quick Reference**:
```bash
# Small deployment (< 100 concurrent users)
export OLLAMA_DB_MAX_OPEN_CONNS=25
export OLLAMA_DB_MAX_IDLE_CONNS=5

# Medium deployment (100-1000 concurrent users)
export OLLAMA_DB_MAX_OPEN_CONNS=100   # Default
export OLLAMA_DB_MAX_IDLE_CONNS=20    # Default

# Large deployment (1000-10000 concurrent users)
export OLLAMA_DB_MAX_OPEN_CONNS=200
export OLLAMA_DB_MAX_IDLE_CONNS=50

# Extra-large deployment (10000+ concurrent users)
export OLLAMA_DB_MAX_OPEN_CONNS=500
export OLLAMA_DB_MAX_IDLE_CONNS=100
```

---

## P2P Networking

### P2P_LISTEN_ADDR
**Type**: String (libp2p multiaddr format)
**Default**: `/ip4/0.0.0.0/tcp/0`

**Examples**:
```bash
# Listen on all IPv4 interfaces, random port
export P2P_LISTEN_ADDR=/ip4/0.0.0.0/tcp/0

# Specific port
export P2P_LISTEN_ADDR=/ip4/0.0.0.0/tcp/4001

# IPv6
export P2P_LISTEN_ADDR=/ip6/::/tcp/4001

# Localhost only
export P2P_LISTEN_ADDR=/ip4/127.0.0.1/tcp/4001
```

### P2P_MAX_CONNECTIONS
**Type**: Integer
**Default**: `100`
**Range**: 10-1000

---

## Development vs Production

### Development Environment

```bash
# .env.development
JWT_SECRET=$(openssl rand -base64 32)

# Relaxed settings for development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080,http://localhost:5173
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_BURST=100

# Database (smaller pools)
OLLAMA_DB_MAX_OPEN_CONNS=25
OLLAMA_DB_MAX_IDLE_CONNS=5

# Optional SMTP (use mock transporter)
# SMTP_PASSWORD not required - emails logged to console

# TLS disabled for local development
API_TLS_ENABLED=false
```

### Production Environment

```bash
# .env.production
JWT_SECRET=$(openssl rand -base64 48)
SMTP_PASSWORD=your-smtp-app-password
POSTGRES_PASSWORD=$(openssl rand -base64 48)
REDIS_PASSWORD=$(openssl rand -base64 48)

# Strict security
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_BURST=10

# Database (production pools)
OLLAMA_DB_MAX_OPEN_CONNS=100
OLLAMA_DB_MAX_IDLE_CONNS=20

# TLS required
API_TLS_ENABLED=true
API_CERT_FILE=/etc/letsencrypt/live/api.yourdomain.com/fullchain.pem
API_KEY_FILE=/etc/letsencrypt/live/api.yourdomain.com/privkey.pem

# SMTP configured
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com

# Monitoring
JAEGER_ENDPOINT=http://jaeger-collector:14268/api/traces
GRAFANA_PASSWORD=$(openssl rand -base64 24)
```

---

## Complete Examples

### Complete Production .env

```bash
# ====================
# OllamaMax Production Environment Configuration
# Generated: 2025-10-27
# NEVER COMMIT THIS FILE TO VERSION CONTROL
# ====================

# --------------------
# Critical Security
# --------------------
JWT_SECRET=XYZ123...base64encoded...ABC789=
AUTH_SECRET_KEY=ABC456...base64encoded...XYZ012=
SMTP_PASSWORD=your-gmail-app-password

# --------------------
# Database Credentials
# --------------------
POSTGRES_DB=ollamamax
POSTGRES_USER=ollama
POSTGRES_PASSWORD=DEF789...base64encoded...GHI345=
REDIS_PASSWORD=GHI012...base64encoded...JKL678=

# --------------------
# Database Connection Pool
# --------------------
OLLAMA_DB_HOST=postgres
OLLAMA_DB_PORT=5432
OLLAMA_DB_NAME=ollamamax
OLLAMA_DB_USER=ollama
OLLAMA_DB_PASSWORD=${POSTGRES_PASSWORD}
OLLAMA_DB_SSL_MODE=require
OLLAMA_DB_MAX_OPEN_CONNS=100
OLLAMA_DB_MAX_IDLE_CONNS=20
OLLAMA_DB_CONN_MAX_LIFETIME=5m

# --------------------
# Redis Configuration
# --------------------
OLLAMA_REDIS_HOST=redis
OLLAMA_REDIS_PORT=6379
OLLAMA_REDIS_PASSWORD=${REDIS_PASSWORD}
OLLAMA_REDIS_DB=0
OLLAMA_REDIS_POOL_SIZE=10
OLLAMA_REDIS_MIN_IDLE_CONNS=5

# --------------------
# API Server
# --------------------
API_LISTEN=0.0.0.0:11434
API_TLS_ENABLED=true
API_CERT_FILE=/etc/letsencrypt/live/api.yourdomain.com/fullchain.pem
API_KEY_FILE=/etc/letsencrypt/live/api.yourdomain.com/privkey.pem
API_MAX_BODY_SIZE=33554432

# --------------------
# Authentication
# --------------------
AUTH_ENABLED=true
AUTH_METHOD=jwt

# --------------------
# SMTP Configuration
# --------------------
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_FROM=OllamaMax <noreply@yourdomain.com>

# --------------------
# CORS Security
# --------------------
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com

# --------------------
# Rate Limiting
# --------------------
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_BURST=10

# --------------------
# Monitoring
# --------------------
JAEGER_ENDPOINT=http://jaeger-collector:14268/api/traces
GRAFANA_USER=admin
GRAFANA_PASSWORD=JKL345...base64encoded...MNO901=

# --------------------
# P2P Networking
# --------------------
P2P_LISTEN_ADDR=/ip4/0.0.0.0/tcp/0
P2P_MAX_CONNECTIONS=100
```

---

## Validation & Testing

### Complete Validation Script

```bash
#!/bin/bash
# validate-environment.sh

echo "🔍 OllamaMax Environment Validation"
echo "===================================="
echo

# Critical secrets
test -n "$JWT_SECRET" || { echo "❌ JWT_SECRET not set"; exit 1; }
test ${#JWT_SECRET} -ge 32 || { echo "❌ JWT_SECRET too weak"; exit 1; }
echo "✅ JWT_SECRET: Strong (${#JWT_SECRET} chars)"

test -n "$SMTP_PASSWORD" && echo "✅ SMTP_PASSWORD: Configured" || echo "⚠️  SMTP_PASSWORD: Not set (email disabled)"

# Database passwords
test "$POSTGRES_PASSWORD" != "secure_password" || { echo "❌ POSTGRES_PASSWORD using default"; exit 1; }
echo "✅ POSTGRES_PASSWORD: Changed from default"

test "$REDIS_PASSWORD" != "ollama_redis_pass" || { echo "❌ REDIS_PASSWORD using default"; exit 1; }
echo "✅ REDIS_PASSWORD: Changed from default"

# CORS configuration
echo "$CORS_ALLOWED_ORIGINS" | grep -q "localhost" && echo "⚠️  CORS: Contains localhost" || echo "✅ CORS: Production-ready"

# TLS configuration (production)
if [ "$API_TLS_ENABLED" = "true" ]; then
    test -f "$API_CERT_FILE" || { echo "❌ TLS certificate not found"; exit 1; }
    test -f "$API_KEY_FILE" || { echo "❌ TLS key not found"; exit 1; }
    echo "✅ TLS: Enabled and configured"
else
    echo "⚠️  TLS: Disabled (development mode)"
fi

# Connection pool validation
test "${OLLAMA_DB_MAX_OPEN_CONNS:-100}" -ge 50 && echo "✅ Connection pool: Production-ready" || echo "⚠️  Connection pool: Small"

echo
echo "===================================="
echo "✅ Environment validation completed!"
```

### Run Validation

```bash
chmod +x validate-environment.sh
source .env
./validate-environment.sh
```

---

## Troubleshooting

See [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md#troubleshooting) for complete troubleshooting guide.

---

## References

- [SECURITY_FIXES_APPLIED.md](./SECURITY_FIXES_APPLIED.md) - Security changes
- [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) - Migration instructions
- [DEPLOYMENT_SECURITY_CHECKLIST.md](./DEPLOYMENT_SECURITY_CHECKLIST.md) - Pre-deployment validation
- [FIXES_IMPACT_ANALYSIS.md](./FIXES_IMPACT_ANALYSIS.md) - Impact assessment
