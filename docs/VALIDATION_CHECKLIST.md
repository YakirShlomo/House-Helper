# Acceptance & Validation Checklist

This document provides a comprehensive checklist for validating the House Helper application is production-ready.

## Table of Contents

1. [Code Quality Validation](#code-quality-validation)
2. [Build Verification](#build-verification)
3. [Testing Validation](#testing-validation)
4. [Security Validation](#security-validation)
5. [Infrastructure Validation](#infrastructure-validation)
6. [Documentation Validation](#documentation-validation)
7. [Performance Validation](#performance-validation)
8. [Deployment Validation](#deployment-validation)
9. [Final Sign-Off](#final-sign-off)

---

## Code Quality Validation

### Backend (Go) Services

- [ ] **API Service**
  - [ ] Code compiles without errors: `cd services/api && go build cmd/server/main.go`
  - [ ] All tests pass: `go test ./... -v`
  - [ ] Code coverage ≥ 80%: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
  - [ ] No linting errors: `golangci-lint run`
  - [ ] No security vulnerabilities: `gosec ./...`
  - [ ] Go modules tidy: `go mod tidy && git diff go.mod go.sum` (should show no changes)

- [ ] **Notifier Service**
  - [ ] Code compiles: `cd services/notifier && go build cmd/server/main.go`
  - [ ] All tests pass: `go test ./... -v`
  - [ ] Code coverage ≥ 70%
  - [ ] No linting errors: `golangci-lint run`
  - [ ] No security vulnerabilities: `gosec ./...`

- [ ] **Temporal Worker**
  - [ ] Code compiles: `cd services/temporal-worker && go build cmd/worker/main.go`
  - [ ] All tests pass: `go test ./... -v`
  - [ ] Workflows tested in isolation
  - [ ] No linting errors

- [ ] **Temporal API**
  - [ ] Code compiles: `cd services/temporal-api && go build cmd/server/main.go`
  - [ ] All tests pass: `go test ./... -v`
  - [ ] No linting errors

- [ ] **Kafka Consumer**
  - [ ] Code compiles: `cd services/kafka-consumer && go build cmd/consumer/main.go`
  - [ ] All tests pass: `go test ./... -v`
  - [ ] Event handlers tested
  - [ ] No linting errors

### Frontend (Flutter) App

- [ ] **Mobile App**
  - [ ] Code compiles (iOS): `cd mobile && flutter build ios --release`
  - [ ] Code compiles (Android): `flutter build apk --release`
  - [ ] All tests pass: `flutter test`
  - [ ] Widget tests pass: `flutter test test/widgets/`
  - [ ] Code coverage ≥ 70%: `flutter test --coverage`
  - [ ] No analyzer errors: `flutter analyze`
  - [ ] Dart formatting correct: `dart format --set-exit-if-changed .`

---

## Build Verification

### Docker Images

- [ ] **All service images build successfully**
  ```bash
  docker build -t house-helper-api:test -f services/api/Dockerfile .
  docker build -t house-helper-notifier:test -f services/notifier/Dockerfile .
  docker build -t house-helper-temporal-worker:test -f services/temporal-worker/Dockerfile .
  docker build -t house-helper-temporal-api:test -f services/temporal-api/Dockerfile .
  docker build -t house-helper-kafka-consumer:test -f services/kafka-consumer/Dockerfile .
  ```

- [ ] **Images are optimized**
  - [ ] Multi-stage builds used
  - [ ] No unnecessary layers
  - [ ] Image sizes reasonable (< 50MB for Go services)
  - [ ] Check: `docker images | grep house-helper`

- [ ] **Images scan clean**
  ```bash
  trivy image house-helper-api:test
  trivy image house-helper-notifier:test
  # etc. for all images
  ```
  - [ ] No HIGH or CRITICAL vulnerabilities

### Docker Compose

- [ ] **Local environment starts successfully**
  ```bash
  docker compose up -d
  docker compose ps  # All services "Up"
  ```

- [ ] **All services healthy**
  ```bash
  curl http://localhost:8080/health  # API
  curl http://localhost:8081/health  # Notifier
  curl http://localhost:8082/health  # Temporal API
  redis-cli ping                      # Redis
  psql -h localhost -U postgres -d househelper -c "SELECT 1;"  # PostgreSQL
  ```

- [ ] **Database migrations run**
  ```bash
  psql -h localhost -U postgres -d househelper -c "\dt"
  # Should show: users, families, family_members, tasks, chores, points, notifications
  ```

---

## Testing Validation

### Unit Tests

- [ ] **Backend unit tests**
  ```bash
  # Run all unit tests
  make test
  
  # Check coverage
  go test ./... -coverprofile=coverage.out
  go tool cover -func=coverage.out | grep total
  # Should show: total coverage ≥ 80%
  ```

- [ ] **Frontend unit tests**
  ```bash
  cd mobile
  flutter test --coverage
  # Check coverage/lcov.info for ≥ 70% coverage
  ```

### Integration Tests

- [ ] **Backend integration tests**
  ```bash
  # Start test dependencies
  docker compose up -d postgres redis
  
  # Run integration tests
  go test -tags=integration ./services/api/tests/integration -v
  
  # All tests should pass
  ```

- [ ] **Test scenarios covered**
  - [ ] User registration and authentication
  - [ ] Family creation and member management
  - [ ] Task creation, assignment, and completion
  - [ ] Points calculation and leaderboards
  - [ ] Concurrent operations (no race conditions)

### Load Tests

- [ ] **k6 load tests run successfully**
  ```bash
  # Start services
  docker compose up -d
  
  # Run smoke test
  k6 run tests/load/k6-api-test.js --env SCENARIO=smoke
  
  # Run load test
  k6 run tests/load/k6-api-test.js --env SCENARIO=load
  ```

- [ ] **Performance thresholds met**
  - [ ] P95 latency < 500ms
  - [ ] P99 latency < 1000ms
  - [ ] Error rate < 1%
  - [ ] Throughput > 100 req/s

### End-to-End Tests

- [ ] **Complete user journeys tested**
  - [ ] User Registration Flow
    - [ ] Register new user
    - [ ] Verify email (simulated)
    - [ ] Login with credentials
    - [ ] Receive JWT token
  
  - [ ] Family Management Flow
    - [ ] Create new family
    - [ ] Invite member (email)
    - [ ] Accept invitation
    - [ ] View family members
  
  - [ ] Task Management Flow
    - [ ] Create task
    - [ ] Assign to user
    - [ ] User sees assigned task
    - [ ] Mark task complete
    - [ ] Points awarded
    - [ ] Task appears in leaderboard
  
  - [ ] Notification Flow
    - [ ] Task assigned → notification sent
    - [ ] Task completed → notification sent
    - [ ] Notification appears in app

---

## Security Validation

### Code Security

- [ ] **Static Analysis Clean**
  ```bash
  # Go security scanner
  gosec -tests ./...
  
  # No HIGH or MEDIUM issues
  ```

- [ ] **Dependency Vulnerabilities**
  ```bash
  # Go vulnerabilities
  govulncheck ./...
  
  # Snyk scan
  snyk test
  
  # No HIGH or CRITICAL vulnerabilities
  ```

- [ ] **License Compliance**
  ```bash
  go-licenses check ./...
  # No GPL or copyleft licenses (unless acceptable)
  ```

### Container Security

- [ ] **Image vulnerabilities scanned**
  ```bash
  trivy image house-helper-api:test
  grype house-helper-api:test
  
  # No HIGH or CRITICAL vulnerabilities
  ```

- [ ] **SBOM generated**
  ```bash
  syft packages house-helper-api:test -o spdx-json > sbom.json
  # Verify SBOM contains all dependencies
  ```

### Infrastructure Security

- [ ] **Terraform security scan**
  ```bash
  cd infra/terraform
  tfsec .
  
  # No HIGH or CRITICAL issues
  ```

- [ ] **Secrets not in code**
  ```bash
  trufflehog filesystem . --only-verified
  
  # No secrets found
  ```

### Security Documentation

- [ ] **Security policies documented**
  - [ ] [SECURITY.md](./SECURITY.md) exists and complete
  - [ ] Vulnerability reporting process documented
  - [ ] Security contact information provided

- [ ] **Privacy policy documented**
  - [ ] [PRIVACY.md](./PRIVACY.md) exists and complete
  - [ ] GDPR compliance addressed
  - [ ] Data retention policies documented

---

## Infrastructure Validation

### Terraform

- [ ] **Terraform plan clean**
  ```bash
  cd infra/terraform
  terraform init
  terraform plan -var="environment=staging"
  
  # Review plan, ensure no unexpected changes
  ```

- [ ] **Terraform validate passes**
  ```bash
  terraform validate
  # Should show: Success! The configuration is valid.
  ```

### Helm Charts

- [ ] **Helm charts lint clean**
  ```bash
  cd infra/helm
  helm lint house-helper
  
  # No errors
  ```

- [ ] **Helm template renders correctly**
  ```bash
  helm template house-helper ./house-helper \
    --values house-helper/values-staging.yaml \
    --debug
  
  # Review output, ensure valid Kubernetes manifests
  ```

- [ ] **Helm charts install successfully (dry-run)**
  ```bash
  helm install house-helper ./house-helper \
    --namespace house-helper-staging \
    --values house-helper/values-staging.yaml \
    --dry-run --debug
  
  # No errors
  ```

### Kubernetes

- [ ] **All required resources defined**
  - [ ] Deployments for all services
  - [ ] Services for all deployments
  - [ ] ConfigMaps for configuration
  - [ ] Secrets (referenced from External Secrets)
  - [ ] HPA (Horizontal Pod Autoscaler)
  - [ ] PDB (Pod Disruption Budget)
  - [ ] Ingress for external access

---

## Documentation Validation

### Completeness

- [ ] **Core documentation exists**
  - [ ] [README.md](../README.md) - Project overview
  - [ ] [ARCHITECTURE.md](./ARCHITECTURE.md) - System design
  - [ ] [API.md](./API.md) - API documentation
  - [ ] [QUICKSTART.md](./QUICKSTART.md) - Developer setup
  - [ ] [PRD.md](./PRD.md) - Product requirements
  - [ ] [TESTING_STRATEGY.md](./TESTING_STRATEGY.md) - Testing approach
  - [ ] [SLO_SLI.md](./SLO_SLI.md) - Service level objectives
  - [ ] [SECURITY.md](./SECURITY.md) - Security policies
  - [ ] [PRIVACY.md](./PRIVACY.md) - Privacy policies

- [ ] **Operational documentation exists**
  - [ ] [RUNBOOKS.md](./runbooks/RUNBOOKS.md) - Operational procedures
  - [ ] Database schema documented
  - [ ] Deployment procedures documented
  - [ ] Monitoring and alerting documented

- [ ] **Architecture Decision Records**
  - [ ] [ADR-001: Go for Backend](./adr/ADR-001-go-for-backend.md)
  - [ ] [ADR-003: PostgreSQL over NoSQL](./adr/ADR-003-postgresql-over-nosql.md)
  - [ ] [ADR-010: Microservices Architecture](./adr/ADR-010-microservices-architecture.md)
  - [ ] Other key decisions documented

### Quality

- [ ] **Documentation is accurate**
  - [ ] All code examples tested
  - [ ] All commands verified
  - [ ] All URLs valid
  - [ ] No outdated information

- [ ] **Documentation is complete**
  - [ ] All features documented
  - [ ] All APIs documented
  - [ ] All configuration options documented
  - [ ] Troubleshooting sections included

---

## Performance Validation

### API Performance

- [ ] **Response times meet SLOs**
  ```bash
  # Run load test
  k6 run tests/load/k6-api-test.js --env SCENARIO=load
  
  # Verify metrics:
  # - http_req_duration p95 < 500ms ✓
  # - http_req_duration p99 < 1000ms ✓
  # - http_req_failed rate < 0.01 ✓
  ```

- [ ] **Specific endpoint performance**
  - [ ] GET /tasks - P95 < 300ms
  - [ ] POST /tasks - P95 < 400ms
  - [ ] GET /families/{id}/leaderboard - P95 < 500ms
  - [ ] POST /tasks/{id}/complete - P95 < 400ms

### Database Performance

- [ ] **Query performance acceptable**
  ```bash
  # Enable slow query log
  psql -h localhost -U postgres -d househelper -c \
    "ALTER SYSTEM SET log_min_duration_statement = 100;"
  psql -h localhost -U postgres -d househelper -c "SELECT pg_reload_conf();"
  
  # Run load test, check logs for slow queries
  docker compose logs postgres | grep "duration:"
  
  # No queries > 1000ms
  ```

- [ ] **Database indexes optimal**
  ```bash
  # Check for missing indexes
  psql -h localhost -U postgres -d househelper -c \
    "SELECT * FROM pg_stat_user_tables WHERE seq_scan > 0 ORDER BY seq_scan DESC;"
  
  # High seq_scan values may indicate missing indexes
  ```

### Resource Usage

- [ ] **Memory usage acceptable**
  ```bash
  docker stats
  
  # Each Go service should use < 100MB
  # PostgreSQL should use < 500MB
  # Redis should use < 100MB
  ```

- [ ] **CPU usage acceptable**
  - [ ] Services idle at < 5% CPU
  - [ ] Services under load at < 50% CPU (with proper scaling)

---

## Deployment Validation

### Staging Environment

- [ ] **Staging deployment successful**
  ```bash
  # Deploy to staging
  cd infra/helm
  helm upgrade --install house-helper ./house-helper \
    --namespace house-helper-staging \
    --values house-helper/values-staging.yaml \
    --create-namespace
  
  # Verify pods running
  kubectl get pods -n house-helper-staging
  # All pods should be "Running" and "Ready"
  ```

- [ ] **Smoke tests pass in staging**
  ```bash
  # Health checks
  curl https://staging-api.house-helper.com/health
  
  # Basic functionality
  ./scripts/smoke-tests.sh staging
  ```

- [ ] **E2E tests pass in staging**
  ```bash
  # Run E2E test suite against staging
  TEST_ENV=staging ./scripts/e2e-tests.sh
  ```

### CI/CD Pipeline

- [ ] **All GitHub Actions workflows pass**
  - [ ] Backend Tests workflow
  - [ ] Frontend Tests workflow
  - [ ] Security Scan workflow
  - [ ] Build and Push workflow
  - [ ] Deploy to Staging workflow

- [ ] **Pipeline stages complete successfully**
  - [ ] Checkout code ✓
  - [ ] Run tests ✓
  - [ ] Security scans ✓
  - [ ] Build images ✓
  - [ ] Push to ECR ✓
  - [ ] Deploy to staging ✓
  - [ ] Smoke tests ✓

### Rollback Capability

- [ ] **Can rollback deployment**
  ```bash
  # Test rollback
  kubectl rollout undo deployment/house-helper-api -n house-helper-staging
  
  # Verify previous version restored
  kubectl get deployment/house-helper-api -n house-helper-staging -o jsonpath='{.spec.template.spec.containers[0].image}'
  ```

- [ ] **Can rollback database migration**
  ```bash
  # Test migration rollback
  migrate -path services/api/migrations \
    -database "postgres://..." \
    down 1
  
  # Verify schema rolled back
  ```

---

## Final Sign-Off

### Checklist Summary

- [ ] ✅ All code compiles without errors
- [ ] ✅ All tests pass (unit, integration, E2E)
- [ ] ✅ Code coverage meets thresholds (≥80% backend, ≥70% frontend)
- [ ] ✅ No security vulnerabilities (HIGH or CRITICAL)
- [ ] ✅ Performance meets SLOs (P95 <500ms, error rate <1%)
- [ ] ✅ Documentation complete and accurate
- [ ] ✅ Infrastructure code validated (Terraform, Helm)
- [ ] ✅ Successfully deployed to staging
- [ ] ✅ Smoke tests pass in staging
- [ ] ✅ Monitoring and alerting configured
- [ ] ✅ Rollback procedures tested

### Team Sign-Off

- [ ] **Engineering Lead**: ___________________________ Date: __________
  - Code quality verified
  - Tests comprehensive
  - Architecture sound

- [ ] **DevOps/SRE Lead**: ___________________________ Date: __________
  - Infrastructure validated
  - Deployment procedures tested
  - Monitoring configured

- [ ] **Security Lead**: ___________________________ Date: __________
  - Security scans clean
  - Compliance requirements met
  - Secrets properly managed

- [ ] **QA Lead**: ___________________________ Date: __________
  - All tests pass
  - Performance validated
  - E2E scenarios verified

- [ ] **Product Owner**: ___________________________ Date: __________
  - Features complete
  - Acceptance criteria met
  - Ready for production

- [ ] **CTO**: ___________________________ Date: __________
  - Overall system validated
  - Ready for production launch
  - Go/No-Go decision: **GO** / NO-GO

---

## Production Launch Checklist

### Pre-Launch (T-7 days)

- [ ] Final security audit completed
- [ ] Load testing completed with production-like data
- [ ] Disaster recovery plan tested
- [ ] Monitoring dashboards reviewed and configured
- [ ] Alert thresholds tuned
- [ ] Runbooks updated
- [ ] On-call rotation established
- [ ] Communication plan prepared

### Launch Day (T-0)

- [ ] **08:00 AM**: Final team sync
- [ ] **09:00 AM**: Deploy to production
- [ ] **09:15 AM**: Verify all services healthy
- [ ] **09:30 AM**: Run smoke tests
- [ ] **10:00 AM**: Open to beta users (10% traffic)
- [ ] **11:00 AM**: Monitor metrics, check for issues
- [ ] **12:00 PM**: Increase to 50% traffic if healthy
- [ ] **02:00 PM**: Increase to 100% traffic if healthy
- [ ] **03:00 PM**: Monitor for stability
- [ ] **05:00 PM**: Launch retrospective meeting

### Post-Launch (T+1 week)

- [ ] Monitor SLOs daily
- [ ] Track error budget consumption
- [ ] Review incident reports
- [ ] Collect user feedback
- [ ] Performance optimization based on real traffic
- [ ] Post-launch retrospective

---

## Success Criteria

The system is considered **production-ready** when:

✅ All validation checklist items completed  
✅ All team leads have signed off  
✅ CTO approval obtained  
✅ Zero HIGH or CRITICAL security issues  
✅ Performance meets SLOs in staging  
✅ All documentation complete  
✅ Rollback procedures tested  
✅ On-call rotation established  
✅ Monitoring and alerting operational  

---

**Document Owner**: Engineering Team  
**Last Updated**: January 2024  
**Next Review**: Before Production Launch

## License

Copyright © 2024 House Helper. All rights reserved.
