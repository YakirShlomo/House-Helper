package temporal

import (
	"context"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Config holds Temporal configuration
type Config struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

// Client wraps the Temporal client
type Client struct {
	client.Client
	taskQueue string
}

// NewClient creates a new Temporal client
func NewClient(config Config) (*Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  config.HostPort,
		Namespace: config.Namespace,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:    c,
		taskQueue: config.TaskQueue,
	}, nil
}

// StartWorker starts a Temporal worker
func (c *Client) StartWorker(workflows []interface{}, activities []interface{}) error {
	w := worker.New(c.Client, c.taskQueue, worker.Options{})

	// Register workflows
	for _, workflow := range workflows {
		w.RegisterWorkflow(workflow)
	}

	// Register activities
	for _, activity := range activities {
		w.RegisterActivity(activity)
	}

	return w.Run(worker.InterruptCh())
}

// TimerWorkflow represents a timer workflow
func TimerWorkflow(ctx context.Context, timerID string, duration time.Duration) error {
	logger := log.Default()
	logger.Printf("Starting timer workflow for %s with duration %v", timerID, duration)

	// Wait for the specified duration
	if err := ctx.Err(); err != nil {
		return err
	}

	// Sleep for the duration
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		logger.Printf("Timer %s completed", timerID)
		return nil
	case <-ctx.Done():
		logger.Printf("Timer %s cancelled", timerID)
		return ctx.Err()
	}
}

// NotificationActivity sends notifications
func NotificationActivity(ctx context.Context, userID, message string) error {
	log.Printf("Sending notification to user %s: %s", userID, message)
	// TODO: Implement actual notification sending via FCM/APNs
	return nil
}
