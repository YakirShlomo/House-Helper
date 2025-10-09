package models

import (
	"time"
	"github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID                string     `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	Phone             *string    `json:"phone,omitempty" db:"phone"`
	PasswordHash      string     `json:"-" db:"password_hash"`
	FirstName         string     `json:"firstName" db:"first_name"`
	LastName          string     `json:"lastName" db:"last_name"`
	IsVerified        bool       `json:"isVerified" db:"is_verified"`
	EmailVerifiedAt   *time.Time `json:"emailVerifiedAt,omitempty" db:"email_verified_at"`
	PhoneVerifiedAt   *time.Time `json:"phoneVerifiedAt,omitempty" db:"phone_verified_at"`
	LastLoginAt       *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
}

// UserProfile represents user preferences and settings
type UserProfile struct {
	UserID               string `json:"userId" db:"user_id"`
	Timezone             string `json:"timezone" db:"timezone"`
	Language             string `json:"language" db:"language"`
	Theme                string `json:"theme" db:"theme"`
	AvatarURL            string `json:"avatarUrl,omitempty" db:"avatar_url"`
	DateFormat           string `json:"dateFormat" db:"date_format"`
	TimeFormat           string `json:"timeFormat" db:"time_format"`
	NotificationsEnabled bool   `json:"notificationsEnabled" db:"notifications_enabled"`
	EmailNotifications   bool   `json:"emailNotifications" db:"email_notifications"`
	PushNotifications    bool   `json:"pushNotifications" db:"push_notifications"`
	CreatedAt            time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time `json:"updatedAt" db:"updated_at"`
}

// Household represents a household that users can belong to
type Household struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	Timezone    string    `json:"timezone" db:"timezone"`
	Currency    string    `json:"currency" db:"currency"`
	CreatedBy   string    `json:"createdBy" db:"created_by"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// HouseholdRole represents the role of a user in a household
type HouseholdRole string

const (
	HouseholdRoleAdmin  HouseholdRole = "admin"
	HouseholdRoleMember HouseholdRole = "member"
	HouseholdRoleGuest  HouseholdRole = "guest"
)

// HouseholdMember represents a user's membership in a household
type HouseholdMember struct {
	ID          string         `json:"id" db:"id"`
	HouseholdID string         `json:"householdId" db:"household_id"`
	UserID      string         `json:"userId" db:"user_id"`
	Role        HouseholdRole  `json:"role" db:"role"`
	JoinedAt    time.Time      `json:"joinedAt" db:"joined_at"`
	LeftAt      *time.Time     `json:"leftAt,omitempty" db:"left_at"`
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time      `json:"updatedAt" db:"updated_at"`
	
	// User information (joined from users table)
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// InvitationStatus represents the status of a household invitation
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusDeclined InvitationStatus = "declined"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// HouseholdInvitation represents an invitation to join a household
type HouseholdInvitation struct {
	ID          string           `json:"id" db:"id"`
	HouseholdID string           `json:"householdId" db:"household_id"`
	InvitedBy   string           `json:"invitedBy" db:"invited_by"`
	InviteCode  string           `json:"inviteCode" db:"invite_code"`
	Email       string           `json:"email" db:"email"`
	Role        HouseholdRole    `json:"role" db:"role"`
	Status      InvitationStatus `json:"status" db:"status"`
	ExpiresAt   time.Time        `json:"expiresAt" db:"expires_at"`
	AcceptedBy  *string          `json:"acceptedBy,omitempty" db:"accepted_by"`
	AcceptedAt  *time.Time       `json:"acceptedAt,omitempty" db:"accepted_at"`
	CreatedAt   time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time        `json:"updatedAt" db:"updated_at"`
}

// Priority represents task priority levels
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task represents a household task
type Task struct {
	ID                string         `json:"id" db:"id"`
	Title             string         `json:"title" db:"title"`
	Description       string         `json:"description,omitempty" db:"description"`
	Category          string         `json:"category" db:"category"`
	Priority          Priority       `json:"priority" db:"priority"`
	Status            TaskStatus     `json:"status" db:"status"`
	AssignedTo        *string        `json:"assignedTo,omitempty" db:"assigned_to"`
	CreatedBy         string         `json:"createdBy" db:"created_by"`
	HouseholdID       string         `json:"householdId" db:"household_id"`
	DueDate           *time.Time     `json:"dueDate,omitempty" db:"due_date"`
	RecurrenceRule    *string        `json:"recurrenceRule,omitempty" db:"recurrence_rule"`
	EstimatedDuration *int           `json:"estimatedDuration,omitempty" db:"estimated_duration"` // in minutes
	ActualDuration    *int           `json:"actualDuration,omitempty" db:"actual_duration"`       // in minutes
	AttachmentURLs    pq.StringArray `json:"attachmentUrls,omitempty" db:"attachment_urls"`
	CompletedAt       *time.Time     `json:"completedAt,omitempty" db:"completed_at"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time      `json:"updatedAt" db:"updated_at"`
}

