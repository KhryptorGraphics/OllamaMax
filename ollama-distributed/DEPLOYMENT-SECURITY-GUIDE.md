# Ollama Distributed - Security Hardened Deployment Guide

## 🔒 Security Overview

This deployment guide covers the security-hardened configuration for Ollama Distributed, addressing all critical vulnerabilities and implementing production-ready security measures.

## ✅ Security Issues Resolved

### 🚨 Critical Security Fixes Applied

1. **JWT Secrets**: Removed hardcoded secrets, implemented environment variable configuration
2. **CORS Policies**: Replaced wildcard origins with specific domain restrictions
3. **TLS Configuration**: Upgraded to TLS 1.3 with modern cipher suites
4. **Port Conflicts**: Resolved all port conflicts across services
5. **Security Headers**: Implemented comprehensive HTTP security headers
6. **Environment Variables**: Created secure environment templates
7. **Certificate Management**: Added complete TLS certificate configuration
8. **Backup Security**: Implemented encrypted backup and restore procedures

## 🚀 Quick Start - Secure Deployment

### Prerequisites

1. **Generate Secure Secrets**
   ```bash
   # JWT Secret (64 characters)
   openssl rand -base64 64
   
   # Encryption Key (32 bytes hex)
   openssl rand -hex 32
   
   # Backup Encryption Key
   openssl rand -base64 32
   ```

2. **TLS Certificates**
   ```bash
   # For production, use Let's Encrypt or CA-signed certificates
   # For development only:
   ./scripts/generate-dev-certs.sh
   ```

### Deployment Options

## 🐳 Docker Deployment (Recommended for Development)

### 1. Environment Setup
```bash
# Copy environment template
cp deploy/docker/.env.docker deploy/docker/.env

# Edit with your secure values
vi deploy/docker/.env
```

### 2. Configure Secrets
```bash
# Replace ALL placeholder values in .env file:
JWT_SECRET=your_secure_64_char_jwt_secret_here
GRAFANA_ADMIN_PASSWORD=your_secure_grafana_password
REDIS_PASSWORD=your_secure_redis_password
DATABASE_PASSWORD=your_secure_database_password
```

### 3. Deploy
```bash
cd deploy/docker
docker-compose up -d
```

### 4. Verify Security
```bash
# Run security audit
./scripts/security-audit.sh

# Check TLS configuration
curl -k https://localhost:18080/health
```

## ☸️ Kubernetes Deployment (Recommended for Production)

### 1. Create Namespace
```bash
kubectl create namespace ollama-distributed
```

### 2. Configure Secrets
```bash
# Create secrets from template
cp deploy/kubernetes/secrets-template.yaml deploy/kubernetes/secrets.yaml

# Base64 encode your actual secrets
echo -n "your_jwt_secret" | base64
echo -n "your_database_password" | base64

# Update secrets.yaml with real values
vi deploy/kubernetes/secrets.yaml

# Apply secrets
kubectl apply -f deploy/kubernetes/secrets.yaml -n ollama-distributed
```

### 3. Configure ConfigMaps
```bash
# Apply configuration
kubectl apply -f deploy/kubernetes/configmap-template.yaml -n ollama-distributed
```

### 4. Deploy with Helm
```bash
cd deploy/kubernetes/helm/ollamacron

# Update values.yaml with your domain
vi values.yaml

# Install
helm install ollama-distributed . -n ollama-distributed
```

### 5. Configure Ingress
```bash
# Update with your domain and SSL certificate
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ollama-ingress
  namespace: ollama-distributed
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: ollama-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ollama-distributed
            port:
              number: 8080
EOF
```

## 🔒 Security Configuration Details

### TLS/SSL Configuration

**Certificates Location:**
- Certificate: `/etc/ssl/certs/ollama-distributed.crt`
- Private Key: `/etc/ssl/private/ollama-distributed.key`
- CA Bundle: `/etc/ssl/certs/ca-bundle.crt`

**TLS Settings:**
- Minimum Version: TLS 1.3
- Cipher Suites: AES-256-GCM, AES-128-GCM, ChaCha20-Poly1305
- HSTS: 1 year, includeSubDomains, preload

### CORS Configuration

**Production Settings:**
```yaml
cors:
  allowed_origins: 
    - "https://app.your-domain.com"
    - "https://api.your-domain.com"
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["Authorization", "Content-Type", "X-Request-ID"]
  allow_credentials: true
  max_age: 3600
```

### Security Headers

