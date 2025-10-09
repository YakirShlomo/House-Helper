package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type TimerStore interface {
	Create(ctx context.Context, timer *models.Timer) error
	GetByID(ctx context.Context, id string) (*models.Timer, error)
	GetUserTimers(ctx context.Context, userID string, filter TimerFilter) ([]*models.Timer, error)
	Update(ctx context.Context, timer *models.Timer) error
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Pause(ctx context.Context, id string) error
	Resume(ctx context.Context, id string) error
	Complete(ctx context.Context, id string) error
	GetActiveTimers(ctx context.Context, userID string) ([]*models.Timer, error)
	
	// Session operations
	AddSession(ctx context.Context, session *models.TimerSession) error
	GetSessions(ctx context.Context, timerID string) ([]*models.TimerSession, error)
	UpdateSession(ctx context.Context, session *models.TimerSession) error
}

type TimerFilter struct {
	Status   *models.TimerStatus `json:"status,omitempty"`
	Category *string             `json:"category,omitempty"`
	Type     *models.TimerType   `json:"type,omitempty"`
	Search   *string             `json:"search,omitempty"`
	Limit    int                 `json:"limit,omitempty"`
	Offset   int                 `json:"offset,omitempty"`
}

type timerStore struct {
	db *sqlx.DB
}

func NewTimerStore(db *sqlx.DB) TimerStore {
	return &timerStore{db: db}
}

func (s *timerStore) Create(ctx context.Context, timer *models.Timer) error {
	query := `
		INSERT INTO timers (
			id, name, description, category, type, duration,
			household_id, created_by, status, workflow_id,
			settings, created_at, updated_at
		) VALUES (
			:id, :name, :description, :category, :type, :duration,
			:household_id, :created_by, :status, :workflow_id,
			:settings, :created_at, :updated_at
		)
	`

	timer.CreatedAt = time.Now()
	timer.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, timer)
	if err != nil {
		return fmt.Errorf("failed to create timer: %w", err)
	}

	return nil
}

func (s *timerStore) GetByID(ctx context.Context, id string) (*models.Timer, error) {
	query := `
		SELECT 
			id, name, description, category, type, duration,
			household_id, created_by, status, workflow_id,
			settings, started_at, completed_at, created_at, updated_at
		FROM timers 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var timer models.Timer
	err := s.db.GetContext(ctx, &timer, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get timer: %w", err)
	}

	return &timer, nil
}

func (s *timerStore) GetUserTimers(ctx context.Context, userID string, filter TimerFilter) ([]*models.Timer, error) {
	query := `
		SELECT 
			t.id, t.name, t.description, t.category, t.type, t.duration,
			t.household_id, t.created_by, t.status, t.workflow_id,
			t.settings, t.started_at, t.completed_at, t.created_at, t.updated_at
		FROM timers t
		JOIN households h ON t.household_id = h.id
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND t.deleted_at IS NULL
	`

	args := []interface{}{userID}
	argCount := 1

	// Apply filters
	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND t.status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	if filter.Category != nil {
		argCount++
		query += fmt.Sprintf(" AND t.category = $%d", argCount)
		args = append(args, *filter.Category)
	}

	if filter.Type != nil {
		argCount++
		query += fmt.Sprintf(" AND t.type = $%d", argCount)
		args = append(args, *filter.Type)
	}

	if filter.Search != nil {
		argCount++
		query += fmt.Sprintf(" AND (t.name ILIKE $%d OR t.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+*filter.Search+"%")
	}

	// Order by created date
	query += " ORDER BY t.created_at DESC"

	// Apply pagination
	if filter.Limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	var timers []*models.Timer
	err := s.db.SelectContext(ctx, &timers, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user timers: %w", err)
	}

	return timers, nil
}

func (s *timerStore) Update(ctx context.Context, timer *models.Timer) error {
	query := `
		UPDATE timers SET
			name = :name,
			description = :description,
			category = :category,
			type = :type,
			duration = :duration,
			status = :status,
			workflow_id = :workflow_id,
			settings = :settings,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	timer.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, timer)
	if err != nil {
		return fmt.Errorf("failed to update timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Start(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET 
			status = $1,
			started_at = NOW(),
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TimerStatusRunning, id)
	if err != nil {
		return fmt.Errorf("failed to start timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Stop(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET 
			status = $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TimerStatusStopped, id)
	if err != nil {
		return fmt.Errorf("failed to stop timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Pause(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET 
			status = $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TimerStatusPaused, id)
	if err != nil {
		return fmt.Errorf("failed to pause timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Resume(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET 
			status = $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TimerStatusRunning, id)
	if err != nil {
		return fmt.Errorf("failed to resume timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) Complete(ctx context.Context, id string) error {
	query := `
		UPDATE timers 
		SET 
			status = $1,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TimerStatusCompleted, id)
	if err != nil {
		return fmt.Errorf("failed to complete timer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *timerStore) GetActiveTimers(ctx context.Context, userID string) ([]*models.Timer, error) {
	runningStatus := models.TimerStatusRunning
	filter := TimerFilter{
		Status: &runningStatus,
		Limit:  50,
	}
	return s.GetUserTimers(ctx, userID, filter)
}

// Session operations
func (s *timerStore) AddSession(ctx context.Context, session *models.TimerSession) error {
	query := `
		INSERT INTO timer_sessions (
			id, timer_id, started_at, ended_at, duration,
			paused_duration, notes, created_at, updated_at
		) VALUES (
			:id, :timer_id, :started_at, :ended_at, :duration,
			:paused_duration, :notes, :created_at, :updated_at
		)
	`

	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to add timer session: %w", err)
	}

	return nil
}

func (s *timerStore) GetSessions(ctx context.Context, timerID string) ([]*models.TimerSession, error) {
	query := `
		SELECT 
			id, timer_id, started_at, ended_at, duration,
			paused_duration, notes, created_at, updated_at
		FROM timer_sessions 
		WHERE timer_id = $1 AND deleted_at IS NULL
		ORDER BY started_at DESC
	`

	var sessions []*models.TimerSession
	err := s.db.SelectContext(ctx, &sessions, query, timerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timer sessions: %w", err)
	}

	return sessions, nil
}

func (s *timerStore) UpdateSession(ctx context.Context, session *models.TimerSession) error {
	query := `
		UPDATE timer_sessions SET
			ended_at = :ended_at,
			duration = :duration,
			paused_duration = :paused_duration,
			notes = :notes,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	session.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to update timer session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
