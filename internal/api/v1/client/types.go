package client

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the API
type APIError struct {
	// StatusCode is the HTTP status code
	StatusCode int

	// Message is the error message
	Message string

	// RawBody is the raw response body
	RawBody string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status=%d, message=%s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsBadRequest returns true if the error is a 400 Bad Request error
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden error
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsServerError returns true if the error is a 5xx Server Error
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// NewAPIError creates a new APIError
func NewAPIError(statusCode int, message, rawBody string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		RawBody:    rawBody,
	}
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) (*APIError, bool) {
	if err == nil {
		return nil, false
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// ErrorResponse represents the standard error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// CreateResponse represents the response from creating infrastructure
type CreateResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

// DeleteResponse represents the response from deleting infrastructure
type DeleteResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}
