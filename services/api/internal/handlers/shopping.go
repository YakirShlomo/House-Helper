package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Shopping List Types
type ShoppingListRequest struct {
	Name string `json:"name" binding:"required"`
}

type ShoppingListResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	HouseholdID string                `json:"household_id"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Items       []ShoppingItemResponse `json:"items"`
}

type ShoppingItemRequest struct {
	Name           string   `json:"name" binding:"required"`
	Note           *string  `json:"note,omitempty"`
	Quantity       *int     `json:"quantity,omitempty"`
	EstimatedPrice *float64 `json:"estimated_price,omitempty"`
}

type ShoppingItemResponse struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Note             *string   `json:"note,omitempty"`
	Quantity         *int      `json:"quantity,omitempty"`
	EstimatedPrice   *float64  `json:"estimated_price,omitempty"`
	Completed        bool      `json:"completed"`
	CreatedAt        time.Time `json:"created_at"`
	CompletedByUser  *string   `json:"completed_by_user,omitempty"`
}

// GetShoppingLists godoc
// @Summary Get shopping lists
// @Description Get all shopping lists for the current user's household
// @Tags shopping
// @Security BearerAuth
// @Produce json
// @Success 200 {array} ShoppingListResponse
// @Failure 401 {object} map[string]string
// @Router /v1/shopping/lists [get]
func (h *Handlers) GetShoppingLists(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	// Mock shopping lists
	lists := []ShoppingListResponse{
		{
			ID:          "list-1",
			Name:        "Weekly Groceries",
			HouseholdID: householdID.(string),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
			Items: []ShoppingItemResponse{
				{
					ID:        "item-1",
					Name:      "Milk",
					Quantity:  intPtr(2),
					Completed: false,
					CreatedAt: time.Now().Add(-24 * time.Hour),
				},
				{
					ID:        "item-2",
					Name:      "Bread",
					Quantity:  intPtr(1),
					Completed: true,
					CreatedAt: time.Now().Add(-24 * time.Hour),
					CompletedByUser: stringPtr("demo-user-123"),
				},
				{
					ID:        "item-3",
					Name:      "Eggs",
					Quantity:  intPtr(12),
					EstimatedPrice: float64Ptr(3.99),
					Completed: false,
					CreatedAt: time.Now().Add(-12 * time.Hour),
				},
			},
		},
		{
			ID:          "list-2",
			Name:        "Cleaning Supplies",
			HouseholdID: householdID.(string),
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-48 * time.Hour),
			Items: []ShoppingItemResponse{
				{
					ID:        "item-4",
					Name:      "Dish Soap",
					Quantity:  intPtr(1),
					Completed: false,
					CreatedAt: time.Now().Add(-48 * time.Hour),
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"lists": lists,
		"total": len(lists),
	})
}

// CreateShoppingList godoc
// @Summary Create shopping list
// @Description Create a new shopping list
// @Tags shopping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param list body ShoppingListRequest true "Shopping list data"
// @Success 201 {object} ShoppingListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/shopping/lists [post]
func (h *Handlers) CreateShoppingList(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	var req ShoppingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list := ShoppingListResponse{
		ID:          uuid.New().String(),
		Name:        req.Name,
		HouseholdID: householdID.(string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       []ShoppingItemResponse{},
	}

	h.logger.Info("Shopping list created",
		zap.String("list_id", list.ID),
		zap.String("name", req.Name),
	)

	c.JSON(http.StatusCreated, list)
}

// GetShoppingList godoc
// @Summary Get shopping list
// @Description Get a specific shopping list by ID
// @Tags shopping
// @Security BearerAuth
// @Produce json
// @Param id path string true "Shopping List ID"
// @Success 200 {object} ShoppingListResponse
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id} [get]
func (h *Handlers) GetShoppingList(c *gin.Context) {
	listID := c.Param("id")
	householdID, _ := c.Get("household_id")

	// Mock shopping list
	list := ShoppingListResponse{
		ID:          listID,
		Name:        "Weekly Groceries",
		HouseholdID: householdID.(string),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-2 * time.Hour),
		Items: []ShoppingItemResponse{
			{
				ID:        "item-1",
				Name:      "Milk",
				Quantity:  intPtr(2),
				Completed: false,
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
		},
	}

	c.JSON(http.StatusOK, list)
}

// UpdateShoppingList godoc
// @Summary Update shopping list
// @Description Update a shopping list's name
// @Tags shopping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Shopping List ID"
// @Param list body ShoppingListRequest true "Updated shopping list data"
// @Success 200 {object} ShoppingListResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id} [put]
func (h *Handlers) UpdateShoppingList(c *gin.Context) {
	listID := c.Param("id")
	householdID, _ := c.Get("household_id")

	var req ShoppingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list := ShoppingListResponse{
		ID:          listID,
		Name:        req.Name,
		HouseholdID: householdID.(string),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		Items:       []ShoppingItemResponse{},
	}

	c.JSON(http.StatusOK, list)
}

// DeleteShoppingList godoc
// @Summary Delete shopping list
// @Description Delete a shopping list
// @Tags shopping
// @Security BearerAuth
// @Param id path string true "Shopping List ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id} [delete]
func (h *Handlers) DeleteShoppingList(c *gin.Context) {
	listID := c.Param("id")
	
	h.logger.Info("Shopping list deleted", zap.String("list_id", listID))
	c.Status(http.StatusNoContent)
}

// GetShoppingItems godoc
// @Summary Get shopping items
// @Description Get all items in a shopping list
// @Tags shopping
// @Security BearerAuth
// @Produce json
// @Param id path string true "Shopping List ID"
// @Success 200 {array} ShoppingItemResponse
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id}/items [get]
func (h *Handlers) GetShoppingItems(c *gin.Context) {
	listID := c.Param("id")

	items := []ShoppingItemResponse{
		{
			ID:        "item-1",
			Name:      "Milk",
			Quantity:  intPtr(2),
			Completed: false,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"list_id": listID,
	})
}

// AddShoppingItem godoc
// @Summary Add shopping item
// @Description Add a new item to a shopping list
// @Tags shopping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Shopping List ID"
// @Param item body ShoppingItemRequest true "Shopping item data"
// @Success 201 {object} ShoppingItemResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id}/items [post]
func (h *Handlers) AddShoppingItem(c *gin.Context) {
	listID := c.Param("id")

	var req ShoppingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := ShoppingItemResponse{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Note:           req.Note,
		Quantity:       req.Quantity,
		EstimatedPrice: req.EstimatedPrice,
		Completed:      false,
		CreatedAt:      time.Now(),
	}

	h.logger.Info("Shopping item added",
		zap.String("list_id", listID),
		zap.String("item_id", item.ID),
		zap.String("name", req.Name),
	)

	c.JSON(http.StatusCreated, item)
}

// UpdateShoppingItem godoc
// @Summary Update shopping item
// @Description Update a shopping item
// @Tags shopping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Shopping List ID"
// @Param item_id path string true "Shopping Item ID"
// @Param item body ShoppingItemRequest true "Updated shopping item data"
// @Success 200 {object} ShoppingItemResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id}/items/{item_id} [put]
func (h *Handlers) UpdateShoppingItem(c *gin.Context) {
	listID := c.Param("id")
	itemID := c.Param("item_id")
	userID, _ := c.Get("user_id")

	var req ShoppingItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := ShoppingItemResponse{
		ID:             itemID,
		Name:           req.Name,
		Note:           req.Note,
		Quantity:       req.Quantity,
		EstimatedPrice: req.EstimatedPrice,
		Completed:      false, // TODO: Handle completion status
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		CompletedByUser: stringPtr(userID.(string)),
	}

	h.logger.Info("Shopping item updated",
		zap.String("list_id", listID),
		zap.String("item_id", itemID),
	)

	c.JSON(http.StatusOK, item)
}

// DeleteShoppingItem godoc
// @Summary Delete shopping item
// @Description Delete a shopping item
// @Tags shopping
// @Security BearerAuth
// @Param id path string true "Shopping List ID"
// @Param item_id path string true "Shopping Item ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /v1/shopping/lists/{id}/items/{item_id} [delete]
func (h *Handlers) DeleteShoppingItem(c *gin.Context) {
	listID := c.Param("id")
	itemID := c.Param("item_id")

	h.logger.Info("Shopping item deleted",
		zap.String("list_id", listID),
		zap.String("item_id", itemID),
	)

	c.Status(http.StatusNoContent)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
