package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/househelper/kafka/pkg/consumer"
	"github.com/househelper/kafka/pkg/eventlog"
	"github.com/househelper/kafka/pkg/events"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get Kafka brokers from environment
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	// Get consumer group ID
	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "house-helper-event-consumer"
	}

	// Topics to consume
	topics := []string{
		"house-helper.tasks",
		"house-helper.shopping",
		"house-helper.bills",
		"house-helper.timers",
		"house-helper.laundry",
		"house-helper.households",
		"house-helper.users",
		"house-helper.notifications",
	}

	// Create consumer
	cons, err := consumer.NewConsumer(consumer.Config{
		Brokers: brokers,
		GroupID: groupID,
		Topics:  topics,
		Logger:  logger,
	})
	if err != nil {
		logger.Fatal("Failed to create consumer", zap.Error(err))
	}
	defer cons.Close()

	// Create event log
	eventLog, err := eventlog.NewEventLog(eventlog.Config{
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("Failed to create event log", zap.Error(err))
	}
	defer eventLog.Close()

	// Register event handlers
	registerHandlers(cons, eventLog, logger)

	logger.Info("Starting event consumer",
		zap.Strings("brokers", brokers),
		zap.String("groupId", groupID),
		zap.Strings("topics", topics),
	)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Shutdown signal received, stopping consumer...")
		cancel()
	}()

	// Start consuming
	err = cons.Start(ctx)
	if err != nil && err != context.Canceled {
		logger.Fatal("Consumer error", zap.Error(err))
	}

	logger.Info("Consumer stopped")
}

// registerHandlers registers all event handlers
func registerHandlers(cons *consumer.Consumer, eventLog *eventlog.EventLog, logger *zap.Logger) {
	// Task event handlers
	cons.RegisterHandler(events.EventTaskCreated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTaskUpdated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTaskCompleted, createTaskCompletedHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTaskDeleted, createEventLogHandler(eventLog, logger))

	// Shopping event handlers
	cons.RegisterHandler(events.EventShoppingItemAdded, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventShoppingItemUpdated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventShoppingItemPurchased, createShoppingPurchasedHandler(eventLog, logger))
	cons.RegisterHandler(events.EventShoppingItemDeleted, createEventLogHandler(eventLog, logger))

	// Bill event handlers
	cons.RegisterHandler(events.EventBillCreated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventBillUpdated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventBillPaid, createBillPaidHandler(eventLog, logger))
	cons.RegisterHandler(events.EventBillOverdue, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventBillDeleted, createEventLogHandler(eventLog, logger))

	// Timer event handlers
	cons.RegisterHandler(events.EventTimerStarted, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTimerPaused, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTimerResumed, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTimerCompleted, createTimerFinishedHandler(eventLog, logger))
	cons.RegisterHandler(events.EventTimerStopped, createEventLogHandler(eventLog, logger))

	// Laundry event handlers
	cons.RegisterHandler(events.EventLaundryStarted, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventLaundryWashComplete, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventLaundryDryStarted, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventLaundryDryComplete, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventLaundryCompleted, createEventLogHandler(eventLog, logger))

	// Household event handlers
	cons.RegisterHandler(events.EventHouseholdCreated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventHouseholdUpdated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventHouseholdMemberAdded, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventHouseholdMemberRemoved, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventHouseholdActivity, createHouseholdActivityHandler(eventLog, logger))

	// User event handlers
	cons.RegisterHandler(events.EventUserRegistered, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventUserLoggedIn, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventUserUpdated, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventUserDeleted, createEventLogHandler(eventLog, logger))

	// Notification event handlers
	cons.RegisterHandler(events.EventNotificationSent, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventNotificationClicked, createEventLogHandler(eventLog, logger))
	cons.RegisterHandler(events.EventNotificationFailed, createEventLogHandler(eventLog, logger))
}

