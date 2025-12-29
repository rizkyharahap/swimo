package middleware

import (
	"net/http"
	"time"

	"github.com/rizkyharahap/swimo/pkg/logger"
)

// wrappedWriter wraps http.ResponseWriter to capture status code
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *wrappedWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

// Logging creates middleware that logs HTTP requests and responses
func Logging(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response wrapper to capture status code
			wrapped := &wrappedWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

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
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"duration", duration.String(),
			)
		})
	}
}
