package workflows

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RecurringTaskWorkflowParams represents parameters for recurring task workflows
type RecurringTaskWorkflowParams struct {
	TaskID           string           `json:"taskId"`
	UserID           string           `json:"userId"`
	HouseholdID      string           `json:"householdId"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	RecurrenceRule   RecurrenceRule   `json:"recurrenceRule"`
	AssignedMembers  []string         `json:"assignedMembers"`
	DueDuration      time.Duration    `json:"dueDuration"`
	ReminderSettings ReminderSettings `json:"reminderSettings"`
	AutoAssign       bool             `json:"autoAssign"`
}

// RecurrenceRule defines how often a task repeats
type RecurrenceRule struct {
	Type           string     `json:"type"`       // daily, weekly, monthly, custom
	Interval       int        `json:"interval"`   // Every N days/weeks/months
	DaysOfWeek     []int      `json:"daysOfWeek"` // For weekly (0=Sunday, 1=Monday, etc.)
	DayOfMonth     int        `json:"dayOfMonth"` // For monthly
	StartDate      time.Time  `json:"startDate"`
	EndDate        *time.Time `json:"endDate,omitempty"`
	MaxOccurrences int        `json:"maxOccurrences,omitempty"`
}

// ReminderSettings defines reminder configuration
type ReminderSettings struct {
	Enabled          bool          `json:"enabled"`
	InitialDelay     time.Duration `json:"initialDelay"`
	ReminderInterval time.Duration `json:"reminderInterval"`
	MaxReminders     int           `json:"maxReminders"`
	EscalateAfter    int           `json:"escalateAfter"`
}

// TaskOccurrence represents a single occurrence of a recurring task
type TaskOccurrence struct {
	OccurrenceID string     `json:"occurrenceId"`
	DueDate      time.Time  `json:"dueDate"`
	AssignedTo   string     `json:"assignedTo"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"createdAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	CompletedBy  string     `json:"completedBy,omitempty"`
}

// RecurringTaskWorkflow manages recurring task creation and lifecycle
func RecurringTaskWorkflow(ctx workflow.Context, params RecurringTaskWorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting recurring task workflow", "taskId", params.TaskID, "name", params.Name)

	// Setup activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	occurrenceCount := 0
	nextDueDate := params.RecurrenceRule.StartDate

	// Continue until end conditions are met
	for {
		// Check if we've reached the end date or max occurrences
		if params.RecurrenceRule.EndDate != nil && nextDueDate.After(*params.RecurrenceRule.EndDate) {
			break
		}
		if params.RecurrenceRule.MaxOccurrences > 0 && occurrenceCount >= params.RecurrenceRule.MaxOccurrences {
			break
		}

		// Create task occurrence
		occurrence := TaskOccurrence{
			OccurrenceID: fmt.Sprintf("%s_%d", params.TaskID, occurrenceCount+1),
			DueDate:      nextDueDate,
			Status:       "pending",
			CreatedAt:    workflow.Now(ctx),
		}

		// Assign task to household member
		if params.AutoAssign && len(params.AssignedMembers) > 0 {
			// Round-robin assignment
			occurrence.AssignedTo = params.AssignedMembers[occurrenceCount%len(params.AssignedMembers)]
		}

		logger.Info("Creating task occurrence", "occurrenceId", occurrence.OccurrenceID, "dueDate", occurrence.DueDate)

		// Wait until it's time to create this occurrence
		currentTime := workflow.Now(ctx)
		createTime := occurrence.DueDate.Add(-params.DueDuration) // Create task X time before due date

		if createTime.After(currentTime) {
			timer := workflow.NewTimer(ctx, createTime.Sub(currentTime))
			err := timer.Get(ctx, nil)
			if err != nil {
				return fmt.Errorf("timer failed: %w", err)
			}
		}

		// Create the task occurrence
		err := workflow.ExecuteActivity(ctx, CreateTaskOccurrenceActivity, CreateTaskOccurrenceRequest{
			TaskID:      params.TaskID,
			Occurrence:  occurrence,
			Name:        params.Name,
			Description: params.Description,
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
		}).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to create task occurrence", "error", err)
			// Continue with next occurrence despite error
		} else {
			// Start reminder workflow for this occurrence
			if params.ReminderSettings.Enabled {
				childWorkflowOptions := workflow.ChildWorkflowOptions{
					WorkflowID: fmt.Sprintf("task-reminders-%s", occurrence.OccurrenceID),
				}
				childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

				reminderParams := TaskReminderWorkflowParams{
					OccurrenceID:     occurrence.OccurrenceID,
					TaskID:           params.TaskID,
					UserID:           params.UserID,
					HouseholdID:      params.HouseholdID,
					AssignedTo:       occurrence.AssignedTo,
					DueDate:          occurrence.DueDate,
					Name:             params.Name,
					ReminderSettings: params.ReminderSettings,
				}

				workflow.ExecuteChildWorkflow(childCtx, TaskReminderWorkflow, reminderParams)
				// Note: Not waiting for child workflow to complete as it runs independently
			}
		}

		// Calculate next due date
		nextDueDate = calculateNextDueDate(params.RecurrenceRule, nextDueDate)
		occurrenceCount++

		// Listen for workflow cancellation
		selector := workflow.NewSelector(ctx)
		cancelChannel := workflow.GetSignalChannel(ctx, "cancel_recurring_task")

		selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
			if more {
				logger.Info("Recurring task workflow cancelled", "taskId", params.TaskID)
				return
			}
		})

		// Add a small delay to prevent tight loops
		delayTimer := workflow.NewTimer(ctx, time.Second)
		selector.AddFuture(delayTimer, func(f workflow.Future) {
			// Continue to next iteration
		})

		selector.Select(ctx)

		// Check if cancellation was requested
		if workflow.GetSignalChannel(ctx, "cancel_recurring_task").ReceiveAsync(nil) {
			break
		}
	}

	logger.Info("Recurring task workflow completed", "taskId", params.TaskID, "occurrences", occurrenceCount)
	return nil
}

