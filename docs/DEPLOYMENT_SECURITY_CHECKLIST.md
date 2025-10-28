# Deployment Security Checklist

## Pre-Deployment Security Checklist

This checklist ensures all critical security fixes from the comprehensive security review have been implemented before production deployment.

### 🔒 Critical Security Issues (MUST FIX BEFORE DEPLOY)

#### ✅ ISSUE-001: Hardcoded SMTP Credentials (CVSS 7.5) - **FIXED**

**Status**: Fixed in commit [hash]

**Changes Made**:
- [x] Removed hardcoded password `teamrsi123teamrsi123` from `api-server/auth-system.js` (lines 70, 86, 99, 113)
- [x] Removed default from `docker-compose.yml` line 27
- [x] Added environment variable requirement with validation
- [x] Added warning when SMTP_PASSWORD not set

**Verification**:
```bash
# Should fail without SMTP_PASSWORD
grep -r "teamrsi123teamrsi123" api-server/ docker-compose.yml
# Expected: No matches found

# Verify environment variable requirement
node -e "process.env.JWT_SECRET='test'; require('./api-server/auth-system.js')"
# Expected: Warning if SMTP_PASSWORD not set
```

**Required Action**:
```bash
export SMTP_PASSWORD=your-secure-password-here
```

#### ✅ ISSUE-002: Weak JWT Secret Defaults (CVSS 8.1) - **FIXED**

**Status**: Fixed in commit [hash]

**Changes Made**:
- [x] Removed default `ollamamax_secret_key_2024` from `api-server/auth-system.js` line 16
- [x] Removed default `your-secret-key-change-this` from `internal/config/config.go` lines 84, 94
- [x] Added panic() if JWT_SECRET not provided (Go)
- [x] Added throw Error() if JWT_SECRET not provided (Node.js)
- [x] Removed weak default from `docker-compose.yml` line 17

**Verification**:
```bash
# Should fail without JWT_SECRET
go run cmd/ollamamax/main.go
# Expected: panic: JWT_SECRET_KEY or JWT_SECRET environment variable is required

# Verify no weak defaults in codebase
grep -r "ollamamax_secret_key_2024\|your-secret-key-change-this" .
# Expected: No matches (except in docs)
```

**Required Action**:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
```

#### ✅ ISSUE-003: Exposed Database Ports (CVSS 7.5) - **FIXED**

**Status**: Fixed in commit [hash]

**Changes Made**:
- [x] Removed PostgreSQL port mapping `5432:5432` from `docker-compose.yml` line 80
- [x] Removed Redis port mapping `6379:6379` from `docker-compose.yml` line 102
- [x] Changed to `expose` directive (internal network only)

**Verification**:
```bash
# After docker-compose up, verify ports not exposed
docker-compose ps
# PostgreSQL and Redis should show only exposed ports, not mapped ports

netstat -tuln | grep -E "5432|6379"
# Expected: No external listeners on 5432 or 6379
```

**Required Action**: None - fixed in docker-compose.yml

#### ✅ ISSUE-006: Permissive CORS (CVSS 5.3) - **FIXED**

**Status**: Fixed in commit [hash]

**Changes Made**:
- [x] Removed `AllowedOrigins: []string{"*"}` from `internal/config/config.go` line 113
- [x] Changed to environment-based configuration: `getEnvListOrDefault("CORS_ALLOWED_ORIGINS", ...)`
- [x] Default to localhost only: `http://localhost:3000,http://localhost:8080`
- [x] Restricted headers to specific list: `Content-Type`, `Authorization`, `X-Request-ID`
- [x] Enabled credentials: `AllowCredentials: true`

**Verification**:
```bash
# Verify wildcard removed
grep -r 'AllowedOrigins.*"\*"' internal/config/
# Expected: No matches

# Test CORS headers (after starting server)
curl -H "Origin: http://evil.com" http://localhost:11434/health -v
# Expected: No Access-Control-Allow-Origin header for unauthorized origin
```

