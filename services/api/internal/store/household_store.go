package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type HouseholdStore interface {
	Create(ctx context.Context, household *models.Household) error
	GetByID(ctx context.Context, id string) (*models.Household, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Household, error)
	Update(ctx context.Context, household *models.Household) error
	Delete(ctx context.Context, id string) error
	
	// Member operations
	AddMember(ctx context.Context, householdID string, userID string, role models.HouseholdRole) error
	RemoveMember(ctx context.Context, householdID string, userID string) error
	UpdateMemberRole(ctx context.Context, householdID string, userID string, role models.HouseholdRole) error
	GetMembers(ctx context.Context, householdID string) ([]*models.HouseholdMember, error)
	IsMember(ctx context.Context, householdID string, userID string) (bool, error)
	
	// Invitation operations
	CreateInvitation(ctx context.Context, invitation *models.HouseholdInvitation) error
	GetInvitation(ctx context.Context, inviteCode string) (*models.HouseholdInvitation, error)
	AcceptInvitation(ctx context.Context, inviteCode string, userID string) error
	DeclineInvitation(ctx context.Context, inviteCode string) error
	GetPendingInvitations(ctx context.Context, householdID string) ([]*models.HouseholdInvitation, error)
}

type householdStore struct {
	db *sqlx.DB
}

func NewHouseholdStore(db *sqlx.DB) HouseholdStore {
	return &householdStore{db: db}
}

func (s *householdStore) Create(ctx context.Context, household *models.Household) error {
	query := `
		INSERT INTO households (
			id, name, description, timezone, currency,
			created_by, created_at, updated_at
		) VALUES (
			:id, :name, :description, :timezone, :currency,
			:created_by, :created_at, :updated_at
		)
	`

	household.CreatedAt = time.Now()
	household.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, household)
	if err != nil {
		return fmt.Errorf("failed to create household: %w", err)
	}

	// Add creator as admin member
	err = s.AddMember(ctx, household.ID, household.CreatedBy, models.HouseholdRoleAdmin)
	if err != nil {
		return fmt.Errorf("failed to add creator as admin: %w", err)
	}

	return nil
}

