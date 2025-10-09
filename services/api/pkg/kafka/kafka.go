package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// Config holds Kafka configuration
type Config struct {
	Brokers []string
	Topic   string
}

// Producer wraps kafka writer
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(config Config) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true,
	}

	return &Producer{writer: writer}
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// SendMessage sends a message to Kafka
func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	}

	return p.writer.WriteMessages(ctx, message)
}

// Consumer wraps kafka reader
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config Config, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{reader: reader}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// ReadMessage reads a message from Kafka
func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Event types
const (
	EventTypeTaskCreated    = "task.created"
	EventTypeTaskUpdated    = "task.updated"
	EventTypeTaskCompleted  = "task.completed"
	EventTypeBillCreated    = "bill.created"
	EventTypeBillPaid       = "bill.paid"
	EventTypeTimerStarted   = "timer.started"
	EventTypeTimerCompleted = "timer.completed"
)

// Event represents a domain event
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Source    string      `json:"source"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// EventPublisher publishes domain events
type EventPublisher struct {
	producer *Producer
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(producer *Producer) *EventPublisher {
	return &EventPublisher{producer: producer}
}

// PublishEvent publishes an event
func (ep *EventPublisher) PublishEvent(ctx context.Context, event Event) error {
	log.Printf("Publishing event: %s", event.Type)
	return ep.producer.SendMessage(ctx, event.ID, event)
}
