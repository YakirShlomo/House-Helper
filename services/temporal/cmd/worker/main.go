package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/househelper/temporal/internal/workflows"
	tlog "github.com/househelper/temporal/pkg/log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

const (
	// TaskQueue is the default task queue name
	TaskQueue = "house-helper-tasks"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get Temporal server address from environment or use default
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = "localhost:7233"
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort:  temporalAddr,
		Namespace: getNamespace(),
		Logger:    tlog.NewZapAdapter(logger),
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", zap.Error(err))
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     10,
		MaxConcurrentWorkflowTaskExecutionSize: 10,
	})

	// Register workflows
	w.RegisterWorkflow(workflows.TimerWorkflow)
	w.RegisterWorkflow(workflows.LaundryWorkflow)
	w.RegisterWorkflow(workflows.RecurringTaskWorkflow)
	w.RegisterWorkflow(workflows.TaskReminderWorkflow)

	// Register activities
	w.RegisterActivity(workflows.StartTimerActivity)
	w.RegisterActivity(workflows.CompleteTimerActivity)
	w.RegisterActivity(workflows.StartLaundryActivity)
	w.RegisterActivity(workflows.CompleteLaundryActivity)
	w.RegisterActivity(workflows.SendNotificationActivity)
	w.RegisterActivity(workflows.UpdateTaskActivity)
	w.RegisterActivity(workflows.UpdateDeviceStateActivity)
	w.RegisterActivity(workflows.LogActivityActivity)
	w.RegisterActivity(workflows.SendWebhookActivity)
	w.RegisterActivity(workflows.CreateTaskOccurrenceActivity)
	w.RegisterActivity(workflows.CheckTaskCompletionActivity)

	logger.Info("Starting Temporal worker",
		zap.String("namespace", getNamespace()),
		zap.String("taskQueue", TaskQueue),
		zap.String("temporalAddress", temporalAddr),
	)

	// Start worker in goroutine
	go func() {
		err = w.Run(worker.InterruptCh())
		if err != nil {
			logger.Fatal("Worker run failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down worker...")
	w.Stop()
	logger.Info("Worker stopped")
}

// getNamespace returns the Temporal namespace from environment or default
func getNamespace() string {
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	return namespace
}
