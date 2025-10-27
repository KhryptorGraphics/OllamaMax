# Deployment Troubleshooting Guide

**Version:** 1.0
**Last Updated:** 2025
**References:** DOCKER_DEPLOYMENT.md, FINAL_DEPLOYMENT_GUIDE.md

## Table of Contents

1. [Docker Issues](#docker-issues)
2. [Kubernetes Issues](#kubernetes-issues)
3. [Performance Issues](#performance-issues)
4. [Security Issues](#security-issues)
5. [Networking Issues](#networking-issues)
6. [Data Issues](#data-issues)
7. [Monitoring Issues](#monitoring-issues)
8. [Diagnostic Commands](#diagnostic-commands)

---

## Docker Issues

### Services Won't Start

**Symptoms:**
- Container exits immediately
- Error: "port already in use"
- Error: "no such file or directory"

**Diagnosis:**
```bash
# Check Docker daemon status
sudo systemctl status docker

# Check available resources
docker system df
docker system info

# Check port conflicts
netstat -tuln | grep :8080
lsof -i :8080

# View container logs
docker compose logs <service-name>
docker logs <container-id> --tail 100
```

**Solutions:**

**Issue 1: Port conflict**
```bash
# Find process using port
sudo lsof -i :8080
sudo kill -9 <PID>

# Or change port in docker-compose.yml
ports:
  - "8081:8080"  # Use different host port
```

**Issue 2: Resource exhaustion**
```bash
# Clean up unused resources
docker system prune -af
docker volume prune -f

# Check disk space
df -h /var/lib/docker

# Increase Docker resources
# Edit Docker Desktop settings or daemon.json
```

**Issue 3: Missing dependencies**
```bash
# Check service dependencies
docker compose config --services
docker compose up --no-deps <service>

# Start dependencies first
docker compose up -d postgres redis
sleep 10
docker compose up -d api
```

### Health Checks Failing

**Symptoms:**
- Container starts but health check fails
- Status shows "unhealthy"

**Diagnosis:**
```bash
# Check health check logs
docker inspect <container-id> | jq '.[].State.Health'

# Test health endpoint manually
docker compose exec api curl -f http://localhost:8080/health

# Check application logs
docker compose logs api --tail 100
```

**Solutions:**

**Issue: Health endpoint not responding**
```bash
# Verify service is listening
docker compose exec api netstat -tuln | grep :8080

# Check application startup
docker compose logs api | grep -i "server listening"

# Increase health check timeout
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s  # Increase from 5s
  retries: 3
  start_period: 40s  # Increase from 30s
```

### Volume Mount Issues

**Symptoms:**
- Permission denied errors
- Data not persisting
- Files not visible in container

**Diagnosis:**
```bash
# Check volume mounts
docker inspect <container-id> | jq '.[].Mounts'

# Check permissions
docker compose exec <service> ls -la /data

# Verify volume exists
docker volume ls
docker volume inspect <volume-name>
```

**Solutions:**

**Issue: Permission denied**
```bash
# Check file ownership
docker compose exec <service> ls -la /data

# Fix permissions on host
sudo chown -R 1000:1000 ./data

# Or set user in docker-compose.yml
services:
  api:
    user: "1000:1000"
```

**Issue: Volume not mounted**
```bash
# Verify volume configuration
docker compose config | grep -A 10 volumes

# Recreate volume
docker compose down -v
docker volume rm <volume-name>
docker compose up -d
```

---

## Kubernetes Issues

### Pods Stuck in Pending

**Symptoms:**
- Pod status: Pending
- Events show scheduling issues

**Diagnosis:**
```bash
# Check pod status
kubectl get pods -n <namespace>
kubectl describe pod <pod-name> -n <namespace>

# Check node resources
kubectl top nodes
kubectl describe nodes

# Check PVC status
kubectl get pvc -n <namespace>
```

**Solutions:**

**Issue: Insufficient resources**
```bash
# Check resource requests
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 5 resources

# Scale down other workloads
kubectl scale deployment/<deployment> --replicas=1 -n <namespace>

# Add more nodes (cloud)
# AWS EKS:
eksctl scale nodegroup --cluster=<cluster> --nodes=5 <nodegroup>

# GKE:
gcloud container clusters resize <cluster> --num-nodes=5
```

**Issue: PVC not bound**
```bash
# Check PVC status
kubectl get pvc -n <namespace>

# Check StorageClass
kubectl get storageclass

# Manually provision PV or fix StorageClass
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: manual-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /mnt/data
EOF
```

### Pods CrashLoopBackOff

**Symptoms:**
- Pod status: CrashLoopBackOff
- Container restarts frequently

**Diagnosis:**
```bash
# Check pod logs
kubectl logs <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace> --previous

# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Check resource limits
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Limits
```

**Solutions:**

**Issue: Application error**
```bash
# View application logs
kubectl logs <pod-name> -n <namespace> --tail=100

# Check environment variables
kubectl get pod <pod-name> -n <namespace> -o yaml | grep -A 20 env

# Exec into container (if it stays up)
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh
```

**Issue: Resource limits too low**
```bash
# Increase resource limits
kubectl edit deployment <deployment> -n <namespace>

# Or patch deployment
kubectl patch deployment <deployment> -n <namespace> -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "api",
          "resources": {
            "limits": {
              "memory": "2Gi",
              "cpu": "1000m"
            }
          }
        }]
      }
    }
  }
}'
```

### Service Not Accessible

**Symptoms:**
- Cannot reach service
- Connection timeout
- DNS resolution fails

**Diagnosis:**
```bash
# Check service endpoints
kubectl get svc -n <namespace>
kubectl get endpoints <service> -n <namespace>

# Check network policies
kubectl get networkpolicies -n <namespace>

# Test from within cluster
kubectl run -it --rm debug --image=busybox --restart=Never -- sh
# Inside pod:
nslookup <service>.<namespace>.svc.cluster.local
wget -O- http://<service>.<namespace>:8080/health
```

**Solutions:**

**Issue: No endpoints**
```bash
# Check selector labels
kubectl get svc <service> -n <namespace> -o yaml | grep selector
kubectl get pods -n <namespace> --show-labels

# Fix label mismatch
kubectl label pod <pod-name> app=<label> -n <namespace>
```

**Issue: Network policy blocking**
```bash
# List network policies
kubectl get networkpolicies -n <namespace>

# Temporarily delete to test
kubectl delete networkpolicy <policy-name> -n <namespace>

# Fix policy
kubectl edit networkpolicy <policy-name> -n <namespace>
```

---

## Performance Issues

### Slow Response Times

**Symptoms:**
- API response time >500ms
- Page load time >3s
- Database queries slow

**Diagnosis:**
```bash
# Measure response time
time curl http://localhost:8080/api/users

# Check resource usage
docker stats
kubectl top pods -n <namespace>

# Database query performance
docker compose exec postgres psql -U postgres -c "\
SELECT query, calls, mean_exec_time \
FROM pg_stat_statements \
ORDER BY mean_exec_time DESC \
LIMIT 10;"

# Network latency
ping -c 10 api.example.com
traceroute api.example.com
```

**Solutions:**

**Issue: High CPU usage**
```bash
# Identify CPU-intensive processes
docker stats --no-stream | sort -k 3 -hr

# Scale horizontally
docker compose up -d --scale api=3

# Kubernetes
kubectl scale deployment/api --replicas=5 -n <namespace>

# Optimize code (profile application)
# Node.js:
docker compose exec api node --prof app.js
# Go:
kubectl exec <pod> -n <namespace> -- curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
```

**Issue: Slow database queries**
```bash
# Enable query logging
docker compose exec postgres psql -U postgres -c "\
ALTER SYSTEM SET log_min_duration_statement = 100;"

# Analyze slow queries
docker compose exec postgres psql -U postgres -c "\
SELECT * FROM pg_stat_activity \
WHERE state = 'active' AND query_start < now() - interval '5 seconds';"

# Add indexes
docker compose exec postgres psql -U postgres -d <database> -c "\
CREATE INDEX idx_users_email ON users(email);"

# Vacuum database
docker compose exec postgres vacuumdb -U postgres -d <database> --analyze
```

### High Memory Usage

**Symptoms:**
- OOM (Out of Memory) errors
- Container restarts due to memory
- Swap usage high

**Diagnosis:**
```bash
# Check memory usage
docker stats
kubectl top pods -n <namespace>

# Check memory limits
docker inspect <container-id> | jq '.[].HostConfig.Memory'
kubectl describe pod <pod-name> -n <namespace> | grep -A 5 Limits

# Application memory profiling
# Node.js heap dump:
docker compose exec api kill -USR2 <pid>
# Go memory profile:
kubectl exec <pod> -n <namespace> -- curl http://localhost:6060/debug/pprof/heap > mem.prof
```

**Solutions:**

**Issue: Memory leak**
```bash
# Identify memory leak
# Monitor memory over time
watch -n 5 'docker stats --no-stream | grep api'

# Restart service temporarily
docker compose restart api

# Fix in code (use profiling tools)
# Node.js: node --inspect
# Go: pprof tool
# Python: memory_profiler
```

**Issue: Insufficient memory limits**
```bash
# Increase memory limits
# Docker Compose:
services:
  api:
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G

# Kubernetes:
kubectl patch deployment api -n <namespace> -p '
{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "api",
          "resources": {
            "limits": {"memory": "2Gi"},
            "requests": {"memory": "1Gi"}
          }
        }]
      }
    }
  }
}'
```

---

## Security Issues

### Authentication Failures

**Symptoms:**
- 401 Unauthorized errors
- JWT token invalid
- Session expired

**Diagnosis:**
```bash
# Check JWT secret
docker compose exec api env | grep JWT_SECRET

# Test authentication endpoint
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Check logs
docker compose logs api | grep -i auth
```

**Solutions:**

**Issue: JWT secret mismatch**
```bash
# Verify JWT secret in all services
docker compose exec api env | grep JWT_SECRET
docker compose exec web env | grep JWT_SECRET

# Regenerate secret
JWT_SECRET=$(openssl rand -base64 64)
echo "JWT_SECRET=$JWT_SECRET" >> .env

# Restart services
docker compose restart api web
```

**Issue: Token expiration**
```bash
# Check token expiration
docker compose exec api env | grep JWT_EXPIRATION

# Increase expiration (if appropriate)
echo "JWT_EXPIRATION=3600" >> .env  # 1 hour
docker compose restart api
```

### TLS Errors

**Symptoms:**
- SSL certificate invalid
- HTTPS connection fails
- Certificate expired

**Diagnosis:**
```bash
# Check certificate validity
openssl s_client -connect localhost:443 -servername ollamamax.io

# Check certificate expiration
echo | openssl s_client -connect localhost:443 2>/dev/null | openssl x509 -noout -dates

# Test SSL configuration
curl -vI https://localhost:443
```

**Solutions:**

**Issue: Certificate expired**
```bash
# Renew Let's Encrypt certificate
certbot renew --force-renewal

# Or manually:
certbot certonly --standalone -d ollamamax.io -d www.ollamamax.io

# Copy new certificates
cp /etc/letsencrypt/live/ollamamax.io/fullchain.pem nginx/ssl/
cp /etc/letsencrypt/live/ollamamax.io/privkey.pem nginx/ssl/

# Reload nginx
docker compose exec nginx nginx -s reload
```

**Issue: Certificate mismatch**
```bash
# Verify certificate CN matches domain
openssl x509 -in nginx/ssl/fullchain.pem -noout -subject

# Generate new certificate for correct domain
certbot certonly --standalone -d correct-domain.com

# Update nginx configuration
nano nginx/nginx-production.conf
# Update ssl_certificate paths
docker compose restart nginx
```

---

## Networking Issues

### Services Can't Communicate

**Symptoms:**
- Connection refused between services
- DNS resolution fails
- Timeout errors

**Diagnosis:**
```bash
# Docker: Test connectivity
docker compose exec api ping -c 3 postgres
docker compose exec api nc -zv postgres 5432

# Check network
docker network ls
docker network inspect <network-name>

# Kubernetes: Test connectivity
kubectl exec -it <pod> -n <namespace> -- ping <service>
kubectl exec -it <pod> -n <namespace> -- nslookup <service>
```

**Solutions:**

**Issue: DNS resolution failure**
```bash
# Docker: Check network configuration
docker compose config | grep -A 10 networks

# Ensure services are on same network
services:
  api:
    networks:
      - app-network
  postgres:
    networks:
      - app-network

networks:
  app-network:
    driver: bridge

# Kubernetes: Check CoreDNS
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl logs -n kube-system -l k8s-app=kube-dns
```

**Issue: Firewall blocking**
```bash
# Check firewall rules
sudo iptables -L -n

# Allow Docker networks
sudo iptables -A INPUT -i docker0 -j ACCEPT

# Kubernetes: Check network policies
kubectl get networkpolicies --all-namespaces
kubectl describe networkpolicy <policy> -n <namespace>
```

### External Access Not Working

**Symptoms:**
- Cannot access from outside
- Load balancer not working
- Ingress not routing

**Diagnosis:**
```bash
# Docker: Check port mappings
docker compose ps
docker port <container-id>

# Kubernetes: Check services
kubectl get svc --all-namespaces
kubectl get ingress --all-namespaces

# Check load balancer status
kubectl describe svc <service> -n <namespace>
```

**Solutions:**

**Issue: Port not exposed**
```bash
# Docker: Add port mapping
services:
  api:
    ports:
      - "8080:8080"

# Kubernetes: Expose service
kubectl expose deployment api --type=LoadBalancer --port=8080 -n <namespace>
```

**Issue: Ingress misconfigured**
```bash
# Check ingress controller
kubectl get pods -n ingress-nginx

# Check ingress rules
kubectl get ingress <ingress> -n <namespace> -o yaml

# Test ingress
curl -H "Host: ollamamax.io" http://<ingress-ip>
```

---

## Data Issues

### Data Not Persisting

**Symptoms:**
- Data lost after restart
- Volume not mounted
- Database empty after restart

**Diagnosis:**
```bash
# Check volume mounts
docker volume ls
docker volume inspect <volume-name>

# Check data in container
docker compose exec postgres ls -la /var/lib/postgresql/data

# Kubernetes: Check PVC
kubectl get pvc -n <namespace>
kubectl describe pvc <pvc-name> -n <namespace>
```

**Solutions:**

**Issue: Volume not configured**
```bash
# Docker: Add volume
services:
  postgres:
    volumes:
      - postgres-data:/var/lib/postgresql/data

volumes:
  postgres-data:
    driver: local

# Kubernetes: Create PVC
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: ollamamax
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF
```

### Database Connection Errors

**Symptoms:**
- Cannot connect to database
- Connection pool exhausted
- Too many connections

**Diagnosis:**
```bash
# Test database connection
docker compose exec api psql postgresql://postgres:password@postgres:5432/ollamamax

# Check active connections
docker compose exec postgres psql -U postgres -c "\
SELECT count(*) FROM pg_stat_activity;"

# Check connection limit
docker compose exec postgres psql -U postgres -c "\
SHOW max_connections;"
```

**Solutions:**

**Issue: Connection pool exhausted**
```bash
# Increase connection pool size
# Node.js example:
const pool = new Pool({
  max: 20,  // Increase from 10
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
})

# Restart application
docker compose restart api
```

**Issue: Database connection limit reached**
```bash
# Increase max_connections
docker compose exec postgres psql -U postgres -c "\
ALTER SYSTEM SET max_connections = 200;"

# Restart PostgreSQL
docker compose restart postgres

# Kill idle connections
docker compose exec postgres psql -U postgres -c "\
SELECT pg_terminate_backend(pid) \
FROM pg_stat_activity \
WHERE state = 'idle' AND state_change < now() - interval '5 minutes';"
```

---

## Monitoring Issues

### Metrics Not Appearing

**Symptoms:**
- Grafana shows no data
- Prometheus targets down
- Missing metrics

**Diagnosis:**
```bash
# Check Prometheus targets
kubectl port-forward -n ollamamax-monitoring svc/prometheus 9090:9090
# Open http://localhost:9090/targets

# Check Prometheus config
kubectl get configmap prometheus-config -n ollamamax-monitoring -o yaml

# Check service discovery
kubectl get pods -n ollamamax --show-labels
```

**Solutions:**

**Issue: Scrape targets not configured**
```bash
# Update Prometheus configuration
kubectl edit configmap prometheus-config -n ollamamax-monitoring

# Add scrape config:
scrape_configs:
  - job_name: 'ollama-api'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
          - ollamamax
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: ollama-api

# Reload Prometheus
kubectl exec -n ollamamax-monitoring <prometheus-pod> -- kill -HUP 1
```

**Issue: Grafana datasource not configured**
```bash
# Check datasources
kubectl exec -n ollamamax-monitoring <grafana-pod> -- \
  curl -u admin:admin http://localhost:3000/api/datasources

# Add Prometheus datasource
kubectl exec -n ollamamax-monitoring <grafana-pod> -- \
  curl -X POST -u admin:admin http://localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Prometheus",
    "type":"prometheus",
    "url":"http://prometheus:9090",
    "access":"proxy",
    "isDefault":true
  }'
```

### Alerts Not Firing

**Symptoms:**
- No alerts received
- Alertmanager not working
- Notification channels not working

**Diagnosis:**
```bash
# Check alert rules
kubectl get prometheusrules -n ollamamax-monitoring

# Check Alertmanager
kubectl port-forward -n ollamamax-monitoring svc/alertmanager 9093:9093
# Open http://localhost:9093

# Check notification config
kubectl get secret alertmanager-config -n ollamamax-monitoring -o yaml
```

**Solutions:**

**Issue: Alert rules not loaded**
```bash
# Create/update alert rules
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: ollamamax-alerts
  namespace: ollamamax-monitoring
spec:
  groups:
    - name: ollamamax
      rules:
        - alert: HighErrorRate
          expr: rate(http_requests_total{status="500"}[5m]) > 0.05
          for: 5m
          annotations:
            summary: "High error rate detected"
EOF
```

**Issue: Notification channel misconfigured**
```bash
# Update Alertmanager config
kubectl edit secret alertmanager-config -n ollamamax-monitoring

# Example Slack configuration:
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
```

---

## Diagnostic Commands

### Docker Diagnostics

```bash
# System information
docker info
docker version

# Container inspection
docker inspect <container-id>
docker stats --no-stream
docker top <container-id>

# Logs
docker logs <container-id> --tail 100 --follow
docker logs <container-id> --since 10m

# Network debugging
docker network inspect bridge
docker exec <container-id> netstat -tuln
docker exec <container-id> ping -c 3 google.com

# Resource usage
docker system df
docker container ls --size
```

### Kubernetes Diagnostics

```bash
# Cluster information
kubectl cluster-info
kubectl get nodes -o wide
kubectl top nodes

# Pod inspection
kubectl describe pod <pod-name> -n <namespace>
kubectl get pod <pod-name> -n <namespace> -o yaml
kubectl top pod <pod-name> -n <namespace>

# Logs
kubectl logs <pod-name> -n <namespace> --tail=100 --follow
kubectl logs <pod-name> -c <container> -n <namespace> --previous

# Events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
kubectl describe pod <pod-name> -n <namespace> | grep -A 10 Events

# Network debugging
kubectl exec -it <pod-name> -n <namespace> -- /bin/sh
# Inside pod:
nslookup <service>
ping <service>
curl http://<service>:8080/health
```

### System-Level Diagnostics

```bash
# Resource usage
top
htop
free -h
df -h

# Network
netstat -tuln
ss -tuln
lsof -i :8080
tcpdump -i any port 8080

# Processes
ps aux | grep docker
ps aux | grep kubelet
systemctl status docker
systemctl status kubelet

# Logs
journalctl -u docker -n 100 --no-pager
journalctl -u kubelet -n 100 --no-pager
tail -f /var/log/syslog
```

---

## Getting Help

If issues persist after following this guide:

1. **Check Documentation:**
   - `/docs` directory
   - Deployment procedures
   - Rollback procedures

2. **Gather Diagnostic Information:**
   ```bash
   # Create diagnostic report
   bash scripts/generate-diagnostic-report.sh
   ```

3. **Contact Support:**
   - Create GitHub issue with diagnostic report
   - Include steps to reproduce
   - Attach relevant logs

4. **Emergency:**
   - Execute rollback: `bash ollama-distributed/scripts/rollback.sh`
   - Contact on-call engineer

---

**Tip:** Keep this guide updated with new issues and solutions encountered in production.
