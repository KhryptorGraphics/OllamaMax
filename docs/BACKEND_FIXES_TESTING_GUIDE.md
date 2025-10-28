# Backend Fixes Testing Guide

Quick reference for testing the security and performance fixes (Issues #006, #007, #009).

## Quick Test Commands

### 1. Test CORS Configuration (ISSUE-006)

```bash
# Test with allowed origin (should succeed)
curl -v -H "Origin: http://localhost:3000" \
  http://localhost:11434/api/v1/health

# Expected: Access-Control-Allow-Origin: http://localhost:3000

# Test with disallowed origin (should use default)
curl -v -H "Origin: http://evil-site.com" \
  http://localhost:11434/api/v1/health

# Expected: Access-Control-Allow-Origin: http://localhost:3000 (or first allowed origin)
```

### 2. Test Rate Limiting (ISSUE-007)

```bash
# Test login rate limit (5 attempts per minute)
for i in {1..10}; do
  echo "Attempt $i:"
  curl -X POST http://localhost:11434/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}' \
    -w "\nHTTP Status: %{http_code}\n\n"
  sleep 1
done

# Expected:
# - Attempts 1-5: 200 OK or 401 Unauthorized (endpoint-dependent)
# - Attempts 6+: 429 Too Many Requests
```

### 3. Test Database Connection Pool (ISSUE-009)

```bash
# Check Prometheus metrics
curl -s http://localhost:11434/metrics | grep -E "ollamamax_database_db_connections"

# Expected output:
# ollamamax_database_db_connections_max 100
# ollamamax_database_db_connections_active <number>
# ollamamax_database_db_connections_idle <number>
```

## Environment Variable Testing

### Test 1: Default Configuration (No Env Vars)

```bash
# Should work with safe defaults
unset CORS_ALLOWED_ORIGINS
unset RATE_LIMIT_LOGIN_REQUESTS
unset OLLAMA_DB_MAX_OPEN_CONNS

# Start server
./ollamamax

# Verify defaults:
# - CORS: localhost:3000, localhost:8080
# - Rate Limit Login: 5/min
# - DB Max Conns: 100
```

### Test 2: Custom Configuration

```bash
# Set custom values
export CORS_ALLOWED_ORIGINS="https://app.example.com,https://api.example.com"
export RATE_LIMIT_LOGIN_REQUESTS=10
export OLLAMA_DB_MAX_OPEN_CONNS=200

# Start server and verify custom values are used
./ollamamax
```

## Load Testing

### Database Connection Pool Load Test

```bash
# Install hey (HTTP load testing tool)
# go install github.com/rakyll/hey@latest

# Run load test (100 concurrent requests, 1000 total)
hey -n 1000 -c 100 http://localhost:11434/api/v1/health

# Monitor metrics during test
watch -n 1 'curl -s http://localhost:11434/metrics | grep db_connections'

# Expected: No connection pool exhaustion with 100 max connections
```

### Rate Limit Load Test

```bash
# Test rate limiting under load
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}' \
  http://localhost:11434/api/v1/auth/login

# Expected: Mix of successful requests and 429 responses
# Check logs for rate limit violations
```

## Integration Tests

### Test with Docker Compose

```bash
# Start services
docker-compose up -d

# Run tests
docker-compose exec api curl http://localhost:11434/api/v1/health

# Check CORS
docker-compose exec api curl -H "Origin: http://localhost:3000" \
  http://localhost:11434/api/v1/health

# Check metrics
docker-compose exec api curl http://localhost:11434/metrics | grep -E "(cors|rate|db)"
```

### Test with Kubernetes

```bash
# Apply configuration
kubectl apply -f k8s/

# Port forward
kubectl port-forward svc/ollamamax-api 11434:11434

# Run tests (same as above)
curl http://localhost:11434/api/v1/health

# Check pod logs
kubectl logs -f deployment/ollamamax-api
```

## Monitoring & Validation

### Check Logs

```bash
# Rate limit violations
tail -f /var/log/ollamamax/api.log | grep "rate_limit"

# CORS issues
tail -f /var/log/ollamamax/api.log | grep -i "cors\|origin"

# Database connection issues
tail -f /var/log/ollamamax/api.log | grep "connection\|pool"
```

### Prometheus Queries

```promql
# Rate limit violations per minute
rate(http_requests_total{status="429"}[1m])

# Database connection pool usage
ollamamax_database_db_connections_active / ollamamax_database_db_connections_max

# CORS-related requests
rate(http_requests_total{path=~".*"}[1m])
```

## Troubleshooting

### CORS Not Working

1. Check environment variable:
   ```bash
   echo $CORS_ALLOWED_ORIGINS
   ```

2. Verify in logs:
   ```bash
   grep -i "cors" /var/log/ollamamax/api.log
   ```

3. Test with curl verbose mode:
   ```bash
   curl -v -H "Origin: YOUR_ORIGIN" http://localhost:11434/api/v1/health
   ```

### Rate Limiting Not Working

1. Check configuration:
   ```bash
   echo $RATE_LIMIT_LOGIN_REQUESTS
   ```

2. Verify middleware is active:
   ```bash
   grep "rate_limit" /var/log/ollamamax/api.log
   ```

3. Test from multiple IPs (rate limiting is per-IP)

### Database Connection Issues

1. Check pool metrics:
   ```bash
   curl http://localhost:11434/metrics | grep db_connections
   ```

2. Check PostgreSQL logs:
   ```bash
   tail -f /var/log/postgresql/postgresql.log
   ```

3. Verify environment variables:
   ```bash
   env | grep OLLAMA_DB
   ```

## Success Criteria

All tests should pass:

- ✅ CORS only allows configured origins
- ✅ Rate limiting blocks after threshold
- ✅ Database connection pool shows 100 max connections
- ✅ No compilation errors
- ✅ All environment variables work correctly
- ✅ Backward compatible (works without new env vars)
- ✅ Prometheus metrics available
- ✅ Logs show expected security events

## Automated Test Script

```bash
#!/bin/bash
# save as test-backend-fixes.sh

echo "Testing Backend Security Fixes..."

# Test 1: CORS
echo "Test 1: CORS Configuration"
RESPONSE=$(curl -s -H "Origin: http://localhost:3000" http://localhost:11434/api/v1/health)
if echo "$RESPONSE" | grep -q "healthy"; then
    echo "✅ CORS test passed"
else
    echo "❌ CORS test failed"
fi

# Test 2: Rate Limiting
echo "Test 2: Rate Limiting"
RATE_LIMIT_HIT=false
for i in {1..10}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
        http://localhost:11434/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username":"test","password":"test"}')

    if [ "$STATUS" = "429" ]; then
        RATE_LIMIT_HIT=true
        break
    fi
    sleep 1
done

if [ "$RATE_LIMIT_HIT" = true ]; then
    echo "✅ Rate limiting test passed"
else
    echo "❌ Rate limiting test failed"
fi

# Test 3: Database Metrics
echo "Test 3: Database Connection Pool"
MAX_CONNS=$(curl -s http://localhost:11434/metrics | grep "ollamamax_database_db_connections_max" | awk '{print $2}')
if [ "$MAX_CONNS" = "100" ]; then
    echo "✅ Database pool test passed"
else
    echo "❌ Database pool test failed (expected 100, got $MAX_CONNS)"
fi

echo "All tests completed!"
```

Run the automated test:
```bash
chmod +x test-backend-fixes.sh
./test-backend-fixes.sh
```