// ShoppingList represents a shopping list
type ShoppingList struct {
	ID                  string         `json:"id" db:"id"`
	Name                string         `json:"name" db:"name"`
	HouseholdID         string         `json:"householdId" db:"household_id"`
	CreatedBy           string         `json:"createdBy" db:"created_by"`
	SharedWith          pq.StringArray `json:"sharedWith,omitempty" db:"shared_with"`
	Settings            string         `json:"settings,omitempty" db:"settings"` // JSON string
	TotalEstimatedCost  *float64       `json:"totalEstimatedCost,omitempty" db:"total_estimated_cost"`
	CreatedAt           time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time      `json:"updatedAt" db:"updated_at"`
}

// ShoppingItem represents an item in a shopping list
type ShoppingItem struct {
	ID             string     `json:"id" db:"id"`
	ListID         string     `json:"listId" db:"list_id"`
	Name           string     `json:"name" db:"name"`
	Quantity       float64    `json:"quantity" db:"quantity"`
	Unit           string     `json:"unit,omitempty" db:"unit"`
	Category       string     `json:"category,omitempty" db:"category"`
	Notes          string     `json:"notes,omitempty" db:"notes"`
	EstimatedPrice *float64   `json:"estimatedPrice,omitempty" db:"estimated_price"`
	ActualPrice    *float64   `json:"actualPrice,omitempty" db:"actual_price"`
	Barcode        *string    `json:"barcode,omitempty" db:"barcode"`
	IsPurchased    bool       `json:"isPurchased" db:"is_purchased"`
	PurchasedBy    *string    `json:"purchasedBy,omitempty" db:"purchased_by"`
	PurchasedAt    *time.Time `json:"purchasedAt,omitempty" db:"purchased_at"`
	AddedBy        string     `json:"addedBy" db:"added_by"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

// Product represents a product in the product database
type Product struct {
	ID           string   `json:"id" db:"id"`
	Name         string   `json:"name" db:"name"`
	Category     string   `json:"category" db:"category"`
	TypicalPrice *float64 `json:"typicalPrice,omitempty" db:"typical_price"`
	Barcode      *string  `json:"barcode,omitempty" db:"barcode"`
	Brand        *string  `json:"brand,omitempty" db:"brand"`
	Description  *string  `json:"description,omitempty" db:"description"`
}

// Permission represents access permissions
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
	PermissionAdmin Permission = "admin"
)

// BillStatus represents the status of a bill
type BillStatus string

const (
	BillStatusPending  BillStatus = "pending"
	BillStatusPaid     BillStatus = "paid"
	BillStatusOverdue  BillStatus = "overdue"
	BillStatusCancelled BillStatus = "cancelled"
)

// Bill represents a household bill
type Bill struct {
	ID              string         `json:"id" db:"id"`
	Name            string         `json:"name" db:"name"`
	Description     string         `json:"description,omitempty" db:"description"`
	Category        string         `json:"category" db:"category"`
	Amount          float64        `json:"amount" db:"amount"`
	Currency        string         `json:"currency" db:"currency"`
	DueDate         time.Time      `json:"dueDate" db:"due_date"`
	IsRecurring     bool           `json:"isRecurring" db:"is_recurring"`
	RecurrenceRule  *string        `json:"recurrenceRule,omitempty" db:"recurrence_rule"`
	Status          BillStatus     `json:"status" db:"status"`
	HouseholdID     string         `json:"householdId" db:"household_id"`
	AssignedTo      *string        `json:"assignedTo,omitempty" db:"assigned_to"`
	CreatedBy       string         `json:"createdBy" db:"created_by"`
	ReminderDays    *int           `json:"reminderDays,omitempty" db:"reminder_days"`
	AutoPayEnabled  bool           `json:"autoPayEnabled" db:"auto_pay_enabled"`
	PaymentMethod   *string        `json:"paymentMethod,omitempty" db:"payment_method"`
	VendorInfo      *string        `json:"vendorInfo,omitempty" db:"vendor_info"` // JSON string
	AttachmentURLs  pq.StringArray `json:"attachmentUrls,omitempty" db:"attachment_urls"`
	PaidAt          *time.Time     `json:"paidAt,omitempty" db:"paid_at"`
	PaidBy          *string        `json:"paidBy,omitempty" db:"paid_by"`
	PaidAmount      *float64       `json:"paidAmount,omitempty" db:"paid_amount"`
	CreatedAt       time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time      `json:"updatedAt" db:"updated_at"`
}

// BillPayment represents a payment made for a bill
type BillPayment struct {
	ID            string    `json:"id" db:"id"`
	BillID        string    `json:"billId" db:"bill_id"`
	Amount        float64   `json:"amount" db:"amount"`
	PaymentMethod string    `json:"paymentMethod" db:"payment_method"`
	TransactionID *string   `json:"transactionId,omitempty" db:"transaction_id"`
	PaidBy        string    `json:"paidBy" db:"paid_by"`
	PaidAt        time.Time `json:"paidAt" db:"paid_at"`
	Notes         *string   `json:"notes,omitempty" db:"notes"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// TimerType represents the type of timer
type TimerType string

const (
	TimerTypeCountdown TimerType = "countdown"
	TimerTypeStopwatch TimerType = "stopwatch"
	TimerTypePomodoro  TimerType = "pomodoro"
)

// TimerStatus represents the status of a timer
type TimerStatus string

const (
	TimerStatusCreated   TimerStatus = "created"
	TimerStatusRunning   TimerStatus = "running"
	TimerStatusPaused    TimerStatus = "paused"
	TimerStatusCompleted TimerStatus = "completed"
	TimerStatusStopped   TimerStatus = "stopped"
)

// Timer represents a timer
type Timer struct {
	ID          string       `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description,omitempty" db:"description"`
	Category    string       `json:"category" db:"category"`
	Type        TimerType    `json:"type" db:"type"`
	Duration    *int         `json:"duration,omitempty" db:"duration"` // in seconds
	HouseholdID string       `json:"householdId" db:"household_id"`
	CreatedBy   string       `json:"createdBy" db:"created_by"`
	Status      TimerStatus  `json:"status" db:"status"`
	WorkflowID  *string      `json:"workflowId,omitempty" db:"workflow_id"` // Temporal workflow ID
	Settings    *string      `json:"settings,omitempty" db:"settings"`     // JSON string
	StartedAt   *time.Time   `json:"startedAt,omitempty" db:"started_at"`
	CompletedAt *time.Time   `json:"completedAt,omitempty" db:"completed_at"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time    `json:"updatedAt" db:"updated_at"`
}

// TimerSession represents a session of timer usage
type TimerSession struct {
	ID             string     `json:"id" db:"id"`
	TimerID        string     `json:"timerId" db:"timer_id"`
	StartedAt      time.Time  `json:"startedAt" db:"started_at"`
	EndedAt        *time.Time `json:"endedAt,omitempty" db:"ended_at"`
	Duration       *int       `json:"duration,omitempty" db:"duration"`        // in seconds
	PausedDuration *int       `json:"pausedDuration,omitempty" db:"paused_duration"` // in seconds
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}
