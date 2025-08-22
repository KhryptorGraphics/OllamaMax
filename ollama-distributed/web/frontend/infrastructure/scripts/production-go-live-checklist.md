# Ollama Frontend - Production Go-Live Checklist
## Comprehensive validation and deployment guide

### [DEPLOY-STARTED] Pre-Deployment Phase

#### Infrastructure Validation
- [ ] **1.1 Kubernetes Cluster Health**
  ```bash
  kubectl cluster-info
  kubectl get nodes
  kubectl top nodes
  kubectl get events --sort-by=.metadata.creationTimestamp | tail -20
  ```
  - All nodes in Ready state
  - No critical events in last 24 hours
  - Resource utilization < 70%

- [ ] **1.2 Namespace and RBAC**
  ```bash
  kubectl get namespace ollama-frontend
  kubectl get serviceaccount -n ollama-frontend
  kubectl auth can-i create pods --as=system:serviceaccount:ollama-frontend:ollama-frontend -n ollama-frontend
  ```
  - Namespace exists with proper labels
  - Service accounts configured
  - RBAC permissions validated

- [ ] **1.3 Network Policies**
  ```bash
  kubectl get networkpolicy -n ollama-frontend
  kubectl describe networkpolicy ollama-frontend-network-policy -n ollama-frontend
  ```
  - Network policies active
  - Ingress/egress rules configured
  - DNS resolution working

- [ ] **1.4 Persistent Storage**
  ```bash
  kubectl get pv | grep ollama
  kubectl get pvc -n ollama-frontend
  kubectl describe storageclass gp3
  ```
  - Storage classes available
  - PVCs bound successfully
  - Backup storage configured

#### Security Validation
- [ ] **2.1 Pod Security Standards**
  ```bash
  kubectl get podsecuritypolicy
  kubectl describe podsecuritypolicy ollama-frontend-psp
  ```
  - Non-root containers enforced
  - ReadOnlyRootFilesystem enabled
  - Privilege escalation disabled

- [ ] **2.2 Secrets Management**
  ```bash
  kubectl get secrets -n ollama-frontend
  kubectl describe secret ollama-frontend-secrets -n ollama-frontend
  ```
  - Secrets properly created
  - No hardcoded credentials
  - External secrets operator working

- [ ] **2.3 Container Image Security**
  ```bash
  # Run security scan on production images
  trivy image ollamamax/frontend:latest
  docker scout cves ollamamax/frontend:latest
  ```
  - No critical vulnerabilities
  - Images signed and verified
  - Base images up to date

#### Application Dependencies
- [ ] **3.1 Backend Services**
  ```bash
  kubectl get svc -n ollama-backend
  curl -f http://ollama-backend.ollama-backend.svc.cluster.local:8080/health
  ```
  - Backend services healthy
  - API endpoints responding
  - Database connections active

- [ ] **3.2 External Dependencies**
  ```bash
  # Test external API connectivity
  curl -f https://api.example.com/health
  nslookup external-service.com
  ```
  - External APIs accessible
  - DNS resolution working
  - CDN endpoints active

### [DEPLOY-STARTED] Deployment Phase

#### Blue-Green Deployment
- [ ] **4.1 Pre-deployment Backup**
  ```bash
  # Create backup before deployment
  velero backup create pre-deployment-$(date +%Y%m%d-%H%M%S) \
    --include-namespaces ollama-frontend \
    --wait
  ```

- [ ] **4.2 Deploy New Version (Green)**
  ```bash
  # Update image tag for green deployment
  kubectl set image deployment/ollama-frontend-green \
    ollama-frontend=ollamamax/frontend:v1.2.3 \
    -n ollama-frontend
  
  # Monitor deployment
  kubectl rollout status deployment/ollama-frontend-green -n ollama-frontend --timeout=300s
  ```

- [ ] **4.3 Health Check Green Deployment**
  ```bash
  # Wait for pods to be ready
  kubectl wait --for=condition=ready pod -l app=ollama-frontend,version=green -n ollama-frontend --timeout=300s
  
  # Test green deployment internally
  kubectl port-forward svc/ollama-frontend-green 8080:3000 -n ollama-frontend &
  curl -f http://localhost:8080/health
  curl -f http://localhost:8080/ready
  ```

- [ ] **4.4 Switch Traffic (Blue → Green)**
  ```bash
  # Update active service to point to green
  kubectl patch service ollama-frontend-active -n ollama-frontend \
    -p '{"spec":{"selector":{"version":"green"}}}'
  
  # Verify traffic switch
  kubectl get service ollama-frontend-active -n ollama-frontend -o yaml | grep version
  ```

