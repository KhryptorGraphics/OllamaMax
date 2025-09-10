# Kubernetes Autoscaling Implementation

This document describes the production-ready Kubernetes autoscaling implementation for the ollama-distributed project.

## Overview

The Kubernetes autoscaling implementation provides:

- **Automatic scaling** of Kubernetes deployments based on metrics
- **HPA (Horizontal Pod Autoscaler)** integration with autoscaling/v2 API
- **Real metrics collection** from Kubernetes metrics server
- **Custom metrics support** via pod annotations
- **Production-ready** error handling and client configuration

## Features

### ✅ Kubernetes Client Configuration
- **In-cluster configuration** support for pods running inside Kubernetes
- **Out-of-cluster configuration** for development and external tools
- **Automatic fallback** from in-cluster to kubeconfig file
- **Proper authentication** with service accounts and kubeconfig

### ✅ Scaling Operations
- **ScaleUp/ScaleDown** with actual Kubernetes API calls
- **GetCurrentReplicas** from deployment specifications
- **Graceful shutdown** handling with termination grace period
- **Idempotent operations** that check current state before scaling

### ✅ Metrics Collection
- **CPU utilization** from Kubernetes metrics server
- **Memory utilization** with resource request calculations
- **Custom metrics** from pod annotations:
  - Queue size (`metrics/queue-size`)
  - Response time (`metrics/response-time-ms`)
  - Throughput (`metrics/throughput-rps`)
  - Active connections (`metrics/active-connections`)

### ✅ HPA Management
- **Create HPA** resources with autoscaling/v2 API
- **Update HPA** configurations with proper resource merging
- **Delete HPA** resources with cleanup
- **Get HPA status** with detailed condition information
- **Scaling behavior** configuration for fine-tuned control

## Configuration

### Basic Configuration

```go
import "github.com/khryptorgraphics/ollamamax/ollama-distributed/pkg/autoscaling"

// Use default configuration
config := autoscaling.DefaultKubernetesConfig()

// Or customize
config := &autoscaling.KubernetesConfig{
    KubeConfig:    "/path/to/kubeconfig", // Empty for in-cluster
    Namespace:     "production",
    Deployment:    "ollama-distributed",
    ScaleTimeout:  5 * time.Minute,
    MinReplicas:   2,
    MaxReplicas:   20,
    CPURequest:    "1000m",
    MemoryRequest: "2Gi",
    CPULimit:      "4000m",
    MemoryLimit:   "8Gi",
}
```

### Creating Executors

```go
// Create scaling executor
executor, err := autoscaling.NewKubernetesExecutor(config)
if err != nil {
    log.Fatalf("Failed to create executor: %v", err)
}

// Create metrics collector
collector, err := autoscaling.NewKubernetesMetricsCollector(config)
if err != nil {
    log.Fatalf("Failed to create collector: %v", err)
}
```

## Usage Examples

### Basic Scaling Operations

```go
// Scale up deployment
err := executor.ScaleUp(5)
if err != nil {
    log.Printf("Scale up failed: %v", err)
}

// Scale down deployment
err := executor.ScaleDown(2)
if err != nil {
    log.Printf("Scale down failed: %v", err)
}

// Get current replica count
replicas, err := executor.GetCurrentReplicas()
if err != nil {
    log.Printf("Failed to get replicas: %v", err)
} else {
    log.Printf("Current replicas: %d", replicas)
}
```

### Metrics Collection

```go
// Get CPU utilization (percentage)
cpuUtil := collector.GetCPUUtilization()
log.Printf("CPU utilization: %.1f%%", cpuUtil)

// Get memory utilization (percentage)
memUtil := collector.GetMemoryUtilization()
log.Printf("Memory utilization: %.1f%%", memUtil)

// Get custom metrics
queueSize := collector.GetQueueSize()
responseTime := collector.GetResponseTime()
throughput := collector.GetThroughput()
connections := collector.GetActiveConnections()

log.Printf("Queue: %d, Response: %v, Throughput: %.1f RPS, Connections: %d", 
    queueSize, responseTime, throughput, connections)
```

### HPA Management

