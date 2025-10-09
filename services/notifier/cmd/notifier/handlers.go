package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/house-helper/notifier/pkg/notifications"
	"github.com/house-helper/notifier/pkg/sse"
	"github.com/house-helper/notifier/pkg/websocket"
)

// FCM notification request
type FCMTokenRequest struct {
	Token   string                 `json:"token"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	ImageURL string                `json:"imageUrl,omitempty"`
	Data    map[string]string      `json:"data,omitempty"`
}

type FCMTopicRequest struct {
	Topic   string                 `json:"topic"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	ImageURL string                `json:"imageUrl,omitempty"`
	Data    map[string]string      `json:"data,omitempty"`
}

// APNS notification request
type APNSRequest struct {
	DeviceToken     string            `json:"deviceToken"`
	BundleID        string            `json:"bundleId"`
	Title           string            `json:"title"`
	Body            string            `json:"body"`
	Badge           *int              `json:"badge,omitempty"`
	Sound           string            `json:"sound,omitempty"`
	Category        string            `json:"category,omitempty"`
	ThreadID        string            `json:"threadId,omitempty"`
	CustomData      map[string]string `json:"customData,omitempty"`
	MutableContent  bool              `json:"mutableContent,omitempty"`
	ContentState    map[string]string `json:"contentState,omitempty"`
	TargetContentID string            `json:"targetContentId,omitempty"`
}

// Broadcast request
type BroadcastRequest struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// handleFCMTokenNotification handles FCM token-based notifications
func handleFCMTokenNotification(fcmService *notifications.FCMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req FCMTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		payload := notifications.NotificationPayload{
			Title:    req.Title,
			Body:     req.Body,
			ImageURL: req.ImageURL,
			Data:     req.Data,
		}

		messageID, err := fcmService.SendToToken(r.Context(), req.Token, payload)
		if err != nil {
			log.Printf("Failed to send FCM notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		response := map[string]string{"messageId": messageID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleFCMTopicNotification handles FCM topic-based notifications
func handleFCMTopicNotification(fcmService *notifications.FCMService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req FCMTopicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		payload := notifications.NotificationPayload{
			Title:    req.Title,
			Body:     req.Body,
			ImageURL: req.ImageURL,
			Data:     req.Data,
		}

		messageID, err := fcmService.SendToTopic(r.Context(), req.Topic, payload)
		if err != nil {
			log.Printf("Failed to send FCM topic notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		response := map[string]string{"messageId": messageID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handleAPNSNotification handles APNS notifications
func handleAPNSNotification(apnsService *notifications.APNSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req APNSRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		payload := notifications.APNSPayload{
			Title:           req.Title,
			Body:            req.Body,
			Badge:           req.Badge,
			Sound:           req.Sound,
			Category:        req.Category,
			ThreadID:        req.ThreadID,
			CustomData:      req.CustomData,
			MutableContent:  req.MutableContent,
			ContentState:    req.ContentState,
			TargetContentID: req.TargetContentID,
		}

		response, err := apnsService.SendNotification(r.Context(), req.BundleID, req.DeviceToken, payload)
		if err != nil {
			log.Printf("Failed to send APNS notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		result := map[string]interface{}{
			"apnsId":     response.ApnsID,
			"statusCode": response.StatusCode,
			"sent":       response.Sent(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// handleAPNSSilentNotification handles APNS silent notifications
func handleAPNSSilentNotification(apnsService *notifications.APNSService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DeviceToken string            `json:"deviceToken"`
			BundleID    string            `json:"bundleId"`
			CustomData  map[string]string `json:"customData,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		response, err := apnsService.SendSilentNotification(r.Context(), req.BundleID, req.DeviceToken, req.CustomData)
		if err != nil {
			log.Printf("Failed to send APNS silent notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		result := map[string]interface{}{
			"apnsId":     response.ApnsID,
			"statusCode": response.StatusCode,
			"sent":       response.Sent(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// handleUserBroadcast handles broadcasting to a specific user
func handleUserBroadcast(wsHub *websocket.Hub, sseHub *sse.SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId is required", http.StatusBadRequest)
			return
		}

		var req BroadcastRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Send via WebSocket
		wsMessage := websocket.Message{
			Type:      req.Type,
			UserID:    userID,
			Data:      req.Data,
			Timestamp: time.Now(),
		}
		wsHub.BroadcastToUser(userID, wsMessage)

		// Send via SSE
		sseHub.SendToUser(userID, req.Type, req.Data)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message sent"))
	}
}

// handleHouseholdBroadcast handles broadcasting to a household
func handleHouseholdBroadcast(wsHub *websocket.Hub, sseHub *sse.SSEHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		householdID := r.URL.Query().Get("householdId")
		if householdID == "" {
			http.Error(w, "householdId is required", http.StatusBadRequest)
			return
		}

		var req BroadcastRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Send via WebSocket
		wsMessage := websocket.Message{
			Type:        req.Type,
			HouseholdID: householdID,
			Data:        req.Data,
			Timestamp:   time.Now(),
		}
		wsHub.BroadcastToHousehold(householdID, wsMessage)

		// Send via SSE
		sseHub.SendToHousehold(householdID, req.Type, req.Data)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Message sent"))
	}
}