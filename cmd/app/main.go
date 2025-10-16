package main

import (
	"context"
	"net/http"
	"os"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/database"

	"github.com/rizkyharahap/swimo/internal/auth"
	"github.com/rizkyharahap/swimo/internal/health"
	"github.com/rizkyharahap/swimo/internal/swagger"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/server"
)

// @title Swimo API
// @version 1.0
// @description This is the API documentation for Swimo - a swimming management and tracking application.
// @termsOfService http://swagger.io/terms/

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1
// @schemes http https

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
	db, err := dbManager.Connect(context.Background(), "primary", &cfg.Database, &cfg.App)
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
	swaggerHandler := swagger.NewSwaggerHandler(cfg, log)
	authHandler := auth.NewAuthHandler(log, authUsecase)

	// Create router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, db, cfg, healthHandler, swaggerHandler, authHandler)

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
	cfg *config.Config,
	healthHandler *health.HealthHandler,
	swaggerHandler *swagger.SwaggerHandler,
	authHandler *auth.AuthHandler,
) {
	// Serve swagger.json file with dynamic host configuration
	mux.HandleFunc("GET /swagger/docs", swaggerHandler.Docs)
	mux.HandleFunc("GET /swagger/", swaggerHandler.Handler)

	// Health check endpoint
	mux.HandleFunc("GET /api/v1/healthz", healthHandler.Check)

	if db != nil {
		// Public endpoints - no authentication required
		mux.HandleFunc("POST /api/v1/sign-up", authHandler.SignUp)
		mux.HandleFunc("POST /api/v1/sign-in", authHandler.SignIn)
		mux.HandleFunc("POST /api/v1/sign-in-guest", authHandler.SignInGuest)
		mux.HandleFunc("POST /api/v1/refresh-token", authHandler.RefreshToken)

		// Protected endpoints - require authentication
		authMiddleware := func(h http.HandlerFunc) http.Handler {
			return middleware.AuthMiddleware(cfg.Auth.JWTSecret, h)
		}

		mux.Handle("POST /api/v1/sign-out", authMiddleware(authHandler.SignOut))
	}
}
