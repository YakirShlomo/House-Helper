package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yakirshlomo/house-helper/services/api/handlers"
	"github.com/yakirshlomo/house-helper/services/api/pkg/models"
)

// MockTaskService is a mock implementation of TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, task *models.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(ctx context.Context, familyID string) ([]*models.Task, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Task), args.Error(1)
}

func (m *MockTaskService) UpdateTask(ctx context.Context, id string, updates *models.TaskUpdate) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockTaskService) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTaskHandler_CreateTask_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskRequest := &models.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		DueDate:     time.Now().Add(24 * time.Hour),
		AssignedTo:  "user-123",
		Points:      10,
	}

	body, _ := json.Marshal(taskRequest)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))

	rec := httptest.NewRecorder()

	expectedTask := &models.Task{
		ID:          "task-123",
		Title:       taskRequest.Title,
		Description: taskRequest.Description,
		DueDate:     taskRequest.DueDate,
		AssignedTo:  taskRequest.AssignedTo,
		Points:      taskRequest.Points,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	mockService.On("CreateTask", mock.Anything, mock.AnythingOfType("*models.Task")).
		Return(nil).
		Run(func(args mock.Arguments) {
			task := args.Get(1).(*models.Task)
			task.ID = expectedTask.ID
		})

	// Act
	handler.CreateTask(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "task-123", response["id"])
	assert.Equal(t, "Test Task", response["title"])

	mockService.AssertExpectations(t)
}

func TestTaskHandler_CreateTask_InvalidInput(t *testing.T) {
	// Table-driven test for various invalid inputs
	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Missing Title",
			request: map[string]interface{}{
				"description": "Test",
				"due_date":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"assigned_to": "user-123",
				"points":      10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title is required",
		},
		{
			name: "Invalid Points",
			request: map[string]interface{}{
				"title":       "Test Task",
				"description": "Test",
				"due_date":    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"assigned_to": "user-123",
				"points":      -5,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "points must be positive",
		},
		{
			name: "Past Due Date",
			request: map[string]interface{}{
				"title":       "Test Task",
				"description": "Test",
				"due_date":    time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"assigned_to": "user-123",
				"points":      10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "due_date must be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(MockTaskService)
			handler := handlers.NewTaskHandler(mockService)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Act
			handler.CreateTask(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedError)
		})
	}
}

func TestTaskHandler_GetTask_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskID := "task-123"
	expectedTask := &models.Task{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      "pending",
		Points:      10,
		CreatedAt:   time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()

	mockService.On("GetTask", mock.Anything, taskID).Return(expectedTask, nil)

	// Act
	handler.GetTask(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedTask.ID, response.ID)
	assert.Equal(t, expectedTask.Title, response.Title)

	mockService.AssertExpectations(t)
}

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskID := "non-existent"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	rec := httptest.NewRecorder()

	mockService.On("GetTask", mock.Anything, taskID).
		Return(nil, models.ErrTaskNotFound)

	// Act
	handler.GetTask(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}

func TestTaskHandler_UpdateTask_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskID := "task-123"
	updates := &models.TaskUpdate{
		Status: stringPtr("completed"),
		Points: intPtr(15),
	}

	body, _ := json.Marshal(updates)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+taskID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	rec := httptest.NewRecorder()

	mockService.On("UpdateTask", mock.Anything, taskID, updates).Return(nil)

	// Act
	handler.UpdateTask(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestTaskHandler_DeleteTask_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskID := "task-123"
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID, nil)
	req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
	rec := httptest.NewRecorder()

	mockService.On("DeleteTask", mock.Anything, taskID).Return(nil)

	// Act
	handler.DeleteTask(rec, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockService.AssertExpectations(t)
}

func TestTaskHandler_ListTasks_Success(t *testing.T) {
	// Arrange
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	familyID := "family-123"
	expectedTasks := []*models.Task{
		{
			ID:     "task-1",
			Title:  "Task 1",
			Status: "pending",
			Points: 10,
		},
		{
			ID:     "task-2",
			Title:  "Task 2",
			Status: "completed",
			Points: 15,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?family_id="+familyID, nil)
	rec := httptest.NewRecorder()

	mockService.On("ListTasks", mock.Anything, familyID).Return(expectedTasks, nil)

	// Act
	handler.ListTasks(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response []*models.Task
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "task-1", response[0].ID)
	assert.Equal(t, "task-2", response[1].ID)

	mockService.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkTaskHandler_CreateTask(b *testing.B) {
	mockService := new(MockTaskService)
	handler := handlers.NewTaskHandler(mockService)

	taskRequest := &models.CreateTaskRequest{
		Title:       "Benchmark Task",
		Description: "Benchmark Description",
		DueDate:     time.Now().Add(24 * time.Hour),
		AssignedTo:  "user-123",
		Points:      10,
	}

	body, _ := json.Marshal(taskRequest)

	mockService.On("CreateTask", mock.Anything, mock.AnythingOfType("*models.Task")).
		Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "user_id", "user-123"))
		rec := httptest.NewRecorder()
		handler.CreateTask(rec, req)
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
