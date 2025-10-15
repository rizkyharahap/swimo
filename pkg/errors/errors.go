package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error codes
const (
	CodeInternalError     = "INTERNAL_ERROR"
	CodeValidationError   = "VALIDATION_ERROR"
	CodeNotFound          = "NOT_FOUND"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeBadRequest        = "BAD_REQUEST"
	CodeConflict          = "CONFLICT"
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// AppError represents an application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWithDetails creates a new AppError with details
func NewWithDetails(code, message string, details any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Wrap wraps an existing error into an AppError
func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// WrapWithDetails wraps an existing error into an AppError with details
func WrapWithDetails(err error, code, message string, details any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
		Cause:   err,
	}
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int      `json:"status"`
	Error      AppError `json:"error"`
}

// ToHTTP converts an AppError to HTTPError with appropriate status code
func ToHTTP(err *AppError) *HTTPError {
	statusCode := http.StatusInternalServerError

	switch err.Code {
	case CodeValidationError:
		statusCode = http.StatusBadRequest
	case CodeNotFound:
		statusCode = http.StatusNotFound
	case CodeUnauthorized:
		statusCode = http.StatusUnauthorized
	case CodeForbidden:
		statusCode = http.StatusForbidden
	case CodeBadRequest:
		statusCode = http.StatusBadRequest
	case CodeConflict:
		statusCode = http.StatusConflict
	case CodeRateLimitExceeded:
		statusCode = http.StatusTooManyRequests
	}

	return &HTTPError{
		StatusCode: statusCode,
		Error:      *err,
	}
}

// ToJSON converts HTTPError to JSON bytes
func (e *HTTPError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Is checks if error matches the target error
func Is(err, target error) bool {
	if appErr, ok := err.(*AppError); ok {
		if targetErr, ok := target.(*AppError); ok {
			return appErr.Code == targetErr.Code
		}
	}
	return false
}

// As finds the first error in err's chain that matches target
func As(err error, target any) bool {
	if appErr, ok := err.(*AppError); ok {
		if targetPtr, ok := target.(**AppError); ok {
			*targetPtr = appErr
			return true
		}
	}
	return false
}

// Convenience functions for common error types
func InternalError(message string) *AppError {
	return New(CodeInternalError, message)
}

func ValidationError(message string) *AppError {
	return New(CodeValidationError, message)
}

func ValidationErrorWithDetails(message string, details any) *AppError {
	return NewWithDetails(CodeValidationError, message, details)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

func Conflict(message string) *AppError {
	return New(CodeConflict, message)
}

func RateLimitExceeded(message string) *AppError {
	return New(CodeRateLimitExceeded, message)
}
