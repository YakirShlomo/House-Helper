package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TimerRequest struct {
	Type     string  `json:"type" binding:"required"`
	Duration int     `json:"duration_seconds" binding:"required,min=1"`
	TaskID   *string `json:"task_id,omitempty"`
	Title    *string `json:"title,omitempty"`
}

type TimerResponse struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Duration    int        `json:"duration_seconds"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	TaskID      *string    `json:"task_id,omitempty"`
	HouseholdID string     `json:"household_id"`
	Status      string     `json:"status"`
}

// GetActiveTimers godoc
// @Summary Get active timers
// @Description Get all active timers for the current user's household
// @Tags timers
// @Security BearerAuth
// @Produce json
// @Success 200 {array} TimerResponse
// @Failure 401 {object} map[string]string
// @Router /v1/timers/active [get]
func (h *Handlers) GetActiveTimers(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	// Mock active timers
	timers := []TimerResponse{
		{
			ID:          "timer-1",
			Type:        "laundry",
			Title:       "Laundry Timer",
			Duration:    2700, // 45 minutes
			StartedAt:   time.Now().Add(-30 * time.Minute),
			HouseholdID: householdID.(string),
			Status:      "running",
		},
		{
			ID:          "timer-2",
			Type:        "cooking",
			Title:       "Cooking Timer",
			Duration:    1800, // 30 minutes
			StartedAt:   time.Now().Add(-10 * time.Minute),
			HouseholdID: householdID.(string),
			Status:      "running",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"timers": timers,
		"total":  len(timers),
	})
}

// StartTimer godoc
// @Summary Start timer
// @Description Start a new timer using Temporal workflow
// @Tags timers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param timer body TimerRequest true "Timer data"
// @Success 201 {object} TimerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/timers/start [post]
func (h *Handlers) StartTimer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	var req TimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create timer response
	timerID := uuid.New().String()
	title := req.Type + " Timer"
	if req.Title != nil {
		title = *req.Title
	}

	timer := TimerResponse{
		ID:          timerID,
		Type:        req.Type,
		Title:       title,
		Duration:    req.Duration,
		StartedAt:   time.Now(),
		TaskID:      req.TaskID,
		HouseholdID: householdID.(string),
		Status:      "running",
	}

	// TODO: Start Temporal workflow
	h.logger.Info("Timer started",
		zap.String("timer_id", timerID),
		zap.String("type", req.Type),
		zap.Int("duration", req.Duration),
		zap.String("user_id", userID.(string)),
		zap.String("household_id", householdID.(string)),
	)

	c.JSON(http.StatusCreated, timer)
}

// CancelTimer godoc
// @Summary Cancel timer
// @Description Cancel a running timer
// @Tags timers
// @Security BearerAuth
// @Param id path string true "Timer ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/timers/{id}/cancel [post]
func (h *Handlers) CancelTimer(c *gin.Context) {
	timerID := c.Param("id")
	userID, _ := c.Get("user_id")

	// TODO: Cancel Temporal workflow
	h.logger.Info("Timer cancelled",
		zap.String("timer_id", timerID),
		zap.String("user_id", userID.(string)),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Timer cancelled successfully",
		"timer_id": timerID,
	})
}

// GetTimer godoc
// @Summary Get timer
// @Description Get a specific timer by ID
// @Tags timers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Timer ID"
// @Success 200 {object} TimerResponse
// @Failure 404 {object} map[string]string
// @Router /v1/timers/{id} [get]
func (h *Handlers) GetTimer(c *gin.Context) {
	timerID := c.Param("id")
	householdID, _ := c.Get("household_id")

	// Mock timer retrieval
	timer := TimerResponse{
		ID:          timerID,
		Type:        "laundry",
		Title:       "Laundry Timer",
		Duration:    2700,
		StartedAt:   time.Now().Add(-30 * time.Minute),
		HouseholdID: householdID.(string),
		Status:      "running",
	}

	c.JSON(http.StatusOK, timer)
}
