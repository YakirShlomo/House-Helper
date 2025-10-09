package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
	"github.com/jmoiron/sqlx"
	"time"
)

type ShoppingStore interface {
	// List operations
	CreateList(ctx context.Context, list *models.ShoppingList) error
	GetListByID(ctx context.Context, id string) (*models.ShoppingList, error)
	GetUserLists(ctx context.Context, userID string) ([]*models.ShoppingList, error)
	UpdateList(ctx context.Context, list *models.ShoppingList) error
	DeleteList(ctx context.Context, id string) error
	ShareList(ctx context.Context, listID string, userID string, permission models.Permission) error
	
	// Item operations
	AddItem(ctx context.Context, item *models.ShoppingItem) error
	GetItem(ctx context.Context, id string) (*models.ShoppingItem, error)
	GetListItems(ctx context.Context, listID string) ([]*models.ShoppingItem, error)
	UpdateItem(ctx context.Context, item *models.ShoppingItem) error
	DeleteItem(ctx context.Context, id string) error
	MarkItemPurchased(ctx context.Context, id string, purchasedBy string) error
	MarkItemUnpurchased(ctx context.Context, id string) error
	
	// Search and suggestions
	SearchProducts(ctx context.Context, query string, limit int) ([]*models.Product, error)
	GetFrequentItems(ctx context.Context, userID string, limit int) ([]*models.Product, error)
}

type shoppingStore struct {
	db *sqlx.DB
}

func NewShoppingStore(db *sqlx.DB) ShoppingStore {
	return &shoppingStore{db: db}
}

// List operations
func (s *shoppingStore) CreateList(ctx context.Context, list *models.ShoppingList) error {
	query := `
		INSERT INTO shopping_lists (
			id, name, household_id, created_by, shared_with, settings,
			total_estimated_cost, created_at, updated_at
		) VALUES (
			:id, :name, :household_id, :created_by, :shared_with, :settings,
			:total_estimated_cost, :created_at, :updated_at
		)
	`

	list.CreatedAt = time.Now()
	list.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, list)
	if err != nil {
		return fmt.Errorf("failed to create shopping list: %w", err)
	}

	return nil
}

func (s *shoppingStore) GetListByID(ctx context.Context, id string) (*models.ShoppingList, error) {
	query := `
		SELECT 
			id, name, household_id, created_by, shared_with, settings,
			total_estimated_cost, created_at, updated_at
		FROM shopping_lists 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var list models.ShoppingList
	err := s.db.GetContext(ctx, &list, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get shopping list: %w", err)
	}

	return &list, nil
}

func (s *shoppingStore) GetUserLists(ctx context.Context, userID string) ([]*models.ShoppingList, error) {
	query := `
		SELECT DISTINCT
			sl.id, sl.name, sl.household_id, sl.created_by, sl.shared_with, 
			sl.settings, sl.total_estimated_cost, sl.created_at, sl.updated_at
		FROM shopping_lists sl
		JOIN households h ON sl.household_id = h.id
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND sl.deleted_at IS NULL
		ORDER BY sl.updated_at DESC
	`

	var lists []*models.ShoppingList
	err := s.db.SelectContext(ctx, &lists, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user shopping lists: %w", err)
	}

	return lists, nil
}

func (s *shoppingStore) UpdateList(ctx context.Context, list *models.ShoppingList) error {
	query := `
		UPDATE shopping_lists SET
			name = :name,
			shared_with = :shared_with,
			settings = :settings,
			total_estimated_cost = :total_estimated_cost,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	list.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, list)
	if err != nil {
		return fmt.Errorf("failed to update shopping list: %w", err)
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

func (s *shoppingStore) DeleteList(ctx context.Context, id string) error {
	query := `
		UPDATE shopping_lists 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete shopping list: %w", err)
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

func (s *shoppingStore) ShareList(ctx context.Context, listID string, userID string, permission models.Permission) error {
	// First check if list exists
	_, err := s.GetListByID(ctx, listID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO shopping_list_shares (id, list_id, user_id, permission, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (list_id, user_id) 
		DO UPDATE SET permission = $3, updated_at = NOW()
	`

	_, err = s.db.ExecContext(ctx, query, listID, userID, permission)
	if err != nil {
		return fmt.Errorf("failed to share shopping list: %w", err)
	}

	return nil
}

// Item operations
func (s *shoppingStore) AddItem(ctx context.Context, item *models.ShoppingItem) error {
	query := `
		INSERT INTO shopping_items (
			id, list_id, name, quantity, unit, category, notes,
			estimated_price, actual_price, barcode, is_purchased,
			purchased_by, purchased_at, added_by, created_at, updated_at
		) VALUES (
			:id, :list_id, :name, :quantity, :unit, :category, :notes,
			:estimated_price, :actual_price, :barcode, :is_purchased,
			:purchased_by, :purchased_at, :added_by, :created_at, :updated_at
		)
	`

	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	_, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to add shopping item: %w", err)
	}

	return nil
}

