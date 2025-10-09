# House Helper - Docker Compose Demo

Complete local development environment with all services orchestrated.

## 🚀 Quick Start

### 1. Prerequisites

- Docker Desktop 4.0+ or Docker Engine 20.10+
- Docker Compose V2
- 8GB RAM minimum (16GB recommended)
- 20GB free disk space

### 2. Start All Services

```bash
# Clone the repository
git clone https://github.com/YourOrg/House-Helper.git
cd House-Helper

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check service status
docker-compose ps
```

### 3. Access Services

| Service | URL | Purpose |
|---------|-----|---------|
| **API** | http://localhost:8080 | REST API endpoints |
| **Adminer** | http://localhost:8081 | PostgreSQL database UI |
| **Kafka UI** | http://localhost:8082 | Kafka topics and messages |
| **Notifier** | http://localhost:8083 | Push notification service |
| **Temporal API** | http://localhost:8084 | Workflow management API |
| **Redis Commander** | http://localhost:8085 | Redis cache UI |
| **Temporal UI** | http://localhost:8088 | Temporal workflow UI |

### 4. Initialize Database

```bash
# Run database migrations
docker-compose exec api /app/migrate up

# Or manually connect to PostgreSQL
docker-compose exec postgres psql -U househelper -d househelper
```

### 5. Create Kafka Topics

```bash
# Topics are auto-created, but you can manually create them:
docker-compose exec kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --topic house-helper.tasks --partitions 3 --replication-factor 1

# List all topics
docker-compose exec kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 --list
```

## 📋 Service Details

### Core Services

#### PostgreSQL Database
- **Port**: 5432
- **Database**: househelper
- **User**: househelper
- **Password**: househelper_dev_pass

#### Redis Cache
- **Port**: 6379
- **No authentication** (development only)

#### Apache Kafka
- **Port**: 9092
- **Topics**: Auto-created on first publish
- **Partitions**: 3 per topic

#### Temporal Server
- **Port**: 7233
- **Namespace**: default
- **UI**: http://localhost:8088

### Application Services

#### API Service
- **Port**: 8080
- **Health**: http://localhost:8080/health
- **Docs**: http://localhost:8080/swagger/index.html

#### Notifier Service
- **Port**: 8083
- **Health**: http://localhost:8083/health
- **Supports**: FCM, APNS, WebSocket, SSE

#### Temporal Worker
- Executes workflow and activity tasks
- Connects to Temporal server
- Processes: timers, laundry, recurring tasks

#### Temporal API
- **Port**: 8084
- **Health**: http://localhost:8084/health
- HTTP endpoints for workflow management

#### Kafka Consumer
- Consumes events from all topics
- Stores events in event log
- Triggers downstream processing

## 🧪 Testing the Stack

### 1. Check All Services Are Running

```bash
docker-compose ps

# All services should show "Up" or "Up (healthy)"
```

### 2. Test API Health

```bash
curl http://localhost:8080/health

# Expected: {"status":"healthy"}
```

### 3. Create a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "fullName": "Test User"
  }'
```

### 4. Create a Task

```bash
# First, login to get JWT token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "SecurePass123!"}' \
  | jq -r '.token')

# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Take out trash",
    "description": "Weekly trash collection",
    "priority": "high",
    "dueDate": "2025-10-10T09:00:00Z"
  }'
```

### 5. Start a Timer Workflow

```bash
curl -X POST http://localhost:8084/api/v1/workflows/timer/start \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

### 6. View Kafka Events

Open Kafka UI at http://localhost:8082 and navigate to topics to see published events.

### 7. View Temporal Workflows

Open Temporal UI at http://localhost:8088 to see running workflows.

## 🛠️ Development Workflow

### Hot Reload

Services support hot reload for development:

```bash
# Edit Go code in services/api/
# Changes are automatically detected and service restarts

# View service logs
docker-compose logs -f api
```

### Debugging

#### View Service Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f temporal-worker
```

#### Access Container Shell

```bash
# API container
docker-compose exec api sh

# PostgreSQL
docker-compose exec postgres psql -U househelper -d househelper
```

#### Database Queries

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U househelper -d househelper

# Example queries
SELECT * FROM users LIMIT 10;
SELECT * FROM tasks WHERE household_id = 'household-001';
SELECT * FROM event_log ORDER BY timestamp DESC LIMIT 20;
```

## 🔧 Configuration

### Environment Variables

Edit `docker-compose.yml` to customize service configurations:

```yaml
environment:
  - DATABASE_URL=postgres://...
  - REDIS_URL=redis://...
  - JWT_SECRET=your_secret_here
```

### Volume Persistence

Data is persisted in Docker volumes:

```bash
# List volumes
docker volume ls | grep househelper

# Inspect volume
docker volume inspect househelper-postgres-data

# Backup database
docker-compose exec postgres pg_dump -U househelper househelper > backup.sql

# Restore database
docker-compose exec -T postgres psql -U househelper househelper < backup.sql
```

## 🧹 Maintenance

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes all data)
docker-compose down -v
```

### Update Services

```bash
# Rebuild specific service
docker-compose build api
docker-compose up -d api

# Rebuild all services
docker-compose build
docker-compose up -d
```

### Clean Up

```bash
# Remove stopped containers
docker-compose rm -f

# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune
```

### View Resource Usage

```bash
# Show resource stats
docker stats

# Show disk usage
docker system df
```

## 🐛 Troubleshooting

### Service Won't Start

```bash
# Check logs
docker-compose logs <service-name>

# Check if port is already in use
netstat -an | grep <port>

# Restart specific service
docker-compose restart <service-name>
```

### Database Connection Issues

```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U househelper
```

### Kafka Connection Issues

```bash
# Check Kafka broker
docker-compose logs kafka

# List topics
docker-compose exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --list

# Describe topic
docker-compose exec kafka kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic house-helper.tasks
```

### Temporal Issues

```bash
# Check Temporal server health
docker-compose exec temporal tctl cluster health

# List workflows
docker-compose exec temporal tctl workflow list

# Describe workflow
docker-compose exec temporal tctl workflow describe -w <workflow-id>
```

## 📚 Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Temporal Documentation](https://docs.temporal.io/)
- [Redis Documentation](https://redis.io/documentation)

## 🔐 Security Notes

⚠️ **This configuration is for local development only!**

For production:
- Change all default passwords
- Enable SSL/TLS for all services
- Use secrets management (e.g., HashiCorp Vault)
- Implement proper authentication/authorization
- Enable security features in each service
- Use private networks and firewalls
- Regular security updates and patches

## 📝 License

MIT License - see LICENSE file for details
