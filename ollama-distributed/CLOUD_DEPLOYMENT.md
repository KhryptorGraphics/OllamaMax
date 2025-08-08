# ☁️ OllamaMax Cloud Deployment Guide

## 🎯 Overview

This guide provides comprehensive instructions for deploying OllamaMax across major cloud providers using Infrastructure as Code (Terraform) and container orchestration.

## 🏗️ Architecture Overview

### **Cloud-Native Architecture**
```
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer                            │
│                  (Cloud Provider)                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Kubernetes Cluster                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ OllamaMax   │  │ OllamaMax   │  │ OllamaMax   │         │
│  │   Node 1    │  │   Node 2    │  │   Node 3    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Prometheus  │  │   Grafana   │  │  AlertMgr   │         │
│  │ Monitoring  │  │ Dashboard   │  │  Alerting   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### **Deployment Components**
- **Container Images**: Multi-architecture (AMD64, ARM64) container images
- **Kubernetes Orchestration**: StatefulSets for persistent AI model storage
- **Infrastructure as Code**: Terraform modules for cloud resources
- **Monitoring Stack**: Prometheus, Grafana, and AlertManager
- **Security**: Network policies, RBAC, and secret management

## 🚀 Quick Start Deployment

### **Prerequisites**
```bash
# Required tools
- Docker (for local testing)
- kubectl (Kubernetes CLI)
- terraform (Infrastructure as Code)
- helm (Kubernetes package manager)
- Cloud CLI (aws/gcloud/az)
```

### **1. Clone and Prepare**
```bash
# Clone the repository
git clone https://github.com/KhryptorGraphics/OllamaMax.git
cd OllamaMax/ollama-distributed

# Build container image (optional - pre-built images available)
docker build -t ollama-distributed:latest .

# Verify container
docker run --rm ollama-distributed:latest ollama-distributed version
```

## ☁️ AWS Deployment

### **AWS EKS Deployment**

#### **1. Prerequisites**
```bash
# Install AWS CLI
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip && sudo ./aws/install

# Configure AWS credentials
aws configure

# Install eksctl (optional)
curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin
```

#### **2. Deploy Infrastructure**
```bash
# Navigate to AWS Terraform
cd infrastructure/terraform/aws

# Initialize Terraform
terraform init

# Plan deployment
terraform plan -var="cluster_name=ollama-prod" \
               -var="region=us-west-2" \
               -var="node_instance_type=m5.large" \
               -var="desired_nodes=3"

# Deploy infrastructure
terraform apply -auto-approve
```

#### **3. Configure kubectl**
```bash
# Update kubeconfig
aws eks update-kubeconfig --region us-west-2 --name ollama-prod

# Verify cluster access
kubectl get nodes
kubectl get pods -n ollama-system
```

#### **4. Access Your Deployment**
```bash
# Get LoadBalancer URL
kubectl get service ollama-loadbalancer -n ollama-system

# Access Web UI
echo "Web UI: http://$(kubectl get service ollama-loadbalancer -n ollama-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'):8081"

# Access API
echo "API: http://$(kubectl get service ollama-loadbalancer -n ollama-system -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')/api/v1/proxy/status"
```

### **AWS Configuration Options**
```hcl
# terraform.tfvars example
region              = "us-west-2"
cluster_name        = "ollama-production"
node_instance_type  = "m5.xlarge"
min_nodes          = 2
max_nodes          = 20
desired_nodes      = 5
domain_name        = "ollama.yourdomain.com"
enable_monitoring  = true
enable_logging     = true
```

## 🌐 Google Cloud Platform (GCP) Deployment

### **GCP GKE Deployment**

#### **1. Prerequisites**
```bash
# Install Google Cloud SDK
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# Initialize and authenticate
gcloud init
gcloud auth application-default login
```

#### **2. Create GCP Infrastructure**
```bash
# Create GCP Terraform configuration
cd infrastructure/terraform/gcp

# Initialize Terraform
terraform init

