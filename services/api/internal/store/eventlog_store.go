package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// EventLog represents an event in the system
type EventLog struct {
	ID          string                 `db:"id" json:"id"`
	EventType   string                 `db:"event_type" json:"event_type"`
	EntityType  string                 `db:"entity_type" json:"entity_type"`
	EntityID    string                 `db:"entity_id" json:"entity_id"`
	UserID      string                 `db:"user_id" json:"user_id"`
	HouseholdID string                 `db:"household_id" json:"household_id"`
	Payload     map[string]interface{} `db:"payload" json:"payload"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
}

type EventLogStore interface {
	Create(ctx context.Context, event *EventLog) error
	GetByID(ctx context.Context, id string) (*EventLog, error)
	GetByEntityID(ctx context.Context, entityID string) ([]*EventLog, error)
	GetByHouseholdID(ctx context.Context, householdID string, limit int) ([]*EventLog, error)
	GetByUserID(ctx context.Context, userID string, limit int) ([]*EventLog, error)
}

type eventLogStore struct {
	db *sqlx.DB
}

func NewEventLogStore(db *sqlx.DB) EventLogStore {
	return &eventLogStore{db: db}
}

func (s *eventLogStore) Create(ctx context.Context, event *EventLog) error {
	query := `
		INSERT INTO event_log (
			id, event_type, entity_type, entity_id, user_id, household_id, payload, created_at
		) VALUES (
			:id, :event_type, :entity_type, :entity_id, :user_id, :household_id, :payload, :created_at
		)
	`
	event.CreatedAt = time.Now()
	_, err := s.db.NamedExecContext(ctx, query, event)
	if err != nil {
		return fmt.Errorf("failed to create event log: %w", err)
	}
	return nil
}

func (s *eventLogStore) GetByID(ctx context.Context, id string) (*EventLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, household_id, payload, created_at
		FROM event_log 
		WHERE id = $1
	`
	var event EventLog
	err := s.db.GetContext(ctx, &event, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get event log by ID: %w", err)
	}
	return &event, nil
}

func (s *eventLogStore) GetByEntityID(ctx context.Context, entityID string) ([]*EventLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, household_id, payload, created_at
		FROM event_log 
		WHERE entity_id = $1
		ORDER BY created_at DESC
	`
	var events []*EventLog
	err := s.db.SelectContext(ctx, &events, query, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event logs by entity ID: %w", err)
	}
	return events, nil
}

func (s *eventLogStore) GetByHouseholdID(ctx context.Context, householdID string, limit int) ([]*EventLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, household_id, payload, created_at
		FROM event_log 
		WHERE household_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var events []*EventLog
	err := s.db.SelectContext(ctx, &events, query, householdID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get event logs by household ID: %w", err)
	}
	return events, nil
}

func (s *eventLogStore) GetByUserID(ctx context.Context, userID string, limit int) ([]*EventLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, household_id, payload, created_at
		FROM event_log 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var events []*EventLog
	err := s.db.SelectContext(ctx, &events, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get event logs by user ID: %w", err)
	}
	return events, nil
}
