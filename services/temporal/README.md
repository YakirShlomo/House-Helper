# Temporal Worker & API Service

Production-ready Temporal.io service for House Helper application, providing durable workflow execution for timers, laundry tracking, recurring tasks, and automation.

## 🌟 Features

### Workflows

- **Timer Workflows**: Countdown, stopwatch, and Pomodoro timers with pause/resume
- **Laundry Workflows**: Complete laundry cycle tracking with wash/dry phases and reminders
- **Recurring Task Workflows**: Automated task scheduling with daily/weekly/monthly patterns
- **Task Reminder Workflows**: Smart reminders with escalation for pending tasks

### Key Capabilities

- **Durable Execution**: Workflows survive process restarts and failures
- **Signal Support**: Real-time control via signals (pause, resume, stop)
- **Activity Retry**: Automatic retry with exponential backoff
- **Child Workflows**: Nested workflows for complex orchestration
- **Testing**: Comprehensive test suite with Temporal test framework

## 📋 Requirements

- Go 1.22 or higher
- Temporal Server (local or cloud)
- PostgreSQL (for workflow persistence)

## 🚀 Quick Start

### 1. Start Temporal Server

Using Docker Compose:

```bash
# Start Temporal server with PostgreSQL
docker run --rm -p 7233:7233 temporalio/auto-setup:latest
```

Or use Temporal Cloud for production.

### 2. Configure Environment

```bash
# .env
TEMPORAL_ADDRESS=localhost:7233
TEMPORAL_NAMESPACE=default
PORT=8084
```

### 3. Start Worker

The worker executes workflow and activity tasks:

```bash
cd services/temporal
go run cmd/worker/main.go
```

### 4. Start API Server

The API server provides HTTP endpoints for workflow management:

```bash
cd services/temporal
go run cmd/api/main.go
```

## 📡 API Endpoints

### Timer Workflows

#### Start Timer
```bash
POST /api/v1/workflows/timer/start
Content-Type: application/json

{
  "timerId": "timer-001",
  "userId": "user-001",
  "householdId": "household-001",
  "name": "Cooking Timer",
  "type": "countdown",
  "duration": "300s",
  "settings": {
    "notifyOnStart": true,
    "notifyOnFinish": true
  }
}
```

#### Pause Timer
```bash
POST /api/v1/workflows/timer/pause?timerId=timer-001
```

#### Resume Timer
```bash
POST /api/v1/workflows/timer/resume?timerId=timer-001
```

#### Stop Timer
```bash
POST /api/v1/workflows/timer/stop?timerId=timer-001
```

### Laundry Workflows

#### Start Laundry
```bash
POST /api/v1/workflows/laundry/start
Content-Type: application/json

{
  "laundryId": "laundry-001",
  "userId": "user-001",
  "householdId": "household-001",
  "loadType": "normal",
  "washTime": "1800s",
  "dryTime": "2700s",
  "settings": {
    "notifyOnStart": true,
    "notifyOnWashDone": true,
    "notifyOnDryDone": true,
    "notifyReminders": true,
    "reminderInterval": "600s",
    "maxReminders": 3,
    "autoStart": true
  }
}
```

#### Signal Wash Complete
```bash
POST /api/v1/workflows/laundry/wash-complete?laundryId=laundry-001
```

#### Start Dry Cycle
```bash
POST /api/v1/workflows/laundry/start-dry?laundryId=laundry-001
```

### Recurring Task Workflows

