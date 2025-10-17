package auth

import (
	"strings"

	"github.com/rizkyharahap/swimo/pkg/validator"
)

// SignUpRequest represents the sign up request data transfer object
type SignUpRequest struct {
	Name            string  `json:"name" example:"John Doe"`
	Email           string  `json:"email" example:"john@example.com"`
	Password        string  `json:"password" example:"SecurePassword123"`
	ConfirmPassword string  `json:"confirmPassword" example:"SecurePassword123"`
	Age             int16   `json:"age" example:"30"`
	Height          float64 `json:"height" example:"180"`
	Weight          float64 `json:"weight" example:"75.5"`
}

// SignInRequest represents the sign in request data transfer object
type SignInRequest struct {
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"SecurePassword123"`
}

// SignInResponse represents the sign in response data transfer object
type SignInResponse struct {
	Name         string  `json:"name" example:"John Doe"`
	Email        string  `json:"email" example:"john@example.com"`
	Age          int16   `json:"age" example:"30"`
	Height       float64 `json:"height" example:"180"`
	Weight       float64 `json:"weight" example:"75.5"`
	Token        string  `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string  `json:"refreshToken" example:"3d3dc788634e05b7d1d5fac06834d3b6a9b62..."`
	ExpiresIn    int64   `json:"expiresIn" example:"1799999"`
}

type SignInGuestRequest struct {
	Age    int16   `json:"age" example:"30"`
	Height float64 `json:"height" example:"180"`
	Weight float64 `json:"weight" example:"75.5"`
}

type SignInGuestResponse struct {
	Name         string  `json:"name" example:"John Doe"`
	Age          int16   `json:"age" example:"30"`
	Height       float64 `json:"height" example:"180"`
	Weight       float64 `json:"weight" example:"75.5"`
	Token        string  `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string  `json:"refreshToken" example:"3d3dc788634e05b7d1d5fac06834d3b6a9b62..."`
	ExpiresIn    int64   `json:"expiresIn" example:"1799999"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" example:"3d3dc788634e05b7d1d5fac06834d3b6a9b62..."`
}

type RefreshTokenResponse struct {
	Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"3d3dc788634e05b7d1d5fac06834d3b6a9b62..."`
	ExpiresIn    int64  `json:"expiresInMs" example:"1799999"`
}

func (r *SignUpRequest) ToUserEntity(accountID string) *User {
	return &User{
		AccountID: accountID,
		Name:      strings.TrimSpace(r.Name),
		WeightKG:  r.Weight,
		HeightCM:  r.Height,
		AgeYears:  r.Age,
	}
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

// Validate validates the sign in request
func (r *SignInRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	r.Email = strings.ToLower(trim(r.Email))
	if r.Email == "" {
		errors["email"] = "Email is required"
	} else if !validator.IsValidEmail(r.Email) {
		errors["email"] = "Email is not a valid format"
	}

	r.Password = trim(r.Password)
	if r.Password == "" {
		errors["password"] = "Password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}

// Validate validates the sign up request
func (r *SignUpRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	r.Email = strings.ToLower(trim(r.Email))
	if r.Email == "" {
		errors["email"] = "Email is required"
	} else if !validator.IsValidEmail(r.Email) {
		errors["email"] = "Email is not a valid format"
	}

	r.Password = trim(r.Password)
	if r.Password == "" {
		errors["password"] = "Password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	r.ConfirmPassword = trim(r.ConfirmPassword)
	if r.ConfirmPassword == "" {
		errors["confirmPassword"] = "Confirm password is required"
	} else if r.Password != r.ConfirmPassword {
		errors["confirmPassword"] = "Confirm passwords do not match"
	}

	r.Name = trim(r.Name)
	if r.Name == "" {
		errors["name"] = "Name is required"
	}

	if r.Weight <= 0 {
		errors["weight"] = "Weight must be a positive number"
	}

	if r.Height <= 0 {
		errors["height"] = "Height cannot be negative"
	}

	if r.Age <= 0 {
		errors["age"] = "Age must be a positive number"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}

// Validate validates the sign in guest request
func (r *SignInGuestRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	if r.Weight <= 0 {
		errors["weight"] = "Weight must be a positive number"
	}

	if r.Height <= 0 {
		errors["height"] = "Height must be a positive number"
	}

	if r.Age <= 0 {
		errors["age"] = "Age must be a positive number"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}

// Validate validates the sign in guest request
func (r *RefreshTokenRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	r.RefreshToken = trim(r.RefreshToken)
	if r.RefreshToken == "" {
		errors["refresh_token"] = "Refresh token is required"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
