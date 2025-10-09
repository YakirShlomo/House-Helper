package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
)

type TaskStore interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	GetUserTasks(ctx context.Context, userID string, filter TaskFilter) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	MarkComplete(ctx context.Context, id string, completedBy string) error
	MarkIncomplete(ctx context.Context, id string) error
	GetTasksByStatus(ctx context.Context, userID string, status models.TaskStatus) ([]*models.Task, error)
	GetTasksByCategory(ctx context.Context, userID string, category string) ([]*models.Task, error)
	GetOverdueTasks(ctx context.Context, userID string) ([]*models.Task, error)
}

type TaskFilter struct {
	Status     *models.TaskStatus `json:"status,omitempty"`
	Category   *string            `json:"category,omitempty"`
	Priority   *models.Priority   `json:"priority,omitempty"`
	AssignedTo *string            `json:"assignedTo,omitempty"`
	DueAfter   *time.Time         `json:"dueAfter,omitempty"`
	DueBefore  *time.Time         `json:"dueBefore,omitempty"`
	Search     *string            `json:"search,omitempty"`
	Limit      int                `json:"limit,omitempty"`
	Offset     int                `json:"offset,omitempty"`
}

type taskStore struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) TaskStore {
	return &taskStore{db: db}
}

func (s *taskStore) Create(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (
			id, title, description, category, priority, status,
			assigned_to, created_by, household_id, due_date, recurrence_rule,
			estimated_duration, actual_duration, attachment_urls,
			created_at, updated_at
		) VALUES (
			:id, :title, :description, :category, :priority, :status,
			:assigned_to, :created_by, :household_id, :due_date, :recurrence_rule,
			:estimated_duration, :actual_duration, :attachment_urls,
			:created_at, :updated_at
		)
	`

	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (s *taskStore) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT 
			id, title, description, category, priority, status,
			assigned_to, created_by, household_id, due_date, recurrence_rule,
			estimated_duration, actual_duration, attachment_urls,
			completed_at, created_at, updated_at
		FROM tasks 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var task models.Task
	err := s.db.GetContext(ctx, &task, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (s *taskStore) GetUserTasks(ctx context.Context, userID string, filter TaskFilter) ([]*models.Task, error) {
	query := `
		SELECT 
			id, title, description, category, priority, status,
			assigned_to, created_by, household_id, due_date, recurrence_rule,
			estimated_duration, actual_duration, attachment_urls,
			completed_at, created_at, updated_at
		FROM tasks t
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

	if filter.Priority != nil {
		argCount++
		query += fmt.Sprintf(" AND t.priority = $%d", argCount)
		args = append(args, *filter.Priority)
	}

	if filter.AssignedTo != nil {
		argCount++
		query += fmt.Sprintf(" AND t.assigned_to = $%d", argCount)
		args = append(args, *filter.AssignedTo)
	}

	if filter.DueAfter != nil {
		argCount++
		query += fmt.Sprintf(" AND t.due_date >= $%d", argCount)
		args = append(args, *filter.DueAfter)
	}

	if filter.DueBefore != nil {
		argCount++
		query += fmt.Sprintf(" AND t.due_date <= $%d", argCount)
		args = append(args, *filter.DueBefore)
	}

	if filter.Search != nil {
		argCount++
		query += fmt.Sprintf(" AND (t.title ILIKE $%d OR t.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+*filter.Search+"%")
	}

	// Order by priority and due date
	query += " ORDER BY t.priority DESC, t.due_date ASC"

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

	var tasks []*models.Task
	err := s.db.SelectContext(ctx, &tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	return tasks, nil
}

func (s *taskStore) Update(ctx context.Context, task *models.Task) error {
	query := `
		UPDATE tasks SET
			title = :title,
			description = :description,
			category = :category,
			priority = :priority,
			status = :status,
			assigned_to = :assigned_to,
			due_date = :due_date,
			recurrence_rule = :recurrence_rule,
			estimated_duration = :estimated_duration,
			actual_duration = :actual_duration,
			attachment_urls = :attachment_urls,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	task.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
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

func (s *taskStore) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE tasks 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
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

func (s *taskStore) MarkComplete(ctx context.Context, id string, completedBy string) error {
	query := `
		UPDATE tasks 
		SET 
			status = $1,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TaskStatusCompleted, id)
	if err != nil {
		return fmt.Errorf("failed to mark task complete: %w", err)
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

func (s *taskStore) MarkIncomplete(ctx context.Context, id string) error {
	query := `
		UPDATE tasks 
		SET 
			status = $1,
			completed_at = NULL,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.TaskStatusPending, id)
	if err != nil {
		return fmt.Errorf("failed to mark task incomplete: %w", err)
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

func (s *taskStore) GetTasksByStatus(ctx context.Context, userID string, status models.TaskStatus) ([]*models.Task, error) {
	filter := TaskFilter{
		Status: &status,
		Limit:  100,
	}
	return s.GetUserTasks(ctx, userID, filter)
}

func (s *taskStore) GetTasksByCategory(ctx context.Context, userID string, category string) ([]*models.Task, error) {
	filter := TaskFilter{
		Category: &category,
		Limit:    100,
	}
	return s.GetUserTasks(ctx, userID, filter)
}

func (s *taskStore) GetOverdueTasks(ctx context.Context, userID string) ([]*models.Task, error) {
	now := time.Now()
	pendingStatus := models.TaskStatusPending
	filter := TaskFilter{
		Status:    &pendingStatus,
		DueBefore: &now,
		Limit:     100,
	}
	return s.GetUserTasks(ctx, userID, filter)
}
