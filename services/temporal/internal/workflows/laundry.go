package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// LaundryWorkflowParams represents the input for laundry workflow
type LaundryWorkflowParams struct {
	LaundryID   string          `json:"laundryId"`
	UserID      string          `json:"userId"`
	HouseholdID string          `json:"householdId"`
	LoadType    string          `json:"loadType"` // normal, delicate, heavy, quick
	WashTime    time.Duration   `json:"washTime"` // Wash cycle duration
	DryTime     time.Duration   `json:"dryTime"`  // Dry cycle duration
	Settings    LaundrySettings `json:"settings"`
}

// LaundrySettings contains laundry-specific configuration
type LaundrySettings struct {
	AutoStart        bool          `json:"autoStart"`
	NotifyOnStart    bool          `json:"notifyOnStart"`
	NotifyOnWashDone bool          `json:"notifyOnWashDone"`
	NotifyOnDryDone  bool          `json:"notifyOnDryDone"`
	NotifyReminders  bool          `json:"notifyReminders"`
	ReminderInterval time.Duration `json:"reminderInterval"`
	MaxReminders     int           `json:"maxReminders"`
	Temperature      string        `json:"temperature"` // cold, warm, hot
	SpinSpeed        string        `json:"spinSpeed"`   // low, medium, high
	DryLevel         string        `json:"dryLevel"`    // low, medium, high, extra
	FabricSoftener   bool          `json:"fabricSoftener"`
	ExtraRinse       bool          `json:"extraRinse"`
}

// LaundryState represents the current state of laundry
type LaundryState struct {
	Status        string    `json:"status"` // created, washing, wash_done, drying, dry_done, completed
	WashStarted   time.Time `json:"washStarted"`
	WashFinished  time.Time `json:"washFinished"`
	DryStarted    time.Time `json:"dryStarted"`
	DryFinished   time.Time `json:"dryFinished"`
	RemindersLeft int       `json:"remindersLeft"`
	LastReminder  time.Time `json:"lastReminder"`
}

// LaundryWorkflow implements a complete laundry cycle with wash and dry phases
func LaundryWorkflow(ctx workflow.Context, params LaundryWorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting laundry workflow", "laundryId", params.LaundryID, "loadType", params.LoadType)

	// Initialize laundry state
	state := LaundryState{
		Status:        "created",
		RemindersLeft: params.Settings.MaxReminders,
	}

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

	// Start laundry tracking
	err := workflow.ExecuteActivity(ctx, StartLaundryActivity, StartLaundryRequest{
		LaundryID:   params.LaundryID,
		UserID:      params.UserID,
		HouseholdID: params.HouseholdID,
		LoadType:    params.LoadType,
		Settings:    params.Settings,
	}).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start laundry tracking: %w", err)
	}

	// Phase 1: Washing
	err = runWashCycle(ctx, params, &state)
	if err != nil {
		return fmt.Errorf("wash cycle failed: %w", err)
	}

	// Phase 2: Drying (optional)
	if params.DryTime > 0 {
		err = runDryCycle(ctx, params, &state)
		if err != nil {
			return fmt.Errorf("dry cycle failed: %w", err)
		}
	}

	// Complete laundry workflow
	state.Status = "completed"
	err = workflow.ExecuteActivity(ctx, CompleteLaundryActivity, CompleteLaundryRequest{
		LaundryID: params.LaundryID,
		UserID:    params.UserID,
		WashTime:  state.WashFinished.Sub(state.WashStarted),
		DryTime:   state.DryFinished.Sub(state.DryStarted),
		TotalTime: time.Since(state.WashStarted),
		Status:    state.Status,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to complete laundry tracking", "error", err)
	}

	logger.Info("Laundry workflow completed", "laundryId", params.LaundryID)
	return nil
}

