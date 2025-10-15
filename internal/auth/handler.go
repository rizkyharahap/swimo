package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/response"
)

type AuthHandler struct {
	logger      *logger.Logger
	authUsecase AuthUsecase
}

func NewAuthHandler(logger *logger.Logger, authUsecase AuthUsecase) *AuthHandler {
	return &AuthHandler{logger, authUsecase}
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Sign up request received", "method", r.Method)
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("signup parse error", "error", err)

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response.Error{Message: "Invalid JSON Body."})
		return
	}

	// Validate request DTO
	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(response.Error{
			Message: "Validation Error.",
			Errors:  err.Errors,
		})
		return
	}

	if err := h.authUsecase.SignUp(r.Context(), req); err != nil {
		if errors.Is(err, ErrAccountExists) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response.Error{Message: "Email already exists."})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response.Error{Message: "Internal server error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response.Error{Message: "User registered successfully."})
}
