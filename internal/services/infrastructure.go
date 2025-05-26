// Package services provides business logic implementation for the API
package services

import (
	"fmt"
	"os"

	"github.com/celestiaorg/talis/internal/constants"
	"github.com/celestiaorg/talis/internal/logger"
)

// GetTalisServerSSHKey returns the SSH key content from the environment variable
func GetTalisServerSSHKey() (string, error) {
	// Get the key content from the environment variable
	keyContent := os.Getenv(constants.EnvTalisSSHKey)
	if keyContent == "" {
		return "", fmt.Errorf("SSH key not found in environment variable %s", constants.EnvTalisSSHKey)
	}

	logger.Debug("Using SSH key content from environment variable")
	return keyContent, nil
}
