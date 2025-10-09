package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/househelper/temporal/internal/workflows"
	tlog "github.com/househelper/temporal/pkg/log"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var (
	temporalClient client.Client
	logger         *zap.Logger
)

const (
	TaskQueue = "house-helper-tasks"
)

func main() {
	// Initialize logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Get Temporal server address from environment or use default
	temporalAddr := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddr == "" {
		temporalAddr = "localhost:7233"
	}

	// Create Temporal client
	temporalClient, err = client.Dial(client.Options{
		HostPort:  temporalAddr,
		Namespace: getNamespace(),
		Logger:    tlog.NewZapAdapter(logger),
	})
	if err != nil {
		logger.Fatal("Failed to create Temporal client", zap.Error(err))
	}
	defer temporalClient.Close()

	// Setup HTTP routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/workflows/timer/start", startTimerHandler)
	http.HandleFunc("/api/v1/workflows/timer/pause", pauseTimerHandler)
	http.HandleFunc("/api/v1/workflows/timer/resume", resumeTimerHandler)
	http.HandleFunc("/api/v1/workflows/timer/stop", stopTimerHandler)
	http.HandleFunc("/api/v1/workflows/laundry/start", startLaundryHandler)
	http.HandleFunc("/api/v1/workflows/laundry/wash-complete", laundryWashCompleteHandler)
	http.HandleFunc("/api/v1/workflows/laundry/start-dry", startDryHandler)
	http.HandleFunc("/api/v1/workflows/laundry/dry-complete", dryCompleteHandler)
	http.HandleFunc("/api/v1/workflows/recurring-task/start", startRecurringTaskHandler)
	http.HandleFunc("/api/v1/workflows/recurring-task/cancel", cancelRecurringTaskHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	logger.Info("Starting Temporal API server", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func startTimerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params workflows.TimerWorkflowParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("timer-%s", params.TimerID),
		TaskQueue: TaskQueue,
	}

	we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.TimerWorkflow, params)
	if err != nil {
		logger.Error("Failed to start timer workflow", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"workflowId": we.GetID(),
		"runId":      we.GetRunID(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func pauseTimerHandler(w http.ResponseWriter, r *http.Request) {
	timerID := r.URL.Query().Get("timerId")
	if timerID == "" {
		http.Error(w, "timerId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("timer-%s", timerID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "pause_timer", nil)
	if err != nil {
		logger.Error("Failed to pause timer", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to pause timer: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "paused"})
}

func resumeTimerHandler(w http.ResponseWriter, r *http.Request) {
	timerID := r.URL.Query().Get("timerId")
	if timerID == "" {
		http.Error(w, "timerId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("timer-%s", timerID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "resume_timer", nil)
	if err != nil {
		logger.Error("Failed to resume timer", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to resume timer: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "resumed"})
}

func stopTimerHandler(w http.ResponseWriter, r *http.Request) {
	timerID := r.URL.Query().Get("timerId")
	if timerID == "" {
		http.Error(w, "timerId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("timer-%s", timerID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "stop_timer", nil)
	if err != nil {
		logger.Error("Failed to stop timer", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to stop timer: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func startLaundryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params workflows.LaundryWorkflowParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("laundry-%s", params.LaundryID),
		TaskQueue: TaskQueue,
	}

	we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.LaundryWorkflow, params)
	if err != nil {
		logger.Error("Failed to start laundry workflow", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"workflowId": we.GetID(),
		"runId":      we.GetRunID(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func laundryWashCompleteHandler(w http.ResponseWriter, r *http.Request) {
	laundryID := r.URL.Query().Get("laundryId")
	if laundryID == "" {
		http.Error(w, "laundryId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("laundry-%s", laundryID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "wash_complete", nil)
	if err != nil {
		logger.Error("Failed to signal wash complete", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to signal: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "wash_complete_signaled"})
}

func startDryHandler(w http.ResponseWriter, r *http.Request) {
	laundryID := r.URL.Query().Get("laundryId")
	if laundryID == "" {
		http.Error(w, "laundryId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("laundry-%s", laundryID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "start_dry", nil)
	if err != nil {
		logger.Error("Failed to start dry cycle", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to signal: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "dry_started"})
}

func dryCompleteHandler(w http.ResponseWriter, r *http.Request) {
	laundryID := r.URL.Query().Get("laundryId")
	if laundryID == "" {
		http.Error(w, "laundryId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("laundry-%s", laundryID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "dry_complete", nil)
	if err != nil {
		logger.Error("Failed to signal dry complete", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to signal: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "dry_complete_signaled"})
}

func startRecurringTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var params workflows.RecurringTaskWorkflowParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("recurring-task-%s", params.TaskID),
		TaskQueue: TaskQueue,
	}

	we, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.RecurringTaskWorkflow, params)
	if err != nil {
		logger.Error("Failed to start recurring task workflow", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"workflowId": we.GetID(),
		"runId":      we.GetRunID(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func cancelRecurringTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("taskId")
	if taskID == "" {
		http.Error(w, "taskId is required", http.StatusBadRequest)
		return
	}

	workflowID := fmt.Sprintf("recurring-task-%s", taskID)
	err := temporalClient.SignalWorkflow(context.Background(), workflowID, "", "cancel_recurring_task", nil)
	if err != nil {
		logger.Error("Failed to cancel recurring task", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to cancel: %v", err), http.StatusInternalServerError)
		return
	}

	// Also terminate the workflow after signal
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = temporalClient.TerminateWorkflow(ctx, workflowID, "", "User requested cancellation")
	if err != nil {
		logger.Warn("Failed to terminate workflow", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

func getNamespace() string {
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	return namespace
}
