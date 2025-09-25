package validator

import (
	"regexp"
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

var EmailPattern = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
var URLPattern = regexp.MustCompile(`^(https?://[^\s]+)$`)
