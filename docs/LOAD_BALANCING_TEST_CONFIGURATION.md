# Load Balancing Test Configuration

## Overview

This document explains the configuration and rationale for the load balancing test script (`scripts/test-load-balancing.sh`) and its interaction with Nginx proxied endpoints.

## Problem Statement

**Original Issue:** The load-balancer test relied on the `X-Served-By` header to identify which backend server handled each request, but Nginx did not set this header, yielding `unknown` backend attribution and inaccurate distribution analysis.

**Root Cause:** The test was sending requests to the `/health` endpoint, which is self-served by Nginx and does not proxy to any upstream backend. Since the request never reaches an upstream, the `X-Served-By` header (which is set to `$upstream_addr`) was not added.

## Solution

### Nginx Configuration

The Nginx configuration (`nginx/nginx-production.conf`) has been updated to add the `X-Served-By` header to all proxied locations:

```nginx
# API Endpoints (line 186)
location /api/ {
    proxy_pass http://api_backend;
    add_header X-Served-By $upstream_addr always;
    # ... other proxy settings
}

# Ollama Endpoints (line 254)
location /ollama/ {
    proxy_pass http://ollama_backend/;
    add_header X-Served-By $upstream_addr always;
    # ... other proxy settings
}

# Frontend (line 290)
location / {
    proxy_pass http://web_backend;
    add_header X-Served-By $upstream_addr always;
    # ... other proxy settings
}

# Self-served health endpoint (NO X-Served-By header)
location /health {
    return 200 "healthy\n";
    # No proxy, no X-Served-By header
}
```

### Test Script Updates

The test script (`scripts/test-load-balancing.sh`) has been updated to:

1. **Use a configurable proxied endpoint:**
   - Default: `/api/health` (proxied to `api_backend`)
   - Configurable via second argument: `LB_PATH="${2:-/api/health}"`

2. **Trim CRLF from headers:**
   - `tr -d '\r'` ensures header parsing works correctly across different systems

3. **Validate endpoint availability:**
   - Checks if the proxied endpoint is responding before running distribution tests
   - Falls back gracefully if the endpoint is unavailable

4. **Use non-proxied endpoints appropriately:**
   - `/nginx-health` is used for liveness checks and failover testing
   - Proxied endpoints are used for distribution analysis

## Endpoint Classification

### Proxied Endpoints (Include X-Served-By)

These endpoints proxy to upstream backends and include the `X-Served-By` header:

| Endpoint Pattern | Upstream Backend | Example Value |
|-----------------|------------------|---------------|
| `/api/*` | `api_backend` | `api:8080` |
| `/ollama/*` | `ollama_backend` | `ollama-node-1:11434`, `ollama-node-2:11434`, `ollama-node-3:11434` |
| `/` (root) | `web_backend` | `web:3000` |
| `/ws` | `api_backend` | `api:8080` |

### Non-Proxied Endpoints (NO X-Served-By)

These endpoints are self-served by Nginx and do NOT proxy to upstreams:

| Endpoint | Purpose | Use Case |
|----------|---------|----------|
| `/health` | Simple health check | K8s liveness probes |
| `/nginx-health` | Nginx-specific health | Internal monitoring |
| `/stub_status` | Nginx metrics | Performance monitoring |

## Usage Examples

### Default Configuration (API Health)

```bash
# Uses /api/health endpoint
bash scripts/test-load-balancing.sh http://localhost
```

### Custom Proxied Endpoint

```bash
# Test Ollama backend distribution
bash scripts/test-load-balancing.sh http://localhost /ollama/api/tags

# Test web backend distribution
bash scripts/test-load-balancing.sh http://localhost /
```

### CI/CD Pipeline Integration

```bash
# In GitHub Actions or CI pipeline
export LB_URL="http://localhost"
export LB_PATH="/api/health"

# Run test with explicit parameters
bash scripts/test-load-balancing.sh "${LB_URL}" "${LB_PATH}"
```

### Docker Compose Environment

```bash
# Start services
docker compose -f docker-compose.prod.yml up -d

# Wait for services to be ready
sleep 30

# Run load balancing test
bash scripts/test-load-balancing.sh http://localhost /api/health
```

