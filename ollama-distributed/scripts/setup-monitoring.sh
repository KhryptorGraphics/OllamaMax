#!/bin/bash

# OllamaMax Monitoring Setup Script
# Sets up comprehensive monitoring for deployed environments

set -e

echo "📊 OllamaMax Monitoring Setup"
echo "============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
ENVIRONMENT=""
MONITORING_TYPE="prometheus"
GRAFANA_ENABLED=true
ALERTING_ENABLED=true
SLACK_WEBHOOK=""
EMAIL_ALERTS=""

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --type)
                MONITORING_TYPE="$2"
                shift 2
                ;;
            --no-grafana)
                GRAFANA_ENABLED=false
                shift
                ;;
            --no-alerting)
                ALERTING_ENABLED=false
                shift
                ;;
            --slack-webhook)
                SLACK_WEBHOOK="$2"
                shift 2
                ;;
            --email-alerts)
                EMAIL_ALERTS="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Show help
show_help() {
    cat << EOF
OllamaMax Monitoring Setup Script

Usage: $0 --environment ENV [OPTIONS]

Options:
    --environment ENV      Target environment (staging, production) [REQUIRED]
    --type TYPE           Monitoring type (prometheus, datadog, newrelic)
    --no-grafana          Skip Grafana dashboard setup
    --no-alerting         Skip alerting configuration
    --slack-webhook URL   Slack webhook URL for notifications
    --email-alerts EMAIL  Email address for alerts
    --help                Show this help message

Examples:
    $0 --environment staging
    $0 --environment production --slack-webhook https://hooks.slack.com/...
    $0 --environment staging --type prometheus --email-alerts admin@company.com
EOF
}

# Validate arguments
validate_args() {
    if [ -z "$ENVIRONMENT" ]; then
        print_status "ERROR" "Environment is required. Use --environment staging|production"
        exit 1
    fi

    if [[ ! "$ENVIRONMENT" =~ ^(staging|production)$ ]]; then
        print_status "ERROR" "Invalid environment: $ENVIRONMENT. Must be staging or production"
        exit 1
    fi
}

# Setup Prometheus monitoring
setup_prometheus() {
    print_status "INFO" "Setting up Prometheus monitoring..."

    # Create monitoring configuration directory
    local config_dir="monitoring/$ENVIRONMENT"
    mkdir -p "$config_dir"

    # Create Prometheus configuration
    cat > "$config_dir/prometheus.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'ollama-distributed'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
    
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']
    
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

    # Create alert rules
    cat > "$config_dir/alert_rules.yml" << EOF
groups:
  - name: ollama-distributed
    rules:
      - alert: OllamaDistributedDown
        expr: up{job="ollama-distributed"} == 0
        for: 1m
        labels:
          severity: critical
          environment: $ENVIRONMENT
        annotations:
          summary: "OllamaMax distributed system is down"
          description: "OllamaMax distributed system has been down for more than 1 minute in {{ \$labels.environment }}"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ \$value }} errors per second in {{ \$labels.environment }}"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ \$value }}s in {{ \$labels.environment }}"

      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "High memory usage"
          description: "Memory usage is above 90% in {{ \$labels.environment }}"

      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "High CPU usage"
          description: "CPU usage is above 80% in {{ \$labels.environment }}"

      - alert: PerformanceRegression
        expr: increase(http_request_duration_seconds{quantile="0.95"}[10m]) > 0.5
        for: 3m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "Performance regression detected"
          description: "95th percentile latency increased by more than 500ms in {{ \$labels.environment }}"

      - alert: ThroughputDrop
        expr: rate(http_requests_total[5m]) < 0.5 * rate(http_requests_total[30m] offset 1h)
        for: 5m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "Throughput drop detected"
          description: "Request throughput dropped significantly in {{ \$labels.environment }}"

      - alert: HighGoroutineCount
        expr: go_goroutines > 1000
        for: 5m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "High goroutine count"
          description: "Goroutine count is above 1000 in {{ \$labels.environment }}"

      - alert: MemoryLeakSuspected
        expr: increase(go_memstats_alloc_bytes[1h]) > 100000000
        for: 10m
        labels:
          severity: warning
          environment: $ENVIRONMENT
        annotations:
          summary: "Potential memory leak"
          description: "Memory allocation increased by more than 100MB in 1 hour in {{ \$labels.environment }}"
