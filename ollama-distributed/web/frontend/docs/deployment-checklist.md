# Production Deployment Checklist

## Pre-Deployment Validation

### Code Quality ✓
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code coverage >80%
- [ ] No linting errors
- [ ] Type checking passes
- [ ] Security scan completed
- [ ] Performance benchmarks met

### Infrastructure ✓
- [ ] Docker images built and tested
- [ ] Kubernetes manifests validated
- [ ] Database migrations prepared
- [ ] Backup systems configured
- [ ] Monitoring setup complete
- [ ] Logging infrastructure ready

### Security ✓
- [ ] SSL certificates installed
- [ ] Security headers configured
- [ ] CSP policies defined
- [ ] Rate limiting enabled
- [ ] Authentication/authorization tested
- [ ] Vulnerability scan passed

## Deployment Process

### Blue-Green Deployment
1. **Prepare Green Environment**
   ```bash
   kubectl apply -f k8s/green/
   kubectl wait --for=condition=ready pod -l app=ollama-green
   ```

2. **Run Health Checks**
   ```bash
   ./scripts/health-check.sh green
   ```

3. **Switch Traffic**
   ```bash
   kubectl patch service ollama-lb -p '{"spec":{"selector":{"version":"green"}}}'
   ```

4. **Monitor Metrics**
   - Response times <200ms
   - Error rate <0.1%
   - CPU usage <70%
   - Memory usage <80%

5. **Rollback if Needed**
   ```bash
   kubectl patch service ollama-lb -p '{"spec":{"selector":{"version":"blue"}}}'
   ```

### Canary Deployment
1. **Deploy Canary Version**
   ```bash
   kubectl apply -f k8s/canary/
   kubectl scale deployment ollama-canary --replicas=1
   ```

2. **Route 10% Traffic**
   ```yaml
   spec:
     rules:
     - http:
         paths:
         - backend:
             service:
               name: ollama-canary
               port:
                 number: 80
           weight: 10
         - backend:
             service:
               name: ollama-stable
               port:
                 number: 80
           weight: 90
   ```

3. **Monitor Canary Metrics**
   - Compare error rates
   - Check latency differences
   - Monitor resource usage

4. **Gradual Rollout**
   - 10% → 25% → 50% → 100%
   - Monitor at each stage

## Post-Deployment Validation

### Functional Tests
- [ ] Critical user journeys working
- [ ] API endpoints responding
- [ ] WebSocket connections stable
- [ ] Database queries optimized
- [ ] Cache hit rates normal

### Performance Tests
- [ ] Load testing passed
- [ ] Response times acceptable
- [ ] Resource usage within limits
- [ ] No memory leaks detected
- [ ] Database connection pooling working

### Security Validation
- [ ] Authentication working
- [ ] Authorization enforced
- [ ] Rate limiting active
- [ ] HTTPS redirect working
- [ ] Security headers present

## Monitoring & Alerting

### Key Metrics
- **Application Metrics**
  - Request rate
  - Response time (p50, p95, p99)
  - Error rate
  - Active users

- **Infrastructure Metrics**
  - CPU usage
  - Memory usage
  - Disk I/O
  - Network throughput

- **Business Metrics**
  - User registrations
  - API usage
  - Model inference count
  - Cost per operation

### Alert Thresholds
- Error rate >1% → Page oncall
- Response time p95 >500ms → Warning
- CPU usage >85% → Scale up
- Memory usage >90% → Investigate
- Disk usage >80% → Cleanup

## Disaster Recovery

### Backup Strategy
- **Database**: Hourly snapshots, 30-day retention
- **File Storage**: Daily backups, 90-day retention
- **Configuration**: Git versioned, replicated
- **Secrets**: Encrypted, multi-region storage

### Recovery Procedures
1. **Database Failure**
   ```bash
   kubectl exec -it postgres-0 -- pg_restore /backup/latest.dump
   ```

2. **Application Failure**
   ```bash
   kubectl rollout undo deployment/ollama-frontend
   ```

3. **Complete System Failure**
   - Provision new infrastructure
   - Restore from backups
   - Verify data integrity
   - Run validation tests

### RTO/RPO Targets
- **RTO** (Recovery Time Objective): <4 hours
- **RPO** (Recovery Point Objective): <1 hour

## Communication Plan

### Stakeholder Notification
1. **Pre-Deployment**
   - Email: 24 hours before
   - Slack: 1 hour before
   - Status page: Update scheduled maintenance

2. **During Deployment**
   - Slack: Real-time updates
   - Status page: In progress

3. **Post-Deployment**
   - Email: Success/failure summary
   - Slack: Metrics report
   - Status page: Operational

### Incident Response
- **Severity 1**: Page oncall + engineering lead
- **Severity 2**: Notify team channel
- **Severity 3**: Create ticket for next sprint

## Go-Live Checklist

### Final Validation
- [ ] All environments tested
- [ ] Performance benchmarks met
- [ ] Security scan clean
- [ ] Documentation updated
- [ ] Runbooks prepared

### Team Readiness
- [ ] Oncall schedule confirmed
- [ ] War room setup
- [ ] Rollback plan tested
- [ ] Communication channels open
- [ ] Monitoring dashboards ready

### Business Readiness
- [ ] Customer support briefed
- [ ] Marketing materials ready
- [ ] Legal compliance verified
- [ ] Analytics tracking enabled
- [ ] Feedback channels open

## Sign-off

- [ ] Engineering Lead: _________________
- [ ] Product Manager: _________________
- [ ] Security Team: _________________
- [ ] Operations Team: _________________
- [ ] Business Stakeholder: _________________

Date: _________________

## Post-Launch Review

### Success Metrics (Day 1)
- [ ] Uptime >99.9%
- [ ] Error rate <0.5%
- [ ] User complaints <5
- [ ] Performance SLAs met
- [ ] No security incidents

### Success Metrics (Week 1)
- [ ] User adoption on track
- [ ] System stability maintained
- [ ] Cost projections accurate
- [ ] Team velocity normal
- [ ] Customer satisfaction positive

### Lessons Learned
- What went well:
- What could improve:
- Action items for next deployment: