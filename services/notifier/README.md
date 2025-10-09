# House Helper Notifier Service

The House Helper Notifier Service handles push notifications and real-time communication for the House Helper application.

## Features

- **Push Notifications**: Firebase Cloud Messaging (FCM) and Apple Push Notification Service (APNS)
- **Real-time Communication**: WebSocket and Server-Sent Events (SSE)
- **Multi-platform Support**: iOS and Android notifications
- **User & Household Broadcasting**: Targeted messaging to specific users or entire households
- **Live Activities**: Support for iOS Live Activities and Dynamic Island
- **Silent Notifications**: Background sync and data updates

## Architecture

```
├── cmd/notifier/         # Application entry point
├── pkg/
│   ├── notifications/    # Push notification services (FCM/APNS)
│   ├── websocket/        # WebSocket hub and client management
│   └── sse/             # Server-Sent Events implementation
└── docs/                # Documentation
```

## Getting Started

### Prerequisites

- Go 1.22+
- Firebase project with FCM enabled
- Apple Developer account with APNS certificates/keys

### Environment Variables

Create a `.env` file:

```bash
# Server
PORT=:8080

# Firebase Cloud Messaging
FCM_CREDENTIALS_PATH=/path/to/firebase-credentials.json

# Apple Push Notification Service (Option 1: Certificate)
APNS_CERT_PATH=/path/to/apns-cert.pem
APNS_KEY_PATH=/path/to/apns-key.pem
APNS_PRODUCTION=false

# Apple Push Notification Service (Option 2: Auth Key)
APNS_KEY_PATH=/path/to/AuthKey_XXXXXXXXXX.p8
APNS_KEY_ID=XXXXXXXXXX
APNS_TEAM_ID=YYYYYYYYYY
APNS_PRODUCTION=false
```

### Local Development

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Start the service:**
   ```bash
   go run ./cmd/notifier
   ```

3. **Test WebSocket connection:**
   ```bash
   # Connect to WebSocket
   wscat -c "ws://localhost:8080/ws?userId=user123&householdId=house456"
   ```

4. **Test Server-Sent Events:**
   ```bash
   # Connect to SSE stream
   curl "http://localhost:8080/events?userId=user123&householdId=house456"
   ```

### Docker Development

1. **Build and run:**
   ```bash
   docker build -t house-helper/notifier .
   docker run -p 8080:8080 -e PORT=:8080 house-helper/notifier
   ```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Real-time Communication
- `GET /ws?userId={id}&householdId={id}` - WebSocket connection
- `GET /events?userId={id}&householdId={id}` - Server-Sent Events stream

### Push Notifications

#### Firebase Cloud Messaging
- `POST /notify/fcm/token` - Send notification to specific device token
- `POST /notify/fcm/topic` - Send notification to topic subscribers

#### Apple Push Notification Service
- `POST /notify/apns` - Send standard APNS notification
- `POST /notify/apns/silent` - Send silent background notification

### Broadcasting
- `POST /broadcast/user?userId={id}` - Broadcast message to specific user
- `POST /broadcast/household?householdId={id}` - Broadcast message to household

## Message Examples

### FCM Token Notification
```bash
curl -X POST http://localhost:8080/notify/fcm/token \
  -H "Content-Type: application/json" \
  -d '{
    "token": "device-fcm-token",
    "title": "Task Completed",
    "body": "John completed the laundry task",
    "imageUrl": "https://example.com/image.png",
    "data": {
      "taskId": "task123",
      "type": "task_completed"
    }
  }'
```

### APNS Notification
```bash
curl -X POST http://localhost:8080/notify/apns \
  -H "Content-Type: application/json" \
  -d '{
    "deviceToken": "apns-device-token",
    "bundleId": "app.househelper.mobile",
    "title": "Timer Finished",
    "body": "Your laundry timer has finished",
    "badge": 1,
    "sound": "default",
    "customData": {
      "timerId": "timer123",
      "type": "timer_completed"
    }
  }'
```

### User Broadcast
```bash
curl -X POST "http://localhost:8080/broadcast/user?userId=user123" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "task_assigned",
    "data": {
      "taskId": "task123",
      "title": "Clean kitchen",
      "assignedBy": "Jane"
    }
  }'
```

### Household Broadcast
```bash
curl -X POST "http://localhost:8080/broadcast/household?householdId=house456" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "shopping_item_purchased",
    "data": {
      "itemId": "item123",
      "name": "Milk",
      "purchasedBy": "John"
    }
  }'
```

## WebSocket Message Format

Messages sent and received via WebSocket follow this format:

```json
{
  "type": "task_completed",
  "userId": "user123",
  "householdId": "house456",
  "data": {
    "taskId": "task123",
    "title": "Laundry",
    "completedBy": "John"
  },
  "timestamp": "2023-10-09T10:30:00Z"
}
```

## Server-Sent Events Format

SSE messages follow the standard format:

```
id: 1234567890
event: task_completed
data: {"taskId":"task123","title":"Laundry","completedBy":"John"}

```

## Push Notification Topics

### FCM Topics
- `household_{householdId}` - All members of a household
- `user_{userId}` - Specific user across all devices
- `tasks_due` - Users with upcoming task deadlines
- `bills_due` - Users with upcoming bill payments

### Notification Categories

#### Task Notifications
- `task_assigned` - Task assigned to user
- `task_completed` - Task marked as complete
- `task_overdue` - Task past due date
- `task_reminder` - Task due soon

#### Shopping Notifications
- `item_added` - Item added to shared shopping list
- `item_purchased` - Item marked as purchased
- `list_shared` - Shopping list shared with user

#### Bill Notifications
- `bill_due` - Bill payment due soon
- `bill_paid` - Bill marked as paid
- `bill_overdue` - Bill payment overdue

#### Timer Notifications
- `timer_started` - Timer started
- `timer_finished` - Timer completed
- `timer_paused` - Timer paused

## Error Handling

The service handles various error scenarios:

- **FCM Errors**: Invalid tokens, quota exceeded, authentication failures
- **APNS Errors**: Invalid device tokens, certificate issues, payload too large
- **WebSocket Errors**: Connection drops, invalid messages, authentication failures
- **SSE Errors**: Client disconnects, network issues

Error responses follow this format:

```json
{
  "error": "invalid_token",
  "message": "The provided device token is invalid",
  "timestamp": "2023-10-09T10:30:00Z"
}
```

## Monitoring & Observability

The service provides several monitoring capabilities:

- **Health Checks**: `/health` endpoint for service status
- **Metrics**: Connection counts, message delivery rates, error rates
- **Logging**: Structured logging for debugging and monitoring
- **Tracing**: Request tracing for performance analysis

## Security

- **CORS**: Configurable CORS settings for web clients
- **Rate Limiting**: Protection against abuse and spam
- **Token Validation**: Verification of device tokens and auth tokens
- **SSL/TLS**: Encrypted connections for production deployments

## Production Deployment

### Environment Configuration
```bash
# Production settings
PORT=:8080
APNS_PRODUCTION=true

# FCM credentials
FCM_CREDENTIALS_PATH=/secrets/firebase-credentials.json

# APNS credentials
APNS_KEY_PATH=/secrets/AuthKey_XXXXXXXXXX.p8
APNS_KEY_ID=XXXXXXXXXX
APNS_TEAM_ID=YYYYYYYYYY
```

### Docker Deployment
```bash
docker run -d \
  --name house-helper-notifier \
  -p 8080:8080 \
  -v /path/to/secrets:/secrets \
  -e FCM_CREDENTIALS_PATH=/secrets/firebase-credentials.json \
  -e APNS_KEY_PATH=/secrets/AuthKey_XXXXXXXXXX.p8 \
  -e APNS_KEY_ID=XXXXXXXXXX \
  -e APNS_TEAM_ID=YYYYYYYYYY \
  -e APNS_PRODUCTION=true \
  house-helper/notifier
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature/new-feature`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.