All HTTP responses include:
- `Strict-Transport-Security`: Force HTTPS
- `X-Frame-Options`: Prevent clickjacking
- `X-Content-Type-Options`: Prevent MIME sniffing
- `X-XSS-Protection`: XSS protection
- `Content-Security-Policy`: Restrict resource loading
- `Referrer-Policy`: Control referrer information

### Authentication & Authorization

**JWT Configuration:**
- Algorithm: HS256
- Expiry: 24 hours
- Refresh Token: 7 days
- Secret: 256-bit secure random key

### Rate Limiting

**Default Limits:**
- API requests: 100 requests/minute
- Upload endpoints: 10 requests/minute
- Burst capacity: 200 requests
- Connection limit: 50 per IP

## 📊 Monitoring & Alerting

### Metrics Collection

**Endpoints:**
- Coordinator: `https://localhost:19090/metrics`
- Node 1: `https://localhost:19091/metrics`
- Node 2: `https://localhost:19092/metrics`
- Prometheus: `https://localhost:19093`

### Grafana Dashboard

**Access:**
- URL: `https://localhost:3000`
- Username: `admin`
- Password: Set in environment variables

### Security Alerts

Configured alerts for:
- High error rates (>5%)
- High latency (>10 seconds)
- Memory usage (>90%)
- CPU usage (>80%)
- Failed authentication attempts
- TLS certificate expiry

## 🛡️ Security Hardening Checklist

### ✅ Completed Configurations

- [x] JWT secrets from environment variables
- [x] TLS 1.3 with strong cipher suites
- [x] Restricted CORS origins
- [x] Security headers implementation
- [x] Port conflict resolution
- [x] Environment variable templates
- [x] Encrypted backup configuration
- [x] Database migration scripts
- [x] Kubernetes security contexts
- [x] Network policies
- [x] Resource limits and requests
- [x] Audit logging configuration

### 🔧 Additional Hardening (Optional)

- [ ] WAF (Web Application Firewall)
- [ ] DDoS protection
- [ ] IP whitelisting
- [ ] OAuth2/OIDC integration
- [ ] Certificate pinning
- [ ] Security scanning automation
- [ ] Penetration testing
- [ ] Compliance auditing (SOC2, ISO27001)

## 🔍 Security Audit

### Run Security Audit
```bash
# Execute comprehensive security check
./scripts/security-audit.sh

# Review report
cat security-audit-report.txt
```

### Manual Verification

1. **TLS Configuration**
   ```bash
   # Test TLS strength
   nmap --script ssl-enum-ciphers -p 443 your-domain.com
   ```

2. **Security Headers**
   ```bash
   # Check headers
   curl -I https://your-domain.com
   ```

3. **CORS Policy**
   ```bash
   # Test CORS
   curl -H "Origin: https://malicious.com" \
        -H "Access-Control-Request-Method: POST" \
        -X OPTIONS https://your-domain.com/api/test
   ```

## 🚨 Incident Response

### Security Incident Procedures

1. **Immediate Response**
   - Isolate affected systems
   - Preserve logs and evidence
   - Notify security team
   - Document timeline

2. **Investigation**
   - Analyze logs and metrics
   - Identify attack vectors
   - Assess damage scope
   - Collect forensic evidence

3. **Recovery**
   - Patch vulnerabilities
   - Restore from clean backups
   - Update security configurations
   - Validate system integrity

4. **Post-Incident**
   - Conduct lessons learned
   - Update security procedures
   - Implement additional controls
   - Schedule follow-up audits

## 📋 Maintenance Schedule

### Daily
- Monitor security alerts
- Review access logs
- Check certificate validity
- Validate backup integrity

### Weekly
- Update security patches
- Review security metrics
- Rotate access keys
- Audit user permissions

### Monthly
- Run security audit script
- Review security policies
- Update threat intelligence
- Conduct security training

### Quarterly
- Penetration testing
- Security architecture review
- Compliance assessment
- Business continuity testing

## 📞 Support & Security Contacts

### Security Issues
- Email: security@your-domain.com
- Emergency: +1-xxx-xxx-xxxx
- PGP Key: [public-key-fingerprint]

### Documentation
- Security Policies: `/docs/security/`
- Incident Response: `/docs/incident-response/`
- Compliance: `/docs/compliance/`

---

## 🔐 Important Security Notes

1. **Never commit secrets to version control**
2. **Regularly rotate passwords and keys**
3. **Keep security patches up to date**
4. **Monitor for suspicious activities**
5. **Backup encryption keys securely**
6. **Test disaster recovery procedures**
7. **Train team on security best practices**

---

*This deployment guide ensures a security-hardened installation of Ollama Distributed suitable for production environments. Regular security audits and updates are essential for maintaining security posture.*