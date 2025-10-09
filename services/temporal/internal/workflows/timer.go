package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// TimerWorkflowParams represents the input parameters for timer workflows
type TimerWorkflowParams struct {
	TimerID     string        `json:"timerId"`
	UserID      string        `json:"userId"`
	HouseholdID string        `json:"householdId"`
	Name        string        `json:"name"`
	Type        string        `json:"type"` // countdown, stopwatch, pomodoro
	Duration    time.Duration `json:"duration"`
	Settings    TimerSettings `json:"settings"`
}

// TimerSettings contains timer-specific configuration
type TimerSettings struct {
	AutoStart       bool          `json:"autoStart"`
	NotifyOnStart   bool          `json:"notifyOnStart"`
	NotifyOnPause   bool          `json:"notifyOnPause"`
	NotifyOnFinish  bool          `json:"notifyOnFinish"`
	WorkDuration    time.Duration `json:"workDuration"`  // For Pomodoro
	ShortBreak      time.Duration `json:"shortBreak"`    // For Pomodoro
	LongBreak       time.Duration `json:"longBreak"`     // For Pomodoro
	BreakInterval   int           `json:"breakInterval"` // For Pomodoro
	Repetitions     int           `json:"repetitions"`   // Number of cycles
	NotificationMsg string        `json:"notificationMsg"`
}

// TimerState represents the current state of a timer
type TimerState struct {
	Status          string        `json:"status"`
	ElapsedTime     time.Duration `json:"elapsedTime"`
	RemainingTime   time.Duration `json:"remainingTime"`
	CurrentCycle    int           `json:"currentCycle"`
	IsBreak         bool          `json:"isBreak"`
	PausedTime      time.Duration `json:"pausedTime"`
	LastPauseStart  time.Time     `json:"lastPauseStart"`
	CompletedCycles int           `json:"completedCycles"`
}

// TimerWorkflow implements a durable timer with pause/resume capabilities
func TimerWorkflow(ctx workflow.Context, params TimerWorkflowParams) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting timer workflow", "timerId", params.TimerID, "type", params.Type)

	// Initialize timer state
	state := TimerState{
		Status:        "created",
		ElapsedTime:   0,
		RemainingTime: params.Duration,
		CurrentCycle:  1,
		IsBreak:       false,
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

	// Start timer activity
	err := workflow.ExecuteActivity(ctx, StartTimerActivity, StartTimerRequest{
		TimerID: params.TimerID,
		UserID:  params.UserID,
		Name:    params.Name,
	}).Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start timer: %w", err)
	}

	state.Status = "running"

	// Send start notification if enabled
	if params.Settings.NotifyOnStart {
		err = workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       "Timer Started",
			Body:        fmt.Sprintf("%s timer has started", params.Name),
			Data: map[string]string{
				"timerId": params.TimerID,
				"type":    "timer_started",
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send start notification", "error", err)
		}
	}

	// Handle different timer types
	switch params.Type {
	case "countdown":
		err = runCountdownTimer(ctx, params, &state)
	case "pomodoro":
		err = runPomodoroTimer(ctx, params, &state)
	case "stopwatch":
		err = runStopwatchTimer(ctx, params, &state)
	default:
		return fmt.Errorf("unsupported timer type: %s", params.Type)
	}

	if err != nil {
		return fmt.Errorf("timer execution failed: %w", err)
	}

	// Complete timer
	state.Status = "completed"
	err = workflow.ExecuteActivity(ctx, CompleteTimerActivity, CompleteTimerRequest{
		TimerID:     params.TimerID,
		UserID:      params.UserID,
		ElapsedTime: state.ElapsedTime,
		Status:      state.Status,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to complete timer", "error", err)
	}

	// Send completion notification
	if params.Settings.NotifyOnFinish {
		err = workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       "Timer Completed",
			Body:        fmt.Sprintf("%s timer has finished", params.Name),
			Data: map[string]string{
				"timerId": params.TimerID,
				"type":    "timer_completed",
			},
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send completion notification", "error", err)
		}
	}

	logger.Info("Timer workflow completed", "timerId", params.TimerID)
	return nil
}

// runCountdownTimer implements countdown timer logic
func runCountdownTimer(ctx workflow.Context, params TimerWorkflowParams, state *TimerState) error {
	logger := workflow.GetLogger(ctx)

	selector := workflow.NewSelector(ctx)
	timerFuture := workflow.NewTimer(ctx, params.Duration)

	// Listen for timer completion
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		// Timer completed naturally
		state.Status = "completed"
		state.RemainingTime = 0
		state.ElapsedTime = params.Duration
	})

	// Listen for pause/resume signals
	pauseChannel := workflow.GetSignalChannel(ctx, "pause_timer")
	resumeChannel := workflow.GetSignalChannel(ctx, "resume_timer")
	stopChannel := workflow.GetSignalChannel(ctx, "stop_timer")

	selector.AddReceive(pauseChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "paused"
			state.LastPauseStart = workflow.Now(ctx)
			logger.Info("Timer paused", "timerId", params.TimerID)
		}
	})

	selector.AddReceive(resumeChannel, func(c workflow.ReceiveChannel, more bool) {
		if more && state.Status == "paused" {
			pauseDuration := workflow.Now(ctx).Sub(state.LastPauseStart)
			state.PausedTime += pauseDuration
			state.Status = "running"
			logger.Info("Timer resumed", "timerId", params.TimerID, "pausedFor", pauseDuration)
		}
	})

	selector.AddReceive(stopChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "stopped"
			logger.Info("Timer stopped", "timerId", params.TimerID)
		}
	})

	// Wait for timer completion or signal
	selector.Select(ctx)

	return nil
}

