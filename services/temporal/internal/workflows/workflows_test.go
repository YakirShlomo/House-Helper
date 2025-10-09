package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type TimerWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *TimerWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	
	// Register activities
	s.env.RegisterActivity(StartTimerActivity)
	s.env.RegisterActivity(CompleteTimerActivity)
	s.env.RegisterActivity(SendNotificationActivity)
}

func (s *TimerWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *TimerWorkflowTestSuite) TestCountdownTimer() {
	params := TimerWorkflowParams{
		TimerID:     "timer-001",
		UserID:      "user-001",
		HouseholdID: "household-001",
		Name:        "Cooking Timer",
		Type:        "countdown",
		Duration:    5 * time.Minute,
		Settings: TimerSettings{
			NotifyOnStart:  true,
			NotifyOnFinish: true,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(StartTimerActivity, mock.Anything, StartTimerRequest{
		TimerID: params.TimerID,
		UserID:  params.UserID,
		Name:    params.Name,
	}).Return(nil)

	s.env.OnActivity(SendNotificationActivity, mock.Anything, NotificationRequest{
		UserID:      params.UserID,
		HouseholdID: params.HouseholdID,
		Title:       "Timer Started",
		Body:        "Cooking Timer timer has started",
		Data: map[string]string{
			"timerId": params.TimerID,
			"type":    "timer_started",
		},
	}).Return(nil)

	s.env.OnActivity(CompleteTimerActivity, mock.Anything, mock.AnythingOfType("CompleteTimerRequest")).Return(nil)

	s.env.OnActivity(SendNotificationActivity, mock.Anything, NotificationRequest{
		UserID:      params.UserID,
		HouseholdID: params.HouseholdID,
		Title:       "Timer Completed",
		Body:        "Cooking Timer timer has finished",
		Data: map[string]string{
			"timerId": params.TimerID,
			"type":    "timer_completed",
		},
	}).Return(nil)

	s.env.ExecuteWorkflow(TimerWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *TimerWorkflowTestSuite) TestPomodoroTimer() {
	params := TimerWorkflowParams{
		TimerID:     "timer-002",
		UserID:      "user-001",
		HouseholdID: "household-001",
		Name:        "Study Session",
		Type:        "pomodoro",
		Duration:    25 * time.Minute, // This will be overridden by settings
		Settings: TimerSettings{
			WorkDuration:   25 * time.Minute,
			ShortBreak:     5 * time.Minute,
			LongBreak:      15 * time.Minute,
			BreakInterval:  4,
			Repetitions:    2, // Short test with 2 cycles
			NotifyOnStart:  true,
			NotifyOnFinish: true,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(StartTimerActivity, mock.Anything, mock.AnythingOfType("StartTimerRequest")).Return(nil)
	s.env.OnActivity(SendNotificationActivity, mock.Anything, mock.AnythingOfType("NotificationRequest")).Return(nil).Times(5) // Start + 2 work periods + 1 break + finish
	s.env.OnActivity(CompleteTimerActivity, mock.Anything, mock.AnythingOfType("CompleteTimerRequest")).Return(nil)

	s.env.ExecuteWorkflow(TimerWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *TimerWorkflowTestSuite) TestTimerPause() {
	params := TimerWorkflowParams{
		TimerID:     "timer-003",
		UserID:      "user-001",
		HouseholdID: "household-001",
		Name:        "Test Timer",
		Type:        "countdown",
		Duration:    10 * time.Minute,
		Settings: TimerSettings{
			NotifyOnPause: true,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(StartTimerActivity, mock.Anything, mock.AnythingOfType("StartTimerRequest")).Return(nil)
	s.env.OnActivity(CompleteTimerActivity, mock.Anything, mock.AnythingOfType("CompleteTimerRequest")).Return(nil)

	// Start workflow
	s.env.ExecuteWorkflow(TimerWorkflow, params)

	// Send pause signal after a short delay
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("pause_timer", nil)
	}, 2*time.Second)

	// Send resume signal after pause
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("resume_timer", nil)
	}, 4*time.Second)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestTimerWorkflowSuite(t *testing.T) {
	suite.Run(t, new(TimerWorkflowTestSuite))
}

// LaundryWorkflowTestSuite tests the laundry workflow
type LaundryWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *LaundryWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	
	// Register activities
	s.env.RegisterActivity(StartLaundryActivity)
	s.env.RegisterActivity(CompleteLaundryActivity)
	s.env.RegisterActivity(SendNotificationActivity)
}

func (s *LaundryWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *LaundryWorkflowTestSuite) TestCompleteLaundryCycle() {
	params := LaundryWorkflowParams{
		LaundryID:   "laundry-001",
		UserID:      "user-001",
		HouseholdID: "household-001",
		LoadType:    "normal",
		WashTime:    30 * time.Minute,
		DryTime:     45 * time.Minute,
		Settings: LaundrySettings{
			NotifyOnStart:    true,
			NotifyOnWashDone: true,
			NotifyOnDryDone:  true,
			AutoStart:        true,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(StartLaundryActivity, mock.Anything, mock.AnythingOfType("StartLaundryRequest")).Return(nil)
	s.env.OnActivity(SendNotificationActivity, mock.Anything, mock.AnythingOfType("NotificationRequest")).Return(nil).Times(3) // Wash start, wash done, dry done
	s.env.OnActivity(CompleteLaundryActivity, mock.Anything, mock.AnythingOfType("CompleteLaundryRequest")).Return(nil)

	s.env.ExecuteWorkflow(LaundryWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *LaundryWorkflowTestSuite) TestLaundryWithReminders() {
	params := LaundryWorkflowParams{
		LaundryID:   "laundry-002",
		UserID:      "user-001",
		HouseholdID: "household-001",
		LoadType:    "delicate",
		WashTime:    25 * time.Minute,
		DryTime:     40 * time.Minute,
		Settings: LaundrySettings{
			NotifyOnWashDone: true,
			NotifyOnDryDone:  true,
			NotifyReminders:  true,
			ReminderInterval: 10 * time.Minute,
			MaxReminders:     2,
		},
	}

	// Mock activity expectations - including reminders
	s.env.OnActivity(StartLaundryActivity, mock.Anything, mock.AnythingOfType("StartLaundryRequest")).Return(nil)
	s.env.OnActivity(SendNotificationActivity, mock.Anything, mock.AnythingOfType("NotificationRequest")).Return(nil).Times(4) // Wash done + 2 reminders + dry done
	s.env.OnActivity(CompleteLaundryActivity, mock.Anything, mock.AnythingOfType("CompleteLaundryRequest")).Return(nil)

	s.env.ExecuteWorkflow(LaundryWorkflow, params)

	// Simulate dry start signal after wash reminders
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("start_dry", nil)
	}, 35*time.Minute)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestLaundryWorkflowSuite(t *testing.T) {
	suite.Run(t, new(LaundryWorkflowTestSuite))
}

// RecurringTaskWorkflowTestSuite tests the recurring task workflow
type RecurringTaskWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *RecurringTaskWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	
	// Register activities and workflows
	s.env.RegisterActivity(CreateTaskOccurrenceActivity)
	s.env.RegisterActivity(CheckTaskCompletionActivity)
	s.env.RegisterActivity(SendNotificationActivity)
	s.env.RegisterWorkflow(TaskReminderWorkflow)
}

func (s *RecurringTaskWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *RecurringTaskWorkflowTestSuite) TestDailyRecurringTask() {
	startDate := time.Now().Truncate(24 * time.Hour)
	
	params := RecurringTaskWorkflowParams{
		TaskID:      "task-001",
		UserID:      "user-001",
		HouseholdID: "household-001",
		Name:        "Take out trash",
		Description: "Weekly trash collection",
		RecurrenceRule: RecurrenceRule{
			Type:           "daily",
			Interval:       1,
			StartDate:      startDate,
			MaxOccurrences: 3, // Test with limited occurrences
		},
		AssignedMembers: []string{"user-001", "user-002"},
		DueDuration:     1 * time.Hour,
		ReminderSettings: ReminderSettings{
			Enabled: false, // Disable reminders for simpler test
		},
		AutoAssign: true,
	}

	// Mock activity expectations
	s.env.OnActivity(CreateTaskOccurrenceActivity, mock.Anything, mock.AnythingOfType("CreateTaskOccurrenceRequest")).Return(nil).Times(3)

	s.env.ExecuteWorkflow(RecurringTaskWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *RecurringTaskWorkflowTestSuite) TestWeeklyRecurringTask() {
	startDate := time.Now().Truncate(24 * time.Hour)
	
	params := RecurringTaskWorkflowParams{
		TaskID:      "task-002",
		UserID:      "user-001",
		HouseholdID: "household-001",
		Name:        "Weekly cleaning",
		Description: "Deep clean the house",
		RecurrenceRule: RecurrenceRule{
			Type:           "weekly",
			Interval:       1,
			DaysOfWeek:     []int{1, 5}, // Monday and Friday
			StartDate:      startDate,
			MaxOccurrences: 2,
		},
		DueDuration: 2 * time.Hour,
		ReminderSettings: ReminderSettings{
			Enabled: false,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(CreateTaskOccurrenceActivity, mock.Anything, mock.AnythingOfType("CreateTaskOccurrenceRequest")).Return(nil).Times(2)

	s.env.ExecuteWorkflow(RecurringTaskWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestRecurringTaskWorkflowSuite(t *testing.T) {
	suite.Run(t, new(RecurringTaskWorkflowTestSuite))
}

// TaskReminderWorkflowTestSuite tests the task reminder workflow
type TaskReminderWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *TaskReminderWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	
	// Register activities
	s.env.RegisterActivity(CheckTaskCompletionActivity)
	s.env.RegisterActivity(SendNotificationActivity)
}

func (s *TaskReminderWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *TaskReminderWorkflowTestSuite) TestTaskReminders() {
	dueDate := time.Now().Add(2 * time.Hour)
	
	params := TaskReminderWorkflowParams{
		OccurrenceID: "occurrence-001",
		TaskID:       "task-001",
		UserID:       "user-001",
		HouseholdID:  "household-001",
		AssignedTo:   "user-002",
		DueDate:      dueDate,
		Name:         "Test Task",
		ReminderSettings: ReminderSettings{
			Enabled:          true,
			InitialDelay:     1 * time.Hour,
			ReminderInterval: 15 * time.Minute,
			MaxReminders:     3,
			EscalateAfter:    2,
		},
	}

	// Mock activity expectations
	s.env.OnActivity(CheckTaskCompletionActivity, mock.Anything, mock.AnythingOfType("CheckTaskCompletionRequest")).Return(false, nil).Times(3)
	s.env.OnActivity(SendNotificationActivity, mock.Anything, mock.AnythingOfType("NotificationRequest")).Return(nil).Times(3)

	s.env.ExecuteWorkflow(TaskReminderWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *TaskReminderWorkflowTestSuite) TestTaskCompletedEarly() {
	dueDate := time.Now().Add(2 * time.Hour)
	
	params := TaskReminderWorkflowParams{
		OccurrenceID: "occurrence-002",
		TaskID:       "task-002",
		UserID:       "user-001",
		HouseholdID:  "household-001",
		AssignedTo:   "user-002",
		DueDate:      dueDate,
		Name:         "Test Task 2",
		ReminderSettings: ReminderSettings{
			Enabled:          true,
			InitialDelay:     1 * time.Hour,
			ReminderInterval: 15 * time.Minute,
			MaxReminders:     3,
		},
	}

	// Mock task completion after first reminder
	s.env.OnActivity(CheckTaskCompletionActivity, mock.Anything, mock.AnythingOfType("CheckTaskCompletionRequest")).Return(false, nil).Once()
	s.env.OnActivity(SendNotificationActivity, mock.Anything, mock.AnythingOfType("NotificationRequest")).Return(nil).Once()
	s.env.OnActivity(CheckTaskCompletionActivity, mock.Anything, mock.AnythingOfType("CheckTaskCompletionRequest")).Return(true, nil).Once()

	s.env.ExecuteWorkflow(TaskReminderWorkflow, params)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestTaskReminderWorkflowSuite(t *testing.T) {
	suite.Run(t, new(TaskReminderWorkflowTestSuite))
}