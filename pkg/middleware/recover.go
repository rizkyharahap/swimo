package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rizkyharahap/swimo/pkg/logger"
)

// RecoverMiddleware creates middleware that recovers from panics
func RecoverMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace
					stack := debug.Stack()

					// Log the panic with stack trace
					log.Error("Panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
						"stack", string(stack),
					)

					// Set content type
					w.Header().Set("Content-Type", "application/json")

					// Return internal server error
					w.WriteHeader(http.StatusInternalServerError)

					// Write error response
					response := fmt.Sprintf(`{"status":%d,"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`, http.StatusInternalServerError)
					w.Write([]byte(response))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
