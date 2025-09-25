package middleware

import (
	"net/http"
	"time"

	"github.com/rizkyharahap/swimo/pkg/logger"
)

// LoggingMiddleware creates middleware that logs HTTP requests and responses
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response wrapper to capture status code
			wrapped := &responseWriter{w, http.StatusOK}

			// Log incoming request
			log.Info("Request started",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
				"proto", r.Proto,
			)

			// Add logger to context
			ctx := log.WithContext(r.Context())
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log completion
			duration := time.Since(start)
			log.Info("Request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration_ms", duration.Milliseconds(),
				"duration", duration.String(),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
