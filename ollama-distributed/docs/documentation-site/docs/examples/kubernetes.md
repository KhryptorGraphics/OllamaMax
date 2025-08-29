# Kubernetes Deployment

Deploy OllamaMax on Kubernetes for production scalability.

## Kubernetes Manifests

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollamamax-cluster
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ollamamax
  template:
    metadata:
      labels:
        app: ollamamax
    spec:
      containers:
      - name: ollamamax
        image: ollamamax/ollama-distributed:latest
        ports:
        - containerPort: 8081
```

See the [Deployment Overview](../deployment/overview.md) for more details.