func (s *householdStore) GetByID(ctx context.Context, id string) (*models.Household, error) {
	query := `
		SELECT 
			id, name, description, timezone, currency,
			created_by, created_at, updated_at
		FROM households 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var household models.Household
	err := s.db.GetContext(ctx, &household, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get household: %w", err)
	}

	return &household, nil
}

func (s *householdStore) GetByUserID(ctx context.Context, userID string) ([]*models.Household, error) {
	query := `
		SELECT 
			h.id, h.name, h.description, h.timezone, h.currency,
			h.created_by, h.created_at, h.updated_at
		FROM households h
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND hm.left_at IS NULL AND h.deleted_at IS NULL
		ORDER BY hm.joined_at ASC
	`

	var households []*models.Household
	err := s.db.SelectContext(ctx, &households, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user households: %w", err)
	}

	return households, nil
}

func (s *householdStore) Update(ctx context.Context, household *models.Household) error {
	query := `
		UPDATE households SET
			name = :name,
			description = :description,
			timezone = :timezone,
			currency = :currency,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	household.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, household)
	if err != nil {
		return fmt.Errorf("failed to update household: %w", err)
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

func (s *householdStore) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE households 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete household: %w", err)
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

// Member operations
func (s *householdStore) AddMember(ctx context.Context, householdID string, userID string, role models.HouseholdRole) error {
	query := `
		INSERT INTO household_members (
			id, household_id, user_id, role, joined_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW()
		)
		ON CONFLICT (household_id, user_id) 
		DO UPDATE SET
			role = $3,
			left_at = NULL,
			updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query, householdID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add household member: %w", err)
	}

	return nil
}

func (s *householdStore) RemoveMember(ctx context.Context, householdID string, userID string) error {
	query := `
		UPDATE household_members 
		SET left_at = NOW(), updated_at = NOW()
		WHERE household_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, householdID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove household member: %w", err)
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

func (s *householdStore) UpdateMemberRole(ctx context.Context, householdID string, userID string, role models.HouseholdRole) error {
	query := `
		UPDATE household_members 
		SET role = $1, updated_at = NOW()
		WHERE household_id = $2 AND user_id = $3 AND left_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, role, householdID, userID)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
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

func (s *householdStore) GetMembers(ctx context.Context, householdID string) ([]*models.HouseholdMember, error) {
	query := `
		SELECT 
			hm.id, hm.household_id, hm.user_id, hm.role,
			hm.joined_at, hm.left_at, hm.created_at, hm.updated_at,
			u.first_name, u.last_name, u.email, u.phone
		FROM household_members hm
		JOIN users u ON hm.user_id = u.id
		WHERE hm.household_id = $1 AND hm.left_at IS NULL AND u.deleted_at IS NULL
		ORDER BY hm.joined_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, householdID)
	if err != nil {
		return nil, fmt.Errorf("failed to get household members: %w", err)
	}
	defer rows.Close()

	var members []*models.HouseholdMember
	for rows.Next() {
		var member models.HouseholdMember
		var firstName, lastName, email sql.NullString
		var phone sql.NullString

		err := rows.Scan(
			&member.ID, &member.HouseholdID, &member.UserID, &member.Role,
			&member.JoinedAt, &member.LeftAt, &member.CreatedAt, &member.UpdatedAt,
			&firstName, &lastName, &email, &phone,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan household member: %w", err)
		}

		// Set user info
		member.FirstName = firstName.String
		member.LastName = lastName.String
		member.Email = email.String
		member.Phone = phone.String

		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate member rows: %w", err)
	}

	return members, nil
}

func (s *householdStore) IsMember(ctx context.Context, householdID string, userID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM household_members 
		WHERE household_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	var count int
	err := s.db.GetContext(ctx, &count, query, householdID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check household membership: %w", err)
	}

	return count > 0, nil
}

// Invitation operations
func (s *householdStore) CreateInvitation(ctx context.Context, invitation *models.HouseholdInvitation) error {
	query := `
		INSERT INTO household_invitations (
			id, household_id, invited_by, invite_code, email,
			role, expires_at, created_at, updated_at
		) VALUES (
			:id, :household_id, :invited_by, :invite_code, :email,
			:role, :expires_at, :created_at, :updated_at
		)
	`

	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, invitation)
	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

func (s *householdStore) GetInvitation(ctx context.Context, inviteCode string) (*models.HouseholdInvitation, error) {
	query := `
		SELECT 
			id, household_id, invited_by, invite_code, email,
			role, status, expires_at, accepted_by, accepted_at,
			created_at, updated_at
		FROM household_invitations 
		WHERE invite_code = $1 AND deleted_at IS NULL
	`

	var invitation models.HouseholdInvitation
	err := s.db.GetContext(ctx, &invitation, query, inviteCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return &invitation, nil
}

func (s *householdStore) AcceptInvitation(ctx context.Context, inviteCode string, userID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get invitation
	invitation, err := s.GetInvitation(ctx, inviteCode)
	if err != nil {
		return err
	}

	// Check if invitation is valid
	if invitation.Status != models.InvitationStatusPending {
		return fmt.Errorf("invitation is not pending")
	}

	if invitation.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("invitation has expired")
	}

	// Update invitation status
	query := `
		UPDATE household_invitations 
		SET status = $1, accepted_by = $2, accepted_at = NOW(), updated_at = NOW()
		WHERE invite_code = $3
	`

	_, err = tx.ExecContext(ctx, query, models.InvitationStatusAccepted, userID, inviteCode)
	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	// Add user to household
	query = `
		INSERT INTO household_members (
			id, household_id, user_id, role, joined_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW()
		)
	`

	_, err = tx.ExecContext(ctx, query, invitation.HouseholdID, userID, invitation.Role)
	if err != nil {
		return fmt.Errorf("failed to add user to household: %w", err)
	}

	return tx.Commit()
}

func (s *householdStore) DeclineInvitation(ctx context.Context, inviteCode string) error {
	query := `
		UPDATE household_invitations 
		SET status = $1, updated_at = NOW()
		WHERE invite_code = $2 AND status = $3
	`

	result, err := s.db.ExecContext(ctx, query, models.InvitationStatusDeclined, inviteCode, models.InvitationStatusPending)
	if err != nil {
		return fmt.Errorf("failed to decline invitation: %w", err)
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

func (s *householdStore) GetPendingInvitations(ctx context.Context, householdID string) ([]*models.HouseholdInvitation, error) {
	query := `
		SELECT 
			id, household_id, invited_by, invite_code, email,
			role, status, expires_at, accepted_by, accepted_at,
			created_at, updated_at
		FROM household_invitations 
		WHERE household_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var invitations []*models.HouseholdInvitation
	err := s.db.SelectContext(ctx, &invitations, query, householdID, models.InvitationStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending invitations: %w", err)
	}

	return invitations, nil
}
