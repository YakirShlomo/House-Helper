# House Helper API

The House Helper API is a RESTful service built with Go and Gin that provides backend functionality for the House Helper household management application.

## Features

- **Authentication & Authorization**: JWT-based authentication with refresh tokens
- **Household Management**: Multi-tenant household system with role-based access
- **Task Management**: Create, assign, and track household tasks with priorities and due dates
- **Shopping Lists**: Collaborative shopping lists with real-time updates
- **Bill Tracking**: Manage and track household bills with payment tracking
- **Timer Management**: Pomodoro and countdown timers with Temporal workflow integration
- **Real-time Updates**: WebSocket support for live collaboration
- **Event Streaming**: Kafka integration for domain events
- **Database**: PostgreSQL with migrations
- **Observability**: OpenTelemetry tracing and Prometheus metrics
- **Documentation**: Swagger/OpenAPI documentation

## Architecture

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── handlers/         # HTTP handlers (controllers)
│   ├── middleware/       # HTTP middleware
│   ├── services/         # Business logic layer
│   └── store/           # Data access layer
├── pkg/
│   ├── models/          # Domain models
│   ├── database/        # Database connection
│   ├── jwt/             # JWT utilities
│   ├── kafka/           # Kafka integration
│   ├── temporal/        # Temporal workflows
│   └── validation/      # Input validation
├── migrations/          # Database migrations
└── docs/               # API documentation
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Redis 7+
- Kafka (optional, for events)
- Temporal (optional, for workflows)

### Environment Variables

Create a `.env` file:

```bash
# Database
DATABASE_URL=postgres://user:password@localhost:5432/house_helper?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_USER=house_helper
DB_PASSWORD=your_password
DB_NAME=house_helper
DB_SSL_MODE=disable

# Redis
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-super-secret-jwt-key
JWT_ACCESS_DURATION=15m
JWT_REFRESH_DURATION=7d

# Server
PORT=8080
GIN_MODE=release
CORS_ORIGINS=http://localhost:3000,https://house-helper.com

# Kafka (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=house-helper-events

# Temporal (optional)
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default
TEMPORAL_TASK_QUEUE=house-helper

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
PROMETHEUS_PORT=9090
```

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run database migrations:**
   ```bash
   make migrate-up
   ```

3. **Start the server:**
   ```bash
   make dev
   ```

4. **Generate Swagger docs:**
   ```bash
   make swagger
   ```

### Docker Development

1. **Build and run with Docker Compose:**
   ```bash
   docker-compose up --build
   ```

2. **Run migrations:**
   ```bash
   docker-compose exec api make migrate-up
   ```

## API Documentation

Once the server is running, visit:
- Swagger UI: `http://localhost:8080/docs/index.html`
- OpenAPI spec: `http://localhost:8080/docs/swagger.json`

## Database Migrations

Create a new migration:
```bash
make migrate-create NAME=add_user_table
```

Run migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

## Testing

Run tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Available Commands

See all available commands:
```bash
make help
```

Common commands:
- `make build` - Build the application
- `make run` - Run the application
- `make test` - Run tests
- `make lint` - Run linter
- `make fmt` - Format code
- `make docker-build` - Build Docker image
- `make docker-run` - Run in Docker

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user

### Users
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `PUT /api/v1/users/password` - Change password

### Households
- `GET /api/v1/households` - List user households
- `POST /api/v1/households` - Create household
- `GET /api/v1/households/:id` - Get household details
- `PUT /api/v1/households/:id` - Update household
- `DELETE /api/v1/households/:id` - Delete household
- `POST /api/v1/households/:id/invite` - Invite user
- `POST /api/v1/households/join/:code` - Join household

### Tasks
- `GET /api/v1/tasks` - List tasks
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks/:id` - Get task details
- `PUT /api/v1/tasks/:id` - Update task
- `DELETE /api/v1/tasks/:id` - Delete task
- `POST /api/v1/tasks/:id/complete` - Mark task complete

### Shopping
- `GET /api/v1/shopping/lists` - List shopping lists
- `POST /api/v1/shopping/lists` - Create shopping list
- `GET /api/v1/shopping/lists/:id` - Get shopping list
- `PUT /api/v1/shopping/lists/:id` - Update shopping list
- `DELETE /api/v1/shopping/lists/:id` - Delete shopping list
- `POST /api/v1/shopping/lists/:id/items` - Add item to list
- `PUT /api/v1/shopping/items/:id` - Update shopping item
- `DELETE /api/v1/shopping/items/:id` - Delete shopping item

### Bills
- `GET /api/v1/bills` - List bills
- `POST /api/v1/bills` - Create bill
- `GET /api/v1/bills/:id` - Get bill details
- `PUT /api/v1/bills/:id` - Update bill
- `DELETE /api/v1/bills/:id` - Delete bill
- `POST /api/v1/bills/:id/pay` - Mark bill as paid

### Timers
- `GET /api/v1/timers` - List timers
- `POST /api/v1/timers` - Create timer
- `GET /api/v1/timers/:id` - Get timer details
- `PUT /api/v1/timers/:id` - Update timer
- `DELETE /api/v1/timers/:id` - Delete timer
- `POST /api/v1/timers/:id/start` - Start timer
- `POST /api/v1/timers/:id/stop` - Stop timer

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/new-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.