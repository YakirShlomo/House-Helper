package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type BillStore interface {
	Create(ctx context.Context, bill *models.Bill) error
	GetByID(ctx context.Context, id string) (*models.Bill, error)
	GetUserBills(ctx context.Context, userID string, filter BillFilter) ([]*models.Bill, error)
	Update(ctx context.Context, bill *models.Bill) error
	Delete(ctx context.Context, id string) error
	MarkPaid(ctx context.Context, id string, paidBy string, amount float64) error
	MarkUnpaid(ctx context.Context, id string) error
	GetUpcomingBills(ctx context.Context, userID string, days int) ([]*models.Bill, error)
	GetOverdueBills(ctx context.Context, userID string) ([]*models.Bill, error)
	
	// Payment operations
	AddPayment(ctx context.Context, payment *models.BillPayment) error
	GetPayments(ctx context.Context, billID string) ([]*models.BillPayment, error)
	DeletePayment(ctx context.Context, paymentID string) error
}

type BillFilter struct {
	Status     *models.BillStatus `json:"status,omitempty"`
	Category   *string            `json:"category,omitempty"`
	DueAfter   *time.Time         `json:"dueAfter,omitempty"`
	DueBefore  *time.Time         `json:"dueBefore,omitempty"`
	Search     *string            `json:"search,omitempty"`
	Limit      int                `json:"limit,omitempty"`
	Offset     int                `json:"offset,omitempty"`
}

type billStore struct {
	db *sqlx.DB
}

func NewBillStore(db *sqlx.DB) BillStore {
	return &billStore{db: db}
}

func (s *billStore) Create(ctx context.Context, bill *models.Bill) error {
	query := `
		INSERT INTO bills (
			id, name, description, category, amount, currency,
			due_date, is_recurring, recurrence_rule, status,
			household_id, assigned_to, created_by, reminder_days,
			auto_pay_enabled, payment_method, vendor_info,
			attachment_urls, created_at, updated_at
		) VALUES (
			:id, :name, :description, :category, :amount, :currency,
			:due_date, :is_recurring, :recurrence_rule, :status,
			:household_id, :assigned_to, :created_by, :reminder_days,
			:auto_pay_enabled, :payment_method, :vendor_info,
			:attachment_urls, :created_at, :updated_at
		)
	`

	bill.CreatedAt = time.Now()
	bill.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, bill)
	if err != nil {
		return fmt.Errorf("failed to create bill: %w", err)
	}

	return nil
}

