package services

import (
	"context"

	"github.com/yakirshlomo/house-helper/services/api/internal/store"
	"github.com/yakirshlomo/house-helper/services/api/pkg/kafka"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/yakirshlomo/house-helper/services/api/pkg/temporal"
)

// Services holds all service dependencies
type Services struct {
	Auth         *AuthService
	Task         *TaskService
	Shopping     *ShoppingService
	Bill         *BillService
	Timer        *TimerService
	Notification *NotificationService
}

// AuthService handles authentication and user management
type AuthService struct {
	userStore store.UserStore
	jwtSecret string
}

// NewAuthService creates a new auth service
func NewAuthService(userStore store.UserStore, jwtSecret string) *AuthService {
	return &AuthService{
		userStore: userStore,
		jwtSecret: jwtSecret,
	}
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	return s.userStore.GetByID(context.Background(), userID)
}

// TaskService handles task operations
type TaskService struct {
	taskStore     store.TaskStore
	eventLog      store.EventLogStore
	kafkaProducer *kafka.Producer
}

// NewTaskService creates a new task service
func NewTaskService(taskStore store.TaskStore, eventLog store.EventLogStore, kafkaProducer *kafka.Producer) *TaskService {
	return &TaskService{
		taskStore:     taskStore,
		eventLog:      eventLog,
		kafkaProducer: kafkaProducer,
	}
}

// ShoppingService handles shopping list operations
type ShoppingService struct {
	shoppingStore store.ShoppingStore
	eventLog      store.EventLogStore
	kafkaProducer *kafka.Producer
}

// NewShoppingService creates a new shopping service
func NewShoppingService(shoppingStore store.ShoppingStore, eventLog store.EventLogStore, kafkaProducer *kafka.Producer) *ShoppingService {
	return &ShoppingService{
		shoppingStore: shoppingStore,
		eventLog:      eventLog,
		kafkaProducer: kafkaProducer,
	}
}

// BillService handles bill operations
type BillService struct {
	billStore     store.BillStore
	eventLog      store.EventLogStore
	kafkaProducer *kafka.Producer
}

// NewBillService creates a new bill service
func NewBillService(billStore store.BillStore, eventLog store.EventLogStore, kafkaProducer *kafka.Producer) *BillService {
	return &BillService{
		billStore:     billStore,
		eventLog:      eventLog,
		kafkaProducer: kafkaProducer,
	}
}

// TimerService handles timer operations
type TimerService struct {
	timerStore     store.TimerStore
	temporalClient *temporal.Client
	eventLog       store.EventLogStore
}

// NewTimerService creates a new timer service
func NewTimerService(timerStore store.TimerStore, temporalClient *temporal.Client, eventLog store.EventLogStore) *TimerService {
	return &TimerService{
		timerStore:     timerStore,
		temporalClient: temporalClient,
		eventLog:       eventLog,
	}
}

// NotificationService handles notification operations
type NotificationService struct {
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{}
}
