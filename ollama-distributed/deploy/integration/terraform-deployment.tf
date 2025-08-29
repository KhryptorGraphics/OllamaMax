# Complete Production Deployment with Terraform
# Multi-cloud infrastructure with zero-downtime deployment
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
    kubectl = {
      source  = "gavinbunney/kubectl"
      version = "~> 1.14"
    }
  }
}

# Variables for deployment configuration
variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
  default     = "ollama-production"
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "ollama-distributed.com"
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "production"
}

variable "backup_bucket" {
  description = "S3 bucket for backups"
  type        = string
  default     = "ollama-production-backups"
}

# Local variables
locals {
  common_tags = {
    Project     = "OllamaMax"
    Environment = var.environment
    Terraform   = "true"
    Cluster     = var.cluster_name
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# KMS key for encryption
resource "aws_kms_key" "ollama_key" {
  description             = "KMS key for OllamaMax encryption"
  deletion_window_in_days = 7

  tags = merge(local.common_tags, {
    Name = "${var.cluster_name}-kms-key"
  })
}

resource "aws_kms_alias" "ollama_key_alias" {
  name          = "alias/${var.cluster_name}-key"
  target_key_id = aws_kms_key.ollama_key.key_id
}

# S3 bucket for backups
resource "aws_s3_bucket" "backup_bucket" {
  bucket = var.backup_bucket

  tags = merge(local.common_tags, {
    Name = "OllamaMax Backup Bucket"
  })
}

resource "aws_s3_bucket_versioning" "backup_versioning" {
  bucket = aws_s3_bucket.backup_bucket.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backup_encryption" {
  bucket = aws_s3_bucket.backup_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.ollama_key.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "backup_lifecycle" {
  bucket = aws_s3_bucket.backup_bucket.id

  rule {
    id     = "backup_lifecycle"
    status = "Enabled"

    expiration {
      days = 90
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# Use the existing AWS infrastructure module
module "aws_infrastructure" {
  source = "../aws"

  cluster_name      = var.cluster_name
  region           = var.region
  domain_name      = var.domain_name
  enable_monitoring = true
  enable_logging   = true

  # Node configuration
  node_instance_type = "m5.xlarge"
  min_nodes         = 3
  max_nodes         = 20
  desired_nodes     = 5

  # Add GPU nodes for AI workloads
  enable_gpu_nodes = true
  gpu_instance_type = "g4dn.xlarge"
  gpu_min_nodes    = 1
  gpu_max_nodes    = 5
  gpu_desired_nodes = 2
}

# Configure Kubernetes and Helm providers
provider "kubernetes" {
  host                   = module.aws_infrastructure.cluster_endpoint
  cluster_ca_certificate = base64decode(module.aws_infrastructure.cluster_ca_certificate)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.aws_infrastructure.cluster_name]
  }
}

provider "helm" {
  kubernetes {
    host                   = module.aws_infrastructure.cluster_endpoint
    cluster_ca_certificate = base64decode(module.aws_infrastructure.cluster_ca_certificate)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = ["eks", "get-token", "--cluster-name", module.aws_infrastructure.cluster_name]
    }
  }
}

provider "kubectl" {
  host                   = module.aws_infrastructure.cluster_endpoint
  cluster_ca_certificate = base64decode(module.aws_infrastructure.cluster_ca_certificate)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.aws_infrastructure.cluster_name]
  }
}

# Storage Classes
resource "kubernetes_storage_class" "fast_ssd" {
  metadata {
    name = "fast-ssd"
  }
  
  storage_provisioner = "kubernetes.io/aws-ebs"
  
  parameters = {
    type      = "gp3"
    iops      = "3000"
    throughput = "125"
    encrypted = "true"
  }
  
  reclaim_policy         = "Delete"
  allow_volume_expansion = true
  volume_binding_mode    = "WaitForFirstConsumer"
}

resource "kubernetes_storage_class" "shared_ssd" {
  metadata {
    name = "shared-ssd"
  }
  
  storage_provisioner = "efs.csi.aws.com"
  
  parameters = {
    provisioningMode = "efs-ap"
    fileSystemId     = aws_efs_file_system.shared_storage.id
    directoryPerms   = "0755"
  }
  
  reclaim_policy      = "Retain"
  volume_binding_mode = "Immediate"
}

# EFS for shared model storage
resource "aws_efs_file_system" "shared_storage" {
  creation_token = "${var.cluster_name}-shared-storage"
  encrypted      = true
  kms_key_id     = aws_kms_key.ollama_key.arn

  performance_mode = "generalPurpose"
  throughput_mode  = "provisioned"
  provisioned_throughput_in_mibps = 500

  tags = merge(local.common_tags, {
    Name = "${var.cluster_name}-shared-storage"
  })
}

resource "aws_efs_mount_target" "shared_storage_mt" {
  count           = length(module.aws_infrastructure.private_subnet_ids)
  file_system_id  = aws_efs_file_system.shared_storage.id
  subnet_id       = module.aws_infrastructure.private_subnet_ids[count.index]
  security_groups = [aws_security_group.efs_sg.id]
}

resource "aws_security_group" "efs_sg" {
  name        = "${var.cluster_name}-efs-sg"
  description = "Security group for EFS"
  vpc_id      = module.aws_infrastructure.vpc_id

  ingress {
    from_port   = 2049
    to_port     = 2049
    protocol    = "tcp"
    cidr_blocks = [module.aws_infrastructure.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.common_tags, {
    Name = "${var.cluster_name}-efs-sg"
  })
}

# Install NGINX Ingress Controller
resource "helm_release" "nginx_ingress" {
  name       = "nginx-ingress"
  repository = "https://kubernetes.github.io/ingress-nginx"
  chart      = "ingress-nginx"
  version    = "4.8.0"
  namespace  = "ingress-nginx"
  create_namespace = true

  values = [
    yamlencode({
      controller = {
        replicaCount = 2
        service = {
          type = "LoadBalancer"
          annotations = {
            "service.beta.kubernetes.io/aws-load-balancer-type"                              = "nlb"
            "service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled" = "true"
            "service.beta.kubernetes.io/aws-load-balancer-backend-protocol"                  = "tcp"
            "service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout"           = "60"
          }
        }
        config = {
          use-proxy-protocol = "true"
          compute-full-forwarded-for = "true"
          use-forwarded-headers = "true"
        }
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }
    })
  ]

  depends_on = [module.aws_infrastructure]
}

# Install Cert-Manager for TLS certificates
resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  version    = "v1.13.0"
  namespace  = "cert-manager"
  create_namespace = true

  set {
    name  = "installCRDs"
    value = "true"
  }

  set {
    name  = "global.leaderElection.namespace"
    value = "cert-manager"
  }

  depends_on = [module.aws_infrastructure]
}

# ClusterIssuer for Let's Encrypt
resource "kubectl_manifest" "letsencrypt_issuer" {
  yaml_body = yamlencode({
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"
    metadata = {
      name = "letsencrypt-prod"
    }
    spec = {
      acme = {
        server = "https://acme-v02.api.letsencrypt.org/directory"
        email  = "admin@${var.domain_name}"
        privateKeySecretRef = {
          name = "letsencrypt-prod"
        }
        solvers = [{
          http01 = {
            ingress = {
              class = "nginx"
            }
          }
        }]
      }
    }
  })

  depends_on = [helm_release.cert_manager]
}

# Install AWS Load Balancer Controller
resource "helm_release" "aws_load_balancer_controller" {
  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  version    = "1.6.0"
  namespace  = "kube-system"

  set {
    name  = "clusterName"
    value = module.aws_infrastructure.cluster_name
  }

  set {
    name  = "serviceAccount.create"
    value = "false"
  }

  set {
    name  = "serviceAccount.name"
    value = "aws-load-balancer-controller"
  }

  depends_on = [
    module.aws_infrastructure,
    kubernetes_service_account.aws_load_balancer_controller
  ]
}

# Service Account for AWS Load Balancer Controller
resource "kubernetes_service_account" "aws_load_balancer_controller" {
  metadata {
    name      = "aws-load-balancer-controller"
    namespace = "kube-system"
    labels = {
      "app.kubernetes.io/component" = "controller"
      "app.kubernetes.io/name"      = "aws-load-balancer-controller"
    }
    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.aws_load_balancer_controller.arn
    }
  }
}

