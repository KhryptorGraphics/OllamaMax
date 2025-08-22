# Ollama Distributed Frontend - Infrastructure as Code
# Production-ready Terraform configuration for Kubernetes deployment

terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.27"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = var.terraform_state_bucket
    key    = "ollama-frontend/terraform.tfstate"
    region = var.aws_region
    
    dynamodb_table = var.terraform_lock_table
    encrypt        = true
  }
}

# Local variables
locals {
  app_name    = "ollama-frontend"
  environment = var.environment
  
  common_tags = {
    Application = local.app_name
    Environment = local.environment
    ManagedBy   = "terraform"
    Owner       = "platform-team"
    CostCenter  = "engineering"
  }
  
  # Blue-Green deployment configuration
  deployment_colors = ["blue", "green"]
  active_color     = var.active_deployment_color
  inactive_color   = local.active_color == "blue" ? "green" : "blue"
}

# Kubernetes provider configuration
provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kubernetes_context
}

# Helm provider configuration
provider "helm" {
  kubernetes {
    config_path    = var.kubeconfig_path
    config_context = var.kubernetes_context
  }
}

# AWS provider configuration (for EKS and services)
provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = local.common_tags
  }
}

# Google provider configuration (for GKE alternative)
provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
}

# Data sources
data "kubernetes_namespace" "ollama" {
  metadata {
    name = var.kubernetes_namespace
  }
  depends_on = [kubernetes_namespace.ollama]
}