package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/response"
)

// ErrorHandler creates middleware that handles unhandled errors and returns JSON responses
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				handleError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// handleError processes errors and returns consistent JSON error responses
func handleError(w http.ResponseWriter) {
	statusCode := http.StatusInternalServerError
	message := "Internal Server Error"

	// Set content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode and send the response
	json.NewEncoder(w).Encode(response.Error{Message: message})
}
