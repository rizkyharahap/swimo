package response

import (
	"encoding/json"
	"net/http"
)

type Success struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// Pagination represents the pagination metadata.
type Pagination struct {
	Page       int `json:"page" example:"1"`
	Limit      int `json:"limit" example:"10"`
	TotalPages int `json:"totalPages" example:"5"`
}

// SuccessPagination is a generic struct for paginated API responses.
// It uses a type parameter 'T' to make the Data field generic.
//
// The 'T' can be any type, and this struct will hold a slice of that type.
type SuccessPagination struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Error struct {
	Message string `json:"message"`
	Errors  any    `json:"errors,omitempty"`
}

// JSON writes any struct as JSON response
func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// BadRequest handles invalid JSON or malformed requests
func BadRequest(w http.ResponseWriter) {
	JSON(w, http.StatusBadRequest, Error{Message: "Invalid request body"})
}

// ValidationError wraps validation errors with 422 Unprocessable Entity
func ValidationError(w http.ResponseWriter, errors any) {
	JSON(w, http.StatusUnprocessableEntity, Error{Errors: errors, Message: "Validation errors"})
}

// InternalError wraps generic 500 Internal Server Error
func InternalError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, Error{Message: "Internal server error"})
}
