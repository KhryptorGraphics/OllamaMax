# Rollback Procedures

**Version:** 1.0
**Last Updated:** 2025
**Reference:** ollama-distributed/scripts/rollback.sh

## Table of Contents

1. [Rollback Decision Criteria](#rollback-decision-criteria)
2. [Docker Rollback](#docker-rollback)
3. [Kubernetes Rollback](#kubernetes-rollback)
4. [Database Rollback](#database-rollback)
5. [Automated Rollback](#automated-rollback)
6. [Manual Rollback](#manual-rollback)
7. [Rollback Testing](#rollback-testing)
8. [Post-Rollback](#post-rollback)

---

## Rollback Decision Criteria

### When to Trigger Rollback

**Critical Issues (Immediate Rollback):**
- ❌ Service unavailable or total outage
- ❌ Data corruption detected
- ❌ Critical security vulnerability exposed
- ❌ Authentication/authorization failures
- ❌ Database connection failures

**High Priority Issues (Rollback within 15 minutes):**
- ⚠️ Performance degradation >50%
- ⚠️ Error rate >5%
- ⚠️ Failed health checks for >5 minutes
- ⚠️ Memory leaks causing OOM
- ⚠️ Critical feature broken

**Medium Priority Issues (Evaluate and decide):**
- 🔶 Performance degradation 20-50%
- 🔶 Error rate 1-5%
- 🔶 Non-critical feature broken
- 🔶 UI/UX issues
- 🔶 Warning-level monitoring alerts

### Rollback Authority

- **Production:** Requires approval from:
  - DevOps Lead
  - Engineering Manager
  - On-call Engineer (for critical issues)

- **Staging:** DevOps team decision

- **Development:** Any developer can rollback

---

## Docker Rollback

### Quick Rollback (Using Script)

```bash
# Execute rollback script
bash ollama-distributed/scripts/rollback.sh --docker

# With specific version
bash ollama-distributed/scripts/rollback.sh --docker --version v1.2.3

# Dry-run mode
bash ollama-distributed/scripts/rollback.sh --docker --dry-run
```

### Manual Step-by-Step Rollback

#### 1. Identify Current Version

```bash
# List current images
docker compose images

# Check container labels
docker inspect $(docker compose ps -q api) | jq '.[].Config.Labels.version'

# Review deployment history
git log --oneline -10
```

#### 2. Stop Current Deployment

```bash
# Stop all services gracefully
docker compose down

# Force stop if not responding
docker compose kill

# Verify all containers stopped
docker compose ps -a
```

#### 3. Restore Previous Version

```bash
# Checkout previous version
git checkout v1.2.3  # or commit hash

# Or use previous images
docker tag ollamamax-api:v1.2.4 ollamamax-api:rollback-backup
docker tag ollamamax-api:v1.2.3 ollamamax-api:latest

# Pull previous images
docker compose pull
```

#### 4. Start Previous Version

```bash
# Start services with previous configuration
docker compose -f docker-compose.prod.yml up -d

# Monitor startup
docker compose logs -f --tail=100

# Verify services started
docker compose ps
```

#### 5. Verify Service Health

```bash
# Run health checks
bash ollama-distributed/scripts/health-check.sh

# Test critical endpoints
curl -f http://localhost:8080/health
curl -f http://localhost:11434/api/tags

# Check logs for errors
docker compose logs --tail=50 | grep -i error
```

### Rollback Time Estimates

| Component | Estimated Time |
|-----------|---------------|
| Stop services | 1-2 minutes |
| Switch versions | 2-3 minutes |
| Start services | 2-3 minutes |
| Health checks | 1-2 minutes |
| **Total** | **6-10 minutes** |

---

## Kubernetes Rollback

### Quick Rollback

```bash
# Rollback to previous revision (all deployments)
kubectl rollout undo deployment/ollamamax-api -n ollamamax
kubectl rollout undo deployment/predictive-scaling-system -n ollamamax-ml

# Rollback all deployments in namespace
for deploy in $(kubectl get deployments -n ollamamax -o name); do
  kubectl rollout undo $deploy -n ollamamax
done
```

### Check Rollout History

```bash
# View deployment history
kubectl rollout history deployment/ollamamax-api -n ollamamax

# Output example:
# REVISION  CHANGE-CAUSE
# 1         Initial deployment
# 2         Update to v1.2.4
# 3         Update to v1.2.5 (CURRENT)

# View specific revision details
kubectl rollout history deployment/ollamamax-api -n ollamamax --revision=2
```

### Rollback to Specific Revision

```bash
# Rollback to revision 2
kubectl rollout undo deployment/ollamamax-api -n ollamamax --to-revision=2

# Monitor rollback progress
kubectl rollout status deployment/ollamamax-api -n ollamamax

# Verify rollback
kubectl describe deployment ollamamax-api -n ollamamax
```

### StatefulSet Rollback

```bash
# StatefulSets require more careful handling
# 1. Check current version
kubectl get statefulset redis-cluster -n ollamamax-redis -o yaml

# 2. Rollback (similar to deployment)
kubectl rollout undo statefulset/redis-cluster -n ollamamax-redis

# 3. Monitor pod recreation (StatefulSets recreate in order)
kubectl get pods -n ollamamax-redis -w

# 4. Verify data integrity
kubectl exec -n ollamamax-redis redis-cluster-0 -- redis-cli PING
```

### Rollback All Namespaces

```bash
#!/bin/bash
NAMESPACES=("ollamamax" "ollamamax-ml" "ollamamax-monitoring")

for ns in "${NAMESPACES[@]}"; do
  echo "Rolling back namespace: $ns"

  for deploy in $(kubectl get deployments -n $ns -o name); do
    echo "  Rolling back $deploy"
    kubectl rollout undo $deploy -n $ns
  done

  # Wait for rollout to complete
  kubectl wait --for=condition=available --timeout=300s \
    deployments --all -n $ns
done
```

### Rollback Time Estimates

| Component | Estimated Time |
|-----------|---------------|
| Rollback command | <30 seconds |
| Pod termination | 30-60 seconds |
| Pod startup | 1-2 minutes |
| Health checks | 30-60 seconds |
| **Total** | **3-5 minutes** |

---

## Database Rollback

### PostgreSQL Rollback

#### 1. Stop Application Services

```bash
# Prevent new connections
docker compose stop api web

# Or in Kubernetes
kubectl scale deployment/ollamamax-api --replicas=0 -n ollamamax
```

#### 2. Restore from Backup

```bash
# List available backups
ls -lh backups/postgres/

# Stop database temporarily
docker compose stop postgres

# Restore from backup
docker compose up -d postgres
sleep 10

docker compose exec -T postgres psql -U postgres < backups/postgres/backup-20250101.sql

# Or restore specific database
docker compose exec -T postgres psql -U postgres -d ollamamax < backups/postgres/ollamamax-20250101.sql
```

#### 3. Verify Data Integrity

```bash
# Connect to database
docker compose exec postgres psql -U postgres -d ollamamax

# Run integrity checks
SELECT COUNT(*) FROM users;
SELECT MAX(created_at) FROM events;

# Check for data corruption
SELECT * FROM pg_stat_database WHERE datname = 'ollamamax';
```

#### 4. Restart Application Services

```bash
# Docker
docker compose start api web

# Kubernetes
kubectl scale deployment/ollamamax-api --replicas=3 -n ollamamax
```

### Redis Rollback

```bash
# Stop application services
docker compose stop api

# Restore Redis backup
docker compose exec redis redis-cli SHUTDOWN SAVE
docker cp backups/redis/dump-20250101.rdb $(docker compose ps -q redis):/data/dump.rdb
docker compose start redis

# Verify data
docker compose exec redis redis-cli PING
docker compose exec redis redis-cli DBSIZE

# Restart application
docker compose start api
```

### Migration Rollback

```bash
# If using migration tool (e.g., Flyway, Liquibase)

# List migrations
docker compose exec api npm run migrate:status

# Rollback last migration
docker compose exec api npm run migrate:down

# Or rollback to specific version
docker compose exec api npm run migrate:down --to 20250101_120000
```

### Rollback Time Estimates

| Operation | Estimated Time |
|-----------|---------------|
| Stop services | 1-2 minutes |
| Database restore | 5-15 minutes (depends on size) |
| Data verification | 2-3 minutes |
| Service restart | 2-3 minutes |
| **Total** | **10-25 minutes** |

**Note:** Database rollback is the longest operation. Consider:
- Database size impacts restore time
- Point-in-time recovery for large databases
- Read replicas for faster recovery

---

## Automated Rollback

### Health Check-Based Rollback

```yaml
# monitoring/alerting-rules.yaml
groups:
  - name: auto-rollback
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          action: "auto_rollback"

      - alert: ServiceDown
        expr: up{job="ollama-api"} == 0
        for: 2m
        annotations:
          summary: "Service is down"
          action: "auto_rollback"
```

### Automated Rollback Script

```bash
#!/bin/bash
# scripts/auto-rollback.sh

DEPLOYMENT="ollamamax-api"
NAMESPACE="ollamamax"
ERROR_THRESHOLD=5  # Error rate percentage
HEALTH_CHECK_FAILURES=3

# Monitor error rate
ERROR_RATE=$(kubectl exec -n ${NAMESPACE} \
  $(kubectl get pod -n ${NAMESPACE} -l app=${DEPLOYMENT} -o jsonpath='{.items[0].metadata.name}') \
  -- curl -s http://localhost:9090/api/v1/query?query=error_rate | jq -r '.data.result[0].value[1]')

if (( $(echo "$ERROR_RATE > $ERROR_THRESHOLD" | bc -l) )); then
  echo "❌ Error rate ${ERROR_RATE}% exceeds threshold ${ERROR_THRESHOLD}%"
  echo "🔄 Initiating automatic rollback..."

  # Execute rollback
  kubectl rollout undo deployment/${DEPLOYMENT} -n ${NAMESPACE}

  # Wait for rollback
  kubectl rollout status deployment/${DEPLOYMENT} -n ${NAMESPACE} --timeout=300s

  # Verify health
  sleep 10
  NEW_ERROR_RATE=$(kubectl exec -n ${NAMESPACE} \
    $(kubectl get pod -n ${NAMESPACE} -l app=${DEPLOYMENT} -o jsonpath='{.items[0].metadata.name}') \
    -- curl -s http://localhost:9090/api/v1/query?query=error_rate | jq -r '.data.result[0].value[1]')

  if (( $(echo "$NEW_ERROR_RATE < $ERROR_THRESHOLD" | bc -l) )); then
    echo "✅ Rollback successful - error rate now ${NEW_ERROR_RATE}%"
    # Send notification
    curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
      -d "{\"text\":\"✅ Auto-rollback successful for ${DEPLOYMENT}\"}"
  else
    echo "❌ Rollback failed - error rate still ${NEW_ERROR_RATE}%"
    # Escalate to on-call
    curl -X POST https://api.pagerduty.com/incidents \
      -H "Authorization: Token YOUR_TOKEN" \
      -d "{\"incident\":{\"type\":\"incident\",\"title\":\"Rollback failed for ${DEPLOYMENT}\"}}"
  fi
fi
```

---

## Manual Rollback

### Manual Rollback Checklist

- [ ] Incident detected and severity assessed
- [ ] Rollback decision approved (if required)
- [ ] Current version documented
- [ ] Rollback command prepared
- [ ] Communication sent to team
- [ ] Rollback executed
- [ ] Health checks verified
- [ ] Monitoring confirmed stable
- [ ] Post-rollback analysis scheduled

### Communication Template

```
🚨 ROLLBACK INITIATED

Service: [Service Name]
Environment: [Production/Staging]
Reason: [Brief description]
Current Version: [v1.2.5]
Rolling back to: [v1.2.4]
Expected completion: [HH:MM]
Impact: [User impact description]

Status updates will be provided every 5 minutes.

- DevOps Team
```

---

## Rollback Testing

### Regular Rollback Drills

```bash
# Run rollback test monthly
bash scripts/test-rollback.sh

# Simulate failure and rollback
kubectl set image deployment/ollamamax-api \
  api=ollamamax-api:broken -n ollamamax

# Wait for failure detection
sleep 30

# Execute rollback
kubectl rollout undo deployment/ollamamax-api -n ollamamax

# Verify success
kubectl rollout status deployment/ollamamax-api -n ollamamax
```

### Rollback Test Report

```markdown
# Rollback Test Report

**Date:** 2025-01-15
**Environment:** Staging
**Tester:** DevOps Team

## Test Results

| Test Scenario | Expected Time | Actual Time | Status |
|---------------|---------------|-------------|--------|
| Docker rollback | 6-10 min | 8 min | ✅ PASS |
| K8s rollback | 3-5 min | 4 min | ✅ PASS |
| Database rollback | 10-25 min | 15 min | ✅ PASS |
| Health verification | 1-2 min | 1.5 min | ✅ PASS |

## Issues Found
- None

## Recommendations
- Update rollback script documentation
- Add automated notifications
- Improve health check coverage
```

---

## Post-Rollback

### Immediate Actions

1. **Verify System Stability**
   ```bash
   # Monitor for 15 minutes
   watch -n 5 'kubectl get pods --all-namespaces'

   # Check error rates
   curl http://prometheus:9090/api/v1/query?query=error_rate
   ```

2. **Document Incident**
   - Create incident report
   - Document timeline
   - Identify root cause
   - Record rollback actions

3. **Communicate Status**
   ```
   ✅ ROLLBACK COMPLETE

   Service: [Service Name]
   Rolled back to: [v1.2.4]
   Completion time: [HH:MM]
   Current status: Stable

   All systems operating normally.
   Post-mortem scheduled for [Date/Time].
   ```

### Root Cause Analysis

Schedule within 24 hours:
- Timeline reconstruction
- Identify failure cause
- Review monitoring gaps
- Update runbooks
- Implement preventive measures

### Preventive Measures

- Add monitoring for detected issue
- Improve testing coverage
- Update deployment checklist
- Enhance staging environment
- Review rollback procedures

---

## RTO/RPO Targets

### Recovery Time Objective (RTO)

| Service | Target RTO |
|---------|-----------|
| API Services | 5 minutes |
| Web Frontend | 5 minutes |
| Database | 15 minutes |
| Monitoring | 10 minutes |

### Recovery Point Objective (RPO)

| Data Type | Target RPO |
|-----------|-----------|
| User Data | 5 minutes (continuous replication) |
| Configuration | 15 minutes (hourly backups) |
| Logs | 1 minute (real-time streaming) |
| Metrics | 1 minute (real-time collection) |

---

## Emergency Contacts

- **On-Call Engineer:** [PagerDuty]
- **DevOps Lead:** [Contact Info]
- **Engineering Manager:** [Contact Info]
- **Database Administrator:** [Contact Info]

---

**Remember:** Practice rollback procedures regularly. The best time to test your rollback is before you need it.
