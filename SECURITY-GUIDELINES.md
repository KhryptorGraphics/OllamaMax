# 🛡️ Security Guidelines for Ollama Distributed

## 🚨 Critical Security Requirements

Before using Ollama Distributed in any environment, please review these security requirements:

### 1. Environment Variables (Required)

**Production deployments MUST set these environment variables:**

```bash
# Required for production
export JWT_SECRET="$(openssl rand -hex 32)"
export DB_PASSWORD="$(openssl rand -base64 32)"
export MINIO_ROOT_USER="admin-$(openssl rand -hex 8)"
export MINIO_ROOT_PASSWORD="$(openssl rand -base64 32)"
export REDIS_PASSWORD="$(openssl rand -base64 32)"
export GRAFANA_PASSWORD="$(openssl rand -base64 16)"
```

### 2. TLS Certificates

**Generate proper TLS certificates for production:**

```bash
# Create certificate directory
mkdir -p certs

# Generate CA certificate
openssl genrsa -out certs/ca-key.pem 4096
openssl req -new -x509 -key certs/ca-key.pem -out certs/ca.pem -days 365

# Generate server certificates
openssl genrsa -out certs/server-key.pem 4096
openssl req -new -key certs/server-key.pem -out certs/server.csr
openssl x509 -req -in certs/server.csr -CA certs/ca.pem -CAkey certs/ca-key.pem -CAcreateserial -out certs/server.pem -days 365

# Set proper permissions
chmod 600 certs/*.pem certs/*.key
```

### 3. Docker Security

**Apply these security measures:**

```yaml
# In docker-compose files
security_opt:
  - no-new-privileges:true
cap_drop:
  - ALL
cap_add:
  - NET_BIND_SERVICE  # Only if binding to privileged ports
read_only: true      # Make containers read-only where possible
```

### 4. Network Security

**Configure secure networking:**

```yaml
networks:
  ollama-net:
    driver: bridge
    internal: true  # Prevent external access
    ipam:
      config:
        - subnet: 172.20.0.0/16  # Use non-conflicting subnet
```

### 5. Input Validation

**Always validate inputs:**

```bash
# Example validation
validate_input() {
    local input="$1"
    if [[ ! "$input" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        echo "❌ Invalid input: $input" >&2
        exit 1
    fi
}
```

## ⚠️ Security Warnings

### Development vs Production

- **NEVER** use development configurations in production
- **NEVER** use default passwords or secrets
- **ALWAYS** enable TLS in production
- **ALWAYS** validate environment variables

### Common Vulnerabilities

1. **Exposed Secrets**: Don't hardcode secrets in configuration files
2. **Weak Passwords**: Use cryptographically secure random passwords
3. **Missing TLS**: Always encrypt data in transit
4. **Privilege Escalation**: Run containers with minimal privileges
5. **Input Injection**: Validate all user inputs

### Security Checklist

Before deploying to production:

- [ ] All secrets are set via environment variables
- [ ] TLS certificates are properly generated and secured
- [ ] Container security options are applied
- [ ] Network access is restricted
- [ ] Input validation is implemented
- [ ] Logs don't contain sensitive information
- [ ] Regular security updates are applied
- [ ] Access controls are properly configured

## 🔍 Security Testing

### Automated Security Scanning

```bash
# Install security tools
pip install safety bandit
npm install -g retire

# Scan dependencies
safety check
bandit -r scripts/
retire --path .

# Docker security scanning
docker scout quickview
trivy image ollamamax/distributed:latest
```

### Manual Security Review

1. **Code Review**: Review all scripts and configurations
2. **Penetration Testing**: Test for common vulnerabilities
3. **Access Control Testing**: Verify authentication and authorization
4. **Network Security**: Test firewall rules and network isolation

## 📞 Security Contact

If you discover security vulnerabilities:

1. **DO NOT** create public issues
2. **DO NOT** publish vulnerabilities
3. **DO** contact the security team privately
4. **DO** provide detailed reproduction steps

## 🔄 Regular Security Maintenance

### Monthly Tasks

- [ ] Update all dependencies
- [ ] Rotate JWT secrets
- [ ] Review access logs
- [ ] Update TLS certificates (if expiring)

### Quarterly Tasks

- [ ] Security audit and penetration testing
- [ ] Review and update security policies
- [ ] Update security documentation
- [ ] Train team on new security practices

---

**Remember**: Security is a shared responsibility. Everyone using Ollama Distributed must follow these guidelines to maintain a secure environment.