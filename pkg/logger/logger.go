package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

type Config struct {
	Level  string // debug|info|warn|error
	Format string // json|text
	File   string
	AddSrc bool
}

func New(cfg Config) *Logger {
	var handler slog.Handler

	// Set log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSrc,
	}

	// Determine output writer
	var writer = os.Stderr
	if cfg.File != "" {
		file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log := slog.New(slog.NewTextHandler(os.Stderr, opts))
			log.Error("failed to open log file, using stderr", "error", err)
		} else {
			writer = file
		}
	}

	// Create handler based on format
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	default:
		handler = slog.NewTextHandler(writer, opts)
	}

	// Create logger
	logger := slog.New(handler)
	return &Logger{Logger: logger}
}

// With returns a new Logger with additional key-value pairs
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// WithContext returns a context with the logger embedded
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// FromContext extracts a logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return logger
	}
	// Return default logger if none found in context
	return New(Config{Level: "info", Format: "text"})
}

type loggerKey struct{}
