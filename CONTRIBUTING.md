# Contributing to House Helper

Thank you for your interest in contributing to House Helper! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- Flutter SDK (latest stable)
- Go 1.22+
- Docker & Docker Compose
- Node.js 18+ (for tooling)
- Make

### Local Development

```bash
# Clone the repository
git clone https://github.com/YakirShlomo/House-Helper.git
cd House-Helper

# Start local services
make dev-up

# Run Flutter app
cd apps/mobile_flutter
flutter pub get
flutter run

# Run API service
cd services/api
make run
```

## Project Structure

```
.
├── apps/
│   └── mobile_flutter/     # Flutter mobile app
├── services/
│   └── api/                # Go REST API service
├── infra/
│   ├── terraform/          # AWS infrastructure
│   └── helm/              # Kubernetes charts
├── deploy/
│   └── docker-compose.yml  # Local development
└── docs/                  # Documentation
```

## Code Style

### Flutter/Dart
- Follow the official Dart style guide
- Use `flutter analyze` and `dart format`
- Write widget tests for UI components

### Go
- Follow effective Go guidelines
- Use `go fmt`, `go vet`, and `golangci-lint`
- Write unit tests with good coverage

## Commit Messages

Use conventional commits format:
```
type(scope): description

feat(mobile): add timer widget
fix(api): resolve JWT refresh issue
docs(readme): update setup instructions
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/description`
3. Make your changes
4. Add tests if applicable
5. Ensure all tests pass
6. Update documentation
7. Submit a pull request

## Testing

- Run Flutter tests: `cd apps/mobile_flutter && flutter test`
- Run Go tests: `cd services/api && go test ./...`
- Run integration tests: `make test-integration`

## Security

- Never commit secrets or credentials
- Use environment variables for configuration
- Follow the security guidelines in SECURITY.md

## Questions?

Open an issue or contact the maintainers.