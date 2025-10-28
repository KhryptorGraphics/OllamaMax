# WAF Integration Guide - ModSecurity with Nginx

## Overview

This guide explains how to integrate ModSecurity WAF (Web Application Firewall) with the OllamaMax Nginx deployment for production security compliance as specified in `production-security.yaml`.

## Current Status

**NOT IMPLEMENTED**: The standard `nginx:alpine` image does not include ModSecurity. To enable WAF protection, you must build a custom Nginx image with the ModSecurity module.

## Implementation Options

### Option 1: Use Pre-Built ModSecurity Image (Recommended for Quick Setup)

Use an existing Nginx image with ModSecurity pre-installed:

```yaml
# docker-compose.yml or docker-compose.prod.yml
services:
  nginx:
    image: owasp/modsecurity-crs:nginx-alpine  # Pre-built with ModSecurity
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx-production.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/modsecurity/modsecurity.conf:/etc/modsecurity.d/modsecurity.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
```

### Option 2: Build Custom Nginx Image with ModSecurity

**Step 1: Create Dockerfile**

Create `nginx/Dockerfile.modsecurity`:

```dockerfile
FROM nginx:alpine

# Install build dependencies
RUN apk add --no-cache \
    git \
    g++ \
    make \
    libtool \
    automake \
    autoconf \
    pcre-dev \
    zlib-dev \
    libxml2-dev \
    curl-dev \
    yajl-dev \
    geoip-dev \
    lmdb-dev

# Build ModSecurity
WORKDIR /tmp
RUN git clone --depth 1 -b v3/master --single-branch https://github.com/SpiderLabs/ModSecurity && \
    cd ModSecurity && \
    git submodule init && \
    git submodule update && \
    ./build.sh && \
    ./configure && \
    make && \
    make install

# Build Nginx ModSecurity connector
RUN git clone --depth 1 https://github.com/SpiderLabs/ModSecurity-nginx.git && \
    cd /tmp && \
    wget http://nginx.org/download/nginx-1.24.0.tar.gz && \
    tar -xzvf nginx-1.24.0.tar.gz && \
    cd nginx-1.24.0 && \
    ./configure --with-compat --add-dynamic-module=../ModSecurity-nginx && \
    make modules && \
    cp objs/ngx_http_modsecurity_module.so /etc/nginx/modules/

# Install OWASP CRS (Core Rule Set)
RUN cd /usr/local && \
    git clone https://github.com/coreruleset/coreruleset && \
    cd coreruleset && \
    mv crs-setup.conf.example crs-setup.conf

# Cleanup
RUN apk del git g++ make libtool automake autoconf && \
    rm -rf /tmp/*

# Load ModSecurity module
RUN echo 'load_module modules/ngx_http_modsecurity_module.so;' > /etc/nginx/modules-enabled/50-mod-security.conf

CMD ["nginx", "-g", "daemon off;"]
```

**Step 2: Build the Custom Image**

```bash
cd nginx
docker build -f Dockerfile.modsecurity -t ollamamax-nginx-modsecurity:latest .
```

**Step 3: Update Docker Compose**

```yaml
# docker-compose.yml
services:
  nginx:
    build:
      context: ./nginx
      dockerfile: Dockerfile.modsecurity
    # ... rest of configuration
```

### Step 3: ModSecurity Configuration

Create `nginx/modsecurity/modsecurity.conf`:

```nginx
# ModSecurity Core Configuration

# Enable ModSecurity and Detection Only mode
SecRuleEngine On
SecRequestBodyAccess On
SecResponseBodyAccess Off

# Logging
SecAuditEngine RelevantOnly
SecAuditLogRelevantStatus "^(?:5|4(?!04))"
SecAuditLogParts ABIJDEFHZ
SecAuditLog /var/log/nginx/modsec_audit.log

# File Upload
SecRequestBodyLimit 13107200
SecRequestBodyNoFilesLimit 131072
SecRequestBodyInMemoryLimit 131072
SecRequestBodyLimitAction Reject

# Temporary Directory
SecTmpDir /tmp/
SecDataDir /tmp/

# Debugging
SecDebugLog /var/log/nginx/modsec_debug.log
SecDebugLogLevel 0

# Include OWASP CRS
Include /usr/local/coreruleset/crs-setup.conf
Include /usr/local/coreruleset/rules/*.conf
```

