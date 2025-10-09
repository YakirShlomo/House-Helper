package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/househelper/kafka/pkg/events"
	"go.uber.org/zap"
)

// Handler is a function that processes an event
type Handler func(ctx context.Context, event *events.Event) error

// Consumer wraps Kafka consumer functionality
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topics        []string
	handlers      map[events.EventType][]Handler
	logger        *zap.Logger
	mu            sync.RWMutex
}

// Config holds consumer configuration
type Config struct {
	Brokers  []string
	GroupID  string
	Topics   []string
	Logger   *zap.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg Config) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V3_6_0_0

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topics:        cfg.Topics,
		handlers:      make(map[events.EventType][]Handler),
		logger:        cfg.Logger,
	}, nil
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(eventType events.EventType, handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.handlers[eventType] = append(c.handlers[eventType], handler)
	c.logger.Info("Registered handler",
		zap.String("eventType", string(eventType)),
	)
}

// Start starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	handler := &consumerGroupHandler{
		consumer: c,
		logger:   c.logger,
	}

	c.logger.Info("Starting consumer",
		zap.Strings("topics", c.topics),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consumer")
			return nil
		default:
			err := c.consumerGroup.Consume(ctx, c.topics, handler)
			if err != nil {
				c.logger.Error("Consumer error", zap.Error(err))
				return err
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	consumer *Consumer
	logger   *zap.Logger
}

// Setup is run at the beginning of a new session
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session started")
	return nil
}

// Cleanup is run at the end of a session
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session ended")
	return nil
}

// ConsumeClaim processes messages from a topic partition
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		ctx := session.Context()
		
		// Parse event
		event, err := events.FromJSON(message.Value)
		if err != nil {
			h.logger.Error("Failed to parse event",
				zap.String("topic", message.Topic),
				zap.Int32("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.Error(err),
			)
			session.MarkMessage(message, "")
			continue
		}

		h.logger.Debug("Processing event",
			zap.String("eventId", event.ID),
			zap.String("eventType", string(event.Type)),
			zap.String("topic", message.Topic),
		)

		// Get handlers for this event type
		h.consumer.mu.RLock()
		handlers, exists := h.consumer.handlers[event.Type]
		h.consumer.mu.RUnlock()

		if !exists || len(handlers) == 0 {
			h.logger.Debug("No handlers registered for event type",
				zap.String("eventType", string(event.Type)),
			)
			session.MarkMessage(message, "")
			continue
		}

		// Execute all registered handlers
		var handlerErrors []error
		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				h.logger.Error("Handler error",
					zap.String("eventId", event.ID),
					zap.String("eventType", string(event.Type)),
					zap.Error(err),
				)
				handlerErrors = append(handlerErrors, err)
			}
		}

		// Mark message as processed if all handlers succeeded
		if len(handlerErrors) == 0 {
			session.MarkMessage(message, "")
			h.logger.Debug("Event processed successfully",
				zap.String("eventId", event.ID),
				zap.String("eventType", string(event.Type)),
			)
		} else {
			h.logger.Error("Failed to process event",
				zap.String("eventId", event.ID),
				zap.Int("handlerErrors", len(handlerErrors)),
			)
			// Message will not be marked, will be reprocessed
		}
	}

	return nil
}

// Example handler functions

// TaskCompletedHandler handles task completed events
func TaskCompletedHandler(ctx context.Context, event *events.Event) error {
	// Extract task data from event
	taskID, _ := event.Data["taskId"].(string)
	householdID, _ := event.Data["householdId"].(string)
	completedBy, _ := event.Data["completedBy"].(string)

	// Business logic here:
	// 1. Update task statistics
	// 2. Award points to user
	// 3. Send notifications to household members
	// 4. Update household activity feed

	fmt.Printf("Task completed: %s by %s in household %s\n", taskID, completedBy, householdID)
	return nil
}

// ShoppingItemPurchasedHandler handles shopping item purchased events
func ShoppingItemPurchasedHandler(ctx context.Context, event *events.Event) error {
	itemID, _ := event.Data["itemId"].(string)
	householdID, _ := event.Data["householdId"].(string)
	purchasedBy, _ := event.Data["purchasedBy"].(string)

	// Business logic here:
	// 1. Update shopping list statistics
	// 2. Send notifications to list creator
	// 3. Check if all items are purchased
	// 4. Update household activity feed

	fmt.Printf("Shopping item purchased: %s by %s in household %s\n", itemID, purchasedBy, householdID)
	return nil
}

// BillPaidHandler handles bill paid events
func BillPaidHandler(ctx context.Context, event *events.Event) error {
	billID, _ := event.Data["billId"].(string)
	householdID, _ := event.Data["householdId"].(string)
	paidBy, _ := event.Data["paidBy"].(string)
	amount, _ := event.Data["amount"].(float64)

	// Business logic here:
	// 1. Update bill payment records
	// 2. Calculate bill splits
	// 3. Send payment confirmations
	// 4. Update household expenses

	fmt.Printf("Bill paid: %s ($%.2f) by %s in household %s\n", billID, amount, paidBy, householdID)
	return nil
}

// TimerFinishedHandler handles timer finished events
func TimerFinishedHandler(ctx context.Context, event *events.Event) error {
	timerID, _ := event.Data["timerId"].(string)
	userID, _ := event.Data["userId"].(string)
	name, _ := event.Data["name"].(string)

	// Business logic here:
	// 1. Send completion notification
	// 2. Log timer completion
	// 3. Update user activity

	fmt.Printf("Timer finished: %s (%s) for user %s\n", name, timerID, userID)
	return nil
}

// HouseholdActivityHandler handles household activity events
func HouseholdActivityHandler(ctx context.Context, event *events.Event) error {
	householdID, _ := event.Data["householdId"].(string)
	actorID, _ := event.Data["actorId"].(string)
	activity, _ := event.Data["activity"].(string)

	// Business logic here:
	// 1. Update activity feed
	// 2. Send real-time updates to household members
	// 3. Update activity statistics

	fmt.Printf("Household activity: %s in %s by %s\n", activity, householdID, actorID)
	return nil
}
