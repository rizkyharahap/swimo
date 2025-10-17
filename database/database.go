package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/pkg/logger"
)

// Database represents a single database connection
type Database struct {
	Pool   *pgxpool.Pool
	Name   string
	log    *logger.Logger
	mu     sync.RWMutex
	closed bool
}

// Manager handles multiple database connections
type Manager struct {
	databases map[string]*Database
	log       *logger.Logger
	mu        sync.RWMutex
}

type pgxTracer struct {
	log *logger.Logger
}

func (t pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	fullQuery := buildFullQuery(data.SQL, data.Args)
	t.log.Debug("[PGX] QUERY START", "sql", fullQuery)
	return ctx
}

func (t pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		t.log.Error("PGX QUERY ERROR", "err", data.Err)
	} else {
		t.log.Debug("PGX QUERY END", "duration", data.CommandTag.String())
	}
}

// buildFullQuery safely substitutes $1, $2... placeholders with real argument values
func buildFullQuery(sql string, args []any) string {
	result := sql

	for i, arg := range args {
		placeholder := fmt.Sprintf(`\$%d`, i+1)
		var replacement string

		switch v := arg.(type) {
		case string:
			replacement = fmt.Sprintf("'%s'", escapeQuotes(v))
		case []byte:
			replacement = fmt.Sprintf("'%x'", v)
		case time.Time:
			replacement = fmt.Sprintf("'%s'", v.Format(time.RFC3339))
		case nil:
			replacement = "NULL"
		default:
			replacement = fmt.Sprintf("%v", v)
		}

		result = regexp.MustCompile(placeholder).ReplaceAllString(result, replacement)
	}

	// Clean multiple spaces & newlines
	result = strings.ReplaceAll(result, "\n", " ")
	result = strings.Join(strings.Fields(result), " ")
	return result
}

// escapeQuotes escapes single quotes to prevent broken SQL logs
func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// NewManager creates a new database manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		databases: make(map[string]*Database),
		log:       log,
	}
}

// Connect connects to a database with the given name and config
func (m *Manager) Connect(ctx context.Context, name string, config *config.DatabaseConfig, appConfig *config.AppConfig) (*Database, error) {
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

	if appConfig.Env == "dev" {
		poolConfig.ConnConfig.Tracer = pgxTracer{log: m.log}
	}

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
		Pool: pool,
		Name: name,
		log:  m.log,
	}

	// Store in manager
	m.databases[name] = db

	m.log.Info("Database connected", "name", name)
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
			m.log.Error("Failed to close database", "name", name, "error", err)
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
		db.log.Info("Database closed", "name", db.Name)
	}

	db.closed = true
	return nil
}