// runWashCycle handles the washing phase
func runWashCycle(ctx workflow.Context, params LaundryWorkflowParams, state *LaundryState) error {
	logger := workflow.GetLogger(ctx)

	state.Status = "washing"
	state.WashStarted = workflow.Now(ctx)

	// Send wash start notification
	if params.Settings.NotifyOnStart {
		err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       "Laundry Started",
			Body:        fmt.Sprintf("Wash cycle started for %s load", params.LoadType),
			Data: map[string]string{
				"laundryId": params.LaundryID,
				"type":      "wash_started",
				"loadType":  params.LoadType,
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send wash start notification", "error", err)
		}
	}

	// Wait for wash cycle to complete
	selector := workflow.NewSelector(ctx)
	washTimer := workflow.NewTimer(ctx, params.WashTime)

	// Listen for wash completion
	selector.AddFuture(washTimer, func(f workflow.Future) {
		state.Status = "wash_done"
		state.WashFinished = workflow.Now(ctx)
		logger.Info("Wash cycle completed", "laundryId", params.LaundryID, "duration", params.WashTime)
	})

	// Listen for manual completion signal
	washDoneChannel := workflow.GetSignalChannel(ctx, "wash_complete")
	selector.AddReceive(washDoneChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "wash_done"
			state.WashFinished = workflow.Now(ctx)
			logger.Info("Wash cycle manually completed", "laundryId", params.LaundryID)
		}
	})

	selector.Select(ctx)

	// Send wash completion notification
	if params.Settings.NotifyOnWashDone {
		err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       "Wash Cycle Complete",
			Body:        "Your laundry is ready to be moved to the dryer",
			Data: map[string]string{
				"laundryId": params.LaundryID,
				"type":      "wash_complete",
				"loadType":  params.LoadType,
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send wash completion notification", "error", err)
		}
	}

	// Start reminder timer if enabled
	if params.Settings.NotifyReminders && params.Settings.ReminderInterval > 0 {
		go runWashReminders(ctx, params, state)
	}

	return nil
}

// runDryCycle handles the drying phase
func runDryCycle(ctx workflow.Context, params LaundryWorkflowParams, state *LaundryState) error {
	logger := workflow.GetLogger(ctx)

	// Wait for signal to start drying (laundry moved to dryer)
	dryStartChannel := workflow.GetSignalChannel(ctx, "start_dry")

	selector := workflow.NewSelector(ctx)

	// Auto-start dry cycle after delay if enabled
	var autoStartTimer workflow.Future
	if params.Settings.AutoStart {
		autoStartTimer = workflow.NewTimer(ctx, 5*time.Minute) // 5 min auto-start delay
		selector.AddFuture(autoStartTimer, func(f workflow.Future) {
			logger.Info("Auto-starting dry cycle", "laundryId", params.LaundryID)
		})
	}

	// Wait for manual dry start signal
	selector.AddReceive(dryStartChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			logger.Info("Dry cycle manually started", "laundryId", params.LaundryID)
		}
	})

	selector.Select(ctx)

	// Start drying phase
	state.Status = "drying"
	state.DryStarted = workflow.Now(ctx)

	// Send dry start notification
	err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
		UserID:      params.UserID,
		HouseholdID: params.HouseholdID,
		Title:       "Dry Cycle Started",
		Body:        fmt.Sprintf("Dry cycle started for %s load", params.LoadType),
		Data: map[string]string{
			"laundryId": params.LaundryID,
			"type":      "dry_started",
			"loadType":  params.LoadType,
		},
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send dry start notification", "error", err)
	}

	// Wait for dry cycle to complete
	drySelector := workflow.NewSelector(ctx)
	dryTimer := workflow.NewTimer(ctx, params.DryTime)

	// Listen for dry completion
	drySelector.AddFuture(dryTimer, func(f workflow.Future) {
		state.Status = "dry_done"
		state.DryFinished = workflow.Now(ctx)
		logger.Info("Dry cycle completed", "laundryId", params.LaundryID, "duration", params.DryTime)
	})

	// Listen for manual dry completion
	dryDoneChannel := workflow.GetSignalChannel(ctx, "dry_complete")
	drySelector.AddReceive(dryDoneChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "dry_done"
			state.DryFinished = workflow.Now(ctx)
			logger.Info("Dry cycle manually completed", "laundryId", params.LaundryID)
		}
	})

	drySelector.Select(ctx)

	// Send dry completion notification
	if params.Settings.NotifyOnDryDone {
		err = workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       "Laundry Complete",
			Body:        "Your laundry is ready to be folded and put away",
			Data: map[string]string{
				"laundryId": params.LaundryID,
				"type":      "dry_complete",
				"loadType":  params.LoadType,
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send dry completion notification", "error", err)
		}
	}

	// Start dry completion reminders if enabled
	if params.Settings.NotifyReminders && params.Settings.ReminderInterval > 0 {
		go runDryReminders(ctx, params, state)
	}

	return nil
}

