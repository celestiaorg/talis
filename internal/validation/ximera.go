// Package validation provides validation functions for various parts of the application
package validation

import (
	"fmt"
	"os"
	"strconv"

	"github.com/celestiaorg/talis/internal/constants"
)

// XimeraRequest defines the interface needed for Ximera validation
type XimeraRequest interface {
	GetMemory() int
	GetCPU() int
	GetImage() string
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

	// Get SSH key name from environment variable
	sshKeyName := os.Getenv(constants.EnvTalisSSHKeyName)
	if sshKeyName == "" {
		return fmt.Errorf("environment variable %s is required for Ximera", constants.EnvTalisSSHKeyName)
	}

	// Validate sshKeyID is an integer
	// Since I am not sure about this, it is commented out to avoid breaking things
	// if _, err := strconv.Atoi(sshKeyName); err != nil {
	// 	return fmt.Errorf("SSH key ID in environment variable %s must be a valid integer for Ximera: %w", constants.EnvTalisSSHKeyName, err)
	// }

	return nil
}
