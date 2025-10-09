package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BillRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	Amount      float64 `json:"amount" binding:"required,min=0"`
	Currency    string  `json:"currency" binding:"required"`
	DueDate     string  `json:"due_date" binding:"required"`
	Category    *string `json:"category,omitempty"`
	Recurrence  *string `json:"recurrence,omitempty"`
}

type BillResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	DueDate     time.Time  `json:"due_date"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	HouseholdID string     `json:"household_id"`
	Status      string     `json:"status"`
	Category    *string    `json:"category,omitempty"`
	Recurrence  *string    `json:"recurrence,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GetBills godoc
// @Summary Get bills
// @Description Get all bills for the current user's household
// @Tags bills
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (pending, paid, overdue)"
// @Param limit query int false "Limit number of bills" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} BillResponse
// @Failure 401 {object} map[string]string
// @Router /v1/bills [get]
func (h *Handlers) GetBills(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	// Mock bills data
	bills := []BillResponse{
		{
			ID:          "bill-1",
			Name:        "Electricity Bill",
			Description: stringPtr("Monthly electricity payment"),
			Amount:      150.75,
			Currency:    "USD",
			DueDate:     time.Now().Add(5 * 24 * time.Hour),
			HouseholdID: householdID.(string),
			Status:      "pending",
			Category:    stringPtr("utilities"),
			Recurrence:  stringPtr("monthly"),
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:          "bill-2",
			Name:        "Internet Bill",
			Description: stringPtr("Monthly internet service"),
			Amount:      79.99,
			Currency:    "USD",
			DueDate:     time.Now().Add(-2 * 24 * time.Hour),
			PaidAt:      timePtr(time.Now().Add(-1 * 24 * time.Hour)),
			HouseholdID: householdID.(string),
			Status:      "paid",
			Category:    stringPtr("utilities"),
			Recurrence:  stringPtr("monthly"),
			CreatedAt:   time.Now().Add(-35 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * 24 * time.Hour),
		},
		{
			ID:          "bill-3",
			Name:        "Water Bill",
			Amount:      45.30,
			Currency:    "USD",
			DueDate:     time.Now().Add(-3 * 24 * time.Hour),
			HouseholdID: householdID.(string),
			Status:      "overdue",
			Category:    stringPtr("utilities"),
			Recurrence:  stringPtr("monthly"),
			CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-40 * 24 * time.Hour),
		},
	}

	// Filter by status if provided
	status := c.Query("status")
	if status != "" {
		filtered := []BillResponse{}
		for _, bill := range bills {
			if bill.Status == status {
				filtered = append(filtered, bill)
			}
		}
		bills = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"bills": bills,
		"total": len(bills),
	})
}

// CreateBill godoc
// @Summary Create bill
// @Description Create a new bill
// @Tags bills
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param bill body BillRequest true "Bill data"
// @Success 201 {object} BillResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/bills [post]
func (h *Handlers) CreateBill(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	var req BillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse due date
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format. Use RFC3339."})
		return
	}

	// Determine status based on due date
	status := "pending"
	if dueDate.Before(time.Now()) {
		status = "overdue"
	}

	bill := BillResponse{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		DueDate:     dueDate,
		HouseholdID: householdID.(string),
		Status:      status,
		Category:    req.Category,
		Recurrence:  req.Recurrence,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	h.logger.Info("Bill created",
		zap.String("bill_id", bill.ID),
		zap.String("name", req.Name),
		zap.Float64("amount", req.Amount),
	)

	c.JSON(http.StatusCreated, bill)
}

// GetBill godoc
// @Summary Get bill
// @Description Get a specific bill by ID
// @Tags bills
// @Security BearerAuth
// @Produce json
// @Param id path string true "Bill ID"
// @Success 200 {object} BillResponse
// @Failure 404 {object} map[string]string
// @Router /v1/bills/{id} [get]
func (h *Handlers) GetBill(c *gin.Context) {
	billID := c.Param("id")
	householdID, _ := c.Get("household_id")

	// Mock bill retrieval
	bill := BillResponse{
		ID:          billID,
		Name:        "Electricity Bill",
		Description: stringPtr("Monthly electricity payment"),
		Amount:      150.75,
		Currency:    "USD",
		DueDate:     time.Now().Add(5 * 24 * time.Hour),
		HouseholdID: householdID.(string),
		Status:      "pending",
		Category:    stringPtr("utilities"),
		Recurrence:  stringPtr("monthly"),
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
	}

	c.JSON(http.StatusOK, bill)
}

// UpdateBill godoc
// @Summary Update bill
// @Description Update an existing bill
// @Tags bills
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Bill ID"
// @Param bill body BillRequest true "Updated bill data"
// @Success 200 {object} BillResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/bills/{id} [put]
func (h *Handlers) UpdateBill(c *gin.Context) {
	billID := c.Param("id")
	householdID, _ := c.Get("household_id")

	var req BillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse due date
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid due_date format. Use RFC3339."})
		return
	}

	bill := BillResponse{
		ID:          billID,
		Name:        req.Name,
		Description: req.Description,
		Amount:      req.Amount,
		Currency:    req.Currency,
		DueDate:     dueDate,
		HouseholdID: householdID.(string),
		Status:      "pending", // TODO: Calculate actual status
		Category:    req.Category,
		Recurrence:  req.Recurrence,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   time.Now(),
	}

	h.logger.Info("Bill updated",
		zap.String("bill_id", billID),
		zap.String("name", req.Name),
	)

	c.JSON(http.StatusOK, bill)
}

// DeleteBill godoc
// @Summary Delete bill
// @Description Delete a bill
// @Tags bills
// @Security BearerAuth
// @Param id path string true "Bill ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /v1/bills/{id} [delete]
func (h *Handlers) DeleteBill(c *gin.Context) {
	billID := c.Param("id")

	h.logger.Info("Bill deleted", zap.String("bill_id", billID))
	c.Status(http.StatusNoContent)
}

// PayBill godoc
// @Summary Pay bill
// @Description Mark a bill as paid
// @Tags bills
// @Security BearerAuth
// @Param id path string true "Bill ID"
// @Success 200 {object} BillResponse
// @Failure 404 {object} map[string]string
// @Router /v1/bills/{id}/pay [post]
func (h *Handlers) PayBill(c *gin.Context) {
	billID := c.Param("id")
	userID, _ := c.Get("user_id")
	householdID, _ := c.Get("household_id")

	// Mock bill payment
	now := time.Now()
	bill := BillResponse{
		ID:          billID,
		Name:        "Electricity Bill",
		Description: stringPtr("Monthly electricity payment"),
		Amount:      150.75,
		Currency:    "USD",
		DueDate:     time.Now().Add(5 * 24 * time.Hour),
		PaidAt:      &now,
		HouseholdID: householdID.(string),
		Status:      "paid",
		Category:    stringPtr("utilities"),
		Recurrence:  stringPtr("monthly"),
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:   now,
	}

	h.logger.Info("Bill paid",
		zap.String("bill_id", billID),
		zap.String("user_id", userID.(string)),
		zap.Float64("amount", bill.Amount),
	)

	c.JSON(http.StatusOK, bill)
}