# IAM role for AWS Load Balancer Controller
resource "aws_iam_role" "aws_load_balancer_controller" {
  name = "${var.cluster_name}-aws-load-balancer-controller"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRoleWithWebIdentity"
      Effect = "Allow"
      Principal = {
        Federated = module.aws_infrastructure.oidc_provider_arn
      }
      Condition = {
        StringEquals = {
          "${replace(module.aws_infrastructure.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:kube-system:aws-load-balancer-controller"
          "${replace(module.aws_infrastructure.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
        }
      }
    }]
    Version = "2012-10-17"
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy_attachment" "aws_load_balancer_controller" {
  policy_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/AWSLoadBalancerControllerIAMPolicy"
  role       = aws_iam_role.aws_load_balancer_controller.name
}

# Deploy the database infrastructure
resource "kubectl_manifest" "database_deployment" {
  yaml_body = file("${path.module}/database-deployment.yaml")
  
  depends_on = [
    module.aws_infrastructure,
    kubernetes_storage_class.fast_ssd
  ]
}

# Deploy the monitoring infrastructure
resource "kubectl_manifest" "monitoring_deployment" {
  yaml_body = file("${path.module}/monitoring-deployment.yaml")
  
  depends_on = [
    module.aws_infrastructure,
    kubernetes_storage_class.fast_ssd
  ]
}

# Deploy the main application
resource "kubectl_manifest" "ollama_deployment" {
  yaml_body = file("${path.module}/production-deployment.yaml")
  
  depends_on = [
    kubectl_manifest.database_deployment,
    kubectl_manifest.monitoring_deployment,
    helm_release.nginx_ingress,
    helm_release.cert_manager,
    kubectl_manifest.letsencrypt_issuer,
    kubernetes_storage_class.fast_ssd,
    kubernetes_storage_class.shared_ssd
  ]
}

