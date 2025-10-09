# House Helper

A comprehensive household management app built for production-scale deployment with Flutter, Go, and modern cloud infrastructure.

## Overview

House Helper is a full-featured household management application that helps families organize tasks, shopping lists, bills, and timers. Built with production-grade architecture including microservices, event-driven patterns, and cloud-native deployment.

### Key Features

- **Smart Task Management**: Create, assign, and track household tasks
- **Live Shopping Lists**: Real-time collaborative shopping with family members
- **Bill Tracking**: Recurring payment reminders and history
- **Durable Timers**: Reliable timers for laundry, cooking, etc. (powered by Temporal)
- **Smart Notifications**: Context-aware push notifications with quiet hours
- **Native Widgets**: iOS and Android home screen widgets and shortcuts
- **Multi-language**: Hebrew and English with full RTL support
- **Offline Support**: Works without internet, syncs when connected

## Technology Stack

### Mobile App
- **Flutter 3.x** with Material 3 design
- **Dart 3.x** with null safety
- **Riverpod** for state management
- **Go Router** for navigation
- **Drift** for local database
- **Firebase Messaging** for push notifications

### Backend Services
- **Go 1.22+** with Gin web framework
- **PostgreSQL** for primary data storage
- **Redis** for caching and sessions
- **Apache Kafka** for event streaming
- **Temporal** for durable workflows and timers
- **OpenTelemetry** for observability

### Infrastructure
- **AWS EKS** (Kubernetes)
- **AWS RDS** (PostgreSQL)
- **AWS ElastiCache** (Redis)
- **AWS MSK** (Kafka)
- **AWS S3** for file storage
- **Terraform** for infrastructure as code
- **Helm** for Kubernetes deployments

### DevOps & CI/CD
- **GitHub Actions** for CI/CD
- **Docker** containerization
- **Amazon ECR** for container registry
- **Prometheus & Grafana** for monitoring
- **Sentry** for error tracking

## Quick Start

### Prerequisites

- Flutter SDK (latest stable)
- Go 1.22+
- Docker & Docker Compose
- Make
- Node.js 18+ (for tooling)

### Local Development

1. **Clone and setup**:
   ```bash
   git clone https://github.com/YakirShlomo/House-Helper.git
   cd House-Helper
   ```

2. **Start local services**:
   ```bash
   make dev-up
   ```

3. **Run the Flutter app**:
   ```bash
   cd apps/mobile_flutter
   flutter pub get
   flutter run
   ```

4. **Run the API service**:
   ```bash
   cd services/api
   make run
   ```

5. **Demo the timer workflow**:
   ```bash
   make demo-timer
   ```

### Production Deployment

1. **Infrastructure setup**:
   ```bash
   cd infra/terraform/envs/prod
   terraform init
   terraform plan
   terraform apply
   ```

2. **Deploy services**:
   ```bash
   cd infra/helm
   helm upgrade --install house-helper-api ./api -f values-prod.yaml
   ```

## Project Structure

```
.
├── apps/
│   └── mobile_flutter/     # Flutter mobile application
│       ├── lib/
│       │   ├── screens/    # UI screens
│       │   ├── providers/  # Riverpod state providers
│       │   ├── services/   # API and local services
│       │   └── models/     # Data models
│       ├── ios/            # iOS-specific code and widgets
│       └── android/        # Android-specific code and widgets
├── services/
│   └── api/                # Go REST API service
│       ├── cmd/api/        # Application entry point
│       ├── internal/       # Private application code
│       │   ├── handlers/   # HTTP handlers
│       │   ├── services/   # Business logic
│       │   ├── store/      # Data access layer
│       │   └── middleware/ # HTTP middleware
│       └── pkg/            # Public packages
├── infra/
│   ├── terraform/          # AWS infrastructure definitions
│   │   ├── modules/        # Reusable modules
│   │   └── envs/           # Environment configurations
│   └── helm/               # Kubernetes deployment charts
│       ├── api/            # API service chart
│       ├── notifier/       # Notification service chart
│       └── temporal/       # Temporal workflow engine chart
├── deploy/
│   └── docker-compose.yml  # Local development environment
└── docs/                   # Documentation
    ├── PRD.md              # Product requirements
    ├── APIs.md             # API documentation
    ├── DB.md               # Database schema
    └── SECURITY.md         # Security guidelines
```

## Performance Targets

- **Cold Start**: ≤ 2.5 seconds
- **API Response**: P95 ≤ 200-250ms (Israel region)
- **Crash-free Rate**: ≥ 99.8%
- **Notification SLA**: ≥ 99%

## Security

- **TLS 1.3** for all communications
- **JWT tokens** with short expiry + refresh
- **RBAC** per household
- **Encryption at rest** and in transit
- **WAF** and rate limiting
- **Secrets management** via AWS Secrets Manager

## Development Commands

```bash
# Local development
make dev-up              # Start all local services
make dev-down            # Stop all local services
make dev-logs            # View service logs
make seed                # Load demo data

# Testing
make test                # Run all tests
make test-mobile         # Run Flutter tests
make test-api            # Run Go tests
make test-integration    # Run integration tests
make test-load           # Run k6 load tests

# Building
make build               # Build all components
make build-mobile        # Build Flutter apps
make build-api           # Build Go API
make docker-build        # Build Docker images

# Deployment
make deploy-dev          # Deploy to development
make deploy-prod         # Deploy to production
make infra-plan          # Terraform plan
make infra-apply         # Terraform apply
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/YakirShlomo/House-Helper/issues)
- **Security**: See [SECURITY.md](SECURITY.md) for reporting security issues

---

**Note**: This is a production-ready codebase with placeholders for sensitive configuration. Follow the security guidelines in `SECURITY.md` for proper secrets management.