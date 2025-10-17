package middleware

import (
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/response"
)

// ErrorHandler creates middleware that handles unhandled errors and returns JSON responses
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				response.JSON(w, http.StatusInternalServerError, response.Message{Message: "Internal Server Error"})
				return
			}
		}()

		next.ServeHTTP(w, r)
	})
}
