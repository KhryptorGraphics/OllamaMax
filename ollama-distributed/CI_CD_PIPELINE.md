# OllamaMax CI/CD Pipeline Documentation

## 🚀 Complete Automated Deployment Pipeline

This document describes the comprehensive CI/CD pipeline for OllamaMax, providing automated testing, security validation, deployment, and monitoring.

## 📋 Pipeline Overview

### **Pipeline Architecture**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Development   │───▶│   Quality Gates │───▶│   Deployment    │
│                 │    │                 │    │                 │
│ • Code Changes  │    │ • Unit Tests    │    │ • Staging       │
│ • Pull Requests │    │ • Integration   │    │ • Production    │
│ • Feature Work  │    │ • Security      │    │ • Rollback      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   Monitoring    │
                       │                 │
                       │ • Health Checks │
                       │ • Metrics       │
                       │ • Alerting      │
                       └─────────────────┘
```

## 🔄 Workflow Triggers

### **Automated Triggers**
- **Push to `main`**: Triggers staging deployment
- **Release tags (`v*`)**: Triggers production deployment
- **Pull Requests**: Runs quality gates only
- **Schedule**: Daily comprehensive tests

### **Manual Triggers**
- **Workflow Dispatch**: Manual deployment to any environment
- **Release Creation**: Manual release with custom options
- **Rollback**: Emergency rollback procedures

## 🛠️ Pipeline Components

### **1. Quality Gates (`.github/workflows/test.yml`)**
```yaml
# Comprehensive testing pipeline
- Unit Tests (Go 1.21, 1.22)
- Integration Tests
- Security Scanning (gosec, nancy, govulncheck)
- Multi-platform builds (Linux, Windows, macOS)
- Code coverage reporting
```

### **2. Deployment Pipeline (`.github/workflows/deploy.yml`)**
```yaml
# Automated deployment workflow
- Deployment strategy determination
- Security validation
- Artifact building (binaries + Docker images)
- Environment-specific deployment
- Health checks and smoke tests
- Rollback on failure
```

### **3. Release Management (`.github/workflows/release.yml`)**
```yaml
# Release automation
- Version validation
- Release artifact creation
- GitHub release publishing
- Docker image tagging
- Release notes generation
- Production deployment trigger
```

## 🎯 Deployment Environments

### **Staging Environment**
- **Trigger**: Push to `main` branch
- **Purpose**: Pre-production validation
- **Configuration**: 2 replicas, Docker deployment
- **URL**: `https://staging.ollamamax.com`
- **Health Checks**: 5-minute timeout

### **Production Environment**
- **Trigger**: Release tags (`v*`)
- **Purpose**: Live production system
- **Configuration**: 3 replicas, Kubernetes deployment
- **URL**: `https://ollamamax.com`
- **Health Checks**: 10-minute timeout
- **Blue-Green Deployment**: Zero-downtime updates

## 🔐 Security Integration

### **Security Validation Pipeline**
```bash
# Automated security checks
1. Security hardening validation
2. Vulnerability scanning (gosec, nancy, govulncheck)
3. Security test suite execution
4. Environment variable validation
5. TLS certificate verification
```

### **Security Gates**
- **All security tests must pass** before deployment
- **No critical vulnerabilities** allowed in production
- **Security headers** validated in smoke tests
- **Authentication** required for sensitive endpoints

## 📦 Artifact Management

### **Binary Artifacts**
```bash
# Multi-platform binaries
- Linux (amd64, arm64)
- macOS (amd64, arm64)  
- Windows (amd64)
- Checksums and signatures
```

### **Container Images**
```bash
# Docker images
- Multi-architecture (linux/amd64, linux/arm64)
- Tagged with version and 'latest'
- Pushed to GitHub Container Registry
- Security scanning included
```

## 🏥 Health Checks & Validation

### **Deployment Validation**
```bash
# Automated health checks
./scripts/health-check.sh --environment staging
./scripts/health-check.sh --environment production

# Smoke tests
go test ./tests/smoke/... -staging-url=$STAGING_URL
go test ./tests/smoke/... -production-url=$PRODUCTION_URL
```

### **Health Check Components**
- **Connectivity**: Basic service accessibility
- **Health Endpoint**: `/health` endpoint validation
- **API Endpoints**: Core API functionality
- **Performance**: Response time validation
- **Security Headers**: Security configuration validation

## 🔄 Rollback Procedures

### **Automatic Rollback**
```yaml
# Triggered on deployment failure
- Health check failures
- Smoke test failures
- Service startup failures
- Performance degradation
```

### **Manual Rollback**
```bash
# Emergency rollback commands
./scripts/rollback.sh --environment staging
./scripts/rollback.sh --environment production --force
./scripts/rollback.sh --environment production --steps 2
```

### **Rollback Strategies**
- **Kubernetes**: Rollout undo to previous revision
- **Docker**: Switch to previous image version
- **Local**: Restore previous binary version

