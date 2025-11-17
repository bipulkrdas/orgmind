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

	"github.com/bipulkrdas/orgmind/backend/internal/config"
	"github.com/bipulkrdas/orgmind/backend/internal/database"
	"github.com/bipulkrdas/orgmind/backend/internal/extraction"
	"github.com/bipulkrdas/orgmind/backend/internal/handler"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/bipulkrdas/orgmind/backend/internal/router"
	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/bipulkrdas/orgmind/backend/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Println("Configuration loaded successfully")
	log.Printf("Server will run on port: %s", cfg.ServerPort)
	log.Printf("Database URL configured: %s", maskDatabaseURL(cfg.DatabaseURL))
	log.Printf("JWT expiration: %d hours", cfg.JWTExpirationHours)

	// Initialize database connection
	log.Println("Connecting to database...")
	db, err := database.Connect(database.Config{
		DatabaseURL:     cfg.DatabaseURL,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		RetryAttempts:   3,
		RetryDelay:      2 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify database health
	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database health check passed")

	// Run migrations
	/*
		log.Println("Running database migrations...")
		migrationRunner := database.NewMigrationRunner(db, "migrations")
		if err := migrationRunner.Up(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")
	*/

	// Initialize repositories
	log.Println("Initializing repositories...")
	userRepo := repository.NewUserRepository(db.DB)
	documentRepo := repository.NewDocumentRepository(db.DB)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db.DB)
	graphRepo := repository.NewGraphRepository(db.DB)

	// Initialize storage service (S3)
	log.Println("Initializing storage service...")
	storageService, err := storage.NewS3StorageService(ctx, storage.S3Config{
		Region:          cfg.AWSRegion,
		Bucket:          cfg.AWSS3Bucket,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
	})
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	log.Println("Storage service initialized successfully")

	// Initialize Zep service
	log.Println("Initializing Zep service...")
	zepService, err := service.NewZepService(cfg.ZepAPIKey)
	if err != nil {
		log.Fatalf("Failed to initialize Zep service: %v", err)
	}
	log.Println("Zep service initialized successfully")

	// Initialize extraction service
	log.Println("Initializing extraction service...")
	extractionService := extraction.NewExtractionRouter(extraction.DefaultConfig())
	log.Println("Extraction service initialized successfully")

	// Initialize business services
	log.Println("Initializing business services...")
	authService := service.NewAuthService(userRepo, resetTokenRepo, cfg)
	graphService := service.NewGraphService(graphRepo, zepService)
	processingService := service.NewProcessingService(documentRepo, zepService)
	documentService := service.NewDocumentService(documentRepo, storageService, processingService, graphService, extractionService)

	// Initialize handlers
	log.Println("Initializing handlers...")
	authHandler := handler.NewAuthHandler(authService)
	documentHandler := handler.NewDocumentHandler(documentService)
	graphHandler := handler.NewGraphHandler(graphService, documentService, zepService)

	// Set up router with all handlers
	log.Println("Setting up router...")
	appRouter := router.NewRouter(authHandler, documentHandler, graphHandler, cfg)
	ginEngine := appRouter.Setup()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      ginEngine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s...", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Println("Server started successfully")
	log.Printf("OrgMind backend is running on http://localhost:%s", cfg.ServerPort)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}

// maskDatabaseURL masks sensitive parts of the database URL for logging
func maskDatabaseURL(url string) string {
	if len(url) > 20 {
		return url[:20] + "..."
	}
	return "***"
}
