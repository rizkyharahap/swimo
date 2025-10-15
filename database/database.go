package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/pkg/logger"
)

// Database represents a single database connection
type Database struct {
	Pool   *pgxpool.Pool
	Name   string
	logger *logger.Logger
	mu     sync.RWMutex
	closed bool
}

// Manager handles multiple database connections
type Manager struct {
	databases map[string]*Database
	logger    *logger.Logger
	mu        sync.RWMutex
}

// NewManager creates a new database manager
func NewManager(logger *logger.Logger) *Manager {
	return &Manager{
		databases: make(map[string]*Database),
		logger:    logger,
	}
}

// Connect connects to a database with the given name and config
func (m *Manager) Connect(ctx context.Context, name string, config *config.DatabaseConfig) (*Database, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already connected
	if db, exists := m.databases[name]; exists && !db.closed {
		return db, nil
	}

	// Parse connection string
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set pool parameters
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create database instance
	db := &Database{
		Pool:   pool,
		Name:   name,
		logger: m.logger,
	}

	// Store in manager
	m.databases[name] = db

	m.logger.Info("Database connected", "name", name)
	return db, nil
}

// Get returns a database connection by name
func (m *Manager) Get(name string) (*Database, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	db, exists := m.databases[name]
	if !exists {
		return nil, fmt.Errorf("database '%s' not found", name)
	}

	if db.closed {
		return nil, fmt.Errorf("database '%s' is closed", name)
	}

	return db, nil
}

// Close closes a specific database connection
func (m *Manager) Close(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	db, exists := m.databases[name]
	if !exists {
		return fmt.Errorf("database '%s' not found", name)
	}

	if err := db.close(); err != nil {
		return err
	}

	delete(m.databases, name)
	return nil
}

// CloseAll closes all database connections
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	for name, db := range m.databases {
		if err := db.close(); err != nil {
			m.logger.Error("Failed to close database", "name", name, "error", err)
			errs = append(errs, err)
		}
	}

	m.databases = make(map[string]*Database)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d databases", len(errs))
	}

	return nil
}

// close internal close method
func (db *Database) close() error {
	if db.closed {
		return nil
	}

	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("Database closed", "name", db.Name)
	}

	db.closed = true
	return nil
}
