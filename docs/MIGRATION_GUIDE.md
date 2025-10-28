# OllamaMax Migration Guide

**Version**: 1.0.0 → 2.0.0 (Security Hardening Release)
**Date**: 2025-10-27
**Migration Difficulty**: Medium
**Estimated Time**: 30-60 minutes

---

## Overview

This guide provides step-by-step instructions for migrating from the previous version of OllamaMax to the security-hardened 2.0.0 release. The new version removes all hardcoded credentials, enforces strong authentication, and implements production-grade security controls.

### What Changed?

| Component | Change Type | Impact Level |
|-----------|-------------|--------------|
| **JWT Authentication** | Breaking | 🔴 Critical |
| **SMTP Configuration** | Breaking | 🟡 High |
| **Database Ports** | Breaking | 🟡 High |
| **CORS Policy** | Breaking | 🟡 High |
| **Connection Pool** | Enhancement | 🟢 Low |
| **Metrics Registry** | Enhancement | 🟢 Low |
| **Distributed Tracing** | New Feature | 🟢 Low |

---

## Pre-Migration Checklist

Before starting the migration, ensure you have:

- [ ] **Backup of current deployment** (database, configuration, secrets)
- [ ] **Access to production environment variables**
- [ ] **OpenSSL installed** for secret generation
- [ ] **Docker and Docker Compose** updated to latest versions
- [ ] **30-60 minutes of scheduled maintenance window**
- [ ] **Rollback plan** documented and tested

---

## Step 1: Generate Required Secrets

### Generate All Secrets at Once

Run this script to generate all required secrets:

```bash
#!/bin/bash
# save as: generate-secrets.sh

echo "🔐 Generating OllamaMax Security Secrets"
echo "========================================"
echo

# Generate JWT Secret (256-bit)
JWT_SECRET=$(openssl rand -base64 32)
echo "JWT_SECRET generated (32 characters)"

# Generate Auth Secret (256-bit)
AUTH_SECRET=$(openssl rand -base64 32)
echo "AUTH_SECRET generated (32 characters)"

# Generate PostgreSQL Password (384-bit)
POSTGRES_PASSWORD=$(openssl rand -base64 48)
echo "POSTGRES_PASSWORD generated (48 characters)"

# Generate Redis Password (384-bit)
REDIS_PASSWORD=$(openssl rand -base64 48)
echo "REDIS_PASSWORD generated (48 characters)"

# Generate Grafana Admin Password
GRAFANA_PASSWORD=$(openssl rand -base64 24)
echo "GRAFANA_PASSWORD generated (24 characters)"

echo
echo "✅ All secrets generated successfully!"
echo
echo "📋 Copy the following to your .env file:"
echo "========================================"
cat > .env << EOF
# Critical Security Secrets (NEVER COMMIT TO GIT)
# Generated: $(date)

# JWT Authentication
JWT_SECRET=$JWT_SECRET
AUTH_SECRET_KEY=$AUTH_SECRET

# Database Credentials
POSTGRES_DB=ollamamax
POSTGRES_USER=ollama
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# Redis Credentials
REDIS_PASSWORD=$REDIS_PASSWORD

# Monitoring
GRAFANA_PASSWORD=$GRAFANA_PASSWORD

# SMTP Configuration (UPDATE WITH YOUR CREDENTIALS)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=YOUR_SMTP_APP_PASSWORD_HERE

# CORS Security (UPDATE WITH YOUR DOMAINS)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# TLS Configuration (for production)
API_TLS_ENABLED=false
API_CERT_FILE=/etc/ssl/certs/ollamamax.crt
API_KEY_FILE=/etc/ssl/private/ollamamax.key

# Connection Pool Tuning (production defaults)
OLLAMA_DB_MAX_OPEN_CONNS=100
OLLAMA_DB_MAX_IDLE_CONNS=20

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_BURST=10

# Jaeger Tracing
JAEGER_ENDPOINT=http://localhost:14268/api/traces
EOF

echo
echo "✅ .env file created successfully!"
echo
echo "⚠️  IMPORTANT: Update the following in .env:"
echo "   1. SMTP_PASSWORD - Replace with your SMTP app password"
echo "   2. CORS_ALLOWED_ORIGINS - Replace with your production domains"
echo "   3. API_TLS_ENABLED - Set to true for production"
echo "   4. API_CERT_FILE and API_KEY_FILE - Update with certificate paths"
echo
echo "🔒 SECURITY WARNING:"
echo "   - NEVER commit .env to version control"
echo "   - Add .env to .gitignore"
echo "   - Store secrets in a secure password manager"
```