// createEventLogHandler creates a handler that stores events in the event log
func createEventLogHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		logger.Debug("Storing event in log",
			zap.String("eventId", event.ID),
			zap.String("eventType", string(event.Type)),
		)
		return eventLog.Store(ctx, event)
	}
}

// createTaskCompletedHandler creates a specialized handler for task completed events
func createTaskCompletedHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		// Store in event log
		if err := eventLog.Store(ctx, event); err != nil {
			return err
		}

		// Additional processing for task completion
		taskID, _ := event.Data["taskId"].(string)
		householdID, _ := event.Data["householdId"].(string)
		completedBy, _ := event.Data["completedBy"].(string)

		logger.Info("Task completed",
			zap.String("taskId", taskID),
			zap.String("householdId", householdID),
			zap.String("completedBy", completedBy),
		)

		// TODO: Implement additional logic:
		// - Update task statistics
		// - Award points to user
		// - Send notifications to household members
		// - Update household activity feed

		return nil
	}
}

// createShoppingPurchasedHandler creates a specialized handler for shopping purchased events
func createShoppingPurchasedHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		// Store in event log
		if err := eventLog.Store(ctx, event); err != nil {
			return err
		}

		// Additional processing
		itemID, _ := event.Data["itemId"].(string)
		householdID, _ := event.Data["householdId"].(string)
		purchasedBy, _ := event.Data["purchasedBy"].(string)

		logger.Info("Shopping item purchased",
			zap.String("itemId", itemID),
			zap.String("householdId", householdID),
			zap.String("purchasedBy", purchasedBy),
		)

		// TODO: Implement additional logic:
		// - Update shopping list statistics
		// - Send notifications to list creator
		// - Check if all items are purchased
		// - Update household activity feed

		return nil
	}
}

// createBillPaidHandler creates a specialized handler for bill paid events
func createBillPaidHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		// Store in event log
		if err := eventLog.Store(ctx, event); err != nil {
			return err
		}

		// Additional processing
		billID, _ := event.Data["billId"].(string)
		householdID, _ := event.Data["householdId"].(string)
		paidBy, _ := event.Data["paidBy"].(string)
		amount, _ := event.Data["amount"].(float64)

		logger.Info("Bill paid",
			zap.String("billId", billID),
			zap.String("householdId", householdID),
			zap.String("paidBy", paidBy),
			zap.Float64("amount", amount),
		)

		// TODO: Implement additional logic:
		// - Update bill payment records
		// - Calculate bill splits
		// - Send payment confirmations
		// - Update household expenses

		return nil
	}
}

// createTimerFinishedHandler creates a specialized handler for timer finished events
func createTimerFinishedHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		// Store in event log
		if err := eventLog.Store(ctx, event); err != nil {
			return err
		}

		// Additional processing
		timerID, _ := event.Data["timerId"].(string)
		userID, _ := event.Data["userId"].(string)
		name, _ := event.Data["name"].(string)

		logger.Info("Timer finished",
			zap.String("timerId", timerID),
			zap.String("userId", userID),
			zap.String("name", name),
		)

		// TODO: Implement additional logic:
		// - Send completion notification
		// - Log timer completion
		// - Update user activity

		return nil
	}
}

// createHouseholdActivityHandler creates a specialized handler for household activity events
func createHouseholdActivityHandler(eventLog *eventlog.EventLog, logger *zap.Logger) consumer.Handler {
	return func(ctx context.Context, event *events.Event) error {
		// Store in event log
		if err := eventLog.Store(ctx, event); err != nil {
			return err
		}

		// Additional processing
		householdID, _ := event.Data["householdId"].(string)
		actorID, _ := event.Data["actorId"].(string)
		activity, _ := event.Data["activity"].(string)

		logger.Info("Household activity",
			zap.String("householdId", householdID),
			zap.String("actorId", actorID),
			zap.String("activity", activity),
		)

		// TODO: Implement additional logic:
		// - Update activity feed
		// - Send real-time updates to household members
		// - Update activity statistics

		return nil
	}
}
