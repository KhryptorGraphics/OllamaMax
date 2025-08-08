# OllamaMax Cluster Terraform Module
# Reusable module for deploying OllamaMax clusters across cloud providers

terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
}

# Variables
variable "cluster_name" {
  description = "Name of the OllamaMax cluster"
  type        = string
  default     = "ollama-cluster"
}

variable "namespace" {
  description = "Kubernetes namespace for OllamaMax"
  type        = string
  default     = "ollama-system"
}

variable "node_count" {
  description = "Number of OllamaMax nodes"
  type        = number
  default     = 3
}

variable "image_tag" {
  description = "OllamaMax container image tag"
  type        = string
  default     = "latest"
}

variable "storage_class" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "standard"
}

variable "storage_size" {
  description = "Storage size for each node"
  type        = string
  default     = "100Gi"
}

variable "enable_monitoring" {
  description = "Enable Prometheus monitoring"
  type        = bool
  default     = true
}

variable "enable_ingress" {
  description = "Enable ingress controller"
  type        = bool
  default     = true
}

variable "domain_name" {
  description = "Domain name for ingress"
  type        = string
  default     = ""
}

variable "tls_enabled" {
  description = "Enable TLS for ingress"
  type        = bool
  default     = true
}

variable "resource_limits" {
  description = "Resource limits for OllamaMax pods"
  type = object({
    cpu_request    = string
    memory_request = string
    cpu_limit      = string
    memory_limit   = string
  })
  default = {
    cpu_request    = "500m"
    memory_request = "1Gi"
    cpu_limit      = "2"
    memory_limit   = "4Gi"
  }
}

# Kubernetes namespace
resource "kubernetes_namespace" "ollama" {
  metadata {
    name = var.namespace
    labels = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
      "app.kubernetes.io/version"  = var.image_tag
    }
  }
}

# ConfigMap for OllamaMax configuration
resource "kubernetes_config_map" "ollama_config" {
  metadata {
    name      = "ollama-config"
    namespace = kubernetes_namespace.ollama.metadata[0].name
  }

  data = {
    "config.yaml" = yamlencode({
      node = {
        name     = "$${HOSTNAME}"
        data_dir = "/app/data"
      }
      api = {
        listen_address   = ":8080"
        enable_cors      = true
        request_timeout  = "30s"
      }
      web = {
        enabled        = true
        listen_address = ":8081"
        static_path    = "/app/web"
      }
      p2p = {
        listen_address   = "/ip4/0.0.0.0/tcp/9000"
        bootstrap_peers  = []
      }
      consensus = {
        algorithm           = "raft"
        election_timeout    = "5s"
        heartbeat_interval  = "1s"
      }
      scheduler = {
        algorithm              = "round_robin"
        health_check_interval  = "30s"
        max_retries           = 3
      }
      models = {
        storage_path = "/app/models"
        cache_size   = "1GB"
        auto_pull    = true
      }
      logging = {
        level  = "info"
        format = "json"
        output = "stdout"
      }
      security = {
        enabled         = true
        jwt_secret      = "auto-generated"
        session_timeout = "24h"
      }
      performance = {
        optimization_enabled = true
        monitoring_enabled   = true
      }
    })
  }
}

# Secret for sensitive configuration
resource "kubernetes_secret" "ollama_secrets" {
  metadata {
    name      = "ollama-secrets"
    namespace = kubernetes_namespace.ollama.metadata[0].name
  }

  data = {
    jwt_secret = base64encode("change-this-in-production")
  }

  type = "Opaque"
}

