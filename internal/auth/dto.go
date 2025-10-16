package auth

import (
	"strings"

	"github.com/rizkyharahap/swimo/pkg/validator"
)

// SignUpRequest represents the sign up request data transfer object
type SignUpRequest struct {
	Email           string   `json:"email" example:"john@example.com"`
	Password        string   `json:"password" example:"SecurePassword123"`
	ConfirmPassword string   `json:"confirmPassword" example:"SecurePassword123"`
	Name            string   `json:"name" example:"John Doe"`
	Weight          *float64 `json:"weight" example:"75.5"`
	Height          *float64 `json:"height" example:"180"`
	Age             *int16   `json:"age" example:"30"`
}

// SignInRequest represents the sign in request data transfer object
type SignInRequest struct {
	Email    string `json:"email" example:"john@example.com"`
	Password string `json:"password" example:"SecurePassword123"`
}

// SignInResponse represents the sign in response data transfer object
type SignInResponse struct {
	Name         string   `json:"name" example:"John Doe"`
	Weight       *float64 `json:"weight" example:"75.5"`
	Height       *float64 `json:"height" example:"180"`
	Age          *int16   `json:"age" example:"30"`
	Email        string   `json:"email" example:"john@example.com"`
	Token        string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string   `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresInMs  int64    `json:"expiresIn" example:"900000"`
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

// Validate validates the sign up request
func (r *SignInRequest) Validate() *validator.ValidationError {
	errors := make(map[string]string)

	sanitizedEmail := strings.TrimSpace(strings.ToLower(r.Email))
	if !validator.EmailPattern.MatchString(sanitizedEmail) {
		errors["email"] = "Email is not a valid format"
	}

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

	sanitizedEmail := strings.TrimSpace(strings.ToLower(r.Email))
	if !validator.EmailPattern.MatchString(sanitizedEmail) {
		errors["email"] = "Email is not a valid format"
	}

	if r.Password == "" {
		errors["password"] = "Password is required"
	} else if len(r.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if r.ConfirmPassword == "" {
		errors["confirmPassword"] = "Confirm password is required"
	}
	if r.Password != r.ConfirmPassword {
		errors["confirmPassword"] = "Confirm passwords do not match"
	}

	if strings.TrimSpace(r.Name) == "" {
		errors["name"] = "Name is required"
	}

	if r.Weight == nil {
		errors["weight"] = "Weight is required"
	} else if *r.Weight < 0 {
		errors["weight"] = "Weight cannot be negative"
	}

	if r.Height == nil {
		errors["height"] = "Height is required"
	} else if *r.Height < 0 {
		errors["height"] = "Height cannot be negative"
	}

	if r.Age == nil {
		errors["age"] = "Age is required"
	} else if *r.Age < 0 {
		errors["age"] = "Age cannot be negative"
	}

	if len(errors) > 0 {
		return &validator.ValidationError{Errors: errors}
	}

	return nil
}
