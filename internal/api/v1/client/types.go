package client

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

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

// IsNotFound returns true if the error is a 404 Not Found error
func IsNotFound(err error) bool {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code == http.StatusNotFound
	}
	return false
}

// IsBadRequest returns true if the error is a 400 Bad Request error
func IsBadRequest(err error) bool {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code == http.StatusBadRequest
	}
	return false
}
