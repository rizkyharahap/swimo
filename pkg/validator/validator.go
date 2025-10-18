package validator

import (
	"net/mail"
	"net/url"
	"strings"
)

// ValidationError is a custom error type to hold multiple validation messages.
type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString("validation failed:")
	for field, msg := range e.Errors {
		sb.WriteString(" ")
		sb.WriteString(field)
		sb.WriteString(": ")
		sb.WriteString(msg)
	}
	return sb.String()
}

func IsValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func IsValidURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}
