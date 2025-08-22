# Ollama Frontend Infrastructure Variables

# Environment Configuration
variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
  
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "region" {
  description = "Primary deployment region"
  type        = string
  default     = "us-west-2"
}

# Kubernetes Configuration
variable "kubernetes_namespace" {
  description = "Kubernetes namespace for deployment"
  type        = string
  default     = "ollama-frontend"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kubernetes_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = "prod-cluster"
}

# Blue-Green Deployment
variable "active_deployment_color" {
  description = "Active deployment color (blue or green)"
  type        = string
  default     = "blue"
  
  validation {
    condition     = contains(["blue", "green"], var.active_deployment_color)
    error_message = "Active deployment color must be blue or green."
  }
}

# Application Configuration
variable "app_image" {
  description = "Container image for the frontend application"
  type        = string
  default     = "ollamamax/frontend"
}

variable "app_version" {
  description = "Application version/tag"
  type        = string
  default     = "latest"
}

variable "app_port" {
  description = "Application port"
  type        = number
  default     = 3000
}

# Scaling Configuration
variable "min_replicas" {
  description = "Minimum number of replicas"
  type        = number
  default     = 3
}

variable "max_replicas" {
  description = "Maximum number of replicas"
  type        = number
  default     = 50
}

variable "target_cpu_utilization" {
  description = "Target CPU utilization for autoscaling"
  type        = number
  default     = 70
}

variable "target_memory_utilization" {
  description = "Target memory utilization for autoscaling"
  type        = number
  default     = 80
}

# Resource Limits
variable "cpu_request" {
  description = "CPU request for containers"
  type        = string
  default     = "100m"
}

variable "cpu_limit" {
  description = "CPU limit for containers"
  type        = string
  default     = "500m"
}

variable "memory_request" {
  description = "Memory request for containers"
  type        = string
  default     = "128Mi"
}

variable "memory_limit" {
  description = "Memory limit for containers"
  type        = string
  default     = "512Mi"
}

# Load Balancer Configuration
variable "load_balancer_type" {
  description = "Type of load balancer (nlb, alb, classic)"
  type        = string
  default     = "alb"
}

variable "ssl_cert_arn" {
  description = "SSL certificate ARN"
  type        = string
  default     = ""
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "ollama.example.com"
}

# Monitoring Configuration
variable "enable_prometheus" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}

variable "enable_grafana" {
  description = "Enable Grafana dashboards"
  type        = bool
  default     = true
}

variable "enable_jaeger" {
  description = "Enable Jaeger tracing"
  type        = bool
  default     = true
}

variable "prometheus_retention" {
  description = "Prometheus data retention period"
  type        = string
  default     = "30d"
}

# Logging Configuration
variable "enable_elk" {
  description = "Enable ELK stack for logging"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Log retention in days"
  type        = number
  default     = 30
}

# Backup Configuration
variable "backup_schedule" {
  description = "Backup schedule in cron format"
  type        = string
  default     = "0 2 * * *"  # Daily at 2 AM
}

variable "backup_retention_days" {
  description = "Backup retention in days"
  type        = number
  default     = 90
}

# AWS Configuration
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "terraform_state_bucket" {
  description = "S3 bucket for Terraform state"
  type        = string
  default     = ""
}

variable "terraform_lock_table" {
  description = "DynamoDB table for Terraform state locking"
  type        = string
  default     = ""
}

# GCP Configuration (alternative cloud provider)
variable "gcp_project" {
  description = "GCP project ID"
  type        = string
  default     = ""
}

variable "gcp_region" {
  description = "GCP region"
  type        = string
  default     = "us-west1"
}

# Security Configuration
variable "enable_pod_security_policy" {
  description = "Enable Pod Security Policy"
  type        = bool
  default     = true
}

variable "enable_network_policy" {
  description = "Enable Network Policy"
  type        = bool
  default     = true
}

variable "enable_admission_controller" {
  description = "Enable admission controller"
  type        = bool
  default     = true
}

# Canary Deployment Configuration
variable "enable_canary" {
  description = "Enable canary deployment"
  type        = bool
  default     = false
}

variable "canary_weight" {
  description = "Canary deployment weight percentage"
  type        = number
  default     = 10
  
  validation {
    condition     = var.canary_weight >= 0 && var.canary_weight <= 100
    error_message = "Canary weight must be between 0 and 100."
  }
}

# Disaster Recovery Configuration
variable "enable_multi_region" {
  description = "Enable multi-region deployment"
  type        = bool
  default     = false
}

variable "dr_region" {
  description = "Disaster recovery region"
  type        = string
  default     = "us-east-1"
}

variable "rto_minutes" {
  description = "Recovery Time Objective in minutes"
  type        = number
  default     = 15
}

variable "rpo_minutes" {
  description = "Recovery Point Objective in minutes"
  type        = number
  default     = 5
}