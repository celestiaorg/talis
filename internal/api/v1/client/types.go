package client

// ErrorResponse represents the standard error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}