Run the script:
```bash
chmod +x generate-secrets.sh
./generate-secrets.sh
```

### Verify .gitignore

Ensure `.env` is not committed:

```bash
# Check if .env is in .gitignore
grep -q "^\.env$" .gitignore || echo ".env" >> .gitignore

# Verify .env is ignored
git check-ignore .env && echo "✅ .env is gitignored" || echo "❌ WARNING: .env will be committed!"
```

---

## Step 2: Update SMTP Configuration

### Get SMTP App Password

#### For Gmail:
1. Go to [Google Account Security](https://myaccount.google.com/security)
2. Enable 2-Step Verification (if not already enabled)
3. Go to [App Passwords](https://myaccount.google.com/apppasswords)
4. Select "Mail" and "Other (Custom name)"
5. Enter "OllamaMax" as the name
6. Click "Generate"
7. Copy the 16-character password

#### Update .env:
```bash
# Replace with your generated app password
SMTP_PASSWORD=abcd efgh ijkl mnop  # (remove spaces in actual password)
SMTP_USER=noreply@yourdomain.com
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

#### For Other SMTP Providers:

**SendGrid**:
```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your-sendgrid-api-key
```

**Mailgun**:
```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@your-domain.mailgun.org
SMTP_PASSWORD=your-mailgun-smtp-password
```

**AWS SES**:
```bash
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=your-ses-smtp-username
SMTP_PASSWORD=your-ses-smtp-password
```

---

## Step 3: Configure CORS for Production

### Development (Default):
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Production:
```bash
# Replace with your actual production domains
CORS_ALLOWED_ORIGINS=https://app.yourdomain.com,https://www.yourdomain.com,https://api.yourdomain.com
```

### Testing CORS Configuration:

```bash
# Start the server with new config
docker-compose up -d

# Test allowed origin (should succeed)
curl -H "Origin: https://app.yourdomain.com" http://localhost:11434/health -v

# Test blocked origin (should fail)
curl -H "Origin: https://evil.com" http://localhost:11434/health -v
```

Expected output for allowed origin:
```
< HTTP/1.1 200 OK
< Access-Control-Allow-Origin: https://app.yourdomain.com
< Access-Control-Allow-Credentials: true
```

Expected output for blocked origin:
```
< HTTP/1.1 200 OK
(No Access-Control-Allow-Origin header)
```

---

## Step 4: Update Docker Compose Configuration

### Backup Current Configuration

```bash
# Backup current docker-compose.yml
cp docker-compose.yml docker-compose.yml.backup.$(date +%Y%m%d)

# Backup current environment variables
env | grep -E "JWT_SECRET|SMTP|POSTGRES|REDIS|CORS" > env.backup.$(date +%Y%m%d)
```

### Key Changes to docker-compose.yml

#### Before (Vulnerable):
```yaml
postgres:
  ports:
    - "5432:5432"  # ❌ Exposed to external network
  environment:
    POSTGRES_PASSWORD: secure_password  # ❌ Weak default

redis:
  ports:
    - "6379:6379"  # ❌ Exposed to external network
  command: redis-server --requirepass ollama_redis_pass  # ❌ Weak default

api:
  environment:
    JWT_SECRET: ${JWT_SECRET:-ollamamax_secret_key_2024}  # ❌ Weak default
```

#### After (Secured):
```yaml
postgres:
  expose:
    - "5432"  # ✅ Internal network only
  environment:
    POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}  # ✅ Required from environment

redis:
  expose:
    - "6379"  # ✅ Internal network only
  command: redis-server --requirepass ${REDIS_PASSWORD}  # ✅ Required from environment

api:
  environment:
    JWT_SECRET: ${JWT_SECRET}  # ✅ Required from environment
    SMTP_PASSWORD: ${SMTP_PASSWORD}  # ✅ Required from environment
    CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS}  # ✅ Required from environment
```

### Pull Latest Changes

```bash
# Pull latest code
git pull origin main

# Verify docker-compose.yml changes
git diff HEAD~1 docker-compose.yml

# Verify api-server changes
git diff HEAD~1 api-server/auth-system.js

# Verify Go config changes
git diff HEAD~1 internal/config/config.go
```

---

## Step 5: Database Migration

### No Schema Changes Required

Good news! This security update does not require any database schema migrations. Your existing data is compatible.

### Verify Database Connectivity

```bash
# Load new environment variables
source .env

# Start only PostgreSQL
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
sleep 10

# Test connection with new password
docker exec -it ollamamax-postgres psql -U ollama -d ollamamax -c "SELECT version();"

