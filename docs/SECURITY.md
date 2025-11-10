# Security Best Practices for OllamaMax

## Overview

This document outlines security best practices for deploying and operating OllamaMax in production environments.

## Table of Contents

1. [Authentication & Authorization](#authentication--authorization)
2. [Data Protection](#data-protection)
3. [Network Security](#network-security)
4. [Infrastructure Security](#infrastructure-security)
5. [Monitoring & Incident Response](#monitoring--incident-response)
6. [Compliance](#compliance)

---

## Authentication & Authorization

### JWT Configuration

**Critical:** Change default JWT secrets before production deployment.

```bash
# Generate strong JWT secret (minimum 32 characters)
openssl rand -base64 48

# Update .env
JWT_SECRET=<generated-secret>
JWT_SECRET_KEY=<generated-secret>
```

### Password Policy

- Minimum 8 characters
- Must include: uppercase, lowercase, number, special character
- Passwords are hashed using bcrypt (cost factor: 10)
- Implement password rotation policy (90 days recommended)

### API Key Management

- Rotate API keys regularly (30-90 days)
- Use different keys for different environments
- Never commit keys to version control
- Use environment variables or secret management systems

### Multi-Factor Authentication (MFA)

Recommended for production:
- Enable MFA for admin accounts
- Use TOTP (Time-based One-Time Password)
- Consider hardware security keys for critical accounts

---

## Data Protection

### Encryption at Rest

**Database:**
```bash
# Enable SQLite encryption (requires SQLCipher)
npm install @journeyapps/sqlcipher

# Update database configuration
DB_ENCRYPTION_KEY=<strong-encryption-key>
```

**File Storage:**
- Encrypt sensitive files before storage
- Use AES-256 encryption
- Store encryption keys in secure key management system

### Encryption in Transit

**HTTPS/TLS:**
```bash
# Generate production certificates
./scripts/setup-ssl.sh

# Or use Let's Encrypt
certbot certonly --standalone -d yourdomain.com

# Update .env
SSL_ENABLED=true
SSL_CERT_PATH=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

**TLS Configuration:**
- Use TLS 1.2 or higher
- Disable weak ciphers
- Enable Perfect Forward Secrecy (PFS)
- Use strong DH parameters (2048-bit minimum)

### Data Retention

- Implement data retention policies
- Automatically delete old logs (30-90 days)
- Anonymize or delete user data on request (GDPR compliance)
- Regular database backups with encryption

---

## Network Security

### Firewall Configuration

```bash
# Allow only necessary ports
ufw allow 22/tcp    # SSH
ufw allow 443/tcp   # HTTPS
ufw deny 13000/tcp  # Block direct API access (use reverse proxy)
ufw enable
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl http2;
    server_name api.ollamamax.com;

    ssl_certificate /etc/letsencrypt/live/api.ollamamax.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.ollamamax.com/privkey.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    location / {
        proxy_pass http://localhost:13000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### CORS Configuration

```javascript
// Restrict CORS in production
const corsOptions = {
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['https://yourdomain.com'],
  credentials: true,
  optionsSuccessStatus: 200
};

app.use(cors(corsOptions));
```

### Rate Limiting

```javascript
const rateLimit = require('express-rate-limit');

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  message: 'Too many requests from this IP'
});

app.use('/api/', limiter);
```

---

## Infrastructure Security

### Container Security

**Dockerfile Best Practices:**
```dockerfile
# Use specific versions, not :latest
FROM node:18.17.0-alpine

# Run as non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001
USER nodejs

# Minimize attack surface
RUN apk add --no-cache dumb-init
ENTRYPOINT ["dumb-init", "--"]

# Read-only filesystem where possible
VOLUME ["/app/data"]
```

**Docker Compose Security:**
```yaml
services:
  api:
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
```

### Environment Variables

**Never commit secrets:**
```bash
# .gitignore
.env
.env.local
.env.production
*.key
*.pem
certs/
```

**Use Secret Management:**
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Google Secret Manager

### Regular Updates

```bash
# Update dependencies weekly
npm audit fix
npm update

# Update system packages
apt update && apt upgrade -y

# Update Docker images
docker-compose pull
docker-compose up -d
```

---

## Monitoring & Incident Response

### Security Monitoring

**Enable audit logging:**
```javascript
// Log all authentication attempts
logger.info('Login attempt', {
  user: email,
  ip: req.ip,
  success: true/false,
  timestamp: new Date()
});
```

**Monitor for suspicious activity:**
- Multiple failed login attempts
- Unusual API usage patterns
- Large data exports
- Access from unusual locations

### Alerting

Configure alerts for:
- Failed authentication attempts (>5 in 5 minutes)
- API error rate >5%
- Unusual traffic patterns
- Certificate expiration (30 days before)
- Disk space <15%
- Memory usage >85%

### Incident Response Plan

1. **Detection:** Automated alerts + manual monitoring
2. **Containment:** Isolate affected systems
3. **Investigation:** Review logs, identify root cause
4. **Remediation:** Apply fixes, update security measures
5. **Recovery:** Restore services, verify integrity
6. **Post-Incident:** Document lessons learned, update procedures

---

## Compliance

### GDPR Compliance

- Implement data subject access requests (DSAR)
- Enable data portability
- Provide data deletion capabilities
- Maintain data processing records
- Implement privacy by design

### SOC 2 Compliance

- Access controls and authentication
- Encryption at rest and in transit
- Audit logging and monitoring
- Incident response procedures
- Regular security assessments

### HIPAA Compliance (if handling health data)

- Business Associate Agreements (BAA)
- Enhanced encryption requirements
- Audit controls
- Access controls and authentication
- Data backup and disaster recovery

---

## Security Checklist

### Pre-Production

- [ ] Change all default passwords and secrets
- [ ] Enable HTTPS/TLS with valid certificates
- [ ] Configure firewall rules
- [ ] Set up reverse proxy with security headers
- [ ] Enable rate limiting
- [ ] Configure CORS properly
- [ ] Run security audit: `./scripts/security-audit.sh`
- [ ] Perform penetration testing
- [ ] Review and update dependencies
- [ ] Set up monitoring and alerting
- [ ] Document incident response procedures
- [ ] Train team on security practices

### Post-Production

- [ ] Monitor security alerts daily
- [ ] Review access logs weekly
- [ ] Update dependencies monthly
- [ ] Rotate secrets quarterly
- [ ] Perform security audits quarterly
- [ ] Review and test incident response annually
- [ ] Update security documentation as needed

---

## Reporting Security Issues

If you discover a security vulnerability, please email: security@ollamamax.com

**Do not** create public GitHub issues for security vulnerabilities.

We will respond within 48 hours and work with you to address the issue.

---

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Node.js Security Best Practices](https://nodejs.org/en/docs/guides/security/)
- [Docker Security Best Practices](https://docs.docker.com/engine/security/)
- [CIS Benchmarks](https://www.cisecurity.org/cis-benchmarks/)

---

**Last Updated:** November 1, 2025  
**Version:** 1.0.0

