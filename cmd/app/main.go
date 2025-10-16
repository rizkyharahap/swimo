package main

import (
	"context"
	"net/http"
	"os"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/database"

	"github.com/rizkyharahap/swimo/internal/auth"
	"github.com/rizkyharahap/swimo/internal/health"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/server"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Swimo API
// @version 1.0
// @description This is the API documentation for Swimo - a swimming management and tracking application.
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @ExternalDocs.url https://github.com/rizkyharahap/swimo
// @ExternalDocs.description Swimo GitHub Repository

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

	// Initialize repositories
	authRepo := auth.NewAuthRepository(db.Pool)

	// Initialize usecases
	authUsecase := auth.NewAuthUsecase(cfg, log, db.Pool, authRepo)

	// Initialize handlers
	healthHandler := health.NewHealthHandler(log, db)
	authHandler := auth.NewAuthHandler(log, authUsecase)

	// Create router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, db, healthHandler, authHandler)

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
func setupRoutes(
	mux *http.ServeMux,
	db *database.Database,
	healthHandler *health.HealthHandler,
	authHandler *auth.AuthHandler,
) {
	// Serve swagger.json file directly
	mux.HandleFunc("GET /swagger/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "docs/swagger/swagger.json")
	})

	// Swagger UI endpoint
	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/docs"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))

	mux.HandleFunc("GET /api/v1/healthz", healthHandler.Check)

	// Auth endpoints
	if db != nil {
		mux.HandleFunc("POST /api/v1/sign-up", authHandler.SignUp)
		mux.HandleFunc("POST /api/v1/sign-in", authHandler.SignIn)
		mux.HandleFunc("POST /api/v1/sign-in-guest", authHandler.SignInGuest)
	}
}
