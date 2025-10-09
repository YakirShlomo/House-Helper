package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/house-helper/notifier/pkg/notifications"
	"github.com/house-helper/notifier/pkg/sse"
	"github.com/house-helper/notifier/pkg/websocket"
)

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize services
	var fcmService *notifications.FCMService
	var apnsService *notifications.APNSService
	var err error

	// Initialize FCM if credentials are provided
	if config.FCMCredentialsPath != "" {
		fcmService, err = notifications.NewFCMService(config.FCMCredentialsPath)
		if err != nil {
			log.Printf("Failed to initialize FCM service: %v", err)
		} else {
			log.Println("FCM service initialized successfully")
		}
	}

	// Initialize APNS if credentials are provided
	if config.APNSCertPath != "" && config.APNSKeyPath != "" {
		apnsService, err = notifications.NewAPNSServiceWithCertificate(
			config.APNSCertPath,
			config.APNSKeyPath,
			config.APNSProduction,
		)
		if err != nil {
			log.Printf("Failed to initialize APNS service with certificate: %v", err)
		} else {
			log.Println("APNS service initialized successfully with certificate")
		}
	} else if config.APNSKeyPath != "" && config.APNSKeyID != "" && config.APNSTeamID != "" {
		apnsService, err = notifications.NewAPNSServiceWithToken(
			config.APNSKeyPath,
			config.APNSKeyID,
			config.APNSTeamID,
			config.APNSProduction,
		)
		if err != nil {
			log.Printf("Failed to initialize APNS service with token: %v", err)
		} else {
			log.Println("APNS service initialized successfully with token")
		}
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize SSE hub
	sseHub := sse.NewSSEHub()
	go sseHub.Run()

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		householdID := r.URL.Query().Get("householdId")

		if userID == "" {
			http.Error(w, "userId is required", http.StatusBadRequest)
			return
		}

		wsHub.ServeWS(w, r, userID, householdID)
	})

	// Server-Sent Events endpoint
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		householdID := r.URL.Query().Get("householdId")

		if userID == "" {
			http.Error(w, "userId is required", http.StatusBadRequest)
			return
		}

		sseHub.ServeSSE(w, r, userID, householdID)
	})

	// Push notification endpoints
	if fcmService != nil {
		mux.HandleFunc("/notify/fcm/token", handleFCMTokenNotification(fcmService))
		mux.HandleFunc("/notify/fcm/topic", handleFCMTopicNotification(fcmService))
	}

	if apnsService != nil {
		mux.HandleFunc("/notify/apns", handleAPNSNotification(apnsService))
		mux.HandleFunc("/notify/apns/silent", handleAPNSSilentNotification(apnsService))
	}

	// Real-time message endpoints
	mux.HandleFunc("/broadcast/user", handleUserBroadcast(wsHub, sseHub))
	mux.HandleFunc("/broadcast/household", handleHouseholdBroadcast(wsHub, sseHub))

	// Create HTTP server
	server := &http.Server{
		Addr:         config.Port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Starting notifier service on %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down notifier service...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close APNS service
	if apnsService != nil {
		apnsService.Close()
	}

	log.Println("Notifier service stopped")
}

// Config holds application configuration
type Config struct {
	Port               string
	FCMCredentialsPath string
	APNSCertPath       string
	APNSKeyPath        string
	APNSKeyID          string
	APNSTeamID         string
	APNSProduction     bool
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port:               getEnv("PORT", ":8080"),
		FCMCredentialsPath: getEnv("FCM_CREDENTIALS_PATH", ""),
		APNSCertPath:       getEnv("APNS_CERT_PATH", ""),
		APNSKeyPath:        getEnv("APNS_KEY_PATH", ""),
		APNSKeyID:          getEnv("APNS_KEY_ID", ""),
		APNSTeamID:         getEnv("APNS_TEAM_ID", ""),
		APNSProduction:     getEnv("APNS_PRODUCTION", "false") == "true",
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
