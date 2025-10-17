package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/database"
	"github.com/rizkyharahap/swimo/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	server          *http.Server
	log             *logger.Logger
	config          config.HTTPConfig
	shutdownTimeout time.Duration
	dbManager       *database.Manager
}

// NewServer creates a new HTTP server with the given configuration
func NewServer(cfg config.HTTPConfig, log *logger.Logger) *Server {
	return &Server{
		config:          cfg,
		log:             log,
		shutdownTimeout: 30 * time.Second, // Default shutdown timeout
		dbManager:       database.NewManager(log),
	}
}

// WithHandler sets the main handler for the server
func (s *Server) WithHandler(handler http.Handler) *Server {
	s.server = &http.Server{
		Addr:         s.getAddress(),
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
	return s
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	if s.server == nil {
		return fmt.Errorf("server handler not set. Call WithHandler() first")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel for errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.log.Info("Starting HTTP server",
			"host", s.config.Host,
			"port", s.config.Port,
			"read_timeout", s.config.ReadTimeout,
			"write_timeout", s.config.WriteTimeout,
			"idle_timeout", s.config.IdleTimeout,
			"prefork", s.config.Prefork,
		)

		var err error
		if s.config.Prefork {
			err = s.startWithPrefork()
		} else {
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Channel for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either error or shutdown signal
	select {
	case err := <-serverErrors:
		return err
	case <-shutdown:
		s.log.Info("Shutdown signal received, starting graceful shutdown")
		return s.gracefulShutdown(ctx)
	}
}

// gracefulShutdown performs graceful shutdown of the server
func (s *Server) gracefulShutdown(ctx context.Context) error {
	// Create context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	s.log.Info("Shutting down server...", "timeout", s.shutdownTimeout)

	// Shutdown the server
	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.log.Error("Server shutdown failed", "error", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Close database connections
	if s.dbManager != nil {
		s.log.Info("Closing database connections...")
		if err := s.dbManager.CloseAll(); err != nil {
			s.log.Error("Failed to close database connections", "error", err)
			return fmt.Errorf("database shutdown failed: %w", err)
		}
		s.log.Info("Database connections closed successfully")
	}

	s.log.Info("Server shutdown completed successfully")
	return nil
}

// Stop stops the server gracefully
func (s *Server) Stop() error {
	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.gracefulShutdown(ctx)
}

// GetServer returns the underlying http.Server instance
func (s *Server) GetServer() *http.Server {
	return s.server
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	if s.server != nil {
		return s.server.Addr
	}
	return s.getAddress()
}

// getAddress returns the server address, handling empty host
func (s *Server) getAddress() string {
	host := s.config.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", host, s.config.Port)
}

// startWithPrefork starts the server with prefork (multiple processes)
func (s *Server) startWithPrefork() error {
	// Simple prefork implementation - just run multiple goroutines
	// This is a simplified version compared to Fiber's actual prefork
	s.log.Info("Starting prefork mode", "workers", runtime.NumCPU())

	// Create a channel to handle server errors
	serverErrors := make(chan error, 1)

	// Start multiple workers
	numWorkers := runtime.NumCPU()
	for i := range numWorkers {
		go func(workerID int) {
			s.log.Info("Starting prefork worker", "worker_id", workerID)

			// Create a new server instance for this worker
			workerServer := &http.Server{
				Addr:         s.getAddress(),
				Handler:      s.server.Handler,
				ReadTimeout:  s.server.ReadTimeout,
				WriteTimeout: s.server.WriteTimeout,
				IdleTimeout:  s.server.IdleTimeout,
			}

			if err := workerServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				select {
				case serverErrors <- fmt.Errorf("worker %d error: %w", workerID, err):
				default: // Don't block if channel is full
				}
			}
		}(i)
	}

	// Wait for any server error
	return <-serverErrors
}