# Deploy GKE cluster
terraform plan -var="project_id=your-project-id" \
               -var="region=us-central1" \
               -var="cluster_name=ollama-gke"

terraform apply -auto-approve
```

#### **3. Configure kubectl for GKE**
```bash
# Get cluster credentials
gcloud container clusters get-credentials ollama-gke --region us-central1

# Verify cluster
kubectl get nodes
```

### **GCP Configuration Example**
```hcl
# GCP terraform.tfvars
project_id         = "your-gcp-project"
region            = "us-central1"
cluster_name      = "ollama-gke"
node_machine_type = "e2-standard-4"
node_count        = 3
disk_size_gb      = 100
```

## 🔷 Microsoft Azure Deployment

### **Azure AKS Deployment**

#### **1. Prerequisites**
```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# Login to Azure
az login

# Set subscription
az account set --subscription "your-subscription-id"
```

#### **2. Deploy AKS Infrastructure**
```bash
# Navigate to Azure Terraform
cd infrastructure/terraform/azure

# Initialize and deploy
terraform init
terraform plan -var="resource_group_name=ollama-rg" \
               -var="location=East US" \
               -var="cluster_name=ollama-aks"

terraform apply -auto-approve
```

#### **3. Configure kubectl for AKS**
```bash
# Get AKS credentials
az aks get-credentials --resource-group ollama-rg --name ollama-aks

# Verify cluster
kubectl get nodes
```

## 🐳 Container Registry Setup

### **GitHub Container Registry (Recommended)**
```bash
# Login to GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Pull pre-built image
docker pull ghcr.io/khryptorgraphics/ollama-distributed:latest

# Use in Kubernetes
kubectl set image statefulset/ollama-distributed \
  ollama-distributed=ghcr.io/khryptorgraphics/ollama-distributed:latest \
  -n ollama-system
```

### **AWS ECR**
```bash
# Create ECR repository
aws ecr create-repository --repository-name ollama-distributed

# Get login token
aws ecr get-login-password --region us-west-2 | \
  docker login --username AWS --password-stdin \
  123456789012.dkr.ecr.us-west-2.amazonaws.com

# Tag and push
docker tag ollama-distributed:latest \
  123456789012.dkr.ecr.us-west-2.amazonaws.com/ollama-distributed:latest
docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/ollama-distributed:latest
```

### **Google Container Registry**
```bash
# Configure Docker for GCR
gcloud auth configure-docker

# Tag and push
docker tag ollama-distributed:latest gcr.io/your-project/ollama-distributed:latest
docker push gcr.io/your-project/ollama-distributed:latest
```

## 📊 Monitoring and Observability

### **Prometheus and Grafana Setup**
```bash
# Install monitoring stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin123

# Access Grafana
kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring
# Open http://localhost:3000 (admin/admin123)
```

### **Custom Dashboards**
```bash
# Import OllamaMax dashboards
kubectl apply -f deploy/monitoring/grafana-dashboards/

# Available dashboards:
# - OllamaMax Performance Monitoring
# - Cluster Resource Usage
# - AI Model Performance
# - Security and Compliance
```

## 🔒 Security Configuration

### **Network Security**
```bash
# Apply network policies
kubectl apply -f deploy/kubernetes/security/network-policies.yaml

# Configure RBAC
kubectl apply -f deploy/kubernetes/security/rbac.yaml

# Setup TLS certificates
kubectl apply -f deploy/kubernetes/security/tls-config.yaml
```

### **Secret Management**
```bash
# Create secrets for production
kubectl create secret generic ollama-secrets \
  --from-literal=jwt-secret="your-secure-jwt-secret" \
  --from-literal=admin-password="your-admin-password" \
  -n ollama-system

# Use external secret management (AWS Secrets Manager, etc.)
kubectl apply -f deploy/kubernetes/security/external-secrets.yaml
```

## 🔧 Production Optimization

### **Resource Scaling**
```bash
# Configure Horizontal Pod Autoscaler
kubectl apply -f deploy/kubernetes/autoscaling/hpa.yaml

