package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/middleware"
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
// @Summary Sign up new user
// @Description Register a new user account with email, password, and profile information
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Sign up request with user details"
// @Success 201 {object} response.Success "User registered successfully"
// @Failure 400 {object} response.Error "Invalid request body"
// @Failure 422 {object} response.Error "Validation errors"
// @Failure 409 {object} response.Error "Email already exists"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /sign-up [post]
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	// Validate request DTO
	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.Errors)
		return
	}

	if err := h.authUsecase.SignUp(r.Context(), req); err != nil {
		if errors.Is(err, ErrAccountExists) {
			response.JSON(w, http.StatusConflict, response.Error{Message: "Email already exists"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, response.Success{Message: "User registered successfully"})
}

// SignIn handles user sign in
// @Summary Sign in user
// @Description Authenticate user with email and password, returns JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Sign in request with user credentials"
// @Success 200 {object} response.Success "Sign in successful"
// @Failure 400 {object} response.Error "Invalid request body"
// @Failure 401 {object} response.Error "Invalid credentials"
// @Failure 422 {object} response.Error "Validation errors"
// @Failure 423 {object} response.Error "Account locked"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /sign-in [post]
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	// Validate request DTO
	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.Errors)
		return
	}

	data, err := h.authUsecase.SignIn(r.Context(), req, r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCreds):
			response.JSON(w, http.StatusUnauthorized, response.Error{Message: "Invalid Email or Password"})
			return

		case errors.Is(err, ErrLocked):
			response.JSON(w, http.StatusForbidden, response.Error{Message: "Your account has been locked"})
			return

		default:
			response.InternalError(w)
			return
		}
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data, Message: "Sign-in successful"})
}

// SignIn handles guest sign in
// @Summary Sign in guest
// @Description Authenticate guest user without credentials, returns limited access tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInGuestRequest true "Guest sign in request with optional user agent"
// @Success 200 {object} response.Success "Guest sign in successful"
// @Failure 400 {object} response.Error "Invalid request body"
// @Failure 403 {object} response.Error "Guest sign in disabled"
// @Failure 429 {object} response.Error "Guest session limit reached"
// @Failure 500 {object} response.Error "Internal server error"
// @Router /sign-in-guest [post]
func (h *AuthHandler) SignInGuest(w http.ResponseWriter, r *http.Request) {

	// Parse request body
	var req SignInGuestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	// Validate request DTO
	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.Errors)
		return
	}

	data, err := h.authUsecase.SignInGuest(r.Context(), req, r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, ErrGuestDisabled):
			response.JSON(w, http.StatusForbidden, response.Error{Message: "Guest sign-in is currently disabled"})
			return

		case errors.Is(err, ErrGuestLimited):
			response.JSON(w, http.StatusTooManyRequests, response.Error{Message: "Guest session limit reached"})
			return

		default:
			response.InternalError(w)
			return
		}
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data, Message: "Sign-in as guest successful"})
}

// SignOut handles user sign out
// @Summary Sign out user
// @Description Revoke user session and invalidate JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Success "Successfully signed out"
// @Failure 401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure 500 {object} response.Error "Internal server error"
// @Security ApiKeyAuth
// @Router /sign-out [post]
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)

	if err := h.authUsecase.SignOut(ctx, claim.SessionID); err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Message: "Sign-out successful"})
}

// RefreshToken handles JWT token refresh
// @Summary Refresh JWT token
// @Description Generate new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.Success "Token refreshed successfully"
// @Failure 401 {object} response.Error "Invalid or expired refresh token"
// @Failure 500 {object} response.Error "Internal server error"
// @Security ApiKeyAuth
// @Router /refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	data, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrExpiredRefreshToken) {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			// Encode and send the response
			json.NewEncoder(w).Encode(response.Error{Message: "Invalid or expired refresh token"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data, Message: "Token refreshed successfully"})
}
