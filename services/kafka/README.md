# Kafka Event Service

Production-ready Apache Kafka event streaming service for House Helper application, providing reliable event-driven communication between microservices.

## 🌟 Features

- **Event Producer**: Publish domain events to Kafka topics
- **Event Consumer**: Subscribe to and process events from multiple topics
- **Event Log**: Persistent storage of all events for audit and replay
- **Topic Organization**: Logical event grouping by domain
- **Idempotent Publishing**: Guaranteed at-least-once delivery
- **Consumer Groups**: Scalable event processing

## 📡 Event Topics

| Topic | Events | Purpose |
|-------|--------|---------|
| `house-helper.tasks` | task.created, task.updated, task.completed, task.deleted | Task lifecycle events |
| `house-helper.shopping` | shopping.item.added, shopping.item.purchased, etc. | Shopping list events |
| `house-helper.bills` | bill.created, bill.updated, bill.paid, bill.overdue | Bill management events |
| `house-helper.timers` | timer.started, timer.paused, timer.completed | Timer lifecycle events |
| `house-helper.laundry` | laundry.started, laundry.wash.complete, laundry.dry.complete | Laundry cycle events |
| `house-helper.households` | household.created, household.member.added, household.activity | Household events |
| `house-helper.users` | user.registered, user.logged.in, user.updated | User events |
| `house-helper.notifications` | notification.sent, notification.clicked | Notification tracking |

## 🚀 Quick Start

### 1. Start Kafka

Using Docker:

```bash
docker run -d --name kafka -p 9092:9092 \
  apache/kafka:latest
```

### 2. Create Topics

```bash
# Create all topics
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.tasks --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.shopping --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.bills --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.timers --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.laundry --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.households --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.users --partitions 3 --replication-factor 1
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic house-helper.notifications --partitions 3 --replication-factor 1
```

### 3. Start Consumer

```bash
cd services/kafka
go run cmd/consumer/main.go
```

## 📦 Usage

### Publishing Events

```go
import (
    "github.com/househelper/kafka/pkg/producer"
    "github.com/househelper/kafka/pkg/events"
)

// Create producer
prod, err := producer.NewProducer(producer.Config{
    Brokers: []string{"localhost:9092"},
    Logger:  logger,
})
defer prod.Close()

// Publish task completed event
err = prod.PublishTaskEvent(
    events.EventTaskCompleted,
    events.TaskEventData{
        TaskID:      "task-001",
        HouseholdID: "household-001",
        Title:       "Take out trash",
        Status:      "completed",
        CompletedBy: "user-001",
        CompletedAt: time.Now(),
    },
    "user-001",
)
```

### Consuming Events

```go
import (
    "github.com/househelper/kafka/pkg/consumer"
    "github.com/househelper/kafka/pkg/events"
)

// Create consumer
cons, err := consumer.NewConsumer(consumer.Config{
    Brokers: []string{"localhost:9092"},
    GroupID: "my-service",
    Topics:  []string{"house-helper.tasks"},
    Logger:  logger,
})
defer cons.Close()

// Register handler
cons.RegisterHandler(events.EventTaskCompleted, func(ctx context.Context, event *events.Event) error {
    taskID, _ := event.Data["taskId"].(string)
    fmt.Printf("Task completed: %s\n", taskID)
    return nil
})

// Start consuming
err = cons.Start(context.Background())
```

## 🏗️ Architecture

```
services/kafka/
├── cmd/
│   └── consumer/       # Event consumer service
├── pkg/
│   ├── events/         # Event type definitions
│   ├── producer/       # Kafka producer
│   ├── consumer/       # Kafka consumer
│   └── eventlog/       # Event persistence
└── go.mod
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BROKERS` | Comma-separated broker list | `localhost:9092` |
| `KAFKA_GROUP_ID` | Consumer group ID | `house-helper-event-consumer` |

## 📝 Event Structure

All events follow a consistent structure:

```json
{
  "id": "20251009120000-abc123",
  "type": "task.completed",
  "source": "api-service",
  "householdId": "household-001",
  "userId": "user-001",
  "timestamp": "2025-10-09T12:00:00Z",
  "version": "1.0",
  "data": {
    "taskId": "task-001",
    "title": "Take out trash",
    "status": "completed",
    "completedBy": "user-001",
    "completedAt": "2025-10-09T12:00:00Z"
  },
  "metadata": {
    "correlationId": "req-xyz789"
  }
}
```

## 🔐 Security

- Use SASL/SSL for production
- Implement ACLs for topic access
- Encrypt sensitive event data
- Audit event access patterns

## 📊 Monitoring

Monitor these metrics:

- Message throughput (messages/sec)
- Consumer lag (messages behind)
- Processing time (ms/message)
- Error rate (%)
- Topic size (GB)

## 📝 License

MIT License - see LICENSE file for details