**Required Action**:
```bash
export CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

#### ⚠️ ISSUE-007: Missing Rate Limiting on Auth Endpoints (CVSS 5.3) - **PARTIAL**

**Status**: Implemented globally, needs authentication-specific limits

**Changes Made**:
- [x] Global rate limiting exists (`middleware.go` line 73-104)
- [ ] TODO: Add stricter rate limits specifically for `/api/v1/auth/*` endpoints
- [ ] TODO: Implement account lockout after N failed attempts

**Verification**:
```bash
# Test rate limiting
for i in {1..150}; do curl http://localhost:11434/health -w "%{http_code}\n" -o /dev/null; done
# Expected: 429 Too Many Requests after ~100 requests
```

**Required Action**: Implement auth-specific rate limiting (see TODO)

#### ⚠️ ISSUE-008: No Compression (CVSS 4.3) - **DOCUMENTED**

**Status**: Documented, implementation pending

**Changes Made**:
- [x] Added TODO comments in `pkg/api/server.go` line 236-238
- [x] Documented compression middleware placeholder in `pkg/api/middleware.go` line 185-194
- [ ] TODO: Install `github.com/gin-contrib/gzip` dependency
- [ ] TODO: Uncomment compression middleware

**Verification**:
```bash
# Test compression (after implementing)
curl -H "Accept-Encoding: gzip" http://localhost:11434/api/v1/models -v
# Expected: Content-Encoding: gzip header
```

**Required Action**:
```bash
# Add dependency
go get github.com/gin-contrib/gzip

# Uncomment in pkg/api/server.go
router.Use(gzip.Gzip(gzip.DefaultCompression))
```

#### ✅ ISSUE-009: Database Connection Pool Too Small (CVSS 5.3) - **FIXED**

**Status**: Fixed in commit [hash]

**Changes Made**:
- [x] Increased `MaxOpenConns` from 25 → 100 in `pkg/database/manager.go` line 86
- [x] Increased `MaxIdleConns` from 5 → 20 in `pkg/database/manager.go` line 90
- [x] Added PERFORMANCE comments documenting changes

**Verification**:
```bash
# Monitor connection pool usage under load
curl http://localhost:11434/metrics | grep db_connections
# Expected: db_connections_max = 100
```

**Required Action**: None - auto-configured to production values

### 📋 Pre-Deployment Checklist

Run through this checklist before deploying to production:

#### Environment Variables

- [ ] **JWT_SECRET** set and strong (32+ characters)
  ```bash
  test ${#JWT_SECRET} -ge 32 && echo "✅ OK" || echo "❌ TOO SHORT"
  ```

- [ ] **SMTP_PASSWORD** configured for email
  ```bash
  test -n "$SMTP_PASSWORD" && echo "✅ OK" || echo "❌ NOT SET"
  ```

- [ ] **POSTGRES_PASSWORD** changed from default
  ```bash
  test "$POSTGRES_PASSWORD" != "secure_password" && echo "✅ OK" || echo "❌ USING DEFAULT"
  ```

- [ ] **REDIS_PASSWORD** changed from default
  ```bash
  test "$REDIS_PASSWORD" != "ollama_redis_pass" && echo "✅ OK" || echo "❌ USING DEFAULT"
  ```

- [ ] **CORS_ALLOWED_ORIGINS** set to production domains
  ```bash
  echo $CORS_ALLOWED_ORIGINS | grep -q "localhost" && echo "⚠️  WARNING: localhost in production" || echo "✅ OK"
  ```

#### Docker Security

- [ ] Database ports NOT exposed externally
  ```bash
  docker-compose config | grep -A2 "postgres:" | grep "ports:" && echo "❌ EXPOSED" || echo "✅ OK"
  ```

- [ ] Redis port NOT exposed externally
  ```bash
  docker-compose config | grep -A2 "redis:" | grep "ports:" && echo "❌ EXPOSED" || echo "✅ OK"
  ```

#### Application Security

- [ ] TLS/SSL enabled in production
  ```bash
  test "$API_TLS_ENABLED" = "true" && echo "✅ OK" || echo "⚠️  TLS NOT ENABLED"
  ```

- [ ] Valid SSL certificates present
  ```bash
  test -f "$API_CERT_FILE" && test -f "$API_KEY_FILE" && echo "✅ OK" || echo "❌ CERTIFICATES MISSING"
  ```

- [ ] No hardcoded secrets in codebase
  ```bash
  grep -r "password.*=.*['\"]" --include="*.js" --include="*.go" . | grep -v "process.env" && echo "❌ FOUND HARDCODED PASSWORDS" || echo "✅ OK"
  ```

#### Rate Limiting

- [ ] Rate limiting enabled
  ```bash
  test "$RATE_LIMIT_ENABLED" = "true" && echo "✅ OK" || echo "❌ DISABLED"
  ```

- [ ] Rate limits configured appropriately
  ```bash
  test "$RATE_LIMIT_REQUESTS" -le 1000 && echo "✅ OK" || echo "⚠️  VERY HIGH LIMIT"
  ```

#### Database & Performance

- [ ] Connection pool sized for expected load
  ```bash
  test "$OLLAMA_DB_MAX_OPEN_CONNS" -ge 50 && echo "✅ OK" || echo "⚠️  SMALL POOL"
  ```

- [ ] Compression enabled (when implemented)
  ```bash
  curl -H "Accept-Encoding: gzip" http://localhost:11434/health -I | grep -q "Content-Encoding: gzip" && echo "✅ OK" || echo "⚠️  NO COMPRESSION"
  ```

### 🚀 Deployment Steps

1. **Generate Secrets**
   ```bash
   # Generate all required secrets
   export JWT_SECRET=$(openssl rand -base64 32)
   export SMTP_PASSWORD=your-smtp-app-password
   export POSTGRES_PASSWORD=$(openssl rand -base64 48)
   export REDIS_PASSWORD=$(openssl rand -base64 48)
   export GRAFANA_PASSWORD=$(openssl rand -base64 24)

   # Save to .env file (DO NOT COMMIT)
   cat > .env << EOF
   JWT_SECRET=$JWT_SECRET
   SMTP_PASSWORD=$SMTP_PASSWORD
   POSTGRES_PASSWORD=$POSTGRES_PASSWORD
   REDIS_PASSWORD=$REDIS_PASSWORD
   GRAFANA_PASSWORD=$GRAFANA_PASSWORD
   CORS_ALLOWED_ORIGINS=https://yourdomain.com
   API_TLS_ENABLED=true
   API_CERT_FILE=/etc/ssl/certs/ollamamax.crt
   API_KEY_FILE=/etc/ssl/private/ollamamax.key
   EOF
   ```

2. **Verify Configuration**
   ```bash
   # Run all verification checks
   ./scripts/verify-security-config.sh
   ```

3. **Deploy with Secrets**
   ```bash
   # Docker Compose
   docker-compose --env-file .env up -d

   # Kubernetes
   kubectl create secret generic ollamamax-secrets --from-env-file=.env
   kubectl apply -f k8s/
   ```

4. **Post-Deployment Verification**
   ```bash
   # Verify no exposed ports
   nmap -p 5432,6379 your-server-ip
   # Expected: ports closed/filtered

   # Verify CORS restrictions
   curl -H "Origin: http://evil.com" https://your-api/health -v
   # Expected: No CORS headers

   # Verify rate limiting
   ab -n 200 -c 10 https://your-api/health
   # Expected: Some requests return 429

   # Verify TLS
   openssl s_client -connect your-api:443 -tls1_2
   # Expected: Valid certificate chain
   ```

### 📊 Security Metrics

Track these metrics post-deployment:

- **Authentication Failures**: Monitor failed login attempts
  ```bash
  curl http://localhost:11434/metrics | grep auth_failures_total
  ```

- **Rate Limit Hits**: Track rate limit violations
  ```bash
  curl http://localhost:11434/metrics | grep rate_limit_exceeded_total
  ```

- **Database Connection Pool**: Monitor pool utilization
  ```bash
  curl http://localhost:11434/metrics | grep db_connections
  ```

- **Response Times**: Ensure P95 < 500ms
  ```bash
  curl http://localhost:11434/metrics | grep http_request_duration_seconds
  ```

### 🔍 Post-Deployment Audit

Within 24 hours of deployment:

- [ ] Run security scanner (e.g., OWASP ZAP)
- [ ] Review audit logs for anomalies
- [ ] Verify all secrets rotated from defaults
- [ ] Confirm monitoring alerts working
- [ ] Test disaster recovery procedures
- [ ] Document incident response procedures

### 📞 Security Incident Response

If a security issue is discovered post-deployment:

1. **Immediate**: Rotate affected secrets
2. **Within 1 hour**: Assess impact and containment
3. **Within 4 hours**: Deploy fix or implement workaround
4. **Within 24 hours**: Root cause analysis and documentation
5. **Within 1 week**: Implement preventive measures

### 🔗 References

- [ENVIRONMENT_VARIABLES.md](./ENVIRONMENT_VARIABLES.md) - Complete environment variable documentation
- [SECURITY_COMPLIANCE_REVIEW.md](./SECURITY_COMPLIANCE_REVIEW.md) - Full security assessment
- [KNOWN_ISSUES.md](../KNOWN_ISSUES.md) - Issue tracking register
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
