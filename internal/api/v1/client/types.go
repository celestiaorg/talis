package client

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