# If connection succeeds, you're good to go!
```

### Backup Database Before Migration

```bash
# Create backup directory
mkdir -p backups/$(date +%Y%m%d)

# Backup PostgreSQL
docker exec ollamamax-postgres pg_dump -U ollama ollamamax > backups/$(date +%Y%m%d)/ollamamax-postgres.sql

# Backup Redis
docker exec ollamamax-redis redis-cli --pass ${REDIS_PASSWORD} SAVE
docker cp ollamamax-redis:/data/dump.rdb backups/$(date +%Y%m%d)/redis-dump.rdb

echo "✅ Database backup completed"
```

---

## Step 6: Deploy New Version

### Option A: Rolling Update (Zero Downtime)

```bash
# Load environment variables
source .env

# Build new images
docker-compose build

# Start services one by one
docker-compose up -d postgres redis
sleep 10

docker-compose up -d ollama-primary ollama-worker-1 ollama-worker-2
sleep 10

docker-compose up -d ollamamax-api
sleep 5

docker-compose up -d ollamamax-web nginx
sleep 5

# Verify all services are healthy
docker-compose ps
```

### Option B: Full Restart (Maintenance Window)

```bash
# Load environment variables
source .env

# Stop all services
docker-compose down

# Pull latest images
docker-compose pull

# Start all services with new configuration
docker-compose up -d

# Wait for services to be ready
sleep 30

# Verify health
curl http://localhost:11434/health
```

---

## Step 7: Post-Migration Verification

### Automated Verification Script

```bash
#!/bin/bash
# save as: verify-migration.sh

echo "🔍 OllamaMax Migration Verification"
echo "===================================="
echo