### Step 4: Update nginx-production.conf

Add ModSecurity directives to `nginx/nginx-production.conf`:

```nginx
# At the top of the file (after worker_processes)
load_module modules/ngx_http_modsecurity_module.so;

# In the http block
http {
    # Enable ModSecurity
    modsecurity on;
    modsecurity_rules_file /etc/modsecurity.d/modsecurity.conf;

    # ... rest of configuration
}
```

### Step 5: Update validation script

Extend `scripts/validate-security.sh`:

```bash
# Phase 10: WAF Detection and Testing
echo -e "\n${BLUE}=== Phase 10: WAF Detection ===${NC}"

# Check if ModSecurity is enabled
if docker compose exec -T nginx nginx -V 2>&1 | grep -q "modsecurity"; then
    log_success "ModSecurity module detected"
    add_result "WAF Module" "pass" "ModSecurity installed"

    # Test WAF with injection payload
    log_info "Testing WAF protection..."

    # SQL Injection test (should be blocked)
    if curl -f "http://localhost/api/users?id=1' OR '1'='1" > /dev/null 2>&1; then
        log_error "WAF did not block SQL injection attempt"
        add_result "WAF SQL Protection" "fail" "Injection not blocked"
    else
        log_success "WAF blocked SQL injection attempt"
        add_result "WAF SQL Protection" "pass" "Injection blocked"
    fi

    # XSS test (should be blocked)
    if curl -f "http://localhost/api/search?q=<script>alert('xss')</script>" > /dev/null 2>&1; then
        log_error "WAF did not block XSS attempt"
        add_result "WAF XSS Protection" "fail" "XSS not blocked"
    else
        log_success "WAF blocked XSS attempt"
        add_result "WAF XSS Protection" "pass" "XSS blocked"
    fi

else
    log_error "ModSecurity module not found"
    add_result "WAF Module" "fail" "ModSecurity not installed"
fi
```

## Testing WAF Rules

### Test SQL Injection Protection

```bash
# This should be blocked (403 Forbidden)
curl -i "http://localhost/api/users?id=1' OR '1'='1"
```

### Test XSS Protection

```bash
# This should be blocked (403 Forbidden)
curl -i "http://localhost/api/search?q=<script>alert('xss')</script>"
```

### Test Path Traversal Protection

```bash
# This should be blocked (403 Forbidden)
curl -i "http://localhost/api/files?path=../../../../etc/passwd"
```

## Performance Considerations

ModSecurity adds overhead:
- **Latency**: +5-20ms per request
- **Memory**: +50-100MB per worker
- **CPU**: +10-20% under load

**Mitigation strategies:**
1. Use `SecRuleEngine DetectionOnly` during initial deployment
2. Tune rules to reduce false positives
3. Exclude static assets from inspection
4. Use caching for WAF decisions
5. Scale horizontally if needed

## Production Deployment Checklist

- [ ] Build custom Nginx image with ModSecurity
- [ ] Configure OWASP CRS rules
- [ ] Test WAF with injection payloads
- [ ] Review and tune rules for false positives
- [ ] Set up WAF logging and monitoring
- [ ] Create alerts for blocked attacks
- [ ] Document WAF bypass procedures for emergencies
- [ ] Train team on WAF management
- [ ] Schedule regular rule updates
- [ ] Integrate WAF logs with SIEM

## References

- [ModSecurity Documentation](https://github.com/SpiderLabs/ModSecurity)
- [OWASP ModSecurity CRS](https://coreruleset.org/)
- [Nginx ModSecurity Connector](https://github.com/SpiderLabs/ModSecurity-nginx)
- [production-security.yaml](../production-security.yaml) - Security requirements

## Next Steps

1. **Immediate**: Document that WAF is not yet implemented
2. **Short-term**: Build custom Nginx image with ModSecurity
3. **Medium-term**: Deploy and tune OWASP CRS rules
4. **Long-term**: Integrate with SIEM and automated response