// calculateNextDueDate calculates the next due date based on recurrence rule
func calculateNextDueDate(rule RecurrenceRule, currentDate time.Time) time.Time {
	switch rule.Type {
	case "daily":
		return currentDate.AddDate(0, 0, rule.Interval)

	case "weekly":
		// Find next occurrence based on days of week
		if len(rule.DaysOfWeek) == 0 {
			return currentDate.AddDate(0, 0, 7*rule.Interval)
		}

		// Find next day of week in the list
		nextDate := currentDate.AddDate(0, 0, 1) // Start checking from tomorrow
		for {
			weekday := int(nextDate.Weekday())
			for _, dow := range rule.DaysOfWeek {
				if weekday == dow {
					return nextDate
				}
			}
			nextDate = nextDate.AddDate(0, 0, 1)
		}

	case "monthly":
		if rule.DayOfMonth > 0 {
			year, month, _ := currentDate.Date()
			nextMonth := month + time.Month(rule.Interval)
			nextYear := year
			if nextMonth > 12 {
				nextYear += int(nextMonth-1) / 12
				nextMonth = ((nextMonth - 1) % 12) + 1
			}

			// Handle day of month overflow (e.g., Feb 31 -> Feb 28/29)
			daysInMonth := time.Date(nextYear, nextMonth+1, 0, 0, 0, 0, 0, currentDate.Location()).Day()
			day := rule.DayOfMonth
			if day > daysInMonth {
				day = daysInMonth
			}

			return time.Date(nextYear, nextMonth, day, currentDate.Hour(), currentDate.Minute(), currentDate.Second(), currentDate.Nanosecond(), currentDate.Location())
		}
		return currentDate.AddDate(0, rule.Interval, 0)

	default:
		// Default to daily
		return currentDate.AddDate(0, 0, rule.Interval)
	}
}

// TaskReminderWorkflowParams represents parameters for task reminder workflow
type TaskReminderWorkflowParams struct {
	OccurrenceID     string           `json:"occurrenceId"`
	TaskID           string           `json:"taskId"`
	UserID           string           `json:"userId"`
	HouseholdID      string           `json:"householdId"`
	AssignedTo       string           `json:"assignedTo"`
	DueDate          time.Time        `json:"dueDate"`
	Name             string           `json:"name"`
	ReminderSettings ReminderSettings `json:"reminderSettings"`
}

