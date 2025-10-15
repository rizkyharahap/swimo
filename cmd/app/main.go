package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/database"
	"github.com/rizkyharahap/swimo/internal/auth"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/server"
)

func main() {
	// Load configuration
	cfg := config.Parse()

	// Initialize logger
	logConfig := logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		File:   cfg.Log.File,
		AddSrc: cfg.Log.AddSrc,
	}
	log := logger.New(logConfig)

	log.Info("Starting application",
		"name", cfg.App.Name,
		"env", cfg.App.Env,
		"version", "1.0.0",
	)

	// Create HTTP server
	httpServer := server.NewServer(cfg.HTTP, log)

	// Initialize database manager through server
	dbManager := database.NewManager(log)

	// Set up database connection
	db, err := dbManager.Connect(context.Background(), "primary", &cfg.Database)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	} else {
		log.Info("Database connection established successfully")
	}

	// Initialize auth components
	authRepo := auth.NewAuthRepository(db.Pool)
	authUsecase := auth.NewAuthUsecase(cfg, log, db.Pool, authRepo)
	authHandler := auth.NewAuthHandler(log, authUsecase)

	// Create router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, authHandler, db != nil)

	// Apply middlewares
	handler := middleware.Chain(
		middleware.ErrorHandler,
		middleware.RecoverMiddleware(log),
		middleware.LoggingMiddleware(log),
		middleware.CORSMiddleware(cfg.CORS),
		middleware.CompressionMiddleware,
	)(mux)

	// Set handler
	httpServer.WithHandler(handler)

	// Start server
	log.Info("Application initialized successfully")
	log.Info("Starting server...")

	if err := httpServer.Start(); err != nil {
		log.Error("Failed to start server", "error", err)
		panic(err)
	}
}

// setupRoutes sets up the application routes
func setupRoutes(mux *http.ServeMux, authHandler *auth.AuthHandler, hasDatabase bool) {
	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Health check requested")

		dbStatus := "disconnected"
		if hasDatabase {
			dbStatus = "connected"
		}

		response := fmt.Sprintf(`{"status":"healthy","timestamp":"%s","service":"swimo-api","database":"%s"}`,
			time.Now().UTC().Format(time.RFC3339), dbStatus)
		logger.Info("Health check response", "response", response)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	})

	// Auth endpoints
	if hasDatabase {
		mux.HandleFunc("POST /api/v1/sign-up", authHandler.SignUp)
	}
}