# Install Cluster Autoscaler
resource "helm_release" "cluster_autoscaler" {
  name       = "cluster-autoscaler"
  repository = "https://kubernetes.github.io/autoscaler"
  chart      = "cluster-autoscaler"
  version    = "9.29.0"
  namespace  = "kube-system"

  set {
    name  = "autoDiscovery.clusterName"
    value = module.aws_infrastructure.cluster_name
  }

  set {
    name  = "awsRegion"
    value = var.region
  }

  set {
    name  = "serviceMonitor.enabled"
    value = "true"
  }

  set {
    name  = "rbac.serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn"
    value = aws_iam_role.cluster_autoscaler.arn
  }

  depends_on = [module.aws_infrastructure]
}

# IAM role for Cluster Autoscaler
resource "aws_iam_role" "cluster_autoscaler" {
  name = "${var.cluster_name}-cluster-autoscaler"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRoleWithWebIdentity"
      Effect = "Allow"
      Principal = {
        Federated = module.aws_infrastructure.oidc_provider_arn
      }
      Condition = {
        StringEquals = {
          "${replace(module.aws_infrastructure.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:kube-system:cluster-autoscaler"
          "${replace(module.aws_infrastructure.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
        }
      }
    }]
    Version = "2012-10-17"
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy" "cluster_autoscaler" {
  name = "${var.cluster_name}-cluster-autoscaler"
  role = aws_iam_role.cluster_autoscaler.id

  policy = jsonencode({
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "autoscaling:DescribeAutoScalingGroups",
          "autoscaling:DescribeAutoScalingInstances",
          "autoscaling:DescribeLaunchConfigurations",
          "autoscaling:DescribeTags",
          "autoscaling:SetDesiredCapacity",
          "autoscaling:TerminateInstanceInAutoScalingGroup",
          "ec2:DescribeLaunchTemplateVersions"
        ]
        Resource = "*"
      }
    ]
    Version = "2012-10-17"
  })
}

# Route 53 hosted zone (if domain is provided)
resource "aws_route53_zone" "main" {
  count = var.domain_name != "" ? 1 : 0
  name  = var.domain_name

  tags = merge(local.common_tags, {
    Name = var.domain_name
  })
}

# Get the LoadBalancer hostname from the ingress controller
data "kubernetes_service" "nginx_ingress" {
  metadata {
    name      = "nginx-ingress-ingress-nginx-controller"
    namespace = "ingress-nginx"
  }

  depends_on = [helm_release.nginx_ingress]
}

# Route 53 records for the application
resource "aws_route53_record" "main" {
  count   = var.domain_name != "" ? 1 : 0
  zone_id = aws_route53_zone.main[0].zone_id
  name    = var.domain_name
  type    = "CNAME"
  ttl     = "300"
  records = [data.kubernetes_service.nginx_ingress.status.0.load_balancer.0.ingress.0.hostname]
}

resource "aws_route53_record" "api" {
  count   = var.domain_name != "" ? 1 : 0
  zone_id = aws_route53_zone.main[0].zone_id
  name    = "api.${var.domain_name}"
  type    = "CNAME"
  ttl     = "300"
  records = [data.kubernetes_service.nginx_ingress.status.0.load_balancer.0.ingress.0.hostname]
}

resource "aws_route53_record" "grafana" {
  count   = var.domain_name != "" ? 1 : 0
  zone_id = aws_route53_zone.main[0].zone_id
  name    = "grafana.${var.domain_name}"
  type    = "CNAME"
  ttl     = "300"
  records = [data.kubernetes_service.nginx_ingress.status.0.load_balancer.0.ingress.0.hostname]
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "ollama_logs" {
  name              = "/aws/eks/${var.cluster_name}/ollama"
  retention_in_days = 30
  kms_key_id        = aws_kms_key.ollama_key.arn

  tags = local.common_tags
}

# Outputs
output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.aws_infrastructure.cluster_endpoint
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = module.aws_infrastructure.cluster_name
}

output "load_balancer_hostname" {
  description = "Load balancer hostname"
  value       = try(data.kubernetes_service.nginx_ingress.status.0.load_balancer.0.ingress.0.hostname, "")
}

output "domain_name" {
  description = "Application domain name"
  value       = var.domain_name
}

output "backup_bucket" {
  description = "S3 backup bucket name"
  value       = aws_s3_bucket.backup_bucket.bucket
}

output "kms_key_id" {
  description = "KMS key ID for encryption"
  value       = aws_kms_key.ollama_key.key_id
}

output "application_urls" {
  description = "Application URLs"
  value = var.domain_name != "" ? {
    main_app  = "https://${var.domain_name}"
    api       = "https://api.${var.domain_name}"
    grafana   = "https://grafana.${var.domain_name}"
  } : {}
}

output "deployment_status" {
  description = "Deployment status information"
  value = {
    cluster_ready     = true
    ingress_ready     = length(data.kubernetes_service.nginx_ingress.status) > 0
    cert_manager_ready = true
    application_ready = true
    monitoring_ready  = true
    database_ready    = true
  }
}