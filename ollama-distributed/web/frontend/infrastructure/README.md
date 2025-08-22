# Ollama Frontend - Production Infrastructure

## [DEPLOY-SUCCESS] Overview

This directory contains the complete production deployment infrastructure for the Ollama Distributed frontend application, implementing enterprise-grade DevOps practices with zero-downtime deployments, comprehensive monitoring, and automated disaster recovery.

## Architecture Components

### 🚀 Deployment Strategies
- **Blue-Green Deployment**: Zero-downtime deployments with instant rollback capability
- **Canary Deployment**: Progressive traffic routing with automated analysis and rollback
- **Infrastructure as Code**: Complete Terraform configuration for reproducible deployments

### 📊 Monitoring & Observability
- **Prometheus**: Metrics collection with custom business metrics
- **Grafana**: Real-time dashboards with alerting
- **Alertmanager**: Multi-channel alerting (Slack, PagerDuty, Email)
- **ELK Stack**: Centralized logging with structured log parsing
- **Jaeger**: Distributed tracing (configured but not shown in detail)

### 🔐 Security & Compliance
- **Pod Security Standards**: Restricted security contexts
- **Network Policies**: Micro-segmentation and traffic isolation
- **OPA Gatekeeper**: Policy enforcement and admission control
- **Falco**: Runtime security monitoring
- **Image Security**: Vulnerability scanning and signed images

### 🔄 Auto-scaling & Performance
- **Horizontal Pod Autoscaler**: CPU/Memory/Custom metrics-based scaling
- **Vertical Pod Autoscaler**: Right-sizing recommendations
- **Cluster Autoscaler**: Node-level scaling
- **Predictive Scaling**: ML-based capacity planning

### 💾 Disaster Recovery
- **Velero Backups**: Automated daily and weekly backups
- **Multi-Region Replication**: Cross-region backup synchronization
- **Automated Failover**: Health-check based failover with RTO < 15 minutes
- **Point-in-Time Recovery**: Granular restore capabilities

## Directory Structure

```
infrastructure/
├── terraform/                 # Infrastructure as Code
│   ├── main.tf                # Main Terraform configuration
│   ├── variables.tf           # Input variables
│   ├── modules.tf             # Module declarations
│   └── modules/               # Terraform modules
├── k8s/                       # Kubernetes manifests
│   ├── namespace.yaml         # Namespace with RBAC and policies
│   ├── blue-green-deployment.yaml  # Blue-green deployment
│   ├── disaster-recovery.yaml # DR and backup configuration
│   ├── autoscaling.yaml       # HPA, VPA, and PDB configs
│   └── canary-deployment.yaml # Argo Rollouts canary config
├── monitoring/                # Monitoring stack
│   ├── prometheus.yaml        # Prometheus with custom rules
│   ├── grafana.yaml           # Grafana with dashboards
│   └── alertmanager.yaml      # Multi-channel alerting
├── logging/                   # Centralized logging
│   └── elk-stack.yaml         # Elasticsearch, Logstash, Kibana
├── security/                  # Security hardening
│   └── security-hardening.yaml # Security policies and scanning
├── cicd/                      # CI/CD pipeline
│   └── github-actions.yaml    # Complete GitHub Actions workflow
├── scripts/                   # Automation scripts
│   ├── deployment-automation.sh    # Production deployment script
│   └── production-go-live-checklist.md  # Go-live validation
└── README.md                  # This file
```

## Quick Start

### Prerequisites

1. **Tools Required**:
   ```bash
   # Install required tools
   kubectl version --client
   helm version
   terraform version
   velero version --client-only
   ```

2. **Cluster Setup**:
   ```bash
   # Ensure cluster access
   kubectl cluster-info
   kubectl get nodes
   ```

3. **Secrets Configuration**:
   ```bash
   # Create required secrets
   kubectl create secret generic ollama-frontend-secrets \
     --from-literal=database-url="postgresql://..." \
     --from-literal=redis-url="redis://..." \
     -n ollama-frontend
   ```

### Deployment Options

#### Option 1: Automated Deployment Script

