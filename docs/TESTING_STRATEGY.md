# Testing Strategy

## Overview

This document outlines the comprehensive testing strategy for the House Helper application, covering all aspects from unit tests to end-to-end testing, load testing, and observability.

## Testing Pyramid

```
                    /\
                   /  \
                  / E2E \
                 /  Tests \
                /----------\
               / Integration \
              /     Tests     \
             /----------------\
            /   Unit Tests     \
           /                    \
          /______________________\
```

### Distribution

- **Unit Tests**: 70% - Fast, isolated, numerous
- **Integration Tests**: 20% - Medium speed, test component interactions
- **E2E Tests**: 10% - Slow, test complete user journeys

## Unit Testing

### Go Services

**Framework**: `testing` (standard library) + `testify` for assertions

**Coverage Target**: > 80% for all services

**Location**: `services/*/tests/*_test.go`

**Running Tests**:
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestTaskHandler_CreateTask ./services/api/handlers/

# Run with race detection
go test -race ./...

# Verbose output
go test -v ./...
```

**Test Structure**:
```go
// services/api/handlers/task_handler_test.go
func TestTaskHandler_CreateTask_Success(t *testing.T) {
    // Arrange
    mockService := new(MockTaskService)
    handler := handlers.NewTaskHandler(mockService)
    // ... setup

    // Act
    handler.CreateTask(rec, req)

    // Assert
    assert.Equal(t, http.StatusCreated, rec.Code)
    mockService.AssertExpectations(t)
}
```

**Best Practices**:
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test edge cases and error conditions
- Use descriptive test names: `TestFunction_Scenario_ExpectedBehavior`
- Benchmark performance-critical code

### Flutter Mobile App

**Framework**: `flutter_test` package

**Coverage Target**: > 70% for business logic

**Location**: `mobile/test/*_test.dart`

**Running Tests**:
```bash
# Run all tests
flutter test

# Run with coverage
flutter test --coverage

# Generate coverage report
genhtml coverage/lcov.info -o coverage/html

# Run specific test
flutter test test/services/task_service_test.dart

# Run widget tests
flutter test test/widgets/

# Run integration tests
flutter test integration_test/
```

**Test Structure**:
```dart
// mobile/test/services/task_service_test.dart
void main() {
  group('TaskService', () {
    late TaskService service;
    late MockHttpClient mockHttp;

    setUp(() {
      mockHttp = MockHttpClient();
      service = TaskService(mockHttp);
    });

    test('createTask returns Task on success', () async {
      // Arrange
      when(() => mockHttp.post(any(), body: any(named: 'body')))
          .thenAnswer((_) async => Response('{"id":"123"}', 201));

      // Act
      final task = await service.createTask(CreateTaskRequest(
        title: 'Test Task',
        points: 10,
      ));

      // Assert
      expect(task.id, '123');
    });
  });
}
```

**Best Practices**:
- Test widgets with `WidgetTester`
- Mock HTTP clients and platform channels
- Use `pumAndSettle()` for animations
- Test state management (Provider, Riverpod, BLoC)
- Test navigation flows

## Integration Testing

### Go Services with Testcontainers

**Framework**: `testcontainers-go`

**Location**: `services/api/tests/integration/`

**Running Tests**:
```bash
# Run integration tests
go test -tags=integration ./services/api/tests/integration/

# With verbose output
go test -tags=integration -v ./services/api/tests/integration/
```

**Test Structure**:
```go
// services/api/tests/integration/integration_test.go
type IntegrationTestSuite struct {
    suite.Suite
    db        *gorm.DB
    container testcontainers.Container
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Start PostgreSQL container
    // Connect to database
    // Run migrations
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    // Cleanup
}

func (suite *IntegrationTestSuite) SetupTest() {
    // Clean database before each test
}

func (suite *IntegrationTestSuite) TestTaskLifecycle() {
    // Test complete task lifecycle
}
```

**Containers Used**:
- PostgreSQL 16
- Redis 7
- Kafka 3.6

**Best Practices**:
- Use real databases, not mocks
- Test complete workflows
- Verify data persistence
- Test transaction rollbacks
- Test concurrent operations

### Flutter Integration Tests

**Location**: `mobile/integration_test/`

**Running Tests**:
```bash
# Run on emulator
flutter test integration_test/

# Run on device
flutter test integration_test/ -d <device-id>

# Generate coverage
flutter test integration_test/ --coverage
```

**Test Structure**:
```dart
// mobile/integration_test/app_test.dart
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('complete task flow', (WidgetTester tester) async {
    // Launch app
    app.main();
    await tester.pumpAndSettle();

    // Navigate to tasks
    await tester.tap(find.byIcon(Icons.task));
    await tester.pumpAndSettle();

    // Create task
    await tester.tap(find.byIcon(Icons.add));
    await tester.pumpAndSettle();
    // ... fill form and submit

    // Verify task created
    expect(find.text('Test Task'), findsOneWidget);
  });
}
```

## End-to-End Testing

### Test Scenarios

**User Registration & Login**:
1. User registers with email/password
2. User receives verification email
3. User verifies email
4. User logs in

**Family Management**:
1. User creates family
2. User invites members
3. Members accept invitation
4. User assigns roles

**Task Management**:
1. User creates task
2. Task assigned to family member
3. Member receives notification
4. Member completes task
5. Points awarded

**Chore Rotation**:
1. Admin creates recurring chore
2. System rotates assignment
3. Notifications sent
4. Members complete chores

### Tools

- **Manual Testing**: Documented test cases
- **Automated E2E**: Playwright/Selenium (future)
- **API Testing**: Postman collections

### Test Data

**Development Environment**:
```yaml
users:
  - email: admin@test.com
    password: Admin123!
    role: admin
  - email: member@test.com
    password: Member123!
    role: member

families:
  - name: Test Family
    owner: admin@test.com
    members:
      - member@test.com

tasks:
  - title: Weekly Groceries
    points: 50
    assigned_to: member@test.com
  - title: Clean Kitchen
    points: 30
    assigned_to: admin@test.com
```

## Load Testing

### k6 Load Testing

**Framework**: k6

**Location**: `tests/load/k6-api-test.js`

**Running Tests**:
```bash
# Local execution
k6 run tests/load/k6-api-test.js

# With environment variables
k6 run \
  -e BASE_URL=https://api.house-helper.com \
  -e API_KEY=your-key \
  tests/load/k6-api-test.js

# Cloud execution
k6 cloud tests/load/k6-api-test.js

# With output to InfluxDB
k6 run \
  --out influxdb=http://localhost:8086/k6 \
  tests/load/k6-api-test.js
```

**Test Scenarios**:

1. **Smoke Test**: 1 VU for 1 minute
2. **Load Test**: Ramp 0→20 users over 15 minutes
3. **Stress Test**: Ramp 0→100 users, find breaking point
4. **Spike Test**: Sudden spike from 10→200 users

**Performance Thresholds**:
```javascript
thresholds: {
  'http_req_duration': ['p(95)<500', 'p(99)<1000'],
  'http_req_failed': ['rate<0.01'],
  'task_creation_duration': ['p(95)<600'],
}
```

### Load Test Results Analysis

**Metrics to Monitor**:
- Request rate (requests/second)
- Response time (P50, P95, P99)
- Error rate
- Throughput (bytes/second)
- Virtual users

**Expected Performance**:
- API P95 latency: < 500ms
- API P99 latency: < 1000ms
- Error rate: < 1%
- Throughput: > 1000 req/s

## Performance Testing

### Database Performance

**Query Performance**:
```bash
# Enable slow query logging
ALTER SYSTEM SET log_min_duration_statement = '100';  -- 100ms
SELECT pg_reload_conf();

# Analyze slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

# Explain query plans
EXPLAIN ANALYZE
SELECT * FROM tasks WHERE family_id = 'xyz' AND status = 'pending';
```

**Connection Pool Testing**:
```go
// Test connection pool exhaustion
func TestConnectionPoolExhaustion(t *testing.T) {
    db := setupDB()
    maxConns := 100
    
    var wg sync.WaitGroup
    for i := 0; i < maxConns + 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            var result int
            db.Raw("SELECT 1").Scan(&result)
        }()
    }
    wg.Wait()
}
```

### API Performance

**Benchmarking**:
```go
func BenchmarkTaskCreation(b *testing.B) {
    handler := setupHandler()
    req := setupRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rec := httptest.NewRecorder()
        handler.ServeHTTP(rec, req)
    }
}
```

**HTTP Load Testing**:
```bash
# Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/v1/tasks

# wrk
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/tasks

# vegeta
echo "GET http://localhost:8080/api/v1/tasks" | \
  vegeta attack -duration=30s -rate=50 | \
  vegeta report
```

## Security Testing

### Static Analysis

**Go**:
```bash
# gosec - security scanner
gosec ./...

# staticcheck
staticcheck ./...

# go vet
go vet ./...
```

**Dependency Scanning**:
```bash
# govulncheck
govulncheck ./...

# Snyk
snyk test

# OWASP Dependency Check
dependency-check --project house-helper --scan .
```

### Dynamic Analysis

**DAST Tools**:
- OWASP ZAP
- Burp Suite
- Nikto

**Penetration Testing**:
- SQL Injection
- XSS
- CSRF
- Authentication bypass
- Authorization bypass
- Rate limiting
- Input validation

## Observability & Monitoring

### Metrics Collection

**Prometheus Exporters**:
- Node exporter (infrastructure metrics)
- Postgres exporter (database metrics)
- Redis exporter (cache metrics)
- Kafka exporter (message queue metrics)
- Custom application metrics

**Custom Metrics**:
```go
// services/api/metrics/metrics.go
var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)
```

### Logging

**Structured Logging**:
```go
log.WithFields(log.Fields{
    "user_id": userID,
    "action":  "create_task",
    "task_id": task.ID,
}).Info("Task created successfully")
```

**Log Levels**:
- DEBUG: Detailed information for debugging
- INFO: General informational messages
- WARN: Warning messages
- ERROR: Error messages
- FATAL: Critical errors that cause shutdown

**Log Aggregation**:
- Loki for log aggregation
- CloudWatch Logs for AWS
- ELK Stack (Elasticsearch, Logstash, Kibana)

### Distributed Tracing

**OpenTelemetry**:
```go
// Initialize tracer
tracer := otel.Tracer("house-helper-api")

// Create span
ctx, span := tracer.Start(ctx, "CreateTask")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.String("task.id", taskID),
)

// Record error
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

**Trace Backends**:
- Jaeger
- Tempo
- AWS X-Ray

## Test Data Management

### Test Fixtures

**Go**:
```go
// testdata/fixtures.go
var TestUsers = []models.User{
    {
        Email: "test1@example.com",
        Name:  "Test User 1",
    },
    {
        Email: "test2@example.com",
        Name:  "Test User 2",
    },
}
```

**SQL**:
```sql
-- testdata/fixtures.sql
INSERT INTO users (email, name, password_hash)
VALUES 
  ('test1@example.com', 'Test User 1', '$2a$10$...'),
  ('test2@example.com', 'Test User 2', '$2a$10$...');
```

### Database Seeding

```bash
# Development environment
go run cmd/seed/main.go

# Test environment
go run cmd/seed/main.go --env=test

# Custom fixtures
go run cmd/seed/main.go --fixtures=testdata/e2e-fixtures.sql
```

## Continuous Testing

### CI/CD Integration

**GitHub Actions** (already configured):
- Run unit tests on every push
- Run integration tests on PR
- Run E2E tests on staging
- Run load tests on schedule
- Security scanning on every build

**Test Reports**:
- Coverage reports to Codecov
- Test results to GitHub Actions
- Performance reports to Grafana

### Test Environments

**Development**:
- Local Docker Compose
- Shared development cluster
- Synthetic test data

**Staging**:
- Production-like environment
- Anonymized production data
- E2E tests run here

**Production**:
- Canary deployments
- Synthetic monitoring
- Real user monitoring (RUM)

## Quality Gates

### Pre-Merge Checks

- [ ] All unit tests pass
- [ ] Code coverage > 80%
- [ ] Integration tests pass
- [ ] No security vulnerabilities (critical/high)
- [ ] Code review approved
- [ ] Linting passes
- [ ] Documentation updated

### Pre-Release Checks

- [ ] All tests pass (unit, integration, E2E)
- [ ] Load tests meet SLOs
- [ ] Security scan clean
- [ ] Performance benchmarks acceptable
- [ ] Database migrations tested
- [ ] Rollback plan documented
- [ ] Monitoring alerts configured

## Testing Best Practices

### General

1. **Write Tests First** (TDD when appropriate)
2. **Keep Tests Fast**: Unit tests < 1s, Integration tests < 10s
3. **Independent Tests**: No dependencies between tests
4. **Deterministic**: Tests should always produce same result
5. **Readable**: Clear test names and structure
6. **Maintainable**: Refactor tests with production code

### Go-Specific

1. Use table-driven tests for multiple scenarios
2. Use `t.Parallel()` for parallel test execution
3. Use `t.Helper()` for test helper functions
4. Use testify for better assertions
5. Mock external dependencies
6. Use build tags for integration tests

### Flutter-Specific

1. Test widgets in isolation
2. Use `WidgetTester` for widget tests
3. Mock platform channels
4. Test state management separately
5. Use golden tests for UI regression
6. Test accessibility features

## Metrics & Reporting

### Test Metrics

- **Test Coverage**: > 80% for services, > 70% for mobile
- **Test Success Rate**: > 99%
- **Test Execution Time**: < 10 minutes for full suite
- **Flaky Test Rate**: < 1%

### Dashboards

**Test Dashboard**:
- Test success rate over time
- Coverage trends
- Test execution time
- Flaky test identification
- Failed test analysis

**Performance Dashboard**:
- API response times
- Database query times
- Memory usage
- CPU usage
- Error rates

## Resources

- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Flutter Testing](https://docs.flutter.dev/testing)
- [k6 Documentation](https://k6.io/docs/)
- [Testcontainers](https://testcontainers.com/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)

## License

Copyright © 2024 House Helper. All rights reserved.
