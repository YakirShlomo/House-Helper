package events

import (
	"encoding/json"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Task events
	EventTaskCreated   EventType = "task.created"
	EventTaskUpdated   EventType = "task.updated"
	EventTaskCompleted EventType = "task.completed"
	EventTaskDeleted   EventType = "task.deleted"

	// Shopping events
	EventShoppingItemAdded    EventType = "shopping.item.added"
	EventShoppingItemUpdated  EventType = "shopping.item.updated"
	EventShoppingItemPurchased EventType = "shopping.item.purchased"
	EventShoppingItemDeleted  EventType = "shopping.item.deleted"
	EventShoppingListShared   EventType = "shopping.list.shared"

	// Bill events
	EventBillCreated  EventType = "bill.created"
	EventBillUpdated  EventType = "bill.updated"
	EventBillPaid     EventType = "bill.paid"
	EventBillOverdue  EventType = "bill.overdue"
	EventBillDeleted  EventType = "bill.deleted"

	// Timer events
	EventTimerStarted   EventType = "timer.started"
	EventTimerPaused    EventType = "timer.paused"
	EventTimerResumed   EventType = "timer.resumed"
	EventTimerCompleted EventType = "timer.completed"
	EventTimerStopped   EventType = "timer.stopped"

	// Laundry events
	EventLaundryStarted      EventType = "laundry.started"
	EventLaundryWashComplete EventType = "laundry.wash.complete"
	EventLaundryDryStarted   EventType = "laundry.dry.started"
	EventLaundryDryComplete  EventType = "laundry.dry.complete"
	EventLaundryCompleted    EventType = "laundry.completed"

	// Household events
	EventHouseholdCreated      EventType = "household.created"
	EventHouseholdUpdated      EventType = "household.updated"
	EventHouseholdMemberAdded  EventType = "household.member.added"
	EventHouseholdMemberRemoved EventType = "household.member.removed"
	EventHouseholdActivity     EventType = "household.activity"

	// User events
	EventUserRegistered EventType = "user.registered"
	EventUserLoggedIn   EventType = "user.logged.in"
	EventUserUpdated    EventType = "user.updated"
	EventUserDeleted    EventType = "user.deleted"

	// Notification events
	EventNotificationSent    EventType = "notification.sent"
	EventNotificationClicked EventType = "notification.clicked"
	EventNotificationFailed  EventType = "notification.failed"
)

// Event represents a domain event in the system
type Event struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	Source      string                 `json:"source"`
	HouseholdID string                 `json:"householdId,omitempty"`
	UserID      string                 `json:"userId,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// NewEvent creates a new event with default values
func NewEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		Data:      data,
		Metadata:  make(map[string]string),
	}
}

// ToJSON converts the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON parses an event from JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// AddMetadata adds metadata to the event
func (e *Event) AddMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
}

// GetMetadata retrieves metadata from the event
func (e *Event) GetMetadata(key string) (string, bool) {
	if e.Metadata == nil {
		return "", false
	}
	value, exists := e.Metadata[key]
	return value, exists
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// In production, use a proper UUID library
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// TaskEventData represents data for task events
type TaskEventData struct {
	TaskID      string    `json:"taskId"`
	HouseholdID string    `json:"householdId"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	AssignedTo  string    `json:"assignedTo,omitempty"`
	CompletedBy string    `json:"completedBy,omitempty"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
	DueDate     time.Time `json:"dueDate,omitempty"`
}

// ShoppingEventData represents data for shopping events
type ShoppingEventData struct {
	ItemID      string    `json:"itemId"`
	ListID      string    `json:"listId"`
	HouseholdID string    `json:"householdId"`
	Name        string    `json:"name"`
	Quantity    int       `json:"quantity"`
	Category    string    `json:"category,omitempty"`
	Status      string    `json:"status"`
	PurchasedBy string    `json:"purchasedBy,omitempty"`
	PurchasedAt time.Time `json:"purchasedAt,omitempty"`
}

// BillEventData represents data for bill events
type BillEventData struct {
	BillID      string    `json:"billId"`
	HouseholdID string    `json:"householdId"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"dueDate"`
	PaidBy      string    `json:"paidBy,omitempty"`
	PaidAt      time.Time `json:"paidAt,omitempty"`
}

// TimerEventData represents data for timer events
type TimerEventData struct {
	TimerID     string        `json:"timerId"`
	UserID      string        `json:"userId"`
	HouseholdID string        `json:"householdId"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Duration    time.Duration `json:"duration"`
	ElapsedTime time.Duration `json:"elapsedTime,omitempty"`
	Status      string        `json:"status"`
}

// LaundryEventData represents data for laundry events
type LaundryEventData struct {
	LaundryID   string        `json:"laundryId"`
	UserID      string        `json:"userId"`
	HouseholdID string        `json:"householdId"`
	LoadType    string        `json:"loadType"`
	Status      string        `json:"status"`
	WashTime    time.Duration `json:"washTime,omitempty"`
	DryTime     time.Duration `json:"dryTime,omitempty"`
}

// HouseholdEventData represents data for household events
type HouseholdEventData struct {
	HouseholdID string `json:"householdId"`
	Name        string `json:"name"`
	MemberID    string `json:"memberId,omitempty"`
	Activity    string `json:"activity,omitempty"`
	ActorID     string `json:"actorId,omitempty"`
}

// NotificationEventData represents data for notification events
type NotificationEventData struct {
	NotificationID string            `json:"notificationId"`
	UserID         string            `json:"userId"`
	HouseholdID    string            `json:"householdId,omitempty"`
	Type           string            `json:"type"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Data           map[string]string `json:"data,omitempty"`
	Status         string            `json:"status"`
	Error          string            `json:"error,omitempty"`
}