```bash
# Production deployment with blue-green strategy
./infrastructure/scripts/deployment-automation.sh \
  --version v1.2.3 \
  --environment production \
  --strategy blue-green

# Canary deployment
./infrastructure/scripts/deployment-automation.sh \
  --version v1.2.3 \
  --environment production \
  --strategy canary

# Emergency rollback
./infrastructure/scripts/deployment-automation.sh \
  --rollback \
  --environment production
```

#### Option 2: Terraform Infrastructure

```bash
# Initialize Terraform
cd infrastructure/terraform
terraform init

# Plan deployment
terraform plan -var="app_version=v1.2.3" -var="environment=production"

# Apply infrastructure
terraform apply -var="app_version=v1.2.3" -var="environment=production"
```

#### Option 3: Manual Kubernetes Deployment

```bash
# Apply namespace and RBAC
kubectl apply -f infrastructure/k8s/namespace.yaml

# Deploy monitoring stack
kubectl apply -f infrastructure/monitoring/

# Deploy application
kubectl apply -f infrastructure/k8s/blue-green-deployment.yaml

# Enable auto-scaling
kubectl apply -f infrastructure/k8s/autoscaling.yaml
```

### CI/CD Integration

#### GitHub Actions

1. **Copy the workflow**:
   ```bash
   cp infrastructure/cicd/github-actions.yaml .github/workflows/frontend-production.yaml
   ```

2. **Configure secrets** in GitHub repository:
   ```
   KUBECONFIG_PRODUCTION    # Base64 encoded kubeconfig
   KUBECONFIG_STAGING       # Base64 encoded kubeconfig  
   SNYK_TOKEN              # Snyk security scanning
   SLACK_WEBHOOK           # Deployment notifications
   ```

3. **Trigger deployment**:
   - Push to `main` branch → staging deployment
   - Create release → production deployment

## Configuration

### Environment Variables

```bash
# Application configuration
NODE_ENV=production
PORT=3000
LOG_LEVEL=info

# Database configuration
DATABASE_URL=postgresql://user:pass@host:5432/db
REDIS_URL=redis://host:6379

# Monitoring configuration
PROMETHEUS_ENDPOINT=http://prometheus:9090
GRAFANA_ENDPOINT=http://grafana:3000
```

### Resource Limits

```yaml
# Default resource configuration
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"

# Auto-scaling configuration
hpa:
  minReplicas: 3
  maxReplicas: 50
  targetCPU: 70%
  targetMemory: 80%
```

## Monitoring & Alerting

### Key Metrics Dashboard

Access Grafana at: `https://grafana.ollama.example.com`

**Critical Metrics**:
- Request Rate (target: < 1000 rps/pod)
- Error Rate (target: < 0.1%)
- Response Time P95 (target: < 500ms)
- CPU Utilization (target: < 70%)
- Memory Utilization (target: < 80%)

### Alert Channels

- **Critical Alerts**: PagerDuty + Slack + Email
- **Warning Alerts**: Slack + Email
- **Info Alerts**: Email only

### Log Analysis

Access Kibana at: `https://kibana.ollama.example.com`

**Log Patterns**:
```bash
# Error logs
level:error AND kubernetes.namespace:ollama-frontend

# Performance issues  
response_time:>1000 AND kubernetes.namespace:ollama-frontend

# Security events
message:"authentication failed" OR message:"unauthorized"
```

## Disaster Recovery

### Backup Strategy

- **Daily Backups**: Full namespace backup at 2 AM UTC
- **Weekly Backups**: Complete cluster backup on Sundays
- **Retention**: 90 days for daily, 180 days for weekly
- **Cross-Region**: Automatic replication to DR region

### Recovery Procedures

#### Point-in-Time Recovery
```bash
# List available backups
velero backup get

# Restore from specific backup
velero restore create --from-backup=backup-20241201-020000 \
  --include-namespaces=ollama-frontend
```

#### Regional Failover
```bash
# Check failover status
kubectl get configmap failover-config -n ollama-frontend -o yaml

# Manual failover trigger
kubectl annotate configmap failover-config -n ollama-frontend \
  failover.triggered="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  failover.reason="manual_trigger"
```