#### Monitoring and Alerting
- [ ] **5.1 Prometheus Targets**
  ```bash
  # Check Prometheus targets
  curl -s http://prometheus.monitoring.svc.cluster.local:9090/api/v1/targets | \
    jq '.data.activeTargets[] | select(.labels.job=="ollama-frontend")'
  ```
  - All targets healthy
  - Metrics being scraped
  - No target errors

- [ ] **5.2 Grafana Dashboards**
  ```bash
  # Verify Grafana access
  curl -f http://grafana.monitoring.svc.cluster.local:3000/api/health
  ```
  - Dashboards loading
  - Data visualization working
  - Alerts configured

- [ ] **5.3 Log Aggregation**
  ```bash
  # Check Elasticsearch cluster health
  curl -X GET "elasticsearch.logging.svc.cluster.local:9200/_cluster/health"
  
  # Verify log ingestion
  curl -X GET "elasticsearch.logging.svc.cluster.local:9200/ollama-frontend-*/_search" \
    -H 'Content-Type: application/json' \
    -d '{"size": 1, "sort": [{"@timestamp": {"order": "desc"}}]}'
  ```

#### Load Balancer and DNS
- [ ] **6.1 Ingress Controller**
  ```bash
  kubectl get ingress -n ollama-frontend
  kubectl describe ingress ollama-frontend-ingress -n ollama-frontend
  ```
  - Ingress rules active
  - SSL certificates valid
  - Load balancer healthy

- [ ] **6.2 DNS Resolution**
  ```bash
  nslookup ollama.example.com
  dig +short ollama.example.com
  ```
  - DNS records propagated
  - TTL values appropriate
  - CDN integration working

### [DEPLOY-STARTED] Post-Deployment Validation

#### Application Health
- [ ] **7.1 Endpoint Testing**
  ```bash
  # Test critical endpoints
  curl -f -H "Accept: application/json" https://ollama.example.com/health
  curl -f -H "Accept: application/json" https://ollama.example.com/api/v1/status
  curl -f https://ollama.example.com/ready
  ```

- [ ] **7.2 Authentication Flow**
  ```bash
  # Test login endpoint
  curl -X POST https://ollama.example.com/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}' \
    -w "%{http_code}"
  ```

- [ ] **7.3 Performance Metrics**
  ```bash
  # Check response times
  curl -w "@curl-format.txt" -o /dev/null -s https://ollama.example.com/
  ```
  - Response time < 2 seconds
  - No 5xx errors
  - Error rate < 0.1%

#### Business Validation
- [ ] **8.1 Critical User Journeys**
  - [ ] User registration flow
  - [ ] Login/logout functionality
  - [ ] Main dashboard loading
  - [ ] Data retrieval and display
  - [ ] Form submissions

- [ ] **8.2 Cross-browser Testing**
  - [ ] Chrome (latest)
  - [ ] Firefox (latest)
  - [ ] Safari (latest)
  - [ ] Edge (latest)
  - [ ] Mobile browsers

#### Performance Validation
- [ ] **9.1 Load Testing**
  ```bash
  # Run load test with k6
  k6 run --vus 100 --duration 5m performance-test.js
  
  # Monitor during load test
  kubectl top pods -n ollama-frontend --sort-by=cpu
  ```

- [ ] **9.2 Core Web Vitals**
  ```bash
  # Run Lighthouse CI
  lhci autorun --upload.target=temporary-public-storage
  ```
  - LCP < 2.5 seconds
  - FID < 100ms
  - CLS < 0.1

#### Security Validation
- [ ] **10.1 HTTPS Configuration**
  ```bash
  # Test SSL configuration
  testssl.sh https://ollama.example.com
  
  # Check security headers
  curl -I https://ollama.example.com
  ```
  - SSL Labs grade A or better
  - Security headers present
  - No mixed content warnings

- [ ] **10.2 Vulnerability Scan**
  ```bash
  # OWASP ZAP baseline scan
  docker run -t owasp/zap2docker-stable zap-baseline.py \
    -t https://ollama.example.com
  ```

### [DEPLOY-STARTED] Auto-scaling and Resilience

#### Auto-scaling Configuration
- [ ] **11.1 Horizontal Pod Autoscaler**
  ```bash
  kubectl get hpa -n ollama-frontend
  kubectl describe hpa ollama-frontend-green-hpa -n ollama-frontend
  ```
  - HPA active and responding
  - Metrics collection working
  - Scaling thresholds appropriate

