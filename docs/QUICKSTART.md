# Developer Quickstart Guide

Get the House Helper application up and running on your local machine in **15 minutes**.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Detailed Setup](#detailed-setup)
4. [Running the Application](#running-the-application)
5. [Running Tests](#running-tests)
6. [Troubleshooting](#troubleshooting)
7. [Next Steps](#next-steps)

---

## Prerequisites

### Required Software

Install the following tools before proceeding:

| Tool | Version | Installation |
|------|---------|--------------|
| **Git** | Latest | [git-scm.com](https://git-scm.com/) |
| **Docker** | 24.0+ | [docker.com](https://www.docker.com/get-started) |
| **Docker Compose** | 2.20+ | Included with Docker Desktop |
| **Go** | 1.21+ | [go.dev/dl](https://go.dev/dl/) |
| **Flutter** | 3.16+ | [flutter.dev/docs/get-started/install](https://flutter.dev/docs/get-started/install) |
| **Make** | Latest | macOS: Xcode Command Line Tools, Windows: [gnuwin32.sourceforge.net/packages/make.htm](http://gnuwin32.sourceforge.net/packages/make.htm) |
| **kubectl** | 1.28+ | [kubernetes.io/docs/tasks/tools/](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.12+ | [helm.sh/docs/intro/install/](https://helm.sh/docs/intro/install/) |

### Verify Installation

```bash
# Check versions
git --version
docker --version
docker compose version
go version
flutter --version
make --version
kubectl version --client
helm version
```

### System Requirements

- **OS**: macOS, Linux, or Windows with WSL2
- **RAM**: 8GB minimum, 16GB recommended
- **Disk Space**: 10GB free space
- **Ports**: 8080, 8081, 8082, 5432, 6379, 9092, 7233 (ensure these are available)

---

## Quick Start

**For the impatient** - Get everything running in 5 commands:

```bash
# 1. Clone repository
git clone https://github.com/your-org/house-helper.git
cd house-helper

# 2. Start infrastructure
docker compose up -d

# 3. Run database migrations
make migrate-up

# 4. Start backend services
make run-all

# 5. Start mobile app (in new terminal)
cd mobile && flutter run
```

✅ You should now have:
- Backend services running on `http://localhost:8080`
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- Kafka on `localhost:9092`
- Temporal on `localhost:7233`
- Mobile app running on emulator/simulator

---

## Detailed Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/house-helper.git
cd house-helper
```

### 2. Configure Environment

Copy environment template and configure:

```bash
# Copy environment template
cp .env.example .env

# Edit .env file with your configuration
# For local development, defaults should work fine
```

**Example `.env`**:

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/househelper?sslmode=disable
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=househelper
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres

# Redis
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka
KAFKA_BROKERS=localhost:9092

# Temporal
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=3600

# API
API_PORT=8080
API_HOST=localhost

# Notification Service
NOTIFIER_PORT=8081
FIREBASE_CREDENTIALS_FILE=./config/firebase-credentials.json

# Temporal Service
TEMPORAL_API_PORT=8082

# Environment
ENVIRONMENT=development
LOG_LEVEL=debug
```

### 3. Start Infrastructure

Start all required infrastructure services using Docker Compose:

```bash
# Start all services in background
docker compose up -d

# Check services are running
docker compose ps

# Expected output:
# NAME                   STATUS   PORTS
# postgres               Up       0.0.0.0:5432->5432/tcp
# redis                  Up       0.0.0.0:6379->6379/tcp
# kafka                  Up       0.0.0.0:9092->9092/tcp
# temporal               Up       0.0.0.0:7233->7233/tcp
# temporal-ui            Up       0.0.0.0:8233->8233/tcp

# View logs
docker compose logs -f
```

**Verify each service**:

```bash
# PostgreSQL
psql -h localhost -U postgres -d househelper -c "SELECT 1;"

# Redis
redis-cli ping
# Should return: PONG

# Kafka
docker exec -it kafka kafka-topics --list --bootstrap-server localhost:9092

# Temporal UI
# Open browser: http://localhost:8233
```

### 4. Setup Database

Run database migrations to create tables:

```bash
# Install migrate tool (if not installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
make migrate-up

# Or manually:
migrate -path services/api/migrations \
  -database "postgres://postgres:postgres@localhost:5432/househelper?sslmode=disable" \
  up

# Verify tables created
psql -h localhost -U postgres -d househelper -c "\dt"
# Should show: users, families, family_members, tasks, chores, points, notifications, schema_migrations
```

**Seed test data** (optional):

```bash
# Load sample data
psql -h localhost -U postgres -d househelper < scripts/seed-data.sql

# Or use seed script
go run scripts/seed/main.go
```

### 5. Install Go Dependencies

```bash
# Install dependencies for all services
cd services/api && go mod download && cd ../..
cd services/notifier && go mod download && cd ../..
cd services/temporal-worker && go mod download && cd ../..
cd services/temporal-api && go mod download && cd ../..
cd services/kafka-consumer && go mod download && cd ../..

# Or use Makefile
make deps
```

### 6. Setup Flutter

```bash
cd mobile

# Get Flutter dependencies
flutter pub get

# Run code generation (for JSON serialization, etc.)
flutter pub run build_runner build --delete-conflicting-outputs

# Verify setup
flutter doctor
# Should show all checkmarks

# Configure API endpoint
# Edit mobile/lib/config/environment.dart
# Set apiBaseUrl to 'http://localhost:8080/api/v1' or your local IP for physical device
```

---

## Running the Application

### Backend Services

You can run all services together or individually.

#### Option 1: Run All Services (Recommended)

```bash
# Using Makefile
make run-all

# Services will start on:
# - API Service:          http://localhost:8080
# - Notifier Service:     http://localhost:8081
# - Temporal API:         http://localhost:8082
# - Temporal Worker:      (background)
# - Kafka Consumer:       (background)
```

#### Option 2: Run Individual Services

Open separate terminal windows for each service:

**Terminal 1 - API Service**:
```bash
cd services/api
go run cmd/server/main.go
```

**Terminal 2 - Notifier Service**:
```bash
cd services/notifier
go run cmd/server/main.go
```

**Terminal 3 - Temporal Worker**:
```bash
cd services/temporal-worker
go run cmd/worker/main.go
```

**Terminal 4 - Temporal API**:
```bash
cd services/temporal-api
go run cmd/server/main.go
```

**Terminal 5 - Kafka Consumer**:
```bash
cd services/kafka-consumer
go run cmd/consumer/main.go
```

### Mobile App

#### Run on iOS Simulator

```bash
cd mobile

# List available simulators
flutter emulators

# Launch simulator
flutter emulators --launch <simulator_id>

# Or use Xcode to launch simulator

# Run app
flutter run

# Or specify device
flutter run -d iPhone
```

#### Run on Android Emulator

```bash
cd mobile

# List available emulators
flutter emulators

# Launch emulator
flutter emulators --launch <emulator_id>

# Or use Android Studio to launch emulator

# Run app
flutter run

# Or specify device
flutter run -d emulator-5554
```

#### Run on Physical Device

```bash
cd mobile

# iOS: Connect iPhone with cable, trust computer, enable Developer Mode
# Android: Enable USB Debugging in Developer Options

# Check connected devices
flutter devices

# Run app
flutter run -d <device_id>
```

**Note**: If using physical device, update API base URL to your computer's IP address:
```dart
// mobile/lib/config/environment.dart
static const String apiBaseUrl = 'http://192.168.1.100:8080/api/v1'; // Use your IP
```

### Verify Everything is Running

**Check Backend Services**:

```bash
# API Health Check
curl http://localhost:8080/health
# Should return: {"status":"ok"}

# Create test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Check Mobile App**:
- Open app on emulator/device
- Should see login/register screen
- Register new account
- Create a family
- Add a task

---

## Running Tests

### Backend Tests

```bash
# Run all tests
make test

# Or manually:

# Run unit tests
go test ./... -v -cover

# Run unit tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run integration tests
go test -tags=integration ./services/api/tests/integration -v

# Run specific test
go test ./services/api/handlers -run TestTaskHandler_CreateTask -v

# Run benchmarks
go test ./services/api/handlers -bench=. -benchmem
```

### Frontend Tests

```bash
cd mobile

# Run all tests
flutter test

# Run with coverage
flutter test --coverage

# View coverage report
genhtml coverage/lcov.info -o coverage/html
open coverage/html/index.html

# Run specific test
flutter test test/models/user_test.dart

# Run integration tests
flutter test integration_test/
```

### Load Tests

```bash
# Install k6
# macOS: brew install k6
# Windows: choco install k6
# Linux: See https://k6.io/docs/getting-started/installation/

# Run load test
k6 run tests/load/k6-api-test.js

# Run specific scenario
k6 run tests/load/k6-api-test.js --env SCENARIO=smoke

# Run with higher load
k6 run tests/load/k6-api-test.js --env SCENARIO=stress
```

---

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
```bash
# Find process using port
# macOS/Linux
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Or change port in .env file
API_PORT=8090
```

#### 2. Database Connection Failed

**Error**: `connection refused` or `could not connect to database`

**Solution**:
```bash
# Check if PostgreSQL is running
docker compose ps postgres

# Restart PostgreSQL
docker compose restart postgres

# Check logs
docker compose logs postgres

# Verify connection
psql -h localhost -U postgres -d househelper
```

#### 3. Migration Failed

**Error**: `Dirty database version` or migration error

**Solution**:
```bash
# Check current version
migrate -path services/api/migrations \
  -database "postgres://postgres:postgres@localhost:5432/househelper?sslmode=disable" \
  version

# Force version (use with caution)
migrate -path services/api/migrations \
  -database "postgres://postgres:postgres@localhost:5432/househelper?sslmode=disable" \
  force <version>

# Or reset database
docker compose down -v
docker compose up -d
make migrate-up
```

#### 4. Flutter Build Failed

**Error**: Build errors or dependency conflicts

**Solution**:
```bash
cd mobile

# Clean build
flutter clean

# Get dependencies
flutter pub get

# Rebuild
flutter run

# If still failing, update Flutter
flutter upgrade
```

#### 5. Kafka Connection Issues

**Error**: `kafka: client has run out of available brokers`

**Solution**:
```bash
# Check Kafka is running
docker compose ps kafka

# Restart Kafka
docker compose restart kafka

# Check logs
docker compose logs kafka

# Test connection
docker exec -it kafka kafka-broker-api-versions --bootstrap-server localhost:9092
```

#### 6. Temporal Not Working

**Error**: Temporal connection failed

**Solution**:
```bash
# Check Temporal is running
docker compose ps temporal

# Restart Temporal
docker compose restart temporal

# Check UI
open http://localhost:8233

# Verify connection
tctl --address localhost:7233 namespace list
```

### Getting Help

- **Documentation**: Check [docs/](../docs/) folder
- **Issues**: [GitHub Issues](https://github.com/your-org/house-helper/issues)
- **Slack**: #house-helper-dev channel
- **Email**: dev@house-helper.com

---

## Next Steps

### Learn the Codebase

1. **Architecture**: Read [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **API Documentation**: See [API.md](./API.md)
3. **Database Schema**: Review [DATABASE.md](./DATABASE.md)
4. **Testing Strategy**: Understand [TESTING_STRATEGY.md](./TESTING_STRATEGY.md)

### Development Workflow

1. **Create Feature Branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Make Changes**: Edit code

3. **Run Tests**:
   ```bash
   make test
   flutter test
   ```

4. **Commit Changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

5. **Push and Create PR**:
   ```bash
   git push origin feature/my-new-feature
   # Create Pull Request on GitHub
   ```

### Recommended Extensions (VS Code)

```json
{
  "recommendations": [
    "golang.go",
    "dart-code.flutter",
    "dart-code.dart-code",
    "ms-azuretools.vscode-docker",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "humao.rest-client",
    "esbenp.prettier-vscode",
    "dbaeumer.vscode-eslint",
    "ms-vscode.vscode-typescript-next"
  ]
}
```

### Useful Commands

```bash
# View all available make commands
make help

# Format Go code
make fmt

# Lint Go code
make lint

# Build all services
make build

# Clean build artifacts
make clean

# View logs
docker compose logs -f <service-name>

# Access database
make db-shell

# Access Redis
make redis-cli

# Restart all services
docker compose restart
```

### Sample Workflow: Add New Endpoint

1. **Define Route** (`services/api/routes/routes.go`):
   ```go
   api.POST("/tasks/:id/archive", handlers.ArchiveTask)
   ```

2. **Create Handler** (`services/api/handlers/task_handler.go`):
   ```go
   func ArchiveTask(c *gin.Context) {
       // Implementation
   }
   ```

3. **Add Service Method** (`services/api/services/task_service.go`):
   ```go
   func (s *TaskService) ArchiveTask(id string) error {
       // Implementation
   }
   ```

4. **Write Tests** (`services/api/handlers/task_handler_test.go`):
   ```go
   func TestArchiveTask(t *testing.T) {
       // Test implementation
   }
   ```

5. **Update API Docs** (`docs/API.md`):
   ```markdown
   ### Archive Task
   POST /tasks/{id}/archive
   ```

6. **Run Tests**:
   ```bash
   go test ./services/api/handlers -v
   ```

7. **Commit**:
   ```bash
   git add .
   git commit -m "feat: add task archive endpoint"
   ```

---

## Additional Resources

- **Go Documentation**: [go.dev/doc](https://go.dev/doc/)
- **Flutter Documentation**: [flutter.dev/docs](https://flutter.dev/docs)
- **Gin Framework**: [gin-gonic.com/docs](https://gin-gonic.com/docs/)
- **PostgreSQL**: [postgresql.org/docs](https://www.postgresql.org/docs/)
- **Docker Compose**: [docs.docker.com/compose](https://docs.docker.com/compose/)
- **Kubernetes**: [kubernetes.io/docs](https://kubernetes.io/docs/)

---

**Congratulations!** 🎉 You now have a fully functional local development environment.

Happy coding! 💻

---

**Last Updated**: January 2024  
**Maintained by**: Engineering Team

## License

Copyright © 2024 House Helper. All rights reserved.
