package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/yakirshlomo/house-helper/services/api/internal/config"
	"github.com/yakirshlomo/house-helper/services/api/internal/handlers"
	"github.com/yakirshlomo/house-helper/services/api/internal/middleware"
	"github.com/yakirshlomo/house-helper/services/api/internal/services"
	"github.com/yakirshlomo/house-helper/services/api/internal/store"
	"github.com/yakirshlomo/house-helper/services/api/pkg/kafka"
	"github.com/yakirshlomo/house-helper/services/api/pkg/temporal"
)

// @title House Helper API
// @version 1.0
// @description A comprehensive household management API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.househelper.app/support
// @contact.email support@househelper.app

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := store.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Kafka producer
	kafkaProducer := kafka.NewProducer(kafka.Config{
		Brokers: cfg.KafkaConfig.Brokers,
		Topic:   "house-helper-events",
	})
	if err != nil {
		logger.Warn("Failed to initialize Kafka producer", zap.Error(err))
		kafkaProducer = nil // Continue without Kafka for local development
	}
	if kafkaProducer != nil {
		defer kafkaProducer.Close()
	}

	// Initialize Temporal client
	temporalClient, err := temporal.NewClient(temporal.Config{
		HostPort:  cfg.TemporalConfig.HostPort,
		Namespace: cfg.TemporalConfig.Namespace,
		TaskQueue: "house-helper-tasks",
	})
	if err != nil {
		logger.Warn("Failed to initialize Temporal client", zap.Error(err))
		temporalClient = nil // Continue without Temporal for local development
	}
	if temporalClient != nil {
		defer temporalClient.Close()
	}

	// Initialize stores
	stores := &store.Stores{
		Users:    store.NewUserStore(db),
		Tasks:    store.NewTaskStore(db),
		Shopping: store.NewShoppingStore(db),
		Bills:    store.NewBillStore(db),
		Timers:   store.NewTimerStore(db),
		EventLog: store.NewEventLogStore(db),
	}

	// Initialize services
	services := &services.Services{
		Auth:         services.NewAuthService(stores.Users, cfg.JWTSecret),
		Task:         services.NewTaskService(stores.Tasks, stores.EventLog, kafkaProducer),
		Shopping:     services.NewShoppingService(stores.Shopping, stores.EventLog, kafkaProducer),
		Bill:         services.NewBillService(stores.Bills, stores.EventLog, kafkaProducer),
		Timer:        services.NewTimerService(stores.Timers, temporalClient, stores.EventLog),
		Notification: services.NewNotificationService(),
	}

	// Initialize handlers
	h := handlers.NewHandlers(services, logger)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Health check endpoint
	router.GET("/healthz", h.HealthCheck)

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", h.Signup)
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth(cfg.JWTSecret))
		{
			// User routes
			protected.GET("/me", h.GetCurrentUser)
			protected.PUT("/me", h.UpdateCurrentUser)

			// Task routes
			tasks := protected.Group("/tasks")
			{
				tasks.GET("", h.GetTasks)
				tasks.POST("", h.CreateTask)
				tasks.GET("/:id", h.GetTask)
				tasks.PUT("/:id", h.UpdateTask)
				tasks.DELETE("/:id", h.DeleteTask)
			}

			// Shopping routes
			shopping := protected.Group("/shopping")
			{
				lists := shopping.Group("/lists")
				{
					lists.GET("", h.GetShoppingLists)
					lists.POST("", h.CreateShoppingList)
					lists.GET("/:id", h.GetShoppingList)
					lists.PUT("/:id", h.UpdateShoppingList)
					lists.DELETE("/:id", h.DeleteShoppingList)

					// Shopping items
					lists.GET("/:id/items", h.GetShoppingItems)
					lists.POST("/:id/items", h.AddShoppingItem)
					lists.PUT("/:id/items/:item_id", h.UpdateShoppingItem)
					lists.DELETE("/:id/items/:item_id", h.DeleteShoppingItem)
				}
			}

			// Bill routes
			bills := protected.Group("/bills")
			{
				bills.GET("", h.GetBills)
				bills.POST("", h.CreateBill)
				bills.GET("/:id", h.GetBill)
				bills.PUT("/:id", h.UpdateBill)
				bills.DELETE("/:id", h.DeleteBill)
				bills.POST("/:id/pay", h.PayBill)
			}

			// Timer routes
			timers := protected.Group("/timers")
			{
				timers.GET("/active", h.GetActiveTimers)
				timers.POST("/start", h.StartTimer)
				timers.POST("/:id/cancel", h.CancelTimer)
				timers.GET("/:id", h.GetTimer)
			}

			// Activity routes
			protected.GET("/activity", h.GetActivity)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
