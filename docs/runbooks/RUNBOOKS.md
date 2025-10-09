# Operational Runbooks

This directory contains step-by-step guides for common operational tasks, incident response, and troubleshooting procedures.

## Table of Contents

1. [Deployments](#deployments)
2. [Rollbacks](#rollbacks)
3. [Incident Response](#incident-response)
4. [Database Operations](#database-operations)
5. [Monitoring & Alerts](#monitoring--alerts)
6. [Scaling](#scaling)
7. [Security Incidents](#security-incidents)
8. [Disaster Recovery](#disaster-recovery)

---

## Deployments

### Standard Deployment

**Prerequisites**:
- [ ] All tests passing in CI/CD
- [ ] Code review approved
- [ ] Changelog updated
- [ ] Database migrations prepared
- [ ] Rollback plan documented

**Procedure**:

```bash
# 1. Verify current state
kubectl get pods -n house-helper-prod

# 2. Check current version
kubectl get deployment house-helper-api -n house-helper-prod -o jsonpath='{.spec.template.spec.containers[0].image}'

# 3. Deploy to staging first
helm upgrade house-helper ./infra/helm/house-helper \
  --namespace house-helper-staging \
  --values infra/helm/house-helper/values-staging.yaml \
  --set api.image.tag=v1.2.3

# 4. Verify staging deployment
kubectl rollout status deployment/house-helper-api -n house-helper-staging
kubectl logs -f deployment/house-helper-api -n house-helper-staging

# 5. Run smoke tests on staging
curl https://staging-api.house-helper.com/health
./scripts/smoke-tests.sh staging

# 6. Deploy to production
helm upgrade house-helper ./infra/helm/house-helper \
  --namespace house-helper-prod \
  --values infra/helm/house-helper/values-prod.yaml \
  --set api.image.tag=v1.2.3

# 7. Monitor deployment
kubectl rollout status deployment/house-helper-api -n house-helper-prod
watch kubectl get pods -n house-helper-prod

# 8. Verify production deployment
curl https://api.house-helper.com/health
./scripts/smoke-tests.sh production

# 9. Monitor metrics
# Check Grafana dashboards for anomalies
# Monitor error rates, latency, CPU/memory usage
```

**Post-Deployment**:
- [ ] Monitor dashboards for 30 minutes
- [ ] Check error rates
- [ ] Verify no increase in response times
- [ ] Check user reports
- [ ] Update deployment log

**Rollback Criteria**:
- Error rate > 1%
- P95 latency > 1000ms
- Critical bugs reported
- Failed health checks

### Canary Deployment

**Use when**: Rolling out high-risk changes

```bash
# 1. Deploy canary version (10% traffic)
helm upgrade house-helper ./infra/helm/house-helper \
  --namespace house-helper-prod \
  --values infra/helm/house-helper/values-prod.yaml \
  --set canary.enabled=true \
  --set canary.weight=10 \
  --set canary.image.tag=v1.2.3-canary

# 2. Monitor canary metrics for 1 hour
# Compare canary vs stable metrics in Grafana

# 3. Gradually increase traffic
# 10% → 25% → 50% → 100%
helm upgrade house-helper ./infra/helm/house-helper \
  --namespace house-helper-prod \
  --reuse-values \
  --set canary.weight=25

# 4. Promote to stable if successful
helm upgrade house-helper ./infra/helm/house-helper \
  --namespace house-helper-prod \
  --values infra/helm/house-helper/values-prod.yaml \
  --set api.image.tag=v1.2.3
```

### Database Migration Deployment

**Prerequisites**:
- [ ] Migration tested in staging
- [ ] Database backup created
- [ ] Rollback SQL prepared
- [ ] Downtime communicated (if required)

**Procedure**:

```bash
# 1. Create backup
kubectl exec -n house-helper-prod postgresql-0 -- \
  pg_dump -U postgres househelper > backup-$(date +%Y%m%d-%H%M%S).sql

# 2. Upload backup to S3
aws s3 cp backup-*.sql s3://house-helper-backups/manual/

# 3. Test migration in staging
migrate -path services/api/migrations \
  -database "postgres://user:pass@staging-db:5432/househelper?sslmode=require" \
  up

# 4. Run migration in production
migrate -path services/api/migrations \
  -database "postgres://user:pass@prod-db:5432/househelper?sslmode=require" \
  up

# 5. Verify migration
psql -h prod-db -U postgres -d househelper -c "\dt"
psql -h prod-db -U postgres -d househelper -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# 6. Deploy application with migration support
# Follow standard deployment procedure
```

---

## Rollbacks

### Application Rollback

**When to rollback**:
- Error rate exceeds 1%
- Critical functionality broken
- Performance degradation > 50%
- Security vulnerability discovered

**Procedure**:

```bash
# 1. Identify previous version
kubectl rollout history deployment/house-helper-api -n house-helper-prod

# 2. Rollback to previous version
kubectl rollout undo deployment/house-helper-api -n house-helper-prod

# Alternative: Rollback to specific revision
kubectl rollout undo deployment/house-helper-api -n house-helper-prod --to-revision=5

# 3. Monitor rollback
kubectl rollout status deployment/house-helper-api -n house-helper-prod

# 4. Verify functionality
curl https://api.house-helper.com/health
./scripts/smoke-tests.sh production

# 5. Check metrics
# Verify error rate decreased
# Verify latency returned to normal

# 6. Incident follow-up
# Document reason for rollback
# Create bug ticket
# Schedule postmortem
```

### Database Migration Rollback

**Procedure**:

```bash
# 1. Check current migration version
migrate -path services/api/migrations \
  -database "postgres://..." \
  version

# 2. Rollback one migration
migrate -path services/api/migrations \
  -database "postgres://..." \
  down 1

# 3. Or restore from backup (if migration irreversible)
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres househelper < backup-20240115-120000.sql

# 4. Verify database state
psql -h prod-db -U postgres -d househelper -c "SELECT * FROM schema_migrations;"

# 5. Rollback application
# Follow application rollback procedure
```

---

## Incident Response

### Incident Response Workflow

```
1. Detect → 2. Triage → 3. Investigate → 4. Mitigate → 5. Resolve → 6. Postmortem
```

### Severity Levels

| Severity | Description | Response Time | Examples |
|----------|-------------|---------------|----------|
| P0 (Critical) | Complete outage | Immediate | API down, database unavailable |
| P1 (High) | Major functionality impacted | < 30 min | Payment processing broken |
| P2 (Medium) | Minor functionality impacted | < 4 hours | Feature not working correctly |
| P3 (Low) | Cosmetic or minor issues | < 24 hours | UI glitch |

### P0 Incident Response

**Example**: API Service Down

```bash
# 1. ASSESS THE SITUATION
# Check service status
kubectl get pods -n house-helper-prod
kubectl get events -n house-helper-prod --sort-by='.lastTimestamp'

# Check recent deployments
kubectl rollout history deployment/house-helper-api -n house-helper-prod

# Check logs
kubectl logs -f deployment/house-helper-api -n house-helper-prod --tail=100

# Check metrics
# Open Grafana → API Performance Dashboard

# 2. COMMUNICATE
# Post to incident channel: #incidents
# Message: "🚨 P0 INCIDENT: API Service Down. Investigating..."
# Update status page: https://status.house-helper.com

# 3. IMMEDIATE MITIGATION
# Option A: Rollback if recent deployment
kubectl rollout undo deployment/house-helper-api -n house-helper-prod

# Option B: Scale up if resource issue
kubectl scale deployment/house-helper-api -n house-helper-prod --replicas=10

# Option C: Restart pods if hanging
kubectl rollout restart deployment/house-helper-api -n house-helper-prod

# 4. VERIFY RECOVERY
curl https://api.house-helper.com/health
watch kubectl get pods -n house-helper-prod

# 5. COMMUNICATE RESOLUTION
# Post to incident channel: "✅ RESOLVED: API Service restored. Root cause: [brief explanation]. Full postmortem to follow."
# Update status page: Incident resolved

# 6. POST-INCIDENT
# Schedule postmortem meeting within 24 hours
# Document timeline, actions, root cause
# Create action items for prevention
```

### Common Incidents

#### High Error Rate

```bash
# 1. Check error logs
kubectl logs deployment/house-helper-api -n house-helper-prod | grep ERROR

# 2. Check error patterns
# Use Grafana or Loki to identify error types

# 3. Check recent changes
git log --since="2 hours ago" --oneline

# 4. Check external dependencies
# Database, Redis, Kafka status

# 5. Mitigate
# Rollback if caused by recent deployment
# Fix and redeploy if quick fix available
# Scale up if capacity issue
```

#### Database Connection Exhaustion

```bash
# 1. Check current connections
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# 2. Identify connection sources
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT application_name, count(*) FROM pg_stat_activity GROUP BY application_name;"

# 3. Check for connection leaks
# Review application logs for unclosed connections

# 4. Immediate mitigation
# Restart application pods to reset connections
kubectl rollout restart deployment/house-helper-api -n house-helper-prod

# 5. Long-term fix
# Adjust connection pool settings
# Fix connection leaks in code
```

#### High Memory Usage / OOM Kills

```bash
# 1. Check memory usage
kubectl top pods -n house-helper-prod

# 2. Check OOM kills
kubectl get events -n house-helper-prod | grep OOMKilled

# 3. Check memory trends in Grafana

# 4. Immediate mitigation
# Increase memory limits temporarily
kubectl set resources deployment/house-helper-api -n house-helper-prod \
  --limits=memory=4Gi

# 5. Investigate memory leak
# Analyze heap dumps
# Review recent code changes

# 6. Scale horizontally if vertical scaling insufficient
kubectl scale deployment/house-helper-api -n house-helper-prod --replicas=10
```

---

## Database Operations

### Backup Database

```bash
# Manual backup
kubectl exec -n house-helper-prod postgresql-0 -- \
  pg_dump -U postgres -F c househelper > backup-$(date +%Y%m%d-%H%M%S).dump

# Upload to S3
aws s3 cp backup-*.dump s3://house-helper-backups/manual/

# Verify backup
pg_restore --list backup-*.dump | head -20
```

### Restore Database

```bash
# 1. Stop application (to prevent writes)
kubectl scale deployment/house-helper-api -n house-helper-prod --replicas=0

# 2. Download backup
aws s3 cp s3://house-helper-backups/manual/backup-20240115-120000.dump .

# 3. Restore database
kubectl exec -i -n house-helper-prod postgresql-0 -- \
  pg_restore -U postgres -d househelper --clean < backup-20240115-120000.dump

# 4. Verify restore
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "SELECT count(*) FROM users;"

# 5. Restart application
kubectl scale deployment/house-helper-api -n house-helper-prod --replicas=3
```

### Run SQL Query

```bash
# Read-only query
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "SELECT count(*) FROM tasks WHERE status = 'pending';"

# Update query (use with caution)
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "UPDATE users SET email_verified = true WHERE email = 'user@example.com';"
```

### Monitor Database Performance

```bash
# Active queries
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC;"

# Slow queries
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT query, mean_exec_time, calls FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Database size
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT pg_size_pretty(pg_database_size('househelper'));"

# Table sizes
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

---

## Monitoring & Alerts

### Check System Health

```bash
# All pods status
kubectl get pods -n house-helper-prod

# Service endpoints
kubectl get endpoints -n house-helper-prod

# Resource usage
kubectl top nodes
kubectl top pods -n house-helper-prod

# Recent events
kubectl get events -n house-helper-prod --sort-by='.lastTimestamp'

# Check ingress
kubectl get ingress -n house-helper-prod
```

### Respond to Alerts

**High Error Rate Alert**:
1. Check Grafana dashboard for error patterns
2. Review application logs
3. Check recent deployments
4. Roll back if needed

**High Latency Alert**:
1. Check Grafana for latency breakdown
2. Identify slow endpoints
3. Check database query performance
4. Check external service response times
5. Scale up if capacity issue

**Pod Restart Loop Alert**:
1. Check pod logs: `kubectl logs <pod-name> -n house-helper-prod --previous`
2. Check pod events: `kubectl describe pod <pod-name> -n house-helper-prod`
3. Common causes: OOM kill, crash loop, liveness probe failure
4. Fix underlying issue and redeploy

---

## Scaling

### Manual Scaling

```bash
# Scale deployment
kubectl scale deployment/house-helper-api -n house-helper-prod --replicas=10

# Verify scaling
kubectl get pods -n house-helper-prod
kubectl top pods -n house-helper-prod
```

### Check HPA Status

```bash
# View HPA
kubectl get hpa -n house-helper-prod

# Describe HPA
kubectl describe hpa house-helper-api -n house-helper-prod

# Check metrics
kubectl top pods -n house-helper-prod
```

### Scale Database Connections

```bash
# Check current max connections
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SHOW max_connections;"

# Update max connections
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "ALTER SYSTEM SET max_connections = 200;"

# Reload configuration
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -c "SELECT pg_reload_conf();"
```

---

## Security Incidents

### Suspected Account Compromise

```bash
# 1. Identify compromised user
USER_ID="uuid-here"

# 2. Immediately revoke all sessions
kubectl exec -n house-helper-prod redis-0 -- \
  redis-cli DEL "session:$USER_ID:*"

# 3. Disable account
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "UPDATE users SET disabled = true WHERE id = '$USER_ID';"

# 4. Force password reset
# Send password reset email

# 5. Review audit logs
kubectl exec -n house-helper-prod postgresql-0 -- \
  psql -U postgres -d househelper -c "SELECT * FROM audit_log WHERE user_id = '$USER_ID' ORDER BY created_at DESC LIMIT 100;"

# 6. Notify user
# Send security notification email
```

### Suspected Data Breach

1. **Immediate Actions**:
   - Activate incident response team
   - Preserve logs and evidence
   - Contain the breach

2. **Investigation**:
   - Review access logs
   - Identify compromised data
   - Determine breach scope

3. **Notification**:
   - Notify affected users (per GDPR requirements)
   - Notify authorities if required
   - Public disclosure if necessary

4. **Remediation**:
   - Patch vulnerabilities
   - Rotate credentials
   - Enhance security measures

---

## Disaster Recovery

### Complete System Failure

**Recovery Steps**:

```bash
# 1. Provision new infrastructure
cd infra/terraform
terraform apply -var="environment=dr"

# 2. Restore database
aws s3 cp s3://house-helper-backups/daily/latest.dump .
# Restore to new database instance

# 3. Deploy application
cd infra/helm
helm install house-helper ./house-helper \
  --namespace house-helper-prod \
  --values values-prod.yaml

# 4. Update DNS
# Point house-helper.com to new infrastructure

# 5. Verify functionality
./scripts/smoke-tests.sh production

# 6. Monitor closely
# Watch dashboards for anomalies
```

### Regional Failure (AWS)

```bash
# 1. Failover to backup region
# Update Route53 to point to us-west-2

# 2. Verify backup region health
kubectl get pods -n house-helper-prod --context=us-west-2

# 3. Monitor traffic shift
# Watch CloudWatch metrics

# 4. Communicate to users
# Post status update
```

---

## Best Practices

1. **Always check before changing**: Verify current state before making changes
2. **Document everything**: Keep detailed notes of actions taken
3. **Communicate proactively**: Update stakeholders frequently
4. **Test in staging first**: Never test in production
5. **Have rollback plan**: Always know how to undo changes
6. **Monitor after changes**: Watch metrics after deployments
7. **Learn from incidents**: Conduct blameless postmortems

## Emergency Contacts

- **On-Call Engineer**: Check PagerDuty rotation
- **Engineering Lead**: lead@house-helper.com
- **CTO**: cto@house-helper.com
- **AWS Support**: 1-800-xxx-xxxx (Enterprise Support)

## Additional Resources

- [Kubernetes Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [AWS CLI Reference](https://docs.aws.amazon.com/cli/latest/reference/)
- [Prometheus Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)

---

**Last Updated**: January 2024  
**Maintained by**: SRE Team

## License

Copyright © 2024 House Helper. All rights reserved.