- [ ] **11.2 Cluster Autoscaler**
  ```bash
  kubectl get nodes
  kubectl describe configmap cluster-autoscaler-config -n kube-system
  ```
  - Node groups configured
  - Auto-scaling policies active
  - Resource limits set

#### Disaster Recovery
- [ ] **12.1 Backup Verification**
  ```bash
  velero backup get
  velero backup describe latest-backup --details
  ```
  - Recent backups successful
  - Backup retention policies active
  - Cross-region replication working

- [ ] **12.2 Failover Testing**
  ```bash
  # Test health check endpoints
  kubectl get configmap failover-config -n ollama-frontend -o yaml
  ```
  - Health checks configured
  - Failover triggers set
  - DR runbooks updated

### [DEPLOY-SUCCESS] Go-Live Validation

#### Final Checks
- [ ] **13.1 Monitoring Dashboards**
  - [ ] Application metrics showing
  - [ ] No critical alerts firing
  - [ ] Log aggregation working
  - [ ] Trace collection active

- [ ] **13.2 Business Metrics**
  - [ ] User sessions tracking
  - [ ] Conversion tracking
  - [ ] Error tracking
  - [ ] Performance metrics

- [ ] **13.3 Rollback Readiness**
  - [ ] Blue deployment still available
  - [ ] Rollback procedures documented
  - [ ] Emergency contacts notified
  - [ ] Incident response ready

#### Communication
- [ ] **14.1 Stakeholder Notification**
  - [ ] Engineering team notified
  - [ ] Product team informed
  - [ ] Customer support briefed
  - [ ] Executive dashboard updated

- [ ] **14.2 Documentation Updates**
  - [ ] Deployment notes recorded
  - [ ] Runbooks updated
  - [ ] Architecture diagrams current
  - [ ] Troubleshooting guides updated

### [DEPLOY-SUCCESS] Production Monitoring

#### First 24 Hours
- [ ] **15.1 Continuous Monitoring**
  ```bash
  # Monitor key metrics every 15 minutes
  watch -n 900 'kubectl top pods -n ollama-frontend; curl -s https://ollama.example.com/health'
  ```

- [ ] **15.2 Alert Response**
  - [ ] On-call rotation active
  - [ ] Escalation procedures ready
  - [ ] Incident response tested
  - [ ] Communication channels open

### Rollback Procedures

#### Emergency Rollback
```bash
# Immediate traffic switch back to blue
kubectl patch service ollama-frontend-active -n ollama-frontend \
  -p '{"spec":{"selector":{"version":"blue"}}}'

# Or use automated rollback
kubectl create job --from=cronjob/automated-rollback-check \
  emergency-rollback-$(date +%Y%m%d-%H%M%S) -n ollama-frontend

# Verify rollback
kubectl get service ollama-frontend-active -n ollama-frontend -o yaml | grep version
curl -f https://ollama.example.com/health
```

#### Post-Rollback Actions
- [ ] Analyze failure root cause
- [ ] Update incident response documentation
- [ ] Plan remediation steps
- [ ] Schedule post-incident review

### Sign-off

#### Technical Sign-off
- [ ] **Platform Engineering**: _________________ Date: _______
- [ ] **Security Team**: _________________ Date: _______
- [ ] **Site Reliability Engineering**: _________________ Date: _______

#### Business Sign-off
- [ ] **Product Owner**: _________________ Date: _______
- [ ] **Engineering Manager**: _________________ Date: _______
- [ ] **Release Manager**: _________________ Date: _______

---

## Emergency Contacts

| Role | Primary | Secondary |
|------|---------|-----------|
| On-Call Engineer | +1-XXX-XXX-XXXX | +1-XXX-XXX-XXXX |
| Platform Team Lead | +1-XXX-XXX-XXXX | +1-XXX-XXX-XXXX |
| Product Manager | +1-XXX-XXX-XXXX | +1-XXX-XXX-XXXX |

## Key Metrics Thresholds

| Metric | Green | Yellow | Red |
|--------|-------|---------|-----|
| Response Time (p95) | < 500ms | < 1s | > 2s |
| Error Rate | < 0.1% | < 1% | > 5% |
| CPU Utilization | < 60% | < 80% | > 90% |
| Memory Utilization | < 70% | < 85% | > 95% |
| Active Connections | < 1000 | < 5000 | > 10000 |

## Additional Resources

- [Incident Response Playbook](./incident-response-playbook.md)
- [Architecture Documentation](./architecture-overview.md)
- [Monitoring Runbooks](./monitoring-runbooks.md)
- [Security Compliance Guide](./security-compliance.md)