# StatefulSet for OllamaMax nodes
resource "kubernetes_stateful_set" "ollama" {
  metadata {
    name      = "ollama-distributed"
    namespace = kubernetes_namespace.ollama.metadata[0].name
    labels = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
      "app.kubernetes.io/version"  = var.image_tag
    }
  }

  spec {
    service_name = "ollama-headless"
    replicas     = var.node_count

    selector {
      match_labels = {
        "app.kubernetes.io/name"     = "ollama-distributed"
        "app.kubernetes.io/instance" = var.cluster_name
      }
    }

    template {
      metadata {
        labels = {
          "app.kubernetes.io/name"     = "ollama-distributed"
          "app.kubernetes.io/instance" = var.cluster_name
          "app.kubernetes.io/version"  = var.image_tag
        }
      }

      spec {
        container {
          name  = "ollama-distributed"
          image = "ghcr.io/khryptorgraphics/ollama-distributed:${var.image_tag}"

          port {
            name           = "api"
            container_port = 8080
            protocol       = "TCP"
          }

          port {
            name           = "web"
            container_port = 8081
            protocol       = "TCP"
          }

          port {
            name           = "p2p"
            container_port = 9000
            protocol       = "TCP"
          }

          port {
            name           = "metrics"
            container_port = 9090
            protocol       = "TCP"
          }

          env {
            name = "HOSTNAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }

          env {
            name  = "OLLAMA_CONFIG_FILE"
            value = "/app/config/config.yaml"
          }

          env {
            name  = "OLLAMA_DATA_DIR"
            value = "/app/data"
          }

          env {
            name  = "OLLAMA_MODELS_DIR"
            value = "/app/models"
          }

          volume_mount {
            name       = "config"
            mount_path = "/app/config"
            read_only  = true
          }

          volume_mount {
            name       = "data"
            mount_path = "/app/data"
          }

          volume_mount {
            name       = "models"
            mount_path = "/app/models"
          }

          resources {
            requests = {
              cpu    = var.resource_limits.cpu_request
              memory = var.resource_limits.memory_request
            }
            limits = {
              cpu    = var.resource_limits.cpu_limit
              memory = var.resource_limits.memory_limit
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 10
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }
        }

        volume {
          name = "config"
          config_map {
            name = kubernetes_config_map.ollama_config.metadata[0].name
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "data"
      }
      spec {
        access_modes       = ["ReadWriteOnce"]
        storage_class_name = var.storage_class
        resources {
          requests = {
            storage = var.storage_size
          }
        }
      }
    }

    volume_claim_template {
      metadata {
        name = "models"
      }
      spec {
        access_modes       = ["ReadWriteOnce"]
        storage_class_name = var.storage_class
        resources {
          requests = {
            storage = var.storage_size
          }
        }
      }
    }
  }
}

# Headless service for StatefulSet
resource "kubernetes_service" "ollama_headless" {
  metadata {
    name      = "ollama-headless"
    namespace = kubernetes_namespace.ollama.metadata[0].name
    labels = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
    }
  }

  spec {
    cluster_ip = "None"
    selector = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
    }

    port {
      name        = "api"
      port        = 8080
      target_port = 8080
      protocol    = "TCP"
    }

    port {
      name        = "web"
      port        = 8081
      target_port = 8081
      protocol    = "TCP"
    }

    port {
      name        = "p2p"
      port        = 9000
      target_port = 9000
      protocol    = "TCP"
    }
  }
}

# LoadBalancer service for external access
resource "kubernetes_service" "ollama_lb" {
  metadata {
    name      = "ollama-loadbalancer"
    namespace = kubernetes_namespace.ollama.metadata[0].name
    labels = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
    }
  }

  spec {
    type = "LoadBalancer"
    selector = {
      "app.kubernetes.io/name"     = "ollama-distributed"
      "app.kubernetes.io/instance" = var.cluster_name
    }

    port {
      name        = "api"
      port        = 80
      target_port = 8080
      protocol    = "TCP"
    }

    port {
      name        = "web"
      port        = 8081
      target_port = 8081
      protocol    = "TCP"
    }
  }
}

# Outputs
output "namespace" {
  description = "Kubernetes namespace"
  value       = kubernetes_namespace.ollama.metadata[0].name
}

output "service_name" {
  description = "LoadBalancer service name"
  value       = kubernetes_service.ollama_lb.metadata[0].name
}

output "cluster_name" {
  description = "Cluster name"
  value       = var.cluster_name
}
