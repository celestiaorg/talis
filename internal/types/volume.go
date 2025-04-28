package types

// Volume represents a storage volume configuration
type Volume struct {
	Name       string `json:"name"`
	SizeGB     int    `json:"size_gb"`
	MountPoint string `json:"mount_point"`
}

// ValidationError represents an error that occurs during validation
type ValidationError struct {
	Message string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}