func (s *billStore) GetByID(ctx context.Context, id string) (*models.Bill, error) {
	query := `
		SELECT 
			id, name, description, category, amount, currency,
			due_date, is_recurring, recurrence_rule, status,
			household_id, assigned_to, created_by, reminder_days,
			auto_pay_enabled, payment_method, vendor_info,
			attachment_urls, paid_at, paid_by, paid_amount,
			created_at, updated_at
		FROM bills 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var bill models.Bill
	err := s.db.GetContext(ctx, &bill, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}

	return &bill, nil
}

func (s *billStore) GetUserBills(ctx context.Context, userID string, filter BillFilter) ([]*models.Bill, error) {
	query := `
		SELECT 
			b.id, b.name, b.description, b.category, b.amount, b.currency,
			b.due_date, b.is_recurring, b.recurrence_rule, b.status,
			b.household_id, b.assigned_to, b.created_by, b.reminder_days,
			b.auto_pay_enabled, b.payment_method, b.vendor_info,
			b.attachment_urls, b.paid_at, b.paid_by, b.paid_amount,
			b.created_at, b.updated_at
		FROM bills b
		JOIN households h ON b.household_id = h.id
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND b.deleted_at IS NULL
	`

	args := []interface{}{userID}
	argCount := 1

	// Apply filters
	if filter.Status != nil {
		argCount++
		query += fmt.Sprintf(" AND b.status = $%d", argCount)
		args = append(args, *filter.Status)
	}

	if filter.Category != nil {
		argCount++
		query += fmt.Sprintf(" AND b.category = $%d", argCount)
		args = append(args, *filter.Category)
	}

	if filter.DueAfter != nil {
		argCount++
		query += fmt.Sprintf(" AND b.due_date >= $%d", argCount)
		args = append(args, *filter.DueAfter)
	}

	if filter.DueBefore != nil {
		argCount++
		query += fmt.Sprintf(" AND b.due_date <= $%d", argCount)
		args = append(args, *filter.DueBefore)
	}

	if filter.Search != nil {
		argCount++
		query += fmt.Sprintf(" AND (b.name ILIKE $%d OR b.description ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+*filter.Search+"%")
	}

	// Order by due date
	query += " ORDER BY b.due_date ASC"

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

	var bills []*models.Bill
	err := s.db.SelectContext(ctx, &bills, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bills: %w", err)
	}

	return bills, nil
}

func (s *billStore) Update(ctx context.Context, bill *models.Bill) error {
	query := `
		UPDATE bills SET
			name = :name,
			description = :description,
			category = :category,
			amount = :amount,
			currency = :currency,
			due_date = :due_date,
			is_recurring = :is_recurring,
			recurrence_rule = :recurrence_rule,
			status = :status,
			assigned_to = :assigned_to,
			reminder_days = :reminder_days,
			auto_pay_enabled = :auto_pay_enabled,
			payment_method = :payment_method,
			vendor_info = :vendor_info,
			attachment_urls = :attachment_urls,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	bill.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, bill)
	if err != nil {
		return fmt.Errorf("failed to update bill: %w", err)
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

func (s *billStore) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE bills 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bill: %w", err)
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

func (s *billStore) MarkPaid(ctx context.Context, id string, paidBy string, amount float64) error {
	query := `
		UPDATE bills 
		SET 
			status = $1,
			paid_at = NOW(),
			paid_by = $2,
			paid_amount = $3,
			updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.BillStatusPaid, paidBy, amount, id)
	if err != nil {
		return fmt.Errorf("failed to mark bill as paid: %w", err)
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

func (s *billStore) MarkUnpaid(ctx context.Context, id string) error {
	query := `
		UPDATE bills 
		SET 
			status = $1,
			paid_at = NULL,
			paid_by = NULL,
			paid_amount = NULL,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, models.BillStatusPending, id)
	if err != nil {
		return fmt.Errorf("failed to mark bill as unpaid: %w", err)
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

func (s *billStore) GetUpcomingBills(ctx context.Context, userID string, days int) ([]*models.Bill, error) {
	endDate := time.Now().AddDate(0, 0, days)
	pendingStatus := models.BillStatusPending
	filter := BillFilter{
		Status:    &pendingStatus,
		DueBefore: &endDate,
		Limit:     100,
	}
	return s.GetUserBills(ctx, userID, filter)
}

func (s *billStore) GetOverdueBills(ctx context.Context, userID string) ([]*models.Bill, error) {
	now := time.Now()
	pendingStatus := models.BillStatusPending
	filter := BillFilter{
		Status:    &pendingStatus,
		DueBefore: &now,
		Limit:     100,
	}
	return s.GetUserBills(ctx, userID, filter)
}

// Payment operations
func (s *billStore) AddPayment(ctx context.Context, payment *models.BillPayment) error {
	query := `
		INSERT INTO bill_payments (
			id, bill_id, amount, payment_method, transaction_id,
			paid_by, paid_at, notes, created_at, updated_at
		) VALUES (
			:id, :bill_id, :amount, :payment_method, :transaction_id,
			:paid_by, :paid_at, :notes, :created_at, :updated_at
		)
	`

	payment.CreatedAt = time.Now()
	payment.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, payment)
	if err != nil {
		return fmt.Errorf("failed to add bill payment: %w", err)
	}

	return nil
}

func (s *billStore) GetPayments(ctx context.Context, billID string) ([]*models.BillPayment, error) {
	query := `
		SELECT 
			id, bill_id, amount, payment_method, transaction_id,
			paid_by, paid_at, notes, created_at, updated_at
		FROM bill_payments 
		WHERE bill_id = $1 AND deleted_at IS NULL
		ORDER BY paid_at DESC
	`

	var payments []*models.BillPayment
	err := s.db.SelectContext(ctx, &payments, query, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill payments: %w", err)
	}

	return payments, nil
}

func (s *billStore) DeletePayment(ctx context.Context, paymentID string) error {
	query := `
		UPDATE bill_payments 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, paymentID)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
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