# Check 1: JWT_SECRET is set and strong
if [ ${#JWT_SECRET} -ge 32 ]; then
    echo "✅ JWT_SECRET is set and strong (${#JWT_SECRET} characters)"
else
    echo "❌ JWT_SECRET is too weak or not set"
    exit 1
fi

# Check 2: SMTP_PASSWORD is configured
if [ -n "$SMTP_PASSWORD" ]; then
    echo "✅ SMTP_PASSWORD is configured"
else
    echo "⚠️  SMTP_PASSWORD not set (email functionality disabled)"
fi

# Check 3: Database passwords changed from defaults
if [ "$POSTGRES_PASSWORD" != "secure_password" ]; then
    echo "✅ PostgreSQL password changed from default"
else
    echo "❌ PostgreSQL still using default password"
    exit 1
fi

if [ "$REDIS_PASSWORD" != "ollama_redis_pass" ]; then
    echo "✅ Redis password changed from default"
else
    echo "❌ Redis still using default password"
    exit 1
fi

# Check 4: CORS configured properly
if echo "$CORS_ALLOWED_ORIGINS" | grep -q "localhost"; then
    echo "⚠️  WARNING: localhost in CORS_ALLOWED_ORIGINS (development mode)"
else
    echo "✅ CORS_ALLOWED_ORIGINS configured for production"
fi

# Check 5: Database ports not exposed
if docker-compose ps | grep -E "5432.*0.0.0.0|6379.*0.0.0.0"; then
    echo "❌ Database ports are exposed externally"
    exit 1
else
    echo "✅ Database ports are internal only"
fi

# Check 6: Services are running
if docker-compose ps | grep -q "Up"; then
    echo "✅ Docker services are running"
else
    echo "❌ Some Docker services are not running"
    exit 1
fi

# Check 7: API is responding
if curl -f http://localhost:11434/health > /dev/null 2>&1; then
    echo "✅ API health check passed"
else
    echo "❌ API health check failed"
    exit 1
fi

# Check 8: Authentication requires JWT
if curl -f http://localhost:11434/api/v1/users/profile > /dev/null 2>&1; then
    echo "❌ Authentication not enforced"
    exit 1
else
    echo "✅ Authentication is enforced"
fi

# Check 9: No hardcoded secrets in codebase
if grep -r "teamrsi123teamrsi123" . 2>/dev/null | grep -v ".backup" | grep -v "docs/"; then
    echo "❌ Found hardcoded SMTP password in codebase"
    exit 1
else
    echo "✅ No hardcoded SMTP passwords found"
fi

if grep -r "ollamamax_secret_key_2024" . 2>/dev/null | grep -v ".backup" | grep -v "docs/"; then
    echo "❌ Found hardcoded JWT secret in codebase"
    exit 1
else
    echo "✅ No hardcoded JWT secrets found"
fi

# Check 10: Metrics are available
if curl -s http://localhost:11434/metrics | grep -q "ollamamax_database_db_connections_max"; then
    echo "✅ Prometheus metrics are exposed"
else
    echo "⚠️  Metrics endpoint not responding correctly"
fi

echo
echo "===================================="
echo "✅ Migration verification completed successfully!"
echo
echo "📊 Connection Pool Stats:"
curl -s http://localhost:11434/metrics | grep ollamamax_database_db_connections_max
echo
echo "🔐 Security Status:"
echo "   - JWT_SECRET: Strong (${#JWT_SECRET} chars)"
echo "   - SMTP_PASSWORD: $([ -n "$SMTP_PASSWORD" ] && echo "Configured" || echo "Not set")"
echo "   - POSTGRES_PASSWORD: Changed from default"
echo "   - REDIS_PASSWORD: Changed from default"
echo "   - CORS: $(echo $CORS_ALLOWED_ORIGINS | grep -q localhost && echo "Development" || echo "Production")"
echo
echo "🎉 Your OllamaMax deployment is secure and ready!"
```

Run verification:
```bash
chmod +x verify-migration.sh
source .env
./verify-migration.sh
```

---

## Step 8: Update Kubernetes Deployments (Optional)

If you're deploying to Kubernetes, update your secrets:

### Create Kubernetes Secrets

```bash
# Create secrets from .env file
kubectl create secret generic ollamamax-secrets \
  --from-literal=jwt-secret=$JWT_SECRET \
  --from-literal=auth-secret=$AUTH_SECRET_KEY \
  --from-literal=smtp-password=$SMTP_PASSWORD \
  --from-literal=postgres-password=$POSTGRES_PASSWORD \
  --from-literal=redis-password=$REDIS_PASSWORD \
  --from-literal=grafana-password=$GRAFANA_PASSWORD \
  --namespace ollamamax

# Create ConfigMap for non-sensitive config
kubectl create configmap ollamamax-config \
  --from-literal=cors-allowed-origins=$CORS_ALLOWED_ORIGINS \
  --from-literal=smtp-host=$SMTP_HOST \
  --from-literal=smtp-port=$SMTP_PORT \
  --from-literal=smtp-user=$SMTP_USER \
  --namespace ollamamax

# Verify secrets created
kubectl get secrets -n ollamamax ollamamax-secrets
kubectl describe configmap -n ollamamax ollamamax-config
```

### Update Deployments

```bash
# Apply updated Kubernetes manifests
kubectl apply -f k8s/

# Restart deployments to pick up new secrets
kubectl rollout restart deployment/ollamamax-api -n ollamamax
kubectl rollout restart deployment/ollamamax-web -n ollamamax

# Watch rollout status
kubectl rollout status deployment/ollamamax-api -n ollamamax
```

---

## Rollback Procedure

If something goes wrong during migration:

### Docker Compose Rollback

```bash
# Stop current deployment
docker-compose down

# Restore backup configuration
cp docker-compose.yml.backup.$(date +%Y%m%d) docker-compose.yml

# Restore environment variables
source env.backup.$(date +%Y%m%d)

# Start with old configuration
docker-compose up -d

# Verify rollback successful
docker-compose ps
curl http://localhost:11434/health
```

### Database Rollback

```bash
# Stop services
docker-compose down

# Restore PostgreSQL
docker-compose up -d postgres
sleep 10
cat backups/$(date +%Y%m%d)/ollamamax-postgres.sql | docker exec -i ollamamax-postgres psql -U ollama -d ollamamax

# Restore Redis
docker cp backups/$(date +%Y%m%d)/redis-dump.rdb ollamamax-redis:/data/dump.rdb
docker-compose restart redis

# Restart all services
docker-compose up -d
```

### Kubernetes Rollback

```bash
# Rollback to previous deployment
kubectl rollout undo deployment/ollamamax-api -n ollamamax
kubectl rollout undo deployment/ollamamax-web -n ollamamax

# Restore old secrets (if needed)
kubectl delete secret ollamamax-secrets -n ollamamax
kubectl create secret generic ollamamax-secrets --from-env-file=env.backup.$(date +%Y%m%d) -n ollamamax

# Restart with old secrets
kubectl rollout restart deployment/ollamamax-api -n ollamamax
```

---

## Troubleshooting

### Issue 1: Application fails to start

**Error**: `panic: JWT_SECRET_KEY or JWT_SECRET environment variable is required`

**Solution**:
```bash
# Verify JWT_SECRET is set
echo $JWT_SECRET

# If not set, generate and export
export JWT_SECRET=$(openssl rand -base64 32)

# Restart application
docker-compose restart ollamamax-api
```

### Issue 2: Email verification not working

**Error**: `SMTP_PASSWORD not set. Email functionality will use mock transporter.`

**Solution**:
```bash
# Set SMTP_PASSWORD
export SMTP_PASSWORD=your-smtp-app-password

# Restart API server
docker-compose restart ollamamax-api

# Test email functionality
docker-compose logs ollamamax-api | grep "Test email sent"
```

### Issue 3: Database connection refused

**Error**: `failed to connect to PostgreSQL: connection refused`

**Solution**:
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Verify password is correct
docker exec -it ollamamax-postgres psql -U ollama -d ollamamax -c "SELECT 1;"

# If password mismatch, update POSTGRES_PASSWORD
export POSTGRES_PASSWORD=correct-password
docker-compose restart postgres
```

### Issue 4: CORS errors in browser

**Error**: `Access to XMLHttpRequest has been blocked by CORS policy`

**Solution**:
```bash
# Add your frontend domain to CORS_ALLOWED_ORIGINS
export CORS_ALLOWED_ORIGINS="https://yourdomain.com,$CORS_ALLOWED_ORIGINS"

# Restart API server
docker-compose restart ollamamax-api

# Verify CORS headers
curl -H "Origin: https://yourdomain.com" http://localhost:11434/health -v
```

### Issue 5: Connection pool exhaustion

**Error**: `database: connection pool exhausted`

**Solution**:
```bash
# Increase connection pool size
export OLLAMA_DB_MAX_OPEN_CONNS=200
export OLLAMA_DB_MAX_IDLE_CONNS=50

# Restart API server
docker-compose restart ollamamax-api

# Monitor pool metrics
curl http://localhost:11434/metrics | grep db_connections
```

---

## Performance Validation

After migration, validate performance improvements:

### Connection Pool Performance

```bash
# Before: MaxOpenConns=25, MaxIdleConns=5
# After:  MaxOpenConns=100, MaxIdleConns=20

# Check current pool configuration
curl -s http://localhost:11434/metrics | grep ollamamax_database_db_connections_max
# Expected: ollamamax_database_db_connections_max 100

# Monitor pool utilization under load
ab -n 10000 -c 100 http://localhost:11434/health

# Check if pool was exhausted (should be 0 or very low)
curl -s http://localhost:11434/metrics | grep ollamamax_database_db_connections_wait_count
```

### Rate Limiting

```bash
# Test rate limiting is active
for i in {1..150}; do
    curl -s -o /dev/null -w "%{http_code}\n" http://localhost:11434/health
done | grep "429" | wc -l

# Expected: >0 (some requests should be rate limited)
```

---

## Production Deployment Checklist

Before going live in production:

- [ ] All secrets generated with strong random values
- [ ] `.env` file created and **NOT** committed to git
- [ ] SMTP credentials tested and working
- [ ] CORS configured with production domains only
- [ ] TLS/SSL certificates obtained and configured
- [ ] Database backups verified and tested
- [ ] Monitoring dashboards configured (Grafana)
- [ ] Alerts configured (Prometheus AlertManager)
- [ ] Load testing completed successfully
- [ ] Security scan completed (OWASP ZAP)
- [ ] Disaster recovery plan documented
- [ ] Team trained on new security requirements
- [ ] Rollback procedure tested
- [ ] Post-migration verification passed

---

## Next Steps

After successful migration:

1. **Monitor Security Metrics**: Watch authentication failures, rate limit violations
2. **Enable Compression**: Install `github.com/gin-contrib/gzip` (see SECURITY_FIXES_APPLIED.md)
3. **Implement WAF**: Follow `docs/WAF_INTEGRATION_GUIDE.md`
4. **Schedule Secret Rotation**: Rotate JWT_SECRET, database passwords quarterly
5. **Review Audit Logs**: Check for suspicious activity

---

## Support

For migration issues:

- **Documentation**: See `docs/ENVIRONMENT_VARIABLES.md` for variable reference
- **Security**: See `docs/SECURITY_FIXES_APPLIED.md` for security details
- **Deployment**: See `docs/DEPLOYMENT_SECURITY_CHECKLIST.md` for deployment steps
- **Known Issues**: See `KNOWN_ISSUES.md` for tracked issues

---

## Conclusion

Congratulations! You've successfully migrated to the security-hardened version of OllamaMax. Your deployment now follows security best practices with:

- ✅ Strong authentication (no weak defaults)
- ✅ Secure credential management (environment-based)
- ✅ Internal database networking (no exposed ports)
- ✅ Restrictive CORS policy
- ✅ Production-grade connection pooling
- ✅ Comprehensive monitoring and tracing

Your system is now production-ready and significantly more secure. 🎉
