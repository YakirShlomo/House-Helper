package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/yakirshlomo/house-helper/services/api/internal/services"
)

type Handlers struct {
	services *services.Services
	logger   *zap.Logger
}

func NewHandlers(services *services.Services, logger *zap.Logger) *Handlers {
	return &Handlers{
		services: services,
		logger:   logger,
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /healthz [get]
func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "house-helper-api",
		"version":   "1.0.0",
		"timestamp": "2025-01-09T12:00:00Z",
	})
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Get the current authenticated user's information
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /v1/me [get]
func (h *Handlers) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	user, err := h.services.Auth.GetUserByID(userID.(string))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", userID.(string)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateCurrentUser godoc
// @Summary Update current user
// @Description Update the current authenticated user's information
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body models.UpdateUserRequest true "User update data"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/me [put]
func (h *Handlers) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var req struct {
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		Language *string `json:"language,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement user update logic
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user_id": userID,
	})
}

// GetActivity godoc
// @Summary Get user activity
// @Description Get the activity history for the current user
// @Tags activity
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit number of activities" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.ActivityEntry
// @Failure 401 {object} map[string]string
// @Router /v1/activity [get]
func (h *Handlers) GetActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	householdID, _ := c.Get("household_id")

	// TODO: Implement activity retrieval
	activities := []map[string]interface{}{
		{
			"id":         "1",
			"type":       "task_completed",
			"message":    "Completed task: Take out trash",
			"timestamp":  "2025-01-09T10:30:00Z",
			"user_id":    userID,
			"household_id": householdID,
		},
		{
			"id":         "2",
			"type":       "timer_started",
			"message":    "Started laundry timer for 45 minutes",
			"timestamp":  "2025-01-09T09:15:00Z",
			"user_id":    userID,
			"household_id": householdID,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"total":     len(activities),
	})
}
