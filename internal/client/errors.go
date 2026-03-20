package client

import (
	"fmt"
	"net/http"
	"strings"
)

// APIError represents an error response from the registry API.
type APIError struct {
	StatusCode int
	Message    string
	Fields     map[string]string
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	if len(e.Fields) > 0 {
		parts := make([]string, 0, len(e.Fields))
		for k, v := range e.Fields {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		}
		msg += " (fields: " + strings.Join(parts, ", ") + ")"
	}
	return msg
}

// IsNotFound returns true if the error is an HTTP 404.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict returns true if the error is an HTTP 409.
func IsConflict(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsUnauthorized returns true if the error is an HTTP 401.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden returns true if the error is an HTTP 403.
func IsForbidden(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusForbidden
	}
	return false
}
