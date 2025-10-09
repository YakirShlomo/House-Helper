package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TaskRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description,omitempty"`
	DueAt       *string `json:"due_at,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      string     `json:"status"`
	Priority    *string    `json:"priority,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	HouseholdID string     `json:"household_id"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
}

// GetTasks godoc
// @Summary Get tasks
// @Description Get all tasks for the current user's household
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status (pending, completed, cancelled)"
// @Param limit query int false "Limit number of tasks" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} TaskResponse
// @Failure 401 {object} map[string]string
// @Router /v1/tasks [get]
func (h *Handlers) GetTasks(c *gin.Context) {
	householdID, exists := c.Get("household_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Household not found in token"})
		return
	}

	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Mock data for demonstration
	tasks := []TaskResponse{
		{
			ID:          "task-1",
			Title:       "Take out trash",
			Description: stringPtr("Weekly trash pickup"),
			Status:      "pending",
			Priority:    stringPtr("medium"),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
			DueAt:       timePtr(time.Now().Add(2 * time.Hour)),
			Tags:        []string{"weekly", "household"},
			HouseholdID: householdID.(string),
		},
		{
			ID:          "task-2",
			Title:       "Water plants",
			Description: stringPtr("Water all indoor plants"),
			Status:      "completed",
			Priority:    stringPtr("low"),
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
			CompletedAt: timePtr(time.Now().Add(-2 * time.Hour)),
			Tags:        []string{"plants", "daily"},
			HouseholdID: householdID.(string),
		},
		{
			ID:          "task-3",
			Title:       "Grocery shopping",
			Description: stringPtr("Buy items from shopping list"),
			Status:      "pending",
			Priority:    stringPtr("high"),
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			UpdatedAt:   time.Now().Add(-12 * time.Hour),
			DueAt:       timePtr(time.Now().Add(24 * time.Hour)),
			Tags:        []string{"shopping", "food"},
			HouseholdID: householdID.(string),
		},
	}

	// Filter by status if provided
	if status != "" {
		filtered := []TaskResponse{}
		for _, task := range tasks {
			if task.Status == status {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(tasks) {
		start = len(tasks)
	}
	if end > len(tasks) {
		end = len(tasks)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks[start:end],
		"total": len(tasks),
		"limit": limit,
		"offset": offset,
	})
}

// CreateTask godoc
// @Summary Create task
// @Description Create a new task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param task body TaskRequest true "Task data"
// @Success 201 {object} TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/tasks [post]
func (h *Handlers) CreateTask(c *gin.Context) {
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

	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create new task
	task := TaskResponse{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending",
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        req.Tags,
		HouseholdID: householdID.(string),
		AssignedTo:  stringPtr(userID.(string)),
	}

	// Parse due date if provided
	if req.DueAt != nil {
		if dueTime, err := time.Parse(time.RFC3339, *req.DueAt); err == nil {
			task.DueAt = &dueTime
		}
	}

	h.logger.Info("Task created",
		zap.String("task_id", task.ID),
		zap.String("user_id", userID.(string)),
		zap.String("household_id", householdID.(string)),
	)

	c.JSON(http.StatusCreated, task)
}

// GetTask godoc
// @Summary Get task
// @Description Get a specific task by ID
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 404 {object} map[string]string
// @Router /v1/tasks/{id} [get]
func (h *Handlers) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	householdID, _ := c.Get("household_id")

	// Mock task retrieval
	task := TaskResponse{
		ID:          taskID,
		Title:       "Sample Task",
		Description: stringPtr("This is a sample task"),
		Status:      "pending",
		Priority:    stringPtr("medium"),
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
		Tags:        []string{"sample"},
		HouseholdID: householdID.(string),
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTask godoc
// @Summary Update task
// @Description Update an existing task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param task body TaskRequest true "Updated task data"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/tasks/{id} [put]
func (h *Handlers) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")
	householdID, _ := c.Get("household_id")

	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock task update
	task := TaskResponse{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "pending", // TODO: Handle status updates
		Priority:    req.Priority,
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now(),
		Tags:        req.Tags,
		HouseholdID: householdID.(string),
		AssignedTo:  stringPtr(userID.(string)),
	}

	if req.DueAt != nil {
		if dueTime, err := time.Parse(time.RFC3339, *req.DueAt); err == nil {
			task.DueAt = &dueTime
		}
	}

	h.logger.Info("Task updated",
		zap.String("task_id", taskID),
		zap.String("user_id", userID.(string)),
	)

	c.JSON(http.StatusOK, task)
}

// DeleteTask godoc
// @Summary Delete task
// @Description Delete a task
// @Tags tasks
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /v1/tasks/{id} [delete]
func (h *Handlers) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := c.Get("user_id")

	h.logger.Info("Task deleted",
		zap.String("task_id", taskID),
		zap.String("user_id", userID.(string)),
	)

	c.Status(http.StatusNoContent)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
