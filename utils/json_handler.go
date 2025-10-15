package utils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Error represents an error response
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// respond is the internal method that writes the JSON response
func SuccessResponse(w http.ResponseWriter, statusCode int, response any) {
	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(statusCode)

	// Encode and write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the response, we can't do much more than log it
		// In a real application, you might want to log this error
		return
	}
}
