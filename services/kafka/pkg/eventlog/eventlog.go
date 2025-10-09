package eventlog

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/househelper/kafka/pkg/events"
	"go.uber.org/zap"
)

// EventLog provides persistent storage for events
type EventLog struct {
	logger *zap.Logger
	// In production, this would use a database connection
	// db *sql.DB
}

// EventRecord represents a persisted event
type EventRecord struct {
	ID          string    `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"`
	Source      string    `json:"source" db:"source"`
	HouseholdID string    `json:"householdId,omitempty" db:"household_id"`
	UserID      string    `json:"userId,omitempty" db:"user_id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Version     string    `json:"version" db:"version"`
	Data        string    `json:"data" db:"data"` // JSON string
	Metadata    string    `json:"metadata,omitempty" db:"metadata"` // JSON string
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// Config holds event log configuration
type Config struct {
	Logger *zap.Logger
	// DatabaseURL string
}

// NewEventLog creates a new event log
func NewEventLog(cfg Config) (*EventLog, error) {
	// In production, initialize database connection here
	// db, err := sql.Open("postgres", cfg.DatabaseURL)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to connect to database: %w", err)
	// }

	return &EventLog{
		logger: cfg.Logger,
		// db: db,
	}, nil
}

// Store persists an event to the log
func (el *EventLog) Store(ctx context.Context, event *events.Event) error {
	// Serialize event data
	dataJSON, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to serialize event data: %w", err)
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to serialize event metadata: %w", err)
	}

	record := EventRecord{
		ID:          event.ID,
		Type:        string(event.Type),
		Source:      event.Source,
		HouseholdID: event.HouseholdID,
		UserID:      event.UserID,
		Timestamp:   event.Timestamp,
		Version:     event.Version,
		Data:        string(dataJSON),
		Metadata:    string(metadataJSON),
		CreatedAt:   time.Now().UTC(),
	}

	// In production, insert into database
	// query := `
	//     INSERT INTO event_log (id, type, source, household_id, user_id, timestamp, version, data, metadata, created_at)
	//     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	// `
	// _, err = el.db.ExecContext(ctx, query,
	//     record.ID, record.Type, record.Source, record.HouseholdID, record.UserID,
	//     record.Timestamp, record.Version, record.Data, record.Metadata, record.CreatedAt,
	// )

	el.logger.Debug("Event stored",
		zap.String("eventId", record.ID),
		zap.String("eventType", record.Type),
	)

	// Mock success
	_ = record
	return nil
}

// GetByID retrieves an event by ID
func (el *EventLog) GetByID(ctx context.Context, eventID string) (*events.Event, error) {
	// In production, query database
	// query := `
	//     SELECT id, type, source, household_id, user_id, timestamp, version, data, metadata
	//     FROM event_log
	//     WHERE id = $1
	// `
	// var record EventRecord
	// err := el.db.QueryRowContext(ctx, query, eventID).Scan(
	//     &record.ID, &record.Type, &record.Source, &record.HouseholdID, &record.UserID,
	//     &record.Timestamp, &record.Version, &record.Data, &record.Metadata,
	// )
	// if err != nil {
	//     return nil, fmt.Errorf("failed to retrieve event: %w", err)
	// }

	// return el.recordToEvent(&record)

	// Mock implementation
	return nil, fmt.Errorf("event not found: %s", eventID)
}

// GetByHousehold retrieves events for a household
func (el *EventLog) GetByHousehold(ctx context.Context, householdID string, limit int, offset int) ([]*events.Event, error) {
	// In production, query database
	// query := `
	//     SELECT id, type, source, household_id, user_id, timestamp, version, data, metadata
	//     FROM event_log
	//     WHERE household_id = $1
	//     ORDER BY timestamp DESC
	//     LIMIT $2 OFFSET $3
	// `
	// rows, err := el.db.QueryContext(ctx, query, householdID, limit, offset)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to query events: %w", err)
	// }
	// defer rows.Close()

	// var events []*events.Event
	// for rows.Next() {
	//     var record EventRecord
	//     err := rows.Scan(
	//         &record.ID, &record.Type, &record.Source, &record.HouseholdID, &record.UserID,
	//         &record.Timestamp, &record.Version, &record.Data, &record.Metadata,
	//     )
	//     if err != nil {
	//         return nil, fmt.Errorf("failed to scan event: %w", err)
	//     }
	//
	//     event, err := el.recordToEvent(&record)
	//     if err != nil {
	//         el.logger.Warn("Failed to convert record to event", zap.Error(err))
	//         continue
	//     }
	//     events = append(events, event)
	// }

	// Mock implementation
	return []*events.Event{}, nil
}

// GetByType retrieves events by type
func (el *EventLog) GetByType(ctx context.Context, eventType events.EventType, limit int, offset int) ([]*events.Event, error) {
	// In production, query database
	// query := `
	//     SELECT id, type, source, household_id, user_id, timestamp, version, data, metadata
	//     FROM event_log
	//     WHERE type = $1
	//     ORDER BY timestamp DESC
	//     LIMIT $2 OFFSET $3
	// `

	// Mock implementation
	return []*events.Event{}, nil
}

// GetByTimeRange retrieves events within a time range
func (el *EventLog) GetByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*events.Event, error) {
	// In production, query database
	// query := `
	//     SELECT id, type, source, household_id, user_id, timestamp, version, data, metadata
	//     FROM event_log
	//     WHERE timestamp >= $1 AND timestamp <= $2
	//     ORDER BY timestamp DESC
	//     LIMIT $3
	// `

	// Mock implementation
	return []*events.Event{}, nil
}

// Delete removes events older than a specified duration
func (el *EventLog) Delete(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().UTC().Add(-olderThan)

	// In production, delete from database
	// query := `
	//     DELETE FROM event_log
	//     WHERE created_at < $1
	// `
	// result, err := el.db.ExecContext(ctx, query, cutoffTime)
	// if err != nil {
	//     return 0, fmt.Errorf("failed to delete events: %w", err)
	// }
	//
	// rowsAffected, _ := result.RowsAffected()

	el.logger.Info("Deleted old events",
		zap.Time("cutoffTime", cutoffTime),
	)

	// Mock implementation
	return 0, nil
}

// recordToEvent converts a database record to an event
func (el *EventLog) recordToEvent(record *EventRecord) (*events.Event, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	var metadata map[string]string
	if record.Metadata != "" {
		if err := json.Unmarshal([]byte(record.Metadata), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event metadata: %w", err)
		}
	}

	return &events.Event{
		ID:          record.ID,
		Type:        events.EventType(record.Type),
		Source:      record.Source,
		HouseholdID: record.HouseholdID,
		UserID:      record.UserID,
		Timestamp:   record.Timestamp,
		Version:     record.Version,
		Data:        data,
		Metadata:    metadata,
	}, nil
}

// Close closes the event log connections
func (el *EventLog) Close() error {
	// In production, close database connection
	// return el.db.Close()
	return nil
}

// Migration SQL for event_log table
const EventLogTableMigration = `
CREATE TABLE IF NOT EXISTS event_log (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL,
    household_id VARCHAR(255),
    user_id VARCHAR(255),
    timestamp TIMESTAMPTZ NOT NULL,
    version VARCHAR(20) NOT NULL,
    data JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_event_log_household_id ON event_log(household_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_event_log_user_id ON event_log(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_event_log_type ON event_log(type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_event_log_timestamp ON event_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_event_log_created_at ON event_log(created_at);

-- Partial index for household events
CREATE INDEX IF NOT EXISTS idx_event_log_household_events ON event_log(household_id, type, timestamp DESC)
    WHERE household_id IS NOT NULL;
`
