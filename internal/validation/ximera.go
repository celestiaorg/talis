// Package validation provides validation functions for various parts of the application
package validation

import (
	"fmt"
	"strconv"
)

// XimeraRequest defines the interface needed for Ximera validation
type XimeraRequest interface {
	GetMemory() int
	GetCPU() int
	GetImage() string
	GetSSHKeyName() string
}

// XimeraInstanceRequest validates Ximera-specific fields in an instance request
func XimeraInstanceRequest(req XimeraRequest) error {
	if req.GetMemory() <= 0 {
		return fmt.Errorf("memory is required and must be > 0 for Ximera")
	}
	if req.GetCPU() <= 0 {
		return fmt.Errorf("cpu is required and must be > 0 for Ximera")
	}
	// Validate osID (Image) is an integer
	if _, err := strconv.Atoi(req.GetImage()); err != nil {
		return fmt.Errorf("image (osID) must be a valid integer for Ximera: %w", err)
	}
	// Validate sshKeyID (SSHKeyName) is an integer
	if _, err := strconv.Atoi(req.GetSSHKeyName()); err != nil {
		return fmt.Errorf("ssh_key_name (sshKeyID) must be a valid integer for Ximera: %w", err)
	}
	return nil
}