#### Start Recurring Task
```bash
POST /api/v1/workflows/recurring-task/start
Content-Type: application/json

{
  "taskId": "task-001",
  "userId": "user-001",
  "householdId": "household-001",
  "name": "Take out trash",
  "description": "Weekly trash collection",
  "recurrenceRule": {
    "type": "weekly",
    "interval": 1,
    "daysOfWeek": [1, 4],
    "startDate": "2025-10-09T09:00:00Z"
  },
  "assignedMembers": ["user-001", "user-002"],
  "dueDuration": "3600s",
  "reminderSettings": {
    "enabled": true,
    "initialDelay": "3600s",
    "reminderInterval": "1800s",
    "maxReminders": 3,
    "escalateAfter": 2
  },
  "autoAssign": true
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./internal/workflows/... -v

# Run specific test suite
go test ./internal/workflows/ -run TestTimerWorkflowSuite -v

# Run with coverage
go test ./internal/workflows/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🏗️ Architecture

```
services/temporal/
├── cmd/
│   ├── worker/         # Temporal worker (executes workflows)
│   └── api/            # HTTP API for workflow management
├── internal/
│   └── workflows/      # Workflow and activity definitions
│       ├── timer.go              # Timer workflows
│       ├── laundry.go            # Laundry workflows
│       ├── recurring_tasks.go    # Recurring task workflows
│       ├── activities.go         # Shared activities
│       └── workflows_test.go     # Comprehensive tests
└── go.mod
```

## 📊 Workflow Details

### Timer Workflow

Supports three types of timers:

1. **Countdown**: Timer counts down from specified duration
2. **Pomodoro**: Work/break cycles with configurable durations
3. **Stopwatch**: Elapsed time tracking

Features:
- Pause/resume capability
- Durable state (survives restarts)
- Notifications on start/pause/finish
- Configurable settings per timer

### Laundry Workflow

Complete laundry cycle management:

1. **Wash Phase**: Track washing cycle with completion notifications
2. **Dry Phase**: Track drying cycle with optional auto-start
3. **Reminders**: Periodic reminders to move laundry or collect clothes

Features:
- Two-phase workflow (wash → dry)
- Manual or automatic phase transitions
- Smart reminders with configurable intervals
- Load type tracking (normal, delicate, heavy, quick)

### Recurring Task Workflow

Automated task creation with flexible scheduling:

- **Daily**: Every N days
- **Weekly**: Specific days of the week
- **Monthly**: Specific day of month
- **Custom**: Advanced patterns

Features:
- Auto-assignment (round-robin among members)
- Child workflows for reminders
- End date or max occurrences limits
- Task completion tracking

### Task Reminder Workflow

Smart reminder system for tasks:

- Initial delay before first reminder
- Configurable reminder intervals
- Escalation after N reminders
- Automatic stop on task completion

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEMPORAL_ADDRESS` | Temporal server address | `localhost:7233` |
| `TEMPORAL_NAMESPACE` | Temporal namespace | `default` |
| `PORT` | API server port | `8084` |

### Worker Configuration

```go
worker.Options{
    MaxConcurrentActivityExecutionSize:     10,
    MaxConcurrentWorkflowTaskExecutionSize: 10,
}
```

### Activity Configuration

```go
workflow.ActivityOptions{
    StartToCloseTimeout: time.Minute,
    RetryPolicy: &workflow.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    3,
    },
}
```

## 🐳 Docker

### Build Worker Image

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o worker cmd/worker/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/worker .
CMD ["./worker"]
```

### Build API Image

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
EXPOSE 8084
CMD ["./api"]
```

## 🔍 Monitoring

### Temporal Web UI

Access Temporal Web UI for workflow monitoring:

```bash
# Local development
http://localhost:8088
```

### Metrics

Temporal SDK provides built-in metrics:

- Workflow execution count
- Activity execution count
- Task queue lag
- Workflow/activity duration
- Error rates

## 🤝 Integration

### With API Service

The API service creates workflow executions via Temporal client:

```go
import "go.temporal.io/sdk/client"

temporalClient, _ := client.Dial(client.Options{
    HostPort: "localhost:7233",
})

workflowOptions := client.StartWorkflowOptions{
    ID:        "timer-001",
    TaskQueue: "house-helper-tasks",
}

we, _ := temporalClient.ExecuteWorkflow(
    context.Background(),
    workflowOptions,
    workflows.TimerWorkflow,
    params,
)
```

### With Notifier Service

Activities call notifier service for push notifications:

```go
func SendNotificationActivity(ctx context.Context, req NotificationRequest) error {
    // Call notifier service HTTP API
    // POST http://notifier:8083/api/v1/notify/fcm
}
```

## 📚 Resources

- [Temporal Documentation](https://docs.temporal.io/)
- [Go SDK Guide](https://docs.temporal.io/docs/go/)
- [Workflow Patterns](https://docs.temporal.io/docs/go/workflows)
- [Testing Guide](https://docs.temporal.io/docs/go/testing)

## 🔐 Security

- Use Temporal Cloud mTLS for production
- Secure workflow IDs (no sensitive data)
- Implement activity authorization
- Encrypt sensitive workflow data
- Rate limit API endpoints

## 📝 License

MIT License - see LICENSE file for details