// runWashReminders sends periodic reminders to move laundry to dryer
func runWashReminders(ctx workflow.Context, params LaundryWorkflowParams, state *LaundryState) {
	logger := workflow.GetLogger(ctx)

	for state.RemindersLeft > 0 && state.Status == "wash_done" {
		// Wait for reminder interval
		reminderTimer := workflow.NewTimer(ctx, params.Settings.ReminderInterval)

		selector := workflow.NewSelector(ctx)
		selector.AddFuture(reminderTimer, func(f workflow.Future) {
			// Send reminder
			err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
				UserID:      params.UserID,
				HouseholdID: params.HouseholdID,
				Title:       "Laundry Reminder",
				Body:        "Don't forget to move your laundry to the dryer",
				Data: map[string]string{
					"laundryId": params.LaundryID,
					"type":      "wash_reminder",
					"loadType":  params.LoadType,
				},
			}).Get(ctx, nil)
			if err != nil {
				logger.Warn("Failed to send wash reminder", "error", err)
			}

			state.RemindersLeft--
			state.LastReminder = workflow.Now(ctx)
			logger.Info("Sent wash reminder", "laundryId", params.LaundryID, "remindersLeft", state.RemindersLeft)
		})

		// Listen for dry start signal to stop reminders
		dryStartChannel := workflow.GetSignalChannel(ctx, "start_dry")
		selector.AddReceive(dryStartChannel, func(c workflow.ReceiveChannel, more bool) {
			if more {
				// Stop reminders as dry cycle is starting
				return
			}
		})

		selector.Select(ctx)
	}
}

// runDryReminders sends periodic reminders to remove laundry from dryer
func runDryReminders(ctx workflow.Context, params LaundryWorkflowParams, state *LaundryState) {
	logger := workflow.GetLogger(ctx)

	remindersLeft := params.Settings.MaxReminders

	for remindersLeft > 0 && state.Status == "dry_done" {
		// Wait for reminder interval
		reminderTimer := workflow.NewTimer(ctx, params.Settings.ReminderInterval)

		selector := workflow.NewSelector(ctx)
		selector.AddFuture(reminderTimer, func(f workflow.Future) {
			// Send reminder
			err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
				UserID:      params.UserID,
				HouseholdID: params.HouseholdID,
				Title:       "Laundry Reminder",
				Body:        "Your laundry is ready to be removed from the dryer",
				Data: map[string]string{
					"laundryId": params.LaundryID,
					"type":      "dry_reminder",
					"loadType":  params.LoadType,
				},
			}).Get(ctx, nil)
			if err != nil {
				logger.Warn("Failed to send dry reminder", "error", err)
			}

			remindersLeft--
			logger.Info("Sent dry reminder", "laundryId", params.LaundryID, "remindersLeft", remindersLeft)
		})

		// Listen for completion signal to stop reminders
		completeChannel := workflow.GetSignalChannel(ctx, "laundry_collected")
		selector.AddReceive(completeChannel, func(c workflow.ReceiveChannel, more bool) {
			if more {
				// Stop reminders as laundry has been collected
				state.Status = "completed"
				return
			}
		})

		selector.Select(ctx)
	}
}
