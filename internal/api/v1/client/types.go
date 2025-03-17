package client

import "github.com/celestiaorg/talis/internal/types/infrastructure"

// ErrorResponse represents the standard error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// InstanceMetadataResponse represents the metadata response for instances
type InstanceMetadataResponse struct {
	Instances []infrastructure.InstanceInfo `json:"instances"` // List of instances
	Total     int                           `json:"total"`     // Total number of instances
	Page      int                           `json:"page"`      // Current page number
	Limit     int                           `json:"limit"`     // Page size limit
	Offset    int                           `json:"offset"`    // Pagination offset
}