EOF

    # Create Alertmanager configuration
    cat > "$config_dir/alertmanager.yml" << EOF
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@ollamamax.com'

route:
  group_by: ['alertname', 'environment']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        environment: production
      receiver: 'production-alerts'

receivers:
  - name: 'default'
    webhook_configs:
      - url: 'http://localhost:5001/'

  - name: 'critical-alerts'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK}'
        channel: '#alerts-critical'
        title: 'Critical Alert - {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
    email_configs:
      - to: '${EMAIL_ALERTS}'
        subject: 'Critical Alert: {{ .GroupLabels.alertname }}'
        body: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'production-alerts'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK}'
        channel: '#alerts-production'
        title: 'Production Alert - {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
EOF

    print_status "SUCCESS" "Prometheus configuration created"
}

# Setup Grafana dashboards
setup_grafana() {
    if [ "$GRAFANA_ENABLED" = false ]; then
        print_status "INFO" "Grafana setup skipped"
        return
    fi

    print_status "INFO" "Setting up Grafana dashboards..."

    local config_dir="monitoring/$ENVIRONMENT/grafana"
    mkdir -p "$config_dir/dashboards"
    mkdir -p "$config_dir/provisioning/dashboards"
    mkdir -p "$config_dir/provisioning/datasources"

    # Create datasource configuration
    cat > "$config_dir/provisioning/datasources/prometheus.yml" << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Create dashboard provisioning
    cat > "$config_dir/provisioning/dashboards/dashboards.yml" << EOF
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF

    # Create OllamaMax performance dashboard
    cat > "$config_dir/dashboards/ollama-performance.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "OllamaMax Performance Monitoring",
    "tags": ["ollama", "performance", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate & Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "Requests/sec"
          },
          {
            "expr": "rate(ollama_proxy_requests_total[5m])",
            "legendFormat": "Proxy Requests/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
        "yAxes": [{"label": "Requests/sec"}]
      },
      {
        "id": 2,
        "title": "Response Time Percentiles",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "99th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.90, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "90th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
        "yAxes": [{"label": "Seconds"}]
      },
      {
        "id": 3,
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
            "legendFormat": "CPU Usage %"
          },
          {
            "expr": "rate(process_cpu_seconds_total[5m]) * 100",
            "legendFormat": "Process CPU %"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8},
        "yAxes": [{"label": "Percentage", "max": 100}]
      },
      {
        "id": 4,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "go_memstats_alloc_bytes / 1024 / 1024",
            "legendFormat": "Allocated Memory (MB)"
          },
          {
            "expr": "go_memstats_heap_inuse_bytes / 1024 / 1024",
            "legendFormat": "Heap In Use (MB)"
          },
          {
            "expr": "go_memstats_stack_inuse_bytes / 1024 / 1024",
            "legendFormat": "Stack In Use (MB)"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8},
        "yAxes": [{"label": "MB"}]
      },
      {
        "id": 5,
        "title": "Goroutines & GC",
        "type": "graph",
        "targets": [
          {
            "expr": "go_goroutines",
            "legendFormat": "Goroutines"
          },
          {
            "expr": "rate(go_gc_duration_seconds_count[5m])",
            "legendFormat": "GC Rate"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 16}
      },
      {
        "id": 6,
        "title": "Network I/O",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(node_network_receive_bytes_total[5m]) / 1024 / 1024",
            "legendFormat": "Network In (MB/s)"
          },
          {
            "expr": "rate(node_network_transmit_bytes_total[5m]) / 1024 / 1024",
            "legendFormat": "Network Out (MB/s)"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 16},
        "yAxes": [{"label": "MB/s"}]
      },
      {
        "id": 7,
        "title": "Error Rates",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"4..\"}[5m])",
            "legendFormat": "4xx errors/sec"
          },
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 24}
      },
      {
        "id": 8,
        "title": "Performance Optimization Status",
        "type": "stat",
        "targets": [
          {
            "expr": "ollama_performance_optimization_active",
            "legendFormat": "Optimization Active"
          },
          {
            "expr": "ollama_performance_score",
            "legendFormat": "Performance Score"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 24}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF

    print_status "SUCCESS" "Grafana dashboards created"
}

# Setup Docker Compose for monitoring stack
setup_docker_compose() {
    print_status "INFO" "Creating Docker Compose monitoring stack..."

    local config_dir="monitoring/$ENVIRONMENT"
    
    cat > "$config_dir/docker-compose.yml" << EOF
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus-$ENVIRONMENT
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - ./alert_rules.yml:/etc/prometheus/alert_rules.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: alertmanager-$ENVIRONMENT
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: grafana-$ENVIRONMENT
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
      - GF_USERS_ALLOW_SIGN_UP=false
    restart: unless-stopped

  node-exporter:
    image: prom/node-exporter:latest
    container_name: node-exporter-$ENVIRONMENT
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    restart: unless-stopped

volumes:
  prometheus_data:
  alertmanager_data:
  grafana_data:
EOF

    print_status "SUCCESS" "Docker Compose configuration created"
}

# Create monitoring startup script
create_startup_script() {
    print_status "INFO" "Creating monitoring startup script..."

    local config_dir="monitoring/$ENVIRONMENT"
    
    cat > "$config_dir/start-monitoring.sh" << EOF
#!/bin/bash

echo "🚀 Starting OllamaMax monitoring stack for $ENVIRONMENT..."

# Start monitoring services
docker-compose up -d

# Wait for services to start
sleep 10

# Check service status
echo "📊 Monitoring Services Status:"
docker-compose ps

echo ""
echo "🎯 Access URLs:"
echo "  Prometheus: http://localhost:9090"
echo "  Grafana:    http://localhost:3000 (admin/admin123)"
echo "  Alertmanager: http://localhost:9093"

echo ""
echo "✅ Monitoring stack started successfully!"
EOF

    chmod +x "$config_dir/start-monitoring.sh"
    
    print_status "SUCCESS" "Startup script created"
}

# Setup monitoring alerts
setup_alerts() {
    if [ "$ALERTING_ENABLED" = false ]; then
        print_status "INFO" "Alerting setup skipped"
        return
    fi

    print_status "INFO" "Setting up monitoring alerts..."

    # Create alert testing script
    local config_dir="monitoring/$ENVIRONMENT"
    
    cat > "$config_dir/test-alerts.sh" << EOF
#!/bin/bash

echo "🧪 Testing monitoring alerts..."

# Test Slack webhook if configured
if [ -n "$SLACK_WEBHOOK" ]; then
    echo "Testing Slack webhook..."
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"🧪 Test alert from OllamaMax monitoring setup"}' \
        "$SLACK_WEBHOOK"
    echo "Slack test sent"
fi

# Test Prometheus alert rules
echo "Validating Prometheus alert rules..."
docker run --rm -v \$(pwd)/alert_rules.yml:/tmp/alert_rules.yml \
    prom/prometheus:latest promtool check rules /tmp/alert_rules.yml

echo "✅ Alert testing completed"
EOF

    chmod +x "$config_dir/test-alerts.sh"
    
    print_status "SUCCESS" "Alert testing script created"
}

# Main function
main() {
    parse_args "$@"
    validate_args
    
    print_status "INFO" "Setting up monitoring for environment: $ENVIRONMENT"
    
    setup_prometheus
    setup_grafana
    setup_docker_compose
    create_startup_script
    setup_alerts
    
    print_status "SUCCESS" "Monitoring setup completed! 🎉"
    
    echo ""
    echo "📋 Next Steps:"
    echo "1. cd monitoring/$ENVIRONMENT"
    echo "2. ./start-monitoring.sh"
    echo "3. Access Grafana at http://localhost:3000"
    echo "4. ./test-alerts.sh (to test alerting)"
    echo ""
    echo "🔧 Configuration files created:"
    echo "  - monitoring/$ENVIRONMENT/prometheus.yml"
    echo "  - monitoring/$ENVIRONMENT/alert_rules.yml"
    echo "  - monitoring/$ENVIRONMENT/alertmanager.yml"
    echo "  - monitoring/$ENVIRONMENT/docker-compose.yml"
    echo ""
    print_status "SUCCESS" "Monitoring is ready for deployment! 📊"
}

# Run main function
main "$@"
