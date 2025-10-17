package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/response"
)

type AuthHandler struct {
	authUsecase AuthUsecase
}

func NewAuthHandler(authUsecase AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase}
}

// SignUp handles user registration
// @Summary Sign up new user
// @Description Register a new user account with email, password, and profile information
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Sign up request with user details"
// @Success 201 {object} response.Message "User registered successfully"
// @Failure 400 {object} response.Message "Invalid request body"
// @Failure 422 {object} response.Error "Validation errors"
// @Failure 409 {object} response.Message "Email already exists"
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
			response.JSON(w, http.StatusConflict, response.Message{Message: "Email already exists"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, response.Message{Message: "User registered successfully"})
}

// SignIn handles user sign in
// @Summary Sign in user
// @Description Authenticate user with email and password, returns JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Sign in request with user credentials"
// @Success 200 {object} response.Success{data=SignInResponse} "Sign in successful"
// @Failure 400 {object} response.Message "Invalid request body"
// @Failure 401 {object} response.Message "Invalid email or password"
// @Failure 422 {object} response.Error "Validation errors"
// @Failure 423 {object} response.Message "Your account has been locked"
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
			response.JSON(w, http.StatusUnauthorized, response.Message{Message: "Invalid email or password"})
			return

		case errors.Is(err, ErrLocked):
			response.JSON(w, http.StatusForbidden, response.Message{Message: "Your account has been locked"})
			return

		default:
			response.InternalError(w)
			return
		}
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data})
}

// SignIn handles guest sign in
// @Summary Sign in guest
// @Description Authenticate guest user without credentials, returns limited access tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInGuestRequest true "Guest sign in request with optional user agent"
// @Success 200 {object} response.Success{data=SignInGuestResponse} "Guest sign in successful"
// @Failure 400 {object} response.Message "Invalid request body"
// @Failure 403 {object} response.Message "Guest sign in disabled"
// @Failure 422 {object} response.Error "Validation errors"
// @Failure 429 {object} response.Message "Guest session limit reached"
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
			response.JSON(w, http.StatusForbidden, response.Message{Message: "Guest sign in disabled"})
			return

		case errors.Is(err, ErrGuestLimited):
			response.JSON(w, http.StatusTooManyRequests, response.Message{Message: "Guest session limit reached"})
			return

		default:
			response.InternalError(w)
			return
		}
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data})
}

// SignOut handles user sign out
// @Summary Sign out user
// @Description Revoke user session and invalidate JWT tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Message "Sign out successfully"
// @Security ApiKeyAuth
// @Router /sign-out [post]
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)

	if err := h.authUsecase.SignOut(ctx, claim.Sub); err != nil {
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Message{Message: "Sign out successfully"})
}

// RefreshToken handles JWT token refresh
// @Summary Refresh JWT token
// @Description Generate new access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} response.Success{data=RefreshTokenResponse} "Token refreshed successfully"
// @Failure 401 {object} response.Message "Invalid or expired refresh token"
// @Security ApiKeyAuth
// @Router /refresh-token [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	// Validate request DTO
	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.Errors)
		return
	}

	data, err := h.authUsecase.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrExpiredRefreshToken) {
			response.JSON(w, http.StatusUnauthorized, response.Message{Message: "Invalid or expired refresh token"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: data})
}
