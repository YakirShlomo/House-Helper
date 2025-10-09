# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build consumer binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer cmd/consumer/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/consumer .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep -f consumer || exit 1

# Run as non-root user
RUN addgroup -g 1000 kafka && \
    adduser -D -u 1000 -G kafka kafka && \
    chown -R kafka:kafka /root

USER kafka

# Default environment variables
ENV KAFKA_BROKERS=kafka:9092
ENV KAFKA_GROUP_ID=house-helper-event-consumer

ENTRYPOINT ["./consumer"]
