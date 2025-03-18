package infrastructure

// Response represents the response from the infrastructure API
type Response struct {
	ID     uint   `json:"id"`     // ID of the job
	Status string `json:"status"` // Status of the job
}
