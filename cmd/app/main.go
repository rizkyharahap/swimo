package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rizkyharahap/swimo/config"
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

	// Create HTTP server with configuration options
	httpServer := server.NewServer(cfg.HTTP, log).

	// Create router
	mux := http.NewServeMux()

	// Setup routes
	setupRoutes(mux, log)

	// Apply middlewares
	handler := middleware.Chain(
		middleware.RecoverMiddleware(log),
		middleware.LoggingMiddleware(log),
		middleware.CORSMiddleware(cfg.CORS),
		middleware.CompressionMiddleware(),
	)(mux)

	// Set the handler
	httpServer.WithHandler(handler)

	// Start the server
	log.Info("Application initialized successfully")
	log.Info("Starting server...")
	if err := httpServer.Start(); err != nil {
		log.Error("Failed to start server", "error", err)
		panic(err)
	}
}

// setupRoutes sets up the application routes
func setupRoutes(mux *http.ServeMux, log *logger.Logger) {
	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Health check requested")

		response := fmt.Sprintf(`{"status":"healthy","timestamp":"%s","service":"swimo-api"}`, time.Now().UTC().Format(time.RFC3339))
		logger.Info("Health check response", "response", response)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, response)
	})

	// Hello endpoint for testing
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Hello endpoint requested")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"Hello, World!","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Test error endpoint
	mux.HandleFunc("GET /error", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Error("Test error endpoint requested")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":500,"error":{"code":"INTERNAL_ERROR","message":"This is a test error"}}`)
	})

	// Test panic endpoint
	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Panic endpoint requested - this will cause a panic")

		// This will be caught by the recover middleware
		panic("This is a test panic!")
	})

	// API routes group
	apiHandler := http.NewServeMux()
	apiHandler.HandleFunc("GET /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Users endpoint requested")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data":[{"id":1,"name":"John Doe","email":"john@example.com"},{"id":2,"name":"Jane Smith","email":"jane@example.com"}],"message":"Users retrieved successfully"}`)
	})

	apiHandler.HandleFunc("POST /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logger.FromContext(ctx)

		logger.Info("Create user endpoint requested", "method", r.Method, "content_length", r.ContentLength)

		// In a real application, you would parse the request body and create a user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"data":{"id":3,"name":"New User","email":"newuser@example.com","created_at":"%s"},"message":"User created successfully"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Mount API routes
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))

	// Static file serving (optional)
	// mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	log.Info("Routes configured successfully")
}
