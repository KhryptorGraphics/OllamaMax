# Ollama Frontend Infrastructure Modules

# Namespace Module
module "namespace" {
  source = "./modules/namespace"
  
  name        = var.kubernetes_namespace
  environment = var.environment
  
  labels = local.common_tags
}

# Blue-Green Deployment Module
module "blue_green_deployment" {
  source = "./modules/blue-green"
  
  namespace           = module.namespace.name
  app_name           = local.app_name
  environment        = var.environment
  
  # Application configuration
  app_image          = var.app_image
  app_version        = var.app_version
  app_port           = var.app_port
  
  # Deployment colors
  active_color       = var.active_deployment_color
  inactive_color     = local.inactive_color
  
  # Scaling configuration
  min_replicas       = var.min_replicas
  max_replicas       = var.max_replicas
  
  # Resource configuration
  cpu_request        = var.cpu_request
  cpu_limit          = var.cpu_limit
  memory_request     = var.memory_request
  memory_limit       = var.memory_limit
  
  # Auto-scaling configuration
  target_cpu_utilization    = var.target_cpu_utilization
  target_memory_utilization = var.target_memory_utilization
  
  # Canary deployment
  enable_canary      = var.enable_canary
  canary_weight      = var.canary_weight
  
  depends_on = [module.namespace]
}

# Monitoring Module
module "monitoring" {
  source = "./modules/monitoring"
  
  namespace           = module.namespace.name
  environment         = var.environment
  
  # Prometheus configuration
  enable_prometheus   = var.enable_prometheus
  prometheus_retention = var.prometheus_retention
  
  # Grafana configuration
  enable_grafana      = var.enable_grafana
  
  # Jaeger configuration
  enable_jaeger       = var.enable_jaeger
  
  # Application monitoring
  app_name           = local.app_name
  app_port           = var.app_port
  
  depends_on = [module.namespace]
}

# Logging Module
module "logging" {
  source = "./modules/logging"
  
  namespace           = module.namespace.name
  environment         = var.environment
  
  # ELK configuration
  enable_elk          = var.enable_elk
  log_retention_days  = var.log_retention_days
  
  depends_on = [module.namespace]
}

# Security Module
module "security" {
  source = "./modules/security"
  
  namespace                    = module.namespace.name
  environment                  = var.environment
  
  # Security policies
  enable_pod_security_policy   = var.enable_pod_security_policy
  enable_network_policy        = var.enable_network_policy
  enable_admission_controller  = var.enable_admission_controller
  
  depends_on = [module.namespace]
}

# Backup and Disaster Recovery Module
module "backup_dr" {
  source = "./modules/backup-dr"
  
  namespace              = module.namespace.name
  environment            = var.environment
  
  # Backup configuration
  backup_schedule        = var.backup_schedule
  backup_retention_days  = var.backup_retention_days
  
  # DR configuration
  enable_multi_region    = var.enable_multi_region
  dr_region             = var.dr_region
  rto_minutes           = var.rto_minutes
  rpo_minutes           = var.rpo_minutes
  
  depends_on = [module.namespace]
}

# Ingress and Load Balancer Module
module "ingress" {
  source = "./modules/ingress"
  
  namespace           = module.namespace.name
  environment         = var.environment
  app_name           = local.app_name
  
  # Load balancer configuration
  load_balancer_type  = var.load_balancer_type
  ssl_cert_arn       = var.ssl_cert_arn
  domain_name        = var.domain_name
  
  # Blue-green service endpoints
  blue_service_name   = module.blue_green_deployment.blue_service_name
  green_service_name  = module.blue_green_deployment.green_service_name
  active_color       = var.active_deployment_color
  
  # Canary configuration
  enable_canary      = var.enable_canary
  canary_weight      = var.canary_weight
  
  depends_on = [module.blue_green_deployment]
}

# Auto-scaling Module
module "autoscaling" {
  source = "./modules/autoscaling"
  
  namespace           = module.namespace.name
  environment         = var.environment
  app_name           = local.app_name
  
  # Scaling configuration
  min_replicas                  = var.min_replicas
  max_replicas                  = var.max_replicas
  target_cpu_utilization        = var.target_cpu_utilization
  target_memory_utilization     = var.target_memory_utilization
  
  # Deployment references
  blue_deployment_name  = module.blue_green_deployment.blue_deployment_name
  green_deployment_name = module.blue_green_deployment.green_deployment_name
  active_color         = var.active_deployment_color
  
  depends_on = [module.blue_green_deployment]
}

# Alerting Module
module "alerting" {
  source = "./modules/alerting"
  
  namespace           = module.namespace.name
  environment         = var.environment
  app_name           = local.app_name
  
  # Alert configuration
  slack_webhook_url   = var.slack_webhook_url
  pagerduty_key      = var.pagerduty_key
  email_alerts       = var.email_alerts
  
  depends_on = [module.monitoring]
}