## 📊 Monitoring & Alerting

### **Monitoring Setup**
```bash
# Automated monitoring deployment
./scripts/setup-monitoring.sh --environment staging
./scripts/setup-monitoring.sh --environment production \
  --slack-webhook $SLACK_WEBHOOK \
  --email-alerts admin@company.com
```

### **Monitoring Stack**
- **Prometheus**: Metrics collection and alerting
- **Grafana**: Dashboards and visualization
- **Alertmanager**: Alert routing and notifications
- **Node Exporter**: System metrics

### **Alert Conditions**
- **Service Down**: Service unavailable for >1 minute
- **High Error Rate**: >10% error rate for >2 minutes
- **High Latency**: 95th percentile >1s for >5 minutes
- **Resource Usage**: CPU >80% or Memory >90% for >5 minutes

## 🔧 Configuration Management

### **Environment Variables**
```bash
# Required for deployment
JWT_SECRET="secure-32-character-secret"
ADMIN_DEFAULT_PASSWORD="secure-admin-password"
TLS_CERT_FILE="/path/to/server.crt"
TLS_KEY_FILE="/path/to/server.key"

# Optional for notifications
SLACK_WEBHOOK_URL="https://hooks.slack.com/..."
```

### **GitHub Secrets**
```bash
# Required repository secrets
GITHUB_TOKEN          # Automatic (for releases)
SLACK_WEBHOOK_URL      # Slack notifications
STAGING_URL           # Staging environment URL
PRODUCTION_URL        # Production environment URL
```

## 🚀 Deployment Commands

### **Manual Deployment**
```bash
# Deploy to staging
gh workflow run deploy.yml -f environment=staging

# Deploy to production
gh workflow run deploy.yml -f environment=production

# Create release
gh workflow run release.yml -f version=v1.2.3
```

### **Local Testing**
```bash
# Test deployment locally
./deploy/scripts/deploy.sh --type local --environment development

# Run health checks
./scripts/health-check.sh --api-url http://localhost:8080

# Test rollback
./scripts/rollback.sh --environment development --dry-run
```

## 📈 Pipeline Metrics

### **Quality Metrics**
- **Test Coverage**: >80% required
- **Security Scan**: 0 critical vulnerabilities
- **Build Time**: <10 minutes for full pipeline
- **Deployment Time**: <5 minutes for staging, <10 minutes for production

### **Reliability Metrics**
- **Pipeline Success Rate**: >95%
- **Deployment Success Rate**: >98%
- **Rollback Time**: <2 minutes
- **Recovery Time**: <5 minutes

## 🔍 Troubleshooting

### **Common Issues**

#### **Deployment Failures**
```bash
# Check deployment logs
kubectl logs -n ollama-staging deployment/ollama-distributed

# Check health status
./scripts/health-check.sh --environment staging --verbose

# Manual rollback
./scripts/rollback.sh --environment staging --force
```

#### **Test Failures**
```bash
# Run specific test suites
go test ./tests/unit/... -v
go test ./tests/integration/... -v
go test ./tests/security/... -v

# Check security issues
gosec -fmt json -out security-report.json ./...
```

#### **Build Issues**
```bash
# Verify build environment
./scripts/build-verify.sh

# Check dependencies
go mod tidy
go mod verify
```

## 📚 Additional Resources

### **Documentation**
- [Security Hardening Guide](SECURITY_HARDENING.md)
- [Build Instructions](BUILD_INSTRUCTIONS.md)
- [CLI Reference](CLI_REFERENCE.md)
- [Testing Framework](TESTING_FRAMEWORK.md)

### **Scripts**
- `scripts/health-check.sh` - Deployment health validation
- `scripts/rollback.sh` - Automated rollback procedures
- `scripts/setup-monitoring.sh` - Monitoring stack deployment
- `scripts/security-hardening.sh` - Security configuration

### **Workflows**
- `.github/workflows/test.yml` - Comprehensive testing
- `.github/workflows/deploy.yml` - Automated deployment
- `.github/workflows/release.yml` - Release management

## ✅ Success Criteria

### **Pipeline Success Indicators**
- ✅ **All tests pass** in quality gates
- ✅ **Security validation** completes successfully
- ✅ **Deployment** completes without errors
- ✅ **Health checks** pass post-deployment
- ✅ **Smoke tests** validate functionality
- ✅ **Monitoring** is operational
- ✅ **Rollback capability** is verified

### **Production Readiness**
- ✅ **Zero-downtime deployments** with blue-green strategy
- ✅ **Automated rollback** on failure detection
- ✅ **Comprehensive monitoring** and alerting
- ✅ **Security hardening** fully implemented
- ✅ **Performance validation** in production
- ✅ **Disaster recovery** procedures tested

The OllamaMax CI/CD pipeline provides enterprise-grade automation for secure, reliable, and efficient software delivery.