### RTO/RPO Targets

- **RTO (Recovery Time Objective)**: 15 minutes
- **RPO (Recovery Point Objective)**: 5 minutes
- **Availability Target**: 99.9% uptime

## Security

### Security Hardening Features

- **Non-root containers** with read-only filesystems
- **Network segmentation** with Kubernetes NetworkPolicies
- **Pod Security Standards** enforced at namespace level
- **Admission controllers** with OPA Gatekeeper
- **Runtime security** monitoring with Falco
- **Vulnerability scanning** with Trivy integration

### Compliance

- **SOC 2 Type II** controls implemented
- **ISO 27001** security practices
- **Automated compliance** monitoring and reporting
- **Security metrics** exposed via Prometheus

## Performance Optimization

### Auto-scaling Configuration

```yaml
# Horizontal scaling triggers
- CPU > 70% for 2 minutes
- Memory > 80% for 2 minutes  
- Request rate > 100 rps/pod
- Response time P99 > 500ms

# Vertical scaling
- Automatic resource right-sizing
- Historical usage analysis
- Predictive capacity planning
```

### Performance Targets

| Metric | Target | Alert Threshold |
|--------|--------|-----------------|
| Response Time (P95) | < 500ms | > 1s |
| Response Time (P99) | < 1s | > 2s |
| Error Rate | < 0.1% | > 1% |
| CPU Utilization | < 70% | > 85% |
| Memory Utilization | < 80% | > 90% |

## Troubleshooting

### Common Issues

#### Deployment Failures
```bash
# Check pod status
kubectl get pods -n ollama-frontend

# Check deployment events
kubectl describe deployment ollama-frontend-blue -n ollama-frontend

# Check rollout status
kubectl rollout status deployment/ollama-frontend-blue -n ollama-frontend
```

#### Performance Issues
```bash
# Check resource usage
kubectl top pods -n ollama-frontend

# Check HPA status
kubectl get hpa -n ollama-frontend

# Check custom metrics
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/ollama-frontend/pods/*/http_requests_per_second"
```

#### Security Alerts
```bash
# Check Falco alerts
kubectl logs -n falco-system -l app=falco

# Check OPA Gatekeeper violations
kubectl get constraintviolations

# Check security scan results
kubectl get configmap security-scan-results -n ollama-frontend -o yaml
```

### Support Contacts

| Role | Primary | Secondary |
|------|---------|-----------|
| Platform Engineering | platform-team@ollama.com | +1-555-0100 |
| Site Reliability | sre-team@ollama.com | +1-555-0101 |
| Security Team | security-team@ollama.com | +1-555-0102 |
| On-Call Engineer | oncall@ollama.com | +1-555-0103 |

## Contributing

### Making Infrastructure Changes

1. **Create feature branch**:
   ```bash
   git checkout -b infrastructure/new-feature
   ```

2. **Test changes locally**:
   ```bash
   # Validate Terraform
   terraform validate
   terraform plan
   
   # Validate Kubernetes manifests
   kubectl apply --dry-run=client -f infrastructure/k8s/
   ```

3. **Submit PR** with:
   - Detailed description of changes
   - Test results and validation
   - Impact assessment
   - Rollback plan

### Infrastructure Updates

- **Terraform versions**: Pin to specific versions
- **Kubernetes manifests**: Use semantic versioning
- **Container images**: Always use specific tags, never `latest`
- **Secrets rotation**: Follow 90-day rotation policy

---

## [DEPLOY-SUCCESS] Summary

This production infrastructure provides:

✅ **Zero-downtime deployments** with blue-green and canary strategies  
✅ **Comprehensive monitoring** with Prometheus, Grafana, and ELK stack  
✅ **Automated disaster recovery** with < 15 minute RTO  
✅ **Security hardening** with multiple layers of protection  
✅ **Auto-scaling** based on multiple metrics and predictive analysis  
✅ **CI/CD integration** with GitHub Actions  
✅ **Infrastructure as Code** with Terraform  
✅ **Production go-live checklist** with 100+ validation steps  

The infrastructure is production-ready and follows enterprise DevOps best practices for reliability, security, and performance.