# Configure Vertical Pod Autoscaler
kubectl apply -f deploy/kubernetes/autoscaling/vpa.yaml

# Configure Cluster Autoscaler (cloud-specific)
kubectl apply -f deploy/kubernetes/autoscaling/cluster-autoscaler.yaml
```

### **Performance Tuning**
```yaml
# Production resource limits
resources:
  requests:
    cpu: "1"
    memory: "2Gi"
  limits:
    cpu: "4"
    memory: "8Gi"

# Storage optimization
volumeClaimTemplates:
- metadata:
    name: models
  spec:
    storageClassName: "fast-ssd"
    resources:
      requests:
        storage: "500Gi"
```

## 🚨 Disaster Recovery

### **Backup Strategy**
```bash
# Backup persistent volumes
kubectl apply -f deploy/kubernetes/backup/velero-backup.yaml

# Backup configuration
kubectl get configmap ollama-config -o yaml > backup/ollama-config.yaml
kubectl get secret ollama-secrets -o yaml > backup/ollama-secrets.yaml
```

### **Multi-Region Deployment**
```bash
# Deploy to multiple regions
terraform workspace new us-west-2
terraform apply -var="region=us-west-2"

terraform workspace new us-east-1
terraform apply -var="region=us-east-1"

# Configure cross-region replication
kubectl apply -f deploy/kubernetes/multi-region/
```

## 📈 Scaling Guidelines

### **Horizontal Scaling**
- **Small deployment**: 3 nodes, 2 CPU, 4GB RAM each
- **Medium deployment**: 5-10 nodes, 4 CPU, 8GB RAM each
- **Large deployment**: 10+ nodes, 8 CPU, 16GB RAM each
- **Enterprise deployment**: 50+ nodes with auto-scaling

### **Storage Scaling**
- **Development**: 50GB per node
- **Production**: 200GB+ per node
- **Enterprise**: 1TB+ per node with high-performance SSD

### **Network Considerations**
- **Bandwidth**: 1Gbps+ for model synchronization
- **Latency**: <10ms between nodes for optimal performance
- **Security**: VPN or private networking for multi-region

## 🎯 Cost Optimization

### **Resource Optimization**
```bash
# Use spot instances for cost savings
terraform apply -var="use_spot_instances=true" \
               -var="spot_instance_types=['m5.large','m5.xlarge']"

# Configure resource quotas
kubectl apply -f deploy/kubernetes/resource-quotas/
```

### **Auto-scaling Configuration**
```yaml
# Cost-effective auto-scaling
spec:
  minReplicas: 2
  maxReplicas: 20
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

## 🔍 Troubleshooting

### **Common Issues**
```bash
# Check pod status
kubectl get pods -n ollama-system

# View logs
kubectl logs -f statefulset/ollama-distributed -n ollama-system

# Debug networking
kubectl exec -it ollama-distributed-0 -n ollama-system -- /bin/bash

# Check resource usage
kubectl top pods -n ollama-system
kubectl top nodes
```

### **Health Checks**
```bash
# Cluster health
kubectl get componentstatuses

# Application health
curl http://your-loadbalancer/health

# Monitoring health
kubectl get pods -n monitoring
```

## 🎉 Success Metrics

### **Deployment Validation**
- ✅ **All pods running**: `kubectl get pods -n ollama-system`
- ✅ **Services accessible**: Web UI and API endpoints responding
- ✅ **Monitoring active**: Prometheus collecting metrics
- ✅ **Security configured**: Network policies and RBAC active
- ✅ **Auto-scaling working**: HPA and cluster autoscaler functional

### **Performance Targets**
- **API Response Time**: <100ms for health checks
- **Model Loading Time**: <30 seconds for standard models
- **Cluster Formation**: <5 minutes for new nodes
- **Failover Time**: <30 seconds for node failures

The OllamaMax cloud deployment provides **enterprise-grade distributed AI infrastructure** that scales automatically and maintains high availability across multiple cloud providers.