func (s *shoppingStore) GetItem(ctx context.Context, id string) (*models.ShoppingItem, error) {
	query := `
		SELECT 
			id, list_id, name, quantity, unit, category, notes,
			estimated_price, actual_price, barcode, is_purchased,
			purchased_by, purchased_at, added_by, created_at, updated_at
		FROM shopping_items 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var item models.ShoppingItem
	err := s.db.GetContext(ctx, &item, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get shopping item: %w", err)
	}

	return &item, nil
}

func (s *shoppingStore) GetListItems(ctx context.Context, listID string) ([]*models.ShoppingItem, error) {
	query := `
		SELECT 
			id, list_id, name, quantity, unit, category, notes,
			estimated_price, actual_price, barcode, is_purchased,
			purchased_by, purchased_at, added_by, created_at, updated_at
		FROM shopping_items 
		WHERE list_id = $1 AND deleted_at IS NULL
		ORDER BY is_purchased ASC, category ASC, name ASC
	`

	var items []*models.ShoppingItem
	err := s.db.SelectContext(ctx, &items, query, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopping list items: %w", err)
	}

	return items, nil
}

func (s *shoppingStore) UpdateItem(ctx context.Context, item *models.ShoppingItem) error {
	query := `
		UPDATE shopping_items SET
			name = :name,
			quantity = :quantity,
			unit = :unit,
			category = :category,
			notes = :notes,
			estimated_price = :estimated_price,
			actual_price = :actual_price,
			barcode = :barcode,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	item.UpdatedAt = time.Now()

	result, err := s.db.NamedExecContext(ctx, query, item)
	if err != nil {
		return fmt.Errorf("failed to update shopping item: %w", err)
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

func (s *shoppingStore) DeleteItem(ctx context.Context, id string) error {
	query := `
		UPDATE shopping_items 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete shopping item: %w", err)
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

func (s *shoppingStore) MarkItemPurchased(ctx context.Context, id string, purchasedBy string) error {
	query := `
		UPDATE shopping_items 
		SET 
			is_purchased = true,
			purchased_by = $1,
			purchased_at = NOW(),
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, purchasedBy, id)
	if err != nil {
		return fmt.Errorf("failed to mark item as purchased: %w", err)
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

func (s *shoppingStore) MarkItemUnpurchased(ctx context.Context, id string) error {
	query := `
		UPDATE shopping_items 
		SET 
			is_purchased = false,
			purchased_by = NULL,
			purchased_at = NULL,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark item as unpurchased: %w", err)
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

// Search and suggestions
func (s *shoppingStore) SearchProducts(ctx context.Context, query string, limit int) ([]*models.Product, error) {
	sqlQuery := `
		SELECT 
			id, name, category, typical_price, barcode, brand, description
		FROM products 
		WHERE name ILIKE $1 OR brand ILIKE $1 OR description ILIKE $1
		ORDER BY name ASC
		LIMIT $2
	`

	var products []*models.Product
	err := s.db.SelectContext(ctx, &products, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

func (s *shoppingStore) GetFrequentItems(ctx context.Context, userID string, limit int) ([]*models.Product, error) {
	query := `
		SELECT 
			p.id, p.name, p.category, p.typical_price, p.barcode, p.brand, p.description,
			COUNT(si.id) as purchase_count
		FROM products p
		JOIN shopping_items si ON p.name = si.name
		JOIN shopping_lists sl ON si.list_id = sl.id
		JOIN households h ON sl.household_id = h.id
		JOIN household_members hm ON h.id = hm.household_id
		WHERE hm.user_id = $1 AND si.is_purchased = true
		GROUP BY p.id, p.name, p.category, p.typical_price, p.barcode, p.brand, p.description
		ORDER BY purchase_count DESC, p.name ASC
		LIMIT $2
	`

	var products []*models.Product
	err := s.db.SelectContext(ctx, &products, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get frequent items: %w", err)
	}

	return products, nil
}