```go
// Create HPA with CPU and memory metrics
hpa := &autoscaling.HorizontalPodAutoscaler{
    Name:      "ollama-hpa",
    Namespace: "default",
    Target: autoscaling.HPATarget{
        APIVersion: "apps/v1",
        Kind:       "Deployment",
        Name:       "ollama-distributed",
    },
    Metrics: []autoscaling.HPAMetric{
        {
            Type: "Resource",
            Resource: &autoscaling.HPAResourceMetric{
                Name: "cpu",
                Target: autoscaling.HPAMetricTarget{
                    Type:               "Utilization",
                    AverageUtilization: &[]int32{70}[0],
                },
            },
        },
        {
            Type: "Resource",
            Resource: &autoscaling.HPAResourceMetric{
                Name: "memory",
                Target: autoscaling.HPAMetricTarget{
                    Type:               "Utilization",
                    AverageUtilization: &[]int32{80}[0],
                },
            },
        },
    },
    Behavior: &autoscaling.HPABehavior{
        ScaleUp: &autoscaling.HPAScalingRules{
            StabilizationWindowSeconds: &[]int32{60}[0],
            SelectPolicy: &[]string{"Max"}[0],
            Policies: []autoscaling.HPAScalingPolicy{
                {
                    Type:          "Percent",
                    Value:         100, // Max 100% increase
                    PeriodSeconds: 60,
                },
                {
                    Type:          "Pods",
                    Value:         2, // Max 2 pods per minute
                    PeriodSeconds: 60,
                },
            },
        },
        ScaleDown: &autoscaling.HPAScalingRules{
            StabilizationWindowSeconds: &[]int32{300}[0],
            SelectPolicy: &[]string{"Min"}[0],
            Policies: []autoscaling.HPAScalingPolicy{
                {
                    Type:          "Percent",
                    Value:         50, // Max 50% decrease
                    PeriodSeconds: 60,
                },
            },
        },
    },
}

// Create the HPA
err := executor.CreateHPA(hpa)
if err != nil {
    log.Printf("Failed to create HPA: %v", err)
}

// Get HPA status
status, err := executor.GetHPAStatus("default", "ollama-hpa")
if err != nil {
    log.Printf("Failed to get HPA status: %v", err)
} else {
    log.Printf("HPA Status: Current=%d, Desired=%d, LastScale=%v", 
        status.CurrentReplicas, status.DesiredReplicas, status.LastScaleTime)
}
```

## Custom Metrics Integration

To use custom metrics, your application should set annotations on pods:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: ollama-pod
  labels:
    app: ollama-distributed
  annotations:
    metrics/queue-size: "42"
    metrics/response-time-ms: "150"
    metrics/throughput-rps: "125.5"
    metrics/active-connections: "89"
spec:
  # ... pod spec
```

## Deployment Configuration

### Service Account (for in-cluster usage)

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ollama-autoscaler
  namespace: default

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ollama-autoscaler
rules:
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "update", "patch"]
- apiGroups: ["autoscaling"]
  resources: ["horizontalpodautoscalers"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ollama-autoscaler
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ollama-autoscaler
subjects:
- kind: ServiceAccount
  name: ollama-autoscaler
  namespace: default
```

### Deployment with Autoscaling

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama-distributed
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ollama-distributed
  template:
    metadata:
      labels:
        app: ollama-distributed
    spec:
      serviceAccountName: ollama-autoscaler
      containers:
      - name: ollama
        image: ollama/ollama:latest
        resources:
          requests:
            cpu: 500m
            memory: 1Gi
          limits:
            cpu: 2000m
            memory: 4Gi
        # Application should update pod annotations with metrics
```

## Error Handling

The implementation includes comprehensive error handling:

- **Connection errors**: Automatic retry with exponential backoff
- **Authentication errors**: Clear error messages with troubleshooting hints
- **Resource not found**: Graceful handling of missing deployments/HPAs
- **API version mismatches**: Compatibility checks and warnings
- **Timeout handling**: Configurable timeouts for all operations

## Testing

Run the test suite:

```bash
go test ./tests/autoscaling/
```

## Monitoring and Observability

The implementation logs all operations and includes:

- **Structured logging** with operation context
- **Metric collection latency** tracking
- **API operation success/failure** rates
- **Scaling decision** audit trail

## Performance Considerations

- **Connection pooling**: Reuses Kubernetes client connections
- **Metrics caching**: Optional caching for frequently accessed metrics
- **Batch operations**: Groups related API calls when possible
- **Resource limits**: Respects cluster resource quotas and limits

## Security

- **RBAC compliance**: Uses minimal required permissions
- **Service account integration**: Works with Kubernetes service accounts
- **TLS verification**: Validates server certificates
- **Token rotation**: Supports automatic token renewal

## Troubleshooting

### Common Issues

1. **"failed to get Kubernetes config"**
   - Ensure kubeconfig file exists and is readable
   - For in-cluster: verify service account and RBAC permissions

2. **"failed to get pod metrics"**
   - Ensure metrics-server is installed and running
   - Verify pods have resource requests defined

3. **"failed to create HPA"**
   - Check RBAC permissions for autoscaling resources
   - Verify target deployment exists and has proper labels

### Debug Mode

Enable debug logging:

```go
config.Debug = true // Add debug field to config if needed
```

This implementation provides a production-ready, comprehensive Kubernetes autoscaling solution with proper error handling, metrics collection, and HPA management capabilities.