// TaskReminderWorkflow handles sending reminders for a specific task occurrence
func TaskReminderWorkflow(ctx workflow.Context, params TaskReminderWorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting task reminder workflow", "occurrenceId", params.OccurrenceID)

	// Setup activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Wait for initial delay before first reminder
	firstReminderTime := params.DueDate.Add(-params.ReminderSettings.InitialDelay)
	currentTime := workflow.Now(ctx)

	if firstReminderTime.After(currentTime) {
		timer := workflow.NewTimer(ctx, firstReminderTime.Sub(currentTime))
		err := timer.Get(ctx, nil)
		if err != nil {
			return fmt.Errorf("initial reminder timer failed: %w", err)
		}
	}

	reminderCount := 0

	for reminderCount < params.ReminderSettings.MaxReminders {
		// Check if task is completed
		var isCompleted bool
		err := workflow.ExecuteActivity(ctx, CheckTaskCompletionActivity, CheckTaskCompletionRequest{
			OccurrenceID: params.OccurrenceID,
		}).Get(ctx, &isCompleted)
		if err != nil {
			logger.Warn("Failed to check task completion", "error", err)
		} else if isCompleted {
			logger.Info("Task completed, stopping reminders", "occurrenceId", params.OccurrenceID)
			break
		}

		// Send reminder
		reminderType := "reminder"
		if reminderCount >= params.ReminderSettings.EscalateAfter {
			reminderType = "escalated_reminder"
		}

		err = workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.AssignedTo,
			HouseholdID: params.HouseholdID,
			Title:       fmt.Sprintf("Task Reminder: %s", params.Name),
			Body:        fmt.Sprintf("Don't forget to complete your task: %s (Due: %s)", params.Name, params.DueDate.Format("Jan 2, 3:04 PM")),
			Data: map[string]string{
				"taskId":       params.TaskID,
				"occurrenceId": params.OccurrenceID,
				"type":         reminderType,
				"dueDate":      params.DueDate.Format(time.RFC3339),
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send reminder", "error", err)
		} else {
			logger.Info("Sent task reminder", "occurrenceId", params.OccurrenceID, "count", reminderCount+1)
		}

		reminderCount++

		// Wait for next reminder interval (unless this is the last reminder)
		if reminderCount < params.ReminderSettings.MaxReminders {
			timer := workflow.NewTimer(ctx, params.ReminderSettings.ReminderInterval)

			selector := workflow.NewSelector(ctx)
			selector.AddFuture(timer, func(f workflow.Future) {
				// Continue to next reminder
			})

			// Listen for task completion signal to stop reminders early
			completionChannel := workflow.GetSignalChannel(ctx, "task_completed")
			selector.AddReceive(completionChannel, func(c workflow.ReceiveChannel, more bool) {
				if more {
					logger.Info("Task completed signal received, stopping reminders", "occurrenceId", params.OccurrenceID)
					return
				}
			})

			selector.Select(ctx)

			// Check if completion signal was received
			if workflow.GetSignalChannel(ctx, "task_completed").ReceiveAsync(nil) {
				break
			}
		}
	}

	logger.Info("Task reminder workflow completed", "occurrenceId", params.OccurrenceID, "reminders", reminderCount)
	return nil
}

// CreateTaskOccurrenceRequest represents a request to create a task occurrence
type CreateTaskOccurrenceRequest struct {
	TaskID      string         `json:"taskId"`
	Occurrence  TaskOccurrence `json:"occurrence"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	UserID      string         `json:"userId"`
	HouseholdID string         `json:"householdId"`
}

// CheckTaskCompletionRequest represents a request to check if a task is completed
type CheckTaskCompletionRequest struct {
	OccurrenceID string `json:"occurrenceId"`
}

// CreateTaskOccurrenceActivity creates a new task occurrence in the database
func CreateTaskOccurrenceActivity(ctx context.Context, req CreateTaskOccurrenceRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating task occurrence", "occurrenceId", req.Occurrence.OccurrenceID)

	// Here you would typically:
	// 1. Insert task occurrence into database
	// 2. Send notifications to assigned user
	// 3. Update household task board
	// 4. Send real-time updates to connected clients

	activity.RecordHeartbeat(ctx, "Creating task occurrence record")

	logger.Info("Task occurrence created successfully", "occurrenceId", req.Occurrence.OccurrenceID)
	return nil
}

// CheckTaskCompletionActivity checks if a task occurrence is completed
func CheckTaskCompletionActivity(ctx context.Context, req CheckTaskCompletionRequest) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking task completion", "occurrenceId", req.OccurrenceID)

	// Here you would typically:
	// 1. Query database for task completion status
	// 2. Return true if completed, false otherwise

	activity.RecordHeartbeat(ctx, "Checking task status")

	// Mock implementation - in real scenario, this would query the database
	// For demo purposes, randomly return false (task not completed)
	logger.Info("Task completion checked", "occurrenceId", req.OccurrenceID)
	return false, nil
}
