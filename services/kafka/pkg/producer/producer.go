package producer

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/househelper/kafka/pkg/events"
	"go.uber.org/zap"
)

// Producer wraps Kafka producer functionality
type Producer struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

// Config holds producer configuration
type Config struct {
	Brokers []string
	Logger  *zap.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg Config) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all replicas
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Idempotent = true // Enable idempotent producer
	config.Net.MaxOpenRequests = 1

	// Set version to enable idempotence
	config.Version = sarama.V3_6_0_0

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		producer: producer,
		logger:   cfg.Logger,
	}, nil
}

// PublishEvent publishes an event to the appropriate topic
func (p *Producer) PublishEvent(event *events.Event) error {
	topic := p.getTopicForEvent(event.Type)
	
	eventJSON, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(event.ID),
		Value: sarama.ByteEncoder(eventJSON),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("event-type"),
				Value: []byte(event.Type),
			},
			{
				Key:   []byte("event-source"),
				Value: []byte(event.Source),
			},
		},
	}

	// Add household ID as partition key if present
	if event.HouseholdID != "" {
		msg.Key = sarama.StringEncoder(event.HouseholdID)
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Failed to publish event",
			zap.String("eventId", event.ID),
			zap.String("eventType", string(event.Type)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Debug("Event published",
		zap.String("eventId", event.ID),
		zap.String("eventType", string(event.Type)),
		zap.String("topic", topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// PublishTaskEvent publishes a task-related event
func (p *Producer) PublishTaskEvent(eventType events.EventType, data events.TaskEventData, userID string) error {
	eventData := map[string]interface{}{
		"taskId":      data.TaskID,
		"householdId": data.HouseholdID,
		"title":       data.Title,
		"status":      data.Status,
		"assignedTo":  data.AssignedTo,
	}

	if eventType == events.EventTaskCompleted {
		eventData["completedBy"] = data.CompletedBy
		eventData["completedAt"] = data.CompletedAt
	}

	event := events.NewEvent(eventType, "api-service", eventData)
	event.HouseholdID = data.HouseholdID
	event.UserID = userID

	return p.PublishEvent(event)
}

// PublishShoppingEvent publishes a shopping-related event
func (p *Producer) PublishShoppingEvent(eventType events.EventType, data events.ShoppingEventData, userID string) error {
	eventData := map[string]interface{}{
		"itemId":      data.ItemID,
		"listId":      data.ListID,
		"householdId": data.HouseholdID,
		"name":        data.Name,
		"quantity":    data.Quantity,
		"category":    data.Category,
		"status":      data.Status,
	}

	if eventType == events.EventShoppingItemPurchased {
		eventData["purchasedBy"] = data.PurchasedBy
		eventData["purchasedAt"] = data.PurchasedAt
	}

	event := events.NewEvent(eventType, "api-service", eventData)
	event.HouseholdID = data.HouseholdID
	event.UserID = userID

	return p.PublishEvent(event)
}

// PublishBillEvent publishes a bill-related event
func (p *Producer) PublishBillEvent(eventType events.EventType, data events.BillEventData, userID string) error {
	eventData := map[string]interface{}{
		"billId":      data.BillID,
		"householdId": data.HouseholdID,
		"name":        data.Name,
		"amount":      data.Amount,
		"currency":    data.Currency,
		"status":      data.Status,
		"dueDate":     data.DueDate,
	}

	if eventType == events.EventBillPaid {
		eventData["paidBy"] = data.PaidBy
		eventData["paidAt"] = data.PaidAt
	}

	event := events.NewEvent(eventType, "api-service", eventData)
	event.HouseholdID = data.HouseholdID
	event.UserID = userID

	return p.PublishEvent(event)
}

// PublishTimerEvent publishes a timer-related event
func (p *Producer) PublishTimerEvent(eventType events.EventType, data events.TimerEventData) error {
	eventData := map[string]interface{}{
		"timerId":     data.TimerID,
		"userId":      data.UserID,
		"householdId": data.HouseholdID,
		"name":        data.Name,
		"type":        data.Type,
		"duration":    data.Duration.String(),
		"status":      data.Status,
	}

	if data.ElapsedTime > 0 {
		eventData["elapsedTime"] = data.ElapsedTime.String()
	}

	event := events.NewEvent(eventType, "temporal-service", eventData)
	event.HouseholdID = data.HouseholdID
	event.UserID = data.UserID

	return p.PublishEvent(event)
}

// PublishLaundryEvent publishes a laundry-related event
func (p *Producer) PublishLaundryEvent(eventType events.EventType, data events.LaundryEventData) error {
	eventData := map[string]interface{}{
		"laundryId":   data.LaundryID,
		"userId":      data.UserID,
		"householdId": data.HouseholdID,
		"loadType":    data.LoadType,
		"status":      data.Status,
	}

	if data.WashTime > 0 {
		eventData["washTime"] = data.WashTime.String()
	}
	if data.DryTime > 0 {
		eventData["dryTime"] = data.DryTime.String()
	}

	event := events.NewEvent(eventType, "temporal-service", eventData)
	event.HouseholdID = data.HouseholdID
	event.UserID = data.UserID

	return p.PublishEvent(event)
}

// PublishHouseholdActivity publishes a household activity event
func (p *Producer) PublishHouseholdActivity(householdID, actorID, activity string) error {
	eventData := map[string]interface{}{
		"householdId": householdID,
		"actorId":     actorID,
		"activity":    activity,
	}

	event := events.NewEvent(events.EventHouseholdActivity, "api-service", eventData)
	event.HouseholdID = householdID
	event.UserID = actorID

	return p.PublishEvent(event)
}

// getTopicForEvent returns the appropriate Kafka topic for an event type
func (p *Producer) getTopicForEvent(eventType events.EventType) string {
	switch eventType {
	case events.EventTaskCreated, events.EventTaskUpdated, events.EventTaskCompleted, events.EventTaskDeleted:
		return "house-helper.tasks"
	
	case events.EventShoppingItemAdded, events.EventShoppingItemUpdated, events.EventShoppingItemPurchased, 
		 events.EventShoppingItemDeleted, events.EventShoppingListShared:
		return "house-helper.shopping"
	
	case events.EventBillCreated, events.EventBillUpdated, events.EventBillPaid, 
		 events.EventBillOverdue, events.EventBillDeleted:
		return "house-helper.bills"
	
	case events.EventTimerStarted, events.EventTimerPaused, events.EventTimerResumed, 
		 events.EventTimerCompleted, events.EventTimerStopped:
		return "house-helper.timers"
	
	case events.EventLaundryStarted, events.EventLaundryWashComplete, events.EventLaundryDryStarted,
		 events.EventLaundryDryComplete, events.EventLaundryCompleted:
		return "house-helper.laundry"
	
	case events.EventHouseholdCreated, events.EventHouseholdUpdated, events.EventHouseholdMemberAdded,
		 events.EventHouseholdMemberRemoved, events.EventHouseholdActivity:
		return "house-helper.households"
	
	case events.EventUserRegistered, events.EventUserLoggedIn, events.EventUserUpdated, events.EventUserDeleted:
		return "house-helper.users"
	
	case events.EventNotificationSent, events.EventNotificationClicked, events.EventNotificationFailed:
		return "house-helper.notifications"
	
	default:
		return "house-helper.misc"
	}
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// PublishBatch publishes multiple events in a batch
func (p *Producer) PublishBatch(events []*events.Event) error {
	messages := make([]*sarama.ProducerMessage, 0, len(events))

	for _, event := range events {
		topic := p.getTopicForEvent(event.Type)
		eventJSON, err := json.Marshal(event)
		if err != nil {
			p.logger.Error("Failed to serialize event", zap.String("eventId", event.ID), zap.Error(err))
			continue
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(event.ID),
			Value: sarama.ByteEncoder(eventJSON),
		}

		if event.HouseholdID != "" {
			msg.Key = sarama.StringEncoder(event.HouseholdID)
		}

		messages = append(messages, msg)
	}

	return p.producer.SendMessages(messages)
}