// runPomodoroTimer implements Pomodoro technique timer
func runPomodoroTimer(ctx workflow.Context, params TimerWorkflowParams, state *TimerState) error {
	logger := workflow.GetLogger(ctx)

	workDuration := params.Settings.WorkDuration
	if workDuration == 0 {
		workDuration = 25 * time.Minute // Default Pomodoro work period
	}

	shortBreak := params.Settings.ShortBreak
	if shortBreak == 0 {
		shortBreak = 5 * time.Minute // Default short break
	}

	longBreak := params.Settings.LongBreak
	if longBreak == 0 {
		longBreak = 15 * time.Minute // Default long break
	}

	breakInterval := params.Settings.BreakInterval
	if breakInterval == 0 {
		breakInterval = 4 // Long break every 4 cycles
	}

	maxCycles := params.Settings.Repetitions
	if maxCycles == 0 {
		maxCycles = 4 // Default 4 Pomodoro cycles
	}

	for state.CurrentCycle <= maxCycles {
		// Work period
		state.IsBreak = false
		logger.Info("Starting work period", "cycle", state.CurrentCycle)

		err := runTimerPeriod(ctx, params, state, workDuration, "Work time!")
		if err != nil || state.Status == "stopped" {
			return err
		}

		state.CompletedCycles++

		// Break period (except after last cycle)
		if state.CurrentCycle < maxCycles {
			state.IsBreak = true
			var breakDuration time.Duration
			var breakMsg string

			if state.CompletedCycles%breakInterval == 0 {
				breakDuration = longBreak
				breakMsg = "Long break time!"
			} else {
				breakDuration = shortBreak
				breakMsg = "Short break time!"
			}

			logger.Info("Starting break period", "cycle", state.CurrentCycle, "duration", breakDuration)

			err = runTimerPeriod(ctx, params, state, breakDuration, breakMsg)
			if err != nil || state.Status == "stopped" {
				return err
			}
		}

		state.CurrentCycle++
	}

	state.Status = "completed"
	return nil
}

// runStopwatchTimer implements stopwatch functionality
func runStopwatchTimer(ctx workflow.Context, params TimerWorkflowParams, state *TimerState) error {
	logger := workflow.GetLogger(ctx)

	// Stopwatch runs indefinitely until stopped
	stopChannel := workflow.GetSignalChannel(ctx, "stop_timer")
	pauseChannel := workflow.GetSignalChannel(ctx, "pause_timer")
	resumeChannel := workflow.GetSignalChannel(ctx, "resume_timer")

	selector := workflow.NewSelector(ctx)
	startTime := workflow.Now(ctx)

	selector.AddReceive(stopChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "stopped"
			state.ElapsedTime = workflow.Now(ctx).Sub(startTime) - state.PausedTime
			logger.Info("Stopwatch stopped", "timerId", params.TimerID, "elapsed", state.ElapsedTime)
		}
	})

	selector.AddReceive(pauseChannel, func(c workflow.ReceiveChannel, more bool) {
		if more && state.Status == "running" {
			state.Status = "paused"
			state.LastPauseStart = workflow.Now(ctx)
			logger.Info("Stopwatch paused", "timerId", params.TimerID)
		}
	})

	selector.AddReceive(resumeChannel, func(c workflow.ReceiveChannel, more bool) {
		if more && state.Status == "paused" {
			pauseDuration := workflow.Now(ctx).Sub(state.LastPauseStart)
			state.PausedTime += pauseDuration
			state.Status = "running"
			logger.Info("Stopwatch resumed", "timerId", params.TimerID)
		}
	})

	// Wait for stop signal
	for state.Status != "stopped" {
		selector.Select(ctx)
	}

	return nil
}

// runTimerPeriod runs a timer for a specific duration with pause/resume support
func runTimerPeriod(ctx workflow.Context, params TimerWorkflowParams, state *TimerState, duration time.Duration, message string) error {
	selector := workflow.NewSelector(ctx)
	timerFuture := workflow.NewTimer(ctx, duration)

	selector.AddFuture(timerFuture, func(f workflow.Future) {
		// Period completed
	})

	// Handle pause/resume during period
	pauseChannel := workflow.GetSignalChannel(ctx, "pause_timer")
	resumeChannel := workflow.GetSignalChannel(ctx, "resume_timer")
	stopChannel := workflow.GetSignalChannel(ctx, "stop_timer")

	selector.AddReceive(pauseChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "paused"
			state.LastPauseStart = workflow.Now(ctx)
		}
	})

	selector.AddReceive(resumeChannel, func(c workflow.ReceiveChannel, more bool) {
		if more && state.Status == "paused" {
			pauseDuration := workflow.Now(ctx).Sub(state.LastPauseStart)
			state.PausedTime += pauseDuration
			state.Status = "running"
		}
	})

	selector.AddReceive(stopChannel, func(c workflow.ReceiveChannel, more bool) {
		if more {
			state.Status = "stopped"
		}
	})

	selector.Select(ctx)

	// Send period completion notification
	if state.Status != "stopped" {
		err := workflow.ExecuteActivity(ctx, SendNotificationActivity, NotificationRequest{
			UserID:      params.UserID,
			HouseholdID: params.HouseholdID,
			Title:       params.Name,
			Body:        message,
			Data: map[string]string{
				"timerId": params.TimerID,
				"type":    "timer_period_complete",
				"cycle":   fmt.Sprintf("%d", state.CurrentCycle),
				"isBreak": fmt.Sprintf("%t", state.IsBreak),
			},
		}).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Warn("Failed to send period notification", "error", err)
		}
	}

	return nil
}