## Expected Output

### Successful Test Run

```
=== Load Balancing Validation ===
Testing endpoint: http://localhost/api/health

=== Phase 1: Load Balancer Configuration ===
[INFO] Checking nginx configuration...
[PASS] Load balancing algorithm: least_conn
[INFO] Backend servers configured: 3

=== Phase 2: Testing Load Distribution ===
[INFO] Sending 1000 requests...
..........
[INFO] Request distribution:
[INFO]   api:8080: 1000 requests (100%)
[PASS] Load distribution is balanced

=== Phase 3: Testing Health Checks ===
[INFO] Testing nginx health check endpoint (non-proxied)...
[PASS] Health check endpoint responding
[INFO] Testing proxied API endpoint...
[PASS] Proxied API endpoint responding
```

### Output with Multiple Ollama Backends

```
[INFO] Request distribution:
[INFO]   ollama-node-1:11434: 334 requests (33%)
[INFO]   ollama-node-2:11434: 333 requests (33%)
[INFO]   ollama-node-3:11434: 333 requests (33%)
[PASS] Load distribution is balanced
```

## Verification Steps

1. **Verify X-Served-By header is present:**
   ```bash
   curl -I http://localhost/api/health | grep -i X-Served-By
   # Expected: X-Served-By: api:8080
   ```

2. **Verify header is absent on non-proxied endpoints:**
   ```bash
   curl -I http://localhost/health | grep -i X-Served-By
   # Expected: (no output - header not present)
   ```

3. **Test distribution manually:**
   ```bash
   for i in {1..10}; do
     curl -s -I http://localhost/api/health | grep X-Served-By
   done
   # Expected: Multiple backend addresses if using clustered upstreams
   ```

4. **Run full test suite:**
   ```bash
   bash scripts/test-load-balancing.sh http://localhost /api/health
   # Expected: Distribution report showing multiple backends
   ```

## Troubleshooting

### Issue: "unknown" backend in distribution report

**Cause:** Test is using a non-proxied endpoint or X-Served-By header is missing.

**Solution:**
1. Verify you're using a proxied endpoint (`/api/`, `/ollama/`, or `/`)
2. Check Nginx config has `add_header X-Served-By $upstream_addr always;`
3. Restart Nginx: `docker compose restart nginx`

### Issue: All requests go to single backend

**Cause:** Only one backend is healthy or load balancing algorithm issue.

**Solution:**
1. Check all backend services are running: `docker compose ps`
2. Verify health checks: `docker compose logs nginx`
3. Check upstream configuration in `nginx/nginx-production.conf`

### Issue: High variance in distribution

**Cause:** Uneven backend load or connection pooling effects.

**Solution:**
1. Increase `NUM_REQUESTS` in script for more accurate sampling
2. Check backend health and performance metrics
3. Verify `least_conn` algorithm is configured correctly

## References

- Nginx configuration: `nginx/nginx-production.conf`
- Test script: `scripts/test-load-balancing.sh`
- Deployment procedures: `docs/DEPLOYMENT_PROCEDURES.md`
- Docker Compose: `docker-compose.prod.yml`

## CI/CD Integration Notes

When invoking this script from CI/CD pipelines:

1. **Set explicit parameters:**
   ```yaml
   - name: Test Load Balancing
     run: bash scripts/test-load-balancing.sh http://localhost /api/health
   ```

2. **Ensure services are ready:**
   ```yaml
   - name: Wait for services
     run: |
       sleep 30
       curl --retry 5 --retry-delay 5 http://localhost/nginx-health
   ```

3. **Parse JSON report:**
   ```yaml
   - name: Parse results
     run: |
       cat deployment-results/loadbalancing-*.json
       # Check "balanced": true
   ```

4. **Archive reports:**
   ```yaml
   - name: Upload report
     uses: actions/upload-artifact@v3
     with:
       name: load-balancing-report
       path: deployment-results/loadbalancing-*.json
   ```

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2025 | Initial documentation |
| 1.1 | 2025 | Updated to use proxied endpoints for X-Served-By header |

---

**Note:** Always test with services fully started and healthy before running distribution analysis.
