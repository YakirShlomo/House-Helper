package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEHub manages Server-Sent Events connections
type SSEHub struct {
	// Connected clients
	clients map[*SSEClient]bool

	// User to client mapping
	userClients map[string][]*SSEClient

	// Register channel for new clients
	register chan *SSEClient

	// Unregister channel for disconnecting clients
	unregister chan *SSEClient

	// Broadcast channel for sending messages
	broadcast chan SSEMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// SSEClient represents an SSE client connection
type SSEClient struct {
	// User ID
	userID string

	// Household ID
	householdID string

	// Response writer
	writer http.ResponseWriter

	// Request context
	ctx context.Context

	// Done channel to signal when client disconnects
	done chan struct{}

	// Hub reference
	hub *SSEHub
}

// SSEMessage represents a server-sent event message
type SSEMessage struct {
	ID          string      `json:"id,omitempty"`
	Event       string      `json:"event"`
	Data        interface{} `json:"data"`
	UserID      string      `json:"userId,omitempty"`
	HouseholdID string      `json:"householdId,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients:     make(map[*SSEClient]bool),
		userClients: make(map[string][]*SSEClient),
		register:    make(chan *SSEClient),
		unregister:  make(chan *SSEClient),
		broadcast:   make(chan SSEMessage),
	}
}

// Run starts the SSE hub
func (h *SSEHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.userID != "" {
				h.userClients[client.userID] = append(h.userClients[client.userID], client)
			}
			h.mu.Unlock()
			log.Printf("SSE client registered: user=%s, household=%s", client.userID, client.householdID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.done)

				// Remove from user clients mapping
				if client.userID != "" {
					clients := h.userClients[client.userID]
					for i, c := range clients {
						if c == client {
							h.userClients[client.userID] = append(clients[:i], clients[i+1:]...)
							break
						}
					}
					if len(h.userClients[client.userID]) == 0 {
						delete(h.userClients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("SSE client unregistered: user=%s, household=%s", client.userID, client.householdID)

		case message := <-h.broadcast:
			h.sendToClients(message)
		}
	}
}

// sendToClients sends a message to relevant clients
func (h *SSEHub) sendToClients(message SSEMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var targetClients []*SSEClient

	if message.UserID != "" {
		// Send to specific user
		targetClients = h.userClients[message.UserID]
	} else if message.HouseholdID != "" {
		// Send to all users in household
		for client := range h.clients {
			if client.householdID == message.HouseholdID {
				targetClients = append(targetClients, client)
			}
		}
	} else {
		// Broadcast to all clients
		for client := range h.clients {
			targetClients = append(targetClients, client)
		}
	}

	for _, client := range targetClients {
		select {
		case <-client.done:
			// Client is disconnected
			continue
		default:
			if err := client.sendMessage(message); err != nil {
				log.Printf("Error sending SSE message to user %s: %v", client.userID, err)
				h.unregister <- client
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *SSEHub) SendToUser(userID string, event string, data interface{}) {
	message := SSEMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Event:     event,
		Data:      data,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("SSE broadcast channel is full, dropping message")
	}
}

// SendToHousehold sends a message to all users in a household
func (h *SSEHub) SendToHousehold(householdID string, event string, data interface{}) {
	message := SSEMessage{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Event:       event,
		Data:        data,
		HouseholdID: householdID,
		Timestamp:   time.Now(),
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("SSE broadcast channel is full, dropping message")
	}
}

// ServeSSE handles SSE requests
func (h *SSEHub) ServeSSE(w http.ResponseWriter, r *http.Request, userID, householdID string) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create client
	client := &SSEClient{
		userID:      userID,
		householdID: householdID,
		writer:      w,
		ctx:         r.Context(),
		done:        make(chan struct{}),
		hub:         h,
	}

	// Register client
	h.register <- client

	// Send initial connection message
	connectionMsg := SSEMessage{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Event:     "connected",
		Data:      map[string]string{"status": "connected"},
		Timestamp: time.Now(),
	}
	client.sendMessage(connectionMsg)

	// Keep connection alive and handle disconnection
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("SSE: ResponseWriter does not support flushing")
		return
	}

	// Send periodic heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.ctx.Done():
			h.unregister <- client
			return
		case <-client.done:
			return
		case <-ticker.C:
			// Send heartbeat
			heartbeat := SSEMessage{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				Event:     "heartbeat",
				Data:      map[string]interface{}{"timestamp": time.Now().Unix()},
				Timestamp: time.Now(),
			}
			if err := client.sendMessage(heartbeat); err != nil {
				h.unregister <- client
				return
			}
			flusher.Flush()
		}
	}
}

// sendMessage sends an SSE message to the client
func (c *SSEClient) sendMessage(message SSEMessage) error {
	data, err := json.Marshal(message.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	// Write SSE formatted message
	if message.ID != "" {
		if _, err := fmt.Fprintf(c.writer, "id: %s\n", message.ID); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(c.writer, "event: %s\n", message.Event); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(c.writer, "data: %s\n\n", string(data)); err != nil {
		return err
	}

	// Flush the data to the client
	if flusher, ok := c.writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}