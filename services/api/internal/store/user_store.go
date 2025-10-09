package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	
	// Profile operations
	UpdateProfile(ctx context.Context, userID string, profile *models.UserProfile) error
	GetProfile(ctx context.Context, userID string) (*models.UserProfile, error)
	
	// Household operations
	JoinHousehold(ctx context.Context, userID string, householdID string, role models.HouseholdRole) error
	LeaveHousehold(ctx context.Context, userID string, householdID string) error
	GetUserHouseholds(ctx context.Context, userID string) ([]*models.Household, error)
}

type userStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) UserStore {
	return &userStore{db: db}
}

func (s *userStore) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, phone, password_hash, first_name, last_name,
			is_verified, email_verified_at, phone_verified_at,
			created_at, updated_at
		) VALUES (
			:id, :email, :phone, :password_hash, :first_name, :last_name,
			:is_verified, :email_verified_at, :phone_verified_at,
			:created_at, :updated_at
		)
	`

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *userStore) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT 
			id, email, phone, password_hash, first_name, last_name,
			is_verified, email_verified_at, phone_verified_at,
			last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := s.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, email, phone, password_hash, first_name, last_name,
			is_verified, email_verified_at, phone_verified_at,
			last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := s.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT 
			id, email, phone, password_hash, first_name, last_name,
			is_verified, email_verified_at, phone_verified_at,
			last_login_at, created_at, updated_at
		FROM users 
		WHERE phone = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := s.db.GetContext(ctx, &user, query, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	return &user, nil
}

func (s *userStore) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			email = :email,
			phone = :phone,
			first_name = :first_name,
			last_name = :last_name,
			is_verified = :is_verified,
			email_verified_at = :email_verified_at,
			phone_verified_at = :phone_verified_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
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

func (s *userStore) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE users 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
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

func (s *userStore) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
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

func (s *userStore) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users 
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// Profile operations
func (s *userStore) UpdateProfile(ctx context.Context, userID string, profile *models.UserProfile) error {
	query := `
		INSERT INTO user_profiles (
			user_id, timezone, language, theme, avatar_url,
			date_format, time_format, notifications_enabled,
			email_notifications, push_notifications, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
		)
		ON CONFLICT (user_id) 
		DO UPDATE SET
			timezone = $2,
			language = $3,
			theme = $4,
			avatar_url = $5,
			date_format = $6,
			time_format = $7,
			notifications_enabled = $8,
			email_notifications = $9,
			push_notifications = $10,
			updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query,
		userID, profile.Timezone, profile.Language, profile.Theme,
		profile.AvatarURL, profile.DateFormat, profile.TimeFormat,
		profile.NotificationsEnabled, profile.EmailNotifications,
		profile.PushNotifications,
	)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (s *userStore) GetProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	query := `
		SELECT 
			timezone, language, theme, avatar_url, date_format, time_format,
			notifications_enabled, email_notifications, push_notifications,
			created_at, updated_at
		FROM user_profiles 
		WHERE user_id = $1
	`

	var profile models.UserProfile
	err := s.db.GetContext(ctx, &profile, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default profile if none exists
			return &models.UserProfile{
				Timezone:             "UTC",
				Language:             "en",
				Theme:                "system",
				DateFormat:           "YYYY-MM-DD",
				TimeFormat:           "24h",
				NotificationsEnabled: true,
				EmailNotifications:   true,
				PushNotifications:    true,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &profile, nil
}

// Household operations
func (s *userStore) JoinHousehold(ctx context.Context, userID string, householdID string, role models.HouseholdRole) error {
	query := `
		INSERT INTO household_members (
			id, household_id, user_id, role, joined_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW()
		)
		ON CONFLICT (household_id, user_id) 
		DO UPDATE SET
			role = $3,
			updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query, householdID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to join household: %w", err)
	}

	return nil
}

func (s *userStore) LeaveHousehold(ctx context.Context, userID string, householdID string) error {
	query := `
		UPDATE household_members 
		SET left_at = NOW(), updated_at = NOW()
		WHERE household_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, householdID, userID)
	if err != nil {
		return fmt.Errorf("failed to leave household: %w", err)
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

func (s *userStore) GetUserHouseholds(ctx context.Context, userID string) ([]*models.Household, error) {
	query := `
		SELECT 
			h.id, h.name, h.description, h.timezone, h.currency,
			h.created_by, h.created_at, h.updated_at,
			hm.role, hm.joined_at
		FROM households h
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND hm.left_at IS NULL AND h.deleted_at IS NULL
		ORDER BY hm.joined_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user households: %w", err)
	}
	defer rows.Close()

	var households []*models.Household
	for rows.Next() {
		var h models.Household
		var role models.HouseholdRole
		var joinedAt time.Time

		err := rows.Scan(
			&h.ID, &h.Name, &h.Description, &h.Timezone, &h.Currency,
			&h.CreatedBy, &h.CreatedAt, &h.UpdatedAt,
			&role, &joinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan household: %w", err)
		}

		households = append(households, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate household rows: %w", err)
	}

	return households, nil
}
